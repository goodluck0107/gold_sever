package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type ConfigHouse struct {
	Id               int64 `gorm:"primary_key;column:id"`
	CreateMax        int   `gorm:"column:createMax"`                 // 创建包厢最大数
	JoinMax          int   `gorm:"column:joinMax"`                   // 加入包厢最大数
	MemMax           int   `gorm:"column:memMax"`                    // 成员最大数
	AdminMax         int   `gorm:"column:adminMax"`                  // 管理员最大数
	ActMax           int   `gorm:"column:actmax"`                    // 活动最大值
	TableNum         int   `gorm:"column:tableNum"`                  // 显示桌子数
	IsChecked        bool  `gorm:"column:ischecked"`                 // 是否开启加入审核
	IsFrozen         bool  `gorm:"column:isfrozen"`                  // 是否冻结
	IsMemHide        bool  `gorm:"column:ismemhide"`                 // 是否成员隐藏
	CardCost         int   `gorm:"column:createCard"`                // 创建所需房卡数
	IsPerControl     bool  `gorm:"column:isperControl"`              // 区域权限
	MixAIAble        bool  `gorm:"column:mixai_able;default:1"`      // 15桌开关
	MixAITableNum    int   `gorm:"column:mixai_tbnum;default:15"`    // 如果当前包厢实时桌数≤15桌则智能筛选即使开启了也不生效
	MixAIOnRoundNum  int   `gorm:"column:mixai_on_rdnum;default:8"`  // 两位玩家今日共同玩牌局数≥8局
	MixAIOffRoundNum int   `gorm:"column:mixai_off_rdnum;default:5"` // 这个玩家和其他玩家（非自身防作弊名单上的玩家）游戏≥5局，即解除 这个玩家和所有玩家的作弊关联；
	NoPartnerJunior  bool  `gorm:"column:no_partner_junior"`         // 不允许队长设置下级
	VicePartnerMax   int   `gorm:"column:vice_partner_max;default:3"`
	DefaultHouse     int   `gorm:"column:create_default;default:0"` // 创建默认圈
}

func (ConfigHouse) TableName() string {
	return "config_house"
}

func initConfigHouse(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigHouse{}) {
		err = db.AutoMigrate(&ConfigHouse{}).Error
	} else {
		err = db.CreateTable(&ConfigHouse{}).Error
	}
	return err
}

// stringer 接口实现
func (c *ConfigHouse) String() string {
	return fmt.Sprintf(
		"\n[---------------%s--------------]:\n"+
			"[Id:]%d\n"+
			"[CreateMax:]%d\n"+
			"[JoinMax:]%d\n"+
			"[MemMax:]%d\n"+
			"[AdminMax:]%d\n"+
			"[TableNum:]%d\n"+
			"[IsChecked:]%t\n"+
			"[IsFrozen:]%t\n"+
			"[IsMemHide:]%t\n"+
			"[CardCost:]%d\n"+
			"[IsPerControl:]%t\n"+
			"[MixAIAble:]%t\n"+
			"[MixAITableNum:]%d\n"+
			"[MixAIOnRoundNum:]%d\n"+
			"[MixAIOffRoundNum]:%d\n",
		c.TableName(),
		c.Id,
		c.CreateMax,
		c.JoinMax,
		c.MemMax,
		c.AdminMax,
		c.TableNum,
		c.IsChecked,
		c.IsFrozen,
		c.IsMemHide,
		c.CardCost,
		c.IsPerControl,
		c.MixAIAble,
		c.MixAITableNum,
		c.MixAIOnRoundNum,
		c.MixAIOffRoundNum,
	)
}
