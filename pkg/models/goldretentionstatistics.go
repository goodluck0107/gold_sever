package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

//! 用户登录留存查询
func GetUserRetentionStatistics(db *gorm.DB, time time.Time) (int, int, int, int) {
	selectDayStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	selectDay1 := time.AddDate(0, 0, 1)
	selectDay1Str := fmt.Sprintf("%d-%02d-%02d", selectDay1.Year(), selectDay1.Month(), selectDay1.Day())
	selectDay3 := time.AddDate(0, 0, 3)
	selectDay3Str := fmt.Sprintf("%d-%02d-%02d", selectDay3.Year(), selectDay3.Month(), selectDay3.Day())
	selectDay7 := time.AddDate(0, 0, 7)
	selectDay7Str := fmt.Sprintf("%d-%02d-%02d", selectDay7.Year(), selectDay7.Month(), selectDay7.Day())

	var uids []int64

	sql := `select id as uid from gold_user where date_format(created_at, "%Y-%m-%d") = ?`
	if err := db.Raw(sql, selectDayStr).Pluck("uid", &uids).Error; err != nil {
		return 0, 0, 0, 0
	}

	if len(uids) == 0 {
		return 0, 0, 0, 0
	}

	type QueryResult struct {
		Count    int    `gorm:"column:count"`            //! 玩家Id
		GameDate string `gorm:"column:gamedate;size:12"` //! 玩游戏时间
	}

	var results []QueryResult
	sql = `select cast(created_at as date) as gamedate, count(DISTINCT uid) as count from gold_user_login_record where (date_format(created_at, "%Y-%m-%d") = ? or date_format(created_at, "%Y-%m-%d") = ? or date_format(created_at, "%Y-%m-%d") = ?) and uid in (?) group by cast(created_at as date)`
	if err := db.Raw(sql, selectDay1Str, selectDay3Str, selectDay7Str, uids).Scan(&results).Error; err != nil {
		return 0, 0, 0, 0
	}

	userRetention := len(uids)
	userRetention1 := 0
	userRetention3 := 0
	userRetention7 := 0
	for _, result := range results {
		if result.GameDate[0:10] == selectDay1Str {
			userRetention1 = result.Count
		} else if result.GameDate[0:10] == selectDay3Str {
			userRetention3 = result.Count
		} else if result.GameDate[0:10] == selectDay7Str {
			userRetention7 = result.Count
		}
	}
	return userRetention, userRetention1, userRetention3, userRetention7
}

// 玩家游戏记录
type GoldGameRetentionStatistics struct {
	Id       int    `gorm:"primary_key;column:id"`   //! 主键Id
	UId      int64  `gorm:"column:uid"`              //! 玩家Id
	GameDate string `gorm:"column:gamedate;size:12"` //! 玩游戏时间
}

func (GoldGameRetentionStatistics) TableName() string {
	return "gold_game_retention_statistics"
}

func initGoldGameRetentionStatistics(db *gorm.DB) error {
	var err error
	if db.HasTable(&GoldGameRetentionStatistics{}) {
		err = db.AutoMigrate(&GoldGameRetentionStatistics{}).Error
	} else {
		err = db.CreateTable(&GoldGameRetentionStatistics{}).Error
		db.Model(GoldGameRetentionStatistics{}).AddUniqueIndex("idx_gamedate_uid", "gamedate", "uid")
	}
	return err
}

//! 玩家游戏记录增加
func GoldGameRetentionStatisticsAdd(db *gorm.DB, uid int64, time time.Time) {
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	db.Create(&GoldGameRetentionStatistics{
		UId:      uid,
		GameDate: selectStr,
	})
}

//! 查询游戏玩家总数
func GoldGameUserCountSelectByTime(db *gorm.DB, time time.Time) (int, error) {
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())

	type QueryResult struct {
		Count int `gorm:"column:count"`
	}

	sql := `select count(*) as count from gold_game_retention_statistics where gamedate = ?`

	var result QueryResult
	err := db.Raw(sql, selectStr).Scan(&result).Error

	return result.Count, err
}

//! 查询新增用户游戏玩家数量
func GoldGameNewUserCountSelectByTime(db *gorm.DB, time time.Time) (int, error) {
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())

	type QueryResult struct {
		Count int `gorm:"column:count"`
	}

	sql := `select count(*) as count from gold_game_retention_statistics where gamedate = ? and uid in (select id as uid from gold_user where date_format(created_at, "%Y-%m-%d") = ?)`

	var result QueryResult
	err := db.Raw(sql, selectStr, selectStr).Scan(&result).Error
	return result.Count, err
}

//! 用户游戏留存查询
func GetGameRetentionStatistics(db *gorm.DB, time time.Time) (int, int, int, int) {
	selectDayStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	selectDay1 := time.AddDate(0, 0, 1)
	selectDay1Str := fmt.Sprintf("%d-%02d-%02d", selectDay1.Year(), selectDay1.Month(), selectDay1.Day())
	selectDay3 := time.AddDate(0, 0, 3)
	selectDay3Str := fmt.Sprintf("%d-%02d-%02d", selectDay3.Year(), selectDay3.Month(), selectDay3.Day())
	selectDay7 := time.AddDate(0, 0, 7)
	selectDay7Str := fmt.Sprintf("%d-%02d-%02d", selectDay7.Year(), selectDay7.Month(), selectDay7.Day())

	var uids []int64

	sql := `select uid from gold_game_retention_statistics where gamedate = ? group by uid`
	if err := db.Raw(sql, selectDayStr).Pluck("uid", &uids).Error; err != nil {
		return 0, 0, 0, 0
	}

	if len(uids) == 0 {
		return 0, 0, 0, 0
	}

	type QueryResult struct {
		Count    int    `gorm:"column:count"`            //! 玩家Id
		GameDate string `gorm:"column:gamedate;size:12"` //! 玩游戏时间
	}

	var results []QueryResult

	var retDay1, retDay3, retDay7 QueryResult
	sql = `select ? as gamedate, count(*) as count from (select uid from gold_game_retention_statistics where gamedate = ? and uid in (?) group by uid) as uidTable`
	if err := db.Raw(sql, selectDay1Str, selectDay1Str, uids).Scan(&retDay1).Error; err != nil {
		return 0, 0, 0, 0
	}

	sql = `select ? as gamedate, count(*) as count from (select uid from gold_game_retention_statistics where gamedate = ? and uid in (?) group by uid) as uidTable`
	if err := db.Raw(sql, selectDay3, selectDay3, uids).Scan(&retDay3).Error; err != nil {
		return 0, 0, 0, 0
	}

	sql = `select ? as gamedate, count(*) as count from (select uid from gold_game_retention_statistics where gamedate = ? and uid in (?) group by uid) as uidTable`
	if err := db.Raw(sql, selectDay7, selectDay7, uids).Scan(&retDay7).Error; err != nil {
		return 0, 0, 0, 0
	}

	results = append(results, retDay1, retDay3, retDay7)

	gameRetention := len(uids)
	gameRetention1 := 0
	gameRetention3 := 0
	gameRetention7 := 0

	for _, result := range results {
		if result.GameDate[0:10] == selectDay1Str {
			gameRetention1 = result.Count
		} else if result.GameDate[0:10] == selectDay3Str {
			gameRetention3 = result.Count
		} else if result.GameDate[0:10] == selectDay7Str {
			gameRetention7 = result.Count
		}
	}
	return gameRetention, gameRetention1, gameRetention3, gameRetention7
}
