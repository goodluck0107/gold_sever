package components

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"time"
)

/**

游戏小结算，分数计算详情解析

**/

type GameCommonJson struct {
	Id           int               `json:"id"`    //= 897
	Player       map[int64]*Player `json:"play"`  //玩家数据
	LookonPlayer map[int64]*Player `json:"lplay"` //旁观用户
	GameTimer    Time              `json:"timer"` //时间计时器
	//GameTable           *Table            //桌子
	GameStatus    byte `json:"status"`    //游戏状态
	GameEndStatus byte `json:"endstatus"` //当前小局游戏的状态

	//属性变量
	GameStartMode byte                 `json:"mode"`   //开始模式
	RenWuAble     bool                 `json:"task"`   //任务是否开启
	FriendInfo    info2.TagTableFriend `json:"friend"` //好友房信息

	//游戏变量
	SiceCount   uint16 `json:"sice"`   //骰子点数
	BankerUser  uint16 `json:"banker"` //庄家用户
	GameStarted bool   `json:"start"`  //是否开始游戏
	TableLocked bool   `json:"Lock"`   //锁定标志

	//设置变量
	HaveDeleteFangKa bool  `json:"deletefk"`  //是否已经扣过卡了
	HaveGetHongbao   bool  `json:"gethb"`     //是否已经获得红包
	CurCompleteCount byte  `json:"complete"`  //当前游戏局数
	TimeStart        int64 `json:"time"`      //游戏开始时间
	TimeEnd          int64 `json:"timeend"`   //游戏大局结束时间
	RoundTimeStart   int64 `json:"roundtime"` //一局游戏开始时间

	//记录分数
	GameTaxScore int                 `json:"score"`
	Rule         rule2.St_FriendRule `json:"rule"` //规则
	//HuType static.TagHuType //胡牌配置
	//20181209 服务器重启后，没有这个记录，小结算录入数据缺失
	GameBeginTime time.Time `json:"gameBeginTime"` //游戏开始时间

	FanScore [4]meta2.Game_mj_fan_score `json:"gamefanscore"` //游戏计算胡牌番数分数

	ContractorXiayuInfo    [4]meta2.Xiayu_Record_info   `json:"contractorxiayuinfo"`  //游戏下雨记录信息
	GameLastRoundInfo      meta2.Game_mj_lastround_info `json:"gamelastroundinfo"`    //游戏上局信息
	GangHouBuPai           [4]bool                      `json:"ganghoubupai"`         //游戏杠后补牌信息
	InvalidGangCards       []byte                       `json:"invalidgangcards"`     //无效的杠牌
	GameGenZhuangInfo      meta2.Game_mj_genzhuang_info `json:"gamegenzhuanginfo"`    //游戏跟庄信息
	BeFirstGang            bool                         `json:"befirstgang"`          //是否完成第一次杠牌
	DaoChePai              [4][2]byte                   `json:"daochepai"`            //倒车牌
	ReBanker               [4]int                       `json:"rebanker"`             //连庄记录
	CurLianGangCount       int                          `json:"curliangangcount"`     //当前这局连杠个数
	UserGuoHuCount         [4]int                       `json:"userguohucount"`       //玩家过胡次数
	UserKanCount           [4]int                       `json:"userkancount"`         //玩家数坎个数
	HasSendNo13Tip         bool                         `json:"hassendno13"`          //不够13张不能亮牌
	FirstXuanPiaoSure      bool                         `json:"firstxuanpiaosure"`    //首局定漂是否选完了
	FirstPiaoNum           [4]int                       `json:"firstpiaonum"`         //首局定漂是几
	GameUserTing           [4]bool                      `json:"userting"`             //玩家是否听牌,跟赔庄相关
	GameUserTingType       [4]uint64                    `json:"usertingtype"`         //玩家听牌的最大牌型
	QiangGangScoreSend     bool                         `json:"xugangscoresend"`      //续杠分是否发送
	QiangGangOperateScore  meta2.Msg_S_OperateScore_K5X `json:"qianggangscore"`       //续杠分
	GameShowCard           [4]meta2.ShowCard            `json:"gameshowcard"`         //亮牌数据
	GameReplay_Record      meta2.K5x_Replay_Record      `json:"gamereplayrecord"`     //回放记录
	GameHuCards            []byte                       `json:"gamehucards"`          //本局胡了哪些牌
	GameNextBanker         uint16                       `json:"gamenextbanker"`       //下一局庄家
	Chatlog                []static.ChatLogs            `json:"chatLogs"`             //聊天记录
	HasFanLaiZi            bool                         `json:"hasfanlaizi"`          //当前这局有没有翻癞子
	NormalDispatchRound    int                          `json:"normaldispatchround"`  //已完成几轮摸牌
	NormalDispatchStatus   [4]bool                      `json:"normaldispatchstatus"` //玩家是否正常摸牌(明杠抓牌不算，暗杠抓牌算抓过一次)
	AlarmStatus            [4]int                       `json:"alarmstatus"`          //玩家报警标识
	TrustPunish            bool                         `json:"trustpunish"`          //20201112 苏大强 第一局打完就可以了托管扣分了，不管是不是流局
	SendGameEnd            bool                         `json:"sendgameend"`          //20201112 苏大强 发送过小结算信息
	GangAfterOperateStatus [4]bool                      //玩家吃碰操作之后还有杠权的标识
	CardRecorder           [10][]meta2.CardRecorder     `json:"cardrecorder"`    //记牌器数据 ,假设最多10个玩家
	BCardRcdNextAble       [10]bool                     `json:"cardrcdnextable"` //记牌器数据 ,true表示新购买的记牌器下局生效
	CardRecordFlag         bool                         `json:"cardrecordflag"`  //记牌器数据 true表示可以购买记牌器也可以使用记牌 器
	HuList                 []int                        `json:"hulist"`          //20210129 苏大强 宜昌血流需要知道最后胡的人是谁，换庄用
	HuUser                 [4]bool                      `json:"huuser"`          //20210130 苏大强 宜昌血流记录已经胡过的玩家，流局用
}

func (self *GameCommonJson) GameCommonToJson(_game *Common) {
	self.Id = _game.KIND_ID                //游戏id
	self.Player = _game.PlayerInfo         //玩家数据，20181209 这个地方可能有问题
	self.LookonPlayer = _game.LookonPlayer //玩家数据，
	//LookonPlayer map[int64]*Player //旁观用户
	//self.GameTimer = _game.GameTimer //20181209 注意这个地方
	//GameTable           *Table  //桌子
	self.GameStatus = _game.GameStatus       //游戏状态
	self.GameEndStatus = _game.GameEndStatus //当前小局状态

	//属性变量
	self.GameStartMode = _game.GameStartMode //开始模式
	self.RenWuAble = _game.RenWuAble         //任务是否开启
	self.FriendInfo = _game.FriendInfo       //好友房信息

	//游戏变量
	self.SiceCount = _game.SiceCount        //骰子点数
	self.BankerUser = _game.BankerUser      //庄家
	self.GameStarted = _game.m_bGameStarted //是否开始游戏
	self.TableLocked = _game.IsTableLocked  //锁定标志

	//设置变量
	self.HaveDeleteFangKa = _game.HaveDeleteFangKa //是否已经扣过卡了
	self.HaveGetHongbao = _game.HaveGetHongbao     //是否已经获得红包
	self.CurCompleteCount = _game.CurCompleteCount //当前游戏局数
	self.TimeStart = _game.TimeStart               //游戏开始时间
	self.TimeEnd = _game.TimeEnd
	self.RoundTimeStart = _game.RoundTimeStart //一局游戏开始时间

	//记录分数
	self.GameTaxScore = _game.GameTaxScore //记录分数
	//self.Rule = _game.Rule
	//20181209 苏大强 恢复场景后，没有游戏开始时间，写入小结算的时候会报错
	self.GameBeginTime = _game.GameBeginTime
	//保存游戏胡牌番数
	self.FanScore = _game.FanScore
	//保存游戏无效的杠牌
	self.InvalidGangCards = _game.InvalidGangCards
	self.BeFirstGang = _game.BeFirstGang
	self.DaoChePai = _game.DaoChePai
	self.ReBanker = _game.ReBanker

	self.CurLianGangCount = _game.LianGangCount //此时的连杠个数
	self.UserGuoHuCount = _game.GuoHuCount      //玩家过胡次数
	self.UserKanCount = _game.KanCount          //玩家数坎个数
	self.HasSendNo13Tip = _game.HasSendNo13Tip  //不够13张不能亮牌
	//首局定漂
	self.FirstXuanPiaoSure = _game.FirstXuanPiaoSure
	self.FirstPiaoNum = _game.FirstPiaoNum

	self.GameUserTing = _game.GameUserTing
	self.GameUserTingType = _game.GameUserTingType
	self.GameReplay_Record = _game.K5xReplayRecord
	self.QiangGangScoreSend = _game.QiangGangScoreSend
	self.QiangGangOperateScore = _game.QiangGangOperateScore
	self.GameShowCard = _game.GameShowCard
	self.GameHuCards = _game.GameHuCards
	self.GameNextBanker = _game.GameNextBanker
	self.Chatlog = _game.Chatlog
	self.HasFanLaiZi = _game.HasFanLaiZi
	self.NormalDispatchRound = _game.NormalDispatchRound
	self.NormalDispatchStatus = _game.NormalDispatchStatus
	self.AlarmStatus = _game.AlarmStatus
	self.TrustPunish = _game.TrustPunish
	self.SendGameEnd = _game.SendGameEnd
	self.CardRecorder = _game.CardRecorder
	self.BCardRcdNextAble = _game.BCardRcdNextAble
	self.CardRecordFlag = _game.CardRecordFlag

}

func (self *GameCommonJson) JsonToStruct(_game *Common) {
	_game.KIND_ID = self.Id
	_game.PlayerInfo = self.Player

	for _, v := range _game.PlayerInfo {
		//syslog.Logger().Debug("断线重连玩家信息：",*v)
		v.UserStatus = static.US_OFFLINE //游戏服重启，玩家初始为离线状态
		v.Ctx.Timer.Init()               //计时事件清空
	}

	//老服务器可能没有旁观，读出来会是个Nil 做个保护处理，否则更新会出问题。
	if nil != self.LookonPlayer {
		_game.LookonPlayer = self.LookonPlayer
	}
	for _, v := range _game.LookonPlayer {
		v.UserStatus = static.US_FREE //游戏服重启，玩家初始为离线状态
		v.LookonFlag = true
		v.Ctx.Timer.Init() //计时事件清空
	}

	//LookonPlayer map[int64]*Player //旁观用户
	//_game.GameTimer = self.GameTimer
	//GameTable           *Table            //桌子
	_game.GameStatus = self.GameStatus
	_game.GameEndStatus = self.GameEndStatus

	//属性变量
	_game.GameStartMode = self.GameStartMode
	_game.RenWuAble = self.RenWuAble
	_game.FriendInfo = self.FriendInfo

	//游戏变量
	_game.SiceCount = self.SiceCount
	_game.BankerUser = self.BankerUser
	_game.m_bGameStarted = self.GameStarted
	_game.IsTableLocked = self.TableLocked

	//设置变量
	_game.HaveDeleteFangKa = self.HaveDeleteFangKa
	_game.HaveGetHongbao = self.HaveGetHongbao
	_game.CurCompleteCount = self.CurCompleteCount
	_game.TimeStart = self.TimeStart
	_game.TimeEnd = self.TimeEnd
	_game.RoundTimeStart = self.RoundTimeStart

	//记录分数
	_game.GameTaxScore = self.GameTaxScore
	_game.Rule = self.Rule
	//20181209 苏大强 恢复场景后，没有游戏开始时间，写入小结算的时候会报错
	_game.GameBeginTime = self.GameBeginTime

	//记录游戏胡牌番数
	_game.FanScore = self.FanScore

	//记录游戏无效的杠牌
	_game.InvalidGangCards = self.InvalidGangCards

	_game.BeFirstGang = self.BeFirstGang

	_game.DaoChePai = self.DaoChePai

	_game.ReBanker = self.ReBanker

	_game.LianGangCount = self.CurLianGangCount
	_game.GuoHuCount = self.UserGuoHuCount     //玩家过胡次数
	_game.KanCount = self.UserKanCount         //玩家数坎个数
	_game.HasSendNo13Tip = self.HasSendNo13Tip //不够13张不能亮牌
	_game.FirstXuanPiaoSure = self.FirstXuanPiaoSure
	_game.FirstPiaoNum = self.FirstPiaoNum
	_game.GameUserTing = self.GameUserTing
	_game.GameUserTingType = self.GameUserTingType
	_game.QiangGangScoreSend = self.QiangGangScoreSend
	_game.QiangGangOperateScore = self.QiangGangOperateScore
	_game.GameShowCard = self.GameShowCard
	_game.K5xReplayRecord = self.GameReplay_Record
	_game.GameHuCards = self.GameHuCards
	_game.GameNextBanker = self.GameNextBanker
	_game.Chatlog = self.Chatlog
	_game.HasFanLaiZi = self.HasFanLaiZi
	_game.NormalDispatchRound = self.NormalDispatchRound
	_game.NormalDispatchStatus = self.NormalDispatchStatus
	_game.AlarmStatus = self.AlarmStatus
	_game.TrustPunish = self.TrustPunish
	_game.SendGameEnd = self.SendGameEnd
	_game.CardRecorder = self.CardRecorder
	_game.BCardRcdNextAble = self.BCardRcdNextAble
	_game.CardRecordFlag = self.CardRecordFlag

}

type GameJsonSerializer struct {
	GameCommonJson
	//游戏变量
	HDAction     bool                `json:"hdaction"`     //是否可以赌海底
	GangCard     bool                `json:"gancard"`      //是否是杠后发牌
	ReplayRecord meta2.Replay_Record `json:"replayrecord"` //回放记录

	//出牌信息
	PiZiCard     byte   `json:"pizi"` //皮子牌值
	PiZiCardTmp  byte   `json:"piziTmp"`
	PiZiCards    []byte `json:"pizis"`    //多个皮子牌值
	MagicCard    byte   `json:"magic"`    //赖子牌值
	OutCardUser  uint16 `json:"outuser"`  //出牌用户
	OutCardData  byte   `json:"outcard"`  //出牌扑克
	OutCardCount byte   `json:"outcount"` //出牌数目
	HasHDGang    bool   `json:"hdgang"`   //咸宁麻将允许在剩余8张牌时（开始海底前），再杠一次，杠完后，从剩余7张牌时算开始海底
	//CheckTimeOut []int64 `json:"checktimeout"` ///<游戏中逻辑踢除时间累计

	//发牌信息
	SendCardData       byte                    `json:"sendcarddata"`      //发牌扑克
	SendCardCount      byte                    `json:"sendcardcount"`     //发牌数目
	LeftBu             byte                    `json:"leftbu"`            //剩下的补牌数目
	LeftCardCount      byte                    `json:"leftcount"`         //剩余数目
	RepertoryCard      []byte                  `json:"repertorycard"`     //库存扑克
	RepertoryCardArray meta2.CardLeftArrayMeta `json:"cardleftarraymeta"` //牌堆

	//运行变量
	ResumeUser       uint16 `json:"resumeuser"`       //还原用户
	CurrentUser      uint16 `json:"currentuser"`      //当前用户
	ProvideUser      uint16 `json:"provideuser"`      //供应用户
	ProvideCard      byte   `json:"providecard"`      //供应扑克
	LastSendCardUser uint16 `json:"lastsendcarduser"` //最近一次发牌的接收者
	LastOutCardUser  uint16 `json:"lastoutcarduser"`  //最近一次出牌人

	//状态变量
	GoUser              byte                         `json:"g0user"`              //参与海底人数
	GangFlower          bool                         `json:"ganflower"`           //杠上开花状态
	SendStatus          bool                         `json:"sendstatus"`          //发牌状态
	GangStatus          bool                         `json:"gangstatus"`          //抢杆状态
	MingGangStatus      bool                         `json:"minggang"`            //明杆状态
	IsHaiDi             bool                         `json:"ishaidi"`             //海底状态
	HuangZhuang         bool                         `json:"huangzhuang"`         //是否荒庄
	VecXiaPao           []static.Msg_C_Xiapao        `json:"vecxiapao"`           //记录每一局下跑
	PayPaoStatus        bool                         `json:"paypaostatus"`        //记录是否在买跑的状态
	ExchangeThreeStatus bool                         `json:"exchangethreestatus"` //记录是否在换三张的状态
	XuanPiaoStatus      bool                         `json:"xuanpiaostatus"`      //记录是否在选漂的状态
	VecPiao             []static.Msg_C_Piao          `json:"vecpiao"`             //记录每一局漂几
	VecGameEnd          []static.Msg_S_GameEnd       `json:"vecgameend"`          //记录每一局的结果
	VecGameDataAllP     [4][]static.CMD_S_StatusPlay `json:"vecgamedataallp"`     //记录每一局的结束时所有人的桌面数据
	Game_start          bool                         `json:"gamestart"`           //转门为下跑做的特殊设计

	//结束信息
	BChiHuCard byte `json:"chihucard"` //吃胡扑克

	//组件变量
	GameLogic logic2.BaseLogic `json:"gamelogic"` //游戏逻辑
	//20181124 苏大强 添加了放放的规则
	//GameLogic_cyff GameLogic_cyff `json:"gameLogic_cyff"` //放放游戏逻辑
	CellScore bool `json:"cellscore"` //底分是否除以10倍
	//TastRateshell  int    `json:"taskshell"`      //贝壳任务倍率
	TimeGameRecord string `json:"timegamerecord"` //记录当前日志的日期
	NameGameRecord string `json:"namegamerecord"` //记录文件名
	GameRecordNum  string `json:"gamerecordnum"`  //记录编号
	LastTime       int64  `json:"lasttime"`       // 用于断线重入 校准客户端操作时间
	//	Config static.GameConfig //游戏配置
	GameReplayId  int    `json:"gamereplayid"`
	LastGangKind  byte   `json:"lastgangkind"`    //最后杠牌类型
	LastGangIndex byte   `json:"last_gang_index"` //最后杠牌小标
	KzhDatas      [4]int `json:"kzh_datas"`       //卡字胡状态 石首捱晃
	GameState     int    `json:"gamestate"`       //游戏中状态
	GameTime      int64  `json:"gametime"`        //下次执行时间
	LastGangScore [4]int `json:"lastgangscore"`   //最后杠牌分
	//BaoQingType   [4]int `json:"baoqingtype"`   //报清类型
	//保存游戏房间记录
	GameConfig    static.GameConfig          `json:"gameConfig"`    //游戏配置
	QuanFeng      byte                       `json:"quanfeng"`      //圈风
	FengQuan      byte                       `json:"fengquan"`      //风圈
	TaiZhuangInfo meta2.TaiZhuangInfoStruct  `json:"taizhuanginfo"` //抬庄信息
	FirstGangGard byte                       `json:"firstganggard"` //首杠牌
	YangPi        bool                       `json:"yangpi"`        //恩施首杠发痞子，如果能养，就不管，不能养就要自动杠出去
	GangHotStatus bool                       `json:"ganghotstatus"` //热统状态
	OutCardIndex  [static.MAX_INDEX]byte     `json:"outcardindex"`  //记牌器
	TuoGuanPlayer [static.MAX_PLAYER_4P]bool `json:"tuoguanplayer"` //谁托管了？
	TrustCounts   [static.MAX_PLAYER_4P]byte `json:"trustcounts"`   //玩家托管次数

	CurPowerSeat           uint16                    // 超时罚分的当前权限玩家
	CurPunishSeat          uint16                    // 超时罚分的当前玩家
	CurPeriod              int                       // 超时罚分的当前阶段
	PunishStartTime        int64                     // 超时罚分的开始时间
	HasAction              [4]bool                   // 超时罚分的动作标志
	PunishCount            [4]int                    // 超时罚分的次数计数
	PlayerPunishScore      [4]int                    `json:"punishscore"` // 超时罚的分
	PlayerChihuCards       [4]meta2.UserChihuCards   //玩家胡的牌
	GangAfterOperateStatus bool                      `json:"gangafteroperatstatus"` //玩家吃碰操作之后还有杠权状态
	MaxPiao                [4]int                    `json:"max_piao"`              //最大漂分
	Tuo3                   bool                      //连续托管3局解散
	TrustCnt               [static.MAX_PLAYER_4P]int // 每个人连续托管次数
	TrustJuShu             int                       `json:"trustjushu"`         //托管局数 不限制0
	TrustLimit             int                       `json:"trustlimit"`         //托管限制 1暂停  2解散
	UserReady              [4]bool                   `json:"userready"`          //玩家是否已经准备
	M_lastOpreatorIsGang   [static.MAX_INDEX]bool    `json:"lastOpreatorIsGang"` //记录每个人的杠牌信息
	TotalScore             [4]int                    `json:"totalscore"`         //玩家累计分数
	UserAction             [4]byte                   `json:"useraction"`         //20210120 苏大强 恩施累计玩法 多牌权记录
}

// 已移除
// func (self *GameJsonSerializer) ToJson_xthh(_game *Game_mj_xthh) string {
// 	self.ToJson(&_game.Metadata)

// 	self.GameCommonToJson(&_game.Common)

// 	return public.HF_JtoA(self)
// }

func (self *GameJsonSerializer) ToJsonCommon(meta *meta2.Metadata, common *Common) string {
	self.ToJson(meta)
	self.GameCommonToJson(common)
	return static.HF_JtoA(self)
}

func (self *GameJsonSerializer) ToJson(_game *meta2.Metadata) {
	//self.HDAction = _game.b_HDAction
	self.GangCard = _game.HaveGangCard
	self.ReplayRecord = _game.ReplayRecord
	self.PiZiCard = _game.PiZiCard
	self.PiZiCards = _game.PiZiCards[:]
	self.PiZiCardTmp = _game.PiZiCardTmp
	self.MagicCard = _game.MagicCard
	self.RepertoryCard = _game.RepertoryCard[:]
	self.RepertoryCardArray = _game.RepertoryCardArray
	self.OutCardUser = _game.OutCardUser
	self.OutCardData = _game.OutCardData
	self.OutCardCount = _game.OutCardCount
	self.SendCardCount = _game.SendCardCount
	self.SendCardData = _game.SendCardData
	self.LeftBu = _game.LeftBu
	self.LeftCardCount = _game.LeftCardCount
	self.ResumeUser = _game.ResumeUser
	self.CurrentUser = _game.CurrentUser
	self.ProvideUser = _game.ProvideUser
	self.ProvideCard = _game.ProvideCard
	self.LastSendCardUser = _game.LastSendCardUser
	self.GoUser = _game.HaiDiUserCount
	self.GangFlower = _game.GangFlower
	self.SendStatus = _game.SendStatus
	self.GangStatus = _game.QiangGangStatus
	self.MingGangStatus = _game.MingGangStatus
	self.IsHaiDi = _game.IsHaiDi
	self.HuangZhuang = _game.HaveHuangZhuang
	self.PayPaoStatus = _game.PayPaoStatus
	self.ExchangeThreeStatus = _game.ExchangeThreeStatus
	self.XuanPiaoStatus = _game.XuanPiaoStatus
	self.Game_start = _game.GameStartForXiapao
	self.VecGameEnd = _game.VecGameEnd
	for i := 0; i < 4; i++ {
		self.VecGameDataAllP[i] = _game.VecGameDataAllP[i]
	}
	//self.GameLogic = _game.m_GameLogic  //每家自己的逻辑
	//HDCard       []public.TagCard //海底牌记录
	//HasHDGang    bool             //咸宁麻将允许在剩余8张牌时（开始海底前），再杠一次，杠完后，从剩余7张牌时算开始海底
	//CheckTimeOut []int64          ///<游戏中逻辑踢除时间累计
	self.BChiHuCard = _game.ChiHuCard
	self.CellScore = _game.HaveCellScore
	self.TimeGameRecord = _game.TimeGameRecord
	self.NameGameRecord = _game.NameGameRecord
	self.GameRecordNum = _game.GameRecordNum
	self.LastTime = _game.LastTime
	self.LastGangKind = _game.LastGangKind
	self.LastGangIndex = _game.LastGangIndex
	self.MaxPiao = _game.MaxPiao
	self.KzhDatas = _game.KzhDatas
	self.GameState = _game.State
	self.GameTime = _game.TheGameTime
	self.LastOutCardUser = _game.LastOutCardUser
	self.LastGangScore = _game.LastGangScore
	//self.BaoQingType = _game.m_BaoQingType
	//添加保存game记录
	self.GameConfig = _game.Config
	self.QuanFeng = _game.QuanFeng
	self.FengQuan = _game.FengQuan
	//保存抬庄信息
	self.TaiZhuangInfo = _game.TaiZhuangInfo
	self.YangPi = _game.YangPi
	self.PiZiCards = _game.PiZiCards
	self.FirstGangGard = _game.FirstGangGard
	//20191015 热铳没保存 苏大强
	self.GangHotStatus = _game.GangHotStatus
	//20191112 苏大强 记牌器保存
	self.OutCardIndex = _game.OutCardIndex
	//20191121 苏大强 托管记录
	self.TuoGuanPlayer = _game.TuoGuanPlayer
	self.TrustCounts = _game.TrustCounts

	self.CurPowerSeat = _game.CurPowerSeat
	self.CurPunishSeat = _game.CurPunishSeat
	self.CurPeriod = _game.CurPeriod
	self.PunishStartTime = _game.PunishStartTime
	self.HasAction = _game.HasAction
	self.PunishCount = _game.PunishCount
	self.PlayerPunishScore = _game.PlayerPunishScore
	self.PlayerChihuCards = _game.PlayerChihuCards
	self.GangAfterOperateStatus = _game.GangAfterOperateStatus
	self.Tuo3 = _game.Tuo3
	self.TrustCnt = _game.TrustCnt
	self.TrustLimit = _game.TrustLimit
	self.TrustJuShu = _game.TrustJuShu
	self.UserReady = _game.UserReady
	self.TotalScore = _game.TotalScore
	self.UserAction = _game.UserAction
	self.HuList = _game.HuList
	self.HuUser = _game.HuUser
}

func (self *GameJsonSerializer) Unmarsha(_game *meta2.Metadata) {
	//_game.b_HDAction = self.HDAction
	_game.HaveGangCard = self.GangCard
	_game.ReplayRecord = self.ReplayRecord
	fmt.Println(_game.ReplayRecord)
	// for k, _ := range _game.RepertoryCard {
	// 	if k >= len(self.RepertoryCard) {
	// 		break
	// 	}
	// 	_game.RepertoryCard[k] = self.RepertoryCard[k]
	// }
	_game.RepertoryCard = self.RepertoryCard[:]
	_game.RepertoryCardArray = self.RepertoryCardArray
	_game.PiZiCard = self.PiZiCard
	_game.PiZiCards = self.PiZiCards[:]
	_game.PiZiCardTmp = self.PiZiCardTmp
	_game.MagicCard = self.MagicCard
	_game.OutCardUser = self.OutCardUser
	_game.OutCardData = self.OutCardData
	_game.OutCardCount = self.OutCardCount
	_game.SendCardCount = self.SendCardCount
	_game.SendCardData = self.SendCardData
	_game.LeftBu = self.LeftBu
	_game.LeftCardCount = self.LeftCardCount
	_game.ResumeUser = self.ResumeUser
	_game.CurrentUser = self.CurrentUser
	_game.ProvideUser = self.ProvideUser
	_game.ProvideCard = self.ProvideCard
	_game.LastSendCardUser = self.LastSendCardUser
	_game.HaiDiUserCount = self.GoUser
	_game.GangFlower = self.GangFlower
	_game.SendStatus = self.SendStatus
	_game.QiangGangStatus = self.GangStatus
	_game.MingGangStatus = self.MingGangStatus
	_game.IsHaiDi = self.IsHaiDi
	_game.HaveHuangZhuang = self.HuangZhuang
	_game.PayPaoStatus = self.PayPaoStatus
	_game.ExchangeThreeStatus = self.ExchangeThreeStatus
	_game.XuanPiaoStatus = self.XuanPiaoStatus
	_game.GameStartForXiapao = self.Game_start
	_game.VecGameEnd = self.VecGameEnd
	for i := 0; i < 4; i++ {
		_game.VecGameDataAllP[i] = self.VecGameDataAllP[i]
	}
	//_game.m_GameLogic = self.GameLogic
	//HDCard       []public.TagCard //海底牌记录
	//HasHDGang    bool             //咸宁麻将允许在剩余8张牌时（开始海底前），再杠一次，杠完后，从剩余7张牌时算开始海底
	//CheckTimeOut []int64          ///<游戏中逻辑踢除时间累计
	_game.ChiHuCard = self.BChiHuCard
	_game.HaveCellScore = self.CellScore
	_game.TimeGameRecord = self.TimeGameRecord
	_game.NameGameRecord = self.NameGameRecord
	_game.GameRecordNum = self.GameRecordNum
	_game.LastTime = self.LastTime
	_game.LastGangKind = self.LastGangKind
	_game.LastGangIndex = self.LastGangIndex
	_game.KzhDatas = self.KzhDatas
	_game.MaxPiao = self.MaxPiao
	_game.State = self.GameState
	_game.TheGameTime = self.GameTime
	_game.LastOutCardUser = self.LastOutCardUser
	_game.LastGangScore = self.LastGangScore
	//_game.m_BaoQingType = self.BaoQingType
	_game.FengQuan = self.FengQuan
	_game.QuanFeng = self.QuanFeng
	//_game.ParseRule(_game.GameTable.Config.GameConfig)
	//_game.m_GameLogic.SetMagicCard(_game.MagicCard)
	//_game.m_GameLogic.SetPiZiCard(_game.PiZiCard)
	_game.PiZiCards = self.PiZiCards
	_game.FirstGangGard = self.FirstGangGard
	_game.YangPi = self.YangPi
	_game.TaiZhuangInfo = self.TaiZhuangInfo
	_game.GangHotStatus = self.GangHotStatus
	_game.OutCardIndex = self.OutCardIndex
	//20191121 苏大强 托管记录
	_game.TuoGuanPlayer = self.TuoGuanPlayer
	_game.TrustCounts = self.TrustCounts

	_game.CurPowerSeat = self.CurPowerSeat
	_game.CurPunishSeat = self.CurPunishSeat
	_game.CurPeriod = self.CurPeriod
	_game.PunishStartTime = self.PunishStartTime
	_game.HasAction = self.HasAction
	_game.PunishCount = self.PunishCount
	_game.PlayerPunishScore = self.PlayerPunishScore
	_game.PlayerChihuCards = self.PlayerChihuCards
	_game.GangAfterOperateStatus = self.GangAfterOperateStatus
	_game.Tuo3 = self.Tuo3
	_game.TrustCnt = self.TrustCnt
	_game.TrustLimit = self.TrustLimit
	_game.TrustJuShu = self.TrustJuShu
	_game.UserReady = self.UserReady
	_game.TotalScore = self.TotalScore
	_game.UserAction = self.UserAction
	_game.HuList = self.HuList
	_game.HuUser = self.HuUser
}

//获取累加的计算
func (self *GameJsonSerializer) getString() string {
	var str string
	return str
}

/////////////////////////////////////////////////*************************************************///////////////////
//打拱场景保存与恢复
type DGGameCommonJson struct {
	Id           int               `json:"id"`    //= 897
	Player       map[int64]*Player `json:"play"`  //玩家数据
	LookonPlayer map[int64]*Player `json:"lplay"` //旁观用户
	GameTimer    Time              `json:"timer"` //时间计时器
	//GameTable           *Table            //桌子
	GameStatus    byte `json:"status"`    //游戏状态
	GameEndStatus byte `json:"endstatus"` //当前小局游戏的状态

	//属性变量
	GameStartMode byte                 `json:"mode"`   //开始模式
	RenWuAble     bool                 `json:"task"`   //任务是否开启
	FriendInfo    info2.TagTableFriend `json:"friend"` //好友房信息

	//游戏变量
	BankerUser  uint16 `json:"banker"` //庄家用户
	GameStarted bool   `json:"start"`  //是否开始游戏
	TableLocked bool   `json:"Lock"`   //锁定标志

	//设置变量
	HaveDeleteFangKa   bool  `json:"deletefk"`    //是否已经扣过卡了
	HaveGetHongbao     bool  `json:"gethb"`       //是否已经获得红包
	CurCompleteCount   byte  `json:"complete"`    //当前游戏局数
	TimeStart          int64 `json:"time"`        //游戏开始时间
	TimeEnd            int64 `json:"timeend"`     //游戏大局结束时间
	RoundTimeStart     int64 `json:"roundtime"`   //一局游戏开始时间
	BAlreadyAddZongZha bool  `json:"zongzha"`     //是否已经算过总炸了
	BSendDissmissReq   bool  `json:"senddissreq"` //每小局是否已经发送了申请解散

	//记录分数
	GameTaxScore int                 `json:"score"`
	Rule         rule2.St_FriendRule `json:"rule"` //规则
	//HuType static.TagHuType //胡牌配置
	//20181209 服务器重启后，没有这个记录，小结算录入数据缺失
	GameBeginTime time.Time `json:"gameBeginTime"` //游戏开始时间

	FanScore [4]meta2.Game_mj_fan_score `json:"gamefanscore"` //游戏计算胡牌番数分数

	DismissRoomTime int `json:"dismissroomtime"` //自动解散房间时间
	OfflineRoomTime int `json:"offlineroomtime"` //离线解散房间时间

	PauseStatus         int                      `json:"pause_status"`        // 是否暂停中
	PauseUsers          []uint16                 `json:"pause_users"`         // 被暂停的玩家座位号
	SendCardOpt         static.TagSendCardInfo   `json:"send_card_opt"`       // 当前发牌信息
	VitaminLowPauseTime int                      `json:"vitaminlowpausetime"` //竞技点过低超时解散房间时间
	LaunchDismissTime   int                      `json:"launchdismisstime"`   //20191126 玩家自动发起解散房间时间
	FirstLiang          uint16                   `json:"firstliang"`          //第一个亮倒的玩家
	IsLastRoundHZ       bool                     `json:"islastroundhz"`       //上一局是否荒庄
	BShangLou           bool                     `json:"bshanglou"`
	CardRecorder        [10][]meta2.CardRecorder `json:"cardrecorder"`    //记牌器数据 ,假设最多10个玩家
	BCardRcdNextAble    [10]bool                 `json:"cardrcdnextable"` //记牌器数据 ,true表示新购买的记牌器下局生效
	CardRecordFlag      bool                     `json:"cardrecordflag"`  //记牌器数据 true表示可以购买记牌器也可以使用记牌器
}

type DGGameJsonSerializer struct {
	DGGameCommonJson
	//游戏变量
	ReplayRecord meta2.DG_Replay_Record `json:"replayrecord"` //回放记录
	ReWriteRec   byte                   `json:"rewriterec"`   //是否重复写回放数据，每小局游戏开始时清理,打拱可以在小结算中申请解散。

	//运行变量
	Banker         uint16                 `json:"banker"`        //庄家
	NextBanker     uint16                 `json:"nextbanker"`    //下一个庄家
	BankParter     uint16                 `json:"bankparter"`    //庄家的队友，2vs2模式下才有
	CurrentUser    uint16                 `json:"currentuser"`   //当前用户
	Player1        uint16                 `json:"player1"`       //3人拱中，除庄家外的另一个人
	Player2        uint16                 `json:"player2"`       //3人拱中，除庄家外的另一个人
	WhoLastOut     uint16                 `json:"wholastout"`    //上一个出牌玩家，pass的不算
	WhoReady       [meta2.MAX_PLAYER]bool `json:"whoready"`      // 谁已经完成了吼牌过程
	WhoJiaoOrQiang [meta2.MAX_PLAYER]int  `json:"jiaoorqiang"`   // 不叫和不抢要区分，有抢庄的3人拱
	WhoRobSpring   [meta2.MAX_PLAYER]int  `json:"whorobspring"`  // 玩家是否选择抢春(-1 没选择，0 不抢，1 抢春)
	RobSpringFlag  bool                   `json:"robspringflag"` // 抢春是否成功
	WhoAnti        uint16                 `json:"whoanti"`       // 谁反牌
	WhoAntic       [meta2.MAX_PLAYER]int  `json:"whoantic"`      // 选择反牌的情况（-1：没选择；0：弃；1：选择）

	//托管和离线数据
	TuoGuanPlayer  [meta2.MAX_PLAYER]bool `json:"tuoguanplayer"`  //谁托管了？
	TrustCounts    [meta2.MAX_PLAYER]byte `json:"trustcounts"`    //玩家托管次数
	AutoCardCounts [meta2.MAX_PLAYER]byte `json:"autocardcounts"` //自动出牌的次数
	WhoBreak       [meta2.MAX_PLAYER]bool `json:"whobreak"`       //谁断线了？
	BreakCounts    [meta2.MAX_PLAYER]byte `json:"breakcounts"`    // 断线次数
	TrustPlayer    []byte                 `json:"trustplayer"`    //托管玩家
	TrustOrder     [meta2.MAX_PLAYER]byte `json:"trustorder"`     //玩家托管顺序
	CurTrustJuShu  [meta2.MAX_PLAYER]byte `json:"curtrustjushu"`  //连续托管局数

	//牌数据
	AllCards_CYDG  [meta2.TS_ALLCARD]byte                  `json:"allcardscydg"`   // 所有牌，崇阳打滚使用
	AllCards       [meta2.TS_ALLCARD]byte                  `json:"allcards"`       // 所有牌
	PlayerCards    [meta2.MAX_PLAYER][static.MAX_CARD]byte `json:"playercards"`    // 玩家分到的牌
	BombSplitCards [meta2.MAX_PLAYER][static.MAX_CARD]byte `json:"bombsplitcards"` // 拆掉炸弹的牌
	LastCards      [static.ALL_CARD]byte                   `json:"lastcard"`       // 上轮洗的牌

	AllPaiOut         [meta2.MAX_PLAYER][static.MAX_CARD]byte `json:"allpaiout"`         // 出的牌
	LastPaiOut        [meta2.MAX_PLAYER][static.MAX_CARD]byte `json:"lastpaiout"`        // 上一轮出的牌
	ThePaiCount       [meta2.MAX_PLAYER]byte                  `json:"thepaicount"`       // 牌数
	LastOutType       int                                     `json:"lastouttype"`       // 最后出牌的类型
	LastOutTypeClient int                                     `json:"lastouttypeclient"` // 最后出牌的类型,客户端端的类型枚举可能和服务器不一致，如果一致就不需要这个了
	AllOutCnt         byte                                    `json:"allountcnt"`        // 几人已经出完了
	WhoPass           [meta2.MAX_PLAYER]bool                  `json:"whopass"`           // 谁放弃了（过）
	BTeamOut          [meta2.MAX_PLAYER]bool                  `json:"bteamout"`          // 我的对家是否走完了？
	WhoAllOutted      [meta2.MAX_PLAYER]bool                  `json:"whoalloutted"`      // 谁出完了，
	PlayerTurn        [meta2.MAX_PLAYER]byte                  `json:"playerturn"`        // 玩家出完牌的顺序，1游、2游、3游
	OutScorePai       [24]byte                                `json:"outscorepai"`       // 已出分牌

	OutCardSequence_CYDG [meta2.TS_ALLCARD]byte `json:"outcardsequencecydg"` //记住所有玩家的出牌数据,发牌拱需要使用
	OutCardSequence      [meta2.TS_ALLCARD]byte `json:"outcardsequence"`     //记住所有玩家的出牌数据,发牌拱需要使用
	OutSequenceIndes     byte                   `json:"outsequenceindes"`    //记住所有玩家的出牌数据的索引,发牌拱需要使用
	CarsLeftNum          [16]byte               `json:"cardsleftnum"`
	BombErrStart         uint16                 `json:"bomberrstart"`

	//积分数据
	LastScore       [meta2.MAX_PLAYER]int      `json:"lastscore"`       // 上一小局输赢金币
	Total           [meta2.MAX_PLAYER]int      `json:"total"`           // 总输赢，若干小局相加的金币
	Playerrich      [meta2.MAX_PLAYER]int      `json:"playerrich"`      // 玩家的财富金币
	CardScore       int                        `json:"cardscore"`       // 每一轮牌的分
	PlayerCardScore [meta2.MAX_PLAYER]int      `json:"playercardscore"` // 每个人最后的分数，有可能没出完的也加进去了
	XiScore         [meta2.MAX_PLAYER]int      `json:"xiscore"`         // 喜钱
	ZongZhaScore    [meta2.MAX_PLAYER]int      `json:"zongzhascore"`    // 总炸分
	PiaoScore       [meta2.MAX_PLAYER]int      `json:"piaoscore"`       // 漂分
	FaScore         [meta2.MAX_PLAYER][]uint16 `json:"fascore"`         // 炸错罚分
	HuapaiScore     [meta2.MAX_PLAYER]int      `json:"huapaiscore"`     // 花牌分
	ExtAddNum       [meta2.MAX_PLAYER][2]int   `json:"extaddnum"`       // 额外加分数据，下标0表示王的数目，下标1表示花牌数目
	XuanPiao        [meta2.MAX_PLAYER]int      `json:"xuanpiao"`        // 首轮选飘飘分记录
	//时间数据
	AutoOutTime    int   `json:"autoouttime"`    // 自动出牌时间，托管的时候出牌时间
	TimeStart      int   `json:"timestart"`      // 坐下之后，按开始的等待时间
	PlayTime       int   `json:"playtime"`       // 出牌时间
	RoarTime       int   `json:"roartime"`       // 吼牌时间
	PowerStartTime int64 `json:"powerstarttime"` //时间辅助变量

	//底
	Qiang            int    `json:"qiang"`          // 是否可以抢庄
	Base             int    `json:"infrastructure"` // 底
	Spay             int    `json:"spay"`           // 服务费
	SerPay           int    `json:"serpay"`         // 茶水费千分比
	FaOfTao          int    `json:"faoftao"`        // 逃跑的惩罚倍数
	JiangOfTao       int    `json:"jiangoftao"`     // 逃跑的补偿倍数
	AddSpecailBeishu int    `json:"addbeishu"`      // 别人有特殊牌奖励时，需要扣除分的倍数
	GXScore          [2]int `json:"gxscore"`        // 仙桃千分的贡献分
	AddXiScore       bool   `json:"addxiscore"`     //是否带喜分
	Piao             int    `json:"piao"`           // 选飘
	PiaoCount        int    `json:"piaocount"`      // 0每局飘一次，1首局飘一次
	//好友房
	LongzhaDing        int  `json:"longzhading"`        // 笼炸封顶
	FakeKingValue      byte `json:"fakekingvalue"`      // 王单出算几
	FapaiMode          int  `json:"fapaimode"`          // 发牌模式，，0表示起牌拱，1表示发牌拱，2表示疯狂拱，3表示变态拱
	ShowCardNum        int  `json:"showcardnum"`        // 是否显示手牌数目
	MaxKingNum         byte `json:"maxkingnum"`         // 4王或8王
	KanJie             int  `json:"kanjie"`             //坎阶梯：5分/坎，10分/坎，20分/坎
	KanScore           int  `json:"kanscore"`           //坎分：1-20，1，2，3，4....20
	BaoDi              int  `json:"baodi"`              //保底：5-100，5，10，15....100,默认5  ,
	Has510k            int  `json:"has510k"`            //510选项：510k是炸弹和510k不是炸弹，单选项，默认选中510k是炸弹，选择510k不是炸弹，癞子用法下面是禁用不能选中
	Magic510k          int  `json:"magic510k"`          //癞子用法：癞子可组510k和癞子不可组510k，单选项，默认选中癞子可组510k
	BombStr            bool `json:"bombstr"`            //是否有摇摆
	HasPiao            bool `json:"haspiao"`            // 是否选漂
	BombErr            bool `json:"bomberr"`            // 是否炸错罚分
	BombErrBankerCount int  `json:"bomberrbankcount"`   // 跟庄炸庄次数
	SkyCnt             int  `json:"skycnt"`             // 花牌数目
	ZhuangBuJie        bool `json:"zhuangbujie"`        //庄家是否不接风
	FristOut           int  `json:"fristout"`           //首出类型 ,0黑桃3先出，庄家首出
	BiYa               int  `json:"biya"`               //有大比压 ,0有大必压，1可以不压，
	TuoGuan            int  `json:"tuoguan"`            //托管 ,0不托管，大于0托管，
	ZhaNiao            bool `json:"zhaniao"`            //是否扎鸟，红桃10是鸟
	FourTake3          bool `json:"fourtake3"`          //是否可以4带3
	BombSplit          bool `json:"bombsplit"`          //炸弹是否可拆
	QuickPass          bool `json:"quickpass"`          //是否快速过牌
	SplitCards         bool `json:"splitcards"`         //是否有切牌
	Bomb3Ace           bool `json:"bomb3ace"`           //3个A是否是炸弹
	LessTake           bool `json:"lesstake"`           //最后一手是否可以少带
	Jiao2King          int  `json:"jiao2king"`          //双王必须叫牌
	TeamCard           int  `json:"teamcard"`           //先跑可看队友牌
	KeFan              bool `json:"kefan"`              //是否可以反春
	FourTake2          bool `json:"fourtake2"`          //是否可以4带2
	FourTake1          bool `json:"fourtake1"`          //4带1是否算炸弹
	CardNum            int  `json:"cardnum"`            //牌数
	KingLai            int  `json:"kinglai"`            //王是否可做赖子，0无癞子，1有癞子不讲硬炸，2有癞子讲硬炸
	FullScoreAward     int  `json:"fullscoreaward"`     //满分奖励
	FourKingScore      int  `json:"fourkingscore"`      //4王换分
	AddDiFen           int  `json:"adddifen"`           //额外加的底分
	ShowHandCardCnt    bool `json:"showhandcardcnt"`    //是否显示手牌数
	GetLastScore       bool `json:"getlastscore"`       //是否可以捡尾分
	SeeTeamerCard      bool `json:"seeteamercard"`      //看队友牌
	Big510k            bool `json:"big510k"`            //6炸7炸不可打510k
	BombMode           bool `json:"bombmode"`           //炸弹被压无分
	ExtAdd             bool `json:"extadd"`             //额外加分，比如打出3王1花得1倍底分，跟打出8喜7喜的加分不一样
	TrusteeCost        bool `json:"trusteecost"`        //托管玩家承担所有输分
	Restart            bool `json:"restart"`            //重新发牌
	TimeOutPunish      bool `json:"timeoutpunish"`      // 超时罚分
	NotDismiss         bool `json:"notdismiss"`         //托管不解散
	FristOutMode       int  `json:"fristoutmode"`       //首出出牌类型 ,0必带黑三或最小牌，1任意出牌
	NoBomb             bool `json:"nobomb"`             //纯净玩法，勾选时不发炸弹
	BombRealTime       bool `json:"bombrealtimescore"`  //炸弹实时计分
	OutCardDismissTime int  `json:"outcarddismisstime"` //出牌时间 超时房间强制解散 -1不限制
	LessTakeFirst      bool `json:"lesstakefirst"`      //最后一手是否可以少带出完
	LessTakeNext       bool `json:"lesstakenext"`       //最后一手是否可以少带接完
	FakeKing           int  `json:"fakeking"`           //王单出当几
	QiangChun          bool `json:"qiangchun"`          //仙桃跑得快抢春
	Hard510KMode       bool `json:"hard510kmode"`       //监利开机纯510K大于四炸
	SeePartnerCards    bool `json:"seepartnercards"`    //可看队友手牌
	IsRed3First        bool `json:"red3firstout"`       //是否红桃3首出
	ZhaNiaoFen         int  `json:"zhaniaofen"`         //2 2分   5 5分   10 10分   20 翻倍
	FengDing           int  `json:"fengding"`           //封顶，0不封顶
	OnlyAuto           bool `json:"onlyauto"`           //唯一可出时是否自动1秒出牌，true表示需要自动出牌
	EndReadyCheck      bool `json:"endreadycheck"`      //小结算自动准备检测
	TrustJuShu         int  `json:"trustjushu"`         // 托管局数 不限制0
	TrustLimit         int  `json:"trustlimit"`         // 托管限制 1 暂停 2 解散

	CurPeriod         int                   `json:"curperiod"`         // 超时罚分的当前阶段
	PunishStartTime   int64                 `json:"punishstarttime"`   // 超时罚分的开始时间
	HasAction         [4]bool               `json:"hasaction"`         // 超时罚分的动作标志
	PunishCount       [4]int                `json:"punishcount"`       // 超时罚分的次数计数
	PunishScore       int                   `json:"punishscore"`       // 超时的罚分
	PlayerPunishScore [meta2.MAX_PLAYER]int `json:"playerpunishscore"` // 玩家因超时被罚的分

	CalScoreMode int  `json:"calscoremode"` //计分模式，0 累计计分，1 不累计计分
	ZongZhaFlag  bool `json:"zongzhaflag"`  //是否需要在小结算计算总炸分
	XiScoreMode  int  `json:"xiscoremode"`  //喜分模式
	No6Xi        bool `json:"no6xi"`        //6喜是否算喜分，勾选为true表示无6喜、不算喜分

	//叫牌相关数据
	MingJiFlag   bool                         `json:"mingjiflag"`   // 是否已经显示明鸡了
	DownPai      byte                         `json:"downpai"`      // 4人拱做牌时的叫牌
	DownPai3P    [static.MAX_DOWNCARDNUM]byte `json:"downpai3p"`    // 3人拱的底牌，做牌也用这个数组
	RoarPai      byte                         `json:"roarpai"`      // 叫的牌
	WhoRoar      uint16                       `json:"whoroar"`      // 谁吼牌了？
	JiabeiType   int                          `json:"jiabeitype"`   // 有抢庄的3人拱的加倍类型
	BSplited     bool                         `json:"bsplited"`     // 是否选择了切牌
	RestartCount int                          `json:"restartcount"` // 重新发牌次数
	WhoRob       uint16                       `json:"whorob"`       // 抢春玩家
	MinCard      byte                         `json:"mincard"`      // 找不到红桃3时顺延找到的牌
	JiaoPaiMate  uint16                       `json:"jiaopaimate"`  //

	WhoHas4KingScore uint16 `json:"whohas4kingscore"` // 谁用4王换分了
	WhoHas4KingPower uint16 `json:"whohas4kingpower"` // 谁拥有4王换分权限（换完就没有该牌权了）

	HasKingNum     [meta2.MAX_PLAYER]int `json:"haskinnum"`      // 每个人手牌中王的个数
	WhoHasKingBomb uint16                `json:"whohaskingbomb"` // 谁有天炸
	Who8Xi         [meta2.MAX_PLAYER]int `json:"who8xi"`         // 每个人的8喜个数
	Who7Xi         [meta2.MAX_PLAYER]int `json:"who7xi"`         // 每个人的7喜个数
	WhoSame510K    [meta2.MAX_PLAYER]int `json:"whosame510k"`    // 每个人的同色510K的个数，如1玩家共有5个510k，其中方块的2个，黑桃的2个，红桃的1个，则有两个同色510k（方块，黑桃）

	MaxScore               [meta2.MAX_PLAYER]int `json:"maxscore"`               // 最高抓分
	TotalFirstTurn         [meta2.MAX_PLAYER]int `json:"totalfirstturn"`         // 一游次数
	TotalDuPai             [meta2.MAX_PLAYER]int `json:"totaldupai"`             // 独牌次数
	WhoKingCount           [meta2.MAX_PLAYER]int `json:"whokingcount"`           // 每个人的4及以上个王的个数
	Who510kCount           [meta2.MAX_PLAYER]int `json:"who510kcount"`           // 每个人的4及以上个硬510k的个数
	WhoTotal8Xi            [meta2.MAX_PLAYER]int `json:"whototal8xi"`            // 每个人大局的8喜个数
	WhoTotal7Xi            [meta2.MAX_PLAYER]int `json:"whototal7xi"`            // 每个人大局的7喜个数
	WhoTotalGonglong       [meta2.MAX_PLAYER]int `json:"whototalgonglong"`       // 每个人大局的笼炸的笼数(大局)
	WhoGonglongCount       [meta2.MAX_PLAYER]int `json:"whogonglongcount"`       // 每个人的笼炸的个数(每小局)
	WhoToTalMore4KingCount [meta2.MAX_PLAYER]int `json:"whototalmore4kingcount"` // 每个人的4及以上个王的次数
	WhoGonglongScore       [meta2.MAX_PLAYER]int `json:"whogonglongscore"`       // 每个人的笼炸的分数(每小局)
	WinCount               [meta2.MAX_PLAYER]int `json:"wincount"`               // 每个人的胜局次数
	TotalAnti              [meta2.MAX_PLAYER]int `json:"totalanti"`              // 反牌次数

	PlayKingBomb          [meta2.MAX_PLAYER]int `json:"playkingbomb"`          // 打出的天炸，4王一起打出
	Play8Xi               [meta2.MAX_PLAYER]int `json:"play8xi"`               // 打出的8喜个数
	Play7Xi               [meta2.MAX_PLAYER]int `json:"play7xi"`               // 打出的7喜个数
	Play510K              [meta2.MAX_PLAYER]int `json:"play510k"`              // 打出的510K的个数
	WhoBombScore          [meta2.MAX_PLAYER]int `json:"whobombscore"`          // 打出的炸弹分数
	WhoOutCount           [meta2.MAX_PLAYER]int `json:"whooutscore"`           // 出牌次数
	WhoToTalChuntianCount [meta2.MAX_PLAYER]int `json:"whototalchuntiancount"` // 春天次数
	BombCount             [meta2.MAX_PLAYER]int `json:"bombcount"`             // 炸弹个数
	MaxBombCount          [meta2.MAX_PLAYER]int `json:"maxbombcount"`          // 最大炸弹个数
	Bird                  [meta2.MAX_PLAYER]int `json:"bird"`                  //抓鸟，1表示有鸟
	ValidBombCount        [meta2.MAX_PLAYER]int `json:"validbombcount"`        // 有效炸弹个数
	MaxBeiShu             [meta2.MAX_PLAYER]int `json:"maxbeishu"`             // 最大倍数
	MaxBird               [meta2.MAX_PLAYER]int `json:"maxbird"`               //最大抓鸟数
	MaxSpring             [meta2.MAX_PLAYER]int `json:"maxspring"`             //最大春天次数

	//蕲春打拱
	Play5Xi       [meta2.MAX_PLAYER]int `json:"play5xi"`       //打出的5张炸个数
	Play6Xi       [meta2.MAX_PLAYER]int `json:"play6xi"`       //打出的6张炸个数
	PlaySame510k  [meta2.MAX_PLAYER]int `json:"playsame510k"`  //打出的正510k
	PlayKingBomb2 [meta2.MAX_PLAYER]int `json:"playkingbomb2"` //打出的王炸

	//状态变量
	VecGameEnd  []static.Msg_S_DG_GameEnd                      `json:"vecgameend"`  //记录每一局的结果
	VecGameData [meta2.MAX_PLAYER][]static.CMD_S_DG_StatusPlay `json:"vecgamedata"` //记录每一局的结果

	I7_8xi   int `json:"i7_8xi"`   //7/8喜出现的次数
	I7_8King int `json:"i7_8king"` //7/8王出现的次数

	//组件变量
	GameLogic      logic2.BaseLogic `json:"gamelogic"`      //游戏逻辑
	TimeGameRecord string           `json:"timegamerecord"` //记录当前日志的日期
	NameGameRecord string           `json:"namegamerecord"` //记录文件名
	GameRecordNum  string           `json:"gamerecordnum"`  //记录编号
	LastTime       int64            `json:"lasttime"`       // 用于断线重入 校准客户端操作时间
	//	Config static.GameConfig //游戏配置
	GameReplayId int   `json:"gamereplayid"`
	GameState    int   `json:"gamestate"` //游戏中状态，1表示吼牌阶段，2表示打牌阶段
	GameType     int   `json:"gametype"`  //游戏类型 1表示普通模式2vs2，2表示吼牌模式1vs3
	GameTime     int64 `json:"gametime"`  //下次执行时间

	//保存游戏房间记录
	GameConfig static.GameConfig `json:"gameConfig"` //游戏配置
}

func (self *DGGameJsonSerializer) ToJsonTS(_game *meta2.GameMetaDG_TS) {
	self.ReplayRecord = _game.ReplayRecord
	self.ReWriteRec = _game.ReWriteRec
	self.Banker = _game.Banker
	self.NextBanker = _game.NextBanker
	self.BankParter = _game.BankParter
	self.CurrentUser = _game.Whoplay
	self.Player1 = _game.Player1
	self.Player2 = _game.Player2
	self.WhoLastOut = _game.WhoLastOut
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.WhoReady[i] = _game.WhoReady[i]
		self.TuoGuanPlayer[i] = _game.TuoGuanPlayer[i]
		self.TrustCounts[i] = _game.TrustCounts[i]
		self.AutoCardCounts[i] = _game.AutoCardCounts[i]
		self.WhoBreak[i] = _game.WhoBreak[i]
		self.BreakCounts[i] = _game.BreakCounts[i]
		self.WhoPass[i] = _game.WhoPass[i]
		self.BTeamOut[i] = _game.TeamOut[i]
		self.WhoAllOutted[i] = _game.WhoAllOutted[i]
		self.PlayerTurn[i] = _game.PlayerTurn[i]
		for j := 0; j < static.MAX_CARD; j++ {
			self.PlayerCards[i][j] = _game.PlayerCards[i][j]
			self.AllPaiOut[i][j] = _game.AllPaiOut[i][j]
			self.LastPaiOut[i][j] = _game.LastPaiOut[i][j]
		}

		self.ThePaiCount[i] = _game.ThePaiCount[i]
	}
	for i := 0; i < meta2.TS_ALLCARD; i++ {
		self.AllCards[i] = _game.AllCards[i]
		self.OutCardSequence[i] = _game.OutCardSequence[i]
	}

	self.LastOutType = _game.LastOutType
	self.LastOutTypeClient = _game.LastOutTypeClient
	self.AllOutCnt = _game.AllOutCnt
	self.OutSequenceIndes = _game.OutSequenceIndes

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.LastScore[i] = _game.LastScore[i]
		self.Total[i] = _game.Total[i]
		self.Playerrich[i] = _game.Playerrich[i]
		self.PlayerCardScore[i] = _game.PlayerCardScore[i]
		self.XiScore[i] = _game.XiScore[i]
	}
	self.CardScore = _game.CardScore
	self.AutoOutTime = _game.AutoOutTime
	self.TimeStart = _game.TimeStart
	self.PlayTime = _game.PlayTime
	self.RoarTime = _game.RoarTime
	self.PowerStartTime = _game.PowerStartTime
	self.Base = _game.Base
	self.Spay = _game.Pay
	self.SerPay = _game.SerPay
	self.FaOfTao = _game.FaOfTao
	self.JiangOfTao = _game.JiangOfTao
	self.AddSpecailBeishu = _game.AddSpecailBeishu
	self.LongzhaDing = _game.LongzhaDing
	self.FakeKingValue = _game.FakeKingValue
	self.FapaiMode = _game.FapaiMode
	self.ShowCardNum = _game.ShowCardNum
	self.MaxKingNum = _game.MaxKingNum
	self.KanJie = _game.KanJie
	self.KanScore = _game.KanScore
	self.BaoDi = _game.BaoDi
	self.Has510k = _game.Has510k
	self.KanJie = _game.Magic510k

	self.MingJiFlag = _game.MingJiFlag
	self.DownPai = _game.DownPai
	self.RoarPai = _game.RoarPai
	self.WhoRoar = _game.WhoRoar
	self.WhoHasKingBomb = _game.WhoHasKingBomb

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		self.DownPai3P[i] = _game.DownPai3P[i]
	}

	for i := 0; i < 2; i++ {
		self.GXScore[i] = _game.GXScore[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.HasKingNum[i] = _game.HasKingNum[i]
		self.Who8Xi[i] = _game.Who8Xi[i]
		self.Who7Xi[i] = _game.Who7Xi[i]
		self.WhoSame510K[i] = _game.WhoSame510K[i]
		self.MaxScore[i] = _game.MaxScore[i]
		self.TotalFirstTurn[i] = _game.TotalFirstTurn[i]
		self.TotalDuPai[i] = _game.TotalDuPai[i]
		self.WhoKingCount[i] = _game.WhoKingCount[i]
		self.Who510kCount[i] = _game.Who510kCount[i]
		self.WhoTotal8Xi[i] = _game.WhoTotal8Xi[i]
		self.WhoTotal7Xi[i] = _game.WhoTotal7Xi[i]
		self.WhoTotalGonglong[i] = _game.WhoTotalGonglong[i]
		self.WhoGonglongCount[i] = _game.WhoGonglongCount[i]
		self.WhoToTalMore4KingCount[i] = _game.WhoToTalMore4KingCount[i]
		self.WinCount[i] = _game.WinCount[i]
		self.WhoGonglongScore[i] = _game.WhoGonglongScore[i]
		self.PlayKingBomb[i] = _game.PlayKingBomb[i]
		self.Play8Xi[i] = _game.Play8Xi[i]
		self.Play7Xi[i] = _game.Play7Xi[i]
		self.Play510K[i] = _game.Play510K[i]
	}

	self.VecGameEnd = _game.VecGameEnd
	self.I7_8xi = _game.Card78xiCount
	self.I7_8King = _game.Card78KingCount
	self.TimeGameRecord = _game.TimeGameRecord
	self.NameGameRecord = _game.NameGameRecord
	self.GameRecordNum = _game.GameRecordNum
	self.LastTime = _game.LastTime
	self.GameState = _game.GameState
	self.GameType = _game.GameType
	//添加保存game记录
	self.GameConfig = _game.Config
}

func (self *DGGameCommonJson) GameCommonToJson(_game *Common) {
	self.Id = _game.KIND_ID                //游戏id
	self.Player = _game.PlayerInfo         //玩家数据，20181209 这个地方可能有问题
	self.LookonPlayer = _game.LookonPlayer //玩家数据
	//LookonPlayer map[int64]*Player //旁观用户
	//self.GameTimer = _game.GameTimer //20181209 注意这个地方
	//GameTable           *Table  //桌子
	self.GameStatus = _game.GameStatus       //游戏状态
	self.GameEndStatus = _game.GameEndStatus //当前小局状态

	//属性变量
	self.GameStartMode = _game.GameStartMode //开始模式
	self.RenWuAble = _game.RenWuAble         //任务是否开启
	self.FriendInfo = _game.FriendInfo       //好友房信息

	//游戏变量
	self.BankerUser = _game.BankerUser      //庄家
	self.GameStarted = _game.m_bGameStarted //是否开始游戏
	self.TableLocked = _game.IsTableLocked  //锁定标志

	//设置变量
	self.HaveDeleteFangKa = _game.HaveDeleteFangKa //是否已经扣过卡了
	self.HaveGetHongbao = _game.HaveGetHongbao     //是否已经获得红包
	self.CurCompleteCount = _game.CurCompleteCount //当前游戏局数
	self.TimeStart = _game.TimeStart               //游戏开始时间
	self.TimeEnd = _game.TimeEnd
	self.RoundTimeStart = _game.RoundTimeStart         //一局游戏开始时间
	self.BAlreadyAddZongZha = _game.BAlreadyAddZongZha //是否已经算过总炸了
	self.BSendDissmissReq = _game.BSendDissmissReq
	self.FirstLiang = _game.FirstLiang
	self.IsLastRoundHZ = _game.IsLastRoundHZ
	self.BShangLou = _game.BShangLou
	self.CardRecorder = _game.CardRecorder
	self.BCardRcdNextAble = _game.BCardRcdNextAble
	self.CardRecordFlag = _game.CardRecordFlag

	//记录分数
	self.GameTaxScore = _game.GameTaxScore //记录分数
	//self.Rule = _game.Rule
	//20181209 苏大强 恢复场景后，没有游戏开始时间，写入小结算的时候会报错
	self.GameBeginTime = _game.GameBeginTime
	//保存游戏胡牌番数
	self.FanScore = _game.FanScore
	//保存游戏无效的杠牌
	self.DismissRoomTime = _game.DismissRoomTime
	self.OfflineRoomTime = _game.OfflineRoomTime
	self.PauseStatus = _game.PauseStatus
	self.PauseUsers = _game.PauseUsers
	self.SendCardOpt = _game.SendCardOpt
	self.VitaminLowPauseTime = _game.VitaminLowPauseTime
	self.LaunchDismissTime = _game.LaunchDismissTime
}

func (self *DGGameCommonJson) JsonToStruct(_game *Common) {
	_game.KIND_ID = self.Id
	_game.PlayerInfo = self.Player

	for _, v := range _game.PlayerInfo {
		//syslog.Logger().Debug("断线重连玩家信息：",*v)
		v.UserStatus = static.US_OFFLINE //游戏服重启，玩家初始为离线状态
		v.Ctx.Timer.Init()               //计时事件清空
	}

	//老服务器可能没有旁观，读出来会是个Nil 做个保护处理，否则更新会出问题。
	if nil != self.LookonPlayer {
		_game.LookonPlayer = self.LookonPlayer
	}
	for _, v := range _game.LookonPlayer {
		v.UserStatus = static.US_FREE //游戏服重启，玩家初始为离线状态
		v.LookonFlag = true
		v.Ctx.Timer.Init() //计时事件清空
	}

	//LookonPlayer map[int64]*Player //旁观用户
	//_game.GameTimer = self.GameTimer
	//GameTable           *Table            //桌子
	_game.GameStatus = self.GameStatus
	_game.GameEndStatus = self.GameEndStatus

	//属性变量
	_game.GameStartMode = self.GameStartMode
	_game.RenWuAble = self.RenWuAble
	_game.FriendInfo = self.FriendInfo

	//游戏变量
	_game.BankerUser = self.BankerUser
	_game.m_bGameStarted = self.GameStarted
	_game.IsTableLocked = self.TableLocked

	//设置变量
	_game.HaveDeleteFangKa = self.HaveDeleteFangKa
	_game.HaveGetHongbao = self.HaveGetHongbao
	_game.CurCompleteCount = self.CurCompleteCount
	_game.TimeStart = self.TimeStart
	_game.TimeEnd = self.TimeEnd
	_game.RoundTimeStart = self.RoundTimeStart
	_game.BAlreadyAddZongZha = self.BAlreadyAddZongZha
	_game.BSendDissmissReq = self.BSendDissmissReq
	_game.FirstLiang = self.FirstLiang
	_game.IsLastRoundHZ = self.IsLastRoundHZ
	_game.BShangLou = self.BShangLou
	_game.CardRecorder = self.CardRecorder
	_game.BCardRcdNextAble = self.BCardRcdNextAble
	_game.CardRecordFlag = self.CardRecordFlag

	//记录分数
	_game.GameTaxScore = self.GameTaxScore
	_game.Rule = self.Rule
	//20181209 苏大强 恢复场景后，没有游戏开始时间，写入小结算的时候会报错
	_game.GameBeginTime = self.GameBeginTime

	//记录游戏胡牌番数
	_game.FanScore = self.FanScore

	_game.DismissRoomTime = self.DismissRoomTime

	_game.OfflineRoomTime = self.OfflineRoomTime

	_game.PauseUsers = self.PauseUsers

	_game.PauseStatus = self.PauseStatus

	_game.SendCardOpt = self.SendCardOpt

	_game.VitaminLowPauseTime = self.VitaminLowPauseTime

	_game.LaunchDismissTime = self.LaunchDismissTime
}

func (self *DGGameJsonSerializer) UnmarshaTS(_game *meta2.GameMetaDG_TS) {

	_game.ReplayRecord = self.ReplayRecord
	fmt.Println(_game.ReplayRecord)
	// for k, _ := range _game.RepertoryCard {
	// 	if k >= len(self.RepertoryCard) {
	// 		break
	// 	}
	// 	_game.RepertoryCard[k] = self.RepertoryCard[k]
	// }

	_game.ReWriteRec = self.ReWriteRec
	_game.Banker = self.Banker
	_game.NextBanker = self.NextBanker
	_game.BankParter = self.BankParter
	_game.Whoplay = self.CurrentUser
	_game.Player1 = self.Player1
	_game.Player2 = self.Player2
	_game.WhoLastOut = self.WhoLastOut

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.WhoReady[i] = self.WhoReady[i]
		_game.TuoGuanPlayer[i] = self.TuoGuanPlayer[i]
		_game.TrustCounts[i] = self.TrustCounts[i]
		_game.AutoCardCounts[i] = self.AutoCardCounts[i]
		_game.WhoBreak[i] = self.WhoBreak[i]
		_game.BreakCounts[i] = self.BreakCounts[i]
		_game.WhoPass[i] = self.WhoPass[i]
		_game.TeamOut[i] = self.BTeamOut[i]
		_game.WhoAllOutted[i] = self.WhoAllOutted[i]
		_game.PlayerTurn[i] = self.PlayerTurn[i]
		for j := 0; j < static.MAX_CARD; j++ {
			_game.PlayerCards[i][j] = self.PlayerCards[i][j]
			_game.AllPaiOut[i][j] = self.AllPaiOut[i][j]
			_game.LastPaiOut[i][j] = self.LastPaiOut[i][j]
		}

		_game.ThePaiCount[i] = self.ThePaiCount[i]
	}
	for i := 0; i < meta2.TS_ALLCARD; i++ {
		_game.AllCards[i] = self.AllCards[i]
		_game.OutCardSequence[i] = self.OutCardSequence[i]
	}

	_game.LastOutType = self.LastOutType
	_game.LastOutTypeClient = self.LastOutTypeClient
	_game.AllOutCnt = self.AllOutCnt
	_game.OutSequenceIndes = self.OutSequenceIndes

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.LastScore[i] = self.LastScore[i]
		_game.Total[i] = self.Total[i]
		_game.Playerrich[i] = self.Playerrich[i]
		_game.PlayerCardScore[i] = self.PlayerCardScore[i]
		_game.XiScore[i] = self.XiScore[i]
	}
	_game.CardScore = self.CardScore
	_game.AutoOutTime = self.AutoOutTime
	_game.TimeStart = self.TimeStart
	_game.PlayTime = self.PlayTime
	_game.RoarTime = self.RoarTime
	_game.PowerStartTime = self.PowerStartTime
	_game.Base = self.Base
	_game.Pay = self.Spay
	_game.SerPay = self.SerPay
	_game.FaOfTao = self.FaOfTao
	_game.JiangOfTao = self.JiangOfTao
	_game.AddSpecailBeishu = self.AddSpecailBeishu
	_game.LongzhaDing = self.LongzhaDing
	_game.FakeKingValue = self.FakeKingValue
	_game.FapaiMode = self.FapaiMode
	_game.ShowCardNum = self.ShowCardNum
	_game.MaxKingNum = self.MaxKingNum
	_game.KanJie = self.KanJie
	_game.KanScore = self.KanScore
	_game.BaoDi = self.BaoDi
	_game.Has510k = self.Has510k
	_game.Magic510k = self.KanJie

	_game.MingJiFlag = self.MingJiFlag
	_game.DownPai = self.DownPai
	_game.RoarPai = self.RoarPai
	_game.WhoRoar = self.WhoRoar
	_game.WhoHasKingBomb = self.WhoHasKingBomb

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		_game.DownPai3P[i] = self.DownPai3P[i]
	}
	for i := 0; i < 2; i++ {
		_game.GXScore[i] = self.GXScore[i]
	}
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.HasKingNum[i] = self.HasKingNum[i]
		_game.Who8Xi[i] = self.Who8Xi[i]
		_game.Who7Xi[i] = self.Who7Xi[i]
		_game.WhoSame510K[i] = self.WhoSame510K[i]
		_game.MaxScore[i] = self.MaxScore[i]
		_game.TotalFirstTurn[i] = self.TotalFirstTurn[i]
		_game.TotalDuPai[i] = self.TotalDuPai[i]
		_game.WhoKingCount[i] = self.WhoKingCount[i]
		_game.Who510kCount[i] = self.Who510kCount[i]
		_game.WhoTotal8Xi[i] = self.WhoTotal8Xi[i]
		_game.WhoTotal7Xi[i] = self.WhoTotal7Xi[i]
		_game.WhoTotalGonglong[i] = self.WhoTotalGonglong[i]
		_game.WhoGonglongCount[i] = self.WhoGonglongCount[i]
		_game.WhoToTalMore4KingCount[i] = self.WhoToTalMore4KingCount[i]
		_game.WinCount[i] = self.WinCount[i]
		_game.WhoGonglongScore[i] = self.WhoGonglongScore[i]
		_game.PlayKingBomb[i] = self.PlayKingBomb[i]
		_game.Play8Xi[i] = self.Play8Xi[i]
		_game.Play7Xi[i] = self.Play7Xi[i]
		_game.Play510K[i] = self.Play510K[i]
	}

	_game.VecGameEnd = self.VecGameEnd
	_game.Card78xiCount = self.I7_8xi
	_game.Card78KingCount = self.I7_8King
	_game.TimeGameRecord = self.TimeGameRecord
	_game.NameGameRecord = self.NameGameRecord
	_game.GameRecordNum = self.GameRecordNum
	_game.LastTime = self.LastTime
	_game.GameState = self.GameState
	_game.GameType = self.GameType

	//_game.Config             =        self.GameConfig
}

func (self *DGGameJsonSerializer) ToJsonCB(_game *meta2.GameMetaDG_CB) {
	self.ReplayRecord = _game.ReplayRecord
	self.ReWriteRec = _game.ReWriteRec
	self.Banker = _game.Banker
	self.NextBanker = _game.NextBanker
	self.BankParter = _game.BankParter
	self.CurrentUser = _game.WhoPlay
	self.Player1 = _game.Player1
	self.Player2 = _game.Player2
	self.WhoLastOut = _game.WhoLastOut
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.WhoReady[i] = _game.WhoReady[i]
		self.TuoGuanPlayer[i] = _game.TuoGuanPlayer[i]
		self.TrustCounts[i] = _game.TrustCounts[i]
		self.AutoCardCounts[i] = _game.AutoCardCounts[i]
		self.WhoBreak[i] = _game.WhoBreak[i]
		self.BreakCounts[i] = _game.BreakCounts[i]
		self.WhoPass[i] = _game.WhoPass[i]
		self.BTeamOut[i] = _game.TeamOut[i]
		self.WhoAllOutted[i] = _game.WhoAllOutted[i]
		self.PlayerTurn[i] = _game.PlayerTurn[i]
		self.TrustOrder[i] = _game.TrustOrder[i]
		for j := 0; j < static.MAX_CARD; j++ {
			self.PlayerCards[i][j] = _game.PlayerCards[i][j]
			self.AllPaiOut[i][j] = _game.AllPaiOut[i][j]
			self.LastPaiOut[i][j] = _game.LastPaiOut[i][j]
		}

		self.ThePaiCount[i] = _game.ThePaiCount[i]
	}
	self.OutScorePai = _game.OutScorePai
	self.TrustPlayer = _game.TrustPlayer[:]
	for i := 0; i < static.ALL_CARD; i++ {
		self.AllCards[i] = _game.AllCards[i]
		self.OutCardSequence[i] = _game.OutCardSequence[i]
	}
	//崇阳打滚使用的
	for i := 0; i < info2.CYDG_CARDS; i++ {
		self.AllCards_CYDG[i] = _game.AllCards_CYDG[i]
		self.OutCardSequence_CYDG[i] = _game.OutCardSequence_CYDG[i]
	}

	self.LastOutType = _game.LastOutType
	self.LastOutTypeClient = _game.LastOutTypeClient
	self.AllOutCnt = _game.AllOutCnt
	self.OutSequenceIndes = _game.OutSequenceIndes
	self.CarsLeftNum = _game.CarsLeftNum
	self.BombErrStart = _game.BombErrStart

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.LastScore[i] = _game.LastScore[i]
		self.Total[i] = _game.Total[i]
		self.Playerrich[i] = _game.PlayeRich[i]
		self.PlayerCardScore[i] = _game.PlayerCardScore[i]
		self.XiScore[i] = _game.XiScore[i]
		self.ZongZhaScore[i] = _game.ZongZhaScore[i]
		self.PiaoScore[i] = _game.PiaoScore[i]
		copy(self.FaScore[i], _game.FaScore[i])
		self.HuapaiScore[i] = _game.HuapaiScore[i]
		copy(self.ExtAddNum[i][:], _game.ExtAddNum[i][:])
	}
	self.CardScore = _game.CardScore
	self.AutoOutTime = _game.AutoOutTime
	self.TimeStart = _game.TimeStart
	self.PlayTime = _game.PlayTime
	self.RoarTime = _game.RoarTime
	self.PowerStartTime = _game.PowerStartTime
	self.Base = _game.Base
	self.Spay = _game.Pay
	self.SerPay = _game.SerPay
	self.FaOfTao = _game.FaOfTao
	self.JiangOfTao = _game.JiangOfTao
	self.AddSpecailBeishu = _game.AddSpecailBeishu
	self.LongzhaDing = _game.LongzhaDing
	self.FakeKingValue = _game.FakeKingValue
	self.FapaiMode = _game.FapaiMode
	self.ShowCardNum = _game.ShowCardNum
	self.MaxKingNum = _game.MaxKingNum
	self.BombStr = _game.HasBombStr
	self.BombErr = _game.HasBombErr
	self.HasPiao = _game.HasPiao
	self.SkyCnt = _game.SkyCnt
	self.ZhuangBuJie = _game.ZhuangBuJie
	self.BombErrBankerCount = _game.BombErrBankerCount
	self.ExtAdd = _game.ExtAdd
	self.TrusteeCost = _game.TrusteeCost
	self.TimeOutPunish = _game.TimeOutPunish
	self.CurPeriod = _game.CurPeriod
	self.PunishStartTime = _game.PunishStartTime
	self.HasAction = _game.HasAction
	self.PunishCount = _game.PunishCount
	self.AddXiScore = _game.AddXiScore
	self.BAlreadyAddZongZha = _game.BHaveZongZha
	self.CalScoreMode = _game.CalScoreMode
	self.No6Xi = _game.No6Xi

	self.MingJiFlag = _game.MingJiFlag
	self.DownPai = _game.DownPai
	self.RoarPai = _game.RoarPai
	self.WhoRoar = _game.WhoRoar
	self.WhoHasKingBomb = _game.WhoHasKingBomb
	self.ZongZhaFlag = _game.ZongZhaFlag

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		self.DownPai3P[i] = _game.DownPai3P[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.HasKingNum[i] = _game.HasKingNum[i]
		self.Who8Xi[i] = _game.Who8Xi[i]
		self.Who7Xi[i] = _game.Who7Xi[i]
		self.WhoSame510K[i] = _game.WhoSame510K[i]
		self.MaxScore[i] = _game.MaxScore[i]
		self.TotalFirstTurn[i] = _game.TotalFirstTurn[i]
		self.TotalDuPai[i] = _game.TotalDuPai[i]
		self.WhoKingCount[i] = _game.WhoKingCount[i]
		self.Who510kCount[i] = _game.Who510kCount[i]
		self.WhoTotal8Xi[i] = _game.WhoTotal8Xi[i]
		self.WhoTotal7Xi[i] = _game.WhoTotal7Xi[i]
		self.WhoTotalGonglong[i] = _game.WhoTotalGonglong[i]
		self.WhoGonglongCount[i] = _game.WhoGonglongCount[i]
		self.WhoToTalMore4KingCount[i] = _game.WhoToTalMore4KingCount[i]
		self.WhoGonglongScore[i] = _game.WhoGonglongScore[i]
		self.PlayKingBomb[i] = _game.PlayKingBomb[i]
		self.Play8Xi[i] = _game.Play8Xi[i]
		self.Play7Xi[i] = _game.Play7Xi[i]
		self.Play510K[i] = _game.Play510K[i]
		self.WhoBombScore[i] = _game.WhoBombScore[i]
		self.WhoOutCount[i] = _game.WhoOutCount[i]
		self.WhoToTalChuntianCount[i] = _game.WhoToTalChuntianCount[i]
	}

	self.VecGameEnd = _game.VecGameEnd
	self.TimeGameRecord = _game.TimeGameRecord
	self.NameGameRecord = _game.NameGameRecord
	self.GameRecordNum = _game.GameRecordNum
	self.LastTime = _game.LastTime
	self.GameState = _game.GameState
	self.GameType = _game.GameType
	//添加保存game记录
	self.GameConfig = _game.Config
}

func (self *DGGameJsonSerializer) UnmarshaCB(_game *meta2.GameMetaDG_CB) {

	_game.ReplayRecord = self.ReplayRecord
	fmt.Println(_game.ReplayRecord)
	// for k, _ := range _game.RepertoryCard {
	// 	if k >= len(self.RepertoryCard) {
	// 		break
	// 	}
	// 	_game.RepertoryCard[k] = self.RepertoryCard[k]
	// }
	_game.ReWriteRec = self.ReWriteRec
	_game.Banker = self.Banker
	_game.NextBanker = self.NextBanker
	_game.BankParter = self.BankParter
	_game.WhoPlay = self.CurrentUser
	_game.Player1 = self.Player1
	_game.Player2 = self.Player2
	_game.WhoLastOut = self.WhoLastOut

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.WhoReady[i] = self.WhoReady[i]
		_game.TuoGuanPlayer[i] = self.TuoGuanPlayer[i]
		_game.TrustCounts[i] = self.TrustCounts[i]
		_game.AutoCardCounts[i] = self.AutoCardCounts[i]
		_game.WhoBreak[i] = self.WhoBreak[i]
		_game.BreakCounts[i] = self.BreakCounts[i]
		_game.WhoPass[i] = self.WhoPass[i]
		_game.TeamOut[i] = self.BTeamOut[i]
		_game.WhoAllOutted[i] = self.WhoAllOutted[i]
		_game.PlayerTurn[i] = self.PlayerTurn[i]
		_game.TrustOrder[i] = self.TrustOrder[i]
		for j := 0; j < static.MAX_CARD; j++ {
			_game.PlayerCards[i][j] = self.PlayerCards[i][j]
			_game.AllPaiOut[i][j] = self.AllPaiOut[i][j]
			_game.LastPaiOut[i][j] = self.LastPaiOut[i][j]
		}

		_game.ThePaiCount[i] = self.ThePaiCount[i]
	}
	_game.OutScorePai = self.OutScorePai
	_game.TrustPlayer = self.TrustPlayer[:]
	for i := 0; i < static.ALL_CARD; i++ {
		_game.AllCards[i] = self.AllCards[i]
		_game.OutCardSequence[i] = self.OutCardSequence[i]
	}
	//崇阳打滚使用的
	for i := 0; i < info2.CYDG_CARDS; i++ {
		_game.AllCards_CYDG[i] = self.AllCards_CYDG[i]
		_game.OutCardSequence_CYDG[i] = self.OutCardSequence_CYDG[i]
	}
	_game.LastOutType = self.LastOutType
	_game.LastOutTypeClient = self.LastOutTypeClient
	_game.AllOutCnt = self.AllOutCnt
	_game.OutSequenceIndes = self.OutSequenceIndes
	_game.CarsLeftNum = self.CarsLeftNum
	_game.BombErrStart = self.BombErrStart

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.LastScore[i] = self.LastScore[i]
		_game.Total[i] = self.Total[i]
		_game.PlayeRich[i] = self.Playerrich[i]
		_game.PlayerCardScore[i] = self.PlayerCardScore[i]
		_game.XiScore[i] = self.XiScore[i]
		_game.ZongZhaScore[i] = self.ZongZhaScore[i]
		copy(_game.FaScore[i], self.FaScore[i])
		_game.HuapaiScore[i] = self.HuapaiScore[i]
		_game.PiaoScore[i] = self.PiaoScore[i]
		copy(_game.ExtAddNum[i][:], self.ExtAddNum[i][:])
	}
	_game.CardScore = self.CardScore
	_game.AutoOutTime = self.AutoOutTime
	_game.TimeStart = self.TimeStart
	_game.PlayTime = self.PlayTime
	_game.RoarTime = self.RoarTime
	_game.PowerStartTime = self.PowerStartTime
	_game.Base = self.Base
	_game.Pay = self.Spay
	_game.SerPay = self.SerPay
	_game.FaOfTao = self.FaOfTao
	_game.JiangOfTao = self.JiangOfTao
	_game.AddSpecailBeishu = self.AddSpecailBeishu
	_game.LongzhaDing = self.LongzhaDing
	_game.FakeKingValue = self.FakeKingValue
	_game.FapaiMode = self.FapaiMode
	_game.ShowCardNum = self.ShowCardNum
	_game.MaxKingNum = self.MaxKingNum
	_game.HasBombStr = self.BombStr
	_game.HasBombErr = self.BombErr
	_game.HasPiao = self.HasPiao
	_game.SkyCnt = self.SkyCnt
	_game.ZhuangBuJie = self.ZhuangBuJie
	_game.BombErrBankerCount = self.BombErrBankerCount
	_game.ExtAdd = self.ExtAdd
	_game.TrusteeCost = self.TrusteeCost
	_game.TimeOutPunish = self.TimeOutPunish
	_game.AddXiScore = self.AddXiScore
	_game.CalScoreMode = self.CalScoreMode
	_game.No6Xi = self.No6Xi

	_game.CurPeriod = self.CurPeriod
	_game.PunishStartTime = self.PunishStartTime
	_game.HasAction = self.HasAction
	_game.PunishCount = self.PunishCount

	_game.MingJiFlag = self.MingJiFlag
	_game.DownPai = self.DownPai
	_game.RoarPai = self.RoarPai
	_game.WhoRoar = self.WhoRoar
	_game.WhoHasKingBomb = self.WhoHasKingBomb
	_game.ZongZhaFlag = self.ZongZhaFlag

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		_game.DownPai3P[i] = self.DownPai3P[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.HasKingNum[i] = self.HasKingNum[i]
		_game.Who8Xi[i] = self.Who8Xi[i]
		_game.Who7Xi[i] = self.Who7Xi[i]
		_game.WhoSame510K[i] = self.WhoSame510K[i]
		_game.MaxScore[i] = self.MaxScore[i]
		_game.TotalFirstTurn[i] = self.TotalFirstTurn[i]
		_game.TotalDuPai[i] = self.TotalDuPai[i]
		_game.WhoKingCount[i] = self.WhoKingCount[i]
		_game.Who510kCount[i] = self.Who510kCount[i]
		_game.WhoTotal8Xi[i] = self.WhoTotal8Xi[i]
		_game.WhoTotal7Xi[i] = self.WhoTotal7Xi[i]
		_game.WhoTotalGonglong[i] = self.WhoTotalGonglong[i]
		_game.WhoGonglongCount[i] = self.WhoGonglongCount[i]
		_game.WhoToTalMore4KingCount[i] = self.WhoToTalMore4KingCount[i]
		_game.WhoGonglongScore[i] = self.WhoGonglongScore[i]
		_game.PlayKingBomb[i] = self.PlayKingBomb[i]
		_game.Play8Xi[i] = self.Play8Xi[i]
		_game.Play7Xi[i] = self.Play7Xi[i]
		_game.Play510K[i] = self.Play510K[i]
		_game.WhoBombScore[i] = self.WhoBombScore[i]
		_game.WhoOutCount[i] = self.WhoOutCount[i]
		_game.WhoToTalChuntianCount[i] = self.WhoToTalChuntianCount[i]
	}

	_game.VecGameEnd = self.VecGameEnd
	_game.TimeGameRecord = self.TimeGameRecord
	_game.NameGameRecord = self.NameGameRecord
	_game.GameRecordNum = self.GameRecordNum
	_game.LastTime = self.LastTime
	_game.GameState = self.GameState
	_game.GameType = self.GameType

	//_game.Config             =        self.GameConfig
}

func (self *DGGameJsonSerializer) ToJsonDG(_game *meta2.GameMetaDG) {
	self.ReplayRecord = _game.ReplayRecord
	self.ReWriteRec = _game.ReWriteRec
	self.Banker = _game.Banker
	self.NextBanker = _game.Nextbanker
	self.BankParter = _game.BankParter
	self.CurrentUser = _game.Whoplay
	self.Player1 = _game.Player1
	self.Player2 = _game.Player2
	self.WhoLastOut = _game.WhoLastOut
	self.TrustPlayer = _game.TrustPlayer
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.WhoJiaoOrQiang[i] = _game.WhoJiaoOrQiang[i]
		self.WhoReady[i] = _game.WhoReady[i]
		self.WhoRobSpring[i] = _game.WhoRobSpring[i]
		self.TuoGuanPlayer[i] = _game.TuoGuanPlayer[i]
		self.TrustCounts[i] = _game.TrustCounts[i]
		self.AutoCardCounts[i] = _game.AutoCardCounts[i]
		self.WhoBreak[i] = _game.WhoBreak[i]
		self.BreakCounts[i] = _game.BreakCounts[i]
		self.WhoPass[i] = _game.WhoPass[i]
		self.BTeamOut[i] = _game.BTeamOut[i]
		self.WhoAllOutted[i] = _game.WhoAllOutted[i]
		self.PlayerTurn[i] = _game.PlayerTurn[i]
		for j := 0; j < static.MAX_CARD; j++ {
			self.PlayerCards[i][j] = _game.PlayerCards[i][j]
			self.AllPaiOut[i][j] = _game.AllPaiOut[i][j]
			self.LastPaiOut[i][j] = _game.LastPaiOut[i][j]
			self.BombSplitCards[i][j] = _game.BombSplitCards[i][j]
		}
		self.ThePaiCount[i] = _game.ThePaiCount[i]
		self.XuanPiao[i] = _game.XuanPiao[i]
	}
	for i := 0; i < static.ALL_CARD; i++ {
		self.AllCards[i] = _game.AllCards[i]
	}
	self.LastCards = _game.LastCards

	self.LastOutType = _game.LastOutType
	self.LastOutTypeClient = _game.LastOutTypeClient
	self.AllOutCnt = _game.AllOutCnt

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.LastScore[i] = _game.LastScore[i]
		self.Total[i] = _game.Total[i]
		self.Playerrich[i] = _game.Playerrich[i]
		self.PlayerCardScore[i] = _game.PlayerCardScore[i]
		self.XiScore[i] = _game.XiScore[i]
		self.PlayerPunishScore[i] = _game.PlayerPunishScore[i]
	}
	self.CardScore = _game.CardScore
	self.AutoOutTime = _game.AutoOutTime
	self.TimeStart = _game.TimeStart
	self.PlayTime = _game.PlayTime
	self.RoarTime = _game.RoarTime
	self.PowerStartTime = _game.PowerStartTime
	self.Qiang = _game.Qiang
	self.Base = _game.IBase
	self.Spay = _game.Spay
	self.SerPay = _game.SerPay
	self.FaOfTao = _game.FaOfTao
	self.JiangOfTao = _game.JiangOfTao
	self.AddSpecailBeishu = _game.AddSpecailBeishu
	self.FristOut = _game.FristOut
	self.BiYa = _game.BiYa
	self.TuoGuan = _game.TuoGuan
	self.ZhaNiao = _game.ZhaNiao
	self.FourTake3 = _game.FourTake3
	self.BombSplit = _game.BombSplit
	self.QuickPass = _game.QuickPass
	self.SplitCards = _game.SplitCards
	self.Bomb3Ace = _game.Bomb3Ace
	self.LessTake = _game.LessTake
	self.Jiao2King = _game.Jiao2King
	self.TeamCard = _game.TeamCard
	self.KeFan = _game.KeFan
	self.FourTake2 = _game.FourTake2
	self.FourTake1 = _game.FourTake1
	self.CardNum = _game.CardNum
	self.KingLai = _game.KingLai
	self.Big510k = _game.Big510k
	self.FullScoreAward = _game.FullScoreAward
	self.FourKingScore = _game.FourKingScore
	self.AddDiFen = _game.AddDiFen
	self.ShowHandCardCnt = _game.ShowHandCardCnt
	self.GetLastScore = _game.GetLastScore
	self.SeeTeamerCard = _game.SeeTeamerCard
	self.BombMode = _game.BombMode
	self.Restart = _game.Restart
	self.Piao = _game.Piao
	self.PiaoCount = _game.PiaoCount
	self.NotDismiss = _game.NotDismiss
	self.NoBomb = _game.NoBomb
	self.FristOutMode = _game.FristOutMode
	self.BombRealTime = _game.BombRealTime
	self.OutCardDismissTime = _game.OutCardDismissTime
	self.LessTakeFirst = _game.LessTakeFirst
	self.LessTakeNext = _game.LessTakeNext
	self.FakeKing = _game.FakeKing
	self.QiangChun = _game.QiangChun
	self.Hard510KMode = _game.Hard510KMode
	self.SeePartnerCards = _game.SeePartnerCards
	self.ZhaNiaoFen = _game.ZhaNiaoFen
	self.IsRed3First = _game.IsRed3First
	self.FengDing = _game.FengDing
	self.OnlyAuto = _game.OnlyAuto
	self.EndReadyCheck = _game.EndReadyCheck
	self.TrustJuShu = _game.TrustJuShu
	self.TrustLimit = _game.TrustLimit
	self.TimeOutPunish = _game.TimeOutPunish
	self.PunishScore = _game.PunishScore
	self.JiaoPaiMate = _game.JiaoPaiMate

	self.MingJiFlag = _game.BMingJiFlag
	self.DownPai = _game.DownPai
	self.RoarPai = _game.RoarPai
	self.WhoRoar = _game.WhoRoar
	self.WhoHasKingBomb = _game.WhoHasKingBomb
	self.JiabeiType = _game.JiabeiType
	self.BSplited = _game.BSplited
	self.AddXiScore = _game.AddXiScore
	self.RestartCount = _game.RestartCount
	self.WhoRob = _game.WhoRob
	self.RobSpringFlag = _game.RobSpringFlag
	self.OutScorePai = _game.OutScorePai
	self.MinCard = _game.MinCard
	self.WhoAnti = _game.WhoAnti
	self.WhoAntic = _game.WhoAntic
	self.CurTrustJuShu = _game.CurTrustJuShu
	self.XiScoreMode = _game.XiScoreMode

	self.WhoHas4KingScore = _game.WhoHas4KingScore
	self.WhoHas4KingPower = _game.WhoHas4KingPower

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		self.DownPai3P[i] = _game.DownPai3P[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		self.HasKingNum[i] = _game.HasKingNum[i]
		self.Who8Xi[i] = _game.Who8Xi[i]
		self.Who7Xi[i] = _game.Who7Xi[i]
		self.WhoSame510K[i] = _game.WhoSame510K[i]
		self.MaxScore[i] = _game.MaxScore[i]
		self.TotalFirstTurn[i] = _game.TotalFirstTurn[i]
		self.TotalDuPai[i] = _game.TotalDuPai[i]
		self.WhoKingCount[i] = _game.WhoKingCount[i]
		self.Who510kCount[i] = _game.Who510kCount[i]
		self.WhoTotal8Xi[i] = _game.WhoTotal8Xi[i]
		self.WhoTotal7Xi[i] = _game.WhoTotal7Xi[i]
		self.WhoTotalGonglong[i] = _game.WhoTotalGonglong[i]
		self.WhoGonglongCount[i] = _game.WhoGonglongCount[i]
		self.WhoToTalMore4KingCount[i] = _game.WhoToTalMore4KingCount[i]
		self.WhoGonglongScore[i] = _game.WhoGonglongScore[i]
		self.TotalAnti[i] = _game.TotalAnti[i]
		self.PlayKingBomb[i] = _game.PlayKingBomb[i]
		self.Play8Xi[i] = _game.Play8Xi[i]
		self.Play7Xi[i] = _game.Play7Xi[i]
		self.Play510K[i] = _game.Play510K[i]
		self.WhoBombScore[i] = _game.WhoBombScore[i]
		self.WhoOutCount[i] = _game.WhoOutCount[i]
		self.WhoToTalChuntianCount[i] = _game.WhoToTalChuntianCount[i]
		self.BombCount[i] = _game.BombCount[i]
		self.MaxBombCount[i] = _game.MaxBombCount[i]
		self.Bird[i] = _game.Bird[i]
		self.ValidBombCount[i] = _game.ValidBombCount[i]
		self.Play5Xi[i] = _game.Play5Xi[i]
		self.Play6Xi[i] = _game.Play6Xi[i]
		self.PlaySame510k[i] = _game.PlaySame510k[i]
		self.PlayKingBomb2[i] = _game.PlayKingBomb2[i]
		self.MaxBeiShu[i] = _game.MaxBeiShu[i]
		self.MaxSpring[i] = _game.MaxSpring[i]
		self.MaxBird[i] = _game.MaxBird[i]
	}

	self.VecGameData = _game.VecGameData
	self.VecGameEnd = _game.VecGameEnd
	self.TimeGameRecord = _game.TimeGameRecord
	self.NameGameRecord = _game.NameGameRecord
	self.GameRecordNum = _game.GameRecordNum
	self.LastTime = _game.LastTime
	self.GameState = _game.GameState
	self.GameType = _game.GameType
	//添加保存game记录
	self.GameConfig = _game.Config
}

func (self *DGGameJsonSerializer) UnmarshaDG(_game *meta2.GameMetaDG) {

	_game.ReplayRecord = self.ReplayRecord
	fmt.Println(_game.ReplayRecord)
	// for k, _ := range _game.RepertoryCard {
	// 	if k >= len(self.RepertoryCard) {
	// 		break
	// 	}
	// 	_game.RepertoryCard[k] = self.RepertoryCard[k]
	// }
	_game.ReWriteRec = self.ReWriteRec
	_game.Banker = self.Banker
	_game.Nextbanker = self.NextBanker
	_game.BankParter = self.BankParter
	_game.Whoplay = self.CurrentUser
	_game.Player1 = self.Player1
	_game.Player2 = self.Player2
	_game.WhoLastOut = self.WhoLastOut
	_game.TrustPlayer = self.TrustPlayer

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.WhoJiaoOrQiang[i] = self.WhoJiaoOrQiang[i]
		_game.WhoReady[i] = self.WhoReady[i]
		_game.WhoRobSpring[i] = self.WhoRobSpring[i]
		_game.TuoGuanPlayer[i] = self.TuoGuanPlayer[i]
		_game.TrustCounts[i] = self.TrustCounts[i]
		_game.AutoCardCounts[i] = self.AutoCardCounts[i]
		_game.WhoBreak[i] = self.WhoBreak[i]
		_game.BreakCounts[i] = self.BreakCounts[i]
		_game.WhoPass[i] = self.WhoPass[i]
		_game.BTeamOut[i] = self.BTeamOut[i]
		_game.WhoAllOutted[i] = self.WhoAllOutted[i]
		_game.PlayerTurn[i] = self.PlayerTurn[i]
		for j := 0; j < static.MAX_CARD; j++ {
			_game.PlayerCards[i][j] = self.PlayerCards[i][j]
			_game.AllPaiOut[i][j] = self.AllPaiOut[i][j]
			_game.LastPaiOut[i][j] = self.LastPaiOut[i][j]
			_game.BombSplitCards[i][j] = self.BombSplitCards[i][j]
		}
		_game.ThePaiCount[i] = self.ThePaiCount[i]
		_game.XuanPiao[i] = self.XuanPiao[i]
	}
	for i := 0; i < static.ALL_CARD; i++ {
		_game.AllCards[i] = self.AllCards[i]
	}
	_game.LastCards = self.LastCards
	_game.LastOutType = self.LastOutType
	_game.LastOutTypeClient = self.LastOutTypeClient
	_game.AllOutCnt = self.AllOutCnt

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.LastScore[i] = self.LastScore[i]
		_game.Total[i] = self.Total[i]
		_game.Playerrich[i] = self.Playerrich[i]
		_game.PlayerCardScore[i] = self.PlayerCardScore[i]
		_game.XiScore[i] = self.XiScore[i]
		_game.PlayerPunishScore[i] = self.PlayerPunishScore[i]
	}
	_game.CardScore = self.CardScore
	_game.AutoOutTime = self.AutoOutTime
	_game.TimeStart = self.TimeStart
	_game.PlayTime = self.PlayTime
	_game.RoarTime = self.RoarTime
	_game.PowerStartTime = self.PowerStartTime
	_game.Qiang = self.Qiang
	_game.IBase = self.Base
	_game.Spay = self.Spay
	_game.SerPay = self.SerPay
	_game.FaOfTao = self.FaOfTao
	_game.JiangOfTao = self.JiangOfTao
	_game.AddSpecailBeishu = self.AddSpecailBeishu
	_game.FristOut = self.FristOut
	_game.BiYa = self.BiYa
	_game.TuoGuan = self.TuoGuan
	_game.ZhaNiao = self.ZhaNiao
	_game.FourTake3 = self.FourTake3
	_game.BombSplit = self.BombSplit
	_game.QuickPass = self.QuickPass
	_game.SplitCards = self.SplitCards
	_game.Bomb3Ace = self.Bomb3Ace
	_game.LessTake = self.LessTake
	_game.TeamCard = self.TeamCard
	_game.Jiao2King = self.Jiao2King
	_game.KeFan = self.KeFan
	_game.FourTake2 = self.FourTake2
	_game.FourTake1 = self.FourTake1
	_game.CardNum = self.CardNum
	_game.KingLai = self.KingLai
	_game.Big510k = self.Big510k
	_game.FullScoreAward = self.FullScoreAward
	_game.FourKingScore = self.FourKingScore
	_game.AddDiFen = self.AddDiFen
	_game.ShowHandCardCnt = self.ShowHandCardCnt
	_game.GetLastScore = self.GetLastScore
	_game.SeeTeamerCard = self.SeeTeamerCard
	_game.BombMode = self.BombMode
	_game.Restart = self.Restart
	_game.Piao = self.Piao
	_game.PiaoCount = self.PiaoCount
	_game.NotDismiss = self.NotDismiss
	_game.FristOutMode = self.FristOutMode
	_game.NoBomb = self.NoBomb
	_game.BombRealTime = self.BombRealTime
	_game.OutCardDismissTime = self.OutCardDismissTime
	_game.LessTakeFirst = self.LessTakeFirst
	_game.LessTakeNext = self.LessTakeNext
	_game.FakeKing = self.FakeKing
	_game.QiangChun = self.QiangChun
	_game.Hard510KMode = self.Hard510KMode
	_game.SeePartnerCards = self.SeePartnerCards
	_game.IsRed3First = self.IsRed3First
	_game.ZhaNiaoFen = self.ZhaNiaoFen
	_game.FengDing = self.FengDing
	_game.OnlyAuto = self.OnlyAuto
	_game.EndReadyCheck = self.EndReadyCheck
	_game.TrustLimit = self.TrustLimit
	_game.TrustJuShu = self.TrustJuShu
	_game.TimeOutPunish = self.TimeOutPunish
	_game.PunishScore = self.PunishScore

	_game.BMingJiFlag = self.MingJiFlag
	_game.DownPai = self.DownPai
	_game.RoarPai = self.RoarPai
	_game.WhoRoar = self.WhoRoar
	_game.WhoHasKingBomb = self.WhoHasKingBomb
	_game.JiabeiType = self.JiabeiType
	_game.JiaoPaiMate = self.JiaoPaiMate
	_game.BSplited = self.BSplited
	_game.AddXiScore = self.AddXiScore
	_game.RestartCount = self.RestartCount
	_game.WhoRob = self.WhoRob
	_game.RobSpringFlag = self.RobSpringFlag
	_game.OutScorePai = self.OutScorePai
	_game.MinCard = self.MinCard
	_game.WhoAnti = self.WhoAnti
	_game.WhoAntic = self.WhoAntic
	_game.CurTrustJuShu = self.CurTrustJuShu
	_game.XiScoreMode = self.XiScoreMode

	_game.WhoHas4KingScore = self.WhoHas4KingScore
	_game.WhoHas4KingPower = self.WhoHas4KingPower

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		_game.DownPai3P[i] = self.DownPai3P[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.HasKingNum[i] = self.HasKingNum[i]
		_game.Who8Xi[i] = self.Who8Xi[i]
		_game.Who7Xi[i] = self.Who7Xi[i]
		_game.WhoSame510K[i] = self.WhoSame510K[i]
		_game.MaxScore[i] = self.MaxScore[i]
		_game.TotalFirstTurn[i] = self.TotalFirstTurn[i]
		_game.TotalDuPai[i] = self.TotalDuPai[i]
		_game.WhoKingCount[i] = self.WhoKingCount[i]
		_game.Who510kCount[i] = self.Who510kCount[i]
		_game.WhoTotal8Xi[i] = self.WhoTotal8Xi[i]
		_game.WhoTotal7Xi[i] = self.WhoTotal7Xi[i]
		_game.WhoTotalGonglong[i] = self.WhoTotalGonglong[i]
		_game.WhoGonglongCount[i] = self.WhoGonglongCount[i]
		_game.WhoToTalMore4KingCount[i] = self.WhoToTalMore4KingCount[i]
		_game.WhoGonglongScore[i] = self.WhoGonglongScore[i]
		_game.TotalAnti[i] = self.TotalAnti[i]
		_game.PlayKingBomb[i] = self.PlayKingBomb[i]
		_game.Play8Xi[i] = self.Play8Xi[i]
		_game.Play7Xi[i] = self.Play7Xi[i]
		_game.Play510K[i] = self.Play510K[i]
		_game.WhoBombScore[i] = self.WhoBombScore[i]
		_game.WhoOutCount[i] = self.WhoOutCount[i]
		_game.WhoToTalChuntianCount[i] = self.WhoToTalChuntianCount[i]
		_game.BombCount[i] = self.BombCount[i]
		_game.MaxBombCount[i] = self.MaxBombCount[i]
		_game.Bird[i] = self.Bird[i]
		_game.ValidBombCount[i] = self.ValidBombCount[i]
		_game.Play5Xi[i] = self.Play5Xi[i]
		_game.Play6Xi[i] = self.Play6Xi[i]
		_game.PlaySame510k[i] = self.PlaySame510k[i]
		_game.PlayKingBomb2[i] = self.PlayKingBomb2[i]
		_game.MaxBeiShu[i] = self.MaxBeiShu[i]
		_game.MaxBird[i] = self.MaxBird[i]
		_game.MaxSpring[i] = self.MaxSpring[i]
	}

	_game.VecGameData = self.VecGameData
	_game.VecGameEnd = self.VecGameEnd
	_game.TimeGameRecord = self.TimeGameRecord
	_game.NameGameRecord = self.NameGameRecord
	_game.GameRecordNum = self.GameRecordNum
	_game.LastTime = self.LastTime
	_game.GameState = self.GameState
	_game.GameType = self.GameType

	//_game.Config             =        self.GameConfig
}

//获取累加的计算
func (self *DGGameJsonSerializer) getString() string {
	var str string
	return str
}
