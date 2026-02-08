package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

// 政府实名认证系统

// 实名认证配置
type ConfigGovAuth struct {
	LimitTimeUnder18         int    `gorm:"column:limit_time_under18"`           // 未成年游戏时长限制(秒为单位)
	LimitTimeNotAuth         int    `gorm:"column:limit_time_not_auth"`          // 未实名游戏时长限制
	IsForceNotice            bool   `gorm:"column:is_force_notice"`              // 是否强制弹窗提示
	LimitTimeAtBeforeUnder18 string `gorm:"column:limit_time_at_before_under18"` // 未成年游戏时间限制 不得在xx:xx之前登陆游戏
	LimitTimeAtAfterUnder18  string `gorm:"column:limit_time_at_after_under18"`  // 未成年游戏时间限制 不得在xx:xx之后登陆游戏
}

func (ConfigGovAuth) TableName() string {
	return "config_government_auth"
}

func initConfigGovAuth(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigGovAuth{}) {
		err = db.AutoMigrate(&ConfigGovAuth{}).Error
	} else {
		err = db.CreateTable(&ConfigGovAuth{}).Error
	}
	return err
}

// 用户的每日游戏时长
type UserDailyPlayTime struct {
	Id        int64     `gorm:"primary_key;column:id"`           // 表id
	Uid       int64     `gorm:"column:uid"`                      // 分享玩家uid
	PlayTime  int       `gorm:"column:play_time"`                // 游戏时长
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (UserDailyPlayTime) TableName() string {
	return "user_daily_play_time"
}

func initDailyPlayTime(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserDailyPlayTime{}) {
		err = db.AutoMigrate(&UserDailyPlayTime{}).Error
	} else {
		err = db.CreateTable(&UserDailyPlayTime{}).Error
		db.Model(UserDailyPlayTime{}).AddIndex("idx_uid_createdAt", "uid", "created_at")
	}
	return err
}

// 更新时长
func UpdateUserPlayTime(db *gorm.DB, uid int64, time time.Time, playTime int) (int, error) {
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	var ret UserDailyPlayTime
	err := db.Model(UserDailyPlayTime{}).Where(`uid = ? and date_format(created_at, "%Y-%m-%d") = ?`, uid, selectStr).First(&ret).Error
	if err == gorm.ErrRecordNotFound {
		// 创建记录
		newRecord := new(UserDailyPlayTime)
		newRecord.Uid = uid
		newRecord.PlayTime = playTime
		// 首次上报时长大于上报时间间隔（兼容跨天）
		if playTime > 5*60 {
			newRecord.PlayTime = 5 * 60
		}
		newRecord.CreatedAt = time
		if err = db.Create(&newRecord).Error; err != nil {
			xlog.Logger().Errorln(err)
		}
		return newRecord.PlayTime, err
	} else {
		// 更新记录
		newRecord := make(map[string]interface{})
		newRecord["play_time"] = playTime
		newRecord["created_at"] = time
		if err = db.Model(UserDailyPlayTime{}).Where(`uid = ? and date_format(created_at, "%Y-%m-%d") = ?`, uid, selectStr).Update(newRecord).Error; err != nil {
			xlog.Logger().Errorln(err)
		}
		return playTime, err
	}
}

// 获取累计时长
func GetUserTotalPlayTime(db *gorm.DB, uid int64) (int, error) {
	type queryResult struct {
		TotalTime int `gorm:"column:total_time"` // 总时长
	}
	ret := new(queryResult)
	err := db.Model(UserDailyPlayTime{}).Select("sum(play_time) as total_time").Where("uid = ?", uid).Scan(&ret).Error
	if err != nil {
		return 0, err
	}
	return ret.TotalTime, err
}
