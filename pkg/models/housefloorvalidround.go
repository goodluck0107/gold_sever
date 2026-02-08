package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type HouseFloorValidRound struct {
	Id            int64     `gorm:"primary_key;column:id"`           //! id
	HId           int64     `gorm:"column:hid"`                      //! 包厢id
	FId           int64     `gorm:"column:fid"`                      //! 楼层id
	ValidMinScore int       `gorm:"column:minscore"`                 //! 包厢有效最低分数
	ValidBigScore int       `gorm:"column:bigscore;default:-1"`      //! 包厢超级分数
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime"` //! 更新时间
}

func (HouseFloorValidRound) TableName() string {
	return "house_floor_valid_round"
}

func initHouseFloorValidRound(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseFloorValidRound{}) {
		err = db.AutoMigrate(&HouseFloorValidRound{}).Error
	} else {
		err = db.CreateTable(&HouseFloorValidRound{}).Error
	}
	return err
}
