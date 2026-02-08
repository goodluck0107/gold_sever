package models

import (
	"github.com/jinzhu/gorm"
)

type UserWhite struct {
	Uid         int64  `gorm:"column:uid;primary_key"`
	MachineCode string `gorm:"column:machine_code"`
}

func (*UserWhite) TableName() string {
	return "user_white"
}

func initUserWhite(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserWhite{}) {
		err = db.AutoMigrate(&UserWhite{}).Error
	} else {
		err = db.CreateTable(&UserWhite{}).Error
	}
	return err
}
