package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

//贡献值系统开关表
type ConfigContributionSystem struct {
	IsOpen          int `gorm:"column:IsOpen"`          //! 总开关
	IsGameRoomOpen  int `gorm:"column:IsGameRoomOpen"`  //! 游戏房间贡献值开关
	IsGameShopOpen  int `gorm:"column:IsGameShopOpen"`  //! 游戏商城贡献值开关
	IsDayScoreOpen  int `gorm:"column:IsDayScoreOpen"`  //! 日净分控制开关
	IsNewPlayerOpen int `gorm:"column:IsNewPlayerOpen"` //!	新玩家强补控制开关
}

func (ConfigContributionSystem) TableName() string {
	return "config_contributionSystem"
}

func initConfigContributionSystem(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigContributionSystem{}) {
		err = db.AutoMigrate(&ConfigContributionSystem{}).Error
	} else {
		err = db.CreateTable(&ConfigContributionSystem{}).Error
	}

	if db.HasTable(&ConfigContributionGame{}) {
		err = db.AutoMigrate(&ConfigContributionGame{}).Error
	} else {
		err = db.CreateTable(&ConfigContributionGame{}).Error
	}

	if db.HasTable(&ConfigContributionNewPlayer{}) {
		err = db.AutoMigrate(&ConfigContributionNewPlayer{}).Error
	} else {
		err = db.CreateTable(&ConfigContributionNewPlayer{}).Error
	}

	if db.HasTable(&ConfigContributionShop{}) {
		err = db.AutoMigrate(&ConfigContributionShop{}).Error
	} else {
		err = db.CreateTable(&ConfigContributionShop{}).Error
	}

	if db.HasTable(&ConfigContributionScore{}) {
		err = db.AutoMigrate(&ConfigContributionScore{}).Error
	} else {
		err = db.CreateTable(&ConfigContributionScore{}).Error
	}

	if db.HasTable(&ContributionGameLog{}) {
		err = db.AutoMigrate(&ContributionGameLog{}).Error
	} else {
		err = db.CreateTable(&ContributionGameLog{}).Error
	}

	if db.HasTable(&ContributionShopLog{}) {
		err = db.AutoMigrate(&ContributionShopLog{}).Error
	} else {
		err = db.CreateTable(&ContributionShopLog{}).Error
	}

	if db.HasTable(&ContributionNewPlayerCtrlLog{}) {
		err = db.AutoMigrate(&ContributionNewPlayerCtrlLog{}).Error
	} else {
		err = db.CreateTable(&ContributionNewPlayerCtrlLog{}).Error
	}

	if db.HasTable(&ContributionNewPlayerLog{}) {
		err = db.AutoMigrate(&ContributionNewPlayerLog{}).Error
	} else {
		err = db.CreateTable(&ContributionNewPlayerLog{}).Error
	}

	if db.HasTable(&ContributionDayScoreCtrlLog{}) {
		err = db.AutoMigrate(&ContributionDayScoreCtrlLog{}).Error
	} else {
		err = db.CreateTable(&ContributionDayScoreCtrlLog{}).Error
	}

	if db.HasTable(&ContributionDayScoreLog{}) {
		err = db.AutoMigrate(&ContributionDayScoreLog{}).Error
	} else {
		err = db.CreateTable(&ContributionDayScoreLog{}).Error
	}

	return err
}

// 游戏贡献值配置表
type ConfigContributionGame struct {
	KindId     int   `gorm:"column:KindId"`     //! 游戏kindid
	IsGameOpen int   `gorm:"column:IsGameOpen"` //! 单个游戏贡献值开关
	SiteType   int   `gorm:"column:SiteType"`   //! 房间级别(初级房、中级房、高级房)
	Extract    int   `gorm:"column:Extract"`    //! 房间贡献值抽取率 (从台费中抽取固定的金额比例填入个人贡献值。充值金币的固定比例填入个人贡献值)
	Trigger    int64 `gorm:"column:Trigger"`    //! 房间触发值,根据各房间的单局平均输赢数额去触发
	Proportion int   `gorm:"column:Proportion"` //! 累计比例,每次抽取金币转换成贡献值时需要乘以比例千分比
	LuckyRate  int   `gorm:"column:LuckyRate"`  //! 额外判定生效率，就是达到触发条件后有多少几率(0-100)触发
}

func (ConfigContributionGame) TableName() string {
	return "config_contribution_game"
}

//新玩家强补控制配置表
type ConfigContributionNewPlayer struct {
	ForceSubsidyLeft int `gorm:"column:ForceSubsidyLeft"` //!	强补剩余次数
	NewPlayerDefine  int `gorm:"column:NewPlayerDefine"`  //! 新玩家定义(距离首次登陆几日为新玩家)
	LuckyRate        int `gorm:"column:LuckyRate"`        //! 强补达到触发条件时，真正启动强补的比例，用来防止玩家摸索规律（千分比）
}

func (ConfigContributionNewPlayer) TableName() string {
	return "config_contribution_newplayer"
}

//商品价格与贡献值比例关系配置表
type ConfigContributionShop struct {
	Price             int `gorm:"column:price"`             //! 商品价格(单位：分)
	ContributionScore int `gorm:"column:contributionScore"` //! 对应的贡献值
}

func (ConfigContributionShop) TableName() string {
	return "config_contribution_shop"
}

//日净分控制配置表
type ConfigContributionScore struct {
	DayScoreLowerLimit1     int `gorm:"column:DayScoreLowerLimit1"`     //! 日净分下限1,日净分的低档上限，当玩家的日净分大于该值时，会稍微提高拿到牌权最小牌的几率
	DayLowerLimitLuckyRate1 int `gorm:"column:DayLowerLimitLuckyRate1"` //! 触发日净分下限1时额外拿牌权最大牌几率
	DayScoreUpperLimit1     int `gorm:"column:DayScoreUpperLimit1"`     //! 日净分的低档上限，当玩家的日净分大于该值时，会稍微提高拿到牌权最小牌的几率
	DayUpperLimitLuckyRate1 int `gorm:"column:DayUpperLimitLuckyRate1"` //! 触发日净分上限1时额外拿牌权最小牌几率
	DayScoreLowerLimit2     int `gorm:"column:DayScoreLowerLimit2"`     //! 日净分下限2,日净分的高档下限，当玩家的日净分小于该值时，会大幅增加拿到牌权最大牌几率
	DayLowerLimitLuckyRate2 int `gorm:"column:DayLowerLimitLuckyRate2"` //! 触发日净分下限2时额外拿牌权最大牌几率
	DayScoreUpperLimit2     int `gorm:"column:DayScoreUpperLimit2"`     //! 日净分上限2,日净分的高档上限，当玩家的日净分大于该值时，会大幅提高拿到牌权最小牌的几率
	DayUpperLimitLuckyRate2 int `gorm:"column:DayUpperLimitLuckyRate2"` //! 触发日净分上限2时额外拿牌权最大牌几率
}

func (ConfigContributionScore) TableName() string {
	return "config_contribution_score"
}

//游戏贡献值变动log表
type ContributionGameLog struct {
	Id                int       `gorm:"primary_key;column:id"`           //! id
	GameNum           string    `gorm:"column:game_num;size:100"`        //! 牌局编号
	UId               int64     `gorm:"column:uid;type:bigint"`          //! 玩家id
	KindId            int       `gorm:"column:kind_id"`                  //! 子游戏id
	SiteType          int       `gorm:"column:site_type"`                //! 场次类型
	Revenue           int       `gorm:"column:revenue"`                  //! 茶水费
	Proportion        int       `gorm:"column:proportion"`               //! 累计比例,每次抽取金币转换成贡献值时需要乘以比例千分比
	ContributionScore int       `gorm:"column:contributionScore"`        //! 本局生成或者消耗的贡献分
	ScoreType         int       `gorm:"column:score_type"`               //! 贡献分类型(1生成, 2消耗)
	CreatedAt         time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (ContributionGameLog) TableName() string {
	return "record_contribution_game_log"
}

//商城充值贡献值变动log表
type ContributionShopLog struct {
	Id                int       `gorm:"primary_key;column:id"`           //! id
	OrderID           string    `gorm:"column:orderID;size:100"`         //! 订单号
	UId               int64     `gorm:"column:uid;type:bigint"`          //! 玩家id
	ProductId         int       `gorm:"column:productID"`                //! 商品id
	Name              string    `gorm:"column:name;size:55"`             //! 商品名称
	Price             int       `gorm:"colunm:price"`                    //! 花费
	ContributionScore int       `gorm:"column:contributionScore"`        //! 购买本商品获得的贡献分
	CreatedAt         time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (ContributionShopLog) TableName() string {
	return "record_contribution_shop_log"
}

//新玩家强补控制log表
type ContributionNewPlayerCtrlLog struct {
	Id          int       `gorm:"primary_key;column:id"`           //! id
	GameNum     string    `gorm:"column:game_num;size:100"`        //! 牌局编号
	UId         int64     `gorm:"column:uid;type:bigint"`          //! 玩家id
	KindId      int       `gorm:"column:kind_id"`                  //! 子游戏id
	SiteType    int       `gorm:"column:site_type"`                //! 场次类型
	Rate        int       `gorm:"column:rate"`                     //! 强补达到触发条件时，真正启动强补的比例，用来防止玩家摸索规律（千分比）
	IsSuccessed int       `gorm:"column:isSuccessed"`              //! 是否成功触发强补(就是通过强补控制让玩家获得了拿到好牌的机会)
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (ContributionNewPlayerCtrlLog) TableName() string {
	return "record_contribution_newplayerCtrl_log"
}

//新玩家强补控制变动log表
type ContributionNewPlayerLog struct {
	Id               int       `gorm:"primary_key;column:id"`           //! id
	UId              int64     `gorm:"column:uid;type:bigint"`          //! 玩家id
	ForceSubsidyLeft int       `gorm:"column:ForceSubsidyLeft"`         //!	强补剩余次数
	CreatedAt        time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (ContributionNewPlayerLog) TableName() string {
	return "record_contribution_newplayer_log"
}

//日净分控制log表
type ContributionDayScoreCtrlLog struct {
	Id          int       `gorm:"primary_key;column:id"`           //! id
	GameNum     string    `gorm:"column:game_num;size:100"`        //! 牌局编号
	UId         int64     `gorm:"column:uid;type:bigint"`          //! 玩家id
	KindId      int       `gorm:"column:kind_id"`                  //! 子游戏id
	SiteType    int       `gorm:"column:site_type"`                //! 场次类型
	CurDayScore int       `gorm:"column:curDayScore"`              //! 当前日净分
	Grade       int       `gorm:"column:grade"`                    //! 日净分所处档次(-1,-2, 1, 2) -1负数代表下限1档, 2代表上限2档
	Rate        int       `gorm:"column:proportion"`               //! 几率
	IsSuccessed int       `gorm:"column:isSuccessed"`              //! 是否成功触发日净分控制
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (ContributionDayScoreCtrlLog) TableName() string {
	return "record_contribution_dayScoreCtrl_log"
}

//日净分log表
type ContributionDayScoreLog struct {
	Id          int       `gorm:"primary_key;column:id"`           //! id
	UId         int64     `gorm:"column:uid;type:bigint"`          //! 玩家id
	CurDayScore int       `gorm:"column:curDayScore"`              //! 当前日净分
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (ContributionDayScoreLog) TableName() string {
	return "record_contribution_dayScore_log"
}
