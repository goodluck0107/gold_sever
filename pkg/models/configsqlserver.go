package models

import (
	"github.com/jinzhu/gorm"
)

type ConfigSqlserver struct {
	Id      int64  `gorm:"primary_key;column:id"`   // 自增id
	Able    int    `gorm:"column:able;default:0"`   // 开关
	Name    string `gorm:"column:name;size:32"`     // 哪个游戏的数据库
	Connect string `gorm:"column:connect;size:255"` // 连接字符串
}

func (ConfigSqlserver) TableName() string {
	return "config_sqlserver"
}

func initConfigSqlserver(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigSqlserver{}) {
		err = db.AutoMigrate(&ConfigSqlserver{}).Error
	} else {
		err = db.CreateTable(&ConfigSqlserver{}).Error
	}
	return err
}
