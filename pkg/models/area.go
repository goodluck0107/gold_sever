package models

import (
	"github.com/jinzhu/gorm"
)

// 区域表
type Area struct {
	AreaId              int `gorm:"primary_key;column:area_id"`    //! 区域ID
	IsShowPlayingRecord int `gorm:"column:is_show_playing_record"` //! 是否显示正在游戏中的战绩记录
}

func (Area) TableName() string {
	return "area"
}

func initArea(db *gorm.DB) error {
	var err error
	if db.HasTable(&Area{}) {
		err = db.AutoMigrate(&Area{}).Error
	} else {
		err = db.CreateTable(&Area{}).Error
	}
	return err
}
