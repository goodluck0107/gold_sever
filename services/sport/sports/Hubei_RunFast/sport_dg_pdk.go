package Hubei_RunFast

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"math/rand"
	"strings"
	"time"
)

type GoodCardType int32

const (
	GoodCardTypeBigCard GoodCardType = 0
	GoodCardTypeBoom    GoodCardType = 1
	GoodCardTypeAir     GoodCardType = 2
)

type GoodCard struct {
	Type   GoodCardType
	Points []byte
}

var goodCardsConfig = []GoodCard{
	{
		Type:   GoodCardTypeBigCard,
		Points: []byte{2, 1, 13, 13},
	},
	{
		Type:   GoodCardTypeBigCard,
		Points: []byte{2, 13, 13},
	},
	{
		Type:   GoodCardTypeBigCard,
		Points: []byte{1, 13, 13},
	},
	{
		Type:   GoodCardTypeBigCard,
		Points: []byte{1, 13, 12, 12, 12},
	},
	{
		Type:   GoodCardTypeBoom,
		Points: nil,
	},
	{
		Type:   GoodCardTypeAir,
		Points: nil,
	},
}

/*
跑得快好友房
*/

type SportPDK struct {
	components2.Common
	//游戏变量
	meta2.GameMetaDG
	m_GameLogic SportLogicPDK //游戏逻辑
}

func (spd *SportPDK) GetGameConfig() *static.GameConfig { //获取游戏相关配置
	return &spd.Config
}

// 重置桌子数据
func (spd *SportPDK) RepositTable(ResetAllData bool) {
	rand.Seed(time.Now().UnixNano())
	for _, v := range spd.PlayerInfo {
		v.Reset()
	}
	//游戏变量
	spd.GameEndStatus = static.GS_MJ_FREE
	if ResetAllData {
		//游戏变量
		spd.GameMetaDG.Reset()
	} else {
		//游戏变量
		spd.GameMetaDG.ResetForNext() //保留部分数据给下一局使用
	}
}

func (spd *SportPDK) switchCard2Ox(_index int) int {
	_hight := (_index - 1) / 13
	_low := (_index-1)%13 + 1

	return (_low + (_hight << 4))
}

func (spd *SportPDK) getCardIndexByOx(_card byte) int {
	low_index := int(0x0F)
	hight_index := int(0xF0)

	return ((low_index & int(_card)) + ((hight_index&int(_card))>>4)*13)
}

// 解析配置的任务
func (spd *SportPDK) ParseRule(strRule string) {

	spd.OnWriteGameRecord(static.INVALID_CHAIR, "rule = "+strRule)

	xlog.Logger().Info("parserRule :" + strRule)

	//表示底分要除以10
	spd.Rule.DiFen = 1
	spd.Rule.BaseScore = 1
	spd.Rule.NineSecondRoom = false

	spd.Rule.JuShu = spd.GetTableInfo().Config.RoundNum
	spd.Spay = 0
	spd.Rule.FangZhuID = spd.GetTableInfo().Creator
	spd.Rule.CreateType = spd.FriendInfo.CreateType

	spd.SerPay = 0 //好友房无茶水

	if len(strRule) == 0 {
		return
	}

	var _msg rule2.FriendRuleDG_qc
	if err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg); err == nil {
		spd.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%#v", _msg))
		if _msg.Radix == 0 {
			spd.IBase = _msg.Difen
			spd.Rule.Radix = 1
		} else {
			spd.IBase = _msg.Difen
			spd.Rule.Radix = _msg.Radix
		}
		spd.SerPay = _msg.SerPay
		spd.FaOfTao = _msg.Fa
		spd.JiangOfTao = _msg.Jiang
		spd.BiYa = _msg.BiYa
		spd.TuoGuan = -1
		if _msg.Overtime_trust > 0 || _msg.Overtime_trust == -1 {
			spd.TuoGuan = _msg.Overtime_trust
		} else {
			spd.TuoGuan = _msg.TuoGuan
		}
		if spd.TuoGuan > 300 {
			spd.TuoGuan = 300 //最多300秒
		}
		spd.Rule.NineSecondRoom = spd.PlayTime > 0
		spd.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		spd.ZhaNiao = (_msg.ZhaNiao == "true")
		spd.FourTake2 = (_msg.FourTake2 == "true")
		spd.FourTake3 = (_msg.FourTake3 == "true")
		spd.BombSplit = (_msg.BombSplit == "true")
		spd.QuickPass = (_msg.QuickPass == "true")
		spd.SplitCards = (_msg.SplitCards == "true")
		spd.Bomb3Ace = (_msg.Bomb3Ace == "true")
		spd.LessTake = (_msg.LessTake == "true")
		spd.KeFan = (_msg.KeFan == "true")
		spd.Rule.DissmissCount = _msg.Dissmiss
		spd.BombMode = (_msg.BombMode == "true")
		spd.BombRealTime = _msg.BombRealTime == "true"
		spd.Rule.NoWan = _msg.Card15 == "true"
		spd.Rule.MingTing = _msg.Must3 == "true"
	} else {
		spd.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%#v", err))
	}
	if spd.IBase < 1 {
		spd.IBase = 1
	}
	//开关
	if spd.Debug > 0 {
		//Rule.JuShu = Debug;
	}
	if spd.Rule.DissmissCount != 0 {
		spd.SetDissmissCount(spd.Rule.DissmissCount)
	}
}

// ! 开局
func (spd *SportPDK) OnBegin() {
	xlog.Logger().Info("onbegin")
	spd.RepositTable(true)

	spd.OnWriteGameRecord(static.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range spd.PlayerInfo {
		v.OnBegin()
		//重置托管状态
		v.ChangeTRUST(false)
	}

	//设置状态
	spd.SetGameStatus(static.GS_MJ_PLAY)

	spd.ParseRule(spd.GetTableInfo().Config.GameConfig)
	spd.m_GameLogic.m_rule = spd.Rule
	spd.CurCompleteCount = 0
	spd.VecGameEnd = []static.Msg_S_DG_GameEnd{}

	// 记录游戏开始时间
	spd.Common.GameBeginTime = time.Now()

	if spd.Rule.NoWan {
		spd.OnWriteGameRecord(static.INVALID_CHAIR, "[15张玩法]")
		_, spd.AllCards = spd.m_GameLogic.CreateCards15()
	} else {
		_, spd.AllCards = spd.m_GameLogic.CreateCards()
	}
	spd.m_GameLogic.SetBombCount(4)                               //设置炸弹的最小长度
	spd.m_GameLogic.SetOnestrCount(5)                             //设置单顺的最小长度，254表示无顺子
	spd.m_GameLogic.SetMaxCardCount(MAXHANDCARD)                  //设置手牌最大长度
	spd.m_GameLogic.SetMaxPlayerCount(byte(spd.GetPlayerCount())) //设置玩家最大数目
	spd.m_GameLogic.SetBombSplit(spd.BombSplit)                   //设置炸弹是否可以拆，true表示不可拆
	spd.m_GameLogic.SetThreeAceBomb(spd.Bomb3Ace)                 //设置3个A是否是炸弹
	spd.m_GameLogic.Set4Take2(spd.FourTake2)
	spd.m_GameLogic.Set4Take3(spd.FourTake3)  //设置是否可以4带3
	spd.m_GameLogic.SetLessTake(spd.LessTake) //设置最后一手是否可以少带
	if spd.Rule.NoWan {
		spd.m_GameLogic.SetCardNum(byte(15)) //设置张数
	} else {
		spd.m_GameLogic.SetCardNum(byte(16)) //设置张数
	}
	spd.PlayTime = 30
	spd.AutoOutTime = 0
	spd.GameMetaDG.TimeStart = 6 //切牌等待时间
	if spd.TuoGuan > 0 {
		spd.PlayTime = spd.TuoGuan
		//spd.AutoOutTime = spd.TuoGuan
	}
	spd.Nextbanker = 0
	spd.CurCompleteCount++
	spd.GetTable().SetBegin(true)
	spd.SetOfflineRoomTime(1800)
	spd.OnGameStart()
}

func (spd *SportPDK) OnGameStart() {
	if !spd.CanContinue() {
		return
	}
	//第一局不切牌
	spd.StartNextGame()
}

// 切牌并且下一局
func (spd *SportPDK) SplitAndGameStart() {
	if !spd.CanContinue() {
		return
	}
	if spd.SplitCards {
		spd.SendSplitCards(static.MAX_PLAYER_3P, spd.GameMetaDG.TimeStart)
		spd.GameState = meta2.GsSplitCards     //设置玩家切牌的状态
		spd.PowerStartTime = time.Now().Unix() //权限开始时间
		spd.setLimitedTime(int64(spd.GameMetaDG.TimeStart + 1))
	} else {
		spd.StartNextGame()
	}
}

// 发送切牌对话框
func (spd *SportPDK) SendSplitCards(Seat uint16, lefttime int) {

	spd.GameEndStatus = static.GS_MJ_PLAY
	//设置状态
	spd.SetGameStatus(static.GS_MJ_PLAY)

	var SplitCards static.Msg_S_DG_SplitCardsStart
	SplitCards.CardCount = MAX_POKER_COUNTS - 6
	SplitCards.LeftTime = lefttime
	SplitCards.SplitUser = uint16((int(spd.Nextbanker) + spd.GetPlayerCount() - 1) % spd.GetPlayerCount())
	if int(Seat) >= spd.GetPlayerCount() {
		spd.SendTableMsg(consts.MsgTypeSplitCardStart, SplitCards)
	} else {
		spd.SendPersonMsg(consts.MsgTypeSplitCardStart, SplitCards, Seat)
	}
}

func (spd *SportPDK) OnUserEndSplitCards(msg *static.Msg_C_DG_SplitCards) bool {
	nChiarID := spd.GetChairByUid(msg.Id)
	_userItem := spd.GetUserItemByChair(nChiarID)
	if _userItem == nil {
		return false
	}
	if int(nChiarID) < spd.GetPlayerCount() && nChiarID == uint16((int(spd.Nextbanker)+spd.GetPlayerCount()-1)%spd.GetPlayerCount()) {
		var SplitCards static.Msg_S_DG_SplitCards
		SplitCards.CardCount = MAX_POKER_COUNTS - 6
		SplitCards.Cardindex = msg.Cardindex
		SplitCards.SplitUser = nChiarID
		SplitCards.SplitStatus = true
		SplitCards.SplitType = msg.SplitType
		spd.SendTableMsg(consts.MsgTypeSplitCard, SplitCards)
	}

	if nChiarID != uint16((int(spd.Nextbanker)+spd.GetPlayerCount()-1)%spd.GetPlayerCount()) {
		recordStr := fmt.Sprintf("椅子编号[%d] 还没有完成切牌", nChiarID)
		spd.OnWriteGameRecord(nChiarID, recordStr)
		return false
	}
	recordStr := "【手动】"
	if msg.SplitType != 0 {
		recordStr = "【超时自动】"
	}
	recordStr += fmt.Sprintf("庄家的上家[%d]完成切牌了，开始游戏", nChiarID)
	spd.OnWriteGameRecord(nChiarID, recordStr)
	//游戏没有开始发牌
	if spd.GameState == meta2.GsSplitCards {
		spd.GameState = meta2.GsStartAnimation
		spd.PowerStartTime = time.Now().Unix() //权限开始时间
		spd.setLimitedTime(3)                  //
	}

	return true
}

// 开始下一局游戏
func (spd *SportPDK) StartNextGame() {
	spd.OnStartNextGame()

	//设置自动解散时间2分钟
	spd.SetDismissRoomTime(120)

	spd.GameState = meta2.GsNull          //吼牌阶段或打牌阶段.
	spd.GameEndStatus = static.GS_MJ_PLAY //当前小局游戏的状态
	spd.ReWriteRec = 0
	spd.BSplited = false //只能放这里复位

	//发送最新状态
	for i := 0; i < spd.GetPlayerCount(); i++ {
		spd.SendUserStatus(i, static.US_PLAY) //把状态发给其他人
	}

	//记录日志
	spd.OnWriteGameRecord(static.INVALID_CHAIR, "开始StartNextGame......")
	spd.OnWriteGameRecord(static.INVALID_CHAIR, spd.GetTableInfo().Config.GameConfig)
	spd.ReplayRecord.UVitamin = make(map[int64]float64)
	spd.ReplayRecord.ReSet()

	//重置所有玩家的状态
	for _, v := range spd.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	//spd.ParseRule(spd.GameTable.Config.GameConfig)
	//spd.m_GameLogic.Rule = spd.Rule

	//设置状态
	spd.SetGameStatus(static.GS_MJ_PLAY)

	// 框架发送开始游戏后开始计算当前这一轮的局数
	//spd.CurCompleteCount++
	//spd.GetTable().SetBegin(true)

	LogNextStr := fmt.Sprintf("【当前第%d局】", spd.CurCompleteCount)
	spd.OnWriteGameRecord(static.INVALID_CHAIR, LogNextStr)

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(spd.GetTableId()+spd.KIND_ID*100+spd.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	spd.SiceCount = components2.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))

	tempCards := spd.AllCards
	//var tempCards [static.ALL_CARD]byte
	//for idx, card := range spd.AllCards {
	//	tempCards[idx] = card
	//}
	//检测上轮牌
	if spd.CurCompleteCount > 1 {
		count := 0
		allcardcnt := MAX_CARDS_16
		if spd.Rule.NoWan {
			allcardcnt = MAX_CARDS_15
		}
		for i := 0; i < allcardcnt; i++ {
			if spd.m_GameLogic.IsValidCard(spd.LastCards[i]) {
				count++
			}
		}
		if count == allcardcnt {
			tempCards = spd.LastCards
		}
	}

	// 检查高权限人本轮是否需要good
	notCheckSeat := make([]uint16, 0)
	//for i := 0; i < spd.GetPlayerCount(); i++ {
	//	_userItem := spd.GetUserItemByChair(uint16(i))
	//	if _userItem != nil {
	//		if _userItem.Ctx.WantGood {
	//			notCheckSeat = append(notCheckSeat, _userItem.GetChairID())
	//		}
	//	}
	//}

	goodCard := make([]int, 0, len(notCheckSeat))
	gcl := len(goodCardsConfig)
	if gcl > 0 {
		for i := 0; i < len(notCheckSeat); i++ {
			idx := static.HF_GetRandom(gcl)
			var find bool
			for j := 0; j < len(goodCard); j++ {
				if goodCard[i] == idx {
					find = true
					break
				}
			}
			if find {
				i--
			} else {
				goodCard = append(goodCard, idx)
			}
		}
	}

	//cardPoints := make([][]byte, 0, len(goodCard))
	//for _, idx := range goodCard {
	//	if idx >= 0 && idx < gcl {
	//		// 不能出现一样的gc
	//		gc := goodCardsConfig[idx]
	//		switch gc.Type {
	//		case GoodCardTypeBigCard:
	//			cardPoints = append(cardPoints, gc.Points)
	//		case GoodCardTypeBoom:
	//			points := make([]byte, 0, 4)
	//			rdm := static.HF_GetRandom(12) + 1 // A-K
	//			for i := 0; i < 4; i++ {
	//				points = append(points, byte(rdm))
	//			}
	//			cardPoints = append(cardPoints, points)
	//		case GoodCardTypeAir:
	//			airCount := static.HF_GetRandom(3) + 1               // 1-3个飞机
	//			startPoint := static.HF_GetRandom(12) + 1 - airCount // A-Q / A-Q / A-J
	//			if startPoint <= 2 {                                 // 不能从 A 2 开始
	//				startPoint = 3
	//			}
	//			points := make([]byte, 0, 3*airCount)
	//			for i := 0; i < airCount; i++ {
	//				for j := 0; j < 3; j++ {
	//					points = append(points, byte(startPoint+i))
	//				}
	//			}
	//			cardPoints = append(cardPoints, points)
	//		}
	//	}
	//}

	// var wantGoodCard [MAX_PLAYER][static.MAX_CARD]byte

	//// 从tempCards从拿到这些想要的牌
	//for idx, pts := range cardPoints {
	//	seat := notCheckSeat[idx]
	//	spd.OnWriteGameRecord(seat, fmt.Sprintf("好牌点数: %s", pts))
	//	wantGoodCardCount := 0
	//	for _, point := range pts {
	//		for i := 0; i < len(tempCards); i++ {
	//			card := tempCards[i]
	//			if spd.m_GameLogic.IsValidCard(card) && spd.m_GameLogic.GetCardPoint(card) == point {
	//				wantGoodCard[seat][wantGoodCardCount] = card
	//				wantGoodCardCount++
	//			}
	//		}
	//	}
	//	spd.OnWriteGameRecord(seat, fmt.Sprintf("好牌: %s", wantGoodCard[seat]))
	//}

	//	var tryTimes int
	//
	//TRYCARD:
	//
	//	tryTimes++

	_randStr, restartcnt := 0, 0

	//50%拆除飞机连对
	_randStr = rand.Intn(1000) % 100

	for i := 0; i < 50; i++ {
		//分发扑克
		if spd.Rule.NoWan { // 15张
			spd.ThePaiCount, spd.PlayerCards, spd.LastCards = spd.m_GameLogic.RandCardData15(tempCards)
		} else {
			spd.ThePaiCount, spd.PlayerCards, spd.LastCards = spd.m_GameLogic.RandCardData(tempCards)
		}

		spd.m_GameLogic.GetPlayerGroupTypeByPoint(spd.PlayerCards)

		if spd.m_GameLogic.PokerStrCheck(3, notCheckSeat) || spd.m_GameLogic.PokerStrCheck(2, notCheckSeat) {
			if _randStr < 0 /*50*/ {
				continue
			}
		}
		restartcnt = i
		break
	}

	splitStr := fmt.Sprintf("拆飞机连对概率:%d,重发次数:%d", _randStr, restartcnt)

	spd.OnWriteGameRecord(uint16(spd.Banker), splitStr)

	//检测是否有炸弹
	if spd.m_GameLogic.PokerBombCheck(notCheckSeat) {
		_temp := rand.Intn(100)
		//是否勾选纯净玩法
		if spd.NoBomb || _temp < 0 /*50*/ {
			//拆炸弹换牌
			spd.PlayerCards = spd.m_GameLogic.BombSplit(spd.PlayerCards)
		}
	}

	//确认手牌
	checkArray := [14]int{}
	for _, handcards := range spd.PlayerCards {
		for _, card := range handcards {
			if card > 0 && card < CARDINDEX_SMALL {
				point := spd.m_GameLogic.GetCardPoint(card)
				checkArray[point]++
				if checkArray[point] > 4 {
					//重新发牌
					if spd.Rule.NoWan { // 15張
						spd.ThePaiCount, spd.PlayerCards, spd.LastCards = spd.m_GameLogic.RandCardData15(tempCards)
					} else {
						spd.ThePaiCount, spd.PlayerCards, spd.LastCards = spd.m_GameLogic.RandCardData(tempCards)
					}
					spd.OnWriteGameRecord(uint16(spd.Banker), "发牌有误，重新发牌")
					spd.m_GameLogic.GetPlayerGroupTypeByPoint(spd.PlayerCards)
					break
				}
			}
		}
	}

	//if len(notCheckSeat) < spd.GetPlayerCount() {
	//	for i := 0; i < spd.GetPlayerCount(); i++ {
	//		_userItem := spd.GetUserItemByChair(uint16(i))
	//		if _userItem != nil {
	//			if _userItem.Ctx.WantGood {
	//				if cards, single := spd.m_GameLogic.PokerSingleCount(_userItem.GetChairID()); single > 3 {
	//					spd.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("vip单张大于3，重新洗牌,手牌%v, 单牌数量%d", cards, single))
	//					if tryTimes >= 200 {
	//						spd.OnWriteGameRecord(_userItem.GetChairID(), "单张大于3，但是超出200次重洗次数，放弃了")
	//						break
	//					} else {
	//						goto TRYCARD
	//					}
	//				}
	//			} else {
	//				if cards, single := spd.m_GameLogic.PokerSingleCount(_userItem.GetChairID()); single < 5 {
	//					spd.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("非vip单张小于5，重新洗牌,手牌%v, 单牌数量%d", cards, single))
	//					if tryTimes >= 200 {
	//						spd.OnWriteGameRecord(_userItem.GetChairID(), "单张大于3，但是超出200次重洗次数，放弃了")
	//						break
	//					} else {
	//						goto TRYCARD
	//					}
	//				}
	//			}
	//		}
	//	}
	//} else {
	//	spd.OnWriteGameRecord(static.INVALID_CHAIR, "桌上都是高权限玩家。")
	//}

	for i := 0; i < spd.GetPlayerCount(); i++ {
		_userItem := spd.GetUserItemByChair(uint16(i))
		if _userItem != nil {
			_userItem.Ctx.WantGood = false
		}
	}

	// spd.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("高权限牌重发次数: %d", tryTimes))

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	spd.initDebugCards("dagongPaoDeiKuai_test", &spd.PlayerCards, &spd.Nextbanker, &spd.DownPai)
	//////////////读取配置文件设置牌型end////////////////////////////////////

	//重新计算牌数目
	for i := 0; i < static.MAX_PLAYER_3P; i++ {
		spd.ThePaiCount[i] = spd.m_GameLogic.GetCardNum(spd.PlayerCards[i], MAXHANDCARD)
	}

	//发送玩家的牌数目
	spd.SendPaiCount(static.MAX_PLAYER_3P)

	if spd.GetPlayerCount() == 3 && spd.CurCompleteCount == 1 {
		//黑桃3先出(三人场)
		for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
			if spd.m_GameLogic.ISHaveSpade3(spd.PlayerCards[i][:]) {
				spd.Nextbanker = i
				break
			}
		}
	} else if spd.GetPlayerCount() == 2 && spd.CurCompleteCount == 1 {
		//最小牌先出(二人场)
		min := byte(56)
		for i := uint16(0); i < static.MAX_PLAYER_2P; i++ {
			iMin := spd.m_GameLogic.GetMinCard(spd.PlayerCards[i][:])
			if !spd.m_GameLogic.ISMin(min, iMin) {
				min = iMin
				spd.Nextbanker = i
			}
		}
	}
	if spd.ZhaNiao {
		//抓鸟
		for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
			if spd.m_GameLogic.ISHaveHeart10(spd.PlayerCards[i][:]) {
				spd.Bird[i] = 1
				break
			}
		}
	}

	//确定庄家，随机坐庄
	if spd.Nextbanker >= uint16(spd.GetPlayerCount()) {
		rand_num := rand.Intn(1000)
		spd.Banker = uint16(rand_num % spd.GetPlayerCount())
	} else {
		spd.Banker = spd.Nextbanker
	}
	spd.Whoplay = spd.Banker

	for seat := 0; seat < spd.GetPlayerCount(); seat++ {
		//详细日志
		handCardStr := string("发牌后手牌:")
		for i := 0; i < MAXHANDCARD; i++ {
			temCardStr := fmt.Sprintf("0x%02x,", spd.switchCard2Ox(int(spd.PlayerCards[seat][i])))
			handCardStr += temCardStr
		}
		if spd.Banker == uint16(seat) {
			handCardStr += "首出玩家"
		}
		spd.OnWriteGameRecord(uint16(seat), handCardStr)

		spd.ReplayRecord.R_HandCards[seat] = append(spd.ReplayRecord.R_HandCards[seat], static.HF_BytesToInts(spd.PlayerCards[seat][:])...)
	}
	//构造数据,发送开始信息
	var GameStart static.Msg_S_DG_GameStart
	GameStart.BankerUser = spd.Banker
	GameStart.CurrentUser = spd.Whoplay
	GameStart.CurCompleteCount = spd.CurCompleteCount

	//向每个玩家发送数据
	for i := 0; i < spd.GetPlayerCount(); i++ {
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//设置变量
		GameStart.MySeat = uint16(i)
		//GameStart.Overtime = spd.LimitTime
		GameStart.CardCount = spd.ThePaiCount[i]
		for c := 0; c < MAXHANDCARD; c++ {
			GameStart.CardData[c] = spd.PlayerCards[i][c]
		}
		//记录玩家初始分

		//TODO 玩家分数设置
		spd.ReplayRecord.R_Score[i] = spd.Total[i]
		spd.ReplayRecord.UVitamin[_item.Uid] = _item.UserScoreInfo.Vitamin
		//发送数据
		spd.SendPersonMsg(consts.MsgTypeGameStart, GameStart, uint16(i))
	}
	spd.GameType = meta2.GT_NORMAL
	spd.StartPlay(spd.Whoplay)

}

// 发送操作
func (spd *SportPDK) SendPaiCount(wChairID uint16) {

	//构造数据
	var SendCount static.Msg_S_DG_SendCount
	for i := 0; i < spd.GetPlayerCount(); i++ {
		SendCount.CardCount[i] = spd.ThePaiCount[i]
	}
	//发送数据
	if wChairID >= uint16(spd.GetPlayerCount()) {
		spd.SendTableMsg(consts.MsgTypeSendPaiCount, SendCount)
	} else {
		for _, v := range spd.PlayerInfo {
			if v.Seat == uint16(wChairID) {
				spd.SendPersonMsg(consts.MsgTypeSendPaiCount, SendCount, v.Seat)
			}
		}
	}
}

// 发送本轮分
func (spd *SportPDK) SendTurnScore(wChairID uint16) {
	var turnScore static.Msg_S_DG_TurnScore
	turnScore.TurnScore = spd.CardScore

	if wChairID >= uint16(spd.GetPlayerCount()) {
		spd.SendTableMsg(consts.MsgTypeTurnScore, turnScore)
	} else {
		spd.SendPersonMsg(consts.MsgTypeTurnScore, turnScore, wChairID)
	}
}

// 发送玩家分
func (spd *SportPDK) SendPlayerScore(wChairID uint16, wGetChairID uint16, iGetScore int) {
	var playScore static.Msg_S_DG_PlayScore
	playScore.ChairID = wGetChairID
	playScore.GetScore = iGetScore

	for i := 0; i < spd.GetPlayerCount(); i++ {
		playScore.PlayScore[i] = spd.PlayerCardScore[i]
	}

	if wChairID >= uint16(spd.GetPlayerCount()) {
		spd.SendTableMsg(consts.MsgTypePlayScore, playScore)
	} else {
		spd.SendPersonMsg(consts.MsgTypePlayScore, playScore, wChairID)
	}
}

// 发送声音
func (spd *SportPDK) SendPlaySoundMsg(seat uint16, bySoundType byte) {
	var soundmsg static.Msg_S_DG_PlaySound
	soundmsg.CurrentUser = seat
	soundmsg.SoundType = bySoundType

	spd.SendTableMsg(consts.MsgTypePlaySound, soundmsg)
}

// 发送玩家是几游
func (spd *SportPDK) SendPlayerTurn(wChairID uint16) {
	var turnmsg static.Msg_S_DG_SendTurn

	///////////////////////////////////////////////////////////////////////////////////////////////////////
	for i := 0; i < spd.GetPlayerCount(); i++ {
		turnmsg.Turn[i] = spd.PlayerTurn[i]
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	if wChairID >= uint16(spd.GetPlayerCount()) {
		spd.SendTableMsg(consts.MsgTypeSendTurn, turnmsg)
	} else {
		spd.SendPersonMsg(consts.MsgTypeSendTurn, turnmsg, wChairID)
	}
}

// 把seat1的牌发给seat2
func (spd *SportPDK) SendPaiToTeamer(seat1 uint16, seat2 uint16) {
	if seat1 < 0 || seat1 >= static.MAX_PLAYER_3P {
		return
	}
	if seat2 < 0 || seat2 >= static.MAX_PLAYER_3P {
		return
	}
	if !spd.BMingJiFlag {
		return
	} //没有明鸡，不发送队友的牌

	var teamerPai static.Msg_S_DG_TeamerPai
	teamerPai.WhoPai = seat1
	for i := 0; i < MAXHANDCARD; i++ {
		if spd.m_GameLogic.IsValidCard(spd.PlayerCards[seat1][i]) {
			teamerPai.CardData[teamerPai.CardCount] = spd.PlayerCards[seat1][i]
			teamerPai.CardCount++
		}
	}
	spd.SendPersonMsg(consts.MsgTypeTeamerPai, teamerPai, seat2)
}

// 发送权限
func (spd *SportPDK) SendPower(whoplay uint16, iPower int, iWaitTime int) {
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	//构造数据
	var power static.Msg_S_DG_Power
	iOvertime := 0
	if iWaitTime > 0 {
		iOvertime = iWaitTime
	}
	power.CurrentUser = whoplay
	power.Power = iPower
	item := spd.GetUserItemByChair(whoplay)
	if item.CheckTRUST() {
		if iOvertime > spd.AutoOutTime {
			iOvertime = spd.AutoOutTime
		}
	}
	power.Overtime = int64(iOvertime)

	//详细日志
	LogStr := fmt.Sprintf("SUB_S_POWER seat=%d power=%d,time=%d ", power.CurrentUser, power.Power, power.Overtime)
	spd.OnWriteGameRecord(power.CurrentUser, LogStr)

	spd.SendTableMsg(consts.MsgTypeSendPower, power)

	spd.PowerStartTime = time.Now().Unix() //权限开始时间
	spd.setLimitedTime(int64(iOvertime + 1))
	if spd.GameState == meta2.GsPlay {
		//SetActionStep(AS_PLAY,nTime + 1);//设置等待时间，服务端多等一下
	} else if spd.GameState == meta2.GsRoarPai {
		//SetActionStep(AS_ROAR,nTime + 1);//设置等待时间，服务端多等一下
	}
}

func (spd *SportPDK) setLimitedTime(iLimitTime int64) {
	// fmt.Println(fmt.Sprintf("limitetimeOP(%d)", spd.Rule.limitetimeOP))
	spd.LimitTime = time.Now().Unix() + iLimitTime
	spd.GameTimer.SetLimitTimer(int(iLimitTime))
}

func (spd *SportPDK) freeLimitedTime() {
	spd.GameTimer.KillLimitTimer()
}

func (spd *SportPDK) LockTimeOut(cUser uint16, iTime int64) {
	if cUser < 0 || cUser > uint16(spd.GetPlayerCount()) {
		return
	}

	_userItem := spd.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = iTime
}

func (spd *SportPDK) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(spd.GetPlayerCount()) {
		return
	}

	_userItem := spd.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0
}

// 计时器事件
func (spd *SportPDK) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	//游戏定时器
	if dwTimerID == components2.GameTime_Nine {
		spd.OnAutoOperate(true)
	}
	return true
}

// 暂时空着
func (spd *SportPDK) OnAutoOperate(bBreakin bool) {
	fmt.Println("自动操作")
	//详细日志
	LogStr := string("OnAutoOperate 自动操作!!! ")
	spd.OnWriteGameRecord(spd.Whoplay, LogStr)

	item := spd.GetUserItemByChair(spd.Whoplay)

	if spd.GameState == meta2.GsSplitCards {
		var msg static.Msg_C_DG_SplitCards
		msg.SplitType = 1
		msg.Cardindex = 8
		msg.Id = spd.GetUidByChair(uint16((int(spd.Nextbanker) + spd.GetPlayerCount() - 1) % spd.GetPlayerCount()))
		spd.OnUserEndSplitCards(&msg)

	} else if spd.GameState == meta2.GsStartAnimation {
		spd.StartNextGame()
	} else if spd.GameState == meta2.GsQuickPass {
		var outmsg static.Msg_C_DG_OutCard
		outmsg.CurrentUser = spd.Whoplay
		outmsg.CardCount = 0
		spd.OnWriteGameRecord(spd.Whoplay, "快速过牌了")
		spd.GameState = meta2.GsPlay
		spd.OnUserOutCard(&outmsg)
	} else if spd.TuoGuan > 0 && spd.GameState == meta2.GsRoarPai {
		if item == nil {
			return
		}
		if spd.Whoplay < uint16(spd.GetPlayerCount()) && !item.CheckTRUST() {
			spd.AutoTuoGuan(spd.Whoplay)
		}
		spd.OnRoarAction(spd.Whoplay, false)
	} else if spd.GameState == meta2.GsPlay {
		if item == nil {
			return
		}
		//托管或最后一手牌时要自动出牌
		LastHandOut := false
		tempCurPlay := spd.Whoplay

		//下家是否只有一张牌了
		bNextOnlyOne := (1 == spd.ThePaiCount[(int(spd.Whoplay)+1)%spd.GetPlayerCount()])

		//是否第一手牌
		isFirstOut := false
		if spd.CurCompleteCount == 1 && spd.Whoplay == spd.Banker && spd.WhoOutCount[spd.Whoplay] == 0 {
			isFirstOut = true
		}

		//构造数据
		var outmsg static.Msg_C_DG_OutCard
		outmsg.CurrentUser = spd.Whoplay
		if spd.WhoLastOut >= uint16(spd.GetPlayerCount()) {
			//智能一点的做法
			spd.m_GameLogic.GetGroupType(spd.PlayerCards[spd.Whoplay])
			_, beepOut := spd.m_GameLogic.BeepFirstCardOut(spd.PlayerCards[spd.Whoplay], bNextOnlyOne, isFirstOut, int(spd.ThePaiCount[spd.Whoplay]))
			for k := 0; k < len(beepOut[0].Indexes); k++ {
				outmsg.CardData[k] = beepOut[0].Indexes[k]
			}
			outmsg.CardCount = beepOut[0].Count
			outmsg.CardType = beepOut[0].Cardtype
			//最后一手时，没有炸弹时,也不是这个情况：3A是炸弹且正好有3个A
			if outmsg.CardCount == spd.ThePaiCount[spd.Whoplay] && spd.m_GameLogic.LastHandCanOut(spd.ThePaiCount[spd.Whoplay]) {
				LastHandOut = true
			}
		} else {
			outmsg.CardCount = 0
			if true {
				//如果跟出要出牌就需要用下面的牌
				spd.m_GameLogic.GetGroupType(spd.PlayerCards[spd.Whoplay])
				_, beepOut := spd.m_GameLogic.BeepCardOut(spd.AllPaiOut[spd.WhoLastOut], spd.LastOutType, bNextOnlyOne, int(spd.ThePaiCount[spd.Whoplay]))
				if len(beepOut) == 0 {
					outmsg.CardCount = 0
				} else {
					if spd.BiYa == 0 || (bNextOnlyOne && spd.LastOutType == TYPE_ONE) || int(spd.ThePaiCount[spd.Whoplay]) == len(beepOut[0].Indexes) {
						//必压 或能出完
						for k := 0; k < len(beepOut[0].Indexes); k++ {
							outmsg.CardData[k] = beepOut[0].Indexes[k]
						}
						outmsg.CardCount = beepOut[0].Count
						outmsg.CardType = beepOut[0].Cardtype
						//最后一手时，没有炸弹时,也不是这个情况：3A是炸弹且正好有3个A
						if outmsg.CardCount == spd.ThePaiCount[spd.Whoplay] && spd.m_GameLogic.LastHandCanOut(spd.ThePaiCount[spd.Whoplay]) {
							LastHandOut = true
						}
						if outmsg.CardCount == spd.ThePaiCount[spd.Whoplay] {
							if !LastHandOut {
								//先清理
								for k := 0; k < len(outmsg.CardData); k++ {
									outmsg.CardData[k] = 0
								}
								if spd.BiYa == 0 {
									//修改最终结果
									//必压时，是最后一手牌，但不全是炸弹，就出炸弹
									bombOut := spd.m_GameLogic.GetLastHandBomb()
									if bombOut.Point > 0 {
										for k := 0; k < len(bombOut.Indexes); k++ {
											outmsg.CardData[k] = bombOut.Indexes[k]
										}
										outmsg.CardCount = bombOut.Count
										outmsg.CardType = bombOut.Cardtype
									}
								} else {
									//修改最终结果
									//不是必压时，是最后一手牌，但不全是炸弹，就不出
									outmsg.CardCount = 0
								}
							}
						}
					} else {
						outmsg.CardCount = 0
					}
				}
			}
		}
		if spd.TuoGuan > 0 && spd.Whoplay < uint16(spd.GetPlayerCount()) && !item.CheckTRUST() && !LastHandOut {
			spd.AutoTuoGuan(spd.Whoplay)
		}
		if spd.TuoGuan > 0 || LastHandOut {
			//详细日志
			LogStr := fmt.Sprintf("托管出牌 OnAutoOperate UserID=%d ,CardCount=%d,牌数据:", outmsg.CurrentUser, outmsg.CardCount)
			for i := 0; i < MAXHANDCARD; i++ {
				if outmsg.CardData[i] > 0 {
					CardStr := fmt.Sprintf("0x%02x,", spd.switchCard2Ox(int(outmsg.CardData[i])))
					LogStr += CardStr
				}
			}
			spd.OnWriteGameRecord(spd.Whoplay, LogStr)
			spd.OnUserOutCard(&outmsg)

			spd.AutoCardCounts[tempCurPlay]++
			if spd.AutoCardCounts[tempCurPlay] >= 5 {
				// 连续5轮自动出牌，结束游戏
				//spd.OnGameEndUserLeft(tempCurPlay, gameserver.GOT_TUOGUAN)
			}
		}
	}
}

func (spd *SportPDK) AutoTuoGuan(theSeat uint16) int {
	item := spd.GetUserItemByChair(theSeat)
	if item == nil {
		return 0
	}
	var msgtg static.Msg_S_DG_Trustee
	msgtg.Trustee = true
	msgtg.ChairID = theSeat

	if spd.GameState == meta2.GsNull || theSeat >= uint16(spd.GetPlayerCount()) {
		//游戏状态 GameState，1表示吼牌阶段，2表示打牌阶段
		return 0
	}
	//详细日志
	LogStr := fmt.Sprintf("超时托管,CMD_S_Tuoguan_CB AutoTuoGuan msgtg.theFlag=%t msgtg.theSeat=%d ", msgtg.Trustee, msgtg.ChairID)
	spd.OnWriteGameRecord(theSeat, LogStr)

	if spd.GameState == meta2.GsPlay || spd.GameState == meta2.GsQuickPass || spd.GameState == meta2.GsRoarPai {
		if true == msgtg.Trustee {
			spd.TuoGuanPlayer[theSeat] = true
			item.ChangeTRUST(true)
			spd.TrustCounts[theSeat]++
			if theSeat == spd.Whoplay { //如果是当前的玩家，那么重新设置一下开始时间
				//spd.setLimitedTime(int64(spd.AutoOutTime))//已经超时了，马上就要切换牌权了，不用在设置他的时间了
			}
			spd.SendTableMsg(consts.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			spd.addReplayOrder(msgtg.ChairID, meta2.DG_REPLAY_OPT_TUOGUAN, 1, []int{})
		} else {
			spd.TuoGuanPlayer[theSeat] = false
			item.ChangeTRUST(false)
			spd.SendTableMsg(consts.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			spd.addReplayOrder(msgtg.ChairID, meta2.DG_REPLAY_OPT_TUOGUAN, 0, []int{})
		}
	}
	return 1
}

// 硬牌动作响应
func (spd *SportPDK) OnRoarAction(seat uint16, bRoar bool) bool {
	if seat < 0 || seat >= uint16(spd.GetPlayerCount()) {
		return false
	}
	if seat != spd.Whoplay {
		return false
	}

	if spd.WhoReady[seat] {
		return false
	}

	spd.WhoReady[seat] = true

	//变量定义
	var roar static.Msg_S_DG_Roar
	roar.CurrentUser = seat
	roar.RoarFlag = 0
	if bRoar {
		roar.RoarFlag = 1
	}

	spd.SendTableMsg(consts.MsgTypeRoar, roar)

	//回放增加吼牌记录
	spd.addReplayOrder(seat, meta2.DG_REPLAY_OPT_HOUPAI, int(roar.RoarFlag), []int{})

	if bRoar { //如果硬牌，那么结束硬牌动作
		spd.EndRoar(true)
		return true
	} else {
		z := 0
		for i := 0; i < spd.GetPlayerCount(); i++ {
			if spd.WhoReady[i] {
				z++
			}
		}
		if z >= spd.GetPlayerCount() { //开始游戏
			spd.EndRoar(false)
		} else {
			spd.GoNextPlayer()
		}
	}
	return true
}

func (spd *SportPDK) StartPlay(whoplay uint16) {
	spd.CardScore = 0
	// 开始
	spd.GameState = meta2.GsPlay

	iPower := 2
	spd.SendPower(whoplay, iPower, spd.PlayTime)
}

func (spd *SportPDK) StartRoar(theSeat uint16) {
	// 开始进入吼牌状态
	spd.GameState = meta2.GsRoarPai
	iPower := 1
	spd.SendPower(theSeat, iPower, spd.RoarTime)
}

func (spd *SportPDK) EndRoar(bRoar bool) {
	var endroarmsg static.Msg_S_DG_EndRoar

	//有人吼牌
	if bRoar {
		spd.WhoRoar = spd.Whoplay
		endroarmsg.RoarUser = spd.WhoRoar
		spd.GameType = meta2.GT_ROAR
	} else {
		//没人吼牌
		spd.GameType = meta2.GT_NORMAL
		spd.WhoRoar = static.INVALID_CHAIR
		endroarmsg.RoarUser = spd.WhoRoar
		spd.GetJiaoPai() //得到叫牌
	}
	//吼牌的为庄家了
	if bRoar {
		spd.Banker = spd.WhoRoar
		spd.BankParter = static.INVALID_CHAIR
		//详细日志
		LogStr := string("包牌(吼牌)")
		spd.OnWriteGameRecord(spd.WhoRoar, LogStr)
	}
	spd.Whoplay = spd.Banker
	endroarmsg.BankUser = spd.Banker
	endroarmsg.JiaoPai = spd.RoarPai

	spd.SendTableMsg(consts.MsgTypeEndRoar, endroarmsg)

	//回放增加结束吼牌记录
	spd.addReplayOrder(spd.Banker, meta2.DG_REPLAY_OPT_END_HOUPAI, int(spd.RoarPai), []int{})

	//详细日志
	LogStr := string("为庄家")
	spd.OnWriteGameRecord(spd.Whoplay, LogStr)

	spd.StartPlay(spd.Whoplay)
}

func (spd *SportPDK) GoNextPlayer() {
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	for iPlayer := 0; iPlayer < spd.GetPlayerCount(); iPlayer++ {
		if spd.Whoplay >= uint16(spd.GetPlayerCount())-1 {
			spd.Whoplay = 0
		} else {
			spd.Whoplay++
		}

		//如果当前玩家出完了
		if spd.WhoAllOutted[spd.Whoplay] {
			if spd.WhoLastOut == spd.Whoplay { //这个玩家是不是上一次出牌玩家
				break
			} else {
				continue
			}
		} else { //没出完？
			break
		}
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////
	if spd.WhoLastOut == spd.Whoplay {
		if spd.EndTurn() {
			return
		}

	}

	//是否管不起
	quickPass := false
	if spd.GameState == meta2.GsPlay && spd.WhoLastOut < uint16(spd.GetPlayerCount()) {
		//下家是否只有一张牌了
		bNextOnlyOne := (1 == spd.ThePaiCount[(int(spd.Whoplay)+1)%spd.GetPlayerCount()])
		spd.m_GameLogic.GetGroupType(spd.PlayerCards[spd.Whoplay])
		_, beepOut := spd.m_GameLogic.BeepCardOut(spd.AllPaiOut[spd.WhoLastOut], spd.LastOutType, bNextOnlyOne, int(spd.ThePaiCount[spd.Whoplay]))
		if len(beepOut) == 0 {
			quickPass = true
		}
	}

	if spd.GameState == meta2.GsRoarPai {
		spd.StartRoar(spd.Whoplay)
	} else if spd.GameState == meta2.GsPlay {
		dwPower := 2
		if quickPass && !spd.QuickPass {
			spd.GameState = meta2.GsQuickPass
			spd.SendPower(spd.Whoplay, 5, 5)
		} else {
			spd.SendPower(spd.Whoplay, dwPower, spd.PlayTime)
		}
	}

	for i := 0; i < MAXHANDCARD; i++ {
		spd.AllPaiOut[spd.Whoplay][i] = 0
	}

	//自动过牌
	if quickPass && spd.QuickPass {
		spd.GameState = meta2.GsQuickPass
		spd.setLimitedTime(1) //延时1秒，否则看不到首出的牌
	} else {
		//判断是否是最后一手
		if spd.IsLastHandOut(false) {
			spd.GameState = meta2.GsPlay
			spd.setLimitedTime(1) //延时2秒，否则看不到首出的牌
		}
	}
}

// 不管是不是必压，最后一手不含炸弹，且能大过别人，就要立即出完。
func (spd *SportPDK) IsLastHandOut(bNextOnlyOne bool) (LastHandOut bool) {
	LastHandOut = false
	if spd.WhoLastOut >= uint16(spd.GetPlayerCount()) {
		//智能一点的做法
		//是否第一手牌
		isFirstOut := false
		if spd.CurCompleteCount == 1 && spd.Whoplay == spd.Banker && spd.WhoOutCount[spd.Whoplay] == 0 {
			isFirstOut = true
		}
		spd.m_GameLogic.GetGroupType(spd.PlayerCards[spd.Whoplay])
		_, beepOut := spd.m_GameLogic.BeepFirstCardOut(spd.PlayerCards[spd.Whoplay], bNextOnlyOne, isFirstOut, int(spd.ThePaiCount[spd.Whoplay]))
		//最后一手时，没有炸弹时,也不是这个情况：3A是炸弹且正好有3个A
		if beepOut[0].Count == spd.ThePaiCount[spd.Whoplay] && spd.m_GameLogic.LastHandCanOut(spd.ThePaiCount[spd.Whoplay]) {
			LastHandOut = true
		}
	} else {
		if true {
			//如果跟出要出牌就需要用下面的牌
			spd.m_GameLogic.GetGroupType(spd.PlayerCards[spd.Whoplay])
			_, beepOut := spd.m_GameLogic.BeepCardOut(spd.AllPaiOut[spd.WhoLastOut], spd.LastOutType, bNextOnlyOne, int(spd.ThePaiCount[spd.Whoplay]))
			if len(beepOut) == 0 {
				LastHandOut = false //
			} else {
				if int(spd.ThePaiCount[spd.Whoplay]) == len(beepOut[0].Indexes) {
					//最后一手时，没有炸弹时,也不是这个情况：3A是炸弹且正好有3个A
					if beepOut[0].Count == spd.ThePaiCount[spd.Whoplay] && spd.m_GameLogic.LastHandCanOut(spd.ThePaiCount[spd.Whoplay]) {
						LastHandOut = true
					}
				} else {
					LastHandOut = false //
				}
			}
		}
	}
	return LastHandOut
}

// 得到叫牌
func (spd *SportPDK) GetJiaoPai() byte {
	temp := [static.MAX_CARD_4P]byte{}
	count := 0
	for i := 1; i < CARDINDEX_BACK; i++ {
		num := 0
		for j := 0; j < MAXHANDCARD; j++ {
			if spd.PlayerCards[spd.Banker][j] == byte(i) {
				num++
			}
		}
		if num == 1 {
			temp[count] = byte(i)
			count++
		}
	}

	seed := rand.Intn(10000) % count
	spd.RoarPai = temp[seed]
	if spd.DownPai != 0 {
		spd.RoarPai = spd.DownPai
	}

	for i := 0; i < spd.GetPlayerCount(); i++ {
		if uint16(i) == spd.Banker {
			continue
		}
		for j := 0; j < MAXHANDCARD; j++ {
			if spd.PlayerCards[i][j] == spd.RoarPai {
				spd.BankParter = uint16(i) //明鸡
				break
			}
		}
		if spd.BankParter < uint16(spd.GetPlayerCount()) {
			break
		}
	}
	//详细日志
	LogStr := fmt.Sprintf("明鸡牌为:0x%02x ", spd.switchCard2Ox(int(spd.RoarPai)))
	spd.OnWriteGameRecord(spd.Banker, LogStr)
	return spd.RoarPai
}

// 结束一轮
func (spd *SportPDK) EndTurn() bool {
	for i := 0; i < MAXHANDCARD; i++ {
		spd.LastPaiOut[spd.WhoLastOut][i] = spd.AllPaiOut[spd.WhoLastOut][i]
	}
	for i := 0; i < spd.GetPlayerCount(); i++ {
		for j := 0; j < MAXHANDCARD; j++ {
			spd.AllPaiOut[i][j] = 0
		}
		spd.WhoPass[i] = false
	}
	//最后是否是炸弹
	if spd.BombMode {
		if spd.LastOutType == TYPE_BOMB_NOMORL {
			spd.ValidBombCount[spd.WhoLastOut]++
			//若炸弹实时算分
			if spd.BombRealTime {
				spd.BombRealTimeScore(spd.WhoLastOut, 10)
			}
		}
	}
	//打分的模式
	if spd.GameType == meta2.GT_NORMAL {
		//回放增加本轮抓分
		if spd.CardScore > 0 {
			spd.addReplayOrder(spd.WhoLastOut, meta2.DG_REPLAY_OPT_TURN_OVER, spd.CardScore, []int{})
		} else {
			spd.addReplayOrder(spd.WhoLastOut, meta2.DG_REPLAY_OPT_TURN_OVER, 0, []int{})
		}
		spd.PlayerCardScore[spd.WhoLastOut] += spd.CardScore
		spd.SendPlayerScore(uint16(spd.GetPlayerCount()), spd.WhoLastOut, spd.CardScore)
		spd.CardScore = 0 //清零
		spd.SendTurnScore(uint16(spd.GetPlayerCount()))
	} else {
		spd.addReplayOrder(spd.WhoLastOut, meta2.DG_REPLAY_OPT_TURN_OVER, 0, []int{})
	}

	spd.WhoLastOut = static.INVALID_CHAIR

	spd.EndOut() //结束一轮
	return false
}
func (spd *SportPDK) EndOut() {
	spd.LastOutType = static.TYPE_NULL
	spd.LastOutTypeClient = static.TYPE_NULL

	var endout static.Msg_S_DG_EndOut
	endout.CurrentUser = spd.Whoplay
	spd.SendTableMsg(consts.MsgTypeEndOut, endout)
}

func (spd *SportPDK) GetTeamer(who uint16) uint16 {
	re := uint16(0)
	if who >= static.MAX_PLAYER_3P {
		return 0
	}
	if who == spd.Banker {
		re = spd.BankParter
	} else if who == spd.BankParter {
		re = spd.Banker
	} else { //闲家
		for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
			if i == who {
				continue
			}
			if i == spd.BankParter {
				continue
			}
			if i == spd.Banker {
				continue
			}
			re = i
			break
		}
	}
	return re
}
func (spd *SportPDK) GetTeamScore(seat uint16) int {
	if spd.GameType != meta2.GT_NORMAL {
		return 0
	}
	if seat < 0 || seat >= static.MAX_PLAYER_3P {
		return 0
	}
	teamer := spd.GetTeamer(seat)
	if spd.PlayerTurn[seat] == 1 || spd.PlayerTurn[teamer] == 1 { //一游玩家的分
		return spd.PlayerCardScore[seat] + spd.PlayerCardScore[teamer]
	} else if spd.PlayerTurn[seat] == 2 || spd.PlayerTurn[teamer] == 2 {
		who2you := teamer
		if spd.PlayerTurn[seat] == 2 {
			who2you = seat
		}
		return spd.PlayerCardScore[who2you]
	}
	return 0
}

func (spd *SportPDK) GetFinalWinLoseScore(score *[MAX_PLAYER]int) {
	wintotal := 0
	losetotal := 0
	total := [MAX_PLAYER]int{}
	n := 0
	for i := 0; i < spd.GetPlayerCount(); i++ {
		total[i] = score[i]
		if score[i] <= 0 {
			losetotal += score[i]
		} else {
			n++
			wintotal += score[i]
		}
	}
	if n <= 0 {
		return
	}
	if wintotal > -losetotal { //扣的分比加的分少，那么只能平分了
		for i := 0; i < static.MAX_PLAYER_3P; i++ {
			if score[i] > 0 {
				total[i] = -int(float64(losetotal) * float64(score[i]) / (float64(wintotal)))
			}
		}
	}
	for i := 0; i < static.MAX_PLAYER_3P; i++ {
		score[i] = total[i]
	}
}

func (spd *SportPDK) AddSpecailScore(Score *[MAX_PLAYER]int, seat uint16, base int) {
	if seat < 0 && seat >= uint16(spd.GetPlayerCount()) {
		return
	}
	spd.OnWriteGameRecord(static.INVALID_CHAIR, "跑得快不计算喜钱")
	return

	iAddFan := spd.WhoSame510K[seat] + spd.Who7Xi[seat] + spd.Who8Xi[seat] + spd.PlayKingBomb[seat]
	iTempScore := iAddFan * base

	//这里来计算各自该赢的，特殊的分
	for i := 0; i < static.MAX_PLAYER_3P; i++ {
		if uint16(i) == seat {
			spd.XiScore[i] += 3 * iTempScore
			Score[i] += 3 * iTempScore
		} else {
			spd.XiScore[i] -= iTempScore //其他人
			Score[i] -= iTempScore
		}
	}
}

// ! 加载测试麻将数据
func (spd *SportPDK) initDebugCards(configName string, cbRepertoryCard *[MAX_PLAYER][static.MAX_CARD]byte, wBankerUser *uint16, byDownPai *byte) (err error) {
	defer func() {
		if err != nil {
			spd.OnWriteGameRecord(static.INVALID_CHAIR, err.Error())
		}
	}()
	//! 做牌文件配置
	var debugCardConfig *meta2.CardConfig = new(meta2.CardConfig)

	configName = fmt.Sprintf("./%s%d", configName, spd.GetPlayerCount())
	spd.OnWriteGameRecord(static.INVALID_CHAIR, "开始读取做牌文件，文件名："+configName)
	if !static.GetJsonMgr().ReadData("./json", configName, debugCardConfig) {
		return errors.New("做牌文件:读取失败")
	}

	cardNum := 16
	if spd.Rule.NoWan {
		cardNum = 15
	}
	// 是否开启做牌
	if debugCardConfig.IsAble == 1 {
		//检查做牌文件是否做牌异常
		for _, handCards := range debugCardConfig.UserCards {
			if len(strings.Split(handCards, ",")) != cardNum {
				return errors.New(fmt.Sprintf("做牌文件:手牌长度不为%d", cardNum))
			}
		}
		////检查牌堆牌是否正常
		//if len(debugCardConfig.RepertoryCard) != debugCardConfig.RepertoryCardCount*5-1 {
		//	return errors.New(fmt.Sprintf("做牌文件:牌库牌数量不一致:::RepertoryCard:[%d]>>>实际做牌牌库数量:[%d]", debugCardConfig.RepertoryCardCount*5-1, len(debugCardConfig.RepertoryCard)))
		//}
		// 设置玩家手牌

		for userIndex, handCards := range debugCardConfig.UserCards {
			byCardsCount := byte(0)
			_item := spd.GetUserItemByChair(uint16(userIndex))
			if _item != nil {
				_item.Ctx.InitCardIndex() //清理手牌
				handCardsSlice := strings.Split(handCards, ",")
				for _, cardStr := range handCardsSlice {

					if _, cardValue := spd.GetCardDataByStr(cardStr); cardValue == static.INVALID_BYTE {
						//return errors.New(fmt.Sprintf("做牌文件:第%d号玩家做牌异常：%s", userIndex, cardStr))
					} else {
						//_item.Ctx.DispatchCard(cardValue)
						(*cbRepertoryCard)[userIndex][byCardsCount] = cardValue
						byCardsCount++
						if byCardsCount >= MAXHANDCARD {
							break
						}
					}
				}
			}
		}
		//设置牌堆牌
		repertoryCards := strings.Split(debugCardConfig.RepertoryCard, ",")
		for _, cardStr := range repertoryCards {
			if _, cardValue := spd.GetCardDataByStr(cardStr); cardValue == static.INVALID_BYTE {
				//return errors.New(fmt.Sprintf("做牌文件:牌堆第%d个做牌异常：%s", cardIndex, cardStr))
			} else {
				(*byDownPai) = cardValue
				break //底牌里面的叫牌只能有一个
			}
		}
		// 设置庄家
		if debugCardConfig.BankerUserSeatId != -1 {
			(*wBankerUser) = uint16(debugCardConfig.BankerUserSeatId)
		}
	} else {
		return errors.New("做牌文件:开关未开启")
	}
	return err
}

// 写日志记录
func (spd *SportPDK) writeGameRecord() {
	//写日志记录
	spd.OnWriteGameRecord(static.INVALID_CHAIR, "开始重阳麻将  发牌......")

	// 玩家手牌
	//for _, v := range spd.PlayerInfo {
	//	if v.Seat != public.INVALID_CHAIR {
	//		handCardStr := fmt.Sprintf("发牌后手牌:%s", spd.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
	//		spd.OnWriteGameRecord(uint16(v.Seat), handCardStr)
	//	}
	//}

	// 牌堆牌
	//leftCardStr := fmt.Sprintf("牌堆牌:%s", spd.m_GameLogic.SwitchToCardNameByDatas(spd.RepertoryCard[0:spd.LeftCardCount+2], 0))
	//spd.OnWriteGameRecord(public.INVALID_CHAIR, leftCardStr)

	//赖子牌
	//magicCardStr := fmt.Sprintf("癞子牌:%s", spd.m_GameLogic.SwitchToCardNameByData(spd.MagicCard, 1))
	//spd.OnWriteGameRecord(public.INVALID_CHAIR, magicCardStr)
}

// ! 解散
func (spd *SportPDK) OnEnd() {
	if spd.IsGameStarted() {
		spd.OnGameOver(static.INVALID_CHAIR, static.GER_DISMISS)
	}
}

// ! 单局结算
func (spd *SportPDK) OnGameOver(wChairID uint16, cbReason byte) bool {
	spd.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 初始化游戏
func (spd *SportPDK) OnInit(table base2.TableBase) {
	spd.KIND_ID = table.GetTableInfo().KindId
	spd.Config.StartMode = static.StartMode_FullReady
	spd.Config.PlayerCount = 4 //玩家人数
	spd.Config.ChairCount = 4  //椅子数量
	spd.PlayerInfo = make(map[int64]*components2.Player)

	spd.RepositTable(true)
	spd.SetGameStartMode(static.StartMode_FullReady)
	spd.GameTable = table
	spd.Init()
	spd.Unmarsha(table.GetTableInfo().GameInfo)
	table.GetTableInfo().GameInfo = ""

	//设置自动解散时间2分钟
	spd.SetDismissRoomTime(120)
	//设置离线解散时间30分钟
	spd.SetOfflineRoomTime(1800)
	//进入防沉迷120秒解散房间
	spd.SetVitaminLowPauseTime(10)

	//if spd.GameTable.GetTableInfo().JoinType==constant.NoCheat || spd.GameTable.GetTableInfo().JoinType==constant.AutoAdd{
	//	spd.SetOfflineRoomTime(30)
	//}

	var _msg rule2.FriendRuleDG_qc
	if err := json.Unmarshal(static.HF_Atobytes(table.GetTableInfo().Config.GameConfig), &_msg); err == nil {
		if _msg.FleeTime != 0 {
			spd.SetOfflineRoomTime(_msg.FleeTime)
		}
	}
}

// ! 发送消息
func (spd *SportPDK) OnMsg(msg *base2.TableMsg) bool {
	wChairID := spd.GetChairByUid(msg.Uid)
	if wChairID == static.INVALID_CHAIR {
		spd.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::找不到玩家的座位号:%d", msg.Uid))
		return true
	}

	switch msg.Head {
	case consts.MsgTypeGameBalanceGameReq: //! 请求总结算信息 //暂时没有

		var _msg static.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spd.CalculateResultTotal_Rep(&_msg)
		}
	case consts.MsgTypeGameOutCard: //! 出牌消息
		var _msg static.Msg_C_DG_OutCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			//详细日志
			LogStr := fmt.Sprintf("OnUserOutCard UserID=%d ,CardCount=%d,牌数据:", _msg.CurrentUser, _msg.CardCount)
			for i := 0; i < MAXHANDCARD; i++ {
				if _msg.CardData[i] > 0 {
					CardStr := fmt.Sprintf("0x%02x,", spd.switchCard2Ox(int(_msg.CardData[i])))
					LogStr += CardStr
				}
			}
			spd.OnWriteGameRecord(_msg.CurrentUser, LogStr)
			outcardflag := spd.OnUserOutCard(&_msg)
			return outcardflag
		}
	case consts.MsgTypeGameTrustee: //用户托管
		var _msg static.Msg_C_Trustee
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spd.onUserTustee(&_msg)
			//详细日志
			LogStr := fmt.Sprintf("主动托管动作(true托管,false取消):TrustFlag=%t ", _msg.Trustee)
			spd.OnWriteGameRecord(spd.GetChairByUid(_msg.Id), LogStr)
		}
	case consts.MsgTypeRoar: //用户吼牌动作
		var _msg static.Msg_C_DG_Roar
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			bRoarFlag := false
			if _msg.RoarFlag == 1 {
				bRoarFlag = true
			}
			//详细日志
			LogStr := fmt.Sprintf("吼牌动作:theFlag=%d ", _msg.RoarFlag)
			spd.OnWriteGameRecord(_msg.CurrentUser, LogStr)
			return spd.OnRoarAction(_msg.CurrentUser, bRoarFlag)
		}
	case consts.MsgTypeGameGoOnNextGame: //下一局
		//详细日志
		LogStr := string("OnUserClientNextGame!!! ")
		spd.OnWriteGameRecord(spd.Whoplay, LogStr)

		var _msg static.Msg_C_GoOnNextGame
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spd.OnUserClientNextGame(&_msg)
		}
	case consts.MsgTypeSplitCard: //玩家切牌
		//详细日志
		LogStr := string("OnUserEndSplitCards!!! ")
		spd.OnWriteGameRecord(spd.Whoplay, LogStr)

		var _msg static.Msg_C_DG_SplitCards
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spd.OnUserEndSplitCards(&_msg)
		}
	case consts.MsgTypeNeedSplitCard: //需要玩家切牌
		//详细日志
		LogStr := string("切牌选择!!! ")
		spd.OnWriteGameRecord(spd.Whoplay, LogStr)

		var _msg static.Msg_C_DG_NeedSplitCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			var _msg_ static.Msg_C_GoOnNextGame
			_msg_.Id = _msg.Id
			if _msg.NeedSplitFlag {
				//详细日志
				LogStr := string("小结算时选择了 切牌!!! ")
				spd.OnWriteGameRecord(spd.Whoplay, LogStr)
				spd.OnUserSplitAndNextGame(&_msg_, _msg.NeedSplitFlag)
			} else {
				if spd.GameState == meta2.GsSplitCards {
					//详细日志
					LogStr := string("游戏开始时选择了 不切牌!!! ")
					spd.OnWriteGameRecord(spd.Whoplay, LogStr)
					spd.StartNextGame() //直接开始
				}
			}
		}
	case consts.MsgTypeGameDismissFriendResult: //申请解散玩家选择
		if spd.GameEndStatus == byte(static.GS_FREE) {
			var _msg static.Msg_C_DismissFriendResult
			if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
				if _msg.Flag == false {
					//不同意解散,托管玩家自动准备
					//是否托管
					for i := 0; i < static.MAX_PLAYER_4P; i++ {
						item := spd.GetUserItemByChair(uint16(i))
						if item == nil || !item.CheckTRUST() {
							continue
						}
						spd.AutoNextGame(item.Uid)
					}
				}
			}
		}
	default:
		//spd.Common.OnMsg(msg)
	}

	return true
}

// 下一局
func (spd *SportPDK) OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(spd.CurCompleteCount) >= spd.Rule.JuShu || spd.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}

	// 记录游戏开始时间
	spd.Common.GameBeginTime = time.Now()

	nChiarID := spd.GetChairByUid(msg.Id)
	if nChiarID >= 0 && nChiarID < uint16(spd.GetPlayerCount()) {
		_item := spd.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}
	//将该消息广播出去。游戏开始后，不用广播
	if spd.GameEndStatus != static.GS_MJ_PLAY {
		spd.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
		spd.SendUserStatus(int(nChiarID), static.US_READY) //把我的状态发给其他人
	}

	//SEND_TABLE_DATA(INVALID_CHAIR,SUB_C_GOON_NEXT_GAME,pDataBuffer);

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < spd.GetPlayerCount(); i++ {
		item := spd.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			break
		}
		if i == spd.GetPlayerCount()-1 {
			spd.RepositTable(false) // 复位桌子
			spd.CurCompleteCount++
			spd.GetTable().SetBegin(true)
			if spd.BSplited {
				spd.SplitAndGameStart()
			} else {
				spd.OnGameStart()
			}
		}
	}
	return true
}

// 下一局
func (spd *SportPDK) OnUserSplitAndNextGame(msg *static.Msg_C_GoOnNextGame, flag bool) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(spd.CurCompleteCount) >= spd.Rule.JuShu || spd.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}

	// 记录游戏开始时间
	spd.Common.GameBeginTime = time.Now()

	nChiarID := spd.GetChairByUid(msg.Id)
	//将该消息广播出去。游戏开始后，不用广播
	if spd.GameEndStatus != static.GS_MJ_PLAY {
		spd.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
		spd.SendUserStatus(int(nChiarID), static.US_READY) //把我的状态发给其他人
	}

	spd.BSplited = true

	if nChiarID >= 0 && nChiarID < uint16(spd.GetPlayerCount()) {
		_item := spd.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < spd.GetPlayerCount(); i++ {
		item := spd.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			break
		}
		if i == spd.GetPlayerCount()-1 {
			spd.RepositTable(false) // 复位桌子
			spd.CurCompleteCount++
			spd.GetTable().SetBegin(true)
			spd.SplitAndGameStart()
		}
	}
	return true
}

// 托管
func (spd *SportPDK) onUserTustee(msg *static.Msg_C_Trustee) bool {
	item := spd.GetUserItemByUid(msg.Id)
	if item == nil {
		return false
	}
	if item.CheckTRUST() == msg.Trustee {
		return true
	}
	//变量定义
	var tuoguan static.Msg_S_DG_Trustee
	tuoguan.ChairID = spd.GetChairByUid(msg.Id)
	tuoguan.Trustee = msg.Trustee
	//校验规则
	if tuoguan.ChairID < uint16(spd.GetPlayerCount()) && ((spd.GameState == meta2.GsPlay) || (spd.GameState == meta2.GsQuickPass) || (spd.GameState == meta2.GsRoarPai) || (spd.GameState == meta2.GsNull)) {
		if tuoguan.Trustee == true && (spd.GameState != meta2.GsNull) {
			spd.TuoGuanPlayer[tuoguan.ChairID] = true
			item.ChangeTRUST(true)
			spd.TrustCounts[tuoguan.ChairID]++
			if tuoguan.ChairID == spd.Whoplay {
				//spd.m_tDownTime = GetCPUTickCount()+spd.AutoOutTime;
				if int64(spd.LimitTime) > time.Now().Unix() {
					tuoguan.Overtime = int64(spd.LimitTime) - time.Now().Unix()
				}
				if time.Now().Unix()+int64(spd.AutoOutTime) < spd.LimitTime { //如果只剩下托管出牌的时间了，就不重新算了，否则跟改为托管出牌的时间
					spd.setLimitedTime(int64(spd.AutoOutTime))
				}
			}
			spd.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//回放托管记录
			spd.addReplayOrder(tuoguan.ChairID, meta2.DG_REPLAY_OPT_TUOGUAN, 1, []int{})
		} else if tuoguan.Trustee == false {
			spd.TuoGuanPlayer[tuoguan.ChairID] = false
			item.ChangeTRUST(false)

			//如果是当前的玩家，那么重新设置一下开始时间
			if tuoguan.ChairID == spd.Whoplay {
				//spd.m_tDownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
				if time.Now().Unix() < spd.LimitTime { //如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < spd.LimitTime
					playTimeTemp := spd.PlayTime
					if spd.GameState == meta2.GsQuickPass && !spd.QuickPass {
						playTimeTemp = 5
					}
					spd.setLimitedTime(int64(playTimeTemp) + spd.PowerStartTime - time.Now().Unix() + 1)
					tuoguan.Overtime = int64(playTimeTemp) + spd.PowerStartTime - time.Now().Unix()
				}
			}

			//tuoguan.theTime = PlayTime-(now-nowTime);
			spd.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//回放托管记录
			spd.addReplayOrder(tuoguan.ChairID, meta2.DG_REPLAY_OPT_TUOGUAN, 0, []int{})
		} else {
			return false
		}
	} else {
		//详细日志
		LogStr := string("托管动作:游戏状态不正确 ")
		spd.OnWriteGameRecord(tuoguan.ChairID, LogStr)
		return false
	}
	return true
}

func (spd *SportPDK) addReplayOrder(chairId uint16, operation int, value int, values []int) {
	var order meta2.DG_Replay_Order
	order.R_ChairId = chairId
	order.R_Opt = operation

	if operation == meta2.DG_REPLAY_OPT_HOUPAI {
		var order_ext meta2.DG_Replay_Order_Ext
		order_ext.Ext_type = meta2.DG_EXT_HOUPAI
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == meta2.DG_REPLAY_OPT_END_HOUPAI {
		order.R_Value = append(order.R_Value, value)
	} else if operation == meta2.DG_REPLAY_OPT_OUTCARD {
		order.R_Value = append(order.R_Value, values[:]...)
	} else if operation == meta2.DG_REPLAY_OPT_END_GAME {

	} else if operation == meta2.DG_REPLAY_OPT_DIS_GAME {

	} else if operation == meta2.DG_REPLAY_OPT_TURN_OVER {
		var order_ext meta2.DG_Replay_Order_Ext
		order_ext.Ext_type = meta2.DG_EXT_GETSCORE
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == meta2.DG_REPLAY_OPT_TUOGUAN {
		var order_ext meta2.DG_Replay_Order_Ext
		order_ext.Ext_type = meta2.DG_EXT_TUOGUAN
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	}

	spd.ReplayRecord.R_Orders = append(spd.ReplayRecord.R_Orders, order)
}

// 不是牌不足的情况，单独的出3张、3带1 和飞机不带满的情况都不允许
func (spd *SportPDK) CardNotEnough(byWhoplay uint16, iTheCardType int, byOutCardStrCount byte, byOutCardCount byte) bool {
	if int(byWhoplay) > spd.GetPlayerCount() {
		return false
	}
	if iTheCardType == TYPE_THREE_TAKE_TWO && byOutCardStrCount*5 > byOutCardCount && byOutCardCount != spd.ThePaiCount[byWhoplay] {
		return false
	}
	if iTheCardType == TYPE_FOUR_TAKE_THREE && byOutCardStrCount*7 > byOutCardCount && byOutCardCount != spd.ThePaiCount[byWhoplay] {
		return false
	}
	return true
}

// 用户出牌
func (spd *SportPDK) OnUserOutCard(msg *static.Msg_C_DG_OutCard) bool {
	xlog.Logger().Info("OnUserOutCard")
	//效验状态
	if spd.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}
	if spd.GameEndStatus != static.GS_MJ_PLAY {
		return true
	}
	if spd.GameState != meta2.GsPlay && spd.GameState != meta2.GsQuickPass {
		return false
	}

	wChairID := msg.CurrentUser
	//效验参数
	if wChairID != spd.Whoplay {
		//详细日志
		LogStr := fmt.Sprintf("座位号 %d, OnUserOutCard 出牌玩家不是当前玩家 ", wChairID)
		spd.OnWriteGameRecord(wChairID, LogStr)
		return false
	}
	_userItem := spd.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	//玩家5s内点不要，将快速过牌阶段置为出牌阶段
	spd.GameState = meta2.GsPlay

	//出牌数目为0，则为放弃的情况
	if msg.CardCount == 0 {
		if spd.WhoLastOut < uint16(spd.GetPlayerCount()) && spd.Whoplay == wChairID {
			//下家是否只有一张牌了
			bNextOnlyOne := (1 == spd.ThePaiCount[(int(spd.Whoplay)+1)%spd.GetPlayerCount()]) && spd.LastOutType == TYPE_ONE
			if spd.BiYa == 0 || bNextOnlyOne {
				//必压时 /下家只有一张牌时 必须压牌
				spd.m_GameLogic.GetGroupType(spd.PlayerCards[spd.Whoplay])
				_, beepOut := spd.m_GameLogic.BeepCardOut(spd.AllPaiOut[spd.WhoLastOut], spd.LastOutType, bNextOnlyOne, int(spd.ThePaiCount[spd.Whoplay]))
				if len(beepOut) > 0 {
					LogStr := fmt.Sprintf("座位号 %d, 非法牌型。必压时有大必压。下家只剩一张牌时出单牌有大必须压牌 ", wChairID)
					spd.OnWriteGameRecord(wChairID, LogStr)
					return false
				}
			}

			var outmsg static.Msg_S_DG_OutCard
			outmsg.CardCount = 0
			outmsg.CurrentUser = msg.CurrentUser
			for i := 0; i < MAXHANDCARD; i++ {
				outmsg.CardData[i] = 0
				spd.AllPaiOut[spd.Whoplay][i] = 0
				spd.LastPaiOut[spd.Whoplay][i] = 0
			}
			spd.SendTableMsg(consts.MsgTypeGameOutCard, outmsg)

			//回放增加出牌日志
			spd.addReplayOrder(spd.Whoplay, meta2.DG_REPLAY_OPT_OUTCARD, int(static.INVALID_BYTE), []int{})

			//详细日志
			LogStr := string("OnUserOutCard 玩家pass ")
			spd.OnWriteGameRecord(spd.Whoplay, LogStr)

			spd.WhoPass[spd.Whoplay] = true
			spd.GoNextPlayer()
		}

		return true
	}

	buf := [static.MAX_CARD_4P]byte{}
	for i := 0; i < MAXHANDCARD; i++ {
		buf[i] = spd.PlayerCards[spd.Whoplay][i]
	}

	z := 0 //重新检验牌数目，并临时删除出的牌
	for i := byte(0); i < msg.CardCount; i++ {
		for j := 0; j < MAXHANDCARD; j++ {
			if msg.CardData[i] == buf[j] {
				buf[j] = 0
				z++
				break
			}
		}
	}
	if spd.BombSplit {
		for ij := 0; ij < int(msg.CardCount); ij++ {
			if 4 == spd.m_GameLogic.GetOneCardNum(spd.PlayerCards[spd.Whoplay], MAXHANDCARD, msg.CardData[ij]) {
				if 4 != spd.m_GameLogic.GetOneCardNum(msg.CardData, msg.CardCount, msg.CardData[ij]) {
					//详细日志
					LogStr := string("炸弹不可拆 ")
					spd.OnWriteGameRecord(spd.Whoplay, LogStr)
					spd.SendGameNotificationMessage(spd.Whoplay, "炸弹不可拆")
					return true
				}
			}
		}
	}
	if spd.Rule.MingTing {
		if spd.GetPlayerCount() == 3 && spd.CurCompleteCount == 1 && spd.Whoplay == spd.Banker && spd.WhoOutCount[spd.Whoplay] == 0 {
			//黑桃3必须先出
			if spd.m_GameLogic.ISHaveSpade3(spd.PlayerCards[spd.Whoplay][:]) {
				if !spd.m_GameLogic.ISHaveSpade3(msg.CardData[:]) {
					spd.OnWriteGameRecord(spd.Whoplay, "黑桃3必须出")
					spd.SendGameNotificationMessage(spd.Whoplay, "黑桃3必须先出")
					return true
				}
			}
		} else if spd.GetPlayerCount() == 2 && spd.CurCompleteCount == 1 && spd.Whoplay == spd.Banker && spd.WhoOutCount[spd.Whoplay] == 0 {
			//黑桃3必须先出
			if spd.m_GameLogic.ISHaveSpade3(spd.PlayerCards[spd.Whoplay][:]) {
				if !spd.m_GameLogic.ISHaveSpade3(msg.CardData[:]) {
					spd.OnWriteGameRecord(spd.Whoplay, "黑桃3必须出")
					spd.SendGameNotificationMessage(spd.Whoplay, "黑桃3必须先出")
					return true
				}
			} else {
				//最小牌
				min := spd.m_GameLogic.GetMinCard(spd.PlayerCards[spd.Whoplay][:])
				if min != spd.m_GameLogic.GetMinCard(msg.CardData[:]) {
					spd.OnWriteGameRecord(spd.Whoplay, "首出请出最小牌")
					spd.SendGameNotificationMessage(spd.Whoplay, "首出请出最小牌")
					return true
				}
			}
		}
	}
	if byte(z) == msg.CardCount {
		iNumOfKing := 0
		for i := 0; i < MAXHANDCARD; i++ {
			spd.AllPaiOut[spd.Whoplay][i] = msg.CardData[i]
		}

		lastOutCards := [static.MAX_CARD]byte{}
		laseOutLen := byte(0)
		//TYPE_NULL代表第一个出
		if spd.WhoLastOut >= uint16(spd.GetPlayerCount()) {
			spd.LastOutType = static.TYPE_NULL
			spd.LastOutTypeClient = static.TYPE_NULL
		} else {
			laseOutLen = spd.m_GameLogic.GetCardNum(spd.AllPaiOut[spd.WhoLastOut], MAXHANDCARD)
			copy(lastOutCards[:], spd.AllPaiOut[spd.WhoLastOut][:])
		}
		iNumOfKing = spd.m_GameLogic.GetKingNum(spd.AllPaiOut[spd.Whoplay], int(msg.CardCount))

		spd.WhoPass[spd.Whoplay] = false
		bRet, re2 := spd.m_GameLogic.CompareCards(lastOutCards[:], laseOutLen, spd.AllPaiOut[spd.Whoplay][:], msg.CardCount, spd.LastOutType, int(spd.ThePaiCount[spd.Whoplay]))
		if bRet {

			// 详细日志 by sam  打出510K或者7喜或者8喜,管得起才算
			if !spd.CardNotEnough(spd.Whoplay, re2.Cardtype, byte(re2.Count), msg.CardCount) {
				spd.PlayKingBomb[spd.Whoplay]++
				//详细日志
				LogStr := fmt.Sprintf("非法牌型。不是最后一手牌，却打出没有带满的牌型。 ", spd.PlayKingBomb[spd.Whoplay])
				spd.OnWriteGameRecord(wChairID, LogStr)
				return false
			}

			if z == 1 {
				bNextOnlyOne := (1 == spd.ThePaiCount[(int(spd.Whoplay)+1)%spd.GetPlayerCount()])
				if bNextOnlyOne {
					maxcard := spd.m_GameLogic.GetMaxSingleCard(spd.PlayerCards[spd.Whoplay])
					if spd.m_GameLogic.GetCardLevel(maxcard) != spd.m_GameLogic.GetCardLevel(spd.AllPaiOut[spd.Whoplay][0]) {
						spd.OnWriteGameRecord(wChairID, "非法牌型。下家只剩一张牌，出单张时必须是手上的最大牌。")
						return false
					}
				}
			}
			//回放增加出牌日志
			spd.addReplayOrder(spd.Whoplay, meta2.DG_REPLAY_OPT_OUTCARD, int(static.INVALID_BYTE), static.HF_BytesToInts(msg.CardData[:]))
			spd.ReplayRecord.R_Orders[len(spd.ReplayRecord.R_Orders)-1].AddReplayExtData(meta2.DG_EXT_CARDTYPE, msg.CardType)

			//管得起才能算分
			if spd.GameType == meta2.GT_NORMAL {
				spd.CardScore += spd.m_GameLogic.GetScore(spd.AllPaiOut[spd.Whoplay], int(msg.CardCount))
				spd.SendTurnScore(uint16(spd.GetPlayerCount()))
				spd.ReplayRecord.R_Orders[len(spd.ReplayRecord.R_Orders)-1].AddReplayExtData(meta2.DG_EXT_TURNSCORE, spd.CardScore)
			} else {
				spd.CardScore = 0
			}

			if re2.Cardtype == TYPE_BOMB_NOMORL {
				spd.BombCount[spd.Whoplay]++
				//若炸弹实时算分
				if spd.BombRealTime && !spd.BombMode {
					spd.BombRealTimeScore(spd.Whoplay, 10)
				}
			}
			spd.WhoOutCount[spd.Whoplay]++ //出牌次数
			spd.LastOutType = re2.Cardtype
			spd.LastOutTypeClient = msg.CardType
			//服务器的3带2的飞机用的类型是TYPE_THREE_TAKE_TWO，而客户端用的是TYPE_THREESTR，托管出牌时需要转换一下
			//服务器的4带3的飞机用的类型是TYPE_FOUR_TAKE_THREE，而客户端用的是TYPE_FOURSTR，托管出牌时需要转换一下
			//是客户端出牌时，不会进入下面的条件
			if spd.LastOutTypeClient == TYPE_THREE_TAKE_TWO && msg.CardCount >= 6 {
				spd.LastOutTypeClient = TYPE_THREESTR
			} else if spd.LastOutTypeClient == TYPE_FOUR_TAKE_THREE && msg.CardCount >= 8 {
				spd.LastOutTypeClient = TYPE_FOURSTR
			}

			spd.ThePaiCount[spd.Whoplay] -= msg.CardCount
			spd.WhoLastOut = spd.Whoplay

			z := 0 //计算剩余多少张牌 m_thePaiCount可能不准
			for i := 0; i < MAXHANDCARD; i++ {
				spd.PlayerCards[spd.Whoplay][i] = 0
				if buf[i] > 0 {
					spd.PlayerCards[spd.Whoplay][z] = buf[i]
					z++
				}
			}
			spd.SendPaiCount(uint16(spd.GetPlayerCount()))
			var msgout static.Msg_S_DG_OutCard
			msgout.CardCount = msg.CardCount
			msgout.CardType = spd.LastOutTypeClient
			msgout.CurrentUser = msg.CurrentUser
			msgout.Overtime = 30
			msgout.ByClient = false
			msgout.CardData = msg.CardData
			spd.SendTableMsg(consts.MsgTypeGameOutCard, msgout)

			//插底提示
			if z == 1 {
				spd.SendPlaySoundMsg(spd.Whoplay, static.TY_ChaDi)
			}

			//判断任务是否完成
			spd.GameTask.IsTaskFinished(spd.Common, re2, spd.Whoplay, iNumOfKing)
			//end

			spd.GameOverOrNextPlayer(z, re2, iNumOfKing)

		} else //压不住或牌型不正确
		{
			spd.WhoPass[spd.Whoplay] = true
			return false
		}
	}

	return true
}

// 游戏是否可以结束或者下一个玩家出牌
func (spd *SportPDK) GameOverOrNextPlayer(byLeftCardNum int, re static.TCardType, iNumOfKing int) int {
	//硬牌，有一个没有牌了就结束游戏
	if spd.GameType == meta2.GT_ROAR {
		//结束游戏
		if byLeftCardNum == 0 {
			//判断最后一手的任务是否完成
			spd.GameTask.IsTaskFinishedOfLastHand(spd.Common, re, spd.Whoplay, iNumOfKing)
			//end

			spd.Nextbanker = spd.Whoplay
			spd.OnGameEndNormal(spd.Whoplay, meta2.GOT_NORMAL)
			return 1
		} else {
			spd.GoNextPlayer()
		}
	} else if spd.GameType == meta2.GT_NORMAL {
		//不硬牌，找朋友的模式
		//m_whoplay的牌没有了，那么检查下我的对家结束没有
		if byLeftCardNum == 0 {
			//判断最后一手的任务是否完成
			spd.GameTask.IsTaskFinishedOfLastHand(spd.Common, re, spd.Whoplay, iNumOfKing)
			//end
			spd.WhoAllOutted[spd.Whoplay] = true
			spd.AllOutCnt++
			spd.PlayerTurn[spd.Whoplay] = spd.AllOutCnt

			spd.SendPlayerTurn(uint16(spd.GetPlayerCount()))
			spd.SendPlaySoundMsg(spd.Whoplay, static.TY_AllOut) //牌出完了需要客户端加特效

			spd.PlayerCardScore[spd.Whoplay] += spd.CardScore
			spd.SendPlayerScore(uint16(spd.GetPlayerCount()), spd.Whoplay, spd.CardScore)

			//最后一手是否是炸弹  ，放在这里可以保证最后一手是手动抢点出牌时也不会有错
			if spd.BombMode {
				if spd.LastOutType == TYPE_BOMB_NOMORL {
					spd.ValidBombCount[spd.WhoLastOut]++
					//若炸弹实时算分
					if spd.BombRealTime {
						spd.BombRealTimeScore(spd.WhoLastOut, 10)
					}
				}
			}
			spd.OnGameEndNormal(spd.Whoplay, meta2.GOT_NORMAL)
			return 1
		} else {
			spd.GoNextPlayer()
		}
	}
	return 1
}

func (spd *SportPDK) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	if wChairID >= uint16(spd.GetPlayerCount()) {
		return false
	}
	if cbReason != meta2.GOT_NORMAL && cbReason != meta2.GOT_DOUBLEKILL && cbReason != meta2.GOT_ZHONGTU {
		return false
	}
	if spd.GameType == meta2.GT_NORMAL {
		spd.CardScore = 0
		spd.SendTurnScore(uint16(spd.GetPlayerCount()))
	}
	spd.GameEndStatus = static.GS_MJ_END

	//定义变量
	iScore := [MAX_PLAYER]int{}
	spd.XiScore = [MAX_PLAYER]int{}
	fan := 1
	byGongType := byte(0)
	byGongType = static.G_Bangong

	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_NORMAL
	endgameMsg.TheOrder = spd.CurCompleteCount
	endgameMsg.WhoKingBomb = static.INVALID_CHAIR //4王一起出了才算

	for i := 0; i < spd.GetPlayerCount(); i++ {
		endgameMsg.Have7Xi[i] = spd.Who7Xi[i]
		endgameMsg.Have8Xi[i] = spd.Who8Xi[i]
		endgameMsg.Have510K[i] = spd.WhoSame510K[i]
	}
	if cbReason == meta2.GOT_NORMAL {
		//普通类型，包括硬牌结束以及找朋友结束;游戏结束
		if spd.GameType == meta2.GT_NORMAL {
			endgameMsg.EndType = static.TY_SCORE
			for i := 0; i < spd.GetPlayerCount(); i++ {
				//if 0 == spd.WhoOutCount[i] || (spd.WhoOutCount[i] == 1 && i == int(spd.Banker)) {
				if 0 == spd.WhoOutCount[i] {
					endgameMsg.ChunTian[i] = 1 //被关
				}
				//反春天
				if spd.KeFan == true && 1 == spd.WhoOutCount[i] && uint16(i) == spd.Banker && spd.ThePaiCount[i] >= 1 {
					endgameMsg.ChunTian[i] = 1
				}
				if spd.ThePaiCount[i] >= 1 {
					//输家处理
					iTempScore := 0
					if spd.ThePaiCount[i] > 1 {
						iTempScore = int(spd.ThePaiCount[i]) * spd.IBase
					}
					if endgameMsg.ChunTian[i] == 1 {
						//春天要乘2
						iTempScore *= 2
					}
					if spd.Bird[i] == 1 {
						//抓鸟要乘2，但炸弹分不乘
						iTempScore *= 2
					} else if spd.Bird[wChairID] == 1 {
						//如果赢家抓鸟，其他两个人都要乘2
						iTempScore *= 2
					}
					iScore[i] += (-1) * iTempScore
					iScore[wChairID] += iTempScore
					//endgameMsg.WinLose[i] = 0
				} else if spd.ThePaiCount[i] == 0 {
					//endgameMsg.WinLose[i] = 1
				}
			}
			//算炸弹分
			bombScore := [MAX_PLAYER]int{}
			for i := 0; i < spd.GetPlayerCount(); i++ {
				spd.MaxBombCount[i] += spd.BombCount[i] //改成总炸弹数目
				if spd.BombCount[i] > 0 && !spd.BombRealTime {
					//if spd.BombCount[i] > spd.MaxBombCount[i] {
					//	spd.MaxBombCount[i] = spd.BombCount[i]
					//}
					//spd.MaxBombCount[i] += spd.BombCount[i]//改成总炸弹数目
					iTempScore := 0
					if spd.BombMode {
						iTempScore = 10 * spd.ValidBombCount[i] * spd.IBase
					} else {
						iTempScore = 10 * spd.BombCount[i] * spd.IBase
					}
					iOther1 := (i + 1) % spd.GetPlayerCount()
					iOther2 := (i + 2) % spd.GetPlayerCount()
					iOtherFan1 := 1
					iOtherFan2 := 1
					//春天影响炸弹分
					if endgameMsg.ChunTian[iOther1] == 1 {
						iOtherFan1 = 2
					}
					if endgameMsg.ChunTian[iOther2] == 1 {
						iOtherFan2 = 2
					}
					bombScore[iOther1] += -iTempScore * iOtherFan1
					bombScore[iOther2] += -iTempScore * iOtherFan2
					bombScore[i] += iTempScore*iOtherFan1 + iTempScore*iOtherFan2
				}
			}

			iScore[0] += bombScore[0]
			iScore[1] += bombScore[1]
			iScore[2] += bombScore[2]
			//这里来计算输赢的分数
			for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
				spd.AddSpecailScore(&iScore, i, spd.IBase)
			}

		} else if spd.GameType == meta2.GT_ROAR {

		}
	}
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		if spd.BombMode {
			endgameMsg.FaScore[i] = spd.ValidBombCount[i]
		} else {
			endgameMsg.FaScore[i] = spd.BombCount[i]
		}
		endgameMsg.ZhuaNiao[i] = spd.Bird[i]
	}
	endgameMsg.FanShu = fan
	endgameMsg.GongType = byGongType
	for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
		cn := 0
		for j := 0; j < MAXHANDCARD; j++ {
			if spd.m_GameLogic.IsValidCard(spd.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spd.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = spd.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)

		endgameMsg.Score[i] = iScore[i] - spd.Spay
		//实时炸弹分
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		iScore[i] += _item.Ctx.StorageScore
		if iScore[i] > 0 {
			endgameMsg.WinLose[i] = 1
		} else if iScore[i] < 0 {
			endgameMsg.WinLose[i] = 0
		} else {
			endgameMsg.WinLose[i] = 2
		}

		spd.LastScore[i] = iScore[i]
		spd.Total[i] += iScore[i]
		endgameMsg.TotalScore[i] = spd.Total[i]

		if spd.MaxScore[i] < spd.LastScore[i] {
			spd.MaxScore[i] = spd.LastScore[i]
		}

		if spd.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			spd.TotalFirstTurn[i]++
		}

		endgameMsg.TheTurn[i] = spd.PlayerTurn[i]
		endgameMsg.GetScore[i] = spd.PlayerCardScore[i]
	}
	if spd.WhoRoar < MAX_PLAYER {
		spd.TotalDuPai[spd.WhoRoar]++
	}
	endgameMsg.HouPaiChair = spd.WhoRoar
	endgameMsg.TheBank = spd.Banker
	endgameMsg.TheParter = spd.BankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【剩余牌数%d，炸弹数%d，春天%d，抓鸟%d】", spd.LastScore[i], spd.Total[i], spd.ThePaiCount[i], spd.BombCount[i], endgameMsg.ChunTian[i], spd.Bird[i])
		spd.OnWriteGameRecord(i, recrodStr)
	}

	for i := 0; i < spd.GetPlayerCount(); i++ {
		endgameMsg.XiScore[i] = spd.XiScore[i] //喜钱
	}
	// 调用结算接口
	_, endgameMsg.UserVitamin = spd.OnSettle(endgameMsg.Score, consts.EventSettleGameOver)

	endgameMsg.TrustTime = 10

	//发送信息
	spd.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	spd.OnWriteGameRecord(static.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	//回放增加结束数据
	spd.addReplayOrder(wChairID, meta2.DG_REPLAY_OPT_END_GAME, 0, []int{})

	// 数据库写出牌记录	// 写完后清除数据
	spd.TableWriteOutDate(endgameMsg.Score)

	endgameMsg.Score = spd.LastScore
	spd.VecGameEnd = append(spd.VecGameEnd, endgameMsg) //保存，用于汇总计算

	for _, v := range spd.PlayerInfo { //i := 0; i < spd.GetPlayerCount(); i++ {
		wintype := static.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinLose[v.Seat] == 1 {
			wintype = static.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = static.ScoreKind_Lost //enScoreKind_Lost;
		}
		//iAward := spd.GetTaskAward(v.Seat)//金豆任务，先留着备用
		spd.TableWriteGameDate(int(spd.CurCompleteCount), v, wintype, spd.LastScore[v.Seat])
	}
	spd.Nextbanker = wChairID

	//扣房卡
	if spd.CurCompleteCount == 1 {
		spd.TableDeleteFangKa(spd.CurCompleteCount)
	}
	//结束游戏
	if int(spd.CurCompleteCount) >= spd.Rule.JuShu { //局数够了
		spd.CalculateResultTotal(static.GER_NORMAL, wChairID, 0) //计算总发送总结算

		spd.UpdateOtherFriendDate(&endgameMsg, false)
		//通知框架结束游戏
		//spd.SetGameStatus(public.GS_MJ_FREE)
		spd.ConcludeGame()

	} else {
		check := false
		if spd.Rule.Overtime_dismiss != -1 {
			for i := 0; i < spd.GetPlayerCount(); i++ {
				item := spd.GetUserItemByChair(uint16(i))
				if item == nil || !item.CheckTRUST() {
					continue
				}
				if check {
					var _msg = &static.Msg_C_DismissFriendResult{
						Id:   item.Uid,
						Flag: true,
					}
					spd.OnDismissResult(item.Uid, _msg)
				} else {
					check = true
					var msg = &static.Msg_C_DismissFriendReq{
						Id: item.Uid,
					}
					spd.SetDismissRoomTime(spd.Rule.Overtime_dismiss)
					spd.OnDismissFriendMsg(item.Uid, msg)
				}
			}
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////
	// 1、第一局随机坐庄；2、胡牌或流局连庄；3、第一局无人胡牌则下家坐庄；
	//if spd.BankerUser != nWinner && nWinner != public.INVALID_CHAIR {
	//	spd.BankerUser = nWinner
	//} else {
	//	//庄家赢了，还是这个庄家做庄
	//}

	spd.OnGameEnd()
	spd.RepositTable(false) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	if int(spd.CurCompleteCount) < spd.Rule.JuShu {
		for _, v := range spd.PlayerInfo {
			v.Ctx.Timer.SetTimer(components2.GameTime_AutoNext, 10)
		}
	}
	return true
}

func (spd *SportPDK) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	if wChairID >= uint16(spd.GetPlayerCount()) {
		return false
	}
	if cbReason != meta2.GOT_ESCAPE && cbReason != meta2.GOT_TUOGUAN {
		return false
	}
	if spd.GameType == meta2.GT_NORMAL {
		spd.CardScore = 0
		spd.SendTurnScore(uint16(spd.GetPlayerCount()))
	}
	spd.GameEndStatus = static.GS_MJ_END

	//定义变量
	iScore := [MAX_PLAYER]int{}
	spd.XiScore = [MAX_PLAYER]int{}

	byGongType := byte(0)
	byGongType = static.G_Bangong

	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_USER_LEFT
	endgameMsg.TheOrder = spd.CurCompleteCount
	endgameMsg.WhoKingBomb = static.INVALID_CHAIR //4王一起出了才算
	if spd.WhoHasKingBomb >= 0 && spd.WhoHasKingBomb < uint16(spd.GetPlayerCount()) {
		if spd.PlayKingBomb[spd.WhoHasKingBomb] > 0 {
			endgameMsg.WhoKingBomb = spd.WhoHasKingBomb
		}
	}

	for i := 0; i < spd.GetPlayerCount(); i++ {
		endgameMsg.Have7Xi[i] = spd.Who7Xi[i]
		endgameMsg.Have8Xi[i] = spd.Who8Xi[i]
		endgameMsg.Have510K[i] = spd.WhoSame510K[i]
	}

	endgameMsg.EndType = static.TY_ESCAPE

	nFan := 0
	if spd.GameState == meta2.GsRoarPai {
		nFan = spd.FaOfTao * 2
	} else {
		nFan = spd.FaOfTao
	}
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		if i == wChairID {
			iScore[i] = -spd.IBase * nFan
			//endgameMsg.WinLose[i] = 0
		} else {
			if spd.GameState == meta2.GsRoarPai {
				iScore[i] = spd.IBase * (spd.JiangOfTao) * 2
			} else {
				iScore[i] = spd.IBase * (spd.JiangOfTao)
			}
			//endgameMsg.WinLose[i] = 1
		}
	}
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		spd.AddSpecailScore(&iScore, i, spd.IBase)
	}

	endgameMsg.FanShu = 1
	endgameMsg.GongType = byGongType
	for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
		cn := 0
		for j := 0; j < MAXHANDCARD; j++ {
			if spd.m_GameLogic.IsValidCard(spd.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spd.PlayerCards[i][j]
				cn++
			}
		}

		endgameMsg.Score[i] = iScore[i] - spd.Spay
		//实时炸弹分
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		iScore[i] += _item.Ctx.StorageScore
		if iScore[i] > 0 {
			endgameMsg.WinLose[i] = 1
		} else if iScore[i] < 0 {
			endgameMsg.WinLose[i] = 0
		} else {
			endgameMsg.WinLose[i] = 2
		}

		spd.LastScore[i] = iScore[i]
		spd.Total[i] += iScore[i]
	}

	//游戏记录
	spd.OnWriteGameRecord(wChairID, "强退游戏结束")

	//记录异常结束数据
	//spd.addReplayOrder(wChairID, E_Li_Xian, 0)

	//回放增加结束数据
	spd.addReplayOrder(wChairID, meta2.DG_REPLAY_OPT_END_GAME, 0, []int{})
	// 调用结算接口
	_, endgameMsg.UserVitamin = spd.OnSettle(endgameMsg.Score, consts.EventSettleGameOver)

	//写入游戏回放数据,写完重置当前回放数据
	spd.TableWriteOutDate(endgameMsg.Score)

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d", spd.LastScore[i], spd.Total[i])
		spd.OnWriteGameRecord(i, recrodStr)
	}

	//发送信息
	spd.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	endgameMsg.Score = spd.LastScore
	spd.VecGameEnd = append(spd.VecGameEnd, endgameMsg) //保存，用于汇总计算

	//数据库写分
	for _, v := range spd.PlayerInfo { //i := 0; i < spd.GetPlayerCount(); i++ {
		wintype := static.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinLose[v.Seat] == 1 {
			wintype = static.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = static.ScoreKind_Lost //enScoreKind_Lost;
		}
		if endgameMsg.EndType == static.TY_ESCAPE {
			if v.Seat == wChairID {
				wintype = static.ScoreKind_Flee
			} else {
				wintype = static.ScoreKind_pass //逃跑活动分数在对战统计中忽略
			}
		}
		//iAward := spd.GetTaskAward(v.Seat)//金豆任务，先留着备用
		spd.TableWriteGameDate(int(spd.CurCompleteCount), v, wintype, spd.LastScore[v.Seat])
	}

	spd.UpdateOtherFriendDate(&endgameMsg, true)
	//结束游戏
	spd.CalculateResultTotal(static.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	spd.CurCompleteCount = 0
	//spd.SetGameStatus(public.GS_MJ_FREE)
	spd.ConcludeGame()
	spd.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 解散，结束游戏
func (spd *SportPDK) OnGameEndDissmiss(wChairID uint16, cbReason byte) bool {
	//if spd.Rule.HasPao && spd.PayPaoStatus {
	//	spd.CurCompleteCount--
	//}
	//变量定义
	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_DISMISS
	endgameMsg.TheOrder = spd.CurCompleteCount
	//定义变量
	iScore := [MAX_PLAYER]int{}

	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		spd.AddSpecailScore(&iScore, i, spd.IBase)
	}

	for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
		cn := 0
		for j := 0; j < MAXHANDCARD; j++ {
			if spd.m_GameLogic.IsValidCard(spd.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spd.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = spd.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)

		endgameMsg.Score[i] = iScore[i] - spd.Spay
		//实时炸弹分
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		iScore[i] += _item.Ctx.StorageScore

		spd.LastScore[i] = iScore[i]
		spd.Total[i] += iScore[i]
		endgameMsg.TotalScore[i] = spd.Total[i]

		if spd.MaxScore[i] < spd.LastScore[i] {
			spd.MaxScore[i] = spd.LastScore[i]
		}

		if spd.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			spd.TotalFirstTurn[i]++
		}

		endgameMsg.TheTurn[i] = spd.PlayerTurn[i]
		endgameMsg.GetScore[i] = spd.PlayerCardScore[i]
	}
	if spd.WhoRoar < MAX_PLAYER {
		spd.TotalDuPai[spd.WhoRoar]++
	}
	endgameMsg.HouPaiChair = spd.WhoRoar
	endgameMsg.TheBank = spd.Banker
	endgameMsg.TheParter = spd.BankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", spd.LastScore[i], spd.Total[i], spd.PlayerCardScore[i])
		spd.OnWriteGameRecord(i, recrodStr)
	}
	for i := 0; i < spd.GetPlayerCount(); i++ {
		endgameMsg.XiScore[i] = spd.XiScore[i] //喜钱
	}

	//记录异常结束数据
	for i := 0; i < spd.GetPlayerCount(); i++ {
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
			//spd.addReplayOrder(uint16(i), E_Li_Xian, 0)
		}
	}

	//回放增加结束数据
	spd.addReplayOrder(wChairID, meta2.DG_REPLAY_OPT_DIS_GAME, 0, []int{})
	// 调用结算接口
	_, endgameMsg.UserVitamin = spd.OnSettle(endgameMsg.Score, consts.EventSettleGameOver)

	//写入游戏回放数据,写完重置当前回放数据
	spd.TableWriteOutDate(endgameMsg.Score)

	//发送信息
	spd.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	endgameMsg.Score = spd.LastScore
	spd.VecGameEnd = append(spd.VecGameEnd, endgameMsg) //保存，用于汇总计算

	//游戏记录
	spd.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

	if spd.GameEndStatus != static.GS_MJ_FREE { //游戏未开局不写记录
		//数据库写分
		for _, v := range spd.PlayerInfo {
			if v.Seat != static.INVALID_CHAIR {
				if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
					spd.TableWriteGameDate(int(spd.CurCompleteCount), v, static.ScoreKind_pass, spd.LastScore[v.Seat])
				} else {
					spd.TableWriteGameDate(int(spd.CurCompleteCount), v, static.ScoreKind_pass, spd.LastScore[v.Seat])
				}
			}

		}
	}

	spd.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	spd.CalculateResultTotal(static.GER_DISMISS, wChairID, 0)
	spd.GameEndStatus = static.GS_MJ_END
	//结束游戏
	//spd.SetGameStatus(public.GS_MJ_FREE)
	spd.ConcludeGame()

	spd.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 程序异常，解散游戏
func (spd *SportPDK) OnGameEndErrorDissmiss(wChairID uint16, cbReason byte) bool {
	//if spd.Rule.HasPao && spd.PayPaoStatus {
	//	spd.CurCompleteCount--
	//}
	//变量定义
	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_DISMISS
	endgameMsg.TheOrder = spd.CurCompleteCount
	//定义变量
	iScore := [MAX_PLAYER]int{}

	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		spd.AddSpecailScore(&iScore, i, spd.IBase)
	}

	for i := uint16(0); i < static.MAX_PLAYER_3P; i++ {
		cn := 0
		for j := 0; j < MAXHANDCARD; j++ {
			if spd.m_GameLogic.IsValidCard(spd.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spd.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = spd.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)

		endgameMsg.Score[i] = iScore[i] - spd.Spay
		//实时炸弹分
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		iScore[i] += _item.Ctx.StorageScore

		spd.LastScore[i] = iScore[i]
		spd.Total[i] += iScore[i]
		endgameMsg.TotalScore[i] = spd.Total[i]

		if spd.MaxScore[i] < spd.LastScore[i] {
			spd.MaxScore[i] = spd.LastScore[i]
		}

		if spd.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			spd.TotalFirstTurn[i]++
		}

		endgameMsg.TheTurn[i] = spd.PlayerTurn[i]
		endgameMsg.GetScore[i] = spd.PlayerCardScore[i]
	}
	if spd.WhoRoar < MAX_PLAYER {
		spd.TotalDuPai[spd.WhoRoar]++
	}
	endgameMsg.HouPaiChair = spd.WhoRoar
	endgameMsg.TheBank = spd.Banker
	endgameMsg.TheParter = spd.BankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spd.GetPlayerCount()); i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", spd.LastScore[i], spd.Total[i], spd.PlayerCardScore[i])
		spd.OnWriteGameRecord(i, recrodStr)
	}
	for i := 0; i < spd.GetPlayerCount(); i++ {
		endgameMsg.XiScore[i] = spd.XiScore[i] //喜钱
	}

	//记录异常结束数据
	for i := 0; i < spd.GetPlayerCount(); i++ {
		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
			//spd.addReplayOrder(uint16(i), E_Li_Xian, 0)
		}
	}

	//回放增加结束数据
	spd.addReplayOrder(wChairID, meta2.DG_REPLAY_OPT_DIS_GAME, 0, []int{})
	// 调用结算接口
	_, endgameMsg.UserVitamin = spd.OnSettle(endgameMsg.Score, consts.EventSettleGameOver)

	//写入游戏回放数据,写完重置当前回放数据
	spd.TableWriteOutDate(endgameMsg.Score)

	//发送信息
	spd.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	endgameMsg.Score = spd.LastScore
	spd.VecGameEnd = append(spd.VecGameEnd, endgameMsg) //保存，用于汇总计算

	//游戏记录
	spd.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

	spd.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	spd.CalculateResultTotal(static.GER_DISMISS, wChairID, 1)
	spd.GameEndStatus = static.GS_MJ_END
	//结束游戏
	//spd.SetGameStatus(public.GS_MJ_FREE)
	spd.ConcludeGame()

	spd.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 游戏结束,流局结束，统计积分
func (spd *SportPDK) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if spd.GameEndStatus == static.GS_MJ_END && cbReason == static.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	//by leon总结算时才能
	//m_pITableFrame->KillGameTimer(IDI_OUT_TIME);
	// 清除超时检测
	for _, v := range spd.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}
	//1.如果提供的用户为空，不可能，直接返回false
	switch cbReason {
	case static.GER_NORMAL: //常规结束

		return spd.OnGameEndNormal(wChairID, meta2.GOT_NORMAL)
	case static.GER_USER_LEFT: //用户强退

		return spd.OnGameEndUserLeft(wChairID, meta2.GOT_ESCAPE)

	case static.GER_DISMISS: //解散游戏

		return spd.OnGameEndDissmiss(wChairID, meta2.GOT_DISMISS)

	case static.GER_GAME_ERROR: //程序异常，解散游戏

		return spd.OnGameEndErrorDissmiss(wChairID, meta2.GOT_DISMISS)

	}
	return false
}

// 计算总发送总结算
func (spd *SportPDK) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	spd.TimeEnd = time.Now().Unix() //大局结束时间
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_DG_BALANCE_GAME
	balanceGame.Userid = spd.Rule.FangZhuID
	balanceGame.CurTotalCount = spd.CurCompleteCount //总盘数
	balanceGame.TimeStart = spd.Common.TimeStart     //游戏大局开始时间
	balanceGame.TimeEnd = spd.TimeEnd
	for i := 0; i < len(spd.VecGameEnd); i++ {
		for j := 0; j < spd.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += spd.VecGameEnd[i].Score[j] //总分
		}
	}

	//打印日志
	recrodStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）", cbReason)
	spd.OnWriteGameRecord(wChairID, recrodStr)

	for i := 0; i < spd.GetPlayerCount(); i++ {
		spd.OnWriteGameRecord(uint16(i), "当前该玩家离线")
	}

	//非正常结束，记录总结算时玩家状态：0正常，1解散，2离线
	if static.GER_USER_LEFT == cbReason {
		for i := 0; i < spd.GetPlayerCount(); i++ {
			if i == int(wChairID) {
				balanceGame.UserEndState[i] = 2
			}

		}
	} else {
		if static.GER_DISMISS == cbReason {
			for i := 0; i < spd.GetPlayerCount(); i++ {
				if i == int(wChairID) {
					balanceGame.UserEndState[i] = 1
				} else {
					if _item := spd.GetUserItemByChair(uint16(i)); _item != nil && _item.UserStatus == static.US_OFFLINE {
						balanceGame.UserEndState[i] = 2
					}
				}
			}
		}
	}

	// 如果是正常结束
	//if (GER_NORMAL == cbReason)
	{
		// 有大赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		//		iChairID := 0
		for i := 0; i < spd.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < spd.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				//				iChairID = j
				iMaxScoreCount++
			}
		}
		if iMaxScoreCount == 1 && spd.Rule.CreateType == 3 { // 大赢家支付
			//IServerUserItem * pIServerUserItem = m_pITableFrame->GetServerUserItem(iChairID);
			//DWORD userid = pIServerUserItem->GetUserID();
			//				m_pITableFrame->TableDeleteDaYingJiaFangKa(userid);
		}
	}

	//! 计算大赢家的座位号
	bigWinScore := 0
	for i := 0; i < spd.GetPlayerCount(); i++ {
		if balanceGame.GameScore[i] >= bigWinScore {
			bigWinScore = balanceGame.GameScore[i]
		}
	}

	//开始就解散时不加，每局结束后游戏状态会复位为GS_MJ_FREE
	wintype := static.ScoreKind_Draw
	if spd.CurCompleteCount == 1 && spd.GetGameStatus() != static.GS_MJ_END {
		if spd.ReWriteRec <= 1 {
			wintype = static.ScoreKind_pass
		}
	} else {
		if spd.CurCompleteCount == 0 { //有可能第一局还没有开始，就解散了（比如在吓跑的过程中解散）
			wintype = static.ScoreKind_pass
		}
	}

	if cbSubReason == 0 {
		for i := 0; i < spd.GetPlayerCount(); i++ {
			_userItem := spd.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.MaxGetScore[i] = spd.MaxScore[i]
			balanceGame.FirstTurnCount[i] = spd.TotalFirstTurn[i]
			balanceGame.RoarCount[i] = spd.MaxBombCount[i]

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
			spd.TableWriteGameDateTotal(int(spd.CurCompleteCount), i, wintype, balanceGame.GameScore[i], isBigWin)
		}
	} else {
		spd.UpdateErrGameTotal(spd.GetTableInfo().GameNum)
	}

	for i := 0; i < spd.GetPlayerCount(); i++ {
		_userItem := spd.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}

		gameendStr := ""
		if len(spd.VecGameEnd) > 0 {
			gameendStr = static.HF_JtoA(spd.VecGameEnd[len(spd.VecGameEnd)-1])
		}
		gamedataStr := ""

		spd.SaveLastGameinfo(_userItem.Uid, gameendStr, static.HF_JtoA(balanceGame), gamedataStr)
	}

	// 记录用户好友房历史战绩
	if wintype != static.ScoreKind_pass {
		spd.TableWriteHistoryRecordWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
		spd.TableWriteHistoryRecordDetailWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
	}

	balanceGame.End = 0

	//发消息
	spd.SendTableMsg(consts.MsgTypeGameBalanceGame, balanceGame)

	spd.resetEndDate()
}

func (spd *SportPDK) resetEndDate() {
	spd.CurCompleteCount = 0
	spd.VecGameEnd = []static.Msg_S_DG_GameEnd{}

	for _, v := range spd.PlayerInfo {
		v.OnEnd()
	}
}

func (spd *SportPDK) UpdateOtherFriendDate(GameEnd *static.Msg_S_DG_GameEnd, bEnd bool) {

}

// ! 给客户端发送总结算
func (spd *SportPDK) CalculateResultTotal_Rep(msg *static.Msg_C_BalanceGameEeq) {
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = spd.Rule.FangZhuID
	balanceGame.CurTotalCount = spd.CurCompleteCount //总盘数

	/*为什么要特意判断下0？之前没有这个协议，是新需求要加的之前很多数据都在RepositTableFrameSink里面重置了，但是由于新功能的开发
	部分数据放到游戏开始的时候重置了，但是存在的问题是，如果一个桌子解散了，这个桌子在服务器内部还是存在的，他的部分数据也是存在的
	当某个玩家进来，还没有点开始的时候点战绩，这时候存在把本包厢桌子上次数据发送过来的情况，但是改动m_MaxFanUserCount之类变量的值又
	有很大风险，因此此处做个特出处理，如果是第0局，没有开始，那就无条件全部返回0*/
	if 0 == balanceGame.CurTotalCount {
		for i := 0; i < len(spd.VecGameEnd); i++ {
			for j := 0; j < spd.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += 0 //总分
			}
		}

		for i := 0; i < spd.GetPlayerCount(); i++ {
			balanceGame.ChiHuUserCount[i] = 0
			balanceGame.ProvideUserCount[i] = 0
			balanceGame.FXMaxUserCount[i] = 0
			balanceGame.HHuUserCount[i] = 0
			balanceGame.UserEndState[i] = 0
		}
	} else {
		for i := 0; i < len(spd.VecGameEnd); i++ {
			for j := 0; j < spd.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += spd.VecGameEnd[i].Score[j] //总分
			}
		}

		for i := 0; i < spd.GetPlayerCount(); i++ {
			balanceGame.UserEndState[i] = 0
		}

		// 有打赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		for i := 0; i < spd.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < spd.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				iMaxScoreCount++
			}
		}

		for i := 0; i < spd.GetPlayerCount(); i++ {
			_userItem := spd.GetUserItemByChair(uint16(i))
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
	spd.SendPersonMsg(consts.MsgTypeGameBalanceGame, balanceGame, spd.GetChairByUid(msg.Id))
}

func (spd *SportPDK) sendGameSceneStatusPlay(player *components2.Player) bool {

	wChiarID := player.GetChairID()

	if wChiarID >= uint16(spd.GetPlayerCount()) {
		xlog.Logger().Info("sendGameSceneStatusPlay invalid chair")
		return false
	}

	//重置离线解散时间
	spd.SetOfflineRoomTime(1800)

	//发送底分
	var msgRule static.Msg_S_DG_GameRule
	msgRule.CellScore = spd.GetCellScore()
	msgRule.FaOfTao = spd.FaOfTao
	spd.SendTableMsg(consts.MsgTypeGameRule, msgRule)

	//spd.WhoBreak[wChiarID] = false;//重连嘛，所以取消断线

	//提示其他玩家，我又回来了！
	var msgTip static.Msg_S_DG_ReLinkTip
	msgTip.ReLinkUser = wChiarID
	msgTip.ReLinkTip = 0
	spd.SendTableMsg(consts.MsgTypeReLinkTip, msgTip)

	//取消托管

	//变量定义
	var StatusPlay static.CMD_S_DG_StatusPlay
	//游戏变量
	StatusPlay.Overtime = 0
	if time.Now().Unix()+1 < spd.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.Overtime = spd.LimitTime - time.Now().Unix()
	}

	StatusPlay.GameState = byte(spd.GameState)         //游戏状态，当前处在哪个阶段，1吼牌阶段，2打牌阶段
	StatusPlay.BankerUser = spd.Banker                 //庄家
	StatusPlay.CurrentUser = spd.Whoplay               //当前牌权玩家
	StatusPlay.CellScore = spd.GetCellScore()          //m_pGameServiceOption->lCellScore;
	StatusPlay.WhoReLink = wChiarID                    //谁断线重连的
	StatusPlay.WhoLastOut = spd.WhoLastOut             //上一个出牌玩家
	StatusPlay.TrustCounts = spd.TrustCounts[wChiarID] //叫我托管了几次了
	StatusPlay.RoarPai = spd.RoarPai                   //叫的什么牌
	StatusPlay.WhoRoar = spd.WhoRoar                   //谁叫了牌
	StatusPlay.WhoMJ = static.INVALID_CHAIR            //初始化谁鸣鸡为无效值
	if spd.BMingJiFlag {
		StatusPlay.WhoMJ = spd.BankParter //谁鸣鸡
	}
	StatusPlay.TurnScore = spd.CardScore           //本轮分
	StatusPlay.LastPaiType = spd.LastOutTypeClient //上一次出牌的类型，可能没用？修改成客户端传过来的
	StatusPlay.TheOrder = spd.CurCompleteCount

	for j := 0; j < MAXHANDCARD; j++ {
		if spd.m_GameLogic.IsValidCard(spd.PlayerCards[wChiarID][j]) {
			StatusPlay.MyCards[j] = spd.PlayerCards[wChiarID][j] //刚才出的牌
			StatusPlay.MyCardsCount++
		}
	}
	for i := 0; i < spd.GetPlayerCount(); i++ {

		_item := spd.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE {
			StatusPlay.WhoBreak[i] = true //掉线的有那些人
		}
		StatusPlay.TuoGuanPlayer[i] = _item.CheckTRUST()                 //托管的有那些人
		StatusPlay.LastScore[i] = int(spd.LastScore[i])                  //上一轮输赢
		StatusPlay.Total[i] = int(spd.Total[i]) + _item.Ctx.StorageScore //总输赢
		StatusPlay.WhoPass[i] = spd.WhoPass[i]                           //谁放弃
		StatusPlay.WhoReady[i] = spd.WhoReady[i]                         //谁已经完成叫牌过程
		StatusPlay.Score[i] = spd.PlayerCardScore[i]                     //每个人的分

		for j := 0; j < MAXHANDCARD; j++ {
			StatusPlay.OutCard[i][j] = spd.AllPaiOut[i][j]      //刚才出的牌
			StatusPlay.LastOutCard[i][j] = spd.LastPaiOut[i][j] //上一轮出的牌，（这个其实可以不要）
		}
	}

	//发送场景
	spd.SendPersonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, wChiarID)

	//出完牌的顺序，游次
	spd.SendPlayerTurn(wChiarID)

	//剩余牌数量
	spd.SendPaiCount(wChiarID)
	if spd.GameType == meta2.GT_NORMAL {
		spd.SendPlayerScore(wChiarID, static.INVALID_CHAIR, 0) //每个人的分,这个找朋友模式才有
	}

	//spd.SendTaskID(false, wChiarID);
	spd.GameTask.SendTaskID(spd.Common, false, wChiarID)

	//发送权限（什么权限、该谁出牌、倒计时等）
	if spd.GameState == meta2.GsRoarPai {
		spd.SendPower(spd.Whoplay, 1, int(StatusPlay.Overtime))
	} else if spd.GameState == meta2.GsPlay {
		spd.SendPower(spd.Whoplay, 2, int(StatusPlay.Overtime))
	} else if spd.GameState == meta2.GsStartAnimation {
		spd.PowerStartTime = time.Now().Unix()             //权限开始时间
		spd.setLimitedTime(int64(StatusPlay.Overtime + 1)) //避免StatusPlay.Overtime=0的问题.go定时器不支持0定时器
	} else if spd.GameState == meta2.GsSplitCards {
		spd.PowerStartTime = time.Now().Unix()             //权限开始时间
		spd.setLimitedTime(int64(StatusPlay.Overtime + 1)) //避免StatusPlay.Overtime=0的问题.go定时器不支持0定时器
	} else if spd.GameState == meta2.GsQuickPass {
		if spd.QuickPass {
			spd.PowerStartTime = time.Now().Unix()             //权限开始时间
			spd.setLimitedTime(int64(StatusPlay.Overtime + 1)) //避免StatusPlay.Overtime=0的问题.go定时器不支持0定时器
		} else {
			spd.SendPower(spd.Whoplay, 5, int(StatusPlay.Overtime))
		}
	}

	//发小结消息
	if byte(len(spd.VecGameEnd)) == spd.CurCompleteCount && spd.CurCompleteCount != 0 && int(wChiarID) < spd.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := spd.VecGameEnd[spd.CurCompleteCount-1]
		gamend.Relink = 1 //表示为断线重连

		spd.SendPersonMsg(consts.MsgTypeGameEnd, gamend, wChiarID)
	}

	//发送解散房间所有玩家的反应
	spd.SendAllPlayerDissmissInfo(player)

	return true
}

// 游戏场景消息发送
func (spd *SportPDK) SendGameScene(uid int64, status byte, secret bool) {
	player := spd.GetUserItemByUid(uid)
	if player == nil {
		return
	}
	switch status {
	case static.GS_MJ_FREE:
		spd.sendGameSceneStatusFree(player)
	case static.GS_MJ_PLAY:
		spd.sendGameSceneStatusPlay(player)
	}
}
func (spd *SportPDK) sendGameSceneStatusFree(player *components2.Player) bool {

	//变量定义
	var StatusFree static.Msg_S_DG_StatusFree
	//构造数据
	StatusFree.BankerUser = spd.Banker
	StatusFree.CellScore = spd.GetCellScore() //spd.m_pGameServiceOption->lCellScore;
	StatusFree.FaOfTao = spd.FaOfTao
	StatusFree.CellMinScore = spd.GetCellScore() //最低分
	StatusFree.CellMaxScore = spd.GetCellScore() //最低分

	//发送场景
	//	spd.SendPersonMsg(constant.MsgTypeGameStatusFree, StatusFree, PlayerInfo.GetChairID())
	spd.SendUserMsg(consts.MsgTypeGameStatusFree, StatusFree, player.Uid)

	return true
}

// ! 游戏退出
func (spd *SportPDK) OnExit(uid int64) {
	spd.Common.OnExit(uid)
}

func (spd *SportPDK) OnTime() {
	spd.Common.OnTime()
}

// ! 写游戏日志
func (spd *SportPDK) OnWriteGameRecord(seatId uint16, recordStr string) {
	spd.GameTable.WriteTableLog(seatId, recordStr)
}

// ! 写入游戏回放数据
func (spd *SportPDK) TableWriteOutDate(score [4]int) {
	if spd.ReWriteRec != 0 {
		spd.ReWriteRec++ //这种情况会>1，表示是在结算时申请解散的。
		// 写完后清除数据
		spd.ReplayRecord.ReSet()
		return
	}

	recordReplay := new(models.RecordGameReplay)
	recordReplay.GameNum = spd.GetTableInfo().GameNum
	recordReplay.RoomNum = spd.GetTableInfo().Id
	recordReplay.PlayNum = int(spd.CurCompleteCount)
	recordReplay.ServerId = server2.GetServer().Con.Id
	recordReplay.HandCard = spd.m_GameLogic.GetWriteHandReplayRecordString(spd.ReplayRecord)
	recordReplay.OutCard = spd.m_GameLogic.GetWriteOutReplayRecordString(spd.ReplayRecord, score)
	recordReplay.UVitaminMap = spd.ReplayRecord.UVitamin
	recordReplay.KindID = spd.GetTableInfo().KindId
	recordReplay.CardsNum = 0

	if id, err := server2.GetDBMgr().InsertGameRecordReplay(recordReplay); err != nil {
		xlog.Logger().Debug(fmt.Sprintf("%d,写游戏出牌记录：（%v）出错（%v）", id, recordReplay, err))
	}

	spd.RoundReplayId = recordReplay.Id

	// 写完后清除数据
	spd.ReplayRecord.ReSet()

	spd.ReWriteRec++ //在小结算过程中解散不在写回放记录了
}

// 场景保存
func (spd *SportPDK) Tojson() string {
	var _json components2.DGGameJsonSerializer

	_json.ToJsonDG(&spd.GameMetaDG)

	_json.GameCommonToJson(&spd.Common)

	return static.HF_JtoA(&_json)
}

// 场景恢复
func (spd *SportPDK) Unmarsha(data string) {
	var _json components2.DGGameJsonSerializer

	if data != "" {
		var err error
		if err = json.Unmarshal([]byte(data), &_json); err != nil {
			panic(err)
		}

		_json.UnmarshaDG(&spd.GameMetaDG)
		_json.JsonToStruct(&spd.Common)

		spd.ParseRule(spd.GetTableInfo().Config.GameConfig)
		spd.m_GameLogic.m_rule = spd.Rule

		spd.m_GameLogic.SetBombCount(4)                               //设置炸弹的最小长度(普通炸弹：四个或者四个以上相同的牌 5 10 K也是炸弹3张牌)
		spd.m_GameLogic.SetOnestrCount(5)                             //设置单顺的最小长度
		spd.m_GameLogic.SetMaxCardCount(MAXHANDCARD)                  //设置手牌最大长度
		spd.m_GameLogic.SetMaxPlayerCount(byte(spd.GetPlayerCount())) //设置玩家最大数目
		spd.m_GameLogic.SetBombSplit(spd.BombSplit)                   //设置炸弹是否可以拆
		spd.m_GameLogic.SetThreeAceBomb(spd.Bomb3Ace)                 //设置3个A是否是炸弹
		spd.m_GameLogic.Set4Take2(spd.FourTake2)                      //设置是否可以4带2
		spd.m_GameLogic.Set4Take3(spd.FourTake3)                      //设置是否可以4带3
		spd.m_GameLogic.SetLessTake(spd.LessTake)                     //设置最后一手是否可以少带
		if spd.Rule.NoWan {
			spd.m_GameLogic.SetCardNum(byte(15)) //设置张数
		} else {
			spd.m_GameLogic.SetCardNum(byte(16)) //设置张数
		}

		spd.PlayTime = 30
		spd.AutoOutTime = 0
		if spd.TuoGuan > 0 {
			spd.PlayTime = spd.TuoGuan
			//spd.AutoOutTime = spd.TuoGuan
		}
	}
}

// 解析配置的任务,格式： "1@5/2@5/33@10"
func (spd *SportPDK) ParseTaskConfig(data string) {
	spd.GameTask.Init()
	//.....
	//spd.Task.AppendTaskMapAndVec(id,award)
}

// 操作分
type Msg_S_OperateScore struct {
	OperateUser uint16     `json:"operateuser"`  //操作用户
	OperateType uint16     `json:"operatetype"`  //操作类型
	GameScore   [4]int     `json:"gamescore"`    //最新总分
	GameVitamin [4]float64 `json:"game_vitamin"` //最新疲劳值信息
	ScoreOffset [4]int     `json:"scoreoffset"`  //分数变化量
}

// 游戏过程中，玩家的分数发生变化事件
func (spd *SportPDK) OnUserScoreOffset(seat uint16, offset int) bool {
	_userItem := spd.GetUserItemByChair(seat)
	if _userItem != nil {
		_userItem.Ctx.StorageScore += offset
	}
	return true
}

// 炸弹实时扣分
func (spd *SportPDK) BombRealTimeScore(chairID uint16, baseCost int) {
	var msg Msg_S_OperateScore
	for tmpChair := 0; tmpChair < spd.GetPlayerCount(); tmpChair++ {
		if uint16(tmpChair) != chairID {
			cost := baseCost * spd.IBase
			msg.ScoreOffset[tmpChair] -= cost
			msg.ScoreOffset[chairID] += cost
		}
	}

	for tmpChair := 0; tmpChair < spd.GetPlayerCount(); tmpChair++ {
		player := spd.GetUserItemByChair(uint16(tmpChair))
		if player != nil {
			// 暂时当做杠分记录，用于大结算和解散游戏时竞技点还原
			//player.Ctx.GangScore += msg.ScoreOffset[tmpChair]
			// 记录游戏中的游戏分变化
			spd.OnUserScoreOffset(uint16(tmpChair), msg.ScoreOffset[tmpChair])
		}
	}

	msg.GameScore, msg.GameVitamin =
		spd.OnSettle(msg.ScoreOffset, consts.EventSettleGaming)

	spd.OnWriteGameRecord(chairID, "BombRealTimeScore(): 炸弹实时算分"+
		fmt.Sprintf("%d", msg.ScoreOffset[chairID]/spd.Rule.Radix)+"分")

	msg.OperateUser = chairID
	spd.SendTableMsg(consts.MsgTypeGameOperateScore, msg)
}

func (spd *SportPDK) GetPlayerHandCard(i uint16) *static.Msg_S2C_PlayCard {
	var msg static.Msg_S2C_PlayCard
	if i < MAX_PLAYER {
		msg.CardCount = spd.m_GameLogic.GetCardNum(spd.PlayerCards[i], MAXHANDCARD)
		for c := 0; c < MAXHANDCARD; c++ {
			msg.CardData[c] = spd.PlayerCards[i][c]
		}
	}
	return &msg
}
