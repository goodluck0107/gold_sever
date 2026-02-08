package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

// 用户表
type GoldUser struct {
	Id          int64     `gorm:"primary_key;column:id"`
	LastLoginAt time.Time `gorm:"column:last_login_at;type:datetime;default:Now()"` // 最后一次登录时间
	TotalCount  int       `gorm:"column:total_count;default:0"`                     // 总局数
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"`                  // 创建时间
}

func (GoldUser) TableName() string {
	return "gold_user"
}

func initGoldUser(db *gorm.DB) error {
	var err error
	if db.HasTable(&GoldUser{}) {
		err = db.AutoMigrate(&GoldUser{}).Error
	} else {
		err = db.CreateTable(&GoldUser{}).Error

		db.Model(GoldUserLoginRecord{}).AddIndex("idx_created_at", "created_at")
	}
	return err
}

//! 添加金币场用户
func GoldUserAdd(db *gorm.DB, uid int64, time time.Time) error {
	return db.Create(&GoldUser{
		Id:          uid,
		CreatedAt:   time,
		TotalCount:  0,
		LastLoginAt: time,
	}).Error
}

//! 更新金币场用户信息
func GoldUserSave(db *gorm.DB, gUser GoldUser) error {
	return db.Save(&gUser).Error
}

//! 获取金币场玩家信息
func GoldUserGet(db *gorm.DB, uid int64) (GoldUser, error) {
	var gUser GoldUser
	err := db.Model(GoldUser{}).Where("id = ?", uid).First(&gUser).Error
	return gUser, err
}

//! 更新玩家对局数
func GoldUserAddRound(db *gorm.DB, uid int64) error {
	return db.Model(GoldUser{}).Where("id = ?", uid).Update("total_count", gorm.Expr("total_count + 1")).Error
}

//! 查询新增玩家总数
func GoldNewUserGetCountByTime(db *gorm.DB, time time.Time) (int, error) {
	var count = 0
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(GoldUser{}).Where(`date_format(created_at, "%Y-%m-%d") = ?`, selectStr).Count(&count).Error
	return count, err
}

//! 查询金币场时间节点玩家总数
func GoldUserGetCountByTime(db *gorm.DB, time time.Time) (int, error) {
	var count = 0
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(GoldUser{}).Where(`date_format(created_at, "%Y-%m-%d") <= ?`, selectStr).Count(&count).Error
	return count, err
}
