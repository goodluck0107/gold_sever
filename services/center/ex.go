package center

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"math"
	"strconv"
	"time"
)

func CheckUserIp(ip string, users []*static.TableUser) bool {
	if len(ip) < 8 {
		return false
	}
	for _, user := range users {
		if user == nil || user.Uid == 0 {
			continue
		}
		var uip string
		p := GetPlayerMgr().GetPlayer(user.Uid)
		if p == nil {
			redisP, _ := GetDBMgr().GetDBrControl().GetPerson(user.Uid)
			if redisP != nil {
				uip = redisP.Ip
			} else {
				continue
			}
		} else {
			uip = p.Info.Ip
		}
		if uip == ip {
			return false
		}
	}
	return true
}

func CheckUserIpById(ip string, uids []int64) error {
	if len(ip) < 8 {
		return fmt.Errorf("invalid ip %s", ip)
	}
	for _, uid := range uids {
		if uid == 0 {
			continue
		}
		p := GetPlayerMgr().GetPlayer(uid)
		var uip string
		if p == nil {
			redisP, _ := GetDBMgr().GetDBrControl().GetPerson(uid)
			if redisP != nil {
				uip = redisP.Ip
			} else {
				continue
			}
		} else {
			uip = p.Ip
		}
		if uip == ip {
			return fmt.Errorf("same ip: uid:%d ip1/ip2: %s/%s", uid, ip, uip)
		}
	}
	return nil
}

func GetUserDistence(longitude1, latitude1 float64, longitude2s, latitude2s string) int64 { //游戏服所用经纬度为string
	if longitude1 == -1 || latitude1 == -1 { //不判断距离信息
		return 9999999
	}
	longitude2, err := strconv.ParseFloat(longitude2s, 64)
	if err != nil {
		xlog.Logger().Errorf("user gps error:%v", err)
		return 9999999
	}

	latitude2, err := strconv.ParseFloat(latitude2s, 64)
	if err != nil {
		xlog.Logger().Errorf("user gps error:%v", err)
		return 9999999
	}

	radLat1 := Rad(latitude1)
	radLat2 := Rad(latitude2)
	a := radLat1 - radLat2
	b := Rad(longitude1) - Rad(longitude2)
	s := 2 * math.Asin(math.Sqrt(math.Pow(math.Sin(a/2), 2)+math.Cos(radLat1)*math.Cos(radLat2)*math.Pow(math.Sin(b/2), 2)))

	s = s * 6378137 //乘地球半径
	return int64(s)
}
func Rad(d float64) float64 {
	return d * math.Pi / 180.0
}

func CheckUserGps(house *Club, longitude1, latitude1 float64, users []int64) error {
	if longitude1 == -1 || latitude1 == -1 { //不判断距离信息
		return nil
	}
	limitDest := house.GetHouseTableLimitDistance()
	if limitDest == -1 {
		limitDest = 100
	}
	for _, uid := range users {
		if uid == 0 {
			continue
		}
		p, _ := GetDBMgr().GetDBrControl().GetPerson(uid)
		if p == nil {
			continue
		}
		d := GetUserDistence(longitude1, latitude1, p.Longitude, p.Latitude)
		if d <= int64(limitDest) {
			return fmt.Errorf("hypotelorism: uid:%d longitude1/latitude1: [%0.2f,%0.2f] longitude2/latitude2: [%s,%s] distance/limitDest:%d/%d",
				p.Uid,
				longitude1,
				latitude1,
				p.Longitude,
				p.Latitude,
				d,
				limitDest,
			)
		}
	}
	return nil
}

const UnreadRedisKey = "unread"

// 玩家未读通知
func UnreadNotifyRedisKey(uid int64) string {
	return fmt.Sprintf("%s:notify:%d", UnreadRedisKey, uid)
}

func AddUnreadNotify(uid int64, data string) error {
	err := GetDBMgr().GetDBrControl().RedisV2.LPush(UnreadNotifyRedisKey(uid), data).Err()
	if err != nil {
		GetDBMgr().GetDBrControl().RedisV2.Del(UnreadNotifyRedisKey(uid))
	}
	return err
}

func GetUnreadNotify(uid int64) ([]string, error) {
	res, err := GetDBMgr().GetDBrControl().RedisV2.LRange(UnreadNotifyRedisKey(uid), 0, -1).Result()
	if err != nil {
		GetDBMgr().GetDBrControl().RedisV2.Del(UnreadNotifyRedisKey(uid))
	}
	return res, err
}

func DelUnreadNotify(uid int64) error {
	return GetDBMgr().GetDBrControl().RedisV2.Del(UnreadNotifyRedisKey(uid)).Err()
}

// 玩家包厢邀请未处理通知
func UnreadHouseJoinInviteRedisKey(uid int64) string {
	return fmt.Sprintf("%s:houseinvite:%d", UnreadRedisKey, uid)
}

func AddUnreadHouseJoinInvite(uid int64, v *static.Msg_HC_HouseInviteJoin) error {
	// 缓存24小时邀请数据，24小时不上线处理 则视为不接受邀请
	// return GetDBMgr().GetDBrControl().RedisV2.Set(UnreadHouseJoinInviteRedisKey(uid), v, time.Hour*24).Err()
	return GetDBMgr().GetDBrControl().RedisV2.Set(UnreadHouseJoinInviteRedisKey(uid), v, 0).Err()
}

func GetUnreadHouseJoinInvite(uid int64) (*static.Msg_HC_HouseInviteJoin, error) {
	v := new(static.Msg_HC_HouseInviteJoin)
	err := GetDBMgr().GetDBrControl().RedisV2.Get(UnreadHouseJoinInviteRedisKey(uid)).Scan(v)
	return v, err
}

func DelUnreadHouseJoinInvite(uid int64) error {
	return GetDBMgr().GetDBrControl().RedisV2.Del(UnreadHouseJoinInviteRedisKey(uid)).Err()
}

func HouseInviteBlackRedisKey(uid int64) string {
	return fmt.Sprintf(consts.REDIS_KEY_HOUSE_BLACK, uid)
}

func AddHouseInviteBlack(uid, tuid int64) error {
	return GetDBMgr().GetDBrControl().RedisV2.SAdd(HouseInviteBlackRedisKey(uid), tuid).Err()
}

func IsHouseInviteBlack(uid, tuid int64) bool {
	return GetDBMgr().GetDBrControl().RedisV2.SIsMember(HouseInviteBlackRedisKey(uid), tuid).Val()
}

// 发放每日礼包奖励
func disposeUserDailyRewards(person *static.Person) error {
	tx := GetDBMgr().db_M.Begin()
	_, aftgold, err := wealthtalk.UpdateGold(person.Uid, GetServer().ConServers.DailyRewards.AwardGold, models.CostTypeDailyReward, tx)
	if err != nil {
		tx.Rollback()
		xlog.Logger().Error("发放每日礼包出错: ", err)
		return err
	}

	// 添加领取每日礼包记录
	record := new(models.UserDailyRewards)
	record.Uid = person.Uid
	record.AwardGold = GetServer().ConServers.DailyRewards.AwardGold
	record.Date = time.Now()
	if err = tx.Create(&record).Error; err != nil {
		tx.Rollback()
		xlog.Logger().Error("发放每日礼包出错: ", err)
		return err
	}

	// 更新内存
	person.Gold = aftgold
	// 更新redis
	if err = GetDBMgr().db_R.UpdatePersonAttrs(person.Uid, "Gold", person.Gold); err != nil {
		tx.Rollback()
		xlog.Logger().Error("发放每日礼包出错: ", err)
		return err
	}
	tx.Commit()
	return nil
}

// 发放低保
func disposeUserAllowances(person *static.Person) error {
	tx := GetDBMgr().db_M.Begin()
	_, aftgold, err := wealthtalk.UpdateGold(person.Uid, GetServer().ConServers.Allowances.AwardGold, models.CostTypeAllowances, tx)
	if err != nil {
		tx.Rollback()
		xlog.Logger().Error("dispose user allowance: ", err)
		return err
	}

	// 添加低保记录
	record := new(models.UserAllowances)
	record.Uid = person.Uid
	record.AwardGold = GetServer().ConServers.Allowances.AwardGold
	record.Date = time.Now()
	if err = tx.Create(&record).Error; err != nil {
		tx.Rollback()
		xlog.Logger().Error("dispose user allowance: ", err)
		return err
	}

	// 更新内存
	person.Gold = aftgold
	// 更新redis
	if err = GetDBMgr().db_R.UpdatePersonAttrs(person.Uid, "Gold", person.Gold); err != nil {
		tx.Rollback()
		xlog.Logger().Error("dispose user allowance: ", err)
		return err
	}
	tx.Commit()
	return nil
}

// 重置用户的缓存数据 GameId/SiteId/TableId
func ResetUserCacheData(p *static.Person) {
	if p == nil {
		return
	}
	updates := make(map[string]interface{})
	if p.GameId != 0 {
		updates["GameId"] = 0
	}
	if p.SiteId != 0 {
		updates["SiteId"] = 0
	}
	if p.TableId != 0 {
		updates["TableId"] = 0
	}
	if len(updates) > 0 {
		err := GetDBMgr().GetDBrControl().UpdatePersonAttrsV2(p.Uid, updates)
		if err != nil {
			xlog.Logger().Error("ResetUserCacheData error:%s", err)
		} else {
			p.GameId = 0
			p.TableId = 0
			p.SiteId = 0
			if person := GetPlayerMgr().GetPlayer(p.Uid); person != nil {
				person.Info.GameId = 0
				person.Info.TableId = 0
				person.Info.SiteId = 0
			}
		}
	}
}
