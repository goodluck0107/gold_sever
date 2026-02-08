package JingZhouHuaPai

import (
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
)

//基础定义
const (
	TCGZ_ALLCARD       = 112 //牌数目
	TCGZ_MAXHANDCARD   = 60  //最大手牌数
	TCGZ_MAXSENDCARD   = 26  //发牌时最大手牌数
	TCGZ_MAX_INDEX     = 23  //最大索引
	TCGZ_MAX_INDEX_HUA = 28  //最大索引,带花牌
	TCGZ_MAX_PLAYER    = 3   //游戏人数
)

const (
	DanBeiFen    = 5  //单倍分数
	ShuangBeiFen = 6  //双倍分数
	BaBeiFen     = 12 //八倍分数
)

//enHuPaiKind
const (
	enHuNull   = iota //没有胡牌
	enQuanHong        //一手红 10倍
	enShiDui          //十对 2倍
	enQuanHei         //一手黑 3倍
	enOther           //其他
)

//ActionStep
const (
	ZP_AS_NULL          = iota
	ZP_AS_GAMESTART     // 游戏开始
	ZP_AS_STARTTIANHU   // 天胡
	ZP_AS_STARTTIANLONG // 天拢
	ZP_AS_STARTPLAY     // 开始出牌
	ZP_AS_SENDCARD      // 开始发牌,上家出牌后没有人要，需要等1s在发下一张牌
	ZP_AS_QZ_KAICHAO    // 强制开朝
	ZP_AS_QZ_DIANPAO    // 点炮必胡
	ZP_AS_ENDGAME       // 结束游戏
	ZP_AS_XUANPIAO      // 选漂
	ZP_AS_GUANSHENG     //观生
	ZP_AS_PLAYCARD      //出牌阶段
	ZP_AS_COUNT
)

//enEstimatKind
const (
	ZP_EstimatKind_NULL         = iota
	ZP_EstimatKind_SendCard     // 发牌效验
	ZP_EstimatKind_OutCard      // 出牌效验
	ZP_EstimatKind_GangCard     // 杠牌效验,杠后校验响应
	ZP_EstimatKind_AnGangCard   // 暗杠牌效验
	ZP_EstimatKind_MingGangCard //明杠牌效验
)

//enGuoZhangKind
const (
	gKind_EatCard  = iota //吃牌效验
	gKind_PengCard        //碰牌效验
	gKind_HuCard          //胡牌效验
)

//牌型设计,做牌时从str转换成byte
var m_strZPCardsMessageUP = [TCGZ_MAX_INDEX_HUA]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09", "0x0A", "0x0B",
	"0x0C", "0x0D", "0x0E", "0x0F", "0x10", "0x11", "0x12", "0x13", "0x14", "0x15", "0x16",
	"0x17", "0x18", "0x19", "0x1A", "0x1B", "0x1C",
}

//牌型设计,做牌时从str转换成byte,做牌小写字母时需要
var m_strZPCardsMessageLW = [TCGZ_MAX_INDEX_HUA]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09", "0x0a", "0x0b",
	"0x0c", "0x0d", "0x0e", "0x0f", "0x10", "0x11", "0x12", "0x13", "0x14", "0x15", "0x16",
	"0x17", "0x18", "0x19", "0x1a", "0x1b", "0x1c",
}

/*
 * 回放相关结构体
 */
//出牌记录类型
const (
	E_ZP_SendCard     = iota //发牌
	E_ZP_OutCard             //出牌
	E_ZP_Wik_Left            //吃
	E_ZP_Wik_Center          //吃
	E_ZP_Wik_Right           //吃
	E_ZP_Wik_1X2D            //吃，绞
	E_ZP_Wik_2X1D            //吃，绞
	E_ZP_Wik_2710            //吃，2710
	E_ZP_Peng                //碰牌
	E_ZP_TianLong            //天拢
	E_ZP_Gang                //杠牌，明杠
	E_ZP_Hu                  //胡
	E_ZP_HuangZhuang         //荒
	E_ZP_Li_Xian             //离线
	E_ZP_Jie_san             //解散
	E_ZP_GameScore           //玩家疲劳值
	E_ZP_Wik_Hua             //滑
	E_ZP_Wik_Jian            //捡
	E_ZP_Wik_Half            //半句
	E_ZP_PIAO                //选漂
	E_ZP_Tuo_Guan            //托管
	E_ZP_Wik_Ta              //踏
	E_ZP_NoOut               //捏牌
	E_SendCardRight          //发送牌权
	E_HandleCardRight        //处理牌权
)

//Replay_Ext_Type
const (
	E_ZP_Ext_NUL         = iota //
	E_ZP_Ext_HuXi               //胡息
	E_ZP_Ext_Provider           //吃碰杠牌提供者
	E_ZP_Ext_ProvideCard        //吃碰杠牌提供牌,通城个子有花色，需要这个标记
	E_ZP_Ext_PiaoScore          //漂分
	E_ZP_Ext_Jian               //捡牌标记
	E_ZP_Ext_Tong               //统数更新
)

type ZP_Replay_Order_Ext struct {
	Ext_type  int `json:"exttype"`  //玩家 //操作类型
	Ext_value int `json:"extvalue"` //操作值
}

type ZP_Replay_Order struct {
	R_ChairId   uint16                `json:"id"`         //玩家
	R_Opt       int                   `json:"operation"`  //记录类型
	R_Value     []int                 `json:"value"`      //出牌
	R_Opt_Ext   []ZP_Replay_Order_Ext `json:"orderext"`   //出牌
	UserScorePL float64               `json:"user_score"` // 玩家疲劳值，目前跟分数一致
}

func (self *ZP_Replay_Order) AddReplayExtData(exttype int, extvalue int) {
	var ext_data ZP_Replay_Order_Ext
	ext_data.Ext_type = exttype
	ext_data.Ext_value = extvalue

	self.R_Opt_Ext = append(self.R_Opt_Ext, ext_data)
}

type ZP_Replay_Record struct {
	R_HandCards     [TCGZ_MAX_PLAYER][]int      `json:"handcard"`      //用户最初手上的牌
	R_Orders        []ZP_Replay_Order           `json:"orders"`        //用户操作
	R_HuType        int                         `json:"hutype"`        //标记胡牌类型
	R_TotalHuxi     int                         `json:"totalhuxi"`     //总胡息
	R_GeZiShu       int                         `json:"gezishu"`       //个子数
	R_HuFen         int                         `json:"hufen"`         //胡分
	R_EndSubType    byte                        `json:"endsubtype"`    //0不是自摸也不是放炮，1自摸，2荒庄，3放炮
	R_ProvideUser   uint16                      `json:"provideuser"`   // 点炮用户
	R_WeaveCount    byte                        `json:"weavecount"`    //组合数目	//胡牌玩家的最终数据
	R_WeaveItemInfo [10][5]byte                 `json:"weaveiteminfo"` //胡牌牌型中每梯牌数据,最多5组组合牌//胡牌玩家的最终数据
	R_Score         [TCGZ_MAX_PLAYER]int        `json:"score"`         //游戏积分
	UVitamin        map[int64]float64           `json:"u_vitamin"`     // 玩家起始疲劳值
	R_LeftCount     int                         `json:"leftcount"`     //牌堆剩余牌数目
	EndInfo         *static.Msg_S_ZP_TC_GameEnd `json:"end_inf"`
}

func (self *ZP_Replay_Record) ReSet() {
	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		self.R_HandCards[i] = []int{}
	}
	self.R_Orders = []ZP_Replay_Order{}
	self.R_HuType = 0
	self.R_TotalHuxi = 0
	self.R_GeZiShu = 0
	self.R_HuFen = 0
	self.R_EndSubType = 0
	self.R_ProvideUser = 0
	self.R_WeaveCount = 0
	self.R_WeaveItemInfo = [10][5]byte{}
	self.R_Score = [TCGZ_MAX_PLAYER]int{}
	self.UVitamin = make(map[int64]float64)
	self.R_LeftCount = 0
	self.EndInfo = nil
}

//类型子项
type TagKindItem struct {
	WeaveKind int     `json:"weavekind"` //组合类型
	CardIndex [5]byte `json:"cardindex"` //扑克索引
}

//组合子项
type TagWeaveItem struct {
	WeaveKind   int     `json:"weavekind"`   //组合类型
	Cards       [5]byte `json:"cards"`       //扑克,大冶字牌最多4个，通城个子最多5个
	PublicCard  byte    `json:"publiccard"`  //公开标志，遮罩要用到这张牌
	ProvideUser uint16  `json:"provideuser"` //供应用户
}

//杠牌结果
type TagGangCardResult struct {
	MagicCard byte     `json:"magiccard"` //赖子牌
	CardCount byte     `json:"cardcount"` //扑克数目
	CardData  [10]byte `json:"carddata"`  //扑克数据
}

//分析子项
type TagAnalyseItem struct {
	CardEye   byte        `json:"cardeye"`   //牌眼扑克
	WeaveKind [10]int     `json:"weavekind"` //组合类型
	Cards     [10][5]byte `json:"cards"`     //扑克数据
	IsWeave   [10]int     `json:"isweave"`   //每梯的牌是否是组合区的牌
}

//胡牌详细结果
type TagChiHuItemInfo struct {
	CardEye       byte         `json:"cardeye"`       //牌眼扑克
	WeaveCount    int          `json:"weavecount"`    //组合数目
	WeaveKind     [10]int      `json:"weavekind"`     //每梯牌数据的组合类型
	WeaveItemInfo [10][10]byte `json:"weaveiteminfo"` //胡牌牌型中每梯牌数据
	HuXiInfo      [10]int      `json:"huxiinfo"`      //每梯牌胡息
	IsWeave       [10]int      `json:"isweave"`       //每梯的牌是否是组合区的牌
	TotalHuxi     int          `json:"totalhuxi"`     //总胡息
	MainJingIndex byte         `json:"mainjingindex"` //主精是谁
}

//胡牌结果
type TagChiHuResult struct {
	ChiHuKind          int                `json:"chihukind"`        //吃胡类型
	ChiHuRight         int                `json:"chihuright"`       //胡牌权位
	ChiHuUser          uint16             `json:"chihuuser"`        //吃胡玩家
	ChiHuItemInfoArray []TagChiHuItemInfo `json:"chihuiteminfoarr"` //吃胡详细信息
}

func (self *TagChiHuResult) reset() {
	self.ChiHuKind = 0
	self.ChiHuRight = 0
	self.ChiHuUser = 0
	self.ChiHuItemInfoArray = []TagChiHuItemInfo{}
}

//统的信息
type TagTongInfo struct {
	CardTongInfo [TCGZ_MAX_INDEX]TagTongItem `json:"cardTongInfo"` //牌
	TongCnt      int                         `json:"tongcnt"`      //统次数
}

//统的信息
type TagTongItem struct {
	Cards    [10]byte `json:"cards"`    //牌
	TongFlag bool     `json:"tongflag"` //是否统过
	TaFlag   bool     `json:"taflag"`   //踏标记
	TongCnt  int      `json:"tongcnt"`  //统次数
	HunCnt   int      `json:"huncnt"`   //使用了几个赖子,当Card是赖子时，这里表示还剩余几个赖子没有使用
}

//听牌结果
type TagTingCardResult struct {
	Seat       uint16                `json:"seat"`
	MaxCount   int                   `json:"maxcount"`
	TingIndex  [static.MAX_CARD]bool `json:"tingindex"`  //听牌索引
	TingNumber [static.MAX_CARD]int  `json:"tingnumber"` //听牌数量
	TingFanShu [static.MAX_CARD]int  `json:"tingfanshu"` //听牌蕃数
}

//游戏流程数据
type SportMetaJZHP struct {
	//游戏变量
	GameState int //游戏状态 GameState，参考ActionStep
	GameType  int //游戏类型 1表示普通模式2vs2，2表示吼牌模式1vs3
	DownTime  int // 权限停止时间

	ReplayRecord ZP_Replay_Record //回放记录
	ReWriteRec   byte             //是否重复写回放数据，每小局游戏开始时清理

	//玩家信息
	Banker     uint16 // 庄家
	Nextbanker uint16 //下一个庄家

	CurrentUser   uint16 //当前玩家	0--MAXPLAYER-1，在等待客户端响应时会置为无效值
	OutCardUser   uint16 //出牌用户	0--MAXPLAYER-1
	ResumeUser    uint16 //还原用户	0--MAXPLAYER-1
	ProvideUser   uint16 //供应用户	0--MAXPLAYER-1
	NoOutUser     uint16 //选择不出牌的用户 0--MAXPLAYER-1
	GuanStartUser uint16 //起牌后第一个检查观生的用户 0--MAXPLAYER-1

	//托管和离线数据
	TuoGuanPlayer  [TCGZ_MAX_PLAYER]bool //谁托管了？
	TrustCounts    [TCGZ_MAX_PLAYER]byte //玩家托管次数
	AutoCardCounts [TCGZ_MAX_PLAYER]byte //自动出牌的次数
	BreakCounts    [TCGZ_MAX_PLAYER]byte // 断线次数
	TrustOrder     [TCGZ_MAX_PLAYER]byte //玩家托管顺序

	//牌数据
	RepertoryCard [TCGZ_ALLCARD]byte                        // 库存扑克，所有牌
	LeftCardCount byte                                      // 剩余数目
	CardIndex     [TCGZ_MAX_PLAYER][TCGZ_MAX_INDEX_HUA]byte // 玩家分到的牌,包括花牌在内共27种
	OutCardCount  byte                                      // 总出牌扑克数目
	OutCardData   byte                                      // 出牌扑克
	ProvideCard   byte                                      // 供应扑克

	DiscardCard  [TCGZ_MAX_PLAYER][static.MAX_CARD]byte // 丢弃记录
	DiscardCount [TCGZ_MAX_PLAYER]byte                  // 丢弃数目

	SendCardData  byte // 发牌扑克
	SendCardCount byte // 发牌数目

	SendStatus         bool                         //发牌状态,发牌和出的牌，牌边框颜色不一样
	HuangZhuang        bool                         //是否荒庄
	WeaveHuxi          [TCGZ_MAX_PLAYER]int         // 每个人组合牌的总胡息
	UserGuanCards      [TCGZ_MAX_PLAYER][10][5]byte // 玩家观生的牌
	UserJianCards      [TCGZ_MAX_PLAYER][]byte      // 玩家捡的牌
	UserJianCardsCur   [TCGZ_MAX_PLAYER][]byte      // 玩家本轮捡的牌，这些牌本轮不能打
	UserGuanCardsCount [TCGZ_MAX_PLAYER]byte        // 玩家观生的牌数目
	UserPengCount      [TCGZ_MAX_PLAYER]byte        // 玩家碰的次数，荆州花牌只能碰2次
	TongInfo           [TCGZ_MAX_PLAYER]TagTongInfo // 玩家统的信息，玩家出牌统数可能会减少
	DispTongCnt        [TCGZ_MAX_PLAYER]int         // 客户端显示的玩家统的信息，玩家出牌统数不会减少，就是不能换统也不减少

	//积分数据
	ProvideUserCount [TCGZ_MAX_PLAYER]int // 点炮次数
	ChiHuUserCount   [TCGZ_MAX_PLAYER]int // 胡牌次数
	JiePaoUserCount  [TCGZ_MAX_PLAYER]int // 接炮次数
	ZiMoUserCount    [TCGZ_MAX_PLAYER]int // 自摸次数
	Total            [TCGZ_MAX_PLAYER]int // 总输赢，若干小局相加的金币
	MaxHUxi          [TCGZ_MAX_PLAYER]int // 单局最高胡数

	//底
	Base int // 底
	Spay int // 服务费

	//好友房
	GeziShu int  //个子数
	HuaShu  int  // 花数 ，10表示10个花，1表示溜花
	Piao    int  // 选漂:0不漂，100带漂，1-3定漂
	DuoHu   bool //true 一炮多响
	BeiShu  int  //倍数
	QuanHei int  //全黑倍数

	DianPaoPei         byte // 点炮包赔
	HunJiang           byte // 混江，0表示没有混
	KeChong            bool //true放铳玩家出分，另外一个人不出；false放铳和另一个人都出自己的分
	NoOut              bool //true不出牌（捏牌）
	FenType            int  //算分类型,数字型,0算胡数，1算坡数，2登庄
	Fleetime           int  //客户端传来的 游戏开始前离线踢人时间
	RoundOverAutoReady bool //小局结束自动准备
	DengZhuang         bool //登庄，false：胡牌玩家的下家当庄，true：胡牌玩家当庄。第一局随机，流局连庄
	OutCardDismissTime int  // 出牌时间 超时房间强制解散 -1不限制

	VecGameEnd  []static.Msg_S_ZP_TC_GameEnd    //记录每一局的结果
	VecGameData []static.CMD_S_ZP_TC_StatusPlay //记录每一局的结束时的桌面数据

	//控制变量
	VecChiHuCard   [TCGZ_MAX_PLAYER][]byte           //本轮弃吃胡的牌，用来控制大冶字牌的过庄
	VecChiCard     [TCGZ_MAX_PLAYER][]byte           //本轮弃吃的牌，用来控制大冶字牌的过庄
	VecPengCard    [TCGZ_MAX_PLAYER][]byte           //本轮弃碰的牌，用来控制大冶字牌的过庄
	Response       [TCGZ_MAX_PLAYER]bool             //响应标志
	UserAction     [TCGZ_MAX_PLAYER]int              //用户动作
	PerformAction  [TCGZ_MAX_PLAYER]int              //执行动作
	OperateCard    [TCGZ_MAX_PLAYER]byte             //操作扑克
	WeaveItemCount [TCGZ_MAX_PLAYER]byte             //组合数目
	WeaveItemArray [TCGZ_MAX_PLAYER][10]TagWeaveItem //组合扑克

	CMD_OperateCard [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard //多个人有牌权时，用于保存摆牌数据
	ChiHuCard       byte                                         //吃胡扑克
	ChiHuResult     [TCGZ_MAX_PLAYER]TagChiHuResult              //吃胡结果
	AutoOut         byte                                         //是否超时自动出牌

	UserReady [TCGZ_MAX_PLAYER]bool //玩家是否已经准备
	//时间数据
	PowerStartTime int64 //时间辅助变量，权限开始时间
	OperateTime    int64 // 出牌时间
	AutoOutTime    int64 // 托管出牌时间

	IsTongAction bool // 统牌阶段的托管需要加个状态，否则玩家响应的倒计时会和流程中的1秒间隔倒计时混淆

	//组件变量
	LastTime int64             // 用于断线重入 校准客户端操作时间
	Config   static.GameConfig //游戏配置

}

func (smj *SportMetaJZHP) GetRepertoryCards() (res [static.MAX_REPERTORY]byte) {
	for i := 0; i < int(smj.LeftCardCount); i++ {
		res[i] = smj.RepertoryCard[i]
	}
	return
}

func (smj *SportMetaJZHP) OnScoreOffset(total []float64) {
	for i, t := range total {
		var order ZP_Replay_Order
		order.R_ChairId = uint16(i)
		order.R_Opt = info2.E_GameScore
		order.UserScorePL = t
		smj.ReplayRecord.R_Orders = append(smj.ReplayRecord.R_Orders, order)
	}
}

//如果有些变量的值在每小局都要保留，建议使用resetForNext
func (smj *SportMetaJZHP) resetForNext() {
	//游戏变量
	smj.GameState = ZP_AS_NULL
	smj.GameType = meta2.GT_NULL
	smj.DownTime = 0

	smj.ReplayRecord.ReSet()

	//玩家信息
	smj.Banker = static.INVALID_CHAIR
	//smj.Nextbanker = public.INVALID_CHAIR
	smj.CurrentUser = static.INVALID_CHAIR
	smj.OutCardUser = static.INVALID_CHAIR
	smj.ResumeUser = static.INVALID_CHAIR
	smj.ProvideUser = static.INVALID_CHAIR
	smj.NoOutUser = static.INVALID_CHAIR
	smj.GuanStartUser = static.INVALID_CHAIR

	//托管和离线数据
	//smj.TuoGuanPlayer = [TCGZ_MAX_PLAYER]bool{}
	smj.TrustCounts = [TCGZ_MAX_PLAYER]byte{}
	smj.AutoCardCounts = [TCGZ_MAX_PLAYER]byte{}
	smj.BreakCounts = [TCGZ_MAX_PLAYER]byte{}
	smj.TrustOrder = [TCGZ_MAX_PLAYER]byte{}

	//牌数据
	//smj.m_allCards = [TS_ALLCARD]byte{}
	for j, _ := range smj.CardIndex {
		for k, _ := range smj.CardIndex[j] {
			smj.CardIndex[j][k] = 0
		}
	}
	for j, _ := range smj.DiscardCard {
		for k, _ := range smj.DiscardCard[j] {
			smj.DiscardCard[j][k] = 0
		}
	}

	smj.DiscardCount = [TCGZ_MAX_PLAYER]byte{}
	smj.WeaveHuxi = [TCGZ_MAX_PLAYER]int{}
	smj.LeftCardCount = 0
	smj.OutCardCount = 0
	smj.OutCardData = 0
	smj.ProvideCard = 0
	smj.SendCardData = 0
	smj.SendCardCount = 0
	smj.SendStatus = false
	smj.HuangZhuang = false
	smj.UserGuanCardsCount = [TCGZ_MAX_PLAYER]byte{}
	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		smj.UserPengCount[i] = 0
		smj.UserJianCards[i] = []byte{}
		smj.UserJianCardsCur[i] = []byte{}
		for j := 0; j < 10; j++ {
			smj.UserGuanCards[i][j] = [5]byte{}
		}
	}
	smj.TongInfo = [TCGZ_MAX_PLAYER]TagTongInfo{}
	smj.DispTongCnt = [TCGZ_MAX_PLAYER]int{}

	//控制变量
	for j, _ := range smj.VecChiHuCard {
		smj.VecChiHuCard[j] = []byte{}
		smj.VecChiCard[j] = []byte{}
		smj.VecPengCard[j] = []byte{}
	}
	smj.Response = [TCGZ_MAX_PLAYER]bool{}
	smj.UserAction = [TCGZ_MAX_PLAYER]int{}
	smj.PerformAction = [TCGZ_MAX_PLAYER]int{}
	smj.OperateCard = [TCGZ_MAX_PLAYER]byte{}
	smj.WeaveItemCount = [TCGZ_MAX_PLAYER]byte{}
	smj.WeaveItemArray = [TCGZ_MAX_PLAYER][10]TagWeaveItem{}

	smj.CMD_OperateCard = [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard{}
	smj.ChiHuCard = 0
	smj.AutoOut = 0
	smj.ChiHuResult = [TCGZ_MAX_PLAYER]TagChiHuResult{}
	smj.PowerStartTime = 0
	smj.IsTongAction = false
	smj.UserReady = [TCGZ_MAX_PLAYER]bool{}

	//状态变量
	smj.LastTime = 0
}

//Reset all
func (smj *SportMetaJZHP) reset() {
	smj.resetForNext()

	smj.Nextbanker = static.INVALID_CHAIR
	smj.VecGameEnd = []static.Msg_S_ZP_TC_GameEnd{}
	smj.VecGameData = []static.CMD_S_ZP_TC_StatusPlay{}
	smj.ReWriteRec = 0

	smj.TuoGuanPlayer = [TCGZ_MAX_PLAYER]bool{}

	//底
	smj.Base = 1
	smj.Spay = 0

	//好友房
	smj.GeziShu = 7
	smj.HuaShu = 10
	smj.Piao = 0
	smj.DuoHu = false
	smj.BeiShu = 1
	smj.QuanHei = 2
	smj.HunJiang = 0
	smj.DianPaoPei = 0
	smj.KeChong = false
	smj.NoOut = false
	smj.FenType = 0
	smj.Fleetime = 0
	smj.RoundOverAutoReady = false
	smj.DengZhuang = false
	smj.OutCardDismissTime = 0

	//积分数据
	smj.Total = [TCGZ_MAX_PLAYER]int{}
	smj.ProvideUserCount = [TCGZ_MAX_PLAYER]int{}
	smj.ChiHuUserCount = [TCGZ_MAX_PLAYER]int{}
	smj.JiePaoUserCount = [TCGZ_MAX_PLAYER]int{}
	smj.ZiMoUserCount = [TCGZ_MAX_PLAYER]int{}
	smj.MaxHUxi = [TCGZ_MAX_PLAYER]int{}

	smj.OperateTime = 15
	smj.AutoOutTime = 1

}

func (smj *SportMetaJZHP) getCardDataByStr(cbCardStr string) (carddataHex byte, carddata byte) {
	var card byte
	defer func() {
		if card != static.INVALID_BYTE {
			carddataHex = card + 1 //十六进制的牌
			carddata = card + 1    //十进制的牌
		} else {
			carddataHex = static.INVALID_BYTE
			carddata = static.INVALID_BYTE
		}
	}()

	for k, v := range m_strZPCardsMessageUP {
		if v == cbCardStr {
			card = byte(k)
			return
		}
	}
	for k, v := range m_strZPCardsMessageLW {
		if v == cbCardStr {
			card = byte(k)
			return
		}
	}

	card = static.INVALID_BYTE
	return
}
