package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//! 游戏场次平均分统计表
type DayAvgScore struct {
	Id          int64     `gorm:"primary_key;AUTO_INCREMENT;column:id"` // id
	KindId      int       `gorm:"column:kind_id"`                       //! 子游戏id
	SiteType    int       `gorm:"column:site_type"`                     //! 场次类型
	AvgScore    int       `gorm:"column:avg_score"`                     //! 平均分
	CollectTime time.Time `gorm:"column:collect_time"`                  //! 数据采集时间
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"`      //! 数据创建时间
}

func (DayAvgScore) TableName() string {
	return "user_day_avgscore"
}

func initDayAvgScore(db *gorm.DB) error {
	var err error
	if db.HasTable(&DayAvgScore{}) {
		err = db.AutoMigrate(&DayAvgScore{}).Error
	} else {
		err = db.CreateTable(&DayAvgScore{}).Error
	}
	return err
}
