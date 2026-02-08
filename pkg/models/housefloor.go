package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type HouseFloor struct {
	Id                   int64     `gorm:"primary_key;column:id" json:"id"` // id
	Name                 string    `gorm:"column:name;default:'';comment:'用户自定义名称'" json:"name"`
	IsMix                bool      `gorm:"column:is_mix;comment:'是否是混排'" json:"is_mix"`
	IsVip                bool      `gorm:"column:is_vip;comment:'是否是vip楼层'" json:"is_vip"`
	DHId                 int64     `gorm:"column:hid" json:"dhid"`                                            // 包厢id
	Rule                 string    `gorm:"column:rule;type:text" json:"rule"`                                 // 楼层规则
	IsVitamin            bool      `gorm:"column:is_vitamin;comment:'防沉迷开关'" json:"is_vitamin"`               // 防沉迷
	IsGamePause          bool      `gorm:"column:is_game_pause;comment:'中途暂停开关'" json:"is_game_pause"`        // 防沉迷队长可见
	VitaminLowLimit      int64     `gorm:"column:vitamin_lowerlimit;comment:'入桌下限'" json:"vitamin_low_limit"` // 疲劳值下限
	VitaminHighLimit     int64     `gorm:"column:vitamin_highlimit;comment:'入桌上限'" json:"vitamin_high_limit"`
	VitaminLowLimitPause int64     `gorm:"column:vitamin_lowerlimit_pause;comment:'暂停下限'" json:"vitamin_low_limit_pause"` // 疲劳值下限
	AiSuperNum           int       `gorm:"column:ai_super_num;comment:'超级防作弊人数'" json:"ai_super_num"`
	IsHide               bool      `gorm:"column:ishide;comment:'超级防作弊人数'" json:"ishide"` // 超级防作弊最大等待人数
	CreatedAt            time.Time `gorm:"column:created_at;type:datetime" json:"-"`      // 创建时间
	UpdatedAt            time.Time `gorm:"column:updated_at;type:datetime" json:"-"`      // 更新时间
	IsCapSetVip          bool      `gorm:"column:is_cap_set_vip;comment:' 队长是否可以设置该vip楼层'" json:"is_cap_set_vip"`
	IsDefJoinVip         bool      `gorm:"column:is_def_join_vip;comment:' 新入圈用户默认加入VIP楼层'" json:"is_def_join_vip"`
	MinTable             int       `gorm:"column:min_table" json:"min_table"`
	MaxTable             int       `gorm:"column:max_table" json:"max_table"`
}

func (HouseFloor) TableName() string {
	return "house_floor"
}

func initHouseFloor(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseFloor{}) {
		err = db.AutoMigrate(&HouseFloor{}).Error
	} else {
		err = db.CreateTable(&HouseFloor{}).Error
	}
	return err
}

type HousemixfloorTable struct {
	Id        int64     `gorm:"primary_key;column:id"`                                                                  //! id
	Hid       int64     `gorm:"column:hid"`                                                                             //! 包厢号
	Fid       int64     `gorm:"column:fid"`                                                                             //! 楼层号
	NTID      int64     `gorm:"column:ntid;not null"`                                                                   //桌子号
	HftInfo   string    `gorm:"column:hft_info"`                                                                        //桌子信息
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;DEFAULT :CURRENT_TIMESTAMP"`                             // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;DEFAULT:CURRENT_TIMESTAMP ;ON_UPDATE:CURRENT_TIMESTAMP"` // 更新时间
}

func (HousemixfloorTable) TableName() string {
	return "house_floor_mixtable"
}
func initHousemixfloorTable(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseFloor{}) {
		err = db.AutoMigrate(&HousemixfloorTable{}).Error
	} else {
		err = db.CreateTable(&HousemixfloorTable{}).Error
	}
	return err
}

func (u *HouseFloor) ConvertDeductModel() *HouseFloorVitaminDeduct {
	return &HouseFloorVitaminDeduct{
		Id:        u.Id,
		DHId:      u.DHId,
		Rule:      u.Rule,
		UpdatedAt: u.UpdatedAt,
	}
}
