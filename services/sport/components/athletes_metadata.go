package components

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
)

//玩家操作数据
type PlayerMeta struct {
	Timer                 Time                   //时间计时器
	CardIndex             [static.MAX_INDEX]byte `json:"cardindex"`             //用户扑克
	DiscardCard           []byte                 `json:"discardcard"`           //丢弃记录
	OutCardRecord         []byte                 `json:"outcardrecord"`         //玩家出牌记录(玩家当局所有打出去的牌,包含被别人吃碰杠胡的牌)
	DiscardCardClass      []byte                 `json:"discardcardclass"`      //丢弃类型
	ShowGang              byte                   `json:"showgang"`              //明杠次数
	DianGang              byte                   `json:"diangang"`              //点杠次数
	XuGang                byte                   `json:"xugang"`                //蓄杠
	DianXuGang            byte                   `json:"dianxugang"`            //点蓄杠
	HidGang               byte                   `json:"hidgang"`               //暗杠次数
	MagicCardGang         byte                   `json:"magiccardgang"`         //赖子杠
	MagicCardChi          byte                   `json:"magiccardchi"`          //吃赖子
	MagicCardOut          byte                   `json:"magiccardout"`          //赖子作普通牌打出，也叫丢赖子
	HongZhongGang         byte                   `json:"hongzhonggang"`         //红中杠
	GangUserCount         byte                   `json:"gangcount"`             //杠次数
	AnGangUserCount       byte                   `json:"angangcount"`           //暗杠次数
	FaCaiGang             byte                   `json:"facaigang"`             //发财杠
	FengGang              byte                   `json:"fenggang"`              //风杠
	JiangGang             byte                   `json:"jianggang"`             //将杠
	PiZiGangCount         byte                   `json:"GangPizi"`              //痞子杠
	PiZiCardOut           byte                   `json:"pizicardout"`           //丢痞子   20191010 恩施麻将
	QiangFailCount        byte                   `json:"qiangfailcount"`        //抢错次数
	JiangPeng             byte                   `json:"jiangpeng"`             //将碰
	FengPeng              byte                   `json:"fengpeng"`              //风碰
	QiangScore            int                    `json:"qiangscore"`            //抢暗杠输赢的分
	HDCard                static.TagCard         `json:"hdcard"`                //海底牌记录
	CheckTimeOut          int64                  `json:"checktimeout"`          ///<游戏中逻辑踢除时间累计
	EndAutoReadyTimeOut   int64                  `json:"endautoreadytimeout"`   //小结算自动准备倒计时开始时间
	LimitedTime           int64                  `json:"limitedtime"`           ///20200227 苏大强 限定托管时间 恩施用来记录超过10秒的总时间
	ClickTime             int64                  `json:"clicktime"`             //20210109 苏大强 用户点击的时间 可能不用
	RecordbeginTime       int64                  `json:"recordbegintime"`       //20210109 苏大强 开始记录累计模式的时间
	DelayerSetBeginTime   int64                  `json:"delayersetbeigntime"`   //20210304 设置累计计时器时的时间 断线重连用
	DelayerSetTime        int64                  `json:"delayersettime"`        //20210304 设置累计计时器时的到期时间 给断线重连用
	EnjoinChiHu           bool                   `json:"enjoinchihu"`           //禁止吃胡
	EnjoinChiPeng         bool                   `json:"enjoinchipeng"`         //禁止吃碰
	HaveGang              bool                   `json:"havegang"`              //杠过牌
	Trustee               bool                   `json:"trustee"`               //是否托管
	HasTrustee            bool                   `json:"has_trustee"`           //是否托管
	TrusteeNum            uint64                 `json:"trusteenum"`            //20191219 苏大强 小局托管次数
	UserPaoReady          bool                   `json:"userpaoready"`          //买跑准备好的用户（点击买跑好的用户）
	UserPiaoReady         bool                   `json:"userpiaoready"`         //自由漂选好的用户
	UserExchangeReady     bool                   `json:"userexchangeready"`     //换三张选好的用户
	UserExchangeThreeCard [3]byte                `json:"userexchangethreecard"` //换三张换的那三张牌
	UserExchangeCard      []byte                 `json:"userexchangecard"`      // 换出去的牌切片
	UserExchangeInCard    []byte                 `json:"userexchangeIncard"`    // 换进来的牌切片
	UserExchangeCardType  int                    `json:"userexchangecardtype"`  //换牌方向:顺时针逆时针对换
	ProvideUserCount      int                    `json:"provideusercount"`      // 点炮次数
	ChiHuUserCount        int                    `json:"chihuusercount"`        // 胡牌次数
	LaiYouHuUserCount     int                    `json:"laiyouhuusercount"`     // 赖油胡牌次数
	HuBySelfCount         int                    `json:"hubyselfcount"`         // 自摸次数
	MaxFanUserCount       int                    `json:"maxfanusercount"`       // 最大番数
	MaxScoreUserCount     int                    `json:"maxscoreusercount"`     // 最大分数
	BigHuUserCount        int                    `json:"bighuusercount"`        // 大胡次数
	ZhuangTimes           int                    `json:"zhuangtimes"`           //做庄次数
	ChiHu                 uint64                 `json:"chihu"`                 //吃胡标志.是否允许吃胡，用来控制过庄
	ChiHuScore            int                    `json:"chihuscore"`            //上一次吃胡分数.用来控制过庄，分数高于上一次分数时不用过庄
	VecChiHuCard          []byte                 `json:"vecchihucard"`          //本轮弃吃胡的牌，用来控制咸宁的过庄
	VecGangCard           []byte                 `json:"vecgangcard"`           //本局弃杠的牌
	VecPengCard           []byte                 `json:"VecPengCard"`           //本局弃碰的牌（20190325 苏大强 通城麻将）
	Response              bool                   `json:"response"`              //响应标志
	UserAction            byte                   `json:"useraction"`            //用户动作
	UserAction32          int                    `json:"useraction32"`          //用户动作32位
	OperateCard           byte                   `json:"operatecard"`           //操作扑克
	PerformAction         byte                   `json:"performaction"`         //执行动作
	PerformAction32       int                    `json:"performaction32"`       //执行动作32
	WeaveItemCount        byte                   `json:"weaveitemcount"`        //组合数目
	WeaveItemArray        [4]static.TagWeaveItem `json:"weaveitemarray"`        //组合扑克
	ThirdOperate          static.ThirdOpreate    `json:"thirdoperate"`          //记录四家第三次开口时,供牌的用户
	ChiHuResult           static.TagChiHuResult  `json:"chihuresult"`           //吃胡结果
	VecXiaPao             static.Msg_C_Xiapao    `json:"vecxiapao"`             //记录每一局下跑
	VecPiao               static.Msg_C_Piao      `json:"vecpiao"`               //记录每一局玩家漂几
	NeedGuoZhuang         bool                   `json:"needguozhuang"`         //记录过庄,需要经过自己一次操作（摸牌、吃、碰）才可以胡
	CurMagicOut           byte                   `json:"curmagicout"`           //记录当前小局飘癞子个数
	MagicScore            int                    `json:"magic_score"`           //赖子带来的得失分记载
	GangScore             int                    `json:"gang_score"`            //杠牌带来的得失分记载
	StorageScore          int                    `json:"storagescore"`          //游戏过程中分数变化量记载
	XiXiangFengScore      int                    `json:"xixiangfengscore"`      //记录一下喜相逢的分数
	XiXiangFengCard       []byte                 `json:"xixiangfengcard"`       //喜相逢的牌，记录下，避免重复计算
	LiangCnt              byte                   `json:"liangcnt"`              //标记亮倒
	LiangCard             byte                   `json:"liangcard"`             //标记亮倒次数
	PlayerMetaSZ
	LastGangScore   int  `json:"lastgangscore"`   //最近一次杠牌玩家的得失分
	IsBaoTingStatus bool `json:"isbaotingstatus"` //玩家是否报听出牌结算
	IsBaoTing       bool `json:"isbaoting"`       //玩家是否报听

	//201905009 苏大强 获嘉 暗听牌其他玩家不可见
	BaoTingCardIndex int   `json:"baoTingCardIndex"` //暗听的牌
	SpecialHuRecord  []int `json:"specialHuRecord"`  //分别记录4对，19句，风句，门清的数量，小结算用
	MengFeng         byte  `json:"mengfeng"`         //门风
	//20201113 点杠玩家记录
	ShowGangProviders []uint16 `json:"showgangproviders"` //点杠提供者
	//20190712 蓄杠玩家记录
	XuGangProvider []uint16 `json:"xuGangProvider"` //蓄杠碰牌提供者
	//20190712 上面的XuGangProvider太简单了，滁州比较恶，每次记录只有当前的，
	XuGangIndex []int `json:"xuGangIndex"` //蓄杠碰牌的牌的记录
	//20190717 滁州全球人包牌信息
	QuanqiurenBao static.BaoPaiInfo `json:"quanqiurenBao"` //全球人包牌信息
	//20190805 滁州老三番一个特殊的玩意 4遇字计数 值钱风计数
	SiYuZhiNum         int    `json:"siYuZhiNum"`
	Zhiqianfeng        int    `json:"zhiqianfeng"`
	BaoQingType        int    `json:"bao_qing_type"`      //报清
	Pilaicardcard      []byte `json:"pilaicardcard"`      //皮籁杠记录
	PilaicardcardClass []byte `json:"pilaicardcardclass"` //皮籁杠来源
	IsJinHu            bool   `json:"isjinhu"`            //20191010 恩施麻将禁胡
	LiangangCard       []byte `json:"LiangangCard"`       //20191025  苏大强 阜阳杠番 连杠记录
	LastUserAction     uint64 `json:"lastuseraction"`     //玩家最近一次的动作(吃碰杠胡弃+出牌)
	//20191228 苏大强 5杠数据
	FiveGangSource static.FiveGangSource `json:"fivegangsource"`
	XiaoChaoTian   int                   `json:"xiao_chao_tian"` //小朝天
	ChaoTian       int                   `json:"chao_tian"`      //大朝天
	FangXiao       int                   `json:"fan_xiao"`

	HuCardInfo []*static.ChihuCardsInfo_xlch `json:"chihucardsinfo"` //玩家胡牌信息
	//
	CardDataHauPai   []byte `json:"carddatahaupai"`   //广东推倒胡 花牌
	ShowGangProvider uint16 `json:"showgangprovider"` //明杠提供者-杠爆全包
	TingQQROutCard   byte   `json:"tingqqroutcard"`   //听全求人的时候玩家打出的那张牌
	TingQQRStatus    bool   `json:"tingqqrstatus"`    //听全求人状态
	GameScoreFen     []int  `json:"gameScoreFen"`     //各局分数记录
	WinCount         int    `json:"wincount"`         //赢牌次数
	PlayTurn1st      int    `json:"playturn1st"`      //1游次数
	//通城打滚统计发牌后手牌赖子个数
	MagicCardNum int `json:"magiccardnum"` //发牌后手上的赖子个数
	//20200902 苏大强 恩施麻将里面出现不能操作的情况记录一下
	IsOutCard   bool `json:"is_out_card"`   //如果是true就表示可以操作（出牌），如果是false就不能操作（出牌）
	IstotalTime bool `json:"istotalTime"`   //是不是累加计时状态恩施玩法要在开局的时候打开，时间到了就关掉，是不是发那个消息要看是不是累计玩法状态
	WantCard    byte `json:"want_card"`     // 想要的牌
	OutCardData byte `json:"out_card_data"` // 本轮出的牌
	WantGood    bool `json:"want_good"`     // 本轮是否需要good
	StatisticsUserGameOpt
}

type StatisticsUserGameOpt struct {
	/**strat** 统计操作数据以下用于统计 不用于游戏业务判断  ****/
	StatisticsGangKaiCount       int `json:"statistics_gang_kai_count"`        //杠开的次数
	StatisticsOpeningUserHu      int `json:"statistics_opening_user_hu"`       //起手胡的次数
	StatisticsHave4MagicCount    int `json:"statistics_have_4_magic_count"`    //摸到4个癞子的次数
	StatisticsOpening4MagicCount int `json:"statistics_opening_4_magic_count"` //起手摸到4个癞子的次数
	/**end*******/
}

func (self *PlayerMeta) SetWant(card byte) {
	self.WantCard = card
}

func (self *PlayerMeta) Reset() {
	self.HaveGang = false
	self.ChiHu = 0
	self.ChiHuScore = 0
	self.BaoQingType = 0

	self.HDCard.Index = int(static.INVALID_BYTE)
	self.HDCard.UserIndex = int(static.INVALID_CHAIR)
	self.HDCard.HDCard = static.INVALID_BYTE

	self.HongZhongGang = 0
	self.FaCaiGang = 0
	self.UserPaoReady = false

	self.Trustee = false
	self.QiangFailCount = 0
	self.QiangScore = 0

	self.Response = false
	self.OperateCard = 0
	self.PerformAction = 0
	self.PerformAction32 = 0

	self.EnjoinChiHu = false
	self.EnjoinChiPeng = false
	self.NeedGuoZhuang = false
	self.UserPiaoReady = false
	self.UserAction32 = 0
	self.VecChiHuCard = make([]byte, 0)
	self.VecGangCard = make([]byte, 0)
	self.VecPengCard = make([]byte, 0)
	self.XiXiangFengCard = make([]byte, 0)
	self.UserAction = 0
	//20190829 苏大强 我觉得有问题，0的话可能是庄家，修改为 public.INVALID_BYTE
	self.ThirdOperate.ProvideUser = static.INVALID_BYTE
	self.ThirdOperate.OperateCount = 0
	//结束信息
	self.InitChiHuResult()

	for k, _ := range self.CardIndex {
		self.CardIndex[k] = 0
	}

	//组合扑克
	self.WeaveItemArray = [4]static.TagWeaveItem{}
	self.WeaveItemCount = 0
	self.LiangCard = 0
	//self.DiscardCount = 0

	self.DiscardCard = []byte{}
	self.DiscardCardClass = []byte{}
	self.ShowGang = 0
	self.DianGang = 0
	self.HidGang = 0
	self.MagicCardGang = 0
	self.MagicCardChi = 0
	self.MagicCardOut = 0
	self.PiZiGangCount = 0
	self.GangUserCount = 0
	self.AnGangUserCount = 0
	self.XuGang = 0
	self.DianXuGang = 0
	self.MagicScore = 0
	self.GangScore = 0
	self.XiXiangFengScore = 0
	self.CurMagicOut = 0
	self.StorageScore = 0
	self.JiangPeng = 0
	self.FengPeng = 0
	self.FengGang = 0
	self.LiangCnt = 0

	self.PlayerMetaSZ.OnEnd()

	self.IsBaoTing = false
	self.IsBaoTingStatus = false
	self.BaoTingCardIndex = -1               //没有暗听就是255
	self.SpecialHuRecord = []int{0, 0, 0, 0} //写死4个
	self.ClearOperateCard()
	self.XuGangProvider = []uint16{}
	self.ShowGangProviders = []uint16{}
	self.XuGangIndex = []int{}
	self.ClearQuanqiurenBaoPaiInfo()
	self.SiYuZhiNum = 0
	self.PiZiCardOut = 0
	self.Pilaicardcard = []byte{}
	self.PilaicardcardClass = []byte{}
	self.HuCardInfo = []*static.ChihuCardsInfo_xlch{}
	self.IsJinHu = false
	self.MengFeng = static.INVALID_BYTE
	self.LiangangCard = []byte{}
	self.CheckTimeOut = 0
	self.EndAutoReadyTimeOut = 0
	self.ClickTime = 0
	self.RecordbeginTime = 0
	self.DelayerSetBeginTime = 0
	self.DelayerSetTime = 0
	self.LastUserAction = 0
	self.TrusteeNum = 0
	self.XiaoChaoTian = 0
	self.ChaoTian = 0
	self.FangXiao = 0
	//------------
	self.FiveGangSource.AreadyFiveGang = false
	self.FiveGangSource.UserStatus = [4]bool{false, false, false, false}
	//--------------

	self.CardDataHauPai = []byte{}
	self.ShowGangProvider = static.INVALID_CHAIR

	self.TingQQROutCard = 0xFF
	self.WantCard = static.INVALID_BYTE
	self.OutCardData = static.INVALID_BYTE
	self.TingQQRStatus = false
	self.OutCardRecord = []byte{}
	self.IsOutCard = true
	self.IstotalTime = false
}

func (self *PlayerMeta) Init() {
	self.Timer.Init()
}

func (self *PlayerMeta) SetHDCard(index byte, userIndex uint16, _card byte) {
	self.HDCard.Index = int(index)
	self.HDCard.UserIndex = int(userIndex)
	self.HDCard.HDCard = _card
}

//托管
func (self *PlayerMeta) SetTrustee(_truteee bool) {
	self.Trustee = _truteee
}

//记录玩家最近一次的操作,吃碰杠胡
func (self *PlayerMeta) SetUserLastAction(action uint64) {
	self.LastUserAction = action
}

//清除记录玩家最近一次的操作,吃碰杠胡
func (self *PlayerMeta) CleanUserLastAction() {
	self.LastUserAction = 0
}

func (self *PlayerMeta) InitCardIndex() {
	self.CardIndex = [static.MAX_INDEX]byte{}
}

func (self *PlayerMeta) InitCardDataHuaPai() {
	//清理花牌数
	self.CardDataHauPai = []byte{}
}

//用户摸牌
func (self *PlayerMeta) DispatchCard(_card byte) {
	self.CardIndex[self.SwitchToCardIndex(_card)]++
}

//删除手牌
func (self *PlayerMeta) DeleteCard(_card byte) {
	self.CardIndex[self.SwitchToCardIndex(_card)]--
}

//清除手牌
func (self *PlayerMeta) CleanCard(_card byte) {
	self.CardIndex[_card] = 0
}

//func (self *PlayerMeta) SetCardIndex(_cards [public.MAX_INDEX]byte) {
//	self.CardIndex = _cards
//}

//牌型转换
func (self *PlayerMeta) SwitchToCardIndex(cbCardData byte) byte {
	return ((cbCardData&static.MASK_COLOR)>>4)*9 + (cbCardData & static.MASK_VALUE) - 1
}

func (self *PlayerMeta) SetCardIndex(m_rule *rule2.St_FriendRule, cbCardData []byte, cbCardCount byte) byte {
	//设置变量
	self.CardIndex = [static.MAX_INDEX]byte{}
	//转换扑克
	for i := byte(0); i < cbCardCount; i++ {
		//ASSERT(IsValidCard(cbCardData[i]));
		if !self.IsValidCard(cbCardData[i], m_rule) {
			return cbCardCount
		}
		self.CardIndex[self.SwitchToCardIndex(cbCardData[i])]++
	}

	return cbCardCount
}

//下跑
func (self *PlayerMeta) XiaPao(_msg *static.Msg_C_Xiapao) {
	self.UserPaoReady = true
	self.VecXiaPao = *_msg
}

//下跑 返回
func (self *PlayerMeta) XiaPaor_safe(_msg *static.Msg_C_Xiapao) bool {
	if self.UserPaoReady {
		return false
	}
	self.UserPaoReady = true
	self.VecXiaPao = *_msg
	return true
}

func (self *PlayerMeta) CleanXiaPao() {
	self.VecXiaPao.Num = 0
	//self.VecXiaPao.Status = false
}

//选漂
func (self *PlayerMeta) XuanPiao(_msg *static.Msg_C_Piao) {
	self.UserPiaoReady = true
	self.VecPiao = *_msg
}
func (self *PlayerMeta) XuanPiao_safe(_msg *static.Msg_C_Piao) bool {
	if self.UserPiaoReady {
		return false
	}
	self.UserPiaoReady = true
	self.VecPiao = *_msg
	return true
}
func (self *PlayerMeta) CleanPiao() {
	self.VecPiao.Num = 0
	self.UserPiaoReady = false
	//self.VecXiaPao.Status = false
}

//记录吃胡类型
func (self *PlayerMeta) SetChiHuKind(_huKind uint64, _card byte, _score int) {
	self.ChiHu = _huKind
	self.VecChiHuCard = append(self.VecChiHuCard, _card)
	self.ChiHuScore = _score
}

func (self *PlayerMeta) InitChiHuResult() {
	//结束信息
	self.ChiHuResult.ChiHuKind = 0
	self.ChiHuResult.ChiHuRight = 0
	self.ChiHuResult.ChiHuKind2 = 0
	self.ChiHuResult.ChiHuUser = 0
}

//记录第三次开口
func (self *PlayerMeta) AddThirdOperate(_user uint16) {
	self.ThirdOperate.OperateCount++
	//如果是第三次开口的提供者,记录供牌的用户
	if self.ThirdOperate.OperateCount == 3 {
		self.ThirdOperate.ProvideUser = byte(_user)
	}
}

//添加吃碰杠的牌池
func (self *PlayerMeta) AddWeaveItemArray(_index int, _public byte, _provideUser uint16, _kind byte, _card byte) {
	self.WeaveItemArray[_index].PublicCard = _public
	self.WeaveItemArray[_index].ProvideUser = _provideUser
	self.WeaveItemArray[_index].WeaveKind = _kind
	self.WeaveItemArray[_index].CenterCard = _card
}

//是否需要报清权限
func (self *PlayerMeta) IsNeedPaoQing() bool {
	_count := self.GetOpenMouth()
	_bNeed := true
	_bJiang := true
	for i := byte(0); i < self.WeaveItemCount; i++ {

		CardValue := int(self.WeaveItemArray[i].CenterCard & static.MASK_VALUE)
		if self.WeaveItemArray[0].CenterCard&static.MASK_COLOR != self.WeaveItemArray[i].CenterCard&static.MASK_COLOR {
			_bNeed = false
		}

		if CardValue != 2 && CardValue != 5 && CardValue != 8 {
			_bJiang = false
		}
	}

	_bNeed = (_bNeed || _bJiang)

	if _count == 2 && _bNeed && (self.BaoQingType == static.WIK_BAOQING_QI || self.BaoQingType == static.WIK_BAOQING_NULL) {
		return true
	}
	return false
}

func (self *PlayerMeta) IsCanChi(cbCenterCard byte) bool {
	if self.BaoQingType == static.WIK_BAOQING_JIANG {
		return false
	}
	return self.IsCanChiPengGang(cbCenterCard)
}

func (self *PlayerMeta) IsCanChiPengGang(cbCenterCard byte) bool {
	if self.BaoQingType == static.WIK_BAOQING_QING {
		if self.WeaveItemArray[0].CenterCard&static.MASK_COLOR != cbCenterCard&static.MASK_COLOR {
			return false
		}
		if 0x30 == cbCenterCard&static.MASK_COLOR {
			return false
		}
	}

	if self.BaoQingType == static.WIK_BAOQING_FENG {
		if 0x30 != cbCenterCard&static.MASK_COLOR {
			return false
		}
	}

	if self.BaoQingType == static.WIK_BAOQING_JIANG {
		if 0x30 == cbCenterCard&static.MASK_COLOR {
			return false
		}
		if 0x02 != cbCenterCard&static.MASK_VALUE && 0x05 != cbCenterCard&static.MASK_VALUE && 0x08 != cbCenterCard&static.MASK_VALUE {
			return false
		}
	}

	return true
}

//是否可湖清一色
func (self *PlayerMeta) IsCanHuQing() bool {
	_count := self.GetOpenMouth()
	if _count > 1 && (self.BaoQingType == static.WIK_BAOQING_QI || self.BaoQingType == static.WIK_BAOQING_NULL) {
		return false
	}
	return true
}

//是否可胡一色
func (self *PlayerMeta) IsCanHuSe(_type int) bool {
	_count := self.GetOpenMouth()

	if _count > 1 && self.BaoQingType&_type == 0 {
		return false
	}
	return true
}

//只能清一色
func (self *PlayerMeta) OnlyCanHuQing(wChiHuKind uint64) bool {
	_count := self.GetOpenMouth()
	if _count > 1 && (self.BaoQingType != static.WIK_BAOQING_QI && self.BaoQingType != static.WIK_BAOQING_NULL) {
		if ((wChiHuKind&static.CHK_FENG_YI_SE) != 0 && self.IsCanHuSe(static.WIK_BAOQING_FENG)) ||
			((wChiHuKind&static.CHK_JIANG_JIANG) != 0 && self.IsCanHuSe(static.WIK_BAOQING_JIANG)) ||
			((wChiHuKind&static.CHK_QING_YI_SE) != 0 && self.IsCanHuSe(static.WIK_BAOQING_QING)) {
			return true
		}
		return false
	}
	return true
}

func (self *PlayerMeta) GetOpenMouth() byte {
	return self.WeaveItemCount - self.HidGang
}

//保留蓄杠的碰牌玩家记录，如果以前是碰，而后是杠，就修改一下
func (self *PlayerMeta) AddWeaveItemArray_Modify(_index int, _public byte, _provideUser uint16, _kind byte, _card byte) {
	self.WeaveItemArray[_index].PublicCard = _public
	self.WeaveItemArray[_index].ProvideUser1 = static.INVALID_CHAIR
	if self.WeaveItemArray[_index].WeaveKind == static.WIK_PENG && _kind == static.WIK_GANG {
		//蓄杠转换
		if self.WeaveItemArray[_index].ProvideUser != static.INVALID_CHAIR {
			self.WeaveItemArray[_index].ProvideUser1 = self.WeaveItemArray[_index].ProvideUser
		}
	}
	self.WeaveItemArray[_index].ProvideUser = _provideUser
	self.WeaveItemArray[_index].WeaveKind = _kind
	self.WeaveItemArray[_index].CenterCard = _card
}

//20190322 苏大强 追加记录是不是带了赖子
func (self *PlayerMeta) AddWeaveItemArrayEx(_index int, _public byte, _provideUser uint16, _kind byte, _card byte, magicCard byte) (laiChi bool) {
	self.AddWeaveItemArray(_index, _public, _provideUser, _kind, _card)
	//标记一下
	if _kind&0x7 > 0 {
		return self.Record_ChiGod(&self.WeaveItemArray[_index], magicCard)
	}
	return false
}

//增加赖子牌记录，实际上可以拓展成皮子牌记录的
func (self *PlayerMeta) Record_ChiGod(weaveItem *static.TagWeaveItem, m_bMagicCard byte) (laiChi bool) {
	laiChi = false
	if m_bMagicCard == static.INVALID_BYTE {
		return laiChi
	}
	//fmt.Println(fmt.Sprintf("CenterCard(%d)--MagicCard(%d)",weaveItem.CenterCard,MagicCard))
	//统一用两位表示，1代表中心牌是鬼牌，2代表边牌是鬼牌（吃牌情况）
	switch weaveItem.WeaveKind {
	case static.WIK_LEFT: //左吃操作
		if m_bMagicCard^(weaveItem.CenterCard+1) == 0 {
			weaveItem.MagicWeave |= 0x2
			laiChi = true
			break
		}
		if m_bMagicCard^(weaveItem.CenterCard+2) == 0 {
			weaveItem.MagicWeave |= 0x2
			laiChi = true
		}
	case static.WIK_RIGHT: //右吃操作
		if m_bMagicCard^(weaveItem.CenterCard-2) == 0 {
			weaveItem.MagicWeave |= 0x2
			laiChi = true
			break
		}
		if m_bMagicCard^(weaveItem.CenterCard-1) == 0 {
			weaveItem.MagicWeave |= 0x2
			laiChi = true
		}
	case static.WIK_CENTER: //中吃操作
		if m_bMagicCard^(weaveItem.CenterCard+1) == 0 {
			weaveItem.MagicWeave |= 0x2
			laiChi = true
			break
		}
		if m_bMagicCard^(weaveItem.CenterCard-1) == 0 {
			weaveItem.MagicWeave |= 0x2
			laiChi = true
		}
	default:
	}
	//杠碰的中心牌就够了，如果吃牌的情况下，吃的是赖子（这个情况不排除）
	if m_bMagicCard^weaveItem.CenterCard == 0 {
		weaveItem.MagicWeave |= 0x1
		laiChi = true
	}
	if weaveItem.MagicWeave == 0x2 {
		fmt.Sprintf("赖子吃")
	}
	if weaveItem.MagicWeave == 0x1 {
		fmt.Sprintf("吃赖子")
	}
	return laiChi
}
func (self *PlayerMeta) CleanWeaveItemArray() {
	self.WeaveItemArray = [4]static.TagWeaveItem{}
	self.WeaveItemCount = 0
}

//清理全球人包牌信息
func (self *PlayerMeta) ClearQuanqiurenBaoPaiInfo() {
	self.QuanqiurenBao.FristQuanQiuRen = false
	self.QuanqiurenBao.BaoPaicard = static.INVALID_BYTE
}

//设定，如果是全球人当前轮
/*
初始设定的时候，才会记录牌值。其他情况不记录,如果已经是初始记录了，再传值，只会置为无效牌
*/
func (self *PlayerMeta) SetQuanqiurenBaoPaiInfo(card byte) {
	if !self.QuanqiurenBao.FristQuanQiuRen {
		self.QuanqiurenBao.FristQuanQiuRen = true
		self.QuanqiurenBao.BaoPaicard = card
	} else {
		self.QuanqiurenBao.BaoPaicard = static.INVALID_BYTE
	}

}

//抢错
func (self *PlayerMeta) QiangFailed(_score int) {
	self.QiangFailCount++
	self.QiangScore -= _score
}

//碰牌
func (self *PlayerMeta) Peng(_card byte, quanfeng byte) {
	cbValue := (_card & static.MASK_VALUE)
	cbColor := (_card & static.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	if cbColor != 3 {
		return //
	}

	if cbValue >= 1 && cbValue <= 4 {

		if quanfeng == _card || _card == self.MengFeng {
			self.FengPeng++
		}
	}

	if cbValue > 4 && cbValue <= 7 {
		self.JiangPeng++
	}
}

//蓄杠，碰牌记录--
func (self *PlayerMeta) PengV2(_card byte, quanfeng byte) {
	cbValue := (_card & static.MASK_VALUE)
	cbColor := (_card & static.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	if cbColor != 3 {
		return //
	}

	if cbValue >= 1 && cbValue <= 4 {
		if quanfeng == _card || _card == self.MengFeng {
			self.FengPeng--
		}
	}

	if cbValue > 4 && cbValue <= 7 {
		self.JiangPeng--
	}
}

//红中杠
func (self *PlayerMeta) HongzhongGang() {
	self.HongZhongGang++
	self.HaveGang = true
}

//红中杠
func (self *PlayerMeta) OutHongzhong() {
	self.HongZhongGang++
}

func (self *PlayerMeta) SetOperateCard(_card byte) {
	self.OperateCard = _card
}

func (self *PlayerMeta) ClearOperateCard() {
	self.Response = false
	self.UserAction = 0
	self.UserAction32 = 0
	self.OperateCard = 0
	self.PerformAction = 0
}

func (self *PlayerMeta) ClearOperateCard32() {
	self.Response = false
	self.UserAction = 0
	self.UserAction32 = 0
	self.OperateCard = 0
	self.PerformAction = 0
	self.PerformAction32 = 0
}

//设置牌权
func (self *PlayerMeta) SetOperate(_card byte, _action byte) {
	self.Response = true
	self.OperateCard = _card
	self.PerformAction = _action
}

//设置牌权
func (self *PlayerMeta) SetOperate32(_card byte, _action int) {
	self.Response = true
	self.OperateCard = _card
	self.PerformAction32 = _action
}
func (self *PlayerMeta) CheckCardExist(_card byte) bool {
	// 效验玩家手上是否有这张牌
	if self.CardIndex[self.SwitchToCardIndex(_card)] <= 0 {
		xlog.Logger().Debug("this PlayerInfo not exist this card: %d", _card)
		return false
	}
	return true
}

//出花牌
func (self *PlayerMeta) OutHuaPaiCard(_card byte) bool {
	//删除花牌
	bRemove := false
	for i := 0; i < len(self.CardDataHauPai); i++ {
		if self.CardDataHauPai[i] == _card {
			self.CardDataHauPai = append(self.CardDataHauPai[:i], self.CardDataHauPai[i+1:]...)
			bRemove = true
			break
		}
	}
	if !bRemove {
		return false
	}
	self.UserAction = static.WIK_NULL
	self.PerformAction = static.WIK_NULL
	self.ChiHu = 0
	self.ChiHuScore = 0

	//玩家出牌了，过庄了，记录的过庄数据需要删除
	self.VecChiHuCard = make([]byte, 0)
	self.FreeQiPeng()
	//syslog.Logger().Debug("玩家已出牌，清空所有牌权动作及过庄数据")
	return true
}

//出牌
func (self *PlayerMeta) OutCard(_rule *rule2.St_FriendRule, _card byte) bool {

	if !self.RemoveCard(_rule, _card) {
		return false
	}
	self.UserAction = static.WIK_NULL
	self.PerformAction = static.WIK_NULL
	self.ChiHu = 0
	self.ChiHuScore = 0
	self.OutCardData = _card
	//玩家出牌了，过庄了，记录的过庄数据需要删除
	self.VecChiHuCard = make([]byte, 0)
	self.FreeQiPeng()
	//syslog.Logger().Debug("玩家已出牌，清空所有牌权动作及过庄数据")
	return true
}

//发财杠
func (self *PlayerMeta) GangFacai() {
	self.FaCaiGang++
	self.HaveGang = true
}

//发财杠
func (self *PlayerMeta) OutFacai() {
	self.FaCaiGang++
}

//痞子杠
func (self *PlayerMeta) GangPizi() {
	self.PiZiGangCount++
	self.HaveGang = true
}

//赖子杠
func (self *PlayerMeta) GangMagic() {
	self.MagicCardGang++
	self.HaveGang = true
}

//吃赖子
func (self *PlayerMeta) ChiMagic() {
	self.MagicCardChi++
}

//赖子杠当普通牌打出，算分
func (self *PlayerMeta) OutMagicCardV2() {
	self.MagicCardGang++
}

//赖子当普通牌打
func (self *PlayerMeta) OutMagicCard() {
	self.MagicCardOut++
	self.CurMagicOut++
}

//皮子当普通牌打
func (self *PlayerMeta) OutPiziCard() {
	self.PiZiCardOut++
}

//点炮
func (self *PlayerMeta) ProvideCard() {
	self.ProvideUserCount++
}

//20190326 接炮次数
func (self *PlayerMeta) HuByChi() {
	self.ChiHuUserCount++

}

//add by zwj for 汉川搓虾子 赖油胡牌次数
func (self *PlayerMeta) HuLaiYou() {
	self.LaiYouHuUserCount++
}

//吃胡
func (self *PlayerMeta) ChiHuCard(_kind uint64) {
	self.ChiHuUserCount++

	if _kind&0x4F00 != 0 { //大胡次数
		self.BigHuUserCount++
	}
}

//添加玩家胡牌,用在血流成河麻将
func (self *PlayerMeta) AddHuCard(_card int, provider uint16, owner uint16) {
	info := new(static.ChihuCardsInfo_xlch)
	info.HuCard = _card
	info.HuCardProvider = provider
	info.HuCardOwner = owner
	self.HuCardInfo = append(self.HuCardInfo, info)
}

//自摸
func (self *PlayerMeta) HuBySelf() {
	self.HuBySelfCount++
}

func (self *PlayerMeta) SetMaxFan(_count int) {
	self.MaxFanUserCount = _count
}

func (self *PlayerMeta) SetMaxScore(_score int) {
	self.MaxScoreUserCount = _score
}

//丢牌
func (self *PlayerMeta) Discard(_card byte) {
	self.DiscardCard = append(self.DiscardCard, _card)
}

func (self *PlayerMeta) AppendOutCardRecord(_card byte) {
	self.OutCardRecord = append(self.OutCardRecord, _card)
}

//丢牌
func (self *PlayerMeta) Discard_ex(_card byte, cardclass byte) {
	self.DiscardCard = append(self.DiscardCard, _card)
	self.DiscardCardClass = append(self.DiscardCardClass, cardclass)
}

func (self *PlayerMeta) SetCardDataHuaPai(_card byte) {
	self.CardDataHauPai = append(self.CardDataHauPai, _card)
}

//丢牌
func (self *PlayerMeta) DiscardPiLai(_card byte) {
	self.Pilaicardcard = append(self.Pilaicardcard, _card)
}

//丢牌
func (self *PlayerMeta) DiscardPiLai_ex(_card byte, cardclass byte) {
	self.Pilaicardcard = append(self.Pilaicardcard, _card)
	self.PilaicardcardClass = append(self.PilaicardcardClass, cardclass)
}

//20181204 苏大强 添加一个，如果有人需求丢弃的牌，再把他拿出来
//因为都是即时的，就删除最后一个就行了
func (self *PlayerMeta) Requiredcard(card byte) {
	if len(self.DiscardCard) != 0 && self.DiscardCard[len(self.DiscardCard)-1] == card {
		// fmt.Println(fmt.Sprintf("玩家的弃牌被需求了，被需求牌（%s）,弃牌（%v）", StrCardsMessage1[self.SwitchToCardIndex(self.DiscardCard[len(self.DiscardCard)-1])], self.DiscardCard))
		self.DiscardCard = self.DiscardCard[:len(self.DiscardCard)-1]
		// fmt.Println(fmt.Sprintf("去牌后（%v）", self.DiscardCard))
	} else {
		fmt.Println(fmt.Sprintf("无牌可弃，或者请求的牌不在牌库中（%d）", card))
	}
}

func (self *PlayerMeta) Gang(_type byte, _card byte, fengquan byte) {

	cbValue := (_card & static.MASK_VALUE)
	cbColor := (_card & static.MASK_COLOR) >> 4

	if fengquan == _card || _card == self.MengFeng {
		self.fengGang(_type, _card, fengquan)
	} else if cbColor == 3 && (cbValue > 4 && cbValue <= 7) {
		self.jiangGang(_type, _card, fengquan)
	}

	switch _type {
	case info2.E_Gang: //杠牌，明杠
		self.ShowGangAction()
	case info2.E_Gang_XuGand: //蓄杠，即回头杠
		self.XuGangAction()
	case info2.E_Gang_AnGang: //暗杠
		self.HidGangAction()
	}

}

func (self *PlayerMeta) fengGang(_type byte, _card byte, fengquan byte) {

	switch _type {
	case info2.E_Gang: //杠牌，明杠
		//self.ShowGang++
	case info2.E_Gang_XuGand: //蓄杠，即回头杠
		//self.XuGang++
		self.PengV2(_card, fengquan)
	case info2.E_Gang_AnGang: //暗杠
		//self.HidGang++
	}
	//self.GangUserCount++
	self.FengGang++
}

func (self *PlayerMeta) jiangGang(_type byte, _card byte, fengquan byte) {

	switch _type {
	case info2.E_Gang: //杠牌，明杠
		//self.ShowGang++
	case info2.E_Gang_XuGand: //蓄杠，即回头杠
		//self.XuGang++
		self.PengV2(_card, fengquan)
	case info2.E_Gang_AnGang: //暗杠
		//self.HidGang++
	}
	//self.GangUserCount++
	self.JiangGang++
}

//明杠
func (self *PlayerMeta) ZuoZhuangAction() {
	self.ZhuangTimes++
}

//明杠
func (self *PlayerMeta) ShowGangAction() {
	self.ShowGang++
	self.HaveGang = true
	self.GangUserCount++
}

//明杠
func (self *PlayerMeta) ShowGangAction_ex(proverUser uint16) {
	self.ShowGang++
	self.HaveGang = true
	self.GangUserCount++
	if proverUser != static.INVALID_CHAIR {
		self.ShowGangProviders = append(self.ShowGangProviders, proverUser)
	}

}

//点杠
func (self *PlayerMeta) DianGangAction() {
	self.DianGang++
}

//蓄杠
func (self *PlayerMeta) XuGangAction() {
	self.XuGang++
	self.HaveGang = true
	self.GangUserCount++
}

func (self *PlayerMeta) XuGangV2(_card byte, fengquan byte) {
	self.XuGangAction()
	self.PengV2(_card, fengquan)
}

//点蓄杠
func (self *PlayerMeta) DianXuGangAction() {
	self.DianXuGang++
}

//暗杠
func (self *PlayerMeta) HidGangAction() {
	self.HidGang++
	self.HaveGang = true
	self.AnGangUserCount++
}

func (self *PlayerMeta) OnNextGame() {
	self.CardIndex = [static.MAX_INDEX]byte{}
	//组合扑克
	self.WeaveItemArray = [4]static.TagWeaveItem{}
	self.WeaveItemCount = 0
	self.LiangCard = 0
	self.HaveGang = false
	self.CheckTimeOut = 0
	self.EndAutoReadyTimeOut = 0
	self.CurMagicOut = 0
	self.MagicScore = 0
	self.GangScore = 0
	self.XiXiangFengScore = 0
	self.XiaoChaoTian = 0
	self.ChaoTian = 0
	self.FangXiao = 0
	self.VecChiHuCard = make([]byte, 0)
	self.VecGangCard = make([]byte, 0)
	/**
	StorageScore
	玩家在小结算之前发生分数变化，用该字段储存起来，在小结算的时候去合计改变玩家分数，之后重置
	保证真正修改玩家分数的时间只发生在小结算的时候
	保证修改玩家分数的方式只有通过TableWriteGameDate()方法
	*/
	self.StorageScore = 0
	self.PlayerMetaSZ.OnNextGame()
	//self.GameTimer.KillTimer(GameTime_AutoNext) //清除自动开始数据
	self.Timer.OnNextClean()
	self.WantCard = static.INVALID_BYTE
	self.OutCardData = static.INVALID_BYTE
}

func (self *PlayerMeta) CleanWant() {
	self.WantCard = static.INVALID_BYTE
}

func (self *PlayerMeta) OnBegin() {
	//一大局开始后重置吓跑数据（一大局里面的小局开始不能重置）
	self.UserPaoReady = false
	self.ProvideUserCount = 0
	self.ChiHuUserCount = 0
	self.HuBySelfCount = 0
	self.MaxFanUserCount = 0
	self.MaxScoreUserCount = -99999
	self.BigHuUserCount = 0
	self.ZhuangTimes = 0
	self.WantCard = static.INVALID_BYTE
	self.OutCardData = static.INVALID_BYTE
	// 纸牌
	self.WinCount = 0
	self.PlayTurn1st = 0
	self.MagicCardNum = 0

	self.CleanXiaPao()
	self.CleanPiao()

}

func (self *PlayerMeta) OnEnd() {
	self.ProvideUserCount = 0
	self.ChiHuUserCount = 0
	self.HuBySelfCount = 0
	self.MaxFanUserCount = 0
	self.MaxScoreUserCount = -99999
	self.BigHuUserCount = 0
	self.ZhuangTimes = 0
	self.WantCard = static.INVALID_BYTE
	self.OutCardData = static.INVALID_BYTE
	self.PlayerMetaSZ.OnEnd()
}

//有效判断
func (self *PlayerMeta) IsValidCard(cbCardData byte, m_rule *rule2.St_FriendRule) bool {
	cbValue := (cbCardData & static.MASK_VALUE)
	cbColor := (cbCardData & static.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	bIsValid := (((cbValue >= 1) && (cbValue <= 9) && (cbColor <= 2)) || ((cbValue >= 1) && (cbValue <= 7) && (cbColor == 3)))
	if !bIsValid {
		return false //非法牌
	}

	//校验：去万
	if m_rule.NoWan && cbColor == 0 {
		return false
	}

	return true //(((cbValue>=1)&&(cbValue<=9)&&(cbColor<=2))||((cbValue>=1)&&(cbValue<=7)&&(cbColor==3)));
}

//删除单张扑克
func (self *PlayerMeta) RemoveCard(m_rule *rule2.St_FriendRule, cbRemoveCard byte) bool {
	//效验扑克
	if !self.IsValidCard(cbRemoveCard, m_rule) {
		xlog.Logger().Debug("RemoveCard failed : %d", cbRemoveCard)
		return false
	}
	//效验扑克
	if self.CardIndex[self.SwitchToCardIndex(cbRemoveCard)] <= 0 {
		xlog.Logger().Debug("RemoveCard failed : %d", cbRemoveCard)
		return false
	}

	cbRemoveIndex := self.SwitchToCardIndex(cbRemoveCard)

	if self.CardIndex[cbRemoveIndex] > 0 {
		self.CardIndex[cbRemoveIndex]--

		return true
	}
	return false
}

//删除多张扑克
func (self *PlayerMeta) RemoveCards(m_rule *rule2.St_FriendRule, cbRemoveCard []byte) bool {
	//删除扑克

	for i := 0; i < len(cbRemoveCard); i++ {
		//效验扑克
		if !self.IsValidCard(cbRemoveCard[i], m_rule) {
			xlog.Logger().Debug("RemoveCard2 failed : %d", cbRemoveCard)
			continue
		}
		//效验扑克
		if self.CardIndex[self.SwitchToCardIndex(cbRemoveCard[i])] <= 0 {
			xlog.Logger().Debug("RemoveCard2 failed : %d", cbRemoveCard)
			continue
		}

		//删除扑克
		cbRemoveIndex := self.SwitchToCardIndex(cbRemoveCard[i])
		if self.CardIndex[cbRemoveIndex] == 0 {
			//还原删除
			for j := 0; j < i; j++ {
				//ASSERT(IsValidCard(cbRemoveCard[j]));
				self.CardIndex[self.SwitchToCardIndex(cbRemoveCard[j])]++
			}
			return false
		} else {
			//删除扑克
			self.CardIndex[cbRemoveIndex]--
		}
	}
	return true
}

//添加多张牌
func (self *PlayerMeta) AddCards(m_rule *rule2.St_FriendRule, cbAddCard []byte) bool {
	for i := 0; i < len(cbAddCard); i++ {
		//效验扑克
		if !self.IsValidCard(cbAddCard[i], m_rule) {
			xlog.Logger().Debug("AddCard2 failed : %d", cbAddCard)
			continue
		}
		//添加
		cbAddIndex := self.SwitchToCardIndex(cbAddCard[i])

		//添加
		self.CardIndex[cbAddIndex]++
	}
	return true
}

// 得到杠明细
func (self *PlayerMeta) WeaveDetail(chair uint16, picard byte, garMask byte) []static.TagWeaveDetail {
	result := make([]static.TagWeaveDetail, 0)
	for _, item := range self.WeaveItemArray {
		var r static.TagWeaveDetail
		switch item.WeaveKind {
		case static.WIK_GANG:
			{
				r.WeaveKind = info2.E_Gang
				if item.ProvideUser == chair {
					r.WeaveKind = info2.E_Gang_AnGang
				}
				r.ProvideUser = item.ProvideUser
			}
		case static.WIK_PENG:
			{
				if garMask&info2.GONG_PIZI > 0 && item.CenterCard == picard {
					r.WeaveKind = info2.E_Gang
					r.ProvideUser = item.ProvideUser
				} else {
					continue
				}
			}
		case static.WIK_FILL | static.WIK_GANG:
			{
				r.WeaveKind = info2.E_Gang_XuGand
				r.ProvideUser = item.ProvideUser
			}
		default:
			{
				xlog.Logger().Debug("未知的组合类型：：", item.WeaveKind)
			}
		}
		result = append(result, r)
	}
	return result
}

// 20190325 苏大强 清理掉弃操作的东西
func (self *PlayerMeta) FreeVerCards() {
	self.VecChiHuCard = []byte{}
	self.VecGangCard = []byte{}
	self.VecPengCard = []byte{}
	self.XiXiangFengCard = []byte{}
}

func (self *PlayerMeta) FreeQiPeng() {
	self.VecPengCard = []byte{}
}

func (self *PlayerMeta) QiPeng(_card byte) {
	self.VecPengCard = append(self.VecPengCard, _card)
}

func (self *PlayerMeta) OnChaoTian() {
	self.ChaoTian++
}

func (self *PlayerMeta) OnXiaoChaoTian() {
	self.XiaoChaoTian++
}

func (self *PlayerMeta) NeedGuoPeng(_card byte) bool {
	for _, v := range self.VecPengCard {
		if v == _card {
			return true
		}
	}
	return false
}

// 20191219 苏大强 小局托管次数
func (self *PlayerMeta) TrustRecord() {
	self.TrusteeNum++
}

//20200401 苏大强
/*
level=0 只要弃 就咔嚓
level=1 必须相等

kind =0 过庄
kind=1 过碰
返回 true 需要过庄或者过碰
*/
func (self *PlayerMeta) CheckNeedGuo(cbCheckCard byte, level int, kind int) bool {
	switch level {
	case 0:
		switch kind {
		case 0:
			//过庄
			if len(self.VecChiHuCard) != 0 {
				return true
			}
		case 1:
			//过碰
			if len(self.VecPengCard) != 0 {
				return true
			}
		default:
			return false
		}
	case 1:
		switch kind {
		case 0:
			//过庄
			return mahlib2.Findcard(self.VecChiHuCard, cbCheckCard)
		case 1:
			//过碰
			return mahlib2.Findcard(self.VecPengCard, cbCheckCard)
		default:
			return false
		}
	}
	return false
}

//20201110 苏大强 直接放这里吧
// 是否为弃杠牌
func (self *PlayerMeta) IsGiveUpGang(card byte) bool {
	for _, c := range self.VecGangCard {
		if c == card {
			return true
		}
	}
	return false
}

// 追加一个弃杠牌
func (self *PlayerMeta) AppendGiveUpGang(card byte) {
	if self.IsGiveUpGang(card) {
		return
	}
	self.VecGangCard = append(self.VecGangCard, card)
}

func (self *PlayerMeta) AppendGiveUpGang_ex(card byte) bool {
	if self.IsGiveUpGang(card) {
		return false
	}
	self.VecGangCard = append(self.VecGangCard, card)
	return true
}
