package wuhan

import (
	"encoding/json"
	"errors"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

const (
	MAX_TABLE_ID = 999999
	MIN_TABLE_ID = 100000

	ForceCloseByOwner          = 1 // 牌桌被房主解散
	ForceCloseByEmpty          = 2 // 空桌解散
	forceCloseByPanic          = 3 // 游戏异常解散
	forceCloseByTimeout        = 4 // 长时间未操作解散牌桌
	forceCloseByHouseManager   = 5 // 牌桌被包厢盟主/管理员强制解散
	forceCloseByServerMaintain = 6 // 服务器维护, 强制解散房间
	forceCloseByServerGM       = 7 // 牌桌被GM强制解散
	forceCloseBySchedule       = 8 // 牌桌定时解散
	forceCloseByKick           = 9 // 被踢了
)

//! 玩家上线下线
type Msg_LinePerson struct {
	Uid  int64 `json:"uid"`
	Line bool  `json:"line"`
}

/////////////////////////////////////////////////
//! 加入牌桌失败
type Msg_JoinTableFail struct {
	Result int `json:"result"`
}

//! 退出牌桌
type Msg_ExitTable struct {
	Uid int64 `json:"uid"`
}

//! 牌桌
type Table struct {
	*static.Table
	Person       []*base2.TablePerson   `json:"-"` // 牌桌上的人
	LookonPerson [][]*base2.TablePerson `json:"-"` // 牌桌上的旁观者
	LiveTime     int64                  `json:"-"` // 活动时间
	DelTime      int64                  `json:"-"` // 解散时间
	reciveChan   chan *base2.TableMsg   `json:"-"` // 操作队列
	game         base2.SportInterface   `json:"-"` // 游戏
	tableLog     *xlog.TableLoger       `json:"-"` // 桌子日志
	SubFloor     *redis.PubSub          `json:"-"` // 订阅楼层信息
	CloseChan    chan struct{}
	shutdown     bool //关闭消息接受
}

//! 恢复牌桌数据
func RestoreTables() {
	// 获取所有table key
	keys, err := GetDBMgr().db_R.Keys(fmt.Sprintf(consts.REDIS_KEY_TABLEINFO_ALL_2, GetServer().Con.Id))
	if err != nil {
		xlog.Logger().Errorln("restore hall table failed: ", err.Error())
		return
	}

	for _, key := range keys {
		data, err := GetDBMgr().db_R.Get(key)
		if err != nil {
			xlog.Logger().Errorln("get game tableinfo failed: ", err.Error())
			continue
		}
		var table Table
		err = json.Unmarshal(data, &table)
		if err != nil {
			xlog.Logger().Errorln("unmarshal game tableinfo failed: ", err.Error())
			continue
		}

		flag := true
		for _, p := range table.Person {
			if p == nil {
				continue
			}

			person, err := GetDBMgr().db_R.GetPerson(p.Uid)
			if err != nil {
				flag = false
				xlog.Logger().Errorln("person not exists: ", p.Uid)
				break
			}

			gp := new(PersonGame)
			gp.Info = *person
			// 添加person
			GetPersonMgr().AddPerson(gp)
		}

		if flag && table.Config.GameType == static.GAME_TYPE_FRIEND {
			// 添加桌子
			table.Init()
			GetTableMgr().AddTable(&table)
		} else {
			xlog.Logger().Errorln("table data error: ", table.Id)
		}
	}
}

func (self *Table) Init() {
	// 获取游戏配置
	gameName := ""
	config := GetServer().GetGameConfig(self.KindId)
	if config != nil {
		gameName = config.EnName
	} else {
		gameName = "UnknownGame"
	}
	self.tableLog = xlog.CreateTableLog(GetServer().Index, self.Id, gameName, self.GameNum, self.HId, self.NFId, self.DHId, self.FId)
	self.Person = make([]*base2.TablePerson, self.Table.Config.MaxPlayerNum)
	self.LookonPerson = make([][]*base2.TablePerson, 10) //最多10个玩家，每个玩家最多有self.Table.Config.MaxLookonPlayerNum个旁观者
	self.reciveChan = make(chan *base2.TableMsg, 2000)
	self.LiveTime = time.Now().Unix()
	self.shutdown = false
	self.IsAnonymous = static.IsAnonymous(self.Config.GameConfig) //匿名游戏

	err := true
	if err, self.game = base2.CreateGameFunc(self.KindId); !err { //没有对应的玩法
		return
	}

	self.game.OnInit(self)
	go self.run()
	// 只有好友房才自动解散牌桌
	if self.Config.GameType == static.GAME_TYPE_FRIEND {
		go self.runLive()
	}
	if self.DelTime != 0 { //! 解散时间不为0
		go self.dissmissThread()
	}
}

func (self *Table) GetTableInfo() *static.Table {
	return self.Table
}

//! 发送操作
func (self *Table) Operator(op *base2.TableMsg) {
	if self.reciveChan != nil && !self.shutdown {
		if op.Head != consts.MsgTypeTableCheckConn {
			xlog.Logger().Debug(fmt.Sprintf("table: %d, op: %+v", self.Id, op))
		}
		self.reciveChan <- op
	}
}

func (self *Table) OnBegin() {
	self.Begin = true

	// 通知大厅
	GetServer().CallHall("NewServerMsg", consts.MsgTtypeTableSetBegin, xerrors.SuccessCode, &static.Msg_GH_SetBegin{TableId: self.Id, Begin: true, Step: self.Step, Fid: self.FId, Hid: self.HId}, 0)
}

func (self *Table) SetXiaPaoIng(xiapaoing bool) {
	self.BXiaPaoIng = xiapaoing
}

//! 查询游戏是否开始
func (self *Table) IsBegin() bool {
	return self.Begin
}

//! 设置开始结束
func (self *Table) SetBegin(begin bool) {
	self.Begin = begin
	if begin {
		self.Step++
		if self.IsTeaHouse() {

		}
	}

	if self.Config.GameType == static.GAME_TYPE_FRIEND {
		// 通知大厅
		GetServer().CallHall("NewServerMsg", consts.MsgTtypeTableSetBegin, xerrors.SuccessCode, &static.Msg_GH_SetBegin{TableId: self.Id, Begin: begin, Step: self.Step, Fid: self.FId, Hid: self.HId}, 0)
	} else if self.Config.GameType == static.GAME_TYPE_GOLD {
		// 重新生成gamenum
		nowTime := time.Now()
		self.GameNum = fmt.Sprintf("%d%02d%02d%02d%02d%02d_%d_%d", nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), nowTime.Minute(), nowTime.Second(), GetServer().Con.Id, self.Id)
		self.NotifyTableChange()
	}
}

//! 牌桌结束
func (self *Table) IsBye() bool {
	return false
}

//! 是否一局都未开始
func (self *Table) IsNew() bool {
	return !self.Begin && self.Step == 0
}

//! 是否一局都未开始
func (self *Table) IsXiaPaoIng() bool {
	return self.BXiaPaoIng
}

func (self *Table) run() {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
			// 如果在解散时panic再解散处理会出错, 先屏蔽
			// msg := new(public.Msg_S2C_TableDel)
			// msg.Type = forceCloseByPanic
			// msg.Msg = fmt.Sprintf("游戏服运行出错,id:%s", GetServer().Index)
			// self.game.OnEnd()
			// self.Clear(msg)
			// 向go服务器钉钉群推送一个消息
			xlog.Logger().Errorln(
				service2.GetDingTalkRobotClient().SendText(
					"gogroup",
					fmt.Sprintf(
						"chess游戏服发现一个bug！\n[gameid:%s][kindid:%d][tableid:%d]\n#ERROR:#:%v\n%s",
						GetServer().Index,
						self.KindId,
						self.Id,
						x,
						string(debug.Stack())),
					true,
				),
			)

			self.clear(&static.Msg_S2C_TableDel{Type: forceCloseByPanic, Msg: "牌桌异常，强制解散。"})
			if self.Config.GameType == static.GAME_TYPE_FRIEND {
				// 通知大厅牌桌解散
				var msg static.GH_TableDel_Ntf
				msg.TableId = self.Id
				msg.Fid = self.FId
				msg.Hid = self.HId
				PushTableStatusMsg(msg.Hid, consts.MsgTypeTableDel_Ntf, msg)
				// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableDel_Ntf, xerrors.SuccessCode, &msg, 0)
			}

		}
	}()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if self.reciveChan == nil {
				return
			}
			self.game.OnTime()
		case op, ok := <-self.reciveChan:
			if !ok {
				continue
			}
			begTime := time.Now().UnixNano()
			xlog.Logger().Infof("【table op: %s】", op.Head)
			if !self.onMsg(op) {
				ticker.Stop()
				return
			}
			endTime := time.Now().UnixNano()
			subTime := (endTime - begTime) / 1e6
			if subTime > 300 {
				xlog.Logger().Errorf("[costs]:%d \t [head]:%s \t [data]:%+v\n", subTime, op.Head, op)
			}
		}
	}
}

//消息处理
func (self *Table) onMsg(op *base2.TableMsg) bool {

	switch op.Head {
	case consts.MsgTypeTableIn: //! 建立连接和加入(session)
		if op.Data == "lookon" {
			self.LookonTablein(op.V.(*PersonGame))
		} else {
			self.Tablein(op.V.(*PersonGame))
		}
	case consts.MsgTypeTableCheckConn: // 若游戏未开始，检测用户连接状态
		v := op.V.(*static.Msg_CheckUserConnection)
		self.CheckUserConnection(v.Uid, v.Seat)
	case consts.MsgTypeTableExit: // 退出牌桌
		// 正常退出
		//ok := self.TableExit(op.Uid)

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			if op.Data == "lookon" {
				self.LookonTableExit(op.Uid, op.Uid)
				self.game.OnStandupLookon(op.Uid, op.Uid)
			} else {
				if !self.Begin {
					self.game.OnStandup(op.Uid, op.Uid)
				} else {
					xlog.Logger().Error("游戏开始不能推出房间")
				}
			}
		} else {
			self.game.OnStandup(op.Uid, op.Uid)
		}
	case consts.MstTypeWatcherQuit: // 退出观战
		// 正常退出
		//ok := self.TableExit(op.Uid)

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			self.LookonTableExit(op.Uid, op.Uid)
			self.game.OnStandupLookon(op.Uid, op.Uid)
		} else {
			self.game.OnStandup(op.Uid, op.Uid)
		}
	case consts.MstTypeWatcherList: // 观战列表
		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			self.game.SendWatcherList(op.Uid)
		}
	case consts.MstTypeWatcherSwitch: // 切换观战位置
		v := op.V.(static.Msg_S_WatcherSwitch)
		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			self.LookonSeatSwitch(v.Uid, v.Seat)
		}
	case consts.MsgTypeTableDel: //! 解散牌桌
		//移除通知大厅解散桌子
		if op.V == nil { //游戏正常结算时候解散
			self.clear(nil)

			if self.Config.GameType == static.GAME_TYPE_FRIEND {
				// 通知大厅牌桌解散
				var msg static.GH_TableDel_Ntf
				msg.TableId = self.Id
				msg.Fid = self.FId
				msg.Hid = self.HId
				PushTableStatusMsg(msg.Hid, consts.MsgTypeTableDel_Ntf, &msg)
				// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableDel_Ntf, xerrors.SuccessCode, &msg, 0)

			}
		} else {
			msg, ok := op.V.(*static.Msg_S2C_TableDel)
			if !ok {
				// ticker.Stop()
				return false
			}
			self.game.OnEnd()
			self.clear(msg)
			if self.Config.GameType == static.GAME_TYPE_FRIEND {
				// 通知大厅牌桌解散
				if msg.Type != forceCloseByHouseManager {
					var msg static.GH_TableDel_Ntf
					msg.TableId = self.Id
					msg.Fid = self.FId
					msg.Hid = self.HId
					PushTableStatusMsg(msg.Hid, consts.MsgTypeTableDel_Ntf, &msg)
					// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableDel_Ntf, xerrors.SuccessCode, &msg, 0)
				}
			}
		}
		//ticker.Stop()
		return false
	case "broadcast": //! 广播
		self.LiveTime = time.Now().Unix()
		self.broadCastMsg(op.Data, xerrors.SuccessCode, op.V)
	case "usersubfloor":
		self.LiveTime = time.Now().Unix()
		self.UserSubFloorMsg(op.Uid, op.Data)
	case "userunsubfloor":
		self.LiveTime = time.Now().Unix()
		self.UserUnSubFloorMsg(op.Uid)
	case "floortableinfo":
		self.LiveTime = time.Now().Unix()
		self.FloorTableInfoPush(op.Uid)
	case "housememonline":
		self.LiveTime = time.Now().Unix()
		self.HouseOnLine(op.Uid)
	default:
		self.LiveTime = time.Now().Unix()
		if self.IsAnswerRequest(op) {
			// syslog.Logger().Debugf("【game op: %+v】", op)
			self.game.OnBaseMsg(op)
		}
	}

	return true
}

// 是不是可以响应的请求
func (self *Table) IsAnswerRequest(op *base2.TableMsg) bool {
	if self.game.IsPausing() {
		switch op.Head {
		case
			consts.MsgCommonToGameContinue,        // 游戏继续
			consts.MsgTypeGameDismissFriendRep,    // 解散房间声请
			consts.MsgTypeGameDismissFriendReq,    // 解散房间声请
			consts.MsgTypeGameDismissFriendResult, // 解散房间响应
			consts.MsgTypeGameinfo,                // 请求游戏信息
			consts.MsgTypeGameReady,               // 玩家准备
			consts.MsgTypeGameUserChat,            // 俏皮话
			consts.MsgTypeGameUserYYInfo,          // yaya 语音
			consts.MsgTypeGameTimeMsg,             // 游戏定时器事件
			consts.MsgTypeHouseVitaminSet_Ntf:     // 玩家疲劳值更新
			return true
		default:
			xlog.Logger().Warnf("【游戏暂停中】不接受操作:%+v。", op)
			return false
		}
	}
	return true
}

// 判断用户是否在桌上
func (self *Table) HasIn(uid int64) bool {
	for i := 0; i < len(self.Person); i++ {
		if self.Person[i] != nil && self.Person[i].Uid == uid {
			return true
		}
	}
	return false
}

func (self *Table) CanBeIn(uid int64) *xerrors.XError {
	if !self.IsLive() {
		return xerrors.TableInError
	}
	if self.DelTime != 0 { //已经申请解散
		return xerrors.TableInError
	}
	// if self.Begin { //! 牌局已经开始
	// 	return false
	// }
	if self.IsFull(uid) {
		return xerrors.TableIsFullError
	}
	return nil

}

// 判断旁观用户是否在桌上 ，默认0号位置
func (self *Table) LookonHasIn(uid int64) bool {
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		if self.LookonPerson[0][i] != nil && self.LookonPerson[0][i].Uid == uid {
			return true
		}
	}
	return false
}

func (self *Table) LookonCanBeIn(uid int64) *xerrors.XError {
	if !self.IsLive() {
		return xerrors.TableInError
	}
	if self.DelTime != 0 { //已经申请解散
		return xerrors.TableInError
	}
	// if self.Begin { //! 牌局已经开始
	// 	return false
	// }
	if !self.IsSupportLookonTable(uid) {
		return xerrors.TableIsNotSupportLookonError
	}
	if self.IsLookonFull(uid) {
		return xerrors.TableIsFullError
	}
	return nil

}

func (self *Table) GetPersonByChair(seat int) *base2.TablePerson {
	//for i := 0; i < len(self.Person); i++ {
	//	if self.Person[i] != nil && self.Person[i].Uid == uid {
	//		return self.Person[i]
	//	}
	//}
	return self.Person[seat]
}

func (self *Table) GetPersons() []*base2.TablePerson {
	return self.Person
}

// 切换旁观位置
func (self *Table) LookonSeatSwitch(uid int, seat int) {
	// 判断是否已经在牌桌
	person := GetPersonMgr().GetLookonPerson(int64(uid))
	if person == nil {
		return
	}
	if person.Info.WatchTable != self.Id { //防止用户在多个桌子导致现在的游戏退出
		return
	}
	if self.LookonHasIn(int64(uid)) {
		tablePerson := self.getLookonPerson(int64(uid))
		if tablePerson == nil {
			return
		}
		tablePerson.Seat = seat
		self.game.OnSwitchLookon(person, uint16(seat))
	}
}
func (self *Table) LookonTablein(person *PersonGame) {
	// 判断是否已经在牌桌
	person.Info.WatchTable = self.Id
	if self.LookonHasIn(person.Info.Uid) {
		self.BroadcastTableInfoLookon(true)
		self.game.OnSendInfoLookon(person)

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			//! 通知大厅用户加入成功
			_msg := static.GH_HTableIn_Ntf{
				Uid:     person.Info.Uid,
				TableId: self.Id,
				Result:  true,
				Ip:      person.Ip,
			}
			PushTableStatusMsg(self.HId, consts.MsgTypeTableIn_Ntf, &_msg)
			// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid)
		}
		go self.SubFloorInfo(self.FId)
		return
	}

	if joinErr := self.LookonCanBeIn(person.Info.Uid); joinErr != nil { //! 是否可以加入
		// 更新用户信息
		person.Info.TableId = 0
		person.Info.GameId = 0
		person.Info.WatchTable = 0
		err := GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "WatchTable", person.Info.WatchTable, "GameId", person.Info.GameId)
		if err != nil {
			xlog.Logger().Errorln("Set user data to redis error: ", err.Error())
		}

		// 断开用户连接
		xlog.Logger().Debug("table,tablein,cannot in")
		person.SendMsg(consts.MsgTypeTableIn, joinErr.Code, joinErr.Error())

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			// 强制断开session
			person.CloseSession(consts.SESSION_CLOED_FORCE)

			//! 通知大厅加入失败
			var _msg static.GH_HTableIn_Ntf
			_msg.TableId = self.Id
			_msg.Uid = person.Info.Uid
			_msg.Result = false
			PushTableStatusMsg(self.HId, consts.MsgTypeTableIn_Ntf, &_msg)
			// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableIn_Ntf, xerrors.TableInErrorCode, &_msg, person.Info.Uid)
		}
	} else {
		xlog.Logger().Debug("lookon tableinsucceed")
		//fmt.Println("============================================================首次加入房间")
		self.addLookonPerson(person, 0)
		//先发table消息，再发game
		self.BroadcastTableInfoLookon(true)
		self.game.OnSendInfoLookon(person)
		self.flush()

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			//! 通知大厅加入成功
			var _msg static.GH_HTableIn_Ntf
			_msg.TableId = self.Id
			_msg.Uid = person.Info.Uid
			_msg.Result = true
			_msg.Ip = person.Ip
			PushTableStatusMsg(self.HId, consts.MsgTypeTableIn_Ntf, &_msg)
			// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid) //! 发送游戏信息
		} else if self.Config.GameType == static.GAME_TYPE_GOLD {
			self.NotifyTableChange()
		}
	}
	go self.SubFloorInfo(self.FId)
}

func (self *Table) Tablein(person *PersonGame) {
	// 判断是否已经在牌桌
	if self.HasIn(person.Info.Uid) {
		self.LiveTime = time.Now().Unix()
		//! 广播房间信息
		self.BroadcastTableInfo(true)
		self.game.OnSendInfo(person)

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			//! 通知大厅用户加入成功
			_msg := static.GH_HTableIn_Ntf{
				Uid:     person.Info.Uid,
				TableId: self.Id,
				Result:  true,
				Ip:      person.Ip,
			}
			PushTableStatusMsg(self.HId, consts.MsgTypeTableIn_Ntf, &_msg)
			// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid)
		}
		go self.SubFloorInfo(self.FId)
		return
	}

	// 获取座位号
	seat, ok := self.GetSeat(person.Info.Uid)
	if !ok {
		xlog.Logger().Debug("用户没有落座权限, 提示连接超时")
		person.SendMsg(consts.MsgTypeTableIn, xerrors.TableInError.Code, "连接超时")
		person.CloseSession(consts.SESSION_CLOED_FORCE)
		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			//! 通知大厅加入失败
			var _msg static.GH_HTableIn_Ntf
			_msg.TableId = self.Id
			_msg.Uid = person.Info.Uid
			_msg.Result = false
			GetServer().CallHall("NewServerMsg", consts.MsgTypeTableIn_Ntf, xerrors.TableInErrorCode, &_msg, person.Info.Uid)
		}
		return
	}

	if joinErr := self.CanBeIn(person.Info.Uid); joinErr != nil { //! 是否可以加入
		// 更新用户信息
		person.Info.TableId = 0
		person.Info.GameId = 0
		err := GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "TableId", person.Info.TableId, "GameId", person.Info.GameId)
		if err != nil {
			xlog.Logger().Errorln("Set user data to redis error: ", err.Error())
		}

		// 断开用户连接
		xlog.Logger().Debug("table,tablein,cannot in")
		person.SendMsg(consts.MsgTypeTableIn, joinErr.Code, joinErr.Error())

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			// 强制断开session
			person.CloseSession(consts.SESSION_CLOED_FORCE)

			//! 通知大厅加入失败
			var _msg static.GH_HTableIn_Ntf
			_msg.TableId = self.Id
			_msg.Uid = person.Info.Uid
			_msg.Result = false
			PushTableStatusMsg(self.HId, consts.MsgTypeTableIn_Ntf, &_msg)
			// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableIn_Ntf, xerrors.TableInErrorCode, &_msg, person.Info.Uid)
		}
	} else {
		xlog.Logger().Debug("tableinsucceed")
		//fmt.Println("============================================================首次加入房间")
		self.OnUserNovice(person.Info.Uid)
		self.LiveTime = time.Now().Unix()
		self.addPerson(person, seat)
		//先发table消息，再发game
		self.BroadcastTableInfo(true)
		self.game.OnSendInfo(person)
		self.flush()

		if self.Config.GameType == static.GAME_TYPE_FRIEND {
			//! 通知大厅加入成功
			var _msg static.GH_HTableIn_Ntf
			_msg.TableId = self.Id
			_msg.Uid = person.Info.Uid
			_msg.Result = true
			_msg.Ip = person.Ip
			PushTableStatusMsg(self.HId, consts.MsgTypeTableIn_Ntf, &_msg)
			// GetServer().CallHall("NewServerMsg", constant.MsgTypeTableIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid) //! 发送游戏信息
		} else if self.Config.GameType == static.GAME_TYPE_GOLD {
			self.NotifyTableChange()
		}
	}
	go self.SubFloorInfo(self.FId)
}

//长时间不操作也不解散的游戏 长生不老
func (self *Table) LongLiveGame() bool {
	//潜江麻将 长时间不操作也不解散
	if self.GetTableInfo().KindId == base2.KIND_ID_QJMJ_QJHH || self.GetTableInfo().KindId == base2.KIND_ID_QJMJ_JQDT || self.GetTableInfo().KindId == base2.KIND_ID_QJMJ_QJHZ {
		return true
	}
	return false
}

func (self *Table) runLive() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		<-ticker.C
		if self.LiveTime == 0 {
			break
		}

		if self.reciveChan == nil {
			break
		}

		// 若牌桌长时间没有进行任何活动则自动解散(长生不老型除外)
		if !self.IsLive() && (!self.LongLiveGame()) {
			msg := new(static.Msg_S2C_TableDel)
			msg.Type = forceCloseByTimeout
			msg.Msg = "长时间未操作，牌桌自动解散"
			self.Operator(base2.NewTableMsg(consts.MsgTypeTableDel, "now", 0, msg))
			break
		}

		// 如果牌桌还未开始, 踢出超时连接的用户
		//if !self.IsBegin() {
		//self.Operator(NewTableMsg(constant.MsgTypeTableCheckConn, "now", 0, nil))
		//}
	}

	//！ 定时器关闭
	ticker.Stop()
}

func (self *Table) GetGame() base2.SportInterface {
	return self.game
}

//! 检测用户连接状态
func (self *Table) CheckUserConnection(uid int64, seat int) {
	if seat < 0 {
		panic("error seat")
	}
	if seat >= len(self.Users) || seat >= len(self.Person) {
		xlog.Logger().Warningf("检查玩家%d(seat:%d)连接时，该玩家已被清理掉了.", uid, seat)
		return
	}

	if self.Users[seat] != nil && self.Users[seat].Uid == uid && self.Person[seat] == nil {
		// 踢出用户
		GetDBMgr().db_R.UpdatePersonAttrs(uid, "TableId", 0, "GameId", 0)

		self.Users[seat] = nil
		self.flush()

		// 通知大厅
		_msg := new(static.GH_TableExit_Ntf)
		_msg.TableId = self.Id
		_msg.Uid = uid
		_msg.Hid = self.HId
		_msg.Fid = self.FId
		PushTableStatusMsg(self.HId, consts.MsgTypeTableExit_Ntf, &_msg)
		// Protocolworkers.CallHall(constant.MsgTypeTableExit_Ntf, xerrors.SuccessCode, &_msg, uid)
	}
}

//暂停桌子消息信道
func (self *Table) pause() {
	self.shutdown = true
	//if self.reciveChan != nil {
	//	close(self.reciveChan)
	//	self.reciveChan = nil
	//}
}

func (self *Table) GetShutDownFlag() bool {
	return self.shutdown
}

//! 清理房间
func (self *Table) clear(promptMsg *static.Msg_S2C_TableDel) {
	if self.reciveChan != nil {
		close(self.reciveChan)
		self.reciveChan = nil
	}
	if self.CloseChan != nil {
		self.CloseChan <- struct{}{}
	}
	self.DelTime = 0
	self.LiveTime = 0

	for i := 0; i < len(self.Person); i++ {
		// 跳过空位
		if self.Person[i] == nil {
			continue
		}

		person := GetPersonMgr().GetPerson(self.Person[i].Uid)
		if person == nil {
			continue
		}
		if person.Info.TableId != self.Id { //防止用户在多个桌子导致现在的游戏退出
			continue
		}

		// 更新person信息
		person.Info.TableId = 0
		person.Info.GameId = 0
		// 不是包厢牌桌 && 房间类型为自己创建自己玩 && 房主是自己, 则删除创建房间记录
		if !self.IsTeaHouse() && self.CreateType == consts.CreateTypeSelf && self.Creator == person.Info.Uid {
			person.Info.CreateTable = 0
			GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "TableId", person.Info.TableId, "GameId", person.Info.GameId, "CreateTable", person.Info.CreateTable)
		} else {
			GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "TableId", person.Info.TableId, "GameId", person.Info.GameId)
		}

		//! 通知退出
		if promptMsg != nil {
			person.SendMsg("forcetabledel", xerrors.SuccessCode, promptMsg)
		}

		//! 断开连接
		person.CloseSession(consts.SESSION_CLOED_FORCE)
		GetPersonMgr().DelPerson(self.Person[i].Uid)
	}
	//旁观也要清理
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		// 跳过空位
		if self.LookonPerson[0][i] == nil {
			continue
		}

		person := GetPersonMgr().GetLookonPerson(self.LookonPerson[0][i].Uid)
		if person == nil {
			continue
		}
		if person.Info.WatchTable != self.Id { //防止用户在多个桌子导致现在的游戏退出
			continue
		}

		// 更新person信息
		person.Info.WatchTable = 0
		// 不是包厢牌桌 && 房间类型为自己创建自己玩 && 房主是自己, 则删除创建房间记录
		if !self.IsTeaHouse() && self.CreateType == consts.CreateTypeSelf && self.Creator == person.Info.Uid {
			person.Info.CreateTable = 0
			GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "WatchTable", person.Info.WatchTable, "CreateTable", person.Info.CreateTable)
		} else {
			GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "WatchTable", person.Info.WatchTable)
		}

		//! 通知退出
		if promptMsg != nil {
			person.SendMsg("forcetabledel", xerrors.SuccessCode, promptMsg)
		}

		//! 断开连接
		person.CloseSession(consts.SESSION_CLOED_FORCE)
		GetPersonMgr().DelLookonPerson(self.LookonPerson[0][i].Uid)
	}

	self.WriteTableLog(static.INVALID_CHAIR, "clear 清理牌桌")
	xlog.Logger().Warnln("clear 清理牌桌", self.Id)
	GetTableMgr().DelTable(self)
	//删除观战牌桌
	GetDBMgr().db_R.RemoveWatchTable(self.Id)

}

/*
	财富更新通用方法, 需要注意以下几点：
	1. wealthType为财富类型, 不同财富类型操作的用户属性不一样
	2. cost为正常消耗/系统返还的数值, 负数消耗, 正数返还
*/
func (self *Table) CostUserWealth(uid int64, wealthType int8, cost int, costType int8, playernum int) (int, error) {
	afterNum := 0
	switch wealthType {
	case consts.WealthTypeCard: // 房卡消耗
		self.IsCost = true
		//tx := GetDBMgr().db_M.Begin()
		//defer tx.Rollback()
		tx := GetDBMgr().GetDBmControl() //20210223 这里经常锁死，去掉事务，
		// 更新房卡
		//var costType int8
		leagueID := self.Table.LeagueID
		var befka int
		var aftka int
		var aftfka int
		var err error
		if leagueID > 0 {
			befka, aftka, _, aftfka, err = updLeagueCard(leagueID, uid, cost, cost, costType, tx)
		} else {
			befka, aftka, _, aftfka, err = updcard(uid, cost, cost, tx, costType)
			// redis
			if self.IsTeaHouse() {
				//// 当日0点时间戳
				//timeStr := time.Now().Format("2006-01-02")
				//t, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
				//
				//recordmini := public.RecordGameCostMini{self.DHId, 1, cost, t.Unix()}
				//err = GetDBMgr().GetDBrControl().HouseRecordCostInsert(&recordmini)
				//if err != nil {
				//	syslog.Logger().Errorln(err)
				//	return -1, err
				//}
			}
		}
		if err != nil {
			xlog.Logger().Errorln(err)
			return -1, err
		}
		if cost <= 0 {
			// 创建房卡消耗记录
			dCostRecord := new(models.RecordGameCost)
			dCostRecord.UId = uid
			dCostRecord.TId = self.Id
			dCostRecord.HId = self.DHId
			dCostRecord.FId = self.FId
			dCostRecord.NTId = self.NTId
			dCostRecord.BefKa = befka
			dCostRecord.AftKa = aftka
			dCostRecord.KaCost = cost
			dCostRecord.KindId = self.KindId
			dCostRecord.Gamenum = self.GameNum
			dCostRecord.LeagueID = leagueID
			dCostRecord.GameConfig = static.HF_JtoA(self.GetGameConfig())
			dCostRecord.CreatedAt = time.Now()
			dCostRecord.PlayerNum = playernum
			//写入主表
			if err = tx.Create(dCostRecord).Error; err != nil {
				xlog.Logger().Errorln(err)
				return -1, err
			}

			//go func() {
			//	//写入备份表
			//	if err = GetDBMgr().GetDBmControl().Create(dCostRecord.ConvertModel()).Error; err != nil {
			//		syslog.Logger().Errorln(err)
			//	}
			//}()

			if cost == 0 {
				//tx.Commit()
				return aftka, nil
			}
		}
		if leagueID > 0 {
			cli := GetDBMgr().Redis
			pipe := cli.TxPipeline()

			err := pipe.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID), "freeze_card", int64(cost)).Err()
			if err != nil {
				return -1, err
			}
			err = pipe.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID), "card_num", int64(cost)).Err()
			if err != nil {
				xlog.Logger().Errorln(err)
				return -1, err
			}
			if self.Table.NotPool {
				sql := `update league_user set used_card = used_card + ? ,freeze_card = freeze_card +? where league_id = ? and uid = ?`
				if err := GetDBMgr().GetDBmControl().Exec(sql, -1*cost, cost, leagueID, uid).Error; err != nil {
					xlog.Logger().Errorf("error:%+v", err)
				}
				err := pipe.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER_H, self.Table.Creator), "freeze_card", int64(cost)).Err()
				if err != nil {
					return -1, err
				}
				err = pipe.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER_H, self.Table.Creator), "used_card", int64(-1*cost)).Err()
				if err != nil {
					return -1, err
				}
			}

			_, err = pipe.Exec()
			if err != nil {
				xlog.Logger().Errorln(err)
				return -1, err
			}
			//err = tx.Commit().Error
			if err != nil {
				xlog.Logger().Errorln(err)
				pipe := cli.TxPipeline()
				err := pipe.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID), "freeze_card", int64(-1*cost)).Err()
				if err != nil {
					return -1, err
				}
				err = pipe.HIncrBy(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID), "card_num", int64(-1*cost)).Err()
				if err != nil {
					xlog.Logger().Errorln(err)
					return -1, err
				}
				pipe.Exec()
				return -1, err
			}
		} else {
			p := GetPersonMgr().GetPerson(uid)
			if p != nil {
				p.Info.Card = aftka
				p.Info.FrozenCard = aftfka
			}
			err = GetDBMgr().db_R.UpdatePersonAttrs(uid, "Card", aftka, "FrozenCard", aftfka)
			if err != nil {
				xlog.Logger().Error(err)
			}
			// 提交事务
			//tx.Commit()
			afterNum = aftka
			_, err = GetServer().CallHall("NewServerMsg", consts.MsgTypeUpdWealth, xerrors.SuccessCode, &static.Msg_UpdWealth{Uid: uid, WealthType: wealthType, WealthNum: afterNum, TableId: self.Id}, 0)
			if err != nil {
				xlog.Logger().Error(err)
			}
		}
	default:
		xlog.Logger().Errorln(fmt.Sprintf("unknown game wealth cost type: %d", wealthType))
	}

	// 通知大厅
	return afterNum, nil
}

func (self *Table) SetCostPaid() {
	self.IsCost = true
	self.flush()
}

//! 广播消息
func (self *Table) broadCastMsg(head string, errCode int16, v interface{}) {
	if head == "lineperson" {
		if !v.(*Msg_LinePerson).Line { //! 下线判断
			person := GetPersonMgr().GetPerson(v.(*Msg_LinePerson).Uid)
			if person != nil && person.session != nil {
				return
			}
		}

		self.game.OnLine(v.(*Msg_LinePerson).Uid, v.(*Msg_LinePerson).Line)
	}

	for i := 0; i < len(self.Person); i++ {
		// 跳过空座
		if self.Person[i] == nil {
			continue
		}

		person := GetPersonMgr().GetPerson(self.Person[i].Uid)
		if person == nil {
			continue
		}
		if person.Info.TableId != self.Id {
			continue
		}
		xlog.Logger().Debug("table broadCastMsg uids:", self.Person[i].Uid)
		person.SendMsg(head, errCode, v)
	}
	//发送给旁观玩家
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		// 跳过空座
		if self.LookonPerson[0][i] == nil {
			continue
		}

		person := GetPersonMgr().GetLookonPerson(self.LookonPerson[0][i].Uid)
		if person == nil {
			continue
		}
		if person.Info.WatchTable != self.Id {
			continue
		}
		xlog.Logger().Debug("table broadCastMsg lookon uids:", self.LookonPerson[0][i].Uid)
		person.SendMsg(head, errCode, v)
	}
}

//! 发送消息
func (self *Table) SendMsg(uid int64, head string, errCode int16, v interface{}) {
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return
	}
	if person.Info.TableId != self.Id {
		return
	}

	person.SendMsg(head, errCode, v)
}

//! 发送消息
func (self *Table) SendLookonMsg(uid int64, head string, errCode int16, v interface{}) {
	person := GetPersonMgr().GetLookonPerson(uid)
	if person == nil {
		return
	}
	if person.Info.WatchTable != self.Id {
		return
	}

	person.SendMsg(head, errCode, v)
}

//! 加入人
func (self *Table) addPerson(person *PersonGame, seat int) {
	self.LiveTime = time.Now().Unix()

	per := new(base2.TablePerson)
	per.Uid = person.Info.Uid
	per.ImgUrl = person.Info.Imgurl
	per.Name = person.Info.Nickname
	per.Sex = person.Info.Sex
	per.IP = person.Ip
	per.Seat = seat
	self.Person[seat] = per

	//self.flush()
}

//! 加入人
func (self *Table) addLookonPerson(person *PersonGame, seat int) {
	per := new(base2.TablePerson)
	per.Uid = person.Info.Uid
	per.ImgUrl = person.Info.Imgurl
	per.Name = person.Info.Nickname
	per.Sex = person.Info.Sex
	per.IP = person.Ip
	per.Seat = seat
	self.LookonPerson[seat] = append(self.LookonPerson[seat], per)

	//self.flush()
}

// 获取座位号
func (self *Table) GetTableSeat(uid int64) int {
	for i, p := range self.Person {
		if p != nil && p.Uid == uid {
			return i
		}
	}
	return -1
}

//! 获取牌桌信息
func (self *Table) GetTableMsg(cur int64, onlySelf bool) *static.TableInfoDetail {
	var msg static.TableInfoDetail
	msg.TableId = self.Id
	msg.IsFloorHideImg = self.IsFloorHideImg
	msg.IsMemUidHide = self.IsMemUidHide
	msg.IsForbidWX = self.IsForbidWX
	msg.LookonPerson = make([][]static.Son_PersonInfo, 10)
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		if self.LookonPerson[0][i] == nil {
			continue
		}
		var son static.Son_PersonInfo
		son.Uid = self.LookonPerson[0][i].Uid
		son.Name = self.LookonPerson[0][i].Name
		son.ImgUrl = self.LookonPerson[0][i].ImgUrl
		son.Sex = self.LookonPerson[0][i].Sex
		son.Seat = self.LookonPerson[0][i].Seat
		msg.LookonPerson[0] = append(msg.LookonPerson[0], son)
	}
	for i := 0; i < len(self.Person); i++ {
		// 跳过空位
		if self.Person[i] == nil {
			// 游戏开始后才需要返回
			if !self.IsBegin() || self.Users[i] == nil {
				continue
			}
			uid := self.Users[i].Uid
			person, err := GetDBMgr().GetDBrControl().GetPerson(uid)

			// 是否隐藏名字和头像
			flag := onlySelf && (self.IsAiSuper || self.IsAnonymous) && person.Uid != cur

			if err != nil {
				xlog.Logger().Errorln(err)
				continue
			}
			var son static.Son_PersonInfo
			son.Uid = person.Uid
			if flag {
				son.Name = "昵称隐藏"
				son.ImgUrl = ""
			} else {
				son.Name = person.Nickname
				son.ImgUrl = person.Imgurl
			}
			son.Sex = person.Sex
			son.Seat = i
			son.Ip = ""
			son.Address = ""
			son.Latitude = ""
			son.Longitude = ""
			msg.Person = append(msg.Person, son)
		} else {
			// 是否隐藏名字和头像
			flag := onlySelf && (self.IsAiSuper || self.IsAnonymous) && self.Person[i].Uid != cur

			var son static.Son_PersonInfo
			son.Uid = self.Person[i].Uid
			if flag {
				son.Name = "昵称隐藏"
				son.ImgUrl = ""
			} else {
				son.Name = self.Person[i].Name
				son.ImgUrl = self.Person[i].ImgUrl
			}
			son.Sex = self.Person[i].Sex
			son.Seat = self.GetTableSeat(son.Uid)

			person := GetPersonMgr().GetPerson(self.Person[i].Uid)
			if person == nil || person.session == nil {
				son.Ip = ""
				son.Address = ""
				son.Latitude = ""
				son.Longitude = ""
			} else {
				son.Ip = person.Info.GameIp
				son.Address = person.Info.Address
				son.Latitude = person.Info.Latitude
				son.Longitude = person.Info.Longitude
				son.Card = person.Info.Card
				son.Gold = person.Info.Gold
				son.WinCount = person.Info.WinCount
				son.LostCount = person.Info.LostCount
				son.DrawCount = person.Info.DrawCount
				son.FleeCount = person.Info.FleeCount
				son.TotalCount = person.Info.TotalCount
			}
			msg.Person = append(msg.Person, son)
		}
	}

	msg.Creator = self.Creator
	msg.HId = self.HId
	msg.NFId = self.NFId
	msg.NTId = self.NTId
	//msg.DelTime = self.DelTime
	msg.Step = self.Step
	msg.KindId = self.KindId

	if self.IsTeaHouse() {
		msg.IsHidHide = self.IsHidHide
		msg.IsVitamin = self.IsVitamin
		msg.IsAiSuper = self.IsAiSuper
		msg.IsAnonymous = self.IsAnonymous
		msg.JoinType = self.JoinType
		if self.IsAiSuper {
			msg.CurrentAiNum = self.CurrentMappingNum
			msg.AiSuperNum = self.TotalMappingNum
			msg.AiSuperPercent = 100 * self.CurrentMappingNum / self.TotalMappingNum
		}

		house := GetDBMgr().GetDBrControl().GetHouse(fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, self.DHId))
		if house != nil {
			msg.PrivateGPS = house.PrivateGPS
		}
	}

	msg.GameConfig = self.GetGameConfig()

	return &msg
}

// 写入redis
func (self *Table) flush() {
	GetDBMgr().db_R.Set(self.GetRedisKey(), self.ToBytes())
}

//写入详细数据
func (self *Table) flushInfo() {
	self.GameInfo = self.game.Tojson()
	GetDBMgr().db_R.Set(self.GetRedisKey(), self.ToBytes())
}

// 清除 redis
func (self *Table) remove() {
	GetDBMgr().db_R.Remove(self.GetRedisKey())
}

// 用户加入桌子 seat=-1时顺序落座
func (self *Table) UserJoinTable(uid int64, seat int, payer int64) (int, int64, *xerrors.XError) {
	p, err := GetDBMgr().GetDBrControl().GetPerson(uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return -1, 0, xerrors.UserNotExistError
	}
	// 判断用户有没有坐下
	for i, u := range self.Users {
		if u != nil && u.Uid == p.Uid {
			// 已经坐下的用户不做任何处理
			return i, u.JoinAt, nil
		}
	}
	curSeat := -1
	if seat < 0 {
		for i, u := range self.Users {
			if u == nil {
				curSeat = i
				break
			}
		}
		if curSeat < 0 {
			return -1, 0, xerrors.TableIsFullError
		}
	} else {
		if self.Users[seat] != nil {
			cuserror := xerrors.NewXError("座位上已经有人了")
			return -1, 0, cuserror
		}
		curSeat = seat
	}

	// 更新内存
	timeNow := time.Now().Unix()
	p.TableId = self.Id
	p.GameId = self.GameId
	self.Users[curSeat] = &static.TableUser{
		Uid:    p.Uid,
		JoinAt: timeNow,
		Payer:  payer,
	}
	// 更新redis
	err = GetDBMgr().db_R.UpdatePersonAttrs(p.Uid, "TableId", p.TableId, "GameId", p.GameId)
	if err != nil {
		xlog.Logger().Errorln("Set user data to redis error: ", err.Error())
		return -1, 0, xerrors.DBExecError
	}

	self.flush()

	if self.Config.GameType == static.GAME_TYPE_FRIEND {
		go func() {
			time.Sleep(30 * time.Second)
			self.Operator(base2.NewTableMsg(consts.MsgTypeTableCheckConn, "now", 0, &static.Msg_CheckUserConnection{uid, curSeat}))
		}()
	}
	go self.SubFloorInfo(self.FId)

	return curSeat, timeNow, nil
}

// 退出牌桌
func (self *Table) LookonTableExit(uid, by int64) bool {
	xlog.Logger().Debug("lookon table exittable 1,uid ", uid)
	person := GetPersonMgr().GetLookonPerson(uid)
	if person == nil {
		return false
	}

	tablePerson := self.getLookonPerson(uid)
	if tablePerson == nil {
		return false
	}

	// 清空用户信息
	if self.Config.GameType == static.GAME_TYPE_FRIEND {
		//删除牌桌观战玩家
		if err := GetDBMgr().db_R.RemoveWatchPlayerToTable(person.Info.WatchTable, person.Info.Uid); err != nil {
			person.SendMsg(consts.MstTypeWatcherQuit, xerrors.DBExecError.Code, xerrors.DBExecError.Msg)
			return false
		}

		person.Info.WatchTable = 0
		if err := GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "WatchTable", person.Info.WatchTable, "GameId", person.Info.GameId); err != nil {
			person.SendMsg(consts.MstTypeWatcherQuit, xerrors.DBExecError.Code, xerrors.DBExecError.Msg)
			return false
		}

		// 此处需要先断开用户链接 不然通知大厅用户的在线状态有问题
		person.CloseSession(consts.SESSION_CLOED_FORCE)
		// 删除用户
		GetPersonMgr().DelLookonPerson(uid)
	}
	self.DelLookonPerson(uid)
	return true
}

// 退出牌桌
func (self *Table) TableExit(uid, by int64) bool {
	xlog.Logger().Debug("table exittable 1,uid ", uid)
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return false
	}

	tablePerson := self.getPerson(uid)
	if tablePerson == nil {
		return false
	}

	if self.Begin && tablePerson != nil && tablePerson.Seat >= 0 && !self.game.CheckExit(uid) {
		xlog.Logger().Debug("table exittable gaming and cannot out")
		var msg static.Msg_Null
		person.SendMsg(consts.MsgTypeTableExit, xerrors.GameBeginCanNotExitCode, &msg)
		return false
	}

	if consts.ClubHouseOwnerPay == false {
		xer := self.RollbackFrozenCardOnExit(tablePerson.Seat)
		if xer != nil {
			xlog.Logger().Errorln(xer.Msg)
			return false
		}
	}

	// 被踢掉了
	if by > 0 && uid != by {
		var promptMsg static.Msg_S2C_TableDel
		promptMsg.Type = forceCloseByKick
		if self.IsTeaHouse() {
			promptMsg.Msg = fmt.Sprintf("您已被盟主/管理员 %d 踢出房间。", by)
		} else {
			promptMsg.Msg = fmt.Sprintf("您已被房主 %d 踢出房间。", by)
		}
		person.SendMsg("forcetabledel", xerrors.SuccessCode, &promptMsg)
	}

	var msg static.Msg_Uid
	msg.Uid = person.Info.Uid
	msg.Fid = self.FId
	msg.Hid = self.HId
	self.broadCastMsg(consts.MsgTypeTableExit, 0, &msg)
	self.game.OnExit(uid)

	// 删除座位号
	self.Users[tablePerson.Seat] = nil
	self.flush()

	// 清空用户信息
	if self.Config.GameType == static.GAME_TYPE_FRIEND {
		person.Info.GameId = 0
		person.Info.TableId = 0
		if err := GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "TableId", person.Info.TableId, "GameId", person.Info.GameId); err != nil {
			person.SendMsg(consts.MsgTypeTableExit, xerrors.DBExecError.Code, xerrors.DBExecError.Msg)
			return false
		}

		// 此处需要先断开用户链接 不然通知大厅用户的在线状态有问题
		person.CloseSession(consts.SESSION_CLOED_FORCE)
		// 删除用户
		GetPersonMgr().DelPerson(uid)
	} else if self.Config.GameType == static.GAME_TYPE_GOLD {
		person.Info.TableId = 0
		if err := GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "TableId", person.Info.TableId); err != nil {
			person.SendMsg(consts.MsgTypeTableExit, xerrors.DBExecError.Code, xerrors.DBExecError.Msg)
			return false
		}
	}
	self.DelPerson(uid)
	return true
}

// 换桌清理数据
func (self *Table) tableChange(uid int64) (bool, int) {
	xlog.Logger().Debug("table exittable 1,uid ", uid)
	person := GetPersonMgr().GetPerson(uid)
	_tableId := person.Info.TableId
	if person == nil {
		return false, _tableId
	}

	tablePerson := self.getPerson(uid)
	if tablePerson == nil {
		return false, _tableId
	}

	if self.Begin && tablePerson != nil && tablePerson.Seat >= 0 && !self.game.CheckExit(uid) {
		xlog.Logger().Debug("table exittable gaming and cannot out")
		//var msg public.Msg_Null
		//person.SendMsg(constant.MsgTypeTableExit, xerrors.GameBeginCanNotExitCode, &msg)
		return false, _tableId
	}

	self.game.OnExit(uid)

	// 删除座位号
	self.Users[tablePerson.Seat] = nil
	self.flush()

	// 清空用户信息
	person.Info.GameId = 0
	person.Info.TableId = 0
	GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "TableId", person.Info.TableId, "GameId", person.Info.GameId)

	self.DelPerson(uid)

	var msg static.Msg_Uid
	msg.Uid = person.Info.Uid
	msg.Fid = self.FId
	msg.Hid = self.HId
	self.broadCastMsg(consts.MsgTypeTableExit, 0, &msg)

	return true, _tableId
}

func (self *Table) Bye() {
	self.Begin = false
	self.Operator(base2.NewTableMsg(consts.MsgTypeTableDel, "now", 0, nil))
}

func (self *Table) DelPerson(uid int64) {
	for i := 0; i < len(self.Person); i++ {
		if self.Person[i] != nil && self.Person[i].Uid == uid {
			self.Person[i] = nil
			self.BroadcastTableInfo(true)
			return
		}
	}
}

func (self *Table) getPerson(uid int64) *base2.TablePerson {
	for i := 0; i < len(self.Person); i++ {
		if self.Person[i] != nil && self.Person[i].Uid == uid {
			return self.Person[i]
		}
	}
	return nil
}
func (self *Table) DelLookonPerson(uid int64) {
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		if self.LookonPerson[0][i] != nil && self.LookonPerson[0][i].Uid == uid {
			//self.LookonPerson[0][i] = nil
			self.LookonPerson[0] = append(self.LookonPerson[0][:i], self.LookonPerson[0][i+1:]...)
			self.BroadcastTableInfoLookon(true)
			return
		}
	}
}
func (self *Table) getLookonPerson(uid int64) *base2.TablePerson {
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		if self.LookonPerson[0][i] != nil && self.LookonPerson[0][i].Uid == uid {
			return self.LookonPerson[0][i]
		}
	}
	return nil
}

func (self *Table) getLookonPersonBySeat(seat int) []*base2.TablePerson {
	retPerson := []*base2.TablePerson{}
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		if self.LookonPerson[0][i] != nil && self.LookonPerson[0][i].Seat == seat {
			retPerson = append(retPerson, self.LookonPerson[0][i])
		}
	}
	return retPerson
}

//! 通过座位编号获取玩家信息
func (self *Table) getPersonWithSeatId(seatId int) *base2.TablePerson {
	return self.Person[seatId]
}

//! 是否活动牌桌
func (self *Table) IsLive() bool {
	if self.Config.GameType == static.GAME_TYPE_GOLD {
		return true
	}
	return time.Now().Unix()-self.LiveTime < 1800
}

func (self *Table) dissmissThread() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		<-ticker.C
		if self.DelTime == 0 {
			break
		}

		if time.Now().Unix() >= self.DelTime {
			msg := new(static.Msg_S2C_TableDel)
			msg.Type = forceCloseBySchedule
			msg.Msg = "牌桌定时解散"
			self.Operator(base2.NewTableMsg("tabledel", "now", 0, msg))
			break
		}
	}

	//！ 定时器关闭
	ticker.Stop()
}

//! 写桌子日志
func (self *Table) WriteTableLog(seatId uint16, recordStr string) {
	var recordUser string = ""

	pUid := int64(0)
	if seatId != static.INVALID_CHAIR {
		var person *base2.TablePerson = self.getPersonWithSeatId(int(seatId))
		if person != nil {
			recordUser = person.Name
			pUid = person.Uid
			recordUser += fmt.Sprintf(" uid:%d", pUid)
		}
	}

	self.tableLog.SetExtraHook(pUid, self.Step)

	//var writeStr = fmt.Sprintf("%s%s", recordUser, recordStr)
	self.tableLog.Output(recordUser, recordStr)
}

// 是否是空桌子
func (self *Table) IsEmpty() bool {
	for _, u := range self.Users {
		if u != nil {
			return false
		}
	}
	return true
}

//! 获取当前桌子玩家人数
func (self *Table) GetPersonNumber() int {
	nmber := 0
	for _, v := range self.Person {
		if v != nil {
			nmber++
		}
	}
	return nmber
}

// 推送牌桌变化消息至场次
func (self *Table) NotifyTableChange() {
	// 通知客户端牌桌信息变化
	site := GetSiteMgr().GetSiteByType(self.KindId, self.SiteType)
	if site != nil {
		site.Operator(NewSiteMsg(consts.MsgTypeSiteTable_Ntf, "", 0, self.NTId))
	}
}

func (self *Table) OnInvite(inviter, invitee int64) error {
	if !self.IsTeaHouse() {
		return fmt.Errorf("非包厢牌桌无法发起邀请")
	}
	person := GetPersonMgr().GetPerson(inviter)
	if person == nil {
		return fmt.Errorf("邀请发起者不存在:%d", inviter)
	}
	msg := static.MsgHouseTableInvite{
		TId:     self.Id,
		Invitee: invitee,
	}
	msg.Inviter.Nickname = person.Info.Nickname
	msg.Inviter.Imgurl = person.Info.Imgurl
	msg.Inviter.Uid = inviter
	msg.Inviter.Gender = person.Info.Sex
	_, err := GetServer().CallHall("NewServerMsg", consts.MsgTypeHouseTableInviteSend, xerrors.SuccessCode, &msg, 0)
	return err
}

// 用户新加入牌桌事件
func (self *Table) OnUserNovice(uid int64) {
	mod := new(models.StatisticsUserGameHistory)
	mod.KindId = self.KindId
	mod.Uid = uid
	if err := GetDBMgr().GetDBrControl().GamePlaysSelect(mod); err != nil {
		xlog.Logger().Errorln("OnUserNovice.OnUserNovice.error.", err)
	}

	xlog.Logger().Infoln("OnUserNovice.kindid:", mod.KindId, "times:", mod.PlayTimes)
	mod.PlayTimes++
	mod.UpdatedAt = time.Now()

	if err := GetDBMgr().GetDBrControl().GamePlaysInsert(mod); err != nil {
		xlog.Logger().Errorln("OnUserNovice.GamePlaysInsert.error.", err)
		return
	}
	tx := GetDBMgr().GetDBmControl().Begin()
	if err := tx.Where("uid = ? and kind_id = ?", uid, self.KindId).First(new(models.StatisticsUserGameHistory)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := tx.Save(mod).Error; err != nil {
				xlog.Logger().Errorln("OnUserNovice.GetDBmControl.Save.error.", err)
				tx.Rollback()
				return
			}
		} else {
			xlog.Logger().Errorln("OnUserNovice.GetDBmControl.Mysql.Error.", err)
			tx.Rollback()
			return
		}
	} else {
		if err := tx.Model(mod).Where("uid = ? and kind_id = ?", uid, self.KindId).Update("play_times", mod.PlayTimes).Error; err != nil {
			xlog.Logger().Errorln("OnUserNovice.GetDBmControl.Update.error.", err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// 检查是否有少人开局的功能
func (self *Table) CanFewer() bool {
	// if !self.Config.IsFriendMode() {
	// 	syslog.Logger().Errorln("canFewer：table is not friend game")
	// 	return false
	// }
	// if self.FewerBegin {
	// 	syslog.Logger().HouseBaseInfo("canFewer：already fewer start")
	// 	return false
	// }
	// if !self.IsNew() {
	// 	syslog.Logger().HouseBaseInfo("canFewer：table is begin")
	// 	return false
	// }
	// if Config := GetServer().GetGameConfig(self.KindId); Config != nil {
	// 	min, max := Config.GetPlayerNum()
	// 	if max-min == 0 {
	// 		syslog.Logger().HouseBaseInfo("canFewer：there is no interval.")
	// 		return false
	// 	}
	// 	if self.Config.IsLeastNum(min) {
	// 		syslog.Logger().HouseBaseInfo("canFewer：there is no interval.")
	// 		return false
	// 	}
	// } else {
	// 	syslog.Logger().Errorln("canFewer：game Config is null.")
	// 	return false
	// }

	if CanBeFewer(self.Table) {
		return self.Config.CanFewer()
	}
	return false
}

// 少人开局成功后
func (self *Table) OnFewerStart(players ...int64) {
	if self.FewerBegin {
		return
	}
	self.Table.OnFewerStart()
	// 释放掉无效的用户
	self.UserFree(players...)
	self.flush()
	if _, err := GetServer().CallHall("NewServerMsg", consts.MsgTypeOnFewerStart, xerrors.SuccessCode, &static.Msg_GH_OnFewer{
		TableId:     self.Id,
		ActiveUsers: players,
	}, 0); err != nil {
		xlog.Logger().Errorln("通知大厅变更桌子信息失败：", err)
	}
}

//重新选择位置
func (self *Table) ReSeat(id1 int64, seat int) bool {
	person1 := self.getPerson(id1)

	if person1.Seat == seat {
		return false
	}

	if self.getPersonWithSeatId(seat) != nil {
		return false
	}

	if person1 != nil {
		var personTmp base2.TablePerson
		personTmp.Seat = seat
		personTmp.Copy(person1)

		var user static.TableUser
		user.Uid = self.Users[person1.Seat].Uid
		user.JoinAt = self.Users[person1.Seat].JoinAt
		user.Payer = self.Users[person1.Seat].Payer
		self.Users[person1.Seat] = nil

		if self.Users[seat] == nil {
			self.Users[seat] = new(static.TableUser)
		}
		self.Users[seat].Uid = user.Uid
		self.Users[seat].JoinAt = user.JoinAt
		self.Users[seat].Payer = user.Payer
		self.Person[person1.Seat] = nil
		self.Person[seat] = &personTmp

		self.flush()
		return true
	}

	return false
}

//更换位置数据
func (self *Table) ExChangeSeat(id1, id2 int64) bool {
	person1 := self.getPerson(id1)
	person2 := self.getPerson(id2)
	if person1 != nil && person2 != nil {
		var personTmp base2.TablePerson
		var user static.TableUser
		user.Uid = self.Users[person1.Seat].Uid
		user.JoinAt = self.Users[person1.Seat].JoinAt
		user.Payer = self.Users[person1.Seat].Payer
		self.Users[person1.Seat].Uid = self.Users[person2.Seat].Uid
		self.Users[person1.Seat].JoinAt = self.Users[person2.Seat].JoinAt
		self.Users[person1.Seat].Payer = self.Users[person2.Seat].Payer
		self.Users[person2.Seat].Uid = user.Uid
		self.Users[person2.Seat].JoinAt = user.JoinAt
		self.Users[person2.Seat].Payer = user.Payer

		personTmp.Copy(person1)
		person1.Copy(person2)
		person2.Copy(&personTmp)

		self.flush()

		//随机换座时旁观玩家要跟着变换位置
		tablePersonLookon1 := self.getLookonPersonBySeat(person1.Seat)
		tablePersonLookon2 := self.getLookonPersonBySeat(person2.Seat)
		if tablePersonLookon1 != nil && len(tablePersonLookon1) > 0 {
			for _, item := range tablePersonLookon1 {
				if item != nil {
					var msg static.Msg_S_WatcherSwitch
					msg.Seat = person2.Seat
					msg.Uid = int(item.Uid)
					msg.TableId = self.Id
					self.Operator(base2.NewTableMsg(consts.MstTypeWatcherSwitch, consts.MstTypeWatcherSwitch, item.Uid, msg))
					self.LookonSeatSwitch(int(item.Uid), person2.Seat)
				}
			}
		}

		if tablePersonLookon2 != nil && len(tablePersonLookon2) > 0 {
			for _, item := range tablePersonLookon2 {
				if item != nil {
					var msg static.Msg_S_WatcherSwitch
					msg.Seat = person1.Seat
					msg.Uid = int(item.Uid)
					msg.TableId = self.Id
					self.Operator(base2.NewTableMsg(consts.MstTypeWatcherSwitch, consts.MstTypeWatcherSwitch, item.Uid, msg))
					self.LookonSeatSwitch(int(item.Uid), person1.Seat)
				}
			}
		}
		return true
	}
	return false
}

// 是否为已满
func (self *Table) IsFull(uid int64) bool {
	config := self.game.GetGameConfig()
	if config == nil {
		xlog.Logger().Errorln("获取牌桌配置失败.")
		return true
	}
	return self.GetOtherUserCount(uid) >= config.ChairCount
}

// 是否为已满
func (self *Table) IsLookonFull(uid int64) bool {
	config := self.game.GetGameConfig()
	if config == nil {
		xlog.Logger().Errorln("旁观入桌获取牌桌配置失败.")
		return true
	}
	return len(self.LookonPerson[0]) >= int(config.LookonCount)
}

// 本桌是否为支持旁观
func (self *Table) IsSupportLookonTable(uid int64) bool {
	config := self.game.GetGameConfig()
	if config == nil {
		xlog.Logger().Errorln("旁观入桌获取牌桌配置失败.")
		return true
	}
	return config.LookonSupport
}

// 清掉无效的人
func (self *Table) UserFree(players ...int64) {
	isInArray := func(uid int64) bool {
		for _, v := range players {
			if v == uid {
				return true
			}
		}
		return false
	}
	self.Table.InvalidUserFree(isInArray)
	for i := 0; i < len(self.Person); {
		if self.Person[i] == nil || !isInArray(self.Person[i].Uid) {
			// 后面的人向前挪
			for j := i; j < len(self.Person); j++ {
				if self.Person[j] != nil && isInArray(self.Person[j].Uid) {
					// 座位号向前挪
					self.Person[j].Seat -= 1
				}
			}
			// 删掉这个空位，后面的向前挪
			copy(self.Person[i:], self.Person[i+1:])
			self.Person = self.Person[:len(self.Person)-1]
			continue
		}
		i++
	}
}

func (self *Table) SubFloorInfo(floorID int64) error {
	if self.SubFloor != nil {
		return nil
	}
	if self.CloseChan == nil {
		self.CloseChan = make(chan struct{})
	}
	cli := GetDBMgr().PubRedis
	sub := cli.Subscribe(fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, floorID))
	defer func() {
		sub.Close()
		if self.CloseChan != nil {
			self.CloseChan = nil
		}
	}()

	self.SubFloor = sub
	ch := sub.Channel()
loop:
	for {
		select {
		case <-self.CloseChan:
			break loop
		case msg := <-ch:
			subMsg := static.MsgRedisSub{}
			err := json.Unmarshal([]byte(msg.Payload), &subMsg)
			if err != nil {
				xlog.Logger().Errorf("json 解析数据失败:%v", err)
				continue
			}
			switch subMsg.Head {
			case consts.MsgTypeHouseVitaminSet_Ntf:
				var msg static.Msg_UserTableId
				if err := json.Unmarshal(static.HF_Atobytes(subMsg.Data), &msg); err == nil {
					if self.Id == msg.TableId {
						self.Operator(&base2.TableMsg{
							Head: consts.MsgTypeHouseVitaminSet_Ntf,
							Uid:  msg.Uid,
							V:    &msg,
						})
					}
				}
			case consts.MsgHouseFloorMappingNumUpdate:
				xlog.Logger().Infoln("收到来自包厢楼层redis的订阅消息。", subMsg.Data)
				var msg static.MsgHouseFloorMappingUpdate
				if err := json.Unmarshal(static.HF_Atobytes(subMsg.Data), &msg); err == nil {
					self.CurrentMappingNum = msg.CurrentMappingNum
					self.TotalMappingNum = msg.TotalMappingNum
					// 如果已经开局则不再广播
					if !self.Begin {
						var Message static.Msg_S_SuperInfo
						Message.Num = self.CurrentMappingNum
						Message.AiSuperNum = self.TotalMappingNum
						Message.Percent = 100 * self.CurrentMappingNum / self.TotalMappingNum
						self.broadCastMsg(consts.MsgTypeGameSupperMessage, 0, &Message)
					}
				}
			default:
				for _, p := range self.Person {
					if p == nil {
						continue
					}
					if !p.SubMsg {
						continue
					}
					person := GetPersonMgr().GetPerson(p.Uid)
					if person == nil {
						continue
					}
					if person.Info.TableId != self.Id {
						continue
					}
					xlog.Logger().Debug("table broadCastMsg uids:", p.Uid, subMsg.Data, subMsg.Head)
					person.SendMsg(subMsg.Head, 0, subMsg.Data)
				}
			}
		}
	}
	return nil
}

func (self *Table) UserSubFloorMsg(uid int64, data string) {
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return
	}
	if person.Info.TableId != self.Id {
		return
	}
	dest := static.MsgUserSubFloor{}

	err := json.Unmarshal([]byte(data), &dest)
	if err != nil {
		person.SendMsg("usersubfloor", 0, "参数错误")
		return
	}
	dataSlice := dest.SubDetail
	if len(dest.SubDetail) == 0 && len(dataSlice) > 5 {
		person.SendMsg("usersubfloor", 0, "参数错误")
		return
	}

	for _, u := range self.Person {
		if u == nil {
			continue
		}
		if u.Uid == uid {
			u.SubMsg = true
			u.SubDetail = dataSlice
			var NeedTable bool
			var NeedOnline bool
			var NeedApply bool
			var NeedUser bool

			for _, v := range dataSlice {
				if v == "table_info" {
					NeedTable = true
				}
				if v == "online_user" {
					NeedOnline = true
				}
				if v == "apply_user" {
					NeedApply = true
				}
				if v == "user_info" {
					NeedUser = true
				}
			}
			var floorInfo *static.GameHfInfo
			if NeedTable {
				floorInfo = self.FloorTableInfo(uid)
			} else {
				floorInfo = nil
			}
			var onLineInfo, applyInfo []*static.GameMember
			var userInfo *static.GameMember

			if NeedApply || NeedOnline || NeedUser {
				onLineInfo, applyInfo, userInfo = self.HouseOnLine(uid)
			}
			ntf := static.Msg_UserSubFloor{}
			if floorInfo != nil {
				ntf.TableInfo = floorInfo.Infos
				ntf.MixType = floorInfo.TableJoinType // 混排类型：0手动加桌 1自动加桌 2智能防作弊
				ntf.IsMix = floorInfo.IsMix
				ntf.IsPartnerApply = floorInfo.IsPartnerApply
				ntf.IsAnonymous = self.IsAnonymous
			}
			ntf.Hid = self.HId
			ntf.Fid = self.FId
			if NeedApply {
				ntf.ApplyUser = applyInfo
			}
			if NeedOnline {
				ntf.OnlineUser = onLineInfo
			}
			if NeedUser {
				ntf.UserInfo = userInfo
			}
			person.SendMsg("usersubfloor", 0, ntf)
			return
		}
	}
	return

}

func (self *Table) UserUnSubFloorMsg(uid int64) error {
	if uid == 0 {
		return errors.New("params error")
	}
	for _, u := range self.Person {
		if u == nil {
			continue
		}
		if u.Uid == uid {
			u.SubMsg = false
			return nil
		}
	}
	return errors.New("params error")
}

func (self *Table) FloorTableInfo(uid int64) *static.GameHfInfo {
	msg := static.MsgHFloorInfo{Hid: self.DHId, Fid: self.FId}
	ack, err := GetServer().CallHall("ServerMethod.ServerMsg", consts.MsgTypeHFInfo, xerrors.SuccessCode, &msg, 0)
	if err != nil {
		return nil
	}
	dest := static.GameHfInfo{}
	err = json.Unmarshal(ack, &dest)
	if err != nil {
		panic(err)
	}
	return &dest
}

func (self *Table) FloorTableInfoPush(uid int64) {
	ack := self.FloorTableInfo(uid)
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return
	}
	if person.Info.TableId != self.Id {
		return
	}
	person.SendMsg("floortableinfo", 0, ack.Infos)
}

//HouseOnLine 获取包厢在线列表
func (self *Table) HouseOnLine(uid int64) ([]*static.GameMember, []*static.GameMember, *static.GameMember) {
	// key := fmt.Sprintf("housemember_%d", self.DHId)
	// cli := GetDBMgr().Redis
	// res := cli.HGetAll(key).Val()
	ackOnline := []*static.GameMember{}
	ackApply := []*static.GameMember{}
	userInfo := &static.GameMember{}
	houseMemberMap, err := GetDBMgr().GetDBrControl().GetHouseMemberMap(self.DHId)
	if err != nil {
		xlog.Logger().Error("house online eve: get house member error:", err)
		return ackOnline, ackApply, userInfo
	}
	mem, ok := houseMemberMap[uid]
	if ok {
		self.CoverHouseMemToGame(mem, userInfo)
		userInfo.Hid = self.HId
	} else {
		xlog.Logger().Error("house online eve: nil member:", uid)
		return ackOnline, ackApply, userInfo
	}

	for _, dest := range houseMemberMap {
		if dest.URole <= consts.ROLE_MEMBER {
			if dest.IsOnline {
				info := &static.GameMember{}
				self.CoverHouseMemToGame(dest, info)
				info.Hid = self.HId
				if !info.IsInGame {
					ackOnline = append(ackOnline, info)
				}
			}
		} else if dest.URole == consts.ROLE_APLLY {
			// 如果是合伙人，则只显示名下玩家申请
			if userInfo.Partner == 1 && dest.Partner != userInfo.Uid {
				continue
			}
			info := &static.GameMember{}
			self.CoverHouseMemToGame(dest, info)
			info.Hid = self.HId
			info.ApplyType = consts.HouseMemberApplyJoin
			ackApply = append(ackApply, info)
		}
	}

	// for _, v := range res {
	// 	dest := public.HouseMember{}
	// 	err := json.Unmarshal([]byte(v), &dest)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	houseMemberMap[dest.UId] = &dest
	// 	if dest.UId == uid {
	// 		CoverHouseMemToGame(&dest, userInfo)
	// 		userInfo.Hid = self.HId
	// 		continue
	// 	}
	// 	if dest.URole <= constant.ROLE_MEMBER {
	// 		if dest.IsOnline {
	// 			eve := &public.GameMember{}
	// 			CoverHouseMemToGame(&dest, eve)
	// 			eve.Hid = self.HId
	// 			if !eve.IsInGame {
	// 				ackOnline = append(ackOnline, eve)
	// 			}
	// 		}
	//
	// 	} else if dest.URole == constant.ROLE_APLLY {
	// 		eve := &public.GameMember{}
	// 		CoverHouseMemToGame(&dest, eve)
	// 		eve.Hid = self.HId
	// 		eve.ApplyType = constant.HouseMemberApplyJoin
	// 		ackApply = append(ackApply, eve)
	// 	}
	// }

	// 退圈申请列表
	exitApplies := make([]*models.HouseExit, 0)
	err = GetDBMgr().GetDBmControl().Find(&exitApplies, "hid = ? and state = ?", self.DHId, models.HouseMemberExitApplying).Error
	if err != nil {
		xlog.Logger().Error(err)
	}
	for i, l := 0, len(exitApplies); i < l; i++ {
		exitUid := exitApplies[i].Uid
		if exitUid == uid {
			continue
		}
		dest, ok := houseMemberMap[exitUid]
		if !ok {
			continue
		}
		// 如果是合伙人，则只显示名下玩家申请
		if userInfo.Partner == 1 && dest.Partner != userInfo.Uid {
			continue
		}
		info := &static.GameMember{}
		self.CoverHouseMemToGame(dest, info)
		info.Hid = self.HId
		info.ApplyType = consts.HouseMemberApplyExit
		ackApply = append(ackApply, info)
	}
	return ackOnline, ackApply, userInfo
}

//// 得到桌子包厢信息 读不加锁
//func (self *Table) GetFloorVitaminOption() (*public.FloorVitaminOptions, error) {
//	if self.IsTeaHouse() {
//		house, err := self.GetHouse()
//		if err != nil {
//			return nil, err
//		}
//		floor, err := self.GetFloor()
//		if err != nil {
//			syslog.Logger().Errorln("get house floor from redis error:", err)
//			return nil, err
//		}
//		floor.FloorVitaminOptions.IsVitamin = floor.IsVitamin && house.IsVitamin
//		floor.FloorVitaminOptions.IsGamePause = floor.IsGamePause && house.IsGamePause
//		return &floor.FloorVitaminOptions, nil
//	}
//	return nil, fmt.Errorf("table is not a house floor table: %d", self.DHId)
//}
//
//// 得到桌子包厢信息 读不加锁
//func (self *Table) GetFloor() (*public.HouseFloor, error) {
//	if self.IsTeaHouse() {
//		floor, err := GetDBMgr().GetDBrControl().HouseFloorSelect(self.DHId, self.FId)
//		if err != nil {
//			syslog.Logger().Errorln("get house floor from redis error:", err)
//			return nil, err
//		}
//		return floor, nil
//	}
//	return nil, fmt.Errorf("table is not a house floor table: %d", self.DHId)
//}
//
//// 得到桌子包厢信息 读不加锁
//func (self *Table) GetHouse() (*model.House, error) {
//	if self.IsTeaHouse() {
//		house, err := GetDBMgr().GetDBrControl().GetHouseInfoById(self.DHId)
//		if err != nil {
//			syslog.Logger().Errorln("get house from redis error:", err)
//			return nil, err
//		}
//		return house, nil
//	}
//	return nil, fmt.Errorf("table is not a house floor table: %d", self.DHId)
//}

//// 得到桌子包厢成员信息 读不加锁
//func (self *Table) GetMember(uid int64) (*public.HouseMember, error) {
//	if self.IsTeaHouse() {
//		return GetDBMgr().GetDBrControl().HouseMemberQueryById(self.DHId, uid)
//	}
//	return nil, fmt.Errorf("table is not a house floor table: %d", self.DHId)
//}
//
//// 更新桌子包厢成员信息
//func (self *Table) SetMember(mem *public.HouseMember) error {
//	if self.IsTeaHouse() {
//		buf, err := json.Marshal(mem)
//		if err != nil {
//			return err
//		}
//		return GetDBMgr().Redis.HSet(fmt.Sprintf("housemember_%d", self.DHId), fmt.Sprintf("%d", mem.UId), fmt.Sprintf("%s", buf)).Err()
//	}
//	return nil
//}

func (self *Table) CoverHouseMemToGame(data *static.HouseMember, dest *static.GameMember) {

	dest.AgreeTime = data.AgreeTime
	dest.ApplyTime = data.ApplyTime
	dest.URole = data.URole
	dest.Uid = data.UId
	dest.Partner = data.Partner
	dest.IsLimitGame = data.IsLimitGame
	dest.IsOnline = data.IsOnline
	p, _ := GetDBMgr().GetDBrControl().GetPerson(dest.Uid)
	if p == nil {
		return
	}
	dest.UUrl = p.Imgurl
	dest.Gender = p.Sex
	dest.Uname = p.Nickname
	dest.IsInGame = p.TableId > 0
	if static.IsAnonymous(self.Config.GameConfig) {
		dest.UUrl = ""
		dest.Uname = "匿名昵称"
	}
	return
}

// 少人开局效验
func CanBeFewer(table *static.Table) bool {
	if !table.Config.IsFriendMode() {
		xlog.Logger().Errorln("CanBeFewer：table is not friend game")
		return false
	}
	if table.FewerBegin {
		xlog.Logger().Info("CanBeFewer：already fewer start")
		return false
	}
	if !(!table.Begin && table.Step == 0) {
		xlog.Logger().Info("CanBeFewer：table is begin")
		return false
	}
	if config := GetServer().GetGameConfig(table.KindId); config != nil {
		min, max := config.GetPlayerNum()
		//fmt.Println("min max", min, max)
		if max-min == 0 {
			xlog.Logger().Trace("CanBeFewer：there is no interval.")
			return false
		}
		if table.Config.IsLeastNum(min) {
			xlog.Logger().Trace("CanBeFewer：there is no interval.")
			return false
		}
	} else {
		xlog.Logger().Errorln("CanBeFewer：game Config is null.")
		return false
	}
	xlog.Logger().Info("CanBeFewer：true.")
	return true
}

func (self *Table) BroadcastTableInfo(private bool /*私密的：如果是超强防作弊且私密，则隐藏其他玩家的名字/头像*/) {
	if self.game.AiSupperLow() && self.IsBegin() { //游戏已经开始，切超级防作弊low标记为true
		private = false
	}

	limitDistance := self.GetHouseTableLimitDistance()
	for i := 0; i < len(self.Person); i++ {
		// 跳过空座
		if self.Person[i] == nil {
			continue
		}
		person := GetPersonMgr().GetPerson(self.Person[i].Uid)
		if person == nil {
			continue
		}
		if person.Info.TableId != self.Id {
			continue
		}

		detail := self.GetTableMsg(person.Info.Uid, private)
		detail.DistanceLimit = limitDistance
		person.SendMsg(consts.MsgTypeTableInfo, 0, detail)
	}
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		// 跳过空座
		if self.LookonPerson[0][i] == nil {
			continue
		}
		person := GetPersonMgr().GetLookonPerson(self.LookonPerson[0][i].Uid)
		if person == nil {
			continue
		}
		if person.Info.WatchTable != self.Id {
			continue
		}

		detail := self.GetTableMsg(person.Info.Uid, private)
		detail.DistanceLimit = limitDistance
		person.SendMsg(consts.MsgTypeTableInfo, 0, detail)
	}
}

func (self *Table) BroadcastTableInfoLookon(private bool /*私密的：如果是超强防作弊且私密，则隐藏其他玩家的名字/头像*/) {
	if self.game.AiSupperLow() && self.IsBegin() { //游戏已经开始，切超级防作弊low标记为true
		private = false
	}

	limitDistance := self.GetHouseTableLimitDistance()
	for i := 0; i < len(self.LookonPerson[0]); i++ {
		// 跳过空座
		if self.LookonPerson[0][i] == nil {
			continue
		}
		person := GetPersonMgr().GetLookonPerson(self.LookonPerson[0][i].Uid)
		if person == nil {
			continue
		}
		if person.Info.WatchTable != self.Id {
			continue
		}

		detail := self.GetTableMsg(person.Info.Uid, private)
		detail.DistanceLimit = limitDistance
		person.SendMsg(consts.MsgTypeTableInfo, 0, detail)
	}
}

func (self *Table) GetHouseTableLimitDistance() int {
	limitDistance := 0
	if self.IsTeaHouse() {
		model := models.HouseTableDistanceLimit{}
		err := GetDBMgr().GetDBmControl().Where("dhid = ?", self.DHId).First(&model).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				xlog.Logger().Errorf("GetHouseTableLimitDistance db error：%v", err)
			}
			limitS := GetDBMgr().Redis.Get("gps_house_limit").Val()
			limitDest, err := strconv.ParseInt(limitS, 10, 64)
			if err != nil {
				limitDistance = -1
			} else {
				limitDistance = int(limitDest)
			}
		} else {
			limitDistance = model.TableDistanceLimit
		}
		return limitDistance
	} else {
		limitS := GetDBMgr().Redis.Get("gps_house_limit").Val()
		limitDest, err := strconv.ParseInt(limitS, 10, 64)
		if err != nil {
			limitDistance = -1
		} else {
			limitDistance = int(limitDest)
		}
	}
	return limitDistance
}

func (self *Table) RollbackFrozenCardOnExit(seat int) *xerrors.XError {
	if !self.IsTeaHouse() {
		return nil
	}
	if self.Begin {
		return nil
	}
	if self.Step > 0 {
		return nil
	}
	if self.IsCost {
		return nil
	}
	player := self.GetUser(seat)
	if player == nil {
		xlog.Logger().Errorln("play is nil")
		return xerrors.UserNotExistError
	}
	if player.Payer <= 0 {
		return nil
	}
	// 回退冻结数据
	payer, err := GetDBMgr().GetDBrControl().GetPerson(player.Payer)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError
	}

	tx := GetDBMgr().GetDBmControl().Begin()
	_, aftka, _, aftfka, err := wealthtalk.UpdateCard(payer.Uid, 0, -self.Config.CardCost, consts.WealthTypeCard, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		tx.Rollback()
		return xerrors.DBExecError
	} else {
		// 更新db
		tx.Commit()
		// 更新内存
		p := GetPersonMgr().GetPerson(payer.Uid)
		if p != nil {
			p.Info.Card = aftka
			p.Info.FrozenCard = aftfka
		}
		// 更新redis
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(payer.Uid, "Card", aftka, "FrozenCard", aftfka)
	}
	return nil
}
