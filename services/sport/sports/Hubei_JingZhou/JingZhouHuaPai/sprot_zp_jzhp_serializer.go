package JingZhouHuaPai

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
)

type SportZPTCJsonSerializer struct {
	components2.DGGameCommonJson
	//游戏变量
	ReplayRecord ZP_Replay_Record `json:"replayrecord"` //回放记录
	ReWriteRec   byte             `json:"rewriterec"`   //是否重复写回放数据，每小局游戏开始时清理,打拱可以在小结算中申请解散。

	//运行变量
	Banker     uint16 `json:"banker"`     //庄家
	NextBanker uint16 `json:"nextbanker"` //下一个庄家

	CurrentUser   uint16 `json:"currentuser"`   //当前用户
	OutCardUser   uint16 `json:"outcarduser"`   //出牌用户
	ResumeUser    uint16 `json:"resumeuser"`    //还原用户
	ProvideUser   uint16 `json:"provideuser"`   //供应用户
	NoOutUser     uint16 `json:"nooutuser"`     //选择不出牌的用户 0--MAXPLAYER-1
	GuanStartUser uint16 `json:"guanstartuser"` //起牌后第一个检查观生的用户 0--MAXPLAYER-1

	//托管和离线数据
	TuoGuanPlayer  [TCGZ_MAX_PLAYER]bool `json:"tuoguanplayer"`  //谁托管了？
	TrustCounts    [TCGZ_MAX_PLAYER]byte `json:"trustcounts"`    //玩家托管次数
	AutoCardCounts [TCGZ_MAX_PLAYER]byte `json:"autocardcounts"` //自动出牌的次数
	BreakCounts    [TCGZ_MAX_PLAYER]byte `json:"breakcounts"`    // 断线次数

	//牌数据
	LeftCardCount byte                                      `json:"leftcardcount"` //  剩余数目
	RepertoryCard [TCGZ_ALLCARD]byte                        `json:"repertorycard"` // 所有牌
	CardIndex     [TCGZ_MAX_PLAYER][TCGZ_MAX_INDEX_HUA]byte `json:"cardindex"`     // 玩家分到的牌
	OutCardCount  byte                                      `json:"outcardcount"`  //  总出牌扑克数目
	OutCardData   byte                                      `json:"outcarddata"`   //  出牌扑克
	ProvideCard   byte                                      `json:"providecard"`   //  供应扑克
	DiscardCount  [TCGZ_MAX_PLAYER]byte                     `json:"discardcount"`  // 丢弃数目
	DiscardCard   [TCGZ_MAX_PLAYER][static.MAX_CARD]byte    `json:"discardcard"`   // 丢弃记录
	SendCardData  byte                                      `json:"sendcarddata"`  //  发牌扑克
	SendCardCount byte                                      `json:"sendcardcount"` //  发牌数目

	SendStatus         bool                         `json:"sendstatus"`     //  发牌状态,发牌和出的牌，牌边框颜色不一样
	HuangZhuang        bool                         `json:"huangzhuang"`    // 是否荒庄
	WeaveHuxi          [TCGZ_MAX_PLAYER]int         `json:"weavehxi"`       // 每个人组合牌的总胡息
	UserGuanCards      [TCGZ_MAX_PLAYER][10][5]byte `json:"guancards"`      // 玩家观生的牌
	UserJianCards      [TCGZ_MAX_PLAYER][]byte      `json:"jiancards"`      // 玩家捡的牌
	UserJianCardsCur   [TCGZ_MAX_PLAYER][]byte      `json:"jiancardscur"`   // 玩家本轮捡的牌，这些牌本轮不能打
	UserGuanCardsCount [TCGZ_MAX_PLAYER]byte        `json:"guancardscount"` // 玩家观生的牌数目
	UserPengCount      [TCGZ_MAX_PLAYER]byte        `json:"pengcount"`      // 玩家碰的次数，荆州花牌只能碰2次
	TongInfo           [TCGZ_MAX_PLAYER]TagTongInfo `json:"tonginfo"`       // 玩家统的信息
	DispTongCnt        [TCGZ_MAX_PLAYER]int         `json:"disptongcnt"`    // 客户端显示的玩家统的信息，玩家出牌统数不会减少，就是不能换统也不减少

	//积分数据
	ProvideUserCount [TCGZ_MAX_PLAYER]int `json:"provideusercount"` // 点炮次数
	ChiHuUserCount   [TCGZ_MAX_PLAYER]int `json:"chihuusercount"`   // 胡牌次数
	JiePaoUserCount  [TCGZ_MAX_PLAYER]int `json:"jiepaousercount"`  // 接炮次数
	ZiMoUserCount    [TCGZ_MAX_PLAYER]int `json:"zimousercount"`    // 自摸次数
	Total            [TCGZ_MAX_PLAYER]int `json:"total"`            // 总输赢，若干小局相加的金币
	MaxHUxi          [TCGZ_MAX_PLAYER]int `json:"maxhuxi"`          // 单局最高胡数

	//底
	Base int `json:"infrastructure"` // 底
	Spay int `json:"spay"`           // 服务费

	//好友房
	GeziShu            int  `json:"gezishu"`            // 个子数
	HuaShu             int  `json:"huashu"`             // 花数 ，10表示10个花，1表示溜花
	Piao               int  `json:"piao"`               // 选漂:0不漂，100带漂，1-3定漂
	DuoHu              bool `json:"duohu"`              // true 一炮多响
	BeiShu             int  `json:"beishu"`             // 倍数
	QuanHei            int  `json:"quanhei"`            //全黑倍数
	DianPaoPei         byte `json:"dianpaopei"`         // 点炮包赔
	HunJiang           byte `json:"hunjiang"`           // 混江
	KeChong            bool `json:"kechong"`            //可以放铳，true可以放铳
	NoOut              bool `json:"noout"`              //true不出牌（捏牌）
	FenType            int  `json:"fentype"`            //算分类型,数字型,0算胡数，1算坡数，2登庄
	Fleetime           int  `json:"fleetime"`           //客户端传来的 游戏开始前离线踢人时间
	RoundOverAutoReady bool `json:"roundoverautoready"` //小局结束自动准备
	DengZhuang         bool `json:"dengzhuang"`         //登庄，false：胡牌玩家的下家当庄，true：胡牌玩家当庄。第一局随机，流局连庄
	OutCardDismissTime int  `json:"outcarddismisstime"` // 出牌时间 超时房间强制解散 -1不限制

	//状态变量
	VecGameEnd  []static.Msg_S_ZP_TC_GameEnd    `json:"vecgameend"`  //记录每一局的结果
	VecGameData []static.CMD_S_ZP_TC_StatusPlay `json:"vecgamedata"` //记录每一局的结果

	//控制变量
	VecChiHuCard   [TCGZ_MAX_PLAYER][]byte           `json:"vecchihucard"`   // 本轮弃吃胡的牌，用来控制大冶字牌的过庄
	VecChiCard     [TCGZ_MAX_PLAYER][]byte           `json:"vecchicard"`     // 本轮弃吃的牌，用来控制大冶字牌的过庄
	VecPengCard    [TCGZ_MAX_PLAYER][]byte           `json:"vecpengcard"`    // 本轮弃碰的牌，用来控制大冶字牌的过庄
	Response       [TCGZ_MAX_PLAYER]bool             `json:"response"`       //  响应标志
	UserAction     [TCGZ_MAX_PLAYER]int              `json:"useraction"`     //  用户动作
	PerformAction  [TCGZ_MAX_PLAYER]int              `json:"performaction"`  //  执行动作
	OperateCard    [TCGZ_MAX_PLAYER]byte             `json:"operatecard"`    //  操作扑克
	WeaveItemCount [TCGZ_MAX_PLAYER]byte             `json:"weaveitemcount"` //  组合数目
	WeaveItemArray [TCGZ_MAX_PLAYER][10]TagWeaveItem `json:"weaveitemarray"` //  组合数目

	CMD_OperateCard [TCGZ_MAX_PLAYER]static.Msg_C_ZP_OperateCard `json:"cmd_operatecard"` //多个人有牌权时，用于保存摆牌数据
	ChiHuResult     [TCGZ_MAX_PLAYER]TagChiHuResult              `json:"chihuresult"`     //吃胡结果
	ChiHuCard       byte                                         `json:"chihucard"`       // 吃胡扑克
	AutoOut         byte                                         `json:"autoout"`         // //是否超时自动出牌
	UserReady       [TCGZ_MAX_PLAYER]bool                        `json:"userready"`       //玩家是否已经准备

	//组件变量
	GameLogic logic2.BaseLogic `json:"gamelogic"` //游戏逻辑

	LastTime int64 `json:"lasttime"` // 用于断线重入 校准客户端操作时间
	//时间数据
	PowerStartTime int64 `json:"powerstarttime"` //时间辅助变量，权限开始时间
	OperateTime    int64 `json:"operatetime"`    // 出牌时间
	AutoOutTime    int64 `json:"autoouttime"`    // 托管出牌时间

	IsTongAction bool `json:"istongstate"` // 统牌阶段的托管需要加个状态，否则玩家响应的倒计时会和流程中的1秒间隔倒计时混淆

	//	Config static.GameConfig //游戏配置
	GameReplayId int   `json:"gamereplayid"`
	GameState    int   `json:"gamestate"` //游戏中状态，1表示吼牌阶段，2表示打牌阶段
	GameType     int   `json:"gametype"`  //游戏类型 1表示普通模式2vs2，2表示吼牌模式1vs3
	GameTime     int64 `json:"gametime"`  //下次执行时间

	//保存游戏房间记录
	GameConfig static.GameConfig `json:"gameConfig"` //游戏配置
}

func (sjs *SportZPTCJsonSerializer) ToJsonZPTC(_game *SportMetaJZHP) {
	sjs.ReplayRecord = _game.ReplayRecord
	sjs.ReWriteRec = _game.ReWriteRec
	sjs.Banker = _game.Banker
	sjs.NextBanker = _game.Nextbanker
	sjs.CurrentUser = _game.CurrentUser
	sjs.OutCardUser = _game.OutCardUser
	sjs.ResumeUser = _game.ResumeUser
	sjs.ProvideUser = _game.ProvideUser
	sjs.NoOutUser = _game.NoOutUser
	sjs.GuanStartUser = _game.GuanStartUser

	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		sjs.TuoGuanPlayer[i] = _game.TuoGuanPlayer[i]
		sjs.TrustCounts[i] = _game.TrustCounts[i]
		sjs.AutoCardCounts[i] = _game.AutoCardCounts[i]
		sjs.BreakCounts[i] = _game.BreakCounts[i]
		sjs.DiscardCount[i] = _game.DiscardCount[i]
		sjs.WeaveHuxi[i] = _game.WeaveHuxi[i]
		sjs.UserGuanCardsCount[i] = _game.UserGuanCardsCount[i]
		sjs.UserPengCount[i] = _game.UserPengCount[i]

		for j := 0; j < static.MAX_CARD; j++ {
			sjs.DiscardCard[i][j] = _game.DiscardCard[i][j]
		}

		for j := 0; j < TCGZ_MAX_INDEX_HUA; j++ {
			sjs.CardIndex[i][j] = _game.CardIndex[i][j]
		}
	}
	for i := 0; i < TCGZ_ALLCARD; i++ {
		sjs.RepertoryCard[i] = _game.RepertoryCard[i]
	}
	sjs.LeftCardCount = _game.LeftCardCount
	sjs.OutCardCount = _game.OutCardCount
	sjs.OutCardData = _game.OutCardData
	sjs.ProvideCard = _game.ProvideCard
	sjs.SendCardData = _game.SendCardData
	sjs.SendCardCount = _game.SendCardCount
	sjs.SendStatus = _game.SendStatus
	sjs.HuangZhuang = _game.HuangZhuang

	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		sjs.ProvideUserCount[i] = _game.ProvideUserCount[i]
		sjs.Total[i] = _game.Total[i]
		sjs.ChiHuUserCount[i] = _game.ChiHuUserCount[i]
		sjs.JiePaoUserCount[i] = _game.JiePaoUserCount[i]
		sjs.ZiMoUserCount[i] = _game.ZiMoUserCount[i]
		sjs.MaxHUxi[i] = _game.MaxHUxi[i]
		for j := 0; j < 10; j++ {
			for k := 0; k < 5; k++ {
				sjs.UserGuanCards[i][j][k] = _game.UserGuanCards[i][j][k]
			}
		}
		sjs.UserJianCards[i] = _game.UserJianCards[i]
		sjs.UserJianCardsCur[i] = _game.UserJianCardsCur[i]
		sjs.TongInfo[i] = _game.TongInfo[i]
		sjs.DispTongCnt[i] = _game.DispTongCnt[i]
	}

	sjs.Base = _game.Base
	sjs.Spay = _game.Spay

	sjs.GeziShu = _game.GeziShu
	sjs.HuaShu = _game.HuaShu
	sjs.Piao = _game.Piao
	sjs.DuoHu = _game.DuoHu
	sjs.BeiShu = _game.BeiShu
	sjs.QuanHei = _game.QuanHei
	sjs.DianPaoPei = _game.DianPaoPei
	sjs.HunJiang = _game.HunJiang
	sjs.KeChong = _game.KeChong
	sjs.NoOut = _game.NoOut
	sjs.FenType = _game.FenType
	sjs.Fleetime = _game.Fleetime
	sjs.RoundOverAutoReady = _game.RoundOverAutoReady
	sjs.DengZhuang = _game.DengZhuang
	sjs.OutCardDismissTime = _game.OutCardDismissTime

	sjs.ChiHuCard = _game.ChiHuCard
	sjs.AutoOut = _game.AutoOut

	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		sjs.VecChiHuCard[i] = _game.VecChiHuCard[i]
		sjs.VecChiCard[i] = _game.VecChiCard[i]
		sjs.VecPengCard[i] = _game.VecPengCard[i]
		sjs.Response[i] = _game.Response[i]
		sjs.UserAction[i] = _game.UserAction[i]
		sjs.PerformAction[i] = _game.PerformAction[i]
		sjs.OperateCard[i] = _game.OperateCard[i]
		sjs.WeaveItemCount[i] = _game.WeaveItemCount[i]
		sjs.CMD_OperateCard[i] = _game.CMD_OperateCard[i]
		sjs.ChiHuResult[i] = _game.ChiHuResult[i]
		for j := 0; j < 10; j++ {
			sjs.WeaveItemArray[i][j] = _game.WeaveItemArray[i][j]
		}
	}

	sjs.VecGameEnd = _game.VecGameEnd
	sjs.VecGameData = _game.VecGameData

	sjs.LastTime = _game.LastTime
	sjs.OperateTime = _game.OperateTime
	sjs.AutoOutTime = _game.AutoOutTime
	sjs.PowerStartTime = _game.PowerStartTime
	sjs.IsTongAction = _game.IsTongAction
	sjs.GameState = _game.GameState
	sjs.GameType = _game.GameType
	sjs.UserReady = _game.UserReady

	sjs.GameConfig = _game.Config //添加保存game记录
}

func (sjs *SportZPTCJsonSerializer) unmarshaZPTC(_game *SportMetaJZHP) {

	_game.ReplayRecord = sjs.ReplayRecord
	fmt.Println(_game.ReplayRecord)

	_game.ReWriteRec = sjs.ReWriteRec
	_game.Banker = sjs.Banker
	_game.Nextbanker = sjs.NextBanker
	_game.CurrentUser = sjs.CurrentUser
	_game.OutCardUser = sjs.OutCardUser
	_game.ResumeUser = sjs.ResumeUser
	_game.ProvideUser = sjs.ProvideUser
	_game.NoOutUser = sjs.NoOutUser
	_game.GuanStartUser = sjs.GuanStartUser

	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		_game.TuoGuanPlayer[i] = sjs.TuoGuanPlayer[i]
		_game.TrustCounts[i] = sjs.TrustCounts[i]
		_game.AutoCardCounts[i] = sjs.AutoCardCounts[i]
		_game.BreakCounts[i] = sjs.BreakCounts[i]
		_game.DiscardCount[i] = sjs.DiscardCount[i]
		_game.WeaveHuxi[i] = sjs.WeaveHuxi[i]
		_game.UserGuanCardsCount[i] = sjs.UserGuanCardsCount[i]
		_game.UserPengCount[i] = sjs.UserPengCount[i]

		for j := 0; j < static.MAX_CARD; j++ {
			_game.DiscardCard[i][j] = sjs.DiscardCard[i][j]
		}

		for j := 0; j < TCGZ_MAX_INDEX_HUA; j++ {
			_game.CardIndex[i][j] = sjs.CardIndex[i][j]
		}
	}
	for i := 0; i < TCGZ_ALLCARD; i++ {
		_game.RepertoryCard[i] = sjs.RepertoryCard[i]
	}

	_game.LeftCardCount = sjs.LeftCardCount
	_game.OutCardCount = sjs.OutCardCount
	_game.OutCardData = sjs.OutCardData
	_game.ProvideCard = sjs.ProvideCard
	_game.SendCardData = sjs.SendCardData
	_game.SendCardCount = sjs.SendCardCount
	_game.SendStatus = sjs.SendStatus
	_game.HuangZhuang = sjs.HuangZhuang

	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		_game.ProvideUserCount[i] = sjs.ProvideUserCount[i]
		_game.Total[i] = sjs.Total[i]
		_game.ChiHuUserCount[i] = sjs.ChiHuUserCount[i]
		_game.JiePaoUserCount[i] = sjs.JiePaoUserCount[i]
		_game.ZiMoUserCount[i] = sjs.ZiMoUserCount[i]
		_game.MaxHUxi[i] = sjs.MaxHUxi[i]

		for j := 0; j < 10; j++ {
			for k := 0; k < 5; k++ {
				_game.UserGuanCards[i][j][k] = sjs.UserGuanCards[i][j][k]
			}
		}
		_game.UserJianCards[i] = sjs.UserJianCards[i]
		_game.UserJianCardsCur[i] = sjs.UserJianCardsCur[i]

		_game.TongInfo[i] = sjs.TongInfo[i]
		_game.DispTongCnt[i] = sjs.DispTongCnt[i]
	}

	_game.Base = sjs.Base
	_game.Spay = sjs.Spay

	_game.GeziShu = sjs.GeziShu
	_game.HuaShu = sjs.HuaShu
	_game.Piao = sjs.Piao
	_game.DuoHu = sjs.DuoHu
	_game.BeiShu = sjs.BeiShu
	_game.QuanHei = sjs.QuanHei
	_game.DianPaoPei = sjs.DianPaoPei
	_game.HunJiang = sjs.HunJiang
	_game.KeChong = sjs.KeChong
	_game.NoOut = sjs.NoOut
	_game.FenType = sjs.FenType
	_game.Fleetime = sjs.Fleetime
	_game.RoundOverAutoReady = sjs.RoundOverAutoReady
	_game.DengZhuang = sjs.DengZhuang
	_game.OutCardDismissTime = sjs.OutCardDismissTime

	_game.ChiHuCard = sjs.ChiHuCard
	_game.AutoOut = sjs.AutoOut

	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		_game.VecChiHuCard[i] = sjs.VecChiHuCard[i]
		_game.VecChiCard[i] = sjs.VecChiCard[i]
		_game.VecPengCard[i] = sjs.VecPengCard[i]
		_game.Response[i] = sjs.Response[i]
		_game.UserAction[i] = sjs.UserAction[i]
		_game.PerformAction[i] = sjs.PerformAction[i]
		_game.OperateCard[i] = sjs.OperateCard[i]
		_game.WeaveItemCount[i] = sjs.WeaveItemCount[i]
		_game.CMD_OperateCard[i] = sjs.CMD_OperateCard[i]
		_game.ChiHuResult[i] = sjs.ChiHuResult[i]
		for j := 0; j < 10; j++ {
			_game.WeaveItemArray[i][j] = sjs.WeaveItemArray[i][j]
		}
	}

	_game.VecGameEnd = sjs.VecGameEnd
	_game.VecGameData = sjs.VecGameData
	_game.LastTime = sjs.LastTime
	_game.OperateTime = sjs.OperateTime
	_game.AutoOutTime = sjs.AutoOutTime
	_game.PowerStartTime = sjs.PowerStartTime
	_game.IsTongAction = sjs.IsTongAction
	_game.GameState = sjs.GameState
	_game.GameType = sjs.GameType
	_game.UserReady = sjs.UserReady

	//_game.Config             =        sjs.GameConfig
}

//获取累加的计算
func (sjs *SportZPTCJsonSerializer) getString() string {
	var str string
	return str
}

func (sjs *SportZPTCJsonSerializer) ToJson_ZP_TCGZ(_game *Sport_zp_jzhp) string {
	sjs.ToJsonZPTC(&_game.SportMetaJZHP)

	sjs.GameCommonToJson(&_game.Common)

	return static.HF_JtoA(sjs)
}
func (sjs *SportZPTCJsonSerializer) Unmarsha_ZP_TCGZ(data string, _game *Sport_zp_jzhp) {
	if data != "" {
		json.Unmarshal([]byte(data), sjs)

		sjs.unmarshaZPTC(&_game.SportMetaJZHP)
		sjs.JsonToStruct(&_game.Common)

		_game.ParseRule(_game.GetTableInfo().Config.GameConfig)
		_game.GameLogic.Rule = _game.Rule
		_game.GameLogic.SetHunJiang(int(_game.HunJiang))
	}
}
