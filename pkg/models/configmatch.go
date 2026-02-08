package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// 任务配置表
type ConfigMatch struct {
	Id         int                    `gorm:"primary_key;column:id"`            //! id
	Name       string                 `gorm:"column:name;size:50"`              // 场次名
	KindId     int                    `gorm:"column:kind_id"`                   //! 子游戏id
	TypeStr    string                 `gorm:"column:typestr"`                   //! 类型(初级场 中级场 高级场)
	ConfigStr  string                 `gorm:"column:match_config;default:'{}'"` //! 游戏参数配置
	State      int                    `gorm:"column:state;default:0"`           //!状态，0未开始，1进行中，2已经结束
	Flag       int                    `gorm:"column:flag;default:0"`            //!是否开启，0未开启，1开启
	BeginDate  time.Time              `gorm:"column:begindate;type:date"`       //! 开始日期
	EndDate    time.Time              `gorm:"column:enddate;type:date"`         //! 结束日期
	BeginTime  time.Time              `gorm:"column:begintime;type:datetime"`   //! 开始时间，解析后只取time
	EndTime    time.Time              `gorm:"column:endtime;type:datetime"`     //! 结束时间，解析后只取time
	CreatedAt  time.Time              `gorm:"column:created_at;type:datetime"`  //! 创建时间
	ShareStr   string                 `gorm:"column:sharestr;default:'{}'"`     //! 分享图片或文案
	ShareWxStr string                 `gorm:"column:sharewxstr;default:'{}'"`   //! 分享图片或文案
	ShareTitle string                 `gorm:"column:sharetitle;default:'{}'"`   //! 分享标题
	Types      []int                  `gorm:"-"`                                // 对应Type配置
	Config     map[string]interface{} `gorm:"-"`                                //对应ConfigStr配置                                         //对应ConfigStr配置
}

func (ConfigMatch) TableName() string {
	return "config_match"
}

func initConfigMatch(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigMatch{}) {
		err = db.AutoMigrate(&ConfigMatch{}).Error
	} else {
		err = db.CreateTable(&ConfigMatch{}).Error
	}
	return err
}
