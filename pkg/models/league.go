package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

//! 加盟商
type League struct {
	Id         int64     `gorm:"primary_key;column:id"` // id
	LeagueID   int64     `gorm:"column:league_id"`      // 加盟商id
	Card       int64     `gorm:"column:card"`           // 房卡数
	FreezeCard int64     `gorm:"column:freeze_card"`    // 房卡数
	AreaCode   int64     `gorm:"column:area_code"`      // 区域编号
	Freeze     bool      `gorm:"column:freeze"`         // 是否冻结
	PoolState  bool      `gorm:"column:pool_state;comment:'卡池是否开启，1为开启，0关闭'"`
	UserNum    int64     `gorm:"column:user_num"` // 是否冻结
	PoolStart  time.Time `gorm:"column:pool_start;comment:'卡池定时开启时间'"`
	PoolEnd    time.Time `gorm:"column:pool_end;comment:'卡池定时关闭时间'"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"`  // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:datetime"` // 更新时间
}

func (League) TableName() string {
	return "league"
}

func initLeague(db *gorm.DB) error {
	var err error
	if db.HasTable(&League{}) {
		err = db.AutoMigrate(&League{}).Error
	} else {
		err = db.CreateTable(&League{}).Error
	}
	return err
}

type LeagueUser struct {
	Id         int64     `gorm:"primary_key;column:id"` // id
	LeagueID   int64     `gorm:"column:league_id"`      // 加盟商id
	Uid        int64     `gorm:"column:uid"`            // 用户id
	Freeze     bool      `gorm:"column:freeze"`         // 是否冻结
	PoolState  bool      `gorm:"column:pool_state;comment:'卡池是否开启，1为开启，0关闭'"`
	PoolStart  time.Time `gorm:"column:pool_start;comment:'卡池定时开启时间'"`
	PoolEnd    time.Time `gorm:"column:pool_end;comment:'卡池定时关闭时间'"`
	Card       int64     `gorm:"column:card;default:0"`            // 房卡数
	FreezeCard int64     `gorm:"column:freeze_card;default:0"`     // 冻结房卡
	UsedCard   int64     `gorm:"column:used_card;default:0"`       // 已使用房卡
	NotPool    bool      `gorm:"column:not_pool;default:false"`    //是否不使用加盟商卡池的卡
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"`  // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:datetime"` // 更新时间
}

func (LeagueUser) TableName() string {
	return "league_user"
}

func initLeagueUser(db *gorm.DB) error {
	var err error
	if db.HasTable(&LeagueUser{}) {
		err = db.AutoMigrate(&LeagueUser{}).Error
	} else {
		err = db.CreateTable(&LeagueUser{}).Error
	}
	return err
}

type LeagueCardRecord struct {
	Id         int64     `gorm:"primary_key;column:id"`  // id
	LeagueID   int64     `gorm:"column:league_id"`       // 加盟商id
	Uid        int64     `gorm:"column:uid"`             // 用户id
	GameID     string    `gorm:"column:game_id;size:30"` // 游戏id
	Cost       int64     `gorm:"column:cost"`
	WealthType int64     `gorm:"column:wealth_type"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (LeagueCardRecord) TableName() string {
	return "league_card_record"
}
func initLeagueCardRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&LeagueUser{}) {
		err = db.AutoMigrate(&LeagueCardRecord{}).Error
	} else {
		err = db.CreateTable(&LeagueCardRecord{}).Error
	}
	return err
}
