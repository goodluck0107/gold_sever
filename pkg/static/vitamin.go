package static

import (
	"github.com/open-source/game/chess.git/pkg/consts"
)

// 包厢疲劳值选项
type HouseVitaminOptions struct {
	IsVitamin       bool `json:"isvitamin"`       // 防沉迷开关
	IsGamePause     bool `json:"isgamepause"`     // 游戏中到达下限暂停
	IsVitaminHide   bool `json:"isvitaminhide"`   // 防沉迷管理员可见
	IsVitaminModi   bool `json:"isvitaminmodi"`   // 防沉迷管理员可调
	IsPartnerHide   bool `json:"ispartnerhide"`   // 防沉迷队长可见
	IsPartnerModi   bool `json:"ispartnermodi"`   // 防沉迷队长可调
	IsMemberSend    bool `json:"ismembersend"`    // 防成谜成员赠送开关
	IsOpenedVitamin bool `json:"isopenedvitamin"` // 开启过防沉迷 // 暂时未用 后面依产品需求用来做保存楼层开关的功能
	IsOpenedGPause  bool `json:"isopenedgpause"`  // 开启过中途暂停 // 暂时未用 后面依产品需求用来做保存楼层开关的功能
}

// 楼层疲劳值选项
type FloorVitaminOptions struct {
	IsVitamin            bool  `json:"is_vitamin"`              // 防沉迷开关
	IsGamePause          bool  `json:"is_game_pause"`           // 游戏中到达下限暂停
	VitaminLowLimit      int64 `json:"vitamin_low_limit"`       // 疲劳值入桌下限
	VitaminHighLimit     int64 `json:"vitamin_high_limit"`      // 疲劳值入桌上限
	VitaminLowLimitPause int64 `json:"vitamin_low_limit_pause"` // 疲劳值暂停下限
	//IsVitaminLowest      bool  `json:"is_vitamin_lowest"`       // 是否勾选了单局低于
	//IsVitaminHighest     bool  `json:"is_vitamin_highest"`      // 是否勾选了单局高于
	//VitaminLowest        int64 `json:"vitamin_lowest"`          // 单局结算低于
	//VitaminHighest       int64 `json:"vitamin_highest"`         // 单局结算高于（或等于）
	//VitaminLowestDeduct  int64 `json:"vitamin_lowest_deduct"`   // 单局结算低于扣除值
	//VitaminHighestDeduct int64 `json:"vitamin_highest_deduct"`  // 单局结算高于或等于 扣除值
	//VitaminDeductCount   int64 `json:"vitamin_deduct_count"`    // 扣费额度
	//VitaminDeductType    int   `json:"vitamin_deduct_type"`     // 扣费方式 0大赢家 1AA
}

//// 扣除相关
//func (fo *FloorVitaminOptions) ConfiguredGameDeduct() bool {
//	return fo.VitaminDeductCount != constant.VitaminInvalidValueSrv
//}
//
//func (fo *FloorVitaminOptions) ConfiguredLowestDeduct() bool {
//	return fo.VitaminLowestDeduct != constant.VitaminInvalidValueSrv ||
//		fo.VitaminLowest != constant.VitaminInvalidValueSrv
//}
//
//func (fo *FloorVitaminOptions) ConfiguredHighestDeduct() bool {
//	return fo.VitaminHighestDeduct != constant.VitaminInvalidValueSrv ||
//		fo.VitaminHighest != constant.VitaminInvalidValueSrv
//}

// 生效相关
func (fo *FloorVitaminOptions) ConfiguredLowLimit() bool {
	return fo.VitaminLowLimit != consts.VitaminInvalidValueSrv
}

// 生效相关
func (fo *FloorVitaminOptions) ConfiguredHighLimit() bool {
	return fo.VitaminHighLimit != consts.VitaminInvalidValueSrv
}

func (fo *FloorVitaminOptions) ConfiguredLowLimitPause() bool {
	return fo.VitaminLowLimitPause != consts.VitaminInvalidValueSrv
}

//func (fo *FloorVitaminOptions) IsSelectLowest() bool {
//	return fo.IsVitaminLowest &&
//		fo.VitaminLowestDeduct != constant.VitaminInvalidValueSrv &&
//		fo.VitaminLowest != constant.VitaminInvalidValueSrv &&
//		fo.VitaminDeductType == 0
//}
//
//func (fo *FloorVitaminOptions) IsSelectHighest() bool {
//	return fo.IsVitaminHighest &&
//		fo.VitaminHighestDeduct != constant.VitaminInvalidValueSrv &&
//		fo.VitaminHighest != constant.VitaminInvalidValueSrv &&
//		fo.VitaminDeductType == 0
//}
