package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 疲劳值每日统计清算数据
type RecordVitaminDayClear struct {
	Id               int       `gorm:"primary_key;column:id"`           //! id
	DHId             int64     `gorm:"column:hid"`                      //! 包厢dhid
	Recording        int       `gorm:"column:recording"`                //! 记录中
	VitaminLeft      int64     `gorm:"column:vitaminleft"`              //! 疲劳值剩余
	VitaminMinus     int64     `gorm:"column:vitaminminus"`             //! 疲劳值负数余额
	VitaminCost      int64     `gorm:"column:vitamincost"`              //! 扣除疲劳值数量
	VitaminCostRound int64     `gorm:"column:vitamincostround"`         //! AA对局扣除疲劳值数量
	VitaminCostBW    int64     `gorm:"column:vitamincostbw"`            //! 大赢家扣除疲劳值数量
	VitaminPayment   int64     `gorm:"column:vitaminpayment"`           //! 收支统计
	BeginAt          time.Time `gorm:"column:begin_at;type:datetime"`   //! 结算起始时间
	EndAt            time.Time `gorm:"column:end_at;type:datetime"`     //！结算结束时间
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	UpdatedAt        time.Time `gorm:"column:updated_at;type:datetime"` //！更新时间
}

func (RecordVitaminDayClear) TableName() string {
	return "house_vitamin_record_clear"
}

func initRecordVitaminDayClear(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordVitaminDayClear{}) {
		err = db.AutoMigrate(&RecordVitaminDayClear{}).Error
	} else {
		err = db.CreateTable(&RecordVitaminDayClear{}).Error
	}
	return err
}
