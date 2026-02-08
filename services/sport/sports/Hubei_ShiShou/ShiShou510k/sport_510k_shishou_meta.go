package ShiShou510k

import (
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	"time"
)

//基础定义
const (
	MAX_PLAYER = 4 //游戏人数
)

/*
 * 回放相关结构体
 */
const (
	DG_REPLAY_OPT_HOUPAI        = 0  //回放吼牌操作
	DG_REPLAY_OPT_END_HOUPAI    = 1  //回放结束吼牌
	DG_REPLAY_OPT_OUTCARD       = 2  //回放出牌
	DG_REPLAY_OPT_END_GAME      = 3  //回放结束游戏
	DG_REPLAY_OPT_DIS_GAME      = 4  //回放解散游戏
	DG_REPLAY_OPT_TURN_OVER     = 5  //回放本轮结束
	DG_REPLAY_OPT_QIANG         = 6  //回放抢庄操作
	DG_REPLAY_OPT_PIAO          = 7  //漂
	DG_REPLAY_OPT_TUOGUAN       = 8  //回放托管
	DG_REPLAY_OPT_4KINGSCORE    = 9  //回放4王换分
	DG_REPLAY_OPT_RESTART       = 10 //回放重新发牌
	DG_REPLAY_OPT_ANTIBRAND     = 11 //回放反牌操作
	DG_REPLAY_OPT_END_ANTIBRAND = 12 //回放结束反牌

	DG_EXT_HOUPAI        = 0  //吼牌
	DG_EXT_MINGJI        = 1  //明鸡
	DG_EXT_TURNSCORE     = 2  //本轮分
	DG_EXT_GETSCORE      = 3  //抓分
	DG_EXT_GONGLNUM      = 4  //拱笼数
	DG_EXT_QIANG         = 5  //抢庄
	DG_EXT_ENDQIANG      = 6  //抢庄结束，抢庄类型
	DG_EXT_PIAO          = 7  //漂
	DG_EXT_CARDTYPE      = 8  //牌类型
	DG_EXT_TUOGUAN       = 9  //托管
	DG_EXT_4KINGSCORE    = 10 //回放4王换分
	DG_EXT_RESTART       = 11 //回放重新发牌
	DG_EXT_MAGICCARD     = 12 //回放加癞子牌
	DG_EXT_CURBEISHU     = 13 //回放当前倍数
	DG_EXT_ANTIBRAND     = 14 //反牌
	DG_EXT_ENE_ANTIBRAND = 15 //反牌结束
	DG_EXT_CURSCORE      = 16 //玩家当前当局分
	DG_EXT_TOTALSCORE    = 17 //玩家当前总分

)

type SS510k_Replay_Order_Ext struct {
	Ext_type  int             `json:"exttype"`  //玩家 //操作类型
	Ext_value int             `json:"extvalue"` //操作值
	Ext_score [MAX_PLAYER]int `json:"extscore"` //操作产生的分值
}

type SS510k_Replay_Order struct {
	R_ChairId   uint16                    `json:"id"`         //玩家
	R_Opt       int                       `json:"operation"`  //记录类型
	R_Value     []int                     `json:"value"`      //出牌
	R_Opt_Ext   []SS510k_Replay_Order_Ext `json:"orderext"`   //出牌
	UserScorePL float64                   `json:"user_score"` // 玩家疲劳值，目前跟分数一致
	R_ScoreCard []int                     `json:"scorecard"`  //分牌
}

func (self *SS510k_Replay_Order) AddReplayExtData(exttype int, extvalue int) {
	var ext_data SS510k_Replay_Order_Ext
	ext_data.Ext_type = exttype
	ext_data.Ext_value = extvalue

	self.R_Opt_Ext = append(self.R_Opt_Ext, ext_data)
}

type SS510k_Replay_Record struct {
	R_HandCards [MAX_PLAYER][]int        `json:"handcard"`  //用户最初手上的牌
	R_Orders    []SS510k_Replay_Order    `json:"orders"`    //用户操作
	R_Score     [MAX_PLAYER]int          `json:"score"`     //游戏积分
	UVitamin    map[int64]float64        `json:"u_vitamin"` // 玩家起始疲劳值
	EndInfo     *static.Msg_S_DG_GameEnd `json:"end_inf"`
}

func (self *SS510k_Replay_Record) ReSet() {
	for i := 0; i < MAX_PLAYER; i++ {
		self.R_HandCards[i] = []int{}
	}
	self.UVitamin = make(map[int64]float64)
	self.R_Orders = []SS510k_Replay_Order{}
	self.R_Score = [MAX_PLAYER]int{}
	self.EndInfo = nil
}

//GameState
const (
	GsNull           = iota
	GsRoarPai        //吼牌？
	GsPlay           //玩牌
	GsQiang          // 抢庄
	GsXuanPiao       // 选漂
	GsSplitCards     // 切牌
	GsStartAnimation // 开始动画播放需要时间
	GsQuickPass      // 快速国牌需要1秒的时间，否则看不到首出的牌
	Gs4KingScore     // 4王换分阶段
	GsBombCheck      // 炸弹检测阶段
	GsRestart        // 重新发牌阶段
	GsAntiBrand      //反牌
)

//ActionStep
const (
	AS_NULL    = iota
	AS_ROAR    // 硬牌
	AS_ENDROAR // 结束硬牌
	AS_PLAY    // 打牌
	AS_ENDPLAY // 打牌结束
	AS_ENDGAME // 结束游戏
	AS_COUNT
)

//RoarState
const (
	H_NULL    = iota
	H_Jiao    // 叫庄
	H_Qiang   // 抢庄
	H_JiaoQi  // 放弃了叫庄
	H_QiangQi // 放弃了抢庄
)

//TGameType
const (
	GT_NULL   = iota
	GT_NORMAL //普通模式，一般为找朋友，就是2vs2,这个为取分数
	GT_ROAR   //吼牌模式，一般为1vs3,这个为争上游
)

//GameOverType
const (
	GOT_NULL       = iota
	GOT_ESCAPE     //有人逃跑，游戏结束
	GOT_NORMAL     //正常的结束游戏，即：在找朋友的模式下，三家走完，在吼牌模式下，一家走完
	GOT_ZHONGTU    //中途结束游戏，即满足一定的条件，结束了游戏
	GOT_DOUBLEKILL //“绑”，意为双杀，在找朋友的模式下，一方为一游和二游走完
	GOT_TUOGUAN    // 达到托管次数限制后结束游戏
	GOT_DISMISS    //解散
)

//游戏流程数据
type SportMetaSS510K struct {
	//游戏变量
	GameState     int //游戏状态 GameState，1表示吼牌阶段，2表示打牌阶段
	GameType      int //游戏类型 1表示普通模式2vs2，2表示吼牌模式1vs3
	DownTime      int // 权限停止时间
	TheActionStep int // 倒计时
	NowActionStep int

	ReplayRecord SS510k_Replay_Record //回放记录
	ReWriteRec   byte                 //是否重复写回放数据，每小局游戏开始时清理

	//玩家信息
	Banker         uint16           // 庄家
	Nextbanker     uint16           //下一个庄家
	BankParter     uint16           //庄家的队友，2vs2模式下才有
	Whoplay        uint16           //当前玩家	0--MAXPLAYER-1
	Player1        uint16           //3人拱中，除庄家外的另一个人
	Player2        uint16           //3人拱中，除庄家外的另一个人
	WhoLastOut     uint16           //上一个出牌玩家，pass的不算
	WhoReady       [MAX_PLAYER]bool // 谁已经完成了吼牌过程
	WhoAntic       [MAX_PLAYER]int  // 选择反牌的情况（-1：没选择；0：弃；1：选择）
	WhoJiaoOrQiang [MAX_PLAYER]int  // 不叫和不抢需要区分
	JiabeiType     int              // 翻倍类型:0,不翻倍；1，一王叫庄；2，一王抢庄,3，其他抢庄,4，无王叫庄

	//托管和离线数据
	TuoGuanPlayer  [MAX_PLAYER]bool //谁托管了？
	TrustCounts    [MAX_PLAYER]byte //玩家托管次数
	AutoCardCounts [MAX_PLAYER]byte //自动出牌的次数
	WhoBreak       [MAX_PLAYER]bool //谁断线了？
	BreakCounts    [MAX_PLAYER]byte // 断线次数
	TrustPlayer    []byte           //托管玩家
	TrustOrder     [MAX_PLAYER]byte //玩家托管顺序

	//牌数据
	AllCards       [static.ALL_CARD]byte             // 所有牌
	PlayerCards    [MAX_PLAYER][static.MAX_CARD]byte // 玩家分到的牌
	BombSplitCards [MAX_PLAYER][static.MAX_CARD]byte // 拆掉炸弹的牌

	AllPaiOut         [MAX_PLAYER][static.MAX_CARD]byte // 出的牌
	LastPaiOut        [MAX_PLAYER][static.MAX_CARD]byte // 上一轮出的牌
	ThePaiCount       [MAX_PLAYER]byte                  // 牌数
	LastOutType       int                               // 最后出牌的类型
	LastOutTypeClient int                               // 最后出牌的类型,客户端端的类型枚举可能和服务器不一致，如果一致就不需要这个了
	AllOutCnt         byte                              // 几人已经出完了
	WhoPass           [MAX_PLAYER]bool                  // 谁放弃了（过）
	BTeamOut          [MAX_PLAYER]bool                  // 我的对家是否走完了？
	WhoAllOutted      [MAX_PLAYER]bool                  // 谁出完了，
	PlayerTurn        [MAX_PLAYER]byte                  // 玩家出完牌的顺序，1游、2游、3游
	//积分数据
	LastScore       [MAX_PLAYER]int // 上一小局输赢金币
	Total           [MAX_PLAYER]int // 总输赢，若干小局相加的金币
	Playerrich      [MAX_PLAYER]int // 玩家的财富金币
	CardScore       int             // 每一轮牌的分
	PlayerCardScore [MAX_PLAYER]int // 每个人最后的分数，有可能没出完的也加进去了
	XiScore         [MAX_PLAYER]int // 喜钱
	//时间数据
	AutoOutTime    int   // 自动出牌时间，托管的时候出牌时间
	TimeStart      int   // 坐下之后，按开始的等待时间
	PlayTime       int   // 出牌时间
	RoarTime       int   // 吼牌时间
	PowerStartTime int64 //时间辅助变量
	//底
	Qiang            int             // 是否可以抢庄，1表示可以抢
	IBase            int             // 底
	Spay             int             // 服务费
	SerPay           int             // 茶水费千分比
	FaOfTao          int             // 逃跑的惩罚倍数
	JiangOfTao       int             // 逃跑的补偿倍数
	AddSpecailBeishu int             // 别人有特殊牌奖励时，需要扣除分的倍数
	Jiao2King        int             // 双王必须叫牌
	TeamCard         int             // 先跑可看队友牌
	AddXiScore       bool            //是否带喜分
	KingLai          int             // 王可否做癞子
	Big510k          bool            //6炸7炸不可打510k
	Piao             int             //选飘
	PiaoCount        int             //选飘次数，0每局飘一次，1首局飘一次
	XuanPiao         [MAX_PLAYER]int // 飘分

	FristOut        int  //首出类型 ,0黑桃3先出，1庄家首出
	BiYa            int  //有大比压 ,0有大必压，1可以不压，
	TuoGuan         int  //托管 ,0不托管，大于0托管，
	ZhaNiao         bool //是否扎鸟，红桃10是鸟
	FourTake3       bool //是否可以4带3
	FourTake2       bool //是否可以4带2
	FourTake1       bool //4带1是否算炸弹
	BombSplit       bool //炸弹是否可拆，true表示不可以拆
	QuickPass       bool //是否快速过牌
	SplitCards      bool //是否有切牌
	Bomb3Ace        bool //3个A是否是炸弹
	LessTake        bool //最后一手是否可以少带
	KeFan           bool //是否可以反春
	CardNum         int  //牌数
	FullScoreAward  int  //满分奖励
	FourKingScore   int  //4王换分
	AddDiFen        int  //额外加的底分
	ShowHandCardCnt bool //是否显示手牌数
	GetLastScore    bool //是否可以捡尾分
	SeeTeamerCard   bool //是否可以看队友牌
	BombMode        bool //炸弹被压无分
	Restart         bool //重新发牌
	NotDismiss      bool //托管不解散
	FristOutMode    int  //首出出牌类型 ,0必带黑三或最小牌，1任意出牌
	NoBomb          bool //不发炸弹
	BombRealTime    bool //炸弹实时计分
	PlayMode        int  //石首510k玩法，0经典 1双赖 2四赖
	RandTeamer      bool //随机队友

	//叫牌相关数据
	BMingJiFlag bool                         // 是否已经显示明鸡了
	DownPai     byte                         // 4人拱做牌时的叫牌
	DownPai3P   [static.MAX_DOWNCARDNUM]byte // 3人拱的底牌，做牌也用这个数组
	RoarPai     byte                         // 叫的牌
	WhoRoar     uint16                       // 谁吼牌了？
	WhoAnti     uint16                       // 谁反牌
	BSplited    bool                         // 是否选择了切牌
	BHaveFinish bool                         // 是否已经完成了某个任务了

	WhoHas4KingScore uint16 // 谁用4王换分了
	WhoHas4KingPower uint16 // 谁拥有4王换分权限（换完就没有该牌权了）

	HasKingNum             [MAX_PLAYER]int // 每个人手牌中王的个数
	WhoHasKingBomb         uint16          // 谁有天炸
	Who8Xi                 [MAX_PLAYER]int // 每个人的8喜个数
	Who7Xi                 [MAX_PLAYER]int // 每个人的7喜个数
	WhoSame510K            [MAX_PLAYER]int // 每个人的同色510K的个数，如1玩家共有5个510k，其中方块的2个，黑桃的2个，红桃的1个，则有两个同色510k（方块，黑桃）
	WinCount               [MAX_PLAYER]int //胜局次数
	MaxScore               [MAX_PLAYER]int // 最高抓分
	TotalFirstTurn         [MAX_PLAYER]int // 一游次数
	TotalDuPai             [MAX_PLAYER]int // 独牌次数
	TotalAnti              [MAX_PLAYER]int // 反牌次数
	WhoKingCount           [MAX_PLAYER]int // 每个人的4及以上个王的个数
	Who510kCount           [MAX_PLAYER]int // 每个人的4及以上个硬510k的个数
	WhoTotal8Xi            [MAX_PLAYER]int // 每个人大局的8喜个数
	WhoTotal7Xi            [MAX_PLAYER]int // 每个人大局的7喜个数
	WhoTotalGonglong       [MAX_PLAYER]int // 每个人大局的笼炸的笼数(大局)
	WhoGonglongCount       [MAX_PLAYER]int // 每个人的笼炸的个数(每小局)
	WhoToTalMore4KingCount [MAX_PLAYER]int // 每个人的4及以上个王的次数
	WhoGonglongScore       [MAX_PLAYER]int // 每个人的笼炸的分数(每小局)
	WhoBombScore           [MAX_PLAYER]int // 每个人的炸弹分(每小局)
	WhoOutCount            [MAX_PLAYER]int // 每个人的出牌次数(每小局)
	WhoToTalChuntianCount  [MAX_PLAYER]int // 每个人的春天次数,包括反春
	WhoToTalplayKingBomb   [MAX_PLAYER]int //每个人的打出的天炸，4王一起打出
	WhoToTalSame510K       [MAX_PLAYER]int //每个人的同色510K的个数，如1玩家共有5个510k，其中方块的2个，黑桃的2个，红桃的1个，则有两个同色510k（方块，黑桃）
	BombCount              [MAX_PLAYER]int //每个人的炸弹个数
	MaxBombCount           [MAX_PLAYER]int //每个人的炸弹最大个数
	ValidBombCount         [MAX_PLAYER]int //每个人的有效炸弹个数
	Bird                   [MAX_PLAYER]int //抓鸟，1表示有鸟
	RestartCount           int             //本局重新发牌次数

	PlayKingBomb [MAX_PLAYER]int // 打出的天炸，4王一起打出
	Play8Xi      [MAX_PLAYER]int // 打出的8喜个数
	Play7Xi      [MAX_PLAYER]int // 打出的7喜个数
	Play510K     [MAX_PLAYER]int // 打出的510K的个数

	//石首规则
	PlayerScore      [MAX_PLAYER]int //玩家当前得分
	PlayerTotalScore [MAX_PLAYER]int //玩家总得分
	JiCount          int             //级数
	BombScore        [MAX_PLAYER]int //当前炸弹分

	VecGameEnd  []static.Msg_S_DG_GameEnd         //记录每一局的结果
	VecGameData [MAX_PLAYER][]CMD_S_DG_StatusPlay //记录每一局的结束时的桌面数据

	//组件变量
	TimeGameRecord string            //记录当前日志的日期
	NameGameRecord string            //记录文件名
	GameRecordNum  string            //记录编号
	LastTime       int64             // 用于断线重入 校准客户端操作时间
	Config         static.GameConfig //游戏配置

	FCB          bool //第一局流局是不是要换庄家 fristFlowChangeBanker  公共
	IsLocalDebug bool
	OutScorePai  [24]byte //20200526 苏大强 武穴510k所出分牌
	JiaoPaiMate  uint16   //20200526 苏大强 武穴510k开局就要知道庄家队友
}

func (smss *SportMetaSS510K) GetRepertoryCards() (res [static.MAX_REPERTORY]byte) {
	return
}

func (smss *SportMetaSS510K) OnScoreOffset(total []float64) {
	for i, t := range total {
		var order SS510k_Replay_Order
		order.R_ChairId = uint16(i)
		order.R_Opt = info2.E_GameScore
		order.UserScorePL = t
		smss.ReplayRecord.R_Orders = append(smss.ReplayRecord.R_Orders, order)
	}
}

//如果有些变量的值在每小局都要保留，建议使用resetForNext
func (smss *SportMetaSS510K) ResetForNext() {
	//游戏变量
	smss.GameState = GsNull
	smss.GameType = GT_NULL
	smss.DownTime = 0
	smss.TheActionStep = AS_NULL
	smss.NowActionStep = AS_NULL

	smss.ReplayRecord.ReSet()

	//玩家信息
	smss.Banker = static.INVALID_CHAIR
	smss.BankParter = static.INVALID_CHAIR
	smss.Whoplay = static.INVALID_CHAIR
	smss.Player1 = static.INVALID_CHAIR
	smss.Player2 = static.INVALID_CHAIR
	smss.WhoLastOut = static.INVALID_CHAIR
	smss.WhoReady = [MAX_PLAYER]bool{}
	smss.WhoAntic = [MAX_PLAYER]int{-1, -1, -1, -1}
	smss.OutScorePai = [24]byte{}
	smss.WhoJiaoOrQiang = [MAX_PLAYER]int{}
	smss.JiabeiType = 0

	//托管和离线数据
	smss.TuoGuanPlayer = [MAX_PLAYER]bool{}
	smss.TrustCounts = [MAX_PLAYER]byte{}
	smss.AutoCardCounts = [MAX_PLAYER]byte{}
	smss.WhoBreak = [MAX_PLAYER]bool{}
	smss.BreakCounts = [MAX_PLAYER]byte{}
	smss.TrustPlayer = []byte{}

	//牌数据
	//smss.AllCards = [public.ALL_CARD]byte{}
	for j, _ := range smss.PlayerCards {
		for k, _ := range smss.PlayerCards[j] {
			smss.PlayerCards[j][k] = 0
		}
	}
	for j, _ := range smss.BombSplitCards {
		for k, _ := range smss.BombSplitCards[j] {
			smss.BombSplitCards[j][k] = 0
		}
	}
	for j, _ := range smss.AllPaiOut {
		for k, _ := range smss.AllPaiOut[j] {
			smss.AllPaiOut[j][k] = 0
		}
	}
	for j, _ := range smss.LastPaiOut {
		for k, _ := range smss.LastPaiOut[j] {
			smss.LastPaiOut[j][k] = 0
		}
	}
	smss.ThePaiCount = [MAX_PLAYER]byte{}
	smss.LastOutType = 0
	smss.LastOutTypeClient = 0
	smss.AllOutCnt = 0
	smss.WhoPass = [MAX_PLAYER]bool{}
	smss.BTeamOut = [MAX_PLAYER]bool{}
	smss.WhoAllOutted = [MAX_PLAYER]bool{}
	for j, _ := range smss.PlayerTurn {
		smss.PlayerTurn[j] = static.INVALID_BYTE
	}
	for j, _ := range smss.DownPai3P {
		smss.DownPai3P[j] = 0
	}

	//积分数据
	smss.Playerrich = [MAX_PLAYER]int{}
	smss.CardScore = 0
	smss.PlayerCardScore = [MAX_PLAYER]int{}
	smss.XiScore = [MAX_PLAYER]int{}

	//时间数据
	smss.PowerStartTime = time.Now().Unix()

	//叫牌相关数据
	smss.BMingJiFlag = false
	smss.DownPai = 0
	smss.RoarPai = 0
	smss.WhoRoar = static.INVALID_CHAIR
	smss.WhoAnti = static.INVALID_CHAIR
	smss.JiaoPaiMate = static.INVALID_CHAIR
	smss.WhoHas4KingScore = static.INVALID_CHAIR
	smss.WhoHas4KingPower = static.INVALID_CHAIR

	//smss.BSplited = false //这里不复位
	smss.BHaveFinish = false

	smss.WhoHasKingBomb = static.INVALID_CHAIR
	smss.Who8Xi = [MAX_PLAYER]int{}
	smss.HasKingNum = [MAX_PLAYER]int{}
	smss.Who7Xi = [MAX_PLAYER]int{}
	smss.WhoSame510K = [MAX_PLAYER]int{}
	smss.PlayKingBomb = [MAX_PLAYER]int{}
	smss.Play8Xi = [MAX_PLAYER]int{}
	smss.Play7Xi = [MAX_PLAYER]int{}
	smss.Play510K = [MAX_PLAYER]int{}
	//

	//状态变量
	smss.LastTime = 0
	smss.IsLocalDebug = false

	smss.WhoGonglongCount = [MAX_PLAYER]int{}
	smss.WhoKingCount = [MAX_PLAYER]int{}
	smss.Who510kCount = [MAX_PLAYER]int{}
	smss.WhoGonglongCount = [MAX_PLAYER]int{}
	smss.WhoGonglongScore = [MAX_PLAYER]int{}
	smss.WhoBombScore = [MAX_PLAYER]int{}
	smss.WhoOutCount = [MAX_PLAYER]int{}
	smss.BombCount = [MAX_PLAYER]int{}
	smss.ValidBombCount = [MAX_PLAYER]int{}
	smss.Bird = [MAX_PLAYER]int{}
	smss.RestartCount = 0
	smss.PlayerScore = [MAX_PLAYER]int{}
	smss.BombScore = [MAX_PLAYER]int{}
}

//Reset all
func (smss *SportMetaSS510K) Reset() {
	smss.ResetForNext()

	smss.AddSpecailBeishu = 0
	smss.Nextbanker = static.INVALID_CHAIR
	smss.AllCards = [static.ALL_CARD]byte{}
	smss.LastScore = [MAX_PLAYER]int{}
	smss.Total = [MAX_PLAYER]int{}
	smss.XuanPiao = [MAX_PLAYER]int{}

	smss.MaxBombCount = [MAX_PLAYER]int{}
	smss.MaxScore = [MAX_PLAYER]int{}
	smss.TotalFirstTurn = [MAX_PLAYER]int{}
	smss.TotalDuPai = [MAX_PLAYER]int{}
	smss.WhoTotal8Xi = [MAX_PLAYER]int{}
	smss.WhoTotal7Xi = [MAX_PLAYER]int{}
	smss.WhoTotalGonglong = [MAX_PLAYER]int{}
	smss.WhoToTalMore4KingCount = [MAX_PLAYER]int{}
	smss.WhoToTalChuntianCount = [MAX_PLAYER]int{}
	smss.WhoToTalplayKingBomb = [MAX_PLAYER]int{}
	smss.WhoToTalSame510K = [MAX_PLAYER]int{}
	smss.WinCount = [MAX_PLAYER]int{}
	smss.VecGameEnd = []static.Msg_S_DG_GameEnd{}
	smss.ReWriteRec = 0
	smss.VecGameData = [MAX_PLAYER][]CMD_S_DG_StatusPlay{}

	smss.PlayerTotalScore = [MAX_PLAYER]int{}
	smss.JiCount = 0

	//时间数据
	smss.AutoOutTime = 0
	smss.TimeStart = 0
	smss.PlayTime = 15
	smss.RoarTime = 15

	//底
	smss.Qiang = 0
	smss.IBase = 1
	smss.Spay = 0
	smss.SerPay = 0
	smss.FaOfTao = 0
	smss.JiangOfTao = 0
	smss.FristOut = 0
	smss.BiYa = 0
	smss.TuoGuan = 0
	smss.ZhaNiao = false
	smss.FourTake3 = false
	smss.BombSplit = false
	smss.QuickPass = false
	smss.SplitCards = false
	smss.Bomb3Ace = false
	smss.LessTake = false
	smss.FullScoreAward = 0
	smss.FourKingScore = 0
	smss.AddDiFen = 0
	smss.ShowHandCardCnt = true
	smss.GetLastScore = true
	smss.SeeTeamerCard = true
	smss.PlayMode = 0
	smss.RandTeamer = false
}
