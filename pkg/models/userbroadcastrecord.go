package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 用户区域广播记录
type UserBroadcastRecord struct {
	Id        int64     `gorm:"primary_key;column:id"`
	Uid       int64     `gorm:"column:uid"`                      // 玩家id
	Area      string    `gorm:"column:area;size:8"`              // 区域
	Content   string    `gorm:"column:content;size:255"`         // 广播内容
	Cost      int       `gorm:"column:cost"`                     //! 消耗金币
	Ip        string    `gorm:"column:ip;size:32"`               // ip
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (UserBroadcastRecord) TableName() string {
	return "user_broadcast_record"
}

func initUserBroadcastRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserBroadcastRecord{}) {
		err = db.AutoMigrate(&UserBroadcastRecord{}).Error
	} else {
		err = db.CreateTable(&UserBroadcastRecord{}).Error
	}
	return err
}
