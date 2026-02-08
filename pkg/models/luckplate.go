package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

//!
type LuckTicket struct {
	Id        int64     `gorm:"primary_key;column:id"`                                                                  // id
	ActId     int64     `gorm:"column:actid"`                                                                           // 加盟商id
	Hid       int64     `gorm:"column:hid"`                                                                             // 房卡数
	Uid       int64     `gorm:"column:uid"`                                                                             // 房卡数
	GameRound int64     `gorm:"column:game_round"`                                                                      // 区域编号
	UsedRound bool      `gorm:"column:used_round"`                                                                      // 是否冻结
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;DEFAULT :CURRENT_TIMESTAMP"`                             // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;DEFAULT:CURRENT_TIMESTAMP ;ON_UPDATE:CURRENT_TIMESTAMP"` // 更新时间
}

func (LuckTicket) TableName() string {
	return "luck_ticket"
}

func initLuckTicket(db *gorm.DB) error {
	var err error
	if db.HasTable(&LuckTicket{}) {
		err = db.AutoMigrate(&LuckTicket{}).Error
	} else {
		err = db.CreateTable(&LuckTicket{}).Error
	}
	return err
}

type LuckRecord struct {
	Id        int64     `gorm:"primary_key;column:id"`                                                                  // id
	ActId     int64     `gorm:"column:actid"`                                                                           // 加盟商id
	Hid       int64     `gorm:"column:hid"`                                                                             // 房卡数
	Uid       int64     `gorm:"column:uid"`                                                                             // 房卡数
	Rank      int64     `gorm:"column:rank"`                                                                            // 区域编号
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;DEFAULT :CURRENT_TIMESTAMP"`                             // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;DEFAULT:CURRENT_TIMESTAMP ;ON_UPDATE:CURRENT_TIMESTAMP"` // 更新时间
}

func (LuckRecord) TableName() string {
	return "luck_record"
}

func initLuckRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&LuckRecord{}) {
		err = db.AutoMigrate(&LuckRecord{}).Error
	} else {
		err = db.CreateTable(&LuckRecord{}).Error
	}
	return err
}

//需要使用sql创建，索引及不重复主键
type LuckConfig struct {
	Id        int64     `gorm:"primary_key;column:id"` // id
	ActId     int64     `gorm:"column:actid"`
	Hid       int64     `gorm:"column:hid"`
	OpUid     int64     `gorm:"column:opuid"`
	Rank      int64     `gorm:"column:rank"`
	Count     int64     `gorm:"count"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;DEFAULT :CURRENT_TIMESTAMP"`                             // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;DEFAULT:CURRENT_TIMESTAMP ;ON_UPDATE:CURRENT_TIMESTAMP"` // 更新时间
}

func (LuckConfig) TableName() string {
	return "luck_config"
}

func initLuckConfig(db *gorm.DB) error {
	var err error
	if db.HasTable(&LuckConfig{}) {
		err = db.AutoMigrate(&LuckConfig{}).Error
	} else {
		err = db.CreateTable(&LuckConfig{}).Error
	}
	return err
}
