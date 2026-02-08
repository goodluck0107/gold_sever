package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type GlodUserStatistics struct {
	Id               int64     `gorm:"primary_key;AUTO_INCREMENT;column:id"` // id
	Date             time.Time `gorm:"column:date;type:date"`                // 日期
	TotalCount       int       `gorm:"column:totalcount"`                    // 平台用户总数
	DayGameCount     int       `gorm:"column:daygamecount"`                  // 平台今天总的游戏玩家总数
	NewUserCount     int       `gorm:"column:newusercount"`                  // 当日新注册的帐号数量
	NewUserGameCount int       `gorm:"column:newusergamecount"`              // 当日新注册的帐号游戏数量
	NewUser1DayCount int       `gorm:"column:newuser1daycount"`              // 新增次日留存
	NewUser3DayCount int       `gorm:"column:newuser3daycount"`              // 新增三日留存
	NewUser7DayCount int       `gorm:"column:newuser7daycount"`              // 新增7日留存
	GameCount        int       `gorm:"column:gamecount"`                     // 当天游戏玩家数量
	Game1DayCount    int       `gorm:"column:game1daycount"`                 // 游戏次日留存
	Game3DayCount    int       `gorm:"column:game3daycount"`                 // 游戏三日留存
	Game7DayCount    int       `gorm:"column:game7daycount"`                 // 游戏七日留存
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime"`      // 创建时间
}

func (GlodUserStatistics) TableName() string {
	return "gold_user_statistics"
}

func initUserStatisticsGlod(db *gorm.DB) error {
	var err error
	if db.HasTable(&GlodUserStatistics{}) {
		err = db.AutoMigrate(&GlodUserStatistics{}).Error
	} else {
		err = db.CreateTable(&GlodUserStatistics{}).Error

		db.Model(GoldUserLoginRecord{}).AddIndex("idx_date", "date")
	}
	return err
}

func GlodUserStatisticsUpdateRetention(db *gorm.DB, time time.Time, userRetention, userRetention1, userRetention3, userRetention7, gameRetention, gameRetention1, gameRetention3, gameRetention7 int) error {
	updataAttr := make(map[string]interface{})
	updataAttr["newusercount"] = userRetention
	updataAttr["newuser1daycount"] = userRetention1
	updataAttr["newuser3daycount"] = userRetention3
	updataAttr["newuser7daycount"] = userRetention7
	updataAttr["gamecount"] = gameRetention
	updataAttr["game1daycount"] = gameRetention1
	updataAttr["game3daycount"] = gameRetention3
	updataAttr["game7daycount"] = gameRetention7

	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	return db.Model(GlodUserStatistics{}).Where(`date_format(date, "%Y-%m-%d") = ?`, selectStr).Update(updataAttr).Error
}
