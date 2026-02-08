package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

const (
	CurrencyType_FangKa  = 0 //房卡
	CurrencyType_Diamond = 1 //钻石
	CurrencyType_Gold    = 2 //金币
)

const (
	StatisticsType_Cost = 0 //消耗
	StatisticsType_Add  = 1 //新增
)

type QueryCurrencyStatisticsResult struct {
	UId             int64 `gorm:"column:uid"`             //! 玩家ID
	StatisticsValue int   `gorm:"column:statisticsvalue"` //! 统计值
}

type CurrencyStatistics struct {
	Id              int       `gorm:"primary_key;column:id"`           //! id
	UId             int64     `gorm:"column:uid"`                      //! 玩家ID
	CurrencyType    int       `gorm:"column:currencytype"`             //! 货币类型(0房卡,1钻石,2金币)
	StatisticsType  int       `gorm:"column:statisticstype"`           //! 统计类型(0消耗,1新增)
	StatisticsValue int       `gorm:"column:statisticsvalue"`          //! 统计值
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (CurrencyStatistics) TableName() string {
	return "currency_statistics"
}

func initCurrencyStatistics(db *gorm.DB) error {
	var err error
	if db.HasTable(&CurrencyStatistics{}) {
		err = db.AutoMigrate(&CurrencyStatistics{}).Error
	} else {
		err = db.CreateTable(&CurrencyStatistics{}).Error
		db.Model(CurrencyStatistics{}).AddIndex("idx_currency_statisticst", "currencytype", "statisticstype")
		db.Model(CurrencyStatistics{}).AddIndex("idx_uid_currency_statisticst", "uid", "currencytype", "statisticstype")
	}

	return err
}

//! 添加伙计统计
func CurrencyStatisticsAdd(db *gorm.DB, uid int64, currencyType, statisticsType int, statisticsValue int) error {
	return db.Save(&CurrencyStatistics{
		UId:             uid,
		CurrencyType:    currencyType,
		StatisticsType:  statisticsType,
		StatisticsValue: statisticsValue,
		CreatedAt:       time.Now()}).Error
}

//! 查询个人货币统计情况
func CurrencyStatisticsSelectByUid(db *gorm.DB, uid int64, currencyType, statisticsType int, beginTime, endTime time.Time) (QueryCurrencyStatisticsResult, error) {
	beginTimeStr := fmt.Sprintf("%d-%02d-%02d", beginTime.Year(), beginTime.Month(), beginTime.Day())
	endTimeStr := fmt.Sprintf("%d-%02d-%02d", endTime.Year(), endTime.Month(), endTime.Day())
	sql := `select uid, sum(statisticsvalue) as statisticsvalue from currency_statistics where uid = ? and currencytype = ? and statisticstype = ? and date_format(created_at, "%Y-%m-%d") >= ? and date_format(created_at, "%Y-%m-%d") <= ? group by uid`

	var result QueryCurrencyStatisticsResult
	err := db.Raw(sql, uid, currencyType, statisticsType, beginTimeStr, endTimeStr).Scan(&result).Error
	return result, err
}

//! 查询批量玩家货币统计情况
func CurrencyStatisticsSelectByUids(db *gorm.DB, uids []int64, currencyType, statisticsType int, beginTime, endTime time.Time) ([]QueryCurrencyStatisticsResult, error) {
	beginTimeStr := fmt.Sprintf("%d-%02d-%02d", beginTime.Year(), beginTime.Month(), beginTime.Day())
	endTimeStr := fmt.Sprintf("%d-%02d-%02d", endTime.Year(), endTime.Month(), endTime.Day())
	sql := `select uid, sum(statisticsvalue) as statisticsvalue from currency_statistics where uid in (?) and currencytype = ? and statisticstype = ? and date_format(created_at, "%Y-%m-%d") >= ? and date_format(created_at, "%Y-%m-%d") <= ? group by uid`

	var result []QueryCurrencyStatisticsResult
	err := db.Raw(sql, uids, currencyType, statisticsType, beginTimeStr, endTimeStr).Scan(&result).Error
	return result, err
}

//! 查询所有货币统计情况
func CurrencyStatisticsSelectByAll(db *gorm.DB, currencyType, statisticsType int, beginTime, endTime time.Time) ([]QueryCurrencyStatisticsResult, error) {
	beginTimeStr := fmt.Sprintf("%d-%02d-%02d", beginTime.Year(), beginTime.Month(), beginTime.Day())
	endTimeStr := fmt.Sprintf("%d-%02d-%02d", endTime.Year(), endTime.Month(), endTime.Day())
	sql := `select uid, sum(statisticsvalue) as statisticsvalue from currency_statistics where currencytype = ? and statisticstype = ? and date_format(created_at, "%Y-%m-%d") >= ? and date_format(created_at, "%Y-%m-%d") <= ? group by uid`

	var result []QueryCurrencyStatisticsResult
	err := db.Raw(sql, currencyType, statisticsType, beginTimeStr, endTimeStr).Scan(&result).Error
	return result, err
}
