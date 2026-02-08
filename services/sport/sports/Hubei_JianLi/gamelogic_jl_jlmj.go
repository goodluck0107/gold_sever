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
	"fmt"
	public "github.com/open-source/game/chess.git/pkg/static"
	common "github.com/open-source/game/chess.git/services/sport/backboard"
	modules "github.com/open-source/game/chess.git/services/sport/components"
	logic "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
)

// ////////////////////////////////////////////////////////////////////////
const (
	//游戏 I D
	//GAME_NAME  = "监利麻将"                               //游戏名字
	GAME_JLMJ_GENRE = (public.GAME_GENRE_GOLD | public.GAME_GENRE_MATCH) //游戏类型
)

// 牌型设计
var jlmj_strCardsMessage = [public.MAX_INDEX]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09",
	"0x11", "0x12", "0x13", "0x14", "0x15", "0x16", "0x17", "0x18", "0x19",
	"0x21", "0x22", "0x23", "0x24", "0x25", "0x26", "0x27", "0x28", "0x29",
	"0x31", "0x32", "0x33", "0x34", "0x35", "0x36", "0x37",
}
var jlmj_strCardsMessage1 = [public.MAX_INDEX]string{
	"1万", "2万", "3万", "4万", "5万", "6万", "7万", "8万", "9万",
	"1条", "2条", "3条", "4条", "5条", "6条", "7条", "8条", "9条",
	"1同", "2同", "3同", "4同", "5同", "6同", "7同", "8同", "9同",
	"东风", "南风", "西风", "北风", "红中", "发财", "白板",
}

// 扑克数据
var jlmj_cbCardDataArray = [public.MAX_REPERTORY]byte{
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

	0x31, 0x31, 0x31, 0x31, //东风
	0x32, 0x32, 0x32, 0x32, //南风
	0x33, 0x33, 0x33, 0x33, //西风
	0x34, 0x34, 0x34, 0x34, //北风
	0x35, 0x35, 0x35, 0x35, //中
	0x36, 0x36, 0x36, 0x36, //发
	0x37, 0x37, 0x37, 0x37, //白
}

// 补花类型
const (
	BuHua_Type_NULL       = 0
	BuHua_Type_HongZhong  = 1
	BuHua_Type_ZhongFaBai = 2
)

// 胡法类型
const (
	HuFa_Type_Zimo    = 1
	HuFa_Type_Dianpao = 2
)

type GameLogic_jlmj struct {
	logic.BaseLogic
	cardsclass int    //牌型，用于生成牌库和牌的有效型判断
	m_cbCard   []byte //每大局的牌库可能不同,通山晃晃都一样
}

// 20200314 沈强 本身就没有风牌
func (self *GameLogic_jlmj) SetCardClass(special int) {
	//self.cardsclass = common.SetCardsClass(common.CARDS_WITHOUT_DRAGON, special)
	self.cardsclass = common.SetCardsClass_ex(common.CARDS_WITHOUT_WIND, special)
}
func (self *GameLogic_jlmj) RandCardData() (byte, []byte) {
	// fmt.Println(fmt.Sprintf("混乱牌型class（%d)", self.Cardsclass))
	//if len(self.m_cbCard) == 0 {
	err := self.CreateCards()
	if err != nil {
		fmt.Println(fmt.Sprintf("重建牌库失败,%v", err))
	}
	//}
	//cardsIndex := common.CardsToCardIndex(self.m_cbCard)
	//common.Print_cards(cardsIndex[:])
	return common.RandCardData(self.m_cbCard)
}

// 20181119 创建牌库的牌
func (self *GameLogic_jlmj) CreateCards() (err error) {
	//预备替换
	// fmt.Println(fmt.Sprintf("初始牌型class（%d)", self.Cardsclass))
	err, self.m_cbCard = common.CreateCards(self.cardsclass)
	if err != nil {
		return err
	}
	//cardsIndex := common.CardsToCardIndex(self.m_cbCard)
	// fmt.Println(fmt.Sprintf("创建牌库的列表数据"))
	//common.Print_cards(cardsIndex[:])
	return nil
}

////4人混乱扑克
//func (self *GameLogic_jlmj) RandCardData4() (byte, []byte) {
//	//混乱准备
//	//监利麻将112张（带红中）和120张（带中发白）
//	if self.Rule.BuHuaType == BuHua_Type_NULL {
//		syslog.Logger().Errorln(fmt.Sprintf("混乱扑克错误,RandCardData4，4人玩法没有设置补花类型"))
//		return 0,nil
//	}
//	cbCardData := [public.MAX_REPERTORY - 4*4]byte{} //去掉风
//	cbCardDataTemp := []byte{}
//	var cbMaxCount byte = public.MAX_REPERTORY - 4*4
//	if self.Rule.BuHuaType == BuHua_Type_HongZhong{
//		cbMaxCount -= 4*2
//		cbCardDataTemp = append(cbCardDataTemp, jlmj_cbCardDataArray[:4*9*3]...) //拷贝万条筒
//		cbCardDataTemp = append(cbCardDataTemp, jlmj_cbCardDataArray[4*9*3+4*4:4*9*3+4*5]...) //拷贝红中
//	}else{
//		cbCardDataTemp = append(cbCardDataTemp, jlmj_cbCardDataArray[:4*9*3]...) //拷贝万条筒
//		cbCardDataTemp = append(cbCardDataTemp, jlmj_cbCardDataArray[4*9*3+4*4:4*9*3+4*7]...) //拷贝中发白
//	}
//
//	//混乱扑克
//	cbRandCount, cbPosition := 0, 0
//	randTmp := 0
//	nAccert := 0
//	for {
//		nAccert++
//		if nAccert > 200 {
//			break
//		}
//		randTmp = int(cbMaxCount) - cbRandCount - 1
//		if randTmp > 0 {
//			cbPosition = rand.Intn(randTmp)
//		} else {
//			cbPosition = 0
//		}
//		cbCardData[cbRandCount] = cbCardDataTemp[cbPosition]
//		cbRandCount++
//		cbCardDataTemp[cbPosition] = cbCardDataTemp[int(cbMaxCount)-cbRandCount]
//		if cbRandCount >= int(cbMaxCount) {
//			break
//		}
//	}
//
//	syslog.Logger().Errorln(cbCardData)
//
//	return cbMaxCount, cbCardData[:]
//}
//
////三人混乱扑克,三人玩法用牌：万条筒或者条筒
//func (self *GameLogic_jlmj) RandCardData3() (byte, []byte){
//	cbCardData := [public.MAX_REPERTORY - 7*4]byte{}
//	cbCardDataTemp := make([]byte, public.MAX_REPERTORY-7*4, public.MAX_REPERTORY-7*4)
//	var cbMaxCount byte = public.MAX_REPERTORY - 7*4
//
//	if self.Rule.NoWan {
//		cbMaxCount -= 4*9	//去万
//		copy(cbCardDataTemp,jlmj_cbCardDataArray[4*9:public.MAX_REPERTORY-7*4])
//	}else{
//		copy(cbCardDataTemp, jlmj_cbCardDataArray[:public.MAX_REPERTORY-7*4])
//	}
//
//	//混乱扑克
//	cbRandCount, cbPosition := 0, 0
//	randTmp := 0
//	nAccert := 0
//	for {
//		nAccert++
//		if nAccert > 200 {
//			break
//		}
//		randTmp = int(cbMaxCount) - cbRandCount - 1
//		if randTmp > 0 {
//			cbPosition = rand.Intn(randTmp)
//		} else {
//			cbPosition = 0
//		}
//		cbCardData[cbRandCount] = cbCardDataTemp[cbPosition]
//		cbRandCount++
//		cbCardDataTemp[cbPosition] = cbCardDataTemp[int(cbMaxCount)-cbRandCount]
//		if cbRandCount >= int(cbMaxCount) {
//			break
//		}
//	}
//
//	syslog.Logger().Errorln(cbCardData)
//
//	return cbMaxCount, cbCardData[:]
//}

// 有效判断
func (self *GameLogic_jlmj) IsValidCard(cbCardData byte) bool {
	cbValue := (cbCardData & public.MASK_VALUE)
	cbColor := (cbCardData & public.MASK_COLOR) >> 4

	//校验：万条同 +中发白
	bIsValid := ((cbValue >= 1) && (cbValue <= 9) && (cbColor <= 2)) || ((cbValue >= 5) && (cbValue <= 7) && (cbColor == 3))
	if !bIsValid {
		return false //非法牌
	}

	if self.Rule.NoWan {
		if cbColor == 0 {
			return false
		}
	}

	return true
}

// 吃牌判断
func (self *GameLogic_jlmj) EstimateEatCard(cbCardIndex [public.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	if !self.IsValidCard(cbCurrentCard) {
		//TODO
		return public.WIK_NULL
	}

	////过滤判断,咸宁晃晃没有中发白,不需要做这个过滤判断,只需要上面那个检验有效牌就可以了
	//if cbCurrentCard >= 0x30 {
	//	return public.WIK_NULL
	//}

	//变量定义
	cbExcursion := [3]byte{0, 1, 2}
	cbItemKind := [3]byte{public.WIK_LEFT, public.WIK_CENTER, public.WIK_RIGHT}

	//吃牌判断
	var cbEatKind, cbFirstIndex byte = 0, 0
	var cbCurrentIndex byte = self.SwitchToCardIndex(cbCurrentCard)
	cbMgicCardIndex := self.SwitchToCardIndex(self.MagicCard)
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

			if i == 0 && (cbMgicCardIndex == cbCurrentIndex+1 || cbMgicCardIndex == cbCurrentIndex+2) {
				continue
			} else if i == 1 && (cbMgicCardIndex == cbCurrentIndex-1 || cbMgicCardIndex == cbCurrentIndex+1) {
				continue
			} else if i == 2 && (cbMgicCardIndex == cbCurrentIndex-1 || cbMgicCardIndex == cbCurrentIndex-2) {
				continue
			}

			//设置类型
			cbEatKind |= cbItemKind[i]
		}
	}

	return cbEatKind
}

// 杠牌分析
func (self *GameLogic_jlmj) AnalyseGangCard(_userItem *modules.Player, cbCardIndex [public.MAX_INDEX]byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte, GangCardResult *public.TagGangCardResult) byte {
	//设置变量
	cbActionMask := public.WIK_NULL

	//手上杠牌
	for i := byte(0); i < public.MAX_INDEX; i++ {
		//中发白花牌
		if i == self.SwitchToCardIndex(0x35) || i == self.SwitchToCardIndex(0x36) || i == self.SwitchToCardIndex(0x37) {
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
				if self.IsGiveUpGang(_userItem, WeaveItem[i].CenterCard) {
					break
				}
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
func (self *GameLogic_jlmj) AnalyseHuKind(wChiHuRight uint16, wChiHuKind uint64, cbTempCard []byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte) uint64 {

	_, AnalyseItemArray := self.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//牌型分析
		wChiHuKind |= public.CHK_PING_HU_NOMAGIC
	}

	// 杠上开花判断
	if self.Rule.AllowGang && self.HuType.HAVE_GANG_SHANG_KAI_HUA && (public.CHR_GANG_SHANG_KAI_HUA&wChiHuRight) != 0 {
		if (wChiHuKind & public.CHK_PING_HU_NOMAGIC) != 0 {
			wChiHuKind |= public.CHK_GANG_SHANG_KAI_HUA
		}
	}

	return wChiHuKind
}

// 带赖子的胡牌
func (self *GameLogic_jlmj) AnalyseMagicHuCard(cbCardIndex []byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16) uint64 {

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

			wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)

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
				wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbTempCard, WeaveItem, cbWeaveCount)
				cbTempCard[i]--
				cbTempCard[j]--
			}
		}
	} else if wMagicCount == 3 {
		byDisCount, byNeedMagicCount, abyDisTable := self.CreatDisperseTable(cbTempCard)
		if false == self.DisperseEstimate(byDisCount, byNeedMagicCount, byte(wMagicCount)) {
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

	if wChiHuKind != public.CHK_NULL {
		//这里只有带癞子的胡牌类型
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
func (self *GameLogic_jlmj) AnalyseChiHuCard(cbCardIndex []byte, WeaveItem []public.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight uint16, ChiHuResult *public.TagChiHuResult) byte {
	//变量定义
	wChiHuKind := uint64(public.CHK_NULL)
	//构造麻将
	cbCardIndexTemp := make([]byte, public.MAX_INDEX, public.MAX_INDEX)

	copy(cbCardIndexTemp, cbCardIndex)

	if cbCurrentCard != 0 {
		cbCardIndexTemp[self.SwitchToCardIndex(cbCurrentCard)]++
	}

	//手上有花牌不能胡牌
	if self.GetFlowerCardCount(cbCardIndexTemp) > 0 {
		return public.WIK_NULL
	}
	//结果判断硬胡
	wChiHuKind |= self.AnalyseHuKind(wChiHuRight, wChiHuKind, cbCardIndexTemp, WeaveItem, cbWeaveCount)
	//wMagicCount := self.GetMagicCount(cbCardIndexTemp)

	//if wChiHuKind != public.CHK_NULL {
	//	//if wMagicCount > 1  && (wChiHuRight & public.CHR_GANG_SHANG_KAI_HUA) == 0 {
	//	//
	//	//}
	//}

	////结果判断软胡
	//if wMagicCount > 0 && wMagicCount <= 4 {
	//	wChiHuKind_magic := uint64(public.CHK_NULL)
	//	wChiHuKind_magic |= self.AnalyseMagicHuCard(cbCardIndexTemp, WeaveItem, cbWeaveCount, cbCurrentCard, wChiHuRight)
	//
	//	//有软胡
	//	if wChiHuKind_magic != public.CHK_NULL {
	//		wChiHuKind |= wChiHuKind_magic
	//	}
	//}
	//
	////既有硬胡有有软胡,算硬胡
	//if (wChiHuKind&public.CHK_PING_HU_MAGIC != 0) && (wChiHuKind&public.CHK_PING_HU_NOMAGIC != 0) {
	//	wChiHuKind ^= public.CHK_PING_HU_MAGIC
	//}

	ChiHuResult.ChiHuKind = wChiHuKind
	if ChiHuResult.ChiHuKind != public.CHK_NULL {
		return public.WIK_CHI_HU
	}

	return public.WIK_NULL
}

// 听牌判断,判断手上听多少张牌，0张表示没听牌
func (self *GameLogic_jlmj) AnalyseTingCardCount(cbCardIndex []byte, WeaveItem []public.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) byte {
	iTingCount := 0
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
		if cbHuCardKind != public.CHK_NULL { //赖子
			y++ //计数
		}
	}

	iTingCount = y

	return byte(iTingCount)
}

// 是否为弃杠牌
func (self *GameLogic_jlmj) IsGiveUpGang(_userItem *modules.Player, card byte) bool {
	for _, c := range _userItem.Ctx.VecGangCard {
		if c == card {
			return true
		}
	}
	return false
}

// 追加一个弃杠牌
func (self *GameLogic_jlmj) AppendGiveUpGang(_userItem *modules.Player, card byte) {
	if self.IsGiveUpGang(_userItem, card) {
		return
	}
	_userItem.Ctx.VecGangCard = append(_userItem.Ctx.VecGangCard, card)
}

// 获取花牌数量
func (self *GameLogic_jlmj) GetFlowerCardCount(cbCardIndex []byte) byte {
	FlowerCardCount := byte(0)
	if self.Rule.BuHuaType == BuHua_Type_HongZhong {
		FlowerCardCount = cbCardIndex[self.SwitchToCardIndex(0x35)]
	} else if self.Rule.BuHuaType == BuHua_Type_ZhongFaBai {
		FlowerCardCount = cbCardIndex[self.SwitchToCardIndex(0x35)] + cbCardIndex[self.SwitchToCardIndex(0x36)] + cbCardIndex[self.SwitchToCardIndex(0x37)]
	}

	return FlowerCardCount
}

// 当前牌是否是花牌
func (self *GameLogic_jlmj) IsFlowerCard(cbCard byte) bool {
	if self.Rule.BuHuaType == BuHua_Type_HongZhong {
		return cbCard == 0x35
	} else if self.Rule.BuHuaType == BuHua_Type_ZhongFaBai {
		return cbCard == 0x35 || cbCard == 0x36 || cbCard == 0x37
	}

	return false
}

// 多张补花牌判断
func (self *GameLogic_jlmj) IsFlowerCardM(cbCardData ...byte) bool {
	for _, card := range cbCardData {
		if !self.IsFlowerCard(card) {
			return false
		}
	}
	return self.IsValidCardM(cbCardData...)
}

// 效验牌
func (self *GameLogic_jlmj) IsValidCardM(cbCardData ...byte) bool {
	for _, card := range cbCardData {
		if !self.IsValidCard(card) {
			return false // 非法牌
		}
	}
	return true
}
