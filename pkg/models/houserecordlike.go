package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

const (
	OptTypeGameTotal  = 0
	OptTypeGameRound  = 1
	OptTypeGameBigWin = 2
	OptTypeUser       = 3
	OptTypeTeam       = 4
)

const (
	OptUserTypeAdmin   = 0
	OptUserTypePartner = 1
)

type HouseRecordLike struct {
	Id          int64     `gorm:"primary_key;column:id"`
	DHid        int64     `gorm:"column:dhid;comment:'包厢数据库id'"`
	GameNum     string    `gorm:"column:gamenum;type:varchar(50);default:'';comment:'点赞游戏唯一编码'"`
	LikeUser    int64     `gorm:"column:likeuser;default:0;comment:'点赞玩家uid'"`
	IsLike      bool      `gorm:"column:islike;comment:'是否点赞'"`
	OptUserType int       `gorm:"column:optusertype;comment:'点赞用户类型[盟主管理员]:[0],[队长,副队长]:[1]'"`
	OptType     int       `gorm:"column:opttype;comment:'点赞类型[用户点赞]:[3],[总战绩点赞]:[0],[对局战绩点赞]:[1],[大赢家战绩点赞]:[2]'"`
	LikeTime    string    `gorm:"column:liketime;type:varchar(20);comment:'点赞时间日期'"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime"`
	TimeRange   string    `gorm:"column:time_range;type:varchar(32);default:'00-24';comment:'时段区间'"`
}

func (HouseRecordLike) TableName() string {
	return "house_record_like"
}

func initHouseRecordLike(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseRecordLike{}) {
		err = db.AutoMigrate(&HouseRecordLike{}).Error
	} else {
		err = db.CreateTable(&HouseRecordLike{}).Error
		if err == nil {
			db.Model(&HouseRecordLike{}).AddUniqueIndex("unique_idx_dhid_lt_gn_lu_outp_ot_tr", "dhid", "liketime", "gamenum", "likeuser", "optusertype", "opttype", "time_range")
			db.Model(&HouseRecordLike{}).AddIndex("idx_dhid_ot_lt_tr", "dhid", "opttype", "liketime", "time_range")
		}
	}
	return err
}
