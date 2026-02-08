package models

import (
	"github.com/open-source/game/chess.git/pkg/static"
	"time"

	"github.com/jinzhu/gorm"
)

type HouseActivity struct {
	Id      int64     `gorm:"primary_key;column:id"`        //! id
	DHId    int64     `gorm:"column:hid"`                   //! 包厢id
	FId     string    `gorm:"column:fid"`                   //! 楼层id
	Kind    int       `gorm:"column:kind"`                  //! 类型
	Name    string    `gorm:"column:name;type:text"`        //! 名称
	Status  int       `gorm:"column:status"`                //! 是否活跃
	BegTime time.Time `gorm:"column:begtime;type:datetime"` // 开始时间
	EndTime time.Time `gorm:"column:endtime;type:datetime"` // 结束时间

	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
	HideInfo    bool      `gorm:"column:hideinfo"`                 //! 是否隐藏信息
	Type        int64     `gorm:"column:type"`                     //! 活动类型，是否为幸运星
	TicketCount int64     `gorm:"column:t_count"`                  //几局抽一次
}

func (HouseActivity) TableName() string {
	return "house_activity"
}

// db -> redis模型
func (u *HouseActivity) ConvertModel() *static.HouseActivity {
	p := new(static.HouseActivity)
	p.Id = u.Id
	p.FId = u.FId
	p.DHId = u.DHId
	p.Kind = u.Kind
	p.Status = u.Status
	p.Name = u.Name
	p.HideInfo = u.HideInfo
	p.BegTime = u.BegTime.Unix()
	p.EndTime = u.EndTime.Unix()

	return p
}

func InitHouseActivity(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseActivity{}) {
		err = db.AutoMigrate(&HouseActivity{}).Error
	} else {
		err = db.CreateTable(&HouseActivity{}).Error
	}
	return err
}

type HouseActivityLog struct {
	Id      int64     `gorm:"primary_key;column:id"`        //! id
	ActId   int64     `gorm:"column:actid"`                 //! 活动id
	DHId    int64     `gorm:"column:hid"`                   //! 包厢id
	FId     string    `gorm:"column:fid"`                   //! 楼层id
	Kind    int       `gorm:"column:kind"`                  //! 类型
	Name    string    `gorm:"column:name;type:text"`        //! 名称
	Status  int       `gorm:"column:status"`                //! 是否活跃
	BegTime time.Time `gorm:"column:begtime;type:datetime"` // 开始时间
	EndTime time.Time `gorm:"column:endtime;type:datetime"` // 结束时间

	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseActivityLog) TableName() string {
	return "house_activity_log"
}

func InitHouseActivityLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseActivityLog{}) {
		err = db.AutoMigrate(&HouseActivityLog{}).Error
	} else {
		err = db.CreateTable(&HouseActivityLog{}).Error
	}
	return err
}
