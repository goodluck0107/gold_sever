package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 疲劳值每管理列表统计
type RecordVitaminMgrList struct {
	Id                 int       `gorm:"primary_key;column:id"`           //! id
	DHId               int64     `gorm:"column:hid"`                      //! 包厢dhid
	UId                int64     `gorm:"column:uid"`                      //! 包厢玩家id
	PreNodeVitamin     int64     `gorm:"column:prenodevitamin"`           //! 上节点疲劳值
	VitaminWinLoseCost int64     `gorm:"column:vitaminwinlosecost"`       //! 输赢扣除疲劳值数量
	VitaminPlayCost    int64     `gorm:"column:vitaminplaycost"`          //! 对局扣除疲劳值数量
	VitaminCostRound   int64     `gorm:"column:vitamincostround"`         //! AA对局扣除疲劳值数量
	VitaminCostBW      int64     `gorm:"column:vitamincostbw"`            //! 大赢家扣除疲劳值数量
	Recording          int       `gorm:"column:recording"`                //! 记录中
	CreatedAt          time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt          time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (RecordVitaminMgrList) TableName() string {
	return "house_vitamin_mgr"
}

func initRecordVitaminMgr(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordVitaminMgrList{}) {
		err = db.AutoMigrate(&RecordVitaminMgrList{}).Error
	} else {
		err = db.CreateTable(&RecordVitaminMgrList{}).Error
	}
	return err
}
