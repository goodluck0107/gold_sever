package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type UserLoginRecord struct {
	Id          int64     `gorm:"primary_key;column:id"`
	Uid         int64     `gorm:"column:uid"`                      // 玩家id
	Platform    int       `gorm:"column:platform"`                 // 平台
	Ip          string    `gorm:"column:ip;size:32"`               // ip
	MachineCode string    `gorm:"column:machine_code;size:32"`     // 机器码
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` // 创建时间(登录时间)
}

func (UserLoginRecord) TableName() string {
	return "user_login_record"
}

func initUserLoginRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserLoginRecord{}) {
		err = db.AutoMigrate(&UserLoginRecord{}).Error
	} else {
		err = db.CreateTable(&UserLoginRecord{}).Error
	}
	return err
}

//! 获取当日活跃用户数量
func GetFangKaUserActiveCount(db *gorm.DB, time time.Time) (int, error) {
	var count int
	selectDayStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(UserLoginRecord{}).Where(`date_format(created_at, "%Y-%m-%d") = ?`, selectDayStr).Group("uid").Count(&count).Error
	return count, err
}

//! 用户登录留存查询
func GetFangKaUserRetentionStatistics(db *gorm.DB, time time.Time) (int, int, int, int) {
	selectDayStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	selectDay1 := time.AddDate(0, 0, 1)
	selectDay1Str := fmt.Sprintf("%d-%02d-%02d", selectDay1.Year(), selectDay1.Month(), selectDay1.Day())
	selectDay3 := time.AddDate(0, 0, 3)
	selectDay3Str := fmt.Sprintf("%d-%02d-%02d", selectDay3.Year(), selectDay3.Month(), selectDay3.Day())
	selectDay7 := time.AddDate(0, 0, 7)
	selectDay7Str := fmt.Sprintf("%d-%02d-%02d", selectDay7.Year(), selectDay7.Month(), selectDay7.Day())

	var uids []int64

	sql := `select id as uid from user where date_format(created_at, "%Y-%m-%d") = ?`
	if err := db.Raw(sql, selectDayStr).Pluck("uid", &uids).Error; err != nil {
		return 0, 0, 0, 0
	}

	if len(uids) == 0 {
		return 0, 0, 0, 0
	}

	type QueryResult struct {
		Count    int    `gorm:"column:count"`            //! 玩家Id
		GameDate string `gorm:"column:gamedate;size:12"` //! 玩游戏时间
	}

	var results []QueryResult
	sql = `select cast(created_at as date) as gamedate, count(DISTINCT uid) as count from user_login_record where (date_format(created_at, "%Y-%m-%d") = ? or date_format(created_at, "%Y-%m-%d") = ? or date_format(created_at, "%Y-%m-%d") = ?) and uid in (?) group by cast(created_at as date)`
	if err := db.Raw(sql, selectDay1Str, selectDay3Str, selectDay7Str, uids).Scan(&results).Error; err != nil {
		return 0, 0, 0, 0
	}

	userRetention := len(uids)
	userRetention1 := 0
	userRetention3 := 0
	userRetention7 := 0
	for _, result := range results {
		if result.GameDate[0:10] == selectDay1Str {
			userRetention1 = result.Count
		} else if result.GameDate[0:10] == selectDay3Str {
			userRetention3 = result.Count
		} else if result.GameDate[0:10] == selectDay7Str {
			userRetention7 = result.Count
		}
	}
	return userRetention, userRetention1, userRetention3, userRetention7
}
