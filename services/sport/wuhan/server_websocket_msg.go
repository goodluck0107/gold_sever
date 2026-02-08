package wuhan

import (
	"encoding/base64"
	"encoding/json"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/open-source/game/chess.git/pkg/xlog"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	"io"
	"net"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"

	"fmt"

	"golang.org/x/net/websocket"
)

type Msg_ClientLog struct {
	Log string `json:"log"`
}

type Session struct {
	ID       int64           //! sessionID
	Uid      int64           //! 用户id
	Ws       *websocket.Conn //! websocket
	IP       string          //! 客户端地址
	SendChan chan []byte     //! 消息管道
	LineTime int64           //! 链接时间
	FirstMsg bool            //! 是否有第一个消息
	MsgTime  []int64         //! 时间切片
	TryNum   int             // 尝试次数
	// RecvTime map[string]time.Time //！收到消息时间
	PingTime int64 //！最新心跳时间

	lockchan *lock2.RWMutex
}

//! 发送消息
func (self *Session) SendMsg(head string, errCode int16, v interface{}, uid int64) {
	if head == "" {
		return
	}

	self.lockchan.CustomLock()
	defer self.lockchan.CustomUnLock()

	if self.SendChan == nil {
		return
	}

	if self == nil {
		return
	}

	buf := static.HF_EncodeMsg(head, errCode, v, GetServer().Con.Encode, GetServer().Con.EncodeClientKey, uid)
	select {
	case self.SendChan <- buf:
		if len(self.SendChan) > sendChanSize-10 && self.Uid > 0 {
			self.SendChan <- []byte("")
		}
		return
	case <-time.After(50 * time.Millisecond):
		xlog.Logger().Errorf("send msg to chan error. chan is nil:%v, chan len:%v ,val:%v", self.SendChan == nil, len(self.SendChan), string(buf))
		return
	}

}

func (self *Session) SafeClose(origin uint8) {

	switch origin {
	case consts.SESSION_CLOED_FORCE:
		self.SendMsg(consts.MsgForceCloseConnection, xerrors.SuccessCode, "", self.Uid)
	case consts.SESSION_CLOED_ASYSTOLE:
		self.SendMsg(consts.MsgForceCloseConnection, xerrors.SuccessCode, "", self.Uid)
	default:
		self.SendMsg(consts.MsgCloseConnection, xerrors.SuccessCode, &static.Msg_Null{}, self.Uid)
	}

	self.lockchan.CustomLock()
	defer self.lockchan.CustomUnLock()

	if self.SendChan != nil {
		select {
		case self.SendChan <- []byte(""):
			return
		case <-time.After(50 * time.Millisecond):
			xlog.Logger().Errorf("safeclose chan error. chan is nil:%v, chan len:%v", self.SendChan == nil, len(self.SendChan))
			return
		}
	}

	return
}

func (self *Session) CloseChan() {
	self.lockchan.CustomLock()
	defer self.lockchan.CustomUnLock()

	if self.SendChan == nil {
		return
	}

	close(self.SendChan)
	self.SendChan = nil
}

//! 消息run
func (self *Session) Run() {
	go self.sendMsgRun()
	self.receiveMsgRun()
}

//! 发送消息循环
func (self *Session) sendMsgRun() {
	for msg := range self.SendChan {
		if GetServer().ShutDown {
			break
		}

		if string(msg) == "" {
			// 连接断开异常处理
			if self.Uid != 0 {
				// onlyExitGame := fmt.Sprintf("userstatus_doing_exitgame_%d", self.Uid)
				// GetDBMgr().Redis.SetNX(onlyExitGame, self.Uid, time.Second*3)
			}
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
						continue
					}
				}
				xlog.Logger().Errorln("sendmsgfail uid:", self.Uid, "-err:", err.Error())
				xlog.Logger().Errorln("sendmsgfail uid:", self.Uid, "-msg:", static.HF_Bytestoa(msg))
				exit = true
				break
			} else {
				//发送成功
				break
			}
		}
		if exit {
			break
		}
	}
	xlog.Logger().Warningln("session client close by wuhan")
	self.Ws.Close()
}

//! 消息循环
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
				xlog.Logger().Warningln("session client disconnet")
			} else {
				xlog.Logger().Warningln("session receive err:", err)
			}

			// 连接断开异常处理
			//if self.Uid != 0 {
			//	onlyExitGame := fmt.Sprintf("userstatus_doing_exitgame_%d", self.Uid)
			//	GetDBMgr().Redis.SetNX(onlyExitGame, self.Uid, time.Second*3)
			//}

			break
		}

		if len(self.MsgTime) >= 10 { //!
			self.MsgTime = make([]int64, 0)
		}
		self.MsgTime = append(self.MsgTime, time.Now().Unix())

		self.onReceive(msg)
	}
	xlog.Logger().Warningln("session client close", self.Uid)
	self.CloseChan()
	self.Ws.Close()
	self.onClose(self.Uid)
}

func (self *Session) onClose(uid int64) {
	if uid == 0 {
		return
	}

	// 连接断开异常处理
	//onlyExitGame := fmt.Sprintf("userstatus_doing_exitgame_%d", uid)
	//GetDBMgr().Redis.Del(onlyExitGame)

	//! 通知该玩家掉线
	// 找不到用户, 用户已经离桌
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return
	}
	if person.session == self || person.session == nil {

		//! 离线消息
		table := GetTableMgr().GetTable(person.Info.TableId)
		if table != nil && table.game != nil {
			//! 通知大厅用户离线
			GetServer().CallHall("NewServerMsg", consts.MsgTypeUserOffline, xerrors.SuccessCode, &static.Msg_Uid{Uid: uid, Fid: table.FId}, uid)
			table.game.OnLine(person.Info.Uid, false)
		} else {
			//! 通知大厅用户离线
			GetServer().CallHall("NewServerMsg", consts.MsgTypeUserOffline, xerrors.SuccessCode, &static.Msg_Uid{Uid: uid}, uid)
		}
	}

	//! 标志为离线
	self = nil
}

func (self *Session) onReceive(msg []byte) {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	head, sign, _, data, ok, msguid, _ := static.HF_DecodeMsg(msg)
	//!解析错误
	if !ok {
		xlog.Logger().Errorln("session onReceive decodemsg err")
		return
	}

	// 消息uid必须>0
	//if msguid <= 0 {
	//	self.SendMsg(head, xerrors.TokenError.Code, xerrors.TokenError.Msg, msguid)
	//	self.SafeClose(constant.SESSION_CLOED_FORCE)
	//	return
	//}

	// 第一条消息必须是入座
	if !self.FirstMsg {
		if head != consts.MsgTypeTableIn {
			self.SendMsg(head, xerrors.TokenError.Code, xerrors.TokenError.Msg, msguid)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}
		self.FirstMsg = true
	}

	// 如果使用加密了但客户端消息未加密则直接报错
	if GetServer().Con.Encode != consts.EncodeNone && sign.Encode == consts.EncodeNone {
		self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
		return
	}

	// 消息解密处理
	var err error
	data, err = decryptData(data, sign.Encode)
	if err != nil {
		xlog.Logger().Errorln("Service decrypt err:", err.Error())
		self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
		return
	}

	if head != "ping" { //屏蔽掉ping的log
		xlog.Logger().WithFields(logrus.Fields{
			"uid":  self.Uid,
			"head": head,
			"data": string(data),
		}).Infoln("【RECEIVED SOCKET】")
	}

	switch head {
	case consts.MsgTypeTableIn: //! 加入牌桌
		var msg static.Msg_Game_TableIn
		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, msguid)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 参数校验
		if msg.Id == 0 {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 关掉老连接
		person := GetPersonMgr().GetPerson(msg.Uid)
		if person != nil && person.session != nil {
			person.SendMsg(consts.MsgTypeRelogin, xerrors.SuccessCode, nil)
			person.CloseSession(consts.SESSION_CLOED_FORCE)
		} else if person == nil {
			person = GetPersonMgr().GetLookonPerson(msg.Uid)
			if person != nil && person.session != nil {
				person.SendMsg(consts.MsgTypeRelogin, xerrors.SuccessCode, nil)
				person.CloseSession(consts.SESSION_CLOED_FORCE)
			}
		}

		// 如果加入失败 关闭连接
		if !self.TableConIn(head, &msg) {
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

	case consts.MsgTypeGetUserInfo: //获取用户详情
		var msg static.Msg_Uid
		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			return
		}

		if msg.Uid <= 0 {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			return
		}

		person := GetPersonMgr().GetPerson(self.Uid)
		if person == nil {
			return
		}

		result := &static.Site_SC_UserInfo{
			Uid:        person.Info.Uid,
			Imgurl:     person.Info.Imgurl,
			Nickname:   person.Info.Nickname,
			Sex:        person.Info.Sex,
			Gold:       person.Info.Gold,
			WinCount:   person.Info.WinCount,
			LostCount:  person.Info.LostCount,
			DrawCount:  person.Info.DrawCount,
			FleeCount:  person.Info.FleeCount,
			TotalCount: person.Info.TotalCount,
		}
		self.SendMsg(head, 0, result, 0)
	case consts.MsgTypeSiteIn: //! 加入场次
		var msg static.Msg_Game_SiteIn
		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 参数校验
		if msg.KindId == 0 || msg.SiteType == 0 {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 如果加入失败关闭连接
		if !self.SiteConIn(head, &msg) {
			self.SafeClose(consts.SESSION_CLOED_FORCE)
		}
	case consts.MsgTypeSiteListIn: // 加入列表
		var msg static.Msg_Game_SiteListIn
		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		p := GetPersonMgr().GetPerson(self.Uid)
		if p == nil {
			self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 判断场次是否存在
		site := GetSiteMgr().GetSite(p.Info.SiteId) //GetSiteMgr().GetSiteByType(msg.Type, msg.KindId)
		if site == nil {                            //! 场次不存在
			xlog.Logger().Errorln("session,tableconin,err:table not exit:")
			self.SendMsg(head, xerrors.SiteNotExistError.Code, xerrors.SiteNotExistError.Msg, p.Info.Uid)
			return
		}

		// 参数校验
		if msg.Start >= msg.End || msg.Start < 0 || msg.End > site.tableNum {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			return
		}

		site.Operator(NewSiteMsg(head, "", p.Info.Uid, &msg))
	case consts.MsgTypeSiteListOut: // 退出列表
		p := GetPersonMgr().GetPerson(self.Uid)
		if p == nil {
			self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 判断场次是否存在
		site := GetSiteMgr().GetSite(p.Info.SiteId) //GetSiteMgr().GetSiteByType(msg.Type, msg.KindId)
		if site == nil {                            //! 场次不存在
			xlog.Logger().Errorln("session,tableconin,err:table not exit:")
			self.SendMsg(head, xerrors.SiteNotExistError.Code, xerrors.SiteNotExistError.Msg, p.Info.Uid)
			return
		}
		site.Operator(NewSiteMsg(head, "", p.Info.Uid, nil))
	case "sitetablein": //! 入桌
		var msg static.Msg_Game_SiteInTable

		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		self.SiteConInTable(head, &msg)
	case "changetablein": //! 换桌
		var msg static.Msg_Game_ChangeInTable
		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 如果加入失败关闭连接
		if !self.ChangeConInTable(head, &msg) {
			self.SafeClose(consts.SESSION_CLOED_FORCE)
		}
	case "ping":
		var msg static.Msg_Null
		self.setPingTime()
		self.SendMsg(head, xerrors.SuccessCode, &msg, self.Uid)

	case "closesession":
		self.onClose(self.Uid)

	case consts.MsgTypeTableExit: // 离开牌桌
		GetDBMgr().Redis.SetNX(fmt.Sprintf("userstatus_doing_exit_%d", self.Uid), 1, time.Second*3)
		table := self.GetTable(self.Uid)
		if table != nil {
			table.Operator(base2.NewTableMsg(consts.MsgTypeTableExit, head, self.Uid, nil))
		}
	case consts.MstTypeWatcherQuit:
		var msg static.Msg_S_WatcherQuit
		err := json.Unmarshal(data, &msg)
		if err == nil {
			GetServer().CallHall("NewServerMsg", consts.MsgTypeTableWatchOut_Ntf, xerrors.SuccessCode, &msg, 0)
		}
		//房间服务器删除观战列表
		table := self.GetLookonTable(self.Uid)
		if table != nil {
			table.Operator(base2.NewTableMsg(consts.MstTypeWatcherQuit, head, self.Uid, nil))
		}
		//通知客户端退出观战
		self.SendMsg(consts.MstTypeWatcherQuit, xerrors.SuccessCode, nil, self.Uid)

	case consts.MstTypeWatcherSwitch:
		//客户端切换旁观位置
		var msg static.Msg_S_WatcherSwitch
		err := json.Unmarshal(data, &msg)
		if err != nil {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, msguid)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}

		// 参数校验
		if msg.Uid == 0 || msg.TableId == 0 {
			// 参数错误
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, 0)
			self.SafeClose(consts.SESSION_CLOED_FORCE)
			return
		}
		//客户端切换旁观位置
		table := self.GetLookonTable(self.Uid)
		if table != nil {
			//通知客户端切换旁观位置
			self.SendMsg(consts.MstTypeWatcherSwitch, xerrors.SuccessCode, msg, self.Uid)
			table.Operator(base2.NewTableMsg(consts.MstTypeWatcherSwitch, head, self.Uid, msg))
		}

	case consts.MstTypeWatcherList:
		//客户端申请旁观列表  服务器支持游戏玩家和旁观玩家申请旁观列表
		table := self.GetLookonTable(self.Uid)
		if table == nil {
			table = self.GetTable(self.Uid)
		}
		if table != nil {
			table.Operator(base2.NewTableMsg(consts.MstTypeWatcherList, head, self.Uid, nil))
		}
	default:
		person := GetPersonMgr().GetPerson(self.Uid)
		if person == nil {
			xlog.Logger().Errorln("head:", head, "session person not exit err uid:", self.Uid)
			return
		}
		table := GetTableMgr().GetTable(person.Info.TableId)
		if table != nil {
			var msg static.Msg_Null
			table.Operator(base2.NewTableMsg(head, static.HF_Bytestoa(data), self.Uid, &msg))
		}
	}
}

func (self *Session) setPingTime() {
	//syslog.Logger().Debug("设置当前pingtime：", self.PingTime, "-->", time.Now().Unix())
	self.PingTime = time.Now().Unix()
}

//! 加入牌桌
func (self *Session) TableConIn(head string, msg *static.Msg_Game_TableIn) bool {

	//获取用户，创建用户
	dperson, err := GetDBMgr().db_R.GetPerson(msg.Uid)
	if err != nil {
		self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, msg.Uid)
		return false
	}

	// 校验token
	if dperson.Token != msg.Token {
		self.SendMsg(head, xerrors.TokenError.Code, xerrors.TokenError.Msg, msg.Uid)
		return false
	}

	// 更新玩家弱数据
	UpdateLazyUser(msg.Uid, dperson.Nickname, dperson.Sex, dperson.Imgurl)

	// 判断牌桌是否存在
	table := GetTableMgr().GetTable(msg.Id)
	if table == nil { //! 牌桌不存在
		if gameinfo, error := GetDBMgr().GetDBrsControl().GetLastGameInfo(msg.Uid); error == nil {
			if gameinfo.TId == msg.Id {
				//其它游戏用这个,消息太大了需要拆分发送
				var gameinfoData static.LastGameInfo
				gameinfoData = *gameinfo
				gameinfoData.GameendInfo = ""
				gameinfoData.ResultTotalInfo = ""
				self.SendMsg(consts.MsgTypeTableInLastGame, xerrors.LastGameCardsInfo.Code, gameinfoData, msg.Uid)
				gameinfoData = *gameinfo
				gameinfoData.CardsDataInfo = ""
				gameinfoData.PlayersInfo = []static.LastGamePlayersInfo{} //不重复发送节省空间
				self.SendMsg(consts.MsgTypeTableInLastGame, xerrors.LastGameEndInfo.Code, gameinfoData, msg.Uid)
				//仙桃一赖到底还是用下面的消息
				self.SendMsg(head, xerrors.LastGameInfo.Code, gameinfo, msg.Uid)
				return false
			}
		}
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(msg.Uid, "TableId", 0, "GameId", 0)
		self.SendMsg(head, xerrors.TableIdError.Code, xerrors.TableIdError.Msg, msg.Uid)
		return false
	}

	// 标记session tag
	self.Uid = msg.Uid

	//！玩家登录成功后，认为第一次ping成功
	self.setPingTime()

	person := new(PersonGame)
	person.Info = *dperson
	person.session = self
	person.Ip = dperson.Ip

	if msg.MInfo != "" && msg.MInfo != "null" {
		err := json.Unmarshal([]byte(msg.MInfo), &person.Minfo)
		if err != nil {
			self.SendMsg(head, xerrors.ArgumentError.Code, xerrors.ArgumentError.Msg, msg.Uid)
			return false
		}
		//GetDBMgr().GetDBrControl().UpdatePersonAttrs(msg.Uid, "Address", person.Minfo.Address,
		//	"Longitude", person.Minfo.Longitude, "Latitude", person.Minfo.Latitude)
	}
	GetDBMgr().GetDBrControl().UpdatePersonAttrs(msg.Uid, "Online", true)

	if person.Info.WatchTable == 0 {
		GetPersonMgr().AddPerson(person)
		table.Operator(base2.NewTableMsg(head, "", person.Info.Uid, person))
	} else {
		GetPersonMgr().AddLookonPerson(person)
		table.Operator(base2.NewTableMsg(head, "lookon", person.Info.Uid, person))
	}
	return true
}

//! 获取牌桌
func (self *Session) GetTable(uid int64) *Table {
	if uid == 0 {
		return nil
	}
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return nil
	}
	table := GetTableMgr().GetTable(person.Info.TableId)
	return table
}

//! 获取牌桌
func (self *Session) GetLookonTable(uid int64) *Table {
	if uid == 0 {
		return nil
	}
	person := GetPersonMgr().GetLookonPerson(uid)
	if person == nil {
		return nil
	}
	table := GetTableMgr().GetTable(person.Info.WatchTable)
	return table
}

//! 加入金币场场次
func (self *Session) SiteConIn(head string, msg *static.Msg_Game_SiteIn) bool {
	//获取用户，创建用户
	person := new(PersonGame)
	dperson, err := GetDBMgr().db_R.GetPerson(msg.Uid)
	if err != nil {
		xlog.Logger().Errorln("session,tableconin,err:", err.Error())
		self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, 0)
		return false
	}
	person.Info = *dperson

	// 校验token
	if person.Info.Token != msg.Token {
		xlog.Logger().Debug("session,tableconin,err:tokenerr:", msg.Token)
		self.SendMsg(head, xerrors.TokenError.Code, xerrors.TokenError.Msg, 0)
		return false
	}
	self.Uid = msg.Uid

	// 判断场次是否存在
	site := GetSiteMgr().GetSiteByType(msg.KindId, msg.SiteType)
	if site == nil { //! 场次不存在
		xlog.Logger().Errorln("session,tableconin,err:table not exit:")
		self.SendMsg(head, xerrors.TableIdError.Code, xerrors.TableIdError.Msg, person.Info.Uid)
		return false
	}

	// 如果用户在该场次的话 无需判断人数上限 直接进入场次(断线重连)
	if site.Person[self.Uid] == nil && len(site.Person) >= site.MaxPeopleNum {
		xlog.Logger().Errorln(fmt.Sprintf("场次[site_id: %d]人数已满", site.Id))
		self.SendMsg(head, xerrors.SiteIsFullError.Code, xerrors.SiteIsFullError.Msg, person.Info.Uid)
		return false
	}

	person.session = self
	person.Ip = self.IP
	if msg.MInfo != "" {
		err = json.Unmarshal([]byte(msg.MInfo), &person.Minfo)
		xlog.Logger().Debug("session,tableconin,minfo :", msg.MInfo)
	}

	GetPersonMgr().AddPerson(person)

	site.Operator(NewSiteMsg(head, "", person.Info.Uid, person))
	return true
}

//! 加入金币场场次
func (self *Session) SiteConInTable(head string, msg *static.Msg_Game_SiteInTable) bool {
	//获取用户，创建用户
	person := GetPersonMgr().GetPerson(msg.Uid)
	if person == nil {
		xlog.Logger().Errorln("session,UserNotExistError,err:")
		self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, 0)
		return false
	}

	// 判断场次是否存在
	site := GetSiteMgr().GetSite(person.Info.SiteId) //GetSiteMgr().GetSiteByType(msg.Type, msg.KindId)
	if site == nil {                                 //! 场次不存在
		xlog.Logger().Errorln("session,tableconin,err:table not exit:")
		self.SendMsg(head, xerrors.TableIdError.Code, xerrors.TableIdError.Msg, person.Info.Uid)
		return false
	}
	person.session = self
	person.Ip = self.IP

	site.Operator(NewSiteMsg(head, "", person.Info.Uid, msg))
	return true
}

//! 换桌
func (self *Session) ChangeConInTable(head string, msg *static.Msg_Game_ChangeInTable) bool {
	//获取用户，创建用户
	person := GetPersonMgr().GetPerson(msg.Uid)
	if person == nil {
		xlog.Logger().Errorln("session,UserNotExistError,err:")
		self.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg, 0)
		return false
	}

	// 判断场次是否存在
	site := GetSiteMgr().GetSite(person.Info.SiteId)
	if site == nil { //! 场次不存在
		xlog.Logger().Errorln("session,tableconin,err:table not exit:")
		self.SendMsg(head, xerrors.TableIdError.Code, xerrors.TableIdError.Msg, person.Info.Uid)
		return false
	}

	site.Operator(NewSiteMsg(head, "", person.Info.Uid, msg))
	return true
}

//！心跳检测
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
	if time.Now().Unix()-self.MsgTime[len(self.MsgTime)-1] >= socketheartbeat {
		xlog.Logger().Debug("[玩家心跳断开]玩家ID:", self.Uid, ":上一次心跳时间:", time.Unix(self.MsgTime[len(self.MsgTime)-1], 0).Format(static.TIMEFORMAT), "，当前时间:", time.Now().Format(static.TIMEFORMAT))
		return false
	}
	return true
}

////////////////////////////////////////////////////////////////////
//! session 管理者

//! 消息管道
const sendChanSize = 200
const socketTimeOut = 5
const socketheartbeat = 12
const socketTryNum = 10

type mapSession map[int64]*Session ///客户端会话表

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

	return session
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
