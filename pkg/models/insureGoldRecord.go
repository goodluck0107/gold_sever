package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

const (
	InsureGoldSave  = 1 // 往保险箱存金币
	InsureGoldFetch = 2 // 从保险箱取金币
)

//! 保险箱存取记录表
type InsureGoldRecord struct {
	Id               int64     `gorm:"primary_key;column:id"`           // id
	Uid              int64     `gorm:"column:uid"`                      // 玩家用户ID
	Type             int8      `gorm:"column:type"`                     // 操作类型
	BeforeGold       int       `gorm:"column:before_gold"`              //! 操作前金币
	AfterGold        int       `gorm:"column:after_gold"`               //! 操作后金币
	BeforeInsureGold int       `gorm:"column:before_insure_gold"`       //! 操作前保险箱金币
	AfterInsureGold  int       `gorm:"column:after_insure_gold"`        //! 操作后保险箱金币
	Num              int       `gorm:"column:num"`                      //! 划账数量
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (InsureGoldRecord) TableName() string {
	return "insure_gold_record"
}

func initInsureGoldRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&InsureGoldRecord{}) {
		err = db.AutoMigrate(&InsureGoldRecord{}).Error
	} else {
		err = db.CreateTable(&InsureGoldRecord{}).Error
	}
	return err
}
