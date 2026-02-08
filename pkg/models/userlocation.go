package models

import "github.com/jinzhu/gorm"

type UserLocation struct {
	Id        int64  `gorm:"primary_key;column:id"`
	Ip        string `gorm:"column:ip;size:20"`        // 玩家ip
	Longitude string `gorm:"column:longitude;size:15"` // 经度
	Latitude  string `gorm:"column:latitude;size:15"`  // 纬度
	Address   string `gorm:"column:address;size:50"`   // 地址
}

func (UserLocation) TableName() string {
	return "user_location"
}

func initUserLocation(db *gorm.DB) error {
	var err error
	if db.HasTable(&User{}) {
		err = db.AutoMigrate(&UserLocation{}).Error
	} else {
		err = db.CreateTable(&UserLocation{}).Error
	}
	return err
}
