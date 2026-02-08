package models

import (
	"github.com/jinzhu/gorm"
)

type PartnerRewardT struct {
	Id          int64  `gorm:"primary_key;column:id"`
	DHid        int64  `gorm:"column:dhid"`
	DFid        int64  `gorm:"column:dfid"`
	GameNum     string `gorm:"column:game_num"`
	PlayerId    int64  `gorm:"column:player_id"`    // 玩家id
	Partner     int64  `gorm:"column:partner"`      // 玩家的队长id
	PartnerType int    `gorm:"column:partner_type"` // 队长类型 0=圈主 1=队长 2=上级队长 3=上上级队长
	Reward      int64  `gorm:"column:reward"`       // 获益多少
	CreatedTime int64  `gorm:"column:created_time"` // 创建时间
	ClearedTime int64  `gorm:"column:cleared_time"` // 结账时间戳
	// 盟主的收益 = 总收益 - 玩家的队长收益 - 玩家的队长的上级收益 = RewardTotal - RewardPartner - RewardSuperior
}

func (PartnerRewardT) TableName() string {
	return "partner_reward_t"
}

func initPartnerRewardT(db *gorm.DB) error {
	var err error
	if db.HasTable(&PartnerRewardT{}) {
		err = db.AutoMigrate(&PartnerRewardT{}).Error
	} else {
		err = db.CreateTable(&PartnerRewardT{}).Error
	}
	return err
}
