package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type ConfigBattleLevel struct {
	Id    int64 `gorm:"column:id;primary_key"` //! 商品ID
	Level int
	Desc  string
	Limit int
}

func (ConfigBattleLevel) TableName() string {
	return "config_battle_level"
}

func initConfigBattleLevel(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigBattleLevel{}) {
		err = db.AutoMigrate(&ConfigBattleLevel{}).Error
	} else {
		err = db.CreateTable(&ConfigBattleLevel{}).Error
	}
	return err
}

type ConfigSpinBase struct {
	Id          int `gorm:"column:id;primary_key"`
	BattleRound int64
}

func (ConfigSpinBase) TableName() string {
	return "config_spin_base"
}

func initConfigSpinBase(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigSpinBase{}) {
		err = db.AutoMigrate(&ConfigSpinBase{}).Error
	} else {
		err = db.CreateTable(&ConfigSpinBase{}).Error
	}
	return err
}

type ConfigSpinAward struct {
	Id     int64 `gorm:"column:id;primary_key"`
	Seq    int
	Desc   string
	Type   int
	Count  int64
	Weight int
	Icon   string `gorm:"column:icon;type:varchar(300)"`
}

func (ConfigSpinAward) TableName() string {
	return "config_spin_award"
}

func (r *ConfigSpinAward) TypeDesc() string {
	if r.Type > 0 {
		return "实物"
	}
	return "金币"
}

func initConfigSpinAward(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigSpinAward{}) {
		err = db.AutoMigrate(&ConfigSpinAward{}).Error
	} else {
		err = db.CreateTable(&ConfigSpinAward{}).Error
	}
	return err
}

type RecordSpinAward struct {
	Id        int64 `gorm:"column:id;primary_key"`
	Uid       int64
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	AwardId   int64
	Seq       int
	Desc      string
	Type      int
	Count     int64
	Weight    int
	Icon      string `gorm:"column:icon;type:varchar(300)"`
}

func (r *RecordSpinAward) TypeDesc() string {
	if r.Type > 0 {
		return "实物"
	}
	return "金币"
}

func (RecordSpinAward) TableName() string {
	return "record_spin_award"
}

func initRecordSpinAward(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordSpinAward{}) {
		err = db.AutoMigrate(&RecordSpinAward{}).Error
	} else {
		err = db.CreateTable(&RecordSpinAward{}).Error
	}
	return err
}

type ConfigCheckin struct {
	Id   int `gorm:"column:id;primary_key"`
	Gold int64
}

func (ConfigCheckin) TableName() string {
	return "config_checkin"
}

func initConfigCheckin(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigCheckin{}) {
		err = db.AutoMigrate(&ConfigCheckin{}).Error
	} else {
		err = db.CreateTable(&ConfigCheckin{}).Error
	}
	return err
}

type RecordCheckin struct {
	Id        int64 `gorm:"column:id;primary_key"`
	Uid       int64 `json:"uid"`
	Day       int
	Gold      int64
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (RecordCheckin) TableName() string {
	return "record_checkin"
}

func initRecordCheckin(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordCheckin{}) {
		err = db.AutoMigrate(&RecordCheckin{}).Error
	} else {
		err = db.CreateTable(&RecordCheckin{}).Error
	}
	return err
}
