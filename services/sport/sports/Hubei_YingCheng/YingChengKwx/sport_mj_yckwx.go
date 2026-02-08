package YingChengKwx

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"github.com/open-source/game/chess.git/services/sport/wuhan"
	"math/rand"
	"strings"
	"time"
)

/*
应城卡五星
*/

const (
	CHK_M4G_1 = 0x400  //明4归1个
	CHK_M4G_2 = 0x800  //明4归2个
	CHK_M4G_3 = 0x8000 //明4归3个
)

const (
	Piao_k5x_ZiYouPiao     = -1 //自由漂
	Piao_k5x_NoPiao        = 0  //不漂分
	Piao_k5x_DingPiaoFirst = 1  //首局定漂
)

const (
	MaiMa_k5x_No           = iota //不买马
	MaiMa_k5x_LiangDaoZiMo        //亮倒自摸买马
	MaiMa_k5x_ZiMo                //自摸买马
)

const (
	PeiZhuang_k5x_NoTing = iota //赔庄_未听牌
	PeiZhuang_k5x_Ting          //赔庄_听牌
	PeiZhuang_k5x_Liang         //赔庄_亮牌
)

const (
	PeiFen_NULL = iota
	PeiFen_True
	PeiFen_False
)

const (
	OutCard_Type_NULL = iota
	OutCard_Type_ByServer
	OutCard_type_ByUser
	OutCard_Type_LiangDao
)

// 好友房规则相关属性
type SportFriendRuleYCKWX struct {
	Difen            int    `json:"difen"`            //	底分
	Dff              int    `json:"dff"`              //	底分倍数
	NineSecondRoom   string `json:"sc9"`              //	九秒场
	FengDing         int    `json:"fengding"`         //  封顶
	DingPiao         int    `json:"dingpiao"`         //	定漂:-1每局选漂0不漂1首局定漂
	IpLimit          string `json:"ip"`               //	ipgps限制
	MaiMa            int    `json:"maima"`            //	不买马0 亮倒自摸买马1 自摸买马2
	GangPaox4        string `json:"gangpaox4"`        //	杠上炮/花 x4
	ShuKan           string `json:"shukan"`           //	数坎
	K5x4             string `json:"k5x4"`             //	卡五星 x4
	PPx4             string `json:"ppx4"`             //	碰碰胡 x4
	DLFF             string `json:"dlff"`             //	对亮翻番
	GHFF             string `json:"ghff"`             //	过胡翻番
	QiHux2           string `json:"qihu"`             //	2分起胡4分起胡
	Overtime_trust   int    `json:"overtime_trust"`   //超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //超时解散
	Endready         string `json:"endready"`         //小结算是否自动准备
	Dissmiss         int    `json:"dissmiss"`         // 解散次数
}

//type K5x_Replay_Order struct {
//	meta2.Replay_Order
//	LiangPai	   string	`json:"liangpai"`	//亮牌
//}
//
//type K5x_Replay_Record struct {
//	RecordHandCard [static.MAX_CHAIR][static.MAX_COUNT]byte `json:"handcard"`      // 用户最初手上的牌
//	VecOrder       []K5x_Replay_Order                       `json:"vecorder"`      // 用户操作
//	BigHuKind      byte                                     `json:"bighukind"`     // 标记大胡类型
//	WBigHuKind     [4]byte									`json:"wbighukind"`	   // 一炮多响大胡类型
//	ProvideUser    uint16                                   `json:"provideuser"`   // 点炮用户
//	PiziCard       byte                                     `json:"pizicard"`      // 癞子皮
//	LeftCardCount  byte                                     `json:"leftcardcount"` // 发完牌后剩余数目
//	Score          [static.MAX_CHAIR]int                    `json:"score"`         // 游戏积分
//	UVitamin       map[int64]float64                        `json:"u_vitamin"`     // 玩家起始疲劳值
//	Fengquan       byte                                     `json:"fengquan"`      // 风圈
//}
//
////游戏记录重置
//func (self *K5x_Replay_Record) Reset() {
//
//	for i := 0; i < static.MAX_CHAIR; i++ {
//		for j := 0; j < static.MAX_COUNT; j++ {
//			sp.RecordHandCard[i][j] = 0
//		}
//	}
//
//	for i := 0; i < len(sp.Score); i++ {
//		sp.Score[i] = 0
//	}
//
//	sp.UVitamin = make(map[int64]float64)
//	sp.VecOrder = make([]K5x_Replay_Order, 0, 30)
//	sp.BigHuKind = 2 //0用来表示将一色了
//	sp.ProvideUser = 9
//	sp.PiziCard = 0
//	sp.Fengquan = 0
//	sp.WBigHuKind = [4]byte{}
//}

type PiaoK5X struct {
	bIsXuanPiaoSure bool                  //是否选漂完成
	iPlayerPiaoNum  [meta2.MAX_PLAYER]int //玩家选漂数
}

//type ShowCard struct{
//	BIsShowCard			bool					//	是否亮牌
//	CbTingCard			[20]byte				//	听牌
//	CbAnPuCard			[static.MAX_COUNT]byte	//	暗铺的牌
//	CbLiangCard			[13]byte	//	亮出的牌
//}
//
//func (self * ShowCard) Reset(){
//	sp.BIsShowCard = false
//	sp.CbTingCard = [20]byte{}
//	sp.CbAnPuCard = [static.MAX_COUNT]byte{}
//	sp.CbLiangCard = [13]byte{}
//}

type MSG_S_GuoHu struct {
	Uid        int64  `json:"uid"`
	GuoHuCount [4]int `json:"guohucount"`
}

// 这里的数组有空改下用切片吧,不然日志看着心累,用int数据就行了
type LiangItem struct {
	LiangCard [13]byte               `json:"liangcard"`
	TingCard  [20]byte               `json:"tingcard"`
	AnPuCard  [static.MAX_COUNT]byte `json:"anpucard"`
}

func (li *LiangItem) Reset() {
	li.TingCard = [20]byte{}
	li.AnPuCard = [static.MAX_COUNT]byte{}
	li.LiangCard = [13]byte{}
}

// 操作命令
type Msg_C_OperateCard_K5X struct {
	static.Msg_C_OperateCard
	Code int `json:"code"` //操作类型
}

type Msg_C_LiangPai struct {
	LiangStruct LiangItem `json:"liangstruct"` //亮牌操作结构体
	OutCard     byte      `json:"data"`        //亮牌打出的扑克
	Code        int       `json:"code"`        //操作类型
	Id          int64     `json:"id"`          //玩家ID
}

// 出牌命令
type Msg_C_OutCard_K5X struct {
	static.Msg_C_OutCard
	IsLiang bool `json:"isliang"` //是否是亮倒出牌
}

type Msg_S_OutCard_K5X struct {
	static.Msg_S_OutCard
	IsLiang     bool `json:"isliang"`     //是否是亮倒时打出的那张牌,用来客户端控制亮倒时的播亮倒bgm
	TuoGuanType int  `json:"tuoguantype"` //出牌类型：标识这张牌是怎么打出来的 0未知类型1服务器自动(托管)2玩家主动打出3亮倒后自动打出的
}

type CMD_S_StatusPlay32 struct {
	static.CMD_S_StatusPlay
	ActionMask       int          `json:"actionmask"`  //动作掩码
	LiangStructArray [4]LiangItem `json:"liangstruct"` //亮牌操作结构体
}

// 操作命令
type Msg_S_OperateResult_K5X struct {
	static.Msg_S_OperateResult
	OperateCode int                    `json:"operatecode"` //操作代码
	LiangCard   [static.MAX_COUNT]byte `json:"liangcard"`
	TingCard    [20]byte               `json:"tingcard"`
}

// 操作分
type Msg_S_OperateScore_K5X struct {
	OperateUser uint16     `json:"operateuser"`  //操作用户
	OperateType uint16     `json:"operatetype"`  //操作类型
	GameScore   [4]int     `json:"gamescore"`    //最新总分
	GameVitamin [4]float64 `json:"game_vitamin"` //最新疲劳值信息
	ScoreOffset [4]int     `json:"scoreoffset"`  //分数变化量
}

type Msg_S_GameStart_K5X struct {
	static.Msg_S_GameStart
	UserAction32 int `json:"useraction"` //用户动作
}

// 赔庄
type K5X_peizhuang struct {
	UserTing     [4]bool   //玩家是否听牌,跟赔庄相关
	UserTingType [4]uint64 //玩家听牌的最大牌型
}

type SportYCKWX struct {
	// 游戏共用部分
	components2.Common
	// 游戏流程数据
	meta2.Metadata
	//漂分结构
	PiaoK5X
	//赔庄
	K5X_peizhuang
	//游戏逻辑
	m_GameLogic           SportLogicYCKWX
	m_ShowCard            [4]meta2.ShowCard //亮牌
	QiangGangOperateScore meta2.Msg_S_OperateScore_K5X
	bQiangGangScoreSend   bool                    //抢杠的结果是不是发送了
	LianGangCount         int                     //连杠个数,连杠中断重置为0,杠1个加1
	GuoHuCount            [4]int                  //玩家过胡次数
	ReplayRecord          meta2.K5x_Replay_Record //回放记录
	KanCount              [4]int                  //玩家数坎个数
	HasSendNo13Tip        bool                    //不够13张不能亮牌
	VecGameDataAllP32     [4][]CMD_S_StatusPlay32 //记录每一局的结束时所有人的桌面数据
}

// ! 设置游戏可胡牌类型
func (sp *SportYCKWX) HuTypeInit(_type *static.TagHuType) {
	_type.HAVE_SIXI_HU = false
	_type.HAVE_QUE_YISE_HU = false
	_type.HAVE_BANBAN_HU = false
	_type.HAVE_LIULIU_HU = false
	_type.HAVE_QING_YISE_HU = true
	_type.HAVE_FENG_YI_SE = true
	_type.HAVE_HAO_HUA_DUI_HU = true
	_type.HAVE_JIANG_JIANG_HU = true
	_type.HAVE_FENG_YISE_HU = false
	_type.HAVE_QUAN_QIU_REN = true
	_type.HAVE_PENG_PENG_HU = true
	_type.HAVE_HAI_DI_HU = true
	_type.HAVE_QIANG_GANG_HU = true
	_type.HAVE_GANG_SHANG_KAI_HUA = true
	_type.HAVE_MENG_QING = false
	_type.HAVE_DI_HU = false
	_type.HAVE_TIAN_HU = false
	_type.HAVE_ZIMO_JIAO_1 = false
}

// ! 获取游戏配置
func (sp *SportYCKWX) GetGameConfig() *static.GameConfig { //获取游戏相关配置
	return &sp.Config
}

// ! 重置桌子数据
func (sp *SportYCKWX) RepositTable() {
	rand.Seed(time.Now().UnixNano())
	for _, v := range sp.PlayerInfo {
		v.Reset()
	}
	//游戏变量
	sp.SiceCount = components2.MAKEWORD(byte(1), byte(1))

	//出牌信息
	sp.OutCardData = 0
	sp.OutCardCount = 0
	sp.OutCardUser = static.INVALID_CHAIR

	//发牌信息
	sp.SendCardData = 0
	sp.SendCardCount = 0
	sp.LeftBu = 0

	//运行变量
	sp.ProvideCard = 0
	sp.ResumeUser = static.INVALID_CHAIR
	sp.CurrentUser = static.INVALID_CHAIR
	sp.ProvideUser = static.INVALID_CHAIR
	sp.PiZiCard = 0x00

	//状态变量
	sp.GangFlower = false
	sp.SendStatus = false
	sp.HaveHuangZhuang = false
	for k, _ := range sp.RepertoryCard {
		sp.RepertoryCard[k] = 0
	}

	sp.FanScore = [4]meta2.Game_mj_fan_score{}
	//赔庄
	sp.UserTing = [4]bool{}
	sp.UserTingType = [4]uint64{}

	for _, v := range sp.PlayerInfo {
		v.Reset()
	}
	//结束信息
	sp.ChiHuCard = 0
	//抢杠胡的杠
	sp.ResetXuGangScore()
	//连杠清零
	sp.LianGangCount = 0
	//听牌玩家
	sp.UserTing = [4]bool{}       //玩家是否听牌,跟赔庄相关
	sp.UserTingType = [4]uint64{} //玩家听牌的最大牌型
	//玩家过胡次数
	sp.GuoHuCount = [4]int{}
	//玩家数坎个数
	sp.KanCount = [4]int{}
	for i := 0; i < 4; i++ {
		//亮牌
		sp.m_ShowCard[i].Reset()
	}

	sp.HasSendNo13Tip = false
}

// ! 解析配置的任务
func (sp *SportYCKWX) ParseRule(strRule string) {

	xlog.Logger().Debug("parserRule :" + strRule)

	sp.Rule.CreateType = 0
	sp.Rule.NineSecondRoom = true

	sp.Rule.FangZhuID = sp.GetTableInfo().Creator
	sp.Rule.JuShu = sp.GetTableInfo().Config.RoundNum
	sp.Rule.CreateType = sp.FriendInfo.CreateType
	sp.Rule.TrusteeCostSharing = true //托管的人扣房费
	sp.Rule.Endready = false
	if len(strRule) == 0 {
		return
	}

	var _msg SportFriendRuleYCKWX
	if err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg); err == nil {
		sp.Rule.DiFen = _msg.Difen
		sp.Rule.DingPiao = _msg.DingPiao
		sp.Rule.MaiMaType = _msg.MaiMa
		sp.Rule.FengDing = _msg.FengDing
		sp.Rule.GangPaox4 = _msg.GangPaox4 == "true"
		sp.Rule.ShuKan = _msg.ShuKan == "true"
		sp.Rule.K5x4 = _msg.K5x4 == "true"
		sp.Rule.PPx4 = _msg.PPx4 == "true"
		sp.Rule.DLFF = _msg.DLFF == "true"
		sp.Rule.GHFF = _msg.GHFF == "true"
		sp.Rule.QiHuX2 = _msg.QiHux2 == "true"
		sp.Rule.Overtime_trust = _msg.Overtime_trust
		if _msg.Overtime_dismiss == 0 {
			sp.Rule.Overtime_dismiss = 1
		} else {
			sp.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		}
		//if sp.Rule.Overtime_dismiss>0{
		//	sp.LaunchDismissTime=sp.Rule.Overtime_dismiss
		//}
		sp.Rule.Endready = _msg.Endready == "true"
		sp.Rule.DissmissCount = _msg.Dissmiss
		if sp.Rule.DiFen <= 0 {
			sp.Rule.DiFen = _msg.Dff
		}
	}
	if sp.Rule.FengDing == 0 {
		sp.Rule.FengDing = 8
	}
	//设置解散次数
	if sp.Rule.DissmissCount != 0 {
		sp.SetDissmissCount(sp.Rule.DissmissCount)
	}
}

// ! 开局
func (sp *SportYCKWX) OnBegin() {
	xlog.Logger().Debug("onbegin")
	sp.RepositTable()

	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range sp.PlayerInfo {
		v.OnBegin()
	}

	// 第一局开放玩家为庄家
	sp.BankerUser = 0
	sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
	sp.m_GameLogic.Rule = sp.Rule
	sp.m_GameLogic.HuType = sp.HuType
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_GameEnd{}
	sp.VecGameDataAllP32 = [4][]CMD_S_StatusPlay32{}

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()

	sp.OnGameStart()
}

func (sp *SportYCKWX) OnGameStart() {
	//设置离线解散时间5分钟
	sp.SetOfflineRoomTime(1800)
	sp.SetDismissRoomTime(120)
	if !sp.CanContinue() {
		return
	}

	// 框架发送开始游戏后开始计算当前这一轮的局数
	sp.CurCompleteCount++
	sp.GetTable().SetBegin(true)
	//开始新的一局记录(提出来)
	sp.ReplayRecord.Reset()
	//for _,v := range sp.PlayerInfo{
	//	v.UserStatus_ex = static.US_PLAY
	//}
	//选漂
	if sp.Rule.DingPiao != Piao_k5x_NoPiao {
		sp.ChoosePiao()
	} else {
		sp.StartNextGame()
	}
}

// ! 发送下跑对话框
func (sp *SportYCKWX) SendPaoSetting(bIsRelink bool) {

	sp.GameEndStatus = static.GS_MJ_PLAY
	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)

	if !bIsRelink {
		for _, v := range sp.PlayerInfo {
			v.Ctx.CleanWeaveItemArray()
			v.Ctx.InitCardIndex()
			v.Ctx.CleanXiaPao()
		}
	}

	sp.PayPaoStatus = true //设置玩家选漂的状态
	var PaoSetting static.Msg_S_PaoSetting
	//向每个玩家发送数据
	for _, v := range sp.PlayerInfo {
		sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
		if v.Ctx.UserPaoReady == true {
			PaoSetting.PaoStatus = true
			PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
			PaoSetting.Overtime = v.Ctx.CheckTimeOut
			PaoSetting.Always = v.Ctx.VecXiaPao.Status
			PaoSetting.BankerUser = sp.BankerUser
		} else {
			PaoSetting.PaoStatus = false
			PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
			PaoSetting.Overtime = v.Ctx.CheckTimeOut
			PaoSetting.Always = v.Ctx.VecXiaPao.Status
			PaoSetting.BankerUser = sp.BankerUser
		}
		sp.SendPersonMsg(consts.MsgTypeGamePaoSetting, PaoSetting, v.Seat)
	}
}

// ! 玩家选择跑
func (sp *SportYCKWX) OnUserClientXiaPao(msg *static.Msg_C_Xiapao) bool {
	nChiarID := sp.GetChairByUid(msg.Id)
	_userItem := sp.GetUserItemByChair(nChiarID)
	if _userItem == nil {
		return true
	}
	if sp.Rule.NineSecondRoom {
		sp.OperateMutex.Lock()
		defer sp.OperateMutex.Unlock()
	}
	if nChiarID >= 0 && nChiarID < meta2.MAX_PLAYER {
		_userItem.Ctx.XiaPao(msg)

		sp.NotifyXiaPao(nChiarID)
		sp.iPlayerPiaoNum[nChiarID] = msg.Num
		fmt.Println(fmt.Sprintf("玩家%d,选跑%d", nChiarID, msg.Num))
	}

	// 如果4个玩家都准备好了，自动开启下一局
	_beginCount := 0
	for _, v := range sp.PlayerInfo {
		if !v.Ctx.UserPaoReady {
			recordStr := fmt.Sprintf("椅子编号[%d] 还没有完成选跑", v.Seat)
			sp.OnWriteGameRecord(uint16(v.Seat), recordStr)
			break
		}
		_beginCount++
	}

	if _beginCount >= sp.GetPlayerCount() {
		sp.OnWriteGameRecord(uint16(nChiarID), "所有人都完成选跑了，开始游戏")
		sp.PayPaoStatus = false
		sp.bIsXuanPiaoSure = true
		//游戏没有开始发牌
		if !sp.GameStartForXiapao {
			sp.StartNextGame()
		}
	}

	return true
}

// ! 广播玩家的状态和选漂的数目
func (sp *SportYCKWX) NotifyXiaPao(wChairID uint16) bool {
	var sXiaPiao static.Msg_S_Xiapao

	for _, v := range sp.PlayerInfo {
		if v.Seat != static.INVALID_CHAIR {
			sXiaPiao.Num[v.Seat] = v.Ctx.VecXiaPao.Num
			sXiaPiao.Always[v.Seat] = v.Ctx.VecXiaPao.Status
			sXiaPiao.Status[v.Seat] = v.Ctx.UserPaoReady
		}
	}

	//发送数据
	if sp.Rule.DingPiao == Piao_k5x_ZiYouPiao || sp.Rule.DingPiao == Piao_k5x_DingPiaoFirst && sp.CurCompleteCount == 1 {
		sp.SendTableMsg(consts.MsgTypeGameXiapao, sXiaPiao)
	}

	//游戏记录
	if wChairID == wChairID {
		recordStr := fmt.Sprintf("发送跑数：%d， 是否默认 %t", sXiaPiao.Num[wChairID], sXiaPiao.Status[wChairID])
		sp.OnWriteGameRecord(wChairID, recordStr)

		sp.addReplayOrder(wChairID, info2.E_Pao, sXiaPiao.Num[wChairID])
	}
	return true
}

// ! 开始下一局游戏
func (sp *SportYCKWX) StartNextGame() {
	sp.OnStartNextGame()
	sp.LastOutCardUser = static.INVALID_CHAIR
	sp.LastSendCardUser = uint16(static.INVALID_CHAIR)

	//组合扑克
	sp.MagicCard = 0x00

	sp.LeftCardCount = 0
	sp.RepertoryCard = []byte{}

	//发送最新状态
	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.SendUserStatus(i, static.US_PLAY) //把状态发给其他人
	}

	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始StartNextGame......")
	sp.OnWriteGameRecord(static.INVALID_CHAIR, sp.GetTableInfo().Config.GameConfig)

	for _, v := range sp.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
	sp.m_GameLogic.Rule = sp.Rule

	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(sp.GetTableId()+sp.KIND_ID*100+sp.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	sp.SiceCount = components2.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	sp.LeftCardCount = byte(len(sp.RepertoryCard))
	sp.LeftBu = 10 //剩下的补牌数

	//这里在没有调用混乱扑克的函数时m_cbRepertoryCard中是空的，当它调用了这个函数之后
	//在这个函数中把固定的牌打乱后放到这个数组中，在放的同时不断增加数组m_cbRepertoryCard
	//的长度
	sp.LeftCardCount, sp.RepertoryCard = sp.m_GameLogic.RandCardData()
	sp.CreateLeftCardArray(sp.GetPlayerCount(), int(sp.LeftCardCount), false)

	//分发扑克--即每一个人解析他的13张牌结果存放在m_cbCardIndex[i]中

	for _, v := range sp.PlayerInfo {
		if v.Seat != static.INVALID_CHAIR {
			sp.LeftCardCount -= (static.MAX_COUNT - 1)
			v.Ctx.SetCardIndex(&sp.Rule, sp.RepertoryCard[sp.LeftCardCount:], static.MAX_COUNT-1)
		}
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	newLeftCount, _ := sp.InitDebugCards_ex("mahjongK5xing_test", &sp.RepertoryCard, &sp.BankerUser)

	//////////////读取配置文件设置牌型end////////////////////////////////////
	//发送扑克---这是发送给庄家的第十四张牌
	sp.SendCardCount++
	sp.LeftCardCount--
	sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount]

	//sp.LeftCardCount--
	//sp.PiZiCard = sp.RepertoryCard[sp.LeftCardCount]
	//
	////转换赖子值
	//cbValue := byte(sp.PiZiCard & static.MASK_VALUE)
	//cbColor := byte(sp.PiZiCard & static.MASK_COLOR)
	//
	//if (cbValue == 9 && cbColor <= 0x20)  {
	//	//值等于9 但是花色是万 同 条 或者值是7（白板）花色是风
	//	cbValue = 0
	//	sp.MagicCard = (cbValue + 1) | cbColor
	//} else if  (cbValue == 7 && cbColor == 0x30){
	//	if sp.Rule.NoFeng{
	//		//如果勾选了去风,荆门双开的是去东风,东风不能成为皮赖
	//		cbValue = 0
	//		sp.MagicCard = (cbValue + 2) | cbColor
	//	}else{
	//		//值是7（白板）花色是风
	//		cbValue = 0
	//		sp.MagicCard = (cbValue + 1) | cbColor
	//	}
	//}else if (cbValue == 4 || cbValue == 5 || cbValue == 6) && cbColor == 0x30 {
	//	//痞子是北风红中发财,都是白板的赖子
	//	cbValue = 7
	//	sp.MagicCard = cbValue | cbColor
	//} else {
	//	sp.MagicCard = (cbValue + 1) | cbColor
	//}
	//
	////最后13-18张不能有痞子赖子
	//for i := 12; i <= 18; i++ {
	//	checkCard := sp.RepertoryCard[i]
	//	if checkCard == gameserver.CARD_HONGZHONG || checkCard == gameserver.CARD_FACAI || checkCard == sp.MagicCard {
	//		for j := 0; j < 12; j++ {
	//			changeCard := sp.RepertoryCard[j]
	//			if changeCard != gameserver.CARD_HONGZHONG && changeCard != gameserver.CARD_FACAI && changeCard != sp.MagicCard {
	//				sp.RepertoryCard[i] = changeCard
	//				sp.RepertoryCard[j] = checkCard
	//				break
	//			}
	//		}
	//	}
	//}
	//
	//sp.m_GameLogic.SetMagicCard(sp.MagicCard)
	//sp.m_GameLogic.SetPiZiCard(sp.PiZiCard)

	//写游戏日志
	sp.WriteGameRecord()

	_userItem := sp.GetUserItemByChair(sp.BankerUser)
	_userItem.Ctx.DispatchCard(sp.SendCardData)

	//设置变量
	sp.ProvideCard = 0
	sp.ProvideUser = static.INVALID_CHAIR
	sp.CurrentUser = sp.BankerUser //供应用户
	sp.LastSendCardUser = sp.BankerUser

	//测试流局
	if newLeftCount != 0 {
		sp.LeftCardCount = newLeftCount
	}

	//动作分析,只分析庄家的牌，只有庄家初始才可以杠
	//杠牌判断
	var GangCardResult static.TagGangCardResult
	_userItem.Ctx.UserAction32 |= sp.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, nil, 0, &GangCardResult, sp.m_ShowCard)
	//胡牌判断
	//wChiHuRight := uint64(0)
	//_userItem.Ctx.UserAction32 |= sp.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:],
	//	_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, sp.SendCardData, wChiHuRight, &_userItem.Ctx.ChiHuResult, false)
	_userItem.Ctx.UserAction32 |= sp.CheckHu(sp.BankerUser, sp.BankerUser, sp.SendCardData, false, false, false, false, false)
	//亮牌判断
	_userItem.Ctx.UserAction32 |= sp.CheckLiangPai(sp.BankerUser, false)

	////卡五星不能起手胡,非清一色的暗四归也不能起手胡
	//if _userItem.Ctx.UserAction32 & static.WIK_CHI_HU != 0 {
	//	if _userItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_KA_5_XING != 0 {
	//		_userItem.Ctx.ChiHuResult.ChiHuKind ^= static.CHK_KA_5_XING
	//	}
	//
	//	if _userItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_SI_GUI_YI_AN != 0 &&
	//		_userItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_QING_YI_SE == 0 {
	//		_userItem.Ctx.ChiHuResult.ChiHuKind ^= static.CHK_SI_GUI_YI_AN
	//	}
	//}

	//构造数据,发送开始信息
	var GameStart Msg_S_GameStart_K5X
	GameStart.SiceCount = sp.SiceCount
	GameStart.BankerUser = sp.BankerUser
	GameStart.CurrentUser = sp.CurrentUser
	GameStart.MagicCard = sp.PiZiCard
	sp.LockTimeOut(sp.BankerUser, sp.Rule.Overtime_trust) //static.GAME_OPERATION_TIME_12)
	GameStart.Overtime = sp.LimitTime                     //time.Now().Unix() + static.GAME_OPERATION_TIME_12
	GameStart.LeftCardCount = sp.LeftCardCount
	GameStart.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	GameStart.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	GameStart.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	GameStart.CurrentCount = sp.CurCompleteCount

	//记录癞子牌
	sp.ReplayRecord.PiziCard = sp.PiZiCard
	//记录发完牌后剩牌数量
	sp.ReplayRecord.LeftCardCount = sp.LeftCardCount

	//向每个玩家发送数据
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//设置变量
		GameStart.UserAction32 = _item.Ctx.UserAction32 //把上面分析过的结果保存再发送到客户端
		_, GameStart.CardData = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameStart.CardData)
		//记录玩家手上初始牌
		_, sp.ReplayRecord.RecordHandCard[i] = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, sp.ReplayRecord.RecordHandCard[i])
		//记录玩家初始分
		UserItem := sp.GetUserItem(i)
		if UserItem != nil {
			//TODO 玩家分数设置
			sp.ReplayRecord.Score[i] = 0
			sp.ReplayRecord.UVitamin[UserItem.Info.Uid] = _item.UserScoreInfo.Vitamin
			if uint16(i) == sp.BankerUser {
				GameStart.SendCardData = sp.SendCardData //发给庄家的第一张牌
			} else {
				GameStart.SendCardData = static.INVALID_BYTE
			}
		}
		if _item.CheckTRUST() {
			GameStart.Whotrust[i] = true
		}
		//发送数据
		sp.SendPersonMsg(consts.MsgTypeGameStart, GameStart, uint16(i))
	}

	if _userItem.Ctx.UserAction32 != 0 {
		sp.ResumeUser = sp.CurrentUser
		sp.SendOperateNotify()
	}
}

// ! 得到某个用户开口的次数,吃，碰，明杠的次数
func (sp *SportYCKWX) GetUserOpenMouth(wChairID uint16) uint16 {
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return uint16(0)
	}

	return uint16(_userItem.Ctx.WeaveItemCount - _userItem.Ctx.HidGang)
}

// ! 初始化游戏
func (sp *SportYCKWX) OnInit(table base2.TableBase) {
	sp.KIND_ID = table.GetTableInfo().KindId
	sp.Config.StartMode = static.StartMode_FullReady
	sp.Config.PlayerCount = 4 //玩家人数
	sp.Config.ChairCount = 4  //椅子数量
	sp.PlayerInfo = make(map[int64]*components2.Player)
	sp.HuTypeInit(&sp.HuType) //设置可胡牌类型

	sp.RepositTable()
	sp.Config.StartIgnoreOffline = true
	sp.SetGameStartMode(static.StartMode_FullReady)
	sp.GameTable = table
	sp.Init()
	sp.Unmarsha(table.GetTableInfo().GameInfo)
	table.GetTableInfo().GameInfo = ""
	//设置自动解散时间5分钟
	sp.SetDismissRoomTime(120)
	//设置离线解散时间5分钟
	sp.SetOfflineRoomTime(1800)
	sp.SetVitaminLowPauseTime(10)

	//防作弊模式下改成30秒,游戏开始后要重新设置成300秒
	if sp.GameTable.GetTableInfo().JoinType == consts.NoCheat || sp.GameTable.GetTableInfo().JoinType == consts.AutoAdd {
		sp.SetOfflineRoomTime(30)
	}
}

// ! 发送消息
func (sp *SportYCKWX) OnMsg(msg *base2.TableMsg) bool {

	switch msg.Head {
	case consts.MsgTypeGameBalanceGameReq: //! 请求总结算信息

		var _msg static.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.CalculateResultTotal_Rep(&_msg)
		}
	case consts.MsgTypeGameOutCard: //! 出牌消息
		//var _msg static.Msg_C_OutCard
		var _msg Msg_C_OutCard_K5X
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			//sp.OnUserOutCard(&_msg)
			_msg.ByClient = true
			if !sp.OnUserOutCard(&_msg, true) && sp.Rule.NineSecondRoom && _msg.ByClient {
				sp.flashClient(_msg.Id, "出牌失败")
			}
		}
	case consts.MsgTypeGameOperateCard: //操作消息
		var _msg Msg_C_OperateCard_K5X
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			//sp.OnUserOperateCard(&_msg)
			_msg.ByClient = true
			if !sp.OnUserOperateCard(&_msg, true) && sp.Rule.NineSecondRoom && _msg.ByClient {
				sp.flashClient(_msg.Id, "限时到，服务器自动选弃")
			}
		}
	case consts.MsgTypeGameGoOnNextGame: //下一局
		var _msg static.Msg_C_GoOnNextGame
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.OnUserClientNextGame(&_msg)
		}
	case consts.MsgTypeGameXiapao: //选漂
		var _msg static.Msg_C_Xiapao
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			//sp.OnUserClientXiaPao(&_msg)
			if !sp.OnUserClientXiaPao(&_msg) && sp.Rule.NineSecondRoom && _msg.ByClient {
				//sp.flashClient(_msg.Id, "限时到，服务器自动选飘")
			}
		}
	case consts.MsgTypeGameLiangPai: //亮牌
		var _msg Msg_C_LiangPai
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if !sp.OnUserOperateLiangPai(&_msg) { //&& sp.Rule.NineSecondRoom && _msg.ByClient {
				sp.flashClient(_msg.Id, "")
			}
		}
	case consts.MsgTypeGameTrustee: //用户托管
		//var _msg static.Msg_C_Trustee
		var _msg static.Msg_S_DG_Trustee
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.OnUserTustee(&_msg)
		}
	case consts.MsgCommonToGameContinue:
		opt, ok := msg.V.(*static.TagSendCardInfo)
		if ok {
			sp.DispatchCardData(opt.CurrentUser, opt.GangFlower, false)
		} else {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "common to game 断言失败。")
		}
	default:
		//sp.Common.OnMsg(msg)
	}
	return true
}

// ! 下一局
func (sp *SportYCKWX) OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu || sp.GetGameStatus() != static.GS_MJ_END {
		return true
	}

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()

	nChiarID := sp.GetChairByUid(msg.Id)

	sp.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
	sp.SendUserStatus(int(nChiarID), static.US_READY)

	if nChiarID >= 0 && nChiarID < uint16(sp.GetPlayerCount()) {
		_item := sp.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < sp.GetPlayerCount(); i++ {
		item := sp.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			break
		}
		if i == sp.GetPlayerCount()-1 {
			// 复位桌子
			sp.RepositTable()
			sp.OnGameStart()
		}
	}
	return true
}

// ! 清除吃胡记录
func (sp *SportYCKWX) initChiHuResult() {
	for _, v := range sp.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// ! 清除单个玩家记录
func (sp *SportYCKWX) ClearChiHuResultByUser(wCurrUser uint16) {
	for _, v := range sp.PlayerInfo {
		if v.GetChairID() == wCurrUser {
			v.Ctx.InitChiHuResult()
			break
		}
	}
}

// ! 反向清除单个玩家记录
func (sp *SportYCKWX) ClearChiHuResultByUserReverse(wCurrUser uint16) {
	for _, v := range sp.PlayerInfo {
		if v.GetChairID() != wCurrUser {
			v.Ctx.InitChiHuResult()
		}
	}
}

// ! 用户操作牌
func (sp *SportYCKWX) OnUserOperateCard(msg *Msg_C_OperateCard_K5X, needlock bool) bool {
	if sp.Rule.NineSecondRoom && needlock {
		sp.OperateMutex.Lock()
		defer sp.OperateMutex.Unlock()
	}

	wChairID := sp.GetChairByUid(msg.Id)

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}

	//效验用户
	if (wChairID != sp.CurrentUser) && (sp.CurrentUser != static.INVALID_CHAIR) {
		return false
	}

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return true
	}

	// 能胡牌没有胡需要过庄
	if (_userItem.Ctx.UserAction32&static.WIK_CHI_HU) != 0 && msg.Code != static.WIK_CHI_HU {
		//别人打牌自己弃胡才算过庄
		if sp.CurrentUser == static.INVALID_CHAIR {
			_userItem.Ctx.NeedGuoZhuang = true
		}
	}

	if msg.Code != static.WIK_NULL {
		// 解锁用户超时操作
		sp.UnLockTimeOut(wChairID)
	}

	if msg.Code == static.WIK_GANG {
		if !sp.CheckCanHaveGangAction() {
			return true
		}
		//如果是杠,则记杠+1
		sp.LianGangCount++
	} else {
		sp.LianGangCount = 0
	}

	//游戏记录
	autoStr := "服务端自动"
	if msg.ByClient {
		autoStr = "玩家主动"
	}
	if msg.Code == static.WIK_NULL {
		sp.OnWriteGameRecord(wChairID, autoStr+"点击弃！")
	}

	// 回放中记录牌权操作
	sp.addReplayOrder(wChairID, info2.E_HandleCardRight, msg.Code)

	//被动动作,被动操作没有红中杠，赖子杠,不分析抢杠
	if sp.CurrentUser == static.INVALID_CHAIR {
		sp.OnUserOperateInvalidChair(msg, _userItem)
		return true
	}

	//主动动作，杠的是红中，赖子，和暗杠，此种情况下蓄杠要考抢杠的操作
	if sp.CurrentUser == wChairID {
		sp.OnUserOperateByChair(msg, _userItem)
		return true
	}

	return true
}

// ! 被动动作，别人打牌吃碰杠胡牌
func (sp *SportYCKWX) OnUserOperateInvalidChair(msg *Msg_C_OperateCard_K5X, _userItem *components2.Player) bool {

	wTargetUser := sp.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card

	//效验状态
	if _userItem.Ctx.Response {
		return false
	}
	if (cbOperateCode != static.WIK_NULL) && ((_userItem.Ctx.UserAction32 & cbOperateCode) == 0) {
		return false
	}
	if cbOperateCard != sp.ProvideCard {
		return false
	}

	//变量定义
	cbTargetAction := cbOperateCode
	//构造结果
	var OperateResult Msg_S_OperateResult_K5X
	var OperateScore Msg_S_OperateScore_K5X

	//设置变量
	_userItem.Ctx.SetOperate32(cbOperateCard, cbOperateCode)
	if cbOperateCard == 0 {
		_userItem.Ctx.SetOperateCard(sp.ProvideCard)
	}

	if cbOperateCode != static.WIK_CHI_HU && _userItem.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
		if sp.m_ShowCard[wTargetUser].BIsShowCard { //&& _userItem.Ctx.UserAction32 & static.WIK_CHI_HU != 0 {
			fmt.Println("玩家亮牌后,可胡后选择了弃胡")
			//游戏记录
			recordStr := fmt.Sprintf("玩家亮牌后,可胡后选择了弃胡")
			sp.OnWriteGameRecord(wTargetUser, recordStr)
			if sp.Rule.GHFF {
				//过胡翻番
				fmt.Println(fmt.Sprintf("玩家%d过胡翻番", wTargetUser))
				sp.GuoHuCount[wTargetUser]++

				var guohu MSG_S_GuoHu
				guohu.Uid = msg.Id
				for i := 0; i < sp.GetPlayerCount(); i++ {
					guohu.GuoHuCount[i] = sp.GuoHuCount[i]
				}
				sp.SendTableMsg(consts.MsgTypeGameGuoHu, guohu)

				//游戏记录
				recordStr := fmt.Sprintf("玩家选择过胡翻番")
				sp.OnWriteGameRecord(wTargetUser, recordStr)
			}
			//清掉胡权
			_userItem.Ctx.UserAction32 ^= static.WIK_CHI_HU
			_userItem.Ctx.ChiHuResult.ChiHuKind = static.CHK_NULL
		}
	}

	//执行判断
	for _, v := range sp.PlayerInfo {
		//获取动作
		cbUserAction := v.Ctx.UserAction32

		if v.Ctx.Response {
			cbUserAction = v.Ctx.PerformAction32
		}

		//优先级别
		cbUserActionRank := sp.m_GameLogic.GetUserActionRank(cbUserAction) // 动作等级
		cbTargetActionRank := sp.m_GameLogic.GetUserActionRank(cbTargetAction)

		//动作判断
		if cbUserActionRank > cbTargetActionRank {
			wTargetUser = v.Seat
			cbTargetAction = cbUserAction
		}
	}

	// 最大操作权限的人还没有操作则返回
	if _userItem = sp.GetUserItemByChair(wTargetUser); _userItem != nil && !_userItem.Ctx.Response {
		return true
	}

	//变量定义
	autoStr := "服务端自动"
	if msg.ByClient {
		autoStr = "玩家主动"
	}

	//可能有多人是最高等级，多人胡牌，先把已经选择胡的人装起来
	var wTargetCHS []uint16
	//if cbTargetAction == static.WIK_CHI_HU {
	iTargetResponse := 0
	for _, v := range sp.PlayerInfo {
		//获取动作 只有多胡才会出现多人能够请求一个这里修改,因为到这里吃碰不可能有多个玩家请求,所以去掉获取玩家的选择
		cbUserAction := v.Ctx.UserAction32 //sp.m_cbUserAction[i]
		//动作判断 把请求是最高权限的用户放在一起(一炮多响)
		if sp.m_GameLogic.GetUserActionRank(int(cbUserAction)) == sp.m_GameLogic.GetUserActionRank(cbTargetAction) {
			wTargetCHS = append(wTargetCHS, v.Seat)
			if v.Ctx.Response {
				iTargetResponse++
			}
		}
	}
	if len(wTargetCHS) > 1 {
		xlog.Logger().Debug(fmt.Sprintf("最高权限人：%v, cbTargetAction:%v", wTargetCHS, cbTargetAction))
	}
	//如果有2个以上的人要胡，其中一个人点了胡，另外的人也就胡了
	if cbTargetAction == static.WIK_CHI_HU {
		if iTargetResponse != len(wTargetCHS) && (_userItem.Ctx.PerformAction32 != static.WIK_CHI_HU) {
			xlog.Logger().Debug("都还未响应,最大玩家选的不是胡牌")
			return true
		}
	} else {
		if userItem := sp.GetUserItemByChair(wTargetUser); userItem != nil {
			if userItem.Ctx.Response == false {
				xlog.Logger().Debug(fmt.Sprintf("玩家（%d）座位号（%d）请求（%d）是否响应（%t）", userItem.Uid, userItem.Seat, userItem.Ctx.PerformAction32, userItem.Ctx.Response))
				return true
			}
		}
	}
	//}

	//如果打出去的牌,所有有操作权的人,都没选择胡
	OperateRightCount := 0
	for _, s := range wTargetCHS {
		if userItem := sp.GetUserItemByChair(s); userItem != nil {
			if userItem.Ctx.UserAction32 != static.WIK_CHI_HU {
				OperateRightCount++
			}
		}
	}
	if len(wTargetCHS) == OperateRightCount {
		if provideItem := sp.GetUserItemByChair(sp.ProvideUser); provideItem != nil {
			if provideItem.Ctx.LastUserAction&static.WIK_GANG != 0 {
				provideItem.Ctx.SetUserLastAction(static.WIK_NULL)
			}
		}
	}

	sp.UnLockAllPlayerTimer() //清除所有人的定时器

	//记录玩家最新的一次操作
	_userItem.Ctx.SetUserLastAction(uint64(msg.Code))

	cbTargetCard := _userItem.Ctx.OperateCard
	//出牌变量
	sp.SendStatus = true
	if cbTargetAction != static.WIK_NULL {
		sp.OutCardData = 0
		sp.OutCardUser = static.INVALID_CHAIR

		if providItem := sp.GetUserItemByChair(sp.ProvideUser); providItem != nil {
			providItem.Ctx.Requiredcard(cbTargetCard)
		}
	}

	if cbTargetAction == static.WIK_NULL {
		//用户状态
		for _, v := range sp.PlayerInfo {
			v.Ctx.ClearOperateCard32()
		}

		if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_QIANG_GANG != 0 && !sp.bQiangGangScoreSend {
			//玩家有抢杠胡权,但是弃了抢杠胡,这个杠分还是要结算的
			sp.ReCoverXuGangScore()
		}

		//放弃操作
		if _userItem = sp.GetUserItemByChair(sp.ResumeUser); _userItem != nil && _userItem.Ctx.PerformAction32 != static.WIK_NULL {
			wTargetUser = sp.ResumeUser
			cbTargetAction = _userItem.Ctx.PerformAction32
		} else {
			if sp.LeftCardCount > 0 {
				_targetUserItem := sp.GetUserItemByChair(wTargetUser)
				if (_targetUserItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_QIANG_GANG) != 0 {
					sp.DispatchCardData(sp.ResumeUser, true, false)
				} else {
					sp.DispatchCardData(sp.ResumeUser, false, false)
				}
			} else {
				sp.ChiHuCard = 0
				sp.ProvideUser = static.INVALID_CHAIR
				sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
			}
			return true
		}
	} else if cbTargetAction == static.WIK_CHI_HU {
		//下面是一炮多响的代码
		//结束信息
		sp.ChiHuCard = cbTargetCard
		sp.ProvideUser = sp.ProvideUser
		sp.OutCardData = 0

		for _, v := range wTargetCHS {
			wTargetUser = v
			if userItem := sp.GetUserItemByChair(wTargetUser); userItem != nil {
				//普通胡牌，有人点炮
				if sp.ChiHuCard != 0 {
					if userItem.Ctx.ChiHuResult.ChiHuKind != static.CHK_NULL {
						userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(sp.ChiHuCard)]++
					}
				} else { //自摸的 自摸只有一个玩家，这里用大海的 正常的自摸没到这里，估计是杠开花的才会来 测试了一下没有来，估计这块代码可以去掉了
					// xlog.Logger().Debug(fmt.Sprintf("自摸玩家（%d）座位号（%d）记录座位号（%d），结算结果（%v）", userItem.Uid, userItem.Seat, wTargetUser, userItem.Ctx.ChiHuResult))
					if userItem.Ctx.UserAction32 != static.WIK_NULL {
						sp.ProvideUser = uint16(wTargetUser)
					}
					if userItem.Ctx.ChiHuResult.ChiHuKind != static.CHK_NULL {
						userItem.Ctx.DeleteCard(sp.ChiHuCard)
					}
				}
				//游戏记录
				recordStr := fmt.Sprintf("%s，胡牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
				sp.OnWriteGameRecord(wTargetUser, recordStr)
				//  xlog.Logger().Debug(fmt.Sprintf("发送给客户端的数据：座位号（%d）,(%s)", wTargetUser, recordStr))
				//记录胡牌
				sp.addReplayOrder(wTargetUser, info2.E_Hu, int(cbTargetCard))
				//xlog.Logger().Debug(fmt.Sprintf("玩家（%d）座位号（%d）结算前的胡牌状态：%v", userItem.Uid, userItem.Seat, userItem.Ctx.ChiHuResult))
			}
		}
		//结束游戏
		sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
		return true
	} else {
		//用户状态
		for _, v := range sp.PlayerInfo {
			v.Ctx.ClearOperateCard32()
		}

		//组合扑克
		wIndex := int(_userItem.Ctx.WeaveItemCount)
		_userItem.Ctx.WeaveItemCount++
		_provideUser := sp.ProvideUser
		if sp.ProvideUser == static.INVALID_CHAIR {
			_provideUser = wTargetUser
		}
		_userItem.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, byte(cbTargetAction), cbTargetCard)

		//删除扑克
		switch cbTargetAction {
		case static.WIK_PENG: //碰牌操作
			{
				//亮牌之后不能再碰了
				if sp.m_ShowCard[wTargetUser].BIsShowCard {
					return true
				}

				var GangCardResult static.TagGangCardResult
				var cbHighAction int

				//判断该玩家是否可以杠这张牌
				_userItem.Ctx.DispatchCard(cbTargetCard)
				cbHighAction |= sp.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex,
					_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult, sp.m_ShowCard)

				_userItem.Ctx.DeleteCard(cbTargetCard)
				bIsCurCardGang := false
				for _, card := range GangCardResult.CardData {
					if card == cbTargetCard {
						bIsCurCardGang = true
					}
				}
				// 玩家在碰的时候有杠的牌权但是没有杠,加入弃杠记录
				if cbHighAction&static.WIK_GANG != 0 && bIsCurCardGang {
					sp.m_GameLogic.AppendGiveUpGang(_userItem, cbTargetCard)
				}
				sp.operateCard(byte(cbTargetAction), cbTargetCard, _userItem) //删除扑克
			}
		case static.WIK_GANG: //杠牌操作
			{
				//删除扑克
				cbRemoveCard := []byte{cbTargetCard, cbTargetCard, cbTargetCard}
				_userItem.Ctx.RemoveCards(&sp.Rule, cbRemoveCard)
				//杠统计
				_userItem.Ctx.ShowGangAction()
				mingGangScore := sp.GetScoreOnGang(info2.E_Gang) * sp.Rule.DiFen * sp.GetLianGangFan()
				OperateScore.ScoreOffset[_provideUser] -= mingGangScore
				OperateScore.ScoreOffset[_userItem.GetChairID()] += mingGangScore
				OperateScore.OperateUser = wTargetUser
				OperateScore.OperateType = info2.E_Gang
				sp.OnUserScoreOffset(_provideUser, -mingGangScore)
				sp.OnUserScoreOffset(_userItem.GetChairID(), mingGangScore)

				//游戏记录
				recordStr := fmt.Sprintf("%s，%s, 杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), autoStr, sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
				sp.OnWriteGameRecord(wTargetUser, recordStr)

				//记录杠牌
				sp.addReplayOrder(wTargetUser, info2.E_Gang, int(cbTargetCard))

				//更新亮牌玩家的亮牌数据
				sp.UpdateShowCard(wTargetUser, cbTargetCard)
			}
		}

		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = cbTargetCard
		OperateResult.OperateCode = cbTargetAction
		OperateResult.ProvideUser = sp.ProvideUser
		//OperateResult.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
		sp.LockTimeOut(_userItem.Seat, sp.Rule.Overtime_trust)
		OperateResult.Overtime = sp.LimitTime
		if sp.ProvideUser == static.INVALID_CHAIR {
			OperateResult.ProvideUser = wTargetUser
		}
		//for i := 0; i < sp.GetPlayerCount(); i++ {
		//	OperateResult.PlayerFan[i] = sp.GetPlayerFan(uint16(i))
		//}
		//操作次数记录
		if sp.ProvideUser != static.INVALID_CHAIR {
			//有人点炮的情况下,增加操作用户的操作次数,并保存第三次供牌的用户
			_userItem.Ctx.AddThirdOperate(sp.ProvideUser)
		}

		OperateResult.HaveGang[wTargetUser] = _userItem.Ctx.HaveGang

		if sp.LastOutCardUser == OperateResult.ProvideUser {
			sp.LastOutCardUser = static.INVALID_CHAIR
		}

		sp.bQiangGangScoreSend = true
		//发送消息
		sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
		//操作分变化
		OperateScore.GameScore, OperateScore.GameVitamin = sp.OnSettle(OperateScore.ScoreOffset, consts.EventSettleGaming)
		//发送分数变化
		sp.SendTableMsg(consts.MsgTypeGameOperateScore, OperateScore)

		//设置用户
		sp.CurrentUser = wTargetUser
		sp.ProvideCard = 0
		sp.ProvideUser = static.INVALID_CHAIR
		sp.SendCardData = static.INVALID_BYTE

		//最大操作用户操作的是杠牌，进行杠牌处理
		if cbTargetAction == static.WIK_GANG {
			//没有人能抢杠
			if sp.LeftCardCount > 0 {
				sp.DispatchCardData(wTargetUser, true, true)
			} else {
				sp.ChiHuCard = 0
				sp.ProvideUser = static.INVALID_CHAIR
				sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
				return true
			}
		}

		//如果是吃碰操作，再判断目标用户是否还有杠牌动作动作判断
		if sp.LeftCardCount > 0 && sp.CheckCanHaveGangAction() {
			//杠牌判断
			//var GangCardResult static.TagGangCardResult

			if _item := sp.GetUserItemByChair(sp.CurrentUser); _item != nil {

				//_item.Ctx.UserAction32 |= sp.m_GameLogic.AnalyseGangCard(_item, _item.Ctx.CardIndex,
				//	_item.Ctx.WeaveItemArray[:], _item.Ctx.WeaveItemCount, &GangCardResult,  sp.m_ShowCard)

				_item.Ctx.UserAction32 |= sp.CheckLiangPai(sp.CurrentUser, true)

				////结果处理
				//if GangCardResult.CardCount > 0 {
				//	//设置变量
				//	_item.Ctx.UserAction32 |= static.WIK_GANG
				//	sp.ProvideCard = 0
				//}

				if _item.Ctx.UserAction32 != 0 {
					//发送动作
					sp.SendOperateNotify()
				}
			}
		}
		return true
	}
	return true
}

// ! 主动动作，自己暗杠痞子杠赖子杠续杠胡牌
func (sp *SportYCKWX) OnUserOperateByChair(msg *Msg_C_OperateCard_K5X, _userItem *components2.Player) bool {
	wChairID := sp.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card
	//效验操作
	if cbOperateCode != static.WIK_CHI_HU && _userItem.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
		if sp.m_ShowCard[wChairID].BIsShowCard { //&& _userItem.Ctx.UserAction32 & static.WIK_CHI_HU != 0 {
			//清掉胡权
			_userItem.Ctx.UserAction32 ^= static.WIK_CHI_HU
			_userItem.Ctx.ChiHuResult.ChiHuKind = static.CHK_NULL
			fmt.Println(fmt.Sprintf("玩家%d亮牌后,可胡后选择了弃胡", wChairID))

			if sp.Rule.GHFF {
				//过胡翻番
				fmt.Println(fmt.Sprintf("玩家%d过胡翻番", wChairID))
				sp.GuoHuCount[wChairID]++

				var guohu MSG_S_GuoHu
				guohu.Uid = msg.Id
				for i := 0; i < sp.GetPlayerCount(); i++ {
					guohu.GuoHuCount[i] = sp.GuoHuCount[i]
				}
				sp.SendTableMsg(consts.MsgTypeGameGuoHu, guohu)

				//如果当前玩家选择过,自动将这张牌打出
				if cbOperateCode == static.WIK_NULL {
					_msg := sp.Greate_OutCardmsg(msg.Id, false, sp.SendCardData, false)
					if !sp.OnUserOutCard(_msg, false) {
						sp.OnWriteGameRecord(wChairID, "玩家选择过胡,服务器自动出牌的时候，可能被定时器抢先了")
					}
				}
			}
		}
	}

	//效验操作
	if cbOperateCode == static.WIK_NULL {
		_userItem.Ctx.InitChiHuResult()
		return true //放弃
	}

	//扑克效验
	if (cbOperateCode != static.WIK_NULL) && (cbOperateCode != static.WIK_CHI_HU) && (sp.m_GameLogic.IsValidCard(cbOperateCard) == false) {
		return false
	}

	//如果当前玩家所杠的牌不是摸来的那张牌,连杠也是要中断的,
	if cbOperateCode&static.WIK_GANG != 0 && cbOperateCard != sp.SendCardData {
		sp.LianGangCount = 0
		sp.LianGangCount++
		//虽然选了杠,但是杠的不是当前这张,也是弃杠
		var GangCardResult static.TagGangCardResult
		sp.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, _userItem.Ctx.WeaveItemArray[:],
			_userItem.Ctx.WeaveItemCount, &GangCardResult, sp.m_ShowCard)

		CanXuGang := false
		for _, card := range GangCardResult.CardData {
			if card == sp.SendCardData && _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(sp.SendCardData)] == 1 {
				CanXuGang = true
			}
		}
		// 玩家有续杠和暗杠的牌权但是没有续杠当前这张牌,加入弃杠记录
		if CanXuGang {
			sp.m_GameLogic.AppendGiveUpGang(_userItem, sp.SendCardData)
		}
	}
	//有杠没杠,弃杠
	if cbOperateCode&static.WIK_GANG == 0 && _userItem.Ctx.UserAction32&static.WIK_GANG != 0 {
		if _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(sp.SendCardData)] == 1 {
			//弃续杠
			sp.m_GameLogic.AppendGiveUpGang(_userItem, sp.SendCardData)
		}
	}

	//记录玩家最新的一次操作
	_userItem.Ctx.SetUserLastAction(uint64(msg.Code))

	//设置变量
	autoStr := "服务端自动"
	if msg.ByClient {
		autoStr = "玩家主动"
	}
	sp.SendStatus = true

	//构造结果,向客户端发送操作结果
	var OperateResult Msg_S_OperateResult_K5X
	var OperateScore meta2.Msg_S_OperateScore_K5X

	//执行动作
	switch cbOperateCode {
	case static.WIK_GANG: //杠牌操作
		{
			bAnGang := false

			//变量定义
			cbWeaveIndex := 0xFF

			cbCardIndex := sp.m_GameLogic.SwitchToCardIndex(cbOperateCard)
			if _userItem.Ctx.CardIndex[cbCardIndex] == 1 {
				//续杠
				for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
					cbWeaveKind := _userItem.Ctx.WeaveItemArray[i].WeaveKind
					cbCenterCard := _userItem.Ctx.WeaveItemArray[i].CenterCard
					if (cbCenterCard == cbOperateCard) && (cbWeaveKind == static.WIK_PENG) {
						cbWeaveIndex = int(i)
						break
					}
				}

				//效验动作
				if cbWeaveIndex == 0xFF {
					return false
				}

				//游戏记录
				recordStr := fmt.Sprintf("%s，%s, 蓄杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), autoStr, sp.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sp.OnWriteGameRecord(wChairID, recordStr)

				//记录蓄杠牌
				sp.addReplayOrder(wChairID, info2.E_Gang_XuGand, int(cbOperateCard))

				//计算续杠分数
				for i := 0; i < sp.GetPlayerCount(); i++ {
					_item := sp.GetUserItemByChair(uint16(i))
					if wChairID != _item.GetChairID() {
						payscore := sp.GetScoreOnGang(info2.E_Gang_XuGand) * sp.Rule.DiFen * sp.GetLianGangFan()
						OperateScore.ScoreOffset[_item.GetChairID()] -= payscore
						OperateScore.ScoreOffset[wChairID] += payscore

						sp.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("被续杠减分%d\n", payscore))
						sp.OnWriteGameRecord(wChairID, fmt.Sprintf("续杠加分%d\n", payscore))
					}
				}
				OperateScore.OperateUser = wChairID
				OperateScore.OperateType = info2.E_Gang_XuGand
				bAnGang = false
				//组合扑克
				_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 1, wChairID, byte(cbOperateCode), cbOperateCard)

				//删除扑克
				_userItem.Ctx.CleanCard(cbCardIndex)
			} else {
				//暗杠
				if _userItem.Ctx.CardIndex[cbCardIndex] != 4 {
					return false
				}

				//设置变量
				cbWeaveIndex = int(_userItem.Ctx.WeaveItemCount)
				_userItem.Ctx.WeaveItemCount++
				_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 0, wChairID, byte(cbOperateCode), cbOperateCard)

				_userItem.Ctx.HidGangAction()
				anGangScore := sp.GetScoreOnGang(info2.E_Gang_AnGang) * sp.Rule.DiFen * sp.GetLianGangFan()
				for i := 0; i < sp.GetPlayerCount(); i++ {
					_item := sp.GetUserItemByChair(uint16(i))
					if wChairID != _item.GetChairID() {
						OperateScore.ScoreOffset[_item.GetChairID()] -= anGangScore
						OperateScore.ScoreOffset[wChairID] += anGangScore

						sp.OnUserScoreOffset(wChairID, anGangScore)
						sp.OnUserScoreOffset(_item.GetChairID(), -anGangScore)
					}
				}
				OperateScore.OperateType = info2.E_Gang_AnGang
				bAnGang = true

				//游戏记录
				recordStr := fmt.Sprintf("%s，%s, 暗杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), autoStr, sp.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sp.OnWriteGameRecord(wChairID, recordStr)

				//记录暗杠牌
				sp.addReplayOrder(wChairID, info2.E_Gang_AnGang, int(cbOperateCard))

				//删除扑克
				_userItem.Ctx.CleanCard(cbCardIndex)
			}

			//更新亮牌玩家的亮牌数据
			sp.UpdateShowCard(wChairID, cbOperateCard)

			OperateResult.OperateUser = wChairID
			OperateResult.ProvideUser = wChairID
			OperateResult.HaveGang[wChairID] = _userItem.Ctx.HaveGang //是否杠过
			OperateResult.OperateCode = cbOperateCode
			OperateResult.OperateCard = cbOperateCard
			//for i := 0; i < sp.GetPlayerCount(); i++ {
			//	OperateResult.PlayerFan[i] = sp.GetPlayerFan(uint16(i))
			//}

			//发送消息
			sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
			if bAnGang {
				OperateScore.GameScore, OperateScore.GameVitamin = sp.OnSettle(OperateScore.ScoreOffset, consts.EventSettleGaming)
				//发送分数变化
				sp.SendTableMsg(consts.MsgTypeGameOperateScore, OperateScore)
				if sp.LeftCardCount > 0 {
					sp.DispatchCardData(wChairID, true, false)
				} else {
					sp.ChiHuCard = 0
					sp.ProvideUser = static.INVALID_CHAIR
					sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
				}
			} else { //如果不是暗杠的情况，才分析这张牌其他用户是否可以抢杠,调用的是分析用户响应操作
				bAroseAction := false
				//允许抢杠胡的条件下才分析抢杠胡
				if sp.HuType.HAVE_QIANG_GANG_HU {
					bAroseAction = sp.EstimateUserRespond(wChairID, cbOperateCard, static.EstimatKind_GangCard, true)
				}

				//发送扑克
				if bAroseAction == false {
					//续杠分写记录
					_userItem.Ctx.XuGangAction()
					for i := 0; i < sp.GetPlayerCount(); i++ {
						if wChairID != uint16(i) {
							sp.OnUserScoreOffset(uint16(i), OperateScore.ScoreOffset[i])
						}
					}
					sp.OnUserScoreOffset(wChairID, OperateScore.ScoreOffset[wChairID])
					OperateScore.GameScore, OperateScore.GameVitamin = sp.OnSettle(OperateScore.ScoreOffset, consts.EventSettleGaming)
					//发送分数变化
					sp.SendTableMsg(consts.MsgTypeGameOperateScore, OperateScore)
					if sp.LeftCardCount > 0 {
						sp.DispatchCardData(wChairID, true, false)
					} else {
						sp.ChiHuCard = 0
						sp.ProvideUser = static.INVALID_CHAIR
						sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
					}
				} else {
					//有玩家可以抢杠胡,这时需要把该玩家的续杠分保存下
					sp.SaveXuGangScore(wChairID, OperateScore)
				}
			}
			return true
		}
	case static.WIK_CHI_HU: //吃胡操作,主动状态下没有抢杠的说法，有自摸胡牌，杠上开花胡牌
		{
			//普通胡牌
			sp.ClearChiHuResultByUserReverse(_userItem.GetChairID())
			sp.ProvideCard = sp.SendCardData

			if sp.ProvideCard != 0 {
				sp.ProvideUser = wChairID
			}

			//游戏记录
			recordStr := fmt.Sprintf("%s，%s, 胡牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), autoStr, sp.m_GameLogic.SwitchToCardNameByData(sp.ProvideCard, 1))
			sp.OnWriteGameRecord(wChairID, recordStr)

			//记录胡牌
			sp.addReplayOrder(wChairID, info2.E_Hu, int(sp.ProvideCard))

			//结束信息
			sp.ChiHuCard = sp.ProvideCard
			//结束游戏
			sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)

			return true
		}
	}
	return true
}

// ! 亮牌动作
func (sp *SportYCKWX) OnUserOperateLiangPai(msg *Msg_C_LiangPai) bool {

	wChairID := sp.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOutCard := msg.OutCard

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	if (_userItem.Ctx.UserAction32 & cbOperateCode) == 0 {
		return false //无效操作
	}

	//校验数据
	if len(msg.LiangStruct.TingCard) <= 0 || len(msg.LiangStruct.LiangCard) <= 0 {
		return false
	}
	//校验数据 听牌必须有数据,不然亮牌失败
	bCanLiang := false
	for _, card := range msg.LiangStruct.TingCard {
		if card != 0 {
			if sp.m_GameLogic.IsValidCard(card) {
				bCanLiang = true
			}
		}
		if bCanLiang {
			break
		}
		//游戏记录
		recordStr := fmt.Sprintf("客户端亮牌发过来所听的牌: %s 没有有效可听牌,亮牌失败!", sp.m_GameLogic.SwitchToCardNameByDatas(msg.LiangStruct.TingCard[:], 1))
		sp.OnWriteGameRecord(wChairID, recordStr)
		return false
	}

	////记录玩家最新的一次操作
	//_userItem.Ctx.SetUserLastAction(uint64(msg.Code))

	// 回放中记录亮牌牌权操作
	sp.addReplayOrder(wChairID, info2.E_HandleCardRight, msg.Code)

	//变量定义
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, _userItem.Ctx.CardIndex[:])

	//先把要打出的牌减掉
	cbCardIndexTemp[sp.m_GameLogic.SwitchToCardIndex(msg.OutCard)]--
	if cbOperateCode == static.WIK_LIANGPAI && _userItem.Ctx.UserAction32&static.WIK_LIANGPAI != 0 {
		//游戏记录
		recordStr := fmt.Sprintf("玩家选择亮牌,打出:%s, 暗铺%s , 亮牌:%s, 听牌:%s",
			sp.m_GameLogic.SwitchToCardNameByData(msg.OutCard, 1),
			sp.m_GameLogic.SwitchToCardNameByDatas(msg.LiangStruct.AnPuCard[:], 1),
			sp.m_GameLogic.SwitchToCardNameByDatas(msg.LiangStruct.LiangCard[:], 1),
			sp.m_GameLogic.SwitchToCardNameByDatas(msg.LiangStruct.TingCard[:], 1))
		sp.OnWriteGameRecord(wChairID, recordStr)
		////校验亮牌
		//if !sp.CheckCardInHand(_userItem, msg.LiangStruct.LiangCard[:]){
		//	fmt.Println("亮牌不在手牌中。。。有毒")
		//	return false
		//}
		////校验暗铺牌
		//if !sp.CheckCardInHand(_userItem, msg.LiangStruct.AnPuCard[:]){
		//	fmt.Println("暗铺牌不在手牌中。。。有毒")
		//	return false
		//}

		//校验玩家是不是听这些张
		for _, cbCurrentCard := range msg.LiangStruct.TingCard {
			//变量定义
			if cbCurrentCard != 0 {
				var ChiHuResult static.TagChiHuResult
				cbHuCardKind := sp.m_GameLogic.AnalyseChiHuCard(cbCardIndexTemp, _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount,
					cbCurrentCard, 0, &ChiHuResult, true, true)

				//判断客户端发过来的听牌是否可以拿来胡
				if cbHuCardKind == static.CHK_NULL {
					fmt.Println(fmt.Sprintf("客户端亮牌发过来所听的牌%x不能胡: ", cbCurrentCard))
					//游戏记录
					recordStr := fmt.Sprintf("客户端亮牌发过来所听的牌: %s 不能胡,亮牌失败!", sp.m_GameLogic.SwitchToCardNameByData(cbCurrentCard, 1),
						msg.LiangStruct.LiangCard, msg.LiangStruct.TingCard)
					sp.OnWriteGameRecord(wChairID, recordStr)
					return false
				}
			}
		}

		//广播给所有玩家
		sp.SendTableMsg(consts.MsgTypeGameLiangPai, msg)
		//回放中记录玩家亮牌
		sp.addLiangPaiReplayOrder(wChairID, msg)
		//记录玩家亮牌状态
		sp.m_ShowCard[wChairID].BIsShowCard = true
		sp.m_ShowCard[wChairID].CbTingCard = msg.LiangStruct.TingCard
		sp.m_ShowCard[wChairID].CbAnPuCard = msg.LiangStruct.AnPuCard
		sp.m_ShowCard[wChairID].CbLiangCard = msg.LiangStruct.LiangCard
		//出牌
		_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbOutCard, true)
		sp.OnUserOutCard(_msg, true)
		return true
	}
	return false
}

// ! 操作牌
func (sp *SportYCKWX) operateCard(cbTargetAction byte, cbTargetCard byte, _userItem *components2.Player) {
	var cbRemoveCard []byte
	var wik_kind int

	//变量定义
	switch cbTargetAction {
	case static.WIK_LEFT: //上牌操作
		cbRemoveCard = []byte{cbTargetCard + 1, cbTargetCard + 2}
		wik_kind = info2.E_Wik_Left

		//游戏记录
		recordStr := fmt.Sprintf("%s，左吃牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sp.OnWriteGameRecord(_userItem.Seat, recordStr)

	case static.WIK_RIGHT:
		cbRemoveCard = []byte{cbTargetCard - 2, cbTargetCard - 1}
		wik_kind = info2.E_Wik_Right

		//游戏记录
		recordStr := fmt.Sprintf("%s，右吃牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sp.OnWriteGameRecord(_userItem.Seat, recordStr)
	case static.WIK_CENTER:
		cbRemoveCard = []byte{cbTargetCard - 1, cbTargetCard + 1}
		wik_kind = info2.E_Wik_Center

		//游戏记录
		recordStr := fmt.Sprintf("%s，中吃牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sp.OnWriteGameRecord(_userItem.Seat, recordStr)
	case static.WIK_PENG: //碰牌操作
		cbRemoveCard = []byte{cbTargetCard, cbTargetCard}
		wik_kind = info2.E_Peng

		//游戏记录
		recordStr := fmt.Sprintf("%s，碰牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sp.OnWriteGameRecord(_userItem.Seat, recordStr)
	default:
	}

	//吃碰过手可以胡牌,不需要过庄
	_userItem.Ctx.NeedGuoZhuang = false

	//记录左吃
	sp.addReplayOrder(_userItem.Seat, wik_kind, int(cbTargetCard))
	//删除扑克
	_userItem.Ctx.RemoveCards(&sp.Rule, cbRemoveCard)

	//sp.LockTimeOut(_userItem.Seat, static.GAME_OPERATION_TIME_12)
	sp.LockTimeOut(_userItem.Seat, sp.Rule.Overtime_trust)
}

// ! 用户出牌
func (sp *SportYCKWX) OnUserOutCard(msg *Msg_C_OutCard_K5X, needlock bool) bool {
	xlog.Logger().Debug("OnUserOutCard")

	if sp.Rule.NineSecondRoom && needlock {
		sp.OperateMutex.Lock()
		defer sp.OperateMutex.Unlock()
	}
	//效验状态
	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}
	wChairID := sp.GetChairByUid(msg.Id)
	//效验参数
	if sp.m_GameLogic.IsValidCard(msg.CardData) == false {
		recordStr := fmt.Sprintf("打出：%s, 无效牌,出牌失败,客户端发来数据是:%v", sp.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1), msg.CardData)
		sp.OnWriteGameRecord(wChairID, recordStr)
		return false
	}

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return true
	}

	if wChairID != sp.CurrentUser {
		recordStr := fmt.Sprintf("当前牌权玩家是%d , 出牌玩家是%d, 没有牌权,出牌失败", sp.CurrentUser, wChairID)
		sp.OnWriteGameRecord(wChairID, recordStr)
		return false
	}
	//如果是续杠杠权选择弃，加入弃杠记录
	if _userItem.Ctx.UserAction32&static.WIK_GANG != 0 {
		var GangCardResult static.TagGangCardResult
		sp.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, _userItem.Ctx.WeaveItemArray[:],
			_userItem.Ctx.WeaveItemCount, &GangCardResult, sp.m_ShowCard)

		CanXuGang := false
		for _, card := range GangCardResult.CardData {
			if card == sp.SendCardData && _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(sp.SendCardData)] == 1 {
				CanXuGang = true
			}
		}
		// 玩家有续杠和暗杠的牌权但是没有续杠当前这张牌,加入弃杠记录
		if CanXuGang {
			sp.m_GameLogic.AppendGiveUpGang(_userItem, sp.SendCardData)
		}
	}

	// 解锁用户超时操作
	sp.UnLockTimeOut(wChairID)

	//出牌丢进弃牌区
	//_userItem.Ctx.Discard(msg.CardData)
	var TuoGuanType byte = 0
	if _userItem.CheckTRUST() {
		TuoGuanType = OutCard_Type_ByServer
	} else {
		if sp.m_ShowCard[wChairID].BIsShowCard {
			TuoGuanType = OutCard_Type_LiangDao
		} else {
			if msg.ByClient {
				TuoGuanType = OutCard_type_ByUser
			} else {
				TuoGuanType = OutCard_Type_NULL
			}
		}
	}
	_userItem.Ctx.Discard_ex(msg.CardData, TuoGuanType)

	//删除扑克
	if !_userItem.Ctx.OutCard(&sp.Rule, msg.CardData) {
		xlog.Logger().Debug("removecard failed")
		recordStr := fmt.Sprintf("手牌:%s, 删除手牌失败：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1))
		sp.OnWriteGameRecord(wChairID, recordStr)
		return true
	}

	//游戏记录
	autoStr := "服务端自动"
	if msg.ByClient {
		autoStr = "玩家主动"
	}
	recordStr := fmt.Sprintf("%s，%s , 打出：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), autoStr, sp.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1))
	sp.OnWriteGameRecord(wChairID, recordStr)

	//设置变量
	sp.SendStatus = true

	//出牌记录
	sp.OutCardCount++
	////////////////出的是一般的牌或赖子牌////////////////
	sp.OutCardUser = wChairID
	sp.LastOutCardUser = wChairID
	sp.OutCardData = msg.CardData

	//构造数据
	//var OutCard static.Msg_S_OutCard
	var OutCard Msg_S_OutCard_K5X
	OutCard.User = int(wChairID)
	OutCard.Data = msg.CardData
	OutCard.ByClient = msg.ByClient
	OutCard.IsLiang = msg.IsLiang
	OutCard.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
	//记录出牌
	if _userItem.CheckTRUST() {
		OutCard.TuoGuanType = OutCard_Type_ByServer
		sp.addReplayOrder(wChairID, info2.E_OutCard_TG, int(msg.CardData))
	} else {
		if sp.m_ShowCard[wChairID].BIsShowCard {
			OutCard.TuoGuanType = OutCard_Type_LiangDao
		} else {
			if msg.ByClient {
				OutCard.TuoGuanType = OutCard_type_ByUser
			} else {
				OutCard.TuoGuanType = OutCard_Type_NULL
			}
		}
		sp.addReplayOrder(wChairID, info2.E_OutCard, int(msg.CardData))
	}

	//发送消息
	sp.SendTableMsg(consts.MsgTypeGameOutCard, OutCard)
	//用户切换
	sp.ProvideUser = wChairID
	sp.ProvideCard = msg.CardData
	sp.CurrentUser = uint16(sp.GetNextSeat(wChairID))

	//响应判断，如果用户出的是一般牌，判断其他用户是否需要该牌，EstimatKind_OutCard只是正常出牌判断
	//如果当前用户自己 出了牌，不能自己对自己进行分析吃，碰杠
	bAroseAction := false
	bAroseAction = sp.EstimateUserRespond(wChairID, msg.CardData, static.EstimatKind_OutCard, false)

	//打了牌，别人没有反应 流局
	if bAroseAction == false {
		//清除玩家最新的一次操作,杠上炮相关
		_userItem.Ctx.CleanUserLastAction()
		if sp.LeftCardCount > 0 {
			// 发牌
			sp.DispatchCardData(sp.CurrentUser, false, false)
		} else {
			// 游戏结束
			sp.ChiHuCard = 0
			sp.ProvideUser = static.INVALID_CHAIR
			sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
		}
	}

	return true
}

// ! 亮倒自动出牌
func (sp *SportYCKWX) OnLiangAutoOperate(wChairID uint16, bBreakin bool) {

	if bBreakin == false {
		return
	}

	if !sp.Rule.NineSecondRoom {
		sp.UnLockTimeOut(wChairID)
		return
	}

	if sp.GetGameStatus() == static.GS_MJ_FREE {
		sp.UnLockTimeOut(wChairID)
		return
	}

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		sp.UnLockTimeOut(wChairID)
		return
	}
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	if (_userItem.Ctx.UserAction32&static.WIK_CHI_HU) != 0 && sp.Rule.GHFF {
		//托管状态下,勾选了过胡翻番也自动胡
		if _userItem.CheckTRUST() {
			_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_CHI_HU, sp.ProvideCard)
			if !sp.OnUserOperateCard(_msg, true) {
				sp.OnWriteGameRecord(wChairID, "服务器自动选胡的时候，可能被客户端抢先了")
			}
		} else {
			fmt.Println(fmt.Sprintf("过胡翻番,有胡牌自己选,系统不代打"))
			//游戏记录
			recordStr := fmt.Sprintf("过胡翻番,有胡牌自己选,系统不代打，手牌%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1))
			sp.OnWriteGameRecord(wChairID, recordStr)
		}
		return
	}

	//能胡吃胡
	if (_userItem.Ctx.UserAction32&static.WIK_CHI_HU) != 0 && !sp.Rule.GHFF { //&& sp.CurrentUser != wChairID {
		_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_CHI_HU, sp.ProvideCard)
		if !sp.OnUserOperateCard(_msg, true) {
			sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
		}
		return
	}

	//点杠
	if _userItem.Ctx.UserAction32&static.WIK_GANG != 0 {
		if sp.CurrentUser == static.INVALID_CHAIR {
			_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_GANG, sp.ProvideCard)
			if !sp.OnUserOperateCard(_msg, true) {
				sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
			}
			return
		} else if sp.CurrentUser == wChairID {
			//这时候可能会抢杠胡
			bIsXuGang := false
			for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
				if _userItem.Ctx.WeaveItemArray[i].WeaveKind == static.WIK_PENG {
					if _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(sp.SendCardData)] == 1 {
						bIsXuGang = true
					}
				}
			}
			//被抢杠胡的牌直接打出
			if sp.IsOthersTingHuCard(wChairID, sp.ProvideCard) && bIsXuGang {
				cbSendCardData := sp.SendCardData
				index := sp.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
				if index >= 0 && index < static.MAX_INDEX {
					if 0 != _userItem.Ctx.CardIndex[index] {
						_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData, false)
						if !sp.OnUserOutCard(_msg, true) {
							sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
						}
						return
					}
				}
			} else {
				_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_GANG, sp.ProvideCard)
				if !sp.OnUserOperateCard(_msg, true) {
					sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
				}
				return
			}
		}
	}

	//if sp.CurrentUser == static.INVALID_CHAIR && _userItem.Ctx.UserAction32 != 0 {
	//	_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_NULL, sp.ProvideCard)
	//	if !sp.OnUserOperateCard(_msg, true){
	//		sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
	//	}
	//	//sp.OnWriteGameRecord(wChairID, "服务端自动操作 ： 选择弃！")
	//	return
	//}

	//暗杠 擦炮直接放弃出牌
	if sp.CurrentUser == wChairID {
		cbSendCardData := sp.SendCardData
		index := sp.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
		if index >= 0 && index < static.MAX_INDEX {
			if 0 != _userItem.Ctx.CardIndex[index] {
				_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData, false)
				if !sp.OnUserOutCard(_msg, true) {
					sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
				}
				return
			}
		}
	}
}

// 托管自动出牌
func (sp *SportYCKWX) OnTrustAutoOperate(wChairID uint16, bBreakin bool) {

	if bBreakin == false {
		return
	}

	if !sp.Rule.NineSecondRoom {
		sp.UnLockTimeOut(wChairID)
		return
	}

	if sp.GetGameStatus() == static.GS_MJ_FREE {
		sp.UnLockTimeOut(wChairID)
		return
	}

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		sp.UnLockTimeOut(wChairID)
		return
	}
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	//处理下跑
	if sp.PayPaoStatus {
		if !_userItem.Ctx.UserPaoReady {
			_msg := sp.Greate_XiaPaomsg(_userItem.Uid, false, _userItem.Ctx.VecXiaPao.Status)
			if !sp.OnUserClientXiaPao(_msg) {
				sp.OnWriteGameRecord(_userItem.Seat, "服务器自动选飘时，可能被客户端抢先了")
			}
			recordStr := fmt.Sprintf("玩家%s 未选漂,超时托管自动选漂为%d ", _userItem.Name, _msg.Num)
			sp.OnWriteGameRecord(wChairID, recordStr)
		}
		return
	}

	//能胡 胡牌 吃胡
	if (_userItem.Ctx.UserAction32 & static.WIK_CHI_HU) != 0 {
		var _msg *Msg_C_OperateCard_K5X
		card := byte(0xFF)
		if sp.CurrentUser == wChairID {
			card = sp.SendCardData
		} else if sp.CurrentUser == static.INVALID_CHAIR {
			card = sp.ProvideCard
		}
		if sp.Rule.GHFF && sp.m_ShowCard[wChairID].BIsShowCard {
			_msg = sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_CHI_HU, card)
			recordStr := fmt.Sprintf("玩家亮倒托管,过胡翻番,有胡权,系统自动胡，手牌%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1))
			sp.OnWriteGameRecord(wChairID, recordStr)
			if !sp.OnUserOperateCard(_msg, true) {
				sp.OnWriteGameRecord(wChairID, "服务器自动操作的时候，可能被客户端抢先了")
			}
			return
		} else {
			if sp.CurrentUser == wChairID {
				sp.TrustOutCard(wChairID)
				return
			} else {
				_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_NULL, sp.ProvideCard)
				if !sp.OnUserOperateCard(_msg, true) {
					sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
				}
				return
			}
		}
	}
	//其它操作全放弃
	if _userItem.Ctx.UserAction32 != 0 {
		if sp.CurrentUser == wChairID {
			sp.TrustOutCard(wChairID)
			return
		} else {
			_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_NULL, sp.ProvideCard)
			if !sp.OnUserOperateCard(_msg, true) {
				sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
			}
			return
		}
	}

	//暗杠 擦炮直接放弃出牌
	sp.TrustOutCard(wChairID)

}

// 创建消息
func (sp *SportYCKWX) Greate_XiaPaomsg(Id int64, byClient bool, status bool) *static.Msg_C_Xiapao {
	_msg := new(static.Msg_C_Xiapao)
	_msg.Num = 0
	_msg.Status = status
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建操作牌消息
func (sp *SportYCKWX) Greate_Operatemsg(Id int64, byClient bool, Code byte, Card byte) *Msg_C_OperateCard_K5X {
	_msg := new(Msg_C_OperateCard_K5X)
	_msg.Card = Card
	_msg.Code = int(Code)
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建出牌消息
func (sp *SportYCKWX) Greate_OutCardmsg(Id int64, byClient bool, Card byte, bLiang bool) *Msg_C_OutCard_K5X {
	//_msg := new(static.Msg_C_OutCard)
	_msg := new(Msg_C_OutCard_K5X)
	_msg.CardData = Card
	_msg.Id = Id
	_msg.ByClient = byClient
	_msg.IsLiang = bLiang
	return _msg
}

// ! 派发扑克
func (sp *SportYCKWX) DispatchCardData(wCurrentUser uint16, bGangFlower bool, bDianGang bool) bool {
	if sp.IsPausing() {
		sp.CurrentUser = static.INVALID_CHAIR
		sp.SetSendCardOpt(static.TagSendCardInfo{
			CurrentUser: wCurrentUser,
			GangFlower:  bGangFlower,
		})
		return true
	}
	//状态效验
	if wCurrentUser == static.INVALID_CHAIR {
		return false
	}

	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return false
	}

	//吃碰过手可以胡牌,不需要过庄
	_userItem.Ctx.NeedGuoZhuang = false

	//剩余牌校验
	if sp.LeftCardCount <= 0 {
		return false
	}

	bEnjoinHu := true
	//发牌处理
	if sp.SendStatus == true {
		sp.CheckPlayerWantCard(_userItem)
		// 设置开始超时操作
		//sp.LockTimeOut(wCurrentUser, static.GAME_OPERATION_TIME_12)
		sp.SendCardCount++
		sp.LeftCardCount--
		sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount]
		sp.CheckWantCard(sp.SendCardData)

		_userItem.Ctx.DispatchCard(sp.SendCardData)
		if bGangFlower {
			sp.HaveGangCard = true
		}
		sp.SetLeftCardArray()
		//游戏记录
		recordStr := fmt.Sprintf("牌型%s，发来：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(sp.SendCardData, 1))
		sp.OnWriteGameRecord(wCurrentUser, recordStr)

		//记录发牌
		sp.addReplayOrder(wCurrentUser, info2.E_SendCard, int(sp.SendCardData))

		//设置变量
		sp.ProvideUser = wCurrentUser
		sp.ProvideCard = sp.SendCardData
		//给用户发牌后，判断用户是否可以杠牌
		_userItem.Ctx.ClearOperateCard32()
		if sp.LeftCardCount > 0 && sp.CheckCanHaveGangAction() {
			var GangCardResult static.TagGangCardResult
			//玩家亮出的牌不能再杠了
			_userItem.Ctx.UserAction32 |= int(sp.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex,
				_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult, sp.m_ShowCard))
			//连杠必须是摸来的那张牌又凑成杠了才算
			bIsCurCardGang := false
			for _, card := range GangCardResult.CardData {
				if card == sp.SendCardData {
					bIsCurCardGang = true
				}
			}
			if _userItem.Ctx.UserAction32&static.WIK_GANG != 0 && !bIsCurCardGang {
				sp.LianGangCount = 0
			}

			//点杠,杠后摸来的牌可以凑成杠的才继续给杠权
			if bDianGang && !bIsCurCardGang {
				if _userItem.Ctx.UserAction32&static.WIK_GANG != 0 {
					_userItem.Ctx.UserAction32 ^= static.WIK_GANG
				}
			}

			_userItem.Ctx.UserAction32 |= sp.CheckLiangPai(wCurrentUser, true)
		}

		// 判断是否胡牌
		sp.initChiHuResult()
		sp.CheckHu(wCurrentUser, wCurrentUser, sp.SendCardData, bGangFlower, false, false, false, true)
	}

	//设置变量
	sp.OutCardData = 0
	sp.CurrentUser = wCurrentUser
	sp.OutCardUser = static.INVALID_CHAIR

	//构造数据
	var SendCard static.Msg_S_SendCard32
	SendCard.CurrentUser = wCurrentUser
	SendCard.ActionMask = _userItem.Ctx.UserAction32
	SendCard.CardData = 0x00
	if sp.SendStatus {
		SendCard.CardData = sp.SendCardData
	}
	SendCard.IsGang = bGangFlower
	SendCard.IsHD = false
	SendCard.EnjoinHu = bEnjoinHu
	if sp.m_ShowCard[wCurrentUser].BIsShowCard {
		if sp.Rule.Overtime_trust > 0 && _userItem.Ctx.UserAction32&static.WIK_CHI_HU != 0 && sp.Rule.GHFF {
			sp.LockTimeOut(wCurrentUser, sp.Rule.Overtime_trust)
			SendCard.Overtime = sp.LimitTime
		} else {
			if sp.m_ShowCard[wCurrentUser].BIsShowCard && sp.Rule.GHFF && sp.Rule.Overtime_trust > 0 && _userItem.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
				sp.LockTimeOut(wCurrentUser, sp.Rule.Overtime_trust)
				SendCard.Overtime = sp.LimitTime
			} else {
				SendCard.Overtime = time.Now().Unix() + static.GAME_LIANGPAI_AUTOOUT_TIME
			}
		}
	} else {
		//SendCard.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
		sp.LockTimeOut(wCurrentUser, sp.Rule.Overtime_trust)
		SendCard.Overtime = sp.LimitTime
	}
	SendCard.VecGangCard = static.HF_BytesToInts(_userItem.Ctx.VecGangCard)
	sp.LastSendCardUser = wCurrentUser

	// 设置开始超时操作
	if sp.m_ShowCard[wCurrentUser].BIsShowCard {
		if sp.m_ShowCard[wCurrentUser].BIsShowCard && sp.Rule.GHFF && sp.Rule.Overtime_trust > 0 && _userItem.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
			sp.LockTimeOut(wCurrentUser, sp.Rule.Overtime_trust)
		} else {
			sp.LockTimeOut(wCurrentUser, static.GAME_LIANGPAI_AUTOOUT_TIME)
		}
	} else {
		//sp.LockTimeOut(wCurrentUser, static.GAME_OPERATION_TIME_12)
		sp.LockTimeOut(wCurrentUser, sp.Rule.Overtime_trust)
	}

	for _, v := range sp.PlayerInfo {
		if v.GetChairID() != wCurrentUser {
			SendCard.CardData = 0x00
		} else {
			SendCard.CardData = sp.SendCardData
		}
		sp.SendPersonMsg(consts.MsgTypeGameSendCard, SendCard, uint16(v.GetChairID()))
	}

	//游戏记录
	recordStr := fmt.Sprintf("发送牌权：%x", _userItem.Ctx.UserAction32)
	sp.OnWriteGameRecord(wCurrentUser, recordStr)

	sp.HaveGangCard = false
	// 回放记录中记录牌权显示
	if _userItem.Ctx.UserAction32 > 0 {
		sp.addReplayOrder(wCurrentUser, info2.E_SendCardRight, _userItem.Ctx.UserAction32)
	}

	return true
}

// ! 响应判断
func (sp *SportYCKWX) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int, isAnalyseQGH bool) bool {
	//变量定义
	bAroseAction := false

	// 响应判断只需要判断出牌以及续杠
	if EstimatKind != static.EstimatKind_OutCard && EstimatKind != static.EstimatKind_GangCard {
		return bAroseAction
	}

	//用户状态
	for _, v := range sp.PlayerInfo {
		v.Ctx.ClearOperateCard32()
	}

	bCanLianGang := false
	//动作判断
	for i := 0; i < sp.GetPlayerCount(); i++ {
		//用户过滤
		if wCenterUser == uint16(i) {
			continue
		}

		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//出牌类型检验
		if EstimatKind == static.EstimatKind_OutCard {
			//吃碰判断
			//碰牌判断,亮牌之后不能再碰了
			if !sp.m_ShowCard[i].BIsShowCard {
				_item.Ctx.UserAction32 |= int(sp.m_GameLogic.EstimatePengCard(_item.Ctx.CardIndex, cbCenterCard))
			}
			if sp.LeftCardCount > 0 {
				//杠牌判断,亮出的牌不能再参与杠了
				_item.Ctx.UserAction32 |= sp.m_GameLogic.EstimateGangCard(_item, cbCenterCard, sp.m_ShowCard)
			}

			if _item.Ctx.UserAction32&static.WIK_GANG != 0 {
				bCanLianGang = true
			}
		}

		bQiangGang := false
		if EstimatKind == static.EstimatKind_GangCard {
			bQiangGang = true
		}

		//杠上炮
		bGangPao := false
		if gangpaouseritem := sp.GetUserItemByChair(wCenterUser); gangpaouseritem != nil {
			if gangpaouseritem.Ctx.LastUserAction == static.WIK_GANG && EstimatKind == static.EstimatKind_OutCard {
				bGangPao = true
			}
		}

		sp.CheckHu(uint16(i), wCenterUser, cbCenterCard, false, bQiangGang, true, bGangPao, true)
		//结果判断
		if _item.Ctx.UserAction32 != static.WIK_NULL {
			bAroseAction = true
		}
	}
	//如果打出的牌玩家不能杠,那连杠就终止了
	if !bCanLianGang && !isAnalyseQGH {
		sp.LianGangCount = 0
	}

	//结果处理，标志为真说明上一个用户所出的牌有人可以所以要发送一个提示信息的框
	if bAroseAction == true {
		//设置变量
		sp.ProvideUser = uint16(wCenterUser)
		sp.ProvideCard = cbCenterCard
		sp.ResumeUser = sp.CurrentUser
		sp.CurrentUser = static.INVALID_CHAIR

		//发送提示
		sp.SendOperateNotify()

		return true
	}

	return false
}

// ! 发送操作
func (sp *SportYCKWX) SendOperateNotify() bool {
	//发送提示
	for _, v := range sp.PlayerInfo {
		if v.Ctx.UserAction32 != static.WIK_NULL {
			//构造数据
			var OperateNotify static.Msg_S_OperateNotify32
			OperateNotify.ResumeUser = sp.ResumeUser
			//抢暗杠时，复用此字段，表示轮到谁抢了
			OperateNotify.ActionCard = sp.ProvideCard
			OperateNotify.ActionMask = v.Ctx.UserAction32
			OperateNotify.EnjoinHu = false
			OperateNotify.VecGangCard = static.HF_BytesToInts(v.Ctx.VecGangCard)
			if sp.m_ShowCard[v.Seat].BIsShowCard {
				if sp.Rule.GHFF && sp.Rule.Overtime_trust > 0 && v.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
					sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
					OperateNotify.Overtime = sp.LimitTime
				} else {
					sp.LockTimeOut(v.Seat, static.GAME_LIANGPAI_AUTOOUT_TIME)
					OperateNotify.Overtime = sp.LimitTime
				}
			} else {
				//OperateNotify.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
				sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
				OperateNotify.Overtime = sp.LimitTime
			}

			//发送数据
			//抢的牌权需要发送给所有玩家，因为其他玩家需要知道轮到谁抢暗杠了
			if v.Ctx.UserAction32 == static.WIK_QIANG {
				OperateNotify.ActionCard = byte(v.Seat)
				sp.SendTableMsg(consts.MsgTypeGameOperateNotify, OperateNotify)
			} else {
				sp.SendPersonMsg(consts.MsgTypeGameOperateNotify, OperateNotify, v.Seat)

				if sp.m_ShowCard[v.Seat].BIsShowCard {
					if sp.Rule.GHFF && sp.Rule.Overtime_trust > 0 && v.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
						sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
					} else {
						sp.LockTimeOut(v.Seat, static.GAME_LIANGPAI_AUTOOUT_TIME)
					}
				} else {
					//sp.LockTimeOut(v.Seat, static.GAME_OPERATION_TIME_12)
					sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
				}
			}

			// 游戏记录
			recrodStr := fmt.Sprintf("发送牌权：%x", v.Ctx.UserAction32)
			sp.OnWriteGameRecord(v.Seat, recrodStr)

			// 回放记录中记录牌权显示
			sp.addReplayOrder(v.Seat, info2.E_SendCardRight, v.Ctx.UserAction32)
		}
	}

	return true
}

// ! 增加回放操作记录
func (sp *SportYCKWX) addReplayOrder(chairId uint16, operation int, card int) {
	var order meta2.K5x_Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	sp.ReplayRecord.VecOrder = append(sp.ReplayRecord.VecOrder, order)
}

// ! 增加亮牌回放操作记录
func (sp *SportYCKWX) addLiangSturctReplayOrder(chairId uint16, operation int, s string) {
	var order meta2.K5x_Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.LiangPai = s
	sp.ReplayRecord.VecOrder = append(sp.ReplayRecord.VecOrder, order)
}

// !亮牌
func (sp *SportYCKWX) addLiangPaiReplayOrder(chairId uint16, msg *Msg_C_LiangPai) {
	str := "LP"
	//亮牌
	for i := 0; i < len(msg.LiangStruct.LiangCard); i++ {
		if msg.LiangStruct.LiangCard[i] != 0 {
			str += fmt.Sprintf("%02x#", msg.LiangStruct.LiangCard[i])
		}
	}
	str += "|"
	//暗铺
	for i := 0; i < len(msg.LiangStruct.AnPuCard); i++ {
		if msg.LiangStruct.AnPuCard[i] != 0 {
			str += fmt.Sprintf("%02x#", msg.LiangStruct.AnPuCard[i])
		}
	}
	//听牌
	str += "|"
	for i := 0; i < len(msg.LiangStruct.TingCard); i++ {
		if msg.LiangStruct.TingCard[i] != 0 {
			str += fmt.Sprintf("%02x#", msg.LiangStruct.TingCard[i])
		}
	}
	str += ","
	sp.addLiangSturctReplayOrder(chairId, info2.E_K5x_LiangPai, str)
}

// ! 检查是否能胡
func (sp *SportYCKWX) CheckHu(wCurrentUser uint16, wProvideUser uint16, cbCurrentCard byte, bGangFlower bool, bQiangGang bool, bDianPao bool, bGangPao bool, checkk5 bool) int {
	sp.ClearChiHuResultByUser(wCurrentUser)

	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return static.WIK_NULL
	}

	//亮牌之后只能胡听牌的那几张
	if sp.m_ShowCard[wCurrentUser].BIsShowCard {
		bCanCheckHu := false
		for _, card := range sp.m_ShowCard[wCurrentUser].CbTingCard {
			if card != 0 {
				if cbCurrentCard != 0 && cbCurrentCard == card {
					bCanCheckHu = true
					break
				}
			}
		}

		if !bCanCheckHu {
			fmt.Println(fmt.Sprintf("玩家%d亮牌之后只能胡所听的牌%v", wCurrentUser, sp.m_ShowCard[wCurrentUser].CbTingCard))

			return static.WIK_NULL
		}
	}

	//牌型权位
	wChiHuRight := uint64(0)
	//杠开权限判断
	if bGangFlower {
		wChiHuRight |= static.CHR_GANG_SHANG_KAI_HUA
	}
	//杠上炮权限判断
	if bGangPao {
		wChiHuRight |= static.CHR_GANG_SHANG_PAO
	}
	//抢杠胡判断权限
	if bQiangGang {
		wChiHuRight |= static.CHR_QIANG_GANG
	}
	////海底权限判断
	//if sp.LeftCardCount <= sp.GetHaidiCount() && !bGangFlower {
	//	if !bDianPao{
	//		//海底捞
	//		wChiHuRight |= static.CHR_HAI_DI
	//	}else{
	//		//海底炮
	//		wChiHuRight |= static.CHR_HAI_DI_PAO
	//	}
	//}

	//给用户发牌后，胡牌判断
	_userItem.Ctx.UserAction32 |= int(sp.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:],
		_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, cbCurrentCard, wChiHuRight, &_userItem.Ctx.ChiHuResult, bDianPao, checkk5))

	//2分起胡
	if _userItem.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
		if !sp.CheckFan(wCurrentUser, wProvideUser) {
			_userItem.Ctx.UserAction32 ^= static.WIK_CHI_HU
			_userItem.Ctx.ChiHuResult.ChiHuKind = static.CHK_NULL
			fmt.Println(fmt.Sprintf("玩家%d胡分不够,不能胡", wCurrentUser))
		}
	}

	//需要过庄才能胡,胡牌牌型变大不需要过庄
	if _userItem.Ctx.NeedGuoZhuang &&
		!(sp.GetPlayerDaHuCount(_userItem.Ctx.ChiHu) < sp.GetPlayerDaHuCount(_userItem.Ctx.ChiHuResult.ChiHuKind)) {
		if (_userItem.Ctx.UserAction32 & static.WIK_CHI_HU) != 0 {
			_userItem.Ctx.UserAction32 ^= static.WIK_CHI_HU
			_userItem.Ctx.ChiHuResult.ChiHuKind = static.CHK_NULL
			sp.SendGameNotificationMessage(_userItem.GetChairID(), "未过庄不能胡")
		}
		return static.WIK_NULL
	}
	//保存玩家上一次可胡牌类型
	_userItem.Ctx.ChiHu = _userItem.Ctx.ChiHuResult.ChiHuKind
	return _userItem.Ctx.UserAction32
}

// ! 检查是否杠牌
func (sp *SportYCKWX) CheckCanHaveGangAction() bool {

	if sp.LeftCardCount >= 1 {
		return true
	}

	return false
}

// ! 单局结算
func (sp *SportYCKWX) OnGameOver(wChairID uint16, cbReason byte) bool {
	sp.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 游戏结束, 流局结束，统计积分
func (sp *SportYCKWX) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if sp.GetGameStatus() == static.GS_MJ_END && cbReason == static.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	// 清除超时检测
	for _, v := range sp.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}

	switch cbReason {
	case static.GER_NORMAL: //常规结束
		return sp.OnGameEndNormal(wChairID, cbReason)
	case static.GER_USER_LEFT: //用户强退
		return sp.OnGameEndUserLeft(wChairID, cbReason)
	case static.GER_DISMISS: //解散游戏
		return sp.OnGameEndDissmiss(wChairID, cbReason, 0)
	case static.GER_GAME_ERROR:
		return sp.OnGameEndDissmiss(wChairID, cbReason, 1)
	}
	return false
}

// ! 结束，结束游戏
func (sp *SportYCKWX) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	//设置游戏结束状态
	sp.SetGameStatus(static.GS_MJ_END)

	//定义变量
	var GameEnd static.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sp.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.MagicCard = sp.MagicCard

	//设置承包用户
	GameEnd.Contractor = static.INVALID_CHAIR
	GameEnd.ProvideUser = wChairID
	GameEnd.ChiHuCard = sp.ChiHuCard
	GameEnd.ChiHuUserCount = 1
	GameEnd.KaiKou = sp.Rule.KouKou
	for i := 0; i < int(sp.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}
	for i := 0; i < 4; i++ {
		GameEnd.JingDingUser[i] = static.DING_NULL
		GameEnd.PeiScore[i] = 0
	}

	var huDetail components2.TagHuCostDetail
	nextBanker := sp.SuanFen(&GameEnd)

	//计算各玩家开口次数，明杠，暗杠，红中，赖子
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_userItem := sp.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}

		//保存四家明杠的次数
		GameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang + _userItem.Ctx.XuGang)
		//保存四家暗杠的次数
		GameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		huDetail.Private(_userItem.Seat, components2.TagShowGan, int(_userItem.Ctx.ShowGang), components2.DetailTypeCost) //明杠
		huDetail.Private(_userItem.Seat, components2.TagXuGan, int(_userItem.Ctx.XuGang), components2.DetailTypeCost)     //蓄杠
		huDetail.Private(_userItem.Seat, components2.TagAnGan, int(_userItem.Ctx.HidGang), components2.DetailTypeCost)    //暗杠

		////漂
		//if sp.Rule.DingPiao != Piao_k5x_NoPiao{
		//	huDetail.Private(_userItem.Seat, components2.TagPiao, _userItem.Ctx.VecXiaPao.Num, components2.DetailTypeCost)
		//}

		if GameEnd.WWinner[i] {
			//过胡翻番
			huDetail.Private(_userItem.Seat, components2.TagGuoHu, sp.GetUserGuoHuFan(i), components2.DetailTypeF)
			//按wKindMask列出的顺序，依次查找是哪个大胡，多个大胡时，按第一个查到的大胡显示
			for k := 0; k < len(k5x_KindMask); k++ {
				if (_userItem.Ctx.ChiHuResult.ChiHuKind & uint64(k5x_KindMask[k])) != 0 {
					switch k5x_KindMask[k] {
					case static.CHK_7_DUI:
						GameEnd.WBigHuKind[i] = static.GameBigHuKind_7
					case static.CHK_7_DUI_1:
						GameEnd.WBigHuKind[i] = static.GameBigHuKind_7Dui_1 //豪华七对
					case static.CHK_7_DUI_2:
						GameEnd.WBigHuKind[i] = static.GameBigHuKind_7Dui_2 //超豪华七对
					case static.CHK_7_DUI_3:
						GameEnd.WBigHuKind[i] = static.GameBigHuKind_7Dui_3 //超超豪华七对
					case static.CHK_QING_YI_SE: //清一色
						GameEnd.WBigHuKind[i] = static.GameBigHuKindQYS
					case static.CHK_XIAO_SAN_YUAN: //小三元
						GameEnd.WBigHuKind[i] = static.GameBigHuKindXSY
					case static.CHK_DA_SAN_YUAN: //大三元
						GameEnd.WBigHuKind[i] = static.GameBigHuKindDSY
					case static.CHK_KA_5_XING: //卡五星
						GameEnd.WBigHuKind[i] = static.GameBigHuKindK5X
					case static.CHK_SHOU_ZHUA_YI: //手抓一
						GameEnd.WBigHuKind[i] = static.GameBigHuKindSZ1
					case static.CHK_SI_GUI_YI_MING: //明四归一
						GameEnd.WBigHuKind[i] = static.GameBigHuKindM4G
						break
					case static.CHK_SI_GUI_YI_AN: //暗四归一
						GameEnd.WBigHuKind[i] = static.GameBigHuKindA4G
						break
					case static.CHK_PENG_PENG: //碰碰胡
						GameEnd.WBigHuKind[i] = static.GameBigHuKindPP
						break
					case static.CHK_QIANG_GANG: //抢杠胡
						GameEnd.WBigHuKind[i] = static.GameBigHuKindQG

					default:
						break
					}
					//大胡次数
					if GameEnd.WBigHuKind[i] != 0 {
						_userItem.Ctx.BigHuUserCount++
					}
					xlog.Logger().Debug(fmt.Sprintf("大胡类型：%d,玩家(%d),ChiHuKind:%v", GameEnd.WBigHuKind[i], _userItem.GetChairID(), _userItem.Ctx.ChiHuResult.ChiHuKind))
					break
				}
			}

			if _userItem.Ctx.ChiHuResult.ChiHuKind != 0 && GameEnd.WBigHuKind[i] == 0 {
				//平胡
				huDetail.Private(_userItem.Seat, components2.TagPingHu, 1, components2.DetailTypeFirst)
			}

			for k := 0; k < len(k5x_KindMask); k++ {
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_QIANG_GANG != 0 {
					//抢杠胡
					huDetail.Private(_userItem.Seat, components2.TagHuQiangGang, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI_1 != 0 {
					//豪华7对
					huDetail.Private(_userItem.Seat, components2.TagHu7Dui_1, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI_2 != 0 {
					//双豪华7对
					huDetail.Private(_userItem.Seat, components2.TagHu7Dui_2, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI_3 != 0 {
					//三豪华7对
					huDetail.Private(_userItem.Seat, components2.TagHu7Dui_3, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI != 0 {
					//7对
					huDetail.Private(_userItem.Seat, components2.TagHu7Dui, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_XIAO_SAN_YUAN != 0 {
					//小三元
					huDetail.Private(_userItem.Seat, components2.TagX3Y, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_DA_SAN_YUAN != 0 {
					//大三元
					huDetail.Private(_userItem.Seat, components2.TagD3Y, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_KA_5_XING != 0 {
					//卡5星
					huDetail.Private(_userItem.Seat, components2.TagK5x, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_SHOU_ZHUA_YI != 0 {
					//手抓一
					huDetail.Private(_userItem.Seat, components2.TagSZ1, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_SI_GUI_YI_MING != 0 {
					//明四归一
					huDetail.Private(_userItem.Seat, components2.TagM4G, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_SI_GUI_YI_AN != 0 {
					//暗四归一
					huDetail.Private(_userItem.Seat, components2.TagA4G, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_QING_YI_SE != 0 {
					//清一色
					huDetail.Private(_userItem.Seat, components2.TagHuQinYiSe, 1, components2.DetailTypeFirst)
					if _userItem.Ctx.ChiHuResult.ChiHuKind&CHK_M4G_1 != 0 {
						huDetail.Private(_userItem.Seat, components2.TagM4G, 1, components2.DetailTypeCost)
					} else if _userItem.Ctx.ChiHuResult.ChiHuKind&CHK_M4G_2 != 0 {
						huDetail.Private(_userItem.Seat, components2.TagM4G, 2, components2.DetailTypeCost)
					} else if _userItem.Ctx.ChiHuResult.ChiHuKind&CHK_M4G_3 != 0 {
						huDetail.Private(_userItem.Seat, components2.TagM4G, 3, components2.DetailTypeCost)
					}
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_PENG_PENG != 0 {
					//碰碰胡
					huDetail.Private(_userItem.Seat, components2.TagHuPengPengHu, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_HAI_DI != 0 {
					//海底捞
					huDetail.Private(_userItem.Seat, components2.TagHuHaiDi, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_GANG_SHANG_KAI_HUA != 0 {
					//杠上开花
					huDetail.Private(_userItem.Seat, components2.TagHuGangKai, 1, components2.DetailTypeFirst)
				}
				if _userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_GANG_SHANG_PAO != 0 {
					//杠上炮
					if provideItem := sp.GetUserItemByChair(sp.ProvideUser); provideItem != nil {
						huDetail.Private(sp.ProvideUser, components2.TagGangPao, 1, components2.DetailTypeFirst)
					}
				}
			}

			//数坎
			if sp.Rule.ShuKan {
				huDetail.Private(_userItem.Seat, components2.TagShuKan, sp.KanCount[i], components2.DetailTypeADD)
			}
		}
	}

	GameEnd.IsQuit = false
	GameEnd.TheOrder = sp.CurCompleteCount
	//如果荒庄了,赔分也要计入游戏分数中
	if GameEnd.HuangZhuang {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			GameEnd.GameScore[i] = GameEnd.PeiScore[i]
		}
	}

	//判断调整分
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		if uint16(i) == static.INVALID_CHAIR {
			GameEnd.MaxFSCount[i] = 0
		} else {
			GameEnd.MaxFSCount[i] = uint16(sp.FanScore[i].FanNum[i])
		}

		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.StrEnd[i] = huDetail.GetSeatString(uint16(i))

		// 统计最大分数
		if sp.CurCompleteCount == 1 {
			_item.Ctx.SetMaxFan(int(GameEnd.GameScore[i] + GameEnd.GameAdjustScore[i]))
		} else {
			if int(GameEnd.GameScore[i]) > _item.Ctx.MaxScoreUserCount {
				_item.Ctx.SetMaxFan(int(GameEnd.GameScore[i] + GameEnd.GameAdjustScore[i]))
			}
		}

		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
	}

	GameEnd.UserScore, GameEnd.UserVitamin = sp.OnSettle(GameEnd.GameScore, consts.EventSettleGameOver)

	//发送信息
	sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SaveGameData()
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	//荒庄
	if sp.HaveHuangZhuang {
		//记录荒庄
		sp.addReplayOrder(0, info2.E_HuangZhuang, 0)
		//记录胡牌类型
		sp.ReplayRecord.BigHuKind = 2
		sp.ReplayRecord.ProvideUser = 9
	} else {
		//记录胡牌类型
		//sp.ReplayRecord.BigHuKind = GameEnd.BigHuKind
		for i := 0; i < sp.GetPlayerCount(); i++ {
			sp.ReplayRecord.WBigHuKind[i] = GameEnd.WBigHuKind[i]
		}
		sp.ReplayRecord.ProvideUser = GameEnd.ProvideUser
	}

	// 数据库写出牌记录
	sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)

	// 写完后清除数据
	sp.ReplayRecord.Reset()

	//数据库写分
	for _, v := range sp.PlayerInfo {
		wintype := static.ScoreKind_Draw
		if GameEnd.GameScore[v.Seat] > 0 {
			wintype = static.ScoreKind_Win
		} else {
			wintype = static.ScoreKind_Lost
		}
		//这里写数据库战绩记录,需要把玩家的杠分记录进去
		sp.TableWriteGameDate(int(sp.CurCompleteCount), v, wintype, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
	}

	//扣房卡
	if sp.CurCompleteCount == 1 {
		sp.TableDeleteFangKa(sp.CurCompleteCount)
	}

	//结束游戏
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu { //局数够了
		sp.CalculateResultTotal(static.GER_NORMAL, wChairID, 0) //计算总发送总结算
		sp.bIsXuanPiaoSure = false
		sp.UpdateOtherFriendDate(&GameEnd, false)
		//通知框架结束游戏
		//sp.SetGameStatus(static.GS_MJ_FREE)
		sp.ConcludeGame()

	} else {
	}

	if sp.BankerUser != nextBanker && nextBanker != static.INVALID_CHAIR {
		sp.BankerUser = nextBanker
	}

	sp.OnGameEnd()
	sp.setDelayLimitedTime(4)
	sp.RepositTable() // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	if sp.Rule.Endready {
		sp.SetAutoNextTimer(15) //自动开始下一局
	}
	return true
}

// ! 强退，结束游戏
func (sp *SportYCKWX) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	//变量定义
	var GameEnd static.Msg_S_GameEnd
	GameEnd.EndStatus = cbReason
	GameEnd.MagicCard = sp.MagicCard
	//设置变量
	GameEnd.ProvideUser = static.INVALID_CHAIR
	GameEnd.Winner = static.INVALID_CHAIR
	GameEnd.Contractor = static.INVALID_CHAIR
	GameEnd.IsQuit = true
	for i := 0; i < int(sp.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}

	//抢杠分数，解散了也要结算
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		//玩家番数
		GameEnd.MaxFSCount[i] = 0 //uint16(sp.GetPlayerFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount
		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
	}

	GameEnd.UserScore, GameEnd.UserVitamin = sp.OnSettle(GameEnd.GameScore, consts.EventSettleGameOver)

	//税收计算,按正常情况下进行计算,如果有调整,在后面进行调整
	var ScoreInfo [meta2.MAX_PLAYER]rule2.TagScoreInfo
	for i := 0; i < sp.GetPlayerCount(); i++ {
		//记录各家的分数
		ScoreInfo[i].Score = GameEnd.GameScore[i]
		//如果分数为0，设置状态和局状态
		if GameEnd.GameScore[i] == 0 {
			ScoreInfo[i].ScoreKind = static.ScoreKind_Draw
		} else { //否则如果分数为正,设置为赢,为负设置为输,即使后面有人承包,也是输
			if GameEnd.GameScore[i] > 0 {
				ScoreInfo[i].ScoreKind = static.ScoreKind_Win
			} else {
				ScoreInfo[i].ScoreKind = static.ScoreKind_Lost
			}
		}
	}

	//游戏记录
	sp.OnWriteGameRecord(wChairID, "强退游戏结束")
	//记录异常结束数据
	sp.addReplayOrder(wChairID, info2.E_Li_Xian, 0)
	//记录胡牌类型
	sp.ReplayRecord.BigHuKind = GameEnd.BigHuKind
	sp.ReplayRecord.ProvideUser = GameEnd.ProvideUser
	//荒庄
	if sp.HaveHuangZhuang {
		//记录荒庄
		sp.addReplayOrder(0, info2.E_HuangZhuang, 0)
		sp.ReplayRecord.BigHuKind = 2
		sp.ReplayRecord.ProvideUser = 9
	}

	nextCardLen := len(GameEnd.NextCard)
	for i := 0; i < nextCardLen; i++ {
		GameEnd.NextCard[i] = 0x00
	}
	//发送几张牌出去
	if int(sp.LeftCardCount) >= nextCardLen {
		for i := 0; i < nextCardLen; i++ {
			sp.LeftCardCount--
			GameEnd.NextCard[i] = sp.RepertoryCard[sp.LeftCardCount]
		}
	}

	//发送信息
	sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SaveGameData()

	if sp.GetGameStatus() != static.GS_MJ_FREE {
		// 数据库写出牌记录
		sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
		// 写完后清除数据
		sp.ReplayRecord.Reset()

		//数据库写分
		for _, v := range sp.PlayerInfo {
			if v.Seat != static.INVALID_CHAIR {
				if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
				} else {
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
				}
			}
		}
	}

	sp.UpdateOtherFriendDate(&GameEnd, true)
	//结束游戏
	sp.CalculateResultTotal(static.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	//sp.SetGameStatus(static.GS_MJ_FREE)
	sp.ConcludeGame()

	return true
}

// ! 解散，结束游戏
func (sp *SportYCKWX) OnGameEndDissmiss(wChairID uint16, cbReason byte, cbSubReason byte) bool {
	//变量定义
	var GameEnd static.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sp.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.ProvideUser = static.INVALID_CHAIR
	GameEnd.Winner = static.INVALID_CHAIR
	GameEnd.Contractor = static.INVALID_CHAIR
	for i := 0; i < int(sp.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}

	GameEnd.MagicCard = sp.MagicCard

	//记录异常结束数据
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
			sp.addReplayOrder(uint16(i), info2.E_Li_Xian, 0)
		}
	}

	sp.addReplayOrder(wChairID, info2.E_Jie_san, 0)

	//抢杠分数，解散了也要结算
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		//玩家番数
		GameEnd.MaxFSCount[i] = 0
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount
		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount

		if sp.Rule.HasPao {
			GameEnd.PaoCount[i] = uint16(_item.Ctx.VecXiaPao.Num)
		} else {
			GameEnd.PaoCount[i] = 0xFF
		}

		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
	}

	GameEnd.UserScore, GameEnd.UserVitamin = sp.OnSettle(GameEnd.GameScore, consts.EventSettleGameOver)

	//税收计算,按正常情况下进行计算,如果有调整,在后面进行调整
	var ScoreInfo [meta2.MAX_PLAYER]rule2.TagScoreInfo
	for i := 0; i < sp.GetPlayerCount(); i++ {
		//记录各家的分数
		ScoreInfo[i].Score = GameEnd.GameScore[i]

		//如果分数为0，设置状态和局状态
		if GameEnd.GameScore[i] == 0 {
			ScoreInfo[i].ScoreKind = static.ScoreKind_Draw
		} else { //否则如果分数为正,设置为赢,为负设置为输,即使后面有人承包,也是输
			if GameEnd.GameScore[i] > 0 {
				ScoreInfo[i].ScoreKind = static.ScoreKind_Win
			} else {
				ScoreInfo[i].ScoreKind = static.ScoreKind_Lost
			}
		}
	}

	GameEnd.IsQuit = true
	nextCardLen := len(GameEnd.NextCard)
	for i := 0; i < nextCardLen; i++ {
		GameEnd.NextCard[i] = 0x00
	}
	//发送几张牌出去
	if int(sp.LeftCardCount) >= nextCardLen {
		for i := 0; i < nextCardLen; i++ {
			sp.LeftCardCount--
			GameEnd.NextCard[i] = sp.RepertoryCard[sp.LeftCardCount]
		}
	}

	//发送信息
	sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SaveGameData()
	switch cbSubReason {
	case 0:
		//游戏记录
		sp.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

		if sp.GetGameStatus() != static.GS_MJ_FREE && !sp.PayPaoStatus {
			//数据库写出牌记录
			sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
			// 写完后清除数据
			sp.ReplayRecord.Reset()

			//数据库写入单局结算
			for _, v := range sp.PlayerInfo {
				if v.Seat != static.INVALID_CHAIR {
					if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
						sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
					} else {
						sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
					}
				}
			}
		}
	case 1:
		sp.OnWriteGameRecord(wChairID, "前面某个时刻程序出错过，需要排查错误日志，无法恢复这局游戏，解散游戏OnGameEndErrorDissmis")
	}

	sp.UpdateOtherFriendDate(&GameEnd, true)
	// 写总计算
	sp.CalculateResultTotal(static.GER_DISMISS, wChairID, cbSubReason)
	//结束游戏
	//sp.SetGameStatus(static.GS_MJ_FREE)
	sp.ConcludeGame()

	return true
}

// ! 解散牌桌
func (sp *SportYCKWX) OnEnd() {
	if sp.IsGameStarted() {
		sp.OnGameOver(static.INVALID_CHAIR, static.GER_DISMISS)
	}
}

// ! 计算总发送总结算
func (sp *SportYCKWX) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	sp.TimeEnd = time.Now().Unix() //大局结束时间
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = sp.Rule.FangZhuID
	balanceGame.CurTotalCount = sp.CurCompleteCount //总盘数
	balanceGame.TimeStart = sp.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = sp.TimeEnd
	for i := 0; i < len(sp.VecGameEnd); i++ {
		for j := 0; j < sp.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += (sp.VecGameEnd[i].GameScore[j] + sp.VecGameEnd[i].GameAdjustScore[j]) //总分
		}
	}

	//打印日志
	recrodStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）", cbReason)
	sp.OnWriteGameRecord(wChairID, recrodStr)

	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.OnWriteGameRecord(uint16(i), "当前该玩家离线")
	}

	//非正常结束，记录总结算时玩家状态：0正常，1解散，2离线
	if static.GER_USER_LEFT == cbReason {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if i == int(wChairID) {
				balanceGame.UserEndState[i] = 2
			}
		}
	} else {
		if static.GER_DISMISS == cbReason {
			for i := 0; i < sp.GetPlayerCount(); i++ {
				if i == int(wChairID) {
					balanceGame.UserEndState[i] = 1
				} else {
					if _item := sp.GetUserItemByChair(uint16(i)); _item != nil && _item.UserStatus == static.US_OFFLINE {
						balanceGame.UserEndState[i] = 2
					}
				}
			}
		}
	}

	//! 计算大赢家的座位号
	bigWinScore := 0
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if balanceGame.GameScore[i] >= bigWinScore {
			bigWinScore = balanceGame.GameScore[i]
		}
	}

	//开始就解散时不加，每局结束后游戏状态会复位为GS_MJ_FREE
	wintype := static.ScoreKind_Draw
	if sp.CurCompleteCount == 1 && sp.GetGameStatus() != static.GS_MJ_END {
		wintype = static.ScoreKind_pass
	} else {
		if sp.CurCompleteCount == 0 {
			wintype = static.ScoreKind_pass
		}
	}
	if cbSubReason == 0 {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_userItem := sp.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			balanceGame.FXMaxUserCount[i] = _userItem.Ctx.MaxFanUserCount
			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
			balanceGame.ZimoCount[i] = _userItem.Ctx.HuBySelfCount

			if wintype != static.ScoreKind_pass {
				if balanceGame.GameScore[i] > 0 {
					wintype = static.ScoreKind_Win
				} else {
					wintype = static.ScoreKind_Lost
				}
			}

			isBigWin := 0
			if bigWinScore == balanceGame.GameScore[i] {
				isBigWin = 1
			}
			//写记录
			sp.TableWriteGameDateTotal(int(sp.CurCompleteCount), i, wintype, balanceGame.GameScore[i], isBigWin)
		}
	} else {
		sp.UpdateErrGameTotal(sp.GetTableInfo().GameNum)
	}

	// 记录用户好友房历史战绩
	if wintype != static.ScoreKind_pass {
		sp.TableWriteHistoryRecord(&balanceGame)
		sp.TableWriteHistoryRecordDetail(&balanceGame)
	}

	balanceGame.End = 0

	for i := 0; i < sp.GetPlayerCount(); i++ {
		_userItem := sp.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}

		gameendStr := ""
		if len(sp.VecGameEnd) > 0 {
			gameendStr = static.HF_JtoA(sp.VecGameEnd[len(sp.VecGameEnd)-1])
		}
		gamedataStr := ""
		if len(sp.VecGameDataAllP32[i]) > 0 {
			gamedataStr = static.HF_JtoA(sp.VecGameDataAllP32[i][len(sp.VecGameDataAllP32[i])-1])
		}
		sp.SaveLastGameinfo(_userItem.Uid, gameendStr, static.HF_JtoA(balanceGame), gamedataStr)
	}

	//发消息
	sp.SendTableMsg(consts.MsgTypeGameBalanceGame, balanceGame)

	sp.resetEndDate()
}

// ! 重置优秀结束数据
func (sp *SportYCKWX) resetEndDate() {
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_GameEnd{}

	for _, v := range sp.PlayerInfo {
		v.OnEnd()
	}
}

func (sp *SportYCKWX) UpdateOtherFriendDate(GameEnd *static.Msg_S_GameEnd, bEnd bool) {

}

// ! 给客户端发送总结算
func (sp *SportYCKWX) CalculateResultTotal_Rep(msg *static.Msg_C_BalanceGameEeq) {
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = sp.Rule.FangZhuID
	balanceGame.CurTotalCount = sp.CurCompleteCount //总盘数

	/*为什么要特意判断下0？之前没有这个协议，是新需求要加的之前很多数据都在RepositTableFrameSink里面重置了，但是由于新功能的开发
	部分数据放到游戏开始的时候重置了，但是存在的问题是，如果一个桌子解散了，这个桌子在服务器内部还是存在的，他的部分数据也是存在的
	当某个玩家进来，还没有点开始的时候点战绩，这时候存在把本茶楼桌子上次数据发送过来的情况，但是改动m_MaxFanUserCount之类变量的值又
	有很大风险，因此此处做个特出处理，如果是第0局，没有开始，那就无条件全部返回0*/
	if 0 == balanceGame.CurTotalCount {
		for i := 0; i < len(sp.VecGameEnd); i++ {
			for j := 0; j < sp.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += 0 //总分
			}
		}

		for i := 0; i < sp.GetPlayerCount(); i++ {
			balanceGame.ChiHuUserCount[i] = 0
			balanceGame.ProvideUserCount[i] = 0
			balanceGame.FXMaxUserCount[i] = 0
			balanceGame.HHuUserCount[i] = 0
			balanceGame.UserEndState[i] = 0
		}
	} else {
		for i := 0; i < len(sp.VecGameEnd); i++ {
			for j := 0; j < sp.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += sp.VecGameEnd[i].GameScore[j] //总分
			}
		}

		for i := 0; i < sp.GetPlayerCount(); i++ {
			balanceGame.UserEndState[i] = 0
		}

		// 有打赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < sp.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				iMaxScoreCount++
			}
		}

		for i := 0; i < sp.GetPlayerCount(); i++ {
			_userItem := sp.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			balanceGame.FXMaxUserCount[i] = _userItem.Ctx.MaxFanUserCount
			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
		}
	}
	balanceGame.End = 1
	//发消息
	sp.SendPersonMsg(consts.MsgTypeGameBalanceGame, balanceGame, sp.GetChairByUid(msg.Id))
}

// ! 发送游戏开始场景数据
func (sp *SportYCKWX) sendGameSceneStatusPlay(player *components2.Player) bool {

	wChiarID := player.GetChairID()

	if wChiarID == static.INVALID_CHAIR {
		xlog.Logger().Debug("sendGameSceneStatusPlay invalid chair")
		return false
	}

	//变量定义
	var StatusPlay CMD_S_StatusPlay32
	//游戏变量
	StatusPlay.SiceCount = sp.SiceCount
	StatusPlay.BankerUser = sp.BankerUser
	StatusPlay.CurrentUser = sp.CurrentUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sp.PiZiCard
	StatusPlay.KaiKou = sp.Rule.KouKou
	StatusPlay.RenWuAble = sp.RenWuAble
	StatusPlay.TheOrder = sp.CurCompleteCount
	if sp.CurrentUser == player.Seat {
		StatusPlay.SendCardData = sp.SendCardData
	} else {
		StatusPlay.SendCardData = static.INVALID_BYTE
	}

	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//玩家番数
		StatusPlay.PlayerFan[i] = 0 //sp.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}

		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount

		//亮牌相关
		if sp.m_ShowCard[i].BIsShowCard {
			StatusPlay.LiangStructArray[i].LiangCard = sp.m_ShowCard[i].CbLiangCard
			StatusPlay.LiangStructArray[i].AnPuCard = sp.m_ShowCard[i].CbAnPuCard
			StatusPlay.LiangStructArray[i].TingCard = sp.m_ShowCard[i].CbTingCard
		}

		//过胡次数
		StatusPlay.GuoHuCount[i] = sp.GuoHuCount[i]
	}

	//状态变量
	StatusPlay.PayPaostatus = sp.PayPaoStatus
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = player.Ctx.UserAction32

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	StatusPlay.VecGangCard = static.HF_BytesToInts(player.Ctx.VecGangCard)

	if player.Ctx.Response {
		StatusPlay.ActionMask = static.WIK_NULL
	}

	//if player.Ctx.CheckTimeOut != 0 {
	//	if sp.m_ShowCard[player.Seat].BIsShowCard{
	//		sp.LockTimeOut(player.Seat, static.GAME_LIANGPAI_AUTOOUT_TIME)
	//	}//else{
	//		//sp.LockTimeOut(player.Seat, static.GAME_OPERATION_TIME_12)
	//	//}
	//	StatusPlay.Overtime = player.Ctx.CheckTimeOut
	//} else {
	CurUserItem := sp.GetUserItemByChair(sp.CurrentUser)
	if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
	}

	for _, v := range sp.PlayerInfo {
		if v.GetChairID() == player.GetChairID() {
			continue
		}
		if v.GetChairID() == sp.CurrentUser {
			continue
		}
		if v.Ctx.UserAction32 > 0 {
			StatusPlay.Overtime = 0
			break
		}
	}
	//}

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData
	StatusPlay.LastOutCardUser = sp.LastOutCardUser

	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardData = sp.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)

	//发送场景
	sp.SendPersonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, wChiarID)
	//发小结消息
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 && int(wChiarID) < sp.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		sp.SendPersonMsg(consts.MsgTypeGameEnd, gamend, wChiarID)
	}

	//发送解散房间所有玩家的反应
	sp.SendAllPlayerDissmissInfo(player)

	//断线重连如果在选漂状态需要通知玩家选漂以及其他玩家选漂状态
	if sp.PayPaoStatus {
		sp.SendPaoSetting(true)

	}

	return true
}

func (sp *SportYCKWX) SaveGameData() {
	//变量定义
	var StatusPlay CMD_S_StatusPlay32
	//游戏变量
	StatusPlay.SiceCount = sp.SiceCount
	StatusPlay.BankerUser = sp.BankerUser
	StatusPlay.CurrentUser = sp.CurrentUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sp.PiZiCard
	StatusPlay.KaiKou = sp.Rule.KouKou
	StatusPlay.RenWuAble = sp.RenWuAble
	StatusPlay.TheOrder = sp.CurCompleteCount

	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//玩家番数
		StatusPlay.PlayerFan[i] = 0 //sp.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}

		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount

		//亮牌相关
		if sp.m_ShowCard[i].BIsShowCard {
			StatusPlay.LiangStructArray[i].LiangCard = sp.m_ShowCard[i].CbLiangCard
			StatusPlay.LiangStructArray[i].AnPuCard = sp.m_ShowCard[i].CbAnPuCard
			StatusPlay.LiangStructArray[i].TingCard = sp.m_ShowCard[i].CbTingCard
		}

		//过胡次数
		StatusPlay.GuoHuCount[i] = sp.GuoHuCount[i]
	}

	//状态变量
	StatusPlay.PayPaostatus = sp.PayPaoStatus
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData
	StatusPlay.LastOutCardUser = sp.LastOutCardUser

	//玩家的个人数据
	for i := 0; i < sp.GetPlayerCount(); i++ {
		player := sp.GetUserItemByChair(uint16(i))
		if player == nil {
			continue
		}
		if sp.CurrentUser == player.Seat {
			StatusPlay.SendCardData = sp.SendCardData
		} else {
			StatusPlay.SendCardData = static.INVALID_BYTE
		}
		StatusPlay.ActionMask = player.Ctx.UserAction32
		StatusPlay.VecGangCard = static.HF_BytesToInts(player.Ctx.VecGangCard)
		if player.Ctx.Response {
			StatusPlay.ActionMask = static.WIK_NULL
		}

		if player.Ctx.CheckTimeOut != 0 {
			if sp.m_ShowCard[player.Seat].BIsShowCard {
				sp.LockTimeOut(player.Seat, static.GAME_LIANGPAI_AUTOOUT_TIME)
			} //else{
			//sp.LockTimeOut(player.Seat, static.GAME_OPERATION_TIME_12)
			//}
			StatusPlay.Overtime = player.Ctx.CheckTimeOut
		} else {
			CurUserItem := sp.GetUserItemByChair(sp.CurrentUser)
			if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
				StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
			}

			for _, v := range sp.PlayerInfo {
				if v.GetChairID() == player.GetChairID() {
					continue
				}
				if v.GetChairID() == sp.CurrentUser {
					continue
				}
				if v.Ctx.UserAction32 > 0 {
					StatusPlay.Overtime = 0
					break
				}
			}
		}
		//扑克数据
		StatusPlay.CardCount, StatusPlay.CardData = sp.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
		sp.VecGameDataAllP32[i] = append(sp.VecGameDataAllP32[i], StatusPlay) //保存，用于汇总计算
	}
}

// 游戏场景消息发送
func (sp *SportYCKWX) SendGameScene(uid int64, status byte, secret bool) {
	player := sp.GetUserItemByUid(uid)
	if player == nil {
		return
	}
	switch status {
	case static.GS_MJ_FREE:
		sp.SendGameSceneStatusFree(player)
	case static.GS_MJ_PLAY:
		sp.sendGameSceneStatusPlay(player)
	case static.GS_MJ_END:
		sp.sendGameSceneStatusPlay(player)
	}
}

// ! 剩余多少张牌进入海底捞
func (sp *SportYCKWX) GetHaidiCount() byte {
	return 0
}

// ! 游戏退出
func (sp *SportYCKWX) OnExit(uid int64) {
	sp.Common.OnExit(uid)
}

// ! 定时器
func (sp *SportYCKWX) OnTime() {
	sp.Common.OnTime()
}

// ! 计时器事件
func (sp *SportYCKWX) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	if sp.IsPausing() {
		return true
	}

	if dwTimerID == components2.GameTime_15 {
		if sp.GetGameStatus() == static.GS_MJ_END {
			//dqjs-7318 防止亮牌后一炮多响时有人异常托管
			return true
		}
		if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
			//到时间了,并且玩家没有离线
			if sp.PayPaoStatus && sp.Rule.Overtime_trust > 0 {
				sp.OnTrustAutoOperate(TablePerson.Seat, true)
			} else {
				//1.检查进入托管
				if !TablePerson.CheckTRUST() {
					if sp.m_ShowCard[TablePerson.Seat].BIsShowCard {
						//亮倒过胡翻番
						if sp.Rule.GHFF && TablePerson.Ctx.UserAction32 != 0 && TablePerson.Ctx.UserAction32&static.WIK_CHI_HU != 0 {
							var msg = &static.Msg_S_DG_Trustee{ChairID: TablePerson.Seat, Trustee: true}
							if sp.OnUserTustee(msg) {
								sp.OnWriteGameRecord(TablePerson.Seat, "超时进入托管")
							}
						}
					} else {
						var msg = &static.Msg_S_DG_Trustee{ChairID: TablePerson.Seat, Trustee: true}
						if sp.OnUserTustee(msg) {
							sp.OnWriteGameRecord(TablePerson.Seat, "超时进入托管")
						}
					}
				}
				//2.优先亮牌操作，再托管操作
				if sp.m_ShowCard[TablePerson.Seat].BIsShowCard { //&& TablePerson.UserOfflineTag == -1{
					sp.OnLiangAutoOperate(TablePerson.Seat, true)
				} else {
					if TablePerson.CheckTRUST() {
						sp.OnTrustAutoOperate(TablePerson.Seat, true)
					}
				}
			}
		}
	}
	if dwTimerID == components2.GameTime_Delayer {
		if sp.GetGameStatus() == static.GS_MJ_END {
			sp.accountManage(sp.Rule.Overtime_dismiss, sp.Rule.Overtime_trust, 15)
		}
	}

	return true
}

// ! 玩家开启超时
func (sp *SportYCKWX) LockTimeOut(cUser uint16, iTime int) {
	if cUser < 0 || cUser > uint16(sp.GetPlayerCount()) {
		return
	}

	_userItem := sp.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}
	if iTime < 1 {
		iTime = static.GAME_OPERATION_TIME_12
	}

	//托管状态
	if _userItem.CheckTRUST() {
		iTime = 1
	}

	sp.LimitTime = time.Now().Unix() + int64(iTime)
	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + int64(iTime)

	if sp.Rule.NineSecondRoom {
		_userItem.Ctx.Timer.SetTimer(components2.GameTime_15, iTime)
	}
}

// ! 玩家关闭超时
func (sp *SportYCKWX) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(sp.GetPlayerCount()) {
		return
	}

	_userItem := sp.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0

	if sp.Rule.NineSecondRoom {
		_userItem.Ctx.Timer.KillTimer(components2.GameTime_15)
	}
}

// 清除所有人的定时器
func (sp *SportYCKWX) UnLockAllPlayerTimer() {
	for ip := uint16(0); int(ip) < sp.GetPlayerCount(); ip++ {
		v := sp.GetPlayerByChair(ip)
		// 解锁超时时间
		if v != nil && v.Ctx.UserAction32 != static.WIK_NULL {
			sp.UnLockTimeOut(ip)
		}
	}
}

// ! 写日志记录
func (sp *SportYCKWX) WriteGameRecord() {
	//写日志记录
	sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("开始卡五星  发牌......第%d局", sp.CurCompleteCount))

	// 玩家手牌
	for i := 0; i < len(sp.PlayerInfo); i++ {
		v := sp.GetUserItemByChair(uint16(i))
		if v != nil {
			handCardStr := fmt.Sprintf("发牌后手牌:%s", sp.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
			sp.OnWriteGameRecord(uint16(v.Seat), handCardStr)
		}
	}

	// 牌堆牌
	leftCardStr := fmt.Sprintf("牌堆牌:%s", sp.m_GameLogic.SwitchToCardNameByDatas(sp.RepertoryCard[0:sp.LeftCardCount+1], 0))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, leftCardStr)

	//赖子牌
	magicCardStr := fmt.Sprintf("癞子牌:%s", sp.m_GameLogic.SwitchToCardNameByData(sp.MagicCard, 1))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, magicCardStr)
}

// ! 场景保存
func (sp *SportYCKWX) Tojson() string {
	var _json components2.GameJsonSerializer
	//return _json.ToJson_jmsk(sp)
	_json.ToJson(&sp.Metadata)
	_json.GameCommonToJson(&sp.Common)
	_json.CurLianGangCount = sp.LianGangCount
	_json.UserGuoHuCount = sp.GuoHuCount     //玩家过胡次数
	_json.UserKanCount = sp.KanCount         //玩家数坎个数
	_json.HasSendNo13Tip = sp.HasSendNo13Tip //不够13张不能亮牌
	_json.FirstXuanPiaoSure = sp.bIsXuanPiaoSure
	_json.FirstPiaoNum = sp.iPlayerPiaoNum
	_json.GameUserTing = sp.UserTing
	_json.GameUserTingType = sp.UserTingType
	_json.QiangGangScoreSend = sp.bQiangGangScoreSend
	_json.QiangGangOperateScore = sp.QiangGangOperateScore
	_json.GameShowCard = sp.m_ShowCard
	_json.GameReplay_Record = sp.ReplayRecord
	return static.HF_JtoA(&_json)
}

// ! 场景恢复
func (sp *SportYCKWX) Unmarsha(data string) {

	//_json.Unmarsha_jmsk(data, sp)
	if data != "" {
		var _json components2.GameJsonSerializer

		json.Unmarshal([]byte(data), &_json)

		_json.Unmarsha(&sp.Metadata)
		_json.JsonToStruct(&sp.Common)

		sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
		sp.m_GameLogic.Rule = sp.Rule
		sp.m_GameLogic.HuType = sp.HuType
		sp.m_GameLogic.SetMagicCard(sp.MagicCard)
		sp.m_GameLogic.SetPiZiCard(sp.PiZiCard)
		sp.LianGangCount = _json.CurLianGangCount
		sp.GuoHuCount = _json.UserGuoHuCount
		sp.KanCount = _json.UserKanCount
		sp.HasSendNo13Tip = _json.HasSendNo13Tip
		sp.bIsXuanPiaoSure = _json.FirstXuanPiaoSure
		sp.iPlayerPiaoNum = _json.FirstPiaoNum
		sp.UserTing = _json.GameUserTing
		sp.UserTingType = _json.GameUserTingType
		sp.bQiangGangScoreSend = _json.QiangGangScoreSend
		sp.QiangGangOperateScore = _json.QiangGangOperateScore
		sp.m_ShowCard = _json.GameShowCard
		sp.ReplayRecord = _json.GameReplay_Record
	}
}

func (sp *SportYCKWX) flashClient(userid int64, msg string) {
	wChairSeat := sp.GetChairByUid(userid)
	sp.OnWriteGameRecord(wChairSeat, msg)
	sp.SendGameNotificationMessage(wChairSeat, msg)
	sp.DisconnectOnMisoperation(wChairSeat)
}

func (sp *SportYCKWX) ChoosePiao() {
	switch sp.Rule.DingPiao {
	case Piao_k5x_ZiYouPiao: //自由漂
		{
			sp.SendPaoSetting(false)
		}
	case Piao_k5x_DingPiaoFirst: //首局定漂
		{
			if sp.bIsXuanPiaoSure {
				for _, v := range sp.PlayerInfo {
					var _msg static.Msg_C_Xiapao
					_msg.Num = sp.iPlayerPiaoNum[v.Seat]
					_msg.Id = sp.GetUidByChair(v.Seat)
					_msg.Status = true
					sp.OnUserClientXiaPao(&_msg)
				}
			} else {
				sp.SendPaoSetting(false)
			}
		}
	}
}

// 杠分
func (sp *SportYCKWX) GetScoreOnGang(gangType uint16) int {
	var Score int = 0
	switch gangType {
	case info2.E_Gang_XuGand:
		Score = 1
	case info2.E_Gang_AnGang:
		Score = 2
	case info2.E_Gang:
		Score = 2
	default:
		xlog.Logger().Debug("杠牌类型找不到")
	}
	return Score
}

// 游戏过程中，玩家的分数发生变化事件
func (sp *SportYCKWX) OnUserScoreOffset(seat uint16, offset int) bool {
	_userItem := sp.GetUserItemByChair(uint16(seat))
	if _userItem != nil {
		_userItem.Ctx.StorageScore += offset
	}
	return true
}

// 获取连杠番数
func (sp *SportYCKWX) GetLianGangFan() int {
	if sp.LianGangCount > 1 {
		num := 1
		for i := 1; i < sp.LianGangCount; i++ {
			num *= 2
		}
		return num
	}

	return 1
}

// 买马,返回马牌和马分
func (sp *SportYCKWX) Maima(wChairID uint16) (byte, byte) {
	if sp.LeftCardCount <= 0 || sp.Rule.MaiMaType == MaiMa_k5x_No {
		return 0, 0
	}

	sp.LeftCardCount--
	maCard := sp.RepertoryCard[sp.LeftCardCount]
	sp.addReplayOrder(wChairID, info2.E_Bird, int(maCard))
	xlog.Logger().Debug(fmt.Sprintf("玩家买马:%x,剩牌:%d", maCard, sp.LeftCardCount))
	//记录买马牌
	maCardStr := fmt.Sprintf("买马 马牌:%s", sp.m_GameLogic.SwitchToCardNameByData(maCard, 1))
	sp.OnWriteGameRecord(wChairID, maCardStr)

	cbValue := byte(maCard & static.MASK_VALUE)
	cbColor := byte(maCard & static.MASK_COLOR)

	if cbColor == 0x30 {
		return maCard, 10
	} else {
		return maCard, cbValue
	}
}

// 检查亮牌
func (sp *SportYCKWX) CheckLiangPai(wChairID uint16, checkk5 bool) int {

	//如果玩家已经亮牌,不需要再次检测
	if sp.m_ShowCard[wChairID].BIsShowCard {
		return static.WIK_NULL
	}

	//小于13张不能亮牌
	if sp.LeftCardCount < 12 {
		if !sp.HasSendNo13Tip {
			sp.SendTableMsg(consts.MsgTypeGameLiangNo13, "")
		}
		sp.HasSendNo13Tip = true
		return static.WIK_NULL
	}

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem != nil {
		AllCount := 0
		bCanLiang, cbOutCard := sp.m_GameLogic.AnalyseLiangCard(_userItem.Ctx.CardIndex, _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, 0, sp.SendCardData, checkk5)
		for _, card := range cbOutCard {
			if sp.IsOthersTingHuCard(wChairID, card) {
				AllCount++
			}
		}
		if bCanLiang && AllCount < len(cbOutCard) {
			////记录亮牌
			//maCardStr := fmt.Sprintf("玩家可打出:%s 亮牌,发送亮牌牌权", sp.m_GameLogic.SwitchToCardNameByDatas(cbOutCard, 1))
			//sp.OnWriteGameRecord(wChairID, maCardStr)
			return static.WIK_LIANGPAI
		}
	}

	return static.WIK_NULL
}

// 检查赔庄听牌
func (sp *SportYCKWX) CheckTing(wChairID uint16) (bool, uint64) {

	bIsTing, cbMaxHuKind := false, uint64(static.CHK_NULL)
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem != nil {
		cbMaxHuKind = sp.m_GameLogic.AnalyseMAXTingType(_userItem.Ctx.CardIndex[:], _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, 0, true)
		if cbMaxHuKind != static.CHK_NULL {
			bIsTing = true
		}
	}

	return bIsTing, cbMaxHuKind
}

func (sp *SportYCKWX) IsDaHu(ChiHuKind uint64) bool {
	if ChiHuKind&static.CHK_PENG_PENG != 0 ||
		ChiHuKind&static.CHK_KA_5_XING != 0 ||
		ChiHuKind&static.CHK_SI_GUI_YI_MING != 0 ||
		ChiHuKind&static.CHK_SI_GUI_YI_AN != 0 ||
		ChiHuKind&static.CHK_SHOU_ZHUA_YI != 0 ||
		ChiHuKind&static.CHK_7_DUI != 0 ||
		ChiHuKind&static.CHK_7_DUI_1 != 0 ||
		ChiHuKind&static.CHK_7_DUI_2 != 0 ||
		ChiHuKind&static.CHK_7_DUI_3 != 0 ||
		ChiHuKind&static.CHK_QING_YI_SE != 0 ||
		ChiHuKind&static.CHK_XIAO_SAN_YUAN != 0 ||
		ChiHuKind&static.CHK_DA_SAN_YUAN != 0 {
		return true
	}
	return false
}

// ! 玩家胡牌分数
func (sp *SportYCKWX) GetPlayerHuScore(ChiHuKind uint64) int {
	huScore := 1
	//碰碰胡
	if (ChiHuKind & static.CHK_PENG_PENG) != 0 {
		if sp.Rule.PPx4 {
			//碰碰胡x4
			huScore *= 4
			fmt.Println("碰碰胡4分")
		} else {
			//碰碰胡x2
			huScore *= 2
			fmt.Println("碰碰胡2分")
		}
	}
	//卡五星
	if (ChiHuKind & static.CHK_KA_5_XING) != 0 {
		if sp.Rule.K5x4 {
			//卡五星x4
			huScore *= 4
			fmt.Println("卡五星4分")
		} else {
			//卡五星x2
			huScore *= 2
			fmt.Println("卡五星2分")
		}
	}
	//明四归一x2
	if (ChiHuKind & static.CHK_SI_GUI_YI_MING) != 0 {
		huScore *= 2
		fmt.Println("明四归一2分")
	}
	//暗四归一x4
	if (ChiHuKind & static.CHK_SI_GUI_YI_AN) != 0 {
		huScore *= 4
		fmt.Println("暗四归一4分")
	}
	//手抓一x4
	if (ChiHuKind & static.CHK_SHOU_ZHUA_YI) != 0 {
		huScore *= 4
		fmt.Println("手抓一4分")
	}
	//7对x4
	if (ChiHuKind&static.CHK_7_DUI) != 0 ||
		(ChiHuKind&static.CHK_7_DUI_1) != 0 ||
		(ChiHuKind&static.CHK_7_DUI_2) != 0 ||
		(ChiHuKind&static.CHK_7_DUI_3) != 0 {
		huScore *= 4
		fmt.Println("七对4分")
	}
	//清一色x4
	if (ChiHuKind & static.CHK_QING_YI_SE) != 0 {
		huScore *= 4
		fmt.Println("清一色4分")
	}
	//小三元x4
	if (ChiHuKind & static.CHK_XIAO_SAN_YUAN) != 0 {
		huScore *= 4
		fmt.Println("小三元4分")
	}
	//大三元x8
	if (ChiHuKind & static.CHK_DA_SAN_YUAN) != 0 {
		huScore *= 8
		fmt.Println("大三元8分")
	}

	fmt.Println(fmt.Sprintf("牌型分：%d", huScore))
	recordStr := fmt.Sprintf("胡牌型是%s, 牌型分:%d", sp.GetTingHuKind(ChiHuKind), huScore)
	sp.OnWriteGameRecord(static.INVALID_CHAIR, recordStr)
	return huScore
}

// ! 玩家胡牌大胡牌型个数
func (sp *SportYCKWX) GetPlayerDaHuCount(ChiHuKind uint64) int {
	DaHuCount := 0
	//碰碰胡
	if (ChiHuKind & static.CHK_PENG_PENG) != 0 {
		DaHuCount++
	}
	//卡五星
	if (ChiHuKind & static.CHK_KA_5_XING) != 0 {
		DaHuCount++
	}
	//明四归一
	if (ChiHuKind & static.CHK_SI_GUI_YI_MING) != 0 {
		DaHuCount++
	}
	//暗四归一
	if (ChiHuKind & static.CHK_SI_GUI_YI_AN) != 0 {
		DaHuCount++
	}
	//手抓一
	if (ChiHuKind & static.CHK_SHOU_ZHUA_YI) != 0 {
		DaHuCount++
	}
	//7对
	if (ChiHuKind & static.CHK_7_DUI) != 0 {
		DaHuCount++
	}
	//清一色
	if (ChiHuKind & static.CHK_QING_YI_SE) != 0 {
		DaHuCount++
	}
	//小三元
	if (ChiHuKind & static.CHK_XIAO_SAN_YUAN) != 0 {
		DaHuCount++
	}
	//大三元
	if (ChiHuKind & static.CHK_DA_SAN_YUAN) != 0 {
		DaHuCount++
	}

	return DaHuCount
}

// ! 得到赢家玩家番数
func (sp *SportYCKWX) GetPlayerFan(WinnerID uint16, wProvideUserID, wCurrentUser uint16, bIsBaoPei bool) (int, int) {
	if WinnerID == static.INVALID_CHAIR {
		return 0, 0
	}

	WinnerItem := sp.GetUserItemByChair(WinnerID)
	if WinnerItem == nil {
		return 0, 0
	}

	winerFanNum := 0
	LoserFanNum := 0
	//杠上开花+1
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_GANG_SHANG_KAI_HUA != 0 {
		if sp.Rule.GangPaox4 {
			//杠上开花 x4
			winerFanNum += 2
			fmt.Println(fmt.Sprintf("赢家%d杠开 +2番", WinnerID))
		} else {
			winerFanNum += 1
			fmt.Println(fmt.Sprintf("赢家%d杠开 +1番", WinnerID))
		}
	}
	//豪华七对+1
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI_1 != 0 {
		winerFanNum += 1
		fmt.Println(fmt.Sprintf("赢家%d豪华7对 +1番", WinnerID))
	}
	//双豪华七对+2
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI_2 != 0 {
		winerFanNum += 2
		fmt.Println(fmt.Sprintf("赢家%d双豪华7对 +2番", WinnerID))
	}
	//三豪华七对+3
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_7_DUI_3 != 0 {
		winerFanNum += 3
		fmt.Println(fmt.Sprintf("赢家%d三豪华7对 +3番", WinnerID))
	}
	//海底捞+1
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_HAI_DI != 0 {
		winerFanNum += 1
		fmt.Println(fmt.Sprintf("赢家%d海底捞 +1番", WinnerID))
	}
	//清一色明四归加番
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_QING_YI_SE != 0 {
		if WinnerItem.Ctx.ChiHuResult.ChiHuKind&CHK_M4G_1 != 0 {
			winerFanNum += 1
			fmt.Println(fmt.Sprintf("赢家%d清一色明四归1个 +1番", WinnerID))
		} else if WinnerItem.Ctx.ChiHuResult.ChiHuKind&CHK_M4G_2 != 0 {
			winerFanNum += 2
			fmt.Println(fmt.Sprintf("赢家%d清一色明四归2个 +2番", WinnerID))
		} else if WinnerItem.Ctx.ChiHuResult.ChiHuKind&CHK_M4G_3 != 0 {
			winerFanNum += 3
			fmt.Println(fmt.Sprintf("赢家%d清一色明四归3个 +3番", WinnerID))
		}
	}

	//过胡翻番,赢了才算,输不翻番
	if sp.Rule.GHFF && WinnerItem.Ctx.ChiHuResult.ChiHuKind != 0 {
		winerFanNum += sp.GuoHuCount[WinnerID]
		fmt.Println(fmt.Sprintf("赢家%d过胡翻番 +%d番", WinnerID, sp.GuoHuCount[WinnerID]))
	}

	//抢杠+1，抢杠算点炮
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_QIANG_GANG != 0 {
		//if wCurrentUser == wProvideUserID && wCurrentUser != WinnerID{
		//LoserFanNum += 1
		winerFanNum += 1
		fmt.Println(fmt.Sprintf("玩家%d被抢杠 +1番", wCurrentUser))
		//}
	}

	////多个大胡叠加
	//if dahuCount := sp.GetPlayerDaHuCount(WinnerItem.Ctx.ChiHuResult.ChiHuKind); dahuCount > 1 {
	//	winerFanNum += (dahuCount - 1 )
	//	fmt.Println(fmt.Sprintf("赢家%d多大胡叠加 +%d番",WinnerID,dahuCount-1))
	//}

	//以上是赢家牌型的加番,下面是涉及到输家相关的加番
	//点炮由点炮玩家付,点炮玩家
	//杠上炮
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_GANG_SHANG_PAO != 0 {
		gangPaoFan := 1
		if sp.Rule.GangPaox4 {
			//杠上炮 x4
			gangPaoFan = 2
		}
		if bIsBaoPei {
			//如果是包赔,这个算赢家杠开,加番记在赢家身上,点炮的一家出两家钱
			winerFanNum += gangPaoFan
		} else {
			//if wCurrentUser == wProvideUserID && wCurrentUser != WinnerID {
			LoserFanNum += gangPaoFan
			//}
		}
		fmt.Println(fmt.Sprintf("玩家%d杠上炮 点炮 +%d番", wCurrentUser, gangPaoFan))
	}
	//海底炮+1
	if WinnerItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_HAI_DI_PAO != 0 {
		if bIsBaoPei {
			//如果是包赔,这个算赢家海底捞
			winerFanNum += 1
		} else {
			//if wCurrentUser == wProvideUserID && wCurrentUser != WinnerID {
			LoserFanNum += 1
			//}
		}
		fmt.Println(fmt.Sprintf("玩家%d海底炮 +1番", wCurrentUser))
	}

	//亮倒+1
	if sp.Rule.DLFF {
		if sp.m_ShowCard[wCurrentUser].BIsShowCard {
			LoserFanNum += 1
			fmt.Println(fmt.Sprintf("对亮翻番 玩家%d亮倒 +1番", wCurrentUser))
		}
		if sp.m_ShowCard[WinnerID].BIsShowCard {
			winerFanNum += 1
			fmt.Println(fmt.Sprintf("对亮翻番 赢家%d亮倒 +1番", WinnerID))
		}
	} else {
		//非对亮翻番
		if sp.m_ShowCard[wCurrentUser].BIsShowCard && sp.m_ShowCard[WinnerID].BIsShowCard {
			winerFanNum += 1
			fmt.Println(fmt.Sprintf("赢家%d亮倒,玩家%d也亮倒,赢家 +1番", WinnerID, wCurrentUser))
		} else {
			if sp.m_ShowCard[wCurrentUser].BIsShowCard {
				LoserFanNum += 1
				fmt.Println(fmt.Sprintf("玩家%d亮倒 +1番", wCurrentUser))
			}
			if sp.m_ShowCard[WinnerID].BIsShowCard {
				winerFanNum += 1
				fmt.Println(fmt.Sprintf("赢家%d亮倒 +1番", wCurrentUser))
			}
		}
	}

	wfan := 1
	lfan := 1
	for i := 0; i < winerFanNum; i++ {
		wfan *= 2
	}
	for i := 0; i < LoserFanNum; i++ {
		lfan *= 2
	}

	fmt.Println(fmt.Sprintf("赢家%d 番数%d, 玩家%d 番数:%d ", WinnerID, wfan, wCurrentUser, lfan))
	return wfan, lfan
}

// 赔庄计算
func (sp *SportYCKWX) CalcPeiZhuang(gameend *static.Msg_S_GameEnd) uint16 {
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "========================[荒庄]========================[进入赔庄]")
	//找出荒庄时玩家中已听牌的玩家
	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.UserTing[i], sp.UserTingType[i] = sp.CheckTing(uint16(i))
		//游戏记录
		recordStr := fmt.Sprintf("玩家可胡的最大牌型是%s", sp.GetTingHuKind(sp.UserTingType[i]))
		sp.OnWriteGameRecord(uint16(i), recordStr)
		fmt.Println(fmt.Sprintf("玩家%d可胡的最大牌型是%s", i, sp.GetTingHuKind(sp.UserTingType[i])))
	}

	IsPei := [4]byte{} //是否赔分
	//赔庄
	var noTingUsers []*components2.Player   //未听牌的玩家
	var LiangDaoUsers []*components2.Player //亮倒的玩家
	var TingUsers []*components2.Player     //已听牌未亮牌的玩家
	for _, v := range sp.PlayerInfo {
		if sp.UserTing[v.Seat] {
			if sp.m_ShowCard[v.Seat].BIsShowCard {
				//追加亮倒玩家
				LiangDaoUsers = append(LiangDaoUsers, v)
			} else {
				//追加听牌未亮倒玩家
				TingUsers = append(TingUsers, v)
			}
		} else {
			//追加未听牌的玩家
			noTingUsers = append(noTingUsers, v)
		}
	}

	//1.（1）未听的赔亮倒的和听牌的
	for _, noting := range noTingUsers {
		//1.1 赔亮倒的
		for _, liangdao := range LiangDaoUsers {
			liangFan, notingFan := sp.GetPeiZhuangFan(liangdao.Seat, noting.Seat)
			peiScore := sp.GetPlayerHuScore(sp.UserTingType[liangdao.Seat]) * liangFan * notingFan
			if peiScore > sp.Rule.FengDing {
				peiScore = sp.Rule.FengDing
			}
			gameend.PeiScore[noting.Seat] -= peiScore * sp.Rule.DiFen
			gameend.PeiScore[liangdao.Seat] += peiScore * sp.Rule.DiFen

			IsPei[noting.Seat] = PeiFen_True
			IsPei[liangdao.Seat] = PeiFen_False
			gameend.TingFlag[liangdao.Seat] = PeiZhuang_k5x_Liang
			//游戏记录
			recordStr := fmt.Sprintf("玩家[%s]未听牌,赔分%d给亮倒玩家[%s],封顶分:%d", noting.Name, peiScore, liangdao.Name, sp.Rule.FengDing)
			sp.OnWriteGameRecord(noting.Seat, recordStr)
		}
		//1.2 赔听牌的
		for _, ting := range TingUsers {
			tingFan, notingFan := sp.GetPeiZhuangFan(ting.Seat, noting.Seat)
			peiScore := sp.GetPlayerHuScore(sp.UserTingType[ting.Seat]) * tingFan * notingFan
			if peiScore > sp.Rule.FengDing {
				peiScore = sp.Rule.FengDing
			}
			gameend.PeiScore[noting.Seat] -= peiScore * sp.Rule.DiFen
			gameend.PeiScore[ting.Seat] += peiScore * sp.Rule.DiFen

			IsPei[noting.Seat] = PeiFen_True
			IsPei[ting.Seat] = PeiFen_False
			gameend.TingFlag[ting.Seat] = PeiZhuang_k5x_Ting
			recordStr := fmt.Sprintf("玩家[%s]未听牌,赔分%d给听牌玩家[%s],封顶分:%d", noting.Name, peiScore, ting.Name, sp.Rule.FengDing)
			sp.OnWriteGameRecord(noting.Seat, recordStr)
		}
		gameend.TingFlag[noting.Seat] = PeiZhuang_k5x_NoTing
	}

	//2.（2）亮倒的赔听牌的
	for _, liangdao := range LiangDaoUsers {
		//2.1赔听牌的
		for _, ting := range TingUsers {
			tingFan, liangFan := sp.GetPeiZhuangFan(ting.Seat, liangdao.Seat)
			peiScore := sp.GetPlayerHuScore(sp.UserTingType[ting.Seat]) * tingFan * liangFan
			if peiScore > sp.Rule.FengDing {
				peiScore = sp.Rule.FengDing
			}
			gameend.PeiScore[liangdao.Seat] -= peiScore * sp.Rule.DiFen
			gameend.PeiScore[ting.Seat] += peiScore * sp.Rule.DiFen

			IsPei[liangdao.Seat] = PeiFen_True
			IsPei[ting.Seat] = PeiFen_False
			gameend.TingFlag[ting.Seat] = PeiZhuang_k5x_Ting
			recordStr := fmt.Sprintf("玩家[%s]亮倒,赔分%d给听牌玩家[%s],封顶分:%d", liangdao.Name, peiScore, ting.Name, sp.Rule.FengDing)
			sp.OnWriteGameRecord(liangdao.Seat, recordStr)
		}
		gameend.TingFlag[liangdao.Seat] = PeiZhuang_k5x_Liang
	}

	var nextBanker uint16 = static.INVALID_CHAIR
	var peiCount int     //赔分人数
	var noPeiSeat uint16 //未赔分玩家
	var peiSeat uint16   //赔分玩家
	for seat, peiType := range IsPei {
		if peiType == PeiFen_True {
			peiCount++
			peiSeat = uint16(seat)
		}
		if peiType == PeiFen_False {
			noPeiSeat = uint16(seat)
		}
	}

	if peiCount > 0 {
		if peiCount == 1 {
			//1人赔分时，次局赔分者坐庄
			nextBanker = peiSeat
		} else if peiCount == 2 {
			//2人赔分时，次局未赔分者坐庄
			nextBanker = noPeiSeat
		}

	}
	return nextBanker
}

// 获取玩家的数坎分
func (sp *SportYCKWX) CalcShuKan(winnerID uint16, chihuCard byte, bZimo bool) int {
	if !sp.Rule.ShuKan {
		return 0
	}

	//胡牌玩家才会计算数坎分
	if winnerID == static.INVALID_CHAIR {
		return 0
	}
	winItem := sp.GetUserItemByChair(winnerID)
	if winItem == nil {
		return 0
	}

	sp.OnWriteGameRecord(static.INVALID_CHAIR, "========================[进入数坎]========================")

	kanCount := sp.m_GameLogic.AnalyseShukan(winItem.Ctx.CardIndex[:], winItem.Ctx.WeaveItemArray[:], winItem.Ctx.WeaveItemCount, chihuCard, bZimo)
	recordStr := fmt.Sprintf("数坎:%d个", kanCount)
	//保存玩家数坎个数
	sp.KanCount[winnerID] = kanCount
	//写记录
	sp.OnWriteGameRecord(winnerID, recordStr)
	fmt.Println(fmt.Sprintf("玩家%d数坎%d个", winnerID, kanCount))
	return kanCount
}

// 更新亮牌数据,亮牌后玩家还可以杠牌,更新数据用于断线重连
func (sp *SportYCKWX) UpdateShowCard(wChairID uint16, operateCard byte) {
	//校验玩家是否亮牌
	if !sp.m_ShowCard[wChairID].BIsShowCard {
		return
	}

	for i := 0; i < len(sp.m_ShowCard[wChairID].CbLiangCard); i++ {
		if sp.m_ShowCard[wChairID].CbLiangCard[i] == operateCard {
			sp.m_ShowCard[wChairID].CbLiangCard[i] = 0
		}
	}
}

// ! 检查是都达到胡牌番数
func (sp *SportYCKWX) SuanFen(GameEnd *static.Msg_S_GameEnd) uint16 {
	var nWinnerCnt = 0                        //胡牌人数
	var nWinner uint16 = static.INVALID_CHAIR //赢家或者一炮多响点炮者
	var winnerList []*components2.Player
	var nextBanker uint16 = static.INVALID_CHAIR

	//找出赢家,丢进winnerList
	for _, v := range sp.PlayerInfo {
		if v.Ctx.ChiHuResult.ChiHuKind != static.CHK_NULL && v.Ctx.UserAction32 != static.WIK_NULL {
			nWinnerCnt++
			nWinner = v.Seat
			winnerList = append(winnerList, v)
			GameEnd.WWinner[nWinner] = true
			GameEnd.WChiHuKind[nWinner] = v.Ctx.ChiHuResult.ChiHuKind

			//游戏记录
			recordStr := fmt.Sprintf("赢了, 胡牌牌型 :%s ", sp.GetTingHuKind(v.Ctx.ChiHuResult.ChiHuKind))
			sp.OnWriteGameRecord(v.Seat, recordStr)
		}
	}

	if nWinnerCnt > 0 {
		//有赢家
		//1.自摸
		if sp.ProvideUser == nWinner && nWinnerCnt == 1 {
			//自摸三家付
			//买马
			maScore := sp.OnMaiMa(nWinner, GameEnd)
			kanCount := sp.CalcShuKan(winnerList[0].Seat, sp.ChiHuCard, true)

			//算分
			winnerHuScore := sp.GetPlayerHuScore(winnerList[0].Ctx.ChiHuResult.ChiHuKind)
			for i := 0; i < sp.GetPlayerCount(); i++ {
				if uint16(i) == nWinner {
					continue
				}
				winnerFan, loserFan := sp.GetPlayerFan(winnerList[0].Seat, sp.ProvideUser, uint16(i), false)
				huScore := winnerHuScore * winnerFan * loserFan
				if huScore > sp.Rule.FengDing {
					huScore = sp.Rule.FengDing
				}
				payscore := (huScore + sp.GetPlayerPiao(winnerList[0].Seat) + sp.GetPlayerPiao(uint16(i)) + int(maScore) + kanCount) * sp.Rule.DiFen
				GameEnd.GameScore[nWinner] += payscore
				GameEnd.GameScore[i] -= payscore
			}
			//自摸统计
			GameEnd.ProvideUser = nWinner
			GameEnd.Winner = nWinner
			winnerList[0].Ctx.HuBySelfCount++
			nextBanker = nWinner
		} else {
			//点炮由点炮玩家付,可能存在一炮多响的情况
			if sp.ProvideUser != static.INVALID_CHAIR {
				//sp.CalcBaoPei(GameEnd, winnerList)
				for _, v := range winnerList {
					//未亮牌玩家点炮亮牌玩家所听的牌,需要包赔
					if sp.IsCurUserTingCard(v.Seat, sp.ProvideCard) && !sp.m_ShowCard[sp.ProvideUser].BIsShowCard {
						GameEnd.Contractor = sp.ProvideUser
					}
					bIsBaoPei := false
					if GameEnd.Contractor != static.INVALID_CHAIR {
						bIsBaoPei = true
					}
					//数坎
					kanCount := sp.CalcShuKan(v.Seat, sp.ChiHuCard, false)
					winnerHuScore := sp.GetPlayerHuScore(v.Ctx.ChiHuResult.ChiHuKind)
					if bIsBaoPei {
						for i := 0; i < sp.GetPlayerCount(); i++ {
							if uint16(i) == v.Seat {
								continue
							}
							winnerFan, loserFan := sp.GetPlayerFan(v.Seat, sp.ProvideUser, uint16(i), true)
							huScore := winnerHuScore * winnerFan * loserFan
							if huScore > sp.Rule.FengDing {
								huScore = sp.Rule.FengDing
							}
							payscore := (huScore + sp.GetPlayerPiao(v.Seat) + sp.GetPlayerPiao(uint16(i)) + kanCount) * sp.Rule.DiFen
							GameEnd.GameScore[v.Seat] += payscore
							GameEnd.GameScore[sp.ProvideUser] -= payscore
						}
					} else {
						winnerFan, loserFan := sp.GetPlayerFan(v.Seat, sp.ProvideUser, uint16(sp.ProvideUser), bIsBaoPei)
						huScore := winnerHuScore * winnerFan * loserFan
						if huScore > sp.Rule.FengDing {
							huScore = sp.Rule.FengDing
						}
						payscore := (huScore + sp.GetPlayerPiao(v.Seat) + sp.GetPlayerPiao(sp.ProvideUser) + kanCount) * sp.Rule.DiFen
						GameEnd.GameScore[v.Seat] += payscore
						GameEnd.GameScore[sp.ProvideUser] -= payscore
					}

					//吃胡统计
					v.Ctx.ChiHuUserCount++
				}

				//点炮统计
				GameEnd.ProvideUser = sp.ProvideUser
				provideItem := sp.GetUserItemByChair(sp.ProvideUser)
				if provideItem == nil {
					return nextBanker
				}
				provideItem.Ctx.ProvideUserCount++

				if nWinnerCnt == 1 {
					nextBanker = winnerList[0].Seat
				} else {
					nextBanker = sp.ProvideUser
				}
			}
		}
	} else {
		//流局
		xlog.Logger().Debug("荒庄")
		// 慌庄所有人都是0分
		for i := 0; i < sp.GetPlayerCount(); i++ {
			GameEnd.GameScore[i] = 0
		}
		//流局处理
		GameEnd.HuangZhuang = true
		GameEnd.ChiHuCard = 0
		GameEnd.ChiHuUserCount = 0
		GameEnd.Winner = static.INVALID_CHAIR
		sp.HaveHuangZhuang = true

		//赔庄
		nextBanker = sp.CalcPeiZhuang(GameEnd)
	}

	return nextBanker
}

////包赔, 返回是否包赔
//func (self *SportYCKWX) CalcBaoPei(GameEnd * static.Msg_S_GameEnd, winnerList []*components2.Player) bool{
//	//校验
//	if len(winnerList) <= 0 {
//		return false
//	}
//
//	//包赔一定是点炮的情况
//	for _, v := range winnerList {
//		if v.Seat == sp.ProvideUser{
//			return false
//		}
//	}
//
//	//点炮由点炮玩家付,可能存在一炮多响的情况
//	if sp.ProvideUser != static.INVALID_CHAIR {
//		for _, v := range winnerList {
//			//未亮牌玩家点炮亮牌玩家所听的牌,需要包赔
//			if sp.IsShowCard(v.Seat, sp.ProvideCard) && !sp.m_ShowCard[sp.ProvideUser].BIsShowCard{
//				GameEnd.Contractor = sp.ProvideUser
//			}
//
//			if GameEnd.Contractor != static.INVALID_CHAIR{
//				//数坎
//				kanCount := sp.CalcShuKan(v.Seat, sp.ChiHuCard, false)
//				winnerHuScore := sp.GetPlayerHuScore(v.Ctx.ChiHuResult.ChiHuKind)
//				for i:=0;i<sp.GetPlayerCount();i++ {
//					if uint16(i) == v.Seat {
//						continue
//					}
//					winnerFan, loserFan := sp.GetPlayerFan(v.Seat, sp.ProvideUser, uint16(i), true)
//				}
//			}
//		}
//	}
//}

// 获取玩家漂分
func (sp *SportYCKWX) GetPlayerPiao(wCurrentUser uint16) int {
	userItem := sp.GetUserItemByChair(wCurrentUser)
	if userItem == nil {
		return 0
	}

	return userItem.Ctx.VecXiaPao.Num
}

// 判断这张牌是不是玩家已经亮出的牌
func (sp *SportYCKWX) IsShowCard(wCurrentUser uint16, card byte) bool {
	if card != 0 {
		for _, v := range sp.m_ShowCard[wCurrentUser].CbLiangCard {
			if v != 0 && v == card {
				return true
			}
		}
	}

	return false
}

// 判断这张牌是不是玩家已经听的牌
func (sp *SportYCKWX) IsCurUserTingCard(wCurrentUser uint16, card byte) bool {
	if card != 0 {
		for _, v := range sp.m_ShowCard[wCurrentUser].CbTingCard {
			if v != 0 && v == card {
				return true
			}
		}
	}

	return false
}

// ! 检查是都达到胡牌番数
func (sp *SportYCKWX) CheckFan(wCurrentUser uint16, wProvideUser uint16) bool {
	//校验
	if wCurrentUser == static.INVALID_CHAIR {
		return false
	}

	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return false
	}

	winnerHuScore := sp.GetPlayerHuScore(_userItem.Ctx.ChiHuResult.ChiHuKind)
	hufan, lfan := sp.GetPlayerFan(wCurrentUser, wProvideUser, wProvideUser, false)

	// 勾选了2番起胡,那屁胡必须满足2番,也就是必须有杠开、杠炮、海底捞、海底炮来加番
	if sp.Rule.QiHuX2 {
		if !sp.IsDaHu(_userItem.Ctx.ChiHuResult.ChiHuKind) && hufan*lfan < 2 {
			sp.OnWriteGameRecord(wCurrentUser, "2番起胡, 屁胡不够2番,胡不了")
			return false
		}
	} else {
		//未勾选2番起胡, 那么捉铳必须2分起胡, 自摸无限制
		if winnerHuScore*hufan*lfan < 2 {
			if wCurrentUser != wProvideUser {
				sp.OnWriteGameRecord(wCurrentUser, "捉铳必须2分起胡,胡不了")
				return false
			}
		}
	}

	return true
}

func (sp *SportYCKWX) GetTingHuKind(kind uint64) string {
	str := ""
	for k := 0; k < len(k5x_KindMask); k++ {
		if (kind & uint64(k5x_KindMask[k])) != 0 {
			switch k5x_KindMask[k] {
			case static.CHK_QIANG_GANG:
				str += " 抢杠胡 "
			case static.CHK_7_DUI:
				str += " 七对 "
			case static.CHK_7_DUI_1:
				str += " 豪华七对 "
			case static.CHK_7_DUI_2:
				str += " 双豪华七对 "
			case static.CHK_7_DUI_3:
				str += " 三豪华七对 "
			case static.CHK_XIAO_SAN_YUAN:
				str += " 小三元 "
			case static.CHK_DA_SAN_YUAN:
				str += " 大三元 "
			case static.CHK_KA_5_XING:
				str += " 卡五星 "
			case static.CHK_SHOU_ZHUA_YI:
				str += " 手抓一 "
			case static.CHK_SI_GUI_YI_MING:
				str += " 明四归一 "
			case static.CHK_SI_GUI_YI_AN:
				str += " 暗四归一 "
			case static.CHK_PENG_PENG:
				str += " 碰碰胡 "
			case static.CHK_QING_YI_SE:
				str += " 清一色 "
			case static.CHK_HAI_DI:
				str += " 海底捞 "
			case static.CHK_GANG_SHANG_KAI_HUA:
				str += " 杠上开花 "
			}
		}
	}

	if kind&static.CHK_PING_HU_NOMAGIC != 0 {
		str += " 平胡"
	}

	if str == "" {
		str += " 胡不了"
	}

	return str
}

func (sp *SportYCKWX) getWriteHandReplayRecordCString(replayRecord meta2.K5x_Replay_Record) string {
	handCardStr := ""
	for i := 0; i < sp.GetPlayerCount(); i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < static.MAX_COUNT; j++ {
			handCardStr += fmt.Sprintf("%02x,", replayRecord.RecordHandCard[i][j])
		}
	}

	//写入分数
	handCardStr += "S:"

	for i := 0; i < sp.GetPlayerCount(); i++ {
		handCardStr += fmt.Sprintf("%d,", replayRecord.Score[i])
	}

	return handCardStr
}

// ! 游戏出牌记录
func (sp *SportYCKWX) TableWriteOutDate(playNum int, replayRecord meta2.K5x_Replay_Record) {
	recordReplay := new(models.RecordGameReplay)
	recordReplay.GameNum = sp.GetTableInfo().GameNum
	recordReplay.RoomNum = sp.GetTableInfo().Id
	recordReplay.PlayNum = playNum
	recordReplay.ServerId = wuhan.GetServer().Con.Id
	recordReplay.HandCard = sp.getWriteHandReplayRecordCString(replayRecord)
	recordReplay.OutCard = sp.getWriteOutReplayRecordCString(replayRecord)
	recordReplay.KindID = sp.GetTableInfo().KindId
	recordReplay.CardsNum = int(replayRecord.LeftCardCount)
	recordReplay.UVitaminMap = replayRecord.UVitamin
	recordReplay.CreatedAt = time.Now()

	wuhan.GetDBMgr().InsertGameRecordReplay(recordReplay)

	sp.RoundReplayId = recordReplay.Id
}

func (sp *SportYCKWX) getWriteOutReplayRecordCString(replayRecord meta2.K5x_Replay_Record) string {
	ourCardStr := ""
	ourCardStr += fmt.Sprintf("P:%02x,", replayRecord.PiziCard)
	if replayRecord.Fengquan > 0 {
		ourCardStr += fmt.Sprintf("f:%02d,", replayRecord.Fengquan)
	}

	hasHu := false
	// 把胡牌的U拿出来
	endMsgUpdateScore := [meta2.MAX_PLAYER]float64{}
	for i := 0; i < len(replayRecord.VecOrder); i++ {
		recordI := replayRecord.VecOrder[i]
		if recordI.Operation == info2.E_Hu {
			for j, count := i, 0; j < len(replayRecord.VecOrder); j++ {
				recordJ := replayRecord.VecOrder[j]
				if recordJ.Operation == info2.E_GameScore {
					count++
					if count > sp.GetPlayerCount() {
						break
					}
					recordJ.Operation = -1 // 置为无效
					endMsgUpdateScore[recordJ.Chair_id] = recordJ.UserScore
					replayRecord.VecOrder[j] = recordJ
				}
			}
			hasHu = true
			break
		}
	}

	for _, record := range replayRecord.VecOrder {
		if record.Operation < 0 {
			continue
		}
		ourCardStr += fmt.Sprintf("%d:", record.Chair_id)
		switch record.Operation {
		case info2.E_SendCard:
			ourCardStr += fmt.Sprintf("S%02x,", record.Value[0])
			break
		case info2.E_OutCard:
			ourCardStr += fmt.Sprintf("O%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG:
			ourCardStr += fmt.Sprintf("o%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG_Magic:
			ourCardStr += fmt.Sprintf("m%02x,", record.Value[0])
			break
		case info2.E_OutCard_TG_PiZi:
			ourCardStr += fmt.Sprintf("z%02x,", record.Value[0])
			break
		case info2.E_Wik_Left:
			ourCardStr += fmt.Sprintf("L%02x,", record.Value[0])
			break
		case info2.E_Wik_Center:
			ourCardStr += fmt.Sprintf("C%02x,", record.Value[0])
			break
		case info2.E_Wik_Right:
			ourCardStr += fmt.Sprintf("R%02x,", record.Value[0])
			break
		case info2.E_Peng:
			ourCardStr += fmt.Sprintf("P%02x,", record.Value[0])
			break
		case info2.E_Gang:
			ourCardStr += fmt.Sprintf("G%02x,", record.Value[0])
			break
		case info2.E_Gang_HongZhongGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_FaCaiGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_PiziGand:
			ourCardStr += fmt.Sprintf("Z%02x,", record.Value[0])
			break
		case info2.E_Gang_LaiziGand:
			ourCardStr += fmt.Sprintf("M%02x,", record.Value[0])
			break
		case info2.E_Gang_XuGand:
			ourCardStr += fmt.Sprintf("X%02x,", record.Value[0])
			break
		case info2.E_Gang_AnGang:
			ourCardStr += fmt.Sprintf("A%02x,", record.Value[0])
			break
		case info2.E_Qiang:
			ourCardStr += fmt.Sprintf("Q%02x,", record.Value[0])
			break
		case info2.E_Hu:
			ourCardStr += fmt.Sprintf("H%02x,", record.Value[0])
			break
		case info2.E_HuangZhuang:
			ourCardStr += fmt.Sprintf("N%02x,", record.Value[0])
			break
		case info2.E_Bird:
			ourCardStr += fmt.Sprintf("B%02x,", record.Value[0])
			break
		case info2.E_Li_Xian:
			ourCardStr += fmt.Sprintf("l%02x,", record.Value[0])
			break
		case info2.E_Jie_san:
			ourCardStr += fmt.Sprintf("j%02x,", record.Value[0])
			break
		case info2.E_Pao:
			ourCardStr += fmt.Sprintf("K%02x,", record.Value[0])
			break
		case info2.E_SendCardRight:
			ourCardStr += fmt.Sprintf("s%02x,", record.Value[0])
			break
		case info2.E_HandleCardRight:
			ourCardStr += fmt.Sprintf("h%02x,", record.Value[0])
			break
		case info2.E_Gang_ChaoTianGand:
			ourCardStr += fmt.Sprintf("T%02x,", record.Value[0])
			break
		case info2.E_Gang_SmallChaoTianGand:
			ourCardStr += fmt.Sprintf("a%02x,", record.Value[0]) //小朝天
			break
		case info2.E_Baoqing:
			ourCardStr += fmt.Sprintf("q%02x,", record.Value[0])
			break
		case info2.E_Baojiang:
			ourCardStr += fmt.Sprintf("J%02x,", record.Value[0])
			break
		case info2.E_Baofeng:
			ourCardStr += fmt.Sprintf("F%02x,", record.Value[0])
			break
		case info2.E_Baoqi:
			ourCardStr += fmt.Sprintf("D%02x,", record.Value[0])
			break
		case info2.E_BaoTing:
			ourCardStr += fmt.Sprintf("t%02x,", record.Value[0])
			break
		case info2.E_ExchageThreeCard: //换三张换的牌
			ourCardStr += fmt.Sprintf("E%02x,", record.Value[0])
			break
		case info2.E_ExchageThreeEnd: //换三张结束
			ourCardStr += fmt.Sprintf("e%02x,", record.Value[0])
			break
		case info2.E_K5x_LiangPai: //卡五星亮牌
			ourCardStr += record.LiangPai
		case info2.E_GameScore:
			if fs := strings.Split(fmt.Sprintf("%v", record.UserScore), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s,", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f,", record.UserScore)
			}
		case info2.E_BuHua:
			ourCardStr += fmt.Sprintf("BH%02x,", record.Value[0])
			break
		case info2.E_Liang:
			ourCardStr += fmt.Sprintf("b%02x,", record.Value[0])
			break
		case info2.E_Change_Pizhi:
			ourCardStr += fmt.Sprintf("p%02x,", record.Value[0])
			break
		default:
			break
		}
	}

	for i := 0; i < sp.GetPlayerCount(); i++ {
		switch replayRecord.WBigHuKind[i] {
		case static.GameBigHuKindK5X: //卡五星
			ourCardStr += fmt.Sprintf("K5X%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindDSY: //大三元
			ourCardStr += fmt.Sprintf("DSY%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindXSY: //小三元
			ourCardStr += fmt.Sprintf("XSY%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindSZ1: //手抓一
			ourCardStr += fmt.Sprintf("SZY%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindA4G: //暗四归一
			ourCardStr += fmt.Sprintf("A4G%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindM4G: //明四归一
			ourCardStr += fmt.Sprintf("M4G%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindHDP: //海底炮
			ourCardStr += fmt.Sprintf("HDP%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindPP: //碰碰胡
			ourCardStr += fmt.Sprintf("PPH%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindHDL: //海底捞
			ourCardStr += fmt.Sprintf("HDL%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindGSK: //杠上开花
			ourCardStr += fmt.Sprintf("GSK%d,", replayRecord.ProvideUser)
		case static.GameBigHuKind_7: //七对
			ourCardStr += fmt.Sprintf("QID%d,", replayRecord.ProvideUser)
		case static.GameBigHuKind_7Dui_1: //豪华七对
			ourCardStr += fmt.Sprintf("QID%d,", replayRecord.ProvideUser)
		case static.GameBigHuKind_7Dui_2: //超豪华七对
			ourCardStr += fmt.Sprintf("QID%d,", replayRecord.ProvideUser)
		case static.GameBigHuKind_7Dui_3: //超超豪华七对
			ourCardStr += fmt.Sprintf("QID%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindQYS: //清一色
			ourCardStr += fmt.Sprintf("QYS%d,", replayRecord.ProvideUser)
		case static.GameBigHuKindQG: //抢杠
			ourCardStr += fmt.Sprintf("QGH%d,", replayRecord.ProvideUser)
		case 13:
			ourCardStr += fmt.Sprintf("QSB%d,", replayRecord.ProvideUser)
		case static.GameNoMagicHu:
			ourCardStr += fmt.Sprintf("YHM%d,", replayRecord.ProvideUser)
			break
		case static.GameMagicHu:
			ourCardStr += fmt.Sprintf("RHM%d,", replayRecord.ProvideUser)
			break
		default:
			ourCardStr += fmt.Sprintf("NIL%d,", replayRecord.ProvideUser)
		}
	}
	// 最后补上胡牌U
	if hasHu {
		for i, s := range endMsgUpdateScore {
			if i >= sp.GetPlayerCount() {
				break
			}
			ourCardStr += fmt.Sprintf("%d:", i)
			if fs := strings.Split(fmt.Sprintf("%v", s), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s,", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f,", s)
			}
		}
	}

	return ourCardStr
}

// 返回马分
func (sp *SportYCKWX) OnMaiMa(nWinner uint16, GameEnd *static.Msg_S_GameEnd) byte {
	//校验
	if nWinner == static.INVALID_CHAIR || sp.LeftCardCount <= 0 {
		return 0
	}
	//自摸买马
	maCard, maScore := byte(0), byte(0)
	if sp.Rule.MaiMaType == MaiMa_k5x_LiangDaoZiMo {
		//亮倒自摸买马
		if sp.m_ShowCard[nWinner].BIsShowCard {
			maCard, maScore = sp.Maima(nWinner)
			fmt.Println(fmt.Sprintf("玩家%d亮倒自摸买马%x,马分%d", nWinner, maCard, maScore))
		}
	} else if sp.Rule.MaiMaType == MaiMa_k5x_ZiMo {
		maCard, maScore = sp.Maima(nWinner)
		fmt.Println(fmt.Sprintf("玩家%d自摸买马%x,马分%d", nWinner, maCard, maScore))
	}

	GameEnd.CbBirdData[nWinner][0] = maCard
	return maScore
}

// 判断当前牌是否是其它亮倒玩家所听胡的牌
func (sp *SportYCKWX) IsOthersTingHuCard(wCurrentUser uint16, cbCurrentCard byte) bool {
	if wCurrentUser == static.INVALID_CHAIR {
		return false
	}
	var LiangDaoUsers []*components2.Player //亮倒的玩家
	for _, v := range sp.PlayerInfo {
		if v.Seat == wCurrentUser {
			//排除当前玩家
			continue
		}
		if sp.m_ShowCard[v.Seat].BIsShowCard {
			//追加亮倒玩家
			LiangDaoUsers = append(LiangDaoUsers, v)
		}
	}
	if len(LiangDaoUsers) > 0 {
		for _, v := range LiangDaoUsers {
			for _, card := range sp.m_ShowCard[v.Seat].CbTingCard {
				if card != 0 {
					if cbCurrentCard != 0 && cbCurrentCard == card {
						return true
					}
				}
			}
		}
	}
	return false
}

func (sp *SportYCKWX) SaveXuGangScore(wChairID uint16, msg meta2.Msg_S_OperateScore_K5X) {
	sp.bQiangGangScoreSend = false
	sp.QiangGangOperateScore = msg
}

func (sp *SportYCKWX) ResetXuGangScore() {
	sp.bQiangGangScoreSend = true
	var tmp meta2.Msg_S_OperateScore_K5X
	sp.QiangGangOperateScore = tmp
}

func (sp *SportYCKWX) ReCoverXuGangScore() {
	if _userItem := sp.GetUserItemByChair(sp.QiangGangOperateScore.OperateUser); _userItem != nil {
		_userItem.Ctx.XuGangAction()
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if sp.QiangGangOperateScore.OperateUser != uint16(i) {
				sp.OnUserScoreOffset(uint16(i), sp.QiangGangOperateScore.ScoreOffset[i])
			}
		}
		sp.OnUserScoreOffset(sp.QiangGangOperateScore.OperateUser, sp.QiangGangOperateScore.ScoreOffset[sp.QiangGangOperateScore.OperateUser])

		sp.QiangGangOperateScore.GameScore, sp.QiangGangOperateScore.GameVitamin = sp.OnSettle(sp.QiangGangOperateScore.ScoreOffset, consts.EventSettleGaming)
		sp.SendTableMsg(consts.MsgTypeGameOperateScore, sp.QiangGangOperateScore)
		sp.ResetXuGangScore()
	}
}

func (sp *SportYCKWX) OnScoreOffset(total []float64) {
	for i, t := range total {
		var order meta2.K5x_Replay_Order
		order.Chair_id = uint16(i)
		order.Operation = info2.E_GameScore
		order.UserScore = t
		sp.ReplayRecord.VecOrder = append(sp.ReplayRecord.VecOrder, order)
	}
}

func (sp *SportYCKWX) GetUserGuoHuFan(wChairID int) int {
	if !sp.Rule.GHFF {
		return 0
	}
	fan := 1
	if sp.GuoHuCount[wChairID] > 0 {
		for i := 0; i < sp.GuoHuCount[wChairID]; i++ {
			fan *= 2
		}
		return fan
	}

	return 0
}

func (sp *SportYCKWX) GetPeiZhuangFan(wCurrentUser uint16, wPeiUser uint16) (int, int) {

	winerFanNum := 0
	LoserFanNum := 0
	//亮倒+1
	if sp.Rule.DLFF {
		if sp.m_ShowCard[wPeiUser].BIsShowCard {
			LoserFanNum += 1

			//游戏记录
			recordStr := fmt.Sprintf("对亮翻番 玩家%d亮倒 赔+1番", wPeiUser)
			fmt.Println(recordStr)
			sp.OnWriteGameRecord(wPeiUser, recordStr)
		}
		if sp.m_ShowCard[wCurrentUser].BIsShowCard {
			winerFanNum += 1

			//游戏记录
			recordStr := fmt.Sprintf("对亮翻番 赢家%d亮倒 +1番", wCurrentUser)
			fmt.Println(recordStr)
			sp.OnWriteGameRecord(wCurrentUser, recordStr)
		}
	} else {
		//非对亮翻番
		if sp.m_ShowCard[wCurrentUser].BIsShowCard && sp.m_ShowCard[wPeiUser].BIsShowCard {
			winerFanNum += 1
			//游戏记录
			recordStr := fmt.Sprintf("赢家%d亮倒,玩家%d也亮倒,赢家 +1番", wCurrentUser, wPeiUser)
			fmt.Println(recordStr)
			sp.OnWriteGameRecord(wCurrentUser, recordStr)
		} else {
			if sp.m_ShowCard[wPeiUser].BIsShowCard {
				LoserFanNum += 1
				//游戏记录
				recordStr := fmt.Sprintf("玩家%d亮倒 +1番", wCurrentUser)
				fmt.Println(recordStr)
				sp.OnWriteGameRecord(wCurrentUser, recordStr)
			}
			if sp.m_ShowCard[wCurrentUser].BIsShowCard {
				winerFanNum += 1
				//游戏记录
				recordStr := fmt.Sprintf("赢家%d亮倒 +1番", wCurrentUser)
				fmt.Println(recordStr)
				sp.OnWriteGameRecord(wCurrentUser, recordStr)
			}
		}
	}

	//豪华七对+1
	if sp.UserTingType[wCurrentUser]&static.CHK_7_DUI_1 != 0 {
		winerFanNum += 1
		//游戏记录
		recordStr := fmt.Sprintf("赢家%d豪华7对 +1番", wCurrentUser)
		fmt.Println(recordStr)
		sp.OnWriteGameRecord(wCurrentUser, recordStr)
	}
	//双豪华七对+2
	if sp.UserTingType[wCurrentUser]&static.CHK_7_DUI_2 != 0 {
		winerFanNum += 2
		//游戏记录
		recordStr := fmt.Sprintf("赢家%d双豪华7对 +2番", wCurrentUser)
		fmt.Println(recordStr)
		sp.OnWriteGameRecord(wCurrentUser, recordStr)
	}
	//三豪华七对+3
	if sp.UserTingType[wCurrentUser]&static.CHK_7_DUI_3 != 0 {
		winerFanNum += 3
		//游戏记录
		recordStr := fmt.Sprintf("赢家%d三豪华7对 +3番", wCurrentUser)
		fmt.Println(recordStr)
		sp.OnWriteGameRecord(wCurrentUser, recordStr)
	}

	//清一色明四归加番
	if sp.UserTingType[wCurrentUser]&static.CHK_QING_YI_SE != 0 {
		if sp.UserTingType[wCurrentUser]&CHK_M4G_1 != 0 {
			winerFanNum += 1
			//游戏记录
			recordStr := fmt.Sprintf("赢家%d清一色明四归1个 +1番", wCurrentUser)
			fmt.Println(recordStr)
			sp.OnWriteGameRecord(wCurrentUser, recordStr)
		} else if sp.UserTingType[wCurrentUser]&CHK_M4G_2 != 0 {
			winerFanNum += 2
			//游戏记录
			recordStr := fmt.Sprintf("赢家%d清一色明四归2个 +2番", wCurrentUser)
			fmt.Println(recordStr)
			sp.OnWriteGameRecord(wCurrentUser, recordStr)
		} else if sp.UserTingType[wCurrentUser]&CHK_M4G_3 != 0 {
			winerFanNum += 3
			//游戏记录
			recordStr := fmt.Sprintf("赢家%d清一色明四归3个 +3番", wCurrentUser)
			fmt.Println(recordStr)
			sp.OnWriteGameRecord(wCurrentUser, recordStr)
		}
	}

	wfan := 1
	lfan := 1
	for i := 0; i < winerFanNum; i++ {
		wfan *= 2
	}
	for i := 0; i < LoserFanNum; i++ {
		lfan *= 2
	}

	fmt.Println(fmt.Sprintf("赢家%d 番数%d, 玩家%d 番数:%d ", wCurrentUser, wfan, wPeiUser, lfan))
	return wfan, lfan
}

// 校验当前牌是否在手牌中
func (sp *SportYCKWX) IsCurCardInHand(_useritem *components2.Player, curCard byte) bool {
	//校验
	if _useritem == nil || curCard == 0 {
		return false
	}

	//构造麻将
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, _useritem.Ctx.CardIndex[:])
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] != 0 {
			if curCard != sp.m_GameLogic.SwitchToCardData(byte(i)) {
				return false
			}
		}
	}

	return true
}

// 校验亮牌数据
func (sp *SportYCKWX) CheckCardInHand(_useritem *components2.Player, cardArray []byte) bool {
	if _useritem == nil || len(cardArray) == 0 {
		return false
	}

	//构造麻将
	cbCardIndexTemp := make([]int, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, static.HF_BytesToInts(_useritem.Ctx.CardIndex[:]))

	for _, card := range cardArray {
		if card != 0 {
			for i := 0; i < static.MAX_INDEX; i++ {
				if cbCardIndexTemp[i] != 0 {
					if card == sp.m_GameLogic.SwitchToCardData(byte(i)) {
						cbCardIndexTemp[i]--
					}
				}

				if cbCardIndexTemp[i] < 0 {
					return false
				}
			}
		}
	}

	return true
}

// 解除托管
func (sp *SportYCKWX) OnUserTustee(msg *static.Msg_S_DG_Trustee) bool {
	if sp.Rule.Overtime_trust < 1 {
		return false
	}
	item := sp.GetUserItemByChair(msg.ChairID)
	if item == nil {
		return false
	}
	if item.CheckTRUST() == msg.Trustee {
		return true
	}
	var tuoguan static.Msg_S_DG_Trustee
	tuoguan.ChairID = msg.ChairID
	tuoguan.Trustee = msg.Trustee
	//校验规则
	if tuoguan.ChairID < static.MAX_PLAYER_4P {
		if tuoguan.Trustee == true /*&& (sp.GameState != gameserver.GsNull)*/ {
			item.ChangeTRUST(true)
			//进入托管啥都不用做
			fmt.Println(fmt.Sprintf("玩家%d进入托管", tuoguan.ChairID))
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			return true
		} else if tuoguan.Trustee == false {
			item.ChangeTRUST(false)
			//如果是当前的玩家，那么重新设置一下开始时间
			if tuoguan.ChairID == sp.CurrentUser {
				_item := sp.GetUserItemByChair(sp.CurrentUser)
				if _item != nil {
					sp.LockTimeOut(_item.Seat, sp.Rule.Overtime_trust)
					tuoguan.Overtime = _item.Ctx.CheckTimeOut
				}
			}
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			return false
		}
	} else {
		//详细日志
		LogStr := string("托管动作:游戏状态不正确 ")
		sp.OnWriteGameRecord(tuoguan.ChairID, LogStr)
		return false
	}
	return false
}
func (sp *SportYCKWX) accountManage(dismisstime int, trusttime int, autonexttimer int) int {
	if !(int(sp.CurCompleteCount) >= sp.Rule.JuShu) && sp.CurCompleteCount != 0 {
		check := false
		if dismisstime != -1 {
			for i := 0; i < static.MAX_PLAYER_4P; i++ {
				if item := sp.GetUserItemByChair(uint16(i)); item != nil {
					if item.CheckTRUST() {
						if check {
							var _msg = &static.Msg_C_DismissFriendResult{
								Id:   item.Uid,
								Flag: true,
							}
							sp.OnDismissResult(item.Uid, _msg)
						} else {
							check = true
							var msg = &static.Msg_C_DismissFriendReq{
								Id: item.Uid,
							}
							sp.SetDismissRoomTime(dismisstime)
							sp.OnDismissFriendMsg(item.Uid, msg)
						}
					}
				}
			}
		}
		if !check && trusttime > 0 && autonexttimer > 0 {
			//sp.SetAutoNextTimer(autonexttimer) //自动开始下一局
			return autonexttimer
		}
	}
	return 0
}

func (sp *SportYCKWX) TrustOutCard(wChairID uint16) {

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	if sp.CurrentUser == wChairID {
		//托管出牌：摸啥打啥，不能打的随机打一张其他的。全手都是不能打的牌时，摸啥打啥。
		if sp.SendCardData == static.INVALID_BYTE {
			IsOK := false
			for i := 0; i < static.MAX_INDEX; i++ {
				if _userItem.Ctx.CardIndex[i] != 0 {
					card := sp.m_GameLogic.SwitchToCardData(byte(i))
					if sp.IsOthersTingHuCard(wChairID, card) {
						continue
					}
					_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, card, false)
					if !sp.OnUserOutCard(_msg, false) {
						sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
					}
					IsOK = true
					return
				}
			}
			if !IsOK {
				for i := 0; i < static.MAX_INDEX; i++ {
					if _userItem.Ctx.CardIndex[i] != 0 {
						card := sp.m_GameLogic.SwitchToCardData(byte(i))
						_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, card, false)
						if !sp.OnUserOutCard(_msg, false) {
							sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
						}
						return
					}
				}
			}
		} else {
			if sp.IsOthersTingHuCard(wChairID, sp.SendCardData) {
				IsOK := false
				for i := 0; i < static.MAX_INDEX; i++ {
					if _userItem.Ctx.CardIndex[i] != 0 {
						card := sp.m_GameLogic.SwitchToCardData(byte(i))
						if sp.IsOthersTingHuCard(wChairID, card) {
							continue
						}
						_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, card, false)
						if !sp.OnUserOutCard(_msg, false) {
							sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
						}
						IsOK = true
						return
					}
				}
				if !IsOK {
					cbSendCardData := sp.SendCardData
					index := sp.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
					if index >= 0 && index < static.MAX_INDEX {
						if 0 != _userItem.Ctx.CardIndex[index] {
							_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData, false)
							if !sp.OnUserOutCard(_msg, false) {
								sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
							}
							return
						}
					}
				}
			} else {
				cbSendCardData := sp.SendCardData
				index := sp.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
				if index >= 0 && index < static.MAX_INDEX {
					if 0 != _userItem.Ctx.CardIndex[index] {
						_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData, false)
						if !sp.OnUserOutCard(_msg, false) {
							sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
						}
						return
					}
				}
			}
		}
	}
}

func (sp *SportYCKWX) setDelayLimitedTime(delaytime int) {

	sp.LimitTime = time.Now().Unix() + int64(delaytime)
	sp.GameTimer.SetDelayLimitTimer(components2.GameTime_Delayer, delaytime)
}

func (sp *SportYCKWX) CheckPlayerWantCard(userItem *components2.Player) {
	if sp.m_GameLogic.IsValidCard(userItem.Ctx.WantCard) {
		defer userItem.Ctx.CleanWant()
		cardStr := sp.m_GameLogic.SwitchToCardNameByData(userItem.Ctx.WantCard, 1)
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("玩家有想要的牌: %s", cardStr))
		index := sp.FindWantCard(userItem)
		if index == static.INVALID_BYTE {
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("高权限玩家换牌失败，牌库不存在此牌: %s", cardStr))
			sp.SendGameNotificationMessage(userItem.GetChairID(), fmt.Sprintf("%q已经被摸完了。", cardStr))
		} else {
			next := sp.LeftCardCount - 1
			sp.RepertoryCard[next], sp.RepertoryCard[index] = sp.RepertoryCard[index], sp.RepertoryCard[next]
			sp.OnWriteGameRecord(userItem.GetChairID(), "高权限玩家换牌成功")
		}
	}
}

func (sp *SportYCKWX) FindWantCard(userItem *components2.Player) byte {
	index := static.INVALID_BYTE
	cardData := userItem.Ctx.WantCard
	if !sp.m_GameLogic.IsValidCard(cardData) {
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("不合法的WANT牌:%s。", sp.m_GameLogic.SwitchToCardNameByData(cardData, 1)))
		return index
	}
	return sp.FindCardIndexInRepertoryCards(cardData)
}

func (sp *SportYCKWX) CheckWantCard(cardData byte) {
	index := sp.FindCardIndexInRepertoryCards(cardData)
	if index == static.INVALID_BYTE {
		cardStr := sp.m_GameLogic.SwitchToCardNameByData(cardData, 1)
		for i := 0; i < meta2.MAX_PLAYER; i++ {
			userItem := sp.GetUserItemByChair(uint16(i))
			if userItem != nil {
				if userItem.Ctx.WantCard == cardData {
					sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("高权限牌 %q 被摸完了，要重选。", cardStr))
					sp.SendGameNotificationMessage(userItem.GetChairID(), fmt.Sprintf("%q被摸完了, 请重选。", cardStr))
				}
			}
		}
	}
}
