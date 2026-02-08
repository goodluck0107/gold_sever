package models

import (
	"github.com/jinzhu/gorm"
)

//! 游戏单局总结算记录
type RecordDismiss struct {
	Id          int64  `gorm:"column:id"`           //! id
	GameNum     string `gorm:"column:game_num"`     //! 游戏ID,唯一标识
	DismissTime string `gorm:"column:dismiss_time"` //! 解散时间
	DismissType int    `gorm:"column:dismiss_type"` //! 解散类型   0：无,  1：盟主解散，2：管理员解散 3：队长解散 4：申请解散 5：超时解散 6：托管解散 7:离线解散 8：出牌超时解散
	DismissDet  string `gorm:"column:dismiss_det"`  //! 详情
}

func (RecordDismiss) TableName() string {
	return "record_dismiss"
}

func initRecordDismiss(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordDismiss{}) {
		err = db.AutoMigrate(&RecordDismiss{}).Error
	} else {
		err = db.CreateTable(&RecordDismiss{}).Error
		db.Model(&RecordDismiss{}).AddIndex("ids_gamenum", "game_num")
	}
	return err
}
