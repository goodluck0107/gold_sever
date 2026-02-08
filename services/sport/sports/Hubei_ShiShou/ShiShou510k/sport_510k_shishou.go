package ShiShou510k

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
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"math/rand"
	"strings"
	"time"
)

/*
石首510k纸牌4人好友房
*/
type FriendRule_ss510k struct {
	Difen            int    `json:"difen"`            //底分(级数分)
	Radix            int    `json:"scoreradix"`       //底分基数
	SerPay           int    `json:"revenue"`          //茶水
	Fa               int    `json:"fa"`               //没逃跑处罚倍数
	Jiang            int    `json:"jiang"`            //逃跑奖励别人的倍数
	ShareScoreType   int    `json:"sharescoretype"`   //单双下贡献分类型 (0:100/200  1:200/400)
	NineSecondRoom   string `json:"sc9"`              //九秒场
	IpLimit          string `json:"ip"`               //ipgps限制
	Overtime_trust   int    `json:"overtime_trust"`   //超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //超时解散
	PlayMode         int    `json:"playmode"`         //玩法(0经典  1双赖   2四赖)
	RandTeamer       string `json:"randomTeam"`       //随机队友
	FleeTime         int    `json:"fleetime"`         // 客户端传来的 游戏开始前离线踢人时间
	AddDiFen         int    `json:"adddifen"`         //额外加的底分
	TrustPunish      string `json:"tgff"`             //托管罚分开关
	Dissmiss         int    `json:"dissmiss"`         //解散次数,0不限制,12345对应限制次数

}

/*
单双下(12游为胜利一方的时候为双下,13游为胜利一方的时候为单下)：
	选择了100-200，双下输家上拱200分，单下输家上供100分;
	选择了200-400，双下输家上拱400分，单下输家上供100分;
*/

//玩家分数
type Msg_S_DG_PlayScore struct {
	//MsgTypePlayScore
	static.Msg_S_DG_PlayScore
	BombScore        int    `json:"bombscore"`        //本轮炸弹分
	PlayerScore      [4]int `json:"playerscore"`      //用户当前得分
	PlayerTotalScore [4]int `json:"playertotalscore"` //用户总得分
}

//游戏状态
type CMD_S_DG_StatusPlay struct {
	//MsgTypeGameStatusPlay
	static.CMD_S_DG_StatusPlay
	PlayerScore      [4]int `json:"playerscore"`      //用户当前得分
	PlayerTotalScore [4]int `json:"playertotalscore"` //用户总得分
}

// 服务器发给客户端 总结算
type Msg_S_DG_BALANCE_GAME struct {
	static.Msg_S_DG_BALANCE_GAME
	TheBank    uint16 `json:"thebank"`   //庄家
	TheParter  uint16 `json:"theparter"` //庄家朋友
	JiFen      int    `json:"jifen"`
	JiCount    [4]int `json:"jicount"`    //级数
	TotalScore [4]int `json:"totalscore"` //总积分
	LostCount  [4]int `json:"lostcount"`  //败局次数
	GongFen    [4]int `json:"gongfen"`    //贡献分
	FaScore    [4]int `json:"fascore"`    //炸弹分
	GetScore   [4]int `json:"getscore"`   //抓分
}

//换座消息
type Msg_C_DG_CHANGESEAT struct {
	//MsgTypeGameOutCard
	ChairID byte `json:"chairid"` //换座位号
}

const (
	YOUTYPE_12 = iota + 1 //一二游
	YOUTYPE_13            //一三游
	YOUTYPE_14            //一四游
	YOUTYPE_34            //三四游
	YOUTYPE_24            //二四游
	YOUTYPE_23            //二三游
)

type SportSS510K struct {
	components2.Common
	//游戏变量
	SportMetaSS510K
	//	metadata.GameMetaDG
	m_GameLogic SportLogicSS510K //游戏逻辑
}

func (spss *SportSS510K) GetGameConfig() *static.GameConfig { //获取游戏相关配置
	return &spss.Config
}

//重置桌子数据
func (spss *SportSS510K) RepositTable(ResetAllData bool) {
	rand.Seed(time.Now().UnixNano())
	for _, v := range spss.PlayerInfo {
		v.Reset()
	}
	//游戏变量
	spss.GameEndStatus = static.GS_MJ_FREE
	if ResetAllData {
		//游戏变量
		spss.SportMetaSS510K.Reset()
	} else {
		//游戏变量
		spss.SportMetaSS510K.ResetForNext() //保留部分数据给下一局使用
	}
}

func (spss *SportSS510K) switchCard2Ox(_index int) int {
	_hight := (_index - 1) / 13
	_low := (_index-1)%13 + 1

	return (_low + (_hight << 4))
}

func (spss *SportSS510K) getCardIndexByOx(_card byte) int {
	low_index := int(0x0F)
	hight_index := int(0xF0)

	return ((low_index & int(_card)) + ((hight_index&int(_card))>>4)*13)
}

//解析配置的任务
func (spss *SportSS510K) ParseRule(strRule string) {

	xlog.Logger().Info("parserRule :" + strRule)

	//表示底分要除以10
	spss.Rule.NineSecondRoom = false

	spss.Rule.JuShu = spss.GetTableInfo().Config.RoundNum
	spss.Rule.Always1Round = true
	spss.Spay = 0
	spss.Rule.FangZhuID = spss.GetTableInfo().Creator
	spss.Rule.CreateType = spss.FriendInfo.CreateType

	spss.SerPay = 0 //好友房无茶水

	if len(strRule) == 0 {
		return
	}

	var _msg FriendRule_ss510k
	if err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg); err == nil {
		if _msg.Radix == 0 {
			spss.IBase = _msg.Difen // 级数底分
			spss.Rule.Radix = 1
		} else {
			spss.IBase = _msg.Difen // 级数底分
			spss.Rule.Radix = _msg.Radix
		}
		spss.SerPay = _msg.SerPay
		spss.FaOfTao = _msg.Fa
		spss.JiangOfTao = _msg.Jiang
		spss.Rule.ShareScoreType = _msg.ShareScoreType //贡分类型
		spss.Rule.Overtime_trust = _msg.Overtime_trust
		spss.Rule.Overtime_dismiss = _msg.Overtime_dismiss
		// 超时托管
		spss.PlayTime = _msg.Overtime_trust
		spss.Rule.NineSecondRoom = spss.PlayTime > 0
		if _msg.Overtime_trust <= 0 {
			spss.PlayTime = 15
		}
		// 底分
		spss.AddDiFen = _msg.AddDiFen
		spss.PlayMode = _msg.PlayMode                 // 游戏模式
		spss.RandTeamer = _msg.RandTeamer == "true"   // 随机队友
		spss.TrustPunish = _msg.TrustPunish == "true" // 托管罚分
		spss.Rule.DissmissCount = _msg.Dissmiss
	}
	if spss.IBase < 1 {
		spss.IBase = 1
	}
	//开关
	if spss.Debug > 0 {
		//Rule.JuShu = Debug;
	}
	if spss.Rule.DissmissCount != 0 {
		spss.SetDissmissCount(spss.Rule.DissmissCount)
	}
	if spss.TrustPunish {
		spss.OnWriteGameRecord(static.INVALID_CHAIR, "本局开启托管罚分")
	}
}

//! 开局
func (spss *SportSS510K) OnBegin() {
	xlog.Logger().Info("onbegin")
	spss.RepositTable(true)

	spss.OnWriteGameRecord(static.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range spss.PlayerInfo {
		v.OnBegin()
		//重置托管状态
		v.ChangeTRUST(false)
	}

	//设置状态
	spss.SetGameStatus(static.GS_MJ_PLAY)
	spss.ParseRule(spss.GetTableInfo().Config.GameConfig)
	spss.m_GameLogic.Rule = spss.Rule
	spss.CurCompleteCount = 0
	spss.VecGameEnd = []static.Msg_S_DG_GameEnd{}

	// 记录游戏开始时间
	spss.Common.GameBeginTime = time.Now()

	if spss.CurCompleteCount == 0 {
		spss.TableDeleteFangKa(1)
		if spss.GetPlayerCount() == static.MAX_PLAYER_4P {
			// 打乱玩家坐位
			spss.RandChangeSeat(true, true)
		}
	}
	//石首510k有花牌癞子
	spss.m_GameLogic.InitMagicPoint(logic2.CARDINDEX_SKY)
	spss.m_GameLogic.SetBombCount(4)                     //设置炸弹的最小长度
	spss.m_GameLogic.SetOnestrCount(254)                 //设置单顺的最小长度，254表示无顺子,石首510K没有顺子
	spss.m_GameLogic.SetPlayMode(spss.PlayMode)          //玩法模式
	spss.m_GameLogic.SetMaxCardCount(static.MAX_CARD_4P) //设置手牌最大长度
	if spss.PlayMode == 0 {
		spss.m_GameLogic.SetMaxCardCount(static.MAX_CARD_4P - 1) //经典玩法26张手牌
	}
	spss.m_GameLogic.SetMaxPlayerCount(static.MAX_PLAYER_4P) //设置玩家最大数目

	_, spss.AllCards = spss.m_GameLogic.CreateCards()

	// 设置离线解散时间30分钟
	spss.SetOfflineRoomTime(1800)

	spss.OnGameStart()
}

func (spss *SportSS510K) OnGameStart() {
	if !spss.CanContinue() {
		return
	}
	spss.StartNextGame()
}

//开始下一局游戏
func (spss *SportSS510K) StartNextGame() {
	spss.OnStartNextGame()

	// 恢复自动解散时间2分钟
	spss.SetDismissRoomTime(120)

	spss.GameState = meta2.GsNull          //吼牌阶段或打牌阶段.
	spss.GameEndStatus = static.GS_MJ_PLAY //当前小局游戏的状态
	spss.ReWriteRec = 0

	//发送最新状态
	for i := 0; i < spss.GetPlayerCount(); i++ {
		spss.SendUserStatus(i, static.US_PLAY) //把状态发给其他人
	}

	//记录日志
	spss.OnWriteGameRecord(static.INVALID_CHAIR, "开始StartNextGame......")
	spss.OnWriteGameRecord(static.INVALID_CHAIR, spss.GetTableInfo().Config.GameConfig)

	//重置所有玩家的状态
	for _, v := range spss.PlayerInfo {
		v.OnNextGame()
	}

	//解析规则
	//spss.ParseRule(spss.GameTable.Config.GameConfig)
	//spss.m_GameLogic.Rule = spss.Rule

	//设置状态
	spss.SetGameStatus(static.GS_MJ_PLAY)

	// 框架发送开始游戏后开始计算当前这一轮的局数
	spss.CurCompleteCount++

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(spss.GetTableId()+spss.KIND_ID*100+spss.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	spss.SiceCount = components2.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	spss.SuperMap = server2.GetDBMgr().LoadSuperAdmin()
	//分发扑克
	checkFlag := true
	spade3seat := -1
	var superManChair uint16 = static.INVALID_CHAIR
	var bigCards []byte
	var superPoint byte
	for i := 0; i < spss.GetPlayerCount(); i++ {
		player := spss.GetUserItemByChair(uint16(i))
		if player != nil {
			if rate, ok := spss.SuperMap[player.Uid]; ok {
				rdm := static.HF_GetRandom(100) + 1
				spss.OnWriteGameRecord(player.GetChairID(), fmt.Sprintf("super man, rate=%d, rdm=%d", rate, rdm))
				if rdm <= rate {
					superManChair = player.GetChairID()
					bigCards, superPoint, _ = GetSuperManBigCards(spss.m_GameLogic.GetAllMagics())
					spss.OnWriteGameRecord(player.GetChairID(), fmt.Sprintf("super man, big cards=%+v", bigCards))
					break
				}
			}
		}
	}
	for checkFlag {
		checkFlag = false
		spss.ThePaiCount, spss.PlayerCards, spss.HasKingNum, spade3seat = spss.m_GameLogic.RandCardData(spss.AllCards, spss.GetTableId(), superManChair, bigCards, superPoint)
		//同点牌加癞子个数>=10则重新发牌
		for i := 0; i < spss.GetPlayerCount(); i++ {
			if spss.m_GameLogic.CheckSamePointAndMagic(spss.PlayerCards[i]) {
				checkFlag = true
				break
			}
		}
	}

	//四人场第一局首先发到黑桃3的人为庄
	if spss.GetPlayerCount() == static.MAX_PLAYER_4P && spss.CurCompleteCount == 1 && spade3seat >= 0 && spade3seat < static.MAX_PLAYER_4P {
		spss.Nextbanker = uint16(spade3seat)
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	spss.initDebugCards("shishou510k_test", &spss.PlayerCards, &spss.Nextbanker, &spss.DownPai)
	//////////////读取配置文件设置牌型end////////////////////////////////////

	//重新计算牌数目
	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		spss.ThePaiCount[i] = spss.m_GameLogic.GetCardNum(spss.PlayerCards[i], static.MAX_CARD_4P)
	}

	//发送玩家的牌数目
	if spss.ShowHandCardCnt {
		spss.SendPaiCount(static.MAX_PLAYER_4P)
	}

	//确定庄家，随机坐庄
	if spss.Nextbanker >= uint16(spss.GetPlayerCount()) {
		rand_num := rand.Intn(1000)
		spss.Banker = uint16(rand_num % spss.GetPlayerCount())
	} else {
		spss.Banker = spss.Nextbanker
	}
	spss.Whoplay = spss.Banker

	// 玩家队友固定为对家
	tmpNextPlayer := spss.GetNextFullSeat(spss.Banker)
	spss.BankParter = spss.GetNextFullSeat(tmpNextPlayer)

	// 存在调换座位的情况下 会再次发送constant.MsgTypeTableInfo消息 字段Step当前第几局 需要在SetBegin里self.Step++之前发送 客户端的已有逻辑中会执行++
	spss.GetTable().SetBegin(true)

	for seat := 0; seat < spss.GetPlayerCount(); seat++ {
		//详细日志
		handCardStr := string("发牌后手牌:")
		for i := 0; i < static.MAX_CARD_4P; i++ {
			temCardStr := fmt.Sprintf("0x%02x,", spss.switchCard2Ox(int(spss.PlayerCards[seat][i])))
			handCardStr += temCardStr
		}
		spss.OnWriteGameRecord(uint16(seat), handCardStr)

		spss.ReplayRecord.R_HandCards[seat] = append(spss.ReplayRecord.R_HandCards[seat], static.HF_BytesToInts(spss.PlayerCards[seat][:])...)
	}

	//开局写入当前玩家分数
	spss.addTurnOverReplayOrder(0, DG_REPLAY_OPT_TURN_OVER, 0, spss.PlayerScore, spss.PlayerTotalScore, true)

	//构造数据,发送开始信息
	var GameStart static.Msg_S_DG_GameStart
	GameStart.BankerUser = spss.Banker
	GameStart.CurrentUser = spss.Whoplay

	//向每个玩家发送数据
	for i := 0; i < spss.GetPlayerCount(); i++ {
		_item := spss.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//设置变量
		GameStart.MySeat = uint16(i)
		//GameStart.Overtime = spss.LimitTime
		GameStart.CardCount = spss.ThePaiCount[i]
		for c := 0; c < static.MAX_CARD_4P; c++ {
			GameStart.CardData[c] = spss.PlayerCards[i][c]
		}
		//记录玩家初始分

		//TODO 玩家分数设置
		spss.ReplayRecord.R_Score[i] = spss.GetUserTotalScore(uint16(i))

		//发送数据
		spss.SendPersonMsg(consts.MsgTypeGameStart, GameStart, uint16(i))
	}

	spss.EndRoar(false)

	//发送随机任务
	//spss.SendTaskID(true,spss.Whoplay);
	spss.GameTask.SendTaskID(spss.Common, true, spss.Whoplay)
}

//发送操作
func (spss *SportSS510K) SendPaiCount(wChairID uint16) {

	//构造数据
	var SendCount static.Msg_S_DG_SendCount
	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		SendCount.CardCount[i] = spss.ThePaiCount[i]
	}
	//发送数据
	if wChairID >= static.MAX_PLAYER_4P {
		spss.SendTableMsg(consts.MsgTypeSendPaiCount, SendCount)
	} else {
		for _, v := range spss.PlayerInfo {
			if v.Seat == uint16(wChairID) {
				spss.SendPersonMsg(consts.MsgTypeSendPaiCount, SendCount, v.Seat)
			}
		}
	}
}

//发送本轮分
func (spss *SportSS510K) SendTurnScore(wChairID uint16) {
	var turnScore static.Msg_S_DG_TurnScore
	turnScore.TurnScore = spss.CardScore

	if wChairID >= static.MAX_PLAYER_4P {
		spss.SendTableMsg(consts.MsgTypeTurnScore, turnScore)
	} else {
		spss.SendPersonMsg(consts.MsgTypeTurnScore, turnScore, wChairID)
	}
}

//发送玩家分
func (spss *SportSS510K) SendPlayerScore(wChairID uint16, wGetChairID uint16, iGetScore int, isSync bool) {
	var playScore Msg_S_DG_PlayScore
	playScore.ChairID = wGetChairID
	playScore.GetScore = iGetScore

	//本轮炸弹分
	if spss.LastOutType == TYPE_BOMB_NOMORL && !isSync {
		l := spss.m_GameLogic.GetCardNum(spss.AllPaiOut[spss.WhoLastOut], static.MAX_CARD_4P)
		re := spss.m_GameLogic.GetType(spss.AllPaiOut[spss.WhoLastOut], int(l), 0, 0, 0)
		playScore.BombScore = spss.m_GameLogic.getBombScore(re.BombLevel)
	}

	scoretemp := playScore.GetScore + playScore.BombScore

	for i := 0; i < spss.GetPlayerCount(); i++ {
		playScore.PlayScore[i] = spss.PlayerCardScore[i]
		if i%2 == int(spss.WhoLastOut)%2 {
			spss.PlayerScore[i] += scoretemp
			spss.PlayerTotalScore[i] += scoretemp
			spss.BombScore[i] += playScore.BombScore
		} else {
			//spss.PlayerScore[i]-=playScore.BombScore
			//spss.PlayerTotalScore[i]-=playScore.BombScore
			//spss.BombScore[i]-=playScore.BombScore
		}
		playScore.PlayerScore[i] = spss.PlayerScore[i]
		playScore.PlayerTotalScore[i] = spss.PlayerTotalScore[i]
	}

	if wChairID >= static.MAX_PLAYER_4P {
		spss.SendTableMsg(consts.MsgTypePlayScore, playScore)
	} else {
		spss.SendPersonMsg(consts.MsgTypePlayScore, playScore, wChairID)
	}
}

//发送声音
func (spss *SportSS510K) SendPlaySoundMsg(seat uint16, bySoundType byte) {
	var soundmsg static.Msg_S_DG_PlaySound
	soundmsg.CurrentUser = seat
	soundmsg.SoundType = bySoundType

	spss.SendTableMsg(consts.MsgTypePlaySound, soundmsg)
}

//发送玩家是几游
func (spss *SportSS510K) SendPlayerTurn(wChairID uint16) {
	var turnmsg static.Msg_S_DG_SendTurn

	///////////////////////////////////////////////////////////////////////////////////////////////////////
	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		turnmsg.Turn[i] = spss.PlayerTurn[i]
	}
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	if wChairID >= static.MAX_PLAYER_4P {
		spss.SendTableMsg(consts.MsgTypeSendTurn, turnmsg)
	} else {
		spss.SendPersonMsg(consts.MsgTypeSendTurn, turnmsg, wChairID)
	}
}

//把seat1的牌发给seat2
func (spss *SportSS510K) SendPaiToTeamer(seat1 uint16, seat2 uint16) {
	if seat1 < 0 || seat1 >= static.MAX_PLAYER_4P {
		return
	}
	if seat2 < 0 || seat2 >= static.MAX_PLAYER_4P {
		return
	}

	//if(!spss.BMingJiFlag){return}//没有明鸡，不发送队友的牌

	var teamerPai static.Msg_S_DG_TeamerPai
	teamerPai.WhoPai = seat1
	for i := 0; i < static.MAX_CARD_4P; i++ {
		if spss.m_GameLogic.IsValidCard(spss.PlayerCards[seat1][i]) {
			teamerPai.CardData[teamerPai.CardCount] = spss.PlayerCards[seat1][i]
			teamerPai.CardCount++
		}
	}
	spss.SendPersonMsg(consts.MsgTypeTeamerPai, teamerPai, seat2)
}

// 发送权限
func (spss *SportSS510K) SendPower(whoplay uint16, iPower int, iWaitTime int) {
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	//构造数据
	var power static.Msg_S_DG_Power
	iOvertime := 0
	if iWaitTime > 0 {
		iOvertime = iWaitTime
	}
	power.CurrentUser = whoplay
	power.Power = iPower
	item := spss.GetUserItemByChair(whoplay)
	if item.CheckTRUST() {
		if iOvertime > spss.AutoOutTime {
			iOvertime = spss.AutoOutTime
		}
	}
	power.Overtime = int64(iOvertime)

	//详细日志
	LogStr := fmt.Sprintf("SUB_S_POWER seat=%d power=%d,time=%d ", power.CurrentUser, power.Power, power.Overtime)
	spss.OnWriteGameRecord(power.CurrentUser, LogStr)

	spss.SendTableMsg(consts.MsgTypeSendPower, power)

	spss.PowerStartTime = time.Now().Unix() //权限开始时间
	spss.setLimitedTime(int64(iOvertime + 1))
	if spss.GameState == meta2.GsPlay {
		//SetActionStep(AS_PLAY,nTime + 1);//设置等待时间，服务端多等一下
	} else if spss.GameState == meta2.GsRoarPai {
		//SetActionStep(AS_ROAR,nTime + 1);//设置等待时间，服务端多等一下
	} else if spss.GameState == meta2.Gs4KingScore {
		//SetActionStep(AS_ROAR,nTime + 1);//设置等待时间，服务端多等一下
	}
}

func (spss *SportSS510K) setLimitedTime(iLimitTime int64) {
	// fmt.Println(fmt.Sprintf("limitetimeOP(%d)", spss.Rule.limitetimeOP))
	spss.LimitTime = time.Now().Unix() + iLimitTime
	spss.GameTimer.SetLimitTimer(int(iLimitTime))
}

func (spss *SportSS510K) freeLimitedTime() {
	spss.GameTimer.KillLimitTimer()
}

func (spss *SportSS510K) LockTimeOut(cUser uint16, iTime int64) {
	if cUser < 0 || cUser > uint16(spss.GetPlayerCount()) {
		return
	}

	_userItem := spss.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = iTime
}

func (spss *SportSS510K) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(spss.GetPlayerCount()) {
		return
	}

	_userItem := spss.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0
}

//计时器事件
func (spss *SportSS510K) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	//游戏定时器
	if dwTimerID == components2.GameTime_Nine {
		if spss.Rule.NineSecondRoom {
			spss.OnAutoOperate(true)
		}
	}
	return true
}

//暂时空着
func (spss *SportSS510K) OnAutoOperate(bBreakin bool) {
	fmt.Println("自动操作")
	//详细日志
	LogStr := string("OnAutoOperate 自动操作!!! ")
	spss.OnWriteGameRecord(spss.Whoplay, LogStr)

	item := spss.GetUserItemByChair(spss.Whoplay)
	if item == nil {
		return
	}

	if spss.GameState == meta2.GsRoarPai {
		if spss.Whoplay < uint16(spss.GetPlayerCount()) && !item.CheckTRUST() {
			spss.AutoTuoGuan(spss.Whoplay)
		}
		spss.OnRoarAction(spss.Whoplay, false)
	} else if spss.GameState == meta2.GsPlay {
		tempCurPlay := spss.Whoplay

		if spss.Whoplay < uint16(spss.GetPlayerCount()) && !item.CheckTRUST() {
			spss.AutoTuoGuan(spss.Whoplay)
		}

		//构造数据
		var outmsg static.Msg_C_DG_OutCard
		outmsg.CurrentUser = spss.Whoplay
		if spss.WhoLastOut >= uint16(spss.GetPlayerCount()) {
			if true {
				//c++蕲春的做法
				buf := [static.MAX_CARD]byte{}
				n := 0
				for i := 0; i < static.MAX_CARD_4P; i++ {
					if spss.m_GameLogic.IsValidCard(spss.PlayerCards[spss.Whoplay][i]) {
						buf[n] = spss.PlayerCards[spss.Whoplay][i]
						n++
					}
				}
				buf = spss.m_GameLogic.SortByIndex(buf, static.MAX_CARD_4P, true)
				////不能只剩下癞子
				_, cFlag := spss.m_GameLogic.IsAllEqualExceptMagic(buf, n)
				if cFlag {
					outmsg.CardCount = spss.ThePaiCount[spss.Whoplay]
					outmsg.CardData = buf
					re := spss.m_GameLogic.GetType(buf, n, 0, 0, 0)
					outmsg.CardType = re.Cardtype
				} else {
					outmsg.CardCount = 1
					outmsg.CardData[0] = buf[0]
					outmsg.CardType = TYPE_ONE
				}
			} else {
				//智能一点的做法
				spss.m_GameLogic.GetGroupType(spss.PlayerCards[spss.Whoplay])
				_, beepOut := spss.m_GameLogic.BeepFirstCardOut()
				for k := 0; k < len(beepOut[0].Indexes); k++ {
					outmsg.CardData[k] = beepOut[0].Indexes[k]
				}
				outmsg.CardCount = beepOut[0].Count
				outmsg.CardType = spss.LastOutType
			}
		} else {
			outmsg.CardCount = 0
			if false {
				//如果跟出要出牌就需要用下面的牌
				spss.m_GameLogic.GetGroupType(spss.PlayerCards[spss.Whoplay])
				_, beepOut := spss.m_GameLogic.BeepCardOut(spss.AllPaiOut[spss.WhoLastOut], spss.LastOutType)
				if len(beepOut) == 0 {
					outmsg.CardCount = 0
				} else {
					for k := 0; k < len(beepOut[0].Indexes); k++ {
						outmsg.CardData[k] = beepOut[0].Indexes[k]
					}
					outmsg.CardCount = beepOut[0].Count
					outmsg.CardType = spss.LastOutType
				}
			}
		}
		//详细日志
		LogStr := fmt.Sprintf("托管出牌 OnAutoOperate UserID=%d ,CardCount=%d,牌数据:", outmsg.CurrentUser, outmsg.CardCount)
		for i := 0; i < static.MAX_CARD_4P; i++ {
			if outmsg.CardData[i] > 0 {
				CardStr := fmt.Sprintf("0x%02x,", spss.switchCard2Ox(int(outmsg.CardData[i])))
				LogStr += CardStr
			}
		}
		spss.OnWriteGameRecord(spss.Whoplay, LogStr)
		spss.OnUserOutCard(&outmsg)

		spss.AutoCardCounts[tempCurPlay]++
		if spss.AutoCardCounts[tempCurPlay] >= 5 {
			// 连续5轮自动出牌，结束游戏
			//spss.OnGameEndUserLeft(tempCurPlay, metadata.GOT_TUOGUAN)
		}
	} else if spss.GameState == meta2.Gs4KingScore {
		if spss.WhoHas4KingPower == static.INVALID_CHAIR {
			return
		}

		// 加入托管
		if spss.WhoHas4KingPower < uint16(spss.GetPlayerCount()) && !spss.TuoGuanPlayer[spss.WhoHas4KingPower] {
			spss.AutoTuoGuan(spss.WhoHas4KingPower)
		}

		// 不换分
		//spss.On4KingScore(spss.WhoHas4KingPower, 0)
		// 日志
		spss.OnWriteGameRecord(spss.WhoHas4KingPower, "托管4王换分 弃")
	}
}

func (spss *SportSS510K) AutoTuoGuan(theSeat uint16) int {
	item := spss.GetUserItemByChair(theSeat)
	if item == nil {
		return 0
	}

	var msgtg static.Msg_S_DG_Trustee
	msgtg.Trustee = true
	msgtg.ChairID = theSeat

	if spss.GameState == meta2.GsNull || theSeat >= uint16(spss.GetPlayerCount()) {
		//游戏状态 GameState，1表示吼牌阶段，2表示打牌阶段
		return 0
	}
	//详细日志
	LogStr := fmt.Sprintf("超时托管,CMD_S_Tuoguan_CB AutoTuoGuan msgtg.theFlag=%t msgtg.theSeat=%d ", msgtg.Trustee, msgtg.ChairID)
	spss.OnWriteGameRecord(theSeat, LogStr)

	if spss.GameState == meta2.GsPlay || spss.GameState == meta2.GsRoarPai || spss.GameState == meta2.Gs4KingScore {
		item.Ctx.HasTrustee = true
		if true == msgtg.Trustee {
			spss.TuoGuanPlayer[theSeat] = true
			spss.TrustCounts[theSeat]++
			item.ChangeTRUST(true)
			if theSeat == spss.Whoplay { //如果是当前的玩家，那么重新设置一下开始时间
				//spss.setLimitedTime(int64(spss.AutoOutTime))//已经超时了，马上就要切换牌权了，不用在设置他的时间了
			}
			spss.SendTableMsg(consts.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			spss.addReplayOrder(msgtg.ChairID, DG_REPLAY_OPT_TUOGUAN, 1, []int{})

		} else {
			spss.TuoGuanPlayer[theSeat] = false
			item.ChangeTRUST(true)
			spss.SendTableMsg(consts.MsgTypeGameTrustee, msgtg)
			//回放托管记录
			spss.addReplayOrder(msgtg.ChairID, DG_REPLAY_OPT_TUOGUAN, 0, []int{})
		}
	}
	return 1
}

//硬牌动作响应
func (spss *SportSS510K) OnRoarAction(seat uint16, bRoar bool) bool {
	if seat < 0 || seat >= static.MAX_PLAYER_4P {
		return false
	}
	if seat != spss.Whoplay {
		return false
	}

	if spss.WhoReady[seat] {
		return false
	}

	spss.WhoReady[seat] = true

	//变量定义
	var roar static.Msg_S_DG_Roar
	roar.CurrentUser = seat
	roar.RoarFlag = 0
	if bRoar {
		roar.RoarFlag = 1
	}

	spss.SendTableMsg(consts.MsgTypeRoar, roar)

	//回放增加吼牌记录
	spss.addReplayOrder(seat, DG_REPLAY_OPT_HOUPAI, int(roar.RoarFlag), []int{})

	if bRoar { //如果硬牌，那么结束硬牌动作
		spss.EndRoar(true)
		return true
	} else {
		z := 0
		for i := 0; i < static.MAX_PLAYER_4P; i++ {
			if spss.WhoReady[i] {
				z++
			}
		}
		if z >= static.MAX_PLAYER_4P { //开始游戏
			spss.EndRoar(false)
		} else {
			spss.GoNextPlayer()
		}
	}
	return true
}

func (spss *SportSS510K) StartPlay(whoplay uint16) {
	spss.CardScore = 0
	// 开始
	spss.GameState = meta2.GsPlay

	iPower := 2
	spss.SendPower(whoplay, iPower, spss.PlayTime+5) //庄家第一次出牌的时间加5秒
}

func (spss *SportSS510K) StartRoar(theSeat uint16) {
	// 开始进入吼牌状态
	spss.GameState = meta2.GsRoarPai
	iPower := 1
	spss.SendPower(theSeat, iPower, spss.RoarTime)
}

func (spss *SportSS510K) EndRoar(bRoar bool) {
	//var endroarmsg public.Msg_S_DG_EndRoar

	////有人吼牌
	//if bRoar {
	//	spss.WhoRoar = spss.Whoplay
	//	endroarmsg.RoarUser = spss.WhoRoar
	//	spss.GameType = GT_ROAR
	//} else {
	///////////////////////////////////////////////////////////////////////////////////////////
	//	//没人吼牌
	spss.GameType = meta2.GT_NORMAL
	spss.WhoRoar = static.INVALID_CHAIR

	tmp1 := spss.GetNextFullSeat(spss.Banker)
	tmp2 := spss.GetNextFullSeat(tmp1)
	spss.BankParter = tmp2
	//	endroarmsg.RoarUser = spss.WhoRoar
	//	spss.GetJiaoPai() //得到叫牌
	//////////////////////////////////////////////////////////////////////////////////////////
	//}
	////吼牌的为庄家了
	//if bRoar {
	//	spss.Banker = spss.WhoRoar
	//	spss.BankParter = public.INVALID_CHAIR
	//	//详细日志
	//	LogStr := string("包牌(吼牌)")
	//	spss.OnWriteGameRecord(spss.WhoRoar, LogStr)
	//}
	//////////////////////////////////////////////////////////////////////////////////////////
	spss.Whoplay = spss.Banker
	//endroarmsg.BankUser = spss.Banker
	//endroarmsg.JiaoPai = spss.RoarPai
	//spss.SendTableMsg(constant.MsgTypeEndRoar, endroarmsg)
	//////////////////////////////////////////////////////////////////////////////////////////
	//回放增加结束吼牌记录
	//spss.addReplayOrder(spss.Banker, DG_REPLAY_OPT_END_HOUPAI, int(spss.RoarPai), []int{})

	//详细日志
	LogStr := string("为庄家")
	spss.OnWriteGameRecord(spss.Whoplay, LogStr)

	spss.StartPlay(spss.Whoplay)
}

func (spss *SportSS510K) GoNextPlayer() {
	///////////////////////////////////////////////////////////////////////////////////////////////////////
	for iPlayer := 0; iPlayer < spss.GetPlayerCount(); iPlayer++ {
		if spss.Whoplay >= uint16(spss.GetPlayerCount())-1 {
			spss.Whoplay = 0
		} else {
			spss.Whoplay++
		}

		//如果当前玩家出完了
		if spss.WhoAllOutted[spss.Whoplay] {
			if spss.WhoLastOut == spss.Whoplay { //这个玩家是不是上一次出牌玩家
				break
			} else {
				continue
			}
		} else { //没出完？
			break
		}
	}

	///////////////////////////////////////////////////////////////////////////////////////////////////////
	if spss.WhoLastOut == spss.Whoplay {
		if spss.EndTurn() {
			return
		}

	}
	if spss.GameState == meta2.GsRoarPai {
		spss.StartRoar(spss.Whoplay)
	} else if spss.GameState == meta2.GsPlay {
		dwPower := 2
		spss.SendPower(spss.Whoplay, dwPower, spss.PlayTime)
	}

	for i := 0; i < static.MAX_CARD_4P; i++ {
		spss.AllPaiOut[spss.Whoplay][i] = 0
	}
}

// 结束一轮
func (spss *SportSS510K) EndTurn() bool {
	for i := 0; i < static.MAX_CARD_4P; i++ {
		spss.LastPaiOut[spss.WhoLastOut][i] = spss.AllPaiOut[spss.WhoLastOut][i]
	}

	//打分的模式
	if spss.GameType == meta2.GT_NORMAL {
		BombFlag := false
		if spss.LastOutType == TYPE_BOMB_NOMORL {
			l := spss.m_GameLogic.GetCardNum(spss.AllPaiOut[spss.WhoLastOut], static.MAX_CARD_4P)
			re := spss.m_GameLogic.GetType(spss.AllPaiOut[spss.WhoLastOut], int(l), 0, 0, 0)
			if spss.m_GameLogic.getBombScore(re.BombLevel) > 0 {
				BombFlag = true
			}
		}
		spss.PlayerCardScore[spss.WhoLastOut] += spss.CardScore
		spss.SendPlayerScore(static.MAX_PLAYER_4P, spss.WhoLastOut, spss.CardScore, false)

		//回放增加本轮抓分
		if BombFlag || spss.CardScore > 0 {
			spss.addTurnOverReplayOrder(spss.WhoLastOut, DG_REPLAY_OPT_TURN_OVER, spss.CardScore, spss.PlayerScore, spss.PlayerTotalScore, true)
		} else {
			spss.addTurnOverReplayOrder(spss.WhoLastOut, DG_REPLAY_OPT_TURN_OVER, spss.CardScore, spss.PlayerScore, spss.PlayerTotalScore, false)
		}

		spss.CardScore = 0 //清零
		spss.OutScorePai = [24]byte{}
		spss.SendTurnScore(static.MAX_PLAYER_4P)
		//这里可以判断下是否结束游戏
		//if spss.JudgeEndGame(spss.WhoLastOut) {
		//	return true
		//}
	} else {
		spss.addTurnOverReplayOrder(spss.WhoLastOut, DG_REPLAY_OPT_TURN_OVER, 0, [4]int{}, [4]int{}, false)
	}

	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		for j := 0; j < static.MAX_CARD_4P; j++ {
			spss.AllPaiOut[i][j] = 0
		}
		spss.WhoPass[i] = false
	}

	spss.WhoLastOut = static.INVALID_CHAIR

	if spss.WhoAllOutted[spss.Whoplay] { //这里说明当前玩家出完了
		teamer := spss.GetTeamer(spss.Whoplay)

		//if spss.BMingJiFlag { //如果明鸡确定，那么是该玩家的队友接风
		spss.Whoplay = teamer
		spss.SendPlaySoundMsg(spss.Whoplay, static.TY_JieFeng)
		//} else { //否则就是下家接风
		//
		//}
	}
	for {
		if spss.WhoAllOutted[spss.Whoplay] {
			if spss.Whoplay >= uint16(spss.GetPlayerCount())-1 {
				spss.Whoplay = 0
			} else {
				spss.Whoplay++
			}
		} else {
			break
		}
	}

	spss.EndOut() //结束一轮
	return false
}
func (spss *SportSS510K) EndOut() {
	spss.LastOutType = TYPE_NULL
	spss.LastOutTypeClient = TYPE_NULL

	var endout static.Msg_S_DG_EndOut
	endout.CurrentUser = spss.Whoplay
	spss.SendTableMsg(consts.MsgTypeEndOut, endout)
}

func (spss *SportSS510K) GetTeamer(who uint16) uint16 {
	re := uint16(0)
	if who >= static.MAX_PLAYER_4P {
		return 0
	}
	if who == spss.Banker {
		re = spss.BankParter
	} else if who == spss.BankParter {
		re = spss.Banker
	} else { //闲家
		for i := uint16(0); i < static.MAX_PLAYER_4P; i++ {
			if i == who {
				continue
			}
			if i == spss.BankParter {
				continue
			}
			if i == spss.Banker {
				continue
			}
			re = i
			break
		}
	}
	return re
}
func (spss *SportSS510K) GetTeamScore(seat uint16) int {
	if spss.GameType != meta2.GT_NORMAL {
		return 0
	}
	if seat < 0 || seat >= static.MAX_PLAYER_4P {
		return 0
	}
	teamer := spss.GetTeamer(seat)
	if spss.PlayerTurn[seat] == 1 || spss.PlayerTurn[teamer] == 1 { //一游玩家的分
		return spss.PlayerCardScore[seat] + spss.PlayerCardScore[teamer]
	} else if spss.PlayerTurn[seat] == 2 || spss.PlayerTurn[teamer] == 2 {
		who2you := teamer
		if spss.PlayerTurn[seat] == 2 {
			who2you = seat
		}
		return spss.PlayerCardScore[who2you]
	}
	return 0
}

func (spss *SportSS510K) AddSpecailScore(Score *[meta2.MAX_PLAYER]int, seat uint16, base int) {
	if seat < 0 && seat >= static.MAX_PLAYER_4P {
		return
	}

	iAddFan := spss.WhoSame510K[seat] + spss.Who7Xi[seat] + spss.Who8Xi[seat] + spss.PlayKingBomb[seat]
	iTempScore := iAddFan * base

	//这里来计算各自该赢的，特殊的分
	for i := 0; i < spss.GetPlayerCount(); i++ {
		if uint16(i) == seat {
			spss.XiScore[i] += 3 * iTempScore
			Score[i] += 3 * iTempScore
		} else {
			spss.XiScore[i] -= iTempScore //其他人
			Score[i] -= iTempScore
		}
	}
}

//! 加载测试麻将数据
func (spss *SportSS510K) initDebugCards(configName string, cbRepertoryCard *[meta2.MAX_PLAYER][static.MAX_CARD]byte, wBankerUser *uint16, byDownPai *byte) (err error) {
	defer func() {
		if err != nil {
			spss.OnWriteGameRecord(static.INVALID_CHAIR, err.Error())
		}
	}()
	//! 做牌文件配置
	var debugCardConfig *meta2.CardConfig = new(meta2.CardConfig)

	fileName := fmt.Sprintf("./%s%d_%d", configName, spss.GetPlayerCount(), spss.Rule.FangZhuID)
	spss.OnWriteGameRecord(static.INVALID_CHAIR, "开始根据房主id读取做牌文件，文件名："+fileName)
	if !static.GetJsonMgr().ReadData("./json", fileName, debugCardConfig) {
		configName = fmt.Sprintf("./%s%d", configName, spss.GetPlayerCount())
		spss.OnWriteGameRecord(static.INVALID_CHAIR, "开始读取做牌文件，文件名："+configName)
		if !static.GetJsonMgr().ReadData("./json", configName, debugCardConfig) {
			return errors.New("做牌文件:读取失败")
		}
	}

	// 是否开启做牌
	if debugCardConfig.IsAble == 1 {
		//检查做牌文件是否做牌异常
		for _, handCards := range debugCardConfig.UserCards {
			if len(strings.Split(handCards, ",")) != 27 {
				return errors.New("做牌文件:手牌长度不为27")
			}
		}
		//检查牌堆牌是否正常
		if len(debugCardConfig.RepertoryCard) != debugCardConfig.RepertoryCardCount*5-1 {
			return errors.New(fmt.Sprintf("做牌文件:牌库牌数量不一致:::RepertoryCard:[%d]>>>实际做牌牌库数量:[%d]", debugCardConfig.RepertoryCardCount*5-1, len(debugCardConfig.RepertoryCard)))
		}

		// 重置HasKingNum
		for i := 0; i < static.MAX_PLAYER_4P; i++ {
			spss.HasKingNum[i] = 0
		}

		// 设置玩家手牌
		for userIndex, handCards := range debugCardConfig.UserCards {
			byCardsCount := byte(0)
			_item := spss.GetUserItemByChair(uint16(userIndex))
			if _item != nil {
				_item.Ctx.InitCardIndex() //清理手牌
				handCardsSlice := strings.Split(handCards, ",")
				for _, cardStr := range handCardsSlice {

					if _, cardValue := spss.GetCardDataByStr(cardStr); cardValue == static.INVALID_BYTE {
						//return errors.New(fmt.Sprintf("做牌文件:第%d号玩家做牌异常：%s", userIndex, cardStr))
					} else {
						//_item.Ctx.DispatchCard(cardValue)
						(*cbRepertoryCard)[userIndex][byCardsCount] = cardValue
						byCardsCount++
						if cardValue == logic2.CARDINDEX_SMALL || cardValue == logic2.CARDINDEX_BIG {
							spss.HasKingNum[userIndex]++
						}
						if byCardsCount >= static.MAX_CARD_4P {
							break
						}
					}
				}
			}
		}
		//设置牌堆牌
		repertoryCards := strings.Split(debugCardConfig.RepertoryCard, ",")
		for _, cardStr := range repertoryCards {
			if _, cardValue := spss.GetCardDataByStr(cardStr); cardValue == static.INVALID_BYTE {
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

//写日志记录
func (spss *SportSS510K) writeGameRecord() {
	//写日志记录
	spss.OnWriteGameRecord(static.INVALID_CHAIR, "开始蕲春打拱  发牌......")

	// 玩家手牌
	//for _, v := range spss.PlayerInfo {
	//	if v.Seat != public.INVALID_CHAIR {
	//		handCardStr := fmt.Sprintf("发牌后手牌:%s", spss.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
	//		spss.OnWriteGameRecord(uint16(v.Seat), handCardStr)
	//	}
	//}

	// 牌堆牌
	//leftCardStr := fmt.Sprintf("牌堆牌:%s", spss.m_GameLogic.SwitchToCardNameByDatas(spss.RepertoryCard[0:spss.LeftCardCount+2], 0))
	//spss.OnWriteGameRecord(public.INVALID_CHAIR, leftCardStr)

	//赖子牌
	//magicCardStr := fmt.Sprintf("癞子牌:%s", spss.m_GameLogic.SwitchToCardNameByData(spss.MagicCard, 1))
	//spss.OnWriteGameRecord(public.INVALID_CHAIR, magicCardStr)
}

//! 解散
func (spss *SportSS510K) OnEnd() {
	if spss.IsGameStarted() {
		spss.OnGameOver(static.INVALID_CHAIR, static.GER_DISMISS)
	}
}

//! 单局结算
func (spss *SportSS510K) OnGameOver(wChairID uint16, cbReason byte) bool {
	spss.OnEventGameEnd(wChairID, cbReason)
	return true
}

//! 初始化游戏
func (spss *SportSS510K) OnInit(table base2.TableBase) {
	spss.KIND_ID = table.GetTableInfo().KindId
	spss.Config.StartMode = static.StartMode_FullReady
	spss.Config.PlayerCount = 4 //玩家人数
	spss.Config.ChairCount = 4  //椅子数量
	spss.PlayerInfo = make(map[int64]*components2.Player)

	spss.RepositTable(true)
	spss.SetGameStartMode(static.StartMode_FullReady)
	spss.GameTable = table
	spss.Init()
	spss.Unmarsha(table.GetTableInfo().GameInfo)
	table.GetTableInfo().GameInfo = ""

	// 设置自动解散时间2分钟
	spss.SetDismissRoomTime(120)
	// 设置离线解散时间30分钟
	spss.SetOfflineRoomTime(1800)
	// 离线60s未准备踢出
	//if spss.GameTable.GetTableInfo().JoinType == constant.NoCheat {
	//	spss.SetOfflineRoomTime(60)
	//}
	var _msg FriendRule_ss510k
	if err := json.Unmarshal(static.HF_Atobytes(table.GetTableInfo().Config.GameConfig), &_msg); err == nil {
		if _msg.FleeTime != 0 {
			spss.SetOfflineRoomTime(_msg.FleeTime)
		}
	}
}

//! 发送消息
func (spss *SportSS510K) OnMsg(msg *base2.TableMsg) bool {
	wChairID := spss.GetChairByUid(msg.Uid)
	if wChairID == static.INVALID_CHAIR {
		spss.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::找不到玩家的座位号:%d", msg.Uid))
		return true
	}

	switch msg.Head {
	case consts.MsgTypeGameBalanceGameReq: //! 请求总结算信息 //暂时没有

		var _msg static.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spss.CalculateResultTotal_Rep(&_msg)
		}
	case consts.MsgTypeGameOutCard: //! 出牌消息
		var _msg static.Msg_C_DG_OutCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			//详细日志
			LogStr := fmt.Sprintf("OnUserOutCard UserID=%d ,CardCount=%d,牌数据:", _msg.CurrentUser, _msg.CardCount)
			for i := 0; i < static.MAX_CARD_4P; i++ {
				if _msg.CardData[i] > 0 {
					CardStr := fmt.Sprintf("0x%02x,", spss.switchCard2Ox(int(_msg.CardData[i])))
					LogStr += CardStr
				}
			}
			spss.OnWriteGameRecord(spss.Whoplay, LogStr)

			return spss.OnUserOutCard(&_msg)
		}
	case consts.MsgTypeGameTrustee: //用户托管
		var _msg static.Msg_C_Trustee
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spss.onUserTustee(&_msg)
			//详细日志
			LogStr := fmt.Sprintf("主动托管动作(true托管,false取消):TrustFlag=%t ", _msg.Trustee)
			spss.OnWriteGameRecord(spss.GetChairByUid(_msg.Id), LogStr)
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
			spss.OnWriteGameRecord(_msg.CurrentUser, LogStr)
			return spss.OnRoarAction(_msg.CurrentUser, bRoarFlag)

		}
	case consts.MsgTypeGameGoOnNextGame: //下一局 //暂时没有下一局
		//详细日志
		LogStr := string("OnUserClientNextGame!!! ")
		spss.OnWriteGameRecord(spss.Whoplay, LogStr)

		var _msg static.Msg_C_GoOnNextGame
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			spss.OnUserClientNextGame(&_msg)
		}
	case consts.MsgTypeGameDismissFriendResult: //申请解散玩家选择
		if spss.GameEndStatus == byte(static.GS_FREE) {
			var _msg static.Msg_C_DismissFriendResult
			if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
				if _msg.Flag == false {
					//不同意解散,托管玩家自动准备
					//是否托管
					for i := 0; i < spss.GetPlayerCount(); i++ {
						item := spss.GetUserItemByChair(uint16(i))
						if item == nil || !item.CheckTRUST() {
							continue
						}
						spss.AutoNextGame(item.Uid)
					}
				}
			}
		}
	case consts.MsgTypeGameChangeSeat: //申请换座
		if spss.GameEndStatus == byte(static.GS_FREE) {
			//若申请换座玩家已准备
			item := spss.GetUserItemByChair(uint16(wChairID))
			if item != nil && item.Ready {
				spss.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::准备状态下无法换座:%d", msg.Uid))
				spss.SendGameNotificationMessage(wChairID, fmt.Sprintf("准备状态下无法换座"))
				break
			}
			var _msg Msg_C_DG_CHANGESEAT
			if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
				_item := spss.GetUserItemByChair(uint16(_msg.ChairID))
				if _msg.ChairID < 0 || _msg.ChairID >= byte(spss.GetPlayerCount()) {
					break
				}
				//若要换的座上有人
				if _item != nil {
					spss.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::该座位%d上已有玩家%d存在", _msg.ChairID, _item.Uid))
					spss.SendGameNotificationMessage(wChairID, fmt.Sprintf("该座位上已有玩家存在"))
					break
				}
				//调换座位
				spss.ReSeat(item.Uid, int(_msg.ChairID))
			}
		}
	default:
		//spss.Common.OnMsg(msg)
	}
	return true
}

//下一局
func (spss *SportSS510K) OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	//if int(spss.CurCompleteCount) >= spss.Rule.JuShu || spss.GetGameStatus() != public.GS_MJ_PLAY {
	//	return true
	//}
	//得分大于等于1000时，不处理此消息
	if spss.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}
	for i := 0; i < spss.GetPlayerCount(); i++ {
		if spss.PlayerTotalScore[i] >= 1000 {
			return true
		}
	}

	// 记录游戏开始时间
	spss.Common.GameBeginTime = time.Now()

	nChiarID := spss.GetChairByUid(msg.Id)
	//将该消息广播出去。游戏开始后，不用广播
	if spss.GameEndStatus != static.GS_MJ_PLAY {
		spss.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
		spss.SendUserStatus(int(nChiarID), static.US_READY) //把我的状态发给其他人
	}

	//SEND_TABLE_DATA(INVALID_CHAIR,SUB_C_GOON_NEXT_GAME,pDataBuffer);

	if nChiarID >= 0 && nChiarID < uint16(spss.GetPlayerCount()) {
		_item := spss.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < spss.GetPlayerCount(); i++ {
		item := spss.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			break
		}
		if i == spss.GetPlayerCount()-1 {
			spss.RepositTable(false) // 复位桌子
			spss.OnGameStart()
		}
	}
	return true
}

//托管
func (spss *SportSS510K) onUserTustee(msg *static.Msg_C_Trustee) bool {
	item := spss.GetUserItemByUid(msg.Id)
	if item == nil {
		return false
	}
	if item.CheckTRUST() == msg.Trustee {
		return true
	}
	//变量定义
	var tuoguan static.Msg_S_DG_Trustee
	tuoguan.ChairID = spss.GetChairByUid(msg.Id)
	tuoguan.Trustee = msg.Trustee
	//校验规则
	if tuoguan.ChairID < uint16(spss.GetPlayerCount()) && ((spss.GameState == meta2.GsPlay) || (spss.GameState == meta2.GsRoarPai) || (spss.GameState == meta2.GsNull)) || (spss.GameState == meta2.Gs4KingScore) {
		if tuoguan.Trustee == true && (spss.GameState != meta2.GsNull) {
			spss.TuoGuanPlayer[tuoguan.ChairID] = true
			spss.TrustCounts[tuoguan.ChairID]++
			item.ChangeTRUST(true)
			if tuoguan.ChairID == spss.Whoplay {
				//spss.DownTime = GetCPUTickCount()+spss.AutoOutTime;
				if int64(spss.LimitTime) > time.Now().Unix() {
					tuoguan.Overtime = spss.LimitTime - time.Now().Unix()
				}
				if time.Now().Unix()+int64(spss.AutoOutTime) < spss.LimitTime { // 如果只剩下托管出牌的时间了，就不重新算了，否则跟改为托管出牌的时间
					spss.setLimitedTime(int64(spss.AutoOutTime))
				}
			}
			spss.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			// 回放托管记录
			spss.addReplayOrder(tuoguan.ChairID, DG_REPLAY_OPT_TUOGUAN, 1, []int{})
		} else if tuoguan.Trustee == false {
			spss.TuoGuanPlayer[tuoguan.ChairID] = false
			item.ChangeTRUST(false)
			// 如果是当前的玩家，那么重新设置一下开始时间
			if tuoguan.ChairID == spss.Whoplay {
				//spss.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
				if time.Now().Unix() < spss.LimitTime { // 如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < spss.LimitTime
					spss.setLimitedTime(int64(spss.PlayTime) + spss.PowerStartTime - time.Now().Unix() + 1)
					tuoguan.Overtime = int64(spss.PlayTime) + spss.PowerStartTime - time.Now().Unix()
				}
			}

			//tuoguan.theTime = PlayTime-(now-nowTime);
			spss.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//回放增加托管记录
			spss.addReplayOrder(tuoguan.ChairID, DG_REPLAY_OPT_TUOGUAN, 0, []int{})
		} else {
			return false
		}
	} else {
		//详细日志
		LogStr := string("托管动作:游戏状态不正确 ")
		spss.OnWriteGameRecord(tuoguan.ChairID, LogStr)
		return false
	}
	return true
}

func (spss *SportSS510K) addTurnOverReplayOrder(chairId uint16, operation int, value int, curscore [meta2.MAX_PLAYER]int, totalscore [meta2.MAX_PLAYER]int, addscore bool) {
	var order SS510k_Replay_Order
	order.R_ChairId = chairId
	order.R_Opt = operation

	if value > 0 {
		var order_ext SS510k_Replay_Order_Ext
		order_ext.Ext_type = DG_EXT_GETSCORE
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	}

	if addscore {
		var curscore0 SS510k_Replay_Order_Ext
		curscore0.Ext_type = DG_EXT_CURSCORE
		curscore0.Ext_score = curscore
		order.R_Opt_Ext = append(order.R_Opt_Ext, curscore0)

		var totalscore0 SS510k_Replay_Order_Ext
		totalscore0.Ext_type = DG_EXT_TOTALSCORE
		totalscore0.Ext_score = totalscore
		order.R_Opt_Ext = append(order.R_Opt_Ext, totalscore0)
	}

	spss.ReplayRecord.R_Orders = append(spss.ReplayRecord.R_Orders, order)

}

func (spss *SportSS510K) addReplayOrder(chairId uint16, operation int, value int, values []int) {
	var order SS510k_Replay_Order
	order.R_ChairId = chairId
	order.R_Opt = operation

	if operation == DG_REPLAY_OPT_HOUPAI {
		var order_ext SS510k_Replay_Order_Ext
		order_ext.Ext_type = DG_EXT_HOUPAI
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == DG_REPLAY_OPT_END_HOUPAI {
		order.R_Value = append(order.R_Value, value)
	} else if operation == DG_REPLAY_OPT_OUTCARD {
		order.R_Value = append(order.R_Value, values[:]...)
	} else if operation == DG_REPLAY_OPT_END_GAME {

	} else if operation == DG_REPLAY_OPT_DIS_GAME {

	} else if operation == DG_REPLAY_OPT_TURN_OVER {
		var order_ext SS510k_Replay_Order_Ext
		order_ext.Ext_type = DG_EXT_GETSCORE
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == DG_REPLAY_OPT_TUOGUAN {
		var order_ext SS510k_Replay_Order_Ext
		order_ext.Ext_type = DG_EXT_TUOGUAN
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	} else if operation == DG_REPLAY_OPT_4KINGSCORE {
		order.R_Value = append(order.R_Value, values[:]...)

		var order_ext SS510k_Replay_Order_Ext
		order_ext.Ext_type = DG_EXT_4KINGSCORE
		order_ext.Ext_value = value
		order.R_Opt_Ext = append(order.R_Opt_Ext, order_ext)
	}

	spss.ReplayRecord.R_Orders = append(spss.ReplayRecord.R_Orders, order)
}

//用户出牌
func (spss *SportSS510K) OnUserOutCard(msg *static.Msg_C_DG_OutCard) bool {
	xlog.Logger().Info("OnUserOutCard")
	//效验状态
	if spss.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}
	if spss.GameEndStatus != static.GS_MJ_PLAY {
		return true
	}
	if spss.GameState != meta2.GsPlay {
		return false
	}
	wChairID := msg.CurrentUser
	//效验参数"tablecreate"
	if wChairID != spss.Whoplay {
		//详细日志
		LogStr := fmt.Sprintf("座位号 %d, OnUserOutCard 出牌玩家不是当前玩家 ", wChairID)
		spss.OnWriteGameRecord(wChairID, LogStr)
		return false
	}
	_userItem := spss.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	//出牌数目为0，则为放弃的情况
	if msg.CardCount == 0 {
		if spss.WhoLastOut < uint16(spss.GetPlayerCount()) && spss.Whoplay == wChairID {
			var outmsg static.Msg_S_DG_OutCard
			outmsg.CardCount = 0
			outmsg.CurrentUser = msg.CurrentUser
			for i := 0; i < static.MAX_CARD_4P; i++ {
				outmsg.CardData[i] = 0
				spss.AllPaiOut[spss.Whoplay][i] = 0
				spss.LastPaiOut[spss.Whoplay][i] = 0
			}
			outmsg.OutScorePai = spss.OutScorePai
			spss.SendTableMsg(consts.MsgTypeGameOutCard, outmsg)

			//回放增加出牌日志
			spss.addReplayOrder(spss.Whoplay, DG_REPLAY_OPT_OUTCARD, int(static.INVALID_BYTE), []int{})

			//详细日志
			LogStr := string("OnUserOutCard 玩家pass ")
			spss.OnWriteGameRecord(spss.Whoplay, LogStr)

			spss.WhoPass[spss.Whoplay] = true

			spss.GoNextPlayer()
		}

		return true
	}

	buf := [static.MAX_CARD_4P]byte{}
	for i := 0; i < static.MAX_CARD_4P; i++ {
		buf[i] = spss.PlayerCards[spss.Whoplay][i]
	}

	z := 0 //重新检验牌数目，并临时删除出的牌
	for i := byte(0); i < msg.CardCount; i++ {
		for j := 0; j < static.MAX_CARD_4P; j++ {
			if msg.CardData[i] == buf[j] {
				buf[j] = 0
				z++
				break
			}
		}
	}
	if byte(z) == msg.CardCount {
		var re1, re2 static.TCardType
		iNumOfKing := 0
		for i := 0; i < static.MAX_CARD_4P; i++ {
			spss.AllPaiOut[spss.Whoplay][i] = msg.CardData[i]
		}

		//TYPE_NULL代表第一个出
		if spss.WhoLastOut >= uint16(spss.GetPlayerCount()) {
			re1.Cardtype = TYPE_NULL
			spss.LastOutType = TYPE_NULL
			spss.LastOutTypeClient = TYPE_NULL
		} else {
			l := spss.m_GameLogic.GetCardNum(spss.AllPaiOut[spss.WhoLastOut], static.MAX_CARD_4P)
			re1 = spss.m_GameLogic.GetType(spss.AllPaiOut[spss.WhoLastOut], int(l), 0, 0, 0)
			//这里为什么要重新设置下呢？因为很可能出现不同的判断
			//比如A走了334455,B走了44王王王5,那么当C走的时候，GetType很可能把B的牌型判断为444555这样的大牌型
			//所以，在这里，一定要把牌型设置回来，根据客户端传过来的为依据
			re1.Cardtype = spss.LastOutType
		}
		re2 = spss.m_GameLogic.GetType(spss.AllPaiOut[spss.Whoplay], int(msg.CardCount), 0, 0, spss.LastOutType)
		iNumOfKing = spss.m_GameLogic.GetKingNum(spss.AllPaiOut[spss.Whoplay], int(msg.CardCount))

		//详细日志/////////
		//		LogStr := fmt.Sprintf("re1的类型%d ,card=%d ", re1.Cardtype, re1.Card)
		//		LogStr += fmt.Sprintf("re2的类型%d ,card=%d ", re2.Cardtype, re2.Card)
		//		spss.OnWriteGameRecord(spss.Whoplay, LogStr)

		spss.WhoPass[spss.Whoplay] = false
		if spss.m_GameLogic.Compare(re1, re2) {
			//回放增加出牌日志
			spss.addReplayOrder(spss.Whoplay, DG_REPLAY_OPT_OUTCARD, int(static.INVALID_BYTE), static.HF_BytesToInts(msg.CardData[:]))
			// 详细日志 by sam  打出510K或者7喜或者8喜,管得起才算
			if re2.Cardtype == TYPE_BOMB_510K {
				spss.Play510K[spss.Whoplay]++
			}

			//管得起才能算分
			if spss.GameType == meta2.GT_NORMAL {
				spss.CardScore += spss.m_GameLogic.GetScore(spss.AllPaiOut[spss.Whoplay], int(msg.CardCount))
				spss.SendTurnScore(static.MAX_PLAYER_4P)
				spss.ReplayRecord.R_Orders[len(spss.ReplayRecord.R_Orders)-1].AddReplayExtData(DG_EXT_TURNSCORE, spss.CardScore)
			} else {
				spss.CardScore = 0
			}

			//回放分牌
			recordScorePai := []int{}

			//加入分牌
			index := 24
			for i := 0; i < 24; i++ {
				if spss.OutScorePai[i] == 0 {
					index = i
					break
				}
			}
			for i := byte(0); i < msg.CardCount; i++ {
				outcard := spss.AllPaiOut[spss.Whoplay][i]
				if spss.m_GameLogic.isScorePai(outcard) {
					if index >= 24 {
						spss.OnWriteGameRecord(wChairID, "分数牌最多24张")
						break
					}
					recordScorePai = append(recordScorePai, int(outcard))
					spss.OutScorePai[index] = outcard
					index++
				}
			}

			if len(recordScorePai) > 0 {
				spss.ReplayRecord.R_Orders[len(spss.ReplayRecord.R_Orders)-1].R_ScoreCard = recordScorePai
			}

			spss.LastOutType = re2.Cardtype
			spss.LastOutTypeClient = msg.CardType

			spss.ThePaiCount[spss.Whoplay] -= msg.CardCount
			spss.WhoLastOut = spss.Whoplay

			z := 0 //计算剩余多少张牌 m_thePaiCount可能不准
			for i := 0; i < static.MAX_CARD_4P; i++ {
				spss.PlayerCards[spss.Whoplay][i] = buf[i]
				if buf[i] > 0 {
					z++
				}
			}

			// 发送手牌数量
			if spss.ShowHandCardCnt {
				spss.SendPaiCount(static.MAX_PLAYER_4P)
			}

			var msgout static.Msg_S_DG_OutCard
			msgout.CardCount = msg.CardCount
			msgout.CardType = msg.CardType
			msgout.CurrentUser = msg.CurrentUser
			msgout.Overtime = 15
			msgout.ByClient = false
			msgout.CardData = msg.CardData
			msgout.OutScorePai = spss.OutScorePai
			spss.SendTableMsg(consts.MsgTypeGameOutCard, msgout)

			//插底提示
			if z == 1 {
				spss.SendPlaySoundMsg(spss.Whoplay, static.TY_ChaDi)
			}

			//判断任务是否完成
			spss.GameTask.IsTaskFinished(spss.Common, re2, spss.Whoplay, iNumOfKing)
			//end
			spss.GameOverOrNextPlayer(z, re2, iNumOfKing)

		} else {
			spss.OnWriteGameRecord(spss.Whoplay, "牌型不正确")
			spss.WhoPass[spss.Whoplay] = true
			return false
		}
	}

	return true
}

//游戏是否可以结束或者下一个玩家出牌
func (spss *SportSS510K) GameOverOrNextPlayer(byLeftCardNum int, re static.TCardType, iNumOfKing int) int {
	//硬牌，有一个没有牌了就结束游戏
	if spss.GameType == meta2.GT_ROAR || spss.GetPlayerCount() == static.MAX_PLAYER_2P {
		//结束游戏
		if byLeftCardNum == 0 {
			//判断最后一手的任务是否完成
			spss.GameTask.IsTaskFinishedOfLastHand(spss.Common, re, spss.Whoplay, iNumOfKing)
			//end

			spss.Nextbanker = spss.Whoplay
			spss.PlayerCardScore[spss.Whoplay] += spss.CardScore
			spss.SendPlayerScore(static.MAX_PLAYER_4P, spss.Whoplay, spss.CardScore, false)
			spss.OnGameEndNormal(spss.Whoplay, meta2.GOT_NORMAL)
			return 1
		} else {
			spss.GoNextPlayer()
		}
	} else if spss.GameType == meta2.GT_NORMAL {
		//不硬牌，找朋友的模式
		//m_whoplay的牌没有了，那么检查下我的对家结束没有
		if byLeftCardNum == 0 {
			//判断最后一手的任务是否完成
			spss.GameTask.IsTaskFinishedOfLastHand(spss.Common, re, spss.Whoplay, iNumOfKing)
			//end

			teamer := spss.GetTeamer(spss.Whoplay)
			spss.BTeamOut[teamer] = true

			spss.WhoAllOutted[spss.Whoplay] = true
			spss.AllOutCnt++
			spss.PlayerTurn[spss.Whoplay] = spss.AllOutCnt

			spss.SendPlayerTurn(static.MAX_PLAYER_4P)
			spss.SendPlaySoundMsg(spss.Whoplay, static.TY_AllOut) //牌出完了需要客户端加特效

			if teamer < static.MAX_PLAYER_4P {
				//我的对家走完，那么结束游戏
				if spss.WhoAllOutted[teamer] {
					//接着检测是不是双扣
					spss.WhoAllOutted[spss.Whoplay] = true
					ncount := 0
					for i := 0; i < static.MAX_PLAYER_4P; i++ {
						if !spss.WhoAllOutted[i] {
							ncount++
						}
					}
					//双扣
					if ncount == 2 {
						for i := 0; i < static.MAX_PLAYER_4P; i++ {
							if spss.PlayerTurn[i] == 1 {
								spss.Nextbanker = uint16(i)
								break
							}
						}

						spss.PlayerCardScore[spss.Whoplay] += spss.CardScore
						spss.SendPlayerScore(static.MAX_PLAYER_4P, spss.Whoplay, spss.CardScore, false)
						spss.OnGameEndNormal(spss.Whoplay, meta2.GOT_DOUBLEKILL)
						return 1
					} else if ncount == 1 {
						//3家结束
						for i := 0; i < static.MAX_PLAYER_4P; i++ {
							if spss.PlayerTurn[i] == 1 {
								spss.Nextbanker = uint16(i)
								break
							}
						}
						spss.PlayerCardScore[spss.Whoplay] += spss.CardScore
						spss.SendPlayerScore(static.MAX_PLAYER_4P, spss.Whoplay, spss.CardScore, false)
						spss.OnGameEndNormal(spss.Whoplay, meta2.GOT_NORMAL)
						return 1
					} else { //可能吗？如果是其他情况，那么逻辑就错了！！

					}
					//这里检查下，是否结束游戏
					//if spss.JudgeEndGame(spss.Whoplay) {
					//	return 1
					//}
				} else //接风
				{
					//这里检查下，是否结束游戏
					//if spss.JudgeEndGame(spss.Whoplay) {
					//	return 1
					//}
					//SendPlaySoundMsg(Whoplay,msgPlaySound::TY_JieFeng);
					spss.SendPaiToTeamer(teamer, spss.Whoplay)

					spss.GoNextPlayer()
				}
			}

		} else {
			if spss.BTeamOut[spss.Whoplay] && spss.SeeTeamerCard {
				teamer := spss.GetTeamer(spss.Whoplay)
				spss.SendPaiToTeamer(spss.Whoplay, teamer)
			}
			spss.GoNextPlayer()
		}
	}
	return 1
}

func (spss *SportSS510K) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	if wChairID >= uint16(spss.GetPlayerCount()) {
		return false
	}
	if cbReason != meta2.GOT_NORMAL && cbReason != meta2.GOT_DOUBLEKILL && cbReason != meta2.GOT_ZHONGTU {
		return false
	}
	if spss.GameType == meta2.GT_NORMAL {
		spss.CardScore = 0
		spss.SendTurnScore(static.MAX_PLAYER_4P)
	}
	spss.GameEndStatus = static.GS_MJ_END

	//定义变量
	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_NORMAL
	endgameMsg.TheOrder = spss.CurCompleteCount
	endgameMsg.WhoKingBomb = static.INVALID_CHAIR //4王一起出了才算
	if spss.WhoHasKingBomb >= 0 && spss.WhoHasKingBomb < static.MAX_PLAYER_4P {
		if spss.PlayKingBomb[spss.WhoHasKingBomb] > 0 {
			endgameMsg.WhoKingBomb = spss.WhoHasKingBomb
		}
	}

	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		endgameMsg.Have7Xi[i] = spss.Who7Xi[i]
		endgameMsg.Have8Xi[i] = spss.Who8Xi[i]
		//endgameMsg.Have510K[i] = spss.WhoSame510K[i]
		endgameMsg.FaScore[i] = spss.BombScore[i] // 炸弹分
	}

	//fullScoreAward := 0								  //满分奖励
	//fourKingScore := 0								  //4王换的分
	//addDiFen := 0									  //保底分数
	gongfen := 0 //贡分

	player1 := uint16(0)
	player2 := uint16(0)
	player3 := uint16(0)
	player4 := uint16(0)

	//二人场
	if spss.GetPlayerCount() == static.MAX_PLAYER_2P {
		for i := uint16(0); i < static.MAX_PLAYER_2P; i++ {
			if wChairID == i {
				spss.PlayerTurn[i] = 1
				player1 = i
			} else {
				player2 = i
			}
		}

		//2游给1游100/200贡献分
		if spss.Rule.ShareScoreType == 0 {
			gongfen = 100
			xlog.Logger().Info(fmt.Sprintf("贡分选的100/200,需要贡分%d", gongfen))
		} else if spss.Rule.ShareScoreType == 1 {
			gongfen = 200
			xlog.Logger().Info(fmt.Sprintf("贡分选的200/400,需要贡分%d", gongfen))
		}

		//2游手上没跑掉的分需要给1游
		player2cardscore := spss.m_GameLogic.GetScore(spss.PlayerCards[player2], static.MAX_CARD_4P)
		spss.PlayerCardScore[player1] += player2cardscore

		endgameMsg.GongFen[player1] += gongfen
		spss.PlayerScore[player1] += player2cardscore + gongfen
		spss.PlayerTotalScore[player1] += player2cardscore + gongfen
		spss.PlayerScore[player2] -= gongfen
		endgameMsg.GongFen[player2] -= gongfen
		spss.PlayerTotalScore[player2] -= gongfen

		if spss.Banker == player1 {
			xlog.Logger().Info(fmt.Sprintf("庄家(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.PlayerScore[player1], spss.PlayerScore[player2]))
			xlog.Logger().Info(fmt.Sprintf("庄家(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.PlayerTotalScore[player1], spss.PlayerTotalScore[player2]))
		} else {
			xlog.Logger().Info(fmt.Sprintf("庄家(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.PlayerScore[player2], spss.PlayerScore[player1]))
			xlog.Logger().Info(fmt.Sprintf("庄家(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.PlayerTotalScore[player2], spss.PlayerTotalScore[player1]))
		}

	} else if spss.GetPlayerCount() == static.MAX_PLAYER_4P {
		//四人场
		//处理没有给最后一个玩家赋值名次号的数据处理
		for i := 0; i < spss.GetPlayerCount(); i++ {
			//若名次号还是INVALID_BYTE，则表示还是初始化的数据，需要重新赋值
			if spss.PlayerTurn[i] >= static.MAX_PLAYER_4P {
				spss.AllOutCnt++
				spss.PlayerTurn[i] = spss.AllOutCnt
			}
		}
		//
		for i := uint16(0); i < static.MAX_PLAYER_4P; i++ {
			if spss.PlayerTurn[i] == 1 {
				player1 = i
			}
			if spss.PlayerTurn[i] == 2 {
				player2 = i
			}
			if spss.PlayerTurn[i] == 3 {
				player3 = i
			}
			if spss.PlayerTurn[i] == 4 {
				player4 = i
			}
		}

		player1parter := spss.GetTeamer(player1) //1游的队友

		//12vs34,双下的情况
		if player1parter == player2 {
			//12对34情况,34游需要给12游贡60分
			if spss.Rule.ShareScoreType == 0 {
				gongfen = 200
				xlog.Logger().Info(fmt.Sprintf("贡分选的100/200,一二游对三四游,需要贡分%d", gongfen))
			} else if spss.Rule.ShareScoreType == 1 {
				gongfen = 400
				xlog.Logger().Info(fmt.Sprintf("贡分选的200/400,一二游对三四游,需要贡分%d", gongfen))
			}

			//34游手上没跑掉的分需要给1游
			player3cardscore := spss.m_GameLogic.GetScore(spss.PlayerCards[player3], static.MAX_CARD_4P)
			player4cardscore := spss.m_GameLogic.GetScore(spss.PlayerCards[player4], static.MAX_CARD_4P)
			otherscardscore := player3cardscore + player4cardscore
			spss.PlayerCardScore[player1] += otherscardscore

			xlog.Logger().Info(fmt.Sprintf("34游玩家贡分%d给12游玩家队伍", gongfen))
			otherscardscore += gongfen
			endgameMsg.GongFen[player1] += gongfen
			spss.PlayerScore[player1] += otherscardscore
			spss.PlayerTotalScore[player1] += otherscardscore
			spss.PlayerScore[player2] = spss.PlayerScore[player1]
			spss.PlayerTotalScore[player2] = spss.PlayerTotalScore[player1]
			spss.PlayerScore[player3] -= gongfen
			endgameMsg.GongFen[player3] -= gongfen
			spss.PlayerTotalScore[player3] -= gongfen
			spss.PlayerScore[player4] = spss.PlayerScore[player3]
			spss.PlayerTotalScore[player4] = spss.PlayerTotalScore[player3]

			if spss.Banker == player1 || spss.Banker == player2 {
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.BankParter, spss.PlayerScore[player1], spss.PlayerScore[player3]))
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.BankParter, spss.PlayerTotalScore[player1], spss.PlayerTotalScore[player3]))
			} else {
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.BankParter, spss.PlayerScore[player3], spss.PlayerScore[player1]))
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.BankParter, spss.PlayerTotalScore[player3], spss.PlayerTotalScore[player1]))
			}

		} else if player1parter == player3 {
			//13对24情况
			if spss.Rule.ShareScoreType == 0 {
				gongfen = 100
				xlog.Logger().Info(fmt.Sprintf("贡分选的100/200,一三游对二四游,需要贡分%d", gongfen))
			} else if spss.Rule.ShareScoreType == 1 {
				gongfen = 200
				xlog.Logger().Info(fmt.Sprintf("贡分选的200/400,一三游对二四游,需要贡分%d", gongfen))
			}

			//4游手上没跑掉的分需要给3游
			player4cardscore := spss.m_GameLogic.GetScore(spss.PlayerCards[player4], static.MAX_CARD_4P)
			otherscardscore := player4cardscore
			spss.PlayerCardScore[player3] += otherscardscore

			xlog.Logger().Info(fmt.Sprintf("24游玩家贡分%d给13游玩家队伍", gongfen))
			otherscardscore += gongfen
			endgameMsg.GongFen[player1] += gongfen
			spss.PlayerScore[player1] += otherscardscore
			spss.PlayerTotalScore[player1] += otherscardscore
			spss.PlayerScore[player3] = spss.PlayerScore[player1]
			spss.PlayerTotalScore[player3] = spss.PlayerTotalScore[player1]
			endgameMsg.GongFen[player2] -= gongfen
			spss.PlayerScore[player2] -= gongfen
			spss.PlayerTotalScore[player2] -= gongfen
			spss.PlayerScore[player4] = spss.PlayerScore[player2]
			spss.PlayerTotalScore[player4] = spss.PlayerTotalScore[player2]

			if spss.Banker == player1 || spss.Banker == player3 {
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.BankParter, spss.PlayerScore[player1], spss.PlayerScore[player2]))
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.BankParter, spss.PlayerTotalScore[player1], spss.PlayerTotalScore[player2]))
			} else {
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.BankParter, spss.PlayerScore[player2], spss.PlayerScore[player1]))
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.BankParter, spss.PlayerTotalScore[player2], spss.PlayerTotalScore[player1]))
			}
		} else {
			//14对23情况
			//4游手上没跑掉的分需要给3游
			player4cardscore := spss.m_GameLogic.GetScore(spss.PlayerCards[player4], static.MAX_CARD_4P)
			otherscardscore := player4cardscore
			spss.PlayerCardScore[player3] += otherscardscore

			xlog.Logger().Info(fmt.Sprintf("14游玩家贡分%d给23游玩家队伍", gongfen))
			spss.PlayerScore[player3] += otherscardscore
			spss.PlayerTotalScore[player3] += otherscardscore
			spss.PlayerScore[player2] = spss.PlayerScore[player3]
			spss.PlayerTotalScore[player2] = spss.PlayerTotalScore[player3]

			if spss.Banker == player1 || spss.Banker == player4 {
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.BankParter, spss.PlayerScore[player1], spss.PlayerScore[player2]))
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.BankParter, spss.PlayerTotalScore[player1], spss.PlayerTotalScore[player2]))
			} else {
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)当局积分%d,闲家当局积分%d", spss.Banker, spss.BankParter, spss.PlayerScore[player2], spss.PlayerScore[player1]))
				xlog.Logger().Info(fmt.Sprintf("庄家(%d)和庄家队友(%d)累计积分%d,闲家累计积分%d", spss.Banker, spss.BankParter, spss.PlayerTotalScore[player2], spss.PlayerTotalScore[player1]))
			}
		}
	}

	//判断赢家
	winteam := 0
	if spss.PlayerScore[1] > spss.PlayerScore[0] {
		winteam = 1
	} else if spss.PlayerScore[0] == spss.PlayerScore[1] {
		winteam = -1 //平局
	}
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		if winteam == -1 {
			endgameMsg.WinLose[i] = 2
		} else if winteam == int(i) || winteam == int(spss.GetTeamer(i)) {
			endgameMsg.WinLose[i] = 1
		} else {
			endgameMsg.WinLose[i] = 0
		}
		endgameMsg.Score[i] = 0
		endgameMsg.PlayerScore[i] = spss.PlayerScore[i]
		endgameMsg.TotalPlayerScore[i] = spss.PlayerTotalScore[i]
	}

	//结束游戏标志
	checkEndFlag := false
	winPlayer := -1
	for i := 0; i < spss.GetPlayerCount(); i++ {
		if spss.PlayerTotalScore[i] >= 1000 {
			checkEndFlag = true
			winPlayer = i
			break
		}
	}
	var isBalanceEndFlag bool
	//小局结束时，有一方玩家达到1000分
	if checkEndFlag {
		//获取输家
		losePlayer := (winPlayer%2 + 1) % 2
		if spss.PlayerTotalScore[winPlayer] < spss.PlayerTotalScore[losePlayer] {
			winPlayer, losePlayer = losePlayer, winPlayer
		} else if spss.PlayerTotalScore[winPlayer] == spss.PlayerTotalScore[losePlayer] {
			isBalanceEndFlag = true
		}
		//计算级数
		winscore := spss.PlayerTotalScore[winPlayer]
		losescore := spss.PlayerTotalScore[losePlayer]
		jishu := spss.m_GameLogic.DealScore(winscore) - spss.m_GameLogic.DealScore(losescore)
		_totalscore := spss.IBase*jishu + spss.AddDiFen

		spss.JiCount = jishu

		if !isBalanceEndFlag {
			winCount := 0
			for i := 0; i < spss.GetPlayerCount(); i++ {
				if spss.PlayerTotalScore[winPlayer] == spss.PlayerTotalScore[i] {
					endgameMsg.Score[i] = _totalscore
					winCount++
				} else {
					endgameMsg.Score[i] = -_totalscore
				}
			}
			if winCount >= spss.GetPlayerCount() {
				isBalanceEndFlag = true
				endgameMsg.Score = [4]int{}
			}
		}
	}

	//胜局次数统计、计算保底得分
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		_userItem := spss.GetUserItemByChair(uint16(i))
		if endgameMsg.WinLose[i] == 1 && _userItem != nil {
			//spss.WinCount[i]++
			// 记录在玩家身上
			_userItem.Ctx.WinCount++
		}
	}

	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		cn := 0
		for j := 0; j < static.MAX_CARD_4P; j++ {
			if spss.m_GameLogic.IsValidCard(spss.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spss.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = spss.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)
		//spss.LastScore[i] = endgameMsg.Score[i]
		//spss.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = spss.GetUserTotalScore(i) + endgameMsg.Score[i]

		if spss.MaxScore[i] < spss.PlayerCardScore[i] {
			spss.MaxScore[i] = spss.PlayerCardScore[i]
		}
		//一游次数
		if spss.PlayerTurn[i] == 1 {
			//spss.TotalFirstTurn[i]++
			// 记录在玩家身上
			_userItem := spss.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				_userItem.Ctx.PlayTurn1st++
			}
		}

		endgameMsg.TheTurn[i] = spss.PlayerTurn[i]
		endgameMsg.GetScore[i] = spss.PlayerCardScore[i]
	}
	if spss.WhoRoar < meta2.MAX_PLAYER {
		spss.TotalDuPai[spss.WhoRoar]++
	}

	banker, bankParter := spss.GetBankerForGameEnd()
	endgameMsg.HouPaiChair = spss.WhoRoar
	endgameMsg.TheBank = banker
	endgameMsg.TheParter = bankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", endgameMsg.Score[i], endgameMsg.TotalScore[i], spss.PlayerCardScore[i])
		spss.OnWriteGameRecord(i, recrodStr)
	}

	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		endgameMsg.XiScore[i] = spss.XiScore[i] //喜钱
	}

	if spss.TrustPunish && checkEndFlag && !isBalanceEndFlag && spss.GetPlayerCount() == static.MAX_PLAYER_4P {
		// 检查托管罚分
		for i := 0; i < static.MAX_PLAYER_4P; i++ {
			_userItem := spss.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				if _userItem.Ctx.HasTrustee {
					result := endgameMsg.Score[i]
					spss.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("托管过，最终输赢分: %d(小于0则触发托管罚分)", result))
					if result < 0 {
						teamerChair := spss.GetTeamer(_userItem.GetChairID())
						teamer := spss.GetUserItemByChair(teamerChair)
						if teamer != nil {
							teamerResult := endgameMsg.Score[teamerChair]
							if teamerResult < 0 {
								if teamer.Ctx.HasTrustee {
									spss.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("触发托管罚分，不过他的队友%d[%s:%d]也托管了", teamerChair, teamer.Name, teamer.Uid))
								} else {
									endgameMsg.Score[i] += teamerResult
									endgameMsg.Score[teamerChair] = 0
									spss.OnWriteGameRecord(_userItem.GetChairID(),
										fmt.Sprintf("触发托管罚分，承包了他的队友%d[%s:%d]的输分:%d,自己的输分:%d, 合计:%d",
											teamerChair, teamer.Name, teamer.Uid, teamerResult, result, endgameMsg.Score[i]))
								}
							} else {
								spss.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("触发托管罚分，不过他的队友%d[%s:%d]的分数[%d]异常", teamerChair, teamer.Name, teamer.Uid, teamerResult))
							}
						} else {
							spss.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("触发托管罚分，但是队友%d没找到", teamerChair))
						}
					} else {
						spss.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("触发托管罚分，但是他的成绩为:%d，队伍没输钱", result))
					}
				}
			}
		}
	}

	// 调用结算接口
	_, endgameMsg.UserVitamin = spss.OnSettle(endgameMsg.Score, consts.EventSettleGameOver)

	//发送信息
	spss.VecGameEnd = append(spss.VecGameEnd, endgameMsg) //保存，用于汇总计算
	spss.SaveGameData()
	spss.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	spss.OnWriteGameRecord(static.INVALID_CHAIR, "游戏正常结束OnGameEndNormal()......")

	//回放增加结束数据
	spss.addReplayOrder(wChairID, DG_REPLAY_OPT_END_GAME, 0, []int{})

	spss.ReplayRecord.EndInfo = &endgameMsg
	// 数据库写出牌记录	// 写完后清除数据
	spss.TableWriteOutDate()

	for _, v := range spss.PlayerInfo { //i := 0; i < spss.GetPlayerCount(); i++ {
		wintype := static.ScoreKind_Draw
		//if endgameMsg.Score[v.Seat] > 0 {
		if endgameMsg.WinLose[v.Seat] == 1 {
			wintype = static.ScoreKind_Win //enScoreKind_Win;
		} else {
			wintype = static.ScoreKind_Lost //enScoreKind_Lost;
		}
		//iAward := spss.GetTaskAward(v.Seat)//金豆任务，先留着备用
		// 记录分数用于总结算汇总
		v.Ctx.GameScoreFen = append(v.Ctx.GameScoreFen, endgameMsg.Score[v.Seat])
		spss.TableWriteGameDate(int(spss.CurCompleteCount), v, wintype, endgameMsg.Score[v.Seat])
	}

	//扣房卡
	//if spss.CurCompleteCount == 1 {
	//	spss.TableDeleteFangKa(spss.CurCompleteCount)
	//}

	//结束游戏
	if checkEndFlag { //积分够了
		// 设置游戏结束状态
		spss.SetGameStatus(static.GS_MJ_END)

		spss.CalculateResultTotal(static.GER_NORMAL, wChairID, 0, banker, bankParter) //计算总发送总结算

		spss.UpdateOtherFriendDate(&endgameMsg, false)
		//通知框架结束游戏
		//spss.SetGameStatus(public.GS_MJ_FREE)
		spss.ConcludeGame()

	} else {
		check := false
		if spss.Rule.Overtime_dismiss != -1 {
			for i := 0; i < spss.GetPlayerCount(); i++ {
				item := spss.GetUserItemByChair(uint16(i))
				if item == nil || !item.CheckTRUST() {
					continue
				}
				if check {
					var _msg = &static.Msg_C_DismissFriendResult{
						Id:   item.Uid,
						Flag: true,
					}
					spss.OnDismissResult(item.Uid, _msg)
				} else {
					check = true
					var msg = &static.Msg_C_DismissFriendReq{
						Id: item.Uid,
					}
					spss.SetDismissRoomTime(spss.Rule.Overtime_dismiss)
					spss.OnDismissFriendMsg(item.Uid, msg)

				}
			}
		}
	}

	// 1、第一局随机坐庄；2、胡牌或流局连庄；3、第一局无人胡牌则下家坐庄；
	if spss.BankerUser != static.INVALID_CHAIR {
		spss.BankerUser = spss.Nextbanker
		spss.Banker = spss.Nextbanker
	}

	spss.OnGameEnd()
	spss.RepositTable(false) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	// OnGameEnd 会清除计时器
	if !checkEndFlag /*&& spss.Rule.Overtime_trust > 0*/ {
		spss.SetAutoNextTimer(15) // 自动开始下一局
	}

	return true
}

func (spss *SportSS510K) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	if wChairID >= static.MAX_PLAYER_4P {
		return false
	}
	if cbReason != meta2.GOT_ESCAPE && cbReason != meta2.GOT_TUOGUAN {
		return false
	}
	if spss.GameType == meta2.GT_NORMAL {
		spss.CardScore = 0
		spss.SendTurnScore(static.MAX_PLAYER_4P)
	}
	spss.GameEndStatus = static.GS_MJ_END

	//定义变量
	iScore := [meta2.MAX_PLAYER]int{}
	spss.XiScore = [meta2.MAX_PLAYER]int{}

	byGongType := byte(0)
	byGongType = static.G_Bangong

	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_USER_LEFT
	endgameMsg.TheOrder = spss.CurCompleteCount
	endgameMsg.WhoKingBomb = static.INVALID_CHAIR //4王一起出了才算
	if spss.WhoHasKingBomb >= 0 && spss.WhoHasKingBomb < static.MAX_PLAYER_4P {
		if spss.PlayKingBomb[spss.WhoHasKingBomb] > 0 {
			endgameMsg.WhoKingBomb = spss.WhoHasKingBomb
		}
	}

	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		endgameMsg.Have7Xi[i] = spss.Who7Xi[i]
		endgameMsg.Have8Xi[i] = spss.Who8Xi[i]
		endgameMsg.Have510K[i] = spss.WhoSame510K[i]
		endgameMsg.PlayerScore[i] = spss.PlayerScore[i]
		endgameMsg.TotalPlayerScore[i] = spss.PlayerTotalScore[i]
		endgameMsg.FaScore[i] = spss.BombScore[i]
	}

	endgameMsg.EndType = static.TY_ESCAPE
	banker, bankParter := spss.GetBankerForGameEnd()
	endgameMsg.TheBank = banker
	endgameMsg.TheParter = bankParter

	nFan := 0
	if spss.GameState == meta2.GsRoarPai {
		nFan = spss.FaOfTao * 2
	} else {
		nFan = spss.FaOfTao
	}
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		if i == wChairID {
			iScore[i] = -spss.IBase * nFan
			endgameMsg.WinLose[i] = 0
		} else {
			if spss.GameState == meta2.GsRoarPai {
				iScore[i] = spss.IBase * (spss.JiangOfTao) * 2
			} else {
				iScore[i] = spss.IBase * (spss.JiangOfTao)
			}
			endgameMsg.WinLose[i] = 1
		}
	}
	for i := uint16(0); i < static.MAX_PLAYER_4P; i++ {
		spss.AddSpecailScore(&iScore, i, spss.IBase)
	}

	endgameMsg.FanShu = 1
	endgameMsg.GongType = byGongType
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		endgameMsg.Score[i] = iScore[i] - spss.Spay
		cn := 0
		for j := 0; j < static.MAX_CARD_4P; j++ {
			if spss.m_GameLogic.IsValidCard(spss.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spss.PlayerCards[i][j]
				cn++
			}
		}
		//spss.LastScore[i] = endgameMsg.Score[i]
		//spss.Total[i] += endgameMsg.Score[i]
	}

	//游戏记录
	spss.OnWriteGameRecord(wChairID, "强退游戏结束")

	//记录异常结束数据
	//spss.addReplayOrder(wChairID, E_Li_Xian, 0)

	//回放增加结束数据
	spss.addReplayOrder(wChairID, DG_REPLAY_OPT_END_GAME, 0, []int{})

	spss.ReplayRecord.EndInfo = &endgameMsg
	//写入游戏回放数据,写完重置当前回放数据
	spss.TableWriteOutDate()

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d", endgameMsg.Score[i], spss.GetUserTotalScore(uint16(i))+endgameMsg.Score[i])
		spss.OnWriteGameRecord(i, recrodStr)
	}

	//发送信息
	spss.VecGameEnd = append(spss.VecGameEnd, endgameMsg) //保存，用于汇总计算
	spss.SaveGameData()
	spss.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)
	//数据库写分
	for _, v := range spss.PlayerInfo { //i := 0; i < spss.GetPlayerCount(); i++ {
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
		//iAward := spss.GetTaskAward(v.Seat)//金豆任务，先留着备用
		spss.TableWriteGameDate(int(spss.CurCompleteCount), v, wintype, endgameMsg.Score[v.Seat])
	}

	spss.UpdateOtherFriendDate(&endgameMsg, true)
	//结束游戏
	spss.CalculateResultTotal(static.GER_USER_LEFT, wChairID, 0, banker, bankParter)
	//结束游戏 不重置局数
	spss.CurCompleteCount = 0
	//spss.SetGameStatus(public.GS_MJ_FREE)
	spss.ConcludeGame()
	spss.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

//解散，结束游戏
func (spss *SportSS510K) OnGameEndDissmiss(wChairID uint16, cbReason byte) bool {
	//if spss.Rule.HasPao && spss.PayPaoStatus {
	//	spss.CurCompleteCount--
	//}
	//变量定义
	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_DISMISS
	endgameMsg.TheOrder = spss.CurCompleteCount

	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		endgameMsg.Score[i] = endgameMsg.Score[i] - spss.Spay
		cn := 0
		for j := 0; j < meta2.TS_MAXHANDCARD; j++ {
			if spss.m_GameLogic.IsValidCard(spss.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spss.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = spss.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)
		//spss.LastScore[i] = endgameMsg.Score[i]
		//spss.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = spss.GetUserTotalScore(i) + endgameMsg.Score[i]
		endgameMsg.FaScore[i] = spss.BombScore[i]

		if spss.MaxScore[i] < spss.PlayerCardScore[i] {
			spss.MaxScore[i] = spss.PlayerCardScore[i]
		}

		if spss.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			//spss.TotalFirstTurn[i]++
			// 记录在玩家身上
			_userItem := spss.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				_userItem.Ctx.PlayTurn1st++
			}
		}

		endgameMsg.TheTurn[i] = spss.PlayerTurn[i]
		endgameMsg.GetScore[i] = spss.PlayerCardScore[i]

		endgameMsg.PlayerScore[i] = spss.PlayerScore[i]
		endgameMsg.TotalPlayerScore[i] = spss.PlayerTotalScore[i]
	}
	if spss.WhoRoar < meta2.MAX_PLAYER {
		spss.TotalDuPai[spss.WhoRoar]++
	}

	banker, bankParter := spss.GetBankerForGameEnd()
	endgameMsg.HouPaiChair = spss.WhoRoar
	endgameMsg.TheBank = banker
	endgameMsg.TheParter = bankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", endgameMsg.Score[i], endgameMsg.TotalScore[i], spss.PlayerCardScore[i])
		spss.OnWriteGameRecord(i, recrodStr)
	}
	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		endgameMsg.XiScore[i] = spss.XiScore[i] //喜钱
	}

	//记录异常结束数据
	for i := 0; i < spss.GetPlayerCount(); i++ {
		_item := spss.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
			//spss.addReplayOrder(uint16(i), E_Li_Xian, 0)
		}
	}

	//回放增加结束数据
	spss.addReplayOrder(wChairID, DG_REPLAY_OPT_DIS_GAME, 0, []int{})

	spss.ReplayRecord.EndInfo = &endgameMsg
	//写入游戏回放数据,写完重置当前回放数据
	spss.TableWriteOutDate()

	//发送信息
	spss.VecGameEnd = append(spss.VecGameEnd, endgameMsg) //保存，用于汇总计算
	spss.SaveGameData()
	spss.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	//游戏记录
	spss.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

	//数据库写分
	for _, v := range spss.PlayerInfo {
		if v.Seat != static.INVALID_CHAIR {
			if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
				spss.TableWriteGameDate(int(spss.CurCompleteCount), v, static.ScoreKind_pass, endgameMsg.Score[v.Seat])
			} else {
				spss.TableWriteGameDate(int(spss.CurCompleteCount), v, static.ScoreKind_pass, endgameMsg.Score[v.Seat])
			}
		}

	}

	spss.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	spss.CalculateResultTotal(static.GER_DISMISS, wChairID, 0, banker, bankParter)
	spss.GameEndStatus = static.GS_MJ_END
	//结束游戏
	//spss.SetGameStatus(public.GS_MJ_FREE)
	spss.ConcludeGame()

	spss.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

//程序异常，解散游戏
func (spss *SportSS510K) OnGameEndErrorDissmiss(wChairID uint16, cbReason byte) bool {
	//if spss.Rule.HasPao && spss.PayPaoStatus {
	//	spss.CurCompleteCount--
	//}
	//变量定义
	var endgameMsg static.Msg_S_DG_GameEnd
	endgameMsg.EndStatus = static.GER_DISMISS
	endgameMsg.TheOrder = spss.CurCompleteCount

	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		endgameMsg.Score[i] = endgameMsg.Score[i] - spss.Spay
		cn := 0
		for j := 0; j < meta2.TS_MAXHANDCARD; j++ {
			if spss.m_GameLogic.IsValidCard(spss.PlayerCards[i][j]) {
				endgameMsg.LeftPai[i][cn] = spss.PlayerCards[i][j]
				cn++
			}
		}
		endgameMsg.HandScore[i] = spss.m_GameLogic.GetScore(endgameMsg.LeftPai[i], cn)
		//spss.LastScore[i] = endgameMsg.Score[i]
		//spss.Total[i] += endgameMsg.Score[i]
		endgameMsg.TotalScore[i] = spss.GetUserTotalScore(i) + endgameMsg.Score[i]
		endgameMsg.FaScore[i] = spss.BombScore[i]

		if spss.MaxScore[i] < spss.PlayerCardScore[i] {
			spss.MaxScore[i] = spss.PlayerCardScore[i]
		}

		if spss.PlayerTurn[i] == 1 { //末游抓的分数归第一游
			//spss.TotalFirstTurn[i]++
			// 记录在玩家身上
			_userItem := spss.GetUserItemByChair(uint16(i))
			if _userItem != nil {
				_userItem.Ctx.PlayTurn1st++
			}
		}

		endgameMsg.TheTurn[i] = spss.PlayerTurn[i]
		endgameMsg.GetScore[i] = spss.PlayerCardScore[i]

		endgameMsg.PlayerScore[i] = spss.PlayerScore[i]
		endgameMsg.TotalPlayerScore[i] = spss.PlayerTotalScore[i]
	}
	if spss.WhoRoar < meta2.MAX_PLAYER {
		spss.TotalDuPai[spss.WhoRoar]++
	}
	banker, bankParter := spss.GetBankerForGameEnd()
	endgameMsg.HouPaiChair = spss.WhoRoar
	endgameMsg.TheBank = banker
	endgameMsg.TheParter = bankParter

	//将玩家分数发送给玩家
	for i := uint16(0); i < uint16(spss.GetPlayerCount()); i++ {
		//注意：抓分是4游的分合并之后的，有的情况不用合并。
		recrodStr := fmt.Sprintf("本局输赢分数%d,当前总分%d，【抓分%d】", endgameMsg.Score[i], endgameMsg.TotalScore[i], spss.PlayerCardScore[i])
		spss.OnWriteGameRecord(i, recrodStr)
	}
	for i := 0; i < static.MAX_PLAYER_4P; i++ {
		endgameMsg.XiScore[i] = spss.XiScore[i] //喜钱
	}

	//记录异常结束数据
	for i := 0; i < spss.GetPlayerCount(); i++ {
		_item := spss.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE && i != int(wChairID) {
			//spss.addReplayOrder(uint16(i), E_Li_Xian, 0)
		}
	}

	//回放增加结束数据
	spss.addReplayOrder(wChairID, DG_REPLAY_OPT_DIS_GAME, 0, []int{})

	spss.ReplayRecord.EndInfo = &endgameMsg
	//写入游戏回放数据,写完重置当前回放数据
	spss.TableWriteOutDate()

	//发送信息
	spss.VecGameEnd = append(spss.VecGameEnd, endgameMsg) //保存，用于汇总计算
	spss.SaveGameData()
	spss.SendTableMsg(consts.MsgTypeGameEnd, endgameMsg)

	//游戏记录
	spss.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

	spss.UpdateOtherFriendDate(&endgameMsg, true)
	// 写总计算
	spss.CalculateResultTotal(static.GER_DISMISS, wChairID, 1, banker, bankParter)
	spss.GameEndStatus = static.GS_MJ_END
	//结束游戏
	//spss.SetGameStatus(public.GS_MJ_FREE)
	spss.ConcludeGame()

	spss.RepositTable(true) // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true
}

//游戏结束,流局结束，统计积分
func (spss *SportSS510K) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if spss.GameEndStatus == static.GS_MJ_END && cbReason == static.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}

	//by leon总结算时才能
	//m_pITableFrame->KillGameTimer(IDI_OUT_TIME);
	// 清除超时检测
	for _, v := range spss.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}
	//1.如果提供的用户为空，不可能，直接返回false
	switch cbReason {
	case static.GER_NORMAL: //常规结束

		return spss.OnGameEndNormal(wChairID, meta2.GOT_NORMAL)
	case static.GER_USER_LEFT: //用户强退

		return spss.OnGameEndUserLeft(wChairID, meta2.GOT_ESCAPE)

	case static.GER_DISMISS: //解散游戏

		return spss.OnGameEndDissmiss(wChairID, meta2.GOT_DISMISS)

	case static.GER_GAME_ERROR: //程序异常，解散游戏

		return spss.OnGameEndErrorDissmiss(wChairID, meta2.GOT_DISMISS)

	}
	return false
}

//计算总发送总结算
func (spss *SportSS510K) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte, banker, partner uint16) {
	// 给客户端发送总结算数据
	var balanceGame Msg_S_DG_BALANCE_GAME
	balanceGame.TheBank = banker
	balanceGame.TheParter = partner
	balanceGame.Userid = spss.Rule.FangZhuID
	balanceGame.CurTotalCount = spss.CurCompleteCount //总盘数
	spss.TimeEnd = time.Now().Unix()                  //大局结束时间
	balanceGame.TimeStart = spss.Common.TimeStart     //游戏大局开始时间
	balanceGame.TimeEnd = spss.TimeEnd
	balanceGame.JiFen = spss.IBase / spss.Rule.Radix
	if spss.AddDiFen > 0 {
		balanceGame.Base = spss.AddDiFen / spss.Rule.Radix
	}
	// 存在换桌的情况下 总结算根据座位相加得到的数据是错误的  需要根据记录在玩家身上的积分计算
	for i := 0; i < spss.GetPlayerCount(); i++ {
		_userItem := spss.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		for _, v := range _userItem.Ctx.GameScoreFen {
			balanceGame.GameScore[i] += v
		}
	}

	//打印日志
	recrodStr := fmt.Sprintf("游戏结束，结束原因:%d（0正常结束，1玩家解散，2超时解散）总分: %d, %d, %d, %d", cbReason, balanceGame.GameScore[0], balanceGame.GameScore[1], balanceGame.GameScore[2], balanceGame.GameScore[3])
	spss.OnWriteGameRecord(wChairID, recrodStr)

	for i := 0; i < spss.GetPlayerCount(); i++ {
		spss.OnWriteGameRecord(uint16(i), "当前该玩家离线")
	}

	//非正常结束，记录总结算时玩家状态：0正常，1解散，2离线
	if static.GER_USER_LEFT == cbReason {
		for i := 0; i < spss.GetPlayerCount(); i++ {
			if i == int(wChairID) {
				balanceGame.UserEndState[i] = 2
			}

		}
	} else {
		if static.GER_DISMISS == cbReason {
			for i := 0; i < spss.GetPlayerCount(); i++ {
				if i == int(wChairID) {
					balanceGame.UserEndState[i] = 1
				} else {
					if _item := spss.GetUserItemByChair(uint16(i)); _item != nil && _item.UserStatus == static.US_OFFLINE {
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
		for i := 0; i < spss.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < spss.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				//				iChairID = j
				iMaxScoreCount++
			}
		}
		if iMaxScoreCount == 1 && spss.Rule.CreateType == 3 { // 大赢家支付
			//IServerUserItem * pIServerUserItem = m_pITableFrame->GetServerUserItem(iChairID);
			//DWORD userid = pIServerUserItem->GetUserID();
			//				m_pITableFrame->TableDeleteDaYingJiaFangKa(userid);
		}
	}

	//! 计算大赢家的座位号
	bigWinScore := 0
	for i := 0; i < spss.GetPlayerCount(); i++ {
		if balanceGame.GameScore[i] >= bigWinScore {
			bigWinScore = balanceGame.GameScore[i]
		}
	}

	//开始就解散时不加，每局结束后游戏状态会复位为GS_MJ_FREE
	wintype := static.ScoreKind_Draw
	if spss.CurCompleteCount == 1 && spss.GetGameStatus() != static.GS_MJ_END {
		if spss.ReWriteRec <= 1 {
			wintype = static.ScoreKind_pass
		}
	} else {
		if spss.CurCompleteCount == 0 { //有可能第一局还没有开始，就解散了（比如在吓跑的过程中解散）
			wintype = static.ScoreKind_pass
		}
	}

	if cbSubReason == 0 {
		for i := 0; i < spss.GetPlayerCount(); i++ {
			_userItem := spss.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			balanceGame.FirstTurnCount[i] = _userItem.Ctx.PlayTurn1st
			balanceGame.WinCount[i] = _userItem.Ctx.WinCount
		}

		for i := 0; i < spss.GetPlayerCount(); i++ {

			balanceGame.LostCount[i] = balanceGame.WinCount[(i+1)%2]
			balanceGame.TotalScore[i] = spss.PlayerTotalScore[i]
			balanceGame.JiCount[i] = spss.JiCount

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
			spss.TableWriteGameDateTotal(int(spss.CurCompleteCount), i, wintype, balanceGame.GameScore[i], isBigWin)
		}
	} else {
		spss.UpdateErrGameTotal(spss.GetTableInfo().GameNum)
	}

	for i := 0; i < spss.GetPlayerCount(); i++ {
		for j := 0; j < len(spss.VecGameEnd); j++ {
			balanceGame.GongFen[i] += spss.VecGameEnd[j].GongFen[i]
			balanceGame.FaScore[i] += spss.VecGameEnd[j].FaScore[i]
			balanceGame.GetScore[i] += spss.VecGameEnd[j].GetScore[i]
		}
		_userItem := spss.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}

		gameendStr := ""
		if len(spss.VecGameEnd) > 0 {
			gameendStr = static.HF_JtoA(spss.VecGameEnd[len(spss.VecGameEnd)-1])
		}
		gamedataStr := ""
		if len(spss.VecGameData[i]) > 0 {
			gamedataStr = static.HF_JtoA(spss.VecGameData[i][len(spss.VecGameData[i])-1])
		}

		spss.SaveLastGameinfo(_userItem.Uid, gameendStr, static.HF_JtoA(balanceGame), gamedataStr)
	}

	// 记录用户好友房历史战绩
	if wintype != static.ScoreKind_pass {
		spss.TableWriteHistoryRecordWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
		spss.TableWriteHistoryRecordDetailWith(int(balanceGame.CurTotalCount), balanceGame.GameScore[:])
	}

	balanceGame.End = 0

	//发消息
	spss.SendTableMsg(consts.MsgTypeGameBalanceGame, balanceGame)

	spss.resetEndDate()
}

func (spss *SportSS510K) resetEndDate() {
	spss.CurCompleteCount = 0
	spss.VecGameEnd = []static.Msg_S_DG_GameEnd{}
	spss.VecGameData = [MAX_PLAYER][]CMD_S_DG_StatusPlay{}

	for _, v := range spss.PlayerInfo {
		v.OnEnd()
	}
}

func (spss *SportSS510K) UpdateOtherFriendDate(GameEnd *static.Msg_S_DG_GameEnd, bEnd bool) {

}

//! 给客户端发送总结算
func (spss *SportSS510K) CalculateResultTotal_Rep(msg *static.Msg_C_BalanceGameEeq) {
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = spss.Rule.FangZhuID
	balanceGame.CurTotalCount = spss.CurCompleteCount //总盘数

	/*为什么要特意判断下0？之前没有这个协议，是新需求要加的之前很多数据都在RepositTableFrameSink里面重置了，但是由于新功能的开发
	部分数据放到游戏开始的时候重置了，但是存在的问题是，如果一个桌子解散了，这个桌子在服务器内部还是存在的，他的部分数据也是存在的
	当某个玩家进来，还没有点开始的时候点战绩，这时候存在把本包厢桌子上次数据发送过来的情况，但是改动m_MaxFanUserCount之类变量的值又
	有很大风险，因此此处做个特出处理，如果是第0局，没有开始，那就无条件全部返回0*/
	if 0 == balanceGame.CurTotalCount {
		for i := 0; i < len(spss.VecGameEnd); i++ {
			for j := 0; j < spss.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += 0 //总分
			}
		}

		for i := 0; i < spss.GetPlayerCount(); i++ {
			balanceGame.ChiHuUserCount[i] = 0
			balanceGame.ProvideUserCount[i] = 0
			balanceGame.FXMaxUserCount[i] = 0
			balanceGame.HHuUserCount[i] = 0
			balanceGame.UserEndState[i] = 0
		}
	} else {
		for i := 0; i < len(spss.VecGameEnd); i++ {
			for j := 0; j < spss.GetPlayerCount(); j++ {
				balanceGame.GameScore[j] += spss.VecGameEnd[i].Score[j] //总分
			}
		}

		for i := 0; i < spss.GetPlayerCount(); i++ {
			balanceGame.UserEndState[i] = 0
		}

		// 有打赢家
		iMaxScore := 0
		iMaxScoreCount := 0
		for i := 0; i < spss.GetPlayerCount(); i++ {
			if iMaxScore < balanceGame.GameScore[i] {
				iMaxScore = balanceGame.GameScore[i]
			}
		}

		// 遍历一下最大值的数量
		for j := 0; j < spss.GetPlayerCount(); j++ {
			if iMaxScore == balanceGame.GameScore[j] {
				iMaxScoreCount++
			}
		}

		for i := 0; i < spss.GetPlayerCount(); i++ {
			_userItem := spss.GetUserItemByChair(uint16(i))
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
	spss.SendPersonMsg(consts.MsgTypeGameBalanceGame, balanceGame, spss.GetChairByUid(msg.Id))
}

//保存桌面数据
func (spss *SportSS510K) SaveGameData() {
	//变量定义
	var StatusPlay CMD_S_DG_StatusPlay
	//游戏变量
	StatusPlay.Overtime = 0
	if time.Now().Unix()+1 < spss.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.Overtime = spss.LimitTime - time.Now().Unix()
	}

	StatusPlay.GameState = byte(spss.GameState) //游戏状态，当前处在哪个阶段，1吼牌阶段，2打牌阶段
	StatusPlay.BankerUser = spss.Banker         //庄家
	StatusPlay.CurrentUser = spss.Whoplay       //当前牌权玩家
	StatusPlay.CellScore = spss.GetCellScore()  //m_pGameServiceOption->lCellScore;
	StatusPlay.WhoLastOut = spss.WhoLastOut     //上一个出牌玩家
	StatusPlay.RoarPai = spss.RoarPai           //叫的什么牌
	StatusPlay.WhoRoar = spss.WhoRoar           //谁叫了牌
	StatusPlay.WhoMJ = static.INVALID_CHAIR     //初始化谁鸣鸡为无效值
	if spss.BMingJiFlag {
		StatusPlay.WhoMJ = spss.BankParter //谁鸣鸡
	}
	StatusPlay.TurnScore = spss.CardScore           //本轮分
	StatusPlay.LastPaiType = spss.LastOutTypeClient //上一次出牌的类型，可能没用？修改成客户端传过来的
	StatusPlay.TheOrder = spss.CurCompleteCount
	StatusPlay.OutScorePai = spss.OutScorePai //所有分牌

	for i := 0; i < spss.GetPlayerCount(); i++ {

		_item := spss.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE {
			StatusPlay.WhoBreak[i] = true //掉线的有那些人
		}
		StatusPlay.TuoGuanPlayer[i] = _item.CheckTRUST() //托管的有那些人
		//StatusPlay.LastScore[i] = int(spss.LastScore[i])    		//上一轮输赢
		StatusPlay.Total[i] = spss.GetUserTotalScore(uint16(i))   //总输赢
		StatusPlay.WhoPass[i] = spss.WhoPass[i]                   //谁放弃
		StatusPlay.WhoReady[i] = spss.WhoReady[i]                 //谁已经完成叫牌过程
		StatusPlay.Score[i] = spss.PlayerCardScore[i]             //每个人的分
		StatusPlay.PlayerScore[i] = spss.PlayerScore[i]           //本局抓分加炸弹分
		StatusPlay.PlayerTotalScore[i] = spss.PlayerTotalScore[i] //累计总分

		for j := 0; j < static.MAX_CARD_4P; j++ {
			StatusPlay.OutCard[i][j] = spss.AllPaiOut[i][j]      //刚才出的牌
			StatusPlay.LastOutCard[i][j] = spss.LastPaiOut[i][j] //上一轮出的牌，（这个其实可以不要）
		}
	}

	//玩家的个人数据
	for i := 0; i < spss.GetPlayerCount() && i < MAX_PLAYER; i++ {
		StatusPlay.TrustCounts = spss.TrustCounts[i] //叫我托管了几次了
		StatusPlay.MyCards = spss.PlayerCards[i]
		StatusPlay.MyCardsCount = spss.ThePaiCount[i]
		StatusPlay.WhoReLink = uint16(i)                              //谁断线重连的
		spss.VecGameData[i] = append(spss.VecGameData[i], StatusPlay) //保存，用于汇总计算
	}
}

func (spss *SportSS510K) sendGameSceneStatusPlay(player *components2.Player) bool {

	wChiarID := player.GetChairID()

	if wChiarID >= uint16(spss.GetPlayerCount()) {
		xlog.Logger().Info("sendGameSceneStatusPlay invalid chair")
		return false
	}

	//重置离线解散时间
	spss.SetOfflineRoomTime(1800)

	//发送底分
	var msgRule static.Msg_S_DG_GameRule
	msgRule.CellScore = spss.GetCellScore()
	msgRule.FaOfTao = spss.FaOfTao
	spss.SendTableMsg(consts.MsgTypeGameRule, msgRule)

	//spss.WhoBreak[wChiarID] = false;//重连嘛，所以取消断线

	//提示其他玩家，我又回来了！
	var msgTip static.Msg_S_DG_ReLinkTip
	msgTip.ReLinkUser = wChiarID
	msgTip.ReLinkTip = 0
	spss.SendTableMsg(consts.MsgTypeReLinkTip, msgTip)

	//取消托管

	//变量定义
	var StatusPlay CMD_S_DG_StatusPlay
	//游戏变量
	StatusPlay.Overtime = 0
	if time.Now().Unix()+1 < spss.LimitTime { //如果只剩下1秒了，就不重新算了
		StatusPlay.Overtime = spss.LimitTime - time.Now().Unix()
	}

	StatusPlay.GameState = byte(spss.GameState)         //游戏状态，当前处在哪个阶段，1吼牌阶段，2打牌阶段
	StatusPlay.BankerUser = spss.Banker                 //庄家
	StatusPlay.CurrentUser = spss.Whoplay               //当前牌权玩家
	StatusPlay.CellScore = spss.GetCellScore()          //m_pGameServiceOption->lCellScore;
	StatusPlay.WhoReLink = wChiarID                     //谁断线重连的
	StatusPlay.WhoLastOut = spss.WhoLastOut             //上一个出牌玩家
	StatusPlay.TrustCounts = spss.TrustCounts[wChiarID] //叫我托管了几次了
	StatusPlay.RoarPai = spss.RoarPai                   //叫的什么牌
	StatusPlay.WhoRoar = spss.WhoRoar                   //谁叫了牌
	StatusPlay.WhoMJ = static.INVALID_CHAIR             //初始化谁鸣鸡为无效值
	if spss.BMingJiFlag {
		StatusPlay.WhoMJ = spss.BankParter //谁鸣鸡
	}
	StatusPlay.TurnScore = spss.CardScore           //本轮分
	StatusPlay.LastPaiType = spss.LastOutTypeClient //上一次出牌的类型，可能没用？修改成客户端传过来的
	StatusPlay.TheOrder = spss.CurCompleteCount
	StatusPlay.OutScorePai = spss.OutScorePai //所有分牌

	for j := 0; j < static.MAX_CARD_4P; j++ {
		if spss.m_GameLogic.IsValidCard(spss.PlayerCards[wChiarID][j]) {
			StatusPlay.MyCards[j] = spss.PlayerCards[wChiarID][j] //刚才出的牌
			StatusPlay.MyCardsCount++
		}
	}
	for i := 0; i < spss.GetPlayerCount(); i++ {

		_item := spss.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.UserStatus == static.US_OFFLINE {
			StatusPlay.WhoBreak[i] = true //掉线的有那些人
		}
		StatusPlay.TuoGuanPlayer[i] = _item.CheckTRUST() //托管的有那些人
		//StatusPlay.LastScore[i] = int(spss.LastScore[i])    		//上一轮输赢
		StatusPlay.Total[i] = spss.GetUserTotalScore(uint16(i))   //总输赢
		StatusPlay.WhoPass[i] = spss.WhoPass[i]                   //谁放弃
		StatusPlay.WhoReady[i] = spss.WhoReady[i]                 //谁已经完成叫牌过程
		StatusPlay.Score[i] = spss.PlayerCardScore[i]             //每个人的分
		StatusPlay.PlayerScore[i] = spss.PlayerScore[i]           //本局抓分加炸弹分
		StatusPlay.PlayerTotalScore[i] = spss.PlayerTotalScore[i] //累计总分

		for j := 0; j < static.MAX_CARD_4P; j++ {
			StatusPlay.OutCard[i][j] = spss.AllPaiOut[i][j]      //刚才出的牌
			StatusPlay.LastOutCard[i][j] = spss.LastPaiOut[i][j] //上一轮出的牌，（这个其实可以不要）
		}
	}

	//发送场景
	spss.SendPersonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, wChiarID)

	//出完牌的顺序，游次
	spss.SendPlayerTurn(wChiarID)

	//剩余牌数量
	if spss.ShowHandCardCnt {
		spss.SendPaiCount(wChiarID)
	}

	if spss.GameType == meta2.GT_NORMAL {
		spss.SendPlayerScore(wChiarID, static.INVALID_CHAIR, 0, true) //每个人的分,这个找朋友模式才有
	}

	//spss.SendTaskID(false, wChiarID);
	spss.GameTask.SendTaskID(spss.Common, false, wChiarID)

	//如果断线之前我的牌出完了,把队友的牌发给我
	if spss.WhoAllOutted[wChiarID] && spss.SeeTeamerCard {
		teamer := spss.GetTeamer(wChiarID)
		spss.SendPaiToTeamer(teamer, wChiarID)
	}

	//发送权限（什么权限、该谁出牌、倒计时等）
	if spss.GameState == meta2.GsRoarPai {
		spss.SendPower(spss.Whoplay, 1, int(StatusPlay.Overtime))
	} else if spss.GameState == meta2.GsPlay {
		spss.SendPower(spss.Whoplay, 2, int(StatusPlay.Overtime))
	} else if spss.GameState == meta2.Gs4KingScore {
		spss.SendPower(spss.WhoHas4KingPower, 4, int(StatusPlay.Overtime))
	}

	//发小结消息
	if byte(len(spss.VecGameEnd)) == spss.CurCompleteCount && spss.CurCompleteCount != 0 && int(wChiarID) < spss.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := spss.VecGameEnd[spss.CurCompleteCount-1]
		gamend.Relink = 1 //表示为断线重连

		spss.SendPersonMsg(consts.MsgTypeGameEnd, gamend, wChiarID)
	}

	//发送解散房间所有玩家的反应
	spss.SendAllPlayerDissmissInfo(player)

	return true
}

//游戏场景消息发送
func (spss *SportSS510K) SendGameScene(uid int64, status byte, secret bool) {
	player := spss.GetUserItemByUid(uid)
	if player == nil {
		return
	}
	switch status {
	case static.GS_MJ_FREE:
		spss.sendGameSceneStatusFree(player)
	case static.GS_MJ_PLAY:
		spss.sendGameSceneStatusPlay(player)
	}
}
func (spss *SportSS510K) sendGameSceneStatusFree(player *components2.Player) bool {

	//变量定义
	var StatusFree static.Msg_S_DG_StatusFree
	//构造数据
	StatusFree.BankerUser = spss.Banker
	StatusFree.CellScore = spss.GetCellScore() //spss.m_pGameServiceOption->lCellScore;
	StatusFree.FaOfTao = spss.FaOfTao
	StatusFree.CellMinScore = spss.GetCellScore() //最低分
	StatusFree.CellMaxScore = spss.GetCellScore() //最低分

	//发送场景
	//	spss.SendPersonMsg(constant.MsgTypeGameStatusFree, StatusFree, PlayerInfo.GetChairID())
	spss.SendUserMsg(consts.MsgTypeGameStatusFree, StatusFree, player.Uid)

	return true
}

//! 游戏退出
func (spss *SportSS510K) OnExit(uid int64) {
	spss.Common.OnExit(uid)
}

func (spss *SportSS510K) OnTime() {
	spss.Common.OnTime()
}

//! 写游戏日志
func (spss *SportSS510K) OnWriteGameRecord(seatId uint16, recordStr string) {
	spss.GameTable.WriteTableLog(seatId, recordStr)
}

//! 写入游戏回放数据
func (spss *SportSS510K) TableWriteOutDate() {
	if spss.ReWriteRec != 0 {
		spss.ReWriteRec++ //这种情况会>1，表示是在结算时申请解散的。
		// 写完后清除数据
		spss.ReplayRecord.ReSet()
		return
	}

	recordReplay := new(models.RecordGameReplay)
	recordReplay.GameNum = spss.GetTableInfo().GameNum
	recordReplay.RoomNum = spss.GetTableInfo().Id
	recordReplay.PlayNum = int(spss.CurCompleteCount)
	recordReplay.ServerId = server2.GetServer().Con.Id
	recordReplay.HandCard = spss.m_GameLogic.GetWriteHandReplayRecordString(spss.ReplayRecord)
	recordReplay.OutCard = spss.m_GameLogic.GetWriteOutReplayRecordString(spss.ReplayRecord)
	recordReplay.KindID = spss.GetTableInfo().KindId
	recordReplay.CardsNum = 0

	if spss.ReplayRecord.EndInfo != nil {
		recordReplay.EndInfo = static.HF_JtoA(spss.ReplayRecord.EndInfo)
	}

	if id, err := server2.GetDBMgr().InsertGameRecordReplay(recordReplay); err != nil {
		xlog.Logger().Debug(fmt.Sprintf("%d,写游戏出牌记录：（%v）出错（%v）", id, recordReplay, err))
	}

	spss.RoundReplayId = recordReplay.Id

	// 写完后清除数据
	spss.ReplayRecord.ReSet()

	spss.ReWriteRec++ //在小结算过程中解散不在写回放记录了
}

//场景保存
func (spss *SportSS510K) Tojson() string {
	var _json SportSS510KJsonSerializer

	_json.ToJsonSS510k(&spss.SportMetaSS510K)

	_json.GameCommonToJson(&spss.Common)

	return static.HF_JtoA(&_json)
}

//场景恢复
func (spss *SportSS510K) Unmarsha(data string) {
	var _json SportSS510KJsonSerializer

	if data != "" {
		json.Unmarshal([]byte(data), &_json)

		_json.UnmarshaSS510k(&spss.SportMetaSS510K)
		_json.JsonToStruct(&spss.Common)

		spss.ParseRule(spss.GetTableInfo().Config.GameConfig)
		spss.m_GameLogic.Rule = spss.Rule

		//_game.m_GameLogic.InitMagicPoint(CARDINDEX_SMALL) //大小王都是赖子
		//_game.m_GameLogic.AddMagicPoint(CARDINDEX_BIG)
		//石首510k有花牌癞子
		spss.m_GameLogic.InitMagicPoint(logic2.CARDINDEX_SKY)
		spss.m_GameLogic.SetBombCount(4)                     //设置炸弹的最小长度(普通炸弹：四个或者四个以上相同的牌 5 10 K也是炸弹3张牌)
		spss.m_GameLogic.SetOnestrCount(254)                 //设置单顺的最小长度
		spss.m_GameLogic.SetPlayMode(spss.PlayMode)          //玩法模式
		spss.m_GameLogic.SetMaxCardCount(static.MAX_CARD_4P) //设置手牌最大长度
		if spss.PlayMode == 0 {
			spss.m_GameLogic.SetMaxCardCount(static.MAX_CARD_4P - 1) //经典玩法26张手牌
		}
		spss.m_GameLogic.SetMaxPlayerCount(static.MAX_PLAYER_4P) //设置玩家最大数目
	}
}

//解析配置的任务,格式： "1@5/2@5/33@10"
func (spss *SportSS510K) ParseTaskConfig(data string) {
	spss.GameTask.Init()
	//.....
	//spss.Task.AppendTaskMapAndVec(id,award)
}

//参数:是否是庄家一个队伍的
func (spss *SportSS510K) GetTeamYouType(isBankerTeam bool) int {
	youb := spss.PlayerTurn[spss.Banker]
	youp := spss.PlayerTurn[spss.BankParter]

	youtype := 0
	if (youb == 1 && youp == 2) || (youb == 2 && youp == 1) {
		//庄家及队友12游
		if isBankerTeam {
			youtype = YOUTYPE_12
		} else {
			youtype = YOUTYPE_34
		}
	} else if (youb == 1 && youp == 3) || (youb == 3 && youp == 1) {
		if isBankerTeam {
			youtype = YOUTYPE_13
		} else {
			youtype = YOUTYPE_24
		}
	} else if (youb == 1 && youp == 4) || (youb == 4 && youp == 1) {
		if isBankerTeam {
			youtype = YOUTYPE_14
		} else {
			youtype = YOUTYPE_23
		}
	} else if (youb == 2 && youp == 3) || (youb == 3 && youp == 2) {
		if isBankerTeam {
			youtype = YOUTYPE_23
		} else {
			youtype = YOUTYPE_14
		}
	} else if (youb == 2 && youp == 4) || (youb == 4 && youp == 2) {
		if isBankerTeam {
			youtype = YOUTYPE_24
		} else {
			youtype = YOUTYPE_13
		}
	} else if (youb == 3 && youp == 4) || (youb == 4 && youp == 3) {
		if isBankerTeam {
			youtype = YOUTYPE_34
		} else {
			youtype = YOUTYPE_12
		}
	}

	return youtype
}

//参数:是否是庄家一个队伍的
func (spss *SportSS510K) IsBankerTeam(seat uint16) bool {
	if seat == spss.Banker || seat == spss.BankParter {
		return true
	}
	return false
}

// 获取玩家总得分
func (spss *SportSS510K) GetUserTotalScore(seat uint16) int {
	score := 0
	_userItem := spss.GetUserItemByChair(seat)
	if _userItem != nil {
		for _, v := range _userItem.Ctx.GameScoreFen {
			score += v
		}
	}
	return score
}

// 得到一张王
func (spss *SportSS510K) GetUserKingCard(seat uint16) byte {
	kingCard := byte(0)
	for i := 0; i < len(spss.PlayerCards[seat]); i++ {
		if spss.PlayerCards[seat][i] == logic2.CARDINDEX_SMALL || spss.PlayerCards[seat][i] == logic2.CARDINDEX_BIG {
			kingCard = spss.PlayerCards[seat][i]
			break
		}
	}
	return kingCard
}

// 用户是否拥有相同的牌
func (spss *SportSS510K) IsHaveSameCard(card byte, seat uint16) bool {
	bFind := false
	for i := 0; i < len(spss.PlayerCards[seat]); i++ {
		if spss.PlayerCards[seat][i] == card {
			bFind = true
			break
		}
	}
	return bFind
}

func (spss *SportSS510K) GetBankerForGameEnd() (banker, bankParter uint16) {
	if spss.Banker == static.INVALID_CHAIR {
		if spss.Nextbanker >= uint16(spss.GetPlayerCount()) {
			rand_num := rand.Intn(1000)
			banker = uint16(rand_num % spss.GetPlayerCount())
		} else {
			banker = spss.Nextbanker
		}
		// 玩家队友固定为对家
		tmpNextPlayer := spss.GetNextFullSeatV2(banker)
		bankParter = spss.GetNextFullSeatV2(tmpNextPlayer)
	} else {
		banker = spss.Banker
		bankParter = spss.BankParter
	}
	return
}
