package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

// 玩家游戏内操作类型
type UserGameOptType uint16

const (
	UserGameOpeningHu UserGameOptType = iota + 1
	UserGameGangKai
	UserGameOpening4Magic
	UserGameHave4Magic
)

func (u UserGameOptType) String() (s string) {
	switch u {
	case UserGameOpeningHu:
		s = "起手胡次数"
	case UserGameGangKai:
		s = "杠开次数"
	case UserGameOpening4Magic:
		s = "起手4个赖子次数"
	case UserGameHave4Magic:
		s = "起到过4个赖子得次数"
	default:
		s = fmt.Sprintf("未知操作类型:%d", u)
	}
	return
}

type StatisticsUserGameOpt struct {
	Id        int64           `gorm:"primary_key;AUTO_INCREMENT;column:id"`            // id
	Uid       int64           `gorm:"column:uid"`                                      // 玩家id
	KindID    int             `gorm:"column:kindid"`                                   // 玩法id
	GameNum   string          `gorm:"column:gamenum"`                                  // 游戏编号（包含房间号）
	Times     int             `gorm:"column:times"`                                    // 操作次数
	Type      UserGameOptType `gorm:"column:type;comment:'1起手胡 2杠开 3起手4个赖子 4起到过4个赖子'"` // 操作类型
	CreatedAt time.Time       `gorm:"column:created_at;type:datetime"`                 // 创建时间，记录只会在游戏结束的时候创建，所以这个创建时间也代表了游戏结束时间
}

func (StatisticsUserGameOpt) TableName() string {
	return "statistics_user_game_opt"
}

func initStatisticsUserGameOpt(db *gorm.DB) error {
	var err error
	if db.HasTable(&StatisticsUserGameOpt{}) {
		err = db.AutoMigrate(&StatisticsUserGameOpt{}).Error
	} else {
		err = db.CreateTable(&StatisticsUserGameOpt{}).Error
	}
	return err
}
