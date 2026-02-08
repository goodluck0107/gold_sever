package Hubei_ShiShou_AH

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
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"github.com/open-source/game/chess.git/services/sport/sports/Hubei_ShiShou_AH/checkHu"
	"math"
	"math/rand"
	"time"
)

type OptHandler func(*static.Msg_C_OperateCard, *components2.Player, bool) error

type ErrorOptWaiting string

func (e ErrorOptWaiting) Error() string {
	return "opt waiting error: " + string(e)
}

type SportSSAH struct {
	// 游戏共用部分
	components2.Common
	// 游戏流程数据
	meta2.Metadata
	//// 游戏逻辑
	Logic SportLogicSSAH
	// 自定义
	cus *FriendRuleSSAH
	// opt map
	OptHandlerHash map[byte]OptHandler
	m_bAutoHu      [4]bool //托管超时自动胡
}

// 超时自动胡
type Msg_C_GameTimeOutAutoHu struct {
	Id     int64 `json:"id"` //玩家ID
	AutoHu bool  `json:"autohu"`
}

// 验胡结果
func (sp *SportSSAH) getHuRes(curUser *components2.Player, curCard int, isZiMo bool, isflag bool, isGH bool, isQGH bool) bool {
	curUser.Ctx.UserAction32 = 0       //清除卡三的数据
	sp.KzhDatas[int(curUser.Seat)] = 0 //清除卡三状态
	//==================算胡日志
	isHaveCHH := sp.Rule.IsCHH
	xlog.Logger().Debug("==========开始查询是否能超豪华======", isHaveCHH)
	curCardIndex, _ := mahlib2.CardToIndex(byte(curCard)) //  赖子牌值  转  索引数组\

	//TODO 要根据规则变动isHaveDH
	huTypeNoMagic := ssahCheckHu.CheckHu(curUser.Ctx, 0, int(curCardIndex), isZiMo, isHaveCHH, sp.OnWriteGameRecord)

	noMagicScore := 0
	if huTypeNoMagic != ssahCheckHu.GameHuNull {
		if huTypeNoMagic&ssahCheckHu.GameHuJYS != 0 && !isZiMo { // 接炮的时候 不是热冲或者不是抢杠胡 去掉将一色
			if sp.GangHotStatus || isQGH {
				//热冲或者抢杠是可以胡的，不用去掉将一色
			} else {
				huTypeNoMagic ^= ssahCheckHu.GameHuJYS
				if huTypeNoMagic == ssahCheckHu.GameHuNull { //有可能只有将一色 但是因为上面条件给 ^= 掉了
					curUser.Ctx.SetChiHuKind(curUser.Ctx.ChiHuResult.ChiHuKind, byte(curCard), 0) //由于小胡过大胡 也需要过庄，那么第三个参数就传0 不做比较
					sp.OnWriteGameRecord(curUser.GetChairID(), fmt.Sprintf("胡将一色点炮非热冲加入弃胡,牌:%s", sp.Logic.SwitchToCardNameByData(sp.ProvideCard, 1)))
				}
			}
		}

		if isQGH {
			huTypeNoMagic |= ssahCheckHu.GameHuQGH
		}
		noMagicScore = sp.getHuScore(huTypeNoMagic, isZiMo)
	}
	//sp.OnWriteGameRecord(curUser.Seat, fmt.Sprintf("【硬】胡结束 表：（%d）硬胡type（%d）胡分（%d）", ssahCheckHu.CheckHuLog(curUser.Ctx, 0,int(curCardIndex),IsDianPao,isHaveCHH) ,huTypeNoMagic,noMagicScore))
	sp.OnWriteGameRecord(curUser.Seat, fmt.Sprintf("【硬】胡结束 硬胡type（%d）胡分（%d）当前牌（%d）是否点炮（%t）", huTypeNoMagic, noMagicScore, curCardIndex, !isZiMo))

	if huTypeNoMagic != ssahCheckHu.GameHuNull {
		if huTypeNoMagic == ssahCheckHu.GameHuPiH {
			if sp.Rule.IsChiHu && !isZiMo { //勾选不能吃胡  并且不是自摸 不能胡
				return false
			}
			if !isZiMo && !sp.GangHotStatus { // 热冲不走卡字胡
				//这里要查询是否满足卡字胡以及卡三张规则
				if sp.getKaZiHuRes(curUser, int(curCardIndex), isZiMo, isflag, isGH) {
					curUser.Ctx.UserAction |= static.WIK_CHI_HU
					curUser.Ctx.ChiHuResult.ChiHuKind2 |= huTypeNoMagic
					return true
				} else {
					sp.KzhDatas[int(curUser.Seat)] = 1
					curUser.Ctx.SetChiHuKind(curUser.Ctx.ChiHuResult.ChiHuKind, byte(curCard), 0) //由于小胡过大胡 也需要过庄，那么第三个参数就传0 不做比较
					sp.OnWriteGameRecord(curUser.GetChairID(), fmt.Sprintf("胡将一色点炮非热冲加入弃胡,牌:%s", sp.Logic.SwitchToCardNameByData(sp.ProvideCard, 1)))
					return false
				}
			} else {
				curUser.Ctx.UserAction |= static.WIK_CHI_HU
				curUser.Ctx.ChiHuResult.ChiHuKind2 |= huTypeNoMagic
				return true
			}
		} else {
			curUser.Ctx.UserAction |= static.WIK_CHI_HU
			curUser.Ctx.ChiHuResult.ChiHuKind2 |= huTypeNoMagic
			return true
		}
	} else {
		return false
	}
}

// 卡字胡情况
func (sp *SportSSAH) getKaZiHuRes(curUser *components2.Player, curCardIndex int, isZiMo bool, isflag bool, isGH bool) bool {
	// 小胡里面要看 普通小胡 还是卡子胡、
	tempTingArr := ssahCheckHu.CheckTing(curUser.Ctx, 0, int(curCardIndex), isZiMo)
	if len(tempTingArr) == 1 { //只听1张 不满足卡字胡牌型 所有听牌值的剩余和>4 默认普通小胡
		if isflag {
			return false
		}
		return true
	}
	//这里找出所有看得见的牌 即：打出去的牌 所有玩家 吃碰杠的牌 和自己的手牌
	seeCards := make([]byte, 0)
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		seeCards = append(seeCards, _item.Ctx.DiscardCard...) //这里加上打出去的牌
		for j := 0; j < len(_item.Ctx.WeaveItemArray); j++ {
			if _item.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_PENG {
				tempPengCards := _item.Ctx.WeaveItemArray[j].CenterCard
				seeCards = append(seeCards, tempPengCards, tempPengCards, tempPengCards)
			} else if _item.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_GANG ||
				_item.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_FILL {
				tempGangCards := _item.Ctx.WeaveItemArray[j].CenterCard
				seeCards = append(seeCards, tempGangCards, tempGangCards, tempGangCards, tempGangCards)
			}
		}
	}
	isDuiDao := false
	// 如果听的牌是两张，且这两张牌在手里是对子，那么就认为他胡对到
	if len(tempTingArr) == 2 {
		isDuiDao = true
		for _, tingIndex := range tempTingArr {
			if curUser.Ctx.CardIndex[tingIndex] < 2 {
				isDuiDao = false
				break
			}
		}
	}
	tmpIndexCard := [static.MAX_INDEX]byte{}
	// static.HF_DeepCopy(&tmpIndexCard, &curUser.Ctx.CardIndex)
	if isDuiDao {
		static.HF_DeepCopy(&tmpIndexCard, &curUser.Ctx.CardIndex)
		//for _, tingIndex := range tempTingArr {
		//	tmpIndexCard[tingIndex] = curUser.Ctx.CardIndex[tingIndex]
		//}
	}
	// 转换扑克
	for i := 0; i < len(seeCards); i++ {
		n, _ := mahlib2.CardToIndex(seeCards[i])
		tmpIndexCard[n]++
	}
	tingCount := 0 //听胡的牌 出现的数量
	for i := 0; i < len(tempTingArr); i++ {
		tingCount += int(tmpIndexCard[byte(tempTingArr[i])])
	}
	noSeeTHCard := len(tempTingArr)*4 - tingCount
	xlog.Logger().Debug("听胡数据,看不见数量,看得见的牌的数据", tempTingArr, noSeeTHCard, tmpIndexCard)
	if noSeeTHCard > 4 { //进入卡字胡牌型限制
		if isGH { //g跟胡直接胡
			return true
		}
		//判断满足卡三张
		seeIndexCard := [static.MAX_INDEX]byte{}
		for i := 0; i < len(seeCards); i++ {
			n, _ := mahlib2.CardToIndex(seeCards[i])
			seeIndexCard[n]++
		}
		kszCardCount := 0
		for i := 0; i < len(tempTingArr); i++ {
			kszCardCount += int(seeIndexCard[tempTingArr[i]])
		}
		//if seeIndexCard[curCardIndex] >= 2 {
		if kszCardCount > 2 {
			if isflag {
				curUser.Ctx.UserAction32 |= static.WIK_K3Z
			}
			return true
		}
		return false
	} else {
		if isflag {
			return false
		}
		return true
	}
}

// 胡分
func (sp *SportSSAH) GetScoreOnHu(msgGameEnd *static.Msg_S_GameEnd) {
	var (
		tempOffset      int
		nextRoundBanker uint16
		needRGS         bool // 是否需要还原杠分
		isQHRGS         bool // 是否需要还原抢杠分
	)
	msgGameEnd.ChiHuUserCount = 0
	offset := [meta2.MAX_PLAYER]int{}
	isHaveDP := false
	//单独写一个提在算分的上面 计算码数
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.Ctx.PerformAction&static.WIK_CHI_HU != 0 && sp.ProvideUser != static.INVALID_CHAIR {
			// end消息结构体改变
			msgGameEnd.WWinner[i] = true
			if !isHaveDP {
				isHaveDP = !(int(sp.ProvideUser) == i)
			}
		}
	}
	isYPDX := 0
	for i := 0; i < len(msgGameEnd.WWinner); i++ {
		if msgGameEnd.WWinner[i] {
			isYPDX++
		}
	}

	msgGameEnd.BingoBirdCount, msgGameEnd.CbBirdData_ex2 = sp.getMaData(sp.HaveHuangZhuang, isHaveDP && (isYPDX > 1), msgGameEnd.WWinner)
	xlog.Logger().Debug("BingoBirdCount", msgGameEnd.BingoBirdCount, msgGameEnd.CbBirdData_ex2)
	if !sp.HaveHuangZhuang {
		for i := 0; i < sp.GetPlayerCount(); i++ { //这里应该是去最大玩家数量，暂时写4 没找到常量
			_item := sp.GetUserItemByChair(uint16(i))
			if _item == nil {
				continue
			}
			isDH := false
			if _item.Ctx.PerformAction&static.WIK_CHI_HU != 0 && sp.ProvideUser != static.INVALID_CHAIR {
				nextRoundBanker = _item.GetChairID()
				msgGameEnd.ChiHuUserCount++
				_item.Ctx.ChiHuUserCount++ //胡牌总数

				// 胡牌种类（不带自摸）
				msgGameEnd.WChiHuKind[i] = _item.Ctx.ChiHuResult.ChiHuKind2
				isZiMo := int(sp.ProvideUser) == i
				pxNum := sp.getHuScore(_item.Ctx.ChiHuResult.ChiHuKind2, isZiMo) //牌型分数
				msgGameEnd.HuFen[i] = pxNum

				if _item.Ctx.ChiHuResult.ChiHuKind2&ssahCheckHu.GameHuQGH != 0 || sp.GangHotStatus {
					needRGS = true
				}
				if _item.Ctx.ChiHuResult.ChiHuKind2&ssahCheckHu.GameHuQH != 0 {
					isQHRGS = true
				}

				if sp.GangFlower {
					msgGameEnd.BigHuKind = static.GameBigHuKindGSK
					msgGameEnd.WBigHuKind[i] = static.GameBigHuKindGSK
					//_item.Ctx.BigHuUserCount++
				} else if (_item.Ctx.ChiHuResult.ChiHuKind2&ssahCheckHu.GameHu7D) != 0 || (_item.Ctx.ChiHuResult.ChiHuKind2&ssahCheckHu.GameHu7D1) != 0 || (_item.Ctx.ChiHuResult.ChiHuKind2&ssahCheckHu.GameHu7D2) != 0 || (_item.Ctx.ChiHuResult.ChiHuKind2&ssahCheckHu.GameHu7D3) != 0 {
					msgGameEnd.BigHuKind = static.GameBigHuKind_7
					msgGameEnd.WBigHuKind[i] = static.GameBigHuKind_7
					_item.Ctx.BigHuUserCount++
					isDH = true
				} else if (_item.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHuQYS) != 0 {
					msgGameEnd.BigHuKind = static.GameBigHuKindQYS
					msgGameEnd.WBigHuKind[i] = static.GameBigHuKindQYS
					_item.Ctx.BigHuUserCount++
					isDH = true
				} else if (_item.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHuJYS) != 0 {
					msgGameEnd.BigHuKind = static.GameBigHuKindJYS
					msgGameEnd.WBigHuKind[i] = static.GameBigHuKindJYS
					_item.Ctx.BigHuUserCount++
					isDH = true
				}
				if isZiMo {
					// _item.Ctx.HuBySelf()
					for j := 0; j < sp.GetPlayerCount(); j++ {
						_itemJ := sp.GetUserItemByChair(uint16(j)) //其他玩家
						if _itemJ == nil {
							continue
						}
						if int(sp.ProvideUser) != j {
							//n := _item.Ctx.MagicCardOut + _itemJ.Ctx.MagicCardOut
							//胡牌扣分={胡牌牌型分*（1+胡牌人中码数）*（1+输家中码数）}*底分
							tempOffset = (pxNum * (1 + int(msgGameEnd.BingoBirdCount[nextRoundBanker])) * (1 + int(msgGameEnd.BingoBirdCount[j]))) * sp.Rule.DiFen
							xlog.Logger().Print("自摸的胡分数：", tempOffset)
							msgGameEnd.GameScore[j] -= tempOffset
							msgGameEnd.GameScore[i] += tempOffset
						}
					}
					//break
				} else {
					// 这里是点炮 需要知道 玩家胡的是不是大胡 还有就是 面板选择是否 大胡点炮奖码
					proUser := sp.GetUserItemByChair(sp.ProvideUser)
					if proUser != nil {
						proUser.Ctx.ProvideUserCount++ //点炮总数
						//胡牌扣分={胡牌牌型分*（1+胡牌人中码数）*（1+输家中码数）}*底分i
						if i != int(sp.ProvideUser) {
							tempOffset = (pxNum * (1 + int(msgGameEnd.BingoBirdCount[i])) * (1 + int(msgGameEnd.BingoBirdCount[sp.ProvideUser]))) * sp.Rule.DiFen
							if isDH && sp.Rule.IsDHJM == 1 {
								//tempOffset = (pxNum * (1 + 0) * ( 1 + 0)) * sp.Rule.DiFen
							}
						}
						proChair := int(proUser.GetChairID())
						msgGameEnd.GameScore[proChair] -= tempOffset
						msgGameEnd.GameScore[i] += tempOffset
					}
				}
			}
		}
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			if _item == nil {
				continue
			}
			if _item.Ctx.MaxScoreUserCount < 1 { // 判断玩家的最大分数 如果没有的话 初始化为0  因为不初始化数据为 -9999
				_item.Ctx.MaxScoreUserCount = 0
			}
			if _item.Ctx.MaxScoreUserCount < msgGameEnd.GameScore[i] {
				_item.Ctx.MaxScoreUserCount = msgGameEnd.GameScore[i]
			}
		}
	}

	for i := 0; i < sp.GetPlayerCount(); i++ {
		userItem := sp.GetUserItemByChair(uint16(i))

		if userItem != nil {
			for j := 0; j < len(userItem.Ctx.WeaveItemArray); j++ {

				if needRGS && (userItem.Seat == sp.ProvideUser) {
					if j == int(sp.LastGangIndex) {
						continue
					}
				}

				if isQHRGS && (j == int(sp.LastGangIndex)) && (userItem.Ctx.WeaveItemArray[j].ProvideUser == sp.ProvideUser) {
					continue
				}
				//杠分扣分={杠牌基础分*（1+杠的人中码数）*（1+扣分人中码数）}*底分   点杠3分  暗杠2分   蓄杠1分
				if userItem.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_FILL {
					for y := 0; y < sp.GetPlayerCount(); y++ {
						if i != y {
							msgGameEnd.GameScore[y] -= (1 * (1 + int(msgGameEnd.BingoBirdCount[i])) * (1 + int(msgGameEnd.BingoBirdCount[y]))) * sp.Rule.DiFen
							msgGameEnd.GameScore[i] += (1 * (1 + int(msgGameEnd.BingoBirdCount[i])) * (1 + int(msgGameEnd.BingoBirdCount[y]))) * sp.Rule.DiFen
						}
					}
				} else if userItem.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_GANG {
					if int(userItem.Ctx.WeaveItemArray[j].ProvideUser) == i {
						for y := 0; y < sp.GetPlayerCount(); y++ {
							if i != y {
								msgGameEnd.GameScore[y] -= (2 * (1 + int(msgGameEnd.BingoBirdCount[i])) * (1 + int(msgGameEnd.BingoBirdCount[y]))) * sp.Rule.DiFen
								msgGameEnd.GameScore[i] += (2 * (1 + int(msgGameEnd.BingoBirdCount[i])) * (1 + int(msgGameEnd.BingoBirdCount[y]))) * sp.Rule.DiFen
							}
						}
					} else {
						msgGameEnd.GameScore[userItem.Ctx.WeaveItemArray[j].ProvideUser] -= (3 * (1 + int(msgGameEnd.BingoBirdCount[i])) * (1 + int(msgGameEnd.BingoBirdCount[userItem.Ctx.WeaveItemArray[j].ProvideUser]))) * sp.Rule.DiFen
						msgGameEnd.GameScore[int(userItem.Seat)] += (3 * (1 + int(msgGameEnd.BingoBirdCount[int(userItem.Seat)])) * (1 + int(msgGameEnd.BingoBirdCount[userItem.Ctx.WeaveItemArray[j].ProvideUser]))) * sp.Rule.DiFen
					}
				}
			}
		}
		//offset[i] += msgGameEnd.GameScore[i]
	}

	//if needRGS {
	//	for i := 0; i < sp.GetPlayerCount(); i++ {
	//		userItem := sp.GetUserItemByChair(sp.ProvideUser)
	//		index:=int(sp.ProvideUser)
	//		if userItem != nil  {
	//			if i != index {
	//				msgGameEnd.GameScore[i] += (int(sp.LastGangKind) * (1 + int(msgGameEnd.BingoBirdCount[index])) * (1 + int(msgGameEnd.BingoBirdCount[i]))) * sp.Rule.DiFen
	//				msgGameEnd.GameScore[index] -= (int(sp.LastGangKind) * (1 + int(msgGameEnd.BingoBirdCount[index])) * (1 + int(msgGameEnd.BingoBirdCount[i]))) * sp.Rule.DiFen
	//			}
	//		}
	//	}
	//}

	for i := 0; i < sp.GetPlayerCount(); i++ {
		offset[i] += msgGameEnd.GameScore[i]
	}

	msgGameEnd.UserScore, msgGameEnd.UserVitamin = sp.OnSettle(offset, consts.EventSettleGameOver)

	//for i := 0; i < sp.GetPlayerCount(); i++ {
	//	_item := sp.GetUserItemByChair(uint16(i))
	//	if _item == nil {
	//		continue
	//	}
	//	msgGameEnd.GameAdjustScore[i] += _item.Ctx.StorageScore
	//}

	if msgGameEnd.ChiHuUserCount > 1 {
		nextRoundBanker = sp.ProvideUser
	}
	sp.BankerUser = nextRoundBanker

}

// 杠分
func (sp *SportSSAH) GetScoreOnGang(curUser, proUser *components2.Player, gangType uint16) [meta2.MAX_PLAYER]int {
	var (
		baseScore     int                         // 基础分
		tempOffsetArr [meta2.MAX_PLAYER]int       // 临时偏移量数组
		curChair      = int(curUser.GetChairID()) // 当前玩家座位号
		tempOffset    int                         // 临时偏移量
	)
	switch gangType {
	case info2.E_Gang_XuGand:
		baseScore = 1
	case info2.E_Gang_AnGang:
		baseScore = 2
	case info2.E_Gang:
		baseScore = 3
	default:
		xlog.Logger().Debug("杠牌类型找不到")
	}
	tempOffset = baseScore
	//暗杠4分 其他2分
	if gangType == info2.E_Gang {
		if proUser != nil {
			proChair := int(proUser.GetChairID())
			tempOffsetArr[proChair] -= tempOffset
			tempOffsetArr[curChair] += tempOffset
		}
	} else {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			if _item == nil {
				continue
			}
			if curChair != i {
				tempOffsetArr[i] -= tempOffset
				tempOffsetArr[curChair] += tempOffset
			}
		}

	}
	return tempOffsetArr
}

// 胡分数
func (sp *SportSSAH) getHuScore(HuType uint64, isZiMo bool) int {
	//① 平胡：1倍
	//② 自摸：2倍
	//⑥ 2 5 8将一色：5分
	//⑤ 七对：5分
	//④ 清 一 色:5分
	//③ 清 七 对:10分
	//③ 豪 华 七 对:10分
	//③ 超 豪 华 七 对:20分
	//③ 三 豪 华 七 对:10分

	HuScore := 0
	if (HuType & ssahCheckHu.GameHuPiH) != 0 { // 屁胡
		HuScore += 1
		if isZiMo {
			HuScore *= 2
		}
		if (HuType & ssahCheckHu.GameHuQGH) != 0 {
			HuScore = 5
		} else if sp.GangHotStatus {
			if sp.LastGangKind == 2 {
				HuScore = 10
			} else {
				HuScore = 5
			}
		}
	} else {
		score7D := 0
		if (HuType & ssahCheckHu.GameHu7D) != 0 {
			score7D = 5
		}

		scoreQYS := 0
		if (HuType & ssahCheckHu.GameHuQYS) != 0 {
			scoreQYS = 5
		}

		scoreJYS := 0
		if (HuType & ssahCheckHu.GameHuJYS) != 0 {
			scoreJYS = 5
		}

		score7D1 := 0
		if (HuType & ssahCheckHu.GameHu7D1) != 0 {
			score7D1 = 10
		}
		score7D2 := 0
		if (HuType & ssahCheckHu.GameHu7D2) != 0 {
			score7D2 = 20
		}
		score7D3 := 0
		if (HuType & ssahCheckHu.GameHu7D3) != 0 {
			score7D3 = 40
		}

		HuScore = scoreQYS + score7D + scoreJYS + score7D1 + score7D2 + score7D3
		if (HuType & ssahCheckHu.GameHuQGH) != 0 {
			HuScore += 5
		} else if sp.GangHotStatus {
			if sp.LastGangKind == 2 {
				HuScore += 10
			} else {
				HuScore += 5
			}
		}
	}

	if sp.Rule.IsCHH {
		if HuScore > 20 {
			HuScore = 20
		}
	} else {
		if HuScore > 10 {
			HuScore = 10
		}
	}

	sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("牌型分数：%d", HuScore))
	// xlog.Logger().Print("牌型：", HuType, HuScore, isZiMo)
	return HuScore
}

func (sp *SportSSAH) getMaCount(WinIndex uint16) int {
	isHaveDh := true
	maCount := 0
	if WinIndex == static.INVALID_CHAIR {
		if sp.LastOutCardUser != static.INVALID_CHAIR {
			WinIndex = uint16(sp.GetNextSeat(sp.LastOutCardUser))
		} else {
			WinIndex = 0
			sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("没有赢家 并且没有最后一个出牌人的数据：%d", sp.LastOutCardUser))
		}
	}

	winer := sp.GetUserItemByChair(WinIndex)
	if winer == nil {
		sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("没有赢家"))
	}

	isK5 := false
	if sp.ProvideCard&static.MASK_VALUE == 0x05 {
		index := sp.Logic.SwitchToCardIndex(sp.ProvideCard)
		if winer.Ctx.CardIndex[index-1] > 0 && winer.Ctx.CardIndex[index+1] > 0 {
			cards := static.HF_BytesToInts(winer.Ctx.CardIndex[:])
			tmpCards := make([]int, len(cards))
			copy(tmpCards, cards)
			tmpCards[index-1] -= 1
			tmpCards[index] -= 1
			tmpCards[index+1] -= 1
			if ssahCheckHu.CheckHuTest(tmpCards, 0) {
				isK5 = true
			}
		}
	}
	isMQQ := false //sp.HidGang
	Pcount := 0
	for i := 0; i < len(winer.Ctx.WeaveItemArray); i++ {
		if winer.Ctx.WeaveItemArray[i].WeaveKind == static.WIK_PENG {
			Pcount++
		}
	}
	if Pcount == 0 {
		if winer.Ctx.WeaveItemCount <= winer.Ctx.HidGang {
			isMQQ = true
		}
	}

	isZiMo := WinIndex == sp.ProvideUser
	isDhCount := 0

	if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHu7D) != 0 {
		isDhCount = 1
	} else if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHu7D1) != 0 {
		isDhCount = 2
	} else if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHu7D2) != 0 {
		isDhCount = 3
	} else if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHu7D3) != 0 {
		isDhCount = 3 //三大胡以上 还是3
	} else if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHuJYS) != 0 {
		isDhCount = 1 //三大胡以上 还是3
	}

	if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHuQYS) != 0 {
		isDhCount++
		if isDhCount > 3 {
			isDhCount = 3
		}
	}

	if (winer.Ctx.ChiHuResult.ChiHuKind2 & ssahCheckHu.GameHuPiH) != 0 {
		isHaveDh = false
	}
	maCountArr := []int{sp.Rule.DiMa, sp.Rule.DiMa + 2, sp.Rule.DiMa + 3, sp.Rule.DiMa + 4}
	if isZiMo {
		if isHaveDh {
			maCount = maCountArr[isDhCount]
		} else {
			maCount = sp.Rule.DiMa + 1
		}
		if isK5 {
			maCount += 1
		}
		if isMQQ {
			maCount += 1
		}
		if sp.GangFlower {
			maCount += 1
		}
	} else {
		maCount = sp.Rule.DiMa
		if isHaveDh {
			if sp.Rule.IsDHJM == 1 {
				//maCount = 0
			} else {
				if isK5 {
					maCount += 1
				}
				if isMQQ {
					maCount += 1
				}
			}
		}
	}

	return maCount
}

// 码
func (sp *SportSSAH) getMaData(isLiuJu bool, isJCma bool, wWiner [4]bool) (b [meta2.MAX_PLAYER]byte, c [meta2.MAX_PLAYER][]int) {
	maCount := 0
	winerIndex := static.INVALID_CHAIR
	if isLiuJu {
		isHaveG := false
		winerIndex = 0
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			if _item == nil {
				continue
			}
			for j := 0; j < len(_item.Ctx.WeaveItemArray); j++ {
				if _item.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_FILL || _item.Ctx.WeaveItemArray[j].WeaveKind == static.WIK_GANG {
					isHaveG = true
					break
				}
			}
		}
		if isHaveG {
			maCount = sp.Rule.DiMa
		}
	} else {
		tempMaxMaNum := []int{0, 0, 0, 0}
		if isJCma {
			for i := 0; i < sp.GetPlayerCount(); i++ {
				_item := sp.GetUserItemByChair(uint16(i))
				if _item == nil {
					continue
				}
				if wWiner[i] {
					tempMaxMaNum[i] = sp.getMaCount(_item.Seat)
				}
			}
			maxVal := tempMaxMaNum[0]
			maxIndex := 0
			for i := 1; i < len(tempMaxMaNum); i++ {
				//从第二个 元素开始循环比较，如果发现有更大的，则交换
				if maxVal < tempMaxMaNum[i] {
					maxVal = tempMaxMaNum[i]
					maxIndex = i
				}
			}
			maCount = sp.getMaCount(uint16(maxIndex))
			winerIndex = uint16(maxIndex)
			xlog.Logger().Debugln("码数的最大值：", maxVal)

		} else {
			firstWinner := static.INVALID_CHAIR
			for _, order := range sp.ReplayRecord.VecOrder {
				if order.Operation == info2.E_Hu {
					firstWinner = order.Chair_id
					break
				}
			}
			winerIndex = firstWinner
			maCount = sp.getMaCount(firstWinner)
		}
	}

	isSuiJi := false
	isSys := false
	// 马牌数据
	houseList := make([]byte, 0)
	if sp.Rule.MaType == 0 {
		if maCount > sp.Rule.DiMa+4 {
			maCount = sp.Rule.DiMa + 4
		}
		for i := 0; i < maCount; i++ {
			houseList = append(houseList, sp.DrawOne())
		}
	} else if sp.Rule.MaType == 1 { //自动
		if int(sp.LeftCardCount) >= maCount {
			for i := 0; i < maCount; i++ {
				houseList = append(houseList, sp.DrawOne())
			}
		} else {
			isSuiJi = true
			for i := 0; i < maCount; i++ {
				houseList = append(houseList, sp.RepertoryCard[static.HF_GetRandom(len(sp.RepertoryCard))])
			}
		}
	} else if sp.Rule.MaType == 2 { //系统
		xlog.Logger().Debug("配码", sp.LeftCardCount, maCount)
		if int(sp.LeftCardCount) >= maCount {
			for i := 0; i < maCount; i++ {
				houseList = append(houseList, sp.DrawOne())
			}
		} else {
			isSys = true
			for i := 0; i < maCount; i++ {
				houseList = append(houseList, sp.RepertoryCard[static.HF_GetRandom(len(sp.RepertoryCard))])
			}
		}
	}

	horseTbl := [meta2.MAX_PLAYER][meta2.MAX_HORSE]byte{}
	switch sp.GetPlayerCount() {
	case 2:
		horseTbl = [meta2.MAX_PLAYER][meta2.MAX_HORSE]byte{
			[meta2.MAX_HORSE]byte{1, 3, 5, 7, 9, 0, 0, 0},
			[meta2.MAX_HORSE]byte{2, 4, 6, 8, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{0, 0, 0, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 3:
		horseTbl = [meta2.MAX_PLAYER][meta2.MAX_HORSE]byte{
			[meta2.MAX_HORSE]byte{1, 4, 7, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{2, 5, 8, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{3, 6, 9, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{0, 0, 0, 0, 0, 0, 0, 0},
		}
	case 4:
		horseTbl = [meta2.MAX_PLAYER][meta2.MAX_HORSE]byte{
			[meta2.MAX_HORSE]byte{1, 5, 9, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{2, 6, 0, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{3, 7, 0, 0, 0, 0, 0, 0},
			[meta2.MAX_HORSE]byte{4, 8, 0, 0, 0, 0, 0, 0},
		}
	}

	// 每个玩家买的什么马
	payHorse := [meta2.MAX_PLAYER][meta2.MAX_HORSE]byte{}
	// TODO  赢家数据需要改成数组
	winerIndexTemp := winerIndex
	for i := 0; i < sp.GetPlayerCount(); i++ {
		if i > 0 {
			winerIndexTemp = uint16(sp.GetNextSeat(winerIndexTemp))
		}
		payHorse[winerIndexTemp] = horseTbl[i]
	}
	// 每个玩家中马数量
	bingohousedetail := [meta2.MAX_PLAYER]byte{}
	// 每个玩家中马的马牌
	cbHouseData := [meta2.MAX_PLAYER][]int{} //最多中马 12张
	//系统配码 平均分配
	if isSys {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			cbHouseData[i] = make([]int, 12)
		}
		playIndex := 0
		if winerIndex != 0 {
			playIndex = int(winerIndex)
		}
		for count := maCount; count > 0; count-- {
			if playIndex >= sp.GetPlayerCount() {
				playIndex = 0
			}
			cbHouseData[playIndex] = append(cbHouseData[playIndex], 255)
			bingohousedetail[playIndex]++
			playIndex++
		}
		for i := 0; i < sp.GetPlayerCount(); i++ {
			sp.OnWriteGameRecord(uint16(i), fmt.Sprintf("中马个数：%d, 中马详情：%+v", bingohousedetail[i], cbHouseData[i]))
		}
	} else {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			cbHouseData[i] = make([]int, 12)
			// 轮询得到中马
			for a := 0; a < len(houseList); a++ {
				for b := 0; b < 12; b++ {
					if payHorse[i][b] <= 0 {
						break
					}
					if payHorse[i][b] == (houseList[a] & static.MASK_VALUE) {
						bingohousedetail[i]++
						// 从小到大 找到第一个不为0的 插入
						for c := 0; c < 12; c++ {
							if cbHouseData[i][c] <= 0 {
								if isSuiJi {
									cbHouseData[i][c] = 255
								} else {
									cbHouseData[i][c] = int(houseList[a])
								}
								break
							}
						}
						break
					}
				}
			}
			sp.OnWriteGameRecord(uint16(i), fmt.Sprintf("中马个数：%d, 中马详情：%+v", bingohousedetail[i], cbHouseData[i]))
		}
	}

	return bingohousedetail, cbHouseData
}

func (sp *SportSSAH) OnInit(table base2.TableBase) {
	sp.KIND_ID = table.GetTableInfo().KindId
	sp.Logic.CreateCheckHu()
	sp.Config.StartMode = static.StartMode_FullReady
	sp.Config.PlayerCount = 4 //玩家人数
	sp.Config.ChairCount = 4  //椅子数量
	sp.PlayerInfo = make(map[int64]*components2.Player)

	sp.ResetTable()
	sp.SetGameStartMode(static.StartMode_FullReady)
	sp.GameTable = table
	sp.Init()
	sp.Unmarsha(table.GetTableInfo().GameInfo)
	var _msg FriendRuleSSAH
	if err := json.Unmarshal(static.HF_Atobytes(table.GetTableInfo().Config.GameConfig), &_msg); err == nil {

		if _msg.LookonSupport == "" {
			sp.Config.LookonSupport = true
		} else {
			sp.Config.LookonSupport = _msg.LookonSupport == "true"
		}
	}
	table.GetTableInfo().GameInfo = ""

	sp.SetVitaminLowPauseTime(10)
	//sp.FriendInfo.DismissType = 1
	sp.InitOptHandler()
	////解析规则
	sp.ParseRule()
}

func (sp *SportSSAH) OnEnd() {
	if sp.IsGameStarted() {
		sp.OnGameOver(static.INVALID_CHAIR, static.GER_DISMISS)
	}
}

func (sp *SportSSAH) OnGameOver(chair uint16, cbReason byte) bool {
	sp.OnEventGameEnd(chair, cbReason)
	return true
}

func (sp *SportSSAH) SendGameScene(uid int64, status byte, secret bool) {
	player := sp.GetUserItemByUid(uid)
	if player == nil {
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

// ! 发送游戏开始场景数据
func (sp *SportSSAH) sendGameSceneStatusPlay(player *components2.Player) bool {
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
	//player.Ctx.SetTrustee(false)
	var Trustee static.Msg_S_Trustee
	Trustee.Trustee = false
	Trustee.ChairID = wChiarID
	sp.SendTableMsg(consts.MsgTypeGameTrustee, Trustee)
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
	StatusPlay.PayPaostatus = sp.PayPaoStatus
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
		if _item.Ctx.IsJinHu {
			StatusPlay.IsJinHu[i] = true
		}
		//玩家番数
		StatusPlay.PlayerFan[i] = sp.GetPlayerFan(uint16(i))
		//追加痞子out
		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang + _item.Ctx.PiZiGangCount + _item.Ctx.PiZiCardOut
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.UserPaoReady == true
		StatusPlay.Whotrust[i] = _item.CheckTRUST()
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		for j := 0; j < len(_item.Ctx.Pilaicardcard) && j < len(StatusPlay.Pilaicardcard[i]); j++ {
			StatusPlay.Pilaicardcard[i][j] = _item.Ctx.Pilaicardcard[j]
		}

		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = player.Ctx.UserAction

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	//StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	//StatusPlay.CardLeft.Random = sp.RepertoryCardArray.Random

	if player.Ctx.Response {
		StatusPlay.ActionMask = static.WIK_NULL
	}
	StatusPlay.VecGangCard = static.HF_BytesToInts(player.Ctx.VecGangCard)
	if (StatusPlay.ActionMask & static.WIK_QIANG) != 0 {
		StatusPlay.ActionCard = byte(wChiarID)
	}

	if player.Ctx.CheckTimeOut != 0 {
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
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData

	//扑克数据
	StatusPlay.CardCount, StatusPlay.CardData = sp.Logic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
	StatusPlay.IsOutCard = true
	if player.Ctx.UserAction32&static.WIK_PENG != 0 {
		StatusPlay.IsOutCard = false
	} else if player.Ctx.UserAction32&static.WIK_GANG != 0 {
		StatusPlay.IsOutCard = false
	}

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

func (sp *SportSSAH) SaveGameData() {
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
	StatusPlay.PayPaostatus = sp.PayPaoStatus

	for i := 0; i < sp.GetPlayerCount(); i++ {

		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		if _item.Ctx.IsJinHu {
			StatusPlay.IsJinHu[i] = true
		}
		//玩家番数
		StatusPlay.PlayerFan[i] = sp.GetPlayerFan(uint16(i))
		//追加痞子out
		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang + _item.Ctx.PiZiGangCount + _item.Ctx.PiZiCardOut
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.UserPaoReady == true
		StatusPlay.Whotrust[i] = _item.CheckTRUST()
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		for j := 0; j < len(_item.Ctx.Pilaicardcard) && j < len(StatusPlay.Pilaicardcard[i]); j++ {
			StatusPlay.Pilaicardcard[i][j] = _item.Ctx.Pilaicardcard[j]
		}

		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	//StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	//StatusPlay.CardLeft.Random = sp.RepertoryCardArray.Random
	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
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
		StatusPlay.VecGangCard = static.HF_BytesToInts(player.Ctx.VecGangCard)
		if (StatusPlay.ActionMask & static.WIK_QIANG) != 0 {
			StatusPlay.ActionCard = byte(i)
		}

		if player.Ctx.CheckTimeOut != 0 {
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
				if v.Ctx.UserAction > 0 {
					StatusPlay.Overtime = 0
					break
				}
			}
		}
		//扑克数据
		StatusPlay.CardCount, StatusPlay.CardData = sp.Logic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
		StatusPlay.IsOutCard = true
		if player.Ctx.UserAction32&static.WIK_PENG != 0 {
			StatusPlay.IsOutCard = false
		} else if player.Ctx.UserAction32&static.WIK_GANG != 0 {
			StatusPlay.IsOutCard = false
		}
		sp.VecGameDataAllP[i] = append(sp.VecGameDataAllP[i], StatusPlay) //保存，用于汇总计算
	}
}

func (sp *SportSSAH) GetGameConfig() *static.GameConfig {
	return &sp.Config
}

func (sp *SportSSAH) Tojson() string {
	var _json components2.GameJsonSerializer
	_json.ToJson(&sp.Metadata)
	_json.GameCommonToJson(&sp.Common)
	return static.HF_JtoA(&_json)
}

func (sp *SportSSAH) Unmarsha(data string) {
	if data != "" {
		var _json components2.GameJsonSerializer
		json.Unmarshal([]byte(data), &_json)
		_json.Unmarsha(&sp.Metadata)
		_json.JsonToStruct(&sp.Common)
		sp.ParseRule()
		sp.Logic.Rule = sp.Rule
		sp.Logic.HuType = sp.HuType
		sp.Logic.SetMagicCard(sp.MagicCard)
		sp.Logic.SetPiZiCard(sp.PiZiCard)
		sp.Logic.SetPiZiCards(sp.PiZiCards)
		sp.SetOfflineRoomTime(sp.Rule.Overtime_offdiss * 60)
	}
}

type FriendRuleSSAH struct {
	Difen            int    `json:"difen"`            //底分
	Quwan            string `json:"quwan"`            //去万
	Nowan            string `json:"nowan"`            //去万
	Fewerstart       string `json:"fewerstart"`       //可少人开局
	BNineSecondRoom  string `json:"b9room"`           //9秒房
	Overtime_trust   int    `json:"overtime_trust"`   // 超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` // 超时解散
	Fleetime         int    `json:"fleetime"`         //客户端传来的 游戏开始前离线踢人时间
	Dissmiss         int    `json:"dissmiss"`         //解散次数,0不限制,12345对应限制次数

	Dima               int    `json:"dima"`               //底码
	Peimafangshi       int    `json:"peimafangshi"`       //配码方式
	Qiangganghu        int    `json:"qiangganghu"`        //自动抢杠胡
	Nochaohaohua       string `json:"nochaohaohua"`       //是否超豪华
	Nochihu            string `json:"nochihu"`            //是否吃胡
	Dahudianpao        int    `json:"dahudianpao"`        //大胡点炮是否奖码
	Endready           string `json:"endready"`           //小结算是否自动准备
	Overtime_offdiss   int    `json:"overtime_offdiss"`   //离线解散时间
	Overtime_applydiss int    `json:"overtime_applydiss"` //申请解散时间
	LookonSupport      string `json:"LookonSupport"`      //本局游戏是否支持旁观

}

func (sp *SportSSAH) InitOptHandler() {
	sp.OptHandlerHash = make(map[byte]OptHandler)
	sp.RegisterHandler(static.WIK_NULL, sp.OnUserQ)
	sp.RegisterHandler(static.WIK_LEFT, sp.OnUserE)
	sp.RegisterHandler(static.WIK_CENTER, sp.OnUserE)
	sp.RegisterHandler(static.WIK_RIGHT, sp.OnUserE)
	sp.RegisterHandler(static.WIK_PENG, sp.OnUserP)
	sp.RegisterHandler(static.WIK_GANG, sp.OnUserG)
	sp.RegisterHandler(static.WIK_FILL, sp.OnUserG)
	sp.RegisterHandler(static.WIK_CHI_HU, sp.OnUserH)
	sp.RegisterHandler(static.WIK_QIANG, sp.OnUserH)
}

func (sp *SportSSAH) RegisterHandler(code byte, handler OptHandler) {
	sp.OptHandlerHash[code] = handler
}

// ! 重置桌子数据
func (sp *SportSSAH) ResetTable() {
	sp.Metadata.Reset()
	for _, v := range sp.PlayerInfo {
		//v.Reset() //重置玩家所有数据 这里面清理了下跑数据 然后定漂不能用
		v.Ctx.Reset()
		v.UserReady = false
		//v.OnBegin() //重置开局玩家数据  【有一些重复的清理操作】
	}
	// 以下GameMeta没有的数据  例如：骰子值
	sp.SiceCount = components2.MAKEWORD(byte(1), byte(1))
	sp.FanScore = [4]meta2.Game_mj_fan_score{}
	sp.LastGangKind = static.INVALID_BYTE
	sp.LastGangIndex = static.INVALID_BYTE
	sp.KzhDatas = [4]int{0, 0, 0, 0}

}

// ! 解析配置的任务
func (sp *SportSSAH) ParseRule() {
	strRule := sp.GetTableInfo().Config.GameConfig
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "strRule:"+strRule)
	sp.Rule.Cardsclass = mahlib2.CARDS_NOMOR
	sp.Rule.Cardsclass ^= mahlib2.CARDS_WITHOUT_WIND
	sp.Rule.Cardsclass ^= mahlib2.CARDS_WITHOUT_ZHONG
	sp.Rule.Cardsclass ^= mahlib2.CARDS_WITHOUT_FA
	sp.Rule.Cardsclass ^= mahlib2.CARDS_WITHOUT_BAI

	sp.Rule.FangZhuID = sp.GetTableInfo().Creator
	sp.Rule.JuShu = sp.GetTableInfo().Config.RoundNum
	sp.Rule.CreateType = sp.FriendInfo.CreateType
	sp.Rule.DiFen = 1
	sp.Rule.NineSecondRoom = false

	sp.Rule.TrusteeCostSharing = true //托管的人扣房费

	// 写游戏默认配置 （案例 ：self.Rule.DiFen = 1）
	if len(strRule) == 0 {
		return
	}
	// TODO 解析自定义房间游戏配置
	var _msg FriendRuleSSAH
	if err := json.Unmarshal(static.HF_Atobytes(strRule), &_msg); err == nil {
		//解析客户端发来的自定义游戏配置
		xlog.Logger().Debugf("规则", _msg)
		sp.Rule.DiFen = _msg.Difen
		//if int(sp.GetConfig().PlayerCount) == 2 {
		//	sp.Rule.NoWan = true
		//	sp.Rule.Cardsclass ^= common.CARDS_WITHOUT_WAN
		//} else if int(sp.GetConfig().PlayerCount) == 3 && _msg.Quwan == "true" {
		//	sp.Rule.NoWan = true
		//	sp.Rule.Cardsclass ^= common.CARDS_WITHOUT_WAN
		//}
		//sp.Rule.HasPao = _msg.Xuanpiao != 0
		//20201013 苏大强 新增2人
		if /*sp.GetPlayerCount() == 2 ||*/ _msg.Quwan == "true" || _msg.Nowan == "true" {
			sp.Rule.Cardsclass ^= mahlib2.CARDS_WITHOUT_WAN
		}
		sp.Rule.IsHaveMagic = false

		sp.Rule.DiMa = _msg.Dima                     //底码
		sp.Rule.MaType = _msg.Peimafangshi           //配码方式
		sp.Rule.IsAutoQGH = _msg.Qiangganghu         //自动抢杠胡
		sp.Rule.IsCHH = _msg.Nochaohaohua == "false" //是否超豪华
		sp.Rule.IsChiHu = _msg.Nochihu == "true"     //是否吃胡
		sp.Rule.IsDHJM = _msg.Dahudianpao            //大胡点炮是否奖码

		sp.Rule.Endready = _msg.Endready == "true"

		sp.Rule.Fewerstart = _msg.Fewerstart == "true"
		sp.Rule.NineSecondRoom = _msg.BNineSecondRoom == "true"
		sp.Rule.Overtime_trust = _msg.Overtime_trust //超时托管
		//sp.Rule.Overtime_trust = 3     //超时托管
		sp.Rule.Overtime_dismiss = _msg.Overtime_dismiss //超时解散
		sp.Rule.DissmissCount = _msg.Dissmiss
		//syslog.Logger().Debug("离线踢人解析", _msg.Fleetime)
		if _msg.Fleetime != 0 {
			sp.SetOfflineRoomTime(_msg.Fleetime)
		}

		if sp.Rule.DissmissCount != 0 {
			sp.SetDissmissCount(sp.Rule.DissmissCount)
		}
		sp.Rule.Overtime_offdiss = _msg.Overtime_offdiss
		sp.Rule.Overtime_applydiss = _msg.Overtime_applydiss

		sp.SetDismissRoomTime(sp.Rule.Overtime_applydiss * 60)
		if _msg.LookonSupport == "" {
			sp.Config.LookonSupport = true
		} else {
			sp.Config.LookonSupport = _msg.LookonSupport == "true"
		}

	}
}

// ! 游戏开场  OnBegin
func (sp *SportSSAH) OnBegin() {
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始OnBegin....大局游戏只会走一次..")
	sp.ResetTable() //重置桌子
	for _, v := range sp.PlayerInfo {
		v.OnBegin() //重置开局玩家数据  【有一些重复的清理操作】
	}
	//  第一局随机坐庄
	sp.BankerUser = uint16(static.HF_GetRandom(100) % sp.GetPlayerCount())

	sp.CurCompleteCount = 0
	sp.VecGameEnd = make([]static.Msg_S_GameEnd, 0)
	sp.VecGameDataAllP = [4][]static.CMD_S_StatusPlay{}
	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()
	sp.OnGameStart()
}

// 游戏开始
func (sp *SportSSAH) OnGameStart() {
	if !sp.CanContinue() {
		return
	}
	sp.GetTable().SetBegin(true)
	////解析规则
	sp.ParseRule()
	sp.Logic.Rule = sp.Rule
	for _, v := range sp.PlayerInfo { //初始化玩家数据
		v.Ctx.CleanWeaveItemArray()
		v.Ctx.InitCardIndex()
		//v.Ctx.CleanXiaPao()
	}
	sp.ReplayRecord.Reset()             //重置战绩回放记录
	sp.OnStartNextGame()                //设置当前小局游戏状态
	sp.SetGameStatus(static.GS_MJ_PLAY) //设置游戏状态 暂时不知是那种表达意思
	sp.SetOfflineRoomTime(1800)
	if sp.Rule.HasPao {
		sp.SendPaoSetting()
	} else {
		sp.StartNextGame()
	}
}

// 发送下跑对话框
func (sp *SportSSAH) SendPaoSetting() {
	sp.PayPaoStatus = true //设置下跑的状态
	var PaoSetting static.Msg_S_PaoSetting
	for _, v := range sp.PlayerInfo { //i := 0; i < self.GetPlayerCount(); i++ {
		if v == nil {
			continue
		}
		if v.Ctx.VecXiaPao.Status { //定漂
			var serverMsgXiaPao static.Msg_C_Xiapao
			serverMsgXiaPao.Num = v.Ctx.VecXiaPao.Num       //下跑数据
			serverMsgXiaPao.Id = v.Ctx.VecXiaPao.Id         //玩家ID
			serverMsgXiaPao.Status = v.Ctx.VecXiaPao.Status //是否以后每局自动下跑
			serverMsgXiaPao.ByClient = false                //来源
			sp.OnUserClientPao(&serverMsgXiaPao)
		} else {
			PaoSetting.PaoCount = 0
			PaoSetting.Overtime = sp.LimitTime
			PaoSetting.Always = false
			PaoSetting.BankerUser = sp.BankerUser
			sp.SendPersonMsg(consts.MsgTypeGamePaoSetting, PaoSetting, v.Seat)
		}
	}
	sp.setPaoTimeOut() //设置下跑的操作时间
	//发送旁观数据
	sp.SendTableLookonMsg(consts.MsgTypeGamePiaoSetting, PaoSetting)

}

// 计时器事件
func (sp *SportSSAH) OnEventTimer(dwTimerID uint16, wBindParam int64) bool {
	//游戏定时器
	//if sp.Rule.NineSecondRoom { //TODO 倒计时 时间到自动操作 ！

	if sp.PayPaoStatus {
		for _, v := range sp.PlayerInfo {
			if v == nil {
				continue
			}
			if !v.Ctx.VecXiaPao.Status {
				_msg := sp.Greate_XiaPaomsg(v.Uid, false, v.Ctx.VecXiaPao.Status, v.Seat == sp.BankerUser)
				// syslog.Logger().Debug(fmt.Sprintf("自动操作（吃胡）玩家（%d）座位号（%d）可执行的操作（%d）放弃,当前用户（%d）消息：%v", _userItem.Uid, wChairID, _userItem.Ctx.UserAction, self.m_wCurrentUser, _msg))
				if !sp.OnUserClientPao(_msg) {
					sp.OnWriteGameRecord(v.Seat, "服务器自动选飘时，可能被客户端抢先了")
				}
			}
		}
	} else {
		//if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
		//	//到时间了
		//	sp.OnAutoOperateByChair(TablePerson.Seat, true)

		//}
		//if sp.Rule.Overtime_trust != -1{
		//	if sp.GetGameStatus() == public.GS_MJ_PLAY {
		//		sp.OnAutoOperate(TablePerson.Seat,true)
		//	}
		//}

		if sp.Rule.Overtime_trust != -1 { //&& dwTimerID == GameTime_Nine {
			if TablePerson := sp.GetUserItemByUid(wBindParam); TablePerson != nil {
				//到时间了
				sp.OnAutoOperate(TablePerson.Seat, true)
			}
		}

	}
	//}
	return true
}

// 桌子定时器
func (sp *SportSSAH) setPaoTimeOut() {
	t := 9                                      //吓跑倒计时 暂时定义 9S
	sp.LimitTime = time.Now().Unix() + int64(t) //public.GAME_OPERATION_TIME_12
	if sp.Rule.NineSecondRoom {
		sp.GameTimer.SetLimitTimer(t) //public.GAME_OPERATION_TIME_12)
	}
}

// ! 玩家开启超时
func (sp *SportSSAH) LockTimeOut(cUser uint16, iTime int) {
	_userItem := sp.GetUserItemByChair(cUser)
	if _userItem == nil || iTime == 0 {
		return
	}
	checktime := iTime
	if checktime < 1 {
		checktime = static.GAME_OPERATION_TIME_12
	}
	sp.LimitTime = time.Now().Unix() + int64(checktime)
	if _userItem.CheckTRUST() {
		//托管状态
		checktime = 1
	}
	_userItem.Ctx.CheckTimeOut = time.Now().Unix() + int64(checktime)
	_userItem.Ctx.Timer.SetTimer(components2.GameTime_Nine, checktime)
}

// ! 玩家关闭超时
func (sp *SportSSAH) UnLockTimeOut(cUser uint16) {
	if cUser < 0 || cUser > uint16(sp.GetPlayerCount()) {
		return
	}
	_userItem := sp.GetUserItemByChair(cUser)
	if _userItem == nil {
		return
	}
	_userItem.Ctx.CheckTimeOut = 0

	if sp.Rule.NineSecondRoom {
		_userItem.Ctx.Timer.KillTimer(components2.GameTime_12)
	}
}

// 发送一张牌 self.m_cbSendCardData存储这张牌值
func (sp *SportSSAH) SetLimitTime() int64 {
	if sp.Rule.Overtime_trust != -1 {
		sp.LimitTime = time.Now().Unix() + int64(sp.Rule.Overtime_trust)
	} else {
		sp.LimitTime = time.Now().Unix() + static.GAME_OPERATION_TIME_12
	}
	return sp.LimitTime
}

// 发送一张牌 self.m_cbSendCardData存储这张牌值
func (sp *SportSSAH) SendOneCard(userItem *components2.Player) {
	if sp.Logic.IsValidCard(userItem.Ctx.WantCard) {
		defer userItem.Ctx.CleanWant()
		cardStr := sp.Logic.SwitchToCardNameByData(userItem.Ctx.WantCard, 1)
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
	sp.SendCardCount++
	sp.SendCardData = sp.DrawOne()
	userItem.Ctx.DispatchCard(sp.SendCardData)
}

// 从牌库摸一张牌
func (sp *SportSSAH) DrawOne() byte {
	sp.LeftCardCount--
	cardData := sp.RepertoryCard[sp.LeftCardCount]
	index := sp.FindCardIndexInRepertoryCards(cardData)
	if index == static.INVALID_BYTE {
		cardStr := sp.Logic.SwitchToCardNameByData(cardData, 1)
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

// 找出痞子 癞子
func (sp *SportSSAH) SetMagicWithPi() {
	if sp.Rule.IsHaveMagic {
		sp.PiZiCard = sp.DrawOne()
		cbValue := byte(sp.PiZiCard & static.MASK_VALUE)
		cbColor := byte(sp.PiZiCard & static.MASK_COLOR)
		if cbValue == 9 && cbColor <= 0x20 {
			cbValue = 0
		}
		sp.MagicCard = (cbValue + 1) | cbColor
		sp.PiZiCards = append(sp.PiZiCards, sp.PiZiCard)
	} else {
		sp.PiZiCard = static.INVALID_BYTE
		sp.MagicCard = static.INVALID_BYTE

	}
	// 确定癞子
	sp.Logic.SetMagicCard(sp.MagicCard)
	// 传递皮子牌
	sp.Logic.SetPiZiCard(sp.PiZiCard)

	sp.Logic.SetPiZiCards(sp.PiZiCards)

}

// ! 开始下一局游戏
func (sp *SportSSAH) StartNextGame() {
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始StartNextGame...=>"+sp.GetTableInfo().Config.GameConfig)
	//设置状态

	sp.CurCompleteCount++ // 框架发送开始游戏后开始计算当前这一轮的局数
	sp.LastSendCardUser = uint16(static.INVALID_CHAIR)
	for _, v := range sp.PlayerInfo {
		v.OnNextGame()
	}
	//组合扑克
	sp.MagicCard = 0x00
	sp.LeftCardCount = 0
	sp.RepertoryCard = []byte{}

	//发送最新状态
	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.SendUserStatus(i, static.US_PLAY) //把状态发给其他人
	}
	//sp.SetOfflineRoomTime(0)
	sp.SetOfflineRoomTime(sp.Rule.Overtime_offdiss * 60)
	sp.SetDismissRoomTime(sp.Rule.Overtime_applydiss * 60)
	//--------------
	//混乱扑克
	_randTmp := time.Now().Unix() + int64(sp.GetTableId()+sp.KIND_ID*100+sp.GetSortId()*1000)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	sp.SiceCount = components2.MAKEWORD(byte(rand.Intn(1000)%6+1), byte(rand.Intn(1000)%6+1))
	sp.LeftCardCount = byte(len(sp.RepertoryCard))
	sp.LeftBu = 10 //剩下的补牌数
	sp.LeftCardCount, sp.RepertoryCard = sp.Logic.RandCardData()
	sp.CreateLeftCardArray(sp.GetProperPNum(), int(sp.LeftCardCount), sp.Rule.IsHaveMagic)
	for _, v := range sp.PlayerInfo {
		if v.Seat != static.INVALID_CHAIR {
			sp.LeftCardCount -= static.MAX_COUNT - 1
			v.Ctx.SetCardIndex(&sp.Rule, sp.RepertoryCard[sp.LeftCardCount:], static.MAX_COUNT-1)
		}
	}

	//////////////读取配置文件设置牌型begin//////////////////////////////////
	e := sp.InitDebugCards("majorTestSSAH", &sp.RepertoryCard, &sp.BankerUser)
	if e != nil {
		fmt.Println(fmt.Sprintf("%v", e))
		sp.OnWriteGameRecord(static.INVALID_CHAIR, e.Error())
	}
	//////////////读取配置文件设置牌型end////////////////////////////////////
	//发送扑克---这是发送给庄家的第十四张牌
	BankerUser := sp.GetUserItemByChair(sp.BankerUser)
	sp.SendOneCard(BankerUser)

	//这一步翻皮子  找癞子
	sp.SetMagicWithPi()

	//写游戏日志
	sp.WriteGameRecord()

	//设置变量
	sp.ProvideCard = sp.SendCardData
	sp.ProvideUser = static.INVALID_CHAIR
	sp.CurrentUser = sp.BankerUser //供应用户
	sp.LastSendCardUser = sp.BankerUser

	//动作分析,只分析庄家的牌，只有庄家初始才可以杠或胡
	var GangCardResult static.TagGangCardResult
	a, b := sp.Logic.AnalyseGangCard(BankerUser, BankerUser.Ctx.CardIndex, nil, 0, &GangCardResult)
	BankerUser.Ctx.UserAction |= a
	if b { //开局必然进不来
		sp.SendGameNotificationMessage(BankerUser.GetChairID(), "弃杠后不能再杠")
	}
	sp.getHuRes(BankerUser, int(sp.SendCardData), true, false, false, false) //验胡
	xlog.Logger().Print("是否可以胡（开局）", BankerUser.Ctx.ChiHuResult.ChiHuKind2)
	//构造数据,发送开始信息
	var msgGameStart static.Msg_S_GameStart
	msgGameStart.SiceCount = sp.SiceCount
	msgGameStart.BankerUser = sp.BankerUser
	msgGameStart.CurrentUser = sp.CurrentUser
	msgGameStart.MagicCard = sp.PiZiCard
	msgGameStart.LeftCardCount = sp.LeftCardCount
	//	msgGameStart.Overtime = sp.SetLimitTime() //操作时间给客户端看的

	msgGameStart.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	msgGameStart.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	//msgGameStart.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou

	msgGameStart.CurrentCount = sp.CurCompleteCount //告知客户端，当前局数

	//开始新的一局记录(提出来)
	//sp.ReplayRecord.Reset()
	sp.ReplayRecord.PiziCard = sp.PiZiCard
	sp.ReplayRecord.LeftCardCount = sp.LeftCardCount

	//sp.LockTimeOut(sp.BankerUser, public.GAME_OPERATION_TIME) //开启玩家倒计时
	xlog.Logger().Print("托管时间：", sp.Rule.Overtime_trust)
	sp.LockTimeOut(sp.BankerUser, sp.Rule.Overtime_trust)

	//向每个玩家发送数据
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item == nil {
			continue
		}
		msgGameStart.Whotrust[i] = _item.CheckTRUST()
		//设置变量
		msgGameStart.UserAction = _item.Ctx.UserAction //把上面分析过的结果保存再发送到客户端
		_, msgGameStart.CardData = sp.Logic.SwitchToCardData2(_item.Ctx.CardIndex, msgGameStart.CardData)

		//记录玩家初始分
		UserItem := sp.GetUserItem(i)
		if UserItem != nil {
			sp.ReplayRecord.RecordHandCard[i] = msgGameStart.CardData
			sp.ReplayRecord.Score[i] = 0
			sp.ReplayRecord.UVitamin[UserItem.Info.Uid] = _item.UserScoreInfo.Vitamin
		}
		if uint16(i) == sp.BankerUser {
			msgGameStart.SendCardData = sp.SendCardData //发给庄家的第一张牌
			msgGameStart.Overtime = _item.Ctx.CheckTimeOut
		} else {
			msgGameStart.SendCardData = static.INVALID_BYTE
			msgGameStart.Overtime = sp.SetLimitTime()
		}
		//发送数据
		sp.SendPersonMsg(consts.MsgTypeGameStart, msgGameStart, uint16(i))
	}

	if BankerUser.Ctx.UserAction != 0 {
		sp.ResumeUser = sp.CurrentUser //不知道有什么用
		sp.SendOperate()
	}
	//发送旁观数据
	msgGameStart.SendCardData = static.INVALID_BYTE
	msgGameStart.CardData = [14]byte{}
	msgGameStart.UserAction = 0
	sp.SendTableLookonMsg(consts.MsgTypeGameStart, msgGameStart)

}

// ! 游戏结束, 流局结束，统计积分
func (sp *SportSSAH) OnEventGameEnd(wChairID uint16, cbReason byte) bool {
	if sp.GetGameStatus() == static.GS_MJ_END && cbReason == static.GER_NORMAL {
		return true //正常结算时防止重入，框架的解散和离开，是可以重入的
	}
	// 清除超时检测
	for _, v := range sp.PlayerInfo {
		v.Ctx.CheckTimeOut = 0
	}
	switch cbReason { //TODO 结束游戏
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

func (sp *SportSSAH) GetMsgGameEnd(wChairID uint16, cbReason byte) *static.Msg_S_GameEnd {
	var msgGameEnd static.Msg_S_GameEnd
	msgGameEnd.LastSendCardUser = sp.LastSendCardUser
	msgGameEnd.EndStatus = cbReason
	msgGameEnd.MagicCard = sp.MagicCard
	msgGameEnd.Contractor = static.INVALID_CHAIR
	msgGameEnd.ProvideUser = wChairID
	msgGameEnd.ChiHuCard = sp.ChiHuCard
	msgGameEnd.ChiHuUserCount = 1
	msgGameEnd.KaiKou = sp.Rule.KouKou
	msgGameEnd.TheOrder = sp.CurCompleteCount
	msgGameEnd.BaseScore = sp.Rule.DiFen

	for i, l := 0, sp.GetPlayerCount(); i < l; i++ {
		_userItem := sp.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		isHuStatus := -1
		if int(sp.ProvideUser) == i {
			isHuStatus = 0
		} else {
			isHuStatus = 1
		}
		msgGameEnd.StrEnd[i] += sp.GetGameEndStr(i, isHuStatus, int(sp.ProvideUser))
		// 保存四家开口的次数（明杠，吃、碰）的次数
		msgGameEnd.OperateCount[i] = uint16(_userItem.Ctx.WeaveItemCount - _userItem.Ctx.HidGang)
		//	GameEnd.OperateCount[i] = uint16(self.m_cbWeaveItemCount[i] - _userItem.ctx.HidGang)
		if sp.Rule.KouKou == false && msgGameEnd.OperateCount[i] > 1 {
			msgGameEnd.OperateCount[i] = 1
		}
		//
		// 保存四家明杠的次数
		msgGameEnd.ShowGangCount[i] = uint16(_userItem.Ctx.ShowGang)
		// 保存四家暗杠的次数
		msgGameEnd.HideGangCount[i] = uint16(_userItem.Ctx.HidGang)
		// 保存四家丢赖子的次数
		msgGameEnd.MagicOutCount[i] = uint16(_userItem.Ctx.MagicCardOut)
		//跑分
		msgGameEnd.PaoCount[i] = 0xFF
		msgGameEnd.WeaveItemArray[i] = _userItem.Ctx.WeaveItemArray
		msgGameEnd.WeaveItemCount[i] = _userItem.Ctx.WeaveItemCount
		//拷贝四个玩家的扑克
		//syslog.Logger().Print("_userItem.Ctx.CardIndex", _userItem.Ctx.CardIndex)
		msgGameEnd.CardCount[i], msgGameEnd.CardData[i] = sp.Logic.SwitchToCardData2(_userItem.Ctx.CardIndex, msgGameEnd.CardData[i])
		//syslog.Logger().Print("玩家的手牌游戏结束了",  msgGameEnd.CardData[i])

	}
	msgGameEnd.LeftCardCount = sp.GetValidCount()
	for i := 0; i < int(sp.GetValidCount()); i++ {
		msgGameEnd.RepertoryCard[i] = sp.RepertoryCard[i]
	}
	msgGameEnd.HuangZhuang = sp.HaveHuangZhuang
	return &msgGameEnd
}

// 写回放
func (sp *SportSSAH) WriteOutDate(gameEnd *static.Msg_S_GameEnd) {
	sp.ReplayRecord.BigHuKind = gameEnd.BigHuKind
	sp.ReplayRecord.ProvideUser = gameEnd.ProvideUser
	sp.ReplayRecord.PiziCard = sp.PiZiCard
	if sp.HaveHuangZhuang {
		sp.addReplayOrder(0, info2.E_HuangZhuang, 0)
		sp.ReplayRecord.BigHuKind = 2
		sp.ReplayRecord.ProvideUser = 9
		sp.OnWriteGameRecord(static.INVALID_CHAIR, "游戏荒庄")
	}
	sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
	sp.ReplayRecord.Reset()
}

// 写分
func (sp *SportSSAH) WriteGameDate(gameEnd *static.Msg_S_GameEnd) {
	if !sp.PayPaoStatus {
		for _, v := range sp.PlayerInfo {
			if v == nil {
				continue
			}
			if v.GetChairID() != static.INVALID_CHAIR {
				wintype := static.ScoreKind_Draw
				if gameEnd.GameScore[v.GetChairID()] > 0 {
					//fmt.Println(fmt.Sprintf("赢家（%d）分数（%d）", v.GetChairID(), GameEnd.GameScore[v.GetChairID()]))
					wintype = static.ScoreKind_Win
				} else {
					//fmt.Println(fmt.Sprintf("平或者输家（%d）分数（%d）", v.GetChairID(), GameEnd.GameScore[v.GetChairID()]))
					wintype = static.ScoreKind_Lost //
				}
				sp.TableWriteGameDate(int(sp.CurCompleteCount), v, wintype, gameEnd.GameScore[v.GetChairID()]) //+ gameEnd.GameAdjustScore[v.GetChairID()]
			}
		}
	}
}

// ! 结束，结束游戏
func (sp *SportSSAH) OnGameEndNormal(wChairID uint16, cbReason byte) bool {
	//设置游戏结束状态
	sp.SetGameStatus(static.GS_MJ_END)

	msgGameEnd := sp.GetMsgGameEnd(wChairID, cbReason)

	sp.GetScoreOnHu(msgGameEnd)

	//漂胡分
	var OperateResult static.Msg_S_OperateResult //构建消息体
	OperateResult.ScoreOffset = msgGameEnd.GameScore
	sp.SendTableMsg(consts.MsgTypeGameOperateResult, OperateResult)
	sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, OperateResult)
	sp.VecGameEnd = append(sp.VecGameEnd, *msgGameEnd) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, msgGameEnd)
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, msgGameEnd)
	sp.SaveGameData()
	// 数据库写出牌记录
	sp.ReplayRecord.EndInfo = msgGameEnd
	// sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
	// 写完后清除数据
	if sp.CurCompleteCount == 1 {
		sp.TableDeleteFangKa(sp.CurCompleteCount)
	}
	sp.WriteOutDate(msgGameEnd)
	sp.WriteGameDate(msgGameEnd)
	sp.ReplayRecord.Reset()
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu { //局数够了
		sp.CalculateResultTotal(static.GER_NORMAL, wChairID, 0) //计算总发送总结算
		//通知框架结束游戏
		sp.SetGameStatus(static.GS_MJ_FREE)
		if !sp.IsLocalDebug {
			sp.ConcludeGame()
		}
	}
	sp.OnGameEnd()
	if !(int(sp.CurCompleteCount) >= sp.Rule.JuShu) && sp.CurCompleteCount != 0 {
		if sp.Rule.Overtime_dismiss != -1 {
			check := false
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
		// //自动开始下一局
		//}
		if sp.Rule.Endready {
			sp.SetAutoNextTimer(10)
		}

	}
	sp.ResetTable() // 复位桌子，如果此时不复位，可能导致解散房间时，发送错误的结算消息给客户端
	return true

}

// ! 强退，结束游戏
func (sp *SportSSAH) OnGameEndUserLeft(wChairID uint16, cbReason byte) bool {
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
		//GameEnd.GameAdjustScore[i] += _item.Ctx.StorageScore
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.Logic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])

		GameEnd.GameScore[i] += _item.Ctx.QiangScore
		//_huDetail.Private(_item.Seat, gameserver.TagQiangCuo, _item.Ctx.QiangScore, gameserver.DetailTypeADD)
		//玩家番数
		GameEnd.MaxFSCount[i] = uint16(sp.GetPlayerFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount

		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount
		//GameEnd.StrEnd[i] = _huDetail.GetSeatString1(_item.Seat)
		GameEnd.StrEnd[i] = sp.GetGameEndStr(-1, -1, -1)
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
	//sp.suanTaiZhuang(&GameEnd)

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
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, GameEnd)

	sp.SaveGameData()

	if sp.GetGameStatus() != static.GS_MJ_FREE {
		//数据库写出牌记录
		sp.ReplayRecord.EndInfo = &GameEnd
		sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
		// 写完后清除数据
		sp.ReplayRecord.Reset()

		//数据库写分
		for _, v := range sp.PlayerInfo {
			if v.Seat != static.INVALID_CHAIR {
				if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
					//sp.TableWriteGameDate(int(sp.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.GetChairID()]) //+ GameEnd.GameAdjustScore[v.GetChairID()]
				} else {
					//sp.TableWriteGameDate(int(sp.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
					sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.GetChairID()]) //+ GameEnd.GameAdjustScore[v.GetChairID()]
				}
			}
		}
	}

	//sp.UpdateOtherFriendDate(&GameEnd, true)
	//结束游戏
	sp.CalculateResultTotal(static.GER_USER_LEFT, wChairID, 0)
	//结束游戏 不重置局数
	//self.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()

	return true
}

// ! 解散，结束游戏
func (sp *SportSSAH) OnGameEndDissmiss(wChairID uint16, cbReason byte, cbSubReason byte) bool {

	var _huDetail components2.TagHuCostDetail
	_huDetail.Init()
	//变量定义
	var GameEnd static.Msg_S_GameEnd
	GameEnd.LastSendCardUser = sp.LastSendCardUser
	GameEnd.EndStatus = cbReason
	GameEnd.ProvideUser = static.INVALID_CHAIR
	GameEnd.Winner = static.INVALID_CHAIR
	GameEnd.Contractor = static.INVALID_CHAIR
	GameEnd.MagicCard = sp.MagicCard
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
		//GameEnd.GameAdjustScore[i] += _item.Ctx.StorageScore
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
		//_huDetail.Private(_item.Seat, gameserver.TagQiangCuo, _item.Ctx.QiangScore, gameserver.DetailTypeADD)
		//玩家番数
		GameEnd.MaxFSCount[i] = uint16(sp.GetPlayerFan(uint16(i)))
		//抢错次数
		GameEnd.QiangFailCount[i] = _item.Ctx.QiangFailCount
		GameEnd.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		GameEnd.WeaveItemCount[i] = _item.Ctx.WeaveItemCount

		if sp.Rule.HasPao {
			GameEnd.PaoCount[i] = uint16(_item.Ctx.VecXiaPao.Num)
		} else {
			GameEnd.PaoCount[i] = 0xFF
		}

		GameEnd.StrEnd[i] = _huDetail.GetSeatString1(_item.Seat)
		//拷贝四个玩家的扑克
		GameEnd.CardCount[i], GameEnd.CardData[i] = sp.Logic.SwitchToCardData2(_item.Ctx.CardIndex, GameEnd.CardData[i])
		GameEnd.StrEnd[i] += sp.GetGameEndStr(-1, -1, -1)
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
	//sp.suanTaiZhuang(&GameEnd)
	//发送信息
	sp.VecGameEnd = append(sp.VecGameEnd, GameEnd) //保存，用于汇总计算
	sp.SendTableMsg(consts.MsgTypeGameEnd, GameEnd)
	sp.SendTableLookonMsg(consts.MsgTypeGameEnd, GameEnd)

	sp.SaveGameData()
	switch cbSubReason {
	case 0:
		//游戏记录
		sp.OnWriteGameRecord(wChairID, "解散游戏OnGameEndDissmiss")

		if sp.GetGameStatus() != static.GS_MJ_FREE {
			//数据库写出牌记录
			sp.ReplayRecord.EndInfo = &GameEnd
			sp.TableWriteOutDate(int(sp.CurCompleteCount), sp.ReplayRecord)
			// 写完后清除数据
			sp.ReplayRecord.Reset()

			//数据库写分
			for _, v := range sp.PlayerInfo {
				if v.Seat != static.INVALID_CHAIR {
					if wChairID != static.INVALID_CHAIR && v.Seat == wChairID {
						//sp.TableWriteGameDate(int(sp.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
						sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.GetChairID()]) //+ GameEnd.GameAdjustScore[v.GetChairID()]
					} else {
						//sp.TableWriteGameDate(int(sp.CurCompleteCount), v, public.ScoreKind_pass, GameEnd.GameScore[v.Seat])
						sp.TableWriteGameDate(int(sp.CurCompleteCount), v, static.ScoreKind_pass, GameEnd.GameScore[v.GetChairID()]) // + GameEnd.GameAdjustScore[v.GetChairID()]
					}
				}
			}
		}
	case 1:
		sp.OnWriteGameRecord(wChairID, "前面某个时刻程序出错过，需要排查错误日志，无法恢复这局游戏，解散游戏OnGameEndErrorDissmis")
	}

	//sp.UpdateOtherFriendDate(&GameEnd, true)
	// 写总计算
	sp.CalculateResultTotal(static.GER_DISMISS, wChairID, cbSubReason)
	//结束游戏
	//self.SetGameStatus(public.GS_MJ_FREE)
	sp.ConcludeGame()

	return true
}

// ! 得到玩家番数
func (sp *SportSSAH) GetPlayerFan(wCurrentUser uint16) int {
	_userItem := sp.GetUserItemByChair(wCurrentUser)
	if _userItem == nil {
		return 0
	}

	FanNum := 0

	FanNum += int(_userItem.Ctx.PiZiGangCount + _userItem.Ctx.PiZiCardOut)

	// 赖子杠+2番
	FanNum += int(_userItem.Ctx.MagicCardGang+_userItem.Ctx.MagicCardOut) * 2
	// 明杠+1番
	FanNum += int(_userItem.Ctx.ShowGang + _userItem.Ctx.XuGang)
	// 暗杠+2番
	FanNum += int(_userItem.Ctx.HidGang) * 2

	return FanNum
}

// ! 得到某个用户开口的次数,吃，碰，明杠的次数
func (sp *SportSSAH) GetUserOpenMouth(wChairID uint16) uint16 {
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return uint16(0)
	}
	return uint16(_userItem.Ctx.WeaveItemCount - _userItem.Ctx.HidGang)
}

// ! 计算总发送总结算
func (sp *SportSSAH) CalculateResultTotal(cbReason byte, wChairID uint16, cbSubReason byte) {
	// 给客户端发送总结算数据
	var balanceGame static.Msg_S_BALANCE_GAME
	balanceGame.Userid = sp.Rule.FangZhuID
	balanceGame.CurTotalCount = sp.CurCompleteCount // 总盘数
	sp.TimeEnd = time.Now().Unix()                  //大局结束时间
	balanceGame.TimeStart = sp.TimeStart            //游戏大局开始时间
	balanceGame.TimeEnd = sp.TimeEnd

	for i := 0; i < len(sp.VecGameEnd); i++ {
		for j := 0; j < sp.GetPlayerCount(); j++ {
			balanceGame.GameScore[j] += sp.VecGameEnd[i].GameScore[j] // 总分
			//balanceGame.GameScore[j] += sp.VecGameEnd[i].GameAdjustScore[j]
			_userItem := sp.GetUserItemByChair(uint16(j))
			if _userItem != nil {
				balanceGame.ChiHuUserCount[j] = _userItem.Ctx.ChiHuUserCount
				balanceGame.ProvideUserCount[j] = _userItem.Ctx.ProvideUserCount
				balanceGame.FXScoreUserCount[j] = _userItem.Ctx.MaxScoreUserCount
				balanceGame.HHuUserCount[j] = _userItem.Ctx.BigHuUserCount
			}
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
	sp.SendTableLookonMsg(consts.MsgTypeGameBalanceGame, balanceGame)

	sp.resetEndDate()
}

// ! 重置优秀结束数据
func (sp *SportSSAH) resetEndDate() {
	sp.CurCompleteCount = 0
	sp.VecGameEnd = []static.Msg_S_GameEnd{}

	for _, v := range sp.PlayerInfo {
		v.OnEnd()
	}
}

func (sp *SportSSAH) CheckFen(wCurrentUser uint16, wProvideUser uint16, cbCurrentCard byte, ChiHuResult *static.TagChiHuResult, _detail *components2.TagHuCostDetail) bool {
	//TODO 算分
	HuScore := 0
	BaseScore := sp.Rule.BaseScore
	var hudetail components2.TagHuCostDetail
	hudetail.Init()

	FengDingScore := [4]int{60, 120, 200, 500}
	JinDingScore := [4]int{80, 150, 300, 600}
	FengDingUserCount := 0
	IsBigBu := sp.IsDaHu(ChiHuResult.ChiHuKind)

	// 胡牌番数
	WinFan := sp.FanScore[wCurrentUser].FanNum[wCurrentUser]

	sp.FanScore[wCurrentUser].Score = [4]int{0}

	if IsBigBu {
		DaHuScore := [4]int{3, 5, 10, 30}
		DaHu7DuiScore := [4][4]int{{3, 6, 12, 24}, {5, 10, 20, 40}, {10, 20, 40, 80}, {30, 60, 120, 480}}

		if (ChiHuResult.ChiHuKind & static.CHK_7_DUI_3) > 0 {
			HuScore += DaHu7DuiScore[BaseScore][3]
			hudetail.Private(wCurrentUser, components2.TagHu7Dui_3, 1, components2.DetailTypeFirst)
		} else if (ChiHuResult.ChiHuKind & static.CHK_7_DUI_2) > 0 {
			hudetail.Private(wCurrentUser, components2.TagHu7Dui_2, 1, components2.DetailTypeFirst)
			HuScore += DaHu7DuiScore[BaseScore][2]
		} else if (ChiHuResult.ChiHuKind & static.CHK_7_DUI_1) > 0 {
			HuScore += DaHu7DuiScore[BaseScore][1]
			hudetail.Private(wCurrentUser, components2.TagHu7Dui_1, 1, components2.DetailTypeFirst)
		} else if (ChiHuResult.ChiHuKind & static.CHK_7_DUI) > 0 {
			hudetail.Private(wCurrentUser, components2.TagHu7Dui, 1, components2.DetailTypeFirst)
			HuScore += DaHu7DuiScore[BaseScore][0]
		}
		if (ChiHuResult.ChiHuKind & static.CHK_QING_YI_SE) > 0 {
			hudetail.Private(wCurrentUser, components2.TagHuQinYiSe, 1, components2.DetailTypeFirst)
			HuScore += DaHuScore[BaseScore]
		}
		if (ChiHuResult.ChiHuKind & static.CHK_JIANG_JIANG) > 0 {
			hudetail.Private(wCurrentUser, components2.TagHuJiangYiSe, 1, components2.DetailTypeFirst)
			HuScore += DaHuScore[BaseScore]
		}
		if (ChiHuResult.ChiHuKind & static.CHK_PENG_PENG) > 0 {
			hudetail.Private(wCurrentUser, components2.TagHuPengPengHu, 1, components2.DetailTypeFirst)
			HuScore += DaHuScore[BaseScore]
		}
		if (ChiHuResult.ChiHuKind & static.CHK_QUAN_QIU_REN) > 0 {
			hudetail.Private(wCurrentUser, components2.TagHuQuanQiuRen, 1, components2.DetailTypeFirst)
			HuScore += DaHuScore[BaseScore]
		}

		for i := 0; i < sp.GetPlayerCount(); i++ {
			if uint16(i) == wCurrentUser {
				continue
			}
			LoseFan := sp.FanScore[wCurrentUser].FanNum[i]

			BeiShu := sp.GetBeiShuWithFan(WinFan + LoseFan)

			WinScore := HuScore * BeiShu
			sp.FanScore[wCurrentUser].Ding[i] = static.DING_NULL
			if WinScore >= FengDingScore[BaseScore] {
				WinScore = FengDingScore[BaseScore]
				FengDingUserCount++
				sp.FanScore[wCurrentUser].Ding[i] = static.DING_FENG
			}
			sp.FanScore[wCurrentUser].Score[i] -= WinScore
			sp.FanScore[wCurrentUser].Score[wCurrentUser] += WinScore
		}
	} else {
		yinhu := ChiHuResult.ChiHuKind&static.CHK_PING_HU_NOMAGIC != 0
		if yinhu {
			hudetail.Private(wCurrentUser, components2.TagHuHard, 1, components2.DetailTypeFirst)
		}
		// 自摸
		if wCurrentUser == wProvideUser {
			DaXueScore := [4]int{3, 5, 10, 30}
			XiaoXueScore := [4]int{2, 3, 5, 20}
			if sp.GetUserOpenMouth(wCurrentUser) > 0 {
				//小血
				HuScore += XiaoXueScore[BaseScore]
			} else {
				//大血
				HuScore += DaXueScore[BaseScore]
			}
			for i := 0; i < sp.GetPlayerCount(); i++ {
				if uint16(i) == wCurrentUser {
					continue
				}

				LoseFan := sp.FanScore[wCurrentUser].FanNum[i]

				BeiShu := sp.GetBeiShuWithFan(WinFan + LoseFan)

				sp.FanScore[wCurrentUser].DianNum[i] = float32(BeiShu)
				WinScore := HuScore * BeiShu
				sp.FanScore[wCurrentUser].Ding[i] = static.DING_NULL
				if WinScore >= FengDingScore[BaseScore] {
					WinScore = FengDingScore[BaseScore]
					FengDingUserCount++
					sp.FanScore[wCurrentUser].Ding[i] = static.DING_FENG
				}
				sp.FanScore[wCurrentUser].Score[i] -= WinScore
				sp.FanScore[wCurrentUser].Score[wCurrentUser] += WinScore
			}
		} else {
			FangChongScore := [4]int{1, 2, 5, 10}
			PeiChongScore := [4]float32{0.5, 1, 2.5, 5}
			for i := 0; i < sp.GetPlayerCount(); i++ {
				if uint16(i) == wCurrentUser {
					continue
				}
				LoseFan := sp.FanScore[wCurrentUser].FanNum[i]

				BeiShu := sp.GetBeiShuWithFan(WinFan + LoseFan)

				WinScore := 0
				if uint16(i) == wProvideUser {
					WinScore = FangChongScore[BaseScore] * BeiShu
				} else {
					WinScore = int(math.Ceil(float64(PeiChongScore[BaseScore]) * float64(BeiShu)))
				}
				sp.FanScore[wCurrentUser].Ding[i] = static.DING_NULL
				if WinScore >= FengDingScore[BaseScore] {
					WinScore = FengDingScore[BaseScore]
					sp.FanScore[wCurrentUser].Ding[i] = static.DING_FENG
					FengDingUserCount++
				}
				sp.FanScore[wCurrentUser].Score[i] -= WinScore
				sp.FanScore[wCurrentUser].Score[wCurrentUser] += WinScore
			}
		}
	}

	// 計算玩家倍數
	for i := 0; i < sp.GetPlayerCount(); i++ {
		sp.FanScore[wCurrentUser].DianNum[i] = float32(sp.GetBeiShuWithFan(sp.FanScore[wCurrentUser].FanNum[i]))
	}
	//fmt.Println(fmt.Sprintf("计算倍数（%v）",self.FanScore[wCurrentUser].DianNum))

	//计算封顶
	if FengDingUserCount == sp.GetPlayerCount()-1 {
		for i := 0; i < sp.GetPlayerCount(); i++ {
			sp.FanScore[wCurrentUser].Score[i] = 0
		}
		for i := 0; i < sp.GetPlayerCount(); i++ {
			if uint16(i) == wCurrentUser {
				continue
			}
			sp.FanScore[wCurrentUser].Score[i] -= JinDingScore[BaseScore]
			sp.FanScore[wCurrentUser].Ding[i] = static.DING_JIN
			sp.FanScore[wCurrentUser].Score[wCurrentUser] += JinDingScore[BaseScore]
		}
	}
	if _detail != nil {
		_detail.Add(&hudetail)
	}

	if FengDingUserCount == sp.GetPlayerCount()-1 {
		return true
	}

	return false
}

func (sp *SportSSAH) IsDaHu(ChiHuKind uint64) bool {
	if (ChiHuKind & ssahCheckHu.GameHuPPH) > 0 {
		return true
	} else if (ChiHuKind & ssahCheckHu.GameHuQYS) > 0 {
		return true
	} else if (ChiHuKind & ssahCheckHu.GameHu7D) > 0 {
		return true
	}
	//else if (ChiHuKind & ssahCheckHu.GameHu7D1) > 0 {
	//	return true
	//} else if (ChiHuKind & ssahCheckHu.GameHu7D2) > 0 {
	//	return true
	//} else if (ChiHuKind & ssahCheckHu.GameHu7D3) > 0 {
	//	return true
	//}
	return false
}

func (sp *SportSSAH) GetBeiShuWithFan(fan int) int {
	BeiShu := 1

	for i := 0; i < fan; i++ {
		BeiShu = BeiShu * 2
	}

	return BeiShu
}

// ! 玩家大胡数量
func (sp *SportSSAH) GetPlayerDaHuCount(ChiHuKind uint64) int {
	DaHuCount := 0
	isRun := true
	if (ChiHuKind&ssahCheckHu.GameHuPPH) != 0 && isRun {
		DaHuCount++
		isRun = false
	}
	if (ChiHuKind&ssahCheckHu.GameHuQYS) != 0 && isRun {
		DaHuCount++
		isRun = false
	}
	if (ChiHuKind&ssahCheckHu.GameHu7D) != 0 && isRun {
		DaHuCount++
		isRun = false
	}
	//if (ChiHuKind&ssahCheckHu.GameHu7D1) != 0 && isRun {
	//	DaHuCount++
	//	isRun = false
	//}
	//if (ChiHuKind&ssahCheckHu.GameHu7D1) != 0 && isRun {
	//	DaHuCount++
	//	isRun = false
	//}
	//if (ChiHuKind&ssahCheckHu.GameHu7D3) != 0 && isRun {
	//	DaHuCount++
	//	isRun = false
	//}
	return DaHuCount
}

// ! 拼装游戏结束玩家信息   //isHuStatus 0 自摸 1接炮 2抢杠
func (sp *SportSSAH) GetGameEndStr(SeatId int, isHuStatus int, provideUser int) string {
	GameEndStr := ""
	if SeatId == -1 {
		return GameEndStr
	}
	useritem := sp.GetUserItemByChair(uint16(SeatId))
	ChiHukind := useritem.Ctx.ChiHuResult.ChiHuKind2
	is7 := false
	if useritem.Ctx.VecXiaPao.Num == 0 {
		GameEndStr += "<color=#ffffff>不漂  </color>"
	} else {
		GameEndStr += "<color=#ffffff>漂  </color>"
	}

	if ChiHukind != ssahCheckHu.GameHuNull {
		if (ChiHukind & ssahCheckHu.GameHuPiH) != 0 {
			GameEndStr += "<color=#ffffff>平胡x1  </color>"
		} else if sp.Rule.TypeLaizi == 0 { //TypeLaizi==0 癞子不能杠 有大胡
			isStrDH := false
			if (ChiHukind & ssahCheckHu.GameHuPPH) != 0 {
				if sp.Rule.PphDouble {
					GameEndStr += "<color=#ffffff>碰碰胡x2  </color>"
					isStrDH = true
				} else {

				}
			}
			if (ChiHukind & ssahCheckHu.GameHu7D) != 0 {
				if sp.Rule.D7Double {
					GameEndStr += "<color=#ffffff>七对x2  </color>"
					isStrDH = true
				} else {
					//if sp.Rule.IsD7{
					//	GameEndStr += "<color=#ffffff>七对  </color>"
					//}
					//GameEndStr += "<color=#ffffff>七对  </color>"
				}
			}
			if (ChiHukind&ssahCheckHu.GameHuQYS) != 0 && !is7 {
				if sp.Rule.QysDouble {
					GameEndStr += "<color=#ffffff>清一色x2  </color>"
					isStrDH = true
				} else {
					//GameEndStr += "<color=#ffffff>清一色  </color>"
				}
			}
			if !isStrDH {
				GameEndStr += "<color=#ffffff>平胡x1  </color>"
			}
		}
		if sp.GangFlower {
			if sp.Rule.GkDouble {
				GameEndStr += "<color=#ffffff>杠开x2  </color>"
			} else {
				GameEndStr += "<color=#ffffff>杠开x1  </color>"
			}
		}

		if isHuStatus == 0 {
			GameEndStr += "<color=#ffffff>自摸x2  </color>"
		} else {
			//GameEndStr += "<color=#ffffff>炮胡  </color>"

			//if (ChiHukind & ssahCheckHu.GameHuQGH) != 0 {
			//	GameEndStr += "<color=#ffffff>抢杠 </color>"
			//} else {
			//	GameEndStr += "<color=#ffffff>接冲 </color>"
			//}
		}
	} else {
		if provideUser == SeatId {
			//GameEndStr += "<color=#ffffff>点炮 </color>"
		}
	}

	if !sp.HaveHuangZhuang {
		str := ""
		if useritem.Ctx.StorageScore > 0 {
			str = fmt.Sprintf("<color=#ffffff>杠分+%d</color>", useritem.Ctx.StorageScore)
		} else if useritem.Ctx.StorageScore < 0 {
			str = fmt.Sprintf("<color=#ffffff>杠分%d</color>", useritem.Ctx.StorageScore)
		}
		GameEndStr += str
	}

	//if useritem.Ctx.MagicCardOut > 0 {
	//	GameEndStr += fmt.Sprintf("<color=#ffffff>癞子</color>%d", useritem.Ctx.MagicCardOut)
	//}
	return GameEndStr
}

/*
  以下是服务器接受消息 ==================================================
*/

// ! 发送消息
func (sp *SportSSAH) OnMsg(msg *base2.TableMsg) bool {
	switch msg.Head {
	case consts.MsgTypeGameXuanpiao: //下跑
		var _msg static.Msg_C_Xiapao
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			if !sp.OnUserClientPao(&_msg) && sp.Rule.NineSecondRoom {
				sp.flashClient(_msg.Id, "限时到，服务器自动选飘")
			}
		}
	case consts.MsgTypeGameOutCard: //! 出牌消息
		var _msg static.Msg_C_OutCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			if sp.OnUserOutCard(&_msg) {
				return true
			} else {
				sp.flashClient(_msg.Id, "不允许的操作！！")
			}
		}
	case consts.MsgTypeGameOperateCard: //操作消息
		var _msg static.Msg_C_OperateCard
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			_msg.ByClient = true
			return sp.OnUserOperateCard(&_msg, false)
			//if !sp.OnUserOperateCard(&_msg, false) {
			//	sp.flashClient(_msg.Id, "限时到，服务器自动选弃")
			//}
		}
	case consts.MsgTypeGameGoOnNextGame: //下一局
		var _msg static.Msg_C_GoOnNextGame
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.OnUserClientNextGame(&_msg)
		}
	case consts.MsgCommonToGameContinue:
		opt, ok := msg.V.(*static.TagSendCardInfo)
		if ok {
			sp.DispatchCardData(opt.CurrentUser, opt.GangFlower)
		} else {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "common to game 断言失败。")
		}
	case consts.MsgTypeGameTrustee: //用户托管
		//var _msg public.Msg_C_Trustee
		var _msg static.Msg_S_DG_Trustee
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			sp.onUserTustee(&_msg)
		}
	case consts.MsgTypeGameTimeOutAutoHu: // 游戏内托管,超时自动胡
		var _msg Msg_C_GameTimeOutAutoHu
		if err := json.Unmarshal(static.HF_Atobytes(msg.Data), &_msg); err == nil {
			nChairID := sp.GetChairByUid(_msg.Id)
			sp.m_bAutoHu[nChairID] = _msg.AutoHu
		}
	default:
		//self.Common.OnMsg(msg)
	}
	return true
}

// 解除托管
func (sp *SportSSAH) onUserTustee(msg *static.Msg_S_DG_Trustee) bool {
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
			item.ChangeTRUST(true)
			//进入托管啥都不用做
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, tuoguan)

			return true
		} else if tuoguan.Trustee == false {
			item.ChangeTRUST(false)
			//如果是当前的玩家，那么重新设置一下开始时间
			if tuoguan.ChairID == sp.CurrentUser {
				//sp.DownTime = PlayTime-(now-nowTime)+GetCPUTickCount();
				_item := sp.GetUserItemByChair(sp.CurrentUser)
				if _item != nil {
					//if time.Now().Unix() < _item.Ctx.CheckTimeOut { //如果只剩下1秒了，就不重新算了//time.Now().Unix() + 1 < sp.LimitTime
					sp.LockTimeOut(_item.Seat, sp.Rule.Overtime_trust)
					//sp.setLimitedTime(int64(sp.PlayTime) + sp.PowerStartTime - time.Now().Unix() + 1)
					tuoguan.Overtime = _item.Ctx.CheckTimeOut
				}
			}
			sp.SendTableMsg(consts.MsgTypeGameTrustee, tuoguan)
			sp.SendTableLookonMsg(consts.MsgTypeGameTrustee, tuoguan)

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

func (sp *SportSSAH) OnUserClientPao(msg *static.Msg_C_Xiapao) bool {
	xlog.Logger().Print("玩家发送过来的下跑数据：", msg)
	chairID := sp.GetChairByUid(msg.Id)
	userItem := sp.GetUserItemByChair(chairID)
	userItem.Ctx.VecXiaPao.Status = msg.Status
	if userItem == nil {
		return true
	}
	if sp.Rule.NineSecondRoom {
		sp.OperateMutex.Lock()
		defer sp.OperateMutex.Unlock()
	}
	if chairID < meta2.MAX_PLAYER {
		if !userItem.Ctx.XiaPaor_safe(msg) { //防止重复
			return false
		}
		//广播吓跑的消息
		sp.BroadPao(chairID)
	}
	// 如果4个玩家都准备好了，自动开启下一局
	_beginCount := 0
	for _, v := range sp.PlayerInfo {
		if !v.Ctx.UserPaoReady {
			recordStr := fmt.Sprintf("椅子编号[%d] 还没有完成下跑", v.Seat)
			sp.OnWriteGameRecord(uint16(v.Seat), recordStr)
			break
		}
		_beginCount++
	}
	if _beginCount >= sp.GetPlayerCount() {
		sp.OnWriteGameRecord(chairID, "所有人都完成下跑了，开始游戏")
		sp.PayPaoStatus = false
		//游戏没有开始发牌
		if !sp.GameStartForXiapao {
			sp.StartNextGame()
		}
	}
	return true
}

// ! 下一局
func (sp *SportSSAH) OnUserClientNextGame(msg *static.Msg_C_GoOnNextGame) bool {
	// 局数校验 大于当前总局数，不处理此消息
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu || sp.GetGameStatus() != static.GS_MJ_END {
		return true
	}

	// 记录游戏开始时间
	sp.Common.GameBeginTime = time.Now()
	nChiarID := sp.GetChairByUid(msg.Id)
	sp.SendTableMsg(consts.MsgTypeGameGoOnNextGame, *msg)
	sp.SendTableLookonMsg(consts.MsgTypeGameGoOnNextGame, *msg)
	if nChiarID < uint16(sp.GetPlayerCount()) {
		//if sp.TuoGuanPlayer[nChiarID] {
		//	sp.onUserTustee(&public.Msg_S_DG_Trustee{
		//		ChairID: nChiarID,
		//		Trustee: false,
		//	})
		//}

		_item := sp.GetUserItemByChair(nChiarID)
		if _item != nil {
			_item.UserReady = true
		}
	}
	sp.SendUserStatus(int(nChiarID), static.US_READY) //把我的状态发给其他人

	// 如果4个玩家都准备好了，自动开启下一局
	for i := 0; i < sp.GetPlayerCount(); i++ {
		item := sp.GetUserItemByChair(uint16(i))
		if item != nil && !item.UserReady {
			if !item.CheckTRUST() {
				break
			} else {
				item.UserReady = true
			}
		}

		if i == sp.GetPlayerCount()-1 {
			// 复位桌子
			sp.ResetTable()
			sp.OnGameStart()
		}
	}
	return true
}

// ! 用户出牌
func (sp *SportSSAH) OnUserOutCard(msg *static.Msg_C_OutCard) bool { //返回false 刷新客户端
	xlog.Logger().Debug("OnUserOutCard")
	if sp.Rule.Overtime_trust != -1 {
		sp.OperateMutex.Lock()
		defer sp.OperateMutex.Unlock()
	}
	//效验状态
	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return false
	}

	ChairID := sp.GetChairByUid(msg.Id)
	//效验参数
	if ChairID != sp.CurrentUser {
		return false
	}

	if !sp.Logic.IsValidCard(msg.CardData) {
		return false
	}

	_userItem := sp.GetUserItemByChair(ChairID)
	if _userItem == nil {
		return false
	}
	sp.ProvideUser = ChairID
	sp.ProvideCard = msg.CardData

	////能杠不杠加入过杠
	if (_userItem.Ctx.UserAction & static.WIK_GANG) != 0 {
		//如果打出的牌不是蓄杠的，就加入弃杠
		for i := byte(0); i < 4; i++ {
			if _userItem.Ctx.WeaveItemArray[i].WeaveKind == static.WIK_PENG {
				if _userItem.Ctx.WeaveItemArray[i].CenterCard == sp.SendCardData {
					if msg.CardData != sp.SendCardData {
						//没有蓄杠，打出了非蓄杠的牌
						if _userItem.Ctx.AppendGiveUpGang_ex(sp.SendCardData) {
							sp.OnWriteGameRecord(_userItem.GetChairID(), fmt.Sprintf("能杠不杠,加入过杠,牌:%s", sp.Logic.SwitchToCardNameByData(sp.SendCardData, 1)))
						}
						break
					}
				}

			}
		}
	}

	//TODO 托管或者计时场
	//if sp.Rule.Overtime_trust != -1 {
	//	sp.OperateMutex.Lock()
	//	defer sp.OperateMutex.Unlock()
	//}

	//TODO 托管或者计时场
	if !msg.ByClient {
		// recordStr += "(托管)"
		//进入托管状态
		if !_userItem.CheckTRUST() {
			var msg = &static.Msg_S_DG_Trustee{
				ChairID: _userItem.Seat,
				Trustee: true,
			}
			if sp.onUserTustee(msg) {
				sp.OnWriteGameRecord(_userItem.Seat, "出牌超时进入托管")
			}
		}
	}

	//删除扑克
	if !_userItem.Ctx.OutCard(&sp.Rule, msg.CardData) {
		xlog.Logger().Debug("removecard failed")
		return false
	}
	//出牌丢进弃牌区
	var class byte = 0
	if _userItem.CheckTRUST() {
		class = 1
	}
	_userItem.Ctx.Discard_ex(msg.CardData, class)
	recordStr := fmt.Sprintf("%s，打出：%s", sp.Logic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(msg.CardData, 1))

	//只要是癞子痞子，就记录
	if msg.CardData == sp.MagicCard {
		if sp.Rule.IsHaveMagic && (sp.Rule.TypeLaizi == 3) {
			recordStr = fmt.Sprintf("%s，丢癞子：%s", sp.Logic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(msg.CardData, 1))
		}
	}
	//else if common.Findcard(sp.Logic.PiZiCards, msg.CardData) {
	//	_userItem.Ctx.PiZiCardOut++
	//	recordStr = fmt.Sprintf("%s，丢痞子：%s", sp.Logic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(msg.CardData, 1))
	//}
	//游戏记录
	sp.OnWriteGameRecord(ChairID, recordStr)

	//设置变量
	sp.SendStatus = true
	//出牌记录
	sp.OutCardCount++
	////////////////出的是一般的牌或赖子牌////////////////
	sp.OutCardUser = ChairID
	sp.OutCardData = msg.CardData

	//构造数据发送消息
	var msgOutCard static.Msg_S_OutCard
	msgOutCard.User = int(ChairID)
	msgOutCard.Data = msg.CardData
	msgOutCard.ByClient = msg.ByClient
	msgOutCard.Overtime = sp.SetLimitTime()
	sp.SendTableMsg(consts.MsgTypeGameOutCard, msgOutCard)
	sp.SendTableLookonMsg(consts.MsgTypeGameOutCard, msgOutCard)

	//用户切换
	//用户出牌，如果是杠开花状态 则 去掉
	if sp.GangFlower {
		sp.GangFlower = false
	}

	sp.CurrentUser = uint16(sp.GetNextFullSeat(ChairID))

	if (sp.ProvideCard == sp.MagicCard) && sp.Rule.IsHaveMagic && (sp.Rule.TypeLaizi == 3) {
		// 这里是癞子可以杠 癞子杠会写回放
		sp.OnUserOutMagic(_userItem)
		sp.OnUserSendCard(true)
	} else {
		//记录出牌
		//记录出牌
		if _userItem.CheckTRUST() {
			//赤壁记录下是不是托管出的牌
			sp.addReplayOrder(ChairID, info2.E_OutCard_TG, msg.CardData)
		} else {
			sp.addReplayOrder(ChairID, info2.E_OutCard, msg.CardData)
		}
		ok, autoHot := sp.EstimateUserRespond(ChairID, msg.CardData, static.EstimatKind_OutCard)
		if !ok {
			sp.OnUserSendCard(false)
			//出的牌没有人要了，可以无条件设置成false
			sp.GangHotStatus = false
		} else {
			for uid, isHot := range autoHot {
				if isHot {
					uItem := sp.GetUserItemByUid(uid)
					if uItem != nil {
						sp.OnWriteGameRecord(uItem.GetChairID(), "自动杠后炮")
						_msg := sp.Greate_Operatemsg(uItem.Uid, true, static.WIK_CHI_HU, sp.ProvideCard)
						if !sp.OnUserOperateCard(_msg, false) {
							sp.OnWriteGameRecord(uItem.GetChairID(), "自动杠后炮错误")
						}
					}
				}
			}
		}
	}

	// 解锁用户超时操作
	// sp.UnLockTimeOut(ChairID)
	return true
}

func (sp *SportSSAH) OnUserSendCard(GangFlower bool) {
	//if sp.GetValidCount() > 0 {
	if sp.GetValidCount() > sp.getValidNum() {
		//TODO     (暂时不让进  这里写判断 其他玩家是否 能操作这张牌)   ps：好像不需要在这里做  先就这么写吧
		if false {

		} else {
			sp.DispatchCardData(sp.CurrentUser, GangFlower)
		}
	} else { //这里是牌库的牌没了
		sp.HaveHuangZhuang = true
		sp.ChiHuCard = 0
		sp.ProvideUser = static.INVALID_CHAIR
		sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
	}
}

// 用户出或杠癞子事件
func (sp *SportSSAH) OnUserOutMagic(_userItem *components2.Player) bool {
	sp.CurrentUser = _userItem.GetChairID()
	_userItem.Ctx.OutMagicCard()
	_userItem.Ctx.MaxFanUserCount++
	var operateResult static.Msg_S_OperateResult
	operateResult.OperateUser = _userItem.GetChairID()
	operateResult.OperateCard = sp.MagicCard
	operateResult.OperateCode = static.WIK_GANG
	operateResult.ProvideUser = _userItem.GetChairID()
	//sp.GetScoreOffsetOnMagic(&operateResult.ScoreOffset, _userItem)
	if _userItem.CheckTRUST() {
		sp.addReplayOrder(_userItem.GetChairID(), info2.E_OutCard_TG_Magic, sp.MagicCard)
	} else {
		sp.addReplayOrder(_userItem.GetChairID(), info2.E_Gang_LaiziGand, sp.MagicCard)
	}

	sp.SendResponseByOperation(_userItem, &operateResult)
	return true
}

// 通过飘癞子得到得失分
func (sp *SportSSAH) GetScoreOffsetOnMagic(scoreOffset *[meta2.MAX_PLAYER]int, _userItem *components2.Player) {
	for i := 0; i < sp.GetPlayerCount(); i++ {
		index := uint16(i)
		if index == _userItem.GetChairID() {
			continue
		}
		cost := sp.GetCostScoreOnMagic(_userItem.GetChairID(), index)
		scoreOffset[index] -= cost
		scoreOffset[_userItem.GetChairID()] += cost
	}
}

// 在玩家飘癞子后，得到当前用户飘癞子每个人要扣多少分
func (sp *SportSSAH) GetCostScoreOnMagic(wCurrentUser uint16, wTagetUser uint16) int {
	if wCurrentUser == wTagetUser {
		sp.OnWriteGameRecord(wCurrentUser, "GetCostScoreOnMagic 自己不能和自己算")
		return 0
	}
	cItem := sp.GetUserItemByChair(wCurrentUser)
	tItem := sp.GetUserItemByChair(wTagetUser)
	if cItem == nil || tItem == nil {
		//syslog.Logger().Debug("GetCostScoreOnMagic 空指针")
		sp.OnWriteGameRecord(static.INVALID_CHAIR, fmt.Sprintf("飘癞子计算分数时空指针::座位号:%d<->%d", wCurrentUser, wTagetUser))
		return 0
	}
	total := cItem.Ctx.CurMagicOut + tItem.Ctx.CurMagicOut
	if total <= 0 {
		// 两家都没飘癞子 则没分
		return 0
	}
	// 倍数 = 2 的 两家癞子总个数 -1 次方
	result := int(math.Pow(float64(2), float64(total-1))) * sp.Rule.DiFen
	return result
}

// ! 派发扑克
func (sp *SportSSAH) DispatchCardData(wCurrentUser uint16, bGangFlower bool) bool {

	if sp.IsPausing() {
		sp.CurrentUser = static.INVALID_CHAIR
		sp.SetSendCardOpt(static.TagSendCardInfo{
			CurrentUser: wCurrentUser,
			GangFlower:  bGangFlower,
		})
		return true
	}

	if !sp.SendStatus {
		return false
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
	//if sp.GetValidCount() <= 0 {
	if sp.GetValidCount() <= sp.getValidNum() {
		return false
	}

	bEnjoinHu := true
	//发牌处理
	sp.SendOneCard(_userItem)
	_userItem.Ctx.OutCardData = static.INVALID_BYTE
	//sp.SendCardCount++
	//sp.LeftCardCount--
	//sp.SendCardData = sp.RepertoryCard[sp.LeftCardCount]
	//_userItem.Ctx.DispatchCard(sp.SendCardData)
	sp.SetLeftCardArray()

	//if bGangFlower{
	//	sp.m_GangBuCardCount++
	//}
	//游戏记录
	recordStr := fmt.Sprintf("牌型%s，发来：%s", sp.Logic.SwitchToCardNameByIndexs(_userItem.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(sp.SendCardData, 1))
	sp.OnWriteGameRecord(wCurrentUser, recordStr)
	//记录发牌
	sp.addReplayOrder(wCurrentUser, info2.E_SendCard, sp.SendCardData)
	//设置变量
	sp.ProvideUser = wCurrentUser
	sp.ProvideCard = sp.SendCardData
	//给用户发牌后，判断用户是否可以杠牌
	//if sp.GetValidCount() > 0 {
	if sp.GetValidCount() > sp.getValidNum() {
		var GangCardResult static.TagGangCardResult
		a, b := sp.Logic.AnalyseGangCard(_userItem, _userItem.Ctx.CardIndex, _userItem.Ctx.WeaveItemArray[:], _userItem.Ctx.WeaveItemCount, &GangCardResult)
		_userItem.Ctx.UserAction |= a
		if b {
			sp.SendGameNotificationMessage(_userItem.GetChairID(), "弃杠后不能再杠")
		}
	}

	// 判断是否胡牌
	sp.initChiHuResult()
	var isGSP bool
	if sp.getHuRes(_userItem, int(sp.SendCardData), true, false, false, false) {
		xlog.Logger().Print("是否可以胡（发牌）", _userItem.Ctx.ChiHuResult.ChiHuKind2)
		if sp.GangFlower && sp.Rule.IsAutoQGH == 0 {
			isGSP = true
		}
	} else {
		if sp.GangFlower {
			sp.GangFlower = false
			sp.GangHotStatus = true
		}
	}
	//设置变量
	sp.OutCardData = 0
	sp.CurrentUser = wCurrentUser
	sp.OutCardUser = static.INVALID_CHAIR
	//构造数据
	var SendCard static.Msg_S_SendCard
	SendCard.CurrentUser = wCurrentUser
	SendCard.ActionMask = _userItem.Ctx.UserAction
	SendCard.CardData = 0x00
	if sp.SendStatus {
		SendCard.CardData = sp.SendCardData
	}

	//如果不是杠开发的牌，那打出去的时候一定不是热冲，这里无条件重置一下
	if !bGangFlower {
		sp.GangHotStatus = false
	}

	SendCard.IsGang = bGangFlower
	SendCard.IsHD = false
	SendCard.EnjoinHu = bEnjoinHu
	//SendCard.Overtime = sp.SetLimitTime()
	sp.LastSendCardUser = wCurrentUser
	// 设置开始超时操作
	sp.LockTimeOut(wCurrentUser, sp.Rule.Overtime_trust)
	for _, v := range sp.PlayerInfo {
		if v.GetChairID() != wCurrentUser {
			SendCard.CardData = 0x00
			SendCard.Overtime = sp.SetLimitTime()
			SendCard.VecGangCard = make([]int, 0)
		} else {
			SendCard.CardData = sp.SendCardData
			SendCard.Overtime = v.Ctx.CheckTimeOut
			SendCard.VecGangCard = static.HF_BytesToInts(_userItem.Ctx.VecGangCard)
			// sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
		}

		sp.SendPersonMsg(consts.MsgTypeGameSendCard, SendCard, uint16(v.GetChairID()))
	}
	//游戏记录
	//发送旁观数据
	SendCard.CardData = 0x00
	SendCard.VecGangCard = make([]int, 0)
	SendCard.ActionMask = 0
	sp.SendTableLookonMsg(consts.MsgTypeGameSendCard, SendCard)

	// 回放记录中记录牌权显示
	if _userItem.Ctx.UserAction > 0 {
		sp.addReplayOrder(wCurrentUser, info2.E_SendCardRight, _userItem.Ctx.UserAction)
		sp.OnWriteGameRecord(wCurrentUser, fmt.Sprintf("发送牌权：%x", _userItem.Ctx.UserAction))
		if isGSP {
			//_msg := sp.Greate_Operatemsg(_userItem.Uid, true, static.WIK_CHI_HU, sp.ProvideCard)
			//if !sp.OnUserOperateCard(_msg, false) {
			//	sp.OnWriteGameRecord(_userItem.GetChairID(), "自动杠上炮")
			//}
		}
	}
	return true
}

// ! 清除吃胡记录
func (sp *SportSSAH) initChiHuResult() {
	for _, v := range sp.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// ! 响应判断
func (sp *SportSSAH) EstimateUserRespond(wCenterUser uint16, cbCenterCard byte, EstimatKind int) (bool, map[int64]bool) {
	//变量定义
	isAutoHot := make(map[int64]bool)
	bAroseAction := false
	// 响应判断只需要判断出牌以及续杠
	if EstimatKind != static.EstimatKind_OutCard && EstimatKind != static.EstimatKind_GangCard {
		return bAroseAction, isAutoHot
	}
	//用户状态
	for _, v := range sp.PlayerInfo {
		if v == nil {
			continue
		}
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
			//吃碰判断
			//if sp.GetValidCount() > 0 {
			if sp.GetValidCount() > sp.getValidNum() {
				if cbCenterCard != sp.MagicCard {
					//碰牌判断
					_item.Ctx.UserAction |= sp.Logic.EstimatePengCard(_item.Ctx.CardIndex, cbCenterCard)
					if _item.Ctx.UserAction&static.WIK_PENG != 0 {
						for _, card := range _item.Ctx.VecPengCard {
							if card == cbCenterCard {
								_item.Ctx.UserAction ^= static.WIK_PENG
								sp.SendGameNotificationMessage(_item.GetChairID(), "过碰后不能再碰")
							}
						}
					}
					//吃牌判断
					//wEatUser := sp.GetNextFullSeat(wCenterUser)
					//if wEatUser == uint16(i) {
					//	_item.Ctx.UserAction |= sp.Logic.EstimateEatCard(_item.Ctx.CardIndex, cbCenterCard)
					//}

					//杠牌判断
					_item.Ctx.UserAction |= sp.Logic.EstimateGangCard(_item.Ctx.CardIndex, cbCenterCard)
				}
			}
		}

		//胡牌判断
		if cbCenterCard != sp.MagicCard { //&& !common.Findcard(sp.Logic.PiZiCards, cbCenterCard)
			if cbCenterCard == _item.Ctx.OutCardData {
				sp.OnWriteGameRecord(_item.GetChairID(),
					fmt.Sprintf("本轮已打出%s, 不用判胡，需要过庄。",
						sp.Logic.SwitchToCardNameByData(cbCenterCard, 1)))
				// sp.SendGameNotificationMessage(_item.GetChairID(), "本轮打出的牌过手前不能再胡")
			} else {
				sp.getHuRes(_item, int(cbCenterCard), false, false, false, false)
				// xlog.Logger().Print("是否可以胡（响应判断）", _item.Ctx.ChiHuResult.ChiHuKind2)
				if _item.Ctx.UserAction&static.WIK_CHI_HU != 0 {
					if len(_item.Ctx.VecChiHuCard) > 0 {
						_item.Ctx.UserAction ^= static.WIK_CHI_HU
						sp.SendGameNotificationMessage(_item.GetChairID(), "过胡后不能再胡")
					} else {
						if sp.GangHotStatus && sp.Rule.IsAutoQGH == 0 {
							isAutoHot[_item.Uid] = true
						}
					}
				}
			}
		}

		//结果判断
		if _item.Ctx.UserAction != static.WIK_NULL {
			bAroseAction = true
			// 开始用户操作
			sp.LockTimeOut(uint16(i), sp.Rule.Overtime_trust)
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
		sp.SendOperate()
		return true, isAutoHot
	}
	return false, isAutoHot
}

func (sp *SportSSAH) CheckNeedGuo(_userItem *components2.Player, cbCheckCard byte, level int, kind int) bool {
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

// ! 用户操作牌
func (sp *SportSSAH) OnUserOperateCard(msg *static.Msg_C_OperateCard, system bool) bool {
	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		return false
	}
	userItem := sp.GetUserItemByUid(msg.Id)
	//效验
	if userItem == nil {
		return false
	}
	if (userItem.Seat != sp.CurrentUser) && (sp.CurrentUser != static.INVALID_CHAIR) {
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("不是可以操作的玩家:%d,%d", userItem.Seat, sp.CurrentUser))
		return false
	}

	if !system { // TODO 过胡过碰 非系统二次调用 判断过户
		//能胡牌没有胡需要过庄
		if (userItem.Ctx.UserAction&static.WIK_CHI_HU) != 0 && msg.Code != static.WIK_CHI_HU && sp.CurrentUser == static.INVALID_CHAIR {
			userItem.Ctx.SetChiHuKind(userItem.Ctx.ChiHuResult.ChiHuKind, sp.ProvideCard, 0) //由于小胡过大胡 也需要过庄，那么第三个参数就传0 不做比较
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("有胡不胡，加入弃胡,牌:%s", sp.Logic.SwitchToCardNameByData(sp.ProvideCard, 1)))
		}

		//能碰不碰加入过碰
		if (userItem.Ctx.UserAction&static.WIK_PENG) != 0 && msg.Code != static.WIK_PENG && sp.CurrentUser == static.INVALID_CHAIR {
			userItem.Ctx.VecPengCard = append(userItem.Ctx.VecPengCard, sp.ProvideCard)
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("能碰不碰，加入过碰,牌:%s", sp.Logic.SwitchToCardNameByData(sp.ProvideCard, 1)))
		}

		//能杠不杠加入过杠
		if (userItem.Ctx.UserAction&static.WIK_GANG) != 0 && msg.Code != static.WIK_GANG {
			userItem.Ctx.VecGangCard = append(userItem.Ctx.VecGangCard, sp.ProvideCard)
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("能杠不杠,加入过杠,牌:%s", sp.Logic.SwitchToCardNameByData(sp.ProvideCard, 1)))
		}
	}

	//if msg.Code != public.WIK_CHI_HU{
	if err := sp.CheckOperationByRank(msg, userItem, system); err != nil {
		_, ok := err.(ErrorOptWaiting)
		if ok {
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("操作需等待：%x, %s", msg.Code, err.Error()))
		}
		return ok
	}
	//}

	if handler, ok := sp.OptHandlerHash[msg.Code]; ok {
		if err := handler(msg, userItem, system); err != nil {
			sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("操作错误：code: %x  err: %s", msg.Code, err.Error()))
			return false
		} else {
			// 回放中记录牌权操作
			sp.addReplayOrder(userItem.Seat, info2.E_HandleCardRight, msg.Code)

			//TODO 托管
			if sp.Rule.Overtime_trust != -1 && !msg.ByClient && !system {
				if !userItem.CheckTRUST() {
					var msg = &static.Msg_S_DG_Trustee{
						ChairID: userItem.Seat,
						Trustee: true,
					}
					if sp.onUserTustee(msg) {
						sp.OnWriteGameRecord(userItem.Seat, "操作超时进入托管")
					}
				}
			}
			return true
		}
	} else {
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("无效的用户动作：%x", msg.Code))
		return false
	}

}

// 判断是否为玩家主动操作
func (sp *SportSSAH) IsActiveOperation(wChairID uint16) bool {
	return sp.CurrentUser == wChairID
}

// 弃
func (sp *SportSSAH) OnUserQ(msg *static.Msg_C_OperateCard, user *components2.Player, system bool) error {
	isUserHaveH := false //因为一炮多响需要玩家自己操作胡 如果有人有点了胡 那么其他可以胡的玩家点非胡操作 会走这里
	for _, player := range sp.PlayerInfo {
		if player.Ctx.UserAction&static.WIK_CHI_HU != 0 && player.Ctx.Response {
			isUserHaveH = true
			break
		}
	}
	chairId := user.GetChairID()
	if user.Ctx.UserAction == static.WIK_NULL {
		return errors.New("没牌权")
	}
	if sp.IsActiveOperation(chairId) {
		sp.OnWriteGameRecord(chairId, "主动操作 弃牌")
		sp.ClearOutAllOperations()
		return nil
	}
	// 检查是否操作需要等待
	//if err := sp.CheckOperationByRank(msg, user, system); err != nil {
	//	return err
	//}

	cbTargetAction, cbTargetCard, _ := sp.GetUserRespondedAction(user)

	if cbTargetAction != static.WIK_NULL && !isUserHaveH {
		return errors.New("杠牌操作越权")
	}
	if sp.GangHotStatus {
		sp.GangHotStatus = false
	}
	sp.KzhDatas[int(chairId)] = 0
	beforeAction := user.Ctx.UserAction32
	user.Ctx.ClearOperateCard()
	user.Ctx.InitChiHuResult()
	if user = sp.GetUserItemByChair(sp.GetMaxResponsedOperationAfterGiveup()); user != nil && user.Ctx.PerformAction != static.WIK_NULL {
		cbTargetAction = user.Ctx.PerformAction
	} else {
		if beforeAction&static.WIK_K3Z != 0 {
			// 走到这里时，已经满足一下条件：
			// 1 已经不存在其他可以卡字胡的玩家没有操作
			// 2 已经不存在其他玩家响应任何牌权
			resumeUser := sp.GetUserItemByChair(sp.ResumeUser)
			if resumeUser == nil {
				return fmt.Errorf("这里有问题1")
			}
			if resumeUser.Ctx.UserAction32&static.WIK_PENG != 0 {
				// 判断用户碰后还能不能杠
				var gangCardResult static.TagGangCardResult
				//EventGiveUp := sp.Logic.AnalyseGangCard(user, user.Ctx.CardIndex, nil, 0, &gangCardResult)

				a, b := sp.Logic.AnalyseGangCard(resumeUser, resumeUser.Ctx.CardIndex, nil, 0, &gangCardResult)
				if b { //开局必然进不来
					sp.SendGameNotificationMessage(resumeUser.GetChairID(), "弃杠后不能再杠")
				}
				//if (a&public.WIK_GANG > 0) && sp.GetValidCount() > 0 {
				if (a&static.WIK_GANG > 0) && (sp.GetValidCount() > sp.getValidNum()) {
					//sp.SendGameNotificationMessage(chairId, constant.MsgContentGameGiveUpGang)
					resumeUser.Ctx.UserAction |= a
					sp.ProvideCard = 0
					sp.CurrentUser = resumeUser.Seat //dqjs-4498 碰后能杠认为是主动杠
					sp.SendOperate()
				}
				//else {
				// todo 通知客户端可以出牌
				sp.CurrentUser = resumeUser.Seat
				resumeUser.Ctx.UserAction32 = 0
				sp.SendPersonMsg(consts.MsgTypeGameIsOutCard, &static.Msg_S_IsOutCard{
					Seat:      resumeUser.Seat,
					IsOutCard: true,
				}, resumeUser.Seat)
				//}
			} else if resumeUser.Ctx.UserAction32&static.WIK_GANG != 0 {
				sp.CurrentUser = resumeUser.Seat
				sp.OnEventUserDrawCard(sp.ResumeUser, false)
			}
		} else {
			// 如果没有其他玩家有操作，则继续还原用户摸牌
			sp.OnEventUserDrawCard(sp.ResumeUser, false)
		}
		return nil
	}

	// 变量赋值
	cbTargetCard = user.Ctx.OperateCard
	// 出牌变量
	sp.OutCardData = 0
	// 状态改变
	sp.SendStatus = true
	// 清除出牌玩家
	sp.OutCardUser = static.INVALID_CHAIR

	// 系统模拟用户操作
	{
		_msg := &static.Msg_C_OperateCard{
			Card:     cbTargetCard,
			Code:     cbTargetAction,
			Id:       user.Uid,
			ByClient: true,
		}
		// 进行再次操作
		sp.OnUserOperateCard(_msg, true)
	}
	return nil
}

// 吃
func (sp *SportSSAH) OnUserE(msg *static.Msg_C_OperateCard, user *components2.Player, system bool) error {
	chairId := user.GetChairID()
	if sp.IsActiveOperation(chairId) {
		return errors.New(" 吃牌必须是被动操作")
	}
	//if err := sp.CheckOperationByRank(msg, user, system); err != nil {
	//	return err
	//}
	return sp.onUserOperateE(user) //处理吃
}

// 碰
func (sp *SportSSAH) OnUserP(msg *static.Msg_C_OperateCard, user *components2.Player, system bool) error {
	for _, player := range sp.PlayerInfo {
		if user.Seat != player.Seat && player.Ctx.PerformAction&static.WIK_CHI_HU != 0 && player.Ctx.Response {
			sp.OnWriteGameRecord(user.GetChairID(), "已经有人点了胡，你还点碰，那就是弃胡")
			return sp.OnUserQ(msg, user, system)
		}
	}
	chairId := user.GetChairID()
	if sp.IsActiveOperation(chairId) {
		return errors.New("碰牌必须是被动操作")
	}
	cbTargetAction, cbTargetCard, _ := sp.GetUserRespondedAction(user)
	// 效验
	if cbTargetAction != static.WIK_PENG {
		return errors.New("碰牌操作越权")
	}
	if !sp.Logic.IsValidCard(cbTargetCard) {
		return errors.New("碰牌牌子不合法:" + sp.Logic.SwitchToCardNameByData(cbTargetCard, 1))
	}
	if sp.GangHotStatus {
		sp.GangHotStatus = false
	}

	// 玩家在碰的时候有杠的牌权但是没有杠
	if user.Ctx.UserAction&static.WIK_GANG != 0 {
		sp.Logic.AppendGiveUpGang(user, cbTargetCard)
		sp.OnWriteGameRecord(user.GetChairID(), fmt.Sprintf("有杠不杠 加入弃杠：%v", user.Ctx.VecGangCard))
	}
	// 过庄检测//TODO 疑问
	//sp.DetectionPassDealer()
	// 追加组合
	wIndex := int(user.Ctx.WeaveItemCount)
	user.Ctx.WeaveItemCount++
	_provideUser := sp.ProvideUser
	if sp.ProvideUser == static.INVALID_CHAIR {
		_provideUser = chairId
	}
	if providItem := sp.GetUserItemByChair(_provideUser); providItem != nil {
		providItem.Ctx.Requiredcard(cbTargetCard)
	}
	user.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, cbTargetAction, cbTargetCard)
	whoGivenPeng := sp.ProvideUser
	// 发送碰牌响应
	var msgOperateResult static.Msg_S_OperateResult
	msgOperateResult.OperateUser = chairId
	msgOperateResult.OperateCard = cbTargetCard
	msgOperateResult.OperateCode = cbTargetAction
	msgOperateResult.ProvideUser = sp.ProvideUser
	if msgOperateResult.ProvideUser == static.INVALID_CHAIR {
		return errors.New(fmt.Sprintf("碰牌，未找到供应玩家:%d", sp.ProvideUser))
	}
	//// 写游戏记录
	recordStr := fmt.Sprintf("%s，碰牌：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(cbTargetCard, 1))
	sp.OnWriteGameRecord(chairId, recordStr)
	sp.addReplayOrder(chairId, info2.E_Peng, cbTargetCard)
	// 删除poker
	user.Ctx.RemoveCards(&sp.Rule, []byte{cbTargetCard, cbTargetCard})
	// 发送响应
	sp.SendResponseByOperation(user, &msgOperateResult)
	//判断用户碰后 有没有卡三张可胡的人
	//允许抢胡的条件下才分析抢胡 卡三张
	isHaveQH := false
	for i := 0; i < sp.GetPlayerCount(); i++ {
		_item := sp.GetUserItemByChair(uint16(i))
		if _item != nil && int(user.GetChairID()) != i {
			if frontSeat := uint16(sp.GetFrontSeat(_item.Seat)); frontSeat != whoGivenPeng /*&& _item.Seat != whoGivenPeng*/ {
				if _item.Seat == whoGivenPeng || _item.Ctx.OutCardData == cbTargetCard {
					sp.OnWriteGameRecord(_item.GetChairID(),
						fmt.Sprintf("本轮打出 %s, 无需判胡，需要过庄。",
							sp.Logic.SwitchToCardNameByData(cbTargetCard, 1),
						))
				} else {
					cards := static.HF_BytesToInts(_item.Ctx.CardIndex[:])
					index := sp.Logic.SwitchToCardIndex(cbTargetCard)
					tmpCards := make([]int, len(cards))
					copy(tmpCards, cards)
					tmpCards[index] += 1
					if sp.KzhDatas[int(_item.Seat)] == 1 && ssahCheckHu.CheckHuTest(tmpCards, 0) {
						_item.Ctx.UserAction |= static.WIK_CHI_HU
						_item.Ctx.UserAction32 |= static.WIK_K3Z
						_item.Ctx.ChiHuResult.ChiHuKind2 |= ssahCheckHu.GameHuPiH
						isHaveQH = true
						sp.SendOperate()
						//sp.ProvideUser = user.GetChairID()
						// sp.ProvideCard = cbTargetCar
						sp.OnWriteGameRecord(_item.GetChairID(), "碰后有玩家可抢胡")
					}

					//if sp.getHuRes(_item, int(cbTargetCard), false, true , false,false) { //可以胡
					//	isHaveQH = true
					//	sp.SendOperate()
					//	//sp.ProvideUser = user.GetChairID()
					//	// sp.ProvideCard = cbTargetCar
					//	sp.OnWriteGameRecord(_item.GetChairID(), "点杠后有玩家可抢胡")
					//}
				}

			} else {
				sp.OnWriteGameRecord(_item.GetChairID(),
					fmt.Sprintf("碰后不能判断胡，座位号(%d), 上家座位号(%d), 点碰座位号(%d)",
						_item.GetChairID(), frontSeat, whoGivenPeng,
					))
			}
		}
	}
	if isHaveQH {
		user.Ctx.UserAction32 |= static.WIK_PENG
		sp.ResumeUser = user.GetChairID()
		sp.CurrentUser = static.INVALID_CHAIR
		// TODO 通知客户端不能出牌
		sp.SendPersonMsg(consts.MsgTypeGameIsOutCard, &static.Msg_S_IsOutCard{
			Seat:      user.Seat,
			IsOutCard: false,
		}, user.Seat)
		return nil
	}

	// 改变运行数据
	sp.OnGameDataChanged(user, info2.E_Peng)
	// 判断用户碰后还能不能杠
	var gangCardResult static.TagGangCardResult
	//EventGiveUp := sp.Logic.AnalyseGangCard(user, user.Ctx.CardIndex, nil, 0, &gangCardResult)

	a, b := sp.Logic.AnalyseGangCard(user, user.Ctx.CardIndex, nil, 0, &gangCardResult)
	if b { //开局必然进不来
		sp.SendGameNotificationMessage(user.GetChairID(), "弃杠后不能再杠")
	}
	//if (a&public.WIK_GANG > 0) && sp.GetValidCount() > 0 {
	if (a&static.WIK_GANG > 0) && (sp.GetValidCount() > sp.getValidNum()) {
		//sp.SendGameNotificationMessage(chairId, constant.MsgContentGameGiveUpGang)
		user.Ctx.UserAction |= a
		sp.ProvideCard = 0
		sp.SendOperate()
	}
	return nil
}

// 杠
func (sp *SportSSAH) OnUserG(msg *static.Msg_C_OperateCard, user *components2.Player, system bool) error {
	for _, player := range sp.PlayerInfo {
		//20210420 苏大强 没看到能吃，就直接这样吧
		if user.Seat != player.Seat && player.Ctx.UserAction&static.WIK_CHI_HU != 0 && player.Ctx.Response && player.Ctx.PerformAction != static.CHK_NULL {
			sp.OnWriteGameRecord(user.GetChairID(), "已经有人点了胡，你还点杠，那就是弃胡")
			sp.OnUserQ(msg, user, system)
			return nil
		}
	}
	// 防止玩家杠不能杠的弃杠牌
	cbTargetAction, cbTargetCard, _ := sp.GetUserRespondedAction(user)
	// 效验
	if cbTargetAction != static.WIK_GANG {
		return errors.New("杠牌操作越权")
	}

	if !sp.Logic.IsValidCard(cbTargetCard) && msg.Card != sp.ProvideCard {
		return errors.New("杠牌牌子不合法:" + sp.Logic.SwitchToCardNameByData(cbTargetCard, 1))
	}

	//if err := sp.CheckOperationByRank(msg, user, system); err != nil {
	//	return err
	//}

	sp.GangHotStatus = true
	if err := sp.OnUserOperateG(msg, user, system); err != nil {
		return err
	}
	return nil
}

// 胡
func (sp *SportSSAH) OnUserH(msg *static.Msg_C_OperateCard, user *components2.Player, system bool) error {
	msgCurUserChairID := sp.GetChairByUid(msg.Id)
	cbTargetAction, cbTargetCard, _ := sp.GetUserRespondedAction(user)
	if cbTargetAction != static.WIK_CHI_HU && cbTargetAction != static.WIK_QIANG {
		return errors.New("胡牌操作越权")
	}
	// 效验
	checkHuRight := func(player *components2.Player) bool {
		return player.Ctx.PerformAction&static.WIK_QIANG != 0 || player.Ctx.PerformAction&static.WIK_CHI_HU != 0 //|| player.Ctx.UserAction&public.WIK_QIANG != 0 || player.Ctx.UserAction&public.WIK_CHI_HU != 0
	}
	if !checkHuRight(user) {
		return errors.New("玩家实际上没有胡牌权限")
	}

	if sp.IsActiveOperation(msgCurUserChairID) {
		sp.ProvideCard = sp.SendCardData
		if sp.ProvideCard == 0 {
			return errors.New("自摸：系统发牌信息有误，完成胡牌失败")
		}
		sp.ProvideUser = user.GetChairID()
		//自摸不存在热冲
		sp.GangHotStatus = false
		// 游戏记录
		recordStr := fmt.Sprintf("手牌%s，自摸：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(sp.ProvideCard, 1))
		sp.OnWriteGameRecord(user.GetChairID(), recordStr)
		// 记录胡牌
		sp.addReplayOrder(user.GetChairID(), info2.E_Hu, sp.ProvideCard)
		// 结束信息
		sp.ChiHuCard = sp.ProvideCard

	} else {
		sp.ChiHuCard = cbTargetCard
		//sp.addReplayOrder(user.GetChairID(), eve.E_Hu, sp.ChiHuCard)
		//for _, huplayer := range sp.PlayerInfo {
		//	if huplayer == nil {
		//		continue
		//	}
		//	//if huplayer.GetChairID() == user.GetChairID() {
		//	//	continue
		//	//}
		//	if checkHuRight(huplayer) {
		//		sp.addReplayOrder(huplayer.GetChairID(), eve.E_HandleCardRight, public.WIK_CHI_HU)
		//		sp.addReplayOrder(huplayer.GetChairID(), eve.E_Hu, sp.ChiHuCard)
		//	} else {
		//		huplayer.Ctx.ClearOperateCard()
		//		huplayer.Ctx.InitChiHuResult()
		//	}
		//}
		huplayer := user
		if checkHuRight(huplayer) {
			sp.addReplayOrder(huplayer.GetChairID(), info2.E_HandleCardRight, static.WIK_CHI_HU)
			sp.addReplayOrder(huplayer.GetChairID(), info2.E_Hu, sp.ChiHuCard)
		}
		//else {
		//	huplayer.Ctx.ClearOperateCard()
		//	huplayer.Ctx.InitChiHuResult()
		//}

		for _, player := range sp.PlayerInfo {
			if player.Ctx.UserAction&static.WIK_CHI_HU != 0 && !player.Ctx.Response {
				return nil
			}
		}

		if providItem := sp.GetUserItemByChair(sp.ProvideUser); providItem != nil {
			sp.OnWriteGameRecord(providItem.GetChairID(), fmt.Sprintf("点炮，点炮牌[%s]", sp.Logic.SwitchToCardNameByData(cbTargetCard, 1)))
			providItem.Ctx.Requiredcard(cbTargetCard)
		}

		//找卡字胡的跟胡
		next := sp.GetNextFullSeat(user.GetChairID())
		for next != sp.ProvideUser && next != user.GetChairID() {
			_item := sp.GetUserItemByChair(next)
			if _item != nil {
				if cbTargetCard == _item.Ctx.OutCardData {
					sp.OnWriteGameRecord(_item.GetChairID(),
						fmt.Sprintf("本轮打出 %s, 无需判胡，需要过庄。",
							sp.Logic.SwitchToCardNameByData(cbTargetCard, 1),
						))
					next = sp.GetNextFullSeat(_item.GetChairID())
				} else {
					cards := static.HF_BytesToInts(_item.Ctx.CardIndex[:])
					index := sp.Logic.SwitchToCardIndex(cbTargetCard)
					tmpCards := make([]int, len(cards))
					copy(tmpCards, cards)
					tmpCards[index] += 1
					if sp.KzhDatas[int(_item.Seat)] == 1 && ssahCheckHu.CheckHuTest(tmpCards, 0) {
						//if sp.getHuRes(_item, int(cbTargetCard), false, true , true,false) { //可以胡
						next = sp.GetNextFullSeat(_item.GetChairID())
						_item.Ctx.PerformAction |= static.WIK_CHI_HU
						_item.Ctx.UserAction |= static.WIK_CHI_HU
						_item.Ctx.ChiHuResult.ChiHuKind2 |= ssahCheckHu.GameHuPiH
						sp.addReplayOrder(_item.GetChairID(), info2.E_Hu, sp.ChiHuCard)
						sp.OnWriteGameRecord(_item.GetChairID(), "跟胡")
					} else {
						break
					}
				}
			}
		}

	}

	// 结束游戏
	if !sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL) {
		return errors.New("胡牌，自然结束游戏异常...")
	}
	return nil
}

// 用户摸牌时间
func (sp *SportSSAH) OnEventUserDrawCard(wChairID uint16, bGangFlower bool) {
	if sp.GetValidCount() > sp.getValidNum() {
		// 有牌发牌
		sp.ClearOutAllOperations()
		sp.DispatchCardData(wChairID, bGangFlower)
	} else {
		sp.OnWriteGameRecord(static.INVALID_CHAIR, "游戏黄庄")
		sp.HaveHuangZhuang = true
		sp.ChiHuCard = 0
		sp.ProvideUser = static.INVALID_CHAIR
		sp.ClearOutAllOperations()
		sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
	}
}

// ! 操作吃
func (sp *SportSSAH) onUserOperateE(user *components2.Player) error {
	optAction, optCard := user.Ctx.PerformAction, user.Ctx.OperateCard
	var cbRemoveCard []byte
	var wik_kind int
	//变量定义
	switch optAction {
	case static.WIK_LEFT: //上牌操作
		cbRemoveCard = []byte{optCard + 1, optCard + 2}
		wik_kind = info2.E_Wik_Left
		//游戏记录
		recordStr := fmt.Sprintf("%s，左吃牌：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(optCard, 1))
		sp.OnWriteGameRecord(user.Seat, recordStr)
	case static.WIK_RIGHT:
		cbRemoveCard = []byte{optCard - 2, optCard - 1}
		wik_kind = info2.E_Wik_Right
		//游戏记录
		recordStr := fmt.Sprintf("%s，右吃牌：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(optCard, 1))
		sp.OnWriteGameRecord(user.Seat, recordStr)
	case static.WIK_CENTER:
		cbRemoveCard = []byte{optCard - 1, optCard + 1}
		wik_kind = info2.E_Wik_Center
		//游戏记录
		recordStr := fmt.Sprintf("%s，中吃牌：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(optCard, 1))
		sp.OnWriteGameRecord(user.Seat, recordStr)
	default:
	}

	wIndex := int(user.Ctx.WeaveItemCount)
	user.Ctx.WeaveItemCount++
	_provideUser := sp.ProvideUser
	if sp.ProvideUser == static.INVALID_CHAIR {
		_provideUser = user.GetChairID()
	}
	user.Ctx.AddWeaveItemArray(wIndex, 1, _provideUser, optAction, optCard)
	// 发送碰牌响应
	var msgOperateResult static.Msg_S_OperateResult
	msgOperateResult.OperateUser = user.GetChairID()
	msgOperateResult.OperateCard = optCard
	msgOperateResult.OperateCode = optAction
	msgOperateResult.ProvideUser = sp.ProvideUser
	if msgOperateResult.ProvideUser == static.INVALID_CHAIR {
		return errors.New(fmt.Sprintf("碰牌，未找到供应玩家:%d", sp.ProvideUser))
	}
	//记录吃
	sp.addReplayOrder(user.Seat, wik_kind, optCard)
	//删除扑克
	user.Ctx.RemoveCards(&sp.Rule, cbRemoveCard)
	// 发送响应
	sp.SendResponseByOperation(user, &msgOperateResult)
	// 改变运行数据
	sp.OnGameDataChanged(user, info2.E_Peng)
	// 判断用户碰后还能不能杠
	var gangCardResult static.TagGangCardResult
	//EventGiveUp := sp.Logic.AnalyseGangCard(user, user.Ctx.CardIndex, nil, 0, &gangCardResult)

	a, b := sp.Logic.AnalyseGangCard(user, user.Ctx.CardIndex, nil, 0, &gangCardResult)
	if b { //开局必然进不来
		sp.SendGameNotificationMessage(user.GetChairID(), "弃杠后不能再杠")
	}

	//if (a&public.WIK_GANG > 0) && sp.GetValidCount() > 0 {
	if (a&static.WIK_GANG > 0) && (sp.GetValidCount() > sp.getValidNum()) {
		//sp.SendGameNotificationMessage(chairId, constant.MsgContentGameGiveUpGang)
		user.Ctx.UserAction |= a
		sp.ProvideCard = 0
		sp.SendOperate()
	}

	sp.GangHotStatus = false
	return nil
}

// ! 操作杠
func (sp *SportSSAH) OnUserOperateG(msg *static.Msg_C_OperateCard, user *components2.Player, system bool) error {
	msgCurUserChairID := sp.GetChairByUid(msg.Id)
	optCode := msg.Code
	optCard := msg.Card
	cbWeaveIndex := 0xFF
	isQGH := false
	isQh := 0 //0不允许抢杠 1都可以抢杠  2有条件枪杠
	//卡三张时用
	GCardIndex := sp.Logic.SwitchToCardIndex(msg.Card) // 杠牌的索引
	if sp.IsActiveOperation(msgCurUserChairID) && user.Ctx.CardIndex[GCardIndex] != 4 {
		if sp.Logic.IsGiveUpGang(user, msg.Card) {
			return errors.New("杠牌牌子为齐杠牌:" + sp.Logic.SwitchToCardNameByData(optCard, 1))
		}
	}
	var OperateResult static.Msg_S_OperateResult //构建消息体
	//设置变量
	sp.SendStatus = true
	user.Ctx.UserAction = static.WIK_NULL
	user.Ctx.PerformAction = static.WIK_NULL
	autoStr := "服务端自动"
	if msg.ByClient {
		autoStr = "玩家主动"
	}
	//是否可以抢杠胡
	if sp.IsActiveOperation(msgCurUserChairID) { //主动
		if user.Ctx.CardIndex[GCardIndex] == 4 { // 暗杠

			//设置变量
			cbWeaveIndex = int(user.Ctx.WeaveItemCount)
			user.Ctx.WeaveItemCount++
			user.Ctx.AddWeaveItemArray(cbWeaveIndex, 0, msgCurUserChairID, msg.Code, msg.Card)
			user.Ctx.HidGangAction()
			sp.LastGangKind = 2
			sp.LastGangIndex = byte(cbWeaveIndex)
			//anGangScore := sp.GetScoreOnGang(gameserver.E_Gang_AnGang)
			//for i := 0; i < sp.GetPlayerCount(); i++ {
			//	_item := sp.GetUserItemByChair(uint16(i))
			//	if msgCurUserChairID != _item.GetChairID() {
			//		OperateResult.ScoreOffset[_item.GetChairID()] -= anGangScore
			//		OperateResult.ScoreOffset[msgCurUserChairID] += anGangScore
			//	}
			//}
			OperateResult.ScoreOffset = sp.GetScoreOnGang(user, nil, info2.E_Gang_AnGang)
			//游戏记录
			recordStr := fmt.Sprintf("%s，%s, 暗杠牌：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), autoStr, sp.Logic.SwitchToCardNameByData(msg.Card, 1))
			sp.OnWriteGameRecord(msgCurUserChairID, recordStr)
			//记录暗杠牌
			sp.addReplayOrder(msgCurUserChairID, info2.E_Gang_AnGang, msg.Card)
			//删除扑克
			user.Ctx.CleanCard(GCardIndex)
		} else if user.Ctx.CardIndex[GCardIndex] == 1 { // 续杠
			cbWeaveIndex := 0xFF
			for i := byte(0); i < user.Ctx.WeaveItemCount; i++ {
				cbWeaveKind := user.Ctx.WeaveItemArray[i].WeaveKind
				cbCenterCard := user.Ctx.WeaveItemArray[i].CenterCard
				if (cbCenterCard == msg.Card) && (cbWeaveKind == static.WIK_PENG) {
					cbWeaveIndex = int(i)
					break
				}
			}
			//效验动作
			if cbWeaveIndex == 0xFF {
				return errors.New("续杠效验动作失败")
			}

			sp.LastGangKind = 1
			sp.LastGangIndex = byte(cbWeaveIndex)
			isQGH = true //自己的续杠是可以抢杠胡的
			//user.Ctx.WeaveItemArray[cbWeaveIndex].WeaveKind = public.WIK_FILL

			OperateResult.ScoreOffset = sp.GetScoreOnGang(user, nil, info2.E_Gang_XuGand)
			// 记录明蓄牌次数
			user.Ctx.ShowGangAction()
			//游戏记录
			recordStr := fmt.Sprintf("%s，%s, 蓄杠牌：%s", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), autoStr, sp.Logic.SwitchToCardNameByData(msg.Card, 1))
			sp.OnWriteGameRecord(msgCurUserChairID, recordStr)
			//记录蓄杠牌
			sp.addReplayOrder(msgCurUserChairID, info2.E_Gang_XuGand, msg.Card)
			//组合扑克
			user.Ctx.AddWeaveItemArray(cbWeaveIndex, 1, msgCurUserChairID, static.WIK_FILL, msg.Card)
			//删除扑克
			user.Ctx.CleanCard(GCardIndex)

		} else {
			return errors.New("杠出问题了！！")
		}
	} else { //被动  只处理明杠 别人打给我杠的牌
		//组合扑克

		cbWeaveIndex := int(user.Ctx.WeaveItemCount)
		user.Ctx.WeaveItemCount++
		_provideUser := sp.ProvideUser
		if sp.ProvideUser == static.INVALID_CHAIR {
			_provideUser = msgCurUserChairID
		}
		if providItem := sp.GetUserItemByChair(_provideUser); providItem != nil {
			providItem.Ctx.Requiredcard(optCard)
		}
		//isQh = true dqjs-4532不能抢点杠
		sp.LastGangKind = 3
		sp.LastGangIndex = byte(cbWeaveIndex)
		OperateResult.ScoreOffset = sp.GetScoreOnGang(user, sp.GetUserItemByChair(_provideUser), info2.E_Gang)
		user.Ctx.AddWeaveItemArray(cbWeaveIndex, 1, _provideUser, optCode, optCard)
		// 明杠
		user.Ctx.ShowGangAction()
		recordStr := fmt.Sprintf("%s，明牌：%s,明杠后打开系统 杠上开花标识", sp.Logic.SwitchToCardNameByIndexs(user.Ctx.CardIndex[:], 1), sp.Logic.SwitchToCardNameByData(optCard, 1))
		sp.OnWriteGameRecord(user.GetChairID(), recordStr)
		sp.addReplayOrder(user.GetChairID(), info2.E_Gang, optCard)
		//删除扑克
		user.Ctx.CleanCard(GCardIndex)

		//卡上不卡下
		if 3 == sp.GetPlayerCount() {
			/*卡上不卡下：abc为出牌顺序，a玩家卡字胡牌型听牌时，b打出的牌a可以胡，但是因为卡三张规则不满足胡牌条件没给胡，
			如果这个时候c碰了这个字，导致卡三张条件满足了，这个时候需要给a胡牌权，这个时候a胡下来算b点炮。*/
			if uint16(sp.GetFrontSeat(sp.ProvideUser)) == user.GetChairID() {
				//自己的下家打的牌，自己杠了，其他人不允许抢杠胡，满足“卡上”条件
				xlog.Logger().Info("下家打牌上家杠了，满足卡上条件，不允许抢杠胡")
				isQh = 0
			} else {
				isQh = 1
			}
		} else if 4 == sp.GetPlayerCount() {
			/*
				(一)		a  b  c  d
						听  打  杠  听  ad都可以
						听  杠  打  听  ad都不可
						杠  打  听  听  cd都不可
						打  杠  听  听  cd都可

				（二）	杠  听  打  听  b可d不可
						打  听  杠  听  b不可d可
			*/

			//第一种情况 听牌和打牌的人中间没有间隔人，是紧挨着的
			if uint16(sp.GetFrontSeat(sp.ProvideUser)) == user.GetChairID() {
				//自己的下家打的牌，自己杠了，其他人不允许抢杠胡，满足“卡上”条件
				xlog.Logger().Info("下家打牌上家杠了，满足卡上条件，不允许抢杠胡")
				isQh = 0
			} else if uint16(sp.GetFrontSeat(user.GetChairID())) == sp.ProvideUser {
				xlog.Logger().Info("上家打牌下家杠了，满足不卡下条件，允许抢杠胡")
				isQh = 1
			} else {
				/*第二种情况了 打和杠不是连着的，中间还有其他人。这个时候部分人是允许抢杠，部分人不允许抢杠
				此处放行，在抢杠的地方再拦截谁可以抢杠，谁不能*/
				xlog.Logger().Info("打牌的人和杠牌的人中间有其他人，是不连续的，部分人可以抢杠，部分人不可以枪杆")
				isQh = 2
			}
		} else {
			//2人玩的话 不存在抢杠情况了，能胡直接点炮胡了
			isQh = 0
		}
	}

	if sp.IsActiveOperation(msgCurUserChairID) {
		OperateResult.ProvideUser = msgCurUserChairID
	} else {
		OperateResult.ProvideUser = sp.ProvideUser
		if sp.ProvideUser == static.INVALID_CHAIR {
			OperateResult.ProvideUser = msgCurUserChairID
		}
	}
	OperateResult.OperateUser = msgCurUserChairID
	OperateResult.OperateCode = msg.Code
	OperateResult.OperateCard = msg.Card
	sp.SendResponseByOperation(user, &OperateResult)

	if isQGH {
		//允许抢杠胡的条件下才分析抢杠胡
		isHaveQGH := false
		var isEnd bool
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			//杠的人过滤掉
			if _item != nil && int(user.GetChairID()) != i {
				if _item.Ctx.OutCardData == optCard {
					sp.OnWriteGameRecord(_item.GetChairID(),
						fmt.Sprintf("本轮打出 %s, 无需判胡，需要过庄。",
							sp.Logic.SwitchToCardNameByData(optCard, 1),
						))
				} else {
					if sp.getHuRes(_item, int(optCard), false, false, false, true) { //可以胡
						isHaveQGH = true
						if sp.Rule.IsAutoQGH == 0 {
							isEnd = true
							sp.ChiHuCard = optCard
							_item.Ctx.PerformAction |= static.WIK_CHI_HU
							sp.addReplayOrder(_item.GetChairID(), info2.E_Hu, sp.ChiHuCard)
							sp.OnWriteGameRecord(_item.GetChairID(), "自动抢杠胡")
						} else {
							xlog.Logger().Print("是否可以胡（抢杠）", _item.Ctx.ChiHuResult.ChiHuKind2)
							sp.SendOperate()
							sp.ProvideUser = user.GetChairID()
							sp.ProvideCard = optCard
							sp.ResumeUser = user.GetChairID()
							sp.CurrentUser = static.INVALID_CHAIR
							sp.OnGameDataChanged(user, info2.E_Qiang)
							sp.OnWriteGameRecord(user.GetChairID(), "回头杠(蓄杠)后有玩家可抢杠胡")
						}
						sp.LastGangScore = OperateResult.ScoreOffset
					}
				}
			}
		}
		if isEnd {
			if providItem := sp.GetUserItemByChair(sp.ProvideUser); providItem != nil {
				sp.OnWriteGameRecord(providItem.GetChairID(), fmt.Sprintf("点炮，点炮牌[%s]", sp.Logic.SwitchToCardNameByData(sp.ChiHuCard, 1)))
				providItem.Ctx.Requiredcard(sp.ChiHuCard)
			}
			sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
			return nil
		}
		if isHaveQGH {
			return nil
		}
	} else if 1 == isQh || 2 == isQh {
		//允许抢胡的条件下才分析抢胡 卡三张
		isHaveQH := false
		for i := 0; i < sp.GetPlayerCount(); i++ {
			_item := sp.GetUserItemByChair(uint16(i))
			//抢杠 过滤掉自己和点杠的人（有可能有人有碰胡的时候点了碰，然后打出一个牌其他人杠了，最后有分析他胡）
			if _item != nil && int(user.GetChairID()) != i && int(sp.ProvideUser) != i {
				bCheck := true
				if 1 == isQh {
					bCheck = true
				} else {
					// 杠  听  打  听  b可d不可
					// 打  听  杠  听  b不可d可
					//这种有间隔的 只有杠的人下家(直接下家，不是下家的下家)才能抢杠
					if uint16(sp.GetFrontSeat(_item.Seat)) == user.Seat {
						bCheck = true
					} else {
						bCheck = false
					}
				}

				if bCheck {
					if _item.Ctx.OutCardData == optCard {
						sp.OnWriteGameRecord(_item.GetChairID(),
							fmt.Sprintf("本轮打出 %s, 无需判胡，需要过庄。",
								sp.Logic.SwitchToCardNameByData(optCard, 1),
							))
					} else {
						cards := static.HF_BytesToInts(_item.Ctx.CardIndex[:])
						index := sp.Logic.SwitchToCardIndex(optCard)
						tmpCards := make([]int, len(cards))
						copy(tmpCards, cards)
						tmpCards[index] += 1
						if sp.KzhDatas[int(_item.Seat)] == 1 && ssahCheckHu.CheckHuTest(tmpCards, 0) {
							//if sp.getHuRes(_item, int(optCard), false, true , false,false) { //可以胡
							_item.Ctx.UserAction |= static.WIK_CHI_HU
							_item.Ctx.UserAction32 |= static.WIK_K3Z
							_item.Ctx.ChiHuResult.ChiHuKind2 |= ssahCheckHu.GameHuPiH
							//点杠的这种“卡上不卡下”的逻辑导致的胡不算抢杠胡(GameHuQH是一个抢胡标记),算点炮，杠的这个分数也不算了
							_item.Ctx.ChiHuResult.ChiHuKind2 |= ssahCheckHu.GameHuQH
							sp.GangHotStatus = false
							isHaveQH = true
							sp.SendOperate()
							sp.OnWriteGameRecord(_item.GetChairID(), "点杠后有玩家可抢胡")
							xlog.Logger().Info(fmt.Sprintf("玩家[%s]可以胡玩家[%s]抢杠胡", _item.Name, user.Name))
						}
					}
				}
			}
		}
		if isHaveQH {
			user.Ctx.UserAction32 |= static.WIK_GANG
			sp.ResumeUser = user.GetChairID()
			sp.CurrentUser = static.INVALID_CHAIR
			// TODO 通知客户端不能出牌
			sp.SendPersonMsg(consts.MsgTypeGameIsOutCard, &static.Msg_S_IsOutCard{
				Seat:      user.Seat,
				IsOutCard: false,
			}, user.Seat)
			return nil
		}
	}

	//if sp.GetValidCount() > 0 {
	if sp.GetValidCount() > sp.getValidNum() {
		sp.GangFlower = true
		sp.DispatchCardData(msgCurUserChairID, true) //TODO 发牌
	} else {
		sp.ChiHuCard = 0
		sp.ProvideUser = static.INVALID_CHAIR
		sp.OnEventGameEnd(sp.ProvideUser, static.GER_NORMAL)
	}

	return nil
}

// 得到有效得剩余牌数
func (sp *SportSSAH) GetValidCount() byte {
	return sp.LeftCardCount
}

// 游戏过程中，游戏用户发生变化
func (sp *SportSSAH) OnGameDataChanged(_userItem /*事件触发者*/ *components2.Player, changType /*事件类型*/ int) {
	switch changType {
	case info2.E_Peng: //碰牌带来游戏运行数据变化   吃和碰是相同
		{
			// 改变变量
			sp.OutCardData = 0
			// TODO
			sp.SendStatus = true
			sp.OutCardUser = static.INVALID_CHAIR
			sp.CurrentUser = _userItem.GetChairID()
			sp.ProvideCard = 0
			sp.ProvideUser = static.INVALID_CHAIR
		}
	case info2.E_Gang: // 明杠带来游戏运行数据发生变化
		{
			sp.OutCardData = 0
			sp.SendStatus = true
			sp.OutCardUser = static.INVALID_CHAIR
		}
	case info2.E_Qiang: // 有玩家可抢杠胡时
		{
			sp.OutCardUser = _userItem.GetChairID()
			sp.ResumeUser = _userItem.GetChairID()
			sp.OutCardData = static.INVALID_BYTE
		}
	case info2.E_OutCard:
		{
			// 置发牌状态
			sp.SendStatus = true
			// 变量赋值
			sp.OutCardUser = _userItem.GetChairID()
			sp.LastOutCardUser = _userItem.GetChairID()
			sp.HaveGangCard = false
			sp.GangFlower = false
			sp.MingGangStatus = false
		}
	case info2.E_SendCard:
		{
			// 摸牌状态下出牌数据为0
			sp.OutCardData = 0
			// 当前用户
			sp.CurrentUser = _userItem.GetChairID()
			// 无人出牌
			sp.OutCardUser = static.INVALID_CHAIR
		}
	}
}

// 用户发送操作所带来的响应
func (sp *SportSSAH) SendResponseByOperation(_userItem /*触发者*/ *components2.Player, operateResult /*消息结构体*/ *static.Msg_S_OperateResult) {
	// 最新数据发送给客户端
	operateResult.HaveGang[_userItem.GetChairID()] = _userItem.Ctx.HaveGang // 是否杠过
	operateResult.Overtime = sp.SetLimitTime()

	// 特殊处理  碰牌被碰走了 把出这张牌从出牌人的dis牌区 清除
	if operateResult.ProvideUser == sp.LastOutCardUser {
		if !sp.IsActiveOperation(_userItem.GetChairID()) {
			if provider := sp.GetUserItemByChair(sp.LastOutCardUser); provider != nil {
				provider.Ctx.Requiredcard(operateResult.OperateCard)
			}
		}
		sp.LastOutCardUser = static.INVALID_CHAIR
	}

	for i, l := 0, sp.GetPlayerCount(); i < l; i++ {
		_user := sp.GetUserItemByChair(uint16(i))
		if _user == nil {
			sp.OnWriteGameRecord(operateResult.OperateUser, "空指针...在操作分数同步时")
			continue
		}
		operateResult.LaiGangCount[i] = _user.Ctx.CurMagicOut
		// 记录分数 //由于吃碰杠胡都会走 然后杠 立即结算 所以 在需要的时候 给ScoreOffset 赋值  不需要则是0
		//_user.Ctx.GangScore += operateResult.ScoreOffset[i]
		_user.Ctx.StorageScore += operateResult.ScoreOffset[i]
	}
	//operateResult.GameScore, operateResult.GameVitamin = sp.OnSettle(operateResult.ScoreOffset, constant.EventSettleGaming)
	if operateResult.OperateCode&static.WIK_GANG > 0 { //抢杠胡 要还原杠分
		sp.LastGangScore = operateResult.ScoreOffset
	}
	sp.LockTimeOut(_userItem.Seat, sp.Rule.Overtime_trust)
	sp.SendTableMsg(consts.MsgTypeGameOperateResult, operateResult)
	sp.SendTableLookonMsg(consts.MsgTypeGameOperateResult, operateResult)
	// 操作响应后，认为所有玩家动作过期
	sp.SendCardData = static.INVALID_BYTE
	sp.ClearOutAllOperations()
	return
}

//// 游戏过程中，玩家的分数发生变化事件
//func (_this *SportSSAH) OnUserScoreOffset(_userItem *gameserver.Player, offset int) error {
//	_userItem.Ctx.StorageScore += offset
//	return nil
//}

// 重置数据
func (sp *SportSSAH) ClearOutAllOperations() {
	//self.OnWriteGameRecord(public.INVALID_CHAIR, "ClearOutAllOperations 清理掉所有玩家旧数据")
	sp.ClearAction()
	sp.ClearChiHuResult()
	sp.ClearOperateCard()
}

// 清除用户动作
func (sp *SportSSAH) ClearAction() {
	for _, v := range sp.PlayerInfo {
		v.Ctx.UserAction = 0
	}
}

// 清除吃胡记录
func (sp *SportSSAH) ClearChiHuResult() {
	for _, v := range sp.PlayerInfo {
		v.Ctx.InitChiHuResult()
	}
}

// 清除用户opcard
func (sp *SportSSAH) ClearOperateCard() {
	for _, v := range sp.PlayerInfo {
		v.Ctx.ClearOperateCard()
	}
}

// 弃牌之后得到最大牌权
func (sp *SportSSAH) GetMaxResponsedOperationAfterGiveup() (maxop uint16) {
	defer func() {
		sp.OnWriteGameRecord(maxop, fmt.Sprintf("当前弃牌后操作权限最大的玩家:%d", maxop))
	}()
	maxop = static.INVALID_CHAIR
	maxoplev := byte(static.WIK_NULL)
	maxoplist := make([]uint16, 0)
	for i, l := 0, sp.GetPlayerCount(); i < l; i++ {
		_userItem := sp.GetUserItemByChair(uint16(i))
		if _userItem == nil {
			continue
		}
		if !_userItem.Ctx.Response {
			continue
		}
		if sp.Logic.GetUserActionRank(_userItem.Ctx.PerformAction) > sp.Logic.GetUserActionRank(maxoplev) {
			maxop = _userItem.GetChairID()
			maxoplev = _userItem.Ctx.PerformAction
		}
	}
	if maxop != static.INVALID_CHAIR {
		maxoplist = append(maxoplist, maxop)
		sp.OnWriteGameRecord(maxop, "在玩家弃牌后 有最大的已响应操作")
		// 找到是不是有多个
		for i, l := 0, sp.GetPlayerCount(); i < l; i++ {
			_userItem := sp.GetUserItemByChair(uint16(i))
			if _userItem == nil {
				continue
			}
			if !_userItem.Ctx.Response {
				continue
			}
			if sp.Logic.GetUserActionRank(_userItem.Ctx.PerformAction) == sp.Logic.GetUserActionRank(maxoplev) {
				maxoplist = append(maxoplist, _userItem.GetChairID())
			}
		}
		// 如果有多个人，则按还原用户的逆时针顺序给牌权
		if sp.ResumeUser != static.INVALID_CHAIR {
			maxop = sp.GetNearestUser(sp.ResumeUser, maxoplist...)
		}
	}
	return
}

// 得到用户响应
func (sp *SportSSAH) GetUserRespondedAction(_userItem *components2.Player) ( /*respondedAction*/ byte /*respondedCard*/, byte /*userAction*/, byte) {
	return _userItem.Ctx.PerformAction, _userItem.Ctx.OperateCard, _userItem.Ctx.UserAction
}

// 根据动作优先级确定出操作玩家是否需要等待
func (sp *SportSSAH) CheckOperationByRank(msg *static.Msg_C_OperateCard, _userItem *components2.Player, system bool) error {
	// 变量定义
	wTargetUser := sp.GetChairByUid(msg.Id)
	cbOperateCode := msg.Code
	cbOperateCard := msg.Card
	// 操作合法性效验
	if err := sp.IsLegalOperation(_userItem, cbOperateCode, system); err != nil {
		sp.OnWriteGameRecord(_userItem.GetChairID(), err.Error())
		return err
	}
	// 被动情况下 玩家操作的牌不是系统提供的牌错误
	if !sp.IsActiveOperation(_userItem.GetChairID()) && sp.ProvideCard != cbOperateCard {
		return errors.New("被动操作的情况下，玩家操作的牌不是系统提供的牌错误")
	}
	// 变量定义
	cbTargetAction := cbOperateCode
	// 设置变量
	// 保存用户的本次操作
	_userItem.Ctx.SetOperate(cbOperateCard, cbOperateCode)
	if cbOperateCard == 0 {
		_userItem.Ctx.SetOperateCard(sp.ProvideCard)
	}
	//isPassive := false
	// 主动操作的情况下 不用效验rank
	if sp.IsActiveOperation(_userItem.GetChairID()) {
		//if !msg.ByClient {   //TODO 设置倒计时
		//	sp.setLimitedTime()
		//}
		sp.OnWriteGameRecord(_userItem.GetChairID(), "主动操作的情况下 不用效验rank")
		return nil
	} else {
		//isPassive = true
	}
	// 找到优先级最大的玩家
	for _, v := range sp.PlayerInfo {
		if v == nil {
			sp.OnWriteGameRecord(static.INVALID_CHAIR, "空指针：CheckOperationByRank")
			continue
		}
		if v.Seat == _userItem.Seat {
			continue
		}
		// 获取动作
		cbUserAction := v.Ctx.UserAction
		// 如果玩家应该响应了某的操作
		if v.Ctx.Response {
			cbUserAction = v.Ctx.PerformAction
		}
		// 得到动作优先级别
		cbUserActionRank := sp.Logic.GetUserActionRank(cbUserAction)
		cbTargetActionRank := sp.Logic.GetUserActionRank(cbTargetAction)
		//动作判断
		if cbUserActionRank > cbTargetActionRank /*|| (isPassive && (cbTargetAction&public.WIK_CHI_HU != 0) && (cbUserAction&public.WIK_CHI_HU != 0))*/ {
			wTargetUser = v.GetChairID()
			cbTargetAction = cbUserAction
		}
	}
	if _userItem = sp.GetUserItemByChair(wTargetUser); _userItem != nil && !_userItem.Ctx.Response {
		return ErrorOptWaiting("请等待更大牌权人响应")
	}
	return nil
}

// 判断玩家是否为合法操作
func (sp *SportSSAH) IsLegalOperation(_userItem *components2.Player, cbOperateCode byte, system bool) error {
	if _userItem == nil {
		return errors.New("空指针:CheckResponse")
	}
	// 非系统模拟的情况下玩家再次响应
	if _userItem.Ctx.Response && !system {
		return errors.New("The PlayerInfo is already responsive")
	}
	// 用户没有这个牌权
	if (cbOperateCode != static.WIK_NULL) && ((_userItem.Ctx.UserAction & cbOperateCode) == 0) {
		return errors.New("Response mismatch")
	}
	return nil
}

// 使客户端强制刷新
func (sp *SportSSAH) flashClient(uid int64, msg string) {
	chairId := sp.GetChairByUid(uid)
	sp.OnWriteGameRecord(chairId, msg)
	sp.SendGameNotificationMessage(chairId, msg)
	sp.DisconnectOnMisoperation(chairId)
}

/*
 向客户端发送消息的函数=======================================================
*/

// 广播玩家的状态和下跑的数目
func (sp *SportSSAH) BroadPao(chairID uint16) bool {
	var Pao static.Msg_S_Xiapao
	for _, v := range sp.PlayerInfo {
		xlog.Logger().Print("要广播的下跑数据：", v.Ctx.VecXiaPao.Num)
		if v.Seat != static.INVALID_CHAIR {
			Pao.Num[v.Seat] = v.Ctx.VecXiaPao.Num
			Pao.Status[v.Seat] = v.Ctx.UserPaoReady
			Pao.Always[v.Seat] = v.Ctx.VecXiaPao.Status //自动下炮
		}
	}
	//发送数据
	sp.SendTableMsg(consts.MsgTypeGameXuanpiao, &Pao)
	sp.SendTableLookonMsg(consts.MsgTypeGameXuanpiao, &Pao)
	//游戏记录
	recordStr := fmt.Sprintf("发送跑数：%d， 是否默认 %t", Pao.Num[chairID], Pao.Status[chairID])
	sp.OnWriteGameRecord(chairID, recordStr)
	sp.addReplayOrder(chairID, info2.E_Pao, byte(Pao.Num[chairID]))
	return true
}

// ! 发送操作
func (sp *SportSSAH) SendOperate() bool {
	//发送提示
	for _, v := range sp.PlayerInfo {
		if v.Ctx.UserAction != static.WIK_NULL {
			//构造数据
			var msgOperateNotify static.Msg_S_OperateNotify
			msgOperateNotify.ResumeUser = sp.ResumeUser
			//抢暗杠时，复用此字段，表示轮到谁抢了
			msgOperateNotify.ActionCard = sp.ProvideCard
			msgOperateNotify.ActionMask = v.Ctx.UserAction
			msgOperateNotify.EnjoinHu = false
			//msgOperateNotify.Overtime = sp.SetLimitTime()
			msgOperateNotify.Overtime = v.Ctx.CheckTimeOut
			msgOperateNotify.VecGangCard = static.HF_BytesToInts(v.Ctx.VecGangCard)
			//sp.LockTimeOut(v.Seat, sp.Rule.Overtime_trust)
			sp.SendPersonMsg(consts.MsgTypeGameOperateNotify, msgOperateNotify, v.Seat)
			// 游戏记录
			if v.Ctx.UserAction > 0 {
				recrodStr := fmt.Sprintf("发送牌权：%x", v.Ctx.UserAction)
				sp.OnWriteGameRecord(v.Seat, recrodStr)
				// 回放记录中记录牌权显示
				sp.addReplayOrder(v.Seat, info2.E_SendCardRight, v.Ctx.UserAction)
			}

		}
	}

	return true
}

/*
此结构必备的一些函数
*/
//! 增加回放操作记录
func (sp *SportSSAH) addReplayOrder(chairId uint16, operation int, card byte) {
	var order meta2.Replay_Order
	order.Chair_id = chairId
	order.Operation = operation
	order.Value = append(order.Value, int(card))
	sp.ReplayRecord.VecOrder = append(sp.ReplayRecord.VecOrder, order)
}

// ! 写日志记录
func (sp *SportSSAH) WriteGameRecord() {
	//写日志记录
	sp.OnWriteGameRecord(static.INVALID_CHAIR, "开始红中赖子杠  发牌......")
	// 玩家手牌
	for i := 0; i < len(sp.PlayerInfo); i++ {
		v := sp.GetUserItemByChair(uint16(i))
		if v != nil {
			handCardStr := fmt.Sprintf("发牌后手牌:%s", sp.Logic.SwitchToCardNameByIndexs(v.Ctx.CardIndex[:], 0))
			sp.OnWriteGameRecord(uint16(v.Seat), handCardStr)
		}
	}
	// 牌堆牌
	leftCardStr := fmt.Sprintf("牌堆牌:%s", sp.Logic.SwitchToCardNameByDatas(sp.RepertoryCard[0:sp.LeftCardCount+2], 0))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, leftCardStr)
	//赖子牌
	magicCardStr := fmt.Sprintf("癞子牌:%s", sp.Logic.SwitchToCardNameByData(sp.MagicCard, 1))
	sp.OnWriteGameRecord(static.INVALID_CHAIR, magicCardStr)
}

/*
9秒场游戏逻辑
*/
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// 创建操作消息
func (sp *SportSSAH) Greate_Operatemsg(Id int64, byClient bool, Code byte, Card byte) *static.Msg_C_OperateCard {
	_msg := new(static.Msg_C_OperateCard)
	_msg.Card = Card
	_msg.Code = Code
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// 创建出牌消息
func (sp *SportSSAH) Greate_OutCardmsg(Id int64, byClient bool, Card byte) *static.Msg_C_OutCard {
	_msg := new(static.Msg_C_OutCard)
	_msg.CardData = Card
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// 创建消息
func (sp *SportSSAH) Greate_XiaPaomsg(Id int64, byClient bool, status bool, isBanker bool) *static.Msg_C_Xiapao {
	_msg := new(static.Msg_C_Xiapao)
	_msg.Num = 0
	if isBanker {
		_msg.Num = 1
	}
	_msg.Status = status
	_msg.Id = Id
	_msg.ByClient = byClient
	return _msg
}

// ! 超时自动出牌
func (sp *SportSSAH) OnAutoOperate(wChairID uint16, bBreakin bool) {
	sp.OnWriteGameRecord(wChairID, "【进入超时自动操作】")
	if bBreakin == false {
		return
	}
	if sp.GetGameStatus() == static.GS_MJ_FREE {
		//self.UnLockTimeOut(wChairID)
		return
	}

	if sp.GetGameStatus() != static.GS_MJ_PLAY {
		//self.UnLockTimeOut(wChairID)
		return
	}
	_userItem := sp.GetUserItemByChair(wChairID)
	if _userItem == nil {
		return
	}

	////能胡 胡牌 吃胡
	//if (_userItem.Ctx.UserAction&public.WIK_CHI_HU) != 0 && sp.CurrentUser != wChairID {
	//	_msg := sp.Greate_Operatemsg(_userItem.Uid, false, public.WIK_NULL, sp.ProvideCard)
	//	sp.OnUserOperateCard(_msg,true)
	//	return
	//}

	//能胡 胡牌 吃胡
	if (_userItem.Ctx.UserAction&static.WIK_CHI_HU) != 0 && sp.CurrentUser != wChairID {
		var _msg *static.Msg_C_OperateCard
		if sp.m_bAutoHu[_userItem.Seat] {
			_msg = sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_CHI_HU, sp.ProvideCard)
		} else {
			_msg = sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_NULL, sp.ProvideCard)
		}
		if !sp.OnUserOperateCard(_msg, false) {
			sp.OnWriteGameRecord(wChairID, "服务器自动选弃的时候，可能被客户端抢先了")
			return
		}
		return
	}

	//点杠 点碰 放弃
	if sp.CurrentUser == static.INVALID_CHAIR && _userItem.Ctx.UserAction != 0 {
		_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_NULL, sp.ProvideCard)
		sp.OnUserOperateCard(_msg, false)
		return
	}

	//暗杠 擦炮直接放弃出牌
	if sp.CurrentUser == wChairID {
		if _userItem.Ctx.UserAction&static.WIK_CHI_HU != 0 && sp.m_bAutoHu[_userItem.Seat] {
			_msg := sp.Greate_Operatemsg(_userItem.Uid, false, static.WIK_CHI_HU, sp.ProvideCard)
			sp.OnWriteGameRecord(_userItem.Seat, "玩家可胡牌，并且选了超时自动胡牌, 胡牌")
			if !sp.OnUserOperateCard(_msg, false) {
				sp.OnWriteGameRecord(wChairID, "服务器自动操作的时候，可能被客户端抢先了")
			}
		} else {
			cbSendCardData := sp.SendCardData
			index := sp.Logic.SwitchToCardIndex(cbSendCardData) // 出牌索引
			if index >= 0 && index < static.MAX_INDEX {
				if 0 != _userItem.Ctx.CardIndex[index] {
					//如果有癞子 并且 癞子不能杠 并且 发的牌是癞子 那就什么都不做
					if sp.Rule.IsHaveMagic && sp.Rule.TypeLaizi != 3 && cbSendCardData == sp.MagicCard {
						sp.OnWriteGameRecord(wChairID, "有癞子并且癞子不能杠并且发的牌是癞子下面随机出一张")
					} else { //非上面的情况 都正常自动出牌
						_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
						if !sp.OnUserOutCard(_msg) {
							sp.OnWriteGameRecord(wChairID, "服务器自动出牌的时候，可能被客户端抢先了")
						} else {
							recordStr := fmt.Sprintf("服务端自动操作：出牌 ：%s", sp.Logic.SwitchToCardNameByData(cbSendCardData, 1))
							sp.OnWriteGameRecord(wChairID, recordStr)
						}
						return
					}
				}
			}
			for i := byte(static.MAX_INDEX - 1); i > 0; i-- {
				if _userItem.Ctx.CardIndex[i] != 0 {
					cbSendCardData := sp.Logic.SwitchToCardData(i)
					if sp.Rule.IsHaveMagic && sp.Rule.TypeLaizi != 3 { //癞子不能杠
						if cbSendCardData != sp.MagicCard {
							_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
							sp.OnUserOutCard(_msg)
							return
						}
						continue
					} else {
						_msg := sp.Greate_OutCardmsg(_userItem.Uid, false, cbSendCardData)
						sp.OnUserOutCard(_msg)
						return
					}

				}
			}
		}
	}

}

func (sp *SportSSAH) SetAutoNextTimer(leftTime int) {
	if int(sp.CurCompleteCount) >= sp.Rule.JuShu { //局数够了
		return
	}
	if leftTime <= 0 {
		leftTime = components2.GAME_OPERATION_TIME_AUTONEXT
	}
	for _, v := range sp.PlayerInfo {
		v.Ctx.Timer.SetTimer(components2.GameTime_AutoNext, leftTime)
	}
}

func (sp *SportSSAH) getCardsADDorDEL(fur byte) (a, b byte) {
	cbValue := byte(fur & static.MASK_VALUE)
	cbColor := byte(fur & static.MASK_COLOR)
	a = (cbValue + 1) | cbColor
	b = (cbValue - 1) | cbColor
	if cbValue == 9 {
		a = 1 | cbColor
	}
	if cbValue == 1 {
		b = 9 | cbColor
	}
	return
}

func (sp *SportSSAH) getValidNum() byte {
	validNum := 0
	if sp.Rule.MaType == 0 {
		validNum = sp.Rule.DiMa + 4
	}
	return byte(validNum)
}

// ! 发送游戏开始场景数据
func (sp *SportSSAH) sendGameSceneStatusPlayLookon(player *components2.Player) bool {
	if player.LookonTableId == 0 {
		return false
	}

	wChiarID := player.GetChairID()

	if int(wChiarID) >= sp.GetPlayerCount() {
		wChiarID = 0
	}
	//是否要获取wChiarID位置真正玩家的信息 ？
	playerOnChair := sp.GetUserItemByChair(wChiarID)

	//取消托管
	//player.Ctx.SetTrustee(false)
	var Trustee static.Msg_S_Trustee
	Trustee.Trustee = false
	Trustee.ChairID = wChiarID
	sp.SendTableMsg(consts.MsgTypeGameTrustee, Trustee)
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
	StatusPlay.PayPaostatus = sp.PayPaoStatus
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
		StatusPlay.PlayerFan[i] = sp.GetPlayerFan(uint16(i))
		//追加痞子out
		StatusPlay.PiGangCount[i] = _item.Ctx.HongZhongGang + _item.Ctx.FaCaiGang + _item.Ctx.PiZiGangCount + _item.Ctx.PiZiCardOut
		StatusPlay.LaiGangCount[i] = _item.Ctx.MagicCardGang + _item.Ctx.MagicCardOut

		StatusPlay.PaoNum[i] = byte(_item.Ctx.VecXiaPao.Num)
		StatusPlay.PaoStatus[i] = _item.Ctx.UserPaoReady == true
		StatusPlay.Whotrust[i] = _item.CheckTRUST()
		for j := 0; j < len(_item.Ctx.DiscardCard) && j < len(StatusPlay.DiscardCard[i]); j++ {
			StatusPlay.DiscardCard[i][j] = _item.Ctx.DiscardCard[j]
		}
		for j := 0; j < len(_item.Ctx.DiscardCardClass) && j < len(StatusPlay.DiscardCardClass[i]); j++ {
			StatusPlay.DiscardCardClass[i][j] = _item.Ctx.DiscardCardClass[j]
		}
		for j := 0; j < len(_item.Ctx.Pilaicardcard) && j < len(StatusPlay.Pilaicardcard[i]); j++ {
			StatusPlay.Pilaicardcard[i][j] = _item.Ctx.Pilaicardcard[j]
		}

		StatusPlay.DiscardCount[i] = byte(len(_item.Ctx.DiscardCard))
		//组合扑克
		StatusPlay.WeaveItemArray[i] = _item.Ctx.WeaveItemArray
		StatusPlay.WeaveCount[i] = _item.Ctx.WeaveItemCount
	}

	//状态变量
	StatusPlay.ActionCard = sp.ProvideCard
	StatusPlay.LeftCardCount = sp.LeftCardCount
	StatusPlay.ActionMask = 0

	StatusPlay.CardLeft.CardArray = make([]int, sp.RepertoryCardArray.MaxCount, sp.RepertoryCardArray.MaxCount)
	copy(StatusPlay.CardLeft.CardArray, sp.RepertoryCardArray.CardArray)
	StatusPlay.CardLeft.MaxCount = sp.RepertoryCardArray.MaxCount
	StatusPlay.CardLeft.Seat = int(sp.RepertoryCardArray.Seat)
	//StatusPlay.CardLeft.Kaikou = sp.RepertoryCardArray.Kaikou
	//StatusPlay.CardLeft.Random = sp.RepertoryCardArray.Random

	if playerOnChair != nil && playerOnChair.Ctx.Response {
		StatusPlay.ActionMask = static.WIK_NULL
	}
	StatusPlay.VecGangCard = static.HF_BytesToInts(player.Ctx.VecGangCard)
	if (StatusPlay.ActionMask & static.WIK_QIANG) != 0 {
		StatusPlay.ActionCard = byte(wChiarID)
	}

	if playerOnChair != nil && playerOnChair.Ctx.CheckTimeOut != 0 {
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
			if v.Ctx.UserAction > 0 {
				StatusPlay.Overtime = 0
				break
			}
		}
	}

	if sp.OutCardUser == static.INVALID_CHAIR {
		if (StatusPlay.ActionMask & static.WIK_GANG) != 0 {
			StatusPlay.ActionCard = 0 //自己杠过，还可以再杠，此时杠牌信息由前台分析。后台传的这个值不正确。
		}
	}

	//历史记录
	StatusPlay.OutCardUser = sp.OutCardUser
	StatusPlay.OutCardData = sp.OutCardData

	//扑克数据
	if playerOnChair != nil {
		StatusPlay.CardCount, StatusPlay.CardData = sp.Logic.SwitchToCardData2(player.Ctx.CardIndex, StatusPlay.CardData)
	}
	StatusPlay.IsOutCard = true
	if player.Ctx.UserAction32&static.WIK_PENG != 0 {
		StatusPlay.IsOutCard = false
	} else if player.Ctx.UserAction32&static.WIK_GANG != 0 {
		StatusPlay.IsOutCard = false
	}

	//发送场景
	sp.SendPersonLookonMsg(consts.MsgTypeGameStatusPlay, StatusPlay, player.Uid)
	//发小结消息
	if byte(len(sp.VecGameEnd)) == sp.CurCompleteCount && sp.CurCompleteCount != 0 && int(wChiarID) < sp.GetPlayerCount() && wChiarID >= 0 {
		//发消息
		gamend := sp.VecGameEnd[sp.CurCompleteCount-1]
		gamend.BirdCard = 1 //表示为断线重连

		sp.SendPersonLookonMsg(consts.MsgTypeGameEnd, gamend, player.Uid)
	}

	//发送解散房间所有玩家的反应
	sp.SendAllPlayerDissmissInfo(player)

	return true
}

func (sp *SportSSAH) FindWantCard(userItem *components2.Player) byte {
	index := static.INVALID_BYTE
	cardData := userItem.Ctx.WantCard
	if !sp.Logic.IsValidCard(cardData) {
		sp.OnWriteGameRecord(userItem.GetChairID(), fmt.Sprintf("不合法的WANT牌:%s。", sp.Logic.SwitchToCardNameByData(cardData, 1)))
		return index
	}
	return sp.FindCardIndexInRepertoryCards(cardData)
}
