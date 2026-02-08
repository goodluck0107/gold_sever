package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type RecordGameCostMini struct {
	HId      int64     `gorm:"column:hid"`                      //! 包厢id
	PlayTime int       `gorm:"column:playtime"`                 //! 对局次数
	KaCost   int       `gorm:"column:kacost"`                   //! 消耗
	Date     time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

//! 游戏总结算记录
type RecordGameCost struct {
	Id         int64     `gorm:"primary_key;column:id"`           //! id
	UId        int64     `gorm:"column:uid"`                      //! 玩家用户ID
	HId        int64     `gorm:"column:hid"`                      //! 包厢id(数据库自增id)
	TId        int       `gorm:"column:tid"`                      //! 牌桌id
	FId        int64     `gorm:"column:fid"`                      //! 楼层id
	NTId       int       `gorm:"column:ntid"`                     //! 牌桌索引
	BefKa      int       `gorm:"column:befka"`                    //! 消耗前
	AftKa      int       `gorm:"column:aftka"`                    //! 消耗后
	KaCost     int       `gorm:"column:kacost"`                   //! 消耗
	KindId     int       `gorm:"column:kindid"`                   //! 游戏类型
	Gamenum    string    `gorm:"column:game_num;size:100"`        //! 牌局编号
	GameConfig string    `gorm:"column:game_config;type:text"`    //! 游戏配置
	LeagueID   int64     `gorm:"column:league;default:0"`         // 加盟商id
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	PlayerNum  int       `gorm:"column:playernum"`                // 游戏人数
}

func (RecordGameCost) TableName() string {
	return "record_game_cost"
}

func initRecordGameCost(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameCost{}) {
		err = db.AutoMigrate(&RecordGameCost{}).Error
	} else {
		err = db.CreateTable(&RecordGameCost{}).Error
	}
	return err
}

// db -> redis模型
func (u *RecordGameCost) ConvertModel() *RecordGameCostBak {
	p := new(RecordGameCostBak)
	p.Id = u.Id
	p.UId = u.UId
	p.HId = u.HId
	p.TId = u.TId
	p.FId = u.FId
	p.NTId = u.NTId
	p.BefKa = u.BefKa
	p.AftKa = u.AftKa
	p.KaCost = u.KaCost
	p.KindId = u.KindId
	p.Gamenum = u.Gamenum
	p.GameConfig = u.GameConfig
	p.LeagueID = u.LeagueID
	p.CreatedAt = u.CreatedAt
	p.PlayerNum = u.PlayerNum
	return p
}
