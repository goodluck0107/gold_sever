package center

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

//AllianceBiz 管理
type AllianceMgr struct {
	lock *lock2.RWMutex
}

var allianceSingleton *AllianceMgr = nil

func GetAllianceMgr() *AllianceMgr {
	if allianceSingleton == nil {
		allianceSingleton = new(AllianceMgr)
		allianceSingleton.lock = new(lock2.RWMutex)
	}
	return allianceSingleton
}

func (alm *AllianceMgr) CheckUserIn(leagueID, uid int64) bool {
	league := AllianceBiz{AllianceBizID: leagueID}
	return league.IsUserIn(uid)
}

// func (self *AllianceMgr) AddLeague(leagueID, areaCode int64) error {
// 	self.lock.CustomLock()
// 	defer self.lock.CustomUnLock()
// 	league := AllianceBiz{AllianceBizID: leagueID, AreaCode: areaCode}
// 	return league.AddLeague()
// }

// func (self *AllianceMgr) AddUser(leagueID, uid int64) error {
// 	self.lock.CustomLock()
// 	defer self.lock.CustomUnLock()
// 	league := AllianceBiz{AllianceBizID: leagueID}
// 	err := league.AddUser(uid)
// 	if err != nil {
// 		return err
// 	}
// 	return self.CardPoolNotify(uid, true)
// }

// func (self *AllianceMgr) FreezeLeagueUser(leagueID, uid int64) error {
// 	self.lock.CustomLock()
// 	defer self.lock.CustomUnLock()
// 	league := AllianceBiz{AllianceBizID: leagueID}
// 	err := league.FreezeUser(uid)
// 	if err != nil {
// 		return err
// 	}
// 	return self.CardPoolNotify(uid, false)
// }

// func (self *AllianceMgr) FreezeLeague(leagueID int64) error {
// 	self.lock.CustomLock()
// 	defer self.lock.CustomUnLock()
// 	league := AllianceBiz{AllianceBizID: leagueID}
// 	err := league.FreezeLeague()
// 	if err != nil {
// 		return err
// 	}
// 	uids := self.GetLeagueUserIDs(leagueID)
// 	for _, uid := range uids {
// 		self.CardPoolNotify(uid, false)
// 	}
// 	return nil
// }

// func (self *AllianceMgr) UnFreezeLeague(leagueID int64) error {
// 	self.lock.CustomLock()
// 	defer self.lock.CustomUnLock()
// 	league := AllianceBiz{AllianceBizID: leagueID}
// 	err := league.UnFreezeLeague()
// 	if err != nil {
// 		return err
// 	}
// 	uids := self.GetLeagueUserIDs(leagueID)
// 	for _, uid := range uids {
// 		self.CardPoolNotify(uid, true)
// 	}
// 	return nil
// }

func (alm *AllianceMgr) GetUserLeagueID(uid int64) int64 {
	userLeague := UserLeague{Uid: uid}
	return userLeague.GetUserLeagueID()
}

func (alm *AllianceMgr) GetUserLeagueInfo(uid int64) *AllianceBiz {
	userLeague := UserLeague{Uid: uid}
	return userLeague.GetUserLeagueInfo()
}

func (alm *AllianceMgr) UpdateLeagueInfo(id int64) error {
	league := AllianceBiz{}
	return league.UpdateFromDb(id)

}
func (alm *AllianceMgr) UpdateLeagueUser(id int64) error {
	userLeague := UserLeague{}
	return userLeague.UpdateFromDb(id)

}

func (alm *AllianceMgr) UpdateLeagueFreezeCard(leagueID int64, kacost int, add bool) error {
	league := AllianceBiz{AllianceBizID: leagueID}
	return league.UpdateFreezeCard(int64(kacost), add)

}

func (alm *AllianceMgr) GetLeagueUserIDs(leagueID int64) []int64 {
	league := AllianceBiz{AllianceBizID: leagueID}
	return league.GetUserIDs()

}
func (alm *AllianceMgr) LeagueCardPoolNotify(uid int64, info map[string]interface{}) error {
	items, err := GetDBMgr().ListHouseIdMemberJoin(uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, item := range items {
		house := GetClubMgr().GetClubHouseByHId(item.HId)
		if house == nil {
			continue
		}
		if house.DBClub.UId != uid {
			continue
		}
		canPool, _ := alm.CheckUserLeagueCardPool(uid)
		ss := make(map[string]interface{})
		ss["pool_state"] = canPool.CanPool
		ss["hid"] = item.HId
		house.Broadcast(consts.ROLE_CREATER, consts.MsgTypeHouseCardPoolChange, ss)
	}
	return nil
}

func (ab *AllianceBiz) LeaguePoolNotify(uid int64, state bool) error {
	ok, _ := GetAllianceMgr().CheckUserLeagueCardPool(uid)
	items, err := GetDBMgr().ListHouseIdMemberJoin(uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, item := range items {
		house := GetClubMgr().GetClubHouseByHId(item.HId)
		if house == nil {
			continue
		}
		if house.DBClub.UId != uid {
			continue
		}
		ss := make(map[string]interface{})
		ss["pool_state"] = ok.CanPool
		ss["hid"] = item.HId
		house.Broadcast(consts.ROLE_CREATER, consts.MsgTypeHouseCardPoolChange, ss)
	}
	return nil
}
func (alm *AllianceMgr) UserCardPoolNotify(uid int64, info map[string]interface{}) error {
	items, err := GetDBMgr().ListHouseIdMemberJoin(uid)
	if err != nil {
		xlog.Logger().Errorln(err)
		return err
	}
	for _, item := range items {
		house := GetClubMgr().GetClubHouseByHId(item.HId)
		if house == nil {
			continue
		}
		if house.DBClub.UId != uid {
			continue
		}
		canPool, _ := alm.CheckUserLeagueCardPool(house.DBClub.UId)
		ss := make(map[string]interface{})
		ss["pool_state"] = canPool.CanPool
		ss["hid"] = item.HId
		house.Broadcast(consts.ROLE_ADMIN, consts.MsgTypeHouseCardPoolChange, ss)
	}
	return nil
}

type PoolState struct {
	CanPool bool
	NotPool bool
}

func (alm *AllianceMgr) CheckUserLeagueCardPool(uid int64) (*PoolState, *AllianceBiz) {
	ul := &UserLeague{Uid: uid}
	err := ul.LoadRedis()
	if err != nil {
		return &PoolState{false, false}, nil
	}
	if ul.Freeze || !ul.PoolState || ul.PoolStart > time.Now().Unix() || ul.PoolEnd < time.Now().Unix() {
		return &PoolState{false, false}, nil
	}
	if ul.NotPool && ul.Card-ul.UsedCard-ul.FreezeCard <= 20 { //对盟主开启了房卡限制
		return &PoolState{false, false}, nil
	}
	lID := ul.GetUserLeagueID()
	l := &AllianceBiz{AllianceBizID: lID}
	l.LoadRedis()
	if l.Freeze || !l.PoolState || l.PoolStart > time.Now().Unix() || l.PoolEnd < time.Now().Unix() {
		return &PoolState{false, false}, nil
	}
	if l.Card-l.FreezeCard <= 20 { //低于20张卡片之后不允许扣卡池的卡
		return &PoolState{false, false}, nil
	}
	return &PoolState{true, ul.NotPool}, l
}

func (alm *AllianceMgr) GetAgentLeagueArea() ([]string, error) {
	agentAreas := make([]string, 0)
	keys, err := GetDBMgr().Redis.Keys("league_??????").Result()
	if err != nil {
		return agentAreas, err
	}
	for _, key := range keys {
		result, err := GetDBMgr().Redis.HGetAll(key).Result()
		if err != nil {
			xlog.Logger().Errorln("GetAgentLeagueArea.error:", key, err)
			continue
		}
		area_code := static.HF_Atoi64(result["area_code"])
		if area_code > 0 {
			agentAreas = append(agentAreas, fmt.Sprint(area_code))
		}
	}
	return agentAreas, nil
}

func (alm *AllianceMgr) UpdateUserCard(uid, usedCard, freezeCard int64) {
	ul := &UserLeague{Uid: uid}
	err := ul.LoadRedis()
	if err != nil {
		xlog.Logger().Errorf("upd league user not exists:uid:%d", uid)
		return
	}
	beforePool, _ := alm.CheckUserLeagueCardPool(uid)
	cli := GetDBMgr().Redis
	if usedCard != 0 {
		cli.HSet(ul.redisKey(), "used_card", ul.UsedCard+usedCard)

	}
	if freezeCard != 0 {
		cli.HSet(ul.redisKey(), "freeze_card", ul.FreezeCard+freezeCard)
	}
	sql := `update league_user set used_card = used_card + ? ,freeze_card = freeze_card +? where league_id = ? and uid = ?`
	if err = GetDBMgr().GetDBmControl().Exec(sql, usedCard, freezeCard, ul.LeagueID, uid).Error; err != nil {
		xlog.Logger().Errorf("error:%+v", err)
	}
	aftPool, _ := alm.CheckUserLeagueCardPool(uid)
	if beforePool.CanPool != aftPool.CanPool {
		alm.UserCardPoolNotify(uid, nil)
	}

	return
}
