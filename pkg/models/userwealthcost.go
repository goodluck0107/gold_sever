package models

import (
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	CostTypeRegister              = 0  // 通过注册赠送
	CostTypeGame                  = 1  // 通过游戏对局结算消耗
	CostTypeReturn                = 2  // 没有正常完成1局游戏 系统返回消耗的
	CostTypeBuy                   = 3  // 通过购买商品获得
	CostTypeGm                    = 4  // 通过gm命令获得
	CostTypePartner               = 5  // 队长发放
	CostTypeCertificationAward    = 6  // 实名认证奖励
	CostTypeInsureGold            = 7  // 保险箱存取
	CostTypeBroadcast             = 8  // 用户广播消耗
	CostTypeAllowances            = 9  // 系统低保补助
	CostTypeCheckin               = 10 // 签到
	CostTypeRevenue               = 11 // 茶水费
	CostTaskReward                = 12 // 任务
	CostTypeSticker               = 13 // 聊天魔法表情
	CostTypeWelcomeGift           = 14 // 见面礼
	CostFrozen                    = 15 // 冻结
	CostAdminOption               = 16 // 权限者修改
	CostTypeGameEx_BW             = 17 // 通过游戏对局扩展 大赢家额外
	CostTypeGameVitamin           = 18 // 疲劳值
	ConstExchange                 = 19 // 商城兑换
	CostTypeVideoAdAward          = 20 // 视频广告奖励
	CostTypeMatchRanking          = 21 // 排位赛奖励
	CostTypeRedBagReturnUserAward = 22 // 微信小游戏红包活动老用户回归奖励金币
	CostTypeRedBagAward           = 23 // 微信小游戏红包活动奖励
	CostTypeCheckinShare          = 24 // 签到+分享双倍领取
	CostTypeDailyReward           = 25 // 每日礼包
	CostTypeShare                 = 26 // 分享
	CostTypeBankruptcyGift        = 27 // 破产礼包
	CostTypeDiamondExchangeGift   = 28 // 钻石兑换礼包
	CostTypeDiamondExchangeRcd    = 29 // 钻石兑换记牌器
	CostTaskRewardGame            = 30 // 房间内的任务
	CostTaskRewardJiabeiGame      = 31 // 房间内的任务加倍领取
	CostTypePayment               = 32 // 支付购买
	CostTypeActivitySpin          = 33 // 活动转盘
	CostTypeActivityCheckin       = 34 // 活动签到
)

// ! 用户房卡流水记录表
type UserWealthCost struct {
	Id         int64     `gorm:"primary_key;column:id"`           // id
	Uid        int64     `gorm:"column:uid;index"`                // 玩家用户ID
	CostType   int8      `gorm:"column:cost_type;index"`          // 流水类型
	WealthType int8      `gorm:"column:wealth_type;index"`        // 财富类型
	BeforeNum  int64     `gorm:"column:before_num"`               //! 消耗前
	AfterNum   int64     `gorm:"column:after_num"`                //! 消耗后
	Cost       int64     `gorm:"column:cost"`                     //! 消耗
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func (UserWealthCost) TableName() string {
	return "user_wealth_cost"
}

func initUserWealthCost(db *gorm.DB) error {
	var err error
	if db.HasTable(&UserWealthCost{}) {
		err = db.AutoMigrate(&UserWealthCost{}).Error
	} else {
		err = db.CreateTable(&UserWealthCost{}).Error
	}
	return err
}

// 得到低保礼包，已经购买了的人数
func GetAllowanceGiftPurchasedCount(db *gorm.DB) (int64, error) {
	var count int64
	if err := db.Table(UserWealthCost{}.TableName()).
		Where("cost_type = ?", CostTypeBankruptcyGift).
		Count(&count).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, nil
		} else {
			xlog.Logger().Errorf("查询所有玩家已购买的破产礼包次数错误%s", err)
			return 0, err
		}
	}
	return count, nil
}

// 今天是否已经购买了低保礼包
func IsAllowanceGiftPurchasedToday(db *gorm.DB, uid int64, nowStr string) (bool, error) {
	var count int
	if err := db.Table(UserWealthCost{}.TableName()).
		Where(`uid = ? and cost_type = ? and date_format(created_at, "%Y-%m-%d") = ?`, uid, CostTypeBankruptcyGift, nowStr).
		Count(&count).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		} else {
			xlog.Logger().Errorf("查询玩家%d今日是否已购买破产礼包错误%s,此时默认不可购买。", uid, err)
			return false, err
		}
	}
	return count > 0, nil
}
