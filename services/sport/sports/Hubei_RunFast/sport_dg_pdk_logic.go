//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  纸牌游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package Hubei_RunFast

//import "fmt"

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"math"
	"math/rand"
	"strings"
)

//游戏常用常量
const (
	MAX_PLAYER   = 4
	MAXHANDCARD  = 16
	MAX_CARDS_15 = 45
	MAX_CARDS_16 = 48
)

//牌型设计
const (
	MAX_POKER_COUNTS      = 54 // 整副牌54张(含小王,大王)
	CARDINDEX_SMALL       = 53 // 小王牌索引
	CARDINDEX_BIG         = 54 // 大王牌索引
	CARDINDEX_BACK        = 55 // 背面牌索引
	CARDINDEX_BACK_HASSKY = 56 // 背面牌索引，包含天牌时
	CARDINDEX_SKY         = 55 // 天牌索引 、有天牌时记得牌背索引要加1 CARDINDEX_BACK+1 = CARDINDEX_BACK_HASSKY
	CARDINDEX_NULL        = 0  // 无效牌索引
)
const (
	//纸牌出牌类型
	TYPE_ERROR           = -1 //错误的类型
	TYPE_NULL            = 0  //没有类型
	TYPE_ONE             = 1  //单张
	TYPE_TWO             = 2  //两张
	TYPE_TWOSTR          = 3  //连对 445566
	TYPE_THREE_TAKE_TWO  = 4  //三张带二 44456
	TYPE_FOUR_TAKE_TWO   = 5  //四张带二 444456
	TYPE_FOUR_TAKE_THREE = 6  //四张带三 4444567
	TYPE_THREESTR        = 7  //三张飞机
	TYPE_FOURSTR         = 8  //四张飞机
	TYPE_ONESTR          = 10 //顺子
	TYPE_BOMB_NOMORL     = 11 //普通炸弹
	TYPE_OTHR            = 12 //其他
)

//扑克花色
const (
	CC_NULL_S    byte = iota // value --> 0  无色
	CC_DIAMOND_S             // value --> 1  方块
	CC_CLUB_S                // value --> 2  梅花
	CC_HEART_S               // value --> 3  红心
	CC_SPADE_S               // value --> 4  黑桃
	CC_TOTAL_S               // value --> 5
)

//扑克点数
const (
	CP_NULL_S byte = iota // value --> 0 牌背
	CP_A_S                // value --> 1 A
	CP_2_S                // value --> 2
	CP_3_S                // value --> 3
	CP_4_S                // value --> 4
	CP_5_S                // value --> 5
	CP_6_S                // value --> 6
	CP_7_S                // value --> 7
	CP_8_S                // value --> 8
	CP_9_S                // value --> 9
	CP_10_S               // value --> 10
	CP_J_S                // value --> 11 J
	CP_Q_S                // value --> 12 Q
	CP_K_S                // value --> 13 K
	CP_BJ_S               // value --> 14 小王
	CP_RJ_S               // value --> 15 大王
	CP_SKY_S              // value --> 16 天牌

	CP_TOTAL_S // value --> 17
)

//游戏默认常量
const (
	INVALID_ONESTR_MIN_LEN = 255 //不支持顺子,设置顺子长度无限长
	INVALID_BOMB_MIN_LEN   = 255 //不支持炸弹,设置炸弹长度无限长
	INVALID_TWOSTR_MIN_LEN = 255 //不支持连对,设置连对长度无限长
)

//////////////////////////////////////////////////////////////////////////

//牌型设计,做牌时从str转换成byte
var m_strDGCardsMessageUP = [54]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09", "0x0A", "0x0B", "0x0C", "0x0D",
	"0x11", "0x12", "0x13", "0x14", "0x15", "0x16", "0x17", "0x18", "0x19", "0x1A", "0x1B", "0x1C", "0x1D",
	"0x21", "0x22", "0x23", "0x24", "0x25", "0x26", "0x27", "0x28", "0x29", "0x2A", "0x2B", "0x2C", "0x2D",
	"0x31", "0x32", "0x33", "0x34", "0x35", "0x36", "0x37", "0x38", "0x39", "0x3A", "0x3B", "0x3C", "0x3D",
	"0x41", "0x42",
}

//牌型设计,做牌时从str转换成byte,做牌小写字母时需要
var m_strDGCardsMessageLW = [54]string{
	"0x01", "0x02", "0x03", "0x04", "0x05", "0x06", "0x07", "0x08", "0x09", "0x0a", "0x0b", "0x0c", "0x0d",
	"0x11", "0x12", "0x13", "0x14", "0x15", "0x16", "0x17", "0x18", "0x19", "0x1a", "0x1b", "0x1c", "0x1d",
	"0x21", "0x22", "0x23", "0x24", "0x25", "0x26", "0x27", "0x28", "0x29", "0x2a", "0x2b", "0x2c", "0x2d",
	"0x31", "0x32", "0x33", "0x34", "0x35", "0x36", "0x37", "0x38", "0x39", "0x3a", "0x3b", "0x3c", "0x3d",
	"0x41", "0x42",
}

//扑克分析
type TagAnalyseResult struct {
	VOne          []static.TPokerGroup //单张牌
	VTwo          []static.TPokerGroup //对子牌
	VThree        []static.TPokerGroup //三张牌
	VFour         []static.TPokerGroup //4张炸弹牌
	VFive         []static.TPokerGroup //
	VSix          []static.TPokerGroup //
	VSeven        []static.TPokerGroup //
	VEight        []static.TPokerGroup // 8张的炸弹牌
	VK105         []static.TPokerGroup //510k，存放的是所有的5、10、k
	VKing         []static.TPokerGroup //一般是打出牌中的所有王
	VMagic        []static.TPokerGroup //一般是打出牌中的所有赖子
	VAllCard      []static.TPokerGroup //所有的3，所有的4...依次入队列3<4<...<2
	VAll3And4Card []static.TPokerGroup //所有的3，所有的4...依次入队列3<4<...<2
}

//////////////////////////////////////////////////////////////////////////

type SportLogicPDK struct {
	//运行时数据
	m_vOne           []static.TPokerGroup           //单张牌
	m_vTwo           []static.TPokerGroup           //对子牌
	m_vThree         []static.TPokerGroup           //三张牌
	m_vFour          []static.TPokerGroup           //4张炸弹牌
	m_vFive          []static.TPokerGroup           //
	m_vSix           []static.TPokerGroup           //
	m_vSeven         []static.TPokerGroup           //
	m_vEight         []static.TPokerGroup           // 8张的炸弹牌
	m_vK105          []static.TPokerGroup           //510k，存放的是所有的5、10、k
	m_vKing          []static.TPokerGroup           //一般是打出牌中的所有王
	m_vAllHandKing   []static.TPokerGroup           //手牌中的所有王
	m_vMagic         []static.TPokerGroup           //一般是打出牌中的所有赖子
	m_vAllHandMagic  []static.TPokerGroup           //手牌中的所有赖子
	m_allCard        []static.TPokerGroup           //所有的3，所有的4...依次入队列3<4<...<2<王
	m_allHandCards   []static.TPokerGroup           //所有的3，所有的4...依次入队列3<4<...<2<王
	m_allMagicPoint  []uint8                        //所有的赖子，初始化时需要设置好
	m_allPlayerCards [MAX_PLAYER][]PlayerPokerGroup //所有的3，所有的4...依次入队列3<4<...<2<王

	m_minOnestrCount uint8 //顺子的最小长度
	m_minBombCount   uint8 //炸弹的最小长度
	m_maxCardCount   uint8 //手牌最大数目
	m_maxPlayerCount uint8 //游戏人数
	m_isBombSplit    bool  //炸弹是否可以拆,true表示不可以拆
	m_isThreeAceBomb bool  //三个A是否是炸弹
	m_isLessTake     bool  //是否可以少带
	m_is4Take3       bool  //是否可以4带3
	m_is4Take2       bool  //是否可以4带2
	m_cardNum        uint8 //手牌数

	m_rule rule2.St_FriendRule
}

//玩家手牌数据
type PlayerPokerGroup struct {
	Point uint8 //牌型值1-15,k105无效
	Count int   //张数
	//	Cards    []uint8               //手牌数据
	Indexes [4]int //对应索引
}

//基础函数，通过花色和point获取牌索引
func (spl *SportLogicPDK) GetCard(byColor byte, byPoint byte) byte {
	//获取小王索引
	if byte(CP_BJ_S) == byPoint {
		return (byte(CC_SPADE_S)-1)*13 + byte(CP_BJ_S)
	} else if byte(CP_RJ_S) == byPoint { //获取大王索引
		return (byte(CC_SPADE_S)-1)*13 + byte(CP_RJ_S)
	} else if byte(CP_A_S) <= byPoint && byPoint <= byte(CP_K_S) { //获取普通牌索引
		if byte(CC_DIAMOND_S) <= byColor && byColor <= byte(CC_SPADE_S) {
			return (byColor-1)*13 + byPoint
		}
	}
	return 0
}

//基础函数，通过牌索引获取花色
func (spl *SportLogicPDK) GetCardColor(byCard byte) byte {
	//处理获取普通牌花色
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return (((byCard - 1) / 13) % 4) + 1
	} else if byCard == CARDINDEX_SMALL { //处理王牌的花色
		return byte(CC_NULL_S) //小王无花色
	} else if byCard == CARDINDEX_BIG {
		return byte(CC_NULL_S) //大王无花色
	}
	return 0
}

//基础函数，通过牌索引获取点数
func (spl *SportLogicPDK) GetCardPoint(byCard byte) byte {
	//处理获取普通牌点数
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return ((byCard - 1) % 13) + 1
	} else if byCard == CARDINDEX_SMALL { //处理王牌的点数
		return byte(CP_BJ_S) //小王
	} else if byCard == CARDINDEX_BIG {
		return byte(CP_RJ_S) //大王
	}
	return 0
}

//初始化赖子列表,byMagicPoint的值约束： A、2...K可以用1、2...13，但大小王只能用53和54，不能用14和15
func (spl *SportLogicPDK) InitMagicPoint(byMagicPoint byte) {
	spl.m_allMagicPoint = []byte{}
	if byMagicPoint >= 1 && byMagicPoint <= CARDINDEX_BIG {
		spl.m_allMagicPoint = append(spl.m_allMagicPoint, byMagicPoint)
	}
}

//设置赖子值,byMagicPoint的值约束： A、2...K可以用1、2...13，但大小王只能用53和54，不能用14和15
func (spl *SportLogicPDK) AddMagicPoint(byMagicPoint byte) {
	if (byMagicPoint >= 1) && (byMagicPoint <= CARDINDEX_BIG) {
		spl.m_allMagicPoint = append(spl.m_allMagicPoint, byMagicPoint)
	}
}

//设置顺子长度值，没有顺子时可以设置顺子最小长度为一个很大的值，比如254
func (spl *SportLogicPDK) SetOnestrCount(byOnestrCount byte) {
	if (byOnestrCount >= 1) && (byOnestrCount <= 255) {
		spl.m_minOnestrCount = byOnestrCount
	}
}

//设置炸弹的张数
func (spl *SportLogicPDK) SetBombCount(byBombCount byte) {
	if (byBombCount >= 3) && (byBombCount <= 255) {
		spl.m_minBombCount = byBombCount
	}
}

//设置手牌数
func (spl *SportLogicPDK) SetCardNum(bCardNum byte) {
	if (bCardNum >= 3) && (bCardNum <= 255) {
		spl.m_cardNum = bCardNum
	} else {
		spl.m_cardNum = static.MAX_CARD
	}
}

//设置手牌的最大张数
func (spl *SportLogicPDK) SetMaxCardCount(byCardNum byte) {
	if (byCardNum >= 3) && (byCardNum <= 255) {
		spl.m_maxCardCount = byCardNum
	} else {
		spl.m_maxCardCount = static.MAX_CARD
	}
}

//设置最大人数
func (spl *SportLogicPDK) SetMaxPlayerCount(byPlayerNum byte) {
	if (byPlayerNum >= 2) && (byPlayerNum <= 10) {
		spl.m_maxPlayerCount = byPlayerNum
	} else {
		spl.m_maxPlayerCount = static.MAX_PLAYER_4P
	}
}

//设置炸弹是否可以拆开
func (spl *SportLogicPDK) SetBombSplit(bBombSplit bool) {
	spl.m_isBombSplit = bBombSplit
}

//设置3个A是否是炸弹
func (spl *SportLogicPDK) SetThreeAceBomb(bThreeAceBomb bool) {
	spl.m_isThreeAceBomb = bThreeAceBomb
}

//设置是否最后一手可以少带
func (spl *SportLogicPDK) SetLessTake(bLessTake bool) {
	spl.m_isLessTake = bLessTake
}

//设置是否可以4带3
func (spl *SportLogicPDK) Set4Take3(b4Take3 bool) {
	spl.m_is4Take3 = b4Take3
}

//设置是否可以4带2
func (spl *SportLogicPDK) Set4Take2(b4Take2 bool) {
	spl.m_is4Take2 = b4Take2
}
func (spl *SportLogicPDK) GetCardLevel(card_id byte) int {
	var level int = 0
	var tmpid int = int(card_id) - 1
	level = tmpid%13 + 1 + 100
	if card_id%13 == 1 {
		level = 114
	}
	if card_id%13 == 2 {
		level = 116
	}
	if card_id == CARDINDEX_SMALL {
		level = 118
	}
	if card_id == CARDINDEX_BIG {
		level = 120
	}
	if card_id > 55 {
		level = 200
	}
	if card_id == 0 {
		level = 0
	}
	return level
}

//得到牌里的分数
func (spl *SportLogicPDK) GetScore(card_list [static.MAX_CARD]byte, cardlen int) int {
	var score int = 0
	for i := 0; i < cardlen && i < int(spl.m_maxCardCount); i++ {
		if spl.GetCardLevel(card_list[i]) == spl.GetCardLevel(5) {
			score += 5
		} else if spl.GetCardLevel(card_list[i]) == spl.GetCardLevel(10) {
			score += 10
		} else if spl.GetCardLevel(card_list[i]) == spl.GetCardLevel(13) {
			score += 10
		}
	}
	return score
}

//牌数量
func (spl *SportLogicPDK) GetOneCardNum(card_list [static.MAX_CARD]byte, cardlen byte, card byte) byte {
	iNum := byte(0)
	if card == 0 || card > CARDINDEX_BIG {
		return 0
	}
	for i := byte(0); i < cardlen && i < static.MAX_CARD && i < spl.m_maxCardCount; i++ {
		if spl.GetCardLevel(card) == spl.GetCardLevel(card_list[i]) {
			iNum++
		}
	}
	return iNum
}

//牌数量
func (spl *SportLogicPDK) GetCardNum(card_list [static.MAX_CARD]byte, cardlen byte) byte {
	iNum := byte(0)
	for i := byte(0); i < cardlen && i < static.MAX_CARD && i < spl.m_maxCardCount; i++ {
		if card_list[i] > 0 && card_list[i] <= CARDINDEX_BIG {
			iNum++
		}
	}
	return iNum
}

func (spl *SportLogicPDK) SortByIndex(card_list [static.MAX_CARD]byte, cardlen int, smalltobig bool) [static.MAX_CARD]byte {
	if smalltobig {
		for i := 0; i < cardlen-1 && i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
			for j := i + 1; j < cardlen && j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
				if card_list[j] > 0 {
					if spl.GetCardLevel(card_list[i]) > spl.GetCardLevel(card_list[j]) {
						temp := card_list[i]
						card_list[i] = card_list[j]
						card_list[j] = temp
					} else {
						if spl.GetCardLevel(card_list[i]) == spl.GetCardLevel(card_list[j]) {
							if card_list[i] > card_list[j] {
								temp := card_list[i]
								card_list[i] = card_list[j]
								card_list[j] = temp
							}
						}
					}
				}
			}
		}
	} else {
		for i := 0; i < cardlen-1 && i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
			for j := i + 1; j < cardlen && j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
				if card_list[j] > 0 {
					if spl.GetCardLevel(card_list[i]) < spl.GetCardLevel(card_list[j]) {
						temp := card_list[i]
						card_list[i] = card_list[j]
						card_list[j] = temp
					} else {
						if spl.GetCardLevel(card_list[i]) == spl.GetCardLevel(card_list[j]) {
							if card_list[i] < card_list[j] {
								temp := card_list[i]
								card_list[i] = card_list[j]
								card_list[j] = temp
							}
						}
					}
				}
			}
		}
	}
	return card_list
}

//当前牌是否是赖子
func (spl *SportLogicPDK) IsMagic(card_id byte) bool {
	for _, v := range spl.m_allMagicPoint {
		if spl.GetCardPoint(card_id) == spl.GetCardPoint(v) {
			return true
		}
	}
	return false
}

//当前牌是否全是赖子
func (spl *SportLogicPDK) IsAllMagic(card_list [static.MAX_CARD]byte, cardlen int) bool {
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
		if card_list[i] > 0 && !spl.IsMagic(card_list[i]) {
			return false
		}
	}
	return true
}

//当前牌是否除赖子外全部相同
func (spl *SportLogicPDK) IsAllEqualExceptMagic(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	if cardlen > static.MAX_CARD || cardlen > int(spl.m_maxCardCount) {
		return 0, false
	}
	k := byte(0)
	for i := cardlen - 1; i >= 0; i-- {
		if spl.IsMagic(card_list[i]) {
			continue
		}
		k = card_list[i]
		break
	}
	if k == 0 {
		return 0, false
	}
	for i := cardlen - 1; i >= 0; i-- {
		if spl.IsMagic(card_list[i]) {
			continue
		}
		if spl.GetCardLevel(k) != spl.GetCardLevel(card_list[i]) {
			return 0, false
		}
	}
	return k, true
}

//获取赖子的数量
func (spl *SportLogicPDK) GetMagicNum(card_list [static.MAX_CARD]byte, cardlen int, fakepoker []static.TFakePoker) (int, []static.TFakePoker) {
	num := 0
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
		if card_list[i] > 0 && spl.IsMagic(card_list[i]) {
			if fakepoker != nil {
				fakepoker[num].Index = card_list[i]
			}
			num++
		}
	}
	return num, fakepoker
}

//当前牌是否是赖子
func (spl *SportLogicPDK) IsKing(card_id byte) bool {
	if spl.GetCardPoint(card_id) == spl.GetCardPoint(CARDINDEX_SMALL) || spl.GetCardPoint(card_id) == spl.GetCardPoint(CARDINDEX_BIG) {
		return true
	}
	return false
}

//当前牌是否全是赖子
func (spl *SportLogicPDK) IsAllKing(card_list [static.MAX_CARD]byte, cardlen int) bool {
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
		if card_list[i] > 0 && !spl.IsKing(card_list[i]) {
			return false
		}
	}
	return true
}

//当前牌是否除赖子外全部相同
func (spl *SportLogicPDK) IsAllEqualExceptKing(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	if cardlen > static.MAX_CARD || cardlen > int(spl.m_maxCardCount) {
		return 0, false
	}
	k := card_list[cardlen-1]
	for i := cardlen - 1; i >= 0; i-- {
		if spl.GetCardLevel(k) != spl.GetCardLevel(card_list[i]) {
			return 0, false
		}
	}
	if k == 0 {
		return 0, false
	}
	return k, true
}

//获取王的数量
func (spl *SportLogicPDK) GetKingNum(card_list [static.MAX_CARD]byte, cardlen int) int {
	num := 0
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
		if card_list[i] > 0 && spl.IsKing(card_list[i]) {
			num++
		}
	}
	return num
}

//判断是不是普通炸弹，如果是，返回true 对王炸、510k炸弹分开来
// 普通炸弹单纯的指的是三张以上相同的牌
func (spl *SportLogicPDK) IsNormalBomb(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	max_bombCnt := 4
	if spl.IsMagic(CARDINDEX_BIG) {
		max_bombCnt = 6 //2个王+4张同样的牌组成的炸弹
	}
	if cardlen > max_bombCnt || cardlen < int(spl.m_minBombCount) { //3到8张同样的牌组成的炸弹
		return 0, false
	}
	p, bBomb := spl.IsAllEqualExceptMagic(card_list, cardlen)
	if bBomb {
		return p, true
	} else {

	}
	return 0, false
}

// 判断是不是四个王
func (spl *SportLogicPDK) IsKingBomb(card_list [static.MAX_CARD]byte, cardlen int) bool {
	if cardlen != 4 { //判断4王的函数
		return false
	}
	bBomb := spl.IsAllKing(card_list, cardlen)
	if bBomb {
		return true
	}
	return false
}

//获取牌型，
func (spl *SportLogicPDK) GetType(card_list [static.MAX_CARD]byte, cardlen int, outtype int, iNextHandCnt int) static.TCardType {
	var re static.TCardType
	re.Len = 0
	re.Card = 0
	re.Color = 0
	re.Cardtype = TYPE_NULL
	re.Count = 0
	re.BombLevel = 0

	card_list = spl.SortByIndex(card_list, cardlen, false)
	re.Len = int(spl.GetCardNum(card_list, byte(cardlen)))
	if re.Len < 1 {
		return re
	}
	re.Card = card_list[0]
	switch re.Len {
	case 0:
		re.Cardtype = TYPE_NULL
		return re
	case 1:
		re.Cardtype = TYPE_ONE
		return re
	case 2:
		if !spl.IsKing(card_list[0]) && !spl.IsKing(card_list[1]) {
			if spl.GetCardLevel(card_list[0]) == spl.GetCardLevel(card_list[1]) {
				re.Cardtype = TYPE_TWO
			} else {
				re.Cardtype = TYPE_ERROR
			}
		}
		return re
	case 3:
		bLogicValue := spl.GetCardLevel(card_list[0])
		i := 1
		for i = 1; i < re.Len; i++ {
			if bLogicValue != spl.GetCardLevel(card_list[i]) {
				break
			}
		}
		if i == re.Len {
			if spl.m_isThreeAceBomb && bLogicValue == spl.GetCardLevel(CP_A_S) {
				//三个A
				re.Cardtype = TYPE_BOMB_NOMORL
				re.Count = 1
				re.BombLevel = 40
			} else if spl.m_isLessTake {
				re.Cardtype = TYPE_THREE_TAKE_TWO
				re.Count = 1
			} else {
				re.Cardtype = TYPE_ERROR
			}
			return re
		}
		re.Cardtype = TYPE_ERROR
		return re
	default: // >= 4张
		if re.Len == 4 {
			card, isBombNormal := spl.IsNormalBomb(card_list, re.Len)
			if isBombNormal {
				re.Cardtype = TYPE_BOMB_NOMORL
				re.Count = 1
				re.BombLevel = re.Len * 10
				re.Card = card
				return re
			}
		}

		if !spl.m_isBombSplit && re.Len == 5 && (outtype == TYPE_NULL || outtype == TYPE_ERROR) {
			bLogicValue := spl.GetCardLevel(card_list[0])
			bSameCount := 1
			for i := 1; i < re.Len; i++ {
				if bLogicValue != spl.GetCardLevel(card_list[i]) {
					bLogicValue = spl.GetCardLevel(card_list[i])
					bSameCount = 1
				} else {
					bSameCount++
				}
				if bSameCount == 4 {
					re.Card = card_list[i]
					re.Cardtype = TYPE_THREE_TAKE_TWO
					re.Count = 1
					return re
				}
			}
		}
		//begin 20191220
		blast := iNextHandCnt == re.Len //允许少带
		if iNextHandCnt == 0 {
			blast = true
		}
		//end 20191220
		tagAnalyseResult := spl.AnalysebCardData(card_list)
		spl.m_allCard, _, _ = spl.GetGroupTypeByPoint(card_list)
		//四牌判断
		if (spl.m_is4Take2 || spl.m_is4Take3) && (outtype == TYPE_FOUR_TAKE_TWO || outtype == TYPE_FOUR_TAKE_THREE || outtype == TYPE_NULL || outtype == TYPE_ERROR) {
			ret := spl.IsFourStr(re.Len, tagAnalyseResult, blast, outtype)
			if ret.Cardtype >= TYPE_ONE {
				return ret
			} else {
				//begin 20191220 3连和4连修改。有时间建议改成托管跟出的方式：先找到所有3连和4连的可能性，然后补牌
				if spl.m_is4Take3 {
					if len(tagAnalyseResult.VAll3And4Card) > 1 && (len(tagAnalyseResult.VAll3And4Card)-1)*7 >= re.Len {
						forI := true
						var TemtagAnalyseResult TagAnalyseResult
						static.HF_DeepCopy(&TemtagAnalyseResult, &tagAnalyseResult)
						//拆除部分3或4张试下
						for forI {
							cnt := len(TemtagAnalyseResult.VAll3And4Card)
							//拆除
							if TemtagAnalyseResult.VAll3And4Card[cnt-1].Count == 4 {
								//炸弹是否可拆
								if !spl.m_isBombSplit {
									for k := 0; k < len(TemtagAnalyseResult.VFour); k++ {
										if TemtagAnalyseResult.VAll3And4Card[cnt-1].Point == TemtagAnalyseResult.VFour[k].Point {
											TemtagAnalyseResult.VFour = append(TemtagAnalyseResult.VFour[:k], TemtagAnalyseResult.VFour[k+1:]...)
											var group static.TPokerGroup
											group.Point = TemtagAnalyseResult.VAll3And4Card[cnt-1].Point
											group.Count = 2
											group.Color = 0
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[0])
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[1])
											TemtagAnalyseResult.VTwo = append(TemtagAnalyseResult.VTwo, group)

											group.Indexes = []uint8{}
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[2])
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[3])
											TemtagAnalyseResult.VTwo = append(TemtagAnalyseResult.VTwo, group)
										}
									}
								}
							}
							TemtagAnalyseResult.VAll3And4Card = append(TemtagAnalyseResult.VAll3And4Card[:cnt-1], TemtagAnalyseResult.VAll3And4Card[cnt:]...)
							ret := spl.IsFourStr(re.Len, TemtagAnalyseResult, blast, outtype)
							if ret.Cardtype >= TYPE_ONE {
								return ret
							}
							if len(TemtagAnalyseResult.VAll3And4Card) < 1 {
								forI = false
							}
							if (cnt-1)*7 < re.Len {
								forI = false
							}
							if !forI {
								break
							}
						}
					}
				}
				if spl.m_is4Take2 {
					if len(tagAnalyseResult.VAll3And4Card) > 1 && (len(tagAnalyseResult.VAll3And4Card)-1)*6 >= re.Len {
						forI := true
						var TemtagAnalyseResult TagAnalyseResult
						static.HF_DeepCopy(&TemtagAnalyseResult, &tagAnalyseResult)
						//拆除部分3或4张试下
						for forI {
							cnt := len(TemtagAnalyseResult.VAll3And4Card)
							//拆除
							if TemtagAnalyseResult.VAll3And4Card[cnt-1].Count == 4 {
								//炸弹是否可拆
								if !spl.m_isBombSplit {
									for k := 0; k < len(TemtagAnalyseResult.VFour); k++ {
										if TemtagAnalyseResult.VAll3And4Card[cnt-1].Point == TemtagAnalyseResult.VFour[k].Point {
											TemtagAnalyseResult.VFour = append(TemtagAnalyseResult.VFour[:k], TemtagAnalyseResult.VFour[k+1:]...)
											var group static.TPokerGroup
											group.Point = TemtagAnalyseResult.VAll3And4Card[cnt-1].Point
											group.Count = 2
											group.Color = 0
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[0])
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[1])
											TemtagAnalyseResult.VTwo = append(TemtagAnalyseResult.VTwo, group)

											group.Count = 1
											group.Indexes = []uint8{}
											group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[2])
											TemtagAnalyseResult.VOne = append(TemtagAnalyseResult.VOne, group)
										}
									}
								}
							}
							TemtagAnalyseResult.VAll3And4Card = append(TemtagAnalyseResult.VAll3And4Card[:cnt-1], TemtagAnalyseResult.VAll3And4Card[cnt:]...)
							ret := spl.IsFourStr(re.Len, TemtagAnalyseResult, blast, outtype)
							if ret.Cardtype >= TYPE_ONE {
								return ret
							}
							if len(TemtagAnalyseResult.VAll3And4Card) < 1 {
								forI = false
							}
							if (cnt-1)*6 < re.Len {
								forI = false
							}
							if !forI {
								break
							}
						}
					}
				}
				//end 20191220
			}
		}

		//三牌判断
		if outtype == TYPE_THREE_TAKE_TWO || outtype == TYPE_NULL || outtype == TYPE_ERROR {
			ret := spl.IsThreeStr(re.Len, tagAnalyseResult, blast)
			if ret.Cardtype >= TYPE_ONE {
				return ret
			} else {
				//begin 20191220 3连和4连修改。有时间建议改成托管跟出的方式：先找到所有3连和4连的可能性，然后补牌
				if len(tagAnalyseResult.VAll3And4Card) > 1 && (len(tagAnalyseResult.VAll3And4Card)-1)*5 >= re.Len {
					forI := true
					var TemtagAnalyseResult TagAnalyseResult
					static.HF_DeepCopy(&TemtagAnalyseResult, &tagAnalyseResult)
					//拆除部分3或4张试下
					for forI {
						cnt := len(TemtagAnalyseResult.VAll3And4Card)
						//拆除
						if TemtagAnalyseResult.VAll3And4Card[cnt-1].Count == 3 {
							for k := 0; k < len(TemtagAnalyseResult.VThree); k++ {
								if TemtagAnalyseResult.VAll3And4Card[cnt-1].Point == TemtagAnalyseResult.VThree[k].Point {
									TemtagAnalyseResult.VThree = append(TemtagAnalyseResult.VThree[:k], TemtagAnalyseResult.VThree[k+1:]...)
									var group static.TPokerGroup
									group.Point = TemtagAnalyseResult.VAll3And4Card[cnt-1].Point
									group.Count = 2
									group.Color = 0
									group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[0])
									group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[1])
									TemtagAnalyseResult.VTwo = append(TemtagAnalyseResult.VTwo, group)

									group.Count = 1
									group.Indexes = []uint8{}
									group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[2])
									TemtagAnalyseResult.VOne = append(TemtagAnalyseResult.VOne, group)
								}
							}
						} else if TemtagAnalyseResult.VAll3And4Card[cnt-1].Count == 4 {
							//炸弹是否可拆
							if !spl.m_isBombSplit {
								for k := 0; k < len(TemtagAnalyseResult.VFour); k++ {
									if TemtagAnalyseResult.VAll3And4Card[cnt-1].Point == TemtagAnalyseResult.VFour[k].Point {
										TemtagAnalyseResult.VFour = append(TemtagAnalyseResult.VFour[:k], TemtagAnalyseResult.VFour[k+1:]...)
										var group static.TPokerGroup
										group.Point = TemtagAnalyseResult.VAll3And4Card[cnt-1].Point
										group.Count = 2
										group.Color = 0
										group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[0])
										group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[1])
										TemtagAnalyseResult.VTwo = append(TemtagAnalyseResult.VTwo, group)

										group.Indexes = []uint8{}
										group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[2])
										group.Indexes = append(group.Indexes, TemtagAnalyseResult.VAll3And4Card[cnt-1].Indexes[3])
										TemtagAnalyseResult.VTwo = append(TemtagAnalyseResult.VTwo, group)
									}
								}
							}
						}
						TemtagAnalyseResult.VAll3And4Card = append(TemtagAnalyseResult.VAll3And4Card[:cnt-1], TemtagAnalyseResult.VAll3And4Card[cnt:]...)
						ret := spl.IsThreeStr(re.Len, TemtagAnalyseResult, blast)
						if ret.Cardtype >= TYPE_ONE {
							return ret
						}
						if len(TemtagAnalyseResult.VAll3And4Card) < 1 {
							forI = false
						}
						if (cnt-1)*5 < re.Len {
							forI = false
						}
						if !forI {
							break
						}
					}
				}
				//end 20191220
			}
		}

		//两张牌判断
		if len(tagAnalyseResult.VTwo) > 0 {
			re.Count = len(tagAnalyseResult.VTwo)
			//连牌判断
			if len(tagAnalyseResult.VFour) > 0 {
				re.Cardtype = TYPE_NULL
				return re
			}
			if !spl.IsStr(re.Len, 2) {
				re.Cardtype = TYPE_NULL
				return re
			}

			if len(tagAnalyseResult.VTwo)*2 == re.Len {
				re.Cardtype = TYPE_TWOSTR
				return re
			}
		}
		//单张判断
		if (len(tagAnalyseResult.VOne) >= int(spl.m_minOnestrCount)) && (len(tagAnalyseResult.VOne) == re.Len) {
			if !spl.IsStr(re.Len, 1) {
				re.Cardtype = TYPE_NULL
				return re
			}
			re.Cardtype = TYPE_ONESTR
			return re
		}
		return re
	}
}

//检查是否是飞机
func (spl *SportLogicPDK) IsThreeStr(cardlen int, tagAnalyseResult TagAnalyseResult, blast bool) static.TCardType {
	var re static.TCardType
	re.Cardtype = TYPE_NULL
	re.Count = len(tagAnalyseResult.VFour) + len(tagAnalyseResult.VThree)
	re.Len = cardlen
	//三牌判断
	if (len(tagAnalyseResult.VThree) > 0) || (!spl.m_isBombSplit && len(tagAnalyseResult.VFour) > 0) {
		//re.Count = len(tagAnalyseResult.VThree)
		//连牌判断
		if spl.m_isBombSplit {
			if len(tagAnalyseResult.VFour) > 0 {
				re.Cardtype = TYPE_NULL
				return re
			}
		}
		if !spl.IsStr3(re.Len, 3, tagAnalyseResult) {
			re.Cardtype = TYPE_NULL
			return re
		}
		//记录能表示最大值的牌点
		if len(tagAnalyseResult.VFour) > 0 {
			re.Card = tagAnalyseResult.VFour[len(tagAnalyseResult.VFour)-1].Point //不比较花色时，可以用牌点表示这个牌
		}
		if len(tagAnalyseResult.VThree) > 0 {
			card := tagAnalyseResult.VThree[len(tagAnalyseResult.VThree)-1].Point //不比较花色时，可以用牌点表示这个牌
			if card > re.Card {
				re.Card = card
			}
		}
		//这个函数里面可以不带满,外面判断是否是最后一手牌
		if len(tagAnalyseResult.VFour)+len(tagAnalyseResult.VTwo)*2+len(tagAnalyseResult.VOne) == re.Count*2 {
			re.Cardtype = TYPE_THREE_TAKE_TWO
			return re
		} else if blast && spl.m_isLessTake && len(tagAnalyseResult.VFour)+len(tagAnalyseResult.VTwo)*2+len(tagAnalyseResult.VOne) < re.Count*2 {
			re.Cardtype = TYPE_THREE_TAKE_TWO
			return re
		}
	}
	return re
}

//
func (spl *SportLogicPDK) IsFourStr(cardlen int, tagAnalyseResult TagAnalyseResult, blast bool, outType int) static.TCardType {
	var re static.TCardType
	re.Cardtype = TYPE_NULL
	re.Count = len(tagAnalyseResult.VFour)
	re.Len = cardlen
	//四牌判断
	if spl.m_is4Take3 || spl.m_is4Take2 {
		if len(tagAnalyseResult.VFour) > 0 {
			re.Count = len(tagAnalyseResult.VFour)
			re.Card = tagAnalyseResult.VFour[re.Count-1].Point //不比较花色时，可以用牌点表示这个牌
			//连牌判断
			if len(tagAnalyseResult.VFour) > 1 {
				//不能有2,从小往大排列的，如果有2肯定在最后一个位置
				if spl.GetCardLevel(tagAnalyseResult.VFour[re.Count-1].Point) == spl.GetCardLevel(CP_2_S) {
					re.Cardtype = TYPE_ERROR
					return re
				}
				for i := 1; i < len(tagAnalyseResult.VFour); i++ {
					if spl.GetCardLevel(tagAnalyseResult.VFour[i].Point) != (spl.GetCardLevel(tagAnalyseResult.VFour[0].Point) + i) {
						re.Cardtype = TYPE_NULL
						return re
					}
				}
			}
			//这个函数里面可以不带满,外面判断是否是最后一手牌
			if spl.m_is4Take3 && (outType == TYPE_FOUR_TAKE_THREE || outType == TYPE_ERROR || outType == TYPE_NULL) {
				if len(tagAnalyseResult.VThree)*3+len(tagAnalyseResult.VTwo)*2+len(tagAnalyseResult.VOne) == re.Count*3 {
					re.Cardtype = TYPE_FOUR_TAKE_THREE
					return re
				} else if blast && spl.m_isLessTake && len(tagAnalyseResult.VThree)*3+len(tagAnalyseResult.VTwo)*2+len(tagAnalyseResult.VOne) < re.Count*3 {
					re.Cardtype = TYPE_FOUR_TAKE_THREE
					return re
				}
			}
			if spl.m_is4Take2 && (outType == TYPE_FOUR_TAKE_TWO || outType == TYPE_ERROR || outType == TYPE_NULL) {
				if len(tagAnalyseResult.VThree)*3+len(tagAnalyseResult.VTwo)*2+len(tagAnalyseResult.VOne) == re.Count*2 {
					re.Cardtype = TYPE_FOUR_TAKE_TWO
					return re
				} else if blast && spl.m_isLessTake && len(tagAnalyseResult.VThree)*3+len(tagAnalyseResult.VTwo)*2+len(tagAnalyseResult.VOne) < re.Count*2 {
					re.Cardtype = TYPE_FOUR_TAKE_TWO
					return re
				}
			}
		}
	}
	return re
}
func (spl *SportLogicPDK) IsStr3(cardlen int, iZhaNum int, tagAnalyseResult TagAnalyseResult) bool {
	var temptype static.FakeType
	if len(tagAnalyseResult.VAll3And4Card) < 0 || iZhaNum < 1 || iZhaNum > 12 {
		temptype.CardType.Cardtype = TYPE_ERROR
		return false
	}

	var templist [static.MAX_CARD]byte

	bLianFlag := true
	iLianNum := int(0)
	iNewcardlen := int(0)
	iMaxIndex := len(tagAnalyseResult.VAll3And4Card)
	for iIndex := 0; iIndex < iMaxIndex && iNewcardlen < cardlen; iIndex++ {
		if iLianNum == 0 && tagAnalyseResult.VAll3And4Card[iIndex].Count < byte(iZhaNum) {
			continue
		}
		iLianNum++
		//构造数据
		for iCardCount := 0; iCardCount < iZhaNum && byte(iCardCount) < tagAnalyseResult.VAll3And4Card[iIndex].Count; iCardCount++ {
			templist[iNewcardlen] = tagAnalyseResult.VAll3And4Card[iIndex].Indexes[iCardCount]
			iNewcardlen++
			temptype.CardType.Card = tagAnalyseResult.VAll3And4Card[iIndex].Point //存放最大点数，暂时赋值
		}
		if tagAnalyseResult.VAll3And4Card[iIndex].Count > byte(iZhaNum) {
			if iZhaNum == 2 || (spl.m_isBombSplit && tagAnalyseResult.VAll3And4Card[iIndex].Count == 4) {
				bLianFlag = false
				break
				//不能组成连对
			}
		} else if tagAnalyseResult.VAll3And4Card[iIndex].Count < byte(iZhaNum) {
			//从这里断开了
			for iIdx := iIndex + 1; iIdx <= iMaxIndex && iNewcardlen < cardlen; iIdx++ {
				if tagAnalyseResult.VAll3And4Card[iIdx].Count >= byte(iZhaNum) {
					bLianFlag = false
					//不能组成连对
				}
			}
			break
		} else {

		}
	}
	if bLianFlag {
		return true
	}
	return false
}

//
func (spl *SportLogicPDK) IsStr(cardlen int, iZhaNum int) bool {
	var temptype static.FakeType
	if len(spl.m_allCard) < 13 || iZhaNum < 1 || iZhaNum > 12 {
		temptype.CardType.Cardtype = TYPE_ERROR
		return false
	}
	if spl.m_allCard[12].Count >= byte(iZhaNum) {
		//有2，不可能组成连对
		temptype.CardType.Cardtype = TYPE_ERROR
		return false
	}
	var templist [static.MAX_CARD]byte

	bLianFlag := true
	iLianNum := int(0)
	iNewcardlen := int(0)
	iMaxIndex := 11
	if cardlen == iZhaNum { //不是连对或飞机或顺子时，可以有2
		iMaxIndex = 12
	}
	for iIndex := 0; iIndex <= iMaxIndex && iNewcardlen < cardlen; iIndex++ {
		if iLianNum == 0 && spl.m_allCard[iIndex].Count < byte(iZhaNum) {
			continue
		}
		iLianNum++
		//构造数据
		for iCardCount := 0; iCardCount < iZhaNum && byte(iCardCount) < spl.m_allCard[iIndex].Count; iCardCount++ {
			templist[iNewcardlen] = spl.m_allCard[iIndex].Indexes[iCardCount]
			iNewcardlen++
			temptype.CardType.Card = spl.m_allCard[iIndex].Point //存放最大点数，暂时赋值
		}
		if spl.m_allCard[iIndex].Count > byte(iZhaNum) {
			if iZhaNum == 2 || (spl.m_isBombSplit && spl.m_allCard[iIndex].Count == 4) {
				bLianFlag = false
				break
				//不能组成连对
			}
		} else if spl.m_allCard[iIndex].Count < byte(iZhaNum) {
			//从这里断开了
			for iIdx := iIndex + 1; iIdx <= iMaxIndex && iNewcardlen < cardlen; iIdx++ {
				if spl.m_allCard[iIdx].Count >= byte(iZhaNum) {
					bLianFlag = false
					//不能组成连对
				}
			}
			break
		} else {

		}
	}
	if bLianFlag {
		return true
	}
	return false
}

//从分组转换成牌列表，delcards表示要排除的牌
func (spl *SportLogicPDK) GetCardsByGroup(delcards static.TPokerGroup) (cardlist []byte) {
	cardlist = []byte{}
	if len(spl.m_vOne) > 0 {
		for i := 0; i < len(spl.m_vOne); i++ {
			delflag := false
			for j := 0; j < len(delcards.Indexes); j++ {
				if spl.m_vOne[i].Indexes[0] == delcards.Indexes[j] {
					delflag = true
				}
			}
			if !delflag {
				cardlist = append(cardlist, spl.m_vOne[i].Indexes[0])
			}
		}
	}
	if len(spl.m_vTwo) > 0 {
		for i := 0; i < len(spl.m_vTwo); i++ {
			for k := 0; k < len(spl.m_vTwo[i].Indexes); k++ {
				delflag := false
				for j := 0; j < len(delcards.Indexes); j++ {
					if spl.m_vTwo[i].Indexes[k] == delcards.Indexes[j] {
						delflag = true
					}
				}
				if !delflag {
					cardlist = append(cardlist, spl.m_vTwo[i].Indexes[k])
				}
			}
		}
	}
	if len(spl.m_vThree) > 0 {
		for i := 0; i < len(spl.m_vThree); i++ {
			for k := 0; k < len(spl.m_vThree[i].Indexes); k++ {
				delflag := false
				for j := 0; j < len(delcards.Indexes); j++ {
					if spl.m_vThree[i].Indexes[k] == delcards.Indexes[j] {
						delflag = true
					}
				}
				if !delflag {
					cardlist = append(cardlist, spl.m_vThree[i].Indexes[k])
				}
			}
		}
	}
	if !spl.m_isBombSplit && len(spl.m_vFour) > 0 {
		for i := 0; i < len(spl.m_vFour); i++ {
			for k := 0; k < len(spl.m_vFour[i].Indexes); k++ {
				delflag := false
				for j := 0; j < len(delcards.Indexes); j++ {
					if spl.m_vFour[i].Indexes[k] == delcards.Indexes[j] {
						delflag = true
					}
				}
				if !delflag {
					cardlist = append(cardlist, spl.m_vFour[i].Indexes[k])
				}
			}
		}
	}
	return cardlist
}
func (spl *SportLogicPDK) GetGroupType(myCard [static.MAX_CARD]byte) {
	spl.m_vOne = []static.TPokerGroup{}
	spl.m_vTwo = []static.TPokerGroup{}
	spl.m_vThree = []static.TPokerGroup{}
	spl.m_vFour = []static.TPokerGroup{}
	spl.m_vFive = []static.TPokerGroup{}
	spl.m_vSix = []static.TPokerGroup{}
	spl.m_vSeven = []static.TPokerGroup{}
	spl.m_vEight = []static.TPokerGroup{}
	spl.m_vKing = []static.TPokerGroup{}
	spl.m_vAllHandKing = []static.TPokerGroup{}
	spl.m_vMagic = []static.TPokerGroup{}
	spl.m_vAllHandMagic = []static.TPokerGroup{}
	spl.m_vK105 = []static.TPokerGroup{}

	var tempPoker [static.MAX_CARD]static.TPoker
	for i := 0; i < len(myCard); i++ {
		tempPoker[i].Set(myCard[i])
	}

	var group static.TPokerGroup
	//首先找王
	for i := 14; i < 15; i++ {
		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == byte(i) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				spl.m_vKing = append(spl.m_vKing, group)
			}
		}
	}
	//找赖子
	for i := 0; i < len(spl.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == spl.GetCardPoint(spl.m_allMagicPoint[i]) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				spl.m_vMagic = append(spl.m_vMagic, group)
			}
		}
	}
	//找其他的
	pokerIndex := [13]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2}
	for i := 0; i < 13; i++ {
		//赖子不重复放，只放在m_vMagic和m_vAllHandMagic中
		bIsMagic := spl.IsMagic(pokerIndex[i])
		if bIsMagic {
			continue
		}

		num := byte(0)
		indexM := [10]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == pokerIndex[i] {
				indexM[num] = tempPoker[j].Index
				num++
				if pokerIndex[i] == 5 || pokerIndex[i] == 10 || pokerIndex[i] == 13 {
					group.Point = pokerIndex[i]
					group.Color = tempPoker[j].Color
					group.Count = 1
					group.Indexes = []uint8{}
					group.Indexes = append(group.Indexes, tempPoker[j].Index)
					spl.m_vK105 = append(spl.m_vK105, group)
				}
			}
		}
		group.Point = pokerIndex[i]
		group.Color = 0 //只针对k105
		group.Count = num
		group.Indexes = []uint8{}
		for k := 0; k < 8; k++ {
			if indexM[k] == 0 {
				continue
			}
			group.Indexes = append(group.Indexes, indexM[k])
		}
		group.Point = pokerIndex[i]
		if num == 8 {
			spl.m_vEight = append(spl.m_vEight, group)
		} else if num == 7 {
			spl.m_vSeven = append(spl.m_vSeven, group)
		} else if num == 6 {
			spl.m_vSix = append(spl.m_vSix, group)
		} else if num == 5 {
			spl.m_vFive = append(spl.m_vFive, group)
		} else if num == 4 {
			spl.m_vFour = append(spl.m_vFour, group)
		} else if num == 3 {
			spl.m_vThree = append(spl.m_vThree, group)
		} else if num == 2 {
			spl.m_vTwo = append(spl.m_vTwo, group)
		} else if num == 1 {
			spl.m_vOne = append(spl.m_vOne, group)
		}
	}
	spl.m_allHandCards, spl.m_vAllHandMagic, spl.m_vAllHandKing = spl.GetGroupTypeByPoint(myCard)
}

func (spl *SportLogicPDK) GetGroupTypeByPoint(myCard [static.MAX_CARD]byte) ([]static.TPokerGroup, []static.TPokerGroup, []static.TPokerGroup) {
	var allCard []static.TPokerGroup
	var vMagic []static.TPokerGroup
	var vKing []static.TPokerGroup

	var tempPoker [static.MAX_CARD]static.TPoker
	for i := 0; i < len(myCard); i++ {
		tempPoker[i].Set(myCard[i])
	}

	//找其他的
	pokerIndex := [24]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	for i := 0; i < 24; i++ {
		var group static.TPokerGroup
		group.Point = pokerIndex[i]
		group.Indexes = []uint8{}
		group.Color = 0
		group.Count = 0
		//赖子不重复放，只放在m_vMagic和m_vAllHandMagic中
		bIsMagic := spl.IsMagic(pokerIndex[i])
		if bIsMagic {
			allCard = append(allCard, group)
			continue
		}
		for j := 0; !bIsMagic && j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == pokerIndex[i] {
				group.Count++
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
			}
		}
		allCard = append(allCard, group)
	}
	//要重新找王
	vKing = []static.TPokerGroup{}
	var group static.TPokerGroup
	group.Point = 0
	group.Indexes = []uint8{}
	group.Color = 0
	group.Count = 0
	for i := 14; i < 15; i++ {
		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == byte(i) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				vKing = append(vKing, group)
			}
		}
	}
	//找赖子
	vMagic = []static.TPokerGroup{}
	for i := 0; i < len(spl.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == spl.GetCardPoint(spl.m_allMagicPoint[i]) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				vMagic = append(vMagic, group)
			}
		}
	}
	return allCard, vMagic, vKing
}

func (spl *SportLogicPDK) AnalysebCardData(myCard [static.MAX_CARD]byte) (tagAnalyseResult TagAnalyseResult) {
	var tempPoker [static.MAX_CARD]static.TPoker
	for i := 0; i < len(myCard); i++ {
		tempPoker[i].Set(myCard[i])
	}

	var group static.TPokerGroup
	//首先找王
	for i := 14; i < 15; i++ {
		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == byte(i) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				tagAnalyseResult.VKing = append(tagAnalyseResult.VKing, group)
			}
		}
	}
	//找赖子
	for i := 0; i < len(spl.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == spl.GetCardPoint(spl.m_allMagicPoint[i]) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				tagAnalyseResult.VMagic = append(tagAnalyseResult.VMagic, group)
			}
		}
	}
	//找其他的
	pokerIndex := [13]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2}
	for i := 0; i < 13; i++ {
		//赖子不重复放，只放在m_vMagic和m_vAllHandMagic中
		bIsMagic := spl.IsMagic(pokerIndex[i])
		if bIsMagic {
			group.Point = pokerIndex[i]
			group.Color = 0
			group.Count = 0
			group.Indexes = []uint8{}
			tagAnalyseResult.VAllCard = append(tagAnalyseResult.VAllCard, group)
			continue
		}

		num := byte(0)
		indexM := [10]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
			if tempPoker[j].Point == pokerIndex[i] {
				indexM[num] = tempPoker[j].Index
				num++
				//if(pokerIndex[i] == 5 || pokerIndex[i] == 10 || pokerIndex[i] == 13) {
				//	group.Point = pokerIndex[i];
				//	group.Color = tempPoker[j].Color;
				//	group.Count = 1;
				//	group.Indexes = []uint8{}
				//	group.Indexes = append(group.Indexes, tempPoker[j].Index)
				//	spl.m_vK105 = append(spl.m_vK105, group)
				//}
			}
		}
		group.Point = pokerIndex[i]
		group.Color = 0 //只针对k105
		group.Count = num
		group.Indexes = []uint8{}
		for k := 0; k < 8; k++ {
			if indexM[k] == 0 {
				continue
			}
			group.Indexes = append(group.Indexes, indexM[k])
		}
		group.Point = pokerIndex[i]
		if num == 8 {
			tagAnalyseResult.VEight = append(tagAnalyseResult.VEight, group)
		} else if num == 7 {
			tagAnalyseResult.VSeven = append(tagAnalyseResult.VSeven, group)
		} else if num == 6 {
			tagAnalyseResult.VSix = append(tagAnalyseResult.VSix, group)
		} else if num == 5 {
			tagAnalyseResult.VFive = append(tagAnalyseResult.VFive, group)
		} else if num == 4 {
			tagAnalyseResult.VFour = append(tagAnalyseResult.VFour, group)
		} else if num == 3 {
			tagAnalyseResult.VThree = append(tagAnalyseResult.VThree, group)
		} else if num == 2 {
			tagAnalyseResult.VTwo = append(tagAnalyseResult.VTwo, group)
		} else if num == 1 {
			tagAnalyseResult.VOne = append(tagAnalyseResult.VOne, group)
		}

		tagAnalyseResult.VAllCard = append(tagAnalyseResult.VAllCard, group)
	}

	//begin 20191220 3连和4连修改。有时间建议改成托管跟出的方式：先找到所有3连和4连的可能性，然后补牌
	//找到所有的连续的3或4连，只有1个2，2不用考虑
	Max_SeqCnt := 0
	SeqCnt := 0
	THCnt := 0
	var VTHCard [13][]static.TPokerGroup //不支持二维都是变长的切片
	VSubTHCard := []static.TPokerGroup{}
	for i := 0; i < len(tagAnalyseResult.VAllCard); i++ {
		if tagAnalyseResult.VAllCard[i].Count >= 3 {
			SeqCnt++
			VSubTHCard = append(VSubTHCard, tagAnalyseResult.VAllCard[i])
			if i+1 == len(tagAnalyseResult.VAllCard) {
				if Max_SeqCnt <= SeqCnt {
					Max_SeqCnt = SeqCnt //记录最大的
				}
				//最后一条直接填充进去
				VTHCard[THCnt] = VSubTHCard
				THCnt++
				VSubTHCard = []static.TPokerGroup{}
			}
		} else {
			if Max_SeqCnt <= SeqCnt {
				Max_SeqCnt = SeqCnt //记录最大的
			}
			//不连续，填充
			SeqCnt = 0
			if len(VSubTHCard) > 0 {
				VTHCard[THCnt] = VSubTHCard
				THCnt++
				VSubTHCard = []static.TPokerGroup{}
			}
		}
	}
	//如果有多个最长的，4张的优先
	FoundFlag := 0 //最优的位置
	MaxCnt := byte(0)
	for i := THCnt - 1; i >= 0; i-- {
		if len(VTHCard[i]) == Max_SeqCnt {
			TempMaxCnt := byte(0)
			for j := 0; j < len(VTHCard[i]); j++ {
				TempMaxCnt += VTHCard[i][j].Count
			}
			if MaxCnt < TempMaxCnt {
				MaxCnt = TempMaxCnt
				FoundFlag = i
			}
		}
	}

	tagAnalyseResult.VAll3And4Card = []static.TPokerGroup{}
	//把不连续的长度不是最大的3连和4连拆了
	for i := THCnt - 1; i >= 0; i-- {
		if FoundFlag == i && len(VTHCard[i]) == Max_SeqCnt {
			for j := 0; j < len(VTHCard[i]); j++ {
				tagAnalyseResult.VAll3And4Card = append(tagAnalyseResult.VAll3And4Card, VTHCard[i][j])
			}
		} else {
			//需要拆了
			for j := 0; j < len(VTHCard[i]); j++ {
				if VTHCard[i][j].Count == 3 {
					for k := 0; k < len(tagAnalyseResult.VThree); k++ {
						if VTHCard[i][j].Point == tagAnalyseResult.VThree[k].Point {
							tagAnalyseResult.VThree = append(tagAnalyseResult.VThree[:k], tagAnalyseResult.VThree[k+1:]...)
							var group static.TPokerGroup
							group.Point = VTHCard[i][j].Point
							group.Count = 2
							group.Color = 0
							group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[0])
							group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[1])
							tagAnalyseResult.VTwo = append(tagAnalyseResult.VTwo, group)

							group.Count = 1
							group.Indexes = []uint8{}
							group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[2])
							tagAnalyseResult.VOne = append(tagAnalyseResult.VOne, group)
						}
					}
				} else if VTHCard[i][j].Count == 4 {
					//炸弹是否可拆
					if !spl.m_isBombSplit {
						for k := 0; k < len(tagAnalyseResult.VFour); k++ {
							if VTHCard[i][j].Point == tagAnalyseResult.VFour[k].Point {
								tagAnalyseResult.VFour = append(tagAnalyseResult.VFour[:k], tagAnalyseResult.VFour[k+1:]...)
								var group static.TPokerGroup
								group.Point = VTHCard[i][j].Point
								group.Count = 2
								group.Color = 0
								group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[0])
								group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[1])
								tagAnalyseResult.VTwo = append(tagAnalyseResult.VTwo, group)

								group.Indexes = []uint8{}
								group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[2])
								group.Indexes = append(group.Indexes, VTHCard[i][j].Indexes[3])
								tagAnalyseResult.VTwo = append(tagAnalyseResult.VTwo, group)
							}
						}
					}
				}
			}
		}
	}
	//end 20191220
	return
}

//返回true表示能出牌，bFirstList表示前一个玩家的牌，这和c++是反过来的
func (spl *SportLogicPDK) CompareCards(bFirstList []byte, bFirstCount byte, bNextList []byte, bNextCount byte, iFirstType int, iNextHandCnt int) (bool, static.TCardType) {
	//由大到小排序
	bTempFirstList := [static.MAX_CARD]byte{}
	for i := byte(0); i < bFirstCount && i < static.MAX_CARD; i++ {
		bTempFirstList[i] = bFirstList[i]
	}
	bTempFirstList = spl.SortByIndex(bTempFirstList, int(bFirstCount), false)

	bTempNextList := [static.MAX_CARD]byte{}
	for i := byte(0); i < bNextCount && i < static.MAX_CARD; i++ {
		bTempNextList[i] = bNextList[i]
	}
	bTempNextList = spl.SortByIndex(bTempNextList, int(bNextCount), false)

	bNextType := spl.GetType(bTempNextList, int(bNextCount), iFirstType, iNextHandCnt)
	bFirstType := spl.GetType(bTempFirstList, int(bFirstCount), iFirstType, iNextHandCnt)
	//类型判断
	if bNextType.Cardtype == TYPE_NULL || bNextType.Cardtype == TYPE_ERROR {
		return false, bNextType
	}
	if bFirstType.Cardtype == TYPE_NULL {
		return true, bNextType
	}
	//炸弹判断
	if (bFirstType.Cardtype == TYPE_BOMB_NOMORL) && (bNextType.Cardtype != TYPE_BOMB_NOMORL) {
		return false, bNextType
	}
	if (bFirstType.Cardtype != TYPE_BOMB_NOMORL) && (bNextType.Cardtype == TYPE_BOMB_NOMORL) {
		return true, bNextType
	}

	//规则判断,同一类型要比较连数是否一致
	if (bFirstType.Cardtype != bNextType.Cardtype) || (bFirstType.Count != bNextType.Count) {
		return false, bNextType
	}
	//规则判断,顺子类型要比较张数是否一致
	if (bFirstType.Cardtype == TYPE_ONESTR) && (bFirstCount != bNextCount) {
		return false, bNextType
	}

	//开始对比
	switch bNextType.Cardtype {
	case TYPE_ONE, TYPE_TWO, TYPE_ONESTR, TYPE_TWOSTR, TYPE_BOMB_NOMORL:
		return spl.GetCardLevel(bFirstType.Card) < spl.GetCardLevel(bNextType.Card), bNextType
	case TYPE_THREE_TAKE_TWO:
		return spl.GetCardLevel(bFirstType.Card) < spl.GetCardLevel(bNextType.Card), bNextType
	case TYPE_FOUR_TAKE_THREE, TYPE_FOUR_TAKE_TWO:
		return spl.GetCardLevel(bFirstType.Card) < spl.GetCardLevel(bNextType.Card), bNextType
	}
	return false, bNextType
}

//获取最小牌
func (spl *SportLogicPDK) GetMinCard(card_list []byte) byte {
	var min byte
	min = CARDINDEX_BACK
	minLevel := 200
	for iIndex := 0; iIndex < len(card_list); iIndex++ {
		iLevel := spl.GetCardLevel(card_list[iIndex])
		if minLevel < iLevel || iLevel == 0 {
			continue
		} else if minLevel == iLevel {
			if min > card_list[iIndex] {
				min = card_list[iIndex]
				minLevel = iLevel
			}
		} else {
			min = card_list[iIndex]
			minLevel = iLevel
		}
	}
	return min
}

//比较两牌大小
func (spl *SportLogicPDK) ISMin(minCard byte, nextCard byte) bool {
	minLevel := spl.GetCardLevel(minCard)
	nextLevel := spl.GetCardLevel(nextCard)
	if minLevel < nextLevel {
		return true
	} else if minLevel == nextLevel {
		if minCard < nextCard {
			return true
		}
	}
	return false
}

//是否有黑桃3
func (spl *SportLogicPDK) ISHaveSpade3(card_list []byte) bool {
	if nil == card_list || len(card_list) == 0 {
		return false
	}
	for iIndex := 0; iIndex < len(card_list); iIndex++ {
		if 42 == card_list[iIndex] {
			return true
		}
	}
	return false
}

//是否有红桃10
func (spl *SportLogicPDK) ISHaveHeart10(card_list []byte) bool {
	if nil == card_list || len(card_list) == 0 {
		return false
	}
	for iIndex := 0; iIndex < len(card_list); iIndex++ {
		if 36 == card_list[iIndex] {
			return true
		}
	}
	return false
}

//计算2连对和3连对、单连、三只、两只的函数。使用StrTool请确保GetGroupType函数被调用过，这样才能保证m_allHandCards 是最新的。
func (spl *SportLogicPDK) StrTool(outpoint byte, iLianNum int, iZhaNum int) (bool, []static.TPokerGroup) {
	var combinelist []static.TPokerGroup
	//添加代码让它适用于顺子
	if (iZhaNum == 1 && (iLianNum < int(spl.m_minOnestrCount) || iLianNum > 13)) || iZhaNum < 1 || iZhaNum > 12 || iLianNum < 1 || iLianNum > 12 {
		return false, nil
	}
	iMagicNum := len(spl.m_vAllHandMagic)
	if iMagicNum < 0 || iMagicNum > 8 {
		return false, nil
	}
	byOutPoint := spl.GetCardPoint(outpoint)
	if byOutPoint != 0 && (int(byOutPoint)-iLianNum+1 < 3 || (iLianNum > 1 && byOutPoint == 1) || (iLianNum <= 1 && byOutPoint == 2)) {
		//他自己已经通天了，别人不可能比它大,连时A最大，非连时(2只或3只)2最大
		return false, nil
	}
	if len(spl.m_allHandCards) < 24 {
		return false, nil
	}

	byPoint := byte(0)
	bHasStr := false

	//从比outpoint - iLianNum + 1大一点的牌开始遍历。跟GetAllGL相比，2不能组成连对这里是<12（从3到A共12个牌型）
	if 1 == byOutPoint {
		byOutPoint = 14
	} //A的point是1，但索引是11，比它大一点索引是12
	iIndex := 0
	if byOutPoint > 0 {
		iIndex = int(byOutPoint) - iLianNum - 1 //byOutPoint - iLianNum + 2 - 3:byOutPoint - iLianNum + 2为point起点，索引要减3
	}
	for ; (iIndex+iLianNum-1 < 12) || (iLianNum == 1 && iIndex == 12); iIndex++ {
		//非连对时要能找到22或222
		if spl.m_allHandCards[iIndex].Count == 0 && (iIndex+iLianNum-1 < 11) {
			//确保遍历了最大的连对，跟GetAllGL相比，2不能组成连对这里是<11
			continue
		}
		//构造数据
		sortRight := 100 //排序权重，出牌优先级,值小的优先出,基础值100
		var fakeking [static.MAX_CARD]static.TFakePoker
		card_list := [static.MAX_CARD]byte{0}
		cardlen := 0
		card_fk_len := 0
		for idx := iIndex; idx < iIndex+iLianNum; idx++ {
			if 4 <= spl.m_allHandCards[idx].Count && iZhaNum < int(spl.m_allHandCards[idx].Count) {
				//炸弹不可拆时却需要用到炸弹的牌
				if spl.m_isBombSplit {
					continue
				} else {
					//拆炸弹需要加权重
					sortRight += 7 //2
				}
			}
			byTmepHaveNum := 0 //手中有这个牌的数量
			for iCardCount := 0; iCardCount < iZhaNum && iCardCount < int(spl.m_allHandCards[idx].Count); iCardCount++ {
				card_list[cardlen] = spl.m_allHandCards[idx].Indexes[iCardCount]
				cardlen++
				byTmepHaveNum++
				if iZhaNum < int(spl.m_allHandCards[idx].Count) {
					sortRight += (int(spl.m_allHandCards[idx].Count) - iZhaNum) //拆牌要加权重
				}
			}
			byPoint = spl.m_allHandCards[idx].Point //直接获取最大的byPoint
			//这些牌将来需要用赖子替换的，肉挨肉需要保存
			if byTmepHaveNum < iZhaNum {
				for i := 0; i < iZhaNum-byTmepHaveNum; i++ {
					fakeking[card_fk_len].Fakeindex = byPoint
					card_fk_len++
				}
			}
		}
		//if (cardlen == 0)//必须至少有一个非赖子牌,但全是赖子且含3首出时可以当3(对3、3个3)出
		//{
		//	if(!(iLianNum == 1 && iIndex == 0 && m_vAllHandMagic.size() > m_vAllHandKing.size()))
		//		continue;
		//}
		//判断需要几个赖子
		if cardlen+iMagicNum >= iLianNum*iZhaNum {
			//可以组成连对
			byNeedMagicNum := iLianNum*iZhaNum - cardlen
			for byMagic := 0; byMagic < byNeedMagicNum; byMagic++ {
				fakeking[byMagic].Index = spl.m_vAllHandMagic[byMagic].Indexes[0] //替换前的值,这个要保证是唯一的
				card_list[cardlen] = spl.m_vAllHandMagic[byMagic].Indexes[0]
				cardlen++
				sortRight += 2 //补牌(要使用赖子)要加权重，使用1个赖子加2个权重
			}
			//是否是连对
			if cardlen == iLianNum*iZhaNum {
				bHasStr = true
				//保存数据
				var tPKstr static.TPokerGroup
				//tPKstr.Indexes = []byte{};//go语音自动初始化了
				for t := 0; t < cardlen; t++ {
					tPKstr.Indexes = append(tPKstr.Indexes, card_list[t])
				}
				tPKstr.Count = byte(cardlen)
				tPKstr.Color = 0
				tPKstr.Point = byPoint
				tPKstr.SortRight = sortRight //排序权重，出牌优先级
				for faker := 0; faker < static.MAX_CARD && faker < int(spl.m_maxCardCount); faker++ {
					tPKstr.Fakeking[faker] = fakeking[faker]
				}
				tPKstr.Cardtype = TYPE_FOUR_TAKE_THREE
				if iZhaNum == 3 {
					tPKstr.Cardtype = TYPE_THREE_TAKE_TWO
				} else if iZhaNum == 2 {
					tPKstr.Cardtype = TYPE_TWOSTR
				} else if iZhaNum == 1 {
					tPKstr.Cardtype = TYPE_ONESTR
				}
				if iLianNum == 1 {
					tPKstr.Cardtype = TYPE_FOUR_TAKE_THREE
					if iZhaNum == 3 {
						tPKstr.Cardtype = TYPE_THREE_TAKE_TWO
					} else if iZhaNum == 2 {
						tPKstr.Cardtype = TYPE_TWO
					} else if iZhaNum == 1 {
						tPKstr.Cardtype = TYPE_ONE //实际上单张不会用这个函数
					}
				}
				combinelist = append(combinelist, tPKstr)
			}
		}
	}

	return bHasStr, combinelist
}

// 得到所有的顺子
func (spl *SportLogicPDK) GetCombineOneStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(spl.m_vAllHandMagic);

	if outlen < int(spl.m_minOnestrCount) || outlen > 12 {
		return
	}
	_, combinelist = spl.StrTool(outpoint, outlen, 1)
	combinelist = spl.SortBeepCardList(combinelist)
	return
}

// 得到所有的两连对
func (spl *SportLogicPDK) GetCombineTwoStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(spl.m_vAllHandMagic);

	if outlen%2 != 0 {
		return
	}
	_, combinelist = spl.StrTool(outpoint, outlen/2, 2)
	combinelist = spl.SortBeepCardList(combinelist)
	return
}

//outlen实际时连数
func (spl *SportLogicPDK) GetCombineThreeStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(spl.m_vAllHandMagic);
	_, combinelist = spl.StrTool(outpoint, outlen, 3)
	//最后可以做一次优化，让没有王的先提示，有王的后提示
	combinelist = spl.SortBeepCardList(combinelist)
	return
} //end GetCombineThreeStr
//outlen实际时连数
func (spl *SportLogicPDK) GetCombineFourStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(spl.m_vAllHandMagic);
	_, combinelist = spl.StrTool(outpoint, outlen, 4)
	//最后可以做一次优化，让没有王的先提示，有王的后提示
	combinelist = spl.SortBeepCardList(combinelist)
	return
} //end GetCombineFourStr
//得到所有的对子
func (spl *SportLogicPDK) GetCombineTwo(outpoint byte) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(spl.m_vAllHandMagic);
	_, combinelist = spl.StrTool(outpoint, 1, 2)
	combinelist = spl.SortBeepCardList(combinelist)
	return
}

//得到所有的三只
func (spl *SportLogicPDK) GetCombineThree(outpoint byte) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(spl.m_vAllHandMagic);
	_, combinelist = spl.StrTool(outpoint, 1, 3)
	combinelist = spl.SortBeepCardList(combinelist)
	return
}

//得到所有大过outpoint的单牌,先直接翻译后面再优化吧
func (spl *SportLogicPDK) GetAllOne(outpoint byte) (combinelist []static.TPokerGroup) {
	defer func() {
		if len(combinelist) > 1 {
			combinelist = spl.SortBeepCardList(combinelist)
		}
	}()

	var group static.TPokerGroup
	group.Cardtype = TYPE_ONE
	for it := 0; it < len(spl.m_vOne); it++ {
		// 过滤掉王提示
		if spl.GetCardLevel(spl.m_vOne[it].Point) == spl.GetCardLevel(53) || spl.GetCardLevel(spl.m_vOne[it].Point) == spl.GetCardLevel(54) {
			continue
		}

		if spl.GetCardLevel(spl.m_vOne[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vOne[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vOne[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}

	bBeepSKing := false
	bBeepBKing := false
	for it := 0; it < len(spl.m_vKing); it++ {
		// 一种花色的王之提示一次就可以// 提示单张小王
		if spl.GetCardLevel(spl.m_vKing[it].Point) > spl.GetCardLevel(outpoint) && spl.GetCardLevel(spl.m_vKing[it].Point) == spl.GetCardLevel(53) && !bBeepSKing {
			group.SortRight = 100 + len(spl.m_vKing) //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vKing[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vKing[it].Indexes[0])
			combinelist = append(combinelist, group)
			bBeepSKing = true
		}
		if spl.GetCardLevel(spl.m_vKing[it].Point) > spl.GetCardLevel(outpoint) && spl.GetCardLevel(spl.m_vKing[it].Point) == spl.GetCardLevel(54) && !bBeepBKing {
			group.SortRight = 100 + len(spl.m_vKing) //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vKing[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vKing[it].Indexes[0])
			combinelist = append(combinelist, group)
			bBeepBKing = true
		}
	}

	for it := 0; it < len(spl.m_vTwo); it++ {
		// 过滤掉对王
		if spl.GetCardLevel(spl.m_vTwo[it].Point) == spl.GetCardLevel(53) || spl.GetCardLevel(spl.m_vTwo[it].Point) == spl.GetCardLevel(54) {
			continue
		}

		if spl.GetCardLevel(spl.m_vTwo[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 1 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vTwo[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vTwo[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}

	for it := 0; it < len(spl.m_vThree); it++ {
		if spl.GetCardLevel(spl.m_vThree[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 2 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vThree[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vThree[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	if spl.m_isBombSplit {
		//不能拆炸弹时需要返回
		return
	}
	for it := 0; it < len(spl.m_vFour); it++ {
		if spl.GetCardLevel(spl.m_vFour[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 3 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vFour[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vFour[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(spl.m_vFive); it++ {
		if spl.GetCardLevel(spl.m_vFive[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 4 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vFive[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vFive[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(spl.m_vSix); it++ {
		if spl.GetCardLevel(spl.m_vSix[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 5 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vSix[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vSix[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(spl.m_vSeven); it++ {
		if spl.GetCardLevel(spl.m_vSeven[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 6 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vSeven[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vSeven[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(spl.m_vEight); it++ {
		if spl.GetCardLevel(spl.m_vEight[it].Point) > spl.GetCardLevel(outpoint) {
			group.SortRight = 100 + 7 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = spl.m_vEight[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, spl.m_vEight[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	return
}

//最后一手牌是否可以自动出
func (spl *SportLogicPDK) GetLastHandBomb() (mybeepout static.TPokerGroup) {
	//没有炸弹时,也不是这个情况：3A是炸弹且正好有3个A
	if len(spl.m_vFour) > 0 {
		mybeepout.Cardtype = TYPE_BOMB_NOMORL
		mybeepout.Count = 4
		mybeepout.Point = spl.m_vFour[len(spl.m_vFour)-1].Point
		mybeepout.Indexes = append(mybeepout.Indexes, spl.m_vFour[len(spl.m_vFour)-1].Indexes[:]...)
		return mybeepout
	}

	if spl.m_isThreeAceBomb && spl.m_allHandCards[11].Count == 3 {
		mybeepout.Cardtype = TYPE_BOMB_NOMORL
		mybeepout.Count = 3
		mybeepout.Point = spl.m_allHandCards[11].Point
		mybeepout.Indexes = append(mybeepout.Indexes, spl.m_allHandCards[11].Indexes[:]...)
		return mybeepout
	}

	return mybeepout
}

//最后一手牌是否可以自动出
func (spl *SportLogicPDK) LastHandCanOut(cardCount byte) bool {
	//没有炸弹时,也不是这个情况：3A是炸弹且正好有3个A
	if len(spl.m_vFour) == 0 && (!spl.m_isThreeAceBomb || spl.m_allHandCards[11].Count != 3) {
		return true
	}
	if len(spl.m_vFour) == 1 && cardCount == 4 {
		return true
	}
	if spl.m_isThreeAceBomb && spl.m_allHandCards[11].Count == 3 && cardCount == 3 {
		return true
	}

	return false
}

//提示出牌
func (spl *SportLogicPDK) BeepCardOut(otherout [static.MAX_CARD]byte, outtype int, bNextOnlyOne bool, iMyHandCnt int) (breturn bool, mybeepout []static.TPokerGroup) {
	defer func() {
		if len(mybeepout) > 1 {
			mybeepout = spl.SortBeepCardList(mybeepout)
		}
	}()

	var group static.TPokerGroup
	group.Color = 0
	len1 := spl.GetCardNum(otherout, spl.m_maxCardCount)
	cardtype1 := spl.GetType(otherout, int(len1), outtype, iMyHandCnt)

	cardtype1.Cardtype = outtype //重新设置下

	outlevel := spl.GetCardLevel(cardtype1.Card)
	if outtype == TYPE_NULL || outtype == TYPE_ERROR {
		return false, nil //不需要提示
		//} else if (outtype == public.TYPE_BOMB_FOUR_KING){
		//	// 四个王
		//	return false,nil;
	} else if outtype == TYPE_FOUR_TAKE_THREE {
		var templist []static.TPokerGroup
		templist = spl.GetCombineFourStr(cardtype1.Card, cardtype1.Count)
		for it := 0; it < len(templist); it++ {
			//补带的牌
			appendCards := []byte{}
			CardsList := spl.GetCardsByGroup(templist[it])
			if len(CardsList) >= 3*cardtype1.Count {
				for i := 0; i < 3*cardtype1.Count; i++ {
					appendCards = append(appendCards, CardsList[i])
				}
			} else if spl.m_isLessTake {
				appendCards = append(appendCards, CardsList[:]...)
			}
			if len(appendCards) == 3*cardtype1.Count || spl.m_isLessTake {
				templist[it].Indexes = append(templist[it].Indexes, appendCards[:]...)
				templist[it].Count = byte(len(templist[it].Indexes))
				mybeepout = append(mybeepout, templist[it])
			} else {
				continue
			}
		}
	} else if outtype == TYPE_THREE_TAKE_TWO {
		var templist []static.TPokerGroup
		templist = spl.GetCombineThreeStr(cardtype1.Card, cardtype1.Count)
		for it := 0; it < len(templist); it++ {
			//补带的牌
			appendCards := []byte{}
			CardsList := spl.GetCardsByGroup(templist[it])
			if len(CardsList) >= 2*cardtype1.Count {
				for i := 0; i < 2*cardtype1.Count; i++ {
					appendCards = append(appendCards, CardsList[i])
				}
			} else if spl.m_isLessTake {
				appendCards = append(appendCards, CardsList[:]...)
			}
			if len(appendCards) == 2*cardtype1.Count || spl.m_isLessTake {
				templist[it].Indexes = append(templist[it].Indexes, appendCards[:]...)
				templist[it].Count = byte(len(templist[it].Indexes))
				mybeepout = append(mybeepout, templist[it])
			} else {
				continue
			}
		}
	} else if outtype == TYPE_TWOSTR {
		var templist []static.TPokerGroup
		templist = spl.GetCombineTwoStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == TYPE_ONESTR {
		var templist []static.TPokerGroup
		templist = spl.GetCombineOneStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == TYPE_TWO {
		var templist []static.TPokerGroup
		templist = spl.GetCombineTwo(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == TYPE_ONE {
		if !bNextOnlyOne {
			var templist []static.TPokerGroup
			templist = spl.GetAllOne(cardtype1.Card)
			for it := 0; it < len(templist); it++ {
				mybeepout = append(mybeepout, templist[it])
			}
		} else {
			//如果炸弹不可拆
			if spl.m_isBombSplit {
				for j := 12; j >= 0 && spl.GetCardLevel(cardtype1.Card) < spl.GetCardLevel(spl.m_allHandCards[j].Point); j-- {
					if spl.m_allHandCards[j].Count >= 4 {
						var tlist static.TPokerGroup
						tlist.Count = spl.m_allHandCards[j].Count
						tlist.SortRight = 106
						tlist.Cardtype = TYPE_BOMB_NOMORL
						tlist.Point = spl.m_allHandCards[j].Point
						tlist.Indexes = append(tlist.Indexes, spl.m_allHandCards[j].Indexes[:]...)
						mybeepout = append(mybeepout, tlist)

						break
					} else if spl.m_allHandCards[j].Count >= 1 && spl.m_allHandCards[j].Count < 4 {
						var tlist static.TPokerGroup
						tlist.Count = 1
						tlist.SortRight = 100 + int(spl.m_allHandCards[j].Count)
						tlist.Cardtype = TYPE_ONE
						tlist.Point = spl.m_allHandCards[j].Point
						tlist.Indexes = append(tlist.Indexes, spl.m_allHandCards[j].Indexes[0])
						mybeepout = append(mybeepout, tlist)
						break
					}
				}
			} else {
				for j := 12; j >= 0 && spl.GetCardLevel(cardtype1.Card) < spl.GetCardLevel(spl.m_allHandCards[j].Point); j-- {
					if spl.m_allHandCards[j].Count >= 1 {
						var tlist static.TPokerGroup
						tlist.Count = 1
						tlist.SortRight = 100 + int(spl.m_allHandCards[j].Count)
						tlist.Cardtype = TYPE_ONE
						tlist.Point = spl.m_allHandCards[j].Point
						tlist.Indexes = append(tlist.Indexes, spl.m_allHandCards[j].Indexes[0])
						mybeepout = append(mybeepout, tlist)
						break
					}
				}
			}
		}
	}

	//找拱拢(炸弹)
	mybombstr := spl.GetAllXiGL()
	for i := 0; i < len(mybombstr); i++ {
		//非拱笼类型的cardtype1.count =0,
		if mybombstr[i].BombLevel > cardtype1.BombLevel || (mybombstr[i].BombLevel == cardtype1.BombLevel && spl.GetCardLevel(mybombstr[i].Point) > outlevel) {
			group.Cardtype = TYPE_BOMB_NOMORL
			group.Indexes = []byte{}
			group.Count = mybombstr[i].Count
			group.Point = mybombstr[i].Point
			group.SortRight = 106 //所有炸弹的排序权重都设置为110
			for j := 0; j < len(mybombstr[i].Indexes); j++ {
				group.Indexes = append(group.Indexes, mybombstr[i].Indexes[j])
			}
			for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
				group.Fakeking[j] = mybombstr[i].Fakeking[j]
			}
			mybeepout = append(mybeepout, group)
		}
	}

	if len(mybeepout) > 0 {
		return true, mybeepout
	} else {
		return false, mybeepout
	}
}

//首出牌提示
func (spl *SportLogicPDK) BeepFirstCardOut(myCard [static.MAX_CARD]byte, bNextOnlyOne bool, isFirstOut bool, iMyHandCnt int) (breturn bool, mybeepout []static.TPokerGroup) {
	var group static.TPokerGroup

	myCardCount := spl.GetCardNum(myCard, spl.m_maxCardCount)
	//由大到小排序
	bTempMyCard := [static.MAX_CARD]byte{}
	for i := byte(0); i < myCardCount && i < static.MAX_CARD; i++ {
		bTempMyCard[i] = myCard[i]
	}
	bTempMyCard = spl.SortByIndex(bTempMyCard, int(myCardCount), false)

	//先判断是不是一手牌，如果是,就一次出完,
	byMyType := spl.GetType(bTempMyCard, int(myCardCount), TYPE_NULL, iMyHandCnt)
	if TYPE_NULL < byMyType.Cardtype {
		group.Cardtype = byMyType.Cardtype
		group.Count = myCardCount
		if spl.LastHandCanOut(myCardCount) {
			for i := byte(0); i < myCardCount; i++ {
				group.Indexes = append(group.Indexes, bTempMyCard[i])
			}
			mybeepout = append(mybeepout, group)
			return true, mybeepout
		}
	}

	for it := 0; it < len(spl.m_allHandCards) && it < 13; it++ {
		if 1 == spl.m_allHandCards[it].Count && !bNextOnlyOne {
			group.Cardtype = TYPE_ONE
			group.Point = spl.m_allHandCards[it].Point
			group.Count = spl.m_allHandCards[it].Count
			for i := byte(0); i < spl.m_allHandCards[it].Count; i++ {
				group.Indexes = append(group.Indexes, spl.m_allHandCards[it].Indexes[i])
			}
			mybeepout = append(mybeepout, group)
			return true, mybeepout
		} else if 2 == spl.m_allHandCards[it].Count {
			group.Cardtype = TYPE_TWO
			group.Point = spl.m_allHandCards[it].Point
			group.Count = spl.m_allHandCards[it].Count
			for i := byte(0); i < spl.m_allHandCards[it].Count; i++ {
				group.Indexes = append(group.Indexes, spl.m_allHandCards[it].Indexes[i])
			}
			mybeepout = append(mybeepout, group)
			return true, mybeepout

		} else if 3 == spl.m_allHandCards[it].Count {
			group.Point = spl.m_allHandCards[it].Point
			if spl.m_allHandCards[it].Point == 1 && spl.m_isThreeAceBomb {
				//三个A是炸弹时
				group.Cardtype = TYPE_BOMB_NOMORL
				group.Count = spl.m_allHandCards[it].Count
				for i := byte(0); i < spl.m_allHandCards[it].Count; i++ {
					group.Indexes = append(group.Indexes, spl.m_allHandCards[it].Indexes[i])
				}
				mybeepout = append(mybeepout, group)
				return true, mybeepout
			}
			if bNextOnlyOne {
				group.Cardtype = TYPE_TWO //3张不知道带哪些牌，拆成对子出，
				group.Count = spl.m_allHandCards[it].Count - 1
				for i := byte(0); i < spl.m_allHandCards[it].Count-1; i++ {
					group.Indexes = append(group.Indexes, spl.m_allHandCards[it].Indexes[i])
				}
				//如果有3，优先出黑桃3
				if spl.m_allHandCards[it].Point == 3 {
					if spl.m_allHandCards[it].Indexes[2] == 42 {
						group.Indexes[1] = 42
					}
				}
			} else {
				group.Cardtype = TYPE_ONE //3张不知道带哪些牌，拆成单张出，
				group.Count = 1
				group.Indexes = append(group.Indexes, spl.m_allHandCards[it].Indexes[0])
				//如果有3，优先出黑桃3
				if spl.m_maxPlayerCount == 3 {
					if spl.m_allHandCards[it].Point == 3 {
						if spl.m_allHandCards[it].Indexes[1] == 42 {
							group.Indexes[0] = 42
						} else if spl.m_allHandCards[it].Indexes[2] == 42 {
							group.Indexes[0] = 42
						}
					}
				} else { //如果是二人场优先出最小牌的情况
					if isFirstOut {
						for i := 1; i < len(spl.m_allHandCards[it].Indexes); i++ {
							itemp := spl.m_allHandCards[it].Indexes[i]
							if itemp < group.Indexes[0] {
								group.Indexes[0] = itemp
							}
						}
					}
				}
			}
			mybeepout = append(mybeepout, group)
			return true, mybeepout

		} else if 4 == spl.m_allHandCards[it].Count {
			group.Cardtype = TYPE_BOMB_NOMORL
			group.Point = spl.m_allHandCards[it].Point
			group.Count = 4
			for i := byte(0); i < 4; i++ {
				group.Indexes = append(group.Indexes, spl.m_allHandCards[it].Indexes[i])
			}
			mybeepout = append(mybeepout, group)
			return true, mybeepout
		}
	}
	//走到这里，说了是下家单张报警了，自己也只有单张了,出最大的单张
	if 0 < len(spl.m_vOne) {
		group.Cardtype = TYPE_ONE
		group.Point = spl.m_vOne[len(spl.m_vOne)-1].Point
		group.Count = 1
		group.Indexes = append(group.Indexes, spl.m_vOne[len(spl.m_vOne)-1].Indexes[0])
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	return false, nil
}
func (spl *SportLogicPDK) GetMaxSingleCard(myCard [static.MAX_CARD]byte) byte {
	num := spl.GetCardNum(myCard, MAXHANDCARD)
	if num == 0 {
		return 0
	}
	maxCard := byte(0)
	spl.GetGroupType(myCard)
	for i := 12; i >= 0; i-- {
		if spl.m_allHandCards[i].Count > 0 {
			if !spl.m_isBombSplit && spl.m_allHandCards[i].Count == 4 {
				maxCard = spl.m_allHandCards[i].Indexes[0]
				break
			}
			if spl.m_allHandCards[i].Count <= 4 {
				maxCard = spl.m_allHandCards[i].Indexes[0]
				break
			}
		}
	}
	return maxCard
}

//获取所有同张的炸弹，比如软4个5 ，或2个王
func (spl *SportLogicPDK) GetAllXiGL() (mybombstr []static.TBombStr) {
	iMagicNum := len(spl.m_vAllHandMagic) //癞子的数量
	if iMagicNum > 8 || iMagicNum < 0 {
		return
	}
	if len(spl.m_allHandCards) < 24 {
		return
	}

	for i := 0; i <= iMagicNum && i < 8; i++ {
		//全由赖子组成的炸弹后面处理、赖子数最多8个
		for j := 0; j < 13; j++ {
			itemsize := len(spl.m_allHandCards[j].Indexes)
			if itemsize == 0 {
				continue
			} //数目为0不能组成喜
			var item static.TBombStr
			item.Indexes = []byte{}
			if i+itemsize >= int(spl.m_minBombCount) {
				//限定为大于等于3张
				var fakeking [static.MAX_CARD]static.TFakePoker
				for k := 0; k < itemsize && k < 8; k++ {
					//同一种牌最多8个，2副牌
					item.Indexes = append(item.Indexes, spl.m_allHandCards[j].Indexes[k])
				}
				for k := 0; k < i; k++ {
					//if (i != 8)//全由赖子组成的炸弹没有特殊性，如果有特殊性可以参考肉挨肉处理
					{
						fakeking[k].Fakeindex = spl.GetCardPoint(spl.m_allHandCards[j].Indexes[0]) //替换的值
						fakeking[k].Index = spl.m_vAllHandMagic[k].Indexes[0]                      //替换前的值,这个要保证是唯一的
					}
					item.Indexes = append(item.Indexes, spl.m_vAllHandMagic[k].Indexes[0])
				}
				item.BombLevel = (i + itemsize) * 10 //3张的炸弹nBombCount为30,4张的炸弹为40,5张炸弹为50
				item.Point = spl.m_allHandCards[j].Point
				item.MaxCount = uint8(i + itemsize)
				item.Count = uint8(i + itemsize)
				for it := 0; it < static.MAX_CARD && it < int(spl.m_maxCardCount); it++ {
					item.Fakeking[it] = fakeking[it]
				}
				mybombstr = append(mybombstr, item)
			}
		}
	}
	//如果3个A是炸弹
	if spl.m_isThreeAceBomb {
		if spl.m_allHandCards[11].Count == 3 {
			var item static.TBombStr
			item.Indexes = []byte{}
			for k := 0; k < int(spl.m_allHandCards[11].Count) && k < 8; k++ {
				//同一种牌最多8个，2副牌
				item.Indexes = append(item.Indexes, spl.m_allHandCards[11].Indexes[k])
			}
			item.BombLevel = 40 //3张的炸弹nBombCount为30,4张的炸弹为40,5张炸弹为50
			item.Point = spl.m_allHandCards[11].Point
			item.MaxCount = 3
			item.Count = 3
			mybombstr = append(mybombstr, item)
		}
	}
	return
}

//出牌排序
func (spl *SportLogicPDK) SortBeepCardList(combinelist []static.TPokerGroup) []static.TPokerGroup {
	//结果低于2个不用排序
	if len(combinelist) < 2 {
		return combinelist
	}

	for i := 0; i < len(combinelist)-1; i++ {
		for j := i + 1; j < len(combinelist); j++ {
			if combinelist[i].SortRight > combinelist[j].SortRight {
				tempItem := combinelist[i]
				combinelist[i] = combinelist[j]
				combinelist[j] = tempItem
			}
		}
	}

	return combinelist
}

//20181214 每个玩法设定自己的限时操作时间
func (spl *SportLogicPDK) Setlimitetime(limitetimeOp bool) {
	//if limitetimeOp {
	//	spl.Rule.limitetimeOP = GameTime_Nine
	//} else {
	//	spl.Rule.limitetimeOP = 0
	//}
}

func (spl *SportLogicPDK) CreateCards() (byte, [static.ALL_CARD]byte) {
	cbCardDataTemp := [static.ALL_CARD]byte{}
	_maxCount := byte(static.ALL_CARD)
	//初始化所有牌点数
	//48张去掉黑桃A和方块2、梅花2、红桃2、大小王
	for i := 0; i < MAX_POKER_COUNTS-6; i++ {
		//先让牌为2-k
		cbCardDataTemp[i] = byte(i) + 2 + byte(i)/12
	}
	//在将红桃和方块、梅花2换成A。
	cbCardDataTemp[0] = 1   //方块A
	cbCardDataTemp[12] = 14 //梅花A
	cbCardDataTemp[24] = 27 //红桃A
	return _maxCount, cbCardDataTemp
}

func (spl *SportLogicPDK) CreateCards15() (byte, [static.ALL_CARD]byte) {
	cbCardDataTemp := [static.ALL_CARD]byte{}
	_maxCount := byte(static.ALL_CARD)
	//初始化所有牌点数
	//48张去掉黑桃K、方块A、梅花A、黑桃A、方块2、梅花2、红桃2、大小王，
	for i := 0; i < MAX_POKER_COUNTS-10; i++ {
		//先让牌为3-k
		cbCardDataTemp[i] = byte(i) + 3 + byte(i)/11*2
	}
	//黑桃K换成红桃A
	cbCardDataTemp[MAX_POKER_COUNTS-11] = 27
	//增加黑桃2
	cbCardDataTemp[MAX_POKER_COUNTS-10] = 41
	return _maxCount, cbCardDataTemp
}

////混乱扑克 并发牌
//func (spl *SportLogicPDK) RandCardData(byAllCards [static.ALL_CARD]byte, wantGoodCard [MAX_PLAYER][static.MAX_CARD]byte) ([MAX_PLAYER]byte, [MAX_PLAYER][static.MAX_CARD]byte, [static.ALL_CARD]byte) {
//	cbCardDataTemp := [MAX_PLAYER][static.MAX_CARD]byte{}
//	_maxCount := [MAX_PLAYER]byte{}
//
//	wantCardCount := 0
//	wantCards := make([]byte, 0)
//	var playerWantCardCount [MAX_PLAYER]int
//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//		for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
//			if wantCard := wantGoodCard[j][i]; spl.IsValidCard(wantCard) {
//				wantCardCount++
//				wantCards = append(wantCards, wantCard)
//				playerWantCardCount[j]++
//			}
//		}
//	}
//	// 从牌库拿掉这些想要的牌
//	var realCards []byte
//	for _, card := range byAllCards {
//		if spl.IsValidCard(card) {
//			var find bool
//			for _, wantCard := range wantCards {
//				if wantCard == card {
//					find = true
//					break
//				}
//			}
//			if !find {
//				realCards = append(realCards, card)
//			}
//		}
//	}
//	//_randTmp := time.Now().Unix()
//	//rand.Seed(_randTmp)
//	//rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
//	//洗牌
//	leftCardCount := len(realCards)
//	for i := 0; i < 2000; i++ {
//		rand_num := rand.Intn(1000)
//		m := rand_num % leftCardCount
//		rand_num_2 := rand.Intn(1000)
//		n := rand_num_2 % leftCardCount
//		//zz := byAllCards[m]
//		//byAllCards[m] = byAllCards[n]
//		//byAllCards[n] = zz
//
//		realCards[m], realCards[n] = realCards[n], realCards[m]
//	}
//
//	//清零
//	for i := 0; i < MAX_PLAYER; i++ {
//		for j := 0; j < static.MAX_CARD; j++ {
//			cbCardDataTemp[i][j] = 0
//		}
//	}
//
//	var start int
//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//		for i := 0; i < playerWantCardCount[j]; i++ {
//			if wantCard := wantGoodCard[j][i]; spl.IsValidCard(wantCard) {
//				cbCardDataTemp[j][i] = wantCard
//				_maxCount[j]++
//			}
//		}
//		other := int(spl.m_maxCardCount) - playerWantCardCount[j] // 还差这么多张
//		otherCards := realCards[start : start+other]
//		start += other
//		for i := 0; i < len(otherCards); i++ {
//			cbCardDataTemp[j][_maxCount[j]] = otherCards[i]
//			_maxCount[j]++
//		}
//	}
//
//	////发牌
//	//for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
//	//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//	//		if sendCard := byAllCards[int(static.MAX_PLAYER_3P)*i+j]; spl.IsValidCard(sendCard) {
//	//			cbCardDataTemp[j][i] = sendCard
//	//			_maxCount[j]++
//	//		}
//	//	}
//	//}
//
//	//遍历玩家的手牌，从其他玩家上手换掉
//
//	return _maxCount, cbCardDataTemp, byAllCards
//}
//
////混乱扑克 并发牌
//func (spl *SportLogicPDK) RandCardData15(byAllCards [static.ALL_CARD]byte, wantGoodCard [MAX_PLAYER][static.MAX_CARD]byte) ([MAX_PLAYER]byte, [MAX_PLAYER][static.MAX_CARD]byte, [static.ALL_CARD]byte) {
//	//cbCardDataTemp := [MAX_PLAYER][static.MAX_CARD]byte{}
//	//_maxCount := [MAX_PLAYER]byte{}
//	//
//	////_randTmp := time.Now().Unix()
//	////rand.Seed(_randTmp)
//	////rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
//	////洗牌
//	//wantCardCount := 0
//	//for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//	//	for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
//	//		if wantCard := wantGoodCard[j][i]; spl.IsValidCard(wantCard) {
//	//			wantCardCount++
//	//		}
//	//	}
//	//}
//	//
//	//leftCardCount := MAX_POKER_COUNTS - 9
//	//for i := 0; i < 2000; i++ {
//	//	rand_num := rand.Intn(1000)
//	//	m := rand_num % leftCardCount
//	//	rand_num_2 := rand.Intn(1000)
//	//	n := rand_num_2 % leftCardCount
//	//	zz := byAllCards[m]
//	//	byAllCards[m] = byAllCards[n]
//	//	byAllCards[n] = zz
//	//}
//	//
//	////清零
//	//for i := 0; i < MAX_PLAYER; i++ {
//	//	for j := 0; j < static.MAX_CARD; j++ {
//	//		cbCardDataTemp[i][j] = 0
//	//	}
//	//}
//	//
//	////发牌
//	//for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
//	//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//	//		if wantCard := wantGoodCard[j][i]; spl.IsValidCard(wantCard) {
//	//			cbCardDataTemp[j][i] = wantCard
//	//			_maxCount[j]++
//	//		} else if sendCard := byAllCards[int(static.MAX_PLAYER_3P)*i+j]; spl.IsValidCard(sendCard) {
//	//			cbCardDataTemp[j][i] = sendCard
//	//			_maxCount[j]++
//	//		}
//	//	}
//	//}
//	//
//	//return _maxCount, cbCardDataTemp, byAllCards
//	cbCardDataTemp := [MAX_PLAYER][static.MAX_CARD]byte{}
//	_maxCount := [MAX_PLAYER]byte{}
//
//	wantCardCount := 0
//	wantCards := make([]byte, 0)
//	var playerWantCardCount [MAX_PLAYER]int
//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//		for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
//			if wantCard := wantGoodCard[j][i]; spl.IsValidCard(wantCard) {
//				wantCardCount++
//				wantCards = append(wantCards, wantCard)
//				playerWantCardCount[j]++
//			}
//		}
//	}
//	// 从牌库拿掉这些想要的牌
//	var realCards []byte
//	for _, card := range byAllCards {
//		if spl.IsValidCard(card) {
//			var find bool
//			for _, wantCard := range wantCards {
//				if wantCard == card {
//					find = true
//					break
//				}
//			}
//			if !find {
//				realCards = append(realCards, card)
//			}
//		}
//	}
//	//_randTmp := time.Now().Unix()
//	//rand.Seed(_randTmp)
//	//rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
//	//洗牌
//	leftCardCount := len(realCards)
//	for i := 0; i < 2000; i++ {
//		rand_num := rand.Intn(1000)
//		m := rand_num % leftCardCount
//		rand_num_2 := rand.Intn(1000)
//		n := rand_num_2 % leftCardCount
//		//zz := byAllCards[m]
//		//byAllCards[m] = byAllCards[n]
//		//byAllCards[n] = zz
//
//		realCards[m], realCards[n] = realCards[n], realCards[m]
//	}
//
//	//清零
//	for i := 0; i < MAX_PLAYER; i++ {
//		for j := 0; j < static.MAX_CARD; j++ {
//			cbCardDataTemp[i][j] = 0
//		}
//	}
//
//	var start int
//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//		for i := 0; i < playerWantCardCount[j]; i++ {
//			if wantCard := wantGoodCard[j][i]; spl.IsValidCard(wantCard) {
//				cbCardDataTemp[j][i] = wantCard
//				_maxCount[j]++
//			}
//		}
//		other := int(spl.m_maxCardCount) - playerWantCardCount[j] // 还差这么多张
//		otherCards := realCards[start : start+other]
//		start += other
//		for i := 0; i < len(otherCards); i++ {
//			cbCardDataTemp[j][_maxCount[j]] = otherCards[i]
//			_maxCount[j]++
//		}
//	}
//
//	////发牌
//	//for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
//	//	for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
//	//		if sendCard := byAllCards[int(static.MAX_PLAYER_3P)*i+j]; spl.IsValidCard(sendCard) {
//	//			cbCardDataTemp[j][i] = sendCard
//	//			_maxCount[j]++
//	//		}
//	//	}
//	//}
//
//	//遍历玩家的手牌，从其他玩家上手换掉
//
//	return _maxCount, cbCardDataTemp, byAllCards
//}


//混乱扑克 并发牌
func (spl *SportLogicPDK) RandCardData(byAllCards [static.ALL_CARD]byte) ([MAX_PLAYER]byte, [MAX_PLAYER][static.MAX_CARD]byte, [static.ALL_CARD]byte) {
	cbCardDataTemp := [MAX_PLAYER][static.MAX_CARD]byte{}
	_maxCount := [MAX_PLAYER]byte{}

	//_randTmp := time.Now().Unix()
	//rand.Seed(_randTmp)
	//rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	//洗牌
	for i := 0; i < 2000; i++ {
		rand_num := rand.Intn(1000)
		m := rand_num % (MAX_POKER_COUNTS - 6)
		rand_num_2 := rand.Intn(1000)
		n := rand_num_2 % (MAX_POKER_COUNTS - 6)
		zz := byAllCards[m]
		byAllCards[m] = byAllCards[n]
		byAllCards[n] = zz
	}

	//清零
	for i := 0; i < MAX_PLAYER; i++ {
		for j := 0; j < static.MAX_CARD; j++ {
			cbCardDataTemp[i][j] = 0
		}
	}

	//发牌
	for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
		for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
			if spl.IsValidCard(byAllCards[int(static.MAX_PLAYER_3P)*i+j]) {
				cbCardDataTemp[j][i] = byAllCards[int(static.MAX_PLAYER_3P)*i+j]
				_maxCount[j]++
			}
		}
	}

	return _maxCount, cbCardDataTemp, byAllCards
}

//混乱扑克 并发牌
func (spl *SportLogicPDK) RandCardData15(byAllCards [static.ALL_CARD]byte) ([MAX_PLAYER]byte, [MAX_PLAYER][static.MAX_CARD]byte, [static.ALL_CARD]byte) {
	cbCardDataTemp := [MAX_PLAYER][static.MAX_CARD]byte{}
	_maxCount := [MAX_PLAYER]byte{}

	//_randTmp := time.Now().Unix()
	//rand.Seed(_randTmp)
	//rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	//洗牌
	for i := 0; i < 2000; i++ {
		rand_num := rand.Intn(1000)
		m := rand_num % (MAX_POKER_COUNTS - 9)
		rand_num_2 := rand.Intn(1000)
		n := rand_num_2 % (MAX_POKER_COUNTS - 9)
		zz := byAllCards[m]
		byAllCards[m] = byAllCards[n]
		byAllCards[n] = zz
	}

	//清零
	for i := 0; i < MAX_PLAYER; i++ {
		for j := 0; j < static.MAX_CARD; j++ {
			cbCardDataTemp[i][j] = 0
		}
	}

	//发牌
	for i := 0; i < static.MAX_CARD && i < int(spl.m_maxCardCount); i++ {
		for j := 0; j < MAX_PLAYER && j < int(static.MAX_PLAYER_3P); j++ {
			if spl.IsValidCard(byAllCards[int(static.MAX_PLAYER_3P)*i+j]) {
				cbCardDataTemp[j][i] = byAllCards[int(static.MAX_PLAYER_3P)*i+j]
				_maxCount[j]++
			}
		}
	}

	return _maxCount, cbCardDataTemp, byAllCards
}


//炸弹检测
func (spl *SportLogicPDK) HasBomb(playercards [MAX_PLAYER][static.MAX_CARD]byte, playercount int) bool {
	for i := 0; i < playercount; i++ {
		cardsNum := [14]int{}
		for j := 0; j < static.MAX_CARD; j++ {
			if !spl.IsValidCard(playercards[i][j]) {
				continue
			}
			point := spl.GetCardPoint(playercards[i][j])
			cardsNum[point]++
			if cardsNum[point] == 4 { //有普通炸弹
				return true
			}
			//有三A炸弹
			if spl.m_isThreeAceBomb && point == CP_A_S && cardsNum[point] == 3 {
				return true
			}
		}
	}
	return false
}

func (spl *SportLogicPDK) GetPlayerGroupTypeByPoint(playercards [MAX_PLAYER][static.MAX_CARD]byte) {
	var allCard [MAX_PLAYER][]PlayerPokerGroup

	for i := 0; i < static.MAX_PLAYER_3P; i++ {
		var tempPoker [static.MAX_CARD]static.TPoker
		for j := 0; j < len(playercards[i]); j++ {
			tempPoker[j].Set(playercards[i][j])
		}

		//找其他的
		pokerIndex := [13]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2}
		for k := 0; k < 13; k++ {
			var group PlayerPokerGroup
			group.Point = pokerIndex[k]
			//group.Cards = []uint8{}
			group.Indexes = [4]int{}
			group.Count = 0

			for j := 0; j < static.MAX_CARD && j < int(spl.m_maxCardCount); j++ {
				if tempPoker[j].Point == pokerIndex[k] {
					group.Indexes[group.Count] = j
					group.Count++
					//group.Cards = append(group.Cards, tempPoker[j].Index)
				}
			}
			allCard[i] = append(allCard[i], group)
		}
	}

	spl.m_allPlayerCards = allCard
}

//是否构成飞机
func (spl *SportLogicPDK) ThreeStrCheck(index byte, player int) bool {
	if index < 0 || index > 11 {
		return false
	}

	if len(spl.m_allPlayerCards[player]) != 13 {
		return false
	}

	//牌值数小于3张
	if spl.m_allPlayerCards[player][index].Count < 2 {
		return false
	}
	//与前一位的索引牌构成飞机
	if index > 0 && spl.m_allPlayerCards[player][index-1].Count >= 3 {
		return true
	}
	//与后一位的索引牌构成飞机
	if index < 11 && spl.m_allPlayerCards[player][index+1].Count >= 3 {
		return true
	}
	return false
}

func (spl *SportLogicPDK) PokerBombCheck(notCheckSeat []uint16) bool {
	for i := 0; i < int(spl.m_maxPlayerCount); i++ {
		if len(spl.m_allPlayerCards[i]) != 13 {
			continue
		}
		var not bool
		for _, notCheck := range notCheckSeat {
			if uint16(i) == notCheck {
				not = true
				break
			}
		}
		if not {
			continue
		}
		//检测每人手上是否有炸弹
		for j := 0; j < 13; j++ {
			if spl.m_allPlayerCards[i][j].Count >= 4 {
				return true
			}
		}
		//有三A炸弹
		if spl.m_isThreeAceBomb && spl.m_allPlayerCards[i][11].Count == 3 {
			return true
		}
	}
	return false
}

//连牌检测
func (spl *SportLogicPDK) PokerStrCheck(strtype int, notCheckSeat []uint16) bool {
	if strtype != 3 && strtype != 2 {
		return false
	}
	//连牌最短长度
	mincnt := 2
	//最少三连对
	if 2 == strtype {
		mincnt = 3
	}
	for i := 0; i < int(spl.m_maxPlayerCount); i++ {
		if len(spl.m_allPlayerCards[i]) != 13 {
			continue
		}
		var not bool
		for _, notCheck := range notCheckSeat {
			if uint16(i) == notCheck {
				not = true
				break
			}
		}
		if not {
			continue
		}
		//检测每人手上是否有飞机/连对
		strCnt := 0
		for j := 0; j < 12; j++ {
			if spl.m_allPlayerCards[i][j].Count >= strtype {
				strCnt++
			} else {
				strCnt = 0
			}
			if strCnt >= mincnt {
				return true
			}
		}
	}
	return false
}

// PokerSingleCount 得到玩家手上的单排数量
func (spl *SportLogicPDK) PokerSingleCount(chair uint16) ([]int, int) {
	if chair >= MAX_PLAYER {
		return nil, 0
	}
	handCard := spl.m_allPlayerCards[chair]
	pokerIndex := [13]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2}
	maxIndex := len(pokerIndex)
	if len(handCard) != maxIndex {
		return nil, 0
	}
	cards := make([]int, 0, maxIndex)
	for i := 0; i < maxIndex; i++ {
		cards = append(cards, handCard[i].Count)
	}
	return cards, CardsSingleCount(cards)
}

func CardsSingleCount(cards []int) int {
	cardsTemp := cards[:len(cards)-1]
	single := singleCount(getNextIndex(0, cardsTemp), cardsTemp)
	if num2 := cards[len(cards)-1]; num2 == 1 {
		single++
	} else if num2 == 3 {
		single--
	}
	return single
}

func getNextIndex(curIdx int, cards []int) int {
	for ; curIdx < len(cards); curIdx++ {
		if cards[curIdx] != 0 {
			return curIdx
		}
	}
	return len(cards)
}

func singleCount(idx int, cards []int) int {
	if idx >= len(cards) {
		return 0
	}
	min := math.MaxInt32
	switch cards[idx] {
	case 1:
		count := 0
		i := idx
		for ; i < len(cards); i++ {
			if cards[i] == 0 {
				break
			}
			cards[i]--
			count++
			if count >= 5 {
				// get next index
				min = static.HF_MinInt(min, singleCount(getNextIndex(idx+1, cards), cards))
			}
		}
		//rollback
		for j := idx; j < i; j++ {
			cards[j]++
		}
		cards[idx]--
		min = static.HF_MinInt(min, singleCount(getNextIndex(idx+1, cards), cards)+1)
		cards[idx]++
		return min
	case 2:
		cards[idx]--
		min = singleCount(getNextIndex(idx, cards), cards) + 1
		cards[idx]++
		return static.HF_MinInt(singleCount(getNextIndex(idx+1, cards), cards), min)
	case 3:
		cards[idx] = 1
		min = singleCount(getNextIndex(idx, cards), cards)
		cards[idx] = 2
		min = static.HF_MinInt(singleCount(getNextIndex(idx, cards), cards)+1, min)
		cards[idx] = 3
		return static.HF_MinInt(singleCount(getNextIndex(idx+1, cards), cards)-1, min)
	case 4:
		cards[idx] = 1
		min = singleCount(getNextIndex(idx, cards), cards) - 1
		cards[idx] = 2
		min = static.HF_MinInt(singleCount(getNextIndex(idx, cards), cards), min)
		cards[idx] = 3
		min = static.HF_MinInt(singleCount(getNextIndex(idx, cards), cards)+1, min)
		cards[idx] = 4
		return static.HF_MinInt(singleCount(getNextIndex(idx+1, cards), cards), min)
	default:
		return math.MaxInt32
	}
}

//获取索引
func (spl *SportLogicPDK) GetPointIndex(point byte) byte {
	if point == 1 {
		return 11
	}
	if point == 2 {
		return 12
	}
	return point - 3
}

//是否构成连对
func (spl *SportLogicPDK) TwoStrCheck(index byte, player int) bool {
	if index < 0 || index > 11 {
		return false
	}

	if len(spl.m_allPlayerCards[player]) != 13 {
		return false
	}

	if spl.m_allPlayerCards[player][index].Count < 1 {
		return false
	}

	//连牌数
	Strcnt := 1
	//与前位的连牌数
	for i := int(index) - 1; i >= 0; i-- {
		if spl.m_allPlayerCards[player][i].Count >= 2 {
			Strcnt++
		} else {
			break
		}
	}

	//与后位的连牌数
	for i := int(index) + 1; i <= 11; i++ {
		if spl.m_allPlayerCards[player][i].Count >= 2 {
			Strcnt++
		} else {
			break
		}
	}

	//是否构成3连对
	if Strcnt >= 3 {
		return true
	}

	return false
}

//是否构成炸弹
func (spl *SportLogicPDK) BombCheck(index byte, player int) bool {
	if index < 0 || index > 12 {
		return false
	}
	if len(spl.m_allPlayerCards[player]) != 13 {
		return false
	}
	if spl.m_allPlayerCards[player][index].Count == 3 || (spl.m_isThreeAceBomb && spl.m_allPlayerCards[player][index].Point == CP_A_S && spl.m_allPlayerCards[player][index].Count == 2) {
		return true
	}
	return false
}

//拆掉炸弹
func (spl *SportLogicPDK) BombSplit(playercards [MAX_PLAYER][static.MAX_CARD]byte) [MAX_PLAYER][static.MAX_CARD]byte {
	for i := 0; i < int(spl.m_maxPlayerCount); i++ {
		if len(spl.m_allPlayerCards[i]) != 13 {
			spl.GetPlayerGroupTypeByPoint(playercards)
			break
		}
	}
	for i := 0; i < int(spl.m_maxPlayerCount); i++ {
		splitIndex := 0
		//是否存在炸弹
		for j := byte(0); j < 13; j++ {
			cardsnum := spl.m_allPlayerCards[i][j].Count
			cardpoint := spl.m_allPlayerCards[i][j].Point
			if cardsnum == 4 || (spl.m_isThreeAceBomb && cardpoint == CP_A_S && cardsnum == 3) {

				//换0到n-1张牌
				splitnum := rand.Intn(100) % (cardsnum - 1)

				//拆掉炸弹与其他玩家手牌替换
				for k := 0; k <= splitnum; k++ {
					bombindex := spl.m_allPlayerCards[i][j].Indexes[cardsnum-1]
					var cIndex, pIndex int
					var changePoint, cPoint, cardtemp1 byte
					//是否成功交换
					changeFlag := false
					//找到其他玩家手牌中符合条件的替换牌
					for k := 0; k < 32; k++ {
						//cIndex, pIndex =spl.RandCard(i,k)
						pIndex, cIndex = spl.GetSplitIndex(i, splitIndex)
						//换人
						if k%2 == 0 {
							splitIndex += int(spl.m_cardNum)
						} else {
							splitIndex++
						}
						if !spl.IsValidCard(playercards[pIndex][cIndex]) {
							continue
						}
						cardtemp1 = playercards[pIndex][cIndex]
						changePoint = spl.GetCardPoint(cardtemp1)
						cPoint = spl.GetPointIndex(changePoint)

						spl.m_allPlayerCards[i][j].Count--
						spl.m_allPlayerCards[pIndex][cPoint].Count--
						//若替换该牌是否仍可构成炸弹
						changeFlag = !(spl.BombCheck(cPoint, i) || spl.TwoStrCheck(cPoint, i) || spl.ThreeStrCheck(cPoint, i) || spl.ThreeStrCheck(j, pIndex) || spl.TwoStrCheck(j, pIndex))
						spl.m_allPlayerCards[i][j].Count++
						spl.m_allPlayerCards[pIndex][cPoint].Count++
						if changeFlag {
							break
						}
					}
					//若可交换手牌
					if changeFlag && cardpoint != changePoint {
						if spl.m_allPlayerCards[pIndex][cPoint].Count == 0 || spl.m_allPlayerCards[i][j].Count == 0 {
							break
						}
						//交换手牌
						playercards[i][bombindex], playercards[pIndex][cIndex] = playercards[pIndex][cIndex], playercards[i][bombindex]
						spl.m_allPlayerCards[i][j].Count--
						spl.m_allPlayerCards[pIndex][j].Indexes[spl.m_allPlayerCards[pIndex][j].Count] = spl.m_allPlayerCards[i][j].Indexes[spl.m_allPlayerCards[i][j].Count]
						spl.m_allPlayerCards[i][j].Indexes[spl.m_allPlayerCards[i][j].Count] = 0
						spl.m_allPlayerCards[pIndex][j].Count++
						cardsnum--

						spl.m_allPlayerCards[pIndex][cPoint].Count--
						spl.m_allPlayerCards[i][cPoint].Indexes[spl.m_allPlayerCards[i][cPoint].Count] = spl.m_allPlayerCards[pIndex][cPoint].Indexes[spl.m_allPlayerCards[pIndex][cPoint].Count]
						spl.m_allPlayerCards[pIndex][cPoint].Indexes[spl.m_allPlayerCards[pIndex][cPoint].Count] = 0
						spl.m_allPlayerCards[i][cPoint].Count++
					}
				}
			}
		}
	}
	return playercards
}

//顺序拆牌
func (spl *SportLogicPDK) GetSplitIndex(splitseat int, temp int) (int, int) {
	tempPlayer := (temp/(int(spl.m_cardNum)))%2 + splitseat + 1
	tempPlayer = tempPlayer % 3
	tempIndex := temp % (int(spl.m_cardNum))

	return tempPlayer, tempIndex
}

//牌有效判断
func (spl *SportLogicPDK) IsValidCard(cbCardData byte) bool {
	//校验
	if cbCardData == 0 || cbCardData > CARDINDEX_BIG {
		return false
	}
	return true
}
func (spl *SportLogicPDK) GetStringByCard(carddata byte) string {
	strColor := ""
	strPoint := ""

	color := spl.GetCardColor(carddata)
	point := spl.GetCardPoint(carddata)

	if point == CP_RJ_S {
		return "rj"
	} else if point == CP_BJ_S {
		return "bj"
	} else if point == CP_SKY_S {
		return "sj"
	} else {
		if color == CC_SPADE_S {
			strColor = "s"
		} else if color == CC_HEART_S {
			strColor = "h"
		} else if color == CC_CLUB_S {
			strColor = "c"
		} else if color == CC_DIAMOND_S {
			strColor = "d"
		}

		if point == CP_A_S {
			strPoint = "A"
		} else if point == CP_2_S {
			strPoint = "2"
		} else if point == CP_3_S {
			strPoint = "3"
		} else if point == CP_4_S {
			strPoint = "4"
		} else if point == CP_5_S {
			strPoint = "5"
		} else if point == CP_6_S {
			strPoint = "6"
		} else if point == CP_7_S {
			strPoint = "7"
		} else if point == CP_8_S {
			strPoint = "8"
		} else if point == CP_9_S {
			strPoint = "9"
		} else if point == CP_10_S {
			strPoint = "10"
		} else if point == CP_J_S {
			strPoint = "J"
		} else if point == CP_Q_S {
			strPoint = "Q"
		} else if point == CP_K_S {
			strPoint = "K"
		}

		return strColor + strPoint
	}
}

func (spl *SportLogicPDK) GetWriteHandReplayRecordString(replayRecord meta2.DG_Replay_Record) string {
	handCardStr := ""
	for i := 0; i < MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < len(replayRecord.R_HandCards[i]); j++ {
			handCardStr += fmt.Sprintf("%s,", spl.GetStringByCard(byte(replayRecord.R_HandCards[i][j])))
		}
	}

	//写入分数
	handCardStr += "S:"
	for i := 0; i < MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d,", replayRecord.R_Score[i])
	}

	return handCardStr
}

func (spl *SportLogicPDK) GetWriteOutReplayRecordString(replayRecord meta2.DG_Replay_Record, score [4]int) string {
	upd := false
	endMsgUpdateScore := [MAX_PLAYER]float64{}
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
		case meta2.DG_REPLAY_OPT_HOUPAI:
			ourCardStr += fmt.Sprintf("|H%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_QIANG:
			ourCardStr += fmt.Sprintf("|Q%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_END_HOUPAI:
			ourCardStr += fmt.Sprintf("|E%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_END_GAME:
			ourCardStr += fmt.Sprintf("|G%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_DIS_GAME:
			ourCardStr += fmt.Sprintf("|J%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_OUTCARD:
			//if len(record.R_Value) > 0 {
			//	ourCardStr += fmt.Sprintf("|C%d:", record.R_ChairId)
			//} else {
			//	ourCardStr += fmt.Sprintf("|C%d", record.R_ChairId)
			//}
			ourCardStr += fmt.Sprintf("|C%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_TURN_OVER:
			if len(record.R_Opt_Ext) > 0 {
				ourCardStr += fmt.Sprintf("|T%d:", record.R_ChairId)
			} else {
				ourCardStr += fmt.Sprintf("|T%d", record.R_ChairId)
			}
			break
		case meta2.DG_REPLAY_OPT_TUOGUAN:
			ourCardStr += fmt.Sprintf("|D%d:", record.R_ChairId)
			break
		default:
			break
		}

		if len(record.R_Value) > 0 {
			for i := 0; i < len(record.R_Value); i++ {
				ourCardStr += fmt.Sprintf("%s", spl.GetStringByCard(byte(record.R_Value[i])))
			}
		}
		if len(record.R_Opt_Ext) > 0 {
			for i := 0; i < len(record.R_Opt_Ext); i++ {
				if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_HOUPAI {
					ourCardStr += fmt.Sprintf(",H%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_QIANG {
					ourCardStr += fmt.Sprintf(",Q%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_TURNSCORE {
					ourCardStr += fmt.Sprintf(",S%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_GETSCORE {
					ourCardStr += fmt.Sprintf(",G%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_MINGJI {
					ourCardStr += fmt.Sprintf(",J")
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_ENDQIANG {
					ourCardStr += fmt.Sprintf(",B%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_CARDTYPE {
					ourCardStr += fmt.Sprintf(",T%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_TUOGUAN {
					ourCardStr += fmt.Sprintf(",D%d", record.R_Opt_Ext[i].Ext_value)
				}
			}
		}
	}

	//增加当局小结算分
	for i, s := range score {
		ourCardStr += fmt.Sprintf(",%d:F%d", i, s)
	}

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
