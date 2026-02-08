package models

import (
	"github.com/jinzhu/gorm"
)

//包厢开关
type HouseMemberSwitch struct {
	Id            int    `gorm:"primary_key;column:id"`
	Hid           int    `gorm:"column:hid"`                      // 包厢ID
	SwitchContent string `gorm:"column:switch_content;type:text"` // 包厢功能开关
}

func (HouseMemberSwitch) TableName() string {
	return "house_member_switch"
}

func initHouseMemberSwitch(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberSwitch{}) {
		err = db.AutoMigrate(&HouseMemberSwitch{}).Error
	} else {
		err = db.CreateTable(&HouseMemberSwitch{}).Error
	}
	return err
}
