// ////////////////////////////////////////////////////////////////////////
//
//	      --
//	  --       --
//	--  游戏逻辑  --
//	  --       --
//	      --
//
// ////////////////////////////////////////////////////////////////////////
package Hubei_JianLi

import (
	public "github.com/open-source/game/chess.git/pkg/static"
	logic "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
)

//////////////////////////////////////////////////////////////////////////

type GameLogic_jlhzlzg struct {
	logic.BaseLogic
}

// 有效判断
func (self *GameLogic_jlhzlzg) IsValidCard(cbCardData byte) bool {
	cbValue := cbCardData & public.MASK_VALUE
	cbColor := (cbCardData & public.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	bIsValid := ((cbValue >= 1) && (cbValue <= 9) && (cbColor <= 2)) ||
		((cbValue >= 1) && (cbValue <= 7) && (cbColor == 3))
	if !bIsValid {
		return false //非法牌
	}

	//校验：去万
	if self.Rule.NoWan && cbColor == 0 {
		return false
	}

	if cbCardData == self.MagicCard {
		return false // 红中为癞子，不能出
	}

	return true
}

// 杠牌分析
func (self *GameLogic_jlhzlzg) AnalyseGangCard(cbCardIndex [public.MAX_INDEX]byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte, GangCardResult *public.TagGangCardResult) byte {
	//设置变量
	cbActionMask := public.WIK_NULL

	//手上杠牌
	for i := byte(0); i < public.MAX_INDEX; i++ {
		if i == self.SwitchToCardIndex(self.MagicCard) {
			continue
		}
		if cbCardIndex[i] == 4 {
			cbActionMask |= public.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = public.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = self.SwitchToCardData(i)
			GangCardResult.CardCount++
		}
	}

	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == public.WIK_PENG {
			if cbCardIndex[self.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				cbActionMask |= public.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = public.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = WeaveItem[i].CenterCard
				GangCardResult.CardCount++
			}
		}
	}

	return byte(cbActionMask)
}

// 基本胡牌分析
func (self *GameLogic_jlhzlzg) AnalyseHuKind(cbTempCard []byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte) uint64 {
	var canHu bool
	wChiHuKind := uint64(public.CHK_NULL)

	//七对
	if self.HuType.HAVE_QIDUI_HU && self.Rule.QiDuiKeHu && cbWeaveCount == 0 {
		hu7Dui, _ := self.IsQiDui(cbTempCard)
		if hu7Dui {
			wChiHuKind |= public.CHK_7_DUI
			wChiHuKind |= public.CHK_PING_HU_NOMAGIC
			canHu = true
		}
	}

	if !canHu {
		_, AnalyseItemArray := self.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
		//胡牌分析
		if len(AnalyseItemArray) > 0 {
			//牌型分析
			for iCount := 0; iCount < len(AnalyseItemArray); iCount++ {
				//变量定义
				var bLianCard, bPengCard bool
				pAnalyseItem := &AnalyseItemArray[iCount]

				//对一个胡牌分析子项进行分析，如果有一个吃牌类型，就记bLianCard为真，如果有一个碰牌类型就记bPengCard为真
				for j := 0; j < len(pAnalyseItem.WeaveKind); j++ {
					cbWeaveKind := pAnalyseItem.WeaveKind[j]
					if cbWeaveKind&(public.WIK_GANG|public.WIK_FILL|public.WIK_PENG) != 0 {
						bPengCard = true
					}

					if (cbWeaveKind & (public.WIK_LEFT | public.WIK_CENTER | public.WIK_RIGHT)) != 0 {
						bLianCard = true
					}
				}

				// 碰碰牌分析子项中，没有吃牌类型，必须都是碰牌类型
				if self.HuType.HAVE_PENG_PENG_HU && (bLianCard == false) && (bPengCard == true) {
					wChiHuKind |= public.CHK_PENG_PENG
				}

				wChiHuKind |= public.CHK_PING_HU_NOMAGIC
			}
			canHu = true
		}
	}
	return wChiHuKind
}

// 带赖子的胡牌
func (self *GameLogic_jlhzlzg) AnalyseMagicHuCard(cbCardIndex []byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte) uint64 {
	wChiHuKind := uint64(public.CHK_NULL)

	//构造扑克
	cbTempCard := make([]byte, public.MAX_INDEX, public.MAX_INDEX)
	copy(cbTempCard, cbCardIndex[:])
	wMagicCount := self.GetMagicCount(cbTempCard)

	if wMagicCount <= 0 {
		return public.WIK_NULL
	}

	// 清空赖子
	cbTempCard[self.SwitchToCardIndex(self.MagicCard)] = 0

	if wMagicCount == 1 {
		for i := byte(0); i < public.MAX_INDEX; i++ {
			if !self.IsValidCard(self.SwitchToCardData(i)) {
				//无效牌值
				continue
			}
			cbTempCard[i]++

			tmpHuKind := self.AnalyseHuKind(cbTempCard, WeaveItem, cbWeaveCount)
			if self.GetHuLevelOnMagic(tmpHuKind) > self.GetHuLevelOnMagic(wChiHuKind) {
				wChiHuKind = tmpHuKind
			}

			cbTempCard[i]--
		}
	} else if wMagicCount == 2 {
		for i := byte(0); i < public.MAX_INDEX; i++ {
			for j := byte(0); j < public.MAX_INDEX; j++ {
				if !self.IsValidCard(self.SwitchToCardData(i)) {
					continue
				}
				if !self.IsValidCard(self.SwitchToCardData(j)) {
					continue
				}
				cbTempCard[i]++
				cbTempCard[j]++
				tmpHuKind := self.AnalyseHuKind(cbTempCard, WeaveItem, cbWeaveCount)
				if self.GetHuLevelOnMagic(tmpHuKind) > self.GetHuLevelOnMagic(wChiHuKind) {
					wChiHuKind = tmpHuKind
				}
				cbTempCard[i]--
				cbTempCard[j]--
			}
		}
	} else if wMagicCount == 3 {
		byDisCount, byNeedMagicCount, abyDisTable := self.CreatDisperseTable(cbTempCard)
		if false == self.DisperseEstimate(byDisCount, byNeedMagicCount, wMagicCount) {
			return public.CHK_NULL //不能胡牌
		}
		for i := byte(0); i < public.MAX_INDEX; i++ {
			if 0 == abyDisTable[i] {
				continue
			}
			for j := byte(0); j < public.MAX_INDEX; j++ {
				if 0 == abyDisTable[j] {
					continue
				}
				for k := byte(0); k < public.MAX_INDEX; k++ {
					if 0 == abyDisTable[k] {
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
					tmpHuKind := self.AnalyseHuKind(cbTempCard, WeaveItem, cbWeaveCount)
					if self.GetHuLevelOnMagic(tmpHuKind) > self.GetHuLevelOnMagic(wChiHuKind) {
						wChiHuKind = tmpHuKind
					}
					cbTempCard[i]--
					cbTempCard[j]--
					cbTempCard[k]--
				}
			}
		}
	} else if wMagicCount == 4 {
		//离散表
		byDisCount, byNeedMagicCount, abyDisTable := self.CreatDisperseTable(cbTempCard)
		if false == self.DisperseEstimate(byDisCount, byNeedMagicCount, wMagicCount) {
			return public.CHK_NULL //不能胡牌
		}
		for i := byte(0); i < public.MAX_INDEX; i++ {
			if 0 == abyDisTable[i] {
				continue
			}
			for j := byte(0); j < public.MAX_INDEX; j++ {
				if 0 == abyDisTable[j] {
					continue
				}
				for k := byte(0); k < public.MAX_INDEX; k++ {
					if 0 == abyDisTable[k] {
						continue
					}
					for m := byte(0); m < public.MAX_INDEX; m++ {
						if 0 == abyDisTable[m] {
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
						tmpHuKind := self.AnalyseHuKind(cbTempCard, WeaveItem, cbWeaveCount)
						if self.GetHuLevelOnMagic(tmpHuKind) > self.GetHuLevelOnMagic(wChiHuKind) {
							wChiHuKind = tmpHuKind
						}
						cbTempCard[i]--
						cbTempCard[j]--
						cbTempCard[k]--
						cbTempCard[m]--

					} //end for
				}
			}
		}
	}

	if wChiHuKind != public.CHK_NULL {
		//这里只有带来自的胡牌类型
		if (wChiHuKind & public.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind ^= public.CHK_DA_HU_NOMAGIC
		}
		if (wChiHuKind & public.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind ^= public.CHK_PING_HU_NOMAGIC
		}
		wChiHuKind |= public.CHK_PING_HU_MAGIC
	}

	return wChiHuKind
}

// 吃胡判断
func (self *GameLogic_jlhzlzg) AnalyseChiHuCard(cbCardIndex []byte, WeaveItem []public.TagWeaveItem, cbWeaveCount,
	cbCurrentCard byte, wChiHuRight uint16, ChiHuResult *public.TagChiHuResult) byte {

	//变量定义
	wChiHuKind := uint64(public.CHK_NULL)
	//构造麻将
	cbCardIndexTemp := make([]byte, public.MAX_INDEX, public.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex)

	if cbCurrentCard != 0 {
		cbCardIndexTemp[self.SwitchToCardIndex(cbCurrentCard)]++
	}

	wMagicCount := self.GetMagicCount(cbCardIndexTemp)

	//判断硬胡
	if wMagicCount == 0 {
		wChiHuKind = self.AnalyseHuKind(cbCardIndexTemp, WeaveItem, cbWeaveCount)
	}

	//判断软胡
	if wMagicCount > 0 && wMagicCount <= 4 {
		wChiHuKindMagic := uint64(public.CHK_NULL)
		wChiHuKindMagic = self.AnalyseMagicHuCard(cbCardIndexTemp, WeaveItem, cbWeaveCount)
		//有软胡
		if wChiHuKindMagic != public.CHK_NULL {
			wChiHuKind = wChiHuKindMagic
		}
	}

	if wChiHuRight&public.CHR_TIAN_HU != 0 && wMagicCount != 4 {
		wChiHuRight ^= public.CHR_TIAN_HU
	}
	if wMagicCount == 4 && wChiHuRight&public.CHR_TIAN_HU != 0 {
		wChiHuKind |= public.CHK_PING_HU_NOMAGIC
	}

	if wChiHuKind != public.CHK_NULL {
		ChiHuResult.ChiHuKind = wChiHuKind
		ChiHuResult.ChiHuRight = wChiHuRight

		return public.WIK_CHI_HU
	}

	return public.WIK_NULL
}

func (self *GameLogic_jlhzlzg) GetHuLevelOnMagic(ChiHuKind uint64) int {
	if ChiHuKind == 0 {
		return 0
	}

	var level int
	level++ // 能胡默认加1
	if (ChiHuKind & public.CHK_7_DUI) != 0 {
		level += 1
	}
	return level
}

// 听牌判断,判断手上听多少张牌，0张表示没听牌
func (self *GameLogic_jlhzlzg) AnalyseTingCardCount(cbCardIndex []byte, WeaveItem []public.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) byte {
	//变量定义
	var ChiHuResult public.TagChiHuResult

	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]

	//听牌分析
	y := 0
	for i := byte(0); i < public.MAX_INDEX; i++ {
		//胡牌分析
		cbCurrentCard := self.SwitchToCardData(i)
		cbHuCardKind := self.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult)

		//结果判断
		if cbHuCardKind != public.CHK_NULL {
			y++ //计数
		}
	}

	return byte(y)
}
