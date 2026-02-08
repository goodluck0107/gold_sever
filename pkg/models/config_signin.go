package models

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
)

type ConfigSignIn struct {
	Id         int64  `gorm:"primary_key;column:id"` // 自增id
	Day        int    `gorm:"column:day;comment:'第一几天'"`
	WealthType int8   `gorm:"column:wealth_type;comment:'财富类型'"`
	WealthNum  int    `gorm:"column:wealth_num;comment:'财富数额'"`
	WealthUrl  string `gorm:"column:wealth_url;comment:'财富图片路径'"`
}

func (ConfigSignIn) TableName() string {
	return "config_signin"
}

func initConfigSignIn(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigSignIn{}) {
		err = db.AutoMigrate(&ConfigSignIn{}).Error
	} else {
		err = db.CreateTable(&ConfigSignIn{}).Error
	}
	return err
}

func (c *ConfigSignIn) ConvertModel() *static.WealthAward {
	return &static.WealthAward{
		WealthType: c.WealthType,
		Num:        c.WealthNum,
		Url:        c.WealthUrl,
	}
}
