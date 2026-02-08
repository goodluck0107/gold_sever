package Hubei_JingZhou

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	base2 "github.com/open-source/game/chess.git/services/sport/infrastructure"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	server2 "github.com/open-source/game/chess.git/services/sport/wuhan"
	"math/rand"
	"strings"
	"time"
)

const (
	GameHu_NoMagic_QingYiSe_Laiyou = 1 //硬清一色赖油   4分或者8分
	GameHu_NoMagicLaiYou           = 2 //硬赖油   固定4分
	GameHu_NoMagic_QingYiSe        = 3 //硬清一色 2分或者4分
	GameHu_Magic_QingYiSe_Laiyou   = 4 //软清一色赖油   2分或者4分
	GameHu_NoMagic                 = 5 //硬自摸  固定2分
	GameHu_MagicLaiYou             = 6 //软赖油固定2分
	GameHu_Magic_QingYiSe          = 7 //软清一色  1分或者2分
	GameHu_Magic                   = 8 //软自摸 固定1分
)

// 换座消息
type Msg_C_JZ_CHANGESEAT struct {
	ChairID byte `json:"chairid"` //换座位号
}

// 好友房规则相关属性
type FriendRule_jzcxz struct {
	Difen              int    `json:"difen"`              // 底分
	Zhangnum           int    `json:"zhangnum"`           // 去万(108张牌就是不去万，72张牌去万)
	Wanfa              int    `json:"wanfa"`              // 0一脚赖油 1半赖 2无赖到底
	Noyy               string `json:"no_yy"`              // 禁止语音
	No_hdbq            string `json:"no_hdbq"`            // 禁止互动表情
	No_qph             string `json:"no_qph"`             // 禁止俏皮话
	Fewerstart         string `json:"fewerstart"`         // 可少人开局
	JianZiHu           string `json:"kjzh"`               // 是否可见字胡
	HuanSanZhang       string `json:"hsz"`                // 是否换三张
	Gpsm               string `json:"gpsm"`               // 杠牌顺摸
	Yzy                string `json:"yzy"`                // 油中油
	Overtimetrust      int    `json:"overtime_trust"`     //超时托管
	Overtimedismiss    int    `json:"overtime_dismiss"`   //超时解散
	Beishu             int    `json:"beishu"`             //放杠倍数 传几就是几倍
	FleeTime           int    `json:"fleetime"`           // 离线踢人
	Endready           string `json:"endready"`           //小结算是否自动准备
	Gmgd               string `json:"gmgd"`               // 各摸各的
	Overtime_offdiss   int    `json:"overtime_offdiss"`   //离线解散时间
	Overtime_applydiss int    `json:"overtime_applydiss"` //申请解散时间
	LookonSupport      string `json:"LookonSupport"`      //本局游戏是否支持旁观
	OutCardDismissTime int    `json:"outcarddismisstime"` // 出牌时间 超时房间强制解散 -1不限制
	TimeOutPunish      string `json:"overtime_score"`     // 超时罚分
	PunishScore        int    `json:"punishscore"`        // 罚分
	Dissmiss           int    `json:"dissmiss"`
}

// 好友房规则相关属性
type FriendRule_jzcxz_old struct {
	Difen        int    `json:"difen"`       // 底分
	SecondRoom   int    `json:"outcardTime"` // 0表示没有限制 12表示12秒场 15表示15秒场
	Zhangnum     int    `json:"zhangnum"`    // 去万(108张牌就是不去万，72张牌去万)
	Wanfa        int    `json:"wanfa"`       // 0一脚赖油 1半赖 2无赖到底
	Noyy         string `json:"no_yy"`       // 禁止语音
	No_hdbq      string `json:"no_hdbq"`     // 禁止互动表情
	No_qph       string `json:"no_qph"`      // 禁止俏皮话
	Fewerstart   string `json:"fewerstart"`  // 可少人开局
	JianZiHu     string `json:"kjzh"`        // 是否可见字胡
	HuanSanZhang string `json:"hsz"`         // 是否换三张
	Gpsm         string `json:"gpsm"`        // 杠牌顺摸
	Yzy          string `json:"yzy"`         // 油中油
	Dissmiss     int    `json:"dissmiss"`
}

type SportJZCXZ struct {
	// 游戏共用部分
	components2.Common
	// 游戏流程数据
	meta2.Metadata
	//游戏逻辑
	m_GameLogic SportLogicJZCXZ

	//记录每个玩家的杠分
	m_playerGangFen [static.MAX_PLAYER_4P][static.MAX_PLAYER_4P]int

	/*当前牌权玩家连续杠开次数（比如牌权到自己了，有两个及以上的赖子 一直杠，第一次杠开（可以不胡），第二次又杠开）然后胡了
	只要有一次出牌了，或者牌权到别人那里了，都不算连续杠开.只有自摸胡，没有点炮胡，所以不用区分记录每个玩家的杠开只记录当前的就行了*/
	magicGangWin [4]bool //玩家最多有4个赖子

	/*1.杠一个赖子杠开后，手牌里面还有朝天杠，这个时候既能朝天杠又能胡牌的时候，如果先选择朝天杠，再选择胡，这个胡下来还是算癞油。（只有朝天杠才是，如果是暗杠蓄杠接着杠然后杠下来补一张牌还是能胡，这个不算癞油）
	2.杠一个赖子杠开后，手牌还有朝天杠，这个时候选择朝天杠，然后接着又赖子杠，然后杠开，这个算油中油。*/
	mLastGangIsMagicGang bool //用来记录最近的杠是不是赖子杠（1 2的这种情形）
	// 出牌时间 超时房间强制解散 -1不限制
	OutCardDismissTime int

	autoPiZiGangOffset [static.MAX_PLAYER_4P]int
}

// ! 设置游戏可胡牌类型
func (sp *SportJZCXZ) HuTypeInit(_type *static.TagHuType) {
	_type.HAVE_SIXI_HU = false
	_type.HAVE_QUE_YISE_HU = false
	_type.HAVE_BANBAN_HU = false
	_type.HAVE_LIULIU_HU = false
	_type.HAVE_HAO_HUA_DUI_HU = false
	_type.HAVE_HAI_DI_HU = false
	_type.HAVE_QIANG_GANG_HU = false
	_type.HAVE_MENG_QING = false
	_type.HAVE_DI_HU = false
	_type.HAVE_TIAN_HU = false
	_type.HAVE_ZIMO_JIAO_1 = false
	_type.HAVE_QIDUI_HU = false
	//清一色 风一色 将一色 碰碰胡 全球人 杠上开花  没有七对 没有抢杠胡
	_type.HAVE_QING_YISE_HU = true
	_type.HAVE_FENG_YI_SE = true
	_type.HAVE_JIANG_JIANG_HU = true
	_type.HAVE_FENG_YISE_HU = true
	_type.HAVE_QUAN_QIU_REN = true
	_type.HAVE_PENG_PENG_HU = true
	_type.HAVE_GANG_SHANG_KAI_HUA = true
}

// ! 获取游戏配置
func (sp *SportJZCXZ) GetGameConfig() *static.GameConfig { //获取游戏相关配置
	return &sp.Config
}

// ! 重置桌子数据
func (sp *SportJZCXZ) RepositTable() {
	rand.Seed(time.Now().UnixNano())
	//游戏变量
	sp.SiceCount = components2.MAKEWORD(byte(1), byte(1))

	//出牌信息
	sp.OutCardData = 0
	sp.OutCardCount = 0
	sp.OutCardUser = static.INVALID_CHAIR

	//发牌信息
	sp.SendCardData = 0
	sp.SendCardCount = 0

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

	//结束信息
	sp.ChiHuCard = 0

}

// ! 解析配置的任务
func (sp *SportJZCXZ) ParseRule(strRule string) {
	xlog.Logger().Debug("parserRule :" + strRule)
	sp.Rule.CreateType = 0
	sp.Rule.FangZhuID = sp.GetTableInfo().Creator
	sp.Rule.CreateType = sp.FriendInfo.CreateType
	sp.Rule.JuShu = sp.GetTableInfo().Config.RoundNum //几局
	sp.Rule.AnGang = false                            //暗杠不可抢（其他杠也不能抢，没有抢杠胡）
	if len(strRule) == 0 {
		return
	}

	if strings.Contains(strRule, "overtime_trust") {
		var _msg FriendRule_jzcxz
		if err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg); err == nil {
			sp.Rule.DiFen = _msg.Difen
			sp.Rule.NoWan = false
			//是否去万
			if _msg.Zhangnum == 72 {
				sp.Rule.NoWan = true
			}
			sp.Rule.Jz_SecondRoom = 0 //新版本界面没有几秒场 只有超时托管和超时解散
			sp.Rule.Jz_Wanfa = _msg.Wanfa
			sp.Rule.Jz_Noyy = _msg.Noyy == "true"
			sp.Rule.Jz_No_hdbq = _msg.Noyy == "true"
			sp.Rule.Jz_No_qph = _msg.No_qph == "true"
			sp.Rule.Jz_Fewerstart = _msg.Fewerstart == "true"
			sp.Rule.Jz_JianZiHu = _msg.JianZiHu == "true"
			sp.Rule.Jz_Gpsm = _msg.Gpsm == "true"
			//连续两次赖子杠开才算油中油（第一次杠开了不胡，胡第二次的杠开）
			sp.Rule.Jz_YouZhongYou = _msg.Yzy == "true"
			sp.Rule.Overtime_trust = _msg.Overtimetrust
			sp.Rule.Overtime_dismiss = _msg.Overtimedismiss
			//离线解散时间和申请解散时间
			sp.Rule.Overtime_offdiss = _msg.Overtime_offdiss
			sp.Rule.Overtime_applydiss = _msg.Overtime_applydiss
			sp.Rule.Jz_Gmgd = _msg.Gmgd == "true"
			sp.Rule.Jz_Beishu = _msg.Beishu
			if 0 == sp.Rule.Overtime_trust {
				sp.Rule.Overtime_trust = -1
			}
			if 0 == sp.Rule.Overtime_dismiss {
				sp.Rule.Overtime_dismiss = -1
			}

			sp.Rule.Endready = _msg.Endready == "true"
			if _msg.LookonSupport == "" {
				sp.Config.LookonSupport = true
			} else {
				sp.Config.LookonSupport = _msg.LookonSupport == "true"
			}
			// 出牌时间 超时解散
			sp.OutCardDismissTime = _msg.OutCardDismissTime
			//sp.OutCardDismissTime=30
			// 超时罚分(选择超时罚分后就不能选择15秒场和xx托管)
			// 限制出牌时间后才能选超时罚分
			//sp.OutCardDismissTime=10
			if sp.OutCardDismissTime > 0 {
				sp.Rule.NineSecondRoom = false
				sp.Rule.Overtime_trust = -1
				//20210320 苏大强 暂时没有罚分吧
				//sp.TimeOutPunish = _msg.TimeOutPunish == "true"
				//if sp.TimeOutPunish {
				//	sp.PunishScore = _msg.PunishScore
				//}
			}
			sp.Rule.DissmissCount = _msg.Dissmiss
		}

	} else {
		var _msg FriendRule_jzcxz_old
		if err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg); err == nil {
			sp.Rule.DiFen = _msg.Difen
			sp.Rule.NoWan = false
			//是否去万
			if _msg.Zhangnum == 72 {
				sp.Rule.NoWan = true
			}
			sp.Rule.Jz_SecondRoom = _msg.SecondRoom
			sp.Rule.Jz_Wanfa = _msg.Wanfa
			sp.Rule.Jz_Noyy = _msg.Noyy == "true"
			sp.Rule.Jz_No_hdbq = _msg.Noyy == "true"
			sp.Rule.Jz_No_qph = _msg.No_qph == "true"
			sp.Rule.Jz_Fewerstart = _msg.Fewerstart == "true"
			sp.Rule.Jz_JianZiHu = _msg.JianZiHu == "true"
			sp.Rule.Jz_Gpsm = _msg.Gpsm == "true"
			//连续两次赖子杠开才算油中油（第一次杠开了不胡，胡第二次的杠开）
			sp.Rule.Jz_YouZhongYou = _msg.Yzy == "true"
			sp.Rule.DissmissCount = _msg.Dissmiss
		}
	}

	if 4 == sp.GetPlayerCount() {
		sp.Rule.NoWan = false
	}

	if sp.Rule.DiFen == 0 {
		sp.Rule.DiFen = 1
	}
	//倍数0的话就默认1倍
	if 0 == sp.Rule.Jz_Beishu {
		sp.Rule.Jz_Beishu = 1
	}

	if sp.Rule.DissmissCount != 0 {
		sp.SetDissmissCount(sp.Rule.DissmissCount)
	}

	sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("游戏低分[%d]", sp.Rule.DiFen))
}

// ! 开局
func (sp *SportJZCXZ) OnBegin() {
	xlog.Logger().Debug("onbegin")
	sp.RepositTable()

	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始OnBegin......")

	for _, v := range sp.PlayerInfo {
		v.OnBegin()
	}

	//  第一局随机坐庄
	rand_num := rand.Intn(1000)
	sp.BankerUser = uint16(rand_num % sp.GetPlayerCount())
	sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
	sp.m_GameLogic.Rule = sp.Rule
	sp.m_GameLogic.HuType = sp.HuType
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_GameEnd{}
	sp.VecGameDataAllP = [4][]static.CMD_S_StatusPlay{}

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()
	//恢复离线踢人时间
	sp.SetOfflineRoomTime(0)
	//开始下一局
	sp.OnGameStart()
}

func (sp *SportJZCXZ) OnGameStart() {
	if sp.CanContinue() {
		sp.StartNextGame()
	}
}

// 发送换三张对话框
func (sp *SportJZCXZ) SendExchangeThreeCardSetting() {
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开启了换三张，给所有客户端发送换三张通知......")
	sp.GameEndStatus = static.GS_MJ_PLAY
	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)
	//牌权设置为无效的椅子（防止发完牌以后没有换三张完毕有人出牌）
	sp.CurrentUser = static.INVALID_CHAIR

	for _, v := range sp.PlayerInfo {
		v.Ctx.CleanWeaveItemArray()
		v.Ctx.UserExchangeReady = false
		v.Ctx.UserExchangeThreeCard = [3]byte{}
	}

	var exchangeSetting static.Msg_S_ExchangeThreeSet
	//向每个玩家发送数据
	for _, v := range sp.PlayerInfo {
		exchangeSetting.ExchangeStatus = v.Ctx.UserExchangeReady
		exchangeSetting.CardThree = [3]byte{0, 0, 0}
		sp.SendPersonMsg(consts.MsgTypeGameExchangeSetting, exchangeSetting, v.Seat)
	}
	//发送旁观数据
	exchangeSetting.ExchangeStatus = false
	exchangeSetting.CardThree = [3]byte{0, 0, 0}
	sp.SendTableLookonMsg(consts.MsgTypeGameExchangeSetting, exchangeSetting)
}

// ! 开始下一局游戏
func (sp *SportJZCXZ) StartNextGame() {
	//本来在RepositTable()这个函数里面清理的，但是客户端要求在一小局打完以后解散的时候还要把每个玩家的牌值发给客户端 因此移到这里清理
	for _, v := range sp.PlayerInfo {
		v.Reset()
	}
	sp.OnStartNextGame()
	if 0 == sp.Rule.Overtime_applydiss {
		//解散房间倒计时120秒
		sp.SetDismissRoomTime(120)
	} else {
		sp.SetDismissRoomTime(sp.Rule.Overtime_applydiss)
	}
	//离线踢人时间
	if 0 == sp.Rule.Overtime_offdiss {
		//如果没有设置，那就默认
		sp.SetOfflineRoomTime(0)
	} else {
		sp.SetOfflineRoomTime(sp.Rule.Overtime_offdiss)
	}
	//sp.SetOfflineRoomTime(10)
	sp.LastSendCardUser = uint16(static.INVALID_CHAIR)
	//赖子牌初始化
	sp.MagicCard = 0x00
	//剩余牌数量和剩余牌牌堆
	sp.LeftCardCount = 0
	sp.RepertoryCard = []byte{}
	sp.autoPiZiGangOffset = [static.MAX_PLAYER_4P]int{}
	sp.ReplayRecord.UVitamin = make(map[int64]float64)
	//发送最新状态
	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.SendUserStatus(i, static.US_PLAY) //把状态发给其他人
	}
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始StartNextGame......")
	sp.OnWriteGameRecord(static.INVALID_CHAIR, sp.GetTableInfo().Config.GameConfig)

	for _, v := range sp.PlayerInfo {
		v.OnNextGame()
	}

	//初始化杠分数组
	for i := 0; i < len(sp.m_playerGangFen); i++ {
		for j := 0; j < len(sp.m_playerGangFen[i]); j++ {
			sp.m_playerGangFen[i][j] = 0
		}
	}

	//初始化油中油连续杠
	for i := 0; i < len(sp.magicGangWin); i++ {
		sp.magicGangWin[i] = false
	}
	//上次赖子杠初始化
	sp.mLastGangIsMagicGang = false

	//解析规则
	sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
	sp.m_GameLogic.Rule = sp.Rule
	//设置状态
	sp.SetGameStatus(static.GS_MJ_PLAY)
	// 框架发送开始游戏后开始计算当前这一轮的局数
	sp.CurCompleteCount++
	sp.GetTable().SetBegin(true)

	//混乱扑克
	_randTmp := time.Now().Unix() + int64(sp.GetTableId()+sp.KIND_ID*100+sp.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	sp.SiceCount = components2.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	sp.LeftCardCount = byte(len(sp.RepertoryCard))

	//这里在没有调用混乱扑克的函数时m_cbRepertoryCard中是空的，当它调用了这个函数之后
	//在这个函数中把固定的牌打乱后放到这个数组中，在放的同时不断增加数组m_cbRepertoryCard
	//的长度
	sp.LeftCardCount, sp.RepertoryCard = sp.m_GameLogic.RandCardData()
	sp.CreateLeftCardArray(sp.GetPlayerCount(), int(sp.LeftCardCount), true)

	//先处理赖子位置的控制，后在发牌
	//20201016 苏大强
	gameControl := server2.GetServer().GetGameControl(sp.KIND_ID)
	if gameControl != nil {
		sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("杠开概率为:(%d%%),四赖概率为:(%d%%)", gameControl.ProbGK, gameControl.ProbMagic4))

		if sp.GetProperPNum() < meta2.MAX_PLAYER {

			var (
				Max_HandCard byte = byte((static.MAX_COUNT - 1) * sp.GetPlayerCount())
				Max_Count    byte = sp.LeftCardCount
				Max_Index         = Max_Count - 1
				Fur_Index         = Max_Count - Max_HandCard - 1 - 1
			)
			must := byte(gameControl.MagicBeforeNum)
			if must == 0 {
				sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("癞子配置未开启"))
			} else {
				if must < Max_HandCard {
					sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("配置癞子必须在前%d张牌（不合法），默认置为前%d张。", must, Max_HandCard))
					must = Max_HandCard
				}
				if must >= Max_Count {
					sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("配置癞子必须在前%d张牌（不合法），不用换牌逻辑。", must))
				} else {
					Must_Count := Max_Count - must
					if Must_Count > 0 {
						rdmIndex := func(mc byte) byte {
							var idx byte = static.INVALID_BYTE
							for idx == static.INVALID_BYTE || idx == Fur_Index || idx < Must_Count || idx > Max_Index || sp.RepertoryCard[idx] == mc {
								idx = byte(static.HF_GetRandom(int(must))+1) + Must_Count
								sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("得到和癞子交换得随机坐标为:%d", idx))
							}
							return idx
						}
						if sp.LeftCardCount == Max_Count {
							// 得到癞子牌
							magic := sp.m_GameLogic.MagicByFur(sp.RepertoryCard[Fur_Index])
							sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("推测的癞子牌为:%s", sp.m_GameLogic.SwitchToCardNameByData(magic, 1)))
							for i := byte(0); i < Must_Count; i++ {
								if sp.RepertoryCard[i] == magic {
									sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("在牌堆第%d张牌找到一个癞子牌", i))
									// 得到随机坐标
									rdmIdx := rdmIndex(magic)
									// 交换位置
									sp.RepertoryCard[i], sp.RepertoryCard[rdmIdx] = sp.RepertoryCard[rdmIdx], sp.RepertoryCard[i]
								}
							}
						} else {
							sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("仙桃麻将三人去万玩法混排出来得牌得数量为:%d,而不是:%d,所以没有操作癞子牌靠前。", sp.LeftCardCount, Max_Count))
						}
					} else {
						sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("癞子配置不合法"))
					}
				}
			}
		}
	} else {
		sp.OnWriteGameRecord(static.INVALID_CHAIR, "获取游戏控制配置为空")
	}

	//分发扑克--即每一个人解析他的14张牌结果存放在m_cbCardIndex[i]中
	for _, v := range sp.PlayerInfo {
		if v.Seat != static.INVALID_CHAIR {
			sp.LeftCardCount -= (static.MAX_COUNT - 1)
			v.Ctx.SetCardIndex(&sp.Rule, sp.RepertoryCard[sp.LeftCardCount:], static.MAX_COUNT-1)
		}
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	//sp.InitDebugCards("mahjongJzcxz_test", &sp.RepertoryCard, &sp.BankerUser)
	newleftcard, _ := sp.InitDebugCards_ex("mahjongJzcxz_test", &sp.RepertoryCard, &sp.BankerUser)
	if newleftcard != 0 {
		sp.LeftCardCount = newleftcard
	}
	//////////////读取配置文件设置牌型end////////////////////////////////////
	//获取皮子值（这个皮子会从牌堆中删除，皮子牌只有三张）
	index := sp.LeftCardCount - 2
	sp.PiZiCard = sp.RepertoryCard[index]

	//转换赖子值
	cbValue := byte(sp.PiZiCard & static.MASK_VALUE)
	cbColor := byte(sp.PiZiCard & static.MASK_COLOR)

	//抽到9同花色1是赖子 值等于9 但是花色是万 同 条
	if cbValue == 9 && cbColor <= 0x20 {
		cbValue = 0
		sp.MagicCard = (cbValue + 1) | cbColor
	} else {
		sp.MagicCard = (cbValue + 1) | cbColor
	}

	sp.m_GameLogic.SetMagicCard(sp.MagicCard)
	sp.m_GameLogic.SetPiZiCard(sp.PiZiCard)

	//把第十四张牌发给庄家
	_userItem := sp.GetUserItemByChair(sp.BankerUser)
	//_userItem.Ctx.DispatchCard(sp.SendCardData)
	if _userItem != nil {
		sp.SendOne(_userItem)
	} else {
		static.HF_CheckErr(errors.New(fmt.Sprintf("庄家不存在:%d", sp.BankerUser)))
	}
	sp.LeftCardCount-- //翻的皮子要去掉

	//写游戏日志
	sp.WriteGameRecord()

	//设置变量
	sp.ProvideCard = 0
	sp.ProvideCard = sp.SendCardData
	sp.ProvideUser = static.INVALID_CHAIR
	sp.CurrentUser = sp.BankerUser //供应用户
	sp.LastSendCardUser = sp.BankerUser
	//构造数据,发送开始信息
	var GameStart static.Msg_S_GameStart
	GameStart.SiceCount = sp.SiceCount
	GameStart.BankerUser = sp.BankerUser
	GameStart.CurrentUser = sp.CurrentUser
	GameStart.MagicCard = sp.PiZiCard
	GameStart.LeftCardCount = sp.LeftCardCount
	GameStart.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	GameStart.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	GameStart.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	GameStart.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(GameStart.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)

	//开始新的一局记录(提出来)
	sp.ReplayRecord.Reset()
	//记录癞子牌
	sp.ReplayRecord.PiziCard = sp.PiZiCard
	sp.ReplayRecord.LeftCardCount = sp.LeftCardCount
	sp.ReplayRecord.LeftCard = 0x00

	if 0 < sp.Rule.Overtime_trust { //托管倒计时
		sp.LockTimeOut(sp.BankerUser, sp.Rule.Overtime_trust)
		GameStart.Overtime = sp.LimitTime
	} else if sp.OutCardDismissTime > 0 {
		sp.LockTimeOut(sp.BankerUser, sp.OutCardDismissTime)
		GameStart.Overtime = sp.LimitTime
	} else if 12 == sp.Rule.Jz_SecondRoom {
		sp.LockTimeOut(sp.BankerUser, static.GAME_OPERATION_TIME_12)
		GameStart.Overtime = sp.LimitTime
	} else if 15 == sp.Rule.Jz_SecondRoom {
		sp.LockTimeOut(sp.BankerUser, static.GAME_OPERATION_TIME_15)
		GameStart.Overtime = sp.LimitTime
	} else {
		GameStart.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME
	}

	//向每个玩家发送数据
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		GameStart.PlayerFan[i] = sp.GetPlayerFan(uint16(i))
		//设置变量
		GameStart.UserAction = _item.Ctx.UserAction //把上面分析过的结果保存再发送到客户端
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
		//发送数据
		sp.SendPersonMsg(consts.MsgTypeGameStart, GameStart, uint16(i))
	}
	//发送旁观数据
	GameStart.SendCardData = static.INVALID_BYTE
	GameStart.CardData = [14]byte{}
	GameStart.UserAction = 0
	sp.SendTableLookonMsg(consts.MsgTypeGameStart, GameStart)

	//动作分析,只分析庄家的牌，只有庄家初始才可以杠
	var GangCardResult static.TagGangCardResult
	//判断杠
	_userItem.Ctx.UserAction |= sp.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex, nil, 0, &GangCardResult, _userItem.Ctx.VecGangCard[:])
	//判断能不能
	//判断下庄家能不能胡牌
	_userItem.Ctx.UserAction |= sp.CheckHu(sp.BankerUser, sp.BankerUser, 0, false)

	if _userItem.Ctx.UserAction != 0 {
		sp.ResumeUser = sp.CurrentUser
		sp.SendOperateNotify()
	}
}

// ! 初始化游戏
func (sp *SportJZCXZ) OnInit(table base2.TableBase) {
	sp.KIND_ID = table.GetTableInfo().KindId
	sp.Config.StartMode = static.StartMode_FullReady
	sp.Config.PlayerCount = 3 //玩家人数
	sp.Config.ChairCount = 3  //椅子数量
	sp.PlayerInfo = make(map[int64]*components2.Player)
	sp.HuTypeInit(&sp.HuType) //设置可胡牌类型

	sp.RepositTable()
	sp.SetGameStartMode(static.StartMode_FullReady)
	sp.GameTable = table
	sp.Init()
	sp.Unmarsha(table.GetTableInfo().GameInfo)
	table.GetTableInfo().GameInfo = ""
	//游戏开始前离线踢人
	//游戏未开始离线踢人时间设置
	if sp.GetGameStatus() == static.GS_MJ_FREE {
		var _msg FriendRule_jzcxz
		if err := json.Unmarshal(static.HF_Atobytes(table.GetTableInfo().Config.GameConfig), &_msg); err == nil {
			if 0 == _msg.FleeTime {
				//如果选择的是不踢人 ，那就设置成一个小时吧 一个小时实际也就是不踢人了，没有人会等这么久
				sp.SetOfflineRoomTime(3600)
			} else {
				sp.SetOfflineRoomTime(_msg.FleeTime)
			}
		}
	}
	sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
	sp.m_GameLogic.Rule = sp.Rule
	sp.m_GameLogic.HuType = sp.HuType
}

// ! 发送消息
func (sp *SportJZCXZ) OnMsg(msg *base2.TableMsg) bool {
	switch msg.Head {
	case consts.MsgTypeGameBalanceGameReq: //! 请求总结算信息
		var _msg static.Msg_C_BalanceGameEeq
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.CalculateResultTotal_Rep(&_msg)
		}
	case consts.MsgTypeGameOutCard: //! 出牌消息
		var _msg static.Msg_C_OutCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			return sp.OnUserOutCard(&_msg)
		}
	case consts.MsgTypeGameOperateCard: //操作消息
		var _msg static.Msg_C_OperateCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return sp.OnUserOperateCard(&_msg, true)
		}
	case consts.MsgTypeGameGoOnNextGame: //下一局
		var _msg static.Msg_C_GoOnNextGame
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.OnUserClientNextGame(&_msg)
		}
	case consts.MsgCommonToGameContinue:
		opt, ok := msg.V.(*static.TagSendCardInfo)
		if ok {
			sp.DispatchCardData(opt.CurrentUser, opt.GangFlower, opt.GangFlower)
		} else {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "common to game 断言失败。")
		}
	case consts.MsgTypeGameTrustee: //用户托管
		//var _msg public.Msg_C_Trustee
		var _msg static.Msg_S_DG_Trustee
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.onUserTustee(&_msg)
		}
	case consts.MsgTypeGameChangeSeat: //申请换座
		wChairID := sp.GetChairByUid(msg.Uid)
		if sp.GameEndStatus == byte(static.GS_FREE) {
			//若申请换座玩家已准备
			item := sp.GetUserItemByChair(uint16(wChairID))
			//if item !=nil && item.Ready{
			//	sp.OnWriteGameRecord(wChairID, fmt.Sprintf("OnMsg::准备状态下无法换座:%d", msg.Uid))
			//	sp.SendGameNotificationMessage(wChairID,fmt.Sprintf("准备状态下无法换座"))
			//	break
			//}
			var _msg Msg_C_JZ_CHANGESEAT
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

// ! 下一局
func (sp *SportJZCXZ) OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu || sp.GetGameStatus() != static.GS_MJ_END {
		return true
	}

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()

	nChiarID := sp.GetChairByUid(msg.Id)

	sp.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameGoOnNextGame, msg)

	if nChiarID >= 0 && nChiarID < uint16(sp.GetPlayerCount()) {
		_item := sp.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
		if sp.OutCardDismissTime > 0 {
			sp.UnLockTimeOut(_item.Seat)
		}
	}
	sp.SendUserStatus(int(nChiarID), static.US_READY) //把我的状态发给其他人

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
func (sp *SportJZCXZ) initChiHuResult() {
	for _, v := range sp.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// ! 清除单个玩家记录
func (sp *SportJZCXZ) ClearChiHuResultByUser(wCurrUser uint16) {
	for _, v := range sp.PlayerInfo {
		if v.GetChairID() == wCurrUser {
			v.Ctx.InitChiHuResult()
			break
		}
	}
}

// ! 反向清除单个玩家记录
func (sp *SportJZCXZ) ClearChiHuResultByUserReverse(wCurrUser uint16) {
	for _, v := range sp.PlayerInfo {
		if v.GetChairID() != wCurrUser {
			v.Ctx.InitChiHuResult()
		}
	}
}

// ! 用户操作牌
func (sp *SportJZCXZ) OnUserOperateCard(msg *static.Msg_C_OperateCard, lock bool) bool {
	//if sp.Rule.Overtime_trust!=-1&&lock {
	//	sp.OperateMutex.Lock()
	//	defer sp.OperateMutex.Unlock()
	//}
	_userItem := sp.GetUserItemByUid(msg.Id)
	wChairID := sp.GetChairByUid(msg.Id)
	if _userItem == nil {
		return false
	}

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}

	//效验用户
	if (_userItem.Seat != sp.CurrentUser) && (sp.CurrentUser != static.INVALID_CHAIR) {
		return false
	}

	if msg.Code != static.WIK_NULL {
		// 解锁用户超时操作
		sp.UnLockTimeOut(_userItem.Seat)
	}

	//能碰不碰加入过碰
	if (_userItem.Ctx.UserAction&static.WIK_PENG) != 0 && msg.Code != static.WIK_PENG && sp.CurrentUser == static.INVALID_CHAIR {
		_userItem.Ctx.VecPengCard = append(_userItem.Ctx.VecPengCard, sp.ProvideCard)
		sp.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("能碰不碰，加入过碰,牌:%s", sp.m_GameLogic.SwitchToCardNameByData(sp.ProvideCard, 1)))
	}

	//游戏记录
	//if msg.Code == public.WIK_NULL {
	//	sp.OnWriteGameRecord(_userItem.Seat, "点击弃！")
	//	sp.UnLockTimeOut(_userItem.Seat)
	//}

	if !msg.ByClient {
		sp.OnWriteGameRecord(wChairID, "点击弃！")
		if !msg.ByClient {
			//进入托管状态
			if !_userItem.CheckTRUST() {
				//var msg= &public.Msg_C_Trustee{
				//	Id :_userItem.Uid,
				//	Trustee :true,
				//}
				var msg = &static.Msg_S_DG_Trustee{
					ChairID: _userItem.Seat,
					Trustee: true,
				}
				if sp.onUserTustee(msg) {
					sp.OnWriteGameRecord(wChairID, "操作超时进入托管")
				}
			}
		}
	}

	// 回放中记录牌权操作
	if !(msg.Code == static.WIK_GANG && (msg.Card == sp.MagicCard)) {
		// 回放中记录牌权操作
		sp.addReplayOrder(_userItem.Seat, info2.E_HandleCardRight, msg.Code)
	}

	//被动动作,被动操作没有红中杠，赖子杠, 别人点杠 不分析抢杠
	if sp.CurrentUser == static.INVALID_CHAIR {
		sp.OnUserOperateInvalidChair(msg, _userItem)
		return true
	}

	//主动动作，杠的是赖子和暗杠
	if sp.CurrentUser == _userItem.Seat {
		sp.OnUserOperateByChair(msg, _userItem)
		return true
	}

	return false
}

// ! 被动动作，别人打牌吃碰杠胡牌
func (sp *SportJZCXZ) OnUserOperateInvalidChair(msg *static.Msg_C_OperateCard, _userItem *components2.Player) bool {

	wTargetUser := sp.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card

	//效验状态
	if _userItem.Ctx.Response {
		return false
	}
	if (cbOperateCode != static.WIK_NULL) && ((_userItem.Ctx.UserAction & cbOperateCode) == 0) {
		return false
	}
	if cbOperateCard != sp.ProvideCard {
		return false
	}

	//变量定义
	cbTargetAction := cbOperateCode
	//设置变量
	_userItem.Ctx.SetOperate(cbOperateCard, cbOperateCode)
	if cbOperateCard == 0 {
		_userItem.Ctx.SetOperateCard(sp.ProvideCard)
	}

	//执行判断
	for _, v := range sp.PlayerInfo {
		//获取动作
		cbUserAction := v.Ctx.UserAction

		if v.Ctx.Response {
			cbUserAction = v.Ctx.PerformAction
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
	//被动操作一定不是赖子杠，赖子杠是主动操作
	sp.mLastGangIsMagicGang = false
	//变量定义
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
			v.Ctx.ClearOperateCard()
		}

		//放弃操作
		if _userItem = sp.GetUserItemByChair(sp.ResumeUser); _userItem != nil && _userItem.Ctx.PerformAction != static.WIK_NULL {
			wTargetUser = sp.ResumeUser
			cbTargetAction = _userItem.Ctx.PerformAction
		} else {
			if sp.GetValidCount() > 0 {
				_targetUserItem := sp.GetUserItemByChair(wTargetUser)
				if (_targetUserItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_QIANG_GANG) != 0 {
					sp.DispatchCardData(sp.ResumeUser, cbTargetCard == sp.MagicCard, true)
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
		//胡牌操作
		for tempIndex := 0; tempIndex < sp.GetPlayerCount(); tempIndex++ {
			wUser := uint16(sp.GetNextFullSeat(sp.ProvideUser + uint16(tempIndex)))

			if _item := sp.GetUserItemByChair(wUser); _item != nil {
				//找到的第一个离放炮的用户最近并且有胡牌操作的用户
				if _item.Ctx.UserAction&static.WIK_CHI_HU != 0 {
					wTargetUser = wUser
					_userItem = _item
					if _userItem.Ctx.OperateCard == 0 {
						_userItem.Ctx.SetOperateCard(sp.ProvideCard)
					}
					break
				}
			}
		}

		//结束信息
		sp.ChiHuCard = cbTargetCard
		sp.ProvideUser = sp.ProvideUser

		//插入扑克
		if _userItem.Ctx.ChiHuResult.ChiHuKind != static.CHK_NULL {
			_userItem.Ctx.DispatchCard(sp.ChiHuCard)
		}

		//清除别人胡牌的牌权
		sp.ClearChiHuResultByUserReverse(_userItem.GetChairID())

		//游戏记录
		recordStr := fmt.Sprintf("%s，胡牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
		sp.OnWriteGameRecord(wTargetUser, recordStr)

		sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)

		return true
	} else {
		//用户状态
		for _, v := range sp.PlayerInfo {
			v.Ctx.ClearOperateCard()
		}

		//组合扑克
		wIndex := int(_userItem.Ctx.WeaveItemCount)
		_userItem.Ctx.WeaveItemCount++
		_provideUser := sp.ProvideUser
		if sp.ProvideUser == static.INVALID_CHAIR {
			_provideUser = wTargetUser
		}

		//荆州搓虾子的特殊处理，三个皮子算杠（但是这种杠只是一种界面显示，不补排，实际还是碰。服务器给客户端发了杠以后客户端发回来的也是杠，内部要转成碰）
		if static.WIK_GANG == cbTargetAction {
			if cbOperateCard == sp.m_GameLogic.PiZiCard {
				cbTargetAction = static.WIK_PENG
				_userItem.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, static.WIK_GANG, cbTargetCard)
			} else {
				_userItem.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, cbTargetAction, cbTargetCard)
			}
		} else {
			_userItem.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, cbTargetAction, cbTargetCard)
		}

		//删除扑克
		switch cbTargetAction {
		case static.WIK_LEFT: //左吃操作
			sp.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case static.WIK_RIGHT: //中吃操作
			sp.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case static.WIK_CENTER: //右吃操作
			sp.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case static.WIK_PENG: //碰牌操作
			sp.operateCard(cbTargetAction, cbTargetCard, _userItem) //删除扑克
		case static.WIK_GANG: //杠牌操作
			{
				//删除扑克
				_userItem.Ctx.ShowGangAction()
				cbRemoveCard := []byte{cbTargetCard, cbTargetCard, cbTargetCard}
				_userItem.Ctx.RemoveCards(&sp.Rule, cbRemoveCard)

				//游戏记录
				recordStr := fmt.Sprintf("%s，杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbTargetCard, 1))
				sp.OnWriteGameRecord(wTargetUser, recordStr)

				//记录对这个放杠的玩家的杠分(点杠低分 乘以 倍数)
				gangFen := sp.Rule.DiFen * sp.Rule.Jz_Beishu
				sp.m_playerGangFen[_userItem.GetChairID()][sp.ProvideUser] += gangFen
				sp.m_playerGangFen[sp.ProvideUser][_userItem.GetChairID()] -= gangFen

				//记录赢家的分数
				winGangItem := sp.GetUserItemByChair(uint16(_userItem.GetChairID()))
				if nil != winGangItem {
					winGangItem.Ctx.GangScore += gangFen
				}
				//记录输家的分数
				loseGangItem := sp.GetUserItemByChair(sp.ProvideUser)
				if nil != winGangItem {
					loseGangItem.Ctx.GangScore += (0 - gangFen)
				}

				//记录杠牌
				sp.addReplayOrder(wTargetUser, info2.E_Gang, cbTargetCard)
				//if !sp.Rule.Jz_Gpsm{
				sp.GangFlower = true
				//}
			}
		}

		//构造结果
		var OperateResult static.Msg_S_OperateResult
		OperateResult.OperateUser = wTargetUser
		OperateResult.OperateCard = cbTargetCard
		OperateResult.OperateCode = cbTargetAction
		//荆州搓虾子的特殊处理，三个皮子算杠（但是这种杠只是一种界面显示，不补排，实际还是碰。服务器给客户端发了杠以后客户端发回来的也是杠，内部要转成碰）
		if static.WIK_PENG == cbTargetAction {
			if cbOperateCard == sp.m_GameLogic.PiZiCard {
				OperateResult.OperateCode = static.WIK_GANG
			}
		}

		OperateResult.ProvideUser = sp.ProvideUser
		if 0 < sp.Rule.Overtime_trust {
			sp.LockTimeOut(_userItem.GetChairID(), sp.Rule.Overtime_trust)
			OperateResult.Overtime = sp.LimitTime
		} else if sp.OutCardDismissTime > 0 {
			sp.LockTimeOut(_userItem.GetChairID(), sp.OutCardDismissTime)
			OperateResult.Overtime = sp.LimitTime
		} else if 12 == sp.Rule.Jz_SecondRoom {
			sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME_12)
			OperateResult.Overtime = sp.LimitTime
		} else if 15 == sp.Rule.Jz_SecondRoom {
			sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME_15)
			OperateResult.Overtime = sp.LimitTime
		} else {
			OperateResult.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME
		}

		gangFen := 0
		if cbTargetAction == static.WIK_GANG {
			gangFen = sp.Rule.DiFen * sp.Rule.Jz_Beishu
			OperateResult.ScoreOffset[uint16(_userItem.GetChairID())] = gangFen
			OperateResult.ScoreOffset[_provideUser] = 0 - gangFen
		} else if cbTargetAction == static.WIK_PENG {
			//荆州搓虾子三个皮子算一个杠
			if cbTargetCard == sp.PiZiCard {
				//记录对这个放杠的玩家的杠分(点杠底分)
				gangFen = sp.Rule.DiFen * sp.Rule.Jz_Beishu
				sp.m_playerGangFen[_userItem.GetChairID()][sp.ProvideUser] += gangFen
				sp.m_playerGangFen[sp.ProvideUser][_userItem.GetChairID()] -= gangFen

				//皮子杠记录一下吧，虽然小结算不显示这个
				_userItem.Ctx.GangPizi()

				OperateResult.ScoreOffset[uint16(_userItem.GetChairID())] = gangFen
				OperateResult.ScoreOffset[_provideUser] = 0 - gangFen

				//记录赢家的分数
				winGangItem := sp.GetUserItemByChair(uint16(_userItem.GetChairID()))
				if nil != winGangItem {
					winGangItem.Ctx.GangScore += gangFen
				}
				//记录输家的分数
				loseGangItem := sp.GetUserItemByChair(sp.ProvideUser)
				if nil != winGangItem {
					loseGangItem.Ctx.GangScore += (0 - gangFen)
				}
			}
		}

		if sp.ProvideUser == static.INVALID_CHAIR {
			OperateResult.ProvideUser = wTargetUser
		}

		//操作次数记录
		if sp.ProvideUser != static.INVALID_CHAIR {
			//有人点炮的情况下,增加操作用户的操作次数,并保存第三次供牌的用户
			_userItem.Ctx.AddThirdOperate(sp.ProvideUser)
		}

		OperateResult.HaveGang[wTargetUser] = _userItem.Ctx.HaveGang
		OperateResult.GameScore, OperateResult.GameVitamin = sp.OnSettle(OperateResult.ScoreOffset, consts.EventSettleGaming)
		//发送消息
		sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
		//发送旁观数据
		sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

		// 清空超时检测
		for _, v := range sp.PlayerInfo {
			v.Ctx.CheckTimeOut = 0
		}

		//设置用户
		sp.CurrentUser = wTargetUser
		sp.ProvideCard = 0
		sp.ProvideUser = static.INVALID_CHAIR
		sp.SendCardData = static.INVALID_BYTE
		//最大操作用户操作的是杠牌，进行杠牌处理
		if cbTargetAction == static.WIK_GANG {
			//荆州撮虾子没有抢杠胡
			if sp.GetValidCount() > 0 {
				//只有杠了赖子以后的杠开才有杠开权位 其他不播放杠开
				sp.DispatchCardData(wTargetUser, cbTargetCard == sp.MagicCard, true)
			} else {
				sp.ChiHuCard = 0
				sp.ProvideUser = static.INVALID_CHAIR
				sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
			}
			return true
		}

		//如果是吃碰操作，再判断目标用户是否还有杠牌动作动作判断
		if sp.LeftCardCount > 0 {
			//杠牌判断
			var GangCardResult static.TagGangCardResult

			_item := sp.GetUserItemByChair(sp.CurrentUser)

			_item.Ctx.UserAction |= sp.m_GameLogic.AnalyseGangCard(_item.Ctx.CardIndex,
				_item.Ctx.WeaveItemArray[:], _item.Ctx.WeaveItemCount, &GangCardResult, _item.Ctx.VecGangCard[:])

			//结果处理
			if GangCardResult.CardCount > 0 {
				//设置变量
				_item.Ctx.UserAction |= static.WIK_GANG
				sp.ProvideCard = 0

				//发送动作
				sp.SendOperateNotify()
			}
		}
		return true
	}
	return true
}

// ! 主动动作，自己暗杠痞子杠赖子杠续杠胡牌
func (sp *SportJZCXZ) OnUserOperateByChair(msg *static.Msg_C_OperateCard, _userItem *components2.Player) bool {
	wChairID := sp.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card
	//效验操作
	if cbOperateCode == static.WIK_NULL {
		return true //放弃
	}

	//扑克效验
	if (cbOperateCode != static.WIK_NULL) && (cbOperateCode != static.WIK_CHI_HU) && (sp.m_GameLogic.IsValidCard(cbOperateCard) == false) {
		return false
	}

	//设置变量
	sp.SendStatus = true
	_userItem.Ctx.UserAction = static.WIK_NULL
	_userItem.Ctx.PerformAction = static.WIK_NULL
	//杠后是否补牌
	sendCardAterGang := true
	//执行动作
	switch cbOperateCode {
	case static.WIK_GANG: //杠牌操作
		{
			sp.ClearChiHuResultByUser(_userItem.GetChairID())
			//if !sp.Rule.Jz_Gpsm{
			sp.GangFlower = true
			//}
			gangScore := 0
			bAnGang := false
			//变量定义
			cbWeaveIndex := 0xFF
			//杠的是哪张牌
			cbCardIndex := sp.m_GameLogic.SwitchToCardIndex(cbOperateCard)
			//赖子牌打出去可以做赖子杠，自动补一张牌给该玩家。其他玩家不能碰杠胡赖子。赖子杠会影响玩家胡牌倍数。赖子杠：1个赖子杠2倍。多次累计计算。
			if cbOperateCard == sp.MagicCard {
				//赖子杠
				if _userItem.Ctx.CardIndex[cbCardIndex] == 0 {
					return false
				}
				//构造数据
				var OutCard static.Msg_S_OutCard
				OutCard.User = int(wChairID)
				OutCard.Data = cbOperateCard
				OutCard.ByClient = msg.ByClient
				if -1 != sp.Rule.Overtime_trust {
					OutCard.Overtime = time.Now().Unix() + int64(sp.Rule.Overtime_trust)
				} else if 12 == sp.Rule.Jz_SecondRoom {
					OutCard.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
				} else if 15 == sp.Rule.Jz_SecondRoom {
					OutCard.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME_15
				} else {
					OutCard.Overtime = time.Now().Unix() + static.GAME_OPERATION_TIME
				}

				//发送消息
				sp.SendTableMsg(consts.MsgTypeGameOutCard, OutCard)
				//发送旁观数据
				sp.SendTableLookonMsg(consts.MsgTypeGameOutCard, OutCard)

				sp.OutCardUser = wChairID
				sp.OutCardData = cbOperateCard

				//删除扑克
				if !_userItem.Ctx.OutCard(&sp.Rule, cbOperateCard) {
					xlog.Logger().Debug("removecard failed")
					return false
				}

				//游戏记录
				recordStr := fmt.Sprintf("%s，杠：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sp.OnWriteGameRecord(wChairID, recordStr)
				//一张牌的的杠 只有赖子杠
				if cbOperateCard == sp.MagicCard {
					if _userItem.CheckTRUST() {
						sp.addReplayOrder(wChairID, info2.E_OutCard_TG_Magic, cbOperateCard)
					} else {
						sp.addReplayOrder(wChairID, info2.E_Gang_LaiziGand, cbOperateCard)
					}
					_userItem.Ctx.GangMagic()
				}
				//是赖子杠
				sp.mLastGangIsMagicGang = true
			} else if _userItem.Ctx.CardIndex[cbCardIndex] == 1 {
				//续杠
				for i := byte(0); i < _userItem.Ctx.WeaveItemCount; i++ {
					cbWeaveKind := _userItem.Ctx.WeaveItemArray[i].WeaveKind
					cbCenterCard := _userItem.Ctx.WeaveItemArray[i].CenterCard
					if (cbCenterCard == cbOperateCard) && (cbWeaveKind == static.WIK_PENG) {
						cbWeaveIndex = int(i)
						break
					}
				}

				//如果是弃杠
				bFind := false
				for _, card := range _userItem.Ctx.VecGangCard {
					if card == cbOperateCard {
						bFind = true
						break
					}
				}

				if bFind {
					recordStr := fmt.Sprintf("%s，弃杠了[%x],不能再杠：%s", _userItem.Name, cbOperateCard)
					sp.OnWriteGameRecord(wChairID, recordStr)
					return true
				}

				//效验动作
				if cbWeaveIndex == 0xFF {
					return false
				}
				//游戏记录
				recordStr := fmt.Sprintf("%s，蓄杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sp.OnWriteGameRecord(wChairID, recordStr)
				//记录蓄杠牌
				sp.addReplayOrder(wChairID, info2.E_Gang_XuGand, cbOperateCard)
				//计录明杠牌次数
				_userItem.Ctx.XuGangAction()
				bAnGang = false
				//组合扑克
				_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 1, wChairID, cbOperateCode, cbOperateCard)
				//删除扑克
				_userItem.Ctx.CleanCard(cbCardIndex)

				gangScore = 1 * sp.Rule.DiFen
				//续杠 每人一分
				for i := 0; i < sp.GetPlayerCount(); i++ {
					if uint16(i) != _userItem.GetChairID() {
						sp.m_playerGangFen[_userItem.GetChairID()][i] += gangScore
						sp.m_playerGangFen[i][_userItem.GetChairID()] -= gangScore
					}
				}
				//不是赖子杠
				sp.mLastGangIsMagicGang = false
			} else {
				//皮子是个特殊情况，有三个就可以杠
				if cbOperateCard == sp.PiZiCard {
					if _userItem.Ctx.CardIndex[cbCardIndex] != 3 {
						return false
					}
					sendCardAterGang = false
					//记录暗杠牌
					sp.addReplayOrder(wChairID, info2.E_Gang_ChaoTianGand, cbOperateCard)
				} else {
					if _userItem.Ctx.CardIndex[cbCardIndex] != 4 {
						return false
					}
					//记录暗杠牌
					sp.addReplayOrder(wChairID, info2.E_Gang_AnGang, cbOperateCard)
				}

				//设置变量
				cbWeaveIndex = int(_userItem.Ctx.WeaveItemCount)
				_userItem.Ctx.WeaveItemCount++
				_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 0, wChairID, cbOperateCode, cbOperateCard)

				_userItem.Ctx.HidGangAction()
				bAnGang = true

				//游戏记录
				recordStr := fmt.Sprintf("%s，暗杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
				sp.OnWriteGameRecord(wChairID, recordStr)

				//删除扑克
				_userItem.Ctx.CleanCard(cbCardIndex)

				//暗杠 每人2分
				gangScore = 2 * sp.Rule.DiFen
				for i := 0; i < sp.GetPlayerCount(); i++ {
					if uint16(i) != _userItem.GetChairID() {
						sp.m_playerGangFen[_userItem.GetChairID()][i] += gangScore
						sp.m_playerGangFen[i][_userItem.GetChairID()] -= gangScore
					}
				}
				//朝天杠的话不改变self.mLastGangIsMagicGang的值
			}

			if msg.ByClient {
				if 0 < sp.Rule.Overtime_trust {
					sp.LockTimeOut(_userItem.GetChairID(), sp.Rule.Overtime_trust)
				} else if sp.OutCardDismissTime > 0 {
					sp.LockTimeOut(_userItem.GetChairID(), sp.OutCardDismissTime)
				} else if 12 == sp.Rule.Jz_SecondRoom {
					sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME_12)
				} else if 15 == sp.Rule.Jz_SecondRoom {
					sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME_15)
				} else {
					sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME)
				}
			}

			//构造结果,向客户端发送操作结果
			var OperateResult static.Msg_S_OperateResult
			OperateResult.OperateUser = wChairID
			OperateResult.ProvideUser = wChairID
			OperateResult.HaveGang[wChairID] = _userItem.Ctx.HaveGang //是否杠过
			OperateResult.OperateCode = cbOperateCode
			OperateResult.OperateCard = cbOperateCard
			if -1 != sp.Rule.Overtime_trust {
				OperateResult.Overtime = _userItem.Ctx.CheckTimeOut
			} else {
				OperateResult.Overtime = sp.LimitTime
			}
			for i := 0; i < sp.GetPlayerCount(); i++ {
				userItem := sp.GetUserItemByChair(uint16(i))
				OperateResult.GameScore[i] = sp.getGangFen(uint16(i))
				if uint16(i) == _userItem.GetChairID() {
					OperateResult.ScoreOffset[i] = gangScore * (sp.GetPlayerCount() - 1)
					if nil != userItem {
						userItem.Ctx.GangScore += (gangScore * (sp.GetPlayerCount() - 1))
					}
				} else {
					OperateResult.ScoreOffset[i] = 0 - gangScore
					if nil != userItem {
						userItem.Ctx.GangScore += (0 - gangScore)
					}
				}
			}

			OperateResult.GameScore, OperateResult.GameVitamin = sp.OnSettle(OperateResult.ScoreOffset, consts.EventSettleGaming)
			//发送消息
			sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)

			if bAnGang && !sp.Rule.AnGang {
				if sp.GetValidCount() > 0 {
					//只有杠了赖子以后的杠开才有杠开权位 其他不播放杠开
					if sendCardAterGang {
						sp.DispatchCardData(wChairID, cbOperateCard == sp.MagicCard, true)
					} else {
						//赖子皮杠了以后不补牌，因此还需要判断下是否还有牌权，比如胡杠 需要再次判断下
						var GangCardResult static.TagGangCardResult
						if sp.LeftCardCount > 0 { //有剩余牌才可以杠
							//判断杠
							_userItem.Ctx.UserAction |= sp.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex, nil, 0, &GangCardResult, _userItem.Ctx.VecGangCard[:])
						}
						//判断下玩家能不能胡牌(如果朝天杠前是一个赖子杠，那胡牌的话也算赖油)
						_userItem.Ctx.UserAction |= sp.CheckHu(wChairID, wChairID, 0, sp.mLastGangIsMagicGang)
						//朝天杠以后，因为不补牌，如果还能胡牌的话，那就发送胡牌，但是不发送牌（因为上一次牌的是朝天牌，已经放下了）
						if (_userItem.Ctx.UserAction & static.WIK_CHI_HU) != 0 {
							//客户端要求发个 0
							sp.ProvideCard = 0
							//客户端要求，如果还能胡，就从手牌上找到最后一张牌发给客户端就好了
							for i := len(_userItem.Ctx.CardIndex[:]) - 1; i >= 0; i-- {
								if _userItem.Ctx.CardIndex[i] > 0 {
									sp.SendCardData = sp.m_GameLogic.SwitchToCardData(byte(i))
									break
								}
							}
						}
						if _userItem.Ctx.UserAction != 0 {
							sp.ResumeUser = sp.CurrentUser
							sp.SendOperateNotify()
						}
					}
				} else {
					sp.ChiHuCard = 0
					sp.ProvideUser = static.INVALID_CHAIR
					sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
				}
			} else { //如果不是暗杠的情况，才分析这张牌其他用户是否可以抢杠,调用的是分析用户响应操作
				bAroseAction := false
				if !sp.IsGangCard(cbOperateCard) {
					//允许抢杠胡的条件下才分析抢杠胡(大冶开口番不允许抢杠胡)
					if sp.HuType.HAVE_QIANG_GANG_HU {
						_type_gang := static.EstimatKind_GangCard
						if bAnGang {
							_type_gang = static.EstimatKind_AnGangCard
						}
						bAroseAction = sp.EstimateUserRespond(wChairID, cbOperateCard, _type_gang)
					}
				}

				//发送扑克
				if bAroseAction == false {
					if sp.GetValidCount() > 0 {
						sp.DispatchCardData(wChairID, cbOperateCard == sp.MagicCard, true)
					} else {
						sp.ChiHuCard = 0
						sp.ProvideUser = static.INVALID_CHAIR
						sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
					}
				}
			}
			return true
		}
	case static.WIK_CHI_HU: //吃胡操作,主动状态下没有抢杠的说法，有自摸胡牌，杠上开花胡牌
		{
			// 检查玩家手上是否有3个皮子，如果有则算一杠
			sp.AutoPiZiGang(wChairID, _userItem)
			//普通胡牌
			sp.ClearChiHuResultByUserReverse(_userItem.GetChairID())
			sp.ProvideCard = sp.SendCardData

			if sp.ProvideCard != 0 {
				sp.ProvideUser = wChairID
			}

			//游戏记录
			recordStr := fmt.Sprintf("%s，胡牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(sp.ProvideCard, 1))
			sp.OnWriteGameRecord(wChairID, recordStr)

			//结束信息
			sp.ChiHuCard = sp.ProvideCard

			//结束游戏
			sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)

			return true
		}
	}
	return true
}

func (sp *SportJZCXZ) AutoPiZiGang(wChairID uint16, _userItem *components2.Player) {
	//皮子是个特殊情况，有三个就可以杠
	cbOperateCard := sp.PiZiCard
	//杠的是哪张牌
	cbCardIndex := sp.m_GameLogic.SwitchToCardIndex(cbOperateCard)
	if _userItem.Ctx.CardIndex[cbCardIndex] == 3 {
		//记录暗杠牌
		//sp.addReplayOrder(wChairID, info2.E_Gang_ChaoTianGand, cbOperateCard)
		////设置变量
		//cbWeaveIndex := int(_userItem.Ctx.WeaveItemCount)
		//_userItem.Ctx.WeaveItemCount++
		//_userItem.Ctx.AddWeaveItemArray(cbWeaveIndex, 0, wChairID, static.WIK_GANG, cbOperateCard)
		_userItem.Ctx.HidGangAction()

		//游戏记录
		recordStr := fmt.Sprintf("%s，自摸自动皮子杠牌：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(cbOperateCard, 1))
		sp.OnWriteGameRecord(wChairID, recordStr)
		//删除扑克
		//_userItem.Ctx.CleanCard(cbCardIndex)

		//暗杠 每人2分
		gangScore := 2 * sp.Rule.DiFen
		//for i := 0; i < sp.GetPlayerCount(); i++ {
		//	if uint16(i) != _userItem.GetChairID() {
		//		sp.m_playerGangFen[_userItem.GetChairID()][i] += gangScore
		//		sp.m_playerGangFen[i][_userItem.GetChairID()] -= gangScore
		//	}
		//}

		sp.autoPiZiGangOffset = [static.MAX_PLAYER_4P]int{}
		for i := 0; i < sp.GetPlayerCount(); i++ {
			// userItem := sp.GetUserItemByChair(uint16(i))
			if uint16(i) == _userItem.GetChairID() {
				sp.autoPiZiGangOffset[i] = gangScore * (sp.GetPlayerCount() - 1)
				//if nil != userItem {
				//	userItem.Ctx.GangScore += (gangScore * (sp.GetPlayerCount() - 1))
				//}
			} else {
				sp.autoPiZiGangOffset[i] = 0 - gangScore
				//if nil != userItem {
				//	userItem.Ctx.GangScore += (0 - gangScore)
				//}
			}
		}
	}
}

// ! 操作牌
func (sp *SportJZCXZ) operateCard(cbTargetAction byte, cbTargetCard byte, _userItem *components2.Player) {
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

		//荆州搓虾子如果又可以碰有可以杠 点了碰以后就不能在杠了。但是蓄杠又有个可以随时杠的需求。此处弃杠了就不让他在杠了
		if _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(cbTargetCard)] == 3 {
			//有杠有碰 选择碰 加入弃杠
			_userItem.Ctx.VecGangCard = append(_userItem.Ctx.VecGangCard, cbTargetCard)
		}
	default:
	}

	//判断一下是不是小朝天
	if cbTargetAction == static.WIK_PENG && cbTargetCard == sp.PiZiCard {
		//记录左吃
		sp.addReplayOrder(_userItem.Seat, info2.E_Gang_SmallChaoTianGand, cbTargetCard)
	} else {
		//记录左吃
		sp.addReplayOrder(_userItem.Seat, wik_kind, cbTargetCard)
	}

	//删除扑克
	_userItem.Ctx.RemoveCards(&sp.Rule, cbRemoveCard)

	//if -1 != sp.Rule.Overtime_trust{
	//	sp.LockTimeOut(_userItem.GetChairID(),sp.Rule.Overtime_trust)
	//}else{
	//	if 12 == sp.Rule.Jz_SecondRoom{
	//		sp.LockTimeOut(_userItem.Seat, public.GAME_OPERATION_TIME_12)
	//	}else if 15 == sp.Rule.Jz_SecondRoom{
	//		sp.LockTimeOut(_userItem.Seat, public.GAME_OPERATION_TIME_15)
	//	}
	//}

}

// ! 用户出牌
func (sp *SportJZCXZ) OnUserOutCard(msg *static.Msg_C_OutCard) bool {
	xlog.Logger().Debug("OnUserOutCard")
	//if sp.Rule.Overtime_trust!=-1 {
	//	sp.OperateMutex.Lock()
	//	defer sp.OperateMutex.Unlock()
	//}
	//效验状态
	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return true
	}

	wChairID := sp.GetChairByUid(msg.Id)
	//效验参数
	if wChairID != sp.CurrentUser {
		return false
	}
	if sp.m_GameLogic.IsValidCard(msg.CardData) == false {
		return false
	}

	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return false
	}

	// 重置胡牌权限
	sp.ClearChiHuResultByUser(wChairID)

	// 游戏记录
	recordStr := fmt.Sprintf("%s，打出：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(msg.CardData, 1))
	if !msg.ByClient && sp.Rule.Overtime_trust != -1 {
		recordStr += "(托管)"
		//进入托管状态
		if !_userItem.CheckTRUST() {
			//var msg= &public.Msg_C_Trustee{
			//	Id :_userItem.Uid,
			//	Trustee :true,
			//}
			var msg = &static.Msg_S_DG_Trustee{
				ChairID: _userItem.Seat,
				Trustee: true,
			}
			if sp.onUserTustee(msg) {
				sp.OnWriteGameRecord(wChairID, "出牌超时进入托管")
			}
		}
	}
	sp.OnWriteGameRecord(wChairID, recordStr)
	var class byte = 0
	if _userItem.CheckTRUST() {
		class = 1
	}

	//有这个牌的情况才放到弃牌区
	if _userItem.Ctx.CheckCardExist(msg.CardData) {
		//出牌丢进弃牌区
		_userItem.Ctx.Discard_ex(msg.CardData, class)
	} else {
		return false
	}

	//打出的是赖子杠 赖子打出加两番
	if sp.MagicCard == msg.CardData {
		_msg := sp.Greate_Operatemsg(_userItem.Uid, msg.ByClient, static.WIK_GANG, msg.CardData)
		return sp.OnUserOperateCard(_msg, false)
	}

	//打出的不是赖子，那一定不是连续的油中油了，无条件全部重置为false
	for i := 0; i < len(sp.magicGangWin); i++ {
		sp.magicGangWin[i] = false
	}
	//如果有人出牌了，那就一定不是两个连续的杠了
	sp.mLastGangIsMagicGang = false

	sp.GangFlower = false

	// 解锁用户超时操作
	sp.UnLockTimeOut(wChairID)

	////能杠不杠加入过杠
	if (_userItem.Ctx.UserAction & static.WIK_GANG) != 0 {
		//如果打出的牌不是蓄杠的，就加入弃杠
		for i := byte(0); i < 4; i++ {
			if _userItem.Ctx.WeaveItemArray[i].WeaveKind == static.WIK_PENG {
				if _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(_userItem.Ctx.WeaveItemArray[i].CenterCard)] >= 1 {
					//有回头杠 选择出牌 加入弃杠
					if _userItem.Ctx.AppendGiveUpGang_ex(_userItem.Ctx.WeaveItemArray[i].CenterCard) {
						sp.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("有回头杠 选择出牌 加入弃杠,牌:%s", sp.m_GameLogic.SwitchToCardNameByData(_userItem.Ctx.WeaveItemArray[i].CenterCard, 1)))
					}
				}
			}
		}
	}

	//删除扑克
	if !_userItem.Ctx.OutCard(&sp.Rule, msg.CardData) {
		xlog.Logger().Debug("removecard failed")
		return false
	}

	//设置变量
	sp.SendStatus = true
	//不管有什么有牌权，出牌了就全部都取消掉
	sp.ClearChiHuResultByUser(_userItem.GetChairID())

	//出牌记录
	sp.OutCardCount++
	////////////////出的是一般的牌或赖子牌////////////////
	sp.OutCardUser = wChairID
	sp.OutCardData = msg.CardData

	//构造数据
	var OutCard static.Msg_S_OutCard
	OutCard.User = int(wChairID)
	OutCard.Data = msg.CardData
	OutCard.ByClient = msg.ByClient
	if sp.OutCardDismissTime > 0 {
		sp.LimitTime = time.Now().Unix() + int64(sp.OutCardDismissTime)
	} else {
		if sp.Rule.Overtime_trust > 0 {
			sp.LimitTime = time.Now().Unix() + int64(sp.Rule.Overtime_trust)
		} else {
			if 12 == sp.Rule.Jz_SecondRoom {
				sp.LimitTime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
			} else if 15 == sp.Rule.Jz_SecondRoom {
				sp.LimitTime = time.Now().Unix() + static.GAME_OPERATION_TIME_15
			} else {
				sp.LimitTime = time.Now().Unix() + static.GAME_OPERATION_TIME
			}
		}
	}
	OutCard.Overtime = sp.LimitTime
	if _userItem.CheckTRUST() {
		if sp.MagicCard == msg.CardData { //赖子
			sp.addReplayOrder(wChairID, info2.E_OutCard_TG_Magic, msg.CardData)
		} else { //普通牌
			sp.addReplayOrder(wChairID, info2.E_OutCard_TG, msg.CardData)
		}
	} else {
		sp.addReplayOrder(wChairID, info2.E_OutCard, msg.CardData)
	}

	//发送消息
	sp.SendTableMsg(consts.MsgTypeGameOutCard, OutCard)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameOutCard, OutCard)

	//用户切换
	sp.ProvideUser = wChairID
	sp.ProvideCard = msg.CardData

	sp.CurrentUser = uint16(sp.GetNextFullSeat(wChairID))

	//响应判断，如果用户出的是一般牌，判断其他用户是否需要该牌，EstimatKind_OutCard只是正常出牌判断
	//如果当前用户自己 出了牌，不能自己对自己进行分析吃，碰杠
	bAroseAction := false
	/*
		4人的时候，剩余牌小于或者等于4张的时候（3人的时候就是3张，2张的时候就是2张）

		a.打出的牌，其他人不能碰和杠
		b.自己能起到的牌不能杠
		c.自己起到的牌能胡就可以胡下来
		d.自己手里有赖子可以杠赖子
	*/
	if sp.Rule.Jz_Gmgd {
		if int(sp.GetValidCount()) > sp.GetPlayerCount() {
			bAroseAction = sp.EstimateUserRespond(wChairID, msg.CardData, static.EstimatKind_OutCard)
		}
	} else {
		bAroseAction = sp.EstimateUserRespond(wChairID, msg.CardData, static.EstimatKind_OutCard)
	}

	//打了牌，别人没有反应 流局
	if bAroseAction == false {
		if sp.GetValidCount() > 0 {
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

// ! 超时自动出牌
func (sp *SportJZCXZ) OnAutoOperate(wChairID uint16, bBreakin bool) {

	if bBreakin == false {
		return
	}
	if sp.GetGameStatus() == static.GS_MJ_FREE {
		//sp.UnLockTimeOut(wChairID)
		return
	}

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		//sp.UnLockTimeOut(wChairID)
		return
	}
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}
	byclient := false
	//能胡 胡牌 吃胡
	if (_userItem.Ctx.UserAction&static.WIK_CHI_HU) != 0 && sp.CurrentUser != wChairID {
		_msg := sp.Greate_Operatemsg(_userItem.Uid, byclient, static.WIK_NULL, sp.ProvideCard)
		sp.OnUserOperateCard(_msg, true)
		return
	}
	//点杠 点碰 放弃
	if sp.CurrentUser == static.INVALID_CHAIR && _userItem.Ctx.UserAction != 0 {
		_msg := sp.Greate_Operatemsg(_userItem.Uid, byclient, static.WIK_NULL, sp.ProvideCard)
		sp.OnUserOperateCard(_msg, true)
		return
	}

	//暗杠 擦炮直接放弃出牌
	if sp.CurrentUser == wChairID {
		sp.OnWriteGameRecord(wChairID, fmt.Sprintf("自动出牌，摸的牌: %s, 癞子牌：%s",
			sp.m_GameLogic.SwitchToCardNameByData(sp.SendCardData, 1),
			sp.m_GameLogic.SwitchToCardNameByData(sp.MagicCard, 1),
		))
		cbSendCardData := sp.SendCardData
		if cbSendCardData != sp.MagicCard {
			index := sp.m_GameLogic.SwitchToCardIndex(cbSendCardData) // 出牌索引
			if index >= 0 && index < static.MAX_INDEX {
				if 0 != _userItem.Ctx.CardIndex[index] {
					_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
					sp.OnUserOutCard(_msg)
					return
				}
			}
		}
		for i := byte(static.MAX_INDEX - 1); i > 0; i-- {
			if _userItem.Ctx.CardIndex[i] != 0 {
				cbSendCardData := sp.m_GameLogic.SwitchToCardData(i)
				if cbSendCardData != sp.MagicCard {
					_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
					sp.OnUserOutCard(_msg)
					return
				}
			}
		}
		// 除了癞子 没牌打了
		for i := byte(static.MAX_INDEX - 1); i > 0; i-- {
			if _userItem.Ctx.CardIndex[i] != 0 {
				cbSendCardData := sp.m_GameLogic.SwitchToCardData(i)
				_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
				sp.OnUserOutCard(_msg)
				return
			}
		}
	}
}

// ! 创建操作牌消息
func (sp *SportJZCXZ) Greate_Operatemsg(Id int64, byClient bool, Code byte, Card byte) *static.Msg_C_OperateCard {
	_msg := new(static.Msg_C_OperateCard)
	_msg.Card = Card
	_msg.Code = Code
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 创建出牌消息
func (sp *SportJZCXZ) Greate_OutCardmsg(Id int64, byClient bool, Card byte) *static.Msg_C_OutCard {
	_msg := new(static.Msg_C_OutCard)
	_msg.CardData = Card
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 派发扑克 其中bGangFlower表示是否有杠上开花牌权（只有赖子杠才需要考虑这个） bGang表示是不是杠牌（杠牌顺摸和杠牌后摸客户端需要知道这个）
func (sp *SportJZCXZ) DispatchCardData(wCurrentUser uint16, bGangFlower bool, bGang bool) bool {
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
	//椅子编号合法校验
	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return false
	}
	//剩余牌校验
	if sp.LeftCardCount <= 0 {
		return false
	}

	bEnjoinHu := true
	//发牌处理
	if sp.SendStatus == true {
		//发送扑克
		//sp.SendCardCount++
		//sp.LeftCardCount--
		//sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount]
		//_userItem.Ctx.DispatchCard(sp.SendCardData)

		sp.SendOne(_userItem)
		sp.SetLeftCardArray()
		//游戏记录
		recordStr := fmt.Sprintf("牌型%s，发来：%s", sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.m_GameLogic.SwitchToCardNameByData(sp.SendCardData, 1))
		sp.OnWriteGameRecord(wCurrentUser, recordStr)

		//记录发牌
		sp.addReplayOrder(wCurrentUser, info2.E_SendCard, sp.SendCardData)

		//设置变量
		sp.ProvideUser = wCurrentUser
		sp.ProvideCard = sp.SendCardData
		//给用户发牌后，判断用户是否可以杠牌
		/*
			4人的时候，剩余牌小于或者等于4张的时候（3人的时候就是3张，2张的时候就是2张）

			a.打出的牌，其他人不能碰和杠
			b.自己能起到的牌不能杠
			c.自己起到的牌能胡就可以胡下来
			d.自己手里有赖子可以杠赖子
		*/
		//if sp.Rule.Jz_Gmgd{
		//	if int(sp.GetValidCount()) > sp.GetPlayerCount() {
		//		var GangCardResult public.TagGangCardResult
		//		_userItem.Ctx.UserAction |= sp.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex,
		//			_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult,_userItem.Ctx.VecGangCard[:])
		//	}
		//}else{
		//	if sp.GetValidCount() > 0 {
		//		var GangCardResult public.TagGangCardResult
		//		_userItem.Ctx.UserAction |= sp.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex,
		//			_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult,_userItem.Ctx.VecGangCard[:])
		//	}
		//}
		if sp.GetValidCount() > 0 {
			var GangCardResult static.TagGangCardResult
			_userItem.Ctx.UserAction |= sp.m_GameLogic.AnalyseGangCard(_userItem.Ctx.CardIndex,
				_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult, _userItem.Ctx.VecGangCard[:])
		}
		//发牌了，清楚所有人的胡牌权限
		sp.initChiHuResult()
		//只有一个赖子并且是一脚赖油的情况下才会存在见字胡的情况，
		if !sp.Rule.Jz_JianZiHu && 0 == sp.m_GameLogic.Rule.Jz_Wanfa && byte(1) == sp.m_GameLogic.GetMagicCount(_userItem.Ctx.CardIndex[:]) { //不允许见字胡（所有的玩法如果勾选了不可见字胡，都不能见字胡）
			cbTempCard := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
			copy(cbTempCard, _userItem.Ctx.CardIndex[:])
			cbTempCard[sp.m_GameLogic.SwitchToCardIndex(sp.SendCardData)]--
			tingCount := sp.m_GameLogic.AnalyseTingCardCount(cbTempCard[:], _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, 0)
			tingMax := 27
			if sp.Rule.NoWan {
				tingMax = 18
			}
			if int(tingCount) < tingMax { //胡牌个数小于tingMax个 那就不是见字胡了
				// 见字胡不能胡
				//发了牌以后判断这个人是否可以自摸胡牌
				sp.CheckHu(wCurrentUser, wCurrentUser, 0, bGangFlower)
			}
		} else {
			sp.CheckHu(wCurrentUser, wCurrentUser, 0, bGangFlower)
		}
	}

	//设置变量
	sp.OutCardData = 0
	sp.CurrentUser = wCurrentUser
	sp.OutCardUser = static.INVALID_CHAIR

	//构造数据
	var SendCard static.Msg_S_SendCard
	SendCard.CurrentUser = wCurrentUser
	//自摸 杠牌和胡牌权
	SendCard.ActionMask = _userItem.Ctx.UserAction
	SendCard.ChiHuKindMask = _userItem.Ctx.ChiHuResult.ChiHuKind

	//判断一下是否胡牌权限
	if sp.Rule.Jz_YouZhongYou && bGangFlower {
		//如果有胡牌
		if _userItem.Ctx.UserAction&static.WIK_CHI_HU != 0 {
			for i := 0; i < len(sp.magicGangWin); i++ {
				//找到最近的一个不是true的赋值为true
				if !sp.magicGangWin[i] {
					sp.magicGangWin[i] = true
					break
				}
			}
		} else {
			//不能胡牌 那就不是连续的了，可以无条件全部重置为false
			for i := 0; i < len(sp.magicGangWin); i++ {
				sp.magicGangWin[i] = false
			}
		}
	} else {
		//荆州搓虾子只有赖子杠才存在杠开 其他的杠没有杠开权限，如果不是赖子杠开的情况下无条件全部重置 因为已经不是连续的了
		for i := 0; i < len(sp.magicGangWin); i++ {
			sp.magicGangWin[i] = false
		}
	}
	SendCard.CardData = 0x00
	if sp.SendStatus {
		SendCard.CardData = sp.SendCardData
	}
	//如果是杠牌顺摸 无条件给客户端发false
	if sp.Rule.Jz_Gpsm {
		SendCard.IsGang = false
	} else {
		SendCard.IsGang = bGang
	}

	SendCard.IsHD = false
	SendCard.EnjoinHu = bEnjoinHu
	//20210320 苏大强 还是保留以前的把
	if sp.OutCardDismissTime > 0 {
		sp.LockTimeOut(_userItem.GetChairID(), sp.OutCardDismissTime)
	} else {
		if 0 < sp.Rule.Overtime_trust {
			sp.LockTimeOut(_userItem.GetChairID(), sp.Rule.Overtime_trust)
		} else {
			if 12 == sp.Rule.Jz_SecondRoom {
				sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME_12)
			} else if 15 == sp.Rule.Jz_SecondRoom {
				sp.LockTimeOut(_userItem.GetChairID(), static.GAME_OPERATION_TIME_15)
			} else {
				sp.LimitTime = time.Now().Unix() + static.GAME_OPERATION_TIME
			}
		}
	}

	sp.LastSendCardUser = wCurrentUser
	for _, v := range sp.PlayerInfo {
		if v.GetChairID() != wCurrentUser {
			SendCard.CardData = 0x00
			SendCard.VecGangCard = make([]int, 0)
			SendCard.Overtime = sp.LimitTime
		} else {
			SendCard.CardData = sp.SendCardData
			SendCard.VecGangCard = static.HF_BytesToInts(_userItem.Ctx.VecGangCard)
			SendCard.Overtime = v.Ctx.CheckTimeOut
		}
		sp.SendPersonMsg(consts.MsgTypeGameSendCard, SendCard, uint16(v.GetChairID()))
	}
	//发送旁观数据
	SendCard.CardData = 0x00
	SendCard.VecGangCard = make([]int, 0)
	SendCard.ActionMask = 0
	sp.SendTableLookonMsg(consts.MsgTypeGameSendCard, SendCard)

	//游戏记录
	recordStr := fmt.Sprintf("发送牌权：%x", _userItem.Ctx.UserAction)
	sp.OnWriteGameRecord(wCurrentUser, recordStr)

	// 回放记录中记录牌权显示
	if _userItem.Ctx.UserAction > 0 {
		sp.addReplayOrderWithHuRight(wCurrentUser, info2.E_JZSendCardRight, _userItem.Ctx.UserAction, _userItem.Ctx.ChiHuResult.ChiHuKind)
		//sp.addReplayOrder(wCurrentUser, eve.E_SendCardRight, _userItem.Ctx.UserAction)
	}

	return true
}

// 返回可以利用的牌张数
func (sp *SportJZCXZ) GetValidCount() byte {
	return sp.LeftCardCount
}

func (sp *SportJZCXZ) CheckNeedGuo(_userItem *components2.Player, cbCheckCard byte, level int, kind int) bool {
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
			return mahlib2.Findcard(_userItem.Ctx.VecChiHuCard, cbCheckCard)
		case 1:
			//过碰
			return mahlib2.Findcard(_userItem.Ctx.VecPengCard, cbCheckCard)
		default:
			return false
		}
	}
	return false
}

// ! 响应判断
func (sp *SportJZCXZ) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int) bool {
	//变量定义
	bAroseAction := false

	// 响应判断只需要判断出牌以及续杠
	if EstimatKind != static.EstimatKind_OutCard && EstimatKind != static.EstimatKind_GangCard && EstimatKind != static.EstimatKind_AnGangCard {
		return bAroseAction
	}

	//用户状态
	for _, v := range sp.PlayerInfo {
		v.Ctx.ClearOperateCard()
	}

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
			//碰杠判断（荆州撮虾子不能吃）
			if sp.GetValidCount() > 0 {
				//碰牌判断
				_item.Ctx.UserAction |= sp.m_GameLogic.JZ_EstimatePengCard(_item.Ctx.CardIndex, cbCenterCard)
				//过碰
				if _item.Ctx.UserAction&static.WIK_PENG != 0 && sp.CheckNeedGuo(_item, cbCenterCard, 1, 1) {
					_item.Ctx.UserAction ^= static.WIK_PENG
					sp.SendGameNotificationMessage(_item.GetChairID(), "过碰后不能再碰")
				}
				//杠牌判断
				_item.Ctx.UserAction |= sp.m_GameLogic.JZ_EstimateGangCard(_item.Ctx.CardIndex, cbCenterCard)
			}
		}

		//结果判断
		if _item.Ctx.UserAction != static.WIK_NULL {
			bAroseAction = true
			if sp.OutCardDismissTime > 0 {
				sp.LockTimeOut(_item.GetChairID(), sp.OutCardDismissTime)
			} else {
				if 0 < sp.Rule.Overtime_trust {
					sp.LockTimeOut(_item.GetChairID(), sp.Rule.Overtime_trust)
				} else {
					// 开始用户操作
					if 12 == sp.Rule.Jz_SecondRoom {
						sp.LockTimeOut(uint16(i), static.GAME_OPERATION_TIME_12)
					} else if 15 == sp.Rule.Jz_SecondRoom {
						sp.LockTimeOut(uint16(i), static.GAME_OPERATION_TIME_15)
					} else {
						//没有托管的话，界面倒计时显示9秒开始
						sp.LockTimeOut(_item.GetChairID(), static.GAME_OPERATION_TIME)
					}
				}
			}

		}
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
func (sp *SportJZCXZ) SendOperateNotify() bool {
	//发送提示
	for _, v := range sp.PlayerInfo {
		if v.Ctx.UserAction != static.WIK_NULL {
			//构造数据
			var OperateNotify static.Msg_S_OperateNotify
			OperateNotify.ResumeUser = sp.ResumeUser
			//抢暗杠时，复用此字段，表示轮到谁抢了
			OperateNotify.ActionCard = sp.ProvideCard
			OperateNotify.ActionMask = v.Ctx.UserAction
			OperateNotify.ChiHuKindMask = v.Ctx.ChiHuResult.ChiHuKind
			OperateNotify.EnjoinHu = false
			OperateNotify.VecGangCard = static.HF_BytesToInts(v.Ctx.VecGangCard)
			//if -1 != sp.Rule.Overtime_trust{
			//	sp.LockTimeOut(v.Seat,sp.Rule.Overtime_trust)
			//	OperateNotify.Overtime = time.Now().Unix() + int64(sp.Rule.Overtime_trust)
			//}else{
			//	if 12 == sp.Rule.Jz_SecondRoom{
			//		sp.LockTimeOut(v.Seat,public.GAME_OPERATION_TIME_12)
			//		OperateNotify.Overtime = time.Now().Unix() + public.GAME_OPERATION_TIME_12
			//	}else if 15 == sp.Rule.Jz_SecondRoom{
			//		sp.LockTimeOut(v.Seat,public.GAME_OPERATION_TIME_15)
			//		OperateNotify.Overtime = time.Now().Unix() + public.GAME_OPERATION_TIME_15
			//	}else{
			//		OperateNotify.Overtime = time.Now().Unix() + public.GAME_OPERATION_TIME
			//	}
			//}
			if -1 != sp.Rule.Overtime_trust {
				OperateNotify.Overtime = v.Ctx.LimitedTime
			} else {
				OperateNotify.Overtime = sp.LimitTime
			}
			//发送数据
			sp.SendPersonMsg(consts.MsgTypeGameOperateNotify, OperateNotify, v.Seat)

			// 游戏记录
			recrodStr := fmt.Sprintf("发送牌权：%x", v.Ctx.UserAction)
			sp.OnWriteGameRecord(v.Seat, recrodStr)

			// 回放记录中记录牌权显示
			sp.addReplayOrderWithHuRight(v.Seat, info2.E_JZSendCardRight, v.Ctx.UserAction, v.Ctx.ChiHuResult.ChiHuKind)
		}
	}

	return true
}

// ! 增加回放操作记录
func (sp *SportJZCXZ) addReplayOrder(chairId uint16, operation int, card byte) {
	var order meta2.Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	order.OpreateRight = static.WIK_NULL
	sp.ReplayRecord.VecOrder = append(sp.ReplayRecord.VecOrder, order)
}

// ! 增加回放操作记录
func (sp *SportJZCXZ) addReplayOrderWithHuRight(chairId uint16, operation int, card byte, opreateRight uint64) {
	var order meta2.Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	order.OpreateRight = opreateRight
	sp.ReplayRecord.VecOrder = append(sp.ReplayRecord.VecOrder, order)
}

// ! 检查是否能胡
func (sp *SportJZCXZ) CheckHu(wCurrentUser uint16, wProvideUser uint16, cbCurrentCard byte, bGangFlower bool) byte {
	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return static.WIK_NULL
	}
	magicCardNum := sp.m_GameLogic.GetMagicCount(_userItem.Ctx.CardIndex[:])

	//无赖到底  只能自摸胡，手中不能有赖子，赖子还原了也不行
	if 2 == sp.m_GameLogic.Rule.Jz_Wanfa {
		if magicCardNum > 0 {
			return static.WIK_NULL
		}
	}
	//所有的玩法 有两个（含）以上的赖子，不能胡牌
	if magicCardNum >= 2 {
		return static.WIK_NULL
	} else if 1 == magicCardNum { //只有一个赖子

	}

	//牌型权位
	wChiHuRight := uint16(0)
	//杠开权限判断
	if bGangFlower {
		wChiHuRight |= static.CHR_GANG_SHANG_KAI_HUA
	}

	//给用户发牌后，胡牌判断
	_userItem.Ctx.UserAction |= sp.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:],
		_userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, cbCurrentCard, wChiHuRight, &_userItem.Ctx.ChiHuResult)

	//可以硬胡，判断下半赖和一脚赖油
	if (_userItem.Ctx.UserAction&static.WIK_CHI_HU) != 0 && (1 == magicCardNum) {
		if 1 == sp.m_GameLogic.Rule.Jz_Wanfa { //半赖
			var outMagicCardNum int
			for _, dc := range _userItem.Ctx.DiscardCard {
				if dc == sp.MagicCard {
					outMagicCardNum++
				}
			}
			if outMagicCardNum > 0 {
				// 打过癞子 可以胡
			} else if (_userItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
				// 有一个赖子，赖子还原了，可以胡牌
				//log.Info("半赖玩法 玩家[%s]手上有一个赖子，赖子不当赖子（赖子还原）可以胡牌！",_userItem.Name)
			} else if (_userItem.Ctx.ChiHuResult.ChiHuKind & static.CHK_DA_HU_NOMAGIC) != 0 {
				// 有一个赖子，赖子还原了，可以胡牌
				//log.Info("半赖玩法 玩家[%s]手上有一个赖子，赖子不当赖子（赖子还原）可以胡牌！",_userItem.Name)
			} else {
				_userItem.Ctx.UserAction ^= static.WIK_CHI_HU //有一个赖子，赖子没有还原 不能胡
				//log.Info("半赖玩法 玩家[%s]手上有一个赖子，赖子没有还原，不能胡牌，取消胡牌牌权！",_userItem.Name)
			}
		} else {
			// 一赖到底不用判断 允许一个赖子胡牌
		}
	}

	//返回胡牌类型
	return _userItem.Ctx.UserAction
}

/*
计算每个人的杠分
点杠、暗杠、续杠（弯杠）立即结算。不算番。
点杠：3张杠1张。被杠玩家出2分。同小朝天。（不能先碰后杠）
暗杠：手中4张一样的牌。每人2分。同大朝天。（可以随时杠）（需要亮明）
续杠：碰完牌后补张成杠。每人1分。（可以随时杠）
*/

func (sp *SportJZCXZ) getGangFen(seat uint16) int {
	_checkuserItem := sp.GetUserItemByChair(seat)
	if _checkuserItem == nil {
		return 0
	}

	fanNum := int(0)
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if uint16(i) != seat {
			fanNum += sp.m_playerGangFen[seat][i]
		}
	}

	return fanNum
}

// ! 得到玩家番数
func (sp *SportJZCXZ) GetPlayerFan(wCurrentUser uint16) int {
	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return 0
	}

	FanNum := 0
	return FanNum
}

// ! 单局结算
func (sp *SportJZCXZ) OnGameOver(wChairID uint16, cbReason byte) bool {
	sp.OnEventGameEnd(wChairID, cbReason)
	return true
}

// ! 游戏结束, 流局结束，统计积分
func (sp *SportJZCXZ) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
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
	case static.GER_DISMISS, static.GER_DISMISS_OVERTIME_OUTCARD: //解散游戏
		return sp.OnGameEndDissmiss(wChairID, cbReason, 0)
	case static.GER_GAME_ERROR:
		return sp.OnGameEndDissmiss(wChairID, cbReason, 1)
	}
	return false
}

func (sp *SportJZCXZ) getWinScore(nWinner uint16) int {
	//胡牌玩家
	_userItem := sp.GetUserItemByChair(nWinner)
	if nil == _userItem {
		return 0
	}

	//封顶玩家个数
	FengDingPlayerCount := 0
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if i == int(nWinner) {
			//过滤自己
			continue
		}
		//赢家的番数加上输家的番数
		FanNum := sp.FanScore[nWinner].FanNum[nWinner] + sp.FanScore[nWinner].FanNum[i]
		//封顶番数先设置为软小胡封顶
		fengDingFan := sp.Rule.FengDingType
		//封顶
		if FanNum >= fengDingFan {
			FengDingPlayerCount++
		}
	}
	// 大胡放铳，自摸都是三家给分 小胡放统一家给分
	bOnePlayerGiveMoney := true
	if (_userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 || (_userItem.Ctx.ChiHuResult.ChiHuKind&static.CHK_DA_HU_MAGIC) != 0 || nWinner == sp.ProvideUser {
		bOnePlayerGiveMoney = false
	}

	var GameScore [4]int = [4]int{0, 0, 0, 0}
	for i := 0; i < sp.GetPlayerCount(); i++ {
		GameScore[i] = 0
		//过滤赢家
		if uint16(i) == nWinner {
			continue
		}

		//一个人输分 过滤没有放炮的玩家
		if bOnePlayerGiveMoney {
			if uint16(i) != sp.ProvideUser {
				continue
			}
		}
	}

	//计算赢家的人数
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if uint16(i) != nWinner {
			GameScore[nWinner] += (0 - GameScore[i])
		}
	}
	GameScore[nWinner] = GameScore[nWinner] * sp.Rule.DiFen
	return GameScore[nWinner]
}

// ! 结束，结束游戏
func (sp *SportJZCXZ) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	//设置游戏结束状态
	sp.SetGameStatus(static.GS_MJ_END)

	var _huDetail components2.TagHuCostDetail
	_huDetail.Init()

	//定义变量
	var GameEnd static.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sp.LastSendCardUser
	GameEnd.EndStatus = cbReason
	//赖子牌
	GameEnd.MagicCard = sp.MagicCard
	for i := 0; i < int(sp.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}
	//金顶
	for i := 0; i < sp.GetPlayerCount(); i++ {
		GameEnd.JingDingUser[i] = static.DING_NULL
	}

	//设置承包用户
	GameEnd.Contractor = static.INVALID_CHAIR
	GameEnd.ProvideUser = wChairID
	GameEnd.ChiHuCard = sp.ChiHuCard
	//大冶开口番只有一个赢家
	GameEnd.ChiHuUserCount = 1
	GameEnd.KaiKou = sp.Rule.KouKou
	nWinner := static.INVALID_CHAIR
	//只有一个赢家，循环判断找出赢家
	for tempIndex := 0; tempIndex < sp.GetPlayerCount(); tempIndex++ {
		wUser := uint16(sp.GetNextFullSeat(sp.ProvideUser + uint16(tempIndex)))
		if _item := sp.GetUserItemByChair(wUser); _item != nil {
			//找到的第一个离放炮的用户最近并且有胡牌操作的用户
			if _item.Ctx.ChiHuResult.ChiHuKind != static.WIK_NULL {

				nWinner = _item.Seat
				//记录胡牌
				sp.addReplayOrder(nWinner, info2.E_Hu, sp.ChiHuCard)
				break
			}
		}
	}

	//胡牌玩家
	GameEnd.Winner = nWinner
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if GameEnd.Winner == uint16(i) {
			GameEnd.WWinner[i] = true
		} else {
			GameEnd.WWinner[i] = false
		}
	}
	//胡牌玩家
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
		//保存四家红中杠的次数
		GameEnd.HongzhongGangCount[i] = uint16(_userItem.Ctx.HongZhongGang)
		//保存四家发财杠的次数
		GameEnd.FaCaiGangCount[i] = uint16(_userItem.Ctx.FaCaiGang)
		//保存四家皮子杠的次数
		GameEnd.PiZiGangCount[i] = uint16(_userItem.Ctx.PiZiGangCount)
		//保存四家赖子杠的次数
		GameEnd.MagicGangCount[i] = uint16(_userItem.Ctx.MagicCardGang)
		//保存四家开口次数
		GameEnd.OperateCount[i] = sp.GetUserOpenMouth(uint16(i))
	}

	//判断一下是否是连续杠开(必须两次及以上连续赖子杠开)
	MagicGangWincontinue := false
	if sp.Rule.Jz_YouZhongYou {
		for i := len(sp.magicGangWin) - 1; i >= 1; i-- {
			if sp.magicGangWin[i] && sp.magicGangWin[i-1] {
				MagicGangWincontinue = true
			}
		}
	}

	noMagicHu := false
	gangKaiHu := false
	qingYiSe := false
	qingYiSeDouble := 1
	//胡牌玩家
	_userItem := sp.GetUserItemByChair(nWinner)
	if _userItem != nil {
		//胡牌权位
		GameEnd.ChiHuKind = _userItem.Ctx.ChiHuResult.ChiHuKind
		//是否是硬胡（硬大胡和软大胡）
		if GameEnd.ChiHuKind&static.CHK_PING_HU_NOMAGIC != 0 || GameEnd.ChiHuKind&static.CHK_DA_HU_MAGIC != 0 {
			noMagicHu = true
		}
		//是否是杠开
		if GameEnd.ChiHuKind&static.CHK_GANG_SHANG_KAI_HUA != 0 {
			gangKaiHu = true
		}
		//是否清一色
		if GameEnd.ChiHuKind&static.CHK_QING_YI_SE != 0 {
			qingYiSe = true
		}
		//清一色翻倍
		if sp.m_GameLogic.Rule.QingYiSeJiaBei && qingYiSe {
			qingYiSeDouble = 2
		}
		//没有承包
		GameEnd.Contractor = static.INVALID_CHAIR
		GameEnd.ContractorType = logic2.Contractor_NULL
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if i == int(nWinner) {
				//过滤自己
				continue
			}

			//硬胡
			if noMagicHu {
				if gangKaiHu {
					GameEnd.GameScore[i] = -4 * qingYiSeDouble //硬赖油
				} else {
					GameEnd.GameScore[i] = -2 * qingYiSeDouble //硬自摸
				}
			} else {
				if gangKaiHu {
					GameEnd.GameScore[i] = -2 * qingYiSeDouble //软赖油
				} else {
					GameEnd.GameScore[i] = -1 * qingYiSeDouble //软自摸
				}
			}
		}

		//获取下赢家的赖子杠的个数
		winUser := sp.GetUserItemByChair(nWinner)
		//记下胡牌次数
		winUser.Ctx.HuByChi()
		//计算赢家的分数(赢家的分数等于所有输家的输分之和)
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if uint16(i) != nWinner {
				//赢家的赖子杠次数
				if winUser.Ctx.MagicCardGang > 0 {
					for k := 0; k < int(winUser.Ctx.MagicCardGang); k++ {
						GameEnd.GameScore[i] = GameEnd.GameScore[i] * 2
					}
					//判断是否是油中油（油中油要多乘以2）
					if MagicGangWincontinue && noMagicHu && gangKaiHu {
						GameEnd.GameScore[i] = GameEnd.GameScore[i] * 2
					}
				}
				//输家的赖子杠次数
				loseUser := sp.GetUserItemByChair(uint16(i))
				if loseUser.Ctx.MagicCardGang > 0 {
					for k := 0; k < int(loseUser.Ctx.MagicCardGang); k++ {
						GameEnd.GameScore[i] = GameEnd.GameScore[i] * 2
					}
				}
				GameEnd.GameScore[nWinner] += (0 - GameEnd.GameScore[i])
			}
		}

		// 所得分数乘以底分
		for i := 0; i < sp.GetPlayerCount(); i++ {
			GameEnd.GameScore[i] = GameEnd.GameScore[i] * sp.Rule.DiFen
		}

		//默认软胡
		GameEnd.HardHu = 0
		if noMagicHu {
			GameEnd.HardHu = 1
		}
		//胡牌类型
		if noMagicHu { //硬胡
			if gangKaiHu { //杠开
				if qingYiSe { //清一色
					//硬胡 + 杠开 + 清一色  = 硬清一色赖油
					GameEnd.BigHuKind = GameHu_NoMagic_QingYiSe_Laiyou
					_huDetail.Private(_userItem.Seat, components2.TagYingQingYiSeLaiYou, 1, components2.DetailTypeFirst)
					_userItem.Ctx.HuLaiYou()
				} else {
					//硬胡 + 杠开  = 硬赖油
					GameEnd.BigHuKind = GameHu_NoMagicLaiYou
					_huDetail.Private(_userItem.Seat, components2.TagYingLaiYou, 1, components2.DetailTypeFirst)
					_userItem.Ctx.HuLaiYou()
				}
			} else {
				if qingYiSe {
					//硬胡 + 清一色 = 硬清一色
					GameEnd.BigHuKind = GameHu_NoMagic_QingYiSe
					_huDetail.Private(_userItem.Seat, components2.TagYingQingYiSe, 1, components2.DetailTypeFirst)
				} else {
					//硬胡 = 硬自摸
					GameEnd.BigHuKind = GameHu_NoMagic
					_huDetail.Private(_userItem.Seat, components2.TagYingZiMo, 1, components2.DetailTypeFirst)
				}
			}

			//判断是否是油中油
			if MagicGangWincontinue && noMagicHu && gangKaiHu {
				_huDetail.Private(_userItem.Seat, components2.TagYouZhongYou, 1, components2.DetailTypeFirst)
			}
		} else { //软胡
			if gangKaiHu { //杠开
				if qingYiSe { //清一色
					//软胡 + 杠开 + 清一色  = 软清一色赖油
					GameEnd.BigHuKind = GameHu_Magic_QingYiSe_Laiyou
					_huDetail.Private(_userItem.Seat, components2.TagRuanQingYiSeLaiYou, 1, components2.DetailTypeFirst)
					_userItem.Ctx.HuLaiYou()
				} else {
					//软胡 + 杠开  = 软赖油
					GameEnd.BigHuKind = GameHu_MagicLaiYou
					_huDetail.Private(_userItem.Seat, components2.TagRuanLaiYou, 1, components2.DetailTypeFirst)
					_userItem.Ctx.HuLaiYou()
				}
			} else {
				if qingYiSe {
					//软胡 + 清一色 = 软清一色
					GameEnd.BigHuKind = GameHu_Magic_QingYiSe
					_huDetail.Private(_userItem.Seat, components2.TagRuanQingYiSe, 1, components2.DetailTypeFirst)
				} else {
					//软胡 = 软自摸
					GameEnd.BigHuKind = GameHu_Magic
					_huDetail.Private(_userItem.Seat, components2.TagRuanZiMo, 1, components2.DetailTypeFirst)
				}
			}
		}
	} else {
		//流局
		//慌庄所有人都是0分
		for i := 0; i < sp.GetPlayerCount(); i++ {
			GameEnd.JingDingUser[i] = static.DING_NULL
			GameEnd.GameScore[i] = 0
		}
		GameEnd.ChiHuCard = 0
		GameEnd.ChiHuUserCount = 0
		sp.HaveHuangZhuang = true
		GameEnd.Winner = static.INVALID_CHAIR
	}

	//拼接一下每个人的赖子杠个数(流局了也要显示赖子打出个数)
	for i := 0; i < sp.GetPlayerCount(); i++ {
		userPlayerItem := sp.GetUserItemByChair(uint16(i))
		if userPlayerItem != nil {
			_huDetail.Private(userPlayerItem.Seat, components2.TagMagic, int(userPlayerItem.Ctx.MagicCardGang), components2.DetailTypeADD)
		}
	}

	GameEnd.IsQuit = false
	GameEnd.TheOrder = sp.CurCompleteCount

	//如果有牌，最后一个牌发给客户端
	if sp.LeftCardCount > 0 {
		GameEnd.LastCard = sp.RepertoryCard[0]
	} else {
		GameEnd.LastCard = 0xff
	}

	sp.ReplayRecord.LeftCard = GameEnd.LastCard

	//判断调整分
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		//没有番数的概念，全部设置为
		GameEnd.MaxFSCount[i] = 0
		GameEnd.GameScore[i] += 0
		GameEnd.GameAdjustScore[i] = GameEnd.GameScore[i]
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.StrEnd[i] = _huDetail.GetSeatString(_item.Seat)
		//小结算不加上杠分，大结算上面要显示杠分，这里要加上杠分
		scoreMax := GameEnd.GameScore[i] + sp.getGangFen(uint16(i))
		if scoreMax > 0 {
			if scoreMax > _item.Ctx.MaxScoreUserCount {
				_item.Ctx.SetMaxScore(scoreMax) //统计最大分数
			}
		}

		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
	}
	for i, s := range sp.autoPiZiGangOffset {
		GameEnd.GameScore[i] += s
		GameEnd.GameAdjustScore[i] += s
	}
	GameEnd.UserScore, GameEnd.UserVitamin = sp.OnSettle(GameEnd.GameScore, consts.EventSettleGameOver)
	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SaveGameData()
	//发送信息（荆州搓虾子是一个特殊情况，小结算显示的分数不包含杠分，但是大结算要显示杠分 所以再发送完小结算以后要把杠分加上在存起来）
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}

		GameEnd.GameScore[i] += sp.getGangFen(uint16(i))
	}
	sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算

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
		sp.ReplayRecord.BigHuKind = GameEnd.BigHuKind
		sp.ReplayRecord.ProvideUser = GameEnd.ProvideUser

		//回放只有8三种，没有那么多
		if GameHu_NoMagic_QingYiSe_Laiyou == sp.ReplayRecord.BigHuKind || GameHu_NoMagicLaiYou == sp.ReplayRecord.BigHuKind {
			sp.ReplayRecord.BigHuKind = static.GameYgk //硬杠开
		} else if GameHu_Magic_QingYiSe_Laiyou == sp.ReplayRecord.BigHuKind || GameHu_MagicLaiYou == sp.ReplayRecord.BigHuKind {
			sp.ReplayRecord.BigHuKind = static.GameRgk //软杠开
		} else if GameHu_NoMagic_QingYiSe == sp.ReplayRecord.BigHuKind || GameHu_NoMagic == sp.ReplayRecord.BigHuKind {
			sp.ReplayRecord.BigHuKind = static.GameNoMagicHu //硬胡
		} else if GameHu_Magic_QingYiSe == sp.ReplayRecord.BigHuKind || GameHu_Magic == sp.ReplayRecord.BigHuKind {
			sp.ReplayRecord.BigHuKind = static.GameMagicHu // 软胡
		}
	}

	sp.ReplayRecord.EndInfo = &GameEnd
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
		sp.TableWriteGameDate(int(sp.CurCompleteCount), v, wintype, GameEnd.GameScore[v.Seat])
	}

	//扣房卡
	if sp.CurCompleteCount == 1 {
		sp.TableDeleteFangKa(sp.CurCompleteCount)
	}
	//结束游戏
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu { //局数够了
		sp.CalculateResultTotal(static.GER_NORMAL, wChairID, 0) //计算总发送总结算

		sp.UpdateOtherFriendDate(&GameEnd, false)
		//通知框架结束游戏
		//sp.SetGameStatus(public.GS_MJ_FREE)
		sp.ConcludeGame()

		for _, v := range sp.PlayerInfo {
			//重置托管状态
			v.ChangeTRUST(false)
		}

	} else {
	}

	//输赢谁继续坐庄
	if nWinner != static.INVALID_CHAIR {
		sp.BankerUser = nWinner
	} else {
		//流局了庄家的下一家坐庄
		sp.BankerUser = sp.GetNextFullSeat(sp.BankerUser)
	}

	sp.OnGameEnd()
	if !(int(sp.CurCompleteCount) >= sp.Rule.JuShu) && sp.CurCompleteCount != 0 {
		check := false
		if sp.Rule.Overtime_dismiss != -1 {
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
							sp.SetDismissRoomTime(sp.Rule.Overtime_dismiss)
							sp.OnDismissFriendMsg(item.Uid, msg)
						}
					}
				}
			}
		}
		//if !check&&sp.Rule.Overtime_trust>0{
		//	sp.SetAutoNextTimer(15) //自动开始下一局
		//}
	}

	if int(sp.CurCompleteCount) < sp.Rule.JuShu {
		if sp.Rule.Endready {
			sp.SetAutoNextTimer(10) //自动开始下一局
		} else {
			if sp.OutCardDismissTime > 0 {
				for i := 0; i < sp.GetPlayerCount(); i++ {
					_item := sp.GetUserItemByChair(uint16(i))
					if _item == nil {
						continue
					}
					sp.LockTimeOut(_item.Seat, sp.OutCardDismissTime)
				}
			}
			//sp.SetAutofree(15)
		}
	}

	sp.RepositTable() // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端

	return true

}

// ! 强退，结束游戏
func (sp *SportJZCXZ) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
	//变量定义

	var GameEnd static.Msg_S_GameEnd
	GameEnd.EndStatus = cbReason
	GameEnd.MagicCard = sp.MagicCard
	var _huDetail components2.TagHuCostDetail
	_huDetail.Init()

	//设置变量
	GameEnd.ProvideUser = static.INVALID_CHAIR
	GameEnd.Winner = static.INVALID_CHAIR
	GameEnd.Contractor = static.INVALID_CHAIR
	GameEnd.RecordTime = time.Now().Unix()
	GameEnd.IsQuit = true
	for i := 0; i < int(sp.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}
	for i := 0; i < 4; i++ {
		GameEnd.JingDingUser[i] = static.DING_NULL
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
		_huDetail.Private(_item.Seat, components2.TagQiangCuo, _item.Ctx.QiangScore, components2.DetailTypeADD)
		//玩家番数
		GameEnd.MaxFSCount[i] = uint16(sp.GetPlayerFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount

		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.StrEnd[i] = _huDetail.GetSeatString(_item.Seat)
	}

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

	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SaveGameData()

	if sp.GetGameStatus() != static.GS_MJ_FREE {
		sp.ReplayRecord.EndInfo = &GameEnd
		//数据库写出牌记录
		sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
		// 写完后清除数据
		sp.ReplayRecord.Reset()

		//数据库写分
		for _, v := range sp.PlayerInfo {
			if v.Seat != static.INVALID_CHAIR {
				if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
					GameEnd.GameScore[v.Seat] += sp.getGangFen(v.Seat)
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat])
				} else {
					GameEnd.GameScore[v.Seat] += sp.getGangFen(v.Seat)
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat])
				}
			}
		}
	}
	sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算
	sp.UpdateOtherFriendDate(&GameEnd, true)
	//结束游戏
	sp.CalculateResultTotal(static.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	//sp.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()
	for _, v := range sp.PlayerInfo {
		//重置托管状态
		v.ChangeTRUST(false)
	}
	sp.SetGameStatus(static.GS_FREE)
	return true
}

// ! 解散，结束游戏
func (sp *SportJZCXZ) OnGameEndDissmiss(wChairID uint16, cbReason byte, cbSubReason byte) bool {

	var _huDetail components2.TagHuCostDetail
	_huDetail.Init()
	//变量定义
	var GameEnd static.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sp.LastSendCardUser
	GameEnd.EndStatus = cbReason
	//如果是出牌超时托管，改回去
	if GameEnd.EndStatus == static.GER_DISMISS_OVERTIME_OUTCARD {
		GameEnd.EndStatus = static.GER_DISMISS
	}
	GameEnd.ProvideUser = static.INVALID_CHAIR
	GameEnd.Winner = static.INVALID_CHAIR
	GameEnd.Contractor = static.INVALID_CHAIR
	GameEnd.MagicCard = sp.MagicCard
	GameEnd.RecordTime = time.Now().Unix()
	for i := 0; i < int(sp.LeftCardCount); i++ {
		GameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}
	for i := 0; i < 4; i++ {
		GameEnd.JingDingUser[i] = static.DING_NULL
	}
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
		GameEnd.GameScore[i] += 0
		//玩家番数
		GameEnd.MaxFSCount[i] = 0
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		GameEnd.PaoCount[i] = 0xFF
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.m_GameLogic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
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
	for i, s := range sp.autoPiZiGangOffset {
		GameEnd.GameScore[i] += s
		GameEnd.GameAdjustScore[i] += s
	}
	GameEnd.UserScore, GameEnd.UserVitamin = sp.OnSettle(GameEnd.GameScore, consts.EventSettleGameOver)
	//发送信息
	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SaveGameData()

	switch cbSubReason {
	case 0:
		//荆州玩法杠分不立即结算，但是如果解散了需要把杠分算到大结算里面去（需要判断下状态，以免一局打完解散重复了）
		if sp.GetGameStatus() == static.GS_MJ_PLAY {
			for i := 0; i < sp.GetPlayerCount(); i++ {
				//记录各家的分数
				GameEnd.GameScore[i] = sp.getGangFen(uint16(i))
				userItem := sp.GetPlayerByChair(uint16(i))
				if nil != userItem {
					if GameEnd.GameScore[i] > 0 {
						if GameEnd.GameScore[i] > userItem.Ctx.MaxScoreUserCount {
							userItem.Ctx.SetMaxScore(GameEnd.GameScore[i]) //统计最大分数
						}
					}
				}
			}
			sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算
		}

		//游戏记录
		sp.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

		if sp.GetGameStatus() == static.GS_MJ_PLAY {
			sp.ReplayRecord.EndInfo = &GameEnd
			//数据库写出牌记录
			sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
		}

		// 写完后清除数据
		sp.ReplayRecord.Reset()

		if sp.GetGameStatus() != static.GS_MJ_FREE {
			//数据库写分
			for _, v := range sp.PlayerInfo {
				if v.Seat != static.INVALID_CHAIR {
					if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
						sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat])
					} else {
						sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.Seat])
					}
				}
			}
		}
	case 1:
		sp.OnWriteGameRecord(wChairID, "前面某个时刻程序出错过，需要排查错误日志，无法恢复这局游戏，解散游戏OnGameEndErrorDissmis")
	}

	sp.UpdateOtherFriendDate(&GameEnd, true)
	// 写总计算
	sp.CalculateResultTotal(cbReason, wChairID, cbSubReason)
	//结束游戏
	//sp.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()
	for _, v := range sp.PlayerInfo {
		//重置托管状态
		v.ChangeTRUST(false)
	}
	sp.SetGameStatus(static.GS_FREE)
	return true
}

// ! 解散牌桌
func (sp *SportJZCXZ) OnEnd() {
	if sp.IsGameStarted() {
		sp.OnGameOver(static.INVALID_CHAIR, static.GER_DISMISS)
	}
}

// ! 计算总发送总结算
func (sp *SportJZCXZ) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	sp.TimeEnd = time.Now().Unix() //大局结束时间
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = sp.Rule.FangZhuID
	balanceGame.CurTotalCount = sp.CurCompleteCount //总盘数
	balanceGame.TimeStart = sp.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = sp.TimeEnd
	for i := 0; i < len(sp.VecGameEnd); i++ {
		for j := 0; j < sp.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += sp.VecGameEnd[i].GameScore[j] //总分
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
		// 有打赢家
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
		wintype = static.ScoreKind_pass
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
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			balanceGame.FXScoreUserCount[i] = _userItem.Ctx.MaxScoreUserCount
			//_userItem.Ctx.MaxScoreUserCount这个被初始化成了-99999了（公共模块，这里就特殊处理下，给0吧）
			if balanceGame.FXScoreUserCount[i] < -10000 {
				balanceGame.FXScoreUserCount[i] = 0
			}

			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
			balanceGame.LaiYouHuCount[i] = _userItem.Ctx.LaiYouHuUserCount

			for j := 0; j < len(sp.VecGameEnd); j++ {
				balanceGame.ShowGangCount[i] += int(sp.VecGameEnd[j].ShowGangCount[i] + sp.VecGameEnd[j].XuGangCount[i])
				balanceGame.HidGangCount[i] += int(sp.VecGameEnd[j].HideGangCount[i])
			}

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
		if len(sp.VecGameDataAllP[i]) > 0 {
			gamedataStr = static.HF_JtoA(sp.VecGameDataAllP[i][len(sp.VecGameDataAllP[i])-1])
		}
		sp.SaveLastGameinfo(_userItem.Uid, gameendStr, static.HF_JtoA(balanceGame), gamedataStr)
	}

	//发消息
	sp.SendTableMsg(consts.MsgTypeGameBalanceGame, balanceGame)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameBalanceGame, balanceGame)
	sp.resetEndDate()
}

// ! 重置优秀结束数据
func (sp *SportJZCXZ) resetEndDate() {
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_GameEnd{}

	for _, v := range sp.PlayerInfo {
		v.OnEnd()
	}
}

func (sp *SportJZCXZ) UpdateOtherFriendDate(GameEnd *static.Msg_S_GameEnd, bEnd bool) {

}

// ! 给客户端发送总结算
func (sp *SportJZCXZ) CalculateResultTotal_Rep(msg *static.Msg_C_BalanceGameEeq) {
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
			balanceGame.FXScoreUserCount[i] = 0
			balanceGame.LaiYouHuCount[i] = 0
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
			//胡牌总数
			balanceGame.ChiHuUserCount[i] = _userItem.Ctx.ChiHuUserCount
			balanceGame.ProvideUserCount[i] = _userItem.Ctx.ProvideUserCount
			//最大分数
			balanceGame.FXScoreUserCount[i] = _userItem.Ctx.MaxScoreUserCount
			//_userItem.Ctx.MaxScoreUserCount这个被初始化成了-99999了（公共模块，这里就特殊处理下，给0吧）
			if balanceGame.FXScoreUserCount[i] < -10000 {
				balanceGame.FXScoreUserCount[i] = 0
			}
			//赖油胡牌次数
			balanceGame.LaiYouHuCount[i] = _userItem.Ctx.LaiYouHuUserCount
			balanceGame.HHuUserCount[i] = _userItem.Ctx.BigHuUserCount
			balanceGame.FXScoreUserCount[i] = 0
		}
	}
	balanceGame.End = 1
	//发消息
	sp.SendPersonMsg(consts.MsgTypeGameBalanceGame, balanceGame, sp.GetChairByUid(msg.Id))
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameBalanceGame, balanceGame)
}

// ! 发送游戏开始场景数据
func (sp *SportJZCXZ) sendGameSceneStatusPlay(player *components2.Player) bool {

	if player.LookonTableId > 0 {
		sp.sendGameSceneStatusPlayLookon(player)
		return true
	}

	wChiarID := player.GetChairID()

	if wChiarID == static.INVALID_CHAIR {
		xlog.Logger().Debug("sendGameSceneStatusPlay invalid chair")
		return false
	}
	//取消托管
	player.Ctx.SetTrustee(false)
	var Trustee static.Msg_S_Trustee
	Trustee.Trustee = false
	Trustee.ChairID = wChiarID
	sp.SendTableMsg(consts.MsgTypeGameTrustee, Trustee)
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, Trustee)

	//变量定义
	var StatusPlay static.CMD_S_StatusPlay
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
	//断线重连回来也把弃杠的牌发出去
	NowUserItem := sp.GetUserItemByChair(sp.CurrentUser)
	if NowUserItem != nil {
		StatusPlay.VecGangCard = static.HF_BytesToInts(NowUserItem.Ctx.VecGangCard)
	}

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	for i := 0; i < sp.GetPlayerCount(); i++ {

		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//玩家番数
		StatusPlay.PlayerFan[i] = sp.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.PiZiGangCount
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		StatusPlay.Whotrust[i] = false
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

		//换三张状态
		StatusPlay.ExchangeThreeState[i].ExchangeState = _item.Ctx.UserExchangeReady
		index := 0
		for _, v := range _item.Ctx.UserExchangeThreeCard {
			StatusPlay.ExchangeThreeState[i].ThreeExchagneCard[index] = v
			index++
		}
	}

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = player.Ctx.UserAction
	StatusPlay.ChiHuKindMask = player.Ctx.ChiHuResult.ChiHuKind

	if player.Ctx.Response {
		StatusPlay.ActionMask = static.WIK_NULL
	}

	if (StatusPlay.ActionMask & static.WIK_QIANG) != 0 {
		StatusPlay.ActionCard = byte(wChiarID)
	}

	if player.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = player.Ctx.CheckTimeOut
	} else {
		var CurUserItem *components2.Player
		CurUserItem = sp.GetUserItemByChair(sp.CurrentUser)

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
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			//StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData

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

	return true
}

// ! 发送游戏开始场景数据
func (sp *SportJZCXZ) sendGameSceneStatusPlayLookon(player *components2.Player) bool {

	if player.LookonTableId == 0 {
		return false
	}
	wChiarID := player.GetChairID()
	if int(wChiarID) >= sp.GetPlayerCount() {
		wChiarID = 0
	}
	//是否要获取wChiarID位置真正玩家的信息 ？
	playerOnChair := sp.GetUserItemByChair(wChiarID)

	//变量定义
	var StatusPlay static.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = sp.SiceCount
	StatusPlay.BankerUser = sp.BankerUser
	StatusPlay.CurrentUser = sp.CurrentUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sp.PiZiCard

	StatusPlay.KaiKou = sp.Rule.KouKou
	StatusPlay.RenWuAble = sp.RenWuAble
	StatusPlay.TheOrder = sp.CurCompleteCount
	if sp.CurrentUser == wChiarID {
		StatusPlay.SendCardData = sp.SendCardData
	} else {
		StatusPlay.SendCardData = static.INVALID_BYTE
	}
	//断线重连回来也把弃杠的牌发出去
	NowUserItem := sp.GetUserItemByChair(sp.CurrentUser)
	if NowUserItem != nil {
		StatusPlay.VecGangCard = static.HF_BytesToInts(NowUserItem.Ctx.VecGangCard)
	}

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	for i := 0; i < sp.GetPlayerCount(); i++ {

		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//玩家番数
		StatusPlay.PlayerFan[i] = sp.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.PiZiGangCount
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		StatusPlay.Whotrust[i] = false
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

		//换三张状态
		StatusPlay.ExchangeThreeState[i].ExchangeState = _item.Ctx.UserExchangeReady
		index := 0
		for _, v := range _item.Ctx.UserExchangeThreeCard {
			StatusPlay.ExchangeThreeState[i].ThreeExchagneCard[index] = v
			index++
		}
	}

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = 0    //player.Ctx.UserAction
	StatusPlay.ChiHuKindMask = 0 //player.Ctx.ChiHuResult.ChiHuKind

	if playerOnChair != nil && playerOnChair.Ctx.Response {
		StatusPlay.ActionMask = static.WIK_NULL
	}

	if (StatusPlay.ActionMask & static.WIK_QIANG) != 0 {
		StatusPlay.ActionCard = byte(wChiarID)
	}

	if playerOnChair != nil && playerOnChair.Ctx.CheckTimeOut != 0 {
		StatusPlay.Overtime = playerOnChair.Ctx.CheckTimeOut
	} else {
		var CurUserItem *components2.Player
		CurUserItem = sp.GetUserItemByChair(sp.CurrentUser)

		if CurUserItem != nil && CurUserItem.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = CurUserItem.Ctx.CheckTimeOut
		}

		for _, v := range sp.PlayerInfo {
			if v.GetChairID() == wChiarID {
				continue
			}
			if v.GetChairID() == sp.CurrentUser {
				continue
			}
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			//StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData

	//扑克数据
	if playerOnChair != nil {
		StatusPlay.CardCount, StatusPlay.CardData = sp.m_GameLogic.SwitchToCardData2(playerOnChair.Ctx.CardIndex, StatusPlay.CardData)
	}

	//发送旁观数据
	sp.SendPersonLookonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, player.Uid)

	//发小结消息
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		//发送旁观数据
		sp.SendPersonLookonMsg(consts.MsgTypeGameEnd, gamend, player.Uid)
	}

	//发送解散房间所有玩家的反应
	sp.SendAllPlayerDissmissInfo(player)

	return true
}

func (sp *SportJZCXZ) SaveGameData() {
	//变量定义
	var StatusPlay static.CMD_S_StatusPlay
	//游戏变量
	StatusPlay.SiceCount = sp.SiceCount
	StatusPlay.BankerUser = sp.BankerUser
	StatusPlay.CurrentUser = sp.CurrentUser
	StatusPlay.CellScore = sp.GetCellScore() //m_pGameServiceOption->lCellScore;
	StatusPlay.PiZiCard = sp.PiZiCard

	StatusPlay.KaiKou = sp.Rule.KouKou
	StatusPlay.RenWuAble = sp.RenWuAble
	StatusPlay.TheOrder = sp.CurCompleteCount
	//断线重连回来也把弃杠的牌发出去
	NowUserItem := sp.GetUserItemByChair(sp.CurrentUser)
	if NowUserItem != nil {
		StatusPlay.VecGangCard = static.HF_BytesToInts(NowUserItem.Ctx.VecGangCard)
	}

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	for i := 0; i < sp.GetPlayerCount(); i++ {

		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		//玩家番数
		StatusPlay.PlayerFan[i] = sp.GetPlayerFan(uint16(i))

		StatusPlay.PiGangCount[i] = _item.Ctx.PiZiGangCount
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.VecXiaPao.Status || _item.Ctx.UserPaoReady
		StatusPlay.Whotrust[i] = false
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

		//换三张状态
		StatusPlay.ExchangeThreeState[i].ExchangeState = _item.Ctx.UserExchangeReady
		index := 0
		for _, v := range _item.Ctx.UserExchangeThreeCard {
			StatusPlay.ExchangeThreeState[i].ThreeExchagneCard[index] = v
			index++
		}
	}

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			//StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData

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
		StatusPlay.ActionMask = player.Ctx.UserAction

		if player.Ctx.Response {
			StatusPlay.ActionMask = static.WIK_NULL
		}
		if (StatusPlay.ActionMask & static.WIK_QIANG) != 0 {
			StatusPlay.ActionCard = byte(i)
		}
		if player.Ctx.CheckTimeOut != 0 {
			StatusPlay.Overtime = player.Ctx.CheckTimeOut
		} else {
			var CurUserItem *components2.Player
			CurUserItem = sp.GetUserItemByChair(sp.CurrentUser)

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
				if v.Ctx.UserAction > 0 {
					StatusPlay.Overtime = 0
					break
				}
			}
		}
		//扑克数据
		StatusPlay.CardCount, StatusPlay.CardData = sp.m_GameLogic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
		sp.VecGameDataAllP[i] = append(sp.VecGameDataAllP[i], StatusPlay) //保存，用于汇总计算
	}
}

// 游戏场景消息发送
func (sp *SportJZCXZ) SendGameScene(uid int64, status byte, secret bool) {
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
		sp.SendGameSceneStatusFree(player)
	case static.GS_MJ_PLAY:
		sp.sendGameSceneStatusPlay(player)
	case static.GS_MJ_END:
		sp.sendGameSceneStatusPlay(player)
	}
}

// ! 判断牌是否可以直接杠出去
func (sp *SportJZCXZ) IsGangCard(cbCardData byte) bool {
	//红中 发财 赖子
	if cbCardData == logic2.CARD_HONGZHONG || cbCardData == sp.MagicCard || cbCardData == logic2.CARD_FACAI {
		return true
	}
	return false
}

// ! 游戏退出
func (sp *SportJZCXZ) OnExit(uid int64) {
	sp.Common.OnExit(uid)
}

// ! 定时器
func (sp *SportJZCXZ) OnTime() {
	sp.Common.OnTime()
}

// ! 计时器事件
func (sp *SportJZCXZ) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {

	if dwTimerID == components2.GameTime_TuoGuan && sp.Rule.Overtime_trust != -1 {
		if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
			//到时间了
			sp.OnAutoOperate(TablePerson.Seat, true)
		}
	} else if dwTimerID == components2.GameTime_12 || dwTimerID == components2.GameTime_15 {
		if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
			//到时间了
			sp.OnAutoOperate(TablePerson.Seat, true)
		}
	}
	// 超时解散
	if sp.OutCardDismissTime > 0 && (sp.GetGameStatus() != static.GS_FREE) {

		wTargetUser := static.INVALID_CHAIR
		if sp.GameEndStatus == static.GS_MJ_PLAY {
			if sp.CurrentUser == static.INVALID_CHAIR {
				// 被动动作  吃、碰等操作 等待最大牌权的玩家
				cbTargetActionRank := byte(0)
				nextUser := sp.ResumeUser
				for i := 0; i < sp.GetPlayerCount(); i++ {
					userItem := sp.GetUserItemByChair(nextUser)
					if userItem != nil {
						if userItem.Ctx.Response || userItem.Ctx.UserAction == static.WIK_NULL {
							nextUser = uint16(sp.GetNextSeat(nextUser))
							continue
						}

						cbUserActionRank := sp.m_GameLogic.GetUserActionRank(userItem.Ctx.UserAction)
						if cbUserActionRank > cbTargetActionRank {
							wTargetUser = userItem.Seat
							cbTargetActionRank = cbUserActionRank
						}
					}
					nextUser = uint16(sp.GetNextSeat(nextUser))
				}
				if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
					if wTargetUser != TablePerson.Seat {
						return true
					}
				}
			} else {
				// 主动操作 出牌、杠操作超时
				if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
					wTargetUser = TablePerson.Seat
				}
			}
		}

		if wTargetUser == static.INVALID_CHAIR {
			wTargetUser = 0
			sp.OnGameOver(wTargetUser, static.GER_DISMISS_OVERTIME_OUTCARD)
			sp.OnWriteGameRecord(wTargetUser, "OnAutoOperate 超时操作强制解散!!!")
		} else {
			// 罚分
			//if sp.TimeOutPunish && sp.CurCompleteCount > 1 {
			//	for i := 0; i < sp.GetPlayerCount(); i++ {
			//		if i != int(wTargetUser) {
			//			sp.PlayerPunishScore[i] += sp.PunishScore
			//			sp.PlayerPunishScore[wTargetUser] -= sp.PunishScore
			//		}
			//	}
			//}
			sp.OnGameOver(wTargetUser, static.GER_DISMISS_OVERTIME_OUTCARD)
			sp.OnWriteGameRecord(wTargetUser, "OnAutoOperate 超时操作强制解散!!!")
		}
	}
	return true
}

// ! 玩家开启超时
func (sp *SportJZCXZ) LockTimeOut(cUser uint16, iTime int) {
	if cUser < 0 || cUser > uint16(sp.GetPlayerCount()) {
		return
	}

	_userItem := sp.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}
	if iTime == 0 {
		iTime = static.GAME_OPERATION_TIME
	}
	checktime := iTime
	if sp.Rule.Overtime_trust != -1 || sp.OutCardDismissTime != -1 {
		sp.LimitTime = time.Now().Unix() + int64(checktime)
		if _userItem.CheckTRUST() {
			//托管状态
			checktime = 1
		}
		_userItem.Ctx.CheckTimeOut = time.Now().Unix() + int64(checktime)
		_userItem.Ctx.Timer.SetTimer(components2.GameTime_TuoGuan, checktime)
	} else {
		sp.LimitTime = time.Now().Unix() + int64(checktime)
		_userItem.Ctx.CheckTimeOut = time.Now().Unix() + int64(checktime)
		//12秒场 本来要取消，但是服务器更新的时候要兼容12秒场 所以保留
		if iTime == static.GAME_OPERATION_TIME_12 {
			_userItem.Ctx.Timer.SetTimer(components2.GameTime_12, checktime)
		} else if iTime == static.GAME_OPERATION_TIME_15 {
			_userItem.Ctx.Timer.SetTimer(components2.GameTime_15, checktime)
		} else if iTime == static.GAME_OPERATION_TIME { //普通操作倒计时
			//_userItem.Ctx.GameTimer.SetTimer(modules.GameTime_Nine, checktime)
		}
	}
}

// ! 玩家关闭超时
func (sp *SportJZCXZ) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(sp.GetPlayerCount()) {
		return
	}

	_userItem := sp.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}

	_userItem.Ctx.CheckTimeOut = 0

	if 12 == sp.Rule.Jz_SecondRoom {
		_userItem.Ctx.Timer.KillTimer(components2.GameTime_12)
	} else if 15 == sp.Rule.Jz_SecondRoom {
		_userItem.Ctx.Timer.KillTimer(components2.GameTime_15)
	}

	if sp.Rule.Overtime_trust != -1 || sp.OutCardDismissTime != -1 {
		_userItem.Ctx.Timer.KillTimer(components2.GameTime_TuoGuan)
	}
}

// ! 写日志记录
func (sp *SportJZCXZ) WriteGameRecord() {
	//写日志记录
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始红中赖子杠  发牌......")

	// 玩家手牌
	for i := 0; i < len(sp.PlayerInfo); i++ {
		v := sp.GetUserItemByChair(uint16(i))
		if v != nil {
			handCardStr := fmt.Sprintf("发牌后手牌:%s", sp.m_GameLogic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
			sp.OnWriteGameRecord(uint16(v.Seat), handCardStr)
		}
	}

	// 牌堆牌
	leftCardStr := fmt.Sprintf("牌堆牌:%s", sp.m_GameLogic.SwitchToCardNameByDatas(sp.RepertoryCard[0:sp.LeftCardCount+2], 0))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, leftCardStr)

	//赖子牌
	magicCardStr := fmt.Sprintf("癞子牌:%s", sp.m_GameLogic.SwitchToCardNameByData(sp.MagicCard, 1))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, magicCardStr)
}

// ! 场景保存
func (sp *SportJZCXZ) Tojson() string {
	var _json components2.GameJsonSerializer
	return _json.ToJsonCommon(&sp.Metadata, &sp.Common)
}

// ! 场景恢复
func (sp *SportJZCXZ) Unmarsha(data string) {
	var _json components2.GameJsonSerializer
	if data != "" {
		json.Unmarshal([]byte(data), &_json)

		_json.Unmarsha(&sp.Metadata)
		_json.JsonToStruct(&sp.Common)

		sp.ParseRule(sp.GetTableInfo().Config.GameConfig)
		sp.m_GameLogic.Rule = sp.Rule
		sp.m_GameLogic.HuType = sp.HuType
		sp.m_GameLogic.SetMagicCard(sp.MagicCard)
		sp.m_GameLogic.SetPiZiCard(sp.PiZiCard)
	}
}

// ! 得到某个用户开口的次数,吃，碰，明杠的次数
func (sp *SportJZCXZ) GetUserOpenMouth(wChairID uint16) uint16 {
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return uint16(0)
	}
	//暗杠不算开口
	return uint16(_userItem.Ctx.WeaveItemCount - _userItem.Ctx.HidGang)
}

// 用户托管事件
func (sp *SportJZCXZ) onUserTustee(msg *static.Msg_S_DG_Trustee) bool {
	if sp.Rule.Overtime_trust < 1 {
		return true
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
			//进入托管啥都不用做
			item.ChangeTRUST(true)
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, tuoguan)
		} else if tuoguan.Trustee == false {
			//如果是当前的玩家，那么重新设置一下开始时间
			item.ChangeTRUST(false)
			if tuoguan.ChairID == sp.CurrentUser {
				//sp.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
				_item := sp.GetUserItemByChair(sp.CurrentUser)
				if _item != nil {
					//if time.Now().Unix() < _item.Ctx.CheckTimeOut { //如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < sp.LimitTime
					sp.LockTimeOut(_item.GetChairID(), sp.Rule.Overtime_trust)
					//sp.setLimitedTime(int64(sp.PlayTime) + sp.PowerStartTime - time.Now().Unix() + 1)
					tuoguan.Overtime = _item.Ctx.CheckTimeOut
				}
			}
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			//发送旁观数据
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, tuoguan)
		}
	} else {
		//详细日志
		LogStr := string("托管动作:游戏状态不正确 ")
		sp.OnWriteGameRecord(tuoguan.ChairID, LogStr)
		return false
	}
	return true
}

func (sp *SportJZCXZ) SendOne(userItem *components2.Player) {
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
	} else {
		// 应要求，仙桃麻将癞子杠开时有比杠开的概率
		// 如果中了这个概率则得到玩家所听的牌，从牌堆中找到一张给他
		gameControl := server2.GetServer().GetGameControl(sp.KIND_ID)
		if gameControl != nil {
			rdm_m4 := static.HF_GetRandom(100) + 1
			userTotalMagic := sp.UserTotalMagic(userItem)
			notbeMagic := userTotalMagic >= 3 && gameControl.ProbMagic4 > 0 && rdm_m4 > gameControl.ProbMagic4
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("四赖随机数为:(%d),四赖概率为:(%d%%),玩家共%d个癞子。下张牌能否能为癞子:%t", rdm_m4, gameControl.ProbMagic4, userTotalMagic, !notbeMagic))
			//if sp.GangFlower && (sp.HuType.HAVE_GANG_SHANG_KAI_HUA || sp.HuType.HAVE_LAI_ZI_GANG_KAI) {
			if sp.GangFlower {
				rdm_gk := static.HF_GetRandom(100) + 1
				prob := gameControl.ProbGK
				sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("杠开生成的随机数为:(%d),杠开概率为:(%d%%)", rdm_gk, prob))
				wChiHuRight := uint16(0)
				// 杠开
				//if sp.GangFlower && (sp.HuType.HAVE_GANG_SHANG_KAI_HUA || sp.HuType.HAVE_LAI_ZI_GANG_KAI) {
				if sp.GangFlower {
					sp.OnWriteGameRecord(userItem.GetChairID(), "杠开预判摸牌：打开杠上开花权位")
					// 杠牌导致摸牌 后打开这个开关
					//if sp.MingGangStatus {
					//	// 如果是杠后摸牌  而不是飘癞子后摸牌 则打开杠上开花 胡牌权位
					//	sp.OnWriteGameRecord(userItem.GetChairID(), "杠开预判摸牌：打开刷开权位")
					//	wChiHuRight |= public.CHR_MENG_QING
					//} else {
					// 如果是飘癞子导致的摸牌  则打开刷开 胡牌权位
					wChiHuRight |= static.CHR_GANG_SHANG_KAI_HUA
					//}
				}
				if sp.HaveGangCard && sp.HuType.HAVE_RETONG_HU {
					sp.GangHotStatus = true
					sp.OnWriteGameRecord(userItem.GetChairID(), "杠开预判摸牌，打开系统 热统 标识")
				}
				index := sp.UserTingCard(userItem, wChiHuRight, notbeMagic)
				if prob > 0 && rdm_gk <= prob {
					if index == static.INVALID_BYTE {
						sp.OnWriteGameRecord(userItem.GetChairID(), "必杠开操作失败！未在牌堆找到玩家能胡的牌或玩家还未能听牌。")
					} else {
						next := sp.LeftCardCount - 1
						sp.RepertoryCard[next], sp.RepertoryCard[index] = sp.RepertoryCard[index], sp.RepertoryCard[next]
						sp.OnWriteGameRecord(userItem.GetChairID(), "必杠开操作成功！")
					}
				} else {
					if index == static.INVALID_BYTE {
						//不能胡，不需要换牌
					} else {
						if sp.LeftCardCount > 1 {
							next := sp.LeftCardCount - 1
							if index != next {
								//下一张牌不能胡，不需要换牌
							} else {
								//循环换牌，直到找到不能胡的牌所在的位置
								for id := next - 1; id >= 0 && id < next; id-- {
									sp.RepertoryCard[next], sp.RepertoryCard[id] = sp.RepertoryCard[id], sp.RepertoryCard[next]
									index2 := sp.UserTingCard(userItem, wChiHuRight, notbeMagic)
									if index2 != next {
										sp.OnWriteGameRecord(userItem.GetChairID(), "不能杠开操作成功！")
										break
									}
									//换牌后还是能胡，把原来的牌换回去，避免改变牌堆
									sp.RepertoryCard[next], sp.RepertoryCard[id] = sp.RepertoryCard[id], sp.RepertoryCard[next]
								}
							}
						}
					}
				}
			}
			// 如果不能为癞子牌
			if notbeMagic {
				sp.OnWriteGameRecord(userItem.GetChairID(), "不允许摸4个癞子。")
				if sp.LeftCardCount > 1 {
					next := sp.LeftCardCount - 1
					if sp.RepertoryCard[next] == sp.MagicCard {
						// 找到下一张不为癞子的牌
						other := next - 1
						find := false
						for ; other >= 0; other-- {
							//此处过滤掉皮子和赖子（原来的写法当倒数第二张是皮子9同 那么赖子是1同。只过滤赖子的话会9同和1同换位置，然后导致变成3个赖子一个皮子）
							if sp.RepertoryCard[other] != sp.MagicCard && sp.RepertoryCard[other] != sp.PiZiCard {
								find = true
								break
							}
						}
						if find {
							sp.RepertoryCard[next], sp.RepertoryCard[other] = sp.RepertoryCard[other], sp.RepertoryCard[next]
							sp.OnWriteGameRecord(userItem.GetChairID(), "不允许四个癞子出现操作成功。")
						} else {
							sp.OnWriteGameRecord(userItem.GetChairID(), "不允许四个癞子出现操作失败，因为剩余牌全是癞子。")
						}
					} else {
						sp.OnWriteGameRecord(static.INVALID_CHAIR, "下一张牌不是癞子，不做处理。")
					}
				} else {
					sp.OnWriteGameRecord(static.INVALID_CHAIR, "剩余牌数小于一张，所以这里不再做摸癞子限制")
				}
			}
		} else {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "获取游戏控制配置为空")
		}
	}
	sp.SendCardCount++
	sp.SendCardData = sp.DrawOne()
	userItem.Ctx.DispatchCard(sp.SendCardData)
}

// 从牌库摸一张牌
func (sp *SportJZCXZ) DrawOne() byte {
	sp.LeftCardCount--
	cardData := sp.RepertoryCard[sp.LeftCardCount]
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
	return cardData
}
func (sp *SportJZCXZ) UserTotalMagic(userItem *components2.Player) byte {
	return userItem.Ctx.CurMagicOut + sp.m_GameLogic.GetMagicCount(userItem.Ctx.CardIndex[:])
}

// 得到玩家听的一张牌存在于牌堆的下标
func (sp *SportJZCXZ) UserTingCard(userItem *components2.Player, wChiHuRight uint16, notbeMagic bool) byte {
	var index byte = static.INVALID_BYTE
	var tingCard byte = static.INVALID_BYTE
	ignoreCards := make([]byte, 0)
	for func() bool {
		if tingCard == static.INVALID_BYTE {
			return true
		}
		for i := sp.LeftCardCount - 1; true; i-- {
			if i < 0 || i == static.INVALID_BYTE {
				break
			}

			if sp.RepertoryCard[i] == tingCard {
				index = byte(i)
				sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("已为其找到胡牌:%s。", sp.m_GameLogic.SwitchToCardNameByData(tingCard, 1)))
				return false
			}
		}
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("在牌堆未找到胡牌:%s。", sp.m_GameLogic.SwitchToCardNameByData(tingCard, 1)))
		ignoreCards = append(ignoreCards, tingCard)
		return true
	}() {
		tingCard = sp.AnalyseTingCard(userItem, wChiHuRight, notbeMagic, ignoreCards...)
		if tingCard == static.INVALID_BYTE {
			sp.OnWriteGameRecord(userItem.GetChairID(), "无法胡牌或没有听牌。")
			break
		}
	}
	return index
}

// 得到玩家的一张听牌
func (sp *SportJZCXZ) AnalyseTingCard(userItem *components2.Player, wChiHuRight uint16, notbeMagic bool, ignoreCards ...byte) byte {
	defer func() {
		userItem.Ctx.InitChiHuResult()
		userItem.Ctx.ClearOperateCard()
	}()
	in := func(card byte) bool {
		for _, icard := range ignoreCards {
			if icard == card {
				return true
			}
		}
		return false
	}
	cbCardIndex := userItem.Ctx.CardIndex[:]
	//听牌分析 获取必要的听牌检查数据
	//for i := byte(0); i < public.MAX_INDEX; i++ {
	//cbCurrentCard := sp.m_GameLogic.SwitchToCardData(i)
	for i := sp.LeftCardCount - 1; true; i-- {
		if i < 0 || i == static.INVALID_BYTE {
			break
		}

		cbCurrentCard := sp.RepertoryCard[i]
		if !sp.m_GameLogic.IsValidCard(cbCurrentCard) {
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("预判胡牌:%s为不合法的非法牌。", sp.m_GameLogic.SwitchToCardNameByData(cbCurrentCard, 1)))
			continue
		}
		if in(cbCurrentCard) {
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("预判胡牌:%s在之前已经判断过，但是牌堆没有，所以忽略掉。", sp.m_GameLogic.SwitchToCardNameByData(cbCurrentCard, 1)))
			continue
		}
		if notbeMagic && cbCurrentCard == sp.MagicCard {
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("预判胡牌:%s是癞子牌，前面判断玩家已经不能再摸癞子，所以忽略掉。", sp.m_GameLogic.SwitchToCardNameByData(cbCurrentCard, 1)))
			continue
		}
		cardIndex := sp.m_GameLogic.SwitchToCardIndex(cbCurrentCard)
		//校验一下，过滤掉不需要检查的孤章
		checkIndex := sp.m_GameLogic.OneTingCard(cbCardIndex, cardIndex)
		if checkIndex != static.INVALID_BYTE || cbCurrentCard == sp.MagicCard {
			userItem.Ctx.DispatchCard(cbCurrentCard)
			//结果判断
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("预判的胡牌:%s。", sp.m_GameLogic.SwitchToCardNameByData(cbCurrentCard, 1)))
			if sp.AnalyseHuPrototype(userItem, cbCurrentCard, wChiHuRight, &userItem.Ctx.ChiHuResult, false) {
				userItem.Ctx.DeleteCard(cbCurrentCard)
				return cbCurrentCard
			}
			userItem.Ctx.DeleteCard(cbCurrentCard)
		}

	}
	return static.INVALID_BYTE
}

// 判断胡牌
func (sp *SportJZCXZ) AnalyseHuPrototype(_userItem *components2.Player, cbCurrentCard byte, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult, otherPlayerOutCard bool) bool {
	// 在查胡之前进行分析
	err, npd := sp.EstimateConditionBeforAnalyseHu(_userItem, otherPlayerOutCard, wChiHuRight)
	if err != nil {
		sp.OnWriteGameRecord(_userItem.GetChairID(), err.Error())
		return false
	}
	// 别人打得癞子或者风牌 不能胡
	if otherPlayerOutCard && !sp.m_GameLogic.IsPengGangCard(cbCurrentCard) {
		//sp.OnWriteGameRecord(_userItem.GetChairID(), "4个红中可胡牌")
		return false
	}
	// 玩家胡牌权位
	huAction := byte(static.WIK_NULL)
	// 红中模式下，手上有4个红中，直接胡牌
	//if sp.IsCBMJ_HZ() {
	//	if sp.m_GameLogic.GetHongZhongCount(_userItem.Ctx.CardIndex[:]) >= 4 {
	//		sp.OnWriteGameRecord(_userItem.GetChairID(), "4个红中可胡牌")
	//		huAction |= public.WIK_CHI_HU
	//		_userItem.Ctx.ChiHuResult.ChiHuKind |= public.CHK_FOUR_LAIZE
	//	}
	//}
	// 如果有七对胡，判断七对
	//if sp.HuType.HAVE_QIDUI_HU {
	//	var cardTemp [public.MAX_INDEX]byte
	//	// 深拷贝一份玩家手牌
	//	public.HF_DeepCopy(&cardTemp, &_userItem.Ctx.CardIndex)
	//	// 如果是别人打得，把这张牌放入手牌变量中
	//	if otherPlayerOutCard && cbCurrentCard != public.INVALID_BYTE {
	//		cardTemp[sp.m_GameLogic.SwitchToCardIndex(cbCurrentCard)]++
	//	}
	//
	//	if sp.m_GameLogic.GetCardCount(cardTemp[:]) != public.MAX_COUNT {
	//		sp.OnWriteGameRecord(_userItem.GetChairID(), "七对牌数不够14张 胡不了")
	//	} else {
	//		// 判断是否七对胡
	//		if chiHuKind7p := sp.m_GameLogic.AnalyseIs7Pairs(cardTemp[:], public.CHK_NULL); chiHuKind7p != public.CHK_NULL {
	//			sp.OnWriteGameRecord(_userItem.GetChairID(), "七对可胡牌")
	//			huAction |= public.WIK_CHI_HU
	//			_userItem.Ctx.ChiHuResult.ChiHuKind |= chiHuKind7p
	//		} else {
	//			sp.OnWriteGameRecord(_userItem.GetChairID(), "七对胡不了")
	//		}
	//	}
	//}
	// 查表法算法原因，当是自己摸牌的时候，由于游戏流程中已经把这张牌放入手牌中，这里需要先拿出来
	// 在胡牌函数退出的时候还原他
	if !otherPlayerOutCard {
		if _userItem.Ctx.CardIndex[sp.m_GameLogic.SwitchToCardIndex(cbCurrentCard)] <= 0 {
			return false
		}
		_userItem.Ctx.DeleteCard(cbCurrentCard)
		defer _userItem.Ctx.DispatchCard(cbCurrentCard)
	}
	// 走到这里时，如果玩家还没有胡牌权位，则进行一次查表
	if huAction == static.WIK_NULL {
		//sp.OnWriteGameRecord(_userItem.GetChairID(),fmt.Sprintf("查胡之前：%s",sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:],1)))
		// 癞子容错
		//sp.FaultTolerantMagic()
		//huAction = sp.m_GameLogic.AnalyseChiHuCard(_userItem.Ctx.CardIndex[:], cbCurrentCard, wChiHuRight, ChiHuResult, otherPlayerOutCard)
		// sp.OnWriteGameRecord(_userItem.GetChairID(),fmt.Sprintf("查胡之后：%s",sp.m_GameLogic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:],1)))
		//20201016 苏大强 写死不是杠开吧看，这里应该就是判断一下能不能胡
		huAction = sp.CheckHu(_userItem.Seat, _userItem.Seat, cbCurrentCard, false)
	}
	// 胡牌函数执行完之后 进行一些校验
	return sp.EstimateConditionAfterAnalyseHu(_userItem, otherPlayerOutCard, huAction, ChiHuResult.ChiHuKind, npd)

}

// 在判胡之前判断玩家是否具备判断胡条件 主要是癞子个数得问题吧
// 赖晃：胡牌时允许有多个赖子，捉统时手中不能有赖子；如果有玩家杠出赖子，则所有玩家只能自摸胡。（赖子还原捉统也不能胡）
// 晃晃：玩家手上不管有几个赖子，只要满足胡牌牌型即可自摸或捉铳。
// 红中玩法：不管有没有红中（癞子）
// 一赖可捉统：手上只允许有一个赖子，即使赖子还原成了普通牌，也只能有一个赖子才能胡牌，且手上有的那个赖子必须还原才能捉统，否则只能自摸。
func (sp *SportJZCXZ) EstimateConditionBeforAnalyseHu(_userItem *components2.Player, byOther bool, wChiHuRight uint16) ( /*err*/ error /*needPassDealer*/, bool) {
	// TODO 实现各个晃晃的胡牌条件判断逻辑
	// 别人打的，规则是只能自摸，则胡不了
	//if byOther && sp.Rule.OnlyBySelf {
	//	// 潜江麻将里面 仅自摸可胡牌的 模式下 抢杠胡和热统还是可以胡的
	//	if !(sp.Rule.HuangHuangType == _HHTYPE_QJLH && (wChiHuRight&public.CHR_QIANG_GANG != 0 || wChiHuRight&public.CHR_RECHONG != 0)) {
	//		return errors.New("仅自摸可胡牌"), false
	//	}
	//}
	if byOther {
		// 潜江麻将里面 仅自摸可胡牌的 模式下 抢杠胡和热统还是可以胡的
		if !(wChiHuRight&static.CHR_QIANG_GANG != 0 || wChiHuRight&static.CHR_RECHONG != 0) {
			return errors.New("仅自摸可胡牌"), false
		}
	}
	//// 别人打得，规则是场上有人打出癞子后不能胡，只要有人打出癞子则胡不了
	//if byOther && sp.Rule.NoHuAfterMagic && sp.GetMagicOutCount() > 0 {
	//	return errors.New("场上有人打癞子，不能捉统"), false
	//}

	// 得到玩家手上的癞子个数
	userHandMagic := sp.m_GameLogic.GetMagicCount(_userItem.Ctx.CardIndex[:])
	// 硬晃模式下，不能有癞子
	//if sp.Rule.HuangHuangType == _HHTYPE_YH && userHandMagic > 0 {
	//	return errors.New("系统错误，硬晃不能有癞子"), false
	//}
	// 潜江麻将里面 抢杠胡 只能胡硬胡
	// if byOther && sp.Rule.HuangHuangType == _HHTYPE_QJLH && wChiHuRight&public.CHR_QIANG_GANG != 0 && userHandMagic > 0 {
	// 	return errors.New("潜江麻将：抢杠胡,胡牌者只能是硬胡的牌型"), false
	// }

	// if !sp.Rule.Hongzhong && !sp.Rule.YiLaiDaoDi && !sp.Rule.bHardHuang && !sp.Rule.bOneMagicHu && userHandMagic > 0 {
	// 	return errors.New("赖晃模式下，手上有癞子不能提统"), false
	// }
	// 一赖到底判断

	if byOther {
		if userHandMagic > 1 {
			return errors.New("手上有2个或以上癞子不能捉统"), false
		}
	} else {
		if userHandMagic > 1 {
			return errors.New("癞子数量大于一"), false
		}
	}

	// 潜江麻将
	//if sp.Rule.HuangHuangType == _HHTYPE_QJLH && userHandMagic > 1 {
	//	return errors.New("潜江麻将 胡牌最多只能有一个癞子"), false
	//}
	// 过庄 别人打的判断是否需要过庄
	//if byOther && sp.m_GameLogic.IsPassDealer(_userItem) {
	//	return nil, true
	//}
	// 校验通过
	return nil, false
}

func (sp *SportJZCXZ) CheckAnyCardHu(_userItem *components2.Player) bool {
	if !sp.Rule.Jz_JianZiHu && 0 == sp.m_GameLogic.Rule.Jz_Wanfa && byte(1) == sp.m_GameLogic.GetMagicCount(_userItem.Ctx.CardIndex[:]) { //不允许见字胡（所有的玩法如果勾选了不可见字胡，都不能见字胡）
		cbTempCard := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
		copy(cbTempCard, _userItem.Ctx.CardIndex[:])
		if sp.SendCardData != static.INVALID_BYTE {
			index := sp.m_GameLogic.SwitchToCardIndex(sp.SendCardData)
			if cbTempCard[index] > 0 {
				cbTempCard[index]--
			}
		}
		tingCount := sp.m_GameLogic.AnalyseTingCardCount(cbTempCard[:], _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, 0)
		tingMax := 27
		if sp.Rule.NoWan {
			tingMax = 18
		}
		if !(int(tingCount) < tingMax) { //胡牌个数小于tingMax个 那就不是见字胡了
			// 见字胡不能胡
			//发了牌以后判断这个人是否可以自摸胡牌
			return true
		}

	}
	return false
}

// 在得到牌权后判断能否胡
// 一赖可捉统：手上只允许有一个赖子，即使赖子还原成了普通牌，也只能有一个赖子才能胡牌，且手上有的那个赖子必须还原才能捉统，否则只能自摸。
func (sp *SportJZCXZ) EstimateConditionAfterAnalyseHu(_userItem *components2.Player, byOther bool, huAction byte, chiHuKind uint64, needPassDealer bool) bool {
	// 玩家没有胡的牌权
	if huAction == static.CHK_NULL {
		// sp.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("3n+2胡不了,牌桌癞子信息:%s,查表癞子信息:%s", sp.m_GameLogic.SwitchToCardNameByData(sp.MagicCard, 1), sp.m_GameLogic.SwitchToCardNameByDatas2(sp.m_GameLogic.HuCore.GetGuiInfo(), 1)))
		sp.OnWriteGameRecord(_userItem.GetChairID(), "判牌结果：胡不了...")
		return false
	}
	if !byOther {
		if sp.CheckAnyCardHu(_userItem) {
			sp.OnWriteGameRecord(_userItem.GetChairID(), "见字胡胡不了")
			return false
		}
	}
	// 不能见字胡 却是见字胡 胡不了

	//// 一赖可捉统模式下 癞子只能有一个 且必须没有充当万能牌 才能捉统
	//if sp.Rule.HuangHuangType == _HHTYPE_YLKZC && byOther && !(chiHuKind&public.CHK_PING_HU_NOMAGIC > 0) {
	//	sp.OnWriteGameRecord(_userItem.GetChairID(), "一赖可捉统模式下 癞子只能有一个 且必须没有充当万能牌 才能捉统")
	//	return false
	//}
	//if sp.Rule.HuangHuangType == _HHTYPE_QJLH {
	//	if sp.Rule.WanFa == _QJMJ_TSJ_YH {
	//		//能走到这里一定是一个赖子自摸的情况，其他情况 EstimateConditionBeforAnalyseHu 已经过滤拦截了
	//		if  chiHuKind&public.CHK_PING_HU_NOMAGIC != 0{
	//			//可以胡硬胡，那就可以胡了（注意不能判断有没有软胡，因为有情况是硬胡软胡可以一起胡的）
	//		}else{
	//			sp.OnWriteGameRecord(_userItem.GetChairID(), "潜江麻将，铁三角硬晃，一个赖子还原可以自摸胡")
	//			return false
	//		}
	//	}else{
	//		if (byOther || sp.Rule.WanFa == _QJMJ_THBG) && !(chiHuKind&public.CHK_PING_HU_NOMAGIC > 0) {
	//			sp.OnWriteGameRecord(_userItem.GetChairID(), "潜江麻将，捉统或土豪玩法，必须硬胡")
	//			return false
	//		}
	//	}
	//}
	// 如果需要过庄，发送过庄提示
	if needPassDealer && byOther {
		sp.SendGameNotificationMessage(_userItem.GetChairID(), consts.MsgContentGamePassDealer)
		return false
	}
	// 给牌权
	sp.OnWriteGameRecord(_userItem.GetChairID(), "3n+2可胡牌")
	_userItem.Ctx.UserAction |= huAction
	return true
}

func (sp *SportJZCXZ) FindWantCard(userItem *components2.Player) byte {
	index := static.INVALID_BYTE
	cardData := userItem.Ctx.WantCard
	if !sp.m_GameLogic.IsValidCard(cardData) {
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("不合法的WANT牌:%s。", sp.m_GameLogic.SwitchToCardNameByData(cardData, 1)))
		return index
	}
	return sp.FindCardIndexInRepertoryCards(cardData)
}
