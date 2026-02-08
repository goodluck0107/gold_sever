package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 疲劳值每日统计数据
type RecordVitaminDay struct {
	Id               int       `gorm:"primary_key;column:id"`           //! id
	DHId             int64     `gorm:"column:hid"`                      //! 包厢dhid
	FId              int64     `gorm:"column:fid"`                      //! 包厢dfid
	DFId             int64     `gorm:"column:dfid"`                     //! 包厢dfid
	UId              int64     `gorm:"column:uid"`                      //! 包厢玩家id
	VitaminCost      int64     `gorm:"column:vitamincost"`              //! 扣除疲劳值数量
	VitaminCostRound int64     `gorm:"column:vitamincostround"`         //! AA对局扣除疲劳值数量
	VitaminCostBW    int64     `gorm:"column:vitamincostbw"`            //! 大赢家扣除疲劳值数量
	VitaminWinLose   int64     `gorm:"column:vitaminwinlose"`           //! 玩家输赢扣除疲劳值数量
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt        time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
	Partner          int64     `gorm:"column:partner;default:0"`        //! 队长归属
}

func (RecordVitaminDay) TableName() string {
	return "house_vitamin_record"
}

func initRecordVitaminDay(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordVitaminDay{}) {
		err = db.AutoMigrate(&RecordVitaminDay{}).Error
	} else {
		err = db.CreateTable(&RecordVitaminDay{}).Error
	}
	return err
}
