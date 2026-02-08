package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 包厢队长属性表
type HousePartnerAttr struct {
	Id             int64     `gorm:"primary_key;column:id"`             //! id
	DHid           int64     `gorm:"column:dhid"`                       //! 圈id
	Uid            int64     `gorm:"column:uid"`                        //! 队长id
	Exp            int64     `gorm:"column:exp"`                        //! 队长经验
	AlarmValue     int64     `gorm:"column:alarm_value;default:-1"`     //! 警戒值
	TeamBan        bool      `gorm:"column:team_ban"`                   //! 全队禁止
	AA             bool      `gorm:"column:aa"`                         //! 是否开启了AA扣卡
	RewardSuperior int       `gorm:"column:reward_superior;default:-1"` //! 上级队长配置的返点比例【盟主>组长】【组长->大队长】【大队长->小队长】
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime"`   //! 创建时间
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime"`   //! 更新时间
}

func (HousePartnerAttr) TableName() string {
	return "house_partner_attr"
}

func initHousePartnerAttr(db *gorm.DB) error {
	var err error
	if db.HasTable(&HousePartnerAttr{}) {
		err = db.AutoMigrate(&HousePartnerAttr{}).Error
	} else {
		err = db.CreateTable(&HousePartnerAttr{}).Error
		if err == nil {
			err = db.Model(&HousePartnerAttr{}).AddUniqueIndex("unique_idx_dhid_uid", "dhid", "uid").Error
		}
	}
	return err
}
