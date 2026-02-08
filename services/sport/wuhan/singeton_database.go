// ! 数据库底层
package wuhan

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/dao"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/open-source/game/chess.git/pkg/models"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// //////////////////////////////////////////////////////////////////////////////
// ! 数据结构
type DBMgr struct {
	db_R       *dao.DB_r      //! redis操作
	db_M       *dao.ORM_Mysql //! mysql操作
	Redis      *redis.Client
	PubRedis   *redis.Client
	db_R_Store *dao.DB_r //! 存储redis操作
	StoreRedis *redis.Client
	RedisLock  *redis.Client //分布式锁的redis库
}

var DBMgrSingleton *DBMgr = nil

// ! 得到包厢管理单例
func GetDBMgr() *DBMgr {

	if DBMgrSingleton == nil {

		DBMgrSingleton = new(DBMgr)
		DBMgrSingleton.db_R = new(dao.DB_r)
		DBMgrSingleton.db_R_Store = new(dao.DB_r)

		con := GetServer().Con
		//! redis
		err := DBMgrSingleton.db_R.Init(con.Redis, con.RedisDB, con.RedisAuth)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("Init redis error: %s", err.Error()))
			return nil

		}
		//! redis
		err = DBMgrSingleton.db_R_Store.Init(con.StoreRedis, con.StoreRedisDB, con.StoreRedisAuth)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("Init Storeredis error: %s", err.Error()))
			return nil
		}
		// 初始化v2版本redis库
		DBMgrSingleton.Redis = static.InitRedisV2(con.Redis, con.RedisDB, con.RedisAuth)
		DBMgrSingleton.PubRedis = static.InitRedisV2(con.PubRedis, con.PubRedisDB, con.PubRedisAuth)
		DBMgrSingleton.RedisLock = static.InitRedisV2(con.Redis, con.RedisDB+1, con.RedisAuth)
		DBMgrSingleton.StoreRedis = static.InitRedisV2(con.StoreRedis, con.StoreRedisDB, con.StoreRedisAuth)
		//! mysql
		DBMgrSingleton.db_M = new(dao.ORM_Mysql)
		_, err = DBMgrSingleton.db_M.Open("mysql", con.DB)
		if err != nil {
			xlog.Logger().Errorln(err)
			panic(fmt.Sprintf("Init database error: %s", err.Error()))
			return nil
		}

		//! basedata
		err = DBMgrSingleton.DataInitialize()
		if err != nil {
			xlog.Logger().Errorln(err)
			return nil
		}

		//！数据库配置
		if err = DBMgrSingleton.ReadAllConfig(); err != nil {
			xlog.Logger().Errorln("DBMgrSingleton.ReadAllConfig error:", err.Error())
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

func (self *DBMgr) GetDBrsControl() *dao.DB_r {
	return self.db_R_Store
}

func (self *DBMgr) Close() error {

	if self.db_R != nil {
		self.db_R.Close()
	}
	if self.db_R_Store != nil {
		self.db_R_Store.Close()
	}

	if self.db_M != nil {
		err := self.db_M.GetConn().Close()
		if err != nil {
			log.Panic("dbmap close error. ", err)
			return err
		}
	}

	return nil
}

func (self *DBMgr) DataInitialize() error {
	// 初始化 依赖 各功能控制器 DB->redis->内存

	return nil
}

// ! 插入单局战绩数据
func (self *DBMgr) InsertGameRecord(gameRecord *models.RecordGameRound) (int64, error) {
	//mysql
	if err := self.db_M.Create(gameRecord).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("gamenum（%s）PlayNum(%d) 写入单局结算出错：%v", gameRecord.GameNum, gameRecord.PlayNum, err))
		return 0, err
	}
	return gameRecord.Id, nil
}

// ! 插入单局战绩回放数据
func (self *DBMgr) InsertGameRecordReplay(gameRecordReplay *models.RecordGameReplay) (int64, error) {
	var err error
	//mysql
	if err = self.db_M.Create(gameRecordReplay).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("gamenum（%s）PlayNum(%d) 写入单局回放出错：%v", gameRecordReplay.GameNum, gameRecordReplay.PlayNum, err))
		return 0, err
	}

	return gameRecordReplay.Id, nil
}

// ! 插入单局战绩回放数据
func (self *DBMgr) UpdataGameRecordReplay(gameRecordReplay *models.RecordGameReplay) (int64, error) {
	var err error
	if gameRecordReplay.Id <= 0 {
		return 0, errors.New("回放id必须大于0")
	}
	//mysql
	if err = self.db_M.Model(models.RecordGameReplay{Id: gameRecordReplay.Id}).UpdateColumn("outcard", gameRecordReplay.OutCard).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("id（%d） 更新单局回放出错：%v", gameRecordReplay.Id, err))
		return 0, err
	}
	if err = self.db_M.Model(models.RecordGameReplay{Id: gameRecordReplay.Id}).UpdateColumn("end_info", gameRecordReplay.EndInfo).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("id（%d） 更新单局回放小结算记录：%v", gameRecordReplay.Id, err))
		return 0, err
	}
	return gameRecordReplay.Id, nil
}

// ! 查询包厢总战绩数据
func (self *DBMgr) SelectGameRecordTotal(gamenum string, uid int64) (models.RecordGameTotal, error) {
	var recordgametotal models.RecordGameTotal

	err := self.db_M.Model(models.RecordGameTotal{}).Where("game_num = ? and uid = ?", gamenum, uid).First(&recordgametotal).Error

	if err != nil {
		return models.RecordGameTotal{}, err
	}

	return recordgametotal, err
}

// ! 插入包厢总战绩数据
func (self *DBMgr) InsertGameRecordTotal(gameRecordTotal *models.RecordGameTotal) (int64, error) {
	// mysql
	if err := self.db_M.Create(gameRecordTotal).Error; err != nil {
		return 0, err
	}

	return gameRecordTotal.Id, nil
}

// ! 更新总战绩数据
func (self *DBMgr) UpdataGameRecordTotal(Id int64, UpdataMap map[string]interface{}) {
	//mysql
	if err := GetDBMgr().GetDBmControl().Model(models.RecordGameTotal{Id: Id}).Updates(UpdataMap).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("id (%d) 更新总结算出错：%v", Id, err))
	}
}

// ! 更新总战绩数据
func (self *DBMgr) UpdateGameRecordTotalByGameNum(gameNum string, uid int64, updateMap map[string]interface{}) {
	// mysql
	if err := GetDBMgr().GetDBmControl().Model(models.RecordGameTotal{}).Where("game_num = ? and uid = ?", gameNum, uid).Updates(updateMap).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("gamenum(%s) uid (%d) 更新总结算出错：%v", gameNum, uid, err))
	}
}

// ! 更新玩家胜率表
func (self *DBMgr) UpdataUserRateOfWinning(uid int64, score int) {
	UpdataMap := make(map[string]interface{})
	if score > 0 {
		UpdataMap["win_count"] = gorm.Expr("win_count + 1")
		UpdataMap["total_count"] = gorm.Expr("total_count + 1")
	} else if score < 0 {
		UpdataMap["lost_count"] = gorm.Expr("lost_count + 1")
		UpdataMap["total_count"] = gorm.Expr("total_count + 1")
	} else {
		UpdataMap["draw_count"] = gorm.Expr("draw_count + 1")
		UpdataMap["total_count"] = gorm.Expr("total_count + 1")
	}

	if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", uid).Updates(UpdataMap).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("id (%d) 更新玩家胜率出错：%v", uid, err))
	}
}

// ! 查询包厢每日统计数据
func (self *DBMgr) SelectUserGameDayRecordToDaySet(hid int, fid int, dfid int, uid int64, selecttime time.Time) ([]models.RecordGameDay, error) {
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var dayrecords []models.RecordGameDay = []models.RecordGameDay{}
	err := self.GetDBmControl().Model(models.RecordGameDay{}).Where("hid = ? and fid = ? and dfid = ? and uid = ? and date_format(created_at, '%Y-%m-%d') = ?", hid, fid, dfid, uid, selectstr).Find(&dayrecords).Error
	return dayrecords, err
}

// ! 查询包厢每日统计数据
func (self *DBMgr) SelectUserGameDayRecordToDaySetWithP(hid int, fid int, dfid int, uid int64, partner int64, superiorId int64, randix int, selecttime time.Time) ([]models.RecordGameDay, error) {
	selectstr := fmt.Sprintf("%d-%02d-%02d", selecttime.Year(), selecttime.Month(), selecttime.Day())
	var dayrecords []models.RecordGameDay = []models.RecordGameDay{}
	err := self.GetDBmControl().Model(models.RecordGameDay{}).Where("hid = ? and fid = ? and dfid = ? and uid = ? and partner = ? and superiorid = ? and radix = ? and date_format(created_at, '%Y-%m-%d') = ?",
		hid, fid, dfid, uid, partner, superiorId, randix, selectstr).Find(&dayrecords).Error
	return dayrecords, err
}

// ! 更新包厢每日统计数据
func (self *DBMgr) HouseUpdataGameDayRecord(DHId int64, FId int64, DFId int, UId int64, PlayTimes int, BwTimes int, TotalScore int, ValidTimes int, BigValidTimes int, Partner int64, SuperiorId int64, randix int) error {
	db_dayRecord, _ := self.SelectUserGameDayRecordToDaySetWithP(int(DHId), int(FId), int(DFId), UId, Partner, SuperiorId, randix, time.Now())
	if db_dayRecord != nil && len(db_dayRecord) >= 1 {
		db_dayRecord[0].PlayTimes += PlayTimes
		db_dayRecord[0].BwTimes += BwTimes
		db_dayRecord[0].TotalScore += TotalScore
		db_dayRecord[0].ValidTimes += ValidTimes
		db_dayRecord[0].BigValidTimes += BigValidTimes
		updateMap := make(map[string]interface{})
		updateMap["play_times"] = db_dayRecord[0].PlayTimes
		updateMap["bw_times"] = db_dayRecord[0].BwTimes
		updateMap["total_score"] = db_dayRecord[0].TotalScore
		updateMap["valid_times"] = db_dayRecord[0].ValidTimes
		updateMap["big_valid_times"] = db_dayRecord[0].BigValidTimes
		//mysql
		if err := self.UpdataGameDayRecord(db_dayRecord[0].Id, updateMap); err != nil {
			return err
		}
	} else if len(db_dayRecord) == 0 {
		newRecord := new(models.RecordGameDay)
		newRecord.DHId = DHId
		newRecord.FId = FId
		newRecord.DFId = int64(DFId)
		newRecord.UId = UId
		newRecord.PlayDate = time.Now()
		newRecord.PlayTimes = PlayTimes
		newRecord.BwTimes = BwTimes
		newRecord.TotalScore = TotalScore
		newRecord.ValidTimes = ValidTimes
		newRecord.BigValidTimes = BigValidTimes
		newRecord.CreatedAt = time.Now()
		newRecord.Partner = Partner
		newRecord.SuperiorId = SuperiorId
		newRecord.Radix = randix
		//mysql
		if err := self.InsertGameDayRecord(newRecord); err != nil {
			return err
		}
	}

	return nil
}

// ! 插入包厢每日统计数据
func (self *DBMgr) InsertGameDayRecord(newRecord *models.RecordGameDay) error {
	//mysql
	var err error
	if err = self.db_M.Create(newRecord).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("hid（%d) uid(%d) 写入每日统计表出错：%v", newRecord.DHId, newRecord.UId, err))
		return err
	}

	return err
}

// ! 更新包厢每日统计数据
func (self *DBMgr) UpdataGameDayRecord(Id int, UpdataMap map[string]interface{}) error {
	//mysql
	var err error
	if err = self.db_M.Model(models.RecordGameDay{Id: Id}).Updates(UpdataMap).Error; err != nil {
		xlog.Logger().Errorln(fmt.Sprintf("id（%d) 更新每日统计表出错：%v", Id, err))
	}

	return err
}

// ! 更新包厢玩家总大赢家以及对局次数
func (self *DBMgr) UpdataGameHouseMemberRecord(DHId int64, UId int64, PlayTimes int, BwTimes int) error {

	//db_houseMember, err := self.HouseMemberQueryById(DHId, UId)
	db_houseMember, err := GetDBMgr().GetDBrControl().HouseMemberQueryById(DHId, UId)
	if err != nil {
		return err
	}
	db_houseMember.BwTimes += BwTimes
	db_houseMember.PlayTimes += PlayTimes

	updateMap := make(map[string]interface{})
	updateMap["play_times"] = db_houseMember.PlayTimes
	updateMap["bw_times"] = db_houseMember.BwTimes

	//mysql
	if err = GetDBMgr().GetDBmControl().Model(models.HouseMember{}).Where("hid = ? and uid = ?", db_houseMember.DHId, db_houseMember.UId).Updates(updateMap).Error; err != nil {
		return err
	}

	//redis
	err = GetDBMgr().GetDBrControl().HouseMemberInsert(db_houseMember)
	//
	//err = GetDBMgr().GetDBrControl().HouseRecordPlayInsert(int64(DHId), db_houseMember.UId, db_houseMember.PlayTimes)
	//err = GetDBMgr().GetDBrControl().HouseRecordBWInsert(int64(DHId), db_houseMember.UId, db_houseMember.BwTimes)

	return err
}

// ! 更新排位赛统计数据
func (self *DBMgr) UpdataGameMatchTotal(gameMatchTotal *models.GameMatchTotal) (int64, error) {

	db_gameMatchTotal, err := self.db_R.SelectGameMatchTotalSet(gameMatchTotal)
	if gameMatchTotal.TotalCount == 0 && gameMatchTotal.WinCount == 0 {
		//total里面不需要记录不参与结算的信息，比如逃跑局的所有玩家都不算
		return 0, nil
	}
	honorAwards := GetServer().GetMatchHonorAwardConfig(gameMatchTotal, gameMatchTotal.SiteType)
	if db_gameMatchTotal != nil {
		db_gameMatchTotal.Score += gameMatchTotal.Score
		if gameMatchTotal.TotalCount > 0 {
			db_gameMatchTotal.TotalCount++
		}
		if gameMatchTotal.WinCount > 0 {
			db_gameMatchTotal.WinCount++
		}

		updateMap := make(map[string]interface{})
		updateMap["score"] = db_gameMatchTotal.Score
		updateMap["wincount"] = db_gameMatchTotal.WinCount
		updateMap["totalcount"] = db_gameMatchTotal.TotalCount
		updateMap["updated_at"] = db_gameMatchTotal.CreatedAt
		if honorAwards != nil {
			db_gameMatchTotal.Coupon += honorAwards.WinHonor*gameMatchTotal.WinCount + honorAwards.JionHonor*gameMatchTotal.TotalCount
			updateMap["coupon"] = db_gameMatchTotal.Coupon
		} else {
			//没有配置奖励列表时，就用参与次数为勋章数
			db_gameMatchTotal.Coupon += gameMatchTotal.TotalCount
			updateMap["coupon"] = db_gameMatchTotal.Coupon
		}
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
		newRecord.MatchId = gameMatchTotal.MatchId
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
	err = self.db_R.InsertGameMatchTotalSet(db_gameMatchTotal)
	if err != nil {
		return 0, nil
	}

	return 0, nil
}

// ! 更新排位赛统计数据
func (self *DBMgr) updataGameMatchState(matchconfig *models.ConfigMatch, state int) (int64, error) {

	updateMap := make(map[string]interface{})
	updateMap["State"] = state

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

// ! 获取包厢有效最低分数
func (self *DBMgr) SelectHouseValidRound(dhid int64, fid int64) (int, int, error) {
	var housefloorvalidrounds []models.HouseFloorValidRound
	err := self.db_M.Model(models.HouseFloorValidRound{}).Where("hid = ? and fid = ?", dhid, fid).Find(&housefloorvalidrounds).Error
	if err != nil {
		xlog.Logger().Errorln("Select HouseFloorValidRound error: ", err.Error())
		return 0, -1, err
	}

	if len(housefloorvalidrounds) > 0 {
		return housefloorvalidrounds[0].ValidMinScore, housefloorvalidrounds[0].ValidBigScore, err
	}

	var housevalidrounds []models.HouseValidRound
	err = self.db_M.Model(models.HouseValidRound{}).Where("hid = ?", dhid).Find(&housevalidrounds).Error
	if err != nil {
		xlog.Logger().Errorln("Select HouseValidRound error: ", err.Error())
		return 0, -1, err
	}
	if len(housevalidrounds) == 0 {
		return 0, -1, err
	}
	return housevalidrounds[0].ValidMinScore, -1, err
}

// ! 插入每天对局疲劳值消耗(playcost, bwcost对局扣除疲劳值,AA以及大赢家)
func (self *DBMgr) InsertVitaminCost(dhid int64, fid int64, dfid int64, uid int64, playcost int64, bwcost int64, winlose int64, partner int64) error {
	var housevitaminrecord models.RecordVitaminDay

	if partner == 1 {
		partner = uid
	}

	db := self.db_M.Model(models.RecordVitaminDay{})

	db = db.Where("hid = ? and fid = ? and dfid = ? and uid = ? and partner = ?", dhid, fid, dfid, uid, partner)

	//查找今天的
	nowtime := time.Now()
	selectstr := fmt.Sprintf("%d-%02d-%02d", nowtime.Year(), nowtime.Month(), nowtime.Day())
	db = db.Where(" date_format(created_at, '%Y-%m-%d') = ?", selectstr)

	err := db.First(&housevitaminrecord).Error

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}

		housevitaminrecord.DHId = dhid
		housevitaminrecord.FId = fid
		housevitaminrecord.DFId = dfid
		housevitaminrecord.UId = uid
		housevitaminrecord.VitaminCost = playcost + bwcost
		housevitaminrecord.VitaminCostRound = playcost
		housevitaminrecord.VitaminCostBW = bwcost
		housevitaminrecord.VitaminWinLose = winlose
		housevitaminrecord.CreatedAt = nowtime
		housevitaminrecord.UpdatedAt = nowtime
		housevitaminrecord.Partner = partner
		err = self.db_M.Create(&housevitaminrecord).Error

		return err
	}

	housevitaminrecord.VitaminCost += playcost + bwcost
	housevitaminrecord.VitaminCostRound += playcost
	housevitaminrecord.VitaminCostBW += bwcost
	housevitaminrecord.VitaminWinLose += winlose

	updateMap := make(map[string]interface{})
	updateMap["vitamincost"] = housevitaminrecord.VitaminCost
	updateMap["vitamincostround"] = housevitaminrecord.VitaminCostRound
	updateMap["vitamincostbw"] = housevitaminrecord.VitaminCostBW
	updateMap["vitaminwinlose"] = housevitaminrecord.VitaminWinLose
	//mysql
	err = self.db_M.Model(models.RecordVitaminDay{Id: housevitaminrecord.Id}).Updates(updateMap).Error

	return err
}

// ! 插入每天对局疲劳值扎帐记录
func (self *DBMgr) InsertVitaminCostClear(dhid int64, playcost int64, bwcost int64) error {
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
		housevitaminrecordclear.VitaminCost = playcost + bwcost
		housevitaminrecordclear.VitaminCostRound = playcost
		housevitaminrecordclear.VitaminCostBW = bwcost

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

	housevitaminrecordclear.VitaminCost += playcost + bwcost
	housevitaminrecordclear.VitaminCostRound += playcost
	housevitaminrecordclear.VitaminCostBW += bwcost

	updateMap := make(map[string]interface{})
	updateMap["vitamincost"] = housevitaminrecordclear.VitaminCost
	updateMap["vitamincostround"] = housevitaminrecordclear.VitaminCostRound
	updateMap["vitamincostbw"] = housevitaminrecordclear.VitaminCostBW

	//mysql
	err = self.db_M.Model(models.RecordVitaminDayClear{Id: housevitaminrecordclear.Id}).Updates(updateMap).Error

	return err
}

// ! 插入记账疲劳值消号(winlosecost输赢扣除疲劳值, playcost, bwcost对局扣除疲劳值,AA以及大赢家)
func (self *DBMgr) InsertVitaminCostFromLastNode(dhid int64, uid int64, before, after, offset, playcost, bwcost int64) error {
	//var housevitaminmgr model.RecordVitaminMgrList
	//
	//db := self.db_M.Model(model.RecordVitaminMgrList{}).Where("hid = ? and uid = ? and recording = 1", dhid, uid)
	//
	//err := db.First(&housevitaminmgr).Error
	//if err != nil {
	//	if err != gorm.ErrRecordNotFound {
	//		return err
	//	}
	//
	//	housevitaminmgr.DHId = dhid
	//	housevitaminmgr.UId = uid
	//	housevitaminmgr.Recording = 1
	//
	//	housevitaminmgr.PreNodeVitamin = 0
	//	housevitaminmgr.VitaminWinLoseCost = offset
	//	housevitaminmgr.VitaminPlayCost = playcost + bwcost
	//	housevitaminmgr.VitaminCostRound = playcost
	//	housevitaminmgr.VitaminCostBW = bwcost
	//	housevitaminmgr.CreatedAt = time.Now()
	//
	//	err = self.db_M.Create(&housevitaminmgr).Error
	//
	//	return err
	//}
	//
	//housevitaminmgr.VitaminWinLoseCost += offset
	//housevitaminmgr.VitaminPlayCost += playcost + bwcost
	//housevitaminmgr.VitaminCostRound += playcost
	//housevitaminmgr.VitaminCostBW += bwcost
	//
	//updateMap := make(map[string]interface{})
	//updateMap["vitaminwinlosecost"] = housevitaminmgr.VitaminWinLoseCost
	//updateMap["vitaminplaycost"] = housevitaminmgr.VitaminPlayCost
	//updateMap["vitamincostround"] = housevitaminmgr.VitaminCostRound
	//updateMap["vitamincostbw"] = housevitaminmgr.VitaminCostBW
	//
	////mysql
	//err = self.db_M.Model(model.RecordVitaminMgrList{Id: housevitaminmgr.Id}).Updates(updateMap).Error
	//
	//if err != nil {
	//	return err
	//}

	// 记录用户财富消耗流水
	record := new(models.UserWealthCost)
	record.Uid = uid
	record.WealthType = models.CostTypeGameVitamin
	record.CostType = models.CostTypeGame
	record.Cost = offset
	record.BeforeNum = before
	record.AfterNum = after
	if err := self.db_M.Create(&record).Error; err != nil {
		xlog.Logger().Errorln(consts.MsgTypeTableRes_Ntf, err.Error())
		return err
	}
	return nil
}

// ！从数据库读取游戏配置
func (self *DBMgr) ReadAllConfig() error {
	var err error

	// 读取服务器配置
	if err := self.GetDBmControl().First(&GetServer().ConServers).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 读取包厢配置
	if err := self.GetDBmControl().First(&GetServer().ConHouse).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 读取房间配置
	var siteConfigs []*models.ConfigSite
	xlog.Logger().Infoln("read config_site....")
	if err = self.GetDBmControl().Find(&siteConfigs).Error; err != nil {
		return err
	}
	GetServer().ConSite = siteConfigs

	// 进程对应的kindid读取
	var cs []*models.ConfigGameTypes
	if err = self.GetDBmControl().Where("game_server_id = ?", GetServer().Index).Find(&cs).Error; err != nil {
		return err
	}
	gameTypes := make(map[int][]*static.ServerGameType)
	for _, item := range cs {
		xlog.Logger().Infoln(item)

		gameType := new(static.ServerGameType)
		gameType.GameType = item.GameType
		gameType.KindId = item.KindId
		gameType.SiteType = item.SiteType
		gameType.TableNum = item.TableNum
		gameType.MaxPeopleNum = item.MaxPeopleNum

		var arr []*static.ServerGameType
		var ok bool

		if arr, ok = gameTypes[item.KindId]; !ok {
			arr = make([]*static.ServerGameType, 0)
		}
		arr = append(arr, gameType)
		gameTypes[item.KindId] = arr
	}

	GetServer().GameTypes = gameTypes
	xlog.Logger().Infoln("GetServer().GameTypes:", GetServer().GameTypes)

	isInArray := func(kindid int) bool {
		_, ok := gameTypes[kindid]
		return ok
	}

	var gs []*models.GameConfig
	if err = self.GetDBmControl().Find(&gs).Error; err != nil {
		return err
	}

	for i := 0; i < len(gs); {
		if isInArray(gs[i].KindId) {
			i++
		} else {
			copy(gs[i:], gs[i+1:])
			gs = gs[:len(gs)-1]
		}
	}
	GetServer().ConGame = gs
	xlog.Logger().Infoln(GetServer().ConGame)

	// 读取房间排位赛的配置
	var matchConfigs []*models.ConfigMatch
	xlog.Logger().Infoln("read config_match....")
	if err = self.GetDBmControl().Find(&matchConfigs).Error; err != nil {
		return err
	}
	GetServer().ConMatch = matchConfigs

	// 读取排位赛奖励清单配置
	xlog.Logger().Infoln("read config_matchaward....")
	if err = self.GetDBmControl().Find(&GetServer().ConMatchAward).Error; err != nil {
		return err
	}

	xlog.Logger().Infoln("read Config game control....")

	if err = self.GetDBmControl().Find(&GetServer().ConGameControl).Error; err != nil {
		return err
	}

	if err = self.GetDBmControl().First(GetServer().ConSpinBase).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// //PoolChange 包厢疲劳值变更记录
// func (self *DBMgr) PoolChange(dhid int64, optUid int64, optType model.VitaminChangeType, value float64, tx *gorm.DB) error {
// 	sql := `select after from house_vitamin_pool where hid = ? and after >= ? for update `
// 	sql1 := `insert house_vitamin_pool(hid,after) values(?,?) ON DUPLICATE KEY UPDATE after = after + ? `
// 	sql2 := `insert into %s(hid,opuid,value,optype,after)
// 	values(?,?,?,?,(select after from house_vitamin_pool where hid = ? order by id desc limit 1))`
// 	if optType == model.BigWinCost || optType == model.GamePay {
// 		sql2 = fmt.Sprintf(sql2, "house_vitamin_pool_tax_log")
// 	} else {
// 		sql2 = fmt.Sprintf(sql2, "house_vitamin_pool_log")
// 	}
// 	if tx == nil {
// 		db := GetDBMgr().GetDBmControl().Begin()
// 		if value < 0 {
// 			if db.Exec(sql, dhid, value).RowsAffected != 1 {
// 				return errors.New("vitamin not enough")
// 			}
// 		}
// 		err := tx.Exec(sql1, dhid, value, value).Error
// 		if err != nil {
// 			db.Rollback()
// 			syslog.Logger().Errorf("%v", err)
// 			return err
// 		}
// 		err = db.Exec(sql2, dhid, optUid, value, optType, dhid).Error
// 		if err != nil {
// 			db.Rollback()
// 			syslog.Logger().Errorf("%v", err)
// 			return err
// 		}
// 		err = db.Commit().Error
// 		if err != nil {
// 			db.Rollback()
// 			syslog.Logger().Errorf("%v", err)
// 			return err
// 		}
// 	} else {
// 		err := tx.Exec(sql1, dhid, value, value).Error
// 		if err != nil {
// 			syslog.Logger().Errorf("%v", err)
// 			return err
// 		}
// 		err = tx.Exec(sql2, dhid, optUid, value, optType, dhid).Error
// 		if err != nil {
// 			syslog.Logger().Errorf("%v", err)
// 			return err
// 		}
// 	}
// 	return nil
// }

func (self *DBMgr) SumHouseMemTodayTotalScore(dHid int64, members ...int64) float64 {
	return self.db_M.Sum(&models.RecordGameDay{},
		"total_score",
		"hid = ? And uid in(?) And DATE_FORMAT(created_at, '%Y-%m-%d') = ?",
		dHid, members, time.Now().Format("2006-01-02"))

}

type HMVitamin struct {
	Vitamin int64
}

func (self *DBMgr) GetUserLatestVitaminFromDataBase(hid, uid int64) (int64, error) {
	sql := `select vitamin from house_member_vitamin where hid = ? and uid = ?`
	dest := HMVitamin{}
	err := GetDBMgr().GetDBmControl().Raw(sql, hid, uid).Scan(&dest).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, nil
		}
		xlog.Logger().Error("GetUserLatestVitaminFromDataBase.Error:", err)
		return 0, err
	}
	return dest.Vitamin, nil
}

func (self *DBMgr) GetUsersLatestVitaminFromDataBase(hid int64, uids ...int64) (map[int64]int64, error) {
	sql := `select vitamin from house_member_vitamin where hid = ? and uid in(?)`
	res := make(map[int64]int64)
	dest := make([]struct {
		Uid     int64
		Vitamin int64
	}, len(uids))
	err := GetDBMgr().GetDBmControl().Raw(sql, hid, uids).Scan(&dest).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return res, nil
		}
		xlog.Logger().Error("GetUserLatestVitaminFromDataBase.Error:", err)
		return res, err
	}

	for _, r := range dest {
		res[r.Uid] = r.Vitamin
	}

	return res, nil
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

// ! 检查是否为vip
func (self *DBMgr) CheckIsHigher(uid int64) bool {
	// 检查版本是否支持
	if !self.db_R.IsHotVersionSupport(fmt.Sprint(uid)) {
		var count int64
		err := self.db_M.Model(&models.User{}).Where("id = ?", uid).
			Where("is_vip = ?", true).Count(&count).Error
		if err != nil {
			if gorm.IsRecordNotFoundError(err) {
				return false
			}
			xlog.Logger().Error(err)
			return false
		}
		return count > 0
	} else {
		return true
	}
}

func (self *DBMgr) LoadSuperAdmin() map[int64]int {
	mapData := make(map[int64]int)
	tmpData := self.GetDBrControl().RedisV2.HGetAll("superman").Val()
	for key, value := range tmpData {
		mapData[static.HF_Atoi64(key)] = static.HF_Atoi(value)
	}
	return mapData
}
