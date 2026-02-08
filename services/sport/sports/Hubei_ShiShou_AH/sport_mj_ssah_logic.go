package Hubei_ShiShou_AH

import (
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
)

//胡牌类型掩码数组
var HuKindMask = []int{
	static.CHK_GANG_SHANG_KAI_HUA,
	static.CHK_7_DUI,
	static.CHK_JIANG_JIANG,
	static.CHK_QING_YI_SE,
	static.CHK_PENG_PENG,
	static.CHK_QUAN_QIU_REN,
}

/**
=============================分隔符=======================================
本游戏logic层    =========================================================
=============================分隔符=======================================
*/

type SportLogicSSAH struct {
	logic2.BaseLogic
}

//碰牌判断
func (self *SportLogicSSAH) EstimatePengCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	//ASSERT(IsValidCard(cbCurrentCard));
	if !self.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}
	//普通碰牌判断
	if cbCardIndex[self.SwitchToCardIndex(cbCurrentCard)] >= 2 {
		if self.BaseLogic.Rule.IsHaveMagic && self.BaseLogic.Rule.TypeLaizi == 0 {
			cardCount := 0
			for i := 0; i < static.MAX_INDEX; i++ {
				if cbCardIndex[i] > 0 {
					cardCount += int(cbCardIndex[i])
				}
			}
			if cbCardIndex[self.SwitchToCardIndex(self.BaseLogic.MagicCard)] == byte(cardCount)-cbCardIndex[self.SwitchToCardIndex(cbCurrentCard)] {
				return static.WIK_NULL
			}
		}
		return static.WIK_PENG
	}
	//
	return static.WIK_NULL
}

//吃牌判断
func (self *SportLogicSSAH) EstimateEatCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	if !self.IsValidCard(cbCurrentCard) {
		//TODO
		return static.WIK_NULL
	}

	//过滤判断
	if cbCurrentCard >= 0x30 {
		return static.WIK_NULL
	}

	//变量定义
	cbExcursion := [3]byte{0, 1, 2}
	cbItemKind := [3]byte{static.WIK_LEFT, static.WIK_CENTER, static.WIK_RIGHT}

	//吃牌判断
	var cbEatKind, cbFirstIndex byte = 0, 0
	var cbCurrentIndex byte = self.SwitchToCardIndex(cbCurrentCard)
	//cbMgicCardIndex := self.SwitchToCardIndex(self.MagicCard)
	for i := 0; i < len(cbItemKind); i++ {
		var cbValueIndex byte = cbCurrentIndex % 9
		if (cbValueIndex >= cbExcursion[i]) && ((cbValueIndex - cbExcursion[i]) <= 6) {
			//吃牌判断
			cbFirstIndex = cbCurrentIndex - cbExcursion[i]
			if (cbCurrentIndex != cbFirstIndex) && (cbCardIndex[cbFirstIndex] == 0) {
				continue
			}
			if (cbCurrentIndex != (cbFirstIndex + 1)) && (cbCardIndex[cbFirstIndex+1] == 0) {
				continue
			}
			if (cbCurrentIndex != (cbFirstIndex + 2)) && (cbCardIndex[cbFirstIndex+2] == 0) {
				continue
			}

			chiMacic := false
			cbMagicCard := self.SwitchToCardIndex(self.MagicCard)
			if i == 0 && (cbMagicCard == cbCurrentIndex+1 || cbMagicCard == cbCurrentIndex+2) {
				chiMacic = true
				break
			} else if i == 1 && (cbMagicCard == cbCurrentIndex-1 || cbMagicCard == cbCurrentIndex+1) {
				chiMacic = true
				break
			} else if i == 2 && (cbMagicCard == cbCurrentIndex-1 || cbMagicCard == cbCurrentIndex-2) {
				chiMacic = true
				break
			}
			if chiMacic {
				continue
			}

			//设置类型
			cbEatKind |= cbItemKind[i]
		}
	}

	return cbEatKind
}

func (self *SportLogicSSAH) findMagicCard(cardsclass int, PiZiCard byte) (byte, error) {
	if PiZiCard == static.INVALID_BYTE {
		//硬晃，没赖子
		return PiZiCard, nil
	}
	return mahlib2.FindMagicCard(cardsclass, PiZiCard)
}

// 听牌判断,当手上是13张牌的时候,分析需要一张什么牌胡牌
func (self *SportLogicSSAH) AnalyseTingCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) bool {
	bIsTing := false
	//变量定义
	var ChiHuResult static.TagChiHuResult

	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]

	// 有红中不能听牌
	if cbCardIndexTemp[31] != 0 {
		return bIsTing
	}
	// 发财杠有发财不能听牌
	if self.Rule.GangType == logic2.Gang_Type_FaCaiGang && cbCardIndexTemp[32] != 0 {
		return bIsTing
	}

	// 痞子杠有痞子不能听牌
	if self.Rule.GangType == logic2.Gang_Type_PiZiGang && cbCardIndexTemp[self.SwitchToCardIndex(self.PiZiCard)] != 0 {
		return bIsTing
	}

	//有红中不听牌

	wMagicCount := self.GetMagicCount(cbCardIndexTemp)

	//听牌分析
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//胡牌分析
		cbCurrentCard := self.SwitchToCardData(i)
		cbHuCardKind := self.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult)
		if wMagicCount > 1 && !self.IsDaHuKind(ChiHuResult.ChiHuKind) {
			continue
		}
		//结果判断
		if cbHuCardKind != static.CHK_NULL {
			bIsTing = true
		}
	}
	return bIsTing
}

//吃胡判断
func (self *SportLogicSSAH) AnalyseChiHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult) byte {
	//变量定义
	wChiHuKind := uint64(static.CHK_NULL)
	//构造麻将
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

	copy(cbCardIndexTemp, cbCardIndex)

	if cbCurrentCard != 0 {
		cbCardIndexTemp[self.SwitchToCardIndex(cbCurrentCard)]++
	}

	// 手里有红中不能胡牌
	if cbCardIndexTemp[31] > 0 {
		return static.WIK_NULL
	}

	// 发财杠手里有发财不能胡
	if self.Rule.GangType == logic2.Gang_Type_FaCaiGang {
		if cbCardIndexTemp[32] > 0 {
			return static.WIK_NULL
		}
	}

	// 痞子杠有痞子不能胡牌
	if self.Rule.GangType == logic2.Gang_Type_PiZiGang {
		if cbCardIndexTemp[self.SwitchToCardIndex(self.PiZiCard)] > 0 {
			return static.WIK_NULL
		}
	}

	//结果判断硬胡
	wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbWeaveCount)

	//结果判断软胡
	wMagicCount := self.GetMagicCount(cbCardIndexTemp)
	if wMagicCount > 0 && wMagicCount <= 4 {
		wChiHuKind_magic := uint64(static.CHK_NULL)
		wChiHuKind_magic |= self.AnalyseMagicHuCard(cbCardIndexTemp, WeaveItem, cbWeaveCount, cbCurrentCard, wChiHuRight)

		//有软胡
		if wChiHuKind_magic != static.CHK_NULL {
			if !self.IsDaHuKind(wChiHuKind_magic) {
				if wMagicCount <= 1 {
					// 没有硬胡的软小胡算软小胡
					if wChiHuKind == static.CHK_NULL {
						wChiHuKind = wChiHuKind_magic
					}
				} else if !self.IsDaHuKind(wChiHuKind) {
					wChiHuKind = uint64(static.CHK_NULL)
				}

			} else {
				if wChiHuKind != static.CHK_NULL {
					if self.IsDaHuKind(wChiHuKind) {
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

	ChiHuResult.ChiHuKind = wChiHuKind
	if ChiHuResult.ChiHuKind != static.CHK_NULL {
		if wMagicCount == 4 {
			ChiHuResult.ChiHuKind |= static.CHK_FOUR_LAIZE
		}
		return static.WIK_CHI_HU
	}

	return static.WIK_NULL
}

//是否是大胡
func (self *SportLogicSSAH) IsDaHuKind(wChiHuKind uint64) bool {
	if (wChiHuKind&static.CHK_FENG_YI_SE) != 0 ||
		(wChiHuKind&static.CHK_JIANG_JIANG) != 0 ||
		(wChiHuKind&static.CHK_QING_YI_SE) != 0 ||
		(wChiHuKind&static.CHK_PENG_PENG) != 0 ||
		(wChiHuKind&static.CHK_QUAN_QIU_REN) != 0 ||
		(wChiHuKind&static.CHK_HAI_DI) != 0 ||
		(wChiHuKind&static.CHK_GANG_SHANG_KAI_HUA) != 0 ||
		(wChiHuKind&static.CHK_QIANG_GANG) != 0 ||
		(wChiHuKind&static.CHK_7_DUI) != 0 {
		return true
	}
	return false
}

//基本胡牌分析
func (self *SportLogicSSAH) AnalyseHuKind(wChiHuRight uint16, wChiHuKind uint64, cbTempCard []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) uint64 {
	//清一色：手中的牌和组合牌都是一种花色，乱将
	if self.HuType.HAVE_QING_YISE_HU && self.IsQingYiSe(cbTempCard, WeaveItem, cbWeaveCount) {
		wChiHuRight |= static.CHR_QING_YI_SE
	}

	//风一色
	if self.HuType.HAVE_FENG_YISE_HU && self.IsFengYiSe(cbTempCard, WeaveItem, cbWeaveCount) {
		wChiHuKind |= static.CHK_FENG_YI_SE
	}

	//将一色
	if self.HuType.HAVE_JIANG_JIANG_HU && self.IsJiangJiangHu(cbTempCard, WeaveItem, cbWeaveCount) {
		wChiHuKind |= static.CHK_JIANG_JIANG
	}

	//七对
	if self.Rule.NoKouKeHu && self.Rule.QiDuiKeHu && cbWeaveCount == 0 {
		if self.HuType.HAVE_QIDUI_HU {
			Hu7Dui, _ := self.IsQiDui(cbTempCard)
			if Hu7Dui {
				wChiHuKind |= static.CHK_7_DUI
			}
		}
	}

	_, AnalyseItemArray := self.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//牌型分析
		for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
			//变量定义
			bLianCard, bPengCard := false, false
			var pAnalyseItem *static.TagAnalyseItem = &AnalyseItemArray[iCount]

			//得到可胡牌类型中牌眼的值
			cbEyeValue := pAnalyseItem.CardEye & static.MASK_VALUE
			//得到胡牌类型中对牌
			cbEyeCard := pAnalyseItem.CardEye

			//判断牌眼是不是将将必须是2 5 8，如果是2的时候对牌不能是风牌（南风） 如果是5的时候不能是风牌（红中）
			bSymbolEye := ((cbEyeValue == 2 && cbEyeCard != 0x32) || (cbEyeValue == 5 && cbEyeCard != 0x35) || (cbEyeValue == 8))

			//对一个胡牌分析子项进行分析，如果有一个吃牌类型，就记bLianCard为真，如果有一个碰牌类型就记bPengCard为真
			for j := 0; j < len(pAnalyseItem.WeaveKind); j++ {
				cbWeaveKind := pAnalyseItem.WeaveKind[j]
				if cbWeaveKind&(static.WIK_GANG|static.WIK_FILL|static.WIK_PENG) != 0 {
					bPengCard = true
				}

				if (cbWeaveKind & (static.WIK_LEFT | static.WIK_CENTER | static.WIK_RIGHT)) != 0 {
					bLianCard = true
				}
			}

			//1.软碰碰牌分析子项中，没有吃牌类型，必须都是碰牌类型(碰碰胡需要2 5 8做将，按小胡计算)
			if self.HuType.HAVE_PENG_PENG_HU && (bLianCard == false) && (bPengCard == true) {
				wChiHuKind |= static.CHK_PENG_PENG
			}

			// 有清一色的牌权并且牌型可以胡
			if self.HuType.HAVE_QING_YISE_HU && (static.CHR_QING_YI_SE&wChiHuRight) != 0 {
				wChiHuKind |= static.CHK_QING_YI_SE
			}

			// 需要成牌型且要258做将的牌型
			if bSymbolEye {
				wChiHuKind |= static.CHK_PING_HU_NOMAGIC
			}

			// 全球人
			if self.HuType.HAVE_QUAN_QIU_REN && (static.CHR_QUAN_QIU_REN&wChiHuRight) != 0 && bSymbolEye {
				if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
					wChiHuKind |= static.CHK_QUAN_QIU_REN
				}
			}

			if self.HuType.HAVE_HAI_DI_HU && (static.CHR_HAI_DI&wChiHuRight) != 0 && bSymbolEye {
				wChiHuKind |= static.CHK_HAI_DI
			}

			if self.IsDaHuKind(wChiHuKind) {
				wChiHuKind |= static.CHK_DA_HU_NOMAGIC
			}
		}
	} else {
		if (wChiHuKind&static.CHK_FENG_YI_SE) != 0 || (wChiHuKind&static.CHK_JIANG_JIANG) != 0 || (wChiHuKind&static.CHK_7_DUI) != 0 {
			wChiHuKind |= static.CHK_DA_HU_NOMAGIC
		}

		// 有清一色的牌权并且牌型可以胡
		if self.HuType.HAVE_QING_YISE_HU && (static.CHR_QING_YI_SE&wChiHuRight) != 0 {
			if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
				wChiHuKind |= static.CHK_QING_YI_SE
			}
		}
	}

	// 杠上开花判断
	if self.HuType.HAVE_GANG_SHANG_KAI_HUA && (static.CHR_GANG_SHANG_KAI_HUA&wChiHuRight) != 0 {
		if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_GANG_SHANG_KAI_HUA
		}
	}

	// 抢杠胡判断
	if self.HuType.HAVE_QIANG_GANG_HU && (static.CHR_QIANG_GANG&wChiHuRight) != 0 {
		if (wChiHuKind&static.CHK_PING_HU_NOMAGIC) != 0 || (wChiHuKind&static.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind |= static.CHK_QIANG_GANG
		}
	}

	return wChiHuKind
}

//带赖子的胡牌
func (self *SportLogicSSAH) AnalyseMagicHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16) uint64 {
	wChiHuKind := uint64(static.CHK_NULL)

	//构造扑克
	cbTempCard := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbTempCard, cbCardIndex[:])

	wMagicCount := self.GetMagicCount(cbTempCard)

	if wMagicCount <= 0 {
		return static.WIK_NULL
	}

	// 清空赖子
	cbTempCard[self.SwitchToCardIndex(self.MagicCard)] = 0

	if wMagicCount == 1 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if !self.IsValidCard(self.SwitchToCardData(i)) {
				//无效牌值
				continue
			}
			if i == 31 || i == 32 {
				//红中发财不能胡
				continue
			}
			cbTempCard[i]++

			wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)

			cbTempCard[i]--
		}
	} else if wMagicCount == 2 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if i == 31 || i == 32 {
				//红中发财不能胡
				continue
			}
			for j := byte(0); j < static.MAX_INDEX; j++ {
				if j == 31 || j == 32 {
					//红中发财不能胡
					continue
				}
				if !self.IsValidCard(self.SwitchToCardData(i)) {
					continue
				}
				if !self.IsValidCard(self.SwitchToCardData(j)) {
					continue
				}
				cbTempCard[i]++
				cbTempCard[j]++
				wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
				cbTempCard[i]--
				cbTempCard[j]--
			}
		}
	} else if wMagicCount == 3 {
		byDisCount, byNeedMagicCount, abyDisTable := self.CreatDisperseTable(cbTempCard)
		if false == self.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
			return static.CHK_NULL //不能胡牌
		}
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if 0 == abyDisTable[i] {
				continue
			}
			if i == 31 || i == 32 {
				//红中发财不能胡
				continue
			}
			for j := byte(0); j < static.MAX_INDEX; j++ {
				if 0 == abyDisTable[j] {
					continue
				}
				if j == 31 || j == 32 {
					//红中发财不能胡
					continue
				}
				for k := byte(0); k < static.MAX_INDEX; k++ {
					if 0 == abyDisTable[k] {
						continue
					}
					if k == 31 || k == 32 {
						//红中发财不能胡
						continue
					}
					if !self.IsValidCard(self.SwitchToCardData(i)) {
						continue
					}
					if !self.IsValidCard(self.SwitchToCardData(j)) {
						continue
					}
					if !self.IsValidCard(self.SwitchToCardData(k)) {
						continue
					}

					cbTempCard[i]++
					cbTempCard[j]++
					cbTempCard[k]++
					//分析记录清理
					wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
					cbTempCard[i]--
					cbTempCard[j]--
					cbTempCard[k]--
				}
			}
		}
	} else if wMagicCount == 4 {
		//离散表
		byDisCount, byNeedMagicCount, abyDisTable := self.CreatDisperseTable(cbTempCard)
		if false == self.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
			return static.CHK_NULL //不能胡牌
		}
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if 0 == abyDisTable[i] {
				continue
			}
			if i == 31 || i == 32 {
				//红中发财不能胡
				continue
			}
			for j := byte(0); j < static.MAX_INDEX; j++ {
				if 0 == abyDisTable[j] {
					continue
				}
				if j == 31 || j == 32 {
					//红中发财不能胡
					continue
				}
				for k := byte(0); k < static.MAX_INDEX; k++ {
					if 0 == abyDisTable[k] {
						continue
					}
					if k == 31 || k == 32 {
						//红中发财不能胡
						continue
					}
					for m := byte(0); m < static.MAX_INDEX; m++ {
						if 0 == abyDisTable[m] {
							continue
						}
						if m == 31 || m == 32 {
							//红中发财不能胡
							continue
						}
						if !self.IsValidCard(self.SwitchToCardData(i)) {
							continue
						}
						if !self.IsValidCard(self.SwitchToCardData(j)) {
							continue
						}
						if !self.IsValidCard(self.SwitchToCardData(k)) {
							continue
						}
						if !self.IsValidCard(self.SwitchToCardData(m)) {
							continue
						}

						cbTempCard[i]++
						cbTempCard[j]++
						cbTempCard[k]++
						cbTempCard[m]++
						//分析记录清理
						wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
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
		if self.IsDaHuKind(wChiHuKind) {
			wChiHuKind |= static.CHK_DA_HU_MAGIC
		} else {
			wChiHuKind |= static.CHK_PING_HU_MAGIC
		}
	}

	return wChiHuKind
}

//杠牌分析
func (self *SportLogicSSAH) AnalyseGangCard(_userItem *components2.Player, cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, GangCardResult *static.TagGangCardResult) (byte, bool) {
	//设置变量
	cbActionMask := static.WIK_NULL

	//癞子杠
	//if cbCardIndex[self.SwitchToCardIndex(self.MagicCard)] > 0 {
	//	cbActionMask |= public.WIK_GANG
	//	GangCardResult.CardData[GangCardResult.CardCount] = public.WIK_GANG
	//	GangCardResult.CardData[GangCardResult.CardCount] = self.MagicCard
	//	GangCardResult.CardCount++
	//}
	//痞子杠
	//for _,v:=range self.PiZiCards{
	//	if cbCardIndex[self.SwitchToCardIndex(v)] > 0 {
	//		cbActionMask |= public.WIK_GANG
	//		GangCardResult.CardData[GangCardResult.CardCount] = public.WIK_GANG
	//		GangCardResult.CardData[GangCardResult.CardCount] = v
	//		GangCardResult.CardCount++
	//	}
	//}

	//手上杠牌
	for i := byte(0); i < static.MAX_INDEX; i++ {
		if i == 31 || i == self.SwitchToCardIndex(self.MagicCard) {
			continue
		}
		if cbCardIndex[i] == 4 {
			//a:=self.SwitchToCardData(i)
			//self.DeleteGiveUpGang(_userItem,a)
			cbActionMask |= static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = self.SwitchToCardData(i)
			GangCardResult.CardCount++
		}
	}
	b := false
	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_PENG {
			if cbCardIndex[self.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				if self.IsGiveUpGang(_userItem, WeaveItem[i].CenterCard) {
					b = true
					break
				}
				cbActionMask |= static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = WeaveItem[i].CenterCard
				GangCardResult.CardCount++
			}
		}
	}

	return byte(cbActionMask), b
}

// 追加一个弃杠牌
func (self *SportLogicSSAH) AppendGiveUpGang(_userItem *components2.Player, card byte) {
	if self.IsGiveUpGang(_userItem, card) {
		return
	}
	_userItem.Ctx.VecGangCard = append(_userItem.Ctx.VecGangCard, card)
}

// 删除一个弃杠牌
func (self *SportLogicSSAH) DeleteGiveUpGang(_userItem *components2.Player, card byte) {
	for i, c := range _userItem.Ctx.VecGangCard {
		if c == card {
			_userItem.Ctx.VecGangCard = append(_userItem.Ctx.VecGangCard[:i], _userItem.Ctx.VecGangCard[i+1:]...)
		}
	}
}

// 是否为弃杠牌
func (self *SportLogicSSAH) IsGiveUpGang(_userItem *components2.Player, card byte) bool {
	for _, c := range _userItem.Ctx.VecGangCard {
		if c == card {
			return true
		}
	}
	return false
}

func (self *SportLogicSSAH) CreateCheckHu() {
	var err error
	_, err = components2.GetHuTable()
	static.HF_CheckErr(err)
}
