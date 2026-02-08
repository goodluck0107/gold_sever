package metadata

import (
	"github.com/open-source/game/chess.git/pkg/static"
)

//做牌数据
type CardConfig struct {
	IsAble             int      `json:"IsAble"`             //! 是否开启做牌
	UserCards          []string `json:"UserCards"`          //! 玩家手牌
	RepertoryCard      string   `json:"RepertoryCard"`      //! 牌堆牌
	RepertoryCardCount int      `json:"RepertoryCardCount"` //! 牌堆牌数量
	BankerUserSeatId   int      `json:"BankerUserSeatId"`   //! 庄家位置，-1不固定庄家
	LeftCardCount      byte     `json:"LeftCardCount"`      //! 测试流局用的
	ShengJiMainOrder0  int      `json:"ShengJiMainOrder0"`  //! 升级游戏的双方级数，必须双方同时配或同时不配置(做牌文件没有这一行或配置为双-1表示不配置级数)
	ShengJiMainOrder1  int      `json:"ShengJiMainOrder1"`  //! 升级游戏的双方级数，必须双方同时配或同时不配置(做牌文件没有这一行或配置为双-1表示不配置级数)
	Max                int      `json:"Max"`                //! 设置总局数
	MagicCard          byte     `json:"magiccard"`          //癞子牌
}

//通山打拱牌库文件
type TSCardConfig struct {
	IsAble  int `json:"IsAble"`  //! 是否开启做牌
	ReadPai int `json:"ReadPai"` //! 牌库版本，版本和保存的不一致，就要重读牌库,0以上的值有效
	FaPaiMs int `json:"FaPaiMs"` //! 发牌模式，0表示只有首牌随机，1表示所有牌都需要随机
}

//双开算分算番统计
type Game_mj_fan_score struct {
	FanNum     [4]int
	Score      [4]int
	ScoreFloat [4]float32
	DianNum    [4]float32
	Ding       [4]int
}

func (self *Game_mj_fan_score) Reset() {
	for i := 0; i < 4; i++ {
		self.FanNum[i] = 0
		self.Score[i] = 0
		self.DianNum[i] = 0
		self.Ding[i] = static.DING_NULL
	}
}

func (self *Game_mj_fan_score) AddFan(seatid uint16, fannum int) {
	if seatid >= 0 && seatid < 4 && fannum > 0 {
		self.FanNum[seatid] += fannum
	}
}

// 记录上局游戏信息
type Game_mj_lastround_info struct {
	LastRoundJingDingUser uint16
	LastRoundScore        [4]int
	LastRoundMagicCard    byte
	LastCun               byte //麻城麻将特有
}

func (self *Game_mj_lastround_info) ReSet() {
	self.LastRoundJingDingUser = static.INVALID_CHAIR
	self.LastRoundMagicCard = static.INVALID_BYTE
	for i := 0; i < 4; i++ {
		self.LastRoundScore[i] = 0
	}
}

type Game_mj_genzhuang_info struct {
	M_cbGenzhuangCard     byte   `json:"genzhuangcard"`     //跟庄的牌
	M_cbGenzhuangNum      int    `json:"genzhuangnum"`      //每个连续跟庄玩家数量
	M_cbGenzhuangLastUser uint16 `json:"genzhuanglastuser"` //上一个跟庄的玩家
	M_cbGenzhuangSuccess  bool   `json:"genzhuangsuccess"`  //是否跟庄成功
}

func (self *Game_mj_genzhuang_info) ReSet() {
	self.M_cbGenzhuangCard = static.INVALID_BYTE
	self.M_cbGenzhuangNum = 0
	self.M_cbGenzhuangLastUser = static.INVALID_CHAIR
	self.M_cbGenzhuangSuccess = false
}

//20181029 苏大强 普通的胡法有258将也有乱将 不排除指定非万饼筒的牌为将  规则
//简陋一点，目前万条筒 最好是这样，当指定是258的时候，结构就根据牌库规则（全牌，去万或去风）转换成（258万，258条，258筒）
type Eyecards_info struct {
	EyeisNotFeng bool   //将牌是不是风牌（258不能是风）
	Eyecards     []byte //将牌 不限数目 无花色
}

type ContractorQysUserItem struct {
	Qysuser uint16 `json:"qysuser"` //清一色陪包用户
	//清一色陪包用户(因为谁开三口了要陪包，可能会开四口，第四口不是清一色的话因为这个玩家的陪包要取消)
	QysReasonChair uint16 `json:"qysreasonchair"`
}

type Xiayu_Record_info struct {
	SeatId                uint16                   `json:"seatid"`           //底分
	XiayuQysKind          bool                     `json:"xiayuqyskind"`     //是否清一色承包
	XiayuQysColor         int                      `json:"xiayuqyscolor"`    //存放清一色下雨包牌色值
	ContractorQysUser     uint16                   `json:"xiayuqysuser"`     //承包清一色用户
	WContractorQysUser    [4]uint16                `json:"wxiayuqysuser"`    //承包清一色用户
	WhanContractorQysUser [4]ContractorQysUserItem `json:"whanxiayuqysuser"` //武汉麻将陪包记录
	XiayuJysKind          bool                     `json:"xiayujyskind"`     //是否将一色承包
	ContractorJysUser     uint16                   `json:"xiayujysuser"`     //承包将一色用户
	WContractorJysUser    [4]uint16                `json:"wxiayujysuser"`    //承包将一色用户
}

func (self *Xiayu_Record_info) ReSet() {
	self.SeatId = static.INVALID_CHAIR
	self.XiayuQysKind = false
	self.XiayuQysColor = -1
	self.WContractorQysUser = [4]uint16{static.INVALID_CHAIR, static.INVALID_CHAIR, static.INVALID_CHAIR, static.INVALID_CHAIR}
	self.WhanContractorQysUser = [4]ContractorQysUserItem{
		ContractorQysUserItem{static.INVALID_CHAIR, static.INVALID_CHAIR},
		ContractorQysUserItem{static.INVALID_CHAIR, static.INVALID_CHAIR},
		ContractorQysUserItem{static.INVALID_CHAIR, static.INVALID_CHAIR},
		ContractorQysUserItem{static.INVALID_CHAIR, static.INVALID_CHAIR}}
	self.ContractorQysUser = static.INVALID_CHAIR
	self.XiayuJysKind = false
	self.WContractorJysUser = [4]uint16{static.INVALID_CHAIR, static.INVALID_CHAIR, static.INVALID_CHAIR, static.INVALID_CHAIR}
	self.ContractorJysUser = static.INVALID_CHAIR
}

func (self *Xiayu_Record_info) EnterXiayuQys(SeatId uint16, QysColor int) {
	self.SeatId = SeatId
	self.XiayuQysKind = true
	self.XiayuQysColor = QysColor
}

func (self *Xiayu_Record_info) EnterXiayuJys(SeatId uint16) {
	self.SeatId = SeatId
	self.XiayuJysKind = true
}

//操作分
type Msg_S_OperateScore_K5X struct {
	OperateUser uint16     `json:"operateuser"`  //操作用户
	OperateType uint16     `json:"operatetype"`  //操作类型
	GameScore   [4]int     `json:"gamescore"`    //最新总分
	GameVitamin [4]float64 `json:"game_vitamin"` //最新疲劳值信息
	ScoreOffset [4]int     `json:"scoreoffset"`  //分数变化量
}

type ShowCard struct {
	BIsShowCard bool                   //	是否亮牌
	CbTingCard  [20]byte               //	听牌
	CbAnPuCard  [static.MAX_COUNT]byte //	暗铺的牌
	CbLiangCard [13]byte               //	亮出的牌
}

func (self *ShowCard) Reset() {
	self.BIsShowCard = false
	self.CbTingCard = [20]byte{}
	self.CbAnPuCard = [static.MAX_COUNT]byte{}
	self.CbLiangCard = [13]byte{}
}

//记牌器数据结构
type CardRecorder struct {
	Point int `json:"point"` //牌值
	Num   int `json:"num"`   //数量
}

type K5x_Replay_Order struct {
	Replay_Order
	LiangPai string `json:"liangpai"` //亮牌
}

type K5x_Replay_Record struct {
	RecordHandCard [static.MAX_CHAIR][static.MAX_COUNT]byte `json:"handcard"`      // 用户最初手上的牌
	VecOrder       []K5x_Replay_Order                       `json:"vecorder"`      // 用户操作
	BigHuKind      byte                                     `json:"bighukind"`     // 标记大胡类型
	WBigHuKind     [4]byte                                  `json:"wbighukind"`    // 一炮多响大胡类型
	ProvideUser    uint16                                   `json:"provideuser"`   // 点炮用户
	PiziCard       byte                                     `json:"pizicard"`      // 癞子皮
	LeftCardCount  byte                                     `json:"leftcardcount"` // 发完牌后剩余数目
	Score          [static.MAX_CHAIR]int                    `json:"score"`         // 游戏积分
	UVitamin       map[int64]float64                        `json:"u_vitamin"`     // 玩家起始疲劳值
	Fengquan       byte                                     `json:"fengquan"`      // 风圈
}

//游戏记录重置
func (self *K5x_Replay_Record) Reset() {

	for i := 0; i < static.MAX_CHAIR; i++ {
		for j := 0; j < static.MAX_COUNT; j++ {
			self.RecordHandCard[i][j] = 0
		}
	}

	for i := 0; i < len(self.Score); i++ {
		self.Score[i] = 0
	}

	self.UVitamin = make(map[int64]float64)
	self.VecOrder = make([]K5x_Replay_Order, 0, 30)
	self.BigHuKind = 2 //0用来表示将一色了
	self.ProvideUser = 9
	self.PiziCard = 0
	self.Fengquan = 0
	self.WBigHuKind = [4]byte{}
}
