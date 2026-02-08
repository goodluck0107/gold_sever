package models

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"time"
)

type ConfigServer struct {
	Webhz         string        `gorm:"column:webhouzui;size:255"`                              // web后缀
	NewCard       int           `gorm:"column:newcard;default:0"`                               // 新注册的用户赠送房卡
	NewGold       int           `gorm:"column:newgold;default:0"`                               // 新注册的用户赠送金币
	NewGoldTime   time.Time     `gorm:"column:newgold_time;default:CURRENT_TIMESTAMP;NOT NULL"` // 金币增送截至时间
	InsureendLine int           `gorm:"column:insureendline;default:20000"`                     // 保险箱最低限制
	BroadcastCost int           `gorm:"column:broadcast_cost;default:3000"`                     // 用户发送广播消耗金币
	AllowancesStr string        `gorm:"column:allowances;size:255;default:'{}'"`                // 低保
	Allowances    *Allowances   `gorm:"-"`
	DailyRewards  *DailyRewards `gorm:"-"`
	Version       int           `gorm:"column:version"`
	Minversion    int           `gorm:"column:minversion"`
	Viewversion   int           `gorm:"column:viewversion"`
	PayUrl        string        `gorm:"column:payurl;size:255"`                                                  //支付链接网址
	PkgUrl        string        `gorm:"column:pkgurl;size:255"`                                                  //最新安装包网址
	AppId         string        `gorm:"column:appid;size:50"`                                                    //支付服务appid
	AppToken      string        `gorm:"column:apptoken;size:50"`                                                 //支付服务appid
	GetImgUrl     string        `gorm:"column:getimg_url;default:'http://apitpg.facai.com/xl/userimg';not null"` // 个人名片图片所在地址
	XianliaoHost  string        `gorm:"column:xianliao_host;default:'https://apitpg.facai.com/xldqjs';not null"` // 个人名片图片所在地址
}

// 每日礼包奖励
type DailyRewards struct {
	AwardGold int `json:"award_gold"` // 奖励金币数量
	Status    int `json:"status"`     // 每日礼包奖励开关(1开0关)
}

// 每日低保
type Allowances struct {
	LimitGold int `json:"limit_gold"` // 当用户金币小于此金币数值, 则触发低保自动发放
	AwardGold int `json:"award_gold"` // 奖励金币数量
	Num       int `json:"num"`        // 每天奖励次数
}

func (ConfigServer) TableName() string {
	return "config_server"
}

func (c *ConfigServer) AfterFind() error {
	var err error
	if err = json.Unmarshal([]byte(c.AllowancesStr), &c.Allowances); err != nil {
		return err
	}
	return nil
}

type ConfigGameTypes struct {
	Id           int `gorm:"primary_key;column:id"`
	GameServerId int `gorm:"column:game_server_id"`             //! 服务器编号
	KindId       int `gorm:"column:kind_id"`                    //! 支持的子游戏类型
	GameType     int `gorm:"column:game_type;default:2"`        //! 比赛场 or 金币场 or 好友场
	SiteType     int `gorm:"column:site_type;default:0"`        //! 场次类型
	TableNum     int `gorm:"column:table_num;default:0"`        //! 牌桌数量
	MaxPeopleNum int `gorm:"column:max_people_num;default:100"` //! 最大限制进入人数
}

func (ConfigGameTypes) TableName() string {
	return "config_server_gametypes"
}

// 房间场次类型
type ConfigSite struct {
	Id        int                    `gorm:"primary_key;column:id"`
	Name      string                 `gorm:"column:name;size:32"`                         // 场次名
	KindId    int                    `gorm:"column:kind_id;unique_index:idx_config_site"` //! 子游戏id
	Type      int                    `gorm:"column:type;unique_index:idx_config_site"`    //! 类型(初级场 中级场 高级场)
	MinScore  int                    `gorm:"column:min_score;default:0"`                  //! 最低进场分数限制(0为不限制)
	MaxScore  int                    `gorm:"column:max_score;default:0"`                  //! 最高进场分数限制(0为不限制)
	ConfigStr string                 `gorm:"column:game_config;default:'{}'"`             //! 游戏参数配置
	BaseNum   int                    `gorm:"column:base_num;default:0"`                   //! 基础在线人数
	Config    map[string]interface{} `gorm:"-"`                                           //! 房间配置(底分、茶水)
	SitMode   int                    `gorm:"column:sit_mode;default:0"`                   //! 是否坐桌模式
	RobotMode int                    `gorm:"column:robot_mode;default:0"`                 //! 是否开启机器人模式
	RobotNum  int                    `gorm:"column:robot_num;default:0"`                  //! 机器人配置数量
}

func (ConfigSite) TableName() string {
	return "config_site"
}

func (c *ConfigSite) AfterFind() error {
	c.Config = make(map[string]interface{})
	return json.Unmarshal([]byte(c.ConfigStr), &c.Config)
}

type ConfigQiNiu struct {
	Id        int    `gorm:"primary_key;column:id"`
	UseQueue  bool   `gorm:"column:use_queue;default:1;comment:'是否使用队列处理'"`
	Able      bool   `gorm:"column:able;default:0;comment:'七牛头像开关'"` // 开关
	AccessKey string `gorm:"column:access_key;default:'hjQcnzaUpe-lhrdVNvEZSbSHdpD6QqADqnd2p7DW';comment:'七牛云AK'"`
	SecretKey string `gorm:"column:secret_key;default:'8U9QFmyGxGrm3nhLZPfxmIgydXVffwiwpYM5SvcF';comment:'七牛云SK'"`
	Bucket    string `gorm:"column:bucket;default:'chess-avatar';comment:'七牛云空间名称'"`
	Domain    string `gorm:"column:domain;default:'';comment:'七牛云空间域名'"`
	Public    bool   `gorm:"column:public;default:0;comment:'七牛云空间访问控制，1公告,0私有'"`
}

func (qn ConfigQiNiu) Verification(newCfg *ConfigQiNiu) bool {
	return qn.AccessKey == newCfg.AccessKey &&
		qn.SecretKey == newCfg.SecretKey &&
		qn.Able == qn.Able
}

func (ConfigQiNiu) TableName() string {
	return "config_qiniu"
}

// type ConfigWebPay struct {
// 	Id    int    `gorm:"primary_key;column:id"`
// 	Host  string `gorm:"column:host;comment:'Host'"`
// 	AppId string `gorm:"column:appid;comment:'AppId'"`
// 	Token string `gorm:"column:token;comment:'Token'"`
// }
//
// func (ConfigWebPay) TableName() string {
// 	return "config_webpay"
// }

func initConfig(db *gorm.DB) error {
	var err error
	if db.HasTable(&ConfigServer{}) {
		err = db.AutoMigrate(&ConfigServer{}).Error
	} else {
		err = db.CreateTable(&ConfigServer{}).Error
	}
	if db.HasTable(&ConfigGameTypes{}) {
		err = db.AutoMigrate(&ConfigGameTypes{}).Error
	} else {
		err = db.CreateTable(&ConfigGameTypes{}).Error
	}
	if db.HasTable(&ConfigSite{}) {
		err = db.AutoMigrate(&ConfigSite{}).Error
	} else {
		err = db.CreateTable(&ConfigSite{}).Error
	}
	if db.HasTable(&ConfigQiNiu{}) {
		err = db.AutoMigrate(&ConfigQiNiu{}).Error
	} else {
		err = db.CreateTable(&ConfigQiNiu{}).Error
	}
	// if db.HasTable(&ConfigWebPay{}) {
	// 	err = db.AutoMigrate(&ConfigWebPay{}).Error
	// } else {
	// 	err = db.CreateTable(&ConfigWebPay{}).Error
	// }
	return err
}

func (c *ConfigServer) String() string {
	return fmt.Sprintf("\n[---------------%s--------------]:\n[Webhz:]:%s\n[NewCard:]%d\n[Viewversion:]%d\n[Minversion:]%d\n[Viewversion:]%d\n[PayUrl:]%s\n[PkgUrl:]%s\n", c.TableName(), c.Webhz, c.NewCard, c.Viewversion, c.Minversion, c.Viewversion, c.PayUrl, c.PkgUrl)
}

func (c ConfigGameTypes) String() string {
	return fmt.Sprintf("[%s-->][id:]<%d>;[GameServerId:]<%d>;[Type:]<%d>;[GameType:]<%d>;[SiteType:]<%d>;[TableNum:]<%d>;",
		c.TableName(), c.Id, c.GameServerId, c.KindId, c.GameType, c.SiteType, c.TableNum)
}
