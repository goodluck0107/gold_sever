package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type HousePartnerInviteCode struct {
	Id         int64     `gorm:"primary_key;column:id"`           //! id
	InviteCode int64     `gorm:"column:invitecode"`               //! 邀请码
	IsUsed     bool      `gorm:"column:isused"`                   //! 是否被使用
	UsedHid    int64     `gorm:"column:usedhid"`                  //! 被使用的包厢id
	UseUid     int64     `gorm:"column:useduid"`                  //! 被使用的用户id
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime"` //! 更新时间
}

func (HousePartnerInviteCode) TableName() string {
	return "house_partner_invitecode"
}

func initHousePartnerInviteCode(db *gorm.DB) error {
	var err error
	if db.HasTable(&HousePartnerInviteCode{}) {
		err = db.AutoMigrate(&HousePartnerInviteCode{}).Error
	} else {
		err = db.CreateTable(&HousePartnerInviteCode{}).Error
	}
	return err
}
