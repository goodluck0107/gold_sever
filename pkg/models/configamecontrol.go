package models

import (
	"github.com/jinzhu/gorm"
)

// 游戏功能概率控制
type ConfigGameControl struct {
	Id             int64 `gorm:"primary_key;column:id"`
	KindId         int   `gorm:"unique;column:kindid"`               // 游戏玩法id
	ProbGK         int   `gorm:"column:prob_gk;default:0"`           // 杠开概率
	ProbMagic4     int   `gorm:"column:prob_m4;default:0"`           // 四个癞子概率
	MagicBeforeNum int   `gorm:"column:magic_before_num;default:61"` // 癞子在多少张之前
}

func (ConfigGameControl) TableName() string {
	return "config_game_control"
}

func initConfigGameControl(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigGameControl{}) {
		err = db.AutoMigrate(&ConfigGameControl{}).Error
	} else {
		err = db.CreateTable(&ConfigGameControl{}).Error
	}
	return err
}
