package models

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"
)

//！ 游戏每日统计数据
type RecordGameDay struct {
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

func (RecordGameDay) TableName() string {
	return "house_day_record"
}

func initRecordGameDay(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameDay{}) {
		err = db.AutoMigrate(&RecordGameDay{}).Error
	} else {
		err = db.CreateTable(&RecordGameDay{}).Error
	}
	return err
}

// db -> redis模型
func (u *RecordGameDay) ConvertModel() *RecordGameDayBak {
	p := new(RecordGameDayBak)
	p.Id = u.Id
	p.DHId = u.DHId
	p.FId = u.FId
	p.DFId = u.DFId
	p.UId = u.UId
	p.PlayTimes = u.PlayTimes
	p.BwTimes = u.BwTimes
	p.TotalScore = u.TotalScore
	p.ValidTimes = u.ValidTimes
	p.PlayDate = u.PlayDate
	p.CreatedAt = u.CreatedAt
	p.UpdatedAt = u.UpdatedAt
	p.BigValidTimes = u.BigValidTimes
	p.Partner = u.Partner
	p.SuperiorId = u.SuperiorId
	p.Radix = u.Radix
	return p
}

// 解析分数
func (u *RecordGameDay) GetRealScore() float64 {
	if u.Radix == 0 {
		return static.HF_DecimalDivide(float64(u.TotalScore/1), 1, 2)
	}
	return static.HF_DecimalDivide(float64(u.TotalScore)/float64(u.Radix), 1, 2)
}
