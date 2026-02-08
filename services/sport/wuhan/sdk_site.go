package wuhan

import (
	// "encoding/json"
	"encoding/json"
	"fmt"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	"log"
	"runtime/debug"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

/////////////////////////////////////////////////////////
type SiteMsg struct {
	Head string
	Data string
	Uid  int64
	V    interface{}
}

func NewSiteMsg(head string, data string, uid int64, v interface{}) *SiteMsg {
	sitemsg := new(SiteMsg)
	sitemsg.Head = head
	sitemsg.Data = data
	sitemsg.Uid = uid
	sitemsg.V = v

	return sitemsg
}

//! 场次上的人
type SitePerson struct {
	Uid    int64  `json:"uid"`    //!
	Name   string `json:"name"`   //! 名字
	ImgUrl string `json:"imgurl"` //! 头像
	Sex    int    `json:"sex"`    //! 性别
	//SiteId int    `json:"site"`   //! 座位编号
	IP    string `json:"ip"`   //! 座位编号
	Gold  int    `json:"gold"` // 金币数
	Start int    `json:"-"`    //! 所在列表起始偏移量,[start, end]
	End   int    `json:"-"`    //! 所在列表结束偏移量,[start, end]
}

//!场次
type Site struct {
	Id           int                    `json:"-"` //! siteid
	SitMode      int                    `json:"-"` //! 是否是坐桌模式
	KindId       int                    `json:"-"` //! 子游戏类型
	SiteType     int                    `json:"-"` //! 场次类型
	TableIds     []int                  `json:"-"` //! 索引号即为牌桌编号,牌桌id规则: kindid+type+index
	Index        int                    `json:"-"`
	Person       map[int64]*SitePerson  `json:"-"` //! 场次里的人
	MaxPeopleNum int                    `json:"-"` //! 最大限制进入人数
	receiveChan  chan *SiteMsg          `json:"-"` //! 操作队列
	MinGold      int                    `json:"-"` //最小入场金币数(0为不限制)
	MaxGold      int                    `json:"-"` //最大入场金币数(0为不限制)
	tableNum     int                    `json:"-"` //场次内牌桌数量
	gameConfig   map[string]interface{} `json:"-"` //玩法配置
	matchConfig  static.MatchConfig     `json:"-"` //! 排位赛配置
}

func (self *Site) Init(c *models.ConfigSite, tableNum int, maxPeopleNum int) {
	// 参数配置
	self.SitMode = c.SitMode
	self.MinGold = c.MinScore
	self.MaxGold = c.MaxScore
	self.tableNum = tableNum
	self.MaxPeopleNum = maxPeopleNum

	// 获取现有桌子数量
	length := len(self.TableIds)
	if length >= tableNum {
		// 如果现有桌子数量超过设置的桌子数量, 则不做任何处理
		return
	}

	// 不足的桌子数量补足
	// 获取游戏参数
	gameConfig := GetServer().GetGameConfig(c.KindId)
	if gameConfig == nil {
		xlog.Logger().Errorln(fmt.Sprintf("can not find game_config: [kindid: %d]", c.KindId))
		return
	}
	json.Unmarshal([]byte(gameConfig.GameConfig), &self.gameConfig)
	// 更新参数
	self.gameConfig["difen"] = int(c.Config["difen"].(float64)) // 底分
	self.gameConfig["fa"] = int(c.Config["fa"].(float64))       // 逃跑罚款倍数
	self.gameConfig["jiang"] = int(c.Config["jiang"].(float64)) // 没逃跑的人奖励倍数
	//configBytes, _ := json.Marshal(self.gameConfig)
	configBytes := c.ConfigStr

	//// 获取游戏排位赛参数
	//matchconfig := GetServer().GetMatchConfig(c.KindId, c.Type)
	//if matchconfig == nil {
	//	syslog.Logger().Errorln(fmt.Sprintf("can not find match_config: [kindid: %d,type：%d]", c.KindId, c.Type))
	//	//return
	//} else {
	//	self.matchConfig.Id = matchconfig.Id
	//	self.matchConfig.Name = matchconfig.Name
	//	self.matchConfig.KindId = matchconfig.KindId
	//	self.matchConfig.Type = matchconfig.Type
	//	self.matchConfig.ConfigStr = matchconfig.ConfigStr
	//	self.matchConfig.State = matchconfig.State
	//	self.matchConfig.Flag = matchconfig.Flag
	//	self.matchConfig.BeginDate = matchconfig.BeginDate.Unix()
	//	self.matchConfig.EndDate = matchconfig.EndDate.Unix()
	//	self.matchConfig.BeginTime = matchconfig.BeginTime.Unix()
	//	self.matchConfig.EndTime = matchconfig.EndTime.Unix()
	//}

	// 初始化桌子
	for i := 0; i < tableNum-length; i++ {
		c := &static.Msg_HG_CreateTable{
			Id:           self.GetTableId(i),
			NTId:         length + i,
			CreateType:   consts.CreateTypeSystem,
			KindId:       self.KindId,
			MinPlayerNum: gameConfig.DefaultPlayerNum,
			MaxPlayerNum: gameConfig.DefaultPlayerNum,
			RoundNum:     gameConfig.DefaultRoundNum,
			CardCost:     int(c.Config["revenue"].(float64)),
			CostType:     gameConfig.DefaultCostType,
			View:         gameConfig.DefaultView,
			Restrict:     false,
			GVoice:       "false",
			GameConfig:   string(configBytes),
			MatchConfig:  self.matchConfig,
		}
		table := GetTableMgr().NewTable(c)
		table.Config.GameType = static.GAME_TYPE_GOLD // 金币房
		table.SiteType = self.SiteType                // 场次类型
		GetTableMgr().CreateTable(table)

		self.TableIds = append(self.TableIds, table.Id)
	}
}

// 获取场次牌桌id
func (self *Site) GetTableId(index int) int {
	return self.KindId*10000 + self.SiteType*1000 + index
}

//! 发送操作
func (self *Site) Operator(op *SiteMsg) {
	if self.receiveChan != nil {
		if op.Head != consts.MsgTypeTableCheckConn {
			xlog.Logger().Debug(fmt.Sprintf("site: %d, op: %s", self.Id, op.Head))
		}
		self.receiveChan <- op
	}
}
func (self *Site) run() {
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			if self.receiveChan == nil {
				return
			}
		case op, ok := <-self.receiveChan:
			if !ok {
				continue
			}
			switch op.Head {
			case consts.MsgTypeSiteIn: //! 建立连接和加入(session)
				self.Sitein(op.V.(*PersonGame))
			case consts.MsgTypeSiteExit: // 用户退出场次
				self.SiteExit(op.Uid)
			case consts.MsgTypeSiteListIn: // 加入列表
				p := self.getPerson(op.Uid)
				if p != nil {
					v := op.V.(*static.Msg_Game_SiteListIn)
					p.Start = v.Start
					p.End = v.End

					// 推送牌桌列表详情
					self.SendMsg(op.Uid, consts.MsgTypeSiteTableList, 0, self.getTableList(p.Start, p.End))
				}
			case consts.MsgTypeSiteListOut: // 退出列表
				p := self.getPerson(op.Uid)
				if p != nil {
					p.Start = 0
					p.End = 0
				}
			case consts.MsgTypeSiteTable_Ntf: // 牌桌信息变化
				ntid := op.V.(int)
				self.NotifyTableInfo(ntid)
			case consts.MsgTypeSiteTableIn: //! 从场次内进入牌桌
				//从场次内申请加入牌桌
				self.TableIn(consts.MsgTypeSiteTableIn, op.V.(*static.Msg_Game_SiteInTable), -1)
			case consts.MsgTypeChangeTableIn: //! 换桌
				self.ChangeTableIn(consts.MsgTypeChangeTableIn, op.V.(*static.Msg_Game_ChangeInTable))
			case "broadcast": //! 广播
				self.broadCastMsg(op.Data, xerrors.SuccessCode, op.V)
			default:

			}
		}
	}
}

//! 通知场内用户牌桌信息变化
func (self *Site) NotifyTableInfo(ntid int) {
	if ntid < 0 || ntid > self.tableNum || self.TableIds[ntid] == 0 {
		return
	}

	table := GetTableMgr().GetTable(self.TableIds[ntid])
	if table == nil {
		return
	}

	// 获取牌桌信息
	t := new(SiteTableInfo)
	t.NTId = table.NTId
	t.Lock = false // 默认不锁桌
	t.MaxPlayerNum = table.Config.MaxPlayerNum
	t.Begin = table.Begin
	t.Person = make([]*SitePerson, t.MaxPlayerNum)
	for i, p := range table.Person {
		if p == nil {
			t.Person[i] = nil
		} else {
			t.Person[i] = self.getPerson(p.Uid)
		}
	}

	// 遍历用户, 推送通知
	for _, p := range self.Person {
		// 不在列表内的直接忽略
		if p.Start == 0 && p.End == 0 {
			continue
		}
		// 在区间内的用户才推送消息
		if p.Start <= ntid && p.End > ntid {
			person := GetPersonMgr().GetPerson(p.Uid)
			if person != nil {
				person.SendMsg(consts.MsgTypeSiteTable_Ntf, 0, t)
			}
		}
	}
}

//! 加入牌桌
func (self *Site) TableIn(head string, msg *static.Msg_Game_SiteInTable, preId int) bool {
	//获取用户，创建用户
	person := GetPersonMgr().GetPerson(msg.Uid)
	if person == nil {
		xlog.Logger().Debug("session,UserNotExistError:")
		person.SendMsg(head, xerrors.UserNotExistError.Code, xerrors.UserNotExistError.Msg)
		return false
	}

	//无效坐桌编号
	if msg.Id >= self.tableNum {
		xlog.Logger().Debug("session,TableIn,err:")
		person.SendMsg(head, xerrors.TableInError.Code, xerrors.TableInError.Msg)
		return false
	}

	//玩家保存桌号和进入桌号不一致
	if msg.Id > 0 && person.Info.TableId != 0 && person.Info.TableId != self.TableIds[msg.Id] {
		xlog.Logger().Debug("session,TableIn,err:")
		person.SendMsg(head, xerrors.TableInLockError.Code, xerrors.TableInLockError.Msg)
		return false
	}

	// 判断牌桌是否存在
	var table *Table //:= GetTableMgr().GetTableInfo(self.TableIds[msg.Id])

	//随机入桌
	if preId > 0 { //换桌
		table = GetTableMgr().GetRandTable(self, preId)
	} else {
		if msg.Id < 0 {
			if person.Info.TableId > 0 {
				table = GetTableMgr().GetTable(person.Info.TableId)
			} else {
				table = GetTableMgr().GetRandTable(self, person.Info.TableId)
			}
		} else {
			table = GetTableMgr().GetTable(self.TableIds[msg.Id])
		}
	}

	if table == nil { //! 坐下失败
		xlog.Logger().Errorln("session,tableconin,err:table not exit:")
		person.SendMsg(head, xerrors.SiteIsFullError.Code, xerrors.SiteIsFullError.Msg)
		return false
	}

	person.Info.TableId = table.Id

	_, _, err := table.UserJoinTable(msg.Uid, msg.Seat, 0)
	if err != nil { //获取座位失败
		log.Println(err)
		person.SendMsg(head, err.Code, err.Msg)
		return false
	}

	if table.Config.GameType != static.GAME_TYPE_FRIEND { //入桌成功返回，暂时只在非好友房加载
		_msg := xerrors.NewXError("")
		_msg.Code = xerrors.SuccessCode
		person.SendMsg(head, _msg.Code, _msg)
	}

	table.Operator(base2.NewTableMsg(consts.MsgTypeTableIn, "", person.Info.Uid, person))

	return true
}

//查询场次桌子
func (self *Site) findTable(tableid int) bool {
	for i := 0; i < self.tableNum; i++ {
		if tableid == self.TableIds[i] {
			return true
		}
	}
	return false
}

//! 换桌
func (self *Site) ChangeTableIn(head string, msg *static.Msg_Game_ChangeInTable) bool {
	person := GetPersonMgr().GetPerson(msg.Uid)
	_tableId := int(-1)
	//玩家保存桌号和进入桌号不一致
	if person.Info.TableId > 0 {
		table := GetTableMgr().GetTable(person.Info.TableId)
		ok := false
		if ok, _tableId = table.tableChange(person.Info.Uid); !ok {
			xlog.Logger().Errorln("session,tableconin,err:table not exit:")
			person.SendMsg(head, xerrors.TableInChangeError.Code, xerrors.TableInChangeError.Msg)
			return false
		}

		person.SendMsg(head, xerrors.SuccessCode, msg)

		// 通知客户端牌桌信息变化
		self.Operator(NewSiteMsg(consts.MsgTypeSiteTable_Ntf, "", 0, table.NTId))
	}

	var _msg static.Msg_Game_SiteInTable
	_msg.Uid = msg.Uid
	_msg.Id = -1   //换桌，随机进入
	_msg.Seat = -1 //换桌随机进入
	//	_msg.Type = msg.Type
	//_msg.KindId = msg.KindId
	return self.TableIn(head, &_msg, _tableId)
}

func (self *Site) Sitein(person *PersonGame) {
	// 判断是否已经在牌桌
	//if self.HasIn(person.Info.Uid) {
	//	//! 广播场次信息
	//	self.broadCastMsg(constant.MsgTypeSiteInfo, 0, self.getSiteMsg())
	//
	//	//! 通知大厅用户加入成功
	//	_msg := public.Msg_SiteIn_Result{
	//		Uid:    person.Info.Uid,
	//		SiteId: self.Id,
	//		Number: true,
	//		Ip:     person.Ip,
	//	}
	//	GetServer().CallHall("NewServerMsg", constant.MsgTypeSiteIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid)
	//	return
	//}
	//
	//var _msg public.Msg_SiteIn_Result
	//_msg.SiteId = self.Id
	//_msg.Uid = person.Info.Uid
	//if !self.CanBeIn(person.Info.Uid) { //! 是否可以加入
	//	// 更新用户信息
	//	person.Info.TableId = 0
	//	person.Info.GameId = 0
	//	person.Info.SiteId = 0
	//
	//	// 断开用户连接
	//	syslog.Logger().Debug("site,sitein,cannot in")
	//	person.SendMsg(constant.MsgTypeSiteIn, xerrors.TableInError.Code, xerrors.TableInError.Error())
	//	person.CloseSession(constant.SESSION_CLOED_FORCE)
	//
	//	//! 通知大厅加入失败
	//	_msg.Number = false
	//	GetServer().CallHall("ServerMethod.ServerMsg", constant.MsgTypeSiteIn_Ntf, xerrors.TableInErrorCode, &_msg, person.Info.Uid)
	//} else {
	//	syslog.Logger().Debug("siteinsucceed")
	//	self.addPerson(person)
	//
	//	self.broadCastMsg(constant.MsgTypeSiteInfo, 0, self.getSiteMsg()) //! 广播房间信息
	//
	//	//! 通知大厅加入成功
	//	_msg.Number = true
	//	_msg.Ip = person.Ip
	//	GetServer().CallHall("NewServerMsg", constant.MsgTypeSiteIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid) //! 发送游戏信息
	//}

	if ok, _errmsg := self.CanBeIn(person.Info); !ok { //! 是否可以加入
		// 断开用户连接
		xlog.Logger().Debug("site,sitein,cannot in")
		cuserror := xerrors.NewXError(_errmsg)
		person.SendMsg(consts.MsgTypeSiteIn, cuserror.Code, cuserror.Error())
		person.CloseSession(consts.SESSION_CLOED_FORCE)
		return
	}

	// 更新用户信息
	person.Info.GameId = GetServer().Con.Id
	person.Info.SiteId = self.Id
	GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "GameId", person.Info.GameId, "SiteId", person.Info.SiteId)

	self.addPerson(person)

	// 推送场次详情
	person.SendMsg(consts.MsgTypeSiteInfo, 0, self.getSiteMsg(&person.Info))
	// 通知大厅用户加入场次成功
	_msg := static.Msg_SiteIn_Result{
		Uid:    person.Info.Uid,
		GameId: person.Info.GameId,
		SiteId: self.Id,
		Result: true,
		Ip:     person.Ip,
	}
	GetServer().CallHall("NewServerMsg", consts.MsgTypeSiteIn_Ntf, xerrors.SuccessCode, &_msg, person.Info.Uid)
}

// 退出牌桌
func (self *Site) SiteExit(uid int64) {
	xlog.Logger().Debug("site exittable 1,uid ", uid)
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return
	}

	SitePerson := self.getPerson(uid)
	if SitePerson == nil {
		return
	}

	if person.Info.TableId != 0 {
		//syslog.Logger().Debug("site sitetable  and cannot out")
		//var msg public.Msg_Null
		//person.SendMsg(constant.MsgTypeTableExit, xerrors.GameBeginCanNotExitCode, &msg)
		return
	}

	// 无需广播用户退出消息
	//var msg public.Msg_Uid
	//msg.Uid = person.Info.Uid
	//self.broadCastMsg(constant.MsgTypeSiteExit, 0, &msg)

	// 清空用户信息
	person.Info.GameId = 0
	person.Info.SiteId = 0
	GetDBMgr().db_R.UpdatePersonAttrs(person.Info.Uid, "GameId", person.Info.GameId, "SiteId", person.Info.SiteId)

	self.DelPerson(uid)
}

// 判断用户是否在场次内
func (self *Site) HasIn(uid int64) bool {
	return self.Person[uid] != nil
}

func (self *Site) CanBeIn(p static.Person) (bool, string) {
	//_errormsg :=nil
	// 入场门槛校验
	// 不在房间内的用户才需校验
	if p.TableId == 0 {
		if self.MinGold != 0 && p.Gold < self.MinGold {
			return false, xerrors.GoldNotEnoughError.Msg //"您持有的欢乐豆不足，无法进入该房间，即可补充欢乐豆，决战到天明吧"
		}
		if self.MaxGold != 0 && p.Gold > self.MaxGold {
			return false, xerrors.GoldExceedingError.Msg //"您持有的欢乐豆超出房间限制"
		}
	}
	return true, ""
}

//! 广播消息
func (self *Site) broadCastMsg(head string, errCode int16, v interface{}) {
	for uid, per := range self.Person {
		// 跳过空座
		if uid == 0 || per == nil {
			continue
		}

		person := GetPersonMgr().GetPerson(uid)
		if person == nil {
			continue
		}
		if person.Info.SiteId != self.Id {
			continue
		}
		xlog.Logger().Debug("site broadCastMsg uids:", uid)
		person.SendMsg(head, errCode, v)
	}
}

//! 发送消息
func (self *Site) SendMsg(uid int64, head string, errCode int16, v interface{}) {
	person := GetPersonMgr().GetPerson(uid)
	if person == nil {
		return
	}

	if person.Info.SiteId != self.Id {
		return
	}

	person.SendMsg(head, errCode, v)
}

//! 加入人
func (self *Site) addPerson(person *PersonGame) {
	per := new(SitePerson)
	per.Uid = person.Info.Uid
	per.ImgUrl = person.Info.Imgurl
	per.Name = person.Info.Nickname
	per.Sex = person.Info.Sex
	per.IP = person.Ip
	per.Gold = person.Info.Gold
	self.Person[per.Uid] = per
}

func (self *Site) DelPerson(uid int64) {
	delete(self.Person, uid)
}

// 获取场次详情
func (self *Site) getSiteMsg(person *static.Person) *static.Site_SC_Msg {
	// 获取用户桌号
	ntid := -1
	if person.TableId != 0 {
		table := GetTableMgr().GetTable(person.TableId)
		if table != nil {
			ntid = table.NTId
		} else {
			// 清除用户信息
			person.TableId = 0
			GetDBMgr().db_R.UpdatePersonAttrs(person.Uid, "TableId", 0)
		}
	}

	siteMsg := new(static.Site_SC_Msg)
	siteMsg.SiteId = self.Id
	siteMsg.SitMode = self.SitMode
	siteMsg.KindId = self.KindId
	siteMsg.SiteType = self.SiteType
	siteMsg.TableNum = self.tableNum
	siteMsg.GameConfig = self.gameConfig
	siteMsg.Person = &static.Person_SC_Msg{
		Uid:      person.Uid,
		Ntid:     ntid,
		Imgurl:   person.Imgurl,
		Nickname: person.Nickname,
		Sex:      person.Sex,
		Gold:     person.Gold,
	}
	return siteMsg
}

//! 详情牌桌
type SiteTableInfo struct {
	NTId         int           `json:"ntid"`
	Lock         bool          `json:"lock"`         // 是否锁住
	MaxPlayerNum int           `json:"maxplayernum"` // 游戏开始最大人数
	Begin        bool          `json:"begin"`        // 游戏是否开始
	Person       []*SitePerson `json:"person"`       // 牌桌上的用户
}

// 获取场次牌桌列表
func (self *Site) getTableList(start, end int) []*SiteTableInfo {
	result := make([]*SiteTableInfo, 0)
	// 判断下标, 防止数组越界
	if start >= len(self.TableIds) {
		return result
	}
	if end+1 > len(self.TableIds) {
		end = len(self.TableIds) - 1
	}

	for _, tid := range self.TableIds[start : end+1] {
		table := GetTableMgr().GetTable(tid)
		if table == nil {
			continue
		}
		t := new(SiteTableInfo)
		t.NTId = table.NTId
		t.Lock = false // 默认不锁桌
		t.MaxPlayerNum = table.Config.MaxPlayerNum
		t.Begin = table.Begin
		t.Person = make([]*SitePerson, t.MaxPlayerNum)
		for i, p := range table.Person {
			if p == nil {
				t.Person[i] = nil
			} else {
				t.Person[i] = self.getPerson(p.Uid)
			}
		}
		result = append(result, t)
	}
	return result
}

func (self *Site) getPerson(uid int64) *SitePerson {
	return self.Person[uid]
}

///////////////////////////////////////////////////////////////////////////////////
//! 场次管理者
type SiteMgr struct {
	MapSite map[int]*Site
	lock    *lock2.RWMutex
}

var sitemgrSingleton *SiteMgr = nil

//! 得到场次管理
func GetSiteMgr() *SiteMgr {
	if sitemgrSingleton == nil {
		sitemgrSingleton = new(SiteMgr)
		sitemgrSingleton.MapSite = make(map[int]*Site)
		sitemgrSingleton.lock = new(lock2.RWMutex)
	}
	return sitemgrSingleton
}

// 新建场次
func (self *SiteMgr) NewSite(kindId int, _type int) *Site {
	site := new(Site)
	// 初始化基本参数
	site.Id = getSiteId(kindId, _type)
	site.KindId = kindId
	site.SiteType = _type
	site.Person = make(map[int64]*SitePerson, 0)
	site.receiveChan = make(chan *SiteMsg, 2000)
	site.TableIds = make([]int, 0)
	site.gameConfig = make(map[string]interface{})

	go site.run()

	return site
}

//! 删除场次
func (self *SiteMgr) DelSite(siteId int) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()

	// 删除内存
	delete(self.MapSite, siteId)
}

//! 添加场次
func (self *SiteMgr) AddSite(site *Site) {
	self.lock.CustomLock()
	defer self.lock.CustomUnLock()

	// 添加
	self.MapSite[site.Id] = site
}

//! 获取场次
func (self *SiteMgr) GetSiteByType(kindId int, _type int) *Site {
	id := getSiteId(kindId, _type)
	return self.GetSite(id)
}

//! 获取场次
func (self *SiteMgr) GetSite(id int) *Site {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()
	site, ok := self.MapSite[id]
	if !ok {
		return nil
	}
	return site
}

//修改site的matchconfig
func (self *SiteMgr) UpdateSiteAttr(kindId int, _type int, state int) {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()
	site := self.GetSiteByType(kindId, _type)
	if site != nil {
		//修改这个site的matchconfig
		site.matchConfig.State = state
		//修改这个site下的所有teble的matchconfig
		for _, tid := range site.TableIds[:] {
			table := GetTableMgr().GetTable(tid)
			if table == nil {
				continue
			}
			table.Config.MatchConfig.State = state
		}
	}
	return
}

// 初始化所有游戏场次
func (self *SiteMgr) InitSites(list map[int][]*static.ServerGameType) {
	for _, v := range list {
		for _, item := range v {
			if item.GameType == static.GAME_TYPE_GOLD {
				// 判断场次是否存在
				site := self.GetSiteByType(item.KindId, item.SiteType)
				if site == nil {
					// 不存在则新建场次
					site = self.NewSite(item.KindId, item.SiteType)
					self.AddSite(site)
				}
				// 获取场次配置
				c := GetServer().GetRoomConfig(item.KindId, item.SiteType)
				if c != nil {
					// 依据配置重新初始化部分参数
					site.Init(c, item.TableNum, item.MaxPeopleNum)
				}
			}
		}
	}
}

func getSiteId(kindId int, _type int) int {
	return kindId*100 + _type
}

//! 用户区域广播
func (self *SiteMgr) SendUserBroadcast(kindid []int, head string, v interface{}) {
	self.lock.RLockWithLog()
	defer self.lock.RUnlock()

	isInArray := func(v int, arr []int) bool {
		for _, item := range arr {
			if v == item {
				return true
			}
		}
		return false
	}

	for _, site := range self.MapSite {
		if isInArray(site.KindId, kindid) {
			site.broadCastMsg(head, xerrors.SuccessCode, v)
			break
		}
	}
}
