package center

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"strconv"
)

type UserLeague struct {
	Uid        int64 `json:"uid"`
	LeagueID   int64 `json:"league_id"`
	Freeze     bool  `json:"freeze"`
	PoolState  bool  `json:"pool_state"`
	PoolStart  int64 `json:"pool_start"`
	PoolEnd    int64 `json:"pool_end"`
	Card       int64 `json:"card"`
	UsedCard   int64 `json:"used_card"`
	FreezeCard int64 `json:"freeze_card"`
	NotPool    bool  `json:"not_pool"`
}

func (self *UserLeague) redisKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER_H, self.Uid)
}

func (self *UserLeague) GetUserLeagueID() int64 {
	self.LoadRedis()
	if self.Freeze {
		return 0
	}
	return self.LeagueID

}

func (self *UserLeague) GetUserLeagueInfo() *AllianceBiz {
	leagueID := self.GetUserLeagueID()
	if leagueID <= 0 {
		return nil
	}
	cli := GetDBMgr().Redis
	league := AllianceBiz{}
	key := fmt.Sprintf(consts.REDIS_KEY_LEAGUE, leagueID)
	leagueInfo := cli.HGetAll(key).Val()
	err := static.CoverMapToStruct(leagueInfo, &league)
	if err != nil || league.Freeze {
		return nil
	}
	return &league
}

func (self *UserLeague) UpdateFromDb(id int64) error {
	var leagueUser models.LeagueUser
	leagueUser.Id = id
	err := GetDBMgr().GetDBmControl().Model(models.LeagueUser{}).Where("id = ?", id).First(&leagueUser).Error
	if err != nil {
		xlog.Logger().Errorf("没有获取到该数据：%+v，%d", err, id)
		return nil
	}
	self.Uid = leagueUser.Uid
	cli := GetDBMgr().Redis
	oldLeagueStr := cli.HGet(self.redisKey(), "league_id").Val()
	oldLeague, err := strconv.ParseInt(oldLeagueStr, 10, 64)
	if err == nil {
		if oldLeague != leagueUser.LeagueID { //更换了加盟商
			err := cli.SRem(fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER, oldLeague), leagueUser.Uid).Err()
			if err != nil {
				xlog.Logger().Errorf("redis 保存数据错误：%+v", err)
			}
		}
	}
	hMap := make(map[string]interface{}, 6)
	hMap["league_id"] = leagueUser.LeagueID
	hMap["freeze"] = leagueUser.Freeze
	hMap["pool_state"] = leagueUser.PoolState
	hMap["pool_start"] = leagueUser.PoolStart.Unix()
	hMap["pool_end"] = leagueUser.PoolEnd.Unix()
	hMap["uid"] = leagueUser.Uid
	hMap["uid"] = leagueUser.Uid
	hMap["card"] = leagueUser.Card
	hMap["freeze_card"] = leagueUser.FreezeCard
	hMap["used_card"] = leagueUser.UsedCard
	hMap["not_pool"] = leagueUser.NotPool
	err = cli.HMSet(self.redisKey(), hMap).Err()
	if err != nil {
		xlog.Logger().Errorf("redis 保存数据错误：%+v", err)
		return err
	}
	err = cli.SAdd(fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER, leagueUser.LeagueID), leagueUser.Uid).Err()
	if err != nil {
		xlog.Logger().Errorf("redis 保存数据错误：%+v", err)
		return err
	}
	// 重读手机号
	var user models.User
	err = GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", leagueUser.Uid).First(&user).Error
	if err != nil {
		return err
	}
	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(leagueUser.Uid, "Tel", user.Tel)
	if err != nil {
		return err
	}
	if p := GetPlayerMgr().GetPlayer(leagueUser.Uid); p != nil {
		p.UpdTel(user.Tel)
	}
	return GetAllianceMgr().UserCardPoolNotify(self.Uid, hMap) // 通知包厢管理员
}

func (self *UserLeague) LoadRedis() error {
	cli := GetDBMgr().Redis
	hmap, err := cli.HGetAll(self.redisKey()).Result()
	if err != nil {
		return err
	}
	return static.CoverMapToStruct(hmap, self)
}

func (ul *UserLeague) NotifyCardPool(updCount int64) error { //通知用，不更新实际房卡
	ul.LoadRedis()
	if ul.Card-ul.UsedCard-updCount <= 20 && ul.Card-ul.FreezeCard-ul.UsedCard > 20 {
		ul.LeaguePoolNotify(true)
	}
	return nil
}

func (ul *UserLeague) LeaguePoolNotify(state bool) error {
	uid := ul.Uid
	ok, _ := GetAllianceMgr().CheckUserLeagueCardPool(uid)
	if ok.CanPool != state {
		return nil
	}
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
		ss["pool_state"] = state
		ss["hid"] = item.HId
		house.Broadcast(consts.ROLE_CREATER, consts.MsgTypeHouseCardPoolChange, ss)
	}
	return nil
}
