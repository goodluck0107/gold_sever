package models

// 包厢权位
const (
	HRightNull      Right = 0x0000 // 无权限
	HRightJoinCheck Right = 0x0001 // 入圈审核权限
	HRightKick      Right = 0x0002 // 踢人权限
	HRightExtends   Right = 0x0004 // 扩展战队权限
	HRightTblDel    Right = 0x0008 // 桌子解散权限
)

// 权限表每个角色对应的可选权位
const (
	HouseRightOptionsAdmin   = HRightJoinCheck | HRightKick | HRightExtends | HRightTblDel
	HouseRightOptionsJudge   = HRightJoinCheck | HRightKick | HRightExtends | HRightTblDel
	HouseRightOptionsPartner = HRightJoinCheck | HRightKick | HRightExtends | HRightTblDel
	HouseRightOptionsVicePtr = HRightJoinCheck | HRightKick | HRightExtends | HRightTblDel
	HouseRightOptionsMember  = HRightJoinCheck | HRightKick | HRightExtends | HRightTblDel
)

type Right int

func (r Right) RightJoinCheck() bool {
	return r&HRightJoinCheck != 0
}

func (r Right) RightKick() bool {
	return r&HRightKick != 0
}

func (r Right) RightExtends() bool {
	return r&HRightExtends != 0
}

func (r Right) RightTblDel() bool {
	return r&HRightJoinCheck != 0
}

// 包厢权限表
type HouseRight struct {
	Id           int64 `gorm:"primary_key;column:id" json:"id"`             // 包厢id
	RightAdmin   Right `gorm:"column:right_admin" json:"right_admin"`       // 管理员权位
	RightJudge   Right `gorm:"column:right_judge" json:"right_judge"`       // 裁判权位
	RightPartner Right `gorm:"column:right_partner" json:"right_partner"`   // 队长权位
	RightVicePtr Right `gorm:"column:right_vice_ptr" json:"right_vice_ptr"` // 副队长权位
	RightMember  Right `gorm:"column:right_member" json:"right_member"`     // 成员权位
}
