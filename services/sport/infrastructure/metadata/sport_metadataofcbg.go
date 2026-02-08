package metadata

import (
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	"time"
)

//基础定义
const ()

//游戏流程数据
type GameMetaDG_CB struct {
	//游戏变量
	GameState     int //游戏状态 GameState，1表示吼牌阶段，2表示打牌阶段,3选漂阶段
	GameType      int //游戏类型 1表示普通模式2vs2，2表示吼牌模式1vs3
	PowerDownTime int // 权限停止时间
	TheActionStep int // 倒计时
	NowActionStep int

	ReplayRecord DG_Replay_Record //回放记录
	ReWriteRec   byte             //是否重复写回放数据，每小局游戏开始时清理

	//玩家信息
	Banker     uint16           // 庄家
	NextBanker uint16           //下一个庄家
	BankParter uint16           //庄家的队友，2vs2模式下才有
	WhoPlay    uint16           //当前玩家	0--MAXPLAYER-1
	Player1    uint16           //3人拱中，除庄家外的另一个人
	Player2    uint16           //3人拱中，除庄家外的另一个人
	WhoLastOut uint16           //上一个出牌玩家，pass的不算
	WhoReady   [MAX_PLAYER]bool // 谁已经完成了吼牌过程
	AlermState [MAX_PLAYER]byte //每个人的报警状态 0没报警（包括出完牌） 1单报 2双报

	//托管和离线数据
	TuoGuanPlayer  [MAX_PLAYER]bool //谁托管了？
	TrustCounts    [MAX_PLAYER]byte //玩家托管次数
	AutoCardCounts [MAX_PLAYER]byte //自动出牌的次数
	WhoBreak       [MAX_PLAYER]bool //谁断线了？
	BreakCounts    [MAX_PLAYER]byte // 断线次数
	TrustPlayer    []byte           //托管玩家
	TrustOrder     [MAX_PLAYER]byte //玩家托管顺序

	//牌数据
	AllCards_CYDG [info2.CYDG_CARDS]byte            // 所有牌 崇阳打滚使用的
	AllCards      [static.ALL_CARD]byte             // 所有牌
	PlayerCards   [MAX_PLAYER][static.MAX_CARD]byte // 玩家分到的牌

	AllPaiOut         [MAX_PLAYER][static.MAX_CARD]byte // 出的牌
	LastPaiOut        [MAX_PLAYER][static.MAX_CARD]byte // 上一轮出的牌
	ThePaiCount       [MAX_PLAYER]byte                  // 牌数
	LastOutType       int                               // 最后出牌的类型
	LastOutTypeClient int                               // 最后出牌的类型,客户端端的类型枚举可能和服务器不一致，如果一致就不需要这个了
	AllOutCnt         byte                              // 几人已经出完了
	WhoPass           [MAX_PLAYER]bool                  // 谁放弃了（过）
	TeamOut           [MAX_PLAYER]bool                  // 我的对家是否走完了？
	WhoAllOutted      [MAX_PLAYER]bool                  // 谁出完了，
	PlayerTurn        [MAX_PLAYER]byte                  // 玩家出完牌的顺序，1游、2游、3游
	OutScorePai       [24]byte                          //所出分牌

	OutCardSequence_CYDG [info2.CYDG_CARDS]byte //记住所有玩家的出牌数据,发牌拱需要使用,崇阳打滚使用
	OutCardSequence      [static.ALL_CARD]byte  //记住所有玩家的出牌数据,发牌拱需要使用
	OutSequenceIndes     byte                   //记住所有玩家的出牌数据的索引,发牌拱需要使用
	CarsLeftNum          [16]byte               //当前所有牌剩余数目，0号位是花牌，1号位是大王，2号位是小王，3号位是2，逐渐减少，一直到最小的3，共16种牌
	BombErrStart         uint16                 //炸错发起者
	BombErrBankerCount   int                    //跟庄炸庄炸错次数

	//积分数据
	LastScore       [MAX_PLAYER]int      // 上一小局输赢金币
	Total           [MAX_PLAYER]int      // 总输赢，若干小局相加的金币
	PlayeRich       [MAX_PLAYER]int      // 玩家的财富金币
	CardScore       int                  // 每一轮牌的分
	PlayerCardScore [MAX_PLAYER]int      // 每个人最后的分数，有可能没出完的也加进去了
	XiScore         [MAX_PLAYER]int      // 喜钱
	ZongZhaScore    [MAX_PLAYER]int      // 总炸分数
	PiaoScore       [MAX_PLAYER]int      // 漂分
	FaScore         [MAX_PLAYER][]uint16 // 炸错罚分
	HuapaiScore     [MAX_PLAYER]int      // 花牌分s
	ExtAddNum       [MAX_PLAYER][2]int   // 额外加分数据，下标0表示王的数目，下标1表示花牌数目

	//时间数据
	AutoOutTime    int   // 自动出牌时间，托管的时候出牌时间
	TimeStart      int   // 坐下之后，按开始的等待时间
	PlayTime       int   // 出牌时间
	RoarTime       int   // 吼牌时间
	PowerStartTime int64 //时间辅助变量
	//底
	Base             int  // 底
	Pay              int  // 服务费
	SerPay           int  // 茶水费千分比
	FaOfTao          int  // 逃跑的惩罚倍数
	JiangOfTao       int  // 逃跑的补偿倍数
	AddSpecailBeishu int  // 别人有特殊牌奖励时，需要扣除分的倍数
	HasPiao          bool // 是否选漂
	PiaoType         int  // 漂类型(上面的bool没法对应多个，重新定义一个)
	HasBombStr       bool // 是否有连炸
	HasBombErr       bool // 是否炸错罚分
	SkyCnt           int  // 花牌数目
	ZhuangBuJie      bool //庄家是否不接风
	TrusteeCost      bool //托管玩家承担所有输分
	Hard510KOf4IsXi  bool //四硬510K是否算喜
	ExtAdd           bool //额外加分，比如打出3王1花得1倍底分，跟打出8喜7喜的加分不一样
	TimeOutPunish    bool //超时罚分
	AddXiScore       bool //是否带喜分
	Endready         bool //小结算是否自动准备
	XiFenHaveOrOut   bool //细分是起到就算还是打出算
	ZongZhaFlag      bool //是否需要在小结算计算总炸分
	No6Xi            bool //6喜是否算喜分，勾选为true表示无6喜、不算喜分

	CurPeriod       int     // 超时罚分的当前阶段
	PunishStartTime int64   // 超时罚分的度过时间
	HasAction       [4]bool // 超时罚分的动作标志
	PunishCount     [4]int  // 超时罚分的次数计数

	//好友房
	LongzhaDing   int  // 笼炸封顶
	FakeKingValue byte // 王单出算几
	FapaiMode     int  // 发牌模式，，0表示起牌拱，1表示发牌拱，2表示疯狂拱，3表示变态拱
	ShowCardNum   int  // 是否显示手牌数目
	MaxKingNum    byte // 4王或8王
	CalScoreMode  int  //0 累计计分 1 不累计计分

	//叫牌相关数据
	MingJiFlag bool                         // 是否已经显示明鸡了
	DownPai    byte                         // 4人拱做牌时的叫牌
	DownPai3P  [static.MAX_DOWNCARDNUM]byte // 3人拱的底牌，做牌也用这个数组
	RoarPai    byte                         // 叫的牌
	WhoRoar    uint16                       // 谁吼牌了？

	HaveFinish bool // 是否已经完成了某个任务了

	HasKingNum     [MAX_PLAYER]int // 每个人手牌中王的个数
	WhoHasKingBomb uint16          // 谁有天炸
	Who8Xi         [MAX_PLAYER]int // 每个人的8喜个数
	Who7Xi         [MAX_PLAYER]int // 每个人的7喜个数
	Who6Xi         [MAX_PLAYER]int // 每个人的6喜个数
	WhoSame510K    [MAX_PLAYER]int // 每个人的同色510K的个数，如1玩家共有5个510k，其中方块的2个，黑桃的2个，红桃的1个，则有两个同色510k（方块，黑桃）

	MaxScore               [MAX_PLAYER]int // 最高抓分
	TotalFirstTurn         [MAX_PLAYER]int // 一游次数
	TotalDuPai             [MAX_PLAYER]int // 独牌次数
	WhoKingCount           [MAX_PLAYER]int // 每个人的4及以上个王的个数
	Who510kCount           [MAX_PLAYER]int // 每个人的4及以上个硬510k的个数
	WhoTotal8Xi            [MAX_PLAYER]int // 每个人大局的8喜个数
	WhoTotal7Xi            [MAX_PLAYER]int // 每个人大局的7喜个数
	WhoTotal6Xi            [MAX_PLAYER]int // 每个人大局的6喜个数
	WhoTotalGonglong       [MAX_PLAYER]int // 每个人大局的笼炸的笼数(大局)
	WhoGonglongCount       [MAX_PLAYER]int // 每个人的笼炸的个数(每小局)
	WhoToTalMore4KingCount [MAX_PLAYER]int // 每个人的4及以上个王的次数
	WhoGonglongScore       [MAX_PLAYER]int // 每个人的笼炸的分数(每小局)
	WhoBombScore           [MAX_PLAYER]int // 每个人的炸弹分(每小局)
	WhoOutCount            [MAX_PLAYER]int // 每个人的出牌次数(每小局)
	WhoToTalChuntianCount  [MAX_PLAYER]int // 每个人的春天次数,包括反春

	PlayKingBomb [MAX_PLAYER]int // 打出的天炸，4王一起打出
	Play6Xi      [MAX_PLAYER]int // 打出的6喜个数
	Play8Xi      [MAX_PLAYER]int // 打出的8喜个数
	Play7Xi      [MAX_PLAYER]int // 打出的7喜个数
	Play510K     [MAX_PLAYER]int // 打出的510K的个数
	PlayHard510K [MAX_PLAYER]int // 打出的硬510K的个数

	VecGameEnd []static.Msg_S_DG_GameEnd //记录每一局的结果

	//组件变量
	TimeGameRecord string            //记录当前日志的日期
	NameGameRecord string            //记录文件名
	GameRecordNum  string            //记录编号
	LastTime       int64             // 用于断线重入 校准客户端操作时间
	Config         static.GameConfig //游戏配置

	fFCB         bool //第一局流局是不是要换庄家 fristFlowChangeBanker  公共
	IsLocalDebug bool
	BHaveZongZha bool //是否有总炸
}

func (self *GameMetaDG_CB) OnScoreOffset(total []float64) {
	for i, t := range total {
		var order DG_Replay_Order
		order.R_ChairId = uint16(i)
		order.R_Opt = info2.E_GameScore
		order.UserScorePL = t
		self.ReplayRecord.R_Orders = append(self.ReplayRecord.R_Orders, order)
	}
}

//如果有些变量的值在每小局都要保留，建议使用resetForNext
func (self *GameMetaDG_CB) ResetForNext() {
	//游戏变量
	self.GameState = GsNull
	self.GameType = GT_NULL
	self.PowerDownTime = 0
	self.TheActionStep = AS_NULL
	self.NowActionStep = AS_NULL

	self.ReplayRecord.ReSet()

	//玩家信息
	self.Banker = static.INVALID_CHAIR
	self.BankParter = static.INVALID_CHAIR
	self.WhoPlay = static.INVALID_CHAIR
	self.Player1 = static.INVALID_CHAIR
	self.Player2 = static.INVALID_CHAIR
	self.WhoLastOut = static.INVALID_CHAIR
	self.WhoReady = [MAX_PLAYER]bool{}

	//托管和离线数据
	self.TuoGuanPlayer = [MAX_PLAYER]bool{}
	self.TrustCounts = [MAX_PLAYER]byte{}
	self.AutoCardCounts = [MAX_PLAYER]byte{}
	self.WhoBreak = [MAX_PLAYER]bool{}
	self.BreakCounts = [MAX_PLAYER]byte{}
	self.TrustOrder = [MAX_PLAYER]byte{}
	self.TrustPlayer = []byte{}
	self.AlermState = [MAX_PLAYER]byte{}

	//牌数据
	//self.AllCards = [public.ALL_CARD]byte{}
	for j, _ := range self.PlayerCards {
		for k, _ := range self.PlayerCards[j] {
			self.PlayerCards[j][k] = 0
		}
	}
	for j, _ := range self.AllPaiOut {
		for k, _ := range self.AllPaiOut[j] {
			self.AllPaiOut[j][k] = 0
		}
	}
	for j, _ := range self.LastPaiOut {
		for k, _ := range self.LastPaiOut[j] {
			self.LastPaiOut[j][k] = 0
		}
	}
	self.ThePaiCount = [MAX_PLAYER]byte{}
	self.LastOutType = 0
	self.LastOutTypeClient = 0
	self.AllOutCnt = 0
	self.WhoPass = [MAX_PLAYER]bool{}
	self.TeamOut = [MAX_PLAYER]bool{}
	self.WhoAllOutted = [MAX_PLAYER]bool{}
	self.OutScorePai = [24]byte{}
	for j, _ := range self.PlayerTurn {
		self.PlayerTurn[j] = static.INVALID_BYTE
	}
	for j, _ := range self.DownPai3P {
		self.DownPai3P[j] = 0
	}
	self.CarsLeftNum = [16]byte{}
	self.BombErrStart = static.INVALID_CHAIR
	self.BombErrBankerCount = 0
	//积分数据
	self.PlayeRich = [MAX_PLAYER]int{}
	self.CardScore = 0
	self.PlayerCardScore = [MAX_PLAYER]int{}
	self.XiScore = [MAX_PLAYER]int{}
	self.ZongZhaScore = [MAX_PLAYER]int{}
	self.PiaoScore = [MAX_PLAYER]int{}
	self.FaScore = [MAX_PLAYER][]uint16{}
	self.HuapaiScore = [MAX_PLAYER]int{}
	self.ExtAddNum = [MAX_PLAYER][2]int{}

	//时间数据
	self.PowerStartTime = time.Now().Unix()

	//叫牌相关数据
	self.MingJiFlag = false
	self.DownPai = 0
	self.RoarPai = 0
	self.WhoRoar = static.INVALID_CHAIR

	self.HaveFinish = false

	self.WhoHasKingBomb = static.INVALID_CHAIR
	self.Who8Xi = [MAX_PLAYER]int{}
	self.HasKingNum = [MAX_PLAYER]int{}
	self.Who7Xi = [MAX_PLAYER]int{}
	self.Who6Xi = [MAX_PLAYER]int{}
	self.WhoSame510K = [MAX_PLAYER]int{}
	self.PlayKingBomb = [MAX_PLAYER]int{}
	self.Play6Xi = [MAX_PLAYER]int{}
	self.Play8Xi = [MAX_PLAYER]int{}
	self.Play7Xi = [MAX_PLAYER]int{}
	self.Play510K = [MAX_PLAYER]int{}
	self.PlayHard510K = [MAX_PLAYER]int{}
	//

	//状态变量
	self.LastTime = 0
	self.IsLocalDebug = false

	self.WhoGonglongCount = [MAX_PLAYER]int{}
	self.WhoKingCount = [MAX_PLAYER]int{}
	self.Who510kCount = [MAX_PLAYER]int{}
	self.WhoGonglongCount = [MAX_PLAYER]int{}
	self.WhoGonglongScore = [MAX_PLAYER]int{}
	self.WhoBombScore = [MAX_PLAYER]int{}
	self.WhoOutCount = [MAX_PLAYER]int{}
	self.ReplayRecord.UVitamin = make(map[int64]float64)
}

//Reset all
func (self *GameMetaDG_CB) Reset() {
	self.ResetForNext()

	self.AddSpecailBeishu = 0
	self.LongzhaDing = 32
	self.FakeKingValue = 2
	self.FapaiMode = 0
	self.ShowCardNum = 0
	self.MaxKingNum = 4

	self.OutCardSequence = [static.ALL_CARD]byte{}
	self.OutSequenceIndes = 0
	self.NextBanker = static.INVALID_CHAIR
	self.AllCards = [static.ALL_CARD]byte{}
	self.AllCards_CYDG = [info2.CYDG_CARDS]byte{}
	self.LastScore = [MAX_PLAYER]int{}
	self.Total = [MAX_PLAYER]int{}

	self.MaxScore = [MAX_PLAYER]int{}
	self.TotalFirstTurn = [MAX_PLAYER]int{}
	self.TotalDuPai = [MAX_PLAYER]int{}
	self.WhoTotal8Xi = [MAX_PLAYER]int{}
	self.WhoTotal7Xi = [MAX_PLAYER]int{}
	self.WhoTotalGonglong = [MAX_PLAYER]int{}
	self.WhoToTalMore4KingCount = [MAX_PLAYER]int{}
	self.WhoToTalChuntianCount = [MAX_PLAYER]int{}
	self.XiScore = [MAX_PLAYER]int{}      //喜钱
	self.ZongZhaScore = [MAX_PLAYER]int{} //总炸
	self.VecGameEnd = []static.Msg_S_DG_GameEnd{}
	self.ReWriteRec = 0

	//时间数据
	self.AutoOutTime = 3
	self.TimeStart = 0
	self.PlayTime = static.GAME_OPERATION_TIME_15 * 2
	self.RoarTime = static.GAME_OPERATION_TIME_15 * 2

	//底
	self.Base = 1
	self.Pay = 0
	self.SerPay = 0
	self.FaOfTao = 0
	self.JiangOfTao = 0
	self.HasPiao = false
	self.PiaoType = 0
	self.HasBombStr = false
	self.HasBombErr = false
	self.ZhuangBuJie = false
	self.SkyCnt = 0
	self.TrusteeCost = false
	self.ExtAdd = false
	self.CalScoreMode = 0
	self.No6Xi = false
}
