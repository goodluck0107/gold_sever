package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 兑换商品配置表
type ConfigShop struct {
	Id      int    `gorm:"column:id"`             //! 商品ID
	Left    int    `gorm:"column:left"`           //! 剩余数量
	Price   int    `gorm:"colunm:price"`          //! 花费
	Deleted int    `gorm:"colum:deleted"`         //! 是否生效：0生效；1不生效
	Name    string `gorm:"column:name;size:55"`   //! 商品名称
	Image   string `gorm:"column:image;size:255"` //! 商品图片
	Order   int    `gorm:"column:order"`          //! 排名(客户端用来排序处理)
	Type    int    `gorm:"column:type"`           //! 商品类型 (0:实物奖励；1：金币奖励；2：房卡奖励；3：话费奖励)
	Num     int    `gorm:"column:num"`            //! 数量
	Gift    int    `gorm:"column:gift"`           //! 赠送
}

func (ConfigShop) TableName() string {
	return "config_shop"
}

func initConfigShop(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigShop{}) {
		err = db.AutoMigrate(&ConfigShop{}).Error
	} else {
		err = db.CreateTable(&ConfigShop{}).Error
	}
	return err
}

// 兑换记录表
type ShopRecord struct {
	Id        int       `gorm:"primary_key;column:id"`           //! id
	UId       int64     `gorm:"column:uid"`                      //! 玩家用户ID
	ProductId int       `gorm:"column:productId"`                //! 商品ID
	Type      int       `gorm:"column:type"`                     //! 商品类型 (0:实物奖励；1：金币奖励；2：房卡奖励；3：话费奖励)
	Name      string    `gorm:"column:name;size:55"`             //! 商品名称
	Price     int       `gorm:"colunm:price"`                    //! 花费
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	Num       int       `gorm:"column:num"`                      //! 数量
	Info      string    `gorm:"column:eve;size:55"`              //! 商品处理数据
	Status    int       `gorm:"column:status"`                   //! 发放状态：0:待发放；1：已发放；2：违规；3：发放失败\
	Operator  string    `gorm:"column:operator;size:55"`         //! 操作人
	Passtime  time.Time `gorm:"column:passtime;type:datetime"`   //! 操作时间
	Tel       string    `gorm:"column:tel"`                      //! 兑换商品绑定的手机
}

func (ShopRecord) TableName() string {
	return "shop_record"
}

func initShopRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&ShopRecord{}) {
		err = db.AutoMigrate(&ShopRecord{}).Error
	} else {
		err = db.CreateTable(&ShopRecord{}).Error
		if err == nil {
			// 修改递增初始值
			err = db.Exec("alter table shop_record AUTO_INCREMENT=1000001").Error
		}
	}
	return err
}

// 兑换手机记录表
type ShopPhone struct {
	Id  int    `gorm:"primary_key;column:id"` //! id
	UId int64  `gorm:"column:uid"`            //! 玩家用户ID
	Tel string `gorm:"column:tel;size:32"`    //! 手机号
}

func (ShopPhone) TableName() string {
	return "shop_phone"
}

func initShopPhone(db *gorm.DB) error {
	var err error
	if db.HasTable(&ShopPhone{}) {
		err = db.AutoMigrate(&ShopPhone{}).Error
	} else {
		err = db.CreateTable(&ShopPhone{}).Error
	}
	return err
}
