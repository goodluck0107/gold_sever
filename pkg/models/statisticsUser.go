package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

//! 用户相关统计表(每天统计一次)
type StatisticsUser struct {
	Id               int64     `gorm:"primary_key;AUTO_INCREMENT;column:id"`         // id
	Date             time.Time `gorm:"column:date;type:date"`                        // 日期
	TotalCount       int       `gorm:"column:total_count;comment:'平台玩家人数'"`          // 平台用户总数
	NewCount         int       `gorm:"column:new_count;comment:'当天新增玩家人数'"`          // 当日新注册的帐号数量
	ActiveCount      int       `gorm:"active_count;comment:'当天活跃玩家人数'"`              // 当日登录过游戏的帐号数
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime"`              // 创建时间
	NewUser1DayCount int       `gorm:"column:newuser1daycount;comment:'新增用户次日留存人数'"` // 新增次日留存
	NewUser3DayCount int       `gorm:"column:newuser3daycount;comment:'新增用三次日留存人数'"` // 新增三日留存
	NewUser7DayCount int       `gorm:"column:newuser7daycount;comment:'新增用七次日留存人数'"` // 新增7日留存
	GameCount        int       `gorm:"column:gamecount;comment:'当天游戏留存人数'"`          // 当天游戏玩家数量
	Game1DayCount    int       `gorm:"column:game1daycount;comment:'次日游戏留存人数'"`      // 游戏次日留存
	Game3DayCount    int       `gorm:"column:game3daycount;comment:'三日游戏留存人数'"`      // 游戏三日留存
	Game7DayCount    int       `gorm:"column:game7daycount;comment:'七日游戏留存人数'"`      // 游戏七日留存
	AndroidNewUser   int       `gorm:"column:android_new_user;comment:'安卓新增用户'"`     // 安卓新增用户
	IOSNewUser       int       `gorm:"column:ios_new_user;comment:'ios新增用户'"`        // ios新增用户
	GoldNewUser      int       `gorm:"column:gold_new_user;comment:'金币场新增用户'"`       // 金币场新增用户
	FangKaNewUser    int       `gorm:"column:fangka_new_user;comment:'房卡场新增用户'"`     // 房卡场新增用户
}

func (StatisticsUser) TableName() string {
	return "statistics_user"
}

func initStatisticsUser(db *gorm.DB) error {
	var err error
	if db.HasTable(&StatisticsUser{}) {
		err = db.AutoMigrate(&StatisticsUser{}).Error
	} else {
		err = db.CreateTable(&StatisticsUser{}).Error
	}
	return err
}

func FangKaUserStatisticsUpdateRetention(db *gorm.DB, time time.Time, userRetention, userRetention1, userRetention3, userRetention7, gameRetention, gameRetention1, gameRetention3, gameRetention7 int) error {
	updataAttr := make(map[string]interface{})
	updataAttr["new_count"] = userRetention
	updataAttr["newuser1daycount"] = userRetention1
	updataAttr["newuser3daycount"] = userRetention3
	updataAttr["newuser7daycount"] = userRetention7
	updataAttr["gamecount"] = gameRetention
	updataAttr["game1daycount"] = gameRetention1
	updataAttr["game3daycount"] = gameRetention3
	updataAttr["game7daycount"] = gameRetention7

	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	return db.Model(StatisticsUser{}).Where(`date_format(date, "%Y-%m-%d") = ?`, selectStr).Update(updataAttr).Error
}
