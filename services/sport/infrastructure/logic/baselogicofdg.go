//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  纸牌游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package logic

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"strings"
)

//////////////////////////////////////////////////////////////////////////
const (
//GAME_GENRE = (public.GAME_GENRE_GOLD | public.GAME_GENRE_MATCH) //游戏类型
)

//牌型设计
const (
	MAX_POKER_COUNTS      = 54 // 整副牌54张(含小王,大王)
	CARDINDEX_SMALL       = 53 // 小王牌索引
	CARDINDEX_BIG         = 54 // 大王牌索引
	CARDINDEX_BACK        = 55 // 背面牌索引
	CARDINDEX_BACK_HASSKY = 56 // 背面牌索引
	CARDINDEX_SKY         = 55 // 天牌索引 、有天牌时记得牌背索引要加1 CARDINDEX_BACK+1 = CARDINDEX_BACK_HASSKY
	CARDINDEX_NULL        = 0  // 无效牌索引
)

//扑克花色
const (
	CC_NULL_S    int = iota // value --> 0  无色
	CC_DIAMOND_S            // value --> 1  方块
	CC_CLUB_S               // value --> 2  梅花
	CC_HEART_S              // value --> 3  红心
	CC_SPADE_S              // value --> 4  黑桃
	CC_TOTAL_S              // value --> 5
)

//扑克点数
const (
	CP_NULL_S int = iota // value --> 0 牌背
	CP_A_S               // value --> 1 A
	CP_2_S               // value --> 2
	CP_3_S               // value --> 3
	CP_4_S               // value --> 4
	CP_5_S               // value --> 5
	CP_6_S               // value --> 6
	CP_7_S               // value --> 7
	CP_8_S               // value --> 8
	CP_9_S               // value --> 9
	CP_10_S              // value --> 10
	CP_J_S               // value --> 11 J
	CP_Q_S               // value --> 12 Q
	CP_K_S               // value --> 13 K
	CP_BJ_S              // value --> 14 小王
	CP_RJ_S              // value --> 15 大王
	CP_SKY_S             // value --> 16 天牌

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

//////////////////////////////////////////////////////////////////////////
//基础函数，通过牌索引获取点数
func GetCardPoint(byCard byte) byte {
	//处理获取普通牌点数
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return ((byCard - 1) % 13) + 1
	} else if byCard == CARDINDEX_SMALL {
		return byte(CP_BJ_S)
	} else if byCard == CARDINDEX_BIG {
		return byte(CP_RJ_S)
	} else if byCard == CARDINDEX_SKY {
		return byte(CP_SKY_S)
	}
	return 0
}

type BaseLogicDG struct {
	//运行时数据
	m_vOne          []static.TPokerGroup //单张牌
	m_vTwo          []static.TPokerGroup //对子牌
	m_vThree        []static.TPokerGroup //三张牌
	m_vFour         []static.TPokerGroup //4张炸弹牌
	m_vFive         []static.TPokerGroup //
	m_vSix          []static.TPokerGroup //
	m_vSeven        []static.TPokerGroup //
	m_vEight        []static.TPokerGroup // 8张的炸弹牌
	m_vK105         []static.TPokerGroup //510k，存放的是所有的5、10、k
	m_vKing         []static.TPokerGroup //一般是打出牌中的所有王
	m_vAllHandKing  []static.TPokerGroup //手牌中的所有王
	m_vMagic        []static.TPokerGroup //一般是打出牌中的所有赖子
	m_vAllHandMagic []static.TPokerGroup //手牌中的所有赖子
	m_allCard       []static.TPokerGroup //所有的3，所有的4...依次入队列3<4<...<2<王
	m_allHandCards  []static.TPokerGroup //所有的3，所有的4...依次入队列3<4<...<2<王
	m_allMagicPoint []uint8              //所有的赖子，初始化时需要设置好

	MinOnestrCount uint8 //顺子的最小长度
	MinBombCount   uint8 //炸弹的最小长度
	MaxCardCount   uint8 //手牌最大数目
	MaxPlayerCount uint8 //游戏人数

	Rule rule2.St_FriendRule
}

//基础函数，通过花色和point获取牌索引
func (self *BaseLogicDG) GetCard(byColor byte, byPoint byte) byte {
	if byte(CP_BJ_S) == byPoint { //获取小王索引
		return (byte(CC_SPADE_S)-1)*13 + byte(CP_BJ_S)
	} else if byte(CP_RJ_S) == byPoint { //获取大王索引
		return (byte(CC_SPADE_S)-1)*13 + byte(CP_RJ_S)
	} else if byte(CP_SKY_S) == byPoint { //获取天牌（花牌）索引
		return (byte(CC_SPADE_S)-1)*13 + byte(CP_SKY_S)
	} else if byte(CP_A_S) <= byPoint && byPoint <= byte(CP_K_S) { //获取普通牌索引
		if byte(CC_DIAMOND_S) <= byColor && byColor <= byte(CC_SPADE_S) {
			return (byColor-1)*13 + byPoint
		}
	}
	return 0
}

//基础函数，通过牌索引获取花色
func (self *BaseLogicDG) GetCardColor(byCard byte) byte {
	//处理获取普通牌花色
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return (((byCard - 1) / 13) % 4) + 1
	} else if byCard == CARDINDEX_SMALL {
		return byte(CC_NULL_S) //小王无花色
	} else if byCard == CARDINDEX_BIG {
		return byte(CC_NULL_S) //大王无花色
	} else if byCard == CARDINDEX_SKY {
		return byte(CC_NULL_S) //天牌（花牌）无花色
	}
	return 0
}

//基础函数，通过牌索引获取点数
func (self *BaseLogicDG) GetCardPoint(byCard byte) byte {
	//处理获取普通牌点数
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return ((byCard - 1) % 13) + 1
	} else if byCard == CARDINDEX_SMALL {
		return byte(CP_BJ_S)
	} else if byCard == CARDINDEX_BIG {
		return byte(CP_RJ_S)
	} else if byCard == CARDINDEX_SKY {
		return byte(CP_SKY_S)
	}
	return 0
}

//初始化赖子列表,byMagicPoint的值约束： A、2...K可以用1、2...13，但大小王只能用53和54，不能用14和15
func (self *BaseLogicDG) InitMagicPoint(byMagicPoint byte) {
	self.m_allMagicPoint = []byte{}
	if byMagicPoint >= 1 && byMagicPoint <= CARDINDEX_SKY {
		self.m_allMagicPoint = append(self.m_allMagicPoint, byMagicPoint)
	}
}

//设置赖子值,byMagicPoint的值约束： A、2...K可以用1、2...13，但大小王只能用53和54，不能用14和15
func (self *BaseLogicDG) AddMagicPoint(byMagicPoint byte) {
	if (byMagicPoint >= 1) && (byMagicPoint <= CARDINDEX_SKY) {
		self.m_allMagicPoint = append(self.m_allMagicPoint, byMagicPoint)
	}
}

//设置顺子长度值，没有顺子时可以设置顺子最小长度为一个很大的值，比如254
func (self *BaseLogicDG) SetOnestrCount(byOnestrCount byte) {
	if (byOnestrCount >= 1) && (byOnestrCount <= 255) {
		self.MinOnestrCount = byOnestrCount
	}
}

//设置炸弹的张数
func (self *BaseLogicDG) SetBombCount(byBombCount byte) {
	if (byBombCount >= 3) && (byBombCount <= 255) {
		self.MinBombCount = byBombCount
	}
}

//设置手牌的最大张数
func (self *BaseLogicDG) SetMaxCardCount(byCardNum byte) {
	if (byCardNum >= 3) && (byCardNum <= 255) {
		self.MaxCardCount = byCardNum
	} else {
		self.MaxCardCount = static.MAX_CARD
	}
}

//设置最大人数
func (self *BaseLogicDG) SetMaxPlayerCount(byPlayerNum byte) {
	if (byPlayerNum >= 2) && (byPlayerNum <= 10) {
		self.MaxPlayerCount = byPlayerNum
	} else {
		self.MaxPlayerCount = static.MAX_PLAYER_4P
	}
}

func (self *BaseLogicDG) GetCardLevel(card_id byte) int {
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
func (self *BaseLogicDG) GetScore(card_list [static.MAX_CARD]byte, cardlen int) int {
	var score int = 0
	for i := 0; i < cardlen && i < int(self.MaxCardCount); i++ {
		if self.GetCardLevel(card_list[i]) == self.GetCardLevel(5) {
			score += 5
		} else if self.GetCardLevel(card_list[i]) == self.GetCardLevel(10) {
			score += 10
		} else if self.GetCardLevel(card_list[i]) == self.GetCardLevel(13) {
			score += 10
		}
	}
	return score
}

//得到8喜炸弹个数
func (self *BaseLogicDG) Get8XibombNum(card_list [static.MAX_CARD]byte) int {
	var nCount int = 0
	for i := 1; i <= 13; i++ {
		num := 0
		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
			if self.GetCardLevel(card_list[j]) == self.GetCardLevel(byte(i)) {
				num++
				//if (num == 8){break;}
			}
		}
		if num == 8 {
			nCount++
		}
	}
	return nCount
}

//得到7喜炸弹个数，注意：不要把8喜也算到7喜里面了
func (self *BaseLogicDG) Get7XibombNum(card_list [static.MAX_CARD]byte) int {
	var nCount int = 0
	for i := 1; i <= 13; i++ {
		num := 0
		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
			if self.GetCardLevel(card_list[j]) == self.GetCardLevel(byte(i)) {
				num++
				//if (num == 7){break;}
			}
		}
		if num == 7 {
			nCount++
		}
	}
	return nCount
}

//是否得到了4王
func (self *BaseLogicDG) HasKingBomb(card_list [static.MAX_CARD]byte) int {
	var nCount int = 0

	if self.GetKingNum(card_list, static.MAX_CARD) >= 4 {
		nCount++
	}

	return nCount
}

//获得了同样花色510k的个数
func (self *BaseLogicDG) GetSame510kNum(card_list [static.MAX_CARD]byte) int {

	//构造m_vK105等变量的数据

	self.GetGroupType(card_list)
	//开始计算
	nCount := 0

	if len(self.m_vK105) == 0 {
		return 0
	}
	bySameK := [4]byte{0} //4种花色，是否有同一花色有2个k
	bySame10 := [4]byte{0}
	bySame5 := [4]byte{0}

	for v := 0; v < len(self.m_vK105); v++ {
		if self.GetCardLevel(self.m_vK105[v].Point) == self.GetCardLevel(13) {
			if self.m_vK105[v].Color <= 4 && self.m_vK105[v].Color >= 1 {
				bySameK[self.m_vK105[v].Color-1]++
			}
		} else if self.GetCardLevel(self.m_vK105[v].Point) == self.GetCardLevel(10) {
			if self.m_vK105[v].Color <= 4 && self.m_vK105[v].Color >= 1 {
				bySame10[self.m_vK105[v].Color-1]++
			}
		} else if self.GetCardLevel(self.m_vK105[v].Point) == self.GetCardLevel(5) {
			if self.m_vK105[v].Color <= 4 && self.m_vK105[v].Color >= 1 {
				bySame5[self.m_vK105[v].Color-1]++
			}
		}
	}
	for byColorIndex := 0; byColorIndex < 4; byColorIndex++ {
		//拥有了2套相同的花色的510k
		if bySameK[byColorIndex] == 2 && bySame10[byColorIndex] == 2 && bySame5[byColorIndex] == 2 {
			nCount++
		}
	}
	return nCount
}

//牌数量
func (self *BaseLogicDG) GetCardNum(card_list [static.MAX_CARD]byte, cardlen byte) byte {
	iNum := byte(0)
	for i := byte(0); i < cardlen && i < static.MAX_CARD && i < self.MaxCardCount; i++ {
		if card_list[i] > 0 && card_list[i] <= CARDINDEX_SKY {
			iNum++
		}
	}
	return iNum
}

func (self *BaseLogicDG) SortByIndex(card_list [static.MAX_CARD]byte, cardlen int, smalltobig bool) [static.MAX_CARD]byte {
	if smalltobig {
		for i := 0; i < cardlen-1 && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
			for j := i + 1; j < cardlen && j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
				if card_list[j] > 0 {
					if self.GetCardLevel(card_list[i]) > self.GetCardLevel(card_list[j]) {
						temp := card_list[i]
						card_list[i] = card_list[j]
						card_list[j] = temp
					} else {
						if self.GetCardLevel(card_list[i]) == self.GetCardLevel(card_list[j]) {
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
		for i := 0; i < cardlen-1 && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
			for j := i + 1; j < cardlen && j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
				if card_list[j] > 0 {
					if self.GetCardLevel(card_list[i]) < self.GetCardLevel(card_list[j]) {
						temp := card_list[i]
						card_list[i] = card_list[j]
						card_list[j] = temp
					} else {
						if self.GetCardLevel(card_list[i]) == self.GetCardLevel(card_list[j]) {
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
func (self *BaseLogicDG) IsMagic(card_id byte) bool {
	for _, v := range self.m_allMagicPoint {
		if self.GetCardPoint(card_id) == self.GetCardPoint(v) {
			return true
		}
	}
	return false
}

//当前牌是否全是赖子
func (self *BaseLogicDG) IsAllMagic(card_list [static.MAX_CARD]byte, cardlen int) bool {
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
		if card_list[i] > 0 && !self.IsMagic(card_list[i]) {
			return false
		}
	}
	return true
}

//当前牌是否除赖子外全部相同
func (self *BaseLogicDG) IsAllEqualExceptMagic(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	if cardlen > static.MAX_CARD || cardlen > int(self.MaxCardCount) {
		return 0, false
	}
	k := byte(0)
	for i := cardlen - 1; i >= 0; i-- {
		if self.IsMagic(card_list[i]) {
			continue
		}
		k = card_list[i]
		break
	}
	if k == 0 {
		return 0, false
	}
	for i := cardlen - 1; i >= 0; i-- {
		if self.IsMagic(card_list[i]) {
			continue
		}
		if self.GetCardLevel(k) != self.GetCardLevel(card_list[i]) {
			return 0, false
		}
	}
	return k, true
}

//获取赖子的数量
func (self *BaseLogicDG) GetMagicNum(card_list [static.MAX_CARD]byte, cardlen int, fakepoker []static.TFakePoker) (int, []static.TFakePoker) {
	num := 0
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
		if card_list[i] > 0 && self.IsMagic(card_list[i]) {
			if fakepoker != nil {
				fakepoker[num].Index = card_list[i]
			}
			num++
		}
	}
	return num, fakepoker
}

//当前牌是否是赖子
func (self *BaseLogicDG) IsKing(card_id byte) bool {
	if self.GetCardPoint(card_id) == self.GetCardPoint(CARDINDEX_SMALL) || self.GetCardPoint(card_id) == self.GetCardPoint(CARDINDEX_BIG) {
		return true
	}
	return false
}

//当前牌是否全是赖子
func (self *BaseLogicDG) IsAllKing(card_list [static.MAX_CARD]byte, cardlen int) bool {
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
		if card_list[i] > 0 && !self.IsKing(card_list[i]) {
			return false
		}
	}
	return true
}

//当前牌是否除赖子外全部相同
func (self *BaseLogicDG) IsAllEqualExceptKing(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	if cardlen > static.MAX_CARD || cardlen > int(self.MaxCardCount) {
		return 0, false
	}
	k := card_list[cardlen-1]
	for i := cardlen - 1; i >= 0; i-- {
		if self.GetCardLevel(k) != self.GetCardLevel(card_list[i]) {
			return 0, false
		}
	}
	if k == 0 {
		return 0, false
	}
	return k, true
}

//获取王的数量
func (self *BaseLogicDG) GetKingNum(card_list [static.MAX_CARD]byte, cardlen int) int {
	num := 0
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
		if card_list[i] > 0 && self.IsKing(card_list[i]) {
			num++
		}
	}
	return num
}

//判断是不是普通炸弹，如果是，返回true 对王炸、510k炸弹分开来
// 普通炸弹单纯的指的是三张以上相同的牌
func (self *BaseLogicDG) IsNormalBomb(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	max_bombCnt := 8
	if self.IsMagic(CARDINDEX_BIG) {
		max_bombCnt = 12 //4个王+8张同样的牌组成的炸弹
	}
	if cardlen > max_bombCnt || cardlen < int(self.MinBombCount) { //3到8张同样的牌组成的炸弹
		return 0, false
	}
	p, bBomb := self.IsAllEqualExceptMagic(card_list, cardlen)
	if bBomb {
		return p, true
	} else {

	}
	return 0, false
}

// 判断是不是四个王
func (self *BaseLogicDG) IsKingBomb(card_list [static.MAX_CARD]byte, cardlen int) bool {
	if cardlen != 4 { //判断4王的函数
		return false
	}
	bBomb := self.IsAllKing(card_list, cardlen)
	if bBomb {
		return true
	}
	return false
}

//判断是不是510k，如果是，返回花色1-4代表了方块，梅花，红桃，黑桃,0代表杂的
func (self *BaseLogicDG) Is510KBomb(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	color := byte(0)
	if cardlen != 3 {
		return color, false
	}
	has_5 := 0
	has_10 := 0
	has_k := 0
	has_m := 0 //赖子数目
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(self.MaxCardCount); i++ {
		if self.GetCardLevel(card_list[i]) == self.GetCardLevel(5) {
			has_5++
		} else if self.GetCardLevel(card_list[i]) == self.GetCardLevel(10) {
			has_10++
		} else if self.GetCardLevel(card_list[i]) == self.GetCardLevel(13) {
			has_k++
		} else if self.IsMagic(card_list[i]) {
			has_m++
		}
	}
	if has_5 == 1 && has_10 == 1 && has_k == 1 {
		var poker [3]static.TPoker
		poker[0].Set(card_list[0])
		poker[1].Set(card_list[1])
		poker[2].Set(card_list[2])
		if poker[0].Color == poker[1].Color && poker[0].Color == poker[2].Color {
			color = poker[0].Color
		} else {
			color = 0
		}
		return color, true
	}
	if has_m > 0 && has_m < 3 {
		//has_m == 3 的情况在外面判断，因为有的玩法3个王当3个2，有的当3个3，有的当正黑桃510k，放在这里反而不方便
		if has_5 > 1 || has_10 > 1 || has_k > 1 {
			//任何一种大于1个，就不可能组成510k，比如2个5和1个王。
			return 0, false
		}
		if has_m+has_5+has_10+has_k == 3 {
			var poker [3]static.TPoker
			poker[0].Set(card_list[0])
			poker[1].Set(card_list[1])
			poker[2].Set(card_list[2])
			for i := 0; i < 3; i++ {
				if poker[i].Color == 0 {
					continue
				}
				if color != 0 && poker[i].Color != color { //有一个颜色不一样，那么就是杂的
					color = 0
					break
				}
				color = poker[i].Color
			}
			return color, true
		}
	}
	return 0, false
}

//获取牌型，公共代码，每次修改这个函数时，都需要测试是否影响了蕲春打拱和大冶打拱，
func (self *BaseLogicDG) GetType(card_list [static.MAX_CARD]byte, cardlen int, outMagicNum byte, byType byte, outtype int) static.TCardType {
	var re static.TCardType
	re.Len = 0
	re.Card = 0
	re.Color = 0
	re.Cardtype = static.TYPE_NULL
	re.Count = 0
	re.BombLevel = 0

	card := byte(0)
	self.SortByIndex(card_list, cardlen, true)
	re.Len = int(self.GetCardNum(card_list, byte(cardlen)))
	if re.Len < 1 {
		return re
	}
	re.Card = card_list[re.Len-1]
	switch re.Len {
	case 0:
		re.Cardtype = static.TYPE_NULL
		return re
	case 1:
		re.Cardtype = static.TYPE_ONE
		//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		if self.IsMagic(CARDINDEX_BIG) && self.IsKing(re.Card) {
			re.Card = 2 //王单出算2
		}
		return re
	case 2:
		if !self.IsKing(card_list[0]) && !self.IsKing(card_list[1]) {
			if self.GetCardLevel(card_list[0]) == self.GetCardLevel(card_list[1]) {
				re.Cardtype = static.TYPE_TWO
			} else {
				re.Cardtype = static.TYPE_ERROR
			}
		} else if self.IsKing(card_list[0]) && self.IsKing(card_list[1]) {
			//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
			if self.IsMagic(CARDINDEX_BIG) {
				re.Card = 2 //大小王按2来算
				re.Cardtype = static.TYPE_TWO
			} else {
				//对王
				if card_list[0] == card_list[1] {
					if card_list[0] == CARDINDEX_SMALL {
						re.BombLevel = 60 + 2 //对小王，比6张的炸弹大一些
					} else {
						re.BombLevel = 60 + 3 //对大王，比6张的炸弹大一些
					}
				} else {
					// 大小王组成的一对
					re.BombLevel = 60 + 1
				}
				re.Cardtype = static.TYPE_BOMB_DOUBLE_KING
			}
		} else {
			//有一张王
			//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
			if self.IsMagic(CARDINDEX_BIG) {
				re.Card = card_list[1]
				if self.IsMagic(card_list[1]) {
					re.Card = card_list[0]
				}
				re.Cardtype = static.TYPE_TWO
			} else {
				re.Cardtype = static.TYPE_ERROR
			}
		}
		return re
	case 3:
		color := byte(0)
		is510k := false
		color, is510k = self.Is510KBomb(card_list, 3)
		isBombNormal := false
		card, isBombNormal = self.IsNormalBomb(card_list, cardlen)
		if is510k {
			re.Color = color
			re.Cardtype = static.TYPE_BOMB_510K
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if self.IsMagic(CARDINDEX_BIG) { //王是赖子
				re.Card = 5
				if color == 0 {
					re.BombLevel = 30 + 1 //比4张的炸弹小一点
				} else {
					re.BombLevel = 30 + int(color) + 1 //比4张的炸弹小一点，比杂510k大(+1)
				}
			} else {
				if color == 0 {
					re.BombLevel = 40 + 1 //比4张的炸弹大一点
				} else {
					re.BombLevel = 50 + int(color) //比5张的炸弹大一点
				}
			}
		} else if self.IsKing(card_list[0]) && self.IsKing(card_list[1]) && self.IsKing(card_list[2]) {
			//三个王
			//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
			if self.IsMagic(CARDINDEX_BIG) {
				re.Color = 4              //3个大小王按黑桃正510k来算
				re.BombLevel = 30 + 4 + 1 //比4张的炸弹小一些，比杂510k大(+1)
				re.Cardtype = static.TYPE_BOMB_510K
				re.Card = 5 //3个大小王按黑桃正510k来算
				//re.Card = 3;//3个大小王按3来算
				//re.Cardtype = public.TYPE_THREE;
			} else {
				re.BombLevel = 60 + 4 //比6张的炸弹大一些
				re.Cardtype = static.TYPE_BOMB_DOUBLE_KING
			}
		} else if isBombNormal {
			re.BombLevel = re.Len * 10
			re.Card = card
			re.Cardtype = static.TYPE_BOMB_NOMORL
		} else {
			//ONESTR
			var refakepoker []static.TFakePoker
			iMagicNum := 0
			iMagicNum, refakepoker = self.GetMagicNum(card_list, cardlen, refakepoker)
			if byType != 0 { //非0表示没有赖子模式
				iMagicNum = 0
			}
			if iMagicNum <= 8 { //有1-4个王
				var ty static.FakeType
				ty = self.GetTypeByMagic(card_list, cardlen, iMagicNum, outtype)
				re.Cardtype = ty.CardType.Cardtype
				re.Card = ty.CardType.Card
			} else {
				re.Cardtype = static.TYPE_ERROR
			}
		}
		return re
	default: // >= 4张
		if re.Len == 4 {
			card = byte(0)
			if self.IsKingBomb(card_list, cardlen) { //四个王最大
				//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
				if self.IsMagic(CARDINDEX_BIG) {
					re.BombLevel = 40 //4个王就是4张的炸弹
					re.Card = 2       //4个王就是4个2
					re.Cardtype = static.TYPE_BOMB_NOMORL
				} else {
					//re.BombLevel = 60 + 2000;
					re.BombLevel = 70 + 1 //比7张的炸弹大些，比8张的炸弹小
					re.Cardtype = static.TYPE_BOMB_FOUR_KING
				}
				return re
			}
		}

		card = byte(0)
		isBombNormal := false
		card, isBombNormal = self.IsNormalBomb(card_list, cardlen)
		if isBombNormal {
			re.Card = card
			if re.Len == 8 {
				re.Cardtype = static.TYPE_BOMB_8XI
			} else {
				re.Cardtype = static.TYPE_BOMB_NOMORL
			}
			re.BombLevel = re.Len * 10
			return re
		}
		//单顺，双顺等
		var refakepoker []static.TFakePoker
		iMagicNum := 0
		iMagicNum, refakepoker = self.GetMagicNum(card_list, cardlen, refakepoker)
		if byType != 0 { //非0表示没有赖子模式，给服务器用
			iMagicNum = 0
		}
		if iMagicNum <= 8 { //有1-4个王
			var ty static.FakeType
			ty = self.GetTypeByMagic(card_list, cardlen, iMagicNum, outtype)
			re.Cardtype = ty.CardType.Cardtype
			re.Card = ty.CardType.Card
		} else {
			re.Cardtype = static.TYPE_ERROR
		}
		return re
	}
	return re
}

//有赖子的情况下（赖子数可以为0），得出他的牌型，这个只针对三连对或者2连对，超过4只的可以参考通山打拱的拱笼
func (self *BaseLogicDG) IsStrByMagic(cardlen int, iMagicNum int, iZhaNum int) static.FakeType {
	var temptype static.FakeType
	if len(self.m_allCard) < 24 || (cardlen > iZhaNum && self.m_allCard[12].Count != 0) || iZhaNum < 1 || iZhaNum > 12 || cardlen%iZhaNum != 0 {
		//有2，不可能组成连对
		temptype.CardType.Cardtype = static.TYPE_ERROR
		return temptype
	}
	byLeftMagicNum := iMagicNum
	var templist [static.MAX_CARD]byte

	bLianFlag := true
	iLianNum := int(0)
	iNewcardlen := int(0)
	iUseMagicCount := int(0)
	iMaxIndex := 11
	if cardlen == iZhaNum { //不是连对或飞机或顺子时，可以有2
		iMaxIndex = 12
	}
	for iIndex := 0; iIndex <= iMaxIndex && iNewcardlen < cardlen; iIndex++ {
		if iLianNum == 0 && self.m_allCard[iIndex].Count == 0 {
			continue
		}
		iLianNum++
		//构造数据
		for iCardCount := 0; iCardCount < iZhaNum && byte(iCardCount) < self.m_allCard[iIndex].Count; iCardCount++ {
			templist[iNewcardlen] = self.m_allCard[iIndex].Indexes[iCardCount]
			iNewcardlen++
			temptype.CardType.Card = self.m_allCard[iIndex].Point //存放最大点数，暂时赋值
		}
		if self.m_allCard[iIndex].Count <= byte(iZhaNum) && byLeftMagicNum >= iZhaNum-int(self.m_allCard[iIndex].Count) {
			byNeedMagicNum := iZhaNum - int(self.m_allCard[iIndex].Count)
			byLeftMagicNum -= byNeedMagicNum
			for byMagic := 0; byMagic < byNeedMagicNum && byLeftMagicNum >= 0; byMagic++ {
				templist[iNewcardlen] = self.m_vMagic[iUseMagicCount].Indexes[0]
				iNewcardlen++
				temptype.CardType.Card = self.m_allCard[iIndex].Point //存放最大点数，暂时赋值
				//temptype.fakeking[iUseMagicCount].fakeindex = iIndex;//代替的值
				//temptype.fakeking[iUseMagicCount].index = fakepoker[iUseMagicCount++].index;//存放原来的值，就是王
				temptype.Fakeking[iUseMagicCount].Fakeindex = self.m_allCard[iIndex].Point         //代替的值
				temptype.Fakeking[iUseMagicCount].Index = self.m_vMagic[iUseMagicCount].Indexes[0] //存放原来的值，就是赖子
				iUseMagicCount++
			}
		} else {
			bLianFlag = false
			break
			//不能组成连对
		}
	}
	//可以组成连对
	if bLianFlag && byLeftMagicNum%iZhaNum == 0 {
		//Go语言不支持？这个3目运算符，只能写if else了
		temptype.CardType.Cardtype = static.TYPE_THREESTR
		if iZhaNum == 2 {
			temptype.CardType.Cardtype = static.TYPE_TWOSTR
		} else if iZhaNum == 1 {
			temptype.CardType.Cardtype = static.TYPE_ONESTR
		}
		if cardlen <= 3 && static.TYPE_ONESTR != temptype.CardType.Cardtype {
			temptype.CardType.Cardtype = static.TYPE_THREE
			if iZhaNum == 2 {
				temptype.CardType.Cardtype = static.TYPE_TWO
			} else if iZhaNum == 1 {
				temptype.CardType.Cardtype = static.TYPE_ONE
			}
		}
	} else {
		temptype.CardType.Cardtype = static.TYPE_ERROR
	}
	return temptype
}

//有赖子的情况下（赖子数可以为0），得出他的牌型，这个只针对三连对或者2连对，超过4只的可以参考通山打拱的拱笼
func (self *BaseLogicDG) GetTypeByMagic(card_list [static.MAX_CARD]byte, cardlen int, iMagicNum int, outtype int) static.FakeType {
	var reType static.FakeType
	if iMagicNum >= 0 {
		var vTypeList []static.FakeType
		self.m_allCard, self.m_vMagic, self.m_vKing = self.GetGroupTypeByPoint(card_list)
		if cardlen >= int(self.MinOnestrCount) { //单顺最小长度在这里判断一下，单张在gettype里面判断
			var temptypeOne static.FakeType
			temptypeOne = self.IsStrByMagic(cardlen, iMagicNum, 1)
			if temptypeOne.CardType.Cardtype != static.TYPE_ERROR {
				vTypeList = append(vTypeList, temptypeOne)
			}
		}
		var temptypeTwo static.FakeType
		temptypeTwo = self.IsStrByMagic(cardlen, iMagicNum, 2)
		if temptypeTwo.CardType.Cardtype != static.TYPE_ERROR {
			vTypeList = append(vTypeList, temptypeTwo)
		}
		var temptypeThree static.FakeType
		temptypeThree = self.IsStrByMagic(cardlen, iMagicNum, 3)
		if temptypeThree.CardType.Cardtype != static.TYPE_ERROR {
			vTypeList = append(vTypeList, temptypeThree)
		}

		if len(vTypeList) > 0 {
			var max_re static.FakeType
			if outtype == static.TYPE_THREESTR || outtype == static.TYPE_TWOSTR || outtype == static.TYPE_ONESTR || outtype == static.TYPE_THREE || outtype == static.TYPE_TWO || outtype == static.TYPE_ONE { //说明手先出的人已经指定了出牌类型，那么跟的就是出牌类型
				first := 0
				for i := 0; i < len(vTypeList); i++ {
					if vTypeList[i].CardType.Cardtype == outtype {
						max_re = vTypeList[i]
						first = i
						break
					}
				}
				for i := first; i < len(vTypeList); i++ {
					if vTypeList[i].CardType.Cardtype == max_re.CardType.Cardtype {
						if self.Compare(max_re.CardType, vTypeList[i].CardType) {
							max_re = vTypeList[i]
						}
					}
				}
			} else {
				max_re = vTypeList[0]
				for i := 1; i < len(vTypeList); i++ {
					if vTypeList[i].CardType.Cardtype > max_re.CardType.Cardtype {
						max_re = vTypeList[i]
					} else if vTypeList[i].CardType.Cardtype == max_re.CardType.Cardtype {
						if self.Compare(max_re.CardType, vTypeList[i].CardType) {
							max_re = vTypeList[i]
						}
					}
				}
			}
			reType = max_re
		} else {
			reType.CardType.Cardtype = static.TYPE_ERROR
		}
	} else {
		reType.CardType.Cardtype = static.TYPE_ERROR
	}
	return reType
}

func (self *BaseLogicDG) GetGroupType(myCard [static.MAX_CARD]byte) {
	self.m_vOne = []static.TPokerGroup{}
	self.m_vTwo = []static.TPokerGroup{}
	self.m_vThree = []static.TPokerGroup{}
	self.m_vFour = []static.TPokerGroup{}
	self.m_vFive = []static.TPokerGroup{}
	self.m_vSix = []static.TPokerGroup{}
	self.m_vSeven = []static.TPokerGroup{}
	self.m_vEight = []static.TPokerGroup{}
	self.m_vKing = []static.TPokerGroup{}
	self.m_vAllHandKing = []static.TPokerGroup{}
	self.m_vMagic = []static.TPokerGroup{}
	self.m_vAllHandMagic = []static.TPokerGroup{}
	self.m_vK105 = []static.TPokerGroup{}

	var tempPoker [static.MAX_CARD]static.TPoker
	for i := 0; i < len(myCard); i++ {
		tempPoker[i].Set(myCard[i])
	}

	var group static.TPokerGroup
	//首先找王
	for i := 14; i < 15; i++ {
		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
			if tempPoker[j].Point == byte(i) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				self.m_vKing = append(self.m_vKing, group)
			}
		}
	}
	//找赖子
	for i := 0; i < len(self.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
			if tempPoker[j].Point == self.GetCardPoint(self.m_allMagicPoint[i]) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				self.m_vMagic = append(self.m_vMagic, group)
			}
		}
	}
	//找其他的
	pokerIndex := [13]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2}
	for i := 0; i < 13; i++ {
		//赖子不重复放，只放在m_vMagic和m_vAllHandMagic中
		bIsMagic := self.IsMagic(pokerIndex[i])
		if bIsMagic {
			continue
		}

		num := byte(0)
		indexM := [10]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
			if tempPoker[j].Point == pokerIndex[i] {
				indexM[num] = tempPoker[j].Index
				num++
				if pokerIndex[i] == 5 || pokerIndex[i] == 10 || pokerIndex[i] == 13 {
					group.Point = pokerIndex[i]
					group.Color = tempPoker[j].Color
					group.Count = 1
					group.Indexes = []uint8{}
					group.Indexes = append(group.Indexes, tempPoker[j].Index)
					self.m_vK105 = append(self.m_vK105, group)
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
			self.m_vEight = append(self.m_vEight, group)
		} else if num == 7 {
			self.m_vSeven = append(self.m_vSeven, group)
		} else if num == 6 {
			self.m_vSix = append(self.m_vSix, group)
		} else if num == 5 {
			self.m_vFive = append(self.m_vFive, group)
		} else if num == 4 {
			self.m_vFour = append(self.m_vFour, group)
		} else if num == 3 {
			self.m_vThree = append(self.m_vThree, group)
		} else if num == 2 {
			self.m_vTwo = append(self.m_vTwo, group)
		} else if num == 1 {
			self.m_vOne = append(self.m_vOne, group)
		}
	}
	self.m_allHandCards, self.m_vAllHandMagic, self.m_vAllHandKing = self.GetGroupTypeByPoint(myCard)
}

func (self *BaseLogicDG) GetGroupTypeByPoint(myCard [static.MAX_CARD]byte) ([]static.TPokerGroup, []static.TPokerGroup, []static.TPokerGroup) {
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
		//赖子不重复放，只放在m_vMagic和m_vAllHandMagic中
		bIsMagic := self.IsMagic(pokerIndex[i])
		if bIsMagic {
			continue
		}
		var group static.TPokerGroup
		group.Point = pokerIndex[i]
		group.Indexes = []uint8{}
		group.Color = 0
		group.Count = 0
		for j := 0; !bIsMagic && j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
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
		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
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
	for i := 0; i < len(self.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
			if tempPoker[j].Point == self.GetCardPoint(self.m_allMagicPoint[i]) {
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

func (self *BaseLogicDG) Compare(typeFirst static.TCardType, typeFollow static.TCardType) bool {
	if typeFirst.Cardtype == static.TYPE_ERROR {
		return false
	}
	if typeFollow.Cardtype == static.TYPE_ERROR || typeFollow.Cardtype == static.TYPE_NULL {
		return false
	}

	//第一种情况，首出
	if typeFirst.Cardtype == static.TYPE_NULL {
		if typeFollow.Cardtype == static.TYPE_ONE || typeFollow.Cardtype == static.TYPE_TWO || typeFollow.Cardtype == static.TYPE_TWOSTR || typeFollow.Cardtype == static.TYPE_ONESTR || typeFollow.Cardtype == static.TYPE_BOMB_510K || typeFollow.Cardtype == static.TYPE_BOMB_8XI || typeFollow.Cardtype == static.TYPE_BOMB_DOUBLE_KING || typeFollow.Cardtype == static.TYPE_BOMB_FOUR_KING || typeFollow.Cardtype == static.TYPE_BOMB_NOMORL {
			return true
		} else if self.IsMagic(CARDINDEX_BIG) && (typeFollow.Cardtype == static.TYPE_THREE || typeFollow.Cardtype == static.TYPE_THREESTR) {
			//如果王是赖子：大冶王是赖子，蕲春王不是赖子，通过王是赖子来区分是否是大冶打拱，大冶有TYPE_THREE和TYPE_THREESTR
			return true
		} else {
			return false
		}
	} else { //第二种情况，跟出的人出的不是炸弹，类型必须和首出一致//其他牌的比较，非炸弹
		if (typeFollow.Cardtype == static.TYPE_BOMB_510K) || (typeFollow.Cardtype == static.TYPE_BOMB_NOMORL) || (typeFollow.Cardtype == static.TYPE_BOMB_8XI) || (typeFollow.Cardtype == static.TYPE_BOMB_DOUBLE_KING) || (typeFollow.Cardtype == static.TYPE_BOMB_FOUR_KING) {
			//非拱笼类型的BombLevel =0,
			if typeFollow.BombLevel > typeFirst.BombLevel || (typeFollow.BombLevel == typeFirst.BombLevel && self.GetCardLevel(typeFollow.Card) > self.GetCardLevel(typeFirst.Card)) {
				return true
			} else {
				return false
			}
		} else if typeFollow.Cardtype == typeFirst.Cardtype { //跟出的人出的不是炸弹，类型必须和首出一致
			if typeFollow.Len == typeFirst.Len && self.GetCardLevel(typeFollow.Card) > self.GetCardLevel(typeFirst.Card) {
				return true
			} else {
				return false
			}
		} else { //跟出的人出的不是炸弹，类型和首出也一致
			return false
		}
	}
	return false
}

//计算2连对和3连对、单连、三只、两只的函数。使用StrTool请确保GetGroupType函数被调用过，这样才能保证m_allHandCards 是最新的。
func (self *BaseLogicDG) StrTool(outpoint byte, iLianNum int, iZhaNum int) (bool, []static.TPokerGroup) {
	var combinelist []static.TPokerGroup
	//添加代码让它适用于顺子
	if (iZhaNum == 1 && (iLianNum < int(self.MinOnestrCount) || iLianNum > 13)) || iZhaNum < 1 || iZhaNum > 12 || iLianNum < 1 || iLianNum > 12 {
		return false, nil
	}
	iMagicNum := len(self.m_vAllHandMagic)
	if iMagicNum < 0 || iMagicNum > 8 {
		return false, nil
	}
	byOutPoint := self.GetCardPoint(outpoint)
	if byOutPoint != 0 && (int(byOutPoint)-iLianNum+1 < 3 || (iLianNum > 1 && byOutPoint == 1) || (iLianNum <= 1 && byOutPoint == 2)) {
		//他自己已经通天了，别人不可能比它大,连时A最大，非连时(2只或3只)2最大
		return false, nil
	}
	if len(self.m_allHandCards) < 24 {
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
		if self.m_allHandCards[iIndex].Count == 0 && (iIndex+iLianNum-1 < 11) {
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
			byTmepHaveNum := 0 //手中有这个牌的数量
			for iCardCount := 0; iCardCount < iZhaNum && iCardCount < int(self.m_allHandCards[idx].Count); iCardCount++ {
				card_list[cardlen] = self.m_allHandCards[idx].Indexes[iCardCount]
				cardlen++
				byTmepHaveNum++
				if iZhaNum < int(self.m_allHandCards[idx].Count) {
					sortRight += (int(self.m_allHandCards[idx].Count) - iZhaNum) //拆牌要加权重
				}
			}
			byPoint = self.m_allHandCards[idx].Point //直接获取最大的byPoint
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
				fakeking[byMagic].Index = self.m_vAllHandMagic[byMagic].Indexes[0] //替换前的值,这个要保证是唯一的
				card_list[cardlen] = self.m_vAllHandMagic[byMagic].Indexes[0]
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
				for faker := 0; faker < static.MAX_CARD && faker < int(self.MaxCardCount); faker++ {
					tPKstr.Fakeking[faker] = fakeking[faker]
				}
				tPKstr.Cardtype = static.TYPE_THREESTR
				if iZhaNum == 2 {
					tPKstr.Cardtype = static.TYPE_TWOSTR
				} else if iZhaNum == 1 {
					tPKstr.Cardtype = static.TYPE_ONESTR
				}
				if iLianNum == 1 {
					tPKstr.Cardtype = static.TYPE_THREE
					if iZhaNum == 2 {
						tPKstr.Cardtype = static.TYPE_TWO
					} else if iZhaNum == 1 {
						tPKstr.Cardtype = static.TYPE_ONE //实际上单张不会用这个函数
					}
				}
				combinelist = append(combinelist, tPKstr)
			}
		}
	}

	return bHasStr, combinelist
}

// 得到所有的顺子
func (self *BaseLogicDG) GetCombineOneStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(self.m_vAllHandMagic);

	if outlen < int(self.MinOnestrCount) || outlen > 12 {
		return
	}
	_, combinelist = self.StrTool(outpoint, outlen, 1)
	combinelist = self.SortBeepCardList(combinelist)
	return
}

// 得到所有的两连对
func (self *BaseLogicDG) GetCombineTwoStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(self.m_vAllHandMagic);

	if outlen%2 != 0 {
		return
	}
	_, combinelist = self.StrTool(outpoint, outlen/2, 2)
	combinelist = self.SortBeepCardList(combinelist)
	return
}
func (self *BaseLogicDG) GetCombineThreeStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(self.m_vAllHandMagic);
	if outlen%3 != 0 {
		return
	}
	_, combinelist = self.StrTool(outpoint, outlen/3, 3)
	//最后可以做一次优化，让没有王的先提示，有王的后提示
	combinelist = self.SortBeepCardList(combinelist)
	return
} //end GetCombineThreeStr

//得到所有的对子
func (self *BaseLogicDG) GetCombineTwo(outpoint byte) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(self.m_vAllHandMagic);
	_, combinelist = self.StrTool(outpoint, 1, 2)
	combinelist = self.SortBeepCardList(combinelist)
	return
}

//得到所有的三只
func (self *BaseLogicDG) GetCombineThree(outpoint byte) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(self.m_vAllHandMagic);
	_, combinelist = self.StrTool(outpoint, 1, 3)
	combinelist = self.SortBeepCardList(combinelist)
	return
}

//得到所有大过outpoint的单牌,先直接翻译后面再优化吧
func (self *BaseLogicDG) GetAllOne(outpoint byte) (combinelist []static.TPokerGroup) {
	var group static.TPokerGroup
	for it := 0; it < len(self.m_vOne); it++ {
		// 过滤掉王提示
		if self.GetCardLevel(self.m_vOne[it].Point) == self.GetCardLevel(53) || self.GetCardLevel(self.m_vOne[it].Point) == self.GetCardLevel(54) {
			continue
		}

		if self.GetCardLevel(self.m_vOne[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vOne[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vOne[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}

	bBeepSKing := false
	bBeepBKing := false
	for it := 0; it < len(self.m_vKing); it++ {
		// 一种花色的王之提示一次就可以// 提示单张小王
		if self.GetCardLevel(self.m_vKing[it].Point) > self.GetCardLevel(outpoint) && self.GetCardLevel(self.m_vKing[it].Point) == self.GetCardLevel(53) && !bBeepSKing {
			group.SortRight = 100 + len(self.m_vKing) //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vKing[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vKing[it].Indexes[0])
			combinelist = append(combinelist, group)
			bBeepSKing = true
		}
		if self.GetCardLevel(self.m_vKing[it].Point) > self.GetCardLevel(outpoint) && self.GetCardLevel(self.m_vKing[it].Point) == self.GetCardLevel(54) && !bBeepBKing {
			group.SortRight = 100 + len(self.m_vKing) //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vKing[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vKing[it].Indexes[0])
			combinelist = append(combinelist, group)
			bBeepBKing = true
		}
	}

	for it := 0; it < len(self.m_vTwo); it++ {
		// 过滤掉对王
		if self.GetCardLevel(self.m_vTwo[it].Point) == self.GetCardLevel(53) || self.GetCardLevel(self.m_vTwo[it].Point) == self.GetCardLevel(54) {
			continue
		}

		if self.GetCardLevel(self.m_vTwo[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 1 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vTwo[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vTwo[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}

	for it := 0; it < len(self.m_vThree); it++ {
		if self.GetCardLevel(self.m_vThree[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 2 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vThree[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vThree[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(self.m_vFour); it++ {
		if self.GetCardLevel(self.m_vFour[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 3 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vFour[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vFour[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(self.m_vFive); it++ {
		if self.GetCardLevel(self.m_vFive[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 4 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vFive[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vFive[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(self.m_vSix); it++ {
		if self.GetCardLevel(self.m_vSix[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 5 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vSix[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vSix[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(self.m_vSeven); it++ {
		if self.GetCardLevel(self.m_vSeven[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 6 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vSeven[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vSeven[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(self.m_vEight); it++ {
		if self.GetCardLevel(self.m_vEight[it].Point) > self.GetCardLevel(outpoint) {
			group.SortRight = 100 + 7 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = self.m_vEight[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, self.m_vEight[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	return
}

//提示出牌
func (self *BaseLogicDG) BeepCardOut(otherout [static.MAX_CARD]byte, outtype int) (breturn bool, mybeepout []static.TPokerGroup) {
	var group static.TPokerGroup
	group.Color = 0
	len1 := self.GetCardNum(otherout, self.MaxCardCount)
	cardtype1 := self.GetType(otherout, int(len1), 0, 0, 0)

	cardtype1.Cardtype = outtype //重新设置下

	outlevel := self.GetCardLevel(cardtype1.Card)
	if outtype == static.TYPE_NULL || outtype == static.TYPE_ERROR {
		return false, nil //不需要提示
		//} else if (outtype == public.TYPE_BOMB_FOUR_KING){
		//	// 四个王
		//	return false,nil;
	} else if outtype == static.TYPE_THREESTR {
		var templist []static.TPokerGroup
		templist = self.GetCombineThreeStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == static.TYPE_TWOSTR {
		var templist []static.TPokerGroup
		templist = self.GetCombineTwoStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == static.TYPE_ONESTR {
		var templist []static.TPokerGroup
		templist = self.GetCombineOneStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == static.TYPE_THREE {
		var templist []static.TPokerGroup
		templist = self.GetCombineThree(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == static.TYPE_TWO {
		var templist []static.TPokerGroup
		templist = self.GetCombineTwo(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == static.TYPE_ONE {
		var templist []static.TPokerGroup
		templist = self.GetAllOne(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	}

	//找拱拢(炸弹)
	mybombstr := self.GetAllXiGL()
	for i := 0; i < len(mybombstr); i++ {
		//非拱笼类型的cardtype1.count =0,
		if mybombstr[i].BombLevel > cardtype1.BombLevel || (mybombstr[i].BombLevel == cardtype1.BombLevel && self.GetCardLevel(mybombstr[i].Point) > outlevel) {
			group.Cardtype = static.TYPE_BOMB_NOMORL
			group.Indexes = []byte{}
			group.Count = mybombstr[i].Count
			group.Point = mybombstr[i].Point
			for j := 0; j < len(mybombstr[i].Indexes); j++ {
				group.Indexes = append(group.Indexes, mybombstr[i].Indexes[j])
			}
			for j := 0; j < static.MAX_CARD && j < int(self.MaxCardCount); j++ {
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
func (self *BaseLogicDG) BeepFirstCardOut() (breturn bool, mybeepout []static.TPokerGroup) {
	var group static.TPokerGroup

	var templist []static.TPokerGroup

	for it := 0; it < len(self.m_vOne); it++ { //1
		group.Indexes = []byte{}
		group.Count = 1
		group.Point = self.m_vOne[it].Point
		for j := 0; j < len(self.m_vOne[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vOne[it].Indexes[j])
		}
		templist = append(templist, group)
	}
	for it := 0; it < len(self.m_vTwo); it++ { //2
		group.Indexes = []byte{}
		group.Count = 2
		group.Point = self.m_vTwo[it].Point
		for j := 0; j < len(self.m_vTwo[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vTwo[it].Indexes[j])
		}
		templist = append(templist, group)
	}

	//3个以内的，都用最小的提示

	size := len(templist)
	for i := 0; i < size-1; i++ {
		for j := i + 1; j < size; j++ {
			if self.GetCardLevel(templist[i].Point) > self.GetCardLevel(templist[j].Point) && templist[j].Point > 0 {
				temp := templist[i]
				templist[i] = templist[j]
				templist[j] = temp
			}
		}
	}
	if size > 0 {
		temp := templist[0]
		mybeepout = append(mybeepout, temp)
		return true, mybeepout
	}
	mybeepout = []static.TPokerGroup{}

	for it := 0; it < len(self.m_vThree); it++ { //3
		group.Indexes = []byte{}
		group.Count = 3
		group.Point = self.m_vThree[it].Point
		for j := 0; j < len(self.m_vThree[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vThree[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(self.m_vFour); it++ { //4
		group.Indexes = []byte{}
		group.Count = 4
		group.Point = self.m_vFour[it].Point
		for j := 0; j < len(self.m_vFour[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vFour[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(self.m_vFive); it++ { //5
		group.Indexes = []byte{}
		group.Count = 5
		group.Point = self.m_vFive[it].Point
		for j := 0; j < len(self.m_vFive[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vFive[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(self.m_vSix); it++ { //6
		group.Indexes = []byte{}
		group.Count = 6
		group.Point = self.m_vSix[it].Point
		for j := 0; j < len(self.m_vSix[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vSix[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(self.m_vSeven); it++ { //7
		group.Indexes = []byte{}
		group.Count = 7
		group.Point = self.m_vSeven[it].Point
		for j := 0; j < len(self.m_vSeven[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vSeven[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(self.m_vEight); it++ { //8
		group.Indexes = []byte{}
		group.Count = 8
		group.Point = self.m_vEight[it].Point
		for j := 0; j < len(self.m_vEight[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vEight[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}

	for it := 0; it < len(self.m_vKing); it++ { //king
		group.Indexes = []byte{}
		group.Count = 1
		group.Point = self.m_vKing[it].Point
		for j := 0; j < len(self.m_vKing[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, self.m_vKing[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	return false, nil
}

//获取所有同张的炸弹，比如软4个5 ，或2个王
func (self *BaseLogicDG) GetAllXiGL() (mybombstr []static.TBombStr) {
	iMagicNum := len(self.m_vAllHandMagic) //癞子的数量
	if iMagicNum > 8 || iMagicNum < 0 {
		return
	}
	if len(self.m_allHandCards) < 24 {
		return
	}

	for i := 0; i <= iMagicNum && i < 8; i++ {
		//全由赖子组成的炸弹后面处理、赖子数最多8个
		for j := 0; j < 13; j++ {
			itemsize := len(self.m_allHandCards[j].Indexes)
			if itemsize == 0 {
				continue
			} //数目为0不能组成喜
			var item static.TBombStr
			item.Indexes = []byte{}
			if i+itemsize >= int(self.MinBombCount) {
				//限定为大于等于3张
				var fakeking [static.MAX_CARD]static.TFakePoker
				for k := 0; k < itemsize && k < 8; k++ {
					//同一种牌最多8个，2副牌
					item.Indexes = append(item.Indexes, self.m_allHandCards[j].Indexes[k])
				}
				for k := 0; k < i; k++ {
					//if (i != 8)//全由赖子组成的炸弹没有特殊性，如果有特殊性可以参考肉挨肉处理
					{
						fakeking[k].Fakeindex = self.GetCardPoint(self.m_allHandCards[j].Indexes[0]) //替换的值
						fakeking[k].Index = self.m_vAllHandMagic[k].Indexes[0]                       //替换前的值,这个要保证是唯一的
					}
					item.Indexes = append(item.Indexes, self.m_vAllHandMagic[k].Indexes[0])
				}
				item.BombLevel = (i + itemsize) * 10 //3张的炸弹nBombCount为30,4张的炸弹为40,5张炸弹为50
				item.Point = self.m_allHandCards[j].Point
				item.MaxCount = uint8(i + itemsize)
				item.Count = uint8(i + itemsize)
				for it := 0; it < static.MAX_CARD && it < int(self.MaxCardCount); it++ {
					item.Fakeking[it] = fakeking[it]
				}
				mybombstr = append(mybombstr, item)
			}
		}
	}
	//非王的癞子有4个，可以组成炸弹
	//最后手牌不是4个，全是赖子且含王
	//王炸
	//这些情况如果需要请参考肉挨肉
	var mybombKingstr []static.TBombStr
	var mybombK105str []static.TBombStr
	//王是赖子，大冶打拱王是赖子，蕲春打拱王不是赖子
	if self.IsMagic(CARDINDEX_BIG) {
	} else {
		mybombKingstr = self.GetCombineKingBomb()
		for i := 0; i < len(mybombKingstr); i++ {
			mybombstr = append(mybombstr, mybombKingstr[i])
		}
	}
	mybombK105str = self.GetAllCombineK105()
	for i := 0; i < len(mybombK105str); i++ {
		mybombstr = append(mybombstr, mybombK105str[i])
	}
	return
}

//得到用双王炸弹
func (self *BaseLogicDG) GetCombineKingBomb() (mybombstr []static.TBombStr) {
	//如果是以王的数目来算大小的，可以用这段代码
	//int kingNum = m_vAllHandKing.size();	//王的数量
	//for (int i = 2; i <= kingNum && i <= 4 ; i++)//2个王以上才是王炸，实际上只有2个王,//防止做牌时有多个王，这里限制最多4个王
	//{
	//	TBombStr item;
	//	item.index.Clear();
	//	for (int k = 0; k < i; k++)
	//	{
	//		item.index.push_back(m_vAllHandKing[k].index[0]);
	//	}
	//	item.nBombCount = GetPower(2, 6-2) + i+(i == 4?2000:0);//比6张的炸弹大一点,4王最大
	//	item.nPoint = CP_BJ_S;
	//	item.nMaxCount = i ;
	//	item.nCount = i ;
	//	mybombstr.push_back(item);
	//}
	//大小王的炸弹不同时，使用下面的代码
	skingNum := 0                       //小王的数量
	bkingNum := 0                       //大王的数量
	kingNum := len(self.m_vAllHandKing) //王的数量
	if kingNum >= 2 {
		for i := 0; i < kingNum && i < 4; i++ {
			if self.GetCardLevel(self.m_vAllHandKing[i].Indexes[0]) == self.GetCardLevel(CARDINDEX_SMALL) {
				// 小王
				skingNum++
			} else {
				bkingNum++
			}
		}
		for skingN := 0; skingN <= skingNum && skingN <= 2; skingN++ {
			for bkingN := 0; bkingN <= bkingNum && bkingN <= 2; bkingN++ {
				if skingN+bkingN < 2 {
					continue
				}
				var item static.TBombStr
				item.Indexes = []byte{}
				for k := 0; k < skingN; k++ {
					//放skingN个小王
					item.Indexes = append(item.Indexes, CARDINDEX_SMALL)
				}
				for j := 0; j < bkingN; j++ {
					//放bkingN个大王
					item.Indexes = append(item.Indexes, CARDINDEX_BIG)
				}
				item.BombLevel = 60 //比6张的炸弹大一点,4王最大
				if skingN == 0 {
					//没有小王，即肯定是一对大王
					item.BombLevel += 3
				} else if bkingN == 0 {
					//没有大王，即肯定是一对小王
					item.BombLevel += 2
				} else if skingN == 2 && bkingN == 2 {
					//4王最大
					//item.BombLevel += 2000;
					item.BombLevel += 10 + 1 //比7张的大，比8张的小
				} else {
					item.BombLevel += 1 //杂王，3个王也是杂王
				}

				item.Point = uint8(CP_BJ_S)
				item.MaxCount = uint8(skingN + bkingN)
				item.Count = uint8(skingN + bkingN)
				mybombstr = append(mybombstr, item)
			}
		}
	}
	return
}

//得到所有的k-10-5
func (self *BaseLogicDG) GetAllCombineK105() (mybombstr []static.TBombStr) {
	if len(self.m_vK105) == 0 {
		return
	}

	var tempK []static.TPoker
	var temp10 []static.TPoker
	var temp5 []static.TPoker
	for it := 0; it < len(self.m_vK105); it++ {
		var temp static.TPoker
		if self.GetCardLevel(self.m_vK105[it].Point) == self.GetCardLevel(13) {
			temp.Index = self.m_vK105[it].Indexes[0]
			temp.Color = self.m_vK105[it].Color
			tempK = append(tempK, temp)
		} else if self.GetCardLevel(self.m_vK105[it].Point) == self.GetCardLevel(10) {
			temp.Index = self.m_vK105[it].Indexes[0]
			temp.Color = self.m_vK105[it].Color
			temp10 = append(temp10, temp)
		} else if self.GetCardLevel(self.m_vK105[it].Point) == self.GetCardLevel(5) {
			temp.Index = self.m_vK105[it].Indexes[0]
			temp.Color = self.m_vK105[it].Color
			temp5 = append(temp5, temp)
		}
	}

	byColor := uint8(0)
	for itK := 0; itK < len(tempK); itK++ {
		for it10 := 0; it10 < len(temp10); it10++ {
			for it5 := 0; it5 < len(temp5); it5++ {
				if tempK[itK].Color == temp10[it10].Color && tempK[itK].Color == temp5[it5].Color {
					byColor = tempK[itK].Color
				} else {
					byColor = 0
				}

				var item static.TBombStr
				item.Indexes = []byte{}
				item.Indexes = append(item.Indexes, tempK[itK].Index)
				item.Indexes = append(item.Indexes, temp10[it10].Index)
				item.Indexes = append(item.Indexes, temp5[it5].Index)
				//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
				if self.IsMagic(CARDINDEX_BIG) { //王是赖子
					if byColor == 0 {
						item.BombLevel = 30 + 1 //比4张的炸弹小一点
					} else {
						item.BombLevel = 30 + int(byColor) + 1 //比4张的炸弹小一点，比杂510k大(+1)
					}
				} else {
					if byColor == 0 {
						item.BombLevel = 40 + 1 //比4张的炸弹大一点
					} else {
						item.BombLevel = 50 + int(byColor) //比5张的炸弹大一点
					}
				}

				item.Point = 0 //这个是无效的
				item.MaxCount = 3
				item.Count = 3
				mybombstr = append(mybombstr, item)
			}
		}
	}

	//带赖子的510k
	if len(self.m_vAllHandMagic) == 0 {
		return
	}
	temp510 := temp10[:] //5和10合并到一个切片中
	temp510 = append(temp510, temp5...)
	//k-10-王和k-王-5
	for itK := 0; itK != len(tempK); itK++ {
		for it510 := 0; it510 != len(temp510); it510++ {
			if tempK[itK].Color == temp510[it510].Color {
				byColor = tempK[itK].Color
			} else {
				byColor = 0
			}

			var item static.TBombStr
			item.Indexes = []byte{}
			item.Indexes = append(item.Indexes, tempK[itK].Index)
			item.Indexes = append(item.Indexes, temp510[it510].Index)
			item.Indexes = append(item.Indexes, self.m_vAllHandMagic[0].Indexes[0])
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if self.IsMagic(CARDINDEX_BIG) { //王是赖子
				if byColor == 0 {
					item.BombLevel = 30 + 1 //比4张的炸弹小一点
				} else {
					item.BombLevel = 30 + int(byColor) + 1 //比4张的炸弹小一点，比杂510k大(+1)
				}
			} else {
				if byColor == 0 {
					item.BombLevel = 40 + 1 //比4张的炸弹大一点
				} else {
					item.BombLevel = 50 + int(byColor) //比5张的炸弹大一点
				}
			}

			item.Point = 0 //这个是无效的
			item.MaxCount = 3
			item.Count = 3
			mybombstr = append(mybombstr, item)
		}
	}
	//王-10-5
	for it10 := 0; it10 != len(temp10); it10++ {
		for it5 := 0; it5 != len(temp5); it5++ {
			if temp10[it10].Color == temp5[it5].Color {
				byColor = temp10[it10].Color
			} else {
				byColor = 0
			}
			var item static.TBombStr
			item.Indexes = []byte{}
			item.Indexes = append(item.Indexes, temp10[it10].Index)
			item.Indexes = append(item.Indexes, temp5[it5].Index)
			item.Indexes = append(item.Indexes, self.m_vAllHandMagic[0].Indexes[0])
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if self.IsMagic(CARDINDEX_BIG) { //王是赖子
				if byColor == 0 {
					item.BombLevel = 30 + 1 //比4张的炸弹小一点
				} else {
					item.BombLevel = 30 + int(byColor) + 1 //比4张的炸弹小一点，比杂510k大(+1)
				}
			} else {
				if byColor == 0 {
					item.BombLevel = 40 + 1 //比4张的炸弹大一点
				} else {
					item.BombLevel = 50 + int(byColor) //比5张的炸弹大一点
				}
			}

			item.Point = 0 //这个是无效的
			item.MaxCount = 3
			item.Count = 3
			mybombstr = append(mybombstr, item)
		}
	}
	if len(self.m_vAllHandMagic) >= 3 {
		//3个王相当于正黑桃510k
		var item static.TBombStr
		item.Indexes = []byte{}
		item.Indexes = append(item.Indexes, self.m_vAllHandMagic[0].Indexes[0])
		item.Indexes = append(item.Indexes, self.m_vAllHandMagic[1].Indexes[0])
		item.Indexes = append(item.Indexes, self.m_vAllHandMagic[2].Indexes[0])
		//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
		if self.IsMagic(CARDINDEX_BIG) { //王是赖子
			item.BombLevel = 30 + 4 + 1 //比4张的炸弹小一点，比杂510k大(+1)
		} else {
			item.BombLevel = 50 + 4 //比5张的炸弹大一点
		}
		item.Point = 0 //这个是无效的
		item.MaxCount = 3
		item.Count = 3
		mybombstr = append(mybombstr, item)
	}
	if len(self.m_vAllHandMagic) >= 2 {
		//2个王的510k，当于正510k
		for it510k := 0; it510k != len(self.m_vK105); it510k++ {
			byColor = self.m_vK105[it510k].Color
			var item static.TBombStr
			item.Indexes = []byte{}
			item.Indexes = append(item.Indexes, self.m_vK105[it510k].Indexes[0])
			item.Indexes = append(item.Indexes, self.m_vAllHandMagic[0].Indexes[0])
			item.Indexes = append(item.Indexes, self.m_vAllHandMagic[1].Indexes[0])
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if self.IsMagic(CARDINDEX_BIG) { //王是赖子
				item.BombLevel = 30 + int(byColor) + 1 //比4张的炸弹小一点，比杂510k大(+1)
			} else {
				item.BombLevel = 50 + int(byColor) //比5张的炸弹大一点
			}

			item.Point = 0 //这个是无效的
			item.MaxCount = 3
			item.Count = 3
			mybombstr = append(mybombstr, item)
		}
	}

	return
}

//出牌排序
func (self *BaseLogicDG) SortBeepCardList(combinelist []static.TPokerGroup) []static.TPokerGroup {
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

func (self *BaseLogicDG) GetStringByCard(carddata byte) string {
	strColor := ""
	strPoint := ""

	color := self.GetCardColor(carddata)
	point := self.GetCardPoint(carddata)

	if point == info2.CARD_P_BIG_KING {
		return "rj"
	} else if point == info2.CARD_P_SMALL_KING {
		return "bj"
	} else if point == info2.CARD_P_SKY_KING {
		return "sj"
	} else {
		if color == info2.CARD_C_SPADE {
			strColor = "s"
		} else if color == info2.CARD_C_HEART {
			strColor = "h"
		} else if color == info2.CARD_C_CLUB {
			strColor = "c"
		} else if color == info2.CARD_C_DIAMOND {
			strColor = "d"
		}

		if point == info2.CARD_P_A {
			strPoint = "A"
		} else if point == info2.CARD_P_2 {
			strPoint = "2"
		} else if point == info2.CARD_P_3 {
			strPoint = "3"
		} else if point == info2.CARD_P_4 {
			strPoint = "4"
		} else if point == info2.CARD_P_5 {
			strPoint = "5"
		} else if point == info2.CARD_P_6 {
			strPoint = "6"
		} else if point == info2.CARD_P_7 {
			strPoint = "7"
		} else if point == info2.CARD_P_8 {
			strPoint = "8"
		} else if point == info2.CARD_P_9 {
			strPoint = "9"
		} else if point == info2.CARD_P_10 {
			strPoint = "10"
		} else if point == info2.CARD_P_J {
			strPoint = "J"
		} else if point == info2.CARD_P_Q {
			strPoint = "Q"
		} else if point == info2.CARD_P_K {
			strPoint = "K"
		}

		return strColor + strPoint
	}
}

func (self *BaseLogicDG) GetWriteHandReplayRecordString(replayRecord meta2.DG_Replay_Record) string {
	handCardStr := ""
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < len(replayRecord.R_HandCards[i]); j++ {
			handCardStr += fmt.Sprintf("%s,", self.GetStringByCard(byte(replayRecord.R_HandCards[i][j])))
		}
	}

	//写入分数
	handCardStr += "S:"
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d,", replayRecord.R_Score[i])
	}

	return handCardStr
}

func (self *BaseLogicDG) GetWriteOutReplayRecordString(replayRecord meta2.DG_Replay_Record) string {
	upd := false
	endMsgUpdateScore := [meta2.MAX_PLAYER]float64{}
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
			if len(record.R_Value) > 0 {
				ourCardStr += fmt.Sprintf("|C%d:", record.R_ChairId)
			} else {
				ourCardStr += fmt.Sprintf("|C%d", record.R_ChairId)
			}
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
		case meta2.DG_REPLAY_OPT_4KINGSCORE:
			ourCardStr += fmt.Sprintf("|K%d:", record.R_ChairId)
			break
		case meta2.DG_REPLAY_OPT_RESTART:
			ourCardStr += fmt.Sprintf("|R%d:", record.R_ChairId)
			break
		default:
			break
		}

		if len(record.R_Value) > 0 {
			for i := 0; i < len(record.R_Value); i++ {
				ourCardStr += fmt.Sprintf("%s", self.GetStringByCard(byte(record.R_Value[i])))
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
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_4KINGSCORE {
					ourCardStr += fmt.Sprintf(",K%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == meta2.DG_EXT_RESTART {
					ourCardStr += fmt.Sprintf(",R%d", record.R_Opt_Ext[i].Ext_value)
				}
			}
		}
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
