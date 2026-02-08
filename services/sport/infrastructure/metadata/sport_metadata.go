package metadata

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	"math/rand"
	"sync"
	"time"
)

const (
	//游戏 I D
	MAX_PLAYER = 4 //游戏人数
	MAX_HORSE  = 8
	//MAX_PLAYER_INDEX = 8//最大人数
)

// 聊天类型
const (
	CHAT_COLOR_MOFA    = 4
	CHAT_COLOR_MOFA_10 = 5
)

// 聊天消耗
const (
	CHAT_COLOR_MOFA_COST    = 100
	CHAT_COLOR_MOFA_10_COST = 1000
)

// 风圈，滁州玩法
const (
	TYPE_DIRECATION_EAST  = 1 // 东风圈
	TYPE_DIRECATION_SOUTH = 2 // 南风圈
	TYPE_DIRECATION_WEST  = 3 // 西风圈
	TYPE_DIRECATION_NORTH = 4 // 北风圈
)

type Replay_Order struct {
	Chair_id     uint16  `json:"id"`           // 玩家
	Operation    int     `json:"operation"`    // 记录类型
	Value        []int   `json:"value"`        // 出牌
	UserScore    float64 `json:"user_score"`   // 玩家分数
	OperateStr   string  `json:"operatestr"`   //其它自定义操作
	TimeSecAt    int     `json:"timesecat"`    // 第几秒发生的
	OpreateRight uint64  `json:"opreateright"` // 牌权（汉川搓虾子需要把胡的牌权组合也写到回放里面去）
}

type Replay_Record struct {
	RecordHandCard [static.MAX_CHAIR][static.MAX_COUNT]byte `json:"handcard"`       // 用户最初手上的牌
	VecOrder       []Replay_Order                           `json:"vecorder"`       // 用户操作
	BigHuKind      byte                                     `json:"bighukind"`      // 标记大胡类型
	BigHuKindArray []byte                                   `json:"bighukindarray"` // 标记大胡类型(存在一炮多响的玩法的回放) add by zwj
	ProvideUser    uint16                                   `json:"provideuser"`    // 点炮用户
	PiziCard       byte                                     `json:"pizicard"`       // 癞子皮
	LeftCardCount  byte                                     `json:"leftcardcount"`  // 发完牌后剩余数目
	Score          [static.MAX_CHAIR]int                    `json:"score"`          // 游戏积分
	UVitamin       map[int64]float64                        `json:"u_vitamin"`      // 玩家起始疲劳值
	Fengquan       byte                                     `json:"fengquan"`       // 风圈
	LeftCard       byte                                     `json:"lastcard"`       // 最后一张牌
	WBigHuKind     [4]byte                                  `json:"wbighukind"`     // 一炮多响大胡类型
	EndInfo        *static.Msg_S_GameEnd                    `json:"end_info"`       //小结算详情
}

// 牌堆数据
type CardLeftArrayMeta struct {
	MaxCount  int     //最大牌数量
	CardArray []int   //牌堆,0代表有牌，-1代表没牌
	OffSet    int     //牌堆前面索引
	EndOffSet int     //牌堆后面索引
	Random    [2]byte //甩子
	Seat      uint16
	Kaikou    int
}

// 游戏记录重置
func (self *CardLeftArrayMeta) Reset() {

	//self.CardArray = make([]int, 0, 0)
	self.OffSet = 0
	self.EndOffSet = 0
	self.MaxCount = 0
	self.Seat = static.INVALID_CHAIR
	self.Kaikou = 0
}

func (self *CardLeftArrayMeta) IsValid() bool {
	if self.MaxCount == 0 || self.OffSet >= self.MaxCount || self.EndOffSet < 0 {
		return false
	}
	return true
}

// 游戏记录重置
func (self *Replay_Record) Reset() {

	for i := 0; i < static.MAX_CHAIR; i++ {
		for j := 0; j < static.MAX_COUNT; j++ {
			self.RecordHandCard[i][j] = 0
		}
	}

	for i := 0; i < len(self.Score); i++ {
		self.Score[i] = 0
	}

	self.UVitamin = make(map[int64]float64)
	self.VecOrder = make([]Replay_Order, 0, 30)
	self.BigHuKind = 2 //0用来表示将一色了
	self.ProvideUser = 9
	self.PiziCard = 0
	self.Fengquan = 0
	self.BigHuKindArray = self.BigHuKindArray[0:0]
	self.WBigHuKind = [4]byte{}
	self.EndInfo = nil
}

type GameMeta = Metadata

// 流程数据
type Metadata struct {
	//游戏变量
	//b_HDAction   bool          //是否可以赌海底
	HaveGangCard bool          //是否是杠后发牌
	ReplayRecord Replay_Record //回放记录

	//出牌信息
	PiZiCard     byte   //皮子牌值
	PiZiCards    []byte //多个皮子存放
	MagicCard    byte   //赖子牌值
	OutCardUser  uint16 //出牌用户
	OutCardData  byte   //出牌扑克
	OutCardCount byte   //出牌数目
	HasHDGang    bool   //咸宁麻将允许在剩余8张牌时（开始海底前），再杠一次，杠完后，从剩余7张牌时算开始海底
	PiZiCardTmp  byte   //皮子，翻出来的没有做过处理的牌

	//发牌信息
	SendCardData       byte              //发牌扑克
	SendCardCount      byte              //发牌数目
	LeftBu             byte              //剩下的补牌数目
	LeftCardCount      byte              //剩余数目
	RepertoryCard      []byte            //库存扑克
	RepertoryCardArray CardLeftArrayMeta //牌堆

	//运行变量
	ResumeUser       uint16 //还原用户
	CurrentUser      uint16 //当前用户
	ProvideUser      uint16 //供应用户
	ProvideCard      byte   //供应扑克
	LastSendCardUser uint16 //最近一次发牌的接收者
	LastOutCardUser  uint16 //最近一次出牌人
	//状态变量
	HaiDiUserCount  byte //参与海底人数
	GangFlower      bool //杠上开花状态
	SendStatus      bool //发牌状态
	QiangGangStatus bool //抢杆状态
	GangHotStatus   bool //热统状态
	MingGangStatus  bool //明杆状态
	IsHaiDi         bool //海底状态
	IsTianDi        bool //天地胡
	HaveHuangZhuang bool //是否荒庄
	//m_vecXiaPao       []static.Msg_C_Xiapao  //记录每一局下跑
	PayPaoStatus        bool                         //记录是否在买跑的状态
	XuanPiaoStatus      bool                         //记录是否在选漂的状态
	ExchangeThreeStatus bool                         //记录是否在换三张的状态
	VecGameEnd          []static.Msg_S_GameEnd       //记录每一局的结果
	VecGameData         []static.CMD_S_StatusPlay    //记录每一局的结束时的桌面数据
	VecGameDataAllP     [4][]static.CMD_S_StatusPlay //记录每一局的结束时所有人的桌面数据
	GameStartForXiapao  bool                         //转门为下跑做的特殊设计
	TianHuStatus        bool                         //天胡状态

	//结束信息
	ChiHuCard byte //吃胡扑克

	//组件变量
	//	bool							m_bLogAble bool;								//日志是否开启
	HaveCellScore  bool   //底分是否除以10倍
	TastRateshell  int    //贝壳任务倍率
	TimeGameRecord string //记录当前日志的日期
	NameGameRecord string //记录文件名
	GameRecordNum  string //记录编号
	//	CLog                            m_mylog;								//log文件标识
	LastTime int64             // 用于断线重入 校准客户端操作时间
	Config   static.GameConfig //游戏配置

	//20181031 苏大强 先扔这里
	Eyecards Eyecards_info // 将牌结构
	//NoFeng      bool                //无风（东南西北） 发牌器属性 规则的属性
	FristFlowBanker bool //第一局流局是不是要换庄家 fristFlowChangeBanker  公共
	BirdCount       int  //抓鸟（买马）个数 基本要公用了
	//CbBirdData   [MAX_PLAYER][8]byte `json:"CbBirdData"` //抓鸟数据
	IsLocalDebug     bool
	HasHDGangCount   int  //海底杠
	HaveMagicCardOut bool //有癞子打出，只能自摸

	LastGangKind  byte   // 最后一次杠的类型
	LastGangIndex byte   // 最后一次杠的下标
	KzhDatas      [4]int //是否是有卡字胡  石首捱晃
	State         int    // 游戏状态 GameState，1表示吼牌阶段，2表示打牌阶段
	TheGameTime   int64  // 游戏时间设置

	LastGangScore [4]int //最近一次的杠分(主要用来记录抢杠胡的杠分,江陵玩法抢杠胡杠分不算)
	DingPiaoNum   int    // 定漂是几
	CurHaiDiCount int    // 本局当前应该进入海底捞的牌数量(海底捞数量可变情况使用)

	//m_BaoQingType [4]int //用户报清情况

	FengQuan byte //风圈
	QuanFeng byte //圈风
	//20190918 苏大强 9秒场还是有牌错的情况，加个锁吧
	OperateMutex sync.Mutex
	//20191010 苏大强 恩施添加
	//恩施 抬庄
	TaiZhuangInfo TaiZhuangInfoStruct
	//恩施 首杠发牌 第一次杠的时候把痞子牌发给玩家
	FirstGangGard byte
	//恩施首杠发痞子，如果能养，就不管，不能养就要自动杠出去
	YangPi bool
	//20191112 苏大强 阜阳杠番打出牌的记录，记牌器。。。。
	OutCardIndex [static.MAX_INDEX]byte //记牌器
	//
	TuoGuanPlayer [static.MAX_PLAYER_4P]bool //谁托管了？
	TrustCounts   [static.MAX_PLAYER_4P]byte //玩家托管次数

	CurPowerSeat           uint16                    // 超时罚分的当前权限玩家
	CurPunishSeat          uint16                    // 超时罚分的当前罚时玩家
	CurPeriod              int                       // 超时罚分的当前阶段
	PunishStartTime        int64                     // 超时罚分的度过时间
	HasAction              [4]bool                   // 超时罚分的动作标志
	PunishCount            [4]int                    // 超时罚分的次数计数
	PlayerPunishScore      [4]int                    // 超时罚的分
	PlayerChihuCards       [4]UserChihuCards         //玩家吃胡记录
	GangAfterOperateStatus bool                      //玩家吃碰操作之后还有杠权状态
	MaxPiao                [static.MAX_PLAYER_4P]int // 每局最高漂数

	Tuo3       bool                      //连续托管3局解散
	TrustCnt   [static.MAX_PLAYER_4P]int // 每个人连续托管次数
	TrustJuShu int                       //托管局数 不限制0
	TrustLimit int                       //托管限制 1暂停  2解散
	UserReady  [4]bool                   //玩家是否已经准备
	TotalScore [4]int                    //玩家累计分数
	UserAction [4]byte                   //20210120 苏大强 累计玩法用的，多牌权记录
	HuList     []int                     //20210129 苏大强 宜昌血流记录顺序胡牌玩家
	HuUser     [4]bool                   //20210129 苏大强 宜昌血流记录已经胡过的玩家
}

func (md *Metadata) GetRepertoryCards() (res [static.MAX_REPERTORY]byte) {
	for i := 0; i < int(md.LeftCardCount); i++ {
		res[i] = md.RepertoryCard[i]
	}
	return
}

func (md *Metadata) FindCardIndexInRepertoryCards(cardData byte) byte {
	index := static.INVALID_BYTE
	for i := md.LeftCardCount - 1; true; i-- {
		if i < 0 || i == static.INVALID_BYTE {
			break
		}
		if md.RepertoryCard[i] == cardData {
			index = i
			break
		}
	}
	return index
}

// 20191010 抬庄 恩施麻将
type TaiZhuangInfoStruct struct {
	CanTZ    bool //是否抬庄
	TZStatus bool //是否抬庄状态（庄家一轮取消）
	TZNumber int  //20200421 苏大强 抬庄次数
	TZCard   byte //抬庄的牌
	TZNum    int  //记录抬庄状态下打出这章牌的个数，一圈下来，个数必须和玩家数相同
	TZFanshu byte //抬庄番数
}

func (md *Metadata) Reset() {
	//游戏变量
	//md.b_HDAction = true
	md.HaveGangCard = false

	//出牌信息
	md.OutCardData = 0
	md.OutCardCount = 0
	md.OutCardUser = static.INVALID_CHAIR

	//发牌信息
	md.SendCardData = 0
	md.SendCardCount = 0
	md.LeftBu = 0
	md.HasHDGang = false

	//运行变量
	md.ProvideCard = 0
	md.ResumeUser = static.INVALID_CHAIR
	md.CurrentUser = static.INVALID_CHAIR
	md.ProvideUser = static.INVALID_CHAIR
	md.PiZiCard = 0x00
	md.LastOutCardUser = uint16(static.INVALID_CHAIR)

	//状态变量
	md.GangFlower = false
	md.SendStatus = false
	md.QiangGangStatus = false
	md.MingGangStatus = false
	md.IsHaiDi = false
	md.IsTianDi = false
	md.HaveHuangZhuang = false
	md.HaiDiUserCount = 0
	//20190911 热冲标准
	md.GangHotStatus = false
	for k, _ := range md.RepertoryCard {
		md.RepertoryCard[k] = 0
	}
	md.TrustJuShu = 0
	md.TrustLimit = 0

	//md.DiscardCard = [MAX_PLAYER][55]byte{}
	//结束信息
	md.ChiHuCard = 0
	md.HaveMagicCardOut = false
	md.HasHDGangCount = 0 //海底杠
	md.ReplayRecord.Reset()
	md.RepertoryCardArray.Reset()
	md.OutCardIndex = [static.MAX_INDEX]byte{}
	md.TuoGuanPlayer = [static.MAX_PLAYER_4P]bool{}
	md.TrustCounts = [static.MAX_PLAYER_4P]byte{}
	md.PlayerChihuCards = [4]UserChihuCards{}
	md.GangAfterOperateStatus = false
	md.UserReady = [4]bool{}
	md.TotalScore = [4]int{}
	md.UserAction = [4]byte{0}
	md.HuList = []int{}
	md.HuUser = [4]bool{false, false, false, false}
}

// 潜江麻将有少三张功能，发牌数不是固定13张，是可以配置的。重定义一个函数
func (md *Metadata) CreateLeftCardArrayBySendCount(_userCount int, _cardCount int, _pizi bool, sendCount int) {

	md.RepertoryCardArray.CardArray = make([]int, _cardCount, _cardCount)
	md.RepertoryCardArray.OffSet = 0
	md.RepertoryCardArray.EndOffSet = _cardCount - 1
	md.RepertoryCardArray.MaxCount = _cardCount

	_sendCount := _userCount*sendCount + 1

	if _pizi {
		_sendCount++
	}

	for i := 0; i < _sendCount && i < _cardCount; i++ { //庄家多起一张
		md.RepertoryCardArray.CardArray[_cardCount-1-i] = -1
	}

	md.RepertoryCardArray.EndOffSet -= _sendCount

	md.RepertoryCardArray.Seat = uint16(rand.Intn(_userCount))
	//掷骰子,两个骰子
	for i := 0; i < 2; i++ {
		md.RepertoryCardArray.Random[i] = byte(rand.Intn(6) + 1)
	}
}

func (md *Metadata) CreateLeftCardArray(_userCount int, _cardCount int, _pizi bool) {

	md.RepertoryCardArray.CardArray = make([]int, _cardCount, _cardCount)
	md.RepertoryCardArray.OffSet = 0
	md.RepertoryCardArray.EndOffSet = _cardCount - 1
	md.RepertoryCardArray.MaxCount = _cardCount

	_sendCount := _userCount*(static.MAX_COUNT-1) + 1

	if _pizi {
		_sendCount++
	}

	for i := 0; i < _sendCount && i < _cardCount; i++ { //庄家多起一张
		md.RepertoryCardArray.CardArray[_cardCount-1-i] = -1
	}

	md.RepertoryCardArray.EndOffSet -= _sendCount

	md.RepertoryCardArray.Seat = uint16(rand.Intn(_userCount))
	//掷骰子,两个骰子
	for i := 0; i < 2; i++ {
		md.RepertoryCardArray.Random[i] = byte(rand.Intn(6) + 1)
	}
}

// 设置牌堆变化
func (md *Metadata) SetLeftCardArray() {

	if !md.RepertoryCardArray.IsValid() {
		return
	}
	if md.GangFlower || md.HaveGangCard {
		md.RepertoryCardArray.CardArray[md.RepertoryCardArray.OffSet] = -1
		md.RepertoryCardArray.OffSet++
	} else {
		md.RepertoryCardArray.CardArray[md.RepertoryCardArray.EndOffSet] = -1
		md.RepertoryCardArray.EndOffSet--
	}
}

func (md *Metadata) SetTime(t int64, state int) {
	if t <= 0 {
		md.TheGameTime = 0
	} else {
		md.TheGameTime = time.Now().Unix() + t
	}
	md.State = state
}

func (md *Metadata) OnScoreOffset(total []float64) {
	for i, t := range total {
		var order Replay_Order
		order.Chair_id = uint16(i)
		order.Operation = info2.E_GameScore
		order.UserScore = t
		md.ReplayRecord.VecOrder = append(md.ReplayRecord.VecOrder, order)
	}
}

// 20191113 苏大强 阜阳杠番里面包牌的要看是不是第一章使用-------------------
func (md *Metadata) RecordOutCard(card byte) error {
	index, err := mahlib2.CardToIndex(card)
	if err != nil {
		return err
	}
	if md.OutCardIndex[index] >= 4 {
		return errors.New(fmt.Sprintf("记牌器（%s）的记录已经（%d），再加越界", mahlib2.G_CardAnother[index], md.OutCardIndex[index]))
	}
	md.OutCardIndex[index]++
	return nil
}

// 获取打出牌的记录数
func (md *Metadata) GetOutCardRecordNum(card byte) (byte, error) {
	index, err := mahlib2.CardToIndex(card)
	if err != nil {
		return static.INVALID_BYTE, err
	}
	return md.OutCardIndex[index], nil
}

// -----------------------------------------
type UserChihuCards struct {
	HasChihu   bool                       //玩家是否胡过
	ChihuInfo  []*static.ChihuDetail_xlch //玩家吃胡牌
	TotalScore int                        //玩家总分
}

func (self *UserChihuCards) Reset() {
	self.HasChihu = false
	self.ChihuInfo = []*static.ChihuDetail_xlch{}
}

func (self *UserChihuCards) AddChihuInfo(info static.ChihuDetail_xlch) {
	if !self.HasChihu {
		self.HasChihu = (info.ChihuType == static.HuType_JiePao) || (info.ChihuType == static.HuType_QiangGang) || (info.ChihuType == static.HuType_Zimo)
	}

	detail := new(static.ChihuDetail_xlch)
	detail.ChihuCard = info.ChihuCard
	detail.ChihuKind = info.ChihuKind
	detail.ChihuType = info.ChihuType
	detail.ChihuProvider = info.ChihuProvider
	detail.ChihuFan = info.ChihuFan
	detail.ChihuScore = info.ChihuScore
	detail.ChihuUser = info.ChihuUser
	detail.ChihuLeaderType = info.ChihuLeaderType
	self.ChihuInfo = append(self.ChihuInfo, detail)
	self.TotalScore += info.ChihuScore
}
