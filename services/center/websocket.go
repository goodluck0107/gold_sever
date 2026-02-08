package center

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static/authentication"
	"github.com/open-source/game/chess.git/pkg/static/util"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"io"
	"net"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/sirupsen/logrus"

	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"

	"golang.org/x/net/websocket"
)

type Session struct {
	ID        int64           //! 自增长id
	Uid       int64           //! uid
	Ws        *websocket.Conn //! websocket
	IP        string          //! ip
	SendChan  chan []byte     //! 发送消息管道
	LineTime  int64           //! 链接时间
	FirstMsg  bool            //! 是否有第一个消息
	MsgTime   []int64         //! 时间切片
	MsgHeader []string

	PingTime int64 //！上次心跳时间
	TryNum   int
	Stop     int64
	lockchan *lock2.RWMutex
}

// ! 发送消息
func (self *Session) SendBuf(buf []byte) {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	if self == nil {
		return
	}
	self.lockchan.Lock()
	if self.Stop == 1 {
		self.lockchan.Unlock()
		return
	}
	select {
	case self.SendChan <- buf:
		if len(self.SendChan) > sendChanSize-10 && self.Uid > 0 {
			self.lockchan.Unlock()
			self.onClose(0)
			return
		}
		self.lockchan.Unlock()
		return
	case <-time.After(50 * time.Millisecond):
		xlog.Logger().Errorf("send msg to chan error. chan is nil:%v, chan len:%v ,val:%v", self.SendChan == nil, len(self.SendChan), string(buf))
		self.lockchan.Unlock()
		self.onClose(0)
		return
	}
}

// ! 发送消息
func (self *Session) SendMsg(head string, errCode int16, v interface{}, uid int64) {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	if head == "" {
		return
	}

	if self == nil {
		return
	}
	self.lockchan.Lock()
	if self.Stop == 1 {
		self.lockchan.Unlock()
		return
	}
	buf := static.HF_EncodeMsg(head, errCode, v, GetServer().Con.Encode, GetServer().Con.EncodeClientKey, uid)

	select {
	case self.SendChan <- buf:
		if len(self.SendChan) > sendChanSize-10 && self.Uid > 0 {
			self.lockchan.Unlock()
			self.onClose(0)
			return
		}
		self.lockchan.Unlock()
		return
	case <-time.After(50 * time.Millisecond):
		xlog.Logger().Errorf("send msg to chan error. chan is nil:%v, chan len:%v ,val:%v", self.SendChan == nil, len(self.SendChan), string(buf))
		self.lockchan.Unlock()
		self.onClose(0)
		return
	}
}

func (self *Session) SafeClose(origin uint8) {

	switch origin {
	case consts.SESSION_CLOED_CONNECT:
		self.SendMsg(consts.MsgCloseConnection, xerrors.SuccessCode, "", self.Uid)
	case consts.SESSION_CLOED_FORCE:
		self.SendMsg(consts.MsgForceCloseConnection, xerrors.SuccessCode, "", self.Uid)
	case consts.SESSION_CLOED_ASYSTOLE:
		self.SendMsg(consts.MsgForceCloseConnection, xerrors.SuccessCode, "", self.Uid)
	default:
		self.SendMsg(consts.MsgCloseConnection, xerrors.SuccessCode, &static.Msg_Null{}, self.Uid)
	}
	self.onClose(0)
	return
}

// ! 消息run
func (self *Session) Run() {
	go self.sendMsgRun()
	self.receiveMsgRun()
}

// ! 发送消息循环
func (self *Session) sendMsgRun() {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	for msg := range self.SendChan {
		if GetServer().ShutDown {
			break
		}

		if string(msg) == "" {
			break
		}

		exit := false
		for {
			self.Ws.SetWriteDeadline(time.Now().Add(socketTimeOut * time.Second))

			err := websocket.Message.Send(self.Ws, msg)
			if err != nil {
				errnet, ok := err.(net.Error)
				if ok && errnet.Timeout() {
					self.TryNum++
					if self.Uid > 0 && self.TryNum < socketTryNum {
						time.Sleep(time.Millisecond * 10)
						continue
					}
				}
				xlog.Logger().Warnf("sendmsgfail uid:", self.Uid, "-err:", err.Error())
				xlog.Logger().Warnf("sendmsgfail uid:", self.Uid, "-msg:", static.HF_Bytestoa(msg))
				exit = true
				break
			} else {
				// 发送成功
				break
			}
		}
		if exit {
			break
		}
	}
	xlog.Logger().Warningln("client close", self.Uid)
	self.onClose(0)
}

// ! 接收消息循环
func (self *Session) receiveMsgRun() {
	for {
		if GetServer().ShutDown { //! 关服
			break
		}
		// if !self.checkHeartbeat() {
		// 	break
		// }
		var msg []byte
		self.Ws.SetReadDeadline(time.Now().Add(socketTimeOut * time.Second))
		err := websocket.Message.Receive(self.Ws, &msg)
		if err != nil {
			neterr, ok := err.(net.Error)
			if ok && neterr.Timeout() {
				continue
			}
			if err == io.EOF {
				xlog.Logger().Warningln("client disconnet")
			} else {
				xlog.Logger().Warningln("receive err:", err)
			}
			break
		}

		if len(msg) >= 1024*50 {
			static.GetBlackIpMgr().AddIp(self.IP, "消息超过1024*50")
			break
		}

		if len(self.MsgTime) >= 50 { //! 取20个消息的间隔如果小于3秒
			if self.MsgTime[49]-self.MsgTime[0] <= 3 {
				static.GetBlackIpMgr().AddIp(self.IP, "消息速度过快")
				var str string
				str += fmt.Sprintf("\nuid:%d", self.Uid)
				for _, header := range self.MsgHeader {
					str += fmt.Sprintf("\n[%s]", header)
				}
				xlog.Logger().Errorln(str)
				break
			}
			self.MsgTime = make([]int64, 0)
			self.MsgHeader = make([]string, 0)
		}
		self.MsgTime = append(self.MsgTime, time.Now().Unix())
		self.onReceive(msg)
	}
	xlog.Logger().Warningln("client close", self.ID)
	self.onClose(self.Uid)
}

func (self *Session) onReceive(msg []byte) {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	head, sign, _, data, ok, msguid, msgip := static.HF_DecodeMsg(msg)
	if !ok {
		static.GetBlackIpMgr().AddIp(self.IP, "消息解析错误1")
		xlog.Logger().Errorln("uid = %, head = %s, ip = %s 被加入黑名单", msguid, head, msgip)
		self.SafeClose(consts.SESSION_CLOED_FORCE)
		return
	}

	ip := self.IP
	if msgip != "" {
		ip = msgip
	}

	//// 消息uid必须>0
	//if msguid <= 0 {
	//	self.SendMsg(head, xerrors.TokenError.Code, xerrors.TokenError.Msg, msguid)
	//	self.SendChan <- []byte("")
	//}

	//! 收到的第一个有效消息必须为setuid
	if !self.FirstMsg {
		// 非登陆协议断开连接
		if head == consts.MsgHouseOwnerChange {
			self.Uid = msguid
			fmt.Println("msg uid:", msguid)
		} else if head != consts.MsgTypeHallLogin {
			self.SafeClose(consts.SESSION_CLOED_CONNECT)

			// 如果在setuid之前请求协议，直接给return 不返回任何提示 并 断开连接
			xlog.Logger().WithFields(logrus.Fields{
				"head": head,
				"uid":  self.Uid,
			}).Errorln("【Request Before Setuid】")
			return
		}
		self.FirstMsg = true
	}

	self.MsgHeader = append(self.MsgHeader, head)

	// 如果使用加密了但客户端消息未加密则直接报错
	if GetServer().Con.Encode != consts.EncodeNone && sign.Encode == consts.EncodeNone {
		self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
		self.SendMsg(consts.MsgCloseConnection, xerrors.SuccessCode, &static.Msg_Null{}, msguid)
		return
	}

	// 消息解密处理
	var err error
	data, err = decryptData(data, sign.Encode)
	if err != nil {
		xlog.Logger().Errorln("Service decrypt err:", err.Error())
		self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, msguid)
		self.SafeClose(consts.SESSION_CLOED_CONNECT)
		return
	}

	if head != "ping" && head != consts.MsgTypeHouseMemberOnline { //屏蔽掉ping的log
		xlog.Logger().WithFields(logrus.Fields{
			"uid":  self.Uid,
			"head": head,
			"data": string(data),
		}).Infoln("【RECEIVED SOCKET】")
	}

	// 统计消息协议的使用次数
	GetProtocolMgr().AddProtocolCnt(head)

	switch head {
	case consts.MsgTypeHallLogin: //登录
		var msg static.Msg_CH_SetUid
		err := json.Unmarshal(data, &msg)
		if err != nil {
			xlog.Logger().Errorln(err)
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_CONNECT)
			return
		}

		// 同一时间只处理1条登陆消息
		onlykey := fmt.Sprintf("userstatus_doing_%d", msg.Uid)
		cli := GetDBMgr().Redis
		if !cli.SetNX(onlykey, msg.Uid, time.Second*3).Val() {
			xlog.Logger().Errorf("SafeClose for userstatus_doing uid : %d", msg.Uid)
			self.SafeClose(consts.SESSION_CLOED_CONNECT)
			return
		}
		defer cli.Del(onlykey)

		// 正在执行加入牌桌流程
		onlyjoin := fmt.Sprintf("userstatus_doing_join_%d", msg.Uid)
		count := 0
		for {
			if count > 10 {
				xlog.Logger().Errorf("SafeClose for onlyjoin uid : %d", msg.Uid)
				self.SendMsg(head, xerrors.UserStatusJoinError.Code, xerrors.UserStatusJoinError.Msg, msg.Uid)
				self.SafeClose(consts.SESSION_CLOED_CONNECT)
				return
			}
			if cli.Exists(onlyjoin).Val() == 1 {
				time.Sleep(time.Millisecond * 200)
				count++
				continue
			}
			break
		}

		// 正在执行退出流程
		onlyExit := fmt.Sprintf("userstatus_doing_exit_%d", msg.Uid)
		count = 0
		for {
			if GetDBMgr().Redis.Exists(onlyExit).Val() == 1 {
				if count > 10 {
					xlog.Logger().Errorf("SafeClose for onlyExit uid : %d", msg.Uid)
					self.SendMsg(head, xerrors.UserStatusExitError.Code, xerrors.UserStatusExitError.Msg, 0)
					self.SafeClose(consts.SESSION_CLOED_CONNECT)
					return
				}
				time.Sleep(time.Millisecond * 800)
				count++
				continue
			}
			break
		}

		//// 正在执行游戏加入流程
		//onlyJoinGame := fmt.Sprintf("userstatus_doing_joingame_%d", msg.Uid)
		//count = 0
		//for {
		//	if GetDBMgr().Redis.Exists(onlyJoinGame).Val() == 1 {
		//		if count > 10 {
		//			self.SendMsg(head, xerrors.UserStatusExitError.Code, xerrors.UserStatusExitError.Msg, 0)
		//			self.SafeClose(constant.SESSION_CLOED_CONNECT)
		//			return
		//		}
		//		time.Sleep(time.Microsecond * 200)
		//		count++
		//		continue
		//	}
		//	break
		//}

		person, err := GetDBMgr().GetDBrControl().GetPerson(msg.Uid)
		if err != nil || person == nil {
			xlog.Logger().WithFields(logrus.Fields{
				"errinfo":    err.Error(),
				"personinfo": person,
			}).Errorln("大厅登录从redis获取玩家信息异常")
			// 用户不存在
			self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, msguid)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 更新玩家弱数据
		UpdateLazyUser(msg.Uid, person.Nickname, person.Sex, person.Imgurl)

		// 玩家登陆存在观战就退出观战
		if person.WatchTable > 0 {
			GetDBMgr().GetDBrControl().RemoveWatchPlayerToTable(person.WatchTable, person.Uid)

			GetClubMgr().RemoveHouseTableWatch(person.WatchTable, person.Uid)

			person.WatchTable = 0
		}

		person.Ip = self.IP
		// 校验token
		if person.Token != msg.Token || person.Token == "" || person.Uid == 0 {
			xlog.Logger().Errorf("SafeClose for Token uid : %d", msg.Uid)
			// 无效token
			self.SendMsg(head, xerrors.TokenError.Code, xerrors.TokenError.Msg, msguid)
			self.SafeClose(consts.SESSION_CLOED_CONNECT)
			return
		}

		isCheckOut := true
		// 校验houseid
		if person.HouseId != 0 {
			house := GetClubMgr().GetClubHouseByHId(person.HouseId)
			if house == nil {
				person.HouseId = 0
				person.FloorId = 0
				isCheckOut = false
			} else {
				hmem := house.GetMemByUId(person.Uid)
				if hmem != nil {
					if hmem.Lower(consts.ROLE_MEMBER) {
						person.HouseId = 0
						person.FloorId = 0
						isCheckOut = false
					}
				}
			}
			if person.FloorId != 0 {
				floor := house.GetFloorByFId(person.FloorId)
				if floor == nil {
					person.FloorId = 0
					isCheckOut = false
				}
			}
		}
		// 校验floorid
		if person.FloorId != 0 {
			if person.HouseId == 0 {
				person.FloorId = 0
				isCheckOut = false
			}
		}

		if person.SiteId == 0 && person.TableId != 0 {
			table := GetTableMgr().GetTable(person.TableId)
			if table == nil {
				person.TableId = 0
				isCheckOut = false
				info := models.CheckUserInBlank(msg.Uid, GetDBMgr().GetDBmControl())
				if info != nil {
					self.SendMsg(head, xerrors.BlankUserErrorCode, info.Reason, 0)
					return
				}
			}
		} else {
			info := models.CheckUserInBlank(msg.Uid, GetDBMgr().GetDBmControl())
			if info != nil {
				self.SendMsg(head, xerrors.BlankUserErrorCode, info.Reason, 0)
				return
			}
			// 如果用户不在游戏中, 则判断服务器是否维护
			if notice := CheckServerMaintainWithWhite(person.Uid, static.NoticeMaintainServerAllServer); notice != nil {
				// 如果有维护, 则屏蔽登录请求
				// self.SendMsg(head, xerrors.ServerMaintainError.Code, xerrors.ServerMaintainError.Msg, 0)
				xlog.Logger().Errorf("SafeClose for CheckMaintainWithWhite uid : %d", msg.Uid)
				self.SendMsg(consts.MsgTypeMaintainNotice, xerrors.SuccessCode, notice, self.Uid)
				self.SafeClose(consts.SESSION_CLOED_CONNECT)
				return
			}
		}

		// 更新玩家的游戏引擎信息
		person.Engine = static.HF_GetGameEngine(msg.Engine)
		if err := GetDBMgr().db_R.UpdatePersonAttrs(person.Uid, "Engine", person.Engine); err != nil {
			xlog.Logger().Errorf("SafeClose for HF_GetGameEngine uid : %d", msg.Uid)
			self.SendMsg(head, xerrors.DBExecError.Code, xerrors.DBExecError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_CONNECT)
			xlog.Logger().Errorln("更新玩家游戏引擎信息失败：", err)
			return
		}

		person.Ip = self.IP
		GetDBMgr().db_R.UpdatePersonAttrs(person.Uid, "Ip", self.IP, "Online", true, "LastLoginTime", time.Now().Unix())

		person.Online = true
		if !isCheckOut {
			if err := GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid,
				"HouseId", person.HouseId,
				"FloorId", person.FloorId,
				"TableId", person.TableId,
				"Engine", person.Engine,
				"Ip", self.IP,
				"Online", true); err != nil {
				self.SendMsg(head, xerrors.ServerMaintainError.Code, xerrors.ServerMaintainError.Msg, 0)
				self.SafeClose(consts.SESSION_CLOED_CONNECT)
				return
			}
		} else {
			if err := GetDBMgr().db_R.UpdatePersonAttrs(person.Uid, "Ip", self.IP, "Online", true); err != nil {
				self.SendMsg(head, xerrors.ServerMaintainError.Code, xerrors.ServerMaintainError.Msg, 0)
				self.SafeClose(consts.SESSION_CLOED_CONNECT)
				return
			}
		}

		//! 踢掉原来的人
		oldperson := GetPlayerMgr().GetPlayer(person.Uid)
		if oldperson != nil {
			if oldperson.session != nil {
				// 踢出大厅的人
				if oldperson.session != self {
					oldperson.SendMsg(consts.MsgTypeRelogin, nil)
					oldperson.CloseSession()
				}
			} else {
				// 判断是否在游戏服 若在则游戏服踢人
				if oldperson.Info.GameId != 0 {

					//onlyExitGame := fmt.Sprintf("userstatus_doing_exitgame_%d", msg.Uid)
					//cli := GetDBMgr().Redis
					//cli.SetNX(onlyExitGame, msg.Uid, time.Second * 3)

					// TODO 此处判定不严格，暂放宽登陆条件
					GetServer().CallGame(oldperson.Info.GameId, 0, "NewServerMsg", consts.MsgTypeRelogin, xerrors.SuccessCode, &static.Msg_HG_Relogin{Uid: oldperson.Info.Uid, TableId: oldperson.Info.TableId})

					//count = 0
					//for {
					//	if GetDBMgr().Redis.Exists(onlyExitGame).Val() == 1 {
					//		if count > 10 {
					//			//self.SendMsg(head, xerrors.UserStatusExitError.Code, xerrors.UserStatusExitError.Msg, 0)
					//			//self.SendMsg(constant.MsgCloseConnection, xerrors.SuccessCode, &public.Msg_Null{}, msguid)
					//			//self.SendChan <- []byte("")
					//			//return
					//			break
					//		}
					//		time.Sleep(time.Microsecond * 200)
					//		count++
					//		continue
					//	}
					//	break
					//}
				}
			}
		}

		//！玩家登录成功后，认为第一次ping成功
		self.setPingTime()
		newPerson := new(PlayerCenterMemory)
		newPerson.session = self
		newPerson.Info = *person
		newPerson.Ip = ip
		self.Uid = msg.Uid
		GetPlayerMgr().AddPerson(newPerson)
		GetClubMgr().CheckDefault(newPerson.Info.Uid)
		// 在玩家所有包厢标记玩家在线
		//GetClubMgr().UserOnline(person)

		// 推送用户信息
		self.SendMsg(head, xerrors.SuccessCode, getPersonInfo(newPerson), newPerson.Info.Uid)

		newPerson.OnLine(true)

		// 未实名用户计算累计时长
		totalTime := 0
		if len(person.IdcardAuthPI) == 0 {
			totalTime, _ = models.GetUserTotalPlayTime(GetDBMgr().GetDBmControl(), person.Uid)
		}

		// 是否需要上报
		bNeedReport := true
		// 超出限制时长 允许继续游戏 但不上报给官方
		if len(person.IdcardAuthPI) == 0 && totalTime > GetServer().ConGovAuth.LimitTimeNotAuth {
			bNeedReport = false
		} else if len(person.IdcardAuthPI) > 0 && static.HF_GetAgeFromIdcard(person.Idcard) < 18 && person.PlayTime >= GetServer().ConGovAuth.LimitTimeUnder18 {
			bNeedReport = false
		}
		// 向官方上报用户上线
		if bNeedReport {
			authentication.AuthenticationcollectionsUpload(newPerson.Info.Uid, newPerson.Info.Token, newPerson.Info.IdcardAuthPI, util.MD5(newPerson.Info.Nickname), authentication.AuthenticationOnlineStatusOnline)
		}

		// 检查未实名或已实名的未成年的游戏时长
		if len(person.IdcardAuthPI) == 0 && totalTime >= GetServer().ConGovAuth.LimitTimeNotAuth {
			if GetServer().ConGovAuth.IsForceNotice && person.GameId == 0 {
				var msg static.Msg_S2C_LimitPlayTime
				msg.Code = 1
				msg.Tips = fmt.Sprintf("亲爱的玩家，非常抱歉，根据国家相关法规规定，未实名账号累计时长不能超过%d小时，当前您已经累计%d小时，为了正常的游戏体验，请尽快完成实名登记！", GetServer().ConGovAuth.LimitTimeNotAuth/3600, totalTime/3600)
				newPerson.SendMsg(consts.MsgTypeLimitPlayTime, &msg)
			}
			// 向官方上报一条下线数据
			authentication.AuthenticationcollectionsUpload(newPerson.Info.Uid, newPerson.Info.Token, newPerson.Info.IdcardAuthPI, util.MD5(newPerson.Info.Nickname), authentication.AuthenticationOnlineStatusOffline)
		} else if len(person.IdcardAuthPI) > 0 && static.HF_GetAgeFromIdcard(person.Idcard) < 18 && person.PlayTime >= GetServer().ConGovAuth.LimitTimeUnder18 {
			if GetServer().ConGovAuth.IsForceNotice && person.GameId == 0 {
				var msg static.Msg_S2C_LimitPlayTime
				msg.Code = 2
				msg.Tips = fmt.Sprintf("亲爱的玩家，非常抱歉，根据国家相关法规规定，未成年用户累计时长不能超过%d小时，当前您已经累计%d小时，您当日不能再进行游戏！", GetServer().ConGovAuth.LimitTimeUnder18/3600, person.PlayTime/3600)
				newPerson.SendMsg(consts.MsgTypeLimitPlayTime, &msg)
			}
			// 向官方上报一条下线数据
			authentication.AuthenticationcollectionsUpload(newPerson.Info.Uid, newPerson.Info.Token, newPerson.Info.IdcardAuthPI, util.MD5(newPerson.Info.Nickname), authentication.AuthenticationOnlineStatusOffline)
		}

		// 检查已实名的未成年人是否在限定的时间范围内登录
		if len(person.IdcardAuthPI) > 0 && static.HF_GetAgeFromIdcard(person.Idcard) < 18 && len(GetServer().ConGovAuth.LimitTimeAtBeforeUnder18) > 0 && len(GetServer().ConGovAuth.LimitTimeAtAfterUnder18) > 0 {
			nowTime := time.Now()
			beforeTime := strings.Split(GetServer().ConGovAuth.LimitTimeAtBeforeUnder18, ":")
			afterTime := strings.Split(GetServer().ConGovAuth.LimitTimeAtAfterUnder18, ":")
			if len(beforeTime) >= 2 && len(afterTime) >= 2 {
				beforeTime_Hour, _ := strconv.Atoi(beforeTime[0])
				beforeTime_Minute, _ := strconv.Atoi(beforeTime[1])
				afterTime_Hour, _ := strconv.Atoi(afterTime[0])
				afterTime_Minute, _ := strconv.Atoi(afterTime[1])
				if nowTime.Hour() < beforeTime_Hour || (beforeTime_Hour == nowTime.Hour() && nowTime.Minute() <= beforeTime_Minute) ||
					nowTime.Hour() > afterTime_Hour || (afterTime_Hour == nowTime.Hour() && nowTime.Minute() >= afterTime_Minute) {
					var msg static.Msg_S2C_LimitPlayTime
					msg.Code = 3
					msg.Tips = fmt.Sprintf("亲爱的玩家，非常抱歉，根据国家相关法规规定，未成年玩家不能在晚%s-次日早%s之间登陆游戏", GetServer().ConGovAuth.LimitTimeAtAfterUnder18, GetServer().ConGovAuth.LimitTimeAtBeforeUnder18)
					newPerson.SendMsg(consts.MsgTypeLimitPlayTime, &msg)
				}
			}
		}

	case "ping":
		var _msg static.Msg_Ping
		//！更新玩家ping的时间
		self.setPingTime()
		_msg.GodTime = self.PingTime
		self.SendMsg(head, xerrors.SuccessCode, &_msg, msguid)
		//	case "closesession":
		//		self.onClose(uid)
	default:

		// 协议对象映射
		if protocolInfo := GetProtocolMgr().GetProtocol(head); protocolInfo != nil {
			person := GetPlayerMgr().GetPlayer(self.Uid)
			if person == nil {
				// 判断是否在维护
				if notice := CheckServerMaintainWithWhite(self.Uid, static.NoticeMaintainServerAllServer); notice != nil {
					// 如果有维护, 则屏蔽登录请求
					self.SendMsg(consts.MsgTypeMaintainNotice, xerrors.SuccessCode, notice, self.Uid)
					self.SafeClose(consts.SESSION_CLOED_FORCE)
					return
				}

				self.SafeClose(consts.SESSION_CLOED_CONNECT)
				return
			}
			protodata := reflect.New(protocolInfo.protoType).Interface()
			err := json.Unmarshal(data, &protodata)
			if err != nil {
				xlog.Logger().Errorln(err)
				self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
				return
			}

			begTime := time.Now().UnixNano()
			code, v := protocolInfo.protoHandler(self, &person.Info, protodata)
			endTime := time.Now().UnixNano()
			subTime := (endTime - begTime) / 1e6
			if subTime > 800 {
				xlog.Logger().Errorf("apicostwarn [costs]:%d ms \t [head]:%s \t [data]:%+v\n", subTime, head, protodata)
			}
			// 2018-12-28: 逻辑修改, 有的请求结果等待rpc异步返回
			if code != xerrors.AsyncRespErrorCode {
				self.SendMsg(head, code, v, self.Uid)
			}
			return
		} else {
			static.GetBlackIpMgr().AddIp(msgip, "消息解析错误3")
			self.SendMsg(head, xerrors.InvalidHeader, "无效的协议请求.", self.Uid)
			self.SendMsg(consts.MsgCloseConnection, xerrors.SuccessCode, &static.Msg_Null{}, msguid)
			return
		}
	}
}

// ！心跳检测
func (self *Session) checkHeartbeat() bool {
	// 没有第一个消息
	if self.Uid <= 0 {
		return true
	}
	// 没有第一个消息
	if len(self.MsgTime) <= 0 {
		return true
	}
	// 没有第一个消息
	if !self.FirstMsg {
		return true
	}
	// 超时判断
	if time.Now().Unix()-self.MsgTime[len(self.MsgTime)-1] >= socketTimeOut {
		xlog.Logger().Debug("[玩家心跳断开]:上一次心跳时间:", time.Unix(self.MsgTime[len(self.MsgTime)-1], 0).Format(static.TIMEFORMAT), "，当前时间:", time.Now().Format(static.TIMEFORMAT))
		return false
	}
	return true
}

func (self *Session) setPingTime() {
	self.PingTime = time.Now().Unix()
}

func (self *Session) closeWs() {
	go func() {
		time.Sleep(time.Second)
		self.Ws.WriteClose(0)
		self.Ws.Close()
	}()
	return
}

func (self *Session) onClose(uid int64) {
	defer func() { //ws可能已关闭
		x := recover()
		if x != nil {
			xlog.Logger().Warnln(x)
		}
	}()
	self.lockchan.Lock()
	defer self.lockchan.Unlock()
	if !atomic.CompareAndSwapInt64(&self.Stop, 0, 1) {
		return
	}
	close(self.SendChan)
	self.closeWs()
	if uid == 0 {
		if self.Uid == 0 {
			return
		}
		uid = self.Uid
	}
	GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "LastOffLineTime", time.Now().Unix())

	person := GetPlayerMgr().GetPlayer(uid)
	if person == nil {
		p, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if p == nil || err != nil {
			return
		}
		person = &PlayerCenterMemory{Info: *p, session: self}
	}

	// 清除包厢活跃数据
	hid := person.Info.HouseId
	fid := person.Info.FloorId
	if hid > 0 {
		house := GetClubMgr().GetClubHouseByHId(hid)
		if house == nil {
			return
		}
		floor := house.GetFloorByFId(fid)
		if floor == nil {
			return
		}
		go floor.SafeMemOut(uid)
	}

	if person.Info.TableId == 0 {
		GetPlayerMgr().DelPerson(uid)
	}

	// 向官方上报用户下线
	authentication.AuthenticationcollectionsUpload(person.Info.Uid, person.Info.Token, person.Info.IdcardAuthPI, util.MD5(person.Info.Nickname), authentication.AuthenticationOnlineStatusOffline)
}

// //////////////////////////////////////////////////////////////////
// ! session 管理者
const sendChanSize = 200
const socketTimeOut = 12
const socketTryNum = 10

type mapSession map[int64]*Session ///定义客户列表类型

var sessionindex int64 = 0

type SessionMgr struct {
}

var sessionmgrsingleton *SessionMgr = nil

func GetSessionMgr() *SessionMgr {
	if sessionmgrsingleton == nil {
		sessionmgrsingleton = new(SessionMgr)
	}

	return sessionmgrsingleton
}

func (self *SessionMgr) GetNewSession(ws *websocket.Conn) *Session {
	sessionindex += 1
	session := new(Session)
	session.ID = sessionindex
	session.Ws = ws
	session.SendChan = make(chan []byte, sendChanSize)
	session.LineTime = time.Now().Unix()
	session.FirstMsg = false
	session.lockchan = new(lock2.RWMutex)
	session.IP = static.HF_GetHttpIP(session.Ws.Request())
	session.Stop = 0

	return session
}

// 获取person信息
func getPersonInfo(person *PlayerCenterMemory) *static.Msg_S2C_Person {
	p := new(static.Msg_S2C_Person)
	p.Uid = person.Info.Uid
	p.Nickname = person.Info.Nickname
	p.Imgurl = person.Info.Imgurl
	p.Card = person.Info.Card
	p.Gold = person.Info.Gold
	p.Diamond = person.Info.Diamond
	p.GoldBean = person.Info.GoldBean
	p.InsureGold = person.Info.InsureGold
	p.HId = person.Info.HouseId
	if GetServer().ConHouse.DefaultHouse > 0 {
		p.HId = GetServer().ConHouse.DefaultHouse
	}
	p.FId = person.Info.FloorId
	p.GameId = person.Info.GameId
	p.SiteId = person.Info.SiteId
	p.TableId = person.Info.TableId
	p.Area = person.Info.Area
	p.Sex = person.Info.Sex
	p.Ip = person.Ip
	p.Games = person.Info.Games
	p.DescribeInfo = person.Info.DescribeInfo
	p.RealName, _ = static.HF_DecodeStr(static.HF_Atobytes(person.Info.ReName), static.UserEncodeKey)
	p.Tel, _ = static.HF_DecodeStr(static.HF_Atobytes(person.Info.Tel), static.UserEncodeKey)
	p.Idcard, _ = static.HF_DecodeStr(static.HF_Atobytes(person.Info.Idcard), static.UserEncodeKey)
	p.UserType = person.Info.UserType
	p.DeliveryImg = person.Info.DeliveryImg
	p.ContributionScore = person.Info.ContributionScore
	str, _ := GetDBMgr().GetDBrControl().RedisV2.Get(HotVersionKey).Result()
	p.HotVersion = str
	// 判断是否实名认证
	if person.Info.Idcard != "" && person.Info.IdcardAuthPI != "" {
		p.Certification = true
	}
	// 传奇参数信息（部分c++用户需要保留迁移前的传奇数据）
	p.ChuanQiParam = GetDBMgr().GetDBrControl().RedisV2.Get(fmt.Sprintf(consts.REDIS_KEY_USER_CHUANQI_INFO, p.Uid)).Val()
	p.RefuseInvite = person.Info.RefuseInvite
	return p
}

// 解密消息
func decryptData(msg []byte, encode int) ([]byte, error) {
	var data []byte
	var err error
	switch encode {
	case consts.EncodeNone: // 不加密不处理
		data = msg
	case consts.EncodeAes: // aes + base64
		data, err = base64.URLEncoding.DecodeString(string(msg))
		if err != nil {
			return nil, err
		}

		data, err = goEncrypt.AesCTR_Decrypt(data, []byte(GetServer().Con.EncodeClientKey))
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
