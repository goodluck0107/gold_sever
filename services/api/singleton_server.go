package api

import (
	"context"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Host                  string `json:"host"` //! GM服务器ip
	RedisDB               int    `json:"redisdb"`
	RedisAuth             string `json:"redisauth"`       //!redis密码
	Redis                 string `json:"redis"`           //! redis
	DB                    string `json:"db"`              //! 数据库
	Encode                int    `json:"encode"`          //! 消息加密方式 0为不加密(使用加密后客户端发的消息也必须加密, 否则忽略)
	EncodeClientKey       string `json:"encodeclientkey"` //! 与客户端通信的aes加密秘钥
	EncodePhpKey          string `json:"encodephpkey"`    //! 与php通信的aes加密秘钥
	service2.CommonConfig        // 第三方服务配置
}

var serverSingleton *Server = nil

//! 得到服务器指针
func GetServer() *Server {
	if serverSingleton == nil {
		serverSingleton = new(Server)
		serverSingleton.ShutDown = false
		serverSingleton.Con = new(Config)
		serverSingleton.ConGame = make([]*models.GameConfig, 0)
		serverSingleton.Wait = new(sync.WaitGroup)
		serverSingleton.BlackIds = make(map[int64]int8)
	}

	return serverSingleton
}

type Server struct {
	Con             *Config              //! 配置
	ConGame         []*models.GameConfig //! 游戏配置
	Wait            *sync.WaitGroup      //! 同步阻塞
	ShutDown        bool                 //! 是否正在执行关闭
	RpcLogin        *static.ClientPool
	HallServer      string
	GateServer      string
	BlackIds        map[int64]int8
	DataSynchronism bool //! 数据同步
}

func (s *Server) Host() (string, string) {
	ip := strings.Split(s.Con.Host, ":")
	return ip[0], ip[1]
}

func (s *Server) Start(ctx context.Context) error {
	// 第三方服务初始化
	service2.InitConfig(s.Con.CommonConfig)
	// 连接数据库
	if GetDBMgr() == nil {
		return fmt.Errorf("db init failed")
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.ShutDown = true
	s.Wait.Wait()
	if err := GetDBMgr().Close(); err != nil {
		return err
	}
	xlog.Logger().Warningln("api wuhan shutdown")
	return nil
}

func (s *Server) RegisterRouter(ctx context.Context) {
	router.NewRouter("/api", NewHttpApi())
}

func (s *Server) RegisterRPC(ctx context.Context) {
	return
}

func (s *Server) LoadConfig(ctx context.Context) error {
	return GetDBMgr().ReadAllConfig()
}

func (s *Server) LoadServerConfig(ctx context.Context) error {
	if err := static.LoadServeYaml(s, true); err != nil {
		xlog.Logger().Debug("InitConfig_yaml:", err)
		return err
	} else {
		xlog.Logger().Debug("Update wuhan yaml etc succeed : ", fmt.Sprintf("%+v", s.Con))
	}
	return nil
}

func (s *Server) SetLoggerLevel() {
	xlog.SetFileLevel("info")
}

func (s *Server) Run(ctx context.Context) {
	xlog.Logger().Errorln("run,time:", time.Now().Hour())
	ticker := time.NewTicker(time.Second * 3600)
	defer ticker.Stop()
	for {
		<-ticker.C
		if s.ShutDown {
			break
		}
	}
}

func (s *Server) Handle(string) error {
	return nil
}

func (s *Server) IsMulti() bool {
	return false
}

func (s *Server) ErrorDeep() int {
	return 1
}

func (s *Server) Name() string {
	return static.ServerNameApi
}

func (s *Server) Etc() interface{} {
	return s.Con
}

// 获取显示免费游戏kind
func (s *Server) GetLimitFreeGameKindIds() map[int]bool {
	freeKindMap := make(map[int]bool)
	for _, gameItem := range s.ConGame {
		var limitFreeFind = false
		for _, roundItem := range gameItem.RoundMap {
			for _, cost := range roundItem {
				if cost == 0 {
					limitFreeFind = true
					goto FOUND
				}
			}
		}
	FOUND:
		freeKindMap[gameItem.KindId] = limitFreeFind
	}
	return freeKindMap
}
