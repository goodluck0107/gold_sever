package components

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"strconv"
	"sync"
)

type BaseFunc struct {
	GameEndStatus byte //当前小局游戏的状态
	GameStatus    byte //游戏状态
	KIND_ID       int
	PlayerInfo    map[int64]*Player   //玩家数据
	Plock         *sync.RWMutex       //玩家数据锁
	LookonPlayer  map[int64]*Player   //旁观用户
	Lplock        *sync.RWMutex       //旁观用户数据锁
	GameTable     base2.TableBase     //桌子
	Rule          rule2.St_FriendRule //规则

	BankerUser uint16 //庄家用户
	//m_wChairCount  uint16 //玩家数量
	m_bGameStarted bool //是否开始游戏
	GameStartMode  byte //开始模式
	IsAiSupperLow  bool // 超级防作弊，在游戏开始时可以线上头像
}

//这种改法可以保证加锁的和没有加锁的游戏并存，否侧只能一次把全部游戏的2125个地方全改了。
//游戏玩家读写锁
func (self *BaseFunc) PlayerInfoRead(f func(*map[int64]*Player)) {
	self.Plock.RLock()
	defer self.Plock.RUnlock()
	f(&self.PlayerInfo)
}

//游戏玩家读写锁
func (self *BaseFunc) PlayerInfoWrite(f func(*map[int64]*Player)) {
	self.Plock.Lock()
	defer self.Plock.Unlock()
	f(&self.PlayerInfo)
}

//旁观玩家读写锁
func (self *BaseFunc) LPlayerInfoRead(f func(*map[int64]*Player)) {
	self.Lplock.RLock()
	defer self.Lplock.RUnlock()
	f(&self.LookonPlayer)
}

//旁观玩家读写锁
func (self *BaseFunc) LPlayerInfoWrite(f func(*map[int64]*Player)) {
	self.Lplock.Lock()
	defer self.Lplock.Unlock()
	f(&self.LookonPlayer)
}

//设置游戏状态
func (self *BaseFunc) SetGameStatus(status byte) {
	xlog.Logger().Debug("setGamestatus :" + strconv.Itoa(int(status)))
	self.GameStatus = status
}

func (self *BaseFunc) GetGameStatus() byte {
	return self.GameStatus
}

// 设置游戏小局状态
func (self *BaseFunc) SetGameRoundStatus(status byte) {
	//play小局开始,end小局结算开始,free小局结算结束
	self.GameEndStatus = status
}

func (self *BaseFunc) GetGameRoundStatus() byte {
	return self.GameEndStatus
}

//给桌子所有玩家发送消息
func (self *BaseFunc) SendTableMsg(head string, date interface{}) bool {
	xlog.Logger().Debug("game SendTableMsg  head :" + head)
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			//syslog.Logger().Debug("send msg head:" + head + "   uid:" + strconv.Itoa(int(v.Uid)))
			if v != nil && v.UserStatus != static.US_OFFLINE {
				self.GameTable.SendMsg(v.Uid, head, xerrors.SuccessCode, date)
			}
		}
	})
	return true
}

//给对应玩家发送消息
func (self *BaseFunc) SendPersonMsg(head string, date interface{}, chairId uint16) bool {
	if chairId == static.INVALID_CHAIR { //无效椅子
		return false
	}
	xlog.Logger().Debug("game SendPersonMsg  head :" + head + " chiar :" + strconv.Itoa(int(chairId)))
	okFlag := false
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Seat == chairId && v.UserStatus != static.US_OFFLINE /*&& head == "gameOperateNotify" */ {
				xlog.Logger().Debug("find charid head:" + head + "   uid:" + strconv.Itoa(int(v.Uid)))
				self.GameTable.SendMsg(v.Uid, head, xerrors.SuccessCode, date)
				okFlag = true
				break
			}
		}
	})
	return okFlag
}

//给对应玩家发送消息
func (self *BaseFunc) SendPersonLookonMsg(head string, date interface{}, uid int64) bool {
	if uid == 0 { //无效uid
		return false
	}
	xlog.Logger().Debug("game SendPersonLookonMsg  head :" + head + " uid :" + strconv.Itoa(int(uid)))
	for _, v := range self.LookonPlayer {
		if v != nil && v.Uid == uid {
			xlog.Logger().Debug("find uid head:" + head + "   uid:" + strconv.Itoa(int(v.Uid)))
			self.GameTable.SendLookonMsg(v.Uid, head, xerrors.SuccessCode, date)
			return true
		}
	}
	return false
}

//给桌子所有旁观玩家发送消息
func (self *BaseFunc) SendTableLookonMsg(head string, date interface{}) bool {
	xlog.Logger().Debug("game SendTableLookonMsg  head :" + head)
	for _, v := range self.LookonPlayer {
		//syslog.Logger().Debug("send msg head:" + head + "   uid:" + strconv.Itoa(int(v.Uid)))
		if v.UserStatus != static.US_OFFLINE {
			self.GameTable.SendLookonMsg(v.Uid, head, xerrors.SuccessCode, date)
		}
	}
	return true
}

func (self *BaseFunc) SendUserMsg(head string, date interface{}, uid int64) {
	xlog.Logger().Debug("send msg head:" + head + "   uid:" + strconv.Itoa(int(uid)))
	self.GameTable.SendMsg(uid, head, xerrors.SuccessCode, date)
}
func (self *BaseFunc) SendLookonUserMsg(head string, date interface{}, uid int64) {
	xlog.Logger().Debug("send lookon msg head:" + head + "   uid:" + strconv.Itoa(int(uid)))
	self.GameTable.SendLookonMsg(uid, head, xerrors.SuccessCode, date)
}

//获取坐桌ID
func (self *BaseFunc) GetTableId() int {
	return self.GetTableInfo().Id
}

//获取房间端口
func (self *BaseFunc) GetSortId() int {
	return 1
}

// 是否紧挨两个玩家之间  不可跨越
func (self *BaseFunc) IsBetweenFull(left, center, right uint16) bool {
	if self.GetNextSeat(center) == int(right) && self.GetFrontSeat(center) == int(left) {
		return true
	}
	return false
}

// 找到临近还原用户的玩家
func (self *BaseFunc) GetNearestUser(origin uint16, userList ...uint16) uint16 {
	nearestUser := static.INVALID_CHAIR
	have := func(cur uint16) bool {
		for _, v := range userList {
			if v == cur {
				return true
			}
		}
		return false
	}
	next := origin
	for {
		if have(next) {
			nearestUser = uint16(next)
			break
		}
		next = uint16(self.GetNextSeat(next))
		if next == origin {
			break
		}
	}
	return nearestUser
}

//根据椅子获取玩家数据
func (self *BaseFunc) GetUserItem(chair int) *server2.PersonGame {
	tmp_Person := self.GameTable.GetPersonByChair(chair)
	if tmp_Person != nil {
		return server2.GetPersonMgr().GetPerson(tmp_Person.Uid)
	}
	return nil
}

//获取下一个玩家位置
func (self *BaseFunc) GetNextSeat(seat uint16) int {
	return (int(seat) + 1) % self.GetPlayerCount()
}

func (self *BaseFunc) GetNextFullSeat(seat uint16) uint16 {
	for i := 0; i < self.GetChairCount(); i++ {
		seat = uint16(self.GetNextSeat(seat))
		if _item := self.GetUserItemByChair(seat); _item != nil && _item.Ctx.IsPlaying {
			break
		}
	}
	return seat
}

func (self *BaseFunc) GetNextFullSeatV2(seat uint16) uint16 {
	for i := 0; i < self.GetChairCount(); i++ {
		seat = uint16(self.GetNextSeat(seat))
		if _item := self.GetUserItemByChair(seat); _item != nil {
			break
		}
	}
	return seat
}

func (self *BaseFunc) GetFrontFullSeat(seat uint16) uint16 {
	for i := 0; i < self.GetTableInfo().Config.MaxPlayerNum; i++ {
		seat = uint16(self.GetFrontSeat(seat))
		if _item := self.GetUserItemByChair(seat); _item != nil && _item.Ctx.IsPlaying {
			break
		}
	}
	return seat
}

// 判断是否是包厢对局
func (self *BaseFunc) IsHouse() bool {
	return self.GetTableInfo().HId != 0
}

//获取上一个玩家位置
func (self *BaseFunc) GetFrontSeat(seat uint16) int {
	return (int(seat) + self.GetPlayerCount() - 1) % self.GetPlayerCount()
}

//设置游戏开始模式
func (self *BaseFunc) SetGameStartMode(mode byte) {
	self.GameStartMode = mode
}

func (self *BaseFunc) SendPlayStatus(play *Player) {
	//效验参数
	if play.IsActive() == false {
		return
	}

	//变量定义
	var UserStatus static.Msg_S_UserStatus
	//构造数据
	UserStatus.UserID = play.Uid
	UserStatus.TableID = int(play.GetTableID())
	UserStatus.ChairID = int(play.GetChairID())
	UserStatus.UserStatus = byte(play.GetUserStatus())
	UserStatus.UserReady = play.UserReady
	//发送数据
	//发给同桌的人
	if play.GetTableID() != static.INVALID_TABLE {
		//通知其他玩家
		self.SendTableMsg(consts.MsgTypeGameUserStatus, UserStatus)
	} else {
		//SendData(pIServerUserItem,MDM_GR_USER,SUB_GR_USER_STATUS,&UserStatus,sizeof(UserStatus));
		self.SendPersonMsg(consts.MsgTypeGameUserStatus, UserStatus, uint16(UserStatus.ChairID))
	}
	//发送旁观数据
	self.SendTableLookonMsg(consts.MsgTypeGameUserStatus, UserStatus)

	//m_AndroidUserManager.SendDataToClient(MDM_GR_USER,SUB_GR_USER_STATUS,&UserStatus,sizeof(UserStatus));

}

func (self *BaseFunc) GetHouseApi() *HouseApi {
	if self.IsHouse() {
		tableInfo := self.GetTableInfo()
		return &HouseApi{
			GameNum:        tableInfo.GameNum,
			TId:            tableInfo.Id,
			HId:            tableInfo.HId,
			NFId:           tableInfo.NFId,
			NTId:           tableInfo.NTId,
			DHId:           tableInfo.DHId,
			FId:            tableInfo.FId,
			Creator:        tableInfo.Creator, // 包厢桌子 creator就是盟主id
			RealPlayerNum:  self.GetPlayerCount(),
			FloorPlayerNum: self.GetProperPNum(),
		}
	}
	return nil
}

//发送用户
func (self *BaseFunc) SendUserItem(item *Player, uid int64, lookon int) bool {
	//效验参数
	if item == nil {
		return false
	}

	flag := (self.GetTableInfo().IsAiSuper || self.GetTableInfo().IsAnonymous) && uid != item.Uid

	if self.GetTable().IsBegin() && self.IsAiSupperLow {
		flag = false
	}

	//构造数据
	var pUserInfoHead static.Msg_S_UserInfoHead
	//填写数据
	//pUserInfoHead.FaceID = item.FaceID
	//pUserInfoHead.CustomFaceVer = 0 //pUserData->dwCustomFaceVer;
	pUserInfoHead.TableID = item.GetTableID()
	pUserInfoHead.ChairID = item.GetChairID()
	pUserInfoHead.Gender = item.Sex
	//pUserInfoHead.UserStatus = item.GetUserStatus()
	pUserInfoHead.Uid = item.GetUserID()
	pUserInfoHead.GameID = item.GameID
	//pUserInfoHead.GroupID = item.GroupID
	//pUserInfoHead.UserRight = item.UserRight
	//pUserInfoHead.Loveliness = item.Loveliness
	//pUserInfoHead.MasterRight = item.MasterRight
	pUserInfoHead.MemberOrder = item.MemberOrder
	pUserInfoHead.MasterOrder = item.MasterOrder
	pUserInfoHead.UserScoreInfo = item.UserScoreInfo.ToV2()
	pUserInfoHead.UserScoreInfo.Score = self.GetRealScore(item.UserScoreInfo.Score)
	item.UserScoreInfo.Vitamin = static.SwitchVitaminToF64(self.GetUserVitamin(item, nil))
	pUserInfoHead.UserScoreInfo.Vitamin = item.UserScoreInfo.Vitamin
	if flag {
		pUserInfoHead.Name = "昵称隐藏" //, pUserData->szAccounts, sizeof(pUserData->szAccounts));
		pUserInfoHead.FaceUrl = ""  //如果有自定义头像，还要发送头像信息
	} else {
		pUserInfoHead.Name = item.Name       //, pUserData->szAccounts, sizeof(pUserData->szAccounts));
		pUserInfoHead.FaceUrl = item.FaceUrl //如果有自定义头像，还要发送头像信息
	}
	//pUserInfoHead.ExtInfo = "ip:" + item.Ip
	if lookon == 0 {
		self.SendUserMsg(consts.MsgTypeGameUserInfoHead, pUserInfoHead, uid)
	} else {
		self.SendLookonUserMsg(consts.MsgTypeGameUserInfoHead, pUserInfoHead, uid)
	}

	return true
}

// 得到玩家疲劳值
func (self *BaseFunc) GetUserVitamin(user *Player, fo *static.FloorVitaminOptions /*为空则从Redis取*/) int64 {
	houseApi := self.GetHouseApi()

	if houseApi == nil {
		return 0
	}
	var err error
	if fo == nil {
		fo, err = houseApi.GetFloorVitaminOption()
		if err != nil {
			self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("GetUserVitamin获取桌子包厢信息失败:%s", err.Error()))
			return 0
		}
	}
	if !fo.IsVitamin {
		self.OnWriteGameRecord(static.INVALID_CHAIR, "House Floor Vitamin False")
		return 0
	}

	mem := &static.HouseMember{
		DHId: self.GetTableInfo().DHId,
		UId:  user.Uid,
	}
	cli := server2.GetDBMgr().Redis
	mem.Lock(cli)
	defer mem.Unlock(cli)

	mem, err = houseApi.GetMember(user.Uid)
	if err != nil {
		self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("GetUserVitamin获取桌子包厢成员%d信息失败:%s", user.Uid, err.Error()))
		return 0
	}
	dbVitamin, err := server2.GetDBMgr().GetUserLatestVitaminFromDataBase(self.GetTableInfo().DHId, user.Uid)
	if err == nil {
		if mem.UVitamin != dbVitamin {
			mem.UVitamin = dbVitamin
			err = houseApi.SetMember(mem)
			if err != nil {
				self.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("GetUserVitamin同步包厢成员数据%d信息失败:%s", user.Uid, err.Error()))
				return 0
			}
		}
	}
	user.UserScoreInfo.Vitamin = static.SwitchVitaminToF64(mem.UVitamin)
	return mem.UVitamin
}

//获取椅子人数
func (self *BaseFunc) GetChairCount() int {
	//return m_pGameServiceOption->wPlayerCount > m_wChairCount ? m_wChairCount : m_pGameServiceOption->wPlayerCount;
	//return len(self.GameTable.Person)
	return int(self.GetConfig().ChairCount)
}

//获取玩家人数
func (self *BaseFunc) GetUserCount() int {
	playUserCount := 0
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Seat != static.INVALID_CHAIR {
				playUserCount++
			}
		}
	})
	return playUserCount
	//return len(self.GameTable.Person)
}

//获取玩家椅子
func (self *BaseFunc) GetChairByUid(uid int64) uint16 {
	seat := static.INVALID_CHAIR
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Uid == uid {
				seat = v.Seat
				break
			}
		}
	})
	return seat
}

//获取玩家uid
func (self *BaseFunc) GetUidByChair(seatId uint16) int64 {
	uid := int64(0)
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.GetChairID() == seatId {
				uid = v.Uid
				break
			}
		}
	})
	return uid
}

//获取玩家player信息
func (self *BaseFunc) GetPlayerByChair(seatId uint16) *Player {
	var player *Player
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.GetChairID() == seatId {
				player = v
				break
			}
		}
	})
	return player
}

// 获取玩家人数
func (self *BaseFunc) GetPlayerCount() int {
	return int(self.GetConfig().PlayerCount)
}

// 获取游戏原本人数 创建桌子时选择的人数
// 写这个函数是因为很多位置判断了3人默认去万，但是少人开局后
// 3人玩4人的规则 不去万，人数又变为了三人，这里加了对应的逻辑
func (self *BaseFunc) GetProperPNum() int {
	count := self.GetPlayerCount()
	if self.GetTableInfo().FewerBegin {
		return count + 1
	}
	return count
}

//是否完毕
func (self *BaseFunc) IsClientReady(userItem *Player) bool {
	//用户判断
	if userItem == nil {
		return false
	}

	return userItem.IsReady()
}

//获取框架，用来调取子类方法
func (self *BaseFunc) getServiceFrame() base2.SportInterface {
	return self.GameTable.GetGame()
}

func (self *BaseFunc) GetTable() base2.TableBase {
	return self.GameTable
}

func (self *BaseFunc) GetTableInfo() *static.Table {
	return self.GameTable.GetTableInfo()
}

//获取游戏房间配置
func (self *BaseFunc) GetConfig() *static.GameConfig {
	return self.getServiceFrame().GetGameConfig()
}

func (self *BaseFunc) CheckExit(uid int64) bool {
	if self.GetTableInfo().Config.GameType == static.GAME_TYPE_FRIEND {
		return false
	}
	okFlag := true
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Uid == uid {
				okFlag = v.LookonFlag
				break
			}
		}
	})
	return okFlag
}

//重新设置庄家
func (self *BaseFunc) standupReSetBanker(seat uint16) bool {
	//庄家设置
	if seat == self.BankerUser {
		self.BankerUser = static.INVALID_CHAIR

		for i := uint16(0); i < uint16(self.GetPlayerCount()); i++ {
			if i != seat && self.GetUserItemByChair(i) != nil {
				self.BankerUser = i
				return true
			}
		}
	}

	return false
}

func (self *BaseFunc) GetUserItemByChair(chair uint16) *Player {
	var player *Player
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Seat == chair {
				player = v
				break
			}
		}
	})
	return player
}

func (self *BaseFunc) ExistsUser(uid int64) bool {
	ok := false
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		_, ok = (*m)[uid]
	})
	return ok
}

func (self *BaseFunc) GetUserItemByUid(uid int64) *Player {
	var player *Player
	self.PlayerInfoRead(func(m *map[int64]*Player) {
		for _, v := range *m {
			if v != nil && v.Uid == uid {
				player = v
				break
			}
		}
	})
	return player
}

func (self *BaseFunc) GetLookonUserItemByUid(uid int64) *Player {
	for _, v := range self.LookonPlayer {
		if v != nil && v.Uid == uid {
			return v
		}
	}
	return nil
}

//获取对应座位的所有旁观者
func (self *BaseFunc) GetLookonUserItemsByChair(chair uint16) []*Player {
	UserItems := []*Player{}
	for _, v := range self.LookonPlayer {
		if v != nil && v.Seat == chair {
			UserItems = append(UserItems, v)
		}
	}
	return UserItems
}

//游戏状态
func (self *BaseFunc) IsUserPlaying(player *Player) bool {
	//游戏状态
	if self.m_bGameStarted == false {
		return false
	}

	//用户状态
	cbUserStatus := player.GetUserStatus()
	if (cbUserStatus != static.US_PLAY) && (cbUserStatus != static.US_OFFLINE) {
		return false
	}
	return !self.getServiceFrame().CheckExit(player.Uid)

	//逻辑判断
	//return true
}

//判断开始
func (self *BaseFunc) StartVerdict() bool {
	//比赛判断
	if self.m_bGameStarted == true {
		//syslog.Logger().Debug("[游戏状态，不开始]")
		return false
	}

	IgnoreOffline := self.GetConfig().StartIgnoreOffline
	OffLineZBFP := self.GetConfig().OffLineZBFP
	if !IgnoreOffline {
		//by leon，有人断线，不开始
		for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
			userItem := self.GetUserItemByChair(i)
			if userItem != nil && userItem.GetUserStatus() == static.US_OFFLINE {
				//syslog.Logger().Debug("[有人断线，不开始]")
				userStatus := userItem.GetUserStatus_ex()
				//20200708 苏大强 京山要求在玩家在点击准备后离线，还是可以发牌的  OffLineZBFP在其他游戏里面默认就是false的
				if !(userStatus == static.US_READY && OffLineZBFP) {
					return false
				}
			}
		}
	}

	//时间模式
	StartMode := self.GetConfig().StartMode
	if StartMode == static.StartMode_TimeControl {
		//syslog.Logger().Debug("[时间模式，不开始]")
		return false
	}
	//准备人数
	wReadyUserCount := uint16(0)
	var pUserData *Player
	for i := uint16(0); i < uint16(self.GetChairCount()); i++ {
		pUserData = self.GetUserItemByChair(i)
		if pUserData != nil {
			wReadyUserCount++
			if !self.GetTableInfo().FewerBegin {
				userStatus := pUserData.GetUserStatus()
				userStatus_ex := pUserData.GetUserStatus_ex()
				if IgnoreOffline {
					if !pUserData.IsReady() {
						//syslog.Logger().Debug("[有人未准备，不开始]")
						return false
					}
				} else {
					//20200708 苏大强 京山需求 如果玩家离线了，但是选了准备，还是可以发牌的
					if userStatus != static.US_READY && !(userStatus_ex == static.US_READY && OffLineZBFP) {
						//syslog.Logger().Debug("[有人未准备，不开始]")
						return false
					}
				}
			}
		}
	}
	//条件判断
	if wReadyUserCount > 1 {
		if StartMode == static.StartMode_Symmetry {
			if (wReadyUserCount % 2) != 0 {
				//syslog.Logger().Debug("[对称模式：准备人数不对称，不开始]")
				return false
			}
			if wReadyUserCount == uint16(self.GetChairCount()) {
				return true
			}
			wHalfCount := uint16(self.GetChairCount()) / 2
			for i := uint16(0); i < wHalfCount; i++ {
				if (self.GetUserItemByChair(i) == nil) && (self.GetUserItemByChair(i+wHalfCount) != nil) {
					//syslog.Logger().Debug("[对称模式：玩家空指针，不开始]")
					return false
				}
				if (self.GetUserItemByChair(i) != nil) && (self.GetUserItemByChair(i+wHalfCount) == nil) {
					//syslog.Logger().Debug("[对称模式：玩家空指针，不开始")
					return false
				}
			}
			return true
		} else {
			if StartMode == static.StartMode_FullReady {
				if wReadyUserCount == uint16(self.GetChairCount()) || wReadyUserCount == self.GetConfig().PlayerCount {
					return true
				} else {
					//syslog.Logger().Debug("[满人模式：人数未满，不开始]", wReadyUserCount, self.GetChairCount(), self.GetConfig().PlayerCount)
					return false
				}
			}
			if StartMode == static.StartMode_AllReady {
				return true
			}
		}
	}
	//特殊判断
	if (wReadyUserCount == 1) && (self.GetConfig().ChairCount == 1) {
		return true
	}
	//syslog.Logger().Debug("[未知原因：不开始]")
	return false
}

//! 获取最终分数
func (self *BaseFunc) GetRealScore(score int) float64 {
	if score == 0 {
		return float64(0)
	}
	if self.Rule.Radix <= 0 {
		return static.HF_DecimalDivide(float64(score), 1, 2)
	}
	return static.HF_DecimalDivide(float64(score)/float64(self.Rule.Radix), 1, 2)
}

//! 写游戏日志
func (self *BaseFunc) OnWriteGameRecord(seatId uint16, recordStr string) {
	self.GameTable.WriteTableLog(seatId, recordStr)
}
