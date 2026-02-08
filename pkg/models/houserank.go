package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type QueryHouseRankResult struct {
	Uid        int64   `gorm:"column:uid"`         //! 玩家ID
	RankRound  int     `gorm:"column:rank_round"`  //! 局数排行榜
	RankWiner  int     `gorm:"column:rank_winer"`  //! 大赢家排行榜
	RankRecord float64 `gorm:"column:rank_record"` //! 战绩总分排行榜
}

//包厢排行榜数据
type HouseRank struct {
	Id         int       `gorm:"primary_key;column:id"`
	Dhid       int       `gorm:"column:dhid"`                 //! 包厢唯一ID
	Uid        int       `gorm:"column:uid"`                  //! 玩家ID
	RankRound  int       `gorm:"column:rank_round"`           //! 局数排行榜
	RankWiner  int       `gorm:"column:rank_winer"`           //! 大赢家排行榜
	RankRecord int       `gorm:"column:rank_record"`          //! 战绩总分排行榜
	TimeType   int       `gorm:"column:time_type;default:0"`  //! 时间类型 0 区间   1 上周  2 上月
	CreatedAt  time.Time `gorm:"column:created_at;type:date"` //! 创建时间

}

func (HouseRank) TableName() string {
	return "house_rank"
}

func initHouseRank(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseRank{}) {
		err = db.AutoMigrate(&HouseRank{}).Error
	} else {
		err = db.CreateTable(&HouseRank{}).Error
		db.Model(&HouseRank{}).AddIndex("ids_hid_createdAt", "dhid", "created_at")
	}
	return err
}
