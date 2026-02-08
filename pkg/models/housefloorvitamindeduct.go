package models

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

type HouseFloorVitaminDeduct struct {
	Id                   int64  `gorm:"primary_key;column:id"`                                 // id
	DHId                 int64  `gorm:"column:hid"`                                            // 包厢id
	Rule                 string `gorm:"column:rule;type:text"`                                 // 楼层规则
	VitaminDeductType    int    `gorm:"column:vitamin_deduct_type;comment:'扣除类型0大赢家非0AA'"`     // 对局扣除
	VitaminDeductCount   int64  `gorm:"column:vitamin_deduct_count;comment:'扣除值'"`             // 大赢家扣除
	VitaminLowest        int64  `gorm:"column:vitamin_lowest;comment:'单局结算低于'"`                // 单局结算低于
	VitaminHighest       int64  `gorm:"column:vitamin_highest;comment:'单局结算高于（或等于）'"`          // 单局结算高于（或等于）
	VitaminLowestDeduct  int64  `gorm:"column:vitamin_lowest_deduct;comment:'单局结算低于扣除值'"`      // 单局结算低于扣除值
	VitaminHighestDeduct int64  `gorm:"column:vitamin_highest_deduct;comment:'单局结算高于或等于 扣除值'"` // 单局结算高于或等于 扣除值
	IsVitaminLowest      bool   `gorm:"column:is_vitamin_lowest;comment:'低于是否勾选'"`             // 单局结算高于或等于 扣除值
	IsVitaminHighest     bool   `gorm:"column:is_vitamin_highest;comment:'高于是否勾选'"`            // 单局结算高于或等于 扣除值
	// CreatedAt            time.Time `gorm:"column:created_at;type:datetime"`                       // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseFloorVitaminDeduct) TableName() string {
	return "house_floor_vitamin_deduct"
}

func initHouseFloorVitaminDeduct(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseFloorVitaminDeduct{}) {
		err = db.AutoMigrate(&HouseFloorVitaminDeduct{}).Error
	} else {
		err = db.CreateTable(&HouseFloorVitaminDeduct{}).Error
		if err == nil {
			sql := `INSERT INTO house_floor_vitamin_deduct SELECT id, hid,rule,vitamin_deduct_type,vitamin_deduct_count,vitamin_lowest,vitamin_highest,vitamin_lowest_deduct,vitamin_highest_deduct,is_vitamin_lowest,is_vitamin_highest,updated_at FROM house_floor`
			err = db.Exec(sql).Error
		}
	}
	return err
}

func (fo *HouseFloorVitaminDeduct) AADeduct() bool {
	return fo.VitaminDeductType != 0
}

func (fo *HouseFloorVitaminDeduct) GetDeduct(bw int64) (int64, error) {
	if !fo.ConfiguredGameDeduct() {
		return 0, fmt.Errorf("floor infrastructure deduct not been configured, %d", fo.VitaminDeductCount)
	}
	var cost int64
	if fo.AADeduct() {
		cost = fo.VitaminDeductCount
	} else {
		switch {
		case fo.IsVitaminHighest &&
			fo.VitaminHighest != consts.VitaminInvalidValueSrv &&
			fo.VitaminHighestDeduct != consts.VitaminInvalidValueSrv &&
			bw >= fo.VitaminHighest:
			cost = fo.VitaminHighestDeduct

		case fo.IsVitaminLowest &&
			fo.VitaminLowest != consts.VitaminInvalidValueSrv &&
			fo.VitaminLowestDeduct != consts.VitaminInvalidValueSrv &&
			bw < fo.VitaminLowest:
			cost = fo.VitaminLowestDeduct

		default:
			cost = fo.VitaminDeductCount
		}
	}

	if cost < 0 {
		cost = 0
	}

	return cost, nil
}

func (fo *HouseFloorVitaminDeduct) BaseDeduct() int64 {
	if fo.ConfiguredGameDeduct() {
		if fo.AADeduct() {
			return fo.VitaminDeductCount
		} else {
			return fo.VitaminDeductCount / fo.GetPlayNum()
		}
	}
	return InvalidPay
}

// 扣除相关
func (fo *HouseFloorVitaminDeduct) ConfiguredGameDeduct() bool {
	return fo.VitaminDeductCount != consts.VitaminInvalidValueSrv
}

// 扣除相关
func (fo *HouseFloorVitaminDeduct) GetPlayNum() int64 {
	var rule static.FRule
	err := json.Unmarshal([]byte(fo.Rule), &rule)
	if err != nil {
		xlog.Logger().Errorln(err)
		return 4
	}
	return int64(rule.PlayerNum)
}
