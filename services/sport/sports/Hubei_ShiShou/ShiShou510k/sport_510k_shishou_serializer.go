package ShiShou510k

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
)

type SportSS510KJsonSerializer struct {
	components2.DGGameCommonJson
	//游戏变量
	ReplayRecord SS510k_Replay_Record `json:"replayrecord"` //回放记录
	ReWriteRec   byte                 `json:"rewriterec"`   //是否重复写回放数据，每小局游戏开始时清理,打拱可以在小结算中申请解散。

	//运行变量
	Banker         uint16                 `json:"banker"`      //庄家
	NextBanker     uint16                 `json:"nextbanker"`  //下一个庄家
	BankParter     uint16                 `json:"bankparter"`  //庄家的队友，2vs2模式下才有
	CurrentUser    uint16                 `json:"currentuser"` //当前用户
	Player1        uint16                 `json:"player1"`     //3人拱中，除庄家外的另一个人
	Player2        uint16                 `json:"player2"`     //3人拱中，除庄家外的另一个人
	WhoLastOut     uint16                 `json:"wholastout"`  //上一个出牌玩家，pass的不算
	WhoReady       [meta2.MAX_PLAYER]bool `json:"whoready"`    // 谁已经完成了吼牌过程
	WhoJiaoOrQiang [meta2.MAX_PLAYER]int  `json:"jiaoorqiang"` // 不叫和不抢要区分，有抢庄的3人拱

	//托管和离线数据
	TuoGuanPlayer  [meta2.MAX_PLAYER]bool `json:"tuoguanplayer"`  //谁托管了？
	TrustCounts    [meta2.MAX_PLAYER]byte `json:"trustcounts"`    //玩家托管次数
	AutoCardCounts [meta2.MAX_PLAYER]byte `json:"autocardcounts"` //自动出牌的次数
	WhoBreak       [meta2.MAX_PLAYER]bool `json:"whobreak"`       //谁断线了？
	BreakCounts    [meta2.MAX_PLAYER]byte `json:"breakcounts"`    // 断线次数
	TrustPlayer    []byte                 `json:"trustplayer"`    //托管玩家
	TrustOrder     [meta2.MAX_PLAYER]byte `json:"trustorder"`     //玩家托管顺序

	//牌数据
	AllCards_CYDG  [meta2.TS_ALLCARD]byte                  `json:"allcardscydg"`   // 所有牌，崇阳打滚使用
	AllCards       [meta2.TS_ALLCARD]byte                  `json:"allcards"`       // 所有牌
	PlayerCards    [meta2.MAX_PLAYER][static.MAX_CARD]byte `json:"playercards"`    // 玩家分到的牌
	BombSplitCards [meta2.MAX_PLAYER][static.MAX_CARD]byte `json:"bombsplitcards"` // 拆掉炸弹的牌

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
	LongzhaDing        int  `json:"longzhading"`       // 笼炸封顶
	FakeKingValue      byte `json:"fakekingvalue"`     // 王单出算几
	FapaiMode          int  `json:"fapaimode"`         // 发牌模式，，0表示起牌拱，1表示发牌拱，2表示疯狂拱，3表示变态拱
	ShowCardNum        int  `json:"showcardnum"`       // 是否显示手牌数目
	MaxKingNum         byte `json:"maxkingnum"`        // 4王或8王
	KanJie             int  `json:"kanjie"`            //坎阶梯：5分/坎，10分/坎，20分/坎
	KanScore           int  `json:"kanscore"`          //坎分：1-20，1，2，3，4....20
	BaoDi              int  `json:"baodi"`             //保底：5-100，5，10，15....100,默认5  ,
	Has510k            int  `json:"has510k"`           //510选项：510k是炸弹和510k不是炸弹，单选项，默认选中510k是炸弹，选择510k不是炸弹，癞子用法下面是禁用不能选中
	Magic510k          int  `json:"magic510k"`         //癞子用法：癞子可组510k和癞子不可组510k，单选项，默认选中癞子可组510k
	BombStr            bool `json:"bombstr"`           //是否有摇摆
	HasPiao            bool `json:"haspiao"`           // 是否选漂
	BombErr            bool `json:"bomberr"`           // 是否炸错罚分
	BombErrBankerCount int  `json:"bomberrbankcount"`  // 跟庄炸庄次数
	SkyCnt             int  `json:"skycnt"`            // 花牌数目
	ZhuangBuJie        bool `json:"zhuangbujie"`       //庄家是否不接风
	FristOut           int  `json:"fristout"`          //首出类型 ,0黑桃3先出，庄家首出
	BiYa               int  `json:"biya"`              //有大比压 ,0有大必压，1可以不压，
	TuoGuan            int  `json:"tuoguan"`           //托管 ,0不托管，大于0托管，
	ZhaNiao            bool `json:"zhaniao"`           //是否扎鸟，红桃10是鸟
	FourTake3          bool `json:"fourtake3"`         //是否可以4带3
	BombSplit          bool `json:"bombsplit"`         //炸弹是否可拆
	QuickPass          bool `json:"quickpass"`         //是否快速过牌
	SplitCards         bool `json:"splitcards"`        //是否有切牌
	Bomb3Ace           bool `json:"bomb3ace"`          //3个A是否是炸弹
	LessTake           bool `json:"lesstake"`          //最后一手是否可以少带
	Jiao2King          int  `json:"jiao2king"`         //双王必须叫牌
	TeamCard           int  `json:"teamcard"`          //先跑可看队友牌
	KeFan              bool `json:"kefan"`             //是否可以反春
	FourTake2          bool `json:"fourtake2"`         //是否可以4带2
	FourTake1          bool `json:"fourtake1"`         //4带1是否算炸弹
	CardNum            int  `json:"cardnum"`           //牌数
	KingLai            int  `json:"kinglai"`           //王是否可做赖子，0无癞子，1有癞子不讲硬炸，2有癞子讲硬炸
	FullScoreAward     int  `json:"fullscoreaward"`    //满分奖励
	FourKingScore      int  `json:"fourkingscore"`     //4王换分
	AddDiFen           int  `json:"adddifen"`          //额外加的底分
	ShowHandCardCnt    bool `json:"showhandcardcnt"`   //是否显示手牌数
	GetLastScore       bool `json:"getlastscore"`      //是否可以捡尾分
	SeeTeamerCard      bool `json:"seeteamercard"`     //看队友牌
	Big510k            bool `json:"big510k"`           //6炸7炸不可打510k
	BombMode           bool `json:"bombmode"`          //炸弹被压无分
	ExtAdd             bool `json:"extadd"`            //额外加分，比如打出3王1花得1倍底分，跟打出8喜7喜的加分不一样
	TrusteeCost        bool `json:"trusteecost"`       //托管玩家承担所有输分
	Restart            bool `json:"restart"`           //重新发牌
	TimeOutPunish      bool `json:"timeoutpunish"`     // 超时罚分
	NotDismiss         bool `json:"notdismiss"`        //托管不解散
	FristOutMode       int  `json:"fristoutmode"`      //首出出牌类型 ,0必带黑三或最小牌，1任意出牌
	NoBomb             bool `json:"nobomb"`            //纯净玩法，勾选时不发炸弹
	BombRealTime       bool `json:"bombrealtimescore"` //炸弹实时计分
	PlayMode           int  `json:"playmode"`          //石首玩法(0经典  1双赖   2四赖)
	RandTeamer         bool `json:"randTeamer"`        //随机队友

	CurPeriod       int     `json:"curperiod"`       // 超时罚分的当前阶段
	PunishStartTime int64   `json:"punishstarttime"` // 超时罚分的开始时间
	HasAction       [4]bool `json:"hasaction"`       // 超时罚分的动作标志
	PunishCount     [4]int  `json:"punishcount"`     // 超时罚分的次数计数

	//叫牌相关数据
	MingJiFlag   bool                         `json:"mingjiflag"`   // 是否已经显示明鸡了
	DownPai      byte                         `json:"downpai"`      // 4人拱做牌时的叫牌
	DownPai3P    [static.MAX_DOWNCARDNUM]byte `json:"downpai3p"`    // 3人拱的底牌，做牌也用这个数组
	RoarPai      byte                         `json:"roarpai"`      // 叫的牌
	WhoRoar      uint16                       `json:"whoroar"`      // 谁吼牌了？
	JiabeiType   int                          `json:"jiabeitype"`   // 有抢庄的3人拱的加倍类型
	BSplited     bool                         `json:"bsplited"`     // 是否选择了切牌
	RestartCount int                          `json:"restartcount"` // 重新发牌次数

	WhoHas4KingScore uint16 `json:"whohas4kingscore"` // 谁用4王换分了
	WhoHas4KingPower uint16 `json:"whohas4kingpower"` // 谁拥有4王换分权限（换完就没有该牌权了）

	PlayerScore      [meta2.MAX_PLAYER]int `json:"playerscore"`      //玩家当前得分
	PlayerTotalScore [meta2.MAX_PLAYER]int `json:"playertotalscore"` //玩家总得分
	JiCount          int                   `json:"jicount"`          //级数
	BombScore        [meta2.MAX_PLAYER]int `json:"bombscore"`        //玩家当前得分

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

	//状态变量
	VecGameEnd  []static.Msg_S_DG_GameEnd         `json:"vecgameend"`  //记录每一局的结果
	VecGameData [MAX_PLAYER][]CMD_S_DG_StatusPlay `json:"vecgamedata"` //记录每一局的结果

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

func (ssjs *SportSS510KJsonSerializer) ToJsonSS510k(_game *SportMetaSS510K) {
	ssjs.ReplayRecord = _game.ReplayRecord
	ssjs.ReWriteRec = _game.ReWriteRec
	ssjs.Banker = _game.Banker
	ssjs.NextBanker = _game.Nextbanker
	ssjs.BankParter = _game.BankParter
	ssjs.CurrentUser = _game.Whoplay
	ssjs.Player1 = _game.Player1
	ssjs.Player2 = _game.Player2
	ssjs.WhoLastOut = _game.WhoLastOut
	ssjs.TrustPlayer = _game.TrustPlayer
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		ssjs.WhoJiaoOrQiang[i] = _game.WhoJiaoOrQiang[i]
		ssjs.WhoReady[i] = _game.WhoReady[i]
		ssjs.TuoGuanPlayer[i] = _game.TuoGuanPlayer[i]
		ssjs.TrustCounts[i] = _game.TrustCounts[i]
		ssjs.AutoCardCounts[i] = _game.AutoCardCounts[i]
		ssjs.WhoBreak[i] = _game.WhoBreak[i]
		ssjs.BreakCounts[i] = _game.BreakCounts[i]
		ssjs.WhoPass[i] = _game.WhoPass[i]
		ssjs.BTeamOut[i] = _game.BTeamOut[i]
		ssjs.WhoAllOutted[i] = _game.WhoAllOutted[i]
		ssjs.PlayerTurn[i] = _game.PlayerTurn[i]
		for j := 0; j < static.MAX_CARD; j++ {
			ssjs.PlayerCards[i][j] = _game.PlayerCards[i][j]
			ssjs.AllPaiOut[i][j] = _game.AllPaiOut[i][j]
			ssjs.LastPaiOut[i][j] = _game.LastPaiOut[i][j]
			ssjs.BombSplitCards[i][j] = _game.BombSplitCards[i][j]
		}

		ssjs.ThePaiCount[i] = _game.ThePaiCount[i]
		ssjs.XuanPiao[i] = _game.XuanPiao[i]
	}
	for i := 0; i < static.ALL_CARD; i++ {
		ssjs.AllCards[i] = _game.AllCards[i]
	}

	ssjs.LastOutType = _game.LastOutType
	ssjs.LastOutTypeClient = _game.LastOutTypeClient
	ssjs.AllOutCnt = _game.AllOutCnt

	ssjs.OutScorePai = _game.OutScorePai

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		ssjs.LastScore[i] = _game.LastScore[i]
		ssjs.Total[i] = _game.Total[i]
		ssjs.Playerrich[i] = _game.Playerrich[i]
		ssjs.PlayerCardScore[i] = _game.PlayerCardScore[i]
		ssjs.XiScore[i] = _game.XiScore[i]
	}
	ssjs.CardScore = _game.CardScore
	ssjs.AutoOutTime = _game.AutoOutTime
	ssjs.TimeStart = _game.TimeStart
	ssjs.PlayTime = _game.PlayTime
	ssjs.RoarTime = _game.RoarTime
	ssjs.PowerStartTime = _game.PowerStartTime
	ssjs.Qiang = _game.Qiang
	ssjs.Base = _game.IBase
	ssjs.Spay = _game.Spay
	ssjs.SerPay = _game.SerPay
	ssjs.FaOfTao = _game.FaOfTao
	ssjs.JiangOfTao = _game.JiangOfTao
	ssjs.AddSpecailBeishu = _game.AddSpecailBeishu
	ssjs.FristOut = _game.FristOut
	ssjs.BiYa = _game.BiYa
	ssjs.TuoGuan = _game.TuoGuan
	ssjs.ZhaNiao = _game.ZhaNiao
	ssjs.FourTake3 = _game.FourTake3
	ssjs.BombSplit = _game.BombSplit
	ssjs.QuickPass = _game.QuickPass
	ssjs.SplitCards = _game.SplitCards
	ssjs.Bomb3Ace = _game.Bomb3Ace
	ssjs.LessTake = _game.LessTake
	ssjs.Jiao2King = _game.Jiao2King
	ssjs.TeamCard = _game.TeamCard
	ssjs.KeFan = _game.KeFan
	ssjs.FourTake2 = _game.FourTake2
	ssjs.FourTake1 = _game.FourTake1
	ssjs.CardNum = _game.CardNum
	ssjs.KingLai = _game.KingLai
	ssjs.Big510k = _game.Big510k
	ssjs.FullScoreAward = _game.FullScoreAward
	ssjs.FourKingScore = _game.FourKingScore
	ssjs.AddDiFen = _game.AddDiFen
	ssjs.ShowHandCardCnt = _game.ShowHandCardCnt
	ssjs.GetLastScore = _game.GetLastScore
	ssjs.SeeTeamerCard = _game.SeeTeamerCard
	ssjs.BombMode = _game.BombMode
	ssjs.Restart = _game.Restart
	ssjs.Piao = _game.Piao
	ssjs.PiaoCount = _game.PiaoCount
	ssjs.NotDismiss = _game.NotDismiss
	ssjs.NoBomb = _game.NoBomb
	ssjs.FristOutMode = _game.FristOutMode
	ssjs.BombRealTime = _game.BombRealTime
	ssjs.PlayMode = _game.PlayMode
	ssjs.RandTeamer = _game.RandTeamer

	ssjs.MingJiFlag = _game.BMingJiFlag
	ssjs.DownPai = _game.DownPai
	ssjs.RoarPai = _game.RoarPai
	ssjs.WhoRoar = _game.WhoRoar
	ssjs.WhoHasKingBomb = _game.WhoHasKingBomb
	ssjs.JiabeiType = _game.JiabeiType
	ssjs.BSplited = _game.BSplited
	ssjs.AddXiScore = _game.AddXiScore
	ssjs.RestartCount = _game.RestartCount

	ssjs.WhoHas4KingScore = _game.WhoHas4KingScore
	ssjs.WhoHas4KingPower = _game.WhoHas4KingPower

	ssjs.JiCount = _game.JiCount

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		ssjs.DownPai3P[i] = _game.DownPai3P[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		ssjs.HasKingNum[i] = _game.HasKingNum[i]
		ssjs.Who8Xi[i] = _game.Who8Xi[i]
		ssjs.Who7Xi[i] = _game.Who7Xi[i]
		ssjs.WhoSame510K[i] = _game.WhoSame510K[i]
		ssjs.MaxScore[i] = _game.MaxScore[i]
		ssjs.TotalFirstTurn[i] = _game.TotalFirstTurn[i]
		ssjs.TotalDuPai[i] = _game.TotalDuPai[i]
		ssjs.WhoKingCount[i] = _game.WhoKingCount[i]
		ssjs.Who510kCount[i] = _game.Who510kCount[i]
		ssjs.WhoTotal8Xi[i] = _game.WhoTotal8Xi[i]
		ssjs.WhoTotal7Xi[i] = _game.WhoTotal7Xi[i]
		ssjs.WhoTotalGonglong[i] = _game.WhoTotalGonglong[i]
		ssjs.WhoGonglongCount[i] = _game.WhoGonglongCount[i]
		ssjs.WhoToTalMore4KingCount[i] = _game.WhoToTalMore4KingCount[i]
		ssjs.WhoGonglongScore[i] = _game.WhoGonglongScore[i]
		ssjs.PlayKingBomb[i] = _game.PlayKingBomb[i]
		ssjs.Play8Xi[i] = _game.Play8Xi[i]
		ssjs.Play7Xi[i] = _game.Play7Xi[i]
		ssjs.Play510K[i] = _game.Play510K[i]
		ssjs.WhoBombScore[i] = _game.WhoBombScore[i]
		ssjs.WhoOutCount[i] = _game.WhoOutCount[i]
		ssjs.WhoToTalChuntianCount[i] = _game.WhoToTalChuntianCount[i]
		ssjs.BombCount[i] = _game.BombCount[i]
		ssjs.MaxBombCount[i] = _game.MaxBombCount[i]
		ssjs.Bird[i] = _game.Bird[i]
		ssjs.ValidBombCount[i] = _game.ValidBombCount[i]
		ssjs.PlayerScore[i] = _game.PlayerScore[i]
		ssjs.PlayerTotalScore[i] = _game.PlayerTotalScore[i]
		ssjs.BombScore[i] = _game.BombScore[i]
		ssjs.VecGameData[i] = _game.VecGameData[i]
	}

	ssjs.VecGameEnd = _game.VecGameEnd
	ssjs.TimeGameRecord = _game.TimeGameRecord
	ssjs.NameGameRecord = _game.NameGameRecord
	ssjs.GameRecordNum = _game.GameRecordNum
	ssjs.LastTime = _game.LastTime
	ssjs.GameState = _game.GameState
	ssjs.GameType = _game.GameType
	//添加保存game记录
	ssjs.GameConfig = _game.Config
}

func (ssjs *SportSS510KJsonSerializer) UnmarshaSS510k(_game *SportMetaSS510K) {

	_game.ReplayRecord = ssjs.ReplayRecord
	fmt.Println(_game.ReplayRecord)
	// for k, _ := range _game.RepertoryCard {
	// 	if k >= len(ssjs.RepertoryCard) {
	// 		break
	// 	}
	// 	_game.RepertoryCard[k] = ssjs.RepertoryCard[k]
	// }
	_game.ReWriteRec = ssjs.ReWriteRec
	_game.Banker = ssjs.Banker
	_game.Nextbanker = ssjs.NextBanker
	_game.BankParter = ssjs.BankParter
	_game.Whoplay = ssjs.CurrentUser
	_game.Player1 = ssjs.Player1
	_game.Player2 = ssjs.Player2
	_game.WhoLastOut = ssjs.WhoLastOut
	_game.TrustPlayer = ssjs.TrustPlayer

	_game.OutScorePai = ssjs.OutScorePai

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.WhoJiaoOrQiang[i] = ssjs.WhoJiaoOrQiang[i]
		_game.WhoReady[i] = ssjs.WhoReady[i]
		_game.TuoGuanPlayer[i] = ssjs.TuoGuanPlayer[i]
		_game.TrustCounts[i] = ssjs.TrustCounts[i]
		_game.AutoCardCounts[i] = ssjs.AutoCardCounts[i]
		_game.WhoBreak[i] = ssjs.WhoBreak[i]
		_game.BreakCounts[i] = ssjs.BreakCounts[i]
		_game.WhoPass[i] = ssjs.WhoPass[i]
		_game.BTeamOut[i] = ssjs.BTeamOut[i]
		_game.WhoAllOutted[i] = ssjs.WhoAllOutted[i]
		_game.PlayerTurn[i] = ssjs.PlayerTurn[i]
		for j := 0; j < static.MAX_CARD; j++ {
			_game.PlayerCards[i][j] = ssjs.PlayerCards[i][j]
			_game.AllPaiOut[i][j] = ssjs.AllPaiOut[i][j]
			_game.LastPaiOut[i][j] = ssjs.LastPaiOut[i][j]
			_game.BombSplitCards[i][j] = ssjs.BombSplitCards[i][j]
		}

		_game.ThePaiCount[i] = ssjs.ThePaiCount[i]
		_game.XuanPiao[i] = ssjs.XuanPiao[i]
	}
	for i := 0; i < static.ALL_CARD; i++ {
		_game.AllCards[i] = ssjs.AllCards[i]
	}

	_game.LastOutType = ssjs.LastOutType
	_game.LastOutTypeClient = ssjs.LastOutTypeClient
	_game.AllOutCnt = ssjs.AllOutCnt

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.LastScore[i] = ssjs.LastScore[i]
		_game.Total[i] = ssjs.Total[i]
		_game.Playerrich[i] = ssjs.Playerrich[i]
		_game.PlayerCardScore[i] = ssjs.PlayerCardScore[i]
		_game.XiScore[i] = ssjs.XiScore[i]
	}
	_game.CardScore = ssjs.CardScore
	_game.AutoOutTime = ssjs.AutoOutTime
	_game.TimeStart = ssjs.TimeStart
	_game.PlayTime = ssjs.PlayTime
	_game.RoarTime = ssjs.RoarTime
	_game.PowerStartTime = ssjs.PowerStartTime
	_game.Qiang = ssjs.Qiang
	_game.IBase = ssjs.Base
	_game.Spay = ssjs.Spay
	_game.SerPay = ssjs.SerPay
	_game.FaOfTao = ssjs.FaOfTao
	_game.JiangOfTao = ssjs.JiangOfTao
	_game.AddSpecailBeishu = ssjs.AddSpecailBeishu
	_game.FristOut = ssjs.FristOut
	_game.BiYa = ssjs.BiYa
	_game.TuoGuan = ssjs.TuoGuan
	_game.ZhaNiao = ssjs.ZhaNiao
	_game.FourTake3 = ssjs.FourTake3
	_game.BombSplit = ssjs.BombSplit
	_game.QuickPass = ssjs.QuickPass
	_game.SplitCards = ssjs.SplitCards
	_game.Bomb3Ace = ssjs.Bomb3Ace
	_game.LessTake = ssjs.LessTake
	_game.TeamCard = ssjs.TeamCard
	_game.Jiao2King = ssjs.Jiao2King
	_game.KeFan = ssjs.KeFan
	_game.FourTake2 = ssjs.FourTake2
	_game.FourTake1 = ssjs.FourTake1
	_game.CardNum = ssjs.CardNum
	_game.KingLai = ssjs.KingLai
	_game.Big510k = ssjs.Big510k
	_game.FullScoreAward = ssjs.FullScoreAward
	_game.FourKingScore = ssjs.FourKingScore
	_game.AddDiFen = ssjs.AddDiFen
	_game.ShowHandCardCnt = ssjs.ShowHandCardCnt
	_game.GetLastScore = ssjs.GetLastScore
	_game.SeeTeamerCard = ssjs.SeeTeamerCard
	_game.BombMode = ssjs.BombMode
	_game.Restart = ssjs.Restart
	_game.Piao = ssjs.Piao
	_game.PiaoCount = ssjs.PiaoCount
	_game.NotDismiss = ssjs.NotDismiss
	_game.FristOutMode = ssjs.FristOutMode
	_game.NoBomb = ssjs.NoBomb
	_game.BombRealTime = ssjs.BombRealTime
	_game.PlayMode = ssjs.PlayMode
	_game.RandTeamer = ssjs.RandTeamer

	_game.BMingJiFlag = ssjs.MingJiFlag
	_game.DownPai = ssjs.DownPai
	_game.RoarPai = ssjs.RoarPai
	_game.WhoRoar = ssjs.WhoRoar
	_game.WhoHasKingBomb = ssjs.WhoHasKingBomb
	_game.JiabeiType = ssjs.JiabeiType
	_game.BSplited = ssjs.BSplited
	_game.AddXiScore = ssjs.AddXiScore
	_game.RestartCount = ssjs.RestartCount

	_game.WhoHas4KingScore = ssjs.WhoHas4KingScore
	_game.WhoHas4KingPower = ssjs.WhoHas4KingPower

	_game.JiCount = ssjs.JiCount

	for i := 0; i < static.MAX_DOWNCARDNUM; i++ {
		_game.DownPai3P[i] = ssjs.DownPai3P[i]
	}

	for i := 0; i < meta2.MAX_PLAYER; i++ {
		_game.HasKingNum[i] = ssjs.HasKingNum[i]
		_game.Who8Xi[i] = ssjs.Who8Xi[i]
		_game.Who7Xi[i] = ssjs.Who7Xi[i]
		_game.WhoSame510K[i] = ssjs.WhoSame510K[i]
		_game.MaxScore[i] = ssjs.MaxScore[i]
		_game.TotalFirstTurn[i] = ssjs.TotalFirstTurn[i]
		_game.TotalDuPai[i] = ssjs.TotalDuPai[i]
		_game.WhoKingCount[i] = ssjs.WhoKingCount[i]
		_game.Who510kCount[i] = ssjs.Who510kCount[i]
		_game.WhoTotal8Xi[i] = ssjs.WhoTotal8Xi[i]
		_game.WhoTotal7Xi[i] = ssjs.WhoTotal7Xi[i]
		_game.WhoTotalGonglong[i] = ssjs.WhoTotalGonglong[i]
		_game.WhoGonglongCount[i] = ssjs.WhoGonglongCount[i]
		_game.WhoToTalMore4KingCount[i] = ssjs.WhoToTalMore4KingCount[i]
		_game.WhoGonglongScore[i] = ssjs.WhoGonglongScore[i]
		_game.PlayKingBomb[i] = ssjs.PlayKingBomb[i]
		_game.Play8Xi[i] = ssjs.Play8Xi[i]
		_game.Play7Xi[i] = ssjs.Play7Xi[i]
		_game.Play510K[i] = ssjs.Play510K[i]
		_game.WhoBombScore[i] = ssjs.WhoBombScore[i]
		_game.WhoOutCount[i] = ssjs.WhoOutCount[i]
		_game.WhoToTalChuntianCount[i] = ssjs.WhoToTalChuntianCount[i]
		_game.BombCount[i] = ssjs.BombCount[i]
		_game.MaxBombCount[i] = ssjs.MaxBombCount[i]
		_game.Bird[i] = ssjs.Bird[i]
		_game.ValidBombCount[i] = ssjs.ValidBombCount[i]
		_game.PlayerScore[i] = ssjs.PlayerScore[i]
		_game.PlayerTotalScore[i] = ssjs.PlayerTotalScore[i]
		_game.BombScore[i] = ssjs.BombScore[i]
		_game.VecGameData[i] = ssjs.VecGameData[i]
	}

	_game.VecGameEnd = ssjs.VecGameEnd
	_game.TimeGameRecord = ssjs.TimeGameRecord
	_game.NameGameRecord = ssjs.NameGameRecord
	_game.GameRecordNum = ssjs.GameRecordNum
	_game.LastTime = ssjs.LastTime
	_game.GameState = ssjs.GameState
	_game.GameType = ssjs.GameType

	//_game.Config             =        ssjs.GameConfig
}

//获取累加的计算
func (ssjs *SportSS510KJsonSerializer) getString() string {
	var str string
	return str
}
