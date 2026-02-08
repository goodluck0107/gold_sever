package origin

import (
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"strings"
	"time"
)

// 用户信息表
type User struct {
	UserId        int64     `gorm:"column:UserID"`        // 用户id
	Nickname      string    `gorm:"column:RegAccounts"`   // 昵称
	OpenId        string    `gorm:"column:OpenID"`        // 微信openid
	UnionId       string    `gorm:"column:UnionID"`       // 微信UnionId
	FaceUrl       string    `gorm:"column:FaceUrl"`       // 头像
	Phone         string    `gorm:"column:Phone"`         // 手机号
	Gender        int       `gorm:"column:Gender"`        // 性别,1和"1"是男,其他是女
	RegisterIp    string    `gorm:"column:RegisterIP"`    // 注册ip
	LastLogonIp   string    `gorm:"column:LastLogonIP"`   // 上次登录ip
	RegisterDate  time.Time `gorm:"column:RegisterDate"`  // 注册时间
	LastLogonDate time.Time `gorm:"column:LastLogonDate"` // 上次登录时间
	IsGuest       int       `gorm:"column:IsGuest"`       // 游客账户：1； 手机账号：0, 3（绑定了微信头像昵称）； 微信账号：2
	MachineSerial string    `gorm:"column:MachineSerial"` // 游客唯一设备码(当账号为游客时需要迁移)\
	Nullity       bool      `gorm:"column:Nullity"`       // 是否被封号
	UnderWrite    string    `gorm:"column:UnderWrite"`    // 注册渠道
	Password      string    `gorm:"column:LogonPass"`     // 登录密码
	MachineCode   string    `gorm:"column:machineCode"`   // 机器码
}

// 表名
func (User) TableName() string {
	return "AccountsInfo"
}

// 数据库名
func (User) Database() string {
	return "QPGameUserDB"
}

// 模型转换
func (u *User) Convert() *models.User {
	user := new(models.User)
	user.Id = u.UserId
	user.Nickname = u.Nickname
	user.Imgurl = u.FaceUrl
	if u.IsGuest == 1 {
		user.Guestid = strings.TrimSpace(u.MachineSerial)
		user.UserType = consts.UserTypeYk
	} else if u.IsGuest == 2 {
		user.UserType = consts.UserTypeWechat
	} else if u.IsGuest == 3 {
		user.UserType = consts.UserTypeMobile2
	} else if u.IsGuest == 0 {
		user.UserType = consts.UserTypeMobile
	}
	user.Openid = u.OpenId
	user.Unionid = u.UnionId
	user.Password = u.Password
	// 1和"1"是男
	if u.Gender == 1 || u.Gender == 49 {
		user.Sex = consts.SexMale
	} else {
		user.Sex = consts.SexFemale
	}
	user.Tel = u.Phone
	//user.Idcard = u.IDCard
	//user.ReName = u.IDName
	user.Ip = u.RegisterIp
	user.CreatedAt = u.RegisterDate
	user.LastLoginIp = u.LastLogonIp
	user.LastLoginAt = &u.LastLogonDate
	if u.Nullity {
		user.IsBlack = consts.BlackStatusForbiddenLogin
	}
	if strings.HasSuffix(u.UnderWrite, "0001") {
		user.Platform = consts.PlatformAndroid
	} else if strings.HasSuffix(u.UnderWrite, "2001") {
		user.Platform = consts.PlatformIos
	} else {
		user.Platform = consts.PlatformWeb
		// 无法识别的渠道
		//log.Println("unknown platform: ", u.UnderWrite)
	}
	user.MachineCode = u.MachineCode
	return user
}
