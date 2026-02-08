package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type HouseMergeState int64

func (hms HouseMergeState) String() string {
	switch hms {
	case HouseMergeStateInvalid:
		return "已撤销合并包厢(无效)"
	case HouseMergeStateHavmind:
		return "有合并包厢意向"
	case HouseMergeStateRevoked:
		return "已撤销申请"
	case HouseMergeStateWaiting:
		return "等待答复"
	case HouseMergeStateRefused:
		return "已拒绝"
	case HouseMergeStateAproved:
		return "已同意"
	case HouseMergeStateRevoking:
		return "撤销合并包厢申请中"
	case HouseMergeStateRevokeRef:
		return "撤销合并包厢申请被拒绝"
	default:
		return "无效状态"
	}
}

// 包厢合并包厢声请状态枚举
// 两个包厢之间最多只能存在一种状态
const (
	HouseMergeStateInvalid   HouseMergeState = -1 // 无效状态/合并后已撤销合并包厢/撤销合并包厢被同意
	HouseMergeStateHavmind   HouseMergeState = 0  // 有意向合并包厢
	HouseMergeStateRefused   HouseMergeState = 1  // 拒绝申请
	HouseMergeStateRevoked   HouseMergeState = 2  // 撤回申请
	HouseMergeStateWaiting   HouseMergeState = 3  // 发出申请/等待响应
	HouseMergeStateAproved   HouseMergeState = 4  // 申请通过
	HouseMergeStateRevoking  HouseMergeState = 5  // 申请撤销合并包厢中
	HouseMergeStateRevokeRef HouseMergeState = 6  // 申请撤销合并包厢被拒绝
)

// 包厢状态枚举
const (
	HouseMergeHidStateNormal   = 0  // 自然状态
	HouseMergeHidStateDevourer = -1 // 已吃掉至少一个圈
	HouseMergeHidStateMerging  = -2 // 合并包厢中
	HouseMergeHidStateRevoking = -3 // 撤销合并包厢中
)

// Js客户端显示状态
const (
	HouseMergeClientStateRsp       = 1  // 收到申请
	HouseMergeClientStateReq       = 2  // 发送申请
	HouseMergeClientStateRspRef    = 3  // 拒绝了收到的申请
	HouseMergeClientStateReqRef    = 4  // 发送申请被拒绝
	HouseMergeClientStateRspApr    = 5  // 同意了收到的申请
	HouseMergeClientStateReqApr    = 6  // 发送的申请通过了
	HouseMergeClientStateReqRvk    = 7  // 合并包厢申请被撤销
	HouseMergeClientStateRvkReq    = 8  // 发送撤销合并包厢申请
	HouseMergeClientStateRvkRsp    = 9  // 收到撤销合并包厢申请
	HouseMergeClientStateRvkReqRef = 10 // 发送撤销合并包厢申请被拒绝
	HouseMergeClientStateRvkRspRef = 11 // 收到撤销合并包厢申请已拒绝
)

// 包厢合并记录
type HouseMergeLog struct {
	Id         int64           `gorm:"primary_key;column:id"`
	Swallowed  int64           `gorm:"column:swallowed;default:0;comment:'原来的/被合并圈DHID'"`
	Devourer   int64           `gorm:"column:devourer;default:0;comment:'新的/合并圈DHID'"`
	Sponsor    int64           `gorm:"column:sponsor;default:0;comment:'发起方'"`
	MergeState HouseMergeState `gorm:"column:merge_state;default:0;comment:'事件状态:0等待答复1申请被撤销2已拒绝3已同意'"`
	MergeAt    int64           `gorm:"column:merge_at"`
	CreatedAt  time.Time       `gorm:"column:created_at;type:datetime"`
	UpdatedAt  time.Time       `gorm:"column:updated_at;type:datetime"`
}

func (HouseMergeLog) TableName() string {
	return "house_merge_log"
}

func initHouseMergeLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMergeLog{}) {
		err = db.AutoMigrate(&HouseMergeLog{}).Error
	} else {
		err = db.CreateTable(&HouseMergeLog{}).Error
	}
	return err
}
