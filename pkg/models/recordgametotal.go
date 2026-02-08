package models

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"

	"github.com/jinzhu/gorm"
)

type QueryMemberStatisticsResult struct {
	Uid         int64   `gorm:"column:uid"`         //! 玩家id
	TotalScore  float64 `gorm:"column:totalscore"`  //! 今日总成绩
	PlayTimes   int     `gorm:"column:playtimes"`   //! 玩家对局数
	ValidTimes  int     `gorm:"column:validtimes"`  //! 玩家有效局数
	BigWinTimes int     `gorm:"column:bigwintimes"` //! 玩家大赢家次数
}

type QueryMemberGameRecordResult struct {
	Uid     int64   `gorm:"column:uid"`      //! 玩家id
	Score   float64 `gorm:"column:score"`    //! 分数
	GameNum string  `gorm:"column:game_num"` //! 游戏ID,唯一标识
	FId     int     `gorm:"column:fid"`      //! 包厢楼层ID
}

//! 游戏总结算记录
type RecordGameTotal struct {
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
	RoomChannel    int       `gorm:"column:room_channel;default:0"` // 渠道标识  0 默认 1 小程序 2 华为
	UserChannel    int       `gorm:"column:user_channel;default:0"` // 渠道标识  0 默认 1 小程序 2 华为
}

func (RecordGameTotal) TableName() string {
	return "record_game_total"
}

func initRecordGameTotal(db *gorm.DB) error {
	var err error
	if db.HasTable(&RecordGameTotal{}) {
		err = db.AutoMigrate(&RecordGameTotal{}).Error
	} else {
		err = db.CreateTable(&RecordGameTotal{}).Error
		if err == nil {
			err = db.Exec("ALTER TABLE `record_game_total` MODIFY COLUMN `is_valid_round`  tinyint(1) NULL DEFAULT 1 COMMENT '是否为有效局', MODIFY COLUMN `is_big_winner`  tinyint(1) NULL DEFAULT 0 COMMENT '是否为大赢家';").Error
		}
	}
	return err
}

// db -> redis模型
func (u *RecordGameTotal) ConvertModel() *RecordGameTotalBak {
	p := new(RecordGameTotalBak)
	p.Id = u.Id
	p.KindId = u.KindId
	p.GameNum = u.GameNum
	p.RoomNum = u.RoomNum
	p.PlayCount = u.PlayCount
	p.Round = u.Round
	p.ServerId = u.ServerId
	p.SeatId = u.SeatId
	p.Uid = u.Uid
	p.UName = u.UName
	p.ScoreKind = u.ScoreKind
	p.WinScore = u.WinScore
	p.Ip = u.Ip
	p.HId = u.HId
	p.IsHeart = u.IsHeart
	p.FId = u.FId
	p.DFId = u.DFId
	p.HalfWayDismiss = u.HalfWayDismiss
	p.CreatedAt = u.CreatedAt
	p.Partner = u.Partner
	p.SuperiorId = u.SuperiorId
	p.Radix = u.Radix
	p.IsValidRound = u.IsValidRound
	p.IsBigWinner = u.IsBigWinner
	return p
}

// 解析分数
func (u *RecordGameTotal) GetRealScore() float64 {
	if u.Radix == 0 {
		return static.HF_DecimalDivide(float64(u.WinScore/1), 1, 2)
	}
	return static.HF_DecimalDivide(float64(u.WinScore)/float64(u.Radix), 1, 2)
}

//! 用户游戏留存查询
func GetFangKaGameRetentionStatistics(db *gorm.DB, time time.Time) (int, int, int, int) {
	selectDayStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	selectDay1 := time.AddDate(0, 0, 1)
	selectDay1Str := fmt.Sprintf("%d-%02d-%02d", selectDay1.Year(), selectDay1.Month(), selectDay1.Day())
	selectDay3 := time.AddDate(0, 0, 3)
	selectDay3Str := fmt.Sprintf("%d-%02d-%02d", selectDay3.Year(), selectDay3.Month(), selectDay3.Day())
	selectDay7 := time.AddDate(0, 0, 7)
	selectDay7Str := fmt.Sprintf("%d-%02d-%02d", selectDay7.Year(), selectDay7.Month(), selectDay7.Day())

	var uids []int64

	sql := `select uid from record_game_total where date_format(created_at, "%Y-%m-%d") = ? group by uid`
	if err := db.Raw(sql, selectDayStr).Pluck("uid", &uids).Error; err != nil {
		return 0, 0, 0, 0
	}

	if len(uids) == 0 {
		return 0, 0, 0, 0
	}

	type QueryResult struct {
		Count    int    `gorm:"column:count"`            //! 玩家Id
		GameDate string `gorm:"column:gamedate;size:12"` //! 玩游戏时间
	}

	var results []QueryResult

	var retDay1, retDay3, retDay7 QueryResult
	sql = `select ? as gamedate, count(*) as count from (select uid from record_game_total where date_format(created_at, "%Y-%m-%d") = ? and uid in (?) group by uid) as uidTable`
	if err := db.Raw(sql, selectDay1Str, selectDay1Str, uids).Scan(&retDay1).Error; err != nil {
		return 0, 0, 0, 0
	}

	sql = `select ? as gamedate, count(*) as count from (select uid from record_game_total where date_format(created_at, "%Y-%m-%d") = ? and uid in (?) group by uid) as uidTable`
	if err := db.Raw(sql, selectDay3, selectDay3, uids).Scan(&retDay3).Error; err != nil {
		return 0, 0, 0, 0
	}

	sql = `select ? as gamedate, count(*) as count from (select uid from record_game_total where date_format(created_at, "%Y-%m-%d") = ? and uid in (?) group by uid) as uidTable`
	if err := db.Raw(sql, selectDay7, selectDay7, uids).Scan(&retDay7).Error; err != nil {
		return 0, 0, 0, 0
	}

	results = append(results, retDay1, retDay3, retDay7)

	gameRetention := len(uids)
	gameRetention1 := 0
	gameRetention3 := 0
	gameRetention7 := 0

	for _, result := range results {
		if result.GameDate[0:10] == selectDay1Str {
			gameRetention1 = result.Count
		} else if result.GameDate[0:10] == selectDay3Str {
			gameRetention3 = result.Count
		} else if result.GameDate[0:10] == selectDay7Str {
			gameRetention7 = result.Count
		}
	}
	return gameRetention, gameRetention1, gameRetention3, gameRetention7
}
