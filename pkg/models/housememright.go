package models

import (
	"github.com/jinzhu/gorm"
)

//默认权限表
type HouseMemberRight struct {
	Id          int    `gorm:"primary_key;column:id"`
	URole       int    `gorm:"column:urole"`        //! 角色 1盟主 2裁判  3管理员  4队长  5副队长  6普通成员
	BigId       int    `gorm:"column:big_id"`       //! 权限大类id
	BigKey      string `gorm:"column:big_key"`      //! 权限大类key
	BigName     string `gorm:"column:big_name"`     //! 权限大类name
	MinorId     int    `gorm:"column:minor_id"`     //子权限ID     可以用作排序参考
	MinorKey    string `gorm:"column:minor_key"`    //子权限Key
	MinorName   string `gorm:"column:minor_name"`   //子权限Name
	MinorStatus int    `gorm:"column:minor_status"` //子权限值说明 { "无权限": 0, "固定权": 1, "可配置并且默认无": 2, "可配置并且默认有": 3, "禁止配置并且无权限": 4}
}

func (HouseMemberRight) TableName() string {
	return "house_member_right"
}

func initHouseMemberRight(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberRight{}) {
		err = db.AutoMigrate(&HouseMemberRight{}).Error
	} else {
		err = db.CreateTable(&HouseMemberRight{}).Error
	}
	return err
}

//个人权限
type HouseMemberUserRight struct {
	Id     int    `gorm:"primary_key;column:id"`
	Hid    int    `gorm:"column:hid"`              //! 包厢ID
	Uid    int    `gorm:"column:uid"`              //! 玩家ID
	Uright string `gorm:"column:uright;type:text"` //! 权限值
	Dhid   int    `gorm:"column:dhid"`             //! 包厢唯一ID
	Role   int    `gorm:"column:role"`             //! 对应包厢的角色
}

func (HouseMemberUserRight) TableName() string {
	return "house_member_right_user"
}

func initHouseMemberUserRight(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberUserRight{}) {
		err = db.AutoMigrate(&HouseMemberUserRight{}).Error
	} else {
		err = db.CreateTable(&HouseMemberUserRight{}).Error
	}
	return err
}
