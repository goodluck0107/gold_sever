package center

import (
	"encoding/json"
	"errors"
	"fmt"
	goRedis "github.com/go-redis/redis"
	"github.com/open-source/game/chess.git/pkg/static/object"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/jinzhu/gorm"
)

// //////////////////////////////////////////////////////////////////////////////
// ! 包厢的数据结构
type Club struct {
	DBClub *models.House
	// 包厢队列
	OptLock *lock2.RWMutex `json:"-"` // 包厢功能
	// 楼层列表
	Floors    map[int64]*HouseFloor `json:"-"` // 包厢楼层
	FloorLock *lock2.RWMutex        `json:"-"`
	// 禁止同桌相关
	TableLimitGropuCount int                       `json:"-"` // 包厢禁止同桌组数
	AddTableLimitLock    *lock2.RWMutex            `json:"-"` // 禁止同桌组锁
	LimitGroups          map[int]map[int64]bool    `json:"-"` // 禁止同桌用户列表
	LimitGroupsUpdateAt  map[int]int64             `json:"-"` // 禁止同桌最后操作时间
	TableLimitUserLock   *lock2.RWMutex            `json:"-"` // 添加禁止同桌用户锁
	LastSum              string                    `json:"-"` // 上次被统计日期
	IsAlive              bool                      `json:"-"`
	LastAutoPay          string                    `json:"-"` // 上次队长自动划账日期
	LiveData             ClubLiveData              `json:"-"` // 包厢实时数据
	UserGroup            map[int64]map[int][]int64 `json:"-"`
	AddUserGroupLock     *lock2.RWMutex            `json:"-"` // 禁止同桌组锁

	lastUpdatetime       int64                     `json:"-"` // 上次更新新时间
	historyData          []models.RecordVitaminDay `json:"-"` // 更新数据
	PrizeVal             []int64                   `json:"prizeval"`
	GroupPrizeVal        []int64                   `json:"groupprizeval"`
	ClubMemberSwitch     map[string]int            `json:"-"` // 包厢功能开关
	ClubMemberSwitchLock *lock2.RWMutex            `json:"-"`
	LastSyncSorted       int64                     `json:"last_sync_sorted"`
}

// 包厢实时数据
type ClubLiveData struct {
	TotalMember  object.SetInt64 `json:"-"` // 总成员数
	OnLineMember object.SetInt64 `json:"-"` // 在线人数
}

// 包厢成员总数新增
func (clb *Club) TotalMemIncr(uid int64) {
	clb.LiveData.TotalMember.Add(uid)
}

// 包厢成员总数减少
func (clb *Club) TotalMemDecr(uid int64) {
	clb.LiveData.TotalMember.Remove(uid)
}

// 包厢成员总数
func (clb *Club) TotalMemCount() int {
	return clb.LiveData.TotalMember.Count()
}

// 包厢成员总数新增
func (clb *Club) OnlineMemIncr(uid int64) {
	clb.LiveData.OnLineMember.Add(uid)
}

// 包厢在线成员总数减少
func (clb *Club) OnlineMemDecr(uid int64) {
	clb.LiveData.OnLineMember.Remove(uid)
}

// 包厢在线成员总数
func (clb *Club) OnlineMemCount() int {
	realCount := clb.LiveData.OnLineMember.Count()
	fakeCount := clb.GetFakeOnlineCounts()
	return realCount + fakeCount
}

func (clb *Club) Init() {
	clb.OptLock = new(lock2.RWMutex)
	clb.Floors = make(map[int64]*HouseFloor)
	clb.FloorLock = new(lock2.RWMutex)
	clb.AddTableLimitLock = new(lock2.RWMutex)
	clb.AddUserGroupLock = new(lock2.RWMutex)
	clb.TableLimitUserLock = new(lock2.RWMutex)
	clb.ClubMemberSwitchLock = new(lock2.RWMutex)
	clb.IsAlive = true
}

func (clb *Club) flush() {

	oldhouse, _ := GetDBMgr().GetDBrControl().GetHouseInfoById(clb.DBClub.Id)
	if oldhouse == nil {
		return
	}

	dbhouse := clb.DBClub
	if err := GetDBMgr().db_M.Transaction(
		func(tx *gorm.DB) error {
			err := GetDBMgr().GetDBrControl().HouseInsert(dbhouse)
			if err != nil {
				return err
			}
			err = tx.Omit("created_at").Save(dbhouse).Error
			if err != nil {
				return err
			}
			return nil
		},
	); err != nil {
		xlog.Logger().Errorln("house.flush.error:", err.Error())
	}
}

func (clb *Club) ClubMemKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, clb.DBClub.Id)
}

func (clb *Club) BroadcastV2(role int, header string, data interface{}) {
	mems := clb.GetMemSimple(false)
	buf := static.HF_EncodeMsg(header, xerrors.SuccessCode, data, GetServer().Con.Encode, GetServer().Con.EncodeClientKey, 0)
	for _, mem := range mems {
		if mem.Lower(role) {
			continue
		}
		if person := GetPlayerMgr().GetPlayer(mem.UId); person != nil {
			person.SendBuf(buf)
		}
	}
}

// 广播
func (clb *Club) Broadcast(role int, header string, data interface{}) {
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.Lower(role) {
			continue
		}
		mem.SendMsg(header, data)
	}
}

// 广播给队长
func (clb *Club) ParnterBroadcast(header string, data interface{}) {
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.Partner == 1 {
			mem.SendMsg(header, data)
			continue
		}
	}
}

// 广播给指定队长、副队长
func (clb *Club) Broadcast2Parnter(partnerId int64, isSync2VicePartner bool, header string, data interface{}) {
	mems := clb.GetMemSimple(false)

	for _, mem := range mems {
		if mem.UId == partnerId {
			mem.SendMsg(header, data)
		} else if isSync2VicePartner && mem.IsVicePartner() && mem.Partner == partnerId {
			mem.SendMsg(header, data)
		}
	}
}

// 广播给队员
func (clb *Club) BroadcastTeam(partnerId int64, header string, data interface{}) {
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.UId == partnerId {
			mem.SendMsg(header, data)
		} else if mem.Partner == partnerId {
			mem.SendMsg(header, data)
		}
	}
}

// 广播
func (clb *Club) CustomBroadcast(role int, partner, vicePartner, vitaminAdmin bool, header string, data interface{}, ids ...int64) {
	mems := clb.GetMemSimple(false)
	l := len(ids)
	in := func(uid int64) bool {
		if l <= 0 {
			return false
		}
		for i := 0; i < l; i++ {
			if ids[i] == uid {
				return true
			}
		}
		return false
	}
	fakers := GetDBMgr().GetDBrControl().RedisV2.SMembersMap("faker_admin").Val()
	for _, mem := range mems {
		_, isFaker := fakers[fmt.Sprint(mem.UId)]
		if isFaker || (mem.URole <= role) || (mem.IsPartner() && partner) || (mem.IsVicePartner() && vicePartner) || (mem.IsVitaminAdmin() && vitaminAdmin) || (in(mem.UId)) {
			mem.SendMsg(header, data)
		}
	}
}

// 获得当前成员数量
func (clb *Club) GetMemCounts() int {
	return clb.GetMemCount(true, false)
}

// GetMemCount 获取当前在线或全部成员数量，只能有一个为true
func (clb *Club) GetMemCount(count bool, online bool) int {
	if !count && !online {
		return 0
	}
	cli := GetDBMgr().Redis
	mems := cli.HGetAll(clb.ClubMemKey()).Val()
	dest := 0

	for _, v := range mems {
		mem := HouseMember{}
		err := json.Unmarshal([]byte(v), &mem)
		if err != nil {
			continue
		}
		if mem.HId == 0 {
			mem.HId = clb.DBClub.HId
		}
		if mem.URole <= consts.ROLE_MEMBER {
			if count {
				dest++
			} else if online {
				if mem.IsOnline {
					dest++
				}
			}
		}
	}

	return dest
}

// 获得在线成员数量
func (clb *Club) GetMemOnlineCounts() int {
	return clb.GetMemCount(false, true)
}

// 　获得有人桌数
func (clb *Club) GetTabOnlineCounts() int {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()

	counts := 0
	for _, f := range clb.Floors {
		f.DataLock.RLock()
		for _, t := range f.Tables {
			if t.UserCount() > 0 {
				counts++
			}
		}
		f.DataLock.RUnlock()
	}

	return counts
}

func (clb *Club) GetFakeOnlineCounts() int {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()

	counts := 0
	for _, f := range clb.Floors {
		f.DataLock.RLock()
		for _, t := range f.Tables {
			if t.FakeEndAt > 0 {
				counts += f.Rule.PlayerNum
			}
		}
		f.DataLock.RUnlock()
	}

	return counts
}

// 　获得有人桌数
func (clb *Club) GetTableByUid(uid int64) *HouseFloorTable {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()
	for _, f := range clb.Floors {
		f.DataLock.RLock()
		for _, t := range f.Tables {
			for i := 0; i < len(t.UserWithOnline); i++ {
				if t.UserWithOnline[i].Uid == uid {
					f.DataLock.RUnlock()
					return t
				}
			}
		}
		f.DataLock.RUnlock()
	}
	return nil
}

// 获得所有成员数据
func (clb *Club) GetAllMem() []HouseMember {
	cli := GetDBMgr().Redis
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.HId == 0 {
			mem.HId = clb.DBClub.HId
		}
		mem.IsLimitGame = cli.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), mem.UId).Val() //TODO 一次获取数据
	}
	return mems
}

// 获得所有成员数据分类
func (clb *Club) GetAllMemberWithClassify() (map[int64]HouseMember, []HouseMember, []HouseMember, []HouseMember, []HouseMember, []HouseMember, []HouseMember, []HouseMember, []HouseMember, []HouseMember) {

	var OnLineArr []HouseMember         // 在线的普通成员
	var OffLineArr []HouseMember        // 离线的普通成员
	var OnLineArrParnter []HouseMember  // 在线的队长
	var OffLineArrParnter []HouseMember // 离线的队长
	var OnLineArrAdmin []HouseMember    // 在线的管理员
	var OffLineArrAdmin []HouseMember   // 离线的管理员
	var CreaterArr []HouseMember        // 盟主
	var ApplyArr []HouseMember          // 申请
	var BlackArr []HouseMember          // 黑名单

	houseMemberMap := make(map[int64]HouseMember) //所有成员数据

	mems := clb.GetMemSimple(true)
	for _, mem := range mems {
		if mem.IsOnline {
			if mem.IsPartner() {
				OnLineArrParnter = append(OnLineArrParnter, mem)
			} else if mem.URole == consts.ROLE_ADMIN {
				OnLineArrAdmin = append(OnLineArrAdmin, mem)
			} else if mem.URole == consts.ROLE_MEMBER {
				OnLineArr = append(OnLineArr, mem)
			}
		} else {
			if mem.IsPartner() {
				OffLineArrParnter = append(OffLineArrParnter, mem)
			} else if mem.URole == consts.ROLE_ADMIN {
				OffLineArrAdmin = append(OffLineArrAdmin, mem)
			} else if mem.URole == consts.ROLE_MEMBER {
				OffLineArr = append(OffLineArr, mem)
			}
		}
		if mem.URole == consts.ROLE_CREATER {
			CreaterArr = append(CreaterArr, mem)
		} else if mem.URole == consts.ROLE_APLLY {
			ApplyArr = append(ApplyArr, mem)
		} else if mem.URole == consts.ROLE_BLACK {
			BlackArr = append(BlackArr, mem)
		}

		houseMemberMap[mem.UId] = mem
	}

	sort.Sort(HouseMemberItemWrapper{OnLineArrAdmin, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	sort.Sort(HouseMemberItemWrapper{OffLineArrAdmin, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	sort.Sort(HouseMemberItemWrapper{OnLineArrParnter, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	sort.Sort(HouseMemberItemWrapper{OffLineArrParnter, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	sort.Sort(HouseMemberItemWrapper{OnLineArr, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	sort.Sort(HouseMemberItemWrapper{OffLineArr, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	return houseMemberMap, CreaterArr, OnLineArr, OffLineArr, OnLineArrParnter, OffLineArrParnter, OnLineArrAdmin, OffLineArrAdmin, ApplyArr, BlackArr
}

// GetMemSimple 获取成员基础信息，all表示所有成员，包含黑名单和申请,不包含是否被禁止游戏的数据
func (clb *Club) GetMemSimple(all bool) []HouseMember {
	cli := GetDBMgr().Redis
	mems := cli.HGetAll(clb.ClubMemKey()).Val()
	dest := make([]HouseMember, 0, len(mems))
	for _, v := range mems {
		mem := HouseMember{}
		err := json.Unmarshal([]byte(v), &mem)
		if err != nil {
			continue
		}
		if mem.HId == 0 {
			mem.HId = clb.DBClub.HId
		}
		if all {
			dest = append(dest, mem)
		} else {
			if mem.URole <= consts.ROLE_MEMBER {
				dest = append(dest, mem)
			}
		}
	}
	return dest
}

// GetMemSimpleToMap 将成员列表以map形式返回，适用于需要多个成员数据时的批量返回
func (clb *Club) GetMemSimpleToMap(all bool) map[int64]*HouseMember {
	cli := GetDBMgr().Redis
	mems := cli.HGetAll(clb.ClubMemKey()).Val()
	dest := make(map[int64]*HouseMember, len(mems))
	for _, v := range mems {
		mem := HouseMember{}
		err := json.Unmarshal([]byte(v), &mem)
		if err != nil {
			continue
		}
		if mem.HId == 0 {
			mem.HId = clb.DBClub.HId
		}
		if all {
			dest[mem.UId] = &mem
		} else {
			if mem.URole <= consts.ROLE_MEMBER {
				dest[mem.UId] = &mem
			}
		}
	}
	return dest

}

// 获得成员数据
func (clb *Club) GetMemById(Id int64) *HouseMember {
	if Id <= 0 {
		return nil
	}
	cli := GetDBMgr().Redis
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.Id == Id {
			mem.IsLimitGame = cli.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), mem.UId).Val()
			if mem.HId == 0 {
				mem.HId = clb.DBClub.HId
			}
			return &mem
		}
	}
	return nil
}
func (clb *Club) GetMemByUId(UId int64) *HouseMember {
	if UId <= 0 {
		return nil
	}
	cli := GetDBMgr().Redis
	v := cli.HGet(clb.ClubMemKey(), fmt.Sprintf("%d", UId)).Val()
	if len(v) == 0 {
		return nil
	}
	mem := HouseMember{}
	err := json.Unmarshal([]byte(v), &mem)
	if err != nil {
		return nil
	}
	if mem.HId == 0 {
		mem.HId = clb.DBClub.HId
	}
	mem.IsLimitGame = cli.SIsMember(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), mem.UId).Val()
	return &mem
}

func (clb *Club) IsPartner(UId int64) bool {
	mem := clb.GetMemByUId(UId)
	if mem != nil {
		return mem.IsPartner()
	}
	return false
}

func (clb *Club) GetIDsByPartner(memMap map[int64]HouseMember, uid int64) []int64 {
	if memMap == nil {
		memMap = clb.GetMemberMap(false)
	}
	items := make([]int64, 0)
	for _, mem := range memMap {
		if mem.Partner == uid {
			items = append(items, mem.Id)
		}
	}
	sort.Sort(util.Int64Slice(items))
	return items
}

func (clb *Club) GetIDsBySuperior(sid int64) []int64 {
	items := make([]int64, 0)
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.Partner == 1 && mem.Superior == sid {
			items = append(items, mem.Id)
		}
	}
	sort.Sort(util.Int64Slice(items))
	return items
}

func (clb *Club) GetMemByPartner(memMap map[int64]HouseMember, uid int64) []HouseMember {
	if memMap == nil {
		memMap = clb.GetMemberMap(false)
	}
	items := make([]HouseMember, 0)
	for _, mem := range memMap {
		if mem.Partner == uid {
			items = append(items, mem)
		}
	}
	return items
}

func (clb *Club) GetAllMemByPartner(memMap map[int64]HouseMember, pid int64) ([]int64, []int64) {
	if memMap == nil {
		memMap = clb.GetMemberMap(false)
	}
	var (
		js = []int64{}
		ms = []int64{}
	)

	if pid == clb.DBClub.UId {
		for _, mem := range memMap {
			if mem.UId != pid && mem.Partner == 0 {
				ms = append(ms, mem.UId)
			}
		}
	} else {
		// 找到所有的队长列表
		var next = []int64{pid}
		for {
			var juniors []int64
			for _, mem := range memMap {
				if mem.IsPartner() && static.In64(next, mem.Superior) {
					juniors = append(juniors, mem.UId)
				}
			}
			if len(juniors) <= 0 {
				break
			}
			js = append(js, juniors...)
			next = juniors
		}
		tempJs := append(js, pid)
		for _, mem := range memMap {
			if mem.Partner > 0 && static.In64(tempJs, mem.Partner) {
				ms = append(ms, mem.UId)
			}
		}
	}
	return js, ms
}

func (clb *Club) GetUIDsByPartner(memMap map[int64]HouseMember, uid int64) []int64 {
	if memMap == nil {
		memMap = clb.GetMemberMap(false)
	}
	items := make([]int64, 0)
	for _, mem := range memMap {
		if mem.Partner == uid {
			items = append(items, mem.UId)
		}
	}
	sort.Sort(util.Int64Slice(items))
	return items
}

func (clb *Club) GetJuniorAndPlayerBySuperior(uid int64) ([]int64, []int64) {
	mems := clb.GetMemSimple(false)
	players := make([]int64, 0)
	juniors := make([]int64, 0)
	for _, mem := range mems {
		if mem.Superior == uid {
			juniors = append(juniors, mem.UId)
		}
		if mem.Partner == uid {
			players = append(players, mem.UId)
		}
	}
	sort.Sort(util.Int64Slice(juniors))
	sort.Sort(util.Int64Slice(players))
	return players, juniors
}

func (clb *Club) GetAllJunior(uid int64) map[int64]int64 {
	mems := clb.GetMemberMap(false)

	isJuniorMap := make(map[int64]int64)
	notJuniorMap := make(map[int64]int64)

	isJuniorMap[uid] = uid
	for _, mem := range mems {
		if mem.IsPartner() {
			clb.CheckSuperiorForGetJunior(&mems, &isJuniorMap, &notJuniorMap, mem, uid)
		}
	}
	return isJuniorMap
}

func (clb *Club) CheckSuperiorForGetJunior(memMap *map[int64]HouseMember, isJuniorMap *map[int64]int64, notJuniorMap *map[int64]int64, mem HouseMember, assignUid int64) bool {
	// 0 表示没有上级
	if mem.Superior == 0 {
		(*notJuniorMap)[mem.UId] = mem.UId
		return false
	}

	if _, ok := (*isJuniorMap)[mem.UId]; ok {
		return true
	}
	if _, ok := (*notJuniorMap)[mem.UId]; ok {
		return false
	}

	if mem.Superior == assignUid {
		(*isJuniorMap)[mem.UId] = mem.UId
		return true
	} else {
		superMem, ok := (*memMap)[mem.Superior]
		if !ok {
			(*notJuniorMap)[mem.UId] = mem.UId
			return false
		}
		in := clb.CheckSuperiorForGetJunior(memMap, isJuniorMap, notJuniorMap, superMem, assignUid)
		if in {
			(*isJuniorMap)[mem.UId] = mem.UId
		} else {
			(*notJuniorMap)[mem.UId] = mem.UId
		}
		return in
	}
}

func (clb *Club) GetIDsByWithOutPartner(uid int64) []int64 {
	mems := clb.GetMemSimple(false)
	items := make([]int64, 0)
	for _, mem := range mems {
		// 非楼主
		if mem.URole <= 1 {
			continue
		}
		// 非当前队长
		if mem.Partner >= 1 {
			continue
		}
		// 非黑名单
		if mem.URole > consts.ROLE_MEMBER {
			continue
		}
		items = append(items, mem.Id)
	}
	return items
}

func (clb *Club) GetPartnerLinkByUid(memMap map[int64]HouseMember, uid int64) []int64 {
	if memMap == nil {
		memMap = clb.GetMemberMap(false)
	}
	items := make([]int64, 0)

	next := uid
	for next > 0 {
		mem, ok := memMap[next]
		if !ok {
			return items
		}
		next = 0
		if mem.IsPartner() {
			items = append(items, mem.UId)
			next = mem.Superior
		} else if mem.Partner > 0 {
			// items = append(items, mem.Partner)
			next = mem.Partner
		}
	}
	return items
}

func (clb *Club) GetPartnerMembersSumVitamin(memMap map[int64]HouseMember, pid int64) int64 {
	if pid <= 0 {
		return 0
	}
	if memMap == nil {
		memMap = clb.GetMemberMap(false)
	}
	ps, ms := clb.GetAllMemByPartner(memMap, pid)
	allMems := []int64{pid}
	allMems = append(allMems, append(ps, ms...)...)
	var sumVitamin int64
	for _, mem := range memMap {
		if static.In64(allMems, mem.UId) {
			sumVitamin += mem.UVitamin
		}
	}
	return sumVitamin
}

// 获取成员id
func (clb *Club) GetMemsMid() []int64 {
	mems := clb.GetMemSimple(false)
	datas := make([]int64, 0, len(mems))
	for _, m := range mems {
		datas = append(datas, m.Id)
	}
	return datas
}

// 获取成员id
func (clb *Club) GetMemsUid() []int64 {
	mems := clb.GetMemSimple(false)
	datas := make([]int64, 0, len(mems))
	for _, m := range mems {
		datas = append(datas, m.UId)
	}
	return datas
}

// 获得成员列表数据
func (clb *Club) GetMemUIdsByRole(role int) []int64 {
	if role == consts.ROLE_CREATER {
		return []int64{clb.DBClub.UId}
	}
	mems := clb.GetMemSimple(true)
	datas := make([]int64, 0, len(mems))
	for _, m := range mems {
		if m.URole == role {
			datas = append(datas, m.UId)
		}
	}
	return datas
}

func (clb *Club) GetMemIdsByRole(role int) []int64 {
	if role == consts.ROLE_CREATER {
		mem := clb.GetMemByUId(clb.DBClub.UId)
		return []int64{mem.Id}
	}
	mems := clb.GetMemSimple(false)
	datas := make([]int64, 0, len(mems))
	for _, m := range mems {
		if m.URole == role {
			datas = append(datas, m.Id)
		}
	}
	return datas

}

// 获得楼层数据
func (clb *Club) GetFloors() ([]int64, []int) {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()

	//　包厢楼层id
	IdArr := make([]int64, 0)
	for k, _ := range clb.Floors {
		IdArr = append(IdArr, k)
	}

	// 排序
	sort.Sort(util.Int64Slice(IdArr))

	var kindIdArr []int
	for _, id := range IdArr {
		kindIdArr = append(kindIdArr, clb.Floors[id].Rule.KindId)
	}

	return IdArr, kindIdArr
}

// 获得楼层数据
func (clb *Club) GetFloorByFId(fId int64) *HouseFloor {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()

	floor := clb.Floors[fId]
	if floor == nil {
		return nil
	}

	return floor
}

// 获得楼层索引
func (clb *Club) GetFloorIndexByFid(fId int64) int {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()

	var IdArr []int64
	for _, hf := range clb.Floors {
		IdArr = append(IdArr, hf.Id)
	}
	sort.Sort(util.Int64Slice(IdArr))

	nfid := -1
	for i := 0; i < len(IdArr); i++ {
		if IdArr[i] == fId {
			nfid = i
			break
		}
	}
	return nfid
}

// 获得楼层索引
func (clb *Club) GetFloorIndexByFidWithoutLock(fId int64) int {

	var IdArr []int64
	for _, hf := range clb.Floors {
		IdArr = append(IdArr, hf.Id)
	}
	sort.Sort(util.Int64Slice(IdArr))

	nfid := -1
	for i := 0; i < len(IdArr); i++ {
		if IdArr[i] == fId {
			nfid = i
			break
		}
	}
	return nfid
}

// 获得楼层索引
func (clb *Club) GetFloorIndex(dfids []int64) []int64 {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()

	//　包厢楼层id
	IdArr := make([]int64, 0)
	for k, _ := range clb.Floors {
		IdArr = append(IdArr, k)
	}

	// 排序
	sort.Sort(util.Int64Slice(IdArr))

	floorIndexs := []int64{}
	for _, dfid := range dfids {
		for i := 0; i < len(IdArr); i++ {
			if dfid == IdArr[i] {
				floorIndexs = append(floorIndexs, int64(i))
			}
		}
	}

	return floorIndexs
}

// 包厢获取活动
func (clb *Club) ExistActivity() bool {

	arr, err := GetDBMgr().GetDBrControl().HouseActivityList(clb.DBClub.Id, true)
	if err != nil {
		xlog.Logger().Errorln(err)
		return false
	}

	if len(arr) > 0 {
		return true
	}

	return false
}

// 包厢获取活动
func (clb *Club) IsActivity() bool {

	arr, err := GetDBMgr().GetDBrControl().HouseActivityList(clb.DBClub.Id, true)
	if err != nil {
		xlog.Logger().Errorln(err)
		return false
	}

	for _, ar := range arr {
		if ar == nil {
			continue
		}
		if time.Now().Unix() < ar.EndTime {
			return true
		}
	}
	return false
}

// 设置审核
func (clb *Club) OptionMemCheck(checked bool, opID int64) *xerrors.XError {
	clb.DBClub.IsChecked = checked
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	if !checked { //将申请列表用户全部通过
		for _, k := range clb.GetMemUIdsByRole(consts.ROLE_APLLY) {
			clb.UserJoin(k, opID)
		}
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsMemCheck)
	ntf.HId = clb.DBClub.HId
	ntf.IsChecked = clb.DBClub.IsChecked
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsMemCheck_Ntf, ntf)

	return nil
}

// 设置队长审核
func (clb *Club) OptionParnterMemCheck(checked bool, opID int64) *xerrors.XError {
	clb.DBClub.IsPartnerApply = checked
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsMemCheck)
	ntf.HId = clb.DBClub.HId
	ntf.IsChecked = clb.DBClub.IsPartnerApply
	clb.CustomBroadcast(consts.ROLE_CREATER, true, true, false, consts.MsgTypeHouseOptIsParnterMemCheck_Ntf, ntf)

	return nil
}

// 设置成员退圈开关
func (clb *Club) OptionIsMemberExit(checked bool, opID int64) *xerrors.XError {
	clb.DBClub.IsMemExit = checked
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}
	memMap := clb.GetMemberMap(false)
	if !checked { // 将申请列表用户全部通过
		exitApplicants := clb.GetExitApplicants()
		for i, l := 0, len(exitApplicants); i < l; i++ {
			hmem, ok := memMap[exitApplicants[i].Uid]
			if ok {
				clb.memberExitAgree(&hmem, opID, true)
			}
		}
	}
	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsMemExitCheck)
	ntf.HId = clb.DBClub.HId
	ntf.IsChecked = clb.DBClub.IsMemExit
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsMemExitCheck_Ntf, ntf)
	return nil
}

// 设置冻结
func (clb *Club) OptionFrozen(frozen bool) *xerrors.XError {
	clb.DBClub.IsFrozen = frozen
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}
	sql := `update house set is_frozen = ? where id = ?`
	GetDBMgr().GetDBmControl().Exec(sql, frozen, clb.DBClub.Id)
	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsFrozen)
	ntf.HId = clb.DBClub.HId
	ntf.IsFrozen = clb.DBClub.IsFrozen
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsFrozen_Ntf, ntf)

	return nil
}

// 包厢设置成员列表隐藏
func (clb *Club) OptionMemHide(hide bool) *xerrors.XError {
	clb.DBClub.IsMemHide = hide
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsMemHide)
	ntf.HId = clb.DBClub.HId
	ntf.IsMemHide = clb.DBClub.IsMemHide
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsMemHide_Ntf, ntf)

	return nil
}

// 包厢设置Hid隐藏
func (clb *Club) OptionHidHide(hide bool) *xerrors.XError {
	clb.DBClub.IsHidHide = hide
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsHidHide)
	ntf.HId = clb.DBClub.HId
	ntf.IsHidHide = clb.DBClub.IsHidHide
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsHidHide_Ntf, ntf)

	return nil
}

// 包厢设置头像隐藏
func (clb *Club) OptionHeadHide(hide bool) *xerrors.XError {
	clb.DBClub.IsHeadHide = hide
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsHeadHide)
	ntf.HId = clb.DBClub.HId
	ntf.IsHeadHide = clb.DBClub.IsHeadHide
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsHeadHide_Ntf, ntf)

	return nil
}

// 包厢设置Uid隐藏
func (clb *Club) OptionUidHide(hide bool) *xerrors.XError {
	clb.DBClub.IsMemUidHide = hide
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseOptIsMemUidHide)
	ntf.HId = clb.DBClub.HId
	ntf.IsMemUidHide = clb.DBClub.IsMemUidHide
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseOptIsMemUidHide_Ntf, ntf)
	return nil
}

// 包厢设置在线人数隐藏
func (clb *Club) OptionPartnerKick(pk bool) *xerrors.XError {
	clb.DBClub.PartnerKick = pk
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.NtfHCHouseOptPartnerKick)
	ntf.HId = clb.DBClub.HId
	ntf.PartnerKick = clb.DBClub.PartnerKick
	clb.CustomBroadcast(consts.ROLE_CREATER, true, true, false, consts.MsgTypeHouseOptIsPartnerKick_Ntf, ntf)
	return nil
}

// 包厢设置在线人数隐藏
func (clb *Club) OptionOnlineHide(hide bool) *xerrors.XError {
	clb.DBClub.IsOnlineHide = hide
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	onlineCur, onlineTotal, onlineTbl := clb.OnlineMemCount(), clb.TotalMemCount(), clb.GetTabOnlineCounts()

	ntf := &static.Ntf_HC_HouseOptIsOnlineHide{
		HId:          clb.DBClub.HId,
		IsOnlineHide: clb.DBClub.IsOnlineHide,
	}
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if clb.IsHideOnlineNum(mem.UId) {
			ntf.OnlineCur = -1
			ntf.OnlineTotal = -1
			ntf.OnlineTable = -1
		} else {
			ntf.OnlineCur = onlineCur
			ntf.OnlineTotal = onlineTotal
			ntf.OnlineTable = onlineTbl
		}
		mem.SendMsg(consts.MsgTypeHouseOptIsOnlineHide_Ntf, ntf)
	}
	return nil
}

// 包厢设置成员列表隐藏
func (clb *Club) OptionAutoPay(auto bool) *xerrors.XError {
	clb.DBClub.AutoPayPartnrt = auto
	clb.flush()
	// 通知
	ntf := new(static.Ntf_HC_HouseOptAutoPay)
	ntf.HId = clb.DBClub.HId
	ntf.AutoPay = clb.DBClub.AutoPayPartnrt
	clb.BroadcastMsg(consts.MsgHouseAutoPayPartnerNtf, ntf, func(member *HouseMember) bool {
		return member.URole == consts.ROLE_CREATER || member.IsVitaminAdmin()
	})
	return nil
}

// 包厢设置成员位置信息隐藏
func (clb *Club) OptionMemGPSHide(hide bool) *xerrors.XError {
	clb.DBClub.PrivateGPS = hide
	clb.flush()
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.NtfHCHouseGpsHide)
	ntf.HId = clb.DBClub.HId
	ntf.PrivateGPS = clb.DBClub.PrivateGPS
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgHousePrivateGPSSet_ntf, ntf)
	return nil
}

func (clb *Club) MemExist(uid int64) bool {
	return GetDBMgr().Redis.HExists(clb.ClubMemKey(), fmt.Sprint(uid)).Val()
}

// 包厢设置2人桌子禁止同桌不生效 是否勾选
func (clb *Club) Option2PTableLimitNotEffect(sta bool) *xerrors.XError {
	clb.DBClub.IsNotEft2PTale = sta
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}
	return nil
}

// 加入包厢
func (clb *Club) MemJoin(uId int64, role int, refId int64, check bool, tx *gorm.DB) *xerrors.XError {
	if check {
		if clb.IsBusyMerging() {
			return xerrors.HouseBusyError
		}
		// 人数上限
		if clb.GetMemCounts() >= GetServer().ConHouse.MemMax {
			return xerrors.HouseMemJoinMaxError
		}
		// 玩家加入上限
		count, err := GetDBMgr().GetDBrControl().HouseMemberJoinCounts(uId)
		if err != nil {
			return xerrors.DBExecError
		}
		if count >= GetServer().ConHouse.JoinMax {
			return xerrors.MemJoinHouseMaxError
		}
	}
	var (
		p   *static.Person
		err error
	)
	session := GetPlayerMgr().GetPlayer(uId)
	if session == nil {
		p, err = GetDBMgr().GetDBrControl().GetPerson(uId)
	} else {
		p = &session.Info
	}
	if p == nil || err != nil {
		return xerrors.UserNotExistError
	}
	// 包厢成员
	hmem := clb.GetMemByUId(uId)
	if hmem != nil {
		// 黑名单
		if hmem.URole == consts.ROLE_BLACK {
			return xerrors.NewXError("你已被盟主限制加入该包厢。")
			// 之前已提交审核，后直接加入
		} else if hmem.URole == consts.ROLE_APLLY {
			/*
				if clb.DBClub.IsChecked && clb.IsNormalNoMerge() {
					return xerrors.InReviewError
				} else {
					hmem.URole = role
				}
			*/
			// 多次入圈申请以最后一次为最新信息
			hmem.URole = role
		}
		if role == consts.ROLE_ADMIN {
			hmem.ApplyTime = time.Now().Unix()
		} else if role == consts.ROLE_MEMBER {
			hmem.ApplyTime = time.Now().Unix()
			hmem.AgreeTime = time.Now().Unix()
		}
		// 更新
		GetDBMgr().HouseMemberUpdate(clb.DBClub.Id, hmem)
		hmem.IsOnline = true
		hmem.UVitamin = 0
		hmem.NickName = p.Nickname
		hmem.ImgUrl = p.Imgurl
		hmem.Sex = p.Sex
		if hmem.Ref > 0 {
			thouse := GetClubMgr().GetClubHouseById(hmem.Ref)
			if thouse == nil {
				return xerrors.InValidHouseError
			}
			if thouse.DBClub.UId == hmem.UId {
				hmem.Partner = 1
			} else {
				hmem.Partner = thouse.DBClub.UId
			}
		}
		hmem.Flush()

	} else {
		// 插入
		hmem = new(HouseMember)
		hmem.Id = 0
		hmem.UId = uId
		hmem.URole = role
		hmem.HId = clb.DBClub.HId
		hmem.DHId = clb.DBClub.Id
		hmem.IsOnline = session != nil
		hmem.NickName = p.Nickname
		hmem.ImgUrl = p.Imgurl
		hmem.Sex = p.Sex
		hmem.Ref = refId
		if hmem.Ref > 0 {
			thouse := GetClubMgr().GetClubHouseById(hmem.Ref)
			if thouse == nil {
				return xerrors.InValidHouseError
			}
			if thouse.DBClub.UId == hmem.UId {
				hmem.Partner = 1
			} else {
				hmem.Partner = thouse.DBClub.UId
			}
		}
		_, err := GetDBMgr().HouseMemberInsert(clb.DBClub.Id, hmem, tx)
		if err != nil {
			xlog.Logger().Errorln(err)
			return xerrors.DBExecError
		}
		if role == consts.ROLE_BLACK {
			hmem.AgreeTime = time.Now().Unix()
			hmem.ApplyTime = time.Now().Unix()
		} else if role == consts.ROLE_APLLY {
			hmem.ApplyTime = time.Now().Unix()
		} else if role == consts.ROLE_MEMBER {
			hmem.ApplyTime = time.Now().Unix()
			hmem.AgreeTime = time.Now().Unix()
		} else if role == consts.ROLE_CREATER {
			hmem.ApplyTime = time.Now().Unix()
			hmem.AgreeTime = time.Now().Unix()
		}
		hmem.Flush()
		clb.TotalMemIncr(hmem.UId)
		//
	}
	if hmem != nil {
		// 向数据库同步同意时间字段
		//更新数据库
		attrMap := make(map[string]interface{})
		attrMap["apply_time"] = time.Unix(hmem.ApplyTime, 0)
		attrMap["agree_time"] = time.Unix(hmem.AgreeTime, 0)
		if tx == nil {
			GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, hmem.UId).Update(attrMap)
		} else {
			tx.Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, hmem.UId).Update(attrMap)
		}

		// 移除此圈历史队长信息
		if !hmem.Lower(consts.ROLE_MEMBER) {
			GetDBMgr().RemoveMemberLeaveHousePartner(hmem.DHId, hmem.UId)
		}
	}
	//30需求 加入的时候 遍历出 默认加入的VIP楼层，然后 floor.AddVipUsers(hmem.UId)
	fids, _ := clb.GetFloors()
	for _, v := range fids {
		floor := clb.GetFloorByFId(v)
		if floor != nil {
			if floor.IsDefJoinVip {
				floor.AddVipUsers(hmem.UId)
				//syslog.Logger().Println("floor000--- ", floor)
				//syslog.Logger().Println("floor001--- ", floor.IsDefJoinVip)
				//syslog.Logger().Println("floor002--- ", hmem.UId)
			}
		}
	}
	return nil
}

func (clb *Club) checkMemExit(hmem *HouseMember) *xerrors.XError {
	if hmem == nil {
		return xerrors.UserNotExistError
	}
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	if _, ok := hmem.CheckRef(); ok {
		return xerrors.NewXError("请先解除合并包厢后再尝试退出包厢。")
	}
	if clb.DBClub.IsMemExit {
		cer := clb.MemExitApply(hmem.UId)
		if cer != nil {
			return cer
		} else {
			if hmem.Partner > 0 {
				clb.Broadcast2Parnter(hmem.Partner, true, consts.MsgTypeHouseMemberApply_Ntf, &static.Msg_HouseMenberApplyNTF{
					HID:       clb.DBClub.HId,
					Uid:       hmem.UId,
					UUrl:      hmem.ImgUrl,
					NickName:  hmem.NickName,
					ApplyTime: time.Now().Unix(),
					IsOnline:  hmem.IsOnline,
					ApplyType: 1,
				})
			}

			clb.CustomBroadcast(consts.ROLE_ADMIN, false, false, false, consts.MsgTypeHouseMemberApply_Ntf, &static.Msg_HouseMenberApplyNTF{
				HID:       clb.DBClub.HId,
				Uid:       hmem.UId,
				UUrl:      hmem.ImgUrl,
				NickName:  hmem.NickName,
				ApplyTime: time.Now().Unix(),
				IsOnline:  hmem.IsOnline,
				ApplyType: 1,
			})

			return xerrors.HouseMemberExitError
		}
	}
	return nil
}

// 离开包厢
func (clb *Club) MemExit(uId int64) *xerrors.XError {
	// 包厢成员
	hmem := clb.GetMemByUId(uId)
	if hmem.UVitamin != 0 {
		return xerrors.NewXError("玩家比赛分必须为0才能退出。")
	}
	err := clb.checkMemExit(hmem)
	if err != nil {
		return err
	}
	houseFloorIds, _ := clb.GetFloors()
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		floor := clb.GetFloorByFId(fid)
		if floor == nil {
			continue
		}
		floor.RemVipUsers(hmem.UId)
	}
	return clb.UserExit(hmem, true, nil)
}

// 离开包厢
func (clb *Club) UserExit(hmem *HouseMember, ntf bool, tx *gorm.DB) *xerrors.XError {
	if hmem == nil {
		return xerrors.UserNotExistError
	}
	hp, err := GetDBMgr().GetDBrControl().GetPerson(hmem.UId)
	if hp == nil || err != nil {
		return xerrors.UserNotExistError
	}
	clb.RedisPub(consts.MsgTypeHouseMemOffline_NTF,
		static.Msg_HC_HouseMemOnline{Hid: clb.DBClub.HId, Uid: hmem.UId, Online: false})
	//msg := fmt.Sprintf("%sID:%d退出了包厢，当前疲劳值为%d", hp.Nickname, hmem.UId, hmem.UVitamin)
	msg := fmt.Sprintf("<color=#00A70C>%sID:%d</color>退出了包厢", hp.Nickname, hmem.UId)
	CreateClubMassage(clb.DBClub.Id, hmem.UId, ExitHouse, msg)
	return clb.MemDelete(hmem.UId, true, tx)
}

// ! 剔除包厢
func (clb *Club) MemKick(optId int64, uId int64) *xerrors.XError {
	if optId == uId {
		return xerrors.InvalidPermission
	}
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	optMember := clb.GetMemByUId(optId)
	if optMember == nil {
		return xerrors.InValidHouseMemberError
	}
	isOK, _ := GetMRightMgr().CheckRight(optMember, MinorOutPlayer)
	if !isOK {
		return xerrors.InvalidPermission
	}
	kickMember := clb.GetMemByUId(uId)
	if kickMember == nil {
		return xerrors.UserNotExistError
	}
	// 权限判定
	if optMember.URole != consts.ROLE_CREATER {
		if optMember.Lower(kickMember.URole) {
			return xerrors.InvalidPermission
		}
	}

	// 如果是队长 则走队长剔除名下玩家流程
	if optMember.IsPartner() || optMember.IsVicePartner() {
		if clb.DBClub.PartnerKick {
			return clb.PartnerMemKick(optMember, kickMember)
		} else {
			return xerrors.NewXError("包厢未开启队长踢人权限，请联系盟主处理。")
		}
	}

	if _, ok := kickMember.CheckRef(); ok {
		return xerrors.NewXError("该队长为合并包厢盟主，无法被踢出包厢。")
	}

	//var kickVitamin float64
	//kickVitamin = float64(kickMember.UVitamin) / 100.0
	if kickMember.UVitamin != 0 {
		return xerrors.NewXError("玩家比赛分必须为0才能退出。")
		//cli := GetDBMgr().Redis
		//kickMember.Lock(cli)
		//tx := GetDBMgr().GetDBmControl().Begin()
		//change := kickMember.UVitamin
		//_, after, err := kickMember.VitaminIncrement(optId, -1*kickMember.UVitamin, models.AdminSend, tx)
		//if err != nil {
		//	xlog.Logger().Errorf("%v", err)
		//	tx.Rollback()
		//	kickMember.Unlock(cli)
		//	return xerrors.DBExecError
		//}
		//
		//xerr := clb.PoolChange(optId, models.AdminSend, change, tx)
		//if xerr != nil {
		//	xlog.Logger().Errorf("%v", xerr.Error())
		//	tx.Rollback()
		//	kickMember.Unlock(cli)
		//	return xerr
		//}
		////修改疲劳值统计管理节点信息
		//err = GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, kickMember.UId, after, tx)
		//if err != nil {
		//	xlog.Logger().Error(err)
		//	tx.Rollback()
		//	kickMember.Unlock(cli)
		//	return xerrors.DBExecError
		//}
		//err = tx.Commit().Error
		//if err != nil {
		//	xlog.Logger().Errorf("%v", err)
		//	tx.Rollback()
		//	kickMember.Unlock(cli)
		//	return xerrors.DBExecError
		//}
		//kickMember.Flush()
		//kickMember.Unlock(cli)
	}
	xer := clb.MemDelete(uId, true, nil)
	if xer != nil {
		xlog.Logger().Errorf("%v", xer.Error())
		return xer
	}
	//踢出玩家之后 要踢出VIP楼层
	for _, floor := range clb.Floors {
		vipUser := floor.GetVipUsersSet()
		_, ok2 := vipUser[uId]
		if ok2 {
			var ferr error
			ferr = floor.RemVipUsers(uId)
			if ferr != nil {
				xlog.Logger().Error(ferr)
			}
		}
	}

	// 在线通知
	kickMember.SendMsg(consts.MsgTypeHouseMemberKick_Ntf, &static.Ntf_HC_HouseMemberKick{clb.DBClub.HId})
	var msg string
	if optMember.URole == consts.ROLE_CREATER {
		msg = fmt.Sprintf("盟主将<color=#00A70C>%sID:%d</color>移除了包厢", kickMember.NickName, kickMember.UId)
	} else {
		msg = fmt.Sprintf("<color=#F93030>%sID:%d</color>将<color=#00A70C>%sID:%d</color>移除了包厢", optMember.NickName, optMember.UId, kickMember.NickName, kickMember.UId)
	}
	CreateClubMassage(clb.DBClub.Id, optMember.UId, KickHouseMem, msg)
	// 删除队长创建的经验
	clb.DelCreatePartnerExp(uId)
	return nil
}

// 队长提出名下玩家
func (clb *Club) PartnerMemKick(hoptmem, hmem *HouseMember) *xerrors.XError {
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}

	if !hoptmem.IsPartner() && !hoptmem.IsVicePartner() {
		return xerrors.InValidHousePartnerError
	}

	// 确定队长ID
	tgtPartnerId := hoptmem.UId
	if hoptmem.IsVicePartner() {
		tgtPartnerId = hoptmem.Partner
	}
	tgtPartner := clb.GetMemByUId(tgtPartnerId)
	if tgtPartner == nil {
		return xerrors.UserNotExistError
	}

	// 队长/副队长 可管理 队长2 下面的成员
	isOK, _ := GetMRightMgr().CheckRight(hoptmem, MinorManageSuperior)
	if isOK {
		bIsPartnerMem, nLv := hmem.IsHaveTgtSuperior(tgtPartner.UId)
		if bIsPartnerMem == false {
			return xerrors.InvalidPermission
		}

		// 队长/副队长 不能 操作 二级队长/副队长
		if (hoptmem.IsPartner() || hoptmem.IsVicePartner()) && nLv > 1 && (hmem.IsPartner() || hmem.IsVicePartner()) {
			return xerrors.InvalidPermission
		}
	} else {
		if hmem.Partner != tgtPartner.UId {
			return xerrors.InvalidPermission
		}
	}

	// 副队长之间 不能相互操作
	if hoptmem.IsVicePartner() && hmem.IsVicePartner() {
		return xerrors.InvalidPermission
	}

	if _, ok := hmem.CheckRef(); ok {
		return xerrors.NewXError("该队长为合并包厢盟主，无法被踢出包厢。")
	}

	//var kickVitamin float64
	//kickVitamin = float64(hmem.UVitamin) / 100.0
	if hmem.UVitamin != 0 {
		if res := tgtPartner.UVitamin + hmem.UVitamin; res < 0 {
			return xerrors.HousePartnerKickVitaminError
		} else {
			cli := GetDBMgr().Redis
			hmem.Lock(cli)
			tgtPartner.Lock(cli)
			tx := GetDBMgr().GetDBmControl().Begin()
			var (
				memAfterVitamin, partnerAfterVitamin int64
				dbError                              error
			)
			deferFunc := func(tx *gorm.DB) (xerr *xerrors.XError) {
				if dbError != nil {
					xlog.Logger().Error(dbError)
					tx.Rollback()
					xerr = xerrors.DBExecError
				} else {
					dbError = GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, hmem.UId, memAfterVitamin, tx)
					if dbError != nil {
						xlog.Logger().Error(dbError)
						tx.Rollback()
						xerr = xerrors.DBExecError
					}
					dbError = GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, tgtPartner.UId, partnerAfterVitamin, tx)
					if dbError != nil {
						xlog.Logger().Error(dbError)
						tx.Rollback()
						xerr = xerrors.DBExecError
					}
					dbError = tx.Commit().Error
					if dbError != nil {
						xlog.Logger().Error(dbError)
						tx.Rollback()
						xerr = xerrors.DBExecError
					} else {
						hmem.Flush()
						tgtPartner.Flush()
					}
				}
				hmem.Unlock(cli)
				tgtPartner.Unlock(cli)
				return xerr
			}

			_, partnerAfterVitamin, dbError = tgtPartner.VitaminIncrement(tgtPartner.UId, hmem.UVitamin, models.PartnerSend, tx)
			if dbError != nil {
				return deferFunc(tx)
			}

			_, memAfterVitamin, dbError = hmem.VitaminIncrement(tgtPartner.UId, -1*hmem.UVitamin, models.PartnerSend, tx)
			if xerr := deferFunc(tx); xerr != nil {
				return xerr
			}
		}
	}
	custerr := clb.MemDelete(hmem.UId, true, nil)
	if custerr != nil {
		return custerr
	}

	// 在线通知
	hp := GetPlayerMgr().GetPlayer(hmem.UId)
	if hp != nil {
		hp.SendMsg(consts.MsgTypeHouseMemberKick_Ntf, &static.Ntf_HC_HouseMemberKick{clb.DBClub.HId})
	}
	p, _ := GetDBMgr().GetDBrControl().GetPerson(hmem.UId)
	if p == nil {
		return xerrors.InvalidIdError
	}
	hpOpt, _ := GetDBMgr().GetDBrControl().GetPerson(hoptmem.UId)
	if hpOpt == nil {
		return xerrors.InvalidIdError
	}
	CreateClubMassage(clb.DBClub.Id, hoptmem.UId, KickHouseMem, fmt.Sprintf("队长/副队长<color=#F93030>%sID:%d</color>将其名下玩家<color=#00A70C>%sID:%d</color>移除了包厢", hpOpt.Nickname, hpOpt.Uid, p.Nickname, p.Uid))
	//CreateClubMassage(clb.DBClub.Id, partner.UId, KickHouseMem, fmt.Sprintf("队长/副队长%sID:%d将其名下玩家%sID:%d移除了包厢,比赛分为:%.2f", hpOpt.Nickname, hpOpt.Uid, p.Nickname, p.Uid, kickVitamin))
	return nil
}

// 通过申请
func (clb *Club) MemAgree(optId int64, uId int64, tx *gorm.DB) *xerrors.XError {
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	// 包厢成员
	hoptmem := clb.GetMemByUId(optId)
	if hoptmem == nil {
		return xerrors.UserNotExistError
	}

	hmem := clb.GetMemByUId(uId)
	if hmem == nil {
		return xerrors.InValidHouseMemberError
	}
	if hmem.URole != consts.ROLE_APLLY {
		return xerrors.InValidHouseMemberRoleError
	}

	// 队长和副队长的权限
	bHavePartnerPower := false
	if clb.DBClub.IsPartnerApply {
		if hoptmem.IsPartner() && hmem.Partner == hoptmem.UId {
			bHavePartnerPower = true
		}
		if hoptmem.IsVicePartner() && hmem.Partner == hoptmem.Partner {
			bHavePartnerPower = true
		}
	}

	if hoptmem.Lower(consts.ROLE_ADMIN) && !(clb.DBClub.IsPartnerApply && bHavePartnerPower) {
		return xerrors.InvalidPermission
	}

	if hmem.UVitamin != 0 {
		_, after, err := hmem.VitaminIncrement(uId, -1*hmem.UVitamin, models.JoinClear, nil)
		if err == nil {
			//修改疲劳值统计管理节点信息
			GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, hmem.UId, after, nil)
		}
	}

	err := clb.ChangeRole(optId, uId, consts.ROLE_MEMBER)
	if err == nil {
		hmem.AgreeTime = time.Now().Unix()
		hmem.URole = consts.ROLE_MEMBER
		hmem.Flush()
		clb.TotalMemIncr(hmem.UId)

		if hmem != nil {
			// 向数据库同步同意时间字段
			//更新数据库
			attrMap := make(map[string]interface{})
			attrMap["apply_time"] = time.Unix(hmem.ApplyTime, 0)
			attrMap["agree_time"] = time.Unix(hmem.AgreeTime, 0)
			GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, hmem.UId).Update(attrMap)

			// 移除此圈历史队长信息
			GetDBMgr().RemoveMemberLeaveHousePartner(hmem.DHId, hmem.UId)
		}
		if hmem.Ref > 0 {
			oldHouse := GetClubMgr().GetClubHouseById(hmem.Ref)
			if oldHouse == nil {
				return xerrors.InValidHouseError
			}
			oldMem := oldHouse.GetMemByUId(hmem.UId)
			if oldMem != nil && oldHouse.DBClub.Id != clb.DBClub.Id && oldMem.URole == consts.ROLE_APLLY {
				err = oldHouse.MemAgree(oldHouse.DBClub.UId, hmem.UId, tx)
			}
		}
	}
	return err
}

// 拒绝申请
func (clb *Club) MemRefused(optId int64, uId int64) *xerrors.XError {
	// 包厢成员
	hoptmem := clb.GetMemByUId(optId)
	if hoptmem == nil {
		return xerrors.UserNotExistError
	}

	hmem := clb.GetMemByUId(uId)
	if hmem == nil {
		return xerrors.InValidHouseMemberError
	}
	if hmem.URole != consts.ROLE_APLLY {
		return xerrors.InValidHouseMemberRoleError
	}

	// 队长和副队长的权限
	bHavePartnerPower := false
	if clb.DBClub.IsPartnerApply {
		if hoptmem.IsPartner() && hmem.Partner == hoptmem.UId {
			bHavePartnerPower = true
		}
		if hoptmem.IsVicePartner() && hmem.Partner == hoptmem.Partner {
			bHavePartnerPower = true
		}
	}

	if hoptmem.Lower(consts.ROLE_ADMIN) && !(clb.DBClub.IsPartnerApply && bHavePartnerPower) {
		return xerrors.InvalidPermission
	}

	custerr := clb.MemDelete(uId, true, nil)
	if custerr != nil {
		return custerr
	}

	hp := GetPlayerMgr().GetPlayer(hmem.UId)
	if hp == nil {
		return nil
	}

	ntf := new(static.Ntf_HC_HouseMemberRoleGen)
	ntf.HId = hmem.HId
	ntf.UId = hmem.UId
	ntf.OURole = hmem.URole
	ntf.URole = -1
	ntf.HName = clb.DBClub.Name
	hp.SendMsg(consts.MsgTypeHouseMemberRoleGen_Ntf, ntf)

	return nil
}

// 角色变更
func (clb *Club) ChangeRole(optId int64, uId int64, role int) *xerrors.XError {
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	// 包厢成员
	hoptmem := clb.GetMemByUId(optId)
	if hoptmem == nil {
		return xerrors.UserNotExistError
	}

	hmem := clb.GetMemByUId(uId)
	if hmem == nil {
		return xerrors.UserNotExistError
	}
	cli := GetDBMgr().Redis
	hmem.Lock(cli)
	defer hmem.Unlock(cli)
	oldRole := hmem.URole
	// 管理数量
	if role == consts.ROLE_ADMIN {
		ids := clb.GetMemUIdsByRole(consts.ROLE_ADMIN)
		if len(ids) >= GetServer().ConHouse.AdminMax {
			return xerrors.HouseMemAdminMaxOverFlowError
		}
		if hmem.Partner == 1 {
			return xerrors.HouseRefusePartnerError
		}
		if hmem.Partner > 1 {
			return xerrors.NewXError(fmt.Sprintf("该玩家已被绑定给队长ID：%d，请先解除绑定关系再做尝试", hmem.Partner))
		}
		if hmem.VitaminAdmin {
			return xerrors.InValidHouseMemberRoleError
		}
	}

	if hoptmem.IsPartner() || hoptmem.IsVicePartner() {
		if !clb.DBClub.IsPartnerApply || role != consts.ROLE_MEMBER {
			return xerrors.InvalidPermission
		}
	} else {
		// 权限
		if role == consts.ROLE_CREATER ||
			role == consts.ROLE_APLLY ||
			hoptmem.Lower(consts.ROLE_ADMIN) ||
			hoptmem.Lower(hmem.URole) {
			return xerrors.InvalidPermission
		}

		// 调整权限
		if hoptmem.URole == consts.ROLE_ADMIN {
			if role <= consts.ROLE_ADMIN {
				return xerrors.InvalidPermission
			}
		}
	}

	// 队长当前是否被限制游戏 新加入的成员状态跟随队长
	if hmem.Partner > 0 {
		hoptPartner := clb.GetMemByUId(hmem.Partner)
		if hoptPartner != nil && hoptPartner.IsLimitGame == true {
			clb.LimitUserGame(hmem.Partner, hmem.UId, false)
		}
	}

	// 更新成员信息
	hmem.URole = role

	//// 队长审核，需要绑定在自己名下
	//if hoptmem.Partner == 1 && hmem.Partner == 0 {
	//	hmem.Partner = hoptmem.UId
	//}

	// db
	err := GetDBMgr().HouseMemberUpdate(clb.DBClub.Id, hmem)
	if err != nil {
		return xerrors.DBExecError
	}

	//更新数据库
	attrMap := make(map[string]interface{})
	attrMap["partner"] = hmem.Partner
	attrMap["urole"] = hmem.URole
	GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, hmem.UId).Update(attrMap)

	GetMRightMgr().setRoleUpdateRight(hmem, false)

	// 通知
	person := GetPlayerMgr().GetPlayer(uId)
	if person == nil {
		return nil
	}
	if hmem.URole == consts.ROLE_BLACK {
		ntf := new(static.Ntf_HC_HouseMemberKick)
		ntf.HId = hmem.HId
		person.SendMsg(consts.MsgTypeHouseMemberKick_Ntf, ntf)
	} else {
		ntf := new(static.Ntf_HC_HouseMemberRoleGen)
		ntf.HId = hmem.HId
		ntf.UId = hmem.UId
		ntf.OURole = oldRole
		ntf.URole = hmem.URole
		ntf.HName = clb.DBClub.Name
		person.SendMsg(consts.MsgTypeHouseMemberRoleGen_Ntf, ntf)
	}
	var msg string
	if oldRole == consts.ROLE_ADMIN && role != consts.ROLE_ADMIN {
		msg = fmt.Sprintf("盟主将%sID:%d取消管理", person.Info.Nickname, person.Info.Uid)
	} else if role == consts.ROLE_ADMIN && oldRole != consts.ROLE_ADMIN {
		msg = fmt.Sprintf("盟主将%sID:%d设为管理", person.Info.Nickname, person.Info.Uid)
	}
	if msg != "" {
		CreateClubMassage(clb.DBClub.Id, hoptmem.UId, AdminChange, msg)
	}
	return nil
}

// ! 创建楼层
func (clb *Club) FloorCreate(uId int64, frule static.FRule) (*HouseFloor, *xerrors.XError) {

	// 获得mem信息
	mem := clb.GetMemByUId(uId)
	if mem == nil {
		return nil, xerrors.InValidHouseMemberError
	}

	//权限判定
	if mem.URole != consts.ROLE_CREATER {
		return nil, xerrors.InvalidPermission
	}

	// 创建楼层时 默认gps/voice为true
	frule.Gps = true
	frule.Voice = true
	frule.GVoiceOk = true

	// 规则合法判定
	var msgct static.Msg_CreateTable
	msgct.KindId = frule.KindId
	msgct.PlayerNum = frule.PlayerNum
	msgct.RoundNum = frule.RoundNum
	msgct.CostType = frule.CostType
	msgct.Restrict = frule.Restrict
	msgct.GVoice = frule.GVoice
	msgct.GameConfig = frule.GameConfig
	msgct.FewerStart = frule.FewerStart
	msgct.Gps = frule.Gps
	msgct.Voice = frule.Voice
	msgct.GVoiceOk = frule.GVoiceOk

	config, custerr := validateCreateTableParam(&msgct, true)
	if custerr != nil {
		return nil, custerr
	}

	// 校验赋值
	frule.KindId = config.KindId
	frule.PlayerNum = config.MaxPlayerNum
	frule.RoundNum = config.RoundNum
	frule.CostType = config.CostType
	frule.Restrict = "false"
	if config.Restrict {
		frule.Restrict = "true"
	}

	frule.GVoice = config.GVoice
	frule.FewerStart = config.FewerStart
	frule.GameConfig = config.GameConfig

	floor := new(HouseFloor)
	floor.Init()
	floor.HId = clb.DBClub.HId
	floor.DHId = clb.DBClub.Id
	floor.Rule = frule
	floor.IsAlive = true
	floor.IsVitamin = clb.DBClub.IsVitamin
	floor.IsGamePause = clb.DBClub.IsGamePause

	floor.DataLock.Lock() //TODO：使用defer
	for i := 0; i < GetServer().ConHouse.TableNum; i++ {
		floor.Tables[i] = NewHFT(frule.PlayerNum, i)
	}
	floor.DataLock.Unlock()

	//数据
	id, err := GetDBMgr().HouseFloorCreate(floor, nil)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, xerrors.DBExecError
	}

	//内存
	floor.Id = id
	clb.FloorLock.CustomLock()
	clb.Floors[floor.Id] = floor
	clb.FloorLock.CustomUnLock()

	go floor.StartSync()

	// 通知
	ntf := new(static.Ntf_HC_HouseFloorCreate)
	ntf.HId = floor.HId
	ntf.FId = floor.Id
	ntf.FRule = floor.Rule
	ntf.VitaminLowLimit = static.SwitchVitaminToF64(floor.VitaminLowLimit)
	ntf.VitaminHighLimit = static.SwitchVitaminToF64(floor.VitaminHighLimit)
	area := GetAreaGameByKid(floor.Rule.KindId)
	if area != nil {
		ntf.ImageUrl = area.Icon
		ntf.KindName = area.Name
		ntf.PackageName = area.PackageName
	}
	house := GetClubMgr().GetClubHouseByHId(floor.HId)
	if house != nil {
		house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorCreate_Ntf, ntf)
	} else {
		floor.BroadCast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorCreate_Ntf, ntf)
	}

	return floor, nil
}

// ! 删除楼层
func (clb *Club) FloorDelete(fId int64, tx *gorm.DB) *xerrors.XError {

	// 楼层
	floor := clb.GetFloorByFId(fId)
	if floor == nil {
		return xerrors.InValidHouseFloorError
	}

	// 桌子
	floor.DataLock.RLock()
	for _, hft := range floor.Tables {
		if hft.TId != 0 {
			floor.DataLock.RUnlock()
			return xerrors.ExistHouseTablePlayingError
		}
	}
	floor.DataLock.RUnlock()

	// 内存
	clb.FloorLock.CustomLock()
	delete(clb.Floors, floor.Id)
	clb.FloorLock.CustomUnLock()

	// DB
	err := GetDBMgr().HouseFloorDelete(floor, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Ntf_HC_HouseFloorDelete)
	ntf.HId = floor.HId
	ntf.FId = floor.Id
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseFloorDelete_Ntf, ntf)

	return nil
}

func (clb *Club) memDelete(uid int64) {
	// 删除活跃数据
	for _, f := range clb.Floors {
		actmem := f.GetHFM(uid)
		if actmem != nil {
			person := GetPlayerMgr().GetPlayer(uid)
			if person != nil {
				ChOptMemOut(f, nil, &person.Info, nil)
			}
			break
		}
	}
	clb.TotalMemDecr(uid)
	clb.OnlineMemDecr(uid)
}

// ! 删除成员
func (clb *Club) MemDelete(uId int64, ntf bool, tx *gorm.DB) *xerrors.XError {
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	mem := clb.GetMemByUId(uId)
	if mem == nil {
		return xerrors.UserNotExistError
	}
	//退圈清理权限
	_, e := GetMRightMgr().deleteRightByHidUid(int(clb.DBClub.Id), uId)
	if e != nil {
		return xerrors.DBExecError
	}
	if mem.IsPartner() {
		if xe := clb.PartnerDelete(mem, tx); xe != nil {
			return xe
		}
	}

	// DB
	err := GetDBMgr().HouseMemberDelete(clb.DBClub.Id, mem.Id, mem.UId, mem.URole, mem.UVitamin, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError
	}

	clb.memDelete(uId)

	if ntf {
		ntf := new(static.Ntf_HC_HouseMemberRoleGen)
		ntf.HId = mem.HId
		ntf.UId = mem.UId
		ntf.OURole = mem.URole
		ntf.URole = -1
		ntf.HName = clb.DBClub.Name
		clb.CustomBroadcast(consts.ROLE_ADMIN, true, false, false, consts.MsgTypeHouseMemberRoleGen_Ntf, ntf)
		// 被删除的人收不到上面的ntf，下补一条
		mem.SendMsg(consts.MsgTypeHouseMemberKick_Ntf, &static.Ntf_HC_HouseMemberKick{HId: clb.DBClub.HId})
	}

	// 如果玩家有队长,记录队长信息
	if mem.Partner > 1 {
		GetDBMgr().AddMemberLeaveHousePartner(mem.DHId, mem.UId, mem.Partner)
	}

	return nil
}

// 修改 名称、公告
func (clb *Club) ModifyNN(uId int64, name string, notify string) *xerrors.XError {

	mem := clb.GetMemByUId(uId)
	if mem == nil {
		return xerrors.InValidHouseMemberError
	}
	isOK, _ := GetMRightMgr().CheckRight(mem, MinorSetTeaName)
	if !isOK && clb.DBClub.Name != name {
		return xerrors.InvalidPermission
	}
	isOK, _ = GetMRightMgr().CheckRight(mem, MinorSetNotice)
	if !isOK && clb.DBClub.Notify != notify {
		return xerrors.InvalidPermission
	}
	clb.DBClub.Name = name
	clb.DBClub.Notify = notify

	clb.flush()

	var ntf static.Ntf_HC_HouseBaseNNmodify
	ntf.HId = clb.DBClub.HId
	ntf.HName = clb.DBClub.Name
	ntf.HNotify = clb.DBClub.Notify
	ntf.HDialog = clb.DBClub.Dialog
	ntf.HDialogActive = clb.DBClub.DialogActive
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseBaseNNModify_Ntf, &ntf)

	return nil
}

// 取消绑定队长下级
func (clb *Club) UnbindPartnerJunior(oid /*操作人uid*/, jid /*下级uid*/ int64) *xerrors.XError {
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	operator := clb.GetMemByUId(oid)
	if operator == nil || operator.Lower(consts.ROLE_CREATER) {
		return xerrors.InvalidPermission
	}

	junior := clb.GetMemByUId(jid)
	cli := GetDBMgr().Redis
	junior.Lock(cli)
	defer junior.Unlock(cli)

	if junior == nil || junior.Lower(consts.ROLE_MEMBER) {
		return xerrors.UserNotExistError
	}

	if !junior.IsPartner() {
		return xerrors.InValidHousePartnerError
	}

	superIor := clb.GetMemByUId(junior.Superior)
	if superIor == nil || superIor.Lower(consts.ROLE_MEMBER) {
		return xerrors.InValidHousePartnerError
	}

	if !superIor.IsPartner() {
		return xerrors.InValidHousePartnerError
	}

	tx := GetDBMgr().GetDBmControl().Begin()
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
			xlog.Logger().Error(err)
		} else {
			if err = tx.Commit().Error; err != nil {
				tx.Rollback()
				xlog.Logger().Error(err)
			}
		}
	}()
	err = static.ErrorsCheck(
		clb.SyncJuniorRoyaltyConfig(tx, superIor, junior.UId),
		tx.Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, junior.UId).Update("superior", 0).Error)

	if err != nil {
		return xerrors.DBExecError
	}
	junior.Superior = 0
	junior.Flush()
	return nil
}

// 绑定队长下级
func (clb *Club) BindPartnerJunior(oid /*操作人uid*/, sid /*上级uid*/, jid /*下级uid*/ int64) *xerrors.XError {
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	operator := clb.GetMemByUId(oid)
	if operator == nil || operator.Lower(consts.ROLE_MEMBER) {
		return xerrors.InvalidPermission
	}
	if operator.URole > consts.ROLE_CREATER && !operator.IsPartner() {
		return xerrors.InvalidPermission
	}

	superior := clb.GetMemByUId(sid)
	if superior == nil || superior.Lower(consts.ROLE_MEMBER) {
		return xerrors.UserNotExistError
	}
	if !superior.IsPartner() {
		return xerrors.InValidHousePartnerError
	}

	var sssid int64

	if superior.IsJunior() {
		ss := clb.GetMemByUId(superior.Superior)
		if ss == nil {
			return xerrors.InValidHousePartnerError
		}
		if ss.IsJunior() {
			sss := clb.GetMemByUId(ss.Superior)
			if sss == nil {
				return xerrors.InValidHousePartnerError
			}
			sssid = sss.UId
			if sss.IsJunior() {
				return xerrors.NewXError("操作失败：超出最大队长级数。")
			}
		}
	}

	junior := clb.GetMemByUId(jid)
	if junior == nil || junior.Lower(consts.ROLE_MEMBER) {
		return xerrors.UserNotExistError
	}
	cli := GetDBMgr().Redis
	junior.Lock(cli)
	defer junior.Unlock(cli)
	if operator.URole == consts.ROLE_CREATER {
		if !junior.IsPartner() {
			return xerrors.InValidHousePartnerError
		}
	} else {
		if junior.Partner != superior.UId {
			if junior.Superior == superior.UId {
				return xerrors.NewXError("对方已经是您的下级队长。")
			}
			return xerrors.HouseRefusePartnerError
		}
	}

	if err := clb.UpdateJuniorRoyaltyBySuperior(superior.UId, superior.Superior, sssid, junior.UId); err != nil {
		return err
	}

	junior.Partner = 1
	junior.Superior = superior.UId
	junior.Flush()
	attrMap := make(map[string]interface{})
	attrMap["partner"] = 1
	attrMap["superior"] = superior.UId
	GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, junior.UId).Update(attrMap)
	GetMRightMgr().setRoleUpdateRight(junior, false)
	return nil
}

// 修改玩家合伙人上级状态
func (clb *Club) UpdateJuniorRoyaltyBySuperior(sid, ssid, sssid, jid int64) *xerrors.XError {
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	var (
		sup models.HousePartnerPyramidFloors
		ok  bool
	)
	houseFloorIds, _ := clb.GetFloors()
	floorCosts, err := clb.GetFloorsSingleCostMap()

	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError
	}

	if sssid > 0 {
		res := GetHousePartPartnersPyramid(clb.DBClub.Id, sid, ssid, sssid)
		sup, ok = res[sid]
		if !ok {
			sup = make(models.HousePartnerPyramidFloors, 0)
		}
		ssup, ok1 := res[ssid]
		if !ok1 {
			ssup = make(models.HousePartnerPyramidFloors, 0)
		}
		sssup, ok1 := res[sssid]
		if !ok1 {
			sssup = make(models.HousePartnerPyramidFloors, 0)
		}
		if err = FixClubPartnersPyramidBySuperSuper(&sup, &ssup, &sssup, clb.DBClub.Id, sid, houseFloorIds); err != nil {
			return xerrors.DBExecError
		}
	} else if ssid > 0 {
		res := GetHousePartPartnersPyramid(clb.DBClub.Id, sid, ssid)
		sup, ok = res[sid]
		if !ok {
			sup = make(models.HousePartnerPyramidFloors, 0)
		}
		ssup, ok1 := res[ssid]
		if !ok1 {
			ssup = make(models.HousePartnerPyramidFloors, 0)
		}

		if err = FixClubPartnersPyramidBySuper(&sup, &ssup, clb.DBClub.Id, sid, houseFloorIds); err != nil {
			return xerrors.DBExecError
		}
	} else {
		res := GetHousePartPartnersPyramid(clb.DBClub.Id, sid)
		sup, ok = res[sid]
		if !ok {
			sup = make(models.HousePartnerPyramidFloors, 0)
		}
		if fixer := FixClubPartnersPyramidForTop(&sup, clb.DBClub.Id, sid, houseFloorIds, floorCosts); fixer != nil {
			xlog.Logger().Error(fixer)
			return xerrors.DBExecError
		}
	}

	costs := make(map[int64]int64)
	for i := 0; i < len(houseFloorIds); i++ {
		fid := houseFloorIds[i]
		config := sup.GetPyramidByFid(fid)
		if config != nil && config.ConfiguredRoyaltyPercent() && config.Configurable() {
			costs[fid] = config.Total * config.RealRoyaltyPercent() / PartnerPercentHigherLimit
		} else {
			costs[fid] = InvalidVitaminCost
		}
	}

	tx := GetDBMgr().GetDBmControl().Begin()
	memMap := GetDBMgr().GetHouseMemMap(clb.DBClub.Id)
	if memMap == nil {
		xlog.Logger().Error("UpdateJuniorRoyaltyBySuperior mem map is nil")
		return xerrors.DBExecError
	}

	if !static.TxCommit(tx, UpdateHousePartnerTotal(tx, memMap, clb.DBClub.Id, jid, costs)) {
		return xerrors.DBExecError
	}

	return nil
}

// 修改玩家队长状态
func (clb *Club) ModifyPartner(memId int64, id int64) *xerrors.XError {
	// 为保证合并包厢/撤销合并包厢效率，及redis复写问题，这里如果正常执行合并包厢/撤销合并包厢，提示用户稍后重试。
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	clb.OptLock.CustomLock()
	defer clb.OptLock.CustomUnLock()

	mem := clb.GetMemById(memId)
	if mem == nil {
		return xerrors.InValidHouseMemberError
	}
	cli := GetDBMgr().Redis

	mem.Lock(cli)
	defer mem.Unlock(cli)
	if mem.UId == clb.DBClub.UId {
		return xerrors.HouseMemAllReadyPartnerError
	}
	if mem.VitaminAdmin {
		return xerrors.InValidHouseMemberError
	}
	if mem.Partner > 1 && id == 1 {
		return xerrors.DelPartnerFirst
	}
	if id == 1 && mem.URole == consts.ROLE_ADMIN {
		return xerrors.HouseRefusePartnerError
	}
	//if id > 1 {
	//	newPartnerU, err := GetDBMgr().GetDBrControl().GetPerson(id)
	//	if err != nil {
	//		xlog.Logger().Error(err)
	//		return xerrors.DBExecError
	//	}
	//	if newPartnerU.Card < 0 {
	//		return xerrors.NewXError("队长房卡不能低于0")
	//	}
	//}

	before := mem.Partner
	mem.Partner = id
	if mem.Partner != before && mem.IsVicePartner() { // 如果副队长的队长发生变化 则自动取消副队长权限
		mem.VicePartner = false
		mem.SendMsg(consts.MsgTypeHouseVicePartnerSet_Ntf, &static.MsgHouseVicePartnerSet{
			Hid:         clb.DBClub.HId,
			OptUid:      clb.DBClub.UId,
			Uid:         mem.UId,
			VicePartner: false,
		})
	}
	mem.Flush()

	//更新数据库
	attrMap := make(map[string]interface{})
	attrMap["partner"] = id
	attrMap["vice_partner"] = mem.VicePartner
	GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, mem.UId).Updates(attrMap)
	GetMRightMgr().setRoleUpdateRight(mem, false)
	return nil
}

// 防沉迷开关
func (clb *Club) ModifyVitaminStatus(status bool) *xerrors.XError {
	clb.OptLock.CustomLock()
	defer clb.OptLock.CustomUnLock()

	clb.DBClub.IsVitamin = status
	return nil
}

// 疲劳值 管理员可见
func (clb *Club) ModifyVitaminAdminHide(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()

	clb.DBClub.IsVitaminHide = status
	return nil
}

// 疲劳值 管理员可调
func (clb *Club) ModifyVitaminAdminModify(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()
	clb.DBClub.IsVitaminModi = status
	return nil
}

// 疲劳值 管理员可调
func (clb *Club) ModifyVitaminNoSkip(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()
	clb.DBClub.NoSkipVitaminSet = status
	return nil
}

// 疲劳值 队长可见
func (clb *Club) ModifyVitaminPartnerHide(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()

	clb.DBClub.IsPartnerHide = status
	return nil
}

// 疲劳值 队长可见
func (clb *Club) ModifyVitaminPartnerModi(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()

	clb.DBClub.IsPartnerModi = status
	return nil
}

// 疲劳值 游戏暂定
func (clb *Club) ModifyVitaminGamePause(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()

	clb.DBClub.IsGamePause = status
	return nil
}

// 疲劳值 赠送
func (clb *Club) ModifyVitaminMemberSend(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()
	clb.DBClub.IsMemberSend = status
	return nil
}

// 疲劳值 赠送
func (clb *Club) ModifyDisSetJuniorVitamin(status bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()
	clb.DBClub.DisVitaminJunior = status
	return nil
}

// 包厢设置在线人数隐藏
func (clb *Club) ModifyRewardBalanced(rb bool) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()
	clb.DBClub.RewardBalanced = rb
	return nil
}

// 疲劳值 赠送
func (clb *Club) ModifyRewardBalancedType(typ int) *xerrors.XError {
	clb.OptLock.Lock()
	defer clb.OptLock.Unlock()
	clb.DBClub.RewardBalancedType = typ
	return nil
}

// 防沉迷开关
// func (self *Club) ModifyVitaminOptValues(deductType int, deductCount, limitGame, limitePause int64) *xerrors.XError {
// 	self.OptLock.Lock()
// 	defer self.OptLock.Unlock()
//
// 	self.VitaminDeductType = deductType
// 	self.VitaminLowLimit = limitGame
// 	self.VitaminDeductCount = deductCount
// 	self.VitaminLowLimitPause = limitePause
// 	return nil
// }

func (clb *Club) GetPartnerIDs() []int64 {
	memlist := make([]int64, 0)
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.Partner == 1 {
			memlist = append(memlist, mem.UId)
		}
	}

	return memlist
}

func (clb *Club) AddMixFloor(fids []int64, tableNum int) error {
	tx := GetDBMgr().GetDBmControl().Begin()
	defer tx.Rollback()
	sql := `update house set  mix_active = true where id = %d;`
	sql = fmt.Sprintf(sql, clb.DBClub.Id)
	db := tx.Exec(sql)
	if db.Error != nil {
		xlog.Logger().Errorf("添加混合大厅失败1：%+v\n", db.Error)
		return errors.New("params error")
	}
	err := tx.Model(models.HouseFloor{}).
		Where("id in (?) and hid = ?", fids, clb.DBClub.Id).
		Updates(map[string]interface{}{"is_mix": true, "mix_table_num": tableNum}).
		Error
	if err != nil {
		xlog.Logger().Errorf("添加混合大厅失败：%+v,\n", db.Error)
		return err
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Errorf("数据提交失败：%+v,\n", db.Error)
		return err
	}
	clb.DBClub.MixActive = true
	clb.DBClub.MixTableNum = tableNum
	GetDBMgr().HouseUpdate(clb)
	for _, fid := range fids {
		floor := clb.GetFloorByFId(fid)
		if floor == nil {
			continue
		}
		// 桌子
		for k, hft := range floor.Tables {
			if hft.TId != 0 {
				return xerrors.ExistHouseTablePlayingError
			} else {
				delete(floor.Tables, k)
			}
		}
		floor.IsMix = true
		err := GetDBMgr().HouseFloorUpdate(floor)
		if err != nil {
			xlog.Logger().Errorf("update house floor error:%+v\n", err)
			continue
		}
		for i := 0; i < tableNum; i++ {
			hft := NewHFT(floor.Rule.PlayerNum, i)
			hft.IsDefault = true
			floor.DataLock.Lock()
			floor.Tables[i] = hft
			floor.DataLock.Unlock()
		}
	}
	return nil
}

func (clb *Club) UpdateMixInfo(fids []int64, tableNum int, isActive bool, emptyTblMax int, createTblType int, newTblSortType int) *xerrors.XError { // isEmptyTblBack bool, tblSortType int, 舍弃参数
	// 如果包厢混排开关没变化 则校验有没有楼层变化
	if clb.DBClub.MixActive == isActive {
		for f, floor := range clb.Floors {
			nowMix := false
			for _, fid := range fids {
				if f == fid {
					nowMix = true
				}
			}
			// 如果楼层混排发生变化，就删掉所有桌子，再新增对应的桌子
			if nowMix != floor.IsMix {
				floor.DataLock.Lock()
				for k, table := range floor.Tables {
					if table.TId > 0 {
						xlog.Logger().Errorf("异常数据出现")
						floor.DataLock.Unlock()
						return xerrors.UserFloorGameError
					} else {
						delete(floor.Tables, k)
					}
				}
				floor.DataLock.Unlock()

				floor.IsMix = nowMix
				// 如果楼层是要开启混排的楼层 则新增选择的桌子数量
				if clb.DBClub.MixActive && floor.IsMix {
					if clb.DBClub.TableJoinType == consts.AutoAdd {
						// 自动加桌不用创建桌子 代码干掉了
					} else {
						for i := 0; i < tableNum; i++ {
							hft := NewHFT(floor.Rule.PlayerNum, i)
							hft.IsDefault = true
							floor.DataLock.Lock()
							floor.Tables[i] = hft
							floor.DataLock.Unlock()
						}
					}
				} else {
					// 如果是普通楼层则新增包厢每个楼层的默认桌子数量
					for i := 0; i < GetServer().ConHouse.TableNum; i++ {
						floor.DataLock.Lock()
						floor.Tables[i] = NewHFT(floor.Rule.PlayerNum, i)
						floor.DataLock.Unlock()
					}
					// 普通楼层开起同步协程
					go floor.StartSync()
				}
				// 更新缓存
				err := GetDBMgr().HouseFloorUpdate(floor)
				if err != nil {
					xlog.Logger().Errorf("update house floor error:%+v\n", err)
					continue
				}
			} else if clb.DBClub.MixTableNum != tableNum {
				// 如果楼层的混排没有发生变化，只是桌子数量发生了变化，这里同样需要重建桌子
				floor.DataLock.Lock()
				for k, table := range floor.Tables {
					if table.TId > 0 {
						xlog.Logger().Errorf("异常数据出现")
						floor.DataLock.Unlock()
						return xerrors.UserFloorGameError
					} else {
						delete(floor.Tables, k)
					}
				}
				floor.DataLock.Unlock()

				if clb.DBClub.MixActive && floor.IsMix {
					for i := 0; i < tableNum; i++ {
						hft := NewHFT(floor.Rule.PlayerNum, i)
						hft.IsDefault = true
						floor.DataLock.Lock()
						floor.Tables[i] = hft
						floor.DataLock.Unlock()
					}
				} else {
					for i := 0; i < GetServer().ConHouse.TableNum; i++ {
						floor.DataLock.Lock()
						floor.Tables[i] = NewHFT(floor.Rule.PlayerNum, i)
						floor.DataLock.Unlock()
					}
					go floor.StartSync()
				}

				err := GetDBMgr().HouseFloorUpdate(floor)
				if err != nil {
					xlog.Logger().Errorf("update house floor error:%+v\n", err)
					continue
				}
			}
			// 如果包厢开关没有变化/楼层开关没有变化/桌子数量没有变化/则不需要再做任何事情
		}
	} else if isActive {
		// 如果包厢的混排开关由关闭变成开启
		// 初始化所有楼层的混排开关为false
		for _, floor := range clb.Floors {
			if floor.IsMix {
				floor.IsMix = false
				err := GetDBMgr().HouseFloorUpdate(floor)
				if err != nil {
					xlog.Logger().Errorf("update house floor error:%+v\n", err)
					continue
				}
			}
		}
		// 再把选中的楼层混排开关置为开同时重置楼层桌子数据
		for _, fid := range fids {
			floor := clb.Floors[fid]
			if floor == nil {
				continue
			}
			floor.IsMix = true

			floor.DataLock.Lock()
			for k, table := range floor.Tables {
				if table.TId > 0 {
					xlog.Logger().Errorf("异常数据出现")
					floor.DataLock.Unlock()
					return xerrors.UserFloorGameError
				} else {
					delete(floor.Tables, k)
				}
			}
			floor.DataLock.Unlock()

			if clb.DBClub.TableJoinType == consts.AutoAdd {
				// 自动加桌不用创建桌子 代码干掉了
			} else {
				for i := 0; i < tableNum; i++ {
					hft := NewHFT(floor.Rule.PlayerNum, i)
					hft.IsDefault = true

					floor.DataLock.Lock()
					floor.Tables[i] = hft
					floor.DataLock.Unlock()
				}
			}
			err := GetDBMgr().HouseFloorUpdate(floor)
			if err != nil {
				xlog.Logger().Errorf("update house floor error:%+v\n", err)
				continue
			}

		}
		// 开启混排后，包厢开起同步协程
		go clb.StartSync()
		go clb.StartSyncAiSuperNum()
	} else {
		// 如果包厢关闭了混排总开关
		for _, floor := range clb.Floors {
			// 把原先混排的楼层重置为普通的楼层
			if floor.IsMix {
				floor.DataLock.Lock()
				for k, table := range floor.Tables {
					if table.TId > 0 {
						xlog.Logger().Errorf("异常数据出现")
						floor.DataLock.Unlock()
						return xerrors.UserFloorGameError
					} else {
						delete(floor.Tables, k)
					}
				}
				floor.DataLock.Unlock()

				for i := 0; i < GetServer().ConHouse.TableNum; i++ {
					floor.DataLock.Lock()
					floor.Tables[i] = NewHFT(floor.Rule.PlayerNum, i)
					floor.DataLock.Unlock()
				}
				err := GetDBMgr().HouseFloorUpdate(floor)
				if err != nil {
					xlog.Logger().Errorf("update house floor error:%+v\n", err)
					continue
				}
				go floor.StartSync()
			}
			nowMix := false
			for _, fid := range fids {
				if floor.Id == fid {
					nowMix = true
				}
			}
			// 最后把客户端发来的的楼层混排开关变动同步
			if floor.IsMix != nowMix {
				floor.IsMix = nowMix
				err := GetDBMgr().HouseFloorUpdate(floor)
				if err != nil {
					xlog.Logger().Errorf("update house floor error:%+v\n", err)
					continue
				}
			}
		}
	}

	clb.DBClub.MixActive = isActive
	clb.DBClub.MixTableNum = tableNum
	// 自动加桌模式下是否往后面追加空桌子/否则就是往前追加空桌子
	//clb.DBClub.EmptyTableBack = isEmptyTblBack
	//clb.DBClub.TableSortType = tblSortType
	// 自动加桌模式下最大空桌子数
	clb.DBClub.EmptyTableMax = emptyTblMax
	clb.DBClub.NewTableSortType = newTblSortType
	clb.DBClub.CreateTableType = createTblType

	GetDBMgr().HouseUpdate(clb)
	tx := GetDBMgr().GetDBmControl().Begin()
	defer tx.Rollback()
	sql := `update house set mix_active = ?,mix_table_num = ? where id = ?`
	db := tx.Exec(sql, isActive, tableNum, clb.DBClub.Id)
	if db.Error != nil {
		xlog.Logger().Errorf("添加混合大厅失败：%+v\n", db.Error)
		return xerrors.DBExecError
	}
	err := GetDBMgr().GetDBmControl().Model(models.HouseFloor{}).
		Where("id not in (?) and  hid = ?", fids, clb.DBClub.Id).
		Update("is_mix", false).
		Error
	if err != nil {
		xlog.Logger().Errorf("更新混合大厅失败%+v\n", err)
		return xerrors.DBExecError
	}

	err = GetDBMgr().GetDBmControl().Model(models.HouseFloor{}).
		Where("id in (?) and  hid = ?", fids, clb.DBClub.Id).
		Update("is_mix", true).
		Error
	if err != nil {
		xlog.Logger().Errorf("更新混合大厅失败%+v\n", err)
		return xerrors.DBExecError
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Errorf("更新混合大厅失败%+v\n", err)
		return xerrors.DBExecError
	}
	return nil
}
func (clb *Club) AddTableLimitGroup() (int, *xerrors.XError) {
	groupID, err := clb.addTableLimitGroup()
	if err != nil {
		return 0, xerrors.DBExecError
	}
	clb.TableLimitGropuCount++
	if clb.LimitGroups == nil {
		clb.LimitGroups = make(map[int]map[int64]bool)
	}
	if clb.LimitGroupsUpdateAt == nil {
		clb.LimitGroupsUpdateAt = make(map[int]int64)
	}
	clb.LimitGroups[groupID] = make(map[int64]bool)
	clb.LimitGroupsUpdateAt[groupID] = time.Now().Unix()
	return groupID, xerrors.RespOk
}

type GroupDb struct {
	GroupId int `gorm:"group_id"`
}

func (clb *Club) addTableLimitGroup() (int, error) {
	clb.AddTableLimitLock.CustomLock()
	defer clb.AddTableLimitLock.CustomUnLock()
	var res GroupDb
	groupIDSql := `select group_id from house_table_limit_user where hid = ? order by group_id desc limit 1`
	err := GetDBMgr().GetDBmControl().Raw(groupIDSql, clb.DBClub.Id).Scan(&res).Error
	sql := `insert into house_table_limit_user(hid,group_id) values 
	(?,?)`
	err = GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, res.GroupId+1).Error
	if err != nil {
		return 0, err
	}
	return res.GroupId + 1, nil
}
func (clb *Club) RemoveTableLimitGroup(id int) error {
	clb.AddTableLimitLock.CustomLock()
	defer clb.AddTableLimitLock.CustomUnLock()
	sql := `update house_table_limit_user set status = 1 where hid= ? and group_id = ?`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, id).Error
	if err != nil {
		return err
	}
	delete(clb.LimitGroups, id)
	delete(clb.LimitGroupsUpdateAt, id)
	clb.TableLimitGropuCount--
	return nil
}

func (clb *Club) AddTableLimitUser(groupID int, uid int64) *xerrors.XError {
	clb.TableLimitUserLock.CustomLock()
	defer clb.TableLimitUserLock.CustomUnLock()
	userMap := clb.LimitGroups[groupID]
	if userMap != nil {
		if len(userMap) >= 200 {
			return xerrors.HouseTableLimitUserMax
		}
	} else {
		return xerrors.HouseTableLimitUserMax
	}
	sql := `insert into house_table_limit_user(hid,group_id,uid) values(?,?,?) ON DUPLICATE KEY UPDATE status = 0`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, groupID, uid).Error
	if err != nil {
		return xerrors.DBExecError
	}
	userMap[uid] = true
	clb.LimitGroupsUpdateAt[groupID] = time.Now().Unix()

	return xerrors.RespOk
}

func (clb *Club) RemoveTableLimitUser(groupID int, uid int64) error {
	clb.TableLimitUserLock.CustomLock()
	defer clb.TableLimitUserLock.CustomUnLock()
	sql := `update house_table_limit_user set status = 1 where hid = ? and group_id = ? and uid = ?`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, groupID, uid).Error
	if err != nil {
		xlog.Logger().Errorf("db error：%v", err)
		return err
	}
	userMap := clb.LimitGroups[groupID]
	clb.LimitGroupsUpdateAt[groupID] = time.Now().Unix()
	if userMap != nil {
		delete(clb.LimitGroups[groupID], uid)
	} else {
		return errors.New("error group id")
	}
	return nil
}

// CheckUsersAllowSameTable 检查用户是否被禁止同桌规则限制，用户id必须作为第一个参数传入
func (clb *Club) CheckUsersAllowSameTable(uids ...int64) (error, []int64) {
	if len(uids) <= 1 {
		return nil, nil
	}
	var err error
	dest := []int64{}
	for _, v := range clb.LimitGroups {
		if _, ok := v[uids[0]]; !ok {
			continue
		}
		for _, uid := range uids[1:] {
			if _, ok := v[uid]; ok {
				dest = append(dest, uid)
				err = fmt.Errorf("house table limit: A：%d  B：%d", uids[0], uid)
			}
		}
	}
	return err, RemoveDuplicates(dest)
}
func RemoveDuplicates(a []int64) (ret []int64) {
	destMap := make(map[int64]struct{}, len(a))
	for _, u := range a {
		if _, ok := destMap[u]; ok {
			continue
		}
		destMap[u] = struct{}{}
	}
	for k, _ := range destMap {
		ret = append(ret, k)
	}
	return
}

func (clb *Club) CheckUserIsLimitInGroup(groupID int, uid int64) bool {
	limGroup := clb.LimitGroups[groupID]
	if limGroup == nil {
		return false
	}
	v, ok := limGroup[uid]
	return ok && v

}

func (clb *Club) GetTableLimitInfo() *static.Msg_HC_HouseTableLimitInfo {
	dest := static.Msg_HC_HouseTableLimitInfo{}
	dest.Is2PNotEffect = clb.DBClub.IsNotEft2PTale
	for k, v := range clb.LimitGroups {
		groupInfo := static.GroupInfo{}
		groupInfo.GroupID = k
		groupInfo.CreateAt = time.Unix(clb.LimitGroupsUpdateAt[k], 0).Format(consts.TIME_Y_M_D)
		for i, _ := range v {
			hmem := clb.GetMemByUId(i)
			if hmem == nil {
				continue
			}
			var titem static.LimitUserInfo
			titem.UId = i
			titem.UName = hmem.NickName
			titem.UUrl = hmem.ImgUrl
			titem.UGender = hmem.Sex
			titem.Limit = true
			groupInfo.Users = append(groupInfo.Users, &titem)
		}
		sort.Sort(static.LimitUserSlie(groupInfo.Users))
		groupInfo.UserCount = len(groupInfo.Users)
		dest.Groups = append(dest.Groups, &groupInfo)
	}
	sort.Sort(static.LimitGroups(dest.Groups))
	dest.TotalGroup = len(dest.Groups)
	return &dest
}

func (clb *Club) MemVitaminClear() *xerrors.XError {

	hmem := clb.GetMemByUId(clb.DBClub.UId)
	if hmem == nil {
		return xerrors.InValidHouseMemberError
	}
	var memTotal int64
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.URole == consts.ROLE_CREATER {
			continue
		}
		memTotal += mem.UVitamin
	}
	if memTotal < 0 && memTotal+hmem.UVitamin < 0 {
		return xerrors.VitaminNotEnoughToZero
	}
	defer func() {
		ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
		ntf.HId = clb.DBClub.HId
		ntf.OptId = hmem.UId
		hp := GetPlayerMgr().GetPlayer(hmem.UId)
		if hp != nil {
			hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
		}
	}()
	cli := GetDBMgr().Redis
	var changeTotal int64
	for _, mem := range mems {
		if mem.UVitamin == 0 {
			continue
		}
		if mem.URole == consts.ROLE_CREATER {
			continue
		}
		mem.Lock(cli)
		oldV := mem.UVitamin
		_, after, err := mem.VitaminChangeNoNtf(clb.DBClub.UId, -mem.UVitamin, models.AdminSend, nil)
		mem.Unlock(cli)

		if err != nil {
			xlog.Logger().Errorln(consts.MsgTypeHouseVitaminSet, "err: ", err.Error())
			return xerrors.DBExecError
		}
		changeTotal += oldV
		//修改疲劳值统计管理节点信息
		GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, mem.UId, after, nil)
	}
	_, after, err := hmem.VitaminChangeNoNtf(hmem.UId, changeTotal, models.AdminSend, nil)
	if err != nil {
		xlog.Logger().Errorln(consts.MsgTypeHouseVitaminSet, "err:,should change vitamin:%d ", err.Error(), changeTotal)
		return xerrors.DBExecError
	}
	GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, hmem.UId, after, nil)

	return nil
}

func (clb *Club) PartnerVitaminClear(uid int64) *xerrors.XError {
	if !clb.DBClub.IsPartnerModi {
		return xerrors.InvalidPermission
	}
	hmem := clb.GetMemByUId(uid)
	if hmem == nil {
		return xerrors.InValidHouseMemberError
	}
	if hmem.Partner != 1 {
		return xerrors.InValidHousePartnerError
	}
	var memTotal int64
	mems := clb.GetMemByPartner(nil, uid)
	for _, mem := range mems {
		if mem.UId == uid { //排除自己
			continue
		}
		memTotal += mem.UVitamin
	}
	if memTotal < 0 && memTotal+hmem.UVitamin < 0 {
		return xerrors.VitaminNotEnoughToZero
	}
	defer func() {
		ntf := new(static.Msg_HC_HouseVitaminSet_Ntf)
		ntf.HId = clb.DBClub.HId
		ntf.OptId = hmem.UId
		hp := GetPlayerMgr().GetPlayer(hmem.UId)
		if hp != nil {
			hp.SendMsg(consts.MsgTypeHouseVitaminSet_Ntf, ntf)
		}
	}()
	cli := GetDBMgr().Redis
	var changeTotal int64
	for _, mem := range mems {
		mem.Lock(cli)
		oldV := mem.UVitamin
		_, after, err := mem.VitaminChangeNoNtf(uid, -mem.UVitamin, models.PartnerSend, nil)
		mem.Unlock(cli)

		if err != nil {
			xlog.Logger().Errorln(consts.MsgTypeHouseVitaminSet, "err: ", err.Error())
			return xerrors.DBExecError
		}
		changeTotal += oldV
		//修改疲劳值统计管理节点信息
		GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, mem.UId, after, nil)

	}
	_, afterHm, err := hmem.VitaminChangeNoNtf(uid, changeTotal, models.PartnerSend, nil)

	if err != nil {
		xlog.Logger().Errorln(consts.MsgTypeHouseVitaminSet, "err:,should change vitamin:%d ", err.Error(), changeTotal)
		return xerrors.DBExecError
	}
	GetDBMgr().UpdateVitaminMgrList(clb.DBClub.Id, hmem.UId, afterHm, nil)

	return nil
}

// 是否为可维护的游戏玩法
func (clb *Club) IsMaintainableGame(kindId, version int) *xerrors.XError {
	pkgList, err := clb.GetMaintainablePkg()
	if err != nil {
		return err
	}
	for _, pkg := range pkgList {
		if pkg == nil {
			continue
		}
		for _, game := range pkg.Games {
			if game == nil {
				continue
			}
			if game.KindId == kindId {
				// if game.ForcedVersion != version {
				// 	return xerrors.HouseFloorRuleChangeStrongError
				// }
				return nil
			}
		}
	}
	return xerrors.NewXError("您无法调整该包厢玩法权限")
}

// 得到包厢支持的玩法
func (clb *Club) GetMaintainablePkg() (static.AreaPkgCompiledList, *xerrors.XError) {
	result := make(static.AreaPkgCompiledList, 0)
	userConfigKindIds := make([]*models.UserAgentConfigKindId, 0)
	err := GetDBMgr().GetDBmControl().Find(&userConfigKindIds, "uid = ?", clb.DBClub.UId).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return nil, xerrors.DBExecError
	}
	// 得到包厢本来就存在的玩法
	_, kindIds := clb.GetFloors()
	xlog.Logger().Debugln("包厢楼层已存在的kindid玩法:", kindIds)
	for _, kindIdConfig := range userConfigKindIds {
		kindIds = append(kindIds, kindIdConfig.KindId)
	}
	xlog.Logger().Debugln("包厢楼层所有kindid玩法:", kindIds)
	pkgKeys := make(map[string][]int)
	for _, kindId := range kindIds {
		pkey := GetAreaPackageKeyByKid(kindId)
		pkgKeys[pkey] = append(pkgKeys[pkey], kindId)
	}
	for pkey, kids := range pkgKeys {
		pkg := GetAreaPackageByPKey(pkey)
		if pkg == nil {
			continue
		}
		pkg.ScreenGame(kids...)
		result = append(result, pkg)
	}
	// 获取限免游戏
	freeGameMap := GetServer().GetLimitFreeGameKindIds()
	for _, pkg := range result {
		for _, game := range pkg.Games {
			if free, ok := freeGameMap[game.KindId]; ok {
				game.TimeLimitFree = free
			} else {
				game.TimeLimitFree = false
			}
		}
	}
	return result, nil
}

// 获得在线的空闲的成员
func (clb *Club) GetMemLeisure() []HouseMember {
	mems := clb.GetAllMem()
	dest := make([]HouseMember, len(mems)/5)
	for _, m := range mems {
		p := GetPlayerMgr().GetPlayer(m.UId)
		if p == nil {
			continue
		}
		if p.Info.TableId > 0 || p.Info.GameId > 0 {
			continue
		}
		dest = append(dest, m)

	}
	return dest
}

// 广播消息
func (clb *Club) BroadcastMsg(head string, v interface{}, filter func(*HouseMember) bool /*过滤器*/) {
	mems := clb.GetMemSimple(false)
	for _, m := range mems {
		// 没有给定过滤器 则广播所有成员
		if filter != nil {
			if !filter(&m) {
				continue
			}
		}
		m.SendMsg(head, v)
	}
}

func (clb *Club) RangeMem(do func(mem *HouseMember)) {
	mems := clb.GetAllMem()
	for _, m := range mems {
		do(&m)
	}
}

type HUserSort struct {
	Uid int64
	Id  int64
}
type HUserSortSlice []*HUserSort

func (p HUserSortSlice) Len() int           { return len(p) }
func (p HUserSortSlice) Less(i, j int) bool { return p[i].Id < p[j].Id }
func (p HUserSortSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (clb *Club) GetMemberUidOrderByID(role int) []int64 {
	if role == consts.ROLE_CREATER {
		return []int64{clb.DBClub.UId}
	}
	var datas []*HUserSort
	mems := clb.GetMemSimple(true)
	for _, m := range mems {
		if m.URole == role {
			datas = append(datas, &HUserSort{Uid: m.UId, Id: m.Id})
		}
	}
	sort.Sort(HUserSortSlice(datas))
	dest := []int64{}
	for _, item := range datas {
		dest = append(dest, item.Uid)
	}
	return dest
}

func (clb *Club) GetMemberOrderByID(role int) []HouseMember {
	if role == consts.ROLE_CREATER {
		return []HouseMember{*clb.GetMemByUId(clb.DBClub.UId)}
	}
	var datas []HouseMember
	mems := clb.GetMemSimple(true)
	for _, m := range mems {
		if m.URole == role {
			datas = append(datas, m)
		}
	}
	sort.Sort(HouseMemberItemWrapper{datas, func(item1, item2 *HouseMember) bool {
		if item1.Id > item2.Id {
			return true
		}
		return false
	}})

	return datas
}

func (clb *Club) GetMemNoPartnerMemOrderById() []HouseMember {
	var datas []HouseMember
	mems := clb.GetMemSimple(true)
	for _, m := range mems {
		if m.URole != consts.ROLE_MEMBER {
			continue
		}
		if m.Partner != 0 {
			continue
		}
		if m.IsVitaminAdmin() {
			continue
		}
		if m.IsVicePartner() {
			continue
		}
		datas = append(datas, m)
	}
	sort.Sort(HouseMemberItemWrapper{datas, func(item1, item2 *HouseMember) bool {
		return item1.Id > item2.Id
	}})
	return datas
}

func (clb *Club) GetMemByPartnerOrderById(partnerId int64) []HouseMember {
	var datas []HouseMember
	mems := clb.GetMemSimple(true)
	for _, m := range mems {
		if m.URole != consts.ROLE_MEMBER {
			continue
		}
		if m.Partner == partnerId {
			datas = append(datas, m)
		}
	}
	sort.Sort(HouseMemberItemWrapper{datas, func(item1, item2 *HouseMember) bool {
		return item1.Id > item2.Id
	}})
	return datas
}

func (clb *Club) GetMemForAllPartner() []HouseMember {
	var datas []HouseMember
	mems := clb.GetMemSimple(true)
	for _, m := range mems {
		if m.URole != consts.ROLE_MEMBER {
			continue
		}
		if m.Partner == 1 {
			datas = append(datas, m)
		}
	}
	sort.Sort(HouseMemberItemWrapper{datas, func(item1, item2 *HouseMember) bool {
		return item1.Id > item2.Id
	}})
	return datas
}

func (clb *Club) GetAllPartnerWithoutJunior(sid int64) []HouseMember {
	var datas []HouseMember
	mems := clb.GetMemSimple(true)
	for _, m := range mems {
		if m.URole != consts.ROLE_MEMBER {
			continue
		}
		if m.Superior > 0 && m.UId != sid {
			continue
		}
		if m.IsVitaminAdmin() {
			continue
		}
		if m.Partner == 1 {
			datas = append(datas, m)
		}
	}
	sort.Sort(HouseMemberItemWrapper{datas, func(item1, item2 *HouseMember) bool {
		if item1.UId == sid {
			return true
		}
		if item2.UId == sid {
			return false
		}
		return item1.Id > item2.Id
	}})
	return datas
}

type HouseMsgDB struct {
	Msg         string `json:"msg" gorm:"msg"`
	CreateStamp int64  `json:"create_stamp"  gorm:"create_stamp"`
}

func (clb *Club) GetMsg(start, end, flag int) ([]*HouseMsgDB, error) {
	dest := []*HouseMsgDB{}
	sql := ""
	if flag == 1 {
		sql = `select msg,create_stamp from house_msg where hid = ? and create_stamp > ? and (msg_type > 0 and msg_type != 15 and msg_type != 16 and msg_type != 17) order by id desc limit ? offset ?`
	} else if flag == 2 {
		// msg_type = JoinHouse
		sql = `select msg,create_stamp from house_msg where hid = ? and create_stamp > ? and msg_type = 16 order by id desc limit ? offset ?`
	} else if flag == 3 {
		// msg_type = KickHouseMem、ExitHouse
		sql = `select msg,create_stamp from house_msg where hid = ? and create_stamp > ? and (msg_type = 15 or msg_type = 17) order by id desc limit ? offset ?`
	} else {
		sql = `select msg,create_stamp from house_msg where hid = ? and create_stamp > ? and msg_type > 0 order by id desc limit ? offset ?`
	}
	err := GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, time.Now().Unix()-7*24*3600, end-start, start).Scan(&dest).Error
	return dest, err
}

func (clb *Club) LimitUserGame(optID, uid int64, allow bool) error {
	sql := `insert into house_user_limit(dhid,operator,uid,status) values(?,?,?,?) ON 
	DUPLICATE KEY UPDATE  status = ?,updated_at= now() `
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, optID, uid, allow, allow).Error
	if err != nil {
		return err
	}
	if allow {
		err = GetDBMgr().Redis.SRem(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), uid).Err()
	} else {
		err = GetDBMgr().Redis.SAdd(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), uid).Err()
	}
	return err
}

func (clb *Club) LimitUsersGame(optID int64, allow bool, ids ...int64) error {
	l := len(ids)
	if l > 0 {
		values := make([]string, l)
		arr := make([]interface{}, l)
		for i := 0; i < l; i++ {
			arr[i] = ids[i]
			values[i] = fmt.Sprintf("(%d,%d,%d,%t)", clb.DBClub.Id, optID, ids[i], allow)
		}
		sql := fmt.Sprintf(`insert into house_user_limit(dhid,operator,uid,status) values%s ON 
				DUPLICATE KEY UPDATE status = ?,updated_at= now();`, strings.Join(values, ","))
		err := GetDBMgr().GetDBmControl().Exec(sql, allow).Error
		if err != nil {
			return err
		}
		if allow {
			err = GetDBMgr().Redis.SRem(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), arr...).Err()
		} else {
			err = GetDBMgr().Redis.SAdd(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id), arr...).Err()
		}
	}
	return nil
}

func (clb *Club) RedisPub(head string, msg interface{}) error {
	// cli1 := GetDBMgr().Redis
	cli2 := GetDBMgr().PubRedis
	buf, err := json.Marshal(msg)
	if err != nil {
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
	for _, f := range clb.Floors {
		// cli1.Publish(f.PubKey(), buf).Err()
		cli2.Publish(f.PubKey(), buf).Err()
	}
	return nil
}

func (clb *Club) UserJoin(uid, opID int64) *xerrors.XError {
	if clb.IsBusyMerging() {
		return xerrors.HouseBusyError
	}
	// 人数上限
	if clb.GetMemCounts() >= GetServer().ConHouse.MemMax {
		return xerrors.HouseMemJoinMaxError
	}

	// 玩家加入上限
	count, err := GetDBMgr().GetDBrControl().HouseMemberJoinCounts(uid)
	if err != nil {
		return xerrors.DBExecError
	}
	if count >= GetServer().ConHouse.JoinMax {
		return xerrors.MemJoinHouseMaxError
	}
	op, _ := GetDBMgr().GetDBrControl().GetPerson(opID)
	if op == nil {
		return xerrors.InvalidIdError
	}
	u, _ := GetDBMgr().GetDBrControl().GetPerson(uid)
	if u == nil {
		return xerrors.InvalidIdError
	}

	// 成员过审
	custerr := clb.MemAgree(opID, uid, nil)
	if custerr != nil {
		return custerr
	}

	msg := fmt.Sprintf("<color=#00A70C>%sID:%d</color>加入包厢", u.Nickname, u.Uid)
	CreateClubMassage(clb.DBClub.Id, opID, JoinHouse, msg)
	return xerrors.RespOk
}

// PoolChange 包厢疲劳值变更记录
func (clb *Club) PoolChange(optUid int64, optType models.VitaminChangeType, value int64, tx *gorm.DB) *xerrors.XError {
	if value < 0 && clb.GetHouseVitaminPool(nil)+value < 0 {
		return xerrors.VitaminPutNotEnoughError
	}
	sql := `select after from house_vitamin_pool where hid = ? and after >= ? for update `
	if tx == nil {
		db := GetDBMgr().GetDBmControl().Begin()
		if value < 0 {
			if db.Exec(sql, clb.DBClub.Id, value).RowsAffected != 1 {
				return xerrors.VitaminPutNotEnoughError
			}
		}
		err := models.HousePoolChange(clb.DBClub.Id, optUid, optType, value, "", db)
		if err != nil {
			db.Rollback()
			xlog.Logger().Errorf("db exec error:%v", err)
			return xerrors.DBExecError
		}
		err = db.Commit().Error
		if err != nil {
			db.Rollback()
			xlog.Logger().Errorf("db exec error:%v", err)
			return xerrors.DBExecError
		}
		return nil
	}
	err := models.HousePoolChange(clb.DBClub.Id, optUid, optType, value, "", tx)
	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError
	}
	return nil
}

type AfterGorm struct {
	After int64 `gorm:"after"`
}

func (clb *Club) GetHouseVitaminPool(userAgentConfig *models.UserAgentConfig) int64 {
	// sql := `select after from house_vitamin_pool where hid = ? `
	// db := GetDBMgr().GetDBmControl()
	// after := AfterGorm{}
	// db.Raw(sql, clb.DBClub.Id).Scan(&after)
	mems := clb.GetMemSimple(false)
	var vitamin int64
	for _, mem := range mems {
		vitamin += mem.UVitamin
	}
	if userAgentConfig != nil {
		return static.SwitchVitaminInt64(userAgentConfig.VitaminPoolMax) - vitamin
	}
	return clb.GetVitaminMax() - vitamin
}

func (clb *Club) GetLogCount() int {
	sql := `select count(*) as after from house_vitamin_pool_log where hid = ? and optype != ? and optype != ? `
	db := GetDBMgr().GetDBmControl()
	var after int
	db.Raw(sql, clb.DBClub.Id, models.AdminSend, models.PoolAdd).Count(&after)
	return after
}

// func (h *Club) BuildHouseTaxSum(t time.Time) {
// 	if t.Unix() > time.Now().Unix() {
// 		return
// 	}
// 	start, end := public.GetTimeLastDayStartAndEnd(t)
// 	tx := GetDBMgr().GetDBmControl().Begin()
// 	var gn []int64
// 	err := tx.Table("house_vitamin_pool_tax_log").Select("SUM(value)").
// 		Where("hid = ? and status = 0 and created_at > ? and created_at <= ?", h.DBClub.Id, start, end).Pluck("SUM(value)", &gn).Error

// 	if err != nil {
// 		if !strings.Contains(err.Error(), "Scan error") {
// 			syslog.Logger().Errorf("query error:%v", err)
// 			return
// 		}
// 	}
// 	var sum int64
// 	if len(gn) == 1 {
// 		sum = gn[0]
// 	}
// 	if sum == 0 {
// 		return
// 	}
// 	tx.Exec(`update  house_vitamin_pool_tax_log set status = 1 where hid = ? and status = 0 and created_at > ? and created_at <= ?`, h.DBClub.Id, start, end)
// 	sql := `insert into house_vitamin_pool_log(hid,opuid,value,optype,created_at,extra) values(?,?,?,?,?,?) `
// 	// err = model.HousePoolChange(h.DBClub.Id, h.DBClub.UId, model.TaxSum, sum, tx)
// 	// if err != nil {
// 	// 	syslog.Logger().Errorf("insert error:%v", err)
// 	// 	return
// 	// }
// 	err = tx.Exec(sql, h.DBClub.Id, h.DBClub.UId, sum, model.TaxSum, end, end.Format("20060102150505")).Error
// 	if err != nil {
// 		syslog.Logger().Errorf("insert error:%v", err)
// 		return
// 	}
// 	tx.Commit()
// 	return
// }

func (clb *Club) GetVitaminPoolLog(start, count int64) (*static.MSgHouseVitaminPoolRecord, *xerrors.XError) {
	db := GetDBMgr().GetDBmControl()

	// 需要实时入账
	var gn []int64
	if start == 0 {
		err := db.Table("house_vitamin_pool_tax_log").Select("SUM(value)").
			Where("hid = ? and status = 0 and optype in (?,?)", clb.DBClub.Id, models.BigWinCost, models.GamePay).Pluck("SUM(value)", &gn).Error
		if err != nil {
			if !strings.Contains(err.Error(), "Scan error") {
				xlog.Logger().Errorf("query error:%v", err)
				return nil, xerrors.DBExecError
			}
		}
	}

	records := make([]*models.HouseVitaminPoolLog, 0, count)
	sql := `select id,hid,opuid,value,after,optype,created_at,extra  
	from house_vitamin_pool_log where hid = ?  and optype != ? and optype != ?  and optype != ?  order by created_at desc limit ?  offset ?  `
	db.Raw(sql, clb.DBClub.Id, models.AdminSend, models.PoolAdd, models.ViAdminSet, count, start).Scan(&records)

	ack := new(static.MSgHouseVitaminPoolRecord)
	ack.Items = make([]*static.HouseVitaminRecord, 0, count)
	ack.TotalCount = clb.GetLogCount()
	if start == 0 && len(gn) == 1 {
		item := new(static.HouseVitaminRecord)
		item.UpdatedTime = time.Now().Unix()
		item.OptType = models.GetTypeName(gn[0], models.TaxSum)
		item.ChangeVitamin = static.SwitchVitaminToF64(gn[0])
		item.OptTypeInt = int64(models.TaxSum)
		item.OptUName = "系统"
		ack.Items = append(ack.Items, item)
	}
	for _, record := range records {
		// if record.Opuid == 0 {
		// 	record.Opuid = clb.DBClub.UId
		// }
		item := new(static.HouseVitaminRecord)
		item.Id = record.Id
		item.UpdatedTime = record.CreatedAt.Unix()
		item.OptType = models.GetTypeName(record.Value, record.Optype)
		item.AftVitamin = static.SwitchVitaminToF64(record.After)
		item.ChangeVitamin = static.SwitchVitaminToF64(record.Value)
		item.OptTypeInt = int64(record.Optype)
		if record.Optype == models.GameTotal ||
			record.Optype == models.HouseCreate || record.Optype == models.TaxSum {
			item.OptUName = "系统"
		} else {
			if record.Opuid == 0 {
				item.OptUName = "系统"
			} else {
				item.OptUName = fmt.Sprintf("ID:%d", record.Opuid)
			}
		}
		ack.Items = append(ack.Items, item)
	}
	return ack, nil
}

type AdminSendSum struct {
	Send  int64
	Opuid int64
}

func (clb *Club) BuildHouseSendSum(t time.Time) {
	if t.Unix() > time.Now().Unix() {
		return
	}
	start, end := static.GetTimeLastDayStartAndEnd(t)
	tx := GetDBMgr().GetDBmControl().Begin()

	sql := `select Sum(value) as send,opuid from house_vitamin_pool_log where hid = ? and opuid !=0 and created_at > ? and created_at <=? and status = 0 
	and (optype = ? or optype = ?) group by opuid `
	dest := []AdminSendSum{}
	tx.Raw(sql, clb.DBClub.Id, start, end, models.AdminSend, models.PoolAdd).Scan(&dest)
	if len(dest) == 0 {
		tx.Rollback()
		return
	}
	sql = `insert into house_vitamin_pool_log(hid,opuid,value,optype,created_at) values(?,?,?,?,?)`
	for _, item := range dest {
		if err := tx.Exec(sql, clb.DBClub.Id, item.Opuid, item.Send, models.AdminSendSum, end).Error; err != nil {
			tx.Rollback()
			xlog.Logger().Errorf("insert admin send sum log  error:%v", err)
			return
		}
	}
	err := tx.Exec(`update house_vitamin_pool_log set status = 1 where hid = ? and opuid !=0 and created_at > ? and created_at <=? and status = 0 
	and (optype = ?  or optype = ?)  `, clb.DBClub.Id, start, end, models.AdminSend, models.PoolAdd).Error
	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorf("update admin send sum log  error:%v", err)
		return
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorf("commit admin send sum log  error:%v", err)
		return
	}
	return
}

func (clb *Club) CheckVitaminPermission(mem *HouseMember) bool {
	if mem.URole == consts.ROLE_CREATER {
		return true
	}

	if mem.IsVitaminAdmin() {
		return true
	}

	if mem.URole == consts.ROLE_ADMIN {
		return clb.DBClub.IsVitaminHide
	}
	if mem.VitaminAdmin {
		return true
	}
	if mem.IsPartner() || mem.IsVicePartner() {
		return clb.DBClub.IsPartnerHide
	}

	return false
}

func (clb *Club) CalculateLeftVitamin() (int64, int64) {
	VitaminLeft := int64(0)
	VitaminMinus := int64(0)
	housemembers, _ := GetDBMgr().GetDBrControl().GetHouseMemberWithUids(clb.DBClub.Id, clb.GetMemsUid())
	for _, hpmem := range housemembers {
		if hpmem.UVitamin < 0 {
			VitaminMinus += hpmem.UVitamin
		} else {
			VitaminLeft += hpmem.UVitamin
		}
	}
	return VitaminLeft, VitaminMinus
}

func (clb *Club) StatisticHouseMemVitaminLeftDay() {
	//timeNow := time.Now()
	//timeNow = timeNow.AddDate(0, 0, -1)
	//selectstr := fmt.Sprintf("%d-%02d-%02d", timeNow.Year(), timeNow.Month(), timeNow.Day())
	//go GetDBMgr().InsertHouseMemberDayLeft(int64(clb.HId), clb.GetAllMem(), selectstr)
}

// 检查包厢是否有正在游戏中的桌子
func (clb *Club) CheckInGameTable() bool {
	clb.FloorLock.Lock()
	defer clb.FloorLock.Unlock()
	for _, f := range clb.Floors {
		if f == nil {
			continue
		}
		f.DataLock.Lock()
		for _, t := range f.Tables {
			if t == nil {
				continue
			}
			if t.TId > 0 {
				if len(t.UserWithOnline) > 0 {
					f.DataLock.Unlock()
					return true
				}
			}
		}
		f.DataLock.Unlock()
	}
	return false
}

// 检查包厢是否有正在游戏中的桌子
func (clb *Club) CheckSpecifiedUsersInGameTable(userIds ...int64) bool {
	if len(userIds) <= 0 {
		return false
	}
	in := func(tableUsers ...FTUsers) bool {
		for _, user := range tableUsers {
			for _, useId := range userIds {
				if user.Uid == useId {
					return true
				}
			}
		}
		return false
	}
	clb.FloorLock.Lock()
	defer clb.FloorLock.Unlock()
	for _, f := range clb.Floors {
		if f == nil {
			continue
		}
		f.DataLock.Lock()
		for _, t := range f.Tables {
			if t == nil {
				continue
			}
			if t.TId > 0 {
				if len(t.UserWithOnline) > 0 && in(t.UserWithOnline...) {
					f.DataLock.Unlock()
					return true
				}
			}
		}
		f.DataLock.Unlock()
	}
	return false
}

// 是否配置了扣除设置
func (clb *Club) ConfiguredDeduct() bool {
	//clb.FloorLock.RLockWithLog()
	//defer clb.FloorLock.RUnlock()
	//
	//for _, floor := range clb.Floors {
	//	if floor == nil {
	//		continue
	//	}
	//	if floor.ConfiguredDeduct() {
	//		return true
	//	}
	//}
	floorIds, _ := clb.GetFloors()

	paySlice := make(models.HouseFloorGearPaySlice, 0)
	err := GetDBMgr().GetDBmControl().Where("hid = ? and id in(?)", clb.DBClub.Id, floorIds).Find(&paySlice).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Error(err)
		}
		return false
	}

	for i := 0; i < len(paySlice); i++ {
		if paySlice[i].Configured() {
			return true
		}
	}
	return false
}

// 是否配置了生效设置
func (clb *Club) ConfiguredEffect() bool {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()
	for _, floor := range clb.Floors {
		if floor == nil {
			continue
		}
		if floor.ConfiguredEffect() {
			return true
		}
	}
	return false
}

func (clb *Club) GetFNoByFid(fid int64) string {
	clb.FloorLock.RLockWithLog()
	var arr []int64
	for _, f := range clb.Floors {
		arr = append(arr, f.Id)
	}
	sort.Sort(util.Int64Slice(arr))
	clb.FloorLock.RUnlock()

	for i, a := range arr {
		if a == fid {
			return fmt.Sprintf("%d楼", i+1)
		}
	}

	return ""
}

func (clb *Club) GetParnterProfit(superiorprofit []int, royalty []int, floorIndex int) (int, int) {
	f_superiorprofit := 0
	if floorIndex < len(superiorprofit) {
		f_superiorprofit = superiorprofit[floorIndex]
	}
	if f_superiorprofit < 0 {
		f_superiorprofit = 0
	}
	f_royalty := 0
	if floorIndex < len(royalty) {
		f_royalty = royalty[floorIndex]
	}
	if f_royalty < 0 {
		f_royalty = 0
	}
	return f_superiorprofit, f_royalty
}

// 目前只刷新包厢和成员 不刷新楼层 后期根据需求可加上
func (clb *Club) ReloadHouse() {
	house := new(models.House)
	// mysql
	if err := GetDBMgr().GetDBmControl().Where("id = ?", clb.DBClub.Id).Find(house).Error; err != nil {
		xlog.Logger().Errorf("sync house %d for db error:%v", clb.DBClub.Id, err)
		return
	}
	// redis
	if err := GetDBMgr().GetDBrControl().HouseInsert(house); err != nil {
		xlog.Logger().Errorf("sync house %d for redis error:%v", clb.DBClub.Id, err)
		return
	}
	clb.DBClub = house
}

func (clb *Club) StartSync() { //混排楼层数据同步
	var syncCount int
	var sleepSec time.Duration = time.Second
	var num int
	for clb.IsAlive {
		syncCount++
		if syncCount > 5 {
			sleepSec = clb.GetSyncDuring()
			syncCount = 0
		}
		time.Sleep(sleepSec)
		if num > 5 {
			num = 0
			if time.Now().Unix()-clb.LastSyncSorted < 5 {
				xlog.Logger().Warnf("StartSync.SyncTablesWithSorted")
				clb.SyncTablesWithSorted()
				continue
			}
		}
		do, err := clb.Sync()
		if err != nil {
			xlog.Logger().Errorf("sync error:%v", err)
			break
		}
		if do {
			num++
		}
	}
	xlog.Logger().Warnf("sync stop!")
}

func (clb *Club) StartSyncAiSuperNum() {
	for {
		if !clb.IsAlive {
			return
		}

		if !clb.DBClub.MixActive {
			return
		}

		time.Sleep(time.Second * 1)

		if !clb.IsAiSuper() {
			continue
		}

		clb.FloorLock.RLock()
		for _, f := range clb.Floors {
			if f == nil {
				continue
			}
			if f.IsMix && f.AiSuperNum > 0 {
				f.PubRedisMappingNumUpdate()
			}
		}
		clb.FloorLock.RUnlock()
	}
}

func (clb *Club) SyncTablesWithSorted() {
	xlog.Logger().Warnf("SyncTablesWithSorted")
	if !clb.IsAlive {
		return
	}

	if !clb.DBClub.MixActive {
		return
	}

	mixFloor := make([]*HouseFloor, 0, len(clb.Floors))
	memActive := make([]*HouseMember, 0, 64)
	clb.FloorLock.RLock()
	for _, f := range clb.Floors {
		if f.IsMix {
			mixFloor = append(mixFloor, f)
			f.MemLock.RLock()
			for _, mem := range f.MemAct {
				memActive = append(memActive, mem)
			}
			f.MemLock.RUnlock()
		}
	}
	clb.FloorLock.RUnlock()
	if len(mixFloor) == 0 || len(memActive) == 0 {
		return
	}
	var acks static.MixFloorAcks
	fids := make([]int64, 0, len(mixFloor))
	for _, f := range mixFloor {
		ack := f.GetTableInfo()
		if ack != nil {
			acks.Acks = append(acks.Acks, ack)
		}
		fids = append(fids, f.Id)
	}
	acks.FIDS = fids
	if len(acks.Acks) > 0 {
		clb.LastSyncSorted = time.Now().Unix()
		buf := static.HF_EncodeMsg(consts.MsgHouseFloorTableSort, xerrors.SuccessCode, &acks, GetServer().Con.Encode, GetServer().Con.EncodeClientKey, 0)
		for _, mem := range memActive {
			if p := GetPlayerMgr().GetPlayer(mem.UId); p != nil {
				p.SendBuf(buf)
			}
		}
	}
}

func (clb *Club) Sync() (bool, error) {
	if !clb.IsAlive {
		return false, errors.New("floor close")
	}

	if !clb.DBClub.MixActive {
		return false, errors.New("mix close")
	}
	mixFloor := make([]*HouseFloor, 0, len(clb.Floors))
	memActive := make([]*HouseMember, 0, 64)
	clb.FloorLock.RLock()
	for _, f := range clb.Floors {
		if f.IsMix {
			mixFloor = append(mixFloor, f)
			f.MemLock.RLock()
			for _, mem := range f.MemAct {
				memActive = append(memActive, mem)
			}
			f.MemLock.RUnlock()
		}
	}
	clb.FloorLock.RUnlock()
	if len(mixFloor) == 0 || len(memActive) == 0 {
		return false, nil
	}
	var acks static.MixFloorAcks
	fids := make([]int64, 0, len(mixFloor))
	for _, f := range mixFloor {
		ack := f.GetDiffInfo()
		if ack != nil {
			acks.Acks = append(acks.Acks, ack)
		}
		fids = append(fids, f.Id)
	}
	acks.FIDS = fids
	if len(acks.Acks) > 0 {
		for _, mem := range memActive {
			mem.SendMsg(consts.MsgHouseFloorTableSync, acks)
		}
		clb.RedisPub(consts.MsgHouseFloorTableSync, acks)
		return true, nil
	}
	return false, nil
}
func (clb *Club) GetRecommendUpdateFloors() []int64 {
	clb.FloorLock.RLockWithLog()
	defer clb.FloorLock.RUnlock()
	floors := make([]int64, 0)
	for _, floor := range clb.Floors {
		if floor == nil {
			continue
		}
		areaGame := GetAreaGameByKid(floor.Rule.KindId)
		if areaGame == nil {
			xlog.Logger().Errorln("不存在的区域游戏:", floor.Rule.KindId)
			continue
		}
		if floor.Rule.RecommendVersion != areaGame.RecommVersion {
			if floor.Rule.RecommendVersion > areaGame.RecommVersion {
				xlog.Logger().Errorf("包厢推荐规则版本异常:hid(%d),fid(%d),fVer(%d),Ver(%d)",
					clb.DBClub.HId, floor.Id, floor.Rule.RecommendVersion, areaGame.RecommVersion)
			}
			floors = append(floors, floor.Id)
		}
	}
	return floors
}

// // 目前只刷新包厢和成员 不刷新楼层 后期根据需求可加上
// func (h *Club) RefreshHouseMember() error {
// 	h.flush()
// 	mems := h.GetAllMem()
// 	for _, mem := range mems {
// 		updateMap := make(map[string]interface{})
// 		updateMap["hid"] = mem.DHId
// 		updateMap["ref"] = mem.Ref
// 		updateMap["fid"] = mem.FId
// 		updateMap["uid"] = mem.UId
// 		updateMap["urole"] = mem.URole
// 		updateMap["uremark"] = mem.URemark
// 		updateMap["apply_time"] = time.Unix(mem.ApplyTime, 0)
// 		updateMap["agree_time"] = time.Unix(mem.AgreeTime, 0)
// 		updateMap["bw_times"] = mem.BwTimes
// 		updateMap["play_times"] = mem.PlayTimes
// 		updateMap["forbid"] = mem.Forbid
// 		updateMap["partner"] = mem.Partner
// 		updateMap["partner_open_id"] = mem.Partner_oid
// 		updateMap["uvitamin"] = mem.UVitamin
// 		var m model.HouseMember
// 		m.Id = mem.Id
// 		err := GetDBMgr().GetDBmControl().Model(m).Updates(updateMap).Error
// 		if err != nil {
// 			syslog.Logger().Errorln(err)
// 			return err
// 		}
// 	}
// 	return nil
// }

func (clb *Club) AutoPayPartner() {
	// 获取包厢所有玩家redis数据
	houseMems := clb.GetMemberMap(false)

	profitMap, err := GetDBMgr().SelectPartnerProfitWithAllPartners(clb.DBClub.Id, -1, 0, -1)
	if err != nil {
		xlog.Logger().Errorf("自动划扣[查询当天分成数据失败: err :%v] hid:%d", err, clb.DBClub.Id)
		return
	}

	var ack static.Msg_HC_HouseParnterFloorStatistics

	for _, hMem := range houseMems {
		if !hMem.IsPartner() {
			continue
		}

		statisticsItem := static.ClubPartnerFloorStatisticsItem{}
		statisticsItem.UId = hMem.UId
		statisticsItem.UName = hMem.NickName
		statisticsItem.UUrl = hMem.ImgUrl
		statisticsItem.UGender = hMem.Sex
		statisticsItem.ValidTimes = 0
		statisticsItem.BigValidTimes = 0
		statisticsItem.RoundProfit = 0
		statisticsItem.SubordinateProfit = 0
		statisticsItem.TotalProfit = 0
		statisticsItem.IsJunior = hMem.Superior > 0
		statisticsItem.VitaminAdmin = hMem.VitaminAdmin
		statisticsItem.VicePartner = hMem.VicePartner

		profit, ok := profitMap[hMem.UId]
		if ok {
			statisticsItem.ValidTimes = profit.ValidTimes
			statisticsItem.BigValidTimes = 0
			statisticsItem.RoundProfit = profit.SelfProfit
			statisticsItem.SubordinateProfit = profit.SubProfit
			statisticsItem.TotalProfit = profit.SelfProfit + profit.SubProfit
		}
		ack.Items = append(ack.Items, &statisticsItem)
	}

	var total int
	for _, item := range ack.Items {
		total += item.TotalProfit
	}
	cli := GetDBMgr().Redis
	cli.Set(clb.partnerPaySumKey(), total, 24*time.Hour) //应付总额
	if !clb.DBClub.AutoPayPartnrt {
		return
	}

	curVitminMax := clb.GetVitaminMax()

	for _, item := range ack.Items {
		hm := clb.GetMemByUId(item.UId)
		if hm == nil {
			xlog.Logger().Errorf("house member not found in static data:%d", item.UId)
			continue
		}
		total := int64(item.TotalProfit)
		left := clb.GetHouseVitaminPool(&models.UserAgentConfig{VitaminPoolMax: curVitminMax})
		if left < total {
			xlog.Logger().Errorf("house pool vitamin not enough:left:%d,need:%d,hid:%d,uid:%d", left, total, clb.DBClub.Id, hm.UId)
			return
		}
		tx := GetDBMgr().GetDBmControl().Begin()
		xerr := clb.PoolChange(item.UId, models.AutoPayPartner, -1*total, tx)
		if xerr != nil {
			xlog.Logger().Errorf("add house vitamin error:%s,user:%d,hid:%d", xerr.Error(), item.UId, clb.DBClub.Id)
			tx.Rollback()
			continue
		}
		_, begin := static.GetTimeLastDayStartAndEnd(time.Now())
		sql := `select 1 from house_member_vitamin_log where dhid = ? and type = ? and uid = ? and created_at >= ? ` //检查今天之后是否有划扣
		var gn []int64
		err := tx.Raw(sql, clb.DBClub.Id, models.AutoPayPartner, item.UId, begin).Scan(&gn).Error
		if err != nil {
			xlog.Logger().Errorf("query paied error:%v", err)
			tx.Rollback()
			continue
		}
		if len(gn) != 0 { //已经划扣
			xlog.Logger().Errorf("already paied")
			tx.Rollback()
			continue
		}
		_, _, err = hm.VitaminIncrement(0, total, models.AutoPayPartner, tx)
		if err != nil {
			xlog.Logger().Errorf("add user vitamin error:%v,user:%d,hid:%d", err, item.UId, clb.DBClub.Id)
			tx.Rollback()
			continue
		}

		err = tx.Commit().Error
		if err != nil {
			xlog.Logger().Errorf("tx commit  error:%v,user:%d,hid:%d", err, item.UId, clb.DBClub.Id)
			tx.Rollback()
			continue
		}
		hm.Flush()
	}
	return
}

// CheckUsersCribbersSameTable 检查用户是否被智能防作弊规则限制，用户id必须作为第一个参数传入
func (clb *Club) CheckUsersCribbersSameTable(uids ...int64) (bool, *xerrors.XError) {
	if len(uids) <= 1 {
		return true, nil
	}
	cur := uids[0]
	userCribbers, err := GetDBMgr().GetDBrControl().GetMemCribberList(clb.DBClub.Id, cur)
	if err != nil {
		xlog.Logger().Error("CheckUsersCribbersSameTable", err)
		return false, xerrors.DBExecError
	}
	for _, uid := range uids {
		if cur == uid {
			continue
		}
		times, ok := userCribbers[uid]
		if ok && times < GetServer().ConHouse.MixAIOffRoundNum {
			//syslog.Logger().Errorf("[作弊玩家]包厢id：%d 玩家A：%d  玩家B：%d", clb.DBClub.HId, uids[0], uid)
			return false, xerrors.NewXError(fmt.Sprintf("[作弊玩家]包厢id：%d 玩家A：%d  玩家B：%d", clb.DBClub.HId, uids[0], uid))
		}
	}
	return true, nil
}

func (clb *Club) SyncAllPartnerPyramid(tx *gorm.DB, houseMemberMap models.HouseMembersMap) error {
	var err error
	if houseMemberMap == nil {
		houseMembers := make([]*models.HouseMember, 0)
		err = tx.Find(&houseMembers, "hid = ?", clb.DBClub.Id).Error
		if err != nil {
			return err
		}
		houseMemberMap = make(models.HouseMembersMap)
		for _, m := range houseMembers {
			houseMemberMap[m.UId] = m
		}
	}
	// 得到楼层房费配置
	paySlice := make(models.HouseFloorGearPaySlice, 0)
	if err := tx.Find(&paySlice, "hid = ?", clb.DBClub.Id).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	floorCost := make(map[int64]int64)
	for i := 0; i < len(paySlice); i++ {
		mod := paySlice[i]
		if floor := clb.GetFloorByFId(mod.FId); floor != nil {
			if c := mod.BaseCost(); c == models.InvalidPay {
				floorCost[mod.FId] = -1
			} else {
				floorCost[mod.FId] = c
			}
		}
	}
	for fid, total := range floorCost {
		err := SyncFloorAllPartnerTotal(tx, houseMemberMap, clb.DBClub.Id, fid, total)
		if err != nil {
			return err
		}
	}
	return nil
}

// 删除合伙人关联
func (clb *Club) deletePartnerRelevance(partnerId int64, tx *gorm.DB) error {
	err := static.ErrorsCheck(
		// 更新队长的上级字段和队长字段
		tx.Model(models.HouseMember{}).Where("hid = ? and uid = ?", clb.DBClub.Id, partnerId).Updates(map[string]interface{}{
			"partner":      0,
			"superior":     0,
			"vice_partner": 0,
		}).Error,
		// 名下玩家置为普通成员
		tx.Model(models.HouseMember{}).Where("hid = ? and partner = ?", clb.DBClub.Id, partnerId).Update(map[string]interface{}{
			"partner":      0,
			"vice_partner": 0,
		}).Error,
		// 名下队长置为一级队长
		tx.Model(models.HouseMember{}).Where("hid = ? and superior = ?", clb.DBClub.Id, partnerId).Update("superior", 0).Error,
		tx.Exec(`delete from house_partner_pyramid where dhid = ? and uid = ? `, clb.DBClub.Id, partnerId).Error,
	)

	if err != nil {
		xlog.Logger().Error(err)
		return err
	}
	return nil
}

// 队长删除
func (clb *Club) PartnerDelete(partner *HouseMember, tx *gorm.DB) *xerrors.XError {
	if partner == nil {
		return xerrors.InValidHousePartnerError
	}

	if !partner.IsPartner() {
		return xerrors.InValidHousePartnerError
	}

	if partner.UId == clb.DBClub.UId {
		return xerrors.HouseMemAllReadyPartnerError
	}
	if partner.VitaminAdmin {
		return xerrors.HouseMemAllReadyPartnerError
	}

	var err error
	if tx == nil {
		tx = GetDBMgr().GetDBmControl().Begin()
		defer static.TxCommit(tx, err)
	}

	err = static.ErrorsCheck(
		// 收益表充值
		clb.RestorePartnerRoyaltyConfig(partner, tx),
		clb.deletePartnerRelevance(partner.UId, tx),
	)

	if err != nil {
		xlog.Logger().Error(err)
		return xerrors.DBExecError
	}

	// Redis
	cli := GetDBMgr().Redis
	allMembers := clb.GetMemSimple(false)
	for i := 0; i < len(allMembers); i++ {
		mem := allMembers[i]
		var flag bool
		if mem.Partner == partner.UId {
			mem.Partner = 0
			if mem.IsVicePartner() {
				mem.VicePartner = false
				mem.SendMsg(consts.MsgTypeHouseVicePartnerSet_Ntf, &static.MsgHouseVicePartnerSet{
					Hid:         clb.DBClub.HId,
					OptUid:      clb.DBClub.UId,
					Uid:         mem.UId,
					VicePartner: false,
				})
			}
			flag = true
		}
		if mem.Superior == partner.UId {
			mem.Superior = 0
			flag = true
		}
		if mem.UId == partner.UId {
			mem.Partner = 0
			mem.Superior = 0
			flag = true
		}
		if flag {
			mem.Lock(cli)
			mem.Flush()
			mem.Unlock(cli)
		}
	}

	return nil
}

// 同步实时数据
func (clb *Club) SyncLiveData() {
	mems := clb.GetMemSimple(false)
	for _, m := range mems {
		clb.TotalMemIncr(m.UId)
		if m.IsOnline {
			clb.OnlineMemIncr(m.UId)
		}
	}
}

func (clb *Club) GetSyncDuring() time.Duration {
	clb.FloorLock.RLock()
	defer clb.FloorLock.RUnlock()
	var count int
	for _, f := range clb.Floors {
		if f.IsMix {
			count += f.GetInGameTableCount()
		}
	}
	if count < 10 {
		return 1 * time.Second
	}
	if count < 30 {
		return 3 * time.Second

	}
	return 5 * time.Second
}

func (clb *Club) GetMemberMap(all bool) map[int64]HouseMember {
	members := clb.GetMemSimple(all)
	res := make(map[int64]HouseMember, len(members))
	for _, m := range members {
		res[m.UId] = m
	}
	return res
}

// 得到玩家最上级队长
func (clb *Club) GetPartnerTopside(curPartner *HouseMember) (topside *HouseMember, xerror *xerrors.XError) {
	if curPartner.Partner != 1 {
		return nil, xerrors.InValidHousePartnerError
	}
	if curPartner.Superior == 0 {
		return curPartner, nil
	}
	var (
		memberMap = clb.GetMemberMap(false) //包厢成员map
		curSuper  = curPartner.Superior     // 当前上级
	)
	for curSuper > 0 {
		super, ok := memberMap[curSuper]
		if !ok {
			return nil, xerrors.InValidHousePartnerError
		}
		topside = &super
		curSuper = topside.Superior
	}
	return topside, nil
}

// 得到合伙人深度
func (clb *Club) GetPartnerDeep(memMap map[int64]HouseMember, cur int64) (deep int, super int64, customError *xerrors.XError) {
	curPartner, ok := memMap[cur]
	if !ok {
		return 0, 0, xerrors.InValidHouseMemberError
	}

	if curPartner.IsPartner() {
		super = curPartner.Superior
		deep++
	} else {
		return deep, super, nil
	}

	if super > 0 {
		nextSuper := super
		for nextSuper > 0 {
			deep++
			nextPartner, ok2 := memMap[nextSuper]
			if !ok2 {
				return 0, 0, xerrors.InValidHousePartnerError
			}
			nextSuper = nextPartner.Superior
		}
	}
	return deep, super, nil
}

func (clb *Club) GetFloorsSingleCostMap() (ClubFloorVitaminCost, error) {
	res := make(ClubFloorVitaminCost)
	floors, _ := clb.GetFloors()
	for i := 0; i < len(floors); i++ {
		res[floors[i]] = InvalidVitaminCost
	}
	paySlice := make(models.HouseFloorGearPaySlice, 0)
	if err := GetDBMgr().GetDBmControl().Where("hid = ?", clb.DBClub.Id).Find(&paySlice).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		// syslog.Logger().Error(err)
		return res, err
	}
	for i := 0; i < len(paySlice); i++ {
		mod := paySlice[i]
		if c := mod.BaseCost(); c == models.InvalidPay {
			res[mod.FId] = InvalidVitaminCost
		} else {
			res[mod.FId] = c
		}
	}
	return res, nil
}

func (clb *Club) IsAiSuper() bool {
	return clb.DBClub.MixActive &&
		clb.DBClub.TableJoinType == consts.NoCheat &&
		clb.DBClub.AiSuper
}

// IsHideOnlineNum 是否隐藏在线数据
func (clb *Club) IsHideOnlineNum(uid int64) bool {
	// 如果不是圈主  且 (开启超级防作弊模式 或者 开启了隐藏开关) 则隐藏在线数据
	return (clb.IsAiSuper() || clb.DBClub.IsOnlineHide) && uid != clb.DBClub.UId
}

// ConfiguredEmptyTblCount 得到配置的每个楼层的空桌子数量
func (clb *Club) ConfiguredEmptyTblCount() int {
	if clb.DBClub.TableJoinType == consts.AutoAdd {
		if clb.DBClub.EmptyTableMax <= 0 {
			return 1
		}
		return clb.DBClub.EmptyTableMax
	}
	return 0
}

func (clb *Club) GetMemWithPartner() map[int64]HouseMember { //获取所有队长及队长名下用户
	mems := clb.GetMemSimple(false)
	items := make(map[int64]HouseMember, len(mems)/3)
	for _, mem := range mems {
		// 非楼主
		if mem.URole <= 1 {
			continue
		}
		// 当前队长
		if mem.Partner >= 1 {
			items[mem.UId] = mem
		}
	}
	return items
}

func (clb *Club) GetPartnerUidRelation() map[int64][]int64 { //获取队长所有下级的uid
	items := clb.GetMemWithPartner()
	dest := make(map[int64][]int64, len(items)/3)
	for _, mem := range items {

		if mem.Partner == 1 {

			if dest[mem.UId] == nil {
				dest[mem.UId] = []int64{}
			}

			uids := clb.GetLowerPartner(mem, []int64{}, items)
			for _, uid := range uids {
				if dest[uid] == nil {
					dest[uid] = []int64{mem.UId}
				} else {
					dest[uid] = append(dest[uid], mem.UId)
				}
			}
			continue
		}
		upPartner := items[mem.Partner]
		if upPartner.UId == 0 {
			continue
		}

		if dest[upPartner.UId] == nil {
			dest[upPartner.UId] = []int64{mem.UId}
		} else {
			dest[upPartner.UId] = append(dest[upPartner.UId], mem.UId)
		}
	}
	return dest
}

func (clb *Club) GetLowerPartner(mem HouseMember, dest []int64, memMap map[int64]HouseMember) []int64 {

	if mem.Superior == 0 || mem.Partner > 1 {
		return dest
	}
	upMem := memMap[mem.Superior]
	if upMem.UId == 0 {
		return dest
	}

	dest = append(dest, upMem.UId)
	return clb.GetLowerPartner(upMem, dest, memMap)
}

type UidGameSum struct {
	Uid        int64
	Send       float64
	Bwtimes    int
	Playtimes  int
	Validtimes int
}

// 团队统计  获取队长及其名下玩家的疲劳值输赢统计，uid为0 则返回全部
func (clb *Club) SelectHouseTeamStatistics(fid int, start, end time.Time, memMap map[int64]*HouseMember, leaveMemMap map[int64]int64) map[int64]*UidGameSum {
	parShip := clb.GetPartnerUidRelation()
	xlog.Logger().Infof("get partner ship:%+v", parShip)

	result, err := GetDBMgr().SelectHouseMemberStatisticsWithTotal(clb.DBClub.Id, fid, start, end)

	xlog.Logger().Infof("quere done sum eve")

	pStatisticsMap := make(map[int64]*UidGameSum)

	if err != nil || len(result) == 0 {
		xlog.Logger().Errorf("get user sum error:%v", err)
		return pStatisticsMap
	}
	xlog.Logger().Infof("get sum eve:%+v", result)

	for _, pMem := range memMap {
		if pMem.IsPartner() {
			pStatisticsMap[pMem.UId] = &UidGameSum{Uid: pMem.UId}
		}
	}
	for _, item := range result {
		mem, ok := memMap[item.Uid]
		if ok {
			pUid := int64(0)
			if mem.URole == consts.ROLE_MEMBER {
				if mem.IsPartner() {
					pUid = mem.UId
				} else if mem.Partner > 1 {
					pUid = mem.Partner
				} else {
					continue
				}
			} else {
				continue
			}
			pStatistics, ok := pStatisticsMap[pUid]
			if !ok {
				pStatistics = &UidGameSum{Uid: pUid}
				pStatisticsMap[pUid] = pStatistics
			}
			pStatistics.Send += item.TotalScore
			pStatistics.Bwtimes += item.BigWinTimes
			pStatistics.Playtimes += item.PlayTimes
			pStatistics.Validtimes += item.ValidTimes
		}
	}

	for _, item := range result {
		_, ok := memMap[item.Uid]
		if !ok {
			if pUid, ok := leaveMemMap[item.Uid]; ok {
				pStatistics, ok := pStatisticsMap[pUid]
				if !ok {
					continue
				}
				pStatistics.Send += item.TotalScore
				pStatistics.Bwtimes += item.BigWinTimes
				pStatistics.Playtimes += item.PlayTimes
				pStatistics.Validtimes += item.ValidTimes
			}
		}
	}

	for id, lowers := range parShip {
		p1Statistics, ok := pStatisticsMap[id]
		if !ok {
			p1Statistics = &UidGameSum{Uid: id}
			pStatisticsMap[id] = p1Statistics
		}
		for _, uid := range lowers {
			p2Statistics, ok := pStatisticsMap[uid]
			if !ok {
				p2Statistics = &UidGameSum{Uid: uid}
				pStatisticsMap[uid] = p2Statistics
			}
			p1Statistics.Send += p2Statistics.Send
			p1Statistics.Bwtimes += p2Statistics.Bwtimes
			p1Statistics.Playtimes += p2Statistics.Playtimes
			p1Statistics.Validtimes += p2Statistics.Validtimes
		}
	}

	return pStatisticsMap
}

func (clb *Club) partnerPaySumKey() string {
	return fmt.Sprintf("partner_house_pay_sum_%d_%s", clb.DBClub.Id, time.Now().Format("20060102"))
}

func (clb *Club) GetLastShouldPay() int {
	cli := GetDBMgr().Redis
	res := cli.Get(clb.partnerPaySumKey()).Val()
	sum, err := strconv.ParseInt(res, 10, 64)
	if err != nil { //TODO:重新从数据库计算，一般不会出现
		xlog.Logger().Errorf("err:%v,%s", err, res)
		return 0
	}
	return int(sum)
}

func (clb *Club) GetLastPaied() int {
	db := GetDBMgr().GetDBmControl()
	_, e := static.GetTimeLastDayStartAndEnd(time.Now())
	var gn []int64
	err := db.Table("house_member_vitamin_log").Select("SUM(value)").
		Where("dhid = ? and type = ? and created_at >= ? ", clb.DBClub.Id, models.AutoPayPartner, e).
		Pluck("SUM(value)", &gn).Error
	if err != nil || len(gn) == 0 {
		xlog.Logger().Errorf("query error:%v", err)
		return 0
	}
	return int(gn[0])

}

func (clb *Club) GetLastDayTaxSum() int {
	db := GetDBMgr().GetDBmControl()
	_, e := static.GetTimeLastDayStartAndEnd(time.Now())
	var gn []int64
	err := db.Table("house_vitamin_pool_log").
		Select("value").
		Where("hid=? and optype = ? and created_at >= ? ", clb.DBClub.Id, models.TaxSum, e).Pluck("value", &gn).Error
	if err != nil || len(gn) == 0 {
		xlog.Logger().Errorf("query error:%v", err)
		return 0
	}
	return int(gn[0])
}

func (clb *Club) GetCardCost() int {
	db := GetDBMgr().GetDBmControl()
	s, e := static.GetTimeLastDayStartAndEnd(time.Now())
	var gn []int64
	err := db.Table("record_game_cost").
		Select("value").
		Where("hid=?  and created_at >= ? and created_at < ?", clb.DBClub.Id, s, e).Pluck("sum(kacost)", &gn).Error
	if err != nil || len(gn) == 0 {
		xlog.Logger().Errorf("query error:%v", err)
		return 0
	}
	return int(gn[0])
}

func (clb *Club) IsVitaminAdmin(UId int64) bool {
	mem := clb.GetMemByUId(UId)
	if mem != nil {
		return mem.IsVitaminAdmin()
	}
	return false
}

func (clb *Club) GetVitaminCount() int {
	mems := clb.GetMemSimple(false)
	var count int
	for _, mem := range mems {
		if mem.VitaminAdmin {
			count++
		}
	}
	return count
}

func (clb *Club) GetVicePartnerCount(partner int64) int {
	mems := clb.GetMemSimple(false)
	var count int
	for _, mem := range mems {
		if mem.Partner == partner && mem.VicePartner {
			count++
		}
	}
	return count
}

func (clb *Club) sureUserGroupMap(uid int64) {
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	if clb.UserGroup == nil {
		clb.UserGroup = make(map[int64]map[int][]int64)
		clb.UserGroup[uid] = make(map[int][]int64)
		return
	}
	if clb.UserGroup[uid] == nil {
		clb.UserGroup[uid] = make(map[int][]int64)
	}
	return
}

func (clb *Club) AddUserGroup(uid int64) (int, *xerrors.XError) {
	clb.sureUserGroupMap(uid)
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	var res GroupDb
	if len(clb.UserGroup[uid]) >= 10 {
		return 0, xerrors.GroupMax
	}
	groupIDSql := `select group_id from house_group_user where hid = ? and puid = ? order by group_id desc limit 1`
	err := GetDBMgr().GetDBmControl().Raw(groupIDSql, clb.DBClub.Id, uid).Scan(&res).Error
	sql := `insert into house_group_user(hid,puid,group_id) values 
	(?,?,?)`
	err = GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, uid, res.GroupId+1).Error
	if err != nil {
		return 0, xerrors.DBExecError
	}
	clb.UserGroup[uid][res.GroupId+1] = make([]int64, 0)
	return res.GroupId + 1, xerrors.RespOk
}
func (clb *Club) RemoveUserGroup(uid int64, id int) error {
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	sql := `update house_group_user set status = 1 where hid= ? and puid = ? and group_id = ?`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, uid, id).Error
	if err != nil {
		xlog.Logger().Errorf("sql error:%v", err)
		return err
	}
	delete(clb.UserGroup[uid], id)
	return nil
}

func (clb *Club) AddUserGroupUser(opuid int64, groupID int, uid int64) *xerrors.XError {
	clb.sureUserGroupMap(opuid)
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	userMap := clb.UserGroup[opuid][groupID]
	if userMap != nil {
		if len(userMap) >= 100 {
			return xerrors.GroupUserMax
		}
	} else {
		return xerrors.GroupMax
	}
	sql := `insert into house_group_user(hid,puid,group_id,uid) values(?,?,?,?) ON DUPLICATE KEY UPDATE status = 0`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, opuid, groupID, uid).Error
	if err != nil {
		return xerrors.DBExecError
	}
	in := false
	for _, u := range userMap {
		if u == uid {
			in = true
		}
	}
	if !in {
		clb.UserGroup[opuid][groupID] = append(userMap, uid)
	}

	return xerrors.RespOk
}

func (clb *Club) RemoveUserGroupUser(opuid int64, groupID int, uid int64) error {
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	sql := `update house_group_user set status = 0 where hid = ? and puid = ? and group_id = ? and uid = ?`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, opuid, groupID, uid).Error
	if err != nil {
		return err
	}
	clb.UserGroup[opuid][groupID] = static.RemoveFromSlice(clb.UserGroup[opuid][groupID], uid)
	return nil
}

func (clb *Club) GetUserGroupInfo(uid int64) *static.MsgHouseGroupInfo {
	clb.sureUserGroupMap(uid)
	dest := static.MsgHouseGroupInfo{}
	dest.Uid = uid
	clb.AddUserGroupLock.CustomLock()
	infoMap := clb.UserGroup[uid]
	if infoMap == nil {
		clb.AddUserGroupLock.CustomUnLock()
		return &dest
	}
	clb.AddUserGroupLock.CustomUnLock()
	mems := clb.GetMemSimpleToMap(false)
	for k, v := range infoMap {
		groupInfo := static.GroupInfo{}
		groupInfo.GroupID = k
		groupInfo.Users = make([]*static.LimitUserInfo, 0, len(v))
		for _, u := range v {
			hmem := mems[u]
			if hmem == nil {
				continue
			}
			var titem static.LimitUserInfo
			titem.UId = u
			titem.UName = hmem.NickName
			titem.UUrl = hmem.ImgUrl
			titem.UGender = hmem.Sex
			titem.Limit = true
			groupInfo.Users = append(groupInfo.Users, &titem)
		}
		sort.Sort(static.LimitUserSlie(groupInfo.Users))
		groupInfo.UserCount = len(groupInfo.Users)
		dest.Groups = append(dest.Groups, &groupInfo)
	}
	sort.Sort(static.LimitGroups(dest.Groups))
	dest.TotalGroup = len(dest.Groups)
	return &dest
}

func (clb *Club) RemoveGroupUser(mem *HouseMember) {
	if mem == nil || mem.Partner <= 1 {
		return
	}
	clb.AddUserGroupLock.CustomLock()
	gmap := clb.UserGroup[mem.Partner]
	if gmap == nil {
		clb.AddUserGroupLock.CustomUnLock()
		return
	}
	clb.AddUserGroupLock.CustomUnLock()
	for k, v := range gmap {
		for _, u := range v {
			if u == mem.UId {
				clb.RemoveUserGroupUser(mem.Partner, k, mem.UId)
				return
			}
		}

	}
}

func (clb *Club) DeletePartner(uid int64) {
	gmap := clb.UserGroup[uid]
	if gmap == nil {
		return
	}
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	sql := `update house_group_user set status = 1 where hid= ? and puid = ? `
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, uid).Error
	if err != nil {
		xlog.Logger().Errorf("delete group user error:%v", err)
		return
	}
	delete(clb.UserGroup, uid)
	return

}

func (clb *Club) GetAllPartner() []int64 {
	dest := []int64{}
	mems := clb.GetMemSimple(false)
	for _, mem := range mems {
		if mem.Partner == 1 {
			dest = append(dest, mem.UId)
		}
	}
	return dest
}

func (clb *Club) SureHouseFiveGroup() {
	partners := clb.GetAllPartner()
	if clb.UserGroup == nil {
		clb.UserGroup = make(map[int64]map[int][]int64)
	}
	for _, uid := range partners {
		if clb.UserGroup[uid] == nil {
			clb.UserGroup[uid] = make(map[int][]int64)
		}
	}

	for u, umap := range clb.UserGroup {
		if len(umap) < 5 {
			left := 5 - len(umap)
			for i := 0; i < left; i++ {
				clb.AddUserGroup(u)
			}
		}
	}

}

func (clb *Club) SearchUser(key string, all bool) []HouseMember {
	if key == "" {
		return clb.GetMemSimple(all)
	}
	mems := clb.GetMemSimple(all)
	dest := make([]HouseMember, 16)
	for _, mem := range mems {
		// ID 包含
		if strings.Contains(fmt.Sprintf("%d", mem.UId), key) {
			dest = append(dest, mem)
			continue
		}
		if strings.Contains(mem.NickName, key) {
			dest = append(dest, mem)
			continue
		}
	}
	return dest
}

func (clb *Club) SearchUserSort(key string, all bool) []HouseMember {
	dest := clb.SearchUser(key, all)
	sort.Sort(HouseMemberItemWrapper{dest, func(item1, item2 *HouseMember) bool {
		if item1.UId > item2.UId {
			return true
		}
		return false
	}})
	return dest
}

func (clb *Club) SearchUserToMap(key string) map[int64]*HouseMember {
	if key == "" {
		return clb.GetMemSimpleToMap(false)
	}
	mems := clb.GetMemSimple(false)
	dest := make(map[int64]*HouseMember, 0)
	for _, mem := range mems {
		// ID 包含
		l := mem
		if strings.Contains(fmt.Sprintf("%d", l.UId), key) {
			dest[l.UId] = &l
			continue
		}
		if strings.Contains(l.NickName, key) {
			dest[mem.UId] = &l
			continue
		}
	}
	return dest
}

func (clb *Club) GetGroupUser(puid int64, groupid int) []int64 {
	if clb.UserGroup == nil {
		return nil
	}
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	userMap := clb.UserGroup[puid]
	if userMap == nil {
		return nil
	}
	var dest []int64
	if groupid == 0 {
		for _, uids := range userMap {
			dest = append(dest, uids...)
		}
		return dest
	} else {
		dest = userMap[groupid]
	}
	checkUser := make([]int64, 0, len(dest))
	for _, uid := range dest {
		mem := clb.GetMemByUId(uid)
		if mem != nil && mem.URole == consts.ROLE_MEMBER && mem.Partner == puid {
			checkUser = append(checkUser, uid)
		}
	}
	return checkUser
}

func (clb *Club) GetGroupUserMap(puid int64, groupid int) map[int64]int64 {
	if clb.UserGroup == nil {
		return nil
	}
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	userMap := clb.UserGroup[puid]
	if userMap == nil {
		return nil
	}
	var dest []int64
	checkUser := make(map[int64]int64)
	if groupid == 0 {
		for _, uids := range userMap {
			dest = append(dest, uids...)
			for _, uid := range uids {
				checkUser[uid] = uid
			}
		}
		return checkUser
	} else {
		dest = userMap[groupid]
	}

	for _, uid := range dest {
		mem := clb.GetMemByUId(uid)
		if mem != nil && mem.URole == consts.ROLE_MEMBER && mem.Partner == puid {
			checkUser[uid] = uid
		}
	}
	return checkUser
}

func (clb *Club) GetUserGroupIndex(puid int64, uid int64) int {
	clb.AddUserGroupLock.CustomLock()
	defer clb.AddUserGroupLock.CustomUnLock()
	userMap, ok := clb.UserGroup[puid]
	if !ok {
		return -1
	}

	inGroupId := -1
	var groupIds []int
	for gId, uidInGList := range userMap {
		for _, uidInG := range uidInGList {
			if uid == uidInG {
				inGroupId = gId
			}
		}
		groupIds = append(groupIds, gId)
	}

	if inGroupId == -1 {
		return -1
	}
	sort.Ints(groupIds)

	for i := 0; i < len(groupIds); i++ {
		if groupIds[i] == inGroupId {
			return i
		}
	}
	return -1
}

// 得到包厢所有的队长
func (clb *Club) GetAllPartners(owner bool) []*HouseMember {
	mems := clb.GetMemSimple(false)
	partners := make([]*HouseMember, 0)
	for _, mem := range mems {
		if mem.Partner == 1 {
			partners = append(partners, &mem)
		}
		if owner && mem.UId == clb.DBClub.UId {
			partners = append(partners, &mem)
		}
	}
	return partners
}

// 得到指定队长的所有下级 self: 是否包含自己在内
func (clb *Club) GetAllJuniors(superior int64, self bool) []*HouseMember {
	mems := clb.GetMemSimple(false)
	juniors := make([]*HouseMember, 0)
	for _, mem := range mems {
		if mem.IsJunior() && mem.Superior == superior {
			juniors = append(juniors, &mem)
		} else if mem.UId == superior {
			if self {
				juniors = append(juniors, &mem)
			}
		}
	}
	return juniors
}

// RestorePartnerRoyaltyConfig 还原/删除队长配置 model.HousePartnerRoyalty{}
// 在队长被取消后，删除其分成配置 并使其名下直属的队长继承他的最上级队长配置，从而保证盟主收益不变
// 因为队长二叉树的某个节点断掉后，会直接将该节点二叉树移动至根节点盟主名下，使得名下队长收益变化，不能保证盟主收益
// 根据业务需求，所有移动至根节点盟主名下变为一级队长的必须继承其原有一级队长（顶级）的分层配置。
// dq 2.14.2
func (clb *Club) RestorePartnerRoyaltyConfig(delPartner *HouseMember, tx *gorm.DB) error {
	// 找到其名下直属下级队长
	_, juniors := clb.GetJuniorAndPlayerBySuperior(delPartner.UId)
	// 如果有直属的下级队长
	if len(juniors) > 0 {
		if err := clb.SyncJuniorRoyaltyConfig(tx, delPartner, juniors...); err != nil {
			return err
		}
	}

	return static.ErrorsCheck(
		tx.Where("dhid = ? and uid = ?", clb.DBClub.Id, delPartner.UId).Delete(models.HousePartnerPyramid{}).Error,
	)
}

// 同步下级队长的顶级队长分成配置至指定的下级队长
func (clb *Club) SyncJuniorRoyaltyConfig(tx *gorm.DB, superior *HouseMember, juniors ...int64) error {
	costs, err := clb.GetFloorsSingleCostMap()
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}

	var (
		dbError error
		memMap  static.HouseMemberMap
	)
	memMap = GetDBMgr().GetHouseMemMap(clb.DBClub.Id)
	if memMap == nil {
		return fmt.Errorf("SyncJuniorRoyaltyConfig mem map is nil")
	}

	for i := 0; i < len(juniors); i++ {
		dbError = UpdateHousePartnerTotal(tx, memMap, clb.DBClub.Id, juniors[i], costs)
		if dbError != nil {
			return dbError
		}
	}

	return nil
}

func (clb *Club) GameSwitch(on bool, admin bool) error {
	defer clb.flush()
	if admin {
		if on {
			clb.DBClub.AdminGameOn = true
			return nil
		}
		clb.DBClub.AdminGameOn = false
		clb.DBClub.GameOn = false
		clb.DBClub.IsVitamin = false //临时增加
		clb.flush()
		return nil
	}
	if on && !clb.DBClub.AdminGameOn {
		return fmt.Errorf("此功能已关闭")
	}

	if clb.DBClub.GameOn == on {
		return nil
	}
	if on == false {
		clb.DBClub.IsVitamin = false //临时增加
		clb.flush()
	}
	clb.DBClub.GameOn = on
	return nil
}

func (clb *Club) savePrize() {
	cli := GetDBMgr().Redis
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_PRIZE, clb.DBClub.Id)
	val, _ := json.Marshal(clb.PrizeVal)
	_, err := cli.Set(key, val, 0).Result()
	if err != nil {
		xlog.Logger().Error("redis save error:%v", err)
	}
	return
}

func (clb *Club) saveGroupPrize() {
	cli := GetDBMgr().Redis
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_PRIZE_GROUP, clb.DBClub.Id)
	val, _ := json.Marshal(clb.GroupPrizeVal)
	_, err := cli.Set(key, val, 0).Result()
	if err != nil {
		xlog.Logger().Error("redis save error:%v", err)
	}
	return

}

func (clb *Club) initPrize() {
	cli := GetDBMgr().Redis
	res := cli.Get(fmt.Sprintf(consts.REDIS_KEY_HOUSE_PRIZE, clb.DBClub.Id)).Val()
	dest := []int64{}
	err := json.Unmarshal([]byte(res), &dest)
	if err != nil {
		xlog.Logger().Warnf("json error：%s", res)
	}
	if len(dest) != 20 {
		clb.PrizeVal = []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		clb.savePrize()
	}
	clb.initGroupPrize()
	return
}

func (clb *Club) delPrize() {
	cli := GetDBMgr().Redis
	err := cli.Del(fmt.Sprintf(consts.REDIS_KEY_HOUSE_PRIZE, clb.DBClub.Id)).Err()
	if err != nil {
		xlog.Logger().Error(err)
	}
	err = cli.Del(fmt.Sprintf(consts.REDIS_KEY_HOUSE_PRIZE_GROUP, clb.DBClub.Id)).Err()
	if err != nil {
		xlog.Logger().Error(err)
	}
	return
}

func (clb *Club) initGroupPrize() {
	cli := GetDBMgr().Redis
	res := cli.Get(fmt.Sprintf(consts.REDIS_KEY_HOUSE_PRIZE_GROUP, clb.DBClub.Id)).Val()
	dest := []int64{}
	err := json.Unmarshal([]byte(res), &dest)
	if err != nil {
		xlog.Logger().Warnf("json error：%s", res)
	}
	if len(dest) != 20 {
		clb.GroupPrizeVal = []int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		clb.saveGroupPrize()
	}
	return
}

type RankCount struct {
	Rank  int64 `gorm:"rank"`
	Count int64 `gorm:"count"`
}

func (clb *Club) GetLuckConfig(actId int64) *[]static.RewordInfo {
	cli := GetDBMgr().Redis
	key := fmt.Sprintf(consts.REDIS_KEY_LUCKCONFIG, clb.DBClub.Id, actId)
	res := cli.Get(key).Val()
	rinfo := []static.RewordInfo{}
	if len(res) == 0 {
		dest := []RankCount{}
		sql := `select rank,count from luck_config where hid = ? and actid = ? `
		GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, actId).Scan(&dest)
		if len(dest) == 0 {
			return nil
		}
		for _, item := range dest {
			rinfo = append(rinfo, static.RewordInfo{item.Rank, item.Count})
		}
		buf, err := json.Marshal(rinfo)
		if err != nil {
			xlog.Logger().Errorf("json error :%v", err)
		} else {
			cli.Set(key, buf, 72*time.Hour)
		}
	} else {
		err := json.Unmarshal([]byte(res), &rinfo)
		if err != nil {
			xlog.Logger().Errorf("json error :%v", err)
		}
	}
	return &rinfo
}
func GetLuckConfig(dhid, actid int64) *[]static.RewordInfo {
	house := GetClubMgr().GetClubHouseById(dhid)
	return house.GetLuckConfig(actid)
}

func (clb *Club) CreateLuckActive(rewords []static.RewordInfo, actID int64) *xerrors.XError {
	for _, item := range rewords {
		if item.Rank < 0 || item.Rank > 10 || item.Count < 0 {
			return xerrors.ArgumentError
		}
	}
	tx := GetDBMgr().GetDBmControl().Begin()
	var fail bool = true
	defer func() {
		if fail {
			tx.Rollback()
		}
	}()
	sql := `insert into luck_config(hid,actid,opuid,rank,count) values(?,?,?,?,?) on DUPLICATE KEY UPDATE count = ?,opuid = ?`
	for _, item := range rewords {
		err := tx.Exec(sql, clb.DBClub.Id, actID, clb.DBClub.UId, item.Rank, item.Count, item.Count, clb.DBClub.UId).Error
		if err != nil {
			xlog.Logger().Errorf("db insert error：%v", err)
			return xerrors.DBExecError
		}
	}
	tx.Commit()
	fail = false
	GetDBMgr().Redis.Del(fmt.Sprintf(consts.REDIS_KEY_LUCKCONFIG, clb.DBClub.Id, actID))
	return nil
}

type TickCount struct {
	Ticket int64 `gorm:"ticket"`
	Used   int64 `gorm:"used"`
}

func (clb *Club) GetUserLuckTimes(uid, actid, tcount int64) int64 {
	if tcount == 0 {
		return 0
	}
	sql := `select ticket,used from luck_ticket where hid = ? and actid = ? and uid = ?`
	dest := []TickCount{}
	GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, actid, uid).Scan(&dest)
	if len(dest) == 1 {
		return (dest[0].Ticket - dest[0].Used) / tcount
	}
	return 0
}

func (clb *Club) GetUserLuckTimesWithTicket(uid, actid, tcount int64) (int64, int64) {
	if tcount == 0 {
		return 0, 0
	}
	sql := `select ticket,used from luck_ticket where hid = ? and actid = ? and uid = ?`
	dest := []TickCount{}
	GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, actid, uid).Scan(&dest)
	if len(dest) == 1 {
		return (dest[0].Ticket - dest[0].Used) / tcount, (dest[0].Ticket - dest[0].Used) % tcount
	}
	return 0, 0
}

type LuckRecord struct {
	Uid       int64     `gorm:"uid"`
	Rank      int64     `gorm:"rank"`
	CreatedAt time.Time `gorm:"created_at"`
}

func (clb *Club) GetLuckDetail(actid int64, uid int64) []LuckRecord {
	dest := []LuckRecord{}
	if uid <= 0 {
		sql := `select uid,rank,created_at from luck_record where hid = ? and actid = ? and rank != 9 order by id desc  limit 300 `
		GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, actid).Scan(&dest)
		return dest
	}
	sql := `select uid,rank,created_at from luck_record where hid = ? and actid = ? and uid = ? order by id desc  limit 300 `
	GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, actid, uid).Scan(&dest)
	return dest

}

type LuckUsed struct {
	Rank int64 `gorm:"rank"`
	Used int64 `gorm:"used"`
}

func (clb *Club) GetUsedLuckRank(actid int64) []LuckUsed {
	sql := `select count(*) used,rank from luck_record where hid= ? and actid = ? group by rank order by rank `
	dest := []LuckUsed{}
	GetDBMgr().GetDBmControl().Raw(sql, clb.DBClub.Id, actid).Scan(&dest)
	return dest
}
func (clb *Club) GetHouseTableLimitDistance() int {
	limitDistance := 0
	model := models.HouseTableDistanceLimit{}
	err := GetDBMgr().GetDBmControl().Where("dhid = ?", clb.DBClub.Id).First(&model).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			xlog.Logger().Errorf("GetHouseTableLimitDistance db error：%v", err)
		}
		limitS, err := GetDBMgr().Redis.Get("gps_house_limit").Result()
		if err != nil {
			xlog.Logger().Warn(err)
		}
		limitDest, err := strconv.ParseInt(limitS, 10, 64)
		if err != nil {
			if err == goRedis.Nil {
				xlog.Logger().Warnf("house %d not config gps_house_limit in mysql && redis", clb.DBClub.Id)
				return -1
			}
			xlog.Logger().Warn(err)
			limitDistance = -1
		} else {
			limitDistance = int(limitDest)
		}
	} else {
		limitDistance = model.TableDistanceLimit
	}
	return limitDistance
}

// 是否为一个没有合并包厢撤销合并包厢动作的圈
func (clb *Club) IsNormalNoMerge() bool {
	return clb.DBClub.MergeHId == models.HouseMergeHidStateNormal
}

// 是被合并的小圈
func (clb *Club) IsBeenMerged() bool {
	return clb.DBClub.MergeHId > models.HouseMergeHidStateNormal
}

// 合并了其他小圈
func (clb *Club) IsMerged() bool {
	return clb.DBClub.MergeHId == models.HouseMergeHidStateDevourer
}

// 是否为其大圈
func (clb *Club) IsDevourer(devourer int64) bool {
	return clb.IsBeenMerged() && clb.DBClub.MergeHId == devourer
}

// 是否为其小圈
func (clb *Club) IsSwallower(swallower int64) bool {
	th := GetClubMgr().GetClubHouseById(swallower)
	if th == nil {
		return false
	}
	return clb.IsMerged() && th.IsBeenMerged() && th.DBClub.MergeHId == clb.DBClub.Id
}

// 设置为合并包厢中/撤销合并包厢中状态
func (clb *Club) SetBusyMerging(merge bool) {
	if merge {
		clb.DBClub.MergeHId = models.HouseMergeHidStateMerging
	} else {
		clb.DBClub.MergeHId = models.HouseMergeHidStateRevoking
	}
}

// 合并过程中
func (clb *Club) IsBusyMerging() bool {
	return clb.DBClub.MergeHId <= models.HouseMergeHidStateMerging
}

func (clb *Club) InitCreatePartnerExp(uid int64) error {
	model := models.HousePartnerAttr{}
	model.DHid = clb.DBClub.Id
	model.Uid = uid
	model.CreatedAt = time.Now()
	err := GetDBMgr().GetDBmControl().Create(&model).Error
	if err != nil {
		xlog.Logger().Errorf("InitCreatePartnerExp db error：%v, hid:%d, uid:%d", err, clb.DBClub.Id, uid)
	}
	return err
}

func (clb *Club) DelCreatePartnerExp(uid int64) {
	err := GetDBMgr().GetDBmControl().Where("dhid = ? and uid = ?", clb.DBClub.Id, uid).Delete(models.HousePartnerAttr{}).Error
	if err != nil {
		xlog.Logger().Errorf("DelCreatePartnerExp db error：%v, hid:%d, uid:%d", err, clb.DBClub.Id, uid)
	}
}

func (clb *Club) MemExitApply(uid int64) *xerrors.XError {
	mod := models.NewHouseExit(clb.DBClub.Id, uid)
	tx := GetDBMgr().GetDBmControl().Begin()
	err := mod.Apply(tx)
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError
	}
	return nil
}

func (clb *Club) MemberExitAgree(optId int64, req *static.Msg_CH_HouseMemberAgree) *xerrors.XError {
	opt := clb.GetMemByUId(optId)
	if opt == nil {
		return xerrors.InValidHouseMemberError
	}

	mem := clb.GetMemByUId(req.UId)
	if mem == nil {
		return xerrors.InValidHouseMemberError
	}

	// 队长和副队长的权限
	bHavePartnerPower := false
	if clb.DBClub.IsPartnerApply {
		if opt.IsPartner() && mem.Partner == opt.UId {
			bHavePartnerPower = true
		}
		if opt.IsVicePartner() && mem.Partner == opt.Partner {
			bHavePartnerPower = true
		}
	}

	if opt.Lower(consts.ROLE_ADMIN) && !(clb.DBClub.IsPartnerApply && bHavePartnerPower) {
		return xerrors.InvalidPermission
	}

	cer := clb.memberExitAgree(mem, opt.UId, true)
	if cer != nil {
		return cer
	}

	return nil
}

func (clb *Club) memberExitAgree(mem *HouseMember, opt int64, ntf bool) *xerrors.XError {
	mod := models.NewHouseExit(clb.DBClub.Id, mem.UId)

	tx := GetDBMgr().GetDBmControl().Begin()

	err := mod.Agree(tx, opt)

	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		if cer, ok := err.(*xerrors.XError); ok {
			return cer
		} else {
			return xerrors.DBExecError
		}
	}

	cer := clb.UserExit(mem, ntf, tx)
	if cer != nil {
		tx.Rollback()
		return cer
	}

	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError
	}
	return nil
}

func (clb *Club) MemberExitRefuse(optId int64, req *static.Msg_CH_HouseMemberRefused) *xerrors.XError {
	opt := clb.GetMemByUId(optId)
	if opt == nil {
		return xerrors.InValidHouseMemberError
	}

	mem := clb.GetMemByUId(req.UId)
	if mem == nil {
		return xerrors.InValidHouseMemberError
	}

	// 队长和副队长的权限
	bHavePartnerPower := false
	if clb.DBClub.IsPartnerApply {
		if opt.IsPartner() && mem.Partner == opt.UId {
			bHavePartnerPower = true
		}
		if opt.IsVicePartner() && mem.Partner == opt.Partner {
			bHavePartnerPower = true
		}
	}

	if opt.Lower(consts.ROLE_ADMIN) && !(clb.DBClub.IsPartnerApply && bHavePartnerPower) {
		return xerrors.InvalidPermission
	}

	cer := clb.memberExitRefuse(mem.UId, opt.UId)
	if cer != nil {
		return cer
	}

	return nil
}

func (clb *Club) memberExitRefuse(uid, opt int64) *xerrors.XError {
	mod := models.NewHouseExit(clb.DBClub.Id, uid)
	tx := GetDBMgr().GetDBmControl().Begin()
	err := mod.Refuse(tx, opt)
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError
	}
	err = tx.Commit().Error
	if err != nil {
		xlog.Logger().Error(err)
		tx.Rollback()
		return xerrors.DBExecError
	}
	return nil
}

func (clb *Club) GetExitApplicants() []*models.HouseExit {
	houseExits := make([]*models.HouseExit, 0)
	err := GetDBMgr().GetDBmControl().Find(&houseExits, "hid = ? and state = ?", clb.DBClub.Id, models.HouseMemberExitApplying).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
	}
	return houseExits
}

func (clb *Club) GetFloorColorArray() []string {
	floorColor := new(models.HouseFloorColor)
	floorColor.Id = clb.DBClub.Id
	err := GetDBMgr().GetDBmControl().First(floorColor).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
	}
	return floorColor.Color()
}

// 设置房卡低于xx时提示盟主
func (clb *Club) SetFangKaTipsMinNum(minNum int) *xerrors.XError {
	clb.DBClub.FangKaTipsMinNum = minNum
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Msg_S2C_SetFangKaTipsMinNum)
	ntf.Hid = clb.DBClub.HId
	ntf.MinNum = clb.DBClub.FangKaTipsMinNum
	clb.Broadcast(consts.ROLE_CREATER, consts.MsgTypeSetFangKaTipsMinNumRsp, ntf)

	return nil
}

// 获取成员基础信息，禁止游戏的数据
func (clb *Club) GetLimitUsers() []int64 {
	res := make([]int64, 0)
	arr, err := GetDBMgr().Redis.SMembers(fmt.Sprintf(consts.REDIS_KEY_USER_LIMIT, clb.DBClub.Id)).Result()
	if err != nil {
		xlog.Logger().Error(err)
		return res
	}
	for _, a := range arr {
		res = append(res, static.HF_Atoi64(a))
	}
	return res
}

// 获取包厢功能开关默认配置
func (clb *Club) GetMemSwitchDefault() map[string]int {
	var tempSwitch map[string]int
	tempSwitch = make(map[string]int)
	tempSwitch["BanWeChat"] = 1       // 禁用微信         0禁用 1启用
	tempSwitch["CapSetDep"] = 1       // 队长设置下级队长权限 0禁止 1启用
	tempSwitch["IsRecShowParent"] = 1 //战绩详情中显示父级归属 仅仅限制 盟主才有的权限 0禁止 1启用
	return tempSwitch
}

// 设置包厢功能开关默认配置
func (clb *Club) SetMemberSwitchDefault() (map[string]int, error) {
	var hmSh models.HouseMemberSwitch
	hmSh.Hid = int(clb.DBClub.Id)
	tempSwitch := clb.GetMemSwitchDefault()
	if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberSwitch{}).Where("hid = ? ", hmSh.Hid).First(&hmSh).Error; err != nil {
		mjson, _ := json.Marshal(tempSwitch)
		hmSh.SwitchContent = string(mjson)
		if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberUserRight{}).Save(&hmSh).Error; err != nil {
			cuserror := xerrors.NewXError("给包厢添加默认开关失败")
			return nil, cuserror
		}
	}
	var hmSwitchData map[string]int
	hmSwitchData = make(map[string]int)
	err := json.Unmarshal([]byte(hmSh.SwitchContent), &hmSwitchData)
	if err != nil {
		cuserror := xerrors.NewXError("包厢开关数据转换失败")
		return nil, cuserror
	}

	for k, v := range tempSwitch {
		_, ok := hmSwitchData[k]
		if !ok {
			hmSwitchData[k] = v
			var hmSh models.HouseMemberSwitch
			hmSh.Hid = int(clb.DBClub.Id)
			mjson, _ := json.Marshal(hmSwitchData)
			if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberSwitch{}).Where("hid = ?", hmSh.Hid).Update("switch_content", string(mjson)).Error; err != nil {
				//这里没有数据存一条默认数据
				cuserror := xerrors.NewXError("更新包厢开关数据失败")
				return nil, cuserror
			}

		}
	}
	return hmSwitchData, nil
}

// 获取包厢功能开关
func (clb *Club) GetMemberSwitch() (map[string]int, error) {
	clb.ClubMemberSwitchLock.RLock()
	defer clb.ClubMemberSwitchLock.RUnlock()
	if len(clb.ClubMemberSwitch) == 0 {
		tempHmSwitch, err := clb.SetMemberSwitchDefault()
		if err != nil {
			return nil, err
		}
		clb.ClubMemberSwitch = tempHmSwitch
	}
	return clb.ClubMemberSwitch, nil
}

// 修改包厢功能开关
func (clb *Club) SetMemberSwitch(switchStr map[string]int) (map[string]int, *xerrors.XError) {
	_, err := clb.GetMemberSwitch() // 给h.HouseMemberSwitch 在这个方法中赋值了
	if err != nil {
		cuserror := xerrors.NewXError("设置失败")
		return nil, cuserror
	}
	clb.ClubMemberSwitchLock.Lock()
	isUpdate := false
	for k, v := range switchStr {
		_, ok := clb.ClubMemberSwitch[k]
		if !ok {
			clb.ClubMemberSwitch[k] = v
			isUpdate = true
		} else {
			for k1, v1 := range clb.ClubMemberSwitch {
				if k == k1 {
					if v != v1 {
						clb.ClubMemberSwitch[k1] = v
						isUpdate = true
					}
				}
			}
		}
	}
	clb.ClubMemberSwitchLock.Unlock()
	if !isUpdate {
		cuserror := xerrors.NewXError("设置成功")
		return nil, cuserror
	}
	var hmSh models.HouseMemberSwitch
	hmSh.Hid = int(clb.DBClub.Id)
	mjson, _ := json.Marshal(clb.ClubMemberSwitch)
	if err := GetDBMgr().GetDBmControl().Model(models.HouseMemberSwitch{}).Where("hid = ?", hmSh.Hid).Update("switch_content", string(mjson)).Error; err != nil {
		//这里没有数据存一条默认数据
		cuserror := xerrors.NewXError("更新包厢开关数据失败")
		return nil, cuserror
	}
	return clb.ClubMemberSwitch, nil
}

// 获取包厢功能开关
func (clb *Club) CheckMemberSwitch(key string) (int, error) {
	shData, err := clb.GetMemberSwitch()
	if err != nil {
		cuserror := xerrors.NewXError("获取包厢开关权限失败")
		return 0, cuserror
	}
	return shData[key], nil
}

// 设置战绩筛选时段
func (clb *Club) SetRecordTimeInterval(timeInterval int) *xerrors.XError {
	clb.DBClub.RecordTimeInterval = timeInterval
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}

	// 通知
	ntf := new(static.Msg_S2C_SetRecordTimeInterval)
	ntf.Hid = clb.DBClub.HId
	ntf.TimeInterval = clb.DBClub.RecordTimeInterval
	clb.Broadcast(consts.ROLE_MEMBER, consts.MsgSetRecordTimeIntervalRsp, ntf)

	return nil
}

// 队长设置低分局
func (clb *Club) PartnerSetLowScoreVal(fid int64, pid int64, val int64) *xerrors.XError {
	sql := `insert into house_partner_setting(hid,fid,pid,low_score_val) values (?,?,?,?) ON DUPLICATE KEY UPDATE low_score_val = ?`
	err := GetDBMgr().GetDBmControl().Exec(sql, clb.DBClub.Id, fid, pid, val, val).Error
	if err != nil {
		return xerrors.DBExecError
	}
	return xerrors.RespOk
}

// 团队统计  获取队长及其名下玩家的疲劳值输赢统计，uid为0 则返回全部
func (clb *Club) SelectPartnerTeamStatistics(pid int64, fid int, start, end time.Time, memMap map[int64]*HouseMember, leaveMemMap map[int64]int64) map[int64]*UidGameSum {
	parShip := clb.GetPartnerUidRelation()
	xlog.Logger().Infof("get partner ship:%+v", parShip)

	result, err := GetDBMgr().SelectPartnerMemberStatisticsWithTotal(pid, clb.DBClub.Id, fid, start, end)

	xlog.Logger().Infof("quere done sum eve")

	pStatisticsMap := make(map[int64]*UidGameSum)

	if err != nil || len(result) == 0 {
		xlog.Logger().Errorf("get user sum error:%v", err)
		return pStatisticsMap
	}
	xlog.Logger().Infof("get sum eve:%+v", result)

	for _, pMem := range memMap {
		if pMem.IsPartner() {
			pStatisticsMap[pMem.UId] = &UidGameSum{Uid: pMem.UId}
		}
	}
	for _, item := range result {
		mem, ok := memMap[item.Uid]
		if ok {
			pUid := int64(0)
			if mem.URole == consts.ROLE_MEMBER {
				if mem.IsPartner() {
					pUid = mem.UId
				} else if mem.Partner > 1 {
					pUid = mem.Partner
				} else {
					continue
				}
			} else {
				continue
			}
			pStatistics, ok := pStatisticsMap[pUid]
			if !ok {
				pStatistics = &UidGameSum{Uid: pUid}
				pStatisticsMap[pUid] = pStatistics
			}
			pStatistics.Send += item.TotalScore
			pStatistics.Bwtimes += item.BigWinTimes
			pStatistics.Playtimes += item.PlayTimes
			pStatistics.Validtimes += item.ValidTimes
		}
	}

	for _, item := range result {
		_, ok := memMap[item.Uid]
		if !ok {
			if pUid, ok := leaveMemMap[item.Uid]; ok {
				pStatistics, ok := pStatisticsMap[pUid]
				if !ok {
					continue
				}
				pStatistics.Send += item.TotalScore
				pStatistics.Bwtimes += item.BigWinTimes
				pStatistics.Playtimes += item.PlayTimes
				pStatistics.Validtimes += item.ValidTimes
			}
		}
	}

	for id, lowers := range parShip {
		p1Statistics, ok := pStatisticsMap[id]
		if !ok {
			p1Statistics = &UidGameSum{Uid: id}
			pStatisticsMap[id] = p1Statistics
		}
		for _, uid := range lowers {
			p2Statistics, ok := pStatisticsMap[uid]
			if !ok {
				p2Statistics = &UidGameSum{Uid: uid}
				pStatisticsMap[uid] = p2Statistics
			}
			p1Statistics.Send += p2Statistics.Send
			p1Statistics.Bwtimes += p2Statistics.Bwtimes
			p1Statistics.Playtimes += p2Statistics.Playtimes
			p1Statistics.Validtimes += p2Statistics.Validtimes
		}
	}

	return pStatisticsMap
}

func (clb *Club) UpdateRankInfo(rankRound int, rankWiner int, rankRecord int, rankOpen bool) *xerrors.XError {
	clb.DBClub.RankRound = rankRound
	clb.DBClub.RankWiner = rankWiner
	clb.DBClub.RankRecord = rankRecord
	clb.DBClub.RankOpen = rankOpen
	err := GetDBMgr().HouseUpdate(clb)
	if err != nil {
		return xerrors.DBExecError
	}
	return nil
}

func (clb *Club) GetVitaminMax() int64 {
	var userAgentConfig models.UserAgentConfig
	err := GetDBMgr().GetDBmControl().First(&userAgentConfig, "uid = ? and state = ?", clb.DBClub.UId, 1).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Warnf("house %d owner %d not config user agent", clb.DBClub.HId, clb.DBClub.UId)
			userAgentConfig.VitaminPoolMax = consts.VitaminStartDefault
		} else {
			xlog.Logger().Errorf("house %d owner %d get config user agent error = %v", clb.DBClub.Id, clb.DBClub.UId, err)
			xlog.Logger().Panic(err)
		}
	}
	return static.SwitchVitaminInt64(userAgentConfig.VitaminPoolMax)
}
