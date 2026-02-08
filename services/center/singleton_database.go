// ! 数据库底层
package center

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/dao"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// //////////////////////////////////////////////////////////////////////////////
// ! 数据结构
type DBMgr struct {
	db_R      *dao.DB_r      //! redis操作
	db_M      *dao.ORM_Mysql //! mysql操作
	Redis     *redis.Client
	PubRedis  *redis.Client
	RedisLock *redis.Client // 分布式锁的redis库
}

var DBMgrSingleton *DBMgr = nil

// ! 得到包厢管理单例
func GetDBMgr() *DBMgr {

	if DBMgrSingleton == nil {

		DBMgrSingleton = new(DBMgr)
		DBMgrSingleton.db_R = new(dao.DB_r)
		DBMgrSingleton.db_M = new(dao.ORM_Mysql)

		con := GetServer().Con
		//! redis
		err := DBMgrSingleton.db_R.Init(con.Redis, con.RedisDB, con.RedisAuth)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("init redis error: %s", err.Error()))
		}
		// 初始化v2版本redis库
		DBMgrSingleton.Redis = static.InitRedisV2(con.Redis, con.RedisDB, con.RedisAuth)
		// 初始化v2版本redis库
		DBMgrSingleton.PubRedis = static.InitRedisV2(con.PubRedis, con.PubRedisDB, con.PubRedisAuth)
		DBMgrSingleton.RedisLock = static.InitRedisV2(con.Redis, con.RedisDB+1, con.RedisAuth)
		//! mysql
		_, err = DBMgrSingleton.db_M.Open("mysql", con.DB)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("init mysql error: %s", err.Error()))
		}

		if err = DBMgrSingleton.ReadAllConfig(); err != nil {
			xlog.Logger().Errorln("DBMgrSingleton.ReadAllConfig:", err)
			return nil
		}
	}
	return DBMgrSingleton
}

func (self *DBMgr) GetDBmControl() *gorm.DB {
	return self.db_M.GetConn()
}

func (self *DBMgr) GetDBrControl() *dao.DB_r {
	return self.db_R
}

func (self *DBMgr) Close() error {

	if self.db_R != nil {
		self.db_R.Close()
	}

	if self.db_M != nil {
		self.db_M.Close()
	}
	return nil
}

func (self *DBMgr) DataInitialize() error {

	// 初始化 依赖 各功能控制器 DB->redis->内存

	// house
	// redis
	hdatas, err := self.GetDBrControl().HouseBaseReLoad()
	if err != nil {
		return err
	}
	// memory
	GetClubMgr().HouseBaseReload(hdatas)

	//housefloor
	// redis
	hfdatas, err := self.GetDBrControl().HouseFloorBaseReLoad()
	if err != nil {
		return err
	}
	// 内存
	GetClubMgr().HouseFloorBaseReload(hfdatas)
	// GetClubMgr().InitHouseVitaminPool() //初始化包厢疲劳值仓库,一生只一次
	// redis
	ftdatas, err := self.GetDBrControl().FloorTableBaseReLoad()
	if err != nil {
		return err
	}
	// 内存
	GetClubMgr().FloorTableBaseReload(ftdatas)

	// 内存

	// 包厢区域检测
	GetClubMgr().HouseAreaCheck()
	GetClubMgr().InitHouseTableLimit()
	GetClubMgr().InitHouseGroupUser()
	GetClubMgr().InitFiveGroup()
	// GetClubMgr().CheckDefault()
	return nil
}

// ! 创建包厢
func (self *DBMgr) HouseInsert(house *Club, tx *gorm.DB) (int64, error) {

	// config := GetServer().ConHouse

	db_house := house.DBClub
	var err error
	//mysql auto set id
	if tx != nil {
		if err = tx.Create(db_house).Error; err != nil {
			return 0, err
		}

		// 日志
		db_house_log := new(models.HouseLog)
		db_house_log.DHId = db_house.Id
		db_house_log.HId = db_house.HId
		db_house_log.UId = db_house.UId
		db_house_log.Area = db_house.Area
		db_house_log.Type = consts.OPTION_INSERT
		if err = tx.Create(db_house_log).Error; err != nil {
			return 0, err
		}
	} else {
		if err = self.GetDBmControl().Create(db_house).Error; err != nil {
			return 0, err
		}

		// 日志
		db_house_log := new(models.HouseLog)
		db_house_log.DHId = db_house.Id
		db_house_log.HId = db_house.HId
		db_house_log.UId = db_house.UId
		db_house_log.Area = db_house.Area
		db_house_log.Type = consts.OPTION_INSERT
		if err = self.GetDBmControl().Create(db_house_log).Error; err != nil {
			return 0, err
		}
	}

	//redis
	err = self.db_R.HouseInsert(db_house)
	if err != nil {
		return 0, err
	}

	return db_house.Id, nil
}

// ! 更新包厢
func (self *DBMgr) HouseUpdate(house *Club) error {

	// config := GetServer().ConHouse

	db_house := house.DBClub
	err := self.db_R.HouseInsert(db_house)
	if err != nil {
		return err
	}
	return nil
}

// ! 创建包厢
func (self *DBMgr) HouseDelete(house *Club, tx *gorm.DB) error {

	db_house := house.DBClub

	var err error
	//mysql auto set id
	if tx != nil {
		if err = tx.Exec(`delete from house where id = ?`, house.DBClub.Id).Error; err != nil {
			return err
		}

		// 日志
		db_house_log := new(models.HouseLog)
		db_house_log.DHId = house.DBClub.Id
		db_house_log.Area = house.DBClub.Area
		db_house_log.HId = house.DBClub.HId
		db_house_log.UId = house.DBClub.UId
		db_house_log.Type = consts.OPTION_DELETE
		if err = tx.Create(db_house_log).Error; err != nil {
			return err
		}
	} else {
		if err = self.GetDBmControl().Exec(`delete from house where id = ?`, house.DBClub.Id).Error; err != nil {
			return err
		}

		// 日志
		db_house_log := new(models.HouseLog)
		db_house_log.DHId = house.DBClub.Id
		db_house_log.HId = house.DBClub.HId
		db_house_log.Area = house.DBClub.Area
		db_house_log.UId = house.DBClub.UId
		db_house_log.Type = consts.OPTION_DELETE
		if err = self.GetDBmControl().Create(db_house_log).Error; err != nil {
			return err
		}
	}
	house.delPrize()
	//redis
	err = self.db_R.HouseDelete(db_house)
	if err != nil {
		return err
	}

	return nil
}

// ! 创建包厢楼层
func (self *DBMgr) HouseFloorCreate(floor *HouseFloor, tx *gorm.DB) (int64, error) {
	db_floor := new(models.HouseFloor)
	db_floor.DHId = floor.DHId
	data, err := json.Marshal(&floor.Rule)
	if err != nil {
		return 0, err
	}
	db_floor.Rule = string(data[:])
	db_floor.IsVitamin = floor.IsVitamin
	db_floor.IsGamePause = floor.IsGamePause
	db_floor.VitaminLowLimit = floor.VitaminLowLimit
	db_floor.VitaminHighLimit = floor.VitaminHighLimit
	db_floor.VitaminLowLimitPause = floor.VitaminLowLimitPause
	db_floor.IsMix = floor.IsMix
	db_floor.IsVip = floor.IsVip
	db_floor.IsCapSetVip = floor.IsCapSetVip
	db_floor.IsDefJoinVip = floor.IsDefJoinVip
	//db_floor.VitaminDeductCount = floor.VitaminDeductCount
	//db_floor.VitaminDeductType = floor.VitaminDeductType
	//db_floor.VitaminLowest = floor.VitaminLowest
	//db_floor.VitaminHighest = floor.VitaminHighest
	//db_floor.VitaminLowestDeduct = floor.VitaminLowestDeduct
	//db_floor.VitaminHighestDeduct = floor.VitaminHighestDeduct
	//db_floor.IsVitaminHighest = floor.IsVitaminHighest
	//db_floor.IsVitaminLowest = floor.IsVitaminLowest

	//mysql
	if tx != nil {
		if err = tx.Create(db_floor).Error; err != nil {
			return 0, err
		}
		payInfo := models.NewHouseFloorGearPay(db_floor.DHId, db_floor.Id, floor.Rule.PlayerNum)
		if err = tx.Create(payInfo).Error; err != nil {
			return 0, err
		}
	} else {
		if err = self.GetDBmControl().Create(db_floor).Error; err != nil {
			return 0, err
		}
		payInfo := models.NewHouseFloorGearPay(db_floor.DHId, db_floor.Id, floor.Rule.PlayerNum)
		if err = self.GetDBmControl().Create(payInfo).Error; err != nil {
			return 0, err
		}
	}

	//redis
	err = self.db_R.HouseFloorInsert(db_floor)
	if err != nil {
		return 0, nil
	}

	return db_floor.Id, nil
}

// ! 删除楼层
func (self *DBMgr) HouseFloorDelete(floor *HouseFloor, tx *gorm.DB) error {
	db_floor := new(models.HouseFloor)
	db_floor.Id = floor.Id

	var err error
	//mysql
	if tx != nil {
		if err = tx.Exec(`delete from house_floor where id = ?`, floor.Id).Error; err != nil {
			return err
		}
	} else {
		if err = self.GetDBmControl().Exec(`delete from house_floor where id = ?`, floor.Id).Error; err != nil {
			return err
		}
	}

	//redis
	err = self.db_R.HouseFloorDelete(floor.DHId, floor.Id)
	if err != nil {
		return err
	}

	self.db_R.RedisV2.Del(floor.RedisKeyVipUsers())
	return nil
}

// ! 更新楼层
func (self *DBMgr) HouseFloorUpdate(floor *HouseFloor) error {
	dbFloor, err := floor.ConvertModel()
	if err != nil {
		return err
	}
	return self.db_M.Transaction(func(tx *gorm.DB) error {
		err = tx.Omit("created_at").Save(dbFloor).Error
		if err != nil {
			return err
		}

		payInfo := new(models.HouseFloorGearPay)
		err = tx.Where("id = ?", floor.Id).First(payInfo).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				payInfo = models.NewHouseFloorGearPay(floor.DHId, floor.Id, floor.Rule.PlayerNum)
				return tx.Create(payInfo).Error
			}
			return err
		}
		if floor.Rule.PlayerNum != payInfo.PlayerNum {
			log := payInfo.GenLog()
			payInfo.PlayerNum = floor.Rule.PlayerNum

			err = tx.Save(payInfo).Error
			if err != nil {
				return err
			}

			err = tx.Create(log).Error
			if err != nil {
				return err
			}
			if nowBaseCost := payInfo.BaseCost(); nowBaseCost != log.BaseCost() {
				memMap := self.GetHouseMemMap(payInfo.DHId)
				err = UpdateTopPartnersTotal(tx, memMap, payInfo.DHId, payInfo.FId, nowBaseCost)
				if err != nil {
					return err
				}
			}
		}

		// 最后写redis
		err = self.db_R.HouseFloorInsert(dbFloor)
		if err != nil {
			return err
		}
		return nil
	})
}

// ! 创建包厢活动
func (self *DBMgr) HouseActivityExist(dhid int64, actid int64) *static.HouseActivity {

	hact, _ := GetDBMgr().GetDBrControl().HouseActivityInfo(dhid, actid)
	return hact
}

// ! 创建包厢活动
func (self *DBMgr) HouseActivityCreate(fActivity *static.HouseActivity, tx *gorm.DB) (int64, error) {

	db_act := new(models.HouseActivity)
	db_act.DHId = fActivity.DHId
	db_act.FId = fActivity.FId
	db_act.Kind = fActivity.Kind
	db_act.Name = fActivity.Name
	db_act.Status = fActivity.Status
	db_act.HideInfo = fActivity.HideInfo
	db_act.BegTime = time.Unix(fActivity.BegTime, 0)
	db_act.EndTime = time.Unix(fActivity.EndTime, 0)
	db_act.Type = fActivity.Type

	//mysql
	if tx != nil {
		if err := tx.Create(db_act).Error; err != nil {
			return 0, err
		}
	} else {
		if err := self.GetDBmControl().Create(db_act).Error; err != nil {
			return 0, err
		}
	}

	fActivity.Id = db_act.Id

	//redis
	err := self.db_R.HouseActivityInsert(fActivity)
	if err != nil {
		return 0, nil
	}

	return db_act.Id, nil
}

// ! 创建包厢活动
func (self *DBMgr) HouseActivityDelete(dhid int64, actid int64) error {

	hact, err := self.GetDBrControl().HouseActivityInfo(dhid, actid)
	if err != nil {
		return err
	}
	// 活动数据
	hactrecord, err := self.GetDBrControl().HouseActivityRecordList(actid, 0, -1)
	if err != nil {
		return err
	}

	// 活动
	db_act := new(models.HouseActivity)
	db_act.Id = hact.Id
	// 活动日志
	db_act_log := new(models.HouseActivityLog)
	db_act_log.ActId = hact.Id
	db_act_log.DHId = hact.DHId
	db_act_log.FId = hact.FId
	db_act_log.Kind = hact.Kind
	db_act_log.Name = hact.Name
	db_act_log.Status = hact.Status
	db_act_log.BegTime = time.Unix(hact.BegTime, 0)
	db_act_log.EndTime = time.Unix(hact.EndTime, 0)

	tx := self.GetDBmControl().Begin()
	if tx != nil {
		// 删除活动
		if err := tx.Exec(`delete from house_activity where id = ?`, hact.Id).Error; err != nil {
			tx.Rollback()
			return err
		}
		// 新增活动日志
		if err = tx.Create(db_act_log).Error; err != nil {
			tx.Rollback()
			return err
		}
		// 删除活动数据
		if err = tx.Where("actid = ?", db_act.Id).Delete(models.HouseActivityRecord{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		// 新增活动日志
		for _, record := range hactrecord {
			// 活动数据
			db_act_record := new(models.HouseActivityRecordLog)
			db_act_record.ActId = db_act.Id
			db_act_record.UId = record.UId
			db_act_record.RankScore = record.Score
			if err = tx.Create(db_act_record).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}

	//redis
	err = self.GetDBrControl().HouseActivityDelete(dhid, actid)
	if err != nil {
		return err
	}
	err = self.GetDBrControl().HouseActivityRecordDelete(actid)
	if err != nil {
		return err
	}

	return nil
}

func (self *DBMgr) HouseMemberInsert(dhid int64, mem *HouseMember, tx *gorm.DB) (int64, error) {

	db_mem := new(models.HouseMember)
	db_mem.Id = mem.Id
	db_mem.DHId = dhid
	db_mem.UId = mem.UId
	db_mem.URole = mem.URole
	db_mem.URemark = mem.URemark
	db_mem.PRemark = mem.PRemark
	db_mem.NoFloors = mem.NoFloors
	db_mem.Ref = mem.Ref
	t := time.Now()
	db_mem.ApplyTime = &t
	if mem.URole == consts.ROLE_MEMBER || mem.URole == consts.ROLE_CREATER {
		db_mem.AgreeTime = db_mem.ApplyTime
	}
	db_mem.BwTimes = mem.BwTimes
	db_mem.PlayTimes = mem.PlayTimes
	db_mem.Forbid = 0
	db_mem.UVitamin = 0
	db_mem.Partner = 0

	//mysql
	if tx != nil {

		// 查询历史数据恢复
		var db_mem_log models.HouseMemberLog
		if err := tx.Model(db_mem_log).Where("dhid = ? AND uid = ? AND type = ? ", dhid, mem.UId, consts.OPTION_DELETE).Order("created_at desc").First(&db_mem_log).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return 0, err
			}
		}

		// db_mem.UVitamin = db_mem_log.UVitamin

		if err := tx.Create(db_mem).Error; err != nil {
			return 0, err
		}

		if db_mem.URole == consts.ROLE_MEMBER ||
			db_mem.URole == consts.ROLE_CREATER {
			// 日志
			db_mem_log := new(models.HouseMemberLog)
			db_mem_log.DHId = db_mem.DHId
			db_mem_log.UId = db_mem.UId
			db_mem_log.UVitamin = db_mem.UVitamin
			db_mem_log.Type = consts.OPTION_INSERT
			db_mem_log.URole = db_mem.URole
			db_mem_log.URemark = db_mem.URemark
			db_mem_log.BwTimes = db_mem.BwTimes
			db_mem_log.PlayTimes = db_mem.PlayTimes
			if db_mem.Ref > 0 {
				db_mem_log.Merge = true
			}
			if err := tx.Create(db_mem_log).Error; err != nil {
				return 0, err
			}
		}
	} else {

		// 查询历史数据恢复
		var db_mem_log models.HouseMemberLog
		if err := self.GetDBmControl().Model(db_mem_log).Where("dhid = ? AND uid = ? AND type = ? ", dhid, mem.UId, consts.OPTION_DELETE).Order("created_at desc").First(&db_mem_log).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return 0, err
			}
		}
		// db_mem.UVitamin = db_mem_log.UVitamin

		if err := self.GetDBmControl().Create(db_mem).Error; err != nil {
			return 0, err
		}

		if db_mem.URole == consts.ROLE_MEMBER ||
			db_mem.URole == consts.ROLE_CREATER {
			// 日志
			db_mem_log := new(models.HouseMemberLog)
			db_mem_log.DHId = mem.DHId
			db_mem_log.UId = mem.UId
			db_mem_log.UVitamin = db_mem.UVitamin
			db_mem_log.Type = consts.OPTION_INSERT
			db_mem_log.URole = db_mem.URole
			db_mem_log.URemark = db_mem.URemark
			db_mem_log.BwTimes = db_mem.BwTimes
			db_mem_log.PlayTimes = db_mem.PlayTimes
			if db_mem.Ref > 0 {
				db_mem_log.Merge = true
			}
			if err := self.GetDBmControl().Create(db_mem_log).Error; err != nil {
				return 0, err
			}
		}
	}

	mem.Id = db_mem.Id
	// mem.UVitamin = db_mem.UVitamin
	mem.UVitamin = 0

	//redis
	err := self.GetDBrControl().HouseMemberInsert(db_mem.ConvertModel())
	if err != nil {
		return 0, err
	}

	if mem.Upper(consts.ROLE_APLLY) {
		err := self.GetDBrControl().MemberHouseJoinInsert(db_mem.UId, db_mem.DHId)
		if err != nil {
			return 0, err
		}
	}

	return db_mem.Id, err
}

func (self *DBMgr) HouseMemberDelete(dhid int64, duid int64, uid int64, urole int, uvitamin int64, tx *gorm.DB) error {

	db_mem := new(models.HouseMember)
	db_mem.Id = duid
	db_mem.DHId = dhid
	db_mem.UId = uid
	db_mem.URole = urole
	db_mem.UVitamin = uvitamin

	var err error
	//mysql
	if tx != nil {
		if err = tx.Exec(`delete from house_member where hid = ? and uid = ?`, dhid, uid).Error; err != nil {
			return err
		}
		// 日志
		db_mem_log := new(models.HouseMemberLog)
		db_mem_log.DHId = db_mem.DHId
		db_mem_log.UId = db_mem.UId
		db_mem_log.UVitamin = db_mem.UVitamin
		db_mem_log.Type = consts.OPTION_DELETE
		db_mem_log.URole = db_mem.URole
		db_mem_log.URemark = db_mem.URemark
		db_mem_log.BwTimes = db_mem.BwTimes
		db_mem_log.PlayTimes = db_mem.PlayTimes
		if err = tx.Create(db_mem_log).Error; err != nil {
			return err
		}
	} else {

		if err = self.GetDBmControl().Exec(`delete from house_member where hid = ? and uid = ?`, dhid, uid).Error; err != nil {
			return err
		}
		// 日志
		db_mem_log := new(models.HouseMemberLog)
		db_mem_log.DHId = db_mem.DHId
		db_mem_log.UId = db_mem.UId
		db_mem_log.UVitamin = db_mem.UVitamin
		db_mem_log.Type = consts.OPTION_DELETE
		db_mem_log.URole = db_mem.URole
		db_mem_log.URemark = db_mem.URemark
		db_mem_log.BwTimes = db_mem.BwTimes
		db_mem_log.PlayTimes = db_mem.PlayTimes
		if err = self.GetDBmControl().Create(db_mem_log).Error; err != nil {
			return err
		}
	}

	//redis
	err = self.db_R.HouseMemberDelete(db_mem.ConvertModel())
	if err != nil {
		return err
	}
	err = self.db_R.MemberHouseJoinDelete(db_mem.UId, db_mem.DHId)
	if err != nil {
		return err
	}

	return nil
}

func (self *DBMgr) HouseMemberUpdate(dhid int64, mem *HouseMember) error {

	dmem, err := self.GetDBrControl().HouseMemberQueryById(dhid, mem.UId)
	if err != nil {
		return err
	}

	if dmem.URole == consts.ROLE_APLLY && mem.URole == consts.ROLE_MEMBER {
		dmem.AgreeTime = time.Now().Unix()
	}
	if mem.URole == consts.ROLE_BLACK {
		dmem.AgreeTime = time.Now().Unix()
	}

	dmem.URole = mem.URole
	dmem.URemark = mem.URemark
	dmem.DHId = mem.DHId
	dmem.FId = mem.FId
	dmem.Ref = mem.Ref

	//redis
	err = self.GetDBrControl().HouseMemberInsert(dmem)
	if err != nil {
		return err
	}

	if mem.Upper(consts.ROLE_APLLY) {
		err := self.GetDBrControl().MemberHouseJoinInsert(dmem.UId, dhid)
		if err != nil {
			return err
		}
	}

	if mem.Lower(consts.ROLE_MEMBER) {
		err := self.GetDBrControl().MemberHouseJoinDelete(dmem.UId, dhid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *DBMgr) ListHouseIdMemIn(uid int64) ([]int64, error) {

	clhIds, err := self.GetDBrControl().ListHouseMemberCreate(uid)
	if err != nil {
		return nil, err
	}
	jlhIds, err := self.GetDBrControl().ListHouseMemberJoin(uid)
	if err != nil {
		return nil, err
	}

	lhIds := append(clhIds, jlhIds...)
	return lhIds, nil
}

func (self *DBMgr) ListHouseMemIn(uid int64) ([]int64, error) {
	lkey := fmt.Sprintf(consts.REDIS_KEY_HOUSE_JOIN_MEMBER, uid)
	lval := self.Redis.LRange(lkey, 0, -1).Val()
	var datas []int64

	for _, val := range lval {
		id, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			self.Redis.LRem(lkey, 1, val) //非整形key，移除
			continue
		}
		datas = append(datas, id)
	}
	return datas, nil
}

func (self *DBMgr) ListHouseIdMemberJoin(uid int64) ([]*static.Msg_HouseItem, error) {

	lhIds, err := self.ListHouseIdMemIn(uid)
	if err != nil {
		return nil, err
	}

	var datas []*static.Msg_HouseItem

	for _, hid := range lhIds {

		// 包厢数据
		dhouse, err := self.GetDBrControl().GetHouseInfoById(hid)
		if err != nil {
			xlog.Logger().Errorln("user:", uid, " | house_", hid, " not exists")
			continue
		}

		house := GetClubMgr().GetClubHouseByHId(dhouse.HId)
		if house == nil {
			xlog.Logger().Errorln("user:", uid, " | house_", hid, " not exists")
			continue
		}

		// 玩家数据

		dmem, err := self.GetDBrControl().HouseMemberQueryById(dhouse.Id, dhouse.UId)
		if err != nil {
			continue
			// return nil, err
		}

		dperson := GetLazyUser(dhouse.UId)

		var data static.Msg_HouseItem
		data.Id = dhouse.Id
		data.HId = dhouse.HId
		data.HName = dhouse.Name
		data.HMems = house.TotalMemCount()
		data.OwnerId = dhouse.UId
		data.OwnerName = dperson.Name
		data.OwnerUrl = dperson.ImageUrl
		data.OwnerGender = dperson.Sex
		data.JoinTime = dmem.AgreeTime
		data.Role = dmem.URole
		if house.IsHideOnlineNum(uid) {
			data.OnlineCur = -1
			data.OnlineTotal = -1
			data.OnlineTable = -1
		} else {
			data.OnlineCur = house.OnlineMemCount()
			data.OnlineTotal = house.TotalMemCount()
			data.OnlineTable = house.GetTabOnlineCounts()
		}
		data.MergeHId = house.DBClub.MergeHId
		data.IsHidHide = house.DBClub.IsHidHide

		idArr, kindIdArr := house.GetFloors()
		data.FloorIDs = idArr
		data.FloorGameUrls = make([]string, 0)
		for _, kindid := range kindIdArr {
			areagame := GetAreaGameByKid(kindid)
			if areagame != nil {
				data.FloorGameUrls = append(data.FloorGameUrls, areagame.Icon)
			}
		}

		datas = append(datas, &data)
	}

	return datas, nil
}

// ! 获取包厢有效最低分数
func (self *DBMgr) SelectHouseValidRound(dhid int64) ([]models.HouseValidRound, error) {
	var housevalidrounds []models.HouseValidRound
	err := GetDBMgr().GetDBmControl().Model(models.HouseValidRound{}).Where("hid = ?", dhid).Find(&housevalidrounds).Error
	if err != nil {
		return []models.HouseValidRound{}, err
	}

	return housevalidrounds, err
}

// ! 获取包厢有效最低分数
func (self *DBMgr) SelectHouseFloorValidRound(dhid int64) ([]models.HouseFloorValidRound, error) {
	var housevalidrounds []models.HouseFloorValidRound
	err := GetDBMgr().GetDBmControl().Model(models.HouseFloorValidRound{}).Where("hid = ?", dhid).Find(&housevalidrounds).Error
	if err != nil {
		return []models.HouseFloorValidRound{}, err
	}

	return housevalidrounds, err
}

// ! 设置包厢有效对局分数
func (self *DBMgr) UpdataHouseValidRound(dhid int64, fid int64, minscore int, bigscore int) (int, error) {
	var housevalidround models.HouseFloorValidRound
	err := GetDBMgr().GetDBmControl().Model(models.HouseFloorValidRound{}).Where("hid = ? and fid = ?", dhid, fid).First(&housevalidround).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return 0, err
		}

		housevalidround.HId = dhid
		housevalidround.FId = fid
		housevalidround.ValidMinScore = minscore
		housevalidround.ValidBigScore = bigscore
		housevalidround.CreatedAt = time.Now()
		housevalidround.UpdatedAt = time.Now()
		if err = self.db_M.Create(&housevalidround).Error; err != nil {
			return minscore, err
		}
	}

	updateMap := make(map[string]interface{})
	updateMap["minscore"] = minscore
	updateMap["bigscore"] = bigscore
	if err = GetDBMgr().GetDBmControl().Model(models.HouseFloorValidRound{Id: housevalidround.Id}).Updates(updateMap).Error; err != nil {
		return minscore, err
	}
	return minscore, err
}

// ! 插入修改对局分数记录
func (self *DBMgr) InsertHouseValidRoundLog(dhid int64, fid int64, minscore int, bigscore int) error {
	var housevalidround models.HouseValidRoundLog
	housevalidround.HId = dhid
	housevalidround.FId = fid
	housevalidround.ValidMinScore = minscore
	housevalidround.ValidBigScore = bigscore
	housevalidround.CreatedAt = time.Now()
	housevalidround.UpdatedAt = time.Now()
	err := self.db_M.Create(&housevalidround).Error

	return err
}

// ! 查询成员统计数据
func (self *DBMgr) SelectHouseMemberStatistics(Hid int, DFid int, DayType int, SortType int) ([]*static.HouseMemberStatisticsItem, error) {
	// 取出相应时间的数据
	gameDayRecords, err := self.ListHouseMemberStatisticsForDayType(Hid, DFid, DayType)
	//计算求和统计
	statisticsItem := self.CalculateMemberStatistics(gameDayRecords)
	return statisticsItem, err
}

func (self *DBMgr) SelectHouseMemberStatisticsWithTotal(Hid int64, DFid int, selectTime1 time.Time, selectTime2 time.Time) (map[int64]models.QueryMemberStatisticsResult, error) {
	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())

	var result []models.QueryMemberStatisticsResult

	resultMap := make(map[int64]models.QueryMemberStatisticsResult)
	var err error
	if DFid == -1 {
		sql := `select uid, sum(win_score / radix) as totalscore, sum(1) as playtimes, sum(is_valid_round) as validtimes, sum(is_valid_round * is_big_winner) as bigwintimes from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 group by uid`
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2).Scan(&result).Error
	} else {
		sql := `select uid, sum(win_score / radix) as totalscore, sum(1) as playtimes, sum(is_valid_round) as validtimes, sum(is_valid_round * is_big_winner) as bigwintimes from record_game_total where hid = ? and created_at >= ? and created_at < ?  and score_kind <> 6 and dfid = ? group by uid`
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2, DFid).Scan(&result).Error
	}
	if err != nil {
		return resultMap, err
	}

	for _, item := range result {
		resultMap[item.Uid] = item
	}

	return resultMap, nil
}

func (self *DBMgr) SelectHouseMemberStatisticsWithCount(Hid int64, DFid int, partner int, selectTime1 time.Time, selectTime2 time.Time) (int, error) {
	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())
	count := 0
	var err error
	whereStr := "hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 "
	if partner == 0 {
		whereStr += "and partner = 0 "
	}
	if DFid == -1 {
		err = GetDBMgr().GetDBmControl().Table("record_game_total").Where(whereStr, Hid, selectDate1, selectDate2).Count(&count).Error
	} else {
		whereStr += "and dfid = ? "
		err = GetDBMgr().GetDBmControl().Table("record_game_total").Where(whereStr, Hid, selectDate1, selectDate2, DFid).Count(&count).Error
	}

	if err != nil {
		return 0, err
	}
	return count, nil
}

// ! 查询成员统计数据
func (self *DBMgr) SelectHouseMemberStatisticsWithPartner(Hid int, DFid int, DayType int, members map[int64]HouseMember) (map[int64]map[int64]*static.HouseMemberStatisticsItem, error) {

	// 取出相应时间的数据
	gameDayRecords, err := self.ListHouseMemberStatisticsForDayType(Hid, DFid, DayType)
	//计算求和统计
	statisticsItem := self.CalculateMemberStatisticsWithP(gameDayRecords, members)

	return statisticsItem, err
}

// ! 查询成员统计数据
func (self *DBMgr) SelectHouseMemberStatisticsWithPartnerUid(Hid int, DFid int, Uid int64, DayType int) ([]static.HousePartnerStatisticsItem, error) {

	var dest []static.HousePartnerStatisticsItem

	db := self.GetDBmControl().Table("house_day_record").Select("play_times as playtimes, bw_times as bwtimes, total_score / radix as totalscore, valid_times as validtimes, big_valid_times as bigvalidtimes, uid as uid").Where("hid = ?", Hid)
	if DFid != -1 {
		db = db.Where("dfid = ?", DFid)
	}
	if DayType == static.DAY_RECORD_TODAY {
		selectstr := fmt.Sprintf("%d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
		db = db.Where("date_format(created_at, '%Y-%m-%d') = ?", selectstr)
	} else if DayType == static.DAY_RECORD_YESTERDAY {
		selectTime := time.Now().AddDate(0, 0, -1)
		selectStr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())
		db = db.Where("date_format(created_at, '%Y-%m-%d') = ?", selectStr)
	} else if DayType == static.DAY_RECORD_3DAYS {
		endTime := time.Now().AddDate(0, 0, -1)
		endTimeStr := fmt.Sprintf("%d-%02d-%02d", endTime.Year(), endTime.Month(), endTime.Day())
		beginTime := time.Now().AddDate(0, 0, -3)
		beginTimeStr := fmt.Sprintf("%d-%02d-%02d", beginTime.Year(), beginTime.Month(), beginTime.Day())
		db = db.Where("date_format(created_at, '%Y-%m-%d') >= ?", beginTimeStr)
		db = db.Where("date_format(created_at, '%Y-%m-%d') <= ?", endTimeStr)
	} else if DayType == static.DAY_RECORD_7DAYS {
		endTime := time.Now().AddDate(0, 0, -1)
		endTimeStr := fmt.Sprintf("%d-%02d-%02d", endTime.Year(), endTime.Month(), endTime.Day())
		beginTime := time.Now().AddDate(0, 0, -7)
		beginTimeStr := fmt.Sprintf("%d-%02d-%02d", beginTime.Year(), beginTime.Month(), beginTime.Day())
		db = db.Where("date_format(created_at, '%Y-%m-%d') >= ?", beginTimeStr)
		db = db.Where("date_format(created_at, '%Y-%m-%d') <= ?", endTimeStr)
	}

	err := db.Where("partner = ?", Uid).Scan(&dest).Error

	return dest, err
}

func (self *DBMgr) ListHouseMemberStatisticsForDayType(Hid int, DFid int, DayType int) ([]models.RecordGameDay, error) {
	var gameDayRecord []models.RecordGameDay
	var err error
	if DayType == static.DAY_RECORD_TODAY {
		tmpgameDayRecord, error := self.SelectGameDetailDayRecordToDaySet(Hid, DFid, time.Now())
		if error == nil {
			gameDayRecord = append(gameDayRecord, tmpgameDayRecord...)
		}
	} else if DayType == static.DAY_RECORD_YESTERDAY {
		tmpgameDayRecord, error := self.SelectGameDetailDayRecordToDaySet(Hid, DFid, time.Now().AddDate(0, 0, -1))
		if error == nil {
			gameDayRecord = append(gameDayRecord, tmpgameDayRecord...)
		}
	} else if DayType == static.DAY_RECORD_3DAYS {
		tmpgameDayRecord, error := self.SelectGameDetailDayRecordToDaySetWithTime(Hid, DFid, time.Now().AddDate(0, 0, -3), time.Now().AddDate(0, 0, -1))
		if error == nil {
			gameDayRecord = append(gameDayRecord, tmpgameDayRecord...)
		}
	} else if DayType == static.DAY_RECORD_7DAYS {
		tmpgameDayRecord, error := self.SelectGameDetailDayRecordToDaySetWithTime(Hid, DFid, time.Now().AddDate(0, 0, -7), time.Now().AddDate(0, 0, -1))
		if error == nil {
			gameDayRecord = append(gameDayRecord, tmpgameDayRecord...)
		}
	} else if DayType < 0 {
		tmpgameDayRecord, error := self.SelectGameDetailDayRecordToDaySet(Hid, DFid, time.Now().AddDate(0, 0, DayType))
		if error == nil {
			gameDayRecord = append(gameDayRecord, tmpgameDayRecord...)
		}
	}

	return gameDayRecord, err
}

func (self *DBMgr) ListHouseFloorPartnerHistoryStatistics(Hid int, delFloorMap map[int64]models.HouseFloorDelMsg) []models.RecordGameDay {
	var gameDayRecord []models.RecordGameDay
	gameDayRecord, err := self.SelectGameDetailDayRecordToDaySetWithTime(Hid, -1, time.Now().AddDate(0, 0, -2), time.Now())
	if err != nil {
		return gameDayRecord
	}

	var deleteData []models.RecordGameDay
	for _, data := range gameDayRecord {
		for floorId, delFloor := range delFloorMap {
			if floorId == data.FId && data.CreatedAt.Day() == time.Unix(delFloor.CreateStamp, 0).Day() {
				deleteData = append(deleteData, data)
			}
		}
	}

	return deleteData
}

func (self *DBMgr) CalculateMemberStatistics(gameDayRecords []models.RecordGameDay) []*static.HouseMemberStatisticsItem {
	var memStatisticsMap = map[int64]*static.HouseMemberStatisticsItem{}
	for _, data := range gameDayRecords {
		memberStatisticsItem, ok := memStatisticsMap[int64(data.UId)]
		if ok {
			memberStatisticsItem.PlayTimes += data.PlayTimes
			memberStatisticsItem.BwTimes += data.BwTimes
			memberStatisticsItem.TotalScore += data.GetRealScore()
			memberStatisticsItem.ValidTimes += data.ValidTimes
			memberStatisticsItem.BigValidTimes += data.BigValidTimes
		} else {
			memberStatisticsItem := new(static.HouseMemberStatisticsItem)
			memberStatisticsItem.UId = data.UId
			memberStatisticsItem.PlayTimes = data.PlayTimes
			memberStatisticsItem.BwTimes = data.BwTimes
			memberStatisticsItem.TotalScore = data.GetRealScore()
			memberStatisticsItem.ValidTimes = data.ValidTimes
			memberStatisticsItem.BigValidTimes = data.BigValidTimes
			memberStatisticsItem.Partner = data.Partner

			memStatisticsMap[data.UId] = memberStatisticsItem
		}
	}

	var items []*static.HouseMemberStatisticsItem
	for _, data := range memStatisticsMap {
		data.TotalScore = static.HF_DecimalDivide(data.TotalScore, 1, 2)
		items = append(items, data)
	}
	return items
}

func (self *DBMgr) CalculateMemberStatisticsWithP(gameDayRecords []models.RecordGameDay, members map[int64]HouseMember) map[int64]map[int64]*static.HouseMemberStatisticsItem {
	memStatisticsMap := make(map[int64]map[int64]*static.HouseMemberStatisticsItem)
	for _, data := range gameDayRecords {
		mem, ok := members[data.UId]
		if !ok {
			continue
		}
		pUid := mem.UId
		if mem.Partner > 1 {
			pUid = mem.Partner
		} else if mem.Partner == 1 {
			pUid = mem.UId
		} else {
			continue
		}
		_, ok = memStatisticsMap[pUid]
		if !ok {
			memStatisticsMap[pUid] = make(map[int64]*static.HouseMemberStatisticsItem)
		}
		_, ok = memStatisticsMap[pUid][data.UId]
		if !ok {
			memberStatisticsItem := new(static.HouseMemberStatisticsItem)
			memberStatisticsItem.UId = data.UId
			memberStatisticsItem.PlayTimes = data.PlayTimes
			memberStatisticsItem.BwTimes = data.BwTimes
			memberStatisticsItem.TotalScore = data.GetRealScore()
			memberStatisticsItem.ValidTimes = data.ValidTimes
			memberStatisticsItem.BigValidTimes = data.BigValidTimes
			memberStatisticsItem.Partner = pUid

			memStatisticsMap[pUid][data.UId] = memberStatisticsItem
		} else {
			memStatisticsMap[pUid][data.UId].PlayTimes += data.PlayTimes
			memStatisticsMap[pUid][data.UId].BwTimes += data.BwTimes
			memStatisticsMap[pUid][data.UId].TotalScore += data.GetRealScore()
			memStatisticsMap[pUid][data.UId].ValidTimes += data.ValidTimes
			memStatisticsMap[pUid][data.UId].BigValidTimes += data.BigValidTimes
		}
	}

	return memStatisticsMap
}

func (self *DBMgr) CalculateMemberStatisticsWithPF(gameDayRecords []models.RecordGameDay, fids []int64) []*static.HouseMemberStatisticsItem {
	memStatisticsMap := make(map[int64]map[int64]map[int64]*static.HouseMemberStatisticsItem)
	for _, data := range gameDayRecords {
		haveDelete := true
		for _, fid := range fids {
			if data.FId == fid {
				haveDelete = false
				break
			}
		}
		if !haveDelete {
			continue
		}

		_, ok := memStatisticsMap[data.UId]
		if !ok {
			memStatisticsMap[data.UId] = make(map[int64]map[int64]*static.HouseMemberStatisticsItem)
		}
		_, ok = memStatisticsMap[data.UId][data.Partner]
		if !ok {
			memStatisticsMap[data.UId][data.Partner] = make(map[int64]*static.HouseMemberStatisticsItem)
		}
		memberStatisticsItem, ok := memStatisticsMap[data.UId][data.Partner][data.FId]
		if !ok {
			memberStatisticsItem := new(static.HouseMemberStatisticsItem)
			memberStatisticsItem.UId = data.UId
			memberStatisticsItem.PlayTimes = data.PlayTimes
			memberStatisticsItem.BwTimes = data.BwTimes
			memberStatisticsItem.TotalScore = data.GetRealScore()
			memberStatisticsItem.ValidTimes = data.ValidTimes
			memberStatisticsItem.BigValidTimes = data.BigValidTimes
			memberStatisticsItem.Partner = data.Partner

			memStatisticsMap[data.UId][data.Partner][data.FId] = memberStatisticsItem
		} else {
			memberStatisticsItem.PlayTimes += data.PlayTimes
			memberStatisticsItem.BwTimes += data.BwTimes
			memberStatisticsItem.TotalScore += data.GetRealScore()
			memberStatisticsItem.ValidTimes += data.ValidTimes
			memberStatisticsItem.BigValidTimes += data.BigValidTimes
		}
	}

	var items []*static.HouseMemberStatisticsItem
	for _, datass := range memStatisticsMap {
		for _, datas := range datass {
			for _, data := range datas {
				data.TotalScore = static.HF_DecimalDivide(data.TotalScore, 1, 2)
				items = append(items, data)
			}
		}
	}
	return items
}

func (self *DBMgr) SelectGameDayRecordToDaySet(hid int, selecttime time.Time) ([]models.RecordGameDay, error) {
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var dayrecords = []models.RecordGameDay{}
	err := self.GetDBmControl().Model(models.RecordGameDay{}).Where("hid = ? AND DATE_FORMAT(created_at, '%Y-%m-%d') = ?", hid, selectstr).Find(&dayrecords).Error
	return dayrecords, err
}

func (self *DBMgr) SelectUserGameDayRecordToDaySet(hid int, uid int64, selecttime time.Time) ([]models.RecordGameDay, error) {
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var dayrecords = []models.RecordGameDay{}
	err := self.GetDBmControl().Model(models.RecordGameDay{}).Where("hid = ? AND uid = ? AND DATE_FORMAT(created_at, '%Y-%m-%d') = ?", hid, uid, selectstr).Find(&dayrecords).Error
	return dayrecords, err
}

func (self *DBMgr) SelectGameDetailDayRecordToDaySet(hid int, dfid int, selecttime time.Time) ([]models.RecordGameDay, error) {
	//dfid 查询所有楼层
	if dfid == -1 {
		return self.SelectGameDayRecordToDaySet(hid, selecttime)
	}
	//指定楼层
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var dayrecords = []models.RecordGameDay{}
	err := self.GetDBmControl().Model(models.RecordGameDay{}).Where("hid = ? and dfid = ? and date_format(created_at, '%Y-%m-%d') = ?", hid, dfid, selectstr).Find(&dayrecords).Error
	return dayrecords, err
}

func (self *DBMgr) SelectGameDetailDayRecordToDaySetWithTime(hid int, dfid int, beginTime time.Time, endTime time.Time) ([]models.RecordGameDay, error) {
	//指定楼层
	beginstr := fmt.Sprintf("%d-%02d-%02d", beginTime.Year(), beginTime.Month(), beginTime.Day())
	endstr := fmt.Sprintf("%d-%02d-%02d", endTime.Year(), endTime.Month(), endTime.Day())
	var dayrecords []models.RecordGameDay
	var err error
	if dfid == -1 {
		err = self.GetDBmControl().Model(models.RecordGameDay{}).
			Where("hid = ? and date_format(created_at, '%Y-%m-%d') >= ? and date_format(created_at, '%Y-%m-%d') <= ?", hid, beginstr, endstr).
			Find(&dayrecords).Error
	} else {
		err = self.GetDBmControl().Model(models.RecordGameDay{}).
			Where("hid = ? and dfid = ? and date_format(created_at, '%Y-%m-%d') >= ? and date_format(created_at, '%Y-%m-%d') <= ?", hid, dfid, beginstr, endstr).
			Find(&dayrecords).Error
	}
	return dayrecords, err
}

func (self *DBMgr) SelectUserGameDetailDayRecordToDaySet(hid int, dfid int, uid int64, selecttime time.Time) ([]models.RecordGameDay, error) {
	//dfid 查询所有楼层
	if dfid == -1 {
		return self.SelectUserGameDayRecordToDaySet(hid, uid, selecttime)
	}
	//指定楼层
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var dayrecords = []models.RecordGameDay{}
	err := self.GetDBmControl().Model(models.RecordGameDay{}).Where("hid = ? and dfid = ? and uid = ? and date_format(created_at, '%Y-%m-%d') = ?", hid, dfid, uid, selectstr).Find(&dayrecords).Error
	return dayrecords, err
}

func (self *DBMgr) TransformTotalToHistory(gametotals []models.RecordGameTotal) []static.GameRecordHistory {
	gametotalmap := make(map[string]static.GameRecordHistory)

	for _, gametotal := range gametotals {
		gamerecord, ok := gametotalmap[gametotal.GameNum]
		if ok {
			person := new(static.GameRecordHistoryPlayer)
			person.Uid = gametotal.Uid
			person.Nickname = gametotal.UName
			person.Score = gametotal.GetRealScore()
			gamerecord.Player = append(gamerecord.Player, person)
		} else {
			gamerecord.KindId = gametotal.KindId
			gamerecord.GameNum = gametotal.GameNum
			gamerecord.RoomNum = gametotal.RoomNum
			gamerecord.PlayedAt = gametotal.CreatedAt.Unix()
			gamerecord.HId = gametotal.HId
			gamerecord.FId = gametotal.FId
			gamerecord.PlayCount = gametotal.PlayCount
			gamerecord.Round = gametotal.Round
			gamerecord.IsHeart = gametotal.IsHeart

			gamerecord.Player = make([]*static.GameRecordHistoryPlayer, 0)

			person := new(static.GameRecordHistoryPlayer)
			person.Uid = gametotal.Uid
			person.Nickname = gametotal.UName
			person.Score = gametotal.GetRealScore()
			gamerecord.Player = append(gamerecord.Player, person)
		}
		gametotalmap[gametotal.GameNum] = gamerecord
	}

	var gamerecords = []static.GameRecordHistory{}
	for _, gamerecord := range gametotalmap {
		gamerecords = append(gamerecords, gamerecord)
	}

	return gamerecords
}

func (self *DBMgr) TransformTotalToDetail(gameTotals []models.RecordGameTotal, bwScoreMap map[string]int, uid int64, puid int64, memMap map[int64]HouseMember, leaveMap map[int64]bool, likeFlagMap map[string]bool) ([]static.GameRecordDetal, int, int, int, int, int, float64, int, int, int, int) {
	gameTotalMap := make(map[string]static.GameRecordDetal)

	var curTotalRound = 0    //对局局数
	var curCompleteRound = 0 //完整局局数
	var curDismissRound = 0  //完整局局数
	var curValidRound = 0    //有效局局数
	var curInValidRound = 0  //低分局局数

	var curUserScore float64 = 0 //玩家总战绩
	var curBwTimes = 0           //大赢家人次
	var curPlayTimes = 0         //总人次
	var curInValidTimes = 0      //低分局人次
	var curTotalLikeRound = 0    //点赞局数

	for _, gameTotal := range gameTotals {
		gameRecord, ok := gameTotalMap[gameTotal.GameNum]
		if ok {
			person := new(static.GameRecordDetalPlayer)
			person.Uid = gameTotal.Uid
			person.NickName = gameTotal.UName
			person.Score = gameTotal.GetRealScore()
			gameRecord.Player = append(gameRecord.Player, *person)
			gameRecord.PartnerIds = append(gameRecord.PartnerIds, gameTotal.Partner)
			gameRecord.PlayerTags = append(gameRecord.PlayerTags, 0)
		} else {
			gameRecord.KindId = gameTotal.KindId
			gameRecord.GameNum = gameTotal.GameNum
			gameRecord.RoomNum = gameTotal.RoomNum
			gameRecord.PlayedAt = gameTotal.CreatedAt.Unix()
			gameRecord.HId = gameTotal.HId
			gameRecord.FId = gameTotal.FId
			gameRecord.DFId = gameTotal.DFId
			gameRecord.PlayRound = gameTotal.PlayCount
			gameRecord.TotalRound = gameTotal.Round
			gameRecord.IsHeart = gameTotal.IsHeart
			gameRecord.PartnerIds = append(gameRecord.PartnerIds, gameTotal.Partner)
			gameRecord.PlayerTags = append(gameRecord.PlayerTags, 0)
			gameRecord.Player = make([]static.GameRecordDetalPlayer, 0)
			if gameTotal.ScoreKind == static.ScoreKind_pass {
				if GetTableMgr().GetTable(gameRecord.RoomNum) == nil {
					gameRecord.FinishType = static.FINISH_STA_1ST_DISMISS
				} else {
					gameRecord.FinishType = static.FINISH_STA_PLAYING
				}
			} else if gameTotal.Round > gameTotal.PlayCount {
				gameRecord.FinishType = static.FINISH_STA_HALF_DISMISS
				// 中途解散局数
				curDismissRound++
			} else {
				if gameTotal.HalfWayDismiss {
					gameRecord.FinishType = static.FINISH_STA_HALF_DISMISS
					// 中途解散局数
					curDismissRound++
				} else {
					gameRecord.FinishType = static.FINISH_STA_NORMAL
					// 统计完整局数
					curCompleteRound++
				}
			}
			// 统计总局数
			curTotalRound++

			// 统计总的有效局
			if gameRecord.FinishType != static.FINISH_STA_PLAYING {
				if gameTotal.IsValidRound {
					curValidRound++
				} else {
					curInValidRound++
				}
			}

			person := new(static.GameRecordDetalPlayer)
			person.Uid = gameTotal.Uid
			person.NickName = gameTotal.UName
			person.Score = gameTotal.GetRealScore()
			gameRecord.Player = append(gameRecord.Player, *person)
		}
		gameTotalMap[gameTotal.GameNum] = gameRecord

		// 查询的是玩家数据,需要统计玩家积分
		if gameTotal.Uid == uid {
			curUserScore += gameTotal.GetRealScore()

			if gameTotal.WinScore >= bwScoreMap[gameTotal.GameNum] && bwScoreMap[gameTotal.GameNum] > 0 && gameTotal.IsValidRound {
				curBwTimes++
			}
		}

		if puid > 0 {
			playerMem, ok := memMap[gameTotal.Uid]
			if ok {
				if playerMem.UId == puid || playerMem.Partner == puid {
					curPlayTimes += 1

					if !gameTotal.IsValidRound && gameRecord.FinishType != static.FINISH_STA_PLAYING {
						curInValidTimes++
					}
				}
			} else {
				if _, ok := leaveMap[gameTotal.Uid]; ok {
					curPlayTimes += 1

					if !gameTotal.IsValidRound && gameRecord.FinishType != static.FINISH_STA_PLAYING {
						curInValidTimes++
					}
				}
			}
		} else {
			curPlayTimes += 1

			if !gameTotal.IsValidRound && gameRecord.FinishType != static.FINISH_STA_PLAYING {
				curInValidTimes++
			}
		}
	}

	var gameRecords []static.GameRecordDetal
	for _, gameRecord := range gameTotalMap {
		if _, ok := likeFlagMap[gameRecord.GameNum]; ok {
			curTotalLikeRound++
			gameRecord.IsHeart = 1
		} else {
			gameRecord.IsHeart = 0
		}
		gameRecords = append(gameRecords, gameRecord)
	}

	return gameRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound
}

// ! 查询大厅我的战绩
func (self *DBMgr) SelectNormalHallRecord(uid int64, selecttime time.Time, kindidRange string) ([]static.GameRecordHistory, error) {
	var gametotals = []models.RecordGameTotal{}
	sqltimestr := selecttime.Format(consts.TIME_Y_M_D)

	var err error
	if kindidRange == "" {
		err = self.GetDBmControl().Model(models.RecordGameTotal{}).Where("game_num in (select game_num from record_game_total where uid = ? and hid = 0 and score_kind <> 6 and date_format(created_at, '%Y-%m-%d') = ?) ", uid, sqltimestr).Order("created_at DESC").Find(&gametotals).Error
	} else {
		sql := "game_num in (select game_num from record_game_total where uid = ? and hid = 0 and score_kind <> 6 and date_format(created_at, '%Y-%m-%d') = ? and kind_id in (" + kindidRange + ") )"
		err = self.GetDBmControl().Model(models.RecordGameTotal{}).Where(sql, uid, sqltimestr).Order("created_at DESC").Find(&gametotals).Error
	}

	if err != nil {
		return []static.GameRecordHistory{}, err
	}

	gamerecords := self.TransformTotalToHistory(gametotals)

	return gamerecords, nil
}

// ! 查询包厢我的战绩(近一天的数据)
func (self *DBMgr) SelectHouseHallMyRecord(hid int64, uid int64) ([]static.GameRecordHistory, error) {
	gamerecords, err := self.SelectHouseHallRecord(hid)
	if err != nil {
		return []static.GameRecordHistory{}, err
	}

	for i := 0; i < len(gamerecords); i++ {
		ingame := false
		for _, player := range gamerecords[i].Player {
			if player.Uid == uid {
				ingame = true
				break
			}
		}
		if !ingame {
			gamerecords = append(gamerecords[:i], gamerecords[i+1:]...)
			i--
		}
	}

	return gamerecords, nil
}

// ! 查询包厢战绩(近一天的数据)
func (self *DBMgr) SelectHouseHallRecord(hid int64) ([]static.GameRecordHistory, error) {
	var gametotals = []models.RecordGameTotal{}
	err := self.GetDBmControl().Model(models.RecordGameTotal{}).Where("hid = ? AND DATE_SUB(CURDATE(), INTERVAL 1 DAY) <= created_at AND score_kind <> 6", hid).Order("created_at DESC").Find(&gametotals).Error

	if err != nil {
		return []static.GameRecordHistory{}, err
	}

	gamerecords := self.TransformTotalToHistory(gametotals)

	return gamerecords, nil
}

// ! 按照日期查询包厢战绩
func (self *DBMgr) SelectHouseGameRecordByDate(hid int64, floorIndex int, uid int64, bw bool, puid int64, memMap map[int64]HouseMember, selectTime1 time.Time, selectTime2 time.Time, searchkey string, likeFlagMap map[string]bool, likeFlag int, roundType int) ([]static.GameRecordDetal, int, int, int, int, int, float64, int, int, int, int, error) {
	// 查询包厢所在区域是否显示正在游戏的战绩记录
	bIsShowPlayingRecord := false
	/*
		house := GetClubMgr().GetClubHouseById(hid)
		if house != nil {
			var areaInfo model.Area
			db := self.GetDBmControl().Model(model.Area{})
			db = db.Where("area_id = ?", house.DBClub.Area)
			err := db.Find(&areaInfo).Error
			if err == nil {
				bIsShowPlayingRecord = areaInfo.IsShowPlayingRecord == 1
			}
		}
	*/

	var gameTotals []models.RecordGameTotal

	db := self.GetDBmControl()

	db = self.GetDBmControl().Model(models.RecordGameTotal{})

	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())

	// 牌局类型 牌局类型 0 全部 1 完整局数 2 中途解散 3 低分局
	if roundType == 0 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ?", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and score_kind != 6", hid, selectDate1, selectDate2)
		}
	} else if roundType == 1 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and play_count = round and halfwaydismiss = 0", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and play_count = round and halfwaydismiss = 0 and score_kind != 6", hid, selectDate1, selectDate2)
		}
	} else if roundType == 2 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and (play_count < round or halfwaydismiss = 1) ", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and (play_count < round or halfwaydismiss = 1) and score_kind != 6", hid, selectDate1, selectDate2)
		}
	} else if roundType == 3 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and is_valid_round = 0 ", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and is_valid_round = 0 and score_kind != 6", hid, selectDate1, selectDate2)
		}
	}

	var err error
	err = db.Find(&gameTotals).Error

	if err != nil {
		return []static.GameRecordDetal{}, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	// 每局的大赢家分数
	bwUserScoreMap := make(map[string]int)
	// 战绩记录
	gameTotalMap := make(map[string][]models.RecordGameTotal)

	for i := 0; i < len(gameTotals); i++ {
		gameRecord := gameTotals[i]

		// 成员战绩--详情--大赢家详情
		if bw && !gameRecord.IsValidRound {
			continue
		}

		// 筛选点赞信息
		if len(likeFlagMap) > 0 {
			if likeFlag == 1 {
				if !likeFlagMap[gameRecord.GameNum] {
					continue
				}
			} else if likeFlag == 2 {
				if likeFlagMap[gameRecord.GameNum] {
					continue
				}
			}
		} else {
			if likeFlag == 1 {
				continue
			}
		}

		// 房间是否已解散
		if gameRecord.ScoreKind == static.ScoreKind_pass && gameRecord.PlayCount == 1 && GetTableMgr().GetTable(gameRecord.RoomNum) == nil {
			continue
		}

		// 是否已保存在战绩map中
		_, ok := gameTotalMap[gameRecord.GameNum]
		if !ok {
			gameTotalMap[gameRecord.GameNum] = []models.RecordGameTotal{}
			bwUserScoreMap[gameRecord.GameNum] = gameRecord.WinScore
		}

		// 更新战绩map
		gameTotalMap[gameRecord.GameNum] = append(gameTotalMap[gameRecord.GameNum], gameRecord)
		// 更新大赢家分数
		if gameRecord.WinScore > bwUserScoreMap[gameRecord.GameNum] {
			bwUserScoreMap[gameRecord.GameNum] = gameRecord.WinScore
		}
	}

	// 已退圈的成员记录
	leaveMemMap := make(map[int64]bool)
	if uid == 0 && puid > 0 {
		leaveMemMap, _ = GetDBMgr().SelectLeaveHousePartnerMember(hid, puid, selectDate1)
	}

	// 筛选记录
	CheckRecordCondition := func(reqUid int64, floorIndex int, gameRecords []models.RecordGameTotal, mpuid int64, mMap map[int64]HouseMember) bool {
		for _, gameRecord := range gameRecords {
			if uid > 0 {
				if gameRecord.Uid == uid {
					if floorIndex != -1 {
						if gameRecord.DFId == floorIndex {
							return true
						}
					} else {
						return true
					}
				}
			} else {
				if mpuid > 0 {
					mem, ok := mMap[gameRecord.Uid]
					if ok {
						if mem.Partner == mpuid || mem.UId == mpuid {
							if floorIndex != -1 {
								if gameRecord.DFId == floorIndex {
									return true
								}
							} else {
								return true
							}
						}
					} else {
						if _, ok := leaveMemMap[gameRecord.Uid]; ok {
							return true
						}
					}
				} else {
					if floorIndex != -1 {
						if gameRecord.DFId == floorIndex {
							return true
						}
					} else {
						return true
					}
				}
			}
		}
		return false
	}

	// 筛选后的战绩map
	searchGameMap := make(map[string][]models.RecordGameTotal)
	// 遍历我的包厢以及楼层战绩信息
	for gameNum, gameRecords := range gameTotalMap {
		if CheckRecordCondition(uid, floorIndex, gameRecords, puid, memMap) {
			searchGameMap[gameNum] = gameRecords
		}
	}

	// 查询汇总的所有记录
	var searchGameTotals []models.RecordGameTotal

	// 遍历搜索
	if searchkey != "" {
		for _, searchRecords := range searchGameMap {
			for _, gameRecord := range searchRecords {
				if static.HF_Itoa(gameRecord.RoomNum) == searchkey || static.HF_I64toa(gameRecord.Uid) == searchkey {
					if bw {
						if static.HF_I64toa(gameRecord.Uid) == searchkey && gameRecord.WinScore >= bwUserScoreMap[gameRecord.GameNum] {
							searchGameTotals = append(searchGameTotals, searchRecords...)
						}
					} else {
						searchGameTotals = append(searchGameTotals, searchRecords...)
					}
					break
				}
			}
		}
	} else {
		for _, searchRecords := range searchGameMap {
			for _, gameRecord := range searchRecords {
				if bw {
					if gameRecord.Uid == uid && gameRecord.WinScore >= bwUserScoreMap[gameRecord.GameNum] && bwUserScoreMap[gameRecord.GameNum] > 0 {
						searchGameTotals = append(searchGameTotals, searchRecords...)
						break
					}
				} else {
					searchGameTotals = append(searchGameTotals, searchRecords...)
					break
				}
			}
		}
	}

	gameRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound := self.TransformTotalToDetail(searchGameTotals, bwUserScoreMap, uid, puid, memMap, leaveMemMap, likeFlagMap)

	return gameRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound, nil
}

// ! 按照时间查询包厢总局数,有效局数量
func (self *DBMgr) SelectHouseGameRound(hid int64, floorIndex int, searchuids []int64, selectTime1 time.Time, selectTime2 time.Time) (int, int, error) {
	type QueryResult struct {
		PlayRound  int `gorm:"column:playround"`  //! 当天总局数
		ValidRound int `gorm:"column:validround"` //! 当天累计有效局数
	}

	// 格式化筛选时间
	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())

	var queryResult QueryResult
	var sql string
	if floorIndex == -1 {
		if len(searchuids) > 0 {
			sql := "select count(1) as playround, sum(is_valid_round) as validround from record_game_total where seat_id = 0 and game_num in (select game_num  from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 and uid in (?) group by game_num)"
			if err := GetDBMgr().GetDBmControl().Raw(sql, hid, selectDate1, selectDate2, searchuids).Scan(&queryResult).Error; err != nil {
				return 0, 0, err
			}
		} else {
			sql = "select count(1) as playround, sum(is_valid_round) as validround from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 and seat_id = 0"
			if err := GetDBMgr().GetDBmControl().Raw(sql, hid, selectDate1, selectDate2).Scan(&queryResult).Error; err != nil {
				return 0, 0, err
			}
		}
	} else {
		if len(searchuids) > 0 {
			sql := "select count(1) as playround, sum(is_valid_round) as validround from record_game_total where seat_id = 0 and game_num in (select game_num  from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 and dfid = ? and uid in (?) group by game_num)"
			if err := GetDBMgr().GetDBmControl().Raw(sql, hid, selectDate1, selectDate2, floorIndex, searchuids).Scan(&queryResult).Error; err != nil {
				return 0, 0, err
			}
		} else {
			sql := "select count(1) as playround, sum(is_valid_round) as validround from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 and dfid = ? and seat_id = 0"
			if err := GetDBMgr().GetDBmControl().Raw(sql, hid, selectDate1, selectDate2, floorIndex).Scan(&queryResult).Error; err != nil {
				return 0, 0, err
			}
		}
	}

	return queryResult.PlayRound, queryResult.ValidRound, nil
}

// ! 按照日期查询包厢无效局数量
func (self *DBMgr) SelectInvalidRoundByDate(hid int64, selecttime time.Time) (int, error) {
	type queryResult struct {
		PlayTimes  int `gorm:"column:play_times"`  //! 当天总局数
		ValidTimes int `gorm:"column:valid_times"` //! 当天累计有效局数
	}

	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	res := new(queryResult)
	err := GetDBMgr().db_M.Model(models.RecordGameDay{}).
		Select("sum(play_times) as play_times, sum(valid_times) as valid_times").
		Where(" hid = ? and date_format(created_at,'%Y-%m-%d') = ?", hid, selectstr).
		Scan(&res).Error
	if err != nil {
		return 0, err
	}
	return res.PlayTimes - res.ValidTimes, err
}

// ! 按照日期查询包厢每层的开局数量
func (self *DBMgr) SelectHouseRoundByDate(hid int64, day int) (map[string]*static.HouseOperationalStatusItem, error) {
	houseOsMap := make(map[string]*static.HouseOperationalStatusItem)
	timeNow := time.Now()

	lastDateStr := ""
	for i := 0; i < day; i++ {
		selectTime := timeNow.AddDate(0, 0, -i)
		selectStr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())
		houseOsMap[selectStr] = new(static.HouseOperationalStatusItem)
		houseOsMap[selectStr].QueryTime = selectTime.Unix()

		if i == day-1 {
			lastDateStr = selectStr
		}
	}

	type HouseFloorRound struct {
		Round    int    `gorm:"column:round"`
		DestFid  int    `gorm:"column:destfid"`
		DestDate string `gorm:"column:destdate"`
	}

	var destRound []HouseFloorRound
	sql := `select cast(created_at as date) as destdate, dfid as destfid, count(1) as round from record_game_total where score_kind <> 6 and seat_id = 0 and hid = ? and date_format(created_at, "%Y-%m-%d") >= ? group by cast(created_at as date), dfid`
	err := self.GetDBmControl().Raw(sql, hid, lastDateStr).Scan(&destRound).Error

	if err != nil {
		return houseOsMap, err
	}

	for _, item := range destRound {
		dateStr := item.DestDate[0:10]
		if _, ok := houseOsMap[dateStr]; ok {
			if item.DestFid < static.MaxFloorIndex {
				houseOsMap[dateStr].PlayRounds[item.DestFid] = item.Round
				houseOsMap[dateStr].TotalRounds += item.Round
			}
		}
	}

	type HouseDateKaCost struct {
		DestCost int    `gorm:"column:destcost"`
		DestDate string `gorm:"column:destdate"`
	}
	var destCost []HouseDateKaCost

	sql = `select cast(created_at as date) as destdate, sum(kacost) as destcost from record_game_cost where hid = ? and date_format(created_at, "%Y-%m-%d") >= ? group by cast(created_at as date)`
	err = self.GetDBmControl().Raw(sql, hid, lastDateStr).Scan(&destCost).Error
	if err != nil {
		return houseOsMap, err
	}

	for _, item := range destCost {
		dateStr := item.DestDate[0:10]
		if _, ok := houseOsMap[dateStr]; ok {
			houseOsMap[dateStr].TotalFangkaCost = -item.DestCost
		}
	}

	return houseOsMap, err
}

// ! 查询战绩详情
func (self *DBMgr) SelectGameRecordInfo(gamenum string) (static.Msg_S2C_GameRecordInfo, error) {
	var gametotals = []models.RecordGameTotal{}
	err := self.GetDBmControl().Model(models.RecordGameTotal{}).Where("game_num = ?", gamenum).Order("seat_id ASC").Find(&gametotals).Error
	if err != nil || len(gametotals) == 0 {
		return static.Msg_S2C_GameRecordInfo{}, err
	}

	playercount := len(gametotals)
	playcount := gametotals[0].PlayCount

	var gamerounds = []models.RecordGameRound{}
	err = self.GetDBmControl().Model(models.RecordGameRound{}).Where("gamenum = ?", gamenum).Order("created_at ASC").Find(&gamerounds).Error
	if err != nil {
		return static.Msg_S2C_GameRecordInfo{}, err
	}

	var gamecost = []models.RecordGameCost{}
	err = self.GetDBmControl().Model(models.RecordGameCost{}).Where("game_num = ?", gamenum).Order("created_at ASC").Find(&gamecost).Error
	if err != nil || len(gamecost) == 0 {
		return static.Msg_S2C_GameRecordInfo{}, err
	}
	gameConfig := gamecost[0].GameConfig
	var gameConfigMap map[string]interface{}
	gameConfigMap = make(map[string]interface{})
	json.Unmarshal([]byte(gameConfig), &gameConfigMap)
	//syslog.Logger().Println("gamecost:", gameConfigMap["dff"], gameConfigMap["difen"], gameConfigMap["scoreradix"])
	difen := 0
	if gameConfigMap["difen"] != nil {
		difen = int(gameConfigMap["difen"].(float64))
	} else if gameConfigMap["dff"] != nil {
		difen = int(gameConfigMap["dff"].(float64))
	} else if gameConfigMap["basescore"] != nil {
		difen = int(gameConfigMap["basescore"].(float64))
	}
	scoreradix := 1
	if gameConfigMap["scoreradix"] != nil {
		scoreradix = int(gameConfigMap["scoreradix"].(float64))
	}

	roundRecord := new(static.Msg_S2C_GameRecordInfo)
	roundRecord.GameNum = gamenum
	roundRecord.KindId = gametotals[0].KindId
	roundRecord.RoomId = gametotals[0].RoomNum
	roundRecord.Time = time.Now().Unix()
	roundRecord.TotalRound = gametotals[0].Round
	roundRecord.FloorIndex = gametotals[0].DFId
	roundRecord.DiFen = difen / scoreradix
	// 玩家列表
	var house *Club
	if gametotals[0].HId > 0 {
		house = GetClubMgr().GetClubHouseById(gametotals[0].HId)
	}
	for i := 0; i < playercount; i++ {
		capNickname := ""
		var capId int64 = -1
		if gametotals[0].HId > 0 {
			mem := house.GetMemByUId(gametotals[i].Uid)
			if mem != nil {
				capId = mem.Partner
				if mem.Partner > 1 {
					capNickname = house.GetMemByUId(mem.Partner).NickName
				}
			}
		}
		roundRecord.UserArr = append(roundRecord.UserArr, &static.Msg_S2C_GameRecordInfoUser{
			Uid:         gametotals[i].Uid,
			Nickname:    gametotals[i].UName,
			Imgurl:      "",
			Sex:         0,
			Score:       gametotals[i].GetRealScore(),
			CapId:       capId,
			CapNickname: capNickname,
		})
	}

	// 每局积分
	scoreArr := make([][]float64, 0)
	uidArr := make([][]int64, 0)

	// 初始化二维数组
	for i := 0; i < playcount; i++ {
		arr := make([]float64, playercount)
		scoreArr = append(scoreArr, arr)

		arr2 := make([]int64, playercount)
		uidArr = append(uidArr, arr2)
	}

	replayIdArr := make([]int64, 0)
	endTimeArr := make([]int64, 0)
	startTimeArr := make([]int64, 0)
	for i := 0; i < playercount; i++ {
		list := self.SelectUserGameRecordInfo(i, gamerounds)
		for j, item := range list {
			// 测试时屏蔽, 为避免异常情况, 上线时应打开
			if j >= playcount {
				continue
			}
			scoreArr[j][i] = item.GetRealScore()
			if i == 0 {
				replayIdArr = append(replayIdArr, item.ReplayId)
				if j == len(list)-1 {
					endTimeArr = append(endTimeArr, gametotals[0].CreatedAt.Unix())
				} else {
					endTimeArr = append(endTimeArr, item.CreatedAt.Unix())
				}
				startTimeArr = append(startTimeArr, item.BeginDate.Unix())
			}
			uidArr[j][i] = item.UId
		}
	}

	for i := 0; i < playcount; i++ {
		if i >= len(replayIdArr) {
			break
		}
		roundRecord.ScoreArr = append(roundRecord.ScoreArr, &static.Msg_S2C_GameRecordInfoScore{
			ReplayId:  replayIdArr[i],
			StartTime: startTimeArr[i],
			EndTime:   endTimeArr[i],
			Score:     scoreArr[i],
			Uids:      uidArr[i],
		})
	}

	return *roundRecord, nil
}

func (self *DBMgr) SelectUserGameRecordInfo(seatid int, gamerounds []models.RecordGameRound) []models.RecordGameRound {
	usergamerounds := []models.RecordGameRound{}
	for _, gameround := range gamerounds {
		if gameround.SeatId == seatid {
			usergamerounds = append(usergamerounds, gameround)
		}
	}
	return usergamerounds
}

// ! 更新战绩点赞数据
func (self *DBMgr) UpdateGameRecordHeart(Hid int64, GameNum string, SeatId int, IsHeart int) error {
	//mysql
	err := self.GetDBmControl().Model(models.RecordGameTotal{}).Where("game_num = ?", GameNum).Update("is_heart", IsHeart).Error

	return err
}

// ! 更新大赢家对局统计
func (self *DBMgr) UpdateHouseRecordBwTimes(DHId int64, UId int64, Times int) error {
	db_houseMember, err := GetDBMgr().GetDBrControl().HouseMemberQueryById(DHId, UId)
	if err != nil {
		return err
	}
	db_houseMember.BwTimes = Times

	//mysql
	if err = self.GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", db_houseMember.DHId, db_houseMember.UId).Update("bw_times", Times).Error; err != nil {
		return err
	}

	//redis
	if err = GetDBMgr().GetDBrControl().HouseMemberInsert(db_houseMember); err != nil {
		return err
	}
	return err
}

// ! 更新对局对局统计
func (self *DBMgr) UpdateHouseRecordPlayTimes(DHId int64, UId int64, Times int) error {
	db_houseMember, err := GetDBMgr().GetDBrControl().HouseMemberQueryById(DHId, UId)
	if err != nil {
		return err
	}
	db_houseMember.PlayTimes = Times

	//mysql
	if err = self.GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", db_houseMember.DHId, db_houseMember.UId).Update("play_times", Times).Error; err != nil {
		return err
	}

	//redis
	if err = GetDBMgr().GetDBrControl().HouseMemberInsert(db_houseMember); err != nil {
		return err
	}

	return err
}

// ! 查询大赢家统计
func (self *DBMgr) HouseRecordBWList(dhid int64) ([]models.HouseMember, error) {
	//mysql
	var housemember = []models.HouseMember{}
	err := self.GetDBmControl().Model(models.HouseMember{}).Where("hid = ?", dhid).Order("bw_times DESC").Find(&housemember).Error
	return housemember, err
}

// ! 查询对局统计
func (self *DBMgr) HouseRecordPlayList(dhid int64) ([]models.HouseMember, error) {
	//mysql
	var housemember = []models.HouseMember{}
	err := self.GetDBmControl().Model(models.HouseMember{}).Where("hid = ?", dhid).Order("play_times DESC").Find(&housemember).Error
	return housemember, err
}

// ! 查询扣卡统计
func (self *DBMgr) HouseRecordCostList(dhid int64) ([]models.RecordGameCostMini, error) {
	//mysql
	var gamecosts = []models.RecordGameCost{}
	err := self.GetDBmControl().Model(models.RecordGameCost{}).Where("hid = ? AND DATE_SUB(CURDATE(), INTERVAL 7 DAY) <= created_at", dhid).Find(&gamecosts).Error
	if err != nil {
		return []models.RecordGameCostMini{}, err
	}

	gamecostminimap := make(map[string]models.RecordGameCostMini)

	gamenummap := make(map[string]map[string]struct{})
	for _, gamecost := range gamecosts {
		timeKey := fmt.Sprintf("%d%-02d-%02d", gamecost.CreatedAt.Year(), gamecost.CreatedAt.Month(), gamecost.CreatedAt.Day())
		// 先按gamenum统计局数
		gamenumm, ok := gamenummap[timeKey]
		if !ok {
			gamenumm = make(map[string]struct{})
		}
		gamenumm[gamecost.Gamenum] = struct{}{}
		gamenummap[timeKey] = gamenumm
		// 再统计房卡消耗
		gamecostmini, ok := gamecostminimap[timeKey]
		if !ok {
			gamecostmini.HId = gamecost.HId
			gamecostmini.Date = gamecost.CreatedAt
			gamecostmini.PlayTime = 0
			gamecostmini.KaCost = 0
		}
		if gamecost.LeagueID > 0 {
			gamecostmini.KaCost -= 0
		} else {
			gamecostmini.KaCost -= gamecost.KaCost
		}
		gamecostminimap[timeKey] = gamecostmini
	}

	var keys []string
	for k, v := range gamecostminimap {
		v.PlayTime = len(gamenummap[k])
		gamecostminimap[k] = v
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var gamecostminis = []models.RecordGameCostMini{}
	for i := len(keys) - 1; i >= 0; i-- {
		gamecostminis = append(gamecostminis, gamecostminimap[keys[i]])
	}
	return gamecostminis, err
}

func (self *DBMgr) GetHouseRecordCardCostAndPlayTimes(dhid int64, gameNums []string, ids ...int64) (map[int64]models.RecordGameCostMini, error) {
	db := self.GetDBmControl().Model(models.RecordGameCost{})

	db = db.Where("hid = ?", dhid)

	db = db.Where("game_num in(?)", gameNums)

	if len(ids) > 0 {
		db = db.Where("uid in(?)", ids)
	}

	//mysql
	gameCosts := make([]*models.RecordGameCost, 0)

	err := db.Find(&gameCosts).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return nil, err
	}
	resultGameNum := make(map[int64]map[string]struct{})
	result := make(map[int64]models.RecordGameCostMini)
	for _, costMini := range gameCosts {
		gameNumMap, ok := resultGameNum[costMini.UId]
		if !ok {
			gameNumMap = make(map[string]struct{})
		}
		gameNumMap[costMini.Gamenum] = struct{}{}
		resultGameNum[costMini.UId] = gameNumMap

		res, ok1 := result[costMini.UId]
		if !ok1 {
			res = models.RecordGameCostMini{
				HId:    costMini.HId,
				KaCost: 0,
				Date:   costMini.CreatedAt,
			}
		}
		res.KaCost -= costMini.KaCost
		result[costMini.UId] = res
	}
	for uid, cost := range result {
		cost.PlayTime = len(resultGameNum[uid])
		result[uid] = cost
	}
	return result, nil
}

// ! 插入用户任务
func (self *DBMgr) InsertUserTask(task *static.Task) error {
	// mysql
	mtask := new(models.UserTask)
	mtask.TaskId = task.TcId
	mtask.Uid = task.Uid
	mtask.Num = task.Num
	mtask.Step = task.Step
	mtask.Sta = task.Sta
	mtask.Time = time.Unix(task.Time, 0)
	if err := self.db_M.Create(mtask).Error; err != nil {
		return err
	}
	task.Id = mtask.Id

	if err := self.db_R.UserTaskInsert(task); err != nil {
		return err
	}

	return nil
}

// ! 更新用户任务
func (self *DBMgr) UpdateUserTask(task *static.Task) error {
	// mysql
	updateMap := make(map[string]interface{})
	updateMap["num"] = task.Num
	updateMap["sta"] = task.Sta
	updateMap["step"] = task.Step
	if err := self.db_M.Model(models.UserTask{Id: task.Id}).Updates(updateMap).Error; err != nil {
		return err
	}

	// redis
	if err := self.db_R.UserTaskUpdate(task); err != nil {
		return err
	}

	return nil
}

// ! 删除用户任务
func (self *DBMgr) DeleteUserTask(task *static.Task) error {
	// mysql
	mtask := new(models.UserTask)
	mtask.Id = task.Id
	if err := self.db_M.Delete(mtask).Error; err != nil {
		return err
	}

	if err := self.db_R.UserTaskDelete(task); err != nil {
		return err
	}

	return nil
}

// ! 查询队长疲劳值统计
func (self *DBMgr) SelectPartnerVitaminStatistic(dhid int64, dfid int64, selectstarttime time.Time, selectendtime time.Time, aftertime int64) ([]models.RecordVitaminDay, error) {
	selectstartstr := fmt.Sprintf("%d-%02d-%02d", selectstarttime.Year(), selectstarttime.Month(), selectstarttime.Day())
	selectendstr := fmt.Sprintf("%d-%02d-%02d", selectendtime.Year(), selectendtime.Month(), selectendtime.Day())
	var vitaminstatistics []models.RecordVitaminDay

	db := self.GetDBmControl().Model(models.RecordVitaminDay{})

	db = db.Where("hid = ?", dhid)

	db = db.Where("date_format(created_at, '%Y-%m-%d') >= ? ", selectstartstr)

	db = db.Where("date_format(created_at, '%Y-%m-%d') <= ? ", selectendstr)

	if aftertime != 0 {
		db = db.Where("created_at > ? ", time.Unix(aftertime, 0))
	}
	if dfid != -1 {
		db = db.Where("dfid = ?", dfid)
	}

	err := db.Find(&vitaminstatistics).Error

	if err != nil {
		return []models.RecordVitaminDay{}, err
	}

	return vitaminstatistics, err
}

// ! 查询队长疲劳值统计
func (self *DBMgr) SelectPartnerVitaminStatisticMgr(dhid int64, dfid int64, selectstarttime time.Time, selectendtime time.Time, searchKey string) ([]models.RecordVitaminDay, error) {
	selectstartstr := fmt.Sprintf("%d-%02d-%02d", selectstarttime.Year(), selectstarttime.Month(), selectstarttime.Day())
	selectendstr := fmt.Sprintf("%d-%02d-%02d", selectendtime.Year(), selectendtime.Month(), selectendtime.Day())
	var vitaminstatistics []models.RecordVitaminDay

	db := self.GetDBmControl().Model(models.RecordVitaminDay{})

	db = db.Where("hid = ?", dhid)

	if searchKey != "" {
		db = db.Where("uid = ? ", searchKey)
	}

	db = db.Where("date_format(created_at, '%Y-%m-%d') >= ? ", selectstartstr)

	db = db.Where("date_format(created_at, '%Y-%m-%d') <= ? ", selectendstr)

	if dfid != -1 {
		db = db.Where("dfid = ?", dfid)
	}

	err := db.Find(&vitaminstatistics).Error

	if err != nil {
		return []models.RecordVitaminDay{}, err
	}

	return vitaminstatistics, err
}

// ! 查询疲劳值结算统计
func (self *DBMgr) SelectVitaminStatisticClear(dhid int64, selectstarttime time.Time, selectendtime time.Time) ([]models.RecordVitaminDayClear, error) {
	selectstartstr := fmt.Sprintf("%d-%02d-%02d", selectstarttime.Year(), selectstarttime.Month(), selectstarttime.Day())
	selectendstr := fmt.Sprintf("%d-%02d-%02d", selectendtime.Year(), selectendtime.Month(), selectendtime.Day())
	var vitaminstatisticclears []models.RecordVitaminDayClear

	db := self.GetDBmControl().Model(models.RecordVitaminDayClear{})

	db = db.Where("hid = ?", dhid)

	db = db.Where("date_format(created_at, '%Y-%m-%d') >= ? ", selectstartstr)

	db = db.Where("date_format(created_at, '%Y-%m-%d') <= ? ", selectendstr)

	err := db.Order("created_at desc").Find(&vitaminstatisticclears).Error

	if err != nil {
		return []models.RecordVitaminDayClear{}, err
	}

	return vitaminstatisticclears, err
}

// ! 查询疲劳值结算统计
func (self *DBMgr) SelectVitaminStatisticClearRecording(dhid int64, selecttime time.Time, vitaminLeft int64, vitaminMinus int64) (models.RecordVitaminDayClear, error) {
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var housevitaminrecordclear models.RecordVitaminDayClear

	db := self.GetDBmControl().Model(models.RecordVitaminDayClear{})

	db = db.Where("hid = ?", dhid)

	db = db.Where("date_format(created_at, '%Y-%m-%d') = ? ", selectstr)

	db = db.Where("recording = 1")

	err := db.Order("created_at desc").First(&housevitaminrecordclear).Error

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return models.RecordVitaminDayClear{}, err
		}

		housevitaminrecordclear.DHId = dhid
		housevitaminrecordclear.Recording = 1
		housevitaminrecordclear.VitaminCost = 0
		housevitaminrecordclear.VitaminCostRound = 0
		housevitaminrecordclear.VitaminCostBW = 0
		housevitaminrecordclear.VitaminLeft = vitaminLeft
		housevitaminrecordclear.VitaminMinus = vitaminMinus
		housevitaminrecordclear.VitaminPayment = 0

		nowTime := selecttime
		housevitaminrecordclear.CreatedAt = selecttime
		housevitaminrecordclear.UpdatedAt = selecttime

		beginTimeStr := fmt.Sprintf("%d-%02d-%02d 00:00:00", nowTime.Year(), nowTime.Month(), nowTime.Day())
		beginTime, _ := time.ParseInLocation("2006-01-02 15:04:05", beginTimeStr, time.Local)
		housevitaminrecordclear.BeginAt = beginTime

		endTimeStr := fmt.Sprintf("%d-%02d-%02d 23:59:59", nowTime.Year(), nowTime.Month(), nowTime.Day())
		endTime, _ := time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, time.Local)
		housevitaminrecordclear.EndAt = endTime

		err = self.db_M.Create(&housevitaminrecordclear).Error

		return housevitaminrecordclear, err
	}

	return housevitaminrecordclear, err
}

// ! 查询疲劳值结算统计
func (self *DBMgr) InsertVitaminStatisticClear(dhid int64, clearItem models.RecordVitaminDayClear) error {
	vitaminClear := models.RecordVitaminDayClear{}

	vitaminClear.DHId = dhid
	vitaminClear.Recording = 1
	vitaminClear.VitaminCost = 0
	vitaminClear.VitaminCostRound = 0
	vitaminClear.VitaminCostBW = 0
	vitaminClear.BeginAt = time.Now()
	vitaminClear.EndAt = clearItem.EndAt

	var err error
	//mysql
	if err = self.db_M.Create(&vitaminClear).Error; err != nil {
		return err
	}

	return nil
}

// ! 更新结扎数据打包
func (self *DBMgr) UpdateVitaminStatisticClear(dhid int64, endat time.Time, vitaminLeft int64, vitaminMinus int64) error {
	var housevitaminrecordclear models.RecordVitaminDayClear

	db := self.db_M.Model(models.RecordVitaminDayClear{})

	selectstr := fmt.Sprintf("%d-%02d-%02d", time.Now().Year(), time.Now().Month(), time.Now().Day())
	db = db.Where("date_format(created_at, '%Y-%m-%d') = ?", selectstr)

	err := db.Where("hid = ? and recording = 1", dhid).First(&housevitaminrecordclear).Error

	if err != nil {
		return err
	}

	housevitaminrecordclear.Recording = 0
	housevitaminrecordclear.EndAt = endat

	updateMap := make(map[string]interface{})
	updateMap["recording"] = housevitaminrecordclear.Recording
	updateMap["end_at"] = housevitaminrecordclear.EndAt
	updateMap["vitaminleft"] = vitaminLeft
	updateMap["vitaminminus"] = vitaminMinus

	//mysql
	err = self.db_M.Model(models.RecordVitaminDayClear{Id: housevitaminrecordclear.Id}).Updates(updateMap).Error

	return err
}

// ! 查询成员疲劳值节点信息
func (self *DBMgr) SelectVitaminMgrList(dhid int64) (map[int64]models.RecordVitaminMgrList, error) {
	//var vitaminmgrlist []model.RecordVitaminMgrList
	//
	//db := self.GetDBmControl().Model(model.RecordVitaminMgrList{})
	//
	//db = db.Where("hid = ?", dhid)
	//
	//db = db.Where("recording = 1")
	//
	//err := db.Find(&vitaminmgrlist).Error
	//
	//vitaminrecordmap := make(map[int64]model.RecordVitaminMgrList)
	//
	//if err != nil {
	//	return vitaminrecordmap, err
	//}
	//for _, record := range vitaminmgrlist {
	//	vitaminrecordmap[record.UId] = record
	//}
	//
	//return vitaminrecordmap, err

	vitaminrecordmap := make(map[int64]models.RecordVitaminMgrList)
	return vitaminrecordmap, nil
}

// ! 更新疲劳值节点信息
func (self *DBMgr) UpdateVitaminMgrList(dhid int64, uid int64, nodeVitamin int64, tx *gorm.DB) (err error) {
	var housevitaminmgr models.RecordVitaminMgrList
	if tx == nil {
		tx = self.GetDBmControl()
	}
	err = tx.Model(models.RecordVitaminMgrList{}).Where("hid = ? and uid = ? and recording = 1", dhid, uid).First(&housevitaminmgr).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			xlog.Logger().Error(err)
			return err
		}
	} else {
		housevitaminmgr.Recording = 0
		updateMap := make(map[string]interface{})
		updateMap["recording"] = housevitaminmgr.Recording
		//mysql
		err = tx.Model(models.RecordVitaminMgrList{Id: housevitaminmgr.Id}).Updates(updateMap).Error
		if err != nil {
			xlog.Logger().Error(err)
			return err
		}
	}

	//加入新节点统计信息
	newhousevitaminmgr := new(models.RecordVitaminMgrList)

	newhousevitaminmgr.DHId = dhid
	newhousevitaminmgr.UId = uid
	newhousevitaminmgr.Recording = 1
	newhousevitaminmgr.PreNodeVitamin = nodeVitamin
	newhousevitaminmgr.VitaminWinLoseCost = 0
	newhousevitaminmgr.VitaminPlayCost = 0
	newhousevitaminmgr.VitaminCostRound = 0
	newhousevitaminmgr.VitaminCostBW = 0
	newhousevitaminmgr.CreatedAt = time.Now()
	newhousevitaminmgr.UpdatedAt = time.Now()
	err = tx.Create(&newhousevitaminmgr).Error
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}
	return nil
}

// ! 获取玩家五天剩余疲劳值
func (self *DBMgr) SelectMemberLeftVitamin(dhid int64, uids []int64, selecttime int) ([]models.HouseMemberVitaminLog, error) {
	var housevitaminlogs []models.HouseMemberVitaminLog

	var queryStr interface{}

	db := self.db_M.Model(models.HouseMemberVitaminLog{})

	db = db.Where("hid = ?", dhid)

	selectTime := time.Now().AddDate(0, 0, selecttime)
	selectTimestr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	db = db.Where("date_format(created_at, '%Y-%m-%d') = ?", selectTimestr).Order("created_at asc")

	queryStr = db.QueryExpr()

	db = self.db_M.Model(models.HouseMemberVitaminLog{})

	err := db.Raw("select * from (?) as orderlogs group by uid", queryStr).Scan(&housevitaminlogs).Error

	if err != nil {
		return []models.HouseMemberVitaminLog{}, err
	}

	return housevitaminlogs, err
}

// ! 插入每天玩家剩余疲劳值
func (self *DBMgr) InsertHouseMemberDayLeft(dhid int64, housememers []HouseMember, statisticsdate string) error {

	db := GetDBMgr().GetDBmControl()
	tx := db.Begin()

	sql := "insert into `house_member_vitamin_day` (`hid`,`uid`,`vitaminleft`,`statisticsdate`) values"

	for key, housemember := range housememers {
		if key == len(housememers)-1 {
			sql += fmt.Sprintf("('%d', '%d', '%d', '%s');", housemember.DHId, housemember.UId, housemember.UVitamin, statisticsdate)
		} else {
			sql += fmt.Sprintf("('%d', '%d', '%d', '%s'),", housemember.DHId, housemember.UId, housemember.UVitamin, statisticsdate)
		}
	}

	err := tx.Exec(sql).Error

	if err != nil {
		tx.Rollback()
		xlog.Logger().Errorln("InsertHouseMemberDayLeft Error: ", err.Error())
	} else {
		tx.Commit()
	}

	return err
}

// ! 查询包厢每天剩余疲劳值
func (self *DBMgr) SelectHouseDayLeftVitamin(dhid int64, selecttime int) (int64, int64, error) {
	var memberleftvitamins []models.HouseMemberVitaminDay

	selectTime := time.Now().AddDate(0, 0, selecttime)
	selectTimestr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	db := self.db_M.Model(models.HouseMemberVitaminDay{})

	db = db.Where("hid = ?", dhid)

	db = db.Where("statisticsdate = ?", selectTimestr)

	err := db.Find(&memberleftvitamins).Error

	if err != nil {
		return 0, 0, err
	}

	VitaminLeft := int64(0)
	VitaminMinus := int64(0)
	for _, memberleftvitamin := range memberleftvitamins {
		if memberleftvitamin.VitaminLeft > 0 {
			VitaminLeft += memberleftvitamin.VitaminLeft
		} else {
			VitaminMinus += memberleftvitamin.VitaminLeft
		}
	}
	return VitaminLeft, VitaminMinus, err
}

// ! 查询玩家没玩剩余疲劳值
func (self *DBMgr) SelectHouseMemberLeftVitamin(dhid int64, uids []int64, selecttime int) (int64, int64, error) {
	var memberleftvitamins []models.HouseMemberVitaminDay

	selectTime := time.Now().AddDate(0, 0, selecttime)
	selectTimestr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	db := self.db_M.Model(models.HouseMemberVitaminDay{})

	if len(uids) == 0 {
		return 0, 0, nil
	}
	db = db.Where("statisticsdate = ?", selectTimestr)

	db = db.Where("hid = ?", dhid)

	sqlstr := "("
	for index, uid := range uids {
		sqlstr += fmt.Sprintf(" uid = %d", uid)
		if index != len(uids)-1 {
			sqlstr += " or"
		}
	}
	sqlstr += " )"

	db = db.Where(sqlstr)

	err := db.Find(&memberleftvitamins).Error

	if err != nil {
		return 0, 0, err
	}

	VitaminLeft := int64(0)
	VitaminMinus := int64(0)
	for _, memberleftvitamin := range memberleftvitamins {
		if memberleftvitamin.VitaminLeft > 0 {
			VitaminLeft += memberleftvitamin.VitaminLeft
		} else {
			VitaminMinus += memberleftvitamin.VitaminLeft
		}
	}
	return VitaminLeft, VitaminMinus, err
}

// ! 更新收支统计
func (self *DBMgr) UpdatePaymentsStatistic(dhid int64, vitaminpayment int64) error {
	var housevitaminrecordclear models.RecordVitaminDayClear

	db := self.db_M.Model(models.RecordVitaminDayClear{})

	db = db.Where("hid = ? and recording = 1", dhid)

	nowTime := time.Now()
	selectstr := fmt.Sprintf("%d-%02d-%02d", nowTime.Year(), nowTime.Month(), nowTime.Day())
	db = db.Where("date_format(created_at, '%Y-%m-%d') = ? ", selectstr)

	err := db.First(&housevitaminrecordclear).Error

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		housevitaminrecordclear.DHId = dhid
		housevitaminrecordclear.Recording = 1
		housevitaminrecordclear.VitaminCost = 0
		housevitaminrecordclear.VitaminCostRound = 0
		housevitaminrecordclear.VitaminCostBW = 0
		housevitaminrecordclear.VitaminPayment = vitaminpayment

		nowTime := time.Now()
		housevitaminrecordclear.CreatedAt = nowTime
		housevitaminrecordclear.UpdatedAt = nowTime

		beginTimeStr := fmt.Sprintf("%d-%02d-%02d 00:00:00", nowTime.Year(), nowTime.Month(), nowTime.Day())
		beginTime, _ := time.ParseInLocation("2006-01-02 15:04:05", beginTimeStr, time.Local)
		housevitaminrecordclear.BeginAt = beginTime

		endTimeStr := fmt.Sprintf("%d-%02d-%02d 23:59:59", nowTime.Year(), nowTime.Month(), nowTime.Day())
		endTime, _ := time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, time.Local)
		housevitaminrecordclear.EndAt = endTime

		err = self.db_M.Create(&housevitaminrecordclear).Error

		return err
	}

	housevitaminrecordclear.VitaminPayment += vitaminpayment

	updateMap := make(map[string]interface{})
	updateMap["vitaminpayment"] = housevitaminrecordclear.VitaminPayment

	//mysql
	err = self.db_M.Model(models.RecordVitaminDayClear{Id: housevitaminrecordclear.Id}).Updates(updateMap).Error

	return err
}

// ! 更新排位赛统计数据
func (self *DBMgr) updataGameMatchState(matchconfig *models.ConfigMatch, state int) (int64, error) {

	updateMap := make(map[string]interface{})
	updateMap["state"] = state

	//mysql
	if err := self.db_M.Model(models.ConfigMatch{}).Where("kind_id = ? AND typestr = ?  AND begindate = ?  AND begintime = ? ", matchconfig.KindId, matchconfig.TypeStr, matchconfig.BeginDate, matchconfig.BeginTime).Updates(updateMap).Error; err != nil {
		return 0, err
	}

	//redis
	err := self.db_R.UpdateMatchAttrs(matchconfig, state)
	if err != nil {
		return 0, nil
	}

	return 0, nil
}

// ! 更新排位赛统计数据
func (self *DBMgr) updataGameMatchFlag(matchconfig *models.ConfigMatch, flag int) (int64, error) {

	updateMap := make(map[string]interface{})
	updateMap["flag"] = flag

	//mysql
	if err := self.db_M.Model(models.ConfigMatch{}).Where("kind_id = ? AND typestr = ?  AND begindate = ?  AND begintime = ? ", matchconfig.KindId, matchconfig.TypeStr, matchconfig.BeginDate, matchconfig.BeginTime).Updates(updateMap).Error; err != nil {
		return 0, err
	}
	//redis
	err := self.db_R.UpdateMatchAttrs(matchconfig, flag)
	if err != nil {
		return 0, nil
	}

	return 0, nil
}

// !排位赛结束时更新排位赛统计数据的排位，
func (self *DBMgr) updataGameMatchTotalByranking(gameMatchTotal *models.GameMatchTotal) (int64, error) {

	db_gameMatchTotal, err := self.db_R.SelectGameMatchTotalSet(gameMatchTotal)

	honorAwards := GetServer().GetMatchHonorAwardConfig(gameMatchTotal, gameMatchTotal.SiteType)
	if db_gameMatchTotal != nil {
		//更新排位
		updateMap := make(map[string]interface{})
		updateMap["ranking"] = gameMatchTotal.Ranking
		db_gameMatchTotal.Ranking = gameMatchTotal.Ranking
		//mysql
		if err = self.db_M.Model(models.GameMatchTotal{Id: db_gameMatchTotal.Id}).Updates(updateMap).Error; err != nil {
			return 0, err
		}
	} else {
		newRecord := new(models.GameMatchTotal)
		newRecord.MatchKey = gameMatchTotal.MatchKey
		newRecord.UId = gameMatchTotal.UId
		newRecord.KindId = gameMatchTotal.KindId
		newRecord.CreatedAt = time.Now()
		newRecord.UpdatedAt = newRecord.CreatedAt
		newRecord.SiteType = gameMatchTotal.SiteType
		newRecord.Score = gameMatchTotal.Score
		newRecord.WinCount = gameMatchTotal.WinCount
		newRecord.TotalCount = gameMatchTotal.TotalCount
		newRecord.BeginDate = gameMatchTotal.BeginDate
		newRecord.BeginTime = gameMatchTotal.BeginTime
		newRecord.EndDate = gameMatchTotal.EndDate
		newRecord.EndTime = gameMatchTotal.EndTime
		newRecord.Ranking = gameMatchTotal.Ranking
		if honorAwards != nil {
			newRecord.Coupon = honorAwards.WinHonor*newRecord.WinCount + honorAwards.JionHonor*newRecord.TotalCount
		} else {
			newRecord.Coupon = newRecord.TotalCount
		}
		//mysql
		if err = self.db_M.Create(newRecord).Error; err != nil {
			return 0, err
		}
		db_gameMatchTotal = newRecord
	}

	//redis
	err = self.db_R.InsertGameMatchTotalSetByranking(db_gameMatchTotal)
	if err != nil {
		return 0, nil
	}

	return 0, nil
}

// !排位赛结束时更新排位赛统计数据的排位，
func (self *DBMgr) InsertGameMatchRecord(gameMatchCouponRecord *models.GameMatchCouponRecord) (int64, error) {

	if gameMatchCouponRecord != nil {
		//mysql
		if err := self.db_M.Create(gameMatchCouponRecord).Error; err != nil {
			return 0, err
		}
		//redis
		err := self.db_R.InsertGameMatchRankingRecordSet(gameMatchCouponRecord)
		if err != nil {
			return 0, nil
		}
	}
	return 0, nil
}

// ! 获取玩家邀请码
func (self *DBMgr) SelectPartnerInviteCode(dhid int64, uid int64) (*models.HousePartnerInviteCode, error) {
	housePartnerInviteCode := new(models.HousePartnerInviteCode)

	db := GetDBMgr().GetDBmControl().Model(models.HousePartnerInviteCode{})
	err := db.Where("usedhid = ? and useduid = ? and isused = ?", dhid, uid, true).First(housePartnerInviteCode).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			xlog.Logger().Error(err)
			return housePartnerInviteCode, err
		}

		updateMap := make(map[string]interface{})
		updateMap["usedhid"] = dhid
		updateMap["useduid"] = uid
		updateMap["isused"] = true
		updateMap["updated_at"] = time.Now()

		err = db.Model(models.HousePartnerInviteCode{}).Where("isused = ?", false).Limit(1).Update(updateMap).Error
		if err != nil {
			return housePartnerInviteCode, err
		}

		err := db.Model(models.HousePartnerInviteCode{}).Where("usedhid = ? and useduid = ? and isused = ?", dhid, uid, true).First(housePartnerInviteCode).Error
		if err != nil {
			return housePartnerInviteCode, err
		}
		return housePartnerInviteCode, err
	}

	return housePartnerInviteCode, nil
}

// ! 通过邀请码获取玩家数据
func (self *DBMgr) SelectPartnerByInviteCode(inviteCode string) (*models.HousePartnerInviteCode, error) {
	housePartnerInviteCode := new(models.HousePartnerInviteCode)

	db := GetDBMgr().GetDBmControl().Model(models.HousePartnerInviteCode{})
	err := db.Where("invitecode = ? and isused = ?", static.HF_Atoi64(inviteCode), true).First(housePartnerInviteCode).Error
	if err != nil {
		return housePartnerInviteCode, err
	}
	return housePartnerInviteCode, err
}

// ！从数据库读取配置
func (self *DBMgr) ReadAllConfig() error {
	var err error
	// 读取服务器配置
	if err := self.GetDBmControl().First(&GetServer().ConServers).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 读取游戏配置
	xlog.Logger().Infoln("read config_games...")
	if err = self.GetDBmControl().Find(&GetServer().ConGame).Error; err != nil {
		return err
	}

	// 加载观战配置
	LoadGameSupportWatch()

	// 读取包厢配置
	xlog.Logger().Infoln("read config_house....")
	if err = self.GetDBmControl().First(GetServer().ConHouse).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	xlog.Logger().Infoln(GetServer().ConHouse)
	// 读取房间配置
	xlog.Logger().Infoln("read config_room....")
	if err = self.GetDBmControl().Find(&GetServer().ConSite).Error; err != nil {
		return err
	}
	// 读取排位赛配置
	xlog.Logger().Infoln("read config_match....")
	if err = self.GetDBmControl().Find(&GetServer().ConMatch).Error; err != nil {
		return err
	}
	// 读取排位赛奖励清单配置
	xlog.Logger().Infoln("read config_matchaward....")
	if err = self.GetDBmControl().Find(&GetServer().ConMatchAward).Error; err != nil {
		return err
	}

	// 读取渠道配置
	if err := self.GetDBmControl().Find(&GetServer().ConApp).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xlog.Logger().Error("err:", err)
		return err
	}

	// 读取政府实名认证配置
	if err := self.GetDBmControl().Find(&GetServer().ConGovAuth).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xlog.Logger().Error("err:", err)
		return err
	}

	// 读取政府实名认证配置
	if err := self.GetDBmControl().Find(&GetServer().ConBattleLevel).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xlog.Logger().Error("err:", err)
		return err
	}

	if err = self.GetDBmControl().First(GetServer().ConSpinBase).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 读取政府实名认证配置
	if err := self.GetDBmControl().Find(&GetServer().ConSpinAward).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xlog.Logger().Error("err:", err)
		return err
	}

	slices.SortFunc(GetServer().ConSpinAward, func(a, b *models.ConfigSpinAward) int {
		return cmp.Compare(a.Seq, b.Seq)
	})

	// 读取政府实名认证配置
	if err := self.GetDBmControl().Find(&GetServer().ConCheckIn).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xlog.Logger().Error("err:", err)
		return err
	}

	slices.SortFunc(GetServer().ConCheckIn, func(a, b *models.ConfigCheckin) int {
		return cmp.Compare(a.Id, b.Id)
	})

	// 读取web支付配置
	// if err := self.GetDBmControl().Find(&GetServer().ConWebPay).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
	// 	syslog.Logger().Error("err:", err)
	// 	return err
	// }
	return err
}

// //! 执行语句
// func (self *DBMgr) Exec(query string, args ...interface{}) (int64, int64) {
//
//		stmt, err := self.m_db.Prepare(query)
//		defer stmt.Close()
//		if err != nil {
//			return 0, 0
//		}
//
//		result, err := stmt.Exec(args...)
//		if err != nil {
//			xlog.Logger().Errorln("db exec fail! query err:", err)
//			xlog.Logger().Errorln("db exec fail! query err:", err)
//			return 0, 0
//		}
//
//		LastInsertId := int64(0)
//		LastInsertId, err = result.LastInsertId()
//		if err != nil {
//			xlog.Logger().Errorln("db exec-LastInsertId fail! query err:", err)
//			xlog.Logger().Errorln("db exec-LastInsertId fail! query err:", err)
//		}
//
//		RowsAffected := int64(0)
//		RowsAffected, err = result.RowsAffected()
//		if err != nil {
//			xlog.Logger().Errorln("db exec-RowsAffected fail! query err:", err)
//			xlog.Logger().Errorln("db exec-RowsAffected fail! query err:", err)
//		}
//
//		return LastInsertId, RowsAffected
//	}
//
// //! 得到一条数据
//
//	func (self *DBMgr) GetOneData(query string, struc interface{}) bool {
//		rows, err := self.m_db.Query(query)
//		defer rows.Close()
//		if rows == nil || err != nil {
//			xlog.Logger().Errorln("db GetOneData fail! query:", query, ",err:", err)
//			xlog.Logger().Errorln("db GetOneData fail! query:", query, ",err:", err)
//			return false
//		}
//
//		//! 得到反射
//		s := reflect.ValueOf(struc).Elem()
//		num := s.NumField()
//		data := make([]interface{}, 0)
//		for i := 0; i < num; i++ {
//			ki := s.Field(i).Kind()
//			if ki != reflect.Slice && ki != reflect.Int && ki != reflect.Int64 && ki != reflect.Int8 && ki != reflect.String && ki != reflect.Float32 && ki != reflect.Float64 {
//				continue
//			}
//			data = append(data, s.Field(i).Addr().Interface())
//		}
//
//		for rows.Next() {
//			err = rows.Scan(data...)
//			if err != nil {
//				xlog.Logger().Errorln("db GetOneData-Scan fail! query:", query, ",err:", err)
//				xlog.Logger().Errorln("db GetOneData-Scan fail! query:", query, ",err:", err)
//				return false
//			}
//			break
//		}
//
//		return true
//	}
//
// //! 得到多条数据
//
//	func (self *DBMgr) GetAllData(query string, struc interface{}) []interface{} {
//		rows, err := self.m_db.Query(query)
//		defer rows.Close()
//		if rows == nil || err != nil {
//			xlog.Logger().Errorln("db GetAllData fail! query:", query, ",err:", err)
//			xlog.Logger().Errorln("db GetAllData fail! query:", query, ",err:", err)
//			return nil
//		}
//
//		//! 得到反射
//		s := reflect.ValueOf(struc).Elem()
//		num := s.NumField()
//		data := make([]interface{}, 0)
//		for i := 0; i < num; i++ {
//			ki := s.Field(i).Kind()
//			if ki != reflect.Slice && ki != reflect.Int && ki != reflect.Int64 && ki != reflect.Int8 && ki != reflect.String && ki != reflect.Float32 && ki != reflect.Float64 {
//				continue
//			}
//			data = append(data, s.Field(i).Addr().Interface())
//		}
//
//		result := make([]interface{}, 0)
//		for rows.Next() {
//			err = rows.Scan(data...)
//			if err != nil {
//				xlog.Logger().Errorln("db GetAllData-Scan fail! query:", query, ",err:", err)
//				xlog.Logger().Errorln("db GetAllData-Scan fail! query:", query, ",err:", err)
//				return nil
//			}
//			newObj := reflect.New(reflect.TypeOf(struc).Elem()).Elem()
//			newObj.Set(s)
//			result = append(result, newObj.Addr().Interface())
//		}
//
//		return result
//	}
func (self *DBMgr) GetHouseMemMap(hid int64) static.HouseMemberMap {
	meme := make([]*static.HouseMember, 0)
	err := self.GetDBrControl().RedisV2.HVals(fmt.Sprintf(consts.REDIS_KEY_HOUSE_MEMBER, hid)).ScanSlice(&meme)
	if err != nil {
		xlog.Logger().Error(err)
	}
	res := make(static.HouseMemberMap)
	for i := 0; i < len(meme); i++ {
		res[meme[i].UId] = meme[i]
	}
	return res
}

// ! 盟主获取删除楼层的历时数据
func (self *DBMgr) SelectDelFloorPartnerProfit(dhid int64, delFloorMap map[int64]models.HouseFloorDelMsg) ([]models.HousePartnerRoyaltyDetailItem, error) {
	var detailItems []models.HousePartnerRoyaltyDetailItem
	var err error
	db := self.GetDBmControl().Table("house_partner_royalty_detail")
	sql := "select sum(selfprofit) as selfprofit, sum(subprofit) as subprofit, Sum(providerround) as validtimes, dfid as dfid from house_partner_royalty_detail where "

	index := 0
	for _, defFloor := range delFloorMap {
		selectTime := time.Unix(defFloor.CreateStamp, 0)
		selectTimeStr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())
		sql += fmt.Sprintf(`(dhid = %d and dfid = %d and `, dhid, defFloor.DFId)
		sql += `date_format(created_at, '%Y-%m-%d') = '`
		sql += selectTimeStr
		sql += `') `
		index++
		if index < len(delFloorMap) {
			sql += " or "
		}
	}

	sql += "group by dfid"

	err = db.Raw(sql).Scan(&detailItems).Error

	return detailItems, err
}

// ! 合伙人获取删除楼层的历时数据
func (self *DBMgr) SelectDelFloorPartnerProfitByPartner(dhid int64, delFloorMap map[int64]models.HouseFloorDelMsg, partner int64) ([]models.HousePartnerRoyaltyDetailItem, error) {
	var detailItems []models.HousePartnerRoyaltyDetailItem
	var err error
	db := self.GetDBmControl().Table("house_partner_royalty_detail")
	sql := "select sum(selfprofit) as selfprofit, sum(subprofit) as subprofit, Sum(providerround) as validtimes, dfid as dfid from house_partner_royalty_detail where beneficiary = ?"

	index := 0
	for _, defFloor := range delFloorMap {
		if index == 0 {
			sql += " and ("
		}
		selectTime := time.Unix(defFloor.CreateStamp, 0)
		selectTimeStr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())
		sql += fmt.Sprintf(`(dhid = %d and dfid = %d and `, dhid, defFloor.DFId)
		sql += `date_format(created_at, '%Y-%m-%d') = '`
		sql += selectTimeStr
		sql += `') `
		index++
		if index < len(delFloorMap) {
			sql += " or "
		} else {
			sql += ")"
		}
	}

	sql += " group by dfid"

	err = db.Raw(sql, partner).Scan(&detailItems).Error

	return detailItems, err
}

// ! 盟主获取删除楼层的所有合伙人历时数据
func (self *DBMgr) SelectDelFloorPartnerProfitWithFloorByPartner(dhid int64, delFloorMap map[int64]models.HouseFloorDelMsg) ([]models.HousePartnerRoyaltyDetailItem, error) {
	var detailItems []models.HousePartnerRoyaltyDetailItem
	var err error
	db := self.GetDBmControl().Table("house_partner_royalty_detail")
	sql := "select sum(selfprofit) as selfprofit, sum(subprofit) as subprofit, Sum(providerround) as validtimes, beneficiary as beneficiary from house_partner_royalty_detail where "

	for _, defFloor := range delFloorMap {
		selectTime := time.Unix(defFloor.CreateStamp, 0)
		selectTimeStr := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())
		sql += fmt.Sprintf(`(dhid = %d and dfid = %d and `, dhid, defFloor.DFId)
		sql += `date_format(created_at, '%Y-%m-%d') = '`
		sql += selectTimeStr
		sql += `') `

		break
	}

	sql += "group by beneficiary"

	err = db.Raw(sql).Scan(&detailItems).Error

	return detailItems, err
}

// ! 查询合伙人分成(盟主视角)
func (self *DBMgr) SelectPartnerProfitWithAllPartners(dhid int64, floorIndex int, floorId int64, day int) (map[int64]models.HousePartnerRoyaltyDetailItem, error) {
	if day > 0 {
		day = -day
	}
	selectDay := time.Now().AddDate(0, 0, day)
	selectstr := fmt.Sprintf("%d-%02d-%02d", selectDay.Year(), selectDay.Month(), selectDay.Day())

	var detailItems []models.HousePartnerRoyaltyDetailItem
	var err error
	db := self.GetDBmControl()
	if floorIndex != -1 {
		sql := `select beneficiary as beneficiary, Sum(selfprofit) as selfprofit, Sum(subprofit) as subprofit, Sum(providerround) as validtimes from house_partner_royalty_detail where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? and (dfloorindex = ? or dfid = ?) group by beneficiary`
		err = db.Raw(sql, dhid, selectstr, floorIndex, floorId).Scan(&detailItems).Error
	} else {
		sql := `select beneficiary as beneficiary, Sum(selfprofit) as selfprofit, Sum(subprofit) as subprofit, Sum(providerround) as validtimes from house_partner_royalty_detail where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? group by beneficiary`
		err = db.Raw(sql, dhid, selectstr).Scan(&detailItems).Error
	}

	detailMap := make(map[int64]models.HousePartnerRoyaltyDetailItem)
	for _, item := range detailItems {
		detailMap[item.Beneficiary] = item
	}
	return detailMap, err
}

// ! 查询合伙人分成(合伙人视角详情)
func (self *DBMgr) SelectPartnerProfitWithPartnerDetail(dhid int64, floorIndex int, floorId int64, day int, partnerId int64) (map[int64]models.HousePartnerRoyaltyDetailItem, map[int64]models.HousePartnerRoyaltyDetailItem, error) {
	selectDay := time.Now().AddDate(0, 0, day)
	selectstr := fmt.Sprintf("%d-%02d-%02d", selectDay.Year(), selectDay.Month(), selectDay.Day())

	var detailItems []models.HousePartnerRoyaltyDetailItem
	var partnerItems []models.HousePartnerRoyaltyDetailItem
	var err error
	db := self.GetDBmControl()
	if floorIndex != -1 {
		sql := `select Sum(selfprofit) as selfprofit, playeruser as beneficiary, Sum(providerround) as validtimes 
                from house_partner_royalty_detail 
                where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? and (dfloorindex = ? or dfid = ?) and beneficiary = ? and playerpartner = ?
                group by playeruser`
		err = db.Raw(sql, dhid, selectstr, floorIndex, floorId, partnerId, partnerId).Scan(&detailItems).Error
	} else {
		sql := `select Sum(selfprofit) as selfprofit, playeruser as beneficiary, Sum(providerround) as validtimes 
                from house_partner_royalty_detail 
                where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? and beneficiary = ? and playerpartner = ?
                group by playeruser`
		err = db.Raw(sql, dhid, selectstr, partnerId, partnerId).Scan(&detailItems).Error
	}

	if floorIndex != -1 {
		sql := `select Sum(subprofit) as subprofit, playerpartner as beneficiary, Count(1) as validtimes 
                from house_partner_royalty_detail 
                where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? and (dfloorindex = ? or dfid = ?) and beneficiary = ? and playerpartner <> ?
                group by playerpartner`
		err = db.Raw(sql, dhid, selectstr, floorIndex, floorId, partnerId, partnerId).Scan(&partnerItems).Error
	} else {
		sql := `select Sum(subprofit) as subprofit, playerpartner as beneficiary, Count(1) as validtimes 
                from house_partner_royalty_detail 
                where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? and beneficiary = ?  and playerpartner <> ? 
                group by playerpartner`
		err = db.Raw(sql, dhid, selectstr, partnerId, partnerId).Scan(&partnerItems).Error
	}

	detailMap := make(map[int64]models.HousePartnerRoyaltyDetailItem)
	for _, item := range detailItems {
		detailMap[item.Beneficiary] = item
	}
	partnerMap := make(map[int64]models.HousePartnerRoyaltyDetailItem)
	for _, item := range partnerItems {
		partnerMap[item.Beneficiary] = item
	}

	return detailMap, partnerMap, err
}

// ! 获得队长属性
func (self *DBMgr) SelectHouseAllPartnerAttr(dhid int64, pid ...int64) (map[int64]models.HousePartnerAttr, error) {
	modelsMeta := make([]models.HousePartnerAttr, 0)
	var err error
	if len(pid) > 0 {
		err = self.GetDBmControl().Find(&modelsMeta, "dhid = ? AND uid in (?)", dhid, pid).Error
	} else {
		err = self.GetDBmControl().Find(&modelsMeta, "dhid = ?", dhid).Error
	}
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	ret := make(map[int64]models.HousePartnerAttr)
	for _, model := range modelsMeta {
		ret[model.Uid] = model
	}
	return ret, nil
}

// ! 获取包厢指定日期游戏点赞信息
func (self *DBMgr) SelectHouseRecordGameLike(dhid int64, optusertype int, opttype int, daytype int, timerange string) (map[string]bool, error) {
	selectTime := time.Now().AddDate(0, 0, daytype)
	liketime := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	var likeList []models.HouseRecordLike

	likeMap := make(map[string]bool)

	optType := opttype
	db := self.GetDBmControl()
	err := db.Model(models.HouseRecordLike{}).Where("dhid = ? and opttype = ? and optusertype = ? and liketime = ? and time_range = ?", dhid, optType, optusertype, liketime, timerange).Find(&likeList).Error
	if err != nil {
		return likeMap, err
	}
	for _, like := range likeList {
		if like.IsLike {
			likeMap[like.GameNum] = true
		}
	}
	return likeMap, nil
}

// ! 获取包厢指定日期用户点赞信息
func (self *DBMgr) SelectHouseRecordUserLike(dhid int64, optusertype int, daytype int, timerange string) (map[int64]bool, error) {
	selectTime := time.Now().AddDate(0, 0, daytype)
	liketime := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	var likeList []models.HouseRecordLike

	likeMap := make(map[int64]bool)

	optType := models.OptTypeUser
	db := self.GetDBmControl()
	err := db.Model(models.HouseRecordLike{}).Where("dhid = ? and opttype = ? and optusertype = ? and liketime = ? and time_range = ?", dhid, optType, optusertype, liketime, timerange).Find(&likeList).Error
	if err != nil {
		return likeMap, err
	}
	for _, like := range likeList {
		if like.IsLike {
			likeMap[like.LikeUser] = true
		}
	}
	return likeMap, nil
}

// ! 获取包厢指定日期团队统计中的用户点赞信息
func (self *DBMgr) SelectHouseRecordTeamLike(dhid int64, optusertype int, liketime string, timerange string) (map[int64]bool, error) {
	var likeList []models.HouseRecordLike
	likeMap := make(map[int64]bool)

	optType := models.OptTypeTeam
	db := self.GetDBmControl()
	err := db.Model(models.HouseRecordLike{}).Where("dhid = ? and opttype = ? and optusertype = ? and liketime = ? and time_range = ?", dhid, optType, optusertype, liketime, timerange).Find(&likeList).Error
	if err != nil {
		return likeMap, err
	}

	for _, like := range likeList {
		if like.IsLike {
			likeMap[like.LikeUser] = true
		}
	}

	return likeMap, nil
}

// ! 更新包厢指定日期游戏点赞信息
func (self *DBMgr) UpdateRecordGameLike(dhid int64, daytype int, gamenum string, optusertype int, opttype int, islike bool, timerange string) error {
	selectTime := time.Now().AddDate(0, 0, daytype)
	liketime := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	var gameLike models.HouseRecordLike

	optType := opttype
	db := self.GetDBmControl()
	err := db.Model(models.HouseRecordLike{}).Where("dhid = ? and opttype = ? and gamenum = ? and liketime = ? and optusertype = ? and time_range = ?", dhid, optType, gamenum, liketime, optusertype, timerange).First(&gameLike).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == nil {
		updateMap := make(map[string]interface{})
		updateMap["islike"] = islike
		updateMap["updated_at"] = time.Now()
		if err = db.Model(models.HouseRecordLike{Id: gameLike.Id}).Updates(updateMap).Error; err != nil {
			return err
		}
	} else {
		gameLike.GameNum = gamenum
		gameLike.DHid = dhid
		gameLike.LikeUser = 0
		gameLike.OptType = optType
		gameLike.OptUserType = optusertype
		gameLike.LikeTime = liketime
		gameLike.IsLike = islike
		gameLike.CreatedAt = time.Now()
		gameLike.UpdatedAt = time.Now()
		gameLike.TimeRange = timerange
		if err := db.Create(&gameLike).Error; err != nil {
			return err
		}
	}

	return nil
}

// ! 更新包厢指定日期玩家点赞信息
func (self *DBMgr) UpdateRecordUserLike(dhid int64, daytype int, likeuser int64, optusertype int, islike bool, timerange string, isTeamLike bool) error {
	selectTime := time.Now().AddDate(0, 0, daytype)
	liketime := fmt.Sprintf("%d-%02d-%02d", selectTime.Year(), selectTime.Month(), selectTime.Day())

	var gameLike models.HouseRecordLike

	// 标识 普通用户点赞 还是 团队统计里的 用户点赞
	optType := models.OptTypeUser
	if isTeamLike {
		optType = models.OptTypeTeam
	}
	db := self.GetDBmControl()
	err := db.Model(models.HouseRecordLike{}).Where("dhid = ? and opttype = ? and likeuser = ? and liketime = ? and optusertype = ? and time_range = ?", dhid, optType, likeuser, liketime, optusertype, timerange).First(&gameLike).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == nil {
		updateMap := make(map[string]interface{})
		updateMap["islike"] = islike
		updateMap["updated_at"] = time.Now()
		if err = db.Model(models.HouseRecordLike{Id: gameLike.Id}).Updates(updateMap).Error; err != nil {
			return err
		}
	} else {
		gameLike.GameNum = ""
		gameLike.DHid = dhid
		gameLike.LikeUser = likeuser
		gameLike.OptType = optType
		gameLike.OptUserType = optusertype
		gameLike.LikeTime = liketime
		gameLike.IsLike = islike
		gameLike.CreatedAt = time.Now()
		gameLike.UpdatedAt = time.Now()
		gameLike.TimeRange = timerange
		if err := db.Create(&gameLike).Error; err != nil {
			return err
		}
	}

	return nil
}

// 记录玩家离开包厢队长信息(离开包厢使用)
func (self *DBMgr) AddMemberLeaveHousePartner(dhid, uid, partner int64) error {
	var leaveLink models.HouseMemberLeaveLink
	leaveLink.DHId = dhid
	leaveLink.UId = uid
	leaveLink.Partner = partner
	leaveLink.CreatedAt = time.Now()

	db := self.GetDBmControl()
	err := db.Save(&leaveLink).Error
	if err != nil {
		xlog.Logger().Errorln("AddMemberLeaveHousePartner err: ", err.Error())
	}
	return err
}

// 移除玩家离开包厢队长信息(加入包厢使用)
func (self *DBMgr) RemoveMemberLeaveHousePartner(dhid, uid int64) error {
	db := self.GetDBmControl()
	err := db.Delete(&models.HouseMemberLeaveLink{}, "dhid = ? and uid = ?", dhid, uid).Error
	if err != nil {
		xlog.Logger().Errorln("RemoveMemberLeaveHousePartner err: ", err.Error())
	}
	return err
}

// 获取队长历史成员
func (self *DBMgr) SelectLeaveHousePartnerMember(dhid int64, partner int64, time string) (map[int64]bool, error) {
	var members []models.HouseMemberLeaveLink

	memMap := make(map[int64]bool)
	db := self.GetDBmControl()
	err := db.Model(models.HouseMemberLeaveLink{}).Where("dhid = ? and partner = ? and created_at >= ? ", dhid, partner, time).Find(&members).Error
	if err != nil {
		return memMap, err
	}

	for _, member := range members {
		memMap[member.UId] = true
	}
	return memMap, err
}

func (self *DBMgr) SelectAllLeaveHousePartnerMember(dhid int64, time string) (map[int64]int64, error) {
	var members []models.HouseMemberLeaveLink

	memMap := make(map[int64]int64)
	db := self.GetDBmControl()
	err := db.Model(models.HouseMemberLeaveLink{}).Where("dhid = ? and created_at >= ?", dhid, time).Find(&members).Error
	if err != nil {
		return memMap, err
	}

	for _, member := range members {
		memMap[member.UId] = member.Partner
	}
	return memMap, err
}

func (self *DBMgr) SelectPlayLastGameHouseAreaCode(uid int64) int {
	var recordTotal models.RecordGameTotal
	db := self.GetDBmControl()
	err := db.Model(models.RecordGameTotal{}).Where("uid = ? and hid > 0", uid).First(&recordTotal).Error
	if err == gorm.ErrRecordNotFound {
		person, err := GetDBMgr().GetDBrControl().GetPerson(uid)
		if err != nil {
			return 0
		}
		return static.HF_Atoi(person.Area)
	} else if err != nil {
		return 0
	}

	house := GetDBMgr().GetDBrControl().GetHouse(fmt.Sprintf(consts.REDIS_KEY_HOUSE_INFO, recordTotal.HId))

	if house != nil {
		xlog.Logger().Errorf("SelectPlayLastGameHouseAreaCode house err: house %d = nil ", recordTotal.HId)
		return 0
	}
	return static.HF_Atoi(house.Area)
}

func (self *DBMgr) SelectRecommentGameByAreaCode(code int) string {
	var reRecommendGame models.ConfigRecommendGame
	db := self.GetDBmControl()
	err := db.Model(models.ConfigRecommendGame{}).Where("areacode = ?", code).First(&reRecommendGame).Error
	if err != nil {
		return "跑的快"
	} else {
		return reRecommendGame.GameName
	}
}

// 查询队长设置的低分局数据
func (self *DBMgr) SelectPartnerLowScoreVal(dhid int64, pid int64) (map[int64]int64, error) {
	itemMap := make(map[int64]int64)

	var itemList []models.HousePartnerSetting
	err := self.GetDBmControl().Model(models.HousePartnerSetting{}).Where("hid = ? and pid = ?", dhid, pid).Find(&itemList).Error
	if err != nil {
		return itemMap, err
	}

	for _, item := range itemList {
		itemMap[item.Fid] = int64(item.LowScoreVal)
	}

	return itemMap, nil
}

func (self *DBMgr) SelectFloorPlayStatistics(Hid int64, DFid int, selectTime1 time.Time, selectTime2 time.Time) (map[int64]int, map[string]map[int64]struct{}, error) {
	// 查询时间
	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())

	// 查询记录
	var (
		retList []models.QueryMemberGameRecordResult
		err     error
	)
	// 查到的都是时间段内的有效局 和 玩家信息
	if DFid == -1 {
		sql := `select uid, game_num, fid, (win_score / radix) as score from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6`
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2).Scan(&retList).Error
	} else {
		sql := `select uid, game_num, fid, (win_score / radix) as score from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 and dfid = ?`
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2, DFid).Scan(&retList).Error
	}
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, nil, err
	}
	// 玩家对应场次
	uidMap := make(map[int64]int)
	// 局数对应玩家人数
	gameNumMap := make(map[string]map[int64]struct{})
	for _, record := range retList {
		// 不要改这里
		uidMap[record.Uid]++
		userMap, ok := gameNumMap[record.GameNum]
		if !ok {
			userMap = make(map[int64]struct{})
		}
		userMap[record.Uid] = struct{}{}
		gameNumMap[record.GameNum] = userMap
	}
	return uidMap, gameNumMap, nil
}

// 查询合伙人名下所有成员的战绩
func (self *DBMgr) SelectPartnerMemberStatisticsWithTotal(Pid int64, Hid int64, DFid int, selectTime1 time.Time, selectTime2 time.Time) (map[int64]models.QueryMemberStatisticsResult, error) {
	// 统计结果
	statisticsMap := make(map[int64]models.QueryMemberStatisticsResult)

	// 查询合伙人设置的楼层低分局数值
	floorLowScoreMap, err := self.SelectPartnerLowScoreVal(Hid, Pid)
	if err != nil {
		return statisticsMap, err
	}

	// 查询时间
	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())

	// 查询记录
	var retList []models.QueryMemberGameRecordResult
	if DFid == -1 {
		sql := `select uid, game_num, fid, (win_score / radix) as score from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6`
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2).Scan(&retList).Error
	} else {
		sql := `select uid, game_num, fid, (win_score / radix) as score from record_game_total where hid = ? and created_at >= ? and created_at < ? and score_kind <> 6 and dfid = ?`
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2, DFid).Scan(&retList).Error
	}
	if err != nil {
		return statisticsMap, err
	}

	// 找到每一局的大赢家分数
	bigWinnerScoreMap := make(map[string]float64)
	for _, record := range retList {
		if _, ok := bigWinnerScoreMap[record.GameNum]; ok {
			if record.Score > bigWinnerScoreMap[record.GameNum] {
				bigWinnerScoreMap[record.GameNum] = record.Score
			}
		} else {
			bigWinnerScoreMap[record.GameNum] = record.Score
		}
	}

	// 遍历战绩记录 生成统计数据
	for _, record := range retList {
		if item, ok := statisticsMap[record.Uid]; ok {
			// 更新记录
			item.TotalScore += record.Score
			item.PlayTimes++
			// 本局的大赢家分数
			var bigWinnerScore float64
			if score, ok := bigWinnerScoreMap[record.GameNum]; ok {
				bigWinnerScore = score
			}
			// 低分局判断
			if lowScoreVal, ok := floorLowScoreMap[int64(record.FId)]; ok {
				// 更新统计信息
				if bigWinnerScore < static.SwitchVitaminToF64(lowScoreVal) {
					// 低分局
				} else {
					// 有效局
					item.ValidTimes++
					if static.SwitchF64ToVitamin(record.Score) == static.SwitchF64ToVitamin(bigWinnerScore) && static.SwitchF64ToVitamin(record.Score) > 0 {
						item.BigWinTimes++
					}
				}
			} else {
				// 没有设置低分局
				item.ValidTimes++
				if static.SwitchF64ToVitamin(record.Score) == static.SwitchF64ToVitamin(bigWinnerScore) && static.SwitchF64ToVitamin(record.Score) > 0 {
					item.BigWinTimes++
				}
			}
			// 更新记录
			statisticsMap[record.Uid] = item
		} else {
			// 新增记录
			var ret models.QueryMemberStatisticsResult
			ret.Uid = record.Uid
			ret.TotalScore = record.Score
			ret.PlayTimes = 1
			// 本局的大赢家分数
			var bigWinnerScore float64
			if score, ok := bigWinnerScoreMap[record.GameNum]; ok {
				bigWinnerScore = score
			}
			// 低分局判断
			if lowScoreVal, ok := floorLowScoreMap[int64(record.FId)]; ok {
				// 更新统计信息
				if bigWinnerScore < static.SwitchVitaminToF64(lowScoreVal) {
					// 低分局
					ret.ValidTimes = 0
					ret.BigWinTimes = 0
				} else {
					// 有效局
					ret.ValidTimes = 1
					if static.SwitchF64ToVitamin(record.Score) == static.SwitchF64ToVitamin(bigWinnerScore) && static.SwitchF64ToVitamin(record.Score) > 0 {
						ret.BigWinTimes = 1
					} else {
						ret.BigWinTimes = 0
					}
				}
			} else {
				// 没有设置低分局
				ret.ValidTimes = 1
				if static.SwitchF64ToVitamin(record.Score) == static.SwitchF64ToVitamin(bigWinnerScore) && static.SwitchF64ToVitamin(record.Score) > 0 {
					ret.BigWinTimes = 1
				} else {
					ret.BigWinTimes = 0
				}
			}
			// 插入新记录
			statisticsMap[record.Uid] = ret
		}
	}

	return statisticsMap, nil
}

// 按照日期查询包厢战绩（合伙人自定义低分局查询）
func (self *DBMgr) SelectPartnerGameRecordByDate(hid int64, floorIndex int, uid int64, bw bool, puid int64, memMap map[int64]HouseMember, selectTime1 time.Time, selectTime2 time.Time, searchkey string, likeFlagMap map[string]bool, likeFlag int, roundType int) ([]static.GameRecordDetal, int, int, int, int, int, float64, int, int, int, int, error) {
	// 查询包厢所在区域是否显示正在游戏的战绩记录
	bIsShowPlayingRecord := false
	/*
		house := GetClubMgr().GetClubHouseById(hid)
		if house != nil {
			var areaInfo model.Area
			db := self.GetDBmControl().Model(model.Area{})
			db = db.Where("area_id = ?", house.DBClub.Area)
			err := db.Find(&areaInfo).Error
			if err == nil {
				bIsShowPlayingRecord = areaInfo.IsShowPlayingRecord == 1
			}
		}
	*/

	var gameTotals []models.RecordGameTotal

	db := self.GetDBmControl()

	db = self.GetDBmControl().Model(models.RecordGameTotal{})

	selectDate1 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day(), selectTime1.Hour(), selectTime1.Minute(), selectTime1.Second())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day(), selectTime2.Hour(), selectTime2.Minute(), selectTime2.Second())

	// 牌局类型 牌局类型 0 全部 1 完整局数 2 中途解散 3 低分局
	if roundType == 0 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ?", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and score_kind != 6", hid, selectDate1, selectDate2)
		}
	} else if roundType == 1 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and play_count = round and halfwaydismiss = 0", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and play_count = round and halfwaydismiss = 0 and score_kind != 6", hid, selectDate1, selectDate2)
		}
	} else if roundType == 2 {
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and (play_count < round or halfwaydismiss = 1) ", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and (play_count < round or halfwaydismiss = 1) and score_kind != 6", hid, selectDate1, selectDate2)
		}
	} else if roundType == 3 {
		// 合伙人可自定义低分局  在 查询出来的结果中 再作剔除
		if bIsShowPlayingRecord {
			db = db.Where("hid = ? and created_at >= ? and created_at < ?", hid, selectDate1, selectDate2)
		} else {
			db = db.Where("hid = ? and created_at >= ? and created_at < ? and score_kind != 6", hid, selectDate1, selectDate2)
		}
	}

	var err error
	err = db.Find(&gameTotals).Error

	if err != nil {
		return []static.GameRecordDetal{}, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, err
	}

	// 每局的大赢家分数
	bwUserScoreMap := make(map[string]int)
	// 战绩记录
	gameTotalMap := make(map[string][]models.RecordGameTotal)

	// 先找出每局的大赢家分数
	for i := 0; i < len(gameTotals); i++ {
		gameRecord := gameTotals[i]
		_, ok := bwUserScoreMap[gameRecord.GameNum]
		if !ok {
			bwUserScoreMap[gameRecord.GameNum] = gameRecord.WinScore
		}
		// 更新大赢家分数
		if gameRecord.WinScore > bwUserScoreMap[gameRecord.GameNum] {
			bwUserScoreMap[gameRecord.GameNum] = gameRecord.WinScore
		}
	}

	// 查询合伙人设置的楼层低分局数值
	floorLowScoreMap, _ := self.SelectPartnerLowScoreVal(hid, puid)

	// 更新有效局
	for i := 0; i < len(gameTotals); i++ {
		gameRecord := gameTotals[i]
		// 低分局判断
		if lowScoreVal, ok := floorLowScoreMap[int64(gameRecord.FId)]; ok {
			if gameRecord.Radix == 0 {
				gameRecord.Radix = 1
			}
			if float64(bwUserScoreMap[gameRecord.GameNum])/float64(gameRecord.Radix) < static.SwitchVitaminToF64(lowScoreVal) {
				gameTotals[i].IsValidRound = false
			} else {
				gameTotals[i].IsValidRound = true
			}
		} else {
			// 没有设置低分局
			gameTotals[i].IsValidRound = true
		}
	}

	// 只筛选低分局记录
	if roundType == 3 {
		var newGameTotals []models.RecordGameTotal
		for i := 0; i < len(gameTotals); i++ {
			gameRecord := gameTotals[i]
			if !gameRecord.IsValidRound {
				newGameTotals = append(newGameTotals, gameRecord)
			}
		}

		// 切片重新赋值
		gameTotals = make([]models.RecordGameTotal, len(newGameTotals))
		copy(gameTotals, newGameTotals)
	}

	for i := 0; i < len(gameTotals); i++ {
		gameRecord := gameTotals[i]

		// 成员战绩--详情--大赢家详情
		if bw && !gameRecord.IsValidRound {
			continue
		}

		// 筛选点赞信息
		if len(likeFlagMap) > 0 {
			if likeFlag == 1 {
				if !likeFlagMap[gameRecord.GameNum] {
					continue
				}
			} else if likeFlag == 2 {
				if likeFlagMap[gameRecord.GameNum] {
					continue
				}
			}
		} else {
			if likeFlag == 1 {
				continue
			}
		}

		// 房间是否已解散
		if gameRecord.ScoreKind == static.ScoreKind_pass && gameRecord.PlayCount == 1 && GetTableMgr().GetTable(gameRecord.RoomNum) == nil {
			continue
		}

		// 是否已保存在战绩map中
		_, ok := gameTotalMap[gameRecord.GameNum]
		if !ok {
			gameTotalMap[gameRecord.GameNum] = []models.RecordGameTotal{}
		}
		// 更新战绩map
		gameTotalMap[gameRecord.GameNum] = append(gameTotalMap[gameRecord.GameNum], gameRecord)
	}

	// 已退圈的成员记录
	leaveMemMap := make(map[int64]bool)
	if uid == 0 && puid > 0 {
		leaveMemMap, _ = GetDBMgr().SelectLeaveHousePartnerMember(hid, puid, selectDate1)
	}

	// 筛选记录
	CheckRecordCondition := func(reqUid int64, floorIndex int, gameRecords []models.RecordGameTotal, mpuid int64, mMap map[int64]HouseMember) bool {
		for _, gameRecord := range gameRecords {
			if uid > 0 {
				if gameRecord.Uid == uid {
					if floorIndex != -1 {
						if gameRecord.DFId == floorIndex {
							return true
						}
					} else {
						return true
					}
				}
			} else {
				if mpuid > 0 {
					mem, ok := mMap[gameRecord.Uid]
					if ok {
						if mem.Partner == mpuid || mem.UId == mpuid {
							if floorIndex != -1 {
								if gameRecord.DFId == floorIndex {
									return true
								}
							} else {
								return true
							}
						}
					} else {
						if _, ok := leaveMemMap[gameRecord.Uid]; ok {
							return true
						}
					}
				} else {
					if floorIndex != -1 {
						if gameRecord.DFId == floorIndex {
							return true
						}
					} else {
						return true
					}
				}
			}
		}
		return false
	}

	// 筛选后的战绩map
	searchGameMap := make(map[string][]models.RecordGameTotal)
	// 遍历我的包厢以及楼层战绩信息
	for gameNum, gameRecords := range gameTotalMap {
		if CheckRecordCondition(uid, floorIndex, gameRecords, puid, memMap) {
			searchGameMap[gameNum] = gameRecords
		}
	}

	// 查询汇总的所有记录
	var searchGameTotals []models.RecordGameTotal

	// 遍历搜索
	if searchkey != "" {
		for _, searchRecords := range searchGameMap {
			for _, gameRecord := range searchRecords {
				if static.HF_Itoa(gameRecord.RoomNum) == searchkey || static.HF_I64toa(gameRecord.Uid) == searchkey {
					if bw {
						if static.HF_I64toa(gameRecord.Uid) == searchkey && gameRecord.WinScore >= bwUserScoreMap[gameRecord.GameNum] {
							searchGameTotals = append(searchGameTotals, searchRecords...)
						}
					} else {
						searchGameTotals = append(searchGameTotals, searchRecords...)
					}
					break
				}
			}
		}
	} else {
		for _, searchRecords := range searchGameMap {
			for _, gameRecord := range searchRecords {
				if bw {
					if gameRecord.Uid == uid && gameRecord.WinScore >= bwUserScoreMap[gameRecord.GameNum] && bwUserScoreMap[gameRecord.GameNum] > 0 {
						searchGameTotals = append(searchGameTotals, searchRecords...)
						break
					}
				} else {
					searchGameTotals = append(searchGameTotals, searchRecords...)
					break
				}
			}
		}
	}

	// 整理并汇总数据
	//////////////////////////////////////////////////////////////////////////////////
	gameRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound := self.TransformTotalToDetail(searchGameTotals, bwUserScoreMap, uid, puid, memMap, leaveMemMap, likeFlagMap)

	return gameRecords, curTotalRound, curCompleteRound, curDismissRound, curValidRound, curInValidRound, curUserScore, curBwTimes, curPlayTimes, curInValidTimes, curTotalLikeRound, nil
}

func (self *DBMgr) GetUserAgentConfig(uid int64) (*models.UserAgentConfig, *xerrors.XError) {
	var userAgentConfig models.UserAgentConfig
	err := self.GetDBmControl().First(&userAgentConfig, "uid = ? and state = ?", uid, 1).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			xlog.Logger().Warnf("玩家%d 没有配置后台", uid)
			return nil, xerrors.UserAgentNotConfigError
		} else {
			xlog.Logger().Error(err)
			return nil, xerrors.DBExecError
		}
	}
	return &userAgentConfig, nil
}

// 包厢排行榜
func (self *DBMgr) SelectHouseMemberRank(Hid int64, isRange int, selectTime1 time.Time, selectTime2 time.Time, rankType int, begin int, end int) (map[int64]models.QueryHouseRankResult, error) {
	selectDate1 := fmt.Sprintf("%d-%02d-%02d", selectTime1.Year(), selectTime1.Month(), selectTime1.Day())
	selectDate2 := fmt.Sprintf("%d-%02d-%02d", selectTime2.Year(), selectTime2.Month(), selectTime2.Day())
	xlog.Logger().Println("开始时间：", selectDate1)
	xlog.Logger().Println("结束时间：", selectDate2)
	var result []models.QueryHouseRankResult
	resultMap := make(map[int64]models.QueryHouseRankResult)
	var err error
	sql2 := ""
	if rankType == static.RANK_TYPE_ROUND {
		sql2 = "ORDER BY rank_round DESC limit "
	} else if rankType == static.RANK_TYPE_WINER {
		sql2 = "ORDER BY rank_winer DESC limit "
	} else if rankType == static.RANK_TYPE_RECORD {
		sql2 = "ORDER BY rank_record DESC limit "
	}
	sql2 += fmt.Sprintf("%d,%d", begin, end)
	if isRange < static.RANK_TIME_CARVE {
		sql := `select uid, rank_round, rank_winer, rank_record from house_rank where dhid = ? and date_format(created_at, '%Y-%m-%d') = ? `
		sql += sql2
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1).Scan(&result).Error
	} else {
		//sql := `select uid, sum(rank_round) as RankRound, sum(rank_winer) as RankWiner, sum(rank_record) as RankRecord from house_rank where dhid = ? and created_at >= ? and created_at < ? `
		sql := `select uid, sum(rank_round) as rank_round, sum(rank_winer) as rank_winer, sum(rank_record) as rank_record from house_rank where dhid = ? and date_format(created_at, '%Y-%m-%d') >= ? and date_format(created_at, '%Y-%m-%d') < ? group by uid `
		sql += sql2
		err = GetDBMgr().GetDBmControl().Raw(sql, Hid, selectDate1, selectDate2).Scan(&result).Error
	}

	if err != nil {
		return resultMap, err
	}
	for _, item := range result {
		resultMap[item.Uid] = item
	}
	xlog.Logger().Println("resultMap   ", resultMap)
	return resultMap, nil
}

// 包厢区域列表
type HouseArea struct {
	Area string // 区域码
	Cnt  int    // 累计出现次数
	Time int64  // 最近的一次入圈时间
}

type HouseAreaList []*HouseArea

func (h HouseAreaList) Len() int {
	return len(h)
}

func (h HouseAreaList) Less(i, j int) bool {
	if h[i].Cnt == h[j].Cnt {
		return h[i].Time > h[j].Time
	}
	return h[i].Cnt > h[j].Cnt
}

func (h HouseAreaList) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (self *DBMgr) GetHouseAreaList(uid int64) HouseAreaList {
	houseAreaList := make(HouseAreaList, 0)

	lhIds, err := self.ListHouseIdMemIn(uid)
	if err != nil {
		return houseAreaList
	}

	for _, hid := range lhIds {
		// 包厢数据
		dhouse, err := self.GetDBrControl().GetHouseInfoById(hid)
		if err != nil || dhouse == nil {
			continue
		}
		// 玩家数据
		dmem, err := self.GetDBrControl().HouseMemberQueryById(dhouse.Id, dhouse.UId)
		if err != nil || dmem == nil {
			continue
		}

		house := GetClubMgr().GetClubHouseByHId(dhouse.HId)
		if house == nil {
			xlog.Logger().Errorln("user:", uid, " | house_", hid, " not exists")
			continue
		}

		houseArea := &HouseArea{
			Area: dhouse.Area,
			Cnt:  1,
			Time: dmem.AgreeTime,
		}

		// 是否已存在
		key := -1
		for k, v := range houseAreaList {
			if v.Area == houseArea.Area {
				key = k
				break
			}
		}

		if key > 0 {
			houseAreaList[key].Cnt++
			if houseArea.Time > houseAreaList[key].Time {
				houseAreaList[key].Time = houseArea.Time
			}
		} else {
			houseAreaList = append(houseAreaList, houseArea)
		}
	}

	// 排序
	sort.Sort(houseAreaList)

	return houseAreaList
}

type RewardDetail struct {
	Total   int64
	Current int64
	Cleared int64
}

// ! 获得队长收益
func (self *DBMgr) SelectHouseAllPartnerReward(dhid int64, dfid int64, startAt *time.Time, endAt *time.Time,
	pid ...int64) (map[int64]RewardDetail, []int64 /*pk*/, []int64 /*可过期pk*/, error) {

	db := GetDBMgr().GetDBmControl().Model(&models.PartnerRewardT{})
	db = db.Where("dhid = ?", dhid)
	if dfid > 0 {
		db = db.Where("dfid = ?", dfid)
	}

	if len(pid) > 0 {
		db = db.Where("partner in(?)", pid)
	}

	if startAt != nil {
		db = db.Where("created_time >= ?", startAt.Unix())
	}

	if endAt != nil {
		db = db.Where("created_time <= ?", endAt.Unix())
	}

	rewards := make([]*models.PartnerRewardT, 0)
	err := db.Find(&rewards).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return nil, nil, nil, err
	}

	timeNowUnix := time.Now().Unix()
	result := make(map[int64]RewardDetail)
	clearPk := make([]int64, 0)
	deletePk := make([]int64, 0)
	for _, reward := range rewards {
		detail, ok := result[reward.Partner]
		if !ok {
			detail = RewardDetail{}
		}
		detail.Total += reward.Reward
		if reward.ClearedTime > 0 {
			detail.Cleared += reward.Reward
		} else {
			clearPk = append(clearPk, reward.Id)
			detail.Current += reward.Reward
		}
		// 超过8天，认为可以清除，如果划扣，由调用者选择清除
		if reward.CreatedTime-timeNowUnix >= 60*60*24*8 {
			deletePk = append(deletePk, reward.Id)
		}
		result[reward.Partner] = detail
	}
	return result, clearPk, deletePk, nil
}

// ! 查询单个队长某局收益
func (self *DBMgr) SelectHousePartnerRewardWithGameNums(dhid int64, pid int64, gameNums []string) (map[string]int64, error) {
	rewards := make([]*models.PartnerRewardT, 0)
	err := GetDBMgr().GetDBmControl().Find(&rewards, "dhid = ? AND partner = ? AND game_num in(?)",
		dhid, pid, gameNums,
	).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return nil, err
	}
	result := make(map[string]int64)
	for _, reward := range rewards {
		result[reward.GameNum] += reward.Reward
	}
	return result, nil
}

type HouseMemberTaxLog struct {
	Value   int64  `gorm:"column:value"`
	GameNum string `gorm:"column:game_num"`
}

// ! 查询单个队长某局收益
func (self *DBMgr) SelectHouseMemberPayWithGameNums(dhid int64, mem int64, gameNums []string) (map[string]int64, error) {
	result := make([]*HouseMemberTaxLog, 0)
	sql := "select value, game_num from house_vitamin_pool_tax_log where hid =? and opuid = ? and game_num in(?) and optype in(?)"
	err := GetDBMgr().GetDBmControl().Raw(sql, dhid, mem, gameNums,
		[]int64{int64(models.BigWinCost), int64(models.GamePay)}).Scan(&result).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		xlog.Logger().Error(err)
		return nil, err
	}
	ret := make(map[string]int64)
	for _, res := range result {
		ret[res.GameNum] -= res.Value
	}
	return ret, nil
}
