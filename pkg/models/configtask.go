package models

import (
	"github.com/jinzhu/gorm"
)

// 任务配置表
type ConfigTask struct {
	Id             int    `gorm:"primary_key;column:id"`         // id
	MainType       int    `gorm:"column:main_type"`              // 任务主类型 0 每日任务 1 系统任务
	SubType        int    `gorm:"column:sub_type"`               // 任务子类型 分享任务 对局次数 胜场次数
	Area           string `gorm:"column:area;default:''"`        // 任务是否绑定区域 默认为空字符串不绑定区域
	Sort           int    `gorm:"column:sort"`                   // 任务排序的标记
	TgtCompleteNum int    `gorm:"column:tgt_complete_num"`       // 任务目标达成所需的数量
	Reward         string `gorm:"column:reward"`                 // 任务奖励
	GameKindId     int    `gorm:"column:game_kind_id"`           // 任务与指定游戏的kind id
	Desc           string `gorm:"column:desc"`                   // 任务描述
	RewardDesc     string `gorm:"column:reward_desc"`            // 奖励描述
	StepTaskId     int    `gorm:"column:step_task_id;default:0"` // 阶段任务表id
}

func (ConfigTask) TableName() string {
	return "config_task"
}

func initConfigTask(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigTask{}) {
		err = db.AutoMigrate(&ConfigTask{}).Error
	} else {
		err = db.CreateTable(&ConfigTask{}).Error
	}
	return err
}

// 游戏内显示的任务配置表
type ConfigTaskGame struct {
	Id         int    `gorm:"primary_key;column:id"`     //! id
	Type       int    `gorm:"column:type"`               //! 任务类型
	RewardType int    `gorm:"column:rewardtype"`         //! 奖励类型
	Kind       int    `gorm:"column:kind"`               //! 任务种类：对战，胜利，分享
	Num        int    `gorm:"column:num"`                //! 任务完成数量
	StepNum    int    `gorm:"column:stepnum"`            //!
	KindId     int    `gorm:"column:kindid"`             //! 对战，胜利相关任务的游戏id
	SiteType   int    `gorm:"column:sitetype;default:0"` //! 场次类型 0表示所有场次
	Value      string `gorm:"column:v;size:255"`         //! 任务扩展数据
	Flag       int    `gorm:"column:flag;default:0"`     //! 0表示关闭，1表示开启
	Text       string `gorm:"column:text;size:255"`      //! 描述性文字说明
}

func (ConfigTaskGame) TableName() string {
	return "config_task_game"
}

func initConfigTaskGame(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigTaskGame{}) {
		err = db.AutoMigrate(&ConfigTaskGame{}).Error
	} else {
		err = db.CreateTable(&ConfigTaskGame{}).Error
	}
	return err
}
