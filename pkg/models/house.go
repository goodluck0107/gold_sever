package models

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"time"

	"github.com/jinzhu/gorm"
)

type HouseTableJoinType int64

const (
	SelfAdd HouseTableJoinType = 0
	AutoAdd HouseTableJoinType = 1
	NoCheat HouseTableJoinType = 2
)

type House struct {
	Id                 int64                     `gorm:"primary_key;column:id" json:"id"`                                        // id
	HId                int                       `gorm:"column:hid" json:"hid"`                                                  // 包厢id
	UId                int64                     `gorm:"column:uid" json:"uid"`                                                  // 玩家id
	Area               string                    `gorm:"column:area" json:"area"`                                                // 包厢区域编码
	Name               string                    `gorm:"column:name;size:100" json:"name"`                                       // 名称
	Notify             string                    `gorm:"column:notify;type:text" json:"notify"`                                  // 公告
	TableSum           int                       `gorm:"column:table_sum" json:"table_sum"`                                      // 桌子数量
	IsChecked          bool                      `gorm:"column:is_checked" json:"ischecked"`                                     // 入楼认证
	IsFrozen           bool                      `gorm:"column:is_frozen" json:"isfrozen"`                                       // 包厢冻结
	IsMemHide          bool                      `gorm:"column:is_member_hide" json:"ismemberhide"`                              // 玩家隐藏
	IsHidHide          bool                      `gorm:"column:is_hid_hide" json:"ishidhide"`                                    // 圈号隐藏
	IsVitaminHide      bool                      `gorm:"column:is_vitamin_hide;comment:'管理员可见'" json:"isvitaminhide"`            // 防沉迷管理员可见
	IsVitaminModi      bool                      `gorm:"column:is_vitamin_modi;comment:'管理员可调'" json:"isvitaminmodi"`            // 防沉迷管理员可调
	IsPartnerHide      bool                      `gorm:"column:is_partner_hide;comment:'队长可见'" json:"ispartnerhide"`             // 防沉迷队长可见
	IsPartnerModi      bool                      `gorm:"column:is_partner_modi;comment:'队长可调'" json:"ispartnermodi"`             // 防沉迷队长可调
	IsVitamin          bool                      `gorm:"column:is_vitamin;comment:'防沉迷开关'" json:"isvitamin"`                     // 防沉迷
	IsGamePause        bool                      `gorm:"column:is_game_pause;comment:'中途暂停开关'" json:"isgamepause"`               // 防沉迷队长可见
	IsMemberSend       bool                      `gorm:"column:is_member_send;comment:'好友赠送开关'" json:"ismembersend"`             // 防沉迷队长可见
	MixActive          bool                      `gorm:"column:mix_active;comment:'混排是否生效'" json:"mix_active"`                   // 混排是否生效
	MixTableNum        int                       `gorm:"column:mix_table_num;default:0;comment:'混排大厅默认桌数'" json:"mix_table_num"` // 混排大厅默认桌数
	MergeHId           int64                     `gorm:"column:merge_hid;default:0;comment:'吞噬该圈的圈hid -1为大圈'" json:"mergehid"`   // 混排大厅默认桌数
	IsPartnerApply     bool                      `gorm:"column:is_partner_apply;comment:'队长批准入圈'" json:"isparnterapply"`         // 队长可批准申请
	OnlyQuickJoin      bool                      `gorm:"column:onlyquick;default:0;comment:'是否只可以快速入桌'" json:"only_quick"`
	CreatedAt          time.Time                 `gorm:"column:created_at;type:datetime" json:"-"`                              // 创建时间
	UpdatedAt          time.Time                 `gorm:"column:updated_at;type:datetime" json:"-"`                              // 更新时间
	TableJoinType      consts.HouseTableJoinType `gorm:"column:table_join_type;default:0" json:"table_join_type"`               // 混排入桌类型
	AICheck            bool                      `gorm:"column:ai_check;default:false" json:"ai_check"`                         // 智能筛选开关
	AITotalScoreLimit  int                       `gorm:"column:ai_total_score_limit;default:false" json:"ai_total_score_limit"` // 智能筛选开关总分上限
	Dialog             string                    `gorm:"column:dialog;type:text" json:"dia_log" `                               // 公告
	DialogActive       bool                      `gorm:"column:dialog_active" json:"dialog_active"`
	AutoPayPartnrt     bool                      `gorm:"column:auto_pay_partner" json:"auto_pay_partner"`     // 队长分账自动划扣
	IsMemExit          bool                      `gorm:"column:is_member_exit" json:"is_member_exit"`         // 成员退圈开关
	AiSuper            bool                      `gorm:"column:ai_super" json:"ai_super"`                     // 超级防作弊开关
	TableShowCount     int                       `gorm:"column:table_show_count" json:"table_show_count"`     // 展示座子数量
	EmptyTableBack     bool                      `gorm:"column:empty_table_back" json:"empty_table_back"`     // 是否空桌子在后面
	EmptyTableMax      int                       `gorm:"column:empty_table_max" json:"empty_table_max"`       // 最大空桌数
	TableSortType      int                       `gorm:"column:table_sort_type" json:"table_sort_type"`       // 桌子排序类型 0 正常 1 极左
	IsHeadHide         bool                      `gorm:"column:is_head_hide" json:"is_head_hide"`             // 隐藏包厢大厅头像
	IsMemUidHide       bool                      `gorm:"column:is_mem_uid_hide" json:"is_mem_uid_hide"`       // 隐藏包厢大厅头像
	DisVitaminJunior   bool                      `gorm:"column:dis_vitamin_junior" json:"dis_vitamin_junior"` // 禁止队长调整下级比赛分
	IsOnlineHide       bool                      `gorm:"column:is_online_hide" json:"is_online_hide"`
	GameOn             bool                      `gorm:"column:game_on" json:"game_on"`
	AdminGameOn        bool                      `gorm:"column:admin_game_on" json:"admin_game_on"`
	PartnerKick        bool                      `gorm:"column:partner_kick" json:"partner_kick"`       // 队长踢人开关
	RewardBalanced     bool                      `gorm:"column:reward_balanced" json:"reward_balanced"` // 奖励均衡开关
	ApplySwitch        bool                      `gorm:"column:apply_switch" json:"apply_switch" `
	RewardBalancedType int                       `json:"reward_balanced_type" gorm:"column:reward_balanced_type;comment:'均摊方式:0低分局1所有局'"`
	PrivateGPS         bool                      `gorm:"column:private_gps;default:0;comment:'隐藏地理位置开关'" json:"private_gps"`                      // 隐藏地理位置
	FangKaTipsMinNum   int                       `gorm:"column:fangka_tips_min_num;default:200;comment:'房卡低于xx时提示盟主'" json:"fangka_tips_min_num"` // 房卡低于xx时提示盟主
	RecordTimeInterval int                       `gorm:"column:record_time_interval;default:12;comment:'战绩筛选时段单位'" json:"record_time_interval"`   // 战绩筛选时段单位
	NewTableSortType   int                       `gorm:"column:new_table_sort_type" json:"new_table_sort_type"`                                   // 新版本桌子排序类型 0-无排序 1-空/未/满 2-空/满/未 3-未/空/满 4-未/满/空 5-满/空/未 6-满/未/空
	CreateTableType    int                       `gorm:"column:create_table_type" json:"create_table_type"`                                       //开桌类型  0-人满开桌  1-另开新卓
	RankRound          int                       `gorm:"column:rank_round;default:0" json:"rank_round"`                                           // Rank排行榜 对应  0001 0010 0100 对应三个复选框
	RankWiner          int                       `gorm:"column:rank_winer;default:0" json:"rank_winer"`
	RankRecord         int                       `gorm:"column:rank_record;default:0" json:"rank_record"`
	RankOpen           bool                      `gorm:"column:rank_open;default:0" json:"rank_open"`               // 开启排行榜 true 开启   false 关闭
	IsNotEft2PTale     bool                      `gorm:"column:is_not_effect_2P;default:1" json:"is_not_effect_2P"` // 2人桌子禁止同桌不生效 是否勾选
	IsCostPartnerCard  bool                      `gorm:"column:is_cost_partner_card;default:0" json:"is_cost_partner_card"`
	NoSkipVitaminSet   bool                      `gorm:"column:no_skip_vitamin_set;default:0" json:"no_skip_vitamin_set"` // 是否禁止跨级调整vitamin
	MinTableNum        int                       `gorm:"column:min_table_num;default:0" json:"min_table_num"`             // 最小桌数
	MaxTableNum        int                       `gorm:"column:max_table_num;default:0" json:"max_table_num"`             // 最大桌数
}

func (House) TableName() string {
	return "house"
}

func initHouse(db *gorm.DB) error {
	var err error
	if db.HasTable(&House{}) {
		err = db.AutoMigrate(&House{}).Error
	} else {
		err = db.CreateTable(&House{}).Error
	}
	return err
}

type HouseLog struct {
	Id        int64     `gorm:"primary_key;column:id"`           //! id
	DHId      int64     `gorm:"column:dhid"`                     //! 包厢id雨
	HId       int       `gorm:"column:hid"`                      //! 包厢id
	Area      string    `gorm:"column:area"`                     //! 包厢区域编码
	UId       int64     `gorm:"column:uid"`                      //! 玩家id
	Type      int       `gorm:"column:type"`                     //! 类型 0创建 1删除
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseLog) TableName() string {
	return "house_log"
}

func initHouseLog(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseLog{}) {
		err = db.AutoMigrate(&HouseLog{}).Error
	} else {
		err = db.CreateTable(&HouseLog{}).Error
	}
	return err
}

type HouseTableLimitUser struct {
	Id        int64     `gorm:"column:id;comment:'主键'"`
	Hid       int       `gorm:"column:hid;comment:'包厢id'"`
	GroupID   int       `gorm:"column:group_id;comment:'分组id'"`
	Uid       int64     `gorm:"column:uid;comment:'用户id'"`
	Status    int       `gorm:"column:status;comment:'状态，0为正常，1为失效'"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime"` // 更新时间
}

func (HouseTableLimitUser) TableName() string {
	return "house_table_limit_user"
}
func initHouseTableLimitUsers(db *gorm.DB) error {
	var err error
	if db.HasTable(&HouseTableLimitUser{}) {
		err = db.AutoMigrate(&HouseTableLimitUser{}).Error
	} else {
		err = db.CreateTable(&HouseTableLimitUser{}).Error
	}
	return err
}

type HouseVitaminPoolLog struct {
	Id        int64             `gorm:"id"`
	Hid       int               `gorm:"hid"`
	Opuid     int64             `gorm:"opuid"`
	Value     int64             `gorm:"value"`
	After     int64             `gorm:"after"`
	Optype    VitaminChangeType `gorm:"optype"`
	Status    int               `gorm:"status"`
	Extra     string            `gorm:"extra"`       // 结算详情时间节点记录
	CreatedAt time.Time         `gorm:"created_at"`  // 创建时间
	UpdatedAt time.Time         `gorm:"update_time"` // 更新时间
}

type CountGorm struct {
	Count int `gorm:"gcount"`
}

// PoolChange 包厢疲劳值变更记录
func HousePoolChange(dhid int64, optUid int64, optType VitaminChangeType, value int64, gameNum string, tx *gorm.DB) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	if optType == HouseCreate {
		sql := `select count(1) as  gcount from house_vitamin_pool_log where hid = ?`
		var count int
		tx.Raw(sql, dhid).Count(&count)
		if count >= 1 {
			return nil
		}
		xlog.Logger().Warnf("init house:%d, vitamin pool:%d", dhid, value)

	}
	sql1 := `insert house_vitamin_pool(hid,after) values(?,?) ON DUPLICATE KEY UPDATE after = after + ? `
	sql2 := `insert into %s(hid,opuid,value,optype,game_num,after) 
	values(?,?,?,?,?,(select after from house_vitamin_pool where hid = ? order by id desc limit 1))`
	if optType == BigWinCost || optType == GamePay {
		sql2 = fmt.Sprintf(sql2, "house_vitamin_pool_tax_log")
	} else {
		if optType == TaxSum { //税收统计已记录
			return nil
		}
		sql2 = fmt.Sprintf(sql2, "house_vitamin_pool_log")
	}
	err := tx.Exec(sql1, dhid, value, value).Error
	if err != nil {
		xlog.Logger().Errorf("%v", err)
		return err
	}
	err = tx.Exec(sql2, dhid, optUid, value, optType, gameNum, dhid).Error
	if err != nil {
		xlog.Logger().Errorf("%v", err)
		return err
	}

	return nil
}
