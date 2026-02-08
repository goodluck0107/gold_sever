package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type HouseRevokeOptKind int

const (
	HouseRevokeOptBack       HouseRevokeOptKind = 0 // 带回小圈
	HouseRevokeOptBackAndUpd HouseRevokeOptKind = 1 // 带回小圈 并不在大圈删除 更新其层级关系
)

type HouseRevokeLog struct {
	Id           int64              `gorm:"primary_key;column:id"`
	Swallowed    int64              `gorm:"column:swallowed;default:0;comment:'原来的/被合并圈DHID'"`
	Devourer     int64              `gorm:"column:devourer;default:0;comment:'新的/合并圈DHID'"`
	Uid          int64              `gorm:"column:uid;comment:'被带走的玩家id'"`
	OptKind      HouseRevokeOptKind `gorm:"column:opt_kind;comment:'操作类型0带回老圈1重复用户更新层级关系和ref'"`
	URole        int                `gorm:"column:urole;comment:'被带走的玩家的角色'"`
	Ref          int64              `gorm:"column:ref;comment:'被带走的玩家的ref'"`
	Partner      int64              `gorm:"column:partner;comment:'被带走的玩家的合伙人标识'"`
	VicePartner  bool               `gorm:"column:vice_partner;comment:'被带走的玩家的副队长标识'"`
	VitaminAdmin bool               `gorm:"column:vitamin_admin;comment:'被带走的玩家的裁判标识'"`
	Superior     int64              `gorm:"column:superior;comment:'被带走的玩家的上级标识'"`
	CreatedAt    time.Time          `gorm:"column:created_at;type:datetime"`
}

func (HouseRevokeLog) TableName() string {
	return "house_revoke_log"
}

func initHouseRevokeLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseRevokeLog{}) {
		err = db.AutoMigrate(&HouseRevokeLog{}).Error
	} else {
		err = db.CreateTable(&HouseRevokeLog{}).Error
	}
	return err
}
