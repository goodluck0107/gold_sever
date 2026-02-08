package Hubei_JingZhou

import (
	"encoding/json"
	"fmt"
	constant "github.com/open-source/game/chess.git/pkg/consts"
	public "github.com/open-source/game/chess.git/pkg/static"
	syslog "github.com/open-source/game/chess.git/pkg/xlog"
	common "github.com/open-source/game/chess.git/services/sport/backboard"
	modules "github.com/open-source/game/chess.git/services/sport/components"
	"github.com/open-source/game/chess.git/services/sport/infrastructure"
	"github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	"github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	"github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	"github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"math/rand"
	"strconv"
	"time"
)

/*
chess麻将-嘉鱼硬巧
*/

// 好友房规则相关属性
type FriendRule_jz_hzg struct {
	DiFen            int    `json:"dff"`              // 底分
	HasPao           string `json:"haspao"`           //跑
	OnlyZimo         string `json:"onlyzimo"`         //自摸胡
	QiDuiJiaBei      string `json:"qiduijiabei"`      //七对加倍
	GangKaiJiaBei    string `json:"gangkaijiabei"`    //杠开加倍
	QingYiSeJiaBei   string `json:"qingyisejiabei"`   //清一色加倍
	WanNengPai       string `json:"wannengpai"`       //万能牌
	NineSecondRoom   string `json:"sc9"`              //九秒场
	IpLimit          string `json:"ip"`               //ipgps限制
	Sihongzhongkehu  int    `json:"sihongzhongkehu"`  //底分
	Mm               int    `json:"mm"`               //买马
	Overtime_trust   int    `json:"overtime_trust"`   //超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //超时解散
	Endready         string `json:"endready"`         //小结算是否自动准备
	Fleetime         int    `json:"fleetime"`         //客户端传来的 游戏开始前离线踢人时间
	LookonSupport    string `json:"LookonSupport"`    //本局游戏是否支持旁观

	//游戏开始后的离线解散时间和申请解散时间
	OfflineDismissTime int    `json:"overtime_offdiss"`   //游戏开始后离线解散的时间，传秒数，如果不解散那就把这个时间传大一点
	ApplyDismissTime   int    `json:"overtime_applydiss"` //申请解散后的自动解散时间，传秒数
	Gmgh               string `json:"gmgd"`               //各摸各胡  勾选了，4人剩余牌≤4，3人剩余牌≤3，2人剩余牌≤2的时候，打出的牌其他人不能碰杠
	Dissmiss           int    `json:"dissmiss"`
}

type SportJZHZG struct {
	// 游戏共用部分
	modules.Common
	// 游戏流程数据
	metadata.Metadata
	//游戏逻辑
	m_GameLogic              SportLogicJZHZG
	m_QiangGangOperateResult public.Msg_S_OperateResult
	m_GangBuCardCount        byte // 杠后摸牌的次数
	m_bQiangGangResultSend   bool // 抢杠的结果是不是发送了
}

// ! 设置游戏可胡牌类型
func (sph *SportJZHZG) HuTypeInit(_type *public.TagHuType) {
	_type.HAVE_SIXI_HU = false
	_type.HAVE_QUE_YISE_HU = false
	_type.HAVE_BANBAN_HU = false
	_type.HAVE_LIULIU_HU = false
	_type.HAVE_QING_YISE_HU = false
	_type.HAVE_FENG_YI_SE = false
	_type.HAVE_HAO_HUA_DUI_HU = false
	_type.HAVE_JIANG_JIANG_HU = false
	_type.HAVE_FENG_YISE_HU = false
	_type.HAVE_QUAN_QIU_REN = false
	_type.HAVE_PENG_PENG_HU = false
	_type.HAVE_HAI_DI_HU = false
	_type.HAVE_QIANG_GANG_HU = true
	_type.HAVE_GANG_SHANG_KAI_HUA = false
	_type.HAVE_MENG_QING = false
	_type.HAVE_DI_HU = false
	_type.HAVE_TIAN_HU = false
	_type.HAVE_ZIMO_JIAO_1 = false
	_type.HAVE_QIDUI_HU = false
}

// ! 获取游戏配置
func (sph *SportJZHZG) GetGameConfig() *public.GameConfig { //获取游戏相关配置
	return &sph.Config
}

// ! 重置桌子数据
func (sph *SportJZHZG) RepositTable() {
	rand.Seed(time.Now().UnixNano())
	for _, v := range sph.PlayerInfo {
		v.Reset()
		v.Ctx.UserPaoReady = false
	}
	//游戏变量
	sph.SiceCount = modules.MAKEWORD(byte(1), byte(1))

	//出牌信息
	sph.OutCardData = 0
	sph.OutCardCount = 0
	sph.OutCardUser = public.INVALID_CHAIR

	//发牌信息
	sph.SendCardData = 0
	sph.SendCardCount = 0
	sph.LeftBu = 0

	//运行变量
	sph.ProvideCard = 0
	sph.ResumeUser = public.INVALID_CHAIR
	sph.CurrentUser = public.INVALID_CHAIR
	sph.ProvideUser = public.INVALID_CHAIR
	sph.PiZiCard = 0x00

	//状态变量
	sph.GangFlower = false
	sph.SendStatus = false
	sph.HaveHuangZhuang = false

	for k, _ := range sph.RepertoryCard {
		sph.RepertoryCard[k] = 0
	}

	sph.FanScore = [4]metadata.Game_mj_fan_score{}

	for _, v := range sph.PlayerInfo {
		v.Reset()
	}
	//结束信息
	sph.ChiHuCard = 0

	sph.ReplayRecord.Reset()

	sph.BeFirstGang = true
	sph.ResetTempGangOperateResult()
}

// ! 解析配置的任务
func (sph *SportJZHZG) ParseRule(strRule string) {

	syslog.Logger().Debug("parserRule :" + strRule)
	fmt.Println(fmt.Sprintf("消息（%v）", strRule))
	sph.Rule.CreateType = 0
	sph.Rule.NineSecondRoom = false

	sph.Rule.FangZhuID = sph.GetTableInfo().Creator
	sph.Rule.JuShu = sph.GetTableInfo().Config.RoundNum
	sph.Rule.CreateType = sph.FriendInfo.CreateType

	sph.Rule.DiFen = 1 //底分
	sph.Rule.NoWan = true
	sph.Config.PlayerCount = uint16(sph.GetTableInfo().Config.MaxPlayerNum)
	sph.Config.ChairCount = uint16(sph.GetTableInfo().Config.MaxPlayerNum)

	if len(strRule) == 0 {
		return
	}

	var _msg FriendRule_jz_hzg
	if err := json.Unmarshal(public.HF_Atobytes(strRule), &_msg); err == nil {
		sph.Rule.HasPao = _msg.HasPao == "true"
		sph.Rule.NineSecondRoom = _msg.NineSecondRoom == "true"
		sph.Rule.QiDuiJiaBei = _msg.QiDuiJiaBei == "true"
		sph.Rule.SiHZkehu = _msg.Sihongzhongkehu
		sph.Rule.Bird = _msg.Mm
		sph.Rule.Overtime_trust = _msg.Overtime_trust
		sph.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		sph.Rule.DiFen = _msg.DiFen

		sph.Rule.Endready = _msg.Endready == "true"
		if _msg.LookonSupport == "" {
			sph.Config.LookonSupport = true
		} else {
			sph.Config.LookonSupport = _msg.LookonSupport == "true"
		}

		//各摸各胡
		sph.Rule.Gmgh = _msg.Gmgh == "true"
		//游戏开始后离线解散时间
		sph.Rule.OfflineDismissTime = 1800
		if _msg.OfflineDismissTime > 0 {
			sph.Rule.OfflineDismissTime = _msg.OfflineDismissTime
		}
		//申请解散后的解散时间
		sph.Rule.ApplyDismissTime = 120
		if _msg.ApplyDismissTime > 0 {
			sph.Rule.ApplyDismissTime = _msg.ApplyDismissTime
		}

		if sph.CurCompleteCount == 0 {
			if _msg.Fleetime > 0 {
				sph.SetOfflineRoomTime(_msg.Fleetime)
			} else {
				//小于0表示不解散 设置时间2小时，2小时已经很长了
				sph.SetOfflineRoomTime(7200)
			}
		} else {
			sph.SetOfflineRoomTime(sph.Rule.OfflineDismissTime)
		}
		sph.Rule.DissmissCount = _msg.Dissmiss
	}
	sph.Rule.OnlyZimo = true
	sph.Rule.WanNengPai = true
	cardclass := 0
	if sph.Rule.NoWan {
		cardclass |= common.CARDS_WITHOUT_WAN
	}
	if sph.Rule.DissmissCount != 0 {
		sph.SetDissmissCount(sph.Rule.DissmissCount)
	}
	//只有红中
	cardclass |= (common.CARDS_WITHOUT_FA | common.CARDS_WITHOUT_BAI)
	sph.m_GameLogic.SetCardClass(cardclass)
}

// ! 开局
func (sph *SportJZHZG) OnBegin() {
	syslog.Logger().Debug("onbegin")
	sph.RepositTable()

	sph.OnWriteGameRecord(public.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range sph.PlayerInfo {
		v.OnBegin()
	}

	// 第一局随机庄家
	rand_num := rand.Intn(1000)
	sph.BankerUser = uint16(rand_num % sph.GetPlayerCount())
	sph.ParseRule(sph.GetTableInfo().Config.GameConfig)
	sph.m_GameLogic.Rule = sph.Rule
	sph.m_GameLogic.HuType = sph.HuType
	sph.CurCompleteCount = 0
	sph.VecGameEnd = []public.Msg_S_GameEnd{}
	sph.VecGameDataAllP = [4][]public.CMD_S_StatusPlay{}

	// 记录游戏开始时间
	sph.Common.GameBeginTime = time.Now()

	sph.CurCompleteCount++
	sph.GetTable().SetBegin(true)

	sph.OnGameStart()
}

func (sph *SportJZHZG) OnGameStart() {
	if !sph.CanContinue() {
		return
	}
	// 有跑就发送跑的设置
	if sph.Rule.HasPao {
		sph.SendPaoSetting()
	} else {
		sph.StartNextGame()
	}
}

// ! 发送下跑对话框
func (sph *SportJZHZG) SendPaoSetting() {

	sph.GameEndStatus = public.GS_MJ_PLAY
	//设置状态
	sph.SetGameStatus(public.GS_MJ_PLAY)

	for _, v := range sph.PlayerInfo {
		v.Ctx.CleanWeaveItemArray()
		v.Ctx.InitCardIndex()
	}

	sph.PayPaoStatus = true //设置玩家选漂的状态
	var PaoSetting public.Msg_S_PaoSetting
	//向每个玩家发送数据
	for _, v := range sph.PlayerInfo {
		if v.Ctx.UserPaoReady == true {
			PaoSetting.PaoStatus = true
			PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
			PaoSetting.Always = v.Ctx.VecXiaPao.Status
			PaoSetting.BankerUser = sph.BankerUser
		} else {
			PaoSetting.PaoStatus = false
			PaoSetting.PaoCount = v.Ctx.VecXiaPao.Num
			PaoSetting.Always = v.Ctx.VecXiaPao.Status
			PaoSetting.BankerUser = sph.BankerUser
		}
		sph.SendPersonMsg(constant.MsgTypeGamePaoSetting, PaoSetting, v.Seat)
	}
	sph.SendTableLookonMsg(constant.MsgTypeGamePaoSetting, PaoSetting)
}

// ! 玩家选择跑
func (sph *SportJZHZG) OnUserClientXiaPao(msg *public.Msg_C_Xiapao) bool {
	nChiarID := sph.GetChairByUid(msg.Id)
	_userItem := sph.GetUserItemByChair(nChiarID)
	if _userItem == nil {
		return false
	}
	if nChiarID >= 0 && nChiarID < metadata.MAX_PLAYER {
		_userItem.Ctx.XiaPao(msg)

		sph.NotifyXiaPao(nChiarID)
		fmt.Println(fmt.Sprintf("玩家%d,选跑%d", nChiarID, msg.Num))
	}

	// 如果4个玩家都准备好了，自动开启下一局
	_beginCount := 0
	for _, v := range sph.PlayerInfo {
		if !v.Ctx.UserPaoReady {
			recordStr := fmt.Sprintf("椅子编号[%d] 还没有完成选跑", v.Seat)
			sph.OnWriteGameRecord(uint16(v.Seat), recordStr)
			break
		}
		_beginCount++
	}

	if _beginCount >= sph.GetPlayerCount() {
		sph.OnWriteGameRecord(uint16(nChiarID), "所有人都完成选跑了，开始游戏")
		sph.PayPaoStatus = false
		//游戏没有开始发牌
		if !sph.GameStartForXiapao {
			sph.StartNextGame()
		}
	}

	return true
}

// ! 广播玩家的状态和选漂的数目
func (sph *SportJZHZG) NotifyXiaPao(wChairID uint16) bool {
	var sXiaPiao public.Msg_S_Xiapao

	for _, v := range sph.PlayerInfo {
		if v.Seat != public.INVALID_CHAIR {
			sXiaPiao.Num[v.Seat] = v.Ctx.VecXiaPao.Num
			sXiaPiao.Always[v.Seat] = v.Ctx.VecXiaPao.Status
			sXiaPiao.Status[v.Seat] = v.Ctx.UserPaoReady
		}
	}

	//发送数据
	sph.SendTableMsg(constant.MsgTypeGameXiapao, sXiaPiao)
	sph.SendTableLookonMsg(constant.MsgTypeGameXiapao, sXiaPiao)
	//游戏记录
	if wChairID == wChairID {
		recordStr := fmt.Sprintf("发送跑数：%d， 是否默认 %t", sXiaPiao.Num[wChairID], sXiaPiao.Status[wChairID])
		sph.OnWriteGameRecord(wChairID, recordStr)

		sph.addReplayOrder(wChairID, eve.E_Pao, byte(sXiaPiao.Num[wChairID]))
	}
	return true
}

// ! 开始下一局游戏
func (sph *SportJZHZG) StartNextGame() {
	sph.OnStartNextGame()
	sph.LastOutCardUser = public.INVALID_CHAIR
	sph.LastSendCardUser = uint16(public.INVALID_CHAIR)
	sph.HaveHuangZhuang = false
	//组合扑克
	sph.MagicCard = 0x00

	sph.LeftCardCount = 0
	sph.RepertoryCard = []byte{}
	sph.IsTianDi = true
	//发送最新状态
	for i := 0; i < sph.GetPlayerCount(); i++ {
		sph.SendUserStatus(i, public.US_PLAY) //把状态发给其他人
	}

	sph.OnWriteGameRecord(public.INVALID_CHAIR, "开始StartNextGame......")
	sph.OnWriteGameRecord(public.INVALID_CHAIR, sph.GetTableInfo().Config.GameConfig)

	for _, v := range sph.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	sph.ParseRule(sph.GetTableInfo().Config.GameConfig)
	//竞技点
	sph.SetVitaminLowPauseTime(10)
	//离线解散
	sph.SetOfflineRoomTime(sph.Rule.OfflineDismissTime)
	//申请解散
	sph.SetDismissRoomTime(sph.Rule.ApplyDismissTime)
	//设置状态
	sph.SetGameStatus(public.GS_MJ_PLAY)

	//混乱扑克
	//_randTmp := time.Now().Unix() + int64(sph.GetTableId()+sph.KIND_ID*100+sph.GetSortId()*1000)
	//rand.Seed(_randTmp)
	//rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	//sph.SiceCount = modules.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	//sph.LeftCardCount = byte(len(sph.RepertoryCard))
	//sph.LeftBu = 10 //剩下的补牌数

	//这里在没有调用混乱扑克的函数时m_cbRepertoryCard中是空的，当它调用了这个函数之后
	//在这个函数中把固定的牌打乱后放到这个数组中，在放的同时不断增加数组m_cbRepertoryCard
	//的长度
	sph.LeftCardCount, sph.RepertoryCard = sph.m_GameLogic.RandCardData()
	//---------------------------
	//cardsIndex := common.CardsToCardIndex(sph.RepertoryCard)
	//common.Print_cards(cardsIndex[:])
	//-----------------------
	//分发扑克--即每一个人解析他的14张牌结果存放在m_cbCardIndex[i]中
	sph.CreateLeftCardArray(sph.GetPlayerCount(), int(sph.LeftCardCount), false)
	for _, v := range sph.PlayerInfo {
		if v.Seat != public.INVALID_CHAIR {
			sph.LeftCardCount -= (public.MAX_COUNT - 1)
			v.Ctx.SetCardIndex(&sph.Rule, sph.RepertoryCard[sph.LeftCardCount:], public.MAX_COUNT-1)
		}
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	newLeftcount, err := sph.InitDebugCards_ex("mahjongjzhzg_test", &sph.RepertoryCard, &sph.BankerUser)
	if err != nil {
		fmt.Println(fmt.Sprintf("%v", err))
		sph.OnWriteGameRecord(public.INVALID_CHAIR, err.Error())
	}
	//////////////读取配置文件设置牌型end////////////////////////////////////
	//发送扑克---这是发送给庄家的第十四张牌
	sph.SendCardCount++
	sph.LeftCardCount--
	sph.SendCardData = sph.RepertoryCard[sph.LeftCardCount]

	//if sph.Rule.WanNengPai {
	//	sph.PiZiCard = sph.RepertoryCard[0]
	//	sph.MagicCard = sph.PiZiCard
	//} else {
	sph.PiZiCard = public.INVALID_BYTE
	//固定红中是癞子
	sph.MagicCard = 0x35
	//}

	sph.m_GameLogic.SetMagicCard(sph.MagicCard)
	//sph.m_GameLogic.SetPiZiCard(sph.PiZiCard)
	//sph.m_GameLogic.SetPiZiCards(sph.PiZiCards)

	//写游戏日志
	sph.WriteGameRecord()

	_userItem := sph.GetUserItemByChair(sph.BankerUser)
	_userItem.Ctx.DispatchCard(sph.SendCardData)

	//设置变量
	sph.ProvideCard = 0
	sph.ProvideUser = public.INVALID_CHAIR
	sph.CurrentUser = sph.BankerUser //供应用户
	sph.LastSendCardUser = sph.BankerUser

	//动作分析,只分析庄家的牌，只有庄家初始才可以杠
	//杠牌判断
	var GangCardResult public.TagGangCardResult
	_userItem.Ctx.UserAction |= sph.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, nil, 0, &GangCardResult)

	_userItem.Ctx.UserAction |= sph.CheckHu(sph.BankerUser, sph.BankerUser, 0, false, false)
	if newLeftcount != 0 {
		sph.LeftCardCount = newLeftcount
	}
	//构造数据,发送开始信息
	var GameStart public.Msg_S_GameStart
	GameStart.SiceCount = sph.SiceCount
	GameStart.BankerUser = sph.BankerUser
	GameStart.CurrentUser = sph.CurrentUser
	sph.setlimitetime(_userItem)
	GameStart.Overtime = sph.LimitTime
	GameStart.MagicCard = sph.PiZiCard

	//记录癞子牌
	sph.ReplayRecord.PiziCard = sph.PiZiCard
	GameStart.CardLeft.MaxCount = sph.RepertoryCardArray.MaxCount
	GameStart.CardLeft.Seat = int(sph.RepertoryCardArray.Seat)
	GameStart.CardLeft.Kaikou = sph.RepertoryCardArray.Kaikou
	GameStart.LeftCardCount = sph.LeftCardCount
	//向每个玩家发送数据
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//设置变量
		GameStart.Whotrust[i] = _item.CheckTRUST()
		GameStart.UserAction = _item.Ctx.UserAction //把上面分析过的结果保存再发送到客户端
		_, GameStart.CardData = sph.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameStart.CardData)
		//记录玩家手上初始牌
		_, sph.ReplayRecord.RecordHandCard[i] = sph.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, sph.ReplayRecord.RecordHandCard[i])
		//记录玩家初始分
		UserItem := sph.GetUserItem(i)
		if UserItem != nil {
			//TODO 玩家分数设置
			sph.ReplayRecord.Score[i] = 0
			sph.ReplayRecord.UVitamin[UserItem.Info.Uid] = _item.UserScoreInfo.Vitamin
			if uint16(i) == sph.BankerUser {
				GameStart.SendCardData = sph.SendCardData //发给庄家的第一张牌
			} else {
				GameStart.SendCardData = public.INVALID_BYTE
			}
		}
		//发送数据
		sph.SendPersonMsg(constant.MsgTypeGameStart, GameStart, uint16(i))
	}

	sph.SendTableLookonMsg(constant.MsgTypeGameStart, GameStart)

	if _userItem.Ctx.UserAction != 0 {
		sph.ResumeUser = sph.CurrentUser
		sph.SendOperateNotify()
	}
}

// ! 得到某个用户开口的次数,吃，碰，明杠的次数
func (sph *SportJZHZG) GetUserOpenMouth(wChairID uint16) uint16 {
	_userItem := sph.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return uint16(0)
	}

	return uint16(_userItem.Ctx.WeaveItemCount - _userItem.Ctx.HidGang)
}

// ! 初始化游戏
func (sph *SportJZHZG) OnInit(table infrastructure.TableBase) {
	sph.KIND_ID = table.GetTableInfo().KindId
	sph.Config.StartMode = public.StartMode_FullReady
	sph.Config.PlayerCount = 4 //玩家人数
	sph.Config.ChairCount = 4  //椅子数量
	sph.PlayerInfo = make(map[int64]*modules.Player)
	sph.HuTypeInit(&sph.HuType) //设置可胡牌类型

	sph.RepositTable()
	sph.SetGameStartMode(public.StartMode_FullReady)
	sph.GameTable = table
	sph.Init()
	sph.Unmarsha(table.GetTableInfo().GameInfo)

	table.GetTableInfo().GameInfo = ""
	sph.ParseRule(sph.GetTableInfo().Config.GameConfig)
}

// ! 发送消息
func (sph *SportJZHZG) OnMsg(msg *infrastructure.TableMsg) bool {

	switch msg.Head {
	case constant.MsgTypeGameOutCard: //! 出牌消息
		var _msg public.Msg_C_OutCard
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return sph.OnUserOutCard(&_msg)
		}
	case constant.MsgTypeGameOperateCard: //操作消息
		var _msg public.Msg_C_OperateCard
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return sph.OnUserOperateCard(&_msg)
		}
	case constant.MsgTypeGameGoOnNextGame: //下一局
		var _msg public.Msg_C_GoOnNextGame
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			sph.OnUserClientNextGame(&_msg)
		}
	case constant.MsgTypeGameXiapao: //选漂
		var _msg public.Msg_C_Xiapao
		if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
			sph.OnUserClientXiaPao(&_msg)
		}
	case constant.MsgTypeGameTrustee: // 托管
		{
			var _msg public.Msg_S_DG_Trustee
			if err := json.Unmarshal(public.HF_Atobytes(msg.Data), &_msg); err == nil {
				sph.OnUserTustee(&_msg)
			}
		}
	default:
		//sph.Common.OnMsg(msg)
	}
	return true
}

// ! 下一局
func (sph *SportJZHZG) OnUserClientNextGame(msg *public.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(sph.CurCompleteCount) >= sph.Rule.JuShu || sph.GetGameStatus() != public.GS_MJ_END {
		return true
	}

	// 记录游戏开始时间
	sph.Common.GameBeginTime = time.Now()

	nChiarID := sph.GetChairByUid(msg.Id)

	sph.SendTableMsg(constant.MsgTypeGameGoOnNextGame, *msg)
	sph.SendTableLookonMsg(constant.MsgTypeGameGoOnNextGame, *msg)
	if nChiarID >= 0 && nChiarID < uint16(sph.GetPlayerCount()) {
		_item := sph.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}
	sph.SendUserStatus(int(nChiarID), public.US_READY) //把我的状态发给其他人

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < sph.GetPlayerCount(); i++ {
		item := sph.GetUserItemByChair(uint16(i))
		if item != nil {
			//fmt.Println(fmt.Sprintf("玩家(%d)状态；准备（%t），托管（%t）",item.Seat,item.UserReady,item.CheckTRUST()))
			if !item.UserReady {
				break
				//if!item.CheckTRUST(){
				//	break
				//}else{
				//	item.UserReady = true
				//}
			}
		}
		if i == sph.GetPlayerCount()-1 {
			// 复位桌子
			sph.RepositTable()
			sph.CurCompleteCount++
			sph.GetTable().SetBegin(true)
			sph.OnGameStart()
		}
	}
	return true
}

// ! 清除吃胡记录
func (sph *SportJZHZG) initChiHuResult() {
	for _, v := range sph.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// ! 清除单个玩家记录
func (sph *SportJZHZG) ClearChiHuResultByUser(wCurrUser uint16) {
	for _, v := range sph.PlayerInfo {
		if v.GetChairID() == wCurrUser {
			v.Ctx.InitChiHuResult()
			break
		}
	}
}

// ! 反向清除单个玩家记录
func (sph *SportJZHZG) ClearChiHuResultByUserReverse(wCurrUser uint16) {
	for _, v := range sph.PlayerInfo {
		if v.GetChairID() != wCurrUser {
			v.Ctx.InitChiHuResult()
		}
	}
}

// ! 用户操作牌
func (sph *SportJZHZG) OnUserOperateCard(msg *public.Msg_C_OperateCard) bool {

	wChairID := sph.GetChairByUid(msg.Id)

	if sph.GetGameStatus() != public.GS_MJ_PLAY {
		return true
	}

	//效验用户
	if (wChairID != sph.CurrentUser) && (sph.CurrentUser != public.INVALID_CHAIR) {
		return false
	}

	_userItem := sph.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	// 能胡牌没有胡需要过庄
	if (_userItem.Ctx.UserAction&public.WIK_CHI_HU) != 0 && msg.Code != public.WIK_CHI_HU {
		//别人大牌自己弃胡才算过庄
		if sph.CurrentUser == public.INVALID_CHAIR {
			_userItem.Ctx.NeedGuoZhuang = true
		}
	}

	if msg.Code != public.WIK_NULL {
		// 解锁用户超时操作
		//sph.UnLockTimeOut(wChairID)
	}

	//游戏记录
	if msg.Code == public.WIK_NULL {
		sph.entryTrust(msg.ByClient, _userItem)
		sph.Greate_OperateRecord(msg, _userItem)
	}
	// 能胡牌没有胡需要过庄
	if (_userItem.Ctx.UserAction32&public.WIK_CHI_HU) != 0 && msg.Code != public.WIK_CHI_HU && sph.CurrentUser == public.INVALID_CHAIR {
		//_userItem.Ctx.SetChiHuKind(_userItem.Ctx.ChiHuResult.ChiHuKind, sph.ProvideCard, sph.getWinScore(_userItem.Seat))
		_userItem.Ctx.VecChiHuCard = append(_userItem.Ctx.VecChiHuCard, sph.ProvideCard)
		sph.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("有胡不胡，加入弃胡,牌:%s", sph.m_GameLogic.SwitchToCardNameByData(sph.ProvideCard, 1)))
	}
	//能碰不碰加入过碰
	if (_userItem.Ctx.UserAction32&public.WIK_PENG) != 0 /*&& msg.Code&(public.WIK_PENG|public.WIK_GANG)==0*/ && sph.CurrentUser == public.INVALID_CHAIR {
		_userItem.Ctx.VecPengCard = append(_userItem.Ctx.VecPengCard, sph.ProvideCard)
		sph.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("能碰不碰，加入过碰,牌:%s", sph.m_GameLogic.SwitchToCardNameByData(sph.ProvideCard, 1)))
	}

	// 回放中记录牌权操作
	sph.addReplayOrder(wChairID, eve.E_HandleCardRight, msg.Code)

	//被动动作,被动操作没有红中杠，赖子杠,不分析抢杠
	if sph.CurrentUser == public.INVALID_CHAIR {
		sph.OnUserOperateInvalidChair(msg, _userItem)
		return true
	}

	//主动动作，杠的是红中，赖子，和暗杠，此种情况下蓄杠要考抢杠的操作
	if sph.CurrentUser == wChairID {
		sph.OnUserOperateByChair(msg, _userItem)
		return true
	}

	return false
}

// ! 被动动作，别人打牌吃碰杠胡牌
func (sph *SportJZHZG) OnUserOperateInvalidChair(msg *public.Msg_C_OperateCard, _userItem *modules.Player) bool {

	wTargetUser := sph.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card

	//效验状态
	if _userItem.Ctx.Response {
		return false
	}
	if (cbOperateCode != public.WIK_NULL) && ((_userItem.Ctx.UserAction & cbOperateCode) == 0) {
		return false
	}
	if cbOperateCard != sph.ProvideCard {
		return false
	}

	//变量定义
	cbTargetAction := cbOperateCode
	//构造结果
	var OperateResult public.Msg_S_OperateResult
	//设置变量
	_userItem.Ctx.SetOperate(cbOperateCard, cbOperateCode)
	if cbOperateCard == 0 {
		_userItem.Ctx.SetOperateCard(sph.ProvideCard)
	}

	//执行判断
	for _, v := range sph.PlayerInfo {
		//获取动作
		cbUserAction := v.Ctx.UserAction

		if v.Ctx.Response {
			cbUserAction = v.Ctx.PerformAction
		}

		//优先级别
		cbUserActionRank := sph.m_GameLogic.GetUserActionRank(cbUserAction) // 动作等级
		cbTargetActionRank := sph.m_GameLogic.GetUserActionRank(cbTargetAction)

		//动作判断
		if cbUserActionRank > cbTargetActionRank {
			wTargetUser = v.Seat
			cbTargetAction = cbUserAction
		}
	}

	// 最大操作权限的人还没有操作则返回
	if _userItem = sph.GetUserItemByChair(wTargetUser); _userItem != nil && !_userItem.Ctx.Response {
		return true
	}

	//变量定义
	cbTargetCard := _userItem.Ctx.OperateCard
	//出牌变量
	sph.SendStatus = true
	if cbTargetAction != public.WIK_NULL {
		sph.OutCardData = 0
		sph.OutCardUser = public.INVALID_CHAIR

		if providItem := sph.GetUserItemByChair(sph.ProvideUser); providItem != nil {
			providItem.Ctx.Requiredcard(cbTargetCard)
		}
	}

	if cbTargetAction == public.WIK_NULL {
		//用户状态
		for _, v := range sph.PlayerInfo {
			v.Ctx.ClearOperateCard()
		}
		//还原蓄杠
		if !sph.m_bQiangGangResultSend {
			if _userItem := sph.GetUserItemByChair(sph.m_QiangGangOperateResult.OperateUser); _userItem != nil {
				for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
					cbWeaveKind := _userItem.Ctx.WeaveItemArray[i].WeaveKind
					cbCenterCard := _userItem.Ctx.WeaveItemArray[i].CenterCard
					if (cbCenterCard == cbOperateCard) && (cbWeaveKind == public.WIK_GANG) {
						_userItem.Ctx.XuGangAction()
						xuGangScore := sph.GetScoreOnGang(eve.E_Gang_XuGand) * sph.Rule.DiFen
						//provider := _userItem.Ctx.WeaveItemArray[i].ProvideUser1
						//sph.m_QiangGangOperateResult.ScoreOffset[provider] -= xuGangScore
						//sph.m_QiangGangOperateResult.ScoreOffset[sph.m_QiangGangOperateResult.ProvideUser] += xuGangScore
						//sph.OnUserScoreOffset(sph.m_QiangGangOperateResult.ProvideUser, xuGangScore)
						//sph.OnUserScoreOffset(provider, -xuGangScore)
						for i := 0; i < sph.GetPlayerCount(); i++ {
							_item := sph.GetUserItemByChair(uint16(i))
							if _userItem.Seat != _item.GetChairID() {
								sph.m_QiangGangOperateResult.ScoreOffset[_item.GetChairID()] -= xuGangScore
								sph.m_QiangGangOperateResult.ScoreOffset[_userItem.Seat] += xuGangScore
								sph.OnUserScoreOffset(_userItem.Seat, xuGangScore)
								sph.OnUserScoreOffset(_item.GetChairID(), -xuGangScore)
							}
						}
						//fmt.Println(fmt.Sprintf("弃蓄杠（%v）(%v)(%v)",sph.m_QiangGangOperateResult.GameScore,sph.m_QiangGangOperateResult.GameVitamin,sph.m_QiangGangOperateResult.ScoreOffset))
						sph.m_QiangGangOperateResult.GameScore, sph.m_QiangGangOperateResult.GameVitamin =
							sph.OnSettle(sph.m_QiangGangOperateResult.ScoreOffset, constant.EventSettleGaming)
						break
					}
				}
			}
			sph.m_bQiangGangResultSend = true
			sph.SendTableMsg(constant.MsgTypeGameOperateScore, sph.m_QiangGangOperateResult)
			sph.SendTableLookonMsg(constant.MsgTypeGameOperateScore, sph.m_QiangGangOperateResult)
			//sph.SendTableMsg(constant.MsgTypeGameOperateResult, sph.m_QiangGangOperateResult)
		}
		//放弃操作
		if _userItem = sph.GetUserItemByChair(sph.ResumeUser); _userItem != nil && _userItem.Ctx.PerformAction != public.WIK_NULL {
			wTargetUser = sph.ResumeUser
			cbTargetAction = _userItem.Ctx.PerformAction
		} else {
			if sph.LeftCardCount > 0 {
				_targetUserItem := sph.GetUserItemByChair(wTargetUser)
				if (_targetUserItem.Ctx.ChiHuResult.ChiHuKind & public.CHK_QIANG_GANG) != 0 {
					sph.DispatchCardData(sph.ResumeUser, true)
				} else {
					sph.DispatchCardData(sph.ResumeUser, false)
				}
			} else {
				sph.HaveHuangZhuang = true
				sph.ChiHuCard = 0
				sph.ProvideUser = public.INVALID_CHAIR
				sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)
			}
			return true
		}
	} else if cbTargetAction == public.WIK_CHI_HU {
		//胡牌操作
		for tempIndex := 0; tempIndex < sph.GetPlayerCount(); tempIndex++ {
			wUser := uint16(sph.GetNextSeat(sph.ProvideUser + uint16(tempIndex)))

			if _item := sph.GetUserItemByChair(wUser); _item != nil {
				//找到的第一个离放炮的用户最近并且有胡牌操作的用户
				if _item.Ctx.UserAction&public.WIK_CHI_HU != 0 {
					wTargetUser = wUser
					_userItem = _item
					if _userItem.Ctx.OperateCard == 0 {
						_userItem.Ctx.SetOperateCard(sph.ProvideCard)
					}
					break
				}
			}
		}

		//结束信息
		sph.ChiHuCard = cbTargetCard
		sph.ProvideUser = sph.ProvideUser

		//插入扑克
		if _userItem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
			_userItem.Ctx.DispatchCard(sph.ChiHuCard)
		}

		//清除别人胡牌的牌权
		sph.ClearChiHuResultByUserReverse(_userItem.GetChairID())

		//游戏记录
		recordStr := fmt.Sprintf("%s，胡牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sph.OnWriteGameRecord(wTargetUser, recordStr)

		//记录胡牌
		sph.addReplayOrder(wTargetUser, eve.E_Hu, cbTargetCard)

		sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)

		return true
	} else {
		//用户状态
		for _, v := range sph.PlayerInfo {
			v.Ctx.ClearOperateCard()
		}

		//组合扑克
		wIndex := int(_userItem.Ctx.WeaveItemCount)
		_userItem.Ctx.WeaveItemCount++
		_provideUser := sph.ProvideUser
		if sph.ProvideUser == public.INVALID_CHAIR {
			_provideUser = wTargetUser
		}
		_userItem.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, cbTargetAction, cbTargetCard)

		//删除扑克
		switch cbTargetAction {
		//case public.WIK_LEFT: //左吃操作
		//	sph.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		//case public.WIK_RIGHT: //中吃操作
		//	sph.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		//case public.WIK_CENTER: //右吃操作
		//	sph.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_PENG: //碰牌操作
			sph.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case public.WIK_GANG: //杠牌操作
			{
				//删除扑克
				_userItem.Ctx.ShowGangAction()
				//点杠
				//sph.GetUserItemByChair(_provideUser).Ctx.DianGangAction()

				cbRemoveCard := []byte{cbTargetCard, cbTargetCard, cbTargetCard}
				_userItem.Ctx.RemoveCards(&sph.Rule, cbRemoveCard)
				mingGangScore := sph.GetScoreOnGang(eve.E_Gang) * sph.Rule.DiFen
				OperateResult.ScoreOffset[_provideUser] -= mingGangScore
				OperateResult.ScoreOffset[_userItem.GetChairID()] += mingGangScore

				sph.OnUserScoreOffset(_provideUser, -mingGangScore)
				sph.OnUserScoreOffset(_userItem.GetChairID(), mingGangScore)

				//游戏记录
				recordStr := fmt.Sprintf("%s，杠牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
				sph.OnWriteGameRecord(wTargetUser, recordStr)

				//记录杠牌
				sph.addReplayOrder(wTargetUser, eve.E_Gang, cbTargetCard)

				sph.GangFlower = true
			}
		}

		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = cbTargetCard
		OperateResult.OperateCode = cbTargetAction
		OperateResult.ProvideUser = sph.ProvideUser
		sph.setlimitetime(_userItem)
		OperateResult.Overtime = sph.LimitTime
		if sph.ProvideUser == public.INVALID_CHAIR {
			OperateResult.ProvideUser = wTargetUser
		}
		for i := 0; i < sph.GetPlayerCount(); i++ {
			if !sph.Rule.NoGangScore {
				// 最新数据发送给客户端
				_user := sph.GetUserItemByChair(uint16(i))
				if _user == nil {
					syslog.Logger().Debug("空指针...杠牌分数同步")
					continue
				}
				// 记录分数
				_user.Ctx.GangScore += OperateResult.ScoreOffset[i]
			}
		}

		//操作次数记录
		if sph.ProvideUser != public.INVALID_CHAIR {
			//有人点炮的情况下,增加操作用户的操作次数,并保存第三次供牌的用户
			_userItem.Ctx.AddThirdOperate(sph.ProvideUser)
		}

		OperateResult.HaveGang[wTargetUser] = _userItem.Ctx.HaveGang

		if sph.LastOutCardUser == OperateResult.ProvideUser {
			sph.LastOutCardUser = public.INVALID_CHAIR
		}

		OperateResult.GameScore, OperateResult.GameVitamin = sph.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)
		//发送消息
		//fmt.Println(fmt.Sprintf("明杠GameScore（%v）GameVitamin(%v)ScoreOffset(%v)",OperateResult.GameScore,OperateResult.GameVitamin,OperateResult.ScoreOffset))
		sph.m_bQiangGangResultSend = true
		sph.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
		sph.SendTableLookonMsg(constant.MsgTypeGameOperateResult, OperateResult)
		//设置用户
		sph.CurrentUser = wTargetUser
		sph.ProvideCard = 0
		sph.ProvideUser = public.INVALID_CHAIR
		sph.SendCardData = public.INVALID_BYTE

		//最大操作用户操作的是杠牌，进行杠牌处理
		if cbTargetAction == public.WIK_GANG {
			//没有人能抢杠
			if sph.LeftCardCount > 0 {
				sph.DispatchCardData(wTargetUser, true)
			} else {
				sph.HaveHuangZhuang = true
				sph.ChiHuCard = 0
				sph.ProvideUser = public.INVALID_CHAIR
				sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)

			}
			return true
		}

		//如果是吃碰操作，再判断目标用户是否还有杠牌动作动作判断
		if sph.LeftCardCount > 0 {
			//杠牌判断
			var GangCardResult public.TagGangCardResult

			_item := sph.GetUserItemByChair(sph.CurrentUser)

			_item.Ctx.UserAction |= sph.m_GameLogic.AnalyseGangCard(_item, _item.Ctx.CardIndex,
				_item.Ctx.WeaveItemArray[:], _item.Ctx.WeaveItemCount, &GangCardResult)

			//结果处理
			if GangCardResult.CardCount > 0 {
				//设置变量
				_item.Ctx.UserAction |= public.WIK_GANG
				sph.ProvideCard = 0

				//发送动作
				sph.SendOperateNotify()
			}
		}
		return true
	}
	return true
}

// ! 主动动作，自己暗杠痞子杠赖子杠续杠胡牌
func (sph *SportJZHZG) OnUserOperateByChair(msg *public.Msg_C_OperateCard, _userItem *modules.Player) bool {
	wChairID := sph.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card

	//扑克效验
	if (cbOperateCode != public.WIK_NULL) && (cbOperateCode != public.WIK_CHI_HU) && (sph.m_GameLogic.IsValidCard(cbOperateCard) == false) {
		return false
	}

	//设置变量
	sph.SendStatus = true
	_userItem.Ctx.UserAction = public.WIK_NULL
	_userItem.Ctx.PerformAction = public.WIK_NULL
	var OperateResult public.Msg_S_OperateResult
	//执行动作
	//var class byte=0
	//if _userItem.CheckTRUST(){
	//	class=1
	//}
	switch cbOperateCode {
	case public.WIK_GANG: //杠牌操作
		{
			bAnGang := false
			//变量定义
			cbWeaveIndex := 0xFF
			cbWeaveProvideUser := public.INVALID_CHAIR
			cbCardIndex := sph.m_GameLogic.SwitchToCardIndex(cbOperateCard)
			if _userItem.Ctx.CardIndex[cbCardIndex] == 1 {
				//续杠
				for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
					cbWeaveKind := _userItem.Ctx.WeaveItemArray[i].WeaveKind
					cbCenterCard := _userItem.Ctx.WeaveItemArray[i].CenterCard
					if (cbCenterCard == cbOperateCard) && (cbWeaveKind == public.WIK_PENG) {
						cbWeaveIndex = int(i)
						cbWeaveProvideUser = _userItem.Ctx.WeaveItemArray[i].ProvideUser
						break
					}
				}
				//效验动作
				if cbWeaveIndex == 0xFF {
					return false
				}
				//游戏记录
				recordStr := fmt.Sprintf("%s，蓄杠牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sph.OnWriteGameRecord(wChairID, recordStr)
				//记录蓄杠牌
				sph.addReplayOrder(wChairID, eve.E_Gang_XuGand, cbOperateCard)
				//记录点续杠次数
				//sph.GetUserItemByChair(cbWeaveProvideUser).Ctx.DianXuGangAction()
				bAnGang = false
				//组合扑克
				_userItem.Ctx.AddWeaveItemArray_Modify(cbWeaveIndex, 1, cbWeaveProvideUser, cbOperateCode, cbOperateCard)
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
				_userItem.Ctx.AddWeaveItemArray_Modify(cbWeaveIndex, 0, wChairID, cbOperateCode, cbOperateCard)
				_userItem.Ctx.HidGangAction()
				bAnGang = true
				anGangScore := sph.GetScoreOnGang(eve.E_Gang_AnGang) * sph.Rule.DiFen
				for i := 0; i < sph.GetPlayerCount(); i++ {
					_item := sph.GetUserItemByChair(uint16(i))
					if wChairID != _item.GetChairID() {
						OperateResult.ScoreOffset[_item.GetChairID()] -= anGangScore
						OperateResult.ScoreOffset[wChairID] += anGangScore
						sph.OnUserScoreOffset(wChairID, anGangScore)
						sph.OnUserScoreOffset(_item.GetChairID(), -anGangScore)
					}
				}
				//游戏记录
				recordStr := fmt.Sprintf("%s，暗杠牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sph.OnWriteGameRecord(wChairID, recordStr)
				//记录暗杠牌
				sph.addReplayOrder(wChairID, eve.E_Gang_AnGang, cbOperateCard)
				//删除扑克
				_userItem.Ctx.CleanCard(cbCardIndex)
			}
			// 最新数据发送给客户端
			if !sph.Rule.NoGangScore {
				for i, l := 0, sph.GetPlayerCount(); i < l; i++ {
					_user := sph.GetUserItemByChair(uint16(i))
					if _user == nil {
						syslog.Logger().Debug("空指针...杠牌分数同步")
						continue
					}
					// 记录分数
					_user.Ctx.GangScore += OperateResult.ScoreOffset[i]
				}
			}
			//构造结果,向客户端发送操作结果
			//var OperateResult public.Msg_S_OperateResult
			OperateResult.OperateUser = wChairID
			OperateResult.ProvideUser = wChairID
			OperateResult.HaveGang[wChairID] = _userItem.Ctx.HaveGang //是否杠过
			OperateResult.OperateCode = cbOperateCode
			OperateResult.OperateCard = cbOperateCard
			OperateResult.GameScore, OperateResult.GameVitamin = sph.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)
			//发送消息
			fmt.Println(fmt.Sprintf("杠GameScore（%v）GameVitamin(%v)ScoreOffset(%v)", OperateResult.GameScore, OperateResult.GameVitamin, OperateResult.ScoreOffset))
			sph.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
			sph.SendTableLookonMsg(constant.MsgTypeGameOperateResult, OperateResult)
			//本局杠牌次数统计
			if bAnGang {
				//sph.SendTableMsg(constant.MsgTypeGameOperateResult, OperateResult)
				if sph.LeftCardCount > 0 {
					sph.GangFlower = true
					sph.DispatchCardData(wChairID, true)
				} else {
					sph.HaveHuangZhuang = true
					sph.ChiHuCard = 0
					sph.ProvideUser = public.INVALID_CHAIR
					sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)
				}
			} else { //如果不是暗杠的情况，才分析这张牌其他用户是否可以抢杠,调用的是分析用户响应操作
				bAroseAction := false
				if cbOperateCard != sph.MagicCard {
					//允许抢杠胡的条件下才分析抢杠胡
					if sph.HuType.HAVE_QIANG_GANG_HU {
						//检查手上红中的个数，如果>=2不能抢杠
						bAroseAction = sph.EstimateUserRespond(wChairID, cbOperateCard, public.EstimatKind_GangCard)
						if bAroseAction {
							//把抢杠之前的这个杠操作保存下来 等待别人是不是弃抢杠了
							OperateResult.OperateUser = _userItem.Seat
							OperateResult.ProvideUser = cbWeaveProvideUser
							sph.SaveGangOperateResult(OperateResult)
						} else {
							for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
								cbWeaveKind := _userItem.Ctx.WeaveItemArray[i].WeaveKind
								cbCenterCard := _userItem.Ctx.WeaveItemArray[i].CenterCard
								if (cbCenterCard == cbOperateCard) && (cbWeaveKind == public.WIK_GANG) {
									_userItem.Ctx.XuGangAction()
									xuGangScore := sph.GetScoreOnGang(eve.E_Gang_XuGand) * sph.Rule.DiFen
									for i := 0; i < sph.GetPlayerCount(); i++ {
										_item := sph.GetUserItemByChair(uint16(i))
										if wChairID != _item.GetChairID() {
											OperateResult.ScoreOffset[_item.GetChairID()] -= xuGangScore
											OperateResult.ScoreOffset[wChairID] += xuGangScore
											sph.OnUserScoreOffset(wChairID, xuGangScore)
											sph.OnUserScoreOffset(_item.GetChairID(), -xuGangScore)
										}
									}
									OperateResult.GameScore, OperateResult.GameVitamin =
										sph.OnSettle(OperateResult.ScoreOffset, constant.EventSettleGaming)
									//fmt.Println(fmt.Sprintf("明杠GameScore（%v）GameVitamin(%v)ScoreOffset(%v)",OperateResult.GameScore,OperateResult.GameVitamin,OperateResult.ScoreOffset))
								}
							}
							//发送消息
							sph.SendTableMsg(constant.MsgTypeGameOperateScore, OperateResult)
							sph.SendTableLookonMsg(constant.MsgTypeGameOperateScore, OperateResult)
							if sph.LeftCardCount > 0 {
								sph.GangFlower = true
								sph.DispatchCardData(wChairID, true)
							} else {
								sph.HaveHuangZhuang = true
								sph.ChiHuCard = 0
								sph.ProvideUser = public.INVALID_CHAIR
								sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)
							}
						}
					}
				}
			}
			return true
		}
	case public.WIK_CHI_HU: //吃胡操作,主动状态下没有抢杠的说法，有自摸胡牌，杠上开花胡牌
		{
			//普通胡牌
			sph.ClearChiHuResultByUserReverse(_userItem.GetChairID())
			sph.ProvideCard = sph.SendCardData

			if sph.ProvideCard != 0 {
				sph.ProvideUser = wChairID
			}

			//游戏记录
			recordStr := fmt.Sprintf("%s，胡牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(sph.ProvideCard, 1))
			sph.OnWriteGameRecord(wChairID, recordStr)

			//记录胡牌
			sph.addReplayOrder(wChairID, eve.E_Hu, sph.ProvideCard)

			//结束信息
			sph.ChiHuCard = sph.ProvideCard

			//结束游戏
			sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)

			return true
		}
	}
	return true
}

// ! 操作牌
func (sph *SportJZHZG) operateCard(cbTargetAction byte, cbTargetCard byte, _userItem *modules.Player) {
	var cbRemoveCard []byte
	var wik_kind int

	//变量定义
	switch cbTargetAction {
	case public.WIK_LEFT: //上牌操作
		cbRemoveCard = []byte{cbTargetCard + 1, cbTargetCard + 2}
		wik_kind = eve.E_Wik_Left

		//游戏记录
		recordStr := fmt.Sprintf("%s，左吃牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sph.OnWriteGameRecord(_userItem.Seat, recordStr)

	case public.WIK_RIGHT:
		cbRemoveCard = []byte{cbTargetCard - 2, cbTargetCard - 1}
		wik_kind = eve.E_Wik_Right

		//游戏记录
		recordStr := fmt.Sprintf("%s，右吃牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sph.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_CENTER:
		cbRemoveCard = []byte{cbTargetCard - 1, cbTargetCard + 1}
		wik_kind = eve.E_Wik_Center

		//游戏记录
		recordStr := fmt.Sprintf("%s，中吃牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sph.OnWriteGameRecord(_userItem.Seat, recordStr)
	case public.WIK_PENG: //碰牌操作
		cbRemoveCard = []byte{cbTargetCard, cbTargetCard}
		wik_kind = eve.E_Peng

		//游戏记录
		recordStr := fmt.Sprintf("%s，碰牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sph.OnWriteGameRecord(_userItem.Seat, recordStr)
	default:
	}

	//吃碰过手可以胡牌,不需要过庄
	_userItem.Ctx.NeedGuoZhuang = false

	//记录左吃
	sph.addReplayOrder(_userItem.Seat, wik_kind, cbTargetCard)
	//删除扑克
	_userItem.Ctx.RemoveCards(&sph.Rule, cbRemoveCard)

}

// ! 用户出牌
func (sph *SportJZHZG) OnUserOutCard(msg *public.Msg_C_OutCard) bool {
	syslog.Logger().Debug("OnUserOutCard")
	//效验状态
	if sph.GetGameStatus() != public.GS_MJ_PLAY {
		return true
	}

	wChairID := sph.GetChairByUid(msg.Id)
	//效验参数
	if wChairID != sph.CurrentUser {
		return false
	}
	if sph.m_GameLogic.IsValidCard(msg.CardData) == false {
		return false
	}

	_userItem := sph.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	//不能出混子
	//if msg.CardData == sph.MagicCard {
	//	return false
	//}

	//出牌丢进弃牌区
	sph.entryTrust(msg.ByClient, _userItem)
	//出牌丢进弃牌区
	var class byte = 0
	if _userItem.CheckTRUST() {
		class = 1
	}
	_userItem.Ctx.Discard_ex(msg.CardData, class)
	// 解锁用户超时操作
	//sph.UnLockTimeOut(wChairID)

	//删除扑克
	if !_userItem.Ctx.OutCard(&sph.Rule, msg.CardData) {
		syslog.Logger().Debug("removecard failed")
		return false
	}
	//发送扑克
	if sph.IsTianDi && (uint16(sph.GetNextSeat(wChairID)) == sph.BankerUser) {
		sph.IsTianDi = false
	}
	//游戏记录
	recordStr := fmt.Sprintf("%s，打出：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1))
	sph.OnWriteGameRecord(wChairID, recordStr)

	//设置变量
	sph.SendStatus = true
	sph.GangFlower = false
	//出牌记录
	sph.OutCardCount++
	////////////////出的是一般的牌或赖子牌////////////////
	sph.OutCardUser = wChairID
	sph.LastOutCardUser = wChairID
	sph.OutCardData = msg.CardData

	//构造数据
	var OutCard public.Msg_S_OutCard
	OutCard.User = int(wChairID)
	OutCard.Data = msg.CardData
	OutCard.ByClient = msg.ByClient
	OutCard.Overtime = time.Now().Unix() + public.GAME_OPERATION_TIME
	//记录出牌
	if class == 1 {
		//赤壁记录下是不是托管出的牌
		sph.addReplayOrder(wChairID, eve.E_OutCard_TG, msg.CardData)
	} else {
		sph.addReplayOrder(wChairID, eve.E_OutCard, msg.CardData)
	}

	//发送消息
	sph.SendTableMsg(constant.MsgTypeGameOutCard, OutCard)
	sph.SendTableLookonMsg(constant.MsgTypeGameOutCard, OutCard)
	//用户切换
	sph.ProvideUser = wChairID
	sph.ProvideCard = msg.CardData
	sph.CurrentUser = uint16(sph.GetNextSeat(wChairID))

	bAroseAction := false
	//各摸各胡 最后几张摸了打出以后，其他人是不能要的
	leftCardSum := int(sph.LeftCardCount)
	if sph.Rule.Gmgh && (leftCardSum < sph.GetPlayerCount()) {
		//直接发牌 不判断其他人的响应
		syslog.Logger().Info(fmt.Sprintf("勾选了各摸各胡，当前剩余牌[%d]小于人数[%d]打出牌其他玩家不可响应", leftCardSum, sph.GetPlayerCount()))
	} else {
		//响应判断，如果用户出的是一般牌，判断其他用户是否需要该牌，EstimatKind_OutCard只是正常出牌判断
		//如果当前用户自己 出了牌，不能自己对自己进行分析吃，碰杠
		bAroseAction = sph.EstimateUserRespond(wChairID, msg.CardData, public.EstimatKind_OutCard)
	}

	//打了牌，别人没有反应 流局
	if bAroseAction == false {
		if sph.LeftCardCount > 0 {
			// 发牌
			sph.DispatchCardData(sph.CurrentUser, false)
		} else {
			// 游戏结束
			sph.HaveHuangZhuang = true
			sph.ChiHuCard = 0
			sph.ProvideUser = public.INVALID_CHAIR
			sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)
		}
	}
	return true
}

// ! 超时自动出牌
func (sph *SportJZHZG) OnAutoOperate(wChairID uint16, bBreakin bool) {

	if bBreakin == false {
		return
	}
	if sph.GetGameStatus() == public.GS_MJ_FREE {
		//sph.UnLockTimeOut(wChairID)
		return
	}

	if sph.GetGameStatus() != public.GS_MJ_PLAY {
		//sph.UnLockTimeOut(wChairID)
		return
	}
	_userItem := sph.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	////处理下跑
	//if sph.PayPaoStatus{
	//	for _, v := range sph.PlayerInfo {
	//		if v == nil {
	//			continue
	//		}
	//		if !v.Ctx.VecXiaPao.Status {
	//			_msg := sph.Greate_XiaPaomsg(v.Uid, false,v.Ctx.VecXiaPao.Status, v.Seat==sph.BankerUser)
	//			// syslog.Logger().Debug(fmt.Sprintf("自动操作（吃胡）玩家（%d）座位号（%d）可执行的操作（%d）放弃,当前用户（%d）消息：%v", _userItem.Uid, wChairID, _userItem.Ctx.UserAction, sph.m_wCurrentUser, _msg))
	//			if !sph.OnUserClientXiapao(_msg){
	//				sph.OnWriteGameRecord(v.Seat, "服务器自动选飘时，可能被客户端抢先了")
	//			}
	//		}
	//	}
	//	return
	//}

	//能胡 胡牌 吃胡
	if (_userItem.Ctx.UserAction&public.WIK_CHI_HU) != 0 && sph.CurrentUser != wChairID {
		_msg := sph.Greate_Operatemsg(_userItem.Uid, false, public.WIK_NULL, sph.ProvideCard)
		sph.OnUserOperateCard(_msg)
		return
	}
	//点杠 点碰 放弃
	if sph.CurrentUser == public.INVALID_CHAIR && _userItem.Ctx.UserAction != 0 {
		_msg := sph.Greate_Operatemsg(_userItem.Uid, false, public.WIK_NULL, sph.ProvideCard)
		sph.OnUserOperateCard(_msg)
		return
	}

	//暗杠 擦炮直接放弃出牌
	if sph.CurrentUser == wChairID {
		cbSendCardData := sph.SendCardData
		index := sph.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
		if index >= 0 && index < public.MAX_INDEX {
			if 0 != _userItem.Ctx.CardIndex[index] {
				_msg := sph.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
				sph.OnUserOutCard(_msg)
				return
			}
		}
		for i := byte(public.MAX_INDEX - 1); i > 0; i-- {
			if _userItem.Ctx.CardIndex[i] != 0 {
				cbSendCardData := sph.m_GameLogic.SwitchToCardData(i)
				_msg := sph.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
				sph.OnUserOutCard(_msg)
				return
			}
		}
	}
}

// ! 创建操作牌消息
func (sph *SportJZHZG) Greate_Operatemsg(Id int64, byClient bool, Code byte, Card byte) *public.Msg_C_OperateCard {
	_msg := new(public.Msg_C_OperateCard)
	_msg.Card = Card
	_msg.Code = Code
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建出牌消息
func (sph *SportJZHZG) Greate_OutCardmsg(Id int64, byClient bool, Card byte) *public.Msg_C_OutCard {
	_msg := new(public.Msg_C_OutCard)
	_msg.CardData = Card
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 派发扑克
func (sph *SportJZHZG) DispatchCardData(wCurrentUser uint16, bGangFlower bool) bool {
	if sph.IsPausing() {
		sph.CurrentUser = public.INVALID_CHAIR
		sph.SetSendCardOpt(public.TagSendCardInfo{
			CurrentUser: wCurrentUser,
			GangFlower:  bGangFlower,
		})
		return true
	}
	//状态效验
	if wCurrentUser == public.INVALID_CHAIR {
		return false
	}

	_userItem := sph.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return false
	}

	//吃碰过手可以胡牌,不需要过庄
	_userItem.Ctx.NeedGuoZhuang = false

	//剩余牌校验
	if sph.LeftCardCount <= 0 {
		return false
	}

	bEnjoinHu := true
	//发牌处理
	if sph.SendStatus == true {

		sph.SendCardCount++
		sph.LeftCardCount--
		//if sph.BeFirstGang && bGangFlower && sph.Rule.WanNengPai {
		//	sph.RepertoryCard[0] = sph.RepertoryCard[sph.LeftCardCount]
		//	sph.RepertoryCard[sph.LeftCardCount] = sph.MagicCard
		//	sph.BeFirstGang = false
		//}
		sph.SendCardData = sph.RepertoryCard[sph.LeftCardCount]
		_userItem.Ctx.DispatchCard(sph.SendCardData)
		sph.SetLeftCardArray()
		//游戏记录
		recordStr := fmt.Sprintf("牌型%s，发来：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(sph.SendCardData, 1))
		sph.OnWriteGameRecord(wCurrentUser, recordStr)

		//记录发牌
		sph.addReplayOrder(wCurrentUser, eve.E_SendCard, sph.SendCardData)

		//设置变量
		sph.ProvideUser = wCurrentUser
		sph.ProvideCard = sph.SendCardData
		//给用户发牌后，判断用户是否可以杠牌
		if sph.LeftCardCount > 0 {
			var GangCardResult public.TagGangCardResult
			_userItem.Ctx.UserAction |= sph.m_GameLogic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex,
				_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult)
		}

		// 判断是否胡牌
		sph.initChiHuResult()
		sph.CheckHu(wCurrentUser, wCurrentUser, 0, bGangFlower, false)
	}

	sph.GangFlower = false
	//设置变量
	sph.OutCardData = 0
	sph.CurrentUser = wCurrentUser
	sph.OutCardUser = public.INVALID_CHAIR

	//构造数据
	var SendCard public.Msg_S_SendCard
	SendCard.CurrentUser = wCurrentUser
	SendCard.ActionMask = _userItem.Ctx.UserAction
	SendCard.ChiHuKindMask = _userItem.Ctx.ChiHuResult.ChiHuKind
	SendCard.CardData = 0x00
	if sph.SendStatus {
		SendCard.CardData = sph.SendCardData
	}
	SendCard.IsGang = bGangFlower
	SendCard.IsHD = false
	SendCard.EnjoinHu = bEnjoinHu
	sph.setlimitetime(_userItem)
	SendCard.Overtime = sph.LimitTime
	SendCard.VecGangCard = public.HF_BytesToInts(sph.InvalidGangCards)

	sph.LastSendCardUser = wCurrentUser
	// 设置开始超时操作
	for _, v := range sph.PlayerInfo {
		if v.GetChairID() != wCurrentUser {
			SendCard.CardData = 0x00
		} else {
			SendCard.CardData = sph.SendCardData
		}
		sph.SendPersonMsg(constant.MsgTypeGameSendCard, SendCard, uint16(v.GetChairID()))
	}

	sph.SendTableLookonMsg(constant.MsgTypeGameSendCard, SendCard)
	//游戏记录
	recordStr := fmt.Sprintf("发送牌权：%x", _userItem.Ctx.UserAction)
	sph.OnWriteGameRecord(wCurrentUser, recordStr)

	// 回放记录中记录牌权显示
	if _userItem.Ctx.UserAction > 0 {
		sph.addReplayOrder(wCurrentUser, eve.E_SendCardRight, _userItem.Ctx.UserAction)
	}

	if sph.m_GameLogic.GetMagicCount(_userItem.Ctx.CardIndex[:]) == sph.m_GameLogic.GetCardCount(_userItem.Ctx.CardIndex[:]) {
		if (_userItem.Ctx.UserAction & public.WIK_CHI_HU) != 0 {
			sph.ChiHuCard = sph.SendCardData
			sph.ProvideUser = wCurrentUser

			//清除别人胡牌的牌权
			sph.ClearChiHuResultByUserReverse(_userItem.GetChairID())

			//游戏记录
			recordStr := fmt.Sprintf("%s，胡牌：%s", sph.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sph.m_GameLogic.SwitchToCardNameByData(sph.ChiHuCard, 1))
			sph.OnWriteGameRecord(wCurrentUser, recordStr)

			//记录胡牌
			sph.addReplayOrder(wCurrentUser, eve.E_Hu, sph.ChiHuCard)

			sph.OnEventGameEnd(sph.ProvideUser, public.GER_NORMAL)
		}
	}
	return true
}

// ! 响应判断
func (sph *SportJZHZG) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int) bool {
	//变量定义
	bAroseAction := false

	// 响应判断只需要判断出牌以及续杠
	if EstimatKind != public.EstimatKind_OutCard && EstimatKind != public.EstimatKind_GangCard {
		return bAroseAction
	}

	//用户状态
	for _, v := range sph.PlayerInfo {
		v.Ctx.ClearOperateCard()
	}

	//动作判断
	for i := 0; i < sph.GetPlayerCount(); i++ {
		//用户过滤
		if wCenterUser == uint16(i) {
			continue
		}

		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//出牌类型检验
		if EstimatKind == public.EstimatKind_OutCard {
			//吃碰判断
			if cbCenterCard != 0x35 {
				_item.Ctx.UserAction |= sph.m_GameLogic.EstimatePengCard(_item.Ctx.CardIndex, cbCenterCard)
				//红中不能杠
				if sph.LeftCardCount > 0 {
					//杠牌判断
					_item.Ctx.UserAction |= sph.m_GameLogic.EstimateGangCard(_item.Ctx.CardIndex, cbCenterCard)
				}
			}
			if _item.Ctx.UserAction&public.WIK_PENG != 0 && sph.CheckNeedGuo(_item, cbCenterCard, 1, 1) {
				_item.Ctx.UserAction32 ^= public.WIK_PENG
				sph.SendGameNotificationMessage(_item.GetChairID(), "过碰后不能再碰")
			}
		}

		bQiangGang := false
		if EstimatKind == public.EstimatKind_GangCard {
			if _item.Ctx.CardIndex[31] < 2 {
				bQiangGang = true
				//判断是否可以胡牌
				sph.CheckHu(uint16(i), wCenterUser, cbCenterCard, false, bQiangGang)
				if _item.Ctx.UserAction&public.WIK_CHI_HU != 0 && sph.CheckNeedGuo(_item, cbCenterCard, 0, 0) {
					_item.Ctx.UserAction32 ^= public.WIK_CHI_HU
					sph.SendGameNotificationMessage(_item.GetChairID(), "弃胡不能胡")
				}
			}
		}

		//结果判断
		if _item.Ctx.UserAction != public.WIK_NULL {
			bAroseAction = true
			sph.setlimitetime(_item)
		}
	}

	//结果处理，标志为真说明上一个用户所出的牌有人可以所以要发送一个提示信息的框
	if bAroseAction == true {
		//设置变量
		sph.ProvideUser = uint16(wCenterUser)
		sph.ProvideCard = cbCenterCard
		sph.ResumeUser = sph.CurrentUser
		sph.CurrentUser = public.INVALID_CHAIR

		//发送提示
		sph.SendOperateNotify()

		return true
	}

	return false
}

// ! 发送操作
func (sph *SportJZHZG) SendOperateNotify() bool {
	//发送提示
	for _, v := range sph.PlayerInfo {
		if v.Ctx.UserAction != public.WIK_NULL {
			//构造数据
			var OperateNotify public.Msg_S_OperateNotify
			OperateNotify.ResumeUser = sph.ResumeUser
			//抢暗杠时，复用此字段，表示轮到谁抢了
			OperateNotify.ActionCard = sph.ProvideCard
			OperateNotify.ActionMask = v.Ctx.UserAction
			OperateNotify.ChiHuKindMask = v.Ctx.ChiHuResult.ChiHuKind
			OperateNotify.EnjoinHu = false
			if v.Ctx.UserAction == 0 {
				OperateNotify.Overtime = sph.LimitTime
			} else {
				OperateNotify.Overtime = v.Ctx.CheckTimeOut
			}
			OperateNotify.VecGangCard = public.HF_BytesToInts(sph.InvalidGangCards)
			//发送数据
			//抢的牌权需要发送给所有玩家，因为其他玩家需要知道轮到谁抢暗杠了
			if v.Ctx.UserAction == public.WIK_QIANG {
				OperateNotify.ActionCard = byte(v.Seat)
				sph.SendTableMsg(constant.MsgTypeGameOperateNotify, OperateNotify)
				sph.SendTableLookonMsg(constant.MsgTypeGameOperateNotify, OperateNotify)
			} else {
				sph.SendPersonMsg(constant.MsgTypeGameOperateNotify, OperateNotify, v.Seat)
			}

			// 游戏记录
			recrodStr := fmt.Sprintf("发送牌权：%x", v.Ctx.UserAction)
			sph.OnWriteGameRecord(v.Seat, recrodStr)

			// 回放记录中记录牌权显示
			sph.addReplayOrder(v.Seat, eve.E_SendCardRight, v.Ctx.UserAction)
		}
	}

	return true
}

// ! 增加回放操作记录
func (sph *SportJZHZG) addReplayOrder(chairId uint16, operation int, card byte) {
	var order metadata.Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	sph.ReplayRecord.VecOrder = append(sph.ReplayRecord.VecOrder, order)
}

// ! 检查是否能胡
func (sph *SportJZHZG) CheckHu(wCurrentUser uint16, wProvideUser uint16, cbCurrentCard byte, bGangFlower bool, bQiangGang bool) byte {
	sph.ClearChiHuResultByUser(wCurrentUser)

	_userItem := sph.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return public.WIK_NULL
	}

	//牌型权位
	wChiHuRight := uint16(0)
	//杠开权限判断
	if bGangFlower {
		wChiHuRight |= public.CHR_GANG_SHANG_KAI_HUA
	}
	if bQiangGang {
		wChiHuRight |= public.CHR_QIANG_GANG
	}
	//20200326 苏大强 是不是要检查4红中
	check4HZ := false

	switch sph.Rule.SiHZkehu {
	case 1:
		if sph.IsTianDi && (wChiHuRight&(public.CHR_GANG_SHANG_KAI_HUA|public.CHR_QIANG_GANG) == 0) {
			check4HZ = true
		}
	case 2:
		check4HZ = true
	default:

	}

	//给用户发牌后，胡牌判断
	_userItem.Ctx.UserAction |= sph.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:],
		_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, cbCurrentCard, wChiHuRight, &_userItem.Ctx.ChiHuResult, check4HZ)

	if _userItem.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
		sph.CheckFen(wCurrentUser, wProvideUser, cbCurrentCard, &_userItem.Ctx.ChiHuResult)
	}

	//需要过庄才能胡
	if _userItem.Ctx.NeedGuoZhuang {
		if (_userItem.Ctx.UserAction & public.WIK_CHI_HU) != 0 {
			_userItem.Ctx.UserAction ^= public.WIK_CHI_HU

			sph.SendGameNotificationMessage(_userItem.GetChairID(), "过庄不能胡")
		}
		return public.WIK_NULL
	}

	return _userItem.Ctx.UserAction
}

// ! 检查是都达到胡牌番数
func (sph *SportJZHZG) CheckFen(wCurrentUser uint16, wProvideUser uint16, cbCurrentCard byte, ChiHuResult *public.TagChiHuResult) {
	_userItem := sph.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return
	}

	//游戏算番算分初始化
	sph.FanScore[wCurrentUser].Reset()
	HuScore := sph.Rule.DiFen
	//自摸都是2分
	if ChiHuResult.ChiHuKind&public.CHK_SI_LAIZI_NO_HUPAI != 0 {
		HuScore *= 10
	} else {
		if wCurrentUser == wProvideUser {
			HuScore *= 2
			//硬胡海要再乘以2
			if ChiHuResult.ChiHuKind&public.CHK_PING_HU_NOMAGIC != 0 {
				HuScore *= 2
			}
		}
	}

	//if (ChiHuResult.ChiHuKind&public.CHK_QING_YI_SE) != 0 && sph.Rule.QingYiSeJiaBei {
	//	HuScore *= 2
	//}
	//if (ChiHuResult.ChiHuKind&public.CHK_GANG_SHANG_KAI_HUA) != 0 && sph.Rule.GangKaiJiaBei {
	//	HuScore *= 2
	//}
	//if (ChiHuResult.ChiHuKind&public.CHK_7_DUI) != 0 && sph.Rule.QiDuiJiaBei {
	//	HuScore *= 2
	//}

	if wCurrentUser != wProvideUser {
		//到这应该就是抢杠胡。再判断一下
		if ChiHuResult.ChiHuKind&public.CHK_SI_LAIZI_NO_HUPAI != 0 {
			HuScore *= 10
		} else {
			if ChiHuResult.ChiHuKind&public.CHK_QIANG_GANG != 0 {
				HuScore *= 4
			}
		}
		sph.FanScore[wCurrentUser].Score[wProvideUser] -= HuScore
		sph.FanScore[wCurrentUser].Score[wCurrentUser] += HuScore

		////计算跑分 没跑分
		//_provideItem := sph.GetUserItemByChair(uint16(wProvideUser))
		//PaoScore := _provideItem.Ctx.VecXiaPao.Num + _userItem.Ctx.VecXiaPao.Num
		//
		//sph.FanScore[wCurrentUser].Score[wProvideUser] -= PaoScore
		//sph.FanScore[wCurrentUser].Score[wCurrentUser] += PaoScore
	} else {
		//自摸
		for i := 0; i < sph.GetPlayerCount(); i++ {
			if i == int(wCurrentUser) {
				continue
			}

			sph.FanScore[wCurrentUser].Score[i] -= HuScore
			sph.FanScore[wCurrentUser].Score[wCurrentUser] += HuScore

			////计算跑分
			//_otherItem := sph.GetUserItemByChair(uint16(i))
			//PaoScore := _otherItem.Ctx.VecXiaPao.Num + _userItem.Ctx.VecXiaPao.Num
			//
			//sph.FanScore[wCurrentUser].Score[i] -= PaoScore
			//sph.FanScore[wCurrentUser].Score[wCurrentUser] += PaoScore
		}
	}

}

// ! 单局结算
func (sph *SportJZHZG) OnGameOver(wChairID uint16, cbReason byte) bool {
	sph.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 游戏结束, 流局结束，统计积分
func (sph *SportJZHZG) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if sph.GetGameStatus() == public.GS_MJ_END && cbReason == public.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	// 清除超时检测
	for _, v := range sph.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}

	switch cbReason {
	case public.GER_NORMAL: //常规结束
		return sph.OnGameEndNormal(wChairID, cbReason)
	case public.GER_USER_LEFT: //用户强退
		return sph.OnGameEndUserLeft(wChairID, cbReason)
	case public.GER_DISMISS: //解散游戏
		return sph.OnGameEndDissmiss(wChairID, cbReason, false)
	case public.GER_GAME_ERROR: //解散游戏
		return sph.OnGameEndDissmiss(wChairID, cbReason, true)
	}
	return false
}

// ! 结束，结束游戏
func (sph *SportJZHZG) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	//设置游戏结束状态
	sph.SetGameStatus(public.GS_MJ_END)

	if wChairID == public.INVALID_CHAIR {
		sph.initChiHuResult()
	}

	//定义变量
	var GameEnd public.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sph.LastSendCardUser
	GameEnd.EndStatus = cbReason

	//设置承包用户
	GameEnd.Contractor = public.INVALID_CHAIR

	GameEnd.ProvideUser = wChairID
	GameEnd.ChiHuCard = sph.ChiHuCard
	GameEnd.ChiHuUserCount = 1
	GameEnd.KaiKou = sph.Rule.KouKou
	for i := 0; i < int(sph.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sph.RepertoryCard[i]
	}

	nWinner := public.INVALID_CHAIR
	//只有一个赢家，循环判断找出赢家
	if !sph.HaveHuangZhuang {
		for _, v := range sph.PlayerInfo {
			if v.Ctx.ChiHuResult.ChiHuKind != public.CHK_NULL {
				nWinner = v.Seat
				break
			}
		}
	}

	//胡牌玩家
	GameEnd.Winner = nWinner
	for i := 0; i < sph.GetPlayerCount(); i++ {
		if GameEnd.Winner == uint16(i) {
			GameEnd.WWinner[i] = true
		} else {
			GameEnd.WWinner[i] = false
		}
	}

	//计算各玩家开口次数，明杠，暗杠，红中，赖子
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_userItem := sph.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		//保存四家开口的次数（明杠，吃、碰）的次数
		GameEnd.OperateCount[i] = uint16(_userItem.Ctx.WeaveItemCount - _userItem.Ctx.HidGang)
		if sph.Rule.KouKou == false && GameEnd.OperateCount[i] > 1 {
			GameEnd.OperateCount[i] = 1
		}
		//保存四家点杠的次数
		GameEnd.DianGangCount[i] = uint16(_userItem.Ctx.DianGang)
		//保存四家续杠的次数
		GameEnd.XuGangCount[i] = uint16(_userItem.Ctx.XuGang)
		//保存四家明杠的次数
		GameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang)
		//保存四家暗杠的次数
		GameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		//保存四家红中杠的次数
		GameEnd.HongzhongGangCount[i] = uint16(_userItem.Ctx.HongZhongGang)
		//保存四家发财杠的次数
		GameEnd.FaCaiGangCount[i] = uint16(_userItem.Ctx.FaCaiGang)
		//保存四家赖子杠的次数
		GameEnd.MagicGangCount[i] = uint16(_userItem.Ctx.MagicCardGang)

	}

	_userItem := sph.GetUserItemByChair(nWinner)

	//记录玩家分数
	var GameUserScore [4]int

	if _userItem != nil {
		//获取买马牌数
		sum := sph.Rule.Bird
		if sum > 0 {
			if _userItem.Ctx.ChiHuResult.ChiHuKind&public.CHK_PING_HU_NOMAGIC == 0 {
				sum -= 2
			}
		}
		//fmt.Println(fmt.Sprintf("买马前（%v）",sph.FanScore[nWinner]))
		//20200326买马
		marecord := sph.GetBingoHorse(sum, nWinner, &GameEnd.CbBirdData_ex, &GameEnd.BingoBirdCount)
		manum := sph.Rule.DiFen * int(marecord) * 2
		//重新加分
		if manum != 0 {
			if nWinner == sph.ProvideUser {
				//自摸，其他玩家给分
				for i := 0; i < sph.GetPlayerCount(); i++ {
					if uint16(i) == nWinner {
						continue
					}
					sph.FanScore[nWinner].Score[i] -= manum
					sph.FanScore[nWinner].Score[nWinner] += manum
				}
			} else {
				//抢杠胡单家给
				sph.FanScore[nWinner].Score[sph.ProvideUser] -= manum
				sph.FanScore[nWinner].Score[nWinner] += manum
			}
		}
		//fmt.Println(fmt.Sprintf("（%d）买马后（%v）",manum,sph.FanScore[nWinner]))
		GameEnd.Contractor = public.INVALID_CHAIR
		GameEnd.ContractorType = logic.Contractor_NULL

		for i := 0; i < sph.GetPlayerCount(); i++ {
			GameUserScore[i] = sph.FanScore[nWinner].Score[i]
		}

		for i := 0; i < sph.GetPlayerCount(); i++ {
			//小结算信息
			GameEnd.StrEnd[i] = sph.GetGameEndStr(uint16(i), GameEnd.WWinner[i], _userItem.Ctx.ChiHuResult.ChiHuKind)
		}
	} else {
		// 慌庄所有人都是0分
		for i := 0; i < sph.GetPlayerCount(); i++ {
			GameUserScore[i] = 0
		}
	}

	if sph.ProvideUser != public.INVALID_CHAIR && _userItem != nil {
		if sph.ProvideUser != nWinner {
			//记录接炮次数
			_userItem.Ctx.HuByChi()
			//记录点炮次数
			sph.GetUserItemByChair(sph.ProvideUser).Ctx.ProvideCard()
		} else {
			//记录自摸次数
			_userItem.Ctx.HuBySelf()
		}

		//默认硬胡
		GameEnd.HardHu = 1
		GameEnd.ChiHuKind = _userItem.Ctx.ChiHuResult.ChiHuKind

		//按wKindMask列出的顺序，依次查找是哪个大胡，多个大胡时，按第一个查到的大胡显示
		for k := 0; k < len(SportJZHuKindMask); k++ {
			if (_userItem.Ctx.ChiHuResult.ChiHuKind & uint64(SportJZHuKindMask[k])) != 0 {
				switch SportJZHuKindMask[k] {
				case public.CHK_QING_YI_SE: //清一色
					GameEnd.BigHuKind = public.GameBigHuKindQYS
					break
				case public.CHK_GANG_SHANG_KAI_HUA: //杠上开花
					GameEnd.BigHuKind = public.GameBigHuKindGSK
					break
				case public.CHK_7_DUI: //七对
					GameEnd.BigHuKind = public.GameBigHuKind_7
					break
				case public.CHK_QIANG_GANG: //抢杠
					GameEnd.BigHuKind = public.GameBigHuKindQG
				default:
					break
				}
				break
			}
		}
	} else { //流局
		//流局处理
		GameEnd.ChiHuCard = 0
		GameEnd.ChiHuUserCount = 0
		sph.HaveHuangZhuang = true
		GameEnd.Winner = public.INVALID_CHAIR
	}

	GameEnd.IsQuit = false
	GameEnd.TheOrder = sph.CurCompleteCount

	//判断调整分
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		GameEnd.MaxFSCount[i] = 0
		GameEnd.GameScore[i] = GameUserScore[i]
		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount

		if sph.CurCompleteCount == 1 {
			_item.Ctx.SetMaxScore(GameEnd.GameScore[i] + GameEnd.GameAdjustScore[i])
		} else {
			if GameEnd.GameScore[i] > _item.Ctx.MaxScoreUserCount {
				_item.Ctx.SetMaxScore(GameEnd.GameScore[i] + GameEnd.GameAdjustScore[i])
			}
		}
		sph.ReplayRecord.Score[i] = GameEnd.GameScore[i]
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sph.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
	}

	GameEnd.UserScore, GameEnd.UserVitamin = sph.OnSettle(GameEnd.GameScore, constant.EventSettleGameOver)

	//发送信息
	sph.VecGameEnd = append(sph.VecGameEnd, GameEnd) //保存，用于汇总计算
	sph.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	sph.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	sph.SaveGameData()
	sph.OnWriteGameRecord(public.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	//荒庄
	if sph.HaveHuangZhuang {
		//记录荒庄
		sph.addReplayOrder(0, eve.E_HuangZhuang, 0)

		//记录胡牌类型
		sph.ReplayRecord.BigHuKind = 2
		sph.ReplayRecord.ProvideUser = 9
	} else {
		//记录胡牌类型
		sph.ReplayRecord.BigHuKind = GameEnd.BigHuKind
		sph.ReplayRecord.ProvideUser = GameEnd.ProvideUser
	}

	// 数据库写出牌记录
	sph.ReplayRecord.EndInfo = &GameEnd
	sph.TableWriteOutDate(int(sph.CurCompleteCount), sph.ReplayRecord)

	// 写完后清除数据
	sph.ReplayRecord.Reset()

	//数据库写分
	for _, v := range sph.PlayerInfo {
		wintype := public.ScoreKind_Draw
		if GameEnd.GameScore[v.Seat] > 0 {
			wintype = public.ScoreKind_Win
		} else {
			wintype = public.ScoreKind_Lost
		}
		//fmt.Println(fmt.Sprintf("（%d）写份(%d)(%d)",v.Seat,GameEnd.GameScore[v.Seat],GameEnd.GameAdjustScore[v.Seat]))
		sph.TableWriteGameDate(int(sph.CurCompleteCount), v, wintype, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
	}

	//扣房卡
	if sph.CurCompleteCount == 1 {
		sph.TableDeleteFangKa(sph.CurCompleteCount)
	}

	//结束游戏
	if int(sph.CurCompleteCount) >= sph.Rule.JuShu { //局数够了
		sph.CalculateResultTotal(public.GER_NORMAL, wChairID, 0) //计算总发送总结算

		sph.UpdateOtherFriendDate(&GameEnd, false)
		//通知框架结束游戏
		//sph.SetGameStatus(public.GS_MJ_FREE)
		sph.ConcludeGame()

	} else {
	}

	if nWinner != public.INVALID_CHAIR {
		sph.BankerUser = nWinner
	}

	sph.OnGameEnd()
	sph.accountManage(sph.Rule.Overtime_dismiss, sph.Rule.Overtime_trust, 10)
	sph.RepositTable() // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	return true

}

// ! 强退，结束游戏
func (sph *SportJZHZG) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	//变量定义
	var GameEnd public.Msg_S_GameEnd
	GameEnd.EndStatus = cbReason

	GameEnd.MagicCard = sph.MagicCard

	//设置变量
	GameEnd.ProvideUser = public.INVALID_CHAIR
	GameEnd.Winner = public.INVALID_CHAIR
	GameEnd.Contractor = public.INVALID_CHAIR
	GameEnd.IsQuit = true
	for i := 0; i < int(sph.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sph.RepertoryCard[i]
	}

	//抢杠分数，解散了也要结算
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sph.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])

		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		//玩家番数
		//GameEnd.MaxFSCount[i] = uint16(sph.GetPlayerFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount

		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore
	}

	//税收计算,按正常情况下进行计算,如果有调整,在后面进行调整
	var ScoreInfo [metadata.MAX_PLAYER]rule.TagScoreInfo

	for i := 0; i < sph.GetPlayerCount(); i++ {
		//记录各家的分数
		ScoreInfo[i].Score = GameEnd.GameScore[i]

		//如果分数为0，设置状态和局状态
		if GameEnd.GameScore[i] == 0 {
			ScoreInfo[i].ScoreKind = public.ScoreKind_Draw
		} else { //否则如果分数为正,设置为赢,为负设置为输,即使后面有人承包,也是输
			if GameEnd.GameScore[i] > 0 {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Win
			} else {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Lost
			}
		}
	}

	//游戏记录
	sph.OnWriteGameRecord(wChairID, "强退游戏结束")

	//记录异常结束数据
	sph.addReplayOrder(wChairID, eve.E_Li_Xian, 0)

	//记录胡牌类型
	sph.ReplayRecord.BigHuKind = GameEnd.BigHuKind
	sph.ReplayRecord.ProvideUser = GameEnd.ProvideUser
	//荒庄
	if sph.HaveHuangZhuang {
		//记录荒庄
		sph.addReplayOrder(0, eve.E_HuangZhuang, 0)

		sph.ReplayRecord.BigHuKind = 2
		sph.ReplayRecord.ProvideUser = 9
	}

	nextCardLen := len(GameEnd.NextCard)
	for i := 0; i < nextCardLen; i++ {
		GameEnd.NextCard[i] = 0x00
	}
	//发送几张牌出去
	if int(sph.LeftCardCount) >= nextCardLen {
		for i := 0; i < nextCardLen; i++ {
			sph.LeftCardCount--
			GameEnd.NextCard[i] = sph.RepertoryCard[sph.LeftCardCount]
		}
	}

	//发送信息
	sph.VecGameEnd = append(sph.VecGameEnd, GameEnd) //保存，用于汇总计算
	sph.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	sph.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	sph.SaveGameData()

	if sph.GetGameStatus() != public.GS_MJ_FREE {
		// 数据库写出牌记录
		sph.TableWriteOutDate(int(sph.CurCompleteCount), sph.ReplayRecord)
		// 写完后清除数据
		sph.ReplayRecord.Reset()

		//数据库写分
		for _, v := range sph.PlayerInfo {
			if v.Seat != public.INVALID_CHAIR {
				if wChairID != public.INVALID_CHAIR && v.Seat == wChairID {
					sph.TableWriteGameDate(int(sph.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
				} else {
					sph.TableWriteGameDate(int(sph.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
				}
			}
		}
	}

	sph.UpdateOtherFriendDate(&GameEnd, true)
	//结束游戏
	sph.CalculateResultTotal(public.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	//sph.SetGameStatus(public.GS_MJ_FREE)
	sph.ConcludeGame()

	return true
}

// ! 解散，结束游戏
func (sph *SportJZHZG) OnGameEndDissmiss(wChairID uint16, cbReason byte, err bool) bool {
	//变量定义
	var GameEnd public.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sph.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.ProvideUser = public.INVALID_CHAIR
	GameEnd.Winner = public.INVALID_CHAIR
	GameEnd.Contractor = public.INVALID_CHAIR
	for i := 0; i < int(sph.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sph.RepertoryCard[i]
	}

	GameEnd.MagicCard = sph.MagicCard

	//记录异常结束数据
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == public.US_OFFLINE && i != int(wChairID) {
			sph.addReplayOrder(uint16(i), eve.E_Li_Xian, 0)
		}
	}

	sph.addReplayOrder(wChairID, eve.E_Jie_san, 0)

	//抢杠分数，解散了也要结算
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		//玩家番数
		//GameEnd.MaxFSCount[i] = uint16(sph.GetPlayerFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.GameAdjustScore[i] = _item.Ctx.StorageScore

		if sph.Rule.HasPao {
			GameEnd.PaoCount[i] = uint16(_item.Ctx.VecXiaPao.Num)
		} else {
			GameEnd.PaoCount[i] = 0xFF
		}

		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sph.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
	}

	//税收计算,按正常情况下进行计算,如果有调整,在后面进行调整
	var ScoreInfo [metadata.MAX_PLAYER]rule.TagScoreInfo

	for i := 0; i < sph.GetPlayerCount(); i++ {
		//记录各家的分数
		ScoreInfo[i].Score = GameEnd.GameScore[i]

		//如果分数为0，设置状态和局状态
		if GameEnd.GameScore[i] == 0 {
			ScoreInfo[i].ScoreKind = public.ScoreKind_Draw
		} else { //否则如果分数为正,设置为赢,为负设置为输,即使后面有人承包,也是输
			if GameEnd.GameScore[i] > 0 {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Win
			} else {
				ScoreInfo[i].ScoreKind = public.ScoreKind_Lost
			}
		}
	}
	GameEnd.IsQuit = true

	nextCardLen := len(GameEnd.NextCard)
	for i := 0; i < nextCardLen; i++ {
		GameEnd.NextCard[i] = 0x00
	}
	//发送几张牌出去
	if int(sph.LeftCardCount) >= nextCardLen {
		for i := 0; i < nextCardLen; i++ {
			sph.LeftCardCount--
			GameEnd.NextCard[i] = sph.RepertoryCard[sph.LeftCardCount]
		}
	}

	//发送信息
	sph.VecGameEnd = append(sph.VecGameEnd, GameEnd) //保存，用于汇总计算
	sph.SendTableMsg(constant.MsgTypeGameEnd, GameEnd)
	sph.SendTableLookonMsg(constant.MsgTypeGameEnd, GameEnd)
	sph.SaveGameData()

	//游戏记录
	sph.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

	if !err {
		if sph.GetGameStatus() != public.GS_MJ_FREE && !sph.PayPaoStatus {
			//数据库写出牌记录
			sph.TableWriteOutDate(int(sph.CurCompleteCount), sph.ReplayRecord)
			// 写完后清除数据
			sph.ReplayRecord.Reset()

			//数据库写入单局结算
			for _, v := range sph.PlayerInfo {
				if v.Seat != public.INVALID_CHAIR {
					if wChairID != public.INVALID_CHAIR && v.Seat == wChairID {
						sph.TableWriteGameDate(int(sph.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
					} else {
						sph.TableWriteGameDate(int(sph.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat]+GameEnd.GameAdjustScore[v.Seat])
					}
				}
			}
		}
		// 写总计算
		sph.CalculateResultTotal(public.GER_DISMISS, wChairID, 0)
	} else {
		// 写总计算
		sph.CalculateResultTotal(public.GER_DISMISS, wChairID, 1)
	}

	sph.UpdateOtherFriendDate(&GameEnd, true)
	//结束游戏
	//sph.SetGameStatus(public.GS_MJ_FREE)
	sph.ConcludeGame()

	return true
}

// ! 解散牌桌
func (sph *SportJZHZG) OnEnd() {
	if sph.IsGameStarted() {
		sph.OnGameOver(public.INVALID_CHAIR, public.GER_DISMISS)
	}
}

// ! 计算总发送总结算
func (sph *SportJZHZG) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	// 给客户端发送总结算数据
	var balanceGame public.Msg_S_BALANCE_GAME
	balanceGame.Userid = sph.Rule.FangZhuID
	balanceGame.CurTotalCount = sph.CurCompleteCount //总盘数
	sph.TimeEnd = time.Now().Unix()                  //大局结束时间
	balanceGame.TimeStart = sph.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = sph.TimeEnd
	for i := 0; i < len(sph.VecGameEnd); i++ {
		for j := 0; j < sph.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += sph.VecGameEnd[i].GameScore[j] + sph.VecGameEnd[i].GameAdjustScore[j] //总分
		}
	}

	//打印日志
	recrodStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）", cbReason)
	sph.OnWriteGameRecord(wChairID, recrodStr)

	for i := 0; i < sph.GetPlayerCount(); i++ {
		sph.OnWriteGameRecord(uint16(i), "当前该玩家离线")
	}

	//非正常结束，记录总结算时玩家状态：0正常，1解散，2离线
	if public.GER_USER_LEFT == cbReason {
		for i := 0; i < sph.GetPlayerCount(); i++ {
			if i == int(wChairID) {
				balanceGame.UserEndState[i] = 2
			}
		}
	} else {
		if public.GER_DISMISS == cbReason {
			for i := 0; i < sph.GetPlayerCount(); i++ {
				if i == int(wChairID) {
					balanceGame.UserEndState[i] = 1
				} else {
					if _item := sph.GetUserItemByChair(uint16(i)); _item != nil && _item.UserStatus == public.US_OFFLINE {
						balanceGame.UserEndState[i] = 2
					}
				}
			}
		}
	}

	//! 计算大赢家的座位号
	bigWinScore := 0
	for i := 0; i < sph.GetPlayerCount(); i++ {
		if balanceGame.GameScore[i] >= bigWinScore {
			bigWinScore = balanceGame.GameScore[i]
		}
	}

	//开始就解散时不加，每局结束后游戏状态会复位为GS_MJ_FREE
	wintype := public.ScoreKind_Draw
	if sph.CurCompleteCount == 1 && sph.GetGameStatus() != public.GS_MJ_END {
		wintype = public.ScoreKind_pass
	} else {
		if sph.CurCompleteCount == 0 {
			wintype = public.ScoreKind_pass
		}
	}

	if cbSubReason == 0 {
		for i := 0; i < sph.GetPlayerCount(); i++ {
			_userItem := sph.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ZimoCount[i] = _userItem.Ctx.HuBySelfCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			balanceGame.FXMaxUserCount[i] = _userItem.Ctx.MaxFanUserCount
			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
			for j := 0; j < len(sph.VecGameEnd); j++ {
				balanceGame.ShowGangCount[i] += int(sph.VecGameEnd[j].ShowGangCount[i] + sph.VecGameEnd[j].XuGangCount[i])
				balanceGame.HidGangCount[i] += int(sph.VecGameEnd[j].HideGangCount[i])
			}

			if wintype != public.ScoreKind_pass {
				if balanceGame.GameScore[i] > 0 {
					wintype = public.ScoreKind_Win
				} else {
					wintype = public.ScoreKind_Lost
				}
			}

			isBigWin := 0
			if bigWinScore == balanceGame.GameScore[i] {
				isBigWin = 1
			}
			//写记录
			sph.TableWriteGameDateTotal(int(sph.CurCompleteCount), i, wintype, balanceGame.GameScore[i], isBigWin)
		}
	} else {
		sph.UpdateErrGameTotal(sph.GetTableInfo().GameNum)
	}

	// 记录用户好友房历史战绩
	if wintype != public.ScoreKind_pass {
		sph.TableWriteHistoryRecord(&balanceGame)
		sph.TableWriteHistoryRecordDetail(&balanceGame)
	}

	balanceGame.End = 0
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_userItem := sph.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}

		gameendStr := ""
		if len(sph.VecGameEnd) > 0 {
			gameendStr = public.HF_JtoA(sph.VecGameEnd[len(sph.VecGameEnd)-1])
		}
		gamedataStr := ""
		if len(sph.VecGameDataAllP[i]) > 0 {
			gamedataStr = public.HF_JtoA(sph.VecGameDataAllP[i][len(sph.VecGameDataAllP[i])-1])
		}
		sph.SaveLastGameinfo(_userItem.Uid, gameendStr, public.HF_JtoA(balanceGame), gamedataStr)
	}
	//发消息
	sph.SendTableMsg(constant.MsgTypeGameBalanceGame, balanceGame)
	sph.SendTableLookonMsg(constant.MsgTypeGameBalanceGame, balanceGame)
	sph.resetEndDate()
}

// ! 重置优秀结束数据
func (sph *SportJZHZG) resetEndDate() {
	sph.CurCompleteCount = 0
	sph.VecGameEnd = []public.Msg_S_GameEnd{}

	for _, v := range sph.PlayerInfo {
		v.OnEnd()
	}
}

func (sph *SportJZHZG) UpdateOtherFriendDate(GameEnd *public.Msg_S_GameEnd, bEnd bool) {

}

// ! 发送游戏开始场景数据
func (sph *SportJZHZG) sendGameSceneStatusPlay(player *modules.Player) bool {

	if player.LookonTableId > 0 {
		sph.sendGameSceneStatusPlayLookon(player)
		return true
	}

	wChiarID := player.GetChairID()

	if wChiarID == public.INVALID_CHAIR {
		syslog.Logger().Debug("sendGameSceneStatusPlay invalid chair")
		return false
	}

	//变量定义
	var StatusPlay public.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = sph.SiceCount
	StatusPlay.BankerUser = sph.BankerUser
	StatusPlay.CurrentUser = sph.CurrentUser
	StatusPlay.CellScore = sph.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sph.PiZiCard
	StatusPlay.KaiKou = sph.Rule.KouKou
	StatusPlay.RenWuAble = sph.RenWuAble
	StatusPlay.TheOrder = sph.CurCompleteCount
	StatusPlay.VecGangCard = public.HF_BytesToInts(sph.InvalidGangCards)
	if sph.CurrentUser == player.Seat {
		StatusPlay.SendCardData = sph.SendCardData
	} else {
		StatusPlay.SendCardData = public.INVALID_BYTE
	}

	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//玩家番数
		//StatusPlay.PlayerFan[i] = sph.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = sph.ProvideCard
	StatusPlay.LeftCardCount = sph.LeftCardCount
	StatusPlay.ActionMask = player.Ctx.UserAction
	StatusPlay.ChiHuKindMask = player.Ctx.ChiHuResult.ChiHuKind
	StatusPlay.CardLeft.CardArray = make([]int, sph.RepertoryCardArray.MaxCount, sph.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sph.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sph.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sph.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sph.RepertoryCardArray.Kaikou
	if player.Ctx.Response {
		StatusPlay.ActionMask = public.WIK_NULL
	}

	if player.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = player.Ctx.CheckTimeOut
	} else {
		CurUserItem := sph.GetUserItemByChair(sph.CurrentUser)
		if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
		}

		for _, v := range sph.PlayerInfo {
			if v.GetChairID() == player.GetChairID() {
				continue
			}
			if v.GetChairID() == sph.CurrentUser {
				continue
			}
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if sph.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sph.OutCardUser
	StatusPlay.OutCardData = sph.OutCardData
	StatusPlay.LastOutCardUser = sph.LastOutCardUser

	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardData = sph.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)

	//发送场景
	sph.SendPersonMsg(constant.MsgTypeGameStatusPlay, StatusPlay, wChiarID)
	//发小结消息
	if byte(len(sph.VecGameEnd)) == sph.CurCompleteCount && sph.CurCompleteCount != 0 && int(wChiarID) < sph.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := sph.VecGameEnd[sph.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		sph.SendPersonMsg(constant.MsgTypeGameEnd, gamend, wChiarID)
	}

	//发送解散房间所有玩家的反应
	sph.SendAllPlayerDissmissInfo(player)

	if sph.PayPaoStatus {
		sph.SendPaoSetting()
	}
	return true
}

func (sph *SportJZHZG) SaveGameData() {
	//变量定义
	var StatusPlay public.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = sph.SiceCount
	StatusPlay.BankerUser = sph.BankerUser
	StatusPlay.CurrentUser = sph.CurrentUser
	StatusPlay.CellScore = sph.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sph.PiZiCard
	StatusPlay.KaiKou = sph.Rule.KouKou
	StatusPlay.RenWuAble = sph.RenWuAble
	StatusPlay.TheOrder = sph.CurCompleteCount
	StatusPlay.VecGangCard = public.HF_BytesToInts(sph.InvalidGangCards)

	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//玩家番数
		//StatusPlay.PlayerFan[i] = sph.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = sph.ProvideCard
	StatusPlay.LeftCardCount = sph.LeftCardCount
	StatusPlay.CardLeft.CardArray = make([]int, sph.RepertoryCardArray.MaxCount, sph.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sph.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sph.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sph.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sph.RepertoryCardArray.Kaikou

	if sph.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sph.OutCardUser
	StatusPlay.OutCardData = sph.OutCardData
	StatusPlay.LastOutCardUser = sph.LastOutCardUser

	//玩家的个人数据
	for i := 0; i < sph.GetPlayerCount(); i++ {
		player := sph.GetUserItemByChair(uint16(i))
		if player == nil {
			continue
		}
		if sph.CurrentUser == player.Seat {
			StatusPlay.SendCardData = sph.SendCardData
		} else {
			StatusPlay.SendCardData = public.INVALID_BYTE
		}
		StatusPlay.ActionMask = player.Ctx.UserAction
		if player.Ctx.Response {
			StatusPlay.ActionMask = public.WIK_NULL
		}

		if player.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = player.Ctx.CheckTimeOut
		} else {
			CurUserItem := sph.GetUserItemByChair(sph.CurrentUser)
			if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
				StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
			}

			for _, v := range sph.PlayerInfo {
				if v.GetChairID() == player.GetChairID() {
					continue
				}
				if v.GetChairID() == sph.CurrentUser {
					continue
				}
				if v.Ctx.UserAction > 0 {
					StatusPlay.Overtime = 0
					break
				}
			}
		}
		//扑克数据
		StatusPlay.CardCount, StatusPlay.CardData = sph.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
		sph.VecGameDataAllP[i] = append(sph.VecGameDataAllP[i], StatusPlay) //保存，用于汇总计算
	}
}

// 游戏场景消息发送
func (sph *SportJZHZG) SendGameScene(uid int64, status byte, secret bool) {
	player := sph.GetUserItemByUid(uid)
	if player == nil {
		//不是游戏玩家就是旁观玩家
		player = sph.GetLookonUserItemByUid(uid)
		if player == nil {
			sph.OnWriteGameRecord(public.INVALID_CHAIR, "SendGameScene 发送游戏场景，玩家空指针")
			return
		}
	}
	switch status {
	case public.GS_MJ_FREE:
		sph.SendGameSceneStatusFree(player)
	case public.GS_MJ_PLAY:
		sph.sendGameSceneStatusPlay(player)
	case public.GS_MJ_END:
		sph.sendGameSceneStatusPlay(player)
	}
}

// ! 拼装游戏结束玩家信息
func (sph *SportJZHZG) GetGameEndStr(SeatId uint16, Win bool, ChiHukind uint64) string {
	//useritem := sph.GetUserItemByChair(SeatId)

	GameEndStr := ""

	UserItem := sph.GetUserItemByChair(SeatId)
	if sph.Rule.HasPao {
		GameEndStr += "<color=#ffffff> 跑" + " </color>" + "<color=#ffff00>" + strconv.Itoa(UserItem.Ctx.VecXiaPao.Num) + " </color>"
	}
	if UserItem.Ctx.HidGang > 0 {
		GameEndStr += "<color=#ffffff>" + "暗杠" + "</color>" + "<color=#ffff00>" + "x" + strconv.Itoa(int(UserItem.Ctx.HidGang)) + " </color> "
	}
	if UserItem.Ctx.ShowGang > 0 {
		GameEndStr += "<color=#ffffff>" + "明杠" + "</color>" + "<color=#ffff00>" + "x" + strconv.Itoa(int(UserItem.Ctx.ShowGang)) + " </color> "
	}
	if UserItem.Ctx.XuGang > 0 {
		GameEndStr += "<color=#ffffff>" + "蓄杠" + "</color>" + "<color=#ffff00>" + "x" + strconv.Itoa(int(UserItem.Ctx.XuGang)) + " </color> "
	}
	if Win {
		if (ChiHukind & public.CHK_SI_LAIZI_NO_HUPAI) != 0 {
			GameEndStr += "<color=#ffffff>四红中胡</color>"
		} else if (ChiHukind & public.CHK_QIANG_GANG) != 0 {
			GameEndStr += "<color=#ffffff>抢杠胡</color>"
		} else if (ChiHukind & public.CHK_PING_HU_NOMAGIC) != 0 {
			GameEndStr += "<color=#ffffff>硬胡</color>"
		} else if (ChiHukind & public.CHK_PING_HU_MAGIC) != 0 {
			GameEndStr += "<color=#ffffff>软胡</color>"
		}
	}

	//
	//if UserItem.Ctx.DianGang > 0 {
	//	GameEndStr += "<color=#ffffff>" + "点杠" + "</color>" + "<color=#ffff00>" + "x" + strconv.Itoa(int(UserItem.Ctx.DianGang)) + " </color> "
	//}

	return GameEndStr
}

// ! 游戏退出
func (sph *SportJZHZG) OnExit(uid int64) {
	sph.Common.OnExit(uid)
}

// ! 定时器
func (sph *SportJZHZG) OnTime() {
	sph.Common.OnTime()
}

// ! 计时器事件
func (sph *SportJZHZG) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	if sph.Rule.Overtime_trust > 0 || sph.Rule.NineSecondRoom {
		if TablePerson := sph.GetUserItemByUid(wBindParam); TablePerson != nil {
			//到时间了
			sph.OnAutoOperate(TablePerson.Seat, true)
		}
	}
	return true
}
func (sph *SportJZHZG) setlimitetime(_userItem *modules.Player) {
	if sph.Rule.NineSecondRoom {
		sph.LockTimeOut_15(_userItem, public.GAME_OPERATION_TIME_15)
		return
	}
	if sph.Rule.Overtime_trust > 0 {
		sph.LockTimeOut(_userItem, sph.Rule.Overtime_trust)
		return
	}
	sph.LimitTime = time.Now().Unix() + public.GAME_OPERATION_TIME_15
	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + public.GAME_OPERATION_TIME_15
}

// ! 玩家开启超时
func (sph *SportJZHZG) LockTimeOut_15(_userItem *modules.Player, iTime int) {

	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + public.GAME_OPERATION_TIME_15
	sph.LimitTime = time.Now().Unix() + public.GAME_OPERATION_TIME_15
	if sph.Rule.NineSecondRoom || sph.Rule.Overtime_trust > 0 {
		_userItem.Ctx.Timer.SetTimer(modules.GameTime_15, iTime)
	}
}

// ! 玩家开启超时
func (sph *SportJZHZG) LockTimeOut(_userItem *modules.Player, iTime int) {
	if _userItem == nil {
		return
	}
	checktime := iTime
	if checktime < 1 {
		checktime = public.GAME_OPERATION_TIME_15
	}
	sph.LimitTime = time.Now().Unix() + int64(checktime)
	if _userItem.CheckTRUST() {
		//托管状态
		checktime = 1
	}
	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + int64(checktime)
	if iTime > 0 {
		_userItem.Ctx.Timer.SetTimer(modules.GameTime_Nine, checktime)
	}
}

// ! 玩家关闭超时
func (sph *SportJZHZG) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(sph.GetPlayerCount()) {
		return
	}

	_userItem := sph.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0

	if sph.Rule.NineSecondRoom {
		_userItem.Ctx.Timer.KillTimer(modules.GameTime_Nine)
	}
}

// ! 写日志记录
func (sph *SportJZHZG) WriteGameRecord() {
	//写日志记录
	sph.OnWriteGameRecord(public.INVALID_CHAIR, "开始焦作 推倒胡  发牌......")

	// 玩家手牌
	for i := 0; i < len(sph.PlayerInfo); i++ {
		v := sph.GetUserItemByChair(uint16(i))
		if v != nil {
			handCardStr := fmt.Sprintf("发牌后手牌:%s", sph.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
			sph.OnWriteGameRecord(uint16(v.Seat), handCardStr)
		}
	}

	// 牌堆牌
	leftCardStr := fmt.Sprintf("牌堆牌:%s", sph.m_GameLogic.SwitchToCardNameByDatas(sph.RepertoryCard[0:sph.LeftCardCount+1], 0))
	sph.OnWriteGameRecord(public.INVALID_CHAIR, leftCardStr)

	//赖子牌
	if sph.Rule.WanNengPai {
		magicCardStr := fmt.Sprintf("癞子牌:%s", sph.m_GameLogic.SwitchToCardNameByData(sph.MagicCard, 1))
		sph.OnWriteGameRecord(public.INVALID_CHAIR, magicCardStr)
	}
}

// ! 场景保存
func (sph *SportJZHZG) Tojson() string {
	var _json modules.GameJsonSerializer
	//return _json.ToJson_jz_hzg(sph)
	_json.ToJson(&sph.Metadata)

	_json.GameCommonToJson(&sph.Common)

	return public.HF_JtoA(&_json)
}

// ! 场景恢复
func (sph *SportJZHZG) Unmarsha(data string) {

	if data != "" {

		var _json modules.GameJsonSerializer

		json.Unmarshal([]byte(data), &_json)
		_json.Unmarsha(&sph.Metadata)
		_json.JsonToStruct(&sph.Common)

		sph.ParseRule(sph.GetTableInfo().Config.GameConfig)
		sph.m_GameLogic.Rule = sph.Rule
		sph.m_GameLogic.HuType = sph.HuType
		sph.m_GameLogic.SetMagicCard(sph.MagicCard)
		sph.m_GameLogic.SetPiZiCard(sph.PiZiCard)
		sph.m_GameLogic.SetPiZiCards(sph.PiZiCards)
	}
}

func (sph *SportJZHZG) OnUserScoreOffset(seat uint16, offset int) bool {
	_userItem := sph.GetUserItemByChair(seat)
	if _userItem != nil {
		_userItem.Ctx.StorageScore += offset
	}
	return true
}

func (sph *SportJZHZG) SaveGangOperateResult(msg public.Msg_S_OperateResult) {
	sph.m_bQiangGangResultSend = false
	sph.m_QiangGangOperateResult = msg
}

func (sph *SportJZHZG) ResetTempGangOperateResult() {
	sph.m_bQiangGangResultSend = true
	var tmp public.Msg_S_OperateResult
	sph.m_QiangGangOperateResult = tmp
}

// 杠分
func (sph *SportJZHZG) GetScoreOnGang(gangType uint16) int {
	var Score = 0
	switch gangType {
	case eve.E_Gang_XuGand:
		Score = 1
	case eve.E_Gang_AnGang:
		Score = 2
	case eve.E_Gang:
		Score = 2
	default:
		syslog.Logger().Debug("杠牌类型找不到")
	}
	return Score
}

// level是0，只要弃了就不能再操作，kind是0管胡，1管碰
func (sph *SportJZHZG) CheckNeedGuo(_userItem *modules.Player, cbCheckCard byte, level int, kind int) bool {
	switch level {
	case 0:
		switch kind {
		case 0:
			//过庄
			if len(_userItem.Ctx.VecChiHuCard) != 0 {
				return true
			}
		case 1:
			//过碰
			if len(_userItem.Ctx.VecPengCard) != 0 {
				return true
			}
		default:
			return false
		}
	case 1:
		switch kind {
		case 0:
			//过庄
			return common.Findcard(_userItem.Ctx.VecChiHuCard, cbCheckCard)
		case 1:
			//过碰
			return common.Findcard(_userItem.Ctx.VecPengCard, cbCheckCard)
		default:
			return false
		}
	}
	return false
}

// 从牌库摸一张牌
func (sph *SportJZHZG) DrawOne() byte {
	sph.LeftCardCount--
	return sph.RepertoryCard[sph.LeftCardCount]
}
func (sph *SportJZHZG) GetBingoHorse(sum int, user uint16, cbHouseData *[metadata.MAX_PLAYER][10]byte, bingohousedetail *[metadata.MAX_PLAYER]byte) byte {
	houseList := make([]byte, 0)
	// 得到买马个数
	if vc := int(sph.LeftCardCount); sum > vc {
		sum = vc
	}
	// 没牌了 就不买马
	if sum <= 0 {
		return 0
	}
	// 从牌堆前面翻
	for i := 0; i < sum; i++ {
		houseList = append(houseList, sph.DrawOne())
	}
	sph.OnWriteGameRecord(public.INVALID_CHAIR, fmt.Sprintf("本局翻马详情：%s", sph.m_GameLogic.SwitchToCardNameByDatas(houseList, 1)))
	endhouseList := []byte{}
	//中马记录 躯壳一下，因为没有风，但有红中35
	for i := 0; i < len(houseList); i++ {
		_, color, value := common.Cardsplit(houseList[i])
		if color < 3 || houseList[i] == 0x35 {
			switch value {
			case 1, 5, 9:
				endhouseList = append(endhouseList, houseList[i])
				bingohousedetail[user]++
				sph.addReplayOrder(user, eve.E_Bird, houseList[i])
			}
		}
	}
	//写一下
	//cbHouseData[user]= make([]byte, len(endhouseList))
	for i := 0; i < len(endhouseList); i++ {
		//cbHouseData[user]=append(cbHouseData[user],endhouseList[i])
		cbHouseData[user][i] = endhouseList[i]
	}
	recrodStr := fmt.Sprintf("中马详情：%s", sph.m_GameLogic.SwitchToCardNameByDatas(endhouseList, 1))
	fmt.Println(fmt.Sprintf("中吗（%v）", cbHouseData[user]))
	sph.OnWriteGameRecord(user, recrodStr)
	return bingohousedetail[user]
}

func (sph *SportJZHZG) OnUserTustee(msg *public.Msg_S_DG_Trustee) bool {
	if sph.Rule.Overtime_trust > 0 {
		item := sph.GetUserItemByChair(msg.ChairID)
		if item == nil {
			return false
		}

		if item.CheckTRUST() == msg.Trustee {
			return true
		}
		var tuoguan public.Msg_S_DG_Trustee
		tuoguan.ChairID = msg.ChairID
		tuoguan.Trustee = msg.Trustee
		//校验规则
		if tuoguan.ChairID < public.MAX_PLAYER_4P {
			if tuoguan.Trustee == true /*&& (sph.GameState != gameserver.GsNull)*/ {
				sph.TrustCounts[tuoguan.ChairID]++
				//进入托管啥都不用做
				item.ChangeTRUST(true)
				sph.SendTableMsg(constant.MsgTypeGameTrustee, tuoguan)
				sph.SendTableLookonMsg(constant.MsgTypeGameTrustee, tuoguan)
				//fmt.Println(fmt.Sprintf("(%d)进入托管（%v）状态（%t）",tuoguan.ChairID,sph.TuoGuanPlayer,item.CheckTRUST()))
				return true
			} else if tuoguan.Trustee == false {
				//如果是当前的玩家，那么重新设置一下开始时间
				item.ChangeTRUST(false)
				if tuoguan.ChairID == sph.CurrentUser {
					//sph.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
					_item := sph.GetUserItemByChair(sph.CurrentUser)
					if _item != nil {
						//if time.Now().Unix() < _item.Ctx.CheckTimeOut { //如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < sph.LimitTime
						sph.LockTimeOut(_item, sph.Rule.Overtime_trust)
						//sph.setLimitedTime(int64(sph.PlayTime) + sph.PowerStartTime - time.Now().Unix() + 1)
						tuoguan.Overtime = _item.Ctx.CheckTimeOut
					}
				}
				sph.SendTableMsg(constant.MsgTypeGameTrustee, tuoguan)
				sph.SendTableLookonMsg(constant.MsgTypeGameTrustee, tuoguan)
				//fmt.Println(fmt.Sprintf("(%d)取消托管（%v）状态（%t）",tuoguan.ChairID,sph.TuoGuanPlayer,item.CheckTRUST()))
				return false
			}
		} else {
			//详细日志
			LogStr := string("托管动作:游戏状态不正确 ")
			sph.OnWriteGameRecord(tuoguan.ChairID, LogStr)
			return false
		}
		return false
	}
	return true
}
func (sph *SportJZHZG) Greate_OutCardRecord(handcards string, msg *public.Msg_C_OutCard, _userItem *modules.Player) string {
	tempstr := "客户端的操作"
	if !msg.ByClient {
		tempstr = "服务器自动操作"
		if _userItem.UserOfflineTag^-1 != 0 {
			tempstr += "(离线)"
		}
		if _userItem.CheckTRUST() {
			tempstr += "(托管)"
		}
	}
	return fmt.Sprintf("%s，打出：%s，来源：%s", handcards, sph.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1), tempstr)
}

// ! 创建弃操作消息
func (sph *SportJZHZG) Greate_OperateRecord(msg *public.Msg_C_OperateCard, _userItem *modules.Player) {
	if msg.Code == public.WIK_NULL {
		if msg.ByClient {
			sph.OnWriteGameRecord(_userItem.Seat, "客户端点击弃！")
		} else {
			tempstr := "服务器点击弃！"
			if _userItem.CheckTRUST() {
				tempstr += "(托管)"
			}
			if _userItem.UserOfflineTag^-1 != 0 {
				tempstr += "（离线）"
			}
			sph.OnWriteGameRecord(_userItem.Seat, tempstr)
		}
	}
}

func (sph *SportJZHZG) Greate_ContendRecord(code byte, operate bool, card byte, byclient bool, _userItem *modules.Player) {
	tempstr := fmt.Sprintf("迟到的消息来源于(服务端)")
	if byclient {
		tempstr = fmt.Sprintf("迟到的消息来源于(客户端)")
	}
	if operate {
		tempstr += fmt.Sprintf(",消息体:操作(%s)", public.GetPaiQuanStr(uint64(code)))
		if card != public.INVALID_BYTE {
			tempstr += fmt.Sprintf(",牌(%s)", sph.m_GameLogic.SwitchToCardNameByData(card, 1))
		}
	} else {
		tempstr += fmt.Sprintf(",消息体:打出(%s)", sph.m_GameLogic.SwitchToCardNameByData(card, 1))
	}
	if byclient {
		tempstr += "，将发送刷新消息给客户端"
	}
	sph.OnWriteGameRecord(_userItem.Seat, tempstr)
}
func (sph *SportJZHZG) entryTrust(byclient bool, _userItem *modules.Player) {
	if !byclient {
		if !_userItem.CheckTRUST() {
			//var msg = &public.Msg_C_Trustee{
			//	Id:      _userItem.Uid,
			//	Trustee: true,
			//}
			var msg = &public.Msg_S_DG_Trustee{
				ChairID: _userItem.Seat,
				Trustee: true,
			}
			if sph.OnUserTustee(msg) {
				sph.OnWriteGameRecord(_userItem.Seat, "超时进入托管")
			}
		}
	}
}

func (sph *SportJZHZG) accountManage(dismisstime int, trusttime int, autonexttimer int) int {
	if !(int(sph.CurCompleteCount) >= sph.Rule.JuShu) && sph.CurCompleteCount != 0 {
		check := false
		if dismisstime != -1 {
			for i := 0; i < public.MAX_PLAYER_4P; i++ {
				if item := sph.GetUserItemByChair(uint16(i)); item != nil {
					//fmt.Println(fmt.Sprintf("玩家（%d）托管（%t）", item.Seat, item.CheckTRUST()))
					if item.CheckTRUST() {
						if check {
							var _msg = &public.Msg_C_DismissFriendResult{
								Id:   item.Uid,
								Flag: true,
							}
							sph.OnDismissResult(item.Uid, _msg)
							//fmt.Println(fmt.Sprintf("玩家（%d）托管（%t）同意解散",item.Seat,item.CheckTRUST() ))
							//发准备
							//_msgauto:= &public.Msg_C_GoOnNextGame{
							//	Id:   item.Uid,
							//}
							//sph.OnUserClientNextGame(_msgauto)
						} else {
							check = true
							var msg = &public.Msg_C_DismissFriendReq{
								Id: item.Uid,
							}
							sph.SetDismissRoomTime(dismisstime)
							sph.OnDismissFriendMsg(item.Uid, msg)
							//fmt.Println(fmt.Sprintf("玩家（%d）托管（%t）申请解散",item.Seat,item.CheckTRUST() ))
							//发准备
							//_msgauto:= &public.Msg_C_GoOnNextGame{
							//	Id:   item.Uid,
							//}
							//sph.OnUserClientNextGame(_msgauto)
						}
					}
				}
			}
		}
		//if !check&&trusttime>0&&autonexttimer>0{
		//	//fmt.Println("自动下一局")
		//	sph.SetAutoNextTimer(autonexttimer) //自动开始下一局
		//	return autonexttimer
		//}
		if sph.Rule.Endready {
			sph.SetAutoNextTimer(10)
		}
	}
	return 0
}

func (sph *SportJZHZG) OnScoreOffset(total []float64) {
	for i, t := range total {
		var order metadata.Replay_Order
		order.Chair_id = uint16(i)
		order.Operation = eve.E_GameScore
		order.UserScore = t
		sph.ReplayRecord.VecOrder = append(sph.ReplayRecord.VecOrder, order)
	}
}

// ! 发送游戏开始场景数据
func (sph *SportJZHZG) sendGameSceneStatusPlayLookon(player *modules.Player) bool {

	if player.LookonTableId == 0 {
		return false
	}
	wChiarID := player.GetChairID()
	if int(wChiarID) >= sph.GetPlayerCount() {
		wChiarID = 0
	}
	//是否要获取wChiarID位置真正玩家的信息 ？
	playerOnChair := sph.GetUserItemByChair(wChiarID)

	//变量定义
	var StatusPlay public.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = sph.SiceCount
	StatusPlay.BankerUser = sph.BankerUser
	StatusPlay.CurrentUser = sph.CurrentUser
	StatusPlay.CellScore = sph.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sph.PiZiCard

	StatusPlay.KaiKou = sph.Rule.KouKou
	StatusPlay.RenWuAble = sph.RenWuAble
	StatusPlay.TheOrder = sph.CurCompleteCount
	//if sph.CurrentUser == wChiarID {
	//	StatusPlay.SendCardData = sph.SendCardData
	//} else {
	StatusPlay.SendCardData = public.INVALID_BYTE
	//}
	//断线重连回来也把弃杠的牌发出去
	NowUserItem := sph.GetUserItemByChair(sph.CurrentUser)
	if NowUserItem != nil {
		StatusPlay.VecGangCard = public.HF_BytesToInts(NowUserItem.Ctx.VecGangCard)
	}

	StatusPlay.CardLeft.CardArray = make([]int, sph.RepertoryCardArray.MaxCount, sph.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sph.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sph.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sph.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sph.RepertoryCardArray.Kaikou
	for i := 0; i < sph.GetPlayerCount(); i++ {
		_item := sph.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//玩家番数
		//StatusPlay.PlayerFan[i] = sph.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status
		if _item.CheckTRUST() {
			StatusPlay.Whotrust[i] = true
		}
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}
	//状态变量
	StatusPlay.ActionCard = sph.ProvideCard
	StatusPlay.LeftCardCount = sph.LeftCardCount
	StatusPlay.ActionMask = 0
	StatusPlay.ChiHuKindMask = 0
	StatusPlay.CardLeft.CardArray = make([]int, sph.RepertoryCardArray.MaxCount, sph.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sph.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sph.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sph.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sph.RepertoryCardArray.Kaikou
	if playerOnChair != nil && playerOnChair.Ctx.Response {
		StatusPlay.ActionMask = public.WIK_NULL
	}

	if playerOnChair != nil && playerOnChair.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = playerOnChair.Ctx.CheckTimeOut
	} else {
		CurUserItem := sph.GetUserItemByChair(sph.CurrentUser)
		if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
		}

		for _, v := range sph.PlayerInfo {
			if v.GetChairID() == player.GetChairID() {
				continue
			}
			if v.GetChairID() == sph.CurrentUser {
				continue
			}
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if sph.OutCardUser == public.INVALID_CHAIR {
		if (StatusPlay.ActionMask & public.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sph.OutCardUser
	StatusPlay.OutCardData = sph.OutCardData

	//扑克数据
	if playerOnChair != nil {
		StatusPlay.CardCount, StatusPlay.CardData = sph.m_GameLogic.SwitchToCardData2(playerOnChair.Ctx.CardIndex, StatusPlay.CardData)
		StatusPlay.CardData = [14]byte{}
	}

	//发送旁观数据
	sph.SendPersonLookonMsg(constant.MsgTypeGameStatusPlay, StatusPlay, player.Uid)

	//发小结消息
	if byte(len(sph.VecGameEnd)) == sph.CurCompleteCount && sph.CurCompleteCount != 0 && int(wChiarID) < sph.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := sph.VecGameEnd[sph.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		sph.SendPersonLookonMsg(constant.MsgTypeGameEnd, gamend, player.Uid)
	}

	//发送解散房间所有玩家的反应
	sph.SendAllPlayerDissmissInfo(player)

	if sph.PayPaoStatus {
		var PaoSetting public.Msg_S_PaoSetting
		sph.SendPersonLookonMsg(constant.MsgTypeGamePaoSetting, PaoSetting, player.Uid)
		//已经选漂的人，需要告诉旁观者
		var sXiaPiao public.Msg_S_Xiapao
		for _, v := range sph.PlayerInfo {
			if v.Seat != public.INVALID_CHAIR {
				sXiaPiao.Num[v.Seat] = v.Ctx.VecXiaPao.Num
				sXiaPiao.Always[v.Seat] = v.Ctx.VecXiaPao.Status
				sXiaPiao.Status[v.Seat] = v.Ctx.UserPaoReady
			}
		}
		sph.SendPersonLookonMsg(constant.MsgTypeGameXiapao, sXiaPiao, player.Uid)
	}

	return true
}
