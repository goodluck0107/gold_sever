package JingZhouHuaPai

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

/*
通山打拱好友房
*/

type Sport_zp_jzhp struct {
	components2.Common
	//游戏变量
	SportMetaJZHP
	GameLogic GameLogic_zp_jzhp //游戏逻辑
}

func (sp *Sport_zp_jzhp) GetGameConfig() *static.GameConfig { //获取游戏相关配置
	return &sp.Config
}

// 重置桌子数据
func (sp *Sport_zp_jzhp) RepositTable(ResetAllData bool) {
	rand.Seed(time.Now().UnixNano())
	for _, v := range sp.PlayerInfo {
		v.Reset()
	}
	//游戏变量
	sp.GameEndStatus = static.GS_MJ_FREE
	if ResetAllData {
		//游戏变量
		sp.SportMetaJZHP.reset()
	} else {
		//游戏变量
		sp.SportMetaJZHP.resetForNext() //保留部分数据给下一局使用
	}
}

// 解析配置的任务
func (sp *Sport_zp_jzhp) ParseRule(strRule string) {

	xlog.Logger().Info("parserRule :" + strRule)

	//表示底分要除以10
	sp.Rule.DiFen = 1
	sp.Rule.BaseScore = 1
	sp.Rule.NineSecondRoom = false

	sp.Rule.JuShu = sp.GetTableInfo().Config.RoundNum
	sp.Spay = 0
	sp.Rule.FangZhuID = sp.GetTableInfo().Creator
	sp.Rule.CreateType = sp.FriendInfo.CreateType
	sp.Rule.TrusteeCostSharing = true
	sp.Rule.Radix = 1

	sp.QuanHei = 2

	if len(strRule) == 0 {
		return
	}

	var _msg rule2.FriendRuleZP_dyzp
	err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg)
	if err == nil {
		sp.Base = _msg.Difen
		if _msg.Radix == 0 {
			sp.Rule.Radix = 1
		} else {
			sp.Rule.Radix = _msg.Radix //大厅需要使用这个，不要删除
		}
		if _msg.DuoHu == "true" {
			sp.DuoHu = true
		} else if _msg.DuoHu == "false" {
			sp.DuoHu = false
		}
		sp.GeziShu = _msg.GeziShu
		sp.HuaShu = _msg.HuaShu
		sp.Piao = _msg.Piao //荆州花牌一直为0
		if 10 == sp.Piao {
			sp.Piao = _msg.DingPiao
		}
		sp.BeiShu = _msg.BeiShu
		if _msg.QuanHei == 2 || _msg.QuanHei == 3 {
			sp.QuanHei = _msg.QuanHei
		}
		sp.Rule.Overtime_trust = _msg.Overtime_trust
		sp.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		if sp.Rule.Overtime_dismiss > 0 {
			//sp.LaunchDismissTime=sp.Rule.Overtime_dismiss
		}
		// 出牌时间 超时解散
		sp.OutCardDismissTime = _msg.OutCardDismissTime
		if sp.OutCardDismissTime > 0 {
			//和托管互斥
			sp.Rule.Overtime_trust = -1
			sp.OperateTime = int64(sp.OutCardDismissTime)
		}
		if sp.Rule.Overtime_trust > 0 {
			sp.OperateTime = int64(sp.Rule.Overtime_trust)
		}
		if _msg.DianPaoPei == "true" {
			sp.DianPaoPei = 1
		} else if _msg.DianPaoPei == "false" {
			sp.DianPaoPei = 0
		}
		if _msg.HunJiangS == 1 {
			sp.HunJiang = 1
			sp.GameLogic.InitMagicPoint(0x1C)
		} else {
			sp.HunJiang = 0
		}
		sp.FenType = _msg.FenType
		if _msg.KeChong == "true" {
			sp.KeChong = true
		} else if _msg.KeChong == "false" {
			sp.KeChong = false
		}
		if _msg.Fleetime != 0 {
			sp.Fleetime = _msg.Fleetime
		}
		sp.Rule.DissmissCount = _msg.Dissmiss
		// 小局结束自动准备
		sp.RoundOverAutoReady = _msg.RoundOverAutoReady == "true"
		// 登庄
		sp.DengZhuang = _msg.DengZhuang == "true"
		if _msg.LookonSupport == "" {
			sp.Config.LookonSupport = true
		} else {
			sp.Config.LookonSupport = _msg.LookonSupport == "true"
		}
		//离线解散时间和申请解散时间
		sp.Rule.Overtime_offdiss = _msg.Overtime_offdiss
		sp.Rule.Overtime_applydiss = _msg.Overtime_applydiss
	}
	if sp.Base < 1 {
		sp.Base = 1
	}
	//开关
	if sp.Debug > 0 {
		//Rule.JuShu = Debug;
	}
	//设置解散次数
	if sp.Rule.DissmissCount != 0 {
		sp.SetDissmissCount(sp.Rule.DissmissCount)
	}
}

// ! 开局
func (sp *Sport_zp_jzhp) OnBegin() {
	xlog.Logger().Info("onbegin")
	sp.RepositTable(true)

	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range sp.PlayerInfo {
		v.OnBegin()
	}

	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)
	sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
	sp.GameLogic.Rule = sp.Rule
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_ZP_TC_GameEnd{}
	sp.GameLogic.SetHunJiang(int(sp.HunJiang))

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()

	sp.CurCompleteCount++
	sp.GetTable().SetBegin(true)

	if 0 == sp.Rule.Overtime_offdiss {
		//设置离线解散时间15分钟
		sp.SetOfflineRoomTime(900)
	} else {
		sp.SetOfflineRoomTime(sp.Rule.Overtime_offdiss)
	}

	sp.OnGameStart()
}

func (sp *Sport_zp_jzhp) OnGameStart() {
	if !sp.CanContinue() {
		return
	}
	if sp.Piao == 100 {
		sp.SendPiaoSetting()
	} else {
		sp.StartNextGame()
	}
}

// 发送选漂对话框
func (sp *Sport_zp_jzhp) SendPiaoSetting() {

	sp.GameEndStatus = static.GS_MJ_PLAY
	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)

	for _, v := range sp.PlayerInfo {
		v.Ctx.CleanWeaveItemArray()
		v.Ctx.InitCardIndex()
	}
	sp.ReplayRecord.ReSet()
	sp.GameState = ZP_AS_XUANPIAO //设置玩家选漂的状态
	var PiaoSetting static.Msg_S_PiaoSetting
	PiaoSetting.CurrentCount = sp.CurCompleteCount //临时加1
	//向每个玩家发送数据
	for _, v := range sp.PlayerInfo {
		if v.Ctx.UserPiaoReady {
			PiaoSetting.PiaoStatus = true
			PiaoSetting.PiaoCount = v.Ctx.VecPiao.Num
			PiaoSetting.Always = false
			if int(sp.Nextbanker) < sp.GetPlayerCount() {
				PiaoSetting.BankerUser = sp.Nextbanker
			}
		} else {
			PiaoSetting.PiaoStatus = false
			PiaoSetting.PiaoCount = 0
			PiaoSetting.Always = false
			if int(sp.Nextbanker) < sp.GetPlayerCount() {
				PiaoSetting.BankerUser = sp.Nextbanker
			}
		}
		sp.SendPersonMsg(consts.MsgTypeGamePiaoSetting, PiaoSetting, v.Seat)
	}
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGamePiaoSetting, PiaoSetting)
}

func (sp *Sport_zp_jzhp) OnUserClientXuanpiao(msg *static.Msg_C_Piao) bool {
	nChiarID := sp.GetChairByUid(msg.Id)
	_userItem := sp.GetUserItemByChair(nChiarID)
	if _userItem == nil {
		return false
	}
	if nChiarID >= 0 && int(nChiarID) < sp.GetPlayerCount() {
		_userItem.Ctx.XuanPiao(msg)
		//广播选漂的消息
		sp.NotifyXuanPiao(nChiarID)
	}

	// 如果4个玩家都准备好了，自动开启下一局
	_beginCount := 0
	for _, v := range sp.PlayerInfo {
		if !v.Ctx.UserPiaoReady {
			recordStr := fmt.Sprintf("椅子编号[%d] 还没有完成选漂", v.Seat)
			sp.OnWriteGameRecord(uint16(v.Seat), recordStr)
			break
		}
		_beginCount++
	}

	if _beginCount >= sp.GetPlayerCount() {
		sp.OnWriteGameRecord(uint16(nChiarID), "所有人都完成选漂了，开始游戏")
		//游戏没有开始发牌
		if sp.GameState == ZP_AS_XUANPIAO {
			sp.StartNextGame()
		}
	}

	return true
}

// 广播玩家的状态和选漂的数目
func (sp *Sport_zp_jzhp) NotifyXuanPiao(wChairID uint16) bool {
	var sXuanPiao static.Msg_S_XuanPiao

	for _, v := range sp.PlayerInfo {
		if v.Seat != static.INVALID_CHAIR {
			sXuanPiao.Num[v.Seat] = v.Ctx.VecPiao.Num
			sXuanPiao.Status[v.Seat] = v.Ctx.UserPiaoReady
		}
	}

	//发送数据
	sp.SendTableMsg(consts.MsgTypeGameXuanpiao, sXuanPiao)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameXuanpiao, sXuanPiao)

	//游戏记录
	if wChairID == wChairID {
		recordStr := fmt.Sprintf("发送漂数：%d， 是否默认 %t", sXuanPiao.Num[wChairID], sXuanPiao.Status[wChairID])
		sp.OnWriteGameRecord(wChairID, recordStr)

		sp.addReplayOrder(wChairID, E_ZP_PIAO, sXuanPiao.Num[wChairID], []int{})
		sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_PiaoScore, int(sXuanPiao.Num[wChairID]))
	}
	return true
}

// 开始下一局游戏
func (sp *Sport_zp_jzhp) StartNextGame() {
	sp.OnStartNextGame()

	if 0 == sp.Rule.Overtime_applydiss {
		// 恢复自动解散时间2分钟
		sp.SetDismissRoomTime(120)
	} else {
		sp.SetDismissRoomTime(sp.Rule.Overtime_applydiss)
	}

	sp.GameState = ZP_AS_GAMESTART       //吼牌阶段或打牌阶段
	sp.GameEndStatus = static.GS_MJ_PLAY //当前小局游戏的状态
	sp.ReWriteRec = 0
	//开始新的一局记录(提出来)
	if 100 != sp.Piao {
		sp.ReplayRecord.ReSet()
	}

	//发送最新状态
	for i := 0; i < sp.GetPlayerCount() && i < TCGZ_MAX_PLAYER; i++ {
		sp.SendUserStatus(i, static.US_PLAY) //把状态发给其他人
	}

	//记录日志
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始StartNextGame......")
	sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("%s,restrict:%t", sp.GetTableInfo().Config.GameConfig, sp.GetTableInfo().Config.Restrict))

	//重置所有玩家的状态
	for _, v := range sp.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	//sp.ParseRule(sp.GameTable.Config.GameConfig)
	//sp.GameLogic.Rule = sp.Rule

	//1-3定漂,服务器记录定票数后就不用管，客户端每小局直接显示；100表示选漂，就要发送选漂对话框；0不漂
	if sp.Piao >= 1 && sp.Piao <= 3 {
		for _, v := range sp.PlayerInfo {
			v.Ctx.VecPiao.Num = sp.Piao
			v.Ctx.UserPiaoReady = true

			//记录定漂的回放
			wChairID := v.GetChairID()
			if int(wChairID) < sp.GetPlayerCount() {
				sp.addReplayOrder(wChairID, E_ZP_PIAO, sp.Piao, []int{})
				sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_PiaoScore, sp.Piao)
			}
		}
	}

	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)

	// 框架发送开始游戏后开始计算当前这一轮的局数
	//sp.CurCompleteCount++
	//sp.GetTable().SetBegin(true)

	LogNextStr := fmt.Sprintf("荆州花牌 【当前第%d局】", sp.CurCompleteCount)
	sp.OnWriteGameRecord(static.INVALID_CHAIR, LogNextStr)

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(sp.GetTableId()+sp.KIND_ID*100+sp.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())

	//分发扑克
	sp.LeftCardCount, sp.RepertoryCard = sp.GameLogic.RandCardData()

	//分发扑克--即每一个人解析他的19张牌结果存放在m_cbCardIndex[i]中
	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.LeftCardCount -= (TCGZ_MAXSENDCARD - 1)
		_, sp.CardIndex[i] = sp.GameLogic.SwitchToCardIndex3(sp.RepertoryCard[sp.LeftCardCount:], TCGZ_MAXSENDCARD-1)
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	if sp.HunJiang == 0 {
		sp.initDebugCards("zipaiJingZhouHuaPai_test", &sp.CardIndex, &sp.Nextbanker)
	} else {
		sp.initDebugCards("zipaiJingZhouHuaPai_M_test", &sp.CardIndex, &sp.Nextbanker)
	}

	//////////////读取配置文件设置牌型end////////////////////////////////////

	//新开局时，随机一个玩家为本轮初始庄家。本轮一游的玩家则为下轮的庄家
	if int(sp.Nextbanker) >= sp.GetPlayerCount() {
		rand_num := rand.Intn(1000)
		sp.Banker = uint16(rand_num % sp.GetPlayerCount())
		sp.Nextbanker = sp.Banker
	} else {
		sp.Banker = sp.Nextbanker
	}
	sp.CurrentUser = sp.Banker

	for seat := 0; seat < sp.GetPlayerCount(); seat++ {
		//详细日志
		handCardStr := string("发牌后手牌:")
		handCardStr += fmt.Sprintf("%s", sp.GameLogic.SwitchToCardName(sp.CardIndex[seat][:]))
		sp.OnWriteGameRecord(uint16(seat), handCardStr)
	}
	RepCardStr := string("牌堆牌: ")
	RepCardStr += fmt.Sprintf("%s", sp.GameLogic.SwitchToCardName5(sp.RepertoryCard[:], sp.LeftCardCount))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, RepCardStr)

	//发送扑克---这是发送给庄家的第26张牌
	sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount-1]
	sp.LeftCardCount--
	sp.CardIndex[sp.Banker][sp.GameLogic.SwitchToCardIndex(sp.SendCardData)]++

	//详细日志
	CardSt21 := string("庄家第26张牌: ")
	card21Idx := sp.GameLogic.SwitchToCardIndex(sp.SendCardData)
	CardSt21 += fmt.Sprintf("%s,%s", sp.GameLogic.SwitchToCardName4(card21Idx), sp.GameLogic.SwitchToCardName2(card21Idx))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, CardSt21)

	//设置变量
	sp.ProvideCard = 0
	sp.ProvideUser = static.INVALID_CHAIR
	sp.ReplayRecord.R_LeftCount = int(sp.LeftCardCount)

	for seat := 0; seat < sp.GetPlayerCount(); seat++ {
		//回放记录玩家手上初始牌
		_, tempCards := sp.GameLogic.SwitchToCardData3(sp.CardIndex[seat])
		tempCardsInt := [TCGZ_MAXSENDCARD]int{}
		for i := 0; i < TCGZ_MAXSENDCARD; i++ {
			tempCardsInt[i] = int(tempCards[i])
		}
		sp.ReplayRecord.R_HandCards[seat] = append(sp.ReplayRecord.R_HandCards[seat], tempCardsInt[:]...)
	}

	//构造数据,发送开始信息
	var GameStart static.Msg_S_ZP_GameStart
	GameStart.BankerUser = sp.Banker
	GameStart.CurrentUser = sp.CurrentUser
	GameStart.SendCard = sp.SendCardData       //发送庄家的最后一张牌
	GameStart.LeftCardCount = sp.LeftCardCount //剩余牌数目

	//向每个玩家发送数据
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//设置变量
		GameStart.MySeat = uint16(i)
		//GameStart.Overtime = sp.LimitTime
		GameStart.CardCount = TCGZ_MAXSENDCARD - 1
		if uint16(i) == sp.Banker {
			GameStart.CardCount = TCGZ_MAXSENDCARD //庄家多一张牌
		}
		_, tempCards := sp.GameLogic.SwitchToCardData3(sp.CardIndex[i])
		for c := 0; c < TCGZ_MAXSENDCARD; c++ {
			GameStart.CardData[c] = tempCards[c]
		}
		//记录玩家初始分

		//TODO 玩家分数设置
		sp.ReplayRecord.R_Score[i] = sp.Total[i]
		sp.ReplayRecord.UVitamin[_item.Uid] = _item.UserScoreInfo.Vitamin
		//发送数据
		sp.SendPersonMsg(consts.MsgTypeGameStart, GameStart, uint16(i))
		//发送旁观数据
		LookonItems := sp.GetLookonUserItemsByChair(uint16(i))
		if len(LookonItems) > 0 {
			for _, item := range LookonItems {
				if item != nil {
					sp.SendPersonLookonMsg(consts.MsgTypeGameStart, GameStart, item.Uid)
				}
			}
		}
	}
	//检查听牌，庄家不检查
	for i := 0; i < TCGZ_MAX_PLAYER && i < sp.GetPlayerCount(); i++ {
		if uint16(i) != sp.Banker {
			tingcnt, tingresult := sp.CheckTing(uint16(i))
			if tingcnt > 0 {
				sp.SendTingInfo(uint16(i), tingresult)
			}
		}
	}

	sp.ProvideUser = sp.CurrentUser
	sp.ProvideCard = sp.SendCardData
	//sp.PowerStartTime = time.Now().Unix() //权限开始时间
	//3s后开始天胡判断
	sp.GuanStartUser = uint16(sp.GetFrontSeat(sp.CurrentUser)) //庄家的上家先检查
	sp.CurrentUser = sp.GuanStartUser                          //修改当前玩家
	sp.GameState = ZP_AS_GUANSHENG
	sp.setLimitedTime(1)

	// 上一局托管 第二局直接进入托管
	for seat, v := range sp.TuoGuanPlayer {
		if v == true {
			sp.AutoTuoGuan(uint16(seat))
		}
	}
}

func (sp *Sport_zp_jzhp) setLimitedTime(iLimitTime int64) {
	// fmt.Println(fmt.Sprintf("limitetimeOP(%d)", sp.Rule.limitetimeOP))
	sp.LimitTime = time.Now().Unix() + iLimitTime
	sp.GameTimer.SetLimitTimer(int(iLimitTime))
}

func (sp *Sport_zp_jzhp) freeLimitedTime() {
	sp.GameTimer.KillLimitTimer()
}

// 计时器事件
func (sp *Sport_zp_jzhp) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	//游戏定时器
	if dwTimerID == components2.GameTime_Nine {
		sp.OnAutoOperate(true)
	}
	return true
}

// 暂时空着
func (sp *Sport_zp_jzhp) OnAutoOperate(bBreakin bool) {
	if sp.GameState == ZP_AS_GUANSHENG {
		if !sp.IsTongAction {
			//先进行观生操作
			sp.JudgeGuanSheng()
		} else {
			if sp.Rule.Overtime_trust <= 0 && sp.OutCardDismissTime <= 0 {
				return
			}
			if sp.OutCardDismissTime > 0 {
				sp.OnGameOver(sp.ResumeUser, static.GER_DISMISS)
				sp.OnWriteGameRecord(sp.ResumeUser, "OnAutoOperate 超时操作强制解散!!!")
				return
			}
			//其实可以新增一个阶段来解决重复状态的问题!!!!!!!!!
			//先进行观生操作超时
			// 牌权处理判断是否托管 多个用户同时等待 取最长定时器
			if int(sp.ResumeUser) < sp.GetPlayerCount() {
				if sp.UserAction[sp.ResumeUser] != ZP_WIK_NULL && !sp.TuoGuanPlayer[sp.ResumeUser] && sp.Response[sp.ResumeUser] == false {
					sp.AutoTuoGuan(sp.ResumeUser)
				}
				item := sp.GetUserItemByChair(sp.ResumeUser)
				if item != nil {
					if sp.UserAction[sp.ResumeUser] != ZP_WIK_NULL && sp.TuoGuanPlayer[sp.ResumeUser] == true {
						sp.OnWriteGameRecord(sp.ResumeUser, "发牌后的统牌超时,系统自动弃")
						var _operatecard_msg static.Msg_C_ZP_OperateCard
						_operatecard_msg.Id = item.GetUserID()
						_operatecard_msg.Code = ZP_WIK_NULL
						_operatecard_msg.Card = 0
						_operatecard_msg.WeaveCount = 0
						_operatecard_msg.WeaveInfo = [10][10]byte{}
						_operatecard_msg.WeaveKind = [10]int{}
						sp.OnUserOperateCard(&_operatecard_msg)
					} else {
						sp.OnWriteGameRecord(sp.ResumeUser, "发牌后的统牌超时 error0")
					}
				} else {
					sp.OnWriteGameRecord(sp.ResumeUser, "发牌后的统牌超时 error1")
				}
			} else {
				sp.OnWriteGameRecord(static.INVALID_CHAIR, "发牌后的统牌超时 error2")
			}
		}
	} else if sp.GameState == ZP_AS_STARTPLAY {
		//观生结束，庄家开始出牌
		if int(sp.CurrentUser) < sp.GetPlayerCount() {
			sp.GameState = ZP_AS_PLAYCARD //修改牌权阶段类型
			// 超时超时时间
			iOvertime := sp.OperateTime
			if sp.TuoGuanPlayer[sp.CurrentUser] == true && iOvertime > sp.AutoOutTime {
				iOvertime = sp.AutoOutTime
			}
			//begin:bugid:17960，在17960基础上又还原了,庄家这时候不能在观或胡了
			sp.UserAction = [TCGZ_MAX_PLAYER]int{}
			var s_Power static.Msg_S_ZP_SendPower
			s_Power.CurrentUser = sp.CurrentUser
			s_Power.LeftTime = int(iOvertime)
			sp.SendTableMsg(consts.MsgTypeSendPower, s_Power)
			sp.OnWriteGameRecord(sp.CurrentUser, "统牌结束，庄家开始出牌")

			sp.PowerStartTime = time.Now().Unix() //权限开始时间

			sp.GameState = ZP_AS_PLAYCARD // 更新为要打牌状态
			sp.setLimitedTime(int64(iOvertime + 1))

			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeSendPower, s_Power)
			////设置变量，在次检查庄家能不能观和胡
			//sp.SendStatus =true;//自摸类型的
			//sp.ProvideUser =sp.CurrentUser;
			//sp.ProvideCard =sp.SendCardData;
			//sp.JudgeDispatch(sp.CurrentUser);
			//end:bugid:17960
		} else {
			sp.OnWriteGameRecord(sp.CurrentUser, "牌权错误!!!")
		}
	} else if sp.GameState == ZP_AS_SENDCARD {
		// 发牌
		if sp.LeftCardCount > 0 {
			sp.DispatchCardData(sp.CurrentUser) // 发牌
		} else // 否则游戏结束
		{
			sp.ChiHuCard = 0
			sp.ProvideUser = static.INVALID_CHAIR
			sp.OnEventGameEnd(static.INVALID_CHAIR, static.GER_NORMAL)
		}
	} else if sp.GameState == ZP_AS_PLAYCARD {
		if sp.Rule.Overtime_trust <= 0 && sp.OutCardDismissTime <= 0 {
			return
		}

		if sp.CurrentUser == static.INVALID_CHAIR {
			if sp.OutCardDismissTime > 0 {
				// 获取最大牌权用户
				wTargetUser := sp.GetMaxPowerUser()
				sp.OnGameOver(wTargetUser, static.GER_DISMISS)
				sp.OnWriteGameRecord(wTargetUser, "OnAutoOperate ZP_AS_PLAYCARD 2 超时操作强制解散!!!")
				return
			}
			// 牌权处理判断是否托管 多个用户同时等待 取最长定时器
			for seat := 0; seat < sp.GetPlayerCount(); seat++ {
				if sp.UserAction[seat] != ZP_WIK_NULL && !sp.TuoGuanPlayer[seat] && sp.Response[seat] == false {
					sp.AutoTuoGuan(uint16(seat))
				}
			}

			// 处理牌权 系统做 弃 处理
			for seat := 0; seat < sp.GetPlayerCount(); seat++ {
				item := sp.GetUserItemByChair(uint16(seat))
				if item == nil {
					continue
				}
				if sp.UserAction[seat] != ZP_WIK_NULL && sp.TuoGuanPlayer[seat] == true {
					sp.OnWriteGameRecord(uint16(seat), "托管自动操作")
					var _operatecard_msg static.Msg_C_ZP_OperateCard
					_operatecard_msg.Id = item.GetUserID()
					_operatecard_msg.Code = 0
					_operatecard_msg.Card = 0
					if sp.GameType == meta2.GT_ROAR && (sp.UserAction[seat]&ZP_WIK_CHI_HU) != ZP_WIK_NULL {
						//海底捞时要自动胡
						_operatecard_msg.Code = ZP_WIK_CHI_HU
						_operatecard_msg.Card = sp.ProvideCard
					}
					_operatecard_msg.WeaveCount = 0
					_operatecard_msg.WeaveInfo = [10][10]byte{}
					_operatecard_msg.WeaveKind = [10]int{}
					sp.OnUserOperateCard(&_operatecard_msg)
				}
			}
		} else {
			if sp.OutCardDismissTime > 0 {
				sp.OnGameOver(sp.CurrentUser, static.GER_DISMISS)
				sp.OnWriteGameRecord(sp.CurrentUser, "OnAutoOperate ZP_AS_PLAYCARD 超时操作强制解散!!!")
				return
			}

			// 出牌处理
			if sp.CurrentUser >= TCGZ_MAX_PLAYER {
				return
			}
			// 出牌判断 是否托管
			if !sp.TuoGuanPlayer[sp.CurrentUser] {
				sp.AutoTuoGuan(sp.CurrentUser)
			}
			sp.OnWriteGameRecord(sp.CurrentUser, "托管出牌")

			// 牌权处理之后由系统自动出牌
			var _outcard_msg static.Msg_C_ZP_OutCard
			_outcard_msg.ChairID = sp.CurrentUser
			_outcard_msg.ByClient = false
			_outcard_msg.Jian = false
			_, tempCards := sp.GameLogic.SwitchToCardData3(sp.CardIndex[sp.CurrentUser])
			if sp.SendStatus {
				_outcard_msg.CardData = sp.SendCardData
			}
			if !sp.GameLogic.IsValidCard(_outcard_msg.CardData) || sp.CardIndex[sp.CurrentUser][sp.GameLogic.SwitchToCardIndex(_outcard_msg.CardData)] <= 0 {
				for i := len(tempCards) - 1; i >= 0; i-- {
					if tempCards[i] <= 0 {
						continue
					}
					_outcard_msg.CardData = tempCards[i]
					break
				}
			}
			sp.OnUserOutCard(&_outcard_msg, 1)
		}
	}
}

func (sp *Sport_zp_jzhp) AutoTuoGuan(theSeat uint16) int {
	var msgtg static.Msg_S_ZP_Trustee
	msgtg.Trustee = true
	msgtg.ChairID = theSeat

	if sp.GameState == meta2.GsNull || theSeat >= TCGZ_MAX_PLAYER {
		return 0
	}

	//详细日志
	LogStr := fmt.Sprintf("超时托管,CMD_S_Tuoguan_CB AutoTuoGuan msgtg.theFlag=%t msgtg.theSeat=%d ", msgtg.Trustee, msgtg.ChairID)
	sp.OnWriteGameRecord(theSeat, LogStr)

	if sp.GameState == ZP_AS_PLAYCARD || sp.GameState == ZP_AS_GUANSHENG {
		if true == msgtg.Trustee {
			sp.TuoGuanPlayer[theSeat] = true
			sp.TrustCounts[theSeat]++
			if theSeat == sp.CurrentUser { //如果是当前的玩家，那么重新设置一下开始时间
				//sp.setLimitedTime(int64(sp.AutoOutTime))//已经超时了，马上就要切换牌权了，不用在设置他的时间了
			}
			sp.SendTableMsg(consts.MsgTypeGameTrustee, msgtg)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			sp.addReplayOrder(msgtg.ChairID, E_ZP_Tuo_Guan, 1, []int{})
		} else {
			sp.TuoGuanPlayer[theSeat] = false
			sp.SendTableMsg(consts.MsgTypeGameTrustee, msgtg)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			sp.addReplayOrder(msgtg.ChairID, E_ZP_Tuo_Guan, 0, []int{})
		}
	}
	return 1
}

// 发送操作
func (sp *Sport_zp_jzhp) SendOperateNotify(iWaitTime int) bool {
	//发送提示
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if sp.UserAction[i] != ZP_WIK_NULL {
			if sp.GameType == meta2.GT_ROAR && (sp.UserAction[i]&ZP_WIK_CHI_HU) != ZP_WIK_NULL {
				//海底捞时能胡时不能弃
				sp.UserAction[i] |= ZP_WIK_FILL
			}
			//构造数据
			var OperateNotify static.Msg_S_ZP_OperateNotify
			OperateNotify.ResumeUser = sp.ResumeUser
			OperateNotify.ActionCard = sp.ProvideCard //抢暗杠时，复用此字段，表示轮到谁抢了
			OperateNotify.ActionMask = sp.UserAction[i]
			OperateNotify.LeftTime = iWaitTime
			OperateNotify.ClockSeat = uint16(i)

			//发送数据
			sp.SendPersonMsg(consts.MsgTypeGameOperateNotify, OperateNotify, uint16(i))

			//游戏记录
			szGameRecord := fmt.Sprintf("发送牌权：0x%x , %s", sp.UserAction[i], sp.GetWikStr(sp.UserAction[i]))
			sp.OnWriteGameRecord(uint16(i), szGameRecord)

			sp.PowerStartTime = time.Now().Unix() //权限开始时间

			// 回放记录中记录牌权显示
			if sp.UserAction[i] > 0 {
				sp.addReplayOrder(uint16(i), E_SendCardRight, sp.UserAction[i], []int{})
			}
		} else {
			//构造数据
			var OperateNotify static.Msg_S_ZP_OperateNotify
			OperateNotify.LeftTime = iWaitTime
			OperateNotify.ClockSeat = sp.ProvideUser
			if int(sp.ProvideUser) >= sp.GetPlayerCount() {
				OperateNotify.ClockSeat = sp.ResumeUser
			}

			//发送数据
			sp.SendPersonMsg(consts.MsgTypeGameOperateNotify, OperateNotify, uint16(i))
		}
	}
	return true
}

// 用户出牌
func (sp *Sport_zp_jzhp) OnUserOutCard(msg *static.Msg_C_ZP_OutCard, byAutoFlag byte) bool {
	xlog.Logger().Info("OnUserOutCard")
	//效验状态
	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}
	if sp.GameEndStatus != static.GS_MJ_PLAY {
		return true
	}
	if sp.GameType == meta2.GT_ROAR {
		sp.OnWriteGameRecord(msg.ChairID, "海底捞时不能出牌")
		return false
	}
	//v1.3新需求，在发牌给自己，在有统的操作牌权时也可以直接出牌，默认弃操作。但这时self.CurrentUser是无效值需要处理下
	currenuser := sp.CurrentUser
	if sp.SendStatus && msg.ChairID < uint16(sp.GetPlayerCount()) && (sp.UserAction[msg.ChairID]&ZP_WIK_TIANLONG) != 0 {
		currenuser = sp.ResumeUser
	}
	if currenuser >= TCGZ_MAX_PLAYER || currenuser >= uint16(sp.GetPlayerCount()) {
		return false
	}
	if sp.GameState != ZP_AS_PLAYCARD {
		return false
	}
	wChairID := msg.ChairID
	cbCardData := msg.CardData
	//效验参数
	if wChairID != currenuser {
		//详细日志
		LogStr := fmt.Sprintf("座位号 %d, OnUserOutCard 出牌玩家不是当前玩家 ", wChairID)
		sp.OnWriteGameRecord(wChairID, LogStr)
		return false
	}
	if sp.GameLogic.IsValidCard(cbCardData) == false {
		sp.OnWriteGameRecord(wChairID, "出牌无效")
		return false
	}
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}
	if sp.CheckJianCur(wChairID, cbCardData) {
		//游戏记录
		szGameRecord := fmt.Sprintf("打出：%s,捡的牌，本次不能出", sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbCardData)))
		sp.OnWriteGameRecord(wChairID, szGameRecord)
		return false
	}
	if sp.NoOutUser == wChairID {
		sp.OnWriteGameRecord(wChairID, "玩家已经选择了捏牌，不允许出牌")
	}
	//v1.5 20200713 统过的牌不能打出，除非能换统
	//判断能不能换统
	_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wChairID])
	index := sp.GameLogic.SwitchToCardIndexNoHua(cbCardData)
	if index < TCGZ_MAX_INDEX && byTempCardIndex[index] >= 1 {
		_, realTongCntBef, _ := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wChairID].TongCnt)
		byTempCardIndex[index] -= 1
		_, realTongCnt, tongInfo := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wChairID].TongCnt)
		if realTongCnt < sp.TongInfo[wChairID].TongCnt {
			if realTongCntBef > realTongCnt {
				//因为这张牌导致的统数减少
				//v1.7 20200723 不能换统也可以打了
				//sp.OnWriteGameRecord(wChairID, "打牌数据有误,不能换统时，不能打");
				//return false
			}
		}
		//v1.5 20200715 泛不换统后，需要确定换到哪里去了
		if realTongCntBef > realTongCnt {
			//随便找一个替换
			for j := 0; j < TCGZ_MAX_INDEX; j++ {
				if tongInfo[j] > sp.TongInfo[wChairID].CardTongInfo[j].TongCnt {
					sp.TongInfo[wChairID].CardTongInfo[j].TongCnt++
					sp.TongInfo[wChairID].CardTongInfo[index].TongCnt--
					sp.OnWriteGameRecord(wChairID, fmt.Sprintf("打牌换统到：%s", sp.GameLogic.SwitchToCardName2(byte(j))))
					break
				}
			}
		}
	}
	//删除扑克
	if msg.Jian {
		if !sp.DeleteJian(wChairID, cbCardData) {
			return false
		}
		sp.OnWriteGameRecord(wChairID, "玩家出的这张牌是以前捡的牌")
	} else {
		bret := false
		bret, sp.CardIndex[wChairID] = sp.GameLogic.RemoveCard(sp.CardIndex[wChairID], cbCardData)
		if bret == false {
			return false
		}
	}

	//打出牌后，把当前轮的捡牌删除
	sp.UserJianCardsCur[wChairID] = []byte{}

	//游戏记录
	szGameRecord := fmt.Sprintf("牌型：%s，打出：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wChairID][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbCardData)))
	sp.OnWriteGameRecord(wChairID, szGameRecord)

	//设置变量
	sp.SendStatus = false
	sp.UserAction[wChairID] = ZP_WIK_NULL
	sp.PerformAction[wChairID] = ZP_WIK_NULL

	sp.DismissGuoZhang(gKind_HuCard, wChairID, 0)   //出牌属于变动手牌了，解除胡牌的过张
	sp.DismissGuoZhang(gKind_PengCard, wChairID, 0) //出牌属于变动手牌了，解除碰牌的过张

	//出牌记录
	sp.OutCardCount++
	sp.OutCardUser = wChairID
	sp.OutCardData = cbCardData

	//构造数据
	var OutCard static.Msg_S_ZP_OutCard
	OutCard.OutCardUser = wChairID
	OutCard.OutCardData = cbCardData
	if byAutoFlag > 0 {
		OutCard.IsAutoOut = 1
	}
	////更新实际的统信息
	//_,byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wChairID])
	//_,realTongCnt:=sp.GameLogic.CheckTong_Op(byTempCardIndex,sp.TongInfo[wChairID].TongCnt)
	//if false && realTongCnt < sp.TongInfo[wChairID].TongCnt{
	//	//这里是false，因为需求时，在有统时需要补齐这次减少的统
	//	sp.TongInfo[wChairID].TongCnt--
	//	//20200114
	//	cardIdx := sp.GameLogic.SwitchToCardIndexNoHua(cbCardData)
	//	if sp.GameLogic.IsValidCard(cbCardData) && cardIdx < TCGZ_MAX_INDEX{
	//		if sp.TongInfo[wChairID].CardTongInfo[cardIdx].TongCnt >0 {
	//			sp.TongInfo[wChairID].CardTongInfo[cardIdx].TongCnt-- //记录那张牌统过
	//		}else {
	//			//随便找一个替换
	//			for j := 0;j<TCGZ_MAX_INDEX;j++{
	//				if sp.TongInfo[wChairID].CardTongInfo[j].TongCnt > 0{
	//					sp.TongInfo[wChairID].CardTongInfo[j].TongCnt--
	//				}
	//			}
	//		}
	//	}
	//}

	//记录出牌
	sp.addReplayOrder(wChairID, E_ZP_OutCard, int(cbCardData), []int{})
	if msg.Jian {
		sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_Jian, int(cbCardData))
	}

	//查听
	tingcnt, tingresult := sp.CheckTing(wChairID)
	if tingcnt > 0 {
		sp.SendTingInfo(wChairID, tingresult)
	}

	//用户切换
	sp.ProvideUser = wChairID
	sp.ProvideCard = cbCardData
	sp.CurrentUser = uint16(sp.GetNextSeat(wChairID))

	//响应判断，如果用户出的是一般牌，判断其他用户是否需要该牌，EstimatKind_OutCard只是正常出牌判断
	//如果当前用户自己 出了牌，不能自己对自己进行分析吃，碰杠
	bAroseAction := sp.EstimateUserRespond(wChairID, cbCardData, ZP_EstimatKind_OutCard)

	//客户端需要知道这张牌有没有人要
	OutCard.ResponseFlag = bAroseAction
	sp.SendTableMsg(consts.MsgTypeGameOutCard, OutCard)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameOutCard, OutCard)

	if bAroseAction == false {
		sp.StartHD()
		sp.GameState = ZP_AS_SENDCARD
		sp.setLimitedTime(1) //牌没有人要，需要1s后在发下一张牌

	}

	return true
}
func (sp *Sport_zp_jzhp) StartHD() bool {
	if sp.GetValidCardCount() == 0 {
		return false //不是海底捞，继续发牌
	} else if sp.GetValidCardCount() == 1 {
		sp.OnWriteGameRecord(sp.CurrentUser, fmt.Sprintf("【海底捞】"))
		sp.GameType = meta2.GT_ROAR //海底捞
		return false
	}
	return false
}

// bool型的返回值不够用了，0表示大于3，1表示等于3，其它表示小于3
func (sp *Sport_zp_jzhp) GetValidCardCount() int {
	if sp.LeftCardCount > byte(sp.GetPlayerCount()) {
		return 0
	} else if sp.LeftCardCount <= byte(sp.GetPlayerCount()) {
		return 1
	}
	return 2
}

// 派发扑克
func (sp *Sport_zp_jzhp) BuCardData(wCurrentUser uint16) bool {
	//状态效验
	if wCurrentUser == static.INVALID_CHAIR || int(wCurrentUser) >= sp.GetPlayerCount() {
		return false
	}

	//剩余牌校验
	if sp.LeftCardCount <= 0 {
		return false
	}

	sp.SendStatus = true
	//发牌处理
	//发送扑克
	sp.SendCardCount++
	sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount-1]
	sp.LeftCardCount--
	sp.CardIndex[wCurrentUser][sp.GameLogic.SwitchToCardIndex(sp.SendCardData)]++

	//牌型
	dispStr := fmt.Sprintf("牌型：%s，发来：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wCurrentUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(sp.SendCardData)))
	sp.OnWriteGameRecord(wCurrentUser, dispStr)

	//记录发牌
	sp.addReplayOrder(wCurrentUser, E_ZP_SendCard, int(sp.SendCardData), []int{})

	//设置变量
	sp.CurrentUser = wCurrentUser

	// 超时超时时间
	iOvertime := sp.OperateTime
	if sp.TuoGuanPlayer[sp.CurrentUser] == true && iOvertime > sp.AutoOutTime {
		iOvertime = sp.AutoOutTime
	}

	//构造数据
	var SendCard static.Msg_S_ZP_SendCard
	SendCard.CurrentUser = wCurrentUser
	SendCard.ActionMask = 0
	SendCard.CardData = sp.SendCardData
	SendCard.LeftTime = int(iOvertime)
	sp.SendTableMsg(consts.MsgTypeGameSendCard, SendCard)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameSendCard, SendCard)
	//设置变量
	sp.ProvideUser = wCurrentUser
	sp.ProvideCard = sp.SendCardData

	return true
}

// 派发扑克
func (sp *Sport_zp_jzhp) DispatchCardData(wCurrentUser uint16) bool {
	//状态效验
	if wCurrentUser == static.INVALID_CHAIR || int(wCurrentUser) >= sp.GetPlayerCount() {
		return false
	}

	//剩余牌校验
	if sp.LeftCardCount <= 0 {
		return false
	}

	//已经过了一轮了，上轮捡牌标记删除
	sp.UserJianCardsCur[wCurrentUser] = []byte{}

	//丢弃扑克DYZP_MAX_PLAYER
	if (sp.OutCardUser != static.INVALID_CHAIR && int(sp.OutCardUser) < sp.GetPlayerCount()) && (sp.OutCardData != 0) {
		sp.DiscardCount[sp.OutCardUser]++
		sp.DiscardCard[sp.OutCardUser][sp.DiscardCount[sp.OutCardUser]-1] = sp.OutCardData
		sp.OutCardUser = static.INVALID_CHAIR
		sp.OutCardData = 0
	}

	sp.SendStatus = true
	//发牌处理
	//发送扑克
	sp.SendCardCount++
	sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount-1]
	sp.LeftCardCount--
	sp.CardIndex[wCurrentUser][sp.GameLogic.SwitchToCardIndex(sp.SendCardData)]++

	//牌型
	dispStr := fmt.Sprintf("牌型：%s，发来：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wCurrentUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(sp.SendCardData)))
	sp.OnWriteGameRecord(wCurrentUser, dispStr)

	//记录发牌
	sp.addReplayOrder(wCurrentUser, E_ZP_SendCard, int(sp.SendCardData), []int{})

	//设置变量
	sp.CurrentUser = wCurrentUser

	// 超时超时时间
	iOvertime := sp.OperateTime
	if sp.TuoGuanPlayer[sp.CurrentUser] == true && iOvertime > sp.AutoOutTime {
		iOvertime = sp.AutoOutTime
	}

	//构造数据
	var SendCard static.Msg_S_ZP_SendCard
	SendCard.CurrentUser = wCurrentUser
	SendCard.ActionMask = 0
	SendCard.CardData = sp.SendCardData
	SendCard.LeftTime = int(iOvertime)
	sp.SendTableMsg(consts.MsgTypeGameSendCard, SendCard)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameSendCard, SendCard)

	sp.DismissGuoZhang(gKind_HuCard, wCurrentUser, 0)   //发牌属于变动手牌了，解除胡牌的过张
	sp.DismissGuoZhang(gKind_PengCard, wCurrentUser, 0) //发牌属于变动手牌了，解除碰牌的过张

	//设置变量
	sp.ProvideUser = wCurrentUser
	sp.ProvideCard = sp.SendCardData

	sp.JudgeDispatch(sp.CurrentUser)
	return true
}

// 是否发牌时可以胡
func (sp *Sport_zp_jzhp) JudgeDispatch(wSeat uint16) {
	if wSeat >= uint16(sp.GetPlayerCount()) {
		return
	}

	sp.UserAction = [TCGZ_MAX_PLAYER]int{}

	if sp.NoOutUser == wSeat {
		//直接给下家发牌
		sp.OnWriteGameRecord(wSeat, "玩家捏牌，不出牌")
		//当前正要出牌
		sp.CurrentUser = uint16(sp.GetNextSeat(wSeat))
		sp.StartHD()
		sp.GameState = ZP_AS_SENDCARD
		sp.setLimitedTime(1) //给下家发牌，需要1s后在发下一张牌
		return
	}

	_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[sp.CurrentUser])

	//杠牌判断
	userAction := 0
	userAction, _ = sp.GameLogic.AnalyseGangCard(byTempCardIndex, sp.TongInfo[sp.CurrentUser].TongCnt)
	sp.UserAction[wSeat] |= userAction

	//踏牌判断
	userAction = 0
	userAction, _ = sp.GameLogic.AnalyseTaCard(byTempCardIndex, sp.WeaveItemArray[sp.CurrentUser], sp.WeaveItemCount[sp.CurrentUser])
	sp.UserAction[wSeat] |= userAction

	//胡牌判断
	if (sp.JudgeHu(wSeat) & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
		sp.OnWriteGameRecord(wSeat, "发牌后可以胡")
		sp.UserAction[wSeat] |= ZP_WIK_CHI_HU
	}

	sp.GameState = ZP_AS_PLAYCARD // 更新为要打牌状态
	// 超时超时时间
	iOvertime := sp.OperateTime
	if sp.TuoGuanPlayer[wSeat] == true && iOvertime > sp.AutoOutTime {
		iOvertime = sp.AutoOutTime
	}
	sp.setLimitedTime(int64(iOvertime + 1))

	if sp.UserAction[wSeat] != ZP_WIK_NULL {
		sp.ResumeUser = sp.CurrentUser
		sp.CurrentUser = static.INVALID_CHAIR
		//发送提示
		sp.SendOperateNotify(int(iOvertime))
	} else {
		if sp.GameType != meta2.GT_ROAR {
			var s_Power static.Msg_S_ZP_SendPower
			s_Power.CurrentUser = wSeat
			s_Power.LeftTime = int(iOvertime)
			sp.SendTableMsg(consts.MsgTypeSendPower, s_Power)
			sp.CurrentUser = wSeat

			sp.OnWriteGameRecord(wSeat, "拥有出牌牌权")
		} else {
			//直接给下家发牌
			sp.CurrentUser = uint16(sp.GetNextSeat(wSeat))
			sp.StartHD()
			sp.GameState = ZP_AS_SENDCARD
			sp.setLimitedTime(1) //给下家发牌，需要1s后在发下一张牌
		}
	}
}

// 是否有观生，有观生要发送给客户端
func (sp *Sport_zp_jzhp) JudgeGuanSheng() {
	sp.IsTongAction = false
	sp.UserAction = [TCGZ_MAX_PLAYER]int{}
	if int(sp.CurrentUser) >= sp.GetPlayerCount() {
		return
	}
	sp.ProvideUser = sp.CurrentUser //整个观生阶段都是自摸类型的

	_, rcvIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[sp.CurrentUser])
	//杠牌判断
	userAction := 0
	userAction, _ = sp.GameLogic.AnalyseGangCard(rcvIndex, sp.TongInfo[sp.CurrentUser].TongCnt)
	sp.UserAction[sp.CurrentUser] |= userAction

	//滑牌判断
	userAction = 0
	//userAction, _= sp.GameLogic.AnalyseHuaCard(rcvIndex)
	sp.UserAction[sp.CurrentUser] |= userAction

	if (sp.UserAction[sp.CurrentUser] & ZP_WIK_TIANLONG) != ZP_WIK_NULL {
		sp.OnWriteGameRecord(sp.CurrentUser, "发牌后可以统")
	}
	if (sp.UserAction[sp.CurrentUser] & ZP_WIK_HUA) != ZP_WIK_NULL {
		sp.OnWriteGameRecord(sp.CurrentUser, "发牌后可以滑")
	}
	if sp.CurrentUser == sp.Banker {
		//判断胡牌
		if (sp.JudgeHu(sp.CurrentUser) & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
			sp.OnWriteGameRecord(sp.CurrentUser, "发牌后可以天胡")
			sp.SendStatus = true //自摸类型的
			sp.UserAction[sp.CurrentUser] |= ZP_WIK_CHI_HU
		}
	}
	if sp.UserAction[sp.CurrentUser] != ZP_WIK_NULL {
		// 超时超时时间
		iOvertime := sp.OperateTime
		if sp.TuoGuanPlayer[sp.CurrentUser] == true && iOvertime > sp.AutoOutTime {
			iOvertime = sp.AutoOutTime
		}
		sp.IsTongAction = true
		sp.setLimitedTime(iOvertime + 1)

		sp.ResumeUser = sp.CurrentUser
		sp.CurrentUser = static.INVALID_CHAIR
		//发送提示
		sp.SendOperateNotify(int(iOvertime))
		return
	}

	if int(sp.CurrentUser) < sp.GetPlayerCount() {
		sp.CurrentUser = uint16(sp.GetFrontSeat(sp.CurrentUser)) //上家
	} else {
		//无效时直接观生结束
		sp.CurrentUser = sp.Banker //
		sp.GameState = ZP_AS_STARTPLAY
	}

	sp.ProvideUser = sp.CurrentUser //整个观生阶段都是自摸类型的 //切换牌权后重新赋值

	if sp.CurrentUser == sp.GuanStartUser {
		//观生结束
		sp.CurrentUser = sp.Banker //
		sp.GameState = ZP_AS_STARTPLAY
	} else {
	}
	//go 框架不支持时间为0的定时器 ,直接调用
	sp.OnAutoOperate(false)
}

// 是否有天胡，庄家发牌后观生，需要判断是否能天胡
func (sp *Sport_zp_jzhp) JudgeHu(wSeat uint16) int {
	//原始牌数据，恢复花牌需要用到这个数组
	CardIndexSrc := [TCGZ_MAX_INDEX_HUA]byte{}
	for c := byte(0); c < TCGZ_MAX_INDEX_HUA; c++ {
		CardIndexSrc[c] = sp.CardIndex[wSeat][c]
	}

	wUserAction := 0
	//胡牌判断
	wChiHuRight := 0
	_, rcvIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wSeat])
	wUserAction, sp.ChiHuResult[wSeat] = sp.GameLogic.AnalyseChiHuCard(CardIndexSrc, rcvIndex, sp.WeaveItemArray[wSeat][:], sp.WeaveItemCount[wSeat], 0, wChiHuRight)
	if (wUserAction & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
		//v1.5 20200715 是否满足换统
		wUserAction, sp.ChiHuResult[wSeat] = sp.CheckHu2(wSeat, 0, sp.ChiHuResult[wSeat], sp.TongInfo[wSeat].TongCnt)
	}
	//需要把WeaveItemInfo补齐就行
	if (wUserAction & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
		//胡息计算
		iMaxIndex, iMaxHuXi, realHuXiInfo := sp.GameLogic.CalculateHuXi(&sp.ChiHuResult[wSeat])
		sp.GameLogic.CalculateHuXi2(&sp.ChiHuResult[wSeat], &iMaxIndex, &iMaxHuXi, &realHuXiInfo)

		//是否满足胡息要求
		bCanHu := false
		if iMaxHuXi >= sp.GeziShu {
			bCanHu = true
		}
		if !bCanHu {
			//不能胡时，去掉胡牌权限
		} else {
			return ZP_WIK_CHI_HU
		}
	}
	return 0
}

// 获取玩家的牌，包括组合牌，观生牌，捡的牌，其他手牌
func (sp *Sport_zp_jzhp) GetAllCardsbyChairid(wSeat uint16, cbCenterCard byte) (byte, [static.MAX_CARD]byte) {
	retCardsData := [static.MAX_CARD]byte{}
	if int(wSeat) > sp.GetPlayerCount() {
		return 0, [static.MAX_CARD]byte{}
	}
	cnt := byte(0)
	num, handCardsData := sp.GameLogic.SwitchToCardData3(sp.CardIndex[wSeat])
	for i := byte(0); i < num; i++ {
		card := handCardsData[i]
		if sp.GameLogic.IsValidCard(card) {
			retCardsData[i] = card
			cnt++
		}
	}
	for i := 0; i < 10; i++ {
		for j := 0; j < 5; j++ {
			card := sp.UserGuanCards[wSeat][i][j]
			if sp.GameLogic.IsValidCard(card) {
				retCardsData[cnt] = card
				cnt++
			}
		}
	}
	//捡的牌要添加到rcvIndex
	for _, jcard := range sp.UserJianCards[wSeat] {
		if sp.GameLogic.IsValidCard(jcard) {
			retCardsData[cnt] = jcard
			cnt++
		}
	}
	for byChiHuCnt := 0; byChiHuCnt < len(sp.WeaveItemArray[wSeat]); byChiHuCnt++ {
		for j := 0; j < 5; j++ {
			card := sp.WeaveItemArray[wSeat][byChiHuCnt].Cards[j]
			if sp.GameLogic.IsValidCard(card) {
				retCardsData[cnt] = card
				cnt++
			}
		}
	}
	if sp.GameLogic.IsValidCard(cbCenterCard) {
		retCardsData[cnt] = cbCenterCard
		cnt++
	}
	return cnt, retCardsData
}

// 获取最大牌权用户，参考建始楚胡
func (sp *Sport_zp_jzhp) GetMaxPowerUser() uint16 {
	// 最大牌权用户
	maxPowerUser := static.INVALID_CHAIR
	maxActionRank := 0

	// 从当前用户开始遍历 sp.CurrentUser
	iNextPlayer := sp.ProvideUser
	for i := 0; i < sp.GetPlayerCount() && int(iNextPlayer) < sp.GetPlayerCount(); i++ {
		// 没有牌权 用户不参与比较
		if sp.UserAction[iNextPlayer] == ZP_WIK_NULL {
			iNextPlayer = uint16(sp.GetNextSeat(iNextPlayer))
			continue
		}

		// 弃权 用户不再参与比较
		if sp.Response[iNextPlayer] && sp.PerformAction[iNextPlayer] == ZP_WIK_NULL {
			iNextPlayer = uint16(sp.GetNextSeat(iNextPlayer))
			continue
		}

		// 牌权等级
		cbUserActionRank := 0
		if sp.Response[iNextPlayer] {
			cbUserActionRank = sp.GameLogic.GetUserActionRank(sp.PerformAction[iNextPlayer])
		} else {
			cbUserActionRank = sp.GameLogic.GetUserActionRank(sp.UserAction[iNextPlayer])
		}

		// 查找最大牌权用户
		if cbUserActionRank > maxActionRank {
			maxActionRank = cbUserActionRank
			maxPowerUser = iNextPlayer
		}

		// 轮询下一位玩家
		iNextPlayer = uint16(sp.GetNextSeat(iNextPlayer))
	}

	return maxPowerUser
}

// ! 用户操作牌
func (sp *Sport_zp_jzhp) OnUserOperateCard(s_C_OperateData *static.Msg_C_ZP_OperateCard) bool {

	wChairID := sp.GetChairByUid(s_C_OperateData.Id)

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}
	if sp.GameEndStatus != static.GS_MJ_PLAY {
		return true
	}

	//效验用户
	if (wChairID != sp.CurrentUser) && (sp.CurrentUser != static.INVALID_CHAIR) {
		if sp.CheckRightAllReleased() {
			sp.OnWriteGameRecord(wChairID, "最高级别牌权玩家已经提前操作了，这条日志说明你的网络比别人慢了一丢丢")
			return true
		}
		return true
	}

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	wOperateCode := s_C_OperateData.Code
	cbOperateCard := s_C_OperateData.Card

	//观生阶段的操作需要先处理
	if sp.GameState == ZP_AS_GUANSHENG && (wOperateCode == ZP_WIK_TIANLONG || wOperateCode == ZP_WIK_HUA || wOperateCode == ZP_WIK_NULL) {
		if sp.UserOperate_GuanSheng_bef(wChairID, cbOperateCard, wOperateCode) {
			return true
		}
	}

	//被动动作,
	if sp.CurrentUser == static.INVALID_CHAIR {
		//效验状态
		if sp.Response[wChairID] == true {
			sp.OnWriteGameRecord(wChairID, "重复操作了牌权")
			return false
		}

		if (wOperateCode != ZP_WIK_NULL) && ((sp.UserAction[wChairID] & wOperateCode) == 0) {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "玩家操作了自己没有的牌权")
			if sp.CheckRightAllReleased() {
				sp.OnWriteGameRecord(wChairID, "最高级别牌权玩家已经提前操作了，这条日志说明你的网络比别人慢了一丢丢")
				return true
			}
			return false
		}
		if cbOperateCard != sp.ProvideCard && wOperateCode != ZP_WIK_TIANLONG && wOperateCode != ZP_WIK_GANG && wOperateCode != ZP_WIK_HUA && wOperateCode != ZP_WIK_NULL && wOperateCode != ZP_WIK_TA {
			//杠和弃时允许不一致,因为玩家有多个牌可以杠时，sp.ProvideCard不知道是哪个，弃时cbOperateCard可能是0
			sp.OnWriteGameRecord(wChairID, "牌数据有误")
			return false
		}

		sp.addReplayOrder(wChairID, E_HandleCardRight, wOperateCode, []int{})

		//变量定义
		wTargetUser := wChairID
		wTargetAction := wOperateCode

		//设置变量
		sp.Response[wChairID] = true
		sp.PerformAction[wChairID] = wOperateCode
		sp.OperateCard[wChairID] = cbOperateCard
		if cbOperateCard == 0 {
			sp.OperateCard[wChairID] = sp.ProvideCard
		}
		sp.CMD_OperateCard[wChairID].Code = s_C_OperateData.Code
		sp.CMD_OperateCard[wChairID].Card = s_C_OperateData.Card
		sp.CMD_OperateCard[wChairID].WeaveCount = s_C_OperateData.WeaveCount
		copy(sp.CMD_OperateCard[wChairID].WeaveKind[:], s_C_OperateData.WeaveKind[:])
		copy(sp.CMD_OperateCard[wChairID].WeaveInfo[:], s_C_OperateData.WeaveInfo[:])

		//放弃操作
		if wTargetAction == ZP_WIK_NULL {
			//游戏记录
			sp.OnWriteGameRecord(wTargetUser, "点击弃！")
		}
		if (sp.UserAction[wChairID]&ZP_WIK_CHI_HU) != 0 && (wTargetAction&ZP_WIK_CHI_HU) == 0 {
			//有胡时选择了弃胡 要过张
			sp.CreateGuoZhang(ZP_WIK_CHI_HU, ZP_WIK_NULL, wChairID, cbOperateCard)
		}
		if (sp.UserAction[wChairID]&ZP_WIK_PENG) != 0 && wTargetAction == 0 {
			//有碰时选择了弃 要过张
			sp.CreateGuoZhang(ZP_WIK_PENG, ZP_WIK_NULL, wChairID, cbOperateCard)
		}

		//判断最大权限用户是否已经操作
		iPlayer := uint16(0) //判断是否有拦胡
		for iPer := 0; iPer < sp.GetPlayerCount(); iPer++ {
			iPlayer = uint16((int(sp.ProvideUser) + iPer + sp.GetPlayerCount()) % sp.GetPlayerCount())
			if (sp.UserAction[iPlayer] & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
				//这个人离出牌的人最近，有优先胡牌的权利（拦胡）
				if sp.Response[iPlayer] == true && (sp.PerformAction[iPlayer]&ZP_WIK_CHI_HU) == ZP_WIK_NULL {
					//但他已经选择弃了，找下一个
					continue
				}
				break //这个人离出牌的人最近，有优先胡牌的权利（拦胡）
			}
		}
		iPlayerPeng := uint16(0) //判断是否有拦碰
		for iPer := 0; iPer < sp.GetPlayerCount(); iPer++ {
			iPlayerPeng = uint16((int(sp.ProvideUser) + iPer + sp.GetPlayerCount()) % sp.GetPlayerCount())
			if (sp.UserAction[iPlayerPeng] & ZP_WIK_PENG) != ZP_WIK_NULL {
				//这个人离出牌的人最近，有优先碰牌的权利（拦碰）
				if sp.Response[iPlayerPeng] == true && (sp.PerformAction[iPlayerPeng]&ZP_WIK_PENG) == ZP_WIK_NULL {
					//但他已经选择弃了，找下一个
					continue
				}
				break //这个人离出牌的人最近，有优先碰牌的权利（拦碰）
			}
		}
		i := uint16(0)
		for i = 0; int(i) < sp.GetPlayerCount(); i++ {
			//获取动作
			cbUserAction := sp.PerformAction[i] //如果之前某个玩家已经点击了弃，对应cbUserAction会是0
			if sp.Response[i] == false {
				cbUserAction = sp.UserAction[i]
			}

			//优先级别
			cbUserActionRank := sp.GameLogic.GetUserActionRank(cbUserAction) // 动作等级
			wTargetActionRank := sp.GameLogic.GetUserActionRank(wTargetAction)
			if iPlayer == i && 8 == cbUserActionRank {
				//离出牌的人最近的人，如果他能胡，他有优先权
				if !sp.DuoHu {
					cbUserActionRank = 9 //不是一炮多响时，优先
				}
			}
			if iPlayer == wTargetUser && 8 == wTargetActionRank {
				//离出牌的人最近的人，如果他能胡，他有优先权
				if !sp.DuoHu {
					wTargetActionRank = 9 //不是一炮多响时，优先
				}
			}
			if iPlayerPeng == i && 3 == cbUserActionRank {
				//离出牌的人最近的人，如果他能碰，他有优先权
				cbUserActionRank = 4 //优先
			}
			if iPlayerPeng == wTargetUser && 3 == wTargetActionRank {
				//离出牌的人最近的人，如果他能碰，他有优先权
				wTargetActionRank = 4 //优先
			}

			//动作判断
			if cbUserActionRank > wTargetActionRank {
				wTargetUser = i
				wTargetAction = cbUserAction
			}
		}
		if sp.Response[wTargetUser] == false {
			//AfxMessageBox("最大操作用户没有响应,请等待!");
			szGameRecord := fmt.Sprintf("OnUserOperateCard wTargetUser=%d,wTargetAction=0x%x 最大操作用户没有响应,请等待!", wTargetUser, wTargetAction)
			sp.OnWriteGameRecord(wChairID, szGameRecord)
			return true

		}

		//放弃操作
		if wTargetAction == ZP_WIK_NULL {
			if sp.UserOperate_Null(wTargetUser, cbOperateCard, wTargetAction) {
				return true
			}
		}

		//变量定义
		cbTargetCard := sp.OperateCard[wTargetUser]

		if sp.GetPriUser(ZP_WIK_PENG) == wTargetUser && (wTargetAction == ZP_WIK_PENG) {
			for byUser := uint16(0); byUser < uint16(sp.GetPlayerCount()); byUser++ {
				if byUser != wTargetUser && (sp.UserAction[byUser]&ZP_WIK_PENG) != 0 && sp.Response[byUser] == true && sp.PerformAction[byUser] == ZP_WIK_NULL {
					sp.DismissGuoZhang(gKind_PengCard, byUser, cbTargetCard) //解除碰牌的过张
				}
			}
		}

		//出牌变量
		sp.OutCardData = 0
		sp.OutCardUser = static.INVALID_CHAIR

		//胡牌操作
		if wTargetAction == ZP_WIK_CHI_HU {
			if sp.UserOperate_Hu(wTargetUser, cbTargetCard, wTargetAction) {
				return true
			}
		}

		GuanHuaFlag := false //是否观生的牌滑牌的情况

		//删除扑克
		switch wTargetAction {
		case ZP_WIK_LEFT, ZP_WIK_RIGHT, ZP_WIK_CENTER, ZP_WIK_JIAN: //实际只有捡牌
			{
				sp.CMD_OperateCard[wTargetUser].WeaveCount = 1 //补充数据
				sp.CMD_OperateCard[wTargetUser].WeaveKind[0] = ZP_WIK_JIAN
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][0] = cbTargetCard

				//游戏记录
				szGameRecord := fmt.Sprintf("牌型：%s，本次捡牌：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wTargetUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbTargetCard)))
				sp.OnWriteGameRecord(wTargetUser, szGameRecord)

				//记录碰牌
				sp.addReplayOrder(wTargetUser, E_ZP_Wik_Jian, int(cbTargetCard), []int{})
				//捡的牌就放到自己手上
				//sp.CardIndex[wTargetUser][sp.GameLogic.SwitchToCardIndex(cbTargetCard)]++

				//记录捡牌
				sp.UserJianCards[wTargetUser] = append(sp.UserJianCards[wTargetUser], cbTargetCard)
				//记录本轮的捡牌
				sp.UserJianCardsCur[wTargetUser] = append(sp.UserJianCardsCur[wTargetUser], cbTargetCard)
				break
			}
		case ZP_WIK_PENG: //碰牌操作
			{
				//判断能不能换统
				_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wTargetUser])
				index := sp.GameLogic.SwitchToCardIndexNoHua(cbTargetCard)
				if index < TCGZ_MAX_INDEX && byTempCardIndex[index] >= 2 {
					_, realTongCntBef, tongInfoBef := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wTargetUser].TongCnt)
					byTempCardIndex[index] -= 2
					_, realTongCnt, tongInfo := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wTargetUser].TongCnt)
					if realTongCnt < sp.TongInfo[wTargetUser].TongCnt {
						if realTongCntBef > realTongCnt {
							//因为这张牌导致的统数减少
							sp.OnWriteGameRecord(wTargetUser, "碰牌数据有误,不能换统时，不能碰")
							sp.Response[wTargetUser] = false
							return false
						}
					}
					//v1.5 20200715 泛不换统后，需要确定换到哪里去了
					if realTongCntBef > realTongCnt {
						needCnt := 1
						if tongInfoBef[index] > tongInfo[index] {
							needCnt = tongInfoBef[index] - tongInfo[index]
						}
						//随便找一个替换
						for needIndex := 0; needIndex < needCnt; needIndex++ {
							for j := 0; j < TCGZ_MAX_INDEX; j++ {
								if tongInfo[j] > sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt {
									sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt++
									sp.TongInfo[wTargetUser].CardTongInfo[index].TongCnt--
									sp.OnWriteGameRecord(wTargetUser, fmt.Sprintf("碰牌换统到：%s", sp.GameLogic.SwitchToCardName2(byte(j))))
									break
								}
							}
						}
					}
				} else {
					sp.OnWriteGameRecord(wTargetUser, "碰牌数据有误")
					sp.Response[wTargetUser] = false
					return false
				}

				wIndex := sp.WeaveItemCount[wTargetUser]
				sp.WeaveItemCount[wTargetUser]++
				sp.WeaveItemArray[wTargetUser][wIndex].PublicCard = cbOperateCard //如果需要遮罩，碰的是这张牌
				sp.WeaveItemArray[wTargetUser][wIndex].WeaveKind = ZP_WIK_PENG
				sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = sp.ProvideUser
				if sp.ProvideUser == static.INVALID_CHAIR {
					sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = wTargetUser
				}
				retNum, cards := sp.GetSameCards(wTargetUser, cbTargetCard, 2)
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[0] = cbTargetCard
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[1] = cards[0]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[2] = cards[1]

				sp.CMD_OperateCard[wTargetUser].WeaveCount = 1 //补充数据
				sp.CMD_OperateCard[wTargetUser].WeaveKind[0] = ZP_WIK_PENG
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][0] = cbTargetCard
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][1] = cards[0]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][2] = cards[1]

				//删除扑克
				cbRemoveCard := []byte{cards[0], cards[1]}

				//游戏记录
				szGameRecord := fmt.Sprintf("牌型：%s，碰牌：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wTargetUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbTargetCard)))
				sp.OnWriteGameRecord(wTargetUser, szGameRecord)

				//记录回放
				values := []int{}
				values = append(values, int(cbTargetCard))
				values = append(values, int(cards[0]))
				values = append(values, int(cards[1]))
				//记录碰牌
				sp.addReplayOrder(wTargetUser, E_ZP_Peng, 0, values)
				_, sp.CardIndex[wTargetUser] = sp.GameLogic.RemoveCard2(sp.CardIndex[wTargetUser], cbRemoveCard, byte(retNum))
				sp.UserPengCount[wTargetUser]++
				break
			}
		case ZP_WIK_GANG: //杠牌操作
			{
				//判断能不能换统
				_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wTargetUser])
				index := sp.GameLogic.SwitchToCardIndexNoHua(cbTargetCard)
				if index < TCGZ_MAX_INDEX && byTempCardIndex[index] >= 3 {
					_, realTongCntBef, tongInfoBef := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wTargetUser].TongCnt)
					byTempCardIndex[index] -= 3
					_, realTongCnt, tongInfo := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wTargetUser].TongCnt)
					if realTongCnt < sp.TongInfo[wTargetUser].TongCnt {
						if realTongCntBef > realTongCnt {
							//因为这张牌导致的统数减少
							sp.OnWriteGameRecord(wTargetUser, "招牌数据有误,不能换统时，不能招")
							sp.Response[wTargetUser] = false
							return false
						}
					}
					//v1.5 20200715 泛不换统后，需要确定换到哪里去了
					if realTongCntBef > realTongCnt {
						needCnt := 1
						if tongInfoBef[index] > tongInfo[index] {
							needCnt = tongInfoBef[index] - tongInfo[index]
						}
						//随便找一个替换
						for needIndex := 0; needIndex < needCnt; needIndex++ {
							for j := 0; j < TCGZ_MAX_INDEX; j++ {
								if tongInfo[j] > sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt {
									sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt++
									sp.TongInfo[wTargetUser].CardTongInfo[index].TongCnt--
									sp.OnWriteGameRecord(wTargetUser, fmt.Sprintf("招牌换统到：%s", sp.GameLogic.SwitchToCardName2(byte(j))))
									break
								}
							}
						}
					}
				} else {
					sp.OnWriteGameRecord(wTargetUser, "招牌数据有误")
					sp.Response[wTargetUser] = false
					return false
				}

				wIndex := sp.WeaveItemCount[wTargetUser]
				sp.WeaveItemCount[wTargetUser]++
				sp.WeaveItemArray[wTargetUser][wIndex].PublicCard = cbOperateCard //如果需要遮罩，杠的是这张牌
				sp.WeaveItemArray[wTargetUser][wIndex].WeaveKind = ZP_WIK_GANG
				sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = sp.ProvideUser
				if sp.ProvideUser == static.INVALID_CHAIR {
					sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = wTargetUser
				}
				iNum := byte(3)
				if sp.SendStatus {
					iNum = 4
				}
				retNum, cards := sp.GetSameCards(wTargetUser, cbTargetCard, iNum)
				if retNum != iNum {
					sp.OnWriteGameRecord(wTargetUser, "招牌数据有误")
					sp.Response[wTargetUser] = false
					return false
				}
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[0] = cbTargetCard
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[1] = cards[0]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[2] = cards[1]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[3] = cards[2]

				sp.CMD_OperateCard[wTargetUser].WeaveCount = 1 //补充数据
				sp.CMD_OperateCard[wTargetUser].WeaveKind[0] = ZP_WIK_GANG
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][0] = cbTargetCard
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][1] = cards[0]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][2] = cards[1]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][3] = cards[2]

				cbRemoveCard := []byte{cards[0], cards[1], cards[2]}
				if sp.SendStatus {
					//说明是暗杠
					cbRemoveCard = append(cbRemoveCard, cbTargetCard)
				}
				_, sp.CardIndex[wTargetUser] = sp.GameLogic.RemoveCard2(sp.CardIndex[wTargetUser], cbRemoveCard, byte(retNum))

				//游戏记录
				szGameRecord := fmt.Sprintf("牌型：%s，杠牌(招)：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wTargetUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbTargetCard)))
				sp.OnWriteGameRecord(wTargetUser, szGameRecord)

				//记录回放
				values := []int{}
				values = append(values, int(cbTargetCard))
				values = append(values, int(cards[0]))
				values = append(values, int(cards[1]))
				values = append(values, int(cards[2]))
				//记录杠牌
				sp.addReplayOrder(wTargetUser, E_ZP_Gang, 0, values)
				break
			}
		case ZP_WIK_TIANLONG: //统牌操作
			{
				iNum := byte(4)
				retNum, cards := sp.GetSameCards(wTargetUser, cbTargetCard, iNum)
				if retNum+sp.CardIndex[wTargetUser][27] < 4 {
					sp.OnWriteGameRecord(wTargetUser, "统牌数据有误")
					sp.Response[wTargetUser] = false
					return false
				}
				sp.CMD_OperateCard[wTargetUser].WeaveCount = 1 //补充数据
				sp.CMD_OperateCard[wTargetUser].WeaveKind[0] = ZP_WIK_TIANLONG
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][0] = cards[3]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][1] = cards[0]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][2] = cards[1]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][3] = cards[2]

				//更新实际的统信息
				_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wTargetUser])
				_, realTongCnt, tongInfo := sp.GameLogic.CheckTong_Op(byTempCardIndex, sp.TongInfo[wTargetUser].TongCnt)
				if realTongCnt > sp.TongInfo[wTargetUser].TongCnt {
					sp.TongInfo[wTargetUser].TongCnt++
					sp.DispTongCnt[wTargetUser]++
					if sp.GameLogic.IsValidCard(cbTargetCard) {
						curIndex := sp.GameLogic.SwitchToCardIndexNoHua(cbTargetCard)
						//避免客户端没有记录换统时出现某一个牌的统数超出了
						if tongInfo[curIndex] <= sp.TongInfo[wTargetUser].CardTongInfo[curIndex].TongCnt {
							//随便找一个替换
							for j := 0; j < TCGZ_MAX_INDEX; j++ {
								if tongInfo[j] > sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt {
									sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt++
									sp.OnWriteGameRecord(wTargetUser, fmt.Sprintf("统牌换统到：%s", sp.GameLogic.SwitchToCardName2(byte(j))))
									break
								}
							}
						} else {
							sp.TongInfo[wTargetUser].CardTongInfo[curIndex].TongCnt++ //记录那张牌统过
						}
					}
				} else {
					sp.OnWriteGameRecord(wTargetUser, "统牌数据有误")
					sp.Response[wTargetUser] = false
					return false
				}

				//cbRemoveCard:=[]byte{cards[0],cards[1],cards[2],cards[3]};
				//_,sp.CardIndex[wTargetUser]=sp.GameLogic.RemoveCard2(sp.CardIndex[wTargetUser],cbRemoveCard,byte(retNum))

				//游戏记录
				szGameRecord := fmt.Sprintf("牌型：%s，统牌：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wTargetUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbTargetCard)))
				sp.OnWriteGameRecord(wTargetUser, szGameRecord)

				//记录回放
				values := []int{}
				values = append(values, int(cards[0]))
				//记录杠牌
				sp.addReplayOrder(wTargetUser, E_ZP_TianLong, 0, values)
				break
			}
		case ZP_WIK_HUA: //滑牌操作
			{
				iNum := byte(4)
				if sp.SendStatus {
					iNum = 5
				}
				retNum, cards := sp.GetSameCards(wTargetUser, cbTargetCard, iNum)
				if retNum >= 4 {
					if retNum == 4 {
						cards[4] = cbTargetCard
					}
					cbRemoveCard := []byte{cards[0], cards[1], cards[2], cards[3]}
					if sp.SendStatus {
						//说明是暗杠
						cbRemoveCard = append(cbRemoveCard, cards[4])
					}
					_, sp.CardIndex[wTargetUser] = sp.GameLogic.RemoveCard2(sp.CardIndex[wTargetUser], cbRemoveCard, byte(retNum))
				}
				//begin：v1.5.1 20200817 泛时换统规则修改：如果泛没有统过的牌要换成泛统过的牌
				////更新实际的统信息
				cardIdx := sp.GameLogic.SwitchToCardIndexNoHua(cbTargetCard)
				if sp.GameLogic.IsValidCard(cbTargetCard) && cardIdx < TCGZ_MAX_INDEX {
					if sp.TongInfo[wTargetUser].CardTongInfo[cardIdx].TongCnt == 0 {
						//找第一个替换，这里得和客户端一致
						for j := 0; j < TCGZ_MAX_INDEX; j++ {
							if sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt > 0 {
								sp.TongInfo[wTargetUser].CardTongInfo[cardIdx].TongCnt++
								sp.TongInfo[wTargetUser].CardTongInfo[j].TongCnt--
								sp.OnWriteGameRecord(wTargetUser, fmt.Sprintf("泛牌换统：由原来的【%s】换统到：%s", sp.GameLogic.SwitchToCardName2(byte(j)), sp.GameLogic.SwitchToCardName2(byte(cardIdx))))
								break
							}
						}
					}
					if sp.TongInfo[wTargetUser].CardTongInfo[cardIdx].TongCnt > 0 {
						GuanHuaFlag = true
						needCnt := sp.TongInfo[wTargetUser].CardTongInfo[cardIdx].TongCnt
						if sp.TongInfo[wTargetUser].CardTongInfo[cardIdx].TongCnt >= 2 {
							//needCnt = 2
							needCnt = 1
						}
						sp.TongInfo[wTargetUser].TongCnt -= needCnt
						sp.TongInfo[wTargetUser].CardTongInfo[cardIdx].TongCnt -= needCnt //记录那张牌统过
						sp.DispTongCnt[wTargetUser] -= needCnt                            //统次数要减少
						sp.OnWriteGameRecord(wTargetUser, fmt.Sprintf("统牌次数减少%d个", needCnt))
					} else {
						sp.OnWriteGameRecord(wTargetUser, "泛牌数据有误，只能泛统过的牌，或者换统")
						sp.Response[wTargetUser] = false
						return false
					}
				}
				//end：v1.5.1 20200817 泛时换统规则修改：如果泛没有统过的牌要换成泛统过的牌

				wIndex := sp.WeaveItemCount[wTargetUser]
				sp.WeaveItemCount[wTargetUser]++
				sp.WeaveItemArray[wTargetUser][wIndex].PublicCard = cbOperateCard //如果需要遮罩，杠的是这张牌
				sp.WeaveItemArray[wTargetUser][wIndex].WeaveKind = ZP_WIK_HUA
				sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = sp.ProvideUser
				if sp.ProvideUser == static.INVALID_CHAIR {
					sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = wTargetUser
				}
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[0] = cards[4]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[1] = cards[0]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[2] = cards[1]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[3] = cards[2]
				sp.WeaveItemArray[wTargetUser][wIndex].Cards[4] = cards[3]

				sp.CMD_OperateCard[wTargetUser].WeaveCount = 1 //补充数据
				sp.CMD_OperateCard[wTargetUser].WeaveKind[0] = ZP_WIK_HUA
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][0] = cards[4]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][1] = cards[0]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][2] = cards[1]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][3] = cards[2]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][4] = cards[3]
				//游戏记录
				szGameRecord := fmt.Sprintf("牌型：%s，泛牌：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wTargetUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbTargetCard)))
				sp.OnWriteGameRecord(wTargetUser, szGameRecord)
				//记录回放
				values := []int{}
				values = append(values, int(cards[4]))
				values = append(values, int(cards[0]))
				values = append(values, int(cards[1]))
				values = append(values, int(cards[2]))
				values = append(values, int(cards[3]))
				//记录杠牌
				sp.addReplayOrder(wTargetUser, E_ZP_Wik_Hua, 0, values)
				break
			}
		case ZP_WIK_TA: //踏牌操作
			{
				iNum := byte(4)
				if sp.SendStatus {
					iNum = 5
				}
				retNum, cards := sp.GetSameCards(wTargetUser, cbTargetCard, iNum)
				if retNum >= 4 {
					if retNum == 4 {
						cards[4] = cbTargetCard
					}
					cbRemoveCard := []byte{cards[0], cards[1], cards[2], cards[3]}
					if sp.SendStatus {
						//说明是暗杠
						cbRemoveCard = append(cbRemoveCard, cards[4])
					}
					_, sp.CardIndex[wTargetUser] = sp.GameLogic.RemoveCard2(sp.CardIndex[wTargetUser], cbRemoveCard, byte(retNum))
				} else {
					//在招里面找
					if sp.WeaveItemCount[wTargetUser] > 0 && retNum <= 1 {
						for k := byte(0); k < sp.WeaveItemCount[wTargetUser]; k++ {
							if sp.WeaveItemArray[wTargetUser][k].WeaveKind != ZP_WIK_GANG {
								continue
							}
							if sp.GameLogic.IsValidCard(sp.WeaveItemArray[wTargetUser][k].Cards[0]) &&
								sp.GameLogic.SwitchToCardIndexNoHua(sp.WeaveItemArray[wTargetUser][k].Cards[0]) == sp.GameLogic.SwitchToCardIndexNoHua(cbTargetCard) {
								for m := byte(0); m < 4; m++ {
									cards[retNum+m] = sp.WeaveItemArray[wTargetUser][k].Cards[m]
								}
								if retNum == 1 {
									cbRemoveCard := []byte{cards[0]}
									_, sp.CardIndex[wTargetUser] = sp.GameLogic.RemoveCard2(sp.CardIndex[wTargetUser], cbRemoveCard, byte(retNum))
								} else {
									cards[4] = cbTargetCard
								}
								GuanHuaFlag = true
								//修改组合牌区的信息为踏
								sp.WeaveItemArray[wTargetUser][k].WeaveKind = ZP_WIK_TA
								if sp.GameLogic.IsValidCard(cards[0]) {
									sp.WeaveItemArray[wTargetUser][k].Cards[4] = cards[0]
								} else {
									sp.WeaveItemArray[wTargetUser][k].Cards[4] = cbTargetCard
								}
								break
							}
						}
						if !GuanHuaFlag {
							sp.OnWriteGameRecord(wTargetUser, "踏牌数据有误")
							sp.Response[wTargetUser] = false
							return false
						}
					} else {
						sp.OnWriteGameRecord(wTargetUser, "踏牌数据有误1")
						sp.Response[wTargetUser] = false
						return false
					}
				}
				if !GuanHuaFlag {
					//踏时不新增，类似续杠
					wIndex := sp.WeaveItemCount[wTargetUser]
					sp.WeaveItemCount[wTargetUser]++
					sp.WeaveItemArray[wTargetUser][wIndex].PublicCard = cbOperateCard //如果需要遮罩，杠的是这张牌
					sp.WeaveItemArray[wTargetUser][wIndex].WeaveKind = ZP_WIK_TA
					sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = sp.ProvideUser
					if sp.ProvideUser == static.INVALID_CHAIR {
						sp.WeaveItemArray[wTargetUser][wIndex].ProvideUser = wTargetUser
					}
					sp.WeaveItemArray[wTargetUser][wIndex].Cards[0] = cards[4]
					sp.WeaveItemArray[wTargetUser][wIndex].Cards[1] = cards[0]
					sp.WeaveItemArray[wTargetUser][wIndex].Cards[2] = cards[1]
					sp.WeaveItemArray[wTargetUser][wIndex].Cards[3] = cards[2]
					sp.WeaveItemArray[wTargetUser][wIndex].Cards[4] = cards[3]
				}
				sp.CMD_OperateCard[wTargetUser].WeaveCount = 1 //补充数据
				sp.CMD_OperateCard[wTargetUser].WeaveKind[0] = ZP_WIK_TA
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][0] = cards[4]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][1] = cards[0]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][2] = cards[1]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][3] = cards[2]
				sp.CMD_OperateCard[wTargetUser].WeaveInfo[0][4] = cards[3]
				//游戏记录
				szGameRecord := fmt.Sprintf("牌型：%s，踏牌：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wTargetUser][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbTargetCard)))
				sp.OnWriteGameRecord(wTargetUser, szGameRecord)
				//记录回放
				values := []int{}
				values = append(values, int(cards[4]))
				values = append(values, int(cards[0]))
				values = append(values, int(cards[1]))
				values = append(values, int(cards[2]))
				values = append(values, int(cards[3]))
				//记录杠牌
				sp.addReplayOrder(wTargetUser, E_ZP_Wik_Ta, 0, values)
				break
			}
		}
		sp.DismissGuoZhang(gKind_HuCard, wTargetUser, 0)   //操作属于变动手牌了，解除胡牌的过张
		sp.DismissGuoZhang(gKind_PengCard, wTargetUser, 0) //操作属于变动手牌了，解除碰牌的过张

		//用户状态
		sp.Response = [TCGZ_MAX_PLAYER]bool{}
		sp.UserAction = [TCGZ_MAX_PLAYER]int{}
		sp.OperateCard = [TCGZ_MAX_PLAYER]byte{}
		sp.PerformAction = [TCGZ_MAX_PLAYER]int{}

		//已经过了一轮了，本轮捡牌标记删除
		if wTargetAction != ZP_WIK_JIAN {
			sp.UserJianCardsCur[wTargetUser] = []byte{}
		}

		//构造结果
		var OperateResult static.Msg_S_ZP_OperateResult
		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = cbTargetCard
		OperateResult.OperateCode = wTargetAction
		OperateResult.HasOutPower = 1 //默认有出牌权
		OperateResult.ProvideUser = sp.ProvideUser
		OperateResult.Type = 0
		if sp.SendStatus && (wTargetAction == ZP_WIK_GANG || wTargetAction == ZP_WIK_HUA) {
			OperateResult.Type = 1 //暗杠
		}
		if sp.ProvideUser == static.INVALID_CHAIR {
			sp.ProvideUser = wTargetUser
			OperateResult.ProvideUser = wTargetUser
		}
		// 操作玩家的超时时间
		iOvertime := sp.OperateTime
		if sp.TuoGuanPlayer[wTargetUser] == true && iOvertime > sp.AutoOutTime {
			iOvertime = sp.AutoOutTime
		}
		OperateResult.LeftCardsCount = sp.GetCardsNum(sp.CardIndex[wTargetUser])
		OperateResult.LeftTime = int(iOvertime)
		OperateResult.WeaveCount = sp.CMD_OperateCard[wTargetUser].WeaveCount
		copy(OperateResult.WeaveInfo[:], sp.CMD_OperateCard[wTargetUser].WeaveInfo[:])
		copy(OperateResult.WeaveKind[:], sp.CMD_OperateCard[wTargetUser].WeaveKind[:])

		//发送消息
		sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

		//记录提供者
		sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_Provider, int(sp.ProvideUser))
		sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_ProvideCard, int(cbTargetCard))

		//滑牌不是观生的牌，需要多补一张牌，在发送消息后补
		if !GuanHuaFlag && wTargetAction == ZP_WIK_HUA {
			sp.BuCardData(wTargetUser)
		}

		sp.SendStatus = false //吃碰杠后 把发牌标记置为false
		sp.CMD_OperateCard = [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard{}

		//重新计算总胡息
		//byHuxi := sp.WeaveHuxi[wTargetUser];
		sp.WeaveHuxi[wTargetUser] = sp.GameLogic.CalculateWeaveHuXi(sp.WeaveItemArray[wTargetUser], sp.WeaveItemCount[wTargetUser])
		//if (byHuxi != sp.WeaveHuxi[wTargetUser]) {
		//胡息有变化
		sp.SendHuxi(wTargetUser)
		sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_HuXi, sp.WeaveHuxi[wTargetUser])
		sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_Tong, sp.DispTongCnt[wTargetUser])
		//}

		//是否能继续操作，或者补牌
		if sp.UserOperate_aft(wTargetUser, cbTargetCard, wTargetAction) {
			return true
		}
		tingcnt, tingresult := sp.CheckTing(wTargetUser)
		if tingcnt > 0 {
			sp.SendTingInfo(wTargetUser, tingresult)
		}
		return true
	}
	sp.OnWriteGameRecord(wChairID, "重复操作了牌权2，此次不需响应")
	return true
}

// 在观生阶段的结果处理，
func (sp *Sport_zp_jzhp) UserOperate_GuanSheng_bef(wChairID uint16, operateCard byte, operateCode int) bool {
	if sp.GameState != ZP_AS_GUANSHENG {
		return true
	}
	sp.IsTongAction = false
	//操作结束，这些变量都必须清理
	sp.Response = [TCGZ_MAX_PLAYER]bool{}
	sp.UserAction = [TCGZ_MAX_PLAYER]int{}
	sp.OperateCard = [TCGZ_MAX_PLAYER]byte{}
	sp.PerformAction = [TCGZ_MAX_PLAYER]int{}
	sp.CMD_OperateCard = [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard{}

	tongflag := false
	GuanHuaFlag := false
	if int(wChairID) < sp.GetPlayerCount() {
		if operateCode == ZP_WIK_TIANLONG || operateCode == ZP_WIK_HUA {
			if operateCode == ZP_WIK_TIANLONG {
				szGameRecord := fmt.Sprintf("开局统时选择统，统牌：%s", sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(operateCard)))
				sp.OnWriteGameRecord(wChairID, szGameRecord)
				index := sp.GameLogic.SwitchToCardIndexNoHua(operateCard)
				_, CardIndexTemp := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[wChairID])
				retb, _, tongInfo := sp.GameLogic.CheckTong_Op(CardIndexTemp, sp.TongInfo[wChairID].TongCnt)
				if retb {
					if retcnt, _ := sp.GameLogic.IsTongCardIndex(CardIndexTemp, index); retcnt > 0 {
						tongflag = true
					} else {
						sp.OnWriteGameRecord(wChairID, "统的牌没有4张或以上---2！！！")
						//继续观生
						sp.CurrentUser = wChairID
						sp.setLimitedTime(1)
						return true
					}
				}
				//记录回放
				values := []int{}
				//保存观生牌
				num, cards := sp.GetSameCards(wChairID, operateCard, 4)

				if !tongflag {
					sp.OnWriteGameRecord(wChairID, "统的牌没有4张或以上！！！")
					//继续观生
					sp.CurrentUser = wChairID
					sp.setLimitedTime(1)
					return true
				}
				//更新玩家统的次数
				sp.TongInfo[wChairID].TongCnt++
				sp.DispTongCnt[wChairID]++
				//sp.TongInfo[wChairID].CardTongInfo[index].TongCnt++ //记录那张牌统过
				//避免客户端没有记录换统时出现某一个牌的统数超出了。同时可以让数目多的牌优先统牌，避免后期在换统出问题
				if tongInfo[index] <= sp.TongInfo[wChairID].CardTongInfo[index].TongCnt {
					//随便找一个替换
					for j := 0; j < TCGZ_MAX_INDEX-1; j++ {
						if tongInfo[j] > sp.TongInfo[wChairID].CardTongInfo[j].TongCnt {
							sp.TongInfo[wChairID].CardTongInfo[j].TongCnt++
							sp.OnWriteGameRecord(wChairID, fmt.Sprintf("统牌换统到：%s", sp.GameLogic.SwitchToCardName2(byte(j))))
							break
						}
					}
				} else {
					sp.TongInfo[wChairID].CardTongInfo[index].TongCnt++ //记录那张牌统过
				}
				for i := byte(0); i < 5 && i < num; i++ {
					values = append(values, int(cards[i]))
					sp.CMD_OperateCard[wChairID].WeaveInfo[0][i] = cards[i]
				}
				sp.CMD_OperateCard[wChairID].WeaveKind[0] = ZP_WIK_TIANLONG
				sp.CMD_OperateCard[wChairID].WeaveCount = 1

				sp.addReplayOrder(wChairID, E_HandleCardRight, ZP_WIK_TIANLONG, []int{})

				//记录胡牌
				sp.addReplayOrder(wChairID, E_ZP_TianLong, 0, values)
			} else if operateCode == ZP_WIK_HUA {
				szGameRecord := fmt.Sprintf("观生时选择滑，滑牌：%s", sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(operateCard)))
				sp.OnWriteGameRecord(wChairID, szGameRecord)
				//记录回放
				values := []int{}
				//保存观生牌
				if sp.UserGuanCardsCount[wChairID] < 10 {
					num, cards := sp.GetSameCards(wChairID, operateCard, 5)
					if num == 0 {
						sp.OnWriteGameRecord(wChairID, "滑的牌没有5张或以上！！！")
						//继续观生
						sp.CurrentUser = wChairID
						sp.setLimitedTime(1)
						return true
					}
					//在观里面找
					if sp.UserGuanCardsCount[wChairID] > 0 && num == 1 {
						for k := byte(0); k < sp.UserGuanCardsCount[wChairID]; k++ {
							if sp.GameLogic.IsValidCard(sp.UserGuanCards[wChairID][k][0]) &&
								sp.GameLogic.SwitchToCardIndexNoHua(sp.UserGuanCards[wChairID][k][0]) == sp.GameLogic.SwitchToCardIndexNoHua(operateCard) {
								for m := byte(0); m < 4; m++ {
									cards[num+m] = sp.UserGuanCards[wChairID][k][m]
								}
								sp.DeleteGuan(wChairID, k)
								GuanHuaFlag = true
								break
							}
						}
						if !GuanHuaFlag {
							sp.OnWriteGameRecord(wChairID, "滑牌数据有误")
							//继续观生
							sp.CurrentUser = wChairID
							sp.setLimitedTime(1)
							return true
						}
					} else if num != 5 {
						sp.OnWriteGameRecord(wChairID, "滑牌数据有误1")
						//继续观生
						sp.CurrentUser = wChairID
						sp.setLimitedTime(1)
						return true
					}
					for i := byte(0); i < 5; i++ {
						if sp.GameLogic.IsValidCard(cards[i]) && sp.CardIndex[wChairID][sp.GameLogic.SwitchToCardIndex(cards[i])] > 0 {
							sp.CardIndex[wChairID][sp.GameLogic.SwitchToCardIndex(cards[i])]--
						}
						values = append(values, int(cards[i]))
						sp.CMD_OperateCard[wChairID].WeaveInfo[0][i] = cards[i]
					}

					sp.CMD_OperateCard[wChairID].WeaveKind[0] = ZP_WIK_HUA
					sp.CMD_OperateCard[wChairID].WeaveCount = 1

					wIndex := sp.WeaveItemCount[wChairID]
					sp.WeaveItemCount[wChairID]++
					sp.WeaveItemArray[wChairID][wIndex].PublicCard = operateCard //如果需要遮罩，杠的是这张牌
					sp.WeaveItemArray[wChairID][wIndex].WeaveKind = ZP_WIK_HUA
					sp.WeaveItemArray[wChairID][wIndex].ProvideUser = wChairID
					sp.WeaveItemArray[wChairID][wIndex].Cards[0] = cards[0]
					sp.WeaveItemArray[wChairID][wIndex].Cards[1] = cards[1]
					sp.WeaveItemArray[wChairID][wIndex].Cards[2] = cards[2]
					sp.WeaveItemArray[wChairID][wIndex].Cards[3] = cards[3]
					sp.WeaveItemArray[wChairID][wIndex].Cards[4] = cards[4]
				}
				sp.addReplayOrder(wChairID, E_HandleCardRight, ZP_WIK_HUA, []int{})

				//记录胡牌
				sp.addReplayOrder(wChairID, E_ZP_Wik_Hua, 0, values)
			}

			//重新计算总胡息
			//byHuxi := sp.WeaveHuxi[wChairID];
			sp.WeaveHuxi[wChairID] = sp.GameLogic.CalculateWeaveHuXi(sp.WeaveItemArray[wChairID], sp.WeaveItemCount[wChairID])
			//if (byHuxi != sp.WeaveHuxi[wChairID]) {
			//胡息有变化
			sp.SendHuxi(wChairID)
			sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_HuXi, sp.WeaveHuxi[wChairID])
			sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_Tong, sp.DispTongCnt[wChairID])
			//}

			//构造结果
			var OperateResult static.Msg_S_ZP_OperateResult
			OperateResult.OperateUser = wChairID
			OperateResult.OperateCard = operateCard
			OperateResult.OperateCode = operateCode
			OperateResult.ProvideUser = wChairID
			OperateResult.LeftCardsCount = sp.GetCardsNum(sp.CardIndex[wChairID])
			OperateResult.WeaveCount = sp.CMD_OperateCard[wChairID].WeaveCount
			copy(OperateResult.WeaveInfo[:], sp.CMD_OperateCard[wChairID].WeaveInfo[:])
			copy(OperateResult.WeaveKind[:], sp.CMD_OperateCard[wChairID].WeaveKind[:])

			sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

			sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_ProvideCard, int(operateCard))

			sp.PowerStartTime = time.Now().Unix() //权限开始时间
			sp.CurrentUser = wChairID

			sp.BuCardData(sp.CurrentUser)
			//滑牌不是观生的牌，需要多补一张牌，在发送消息后补
			if !GuanHuaFlag && operateCode == ZP_WIK_HUA {
				sp.BuCardData(sp.CurrentUser)
			}

			//继续观生
			sp.setLimitedTime(1)
			return true
		}
	}
	//放弃操作
	if operateCode == ZP_WIK_NULL {
		//游戏记录
		sp.OnWriteGameRecord(wChairID, "开局统时点击弃！")

		//构造结果
		var OperateResult static.Msg_S_ZP_OperateResult
		OperateResult.OperateUser = wChairID
		OperateResult.OperateCard = operateCard
		OperateResult.OperateCode = ZP_WIK_NULL
		OperateResult.ProvideUser = wChairID
		OperateResult.LeftCardsCount = sp.GetCardsNum(sp.CardIndex[wChairID])
		sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

		sp.CurrentUser = uint16(sp.GetFrontSeat(wChairID)) //上家

		if sp.CurrentUser == sp.GuanStartUser {
			sp.CurrentUser = sp.Banker
			//结束观生,开始出牌
			sp.GameState = ZP_AS_STARTPLAY
			sp.setLimitedTime(1)
		} else {
			sp.setLimitedTime(1)
		}

		return true
	}
	return false
}

// 所有用户都选择了弃，
func (sp *Sport_zp_jzhp) UserOperate_Null(wTargetUser uint16, operateCard byte, operateCode int) bool {
	//放弃操作
	if operateCode == ZP_WIK_NULL {

		//用户状态
		sp.UserAction = [TCGZ_MAX_PLAYER]int{}
		sp.Response = [TCGZ_MAX_PLAYER]bool{}
		sp.OperateCard = [TCGZ_MAX_PLAYER]byte{}
		sp.PerformAction = [TCGZ_MAX_PLAYER]int{}
		sp.CMD_OperateCard = [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard{}

		//构造结果
		var OperateResult static.Msg_S_ZP_OperateResult
		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = operateCard
		OperateResult.OperateCode = ZP_WIK_NULL
		OperateResult.ProvideUser = sp.ProvideUser
		if sp.ProvideUser == static.INVALID_CHAIR {
			OperateResult.ProvideUser = wTargetUser
		}
		OperateResult.LeftCardsCount = sp.GetCardsNum(sp.CardIndex[wTargetUser])
		sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

		sp.CurrentUser = sp.ResumeUser

		if !sp.SendStatus {
			sp.StartHD()
			//出牌后，别人都放弃了，需要1s后在发下一张牌
			sp.GameState = ZP_AS_SENDCARD
			sp.setLimitedTime(1)
		} else {
			if sp.GameType != meta2.GT_ROAR {
				// 超时超时时间
				iOvertime := sp.OperateTime
				if sp.TuoGuanPlayer[wTargetUser] == true && iOvertime > sp.AutoOutTime {
					iOvertime = sp.AutoOutTime
				}
				//发牌后，别人都放弃了，需要出牌
				var s_Power static.Msg_S_ZP_SendPower
				s_Power.CurrentUser = sp.CurrentUser
				s_Power.LeftTime = int(iOvertime)
				sp.SendTableMsg(consts.MsgTypeSendPower, s_Power)
				sp.setLimitedTime(int64(iOvertime + 1))

				sp.OnWriteGameRecord(sp.CurrentUser, "拥有出牌牌权")
			} else {
				//直接给下家发牌
				sp.CurrentUser = uint16(sp.GetNextSeat(wTargetUser))
				sp.StartHD()
				sp.GameState = ZP_AS_SENDCARD
				sp.setLimitedTime(1) //给下家发牌，需要1s后在发下一张牌
			}
		}

		return true
	}
	return false
}

// 有用户都选择了胡，
func (sp *Sport_zp_jzhp) UserOperate_Hu(wTargetUser uint16, operateCard byte, operateCode int) bool {
	//胡牌操作
	if operateCode == ZP_WIK_CHI_HU {
		//结束信息
		sp.ChiHuCard = operateCard
		sp.ProvideUser = sp.ProvideUser

		//普通胡牌，有人点炮
		if sp.ChiHuCard != 0 {
			//插入扑克
			if sp.ChiHuResult[wTargetUser].ChiHuKind != ZP_WIK_NULL {
				//庄家天胡时不能插入扑克
				if sp.OutCardCount == 0 && wTargetUser == sp.Banker || sp.SendStatus {
				} else {
					sp.CardIndex[wTargetUser][sp.GameLogic.SwitchToCardIndex(sp.ChiHuCard)]++
				}
			}
		}
		for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
			WinerItem := sp.GetUserItemByChair(uint16(i))
			if WinerItem == nil {
				continue
			}
			if !sp.DuoHu && i != wTargetUser {
				continue
			}
			if (sp.UserAction[i] & ZP_WIK_CHI_HU) == ZP_WIK_NULL {
				continue
			}
			//构造结果
			var OperateResult static.Msg_S_ZP_OperateResult
			OperateResult.OperateUser = i
			OperateResult.OperateCard = operateCard
			OperateResult.OperateCode = ZP_WIK_CHI_HU
			OperateResult.ProvideUser = sp.ProvideUser
			if sp.ProvideUser == static.INVALID_CHAIR {
				OperateResult.ProvideUser = i
			}
			OperateResult.LeftCardsCount = sp.GetCardsNum(sp.CardIndex[i])
			sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

			//游戏记录
			szGameRecord := "牌型："
			if sp.GameLogic.IsValidCard(operateCard) {
				szGameRecord += fmt.Sprintf("%s，胡牌：%s", sp.GameLogic.SwitchToCardName1(sp.CardIndex[i][:]), sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(operateCard)))
			}
			sp.OnWriteGameRecord(i, szGameRecord)

			//记录胡牌
			sp.addReplayOrder(i, E_ZP_Hu, int(operateCard), []int{})
			//胡息计算
			iMaxIndex, iMaxHuXi, realHuXiInfo := sp.GameLogic.CalculateHuXi(&sp.ChiHuResult[i])
			sp.GameLogic.CalculateHuXi2(&sp.ChiHuResult[i], &iMaxIndex, &iMaxHuXi, &realHuXiInfo)

			//记录总胡息，并发送给客户端
			sp.WeaveHuxi[i] = iMaxHuXi
			sp.SendHuxi(i)
			//总胡息记录到回放
			if len(sp.ReplayRecord.R_Orders) > 0 && sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].R_Opt == E_ZP_Hu {
				//总胡息记录到回放
				sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_HuXi, sp.WeaveHuxi[i])
			}
		}
		//结束游戏
		sp.OnEventGameEnd(wTargetUser, static.GER_NORMAL)

		return true
	}
	return false
}

// 用户操作后出是否还能继续操作或者补牌，
func (sp *Sport_zp_jzhp) UserOperate_aft(wTargetUser uint16, operateCard byte, wTargetAction int) bool {

	if int(wTargetUser) > sp.GetPlayerCount() {
		return false
	}
	//设置用户
	sp.CurrentUser = wTargetUser
	sp.ProvideCard = 0
	sp.ProvideUser = static.INVALID_CHAIR

	//最大操作用户操作的是开朝，记录第几次开朝，开朝后还要检查他是否还可以胡
	if wTargetAction == ZP_WIK_GANG || wTargetAction == ZP_WIK_TIANLONG || wTargetAction == ZP_WIK_HUA || wTargetAction == ZP_WIK_TA {
		//补牌，需要1s后在发下一张牌.滑牌时如果要补两张牌，第一张在前面补
		sp.StartHD()
		sp.GameState = ZP_AS_SENDCARD
		sp.setLimitedTime(1)
		return true
	}
	//其他操作时需要判断目标用户是否还能杠
	//if sp.LeftCardCount > 0{
	if wTargetAction != ZP_WIK_PENG {
		//碰了后不能统
		//杠牌判断
		_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[sp.CurrentUser])

		userAction := 0
		userAction, _ = sp.GameLogic.AnalyseGangCard(byTempCardIndex, sp.TongInfo[sp.CurrentUser].TongCnt)
		sp.UserAction[wTargetUser] |= userAction

		//滑牌判断
		userAction = 0
		userAction, _ = sp.GameLogic.AnalyseTaCard(byTempCardIndex, sp.WeaveItemArray[sp.CurrentUser], sp.WeaveItemCount[sp.CurrentUser])
		sp.UserAction[wTargetUser] |= userAction
	}

	sp.GameState = ZP_AS_PLAYCARD // 更新为要打牌状态
	// 超时超时时间
	iOvertime := sp.OperateTime
	if sp.TuoGuanPlayer[wTargetUser] == true && iOvertime > sp.AutoOutTime {
		iOvertime = sp.AutoOutTime
	}
	sp.setLimitedTime(iOvertime + 1)

	if sp.UserAction[wTargetUser] != ZP_WIK_NULL {
		sp.ResumeUser = sp.CurrentUser
		sp.CurrentUser = static.INVALID_CHAIR
		sp.SendStatus = true //吃碰杠后 还能杠牌，把发牌标记置为true
		//发送提示
		sp.SendOperateNotify(int(iOvertime))
	} else {
		if sp.GameType != meta2.GT_ROAR {
			var s_Power static.Msg_S_ZP_SendPower
			s_Power.CurrentUser = wTargetUser
			s_Power.LeftTime = int(iOvertime)
			sp.SendTableMsg(consts.MsgTypeSendPower, s_Power)
			sp.CurrentUser = wTargetUser

			sp.OnWriteGameRecord(wTargetUser, "拥有出牌牌权")
		} else {
			//直接给下家发牌
			sp.CurrentUser = uint16(sp.GetNextSeat(wTargetUser))
			sp.StartHD()
			sp.GameState = ZP_AS_SENDCARD
			sp.setLimitedTime(1) //给下家发牌，需要1s后在发下一张牌
		}
	}
	return true
	//}
	return false
}

// 发送每个人的胡息
func (sp *Sport_zp_jzhp) SendHuxi(wTheSeat uint16) {
	var sHuXi static.Msg_S_ZP_HuXi
	sHuXi.TheSeat = wTheSeat
	for bySeat := 0; bySeat < sp.GetPlayerCount(); bySeat++ {
		sHuXi.HuXi[bySeat] = sp.WeaveHuxi[bySeat]
		sHuXi.Tong[bySeat] = sp.DispTongCnt[bySeat]
	}
	sp.SendTableMsg(consts.MsgTypeHUXI, sHuXi)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeHUXI, sHuXi)
}

// 获取同种牌，花牌优先使用
func (sp *Sport_zp_jzhp) GetSameCards(wUser uint16, cbCard byte, iNum byte) (byte, [5]byte) {
	retcards := [5]byte{}
	retnum := byte(0)
	if int(wUser) >= sp.GetPlayerCount() {
		return 0, retcards
	}
	if !sp.GameLogic.IsValidCard(cbCard) {
		return 0, retcards
	}
	cardIndex := sp.GameLogic.SwitchToCardIndex(cbCard)
	cardIndexP := byte(0)
	if cardIndex < 22 {
		if cardIndex == 4 {
			cardIndexP = 22
		} else if cardIndex == 10 {
			cardIndexP = 23
		} else if cardIndex == 20 {
			cardIndexP = 24
		} else if cardIndex == 12 {
			cardIndexP = 25
		} else if cardIndex == 16 {
			cardIndexP = 26
		}
		//优先用掉花牌
		if cardIndexP > 0 && cardIndexP < 27 && sp.CardIndex[wUser][cardIndexP] > 0 {
			for i := 0; i < int(sp.CardIndex[wUser][cardIndexP]) && retnum < 5 && retnum < iNum; i++ {
				retcards[retnum] = sp.GameLogic.SwitchToCardData(cardIndexP)
				retnum++
			}
		}
		if cardIndex < 27 {
			for i := 0; i < int(sp.CardIndex[wUser][cardIndex]) && retnum < 5 && retnum < iNum; i++ {
				retcards[retnum] = sp.GameLogic.SwitchToCardData(cardIndex)
				retnum++
			}
		}
	} else if cardIndex >= 22 && cardIndex < 27 {
		if cardIndex == 22 {
			cardIndexP = 4
		} else if cardIndex == 23 {
			cardIndexP = 10
		} else if cardIndex == 24 {
			cardIndexP = 20
		} else if cardIndex == 25 {
			cardIndexP = 12
		} else if cardIndex == 26 {
			cardIndexP = 16
		}
		//优先用掉花牌
		if cardIndex > 0 && cardIndex < 27 && sp.CardIndex[wUser][cardIndex] > 0 {
			for i := 0; i < int(sp.CardIndex[wUser][cardIndex]) && retnum < 5 && retnum < iNum; i++ {
				retcards[retnum] = sp.GameLogic.SwitchToCardData(cardIndex)
				retnum++
			}
		}
		if cardIndexP < 27 {
			for i := 0; i < int(sp.CardIndex[wUser][cardIndexP]) && retnum < 5 && retnum < iNum; i++ {
				retcards[retnum] = sp.GameLogic.SwitchToCardData(cardIndexP)
				retnum++
			}
		}
	}

	return retnum, retcards
}

// 删除一套观的牌
func (sp *Sport_zp_jzhp) DeleteGuan(wCenterUser uint16, index byte) bool {
	if int(wCenterUser) >= sp.GetPlayerCount() {
		return false
	}
	if index >= sp.UserGuanCardsCount[wCenterUser] {
		return false
	}
	if sp.UserGuanCardsCount[wCenterUser] == 0 {
		return false
	}

	for i := byte(0); i < 10; i++ {
		if i <= index {
			continue
		} else {
			copy(sp.UserGuanCards[wCenterUser][i-1][:], sp.UserGuanCards[wCenterUser][i][:])
		}
	}
	sp.UserGuanCards[wCenterUser][9] = [5]byte{} //最后一位必定是空的
	sp.UserGuanCardsCount[wCenterUser]--
	return true
}

// 判断牌是不是捡的
func (sp *Sport_zp_jzhp) CheckJian(wCenterUser uint16, cbCenterCard byte) bool {
	if int(wCenterUser) >= sp.GetPlayerCount() {
		return true
	}
	if !sp.GameLogic.IsValidCard(cbCenterCard) {
		return true
	}
	for _, card := range sp.UserJianCards[wCenterUser] {
		if card == cbCenterCard {
			return true
		} else if card == sp.GameLogic.GetParterCard(cbCenterCard) {
			return true
		}
	}
	return false
}

// 删除一张捡的牌
func (sp *Sport_zp_jzhp) DeleteJian(wCenterUser uint16, cbCenterCard byte) bool {
	if int(wCenterUser) >= sp.GetPlayerCount() {
		return false
	}
	if !sp.GameLogic.IsValidCard(cbCenterCard) {
		return false
	}
	for i, card := range sp.UserJianCards[wCenterUser] {
		if card == cbCenterCard {
			sp.UserJianCards[wCenterUser] = append(sp.UserJianCards[wCenterUser][:i], sp.UserJianCards[wCenterUser][i+1:]...)
			return true
		}
	}
	return false
}

// 判断牌是不是本轮捡的
func (sp *Sport_zp_jzhp) CheckJianCur(wCenterUser uint16, cbCenterCard byte) bool {
	if int(wCenterUser) >= sp.GetPlayerCount() {
		return true
	}
	if !sp.GameLogic.IsValidCard(cbCenterCard) {
		return true
	}
	for _, card := range sp.UserJianCardsCur[wCenterUser] {
		if card == cbCenterCard {
			return true
		} else if card == sp.GameLogic.GetParterCard(cbCenterCard) {
			return true
		}
	}
	return false
}

// 响应判断
func (sp *Sport_zp_jzhp) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int) bool {
	//变量定义
	bAroseAction := false

	//强制开朝需要用的2个变量

	//用户状态
	sp.Response = [TCGZ_MAX_PLAYER]bool{}
	sp.UserAction = [TCGZ_MAX_PLAYER]int{}
	sp.PerformAction = [TCGZ_MAX_PLAYER]int{}
	sp.CMD_OperateCard = [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard{}

	//动作判断
	for k := 0; k < sp.GetPlayerCount(); k++ {
		i := uint16((int(wCenterUser) + k + sp.GetPlayerCount()) % sp.GetPlayerCount())
		if i == sp.NoOutUser {
			continue
		}
		//用户过滤
		if wCenterUser == i && EstimatKind != ZP_EstimatKind_SendCard && EstimatKind != ZP_EstimatKind_GangCard {
			//发牌时自己也要分析，
			continue
		}
		if wCenterUser != i && (EstimatKind == ZP_EstimatKind_GangCard || EstimatKind == ZP_EstimatKind_SendCard) {
			//开朝后只分析自己还能不能胡,发牌时只有自己能判断能不能观滑胡
			continue
		}

		_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[i])

		//吃碰判断
		if EstimatKind != ZP_EstimatKind_GangCard && EstimatKind != ZP_EstimatKind_SendCard && sp.LeftCardCount >= 0 && !sp.CheckNeedGuoZhang(gKind_PengCard, i, cbCenterCard) {
			//碰牌判断
			sp.UserAction[i] |= sp.GameLogic.EstimatePengCard(byTempCardIndex, cbCenterCard, sp.TongInfo[i].TongCnt)
			if (sp.UserAction[i]&ZP_WIK_PENG) != 0 && sp.UserPengCount[i] >= 2 {
				sp.UserAction[i] ^= ZP_WIK_PENG
				sp.OnWriteGameRecord(i, "一局只能碰2次")
			}
			//捡牌判断--就判断出牌的人的下一个人
			wEatUser := uint16(sp.GetNextSeat(wCenterUser)) //位置和之前反过来
			if wEatUser == i {
				//sp.UserAction[i]|= ZP_WIK_JIAN
			}
		}

		//招牌判断
		if EstimatKind != ZP_EstimatKind_GangCard && sp.LeftCardCount >= 0 {
			sp.UserAction[i] |= sp.GameLogic.EstimateGangCard(byTempCardIndex, cbCenterCard, sp.TongInfo[i].TongCnt)
		}
		//滑牌判断
		if EstimatKind != ZP_EstimatKind_GangCard && sp.LeftCardCount >= 0 {
			sp.UserAction[i] |= sp.GameLogic.EstimateHuaCard(byTempCardIndex, cbCenterCard, sp.TongInfo[i].TongCnt)
		}

		//胡牌判断
		if len(sp.VecChiHuCard[i]) == 0 && sp.LeftCardCount >= 0 {
			//原始牌数据，恢复花牌需要用到这个数组
			CardIndexSrc := [TCGZ_MAX_INDEX_HUA]byte{}
			for c := byte(0); c < TCGZ_MAX_INDEX_HUA; c++ {
				CardIndexSrc[c] = sp.CardIndex[i][c]
			}
			//牌型权位
			//吃牌权位
			wChiHuRight := 0
			//吃胡判断
			cbWeaveCount := sp.WeaveItemCount[i]
			tempAction := 0
			tempAction, sp.ChiHuResult[i] = sp.GameLogic.AnalyseChiHuCard(CardIndexSrc, byTempCardIndex, sp.WeaveItemArray[i][:], cbWeaveCount, cbCenterCard, wChiHuRight)
			if (tempAction & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
				//v1.5 20200715 是否满足换统
				tempAction, sp.ChiHuResult[i] = sp.CheckHu2(i, cbCenterCard, sp.ChiHuResult[i], sp.TongInfo[i].TongCnt)
			}
			sp.UserAction[i] |= tempAction

			//10胡息起胡
			if (sp.UserAction[i] & ZP_WIK_CHI_HU) != ZP_WIK_NULL {

				//胡息计算
				iMaxIndex, iMaxHuXi, realHuXiInfo := sp.GameLogic.CalculateHuXi(&sp.ChiHuResult[i])
				sp.GameLogic.CalculateHuXi2(&sp.ChiHuResult[i], &iMaxIndex, &iMaxHuXi, &realHuXiInfo)

				//是否满足胡息要求
				bCanHu := false
				if iMaxHuXi >= sp.GeziShu {
					bCanHu = true
				}

				if !bCanHu {
					//不能胡时，去掉胡牌权限
					sp.UserAction[i] ^= ZP_WIK_CHI_HU
					sp.ChiHuResult[i].reset()
				}
			}
		}

		//结果判断
		if sp.UserAction[i] != ZP_WIK_NULL {
			bAroseAction = true
		}
	}
	//结果处理，标志为真说明上一个用户所出的牌有人可以所以要发送一个提示信息的框
	if bAroseAction == true {
		//设置变量
		sp.ProvideUser = wCenterUser
		sp.ProvideCard = cbCenterCard
		sp.ResumeUser = sp.CurrentUser
		sp.CurrentUser = static.INVALID_CHAIR

		// 是否存在多个玩家同时拥有牌权 且 不是全部托管
		bAllTuoGuan := true
		for seat := 0; seat < sp.GetPlayerCount(); seat++ {
			if sp.UserAction[seat] != ZP_WIK_NULL && !sp.TuoGuanPlayer[seat] {
				bAllTuoGuan = false
				break
			}
		}

		iOvertime := int64(0)
		if bAllTuoGuan {
			iOvertime = sp.AutoOutTime
		} else {
			iOvertime = sp.OperateTime
		}

		sp.GameState = ZP_AS_PLAYCARD // 更新为要打牌状态
		sp.setLimitedTime(int64(iOvertime + 1))

		//发送提示
		sp.SendOperateNotify(int(iOvertime))
		return true
	}
	return false
}

// 生成过张数据
func (sp *Sport_zp_jzhp) CreateGuoZhang(wUserAction int, wUserOperateCode int, wChairID uint16, cbCheckCard byte) bool {
	if wChairID >= TCGZ_MAX_PLAYER || wChairID >= uint16(sp.GetPlayerCount()) || !sp.GameLogic.IsValidCard(cbCheckCard) {
		return false
	}
	//游戏记录

	guozhangStr := "过张：不能 "

	bGuoZhangFlag := false
	if wUserOperateCode == ZP_WIK_NULL {
		//用户选择弃，其它牌权都需要过张
		if (wUserAction&ZP_WIK_HALF) != 0 || (wUserAction&ZP_WIK_LEFT) != 0 || (wUserAction&ZP_WIK_CENTER) != 0 || (wUserAction&ZP_WIK_RIGHT) != 0 || (wUserAction&ZP_WIK_JIAN) != 0 {
			//gKind_EatCard过张类型
			sp.VecChiCard[wChairID] = append(sp.VecChiCard[wChairID], cbCheckCard)
			bGuoZhangFlag = true
			guozhangStr += "吃 "
		}
		if (wUserAction & ZP_WIK_PENG) != 0 {
			//弃绍不能碰
			//gKind_PengCard过张类型
			sp.VecPengCard[wChairID] = append(sp.VecPengCard[wChairID], cbCheckCard)
			bGuoZhangFlag = true
			guozhangStr += "碰 "
		}
		if (wUserAction & ZP_WIK_CHI_HU) != 0 {
			//gKind_ChiHuCard过张类型
			sp.VecChiHuCard[wChairID] = append(sp.VecChiHuCard[wChairID], cbCheckCard)
			bGuoZhangFlag = true
			guozhangStr += "胡 "
		}
	} else if (wUserOperateCode&ZP_WIK_LEFT) != 0 || (wUserOperateCode&ZP_WIK_CENTER) != 0 || (wUserOperateCode&ZP_WIK_RIGHT) != 0 || (wUserOperateCode&ZP_WIK_HALF) != 0 || (wUserOperateCode&ZP_WIK_JIAN) != 0 {
		//用户选择吃，碰和胡的牌权都需要过张
		if (wUserAction & ZP_WIK_PENG) != 0 {
			//gKind_PengCard过张类型
			sp.VecPengCard[wChairID] = append(sp.VecPengCard[wChairID], cbCheckCard)
			bGuoZhangFlag = true
			guozhangStr += "碰 "
		}

		if (wUserAction & ZP_WIK_CHI_HU) != 0 {
			//gKind_ChiHuCard过张类型
			sp.VecChiHuCard[wChairID] = append(sp.VecChiHuCard[wChairID], cbCheckCard)
			bGuoZhangFlag = true
			guozhangStr += "胡 "
		}
	} else if (wUserOperateCode & ZP_WIK_PENG) != 0 {

		//用户选择碰，胡的牌权需要过张
		if (wUserAction & ZP_WIK_CHI_HU) != 0 {
			//gKind_ChiHuCard过张类型
			sp.VecChiHuCard[wChairID] = append(sp.VecChiHuCard[wChairID], cbCheckCard)
			bGuoZhangFlag = true
			guozhangStr += "胡 "
		}
	}
	if bGuoZhangFlag == true {
		//记录
		szGameRecord := fmt.Sprintf("这张牌，%s", sp.GameLogic.SwitchToCardName2(sp.GameLogic.SwitchToCardIndex(cbCheckCard)))
		guozhangStr += szGameRecord
		sp.OnWriteGameRecord(wChairID, guozhangStr)
	}
	return true
}

// 校验是否需要等待过张
func (sp *Sport_zp_jzhp) CheckNeedGuoZhang(wKind int, wChairID uint16, cbCheckCard byte) bool {
	if wChairID >= TCGZ_MAX_PLAYER || wChairID >= uint16(sp.GetPlayerCount()) || !sp.GameLogic.IsValidCard(cbCheckCard) {
		return false
	}
	if wKind == gKind_EatCard {
		for i := 0; i < len(sp.VecChiCard[wChairID]); i++ {
			if cbCheckCard == sp.VecChiCard[wChairID][i] || sp.GameLogic.GetParterCard(cbCheckCard) == sp.VecChiCard[wChairID][i] {
				return true
			}
		}
	} else if wKind == gKind_PengCard {
		for i := 0; i < len(sp.VecPengCard[wChairID]); i++ {
			if cbCheckCard == sp.VecPengCard[wChairID][i] || sp.GameLogic.GetParterCard(cbCheckCard) == sp.VecPengCard[wChairID][i] {
				return true
			}
		}
	} else if wKind == gKind_HuCard {
		//for i := 0; i < len(sp.VecChiHuCard[wChairID]); i++{
		//	if (cbCheckCard == sp.VecChiHuCard[wChairID][i]||sp.GameLogic.GetParterCard(cbCheckCard) == sp.VecChiHuCard[wChairID][i]) {
		//		return true;
		//	}
		//}
		if len(sp.VecChiHuCard[wChairID]) > 0 {
			return true
		}
	}
	return false
}

// 解除过张，目前只有胡牌的过张可以被解除（在手牌变动的情况下），20180809添加吃的解除过张
func (sp *Sport_zp_jzhp) DismissGuoZhang(wKind int, wChairID uint16, cbCheckCard byte) bool {
	if wChairID >= TCGZ_MAX_PLAYER || wChairID >= uint16(sp.GetPlayerCount()) {
		return false
	}
	if wKind == gKind_HuCard && len(sp.VecChiHuCard[wChairID]) > 0 {
		byCardIndex := [TCGZ_MAX_INDEX_HUA]byte{} //记录日志需要的数据
		for byCardI := 0; byCardI < len(sp.VecChiHuCard[wChairID]); byCardI++ {
			if sp.GameLogic.IsValidCard(sp.VecChiHuCard[wChairID][byCardI]) {
				byIndex := sp.GameLogic.SwitchToCardIndex(sp.VecChiHuCard[wChairID][byCardI])
				if byIndex < TCGZ_MAX_INDEX_HUA {
					byCardIndex[byIndex]++
				}
			}
		}

		sp.VecChiHuCard[wChairID] = []byte{}

		//游戏记录
		guozhangStr := "变动了手牌，解除胡牌的过张"
		szGameRecord := fmt.Sprintf(",牌数据: %s", sp.GameLogic.SwitchToCardName1(byCardIndex[:]))
		guozhangStr += szGameRecord
		sp.OnWriteGameRecord(wChairID, guozhangStr)
	} else if wKind == gKind_EatCard && len(sp.VecChiCard[wChairID]) > 0 {
		byCardI := len(sp.VecChiCard[wChairID]) - 1
		if cbCheckCard == sp.VecChiCard[wChairID][byCardI] {
			byIndex := sp.GameLogic.SwitchToCardIndex(cbCheckCard)

			sp.VecChiCard[wChairID] = append(sp.VecChiCard[wChairID][0 : len(sp.VecChiCard[wChairID])-1])

			//游戏记录
			guozhangStr := "高权限用户操作后，解除吃牌的过张"
			szGameRecord := fmt.Sprintf(",牌数据: %s", sp.GameLogic.SwitchToCardName2(byIndex))
			guozhangStr += szGameRecord
			sp.OnWriteGameRecord(wChairID, guozhangStr)
		}
	} else if wKind == gKind_PengCard && len(sp.VecPengCard[wChairID]) > 0 {
		//byCardI := len(sp.VecPengCard[wChairID]) - 1;
		//if (cbCheckCard == sp.VecPengCard[wChairID][byCardI]){
		//	byIndex := sp.GameLogic.SwitchToCardIndex(cbCheckCard);
		//
		//	sp.VecPengCard[wChairID]=append(sp.VecPengCard[wChairID][0:len(sp.VecPengCard[wChairID])-1]);
		//
		//	//游戏记录
		//	guozhangStr :="高权限用户操作后，解除碰牌的过张"
		//	szGameRecord:=fmt.Sprintf(",牌数据: %s",sp.GameLogic.SwitchToCardName2(byIndex));
		//	guozhangStr+=szGameRecord
		//	sp.OnWriteGameRecord(wChairID,guozhangStr)
		//}
		byCardIndex := [TCGZ_MAX_INDEX_HUA]byte{} //记录日志需要的数据
		for byCardI := 0; byCardI < len(sp.VecPengCard[wChairID]); byCardI++ {
			if sp.GameLogic.IsValidCard(sp.VecPengCard[wChairID][byCardI]) {
				byIndex := sp.GameLogic.SwitchToCardIndex(sp.VecPengCard[wChairID][byCardI])
				if byIndex < TCGZ_MAX_INDEX_HUA {
					byCardIndex[byIndex]++
				}
			}
		}

		sp.VecPengCard[wChairID] = []byte{}

		//游戏记录
		guozhangStr := "变动了手牌，解除碰牌的过张"
		szGameRecord := fmt.Sprintf(",牌数据: %s", sp.GameLogic.SwitchToCardName1(byCardIndex[:]))
		guozhangStr += szGameRecord
		sp.OnWriteGameRecord(wChairID, guozhangStr)
	}
	return true
}

// v1.5 20200715 获取优先玩家，不考虑是否已经选择了牌权
func (sp *Sport_zp_jzhp) GetPriUser(wTargetAction int) uint16 {
	//if (wTargetUser >= MAX_PLAYER || wTargetUser >= uint16(sp.GetPlayerCount())) {
	//	return 0,0,0;
	//}
	firstUser := sp.ProvideUser % uint16(sp.GetPlayerCount())
	if !sp.SendStatus && int(sp.ProvideUser) < sp.GetPlayerCount() {
		firstUser = (sp.ProvideUser + 1) % uint16(sp.GetPlayerCount())
	}
	if wTargetAction != 0 {
		for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
			tempUser := (firstUser + i) % uint16(sp.GetPlayerCount())
			if (sp.UserAction[tempUser] & wTargetAction) != 0 {
				return tempUser
			}
		}
	}
	return firstUser
}

// 获取供牌在胡牌组合中的位置，0xFF表示在牌眼中，0-9表示在胡牌组合中,entype ==1 表示自摸
func (sp *Sport_zp_jzhp) GetProvidCardPos(entype byte, wChairID uint16, byIndex byte, byProvidCard byte) byte {
	if int(wChairID) >= sp.GetPlayerCount() || byIndex >= 10 || !sp.GameLogic.IsValidCard(byProvidCard) {
		return 0xFE
	}
	byWeaveCnt := sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveCount
	byWeaveCnt = byWeaveCnt
	if byWeaveCnt > 10 {
		byWeaveCnt = 10
	}
	for byWeaveIdx := byte(0); byWeaveIdx < byte(byWeaveCnt); byWeaveIdx++ {
		wWeaveK := sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveKind[byWeaveIdx]
		if wWeaveK == ZP_WIK_KAN && entype != 1 {
			continue
		}
		if sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].IsWeave[byWeaveIdx] == 1 {
			continue
		}

		byFindFlag := 0
		if sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveItemInfo[byWeaveIdx][0] == byProvidCard {
			byFindFlag = 1
		} else if sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveItemInfo[byWeaveIdx][1] == byProvidCard {
			byFindFlag = 1
		} else if sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveItemInfo[byWeaveIdx][2] == byProvidCard {
			byFindFlag = 1
		} else if sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveItemInfo[byWeaveIdx][3] == byProvidCard {
			byFindFlag = 1
		} else if sp.ChiHuResult[wChairID].ChiHuItemInfoArray[byIndex].WeaveItemInfo[byWeaveIdx][4] == byProvidCard {
			byFindFlag = 1
		}

		if byFindFlag == 1 {
			return byWeaveIdx
		}
	}
	return 0xFF
}

// 获取一张牌的原始数目
func (sp *Sport_zp_jzhp) GetCardsOriginalNum(cardIndex byte) int {
	if cardIndex == 4 || cardIndex == 10 || cardIndex == 12 || cardIndex == 16 || cardIndex == 20 {
		return 3
	} else if cardIndex >= 22 && cardIndex <= 26 {
		return 2
	} else if cardIndex == 27 {
		if sp.HunJiang == 0 {
			return 0
		} else {
			return 2
		}
	} else {
		return 5
	}
	return 5
}

// 获取所有牌的总数目
func (sp *Sport_zp_jzhp) GetCardsNum(cbCardIndex [TCGZ_MAX_INDEX_HUA]byte) int {
	leftCardsNum := 0
	for i := 0; i < TCGZ_MAX_INDEX_HUA; i++ {
		if sp.GameLogic.IsValidCard(cbCardIndex[i]) {
			leftCardsNum++
		}
	}
	return leftCardsNum
}

// 检查胡牌要求，换统的需求 v1.5 20200715。  线上问题20200724 cbCenterCard == 0表示自摸
func (sp *Sport_zp_jzhp) CheckHu2(wSeat uint16, cbCenterCard byte, ChiHuResult TagChiHuResult, TongCnt int) (int, TagChiHuResult) {
	var OKChiHuResult TagChiHuResult //确定满足要求的胡牌数据
	if ChiHuResult.ChiHuKind == 0 {
		return 0, OKChiHuResult
	}
	if len(ChiHuResult.ChiHuItemInfoArray) == 0 {
		return 0, OKChiHuResult
	}
	if /*!sp.GameLogic.IsValidCard(cbCenterCard) ||*/ int(wSeat) >= sp.GetPlayerCount() || TongCnt <= 0 {
		return ZP_WIK_CHI_HU, ChiHuResult
	}
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	for byIndex := 0; byIndex < byCount; byIndex++ {
		okFlag := true
		tempTongCnt := 0
		cardsNum := sp.GameLogic.GetCardsNumInResult(cbCenterCard, ChiHuResult.ChiHuItemInfoArray[byIndex])
		for j := 0; j < ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveCount && j < 10; j++ {
			tempKind := ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j]
			if ZP_WIK_TIANLONG == tempKind {
				tempTongCnt++
			} else if ZP_WIK_HUA == tempKind && ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[j] == 0 {
				//20200724真正的滑没有统数
				if ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[j] == 0 {
					//20200724如果是滑胡的没有统数
					if sp.GameLogic.IsValidCard(cbCenterCard) && sp.GameLogic.SwitchToCardIndexNoHua(cbCenterCard) == sp.GameLogic.SwitchToCardIndexNoHua(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) && cardsNum <= 5 {
					} else {
						tempTongCnt += 2
					}
				}
			}
		}
		if tempTongCnt < TongCnt {
			okFlag = false
		}
		if okFlag {
			OKChiHuResult.ChiHuItemInfoArray = append(OKChiHuResult.ChiHuItemInfoArray, ChiHuResult.ChiHuItemInfoArray[byIndex])
		}
	}
	if len(OKChiHuResult.ChiHuItemInfoArray) > 0 {
		OKChiHuResult.ChiHuKind = ZP_WIK_CHI_HU
		OKChiHuResult.ChiHuRight = ChiHuResult.ChiHuRight
		return ZP_WIK_CHI_HU, OKChiHuResult
	}
	return 0, OKChiHuResult
}

// 发送听牌信息
func (sp *Sport_zp_jzhp) SendTingInfo(wTheSeat uint16, tingresult TagTingCardResult) {
	if int(wTheSeat) >= sp.GetPlayerCount() {
		return
	}
	var tinginfo static.Msg_S_ZP_TingInfo
	tinginfo.TheSeat = wTheSeat
	tinginfo.TingInfo.MaxCount = tingresult.MaxCount
	tinginfo.TingInfo.Seat = tingresult.Seat
	copy(tinginfo.TingInfo.TingIndex[:], tingresult.TingIndex[:])
	copy(tinginfo.TingInfo.TingNumber[:], tingresult.TingNumber[:])
	copy(tinginfo.TingInfo.TingFanShu[:], tingresult.TingFanShu[:])
	sp.SendPersonMsg(consts.MsgTypeTingInfo, tinginfo, wTheSeat)
}
func (sp *Sport_zp_jzhp) CheckTing(wChairid uint16) (int, TagTingCardResult) {
	//玩家组合牌区数据
	cbWeaveCount := [TCGZ_MAX_PLAYER]byte{}
	copy(cbWeaveCount[:], sp.WeaveItemCount[:])
	WeaveItem := [TCGZ_MAX_PLAYER][10]TagWeaveItem{}
	copy(WeaveItem[:], sp.WeaveItemArray[:])

	//丢弃的牌
	cbDiscardCount := [TCGZ_MAX_PLAYER]byte{}
	cbDiscardCard := [TCGZ_MAX_PLAYER][]byte{}
	for seat := 0; seat < TCGZ_MAX_PLAYER && seat < sp.GetPlayerCount(); seat++ {
		//DiscardCard中包括了DispatchCard
		//cbDiscardCount[seat]= byte(len(sp.DispatchCard[seat]))
		//for i := 0;i < len(sp.DispatchCard[seat]);i++{
		//	cbDiscardCard[seat]= append(cbDiscardCard[seat],byte(sp.DispatchCard[seat][i]))
		//}
		cbDiscardCount[seat] += sp.DiscardCount[seat]
		for i := byte(0); i < sp.DiscardCount[seat]; i++ {
			cbDiscardCard[seat] = append(cbDiscardCard[seat], sp.DiscardCard[seat][i])
		}
	}
	//发的牌或打的牌正在等待响应的话怎么处理？
	tingcount, TingCardResult := sp.AnalyseTingCardCount(wChairid == sp.Banker, wChairid, sp.CardIndex[wChairid], WeaveItem, cbWeaveCount, cbDiscardCount, cbDiscardCard)

	return tingcount, TingCardResult
}
func (sp *Sport_zp_jzhp) AnalyseTingCardCount(isbank bool, bySeat uint16, cbCardIndex [TCGZ_MAX_INDEX_HUA]byte, WeaveItem [TCGZ_MAX_PLAYER][10]TagWeaveItem, cbItemCount [TCGZ_MAX_PLAYER]byte, cbDiscardCount [TCGZ_MAX_PLAYER]byte, cbDiscardCard [TCGZ_MAX_PLAYER][]byte) (int, TagTingCardResult) {
	var TingCardResult TagTingCardResult
	TingCardResult.Seat = bySeat
	TingCardResult.MaxCount = TCGZ_MAX_INDEX_HUA
	if bySeat >= TCGZ_MAX_PLAYER {
		return 0, TingCardResult
	}

	//变量定义
	var ChiHuResult TagChiHuResult
	ChiHuResult.reset()
	//听牌分析
	y := 0
	for byI := byte(0); byI < TCGZ_MAX_INDEX_HUA-1; byI++ {
		//-1是因为胡牌算法的恢复花牌的部分代码不支持赖子，赖子的数目由客户端去算
		//胡牌分析
		cbCurrentCard := sp.GameLogic.SwitchToCardData(byI)
		cbDiscartNum := 0 //丢弃扑克数
		for i := 0; i < TCGZ_MAX_PLAYER; i++ {
			for j := byte(0); j < cbDiscardCount[i]; j++ {
				if sp.GameLogic.IsValidCard(cbDiscardCard[i][j]) && byI == sp.GameLogic.SwitchToCardIndex(cbDiscardCard[i][j]) {
					cbDiscartNum++
				}
			}
		}
		cbHandCard := int(cbCardIndex[byI]) //手中这张扑克数

		cbWeaveNum := 0 //组合牌中这张扑克数
		for i := 0; i < TCGZ_MAX_PLAYER; i++ {
			for k := byte(0); k < 10 && k < cbItemCount[i]; k++ {
				for m := 0; m < 5; m++ {
					if sp.GameLogic.IsValidCard(WeaveItem[i][k].Cards[m]) && byI == sp.GameLogic.SwitchToCardIndex(WeaveItem[i][k].Cards[m]) {
						cbWeaveNum++
					}
				}
			}
		}
		if cbDiscartNum+cbHandCard+cbWeaveNum >= sp.GetCardsOriginalNum(byI) {
			continue //不能胡，没有牌了//bugid 7567 ,已经没有牌了，听牌提示要提示0张
		}

		//构造扑克
		_, byTempCardIndex := sp.GameLogic.SwitchIndexToIndexNoHua(sp.CardIndex[bySeat])

		//原始牌数据，恢复花牌需要用到这个数组
		CardIndexSrc := [TCGZ_MAX_INDEX_HUA]byte{}
		for c := byte(0); c < TCGZ_MAX_INDEX_HUA; c++ {
			CardIndexSrc[c] = sp.CardIndex[bySeat][c]
		}

		//按照自摸处理，将cbCurrentCard放到手上后在将AnalyseChiHuCard 中的cbCurrentCard参数置0
		if cbCurrentCard != 0 {
			byTempCardIndex[sp.GameLogic.SwitchToCardIndexNoHua(cbCurrentCard)]++
			CardIndexSrc[sp.GameLogic.SwitchToCardIndex(cbCurrentCard)]++
		}
		//牌型权位
		//吃牌权位
		ChiHuResult.reset()
		wChiHuRight := 0
		//吃胡判断
		cbWeaveCount := sp.WeaveItemCount[bySeat]
		tempAction := 0
		tempAction, ChiHuResult = sp.GameLogic.AnalyseChiHuCard(CardIndexSrc, byTempCardIndex, sp.WeaveItemArray[bySeat][:], cbWeaveCount, 0, wChiHuRight)
		if (tempAction & ZP_WIK_CHI_HU) != ZP_WIK_NULL {
			//v1.5 20200715 是否满足换统
			tempAction, ChiHuResult = sp.CheckHu2(bySeat, 0, ChiHuResult, sp.TongInfo[bySeat].TongCnt)
		}
		//10胡息起胡
		if (tempAction & ZP_WIK_CHI_HU) != ZP_WIK_NULL {

			//胡息计算
			//胡息计算
			iMaxIndex, iMaxHuXi, realHuXiInfo := sp.GameLogic.CalculateHuXi(&ChiHuResult)
			sp.GameLogic.CalculateHuXi2(&ChiHuResult, &iMaxIndex, &iMaxHuXi, &realHuXiInfo)

			//是否满足胡息要求
			bCanHu := false
			if iMaxHuXi >= sp.GeziShu {
				bCanHu = true
			}
			if bCanHu {
				y++ //计数
				TingCardResult.TingIndex[byI] = true
				TingCardResult.TingNumber[byI] = 0
				if cbDiscartNum+cbHandCard+cbWeaveNum < sp.GetCardsOriginalNum(byI) {
					TingCardResult.TingNumber[byI] = sp.GetCardsOriginalNum(byI) - (cbDiscartNum + cbHandCard + cbWeaveNum)
				}
			} else {
				TingCardResult.TingIndex[byI] = false
				TingCardResult.TingNumber[byI] = 0
				TingCardResult.TingFanShu[byI] = 0
			}
		} else {
			TingCardResult.TingIndex[byI] = false
			TingCardResult.TingNumber[byI] = 0
			TingCardResult.TingFanShu[byI] = 0
		}
	}
	return y, TingCardResult
}

// 获取牌权名字
func (sp *Sport_zp_jzhp) GetWikStr(wUserAction int) string {
	if wUserAction == 0 {
		return ""
	}
	//游戏牌权记录
	wikStr := "牌权:"
	if (wUserAction & ZP_WIK_CHI_HU) != 0 {
		wikStr += "胡|"
	}
	if (wUserAction & ZP_WIK_TA) != 0 {
		wikStr += "踏|"
	}
	if (wUserAction & ZP_WIK_HUA) != 0 {
		wikStr += "泛|"
	}
	if (wUserAction & ZP_WIK_TIANLONG) != 0 {
		wikStr += "统|"
	}
	if (wUserAction & ZP_WIK_GANG) != 0 {
		wikStr += "招|"
	}
	if (wUserAction & ZP_WIK_PENG) != 0 {
		wikStr += "碰|"
	}
	if (wUserAction&ZP_WIK_HALF) != 0 || (wUserAction&ZP_WIK_LEFT) != 0 || (wUserAction&ZP_WIK_CENTER) != 0 || (wUserAction&ZP_WIK_RIGHT) != 0 {
		wikStr += "吃|"
	}
	if (wUserAction & ZP_WIK_JIAN) != 0 {
		wikStr += "捡|"
	}

	return wikStr
}

// ! 加载测试麻将数据
func (sp *Sport_zp_jzhp) initDebugCards(configName string, cbRepertoryCard *[TCGZ_MAX_PLAYER][TCGZ_MAX_INDEX_HUA]byte, wBankerUser *uint16) (err error) {
	defer func() {
		if err != nil {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, err.Error())
		}
	}()
	repertoryCnt := byte(37) //牌库剩余牌数量，3人36,2人61 含2张赖子
	if sp.GetPlayerCount() == 2 {
		repertoryCnt = 62
	}
	if sp.HunJiang != 1 {
		repertoryCnt -= 2
	}
	//! 做牌文件配置
	var debugCardConfig *meta2.CardConfig = new(meta2.CardConfig)

	fileName := fmt.Sprintf("./%s%d_%d", configName, sp.GetPlayerCount(), sp.Rule.FangZhuID)
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始根据房主id读取做牌文件，文件名："+fileName)
	if !static.GetJsonMgr().ReadData("./json", fileName, debugCardConfig) {
		configName = fmt.Sprintf("./%s%d", configName, sp.GetPlayerCount())
		sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始读取做牌文件，文件名："+configName)
		if !static.GetJsonMgr().ReadData("./json", configName, debugCardConfig) {
			return errors.New("做牌文件:读取失败")
		}
	}

	// 是否开启做牌
	if debugCardConfig.IsAble == 1 {
		//检查做牌文件是否做牌异常
		for _, handCards := range debugCardConfig.UserCards {
			if len(strings.Split(handCards, ",")) != TCGZ_MAXSENDCARD-1 {
				///25张牌
				return errors.New("做牌文件:手牌长度不对")
			}
		}
		//检查牌堆牌是否正常
		if len(debugCardConfig.RepertoryCard) != debugCardConfig.RepertoryCardCount*5-1 {
			return errors.New(fmt.Sprintf("做牌文件:牌库牌数量不一致:::RepertoryCard:[%d]>>>实际做牌牌库数量:[%d]", debugCardConfig.RepertoryCardCount, (len(debugCardConfig.RepertoryCard)+1)/5))
		}
		// 设置玩家手牌

		for userIndex, handCards := range debugCardConfig.UserCards {
			byCardsCount := 0
			_item := sp.GetUserItemByChair(uint16(userIndex))
			if _item != nil {
				_item.Ctx.InitCardIndex()                                  //清理手牌
				(*cbRepertoryCard)[userIndex] = [TCGZ_MAX_INDEX_HUA]byte{} //清理手牌
				handCardsSlice := strings.Split(handCards, ",")
				for _, cardStr := range handCardsSlice {

					if cardValue, _ := sp.SportMetaJZHP.getCardDataByStr(cardStr); cardValue == static.INVALID_BYTE {
						//return errors.New(fmt.Sprintf("做牌文件:第%d号玩家做牌异常：%s", userIndex, cardStr))
					} else {
						//_item.Ctx.DispatchCard(cardValue)
						if !sp.GameLogic.IsValidCard(cardValue) {
							sp.OnWriteGameRecord(uint16(userIndex), "做牌数据异常")
							continue
						}
						(*cbRepertoryCard)[userIndex][sp.GameLogic.SwitchToCardIndex(cardValue)]++
						byCardsCount++
						if byCardsCount >= TCGZ_MAXSENDCARD-1 {
							break
						}
					}
				}
			}
		}
		//设置牌堆牌
		sp.LeftCardCount = 0 //做牌的牌堆数目要重新算
		sp.RepertoryCard = [TCGZ_ALLCARD]byte{}
		repertoryCards := strings.Split(debugCardConfig.RepertoryCard, ",")
		for _, cardStr := range repertoryCards {
			if cardValue, _ := sp.SportMetaJZHP.getCardDataByStr(cardStr); cardValue == static.INVALID_BYTE {
				//return errors.New(fmt.Sprintf("做牌文件:牌堆第%d个做牌异常：%s", cardIndex, cardStr))
			} else {
				if !sp.GameLogic.IsValidCard(cardValue) {
					sp.OnWriteGameRecord(static.INVALID_CHAIR, "做牌数据异常")
					continue
				}
				sp.RepertoryCard[sp.LeftCardCount] = cardValue
				sp.LeftCardCount++
				if sp.LeftCardCount >= repertoryCnt {
					break //底牌最多53个或34个
				}
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

func (sp *Sport_zp_jzhp) addReplayOrder(chairId uint16, operation int, value int, values []int) {
	var order ZP_Replay_Order
	order.R_ChairId = chairId
	order.R_Opt = operation

	if operation == E_ZP_SendCard {
	} else if operation == E_ZP_OutCard {
	} else if operation == E_ZP_Wik_Left || operation == E_ZP_Wik_Center || operation == E_ZP_Wik_Right || operation == E_ZP_Wik_1X2D || operation == E_ZP_Wik_2X1D || operation == E_ZP_Wik_2710 {
	} else if operation == E_ZP_Peng {
	} else if operation == E_ZP_TianLong {
	} else if operation == E_ZP_Gang {
	} else if operation == E_ZP_Hu {
	} else if operation == E_ZP_HuangZhuang {
	} else if operation == E_ZP_Li_Xian {
	} else if operation == E_ZP_Jie_san {
	}
	//所有动作是一样的操作
	if len(values) != 0 {
		order.R_Value = append(order.R_Value, values[:]...)
	} else {
		order.R_Value = append(order.R_Value, value)
	}
	sp.ReplayRecord.R_Orders = append(sp.ReplayRecord.R_Orders, order)
}

// ! 解散
func (sp *Sport_zp_jzhp) OnEnd() {
	if sp.IsGameStarted() {
		sp.OnGameOver(static.INVALID_CHAIR, static.GER_DISMISS)
	}
}

// ! 单局结算
func (sp *Sport_zp_jzhp) OnGameOver(wChairID uint16, cbReason byte) bool {
	sp.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 游戏结束, 流局结束，统计积分
func (sp *Sport_zp_jzhp) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if sp.GetGameStatus() == static.GS_MJ_END && cbReason == static.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}
	if sp.GameEndStatus == static.GS_MJ_END {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	// 清除超时检测
	for _, v := range sp.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
		//20200424 小结算的时候玩家不准备
		v.Ready = false
	}

	switch cbReason {
	case static.GER_NORMAL: //常规结束
		return sp.OnGameEndNormal(wChairID, cbReason)
	case static.GER_USER_LEFT: //用户强退
		return sp.OnGameEndUserLeft(wChairID, cbReason)
	case static.GER_DISMISS: //解散游戏
		return sp.OnGameEndDissmiss(wChairID, cbReason)
	case static.GER_GAME_ERROR: //程序异常，解散游戏
		return sp.OnGameEndErrorDissmiss(wChairID, cbReason)
	}
	return false
}

func (sp *Sport_zp_jzhp) FakeSwitch(ProvideCard byte, wChairID uint16, iswave [10]int, info *[10][10]byte, weavekind [10]int) []byte {
	//接炮时，如果牌类型出现了碰，说明接的牌在碰里面，要保留一个不被赖子替换
	findflag := false
	Fakecards := []byte{}
	cardIndex := [TCGZ_MAX_INDEX_HUA]byte{}
	copy(cardIndex[:], sp.CardIndex[wChairID][:])
	for i := len(*info) - 1; i >= 0; i-- {
		if iswave[i] == 1 {
			continue
		}
		for j := 0; j < 5; j++ {
			card := (*info)[i][j]
			if sp.GameLogic.IsValidCard(card) {
				cardindex := sp.GameLogic.SwitchToCardIndex(card)
				if cardIndex[cardindex] > 0 {
					cardIndex[cardindex]--
				} else {
					if findflag == false && weavekind[i] == ZP_WIK_PENG && ProvideCard == card {
						//不被替换
						findflag = true
					} else {
						(*info)[i][j] = 0x1C //这里是赖子
						Fakecards = append(Fakecards, card)
					}
				}
			}
		}
	}
	return Fakecards
}

func (sp *Sport_zp_jzhp) AdjustKanHuXi(iswave [10]int, info [10][10]byte, weavekind [10]int, jingIndex byte) (bool, [10]int, [10]int) {
	retKan := [10]int{}
	retMainJing := [10]int{}
	retFlag := false
	for i := 0; i < 10; i++ {
		if iswave[i] == 1 {
			continue
		}
		if weavekind[i] != ZP_WIK_KAN {
			continue
		}
		bret := false
		retk := 0
		hua := 0
		for j := 0; j < 5; j++ {
			card := info[i][j]
			if sp.GameLogic.IsValidCard(card) {
				if 0x1C == card {
					retk++
					bret = true
				}
				if 1 == sp.GameLogic.IsHuaPai(card) && 1 == sp.GameLogic.IsJingPai(card) {
					hua++
					if sp.GameLogic.SwitchToCardIndexNoHua(card) == jingIndex {
						retMainJing[i] = 1
					}
				}
			}
		}
		if hua+retk == 3 {
			retFlag = bret
			retKan[i] = retk
		}
	}
	return retFlag, retKan, retMainJing
}

// ! 结束，结束游戏 wChairID胜者、解散者，或无效值
func (sp *Sport_zp_jzhp) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	if cbReason != static.GER_NORMAL {
		return false
	}
	//设置游戏结束状态
	//sp.SetGameStatus(public.GS_MJ_END)
	sp.GameEndStatus = static.GS_MJ_END

	//定义变量
	iWinnerCnt := 0
	nWinner := static.INVALID_CHAIR

	var endgameMsg static.Msg_S_ZP_TC_GameEnd
	endgameMsg.EndSubType = 0
	endgameMsg.EndStatus = cbReason

	endgameMsg.ProvideUser = sp.ProvideUser
	endgameMsg.BankUser = sp.Banker
	endgameMsg.ChiHuCard = sp.ChiHuCard
	endgameMsg.ChiHuUserCount = 1

	////////////////////////////////////////////////////////////////////////////////////////////
	//	有人胡牌：//1.自摸   2.放炮
	if sp.ProvideUser != static.INVALID_CHAIR {
		for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
			WinerItem := sp.GetUserItemByChair(uint16(i))
			if WinerItem == nil {
				continue
			}
			if (sp.UserAction[i]&ZP_WIK_CHI_HU) == ZP_WIK_NULL || sp.ChiHuResult[i].ChiHuKind == ZP_WIK_NULL {
				continue
			}
			//非一炮多响时必须要有主动点击事件
			if !sp.DuoHu && ((sp.PerformAction[i]&ZP_WIK_CHI_HU) == ZP_WIK_NULL || i != wChairID) {
				continue
			}
			//先排序 ,句子在后
			sp.GameLogic.SortChiHuResult(1, &sp.ChiHuResult[i])

			iWinnerCnt++
			nWinner = i
			//胡息计算
			byMaxIndex, iMaxHuXi, realHuXiInfo := sp.GameLogic.CalculateHuXi(&sp.ChiHuResult[i])
			sp.GameLogic.CalculateHuXi2(&sp.ChiHuResult[i], &byMaxIndex, &iMaxHuXi, &realHuXiInfo)
			if byMaxIndex >= len(sp.ChiHuResult[i].ChiHuItemInfoArray) {
				sp.OnWriteGameRecord(static.INVALID_CHAIR, "byMaxIndex 越界0")
				byMaxIndex = 0 //这时已经错了，只是不让它崩溃
			}
			endgameMsg.JingPai = 0
			byJingIndex := sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].MainJingIndex
			if byJingIndex < 23 {
				endgameMsg.JingPai = sp.GameLogic.SwitchToCardData(byJingIndex)
			}

			cardsNum := sp.GameLogic.GetCardsNumInResult(endgameMsg.ChiHuCard, sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex])

			//FakeSwitch要放在CalculateMainJing之后使用。
			sp.FakeSwitch(sp.ProvideCard, i, sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].IsWeave, &sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveItemInfo, sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveKind)

			//在这里修正带精的坎的赖子数不正确的问题，从4张和5张的看，胡数和赖子无关的，但3张的却有关?
			//........有点麻烦
			retFlag, retKan, retM := sp.AdjustKanHuXi(sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].IsWeave, sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveItemInfo, sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveKind, byJingIndex)
			if retFlag {
				for ii := 0; ii < 10; ii++ {
					if retKan[ii] == 1 {
						adjust := 3 //12-9
						if retM[ii] == 1 {
							adjust = 3 * 2
						}
						realHuXiInfo[ii] -= adjust
						iMaxHuXi -= adjust
					}
				}
			}

			byHutye := 0
			//是否满足胡息要求
			bCanHu := false
			if iMaxHuXi >= sp.GeziShu {
				bCanHu = true
			}
			_, retcards := sp.GetAllCardsbyChairid(i, 0)

			if bCanHu {
				//胡息等
				endgameMsg.TotalHuxi[i] = iMaxHuXi
				endgameMsg.WinOrLose[i] = 1
				endgameMsg.ChiHuKind[i] = byHutye
				if sp.MaxHUxi[i] < iMaxHuXi {
					sp.MaxHUxi[i] = iMaxHuXi
				}

				if byMaxIndex >= len(sp.ChiHuResult[i].ChiHuItemInfoArray) {
					sp.OnWriteGameRecord(static.INVALID_CHAIR, "byMaxIndex 越界")
					byMaxIndex = 0 //这时已经错了，只是不让它崩溃
				} else {
					endgameMsg.WeaveCount[i] = byte(sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveCount)
					if endgameMsg.WeaveCount[i] > 10 {
						endgameMsg.WeaveCount[i] = 10
					}
					//copy(endgameMsg.HuXi[i][:],sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].HuXiInfo[:])
					copy(endgameMsg.HuXi[i][:], realHuXiInfo[:])
					copy(endgameMsg.WeaveInfo[i][:], sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveItemInfo[:])
					copy(endgameMsg.WeaveKind[i][:], sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].WeaveKind[:])
					copy(endgameMsg.IsWeave[i][:], sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].IsWeave[:])
					endgameMsg.EyeCard[i] = sp.ChiHuResult[i].ChiHuItemInfoArray[byMaxIndex].CardEye //如果为0表示没有牌眼
					//20200724如果是自摸，手上的5张的牌的泛全改成统2
					for j := byte(0); j < endgameMsg.WeaveCount[i] && j < 10; j++ {
						tempKind := endgameMsg.WeaveKind[i][j]
						if ZP_WIK_HUA == tempKind && endgameMsg.IsWeave[i][j] == 0 {
							if sp.ProvideUser == i {
								endgameMsg.WeaveKind[i][j] = ZP_WIK_TIANLONG
							} else {
								//20200724如果是滑胡的没有统数
								if sp.GameLogic.IsValidCard(endgameMsg.ChiHuCard) && sp.GameLogic.SwitchToCardIndexNoHua(endgameMsg.ChiHuCard) == sp.GameLogic.SwitchToCardIndexNoHua(endgameMsg.WeaveInfo[i][j][0]) && cardsNum <= 5 {
								} else {
									endgameMsg.WeaveKind[i][j] = ZP_WIK_TIANLONG
								}
							}
						}
					}
				}

				//分数计算
				iRealHuaShu := sp.GameLogic.GetHuaPaiNum(retcards[:])
				sp.ChiHuUserCount[i]++
				iPiaoFen := [TCGZ_MAX_PLAYER]int{}
				if sp.ProvideUser == i {
					sp.ZiMoUserCount[i]++
					endgameMsg.EndSubType = 1 //自摸
					endgameMsg.StrEnd = "自摸"
					sp.OnWriteGameRecord(i, "自摸")
					//自摸三家给分
					item := sp.GetPlayerByChair(i)
					for lose := uint16(0); int(lose) < sp.GetPlayerCount(); lose++ {
						if lose == i {
							continue
						}
						itemlose := sp.GetPlayerByChair(lose)
						iPiaoFen[lose] = item.Ctx.VecPiao.Num + itemlose.Ctx.VecPiao.Num
						iPiaoFen[i] += iPiaoFen[lose]
						endgameMsg.Score[lose] = sp.GameLogic.GetHuFen(iMaxHuXi, sp.GeziShu, sp.FenType, iPiaoFen[lose], sp.Base) + sp.Base
						endgameMsg.Score[i] += endgameMsg.Score[lose]
						endgameMsg.Score[lose] = endgameMsg.Score[lose] * (-1)
					}
				} else if sp.ProvideUser != static.INVALID_CHAIR && sp.ProvideUser != i {
					endgameMsg.EndSubType = 3
					sp.JiePaoUserCount[i]++
					sp.ProvideUserCount[sp.ProvideUser]++

					endgameMsg.StrEnd = "接炮"
					sp.OnWriteGameRecord(i, "接炮")
					//接炮输家给分
					item := sp.GetPlayerByChair(i)
					itemlose := sp.GetPlayerByChair(sp.ProvideUser)
					iPiaoFen[i] = item.Ctx.VecPiao.Num + itemlose.Ctx.VecPiao.Num
					iPiaoFen[sp.ProvideUser] += iPiaoFen[i]
					iBaseJiScore := sp.GameLogic.GetHuFen(iMaxHuXi, sp.GeziShu, sp.FenType, iPiaoFen[i], sp.Base)
					if sp.KeChong {
						//放炮玩家一个人出
						endgameMsg.Score[i] = iBaseJiScore + sp.Base
						endgameMsg.Score[sp.ProvideUser] -= (iBaseJiScore + sp.Base)
						if sp.DianPaoPei == 1 {
							//包另一家
							endgameMsg.Score[i] += iBaseJiScore
							endgameMsg.Score[sp.ProvideUser] -= iBaseJiScore
						}
					} else {
						//另一个人也出
						endgameMsg.Score[i] = iBaseJiScore + sp.Base
						endgameMsg.Score[sp.ProvideUser] -= (iBaseJiScore + sp.Base) //先只扣自己的
						lose2 := uint16(0)
						for ; int(lose2) < sp.GetPlayerCount(); lose2++ {
							if lose2 == i || lose2 == sp.ProvideUser {
								continue
							}
							break
						}
						//扣另一家
						if int(lose2) < sp.GetPlayerCount() {
							endgameMsg.Score[i] += iBaseJiScore
							endgameMsg.Score[lose2] -= iBaseJiScore
						}
					}
				}
				endgameMsg.HuaPaiShu[i] = iRealHuaShu

				endgameMsg.ProvideIndex = sp.GetProvidCardPos(endgameMsg.EndSubType, i, byte(byMaxIndex), sp.ProvideCard) //用于显示供牌的位置

				//记录总胡息，并发送给客户端
				sp.WeaveHuxi[i] = iMaxHuXi
				sp.SendHuxi(i)
				//总胡息记录到回放
				if len(sp.ReplayRecord.R_Orders) > 0 && sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].R_Opt == E_ZP_Hu {
					//总胡息记录到回放
					sp.ReplayRecord.R_Orders[len(sp.ReplayRecord.R_Orders)-1].AddReplayExtData(E_ZP_Ext_HuXi, sp.WeaveHuxi[i])
				}
			}
		}
	} else {
		//流局
		//流局处理
		endgameMsg.ChiHuCard = 0
		endgameMsg.ChiHuUserCount = 0
		endgameMsg.EndSubType = 2 //荒庄
		sp.HuangZhuang = true

		endgameMsg.StrEnd = "流局"

		sp.addReplayOrder(0, E_ZP_HuangZhuang, 0, []int{})
	}

	//保存基础积分
	endgameMsg.BaseScore = 0

	//拷贝四个玩家的扑克和分数
	for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
		endgameMsg.CardCount[i], endgameMsg.CardDataJZ[i] = sp.GameLogic.SwitchToCardData3(sp.CardIndex[i])
		for _, data := range sp.UserJianCards[i] {
			endgameMsg.UserJianCards[i] = append(endgameMsg.UserJianCards[i], int(data))
		}
		endgameMsg.UserGuanCardsCount[i] = sp.UserGuanCardsCount[i]
		//存放消息结构体的公共代码有
		for j := 0; j < 10; j++ {
			endgameMsg.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
			endgameMsg.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
			endgameMsg.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
			copy(endgameMsg.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])

			copy(endgameMsg.UserGuanCards[i][j][:], sp.UserGuanCards[i][j][:])
		}
		endgameMsg.TWeaveCount[i] = sp.WeaveItemCount[i]
	}
	//拷贝底牌
	endgameMsg.LeftCardCount = sp.LeftCardCount
	if sp.LeftCardCount > 0 {
		copy(endgameMsg.LeftCardData[:], sp.RepertoryCard[:])
	}

	endgameMsg.IsQuit = false

	endgameMsg.TheOrder = sp.CurCompleteCount

	//判断调整分
	for i := 0; i < sp.GetPlayerCount(); i++ {
		endgameMsg.Score[i] = endgameMsg.Score[i]
		endgameMsg.GameAdjustScore[i] = endgameMsg.Score[i]
		sp.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = sp.Total[i]
	}

	//发送信息
	//必须MAX_PLAYER人
	scoreOffset := [meta2.MAX_PLAYER]int{}
	for ip := 0; ip < meta2.MAX_PLAYER && ip < sp.GetPlayerCount(); ip++ {
		scoreOffset[ip] = endgameMsg.Score[ip]
	}
	_, endgameMsg.UserVitamin = sp.OnSettle(scoreOffset, consts.EventSettleGameOver)
	sp.VecGameEnd = append(sp.VecGameEnd, endgameMsg) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, endgameMsg)
	sp.SaveGameData()

	sp.OnWriteGameRecord(static.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	sp.UserAction = [TCGZ_MAX_PLAYER]int{} //bugid 7475，防止小结算断线重连回来还能看到上局没有处理的牌权
	//记录胡牌类型
	if int(nWinner) < sp.GetPlayerCount() {
		sp.ReplayRecord.R_WeaveCount = 0
		for byWcnt := byte(0); byWcnt < endgameMsg.WeaveCount[nWinner] && byWcnt < 10; byWcnt++ {
			if endgameMsg.IsWeave[nWinner][byWcnt] == 0 {
				//组合牌区的牌不发，对牌在写库时在放入
				copy(sp.ReplayRecord.R_WeaveItemInfo[sp.ReplayRecord.R_WeaveCount][:], endgameMsg.WeaveInfo[nWinner][byWcnt][:])
				sp.ReplayRecord.R_WeaveCount++
			}
		}
	}

	sp.ReplayRecord.R_ProvideUser = endgameMsg.ProvideUser
	sp.ReplayRecord.R_EndSubType = endgameMsg.EndSubType

	// 数据库写出牌记录	// 写完后清除数据
	sp.ReplayRecord.EndInfo = &endgameMsg
	sp.TableWriteOutDate()

	for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d", endgameMsg.Score[i], sp.Total[i])
		sp.OnWriteGameRecord(i, recrodStr)
	}

	for _, v := range sp.PlayerInfo { //i := 0; i < sp.GetPlayerCount(); i++ {
		wintype := static.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinOrLose[v.Seat] == 1 {
			wintype = static.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = static.ScoreKind_Lost //enScoreKind_Lost;
		}
		sp.TableWriteGameDate(int(sp.CurCompleteCount), v, wintype, endgameMsg.Score[v.Seat])
	}

	//扣房卡
	if sp.CurCompleteCount == 1 {
		sp.TableDeleteFangKa(sp.CurCompleteCount)
	}

	tempCount := int(sp.CurCompleteCount)

	//结束游戏
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu { //局数够了
		sp.CalculateResultTotal(static.GER_NORMAL, wChairID, 0) //计算总发送总结算

		//sp.UpdateOtherFriendDate(&endgameMsg, false)
		//通知框架结束游戏
		//sp.SetGameStatus(public.GS_MJ_FREE)
		sp.ConcludeGame()

	} else {

	}

	/////////////////////////////////////////////////////////////////////////////////////////
	// 1、第一局随机坐庄；2、一人胡牌胡牌玩家当庄，流局连庄；3、多人胡牌点炮这当庄
	if sp.ProvideUser == static.INVALID_CHAIR {
		sp.Nextbanker = sp.Banker
	} else if iWinnerCnt > 1 {
		sp.Nextbanker = sp.ProvideUser
	} else if nWinner != static.INVALID_CHAIR {
		if sp.DengZhuang {
			sp.Nextbanker = nWinner
		} else {
			sp.Nextbanker = uint16(sp.GetNextSeat(nWinner))
		}
	} else {
		//庄家赢了，还是这个庄家做庄
	}

	sp.OnGameEnd()

	//超时解散
	if !(tempCount >= sp.Rule.JuShu) {
		check := false
		if sp.Rule.Overtime_dismiss != -1 {
			for i := 0; i < sp.GetPlayerCount(); i++ {
				if item := sp.GetUserItemByChair(uint16(i)); item != nil {
					if sp.TuoGuanPlayer[i] {
						if check {
							var _msg = &static.Msg_C_DismissFriendResult{
								Id:   item.Uid,
								Flag: true,
							}
							sp.OnDismissResult(item.Uid, _msg)
							sp.OnWriteGameRecord(uint16(i), "托管中，系统帮你【同意】了解散")
						} else {
							check = true
							var msg = &static.Msg_C_DismissFriendReq{
								Id: item.Uid,
							}
							sp.SetDismissRoomTime(sp.Rule.Overtime_dismiss)
							sp.OnDismissFriendMsg(item.Uid, msg)
							sp.OnWriteGameRecord(uint16(i), "托管中，系统帮你【申请】了解散")
						}
					}
				}
			}
		}
		//if !check&&sp.Rule.Overtime_trust>0{
		if sp.RoundOverAutoReady {
			//自动准备由选项控制，跟托管没有关系了
			sp.SetAutoNextTimer(10) //自动开始下一局
		} else if sp.OutCardDismissTime > 0 {
			//不准备也要算超时时间，到时间解散，由于是解散，用ZP_AS_PLAYCARD没有关系
			// 超时超时时间
			iOvertime := sp.OperateTime
			sp.PowerStartTime = time.Now().Unix() //权限开始时间
			sp.GameState = ZP_AS_PLAYCARD         // 更新为要打牌状态
			sp.setLimitedTime(int64(iOvertime + 1))
		}
	}
	if tempCount >= sp.Rule.JuShu {
		sp.RepositTable(false) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	}
	sp.UserReady = [TCGZ_MAX_PLAYER]bool{}
	return true
}

func (sp *Sport_zp_jzhp) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	if int(wChairID) >= sp.GetPlayerCount() {
		return false
	}
	if cbReason != static.GER_USER_LEFT {
		return false
	}

	sp.GameEndStatus = static.GS_MJ_END

	//定义变量
	var endgameMsg static.Msg_S_ZP_TC_GameEnd
	endgameMsg.EndStatus = static.GER_USER_LEFT
	endgameMsg.TheOrder = sp.CurCompleteCount
	endgameMsg.EndSubType = 0

	//设置变量
	endgameMsg.ProvideUser = static.INVALID_CHAIR
	endgameMsg.BankUser = sp.Banker
	endgameMsg.IsQuit = true
	endgameMsg.WhoQuit = wChairID

	//拷贝四个玩家的扑克和分数
	for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
		endgameMsg.TotalScore[i] = sp.Total[i]

		endgameMsg.CardCount[i], endgameMsg.CardDataJZ[i] = sp.GameLogic.SwitchToCardData3(sp.CardIndex[i])
		for _, data := range sp.UserJianCards[i] {
			endgameMsg.UserJianCards[i] = append(endgameMsg.UserJianCards[i], int(data))
		}
		endgameMsg.UserGuanCardsCount[i] = sp.UserGuanCardsCount[i]
		//存放消息结构体的公共代码有
		for j := 0; j < 10; j++ {
			endgameMsg.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
			endgameMsg.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
			endgameMsg.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
			copy(endgameMsg.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])

			copy(endgameMsg.UserGuanCards[i][j][:], sp.UserGuanCards[i][j][:])
		}
		endgameMsg.TWeaveCount[i] = sp.WeaveItemCount[i]
	}
	//拷贝底牌
	endgameMsg.LeftCardCount = sp.LeftCardCount
	if sp.LeftCardCount > 0 {
		copy(endgameMsg.LeftCardData[:], sp.RepertoryCard[:])
	}

	//游戏记录
	sp.OnWriteGameRecord(wChairID, "强退游戏结束")

	//记录异常结束数据
	sp.addReplayOrder(wChairID, E_ZP_Li_Xian, 0, []int{})

	//写入游戏回放数据,写完重置当前回放数据
	sp.ReplayRecord.EndInfo = &endgameMsg
	sp.TableWriteOutDate()

	//将玩家分数发送给玩家
	for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d", endgameMsg.Score[i], sp.Total[i])
		sp.OnWriteGameRecord(i, recrodStr)
	}

	//必须MAX_PLAYER人
	scoreOffset := [meta2.MAX_PLAYER]int{}
	for ip := 0; ip < meta2.MAX_PLAYER && ip < sp.GetPlayerCount(); ip++ {
		scoreOffset[ip] = endgameMsg.Score[ip]
	}
	_, endgameMsg.UserVitamin = sp.OnSettle(scoreOffset, consts.EventSettleGameOver)

	//发送信息
	sp.VecGameEnd = append(sp.VecGameEnd, endgameMsg) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, endgameMsg)
	sp.SaveGameData()
	//数据库写分
	for _, v := range sp.PlayerInfo { //i := 0; i < sp.GetPlayerCount(); i++ {
		wintype := static.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinOrLose[v.Seat] == 1 {
			wintype = static.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = static.ScoreKind_Lost //enScoreKind_Lost;
		}
		if endgameMsg.EndStatus == static.GER_USER_LEFT {
			if v.Seat == wChairID {
				wintype = static.ScoreKind_Flee
			} else {
				wintype = static.ScoreKind_pass //逃跑活动分数在对战统计中忽略
			}
		}
		//iAward := sp.GetTaskAward(v.Seat)//金豆任务，先留着备用
		sp.TableWriteGameDate(int(sp.CurCompleteCount), v, wintype, endgameMsg.Score[v.Seat])
	}

	//sp.UpdateOtherFriendDate(&endgameMsg, true)
	//结束游戏
	sp.CalculateResultTotal(static.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	sp.CurCompleteCount = 0
	//sp.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()
	sp.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 解散，结束游戏
func (sp *Sport_zp_jzhp) OnGameEndDissmiss(wChairID uint16, cbReason byte) bool {
	if cbReason != static.GER_DISMISS {
		return false
	}
	if sp.GetGameStatus() == static.GS_MJ_FREE {
		//不要重复进解散流程
		return false
	}
	//变量定义
	var endgameMsg static.Msg_S_ZP_TC_GameEnd
	//在小结算期间申请解散的需要发送上次小结算数据
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		endgameMsg = gamend
		endgameMsg.EndStatus = static.GER_DISMISS
		sp.VecGameEnd[sp.CurCompleteCount-1].EndStatus = static.GER_DISMISS //修改到VecGameEnd里面去
		sp.ReWriteRec++                                                     //这种情况会>1，表示是在结算时申请解散的。
	} else {
		endgameMsg.EndStatus = static.GER_DISMISS
		endgameMsg.TheOrder = sp.CurCompleteCount
		endgameMsg.ProvideUser = static.INVALID_CHAIR
		endgameMsg.BankUser = sp.Banker
		copy(endgameMsg.TotalScore[:], sp.Total[:])

		//拷贝四个玩家的扑克和分数
		for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
			endgameMsg.CardCount[i], endgameMsg.CardDataJZ[i] = sp.GameLogic.SwitchToCardData3(sp.CardIndex[i])
			for _, data := range sp.UserJianCards[i] {
				endgameMsg.UserJianCards[i] = append(endgameMsg.UserJianCards[i], int(data))
			}
			endgameMsg.UserGuanCardsCount[i] = sp.UserGuanCardsCount[i]
			//存放消息结构体的公共代码有
			for j := 0; j < 10; j++ {
				endgameMsg.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
				endgameMsg.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
				endgameMsg.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
				copy(endgameMsg.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])

				copy(endgameMsg.UserGuanCards[i][j][:], sp.UserGuanCards[i][j][:])
			}
			endgameMsg.TWeaveCount[i] = sp.WeaveItemCount[i]
		}
		//拷贝底牌
		endgameMsg.LeftCardCount = sp.LeftCardCount
		if sp.LeftCardCount > 0 {
			copy(endgameMsg.LeftCardData[:], sp.RepertoryCard[:])
		}

		endgameMsg.IsQuit = true
		//记录异常结束数据
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			if _item == nil {
				continue
			}
			if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
				sp.addReplayOrder(uint16(i), E_ZP_Li_Xian, 0, []int{})
			}
		}

		sp.addReplayOrder(wChairID, E_ZP_Jie_san, 0, []int{})

		//写入游戏回放数据,写完重置当前回放数据
		sp.ReplayRecord.EndInfo = &endgameMsg
		sp.TableWriteOutDate()

		//必须MAX_PLAYER人
		scoreOffset := [meta2.MAX_PLAYER]int{}
		for ip := 0; ip < meta2.MAX_PLAYER && ip < sp.GetPlayerCount(); ip++ {
			scoreOffset[ip] = endgameMsg.Score[ip]
		}
		_, endgameMsg.UserVitamin = sp.OnSettle(scoreOffset, consts.EventSettleGameOver)

		//数据库写分
		for _, v := range sp.PlayerInfo {
			if v.Seat != static.INVALID_CHAIR {
				if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, endgameMsg.Score[v.Seat])
				} else {
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, endgameMsg.Score[v.Seat])
				}
			}
		}

		//发送信息
		sp.VecGameEnd = append(sp.VecGameEnd, endgameMsg) //保存，用于汇总计算
		sp.SaveGameData()
	}
	sp.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, endgameMsg)

	//游戏记录
	sp.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

	//sp.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	sp.CalculateResultTotal(static.GER_DISMISS, wChairID, 0)
	sp.GameEndStatus = static.GS_MJ_END
	//结束游戏
	//sp.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()

	sp.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

// 解散，结束游戏
func (sp *Sport_zp_jzhp) OnGameEndErrorDissmiss(wChairID uint16, cbReason byte) bool {
	if cbReason != static.GER_GAME_ERROR {
		return false
	}
	//变量定义
	var endgameMsg static.Msg_S_ZP_TC_GameEnd
	//在小结算期间申请解散的需要发送上次小结算数据
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		endgameMsg = gamend
		endgameMsg.EndStatus = static.GER_DISMISS
		sp.VecGameEnd[sp.CurCompleteCount-1].EndStatus = static.GER_DISMISS //修改到VecGameEnd里面去
		sp.ReWriteRec++                                                     //这种情况会>1，表示是在结算时申请解散的。
	} else {
		endgameMsg.EndStatus = static.GER_DISMISS
		endgameMsg.TheOrder = sp.CurCompleteCount
		endgameMsg.ProvideUser = static.INVALID_CHAIR
		endgameMsg.BankUser = sp.Banker
		copy(endgameMsg.TotalScore[:], sp.Total[:])

		//拷贝四个玩家的扑克和分数
		for i := uint16(0); int(i) < sp.GetPlayerCount(); i++ {
			endgameMsg.CardCount[i], endgameMsg.CardDataJZ[i] = sp.GameLogic.SwitchToCardData3(sp.CardIndex[i])
			for _, data := range sp.UserJianCards[i] {
				endgameMsg.UserJianCards[i] = append(endgameMsg.UserJianCards[i], int(data))
			}
			endgameMsg.UserGuanCardsCount[i] = sp.UserGuanCardsCount[i]
			//存放消息结构体的公共代码有
			for j := 0; j < 10; j++ {
				endgameMsg.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
				endgameMsg.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
				endgameMsg.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
				copy(endgameMsg.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])

				copy(endgameMsg.UserGuanCards[i][j][:], sp.UserGuanCards[i][j][:])
			}
			endgameMsg.TWeaveCount[i] = sp.WeaveItemCount[i]
		}
		//拷贝底牌
		endgameMsg.LeftCardCount = sp.LeftCardCount
		if sp.LeftCardCount > 0 {
			copy(endgameMsg.LeftCardData[:], sp.RepertoryCard[:])
		}

		endgameMsg.IsQuit = true
		//记录异常结束数据
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			if _item == nil {
				continue
			}
			if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
				sp.addReplayOrder(uint16(i), E_ZP_Li_Xian, 0, []int{})
			}
		}

		sp.addReplayOrder(wChairID, E_ZP_Jie_san, 0, []int{})

		//写入游戏回放数据,写完重置当前回放数据
		sp.ReplayRecord.EndInfo = &endgameMsg
		//sp.TableWriteOutDate()

		//必须MAX_PLAYER人
		scoreOffset := [meta2.MAX_PLAYER]int{}
		for ip := 0; ip < meta2.MAX_PLAYER && ip < sp.GetPlayerCount(); ip++ {
			scoreOffset[ip] = endgameMsg.Score[ip]
		}
		_, endgameMsg.UserVitamin = sp.OnSettle(scoreOffset, consts.EventSettleGameOver)

		////数据库写分
		//for _, v := range sp.PlayerInfo {
		//	if v.Seat != public.INVALID_CHAIR {
		//		if wChairID != public.INVALID_CHAIR && v.Seat == wChairID {
		//			sp.TableWriteGameDate(int(sp.CurCompleteCount), v, public.ScoreKind_pass, endgameMsg.Score[v.Seat])
		//		} else {
		//			sp.TableWriteGameDate(int(sp.CurCompleteCount), v, public.ScoreKind_pass, endgameMsg.Score[v.Seat])
		//		}
		//	}
		//}

		//发送信息
		sp.VecGameEnd = append(sp.VecGameEnd, endgameMsg) //保存，用于汇总计算
		sp.SaveGameData()
	}
	sp.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, endgameMsg)

	//游戏记录
	sp.OnWriteGameRecord(wChairID, "前面某个时刻程序出错过，需要排查错误日志，无法恢复这局游戏，解散游戏OnGameEndErrorDissmis")

	//sp.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	sp.CalculateResultTotal(static.GER_DISMISS, wChairID, 1)
	sp.GameEndStatus = static.GS_MJ_END
	//结束游戏
	//sp.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()

	sp.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}
func (sp *Sport_zp_jzhp) SaveGameData() {
	//变量定义
	var StatusPlay static.CMD_S_ZP_TC_StatusPlay
	//游戏变量
	StatusPlay.GameStatus = sp.GameState
	StatusPlay.TotalTime = sp.OperateTime
	StatusPlay.LeftTime = 0

	StatusPlay.BankerUser = sp.Banker       //庄家
	StatusPlay.CurrentUser = sp.CurrentUser //当前牌权玩家
	if StatusPlay.CurrentUser > uint16(sp.GetPlayerCount()) {
		StatusPlay.CurrentUser = sp.OutCardUser
	}
	StatusPlay.NoOutUser = sp.NoOutUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;

	StatusPlay.TheOrder = sp.CurCompleteCount

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = ZP_WIK_NULL //断线重连后要恢复之前的牌权

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData
	copy(StatusPlay.DiscardCard[:], sp.DiscardCard[:])
	copy(StatusPlay.DiscardCount[:], sp.DiscardCount[:])

	StatusPlay.ToTalOutCount = sp.OutCardCount //所有人的总出牌次数
	StatusPlay.IsOutFlag = 1                   //是否是出牌，发牌和出牌时，牌的边框颜色不一样
	if sp.SendStatus {
		StatusPlay.IsOutFlag = 0
	}

	//组合扑克
	//copy(StatusPlay.WeaveItemArray[:],sp.WeaveItemArray[:]);
	for i := 0; i < sp.GetPlayerCount(); i++ {
		for j := 0; j < 10; j++ {
			StatusPlay.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
			StatusPlay.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
			StatusPlay.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
			copy(StatusPlay.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])
		}
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecPiao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.UserPiaoReady
		StatusPlay.DispTongCnt[i] = sp.DispTongCnt[i]
		StatusPlay.TuoGuanPlayer[i] = sp.TuoGuanPlayer[i] //托管的有那些人
	}

	copy(StatusPlay.WeaveCount[:], sp.WeaveItemCount[:])
	copy(StatusPlay.TotalScore[:], sp.Total[:])
	//玩家的个人数据在gameend消息里面

	sp.VecGameData = append(sp.VecGameData, StatusPlay) //保存，用于汇总计算
}

// ! 初始化游戏
func (sp *Sport_zp_jzhp) OnInit(table base2.TableBase) {
	sp.KIND_ID = table.GetTableInfo().KindId
	sp.Config.StartMode = static.StartMode_FullReady
	sp.Config.PlayerCount = 3 //玩家人数
	sp.Config.ChairCount = 3  //椅子数量
	sp.PlayerInfo = make(map[int64]*components2.Player)

	sp.RepositTable(true)
	sp.SetGameStartMode(static.StartMode_FullReady)
	sp.GameTable = table
	sp.Init()
	sp.Unmarsha(table.GetTableInfo().GameInfo)
	if len(table.GetTableInfo().GameInfo) != 0 {
		if sp.LastTime > 0 {
			//fmt.Println(fmt.Sprintf("参数1（%d）参数2（%d） 参数3（%d）参数4（%d）参数5（%d）",sp.PlayTime,sp.PowerStartTime ,sp.LimitTime,sp.LastTime,time.Now().Unix()))
			sp.setLimitedTime(sp.LastTime)
		}
	}
	table.GetTableInfo().GameInfo = ""

	if 0 == sp.Rule.Overtime_applydiss {
		//设置自动解散时间2分钟
		sp.SetDismissRoomTime(120)
	} else {
		sp.SetDismissRoomTime(sp.Rule.Overtime_applydiss)
	}
	if 0 == sp.Rule.Overtime_offdiss {
		//设置离线解散时间15分钟
		sp.SetOfflineRoomTime(900)
	} else {
		sp.SetOfflineRoomTime(sp.Rule.Overtime_offdiss)
	}
	// 离线60s未准备踢出
	if len(table.GetTableInfo().Config.GameConfig) > 0 {
		var _msg rule2.FriendRuleZP_dyzp
		if err := json.Unmarshal(static.HF_Atobytes(table.GetTableInfo().Config.GameConfig), &_msg); err == nil {
			sp.Fleetime = _msg.Fleetime
			if sp.GetGameStatus() == static.GS_MJ_FREE && sp.Fleetime > 0 {
				sp.SetOfflineRoomTime(sp.Fleetime)
			}
			if _msg.LookonSupport == "" {
				sp.Config.LookonSupport = true
			} else {
				sp.Config.LookonSupport = _msg.LookonSupport == "true"
			}
		}
	}
	// 竞技点低于暂停下限 发起解散时间
	sp.SetVitaminLowPauseTime(10)
}

// ! 发送消息
func (sp *Sport_zp_jzhp) OnMsg(msg *base2.TableMsg) bool {
	wChairID := sp.GetChairByUid(msg.Uid)
	if wChairID == static.INVALID_CHAIR {
		sp.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::找不到玩家的座位号:%d", msg.Uid))
		return false
	}

	switch msg.Head {
	case consts.MsgTypeGameBalanceGameReq: //! 请求总结算信息 //暂时没有

		var _msg static.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.CalculateResultTotal_Rep(&_msg)
		} else {
			xlog.Logger().Debug(fmt.Sprintf("请求总结算消息错误：%v", err))
		}
	case consts.MsgTypeGameOutCard: //! 出牌消息
		var _msg static.Msg_C_ZP_OutCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if !sp.OnUserOutCard(&_msg, 0) {
				sp.DisconnectOnMisoperation(wChairID)
			}
		}
	case consts.MsgTypeGameTrustee: // 用户托管
		var _msg static.Msg_C_Trustee
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.onUserTustee(&_msg)
			// 详细日志
			LogStr := fmt.Sprintf("主动托管动作(true托管,false取消):TrustFlag=%t ", _msg.Trustee)
			sp.OnWriteGameRecord(sp.GetChairByUid(_msg.Id), LogStr)
		}
	case consts.MsgTypeGameOperateCard: //操作消息
		var _msg static.Msg_C_ZP_OperateCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			return sp.OnUserOperateCard(&_msg)
		}
	case consts.MsgTypeGameGoOnNextGame: //下一局 //暂时没有下一局
		//详细日志
		LogStr := string("OnUserClientNextGame!!! ")
		sp.OnWriteGameRecord(wChairID, LogStr)

		var _msg static.Msg_C_GoOnNextGame
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.OnUserClientNextGame(&_msg)
		}
	case consts.MsgTypeGameXuanpiao: //选漂
		var _msg static.Msg_C_Piao
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			return sp.OnUserClientXuanpiao(&_msg)
		}
	case consts.MsgTypeNoOut: //捏牌
		var _msg static.Msg_C_ZP_NoOut
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			return sp.OnUserNoOut(&_msg)
		}
	case consts.MsgTypeGameChangeSeat: //申请换座
		if sp.GameEndStatus == byte(static.GS_FREE) {
			//若申请换座玩家已准备
			item := sp.GetUserItemByChair(uint16(wChairID))
			//if item !=nil && item.Ready{
			//	sp.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::准备状态下无法换座:%d", msg.Uid))
			//	sp.SendGameNotificationMessage(wChairID,fmt.Sprintf("准备状态下无法换座"))
			//	break
			//}
			var _msg static.Msg_C_CHANGESEAT
			if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
				if _msg.ChairID >= 0 && int(_msg.ChairID) < sp.GetPlayerCount() {
					_item := sp.GetUserItemByChair(uint16(_msg.ChairID))
					if _msg.ChairID < 0 || _msg.ChairID >= static.MAX_PLAYER_4P {
						break
					}
					//若要换的座上有人
					if _item != nil {
						sp.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::该座位%d上已有玩家%d存在", _msg.ChairID, _item.Uid))
						sp.SendGameNotificationMessage(wChairID, fmt.Sprintf("该座位上已有玩家存在"))
						break
					}
					//调换座位
					sp.ReSeat(item.Uid, int(_msg.ChairID))
				} else {
					sp.OnWriteGameRecord(wChairID, fmt.Sprintf("换座客户端传过来的椅子编号不合法[%d]", _msg.ChairID))
				}
			}
		}
	default:
		//sp.Common.OnMsg(msg)
	}
	return true
}
func (sp *Sport_zp_jzhp) OnUserNoOut(msg *static.Msg_C_ZP_NoOut) bool {
	_userItem := sp.GetUserItemByChair(msg.ChairID)
	if _userItem == nil {
		return false
	}
	if sp.NoOutUser != static.INVALID_CHAIR {
		sp.OnWriteGameRecord(msg.ChairID, "已经有人先选择了捏牌")
		return false
	}
	sp.OnWriteGameRecord(msg.ChairID, "捏牌动作")
	if int(msg.ChairID) < sp.GetPlayerCount() {
		sp.NoOutUser = msg.ChairID
		//记录捏牌
		sp.addReplayOrder(msg.ChairID, E_ZP_NoOut, 0, []int{})
		var nooutmsg static.Msg_S_ZP_NoOut
		nooutmsg.ChairID = msg.ChairID
		nooutmsg.Status = msg.Status
		sp.SendTableMsg(consts.MsgTypeNoOut, nooutmsg)
		//发送旁观数据
		sp.SendTableLookonMsg(consts.MsgTypeNoOut, nooutmsg)

		if sp.GameState == ZP_AS_PLAYCARD {
			if sp.CurrentUser == msg.ChairID {
				//当前正要出牌
				sp.CurrentUser = uint16(sp.GetNextSeat(sp.CurrentUser))
				sp.StartHD()
				sp.GameState = ZP_AS_SENDCARD
				sp.setLimitedTime(1) //给下家发牌，需要1s后在发下一张牌
			} else {
				//等待响应
				if sp.UserAction[msg.ChairID] != ZP_WIK_NULL && !sp.Response[msg.ChairID] {
					//弃
					var _msg static.Msg_C_ZP_OperateCard
					_msg.Code = ZP_WIK_NULL
					_msg.Card = sp.ProvideCard
					_msg.Id = sp.GetUidByChair(msg.ChairID)
					sp.OnUserOperateCard(&_msg)
				}
			}
		}
	}
	return true
}

// 下一局
func (sp *Sport_zp_jzhp) OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu || sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()

	nChiarID := sp.GetChairByUid(msg.Id)
	if nChiarID >= 0 && nChiarID < uint16(sp.GetPlayerCount()) {
		_item := sp.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
			sp.UserReady[nChiarID] = true
			_item.Ready = true
		}
	}
	//将该消息广播出去。游戏开始后，不用广播
	if sp.GameEndStatus != static.GS_MJ_PLAY {
		sp.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
		//发送旁观数据
		sp.SendTableLookonMsg(consts.MsgTypeGameGoOnNextGame, msg)
		sp.SendUserStatus(int(nChiarID), static.US_READY) //把我的状态发给其他人
	}

	//SEND_TABLE_DATA(INVALID_CHAIR,SUB_C_GOON_NEXT_GAME,pDataBuffer);

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < sp.GetPlayerCount(); i++ {
		item := sp.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			break
		}
		if i == sp.GetPlayerCount()-1 {
			sp.RepositTable(false) // 复位桌子
			sp.CurCompleteCount++
			sp.GetTable().SetBegin(true)
			sp.OnGameStart()
		}
	}
	return true
}

// 托管
func (sp *Sport_zp_jzhp) onUserTustee(msg *static.Msg_C_Trustee) bool {
	// 变量定义
	var tuoguan static.Msg_S_ZP_Trustee
	tuoguan.ChairID = sp.GetChairByUid(msg.Id)
	tuoguan.Trustee = msg.Trustee
	// 校验规则
	if tuoguan.ChairID < TCGZ_MAX_PLAYER && (sp.GameState == ZP_AS_PLAYCARD || sp.GameState == ZP_AS_SENDCARD || sp.GameState == ZP_AS_GUANSHENG) {
		if tuoguan.Trustee == true {
			sp.TuoGuanPlayer[tuoguan.ChairID] = true
			sp.TrustCounts[tuoguan.ChairID]++

			if tuoguan.ChairID == sp.CurrentUser {
				//sp.DownTime = GetCPUTickCount()+sp.AutoOutTime;
				if int64(sp.LimitTime) > time.Now().Unix() {
					tuoguan.Overtime = sp.LimitTime - time.Now().Unix()
				}
				if time.Now().Unix()+int64(sp.AutoOutTime) < sp.LimitTime { // 如果只剩下托管出牌的时间了，就不重新算了，否则跟改为托管出牌的时间
					sp.setLimitedTime(int64(sp.AutoOutTime))
				}
			}
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, tuoguan)
			// 回放托管记录
			sp.addReplayOrder(tuoguan.ChairID, E_ZP_Tuo_Guan, 1, []int{})
		} else if tuoguan.Trustee == false {
			sp.TuoGuanPlayer[tuoguan.ChairID] = false
			// 如果是当前的玩家，那么重新设置一下开始时间//sp.GameState == ZP_AS_SENDCARD阶段永远都是1秒出牌
			if tuoguan.ChairID == sp.CurrentUser && sp.GameState != ZP_AS_SENDCARD {
				//sp.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
				if time.Now().Unix() < sp.LimitTime { // 如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < sp.LimitTime
					sp.setLimitedTime(sp.OperateTime + sp.PowerStartTime - time.Now().Unix() + 1)
					tuoguan.Overtime = sp.OperateTime + sp.PowerStartTime - time.Now().Unix()
				}
			}

			//tuoguan.theTime = PlayTime-(now-nowTime);
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, tuoguan)
			//回放增加托管记录
			sp.addReplayOrder(tuoguan.ChairID, E_ZP_Tuo_Guan, 0, []int{})
		} else {
			return false
		}
	} else {
		// 详细日志
		LogStr := string("托管动作:游戏状态不正确 ")
		sp.OnWriteGameRecord(tuoguan.ChairID, LogStr)
		return false
	}

	return true
}

// ! 给客户端发送总结算
func (sp *Sport_zp_jzhp) CalculateResultTotal_Rep(msg *static.Msg_C_BalanceGameEeq) {
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = sp.Rule.FangZhuID
	balanceGame.CurTotalCount = sp.CurCompleteCount //总盘数

	/*为什么要特意判断下0？之前没有这个协议，是新需求要加的之前很多数据都在RepositTableFrameSink里面重置了，但是由于新功能的开发
	部分数据放到游戏开始的时候重置了，但是存在的问题是，如果一个桌子解散了，这个桌子在服务器内部还是存在的，他的部分数据也是存在的
	当某个玩家进来，还没有点开始的时候点战绩，这时候存在把本包厢桌子上次数据发送过来的情况，但是改动m_MaxFanUserCount之类变量的值又
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
				balanceGame.GameScore[j] += sp.VecGameEnd[i].Score[j] //总分
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

func (sp *Sport_zp_jzhp) sendGameSceneStatusPlay(player *components2.Player) bool {

	if player.LookonTableId > 0 {
		sp.sendGameSceneStatusPlayLookon(player)
		return true
	}

	wChiarID := player.GetChairID()

	if int(wChiarID) >= sp.GetPlayerCount() {
		xlog.Logger().Info("sendGameSceneStatusPlay invalid chair")
		return false
	}

	//变量定义
	var StatusPlay static.CMD_S_ZP_TC_StatusPlay
	//游戏变量
	StatusPlay.GameStatus = sp.GameState
	StatusPlay.TotalTime = sp.OperateTime
	StatusPlay.LeftTime = 0
	if time.Now().Unix()+1 < sp.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.LeftTime = sp.LimitTime - time.Now().Unix()
	}
	StatusPlay.BankerUser = sp.Banker       //庄家
	StatusPlay.CurrentUser = sp.CurrentUser //当前牌权玩家
	if StatusPlay.CurrentUser > uint16(sp.GetPlayerCount()) {
		StatusPlay.CurrentUser = sp.OutCardUser
		if StatusPlay.CurrentUser > uint16(sp.GetPlayerCount()) {
			StatusPlay.CurrentUser = sp.ResumeUser
		}
	}
	StatusPlay.NoOutUser = sp.NoOutUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;

	StatusPlay.TheOrder = sp.CurCompleteCount

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = ZP_WIK_NULL //断线重连后要恢复之前的牌权
	if sp.Response[wChiarID] == false {
		StatusPlay.ActionMask = sp.UserAction[wChiarID]
		if sp.UserAction[wChiarID] != ZP_WIK_NULL {
			sp.OnWriteGameRecord(wChiarID, fmt.Sprintf("重新发送了牌权：0x%x , %s", sp.UserAction[wChiarID], sp.GetWikStr(sp.UserAction[wChiarID])))
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData
	copy(StatusPlay.DiscardCard[:], sp.DiscardCard[:])
	copy(StatusPlay.DiscardCount[:], sp.DiscardCount[:])

	StatusPlay.ToTalOutCount = sp.OutCardCount //所有人的总出牌次数
	StatusPlay.IsOutFlag = 1                   //是否是出牌，发牌和出牌时，牌的边框颜色不一样
	if sp.SendStatus {
		StatusPlay.IsOutFlag = 0
	}

	//组合扑克
	//copy(StatusPlay.WeaveItemArray[:],sp.WeaveItemArray[:]);
	for i := 0; i < sp.GetPlayerCount(); i++ {
		for j := 0; j < 10; j++ {
			StatusPlay.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
			StatusPlay.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
			StatusPlay.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
			copy(StatusPlay.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])
		}
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecPiao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.UserPiaoReady
		StatusPlay.DispTongCnt[i] = sp.DispTongCnt[i]
		StatusPlay.TuoGuanPlayer[i] = sp.TuoGuanPlayer[i] //托管的有那些人
	}

	copy(StatusPlay.WeaveCount[:], sp.WeaveItemCount[:])
	copy(StatusPlay.TotalScore[:], sp.Total[:])

	StatusPlay.UserGuanCardsCount = sp.UserGuanCardsCount[wChiarID]
	for j := 0; j < 10; j++ {
		for k := 0; k < 5; k++ {
			StatusPlay.UserGuanCards[j] = append(StatusPlay.UserGuanCards[j], int(sp.UserGuanCards[wChiarID][j][k]))
		}
	}
	for j := 0; j < len(sp.UserJianCards[wChiarID]); j++ {
		StatusPlay.UserJianCards = append(StatusPlay.UserJianCards, int(sp.UserJianCards[wChiarID][j]))
	}
	for j := 0; j < len(sp.UserJianCardsCur[wChiarID]); j++ {
		StatusPlay.UserJianCardsCur = append(StatusPlay.UserJianCardsCur, int(sp.UserJianCardsCur[wChiarID][j]))
	}

	for j := 0; j < TCGZ_MAX_INDEX; j++ {
		StatusPlay.MyTongCnt[j] = sp.TongInfo[wChiarID].CardTongInfo[j].TongCnt
	}

	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardDataJZ = sp.GameLogic.SwitchToCardData3(sp.CardIndex[wChiarID])

	//发送场景
	sp.SendPersonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, wChiarID)

	//断线回来后记录下手牌数据
	//游戏记录

	szGameRecord := fmt.Sprintf("牌型：%s,断线重连记录的手牌数据", sp.GameLogic.SwitchToCardName1(sp.CardIndex[wChiarID][:]))
	sp.OnWriteGameRecord(wChiarID, szGameRecord)

	if int(wChiarID) < sp.GetPlayerCount() && sp.CurrentUser == wChiarID {
		isOutting := true
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if sp.UserAction[wChiarID] != ZP_WIK_NULL {
				isOutting = false
				break
			}
		}
		//正在出牌时要发送出牌的牌权
		if isOutting {
			var s_Power static.Msg_S_ZP_SendPower
			s_Power.CurrentUser = wChiarID
			s_Power.LeftTime = int(StatusPlay.LeftTime)
			sp.SendTableMsg(consts.MsgTypeSendPower, s_Power)
			sp.CurrentUser = wChiarID

			sp.OnWriteGameRecord(wChiarID, "拥有出牌牌权")
		} else {
			//查听
			tingcnt, tingresult := sp.CheckTing(wChiarID)
			if tingcnt > 0 {
				sp.SendTingInfo(wChiarID, tingresult)
			}
		}
	} else {
		//查听
		tingcnt, tingresult := sp.CheckTing(wChiarID)
		if tingcnt > 0 {
			sp.SendTingInfo(wChiarID, tingresult)
		}
	}
	//重新计算总胡息
	//WeaveHuxi[wChiarID] = GameLogic.CalculateWeaveHuXi(WeaveItemArray[wChiarID],WeaveItemCount[wChiarID]);
	sp.SendHuxi(wChiarID)

	//发小结消息
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 && int(wChiarID) < sp.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		gamend.ReLinkFlag = 1 //表示为断线重连

		sp.SendPersonMsg(consts.MsgTypeGameEnd, gamend, wChiarID)
		for i := 0; i < TCGZ_MAX_PLAYER && i < sp.GetPlayerCount(); i++ {
			if sp.UserReady[i] {
				sp.OnWriteGameRecord(wChiarID, "已经准备的人都要重新补发准备")
				var _msg static.Msg_C_GoOnNextGame
				_msg.Id = sp.GetUidByChair(uint16(i))
				sp.OnUserClientNextGame(&_msg)
			}
		}
	}

	//发送解散房间所有玩家的反应
	sp.SendAllPlayerDissmissInfo(player)

	return true
}

func (sp *Sport_zp_jzhp) sendGameSceneStatusPlayLookon(player *components2.Player) bool {

	if player.LookonTableId == 0 {
		return false
	}
	wChiarID := player.GetChairID()
	if int(wChiarID) >= sp.GetPlayerCount() {
		wChiarID = 0
	}
	//是否要获取wChiarID位置真正玩家的信息 ？
	//playerOnChair := sp.GetUserItemByChair(wChiarID)

	//变量定义
	var StatusPlay static.CMD_S_ZP_TC_StatusPlay
	//游戏变量
	StatusPlay.GameStatus = sp.GameState
	StatusPlay.TotalTime = sp.OperateTime
	StatusPlay.LeftTime = 0
	if time.Now().Unix()+1 < sp.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.LeftTime = sp.LimitTime - time.Now().Unix()
	}
	StatusPlay.BankerUser = sp.Banker       //庄家
	StatusPlay.CurrentUser = sp.CurrentUser //当前牌权玩家
	if StatusPlay.CurrentUser > uint16(sp.GetPlayerCount()) {
		StatusPlay.CurrentUser = sp.OutCardUser
		if StatusPlay.CurrentUser > uint16(sp.GetPlayerCount()) {
			StatusPlay.CurrentUser = sp.ResumeUser
		}
	}
	StatusPlay.NoOutUser = sp.NoOutUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;

	StatusPlay.TheOrder = sp.CurCompleteCount

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = ZP_WIK_NULL //断线重连后要恢复之前的牌权
	//if (sp.Response[wChiarID]==false){
	//	StatusPlay.ActionMask= sp.UserAction[wChiarID]
	//	if sp.UserAction[wChiarID] != ZP_WIK_NULL{
	//		sp.OnWriteGameRecord(wChiarID,fmt.Sprintf("重新发送了牌权：0x%x , %s", sp.UserAction[wChiarID],sp.GetWikStr(sp.UserAction[wChiarID])));
	//	}
	//}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData
	copy(StatusPlay.DiscardCard[:], sp.DiscardCard[:])
	copy(StatusPlay.DiscardCount[:], sp.DiscardCount[:])

	StatusPlay.ToTalOutCount = sp.OutCardCount //所有人的总出牌次数
	StatusPlay.IsOutFlag = 1                   //是否是出牌，发牌和出牌时，牌的边框颜色不一样
	if sp.SendStatus {
		StatusPlay.IsOutFlag = 0
	}

	//组合扑克
	//copy(StatusPlay.WeaveItemArray[:],sp.WeaveItemArray[:]);
	for i := 0; i < sp.GetPlayerCount(); i++ {
		for j := 0; j < 10; j++ {
			StatusPlay.WeaveItemArray[i][j].WeaveKind = sp.WeaveItemArray[i][j].WeaveKind
			StatusPlay.WeaveItemArray[i][j].ProvideUser = sp.WeaveItemArray[i][j].ProvideUser
			StatusPlay.WeaveItemArray[i][j].PublicCard = sp.WeaveItemArray[i][j].PublicCard
			copy(StatusPlay.WeaveItemArray[i][j].Cards[:], sp.WeaveItemArray[i][j].Cards[:])
		}
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecPiao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.UserPiaoReady
		StatusPlay.DispTongCnt[i] = sp.DispTongCnt[i]
		StatusPlay.TuoGuanPlayer[i] = sp.TuoGuanPlayer[i] //托管的有那些人
	}

	copy(StatusPlay.WeaveCount[:], sp.WeaveItemCount[:])
	copy(StatusPlay.TotalScore[:], sp.Total[:])

	StatusPlay.UserGuanCardsCount = sp.UserGuanCardsCount[wChiarID]
	for j := 0; j < 10; j++ {
		for k := 0; k < 5; k++ {
			StatusPlay.UserGuanCards[j] = append(StatusPlay.UserGuanCards[j], int(sp.UserGuanCards[wChiarID][j][k]))
		}
	}
	for j := 0; j < len(sp.UserJianCards[wChiarID]); j++ {
		StatusPlay.UserJianCards = append(StatusPlay.UserJianCards, int(sp.UserJianCards[wChiarID][j]))
	}
	for j := 0; j < len(sp.UserJianCardsCur[wChiarID]); j++ {
		StatusPlay.UserJianCardsCur = append(StatusPlay.UserJianCardsCur, int(sp.UserJianCardsCur[wChiarID][j]))
	}

	for j := 0; j < TCGZ_MAX_INDEX; j++ {
		StatusPlay.MyTongCnt[j] = sp.TongInfo[wChiarID].CardTongInfo[j].TongCnt
	}

	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardDataJZ = sp.GameLogic.SwitchToCardData3(sp.CardIndex[wChiarID])

	//发送旁观数据
	sp.SendPersonLookonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, player.Uid)

	if int(wChiarID) < sp.GetPlayerCount() && sp.CurrentUser == wChiarID {
		isOutting := true
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if sp.UserAction[wChiarID] != ZP_WIK_NULL {
				isOutting = false
				break
			}
		}
		//正在出牌时要发送出牌的牌权
		if isOutting {
			var s_Power static.Msg_S_ZP_SendPower
			s_Power.CurrentUser = wChiarID
			s_Power.LeftTime = int(StatusPlay.LeftTime)
			//发送旁观数据
			sp.SendPersonLookonMsg(consts.MsgTypeSendPower, s_Power, player.Uid)

		} else {
			//	//查听
			//	tingcnt,tingresult:=sp.CheckTing(wChiarID)
			//	if tingcnt > 0{
			//		sp.SendTingInfo(wChiarID,tingresult)
			//	}
		}
	} else {
		////查听
		//tingcnt,tingresult:=sp.CheckTing(wChiarID)
		//if tingcnt > 0{
		//	sp.SendTingInfo(wChiarID,tingresult)
		//}
	}
	//重新计算总胡息
	//WeaveHuxi[wChiarID] = GameLogic.CalculateWeaveHuXi(WeaveItemArray[wChiarID],WeaveItemCount[wChiarID]);
	sp.SendHuxi(wChiarID)

	//发小结消息
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 && int(wChiarID) < sp.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		gamend.ReLinkFlag = 1 //表示为断线重连

		//发送旁观数据
		sp.SendPersonLookonMsg(consts.MsgTypeGameEnd, gamend, player.Uid)
		//for i:=0 ;i<TCGZ_MAX_PLAYER && i<sp.GetPlayerCount();i++ {
		//	if sp.UserReady[i] {
		//		sp.OnWriteGameRecord(wChiarID,"已经准备的人都要重新补发准备");
		//		var _msg public.Msg_C_GoOnNextGame
		//		_msg.Id = sp.GetUidByChair(uint16(i))
		//		sp.OnUserClientNextGame(&_msg)
		//	}
		//}
	}

	//发送解散房间所有玩家的反应
	sp.SendAllPlayerDissmissInfo(player)

	return true
}

// 游戏场景消息发送
func (sp *Sport_zp_jzhp) SendGameScene(uid int64, status byte, secret bool) {
	player := sp.GetUserItemByUid(uid)
	if player == nil {
		//不是游戏玩家就是旁观玩家
		player = sp.GetLookonUserItemByUid(uid)
		if player == nil {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "SendGameScene 发送游戏场景，玩家空指针")
			return
		}
	}
	switch status {
	case static.GS_MJ_FREE:
		sp.sendGameSceneStatusFree(player)
	case static.GS_MJ_PLAY:
		sp.sendGameSceneStatusPlay(player)
	case static.GS_MJ_END:
		sp.sendGameSceneStatusPlay(player)
	}
}
func (sp *Sport_zp_jzhp) sendGameSceneStatusFree(player *components2.Player) bool {

	//变量定义
	var StatusFree static.Msg_S_ZP_StatusFree
	//构造数据
	StatusFree.BankerUser = sp.Banker
	StatusFree.CellScore = sp.GetCellScore()    //sp.m_pGameServiceOption->lCellScore;
	StatusFree.CellMinScore = sp.GetCellScore() //最低分
	StatusFree.CellMaxScore = sp.GetCellScore() //最低分

	//发送场景
	//	sp.SendPersonMsg(constant.MsgTypeGameStatusFree, StatusFree, PlayerInfo.GetChairID())
	sp.SendUserMsg(consts.MsgTypeGameStatusFree, StatusFree, player.Uid)

	return true
}

// 计算总发送总结算
func (sp *Sport_zp_jzhp) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	sp.TimeEnd = time.Now().Unix() //大局结束时间
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_ZP_BALANCE_GAME
	balanceGame.Userid = sp.Rule.FangZhuID
	balanceGame.CurTotalCount = sp.CurCompleteCount //总盘数
	balanceGame.TimeStart = sp.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = sp.TimeEnd

	for i := 0; i < len(sp.VecGameEnd); i++ {
		for j := 0; j < sp.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += sp.VecGameEnd[i].Score[j] //总分
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

	// 如果是正常结束
	//if (GER_NORMAL == cbReason)
	{
		// 有大赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		//		iChairID := 0
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < sp.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				//				iChairID = j
				iMaxScoreCount++
			}
		}
		if iMaxScoreCount == 1 && sp.Rule.CreateType == 3 { // 大赢家支付
			//IServerUserItem * pIServerUserItem = m_pITableFrame->GetServerUserItem(iChairID);
			//DWORD userid = pIServerUserItem->GetUserID();
			//				m_pITableFrame->TableDeleteDaYingJiaFangKa(userid);
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
		if sp.ReWriteRec <= 1 {
			wintype = static.ScoreKind_pass
		}
	} else {
		if sp.CurCompleteCount == 0 { //有可能第一局还没有开始，就解散了（比如在吓跑的过程中解散）
			wintype = static.ScoreKind_pass
		}
	}

	if cbSubReason == 0 {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_userItem := sp.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.ChiHuUserCount[i] = sp.ChiHuUserCount[i]
			balanceGame.ProvideUserCount[i] = sp.ProvideUserCount[i]
			balanceGame.JiePaoUserCount[i] = sp.JiePaoUserCount[i]
			balanceGame.ZiMoUserCount[i] = sp.ZiMoUserCount[i]
			balanceGame.MaxHuxi[i] = sp.MaxHUxi[i]

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
		if len(sp.VecGameData) > 0 {
			gamedataStr = static.HF_JtoA(sp.VecGameData[len(sp.VecGameData)-1])
		}
		sp.SaveLastGameinfo(_userItem.Uid, gameendStr, static.HF_JtoA(balanceGame), gamedataStr)
	}

	// 记录用户好友房历史战绩
	if wintype != static.ScoreKind_pass {
		sp.TableWriteHistoryRecordWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
		sp.TableWriteHistoryRecordDetailWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
	}

	//发消息
	sp.SendTableMsg(consts.MsgTypeGameBalanceGame, balanceGame)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameBalanceGame, balanceGame)
	sp.resetEndDate()
}

func (sp *Sport_zp_jzhp) resetEndDate() {
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_ZP_TC_GameEnd{}
	sp.VecGameData = []static.CMD_S_ZP_TC_StatusPlay{}

	for _, v := range sp.PlayerInfo {
		v.OnEnd()
	}
}

// ! 写入游戏回放数据
func (sp *Sport_zp_jzhp) TableWriteOutDate() {

	if sp.ReWriteRec != 0 {
		sp.ReWriteRec++ //这种情况会>1，表示是在结算时申请解散的。
		// 写完后清除数据
		sp.ReplayRecord.ReSet()
		return
	}

	recordReplay := new(models.RecordGameReplay)
	recordReplay.GameNum = sp.GetTableInfo().GameNum
	recordReplay.RoomNum = sp.GetTableInfo().Id
	recordReplay.PlayNum = int(sp.CurCompleteCount)
	recordReplay.ServerId = server2.GetServer().Con.Id
	recordReplay.HandCard = sp.GameLogic.GetWriteHandReplayRecordString(sp.ReplayRecord)
	recordReplay.OutCard = sp.GameLogic.GetWriteOutReplayRecordString(sp.ReplayRecord)
	recordReplay.UVitaminMap = sp.ReplayRecord.UVitamin
	recordReplay.KindID = sp.GetTableInfo().KindId
	recordReplay.CardsNum = 0

	if sp.ReplayRecord.EndInfo != nil {
		recordReplay.EndInfo = static.HF_JtoA(sp.ReplayRecord.EndInfo)
	}

	if id, err := server2.GetDBMgr().InsertGameRecordReplay(recordReplay); err != nil {
		xlog.Logger().Debug(fmt.Sprintf("%d,写游戏出牌记录：（%v）出错（%v）", id, recordReplay, err))
	}

	sp.RoundReplayId = recordReplay.Id

	// 写完后清除数据
	sp.ReplayRecord.ReSet()

	sp.ReWriteRec++ //在小结算过程中解散不在写回放记录了
}

// 检查是否所有的牌权都已经释放
func (sp *Sport_zp_jzhp) CheckRightAllReleased() bool {
	AllReleased := true
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if sp.UserAction[i] != ZP_WIK_NULL {
			AllReleased = false
			break
		}
	}

	return AllReleased
}

// ! 游戏退出
func (sp *Sport_zp_jzhp) OnExit(uid int64) {
	sp.Common.OnExit(uid)
}

func (sp *Sport_zp_jzhp) OnTime() {
	sp.Common.OnTime()
}

// 场景保存
func (sp *Sport_zp_jzhp) Tojson() string {
	sp.LastTime = sp.LimitTime - time.Now().Unix() + 1
	var _json SportZPTCJsonSerializer
	return _json.ToJson_ZP_TCGZ(sp)
}

// 场景恢复
func (sp *Sport_zp_jzhp) Unmarsha(data string) {
	var _json SportZPTCJsonSerializer

	_json.Unmarsha_ZP_TCGZ(data, sp)
}
