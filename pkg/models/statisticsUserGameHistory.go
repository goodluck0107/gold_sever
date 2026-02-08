package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 用户相关历史游戏记录统计表
type StatisticsUserGameHistory struct {
	Id        int64     `gorm:"primary_key;column:id;AUTO_INCREMENT" json:"id"`     // id
	Uid       int64     `gorm:"column:uid;index:idx_uid2kindId" json:"uid"`         // 玩家uid
	KindId    int       `gorm:"column:kind_id;index:idx_uid2kindId" json:"kind_id"` // 游戏id
	PlayTimes int       `gorm:"column:play_times" json:"play_times"`                // 游戏次数
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime" json:"updated_at"`  // 更新时间
}

func (StatisticsUserGameHistory) TableName() string {
	return "statistics_user_gamehistory"
}

func initStatisticsUserGameHistory(db *gorm.DB) error {
	var err error
	if db.HasTable(&StatisticsUserGameHistory{}) {
		err = db.AutoMigrate(&StatisticsUserGameHistory{}).Error
	} else {
		err = db.CreateTable(&StatisticsUserGameHistory{}).Error
	}
	return err
}

type UserGameHistoryList []*StatisticsUserGameHistory

func (self UserGameHistoryList) Len() int {
	return len(self)
}

func (self UserGameHistoryList) Less(i, j int) bool {
	if self[i].UpdatedAt.Unix() == self[j].UpdatedAt.Unix() {
		return self[i].PlayTimes > self[j].PlayTimes
	}
	return self[i].UpdatedAt.Unix() > self[j].UpdatedAt.Unix()
}

func (self UserGameHistoryList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
