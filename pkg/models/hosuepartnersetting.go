package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 合伙人对包厢设置表
type HousePartnerSetting struct {
	Id          int64     `gorm:"primary_key;column:id;comment:'主键id'"`
	Hid         int       `gorm:"column:hid;comment:'包厢id'"`
	Fid         int64     `gorm:"column:fid;comment:'楼层id'"`
	Pid         int64     `gorm:"column:pid;comment:'合伙人id'"`
	LowScoreVal int       `gorm:"column:low_score_val;comment:'低分局数值'"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;default:current_timestamp;on_update:current_timestamp"`
}

func (HousePartnerSetting) TableName() string {
	return "house_partner_setting"
}

func initHousePartnerSetting(db *gorm.DB) error {
	var err error
	if db.HasTable(&HousePartnerSetting{}) {
		err = db.AutoMigrate(&HousePartnerSetting{}).Error
	} else {
		err = db.CreateTable(&HousePartnerSetting{}).Error
		db.Model(HousePartnerSetting{}).AddUniqueIndex("hid_fid_pid", "Hid", "Fid", "Pid")
		db.Model(HousePartnerSetting{}).AddIndex("hid_pid", "Hid", "Pid")
	}
	return err
}
