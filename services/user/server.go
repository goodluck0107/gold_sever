package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/rpc"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/sirupsen/logrus"

	"sync"
	"time"

	"github.com/open-source/game/chess.git/pkg/models/origin"
	"github.com/open-source/game/chess.git/pkg/static/chanqueue"
	"github.com/open-source/game/chess.git/pkg/static/chanqueue/proto"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/util"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"github.com/jinzhu/gorm"
)

type Config struct {
	Host                  string `json:"host"`     //! 登錄服务器ip
	Center                string `json:"center" `  //! hall服务器ip
	InCenter              string `json:"incenter"` //! hall服务器内网ip
	Gate                  string `json:"gate"`     //! gate服务器ip
	Api                   string `json:"api"`      //! api服务器地址
	Redis                 string `json:"redis"`    //! redis
	RedisDB               int    `json:"redisdb"`
	RedisAuth             string `json:"redisauth"`       //! redis密码
	DB                    string `json:"db"`              //! 数据库
	SafeHost              string `json:"safehost"`        //! 服务器ip_高仿地址
	SafeHall              string `json:"safehall"`        //! 大厅服务器ip_高仿地址
	UseSafeIp             int8   `json:"usesafeip"`       //! 是否用高仿地址：0不用，1用
	EnabledYk             bool   `json:"enabledyk"`       //! 是否允许游客注册账号
	Encode                int    `json:"encode"`          // 消息加密方式 0为不加密(使用加密后客户端发的消息也必须加密, 否则忽略)
	EncodeClientKey       string `json:"encodeclientkey"` // 与客户端通信的aes加密秘钥
	EncodePhpKey          string `json:"encodephpkey"`    // 与php通信的aes加密秘钥
	service2.CommonConfig        // 第三方服务配置
}

type ServerQueue struct {
	Head string
	V    interface{}
}

type Server struct {
	Con          *Config                 //! 配置
	ConApp       []*models.ConfigApp     //! 渠道配置
	ConServers   *models.ConfigServer    //！服务器配置
	ConSqlServer *models.ConfigSqlserver //! sqlserver配置
	ConQiNiu     *models.ConfigQiNiu     //! 七牛云配置
	Wait         *sync.WaitGroup         //! 同步阻塞
	ShutDown     bool                    //! 是否正在执行关闭
	// RpcHall   *network.RPCClient
	HallServer string
	GateServer string
	BlackIds   map[int64]int8
	WechatLock *lock2.RWMutex // 微信登录锁
	Queue      *chanqueue.Queue
}

func (s *Server) Host() (string, string) {
	ip := strings.Split(s.Con.Host, ":")
	return ip[0], ip[1]
}

func (s *Server) Start(ctx context.Context) error {
	// 第三方服务初始化
	service2.InitConfig(s.Con.CommonConfig)
	// 单独进程处理
	s.LoadDBData()
	//注册玩家ID
	s.RegistUserID()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	//! 告诉服务器关闭
	var msg static.Msg_Null
	GetServer().CallHall("NewServerMsg", "loginserverclose", &msg)
	s.ShutDown = true
	if s.Queue != nil {
		xlog.AnyErrors(s.Queue.Close(), "close wuhan message queue error")
	}
	s.Wait.Wait()
	GetDBMgr().Close()
	xlog.Logger().Warn("wuhan shutdown")
	return nil
}

func (s *Server) RegisterRouter(ctx context.Context) {
	http.HandleFunc("/service", Service)
}

func (s *Server) RegisterRPC(ctx context.Context) {
	rpc.NewRpcxServer(s.Con.Host, "ServerMethod", new(ServerMethod), nil)
}

func (s *Server) LoadConfig(ctx context.Context) error {
	return GetDBMgr().ReadAllConfig()
}

func (s *Server) LoadServerConfig(ctx context.Context) error {
	if err := static.LoadServeYaml(s, true); err != nil {
		xlog.Logger().Error("InitConfig_yaml:", err)
		return err
	} else {
		xlog.Logger().Debug("load wuhan yaml etc succeed : ", fmt.Sprintf("%+v", s.Con))
	}
	return nil
}

func (s *Server) SetLoggerLevel() {
	xlog.SetFileLevel("debug")
}

func (s *Server) Run(ctx context.Context) {
	go s.RunQueue()
}

func (s *Server) Handle(string) error {
	return nil
}

func (s *Server) IsMulti() bool {
	return false
}

func (s *Server) ErrorDeep() int {
	return 3
}

func (s *Server) Name() string {
	return static.ServerNameUser
}

func (s *Server) Etc() interface{} {
	return s.Con
}

var serverSingleton *Server = nil

// ! 得到服务器指针
func GetServer() *Server {
	if serverSingleton == nil {
		serverSingleton = new(Server)
		serverSingleton.ShutDown = false
		serverSingleton.Con = new(Config)
		serverSingleton.ConServers = new(models.ConfigServer)
		serverSingleton.ConSqlServer = new(models.ConfigSqlserver)
		serverSingleton.ConQiNiu = new(models.ConfigQiNiu)
		serverSingleton.Wait = new(sync.WaitGroup)
		serverSingleton.WechatLock = new(lock2.RWMutex)
		serverSingleton.BlackIds = make(map[int64]int8)
	}

	return serverSingleton
}

// ! 调用大厅服
func (s *Server) CallHall(method string, msghead string, v interface{}) ([]byte, error) {
	args := &static.Msg_MsgBase{}
	buf, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	args.Data = string(buf)
	args.Head = msghead
	buf, err = json.Marshal(args)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	rep := CallHall(&static.Rpc_Args{MsgData: buf})
	return rep, nil
}

// 游客登录
func (s *Server) loginByYK(guestid string, platform int, ip string, machineCode string) (*models.User, *static.Person, bool, error) {
	var err error
	var person *static.Person
	var user models.User
	isNew := false
	// 判断用户是否存在
	if err = GetDBMgr().db_M.Model(user).Where("guest_id = ? and user_type = ? and account_type = 0", guestid, consts.UserTypeYk).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			// 异常报错
			xlog.Logger().Errorln("create user data failed: ", err.Error())
			return nil, nil, false, err
		}

		// 判断是否允许游客注册
		if !GetServer().Con.EnabledYk {
			err = errors.New("禁止游客注册")
			xlog.Logger().Errorln(err)
			return nil, nil, false, err
		}

		// 如果不存在则创建用户
		xlog.Logger().Errorln("新用戶", guestid)
		isNew = true
		var uid int64
		uid, err = s.AllocUserID()
		if err != nil {
			err = errors.New("生成的账号已经用完")
			xlog.Logger().Errorln(err)
			return nil, nil, false, err
		}
		user.Id = uid
		user.Nickname = getGuestRandomNickname()
		user.Imgurl = static.GenRobotUrl()
		user.Guestid = guestid
		user.Sex = consts.SexFemale // 默认女性
		user.Platform = platform
		user.Ip = ip
		user.MachineCode = machineCode
		user.UserType = consts.UserTypeYk
		person, err = s.creatUserData(&user)
		if err != nil {
			xlog.Logger().Errorln("create user data failed: ", err.Error())
			return nil, nil, false, err
		}
		xlog.Logger().Errorln("login creat youke user:", person.Uid, guestid)
	} else {
		// 更新玩家得机器码
		if !static.HF_IsValidMachineCode(user.MachineCode) && static.HF_IsValidMachineCode(machineCode) {
			user.MachineCode = machineCode
			if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
				xlog.Logger().Errorln("update user data failed: ", err.Error())
				return nil, nil, false, errors.New("更新用户数据失败")
			}
		}

		// 更新msyql
		user.Platform = platform
		if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
			xlog.Logger().Errorln("update user data failed: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}

		// 存在 去redis取
		person, err = GetDBMgr().db_R.GetPerson(user.Id)
		if err != nil {
			xlog.Logger().Errorln("get user from redis failed: ", err.Error())
			return nil, nil, false, err
		}

		// 更新redis
		person.Platform = user.Platform
		err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Platform", platform)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			return nil, nil, false, err
		}
	}

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, nil, false, err
	}

	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, nil, false, err
	}

	return &user, person, isNew, nil
}

// 手机登录
func (s *Server) loginByMobile(mobile string, platform int, ip string, machineCode string, allowRegister bool) (*models.User, *static.Person, bool, error) {
	isNew := false
	var err error
	// 判断用户是否存在
	var person *static.Person
	var user models.User
	xlog.Logger().Errorln("mobile is1: ", mobile)
	if err = GetDBMgr().db_M.Model(user).Where("tel = ? and account_type = 0", mobile).First(&user).Error; err != nil {
		if !allowRegister {
			// 不允许注册, 不存在则提示绑定微信
			return nil, nil, false, nil
		} else {
			// 注册账号
			isNew = true
			var uid int64
			uid, err = s.AllocUserID()
			if err != nil {
				err = errors.New("生成的账号已经用完")
				xlog.Logger().Errorln(err)
				return nil, nil, false, err
			}
			user.Id = uid
			user.Nickname = getGuestRandomNickname()
			user.Imgurl = static.GenRobotUrl()
			user.Tel = mobile
			user.Sex = consts.SexFemale // 默认女性
			user.Platform = platform
			user.Ip = ip
			user.MachineCode = machineCode
			user.UserType = consts.UserTypeMobile
			person, err = s.creatUserData(&user)
			xlog.Logger().Errorln("mobile is2: ", person.Uid, user.Tel)
			if err != nil {
				xlog.Logger().Errorln("create user data failed: ", err.Error())
				return nil, nil, false, err
			}
			xlog.Logger().Errorln("login create mobile user:", person.Uid, mobile)
		}
	}

	// 更新mysql
	user.Platform = platform
	if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
		xlog.Logger().Errorln("update user data failed: ", err.Error())
		return nil, nil, false, errors.New("更新用户数据失败")
	}

	person, err = GetDBMgr().db_R.GetPerson(user.Id)
	if err != nil {
		xlog.Logger().Errorln("get user from redis failed: ", err.Error())
		return nil, nil, false, err
	}
	person.Platform = platform

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, nil, false, err
	}

	// 更新redis
	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token, "Platform", platform)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, nil, false, err
	}

	//GetPersonMgr().AddPerson(person)
	return &user, person, isNew, nil
}

// 手机登录v2
func (s *Server) loginByMobilev2(mobile string, password string, platform int) (*models.User, *static.Person, *xerrors.XError) {
	var err error
	// 判断用户是否存在
	var person *static.Person
	var user models.User
	if err = GetDBMgr().db_M.Model(user).Where("tel = ? and password = ? and account_type = 0", mobile, util.MD5(password)).First(&user).Error; err != nil {
		return nil, nil, xerrors.NewXError("账号或密码错误")
	}

	// 更新mysql
	user.Platform = platform
	if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
		xlog.Logger().Errorln("update user data failed: ", err.Error())
		return nil, nil, xerrors.NewXError("更新用户数据失败")
	}

	// 更新redis
	person, err = GetDBMgr().db_R.GetPerson(user.Id)
	if err != nil {
		xlog.Logger().Errorln("get user from redis failed: ", err.Error())
		return nil, nil, xerrors.DBExecError
	}
	person.Platform = platform

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, nil, xerrors.DBExecError
	}

	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token, "Platform", platform)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, nil, xerrors.DBExecError
	}

	return &user, person, nil
}

// 手机注册
func (s *Server) registerByMobile(mobile string, platform int, ip string, machineCode string, password string, nickname string) (*static.Person, *xerrors.XError) {
	var err error
	// 判断用户是否存在
	var person *static.Person
	var user models.User

	// 注册账号
	var uid int64
	uid, err = s.AllocUserID()
	if err != nil {
		err = errors.New("生成的账号已经用完")
		xlog.Logger().Errorln(err)
		return nil, xerrors.DBExecError
	}
	user.Id = uid
	user.Nickname = nickname
	user.Imgurl = static.GenRobotUrl()
	user.Password = util.MD5(password)
	user.Tel = mobile
	user.Sex = consts.SexFemale // 默认女性
	user.Platform = platform
	user.Ip = ip
	user.MachineCode = machineCode
	user.UserType = consts.UserTypeMobile
	person, err = s.creatUserData(&user)
	if err != nil {
		xlog.Logger().Errorln("create user data failed: ", err.Error())
		return nil, xerrors.DBExecError
	}
	xlog.Logger().Errorln("login create mobile user:", person.Uid, mobile)

	person, err = GetDBMgr().db_R.GetPerson(user.Id)
	if err != nil {
		xlog.Logger().Errorln("get user from redis failed: ", err.Error())
		return nil, xerrors.DBExecError
	}

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, xerrors.DBExecError
	}
	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token, "Platform", platform)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, xerrors.DBExecError
	}

	//GetPersonMgr().AddPerson(person)
	return person, nil
}

// 微信登录
func (s *Server) loginByWechat(info *service2.WeixinUserInfo, mobile string, platform int, ip string, machineCode string) (*models.User, *static.Person, bool, error) {
	// 避免同一客户端重复发送多个请求过来，导致注册的微信账号重复
	GetServer().WechatLock.CustomLock()
	defer GetServer().WechatLock.CustomUnLock()

	var err error
	var person *static.Person
	var user models.User
	isNew := false
	xlog.Logger().Infoln("wxlogin mobile is: ", mobile)
	// 判断用户是否存在
	if user, err = s.AccountCheck(info, mobile, platform, ip, machineCode); err != nil {
		if err != gorm.ErrRecordNotFound {
			// 异常报错
			xlog.Logger().Errorln("query user data failed: ", err.Error())
			return nil, nil, false, err
		}

		// 首次快速登录注册 使用游客昵称
		if len(info.Nickname) == 0 {
			info.Nickname = getGuestRandomNickname()
			info.Headimgurl = ""
			info.Sex = consts.SexMale
		}

		// 如果不存在则创建用户
		isNew = true
		var uid int64
		uid, err = s.AllocUserID()
		if err != nil {
			err = errors.New("生成的账号已经用完")
			xlog.Logger().Errorln(err)
			return nil, nil, false, err
		}
		user.Id = uid
		user.Nickname = info.Nickname
		user.Openid = info.OpenId
		user.Unionid = info.UnionId
		user.Imgurl = info.Headimgurl
		user.Sex = info.Sex
		user.Tel = mobile
		user.Platform = platform
		user.Ip = ip
		user.MachineCode = machineCode
		// 默认为男性
		if user.Sex == consts.SexUnknown {
			user.Sex = consts.SexFemale // 默认女性
		}
		user.UserType = consts.UserTypeWechat
		person, err = s.creatUserData(&user)
		if err != nil {
			xlog.Logger().Errorln("create user data failed: ", err.Error())
			return nil, nil, false, errors.New("创建用户失败")
		}
		xlog.Logger().Warningln("login creatUser:", person.Uid, info.OpenId, "--nickname:", info.Nickname)
	} else {
		xlog.Logger().Warningln(user)
		if mobile != "" {
			// 判断手机信息是否一致, 不一致则认为是新手机新建账号, 提示微信已绑定
			if user.Tel != "" && mobile != user.Tel {
				return nil, nil, false, errors.New("微信账户已经绑定, 请更换微信账号后重试")
			}

			// 否则将手机绑定至该微信账号
			user.Tel = mobile
		}

		// 更新用户信息 快速登录 没有昵称头像信息
		if len(info.Nickname) > 0 && len(info.Headimgurl) > 0 {
			user.Nickname = info.Nickname
			user.Imgurl = info.Headimgurl
		}
		user.Openid = info.OpenId
		user.Unionid = info.UnionId
		user.Platform = platform
		//user.Sex = eve.Sex  // 关闭 这里是重新获取微信性别
		// syslog.Logger().Infoln("判断 machinecode :", user.MachineCode, ">>>", machineCode)
		if !static.HF_IsValidMachineCode(user.MachineCode) && static.HF_IsValidMachineCode(machineCode) {
			xlog.Logger().Infoln("MachineCode Null")
			user.MachineCode = machineCode
		}

		if user.Sex == consts.SexUnknown {
			user.Sex = consts.SexMale
		}

		// 更新redis
		person, err = GetDBMgr().db_R.GetPerson(user.Id)
		if err != nil {
			xlog.Logger().Errorln("get user from redis failed: ", err.Error())
			return nil, nil, false, errors.New("获取用户数据失败")
		}

		if user.Imgurl == "" {
			user.Imgurl = static.GenRobotUrl()
		}

		if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
			xlog.Logger().Errorln("update user data failed: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}

		person.Nickname = user.Nickname
		person.Imgurl = user.Imgurl
		person.Openid = user.Openid
		person.Sex = user.Sex
		person.Tel = user.Tel
		person.Platform = user.Platform
		//syslog.Logger().Infoln("wxlogin update SQl mobile is: ", person.Tel)
		err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Nickname", person.Nickname, "Imgurl", person.Imgurl, "Openid", person.Openid, "Sex", person.Sex, "Tel", person.Tel, "Platform", platform)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}
	}

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, nil, false, errors.New("更新用户数据失败")
	}
	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, nil, false, errors.New("更新用户数据失败")
	}

	return &user, person, isNew, nil
}

// 苹果登录
func (s *Server) loginByApple(appleUser *service2.AppleUserInfo, msgData *static.Msg_LoginApple, cliIP string) (*models.User, *static.Person, bool, error) {
	// 用户信息
	var person *static.Person
	var user models.User
	var err error
	isNew := false

	// 判断用户是否存在
	if err = GetDBMgr().db_M.Where("union_id = ? and account_type = 0", appleUser.UnionId).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			// 异常报错
			xlog.Logger().Errorln("query user data failed: ", err.Error())
			return nil, nil, false, err
		}

		// 创建新用户
		isNew = true
		// 注册账号
		var uid int64
		uid, err = s.AllocUserID()
		if err != nil {
			err = errors.New("生成的账号已经用完")
			xlog.Logger().Errorln(err)
			return nil, nil, false, errors.New("生成的账号已经用完")
		}
		user.Id = uid
		user.Nickname = getAppleRandomNickname() // 不使用苹果用户的真实姓名,由服务器生成apple+随机码
		user.Unionid = appleUser.UnionId
		user.Openid = appleUser.UnionId // 苹果账号没有openid
		user.UserType = consts.UserTypeApple
		user.Platform = msgData.Platform
		user.Sex = consts.SexFemale
		user.MachineCode = msgData.MachineCode
		user.Ip = cliIP
		user.Imgurl = ""
		user.Tel = ""

		// 创建用户数据
		person, err = s.creatUserData(&user)
		if err != nil {
			xlog.Logger().Errorln("create user data failed: ", err.Error())
			return nil, nil, false, errors.New("创建用户失败")
		}

		xlog.Logger().Warningln("login creatUser:", person.Uid, msgData.UserUnionId, "--nickname:", msgData.Nickname)
	} else {
		// 更新用户信息
		if len(msgData.Nickname) > 0 {
			user.Nickname = msgData.Nickname
		}
		if !static.HF_IsValidMachineCode(user.MachineCode) && static.HF_IsValidMachineCode(msgData.MachineCode) {
			xlog.Logger().Infoln("MachineCode Null")
			user.MachineCode = msgData.MachineCode
		}

		// 更新平台
		user.Platform = msgData.Platform

		// 更新mysql
		if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
			xlog.Logger().Errorln("update user data failed: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}

		// 更新redis
		person, err = GetDBMgr().db_R.GetPerson(user.Id)
		if err != nil {
			xlog.Logger().Errorln("get user from redis failed: ", err.Error())
			return nil, nil, false, errors.New("获取用户数据失败")
		}

		person.Nickname = user.Nickname
		person.Platform = user.Platform
		err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Nickname", person.Nickname, "Platform", msgData.Platform)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}
	}

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, nil, false, errors.New("创建用户token失败")
	}

	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, nil, false, errors.New("更新用户数据失败")
	}

	return &user, person, isNew, nil
}

// 华为登录
func (s *Server) loginByHW(hwUser *service2.HuaWeiUserInfo, msgData *static.Msg_LoginHW, cliIP string) (*models.User, *static.Person, bool, error) {
	// 用户信息
	var person *static.Person
	var user models.User
	var err error
	isNew := false

	// 判断用户是否存在
	if err = GetDBMgr().db_M.Where("union_id = ? and account_type = 0", hwUser.UnionId).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			// 异常报错
			xlog.Logger().Errorln("query user data failed: ", err.Error())
			return nil, nil, false, err
		}

		// 创建新用户
		isNew = true
		// 注册账号
		var uid int64
		uid, err = s.AllocUserID()
		if err != nil {
			err = errors.New("生成的账号已经用完")
			xlog.Logger().Errorln(err)
			return nil, nil, false, errors.New("生成的账号已经用完")
		}
		user.Id = uid
		user.Nickname = hwUser.Nickname
		user.Unionid = hwUser.UnionId
		user.Openid = hwUser.OpenId // 华为账号的openId非固定长度,最大允许长度256,需要调整一下user表open_id的长度
		user.UserType = consts.UserTypeHW
		user.Platform = msgData.Platform
		user.Sex = consts.SexFemale
		user.MachineCode = msgData.MachineCode
		user.Imgurl = hwUser.AvatarUrl
		user.Ip = cliIP
		user.Tel = ""

		// 创建用户数据
		person, err = s.creatUserData(&user)
		if err != nil {
			xlog.Logger().Errorln("create user data failed: ", err.Error())
			return nil, nil, false, errors.New("创建用户失败")
		}

		xlog.Logger().Warningln("login creatUser:", person.Uid, msgData.UnionId, "--nickname:", msgData.Nickname)
	} else {
		// 更新用户信息
		user.Nickname = msgData.Nickname
		user.Imgurl = msgData.AvatarUrl

		if !static.HF_IsValidMachineCode(user.MachineCode) && static.HF_IsValidMachineCode(msgData.MachineCode) {
			xlog.Logger().Infoln("MachineCode Null")
			user.MachineCode = msgData.MachineCode
		}

		// 更新平台
		user.Platform = msgData.Platform

		// 更新mysql
		if err = GetDBMgr().db_M.Save(&user).Error; err != nil {
			xlog.Logger().Errorln("update user data failed: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}

		// 更新redis
		person, err = GetDBMgr().db_R.GetPerson(user.Id)
		if err != nil {
			xlog.Logger().Errorln("get user from redis failed: ", err.Error())
			return nil, nil, false, errors.New("获取用户数据失败")
		}

		person.Nickname = user.Nickname
		person.Imgurl = user.Imgurl
		person.Platform = user.Platform
		err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Nickname", person.Nickname, "Imgurl", person.Imgurl, "Platform", msgData.Platform)
		if err != nil {
			xlog.Logger().Errorln("set user data to redis error: ", err.Error())
			return nil, nil, false, errors.New("更新用户数据失败")
		}
	}

	// 生成token
	person.Token, err = GetServer().createUserToken(person)
	if err != nil {
		xlog.Logger().Errorln("create user token error: ", err.Error())
		return nil, nil, false, errors.New("创建用户token失败")
	}

	err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Token", person.Token)
	if err != nil {
		xlog.Logger().Errorln("set user data to redis error: ", err.Error())
		return nil, nil, false, errors.New("更新用户数据失败")
	}

	return &user, person, isNew, nil
}

func (s *Server) UpdateUserImgUrlFromQiNiu(uid int64, imgUrl string) {
	if imgUrl == "" {
		return
	}
	if s.ConQiNiu.Able {
		if s.ConQiNiu.UseQueue {
			go s.Queue.Push(chanqueue.NewAsyncMsg(proto.MsgHeadQiNiuUserImg, &proto.QiNiuUserImgMsg{
				Uid:    uid,
				ImgUrl: imgUrl,
			}))
		} else {
			xlog.AnyErrors(QueueFetchUserImgToQiNiu(&proto.QiNiuUserImgMsg{
				Uid:    uid,
				ImgUrl: imgUrl,
			}), "更新玩家七牛头像失败。")
		}
	}
}

// 生成token
func (s *Server) createUserToken(person *static.Person) (string, error) {
	timeNow := time.Now()
	tokenStr := util.MD5(fmt.Sprintf("%d_%s", person.Uid, timeNow.Format("20060102150405")))

	if err := GetDBMgr().GetDBmControl().Model(models.User{}).Where("id = ?", person.Uid).Update("token", tokenStr).Error; err != nil {
		xlog.Logger().Errorln("update account_user failed: ", err.Error())
		return "", err
	}

	return tokenStr, nil
}

func (s *Server) creatUserData(user *models.User) (*static.Person, error) {
	if user.MachineCode == "" || !static.HF_IsValidMachineCode(user.MachineCode) {
		if user.Platform != consts.PlatformIos {
			return new(static.Person), fmt.Errorf("无效的设备码")
		}
	} else {
		if n := GetDBMgr().GetDBrControl().RedisV2.SCard(fmt.Sprintf("machine_code:%s", user.MachineCode)).Val(); n >= 2 {
			xlog.Logger().Errorf("用户机器码[%s]达到限制[%d]", user.MachineCode, n)
			return new(static.Person), fmt.Errorf("同一设备注册账号超出限制")
		}
	}

	if user.Imgurl == "" {
		user.Imgurl = static.GenRobotUrl()
	}
	xlog.Logger().Errorln("当前newcard：", s.ConServers.NewCard)
	tx := GetDBMgr().db_M.Begin()
	var err error
	//user.Id = GetPersonMgr().GetNextUid() // 分配用户id
	if user.Origin == consts.MigratorOriginHBMJ {
		// 迁移湖北麻将用户房卡按1:10转换
	} else {
		user.Card = s.ConServers.NewCard // 初始化房卡
	}
	user.Gold = s.ConServers.NewGold // 初始化金币

	if err = tx.Create(&user).Error; err != nil {
		xlog.Logger().Errorln("create user failed: ", err.Error())
		tx.Rollback()
		return nil, err
	}

	// 记录用户财富消耗流水
	if user.Card > 0 {
		record := new(models.UserWealthCost)
		record.Uid = user.Id
		record.WealthType = consts.WealthTypeCard
		record.CostType = models.CostTypeRegister
		record.Cost = int64(user.Card)
		record.BeforeNum = 0
		record.AfterNum = int64(user.Card)
		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			xlog.Logger().Errorln(err.Error())
			return nil, err
		}
	}

	// 记录用户财富消耗流水
	if user.Gold > 0 {
		record := new(models.UserWealthCost)
		record.Uid = user.Id
		record.WealthType = consts.WealthTypeGold
		record.CostType = models.CostTypeRegister
		record.Cost = int64(user.Gold)
		record.BeforeNum = 0
		record.AfterNum = int64(user.Gold)
		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			xlog.Logger().Errorln(err.Error())
			return nil, err
		}
	}

	// db对象 -> redis对象
	person := user.ConvertModel()
	err = GetDBMgr().db_R.AddPerson(person)
	if err != nil {
		xlog.Logger().Errorln("insert user to redis failed: ", err.Error())
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	GetDBMgr().GetDBrControl().RedisV2.SAdd(fmt.Sprintf("machine_code:%s", user.MachineCode), person.Uid)
	xlog.Logger().Errorln("InsertTable user uid:", person.Uid)

	return person, nil
}

func (s *Server) LoadDBData() {
	// 读取服务器配置
	if err := GetDBMgr().ReadAllConfig(); err != nil {
		xlog.Logger().Errorln("ReadAllConfig err:", err)
		return
	}
}

// 获取随机昵称
func getGuestRandomNickname() string {
	return static.GenGuestRandomNickname()
}

// 获取苹果随机昵称
func getAppleRandomNickname() string {
	return static.GenAppleRandomNickname()
}

// 获取渠道配置
func (s *Server) GetAppConfig(appId string) *models.ConfigApp {
	for _, item := range s.ConApp {
		if item.AppId == appId {
			return item
		}
	}
	return nil
}

// 账号检查
func (s *Server) AccountCheck(info *service2.WeixinUserInfo, mobile string, platform int, ip string, machineCode string) (models.User, error) {
	// 先从go的数据库查union_id
	var dberr error
	var user models.User
	// 先找迁移过来的chess湖北麻将用户
	dberr = GetDBMgr().db_M.Where("union_id = ? and account_type = 0 and origin = 6", info.UnionId).First(&user).Error
	// 玩家存在
	if dberr == nil && user.Id > 0 {
		xlog.Logger().Debugln("按union_id找到chess湖北麻将玩家了", user.Id, user.MachineCode)
		return user, nil
	}

	// 再找正常的js用户
	dberr = GetDBMgr().db_M.Where("union_id = ? and account_type = 0", info.UnionId).First(&user).Error
	// 玩家存在
	if dberr == nil && user.Id > 0 {
		xlog.Logger().Debugln("按union_id找到chess湖北麻将玩家了", user.Id, user.MachineCode)
		return user, nil
	}

	// go数据库未找到用户 再去湖北麻将数据库查找
	if dberr == gorm.ErrRecordNotFound && s.ConSqlServer.Able == 1 {
		cppUser := new(origin.User)
		dbSQLErr := GetDBMgr().GetDBsControl().Exec("use QPGameUserDB").Table("QPGameUserDB.dbo.AccountsInfo").Where("UnionID = ?", info.UnionId).Find(cppUser).Error
		if dbSQLErr == nil && cppUser.UserId > 0 {
			xlog.Logger().Debugln("通过UnionID:%d找到湖北麻将用户:%d", info.OpenId, cppUser.UserId)

			// 用户房卡
			type CppUserWealth struct {
				Card int64 `gorm:"column:Score"` // 湖北麻将的房卡本身已x10
			}
			var userWealth CppUserWealth
			dbSQLErr = GetDBMgr().GetDBsControl().Exec("use QPTreasureDB").Table("QPTreasureDB.dbo.GameScoreInfo").Where("UserID = ?", cppUser.UserId).Find(&userWealth).Error
			if dbSQLErr != nil {
				xlog.Logger().Errorln("select cpp user score failed ", dbSQLErr.Error())
				return user, dberr
			}

			// ID不变
			user.Id = cppUser.UserId
			user.Openid = info.OpenId
			user.Unionid = info.UnionId

			// 迁移用户信息、及财富信息(1:10)
			user.Card = int(userWealth.Card)
			user.Nickname = cppUser.Nickname
			user.Imgurl = cppUser.FaceUrl
			user.Sex = cppUser.Gender
			user.Tel = mobile
			user.Platform = platform
			user.Ip = ip
			user.MachineCode = machineCode
			user.Origin = consts.MigratorOriginHBMJ
			if cppUser.Gender == 1 {
				user.Sex = consts.SexMale
			} else {
				user.Sex = consts.SexFemale
			}

			user.UserType = consts.UserTypeWechat
			_, err := s.creatUserData(&user)
			if err != nil {
				xlog.Logger().Errorln("create user data failed: ", err.Error())
			}

			xlog.Logger().Warningln("迁移湖北麻将用户创建 成功 creatUser:", user.Id)

			return user, nil
		}
	}

	return user, dberr

	// db error
	// if dberr != gorm.ErrRecordNotFound {
	// 	syslog.Logger().Debugln("db error:", dberr)
	// 	return user, dberr
	// }
	//
	// // 没找到 就根据machine_code找
	// if !public.HF_IsValidMachineCode(machineCode) {
	// 	return user, dberr
	// }
	// dberr = GetDBMgr().db_M.Where("machine_code = ?", machineCode).First(&user).Error
	// // 玩家找到了
	// if dberr == nil && user.Id > 0 {
	// 	syslog.Logger().Infoln("按machine_code找到玩家了", user.Id, user.MachineCode)
	// 	// 这里注意 如果根据machinecode找到的玩家存在unionid 则认为该机器码已被使用，创建一个新用户
	// 	if len(user.Unionid) == 28 {
	// 		syslog.Logger().Errorf("机器码:%s.已被微信用户%d.使用，此时认为是新用户", machineCode, user.Id)
	// 		// 重置玩家
	// 		user = model.User{}
	// 		return user, gorm.ErrRecordNotFound
	// 	}
	// 	return user, nil
	// }
	//
	// if dberr != gorm.ErrRecordNotFound {
	// 	syslog.Logger().Errorln("db error:", dberr)
	// 	return user, dberr
	// }
	//
	// // go数据库按union_id和machine_code都未找到该玩家 去c++数据库找
	//
	// // 如果未初始化sqlserver的连接 就返回未找到玩家
	// if GetDBMgr().db_S == nil {
	// 	if self.ConSqlServer.Able > 0 {
	// 		// 如果sqlserver开启了登录查询
	// 		// 因为此时还未找到玩家信息，所以不知道用户来源于c++的哪个游戏
	// 		syslog.Logger().Errorln("根据用户openid和机器码都未找到玩家信息，C++ SqlServer数据库未初始化，此时认为异常")
	// 		return user, fmt.Errorf("登录失败:sqlserver error")
	// 	} else {
	// 		// 如果未开启查询 此时认为新用户
	// 		return user, gorm.ErrRecordNotFound
	// 	}
	// }
	//
	// // 开始查 sqlserver
	// cppUser := new(origin.User)
	// dberr = GetDBMgr().GetDBsControl().Exec(
	// 	fmt.Sprintf("use %s", cppUser.Database())).Table(
	// 	cppUser.TableName()).Where(
	// 	"machineCode = ? and IsGuest = ?", machineCode, 2).Find(cppUser).Error
	//
	// // 如果找到了
	// if dberr == nil && cppUser.UserId > 0 {
	// 	syslog.Logger().Infoln("从 sqlserver 找到玩家：", cppUser.UserId, cppUser.MachineSerial)
	// 	// 根据这个人的id去找数据
	// 	if err := GetDBMgr().GetDBmControl().Where("id = ?", cppUser.UserId).First(&user).Error; err != nil {
	// 		syslog.Logger().Errorln("sqlserver存在玩家:", cppUser.UserId, cppUser.IsGuest, "但是mysql未找到:", err)
	// 		return user, err
	// 	}
	// 	syslog.Logger().Warningln(fmt.Sprintf("%+v", user))
	// 	return user, nil
	// }
	// return user, dberr
}

func (s *Server) RunQueue() {
	s.Queue = chanqueue.NewQueue(&QueueHandler{})
	s.Queue.Run(s.Wait)
}

func (s *Server) AllocUserID() (int64, error) {
	//先找redis
	cli := GetDBMgr().GetDBrControl()
	ret, err := cli.RedisV2.SPop("UidListV2").Int64()
	if err != nil {
		logrus.Errorln("redis客户端获取数据失败")
		return 0, err
	}

	return ret, nil
}
func (s *Server) RegistUserID() {

	cli := GetDBMgr().GetDBrControl()
	n, err := cli.RedisV2.SCard("UidListV2").Result()
	if err != nil {
		xlog.Logger().Errorf("获取uid list失败: %v", err)
		return
	}
	if n == 0 {
		idStart := 1000001
		idEnd := 9999999
		reservedSuffixList := make([]int, len(static.RobotIdSuffix))
		for k, v := range static.RobotIdSuffix {
			reservedSuffixList[k] = int(v)
		}
		var playerIds []string
		for id := idStart; id <= idEnd; id++ {
			suffix := id % 1000000
			if slices.Contains(reservedSuffixList, suffix) {
				continue
			}
			playerId := strconv.FormatUint(uint64(id), 10)
			playerIds = append(playerIds, playerId)

			if id%1000 == 0 || id == 9999999 {
				cli.RedisV2.SAdd("UidListV2", playerIds)
				playerIds = []string{}
			}
		}
		xlog.Logger().Warningln("userid添加成功")
	}
	// person, err = GetDBMgr().db_R.GetPerson(1)
	// if err != nil {
	// 	xlog.Logger().Errorln("get user from redis failed: ", err.Error())
	// 	return nil, nil, false, err
	// }

	// // 更新redis
	// person.Platform = user.Platform
	// err = GetDBMgr().GetDBrControl().UpdatePersonAttrs(person.Uid, "Platform", platform)
	// if err != nil {
	// 	xlog.Logger().Errorln("set user data to redis error: ", err.Error())
	// 	return nil, nil, false, err
	// }
}
