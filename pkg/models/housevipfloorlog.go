package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type HouseVipFloorLog struct {
	Id          int64     `gorm:"primary_key;column:id"` //! id
	Fid         int64     `gorm:"column:fid"`
	OptId       int64     `gorm:"column:optid"`
	Uid         int64     `gorm:"column:uid"`
	UVip        bool      `gorm:"column:uvip"`
	FVip        bool      `gorm:"column:fvip"`
	FCapSetVip  bool      `gorm:"column:f_cap_set_vip"`
	FDefJoinVip bool      `gorm:"column:f_def_join_vip"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (HouseVipFloorLog) TableName() string {
	return "house_vip_floor_log"
}

func initHouseVipFloorLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseVipFloorLog{}) {
		err = db.AutoMigrate(&HouseVipFloorLog{}).Error
	} else {
		err = db.CreateTable(&HouseVipFloorLog{}).Error
	}
	return err
}
