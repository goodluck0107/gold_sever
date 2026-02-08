package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

//! 游戏总结算记录
type RecordGameTotalBak struct {
	Id             int64     `gorm:"primary_key;column:id"`           //! id
	KindId         int       `gorm:"kindid"`                          //! 游戏类型
	GameNum        string    `gorm:"column:game_num"`                 //! 游戏ID,唯一标识
	RoomNum        int       `gorm:"column:room_num"`                 //! 游戏房间ID
	PlayCount      int       `gorm:"column:play_count"`               //! 游戏局数
	Round          int       `gorm:"column:round"`                    //! 游戏总局数
	ServerId       int       `gorm:"column:server_id"`                //! 游戏服务ID
	SeatId         int       `gorm:"column:seat_id"`                  //! 玩家座位ID
	Uid            int64     `gorm:"column:uid"`                      //! 玩家用户ID
	UName          string    `gorm:"column:uname"`                    //! 玩家名称
	ScoreKind      int       `gorm:"column:score_kind"`               //! 游戏结束类型
	WinScore       int       `gorm:"column:win_score"`                //! 玩家积分
	Ip             string    `gorm:"column:ip"`                       //! 玩家IP地址
	HId            int64     `gorm:"column:hid"`                      //! 包厢ID
	IsHeart        int       `gorm:"column:is_heart"`                 //! 该战绩是否点赞
	FId            int       `gorm:"column:fid"`                      //! 包厢楼层ID
	DFId           int       `gorm:"column:dfid"`                     //! 包厢楼层索引
	HalfWayDismiss bool      `gorm:"column:halfwaydismiss"`           //! 是否中途解散
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	Partner        int64     `gorm:"column:partner;default:0"`        //! 队长归属
	SuperiorId     int64     `gorm:"column:superiorid;default:0"`     //! 上级收益
	Radix          int       `gorm:"column:radix;default:1"`          //! 算分基数,默认100
	IsValidRound   bool      `gorm:"column:is_valid_round;comment:'是否为有效局'"`
	IsBigWinner    bool      `gorm:"column:is_big_winner;comment:'是否为大赢家'"`
}

func (RecordGameTotalBak) TableName() string {
	return "record_game_total_bak"
}

func initRecordGameTotalBak(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameTotalBak{}) {
		err = db.AutoMigrate(&RecordGameTotalBak{}).Error
	} else {
		err = db.CreateTable(&RecordGameTotalBak{}).Error
	}
	return err
}
