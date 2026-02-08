//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package Hubei_JingZhou

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/services/sport/backboard"
	"github.com/open-source/game/chess.git/services/sport/components"
	"github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
)

//////////////////////////////////////////////////////////////////////////
const (
	//游戏 I D
	GAME_JZhzg_GENRE = (static.GAME_GENRE_GOLD | static.GAME_GENRE_MATCH) //游戏类型
)

//////////////////////////////////////////////////////////////////////////

//////////////////////////t///////////////////////////////////////////////

type SportLogicJZHZG struct {
	logic.BaseLogic
	cardClazz int    //牌型，用于生成牌库和牌的有效型判断
	m_cbCard  []byte //每大局的牌库可能不同,通山晃晃都一样
}

func (slh *SportLogicJZHZG) SetCardClass(special int) {
	//20200325 默认 无风
	slh.cardClazz = backboard.SetCardsClass_ex(backboard.CARDS_WITHOUT_WIND, special)
}
func (slh *SportLogicJZHZG) CreateCards() (err error) {
	//预备替换
	// fmt.Println(fmt.Sprintf("初始牌型class（%d)", slh.Cardsclass))
	err, slh.m_cbCard = backboard.CreateCards(slh.cardClazz)
	if err != nil {
		return err
	}
	//cardsIndex := backboard.CardsToCardIndex(slh.m_cbCard)
	// fmt.Println(fmt.Sprintf("创建牌库的列表数据"))
	//backboard.Print_cards(cardsIndex[:])
	return nil
}

//混乱扑克
func (slh *SportLogicJZHZG) RandCardData() (byte, []byte) {
	// fmt.Println(fmt.Sprintf("混乱牌型class（%d)", slh.Cardsclass))
	//if len(slh.m_cbCard) == 0 {
	err := slh.CreateCards()
	if err != nil {
		fmt.Println(fmt.Sprintf("重建牌库失败,%v", err))
	}
	//}
	//cardsIndex := backboard.CardsToCardIndex(slh.m_cbCard)
	//backboard.Print_cards(cardsIndex[:])
	return backboard.RandCardData(slh.m_cbCard)
}

//有效判断
func (slh *SportLogicJZHZG) IsValidCard(cbCardData byte) bool {
	cbValue := (cbCardData & static.MASK_VALUE)
	cbColor := (cbCardData & static.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	bIsValid := cbValue >= 1 && cbValue <= 9 && cbColor >= 0
	if !bIsValid {
		return false //非法牌
	}

	if slh.Rule.NoWan {
		if cbColor == 0 {
			return false
		}
	}
	return true
}

//吃牌判断
func (slh *SportLogicJZHZG) EstimateEatCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	return static.WIK_NULL
}

//杠牌分析
func (slh *SportLogicJZHZG) AnalyseGangCard(_userItem *components.Player, cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, GangCardResult *static.TagGangCardResult) byte {
	//设置变量
	cbActionMask := static.WIK_NULL

	//手上杠牌
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//过滤混子
		if i == slh.SwitchToCardIndex(slh.MagicCard) {
			continue
		}
		if cbCardIndex[i] == 4 {
			cbActionMask |= static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = slh.SwitchToCardData(i)
			GangCardResult.CardCount++
		}
	}

	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_PENG {
			if cbCardIndex[slh.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				cbActionMask |= static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = WeaveItem[i].CenterCard
				GangCardResult.CardCount++
			}
		}
	}

	return byte(cbActionMask)
}

//基本胡牌分析
func (slh *SportLogicJZHZG) AnalyseHuKind(wChiHuRight uint16, wChiHuKind uint64, cbTempCard []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) uint64 {
	//清一色：手中的牌和组合牌都是一种花色，乱将
	if slh.HuType.HAVE_QING_YISE_HU && slh.IsQingYiSe(cbTempCard, WeaveItem, cbWeaveCount) == true {
		wChiHuRight |= static.CHR_QING_YI_SE
	}
	//七对
	if cbWeaveCount == 0 && slh.Rule.QiDuiJiaBei {
		if slh.HuType.HAVE_QIDUI_HU {
			is7dui, _ := slh.IsHaoHuaQiDui(cbTempCard)
			if is7dui {
				wChiHuKind |= static.CHK_7_DUI
			}
		}
	}

	_, AnalyseItemArray := slh.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//牌型分析
		for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
			// 有清一色的牌权并且牌型可以胡
			if slh.HuType.HAVE_QING_YISE_HU && (static.CHR_QING_YI_SE&wChiHuRight) != 0 {
				wChiHuKind |= static.CHK_QING_YI_SE
			}

			wChiHuKind |= static.CHK_PING_HU_NOMAGIC
		}
	} else {
		if (wChiHuKind&static.CHK_7_DUI) != 0 || (wChiHuKind&static.CHK_FOUR_LAIZE) != 0 {
			wChiHuKind |= static.CHK_DA_HU_NOMAGIC
		}

		// 有清一色的牌权并且牌型可以胡
		if slh.HuType.HAVE_QING_YISE_HU && (static.CHR_QING_YI_SE&wChiHuRight) != 0 {
			if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
				wChiHuKind |= static.CHK_QING_YI_SE
			}
		}
	}

	// 杠上开花判断
	if slh.HuType.HAVE_GANG_SHANG_KAI_HUA && (static.CHR_GANG_SHANG_KAI_HUA&wChiHuRight) != 0 {
		if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_GANG_SHANG_KAI_HUA
		}
	}
	// 杠上开花判断
	if slh.HuType.HAVE_QIANG_GANG_HU && (static.CHR_QIANG_GANG&wChiHuRight) != 0 {
		if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_QIANG_GANG
		}
	}
	return wChiHuKind
}

//带赖子的胡牌
func (slh *SportLogicJZHZG) AnalyseMagicHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16, canUseMagic int) uint64 {
	wChiHuKind := uint64(static.CHK_NULL)

	//构造扑克
	cbTempCard := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbTempCard, cbCardIndex[:])

	wMagicCount := byte(0)
	if canUseMagic == -1 {
		wMagicCount = slh.GetMagicCount(cbTempCard)
	} else {
		wMagicCount = byte(canUseMagic)
	}

	if wMagicCount <= 0 {
		return static.WIK_NULL
	}

	//if wMagicCount == 4 {
	//	wChiHuKind |= static.CHK_FOUR_LAIZE
	//}

	// 清空赖子
	cbTempCard[slh.SwitchToCardIndex(slh.MagicCard)] = 0

	if wMagicCount == 1 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if !slh.IsValidCard(slh.SwitchToCardData(i)) {
				//无效牌值
				continue
			}
			cbTempCard[i]++

			wChiHuKind |= slh.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)

			cbTempCard[i]--
		}
	} else if wMagicCount == 2 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			for j := byte(0); j < static.MAX_INDEX; j++ {
				if !slh.IsValidCard(slh.SwitchToCardData(i)) {
					continue
				}
				if !slh.IsValidCard(slh.SwitchToCardData(j)) {
					continue
				}
				cbTempCard[i]++
				cbTempCard[j]++
				wChiHuKind |= slh.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
				cbTempCard[i]--
				cbTempCard[j]--
			}
		}
	} else if wMagicCount == 3 {
		byDisCount, byNeedMagicCount, abyDisTable := slh.CreatDisperseTable(cbTempCard)
		if false == slh.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
			return static.CHK_NULL //不能胡牌
		}
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if 0 == abyDisTable[i] {
				continue
			}
			for j := byte(0); j < static.MAX_INDEX; j++ {
				if 0 == abyDisTable[j] {
					continue
				}
				for k := byte(0); k < static.MAX_INDEX; k++ {
					if 0 == abyDisTable[k] {
						continue
					}
					if !slh.IsValidCard(slh.SwitchToCardData(i)) {
						continue
					}
					if !slh.IsValidCard(slh.SwitchToCardData(j)) {
						continue
					}
					if !slh.IsValidCard(slh.SwitchToCardData(k)) {
						continue
					}

					cbTempCard[i]++
					cbTempCard[j]++
					cbTempCard[k]++
					//分析记录清理
					wChiHuKind |= slh.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
					cbTempCard[i]--
					cbTempCard[j]--
					cbTempCard[k]--
				}
			}
		}
	} else if wMagicCount == 4 {
		//离散表
		byDisCount, byNeedMagicCount, abyDisTable := slh.CreatDisperseTable(cbTempCard)
		if false == slh.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
			return static.CHK_NULL //不能胡牌
		}
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if 0 == abyDisTable[i] {
				continue
			}
			for j := byte(0); j < static.MAX_INDEX; j++ {
				if 0 == abyDisTable[j] {
					continue
				}
				for k := byte(0); k < static.MAX_INDEX; k++ {
					if 0 == abyDisTable[k] {
						continue
					}
					for m := byte(0); m < static.MAX_INDEX; m++ {
						if 0 == abyDisTable[m] {
							continue
						}
						if !slh.IsValidCard(slh.SwitchToCardData(i)) {
							continue
						}
						if !slh.IsValidCard(slh.SwitchToCardData(j)) {
							continue
						}
						if !slh.IsValidCard(slh.SwitchToCardData(k)) {
							continue
						}
						if !slh.IsValidCard(slh.SwitchToCardData(m)) {
							continue
						}

						cbTempCard[i]++
						cbTempCard[j]++
						cbTempCard[k]++
						cbTempCard[m]++
						//分析记录清理
						wChiHuKind |= slh.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
						cbTempCard[i]--
						cbTempCard[j]--
						cbTempCard[k]--
						cbTempCard[m]--

					} //end for
				}
			}
		}
	}

	if wChiHuKind != static.CHK_NULL {
		//这里只有带来自的胡牌类型
		if (wChiHuKind & static.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind ^= static.CHK_DA_HU_NOMAGIC
		}
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind ^= static.CHK_PING_HU_NOMAGIC
		}
		if slh.IsDaHuKind(wChiHuKind) {
			wChiHuKind |= static.CHK_DA_HU_MAGIC
		} else {
			wChiHuKind |= static.CHK_PING_HU_MAGIC
		}
	}

	return wChiHuKind
}

//吃胡判断
func (slh *SportLogicJZHZG) AnalyseChiHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult, checkhz bool) byte {
	//变量定义
	wChiHuKind := uint64(static.CHK_NULL)
	//构造麻将
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

	copy(cbCardIndexTemp, cbCardIndex)

	if cbCurrentCard != 0 {
		cbCardIndexTemp[slh.SwitchToCardIndex(cbCurrentCard)]++
	}

	//结果判断硬胡,红中不可能是硬胡，不管
	wChiHuKind |= slh.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbWeaveCount)

	if slh.Rule.WanNengPai && wChiHuKind == static.CHK_NULL {
		//结果判断软胡
		wMagicCount := slh.GetMagicCount(cbCardIndexTemp)
		if wMagicCount == 4 && checkhz {
			ChiHuResult.ChiHuKind |= static.CHK_SI_LAIZI_NO_HUPAI
			return static.WIK_CHI_HU
		}
		if wMagicCount > 0 && wMagicCount <= 4 {
			wChiHuKind_magic := uint64(static.CHK_NULL)
			wChiHuKind_magic |= slh.AnalyseMagicHuCard(cbCardIndexTemp, WeaveItem, cbWeaveCount, cbCurrentCard, wChiHuRight, -1)

			//有软胡
			if wChiHuKind_magic != static.CHK_NULL {
				if !slh.IsDaHuKind(wChiHuKind_magic) {
					if wChiHuKind != static.CHK_NULL {

					} else {
						// 没有硬胡的软小胡算软小胡
						if wChiHuKind_magic != static.CHK_NULL {
							wChiHuKind = wChiHuKind_magic
						}
					}
				} else {
					if wChiHuKind != static.CHK_NULL {
						if slh.IsDaHuKind(wChiHuKind) {
							//有硬大胡的软大胡叠加
							wChiHuKind |= wChiHuKind_magic
						} else {
							//没有硬大胡的软大胡算软大胡
							wChiHuKind = wChiHuKind_magic
						}
					} else {
						//只有软大胡
						wChiHuKind = wChiHuKind_magic
					}
				}
			}
		}
	}

	ChiHuResult.ChiHuKind = wChiHuKind
	if ChiHuResult.ChiHuKind != static.CHK_NULL {
		return static.WIK_CHI_HU
	}

	return static.WIK_NULL
}

//是否是大胡
func (slh *SportLogicJZHZG) IsDaHuKind(wChiHuKind uint64) bool {
	if (wChiHuKind&static.CHK_QING_YI_SE) != 0 ||
		(wChiHuKind&static.CHK_GANG_SHANG_KAI_HUA) != 0 ||
		(wChiHuKind&static.CHK_7_DUI) != 0 {
		return true
	}
	return false
}

//听牌判断,判断手上听多少张牌，0张表示没听牌
//func (self *SportLogicJZHZG) AnalyseTingCardCount(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) byte {
//	iTingCount := 0
//	//变量定义
//	var ChiHuResult static.TagChiHuResult
//
//	//构造扑克
//	cbCardIndexTemp := cbCardIndex[:]
//
//	//听牌分析
//	y := 0
//	for i := byte(0); i < static.MAX_INDEX; i++ {
//		//胡牌分析
//		cbCurrentCard := self.SwitchToCardData(i)
//		cbHuCardKind := self.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult)
//
//		//结果判断
//		if cbHuCardKind != static.CHK_NULL { //赖子
//			y++ //计数
//		}
//	}
//
//	iTingCount = y
//
//	return byte(iTingCount)
//}

// 听牌判断,当手上是13张牌的时候,分析需要一张什么牌胡牌
//func (self *SportLogicJZHZG) AnalyseTingCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) bool {
//	bIsTing := false
//	//变量定义
//	var ChiHuResult static.TagChiHuResult
//
//	//构造扑克
//	cbCardIndexTemp := cbCardIndex[:]
//
//	//听牌分析
//	for i := byte(0); i < static.MAX_INDEX; i++ {
//		//胡牌分析
//		cbCurrentCard := self.SwitchToCardData(i)
//		cbHuCardKind := self.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult)
//		//结果判断
//		if cbHuCardKind != static.CHK_NULL {
//			bIsTing = true
//		}
//	}
//	return bIsTing
//}
