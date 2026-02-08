package center

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/models"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"maps"
	"math/rand"
	"slices"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
)

const (
	InvalidHouseTableSeat = -1
	ElkHouseTbl           = "hft"
)

// //////////////////////////////////////////////////////////////////////////////
// ! 楼层牌桌数据结构
type HouseFloorTable struct {
	NTId           int            // 序号
	TId            int            // 牌桌id
	CreateStamp    int64          // 创建时间戳
	IsDefault      bool           // 是否为默认桌子
	Begin          bool           //是否开始
	Step           int            //游戏局数
	DataLock       *lock2.RWMutex `json:"-"`
	IsOccupy       int64          `json:"-"` // 桌子状态
	Deleted        bool
	LockClose      chan struct{} `json:"-"`
	UserWithOnline []FTUsers
	Watchers       []int64
	FakeEndAt      int64

	Changed bool
}

type FTUsers struct {
	Uid     int64
	OnLine  bool
	Ready   bool
	UName   string `json:"uname"`
	UUrl    string `json:"uurl"`
	UGender int    `json:"ugender"`
	Ip      string `json:"ip"`
}

type HouseFloorTableTmp struct {
	NTId        int     // 序号
	TId         int     // 牌桌id
	Users       []int64 // 玩家
	CreateStamp int64   // 创建时间戳
	Begin       bool    //是否开始
	Step        int     //游戏局数
	Deleted     bool
}

func NewHFT(ucount, ntid int) *HouseFloorTable {
	hft := new(HouseFloorTable)
	hft.NTId = ntid
	hft.TId = 0
	hft.UserWithOnline = make([]FTUsers, ucount)
	hft.DataLock = new(lock2.RWMutex)
	hft.CreateStamp = time.Now().UnixNano()
	return hft
}

func NewFakeHFT(hid int64, ucount, ntid int) *HouseFloorTable {
	hft := new(HouseFloorTable)
	hft.NTId = ntid
	hft.TId = 0
	hft.UserWithOnline = make([]FTUsers, 0, ucount)
	for i := 0; i < ucount; i++ {
		robotName, robotUrl := static.GenRobotNameUrl(hid)
		if robotName == "" {
			xlog.Logger().Warn("NewFakeHFT: no robot name")
			return nil
		}
		hft.UserWithOnline = append(hft.UserWithOnline, FTUsers{
			Uid:     static.GenRobotId(),
			OnLine:  true,
			Ready:   true,
			UName:   robotName,
			UUrl:    robotUrl,
			UGender: static.GenGender(),
			Ip:      static.GenIPv4(),
		})
	}

	hft.DataLock = new(lock2.RWMutex)
	hft.CreateStamp = time.Now().UnixNano()
	hft.Begin = true
	hft.FakeEndAt = time.Now().Unix() + rand.Int63n(60) + 60*5
	hft.Changed = true
	return hft
}

type HFTableList struct {
	hfts []*HouseFloorTable
	less func(i, j *HouseFloorTable) bool
}

func (t *HFTableList) Len() int {
	return len(t.hfts)
}

func (t *HFTableList) Less(i, j int) bool {
	if t.less == nil {
		xlog.Logger().Errorln("house table sort error: less func is nil.")
		return false
	}
	return t.less(t.hfts[i], t.hfts[j])
}

func (t *HFTableList) Swap(i, j int) {
	t.hfts[i], t.hfts[j] = t.hfts[j], t.hfts[i]
}

// 锁定位置
func (self *HouseFloorTable) LockSeat(seat int, uid int64) bool {
	self.Changed = true
	if seat >= len(self.UserWithOnline) || seat < 0 {
		return false
	}
	if self.UserWithOnline[seat].Uid != 0 {
		return false
	}
	self.UserWithOnline[seat] = FTUsers{Uid: uid, OnLine: true}
	return true
}

// 解锁位置
func (self *HouseFloorTable) UnlockSeat(seat int, uid int64) bool {
	self.Changed = true
	if seat >= len(self.UserWithOnline) || seat < 0 {
		return false
	}
	if self.UserWithOnline[seat].Uid == 0 || self.UserWithOnline[seat].Uid != uid {
		return false
	}
	self.UserWithOnline[seat] = FTUsers{}
	return true
}

// 玩家座位
func (self *HouseFloorTable) GetUserSeat(uid int64) int {
	for seat, u := range self.UserWithOnline {
		if u.Uid == uid {
			return seat
		}
	}
	return -1
}

// 玩家数量
func (self *HouseFloorTable) UserCount() int {
	count := 0
	for _, u := range self.UserWithOnline {
		if u.Uid != 0 {
			count++
		}
	}
	return count
}

// 获取空座位
func (self *HouseFloorTable) GetEmptySeat() (int, bool) {
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.UserWithOnline[i].Uid == 0 {
			return i, true
		}
	}
	return -1, false
}

// UserSeat 获取座位然后坐下
func (self *HouseFloorTable) UserSeat(uid int64) int {
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.UserWithOnline[i].Uid == 0 {
			self.UserWithOnline[i].Uid = uid
			self.Changed = true
			return i
		}
	}
	return -1
}

// 清空数据
func (self *HouseFloorTable) Clear(hid int64) {
	self.TId = 0
	self.Begin = false
	self.Step = 0
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.FakeEndAt > 0 {
			static.RecycleRobotNameUrl(hid, self.UserWithOnline[i].UName)
		}
		self.UserWithOnline[i].Uid = 0
		self.UserWithOnline[i].OnLine = false
		self.UserWithOnline[i].Ready = false
		self.UserWithOnline[i].UName = ""
		self.UserWithOnline[i].UUrl = ""
		self.UserWithOnline[i].UGender = 0
		self.UserWithOnline[i].Ip = ""
	}
	self.Watchers = []int64{}
	self.FakeEndAt = 0

	self.Changed = true
}

// 释放无效的玩家
func (self *HouseFloorTable) InvalidUserFree(players ...int64) {
	self.DataLock.RLockWithLog()
	defer self.DataLock.RUnlock()
	isInArray := func(uid int64) bool {
		for _, v := range players {
			if v == uid {
				return true
			}
		}
		return false
	}
	for i := 0; i < len(self.UserWithOnline); {
		if self.UserWithOnline[i].Uid == 0 || !isInArray(self.UserWithOnline[i].Uid) {
			copy(self.UserWithOnline[i:], self.UserWithOnline[i+1:])
			self.UserWithOnline = self.UserWithOnline[:len(self.UserWithOnline)-1]
			continue
		}
		i++
	}
}

func (self *HouseFloorTable) UserSiteDown(uid int64) int {
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.UserWithOnline[i].Uid == 0 {
			self.UserWithOnline[i].Uid = uid
			return i
		}
	}
	return -1 //正常状况不会出现这个
}

func (self *HouseFloorTable) UserStandUp(uid int64) int {
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.UserWithOnline[i].Uid == uid {
			self.UserWithOnline[i].Uid = 0
			return i
		}
	}
	return -1 //正常状况不会出现这个
}

// 根据玩家上局对战玩家的uid判断该包厢桌子是否可以再来一局加入
func (self *HouseFloorTable) RestartIn(tid int, curUid int64 /*当前玩家*/, uids ...int64 /*当前玩家的上局对战玩家们*/) bool {
	table := GetTableMgr().GetTable(self.TId)
	if table == nil {
		return false
	}
	if table.IsBegin() {
		return false
	}
	if table.GetEmptySeat() < 0 {
		return false
	}

	for _, uid := range uids {
		if uid > 0 && uid != curUid {
			tUser := table.GetUser(uid)
			if tUser == nil {
				continue
			}
			if tUser.Uid == uid {
				// 20190829 hex 再来一局逻辑修改，只要桌上有上局的人就行，不必关心是否通过再来一局加入
				// if tUser.Restart {
				xlog.Logger().Debugf("[再来一局]在包厢牌桌初步找到符合再来一局的玩家%d", tUser.Uid)
				// 核实上局玩家
				record, err := GetDBMgr().GetDBrControl().SelectHRecordPlayers(tUser.Uid)
				if err != nil {
					xlog.Logger().Debugf("[再来一局]玩家%d.Redis操作错误：%v,", tUser.Uid, err)
					continue
				}
				if record.TId == tid {
					// 从该玩家的上局记录中找当前请求再来一局的玩家
					for _, ruid := range record.Users {
						if ruid > 0 && ruid == curUid {
							xlog.Logger().Debugf("[再来一局]当前包厢牌桌%d有符合再来一局的玩家%d", self.TId, tUser.Uid)
							return true
						}
					}
				}
				xlog.Logger().Errorf("[再来一局]在包厢牌桌%d.找到当前玩家%d.上局对战玩家%d,但其上局对战信息与之不匹配。", self.TId, curUid, tUser.Uid)
				// } else {
				// 	syslog.Logger().Infof("[再来一局]在包厢牌桌%d.找到当前玩家%d.上局对战的玩家%d,但其并不是通过再来一局加入牌桌。", self.TId, curUid, tUser.Uid)
				// }
			}
		}
	}
	return false
}

// 得到桌子实例
func (self *HouseFloorTable) GetTableInstance() *Table {
	if self.TId > 0 {
		return GetTableMgr().GetTable(self.TId)
	}
	return nil
}

// 清空数据
func (self *HouseFloorTable) UserOnlineChange(uid int64, online bool) {
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.UserWithOnline[i].Uid == uid {
			self.UserWithOnline[i].OnLine = online
			self.Changed = true
		}
	}
}
func (self *HouseFloorTable) UserReadyChange(uid int64, ready bool) {
	for i := 0; i < len(self.UserWithOnline); i++ {
		if self.UserWithOnline[i].Uid == uid {
			self.UserWithOnline[i].Ready = ready
			self.Changed = true
		}
	}
}

func (self *HouseFloorTable) AddWatcher(uid int64) {
	for _, wUid := range self.Watchers {
		if wUid == uid {
			return
		}
	}
	self.Watchers = append(self.Watchers, uid)
	self.Changed = true
}

func (self *HouseFloorTable) RemoveWatcher(uid int64) {
	for index, wUid := range self.Watchers {
		if wUid == uid {
			self.Watchers = append(self.Watchers[:index], self.Watchers[index+1:]...)
			self.Changed = true
			return
		}
	}
}

// //////////////////////////////////////////////////////////////////////////////
// ! 包厢楼层数据结构
type HouseFloor struct {
	Id   int64
	HId  int          //! 历史包厢
	DHId int64        //! 历史包厢
	Rule static.FRule //! 楼层规则

	IsAlive bool //! 活跃

	MemAct  map[int64]*HouseMember //! 活跃玩家
	MemLock *lock2.RWMutex

	DataLock *lock2.RWMutex
	Tables   map[int]*HouseFloorTable

	Name  string // 用户自定义名称
	IsMix bool   // 是否是混排大厅
	IsVip bool   // 是否为混排楼层
	static.FloorVitaminOptions
	SyncTime        int64
	TotalTableCount int
	MaxMatchCount   int //已开始桌子数
	AiSuperNum      int `json:"ai_super_num"`
	HouseFloorLiveData
	IsHide       bool `json:"ishide"`
	IsCapSetVip  bool // 队长是否可以设置VIP
	IsDefJoinVip bool // 新入圈的玩家 是否 默认加入VIP  true 加入 false不加入
	MinTable     int
	MaxTable     int
}

type HouseFloorLiveData struct {
	LastMappingNum   int
	LastMappingTotal int
}

func (hf *HouseFloor) Init() {
	hf.MemAct = make(map[int64]*HouseMember)
	hf.MemLock = new(lock2.RWMutex)
	hf.Tables = make(map[int]*HouseFloorTable)
	hf.DataLock = new(lock2.RWMutex)
	hf.VitaminLowLimit = consts.VitaminInvalidValueSrv
	hf.VitaminHighLimit = consts.VitaminInvalidValueSrv
	hf.VitaminLowLimitPause = consts.VitaminInvalidValueSrv
}

// 广播
func (hf *HouseFloor) BroadCast(role int, header string, data interface{}) {
	hf.MemLock.RLockWithLog()
	defer func() {
		hf.MemLock.RUnlock()
	}()

	for _, mem := range hf.MemAct {
		if mem == nil {
			continue
		}
		if mem.Lower(role) {
			continue
		}
		SendPersonMsg(mem.UId, header, data)
	}

	return
}
func SendPersonMsg(uid int64, head string, data interface{}) {
	p := GetPlayerMgr().GetPlayer(uid)
	if p == nil {
		return
	}
	// 在线
	p.SendMsg(head, data)
}

func (hf *HouseFloor) BroadCastMix(role int, head string, data interface{}) {
	house := GetClubMgr().GetClubHouseById(hf.DHId)
	if house == nil {
		return
	}
	if hf.IsMix && house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				floor.BroadCast(role, head, data)
			}
		}
	} else {
		hf.BroadCast(role, head, data)
	}
}

// BroadCastMuteMsg 广播两条消息，先执行head1
func (hf *HouseFloor) BroadCastMuteMsg(role int, head1, head2 string, data1, data2 interface{}) {
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if hf.IsMix && house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				floor.BroadCast(role, head1, data1)
			}
		}
		for _, floor := range house.Floors {
			if floor.IsMix {
				floor.BroadCast(role, head2, data2)
			}
		}
	} else {
		hf.BroadCast(role, head1, data1)
		hf.BroadCast(role, head2, data2)
	}
}

// 离开包厢楼层
func (hf *HouseFloor) MemOut(uid int64) {

	// 查找玩家是否进入其余楼层
	hf.SafeMemOut(uid)

	// 当前玩家数据
	person := GetPlayerMgr().GetPlayer(uid)
	if person == nil {
		return
	}

	if person.Info.HouseId == hf.HId && person.Info.FloorId == hf.Id {
		// DBr缓存数据
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(uid, "HouseId", 0, "FloorId", 0)

		// 内存数据
		person.Info.HouseId = 0
		person.Info.FloorId = 0
	}
}

// MemIn 带锁进入楼层
func (hf *HouseFloor) MemIn(mem *HouseMember) {
	hf.MemLock.CustomLock()
	defer hf.MemLock.CustomUnLock()
	hf.MemAct[mem.UId] = mem
	return
}

// SafeMemOut 带锁退出楼层
func (hf *HouseFloor) SafeMemOut(uid int64) {
	hf.MemLock.CustomLock()
	defer hf.MemLock.CustomUnLock()
	delete(hf.MemAct, uid)
	return

}

// ModifUserNum 修改楼层牌桌人数
func (hf *HouseFloor) ModifUserNum() {
	hf.DataLock.Lock()
	defer hf.DataLock.Unlock()
	for _, ftb := range hf.Tables {
		ftb.DataLock.RLockWithLog()
		if ftb.UserCount() > 0 {
			ftb.DataLock.RUnlock()
			continue
		}
		ftb.UserWithOnline = make([]FTUsers, hf.Rule.PlayerNum)
		ftb.DataLock.RUnlock()
	}
}

// GetTableUser 根据楼层号获取桌上玩家id
func (hf *HouseFloor) GetTableUser(hft *HouseFloorTable, u int64) []int64 {
	inUids := []int64{}
	for _, uid := range hft.UserWithOnline {
		if uid.Uid > 0 && uid.Uid != u {
			inUids = append(inUids, uid.Uid)
		}
	}
	return inUids
}

// GetHFT 获取包厢桌子
func (hf *HouseFloor) GetHFT(ntid int) *HouseFloorTable {
	hf.DataLock.RLockWithLog()
	defer hf.DataLock.RUnlock()
	return hf.Tables[ntid]
}

// GetHFM 获取包厢用户
func (hf *HouseFloor) GetHFM(uid int64) *HouseMember {
	hf.MemLock.RLockWithLog()
	defer hf.MemLock.RUnlock()
	return hf.MemAct[uid]
}

// TableLimitMsg 禁止同桌检查
func (hf *HouseFloor) TableLimitMsg(uid int64, hft *HouseFloorTable) (int16, string) {
	inUids := hf.GetTableUser(hft, uid)
	if inUids == nil {
		return xerrors.SuccessCode, ""
	}
	uids := make([]int64, 0, 4)
	uids = append(uids, uid)
	uids = append(uids, inUids...)
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		return xerrors.SuccessCode, ""
	}
	e, notAllowUids := house.CheckUsersAllowSameTable(uids...)
	if e != nil {
		var msg []string
		for _, uid := range notAllowUids {
			dmem, err := GetDBMgr().GetDBrControl().GetPerson(uid)
			if err != nil {
				xlog.Logger().Errorln(err)
				continue
			}
			msg = append(msg, dmem.Nickname)
		}
		var respMsg string
		switch len(msg) {
		case 0: //假如没有获取到用户昵称
			respMsg = "您暂时无法与该桌用户同桌游戏，如有疑问请联系盟主"
		case 1:
			respMsg = fmt.Sprintf("您暂时无法与%s同桌游戏，如有疑问请联系盟主", msg[0])
		case 2:
			respMsg = fmt.Sprintf("您暂时无法与%s,%s同桌游戏，如有疑问请联系盟主", msg[0], msg[1])
		case 3:
			respMsg = fmt.Sprintf("您暂时无法与%s,%s,%s同桌游戏，如有疑问请联系盟主", msg[0], msg[1], msg[2])
		default: //超过三人
			respMsg = fmt.Sprintf("您暂时无法与%s,%s,%s等用户同桌游戏，如有疑问请联系盟主", msg[0], msg[1], msg[2])
		}
		return xerrors.HouseTableLimitJoin.Code, respMsg
	}
	return xerrors.SuccessCode, ""
}

func (hf *HouseFloor) TableLogger() *logrus.Entry {
	return xlog.ELogger(ElkHouseTbl).WithFields(map[string]interface{}{
		"hid": hf.HId,
		"fid": hf.Id,
	})
}

func (hf *HouseFloor) checkTableIn(house *Club, uid int64, gpsInfo *static.GpsInfo, tableUser []int64) (bool, *xerrors.XError) {
	if len(tableUser) > 0 {
		ids := make([]int64, 0)
		ids = append(ids, uid)
		ids = append(ids, tableUser...)

		if e, _ := house.CheckUsersAllowSameTable(ids...); e != nil {
			hf.TableLogger().Errorf("[禁止同桌]信息:%s", e.Error())
			return false, nil
		}

		if hf.Rule.Restrict == "true" {
			if e := CheckUserIpById(gpsInfo.Ip, tableUser); e != nil {
				hf.TableLogger().Errorf("[ip相同]玩家A：%d 信息:%s", uid, e.Error())
				return false, nil
			}
			if e := CheckUserGps(house, gpsInfo.Longitude, gpsInfo.Latitude, tableUser); e != nil {
				hf.TableLogger().Errorf("[Gps限制]玩家A：%d 信息:%s", uid, e.Error())
				return false, nil
			}
		}

		// 如果开启了智能防作弊且当前包厢桌子数量大于或等于设置值，则检查作弊玩家
		if house.DBClub.MixActive && house.DBClub.AICheck && hf.IsMix {
			if !GetServer().ConHouse.MixAIAble || house.GetTabOnlineCounts() > GetServer().ConHouse.MixAITableNum {
				if ok, xe := house.CheckUsersCribbersSameTable(ids...); !ok {
					hf.TableLogger().Error(xe)
					return false, xe
				}
			}
		}
	}
	return true, nil
}

func (hf *HouseFloor) CheckTableIn(house *Club, hft *HouseFloorTable, uid int64, gpsInfo *static.GpsInfo, max bool, want int, ignoreAtomic bool) (seat int, cusErr *xerrors.XError) {
	// hf.TableLogger().Infof("CheckTableIn：%d  hft: %+v", uid, hft)
	ok := atomic.CompareAndSwapInt64(&hft.IsOccupy, 0, 1)
	if !ok {
		if ignoreAtomic {
			// 如果过滤掉取不到原子操作的玩家，则返回 -1 本次过滤这张桌子
			return InvalidHouseTableSeat, nil
		} else {
			// 如果不允许过滤，则在指定时间内取指定次数，如果一直取不到 则认为入桌失败
			tryNum := 5
			delayTime := time.Millisecond * 200
			for {
				tryNum--
				time.Sleep(delayTime)
				ok = atomic.CompareAndSwapInt64(&hft.IsOccupy, 0, 1)
				if ok {
					break
				}
				if tryNum < 0 {
					return 0, xerrors.HouseFloorTableJoiningError
				}
			}
		}
	}

	hft.DataLock.Lock()
	HftTicker(hf, hft)
	defer func() {
		if seat == InvalidHouseTableSeat {
			hft.DataLock.Unlock()
			hft.LockClose <- struct{}{}
			hft.IsOccupy = 0
		}
	}()

	if hft.Begin || hft.FakeEndAt > 0 {
		return InvalidHouseTableSeat, nil
	}

	tableUser := make([]int64, 0)
	for _, user := range hft.UserWithOnline {
		if user.Uid > 0 {
			tableUser = append(tableUser, user.Uid)
		}
	}

	tableUserCount := len(tableUser)

	if tableUserCount >= hf.Rule.PlayerNum {
		return InvalidHouseTableSeat, nil
	}

	if max {
		if tableUserCount != want {
			return InvalidHouseTableSeat, nil
		}
	}

	ok, cusErr = hf.checkTableIn(house, uid, gpsInfo, tableUser)
	if !ok {
		return InvalidHouseTableSeat, cusErr
	}

	for i := 0; i < hf.Rule.PlayerNum; i++ {
		if i < len(hft.UserWithOnline) {
			if hft.UserWithOnline[i].Uid == 0 {
				key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_TABLE_LOCK, hf.Id, hft.NTId, i)
				if GetDBMgr().Redis.Exists(key).Val() == 1 {
					if GetDBMgr().Redis.Get(key).Val() != fmt.Sprintf("%d", uid) {
						continue
					}
				}
				hft.UserWithOnline[i] = FTUsers{Uid: uid, OnLine: true}
				// 入桌校验通过
				hf.TableLogger().Infof("[玩家入桌]玩家: %d 入座成功", uid)
				hft.Changed = true
				return i, nil
			}
		} else {
			hft.UserWithOnline = append(hft.UserWithOnline, FTUsers{Uid: uid, OnLine: true})
			return i, nil
		}
	}
	hf.TableLogger().Errorf("[入座异常]玩家：%d 座子信息：%+v", uid, hft)
	return InvalidHouseTableSeat, nil
}

// GetQuickTable 快速入桌
func (hf *HouseFloor) GetQuickTable(p *static.Person, gpsInfo *static.GpsInfo, untid int, restartTid int, priorityUsers ...int64 /*指定匹配优先级高的人*/) (hft *HouseFloorTable, seat int, cusErr *xerrors.XError) {
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		return
	}
	hf.TableLogger().Infof("[Join Entrance] table num=%d", len(hf.Tables))
	if house.DBClub.MixActive && house.DBClub.TableJoinType != consts.SelfAdd /* && len(hf.Tables) == 0 */ {
		hf.TableLogger().Infof("[Join Entrance] CheckEmptyAndAdd")
		hf.CheckEmptyAndAdd()
	}
	if house.IsAiSuper() && hf.IsMix && hf.AiSuperNum > 0 {
		hf.TableLogger().Infof("[Join Entrance] CheckEligibilityAndAdd")
		hf.CheckEligibilityAndAdd(house, p.Uid, gpsInfo)
	}
	switch untid {
	case consts.HOUSETABLEINQUICK:
		return hf.JoinQuick(house, p, gpsInfo, untid, restartTid, priorityUsers...)
	case consts.HOUSETABLEINAGAIN:
		return hf.JoinAntherGame(house, p, gpsInfo, untid, restartTid, priorityUsers...)
	case consts.HOUSETABLEINMIXIN:
		return hf.JoinMixAutoEntrance(house, p, gpsInfo, untid, restartTid, priorityUsers...)
	default:
		return hf.JoinNormal(house, p, gpsInfo, untid, restartTid, priorityUsers...)
	}
}

// 存在活动
func (hf *HouseFloor) ExistActivity() bool {
	arr, err := GetDBMgr().GetDBrControl().HouseActivityList(hf.DHId, true)
	if err != nil {
		xlog.Logger().Errorln(err)
		return false
	}

	for _, act := range arr {
		fids := GetClubMgr().GetHouseFidsByStr(act.FId)
		for _, fid := range fids {
			if fid == hf.Id {
				return true
			}
		}
	}

	return false
}

// 存在没有结束的活动
func (hf *HouseFloor) ExistActivityNoFinish() bool {
	arr, err := GetDBMgr().GetDBrControl().HouseActivityList(hf.DHId, true)
	if err != nil {
		xlog.Logger().Errorln(err)
		return false
	}

	for _, act := range arr {
		fids := GetClubMgr().GetHouseFidsByStr(act.FId)
		for _, fid := range fids {
			if fid == hf.Id && time.Now().Unix() < act.EndTime {
				return true
			}
		}
	}

	return false
}

// 修改玩法
func (hf *HouseFloor) KindIdModify(frule static.FRule) *xerrors.XError {

	// 规则合法判定
	var msgct static.Msg_CreateTable
	msgct.KindId = frule.KindId
	msgct.PlayerNum = frule.PlayerNum
	msgct.RoundNum = frule.RoundNum
	msgct.CostType = frule.CostType
	msgct.Restrict = frule.Restrict
	msgct.GameConfig = frule.GameConfig
	msgct.Gps = frule.Gps
	msgct.GVoice = frule.GVoice
	msgct.GVoiceOk = frule.GVoiceOk

	_, custerr := validateCreateTableParam(&msgct, true)
	if custerr != nil {
		return custerr
	}

	hf.Rule = frule

	err := GetDBMgr().HouseFloorUpdate(hf)
	if err != nil {
		return xerrors.DBExecError
	}

	ntf := new(static.Ntf_HC_HouseFloorKIdModify)
	ntf.HId = hf.HId
	ntf.FId = hf.Id
	ntf.FRule = frule
	go hf.BroadCastMix(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorRuleModify_Ntf, ntf)

	return nil
}

// 获得牌桌
func (hf *HouseFloor) GetTableByNTId(id int) *HouseFloorTable {
	hf.DataLock.RLockWithLog()
	defer hf.DataLock.RUnlock()

	return hf.Tables[id]
}

func (hf *HouseFloor) GetTableByTId(tid int) *HouseFloorTable {
	hf.DataLock.RLockWithLog()
	defer hf.DataLock.RUnlock()

	for _, hft := range hf.Tables {
		if hft.TId == tid {
			return hft
		}
	}
	return nil
}

// 获得桌子数据
func (hf *HouseFloor) MemTableBaseInfo(table *Table) *static.Msg_HouseTableItem {

	if table == nil {
		return nil
	}

	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		return nil
	}

	tBaseInfo := new(static.Msg_HouseTableItem)
	tBaseInfo.NTId = table.NTId
	tBaseInfo.TId = table.Id
	tBaseInfo.ATId = table.Id
	tBaseInfo.Step = table.Step
	tBaseInfo.Begin = table.Begin
	tBaseInfo.TRule.KindId = table.KindId
	tBaseInfo.TRule.PlayerNum = table.Config.MaxPlayerNum
	tBaseInfo.TRule.RoundNum = table.Config.RoundNum
	tBaseInfo.TRule.Difen = table.Config.Difen
	tBaseInfo.TMemItems = make([]static.FloorTableUser, 0)

	table.lock.RLockWithLog()
	for _, u := range table.Users {
		if u == nil {
			continue
		}
		person, err := GetDBMgr().GetDBrControl().GetPerson(u.Uid)
		if err != nil {
			table.lock.RUnlock()
			return nil
		}
		nmem := new(static.FloorTableUser)
		nmem.UId = person.Uid
		nmem.UName = person.Nickname
		nmem.UUrl = person.Imgurl
		nmem.UGender = person.Sex

		tBaseInfo.TMemItems = append(tBaseInfo.TMemItems, *nmem)
	}
	table.lock.RUnlock()

	return tBaseInfo
}

func (hf *HouseFloor) Rename(name string) error {
	sql := `update house_floor set name = ? where id = ? and hid = ? limit 1`
	err := GetDBMgr().GetDBmControl().Exec(sql, name, hf.Id, hf.DHId).Error
	if err != nil {
		return err
	}
	hf.Name = name
	err = GetDBMgr().HouseFloorUpdate(hf)
	if err != nil {
		return xerrors.DBExecError
	}
	return err
}

// 获取并锁定条件牌桌
func (hf *HouseFloor) GetTableByQuick() int {
	hf.DataLock.RLockWithLog()
	defer hf.DataLock.RUnlock()

	maxnum := hf.Rule.PlayerNum
	tnum := GetServer().ConHouse.TableNum
	for num := 1; num <= maxnum; num++ {

		for ntid := 0; ntid < tnum; ntid++ {
			hft := hf.Tables[ntid]

			// 无人寻找空桌子
			if hft.TId == 0 && num == maxnum {
				return ntid
			}
			if hft.UserCount() == 0 {
				continue
			}
			// 牌桌最大人数
			if len(hft.UserWithOnline) > maxnum {
				maxnum = len(hft.UserWithOnline)
			}
			// 牌桌最优人数
			if len(hft.UserWithOnline)-num <= 0 {
				continue
			}
			//牌桌匹配
			if len(hft.UserWithOnline)-num == hft.UserCount() {
				return ntid
			}
		}
	}

	return -1
}

func (hf *HouseFloor) Delete(wg *sync.WaitGroup) error {
	// DB
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	err := GetDBMgr().HouseFloorDelete(hf, nil)
	if err != nil {
		hf.IsAlive = true
		xlog.Logger().Errorln(err)
		return err
	}
	// 包厢
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house != nil {
		house.FloorLock.CustomLock()
		delete(house.Floors, hf.Id)
		house.FloorLock.CustomUnLock()
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseFloorDelete)
	ntf.HId = hf.HId
	ntf.FId = hf.Id

	if wg == nil {
		go house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorDelete_Ntf, ntf)
	}
	hf.MemAct = make(map[int64]*HouseMember, 0)
	hf.Tables = make(map[int]*HouseFloorTable, 0)
	hf.DeleteHft()
	return nil
}

func (hf *HouseFloor) DeleteHft() error {
	if !hf.IsMix {
		return nil
	}
	cli := GetDBMgr().Redis
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR_MIX, hf.Id)
	xlog.Logger().Errorf("deleted floor:%s", key)
	cli.Del(key)
	sql := `delete from house_floor_mixtable where fid = ? `
	db := GetDBMgr().GetDBmControl()
	db.Exec(sql, hf.Id)
	return nil
}

func (hf *HouseFloor) SaveHft() error {
	if !hf.IsMix {
		return nil
	}
	cli := GetDBMgr().Redis
	key1 := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR_MIX, hf.Id)
	if len(hf.Tables) == 0 {
		cli.Del(key1)
		return nil
	}

	tmpMap := make(map[string]interface{}, len(hf.Tables))
	for _, hft := range hf.Tables {
		if hft.FakeEndAt > 0 {
			continue
		}
		buf, err := json.Marshal(hft)
		if err != nil {
			xlog.Logger().Errorf("json marshal error:%v", err)
			continue
		}
		tmpMap[fmt.Sprintf("%d", hft.NTId)] = buf
	}
	pip := cli.Pipeline()
	pip.Del(key1)
	pip.HMSet(key1, tmpMap)
	pip.Exec()
	return nil
}

func (hf *HouseFloor) ReloadHft() error {
	cli := GetDBMgr().Redis
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR_MIX, hf.Id)
	res := cli.HGetAll(key).Val()
	for _, v := range res {
		hft := HouseFloorTable{}
		err := json.Unmarshal([]byte(v), &hft)
		if err != nil {
			hftTmp := HouseFloorTableTmp{}
			err := json.Unmarshal([]byte(v), &hftTmp)
			if err != nil {
				xlog.Logger().Errorf("json error:%v,%s", err, v)
				continue
			}
			hft.NTId = hftTmp.NTId
			hft.TId = hftTmp.TId
			hft.UserWithOnline = make([]FTUsers, len(hftTmp.Users))
			for _, uid := range hftTmp.Users {
				hft.UserWithOnline = append(hft.UserWithOnline, FTUsers{Uid: uid})
			}
			hft.CreateStamp = hftTmp.CreateStamp
			hft.Begin = hftTmp.Begin
			hft.Step = hftTmp.Step
			hft.Deleted = hftTmp.Deleted
		}
		hft.DataLock = new(lock2.RWMutex)
		if hft.TId > 0 {
			table := GetTableMgr().GetTable(hft.TId)
			if table == nil {
				hft.TId = 0
				hft.UserWithOnline = make([]FTUsers, hf.Rule.PlayerNum)
			}
		} else {
			hft.UserWithOnline = make([]FTUsers, hf.Rule.PlayerNum)
		}
		hf.Tables[hft.NTId] = &hft
	}
	return nil
}

func (hf *HouseFloor) SaveHftRedisToDB() error {
	if !hf.IsMix {
		return nil
	}
	cli := GetDBMgr().Redis
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR_MIX, hf.Id)
	res := cli.HGetAll(key).Val()
	if len(res) == 0 {
		return nil
	}
	sql := `delete from house_floor_mixtable where fid = ? `
	db := GetDBMgr().GetDBmControl()
	db.Exec(sql, hf.Id)
	sql = `insert into house_floor_mixtable(hid,fid,ntid,hft_info) values(?,?,?,?)`
	for k, v := range res {
		err := db.Exec(sql, hf.DHId, hf.Id, k, v).Error
		if err != nil {
			xlog.Logger().Errorf("insert into housemixfloor_table error :%v", err)
			continue
		}
	}
	return nil
}

func (hf *HouseFloor) GetHftByTid(tid int) *HouseFloorTable {
	hf.DataLock.RLockWithLog()
	defer hf.DataLock.RUnlock()
	for _, h := range hf.Tables {
		if h.TId == tid {
			return h
		}
	}
	return nil
}

func (hf *HouseFloor) RedisPub(head string, msg interface{}) error {
	// cli1 := GetDBMgr().Redis
	cli2 := GetDBMgr().PubRedis
	buf, err := json.Marshal(msg)
	if err != nil {
		xlog.Logger().Errorf("json mas err:%s", buf)
		return err
	}
	pubData := static.MsgRedisSub{}
	pubData.Head = head
	pubData.Data = fmt.Sprintf("%s", buf)
	buf, err = json.Marshal(pubData)
	if err != nil {
		xlog.Logger().Errorf("json mas err:%s", buf)
		return err
	}
	// err = cli1.Publish(hf.PubKey(), buf).Err()
	// if err != nil {
	// 	syslog.Logger().Errorf("pub error:%v", err)
	// 	// return err
	// }
	err = cli2.Publish(hf.PubKey(), buf).Err() //兼容不更新的游戏服
	if err != nil {
		xlog.Logger().Errorf("pub error:%v", err)
		return err
	}
	return err
}

func (hf *HouseFloor) PubKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, hf.Id)
}

func (hf *HouseFloor) RedisPubMulti(data map[string]interface{}) error {
	for k, v := range data {
		err := hf.RedisPub(k, v)
		if err != nil {
			xlog.Logger().Errorf("redis pub error:%v", err)
			return err
		}
	}
	return nil
}

//func (hf *HouseFloor) ConfiguredDeduct() bool {
//	return hf.ConfiguredGameDeduct() || hf.ConfiguredHighestDeduct() || hf.ConfiguredLowestDeduct()
//}

func (hf *HouseFloor) ConfiguredEffect() bool {
	return hf.ConfiguredLowLimit() || hf.ConfiguredLowLimitPause()
}

func (hf *HouseFloor) CheckInGameTable() bool {
	hf.DataLock.Lock()
	defer hf.DataLock.Unlock()
	for _, t := range hf.Tables {
		if t == nil {
			continue
		}
		if t.TId > 0 {
			if len(t.UserWithOnline) > 0 {
				return true
			}
		}
	}
	return false
}

func (hf *HouseFloor) StartSync() {
	ticker := time.NewTicker(time.Second)
loop:
	for hf.IsAlive {
		select {
		case <-ticker.C:
			err := hf.Sync()
			if err != nil {
				xlog.Logger().Errorf("sync error:%v", err)
				break loop
			}
		}
	}
	ticker.Stop()
	xlog.Logger().Warnf("sync stop!")
}

func (hf *HouseFloor) Sync() error {
	if !hf.IsAlive {
		return errors.New("floor close")
	}
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		return errors.New("house not exists")
	}
	if house.DBClub.MixActive && hf.IsMix {
		return errors.New("floor is mix")
	}

	if len(hf.MemAct) == 0 {
		return nil
	}
	var acks static.MixFloorAcks

	acks.FIDS = append(acks.FIDS, hf.Id)

	hf.SyncTime = time.Now().UnixNano()
	ack := hf.GetDiffInfo()
	if ack != nil {
		acks.Acks = append(acks.Acks, ack)
	}
	if len(acks.Acks) > 0 {
		hf.PushTabChange(acks)
	}
	return nil
}

func (hf *HouseFloor) GetTableInfo() *static.Msg_HC_HouseMemberIn {
	if !hf.IsAlive {
		return nil
	}
	backInfo := hf.GetAllTables()
	return hf.BuildAck(backInfo, true)
}

func (hf *HouseFloor) GetDiffInfo() *static.Msg_HC_HouseMemberIn {
	if !hf.IsAlive {
		return nil
	}
	backInfo := make(map[int]*HouseFloorTable)

	backInfo, hf.TotalTableCount, hf.MaxMatchCount = hf.GetChangeTable()
	return hf.BuildAck(backInfo, true)
}

func (hf *HouseFloor) GetAllTables() map[int]*HouseFloorTable {
	hf.DataLock.RLock()
	defer hf.DataLock.RUnlock()
	dest := make(map[int]*HouseFloorTable)
	for _, table := range hf.Tables {
		dest[table.NTId] = table
	}
	return dest
}

func (hf *HouseFloor) GetChangeTable() (map[int]*HouseFloorTable, int, int) {
	hf.DataLock.Lock()
	defer hf.DataLock.Unlock()
	dest := make(map[int]*HouseFloorTable)
	var total, maxWait int
	for _, table := range hf.Tables {
		table.DataLock.Lock()
		if table.Begin {
			total++
		} else {
			if matchUser := table.UserCount(); matchUser < hf.Rule.PlayerNum && matchUser > maxWait {
				maxWait = matchUser
			}
		}
		if table.Changed {
			dest[table.NTId] = table
			table.Changed = false
		}
		table.DataLock.Unlock()
	}
	return dest, total, maxWait
}

func compireSlice(s1, s2 []FTUsers) bool {
	if len(s1) != len(s2) {
		return false
	}
	for k, uid := range s1 {
		if uid.Uid != s2[k].Uid || uid.OnLine != s2[k].OnLine || uid.Ready != s2[k].Ready {
			return false
		}
	}
	return true
}

func compireIcon(s1, s2 []int64) bool {
	if len(s1) != len(s2) {
		return false
	}
	for k, icon := range s1 {
		if icon != s2[k] {
			return false
		}
	}
	return true
}

func (hf *HouseFloor) BuildAck(newInfo map[int]*HouseFloorTable, sync bool) *static.Msg_HC_HouseMemberIn {
	if len(newInfo) == 0 {
		return nil
	}
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		return nil
	}
	var onlyBegin bool
	if house.DBClub.MixActive && house.DBClub.TableJoinType == consts.NoCheat && hf.IsMix {
		onlyBegin = true
	}
	ack := new(static.Msg_HC_HouseMemberIn)
	ack.FId = hf.Id
	ack.FTableItems = make([]static.Msg_HouseTableItem, 0, len(newInfo))
	ack.MaxMatchNum = hf.GetMaxMatchCount()
	ack.TotalNum = hf.GetInGameTableCount()

	nowUnix := time.Now().Unix()

	for _, table := range newInfo {
		if onlyBegin {
			if !table.Begin {
				matchUser := table.UserCount()
				if matchUser < hf.Rule.PlayerNum && matchUser > ack.MaxMatchNum {
					ack.MaxMatchNum = matchUser
				}
				if !sync {
					continue
				}
			}
		}
		var titem static.Msg_HouseTableItem
		titem.NTId = table.NTId
		titem.TId = table.TId
		titem.ATId = table.TId
		titem.TMemItems = make([]static.FloorTableUser, 0)
		titem.Deleted = table.Deleted

		if table.FakeEndAt > 0 {
			titem.TRule.KindId = hf.Rule.KindId
			titem.TRule.PlayerNum = hf.Rule.PlayerNum
			titem.TRule.RoundNum = hf.Rule.RoundNum
			titem.TRule.Difen = hf.Rule.Difen
			titem.CanWatch = IsGameSupportWatch(hf.Rule.KindId)
			titem.WatcherIcons = []string{}
			for _, watcher := range table.Watchers {
				lazyUser := GetLazyUser(watcher)
				if lazyUser != nil {
					titem.WatcherIcons = append(titem.WatcherIcons, lazyUser.ImageUrl)

					if len(titem.WatcherIcons) >= 4 {
						break
					}
				}
			}
			titem.Begin = true
			createdAt := table.CreateStamp / 1e9
			exsitSec := int(nowUnix - createdAt)
			singleRoundSec := int(table.FakeEndAt-table.CreateStamp/1e9) / hf.Rule.RoundNum
			titem.Step = exsitSec / singleRoundSec
			if titem.Step <= 0 {
				titem.Step = 1
			}
			if titem.Step > hf.Rule.RoundNum {
				titem.Step = hf.Rule.RoundNum
			}

			for _, uid := range table.UserWithOnline {
				if uid.Uid == 0 {
					continue
				}

				var mitem static.FloorTableUser

				mitem.UId = uid.Uid
				mitem.UName = uid.UName
				mitem.UUrl = uid.UUrl
				if static.IsAnonymous(hf.Rule.GameConfig) {
					mitem.UName = "匿名昵称"
					mitem.UUrl = ""
				}

				mitem.UGender = uid.UGender

				mitem.IsOnline = uid.OnLine
				mitem.Ready = uid.Ready
				titem.TMemItems = append(titem.TMemItems, mitem)
			}
			ack.FTableItems = append(ack.FTableItems, titem)
			continue
		}

		t := GetTableMgr().GetTable(table.TId)
		if t == nil {
			titem.TRule.KindId = hf.Rule.KindId
			titem.TRule.PlayerNum = hf.Rule.PlayerNum
			titem.TRule.RoundNum = hf.Rule.RoundNum
			titem.TRule.Difen = hf.Rule.Difen
			titem.CanWatch = IsGameSupportWatch(hf.Rule.KindId)
			ack.FTableItems = append(ack.FTableItems, titem)
			continue
		}
		if static.IsAnonymous(t.Config.GameConfig) {
			titem.TId = 0
		}
		titem.TRule.KindId = t.KindId
		titem.TRule.PlayerNum = t.Config.MaxPlayerNum
		titem.TRule.RoundNum = t.Config.RoundNum
		titem.TRule.Difen = hf.Rule.Difen
		titem.WatcherIcons = []string{}
		for _, watcher := range table.Watchers {
			lazyUser := GetLazyUser(watcher)
			if lazyUser != nil {
				titem.WatcherIcons = append(titem.WatcherIcons, lazyUser.ImageUrl)

				if len(titem.WatcherIcons) >= 4 {
					break
				}
			}
		}

		titem.Begin = t.Begin
		titem.Step = t.Step

		for _, uid := range table.UserWithOnline {
			if uid.Uid == 0 {
				continue
			}
			lazyUser := GetLazyUser(uid.Uid)

			var mitem static.FloorTableUser

			mitem.UId = lazyUser.Uid
			mitem.UName = lazyUser.Name
			mitem.UUrl = lazyUser.ImageUrl
			if static.IsAnonymous(t.Config.GameConfig) {
				mitem.UName = "匿名昵称"
				mitem.UUrl = ""
			}

			mitem.UGender = lazyUser.Sex

			mitem.IsOnline = uid.OnLine
			mitem.Ready = uid.Ready
			titem.TMemItems = append(titem.TMemItems, mitem)
		}
		ack.FTableItems = append(ack.FTableItems, titem)
	}
	return ack
}

func (hf *HouseFloor) PushTabChange(acks static.MixFloorAcks) {
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		return
	}
	hf.MemLock.RLock()
	for _, mem := range hf.MemAct {
		mem.SendMsg(consts.MsgHouseFloorTableSync, acks)
	}
	hf.MemLock.RUnlock()
	hf.RedisPub(consts.MsgHouseFloorTableSync, acks)
	return
}

func (hf *HouseFloor) CheckEmptyAndAdd() {
	house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		xlog.Logger().Errorf("nil house:%d", hf.DHId)
		return
	}

	var empty int

	hf.DataLock.Lock()

	for _, hft := range hf.Tables {
		if hft.TId == 0 {
			if hft.Begin {
				hft.Clear(hf.DHId)
			}
			empty++
		}
	}

	if len(hf.Tables) >= GetServer().ConHouse.TableNum {
		hf.DataLock.Unlock()
		return
	}

	hf.TableLogger().Infof("current empty table count:%d", empty)

	if empty > 1 {
		hf.DataLock.Unlock()
		return
	}

	hf.Tables[999] = NewHFT(hf.Rule.PlayerNum, 999) //桌子ntid为乱序，防止覆盖其他桌子，ntid在请求楼层信息会重新排序
	hf.DataLock.Unlock()
	var mixFloor []*HouseFloor
	fids := []int64{}
	if house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				mixFloor = append(mixFloor, floor)
				fids = append(fids, floor.Id)
			}
		}
		if sortTableByCreateTime(mixFloor) {
			xlog.Logger().Warnf("sortTableByCreateTime after add")
			house.SyncTablesWithSorted()
		}
	}
	return
}

func (hf *HouseFloor) CheckRunningAndAdd(house *Club) {
	//house := GetClubMgr().GetClubHouseByHId(hf.HId)
	if house == nil {
		xlog.Logger().Errorf("nil house:%d", hf.DHId)
		return
	}

	var realNum int
	var fakeNum int

	hf.DataLock.Lock()

	for _, hft := range hf.Tables {
		if hft.FakeEndAt > 0 && time.Now().Unix() >= hft.FakeEndAt {
			if tbl := GetTableMgr().GetTable(hft.TId); tbl != nil {
				GetTableMgr().DelTable(tbl)
			}
			hft.Clear(hf.DHId)
			continue
		}
		if hft.TId == 0 && hft.Begin {
			hft.Clear(hf.DHId)
			continue
		}
		if hft.TId > 0 {
			if hft.FakeEndAt > 0 {
				fakeNum++
			} else {
				realNum++
			}
		}
	}

	if house.DBClub.IsFrozen {
		hf.DataLock.Unlock()
		return
	}
	minTableNum := house.DBClub.MinTableNum
	maxTableNum := house.DBClub.MaxTableNum
	if minTableNum <= 0 {
		hf.DataLock.Unlock()
		return
	}

	hasNum := realNum + fakeNum

	// xlog.Logger().Warnf("current has table count:%d, real:%d, fake:%d", hasNum, realNum, fakeNum)

	if hasNum >= maxTableNum {
		hf.DataLock.Unlock()
		return
	}

	if hasNum >= minTableNum {
		// 区间内
		if rand.Intn(100) >= 50 {
			hf.DataLock.Unlock()
			return
		}
	}

	// 先找一张空桌子
	var (
		hft   *HouseFloorTable
		found bool
	)

	for _, key := range slices.Sorted(maps.Keys(hf.Tables)) {
		_hft, ok := hf.Tables[key]
		if !ok || _hft == nil {
			continue
		}
		if _hft.TId == 0 {
			found = true
			hft = NewFakeHFT(hf.DHId, hf.Rule.PlayerNum, _hft.NTId) //桌子ntid为乱序，防止覆盖其他桌子，ntid在请求楼层信息会重新排序
			break
		}
	}

	if !found {
		if len(hf.Tables) >= GetServer().ConHouse.TableNum {
			hf.DataLock.Unlock()
			return
		}
		hft = NewFakeHFT(hf.DHId, hf.Rule.PlayerNum, 999) //桌子ntid为乱序，防止覆盖其他桌子，ntid在请求楼层信息会重新排序
	}

	if hft == nil {
		hf.DataLock.Unlock()
		return
	}

	// 创建内存牌桌
	table := new(static.Table)
	table.Id = GetTableMgr().GetRandomTableId()
	if table.Id <= 0 {
		hf.DataLock.Unlock()
		xlog.Logger().Errorln("GetRandomTableId error")
		return
	} else {
		table.NTId = hft.NTId
		table.HId = hf.HId
		table.DHId = hf.DHId
		table.FId = hf.Id
		table.NFId = house.GetFloorIndexByFidWithoutLock(hf.Id)
		table.IsCost = false
		table.Creator = house.DBClub.UId
		table.CreateType = consts.CreateTypeHouse
		table.KindId = hf.Rule.KindId
		table.KindVersion = hf.Rule.Version
		table.Users = make([]*static.TableUser, hf.Rule.PlayerNum)
		table.Begin = true
		//table.GameNum = fmt.Sprintf("%s_%d_%d", nowTime.Format("20060102150405"), gameserver.Id, table.Id)
		table.IsFloorHideImg = hf.IsHide
		table.IsMemUidHide = house.DBClub.IsMemUidHide
		//if houseSwitch != nil {
		//	table.IsForbidWX = houseSwitch["BanWeChat"] == 0
		//}
		var msgct static.Msg_CreateTable
		msgct.KindId = hf.Rule.KindId
		msgct.PlayerNum = hf.Rule.PlayerNum
		msgct.RoundNum = hf.Rule.RoundNum
		msgct.CostType = hf.Rule.CostType
		msgct.Restrict = hf.Rule.Restrict
		msgct.GameConfig = hf.Rule.GameConfig
		msgct.GVoice = hf.Rule.GVoice
		msgct.Gps = true
		msgct.Voice = true
		msgct.GVoiceOk = true
		msgct.FewerStart = ""
		newConfig, cusErr := validateCreateTableParam(&msgct, true)
		if cusErr != nil {
			hf.DataLock.Unlock()
			xlog.Logger().Error(cusErr.Msg)
			return
		}
		table.Config = &static.TableConfig{
			MaxPlayerNum: newConfig.MaxPlayerNum,
			MinPlayerNum: newConfig.MinPlayerNum,
			RoundNum:     newConfig.RoundNum,
			CardCost:     newConfig.CardCost,
			CostType:     newConfig.CostType,
			View:         false,
			Restrict:     newConfig.Restrict,
			GameConfig:   newConfig.GameConfig,
			FewerStart:   newConfig.FewerStart,
			GVoice:       newConfig.GVoice,
			GameType:     2, // 默认好友房
			Difen:        newConfig.Difen,
		}
		table.IsVitamin = house.DBClub.IsVitamin && hf.IsVitamin
		table.IsHidHide = house.DBClub.IsHidHide
		table.IsAiSuper = house.DBClub.MixActive &&
			// 楼层是否为混排楼层
			hf.IsMix &&
			// 防作弊模式
			house.DBClub.TableJoinType == consts.NoCheat &&
			// 超级防作弊模式
			house.DBClub.AiSuper && hf.AiSuperNum > 0

		table.CreateStamp = hft.CreateStamp
		// 得到匹配信息
		//if table.IsAiSuper {
		//	table.CurrentMappingNum, _, table.TotalMappingNum = hf.MappingInfo(false)
		//}
		//table.CurrentMappingNum = floor.CurrentMappingInfo()
		//table.TotalMappingNum = floor.NumTotalMapping()
		if table.CurrentMappingNum >= table.TotalMappingNum {
			table.CurrentMappingNum = table.TotalMappingNum - 1
		}

		if house.DBClub.MixActive && hf.IsMix {
			table.JoinType = house.DBClub.TableJoinType
		}
		ntable := GetTableMgr().CreateTable(table)
		if ntable == nil {
			hf.DataLock.Unlock()
			xlog.Logger().Errorln("protocol_housefloor 408 table创建失败", table.Id)
			return
		}
	}

	hft.TId = table.Id
	// xlog.Logger().Warnf("hid %d created hft: %d", hf.HId, table.Id)
	hf.Tables[hft.NTId] = hft

	hf.DataLock.Unlock()
	var mixFloor []*HouseFloor
	fids := []int64{}
	if house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				mixFloor = append(mixFloor, floor)
				fids = append(fids, floor.Id)
			}
		}
		if sortTableByCreateTime(mixFloor) {
			// xlog.Logger().Warnf("sortTableByCreateTime after CheckRunningAndAdd")
			house.SyncTablesWithSorted()
		}
	}
	return
}

// 提前检查符合条件的座子
func (hf *HouseFloor) CheckEligibilityAndAdd(house *Club, uid int64, gpsInfo *static.GpsInfo) {
	var (
		min, eligibility, empty int
		superAi                 = house.IsAiSuper() && hf.IsMix
	)

	if !superAi {
		return
	}

	min = hf.NumAiSuperNeedEmptyTbl()

	hf.DataLock.Lock()
	// syslog.Logger().Info("当前楼层桌子数量:", len(hf.Tables))
	if len(hf.Tables) >= GetServer().ConHouse.TableNum {
		hf.DataLock.Unlock()
		return
	}
	for _, hft := range hf.Tables {
		if !hft.Begin && hft.UserCount() < hf.Rule.PlayerNum {
			empty++

			tableUsers := make([]int64, 0)
			for _, uid := range hft.UserWithOnline {
				if uid.Uid > 0 {
					tableUsers = append(tableUsers, uid.Uid)
				}
			}
			ok, err := hf.checkTableIn(house, uid, gpsInfo, tableUsers)

			if ok {
				eligibility++
			} else {
				if err != nil {
					hf.TableLogger().Error("check table in error:", err)
				}
			}
		}
	}

	// syslog.Logger().Infof("当前有%d张空桌子,有%d符合条件的桌子。", empty, eligibility)

	hf.TableLogger().Infof("current empty table count:%d, eligibility table count:%d.", empty, eligibility)

	if empty > min && eligibility > min {
		hf.DataLock.Unlock()
		return
	}

	hf.TableLogger().Info("add a empty table.")
	hf.Tables[999] = NewHFT(hf.Rule.PlayerNum, 999)
	hf.DataLock.Unlock()

	var mixFloor []*HouseFloor
	fids := []int64{}
	if house.DBClub.MixActive {
		for _, floor := range house.Floors {
			if floor.IsMix {
				mixFloor = append(mixFloor, floor)
				fids = append(fids, floor.Id)
			}
		}
		if sortTableByCreateTime(mixFloor) {
			xlog.Logger().Warnf("sortTableByCreateTime after CheckEligibilityAndAdd")
			house.SyncTablesWithSorted()
		}
	}
	return
}

// HftTicker 当桌子发生死锁时候创建新桌子替换之前的桌子
func HftTicker(hf *HouseFloor, hft *HouseFloorTable) {
	if hft.LockClose == nil {
		hft.LockClose = make(chan struct{}, 0)
	}
	go func(hf *HouseFloor) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		count := 1
	loop:
		for {
			select {
			case <-ticker.C:
				if count > 3 {
					xlog.Logger().Errorf("start reset hft")
					newHft := &HouseFloorTable{}
					newHft.TId = hft.TId
					newHft.UserWithOnline = hft.UserWithOnline
					newHft.CreateStamp = hft.CreateStamp
					newHft.NTId = hft.NTId
					newHft.DataLock = new(lock2.RWMutex)
					newHft.Deleted = true
					newHft.Begin = hft.Begin
					newHft.Step = hft.Step
					if !hf.IsMix {
						hf.DataLock.Lock()
						hf.Tables[hft.NTId] = newHft
						hf.DataLock.Unlock()
					} else { //混排前面hf已经上锁
						hf.Tables[hft.NTId] = newHft
					}

					break loop
				}
				xlog.Logger().Errorf("hft lock time out:tid:%d,hfid:%d,count:%d", hft.TId, hf.Id, count)
				count++
			case <-hft.LockClose:
				// 如果解锁了，则座子的原子操作字段应该还原
				hft.IsOccupy = 0
				break loop
			}
		}
	}(hf)
	return
}

func (hf *HouseFloor) GetInGameTableCount() int {
	if hf.TotalTableCount < 0 {
		hf.TotalTableCount = 0
	}
	return hf.TotalTableCount
}

func (hf *HouseFloor) GetMaxMatchCount() int {
	if hf.MaxMatchCount < 0 {
		hf.MaxMatchCount = 0
	}
	return hf.MaxMatchCount
}

// 快速入座
func (hf *HouseFloor) JoinQuick(house *Club, p *static.Person, gpsInfo *static.GpsInfo, untid int, restartTid int, priorityUsers ...int64 /*指定匹配优先级高的人*/) (hft *HouseFloorTable, seat int, cusErr *xerrors.XError) {
	floorMix := hf.IsMix && house.DBClub.MixActive
	if floorMix && house.IsAiSuper() && hf.AiSuperNum > 0 {
		return hf.JoinAiSuper(house, p, gpsInfo, untid, restartTid, priorityUsers...)
	}

	maxnum := hf.Rule.PlayerNum

	hf.DataLock.RLockWithLog()
	defer hf.DataLock.RUnlock()

	if floorMix {
		for num := 1; num <= maxnum; num++ {
			for _, hft := range hf.Tables {
				if hft.Begin {
					continue
				}
				if len(hft.UserWithOnline) > maxnum {
					maxnum = len(hft.UserWithOnline)
				}
				seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, true, maxnum-num, false)
				if seat != InvalidHouseTableSeat {
					return hft, seat, cusErr
				}
			}
		}
	} else {
		tableNum := len(hf.Tables)
		for num := 1; num <= maxnum; num++ {
			for i := 0; i < tableNum; i++ {
				//hf.DataLock.RLockWithLog()
				hft = hf.Tables[i]
				if hft.Begin {
					continue
				}
				//hf.DataLock.RUnlock()
				if len(hft.UserWithOnline) > maxnum {
					maxnum = len(hft.UserWithOnline)
				}
				seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, true, maxnum-num, false)
				if seat != InvalidHouseTableSeat {
					return hft, seat, cusErr
				}
			}
		}
	}
	return nil, InvalidHouseTableSeat, xerrors.HouseNoConditionsError
}

// 再来一局
func (hf *HouseFloor) JoinAntherGame(house *Club, p *static.Person, gpsInfo *static.GpsInfo, untid int, restartTid int, priorityUsers ...int64 /*指定匹配优先级高的人*/) (hft *HouseFloorTable, seat int, cusErr *xerrors.XError) {
	// 再来一局逻辑处理
	// 1.优先找已经有玩家点了再来一局的桌子
	if hf.IsMix && house.DBClub.MixActive {
		if house.IsAiSuper() && hf.AiSuperNum > 0 {
			return hf.JoinAiSuper(house, p, gpsInfo, untid, restartTid, priorityUsers...)
		}
	}

	hf.TableLogger().Infof("[Join Anther] uid:%d max:%d", p.Uid, hf.Rule.PlayerNum)

	houseTables := make([]*HouseFloorTable, 0, len(hf.Tables))
	hf.DataLock.RLockWithLog()
	for _, hft := range hf.Tables {
		if hft == nil {
			continue
		}
		houseTables = append(houseTables, hft)
	}
	hf.DataLock.RUnlock()
	sort.Sort(&HFTableList{
		hfts: houseTables,
		less: func(i, j *HouseFloorTable) bool {
			return i.NTId < j.NTId
		},
	})

	tableCount := len(houseTables)
	for i := 0; i < tableCount; i++ {
		hft = houseTables[i]
		if hft.RestartIn(restartTid, p.Uid, priorityUsers...) {
			seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, false, 0, true)
			if seat != -1 {
				// syslog.Logger().Info("[再来一局]找到了符合条件的桌子：", hft.TId, hft.NTId)
				return hft, seat, cusErr
			}
		}
	}

	// 2.未找到符合再来一局的桌子(前面没有玩家点再来一局或有人点了再来一局但是桌子被其他玩家桌满了), 则优先找没有玩家的空桌子
	//syslog.Logger().Info("[再来一局]未匹配到符合条件的桌子，去找空桌子")
	for i := 0; i < tableCount; i++ {
		hft = houseTables[i]
		if table := hft.GetTableInstance(); table != nil && hft.UserCount() == 0 {
			seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, false, 0, true)
			if seat != -1 {
				// syslog.Logger().Info("[再来一局]找到了空桌子：", hft.TId, hft.NTId)
				return hft, seat, cusErr
			}
		}
	}

	// 3.连空桌子都没有找到。则去找一个没有桌子实例的包厢桌子创建新桌子
	//syslog.Logger().Info("[再来一局]未匹配到空桌子，去创建新桌子")
	for i := 0; i < tableCount; i++ {
		hft = houseTables[i]
		if table := hft.GetTableInstance(); table == nil {
			seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, false, 0, true)
			if seat != -1 {
				// syslog.Logger().Debug("[再来一局]找到了可以创建新桌子的包厢牌桌", hft.TId, hft.NTId)
				return hft, seat, cusErr
			}
		}
	}
	// 4.创建不了新桌子了，其实也就是楼层桌子满了，则提示。
	// 如果包厢所有桌子都有房间存在，那么点再来一局，就提醒：包厢房间过多，暂时不能创建新房间。
	// syslog.Logger().Info("[再来一局]包厢房间过多，暂时不能创建新房间")
	return nil, -1, xerrors.HouseFloorTableFullError
}

// 智能加桌入口
func (hf *HouseFloor) JoinMixAutoEntrance(house *Club, p *static.Person, gpsInfo *static.GpsInfo, untid int, restartTid int, priorityUsers ...int64 /*指定匹配优先级高的人*/) (hft *HouseFloorTable, seat int, cusErr *xerrors.XError) {

	maxnum := hf.Rule.PlayerNum
	hf.TableLogger().Infof("[Join Entrance] uid:%d max:%d", p.Uid, maxnum)

	if hf.IsMix && house.DBClub.MixActive && house.DBClub.TableJoinType == consts.AutoAdd {
		hf.TableLogger().Infof("[Join Entrance] do")
		hf.DataLock.RLockWithLog()

		if house.DBClub.CreateTableType != 1 { // 1-另开新卓  不为另开新桌模式 就往下执行
			var emptyTblCount int
			for _, hft := range hf.Tables {
				// hft.DataLock.RLockWithLog()
				if !hft.Begin {
					if uc := hft.UserCount(); uc > 0 && uc < hf.Rule.PlayerNum {
						emptyTblCount++
						if emptyTblCount >= house.DBClub.EmptyTableMax {
							hf.DataLock.RUnlock()
							hf.TableLogger().Infof("[Join Entrance] emptyTblCount%d >= house.DBClub.EmptyTableMax%d", emptyTblCount, house.DBClub.EmptyTableMax)
							return hf.JoinQuick(house, p, gpsInfo, consts.HOUSETABLEINQUICK, restartTid, priorityUsers...)
						}
					}
				}
				// hft.DataLock.RUnlock()
			}
		}

		for _, hft := range hf.Tables {
			userCount := len(hft.UserWithOnline)
			if userCount > maxnum {
				maxnum = len(hft.UserWithOnline)
			}
			seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, true, 0, true)
			if seat != -1 {
				hf.TableLogger().Infof("[Join Entrance] CheckTableIn seat%d, err=%v", seat, cusErr)
				hf.DataLock.RUnlock()
				return hft, seat, cusErr
			}
		}
		hf.DataLock.RUnlock()
	}
	hf.TableLogger().Infof("[Join Entrance] error, tabel num=%d", len(hf.Tables))
	return nil, -1, xerrors.HouseNoConditionsError
}

// 普通 | 正常 入座
func (hf *HouseFloor) JoinNormal(house *Club, p *static.Person, gpsInfo *static.GpsInfo, untid int, restartTid int, priorityUsers ...int64 /*指定匹配优先级高的人*/) (hft *HouseFloorTable, seat int, cusErr *xerrors.XError) {
	hf.TableLogger().Infof("[Join Normal]uid:%d max:%d", p.Uid, hf.Rule.PlayerNum)

	// 正常入桌
	hf.DataLock.RLockWithLog()
	hft = hf.Tables[untid]
	hf.DataLock.RUnlock()
	if hft == nil {
		return
	}
	if !atomic.CompareAndSwapInt64(&hft.IsOccupy, 0, 1) {
		return nil, -1, xerrors.HouseFloorTableJoiningError
	}
	hft.DataLock.Lock()
	HftTicker(hf, hft)
	for seat, u := range hft.UserWithOnline {
		if u.Uid == 0 {
			key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_TABLE_LOCK, hf.Id, hft.NTId, seat)
			cli := GetDBMgr().Redis
			if cli.Exists(key).Val() == 1 {
				if cli.Get(key).Val() != fmt.Sprintf("%d", p.Uid) {
					continue
				}
			}
			hft.UserWithOnline[seat] = FTUsers{Uid: p.Uid, OnLine: true}
			return hft, seat, nil
		}
	}

	//检查卡桌数据
	table := GetTableMgr().GetTable(hft.TId)
	if table == nil {
		xlog.Logger().Errorf("get_table_id :%d nil in floot id:%d  ,ntid:%d", hft.TId, hft.NTId, hf.Id)
		hft.Clear(hf.DHId)
		if len(hft.UserWithOnline) != hf.Rule.PlayerNum {
			hft.UserWithOnline = make([]FTUsers, hf.Rule.PlayerNum)
		}
		// go hf.BroadCastMix(constant.ROLE_MEMBER, constant.MsgTypeHouseTableDissovle_Ntf, msg)
	}
	hft.DataLock.Unlock()
	hft.LockClose <- struct{}{}
	hft.IsOccupy = 0
	return nil, -1, xerrors.HouseTableFullError
}

// 超级防作弊入座
// 快速入座
func (hf *HouseFloor) JoinAiSuper(house *Club, p *static.Person, gpsInfo *static.GpsInfo, untid int, restartTid int, priorityUsers ...int64 /*指定匹配优先级高的人*/) (hft *HouseFloorTable, seat int, cusErr *xerrors.XError) {
	// 得到当前匹配人数，最大匹配中座子的人数，总共的匹配人数
	currentMapping, maxMapping, totalMapping := hf.MappingInfo(true)
	// 得到包含当前玩家的剩余匹配人数
	surplusMapping := totalMapping - currentMapping
	if surplusMapping < 0 {
		surplusMapping = 0
	}
	hf.TableLogger().Infof("[Join AiSuper]uid:%d max:%d surplus:%d total:%d maxnum:%d", p.Uid, maxMapping, surplusMapping, totalMapping, hf.Rule.PlayerNum)

	hf.DataLock.RLock()
	defer hf.DataLock.RUnlock()
	if surplusMapping <= 1 && maxMapping+surplusMapping <= hf.Rule.PlayerNum {
		hf.TableLogger().Info("ai super join by quick join.")
		// 如果剩余人数不足以满足开桌条件 则优先分配给人数最多的（也就是之前的快速入座逻辑）
		maxNum := hf.Rule.PlayerNum
		for num := 1; num <= maxNum; num++ {
			for _, hft := range hf.Tables {
				if hft.Begin {
					continue
				}
				if len(hft.UserWithOnline) > maxNum {
					maxNum = len(hft.UserWithOnline)
				}
				seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, true, maxNum-num, false)
				if seat != -1 {
					return hft, seat, cusErr
				}
			}
		}
	} else {
		// 如果剩余人数足以满足开桌条件 则随机的分配给一个未开局且人数低于当前人数最大的桌子
		hf.TableLogger().Info("ai super join normal.")
		maxNum := hf.Rule.PlayerNum - 2
		if maxNum < 0 {
			maxNum = 0
		}
		for _, hft := range hf.Tables {
			if !hft.Begin && hft.UserCount() <= maxNum {
				seat, cusErr = hf.CheckTableIn(house, hft, p.Uid, gpsInfo, false, 0, true)
				if seat != -1 {
					return hft, seat, cusErr
				}
			}
		}
	}
	return nil, -1, xerrors.HouseNoConditionsError
}

// 得到超级防作弊最大空座子数
func (hf *HouseFloor) NumAiSuperNeedEmptyTbl() int {
	if !hf.IsMix || hf.AiSuperNum < consts.HOUSEAISUPERMIN || hf.AiSuperNum > consts.HOUSEAISUPERMAX {
		return 1
	}

	return hf.AiSuperNum/hf.Rule.PlayerNum + 1
}

// 得到当前匹配中的信息
func (hf *HouseFloor) MappingInfo(detail bool) (c, max, sum int) {
	// 如果没有开启超级防作弊则不返回匹配信息
	sum = hf.AiSuperNum
	if sum <= 0 {
		return
	}
	// lock
	hf.DataLock.RLock()
	defer hf.DataLock.RUnlock()

	var buf bytes.Buffer
	for _, hft := range hf.Tables {
		if !hft.Begin {
			if detail {
				buf.WriteString(fmt.Sprintf("unbegin table eve: ntid:%d, tableid:%d, users:%v[count:%d]. \n", hft.NTId, hft.TId, hft.UserWithOnline, hft.UserCount()))
			}
			if uc := hft.UserCount(); uc < hf.Rule.PlayerNum {
				if detail {
					xlog.Logger().Infof("house %d floor %s table %d current mapping: %d", hf.HId, hf.Name, hft.TId, uc)
				}
				if uc > max {
					max = uc
				}
				c += uc
			} else {
				if detail {
					xlog.Logger().Infof("house %d floor %s table %d current user already full[%d/%d]", hf.HId, hf.Name, hft.TId, uc, hf.Rule.PlayerNum)
				}
			}
		}
	}

	if detail {
		xlog.Logger().Infof("house %d floor %s current mapping: %d, max: %d.", hf.HId, hf.Name, c, max)
		xlog.Logger().Infof("house %d floor %s detail mapping eve: \n %s", hf.HId, hf.Name, buf.String())
	}
	return
}

func (hf *HouseFloor) PubRedisMappingNumUpdate() {
	curMapping, _, totalMapping := hf.MappingInfo(false)

	// 防止为0操作
	if curMapping <= 0 {
		curMapping = 1
	}

	// 为防止进度达到100%
	if curMapping >= totalMapping {
		curMapping = totalMapping - 1
	}

	// 如果两数字完全相等 则不用推送
	if curMapping == hf.LastMappingNum && totalMapping == hf.LastMappingTotal {
		return
	}

	xlog.Logger().Infof("house %d floor %s ai super mapping num change, cur[before/now]:%d/%d, total[before/now]:%d/%d",
		hf.HId, hf.Name, hf.LastMappingNum, curMapping, hf.LastMappingTotal, totalMapping)

	hf.LastMappingNum = curMapping
	hf.LastMappingTotal = totalMapping

	go hf.RedisPub(
		consts.MsgHouseFloorMappingNumUpdate,
		static.MsgHouseFloorMappingUpdate{
			CurrentMappingNum: hf.LastMappingNum,
			TotalMappingNum:   hf.LastMappingTotal,
		},
	)
}

func (hf *HouseFloor) GetWanFaName() string {
	if hf.Name != "" {
		return hf.Name
	} else {
		area := GetAreaGameByKid(hf.Rule.KindId)
		if area != nil {
			return area.Name
		}
	}
	return ""
}

func (hf *HouseFloor) ConvertModel() (*models.HouseFloor, error) {
	mod := new(models.HouseFloor)
	mod.Id = hf.Id
	mod.DHId = hf.DHId
	data, err := json.Marshal(hf.Rule)
	if err != nil {
		return mod, err
	}
	mod.Rule = string(data[:])
	mod.IsMix = hf.IsMix
	mod.IsVip = hf.IsVip
	mod.Name = hf.Name
	mod.IsVitamin = hf.IsVitamin
	mod.IsGamePause = hf.IsGamePause
	mod.VitaminLowLimit = hf.VitaminLowLimit
	mod.VitaminHighLimit = hf.VitaminHighLimit
	mod.VitaminLowLimitPause = hf.VitaminLowLimitPause
	mod.AiSuperNum = hf.AiSuperNum
	mod.IsHide = hf.IsHide
	mod.IsCapSetVip = hf.IsCapSetVip
	mod.IsDefJoinVip = hf.IsDefJoinVip
	mod.MinTable = hf.MinTable
	mod.MaxTable = hf.MaxTable
	return mod, nil
}

// 包厢vip楼层模块
func (hf *HouseFloor) IsVipFloor() bool {
	return hf.IsVip
}

func (hf *HouseFloor) RedisKeyVipUsers() string {
	return RedisKeyFloorVipUsers(hf.Id)
}

func RedisKeyFloorVipUsers(fid int64) string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_VIP, fid)
}

func (hf *HouseFloor) AddVipUsers(ids ...int64) error {
	addCount := len(ids)
	if addCount > 0 {
		cli := GetDBMgr().GetDBrControl()
		objs := make([]interface{}, addCount)
		for i := 0; i < addCount; i++ {
			objs[i] = ids[i]
		}
		return cli.RedisV2.SAdd(hf.RedisKeyVipUsers(), objs...).Err()
	}
	return nil
}

func (hf *HouseFloor) RemVipUsers(ids ...int64) error {
	addCount := len(ids)
	if addCount > 0 {
		cli := GetDBMgr().GetDBrControl()
		objs := make([]interface{}, addCount)
		for i := 0; i < addCount; i++ {
			objs[i] = ids[i]
		}
		return cli.RedisV2.SRem(hf.RedisKeyVipUsers(), objs...).Err()
	}
	return nil
}

func RemVipUsers(fid int64, ids ...int64) error {
	addCount := len(ids)
	if addCount > 0 {
		cli := GetDBMgr().GetDBrControl()
		objs := make([]interface{}, addCount)
		for i := 0; i < addCount; i++ {
			objs[i] = ids[i]
		}
		return cli.RedisV2.SRem(RedisKeyFloorVipUsers(fid), objs...).Err()
	}
	return nil
}

func (hf *HouseFloor) NumVipUsers() int64 {
	cli := GetDBMgr().GetDBrControl()
	return cli.RedisV2.SCard(hf.RedisKeyVipUsers()).Val()
}

func (hf *HouseFloor) IsVipUser(id int64) bool {
	cli := GetDBMgr().GetDBrControl()
	return cli.RedisV2.SIsMember(hf.RedisKeyVipUsers(), id).Val()
}

func (hf *HouseFloor) GetVipUsersSet() map[int64]struct{} {
	res := make(map[int64]struct{})
	cli := GetDBMgr().GetDBrControl()
	data, err := cli.RedisV2.SMembers(hf.RedisKeyVipUsers()).Result()
	if err != nil {
		xlog.Logger().Error(err)
		return res
	}
	for i := 0; i < len(data); i++ {
		res[static.HF_Atoi64(data[i])] = struct{}{}
	}
	return res
}

func (hf *HouseFloor) GetVipUsersNumByCap(capId int64, memMap map[int64]HouseMember) int64 {
	var num int64 = 0
	cli := GetDBMgr().GetDBrControl()
	data, err := cli.RedisV2.SMembers(hf.RedisKeyVipUsers()).Result()
	if err != nil {
		xlog.Logger().Error(err)
		return 0
	}
	for _, mem := range memMap {
		for i := 0; i < len(data); i++ {
			userId := static.HF_Atoi64(data[i])
			if mem.UId == userId && (mem.Partner == capId || mem.UId == capId) {
				num++
			}
		}
	}
	return num
}

//// 此楼层的游戏 玩法是否 是匿名游戏
//func (hf *HouseFloor) IsAnonymous(rule string) bool {
//	var gameConfig map[string]interface{}
//	gameConfig = make(map[string]interface{})
//	err := json.Unmarshal([]byte(rule), &gameConfig)
//	if err != nil {
//		syslog.Logger().Errorf("查找匿名游戏时游戏玩法解析失败")
//		return false
//	}
//	key := "anonymity"
//	b, ok := gameConfig[key]
//	if !ok {
//		return false
//	}
//	return b.(string) == "true"
//}
