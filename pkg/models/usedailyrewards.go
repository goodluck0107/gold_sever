package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 每日礼包
type UserDailyRewards struct {
	Id        int64 `gorm:"primary_key;column:id"`
	Uid       int64 `gorm:"column:uid"`   // 玩家id
	AwardGold int   `gorm:"column:award"` // 奖励金币数量
	// Double    bool      `gorm:"column:double"`                   // 是否领取了双倍奖励
	Date      time.Time `gorm:"column:date;type:date"`           // 领取日期
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (UserDailyRewards) TableName() string {
	return "user_dailyrewards"
}

func initUserDailyRewards(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserDailyRewards{}) {
		err = db.AutoMigrate(&UserDailyRewards{}).Error
	} else {
		err = db.CreateTable(&UserDailyRewards{}).Error
	}
	return err
}
