//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package logic

//import "fmt"

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"strconv"
)

//////////////////////////////////////////////////////////////////////////
const (
	GAME_GENRE = (static.GAME_GENRE_GOLD | static.GAME_GENRE_MATCH) //游戏类型

	CARD_TYPE_BACK = 0xF0 //麻将牌背
)

//////////////////////////////////////////////////////////////////////////

const (
	CARD_DONGFENG  = 0x31
	CARD_NANFENG   = 0x32
	CARD_XIFENG    = 0x33
	CARD_BEIFENG   = 0x34
	CARD_HONGZHONG = 0x35
	CARD_FACAI     = 0x36
	CARD_BAIBAN    = 0x37
)

const (
	Contractor_NULL       = 0
	Contractor_XiaYu      = 1
	Contractor_QingYiSe   = 2
	Contractor_QiangGang  = 3
	Contractor_QuanQiuRen = 4
	Contractor_XiaXiaoYu  = 5
	Contractor_XiaDaYu    = 6
	Contractor_JiangYiSe  = 7
	Contractor_FengYiSe   = 8
	Contractor_PeiBao     = 9
)

const (
	Gang_Type_FaCaiGang = 1
	Gang_Type_PiZiGang  = 2
	Gang_Type_LaiZiGang = 3
)

//////////////////////////////////////////////////////////////////////////

type BaseLogic struct {
	//	m_cbCardDataArray  [MAX_REPERTORY]byte //扑克数据
	MagicCard    byte   //赖子牌值
	PiZiCard     byte   //皮子牌值
	PiZiCards    []byte //皮子杠 红中杠 发财杠保存在这里
	Specialcards []byte //20191009 通山麻将自动出牌不能复用上面的m_bPiZiCards
	//	m_strCardsMessage  [public.MAX_INDEX]string
	//	m_strCardsMessage1 [public.MAX_INDEX]string

	Rule   rule2.St_FriendRule
	HuType static.TagHuType //胡牌配置
}

//设置赖子值
func (self *BaseLogic) SetMagicCard(wMagicCard byte) {
	self.MagicCard = wMagicCard
}

//设置皮子值
func (self *BaseLogic) SetPiZiCard(wPiZiCard byte) {
	self.PiZiCard = wPiZiCard
}

//设置皮子值
func (self *BaseLogic) SetPiZiCards(wPiZiCards []byte) {
	self.PiZiCards = wPiZiCards
}

//设置皮子值
func (self *BaseLogic) AddPiZiCard(wPiZiCards byte) {
	if len(self.PiZiCards) == 0 {
		self.PiZiCards = []byte{wPiZiCards}
	} else {
		self.PiZiCards = append(self.PiZiCards, wPiZiCards)
	}
}

func (self *BaseLogic) GetPiZiCount(cbCardIndex []byte) byte {
	if self.PiZiCard > 0x37 || self.PiZiCard == static.INVALID_BYTE {
		return 0
	}
	return cbCardIndex[self.SwitchToCardIndex(self.PiZiCard)]
}

func (self *BaseLogic) GetMagicCount(cbCardIndex []byte) byte {
	if self.MagicCard > 0x37 || self.MagicCard == static.INVALID_BYTE {
		return 0
	}
	return cbCardIndex[self.SwitchToCardIndex(self.MagicCard)]
}

func (self *BaseLogic) GetHongZhongCount(cbCardIndex []byte) byte {
	return cbCardIndex[self.SwitchToCardIndex(0x35)]
}

func (self *BaseLogic) CreateCards() (byte, []byte) {
	cbCardDataTemp := make([]byte, static.MAX_REPERTORY, static.MAX_REPERTORY)
	_maxCount := byte(static.MAX_REPERTORY)
	if self.Rule.NoWan {
		_maxCount -= 4 * 9
		copy(cbCardDataTemp, CardDataArray[4*9:])

	} else {
		copy(cbCardDataTemp, CardDataArray[:])
	}

	return _maxCount, cbCardDataTemp
}

//混乱扑克
func (self *BaseLogic) RandCardData() (byte, []byte) {
	//CopyMemory(cbCardData,CardDataArray,sizeof(CardDataArray));
	_, cbRepertoryCard := mahlib2.CreateCards(self.Rule.Cardsclass)

	return mahlib2.RandCardData(cbRepertoryCard)

	////混乱准备
	//cbCardData := [public.MAX_REPERTORY]byte{}
	//cbCardDataTemp := make([]byte, public.MAX_REPERTORY, public.MAX_REPERTORY)
	//var cbMaxCount = byte(len(cbRepertoryCard)) //public.MAX_REPERTORY
	//if self.Rule.NoWan {
	//	cbMaxCount -= 4 * 9 //去万字牌时，少4*9张牌
	//	//CopyMemory(cbCardDataTemp, CardDataArray + 4*9, cbMaxCount); //不copy前面的4*9张万字牌
	//	copy(cbCardDataTemp, CardDataArray[4*9:])
	//} else {
	//	//CopyMemory(cbCardDataTemp,CardDataArray,sizeof(CardDataArray));
	//	copy(cbCardDataTemp, CardDataArray[:])
	//}
	//
	//copy(cbCardDataTemp, cbRepertoryCard)
	//
	//syslog.Logger().Debug("RandCardData tempcard :")
	//syslog.Logger().Debug(cbCardDataTemp)
	//
	////混乱扑克
	//cbRandCount, cbPosition := 0, 0
	//randTmp := 0
	//nAccert := 0
	//for {
	//	nAccert++
	//	if nAccert > 200 {
	//		//			m_mylog.Log("混乱扑克时死循环啦")
	//		break
	//	}
	//	randTmp = int(cbMaxCount) - cbRandCount - 1
	//	if randTmp > 0 {
	//		cbPosition = rand.Intn(randTmp)
	//	} else {
	//		cbPosition = 0
	//	}
	//	//cbPosition=rand()%(cbMaxCount-cbRandCount);
	//	cbCardData[cbRandCount] = cbCardDataTemp[cbPosition]
	//	cbRandCount++
	//	cbCardDataTemp[cbPosition] = cbCardDataTemp[int(cbMaxCount)-cbRandCount]
	//	if cbRandCount >= int(cbMaxCount) {
	//		break
	//	}
	//}
	//
	//syslog.Logger().Debug(cbCardData)
	//
	//return cbMaxCount, cbCardData[:]
}

//20200520 苏大强 混乱扑克加次数
func (self *BaseLogic) RandCardData_ex(num int) (byte, []byte) {
	_, cbRepertoryCard := mahlib2.CreateCards(self.Rule.Cardsclass)
	if num == 0 {
		num = 1
	}
	return self.RandCardData_recursion(num, cbRepertoryCard)
}
func (self *BaseLogic) RandCardData_recursion(randCardnum int, baseCards []byte) (byte, []byte) {
	if randCardnum == 0 {
		return byte(len(baseCards)), baseCards
	}
	cardnum, temp1 := mahlib2.RandCardData(baseCards)
	xlog.Logger().Debug(fmt.Sprintf("混乱扑克倒数（%d）:牌库牌数（%d）数据（%v）", randCardnum, cardnum, temp1))
	return self.RandCardData_recursion(randCardnum-1, temp1)
}

//删除单张扑克
func (self *BaseLogic) RemoveCard(cbCardIndex [static.MAX_INDEX]byte, cbRemoveCard byte) (bool, [static.MAX_INDEX]byte) {
	//效验扑克

	xlog.Logger().Debug("RemoveCard card :" + strconv.Itoa(int(cbRemoveCard)))
	xlog.Logger().Debug(cbCardIndex)
	if !self.IsValidCard(cbRemoveCard) {
		xlog.Logger().Errorln("RemoveCard inValidCard")
		//TODO
		return false, cbCardIndex
	}
	//效验扑克
	if cbCardIndex[self.SwitchToCardIndex(cbRemoveCard)] <= 0 {
		//TODO
		xlog.Logger().Errorln("RemoveCard inValidCard")
		return false, cbCardIndex
	}

	cbRemoveIndex := self.SwitchToCardIndex(cbRemoveCard)

	if cbCardIndex[cbRemoveIndex] > 0 {
		cbCardIndex[cbRemoveIndex]--

		xlog.Logger().Debug(cbCardIndex)
		return true, cbCardIndex
	}
	xlog.Logger().Errorln("RemoveCard inValidCard")
	return false, cbCardIndex
}

//删除多张扑克
func (self *BaseLogic) RemoveCard2(cbCardIndex [static.MAX_INDEX]byte, cbRemoveCard []byte, cbRemoveCount byte) (bool, [static.MAX_INDEX]byte) {
	//删除扑克
	xlog.Logger().Debug("RemoveCard2 card :")
	xlog.Logger().Debug(cbRemoveCard)
	xlog.Logger().Debug(cbCardIndex)
	for i := 0; i < int(cbRemoveCount); i++ {
		//效验扑克
		if !self.IsValidCard(cbRemoveCard[i]) {
			//TODO
			continue
		}
		//效验扑克
		if cbCardIndex[self.SwitchToCardIndex(cbRemoveCard[i])] <= 0 {
			//TODO
			continue
		}

		//删除扑克
		cbRemoveIndex := self.SwitchToCardIndex(cbRemoveCard[i])
		if cbCardIndex[cbRemoveIndex] == 0 {
			//还原删除
			for j := 0; j < i; j++ {
				//ASSERT(IsValidCard(cbRemoveCard[j]));
				cbCardIndex[self.SwitchToCardIndex(cbRemoveCard[j])]++
			}

			return false, cbCardIndex
		} else {
			//删除扑克
			cbCardIndex[cbRemoveIndex]--
		}
	}
	xlog.Logger().Debug(cbCardIndex)
	return true, cbCardIndex
}

//删除固定数量手牌中的多个扑克
func (self *BaseLogic) RemoveCard3(cbCardData [static.MAX_INDEX]byte, cbCardCount int, cbRemoveCard []byte, cbRemoveCount int) (bool, [static.MAX_INDEX]byte) {
	//检验数据
	if cbCardCount > 14 {
		//TODO
		return false, cbCardData
	}
	if cbRemoveCount > cbCardCount {
		//TODO
		return false, cbCardData
	}

	//定义变量
	cbDeleteCount := 0
	cbTempCardData := make([]byte, 14, 14)
	if cbCardCount > len(cbTempCardData) {
		return false, cbCardData
	}
	//cbTempCardData = cbCardData[:cbCardCount]
	copy(cbTempCardData, cbCardData[:cbCardCount])

	//置零扑克
	for i := 0; i < cbRemoveCount; i++ {
		for j := 0; j < cbCardCount; j++ {
			if cbRemoveCard[i] == cbTempCardData[j] {
				cbDeleteCount++
				cbTempCardData[j] = 0
				break
			}
		}
	}

	//成功判断
	if cbDeleteCount != cbRemoveCount {
		//ASSERT(FALSE);
		return false, cbCardData
	}

	//清理扑克
	cbCardPos := 0
	for i := 0; i < cbCardCount; i++ {
		if cbTempCardData[i] != 0 {
			cbCardData[cbCardPos] = cbTempCardData[i]
			cbCardPos++
		}
	}
	return true, cbCardData
}

//有效判断
func (self *BaseLogic) IsValidCard(cbCardData byte) bool {
	cbValue := (cbCardData & static.MASK_VALUE)
	cbColor := (cbCardData & static.MASK_COLOR) >> 4

	//校验：万条同+东南西北中发白
	bIsValid := (((cbValue >= 1) && (cbValue <= 9) && (cbColor <= 2)) || ((cbValue >= 1) && (cbValue <= 7) && (cbColor == 3)))
	if !bIsValid {
		return false //非法牌
	}

	//校验：去万
	if self.Rule.NoWan && cbColor == 0 {
		return false
	}

	return true //(((cbValue>=1)&&(cbValue<=9)&&(cbColor<=2))||((cbValue>=1)&&(cbValue<=7)&&(cbColor==3)));
}

//扑克数目
func (self *BaseLogic) GetCardCount(cbCardIndex []byte) byte {
	//数目统计
	cbCardCount := 0
	for i := 0; i < static.MAX_INDEX; i++ {
		cbCardCount += int(cbCardIndex[i])
	}
	return byte(cbCardCount)
}

//获取组合，为了安全，添加个东西 苏大强
func (self *BaseLogic) GetWeaveCard(cbWeaveKind byte, cbCenterCard byte, cbCardBuffer [4]byte) (byte, [4]byte) {
	//组合扑克
	switch cbWeaveKind {
	case static.WIK_LEFT: //上牌操作
		{
			//设置变量
			cbCardBuffer[0] = cbCenterCard
			cbCardBuffer[1] = cbCenterCard + 1
			cbCardBuffer[2] = cbCenterCard + 2
			cbCardBuffer[3] = static.INVALID_BYTE
			return 3, cbCardBuffer
		}
	case static.WIK_RIGHT: //上牌操作
		{
			//设置变量
			cbCardBuffer[0] = cbCenterCard - 2
			cbCardBuffer[1] = cbCenterCard - 1
			cbCardBuffer[2] = cbCenterCard
			cbCardBuffer[3] = static.INVALID_BYTE
			return 3, cbCardBuffer
		}
	case static.WIK_CENTER: //上牌操作
		{
			//设置变量
			cbCardBuffer[0] = cbCenterCard - 1
			cbCardBuffer[1] = cbCenterCard
			cbCardBuffer[2] = cbCenterCard + 1
			cbCardBuffer[3] = static.INVALID_BYTE
			return 3, cbCardBuffer
		}
	case static.WIK_PENG: //碰牌操作
		{
			//设置变量
			cbCardBuffer[0] = cbCenterCard
			cbCardBuffer[1] = cbCenterCard
			cbCardBuffer[2] = cbCenterCard
			cbCardBuffer[3] = static.INVALID_BYTE
			return 3, cbCardBuffer
		}
	case static.WIK_FILL: //补牌操作
		{
			//设置变量
			cbCardBuffer[0] = cbCenterCard
			cbCardBuffer[1] = cbCenterCard
			cbCardBuffer[2] = cbCenterCard
			cbCardBuffer[3] = cbCenterCard

			return 4, cbCardBuffer
		}
	case static.WIK_GANG: //杠牌操作
		{
			//设置变量
			cbCardBuffer[0] = cbCenterCard
			cbCardBuffer[1] = cbCenterCard
			cbCardBuffer[2] = cbCenterCard
			cbCardBuffer[3] = cbCenterCard

			return 4, cbCardBuffer
		}
	default:
		{
			//ASSERT(FALSE);
		}
	}

	return 0, cbCardBuffer
}

//动作等级
func (self *BaseLogic) GetUserActionRank(cbUserAction byte) byte {
	//抢暗杠等级
	if cbUserAction&static.WIK_QIANG != 0 {
		return 5
	}

	//胡牌等级
	if cbUserAction&static.WIK_CHI_HU != 0 {
		return 4
	}

	//杠牌等级
	if cbUserAction&(static.WIK_GANG) != 0 {
		return 3
	}

	//碰牌等级
	if cbUserAction&static.WIK_PENG != 0 {
		return 2
	}

	//上牌等级
	if cbUserAction&(static.WIK_RIGHT|static.WIK_CENTER|static.WIK_LEFT) != 0 {
		return 1
	}

	return 0
}

//胡牌等级
func (self *BaseLogic) GetChiHuActionRank(ChiHuResult *static.TagChiHuResult) byte {
	//变量定义
	cbChiHuOrder := 0
	wChiHuRight := ChiHuResult.ChiHuRight
	wChiHuKind := (ChiHuResult.ChiHuKind & 0xFF00) >> 4

	//大胡升级
	for i := 0; i < 8; i++ {
		wChiHuKind >>= 1
		if (wChiHuKind & 0x0001) != 0 {
			cbChiHuOrder++
		}
	}

	//权位升级
	for i := 0; i < 16; i++ {
		wChiHuRight >>= 1
		if (wChiHuRight & 0x0001) != 0 {
			cbChiHuOrder++
		}
	}

	return byte(cbChiHuOrder)
}

//吃牌判断
func (self *BaseLogic) EstimateEatCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
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
	var cbCurrentIndex byte = self.SwitchToCardIndex(cbCurrentCard) //3
	//var cbMagicIndex byte = self.SwitchToCardIndex(self.MagicCard)
	for i := 0; i < len(cbItemKind); i++ {
		var cbValueIndex byte = cbCurrentIndex % 9 //3
		//i=0--3>0 && 3-0<6 i=1--3>1&&3-1<6  i==3--3==3&&3-2<6
		if (cbValueIndex >= cbExcursion[i]) && ((cbValueIndex - cbExcursion[i]) <= 6) {
			//吃牌判断
			cbFirstIndex = cbCurrentIndex - cbExcursion[i] //i==0--3 i==1--2 i==2--1
			//i==0--3==3 do cbEatKind|=cbItemKind[i];
			//i==1--3==2+1并且没有值为2的牌那么就构不成左吃
			if (cbCurrentIndex != cbFirstIndex) && (cbCardIndex[cbFirstIndex] == 0) {
				continue
			}
			if (cbCurrentIndex != (cbFirstIndex + 1)) && (cbCardIndex[cbFirstIndex+1] == 0) {
				continue
			}
			if (cbCurrentIndex != (cbFirstIndex + 2)) && (cbCardIndex[cbFirstIndex+2] == 0) {
				continue
			}

			if self.Rule.TypeLaizi == static.TYPE_LAIZI_GANG_PIZI {
				cbPiZiIndex := self.SwitchToCardIndex(self.PiZiCard)
				//吃牌里面不能有皮子
				switch i {
				case 0:
					{
						if cbPiZiIndex == cbCurrentIndex+1 || cbPiZiIndex == cbCurrentIndex+2 {
							continue
						}
					}
					break
				case 1:
					{
						if cbPiZiIndex == cbCurrentIndex-1 || cbPiZiIndex == cbCurrentIndex+1 {
							continue
						}
					}
					break
				case 2:
					{
						if cbPiZiIndex == cbCurrentIndex-1 || cbPiZiIndex == cbCurrentIndex-2 {
							continue
						}
					}
					break
				}
			}

			//设置类型
			cbEatKind |= cbItemKind[i]
		}
	}

	return cbEatKind
}

//新加锁牌判碰
/*
返回的第2个参数标记是哪种碰牌
0：就是普通碰牌
1：表示锁碰（赖子补充的，碰牌的话，只能有一个赖子）
*/
func (self *BaseLogic) EstimatePengCard_Ex(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) (byte, byte) {
	//参数效验
	//ASSERT(IsValidCard(cbCurrentCard));
	if self.EstimatePengCard(cbCardIndex, cbCurrentCard) == static.WIK_PENG {
		return static.WIK_PENG, 0
	}
	if self.MagicCard != static.INVALID_BYTE {
		if cbCardIndex[self.SwitchToCardIndex(cbCurrentCard)] < 2 && cbCardIndex[self.SwitchToCardIndex(self.MagicCard)] > 1 {
			return static.WIK_PENG, 1
		}
	}
	return static.WIK_NULL, 0
}

//碰牌判断
func (self *BaseLogic) EstimatePengCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	//ASSERT(IsValidCard(cbCurrentCard));
	if !self.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}

	//普通碰牌判断
	if cbCardIndex[self.SwitchToCardIndex(cbCurrentCard)] >= 2 {
		return static.WIK_PENG
	}
	//
	return static.WIK_NULL
}

//杠牌判断
func (self *BaseLogic) EstimateGangCard(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//参数效验
	if !self.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}

	//杠牌判断
	if cbCardIndex[self.SwitchToCardIndex(cbCurrentCard)] == 3 {
		return static.WIK_GANG
	}
	return static.WIK_NULL
}

//补牌判断
func (self *BaseLogic) EstimateBu(cbCardIndex [static.MAX_INDEX]byte, cbCurrentCard byte) byte {
	//	//参数效验
	if !self.IsValidCard(cbCurrentCard) {
		return static.WIK_NULL
	}

	//杠牌判断
	if cbCardIndex[self.SwitchToCardIndex(cbCurrentCard)] == 3 {
		return static.WIK_FILL
	}
	return static.WIK_NULL
}

//杠牌分析
func (self *BaseLogic) AnalyseGangCard(cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, GangCardResult *static.TagGangCardResult) byte {
	//设置变量
	cbActionMask := static.WIK_NULL
	//ZeroMemory(&GangCardResult,sizeof(GangCardResult));

	if self.Rule.TypeLaizi == static.TYPE_LAIZI_GANG_FACAI {
		//发财杠
		if cbCardIndex[32] != 0 {
			cbActionMask |= static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = 0x36
			GangCardResult.CardCount++
		}
	} else if self.Rule.TypeLaizi == static.TYPE_LAIZI_GANG_PIZI {
		//皮子杠
		if cbCardIndex[self.SwitchToCardIndex(self.PiZiCard)] != 0 {
			cbActionMask |= static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = self.PiZiCard
			GangCardResult.CardCount++
		}
	}

	//赖子杠,癞晃不提示癞子杠，打出癞子为癞子杠
	if false && cbCardIndex[self.SwitchToCardIndex(self.MagicCard)] != 0 {
		cbActionMask |= static.WIK_GANG
		GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
		GangCardResult.MagicCard = self.MagicCard
		GangCardResult.CardData[GangCardResult.CardCount] = self.MagicCard
		GangCardResult.CardCount++
	}

	//皮子赖子不提示杠
	//手上杠牌
	for i := byte(0); i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] == 4 {
			cbActionMask |= static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
			GangCardResult.CardData[GangCardResult.CardCount] = self.SwitchToCardData(i)
			GangCardResult.CardCount++
		}
	}

	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_PENG {
			if cbCardIndex[self.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				cbActionMask |= static.WIK_FILL
				GangCardResult.CardData[GangCardResult.CardCount] = static.WIK_GANG
				GangCardResult.CardData[GangCardResult.CardCount] = WeaveItem[i].CenterCard
				GangCardResult.CardCount++
			}
		}
	}

	return byte(cbActionMask)
}

//补牌分析
func (self *BaseLogic) AnalyseBuCard(cbCardIndex [static.MAX_INDEX]byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, BuCardResult *static.TagBuResult) byte {
	//设置变量
	cbActionMask := byte(static.WIK_NULL)
	//ZeroMemory(&BuCardResult,sizeof(BuCardResult));

	//手上杠牌
	for i := byte(0); i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] == 4 {
			cbActionMask |= static.WIK_FILL
			BuCardResult.CardData[BuCardResult.CardCount] = static.WIK_FILL
			BuCardResult.CardData[BuCardResult.CardCount] = self.SwitchToCardData(i)
			BuCardResult.CardCount++
		}
	}

	//组合补牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_PENG {
			if cbCardIndex[self.SwitchToCardIndex(WeaveItem[i].CenterCard)] == 1 {
				cbActionMask |= static.WIK_FILL
				BuCardResult.CardData[BuCardResult.CardCount] = static.WIK_FILL
				BuCardResult.CardData[BuCardResult.CardCount] = WeaveItem[i].CenterCard
				BuCardResult.CardCount++
			}
		}
	}

	return cbActionMask
}

//吃胡判断
func (self *BaseLogic) EstimateChiHu(cbCardIndex [static.MAX_INDEX]byte, ChiHuResult *static.TagChiHuResult, i uint16) byte {
	return byte(static.WIK_NULL)
}

//四喜胡牌
func (self *BaseLogic) IsSiXi(cbCardIndex []byte) bool {
	//胡牌判断
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] == 4 {
			return true
		}
	}
	return false
}

//缺一色牌
func (self *BaseLogic) IsQueYiSe(cbCardIndex []byte) bool {
	//胡牌判断
	cbIndex := []byte{0, 9, 18}
	for i := 0; i < len(cbIndex); i++ {
		var j byte
		for j = cbIndex[i]; j < (cbIndex[i] + 9); j++ {
			if cbCardIndex[j] != 0 {
				break
			}
		}
		if j == (cbIndex[i] + 9) {
			return true
		}
	}

	return false
}

//判断板板胡牌
func (self *BaseLogic) IsBanBanHu(cbCardIndex []byte) bool {
	//胡牌判断
	for i := 1; i < static.MAX_INDEX; i += 3 {
		if cbCardIndex[i] != 0 {
			return false
		}
	}
	return true
}

//六六顺牌
func (self *BaseLogic) IsLiuLiuShun(cbCardIndex []byte) bool {
	//胡牌判断
	cbPengCount := byte(0)
	for i := 0; i < static.MAX_INDEX; i++ {
		cbPengCount++
		if (cbCardIndex[i] >= 3) && (cbPengCount >= 2) {
			return true
		}
	}
	return false
}

//无赖子的清一色牌，包括万一色，条一色，筒一色
func (self *BaseLogic) IsQingYiSe(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte) bool {
	//构造麻将
	cbCardIndexTemp := cbCardIndex[:]
	for i := 27; i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] != 0 {
			return false
		}
	}

	//胡牌判断
	cbCardColor := byte(0xFF)
	for i := byte(0); i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] != 0 {
			//花色判断
			if cbCardColor != 0xFF {
				return false
			}

			//设置花色
			cbCardColor = (self.SwitchToCardData(i) & static.MASK_COLOR)

			//设置索引
			if i <= 27 {
				i = (i/9+1)*9 - 1
			} else {
				break
				//i=28;
			}
		}
	}

	//对组合牌进行判断
	for i := byte(0); i < cbItemCount; i++ {
		cbCenterCard := WeaveItem[i].CenterCard
		if (cbCenterCard & static.MASK_COLOR) != cbCardColor {
			return false
		}
	}
	return true
}

//风一色牌,不带赖子
func (self *BaseLogic) IsFengYiSe(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) bool {
	//构造麻将
	cbCardIndexTemp := cbCardIndex[:]

	cbCardColor := byte(0x30)

	for i := byte(0); i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] != 0 {
			//花色判断
			if (self.SwitchToCardData(i) & static.MASK_COLOR) != cbCardColor {
				return false
			}
		}
	}

	//对组合牌进行判断
	for i := byte(0); i < cbWeaveCount; i++ {
		cbCenterCard := WeaveItem[i].CenterCard
		if (cbCenterCard & static.MASK_COLOR) != cbCardColor {
			return false
		}
	}
	return true
}

//检查七对
func (self *BaseLogic) IsQiDui(cbCardIndex []byte) (bool, int) {
	//每组牌的张数
	iNotDui := uint16(0)
	HaoHuaNum := 0
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i]%2 != 0 {
			//非偶数
			iNotDui++
		}
		if cbCardIndex[i] == 4 {
			HaoHuaNum++
		}
	}

	if iNotDui == 0 {
		return true, HaoHuaNum
	}

	return false, HaoHuaNum
}

//检查豪华七对
func (self *BaseLogic) IsHaoHuaQiDuiNoMargic(cbCardIndex []byte) (bool, int) {
	MagicCount := int(self.GetMagicCount(cbCardIndex))
	//每组牌的张数
	iHaoHuaDui := 0
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i]%2 != 0 && self.SwitchToCardIndex(self.MagicCard) != byte(i) {
			MagicCount--
		}
		if cbCardIndex[i] == 4 && self.SwitchToCardIndex(self.MagicCard) != byte(i) {
			iHaoHuaDui++
		}

		if MagicCount < 0 {
			return false, 0
		}
	}

	//赖子匹配玩七对,剩下只可能是双数,是单数就有有问题的
	if MagicCount >= 0 && MagicCount%2 == 0 {
		iHaoHuaDui += MagicCount / 4
		return true, iHaoHuaDui
	}

	return false, 0
}

//检查豪华七对
func (self *BaseLogic) IsHaoHuaQiDui(cbCardIndex []byte) (bool, int) {
	MagicCount := int(self.GetMagicCount(cbCardIndex))
	//每组牌的张数
	iHaoHuaDui := 0
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i]%2 != 0 && self.SwitchToCardIndex(self.MagicCard) != byte(i) {
			if cbCardIndex[i] == 1 {

			} else if cbCardIndex[i] == 3 {
				iHaoHuaDui++
			}
			MagicCount--
		}
		if cbCardIndex[i] == 4 && self.SwitchToCardIndex(self.MagicCard) != byte(i) {
			iHaoHuaDui++
		}

		if MagicCount < 0 {
			return false, 0
		}
	}

	//赖子匹配玩七对,剩下只可能是双数,是单数就有有问题的
	if MagicCount >= 0 && MagicCount%2 == 0 {
		iHaoHuaDui += MagicCount / 2
		return true, iHaoHuaDui
	}

	return false, 0
}

//7对牌
func (self *BaseLogic) Is7Dui(cbCardIndex []byte, bMagic *bool, wMagicCount uint16) bool {
	*bMagic = false

	//每组牌的张数
	iNotDui := uint16(0)
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i]%2 != 0 && self.SwitchToCardIndex(self.MagicCard) != byte(i) {
			//非偶数
			iNotDui++
		}
	}

	if iNotDui == 0 {
		*bMagic = false
		return true
	}
	//癞子够配
	if wMagicCount >= 1 && iNotDui <= wMagicCount {
		*bMagic = true
		return true
	}

	return false
}

//7对牌
func (self *BaseLogic) Is7DuiNoMagic(cbCardIndex []byte) bool {
	//每组牌的张数
	iNotDui := uint16(0)
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i]%2 != 0 {
			//非偶数
			iNotDui++
		}
	}

	if iNotDui == 0 {
		return true
	}

	return false
}

//豪华对牌
func (self *BaseLogic) IsHaoHuaDui(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte) bool {
	//变量定义
	bFourCard := false

	//组合判断
	for i := byte(0); i < cbItemCount; i++ {
		//杆补判断
		if WeaveItem[i].WeaveKind != static.WIK_FILL {
			return false
		}
		if WeaveItem[i].WeaveKind != static.WIK_GANG {
			return false
		}

		//设置变量
		bFourCard = true
	}

	//扑克判断
	for i := 0; i < static.MAX_INDEX; i++ {
		//四牌判断
		if cbCardIndex[i] == 4 {
			bFourCard = true
			continue
		}

		//对牌判断
		if (cbCardIndex[i] != 0) && (cbCardIndex[i] != 2) {
			return false
		}
	}

	//结果判断
	if bFourCard == false {
		return false
	}

	return true
}

//不带赖子的将将胡牌,
func (self *BaseLogic) IsJiangJiangHu(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) bool {
	//构造麻将
	cbCardIndexTemp := cbCardIndex[:]

	//扑克判断
	for i := 0; i < static.MAX_INDEX; i++ {
		if (i%3 != 1) && (cbCardIndexTemp[i] != 0) {
			return false
		}
	}

	for i := 27; i < static.MAX_INDEX; i++ {
		if cbCardIndexTemp[i] != 0 {
			return false
		}
	}

	//组合判断
	for i := byte(0); i < cbWeaveCount; i++ {
		//类型判断
		cbWeaveKind := WeaveItem[i].WeaveKind
		if (cbWeaveKind != static.WIK_PENG) && (cbWeaveKind != static.WIK_GANG) && (cbWeaveKind != static.WIK_FILL) {
			return false
		}

		//数值判断
		cbCenterValue := (WeaveItem[i].CenterCard & static.MASK_VALUE)
		cbCenterColor := (WeaveItem[i].CenterCard & static.MASK_COLOR)

		//不是同条万也不能胡将一色
		if cbCenterColor != 0x00 && cbCenterColor != 0x10 && cbCenterColor != 0x20 {
			return false
		}
		if (cbCenterValue != 2) && (cbCenterValue != 5) && (cbCenterValue != 8) {
			return false
		}
	}

	return true
}

//扑克转换成牌数据
func (self *BaseLogic) SwitchToCardData(cbCardIndex byte) byte {
	//	ASSERT(cbCardIndex<public.MAX_INDEX);
	return ((cbCardIndex / 9) << 4) | (cbCardIndex%9 + 1)
}

//扑克数据转换成扑克点数转换的结果：
//  万，  条， 筒
// 0-8   9-15  16-23

func (self *BaseLogic) SwitchToCardIndex(cbCardData byte) byte {
	cardIndex := ((cbCardData&static.MASK_COLOR)>>4)*9 + (cbCardData & static.MASK_VALUE) - 1
	return cardIndex
}

func (self *BaseLogic) SwitchToCardData2(cbCardIndex [static.MAX_INDEX]byte, cbCardData [static.MAX_COUNT]byte) (byte, [static.MAX_COUNT]byte) {
	//转换扑克
	cbCardData = [static.MAX_COUNT]byte{0}
	cbPosition := byte(0)
	for i := byte(0); i < static.MAX_INDEX; i++ { // 34种牌，每种牌有几张
		if cbCardIndex[i] != 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				if cbPosition >= static.MAX_COUNT {
					xlog.Logger().Errorln("SwitchToCardData2 error cardId :" + strconv.Itoa(int(cbPosition)))
					continue
				}
				cbCardData[cbPosition] = self.SwitchToCardData(i)
				cbPosition++
			}
		}
	}

	return cbPosition, cbCardData
}

//扑克转换
func (self *BaseLogic) SwitchToCardIndex2(cbCardData []byte, cbCardCount byte, cbCardIndex [static.MAX_INDEX]byte) (byte, [static.MAX_INDEX]byte) {
	//设置变量
	cbCardIndex = [static.MAX_INDEX]byte{}
	//转换扑克
	for i := byte(0); i < cbCardCount; i++ {
		//ASSERT(IsValidCard(cbCardData[i]));
		if !self.IsValidCard(cbCardData[i]) {
			return cbCardCount, cbCardIndex
		}
		cbCardIndex[self.SwitchToCardIndex(cbCardData[i])]++
	}

	return cbCardCount, cbCardIndex
}

//通过手牌产生离散表
//原则：1，手牌中已经存在的牌，和距离已存在的牌为1的空位，都可以用癞子代替，其他位置没有必要用癞子代替；
//原则：2，单牌指左右两边距离为2以内都是空位，距离单牌为1的空位不用癞子替，比如4万7万8万，没有必要用两个癞子组成345万或456万，因为组成444万照样可以胡牌
//原则：3，2连牌指两个牌中间有一个空位或两个连续的牌前后都有空位，需要在中间空位用癞子替或者后面替，比如4万6万，用一个癞子组成456万，比如4万5万，用一个癞子组成456万，没有必要在前面替癞子组成345万。
//原则：4，只有2连牌是89时需要在前面替组成789
//用例：   23  67 9
//离散表为:11111111,即2到9的位置都可能需要用到癞子
//用例：   4   89
//离散表为:100111,即4和789的位置可能需要用到癞子
//经验表明当用户有3或4个癞子时，在能胡牌时离散值不会超过13，需要的byNeedMagicCount参考值不会超过6个，所以癞子替换时不需要遍历34张牌，遍历15张牌就可以保证所有胡牌类型都包括在内
//输入：abyCardIndex 手牌索引，对应位置为0表示没有这个牌
//输出：byNeedMagicCount 需要用到的最少的癞子数量，参考值；abyDisperseTable 离散表，癞子可以代替的值
//返回值：返回离散值，值越大表示牌分布越分散
func (self *BaseLogic) CreatDisperseTable(abyCardIndex []byte) (byte, byte, [static.MAX_INDEX]byte) {
	//设置变量
	//ZeroMemory(abyDisperseTable,sizeof(BYTE)*public.MAX_INDEX);
	byNeedMagicCount := byte(0)  //需要用到的最少的癞子数量，参考值
	bySeriesCount := byte(0)     //连牌计数
	bySeriesNULLCount := byte(0) //连续空位计数
	abyDisperseTable := [static.MAX_INDEX]byte{0}

	//产生离散表
	for i := 0; i < static.MAX_INDEX; i++ {
		//风牌先判断，不能构成连牌
		if i > 26 {
			if 0 != abyCardIndex[i] {
				abyDisperseTable[i] = 1
				if 1 == abyCardIndex[i] || 4 == abyCardIndex[i] {
					byNeedMagicCount++ //这里需要用癞子
				}
			} else {
				abyDisperseTable[i] = 0
			}
			continue
		}
		if 0 != abyCardIndex[i] {
			abyDisperseTable[i] = 1
			bySeriesCount++
			//2连，且中间只有一个空位，需要把空位补齐，表示空位可以用癞子代替
			if 2 == bySeriesCount && 1 == bySeriesNULLCount {
				abyDisperseTable[i-1] = 1
				bySeriesCount++    //2连变成3连
				byNeedMagicCount++ //这里需要用癞子
			} else {
				if 2 == bySeriesCount && 8 == i%9 {
					abyDisperseTable[i-2] = 1
					bySeriesCount++    //2连变成3连
					byNeedMagicCount++ //这里需要用癞子
				}
			}
			bySeriesNULLCount = 0

		} else {
			abyDisperseTable[i] = 0
			//重新计算连牌个数
			if bySeriesCount >= 3 {
				bySeriesCount = 1
			}
			//2连，且末位有一个空位，需要把空位补齐，表示空位可以用癞子代替
			if 2 == bySeriesCount && 0 == bySeriesNULLCount {
				abyDisperseTable[i] = 1
				bySeriesCount++    //2连变成3连
				byNeedMagicCount++ //这里需要用癞子
			} else {
				bySeriesNULLCount++
				//连续空位超过2个，不需要补癞子了
				if bySeriesNULLCount >= 2 {
					bySeriesCount = 0
				}
			}

		}
		if 8 == i%9 {
			bySeriesCount = 0
			bySeriesNULLCount = 0
		}

	}
	//统计离散表的离散值，值越大表示越分散
	byDisperseCount := byte(0)
	for i := 0; i < static.MAX_INDEX; i++ {
		byDisperseCount += abyDisperseTable[i]
	}
	//赖子本身也放在离散表中
	abyDisperseTable[self.SwitchToCardIndex(self.MagicCard)] = 1
	//加一个将牌
	abyDisperseTable[self.SwitchToCardIndex(0x15)] = 1

	return byDisperseCount, byNeedMagicCount, abyDisperseTable
}

//离散值估计
func (self *BaseLogic) DisperseEstimate(byDisperseCount byte, byNeedMagicCount byte, byRealMagicCount byte) bool {
	//认为离散值超过15，癞子数为4个或3个的情况下永远不能胡牌
	if byDisperseCount > 18 {
		return false
	}
	//需要使用的癞子数量远大于癞子的实际数量，认为不可能胡牌
	if byNeedMagicCount > byRealMagicCount+3 {
		return false
	}
	//离散值少于15时，癞子替换牌时使用离散表遍历
	return true
}

//分析扑克，cbCardIndex：最终的点数数组，包含了要分析的牌
func (self *BaseLogic) AnalyseCard(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) (bool, []static.TagAnalyseItem) {
	var AnalyseItemArray []static.TagAnalyseItem
	//统计索引数组中的所有牌数
	cbCardCount := byte(0)
	for i := byte(0); i < static.MAX_INDEX; i++ {
		cbCardCount += cbCardIndex[i]
	}

	//效验数目，索引数组中牌的总数-2（一对将）后剩下的数是3的倍数
	if (cbCardCount < 2) || (cbCardCount > static.MAX_COUNT) || ((cbCardCount-2)%3 != 0) {
		return false, AnalyseItemArray
	}

	////////////////////////下面出现的情况只能是：索引数组中牌的总数-2（一对将）后剩下的数是3的倍数，不可能出现4，6。。张的情况///////////////////////////////////
	//类型子项
	//变量定义
	cbKindItemCount := byte(0)
	var KindItem [static.MAX_COUNT - 2]static.TagKindItem

	//需求判断，cbLessKindItem必须是3的倍数，cbCardCount不可能为4
	cbLessKindItem := byte((cbCardCount - 2) / 3)
	//	ASSERT((cbLessKindItem+cbWeaveCount)==4);

	//单吊判断，cbCardCount=2的情况：原来手中牌只有一张牌，加入要分析的牌后正好构成两张，其他的牌都在组合牌中
	if cbLessKindItem == 0 {
		//效验参数
		//此种情况是：原来手中只有一张牌，加入一张待分析的牌正好构成两张，并且有四组组合牌型
		//ASSERT((cbCardCount==2)&&(cbWeaveCount==4));

		//判断加入了待分析的牌后是否构成对子
		for i := byte(0); i < static.MAX_INDEX; i++ {
			//如果牌索引数组中有一个对子，保存分析结果
			if cbCardIndex[i] == 2 {
				//分析子项
				//变量定义
				var AnalyseItem static.TagAnalyseItem

				//分析每一组组合牌，得到组合牌的组合牌型和中间牌，比如说WK_PENG,WK_CHI保存到分析子项中
				for j := byte(0); j < cbWeaveCount; j++ {
					AnalyseItem.WeaveKind[j] = WeaveItem[j].WeaveKind
					AnalyseItem.CenterCard[j] = self.SwitchToCardIndex(WeaveItem[j].CenterCard) //centercard统一风格
				}
				//将待分析的牌索引转换成牌值，作为牌眼保存起来
				AnalyseItem.CardEye = self.SwitchToCardData(i)

				//将分析结果插入到分析数组中
				//AnalyseItemArray.Add(AnalyseItem);
				AnalyseItemArray = append(AnalyseItemArray, AnalyseItem)

				return true, AnalyseItemArray
			}
		}
		return false, AnalyseItemArray
	}
	//加入待分析的牌后，手中牌>=3的情况，对手中牌索引数组进行分析
	if cbCardCount >= 3 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			//1.同牌判断，胡牌时，杠牌算碰牌
			if cbCardIndex[i] >= 3 {
				KindItem[cbKindItemCount].CenterCard = i
				KindItem[cbKindItemCount].CardIndex[0] = i
				KindItem[cbKindItemCount].CardIndex[1] = i
				KindItem[cbKindItemCount].CardIndex[2] = i
				KindItem[cbKindItemCount].WeaveKind = static.WIK_PENG
				cbKindItemCount++
				//4个赖子的情况，最多可能出现8个铜牌
				if cbCardIndex[i] >= 6 {
					KindItem[cbKindItemCount].CenterCard = i
					KindItem[cbKindItemCount].CardIndex[0] = i
					KindItem[cbKindItemCount].CardIndex[1] = i
					KindItem[cbKindItemCount].CardIndex[2] = i
					KindItem[cbKindItemCount].WeaveKind = static.WIK_PENG
					cbKindItemCount++
				}
			}
			//2.连牌判断
			//????(i<(public.MAX_INDEX-2)不清楚为什么会这样写，麻将总点数也才到26，就算能到34，风也不能构成连牌?????????
			if (i < (29 - 2)) && (cbCardIndex[i] > 0) && ((i % 9) < 7) {
				for j := byte(1); j <= cbCardIndex[i]; j++ {
					if (cbCardIndex[i+1] >= j) && (cbCardIndex[i+2] >= j) {
						KindItem[cbKindItemCount].CenterCard = i
						KindItem[cbKindItemCount].CardIndex[0] = i
						KindItem[cbKindItemCount].CardIndex[1] = i + 1
						KindItem[cbKindItemCount].CardIndex[2] = i + 2
						KindItem[cbKindItemCount].WeaveKind = static.WIK_LEFT
						cbKindItemCount++
					}
				}
			}
		}
	}
	//组合分析，cbLessKindItem是手中牌总数-2后得到的3的倍数，cbKindItemCount是对手中牌进行分析后得出来的最多的组合类型
	//比如手中牌有：5万，6万，6万，7万，8万，cbKindItemCount=2 cbLessKindItem=1
	if cbKindItemCount >= cbLessKindItem {
		//变量定义
		cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)
		//变量定义
		cbIndex := [4]byte{0, 1, 2, 3}
		var pKindItem [4]*static.TagKindItem
		//do循环的作用：
		/*
				cbLessKindItem：一组牌中如果可以胡牌的话，需要的组合数，
				cbKindItemCount：对手中牌的组合进行分析，可以得出的最多的组合数
				1.将待分析的牌保存到临时数组中，取前面分析出来的所有组合，每次分析cbKindItemCount个，
				  在临时数组中减去这cbKindItemCount个组合中的牌，对剩下的牌进行分析，如果还有对牌，可以
				  胡牌，保存这种组合类型到分析数组中，
			   2.设置索引数组，将后面的组合下标放到索引数组中，下一次循环的时候就取新设置的索引数组对应的
			     分析子项。再进行判断。
		*/
		nAccert := 0
		for {
			nAccert++
			if nAccert > 600 {
				//m_mylog.Log("分析扑克时死循环啦");
				break
			}
			//每次循环将传进来的牌索引数组拷贝到临时数组中，进行分析
			copy(cbCardIndexTemp, cbCardIndex[:])
			//每次从上面分析得出的分析子项中取cbLessKindItem个分析子项进行分析，
			//注意：索引数组cbIndex[]在每次循环结束时都重新设置了
			for i := byte(0); i < cbLessKindItem; i++ {
				pKindItem[i] = &KindItem[cbIndex[i]]
			}
			//数量判断
			bEnoughCard := true
			//修改临时数组的值，把临时数组中构成cbLessKindItem个分析子项里的每一张牌，牌数减1，
			for i := byte(0); i < cbLessKindItem*3; i++ {
				//存在判断
				cbCardIndex := pKindItem[i/3].CardIndex[i%3]
				if cbCardIndexTemp[cbCardIndex] == 0 {
					bEnoughCard = false
					break
				} else {
					cbCardIndexTemp[cbCardIndex]--
				}
			}

			//胡牌判断，注意下面使用到的cbCardIndexTemp[]数组是经前面修改过后的
			if bEnoughCard == true {
				//牌眼判断
				cbCardEye := byte(0)
				//检查cbCardIndexTemp[]数组中是否还有对牌，如果有就将该对牌设置为牌眼，说明此种组合有可能
				//会胡牌，跳出
				for i := byte(0); i < static.MAX_INDEX; i++ {
					if cbCardIndexTemp[i] == 2 {
						cbCardEye = self.SwitchToCardData(i)
						break
					}
				}

				//组合类型
				if cbCardEye != 0 {
					//变量定义
					var AnalyseItem static.TagAnalyseItem

					//得到组合牌中的牌型，保存到分析子项中
					for i := byte(0); i < cbWeaveCount; i++ {
						AnalyseItem.WeaveKind[i] = WeaveItem[i].WeaveKind
						AnalyseItem.CenterCard[i] = self.SwitchToCardIndex(WeaveItem[i].CenterCard)
					}

					//得到手中牌的牌型，保存到分析子项中
					for i := byte(0); i < cbLessKindItem; i++ {
						AnalyseItem.WeaveKind[i+cbWeaveCount] = pKindItem[i].WeaveKind
						AnalyseItem.CenterCard[i+cbWeaveCount] = pKindItem[i].CenterCard
					}

					//设置牌眼
					AnalyseItem.CardEye = cbCardEye

					//将分析子项插入到分析数组中
					//AnalyseItemArray.Add(AnalyseItem);
					AnalyseItemArray = append(AnalyseItemArray, AnalyseItem)
				}
			}

			//设置索引，索引数组中存放的是分析子项数组的下标，每次取分析子项进行分析时，都是按照索引数组
			//里面存放的下标值进行存取，当cbIndex[cbLessKindItem-1]的最后一位存放的值与得出的分析子项下标相同，
			//重新调整索引数组，下一次取值就会取新的组合
			if cbIndex[cbLessKindItem-1] == (cbKindItemCount - 1) {
				var i byte
				for i = cbLessKindItem - 1; i > 0; i-- {
					if (cbIndex[i-1] + 1) != cbIndex[i] {
						cbNewIndex := cbIndex[i-1]
						for j := (i - 1); j < cbLessKindItem; j++ {
							cbIndex[j] = cbNewIndex + j - i + 2
						}
						break
					}
				}
				//跳出整个while循环
				if i == 0 {
					break
				}
			} else {
				cbIndex[cbLessKindItem-1]++
			}

		}

	}

	return (len(AnalyseItemArray) > 0), AnalyseItemArray
}

//基本组合分析
func (self *BaseLogic) AnalyHuArray(cbCardIndex []byte) bool {
	cbCardCount := byte(0)
	for i := byte(0); i < static.MAX_INDEX; i++ {
		cbCardCount += cbCardIndex[i]
	}

	//效验数目，索引数组中牌的总数-2（一对将）后剩下的数是3的倍数
	if (cbCardCount < 2) || (cbCardCount > static.MAX_COUNT) || ((cbCardCount-2)%3 != 0) {
		return false
	}

	////////////////////////下面出现的情况只能是：索引数组中牌的总数-2（一对将）后剩下的数是3的倍数，不可能出现4，6。。张的情况///////////////////////////////////
	//变量定义
	cbKindItemCount := byte(0)
	var KindItem [static.MAX_COUNT - 2]static.TagKindItem
	//需求判断，cbLessKindItem必须是3的倍数，cbCardCount不可能为4
	cbLessKindItem := byte((cbCardCount - 2) / 3)

	//单吊判断，cbCardCount=2的情况：原来手中牌只有一张牌，加入要分析的牌后正好构成两张，其他的牌都在组合牌中
	if cbLessKindItem == 0 {
		//判断是否构成对子
		for i := byte(0); i < static.MAX_INDEX; i++ {
			//如果牌索引数组中有一个对子，保存分析结果
			if cbCardIndex[i] == 2 {
				return true
			}
		}
		return false
	}

	//加入待分析的牌后，手中牌>=3的情况，对手中牌索引数组进行分析
	if cbCardCount >= 3 {
		for i := byte(0); i < static.MAX_INDEX; i++ {
			//1.同牌判断，胡牌时，杠牌算碰牌
			if cbCardIndex[i] >= 3 {

				KindItem[cbKindItemCount].CenterCard = i
				KindItem[cbKindItemCount].CardIndex[0] = i
				KindItem[cbKindItemCount].CardIndex[1] = i
				KindItem[cbKindItemCount].CardIndex[2] = i

				KindItem[cbKindItemCount].WeaveKind = static.WIK_PENG
				cbKindItemCount++

				//4个赖子的情况，最多可能出现8个铜牌
				if cbCardIndex[i] >= 6 {
					KindItem[cbKindItemCount].CenterCard = i
					KindItem[cbKindItemCount].CardIndex[0] = i
					KindItem[cbKindItemCount].CardIndex[1] = i
					KindItem[cbKindItemCount].CardIndex[2] = i

					KindItem[cbKindItemCount].WeaveKind = static.WIK_PENG
					cbKindItemCount++
				}
			}

			//2.连牌判断
			//????(i<(public.MAX_INDEX-2)不清楚为什么会这样写，麻将总点数也才到26，就算能到34，风也不能构成连牌?????????
			if (i < (29 - 2)) && (cbCardIndex[i] > 0) && ((i % 9) < 7) {
				for j := byte(1); j <= cbCardIndex[i]; j++ {
					if (cbCardIndex[i+1] >= j) && (cbCardIndex[i+2] >= j) {
						KindItem[cbKindItemCount].CenterCard = i
						KindItem[cbKindItemCount].CardIndex[0] = i
						KindItem[cbKindItemCount].CardIndex[1] = i + 1
						KindItem[cbKindItemCount].CardIndex[2] = i + 2
						KindItem[cbKindItemCount].WeaveKind = static.WIK_LEFT
						cbKindItemCount++
					}
				}
			}
		}
	}

	//组合分析，cbLessKindItem是手中牌总数-2后得到的3的倍数，cbKindItemCount是对手中牌进行分析后得出来的最多的组合类型
	//比如手中牌有：5万，6万，6万，7万，8万，cbKindItemCount=2 cbLessKindItem=1
	if cbKindItemCount >= cbLessKindItem {
		//变量定义
		cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

		//变量定义
		cbIndex := [4]byte{0, 1, 2, 3}

		var pKindItem [4]*static.TagKindItem
		nAccert := 0
		for {
			nAccert++
			if nAccert > 600 {
				//m_mylog.Log("分析扑克时死循环啦");
				break
			}
			//每次循环将传进来的牌索引数组拷贝到临时数组中，进行分析
			copy(cbCardIndexTemp, cbCardIndex[:])

			//每次从上面分析得出的分析子项中取cbLessKindItem个分析子项进行分析，
			//注意：索引数组cbIndex[]在每次循环结束时都重新设置了
			for i := byte(0); i < cbLessKindItem; i++ {
				pKindItem[i] = &KindItem[cbIndex[i]]
			}

			//数量判断
			bEnoughCard := true

			//修改临时数组的值，把临时数组中构成cbLessKindItem个分析子项里的每一张牌，牌数减1，
			for i := byte(0); i < cbLessKindItem*3; i++ {
				//存在判断
				cbCardIndex := pKindItem[i/3].CardIndex[i%3]
				if cbCardIndexTemp[cbCardIndex] == 0 {
					bEnoughCard = false
					break
				} else {
					cbCardIndexTemp[cbCardIndex]--
				}
			}

			//胡牌判断，注意下面使用到的cbCardIndexTemp[]数组是经前面修改过后的
			if bEnoughCard == true {
				//牌眼判断
				//会胡牌，跳出
				for i := byte(0); i < static.MAX_INDEX; i++ {
					if cbCardIndexTemp[i] == 2 {
						//cbCardEye = self.SwitchToCardData(i)
						return true
					}
				}
			}

			//设置索引，索引数组中存放的是分析子项数组的下标，每次取分析子项进行分析时，都是按照索引数组
			//里面存放的下标值进行存取，当cbIndex[cbLessKindItem-1]的最后一位存放的值与得出的分析子项下标相同，
			//重新调整索引数组，下一次取值就会取新的组合
			if cbIndex[cbLessKindItem-1] == (cbKindItemCount - 1) {
				var i byte
				for i = cbLessKindItem - 1; i > 0; i-- {
					if (cbIndex[i-1] + 1) != cbIndex[i] {
						cbNewIndex := cbIndex[i-1]
						for j := (i - 1); j < cbLessKindItem; j++ {
							cbIndex[j] = cbNewIndex + j - i + 2
						}
						break
					}
				}
				//跳出整个while循环
				if i == 0 {
					break
				}
			} else {
				cbIndex[cbLessKindItem-1]++
			}
		}
	}

	return false
}

//转换牌名
//SwitchToCardName1(const BYTE cbCardIndex[])
func (self *BaseLogic) SwitchToCardName2(cbCardIndex []byte, msgType byte) string {
	szCardName := string("")
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] > 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				//szCardName.Append(StrCardsMessage[i]+",");

				if msgType == 0 {
					szCardName += StrCardsMessage[i] + ","
				} else {
					szCardName += StrCardsMessage1[i]
				}
			}
		}
	}
	return szCardName
}

//转换牌名
//SwitchToCardName1(const BYTE cbCardIndex)
func (self *BaseLogic) SwitchToCardName(cbCardIndex byte, msgType byte) string {
	if msgType == 0 {
		return StrCardsMessage[cbCardIndex]
	}
	return StrCardsMessage1[cbCardIndex]
}

//SwitchToCardName1(const BYTE cbCardData[], BYTE cbCardCount)
func (self *BaseLogic) SwitchToCardName3(cbCardData []byte, cbCardCount byte, msgType byte) string {
	szCardName := string("")

	for i := byte(0); i < cbCardCount; i++ {
		if !self.IsValidCard(cbCardData[i]) {
			szCardName += strconv.Itoa(int(cbCardData[i]))
			continue
		}
		index := self.SwitchToCardIndex(cbCardData[i])
		if self.IsValidCard(cbCardData[index]) {
			if msgType == 0 {
				szCardName += StrCardsMessage[index] + ","
			} else {
				szCardName += StrCardsMessage1[index] + ","
			}

		}
	}
	return szCardName
}

// replace SwitchToCardName, switch card index to name
func (self *BaseLogic) SwitchToCardNameByIndex(cbCardIndex byte, msgType byte) string {
	if msgType == 0 {
		return StrCardsMessage[cbCardIndex]
	}
	return StrCardsMessage1[cbCardIndex]
}

// replace SwitchToCardName2, switch card indexs to name
func (self *BaseLogic) SwitchToCardNameByIndexs(cbCardIndex []byte, msgType byte) string {
	szCardName := string("")
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] > 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				//szCardName.Append(StrCardsMessage[i]+",");

				if msgType == 0 {
					szCardName += StrCardsMessage[i] + ","
				} else {
					szCardName += StrCardsMessage1[i]
				}
			}
		}
	}
	return szCardName
}

// replace SwitchToCardName3, switch card data to name
func (self *BaseLogic) SwitchToCardNameByData(cbCardData byte, msgType byte) string {
	szCardName := string("")
	if !self.IsValidCard(cbCardData) {
		szCardName += strconv.Itoa(int(cbCardData))
		return szCardName
	}
	index := self.SwitchToCardIndex(cbCardData)
	if msgType == 0 {
		szCardName += StrCardsMessage[index]
	} else {
		szCardName += StrCardsMessage1[index]
	}
	return szCardName
}

// replace SwitchToCardName3, switch card datas to name
func (self *BaseLogic) SwitchToCardNameByDatas(cbCardData []byte, msgType byte) string {
	szCardName := string("")

	for i := 0; i < len(cbCardData); i++ {
		if !self.IsValidCard(cbCardData[i]) {
			szCardName += strconv.Itoa(int(cbCardData[i]))
			continue
		}
		index := self.SwitchToCardIndex(cbCardData[i])
		//fmt.Println("\n","index", index, "cb", len(cbCardData), "---", cbCardData,"---------",cbCardData[i])
		if self.IsValidCard(cbCardData[i]) {
			if msgType == 0 {
				szCardName += StrCardsMessage[index] + ","
			} else {
				szCardName += StrCardsMessage1[index] + ","
			}

		}
	}
	return szCardName
}

//判断是否符合甩字胡的规则  是甩字胡 返回false 不是返回true
func (self *BaseLogic) AnalyseShuaiZiHu(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbItemCount byte, wChiHuRight uint16) bool {
	//构造扑克
	cbCardIndexTemp := make([]byte, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex)
	wMagicCount := self.GetMagicCount(cbCardIndexTemp)

	//甩字胡时必须有且仅有一个赖子
	if wMagicCount != 1 {
		return true
	}

	//听牌分析
	for i := 0; i < len(cbCardIndex); i++ {
		//胡牌分析
		cbCurrentCard := self.SwitchToCardData(byte(i))
		//只要有一组牌型构成甩字胡，即为甩字胡牌型
		if self.CheckShuaiZiHu(cbCardIndexTemp, WeaveItem, cbItemCount, cbCurrentCard) == false {
			return false
		}
	}
	return true
}

//判断是否符合甩字胡的规则
//甩字胡规则：其他牌都成句或成刻，有一对将，一个散牌以及一个赖子，此为甩字胡，此时只能自摸不能捉铳
func (self *BaseLogic) CheckShuaiZiHu(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte) bool {
	//cbCardIndex中不包含cbCurrentCard,包含赖子

	var AnalyseItemArray []static.TagAnalyseItem

	//构造扑克
	cbTempCard := make([]byte, static.MAX_INDEX)
	copy(cbTempCard, cbCardIndex)
	if cbCurrentCard != 0 {
		cbTempCard[self.SwitchToCardIndex(cbCurrentCard)]++
	}

	wMagicCount := self.GetMagicCount(cbTempCard)
	if wMagicCount != 1 {
		return true //甩字胡只能有一个赖子
	}
	if cbCurrentCard == self.MagicCard {
		return true //唯一的赖子是别人打出的牌
	}

	//清除赖子牌
	if wMagicCount > 0 {
		cbTempCard[(self.SwitchToCardIndex(self.MagicCard))] = 0
	}
	//一张赖子的情况
	if wMagicCount == 1 {
		//ASSERT(0);
		//判断用赖子替值时能否胡牌
		for i := byte(0); i < 27; i++ {
			if i == 31 {
				continue
			}
			if self.SwitchToCardData(byte(i)) == cbCurrentCard {
				continue
			}

			cbTempCard[i]++

			//分析记录清理
			_, AnalyseItemArray = self.AnalyseCard(cbTempCard, WeaveItem, cbWeaveCount)
			//胡牌分析

			for iItem := 0; iItem < len(AnalyseItemArray); iItem++ {
				if cbWeaveCount > byte(len(AnalyseItemArray[iItem].WeaveKind)) {
					continue
				}

				//判断是否为甩字胡时，要去掉已经吃、碰、杠的组合牌
				for j := cbWeaveCount; j < byte(len(AnalyseItemArray[iItem].WeaveKind)); j++ {
					if static.WIK_PENG == AnalyseItemArray[iItem].WeaveKind[j] {
						if self.SwitchToCardIndex(cbCurrentCard) == AnalyseItemArray[iItem].CenterCard[j] && self.SwitchToCardData(byte(i)) == cbCurrentCard {
							return false
						}
					} else if static.WIK_LEFT == AnalyseItemArray[iItem].WeaveKind[j] { //分析胡牌牌型产生的新的吃牌类型只有WIK_LEFT
						//cbCenterCard 实际上是最左的一张牌，bugid 5891
						if self.SwitchToCardIndex(cbCurrentCard) == AnalyseItemArray[iItem].CenterCard[j] || self.SwitchToCardIndex(cbCurrentCard) == AnalyseItemArray[iItem].CenterCard[j]+1 || self.SwitchToCardIndex(cbCurrentCard) == AnalyseItemArray[iItem].CenterCard[j]+2 {
							if i == AnalyseItemArray[iItem].CenterCard[j] || i == AnalyseItemArray[iItem].CenterCard[j]+1 || i == AnalyseItemArray[iItem].CenterCard[j]+2 {
								return false
							}
						}
					}
				}
			}

			cbTempCard[i]--
		}
		return true
	} //end if赖子数为1

	return true
}

//连6连9判断
func (self *BaseLogic) CheckSeries(cbCardIndex []byte, cbCurrentCard byte, seriesnum int) bool {
	if seriesnum < 6 || seriesnum > 9 {
		return false
	}

	cbAllCardIndex := make([]byte, static.MAX_INDEX)
	copy(cbAllCardIndex, cbCardIndex[:])
	if cbCurrentCard != 0 {
		cbAllCardIndex[self.SwitchToCardIndex(cbCurrentCard)]++
	}

	for i := 0; i < 3; i++ {
		seriesbeginindex := i * 9
		seriesmaxbeginindex := i*9 + 9 - seriesnum
		for j := 0; j < 9; j++ {
			if seriesbeginindex > seriesmaxbeginindex {
				break
			}
			haveseries := true
			for k := seriesbeginindex; k < seriesbeginindex+seriesnum; k++ {
				if cbAllCardIndex[k] <= 0 {
					haveseries = false
					break
				}
			}
			if haveseries {
				cbCardIndexTemp := make([]byte, static.MAX_INDEX)
				copy(cbCardIndexTemp, cbAllCardIndex[:])

				for k := seriesbeginindex; k < seriesbeginindex+seriesnum; k++ {
					cbCardIndexTemp[k] = cbCardIndexTemp[k] - 1
				}

				_, AnalyseItemArray := self.AnalyseCard(cbCardIndexTemp, []static.TagWeaveItem{}, 0)
				if len(AnalyseItemArray) > 0 {
					return true
				} else {
					seriesbeginindex++
				}
			} else {
				seriesbeginindex++
			}
		}
	}

	return false
}

//混一色判断,手牌只有一种花色和字牌组成
func (self *BaseLogic) IsHunYiSe(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) bool {
	//判断是否有字牌
	haveZiPai := false
	for i := 27; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] > 0 {
			haveZiPai = true
			break
		}
	}

	//判断手牌是不是同一色花
	var colorUsed [3]int = [3]int{0, 0, 0}
	for i := 0; i < 27; i++ {
		if cbCardIndex[i] > 0 {
			recordColor := i / 9
			colorUsed[recordColor] = 1
		}
	}

	//判断吃碰杠是否是一色
	for i := 0; i < int(cbWeaveCount); i++ {
		recordColor := int((WeaveItem[i].CenterCard & static.MASK_COLOR) >> 4)
		if recordColor >= 0 && recordColor <= 2 {
			colorUsed[recordColor] = 1
		}

		if recordColor == 3 {
			haveZiPai = true
		}
	}

	if !haveZiPai {
		return false
	}

	useCount := colorUsed[0] + colorUsed[1] + colorUsed[2]

	if useCount == 1 {
		return true
	} else {
		return false
	}
}

//硬缺,筒条万缺一门,缺两门,缺三门,可以有风
func (self *BaseLogic) IsYingQue(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) bool {
	//判断手牌是不是同一色花
	var colorUsed [3]int = [3]int{0, 0, 0}
	for i := 0; i < 27; i++ {
		if cbCardIndex[i] > 0 {
			recordColor := i / 9
			colorUsed[recordColor] = 1
		}
	}

	//判断吃碰杠是否是一色
	for i := 0; i < int(cbWeaveCount); i++ {
		recordColor := int((WeaveItem[i].CenterCard & static.MASK_COLOR) >> 4)
		if recordColor >= 0 && recordColor <= 2 {
			colorUsed[recordColor] = 1
		}
	}

	useCount := colorUsed[0] + colorUsed[1] + colorUsed[2]

	if useCount <= 2 {
		return true
	} else {
		return false
	}
}

//硬缺,手上只有2色牌，没风将牌
// 0 :非缺牌型
// 1 ：硬缺（手上只有2色牌，没风将牌）
// 2 ：软缺（手上有2色牌，另有风将牌）
func (self *BaseLogic) IsYingQue_v2(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte) byte {
	//判断手牌是不是同一色花
	var colorUsed [3]int = [3]int{0, 0, 0}
	bFen := false
	for i := 0; i < static.MAX_INDEX; i++ {
		if cbCardIndex[i] > 0 {
			if i >= 27 {
				bFen = true
				break
			}
			recordColor := i / 9
			colorUsed[recordColor] = 1
		}
	}

	//判断吃碰杠是否是一色
	for i := 0; i < int(cbWeaveCount); i++ {
		recordColor := int((WeaveItem[i].CenterCard & static.MASK_COLOR) >> 4)
		if recordColor >= 0 && recordColor <= 2 {
			colorUsed[recordColor] = 1
		} else {
			bFen = true
		}
	}

	useCount := colorUsed[0] + colorUsed[1] + colorUsed[2]

	if useCount == 2 {
		if bFen {
			return 2
		}
		return 1
	} else {
		return 0
	}
}

//补全手牌
func (self *BaseLogic) CardIndexFill(cbTempCard []byte, WeaveItem []static.TagWeaveItem, cbWeaveCount byte, _weaveKind byte) []byte {
	cbCardIndexTemp := make([]byte, static.MAX_INDEX, static.MAX_INDEX)

	copy(cbCardIndexTemp, cbTempCard)

	//判断吃碰杠
	var cbCardBuffer [4]byte
	_count := byte(0)
	for i := 0; i < int(cbWeaveCount); i++ {
		//recordColor := int((WeaveItem[i].CenterCard & public.MASK_COLOR) >> 4)
		if _weaveKind&WeaveItem[i].WeaveKind != 0 && self.IsValidCard(WeaveItem[i].CenterCard) { //recordColor >= 0 && recordColor <= 2 {

			_count, cbCardBuffer = self.GetWeaveCard(WeaveItem[i].WeaveKind, WeaveItem[i].CenterCard, cbCardBuffer)
			for i := byte(0); i < _count; i++ {
				cbCardIndexTemp[self.SwitchToCardIndex(cbCardBuffer[i])]++
			}
		}
	}

	return cbCardIndexTemp
}

//获取有钱风0x31
func (self *BaseLogic) GetYouQianFeng(quanfeng byte) byte {
	switch quanfeng {
	case meta2.TYPE_DIRECATION_EAST: //东风圈
		return 0x31
	case meta2.TYPE_DIRECATION_SOUTH: //南风圈
		return 0x32
	case meta2.TYPE_DIRECATION_WEST: //西风圈
		return 0x33
	case meta2.TYPE_DIRECATION_NORTH: //北风圈
		return 0x34
	default:
		return static.INVALID_BYTE
	}
}

//获取门风
func (self *BaseLogic) GetMengFeng(bank uint16, _seat uint16) byte {
	return self.GetYouQianFeng(byte(meta2.TYPE_DIRECATION_EAST + (_seat+4-bank)%4))
}

//获取门风根据人头
func (self *BaseLogic) GetMengFengbyPlayerNum(bank uint16, _seat uint16, PlayerNum uint16) byte {
	return self.GetYouQianFeng(byte(meta2.TYPE_DIRECATION_EAST + (_seat+PlayerNum-bank)%PlayerNum))
}
