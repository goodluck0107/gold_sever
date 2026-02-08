package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//！ 金币场排位赛对局结算详情
type GameMatchDetail struct {
	Id        int       `gorm:"primary_key;column:id"`           //! id
	MatchKey  string    `gorm:"column:match_key;size:100"`       //! 排位赛编号
	GameNum   string    `gorm:"column:game_num;size:100"`        //! 牌局编号
	UId       int64     `gorm:"column:uid"`                      //! 玩家id
	KindId    int       `gorm:"column:kind_id"`                  //! 子游戏id
	SiteType  int       `gorm:"column:site_type"`                //! 场次类型
	Ntid      int       `gorm:"column:ntid"`                     //! 桌号
	Score     int       `gorm:"column:score"`                    //! 当前得分,扣除茶水的
	Result    int       `gorm:"column:result"`                   //! 玩家当前对局状态，0胜，1输，2平，3逃
	ClientIp  string    `gorm:"column:client_ip;size:32"`        //! ip地址
	BeginDate time.Time `gorm:"column:begindate;type:date"`      //! 开始日期
	EndDate   time.Time `gorm:"column:enddate;type:date"`        //! 结束日期
	BeginTime time.Time `gorm:"column:begintime;type:datetime"`  //! 开始时间，解析后只取time
	EndTime   time.Time `gorm:"column:endtime;type:datetime"`    //! 结束时间，解析后只取time
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (GameMatchDetail) TableName() string {
	return "game_match_detail"
}

func initGameMatchDetail(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameMatchDetail{}) {
		err = db.AutoMigrate(&GameMatchDetail{}).Error
	} else {
		err = db.CreateTable(&GameMatchDetail{}).Error
	}
	return err
}

//！ 金币场排位赛对局总结算
type GameMatchTotal struct {
	Id          int       `gorm:"primary_key;column:id"`           // ! id
	MatchId     int       `gorm:"column:match_id"`                 // ! id
	MatchKey    string    `gorm:"column:match_key;size:100"`       // ! 排位赛编号
	UId         int64     `gorm:"column:uid"`                      // ! 玩家id
	KindId      int       `gorm:"column:kind_id"`                  // ! 子游戏id
	SiteType    int       `gorm:"column:site_type"`                // ! 场次类型
	SiteTypeStr string    `gorm:"column:site_typestr"`             // ! 场次类型集，标识排位需要这个字段
	Score       int       `gorm:"column:score"`                    // ! 当前得分,扣除茶水的
	WinCount    int       `gorm:"column:wincount;default:0"`       // ! 玩家胜利次数
	TotalCount  int       `gorm:"column:totalcount"`               // ! 玩家总对局次数，逃跑不算
	Coupon      int       `gorm:"column:coupon;default:0"`         // ! 礼券总数
	Ranking     int       `gorm:"column:ranking;default:0"`        // ! 排名
	BeginDate   time.Time `gorm:"column:begindate;type:date"`      // ! 开始日期
	EndDate     time.Time `gorm:"column:enddate;type:date"`        // ! 结束日期
	BeginTime   time.Time `gorm:"column:begintime;type:datetime"`  // ! 开始时间，解析后只取time
	EndTime     time.Time `gorm:"column:endtime;type:datetime"`    // ! 结束时间，解析后只取time
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime"` // 最后更新时间
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (GameMatchTotal) TableName() string {
	return "game_match_Total"
}

func initGameMatchTotal(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameMatchTotal{}) {
		err = db.AutoMigrate(&GameMatchTotal{}).Error
	} else {
		err = db.CreateTable(&GameMatchTotal{}).Error
	}
	return err
}

//！ 金币场排位赛礼券奖励清单
type GameMatchCoupon struct {
	Id               int             `gorm:"primary_key;column:id"`                       //! id
	MatchId          int             `gorm:"column:match_id"`                             //! 排位赛新的标识
	Name             string          `gorm:"column:name;size:50"`                         // 场次名
	KindId           int             `gorm:"column:kind_id"`                              //! 子游戏id
	SiteTypeStr      string          `gorm:"column:site_typestr"`                         //! 场次类型
	SiteTypes        []int           `gorm:"-"`                                           // 对应SiteType配置
	HonorAwardsStr   string          `gorm:"column:honorawardsstr;size:255;default:'{}'"` // 参入奖励
	HonorAwards      []HonorAwards   `gorm:"-"`
	RankingAwardsStr string          `gorm:"column:rankingawardsstr;size:512;default:'{}'"` // 排位奖励
	RankingAwards    []RankingAwards `gorm:"-"`
	BeginDate        time.Time       `gorm:"column:begindate;type:date"`      //! 开始日期
	EndDate          time.Time       `gorm:"column:enddate;type:date"`        //! 结束日期
	BeginTime        time.Time       `gorm:"column:begintime;type:datetime"`  //! 开始时间，解析后只取time
	EndTime          time.Time       `gorm:"column:endtime;type:datetime"`    //! 结束时间，解析后只取time
	CreatedAt        time.Time       `gorm:"column:created_at;type:datetime"` // 创建时间
	MaxRange         int             `gorm:"-"`                               //! UpperRange的最大值，最多MaxRange个玩家可以获奖
}

// 勋章奖励
type HonorAwards struct {
	SiteType  int `json:"site_type"` // 对应的场次
	JionHonor int `json:"jionhonor"` // 参与一局奖励勋章数量，逃跑局不算
	WinHonor  int `json:"winhonor"`  // 胜利一局奖励勋章数量
}

// 排位奖励
type RankingAwards struct {
	LowerRange int    `json:"lowerrange"` // ! 排名范围的起始点  4-10名 奖励100个礼券，11-50名奖励1元话费
	UpperRange int    `json:"upperrange"` // ! 排名范围的起终点
	AwardType  int    `json:"awardtype"`  // //! 奖励类别，礼券或话费，0表示礼券，1表示话费
	Awards     int    `json:"awards"`     // ! 奖励数目
	AwardStr   string `json:"awardstr"`   // 奖励的文案提醒
	AwardUrl   string `json:"awardutl"`   // 奖励的图片地址
}

func (GameMatchCoupon) TableName() string {
	return "game_match_coupon"
}

func initGameMatchCoupon(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameMatchCoupon{}) {
		err = db.AutoMigrate(&GameMatchCoupon{}).Error
	} else {
		err = db.CreateTable(&GameMatchCoupon{}).Error
	}
	return err
}

//！ 金币场排位赛以领取奖励表
type GameMatchCouponRecord struct {
	Id          int       `gorm:"primary_key;column:id"`           //! id
	MatchId     int       `gorm:"column:match_id"`                 //! 排位赛新的标识
	UId         int64     `gorm:"column:uid"`                      //! 玩家id
	KindId      int       `gorm:"column:kind_id"`                  //! 子游戏id
	SiteTypeStr string    `gorm:"column:site_typestr"`             //! 场次类型
	MatchKey    string    `gorm:"column:match_key;size:100"`       //! 排位赛编号
	AwardType   int       `gorm:"column:awardtype;default:0"`      //! 奖励类别，礼券或话费，0表示礼券，1表示话费
	Awards      int       `gorm:"column:awards;default:0"`         //! 奖励数目
	Ranking     int       `gorm:"column:ranking;default:0"`        //! 排名
	State       int       `gorm:"column:state;default:0"`          //! 是否已经领取，0表示未领取，1表示发放中，2表示已经领取
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (GameMatchCouponRecord) TableName() string {
	return "game_match_coupon_record"
}

func initGameMatchCouponRecord(db *gorm.DB) error {
	var err error
	if db.HasTable(&GameMatchCouponRecord{}) {
		err = db.AutoMigrate(&GameMatchCouponRecord{}).Error
	} else {
		err = db.CreateTable(&GameMatchCouponRecord{}).Error
	}
	return err
}
