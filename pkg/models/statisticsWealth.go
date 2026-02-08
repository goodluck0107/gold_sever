package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//! 财富相关统计表(每天统计一次)
type StatisticsWealth struct {
	Id         int64     `gorm:"primary_key;column:id"`           // id
	Date       time.Time `gorm:"column:date;type:date"`           // 日期
	WealthType int8      `gorm:"column:wealth_type"`              // 财富类型
	CostType   int8      `gorm:"column:cost_type"`                // 流水类型
	Num        int       `gorm:"column:num"`                      // 总量
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (StatisticsWealth) TableName() string {
	return "statistics_wealth"
}

func initStatisticsWealth(db *gorm.DB) error {
	var err error
	if db.HasTable(&StatisticsWealth{}) {
		err = db.AutoMigrate(&StatisticsWealth{}).Error
	} else {
		err = db.CreateTable(&StatisticsWealth{}).Error
	}
	return err
}
