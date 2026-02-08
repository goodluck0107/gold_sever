//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  纸牌游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package JingZhouHuaPai

//import "fmt"

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"math/rand"
	"strconv"
	"strings"
)

//牌型设计
//上、大、人、孔、乙、己、可、知、礼、化、三、千、七、十、士、八、九、子、二、四、五、六,花乙,花三,花五,花七,花九,别杠
var zp_tcgz_strCardsMessage = [TCGZ_MAX_INDEX_HUA]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09", "0x0A", "0x0B",
	"0x0C", "0x0D", "0x0E", "0x0F", "0x10", "0x11", "0x12", "0x13", "0x14", "0x15", "0x16",
	"0x17", "0x18", "0x19", "0x1A", "0x1B", "0x1C",
}
var zp_tcgz_strCardsMessage1 = [TCGZ_MAX_INDEX_HUA]string{
	"上", "大", "人", "孔", "乙", "己", "可", "知", "礼", "化", "三",
	"千", "七", "十", "士", "八", "九", "子", "二", "四", "五", "六",
	"花乙", "花三", "花五", "花七", "花九", "混",
}

//扑克数据,已经用花牌替换了素牌，位置随机的
var zp_tcgz_cbCardDataArray = [TCGZ_ALLCARD]byte{
	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B,
	0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,

	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B,
	0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,

	0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B,
	0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16,

	0x01, 0x02, 0x03, 0x04, 0x17, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x18,
	0x0C, 0x1A, 0x0E, 0x0F, 0x10, 0x1B, 0x12, 0x13, 0x14, 0x19, 0x16,

	0x01, 0x02, 0x03, 0x04, 0x17, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x18,
	0x0C, 0x1A, 0x0E, 0x0F, 0x10, 0x1B, 0x12, 0x13, 0x14, 0x19, 0x16,

	0x1C, 0x1C,
}

//牌索引：上0、大1、人2、孔3、乙4、己5、可6、知7、礼8、化9、三10、千11、七12、十13、士14、八15、九16、子17、二18、四19、五20、六21
//句子牌和顺序：上大人、孔乙己、可知礼、化三千、七十士、八九子、乙二三四五六七八九十
//每张牌的半句牌索引，避免重复，只往后找，比如下标为0表示上的半句牌有上大和人（0、1、2），乙（4）的半句牌有乙己、二、三（4、5、18、10）
var zp_tcgz_CardsHalf = [TCGZ_MAX_INDEX][]byte{
	{0, 1, 2},
	{1, 2},
	{2},
	{3, 4, 5},
	{4, 5, 18, 10},
	{5},
	{6, 7, 8},
	{7, 8},
	{8},
	{9, 10, 11},
	{10, 11, 18, 19, 20},
	{11},
	{12, 13, 14, 15, 16, 20, 21},
	{13, 14, 15, 16},
	{14},
	{15, 16, 17, 21},
	{16, 17},
	{17},
	{18, 19},
	{19, 20, 21},
	{20, 21},
	{21},
}

//可以组成常规顺子的首牌位置
var zp_tcgz_CardsStraight = [TCGZ_MAX_INDEX]byte{
	1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0,
}

//可以组成跳跃顺子的所有牌位置
var zp_tcgz_Cards123 = [TCGZ_MAX_INDEX][]byte{
	{},
	{},
	{},
	{},
	{4, 18, 10},
	{},
	{},
	{},
	{},
	{},
	{10, 19, 20},
	{},
	{12, 15, 16},
	{},
	{},
	{15, 16, 13},
	{},
	{},
	{18, 10, 19},
	{19, 20, 21},
	{20, 21, 12},
	{21, 12, 15},
	{},
}

//可以组成跳跃顺子的所有牌位置  如果有赖子时这里总是漏掉牌型，可以把一个数字牌索引相关的所有组合都加上，比如索引10表示三，乙二三、二三四、三四五的都加上
var zp_tcgz_Cards123new = [TCGZ_MAX_INDEX][][3]byte{
	{},
	{},
	{},
	{},
	{{4, 18, 10}},
	{},
	{},
	{},
	{},
	{},
	{{4, 18, 10}, {18, 10, 19}, {10, 19, 20}},
	{},
	{{20, 21, 12}, {21, 12, 15}, {12, 15, 16}},
	{{15, 16, 13}},
	{},
	{{21, 12, 15}, {12, 15, 16}, {15, 16, 13}},
	{},
	{},
	{{18, 10, 19}},
	{{19, 20, 21}, {10, 19, 20}},
	{{20, 21, 12}, {10, 19, 20}, {19, 20, 21}},
	{{19, 20, 21}, {20, 21, 12}},
	{},
}

const (
	//逻辑掩码
	MASK_COLOR = 0xF0 //花色掩码
	MASK_VALUE = 0x0F //数值掩码
)

//动作定义
const (
	ZP_WIK_NULL       = 0x00    //没有类型
	ZP_WIK_LEFT       = 0x01    //左吃类型
	ZP_WIK_CENTER     = 0x02    //中吃类型
	ZP_WIK_RIGHT      = 0x04    //右吃类型
	ZP_WIK_PENG       = 0x08    //碰牌类型
	ZP_WIK_JIAO_1X_2D = 0x10    //绞牌吃牌类型(1小2大)
	ZP_WIK_GANG       = 0x20    //杠牌类型,通城个子里面这里是招
	ZP_WIK_CHI_HU     = 0x40    //吃胡类型
	ZP_WIK_QIANG      = 0x80    //抢暗杠类型
	ZP_WIK_FILL       = 0x100   //补牌类型
	ZP_WIK_JIAO_2X_1D = 0x200   //绞牌吃牌类型(2小1大)
	ZP_WIK_2710       = 0x400   //二七十吃牌类型
	ZP_WIK_KAN        = 0x800   //坎（没有动作）
	ZP_WIK_TIANLONG   = 0x1000  //天拢（没有动作）,通城个子里面这里是观
	ZP_WIK_OUT        = 0x2000  //出牌，大冶字牌有的时候需要用户选择是否出牌，服务器发给客户端时表示需要选择，客户端发给服务器时表示选择了出牌
	ZP_WIK_OUT_NULL   = 0x4000  //选择出牌时，选择了放弃
	ZP_WIK_HUA        = 0x8000  //通城个子，滑，手中有4张牌，又发来或别人打出第5张时可以滑
	ZP_WIK_HALF       = 0x10000 //通城个子，半句，两张扯得上关系的牌
	ZP_WIK_JIAN       = 0x20000 //通城个子，捡
	ZP_WIK_TA         = 0x40000 //荆州花牌，踏
)

type GameLogic_zp_jzhp struct {
	m_AnalyseItemArray []TagAnalyseItem

	Rule rule2.St_FriendRule

	m_allMagicPoint []uint8 //所有的赖子，初始化时需要设置好
	HunJiang        int     //混江
}

//20181214 每个玩法设定自己的限时操作时间
func (self *GameLogic_zp_jzhp) Setlimitetime(limitetimeOp bool) {
	//if limitetimeOp {
	//	self.Rule.limitetimeOP = GameTime_Nine
	//} else {
	//	self.Rule.limitetimeOP = 0
	//}
}

//混乱扑克
func (self *GameLogic_zp_jzhp) RandCardData() (byte, [TCGZ_ALLCARD]byte) {
	//混乱准备
	cbCardData := [TCGZ_ALLCARD]byte{}
	cbCardDataTemp := make([]byte, TCGZ_ALLCARD, TCGZ_ALLCARD)
	var cbMaxCount byte = TCGZ_ALLCARD
	if self.HunJiang == 0 {
		cbMaxCount -= 2
	}
	copy(cbCardDataTemp, zp_tcgz_cbCardDataArray[:cbMaxCount])

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

	return cbMaxCount, cbCardData
}

//有效判断
func (self *GameLogic_zp_jzhp) IsValidCard(cbCardData byte) bool {
	//校验
	bIsValid := ((cbCardData >= 1) && (cbCardData <= 28))
	if !bIsValid {
		return false //非法牌
	}

	return true
}

//设置最大人数
func (self *GameLogic_zp_jzhp) SetHunJiang(iHunJiang int) {
	if (iHunJiang >= 0) && (iHunJiang <= 1) {
		self.HunJiang = iHunJiang
	} else {
		self.HunJiang = 0
	}
}

//初始化赖子列表,byMagicPoint的值约束： 0x1C是赖子
func (self *GameLogic_zp_jzhp) InitMagicPoint(byMagicPoint byte) {
	self.m_allMagicPoint = []byte{}
	if byMagicPoint >= 1 && byMagicPoint <= 0x1C {
		self.m_allMagicPoint = append(self.m_allMagicPoint, byMagicPoint)
	}
}

//设置赖子值,byMagicPoint的值约束： 0x19是赖子
func (self *GameLogic_zp_jzhp) AddMagicPoint(byMagicPoint byte) {
	if (byMagicPoint >= 1) && (byMagicPoint <= 0x1C) {
		self.m_allMagicPoint = append(self.m_allMagicPoint, byMagicPoint)
	}
}

//当前牌是否是赖子
func (self *GameLogic_zp_jzhp) IsMagic(cardData byte) bool {
	for _, v := range self.m_allMagicPoint {
		if cardData == v {
			return true
		}
	}
	return false
}

//不带花的索引
func (self *GameLogic_zp_jzhp) IsMagicIndex(cardIndex byte) bool {
	if cardIndex == TCGZ_MAX_INDEX-1 {
		cardIndex = TCGZ_MAX_INDEX_HUA - 1
	}
	for _, v := range self.m_allMagicPoint {
		if self.SwitchToCardData(cardIndex) == v {
			return true
		}
	}
	return false
}

//带花的索引
func (self *GameLogic_zp_jzhp) IsMagicIndexHua(cardIndex byte) bool {
	for _, v := range self.m_allMagicPoint {
		if self.SwitchToCardData(cardIndex) == v {
			return true
		}
	}
	return false
}

//当前牌是否全是赖子,card_list是索引
func (self *GameLogic_zp_jzhp) IsAllMagic(card_list []byte, cardlen int) bool {
	for i := 0; i < cardlen && i < TCGZ_MAX_INDEX && i < int(len(card_list)); i++ {
		if card_list[i] > 0 && !self.IsMagicIndex(byte(i)) {
			return false
		}
	}
	return true
}

//获取赖子的数量,card_list是索引
func (self *GameLogic_zp_jzhp) GetMagicNum(card_list []byte, cardlen int, fakepoker []static.TFakePoker) (int, []static.TFakePoker) {
	num := 0
	for i := 0; i < cardlen && i < TCGZ_MAX_INDEX && i < int(len(card_list)); i++ {
		if card_list[i] > 0 && self.IsMagicIndex(byte(i)) {
			if fakepoker != nil {
			}
			num += int(card_list[i])
		}
	}
	return num, fakepoker
}
func (self *GameLogic_zp_jzhp) ClearMagic(card_list []byte, cardlen int) (recardlist [TCGZ_MAX_INDEX]byte) {
	for i := 0; i < cardlen && i < TCGZ_MAX_INDEX && i < int(len(card_list)); i++ {
		if card_list[i] > 0 && self.IsMagicIndex(byte(i)) {
			recardlist[i] = 0
		} else {
			recardlist[i] = card_list[i]
		}
	}
	return recardlist
}

/********************************从C++项目复制过来********************************************************************/
//扑克转换成牌数据
func (self *GameLogic_zp_jzhp) SwitchToCardData(cbCardIndex byte) byte {
	if cbCardIndex < TCGZ_MAX_INDEX_HUA {
		return cbCardIndex + 1
	} else {
		return 0
	}
}

//扑克数据转换成扑克点数转换的结果：
func (self *GameLogic_zp_jzhp) SwitchToCardIndex(cbCardData byte) byte {
	if !self.IsValidCard(cbCardData) {
		return 255
	}
	if cbCardData <= 0x1C && cbCardData > 0 {
		return cbCardData - 1
	} else {
		return 255
	}
}

//扑克数据转换成扑克点数转换的结果：不考虑花牌
func (self *GameLogic_zp_jzhp) SwitchToCardIndexNoHua(cbCardData byte) byte {
	if !self.IsValidCard(cbCardData) {
		return 0 //为避免崩溃，这里返回的0
	}
	if cbCardData <= 0x16 && cbCardData > 0 {
		return cbCardData - 1
	} else if cbCardData == 0x17 {
		return 4
	} else if cbCardData == 0x18 {
		return 10
	} else if cbCardData == 0x19 {
		return 20
	} else if cbCardData == 0x1A {
		return 12
	} else if cbCardData == 0x1B {
		return 16
	} else if cbCardData == 0x1C {
		return 22
	} else {
		return 255
	}
}

func (self *GameLogic_zp_jzhp) SwitchToCardData2(cbCardIndex [TCGZ_MAX_INDEX]byte) (byte, [TCGZ_MAXHANDCARD]byte) {
	//转换扑克
	cbCardData := [TCGZ_MAXHANDCARD]byte{}
	cbPosition := byte(0)
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ { // 20种牌，每种牌有几张
		if cbCardIndex[i] != 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				if cbPosition >= TCGZ_MAXHANDCARD {
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
func (self *GameLogic_zp_jzhp) SwitchToCardIndex2(cbCardData []byte, cbCardCount byte) (byte, [TCGZ_MAX_INDEX]byte) {
	//设置变量
	cbCardIndex := [TCGZ_MAX_INDEX]byte{}
	//转换扑克
	for i := byte(0); i < cbCardCount; i++ {
		if !self.IsValidCard(cbCardData[i]) {
			return cbCardCount, cbCardIndex
		}
		cbCardIndex[self.SwitchToCardIndexNoHua(cbCardData[i])]++
	}

	return cbCardCount, cbCardIndex
}

func (self *GameLogic_zp_jzhp) SwitchToCardData3(cbCardIndex [TCGZ_MAX_INDEX_HUA]byte) (byte, [TCGZ_MAXHANDCARD]byte) {
	//转换扑克
	cbCardData := [TCGZ_MAXHANDCARD]byte{}
	cbPosition := byte(0)
	for i := byte(0); i < TCGZ_MAX_INDEX_HUA; i++ { // 20种牌，每种牌有几张
		if cbCardIndex[i] != 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				if cbPosition >= TCGZ_MAXHANDCARD {
					xlog.Logger().Errorln("SwitchToCardData3 error cardId :" + strconv.Itoa(int(cbPosition)))
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
func (self *GameLogic_zp_jzhp) SwitchToCardIndex3(cbCardData []byte, cbCardCount byte) (byte, [TCGZ_MAX_INDEX_HUA]byte) {
	//设置变量
	cbCardIndex := [TCGZ_MAX_INDEX_HUA]byte{}
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

//把带有花牌的index转换成没有花的index，
func (self *GameLogic_zp_jzhp) SwitchIndexToIndexNoHua(cbCardIndex [TCGZ_MAX_INDEX_HUA]byte) (byte, [TCGZ_MAX_INDEX]byte) {
	//转换扑克
	rtCardIndex := [TCGZ_MAX_INDEX]byte{}
	cbPosition := byte(0)
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ { // 27种牌，每种牌有几张
		rtCardIndex[i] = cbCardIndex[i]
	}
	rtCardIndex[4] += cbCardIndex[22]
	rtCardIndex[10] += cbCardIndex[23]
	rtCardIndex[20] += cbCardIndex[24]
	rtCardIndex[12] += cbCardIndex[25]
	rtCardIndex[16] += cbCardIndex[26]
	rtCardIndex[22] = cbCardIndex[27]

	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		cbPosition += rtCardIndex[i]
	}
	return cbPosition, rtCardIndex
}

//删除单张扑克
func (self *GameLogic_zp_jzhp) RemoveCard(cbCardIndex [TCGZ_MAX_INDEX_HUA]byte, cbRemoveCard byte) (bool, [TCGZ_MAX_INDEX_HUA]byte) {
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
func (self *GameLogic_zp_jzhp) RemoveCard2(cbCardIndex [TCGZ_MAX_INDEX_HUA]byte, cbRemoveCard []byte, cbRemoveCount byte) (bool, [TCGZ_MAX_INDEX_HUA]byte) {
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

//扑克数目
func (self *GameLogic_zp_jzhp) GetCardCount(cbCardIndex []byte) byte {
	//数目统计
	cbCardCount := 0
	for i := 0; i < len(cbCardIndex) && i < TCGZ_MAX_INDEX_HUA; i++ {
		cbCardCount += int(cbCardIndex[i])
	}
	return byte(cbCardCount)
}

//动作等级
func (self *GameLogic_zp_jzhp) GetUserActionRank(cbUserAction int) int {

	//胡牌等级
	if cbUserAction&ZP_WIK_CHI_HU != 0 {
		return 8
	}
	//踏牌等级
	if cbUserAction&(ZP_WIK_TA) != 0 {
		return 7
	}
	//滑牌等级
	if cbUserAction&(ZP_WIK_HUA) != 0 {
		return 6
	}

	//杠牌等级
	if cbUserAction&(ZP_WIK_GANG) != 0 || cbUserAction&(ZP_WIK_TIANLONG) != 0 {
		return 5
	}

	//碰牌等级
	if cbUserAction&ZP_WIK_PENG != 0 {
		return 3
	}

	//上牌等级
	if cbUserAction&(ZP_WIK_RIGHT|ZP_WIK_CENTER|ZP_WIK_LEFT|ZP_WIK_JIAN) != 0 {
		return 1
	}

	return 0
}

//删除一些不能参与分析的牌，观的牌不能拆，捡的牌不能参入操作响应
func (self *GameLogic_zp_jzhp) DeleteSomeCards(cbCardIndex [TCGZ_MAX_INDEX]byte, delcardI byte, delcnt byte) [TCGZ_MAX_INDEX]byte {
	//参数效验
	if delcardI < TCGZ_MAX_INDEX {
		if cbCardIndex[delcardI] >= delcnt {
			cbCardIndex[delcardI] -= delcnt
		}
	}
	return cbCardIndex
}

//索引是不是统牌的索引，入参cbCardIndex会被修改；返回0表示不是，1表示可以组成1个统，2表示可以组成2个统
func (self *GameLogic_zp_jzhp) IsTongCardIndex(cbCardIndex [TCGZ_MAX_INDEX]byte, cbIndex byte) (int, [TCGZ_MAX_INDEX]byte) {
	tongcnt := 0
	if cbIndex >= TCGZ_MAX_INDEX {
		return tongcnt, cbCardIndex
	}
	if cbCardIndex[cbIndex] >= 5 {
		tongcnt = 2
	} else if cbCardIndex[cbIndex] == 4 && 1 != self.IsJingPaiIndex(cbIndex) {
		tongcnt = 1
	} else if cbCardIndex[cbIndex] == 4 && 1 == self.IsJingPaiIndex(cbIndex) {
		tongcnt = 1
		if cbCardIndex[22] >= 1 {
			cbCardIndex[22]-- //使用了一个赖子
			tongcnt = 2
		}
	} else if cbCardIndex[cbIndex] >= 2 && 1 == self.IsJingPaiIndex(cbIndex) && cbCardIndex[cbIndex]+cbCardIndex[22] >= 4 {
		tongcnt = 1
		delecnt := byte(0)
		if cbCardIndex[cbIndex] < 4 {
			delecnt = 4 - cbCardIndex[cbIndex]
		}
		if cbCardIndex[cbIndex]+cbCardIndex[22] >= 5 {
			tongcnt = 2
			if cbCardIndex[cbIndex] < 5 {
				delecnt = 5 - cbCardIndex[cbIndex]
			}
		}
		cbCardIndex[22] -= delecnt
	}
	return tongcnt, cbCardIndex
}
func (self *GameLogic_zp_jzhp) IsTongCard(cbCardIndex [TCGZ_MAX_INDEX]byte, cbCurrentCard byte) int {
	if !self.IsValidCard(cbCurrentCard) {
		return 0
	}
	cbIndex := self.SwitchToCardIndexNoHua(cbCurrentCard)
	ret, _ := self.IsTongCardIndex(cbCardIndex, cbIndex)
	return ret
}

//操作是否满足统的要求 v1.5 20200715 泛不换统后，需要确定换到哪里去了
func (self *GameLogic_zp_jzhp) CheckTong_Op(cbCardIndex [TCGZ_MAX_INDEX]byte, curTongCnt int) (bool, int, [TCGZ_MAX_INDEX]int) {
	//理论上多少个统
	realTongCnt := 0
	//IsTongCardIndex函数会修改cbCardIndex数组，需要临时数组
	CardIndexTemp := [TCGZ_MAX_INDEX]byte{}
	copy(CardIndexTemp[:], cbCardIndex[:])
	//三五七，要优先找数目多的
	split := self.CheckMaxIndex(CardIndexTemp)
	//统计理论上的统数
	tongInfo := [TCGZ_MAX_INDEX]int{}
	for i := byte(0); i < TCGZ_MAX_INDEX-1; i++ {
		retcnt := 0
		//三五七，要优先找数目多的
		byi := self.SwitchIndex(split, i)
		retcnt, CardIndexTemp = self.IsTongCardIndex(CardIndexTemp, byi)
		realTongCnt += retcnt
		tongInfo[byi] = retcnt
	}
	if realTongCnt > curTongCnt {
		return true, realTongCnt, tongInfo
	}
	return false, realTongCnt, tongInfo
}

//三五七，要优先找数目多的
func (self *GameLogic_zp_jzhp) SwitchIndex(split [3][2]byte, cbIndex byte) byte {

	if cbIndex >= TCGZ_MAX_INDEX-1 {
		return TCGZ_MAX_INDEX - 2
	}
	if cbIndex == 10 {
		return split[0][1]
	} else if cbIndex == 12 {
		return split[1][1]
	} else if cbIndex == 20 {
		return split[2][1]
	} else {
		return cbIndex
	}
}

//三五七，要优先找数目多的
func (self *GameLogic_zp_jzhp) CheckMaxIndex(cbCardIndex [TCGZ_MAX_INDEX]byte) [3][2]byte {
	retSlipt := [3][2]byte{{10, 10}, {12, 12}, {20, 20}}
	if cbCardIndex[22] == 0 {
		return retSlipt
	}
	if cbCardIndex[retSlipt[1][1]] > cbCardIndex[retSlipt[0][1]] {
		retSlipt[0][1], retSlipt[1][1] = retSlipt[1][1], retSlipt[0][1]
	}
	if cbCardIndex[retSlipt[2][1]] > cbCardIndex[retSlipt[0][1]] {
		retSlipt[0][1], retSlipt[2][1] = retSlipt[2][1], retSlipt[0][1]
	}
	if cbCardIndex[retSlipt[2][1]] > cbCardIndex[retSlipt[1][1]] {
		retSlipt[1][1], retSlipt[2][1] = retSlipt[2][1], retSlipt[1][1]
	}
	return retSlipt
}

//返回值大于0表示有冲突，需要把357按照数目排序后，把数目少的换成数目多的
func (self *GameLogic_zp_jzhp) Check357Conflict(cbCardIndex [TCGZ_MAX_INDEX]byte, TongInfo TagTongInfo) int {
	CardIndexTemp := [TCGZ_MAX_INDEX]byte{}
	copy(CardIndexTemp[:], cbCardIndex[:])
	//三五七，要优先找数目多的
	split := self.CheckMaxIndex(CardIndexTemp)
	needHunCnt := byte(0)
	for i := 0; i < 3; i++ {
		curIndex := split[i][1]
		if TongInfo.CardTongInfo[curIndex].TongCnt > 0 {
			if TongInfo.CardTongInfo[curIndex].TongCnt >= 2 {
				if 5 >= CardIndexTemp[curIndex] {
					needHunCnt += 5 - CardIndexTemp[curIndex]
				}
			} else if TongInfo.CardTongInfo[curIndex].TongCnt == 1 {
				if 4 >= CardIndexTemp[curIndex] {
					needHunCnt += 4 - CardIndexTemp[curIndex]
				}
			}
		}
	}
	if needHunCnt > cbCardIndex[22] {
		return int(needHunCnt - cbCardIndex[22])
	}
	return 0
}

//碰牌判断
func (self *GameLogic_zp_jzhp) EstimatePengCard(cbCardIndex [TCGZ_MAX_INDEX]byte, cbCurrentCard byte, curTongCnt int) int {
	//参数效验
	if !self.IsValidCard(cbCurrentCard) || cbCurrentCard >= 0x1C {
		return ZP_WIK_NULL
	}

	//碰牌判断
	if cbCardIndex[self.SwitchToCardIndexNoHua(cbCurrentCard)] >= 2 {
		//碰要删除2张牌，需要临时数组
		CardIndexTemp := [TCGZ_MAX_INDEX]byte{}
		copy(CardIndexTemp[:], cbCardIndex[:])
		_, realTongCntBef, _ := self.CheckTong_Op(CardIndexTemp, curTongCnt)
		CardIndexTemp[self.SwitchToCardIndexNoHua(cbCurrentCard)] -= 2
		_, realTongCnt, _ := self.CheckTong_Op(CardIndexTemp, curTongCnt)
		if realTongCnt < curTongCnt {
			//统过不能在碰，可以换统的除外
			if realTongCntBef > realTongCnt {
				//因为这张牌导致的统数减少
				return ZP_WIK_NULL
			}
		}
		return ZP_WIK_PENG
	} else {
		return ZP_WIK_NULL
	}
}

//杠牌判断  招
func (self *GameLogic_zp_jzhp) EstimateGangCard(cbCardIndex [TCGZ_MAX_INDEX]byte, cbCurrentCard byte, curTongCnt int) int {
	//参数效验
	if !self.IsValidCard(cbCurrentCard) || cbCurrentCard >= 0x1C {
		return ZP_WIK_NULL
	}

	//杠牌判断
	if cbCardIndex[self.SwitchToCardIndexNoHua(cbCurrentCard)] >= 3 {
		//招要删除3张牌，需要临时数组
		CardIndexTemp := [TCGZ_MAX_INDEX]byte{}
		copy(CardIndexTemp[:], cbCardIndex[:])
		_, realTongCntBef, _ := self.CheckTong_Op(CardIndexTemp, curTongCnt)
		CardIndexTemp[self.SwitchToCardIndexNoHua(cbCurrentCard)] -= 3
		_, realTongCnt, _ := self.CheckTong_Op(CardIndexTemp, curTongCnt)
		if realTongCnt < curTongCnt {
			//统过不能在招，可以换统的除外
			if realTongCntBef > realTongCnt {
				//因为这张牌导致的统数减少
				return ZP_WIK_NULL
			}
		}
		return ZP_WIK_GANG
	} else {
		return ZP_WIK_NULL
	}
}

//杠牌分析 统
func (self *GameLogic_zp_jzhp) AnalyseGangCard(cbCardIndex [TCGZ_MAX_INDEX]byte, curTongCnt int) (int, TagGangCardResult) {
	//设置变量
	cbActionMask := ZP_WIK_NULL
	var GangCardResult TagGangCardResult
	////手上杠牌
	//for  i:=byte(0);i<TCGZ_MAX_INDEX;i++{
	//	if (cbCardIndex[i]>=4 ) {
	//		cbActionMask|=ZP_WIK_TIANLONG;
	//		GangCardResult.CardData[GangCardResult.CardCount]=self.SwitchToCardData(i);
	//		GangCardResult.CardCount++
	//	}
	//}
	_, realTongCnt, _ := self.CheckTong_Op(cbCardIndex, curTongCnt)
	if realTongCnt > curTongCnt {
		cbActionMask |= ZP_WIK_TIANLONG
	}
	return cbActionMask, GangCardResult
}

//滑牌判断 泛  v1.5 20200713 要考虑统
func (self *GameLogic_zp_jzhp) EstimateHuaCard(cbCardIndex [TCGZ_MAX_INDEX]byte, cbCurrentCard byte, curTongCnt int) int {
	//参数效验
	if !self.IsValidCard(cbCurrentCard) || cbCurrentCard >= 0x1C {
		return ZP_WIK_NULL
	}

	//杠牌判断
	if cbCardIndex[self.SwitchToCardIndexNoHua(cbCurrentCard)] >= 4 {
		//别人打的牌开泛不受统的限制，但统数可能会减少
		//v1.5 20200713 要考虑统
		if curTongCnt == 0 {
			return ZP_WIK_NULL
		}
		return ZP_WIK_HUA
	} else {
		return ZP_WIK_NULL
	}
}

//滑牌分析  泛
func (self *GameLogic_zp_jzhp) AnalyseHuaCard(cbCardIndex [TCGZ_MAX_INDEX]byte) (int, TagGangCardResult) {
	//设置变量
	cbActionMask := ZP_WIK_NULL
	var GangCardResult TagGangCardResult
	//手上杠牌
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {

		if cbCardIndex[i] >= 5 {
			//做牌时可能会做很多个，这里将==该成>=
			cbActionMask |= ZP_WIK_HUA
			GangCardResult.CardData[GangCardResult.CardCount] = self.SwitchToCardData(i)
			GangCardResult.CardCount++
		}
	}

	return cbActionMask, GangCardResult
}

//踏牌分析
func (self *GameLogic_zp_jzhp) AnalyseTaCard(cbCardIndex [TCGZ_MAX_INDEX]byte, WeaveItem [10]TagWeaveItem, cbWeaveCount byte) (int, TagGangCardResult) {
	//设置变量
	cbActionMask := ZP_WIK_NULL
	var GangCardResult TagGangCardResult
	//组合杠牌
	for i := byte(0); i < cbWeaveCount; i++ {
		if WeaveItem[i].WeaveKind == static.WIK_GANG {
			if cbCardIndex[self.SwitchToCardIndexNoHua(WeaveItem[i].PublicCard)] == 1 {
				cbActionMask |= ZP_WIK_TA
				GangCardResult.CardData[GangCardResult.CardCount] = self.SwitchToCardData(i)
				GangCardResult.CardCount++
			}
		}
	}

	return cbActionMask, GangCardResult
}

//自摸的牌可以是坎，接炮的牌如果只能组成3张一样的，只能当作碰,cbCurrentCard有可能是0，需要新增一个参数cbCenterCard
func (self *GameLogic_zp_jzhp) AnalyseChiHuCard(cbCardIndexSrc [TCGZ_MAX_INDEX_HUA]byte, cbCardIndex [TCGZ_MAX_INDEX]byte, WeaveItem []TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight int) (int, TagChiHuResult) {
	//变量定义
	wChiHuKind := ZP_WIK_NULL
	var ChiHuResult TagChiHuResult
	AnalyseItemArray := []TagAnalyseItem{}

	if cbCurrentCard >= 0x1C {
		return ZP_WIK_NULL, ChiHuResult
	}

	//构造麻将
	cbCardIndexTemp := [TCGZ_MAX_INDEX]byte{}
	copy(cbCardIndexTemp[:], cbCardIndex[:])

	currentCards := byte(0)
	if cbCurrentCard != 0 {
		cbCardIndexTemp[self.SwitchToCardIndexNoHua(cbCurrentCard)]++
		cbCardIndexSrc[self.SwitchToCardIndex(cbCurrentCard)]++
		currentCards = cbCardIndexTemp[self.SwitchToCardIndexNoHua(cbCurrentCard)]
	}
	wMagicCount, _ := self.GetMagicNum(cbCardIndexTemp[:], TCGZ_MAX_INDEX, nil)

	//用AnalyseItemArray保存可以胡牌的所有组合
	if wMagicCount == 0 {
		_, AnalyseItemArray = self.AnalyseCard_new(cbCardIndexTemp, WeaveItem, cbWeaveCount, wChiHuRight)
	}

	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//牌型分析
		for i := 0; i < len(AnalyseItemArray); i++ {
			//变量定义
			var tChiHuItemInfo TagChiHuItemInfo
			pAnalyseItem := &AnalyseItemArray[i]
			tChiHuItemInfo.CardEye = pAnalyseItem.CardEye
			for j := 0; j < len(pAnalyseItem.WeaveKind) && j < 10; j++ {
				tChiHuItemInfo.WeaveCount = j
				if pAnalyseItem.WeaveKind[j] == ZP_WIK_NULL {
					break
				}
				tChiHuItemInfo.WeaveKind[j] = pAnalyseItem.WeaveKind[j]
				tChiHuItemInfo.WeaveItemInfo[j][0] = pAnalyseItem.Cards[j][0]
				tChiHuItemInfo.WeaveItemInfo[j][1] = pAnalyseItem.Cards[j][1]
				tChiHuItemInfo.WeaveItemInfo[j][2] = pAnalyseItem.Cards[j][2]
				tChiHuItemInfo.WeaveItemInfo[j][3] = pAnalyseItem.Cards[j][3]
				tChiHuItemInfo.WeaveItemInfo[j][4] = pAnalyseItem.Cards[j][4]
				tChiHuItemInfo.IsWeave[j] = pAnalyseItem.IsWeave[j]
			}
			if pAnalyseItem.WeaveKind[9] != ZP_WIK_NULL {
				//数组用满了时，这个长度少了一个
				tChiHuItemInfo.WeaveCount = 10
			}
			ChiHuResult.ChiHuItemInfoArray = append(ChiHuResult.ChiHuItemInfoArray, tChiHuItemInfo)
		}
		if len(ChiHuResult.ChiHuItemInfoArray) > 0 {
			wChiHuKind |= ZP_WIK_CHI_HU
		}
	}

	//结果判断
	if wChiHuKind != ZP_WIK_NULL {
		//设置结果
		ChiHuResult.ChiHuKind |= wChiHuKind
		ChiHuResult.ChiHuRight = wChiHuRight

		//恢复花牌数据
		cardIndex := [TCGZ_MAX_INDEX_HUA]byte{}
		for k := 0; k < len(ChiHuResult.ChiHuItemInfoArray); k++ {
			copy(cardIndex[:], cbCardIndexSrc[:])
			for i := 0; i < len(ChiHuResult.ChiHuItemInfoArray[k].WeaveItemInfo); i++ {
				if ChiHuResult.ChiHuItemInfoArray[k].IsWeave[i] == 1 {
					continue
				}
				for j := 0; j < 5; j++ {
					card := ChiHuResult.ChiHuItemInfoArray[k].WeaveItemInfo[i][j]
					if self.IsValidCard(card) {
						cardindex := self.SwitchToCardIndex(card)
						if cardIndex[cardindex] > 0 {
							cardIndex[cardindex]--
						} else {
							//换成parter
							if self.GetParterCard(card) != 0 {
								ChiHuResult.ChiHuItemInfoArray[k].WeaveItemInfo[i][j] = self.GetParterCard(card)
							}
						}
					}
				}
			}
		}
		if self.IsValidCard(cbCurrentCard) {
			//self.CheckHu_KanPeng(cbCurrentCard,&ChiHuResult)
			self.CheckHu_ZhaoShaoKanPeng(false, cbCurrentCard, currentCards, &ChiHuResult)
		}
		return ZP_WIK_CHI_HU, ChiHuResult
	}

	if wChiHuKind == ZP_WIK_NULL && wMagicCount != 0 && wMagicCount <= 4 {
		//调用赖子分析函数进行分析
		wChiHuKind, ChiHuResult = self.AnalyseMagicChiHuCard(cbCardIndexSrc, cbCardIndexTemp, WeaveItem, cbWeaveCount, 0, wChiHuRight)
		if wChiHuKind != ZP_WIK_NULL {
			if self.IsValidCard(cbCurrentCard) {
				//self.CheckHu_KanPeng(cbCurrentCard,&ChiHuResult)
				self.CheckHu_ZhaoShaoKanPeng(false, cbCurrentCard, currentCards, &ChiHuResult)
			}
			return ZP_WIK_CHI_HU, ChiHuResult
		}
	}

	return ZP_WIK_NULL, ChiHuResult
}

//赖子只能做三五七，优先做花牌，且要保证赖子组合的胡数最大（赖子，花牌，精牌优先组成高胡数的类型），
//自摸的牌可以是坎，接炮的牌如果只能组成3张一样的，只能当作碰
func (self *GameLogic_zp_jzhp) AnalyseMagicChiHuCard(cbCardIndexSrc [TCGZ_MAX_INDEX_HUA]byte, cbCardIndex [TCGZ_MAX_INDEX]byte, WeaveItem []TagWeaveItem, cbWeaveCount byte, cbCurrentCard byte, wChiHuRight int) (int, TagChiHuResult) {
	//变量定义
	wChiHuKind := ZP_WIK_NULL
	var ChiHuResult TagChiHuResult
	AnalyseItemArray := []TagAnalyseItem{}

	//构造麻将
	cbCardIndexTemp := [TCGZ_MAX_INDEX]byte{}
	copy(cbCardIndexTemp[:], cbCardIndex[:])

	if cbCurrentCard != 0 {
		cbCardIndexTemp[self.SwitchToCardIndex(cbCurrentCard)]++
	}

	wMagicCount, _ := self.GetMagicNum(cbCardIndexTemp[:], TCGZ_MAX_INDEX, nil)

	if wMagicCount <= 0 {
		return ZP_WIK_NULL, ChiHuResult
	}
	//清除赖子牌
	if wMagicCount > 0 {
		cbCardIndexTemp = self.ClearMagic(cbCardIndexTemp[:], TCGZ_MAX_INDEX)
	}

	//用AnalyseItemArray保存可以胡牌的所有组合
	_, AnalyseItemArray = self.AnalyseCard_Magic_new(cbCardIndexTemp, WeaveItem, cbWeaveCount, wChiHuRight, byte(wMagicCount))

	//胡牌分析
	if len(AnalyseItemArray) > 0 {
		//牌型分析
		for i := 0; i < len(AnalyseItemArray); i++ {
			//变量定义
			cbCardIndexAnalyse := [TCGZ_MAX_INDEX]byte{} //赖子只能当作三五七使用
			var tChiHuItemInfo TagChiHuItemInfo
			pAnalyseItem := &AnalyseItemArray[i]
			tChiHuItemInfo.CardEye = pAnalyseItem.CardEye
			for j := 0; j < len(pAnalyseItem.WeaveKind) && j < 10; j++ {
				tChiHuItemInfo.WeaveCount = j
				if pAnalyseItem.WeaveKind[j] == ZP_WIK_NULL {
					break
				}
				tChiHuItemInfo.WeaveKind[j] = pAnalyseItem.WeaveKind[j]
				tChiHuItemInfo.WeaveItemInfo[j][0] = pAnalyseItem.Cards[j][0]
				tChiHuItemInfo.WeaveItemInfo[j][1] = pAnalyseItem.Cards[j][1]
				tChiHuItemInfo.WeaveItemInfo[j][2] = pAnalyseItem.Cards[j][2]
				tChiHuItemInfo.WeaveItemInfo[j][3] = pAnalyseItem.Cards[j][3]
				tChiHuItemInfo.WeaveItemInfo[j][4] = pAnalyseItem.Cards[j][4]
				tChiHuItemInfo.IsWeave[j] = pAnalyseItem.IsWeave[j]
				if pAnalyseItem.WeaveKind[9] != ZP_WIK_NULL {
					//数组用满了时，这个长度少了一个
					tChiHuItemInfo.WeaveCount = 10
				}
				//赖子只能当作三五七使用
				if pAnalyseItem.IsWeave[j] == 0 {
					for k := 0; k < 5; k++ {
						card := pAnalyseItem.Cards[j][k]
						if self.IsValidCard(card) && self.SwitchToCardIndexNoHua(card) < TCGZ_MAX_INDEX {
							cbCardIndexAnalyse[self.SwitchToCardIndexNoHua(card)]++
						}
					}
				}
			}
			//赖子只能当作三五七使用  三的indexer10，五的indexer20，七的indexer12
			if cbCardIndexAnalyse[10] >= cbCardIndex[10] && cbCardIndexAnalyse[20] >= cbCardIndex[20] && cbCardIndexAnalyse[12] >= cbCardIndex[12] {
				if byte(wMagicCount) == (cbCardIndexAnalyse[10]-cbCardIndex[10])+(cbCardIndexAnalyse[20]-cbCardIndex[20])+(cbCardIndexAnalyse[12]-cbCardIndex[12]) {
					ChiHuResult.ChiHuItemInfoArray = append(ChiHuResult.ChiHuItemInfoArray, tChiHuItemInfo)
					wChiHuKind |= ZP_WIK_CHI_HU
				}
			}
		}
	}

	//结果判断
	if wChiHuKind != ZP_WIK_NULL {
		//设置结果
		ChiHuResult.ChiHuKind = wChiHuKind
		ChiHuResult.ChiHuRight = wChiHuRight
		//恢复花牌数据
		//先排序 ,句子在前(先把非花牌在句子中用掉，保证花牌在后面使用)
		self.SortChiHuResult(0, &ChiHuResult)
		cardIndex := [TCGZ_MAX_INDEX_HUA]byte{}
		for k := 0; k < len(ChiHuResult.ChiHuItemInfoArray); k++ {
			copy(cardIndex[:], cbCardIndexSrc[:])
			for i := 0; i < len(ChiHuResult.ChiHuItemInfoArray[k].WeaveItemInfo); i++ {
				if ChiHuResult.ChiHuItemInfoArray[k].IsWeave[i] == 1 {
					continue
				}
				for j := 0; j < 5; j++ {
					card := ChiHuResult.ChiHuItemInfoArray[k].WeaveItemInfo[i][j]
					if self.IsValidCard(card) {
						cardindex := self.SwitchToCardIndex(card)
						if cardIndex[cardindex] > 0 {
							cardIndex[cardindex]--
						} else {
							//换成parter
							if self.GetParterCard(card) != 0 {
								ChiHuResult.ChiHuItemInfoArray[k].WeaveItemInfo[i][j] = self.GetParterCard(card)
							}
						}
					}
				}
			}
		}
		return ZP_WIK_CHI_HU, ChiHuResult
	}
	return ZP_WIK_NULL, ChiHuResult
}

//胡牌详情排序，不排组合排区的牌。SortFlag 0句子在前，1句子在后，
func (self *GameLogic_zp_jzhp) SortChiHuResult(SortFlag int, ChiHuResult *TagChiHuResult) {
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	for byIndex := 0; byIndex < byCount; byIndex++ {
		for j := 0; j < ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveCount-1 && j < 9; j++ {
			for k := 0; k < ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveCount-1-j && k < 9-j; k++ {
				if self.GetWeaveLevel(SortFlag, ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[k], ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[k]) < self.GetWeaveLevel(SortFlag, ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[k+1], ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[k+1]) {
					//1
					weavekind := ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[k]
					ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[k] = ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[k+1]
					ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[k+1] = weavekind
					//2
					WeaveInfo := [5]byte{}
					copy(WeaveInfo[:], ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[k][:])
					copy(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[k][:], ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[k+1][:])
					copy(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[k+1][:], WeaveInfo[:])
					//3
					huxi := ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[k]
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[k] = ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[k+1]
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[k+1] = huxi
					//4
					isweave := ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[k]
					ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[k] = ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[k+1]
					ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[k+1] = isweave
				}
			}
		}
	}
}

//获取一句话的级别,SortFlag ==0时句子优先级比坎等高，SortFlag ==1时句子优先级比坎等低 ,isweave==1 表示组合牌区的一句话，不参入排序
func (self *GameLogic_zp_jzhp) GetWeaveLevel(SortFlag int, wWeaveKind int, isweave int) int {
	level := 0
	if isweave == 1 {
		level = 100
	} else {
		if wWeaveKind == ZP_WIK_TA {
			level = 35
		} else if wWeaveKind == ZP_WIK_HUA {
			level = 40
		} else if wWeaveKind == ZP_WIK_TIANLONG {
			level = 45
		} else if wWeaveKind == ZP_WIK_GANG {
			level = 50
		} else if wWeaveKind == ZP_WIK_KAN {
			level = 55
		} else if wWeaveKind == ZP_WIK_PENG {
			level = 60
		} else if wWeaveKind == ZP_WIK_LEFT || wWeaveKind == ZP_WIK_CENTER || wWeaveKind == ZP_WIK_RIGHT {
			if SortFlag == 1 {
				level = 30
			} else {
				level = 65
			}
		} else if wWeaveKind == ZP_WIK_HALF {
			if SortFlag == 1 {
				level = 20
			} else {
				level = 70
			}
		}
	}
	return level
}

//自摸的牌可以是坎、绍，接炮的牌当作碰招。 zimoFlag bool, 这个函数要放在恢复花牌之后使用
func (self *GameLogic_zp_jzhp) CheckHu_ZhaoShaoKanPeng(zimoFlag bool, cbCenterCard byte, currentcards byte, ChiHuResult *TagChiHuResult) {
	if zimoFlag || !self.IsValidCard(cbCenterCard) {
		return
	}
	//先检查一下是否有cbCenterCard相关的非坎牌
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	for byIndex := 0; byIndex < byCount; byIndex++ {
		cardsNum := currentcards
		if 1 == self.IsJingPai(cbCenterCard) {
			cardsNum = self.GetCardsNumInResult(cbCenterCard, ChiHuResult.ChiHuItemInfoArray[byIndex])
		}
		for j := 0; j < ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveCount && j < 10; j++ {
			if (ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j]&(ZP_WIK_TIANLONG|ZP_WIK_KAN)) != ZP_WIK_NULL && self.SwitchToCardIndexNoHua(cbCenterCard) == self.SwitchToCardIndexNoHua(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) && self.IsValidCard((ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0])) {
				if ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[j] == 1 {
					//外面已经处理了
				} else if ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[j] == 0 {
					//修改为碰或招
					if self.IsValidCard(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][3]) {
						if cardsNum <= 4 {
							ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] = ZP_WIK_GANG
						}
					} else {
						if cardsNum <= 3 {
							ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] = ZP_WIK_PENG
						}
					}
				}
			}
		}
	}
}

//自摸的牌可以是坎，接炮的牌如果只能组成3张一样的，只能当作碰。 cbCenterCard == 0表示自摸， 这个函数要放在恢复花牌之后使用
func (self *GameLogic_zp_jzhp) CheckHu_KanPeng(cbCenterCard byte, ChiHuResult *TagChiHuResult) {
	if !self.IsValidCard(cbCenterCard) {
		return
	}
	//先检查一下是否有cbCenterCard相关的非坎牌
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	for byIndex := 0; byIndex < byCount; byIndex++ {
		//记录第一个带cbCenterCard牌的坎
		kanpos := 100
		totalcardCnt := byte(0)
		cardCnt := byte(0)
		for j := 0; j < ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveCount && j < 10; j++ {
			if ChiHuResult.ChiHuItemInfoArray[byIndex].IsWeave[j] == 1 {
				continue
			}
			hasFlag := false
			for k := 0; k < 5; k++ {
				if self.IsValidCard(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][k]) && self.SwitchToCardIndexNoHua(cbCenterCard) == self.SwitchToCardIndexNoHua(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][k]) {
					hasFlag = true
					totalcardCnt++
					if ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] != ZP_WIK_KAN {
						cardCnt++
					}
				}
			}
			if hasFlag {
				if ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] == ZP_WIK_KAN && kanpos > 10 {
					kanpos = j
				}
			}
		}
		if kanpos < 10 {
			if cardCnt == 0 {
				//修改为碰
				ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[kanpos] = ZP_WIK_PENG
			}
		}
	}
}

//获取赖子可以替换的牌（三五七）在胡牌结果中的数目
func (self *GameLogic_zp_jzhp) GetCardsNumInResult(cbCenterCard byte, ChiHuItemInfo TagChiHuItemInfo) byte {
	if !self.IsValidCard(cbCenterCard) {
		return 0
	}
	//记录带cbCenterCard牌的
	cardCnt := byte(0)
	for j := 0; j < ChiHuItemInfo.WeaveCount && j < 10; j++ {
		if ChiHuItemInfo.IsWeave[j] == 1 {
			continue
		}
		for k := 0; k < 5; k++ {
			if self.IsValidCard(ChiHuItemInfo.WeaveItemInfo[j][k]) && self.SwitchToCardIndexNoHua(cbCenterCard) == self.SwitchToCardIndexNoHua(ChiHuItemInfo.WeaveItemInfo[j][k]) {
				cardCnt++
			}
		}
	}
	return cardCnt
}
func (self *GameLogic_zp_jzhp) Check10Dui(cbCardIndex [TCGZ_MAX_INDEX]byte, WeaveItem []TagWeaveItem, cbWeaveCount byte, iGuanShengCardsCnt byte) (byte, TagChiHuResult) {
	var ChiHuResult TagChiHuResult
	if cbWeaveCount > 0 || iGuanShengCardsCnt > 0 {
		return ZP_WIK_NULL, ChiHuResult
	}
	weaveCnt := 0
	is10Dui := 1
	var tChiHuItemInfo TagChiHuItemInfo
	tChiHuItemInfo.CardEye = enShiDui
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] == 2 {
			tChiHuItemInfo.WeaveKind[weaveCnt] = ZP_WIK_HALF
			tChiHuItemInfo.WeaveItemInfo[weaveCnt][0] = self.SwitchToCardData(i)
			tChiHuItemInfo.WeaveItemInfo[weaveCnt][1] = self.SwitchToCardData(i)
			weaveCnt++
		} else if cbCardIndex[i] == 4 {
			tChiHuItemInfo.WeaveKind[weaveCnt] = ZP_WIK_HALF
			tChiHuItemInfo.WeaveItemInfo[weaveCnt][0] = self.SwitchToCardData(i)
			tChiHuItemInfo.WeaveItemInfo[weaveCnt][1] = self.SwitchToCardData(i)
			weaveCnt++
			tChiHuItemInfo.WeaveKind[weaveCnt] = ZP_WIK_HALF
			tChiHuItemInfo.WeaveItemInfo[weaveCnt][0] = self.SwitchToCardData(i)
			tChiHuItemInfo.WeaveItemInfo[weaveCnt][1] = self.SwitchToCardData(i)
			weaveCnt++
		} else if cbCardIndex[i] == 0 {
			continue
		} else {
			is10Dui = 0
			break
		}
	}
	if is10Dui == 0 {
		return ZP_WIK_NULL, ChiHuResult
	}
	tChiHuItemInfo.WeaveCount = weaveCnt
	ChiHuResult.ChiHuKind = ZP_WIK_CHI_HU
	ChiHuResult.ChiHuItemInfoArray = append(ChiHuResult.ChiHuItemInfoArray, tChiHuItemInfo)
	return ZP_WIK_CHI_HU, ChiHuResult
}

//获取同种牌，如果当前是花牌就获取它对应的非花牌，如果是非花牌就获取它对应的花牌，如果没有花牌返回0
func (self *GameLogic_zp_jzhp) GetParterCard(cbCard byte) byte {
	retcard := byte(0)
	if !self.IsValidCard(cbCard) {
		return retcard
	}
	if cbCard <= 0x16 {
		if cbCard == 0x05 {
			retcard = 0x17
		} else if cbCard == 0x0B {
			retcard = 0x18
		} else if cbCard == 0x15 {
			retcard = 0x19
		} else if cbCard == 0x0D {
			retcard = 0x1A
		} else if cbCard == 0x11 {
			retcard = 0x1B
		}
	} else {
		if cbCard == 0x17 {
			retcard = 0x05
		} else if cbCard == 0x18 {
			retcard = 0x0B
		} else if cbCard == 0x19 {
			retcard = 0x15
		} else if cbCard == 0x1A {
			retcard = 0x0D
		} else if cbCard == 0x1B {
			retcard = 0x11
		}
	}
	return retcard
}

//只计算一个玩家组合牌的胡息
func (self *GameLogic_zp_jzhp) CalculateWeaveHuXi(WeaveItemArray [10]TagWeaveItem, cbWeaveItemCount byte) int {
	wHeiPeng := 0
	wHuXi := 0
	for j := byte(0); j < cbWeaveItemCount && j < 10; j++ {
		byCnt := 0
		wTempHuXi := 0
		if ZP_WIK_PENG == WeaveItemArray[j].WeaveKind {
			if 1 == self.IsHongPai(WeaveItemArray[j].PublicCard) {
				wTempHuXi = 1
			} else {
				wHeiPeng++
			}
		} else if ZP_WIK_KAN == WeaveItemArray[j].WeaveKind {
			byCnt = 3
			wTempHuXi = 1
			if 1 == self.IsHongPai(WeaveItemArray[j].PublicCard) {
				wTempHuXi = 2
			}
		} else if ZP_WIK_GANG == WeaveItemArray[j].WeaveKind {
			byCnt = 4
			wTempHuXi = 2
			if 1 == self.IsHongPai(WeaveItemArray[j].PublicCard) {
				wTempHuXi = 4
			}
		} else if ZP_WIK_TIANLONG == WeaveItemArray[j].WeaveKind {
			wTempHuXi = 2
			if 1 == self.IsHongPai(WeaveItemArray[j].PublicCard) {
				wTempHuXi = 4
			}
		} else if ZP_WIK_HUA == WeaveItemArray[j].WeaveKind {
			byCnt = 4
			wTempHuXi = 4
			if 1 == self.IsHongPai(WeaveItemArray[j].PublicCard) {
				wTempHuXi = 8
			}
		} else if ZP_WIK_TA == WeaveItemArray[j].WeaveKind {
			byCnt = 5
			wTempHuXi = 4
			if 1 == self.IsHongPai(WeaveItemArray[j].PublicCard) {
				wTempHuXi = 8
			}
		} else if ZP_WIK_LEFT == WeaveItemArray[j].WeaveKind || ZP_WIK_CENTER == WeaveItemArray[j].WeaveKind || ZP_WIK_RIGHT == WeaveItemArray[j].WeaveKind {
			//上大人、可知礼的吃牌类型
			cbCurrentIndex := self.SwitchToCardIndex(WeaveItemArray[j].Cards[0])
			if cbCurrentIndex == 0 || cbCurrentIndex == 6 {
				wTempHuXi = 1
			}
		}
		byJingCnt := 0
		byHuaCnt := 0
		for _, card := range WeaveItemArray[j].Cards {
			wTempHuXi += self.GetSinglePaiHuxi(card)
			if 1 == self.IsJingPai(card) {
				byJingCnt++
			}
			if 1 == self.IsHuaPai(card) {
				byHuaCnt++
			}
		}
		if byCnt > 2 && byJingCnt > 0 {
			temHuXi2 := self.CalculateJingHuXi(byCnt, byHuaCnt, 2, false)
			if wTempHuXi < temHuXi2 {
				wTempHuXi = temHuXi2
			}
		}

		wHuXi += wTempHuXi
	}
	//3个黑碰算一个子
	//wHuXi += (wHeiPeng/3)
	return wHuXi
}

//计算胡息
func (self *GameLogic_zp_jzhp) CalculateHuXi(ChiHuResult *TagChiHuResult) (iMaxIndex int, iMaxHuXi int, realHuXiInfo [10]int) {
	iMaxIndex = 0
	iMaxHuXi = 0
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	for byIndex := 0; byIndex < byCount; byIndex++ {
		iMainJingIndex := self.CalculateMainJing(ChiHuResult.ChiHuItemInfoArray[byIndex])
		wHeiPeng := 0
		wHuXi := 0
		tmpHuXiInfo := [10]int{}
		for j := 0; j < ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveCount && j < 10; j++ {
			byCnt := 0
			if ZP_WIK_PENG == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 0
				if 1 == self.IsHongPai(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 1
				} else {
					wHeiPeng++
				}
			} else if ZP_WIK_KAN == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				byCnt = 3
				ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 1
				if 1 == self.IsHongPai(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 2
				}
			} else if ZP_WIK_GANG == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				byCnt = 4
				ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 2
				if 1 == self.IsHongPai(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 4
				}
			} else if ZP_WIK_TIANLONG == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				byCnt = 4
				ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 2
				if 1 == self.IsHongPai(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 4
				}
			} else if ZP_WIK_HUA == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				byCnt = 5
				ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 4
				if 1 == self.IsHongPai(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 8
				}
			} else if ZP_WIK_TA == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				byCnt = 5
				ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 4
				if 1 == self.IsHongPai(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0]) {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 8
				}
			} else if ZP_WIK_LEFT == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] || ZP_WIK_CENTER == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] || ZP_WIK_RIGHT == ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveKind[j] {
				//上大人、可知礼的吃牌类型
				cbCurrentIndex := self.SwitchToCardIndex(ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j][0])
				if cbCurrentIndex == 0 || cbCurrentIndex == 6 {
					ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j] = 1
				}
			}
			tmpHuXi := ChiHuResult.ChiHuItemInfoArray[byIndex].HuXiInfo[j]
			byJingCnt := 0
			byHuaCnt := 0
			for _, card := range ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j] {
				tmpHuXi += self.GetSinglePaiHuxi(card)
				if 1 == self.IsJingPai(card) {
					byJingCnt++
				}
				if 1 == self.IsHuaPai(card) {
					byHuaCnt++
				}
			}
			//主精？
			IsHasMain := false
			for _, card := range ChiHuResult.ChiHuItemInfoArray[byIndex].WeaveItemInfo[j] {
				if self.IsValidCard(card) && iMainJingIndex == self.SwitchToCardIndexNoHua(card) {
					IsHasMain = true
					// 不是整体加倍
					tmpHuXi += self.GetSinglePaiHuxi(card)
				}
			}
			if IsHasMain {
				//tmpHuXi *=2
				// 不是整体加倍
			}
			if byCnt > 2 && byJingCnt > 0 {
				temHuXi2 := self.CalculateJingHuXi(byCnt, byHuaCnt, 2, IsHasMain)
				if tmpHuXi < temHuXi2 {
					tmpHuXi = temHuXi2
				}
			}
			wHuXi += tmpHuXi
			tmpHuXiInfo[j] = tmpHuXi
		}
		//3个黑碰算一个子
		//wHuXi += (wHeiPeng/3)
		ChiHuResult.ChiHuItemInfoArray[byIndex].TotalHuxi = wHuXi
		ChiHuResult.ChiHuItemInfoArray[byIndex].MainJingIndex = iMainJingIndex
		if iMaxHuXi < wHuXi {
			iMaxHuXi = wHuXi
			iMaxIndex = byIndex
			copy(realHuXiInfo[:], tmpHuXiInfo[:])
		}
	}
	return iMaxIndex, iMaxHuXi, realHuXiInfo
}

//计算胡息 20200914 dqjs-4587 修改规则，主精的计算优先能胡牌
func (self *GameLogic_zp_jzhp) CalculateHuXi2(ChiHuResult *TagChiHuResult, iMaxIndex *int, iMaxHuXi *int, realHuXiInfo *[10]int) {
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	adjustIndex := byCount
	for byIndex := 0; byIndex < byCount; byIndex++ {
		iMainJingIndex := self.CalculateMainJing2(ChiHuResult.ChiHuItemInfoArray[byIndex])
		if (len(iMainJingIndex) == 1 && iMainJingIndex[0] == ChiHuResult.ChiHuItemInfoArray[byIndex].MainJingIndex) || len(iMainJingIndex) == 0 {
			continue
		}
		//轮流当主精,胡数优先
		for mainJinI := 0; mainJinI < len(iMainJingIndex); mainJinI++ {
			mainJinIndex := iMainJingIndex[mainJinI]
			if mainJinIndex == ChiHuResult.ChiHuItemInfoArray[byIndex].MainJingIndex {
				continue
			}
			wHeiPeng := 0
			wHuXi := 0
			tmpHuXiInfo := [10]int{}
			var tempTagChiHuItemInfo TagChiHuItemInfo
			tempTagChiHuItemInfo = ChiHuResult.ChiHuItemInfoArray[byIndex]
			for j := 0; j < tempTagChiHuItemInfo.WeaveCount && j < 10; j++ {
				byCnt := 0
				if ZP_WIK_PENG == tempTagChiHuItemInfo.WeaveKind[j] {
					tempTagChiHuItemInfo.HuXiInfo[j] = 0
					if 1 == self.IsHongPai(tempTagChiHuItemInfo.WeaveItemInfo[j][0]) {
						tempTagChiHuItemInfo.HuXiInfo[j] = 1
					} else {
						wHeiPeng++
					}
				} else if ZP_WIK_KAN == tempTagChiHuItemInfo.WeaveKind[j] {
					byCnt = 3
					tempTagChiHuItemInfo.HuXiInfo[j] = 1
					if 1 == self.IsHongPai(tempTagChiHuItemInfo.WeaveItemInfo[j][0]) {
						tempTagChiHuItemInfo.HuXiInfo[j] = 2
					}
				} else if ZP_WIK_GANG == tempTagChiHuItemInfo.WeaveKind[j] {
					byCnt = 4
					tempTagChiHuItemInfo.HuXiInfo[j] = 2
					if 1 == self.IsHongPai(tempTagChiHuItemInfo.WeaveItemInfo[j][0]) {
						tempTagChiHuItemInfo.HuXiInfo[j] = 4
					}
				} else if ZP_WIK_TIANLONG == tempTagChiHuItemInfo.WeaveKind[j] {
					byCnt = 4
					tempTagChiHuItemInfo.HuXiInfo[j] = 2
					if 1 == self.IsHongPai(tempTagChiHuItemInfo.WeaveItemInfo[j][0]) {
						tempTagChiHuItemInfo.HuXiInfo[j] = 4
					}
				} else if ZP_WIK_HUA == tempTagChiHuItemInfo.WeaveKind[j] {
					byCnt = 5
					tempTagChiHuItemInfo.HuXiInfo[j] = 4
					if 1 == self.IsHongPai(tempTagChiHuItemInfo.WeaveItemInfo[j][0]) {
						tempTagChiHuItemInfo.HuXiInfo[j] = 8
					}
				} else if ZP_WIK_TA == tempTagChiHuItemInfo.WeaveKind[j] {
					byCnt = 5
					tempTagChiHuItemInfo.HuXiInfo[j] = 4
					if 1 == self.IsHongPai(tempTagChiHuItemInfo.WeaveItemInfo[j][0]) {
						tempTagChiHuItemInfo.HuXiInfo[j] = 8
					}
				} else if ZP_WIK_LEFT == tempTagChiHuItemInfo.WeaveKind[j] || ZP_WIK_CENTER == tempTagChiHuItemInfo.WeaveKind[j] || ZP_WIK_RIGHT == tempTagChiHuItemInfo.WeaveKind[j] {
					//上大人、可知礼的吃牌类型
					cbCurrentIndex := self.SwitchToCardIndex(tempTagChiHuItemInfo.WeaveItemInfo[j][0])
					if cbCurrentIndex == 0 || cbCurrentIndex == 6 {
						tempTagChiHuItemInfo.HuXiInfo[j] = 1
					}
				}
				tmpHuXi := tempTagChiHuItemInfo.HuXiInfo[j]
				byJingCnt := 0
				byHuaCnt := 0
				for _, card := range tempTagChiHuItemInfo.WeaveItemInfo[j] {
					tmpHuXi += self.GetSinglePaiHuxi(card)
					if 1 == self.IsJingPai(card) {
						byJingCnt++
					}
					if 1 == self.IsHuaPai(card) {
						byHuaCnt++
					}
				}
				//主精？
				IsHasMain := false
				for _, card := range tempTagChiHuItemInfo.WeaveItemInfo[j] {
					if self.IsValidCard(card) && mainJinIndex == self.SwitchToCardIndexNoHua(card) {
						IsHasMain = true
						// 不是整体加倍
						tmpHuXi += self.GetSinglePaiHuxi(card)
					}
				}
				if IsHasMain {
					//tmpHuXi *=2
					// 不是整体加倍
				}
				if byCnt > 2 && byJingCnt > 0 {
					temHuXi2 := self.CalculateJingHuXi(byCnt, byHuaCnt, 2, IsHasMain)
					if tmpHuXi < temHuXi2 {
						tmpHuXi = temHuXi2
					}
				}
				wHuXi += tmpHuXi
				tmpHuXiInfo[j] = tmpHuXi
			}
			//3个黑碰算一个子
			//wHuXi += (wHeiPeng/3)
			tempTagChiHuItemInfo.TotalHuxi = wHuXi
			tempTagChiHuItemInfo.MainJingIndex = mainJinIndex

			if *iMaxHuXi < wHuXi {
				*iMaxHuXi = wHuXi
				*iMaxIndex = adjustIndex
				copy(realHuXiInfo[:], tmpHuXiInfo[:])

				ChiHuResult.ChiHuItemInfoArray = append(ChiHuResult.ChiHuItemInfoArray, tempTagChiHuItemInfo)
				adjustIndex++
			}
		}
	}
	return
}

//坎、招、统、滑、踏 使用这个函数计算胡数，碰、句不用这个函数
func (self *GameLogic_zp_jzhp) CalculateJingHuXi(cardNum int, huaNum int, hunNum int, mainFlag bool) int {
	huxi := 0
	if cardNum < 3 {
		return huxi
	}
	if cardNum == 3 {
		huxi = self.Calculate3SameJingHuXi(cardNum, huaNum, hunNum, mainFlag)
	} else if cardNum == 4 {
		if huaNum <= 2 {
			huxi = 2 * self.Calculate3SameJingHuXi(3, huaNum, hunNum, mainFlag)
		} else {
			huxi = self.CalculateSpecialJingHuXi(cardNum, huaNum, hunNum, mainFlag)
		}
	} else if cardNum == 5 {
		if huaNum <= 2 {
			huxi = 4 * self.Calculate3SameJingHuXi(3, huaNum, hunNum, mainFlag)
		} else {
			huxi = self.CalculateSpecialJingHuXi(cardNum, huaNum, hunNum, mainFlag)
		}
	}
	return huxi
}
func (self *GameLogic_zp_jzhp) Calculate3SameJingHuXi(cardNum int, huaNum int, hunNum int, mainFlag bool) int {
	huxi := 0
	if cardNum != 3 {
		return huxi
	}
	if cardNum == 3 && huaNum == 0 {
		huxi = 5
	} else if cardNum == 3 && huaNum == 1 {
		huxi = 6
	} else if cardNum == 3 && huaNum == 2 {
		huxi = 7
	} else if cardNum == 3 && huaNum == 3 && hunNum == 1 {
		huxi = 9
	} else if cardNum == 3 && huaNum == 3 && hunNum == 2 {
		huxi = 12
	} else {

	}
	if mainFlag {
		huxi *= 2
	}
	return huxi
}
func (self *GameLogic_zp_jzhp) CalculateSpecialJingHuXi(cardNum int, huaNum int, hunNum int, mainFlag bool) int {
	huxi := 0
	if huaNum <= 2 || cardNum <= 3 {
		return huxi
	}
	if cardNum == 4 && huaNum == 3 && hunNum == 1 {
		huxi = 18
	} else if cardNum == 4 && huaNum == 3 && hunNum == 2 {
		huxi = 18
	} else if cardNum == 4 && huaNum == 4 && hunNum == 2 {
		huxi = 24
	} else if cardNum == 5 && huaNum == 3 && hunNum == 1 {
		huxi = 36
	} else if cardNum == 5 && huaNum == 3 && hunNum == 2 {
		huxi = 36
	} else if cardNum == 5 && huaNum == 4 && hunNum == 2 {
		huxi = 48
	}
	if mainFlag {
		huxi *= 2
	}
	return huxi
}

//统计主精时需要使用的组合级别，相同数目的精牌要比较组合级别。半句和1句级别为1，碰级别为3，坎级别为5，招统为7，滑和踏为9
func (self *GameLogic_zp_jzhp) GetWeaveLevel2(wKind int) (iLevel int) {
	iLevel = 1
	if (wKind & (ZP_WIK_LEFT | ZP_WIK_CENTER | ZP_WIK_RIGHT)) != 0 {
		iLevel = 1
	} else if wKind == ZP_WIK_HALF {
		iLevel = 1
	} else if wKind == ZP_WIK_PENG {
		iLevel = 3
	} else if wKind == ZP_WIK_KAN {
		iLevel = 5
	} else if wKind == ZP_WIK_GANG {
		iLevel = 7
	} else if wKind == ZP_WIK_TIANLONG {
		iLevel = 7
	} else if wKind == ZP_WIK_HUA {
		iLevel = 9
	} else if wKind == ZP_WIK_TA {
		iLevel = 9
	}
	return iLevel
}

//统计主精
func (self *GameLogic_zp_jzhp) CalculateMainJing(ChiHuItemInfoArray TagChiHuItemInfo) (byMainJingIndex byte) {
	i3Cnt := 0
	i5Cnt := 0
	i7Cnt := 0
	i3HuaCnt := 0
	i5HuaCnt := 0
	i7HuaCnt := 0
	i3WLevel := 0 //20200723 在精牌数目相同，花牌数目也相同时，要考虑胡数最大的优先
	i5WLevel := 0 //20200723 在精牌数目相同，花牌数目也相同时，要考虑胡数最大的优先
	i7WLevel := 0 //20200723 在精牌数目相同，花牌数目也相同时，要考虑胡数最大的优先
	for j := 0; j < ChiHuItemInfoArray.WeaveCount && j < 10; j++ {
		if ZP_WIK_NULL == ChiHuItemInfoArray.WeaveKind[j] {
			continue
		} else {
			for k := 0; k < 5; k++ {
				card := ChiHuItemInfoArray.WeaveItemInfo[j][k]
				if self.IsValidCard(card) {
					if 10 == self.SwitchToCardIndexNoHua(card) {
						i3Cnt++
						if 23 == self.SwitchToCardIndex(card) {
							i3HuaCnt++
						}
						Level := self.GetWeaveLevel2(ChiHuItemInfoArray.WeaveKind[j])
						if i3WLevel < Level {
							i3WLevel = Level
						}
					} else if 20 == self.SwitchToCardIndexNoHua(card) {
						i5Cnt++
						if 24 == self.SwitchToCardIndex(card) {
							i5HuaCnt++
						}
						Level := self.GetWeaveLevel2(ChiHuItemInfoArray.WeaveKind[j])
						if i5WLevel < Level {
							i5WLevel = Level
						}
					} else if 12 == self.SwitchToCardIndexNoHua(card) {
						i7Cnt++
						if 25 == self.SwitchToCardIndex(card) {
							i7HuaCnt++
						}
						Level := self.GetWeaveLevel2(ChiHuItemInfoArray.WeaveKind[j])
						if i7WLevel < Level {
							i7WLevel = Level
						}
					}
				}
			}
		}
	}
	if i3Cnt == 0 && i5Cnt == 0 && i7Cnt == 0 {
		byMainJingIndex = 254 //给个比较大的值，
		return
	}
	iMainJing := i3Cnt
	iMainHua := i3HuaCnt
	iMainLevel := i3WLevel
	byMainJingIndex = 10
	if iMainJing < i5Cnt || (iMainJing == i5Cnt && iMainHua < i5HuaCnt) || (iMainJing == i5Cnt && iMainHua == i5HuaCnt && iMainLevel < i5WLevel) {
		iMainJing = i5Cnt
		iMainHua = i5HuaCnt
		iMainLevel = i5WLevel
		byMainJingIndex = 20
	}
	if iMainJing < i7Cnt || (iMainJing == i7Cnt && iMainHua < i7HuaCnt) || (iMainJing == i7Cnt && iMainHua == i7HuaCnt && iMainLevel < i7WLevel) {
		iMainJing = i7Cnt
		iMainHua = i7HuaCnt
		iMainLevel = i7WLevel
		byMainJingIndex = 12
	}
	return
}

//统计主精 20200914 dqjs-4587 修改规则，主精的计算优先能胡牌(干脆什么都不比了，直接轮流当主精，然后看谁的胡数大？)
func (self *GameLogic_zp_jzhp) CalculateMainJing2(ChiHuItemInfoArray TagChiHuItemInfo) (byMainJingIndex []byte) {
	i3Cnt := 0
	i5Cnt := 0
	i7Cnt := 0
	for j := 0; j < ChiHuItemInfoArray.WeaveCount && j < 10; j++ {
		if ZP_WIK_NULL == ChiHuItemInfoArray.WeaveKind[j] {
			continue
		} else {
			for k := 0; k < 5; k++ {
				card := ChiHuItemInfoArray.WeaveItemInfo[j][k]
				if self.IsValidCard(card) {
					if 10 == self.SwitchToCardIndexNoHua(card) {
						i3Cnt++
					} else if 20 == self.SwitchToCardIndexNoHua(card) {
						i5Cnt++
					} else if 12 == self.SwitchToCardIndexNoHua(card) {
						i7Cnt++
					}
				}
			}
		}
	}
	if i3Cnt == 0 && i5Cnt == 0 && i7Cnt == 0 {
		byMainJingIndex = []byte{}
		return
	}
	//轮流当主精吧
	if i3Cnt > 0 {
		byMainJingIndex = append(byMainJingIndex, 10)
	}
	if i5Cnt > 0 {
		byMainJingIndex = append(byMainJingIndex, 20)
	}
	if i7Cnt > 0 {
		byMainJingIndex = append(byMainJingIndex, 12)
	}

	return
}

func (sp *Sport_zp_jzhp) GetMaxHuXi(ChiHuResult TagChiHuResult) (iMaxIndex int, iMaxHuXi int) {
	iMaxIndex = 0
	iMaxHuXi = 0
	byCount := len(ChiHuResult.ChiHuItemInfoArray)
	for byIndex := 0; byIndex < byCount; byIndex++ {
		if iMaxHuXi < ChiHuResult.ChiHuItemInfoArray[byIndex].TotalHuxi {
			iMaxHuXi = ChiHuResult.ChiHuItemInfoArray[byIndex].TotalHuxi
			iMaxIndex = byIndex
		}
	}
	return iMaxIndex, iMaxHuXi
}

//分析扑克，cbCardIndex：最终的点数数组，包含了要分析的牌
//Add by luoqifu 20180329 优化胡牌算法
//原理：
//1.首先找出所有包含一对的情形，移除对子（作为将牌，注意258将的条件），记下剩余牌的所有集合为Tn;
//2.针对每个Tn中的数组尝试移除一个顺子，成功转到2，失败到3。
//3.针对每个Tn中的数组尝试移除一个刻子（DDD），成功转到2。
//4.若当前的数组的数量变为0，则表示，当前的方案可以胡牌。
//大冶字牌由于有时需要满足3n+2，有时满足3n，所以添加wChiHuRight区分
func (self *GameLogic_zp_jzhp) AnalyseCard_new(cbCardIndex [TCGZ_MAX_INDEX]byte, WeaveItem []TagWeaveItem, cbWeaveCount byte, wChiHuRight int) (bool, []TagAnalyseItem) {
	AnalyseItemArray := []TagAnalyseItem{}
	//统计索引数组中的所有牌数
	cbCardCount := byte(0)
	for k := 0; k < TCGZ_MAX_INDEX; k++ {
		cbCardCount += cbCardIndex[k]
	}

	//效验数目，索引数组中牌的总数-2（一对将）后剩下的数是3的倍数
	//if ((cbCardCount<2)||(cbCardCount>TCGZ_MAXHANDCARD)) {
	if cbCardCount < 2 {
		return false, AnalyseItemArray
	}

	//判断加入了待分析的牌后是否构成半句
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ { //设置变量
		if cbCardIndex[i] > 0 {
			cardHalf := self.SearchHalf(cbCardIndex, i)
			if len(cardHalf) > 0 {
				//删除一个半句，如乙的话，乙己，乙二，乙三，避免重复只往后找
				for _, is := range cardHalf {
					self.m_AnalyseItemArray = []TagAnalyseItem{}
					abTempCardIndex := [TCGZ_MAX_INDEX]byte{}
					copy(abTempCardIndex[:], cbCardIndex[:])
					if abTempCardIndex[i] > 0 {
						abTempCardIndex[i]--
					} else {
						continue
					}
					if abTempCardIndex[is] > 0 {
						abTempCardIndex[is]--
					} else {
						continue
					}
					var KindItemArray []TagKindItem
					self.Check_3N(abTempCardIndex, &KindItemArray)

					if len(self.m_AnalyseItemArray) > 0 {
						for j := 0; j < len(self.m_AnalyseItemArray); j++ {
							//变量定义
							var AnalyseItem TagAnalyseItem

							//分析每一组组合牌，得到组合牌的组合牌型和中间牌，比如说WK_PENG,WK_CHI保存到分析子项中
							for w := byte(0); w < cbWeaveCount; w++ {
								AnalyseItem.WeaveKind[w] = WeaveItem[w].WeaveKind
								AnalyseItem.Cards[w][0] = WeaveItem[w].Cards[0]
								AnalyseItem.Cards[w][1] = WeaveItem[w].Cards[1]
								AnalyseItem.Cards[w][2] = WeaveItem[w].Cards[2]
								AnalyseItem.Cards[w][3] = WeaveItem[w].Cards[3]
								AnalyseItem.Cards[w][4] = WeaveItem[w].Cards[4]
								AnalyseItem.IsWeave[w] = 1 //组合牌区的牌
							}
							nKindItemCount := byte(10) //wWeaveKind数组长度为10 //sizeof(m_AnalyseItemArray[j].wWeaveKind)/sizeof(WORD);
							for k := byte(0); cbWeaveCount+k < nKindItemCount; k++ {
								AnalyseItem.WeaveKind[cbWeaveCount+k] = self.m_AnalyseItemArray[j].WeaveKind[k]
								AnalyseItem.Cards[cbWeaveCount+k][0] = self.m_AnalyseItemArray[j].Cards[k][0]
								AnalyseItem.Cards[cbWeaveCount+k][1] = self.m_AnalyseItemArray[j].Cards[k][1]
								AnalyseItem.Cards[cbWeaveCount+k][2] = self.m_AnalyseItemArray[j].Cards[k][2]
								AnalyseItem.Cards[cbWeaveCount+k][3] = self.m_AnalyseItemArray[j].Cards[k][3]
								AnalyseItem.Cards[cbWeaveCount+k][4] = self.m_AnalyseItemArray[j].Cards[k][4]
								AnalyseItem.IsWeave[cbWeaveCount+k] = 0 //不是组合牌区的牌
							}
							//将半句加入到分析结果中去
							nKindItemCount = 0
							for n := 0; n < 10; n++ {
								if AnalyseItem.WeaveKind[n] > 0 {
									if nKindItemCount < 9 {
										nKindItemCount++
									}
								}
							}
							AnalyseItem.WeaveKind[nKindItemCount] = ZP_WIK_HALF
							AnalyseItem.Cards[nKindItemCount][0] = self.SwitchToCardData(i)
							AnalyseItem.Cards[nKindItemCount][1] = self.SwitchToCardData(is)
							AnalyseItem.IsWeave[nKindItemCount] = 0 //不是组合牌区的牌
							//将分析结果插入到分析数组中
							AnalyseItemArray = append(AnalyseItemArray, AnalyseItem)
						}
					}
				}
			}
		}
	}
	return (len(AnalyseItemArray) > 0), AnalyseItemArray
}

//查找半句牌的索引，避免重复，只往后找
func (self *GameLogic_zp_jzhp) SearchHalf(cbCardIndex [TCGZ_MAX_INDEX]byte, cardIndex byte) []byte {
	cardHalf := []byte{}
	if cardIndex < TCGZ_MAX_INDEX {
		for _, ch := range zp_tcgz_CardsHalf[cardIndex] {
			//赖子函数也需要用到这个函数，把cbCardIndex[ch] > 0的限制去掉
			//if ch < TCGZ_MAX_INDEX && cbCardIndex[ch] > 0{
			if ch < TCGZ_MAX_INDEX {
				cardHalf = append(cardHalf, ch)
			}
		}
	}
	return cardHalf
}

//分析扑克,去掉牌眼后的牌符合牌数量为3*N的要求，cbCardIndex：最终的点数数组，包含了要分析的牌
func (self *GameLogic_zp_jzhp) Check_3N(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) bool {
	//统计索引数组中的所有牌数
	cbCardCount := byte(0)
	for i := 0; i < TCGZ_MAX_INDEX; i++ {
		cbCardCount += cbCardIndex[i]
	}
	if cbCardCount == 0 {
		//只能在这里保存这次找到的胡牌结果了
		//变量定义
		var AnalyseItem TagAnalyseItem
		nKindItemCount := len(*KindItemArray)
		for j := 0; j < nKindItemCount && j < 10; j++ {
			//wWeaveKind数组长度为10
			AnalyseItem.WeaveKind[j] = (*KindItemArray)[j].WeaveKind
			AnalyseItem.Cards[j][0] = (*KindItemArray)[j].CardIndex[0]
			AnalyseItem.Cards[j][1] = (*KindItemArray)[j].CardIndex[1]
			AnalyseItem.Cards[j][2] = (*KindItemArray)[j].CardIndex[2]
			AnalyseItem.Cards[j][3] = (*KindItemArray)[j].CardIndex[3]
			AnalyseItem.Cards[j][4] = (*KindItemArray)[j].CardIndex[4]
		}
		self.m_AnalyseItemArray = append(self.m_AnalyseItemArray, AnalyseItem)
		return false //需要查找所有的可能性
	}

	//避免修改了cbCardIndex，需要定义两个临时数组
	abTempCardIndex1 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex2 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex3 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex4 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex5 := [TCGZ_MAX_INDEX]byte{}
	copy(abTempCardIndex1[:], cbCardIndex[:])
	bRet := false
	//先删除一个刻子，由于需要碰碰胡优先，所以需要先判断刻子
	//大冶字牌的坎(刻子)不参入分析，实际上不用执行Remove3Same的
	if bRet, abTempCardIndex1 = self.Remove3Same(abTempCardIndex1, KindItemArray); bRet == true {
		if self.Check_3N(abTempCardIndex1, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}
	copy(abTempCardIndex2[:], cbCardIndex[:])
	//在删除一个顺子
	if bRet, abTempCardIndex2 = self.RemoveStraight(abTempCardIndex2, KindItemArray); bRet == true {
		if self.Check_3N(abTempCardIndex2, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}
	copy(abTempCardIndex3[:], cbCardIndex[:])
	////在删除一个123
	//if bRet,abTempCardIndex3 = self.Remove1_2_3(abTempCardIndex3,KindItemArray);bRet == true  {
	//	if (self.Check_3N(abTempCardIndex3,KindItemArray)) {
	//		return true;
	//	}
	//	nKindItemCount:=len(*KindItemArray)
	//	if (nKindItemCount > 0) {
	//		*KindItemArray= (*KindItemArray)[:nKindItemCount-1] //删除最后一个
	//	}
	//}
	self.Remove1_2_3new(abTempCardIndex3, KindItemArray)

	copy(abTempCardIndex4[:], cbCardIndex[:])
	//在删除一个观
	if bRet, abTempCardIndex4 = self.Remove4Same(abTempCardIndex4, KindItemArray); bRet == true {
		if self.Check_3N(abTempCardIndex4, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}
	copy(abTempCardIndex5[:], cbCardIndex[:])
	//在删除一个滑
	if bRet, abTempCardIndex5 = self.Remove5Same(abTempCardIndex5, KindItemArray); bRet == true {
		if self.Check_3N(abTempCardIndex5, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}
	return false
}

//删除最前面的一个乙二三等跳跃顺子，适合1、2、3...9、10杂乱无章的顺序。
func (self *GameLogic_zp_jzhp) Remove1_2_3new(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 {
			cards123 := zp_tcgz_Cards123new[i]
			if len(cards123) == 0 {
				break
			}
			for j := 0; j < len(cards123); j++ {
				abTempCardIndex0 := [TCGZ_MAX_INDEX]byte{}
				copy(abTempCardIndex0[:], cbCardIndex[:])
				eachcards123 := cards123[j]
				bRet := false
				//在删除一个123
				if bRet, abTempCardIndex0 = self.EachRemove1_2_3new(abTempCardIndex0, eachcards123, KindItemArray); bRet == true {
					if self.Check_3N(abTempCardIndex0, KindItemArray) {
						break
					}
					nKindItemCount := len(*KindItemArray)
					if nKindItemCount > 0 {
						*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
					}
				}
			}
			break
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个乙二三等跳跃顺子
func (self *GameLogic_zp_jzhp) EachRemove1_2_3new(cbCardIndex [TCGZ_MAX_INDEX]byte, cards123new [3]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	if cards123new[0] >= TCGZ_MAX_INDEX || cards123new[1] >= TCGZ_MAX_INDEX || cards123new[2] >= TCGZ_MAX_INDEX {
		return false, cbCardIndex
	}
	//变量定义
	var KindItem TagKindItem

	//删除1个顺子
	if cbCardIndex[cards123new[0]] >= 1 {
		if (cbCardIndex[cards123new[1]] >= 1) && (cbCardIndex[cards123new[2]] >= 1) {
			KindItem.CardIndex[0] = self.SwitchToCardData(cards123new[0])
			KindItem.CardIndex[1] = self.SwitchToCardData(cards123new[1])
			KindItem.CardIndex[2] = self.SwitchToCardData(cards123new[2])
			KindItem.WeaveKind = ZP_WIK_LEFT
			cbCardIndex[cards123new[0]]--
			cbCardIndex[cards123new[1]]--
			cbCardIndex[cards123new[2]]--
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		} else {
			return false, cbCardIndex
		}
	}

	return false, cbCardIndex
}

//删除最前面的一个乙二三等跳跃顺子，适合1、2、3...9、10基本有序，最好完全有序。
func (self *GameLogic_zp_jzhp) Remove1_2_3(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem

	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 {
			if len(zp_tcgz_Cards123[i]) == 3 && (cbCardIndex[zp_tcgz_Cards123[i][1]] >= 1) && (cbCardIndex[zp_tcgz_Cards123[i][2]] >= 1) {
				KindItem.CardIndex[0] = self.SwitchToCardData(i)
				KindItem.CardIndex[1] = self.SwitchToCardData(zp_tcgz_Cards123[i][1])
				KindItem.CardIndex[2] = self.SwitchToCardData(zp_tcgz_Cards123[i][2])
				KindItem.WeaveKind = ZP_WIK_LEFT
				cbCardIndex[i]--
				cbCardIndex[zp_tcgz_Cards123[i][1]]--
				cbCardIndex[zp_tcgz_Cards123[i][2]]--
				*KindItemArray = append(*KindItemArray, KindItem)
				return true, cbCardIndex
			} else {
				return false, cbCardIndex
			}
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个顺子
func (self *GameLogic_zp_jzhp) RemoveStraight(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem

	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 && zp_tcgz_CardsStraight[i] > 0 {
			if (cbCardIndex[i+1] >= 1) && (cbCardIndex[i+2] >= 1) {
				KindItem.CardIndex[0] = self.SwitchToCardData(i)
				KindItem.CardIndex[1] = self.SwitchToCardData(i + 1)
				KindItem.CardIndex[2] = self.SwitchToCardData(i + 2)
				KindItem.WeaveKind = ZP_WIK_LEFT
				cbCardIndex[i]--
				cbCardIndex[i+1]--
				cbCardIndex[i+2]--
				*KindItemArray = append(*KindItemArray, KindItem)
				return true, cbCardIndex
			} else {
				return false, cbCardIndex
			}
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个刻子
func (self *GameLogic_zp_jzhp) Remove3Same(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem

	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 3 {
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i)
			KindItem.CardIndex[2] = self.SwitchToCardData(i)

			KindItem.WeaveKind = ZP_WIK_KAN
			cbCardIndex[i] -= 3
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		} else if cbCardIndex[i] >= 1 {
			//牌数大于1个小于3个，返回不能组成刻子
			return false, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个观
func (self *GameLogic_zp_jzhp) Remove4Same(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem

	//删除1个观
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 4 {
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i)
			KindItem.CardIndex[2] = self.SwitchToCardData(i)
			KindItem.CardIndex[3] = self.SwitchToCardData(i)
			KindItem.WeaveKind = ZP_WIK_TIANLONG //观
			cbCardIndex[i] -= 4
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		} else if cbCardIndex[i] >= 1 {
			//牌数大于1个小于4个，返回不能组成观
			return false, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个滑
func (self *GameLogic_zp_jzhp) Remove5Same(cbCardIndex [TCGZ_MAX_INDEX]byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem

	//删除1个滑
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 5 {
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i)
			KindItem.CardIndex[2] = self.SwitchToCardData(i)
			KindItem.CardIndex[3] = self.SwitchToCardData(i)
			KindItem.CardIndex[4] = self.SwitchToCardData(i)
			KindItem.WeaveKind = ZP_WIK_HUA
			cbCardIndex[i] -= 5
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		} else if cbCardIndex[i] >= 1 {
			//牌数大于1个小于5个，返回不能组成滑
			return false, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//分析扑克,带赖子,cbCardIndex需要先删除赖子,基本算法与AnalyseCard_new一致，当缺n个牌时就用n个赖子替
func (self *GameLogic_zp_jzhp) AnalyseCard_Magic_new(cbCardIndex [TCGZ_MAX_INDEX]byte, WeaveItem []TagWeaveItem, cbWeaveCount byte, wChiHuRight int, cbMagicNum byte) (bool, []TagAnalyseItem) {
	AnalyseItemArray := []TagAnalyseItem{}
	//统计索引数组中的所有牌数
	cbCardCount := byte(0)
	for k := 0; k < TCGZ_MAX_INDEX; k++ {
		cbCardCount += cbCardIndex[k]
	}

	//效验数目，索引数组中牌的总数-2（一对将）后剩下的数是3的倍数
	//if ((cbCardCount + cbMagicNum<2)||(cbCardCount + cbMagicNum>HCSJ_MAXHANDCARD)||((cbCardCount  + cbMagicNum - 2)%3!=0)) {
	if cbCardCount+cbMagicNum < 2 {
		return false, AnalyseItemArray
	}

	//判断加入了待分析的牌后是否构成半句
	for i := byte(0); i < TCGZ_MAX_INDEX-1; i++ { //设置变量
		cardHalf := self.SearchHalf(cbCardIndex, i)
		if len(cardHalf) > 0 {
			//删除一个半句，如乙的话，乙己，乙乙,避免重复只往后找
			for _, is := range cardHalf {
				self.m_AnalyseItemArray = []TagAnalyseItem{}
				byTotalMagicNum := cbMagicNum //总赖子数
				byNeedMagicCount := byte(0)

				//如果牌索引数组中有一个对子(将牌)，可以胡牌时保存分析结果
				abTempCardIndex := [TCGZ_MAX_INDEX]byte{}
				copy(abTempCardIndex[:], cbCardIndex[:])

				if abTempCardIndex[i] > 0 {
					abTempCardIndex[i]--
				} else {
					byNeedMagicCount++
				}
				if abTempCardIndex[is] > 0 {
					abTempCardIndex[is]--
				} else {
					byNeedMagicCount++
				}
				if byNeedMagicCount > byTotalMagicNum {
					continue
				}
				byTotalMagicNum -= byNeedMagicCount

				var KindItemArray []TagKindItem
				self.Check_3N_Magic(abTempCardIndex, &byTotalMagicNum, &KindItemArray)

				if len(self.m_AnalyseItemArray) > 0 {
					for j := 0; j < len(self.m_AnalyseItemArray); j++ {
						//变量定义
						var AnalyseItem TagAnalyseItem

						//分析每一组组合牌，得到组合牌的组合牌型和中间牌，比如说WK_PENG,WK_CHI保存到分析子项中
						for w := byte(0); w < cbWeaveCount; w++ {
							AnalyseItem.WeaveKind[w] = WeaveItem[w].WeaveKind
							AnalyseItem.Cards[w][0] = WeaveItem[w].Cards[0]
							AnalyseItem.Cards[w][1] = WeaveItem[w].Cards[1]
							AnalyseItem.Cards[w][2] = WeaveItem[w].Cards[2]
							AnalyseItem.Cards[w][3] = WeaveItem[w].Cards[3]
							AnalyseItem.Cards[w][4] = WeaveItem[w].Cards[4]
							AnalyseItem.IsWeave[w] = 1 //组合牌区的牌
						}
						nKindItemCount := byte(10) //wWeaveKind数组长度为10 //sizeof(m_AnalyseItemArray[j].wWeaveKind)/sizeof(WORD);
						for k := byte(0); cbWeaveCount+k < nKindItemCount; k++ {
							AnalyseItem.WeaveKind[cbWeaveCount+k] = self.m_AnalyseItemArray[j].WeaveKind[k]
							AnalyseItem.Cards[cbWeaveCount+k][0] = self.m_AnalyseItemArray[j].Cards[k][0]
							AnalyseItem.Cards[cbWeaveCount+k][1] = self.m_AnalyseItemArray[j].Cards[k][1]
							AnalyseItem.Cards[cbWeaveCount+k][2] = self.m_AnalyseItemArray[j].Cards[k][2]
							AnalyseItem.Cards[cbWeaveCount+k][3] = self.m_AnalyseItemArray[j].Cards[k][3]
							AnalyseItem.Cards[cbWeaveCount+k][4] = self.m_AnalyseItemArray[j].Cards[k][4]
							AnalyseItem.IsWeave[cbWeaveCount+k] = 0 //不是组合牌区的牌
						}
						//将半句加入到分析结果中去
						nKindItemCount = 0
						for n := 0; n < 10; n++ {
							if AnalyseItem.WeaveKind[n] > 0 {
								if nKindItemCount < 9 {
									nKindItemCount++
								}
							}
						}
						AnalyseItem.WeaveKind[nKindItemCount] = ZP_WIK_HALF
						AnalyseItem.Cards[nKindItemCount][0] = self.SwitchToCardData(i)
						AnalyseItem.Cards[nKindItemCount][1] = self.SwitchToCardData(is)
						AnalyseItem.IsWeave[nKindItemCount] = 0 //不是组合牌区的牌
						//将分析结果插入到分析数组中
						AnalyseItemArray = append(AnalyseItemArray, AnalyseItem)
					}
				}
			}
		}

	}
	return (len(AnalyseItemArray) > 0), AnalyseItemArray
}

//分析扑克,去掉牌眼后的牌符合牌数量为3*N的要求，cbCardIndex：最终的点数数组，包含了要分析的牌
func (self *GameLogic_zp_jzhp) Check_3N_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) bool {
	//统计索引数组中的所有牌数
	cbCardCount := byte(0)
	for i := 0; i < TCGZ_MAX_INDEX; i++ {
		cbCardCount += cbCardIndex[i]
	}
	if cbCardCount == 0 && (*cbMagicNum) == 0 {
		//只能在这里保存这次找到的胡牌结果了
		//变量定义
		var AnalyseItem TagAnalyseItem
		nKindItemCount := len(*KindItemArray)
		for j := 0; j < nKindItemCount && j < 10; j++ {
			//wWeaveKind数组长度为10
			AnalyseItem.WeaveKind[j] = (*KindItemArray)[j].WeaveKind
			AnalyseItem.Cards[j][0] = (*KindItemArray)[j].CardIndex[0]
			AnalyseItem.Cards[j][1] = (*KindItemArray)[j].CardIndex[1]
			AnalyseItem.Cards[j][2] = (*KindItemArray)[j].CardIndex[2]
			AnalyseItem.Cards[j][3] = (*KindItemArray)[j].CardIndex[3]
			AnalyseItem.Cards[j][4] = (*KindItemArray)[j].CardIndex[4]
		}
		////多余的赖子组成砍，
		//if (*cbMagicNum >= 3 && (*cbMagicNum)%3 == 0){
		//	Leftcnt := (*cbMagicNum)/3
		//	for j:=0;j<int(Leftcnt) && (j+nKindItemCount < 10);j++{
		//		//wWeaveKind数组长度为10
		//		AnalyseItem.WeaveKind[j+nKindItemCount]=ZP_WIK_KAN;
		//		AnalyseItem.Cards[j+nKindItemCount][0]=11;//补成精“三”
		//		AnalyseItem.Cards[j+nKindItemCount][1]=11;
		//		AnalyseItem.Cards[j+nKindItemCount][2]=11;
		//	}
		//}
		self.m_AnalyseItemArray = append(self.m_AnalyseItemArray, AnalyseItem)
		return false //需要查找所有的可能性
	}

	//避免修改了cbCardIndex，需要定义两个临时数组
	abTempCardIndex1 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex2 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex3 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex4 := [TCGZ_MAX_INDEX]byte{}
	abTempCardIndex5 := [TCGZ_MAX_INDEX]byte{}
	bRet := false

	byTotalMagicNum5 := *cbMagicNum
	copy(abTempCardIndex5[:], cbCardIndex[:])
	//在删除一个滑
	if bRet, abTempCardIndex5 = self.Remove5Same_Magic(abTempCardIndex5, &byTotalMagicNum5, KindItemArray); bRet == true {
		if self.Check_3N_Magic(abTempCardIndex5, &byTotalMagicNum5, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}

	byTotalMagicNum4 := *cbMagicNum
	copy(abTempCardIndex4[:], cbCardIndex[:])
	//在删除一个观
	if bRet, abTempCardIndex4 = self.Remove4Same_Magic(abTempCardIndex4, &byTotalMagicNum4, KindItemArray); bRet == true {
		if self.Check_3N_Magic(abTempCardIndex4, &byTotalMagicNum4, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}

	copy(abTempCardIndex1[:], cbCardIndex[:])
	byTotalMagicNum := *cbMagicNum
	//先删除一个刻子，由于需要碰碰胡优先，所以需要先判断刻子
	//大冶字牌的坎(刻子)不参入分析，实际上不用执行Remove3Same的
	if bRet, abTempCardIndex1 = self.Remove3Same_Magic(abTempCardIndex1, &byTotalMagicNum, KindItemArray); bRet == true {
		if self.Check_3N_Magic(abTempCardIndex1, &byTotalMagicNum, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}
	byTotalMagicNum2 := *cbMagicNum
	copy(abTempCardIndex2[:], cbCardIndex[:])
	//在删除一个顺子
	if bRet, abTempCardIndex2 = self.RemoveStraight_Magic(abTempCardIndex2, &byTotalMagicNum2, KindItemArray); bRet == true {
		if self.Check_3N_Magic(abTempCardIndex2, &byTotalMagicNum2, KindItemArray) {
			return true
		}
		nKindItemCount := len(*KindItemArray)
		if nKindItemCount > 0 {
			*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
		}
	}
	byTotalMagicNum3 := *cbMagicNum
	copy(abTempCardIndex3[:], cbCardIndex[:])
	////在删除一个123
	//if bRet,abTempCardIndex3 = self.Remove1_2_3(abTempCardIndex3,KindItemArray);bRet == true  {
	//	if (self.Check_3N(abTempCardIndex3,KindItemArray)) {
	//		return true;
	//	}
	//	nKindItemCount:=len(*KindItemArray)
	//	if (nKindItemCount > 0) {
	//		*KindItemArray= (*KindItemArray)[:nKindItemCount-1] //删除最后一个
	//	}
	//}
	self.Remove1_2_3new_Magic(abTempCardIndex3, &byTotalMagicNum3, KindItemArray)

	return false
}

//删除最前面的一个顺子
func (self *GameLogic_zp_jzhp) RemoveStraight_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem
	byNeedMagicCount := byte(0)
	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		//风牌，只有刻子类型
		if i > 26 {
			continue
		}
		//if (((i%9) < 6 && cbCardIndex[i] >= 1) ||((i%9) == 6 && (cbCardIndex[i] >= 1 || (cbCardIndex[i+1] >= 1) || (cbCardIndex[i+2] >= 1))))
		if zp_tcgz_CardsStraight[i] > 0 && i < 22 && (cbCardIndex[i] >= 1 || cbCardIndex[i+1] >= 1 || cbCardIndex[i+2] >= 1) {
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i + 1)
			KindItem.CardIndex[2] = self.SwitchToCardData(i + 2)
			KindItem.WeaveKind = ZP_WIK_LEFT
			if cbCardIndex[i] == 0 {
				byNeedMagicCount++
			}
			if cbCardIndex[i+1] == 0 {
				byNeedMagicCount++
			}
			if cbCardIndex[i+2] == 0 {
				byNeedMagicCount++
			}
			if byNeedMagicCount > *cbMagicNum {
				//赖子不够配
				return false, cbCardIndex
			}
			if cbCardIndex[i] > 0 {
				cbCardIndex[i]--
			}
			if cbCardIndex[i+1] > 0 {
				cbCardIndex[i+1]--
			}
			if cbCardIndex[i+2] > 0 {
				cbCardIndex[i+2]--
			}
			*cbMagicNum -= byNeedMagicCount
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个乙二三等跳跃顺子，适合1、2、3...9、10杂乱无章的顺序。
func (self *GameLogic_zp_jzhp) Remove1_2_3new_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 {
			cards123 := zp_tcgz_Cards123new[i]
			if len(cards123) == 0 {
				break
			}
			for j := 0; j < len(cards123); j++ {
				byTotalMagicNum := *cbMagicNum
				abTempCardIndex0 := [TCGZ_MAX_INDEX]byte{}
				copy(abTempCardIndex0[:], cbCardIndex[:])
				eachcards123 := cards123[j]
				bRet := false
				//在删除一个123
				if bRet, abTempCardIndex0 = self.EachRemove1_2_3new_Magic(abTempCardIndex0, eachcards123, &byTotalMagicNum, KindItemArray); bRet == true {
					if self.Check_3N_Magic(abTempCardIndex0, &byTotalMagicNum, KindItemArray) {
						break
					}
					nKindItemCount := len(*KindItemArray)
					if nKindItemCount > 0 {
						*KindItemArray = (*KindItemArray)[:nKindItemCount-1] //删除最后一个
					}
				}
			}
			break
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个乙二三等跳跃顺子
func (self *GameLogic_zp_jzhp) EachRemove1_2_3new_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cards123new [3]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	if cards123new[0] >= TCGZ_MAX_INDEX || cards123new[1] >= TCGZ_MAX_INDEX || cards123new[2] >= TCGZ_MAX_INDEX {
		return false, cbCardIndex
	}
	//变量定义
	var KindItem TagKindItem
	byNeedMagicCount := byte(0)

	//删除1个顺子
	if (cbCardIndex[cards123new[0]] >= 1) || (cbCardIndex[cards123new[1]] >= 1) || (cbCardIndex[cards123new[2]] >= 1) {
		KindItem.CardIndex[0] = self.SwitchToCardData(cards123new[0])
		KindItem.CardIndex[1] = self.SwitchToCardData(cards123new[1])
		KindItem.CardIndex[2] = self.SwitchToCardData(cards123new[2])
		KindItem.WeaveKind = ZP_WIK_LEFT
		if cbCardIndex[cards123new[0]] == 0 {
			byNeedMagicCount++
		}
		if cbCardIndex[cards123new[1]] == 0 {
			byNeedMagicCount++
		}
		if cbCardIndex[cards123new[2]] == 0 {
			byNeedMagicCount++
		}
		if byNeedMagicCount > *cbMagicNum {
			//赖子不够配
			return false, cbCardIndex
		}
		if cbCardIndex[cards123new[0]] > 0 {
			cbCardIndex[cards123new[0]]--
		}
		if cbCardIndex[cards123new[1]] > 0 {
			cbCardIndex[cards123new[1]]--
		}
		if cbCardIndex[cards123new[2]] > 0 {
			cbCardIndex[cards123new[2]]--
		}
		*cbMagicNum -= byNeedMagicCount
		*KindItemArray = append(*KindItemArray, KindItem)
		return true, cbCardIndex
	}

	return false, cbCardIndex
}

//删除最前面的一个刻子
func (self *GameLogic_zp_jzhp) Remove3Same_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem
	byNeedMagicCount := byte(0)

	//删除1个顺子
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 {
			byNeedMagicCount = 0
			if cbCardIndex[i] < 3 {
				byNeedMagicCount = 3 - cbCardIndex[i]
			}
			if byNeedMagicCount > *cbMagicNum {
				//赖子不够配
				return false, cbCardIndex
			}
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i)
			KindItem.CardIndex[2] = self.SwitchToCardData(i)

			KindItem.WeaveKind = ZP_WIK_KAN
			cbCardIndex[i] -= (3 - byNeedMagicCount)
			*cbMagicNum -= byNeedMagicCount
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个观
func (self *GameLogic_zp_jzhp) Remove4Same_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem
	byNeedMagicCount := byte(0)

	//删除1个观
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 {
			byNeedMagicCount = 0
			if cbCardIndex[i] < 4 {
				byNeedMagicCount = 4 - cbCardIndex[i]
			}
			if byNeedMagicCount > *cbMagicNum {
				//赖子不够配
				return false, cbCardIndex
			}
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i)
			KindItem.CardIndex[2] = self.SwitchToCardData(i)
			KindItem.CardIndex[3] = self.SwitchToCardData(i)

			KindItem.WeaveKind = ZP_WIK_TIANLONG
			cbCardIndex[i] -= (4 - byNeedMagicCount)
			*cbMagicNum -= byNeedMagicCount
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//删除最前面的一个滑
func (self *GameLogic_zp_jzhp) Remove5Same_Magic(cbCardIndex [TCGZ_MAX_INDEX]byte, cbMagicNum *byte, KindItemArray *[]TagKindItem) (bool, [TCGZ_MAX_INDEX]byte) {
	//变量定义
	var KindItem TagKindItem
	byNeedMagicCount := byte(0)

	//删除1个滑
	for i := byte(0); i < TCGZ_MAX_INDEX; i++ {
		if cbCardIndex[i] >= 1 {
			byNeedMagicCount = 0
			if cbCardIndex[i] < 5 {
				byNeedMagicCount = 5 - cbCardIndex[i]
			}
			if byNeedMagicCount > *cbMagicNum {
				//赖子不够配
				return false, cbCardIndex
			}
			KindItem.CardIndex[0] = self.SwitchToCardData(i)
			KindItem.CardIndex[1] = self.SwitchToCardData(i)
			KindItem.CardIndex[2] = self.SwitchToCardData(i)
			KindItem.CardIndex[3] = self.SwitchToCardData(i)
			KindItem.CardIndex[4] = self.SwitchToCardData(i)

			KindItem.WeaveKind = ZP_WIK_HUA
			cbCardIndex[i] -= (5 - byNeedMagicCount)
			*cbMagicNum -= byNeedMagicCount
			*KindItemArray = append(*KindItemArray, KindItem)
			return true, cbCardIndex
		}
	}
	return false, cbCardIndex
}

//一张牌是否是红牌
func (self *GameLogic_zp_jzhp) IsHongPai(byData byte) byte {
	if !self.IsValidCard(byData) {
		return 0
	}
	cbCurrentIndex := self.SwitchToCardIndexNoHua(byData) //索引
	if cbCurrentIndex == 0 || cbCurrentIndex == 1 || cbCurrentIndex == 2 || cbCurrentIndex == 6 || cbCurrentIndex == 7 || cbCurrentIndex == 8 || cbCurrentIndex == 10 || cbCurrentIndex == 20 || cbCurrentIndex == 12 {
		return 1
	}
	return 0
}

//一组牌是否是红牌,返回红牌数目,返回0表示没有红牌
func (self *GameLogic_zp_jzhp) IsHongPaiSZ(byData []byte, byLen byte) (byAllHong byte, byHongPaiNum byte) {
	if byData == nil || byLen == 0 || byLen >= static.MAX_CARD {
		return 0, 0
	}
	byAllHong = 0
	byHongPaiNum = 0
	byCardNum := byte(0)
	for byCardIndex := byte(0); byCardIndex < byLen; byCardIndex++ {
		if !self.IsValidCard(byData[byCardIndex]) {
			continue
		}
		if 1 == self.IsHongPai(byData[byCardIndex]) {
			byHongPaiNum++
		}
		byCardNum++
	}
	if byCardNum == byHongPaiNum {
		byAllHong = 1
	} else {
		byAllHong = 0
	}

	return byAllHong, byHongPaiNum
}

//一张牌是否是花牌
func (self *GameLogic_zp_jzhp) GetHuaPaiNum(byData []byte) int {
	iHuaPaiNum := 0
	for _, data := range byData {
		if 1 == self.IsHuaPai(data) {
			iHuaPaiNum++
		}
	}
	return iHuaPaiNum
}

//一张牌是否是花牌
func (self *GameLogic_zp_jzhp) IsHuaPai(byData byte) byte {
	if !self.IsValidCard(byData) {
		return 0
	}
	if byData == 0x17 {
		return 1
	} else if byData == 0x18 {
		return 1
	} else if byData == 0x19 {
		return 1
	} else if byData == 0x1A {
		return 1
	} else if byData == 0x1B {
		return 1
	}
	return 0
}

//一张牌索引是否是精牌的索引  三五七为精牌
func (self *GameLogic_zp_jzhp) IsJingPaiIndex(byI byte) byte {
	if byI >= 27 {
		return 0
	}
	if byI == 10 {
		return 1
	} else if byI == 12 {
		return 1
	} else if byI == 20 {
		return 1
	} else if byI == 23 {
		return 1
	} else if byI == 024 {
		return 1
	} else if byI == 25 {
		return 1
	}
	return 0
}

//一张牌是否是精牌   三五七为精牌
func (self *GameLogic_zp_jzhp) IsJingPai(byData byte) byte {
	if !self.IsValidCard(byData) {
		return 0
	}
	if byData == 0x0B {
		return 1
	} else if byData == 0x0D {
		return 1
	} else if byData == 0x15 {
		return 1
	} else if byData == 0x18 {
		return 1
	} else if byData == 0x19 {
		return 1
	} else if byData == 0x1A {
		return 1
	}
	return 0
}

//获取单张牌的胡息（个子） 三五七为精牌
//带花精牌+2胡，无花精牌+1胡，非精花牌+1胡，无花非精牌没有胡数
func (self *GameLogic_zp_jzhp) GetSinglePaiHuxi(byData byte) int {
	if !self.IsValidCard(byData) {
		return 0
	}
	if byData == 0x17 {
		//花乙
		return 1
	} else if byData == 0x18 {
		//花三
		return 2
	} else if byData == 0x19 {
		//花五
		return 2
	} else if byData == 0x1A {
		//花七
		return 2
	} else if byData == 0x1B {
		//花九
		return 1
	} else if byData == 0x0005 {
		//素乙
		return 0
	} else if byData == 0x000B {
		//素三
		return 1
	} else if byData == 0x0015 {
		//素五
		return 1
	} else if byData == 0x000D {
		//素七
		return 1
	} else if byData == 0x0011 {
		//素九
		return 0
	}
	return 0
}

//计算胡牌的最大胡息（个子），返回值为0时表示不能胡牌
func (self *GameLogic_zp_jzhp) GetMaxHuxi(ChiHuResult TagChiHuResult) (retby byte, byMaxIndex byte, byMaxTotalHuXi int) {
	//BYTE	byMaxIndex= 0;//最大分数时索引
	//BYTE	byMaxTotalHuXi = 0;;//最大分数时总胡息
	byMaxTotalHuXi = 0
	retby = 0
	if ChiHuResult.ChiHuKind != ZP_WIK_NULL {
		//
		byHuInfoNum := len(ChiHuResult.ChiHuItemInfoArray)
		for byHuIndex := 0; byHuIndex < byHuInfoNum; byHuIndex++ {
			byTotalHuXi := ChiHuResult.ChiHuItemInfoArray[byHuIndex].TotalHuxi
			//记录最大的
			if byTotalHuXi > byMaxTotalHuXi {
				byMaxIndex = byte(byHuIndex) //最大分数时索引
				byMaxTotalHuXi = byTotalHuXi //最大分数时总胡息
			}
		}
	} else {
		retby = 0
	}
	if byMaxTotalHuXi > 0 {
		retby = 1
	}
	return retby, byMaxIndex, byMaxTotalHuXi
}

//计算胡牌的分数
//iHuXi 总个子数，iBaseGeZi 起胡个子数，iFenType算分类型，iPaoFen 跑分，iBaseScore底分
func (self *GameLogic_zp_jzhp) GetHuFen(iHuXi int, iBaseGeZi int, iFenType int, iPaoFen int, iBaseScore int) (iScore int) {
	//iFenType == 0 表示按照胡数来算，有多少胡数就多少分；1表示按照坡算分，5分一坡
	iBase := 1
	if iFenType == 1 {
		if iBaseGeZi == 0 {
			//选了算坡时默认17起
			iBaseGeZi = 17
		}
		iBase = 1
		if iHuXi > iBaseGeZi {
			iBase += ((iHuXi - iBaseGeZi) / 5)
		}
	} else {
		iBase = iHuXi
	}
	if iBase < 1 {
		iBase = 1 //最少1个基础积分
	}

	iScore = iBase*iBaseScore + iPaoFen

	return iScore
}

//转换牌名 十六进制
func (self *GameLogic_zp_jzhp) SwitchToCardName(cbCardIndex []byte) string {
	szCardName := ""
	for i := 0; i < TCGZ_MAX_INDEX_HUA && i < len(cbCardIndex); i++ {
		if cbCardIndex[i] > 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				szCardName += zp_tcgz_strCardsMessage[i] + ","
			}
		}
	}
	return szCardName
}

//转换牌名 十六进制
func (self *GameLogic_zp_jzhp) SwitchToCardName4(cbCardIndex byte) string {
	if cbCardIndex > TCGZ_MAX_INDEX_HUA {
		return ""
	}
	return zp_tcgz_strCardsMessage[cbCardIndex]
}

//转换牌名 十六进制
func (self *GameLogic_zp_jzhp) SwitchToCardName5(cbCardData []byte, cbCardCount byte) string {
	szCardName := ""

	for i := byte(0); i < cbCardCount; i++ {
		if !self.IsValidCard(cbCardData[i]) {
			cardstr := fmt.Sprintf("%d", cbCardData[i])
			szCardName += cardstr
			continue
		}
		index := self.SwitchToCardIndex(cbCardData[i])
		if self.IsValidCard(cbCardData[index]) {
			szCardName += zp_tcgz_strCardsMessage[index] + ","
		}
	}
	return szCardName
}

//转换牌名，汉字
func (self *GameLogic_zp_jzhp) SwitchToCardName1(cbCardIndex []byte) string {
	szCardName := ""
	for i := 0; i < TCGZ_MAX_INDEX_HUA && i < len(cbCardIndex); i++ {
		if cbCardIndex[i] > 0 {
			for j := byte(0); j < cbCardIndex[i]; j++ {
				szCardName += zp_tcgz_strCardsMessage1[i]
			}
		}
	}
	return szCardName
}

//转换牌名，汉字
func (self *GameLogic_zp_jzhp) SwitchToCardName2(cbCardIndex byte) string {
	if cbCardIndex > TCGZ_MAX_INDEX_HUA {
		return ""
	}
	return zp_tcgz_strCardsMessage1[cbCardIndex]
}

//转换牌名，汉字
func (self *GameLogic_zp_jzhp) SwitchToCardName3(cbCardData []byte, cbCardCount byte) string {
	szCardName := ""

	for i := byte(0); i < cbCardCount; i++ {
		if !self.IsValidCard(cbCardData[i]) {
			cardstr := fmt.Sprintf("%d", cbCardData[i])
			szCardName += cardstr
			continue
		}
		index := self.SwitchToCardIndex(cbCardData[i])
		if index < TCGZ_MAX_INDEX_HUA {
			szCardName += zp_tcgz_strCardsMessage1[index] + ","
		}
	}
	return szCardName
}

/****************************************************************************************************/

func (self *GameLogic_zp_jzhp) GetStringByCard(carddata byte) string {
	strColor := ""
	strPoint := ""

	if carddata == 0 || carddata > 28 {
		return strColor
	}

	//牌值
	strPoint = fmt.Sprintf("%02d", carddata)

	strColor += strPoint

	return strColor
}

func (self *GameLogic_zp_jzhp) GetWriteHandReplayRecordString(replayRecord ZP_Replay_Record) string {
	handCardStr := ""
	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < len(replayRecord.R_HandCards[i]); j++ {
			if self.IsValidCard(byte(replayRecord.R_HandCards[i][j])) {
				handCardStr += fmt.Sprintf("%s,", self.GetStringByCard(byte(replayRecord.R_HandCards[i][j])))
			}
		}
		handCardStr += fmt.Sprintf("|")
	}

	//写入分数
	handCardStr += "S:"
	for i := 0; i < TCGZ_MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d", replayRecord.R_Score[i])
		//最后一位不要"，"
		if i < TCGZ_MAX_PLAYER-1 {
			handCardStr += fmt.Sprintf(",")
		}
	}
	handCardStr += fmt.Sprintf("|LC:%d", replayRecord.R_LeftCount) //剩余数目
	return handCardStr
}

func (self *GameLogic_zp_jzhp) GetWriteOutReplayRecordString(replayRecord ZP_Replay_Record) string {
	upd := false
	endMsgUpdateScore := [TCGZ_MAX_PLAYER]float64{}
	ourCardStr := ""
	for k, record := range replayRecord.R_Orders {
		// 如果是分数变化 ,为什么不能写在case里面？
		if record.R_Opt == info2.E_GameScore {
			flag := false
			// 如果是最后结算的update 则挪到最后追加
			for j := k; j < len(replayRecord.R_Orders); j++ {
				// 只要后面还有别的操作，说明是中途及时结算，按正常的逻辑走
				if replayRecord.R_Orders[j].R_Opt != info2.E_GameScore {
					flag = true
					break
				}
			}
			// 如果是中途结算 或者没有大胡 这里就直接追加
			if flag {
				ourCardStr += fmt.Sprintf(",%d:", record.R_ChairId)
				if fs := strings.Split(fmt.Sprintf("%v", record.UserScorePL), "."); len(fs) == 1 {
					ourCardStr += fmt.Sprintf("U%s", fs[0])
				} else {
					ourCardStr += fmt.Sprintf("U%0.2f", record.UserScorePL)
				}
			} else {
				upd = true
				endMsgUpdateScore[record.R_ChairId] = record.UserScorePL
			}
			continue
		}

		switch record.R_Opt {
		case E_ZP_SendCard:
			ourCardStr += fmt.Sprintf("|SA%d:", record.R_ChairId)
			break
		case E_ZP_OutCard:
			ourCardStr += fmt.Sprintf("|OA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_Left:
			ourCardStr += fmt.Sprintf("|LA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_Center:
			ourCardStr += fmt.Sprintf("|CA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_Right:
			ourCardStr += fmt.Sprintf("|RA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_1X2D:
			ourCardStr += fmt.Sprintf("|XA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_2X1D:
			ourCardStr += fmt.Sprintf("|YA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_2710:
			ourCardStr += fmt.Sprintf("|ZA%d:", record.R_ChairId)
			break
		case E_ZP_TianLong:
			ourCardStr += fmt.Sprintf("|TA%d:", record.R_ChairId)
			break
		case E_ZP_Peng:
			ourCardStr += fmt.Sprintf("|PA%d:", record.R_ChairId)
			break
		case E_ZP_Gang:
			ourCardStr += fmt.Sprintf("|GA%d:", record.R_ChairId)
			break
		case E_ZP_Hu:
			ourCardStr += fmt.Sprintf("|HA%d:", record.R_ChairId)
			break
		case E_ZP_HuangZhuang:
			ourCardStr += fmt.Sprintf("|NA%d:", record.R_ChairId)
			break
		case E_ZP_Li_Xian:
			ourCardStr += fmt.Sprintf("|LB%d:", record.R_ChairId)
			break
		case E_ZP_Jie_san:
			ourCardStr += fmt.Sprintf("|JA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_Hua:
			ourCardStr += fmt.Sprintf("|HB%d:", record.R_ChairId)
			break
		case E_ZP_Wik_Jian:
			ourCardStr += fmt.Sprintf("|JB%d:", record.R_ChairId)
			break
		case E_ZP_PIAO:
			ourCardStr += fmt.Sprintf("|PB%d:", record.R_ChairId)
			break
		case E_ZP_Tuo_Guan:
			ourCardStr += fmt.Sprintf("|DA%d:", record.R_ChairId)
			break
		case E_ZP_Wik_Ta:
			ourCardStr += fmt.Sprintf("|TB%d:", record.R_ChairId)
			break
		case E_ZP_NoOut:
			ourCardStr += fmt.Sprintf("|NB%d:", record.R_ChairId)
			break
		case E_SendCardRight:
			ourCardStr += fmt.Sprintf("|AA%d:", record.R_ChairId)
			break
		case E_HandleCardRight:
			ourCardStr += fmt.Sprintf("|BA%d:", record.R_ChairId)
			break
		default:
			break
		}
		//写牌数据
		//胡牌数据
		if record.R_Opt == E_ZP_Hu {
			for iCount := byte(0); iCount < replayRecord.R_WeaveCount && iCount < 10; iCount++ {
				for iCardCount := 0; iCardCount < 4; iCardCount++ {
					if self.IsValidCard(replayRecord.R_WeaveItemInfo[iCount][iCardCount]) {
						ourCardStr += fmt.Sprintf("%s", self.GetStringByCard(replayRecord.R_WeaveItemInfo[iCount][iCardCount]))
					}
				}
				ourCardStr += fmt.Sprintf("+")
			}

		} else {
			if len(record.R_Value) > 0 {
				for i := 0; i < len(record.R_Value); i++ {
					if record.R_Opt == E_ZP_Tuo_Guan && byte(record.R_Value[i]) == 0 {
						ourCardStr += fmt.Sprintf("%02d", byte(record.R_Value[i]))
					} else if record.R_Opt == E_SendCardRight || record.R_Opt == E_HandleCardRight {
						ourCardStr += fmt.Sprintf("%02d", record.R_Value[i])
					} else {
						ourCardStr += fmt.Sprintf("%s", self.GetStringByCard(byte(record.R_Value[i])))
					}
				}
			}
		}
		//写附加信息，例如本轮胡息等
		if len(record.R_Opt_Ext) > 0 {
			for i := 0; i < len(record.R_Opt_Ext); i++ {
				if record.R_Opt_Ext[i].Ext_type == E_ZP_Ext_HuXi {
					ourCardStr += fmt.Sprintf(",EA%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == E_ZP_Ext_Provider {
					ourCardStr += fmt.Sprintf(",EC%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == E_ZP_Ext_ProvideCard {
					ourCardStr += fmt.Sprintf(",ED%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == E_ZP_Ext_PiaoScore {
					ourCardStr += fmt.Sprintf(",EE%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == E_ZP_Ext_Jian {
					ourCardStr += fmt.Sprintf(",EF%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == E_ZP_Ext_Tong {
					ourCardStr += fmt.Sprintf(",EH%d", record.R_Opt_Ext[i].Ext_value)
				}
			}
		}
	}
	ourCardStr += fmt.Sprintf("|HT:")
	switch replayRecord.R_HuType {
	case enQuanHong:
		ourCardStr += fmt.Sprintf("QB,")
		break
	case enShiDui:
		ourCardStr += fmt.Sprintf("SD,")
		break
	case enQuanHei:
		ourCardStr += fmt.Sprintf("QY,")
		break
	default:
		ourCardStr += fmt.Sprintf("NIL,")
		break
	}

	ourCardStr += fmt.Sprintf("|HX:%d,", replayRecord.R_TotalHuxi)
	ourCardStr += fmt.Sprintf("|GS:%d,", replayRecord.R_GeZiShu)
	ourCardStr += fmt.Sprintf("|HF:%d,", replayRecord.R_HuFen)
	ourCardStr += fmt.Sprintf("|ET:%d,", replayRecord.R_EndSubType)

	if upd {
		for i, s := range endMsgUpdateScore {
			ourCardStr += fmt.Sprintf(",%d:", i)
			if fs := strings.Split(fmt.Sprintf("%v", s), "."); len(fs) == 1 {
				ourCardStr += fmt.Sprintf("U%s", fs[0])
			} else {
				ourCardStr += fmt.Sprintf("U%0.2f", s)
			}
		}
	}

	return ourCardStr
}
