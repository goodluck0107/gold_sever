package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 在线人数表
type GameOnline struct {
	Id        int64     `gorm:"primary_key;column:id"`           //! id
	KindId    int       `gorm:"column:kindid"`                   //! 玩法id
	Value     int       `gorm:"column:value"`                    //! 在线人数
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (GameOnline) TableName() string {
	return "game_online"
}

func initGameOnline(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameOnline{}) {
		err = db.AutoMigrate(&GameOnline{}).Error
	} else {
		err = db.CreateTable(&GameOnline{}).Error
	}
	return err
}
