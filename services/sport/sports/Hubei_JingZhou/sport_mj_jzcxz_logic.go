//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  汉川搓虾子游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package Hubei_JingZhou

import (
	"github.com/open-source/game/chess.git/pkg/static"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	"math/rand"
)

//////////////////////////////////////////////////////////////////////////
//扑克数据
var jzcxz_cbCardDataArray = [static.ALL_CARD]byte{
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, //万子

	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, //索子

	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
	0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, //同子
}

//胡牌类型掩码数组  软自摸，硬自摸，软癞油，硬癞油，软清一色癞油，硬清一色癞油，软清一色，硬清一色
var jzcxz_KindMask = [...]int{
	static.CHK_FENG_YI_SE,   //风一色
	static.CHK_JIANG_JIANG,  //将一色
	static.CHK_QING_YI_SE,   //清一色
	static.CHK_QUAN_QIU_REN, //全求人
	static.CHK_PENG_PENG,    //碰碰胡
	static.CHK_GANG_SHANG_KAI_HUA,
}

type SportLogicJZCXZ struct {
	logic2.BaseLogic
}

//混乱扑克
func (spl *SportLogicJZCXZ) RandCardData() (byte, []byte) {
	//混乱准备
	//汉川搓虾子108张牌，不包含东南西北中发白
	cbCardDataTemp := []byte{}
	var cbMaxCount byte = static.ALL_CARD
	if spl.Rule.NoWan {
		for i := 0; i < len(jzcxz_cbCardDataArray[36:]); i++ {
			cbCardDataTemp = append(cbCardDataTemp, jzcxz_cbCardDataArray[36:][i])
		}
		cbMaxCount = cbMaxCount - 36
	} else {
		for i := 0; i < len(jzcxz_cbCardDataArray); i++ {
			cbCardDataTemp = append(cbCardDataTemp, jzcxz_cbCardDataArray[i])
		}
	}

	cbCardData := make([]byte, cbMaxCount)
	//混乱扑克
	cbRandCount, cbPosition := 0, 0
	randTmp := 0
	nAccert := 0
	for {
		nAccert++
		if nAccert > 200 {
			break
		}
		randTmp = int(cbMaxCount) - cbRandCount - 1
		if randTmp > 0 {
			cbPosition = rand.Intn(randTmp)
		} else {
			cbPosition = 0
		}
		cbCardData[cbRandCount] = cbCardDataTemp[cbPosition]
		cbRandCount++
		cbCardDataTemp[cbPosition] = cbCardDataTemp[int(cbMaxCount)-cbRandCount]
		if cbRandCount >= int(cbMaxCount) {
			break
		}
	}

	//syslog.Logger().Infoln(cbCardData)

	return cbMaxCount, cbCardData[:]
}

//有效判断
func (spl *SportLogicJZCXZ) IsValidCard(cbCardData ...byte) bool {
	for _, card := range cbCardData {
		cbValue := (card & static.MASK_VALUE)
		cbColor := (card & static.MASK_COLOR) >> 4

		//校验：万条同+东南西北中发白
		bIsValid := (((cbValue >= 1) && (cbValue <= 9) && (cbColor <= 2)) || ((cbValue >= 1) && (cbValue <= 7) && (cbColor == 3)))
		if !bIsValid {
			return false //非法牌
		}
	}
	return true
}

// 通过皮子牌得到癞子牌
func (spl *SportLogicJZCXZ) MagicByFur(fur byte) byte {
	cbValue := byte(fur & static.MASK_VALUE)
	cbColor := byte(fur & static.MASK_COLOR)
	if cbValue == 9 && cbColor <= 0x20 {
		cbValue = 0
	}
	return (cbValue + 1) | cbColor
}

//杠牌分析
func (spl *SportLogicJZCXZ) AnalyseGangCard(cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, GangCardResult *static.TagGangCardResult, giveUpGangCard []byte) byte {
	//设置变量
	cbActionMask := static.WIK_NULL
	//手上杠牌
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//皮子的话三张皮子就可以杠牌
		if spl.BaseLogic.SwitchToCardIndex(spl.PiZiCard) == i {
			if cbCardIndex[i] == 3 {
				cbActionMask |= static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = spl.SwitchToCardData(i)
				GangCardResult.CardCount++
			}
		} else {
			if cbCardIndex[i] == 4 {
				//4个赖子不显示杠
				if spl.BaseLogic.SwitchToCardIndex(spl.MagicCard) != i {
					cbActionMask |= static.WIK_GANG
					GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
					GangCardResult.CardData[GangCardResult.CardCount] = spl.SwitchToCardData(i)
					GangCardResult.CardCount++
				}
			}
		}
	}

	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_PENG {
			if cbCardIndex[spl.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				giveUpGang := false
				//不分析弃杠的牌
				for _, giveUpGangCard := range giveUpGangCard {
					if giveUpGangCard == WeaveItem[i].CenterCard {
						giveUpGang = true
						break
					}
				}

				//有碰有杠 选碰弃杠 不允许在回头杠
				if giveUpGang {
					continue
				}

				cbActionMask |= static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = WeaveItem[i].CenterCard
				GangCardResult.CardCount++
			}
		}
	}

	return byte(cbActionMask)
}

//基本胡牌分析
func (spl *SportLogicJZCXZ) AnalyseHuKind(wChiHuRight uint16, wChiHuKind uint64, cbTempCard []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, isMagicHuCheck bool) uint64 {
	_, AnalyseItemArray := spl.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//胡牌分析
		//清一色：手中的牌和组合牌都是一种花色，乱将,没能赖子或者赖子不替值
		if spl.IsQingYiSe(cbTempCard, WeaveItem, cbWeaveCount) {
			wChiHuRight |= static.CHR_QING_YI_SE
		}

		//牌型分析
		if isMagicHuCheck {
			wChiHuKind |= static.CHK_PING_HU_MAGIC
		} else {
			wChiHuKind |= static.CHK_PING_HU_NOMAGIC
		}

		//杠上开花
		if (static.CHR_GANG_SHANG_KAI_HUA & wChiHuRight) != 0 {
			wChiHuKind |= static.CHK_GANG_SHANG_KAI_HUA
		}

		//可以胡清一色
		if (static.CHR_QING_YI_SE & wChiHuRight) != 0 {
			wChiHuKind |= static.CHK_QING_YI_SE
		}
	}

	return wChiHuKind
}

//带赖子的胡牌
func (spl *SportLogicJZCXZ) AnalyseMagicHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16) uint64 {

	wChiHuKind := uint64(static.CHK_NULL)

	//构造扑克
	cbTempCard := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbTempCard, cbCardIndex[:])

	wMagicCount := spl.GetMagicCount(cbTempCard)

	if wMagicCount <= 0 {
		return static.WIK_NULL
	}

	// 清空赖子
	cbTempCard[spl.SwitchToCardIndex(spl.MagicCard)] = 0

	if wMagicCount == 1 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			if !spl.IsValidCard(spl.SwitchToCardData(i)) {
				//无效牌值
				continue
			}

			cbTempCard[i]++

			wChiHuKind |= spl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount, true)

			cbTempCard[i]--
		}
	} else if wMagicCount == 2 {
		for i := byte(0); i < static.MAX_INDEX; i++ {

			for j := byte(0); j < static.MAX_INDEX; j++ {

				if !spl.IsValidCard(spl.SwitchToCardData(i)) {
					continue
				}
				if !spl.IsValidCard(spl.SwitchToCardData(j)) {
					continue
				}
				cbTempCard[i]++
				cbTempCard[j]++
				wChiHuKind |= spl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount, true)
				cbTempCard[i]--
				cbTempCard[j]--
			}
		}
	} else if wMagicCount == 3 {
		byDisCount, byNeedMagicCount, abyDisTable := spl.CreatDisperseTable(cbTempCard)
		if false == spl.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
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

					if !spl.IsValidCard(spl.SwitchToCardData(i)) {
						continue
					}
					if !spl.IsValidCard(spl.SwitchToCardData(j)) {
						continue
					}
					if !spl.IsValidCard(spl.SwitchToCardData(k)) {
						continue
					}

					cbTempCard[i]++
					cbTempCard[j]++
					cbTempCard[k]++
					//分析记录清理
					wChiHuKind |= spl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount, true)
					cbTempCard[i]--
					cbTempCard[j]--
					cbTempCard[k]--
				}
			}
		}
	} else if wMagicCount == 4 {
		//离散表
		byDisCount, byNeedMagicCount, abyDisTable := spl.CreatDisperseTable(cbTempCard)
		if false == spl.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
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

						if !spl.IsValidCard(spl.SwitchToCardData(i)) {
							continue
						}
						if !spl.IsValidCard(spl.SwitchToCardData(j)) {
							continue
						}
						if !spl.IsValidCard(spl.SwitchToCardData(k)) {
							continue
						}
						if !spl.IsValidCard(spl.SwitchToCardData(m)) {
							continue
						}

						cbTempCard[i]++
						cbTempCard[j]++
						cbTempCard[k]++
						cbTempCard[m]++
						//分析记录清理
						wChiHuKind |= spl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount, true)
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
		//这里只有带癞子的胡牌类型
		if (wChiHuKind & static.CHK_DA_HU_NOMAGIC) != 0 {
			wChiHuKind ^= static.CHK_DA_HU_NOMAGIC
		}
		if (wChiHuKind & static.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind ^= static.CHK_PING_HU_NOMAGIC
		}

		wChiHuKind |= static.CHK_PING_HU_MAGIC
	}
	return wChiHuKind
}

//吃胡判断
func (spl *SportLogicJZCXZ) AnalyseChiHuCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16, ChiHuResult *static.TagChiHuResult) byte {
	//变量定义
	wChiHuKind := uint64(static.CHK_NULL)
	//构造麻将
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex)
	if cbCurrentCard != 0 {
		cbCardIndexTemp[spl.SwitchToCardIndex(cbCurrentCard)]++
	}

	//结果判断硬胡
	wChiHuKind |= spl.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbWeaveCount, false)
	wMagicCount := spl.GetMagicCount(cbCardIndexTemp)

	//如果硬胡能胡，再考虑软胡（可能存在又有硬胡又有软胡的情况，这种情况下从业务上就能保证硬胡的分数肯定是多一些的，所以不需要判断有硬胡时又有软胡的情况）
	if wChiHuKind == static.CHK_NULL {
		if wMagicCount > 0 && wMagicCount <= 1 {
			wChiHuRight_magic := wChiHuRight
			wChiHuKind_magic := uint64(static.CHK_NULL)
			wChiHuKind_magic |= spl.AnalyseMagicHuCard(cbCardIndexTemp, WeaveItem, cbWeaveCount, cbCurrentCard, wChiHuRight_magic)
			wChiHuKind = wChiHuKind_magic
			wChiHuRight = wChiHuRight_magic
		}
	}
	if wChiHuKind == static.CHK_NULL {
		ChiHuResult.ChiHuKind = static.CHK_NULL
		ChiHuResult.ChiHuRight = static.CHK_NULL
		return static.CHK_NULL
	} else {
		ChiHuResult.ChiHuKind = wChiHuKind
		ChiHuResult.ChiHuRight = wChiHuRight
		return static.WIK_CHI_HU
	}
}

// 听牌判断,当手上是13张牌的时候,分析需要一张什么牌胡牌
func (spl *SportLogicJZCXZ) AnalyseTingCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) bool {
	bIsTing := false
	//变量定义
	var ChiHuResult static.TagChiHuResult
	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]
	//听牌分析
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//胡牌分析
		cbCurrentCard := spl.SwitchToCardData(i)
		cbHuCardKind := spl.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult)
		//结果判断
		if cbHuCardKind != static.CHK_NULL {
			bIsTing = true
		}
	}
	return bIsTing
}

//听牌判断,判断手上听多少张牌，0张表示没听牌
func (spl *SportLogicJZCXZ) AnalyseTingCardCount(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) byte {
	iTingCount := 0
	//变量定义
	var ChiHuResult static.TagChiHuResult
	//构造扑克
	cbCardIndexTemp := cbCardIndex[:]
	//听牌分析
	y := 0
	for i := byte(0); i < static.MAX_INDEX; i++ {
		//胡牌分析
		cbCurrentCard := spl.SwitchToCardData(i)
		cbHuCardKind := spl.AnalyseChiHuCard(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard, wChiHuRight, &ChiHuResult)
		//结果判断
		if cbHuCardKind != static.CHK_NULL { //赖子
			y++ //计数
		}
	}

	iTingCount = y
	return byte(iTingCount)
}

//碰牌判断
func (spl *SportLogicJZCXZ) JZ_EstimatePengCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	if !spl.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}

	//普通碰牌判断(荆州搓虾子是个两个皮子算杠，不算碰)
	if cbCardIndex[spl.SwitchToCardIndex(cbCurrentCard)] >= 2 && spl.PiZiCard != cbCurrentCard {
		return static.WIK_PENG
	}

	return static.WIK_NULL
}

//杠牌判断
func (spl *SportLogicJZCXZ) JZ_EstimateGangCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	if !spl.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}

	if spl.PiZiCard != cbCurrentCard {
		//杠牌判断
		if cbCardIndex[spl.SwitchToCardIndex(cbCurrentCard)] == 3 {
			return static.WIK_GANG
		}
	} else {
		//荆州搓虾子是两个皮子的话算杠（仅仅界面显示，实际不是杠，不补牌），不算碰
		if cbCardIndex[spl.SwitchToCardIndex(cbCurrentCard)] == 2 {
			return static.WIK_GANG
		}
	}

	return static.WIK_NULL
}

func (spl *SportLogicJZCXZ) OneTingCard(cbCardIndex []byte, index byte) byte {
	if cbCardIndex[index] == 4 {
		//fmt.Println(fmt.Sprintf("手牌中（%s）已经4章了",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
		return static.INVALID_BYTE
	}
	if index == 0 || index == 9 || index == 18 {
		//1万 1条 1筒
		if cbCardIndex[index] == 0 {
			if cbCardIndex[index+2] != 0 && cbCardIndex[index+1] != 0 {
				//fmt.Println(fmt.Sprintf("1序列牌可成刻（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
				return index
			} else {
				return static.INVALID_BYTE
			}
		} else {
			//fmt.Println(fmt.Sprintf("1序列牌可成对（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		}
	}
	if index == 8 || index == 17 || index == 26 {
		//9万 9条 9筒
		if cbCardIndex[index] == 0 {
			if cbCardIndex[index-1] != 0 && cbCardIndex[index-2] != 0 {
				//fmt.Println(fmt.Sprintf("9序列牌可成刻（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
				return index
			} else {
				return static.INVALID_BYTE
			}
		} else {
			//fmt.Println(fmt.Sprintf("9序列牌可成对（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		}
	}
	if index == 1 || index == 10 || index == 19 {
		//2万 2条 2筒
		if cbCardIndex[index] == 0 {
			if (cbCardIndex[index-1] != 0 && cbCardIndex[index+1] != 0) || (cbCardIndex[index+2] != 0 && cbCardIndex[index+1] != 0) {
				//fmt.Println(fmt.Sprintf("2序列牌可成刻（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
				return index
			} else {
				return static.INVALID_BYTE
			}
		} else {
			//fmt.Println(fmt.Sprintf("2序列牌可成对（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		}
	}
	if index == 7 || index == 16 || index == 25 {
		//8万 8条 8筒 的检查
		if cbCardIndex[index] == 0 {
			if (cbCardIndex[index-1] != 0 && cbCardIndex[index+1] != 0) || (cbCardIndex[index-2] != 0 && cbCardIndex[index+1] != 0) {
				//fmt.Println(fmt.Sprintf("8序列牌可成刻（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
				return index
			} else {
				return static.INVALID_BYTE
			}
		} else {
			//fmt.Println(fmt.Sprintf("8序列牌可成对（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		}
	}
	//风
	if index > 26 && index < 31 {
		if cbCardIndex[index] == 0 {
			return static.INVALID_BYTE
		} else {
			//fmt.Println(fmt.Sprintf("风牌可成对（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		}
	}
	//箭
	if index > 30 && index < 34 {
		if cbCardIndex[index] == 0 {
			return static.INVALID_BYTE
		} else {
			//fmt.Println(fmt.Sprintf("箭牌（%s）可成对",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		}
	}
	//正常牌
	if cbCardIndex[index] == 0 {
		if cbCardIndex[index-1] != 0 && cbCardIndex[index+1] != 0 || cbCardIndex[index-1] != 0 && cbCardIndex[index-2] != 0 || cbCardIndex[index+1] != 0 && cbCardIndex[index+2] != 0 {
			//fmt.Println(fmt.Sprintf("正常牌可成刻（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
			return index
		} else {
			return static.INVALID_BYTE
		}
	} else {
		//fmt.Println(fmt.Sprintf("正常牌可成对（%s）",spl.SwitchToCardNameByData(common.IndexToCard(index), 1)))
		return index
	}
	return static.INVALID_BYTE
}

// 检查是不是能碰杠的牌
func (spl *SportLogicJZCXZ) IsPengGangCard(card byte) bool {
	if card == spl.MagicCard || spl.IsWindCard(card) {
		return false
	}
	return true
}

// 风牌效验
func (spl *SportLogicJZCXZ) IsWindCard(cbCardData ...byte) bool {
	for _, card := range cbCardData {
		cbValue := (card & static.MASK_VALUE)
		cbColor := (card & static.MASK_COLOR) >> 4
		if !(cbValue >= 1 && cbValue <= 7 && cbColor == 3) {
			return false
		}
	}
	return spl.IsValidCard(cbCardData...)
}
