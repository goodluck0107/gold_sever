package models

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"time"

	"github.com/jinzhu/gorm"
)

type UserHeadInfo struct {
	Id       int64  `gorm:"primaryKey;column:id"`
	Nickname string `redis:"nickname" gorm:"column:nickname;size:100;index:idx_nickname"` // 昵称
	Imgurl   string `redis:"imgurl" gorm:"column:imgurl;size:512"`                        // 头像
	Sex      int    `redis:"sex" gorm:"column:sex"`                                       // 性别
}

// 用户表
type User struct {
	Id                int64      `gorm:"primary_key;column:id"`
	Nickname          string     `gorm:"column:nickname;size:100"`                           // 昵称
	Imgurl            string     `gorm:"column:imgurl;size:255"`                             // 头像
	Guestid           string     `gorm:"column:guest_id;size:32"`                            // 游客id
	Openid            string     `gorm:"column:open_id;size:255"`                            // openid
	Unionid           string     `gorm:"column:union_id;size:64;index:idx_wechat"`           // Unionid
	Card              int        `gorm:"column:card;default:0"`                              // 房卡数量
	FrozenCard        int        `gorm:"column:frozencard;default:0"`                        // 冻结房卡数量
	Gold              int        `gorm:"column:gold;default:0"`                              // 金币数量
	GoldBean          int        `gorm:"column:gold_bean;default:0"`                         // 礼券
	InsureGold        int        `gorm:"column:insure_gold;default:0"`                       // 保险箱金币
	Revenue           int        `gorm:"column:revenue;default:0"`                           // 茶水费总消耗
	WinCount          int        `gorm:"column:win_count;default:0"`                         // 胜利局数
	LostCount         int        `gorm:"column:lost_count;default:0"`                        // 失败局数
	DrawCount         int        `gorm:"column:draw_count;default:0"`                        // 和局局数
	FleeCount         int        `gorm:"column:flee_count;default:0"`                        // 逃跑局数
	TotalCount        int        `gorm:"column:total_count;default:0"`                       // 总局数
	DescribeInfo      string     `gorm:"column:describe_info;size:255"`                      // 个性签名
	Sex               int        `gorm:"column:sex"`                                         // 性别
	Tel               string     `gorm:"column:tel;size:32"`                                 // 手机号
	ReName            string     `gorm:"column:rename;size:50;default:null"`                 // 真实姓名
	Idcard            string     `gorm:"column:idcard;size:32;default:null"`                 // 身份证号
	Games             string     `gorm:"column:games;type:text"`                             // 用户游戏列表
	Token             string     `gorm:"column:token;size:32"`                               // 登录凭证
	IsBlack           int8       `gorm:"column:is_black"`                                    // 封号类型, 0为不封号
	UserType          int        `gorm:"column:user_type;index:idx_wechat"`                  // 用户注册类型(区分用户是哪种类型的账号)
	Password          string     `gorm:"column:password;size:32"`                            // 手机注册账号特有
	Platform          int        `gorm:"column:platform"`                                    // 注册平台
	Ip                string     `gorm:"column:ip;size:32"`                                  // 注册ip
	MachineCode       string     `gorm:"column:machine_code;size:32"`                        // 机器码
	Area              string     `gorm:"column:area;size:32"`                                // 区域码
	LastLoginIp       string     `gorm:"column:last_login_ip;size:32"`                       // 最后一次登录ip
	LastLoginAt       *time.Time `gorm:"column:last_login_at;type:datetime;default:Now()"`   // 最后一次登录时间
	LastOffLineAt     *time.Time `gorm:"column:last_offline_at;type:datetime;default:Now()"` // 最后一次离线时间
	Origin            int8       `gorm:"column:origin;default:0"`                            // 用户来源
	DeliveryImg       string     `gorm:"column:delivery_img;default:''"`                     // 用户个人名片图片所在地址
	CreatedAt         time.Time  `gorm:"column:created_at;type:datetime"`                    // 创建时间
	UpdatedAt         time.Time  `gorm:"column:updated_at;type:datetime"`                    // 更新时间
	ContributionScore int64      `gorm:"column:contributionScore;default:0"`                 // 玩家贡献值
	Diamond           int        `gorm:"column:diamond"`                                     // 钻石
	RefuseInvite      bool       `gorm:"column:refuse_invite"`                               // 拒绝入圈邀请
	AccountType       int        `gorm:"column:account_type;default:0"`                      // 账号状态  0 是正常  1是注销
	Area2nd           string     `gorm:"column:area2nd;default:''"`                          // 第二区域码（gps区域码）
	Area3rd           string     `gorm:"column:area3rd;default:''"`                          // 第三区域码（包厢位置区域码）
	IdcardAuthPI      string     `gorm:"column:idcard_auth_pi;default:''"`                   // 实名认证的官方返回的唯一认证码pi
	IsVip             bool       `gorm:"column:is_vip"`
}

func (User) TableName() string {
	return "user"
}

func initUser(db *gorm.DB) error {
	var err error
	if db.HasTable(&User{}) {
		err = db.AutoMigrate(&User{}).Error
	} else {
		err = db.CreateTable(&User{}).Error
		if err == nil {
			// 修改递增初始值
			err = db.Exec("alter table user AUTO_INCREMENT=1000000").Error
		}
	}
	return err
}

// db -> redis模型
func (u *User) ConvertModel() *static.Person {
	p := new(static.Person)
	p.Uid = u.Id
	p.Area = u.Area
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
	p.DeliveryImg = u.DeliveryImg
	p.Platform = u.Platform
	if u.LastLoginAt == nil {
		p.LastLoginTime = time.Now().Unix()
	} else {
		p.LastLoginTime = u.LastLoginAt.Unix()
	}
	if u.LastOffLineAt == nil {
		p.LastOffLineTime = time.Now().Unix()
	} else {
		p.LastOffLineTime = u.LastOffLineAt.Unix()
	}
	p.ContributionScore = u.ContributionScore
	p.Diamond = u.Diamond
	p.RefuseInvite = u.RefuseInvite
	p.AccountType = u.AccountType
	p.Area2nd = u.Area2nd
	p.Area3rd = u.Area3rd
	p.IdcardAuthPI = u.IdcardAuthPI
	p.IsVip = u.IsVip
	return p
}

// ! 查询新增玩家总数
func FangKaNewUserGetCountByTime(db *gorm.DB, time time.Time) (int, error) {
	var count = 0
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(User{}).Where(`date_format(created_at, "%Y-%m-%d") = ?`, selectStr).Count(&count).Error
	return count, err
}

// 查询房卡场时间节点玩家总数
func FangKaUserGetCountByTime(db *gorm.DB, time time.Time) (int, error) {
	var count = 0
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(User{}).Where(`date_format(created_at, "%Y-%m-%d") <= ?`, selectStr).Count(&count).Error
	return count, err
}

// 查询金币场时间节点玩家新增总数(统计方式2)
func GoldNewUserNewStataGetCountByTime(db *gorm.DB, time time.Time) (int, error) {
	selectStr := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())

	sql1 := `select count(*) as count from (
			select id as uid from user where date_format(created_at, "%Y-%m-%d") = ? and
			id not in (select uid from house_member where date_format(created_at, "%Y-%m-%d") = ?) and
			id in (select uid from gold_game_retention_statistics where date_format(gamedate, "%Y-%m-%d") = ?)) as tb`

	sql2 := `select count(*) as count from (
			select id as uid from user where date_format(created_at, "%Y-%m-%d") = ? and
			id not in (select uid from house_member where date_format(created_at, "%Y-%m-%d") = ?) and
			id not in (select uid from gold_game_retention_statistics where date_format(gamedate, "%Y-%m-%d") = ?) and
			id not in (select uid from record_game_total where date_format(created_at, "%Y-%m-%d") = ? and hid = 0)) as tb`

	/*
		if (未创建包厢 或者 加入包厢) {
			if （玩过金币场）
			{
				// 是否玩过房卡场不影响
				// 金币用户
			}
			else if (未玩过大厅好友房 且 未玩过金币场)
			{
				// 金币用户
			}
		}
	*/

	type QueryResult struct {
		Count int `gorm:"column:count"`
	}

	var result1, result2 QueryResult
	err := db.Raw(sql1, selectStr, selectStr, selectStr).Scan(&result1).Error

	if err == nil {
		err = db.Raw(sql2, selectStr, selectStr, selectStr, selectStr).Scan(&result2).Error
	}

	return result1.Count + result2.Count, err
}

func PlatformNewUserGetCountByTime(db *gorm.DB, time time.Time) (int, int, error) {
	var count1 = 0
	selectStr1 := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
	err := db.Model(User{}).Where(`date_format(created_at, "%Y-%m-%d") = ? and platform = 1 `, selectStr1).Count(&count1).Error

	var count2 = 0
	if err == nil {
		selectStr2 := fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
		err = db.Model(User{}).Where(`date_format(created_at, "%Y-%m-%d") = ? and platform = 2 `, selectStr2).Count(&count2).Error
	}

	return count1, count2, err
}

////! 注销用户 database 导致循环导包了
//func WriteOffUser(db *gorm.DB, dbR *database.DB_r, uid int64, accType int) (interface{}, *xerrors.CustomError) {
//	tx := db.Begin()
//	var user User
//	var err error
//	if err = tx.Where("AccountType = ? and id = ?", accType, uid).First(&user).Error; err == nil {
//		cuserror := xerrors.NewResultError("该账号已经被注销")
//		tx.Rollback()
//		return xerrors.ResultErrorCode, cuserror
//	}
//	if err = tx.Model(User{Id: uid}).Update("AccountType", accType).Error; err != nil {
//		syslog.Logger().Errorln(err)
//		cuserror := xerrors.NewResultError("账号注销失败")
//		tx.Rollback()
//		return xerrors.ResultErrorCode, cuserror
//	}
//	// 更新redis
//	err = dbR.UpdatePersonAttrs(uid, "AccountType", accType)
//	if err != nil {
//		syslog.Logger().Errorln("set user data to redis error: ", err.Error())
//		cuserror := xerrors.NewResultError("账号注销失败")
//		tx.Rollback()
//		return xerrors.ResultErrorCode, cuserror
//	}
//	tx.Commit()
//	return nil, xerrors.RespOk
//}
