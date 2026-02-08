package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type HouseTableDistanceLimit struct {
	Id                 int64     `gorm:"primary_key;column:id"`           // id
	DHId               int64     `gorm:"column:dhid"`                     // 包厢id
	TableDistanceLimit int       `gorm:"column:table_distance_limit"`     // 包厢入桌距离限制
	CreatedAt          time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt          time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseTableDistanceLimit) TableName() string {
	return "house_table_distance_limit"
}

func initHouseTableDistanceLimit(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseTableDistanceLimit{}) {
		err = db.AutoMigrate(&HouseTableDistanceLimit{}).Error
	} else {
		err = db.CreateTable(&HouseTableDistanceLimit{}).Error
		db.Model(&HouseTableDistanceLimit{}).AddIndex("idx_dhid", "dhid")
		db.Model(&HouseTableDistanceLimit{}).AddUniqueIndex("unique_idx_dhid", "dhid")
	}
	return err
}

type HouseTableDistanceLimitLog struct {
	Id                 int64     `gorm:"primary_key;column:id"`           // id
	DHId               int64     `gorm:"column:dhid"`                     // 包厢id
	TableDistanceLimit int       `gorm:"column:table_distance_limit"`     // 包厢入桌距离限制
	CreatedAt          time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (HouseTableDistanceLimitLog) TableName() string {
	return "house_table_distance_limit_log"
}

func initHouseTableDistanceLimitLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseTableDistanceLimitLog{}) {
		err = db.AutoMigrate(&HouseTableDistanceLimitLog{}).Error
	} else {
		err = db.CreateTable(&HouseTableDistanceLimitLog{}).Error
		db.Model(&HouseTableDistanceLimit{}).AddIndex("idx_dhid", "dhid")
	}
	return err
}
