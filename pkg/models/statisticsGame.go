package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//! 游戏相关统计表(每天统计一次)
type StatisticsGame struct {
	Id                int64     `gorm:"primary_key;AUTO_INCREMENT;column:id"` // id
	Date              time.Time `gorm:"column:date;type:date"`                // 日期
	KindId            int       `gorm:"column:kind_id"`                       //! 子游戏id
	SiteType          int       `gorm:"column:site_type"`                     //! 场次类型
	AvgPlayCount      int       `gorm:"column:avg_play_count"`                // 平均玩牌局数
	AvgPlayTime       int       `gorm:"column:avg_play_time"`                 // 平均玩牌时长(秒)
	AvgWinScore       int       `gorm:"column:avg_win_score;default:0"`       // 平均赢分
	AvgLoseScore      int       `gorm:"column:avg_lose_score;default:0"`      // 平均输分
	RobotAvgScore     int       `gorm:"column:robot_avg_score;default:0"`     // 机器人平均输分
	RobotDayCost      int       `gorm:"column:robot_day_cost;default:0"`      // 每日机器人具体输赢金币总数
	DayPlayUser       int       `gorm:"column:day_play_user;default:0"`       // 每日玩牌人数统计
	CreatedAt         time.Time `gorm:"column:created_at;type:datetime"`      // 创建时间
	RobotAvgPlayCount int       `gorm:"column:robot_avg_play_count"`          // 机器人平均玩牌局数
	//RobotAvgWinScore  int       `gorm:"column:robot_avg_winscore;default:0"`  // 机器人平均赢分
	//RobotAvgLoseScore int       `gorm:"column:robot_avg_losescore;default:0"` // 机器人平均输分
}

func (StatisticsGame) TableName() string {
	return "statistics_game"
}

func initStatisticsGame(db *gorm.DB) error {
	var err error
	if db.HasTable(&StatisticsGame{}) {
		err = db.AutoMigrate(&StatisticsGame{}).Error
	} else {
		err = db.CreateTable(&StatisticsGame{}).Error
	}
	return err
}
