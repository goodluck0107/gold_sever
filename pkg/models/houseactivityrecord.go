package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type HouseActivityRecord struct {
	Id        int64     `gorm:"primary_key;column:id"`                               //! id
	ActId     int64     `gorm:"column:actid;unique_index:idx_user_act" json:"actid"` //! 包厢id
	UId       int64     `gorm:"column:uid;unique_index:idx_user_act" json:"uid"`     //! 楼层id
	RankScore float64   `gorm:"column:rankscore" json:"score"`                       //! 分数
	CreatedAt time.Time `gorm:"column:created_at;type:datetime" json:"-"`            //! 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime" json:"-"`            //! 更新时间
}

func (HouseActivityRecord) TableName() string {
	return "house_activity_record"
}

func InitHouseActRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseActivityRecord{}) {
		err = db.AutoMigrate(&HouseActivityRecord{}).Error
	} else {
		err = db.CreateTable(&HouseActivityRecord{}).Error
	}
	return err
}

type HouseActivityRecordLog struct {
	Id        int64     `gorm:"primary_key;column:id"`           //! id
	ActId     int64     `gorm:"column:actid"`                    //! 包厢id
	UId       int64     `gorm:"column:uid"`                      //! 楼层id
	RankScore float64   `gorm:"column:rankscore"`                //! 分数
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` //! 更新时间
}

func (HouseActivityRecordLog) TableName() string {
	return "house_activity_record_log"
}

func InitHouseActRecordLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseActivityRecordLog{}) {
		err = db.AutoMigrate(&HouseActivityRecordLog{}).Error
	} else {
		err = db.CreateTable(&HouseActivityRecordLog{}).Error
	}
	return err
}
