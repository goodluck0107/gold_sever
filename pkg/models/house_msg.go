package models

import "github.com/jinzhu/gorm"

type HouseMsg struct {
	Id          int64  `gorm:"primary_key;column:id"`    //! id
	HId         int    `gorm:"column:hid;index"`         //! 包厢id
	Creater     int64  `gorm:"column:creater;default:0"` //! 玩家id
	Msg         string `gorm:"column:msg;type:varchar(200)"`
	MsgType     int64  `gorm:"column:msg_type;default:0"`
	CreateStamp int64  `gorm:"column:create_stamp;not null"`
}

func (HouseMsg) TableName() string {
	return "house_msg"
}

func initHouseMsg(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMsg{}) {
		err = db.AutoMigrate(&HouseMsg{}).Error
	} else {
		err = db.CreateTable(&HouseMsg{}).Error
	}
	return err
}

type HouseFloorDelMsg struct {
	Id           int64  `gorm:"primary_key;column:id"` //! id
	DHId         int64  `gorm:"column:dhid"`           //! 包厢id
	DFId         int64  `gorm:"column:dfid"`           //! 楼层id
	DFIndex      int    `gorm:"column:dfindex"`        //! 楼层索引
	CreateStamp  int64  `gorm:"column:create_stamp;not null"`
	FloorRoyalty string `gorm:"column:floorroyalty;type:text"` //! 包厢合伙人分层配置
	FloorPartner string `gorm:"column:floorpartner;type:text"` //! 包厢合伙人关系
}

func (HouseFloorDelMsg) TableName() string {
	return "house_floor_del_msg"
}

func initHouseFloorDelMsg(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseFloorDelMsg{}) {
		err = db.AutoMigrate(&HouseFloorDelMsg{}).Error
	} else {
		err = db.CreateTable(&HouseFloorDelMsg{}).Error
		db.Model(&HouseFloorDelMsg{}).AddIndex("idx_dhid_dfindex", "dhid", "dfindex")
		db.Model(&HouseFloorDelMsg{}).AddIndex("idx_create_stamp", "create_stamp")
	}
	return err
}
