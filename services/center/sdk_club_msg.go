package center

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"
)

type ClubMsg int

const (
	ActiveCreate      ClubMsg = 1
	ActiveClose       ClubMsg = 2
	FloorRuleChange   ClubMsg = 3
	AddBlackList      ClubMsg = 4
	RemoveBlackList   ClubMsg = 5
	PrivacyChange     ClubMsg = 6
	FrozenChange      ClubMsg = 7
	AdminChange       ClubMsg = 8
	UnSetAdmin        ClubMsg = 9
	PartnerChange     ClubMsg = 10
	HouseCheckChange  ClubMsg = 11
	DeleteFloor       ClubMsg = 12
	CreateFloor       ClubMsg = 13
	PartnerUserChange ClubMsg = 14
	KickHouseMem      ClubMsg = 15
	JoinHouse         ClubMsg = 16
	ExitHouse         ClubMsg = 17
	GameStatueChange  ClubMsg = 18
	PartnerDelTable   ClubMsg = 19
	HouseVitaminClear ClubMsg = 20
	PartnerVitamin    ClubMsg = 21
	MemVitaminSend    ClubMsg = 22
	MemLimitTable     ClubMsg = 23
	VitaminAdmin      ClubMsg = 24
	AddUserGroup      ClubMsg = 25
	AddGroupUser      ClubMsg = 26
	MsgBanWeChat      ClubMsg = 27
	FloorVipChange    ClubMsg = 28
	TeamOffWork       ClubMsg = 29
	TeamBan           ClubMsg = 30
	TeamKick          ClubMsg = 31
	AlarmValue        ClubMsg = 32
	GameReward        ClubMsg = 33
	AACost            ClubMsg = 34
	FloorPayChange    ClubMsg = -1 //小于0类型客户端不展示
	VicePartner       ClubMsg = -2
)

//CreateClubMassage 新增包厢消息
func CreateClubMassage(hid, creater int64, msgType ClubMsg, msg string) {
	sql := `insert into house_msg(hid,creater,msg,msg_type,create_stamp) values(?,?,?,?,?)`
	err := GetDBMgr().GetDBmControl().Exec(sql, hid, creater, msg, msgType, time.Now().Unix()).Error
	if err != nil {
		xlog.Logger().Errorf("插入消息失败：%v", err)
	}
}

//CreateClubMassageWithTx 新增包厢消息
func CreateClubMassageWithTx(tx *gorm.DB, hid, creater int64, msgType ClubMsg, msg string) {
	if tx == nil {
		tx = GetDBMgr().GetDBmControl()
	}
	sql := `insert into house_msg(hid,creater,msg,msg_type,create_stamp) values(?,?,?,?,?)`
	err := tx.Exec(sql, hid, creater, msg, msgType, time.Now().Unix()).Error
	if err != nil {
		xlog.Logger().Errorf("插入消息失败：%v", err)
	}
}

func CreateClubFloorDelMsg(dhid, dfid int64, dfindex int) {
	//royalty, partner := GetClubRoyaltyByPartner(dhid, []int64{dfid})
	//royaltyData, err := json.Marshal(royalty)
	royaltyStr := ""
	//if err == nil {
	//	royaltyStr = public.HF_Bytestoa(royaltyData)
	//} else {
	//	syslog.Logger().Errorf("楼层配置转json失败:dhid:%d, dfid：%d err:%v", dhid, dfid, err)
	//}

	//partnerData, err := json.Marshal(partner)
	partnerStr := ""
	//if err == nil {
	//	partnerStr = public.HF_Bytestoa(partnerData)
	//} else {
	//	syslog.Logger().Errorf("楼层关系转json失败:dhid:%d, dfid：%d err:%v", dhid, dfid, err)
	//}
	sql := `insert into house_floor_del_msg (dhid,dfid,dfindex,create_stamp,floorroyalty,floorpartner) values(?,?,?,?,?,?)`
	err := GetDBMgr().GetDBmControl().Exec(sql, dhid, dfid, dfindex, time.Now().Unix(), royaltyStr, partnerStr).Error
	if err != nil {
		xlog.Logger().Errorf("插入消息失败：%v", err)
	}
}
