package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 疲劳值每日统计数据
type HouseMemberVitaminDay struct {
	Id             int       `gorm:"primary_key;column:id"`           //! id
	DHId           int64     `gorm:"column:hid"`                      //! 包厢dhid
	UId            int64     `gorm:"column:uid"`                      //! 包厢玩家id
	VitaminLeft    int64     `gorm:"column:vitaminleft"`              //! 扣除疲劳值数量
	StatisticsDate string    `gorm:"column:statisticsdate"`           //! 统计时间
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseMemberVitaminDay) TableName() string {
	return "house_member_vitamin_day"
}

func initHouseMemberVitaminDay(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberVitaminDay{}) {
		err = db.AutoMigrate(&HouseMemberVitaminDay{}).Error
	} else {
		err = db.CreateTable(&HouseMemberVitaminDay{}).Error
	}
	return err
}
