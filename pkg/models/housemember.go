package models

import (
	"errors"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"

	"github.com/jinzhu/gorm"
)

type HouseMember struct {
	Id           int64      `gorm:"primary_key;column:id"`    //! id
	DHId         int64      `gorm:"column:hid"`               //! 包厢id
	FId          int64      `gorm:"column:fid"`               //! 历史楼层id
	UId          int64      `gorm:"column:uid"`               //! 玩家id
	UVitamin     int64      `gorm:"column:uvitamin"`          //! 疲劳值
	URole        int        `gorm:"column:urole"`             //! 玩家角色
	URemark      string     `gorm:"column:uremark;size:100"`  //! 玩家备注
	PRemark      string     `gorm:"column:p_remark;size:100"` //! 玩家备注
	ApplyTime    *time.Time `gorm:"column:apply_time"`        //! 申请时间
	AgreeTime    *time.Time `gorm:"column:agree_time"`        //! 进入时间
	BwTimes      int        `gorm:"column:bw_times"`          //! 大赢家次数
	PlayTimes    int        `gorm:"column:play_times"`        //! 对局次数
	Forbid       int        `gorm:"column:forbid"`            //! 0正常娱乐1禁止娱乐
	Partner      int64      `gorm:"column:partner"`           //! 队长 0否1是 >1 挂载队长id
	Superior     int64      `gorm:"column:superior"`          //! 队长的上级
	Agent        int64      `gorm:"column:agent"`             //! 队长隶属于哪个代理
	Ref          int64      `gorm:"column:ref"`               //! 队长绑定微信ID
	VitaminAdmin bool       `gorm:"column:vitamin_admin"`
	VicePartner  bool       `gorm:"column:vice_partner"`
	NoFloors     string     `gorm:"column:no_floors" json:"no_floors"` // 不能加入的楼层 ”101,102,103,,,,“
	CreatedAt    time.Time  `gorm:"column:created_at;type:datetime"`   //! 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:datetime"`   //！ 更新时间
}

func (HouseMember) TableName() string {
	return "house_member"
}

func initHouseMember(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMember{}) {
		err = db.AutoMigrate(&HouseMember{}).Error
	} else {
		err = db.CreateTable(&HouseMember{}).Error
	}
	return err
}

// db -> redis模型
func (u *HouseMember) ConvertModel() *static.HouseMember {
	p := new(static.HouseMember)
	p.Id = u.Id
	p.DHId = u.DHId
	p.FId = u.FId
	p.UId = u.UId
	p.UVitamin = u.UVitamin
	p.URole = u.URole
	p.URemark = u.URemark
	p.PRemark = u.PRemark
	if u.ApplyTime != nil {
		p.ApplyTime = u.ApplyTime.Unix()
	}
	if u.AgreeTime != nil {
		p.AgreeTime = u.AgreeTime.Unix()
	}
	p.BwTimes = u.BwTimes
	p.PlayTimes = u.PlayTimes
	p.Forbid = u.Forbid
	p.Partner = u.Partner
	// p.Partner_oid = u.Partner_oid
	p.Ref = u.Ref
	p.Superior = u.Superior
	p.Agent = u.Agent
	p.VitaminAdmin = u.VitaminAdmin
	p.VicePartner = u.VicePartner
	p.NoFloors = u.NoFloors
	return p
}

type HouseMembersMap map[int64]*HouseMember

func (hmm *HouseMembersMap) JuniorsBySuperiors(superiors ...int64) []int64 {
	in := func(sid int64) bool {
		for i := 0; i < len(superiors); i++ {
			if sid == superiors[i] {
				return true
			}
		}
		return false
	}
	res := make([]int64, 0)
	for id, mem := range *hmm {
		if mem.Partner == 1 && in(mem.Superior) {
			res = append(res, id)
		}
	}
	return res
}
func (hmm *HouseMembersMap) GetJuniorsAndNamePlayers(superiorIds ...int64) ([]int64, []*HouseMember) {
	in := func(sid int64) bool {
		for i := 0; i < len(superiorIds); i++ {
			if sid == superiorIds[i] {
				return true
			}
		}
		return false
	}
	ids := make([]int64, 0)
	res := make([]*HouseMember, 0)
	for _, mem := range *hmm {
		if mem.Partner > 0 && in(mem.Partner) {
			res = append(res, mem)
			ids = append(ids, mem.UId)
		} else if mem.Superior > 0 && in(mem.Superior) {
			res = append(res, mem)
			ids = append(ids, mem.UId)
		}
	}
	return ids, res
}

type HouseMemberLog struct {
	Id        int64     `gorm:"primary_key;column:id"`                   //! id
	DHId      int64     `gorm:"column:dhid"`                             //! 包厢id
	UId       int64     `gorm:"column:uid"`                              //! 玩家id
	UVitamin  int64     `gorm:"column:uvitamin"`                         //! 疲劳值
	Type      int       `gorm:"column:type"`                             //! 操作类型
	URole     int       `gorm:"column:urole"`                            //! 玩家角色
	URemark   string    `gorm:"column:uremark;size:100"`                 //! 玩家备注
	BwTimes   int       `gorm:"column:bw_times"`                         //! 大赢家次数
	PlayTimes int       `gorm:"column:play_times"`                       //! 对局次数
	Merge     bool      `gorm:"column:merge;comment:'是否为合并包厢/撤销合并包厢操作'"` //! 是否为合并包厢 撤销合并包厢操作
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"`         // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"`         // 更新时间
}

func (HouseMemberLog) TableName() string {
	return "house_member_log"
}

func initHouseMemberLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberLog{}) {
		err = db.AutoMigrate(&HouseMemberLog{}).Error
	} else {
		err = db.CreateTable(&HouseMemberLog{}).Error
	}
	return err
}

type VitaminChangeType int64

const (
	GameCost        VitaminChangeType = 0  //游戏输赢
	MemSend         VitaminChangeType = 1  //玩家赠送及赠送给玩家
	AdminSend       VitaminChangeType = 2  //管理赠送、扣除
	BigWinCost      VitaminChangeType = 3  //大赢家抽水
	GamePay         VitaminChangeType = 4  //牌局抽水
	SysSend         VitaminChangeType = 5  //系统赠送、扣除
	GameTotal       VitaminChangeType = 6  //每天游戏输赢疲劳值统计
	HouseCreate     VitaminChangeType = 7  //创建赠送
	PoolAdd         VitaminChangeType = 8  //管理从玩家扣除
	JoinClear       VitaminChangeType = 9  //入圈清理
	TaxSum          VitaminChangeType = 10 //房费总收入
	PoolAdminAdd    VitaminChangeType = 11 //管理充入，提取
	PartnerSend     VitaminChangeType = 12 //队长赠送、扣除
	AdminSendSum    VitaminChangeType = 13 //管理扣除统计
	AutoPayPartner  VitaminChangeType = 14 //队长自动划扣
	ViAdminSet      VitaminChangeType = 15 //比赛分管理员 赠送，扣除
	ViAdminSum      VitaminChangeType = 16 //比赛分管理员 赠送，扣除统计
	BackPool        VitaminChangeType = 17 //管理送玩家划到仓库
	CardCost        VitaminChangeType = 18 //奖励均衡房卡扣除
	TaxLeft         VitaminChangeType = 19 //房费剩余
	MemberCost      VitaminChangeType = 20 //扣除玩家
	GameReward      VitaminChangeType = 21 //赛分奖励
	Payment         VitaminChangeType = 32
	ActivitySpin    VitaminChangeType = 33 // 活动转盘获得
	ActivityCheckin VitaminChangeType = 34 // 活动签到获得
)

func GetTypeName(val int64, utype VitaminChangeType) string {
	switch utype {
	case GameCost:
		return "游戏输赢"
	case GameTotal:
		return "牌局总输赢"
	case MemSend:
		if val >= 0 {
			return "收到赠送"
		} else {
			return "主动赠送"
		}
	case AdminSend:
		if val >= 0 {
			return "管理增加"
		} else {
			return "管理扣除"
		}
	case HouseCreate:
		return "初始赠送"
	case PoolAdd:
		if val > 0 {
			return "扣分收入"
		} else {
			return "提取"
		}
	case TaxSum:
		return "总报名费"
	case PoolAdminAdd:
		if val > 0 {
			return "减少配额"
		} else {
			return "增加配额"
		}
	case PartnerSend:
		if val > 0 {
			return "扣除下级/玩家"
		} else {
			return "赠送下级/玩家"
		}
	case AdminSendSum:
		return "罚分统计"
	case AutoPayPartner:
		return "队长划账"
	case ViAdminSet:
		if val > 0 {
			return "管理员赠送"
		}
		return "管理员扣除"
	case ViAdminSum:
		return "竞技管理扣除"
	case BackPool:
		return "扣回仓库"
	case CardCost:
		return "房卡扣除"
	case TaxLeft:
		return "赛事服务"
	case MemberCost:
		if val > 0 {
			return "扣除玩家"
		}
		return "管理扣除"
	case GameReward:
		return "赛点清零"
	case BigWinCost:
		return "大贏家房费"
	case GamePay:
		return "AA房费"
	default:
		return ""
	}
}

type HouseMemberVitaminLog struct {
	Id         int64             `gorm:"primary_key;column:id"` //! id
	DHId       int64             `gorm:"column:dhid"`           //! 包厢id
	OptUid     int64             `gorm:"column:optuid"`         //! 操作者id
	UId        int64             `gorm:"column:uid"`            //! 玩家id
	BefVitamin int64             `gorm:"column:befvitamin"`     //! 变更前疲劳值
	AftVitamin int64             `gorm:"column:aftvitamin"`     //! 变更后疲劳值
	Value      int64             `gorm:"column:value"`
	Type       VitaminChangeType `gorm:"column:type;type:tinyint;default:2"`
	GameNum    string            `gorm:"column:game_num"`                      //! 如果是游戏里面变动 则关联gamenum
	Status     int64             `gorm:"column:status;type:tinyint;default:0"` //0表示正常，1表示已统计
	CreatedAt  time.Time         `gorm:"column:created_at;type:datetime"`      // 创建时间
}

func (HouseMemberVitaminLog) TableName() string {
	return "house_member_vitamin_log"
}

type HouseVitaminGoldLog struct {
	Id        int64     `gorm:"primary_key;column:id"` //! id
	DHId      int64     `gorm:"column:dhid"`           //! 包厢id
	UId       int64     `gorm:"column:uid"`            //! 玩家id
	Afterval  int64     `gorm:"afterval"`              // 变动后
	Cost      int64     `gorm:"column:cost"`
	CostType  int       `gorm:"column:cost_type"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
}

func initHouseMemberVitaminLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseMemberVitaminLog{}) {
		err = db.AutoMigrate(&HouseMemberVitaminLog{}).Error
	} else {
		err = db.CreateTable(&HouseMemberVitaminLog{}).Error
	}
	if db.HasTable(&HouseVitaminGoldLog{}) {
		err = db.AutoMigrate(&HouseVitaminGoldLog{}).Error
	} else {
		err = db.CreateTable(&HouseVitaminGoldLog{}).Error
	}
	return err
}

func AddVitaminLog(hid, opUid, uid int64, befvitamin, aftvitamin int64, vtype VitaminChangeType, gameNum string, tx *gorm.DB) error {
	if tx == nil {
		return errors.New("tx nil ")
	}
	value := aftvitamin - befvitamin
	sql := `insert into house_member_vitamin(uid,hid,vitamin,ex_opt,ex_value,ex_type) values(?,?,?,?,?,?) 
       on DUPLICATE KEY UPDATE vitamin = vitamin + ?, ex_opt = ?,ex_value = ?,ex_type = ?`

	if vtype == BigWinCost || vtype == GamePay {
		err := HousePoolChange(hid, uid, vtype, -1*value, gameNum, tx)
		if err != nil {
			return err
		}
	}

	record := &HouseVitaminGoldLog{
		DHId:      hid,
		UId:       uid,
		Afterval:  aftvitamin,
		Cost:      value,
		CostType:  int(vtype),
		CreatedAt: time.Now(),
	}
	err := tx.Create(record).Error
	if err != nil {
		println(err)
	}

	return tx.Exec(sql, uid, hid, aftvitamin, opUid, value, vtype, value, opUid, value, vtype).Error
}

type HouseUserLimit struct {
	Id        int64     `gorm:"primary_key;column:id"`                                                                  //! id
	DHId      int64     `gorm:"column:dhid;NOT NUll;unique_index:dhid_uid"`                                             //! 包厢id
	OptUid    int64     `gorm:"column:operator;NOT NUll"`                                                               //! 操作者id
	UId       int64     `gorm:"column:uid;NOT NUll;unique_index:dhid_uid"`                                              //! 玩家id
	Status    bool      `gorm:"column:status;default:0"`                                                                //状态0:被禁止，1：取消禁止
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;DEFAULT :CURRENT_TIMESTAMP"`                             // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;DEFAULT:CURRENT_TIMESTAMP ;ON_UPDATE:CURRENT_TIMESTAMP"` // 更新时间
}

func (HouseUserLimit) TableName() string {
	return "house_user_limit"
}
func initHouseUserLimit(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseUserLimit{}) {
		err = db.AutoMigrate(&HouseUserLimit{}).Error
	} else {
		err = db.CreateTable(&HouseUserLimit{}).Error
	}
	return err
}
