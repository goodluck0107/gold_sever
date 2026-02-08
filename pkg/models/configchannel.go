package models

import (
	"github.com/jinzhu/gorm"
)

// 应用配置表
type ConfigApp struct {
	AppId       string `gorm:"column:appid;size:32"`       // 应用id
	WxAppId     string `gorm:"column:wxappid;size:64"`     // 微信登录app_id
	WxAppSecret string `gorm:"column:wxappsecert;size:64"` // 微信登录app_secert
	PayId       string `gorm:"column:payid;size:64"`       // 支付id
	PayToken    string `gorm:"column:paytoken;size:64"`    // 支付token
}

func (ConfigApp) TableName() string {
	return "config_app"
}

func initConfigApp(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigApp{}) {
		err = db.AutoMigrate(&ConfigApp{}).Error
	} else {
		err = db.CreateTable(&ConfigApp{}).Error
	}
	return err
}
