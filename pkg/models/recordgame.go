package models

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"
)

//! 游戏单局总结算记录
type RecordGameRound struct {
	Id        int64     `gorm:"primary_key;column:id"`           //! id
	GameNum   string    `gorm:"column:gamenum"`                  //! 游戏ID,唯一标识
	RoomNum   int       `gorm:"column:roomnum"`                  //! 游戏房间ID
	PlayNum   int       `gorm:"column:playnum"`                  //! 游戏局数
	ServerId  int       `gorm:"column:serverid"`                 //! 游戏服务ID
	SeatId    int       `gorm:"column:seatid"`                   //! 玩家座位ID
	UId       int64     `gorm:"column:uid"`                      //! 玩家用户ID
	UName     string    `gorm:"column:uname"`                    //! 玩家名称
	ScoreKind int       `gorm:"column:scorekind"`                //! 游戏结束类型
	WinScore  int       `gorm:"column:winscore"`                 //! 玩家积分
	Ip        string    `gorm:"column:ip"`                       //! 玩家IP地址
	ReplayId  int64     `gorm:"column:replayid"`                 //! 该局游戏回放ID
	UUrl      string    `gorm:"column:uurl"`                     //! 玩家头像
	UGenber   int       `gorm:"column:ugender"`                  //! 玩家性别
	BeginDate time.Time `gorm:"column:begindate;type:datetime"`  //! 游戏开始时间
	EndDate   time.Time `gorm:"column:writedate;type:datetime"`  //! 游戏结束时间
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	Radix     int       `gorm:"column:radix;default:1"`          //! 算分基数,默认100
}

func (RecordGameRound) TableName() string {
	return "record_game_round"
}

func initRecordGame(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameRound{}) {
		err = db.AutoMigrate(&RecordGameRound{}).Error
	} else {
		err = db.CreateTable(&RecordGameRound{}).Error
	}
	return err
}

// db -> redis模型
func (u *RecordGameRound) ConvertModel() *RecordGameRoundBak {
	p := new(RecordGameRoundBak)
	p.Id = u.Id
	p.GameNum = u.GameNum
	p.RoomNum = u.RoomNum
	p.PlayNum = u.PlayNum
	p.ServerId = u.ServerId
	p.SeatId = u.SeatId
	p.UId = u.UId
	p.UName = u.UName
	p.ScoreKind = u.ScoreKind
	p.WinScore = u.WinScore
	p.Ip = u.Ip
	p.ReplayId = u.ReplayId
	p.UUrl = u.UUrl
	p.UGenber = u.UGenber
	p.BeginDate = u.BeginDate
	p.EndDate = u.EndDate
	p.CreatedAt = u.CreatedAt
	p.Radix = u.Radix
	return p
}

// 解析分数
func (u *RecordGameRound) GetRealScore() float64 {
	if u.Radix == 0 {
		return static.HF_DecimalDivide(float64(u.WinScore/1), 1, 2)
	}
	return static.HF_DecimalDivide(float64(u.WinScore)/float64(u.Radix), 1, 2)
}
