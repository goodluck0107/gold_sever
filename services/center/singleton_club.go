package center

import (
	"encoding/json"
	"fmt"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/jinzhu/gorm"
)

const MAX_House_ID = 999999
const MIN_House_ID = 100000

// ////////////////////////////////////////////////////////////
// ! 俱乐部管理者
type ClubMgr struct {
	ClubMap map[int]*Club
	//MapMem   map[int64]*Club
	lock           *lock2.RWMutex
	ClubIdToHidMap map[int64]int
	hidLock        *lock2.RWMutex
	initLock       *lock2.RWMutex
}

var clubMgrSingleton *ClubMgr = nil

// ! 得到包厢管理单例
func GetClubMgr() *ClubMgr {
	if clubMgrSingleton == nil {
		clubMgrSingleton = new(ClubMgr)
		clubMgrSingleton.ClubMap = make(map[int]*Club, 10000)
		clubMgrSingleton.lock = new(lock2.RWMutex)
		clubMgrSingleton.ClubIdToHidMap = make(map[int64]int, 10000) //目前有一万包厢
		clubMgrSingleton.hidLock = new(lock2.RWMutex)
		clubMgrSingleton.initLock = new(lock2.RWMutex)
		go clubMgrSingleton.Run()
	}
	return clubMgrSingleton
}

func (clm *ClubMgr) Run() {
	defer func() {
		if x := recover(); x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		<-ticker.C
		clm.CheckRunningAndAdd()
	}
}

func (clm *ClubMgr) CheckRunningAndAdd() {
	clm.lock.RLock()
	defer clm.lock.RUnlock()
	for _, clb := range clm.ClubMap {
		clb.FloorLock.RLock()
		for _, floor := range clb.Floors {
			floor.CheckRunningAndAdd(clb)
		}
		clb.FloorLock.RUnlock()
	}
}

// 包厢 DB数据初始化
func (clm *ClubMgr) HouseBaseReload(datas []*models.House) {
	clm.lock.CustomLock()
	clm.hidLock.CustomLock()

	for _, data := range datas {
		club := new(Club)
		club.DBClub = data

		club.OptLock = new(lock2.RWMutex)
		club.Floors = make(map[int64]*HouseFloor)
		club.FloorLock = new(lock2.RWMutex)
		club.AddTableLimitLock = new(lock2.RWMutex)
		club.TableLimitUserLock = new(lock2.RWMutex)
		club.AddUserGroupLock = new(lock2.RWMutex)
		club.ClubMemberSwitchLock = new(lock2.RWMutex)
		club.initPrize()
		club.IsAlive = true
		if club.DBClub.MixActive {
			go club.StartSync()
			go club.StartSyncAiSuperNum()
		}
		// 初始化总人数及在线人数
		club.SyncLiveData()
		// club.AlterOnlineMemCount(club.GetMemOnlineCounts())
		clm.ClubMap[club.DBClub.HId] = club
		clm.ClubIdToHidMap[club.DBClub.Id] = club.DBClub.HId
	}
	clm.hidLock.CustomUnLock()
	clm.lock.CustomUnLock()
}

// 包厢楼层 DB数据初始化
func (clm *ClubMgr) HouseFloorBaseReload(datas []*static.HouseFloor) {

	for _, dhfloor := range datas {
		house := clm.GetClubHouseById(dhfloor.DHId)
		if house == nil {
			// syslog.Logger().Errorln("housefloor ", dhfloor.Id, " not in house ", dhfloor.DHId)
			continue
		}

		nhfloor := new(HouseFloor)
		nhfloor.Id = dhfloor.Id
		err := json.Unmarshal([]byte(dhfloor.Rule), &nhfloor.Rule)
		if err != nil {
			xlog.Logger().Errorln(err)
			continue
		}

		nhfloor.IsAlive = true
		nhfloor.MemAct = make(map[int64]*HouseMember)
		nhfloor.Tables = make(map[int]*HouseFloorTable)
		nhfloor.MemLock = new(lock2.RWMutex)
		nhfloor.DataLock = new(lock2.RWMutex)
		nhfloor.Name = dhfloor.Name
		nhfloor.AiSuperNum = dhfloor.AiSuperNum

		nhfloor.HId = house.DBClub.HId
		nhfloor.DHId = house.DBClub.Id
		nhfloor.IsMix = dhfloor.IsMix
		nhfloor.IsVip = dhfloor.IsVip
		nhfloor.IsCapSetVip = dhfloor.IsCapSetVip
		nhfloor.IsDefJoinVip = dhfloor.IsDefJoinVip
		nhfloor.FloorVitaminOptions = dhfloor.FloorVitaminOptions
		nhfloor.IsHide = dhfloor.IsHide
		nhfloor.MinTable = dhfloor.MinTable
		nhfloor.MaxTable = dhfloor.MaxTable
		house.Floors[nhfloor.Id] = nhfloor

		if nhfloor.IsMix && house.DBClub.MixActive {
			nhfloor.ReloadHft()
		} else {
			for i := 0; i < GetServer().ConHouse.TableNum; i++ {
				hft := new(HouseFloorTable)
				hft.NTId = i
				hft.TId = 0
				hft.UserWithOnline = make([]FTUsers, nhfloor.Rule.PlayerNum)
				hft.DataLock = new(lock2.RWMutex)
				hft.CreateStamp = time.Now().UnixNano()
				nhfloor.Tables[i] = hft
			}
		}
		// go nhfloor.RunOpt()
		if !nhfloor.IsMix || !house.DBClub.MixActive {
			go nhfloor.StartSync()
		}
	}
}

// 楼层牌桌 DB数据初始化
func (clm *ClubMgr) FloorTableBaseReload(datas []interface{}) {

	for _, data := range datas {
		nftable := data.(*static.FloorTable)

		house := clm.GetClubHouseById(nftable.DHId)
		if house == nil {
			continue
		}

		floor := house.GetFloorByFId(nftable.FId)
		if floor == nil {
			continue
		}
		table := GetTableMgr().GetTable(nftable.TId)
		if table == nil {
			continue
		}
		hft := new(HouseFloorTable)
		hft.TId = table.Id
		hft.NTId = nftable.NTId
		hft.DataLock = new(lock2.RWMutex)
		hft.Begin = table.Begin
		hft.Step = table.Step
		hft.UserWithOnline = make([]FTUsers, table.Config.MaxPlayerNum)
		hft.CreateStamp = table.Table.CreateStamp
		if len(table.Users) > table.Config.MaxPlayerNum {
			xlog.Logger().Errorf("error data:,len users:%d,table config users:%d", len(table.Users), table.Config.MaxPlayerNum)
			continue
		}
		for i := 0; i < len(table.Users); i++ {
			if table.Users[i] != nil {
				p, err := GetDBMgr().GetDBrControl().GetPerson(table.Users[i].Uid)
				if err != nil || p == nil {
					continue
				}
				hft.UserWithOnline[i] = FTUsers{Uid: table.Users[i].Uid, OnLine: p.Online, Ready: table.Users[i].Ready}
			}
		}
		hft.Watchers = GetDBMgr().GetDBrControl().GetWatchTablePlayer(table.Id)

		floor.Tables[nftable.NTId] = hft
	}
	return
}

// 包厢成员 DB数据初始化
func (clm *ClubMgr) HouseAreaCheck() {
	clm.lock.CustomLock()
	defer clm.lock.CustomUnLock()

	for _, house := range clm.ClubMap {

		if house.DBClub.Area == "" {
			ids, kinds := house.GetFloors()
			if len(kinds) == 0 {
				house.DBClub.Area = "420000"
			} else {
				apkg := GetAreaPackageByKid(kinds[0])
				if len(ids) > 0 && apkg != nil {
					// 有楼层以第一层楼id为准
					house.DBClub.Area = apkg.Code
				} else {
					// 无楼层以湖北省为区域id
					house.DBClub.Area = "420000"
				}
			}

			db_house := house.DBClub
			err := GetDBMgr().GetDBrControl().HouseInsert(db_house)
			if err != nil {
				continue
			}

			//　多层楼区域均以第一层楼为准　不同就删除
			if len(ids) > 1 {
				for idx, id := range ids {
					area := GetAreaPackageByKid(kinds[idx])
					if area != nil && house.DBClub.Area != area.Code {
						house.FloorDelete(id, nil)
					}
				}
			}
		}
	}
}

// ! 获取包厢 by hid (唯一识别码)
func (clm *ClubMgr) GetClubHouseByHId(hid int) *Club {
	clm.lock.RLock()
	defer clm.lock.RUnlock()

	house, ok := clm.ClubMap[hid]
	if ok {
		return house
	}
	return nil
}

// ! 获取包厢 by id (自增id)
func (clm *ClubMgr) GetClubHouseById(id int64) *Club {

	clm.hidLock.RLock()
	hid, ok := clm.ClubIdToHidMap[id]
	if !ok || hid <= 0 {
		clm.hidLock.RUnlock()
		return nil
	}
	clm.hidLock.RUnlock()
	return clm.GetClubHouseByHId(hid)
}

// ! 获取包厢楼层信息
func (clm *ClubMgr) GetHouseFloorById(dhid int64, dfid int64) *HouseFloor {
	h := clm.GetClubHouseById(dhid)
	if h != nil {
		return h.Floors[dfid]
	}
	return nil
}

func (clm *ClubMgr) GetRandomHouseId() int {

	//随机获取id
	id := 0
	count := 0
	for {
		count++
		if count > 100 {
			break
		}
		num := static.HF_GetRandom(899999)
		id = 100000 + num
		isE := false

		if clm.GetClubHouseByHId(id) != nil {
			isE = true
			id = 0
			continue
		}
		if !isE {
			break
		}
	}

	return id
}

// ! 获取包厢成员数据
func (clm *ClubMgr) GetHouseMember(hid int64, uid int64) *HouseMember {
	house := clm.GetClubHouseById(hid)
	if house != nil {
		return house.GetMemByUId(uid)
	}
	return nil
}

// ! 修改包厢hid
func (clm *ClubMgr) HouseIDChange(req *static.Msg_HosueIDChange) *xerrors.XError {
	if req == nil {
		return xerrors.InValidHouseError
	}
	oldh := clm.GetClubHouseByHId(req.OldHId)
	if oldh == nil {
		return xerrors.InValidHouseError
	}
	if !oldh.DBClub.IsFrozen {
		return xerrors.NotFrozenError
	}
	if oldh.CheckInGameTable() {
		return xerrors.ExistHouseTablePlayingError
	}
	if req.NewHId > MAX_House_ID || req.NewHId < MIN_House_ID {
		return xerrors.InValidHouseError
	}
	newh := clm.GetClubHouseByHId(req.NewHId)
	if newh != nil {
		return xerrors.HouseIdOccupyError
	}

	oldh.DBClub.HId = req.NewHId
	oldh.flush()
	clm.lock.CustomLock()
	clm.ClubMap[req.NewHId] = oldh
	clm.lock.CustomUnLock()
	clm.hidLock.CustomLock()
	clm.ClubIdToHidMap[oldh.DBClub.Id] = req.NewHId
	clm.hidLock.CustomUnLock()
	oldh.FloorLock.CustomLock()
	for _, f := range oldh.Floors {
		f.HId = req.NewHId
		err := GetDBMgr().HouseFloorUpdate(f)
		if err != nil {
			xlog.Logger().Errorf("update house floor error:%+v\n", err)
			continue
		}
	}
	oldh.FloorLock.CustomUnLock()
	mems := oldh.GetMemSimple(true)
	for _, mem := range mems {
		mem.HId = req.NewHId
		p, err := GetDBMgr().GetDBrControl().GetPerson(mem.UId)
		if p == nil || err != nil {
			continue
		}
		p.HouseId = req.NewHId
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(p.Uid, "HouseId", req.NewHId)
		mem.Flush()
	}
	clm.lock.CustomLock()
	delete(clm.ClubMap, req.OldHId)
	clm.lock.CustomUnLock()

	sql := `update house set hid = ? where id = ?`
	GetDBMgr().GetDBmControl().Exec(sql, req.NewHId, oldh.DBClub.Id)
	// 发送通知
	oldh.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseIDChange_NTF, req)
	return nil
}

// ! 创建包厢
func (clm *ClubMgr) HouseCreate(req *static.Msg_CH_HouseCreate, uId int64) (*Club, *xerrors.XError) {
	//获取牌桌id
	HouseId := clm.GetRandomHouseId()
	if HouseId <= 0 {
		return nil, xerrors.IdAuthorError
	}
	//! 包厢
	house := new(Club)
	house.Init()
	house.DBClub = new(models.House)
	house.DBClub.HId = HouseId
	house.DBClub.UId = uId
	house.DBClub.IsVitamin = true

	//! 包厢区域编码
	if p := GetPlayerMgr().GetPlayer(house.DBClub.UId); p == nil {
		return nil, xerrors.UserNotExistError
	} else {
		//if !GetServer().ConHouse.IsPerControl {
		//	league := GetAllianceMgr().GetUserLeagueInfo(p.Info.Uid)
		//	if league == nil || league.Freeze {
		//		return nil, xerrors.InvalidPermission
		//	}
		//	house.DBClub.Area = static.HF_I64toa(league.AreaCode)
		//} else {
		if AreaCodeCheck(p.Info.Area) {
			house.DBClub.Area = p.Info.Area
		} else {
			return nil, xerrors.InvalidAreaCodeError
		}
		// }
	}
	res := GetDBMgr().GetDBrControl().RedisV2.HGet(fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uId), "admin_game_on").Val()
	adminOn, err := strconv.ParseBool(res)
	if err != nil {
		adminOn = false
	}
	house.DBClub.Name = req.HName
	house.DBClub.Notify = req.HNotify
	house.DBClub.IsChecked = GetServer().ConHouse.IsChecked
	house.DBClub.IsFrozen = GetServer().ConHouse.IsFrozen
	house.DBClub.AdminGameOn = adminOn

	tx := GetDBMgr().GetDBmControl().Begin()
	id, err := GetDBMgr().HouseInsert(house, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		tx.Rollback()
		return nil, xerrors.DBExecError
	}
	// 内存
	house.DBClub.Id = id
	house.initPrize()
	//! 成员
	custerr := house.MemJoin(uId, consts.ROLE_CREATER, 0, true, tx)
	if custerr != nil {
		tx.Rollback()
		return nil, custerr
	}

	xerr := house.PoolChange(uId, models.HouseCreate, house.GetVitaminMax(), tx)
	if xerr != nil {
		xlog.Logger().Errorln(err.Error())
		tx.Rollback()
		return nil, xerr
	}
	tx.Commit()
	clm.lock.CustomLock()
	clm.ClubMap[house.DBClub.HId] = house
	clm.lock.CustomUnLock()
	clm.hidLock.CustomLock()
	clm.ClubIdToHidMap[house.DBClub.Id] = house.DBClub.HId
	clm.hidLock.CustomUnLock()
	// 更新包厢盟主数据
	person, _ := GetDBMgr().GetDBrControl().GetPerson(house.DBClub.UId)
	if person != nil {
		person.HouseId = house.DBClub.HId
		person.FloorId = 0
		person.TableId = 0
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "HouseId", person.HouseId, "FloorId", person.FloorId, "TableId", person.TableId)
	}
	return house, nil
}

// ! 删除包厢
func (clm *ClubMgr) HouseDelete(house *Club) *xerrors.XError {

	// 判定是否存在正在游戏牌桌
	house.DBClub.IsFrozen = true
	defer func() {
		house.DBClub.IsFrozen = false
	}()
	for _, floor := range house.Floors {
		floor.DataLock.RLock()
		for _, t := range floor.Tables {
			if t.TId != 0 {
				floor.DataLock.RUnlock()
				house.DBClub.IsFrozen = false
				return xerrors.ExistHouseTablePlayingError
			}
		}
		floor.DataLock.RUnlock()
	}

	// 删除所有楼层数据
	waitGroup := new(sync.WaitGroup)
	for _, floor := range house.Floors {
		waitGroup.Add(1)
		go floor.Delete(waitGroup)
	}
	cli := GetDBMgr().Redis
	cli.Del(fmt.Sprintf(consts.REDIS_KEY_HOUSE_TABLENUM, house.DBClub.Id))
	waitGroup.Wait()
	// 发送通知
	ntf := new(static.Ntf_HC_HouseDelete)
	ntf.HId = house.DBClub.HId
	house.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseDelete_Ntf, ntf)
	// 删除所有玩家数据
	arruid := house.GetMemsUid()
	tx := GetDBMgr().GetDBmControl().Begin()
	for _, id := range arruid {
		custerr := house.MemDelete(id, false, tx)
		if custerr != nil {
			tx.Rollback()
			return custerr
		}
	}

	err := GetDBMgr().HouseDelete(house, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		tx.Rollback()
		return xerrors.DBExecError
	}
	tx.Commit()
	// 更新包厢盟主数据
	person, _ := GetDBMgr().GetDBrControl().GetPerson(house.DBClub.UId)
	if person != nil {
		person.HouseId = 0
		person.FloorId = 0
		person.TableId = 0
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "HouseId", person.HouseId, "FloorId", person.FloorId, "TableId", person.TableId)
	}

	house.OptLock = nil
	house.Floors = make(map[int64]*HouseFloor, 0)
	// house.FloorLock = nil
	// house.Config = nil
	clm.lock.CustomLock()
	delete(clm.ClubMap, house.DBClub.HId)
	clm.lock.CustomUnLock()
	return nil
}

func (clm *ClubMgr) HouseTableLimitGroupAdd(hid int) *xerrors.XError {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		return xerrors.InValidHouseError
	}
	_, err := house.AddTableLimitGroup()
	if err != xerrors.RespOk {
		xlog.Logger().Errorln(err)
		return err
	}
	return err
}

func (clm *ClubMgr) HouseTableLimitGroupRemove(hid, groupID int) *xerrors.XError {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		return xerrors.InValidHouseError
	}
	err := house.RemoveTableLimitGroup(groupID)
	if err != xerrors.RespOk {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError
	}
	return xerrors.RespOk
}

func (clm *ClubMgr) HouseTableLimitUserAdd(hid, groupID int, uid int64) *xerrors.XError {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		return xerrors.InValidHouseError
	}
	mem := house.GetMemByUId(uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError
	}
	err := house.AddTableLimitUser(groupID, uid)
	if err != xerrors.RespOk {
		xlog.Logger().Errorln(err)
		return err
	}
	return err
}
func (clm *ClubMgr) HouseTableLimitUserRemove(hid, groupID int, uid int64) *xerrors.XError {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		return xerrors.InValidHouseError
	}
	mem := house.GetMemByUId(uid)
	if mem == nil {
		return xerrors.InValidHouseMemberError
	}
	err := house.RemoveTableLimitUser(groupID, uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return xerrors.DBExecError
	}
	return xerrors.RespOk
}

func (clm *ClubMgr) HouseTableLimitCheck(hid int, uids ...int64) (bool, []int64) {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		return true, nil
	}
	e, users := house.CheckUsersAllowSameTable(uids...)
	return e == nil, users
}

func (clm *ClubMgr) CheckUserIsLimitInGroup(hid, groupID int, uid int64) bool {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		//前面有检查，不可能出现
		return false
	}
	return house.CheckUserIsLimitInGroup(groupID, uid)
}

type HTableLimitDb struct {
	Hid      int       `gorm:"hid"`
	GroupId  int       `gorm:"group_id"`
	Uid      int64     `gorm:"uid"`
	UpdateAt time.Time `gorm:"column:updated_at;type:datetime"`
}

func (clm *ClubMgr) InitHouseTableLimit() {
	sql := `select hid,group_id,uid,updated_at from house_table_limit_user where status = 0`
	dest := []HTableLimitDb{}
	err := GetDBMgr().GetDBmControl().Raw(sql).Scan(&dest).Error
	if err != nil {
		panic(err)
	}
	for _, item := range dest {
		house := clm.GetClubHouseById(int64(item.Hid))
		if house == nil {
			continue
		}
		if house.LimitGroupsUpdateAt == nil {
			house.LimitGroupsUpdateAt = make(map[int]int64)
		}
		if house.LimitGroups == nil {
			house.LimitGroups = make(map[int]map[int64]bool)
		}
		if house.LimitGroups[item.GroupId] == nil {
			house.LimitGroups[item.GroupId] = make(map[int64]bool)
		}
		if item.Uid == 0 {
			house.LimitGroupsUpdateAt[item.GroupId] = item.UpdateAt.Unix()
			continue
		}
		house.LimitGroups[item.GroupId][item.Uid] = true
		house.TableLimitGropuCount = len(house.LimitGroups)
	}
}

func (clm *ClubMgr) UserInToHouse(p *static.Person, hid int) {
	items, err := GetDBMgr().ListHouseMemIn(p.Uid)
	if err != nil {
		xlog.Logger().Errorf("获取用户包厢列表失败:%v", err)
		return
	}
	cli := GetDBMgr().Redis
	for _, item := range items {
		house := clm.GetClubHouseById(item)
		if house == nil {
			continue
		}
		mem := house.GetMemByUId(p.Uid)
		if mem == nil {
			continue
		}
		mem.Lock(cli)
		if house.DBClub.HId == hid {
			if p.TableId == 0 {
				house.RedisPub(consts.MsgTypeHouseMemOnline_NTF,
					static.GameMember{Hid: house.DBClub.HId, Uid: mem.UId,
						IsOnline: true, URole: mem.URole, Uname: p.Nickname, Gender: p.Sex,
						UUrl: p.Imgurl, IsLimitGame: mem.IsLimitGame, Partner: mem.Partner})
			}
			mem.IsOnline = true
			house.OnlineMemIncr(mem.UId)
		} else {
			mem.IsOnline = false
			house.OnlineMemDecr(mem.UId)
		}

		mem.NickName = p.Nickname
		mem.Sex = p.Sex
		mem.ImgUrl = p.Imgurl
		mem.Flush()
		mem.Unlock(cli)
	}
	return
}

func (clm *ClubMgr) UserLeaveHouse(uid int64, hid int) {
	house := clm.GetClubHouseByHId(hid)
	if house == nil {
		return
	}
	mem := house.GetMemByUId(uid)
	if mem == nil {
		return
	}
	house.RedisPub(consts.MsgTypeHouseMemOffline_NTF,
		static.Msg_HC_HouseMemOnline{Hid: house.DBClub.HId, Uid: mem.UId, Online: false})
	mem.IsOnline = false
	mem.Flush()
	house.OnlineMemDecr(mem.UId)
}

func (clm *ClubMgr) UserOffLine(uid int64) {
	items, err := GetDBMgr().ListHouseIdMemIn(uid)
	if err != nil {
		xlog.Logger().Errorf("获取用户包厢列表失败:%v", err)
	}
	cli := GetDBMgr().Redis
	for _, item := range items {
		house := clm.GetClubHouseById(item)
		if house == nil {
			continue
		}
		mem := house.GetMemByUId(uid)
		if mem == nil {
			continue
		}
		mem.Lock(cli)
		house.RedisPub(consts.MsgTypeHouseMemOffline_NTF,
			static.Msg_HC_HouseMemOnline{Hid: house.DBClub.HId, Uid: mem.UId, Online: false})
		mem.IsOnline = false
		mem.Flush()
		mem.Unlock(cli)
		house.OnlineMemDecr(mem.UId)
	}

	return
}
func (clm *ClubMgr) UserOutGame(uid int64) {
	items, err := GetDBMgr().ListHouseIdMemIn(uid)
	if err != nil {
		xlog.Logger().Errorf("获取用户包厢列表失败:%v", err)
	}
	for _, item := range items {
		house := clm.GetClubHouseById(item)
		if house == nil {
			continue
		}
		mem := house.GetMemByUId(uid)
		if mem == nil {
			continue
		}
		house.RedisPub(consts.MsgTypeHouseMemOutGame_NTF,
			static.MsgUserInOutGame{Uid: mem.UId})
	}
	return
}
func (clm *ClubMgr) UserInGame(uid int64) {
	items, err := GetDBMgr().ListHouseIdMemIn(uid)
	if err != nil {
		xlog.Logger().Errorf("获取用户包厢列表失败:%v", err)
	}
	for _, item := range items {
		house := clm.GetClubHouseById(item)
		if house == nil {
			continue
		}
		mem := house.GetMemByUId(uid)
		if mem == nil {
			continue
		}
		house.RedisPub(consts.MsgTypeHouseMemInGame_NTF,
			static.MsgUserInOutGame{Uid: mem.UId})
	}
	return
}

func (clm *ClubMgr) SaveFloor() error {
	clm.lock.CustomLock()
	defer clm.lock.CustomUnLock()
	for _, house := range clm.ClubMap {
		if house == nil {
			continue
		}
		if !house.DBClub.MixActive {
			continue
		}
		for _, f := range house.Floors {
			f.SaveHft()
		}
	}
	return nil
}

func (clm *ClubMgr) InitHouseVitaminPool() {
	clm.lock.RLock()
	defer clm.lock.RUnlock()
	for _, h := range clm.ClubMap {
		var sum int64
		mems := h.GetMemSimple(false)
		for _, mem := range mems {
			sum += mem.UVitamin
		}
		h.PoolChange(h.DBClub.UId, models.HouseCreate, h.GetVitaminMax()-sum, nil)
	}
}

func (clm *ClubMgr) StatisticHouseMemVitaminLeftDay() {
	//clm.lock.RLock()
	//defer clm.lock.RUnlock()
	//for _, h := range clm.ClubMap {
	//	h.StatisticHouseMemVitaminLeftDay()
	//}
}

func (clm *ClubMgr) HouseFidsToStr(fids []int64) string {
	fidstr := ""
	for i := 0; i < len(fids); i++ {
		fidstr += static.HF_I64toa(fids[i])
		if i != len(fids)-1 {
			fidstr += ","
		}
	}
	return fidstr
}

func (clm *ClubMgr) GetHouseFidsByStr(fidstr string) []int64 {
	fids := []int64{}
	fidstrs := strings.Split(fidstr, ",")
	for i := 0; i < len(fidstrs); i++ {
		fids = append(fids, static.HF_Atoi64(fidstrs[i]))
	}
	return fids
}

// 吞吐结果
type HouseMergeResult struct {
	num      int                   // 失败个数
	err      error                 // 错误
	opMem    []*models.HouseMember // 被操作的用户列表/合并包厢是代指被合过去的uid集合/撤销合并包厢时代指被删掉的uid集合
	repCount int                   // 重复个数
}

func (hpr *HouseMergeResult) NumRepeat() int {
	return hpr.repCount
}

// 失败个数
func (hpr *HouseMergeResult) NumFailed() int {
	return hpr.num
}

func (hpr *HouseMergeResult) OptCount() int {
	return len(hpr.opMem)
}

func (hpr *HouseMergeResult) OptMember() []*models.HouseMember {
	return hpr.opMem
}

// 操作结果
func (hpr *HouseMergeResult) Error() error {
	return hpr.err
}

// 操作结果
func (hpr *HouseMergeResult) SetDBErr() {
	hpr.err = DBError(hpr.err.Error())
}

type DBError string

func (dbe DBError) Error() string {
	return fmt.Sprintf("数据库操作异常:%s", dbe)
}

// 包厢吞并(合并包厢)
func (clm *ClubMgr) HouseMerge(devourer /*吞噬者HID*/, swallower /*被吞噬者HID*/ int) *HouseMergeResult {
	hpr := new(HouseMergeResult)
	hpr.opMem = make([]*models.HouseMember, 0)

	// 得到大圈
	houseDev := clm.GetClubHouseByHId(devourer) // 吞噬圈
	if houseDev == nil {
		hpr.err = fmt.Errorf("包厢:%d不存在", devourer)
		return hpr
	}

	// 得到小圈
	houseSwa := clm.GetClubHouseByHId(swallower) // 被吞噬圈
	if houseSwa == nil {
		hpr.err = fmt.Errorf("包厢:%d不存在", swallower)
		return hpr
	}

	// 校验状态
	if houseDev.IsBusyMerging() {
		hpr.err = xerrors.HouseBusyError
		return hpr
	}
	if houseSwa.IsBusyMerging() {
		hpr.err = xerrors.HouseBusyError
		return hpr
	}
	if houseSwa.IsBeenMerged() {
		hpr.err = fmt.Errorf("对方包厢已被其他包厢合并。")
		return hpr
	}
	if houseSwa.IsMerged() {
		hpr.err = fmt.Errorf("对方包厢已合并其他包厢，无法被合并。")
		return hpr
	}
	if houseDev.IsBeenMerged() {
		hpr.err = fmt.Errorf("当前包厢已被合并，无法再合并其他包厢。")
		return hpr
	}

	if mem := houseDev.GetMemByUId(houseSwa.DBClub.UId); mem != nil && mem.Upper(consts.ROLE_APLLY) {
		hpr.err = fmt.Errorf("合并包厢失败，对方盟主已经是当前包厢成员。")
		return hpr
	}

	// 进来先把包厢最新的数据同步一遍，因为可能之前有很多数据只改了内存和redis并没有落地到mysql
	houseDev.flush()
	houseSwa.flush()
	// 记录之前的状态
	beforeHouseStateDev, beforeHouseStateSwa := houseDev.DBClub.MergeHId, houseSwa.DBClub.MergeHId

	// 开启合并包厢中状态
	houseDev.SetBusyMerging(true)
	houseSwa.SetBusyMerging(true)

	startAt := time.Now()
	xlog.Logger().Info("开始合并包厢 at:", startAt.Format(static.TIMEFORMAT))
	tx := GetDBMgr().GetDBmControl().Begin()
	defer func() {
		if hpr.Error() == nil {
			xlog.Logger().Warningln("失败人数：", hpr.NumFailed())
			tx.Commit()
			houseSwa.ReloadHouse()
			houseDev.ReloadHouse()
			endAt := time.Now()
			xlog.Logger().Infof("合并包厢成功 at:%s, 耗时 [%.2fms].", endAt.Format(static.TIMEFORMAT), float64(endAt.Sub(startAt).Nanoseconds()/1e4)/100.00)
		} else {
			tx.Rollback()
			houseDev.DBClub.MergeHId = beforeHouseStateDev
			houseSwa.DBClub.MergeHId = beforeHouseStateSwa
		}
	}()

	// 得到要合过去的玩家集合
	hpr.err = GetDBMgr().GetDBmControl().
		Where("hid = ? and urole != 3 and urole != 4 ", houseSwa.DBClub.Id).
		Where("uid not in (select uid from house_member where hid = ?)", houseDev.DBClub.Id).Find(&hpr.opMem).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	hpr.repCount = houseSwa.GetMemCounts() - hpr.OptCount()
	if hpr.repCount < 0 {
		hpr.repCount = 0
	}

	xlog.Logger().Info("本次操作受影响人数 : ", hpr.OptCount())

	for _, mem := range hpr.opMem {
		if mem == nil {
			continue
		}

		if xerr := houseDev.MemJoin(mem.UId, consts.ROLE_MEMBER, houseSwa.DBClub.Id, false, tx); xerr != nil {
			hpr.num++
			continue
		}
	}

	// existOldOwner := houseDev.GetMemByUId(houseSwa.OwnerId)
	// if existOldOwner == nil {
	// 	if err := houseDev.MemJoin(houseSwa.OwnerId, constant.ROLE_MEMBER, houseSwa.DBClub.Id, tx); err != nil {
	// 		hpr.num++
	// 	}
	// } else {
	// 	if existOldOwner.Partner != 1 {
	// 		existOldOwner.Partner = 1
	// 		existOldOwner.Flush()
	// 	}
	// }

	// 加入完成后將所有更新队长
	hpr.err = tx.Model(&models.HouseMember{}).
		Where("hid = ?", houseDev.DBClub.Id).
		Where("ref = ?", houseSwa.DBClub.Id).
		Update("partner", houseSwa.DBClub.UId).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// 将盟主更新为队长
	hpr.err = tx.Model(&models.HouseMember{}).
		Where("hid = ?", houseDev.DBClub.Id).
		Where("uid = ?", houseSwa.DBClub.UId).
		Update("partner", 1).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// 将合并包厢记录更新
	hpr.err = tx.Model(&models.HouseMergeLog{}).
		Where("swallowed = ?", houseSwa.DBClub.Id).
		Where("devourer = ?", houseDev.DBClub.Id).
		Where("merge_state = ?", models.HouseMergeStateWaiting).
		Update("merge_state", models.HouseMergeStateAproved, "merge_at", time.Now().Unix()).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// 更新合并包厢状态及冻结状态
	hpr.err = tx.Model(&models.House{}).Where("id = ?", houseSwa.DBClub.Id).Update("merge_hid", houseDev.DBClub.Id, "is_frozen", false).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// 更新合并包厢状态及冻结状态
	hpr.err = tx.Model(&models.House{}).Where("id = ?", houseDev.DBClub.Id).Update("merge_hid", models.HouseMergeHidStateDevourer).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// 被合的圈被合后 置所有向他声请合并且未处理的请求为已拒绝
	hpr.err = tx.Model(&models.HouseMergeLog{}).Where("devourer = ?", houseSwa.DBClub.Id).Where("merge_state = ?", models.HouseMergeStateWaiting).Update("merge_state", models.HouseMergeStateRefused).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// 合过圈后，其发起的所有合并请求变为已撤回
	hpr.err = tx.Model(&models.HouseMergeLog{}).Where("swallowed = ?", houseDev.DBClub.Id).Where("merge_state = ?", models.HouseMergeStateWaiting).Update("merge_state", models.HouseMergeStateRevoked).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	// // 3.执行合并包厢sql
	// sql := `insert into house_member(id,hid,ref,fid,uid,uvitamin,urole,uremark,apply_time,agree_time,bw_times,play_times,forbid,partner,partner_open_id,created_at,updated_at)
	// 		select 0,?,?,0,uid,0,2,uremark,apply_time,agree_time,0,0,forbid,case uid when ? then 1 else ? end,partner_open_id,now(),now()
	// 		from house_member where hid = ? and uid not in(select uid from house_member where hid = ?)`
	//
	// // addHms := make([]*model.HouseMember, 0)
	// // hpr.err = tx.Where("hid = ?", houseSwa.DBClub.Id).Where("uid not in(select uid from house_member where hid = ?)", houseDev.DBClub.Id).Find(&addHms).Error
	// //
	// // houseDev.MemJoin()
	//
	// hpr.err = tx.Exec(sql, houseDev.DBClub.Id, houseSwa.DBClub.Id, houseSwa.OwnerId, houseSwa.OwnerId, houseSwa.DBClub.Id, houseDev.DBClub.Id).Error
	// if hpr.err != nil {
	// 	hpr.SetDBErr()
	// 	return hpr
	// }

	return hpr
}

// 包厢吐出(撤销合并包厢)
func (clm *ClubMgr) HouseRevoke(A /*大圈hid*/, B /*小圈hid*/ int) *HouseMergeResult {
	// 定义撤销合并包厢结果
	hpr := new(HouseMergeResult)
	hpr.opMem = make([]*models.HouseMember, 0)
	// 得到大圈
	AHouse := clm.GetClubHouseByHId(A) // 吞噬圈
	if AHouse == nil {
		hpr.err = fmt.Errorf("包厢:%d不存在", A)
		return hpr
	}
	// 得到小圈
	BHouse := clm.GetClubHouseByHId(B) // 被吞噬圈
	if BHouse == nil {
		hpr.err = fmt.Errorf("包厢:%d不存在", B)
		return hpr
	}
	// 包厢合并包厢撤销合并包厢状态判断（涉及到数据量极大的时候保证单个包厢同时只存在一个合并包厢/撤销合并包厢状态,这里并没有使用atomic，原因是某些情况下的状态并不是很好处理）
	if AHouse.IsBusyMerging() {
		hpr.err = xerrors.HouseBusyError
		return hpr
	}
	if BHouse.IsBusyMerging() {
		hpr.err = xerrors.HouseBusyError
		return hpr
	}

	// 如果大圈已经被合并过，则不存在撤销合并包厢之说
	if !AHouse.IsMerged() {
		hpr.err = fmt.Errorf("包厢%d没有合并过任何子包厢，无法撤销合并包厢。", A)
		return hpr
	}
	// 如果小圈的被合并者并不是该大圈，则不存在撤销合并包厢之说
	if !BHouse.IsDevourer(AHouse.DBClub.Id) {
		hpr.err = fmt.Errorf("包厢%d没有被包厢%d合并，无效撤销合并包厢。", B, A)
		return hpr
	}

	// 进来先把包厢最新的数据同步一遍，因为可能之前有很多数据只改了内存和redis并没有落地到mysql
	AHouse.flush()
	BHouse.flush()

	// 记录之前的状态
	beforeHouseStateDev, beforeHouseStateSwa := AHouse.DBClub.MergeHId, BHouse.DBClub.MergeHId

	// 开启撤销合并包厢中状态
	AHouse.SetBusyMerging(false)
	BHouse.SetBusyMerging(false)

	defer func() {
		if hpr.Error() == nil {
			BHouse.ReloadHouse() // mysql数据已经改变 此时重复读取
			AHouse.ReloadHouse() // mysql数据已经改变 此时重复读取
			// 广播撤销合并包厢成功
			BHouse.Broadcast(consts.ROLE_MEMBER, consts.MsgTypeHouseRevokeOk, &static.MsgHouseMergeRevokeOk{
				Hid:    AHouse.DBClub.HId,
				THid:   AHouse.DBClub.HId,
				THName: AHouse.DBClub.Name,
				Msg:    fmt.Sprintf("包厢(%d)已被解除合并，请在包厢列表重新进入。", BHouse.DBClub.HId),
			})
		} else {
			AHouse.DBClub.MergeHId = beforeHouseStateDev // 撤销合并包厢失败 还原大圈状态
			BHouse.DBClub.MergeHId = beforeHouseStateSwa // 撤销合并包厢失败 还原小圈状态
		}
	}()

	// 查询B圈应带走的玩家Uid
	revokeUsersSet := make(map[int64]*models.HouseMember)
	revokeUsersIds := func() []int64 {
		res := make([]int64, 0)
		for _, ru := range revokeUsersSet {
			res = append(res, ru.UId)
		}
		return res
	}

	AHouseMembers := make([]*models.HouseMember, 0)
	hpr.err = GetDBMgr().GetDBmControl().Find(&AHouseMembers, "hid = ?", AHouse.DBClub.Id).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	AMysqlHouseMembersMap := make(models.HouseMembersMap)
	for _, m := range AHouseMembers {
		AMysqlHouseMembersMap[m.UId] = m
		if m.Ref == BHouse.DBClub.Id {
			revokeUsersSet[m.UId] = m
		}
	}

	for nextIds, next := AMysqlHouseMembersMap.GetJuniorsAndNamePlayers(BHouse.DBClub.UId); len(next) > 0; nextIds, next = AMysqlHouseMembersMap.GetJuniorsAndNamePlayers(nextIds...) {
		for _, n := range next {
			revokeUsersSet[n.UId] = n
		}
	}

	// 查找合并记录，找到除要撤走的圈外其他已经合并了的子圈，并将他们按合并时间从早到晚排序
	houseMergeLogs := make([]models.HouseMergeLog, 0)
	hpr.err = GetDBMgr().GetDBmControl().
		Where("devourer = ? AND merge_state >= ? AND swallowed != ?", AHouse.DBClub.Id, models.HouseMergeStateAproved, BHouse.DBClub.Id).
		Order("merge_at ASC").
		Find(&houseMergeLogs).Error
	if hpr.err != nil {
		if hpr.err == gorm.ErrRecordNotFound {
			xlog.Logger().Debug("撤销合并包厢时:没有其他已合并的子圈,此时不用更新标志位")
		} else {
			hpr.SetDBErr()
			return hpr
		}
	}

	// 得到除该小圈外其他还处于合并包厢状态的小圈
	otherMergeListOrder := make([]int64, 0)
	// 按时间顺序写入合并列表
	for _, hml := range houseMergeLogs {
		otherMergeListOrder = append(otherMergeListOrder, hml.Swallowed)
	}
	// 给定两个hid 筛选除合并包厢时间最早的hid
	pickOldestMergeHid := func(id1, id2 int64) int64 {
		for _, id := range otherMergeListOrder {
			if id == id1 || id == id2 {
				return id
			}
		}
		return 0
	}

	// 得到要带走的成员中，除在该小圈中外，是否还存在于其他小圈中，这一部分用户我们将其定名为 “重复成员”
	repeatHms := make([]*models.HouseMember, 0)

	// 查询子圈包含的重复成员
	if len(otherMergeListOrder) > 0 {
		hpr.err = GetDBMgr().GetDBmControl().
			Where("hid in(?)", otherMergeListOrder).
			Where("uid in(?)", revokeUsersIds()).
			Find(&repeatHms).Error
		if hpr.err != nil {
			if hpr.err == gorm.ErrRecordNotFound {
				xlog.Logger().Debug("撤销合并包厢时:没有存在于其他已合并的子圈玩家,此时不用更新标志位")
			} else {
				hpr.SetDBErr()
				return hpr
			}
		}
	}

	uidToRef := make(map[int64]int64)
	for _, hm := range repeatHms {
		if hm == nil {
			continue
		}
		if oldRef, ok := uidToRef[hm.UId]; ok {
			uidToRef[hm.UId] = pickOldestMergeHid(hm.DHId, oldRef)
		} else {
			uidToRef[hm.UId] = hm.DHId
		}
	}

	// 得到真正要带走的玩家
	for _, m := range revokeUsersSet {
		if _, ok := uidToRef[m.UId]; !ok {
			hpr.opMem = append(hpr.opMem, m)
		}
	}

	// 检查入桌情况
	realOptIds := make([]int64, 0)
	for _, m := range hpr.opMem {
		if m == nil {
			continue
		}
		realOptIds = append(realOptIds, m.UId)
	}

	if AHouse.CheckSpecifiedUsersInGameTable(realOptIds...) {
		hpr.err = fmt.Errorf("有玩家入桌，解除条件不满足，请处理后再行尝试。")
		return hpr
	}

	startAt := time.Now()
	xlog.Logger().Info("开始撤销合并包厢 at:", startAt.Format(static.TIMEFORMAT))
	xlog.Logger().Info("本次操作受影响人数 : ", hpr.OptCount())
	tx := GetDBMgr().GetDBmControl().Begin()
	defer func() {
		if hpr.Error() == nil {
			tx.Commit()
			endAt := time.Now()
			xlog.Logger().Infof("撤销合并包厢成功 at:%s, 耗时 [%.2fms].", endAt.Format(static.TIMEFORMAT), float64(endAt.Sub(startAt).Nanoseconds()/1e4)/100.00)
		} else {
			tx.Rollback()
		}
	}()

	// 如果有需要更新标志为的用户 则更新
	if len(uidToRef) > 0 {
		for uid, ref := range uidToRef {
			if m, ok := AMysqlHouseMembersMap[uid]; ok {
				hpr.err = tx.Create(&models.HouseRevokeLog{
					Id:           0,
					Swallowed:    BHouse.DBClub.Id,
					Devourer:     AHouse.DBClub.Id,
					Uid:          uid,
					OptKind:      models.HouseRevokeOptBackAndUpd,
					URole:        m.URole,
					Ref:          m.Ref,
					Partner:      m.Partner,
					VicePartner:  m.VicePartner,
					VitaminAdmin: m.VitaminAdmin,
					Superior:     m.Superior,
				}).Error
				if hpr.err != nil {
					hpr.SetDBErr()
					return hpr
				}
			}
			sql := `update house_member set ref = ?,vice_partner = 0,superior = 0,vitamin_admin = 0,partner = (select case uid when ? then 1 else uid END FROM house WHERE id = ?) where hid = ? and uid = ?`
			hpr.err = tx.Exec(sql, ref, uid, ref, AHouse.DBClub.Id, uid).Error
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
		}
	}

	now := time.Now()
	BHouseMembers := make([]*models.HouseMember, 0)
	hpr.err = tx.Find(&BHouseMembers, "hid = ?", BHouse.DBClub.Id).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}
	BMysqlHouseMembersMap := make(models.HouseMembersMap)
	for _, m := range BHouseMembers {
		BMysqlHouseMembersMap[m.UId] = m
	}

	for _, m := range revokeUsersSet {
		// 不在更新列表中需要删除
		if _, ok := uidToRef[m.UId]; !ok {
			hpr.err = tx.Create(&models.HouseRevokeLog{
				Id:           0,
				Swallowed:    BHouse.DBClub.Id,
				Devourer:     AHouse.DBClub.Id,
				Uid:          m.UId,
				OptKind:      models.HouseRevokeOptBack,
				URole:        m.URole,
				Ref:          m.Ref,
				Partner:      m.Partner,
				VicePartner:  m.VicePartner,
				VitaminAdmin: m.VitaminAdmin,
				Superior:     m.Superior,
			}).Error
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			hpr.err = tx.Create(&models.HouseMemberLog{
				DHId:      m.DHId,
				UId:       m.UId,
				UVitamin:  m.UVitamin,
				Type:      consts.OPTION_DELETE,
				URole:     m.URole,
				URemark:   m.URemark,
				BwTimes:   m.BwTimes,
				PlayTimes: m.PlayTimes,
				Merge:     true,
			}).Error
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			if m.Partner == 1 {
				hpr.err = AHouse.deletePartnerRelevance(m.UId, tx)
				if hpr.err != nil {
					hpr.SetDBErr()
					return hpr
				}
			}

			if hpr.err = tx.Exec(`delete from house_member where hid = ? and uid = ?`, m.DHId, m.UId).Error; hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
		}

		if _, ok := BMysqlHouseMembersMap[m.UId]; !ok {
			joinMember := &models.HouseMember{
				DHId:      BHouse.DBClub.Id,
				UId:       m.UId,
				URole:     consts.ROLE_MEMBER,
				ApplyTime: &now,
				AgreeTime: &now,
			}
			if hpr.err = tx.Create(joinMember).Error; hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			if hpr.err = tx.Create(&models.HouseMemberLog{
				DHId:  joinMember.DHId,
				UId:   joinMember.UId,
				Type:  consts.OPTION_INSERT,
				URole: joinMember.URole,
				Merge: true,
			}).Error; hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			BMysqlHouseMembersMap[m.UId] = joinMember
		}
	}

	hpr.err = tx.Model(&models.HouseMergeLog{}).
		Where("swallowed = ?", BHouse.DBClub.Id).
		Where("devourer = ?", AHouse.DBClub.Id).
		Update("merge_state", models.HouseMergeStateInvalid, "merge_at", 0).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	hpr.err = tx.Model(&models.House{}).Where("id = ?", BHouse.DBClub.Id).Update("merge_hid", models.HouseMergeHidStateNormal).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}

	if len(otherMergeListOrder) <= 0 {
		// 更新状态
		hpr.err = tx.Model(&models.House{}).Where("id = ?", AHouse.DBClub.Id).Where("merge_hid = ?", models.HouseMergeHidStateDevourer).Update("merge_hid", models.HouseMergeHidStateNormal).Error
		if hpr.err != nil {
			hpr.SetDBErr()
			return hpr
		}
	}

	AHouseMembers = make([]*models.HouseMember, 0)
	hpr.err = tx.Find(&AHouseMembers, "hid = ?", AHouse.DBClub.Id).Error
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}
	AMysqlHouseMembersMap = make(models.HouseMembersMap)
	for _, m := range AHouseMembers {
		AMysqlHouseMembersMap[m.UId] = m
	}
	hpr.err = AHouse.SyncAllPartnerPyramid(tx, AMysqlHouseMembersMap)
	if hpr.err != nil {
		hpr.SetDBErr()
		return hpr
	}
	// 开始操作redis
	cli := GetDBMgr().GetDBrControl()
	ARedisHouseMember := GetDBMgr().GetHouseMemMap(AHouse.DBClub.Id)
	for uid, rdsMem := range ARedisHouseMember {
		if dbMem, ok := AMysqlHouseMembersMap[uid]; !ok {
			hpr.err = cli.HouseMemberDelete(rdsMem)
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			hpr.err = cli.MemberHouseJoinDelete(uid, AHouse.DBClub.Id)
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			AHouse.memDelete(uid)
		} else {
			var flush bool
			if rdsMem.Partner != dbMem.Partner {
				flush = true
				rdsMem.Partner = dbMem.Partner
			}
			if rdsMem.Superior != dbMem.Superior {
				flush = true
				rdsMem.Superior = dbMem.Superior
			}
			if rdsMem.Ref != dbMem.Ref {
				flush = true
				rdsMem.Ref = dbMem.Ref
			}
			if rdsMem.VicePartner != dbMem.VicePartner {
				flush = true
				rdsMem.VicePartner = dbMem.VicePartner
			}
			if rdsMem.VitaminAdmin != dbMem.VitaminAdmin {
				flush = true
				rdsMem.VitaminAdmin = dbMem.VitaminAdmin
			}
			if flush {
				hpr.err = cli.HouseMemberInsert(rdsMem)
				if hpr.err != nil {
					hpr.SetDBErr()
					return hpr
				}
			}
		}
	}

	BRedisHouseMember := GetDBMgr().GetHouseMemMap(BHouse.DBClub.Id)
	for uid, dbMem := range BMysqlHouseMembersMap {
		if _, ok := BRedisHouseMember[uid]; !ok {
			bRdsMem := dbMem.ConvertModel()
			if aRdsMem, ok2 := ARedisHouseMember[uid]; ok2 {
				bRdsMem.ImgUrl = aRdsMem.ImgUrl
				bRdsMem.NickName = aRdsMem.NickName
				bRdsMem.Sex = aRdsMem.Sex
			}
			hpr.err = cli.HouseMemberInsert(bRdsMem)
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
			hpr.err = cli.MemberHouseJoinInsert(uid, BHouse.DBClub.Id)
			if hpr.err != nil {
				hpr.SetDBErr()
				return hpr
			}
		}
	}

	return hpr
}

type HGroupUserDb struct {
	Hid     int   `gorm:"hid"`
	Puid    int64 `gorm:"puid"`
	GroupId int   `gorm:"group_id"`
	Uid     int64 `gorm:"uid"`
}

func (clm *ClubMgr) InitHouseGroupUser() {
	sql := `select hid,puid,group_id,uid from house_group_user where status = 0 `
	dest := []HGroupUserDb{}
	err := GetDBMgr().GetDBmControl().Raw(sql).Scan(&dest).Error
	if err != nil {
		panic(err)
	}
	for _, item := range dest {
		house := clm.GetClubHouseById(int64(item.Hid))
		if house == nil {
			continue
		}
		if house.UserGroup == nil {
			house.UserGroup = make(map[int64]map[int][]int64)
		}
		if house.UserGroup[item.Puid] == nil {
			house.UserGroup[item.Puid] = make(map[int][]int64)
		}
		if item.Uid == 0 { //分组
			house.UserGroup[item.Puid][item.GroupId] = make([]int64, 0)
			continue
		}
		slice := house.UserGroup[item.Puid][item.GroupId]
		if slice == nil {
			slice = make([]int64, 0)
			house.UserGroup[item.Puid][item.GroupId] = slice
		}
		house.UserGroup[item.Puid][item.GroupId] = append(house.UserGroup[item.Puid][item.GroupId], item.Uid)
	}
}

func (clm *ClubMgr) InitFiveGroup() {
	clm.lock.RLock()
	defer clm.lock.RUnlock()
	for _, h := range clm.ClubMap {
		go h.SureHouseFiveGroup()
	}
}

func (clm *ClubMgr) RemoveHouseTableWatch(tableId int, uid int64) {
	table := GetTableMgr().GetTable(tableId)
	if table == nil {
		return
	}
	house := GetClubMgr().GetClubHouseByHId(table.HId)
	if house == nil {
		return
	}
	floor := house.GetFloorByFId(table.FId)
	if floor == nil {
		return
	}
	hft := floor.GetHftByTid(tableId)
	if hft != nil {
		hft.RemoveWatcher(uid)
	}
}

func (clm *ClubMgr) CheckDefault(uid int64) {
	clm.initLock.Lock()
	defer clm.initLock.Unlock()
	defaultHouseId := GetServer().ConHouse.DefaultHouse
	if defaultHouseId < MIN_House_ID || defaultHouseId > MAX_House_ID {
		return
	}
	if clm.GetClubHouseByHId(defaultHouseId) != nil {
		clm.DefaultHouseMemberJoin(uid)
		return
	}
	_, xe := clm.DefaultHouseCreate(defaultHouseId, uid)
	if xe != nil {
		xlog.Logger().Errorf("create default house error: %v", xe)
	}
}

func (clm *ClubMgr) DefaultHouseCreate(HouseId int, uId int64) (*Club, *xerrors.XError) {
	//! 包厢
	house := new(Club)
	house.Init()
	house.DBClub = new(models.House)
	house.DBClub.HId = HouseId
	house.DBClub.UId = uId
	house.DBClub.IsVitamin = true
	house.DBClub.Area = consts.DefaultAreaCode
	//res := GetDBMgr().GetDBrControl().RedisV2.HGet(fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uId), "admin_game_on").Val()
	//adminOn, err := strconv.ParseBool(res)
	//if err != nil {
	//	adminOn = false
	//}
	house.DBClub.Name = "比赛"
	house.DBClub.Notify = ""
	house.DBClub.IsChecked = false
	house.DBClub.IsFrozen = false
	house.DBClub.AdminGameOn = true

	tx := GetDBMgr().GetDBmControl().Begin()
	id, err := GetDBMgr().HouseInsert(house, tx)
	if err != nil {
		xlog.Logger().Errorln(err)
		tx.Rollback()
		return nil, xerrors.DBExecError
	}
	// 内存
	house.DBClub.Id = id
	house.initPrize()
	//! 成员
	custerr := house.MemJoin(uId, consts.ROLE_CREATER, 0, true, tx)
	if custerr != nil {
		tx.Rollback()
		return nil, custerr
	}

	xerr := house.PoolChange(uId, models.HouseCreate, house.GetVitaminMax(), tx)
	if xerr != nil {
		xlog.Logger().Errorln(err.Error())
		tx.Rollback()
		return nil, xerr
	}
	tx.Commit()
	clm.lock.CustomLock()
	clm.ClubMap[house.DBClub.HId] = house
	clm.lock.CustomUnLock()
	clm.hidLock.CustomLock()
	clm.ClubIdToHidMap[house.DBClub.Id] = house.DBClub.HId
	clm.hidLock.CustomUnLock()
	// 更新包厢盟主数据
	person, _ := GetDBMgr().GetDBrControl().GetPerson(house.DBClub.UId)
	if person != nil {
		person.HouseId = house.DBClub.HId
		person.FloorId = 0
		person.TableId = 0
		GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "HouseId", person.HouseId, "FloorId", person.FloorId, "TableId", person.TableId)
	}
	return house, nil
}

func (clm *ClubMgr) DefaultHouseMemberJoin(uid int64) {
	defaultHouseId := GetServer().ConHouse.DefaultHouse
	if defaultHouseId < MIN_House_ID || defaultHouseId > MAX_House_ID {
		return
	}
	house := clm.GetClubHouseByHId(defaultHouseId)
	if house == nil {
		xlog.Logger().Errorf("defaultHouseMemberJoin get default house is nil: %v", defaultHouseId)
		return
	}
	hmem := house.GetMemByUId(uid)
	if hmem != nil {
		if hmem.Upper(consts.ROLE_APLLY) {
			return
		}
	}
	role := consts.ROLE_MEMBER
	if house.DBClub.UId == 0 {
		role = consts.ROLE_CREATER
	}
	custerr := house.MemJoin(uid, role, 0, true, nil)
	if custerr != nil {
		xlog.Logger().Errorf("defaultHouseMemberJoin got error: %v", custerr)
		return
	} else if role == consts.ROLE_CREATER {
		house.DBClub.UId = uid
		house.flush()
	}
}
