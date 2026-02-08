package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type HouseMemberLeaveLink struct {
	Id        int64     `gorm:"primary_key;column:id"`           //! id
	DHId      int64     `gorm:"column:dhid"`                     //! 包厢id
	UId       int64     `gorm:"column:uid"`                      //! 玩家id
	Partner   int64     `gorm:"column:partner"`                  //! 玩家id
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (HouseMemberLeaveLink) TableName() string {
	return "house_member_leavelink"
}

func initHouseMemberLeaveLink(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberLeaveLink{}) {
		err = db.AutoMigrate(&HouseMemberLeaveLink{}).Error
	} else {
		err = db.CreateTable(&HouseMemberLeaveLink{}).Error
		db.Model(HouseMemberLeaveLink{}).AddUniqueIndex("unique_idx_dhid_uid", "dhid", "uid")
		db.Model(HouseMemberLeaveLink{}).AddIndex("unique_idx_dhid_partner", "dhid", "partner")
	}
	return err
}
