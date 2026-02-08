package models

import (
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"
)

//机器人配置
type ConfigRobot struct {
	IsOpen     int `gorm:"column:IsOpen"`      //! 机器人开关
	Site1Count int `gorm:"column:site1_count"` //! 初级房机器人数量配置
	Site2Count int `gorm:"column:site2_count"` //! 中级房机器人数量配置
	Site3Count int `gorm:"column:site3_count"` //! 高级房机器人数量配置
}

func (ConfigRobot) TableName() string {
	return "config_robot"
}

// 用户表
type Robot struct {
	Mid               int64      `gorm:"primary_key;column:mid"`                   // 机器人mid
	Nickname          string     `gorm:"column:nickname;size:100"`                 // 昵称
	Imgurl            string     `gorm:"column:imgurl;size:255"`                   // 头像
	Guestid           string     `gorm:"column:guest_id;size:32"`                  // 游客id
	Openid            string     `gorm:"column:open_id;size:50"`                   // 微信openid
	Unionid           string     `gorm:"column:union_id;size:50;index:idx_wechat"` // Unionid
	Card              int        `gorm:"column:card;default:0"`                    // 房卡数量
	FrozenCard        int        `gorm:"column:frozencard;default:0"`              // 冻结房卡数量
	Gold              int        `gorm:"column:gold;default:0"`                    // 金币数量
	GoldBean          int        `gorm:"column:gold_bean;default:0"`               // 礼券
	InsureGold        int        `gorm:"column:insure_gold;default:0"`             // 保险箱金币
	Revenue           int        `gorm:"column:revenue;default:0"`                 // 茶水费总消耗
	WinCount          int        `gorm:"column:win_count;default:0"`               // 胜利局数
	LostCount         int        `gorm:"column:lost_count;default:0"`              // 失败局数
	DrawCount         int        `gorm:"column:draw_count;default:0"`              // 和局局数
	FleeCount         int        `gorm:"column:flee_count;default:0"`              // 逃跑局数
	TotalCount        int        `gorm:"column:total_count;default:0"`             // 总局数
	DescribeInfo      string     `gorm:"column:describe_info;size:255"`            // 个性签名
	Sex               int        `gorm:"column:sex"`                               // 性别
	Tel               string     `gorm:"column:tel;size:32"`                       // 手机号
	ReName            string     `gorm:"column:rename;size:50;default:null"`       // 真实姓名
	Idcard            string     `gorm:"column:idcard;size:32;default:null"`       // 身份证号
	Games             string     `gorm:"column:games;type:text"`                   // 用户游戏列表
	Token             string     `gorm:"column:token;size:32"`                     // 登录凭证
	IsBlack           int8       `gorm:"column:is_black"`                          // 封号类型, 0为不封号
	UserType          int        `gorm:"column:user_type;index:idx_wechat"`        // 用户注册类型(区分用户是哪种类型的账号)
	Password          string     `gorm:"column:password;size:32"`                  // 手机注册账号特有
	Platform          int        `gorm:"column:platform"`                          // 注册平台
	Ip                string     `gorm:"column:ip;size:32"`                        // 注册ip
	MachineCode       string     `gorm:"column:machine_code;size:32"`              // 机器码
	LastLoginIp       string     `gorm:"column:last_login_ip;size:32"`             // 最后一次登录ip
	LastLoginAt       *time.Time `gorm:"column:last_login_at;type:datetime"`       // 最后一次登录时间
	Origin            int8       `gorm:"column:origin;default:0"`                  // 用户来源
	CreatedAt         time.Time  `gorm:"column:created_at;type:datetime"`          // 创建时间
	UpdatedAt         time.Time  `gorm:"column:updated_at;type:datetime"`          // 更新时间
	UnionCode         string     `gorm:"column:union_code;size:50;index:idx_h5"`   // 联运id
	ContributionScore int64      `gorm:"column:contributionScore;default:0"`       // 玩家贡献值
	IsUsed            int        `gorm:"column:is_used;default:0"`                 // 本机器人信息是否已被使用是否已使用(0未使用1使用)
}

func (Robot) TableName() string {
	return "robot"
}

func initRobot(db *gorm.DB) error {
	var err error
	if db.HasTable(&Robot{}) {
		err = db.AutoMigrate(&Robot{}).Error
	} else {
		err = db.CreateTable(&Robot{}).Error
		if err == nil {
			// 修改递增初始值
			err = db.Exec("alter table robot AUTO_INCREMENT=3000").Error
		}
	}
	if db.HasTable(&ConfigRobot{}) {
		err = db.AutoMigrate(&ConfigRobot{}).Error
	} else {
		err = db.CreateTable(&ConfigRobot{}).Error
	}

	if db.HasTable(&ConfigRobotJoinCtrl{}) {
		err = db.AutoMigrate(&ConfigRobotJoinCtrl{}).Error
	} else {
		err = db.CreateTable(&ConfigRobotJoinCtrl{}).Error
	}

	if db.HasTable(&ConfigRobotPlayCtrl{}) {
		err = db.AutoMigrate(&ConfigRobotPlayCtrl{}).Error
	} else {
		err = db.CreateTable(&ConfigRobotPlayCtrl{}).Error
	}

	return err
}

// robot -> person模型
func (u *Robot) Convert2Person() *static.Person {
	p := new(static.Person)
	p.Uid = u.Mid
	p.UserType = u.UserType
	p.Nickname = u.Nickname
	p.Imgurl = u.Imgurl
	p.Guestid = u.Guestid
	p.Openid = u.Openid
	p.Card = u.Card
	p.Gold = u.Gold
	p.InsureGold = u.InsureGold
	p.GoldBean = u.GoldBean
	p.FrozenCard = u.FrozenCard
	p.DescribeInfo = u.DescribeInfo
	p.Sex = u.Sex
	p.Tel = u.Tel
	p.ReName = u.ReName
	p.Idcard = u.Idcard
	p.Token = u.Token
	p.Games = u.Games
	p.IsBlack = u.IsBlack
	p.CreateTime = u.CreatedAt.Unix()
	p.WinCount = u.WinCount
	p.LostCount = u.LostCount
	p.DrawCount = u.DrawCount
	p.FleeCount = u.FleeCount
	p.TotalCount = u.TotalCount
	p.UnionCode = u.UnionCode
	p.ContributionScore = u.ContributionScore
	p.IsRobot = 1
	return p
}

// db -> redis
func (u *Robot) ConvertModel() *static.Robot {
	r := new(static.Robot)
	r.Mid = u.Mid
	r.Nickname = u.Nickname
	r.Imgurl = u.Imgurl
	r.Guestid = u.Guestid
	r.Openid = u.Openid
	r.Card = u.Card
	r.FrozenCard = u.FrozenCard
	r.Gold = u.Gold
	r.GoldBean = u.GoldBean
	r.InsureGold = u.InsureGold
	r.Sex = u.Sex
	r.Tel = u.Tel
	r.ReName = u.ReName
	r.Idcard = u.Idcard
	r.Token = u.Token
	r.Games = u.Games
	r.IsBlack = u.IsBlack
	r.CreateTime = u.CreatedAt.Unix()
	r.DescribeInfo = u.DescribeInfo
	r.UserType = u.UserType
	r.Platform = u.Platform
	r.WinCount = u.WinCount
	r.LostCount = u.LostCount
	r.DrawCount = u.DrawCount
	r.FleeCount = u.FleeCount
	r.TotalCount = u.TotalCount
	r.UnionCode = u.UnionCode
	r.MachineCode = u.MachineCode
	r.ContributionScore = u.ContributionScore
	r.IsRobot = 1

	return r
}

//机器人工作状态
type RobotWorkDetail struct {
	Mid      int64 `gorm:"primary_key;column:mid"` //! 机器人mid
	KindId   int   `gorm:"column:kind_id"`         //! 子游戏id
	SiteType int   `gorm:"column:site_type"`       //! 场次类型
	GameId   int   `gorm:"column:game_id"`         //! 游戏进程id
	Ntid     int   `gorm:"column:ntid"`            //! 桌号
}

func (RobotWorkDetail) TableName() string {
	return "robot_wrok_detail"
}

func initRobotWorkDetail(db *gorm.DB) error {
	var err error
	if db.HasTable(&RobotWorkDetail{}) {
		err = db.AutoMigrate(&RobotWorkDetail{}).Error
	} else {
		err = db.CreateTable(&RobotWorkDetail{}).Error
	}
	return err
}

//机器人入场规则
type ConfigRobotJoinCtrl struct {
	KindId       int `gorm:"column:kind_id"`       //! 子游戏id
	SiteType     int `gorm:"column:site_type"`     //!
	Num          int `gorm:"column:num"`           //! 默认机器人：当用户进入房间时，如果房间为空，则：默认加d名机器人进入该房间,初始设定：随机1~2(d为一个随机范围)
	TimeInterval int `gorm:"column:time_interval"` //! 增加机器人：每隔时间a，加一名机器人进入该房间,数值a的初始设定：随机5~10s
	CancelNum    int `gorm:"column:cancel_num"`    //! 达到c人时，则：停止添加机器人
}

func (ConfigRobotJoinCtrl) TableName() string {
	return "config_robot_joinctrl"
}

//机器人发牌、出牌控制策略
type ConfigRobotPlayCtrl struct {
	KindId          int `gorm:"column:kind_id"`          //! 子游戏id
	SiteType        int `gorm:"column:site_type"`        //!
	DispatchMaxRate int `gorm:"column:dispatchmax_rate"` //发最大牌几率(0关闭,1-100之间,随机发牌几率为100减去最大牌几率)
	PassTime        int `gorm:"column:pass_time"`        //不要出牌操作时间时间,单位秒
	OutTime         int `gorm:"column:out_time"`         //出牌时间,单位秒
	GuanPaiTime     int `gorm:"column:guanpai_time"`     //管牌出牌时间,单位秒
	WorkTime        int `gorm:"column:work_time"`        //机器人工作时间,单位秒
}

func (ConfigRobotPlayCtrl) TableName() string {
	return "config_robot_playctrl"
}
