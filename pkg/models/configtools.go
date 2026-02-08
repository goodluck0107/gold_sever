package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 兑换商道具配置表
type ConfigTool struct {
	ToolId int `gorm:"column:toolid"` //! 商品ID
	Price  int `gorm:"colunm:price"`  //! 花费
	Num    int `gorm:"column:num"`    //! 数量
}

func (ConfigTool) TableName() string {
	return "config_tool"
}

func initConfigTool(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigTool{}) {
		err = db.AutoMigrate(&ConfigTool{}).Error
	} else {
		err = db.CreateTable(&ConfigTool{}).Error
	}
	return err
}

//! 用户道具表
type UserTools struct {
	Id        int64     `gorm:"primary_key;column:id"`           // id
	Uid       int64     `gorm:"column:uid;index"`                // 玩家用户ID
	ToolType  int16     `gorm:"column:tool_type;index"`          // 道具类型
	Count     int64     `gorm:"column:count"`                    //! 道具数目
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	DeadAt    time.Time `gorm:"column:dead_at;type:datetime"`    // 失效时间
}

func (UserTools) TableName() string {
	return "user_tools"
}

func initUserTools(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserTools{}) {
		err = db.AutoMigrate(&UserTools{}).Error
	} else {
		err = db.CreateTable(&UserTools{}).Error
	}
	return err
}
