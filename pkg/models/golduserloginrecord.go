package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type GoldUserLoginRecord struct {
	Id          int64     `gorm:"primary_key;column:id"`
	Uid         int64     `gorm:"column:uid"`                      // 玩家id
	Platform    int       `gorm:"column:platform"`                 // 平台
	Ip          string    `gorm:"column:ip;size:32"`               // ip
	MachineCode string    `gorm:"column:machine_code;size:32"`     // 机器码
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` // 创建时间(登录时间)
}

func (GoldUserLoginRecord) TableName() string {
	return "gold_user_login_record"
}

func initGoldUserLoginRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&GoldUserLoginRecord{}) {
		err = db.AutoMigrate(&GoldUserLoginRecord{}).Error
	} else {
		err = db.CreateTable(&GoldUserLoginRecord{}).Error

		db.Model(GoldUserLoginRecord{}).AddIndex("idx_created_at", "created_at")
	}
	return err
}

func GoldUserLoginRecordAdd(db *gorm.DB, uid int64, platform int, ip string, machinecode string, time time.Time) {
	db.Create(&GoldUserLoginRecord{
		Uid:         uid,
		Platform:    platform,
		Ip:          ip,
		MachineCode: machinecode,
		CreatedAt:   time,
	})
}
