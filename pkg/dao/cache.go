package dao

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"

	goRedis "github.com/go-redis/redis"
)

//! Redis操作
type DB_r struct {
	// Deprecated: Use RedisV2 instead.
	m_redis *static.RedisCli //! redis
	RedisV2 *goRedis.Client
}

//! 连接Redis
func (rds *DB_r) Init(ip string, db int, auth string) error {
	rds.m_redis = static.InitRedis(ip, db, auth)
	if rds.m_redis == nil {
		return errors.New("redis init error")
	}
	rds.RedisV2 = static.InitRedisV2(ip, db, auth)
	return nil
}

func (rds *DB_r) Close() error {
	return rds.RedisV2.Close()
}

func (rds *DB_r) Pool() *static.RedisCli {
	return rds.m_redis
}

// Deprecated: Use RedisV2 instead.
func (rds *DB_r) GetRedisCli() *static.RedisCli {
	return rds.m_redis
}

//! 通用set
// Deprecated: Use RedisV2 instead.
func (rds *DB_r) Set(key string, val []byte) error {
	return rds.RedisV2.Set(key, val, 0).Err()
}

// Deprecated: Use RedisV2 instead.
//! 通用get
func (rds *DB_r) Get(key string) ([]byte, error) {
	return rds.m_redis.Get(key)
}

// Deprecated: Use RedisV2 instead.
//! 通用get
func (rds *DB_r) HGet(key string, field string) ([]byte, error) {
	return rds.m_redis.HGet(key, field)
}

// Deprecated: Use RedisV2 instead.
//! 通用get
func (rds *DB_r) HSet(key string, field string, val []byte) error {
	return rds.m_redis.HSet(key, field, val)
}

// Deprecated: Use RedisV2 instead.
//! 通用del
func (rds *DB_r) Remove(key string) (bool, error) {
	return rds.m_redis.Del(key)
}

// Deprecated: Use RedisV2 instead.
//! 设置过期时间
func (rds *DB_r) Expire(key string, sec int64) error {
	return rds.m_redis.Expire(key, sec)
}

// Deprecated: Use RedisV2 instead.
//! 判断key存不存在
func (rds *DB_r) Exists(key string) bool {
	return rds.m_redis.Exists(key)
}

// Deprecated: Use RedisV2 instead.
//! 模糊查找key
func (rds *DB_r) Keys(key string) ([]string, error) {
	data, err := rds.m_redis.Keys(key)
	if err != nil {
		xlog.Logger().Errorln(err)
		return nil, err
	}

	result := make([]string, 0)
	for _, item := range data {
		result = append(result, string(item))
	}

	return result, nil
}

// // redis锁
// // 依次传入 Key ，Value， expiry自动解锁时间（过期时间），tries锁尝试次数，tryInterval尝试间隔
// func (self *DB_r) Lock(key string, value interface{}, expiry time.Duration, tries int, tryInterval time.Duration) error {
// 	// 至少尝试一次
// 	if tries < 1 {
// 		tries = 1
// 	}
// 	for i := 0; i < tries; i++ {
// 		if err := self.RedisV2.SetNX(key, value, expiry).Err(); err == nil {
// 			return nil
// 		}
// 		time.Sleep(tryInterval)
// 	}
// 	err := fmt.Errorf("redis lock failed after %0.2f ms", float64(tries)*float64(tryInterval.Nanoseconds()/1e4/100.00))
// 	syslog.Logger().Error(err)
// 	return err
// }
//
// func (self *DB_r) UnLock(key string) error {
// 	return self.RedisV2.Del(key).Err()
// }

////////////////////////////////////////////////////////////////////////////////
// 包厢相关
func (rds *DB_r) HouseBaseReLoad() ([]*models.House, error) {
	res := rds.RedisV2.Keys(consts.REDIS_KEY_HOUSE_INFO_ALL).Val()
	lenRes := len(res)
	if lenRes == 0 {
		return nil, nil
	}

	datas := make([]*models.House, 0, lenRes)
	for _, key := range res {
		tmpKey := key
		data := rds.GetHouse(tmpKey)
		if data != nil {
			datas = append(datas, data)
		}
	}
	return datas, nil
}

func (rds *DB_r) GetHouse(key string) *models.House {
	data := models.House{}
	buf, err := rds.RedisV2.Get(key).Result()
	if err != nil {
		xlog.Logger().Errorf("house init error:%v,key:%s", err, key)
		return nil
	}
	err = json.Unmarshal([]byte(buf), &data)
	if err != nil {
		xlog.Logger().Errorf("json house error:%v", err)
		return nil
	}

	return &data
}

func (rds *DB_r) hotVersionKey() string {
	return "ServerHotVersionRecord"
}

func Md5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}

func (rds *DB_r) hotVersionMember(v string) string {
	timer := "20170101_12:00:00"
	m := md5.New()
	m.Write([]byte(fmt.Sprintf("%s_%s", v, timer)))
	return hex.EncodeToString(m.Sum(nil))
}

func (rds *DB_r) HotVersionRecord(v string) error {
	return rds.RedisV2.SAdd(rds.hotVersionKey(), rds.hotVersionMember(v)).Err()
}

func (rds *DB_r) IsHotVersionSupport(v string) bool {
	return rds.RedisV2.SIsMember(rds.hotVersionKey(), rds.hotVersionMember(v)).Val()
}

func (rds *DB_r) HouseBaseReSave(datas []*models.House) error {

	for _, data := range datas {
		// key
		key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, data.Id)
		if rds.RedisV2.Exists(key).Val() == 1 {
			continue
		}
		rds.HouseInsert(data)
	}

	return nil
}

//! 包厢数据
func (rds *DB_r) GetHouseInfoById(id int64) (*models.House, error) {

	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, id)

	// res
	res, err := rds.RedisV2.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	db_hosue := new(models.House)
	err = json.Unmarshal(res, db_hosue)
	if err != nil {
		return nil, err
	}

	return db_hosue, nil
}

//! 创建包厢
func (rds *DB_r) HouseInsert(db_house *models.House) error {

	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, db_house.Id)
	// val
	val := static.HF_JtoB(db_house)

	// "house_id"
	return rds.RedisV2.Set(key, val, 0).Err()
}

//! 删除包厢
func (rds *DB_r) HouseDelete(db_house *models.House) error {

	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, db_house.Id)

	// res
	_, err := rds.RedisV2.Del(key).Result()

	return err
}

//! 创建包厢ID列表
func (rds *DB_r) ListHouseMemberCreate(uid int64) ([]int64, error) {

	// lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)
	// lval
	lval := rds.RedisV2.LRange(lkey, 0, -1).Val()
	var datas []int64

	for _, val := range lval {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			rds.RedisV2.LRem(lkey, 1, val) //非整形key，移除
			continue
		}
		house, err := rds.GetHouseInfoById(id)
		if err != nil {
			rds.RedisV2.LRem(lkey, 1, val) //house 不存在
			xlog.Logger().Errorf("remove user house:%s,%d", val, uid)
			continue
		}
		if house.UId == uid {
			datas = append(datas, id)
		}
	}

	return datas, nil
}

//! 入驻包厢ID列表
func (rds *DB_r) ListHouseMemberJoin(uid int64) ([]int64, error) {

	// lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)
	// lval
	lval := rds.RedisV2.LRange(lkey, 0, -1).Val()
	var datas []int64

	for _, val := range lval {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			rds.RedisV2.LRem(lkey, 1, val) //非整形key，移除
			continue
		}
		house, err := rds.GetHouseInfoById(id)
		if err != nil {
			rds.RedisV2.LRem(lkey, 1, val) //house 不存在
			xlog.Logger().Errorf("remove user house:%s,%d", val, uid)
			continue
		}
		if house.UId != uid {
			datas = append(datas, id)
		}
	}

	return datas, nil
}

//! 入驻包厢数量 不包含创建数
func (rds *DB_r) HouseMemberJoinCounts(uid int64) (int, error) {

	// lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)

	// lval
	lval, err := rds.m_redis.Lrange(lkey, 0, -1)
	if err != nil {
		return 0, err
	}

	sum := 0
	for _, val := range lval {
		house, err := rds.GetHouseInfoById(static.HF_Bytestoi64(val))
		if err != nil {
			return 0, err
		}
		if house.UId != uid {
			sum++
		}
	}

	return sum, nil
}

//! 创建包厢数量
func (rds *DB_r) HouseMemberCreateCounts(uid int64) (int, error) {

	// lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)

	// lval
	lval, err := rds.m_redis.Lrange(lkey, 0, -1)
	if err != nil {
		return 0, err
	}

	sum := 0
	for _, val := range lval {
		house, err := rds.GetHouseInfoById(static.HF_Bytestoi64(val))
		if err != nil {
			return 0, err
		}
		if house.UId == uid {
			sum++
		}
	}

	return sum, nil
}

//! 包厢名称公告修改
func (rds *DB_r) HouseBaseNNModify(hid int64, name string, notify string) error {

	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, hid)
	// val
	val, err := rds.RedisV2.Get(key).Bytes()
	if err != nil {
		return err
	}

	var dhouse models.House
	err = json.Unmarshal(val, &dhouse)
	if err != nil {
		return err
	}

	dhouse.Name = name
	dhouse.Notify = notify

	datas, err := json.Marshal(dhouse)
	if err != nil {
		return err
	}

	// new val
	err = rds.m_redis.Set(key, datas)
	if err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// 包厢楼层相关
func (rds *DB_r) HouseFloorBaseReLoad() ([]*static.HouseFloor, error) {

	res, err := rds.RedisV2.Keys(consts.REDIS_KEY_HOUSE_FLOOR_ALL).Result()
	if err != nil {
		return nil, err
	}
	var datas []*static.HouseFloor
	for _, key := range res {
		tmpKey := key
		hfs := rds.GetHouseFloor(tmpKey)
		if len(hfs) > 0 {
			datas = append(datas, hfs...)
		}
	}

	return datas, nil
}

func (rds *DB_r) GetHouseFloor(key string) (datas []*static.HouseFloor) {
	buf, err := rds.RedisV2.HGetAll(key).Result()
	if err != nil {
		xlog.Logger().Errorf("floor init:%v", err)
		return datas
	}
	for _, v := range buf {
		data := static.HouseFloor{}
		err := json.Unmarshal([]byte(v), &data)
		if err != nil {
			xlog.Logger().Errorf("json house error:%v", err)
			continue
		}
		datas = append(datas, &data)
	}
	return
}

func (rds *DB_r) GetHouseMember(hid int64) ([]*static.HouseMember, error) {
	mems := make([]*static.HouseMember, 0)
	err := rds.RedisV2.HVals(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, hid)).ScanSlice(mems)
	if err != nil {
		return mems, err
	}
	return mems, nil
}

func (rds *DB_r) HouseFloorBaseReSave(datas []*models.HouseFloor) error {

	for _, data := range datas {
		// key
		key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, data.DHId)
		// field
		field := fmt.Sprintf("%d", data.Id)

		if rds.m_redis.HExists(key, field) {
			continue
		}

		rds.HouseFloorInsert(data)
	}

	return nil
}

//! 创建包厢楼层
func (rds *DB_r) HouseFloorInsert(db_housefloor *models.HouseFloor) error {

	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, db_housefloor.DHId)
	// key
	key := fmt.Sprintf("%d", db_housefloor.Id)
	// val
	val := static.HF_JtoB(db_housefloor)

	// "housefloor_id"
	return rds.m_redis.HSet(hkey, key, val)
}

//! 删除包厢楼层
func (rds *DB_r) HouseFloorDelete(dhid int64, fid int64) error {

	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, dhid)
	// key
	key := fmt.Sprintf("%d", fid)

	return rds.m_redis.HDel(hkey, key)
}

func (rds *DB_r) HouseFloorSelect(dhid int64, fid int64) (*static.HouseFloor, error) {

	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR, dhid)
	// key
	key := fmt.Sprintf("%d", fid)

	data, err := rds.m_redis.HGet(hkey, key)

	if err != nil {
		return nil, err
	}

	floor := new(static.HouseFloor)

	err = json.Unmarshal(data, floor)
	if err != nil {
		return nil, err
	}

	return floor, nil
}

func (rds *DB_r) FloorTableBaseReLoad() ([]interface{}, error) {

	bytes, err := rds.m_redis.Keys(consts.REDIS_KEY_FLOOR_TABLE_ALL)
	if err != nil {
		return nil, err
	}

	var items []interface{}
	for _, key := range bytes {

		datas, err := rds.m_redis.Hvals(static.HF_Bytestoa(key))
		if err != nil {
			return nil, err
		}

		for _, data := range datas {
			var dhfloor static.FloorTable
			err = json.Unmarshal(data, &dhfloor)
			if err != nil {
				return nil, err
			}
			items = append(items, &dhfloor)
		}
	}

	return items, nil
}

func (rds *DB_r) HouseActivityReLoad() ([]interface{}, error) {

	bytes, err := rds.m_redis.Keys(consts.REDIS_KEY_HOUSE_ACTIVITY_ALL)
	if err != nil {
		return nil, err
	}

	var items []interface{}
	for _, key := range bytes {

		datas, err := rds.m_redis.Hvals(static.HF_Bytestoa(key))
		if err != nil {
			return nil, err
		}

		for _, data := range datas {
			var dhActivity static.HouseActivity
			err = json.Unmarshal(data, &dhActivity)
			if err != nil {
				return nil, err
			}
			items = append(items, &dhActivity)
		}
	}
	return items, nil
}

func (rds *DB_r) HouseActivityReSave(datas []*models.HouseActivity) error {

	for _, data := range datas {
		hact := data.ConvertModel()
		// key
		key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_ACTIVITY, hact.DHId)
		// field
		field := fmt.Sprintf("%d", data.Id)
		if rds.m_redis.HExists(key, field) {
			continue
		}
		rds.HouseActivityInsert(data.ConvertModel())
	}

	return nil
}

//! 创建包厢楼层活动
func (rds *DB_r) HouseActivityInsert(db_hActivity *static.HouseActivity) error {

	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_ACTIVITY, db_hActivity.DHId)
	// key
	key := fmt.Sprintf("%d", db_hActivity.Id)
	// val
	val := static.HF_JtoB(db_hActivity)

	// "houseactivity_id"
	return rds.m_redis.HSet(hkey, key, val)
}

//! 删除包厢活动
func (rds *DB_r) HouseActivityDelete(hid int64, actid int64) error {

	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_ACTIVITY, hid)
	// key
	key := fmt.Sprintf("%d", actid)

	return rds.m_redis.HDel(hkey, key)
}

//! 不活跃包厢活动
func (rds *DB_r) HouseActivityUnActive(hid int64, actid int64) error {

	hact, err := rds.HouseActivityInfo(hid, actid)
	if err != nil {
		return err
	}

	hact.Status = 0

	return rds.HouseActivityInsert(hact)
}

//! 查询包厢活动
func (rds *DB_r) HouseActivityInfo(hid int64, actid int64) (*static.HouseActivity, error) {
	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_ACTIVITY, hid)
	// key
	key := fmt.Sprintf("%d", actid)
	// res
	bytes, err := rds.m_redis.HGet(hkey, key)
	if err != nil {
		return nil, err
	}

	// instance
	var data static.HouseActivity
	if bytes != nil {
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			return nil, err
		}
		if data.Status == 0 {
			return nil, nil
		}
	}
	return &data, nil
}

//! 查询包厢楼层活动
func (rds *DB_r) HouseActivityList(dhid int64, passEnd bool) ([]*static.HouseActivity, error) {
	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_ACTIVITY, dhid)
	datas, err := rds.m_redis.Hvals(hkey)
	if err != nil {
		return nil, err
	}

	var items []*static.HouseActivity
	for _, data := range datas {
		var dhActivity static.HouseActivity
		err = json.Unmarshal(data, &dhActivity)
		if err != nil {
			return nil, err
		}
		if dhActivity.Status == 0 {
			continue
		}
		if passEnd {
			if dhActivity.EndTime < time.Now().Unix() && dhActivity.Type == 1 {
				continue
			}
		}
		items = append(items, &dhActivity)
	}

	return items, nil
}

//! 包厢楼层活动数据
func (rds *DB_r) HouseActivityRecordReLoad() ([]interface{}, error) {
	bytes, err := rds.m_redis.Keys(consts.REDIS_KEY_HOUSE_RECORD_ALL)
	if err != nil {
		return nil, err
	}

	var items []interface{}
	for _, key := range bytes {

		stractid := static.HF_Bytestoa(key[len(consts.REDIS_KEY_HOUSE_RECORD_TAG):])
		actid, err := strconv.ParseInt(stractid, 10, 64)
		if err != nil {
			continue
		}

		datas, err := rds.m_redis.ZrevrangeWithScore(static.HF_Bytestoa(key), 0, -1)
		if err != nil {
			return nil, err
		}

		var uid []int64
		var res []float64
		for index, data := range datas {
			if index%2 == 0 {
				uid = append(uid, static.HF_Bytestoi64(data))
			} else {
				res = append(res, static.HF_Bytestof64(data))
			}
		}

		var items []*static.ActRecordItem
		for i := 0; i < len(uid); i++ {
			item := new(static.ActRecordItem)
			item.ActId = actid
			item.UId = uid[i]
			item.Score = res[i]
			items = append(items, item)
		}
	}
	return items, nil
}

func (rds *DB_r) HouseActivityRecordReSave(datas []*models.HouseActivityRecord) error {

	for _, data := range datas {
		// key
		key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD, data.ActId)
		// field
		field := fmt.Sprintf("%d", data.UId)
		if rds.m_redis.HExists(key, field) {
			continue
		}
		rds.HouseActivityRecordInsert(data.ActId, data.UId, data.RankScore)
	}

	return nil
}

//! 包厢楼层活动数据新增
func (rds *DB_r) HouseActivityRecordInsert(actId int64, uid int64, score float64) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD, actId)
	// val
	val := uid
	//insert
	_, err := rds.m_redis.Zincrby(key, static.HF_I64tobytes(val), score)
	return err
}

//! 包厢楼层活动数据删除
func (rds *DB_r) HouseActivityRecordDelete(actId int64) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD, actId)
	//delete
	_, err := rds.m_redis.Del(key)
	return err
}
func (rds *DB_r) HouseActivityRecordItemDelete(actId int64, uid int64) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD, actId)
	// val
	val := uid
	//delete
	_, err := rds.m_redis.Zrem(key, static.HF_I64tobytes(val))
	return err
}

//! 包厢楼层活动数据查询
func (rds *DB_r) HouseActivityRecordList(actId int64, ibeg int, iend int) ([]*static.ActRecordItem, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD, actId)
	datas, err := rds.m_redis.ZrevrangeWithScore(key, ibeg, iend)
	if err != nil {
		return nil, err
	}

	var uid []int64
	var res []float64
	for index, data := range datas {
		if index%2 == 0 {
			uid = append(uid, static.HF_Bytestoi64(data))
		} else {
			res = append(res, static.HF_DecimalDivide(static.HF_Bytestof64(data), 1, 2))
		}
	}

	var houseRecordItems []*static.ActRecordItem
	for i := 0; i < len(res); i++ {

		person, err := rds.GetPerson(uid[i])
		if person == nil && err == nil {
			i++
			continue
		}

		item := new(static.ActRecordItem)
		item.ActId = actId
		item.UId = uid[i]
		item.UName = person.Nickname
		item.Score = res[i]
		houseRecordItems = append(houseRecordItems, item)
	}

	return houseRecordItems, nil
}

//! 楼层桌子数据插入
func (rds *DB_r) FloorTableInsert(db_floortable *static.FloorTable) error {
	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_FLOOR_TABLE, db_floortable.FId)
	// key
	key := fmt.Sprintf("%d", db_floortable.NTId)
	// val
	val := static.HF_JtoB(db_floortable)

	// "floortable_id"
	return rds.m_redis.HSet(hkey, key, val)
}

//! 楼层桌子数据删除
func (rds *DB_r) FloorTableDelete(db_floortable *static.FloorTable) error {
	//包厢成员
	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_FLOOR_TABLE, db_floortable.FId)
	// key
	key := fmt.Sprintf("%d", db_floortable.NTId)
	// res
	err := rds.m_redis.HDel(hkey, key)
	if err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// 包厢角色相关
func (rds *DB_r) HouseMemberBaseReLoad() ([]interface{}, error) {

	bytes, err := rds.m_redis.Keys(consts.REDIS_KEY_HOUSE_MEMBER_ALL)
	if err != nil {
		return nil, err
	}

	var items []interface{}
	for _, key := range bytes {

		datas, err := rds.m_redis.Hvals(static.HF_Bytestoa(key))
		if err != nil {
			xlog.Logger().Errorf("HouseMemberBaseReLoad Hvals： %v, key = %s", err, key)
			continue
		}

		for _, data := range datas {
			var dhmem static.HouseMember
			err = json.Unmarshal(data, &dhmem)
			if err != nil {
				xlog.Logger().Errorf("HouseMemberBaseReLoad Unmarshal： %v, key = %s", err, key)
				continue
			}
			items = append(items, &dhmem)
		}
	}

	return items, nil
}

func (rds *DB_r) HouseMemberBaseReSave(datas []*models.HouseMember) error {

	for _, data := range datas {
		dmem := data.ConvertModel()

		if dmem.URole > consts.ROLE_MEMBER {
			continue
		}

		p, _ := rds.GetPerson(dmem.UId)
		if p != nil {
			dmem.Sex = p.Sex
			dmem.ImgUrl = p.Imgurl
			dmem.NickName = p.Nickname
		}

		rds.HouseMemberInsert(dmem)

		// 玩家包厢数据
		rds.MemberHouseJoinInsert(dmem.UId, dmem.DHId)
		//// 玩家包厢大赢家统计
		//rds.HouseRecordBWInsert(dmem.DHId, dmem.UId, dmem.BwTimes)
		//// 玩家包厢对局统计
		//rds.HouseRecordPlayInsert(dmem.DHId, dmem.UId, dmem.PlayTimes)
	}

	return nil
}

func (rds *DB_r) HouseMemberClear(dhid int64) error {
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, dhid)
	return rds.RedisV2.Del(key).Err()
}

// 查询角色信息
func (rds *DB_r) HouseMemberQueryById(DHid int64, UId int64) (*static.HouseMember, error) {
	// jkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, DHid)
	// key
	key := fmt.Sprintf("%d", UId)
	// res
	bytes, err := rds.m_redis.HGet(hkey, key)
	if err != nil {
		return nil, err
	}

	// instance
	var data static.HouseMember
	if bytes != nil {
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			return nil, err
		}
	}

	return &data, nil
}

//! 创建包厢成员
func (rds *DB_r) HouseMemberInsert(db_houseMember *static.HouseMember) error {
	//包厢成员
	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, db_houseMember.DHId)
	// key
	key := fmt.Sprintf("%d", db_houseMember.UId)
	// val
	val := static.HF_JtoB(db_houseMember)
	// "housemem_dhid" - "uid-data"
	err := rds.m_redis.HSet(hkey, key, val)
	if err != nil {
		return err
	}
	return nil
}

//! 删除包厢成员
func (rds *DB_r) HouseMemberDelete(db_houseMember *static.HouseMember) error {
	//包厢成员
	// hkey
	hkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, db_houseMember.DHId)
	// key
	key := fmt.Sprintf("%d", db_houseMember.UId)
	// res
	err := rds.m_redis.HDel(hkey, key)
	if err != nil {
		return err
	}
	return nil
}

//! 插入包厢成员加入包厢数据
func (rds *DB_r) MemberHouseJoinDelete(uid int64, hid int64) error {
	// lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)
	// lval
	lval := hid
	// "memberhousejoin_id"
	_, err := rds.m_redis.Lrem(lkey, static.HF_I64tobytes(lval))
	return err
}

//! 删除包厢成员加入包厢数据
func (rds *DB_r) MemberHouseJoinInsert(uid int64, hid int64) error {
	//成员包厢
	// lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)
	// lval
	lval := hid
	// "memberhousejoin_id"
	//remove
	_, err := rds.m_redis.Lrem(lkey, static.HF_I64tobytes(lval))
	if err != nil {
		return err
	}
	//insert
	return rds.m_redis.Lpush(lkey, static.HF_I64tobytes(lval))
}

////! 包厢成员大赢家统计
//func (self *DB_r) HouseRecordBWList(dhid int64, ibeg int, iend int) ([]int64, error) {
//	// key
//	key := fmt.Sprintf("houserecord_bw_%d", dhid)
//	datas, err := self.m_redis.Zrevrange(key, ibeg, iend)
//	if err != nil {
//		return nil, err
//	}
//
//	var res []int64
//	for _, data := range datas {
//		res = append(res, public.HF_Bytestoi64(data))
//	}
//
//	var houseRecordItems []int64
//	for _, uid := range res {
//		houseRecordItems = append(houseRecordItems, uid)
//	}
//	return houseRecordItems, nil
//}
//
//func (self *DB_r) HouseRecordBWInsert(dhid int64, uid int64, bwtime int) error {
//	// key
//	key := fmt.Sprintf("houserecord_bw_%d", dhid)
//	// val
//	val := uid
//	// score
//	score := bwtime
//	//insert
//	_, err := self.m_redis.Zadd(key, public.HF_I64tobytes(val), float64(score))
//	return err
//}
//
//func (self *DB_r) HouseRecordBWDelete(dhid int64, uid int64) error {
//	// key
//	key := fmt.Sprintf("houserecord_bw_%d", dhid)
//	// val
//	val := uid
//	//delete
//	_, err := self.m_redis.Zrem(key, public.HF_I64tobytes(val))
//	return err
//}
//
////! 包厢成员对局统计
//func (self *DB_r) HouseRecordPlayList(dhid int64, ibeg int, iend int) ([]int64, error) {
//	// key
//	key := fmt.Sprintf("houserecord_play_%d", dhid)
//	datas, err := self.m_redis.Zrevrange(key, ibeg, iend)
//	if err != nil {
//		return nil, err
//	}
//
//	var res []int64
//	for _, data := range datas {
//		res = append(res, public.HF_Bytestoi64(data))
//	}
//
//	var houseRecordItems []int64
//	for _, uid := range res {
//		houseRecordItems = append(houseRecordItems, uid)
//	}
//	return houseRecordItems, nil
//}
//
//func (self *DB_r) HouseRecordPlayInsert(dhid int64, uid int64, playtime int) error {
//	// key
//	key := fmt.Sprintf("houserecord_play_%d", dhid)
//	// val
//	val := uid
//	// score
//	score := playtime
//	//insert
//	_, err := self.m_redis.Zadd(key, public.HF_I64tobytes(val), float64(score))
//	return err
//}
//
//func (self *DB_r) HouseRecordPlayDelete(dhid int64, uid int64) error {
//	// key
//	key := fmt.Sprintf("houserecord_play_%d", dhid)
//	// val
//	val := uid
//	//delete
//	_, err := self.m_redis.Zrem(key, public.HF_I64tobytes(val))
//	return err
//}
//
////! 包厢房卡统计
//func (self *DB_r) HouseRecordCostReSave(datas []interface{}) error {
//
//	var hid int64
//	for _, data := range datas {
//		item := data.(*public.RecordGameCostMini)
//
//		// 删除历史key
//		if hid != item.HId {
//			hid = item.HId
//			// key
//			key := fmt.Sprintf("houserecord_cost_%d", item.HId)
//			// delete
//			_, err := self.m_redis.Del(key)
//			if err != nil {
//				return err
//			}
//		}
//
//		// 插入数据
//		err := self.HouseRecordCostInsert(item)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func (self *DB_r) HouseRecordCostList(dhid int64) (*[]public.RecordGameCostMini, error) {
//	// key
//	key := fmt.Sprintf("houserecord_cost_%d", dhid)
//	datas, err := self.m_redis.Lrange(key, 0, 7)
//	if err != nil {
//		return nil, err
//	}
//
//	if datas == nil {
//		return nil, nil
//	}
//
//	var houseRecordItems []public.RecordGameCostMini
//	for _, data := range datas {
//		var record public.RecordGameCostMini
//		err := json.Unmarshal(data, &record)
//		if err != nil {
//			continue
//		}
//		houseRecordItems = append(houseRecordItems, record)
//	}
//
//	return &houseRecordItems, nil
//}
//
//func (self *DB_r) HouseRecordCostInsert(record *public.RecordGameCostMini) error {
//
//	//// 当日0点时间戳
//	//timeStr := time.Now().Format("2006-01-02")
//	//t, _ := time.Parse("2006-01-02", timeStr)
//	//timeNumber := t.Unix()-3600*8
//	//fmt.Println("timeNumber:", timeNumber)
//	//curZero := fmt.Sprintf("%d", timeNumber)
//
//	// key
//	key := fmt.Sprintf("houserecord_cost_%d", record.HId)
//	// 取出最后一条判定是否同一天数据
//	bytes, err := self.m_redis.Lpop(key)
//	if err != nil {
//		return err
//	}
//	if bytes != nil {
//		var lastval public.RecordGameCostMini
//		err = json.Unmarshal(bytes, &lastval)
//		if err != nil {
//			return err
//		}
//
//		if record.Date == lastval.Date {
//			// 同一天数据增加
//			record.PlayTime += lastval.PlayTime
//			record.KaCost += lastval.KaCost
//		} else {
//			// 非同天回写数据
//			err = self.m_redis.Lpush(key, bytes)
//		}
//	}
//
//	bval, err := json.Marshal(record)
//	if err != nil {
//		return err
//	}
//
//	//insert
//	err = self.m_redis.Lpush(key, bval)
//	if err != nil {
//		return err
//	}
//	// 长度判定
//	llen, err := self.m_redis.Llen(key)
//	if err != nil {
//		return err
//	}
//	// 大于7天数据清除
//	if llen > 7 {
//		self.m_redis.Rpop(key)
//	}
//
//	return nil
//}

//! 玩家包厢列表
func (rds *DB_r) MemberHouseIDList(uid int64) ([]int64, error) {

	//成员包厢
	//lkey
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)
	datas, err := rds.m_redis.Lrange(lkey, 0, -1)
	if err != nil {
		return nil, err
	}

	var res []int64
	for _, d := range datas {
		res = append(res, static.HF_Bytestoi64(d))
	}
	return res, nil
}

//! 获取Person
func (rds *DB_r) GetPerson(uid int64) (*static.Person, error) {
	if uid == 0 {
		return nil, errors.New("uid is zero")
	}
	person := new(static.Person)
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uid)
	err := rds.m_redis.Hgetall(key, person)
	//robot, e := rds.GetRobot(uid)
	//if e == nil {
	//	person = robot.Convert2Person()
	//}
	//if err != nil && e != nil {
	//	return nil, err
	//}
	if err != nil {
		return nil, err
	}
	return person, nil
}

//! 创建Person
func (rds *DB_r) AddPerson(person *static.Person) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_INFO, person.Uid)
	return rds.m_redis.HMsetall(key, person)
}

//! 获取Person属性
func (rds *DB_r) GetPersonAttrs(uid int64, field string) (string, error) {
	key := fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uid)
	if uid >= static.RobortIdMin && uid <= static.RobortIdMax {
		key = fmt.Sprintf(consts.REDIS_KEY_ROBOTS, uid)
	}
	ret := rds.RedisV2.HGet(key, field)
	return ret.Val(), ret.Err()
}

// 更新person属性
func (rds *DB_r) UpdatePersonAttrs(uid int64, args ...interface{}) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uid)
	if uid >= static.RobortIdMin && uid <= static.RobortIdMax {
		key = fmt.Sprintf(consts.REDIS_KEY_ROBOTS, uid)
	}
	err := rds.m_redis.HMset(key, args...)
	if err != nil {
		xlog.Logger().Error("UpdatePersonAttrs.error=", err)
		return err
	}
	return nil
}

// 更新person属性
func (rds *DB_r) UpdatePersonAttrsV2(uid int64, updates map[string]interface{}) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uid)
	if uid >= static.RobortIdMin && uid <= static.RobortIdMax {
		key = fmt.Sprintf(consts.REDIS_KEY_ROBOTS, uid)
	}
	return rds.RedisV2.HMSet(key, updates).Err()
}

//！根据id删除person
func (rds *DB_r) DelPerson(uid int64) (bool, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_INFO, uid)
	return rds.m_redis.Del(key)
}

// 更新在包厢牌桌的最后一局的对战玩家
func (rds *DB_r) UpdateHRecordPlayers(uid int64, record *static.HTableRecordPlayers) error {
	xlog.Logger().Infoln("写入包厢最新一局对战玩家.")
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD_PLAYER, uid)
	if record == nil {
		_, _ = rds.m_redis.Del(key)
		return nil
	}
	return rds.m_redis.Set(key, static.HF_JtoB(record))
}

// 得到玩家在任意包厢牌桌的最后一局的对战玩家信息
func (rds *DB_r) SelectHRecordPlayers(uid int64) (*static.HTableRecordPlayers, error) {
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_RECORD_PLAYER, uid)
	data, err := rds.m_redis.Get(key)
	if err != nil {
		return nil, err
	}
	record := new(static.HTableRecordPlayers)
	record.Users = make([]int64, 0)
	err = json.Unmarshal(data, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

//! 创建Account
//func (self *DB_r) InsertAccount(account *public.DB_Account) error {
//	// key
//	key := fmt.Sprintf("account_%d", account.Uid)
//	// val
//	val := public.HF_JtoB(account)
//	return self.m_redis.Set(key, val)
//}

//！ 保存战绩回放数据
//func (self *DB_r) GameRecordReloadReSave(datas []interface{}) error {
//
//	for _, data := range datas {
//		self.InsertGameRecord(data.(*static.GameRecord))
//	}
//
//	return nil
//}
// 获取玩家最后一局游戏小结算数据
func (rds *DB_r) GetLastGameInfo(uid int64) (*static.LastGameInfo, error) {
	xlog.Logger().Infoln("读取玩家最新一局的小结算.")
	key := fmt.Sprintf(consts.REDIS_KEY_LAST_GAME_INFO, uid)
	data, err := rds.m_redis.Get(key)
	if err != nil {
		return nil, err
	}
	record := new(static.LastGameInfo)
	record.Users = make([]int64, 0)
	err = json.Unmarshal(data, record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// 更新玩家最后一局游戏小结算数据
func (rds *DB_r) UpdateLastGameInfo(uid int64, record *static.LastGameInfo) error {
	xlog.Logger().Infoln("写入玩家最新一局的小结算.")
	key := fmt.Sprintf(consts.REDIS_KEY_LAST_GAME_INFO, uid)
	if record == nil {
		_, _ = rds.m_redis.Del(key)
		return nil
	}
	return rds.m_redis.Set(key, static.HF_JtoB(record))
}

// 设置玩家最后一局记录的过期时间
func (rds *DB_r) SetLastGameInfoExpire(uid int64) {
	key := fmt.Sprintf(consts.REDIS_KEY_LAST_GAME_INFO, uid)
	//rds.m_redis.Set(key, []byte("1"))
	rds.m_redis.Expire(key, 600) // 600s有效期
}

//! 插入游戏次数
func (rds *DB_r) GamePlaysInsert(model *models.StatisticsUserGameHistory) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_PLAYTIMES, model.Uid)

	return rds.m_redis.HSet(key, static.HF_Itoa(model.KindId), static.HF_JtoB(model))
}

//! 获得游戏次数
func (rds *DB_r) GamePlaysSelect(model *models.StatisticsUserGameHistory) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_PLAYTIMES, model.Uid)
	// field
	field := fmt.Sprintf("%d", model.KindId)
	// val
	val, err := rds.m_redis.HGet(key, field)
	if err != nil {
		return err
	}
	return json.Unmarshal(val, model)
}

//! 获得用户所有游戏历史信息
func (rds *DB_r) GamePlaysSelectAll(uid int64) (models.UserGameHistoryList, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_PLAYTIMES, uid)

	val, err := rds.m_redis.Hvals(key)

	if err != nil {
		return nil, err
	}

	result := make(models.UserGameHistoryList, 0)
	for _, data := range val {
		history := new(models.StatisticsUserGameHistory)
		if err := json.Unmarshal(data, history); err != nil {
			return nil, err
		}
		result = append(result, history)
	}

	return result, nil
}

//! 获得游戏次数列表
func (rds *DB_r) GamePlayList(uid int64) ([]int, error) {

	var res []int

	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_PLAYS, uid)
	datas, err := rds.m_redis.Zrevrange(key, 0, -1)
	if err != nil {
		return res, err
	}

	for _, data := range datas {
		res = append(res, static.HF_Bytestoi(data))
	}

	return res, nil
}

////! 插入单局战绩数据
//func (self *DB_r) InsertGameRecord(gameRecord *public.GameRecord) error {
//	// key
//	key := fmt.Sprintf("record_gamenum_%s_%d", gameRecord.GameNum, gameRecord.UId)
//	// val
//	val := public.HF_JtoB(gameRecord)
//	return self.m_redis.Rpush(key, val)
//}
//
////! 查询单局战绩数据
//func (self *DB_r) SelectGameRecord(GameNum string, uid int64) ([]*public.GameRecord, error) {
//	// key
//	key := fmt.Sprintf("record_gamenum_%s_%d", GameNum, uid)
//	// lval
//	lval, err := self.m_redis.Lrange(key, 0, -1)
//	if err != nil {
//		return nil, err
//	}
//
//	datas := make([]*public.GameRecord, 0)
//	for _, val := range lval {
//		var record public.GameRecord
//		err = json.Unmarshal([]byte(val), &record)
//		datas = append(datas, &record)
//	}
//
//	return datas, nil
//}
//
////! 删除单局战绩数据
//func (self *DB_r) DeleteGameRecord(GameNum string, uid int64) error {
//	key := fmt.Sprintf("record_gamenum_%s_%d", GameNum, uid)
//
//	return self.m_redis.Expire(key, 1)
//}
//
////! 保存战绩回放数据
//func (self *DB_r) GameRecordReplayReSave(datas *[]model.RecordGameReplay) error {
//	for _, data := range *datas {
//		gameReplay := data.ConvertModel()
//
//		// key
//		key := fmt.Sprintf("record_replayid_%d", gameReplay.Id)
//		if self.m_redis.Exists(key) {
//			continue
//		}
//
//		err := self.InsertGameRecordReplay(gameReplay)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
////! 插入单局回放数据
//func (self *DB_r) InsertGameRecordReplay(gameRecord *public.GameRecordReplay) error {
//	// key
//	key := "record_replay"
//	// filed
//	field := fmt.Sprintf("%d", gameRecord.Id)
//	// val
//	val := public.HF_JtoB(gameRecord)
//
//	return self.m_redis.HSet(key, field, val)
//}
//
////! 查询单局战绩数据
//func (self *DB_r) SelectGameRecordReplay(ReplayId int64) (*public.GameRecordReplay, error) {
//	// key
//	key := "record_replay"
//	// filed
//	field := fmt.Sprintf("%d", ReplayId)
//	// val
//	bytes, err := self.m_redis.HGet(key, field)
//	if err != nil {
//		return nil, err
//	}
//
//	if bytes == nil {
//		err = errors.New("record not exists")
//		return nil, err
//	}
//
//	var recordReplay public.GameRecordReplay
//	err = json.Unmarshal([]byte(bytes), &recordReplay)
//	if err != nil {
//		return nil, err
//	}
//
//	return &recordReplay, nil
//}
//
////！ 保存总战绩数据
//func (self *DB_r) GameRecordTotalReSave(datas *[]model.RecordGameTotal) error {
//	var recordmap = make(map[string][]model.RecordGameTotal)
//
//	for _, data := range *datas {
//		recordmap[data.GameNum] = append(recordmap[data.GameNum], data)
//	}
//
//	for _, mapdata := range recordmap {
//		record := new(public.GameRecordHistory)
//		record.Player = make([]*public.GameRecordHistoryPlayer, 0)
//
//		for _, data := range mapdata {
//			if data.SeatId == 0 {
//				gameround, err := self.SelectGameRecordInfo(data.GameNum)
//				if err == nil && gameround != nil {
//					record.KindId = gameround.KindId
//				} else {
//					record.KindId = 0
//				}
//				record.GameNum = data.GameNum
//				record.RoomNum = data.RoomNum
//				record.PlayedAt = data.CreatedAt.Unix()
//				record.HId = data.HId
//				record.FId = data.FId
//				record.PlayCount = data.PlayCount
//				record.Round = data.Round
//				record.IsHeart = data.IsHeart
//			}
//
//			person := new(public.GameRecordHistoryPlayer)
//			person.Uid = data.Uid
//			person.Nickname = data.UName
//			person.Score = data.WinScore
//			record.Player = append(record.Player, person)
//		}
//		if record.HId > 0 {
//			//记录一条包厢战绩(包厢战绩)
//			self.InsertHouseHallRecord(record.HId, record)
//			for _, p := range record.Player {
//				// 每人记录一条(包厢我的战绩)
//				self.InsertHouseHallMyRecord(record.HId, p.Uid, record)
//			}
//		} else {
//			for _, p := range record.Player {
//				// 每人记录一条
//				self.InsertNormalHallRecord(p.Uid, record)
//			}
//		}
//	}
//	return nil
//}
//
////! 保存单局回放数据
//func (self *DB_r) GameRecordRoundReSave(datas *[]model.RecordGameRound) error {
//	var recordmap = make(map[string]map[int]map[int]model.RecordGameRound)
//
//	for _, data := range *datas {
//		_, ok := recordmap[data.GameNum]
//		if !ok {
//			recordmap[data.GameNum] = make(map[int]map[int]model.RecordGameRound)
//		}
//		_, ok = recordmap[data.GameNum][data.PlayNum]
//		if !ok {
//			recordmap[data.GameNum][data.PlayNum] = make(map[int]model.RecordGameRound)
//		}
//		recordmap[data.GameNum][data.PlayNum][data.SeatId] = data
//	}
//
//	for _, mapdata := range recordmap {
//		// 玩家局数
//		PlayCount := len(mapdata)
//
//		// 记录对局详情
//		roundRecord := new(public.Msg_S2C_GameRecordInfo)
//		// 每局积分
//		scoreArr := make([][]int, 0)
//
//		replayIdArr := make([]int64, 0)
//		endTimeArr := make([]int64, 0)
//		startTimeArr := make([]int64, 0)
//
//		//玩家数量
//		UserCount := 0
//		for playnum := 0; playnum < PlayCount; playnum++ {
//			UserCount = len(mapdata[playnum+1])
//
//			if playnum == 0 {
//				// 初始化二维数组
//				for i := 0; i < PlayCount; i++ {
//					arr := make([]int, UserCount)
//					scoreArr = append(scoreArr, arr)
//				}
//
//				// 玩家列表
//				for seatid := 0; seatid < UserCount; seatid++ {
//					p := mapdata[playnum+1][seatid]
//					roundRecord.UserArr = append(roundRecord.UserArr, &public.Msg_S2C_GameRecordInfoUser{
//						Uid:      p.UId,
//						Nickname: p.UName,
//						Imgurl:   p.UUrl,
//						Sex:      p.UGenber,
//						Score:    0,
//					})
//				}
//			}
//
//			for seatid := 0; seatid < UserCount; seatid++ {
//				p := mapdata[playnum+1][seatid]
//
//				scoreArr[playnum][seatid] = p.WinScore
//				if seatid == 0 {
//					if playnum == 0 {
//						recordreplay, err := self.SelectGameRecordReplay(p.ReplayId)
//						if err == nil {
//							roundRecord.KindId = recordreplay.KindID
//						} else {
//							roundRecord.KindId = 0
//						}
//						roundRecord.GameNum = p.GameNum
//						roundRecord.RoomId = p.RoomNum
//						roundRecord.Time = p.CreatedAt.Unix()
//					}
//					replayIdArr = append(replayIdArr, p.ReplayId)
//					endTimeArr = append(endTimeArr, p.EndDate.Unix())
//					startTimeArr = append(startTimeArr, p.BeginDate.Unix())
//				}
//			}
//		}
//
//		for i := 0; i < PlayCount; i++ {
//			roundRecord.ScoreArr = append(roundRecord.ScoreArr, &public.Msg_S2C_GameRecordInfoScore{
//				ReplayId:  replayIdArr[i],
//				StartTime: startTimeArr[i],
//				EndTime:   endTimeArr[i],
//				Score:     scoreArr[i],
//			})
//
//			for j := 0; j < UserCount; j++ {
//				roundRecord.UserArr[j].Score += scoreArr[i][j]
//			}
//		}
//
//		//! 插入redis
//		err := self.InsertGameRecordInfo(roundRecord)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
////! 插入大厅个人历史战绩数据
//func (self *DB_r) InsertNormalHallRecord(uid int64, gameRecord *public.GameRecordHistory) error {
//	playTime := time.Unix(gameRecord.PlayedAt, 0)
//	timeKey := fmt.Sprintf("%d%02d%02d", playTime.Year(), playTime.Month(), playTime.Day())
//	key := fmt.Sprintf("record_hall_%d_%s", uid, timeKey)
//
//	// val
//	val := public.HF_JtoB(gameRecord)
//	return self.m_redis.Lpush(key, val)
//
//	//! 设置改数据八天过期
//	return self.m_redis.Expire(key, public.EIGHT_DAY_SECONDS)
//}
//
////! 查询大厅个人历史战绩
//func (self *DB_r) SelectNormalHallRecord(uid int64, selecttime time.Time) ([]*public.GameRecordHistory, error) {
//	// key
//	timeKey := fmt.Sprintf("%d%02d%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
//	key := fmt.Sprintf("record_hall_%d_%s", uid, timeKey)
//	// val
//	lval, err := self.m_redis.Lrange(key, 0, -1)
//	if err != nil {
//		return nil, err
//	}
//
//	datas := make([]*public.GameRecordHistory, 0)
//	for _, val := range lval {
//		var record public.GameRecordHistory
//		err = json.Unmarshal([]byte(val), &record)
//		datas = append(datas, &record)
//	}
//	return datas, nil
//}
//
////! 插入包厢历史战绩
//func (self *DB_r) InsertHouseHallRecord(hid int64, gameRecord *public.GameRecordHistory) error {
//	// key
//	key := fmt.Sprintf("record_house_%d", hid)
//
//	//field
//	field := fmt.Sprintf("%s", gameRecord.GameNum)
//
//	// val
//	val := public.HF_JtoB(gameRecord)
//	return self.m_redis.HSet(key, field, val)
//}
//
////! 查询包厢历史战绩
//func (self *DB_r) SelectHouseHallRecord(hid int64) ([]*public.GameRecordHistory, error) {
//	// key
//	key := fmt.Sprintf("record_house_%d", hid)
//	// val
//	lval, err := self.m_redis.Hvals(key)
//	if err != nil {
//		return nil, err
//	}
//
//	datas := make([]*public.GameRecordHistory, 0)
//	for _, val := range lval {
//		var record public.GameRecordHistory
//		err = json.Unmarshal([]byte(val), &record)
//		datas = append(datas, &record)
//	}
//	return datas, nil
//}
//
////! 查询包厢历史战绩
//func (self *DB_r) SelectHouseHallRecordWithGameNum(hid int64, gamenum string) (*public.GameRecordHistory, error) {
//	// key
//	key := fmt.Sprintf("record_house_%d", hid)
//	//field
//	field := fmt.Sprintf("%s", gamenum)
//
//	// val
//	val, err := self.m_redis.HGet(key, field)
//	if err != nil {
//		return nil, err
//	}
//
//	var record public.GameRecordHistory
//	err = json.Unmarshal([]byte(val), &record)
//	if err != nil {
//		return nil, err
//	}
//	return &record, nil
//}
//
////! 删除过期包厢战绩
//func (self *DB_r) DeleteHouseHallRecordWithGameNum(hid int64, gamenum string) error {
//	// key
//	key := fmt.Sprintf("record_house_%d", hid)
//	//field
//	field := fmt.Sprintf("%s", gamenum)
//
//	// val
//	err := self.m_redis.HDel(key, field)
//	return err
//}
//
////! 插入包厢我的历史战绩
//func (self *DB_r) InsertHouseHallMyRecord(hid int64, uid int64, gameRecord *public.GameRecordHistory) error {
//	// key
//	key := fmt.Sprintf("record_house_%d_%d", hid, uid)
//
//	//field
//	field := fmt.Sprintf("%s", gameRecord.GameNum)
//	// val
//	val := public.HF_JtoB(gameRecord)
//	return self.m_redis.HSet(key, field, val)
//}
//
////! 查询包厢历史战绩
//func (self *DB_r) SelectHouseHallMyRecord(hid int, uid int64) ([]*public.GameRecordHistory, error) {
//	// key
//	key := fmt.Sprintf("record_house_%d_%d", hid, uid)
//	// val
//	lval, err := self.m_redis.Hvals(key)
//	if err != nil {
//		return nil, err
//	}
//
//	datas := make([]*public.GameRecordHistory, 0)
//	for _, val := range lval {
//		var record public.GameRecordHistory
//		err = json.Unmarshal([]byte(val), &record)
//		datas = append(datas, &record)
//	}
//	return datas, nil
//}
//
////! 删除过期我的战绩
//func (self *DB_r) DeleteHouseHallMyRecordWithGameNum(hid int, uid int64, gamenum string) error {
//	// key
//	key := fmt.Sprintf("record_house_%d_%d", hid, uid)
//	//field
//	field := fmt.Sprintf("%s", gamenum)
//
//	// val
//	err := self.m_redis.HDel(key, field)
//	return err
//}
//
////! 每日统计数据数据库恢复
//func (self *DB_r) GameDayRecordReSave(datas *[]model.RecordGameDay) error {
//
//	for _, data := range *datas {
//		dayRecord := data.ConvertModel()
//
//		self.InsertGameDayRecordToDaySet(dayRecord)
//	}
//
//	return nil
//}
//
////! 插入每日统计数据到天数据中
//func (self *DB_r) InsertGameDayRecordToDaySet(dayRecord *public.GameDayRecord) error {
//	// key
//	playTime, err := time.Parse("2006-01-02", dayRecord.PlayDate)
//	if err != nil {
//		return err
//	}
//
//	timeKey := fmt.Sprintf("%d%02d%02d", playTime.Year(), playTime.Month(), playTime.Day())
//	key := fmt.Sprintf("day_record_%d_%s", dayRecord.DHId, timeKey)
//	// field
//	field := fmt.Sprintf("%d_%d", dayRecord.UId, dayRecord.DFId)
//	// val
//	val := public.HF_JtoB(dayRecord)
//
//	err = self.m_redis.HSet(key, field, val)
//
//	//! 设置改数据八天过期
//	return self.m_redis.Expire(key, public.EIGHT_DAY_SECONDS)
//}
//
////! 查询单个玩家每天统计数据
//func (self *DB_r) SelectUserGameDayRecordToDaySet(DHid int64, DFId int64, UId int64, selecttime time.Time) (*public.GameDayRecord, error) {
//	// key
//	timeKey := fmt.Sprintf("%d%02d%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
//	key := fmt.Sprintf("day_record_%d_%s", DHid, timeKey)
//	// field
//	field := fmt.Sprintf("%d_%d", UId, DFId)
//	// val
//	bytes, err := self.m_redis.HGet(key, field)
//	if err != nil {
//		return nil, err
//	}
//
//	var dayRecord public.GameDayRecord
//	err = json.Unmarshal([]byte(bytes), &dayRecord)
//	if err != nil {
//		return nil, err
//	}
//	return &dayRecord, nil
//}
//
////! 查询每日统计天数据
//func (self *DB_r) SelectGameDayRecordToDaySet(hid int, selecttime time.Time) (*[]public.GameDayRecord, error) {
//	// key
//	timeKey := fmt.Sprintf("%d%02d%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
//	key := fmt.Sprintf("day_record_%d_%s", hid, timeKey)
//	// val
//	lval, err := self.m_redis.Hvals(key)
//	if err != nil {
//		return nil, err
//	}
//	var datas []public.GameDayRecord
//	for _, val := range lval {
//		var dayRecord public.GameDayRecord
//		err = json.Unmarshal([]byte(val), &dayRecord)
//		datas = append(datas, dayRecord)
//	}
//	return &datas, nil
//}
//
////! 插入对局详情
//func (self *DB_r) InsertGameRecordInfo(gameRecord *public.Msg_S2C_GameRecordInfo) error {
//	// key
//	key := "record_info"
//	// filed
//	filed := fmt.Sprintf("%s", gameRecord.GameNum)
//	// val
//	val := public.HF_JtoB(gameRecord)
//
//	return self.m_redis.HSet(key, filed, val)
//}
//
////! 获取对局详情
//func (self *DB_r) SelectGameRecordInfo(gameNum string) (*public.Msg_S2C_GameRecordInfo, error) {
//	// key
//	key := "record_info"
//	// filed
//	filed := fmt.Sprintf("%s", gameNum)
//	// val
//	bytes, err := self.m_redis.HGet(key, filed)
//	if err != nil {
//		return nil, err
//	}
//
//	if bytes == nil {
//		err = errors.New("game record eve not exists")
//		return nil, err
//	}
//
//	var record public.Msg_S2C_GameRecordInfo
//	err = json.Unmarshal([]byte(bytes), &record)
//	if err != nil {
//		return nil, err
//	}
//
//	return &record, nil
//}

// 检查短信验证码是否发送过
func (rds *DB_r) CheckSmsCodeSend(mobile string, _type uint8) bool {
	key := fmt.Sprintf(consts.REDIS_KEY_SMSCODE, _type, mobile)
	return rds.m_redis.Exists(key)
}

// 设置短信验证码使用过
func (rds *DB_r) SetSmsCodeSend(mobile string, _type uint8) {
	key := fmt.Sprintf(consts.REDIS_KEY_SMSCODE, _type, mobile)
	rds.m_redis.Set(key, []byte("1"))
	rds.m_redis.Expire(key, 60) // 60s有效期
}

// 用户相关
func (rds *DB_r) UserBaseReLoad() ([]*static.Person, error) {
	keysArr, err := rds.Keys(consts.REDIS_KEY_USER_INFO_ALL)
	if err != nil {
		return nil, err
	}

	result := make([]*static.Person, 0)
	for _, key := range keysArr {
		tempArr := strings.Split(key, "_")
		if len(tempArr) != 2 {
			continue
		}
		uid, _ := strconv.ParseInt(tempArr[1], 10, 64)
		person, err := rds.GetPerson(uid)
		if err != nil {
			continue
		}
		result = append(result, person)
	}
	return result, nil
}

// 用户相关
func (rds *DB_r) RobotsBaseReLoad() ([]*static.Robot, error) {
	keysArr, err := rds.Keys(consts.REDIS_KEY_ROBOTS_ALL)
	if err != nil {
		return nil, err
	}

	result := make([]*static.Robot, 0)
	for _, key := range keysArr {
		tempArr := strings.Split(key, "_")
		if len(tempArr) != 2 {
			continue
		}
		uid, _ := strconv.ParseInt(tempArr[1], 10, 64)
		r, err := rds.GetRobot(uid)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

// 任务相关
func (rds *DB_r) UserTaskReLoad() ([]*static.Task, error) {
	keysArr, err := rds.Keys(consts.REDIS_KEY_USER_TASK_ALL)
	if err != nil {
		return nil, err
	}

	result := make([]*static.Task, 0)
	for _, key := range keysArr {
		datas, err := rds.m_redis.Hvals(key)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			task := new(static.Task)
			err = json.Unmarshal(data, task)
			if err != nil {
				return nil, err
			}
			result = append(result, task)
		}
	}
	return result, nil
}

// 任务相关
func (rds *DB_r) UserTaskReSave(datas []*models.UserTask) error {

	for _, data := range datas {
		// key
		key := fmt.Sprintf(consts.REDIS_KEY_USER_TASK, data.Uid)
		// field
		field := fmt.Sprintf("%d", data.TaskId)
		if rds.m_redis.HExists(key, field) {
			continue
		}
		rds.UserTaskInsert(data.ConvertModel())
	}
	return nil
}

func (rds *DB_r) UserTaskList(uid int64) (static.TaskList, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_TASK, uid)

	datas, err := rds.m_redis.Hvals(key)
	if err != nil {
		return nil, err
	}

	tasks := make([]*static.Task, 0)
	for _, data := range datas {
		task := new(static.Task)
		err = json.Unmarshal(data, task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

type League struct {
	LeagueId int64 `redis:"league_id"`
	Card     int64 `redis:"card_num"`
	Freeze   bool  `redis:"freeze"`
	UserNum  int64 `redis:"user_num"`
	AreaCode int64 `redis:"area_code"`
}

func (rds *DB_r) LeagueBaseInfo() ([]*League, error) {
	keysArr, err := rds.Keys(consts.REDIS_KEY_LEAGUE_ALL)
	if err != nil {
		return nil, err
	}
	res := make([]*League, 0, len(keysArr))
	cli := rds.GetRedisCli()
	for _, key := range keysArr {
		if len(key) >= 16 {
			continue
		}
		leagueInfo := League{}
		err := cli.Hgetall(key, &leagueInfo)
		if err != nil {
			xlog.Logger().Errorln("unmarshal task eve failed: ", err.Error(), key)
			continue
		}
		res = append(res, &leagueInfo)
	}
	return res, nil
}

func (rds *DB_r) LeagueUserInfo() (map[interface{}][]interface{}, error) {
	keysArr, err := rds.Keys("user_league_h*")
	if err != nil {
		return nil, err
	}
	res := make(map[interface{}][]interface{})
	cli := rds.RedisV2
	for _, key := range keysArr {
		resp := cli.HMGet(key, "league_id", "uid", "").Val()
		if len(resp) < 2 {
			continue
		}
		if res[resp[0]] == nil {
			res[resp[0]] = []interface{}{}
		}
		res[resp[0]] = append(res[resp[0]], resp[1])
	}
	return res, nil
}

func (rds *DB_r) LeagueBaseReSave(datas []*models.League) error {
	cli := rds.GetRedisCli()
	for _, data := range datas {
		legaueField := make(map[string]interface{}, 4)
		legaueField["card_num"] = data.Card
		legaueField["freeze_card"] = data.FreezeCard
		legaueField["freeze"] = false
		legaueField["user_num"] = data.UserNum
		legaueField["league_id"] = data.LeagueID
		legaueField["area_code"] = data.AreaCode
		legaueField["pool_state"] = data.PoolState
		legaueField["pool_start"] = data.PoolStart.Unix()
		legaueField["pool_end"] = data.PoolEnd.Unix()
		cli.HMsetall(fmt.Sprintf(consts.REDIS_KEY_LEAGUE, data.LeagueID), legaueField)
	}
	return nil

}

func (rds *DB_r) LeagueUserReSave(data []*models.LeagueUser) error {
	cli := rds.GetRedisCli()
	for _, lu := range data {
		redisKey := fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER, lu.LeagueID)
		cli.Sadd(redisKey, lu.Uid)
		hMap := make(map[string]interface{}, 6)
		hMap["league_id"] = lu.LeagueID
		hMap["freeze"] = lu.Freeze
		hMap["pool_state"] = lu.PoolState
		hMap["pool_start"] = lu.PoolStart.Unix()
		hMap["pool_end"] = lu.PoolEnd.Unix()
		hMap["uid"] = lu.Uid
		key := fmt.Sprintf(consts.REDIS_KEY_LEAGUE_USER_H, lu.Uid)
		err := cli.HMsetall(key, hMap)
		if err != nil {
			xlog.Logger().Errorf("redis 保存数据错误：%+v", err)
			continue
		}
	}
	return nil
}

//大厅显示的任务
func (rds *DB_r) UserTaskInsert(task *static.Task) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_TASK, task.Uid)
	// field
	field := fmt.Sprintf("%d", task.TcId)
	// val
	val := static.HF_JtoB(task)

	return rds.m_redis.HSet(key, field, val)
}

func (rds *DB_r) UserTaskUpdate(task *static.Task) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_TASK, task.Uid)
	// field
	field := fmt.Sprintf("%d", task.TcId)
	// val
	val := static.HF_JtoB(task)

	return rds.m_redis.HSet(key, field, val)
}

//大厅显示的任务
func (rds *DB_r) UserTaskDelete(task *static.Task) error {
	// key
	hkey := fmt.Sprintf(consts.REDIS_KEY_USER_TASK, task.Uid)
	// key
	key := fmt.Sprintf("%d", task.TcId)

	return rds.m_redis.HDel(hkey, key)
}

//游戏内显示的任务
func (rds *DB_r) UserGameTaskInsert(task *static.Task) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_GAME_TASK, task.Uid)
	// field
	field := fmt.Sprintf("%d", task.TcId)
	// val
	val := static.HF_JtoB(task)

	return rds.m_redis.HSet(key, field, val)
}

//游戏内显示的任务
func (rds *DB_r) UserGameTaskDelete(task *static.Task) error {
	// key
	hkey := fmt.Sprintf(consts.REDIS_KEY_USER_GAME_TASK, task.Uid)
	// key
	key := fmt.Sprintf("%d", task.TcId)

	return rds.m_redis.HDel(hkey, key)
}

// 任务相关
func (rds *DB_r) UserGameTaskReLoad() ([]*static.Task, error) {
	keysArr, err := rds.Keys(consts.REDIS_KEY_USER_GAME_TASK_ALL)
	if err != nil {
		return nil, err
	}

	result := make([]*static.Task, 0)
	for _, key := range keysArr {
		datas, err := rds.m_redis.Hvals(key)
		if err != nil {
			return nil, err
		}
		for _, data := range datas {
			task := new(static.Task)
			err = json.Unmarshal(data, task)
			if err != nil {
				return nil, err
			}
			result = append(result, task)
		}
	}
	return result, nil
}

// 任务相关
func (rds *DB_r) UserGameTaskReSave(datas []*models.UserTaskGame) error {

	for _, data := range datas {
		// key
		key := fmt.Sprintf(consts.REDIS_KEY_USER_GAME_TASK, data.Uid)
		// field
		field := fmt.Sprintf("%d", data.TaskId)
		if rds.m_redis.HExists(key, field) {
			continue
		}
		rds.UserGameTaskInsert(data.ConvertModel())
	}
	return nil
}

func (rds *DB_r) UserGameTaskList(uid int64) ([]*static.Task, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_USER_GAME_TASK, uid)

	datas, err := rds.m_redis.Hvals(key)
	if err != nil {
		return nil, err
	}

	tasks := make([]*static.Task, 0)
	for _, data := range datas {
		task := new(static.Task)
		err = json.Unmarshal(data, task)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (rds *DB_r) SaveHftDBToRedis(dest []*models.HousemixfloorTable) error {
	cli := rds.GetRedisCli()
	for _, v := range dest {
		key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_FLOOR_MIX, v.Fid)
		_ = cli.HSet(key, fmt.Sprintf("%d", v.NTID), []byte(v.HftInfo))
	}
	return nil
}

func (rds *DB_r) GetHouseMemberWithUids(dhid int64, uids []int64) ([]*static.HouseMember, error) {
	Int64ToStringSlice := func(values []int64) []string {
		ss := make([]string, 0)
		for _, v := range values {
			ss = append(ss, fmt.Sprintf("%d", v))
		}
		return ss
	}

	InterfaceToStringSlice := func(values []interface{}) []string {
		ss := make([]string, 0)
		for _, v := range values {
			ss = append(ss, fmt.Sprintf("%v", v))
		}
		return ss
	}

	ms := make([]*static.HouseMember, 0)

	cmd := rds.RedisV2.HMGet(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, dhid), Int64ToStringSlice(uids)...)

	err := goRedis.NewStringSliceResult(InterfaceToStringSlice(cmd.Val()), cmd.Err()).ScanSlice(&ms)

	return ms, err
}

func (rds *DB_r) GetHouseMemberMap(dhid int64) (map[int64]*static.HouseMember, error) {
	ms := make([]*static.HouseMember, 0)
	err := rds.RedisV2.HVals(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, dhid)).ScanSlice(&ms)
	if err != nil {
		return nil, err
	}
	msMap := make(map[int64]*static.HouseMember)
	for _, mem := range ms {
		msMap[mem.UId] = mem
	}
	return msMap, nil
}

func (rds *DB_r) GetHouseMemberMapWithUids(dhid int64, uids []int64) (map[int64]*static.HouseMember, error) {
	Int64ToStringSlice := func(values []int64) []string {
		ss := make([]string, 0)
		for _, v := range values {
			ss = append(ss, fmt.Sprintf("%d", v))
		}
		return ss
	}

	InterfaceToStringSlice := func(values []interface{}) []string {
		ss := make([]string, 0)
		for _, v := range values {
			ss = append(ss, fmt.Sprintf("%v", v))
		}
		return ss
	}

	ms := make([]*static.HouseMember, 0)

	cmd := rds.RedisV2.HMGet(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, dhid), Int64ToStringSlice(uids)...)

	err := goRedis.NewStringSliceResult(InterfaceToStringSlice(cmd.Val()), cmd.Err()).ScanSlice(&ms)

	msMap := make(map[int64]*static.HouseMember)
	if err != nil {
		return msMap, err
	}

	for _, mem := range ms {
		msMap[mem.UId] = mem
	}
	return msMap, err
}

func (rds *DB_r) AreaCodeToPkgsSave(kind static.AreaPackageKind, areaPkgs map[string]*static.PackKeys) error {
	keys, err := rds.RedisV2.Keys(static.AreaCodeRedisKey(kind, "*")).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		err = rds.RedisV2.Del(keys...).Err()
		if err != nil {
			return err
		}
	}

	for key, val := range areaPkgs {
		if len(*val) == 0 {
			continue
		}
		if err := rds.RedisV2.SAdd(static.AreaCodeRedisKey(kind, key), val.Convert()...).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (rds *DB_r) AreaPackageSave(key string, data map[string]interface{}) error {
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}
	if len(data) > 0 {
		return rds.RedisV2.HMSet(key, data).Err()
	}
	return nil
}

func (rds *DB_r) AreaKidToRuleSave(key string, data map[string]interface{}) error {
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}
	if len(data) > 0 {
		return rds.RedisV2.HMSet(key, data).Err()
	}
	return nil
}

func (rds *DB_r) AreaKidToPackageKeySave(key string, data map[string]interface{}) error {
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}
	if len(data) > 0 {
		return rds.RedisV2.HMSet(key, data).Err()
	}
	return nil
}

func (rds *DB_r) AreaKidToExplainSave(data map[string]interface{}) error {
	if len(data) > 0 {
		return rds.RedisV2.HMSet(static.AreaExplainRedisKey(), data).Err()
	}
	return nil
}

func (rds *DB_r) AreaPackageRecommendSave(kind static.AreaPackageKind, data []interface{}) error {
	key := static.AreaRecommendRedisKey(kind)
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}

	if len(data) > 0 {
		return rds.RedisV2.SAdd(key, data...).Err()
	}

	return nil
}

func (rds *DB_r) AreaPackageUniversalSave(kind static.AreaPackageKind, data []interface{}) error {
	key := static.AreaUniversalRedisKey(kind)
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}

	if len(data) > 0 {
		return rds.RedisV2.SAdd(key, data...).Err()
	}

	return nil
}

func (rds *DB_r) AreaPackageCodeListSave(key string, data []interface{}) error {
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}
	if len(data) > 0 {
		return rds.RedisV2.SAdd(key, data...).Err()
	}
	return nil
}

func (rds *DB_r) AreaCodeToGamesSave(key string, data map[string]interface{}) error {
	if rds.RedisV2.Exists(key).Val() == 1 {
		if err := rds.RedisV2.Del(key).Err(); err != nil {
			return err
		}
	}
	if len(data) > 0 {
		return rds.RedisV2.HMSet(key, data).Err()
	}
	return nil
}

func (rds *DB_r) SelectAreaPkgByPkgKey(kind static.AreaPackageKind, pkgKey string) (*static.AreaPackageCompiled, error) {
	apc := new(static.AreaPackageCompiled)
	err := goRedis.NewStringResult(
		rds.RedisV2.HGet(static.AreaPackageRedisKey(kind), pkgKey).Result(),
	).Scan(apc)
	return apc, err
}

func (rds *DB_r) SelectAreaPkgsByPkgKeys(kind static.AreaPackageKind, pkgKeys ...string) (static.AreaPkgCompiledList, error) {
	apcs := make(static.AreaPkgCompiledList, 0)
	err := goRedis.NewStringSliceResult(static.RedisSwitchObjectsToStrings(rds.RedisV2.HMGet(static.AreaPackageRedisKey(kind), pkgKeys...).Result())).ScanSlice(&apcs)
	return apcs, err
}

func (rds *DB_r) SelectAreaPkgByKindId(kind static.AreaPackageKind, kid int) (*static.AreaPackageCompiled, error) {
	apc := new(static.AreaPackageCompiled)
	pkg, err := rds.RedisV2.HGet(static.AreaKindIdRedisKey(kind), fmt.Sprint(kid)).Result()
	if err != nil {
		return apc, err
	}
	err = rds.RedisV2.HGet(static.AreaPackageRedisKey(kind), pkg).Scan(apc)
	// err := goRedis.NewStringResult(
	// 	public.RedisHGetHGet.Run(
	// 		rds.RedisV2,
	// 		[]string{public.AreaKindIdRedisKey(), public.AreaPackageRedisKey()},
	// 		kid).String()).Scan(apc)
	return apc, err
}

/*
func (self *DB_r) SelectAreaPkgsByKindIds(kid ...interface{}) (public.AreaPkgCompiledList, error) {
	apcs := make(public.AreaPkgCompiledList, 0)

	if len(kid) <= 0 {
		return apcs, nil
	}

	cmd := public.RedisHMGetHMGet.Run(self.RedisV2, []string{
		public.AreaKindIdRedisKey(),
		public.AreaPackageRedisKey(),
	}, kid...)

	if cmd.Err() != nil || cmd.Val() == nil {
		return apcs, cmd.Err()
	}

	err := goRedis.NewStringSliceResult(public.RedisSwitchObjectsToStrings(cmd.Val().([]interface{}), cmd.Err())).ScanSlice(&apcs)
	apcs.UnDuplicate()

	return apcs, err
}
*/

func (rds *DB_r) SelectAreaPkgsByCode(kind static.AreaPackageKind, code ...string) (static.AreaPkgCompiledList, error) {
	keys := make([]string, 0)
	for _, c := range code {
		keys = append(keys, static.AreaCodeRedisKey(kind, c))
	}
	apcs := make(static.AreaPkgCompiledList, 0)

	codeUnion, err := rds.RedisV2.SUnion(keys...).Result()

	if len(code) <= 0 {
		return apcs, nil
	}

	if err != nil {
		return apcs, err
	}

	err = goRedis.NewStringSliceResult(
		static.RedisSwitchObjectsToStrings(rds.RedisV2.HMGet(static.AreaPackageRedisKey(kind), codeUnion...).Result()),
	).ScanSlice(&apcs)

	// 对游戏包排序（区域包自带排序）
	for _, pkg := range apcs {
		for i := 0; i < len(pkg.Games)-1; i++ {
			for j := i + 1; j < len(pkg.Games); j++ {
				if pkg.Games[i].Sort > pkg.Games[j].Sort {
					var tmp *static.AreaGameCompiled
					tmp = pkg.Games[j]
					pkg.Games[j] = pkg.Games[i]
					pkg.Games[i] = tmp
				}
			}
		}
	}

	return apcs, err
}

func (rds *DB_r) SelectAppletAreaGamesByCode(code string) (*static.AppletAreaGamesCompiled, error) {
	key := static.AppletAreaGamesRedisKey(static.AreaPackageKindApplet)
	agcs := new(static.AppletAreaGamesCompiled)
	err := goRedis.NewStringResult(rds.RedisV2.HGet(key, code).Result()).Scan(agcs)
	return agcs, err
}

func (rds *DB_r) SelectAreaPkgsUniversal(kind static.AreaPackageKind) (static.AreaPkgCompiledList, error) {
	apcs := make(static.AreaPkgCompiledList, 0)
	codeUnion, err := rds.RedisV2.SMembers(static.AreaUniversalRedisKey(kind)).Result()
	if err != nil {
		return apcs, err
	}

	err = goRedis.NewStringSliceResult(
		static.RedisSwitchObjectsToStrings(rds.RedisV2.HMGet(static.AreaPackageRedisKey(kind), codeUnion...).Result()),
	).ScanSlice(&apcs)

	// 对游戏包排序（区域包自带排序）
	for _, pkg := range apcs {
		for i := 0; i < len(pkg.Games)-1; i++ {
			for j := i + 1; j < len(pkg.Games); j++ {
				if pkg.Games[i].Sort > pkg.Games[j].Sort {
					var tmp *static.AreaGameCompiled
					tmp = pkg.Games[j]
					pkg.Games[j] = pkg.Games[i]
					pkg.Games[i] = tmp
				}
			}
		}
	}

	return apcs, err
}

func (rds *DB_r) SelectAreaPkgsRecommend(kind static.AreaPackageKind) (static.AreaPkgCompiledList, error) {
	apcs := make(static.AreaPkgCompiledList, 0)
	codeUnion, err := rds.RedisV2.SMembers(static.AreaRecommendRedisKey(kind)).Result()
	if err != nil {
		return apcs, err
	}

	err = goRedis.NewStringSliceResult(
		static.RedisSwitchObjectsToStrings(rds.RedisV2.HMGet(static.AreaPackageRedisKey(kind), codeUnion...).Result()),
	).ScanSlice(&apcs)

	// 对游戏包排序（区域包自带排序）
	for _, pkg := range apcs {
		for i := 0; i < len(pkg.Games)-1; i++ {
			for j := i + 1; j < len(pkg.Games); j++ {
				if pkg.Games[i].Sort > pkg.Games[j].Sort {
					var tmp *static.AreaGameCompiled
					tmp = pkg.Games[j]
					pkg.Games[j] = pkg.Games[i]
					pkg.Games[i] = tmp
				}
			}
		}
	}

	return apcs, err
}

func (rds *DB_r) SelectAllAreaPkgsByKind(kind static.AreaPackageKind) (static.AreaPkgCompiledList, error) {
	apcs := make(static.AreaPkgCompiledList, 0)
	err := goRedis.NewStringSliceResult(rds.RedisV2.HVals(static.AreaPackageRedisKey(kind)).Result()).ScanSlice(&apcs)
	return apcs, err
}

func (rds *DB_r) SelectAllAreaPkgKeysByKind(kind static.AreaPackageKind) static.PackKeys {
	return rds.RedisV2.HKeys(static.AreaPackageRedisKey(kind)).Val()
}

func (rds *DB_r) SelectAllAreaGamesIdByKind(kind static.AreaPackageKind) static.PackKeys {
	return rds.RedisV2.HKeys(static.AreaKindIdRedisKey(kind)).Val()
}

func (rds *DB_r) SelectAllAreaCodesByKind(kind static.AreaPackageKind) static.PackKeys {
	return rds.RedisV2.SMembers(static.AreaCodeListRedisKey(kind)).Val()
}

func (rds *DB_r) SelectAllAreaPkgs() static.AreaPkgCompiledList {
	cardApcs := make(static.AreaPkgCompiledList, 0)
	err := goRedis.NewStringSliceResult(rds.RedisV2.HVals(static.AreaPackageRedisKey(static.AreaPackageKindCard)).Result()).ScanSlice(&cardApcs)
	if err != nil {
		xlog.Logger().Error(err)
	}
	goldApcs := make(static.AreaPkgCompiledList, 0)
	err = goRedis.NewStringSliceResult(rds.RedisV2.HVals(static.AreaPackageRedisKey(static.AreaPackageKindGold)).Result()).ScanSlice(&goldApcs)
	if err != nil {
		xlog.Logger().Error(err)
	}
	res := append(cardApcs, goldApcs...)
	res.UnDuplicate()
	return res
}

func (rds *DB_r) SelectAllAreaPkgKeys() static.PackKeys {
	return append(rds.RedisV2.HKeys(static.AreaPackageRedisKey(static.AreaPackageKindCard)).Val(),
		rds.RedisV2.HKeys(static.AreaPackageRedisKey(static.AreaPackageKindGold)).Val()...)
}

func (rds *DB_r) SelectAllAreaGamesId() static.PackKeys {
	return append(rds.RedisV2.HKeys(static.AreaKindIdRedisKey(static.AreaPackageKindCard)).Val(),
		rds.RedisV2.HKeys(static.AreaKindIdRedisKey(static.AreaPackageKindGold)).Val()...)
}

func (rds *DB_r) SelectAllAreaCodes() static.PackKeys {
	return append(rds.RedisV2.SMembers(static.AreaCodeListRedisKey(static.AreaPackageKindCard)).Val(),
		rds.RedisV2.SMembers(static.AreaCodeListRedisKey(static.AreaPackageKindGold)).Val()...)
}

func (rds *DB_r) SelectAreaWxByCode(code string) *static.AreaGmWeChatInfo {
	wx := new(static.AreaGmWeChatInfo)
	err := goRedis.NewStringResult(rds.RedisV2.HGet(static.AreaWeChatRedisKey(), code).Result()).Scan(wx)
	if err == nil {
		return wx
	}
	return static.GenAreaDefaultGMWeChat(code)
}

func (rds *DB_r) CheckAreaCodeExist(kind static.AreaPackageKind, code string) bool {
	return rds.RedisV2.Exists(static.AreaCodeRedisKey(kind, code)).Val() == 1
}

type DictionaryInt64Toi map[int64]int

func (i64 *DictionaryInt64Toi) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, i64)
}

func (i64 *DictionaryInt64Toi) MarshalBinary() (data []byte, err error) {
	return json.Marshal(i64)
}

func (rds *DB_r) GetTeamMemGameRecord(dHid int64, uid int64) (map[int64]int, error) {
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_TEAM, dHid)
	record := make(DictionaryInt64Toi)
	err := rds.RedisV2.HGet(key, fmt.Sprint(uid)).Scan(&record)
	if err != nil {
		if err == goRedis.Nil {
			return record, nil
		}
		return nil, err
	}
	return record, nil
}

func (rds *DB_r) SetTeamMemGameRecord(dHid int64, uid int64, record map[int64]int) error {
	res := DictionaryInt64Toi(record)
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_TEAM, dHid)
	exists := rds.RedisV2.Exists(key).Val() == 1
	err := rds.RedisV2.HSet(key, fmt.Sprint(uid), &res).Err()
	if err != nil {
		return err
	}
	if !exists {
		err := rds.RedisV2.Expire(key, time.Second*time.Duration(static.HF_GetTodayRemainSecond())).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rds *DB_r) GetMemCribberList(dHid int64, uid int64) (map[int64]int, error) {
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_CRIBBER, dHid)
	record := make(DictionaryInt64Toi)
	err := rds.RedisV2.HGet(key, fmt.Sprint(uid)).Scan(&record)
	if err != nil {
		if err == goRedis.Nil {
			return record, nil
		}
		return nil, err
	}
	return record, nil
}

func (rds *DB_r) GetMultiMemCribberList(dHid int64, uid ...int64) (map[int64]map[int64]int, error) {
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_CRIBBER, dHid)
	res, err := rds.RedisV2.HMGet(key, static.HF_I64stoas(uid...)...).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println(res)
	if len(res) == len(uid) {
		result := make(map[int64]map[int64]int)
		for k, v := range uid {
			r := make(map[int64]int)
			if data, ok := res[k].(string); ok {
				err = json.Unmarshal([]byte(data), &r)
				if err != nil {
					return nil, err
				}
			}
			result[v] = r
		}
		return result, nil
	}
	return nil, fmt.Errorf("an unequal result")
}

func (rds *DB_r) SetMemCribberList(dHid int64, uid int64, record map[int64]int) error {
	res := DictionaryInt64Toi(record)
	for k, v := range record {
		res[k] = v
	}
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_CRIBBER, dHid)
	return rds.RedisV2.HSet(key, fmt.Sprint(uid), &res).Err()
}

func (rds *DB_r) SetMultiMemCribberList(dHid int64, record map[int64]map[int64]int) error {
	res := make(map[string]interface{})
	for k, v := range record {
		val := DictionaryInt64Toi(v)
		res[fmt.Sprint(k)] = &val
	}
	key := fmt.Sprintf(consts.REDIS_KEY_HOUSE_CRIBBER, dHid)
	return rds.RedisV2.HMSet(key, res).Err()
}

//! 删除单局战绩数据
func (rds *DB_r) DeleteGameRecord(GameNum string, uid int64) error {
	key := fmt.Sprintf(consts.REDIS_KEY_GAMENUM, GameNum, uid)

	return rds.RedisV2.Expire(key, time.Second*1).Err()
}

//向redis中插入贡献值系统配置
func (rds *DB_r) ContributionSystemConfigInsert(config *models.ConfigContributionSystem) error {
	key := "contributionSystem"
	// filed
	filed := "contributionSystem_config"
	// val
	val := static.HF_JtoB(config)

	return rds.RedisV2.HSet(key, filed, val).Err()
}

//向redis中插入游戏房间贡献值配置
func (rds *DB_r) ContributionSystemGameInsert(config *models.ConfigContributionGame) error {
	key := "contributionSystem"
	// filed
	filed := fmt.Sprintf(consts.REDIS_KEY_CONTRIBUTION, config.KindId, config.SiteType)
	// val
	val := static.HF_JtoB(config)

	return rds.RedisV2.HSet(key, filed, val).Err()
}

//向redis中插入新玩家强补控制配置
func (rds *DB_r) ContributionSystemNewPlayerInsert(config *models.ConfigContributionNewPlayer) error {
	key := "contributionSystem"
	// filed
	filed := "contributionNewPlayer_config"
	// val
	val := static.HF_JtoB(config)

	return rds.RedisV2.HSet(key, filed, val).Err()
}

//向redis中插入商城商品充值转换贡献值配置
func (rds *DB_r) ContributionSystemShopInsert(config *models.ConfigContributionShop) error {
	key := "contributionSystem"
	// filed
	filed := "contributionShop_config"
	// val
	val := static.HF_JtoB(config)

	return rds.RedisV2.HSet(key, filed, val).Err()
}

//向redis中插入日净分控制配置
func (rds *DB_r) ContributionSystemDayScoreInsert(config *models.ConfigContributionScore) error {
	key := "contributionSystem"
	// filed
	filed := "contributionDayScore_config"
	// val
	val := static.HF_JtoB(config)

	return rds.RedisV2.HSet(key, filed, val).Err()
}

//向redis中插入贡献值系统配置
func (rds *DB_r) GetContributionSystemConfig() *models.ConfigContributionSystem {
	key := "contributionSystem"
	// filed
	filed := "contributionSystem_config"

	config := new(models.ConfigContributionSystem)
	if bytes, err := rds.RedisV2.HGet(key, filed).Bytes(); err == nil {
		if err := json.Unmarshal(bytes, config); err == nil {
			return config
		}
	}

	return nil
}

//向redis中插入游戏房间贡献值配置
func (rds *DB_r) GetContributionSystemGameConfig(kindid int, siteType int) *models.ConfigContributionGame {
	key := "contributionSystem"
	filed := fmt.Sprintf(consts.REDIS_KEY_CONTRIBUTION, kindid, siteType)

	config := new(models.ConfigContributionGame)
	if bytes, err := rds.RedisV2.HGet(key, filed).Bytes(); err == nil {
		if err := json.Unmarshal(bytes, config); err == nil {
			return config
		}
	}

	return nil
}

//向redis中插入新玩家强补控制配置
func (rds *DB_r) GetContributionSystemNewPlayerConfig() *models.ConfigContributionNewPlayer {
	key := "contributionSystem"
	filed := "contributionNewPlayer_config"

	config := new(models.ConfigContributionNewPlayer)
	if bytes, err := rds.RedisV2.HGet(key, filed).Bytes(); err == nil {
		if err := json.Unmarshal(bytes, config); err == nil {
			return config
		}
	}

	return nil
}

//向redis中插入日净分控制配置
func (rds *DB_r) GetContributionSystemDayScoreConfig() *models.ConfigContributionScore {
	key := "contributionSystem"
	filed := "contributionDayScore_config"

	config := new(models.ConfigContributionScore)
	if bytes, err := rds.RedisV2.HGet(key, filed).Bytes(); err == nil {
		if err := json.Unmarshal(bytes, config); err == nil {
			return config
		}
	}

	return nil
}

func (rds *DB_r) GetContributionSystemShopConfig() *models.ConfigContributionShop {
	key := "contributionSystem"
	filed := "contributionShop_config"

	config := new(models.ConfigContributionShop)
	if bytes, err := rds.RedisV2.HGet(key, filed).Bytes(); err == nil {
		if err := json.Unmarshal(bytes, config); err == nil {
			return config
		}
	}

	return nil
}

// getTable
func (rds *DB_r) GetTableInfo(gameId, tid int) (*static.Table, error) {
	key := fmt.Sprintf(consts.REDIS_KEY_TABLEINFO, gameId, tid)
	var table static.Table
	err := rds.RedisV2.Get(key).Scan(&table)
	if err != nil {
		return nil, err
	}
	return &table, nil
}

// getTableBy pattern
func (rds *DB_r) GetTableInfoPattern(tid int) (*static.Table, error) {
	pattern := fmt.Sprintf(consts.REDIS_KEY_TABLEINFO_ALL_1, tid)
	keys, err := rds.RedisV2.Keys(pattern).Result()
	if err != nil {
		return nil, err
	} else {
		if len(keys) > 0 {
			key := keys[0] // 始终取匹配到的第一个
			var table static.Table
			err = rds.RedisV2.Get(key).Scan(&table)
			if err != nil {
				return nil, err
			} else {
				return &table, nil
			}
		}
	}
	return nil, goRedis.Nil
}

//! 获取Robot信息
func (rds *DB_r) GetRobot(uid int64) (*static.Robot, error) {
	r := new(static.Robot)
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_ROBOTS, uid)
	err := rds.m_redis.Hgetall(key, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

//! 获取Person
func (rds *DB_r) GetRobotByCode(code string) (*static.Robot, error) {
	r := new(static.Robot)
	robots, err := rds.GetRobots()
	if err != nil {
		return nil, err
	}

	for _, v := range robots {
		if v.MachineCode == code {
			r = v
		}
	}

	return r, nil
}

func (rds *DB_r) ResetRobot(r *static.Robot) error {
	//先删
	if ok, err := rds.DelRobot(r.Mid); !ok {
		return err
	}
	//再建
	return rds.CreateRobot(r)
}

//! 获取redis中所有已创建的Robot
func (rds *DB_r) GetRobots() ([]*static.Robot, error) {
	robotSlice := make([]*static.Robot, 0)
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_ROBOTS_ALL)
	bytes, err := rds.m_redis.Keys(key)
	if err != nil {
		return nil, err
	}

	for _, key := range bytes {
		r := new(static.Robot)
		err = rds.m_redis.Hgetall(static.HF_Bytestoa(key), r)
		if err != nil {
			return nil, err
		}
		robotSlice = append(robotSlice, r)
	}

	if len(robotSlice) > 0 {
		return robotSlice, nil
	}

	return nil, err
}

//! 获取redis中所有已创建的Robot
func (rds *DB_r) GetTableRobots(tableid int) ([]*static.Robot, error) {
	robotSlice := make([]*static.Robot, 0)
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_ROBOTS_ALL)
	bytes, err := rds.m_redis.Keys(key)
	if err != nil {
		return nil, err
	}

	for _, key := range bytes {
		r := new(static.Robot)
		err = rds.m_redis.Hgetall(static.HF_Bytestoa(key), r)
		if err != nil {
			return nil, err
		}
		if r.TableId == tableid {
			robotSlice = append(robotSlice, r)
		}
	}

	if len(robotSlice) > 0 {
		return robotSlice, nil
	}

	return nil, err
}

//获取redis中已被使用但是没在工作的robots
func (rds *DB_r) GetFreeRobots() ([]*static.Robot, error) {
	robotSlice, err := rds.GetRobots()
	if err != nil {
		return nil, err
	}

	if len(robotSlice) == 0 {
		return nil, err
	}
	usedRobots := make([]*static.Robot, 0)
	for _, r := range robotSlice {
		if r.IsWorking == 0 {
			usedRobots = append(usedRobots, r)
		}
	}

	return usedRobots, nil
}

//! 创建Robot
func (rds *DB_r) CreateRobot(r *static.Robot) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_ROBOTS, r.Mid)
	return rds.m_redis.HMsetall(key, r)
}

// 更新robot属性
func (rds *DB_r) UpdateRobotAttrs(uid int64, args ...interface{}) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_ROBOTS, uid)
	return rds.m_redis.HMset(key, args...)
}

//！根据id删除robot
func (rds *DB_r) DelRobot(uid int64) (bool, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_ROBOTS, uid)
	return rds.m_redis.Del(key)
}

//! 插入大厅个人历史战绩数据
func (rds *DB_r) InsertNormalHallRecord(uid int64, gameRecord *static.GameRecordHistory) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_HALL_RECORD, uid)
	// 先获取长度
	len, err := rds.m_redis.Llen(key)
	if err != nil {
		return err
	}
	// 超过8条数据删除最旧的一条记录
	if len >= 8 {
		_, err = rds.m_redis.Rpop(key)
		if err != nil {
			return err
		}
	}

	// val
	val := static.HF_JtoB(gameRecord)
	return rds.m_redis.Lpush(key, val)
}

//! 插入每日排位赛统计数据到天数据中
func (rds *DB_r) InsertGameMatchTotalSet(gameMatchTotal *models.GameMatchTotal) error {
	// key
	key := fmt.Sprintf("%s", gameMatchTotal.MatchKey)
	// field
	field := fmt.Sprintf("%d", gameMatchTotal.UId)
	// val
	val := static.HF_JtoB(gameMatchTotal)

	rds.m_redis.HSet(key, field, val)

	//更新排名
	rds.GameMatchRankingInsert(gameMatchTotal)
	// 计算过期时间
	//expireTime := playTime.Unix() + public.SEVEN_DAY_SECONDS - time.Now().Unix()
	//if expireTime <= 0 {
	//	expireTime = 1
	//}
	//! 设置改数据七天过期
	//return rds.m_redis.Expire(key, int(expireTime))
	return nil
}

//! 查询单个玩家每次排位统计数据
func (rds *DB_r) SelectGameMatchTotalSet(gameMatchTotal *models.GameMatchTotal) (*models.GameMatchTotal, error) {
	// key
	key := fmt.Sprintf("%s", gameMatchTotal.MatchKey)
	// field
	field := fmt.Sprintf("%d", gameMatchTotal.UId)
	// val
	bytes, err := rds.m_redis.HGet(key, field)
	if err != nil {
		return nil, err
	}

	var gameMatchTotalr models.GameMatchTotal
	err = json.Unmarshal([]byte(bytes), &gameMatchTotalr)
	if err != nil {
		return nil, err
	}
	return &gameMatchTotalr, nil
}

//! 排位结束时更新排位赛的排位名次
func (rds *DB_r) InsertGameMatchTotalSetByranking(gameMatchTotal *models.GameMatchTotal) error {
	// key
	key := fmt.Sprintf("%s", gameMatchTotal.MatchKey)
	// field
	field := fmt.Sprintf("%d", gameMatchTotal.UId)
	// val
	val := static.HF_JtoB(gameMatchTotal)

	rds.m_redis.HSet(key, field, val)

	//更新排名
	rds.GameMatchRankingInsertByRanking(gameMatchTotal)
	// 计算过期时间
	//expireTime := playTime.Unix() + public.SEVEN_DAY_SECONDS - time.Now().Unix()
	//if expireTime <= 0 {
	//	expireTime = 1
	//}
	//! 设置改数据七天过期
	//return rds.m_redis.Expire(key, int(expireTime))
	return nil
}

//排位赛结束后的排位
func (rds *DB_r) GameMatchRankingInsertByRanking(gameMatchTotal *models.GameMatchTotal) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING_MATCH_OVER, gameMatchTotal.MatchKey)
	// val
	val := gameMatchTotal.UId
	// score
	score := gameMatchTotal.Ranking
	//insert
	_, err := rds.m_redis.Zadd(key, static.HF_I64tobytes(val), float64(score))
	return err
}

//! 查询排位赛最终的排位
func (rds *DB_r) GameMatchRankingListByranking(matchKey string, ibeg int, iend int) ([]int64, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING_MATCH_OVER, matchKey)
	datas, err := rds.m_redis.Zrange(key, ibeg, iend)
	if err != nil {
		return nil, err
	}

	var res []int64
	for _, data := range datas {
		res = append(res, static.HF_Bytestoi64(data))
	}

	var matchRecordItems []int64
	for _, uid := range res {
		matchRecordItems = append(matchRecordItems, uid)
	}
	return matchRecordItems, nil
}

//! 查询排位赛临时排位
func (rds *DB_r) GameMatchRankingList(matchKey string, ibeg int, iend int) ([]int64, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING, matchKey)
	datas, err := rds.m_redis.Zrevrange(key, ibeg, iend)
	if err != nil {
		return nil, err
	}

	var res []int64
	for _, data := range datas {
		res = append(res, static.HF_Bytestoi64(data))
	}

	var matchRecordItems []int64
	for _, uid := range res {
		matchRecordItems = append(matchRecordItems, uid)
	}
	return matchRecordItems, nil
}

//排位赛进行中时的临时排位
func (rds *DB_r) GameMatchRankingInsert(gameMatchTotal *models.GameMatchTotal) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING, gameMatchTotal.MatchKey)
	// val
	val := gameMatchTotal.UId
	// score
	coupon := gameMatchTotal.Coupon
	//insert
	_, err := rds.m_redis.Zadd(key, static.HF_I64tobytes(val), float64(coupon))
	return err
}

func (rds *DB_r) GameMatchRankingDelete(gameMatchTotal *models.GameMatchTotal) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING, gameMatchTotal.MatchKey)
	// val
	val := gameMatchTotal.UId
	//delete
	_, err := rds.m_redis.Zrem(key, static.HF_I64tobytes(val))
	return err
}

//! 查询排位赛排位
func (rds *DB_r) GameMatchRankingRecordList(matchKey string, ibeg int, iend int) ([]int64, error) {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING_OVER_BY_RANKING, matchKey)
	datas, err := rds.m_redis.Zrevrange(key, ibeg, iend)
	if err != nil {
		return nil, err
	}

	var res []int64
	for _, data := range datas {
		res = append(res, static.HF_Bytestoi64(data))
	}

	var matchRecordItems []int64
	for _, uid := range res {
		matchRecordItems = append(matchRecordItems, uid)
	}
	return matchRecordItems, nil
}

//! 排位结束时插入排位赛的排位名次
func (rds *DB_r) InsertGameMatchRankingRecordSet(gameMatchCouponRecord *models.GameMatchCouponRecord) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING_OVER_RECORD, gameMatchCouponRecord.MatchKey)
	// field
	field := fmt.Sprintf("%d", gameMatchCouponRecord.UId)
	// val
	val := static.HF_JtoB(gameMatchCouponRecord)

	rds.m_redis.HSet(key, field, val)

	//更新排名
	rds.GameMatchRankingRecordInsert(gameMatchCouponRecord)
	// 计算过期时间
	//expireTime := playTime.Unix() + public.SEVEN_DAY_SECONDS - time.Now().Unix()
	//if expireTime <= 0 {
	//	expireTime = 1
	//}
	//! 设置改数据七天过期
	//return rds.m_redis.Expire(key, int(expireTime))
	return nil
}

func (rds *DB_r) GameMatchRankingRecordInsert(gameMatchCouponRecord *models.GameMatchCouponRecord) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_RANKING_OVER_BY_RANKING, gameMatchCouponRecord.MatchKey)
	// val
	val := gameMatchCouponRecord.UId
	// score
	score := gameMatchCouponRecord.Ranking
	//insert
	_, err := rds.m_redis.Zadd(key, static.HF_I64tobytes(val), float64(score))
	return err
}

// 更新match属性
func (rds *DB_r) UpdateMatchAttrs(matchconfig *models.ConfigMatch, state int) error {
	// key
	key := fmt.Sprintf(consts.REDIS_KEY_MATCH, matchconfig.KindId, matchconfig.TypeStr) + matchconfig.BeginDate.Format("20060102") + matchconfig.BeginTime.Format("150405")
	val := state
	return rds.m_redis.HMset(key, static.HF_Itobytes(val))
}

// 得到低保礼包配置
func (rds *DB_r) GetAllowanceGift() (*static.AllowanceGift, error) {
	ret := new(static.AllowanceGift)
	err := rds.RedisV2.Get(consts.REDIS_KEY_ALLOWANCE_GIFT).Scan(ret)
	return ret, err
}

//! 添加牌桌观战
func (rds *DB_r) AddWatchPlayerToTable(tableId int, uid int64) error {
	key := fmt.Sprintf(consts.REDIS_KEY_TABLE_WATCH, tableId)
	_, err := rds.RedisV2.LPush(key, uid).Result()
	if err != nil {
		xlog.Logger().Errorf("AddWatchPlayerToTable LPush Err, tableid = %d, uid = %d", tableId, uid)
	}
	// 设置玩家为观战人员
	err = rds.UpdatePersonAttrs(uid, "WatchTable", tableId)
	if err != nil {
		xlog.Logger().Errorf("AddWatchPlayerToTable Update Err, tableid = %d, uid = %d", tableId, uid)
	}
	return err
}

//! 删除牌桌观战玩家
func (rds *DB_r) RemoveWatchPlayerToTable(tableId int, uid int64) error {
	key := fmt.Sprintf(consts.REDIS_KEY_TABLE_WATCH, tableId)
	_, err := rds.RedisV2.LRem(key, 1, uid).Result()
	if err != nil {
		xlog.Logger().Errorf("RemoveWatchPlayerToTable LRem Err, tableid = %d, uid = %d", tableId, uid)
	}
	// 设置玩家为观战人员
	err = rds.UpdatePersonAttrs(uid, "WatchTable", 0)
	if err != nil {
		xlog.Logger().Errorf("RemoveWatchPlayerToTable Update Err, tableid = %d, uid = %d", tableId, uid)
	}
	return err
}

//! 删除观战牌桌
func (rds *DB_r) RemoveWatchTable(tableId int) error {
	key := fmt.Sprintf(consts.REDIS_KEY_TABLE_WATCH, tableId)
	_, err := rds.RedisV2.Del(key).Result()
	return err
}

//! 获取观战列表
func (rds *DB_r) GetWatchTablePlayer(tableId int) []int64 {
	var list []int64
	key := fmt.Sprintf(consts.REDIS_KEY_TABLE_WATCH, tableId)
	result, err := rds.RedisV2.LRange(key, 0, -1).Result()
	if err != nil {
		return list
	}
	for _, strVaule := range result {
		intValue := static.HF_Atoi64(strVaule)
		list = append(list, intValue)
	}
	return list
}
