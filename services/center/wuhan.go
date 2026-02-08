package center

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/router"
	"github.com/open-source/game/chess.git/pkg/rpc"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/authentication"
	"github.com/open-source/game/chess.git/pkg/static/wealthtalk"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net/http"
	"strings"
	"sync"
	"time"

	// "github.com/open-source/game/chess.git/pkg/public/network.go"

	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"

	"golang.org/x/net/websocket"
)

type RpcMaps struct {
	data map[int]*client.XClient
	lock *sync.RWMutex
}

func (self *RpcMaps) Read(f func(*map[int]*client.XClient)) {
	self.lock.RLock()
	defer self.lock.RUnlock()
	f(&self.data)
}

func (self *RpcMaps) Write(f func(*map[int]*client.XClient)) {
	self.lock.Lock()
	defer self.lock.Unlock()
	f(&self.data)
}

type Config struct {
	Host                  string `json:"host"`   //! 服务器ip
	InHost                string `json:"inhost"` //! 服务器inip
	UserServerAddr        string `json:"user" `  //!登錄服
	Redis                 string `json:"redis"`  //! redis
	RedisDB               int    `json:"redisdb"`
	RedisAuth             string `json:"redisauth"`       //!redis密码
	DB                    string `json:"db"`              //！数据库
	RcRedis               string `json:"rcrdis"`          //! 数据库
	SafeHost              string `json:"safehost"`        //! 服务器ip_高仿地址
	UseSafeIp             int8   `json:"usesafeip"`       //! 是否用高仿地址：0不用，1用
	Encode                int    `json:"encode"`          // 消息加密方式 0为不加密(使用加密后客户端发的消息也必须加密, 否则忽略)
	EncodeClientKey       string `json:"encodeclientkey"` // 与客户端通信的aes加密秘钥
	EncodePhpKey          string `json:"encodephpkey"`    // 与php通信的aes加密秘钥
	PubRedis              string `json:"pubredis"`        //! redis
	PubRedisDB            int    `json:"pubredisdb"`
	PubRedisAuth          string `json:"pubredisauth"` //!redis密码
	service2.CommonConfig                              // 第三方服务配置
}

var serverSingleton *Server = nil

// ! 得到服务器指针
func GetServer() *Server {
	if serverSingleton == nil {
		serverSingleton = new(Server)
		serverSingleton.ShutDown = false
		serverSingleton.Con = new(Config)
		serverSingleton.ConSite = make([]*models.ConfigSite, 0)
		serverSingleton.ConMatch = make([]*models.ConfigMatch, 0)
		serverSingleton.ConMatchAward = make([]*models.GameMatchCoupon, 0)
		serverSingleton.ConGame = make([]*models.GameConfig, 0)
		serverSingleton.ConHouse = new(models.ConfigHouse)
		serverSingleton.ConServers = new(models.ConfigServer)
		serverSingleton.ConGovAuth = new(models.ConfigGovAuth)
		serverSingleton.ConSpinBase = new(models.ConfigSpinBase)
		serverSingleton.ConBattleLevel = make([]*models.ConfigBattleLevel, 0)
		serverSingleton.ConSpinAward = make([]*models.ConfigSpinAward, 0)
		serverSingleton.ConCheckIn = make([]*models.ConfigCheckin, 0)
		// serverSingleton.ConWebPay = new(model.ConfigWebPay)
		serverSingleton.Wait = new(sync.WaitGroup)
		serverSingleton.BlackIds = make(map[int64]int8)
		serverSingleton.GameServers = make(map[int]*static.Msg_GameServer)
		serverSingleton.RpcGames = &RpcMaps{
			data: make(map[int]*client.XClient, 10),
			lock: new(sync.RWMutex),
		}
		serverSingleton.GameLock = new(lock2.RWMutex)
		serverSingleton.MatchGameLock = new(sync.RWMutex)
		serverSingleton.MatchOverLock = new(sync.Mutex)
		serverSingleton.RpcGamesLock = new(sync.RWMutex)
	}

	return serverSingleton
}

type Server struct {
	Con            *Config              //! 配置
	ConApp         []*models.ConfigApp  //! 渠道配置
	ConServers     *models.ConfigServer //！服务器配置
	ConSite        []*models.ConfigSite
	ConGame        []*models.GameConfig      //! 游戏配置
	ConMatch       []*models.ConfigMatch     //! 排位赛配置
	ConMatchAward  []*models.GameMatchCoupon //! 排位赛配置
	ConHouse       *models.ConfigHouse       //! 包厢配置
	ConGovAuth     *models.ConfigGovAuth     // 政府实名认证配置
	ConBattleLevel []*models.ConfigBattleLevel
	ConSpinBase    *models.ConfigSpinBase
	ConSpinAward   []*models.ConfigSpinAward
	ConCheckIn     []*models.ConfigCheckin
	// ConWebPay     *model.ConfigWebPay      //! 七牛云配置
	Wait     *sync.WaitGroup //! 同步阻塞
	ShutDown bool            //! 是否正在执行关闭
	// RpcLogin      *network.RPCClient
	BlackIds      map[int64]int8 //0 不在黑名单，1在黑名单
	GameServers   map[int]*static.Msg_GameServer
	RpcGamesLock  *sync.RWMutex
	GameLock      *lock2.RWMutex
	MatchGameLock *sync.RWMutex //用于控制ConMatch和ConMatchAward的初始化和读取
	MatchOverLock *sync.Mutex   //用于控制排位赛结束时结算中，读结算和写结算互斥
	RpcGames      *RpcMaps
}

func (s *Server) RegisterRouter(ctx context.Context) {
	router.NewRouter("/service", NewHttpService())
}

func (s *Server) RegisterRPC(ctx context.Context) {
	InitByIP(s.Con.UserServerAddr)
	rpc.NewRpcxServer(s.Con.Host, "ServerMethod", &ServerMethod{}, &ServerPlugin{})
}

func (s *Server) Host() (string, string) {
	ip := strings.Split(s.Con.Host, ":")
	return ip[0], ip[1]
}

func (s *Server) LoadConfig(ctx context.Context) error {
	return GetDBMgr().ReadAllConfig()
}

func (s *Server) SetLoggerLevel() {
	xlog.SetFileLevel("info")
}

func (s *Server) Name() string {
	return static.ServerNameCenter
}

func (s *Server) Etc() interface{} {
	return s.Con
}

// ! 初始化
func (s *Server) Start(ctx context.Context) error {
	err := s.LoadServerConfig(ctx)
	if err != nil {
		return err
	}
	// 第三方服务初始化
	service2.InitConfig(s.Con.CommonConfig)
	// 包厢控制器实例
	if GetClubMgr() == nil {
		return fmt.Errorf("包厢实例化失败")
	}
	// 协议实例
	if GetProtocolMgr() == nil {
		return fmt.Errorf("协议实例初始化失败.")
	}
	// 数据实例
	if GetDBMgr() == nil {
		return fmt.Errorf("数据实例初始化失败.")
	}
	// 牌桌数据恢复
	GetTableMgr().Restore()
	// 数据库数据恢复
	err = GetDBMgr().DataInitialize()
	if err != nil {
		return fmt.Errorf("DBMgrSingleton.DataInitialize: %v", err)
	}

	// 初始化任务配置
	GetTasksMgr().Init()

	if protocolworkers == nil {
		protocolworkers = new(ProtocolWorkers)
		protocolworkers.Init()
	}

	// 政府实名上报系统
	authentication.AuthenticationThread()

	return nil
}

// ! 得到一个websocket处理句柄
func (s *Server) GetConnectHandler() websocket.Handler {
	connectHandler := func(ws *websocket.Conn) {
		if s.ShutDown { //! 关服了
			ws.Close()
			return
		}
		if static.GetBlackIpMgr().IsIp(static.HF_GetHttpIP(ws.Request())) { //! ip在黑名单里
			xlog.Logger().Errorln("黑名单ip登陆")
			ws.Close()
			return
		}
		session := GetSessionMgr().GetNewSession(ws)
		session.Run()
	}
	return websocket.Handler(connectHandler)
}

func (s *Server) LoadServerConfig(ctx context.Context) error {
	if err := static.LoadServeYaml(s, true); err != nil {
		return err
	} else {
		xlog.Logger().Debug("load wuhan yaml etc succeed : ", fmt.Sprintf("%+#v", s.Con))
	}
	return nil
}

// ! 关闭服务器
func (s *Server) Stop(ctx context.Context) error {
	//! 告诉服务器关闭
	s.ShutDown = true
	s.Wait.Wait()
	// 保存任务状态
	s.NoticeGameServer()

	s.RpcGames.Write(func(m *map[int]*client.XClient) {
		for _, v := range *m {
			err := (*v).Close()
			xlog.Logger().Warnf("close client:%v", err)
		}
	})

	GetClubMgr().lock.RLock()
	for _, h := range GetClubMgr().ClubMap {
		if !h.DBClub.MixActive {
			continue
		}
		h.FloorLock.RLock()
		for _, f := range h.Floors {
			if f.IsMix {
				f.SaveHft()
			}
		}
		h.FloorLock.RUnlock()
	}
	GetClubMgr().lock.RUnlock()
	err := GetDBMgr().Close()
	if err != nil {
		return err
	}
	// 统计一下大厅协议使用情况
	GetProtocolMgr().StataProtocolCnt()
	return nil
}

// ! 通知服务器
func (s *Server) NoticeGameServer() {
	//! 告诉游戏服务器关闭
	for id, _ := range s.GameServers {
		xlog.Logger().Infof("noticegameserver ,gameid:%d", id)
		var msg static.Msg_Null
		s.CallGame(id, 0, "NewServerMsg", "closehallserver", xerrors.SuccessCode, &msg)
	}
}

// ! 调用游戏服务器
func (s *Server) CallGame(gameId int, uid int64, method string, msghead string, errCode int16, v interface{}) ([]byte, error) {
	var cli *client.XClient
	GetServer().RpcGames.Read(func(m *map[int]*client.XClient) {
		cli = (*m)[gameId]
	})
	if cli == nil {
		return nil, errors.New("game cli error")
	}
	args := &static.Msg_MsgBase{}
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	args.Data = string(buf)
	args.Head = msghead
	args.Uid = uid
	buf, err = json.Marshal(args)
	if err != nil {
		return nil, err
	}
	rep, err := CallGame(&static.Rpc_Args{MsgData: buf}, cli, gameId)
	if err != nil {
		s.GameLock.RLock()
		gameInfo := s.GameServers[gameId]
		s.GameLock.RUnlock()

		if gameInfo != nil {
			GetServer().RpcGames.Write(func(m *map[int]*client.XClient) {
				(*m)[gameId] = InitByIP(gameInfo.InIp)
			})
		} else {
			xlog.Logger().Errorf("reconnect error: game wuhan nil:%d", gameId)
		}
	}
	return rep, nil
}

// ! 向所有游戏服广播rpc消息
func (s *Server) BroadcastGame(uid int64, method string, msghead string, errCode int16, v interface{}) {
	s.GameLock.CustomLock()
	defer s.GameLock.CustomUnLock()
	for _, gsvr := range s.GameServers {
		if gsvr == nil {
			continue
		}
		//syslog.Logger().Infoln("beg call wuhan ...", gsvr.Id)
		go func(gid int) {
			_, err := s.CallGame(gid, uid, method, msghead, errCode, v)
			//syslog.Logger().Infoln("end call wuhan ...", gsvr.Id)
			if err != nil {
				xlog.Logger().WithFields(logrus.Fields{
					"gamesvrid": gid,
					"errinfo":   err,
				}).Errorln("hall broadcost rpc to game err.")
			}
		}(gsvr.Id)
	}
}

func (s *Server) Run(ctx context.Context) {
	xlog.Logger().Info("run,time:", time.Now().Hour())
	go func() { xlog.Logger().Error(http.ListenAndServe(":6060", nil)) }()
	go GetMsg()
	/*
		大厅服务器启动15秒钟后开始检查公告
		并不是在检查之后才会有公告数据
	*/
	go checkMaintainNotice(time.Second * 15)

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if s.ShutDown {
				break
			}

			if time.Now().Minute() == 0 {
				//每小时
			}
			if time.Now().Hour() == 0 {
				// 每天凌晨统计一下协议使用次数
				go GetProtocolMgr().StataProtocolCnt()
			}
		case <-ctx.Done():
			return
		}
	}
}

// ! 获取游戏服务
func (s *Server) GetGame(gameId int) *static.Msg_GameServer {
	s.GameLock.CustomLock()
	defer s.GameLock.CustomUnLock()

	return s.GameServers[gameId]
}

// ! 获取游戏服务器rpcclient
func (s *Server) GetGameRpcClient(gameId int) *client.XClient {
	var cli *client.XClient
	GetServer().RpcGames.Read(func(m *map[int]*client.XClient) {
		cli = (*m)[gameId]
	})
	return cli
}

// ! 获取某个游戏支持的场次类型
func (s *Server) GetSiteTypeByKindId(gameType int, kindId int) []int {
	s.GameLock.CustomLock()
	defer s.GameLock.CustomUnLock()

	isInArray := func(v int, arr []int) bool {
		for _, item := range arr {
			if v == item {
				return true
			}
		}
		return false
	}

	// 合并去重
	result := make([]int, 0)
	for _, value := range s.GameServers {
		// 判断服务器是否维护
		if value.Status == consts.ServerStatusMaintain {
			continue
		}

		if v, ok := value.GameTypes[kindId]; ok {
			for _, item := range v {
				if item.GameType == gameType {
					if !isInArray(item.SiteType, result) {
						result = append(result, item.SiteType)
					}
				}
			}
		}
	}
	return result
}

// ! 获取某kindid下的所有gameserver
func (s *Server) GetGamesByKindId(kindId int) []*static.Msg_GameServer {
	s.GameLock.RLockWithLog()
	defer s.GameLock.RUnlock()

	result := make([]*static.Msg_GameServer, 0)
	for _, value := range s.GameServers {
		if _, ok := value.GameTypes[kindId]; ok {
			result = append(result, value)
		}
	}
	return result
}

type GameInfo struct {
	Id       int
	TableNum int
}

// ! 好友房分配gameserver(得到room最少的gameserver)
func (s *Server) GetGameByKindId(uid int64, gameType int, kindId int) (*static.Msg_GameServer, *static.Notice) {
	s.GameLock.RLockWithLog()
	defer s.GameLock.RUnlock()

	var notice *static.Notice

	ids := []*GameInfo{}
	for key, value := range s.GameServers {
		// 判断服务器是否维护
		if value.Status == consts.ServerStatusMaintain {
			continue
		}

		// 判断是否支持子游戏
		if v, ok := value.GameTypes[kindId]; !ok {
			continue
		} else {
			// 判断是否支持gameType
			support := false
			for _, item := range v {
				if item.GameType == gameType {
					support = true
					break
				}
			}
			if !support {
				continue
			}
			if _notice := CheckServerMaintainWithWhite(uid, static.NoticeMaintainServerType(value.Id)); _notice != nil {
				if notice == nil || notice.GameServerId > _notice.GameServerId {
					notice = _notice
				}
				continue
			}
		}

		// 获取游戏服牌桌数量
		gameTableNum := GetTableMgr().GetGameTableNum(value.Id)
		// if gameTableNum < num {
		// 	num = gameTableNum
		// 	id = key
		// }
		ids = append(ids, &GameInfo{Id: key, TableNum: gameTableNum})
	}
	if len(ids) == 0 {
		return nil, notice
	}
	minTab := &GameInfo{Id: 0, TableNum: 1000000000}
	for _, game := range ids {
		if game.TableNum < minTab.TableNum {
			minTab = game
		}
	}
	xlog.Logger().Warningln("分配的game进程：", minTab.Id)
	return s.GameServers[minTab.Id], notice
}

// ! 金币场分配gameserver(人数最少的服务器)
func (s *Server) GetGameBySiteType(uid int64, kindId int, siteType int) (*static.Msg_GameServer, *static.Notice) {
	s.GameLock.RLockWithLog()
	defer s.GameLock.RUnlock()

	var notice *static.Notice

	isInArray := func(v int, arr []*static.ServerGameType) bool {
		for _, item := range arr {
			if item.GameType == static.GAME_TYPE_GOLD && v == item.SiteType {
				return true
			}
		}
		return false
	}

	// 寻找人数最少的服务器
	var svr *static.Msg_GameServer
	var minNumber int = 999999
	for _, value := range s.GameServers {
		// 判断服务器是否维护
		if value.Status == consts.ServerStatusMaintain {
			continue
		}

		// 判断是否支持子游戏
		if v, ok := value.GameTypes[kindId]; !ok {
			continue
		} else {
			if isInArray(siteType, v) {
				if _notice := CheckServerMaintainWithWhite(uid, static.NoticeMaintainServerType(value.Id)); _notice != nil {
					if notice == nil || notice.GameServerId > _notice.GameServerId {
						notice = _notice
					}
					continue
				}
				num := GetPlayerMgr().GetOnlineNumberByKindId(kindId, siteType)
				if num < minNumber {
					num = minNumber
					svr = value
				}
			}
		}
	}

	if svr == nil {
		return nil, notice
	}

	xlog.Logger().Warningln("分配的game进程：", svr.Id)
	return svr, notice
}

// 获取游戏配置
func (s *Server) GetGameConfig(kindid int) *models.GameConfig {
	for _, item := range s.ConGame {
		if item.KindId == kindid {
			return item
		}
	}
	return nil
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

// 获取房间场次配置
func (s *Server) GetRoomConfig(kindId int, siteType int) *models.ConfigSite {
	for _, item := range s.ConSite {
		if item.KindId == kindId && item.Type == siteType {
			return item
		}
	}
	return nil
}

// 获取房间场次的排位赛配置
func (s *Server) GetMatchConfig(kindid int, siteType int) []*models.ConfigMatch {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	matchconfig := make([]*models.ConfigMatch, 0)
	for _, item := range s.ConMatch {
		if item.KindId == kindid {
			for _, typeitem := range item.Types {
				if typeitem == siteType {
					matchconfig = append(matchconfig, item)
				}
			}
		}
	}
	if len(matchconfig) > 0 {
		return matchconfig
	}
	return nil
}

// 获取房间场次的排位赛配置
func (s *Server) GetMatchConfigByKindid(kindid int) []*models.ConfigMatch {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	matchconfig := make([]*models.ConfigMatch, 0)
	for _, item := range s.ConMatch {
		if item.KindId == kindid {
			matchconfig = append(matchconfig, item)
		}
	}
	if len(matchconfig) > 0 {
		return matchconfig
	}
	return nil
}

// 获取指定排位赛的排位赛配置
func (s *Server) GetMatchConfigByStr(kindid int, siteType string, beginDate int64, beginTime int64) *models.ConfigMatch {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	for _, item := range s.ConMatch {
		if item.KindId == kindid && item.TypeStr == siteType && item.BeginDate.Unix() == beginDate && item.BeginTime.Unix() == beginTime {
			return item
		}
	}
	return nil
}

// 获取房间场次的排位赛配置
func (s *Server) GetMatchAwardConfig(kindId int, siteType string, beginDate int64, beginTime int64) *models.GameMatchCoupon {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	for _, item := range s.ConMatchAward {
		if item.KindId == kindId && item.SiteTypeStr == siteType && item.BeginDate.Unix() == beginDate && item.BeginTime.Unix() == beginTime {
			return item
		}
	}
	return nil
}

// 获取房间场次的排位赛配置
func (s *Server) GetMatchHonorAwardConfig(gameMatchTotal *models.GameMatchTotal, siteType int) *models.HonorAwards {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	for _, item := range s.ConMatchAward {
		if item.KindId == gameMatchTotal.KindId && item.SiteTypeStr == gameMatchTotal.SiteTypeStr && item.BeginDate == gameMatchTotal.BeginDate && item.BeginTime == gameMatchTotal.BeginTime {
			for i, typeitem := range item.SiteTypes {
				if typeitem == siteType {
					return &item.HonorAwards[i]
					break
				}
			}
		}
	}
	return nil
}

// 获取房间场次的排位赛配置,排位奖励信息
func (s *Server) GetMatchRankingAwardConfig(kindId int, siteType string, beginDate int64, beginTime int64) []models.RankingAwards {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	for _, item := range s.ConMatchAward {
		if item.KindId == kindId && item.SiteTypeStr == siteType && item.BeginDate.Unix() == beginDate && item.BeginTime.Unix() == beginTime {
			return item.RankingAwards
		}
	}
	return nil
}

// ! 初始化
func (s *Server) InitRpcGame(idGame int) {
	//! 连接rpc
	ipGame := s.GameServers[idGame].InIp
	GetServer().RpcGames.Write(func(m *map[int]*client.XClient) {
		(*m)[idGame] = InitByIP(ipGame)
	})
}

// ! 添加一个游戏服务器
func (s *Server) AddOneGameServer(idGame int, gameserver *static.Msg_GameServer) {
	s.GameLock.CustomLock()
	_, ok := s.GameServers[idGame]
	s.GameServers[idGame] = gameserver
	s.GameLock.CustomUnLock()

	s.RpcGames.Write(func(m *map[int]*client.XClient) {
		if !ok || (*m)[idGame] == nil {
			(*m)[idGame] = AddGameServer(gameserver.InIp)
			xlog.Logger().Infoln("game_", idGame, ":连接")
		}
	})
}

// ! 删除一个游戏服务器
func (s *Server) DelOneGameServer(idGame int) {
	s.RpcGames.Write(func(m *map[int]*client.XClient) {
		rpcgame := (*m)[idGame]
		if rpcgame != nil {
			(*rpcgame).Close()
		}
		delete(*m, idGame)
	})
	s.GameLock.CustomLock()
	delete(s.GameServers, idGame)
	s.GameLock.CustomUnLock()
	xlog.Logger().Infoln("game_", idGame, ":断开")

	// 通知客户端服务器变化
	// GetPlayerMgr().NotifyGameServerChange([]string{})
}

// 更新服务器状态
func (s *Server) UpdateServerStatus(idGame, status int) {
	s.GameLock.CustomLock()
	defer s.GameLock.CustomUnLock()
	gameserver, ok := s.GameServers[idGame]
	if !ok {
		return
	}
	gameserver.Status = status
	xlog.Logger().Infoln("game_", idGame, ":维护")

	// 通知客户端服务器变化
	// GetPlayerMgr().NotifyGameServerChange([]string{})
}

// 根据子游戏获取服务器列表
func (s *Server) GetServersByKindId(kindids []int) []int {
	result := make([]int, 0)
	isInArray := func(v int, arr []int) bool {
		for _, item := range arr {
			if v == item {
				return true
			}
		}
		return false
	}

	for gameid, value := range s.GameServers {
		// 判断服务器是否维护
		if value.Status == consts.ServerStatusMaintain {
			continue
		}

		// 判断是否支持子游戏
		for kindid, _ := range value.GameTypes {
			if isInArray(kindid, kindids) {
				result = append(result, gameid)
				break
			}
		}
	}
	return result
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

//! 更新游戏服务器桌子数
//func (self *Server) UpdateTableNum(idGame, tablenum int) {
//	self.GameLock.CustomLock()
//	defer self.GameLock.CustomUnLock()
//	syslog.Logger().Errorln("UpdateTableNum:gameid->", idGame, "tablenum->", tablenum)
//	gameserver, ok := self.GameServers[idGame]
//	if !ok {
//		return
//	}
//	gameserver.TableNum = tablenum
//}

/*
财富更新通用方法, 需要注意以下几点：
1. wealthType为财富类型, 不同财富类型操作的用户属性不一样
2. cost为正常消耗/系统返还的数值, 负数消耗, 正数返还
*/
func updateWealth(uid int64, wealthType int8, cost int, costType int8) (int64, error) {
	var afterNum int64
	switch wealthType {
	case consts.WealthTypeCard: // 房卡消耗
		tx := GetDBMgr().db_M.Begin()
		_, afterCard, _, _, err := wealthtalk.UpdateCard(uid, cost, 0, wealthType, tx)
		if err != nil {
			xlog.Logger().Error(err)
			tx.Rollback()
			return 0, err
		}

		tx.Commit()

		// 更新用户信息
		GetDBMgr().db_R.UpdatePersonAttrs(uid, "Card", afterCard)
		// 通知客户端
		hp := GetPlayerMgr().GetPlayer(uid)
		if hp != nil {
			hp.Info.Card = afterCard
			hp.UpdCard(costType)
		}
		afterNum = int64(afterCard)
	case consts.WealthTypeGold: // 金币消耗
		defaultHouseID := GetServer().ConHouse.DefaultHouse
		if defaultHouseID >= MIN_House_ID && defaultHouseID <= MAX_House_ID {
			// tx := GetDBMgr().db_M.Begin()
			house := GetClubMgr().GetClubHouseByHId(defaultHouseID)
			if house == nil {
				// tx.Rollback()
				return 0, fmt.Errorf("cost vitamin for house %d member %d: house is nil", defaultHouseID, uid)
			}
			mem := house.GetMemByUId(uid)
			if mem == nil {
				// tx.Rollback()
				return 0, fmt.Errorf("cost vitamin for house %d member %d: member is nil", defaultHouseID, uid)
			}
			_, afterVitamin, err := mem.VitaminIncrement(0, int64(cost), models.VitaminChangeType(costType), nil)
			if err != nil {
				// tx.Rollback()
				xlog.Logger().Error(err)
				return 0, err
			}
			// tx.Commit()
			afterNum = afterVitamin
		} else {
			tx := GetDBMgr().db_M.Begin()
			_, afterGold, err := wealthtalk.UpdateGold(uid, cost, costType, tx)
			if err != nil {
				xlog.Logger().Error(err)
				tx.Rollback()
				return 0, err
			}
			tx.Commit()

			afterNum = int64(afterGold)

			// 更新用户信息
			GetDBMgr().db_R.UpdatePersonAttrs(uid, "Gold", afterGold)
			// 通知客户端
			hp := GetPlayerMgr().GetPlayer(uid)
			if hp != nil {
				hp.Info.Gold = afterGold
				hp.UpdGold(costType, cost)
			}
		}
	case consts.WealthTypeCoupon: // 礼券
		tx := GetDBMgr().db_M.Begin()
		_, afterCoupon, err := wealthtalk.UpdateCoupon(uid, cost, costType, tx)
		if err != nil {
			xlog.Logger().Error(err)
			tx.Rollback()
			return 0, err
		}
		tx.Commit()
		afterNum = int64(afterCoupon)
		// 更新用户信息
		GetDBMgr().db_R.UpdatePersonAttrs(uid, "GoldBean", afterCoupon)
		// 通知客户端
		hp := GetPlayerMgr().GetPlayer(uid)
		if hp != nil {
			hp.Info.GoldBean = afterCoupon
			hp.UpdCoupon(costType)
		}
	case consts.WealthTypeDiamond: // 钻石
		tx := GetDBMgr().db_M.Begin()
		_, afterDiamond, err := wealthtalk.UpdateDiamond(uid, cost, costType, tx)
		if err != nil {
			xlog.Logger().Error(err)
			tx.Rollback()
			return 0, err
		}
		tx.Commit()
		afterNum = int64(afterDiamond)
		// 更新用户信息
		GetDBMgr().db_R.UpdatePersonAttrs(uid, "Diamond", afterDiamond)
		// 通知客户端
		hp := GetPlayerMgr().GetPlayer(uid)
		if hp != nil {
			hp.Info.GoldBean = afterDiamond
			hp.UpdDiamond(costType)
		}
	default:
		xlog.Logger().Info(fmt.Sprintf("unknown game wealth cost type: %d", wealthType))
	}

	return afterNum, nil
}
