package wuhan

import (
	"context"
	"encoding/json"
	"fmt"
	service2 "github.com/open-source/game/chess.git/pkg/client"
	"github.com/open-source/game/chess.git/pkg/rpc"
	lock2 "github.com/open-source/game/chess.git/pkg/xlock"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"

	"golang.org/x/net/websocket"
)

type Config struct {
	Id                    int    `json:"id"`      //! 服务器id
	Host                  string `json:"host" `   //! 服务器外网ip
	InHost                string `json:"inhost" ` //! 内网ip
	Redis                 string `json:"redis"`   //! redis
	RedisDB               int    `json:"redisdb"`
	Center                string `json:"center"`          //! 大厅服务器ip
	DB                    string `json:"db"`              //! 数据库
	RedisAuth             string `json:"redisauth"`       //! 数据库
	RcRedis               string `json:"rcredis"`         //! 战绩redis
	SafeHost              string `json:"safehost"`        //! 服务器ip_高仿地址
	Encode                int    `json:"encode"`          // 消息加密方式 0为不加密(使用加密后客户端发的消息也必须加密, 否则忽略)
	EncodeClientKey       string `json:"encodeclientkey"` // 与客户端通信的aes加密秘钥
	EncodePhpKey          string `json:"encodephpkey"`    // 与php通信的aes加密秘钥
	PubRedis              string `json:"pubredis"`        //! redis
	PubRedisDB            int    `json:"pubredisdb"`
	PubRedisAuth          string `json:"pubredisauth"` //!redis密码
	StoreRedis            string `json:"storeredis"`   //! redis
	StoreRedisDB          int    `json:"storeredisdb"`
	StoreRedisAuth        string `json:"storeredisauth"` //!redis密码
	service2.CommonConfig        // 第三方服务配置
}

// 打拱的牌库
type DGCase struct {
	KindId            int      `json:"kindid"`     //! 玩法id
	DGCaseList        []string `json:"dgcaselist"` // 通山打拱疯狂拱对应的牌库
	Lock              lock2.RWMutex
	DGCaseListBt      []string `json:"dgcaselistbt"` // 通山打拱变态拱对应的牌库
	LockBt            lock2.RWMutex
	DGCaseList8k      []string `json:"dgcaselist8k"` // 通山打拱疯狂拱对应的牌库
	Lock8k            lock2.RWMutex
	DGCaseList8kBt    []string `json:"dgcaselist8kbt"` // 通山打拱变态拱对应的牌库
	Lock8kBt          lock2.RWMutex
	DGCaseListNew     []string `json:"dgcaselistnew"` // 通山打拱新疯狂拱对应的牌库
	LockNew           lock2.RWMutex
	DGCaseList8kNew   []string `json:"dgcaselist8knew"` // 通山打拱新疯狂拱对应的牌库
	Lock8kNew         lock2.RWMutex
	DGCaseListNewBt   []string `json:"dgcaselistnewbt"` // 通山打拱新变态拱对应的牌库 4王
	LockNewBt         lock2.RWMutex
	DGCaseList8kNewBt []string `json:"dgcaselist8knewbt"` // 通山打拱新变态拱对应的牌库 8王
	Lock8kNewBt       lock2.RWMutex
	DDZBombCnt        string `json:"lcddzbombcnt"` // 利川斗地主对应的炸弹概率
	LockBombCnt       lock2.RWMutex
	ReadVer           int
	FkgFapai          int //疯狂拱发牌模式
}

type Server struct {
	Con            *Config                     //! 配置
	ConSite        []*models.ConfigSite        //! 房间场次配置
	ConGame        []*models.GameConfig        //! 游戏配置
	ConMatch       []*models.ConfigMatch       //! 游戏排位赛配置
	ConMatchAward  []*models.GameMatchCoupon   //! 游戏排位赛奖励清单配置
	ConGameControl []*models.ConfigGameControl //! 游戏控制配置
	DaCaseCon      *DGCase                     //! 打拱牌库配置
	ConServers     *models.ConfigServer        //！服務器配置
	ConHouse       *models.ConfigHouse         //! 包厢配置
	ConSpinBase    *models.ConfigSpinBase
	GameTypes      map[int][]*static.ServerGameType //! 游戏服对应包含的玩法类型,场次类型,桌子数量
	Wait           *sync.WaitGroup                  //! 同步阻塞
	ShutDown       bool                             //! 是否正在执行关闭
	Index          string
	Login          string
	MatchGameLock  *sync.RWMutex
}

func (s *Server) Host() (string, string) {
	ip := strings.Split(s.Con.Host, ":")
	return ip[0], ip[1]
}

func (s *Server) Start(ctx context.Context) error {
	// 第三方服务初始化
	service2.InitConfig(s.Con.CommonConfig)

	// 数据实例
	if GetDBMgr() == nil {
		return fmt.Errorf("数据实例初始化失败.")
	}

	// 初始化协议映射
	if Protocolworkers == nil {
		Protocolworkers = new(ProtocolWorkers)
		Protocolworkers.Init()
	}

	// 恢复牌桌数据
	RestoreTables()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	//! 服务器关闭
	var msg static.Msg_GameServer
	msg.Id = s.Con.Id
	msg.Status = consts.ServerStatusOffline
	GetServer().CallHall("NewServerMsg", "regameserver", xerrors.SuccessCode, &msg, 0)

	s.ShutDown = true
	GetTableMgr().reflushTable()
	s.Wait.Wait()
	GetDBMgr().Close()
	for _, v := range rpcCliMap {
		v.Close()
	}
	xlog.Logger().Warn("wuhan shutdown")
	return nil
}

func (s *Server) RegisterRouter(ctx context.Context) {
	return
}

func (s *Server) RegisterRPC(ctx context.Context) {
	InitCliByIP(s.Con.Center)
	rpc.NewRpcxServer(s.Con.Host, "ServerMethod", new(ServerMethod), &ServerPlugin{})
}

func (s *Server) LoadConfig(ctx context.Context) error {
	return GetDBMgr().ReadAllConfig()
}

func (s *Server) LoadServerConfig(ctx context.Context) error {
	xlog.Logger().Warn("LoadServerConfig")
	if err := static.LoadServeYaml(s, true); err != nil {
		xlog.Logger().Error("InitConfig_yaml:", err)
		return err
	}
	xlog.Logger().Info("load wuhan yaml etc succeed : ", fmt.Sprintf("%+v", s.Con))
	return nil
}

func (s *Server) SetLoggerLevel() {
	xlog.SetFileLevel("error")
}

func (s *Server) InitServerIndex() {
	if len(os.Args) > 1 {
		s.Index = os.Args[1]
	} else {
		s.Index = "10"
	}
}

func (s *Server) Run(ctx context.Context) {
	//!心跳包
	mticker := time.NewTicker(time.Second)
	defer mticker.Stop()
	for {
		<-mticker.C
		dateN := time.Now()
		d := time.Date(dateN.Year(), dateN.Month(), dateN.Day(), 0, 0, 0, 0, dateN.Location()) //获取当前日期的零点时间，否则无法正确比较临界点时间
		timeStr := time.Now().Format("15:04:05")
		t, _ := time.Parse("15:04:05", timeStr)
		for _, matchcon := range s.ConMatch {
			//如果排位赛开启了
			if matchcon != nil && matchcon.Flag == 1 {
				//日期对得上
				if matchcon.BeginDate.Unix() <= d.Unix() && matchcon.EndDate.Unix() >= d.Unix() {
					timeStrb := matchcon.BeginTime.Format("15:04:05")
					tb, _ := time.Parse("15:04:05", timeStrb)
					timeStre := matchcon.EndTime.Format("15:04:05")
					te, _ := time.Parse("15:04:05", timeStre)
					//if tb == t{
					//可以误差1秒
					if tb.Hour() == t.Hour() && tb.Minute() == t.Minute() && (tb.Second() == t.Second() || tb.Second() == t.Second()+1) {
						matchcon.State = 1 //排位赛开启
						//写库
						GetDBMgr().updataGameMatchState(matchcon, 1)
						//} else if te == t {
						//可以误差1秒
					} else if te.Hour() == t.Hour() && te.Minute() == t.Minute() && (te.Second() == t.Second() || te.Second() == t.Second()+1) {
						matchcon.State = 2 //排位赛已经结束
						//写库
						GetDBMgr().updataGameMatchState(matchcon, 2)
					}
					if time.Now().Hour() == 23 && time.Now().Minute() == 59 && (time.Now().Second() == 58 || time.Now().Second() == 59) {
						//最后一天了
						if matchcon.EndDate.Unix() == d.Unix() {
							matchcon.Flag = 0 //排位赛已经下线
							//写库
							GetDBMgr().updataGameMatchFlag(matchcon, 0)
						}
					}
					if time.Now().Hour() == 00 && time.Now().Minute() == 00 && (time.Now().Second() == 1 || time.Now().Second() == 0) {
						matchcon.State = 0 //排位赛未开启
						//写库
						GetDBMgr().updataGameMatchState(matchcon, 0)
					}
				}
			}
		}
	}
}

func (s *Server) Handle(string) error {
	return nil
}

func (s *Server) IsMulti() bool {
	return true
}

func (s *Server) ErrorDeep() int {
	return 3
}

func (s *Server) Name() string {
	return static.ServerNameSport + s.Index
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
		serverSingleton.Wait = new(sync.WaitGroup)
		serverSingleton.ConGame = make([]*models.GameConfig, 0)
		serverSingleton.ConServers = new(models.ConfigServer)
		serverSingleton.ConHouse = new(models.ConfigHouse)
		serverSingleton.ConSpinBase = new(models.ConfigSpinBase)
		serverSingleton.DaCaseCon = new(DGCase)
		serverSingleton.ConGameControl = make([]*models.ConfigGameControl, 0)
		serverSingleton.MatchGameLock = new(sync.RWMutex)
		serverSingleton.InitServerIndex()
	}

	return serverSingleton
}

// ! 关闭服务器
func (s *Server) Close() {

}

func (s *Server) RegistGame(status int) error {
	var msg static.Msg_GameServer
	msg.Id = s.Con.Id
	msg.InIp = s.Con.InHost
	msg.ExIp = s.Con.Host
	msg.SafeIp = s.Con.SafeHost
	msg.GameTypes = s.GameTypes
	msg.Status = consts.ServerStatusOnline

	_, err := s.CallHall("NewServerMsg", "regameserver", xerrors.SuccessCode, &msg, 0)

	return err
}

// ! 得到一个websocket处理句柄
func (s *Server) GetConnectHandler() websocket.Handler {
	connectHandler := func(ws *websocket.Conn) {
		if s.ShutDown { //! 关服了
			ws.Close()
			return
		}

		session := GetSessionMgr().GetNewSession(ws)
		session.Run()
	}
	return websocket.Handler(connectHandler)
}

//! 获取一个数据
//func (self *Server) DB_GetData(table string, uid int64) []byte {
//	v, err := self.Redis.Get(fmt.Sprintf("%s_%d", table, uid))
//	if err == nil {
//		return v
//	}
//	return []byte("")
//}
//
////! 改变数据
//func (self *Server) DB_SetData(table string, uid int64, value []byte) {
//
//	self.Redis.Set(fmt.Sprintf("%s_%d", table, uid), value)
//}

// ! 调用大厅服务器
func (s *Server) CallHall(method string, msghead string, errCode int16, v interface{}, uid int64) ([]byte, error) {
	args := &static.Msg_MsgBase{}
	args.Uid = uid
	args.Head = msghead
	args.ErrCode = errCode
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	args.Data = string(buf)
	reply := []byte{}
	err = CallHall(args, &reply)
	return reply, err
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

// 获取游戏配置
func (s *Server) GetMatchConfig(kindid int, siteType int) []*models.ConfigMatch {
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
func (s *Server) GetMatchAwardConfig(gameMatchTotal *models.GameMatchTotal) *models.GameMatchCoupon {
	s.MatchGameLock.RLock()
	defer s.MatchGameLock.RUnlock()

	for _, item := range s.ConMatchAward {
		if item.KindId == gameMatchTotal.KindId && item.SiteTypeStr == gameMatchTotal.SiteTypeStr && item.BeginDate == gameMatchTotal.BeginDate && item.BeginTime == gameMatchTotal.BeginTime {
			return item
		}
	}
	return nil
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

// 更新任务完成状态
// uid: 用户id，kind：完成游戏or胜利游戏，num：完成数量，kindId:游戏id  ，
func (s *Server) SendTaskCompleteState(uid int64, kind int, num int, kindId int, cardType int) bool {
	var msg static.Msg_Task_Complete
	msg.Kind = kind
	msg.Uid = uid
	msg.KindId = kindId
	msg.Num = num
	msg.CardType = cardType
	s.CallHall("NewServerMsg", consts.MsgGameTaskComplete, xerrors.SuccessCode, &msg, 0)
	return true
}

func (s *Server) GetGameControl(kindId int) *models.ConfigGameControl {
	for _, item := range s.ConGameControl {
		if item.KindId == kindId {
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

func (s *Server) OnBettleDone(uid int64, gameScore float64) {
	key := fmt.Sprintf("user_activity:spin:battle:%d", uid)
	aft, err := GetDBMgr().GetDBrControl().RedisV2.Incr(key).Result()
	if err != nil {
		xlog.Logger().Errorf("on %d bettle done err:%v", uid, err)
		return
	}
	if aft >= s.ConSpinBase.BattleRound {
		GetDBMgr().GetDBrControl().RedisV2.IncrBy(key, -s.ConSpinBase.BattleRound)
		chanceKey := fmt.Sprintf("user_activity:spin:times:%d", uid)
		GetDBMgr().GetDBrControl().RedisV2.Incr(chanceKey)
	}
	dateStr := time.Now().Format(time.DateOnly)
	rankWinKey := fmt.Sprintf("rank:winround:%s", dateStr)
	rankTotalKey := fmt.Sprintf("rank:totalround:%s", dateStr)
	pipe := GetDBMgr().GetDBrControl().RedisV2.Pipeline()
	defer pipe.Close()
	pipe.ZIncrBy(rankTotalKey, 1, static.HF_I64toa(uid))
	if gameScore > 0 {
		pipe.ZIncrBy(rankWinKey, 1, static.HF_I64toa(uid))
	}
	pipe.Expire(rankTotalKey, time.Hour*48)
	pipe.Expire(rankWinKey, time.Hour*48)
	_, err = pipe.Exec()
	if err != nil {
		xlog.Logger().Error(err)
	} else {
		xlog.Logger().Infof("%d 更新排行榜成功", uid)
	}
}
