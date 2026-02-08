package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/pkg/errors"
	"time"
)

const (
	InvalidPay = -1
	IgnorePay  = -2
)

type HouseFloorGearPayCore struct {
	DHId      int64 `gorm:"column:hid"`      // 包厢id
	PlayerNum int   `gorm:"column:play_num"` // 楼层规则
	// 1
	AAPay     bool  `gorm:"column:aa_pay"`      // 是否为AA支付
	Gear1Cost int64 `gorm:"column:gear_1_cost"` // 基础花销 (基础档/一档承担)
	// 2
	Gear2      bool  `gorm:"column:gear_2"`       // 是否勾选低分局(二档)
	Gear2Under int64 `gorm:"column:gear_2_under"` // 低分局低于(二档)
	// 3
	Gear3      bool  `gorm:"column:gear_3"`       // 是否勾选第三档
	Gear3Under int64 `gorm:"column:gear_3_under"` // 第三档低于
	Gear3Cost  int64 `gorm:"column:gear_3_cost"`  // 第三档扣除
	// 4
	Gear4      bool  `gorm:"column:gear_4"`       // 是否勾选第四档
	Gear4Under int64 `gorm:"column:gear_4_under"` // 第四档低于
	Gear4Cost  int64 `gorm:"column:gear_4_cost"`  // 第四档扣除
	// 5
	Gear5 bool `gorm:"column:gear_5"` // 是否勾选第五档

	Gear5Under int64 `gorm:"column:gear_5_under"` // 第五档低于
	Gear5Cost  int64 `gorm:"column:gear_5_cost"`  // 第五档扣除
	// 6
	Gear6      bool  `gorm:"column:gear_6"`       // 是否勾选第六档
	Gear6Under int64 `gorm:"column:gear_6_under"` // 第六档低于
	Gear6Cost  int64 `gorm:"column:gear_6_cost"`  // 第六档扣除

	// 7
	Gear7      bool  `gorm:"column:gear_7"`       // 是否勾选第六档
	Gear7Under int64 `gorm:"column:gear_7_under"` // 第7档低于
	Gear7Cost  int64 `gorm:"column:gear_7_cost"`  // 第7档扣除

	// 8
	Gear8      bool  `gorm:"column:gear_8"`       // 是否勾选第六档
	Gear8Under int64 `gorm:"column:gear_8_under"` // 第8档低于
	Gear8Cost  int64 `gorm:"column:gear_8_cost"`  // 第8档扣除

	// 9
	Gear9      bool  `gorm:"column:gear_9"`       // 是否勾选第六档
	Gear9Under int64 `gorm:"column:gear_9_under"` // 第9档低于
	Gear9Cost  int64 `gorm:"column:gear_9_cost"`  // 第9档扣除

	// 9
	Gear10      bool  `gorm:"column:gear_10"`       // 是否勾选第六档
	Gear10Under int64 `gorm:"column:gear_10_under"` // 第10档低于
	Gear10Cost  int64 `gorm:"column:gear_10_cost"`  // 第10档扣除
}

// 包厢楼层报名费
type HouseFloorGearPay struct {
	FId int64 `gorm:"primary_key;column:id"` // id
	HouseFloorGearPayCore
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` //! 更新时间
}

type HouseFloorGearPayLog struct {
	Id  int64 `gorm:"primary_key;column:id"` // id
	FId int64 `gorm:"column:fid"`            // id
	HouseFloorGearPayCore
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` //! 创建时间
}

func (HouseFloorGearPayLog) TableName() string {
	return "house_floor_gear_pay_log"
}

// 生产一个默认的配置
func NewHouseFloorGearPay(hid, fid int64, playerNum int) *HouseFloorGearPay {
	return &HouseFloorGearPay{
		FId: fid,
		HouseFloorGearPayCore: HouseFloorGearPayCore{
			DHId:       hid,
			PlayerNum:  playerNum,
			AAPay:      false,
			Gear1Cost:  InvalidPay,
			Gear2:      false,
			Gear2Under: InvalidPay,
			Gear3:      false,
			// Gear3Above: InvalidPay,
			Gear3Under: InvalidPay,
			Gear3Cost:  InvalidPay,
			Gear4:      false,
			// Gear4Above: InvalidPay,
			Gear4Under: InvalidPay,
			Gear4Cost:  InvalidPay,
			Gear5:      false,
			// Gear5Above: InvalidPay,
			Gear5Under: InvalidPay,
			Gear5Cost:  InvalidPay,
			Gear6:      false,
			// Gear6Above: InvalidPay,
			Gear6Under:  InvalidPay,
			Gear6Cost:   InvalidPay,
			Gear7:       false,
			Gear7Under:  InvalidPay,
			Gear7Cost:   InvalidPay,
			Gear8:       false,
			Gear8Under:  InvalidPay,
			Gear8Cost:   InvalidPay,
			Gear9:       false,
			Gear9Under:  InvalidPay,
			Gear9Cost:   InvalidPay,
			Gear10:      false,
			Gear10Under: InvalidPay,
			Gear10Cost:  InvalidPay,
		},
		//CreatedAt:             time.Time{},
		//UpdatedAt:             time.Time{},
	}
}

type HouseFloorGearPaySlice []*HouseFloorGearPay

func (hfg *HouseFloorGearPaySlice) FindByFid(fid int64) *HouseFloorGearPay {
	for i := 0; i < len(*hfg); i++ {
		fgp := (*hfg)[i]
		if fgp.FId == fid {
			return fgp
		}
	}
	return nil
}

func (HouseFloorGearPay) TableName() string {
	return "house_floor_gear_pay"
}

func initHouseFloorGearPay(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseFloorGearPay{}) {
		err = db.AutoMigrate(&HouseFloorGearPay{}).Error
	} else {
		err = db.CreateTable(&HouseFloorGearPay{}).Error
	}
	if err != nil {
		return err
	}
	if db.HasTable(&HouseFloorGearPayLog{}) {
		err = db.AutoMigrate(&HouseFloorGearPayLog{}).Error
	} else {
		err = db.CreateTable(&HouseFloorGearPayLog{}).Error
	}
	return err
}

func (fgp *HouseFloorGearPayCore) BaseCost() int64 {
	if fgp.ConfiguredBase() {
		if fgp.AAPay {
			return fgp.Gear1Cost
		} else {
			var playerNum int64
			if fgp.PlayerNum > 0 {
				playerNum = int64(fgp.PlayerNum)
			} else {
				playerNum = 4
			}
			if fgp.Gear1Cost > 0 {
				return fgp.Gear1Cost / playerNum
			} else {
				return 0
			}
		}
	}
	return InvalidPay
}

func (fgp *HouseFloorGearPay) CovertModel() *static.MsgHouseFloorPayItem {
	var item static.MsgHouseFloorPayItem
	item.Fid = fgp.FId
	item.AA = fgp.AAPay
	item.Gear2 = fgp.Gear2
	item.Gear3 = fgp.Gear3
	item.Gear4 = fgp.Gear4
	item.Gear5 = fgp.Gear5
	item.Gear6 = fgp.Gear6
	item.Gear7 = fgp.Gear7
	item.Gear8 = fgp.Gear8
	item.Gear9 = fgp.Gear9
	item.Gear10 = fgp.Gear10
	if fgp.Gear1Cost == InvalidPay {
		item.Gear1Cost = InvalidPay
	} else {
		item.Gear1Cost = static.SwitchVitaminToF64(fgp.Gear1Cost)
	}
	if fgp.Gear3Cost == InvalidPay {
		item.Gear3Cost = InvalidPay
	} else {
		item.Gear3Cost = static.SwitchVitaminToF64(fgp.Gear3Cost)
	}
	if fgp.Gear4Cost == InvalidPay {
		item.Gear4Cost = InvalidPay
	} else {
		item.Gear4Cost = static.SwitchVitaminToF64(fgp.Gear4Cost)
	}
	if fgp.Gear5Cost == InvalidPay {
		item.Gear5Cost = InvalidPay
	} else {
		item.Gear5Cost = static.SwitchVitaminToF64(fgp.Gear5Cost)
	}
	if fgp.Gear6Cost == InvalidPay {
		item.Gear6Cost = InvalidPay
	} else {
		item.Gear6Cost = static.SwitchVitaminToF64(fgp.Gear6Cost)
	}
	if fgp.Gear7Cost == InvalidPay {
		item.Gear7Cost = InvalidPay
	} else {
		item.Gear7Cost = static.SwitchVitaminToF64(fgp.Gear7Cost)
	}
	if fgp.Gear8Cost == InvalidPay {
		item.Gear8Cost = InvalidPay
	} else {
		item.Gear8Cost = static.SwitchVitaminToF64(fgp.Gear8Cost)
	}
	if fgp.Gear9Cost == InvalidPay {
		item.Gear9Cost = InvalidPay
	} else {
		item.Gear9Cost = static.SwitchVitaminToF64(fgp.Gear9Cost)
	}
	if fgp.Gear10Cost == InvalidPay {
		item.Gear10Cost = InvalidPay
	} else {
		item.Gear10Cost = static.SwitchVitaminToF64(fgp.Gear10Cost)
	}

	if fgp.Gear2Under == InvalidPay {
		item.Gear2Under = InvalidPay
	} else {
		item.Gear2Under = static.SwitchVitaminToF64(fgp.Gear2Under)
	}
	if fgp.Gear3Under == InvalidPay {
		item.Gear3Under = InvalidPay
	} else {
		item.Gear3Under = static.SwitchVitaminToF64(fgp.Gear3Under)
	}
	if fgp.Gear4Under == InvalidPay {
		item.Gear4Under = InvalidPay
	} else {
		item.Gear4Under = static.SwitchVitaminToF64(fgp.Gear4Under)
	}
	if fgp.Gear5Under == InvalidPay {
		item.Gear5Under = InvalidPay
	} else {
		item.Gear5Under = static.SwitchVitaminToF64(fgp.Gear5Under)
	}
	if fgp.Gear6Under == InvalidPay {
		item.Gear6Under = InvalidPay
	} else {
		item.Gear6Under = static.SwitchVitaminToF64(fgp.Gear6Under)
	}
	if fgp.Gear7Under == InvalidPay {
		item.Gear7Under = InvalidPay
	} else {
		item.Gear7Under = static.SwitchVitaminToF64(fgp.Gear7Under)
	}
	if fgp.Gear8Under == InvalidPay {
		item.Gear8Under = InvalidPay
	} else {
		item.Gear8Under = static.SwitchVitaminToF64(fgp.Gear8Under)
	}
	if fgp.Gear9Under == InvalidPay {
		item.Gear9Under = InvalidPay
	} else {
		item.Gear9Under = static.SwitchVitaminToF64(fgp.Gear9Under)
	}
	if fgp.Gear10Under == InvalidPay {
		item.Gear10Under = InvalidPay
	} else {
		item.Gear10Under = static.SwitchVitaminToF64(fgp.Gear10Under)
	}

	return &item
}

func (fgp *HouseFloorGearPay) GenLog() *HouseFloorGearPayLog {
	return &HouseFloorGearPayLog{
		FId: fgp.FId,
		HouseFloorGearPayCore: HouseFloorGearPayCore{
			DHId:        fgp.DHId,
			PlayerNum:   fgp.PlayerNum,
			AAPay:       fgp.AAPay,
			Gear1Cost:   fgp.Gear1Cost,
			Gear2:       fgp.Gear2,
			Gear2Under:  fgp.Gear2Under,
			Gear3:       fgp.Gear3,
			Gear3Under:  fgp.Gear3Under,
			Gear3Cost:   fgp.Gear3Cost,
			Gear4:       fgp.Gear4,
			Gear4Under:  fgp.Gear4Under,
			Gear4Cost:   fgp.Gear4Cost,
			Gear5:       fgp.Gear5,
			Gear5Under:  fgp.Gear5Under,
			Gear5Cost:   fgp.Gear5Cost,
			Gear6:       fgp.Gear6,
			Gear6Under:  fgp.Gear6Under,
			Gear6Cost:   fgp.Gear6Cost,
			Gear7:       fgp.Gear7,
			Gear7Under:  fgp.Gear7Under,
			Gear7Cost:   fgp.Gear7Cost,
			Gear8:       fgp.Gear8,
			Gear8Under:  fgp.Gear8Under,
			Gear8Cost:   fgp.Gear8Cost,
			Gear9:       fgp.Gear9,
			Gear9Under:  fgp.Gear9Under,
			Gear9Cost:   fgp.Gear9Cost,
			Gear10:      fgp.Gear10,
			Gear10Under: fgp.Gear10Under,
			Gear10Cost:  fgp.Gear10Cost,
		},
	}
}

// 只要有一个不是默认值 则认为已经配置了
func (fgp *HouseFloorGearPayCore) Configured() bool {
	return fgp.Gear1Cost != InvalidPay ||
		fgp.Gear2Under != InvalidPay ||
		fgp.Gear3Under != InvalidPay ||
		fgp.Gear3Cost != InvalidPay ||
		fgp.Gear4Under != InvalidPay ||
		fgp.Gear4Cost != InvalidPay ||
		fgp.Gear5Under != InvalidPay ||
		fgp.Gear5Cost != InvalidPay ||
		fgp.Gear6Under != InvalidPay ||
		fgp.Gear6Cost != InvalidPay ||
		fgp.Gear7Under != InvalidPay ||
		fgp.Gear8Cost != InvalidPay ||
		fgp.Gear9Under != InvalidPay ||
		fgp.Gear10Cost != InvalidPay
}

// 检查各配置项是否合法
func (fgp *HouseFloorGearPayCore) Check() error {
	if fgp.AAPay {
		return nil
	}

	if ((fgp.Gear10) && (fgp.Gear9Under == InvalidPay || fgp.Gear9Cost == InvalidPay || !fgp.Gear9)) ||
		((fgp.Gear9) && (fgp.Gear8Under == InvalidPay || fgp.Gear8Cost == InvalidPay || !fgp.Gear8)) ||
		((fgp.Gear8) && (fgp.Gear7Under == InvalidPay || fgp.Gear7Cost == InvalidPay || !fgp.Gear7)) ||
		((fgp.Gear7) && (fgp.Gear6Under == InvalidPay || fgp.Gear6Cost == InvalidPay || !fgp.Gear6)) ||
		((fgp.Gear6) && (fgp.Gear5Under == InvalidPay || fgp.Gear5Cost == InvalidPay || !fgp.Gear5)) ||
		((fgp.Gear5) && (fgp.Gear4Under == InvalidPay || fgp.Gear4Cost == InvalidPay || !fgp.Gear4)) ||
		((fgp.Gear4) && (fgp.Gear3Under == InvalidPay || fgp.Gear3Cost == InvalidPay || !fgp.Gear3)) ||
		((fgp.Gear3) && !fgp.ConfiguredBase()) ||
		((fgp.Gear2) && !fgp.ConfiguredBase()) {
		return errors.New("请按序号填写各配置。")
	}

	if (fgp.Gear2Under != InvalidPay && fgp.Gear2Under <= -1) ||
		(fgp.Gear3Under != InvalidPay && fgp.Gear2Under != InvalidPay && fgp.Gear3Under <= fgp.Gear2Under) ||
		(fgp.Gear4Under != InvalidPay && fgp.Gear3Under != InvalidPay && fgp.Gear4Under <= fgp.Gear3Under) ||
		(fgp.Gear5Under != InvalidPay && fgp.Gear4Under != InvalidPay && fgp.Gear5Under <= fgp.Gear4Under) ||
		(fgp.Gear6Under != InvalidPay && fgp.Gear5Under != InvalidPay && fgp.Gear6Under <= fgp.Gear5Under) ||
		(fgp.Gear7Under != InvalidPay && fgp.Gear6Under != InvalidPay && fgp.Gear7Under <= fgp.Gear6Under) ||
		(fgp.Gear8Under != InvalidPay && fgp.Gear7Under != InvalidPay && fgp.Gear8Under <= fgp.Gear7Under) ||
		(fgp.Gear9Under != InvalidPay && fgp.Gear8Under != InvalidPay && fgp.Gear9Under <= fgp.Gear8Under) ||
		(fgp.Gear10Under != InvalidPay && fgp.Gear9Under != InvalidPay && fgp.Gear10Under <= fgp.Gear9Under) {
		xlog.Logger().Errorf("%#v", fgp)
		return errors.New("低于值必须大于高于值。")
	}

	return nil
}

// 是否配置了基础档
func (fgp *HouseFloorGearPayCore) ConfiguredBase() bool {
	return fgp.Gear1Cost >= 0
}

func (fgp *HouseFloorGearPayCore) InGear2(result int64) bool {
	if fgp.SelectedGear2() {
		return result >= -1 && result < fgp.Gear2Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear3(result int64) bool {
	if fgp.SelectedGear3() {
		return result >= fgp.Gear2Under && result < fgp.Gear3Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear4(result int64) bool {
	if fgp.SelectedGear4() {
		return result >= fgp.Gear3Under && result < fgp.Gear4Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear5(result int64) bool {
	if fgp.SelectedGear5() {
		return result >= fgp.Gear4Under && result < fgp.Gear5Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear6(result int64) bool {
	if fgp.SelectedGear6() {
		return result >= fgp.Gear5Under && result < fgp.Gear6Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear7(result int64) bool {
	if fgp.SelectedGear7() {
		return result >= fgp.Gear6Under && result < fgp.Gear7Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear8(result int64) bool {
	if fgp.SelectedGear8() {
		return result >= fgp.Gear7Under && result < fgp.Gear8Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear9(result int64) bool {
	if fgp.SelectedGear9() {
		return result >= fgp.Gear8Under && result < fgp.Gear9Under
	}
	return false
}

func (fgp *HouseFloorGearPayCore) InGear10(result int64) bool {
	if fgp.SelectedGear10() {
		return result >= fgp.Gear9Under && result < fgp.Gear10Under
	}
	return false
}

//// 是否勾选了2级挡位 2级挡位就是低分挡
//func (fgp *HouseFloorGearPayCore) ConfiguredGear2() bool {
//	return fgp.Gear2 &&
//		fgp.Gear2Under != InvalidPay
//}
//
//func (fgp *HouseFloorGearPayCore) ConfiguredGear3() bool {
//	return fgp.Gear3 &&
//		// fgp.Gear3Above != InvalidPay &&
//		fgp.Gear3Under != InvalidPay &&
//		fgp.Gear3Cost != InvalidPay
//}
//
//func (fgp *HouseFloorGearPayCore) ConfiguredGear4() bool {
//	return fgp.Gear4 &&
//		// fgp.Gear4Above != InvalidPay &&
//		fgp.Gear4Under != InvalidPay &&
//		fgp.Gear4Cost != InvalidPay
//}
//
//func (fgp *HouseFloorGearPayCore) ConfiguredGear5() bool {
//	return fgp.Gear5 &&
//		// fgp.Gear5Above != InvalidPay &&
//		fgp.Gear5Under != InvalidPay &&
//		fgp.Gear5Cost != InvalidPay
//}
//
//func (fgp *HouseFloorGearPayCore) ConfiguredGear6() bool {
//	return fgp.Gear6 &&
//		// fgp.Gear6Above != InvalidPay &&
//		fgp.Gear6Under != InvalidPay &&
//		fgp.Gear6Cost != InvalidPay
//}

// 是否勾选了2级挡位 2级挡位就是低分挡
func (fgp *HouseFloorGearPayCore) SelectedGear2() bool {
	return fgp.Gear2 &&
		fgp.Gear2Under != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear3() bool {
	return fgp.Gear3 &&
		// fgp.Gear3Above != InvalidPay &&
		fgp.Gear3Under != InvalidPay &&
		fgp.Gear3Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear4() bool {
	return fgp.Gear4 &&
		// fgp.Gear4Above != InvalidPay &&
		fgp.Gear4Under != InvalidPay &&
		fgp.Gear4Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear5() bool {
	return fgp.Gear5 &&
		// fgp.Gear5Above != InvalidPay &&
		fgp.Gear5Under != InvalidPay &&
		fgp.Gear5Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear6() bool {
	return fgp.Gear6 &&
		// fgp.Gear6Above != InvalidPay &&
		fgp.Gear6Under != InvalidPay &&
		fgp.Gear6Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear7() bool {
	return fgp.Gear7 &&
		// fgp.Gear6Above != InvalidPay &&
		fgp.Gear7Under != InvalidPay &&
		fgp.Gear7Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear8() bool {
	return fgp.Gear8 &&
		// fgp.Gear6Above != InvalidPay &&
		fgp.Gear8Under != InvalidPay &&
		fgp.Gear8Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear9() bool {
	return fgp.Gear9 &&
		// fgp.Gear6Above != InvalidPay &&
		fgp.Gear9Under != InvalidPay &&
		fgp.Gear9Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) SelectedGear10() bool {
	return fgp.Gear10 &&
		// fgp.Gear6Above != InvalidPay &&
		fgp.Gear10Under != InvalidPay &&
		fgp.Gear10Cost != InvalidPay
}

func (fgp *HouseFloorGearPayCore) GetPay(result int64) (int64, error) {
	if !fgp.ConfiguredBase() {
		return 0, fmt.Errorf("gear infrastructure not been configured, %d", fgp.Gear1Cost)
	}

	var cost int64
	if fgp.AAPay {
		cost = fgp.Gear1Cost
	} else {
		switch {
		case fgp.InGear2(result):
			cost = 0
		case fgp.InGear3(result):
			cost = fgp.Gear3Cost
		case fgp.InGear4(result):
			cost = fgp.Gear4Cost
		case fgp.InGear5(result):
			cost = fgp.Gear5Cost
		case fgp.InGear6(result):
			cost = fgp.Gear6Cost
		case fgp.InGear7(result):
			cost = fgp.Gear7Cost
		case fgp.InGear8(result):
			cost = fgp.Gear8Cost
		case fgp.InGear9(result):
			cost = fgp.Gear9Cost
		case fgp.InGear10(result):
			cost = fgp.Gear10Cost
		default:
			cost = fgp.Gear1Cost
		}
	}

	if cost < 0 {
		cost = 0
	}

	return cost, nil
}

func (fgp *HouseFloorGearPayCore) IsValidRound(result int64) bool {
	return !(fgp.AAPay == false && fgp.ConfiguredBase() && fgp.SelectedGear2() && result < fgp.Gear2Under)
}

func (fgp *HouseFloorGearPayCore) Identical(target *HouseFloorGearPayCore) bool {
	return fgp.AAPay == target.AAPay &&
		fgp.Gear1Cost == target.Gear1Cost &&
		fgp.Gear2 == target.Gear2 &&
		fgp.Gear2Under == target.Gear2Under &&
		fgp.Gear3 == target.Gear3 &&
		fgp.Gear3Cost == target.Gear3Cost &&
		fgp.Gear3Under == target.Gear3Under &&
		fgp.Gear4 == target.Gear4 &&
		fgp.Gear4Cost == target.Gear4Cost &&
		fgp.Gear4Under == target.Gear4Under &&
		fgp.Gear5 == target.Gear5 &&
		fgp.Gear5Cost == target.Gear5Cost &&
		fgp.Gear5Under == target.Gear5Under &&
		fgp.Gear6 == target.Gear6 &&
		fgp.Gear6Cost == target.Gear6Cost &&
		fgp.Gear6Under == target.Gear6Under &&
		fgp.Gear7 == target.Gear7 &&
		fgp.Gear7Cost == target.Gear7Cost &&
		fgp.Gear7Under == target.Gear7Under &&
		fgp.Gear8 == target.Gear8 &&
		fgp.Gear8Cost == target.Gear8Cost &&
		fgp.Gear8Under == target.Gear8Under &&
		fgp.Gear9 == target.Gear9 &&
		fgp.Gear9Cost == target.Gear9Cost &&
		fgp.Gear9Under == target.Gear9Under &&
		fgp.Gear10 == target.Gear10 &&
		fgp.Gear10Cost == target.Gear10Cost &&
		fgp.Gear10Under == target.Gear10Under

}
