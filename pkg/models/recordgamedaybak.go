package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 游戏每日统计数据
type RecordGameDayBak struct {
	Id            int       `gorm:"primary_key;column:id"`            //! id
	DHId          int64     `gorm:"column:hid"`                       //! 包厢dhid
	FId           int64     `gorm:"column:fid"`                       //! 包厢dfid
	DFId          int64     `gorm:"column:dfid"`                      //! 包厢dfid
	UId           int64     `gorm:"column:uid"`                       //! 包厢玩家id
	PlayTimes     int       `gorm:"column:play_times"`                //! 当天总局数
	BwTimes       int       `gorm:"column:bw_times"`                  //! 当天大赢家次数
	TotalScore    int       `gorm:"column:total_score"`               //! 当天累计积分
	ValidTimes    int       `gorm:"column:valid_times"`               //! 当天累计有效局数
	BigValidTimes int       `gorm:"column:big_valid_times;default:0"` //! 当天累计超级有效局数
	PlayDate      time.Time `gorm:"column:playdate;type:date"`        //! 当天游戏时间
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime"`  //! 创建时间
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime"`  //! 更新时间
	Partner       int64     `gorm:"column:partner;default:0"`         //! 队长归属
	SuperiorId    int64     `gorm:"column:superiorid;default:0"`      //! 上级收益
	Radix         int       `gorm:"column:radix;default:1"`           //! 算分基数,默认100
}

func (RecordGameDayBak) TableName() string {
	return "house_day_record_bak"
}

func initRecordGameDayBak(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameDayBak{}) {
		err = db.AutoMigrate(&RecordGameDayBak{}).Error
	} else {
		err = db.CreateTable(&RecordGameDayBak{}).Error
	}
	return err
}
