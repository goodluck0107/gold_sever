package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 金币场对局结算详情
type GameResultDetail struct {
	Id             int       `gorm:"primary_key;column:id"`           //! id
	GameNum        string    `gorm:"column:game_num;size:100"`        //! 牌局编号
	UId            int64     `gorm:"column:uid"`                      //! 玩家id
	BeforeScore    int       `gorm:"column:before_score"`             //! 结算前总分
	AfterScore     int       `gorm:"column:after_score"`              //! 结算后总分
	Score          int       `gorm:"column:score"`                    //! 当前得分
	Revenue        int       `gorm:"column:revenue"`                  //! 茶水费
	Award          int       `gorm:"column:award"`                    //! 任务奖励分数(预留字段)
	Result         int       `gorm:"column:result"`                   //! 玩家当前对局状态
	KindId         int       `gorm:"column:kind_id"`                  //! 子游戏id
	SiteType       int       `gorm:"column:site_type"`                //! 场次类型
	GameId         int       `gorm:"column:game_id"`                  //! 游戏进程id
	Ntid           int       `gorm:"column:ntid"`                     //! 桌号
	ClientIp       string    `gorm:"column:client_ip;size:32"`        //! ip地址
	Fanshu         int       `gorm:"column:fanshu"`                   //! 番数(预留字段)
	Difen          int       `gorm:"column:difen"`                    //! 底分
	BeginTime      time.Time `gorm:"column:begintime;type:datetime"`  //! 游戏开始时间
	EndTime        time.Time `gorm:"column:endtime;type:datetime"`    //! 游戏结束时间
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	BeforeDiamond  int       `gorm:"column:before_diamond"`           // 游戏开始前的钻石
	DiamondChanged bool      `gorm:"column:diamond_changed"`          // 是否有钻石消耗
	JiaBeiType     byte      `gorm:"column:jia_bei_type"`             // 加倍类型 0不加倍 1加倍 2超级加倍
	UserChannel    int       `gorm:"column:user_channel;default:0"`   // 渠道标识 0 默认 1 小程序 2 华为
}

func (GameResultDetail) TableName() string {
	return "game_result_detail"
}

func initGameResultDetail(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameResultDetail{}) {
		err = db.AutoMigrate(&GameResultDetail{}).Error
	} else {
		err = db.CreateTable(&GameResultDetail{}).Error
	}
	return err
}
