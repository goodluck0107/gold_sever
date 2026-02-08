package center

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"strconv"
	"time"
)

type AllianceBiz struct {
	AllianceBizID int64 `json:"league_id"`
	Card          int64 `json:"card_num"`
	Freeze        bool  `json:"freeze"`
	UserNum       int64 `json:"user_num"`
	AreaCode      int64 `json:"area_code"`
	FreezeCard    int64 `json:"freeze_card"`
	PoolState     bool  `json:"pool_state"`
	PoolStart     int64 `json:"pool_start"`
	PoolEnd       int64 `json:"pool_end"`
	Lock          *lock2.RWMutex
}

func (ab *AllianceBiz) AddLeague() error {
	if ab.AllianceBizID <= 0 {
		return errors.New("params error")
	}
	sql := `insert into league(league_id,area_code,card) values(?,?,?)`
	if err := GetDBMgr().GetDBmControl().Exec(sql, ab.AllianceBizID, ab.AreaCode, ab.Card).Error; err != nil {
		return err
	}
	legaueField := make(map[string]interface{}, 6)
	legaueField[ab.CardKey()] = ab.Card
	legaueField[ab.FreezeKey()] = false
	legaueField[ab.UserNumKey()] = 0
	legaueField[ab.LeagueIDKey()] = ab.AllianceBizID
	legaueField[ab.AreaCodeKey()] = ab.AreaCode
	legaueField[ab.FreezeCardKey()] = 0
	cli := GetDBMgr().Redis
	return cli.HMSet(ab.redisKey(), legaueField).Err()
}

func (ab *AllianceBiz) FreezeLeague() error {
	if ab.AllianceBizID <= 0 {
		return errors.New("params error")
	}
	cli := GetDBMgr().Redis
	return cli.HSet(ab.redisKey(), ab.FreezeKey(), true).Err()
}
func (ab *AllianceBiz) UnFreezeLeague() error {
	if ab.AllianceBizID <= 0 {
		return errors.New("params error")
	}
	cli := GetDBMgr().Redis
	return cli.HSet(ab.redisKey(), ab.FreezeKey(), false).Err()
}

func (ab *AllianceBiz) AddUser(uid int64) error {
	if ab.IsUserIn(uid) {
		return errors.New("user already in")
	}
	cli := GetDBMgr().Redis
	pipe := cli.TxPipeline()
	pipe.HIncrBy(ab.redisKey(), ab.UserNumKey(), 1)
	pipe.SAdd(ab.userRedisKey(), uid)
	userLeague := UserLeague{Uid: uid}
	pipe.SAdd(userLeague.redisKey(), ab.AllianceBizID)
	_, err := pipe.Exec()
	return err
}

func (ab *AllianceBiz) FreezeUser(uid int64) error {
	if !ab.IsUserIn(uid) {
		return errors.New("user already freeze or not in league")
	}
	cli := GetDBMgr().Redis
	pipe := cli.TxPipeline()
	pipe.HIncrBy(ab.redisKey(), ab.UserNumKey(), -1)
	pipe.SRem(ab.userRedisKey(), uid)
	userLeague := UserLeague{Uid: uid}
	pipe.SRem(userLeague.redisKey(), ab.AllianceBizID)
	_, err := pipe.Exec()
	return err
}

func (ab *AllianceBiz) AddCard(count int64) error {
	cli := GetDBMgr().Redis
	return cli.HIncrBy(ab.redisKey(), ab.UserNumKey(), count).Err()
}

func (ab *AllianceBiz) IsUserIn(uid int64) bool {
	cli := GetDBMgr().Redis
	return cli.SIsMember(ab.userRedisKey(), uid).Val()
}
func (ab *AllianceBiz) UpdateFreezeCard(count int64, add bool) error {
	cli := GetDBMgr().Redis
	ab.LoadRedis()
	if !add {
		count = 0 - count
		cli.HIncrBy(ab.redisKey(), ab.FreezeCardKey(), count).Err()
		if ab.Card-ab.FreezeCard <= 20 && ab.Card-ab.FreezeCard-count > 20 {
			uids := ab.GetUserIDs()
			for _, uid := range uids {
				ab.LeaguePoolNotify(uid, true)
			}
		}
	} else {
		cli.HIncrBy(ab.redisKey(), ab.FreezeCardKey(), count).Err()
		if ab.Card-ab.FreezeCard > 20 && ab.Card-ab.FreezeCard-count <= 20 {
			uids := ab.GetUserIDs()
			for _, uid := range uids {
				ab.LeaguePoolNotify(uid, false)
			}
		}
	}
	return nil
}

// 自运营判断
func (ab *AllianceBiz) IsOfficialWork() bool {
	return ab.IsExist() && ab.AreaCode == consts.OfficialWorkLeagueAreaCode
}

func (ab *AllianceBiz) IsExist() bool {
	return ab.AllianceBizID > 0
}

// 是否支持该区域
func (ab *AllianceBiz) IsSupportArea(area_code int64) bool {
	if !ab.IsExist() {
		return false
	}
	if ab.IsOfficialWork() {
		// 检查这个area_code在不在非代理区域列表里面
		return IsOfficialWorkArea(static.HF_I64toa(area_code))
	}
	return ab.AreaCode == area_code
}

func (ab *AllianceBiz) GetUserIDs() []int64 {
	cli := GetDBMgr().Redis
	res := cli.SMembers(ab.userRedisKey()).Val()
	ids := make([]int64, 0, len(res))
	for _, item := range res {
		id, err := strconv.ParseInt(item, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids

}
func (ab *AllianceBiz) redisKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_LEAGUE, ab.AllianceBizID)
}

func (ab *AllianceBiz) userRedisKey() string {
	return fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER, ab.AllianceBizID)
}

func (ab *AllianceBiz) LeagueIDKey() string {
	return "league_id"
}
func (ab *AllianceBiz) FreezeKey() string {
	return "freeze"
}
func (ab *AllianceBiz) PoolStateKey() string {
	return "pool_state"
}
func (ab *AllianceBiz) PoolStartKey() string {
	return "pool_start"
}
func (ab *AllianceBiz) PoolEndKey() string {
	return "pool_end"
}
func (ab *AllianceBiz) FreezeCardKey() string {
	return "freeze_card"
}
func (ab *AllianceBiz) CardKey() string {
	return "card_num"
}
func (ab *AllianceBiz) UserNumKey() string {
	return "user_num"
}
func (ab *AllianceBiz) AreaCodeKey() string {
	return "area_code"
}

func (ab *AllianceBiz) UpdateFromDb(id int64) error {
	var league models.League
	league.Id = id
	err := GetDBMgr().GetDBmControl().Model(models.League{}).Where("id = ?", id).First(&league).Error
	if err != nil {
		fmt.Println("league error ", err)
		return err
	}
	legaueField := make(map[string]interface{}, 9)
	legaueField[ab.CardKey()] = league.Card
	legaueField[ab.FreezeKey()] = league.Freeze
	legaueField[ab.UserNumKey()] = league.UserNum
	legaueField[ab.LeagueIDKey()] = league.LeagueID
	legaueField[ab.AreaCodeKey()] = league.AreaCode
	legaueField[ab.FreezeCardKey()] = league.FreezeCard
	legaueField[ab.PoolStateKey()] = league.PoolState
	legaueField[ab.PoolStartKey()] = league.PoolStart.Unix()
	legaueField[ab.PoolEndKey()] = league.PoolEnd.Unix()
	ab.AllianceBizID = league.LeagueID
	cli := GetDBMgr().Redis
	err = cli.HMSet(ab.redisKey(), legaueField).Err()
	if err == nil {
		uids := ab.GetUserIDs()
		for _, uid := range uids {
			GetAllianceMgr().LeagueCardPoolNotify(uid, legaueField)
		}
	}
	return err
}

func (ab *AllianceBiz) LoadRedis() error {
	cli := GetDBMgr().Redis
	hMap, err := cli.HGetAll(ab.redisKey()).Result()
	if err != nil {
		return err
	}
	return static.CoverMapToStruct(hMap, ab)
}

//SetNx 设置锁
func (ab *AllianceBiz) SetNx() bool {
	cli := GetDBMgr().Redis
	return cli.SetNX(fmt.Sprintf("lock_l_%d", ab.AllianceBizID), 1, 5*time.Second).Val()
}

//DelLock 删除锁
func (ab *AllianceBiz) DelLockWithLog() {
	cli := GetDBMgr().Redis
	cli.Del(fmt.Sprintf("lock_l_%d", ab.AllianceBizID))
}

func (ab *AllianceBiz) NotifyCardPool(updCount int64) error { //通知用，不更新实际房卡
	ab.LoadRedis()
	if ab.Card-ab.FreezeCard-updCount <= 20 && ab.Card-ab.FreezeCard > 20 {
		uids := ab.GetUserIDs()
		for _, uid := range uids {
			ab.LeaguePoolNotify(uid, true)
		}
	}
	return nil
}
