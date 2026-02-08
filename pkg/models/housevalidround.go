package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type HouseValidRound struct {
	Id            int64     `gorm:"primary_key;column:id"`           //! id
	HId           int64     `gorm:"column:hid"`                      //! 包厢id
	ValidMinScore int       `gorm:"column:minscore"`                 //! 包厢有效最低分数
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime"` //! 更新时间
}

func (HouseValidRound) TableName() string {
	return "house_valid_round"
}

func initHouseValidRound(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseValidRound{}) {
		err = db.AutoMigrate(&HouseValidRound{}).Error
	} else {
		err = db.CreateTable(&HouseValidRound{}).Error
	}
	return err
}
