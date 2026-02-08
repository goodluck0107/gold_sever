package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 发放商品记录表
type DeliverProductRecord struct {
	Id          int       `gorm:"primary_key;column:id"`         //! id
	UId         int64     `gorm:"column:uid"`                    //! 玩家用户ID
	ProductName string    `gorm:"column:product_name;size:55"`   //! 商品名称
	Type        int8      `json:"type"`                          //! 下发类型
	WealthType  int8      `json:"product_type"`                  //! 财富类型
	Num         int       `json:"num"`                           //! 数量
	Price       int       `json:"price"`                         //! 商品价格(单位：分)
	Order       string    `gorm:"column:order"`                  //！订单号
	Appid       string    `gorm:"column:appid;size:55"`          //! 订单渠道来源
	CreatedAt   time.Time `gorm:"column:ctime;type:datetime"`    //! 订单创建时间
	PayWay      int       `gorm:"column:way;"`                   //! 支付方式
	Passtime    time.Time `gorm:"column:passtime;type:datetime"` //! 支付时间
}

func (DeliverProductRecord) TableName() string {
	return "record_deliver_product"
}

func initDeliverProductRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&DeliverProductRecord{}) {
		err = db.AutoMigrate(&DeliverProductRecord{}).Error
	} else {
		err = db.CreateTable(&DeliverProductRecord{}).Error
		if err == nil {
			// 修改递增初始值
			err = db.Exec("alter table shop_record AUTO_INCREMENT=1000000").Error
		}
	}
	return err
}
