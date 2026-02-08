package backboard

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"math/rand"
	"reflect"
	"time"
)

var G_CardSerial = [static.MAX_INDEX]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09",
	"0x11", "0x12", "0x13", "0x14", "0x15", "0x16", "0x17", "0x18", "0x19",
	"0x21", "0x22", "0x23", "0x24", "0x25", "0x26", "0x27", "0x28", "0x29",
	"0x31", "0x32", "0x33", "0x34", "0x35", "0x36", "0x37",
}

//牌型别名
var G_CardAnother = [static.MAX_INDEX]string{
	"1万", "2万", "3万", "4万", "5万", "6万", "7万", "8万", "9万",
	"1条", "2条", "3条", "4条", "5条", "6条", "7条", "8条", "9条",
	"1同", "2同", "3同", "4同", "5同", "6同", "7同", "8同", "9同",
	"东风", "南风", "西风", "北风", "红中", "发财", "白板",
}

//扑克数据
var G_cbCardDataArray = [static.MAX_REPERTORY]byte{
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

//20190108 目测，极大话的处理，可能会出现只要风牌的跳牌选择，那就比较蛋疼了，将来再看吧
const (
	CARDS_NOMOR          = 0xFF   //全牌   1111 1111
	CARDS_WITHOUT_WAN    = 1      // 去万 1111 1110
	CARDS_WITHOUT_BAMBOO = 1 << 1 // 去条 1111 1101
	CARDS_WITHOUT_DOT    = 1 << 2 // 去筒 1111 1011
	CARDS_WITHOUT_WIND   = 1 << 3 //去风  1111 0111
	CARDS_WITHOUT_DRAGON = 1 << 4 //去箭  1110 1111  现在如果去箭，就全去了，如果不去，再判断是不是保留了中发白
	CARDS_WITHOUT_ZHONG  = 1 << 5 //去中  1101 1111
	CARDS_WITHOUT_FA     = 1 << 6 //去发  1011 1111
	CARDS_WITHOUT_BAI    = 1 << 7 //去白  0111 1111

)

//20181203 添加，追加新孤章，比如去了风字后，还要留了红中。。。。。。
type CardLibrary struct {
	cardclass   int    //基础牌库类型
	specialcard []byte //特殊的孤章 20181207 因调整后，基本不用了，这个先放置
}

//可执行操作的结构 目前用做杠碰
type CardOpeInfo struct {
	CardID         byte `json:"cardid"`         //杠碰牌标志就是一张
	OpeChiPengGang byte `json:"opechipenggang"` //8位暂时够了
}

//可操作的选项
func NewCardOpe(cardID byte, operation byte) (*CardOpeInfo, error) {
	if !IsMahjongCard(cardID) {
		return nil, errors.New(fmt.Sprintf("牌值（%d）~（%d），传入值（%d）", 1, static.MAX_INDEX, cardID))
	}
	newCardOpe := &CardOpeInfo{
		CardID:         cardID,
		OpeChiPengGang: operation,
	}
	return newCardOpe, nil
}

//下标转换成牌编号
func IndexToCardSafe(cbCardIndex byte) (byte, error) {
	if cbCardIndex >= static.MAX_INDEX {
		return static.INVALID_BYTE, errors.New(fmt.Sprintf("牌型最大值（%d），传入值（%d）", static.MAX_INDEX, cbCardIndex))
	}
	return ((cbCardIndex / 9) << 4) | (cbCardIndex%9 + 1), nil
}

func IndexToCard(cbCardIndex byte) (card byte) {
	var err error
	card, err = IndexToCardSafe(cbCardIndex)
	if err != nil {
		fmt.Println(err)
	}
	return
}

//20181226 检查是什么花色的牌 5位值
/*
花色      牌值
万
1         1-9
条
1《1       1-9
筒
1《2       1-9
风
1《3       1-4
箭
1《4       5-7
*/
func DifferentColor(card byte) (byte, byte) {
	err, color, value := Cardsplit(card)
	if err != nil {
		return 0, 0
	}
	if color > 2 && value > 4 {
		//箭牌
		return CARDS_WITHOUT_DRAGON, value
	}
	return 1 << color, value
}

//牌编号成花色和牌值
func Cardsplit(card byte) (error, byte, byte) {
	if card == static.INVALID_BYTE {
		return nil, static.INVALID_BYTE, static.INVALID_BYTE
	}
	if card == 0 || card > 0x37 {
		return errors.New(fmt.Sprintf("Cardsplit牌编号最大值（%b），传入值（%b）", static.MAX_INDEX, card)), static.INVALID_BYTE, static.INVALID_BYTE
	}
	return nil, card >> 4, card & static.MASK_VALUE
}

//组合牌
func CombinCardSafe(color byte, value byte) (error, byte) {
	card := color<<4 | value
	if card == 0 || card > 0x37 {
		return errors.New(fmt.Sprintf("牌编号最大值（%b），传入花色值（%b），传入value（%b）", 0x37, color, value)), static.INVALID_BYTE
	}
	return nil, card
}

//组合牌
func CombinCard(color byte, value byte) byte {
	return color<<4 | value
}

//牌编号转变成下标
func CardToIndex(cbCardData byte) (byte, error) {
	err, color, value := Cardsplit(cbCardData)
	if err != nil || color == static.INVALID_BYTE {
		return static.INVALID_BYTE, err
	}
	return color*9 + value - 1, nil
}

// 必须是麻将牌值
func IsMahjongCard(cbCardData ...byte) bool {
	for _, card := range cbCardData {
		if !(card > 0 && card <= 0x37) {
			return false
		}
	}
	return true
}

// 必须是麻将牌类别
func IsMahjongClass(cardclass int) (error, bool) {
	if cardclass == 0 || cardclass > CARDS_NOMOR {
		return errors.New(fmt.Sprintf("牌库最大支持（%d）,传入值（%d）", CARDS_NOMOR, cardclass)), false
	}
	return nil, true

}
func IsLegalCardAll(cardclass int, specialcard []byte, cbCardData byte) (err error, color byte, value byte) {
	if len(specialcard) != 0 {
		err, color, value = IsLegalCardSpecial(specialcard, cbCardData)
	}
	if err != nil {
		return err, 0, 0
	}
	if value != 0 {
		return nil, color, value
	}
	return IsLegalCard(cardclass, cbCardData)
}

func IsLegalCardSpecial(specialcard []byte, cbCardData byte) (error, byte, byte) {
	//根据cordclass做进一步确认
	err, color, value := Cardsplit(cbCardData)
	if err != nil {
		return err, 0, 0
	}
	if len(specialcard) == 0 {
		return nil, 0, 0
	}
	for _, v := range specialcard {
		if v == cbCardData {
			return nil, color, value
		}
	}
	return nil, 0, 0
}

//根据每个牌库的规则进行二次判断.目前value不可能为0的
func IsLegalCard(cardclass int, cbCardData byte) (error, byte, byte) {

	//根据cordclass做进一步确认
	err, color, value := Cardsplit(cbCardData)
	if err != nil {
		return err, 0, 0
	}
	checkindex := color
	if color < 3 {
		if cardclass&(1<<checkindex) == 0 {
			v, _ := CardToIndex(cbCardData)
			xlog.Logger().Debug(fmt.Sprintf("库牌不支持（%s）", G_CardAnother[v]))
			return nil, 0, 0
		}
		return nil, color, value
	}
	//检查风和字要分开 字牌+1，目前如果没有字牌就不用检查了，如果有字牌，因调整，红中，发财，白板变成了更高级别的了
	if value > 4 {
		checkindex += 1
		if cardclass&(1<<checkindex) == 0 {
			v, _ := CardToIndex(cbCardData)
			xlog.Logger().Debug(fmt.Sprintf("库牌不支持（%s）", G_CardAnother[v]))
			return nil, 0, 0
		}
		//2018127，根据调整，来个提升
		switch value {
		case 5:
			checkindex += 1
		case 6:
			checkindex += 2
		case 7:
			checkindex += 3
		}
		if cardclass&(1<<checkindex) == 0 {
			v, _ := CardToIndex(cbCardData)
			xlog.Logger().Debug(fmt.Sprintf("库牌不支持（%s）", G_CardAnother[v]))
			return nil, 0, 0
		}
		return nil, color, value
	}
	if cardclass&(1<<checkindex) == 0 {
		v, _ := CardToIndex(cbCardData)
		xlog.Logger().Debug(fmt.Sprintf("库牌不支持（%s）", G_CardAnother[v]))
		return nil, 0, 0
	}
	return nil, color, value

}

//有效判断 20181122修改
func IsValidCard(cardclass int, cbCardData byte) bool {
	if !IsMahjongCard(cbCardData) {
		return false
	}
	//根据cordclass再检查
	_, _, value := IsLegalCard(cardclass, cbCardData)
	return value != 0
}

//20190307 潜江红中 12张风中 3张白板 1张东风
/*
{
code
num
}
*/
func CreateCardLibrary(cardclass int, specialcard ...byte) (error, []byte) {
	err, cards := CreateCards(cardclass)
	if err != nil {
		return err, nil
	}
	if !IsMahjongCard(specialcard...) {
		return errors.New(fmt.Sprintf("特殊添加的牌值越界（%v）", specialcard)), nil
	}
	cards = append(cards, specialcard...)
	//if len(specialcard) != 0 {
	//	for _, v := range specialcard {
	//		if !IsMahjongCard(v) {
	//			return errors.New(fmt.Sprintf("特殊牌库的牌值越界（%x）", v)), nil
	//		}
	//		cards = append(cards, v)
	//	}
	//}
	return nil, cards
}

//20181119 创建牌库的牌
/*参数说明
1、牌型值
*/
func CreateCards(cardclass int) (error, []byte) {
	//目前没有4季牌
	if cardclass == 0 || cardclass > CARDS_NOMOR {
		return errors.New(fmt.Sprintf("classify值异常目前全牌最大值31，传入值（%d）", cardclass)), nil
	}
	tempCards := []byte{}
	var i uint8
	//万条筒
	for i = 0; i < 3; i++ {
		if cardclass&(1<<i) > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[4*9*i:4*(i+1)*9]...)
		}
	}
	//风
	if cardclass&CARDS_WITHOUT_WIND > 0 {
		tempCards = append(tempCards, G_cbCardDataArray[4*9*3:(4*9*3)+16]...)
	}
	//中发白
	if cardclass&CARDS_WITHOUT_DRAGON > 0 {
		if cardclass&CARDS_WITHOUT_ZHONG > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[(4*9*3)+4*4:(4*9*3)+4*5]...)
		}
		if cardclass&CARDS_WITHOUT_FA > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[(4*9*3)+4*5:(4*9*3)+4*6]...)
		}
		if cardclass&CARDS_WITHOUT_BAI > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[(4*9*3)+4*6:]...)
		}
	}
	return nil, tempCards
}

//20201030 苏大强 卡五星转用
/*
要留一色必然在 1~9之间会留一张，默认是0，全有，byte不够，9张的话只能用word了
至少写3个，因为万、条、同
风字暂不判断
*/
func CreateCards_k5x(cardclass int, mask []uint16) (error, []byte) {
	//目前没有4季牌
	if cardclass == 0 || cardclass > CARDS_NOMOR {
		return errors.New(fmt.Sprintf("classify值异常目前全牌最大值31，传入值（%d）", cardclass)), nil
	}
	tempCards := []byte{}
	var i uint16
	//万条筒
	for i = 0; i < 3; i++ {
		if cardclass&(1<<i) > 0 {
			if mask[i] == 0 {
				tempCards = append(tempCards, G_cbCardDataArray[4*9*i:4*(i+1)*9]...)
			} else {
				//从底到高开始
				var j uint16
				for j = 0; j < 9; j++ {
					//如果是0就要
					if mask[i]&(1<<j) == 0 {
						for z := 0; z < 4; z++ {
							tempCards = append(tempCards, G_cbCardDataArray[4*9*i+j])
						}
					}
				}
			}
		}
	}
	//风
	if cardclass&CARDS_WITHOUT_WIND > 0 {
		tempCards = append(tempCards, G_cbCardDataArray[4*9*3:(4*9*3)+16]...)
	}
	//中发白
	if cardclass&CARDS_WITHOUT_DRAGON > 0 {
		if cardclass&CARDS_WITHOUT_ZHONG > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[(4*9*3)+4*4:(4*9*3)+4*5]...)
		}
		if cardclass&CARDS_WITHOUT_FA > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[(4*9*3)+4*5:(4*9*3)+4*6]...)
		}
		if cardclass&CARDS_WITHOUT_BAI > 0 {
			tempCards = append(tempCards, G_cbCardDataArray[(4*9*3)+4*6:]...)
		}
	}
	return nil, tempCards
}

//这个要跟牌型检查的 万条筒是一次挖空，我们这里风红发财在一起，处理起来要麻烦点
//2018123 不支持牌库中孤章红中的情况，切记，目前只支持挖了风，或者挖了字的情况
//20181207 再次提示，不支持牌中出现孤章，必然字牌中只有一个红中的情况，这个情况可以在各自的逻辑中单独添加
func FindMagicCard(cardclass int, PiZiCard byte) (byte, error) {
	err, color, value := Cardsplit(PiZiCard)
	if err != nil {
		return 0, err
	}
	checkindex := color
	//检查风和字要分开
	if color == 3 && value > 4 {
		checkindex += 1
	}
	if cardclass&(1<<checkindex) == 0 {
		v, _ := CardToIndex(PiZiCard)
		return 0, errors.New(fmt.Sprintf("库牌中不支持赖子（%s）mask(%b)", G_CardAnother[v], cardclass))
	}
	if color < 3 {
		//万条筒的
		if value == 9 {
			value = 0
		}
	} else {
		//去风的 生成牌的时候不可能有1234
		//若去中发白，目前（风和字在一起）则这里7就不可能是value 所以还是有局限的
		switch value {
		case 7:
			//去 风
			if cardclass&(CARDS_WITHOUT_WIND) == 0 {
				value = 4
			} else {
				value = 0
			}
		case 4:
			//去 中发白
			if cardclass&(CARDS_WITHOUT_DRAGON) == 0 {
				value = 0
			}
		}
	}
	//	fmt.Println(fmt.Sprintf("皮子（%s）,赖子（%s）", g_CardAnother[CardToIndex(PiZiCard)], g_CardAnother[CardToIndex(CombinCard(color, value+1))]))
	return CombinCard(color, value+1), nil
}

func FindPiziCard(cardclass int, PiZiCard byte) (byte, error) {
	err, color, value := Cardsplit(PiZiCard)
	if err != nil {
		return 0, err
	}
	checkindex := color
	//检查风和字要分开
	if color == 3 && value > 4 {
		checkindex += 1
	}
	if cardclass&(1<<checkindex) == 0 {
		v, _ := CardToIndex(PiZiCard)
		return 0, errors.New(fmt.Sprintf("库牌中不支持皮子（%s）", G_CardAnother[v]))
	}
	if color < 3 {
		//万条筒的
		if value == 1 {
			value = 10
		}
	} else {
		//去风的 生成牌的时候不可能有1234
		//若去中发白，目前（风和字在一起）则这里7就不可能是value 所以还是有局限的
		switch value {
		case 5:
			//去 风
			if cardclass&(CARDS_WITHOUT_WIND) == 0 {
				value = 8
			}
		case 1:
			//去 中发白
			if cardclass&(CARDS_WITHOUT_DRAGON) == 0 {
				value = 5
			} else {
				value = 8
			}
		}
	}
	//	fmt.Println(fmt.Sprintf("皮子（%s）,赖子（%s）", g_CardAnother[CardToIndex(PiZiCard)], g_CardAnother[CardToIndex(CombinCard(color, value+1))]))
	return CombinCard(color, value-1), nil
}

//混乱扑克
func RandCardData(repertoryCard []byte) (byte, []byte) {
	cbCardData := [static.MAX_REPERTORY]byte{}
	cbCardDataTemp := make([]byte, static.MAX_REPERTORY, static.MAX_REPERTORY)
	var cbMaxCount byte = byte(len(repertoryCard))
	copy(cbCardDataTemp, repertoryCard[:])
	//混乱扑克
	cbRandCount, cbPosition := 0, 0
	randTmp := 0
	nAccert := 0
	for {
		nAccert++
		if nAccert > 200 {
			//			m_mylog.Log("混乱扑克时死循环啦")
			break
		}
		randTmp = int(cbMaxCount) - cbRandCount - 1
		if randTmp > 0 {
			cbPosition = rand.Intn(randTmp)
		} else {
			cbPosition = 0
		}
		//cbPosition=rand()%(cbMaxCount-cbRandCount);
		cbCardData[cbRandCount] = cbCardDataTemp[cbPosition]
		cbRandCount++
		cbCardDataTemp[cbPosition] = cbCardDataTemp[int(cbMaxCount)-cbRandCount]
		if cbRandCount >= int(cbMaxCount) {
			break
		}
	}
	// fmt.Println(fmt.Sprintf("混牌后的长度（%d）", len(cbCardData)))
	result := removezero(cbCardData[:])
	// fmt.Println(fmt.Sprintf("去0后的长度（%d）", len(result)))
	return cbMaxCount, result
}

//手牌转换成牌型表
func CardsToCardIndex(cbCardData []byte) [static.MAX_INDEX]byte {
	//设置变量
	cbCardIndex := [static.MAX_INDEX]byte{}
	//转换扑克
	for i := 0; i < len(cbCardData); i++ {
		if !IsMahjongCard(cbCardData[i]) {
			return [static.MAX_INDEX]byte{}
		}
		v, err := CardToIndex(cbCardData[i])
		if err != nil {
			xlog.Logger().Debug(fmt.Sprintf("手牌转换成牌型表出错，不支持（%s）", G_CardAnother[v]))
			continue
		}

		cbCardIndex[v]++
	}

	return cbCardIndex
}

//这个函数会根据实际情况调整，目前还不能完全定型
//目前对风牌的处理是一次挖干净
/*
20190108 定型  由参数一生成牌型默认，由参数二做进一步处理，比如有三人玩法还要再去万的情况
考虑到特殊的处理也可能出现多挖的情况，感觉只要大家知道判型的值就行了，这个功能有点鸡肋了
参数一：由gameID（游戏ID）修改为baseCardsClass(初始牌库类型需要丢弃的牌)
参数二：当玩家人数为3人的时候，再去除的牌（目前环境是这样的）
*/
func SetCardsClass(baseCardsClass int, special int) (cardsclass int) {
	//两个参数如果有问题就直接出去，不再生成
	if baseCardsClass < 0 || baseCardsClass > CARDS_NOMOR || special < 0 || special > baseCardsClass {
		xlog.Logger().Debug(fmt.Sprintf("输入参数有误baseCardsClass（%d），special（%d），范围值（%d）", baseCardsClass, special, CARDS_NOMOR))
		return 0
	}
	cardsclass = CARDS_NOMOR
	//去掉初始丢弃牌
	cardsclass ^= baseCardsClass
	//20190108 目前只看到箭牌会出现只要其中一张或挑张的选择
	switch special {
	//------------目前再次剔除只会在下面一种选择----------------------------
	case CARDS_WITHOUT_WAN: //去万
		cardsclass ^= CARDS_WITHOUT_WAN
	case CARDS_WITHOUT_BAMBOO: //去条
		cardsclass ^= CARDS_WITHOUT_BAMBOO
	case CARDS_WITHOUT_DOT: //去筒
		cardsclass ^= CARDS_WITHOUT_DOT
	case CARDS_WITHOUT_WIND: //去风
		cardsclass ^= CARDS_WITHOUT_WIND
	case CARDS_WITHOUT_DRAGON: //去字 20181207 如果去字，就是去掉中发白 下面可能要用到 fallthrough
		cardsclass ^= CARDS_WITHOUT_DRAGON
	//-------------------------------------------
	//20190108 从这里开始 中发白支持跳章
	case CARDS_WITHOUT_ZHONG: //去中
		cardsclass ^= CARDS_WITHOUT_ZHONG
		fallthrough
	case CARDS_WITHOUT_FA: //去发
		cardsclass ^= CARDS_WITHOUT_FA
		fallthrough
	case CARDS_WITHOUT_BAI: //去白
		cardsclass ^= CARDS_WITHOUT_BAI
	}
	return cardsclass
}
func SetCardsClass_ex(baseCardsClass int, special int) (cardsclass int) {
	//两个参数如果有问题就直接出去，不再生成
	if baseCardsClass < 0 || baseCardsClass > CARDS_NOMOR || special < 0 {
		xlog.Logger().Debug(fmt.Sprintf("输入参数有误baseCardsClass（%d），special（%d），范围值（%d）", baseCardsClass, special, CARDS_NOMOR))
		return 0
	}
	cardsclass = CARDS_NOMOR
	//去掉初始丢弃牌
	cardsclass ^= baseCardsClass
	//20190108 目前只看到箭牌会出现只要其中一张或挑张的选择
	if special > 0 {
		cardsclass ^= special
	}
	return cardsclass
}

func Print_cards(cards []byte) {
	for i := 0; i < 9; i++ {
		fmt.Printf("%d,", cards[i])
	}
	fmt.Printf("\n")

	for i := 9; i < 18; i++ {
		fmt.Printf("%d,", cards[i])
	}
	fmt.Printf("\n")

	for i := 18; i < 27; i++ {
		fmt.Printf("%d,", cards[i])
	}
	fmt.Printf("\n")

	for i := 27; i < 34; i++ {
		fmt.Printf("%d,", cards[i])
	}
	fmt.Printf("\n")
}

//去除切片中的0值
func removezero(in []byte) []byte {
	if len(in) == 0 {
		return in
	}
	for i, v := range in {
		if v == 0 {
			in = append(in[:i], in[i+1:]...)
			return removezero(in)
			break
		}
	}
	return in
}

//移除切片中的指定的值
func Removecard(slice []byte, elem byte) []byte {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == elem {
			slice = append(slice[:i], slice[i+1:]...)
			return Removecard(slice, elem)
			break
		}
	}
	return slice
}

//找牌
func Findcard(slice []byte, elem byte) bool {
	if len(slice) == 0 {
		return false
	}
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

//去刻 这个是每种花色的单独去除 碰 暗杠 暗杠
func RemoveSameCard(cards []byte, card byte, num byte) (item *static.TagWeaveItem, err error) {
	if !IsMahjongCard(card) || num > 4 {
		return nil, errors.New(fmt.Sprintf("牌值越界（%d）或者删除数据越界（%d）最多删除4个", card, num))
	}
	index, err := CardToIndex(card)
	if err != nil {
		return nil, err
	}
	if int(index) > len(cards) {
		return nil, errors.New(fmt.Sprintf("要去的牌编号（%d）越界了(%d)", index, len(cards)))
	}
	if cards[index] >= num {
		cards[index] -= num
		item = new(static.TagWeaveItem)
		item.CenterCard = card
		switch num {
		case 4:
			item.WeaveKind = static.WIK_GANG
		case 3:
			item.WeaveKind = static.WIK_PENG
		}
	} else {
		return nil, errors.New(fmt.Sprintf("要去的牌数目（%d）越界了(%d)", num, cards[index]))
	}
	return
}

//蓄杠 回头杠 从碰牌升级 20181225 这个将来扔到判碰杠模块中去
func UpdateWeaveItem(WeaveArray []static.TagWeaveItem, card byte, provideUser uint16) (err error) {
	if len(WeaveArray) == 0 {
		return nil
	}
	for index, v := range WeaveArray {
		if v.CenterCard == card && v.WeaveKind == static.WIK_PENG {
			WeaveArray[index].WeaveKind = static.WIK_GANG
			WeaveArray[index].ProvideUser = provideUser
			return nil
		}
	}
	v, _ := CardToIndex(card)
	return errors.New(fmt.Sprintf("未发现牌（%s）的碰牌数据", G_CardAnother[v]))
}

//20181225 调整移动过来
//单一加
func AddCardSingle(cbCardIndex []byte, card byte) (err error) {
	index, err := CardToIndex(card)
	if err != nil {
		return err
	}
	if cbCardIndex[index] == 4 {
		return errors.New(fmt.Sprintf("玩家手牌中(%s)不能再加数了", G_CardAnother[index]))
	}
	cbCardIndex[index] += 1
	return nil
}

//据说是去重复的
func SliceRemoveDuplicate(a interface{}) (err error, ret []interface{}) {
	if reflect.TypeOf(a).Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("去重对象不是切片，类型为（%T）", a)), ret
	}
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return nil, ret
}

//多加
func AddCardBatch(cbCardIndex []byte, card []byte) (err error) {
	for _, value := range card {
		if err = AddCardSingle(cbCardIndex, value); err != nil {
			return err
		}
	}
	return nil
}

//n次方？
func Powerf2(x int, n int) int {
	if n == 0 {
		return 1
	} else {
		return x * Powerf2(x, n-1)
	}
}

// //单一减
// func subCardSingle(cbCardIndex []byte, card byte, isguis bool) error {
// 	index := CardToIndex(card)
// 	if cbCardIndex[index] == 0 {
// 		v, err := common.CardToIndex(card)
// 		if err != nil {
// 			return err
// 		}
// 		return errors.New(fmt.Sprintf("玩家手牌中(%s)不够减", common.G_CardAnother[v]))
// 	}
// 	if isguis {
// 		cbCardIndex[index] = 0
// 	} else {
// 		cbCardIndex[index] -= 1
// 	}
// 	return nil
// }

// //减
// func subCardBatch(cbCardIndex []byte, card []byte, isguis bool) (err error) {
// 	for _, value := range card {
// 		if err = subCardSingle(cbCardIndex, value, isguis); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
//这个只能检查手牌，但是有的情况是已经把判牌也加进去了
func CheckHandCardsSafe(cbCardIndex []byte) (num int, err error) {
	num = 0
	cardNum := len(cbCardIndex)
	if cardNum == 0 || cardNum > 35 {
		return num, errors.New(fmt.Sprintf("手牌序列数量异常（%d）", cardNum))
	}
	for i := 0; i < cardNum; i++ {
		if cbCardIndex[i] > 4 {
			return num, errors.New(fmt.Sprintf("手牌(%s)数量（%d）超过4章了", G_CardAnother[i], cbCardIndex[i]))
		}
		num = num + int(cbCardIndex[i])
	}
	if num%3 != 1 {
		return num, errors.New(fmt.Sprintf("手牌数量应该是（3n+1）当前牌数为（%d）", num))
	}
	return num, nil
}

// 如果检查的牌已经加到手牌里面，那么checkcards为public.INVALID_BYTE
func CheckHandCardsSafe_ex(cbCardIndex []byte, checkcard byte) (num int, err error) {
	num = 0
	cardNum := len(cbCardIndex)
	if cardNum == 0 || cardNum > 35 {
		return num, errors.New(fmt.Sprintf("手牌序列数量异常（%d）应该是34", cardNum))
	}
	for i := 0; i < cardNum; i++ {
		//if cbCardIndex[i]>4{
		//	return num,errors.New(fmt.Sprintf("手牌(%s)数量（%d）超过4章了",G_CardAnother[i],cbCardIndex[i]))
		//}
		num = num + int(cbCardIndex[i])
	}
	if checkcard != static.INVALID_BYTE {
		num += 1
	}
	checkNum := num % 3
	if checkNum != 2 {
		return num, errors.New(fmt.Sprintf("手牌数量应该是（3n+2）当前牌数为（%d）", num))
	}
	return num, nil
}

func CheckHuCardsSafe(cbCardIndex []byte) (num int, err error) {
	num = 0
	cardNum := len(cbCardIndex)
	if cardNum == 0 || cardNum > 35 {
		return num, errors.New(fmt.Sprintf("手牌序列数量异常（%d）", cardNum))
	}
	for i := 0; i < cardNum; i++ {
		if cbCardIndex[i] > 4 {
			return num, errors.New(fmt.Sprintf("手牌(%s)数量（%d）超过4章了", G_CardAnother[i], cbCardIndex[i]))
		}
		num = num + int(cbCardIndex[i])
	}
	if num%4 != 0 {
		return num, errors.New(fmt.Sprintf("手牌数量应该是（3n+1）当前牌数为（%d）", num))
	}
	return num, nil
}

func Pow(x int, n int) int {
	if x == 0 {
		return 0
	}
	result := calPow(x, n)
	if n < 0 {
		result = 1 / result
	}
	return result
}

func calPow(x int, n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}

	// 向右移动一位
	result := calPow(x, n>>1)
	result *= result

	// 如果n是奇数
	if n&1 == 1 {
		result *= x
	}

	return result
}

func BubbleSort(values []byte) {
	var arrlen int = len(values)
	for i := 0; i < arrlen-1; i++ {
		for j := 0; j < len(values)-1-i; j++ {
			if values[j] <= values[j+1] {
				values[j], values[j+1] = values[j+1], values[j]
			}
		}
	}
}

func RandSlice(slice interface{}) {
	rv := reflect.ValueOf(slice)
	if rv.Type().Kind() != reflect.Slice {
		return
	}

	length := rv.Len()
	if length < 2 {
		return
	}

	swap := reflect.Swapper(slice)
	rand.Seed(time.Now().Unix())
	for i := length - 1; i >= 0; i-- {
		j := rand.Intn(length)
		swap(i, j)
	}
	return
}
