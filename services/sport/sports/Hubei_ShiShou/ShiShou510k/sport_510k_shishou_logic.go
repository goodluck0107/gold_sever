//////////////////////////////////////////////////////////////////////////
//                            --
//                        --       --
//                      --  纸牌游戏逻辑  --
//                        --       --
//                            --
//////////////////////////////////////////////////////////////////////////
package ShiShou510k

//import "fmt"

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	info2 "github.com/open-source/game/chess.git/services/sport/infrastructure/eve"
	logic2 "github.com/open-source/game/chess.git/services/sport/infrastructure/logic"
	meta2 "github.com/open-source/game/chess.git/services/sport/infrastructure/metadata"
	rule2 "github.com/open-source/game/chess.git/services/sport/infrastructure/rule"
	"math/rand"
	"strings"
	"time"
)

type SportLogicSS510K struct {
	//logic.BaseLogicDG
	gameTimer       components2.Time     //时间计时器
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

	m_playMode     int   //玩法模式
	MinOnestrCount uint8 //顺子的最小长度
	MinBombCount   uint8 //炸弹的最小长度
	MaxCardCount   uint8 //手牌最大数目
	MaxPlayerCount uint8 //游戏人数

	Rule rule2.St_FriendRule
}

//20181214 每个玩法设定自己的限时操作时间
func (slss *SportLogicSS510K) Setlimitetime(limitetimeOp bool) {
	//if limitetimeOp {
	//	slss.Rule.limitetimeOP = GameTime_Nine
	//} else {
	//	slss.Rule.limitetimeOP = 0
	//}
}

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

const (
	//纸牌出牌类型
	TYPE_ERROR       = -1    //错误的类型
	TYPE_NULL        = 0     //没有类型
	TYPE_ONE         = 10000 //单张
	TYPE_TWO         = 20000 //两张
	TYPE_THREE       = 30000 //三张
	TYPE_TWOSTR      = 40000 //连对 445566
	TYPE_THREESTR    = 50000 //三张的连对 444555
	TYPE_BOMB_510K   = 60000 //510k炸弹
	TYPE_BOMB_NOMORL = 70000 //普通炸弹
	TYPE_OTHR        = 80000 //其他
)

//炸弹对应分数
const (
	TYPE_BOMB5       = 200  //5张炸弹
	TYPE_BOMB6       = 500  //6张炸弹
	TYPE_BOMB7_DL    = 1500 //7张炸弹(双赖/经典模式)
	TYPE_BOMB7_SL    = 1000 //7张炸弹(四赖/经典模式)
	TYPE_SKYKINGBOMB = 2000 //四大天王(双王双赖)
	TYPE_BOMB8_DL    = 3000 //8张炸弹(双赖/经典模式)
	TYPE_BOMB8_SL    = 2000 //8张炸弹(双赖/经典模式)
	TYPE_BOMB9_DL    = 6000 //9张炸弹(双赖/经典模式)
	TYPE_BOMB9_SL    = 4000 //9张炸弹(双赖/经典模式)
	TYPE_BOMB10_DL   = 12000 //10张炸弹(双赖/经典模式)
	TYPE_BOMB10_SL   = 6000 //10张炸弹(双赖/经典模式)
	TYPE_BOMB11      = 8000 //11张炸弹
	TYPE_BOMB12      = 12000 //12张炸弹
)

//设置玩法模式
func (slss *SportLogicSS510K) SetPlayMode(byPlayMode int) {
	if byPlayMode >= 0 && byPlayMode <= 2 {
		slss.m_playMode = byPlayMode
	}
}

//基础函数，通过花色和point获取牌索引
func (slss *SportLogicSS510K) GetCard(byColor byte, byPoint byte) byte {
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
func (slss *SportLogicSS510K) GetCardColor(byCard byte) byte {
	//处理获取普通牌花色
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return (((byCard - 1) / 13) % 4) + 1
	} else if byCard == CARDINDEX_SMALL { //处理王牌的花色
		return byte(CC_NULL_S) //小王无花色
	} else if byCard == CARDINDEX_BIG {
		return byte(CC_NULL_S) //大王无花色
	} else if byCard == logic2.CARDINDEX_SKY {
		return byte(logic2.CC_NULL_S) //花牌无花色
	}
	return 0
}

//基础函数，通过牌索引获取点数
func (slss *SportLogicSS510K) GetCardPoint(byCard byte) byte {
	//处理获取普通牌点数
	if CARDINDEX_NULL < byCard && byCard < CARDINDEX_SMALL {
		return ((byCard - 1) % 13) + 1
	} else if byCard == CARDINDEX_SMALL { //处理王牌的点数
		return byte(CP_BJ_S) //小王
	} else if byCard == CARDINDEX_BIG {
		return byte(CP_RJ_S) //大王
	} else if byCard == logic2.CARDINDEX_SKY {
		return byte(logic2.CP_SKY_S) //花牌
	}
	return 0
}

//初始化赖子列表,byMagicPoint的值约束： A、2...K可以用1、2...13，但大小王只能用53和54，不能用14和15
func (slss *SportLogicSS510K) InitMagicPoint(byMagicPoint byte) {
	slss.m_allMagicPoint = []byte{}
	if byMagicPoint >= 1 && byMagicPoint <= logic2.CARDINDEX_SKY {
		slss.m_allMagicPoint = append(slss.m_allMagicPoint, byMagicPoint)
	}
}

//设置赖子值,byMagicPoint的值约束： A、2...K可以用1、2...13，但大小王只能用53和54，不能用14和15
func (slss *SportLogicSS510K) AddMagicPoint(byMagicPoint byte) {
	if (byMagicPoint >= 1) && (byMagicPoint <= CARDINDEX_SKY) {
		slss.m_allMagicPoint = append(slss.m_allMagicPoint, byMagicPoint)
	}
}

//设置顺子长度值，没有顺子时可以设置顺子最小长度为一个很大的值，比如254
func (slss *SportLogicSS510K) SetOnestrCount(byOnestrCount byte) {
	if (byOnestrCount >= 1) && (byOnestrCount <= 255) {
		slss.MinOnestrCount = byOnestrCount
	}
}

//设置炸弹的张数
func (slss *SportLogicSS510K) SetBombCount(byBombCount byte) {
	if (byBombCount >= 3) && (byBombCount <= 255) {
		slss.MinBombCount = byBombCount
	}
}

//设置手牌的最大张数
func (slss *SportLogicSS510K) SetMaxCardCount(byCardNum byte) {
	if (byCardNum >= 3) && (byCardNum <= 255) {
		slss.MaxCardCount = byCardNum
	} else {
		slss.MaxCardCount = static.MAX_CARD
	}
}

//设置最大人数
func (slss *SportLogicSS510K) SetMaxPlayerCount(byPlayerNum byte) {
	if (byPlayerNum >= 2) && (byPlayerNum <= 10) {
		slss.MaxPlayerCount = byPlayerNum
	} else {
		slss.MaxPlayerCount = static.MAX_PLAYER_4P
	}
}

func (slss *SportLogicSS510K) GetCardLevel(card_id byte) int {
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
	if card_id == CARDINDEX_SKY {
		level = 150
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
func (slss *SportLogicSS510K) GetScore(card_list [static.MAX_CARD]byte, cardlen int) int {
	var score int = 0
	for i := 0; i < cardlen && i < int(slss.MaxCardCount); i++ {
		if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(5) {
			score += 5
		} else if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(10) {
			score += 10
		} else if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(13) {
			score += 10
		}
	}
	return score
}

func (slss *SportLogicSS510K) SortByIndex(card_list [static.MAX_CARD]byte, cardlen int, smalltobig bool) [static.MAX_CARD]byte {
	if smalltobig {
		for i := 0; i < cardlen-1 && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
			for j := i + 1; j < cardlen && j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
				if card_list[j] > 0 {
					if slss.GetCardLevel(card_list[i]) > slss.GetCardLevel(card_list[j]) {
						temp := card_list[i]
						card_list[i] = card_list[j]
						card_list[j] = temp
					} else {
						if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(card_list[j]) {
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
		for i := 0; i < cardlen-1 && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
			for j := i + 1; j < cardlen && j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
				if card_list[j] > 0 {
					if slss.GetCardLevel(card_list[i]) < slss.GetCardLevel(card_list[j]) {
						temp := card_list[i]
						card_list[i] = card_list[j]
						card_list[j] = temp
					} else {
						if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(card_list[j]) {
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
func (slss *SportLogicSS510K) IsMagic(card_id byte) bool {
	for _, v := range slss.m_allMagicPoint {
		if slss.GetCardPoint(card_id) == slss.GetCardPoint(v) {
			return true
		}
	}
	return false
}

//当前牌是否是赖子
func (slss *SportLogicSS510K) IsKing(card_id byte) bool {
	if slss.GetCardPoint(card_id) == slss.GetCardPoint(CARDINDEX_SMALL) || slss.GetCardPoint(card_id) == slss.GetCardPoint(CARDINDEX_BIG) {
		return true
	}
	return false
}

//当前牌是否全是赖子
func (slss *SportLogicSS510K) IsAllKing(card_list [static.MAX_CARD]byte, cardlen int) bool {
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
		if card_list[i] > 0 && !slss.IsKing(card_list[i]) {
			return false
		}
	}
	return true
}

//当前牌是否除赖子外全部相同
func (slss *SportLogicSS510K) IsAllEqualExceptKing(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	if cardlen > static.MAX_CARD || cardlen > int(slss.MaxCardCount) {
		return 0, false
	}
	k := card_list[cardlen-1]
	for i := cardlen - 1; i >= 0; i-- {
		if slss.GetCardLevel(k) != slss.GetCardLevel(card_list[i]) {
			return 0, false
		}
	}
	if k == 0 {
		return 0, false
	}
	return k, true
}

//获取王的数量
func (slss *SportLogicSS510K) GetKingNum(card_list [static.MAX_CARD]byte, cardlen int) int {
	num := 0
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
		if card_list[i] > 0 && slss.IsKing(card_list[i]) {
			num++
		}
	}
	return num
}

func (slss *SportLogicSS510K) GetGroupType(myCard [static.MAX_CARD]byte) {
	slss.m_vOne = []static.TPokerGroup{}
	slss.m_vTwo = []static.TPokerGroup{}
	slss.m_vThree = []static.TPokerGroup{}
	slss.m_vFour = []static.TPokerGroup{}
	slss.m_vFive = []static.TPokerGroup{}
	slss.m_vSix = []static.TPokerGroup{}
	slss.m_vSeven = []static.TPokerGroup{}
	slss.m_vEight = []static.TPokerGroup{}
	slss.m_vKing = []static.TPokerGroup{}
	slss.m_vAllHandKing = []static.TPokerGroup{}
	slss.m_vMagic = []static.TPokerGroup{}
	slss.m_vAllHandMagic = []static.TPokerGroup{}
	slss.m_vK105 = []static.TPokerGroup{}

	var tempPoker [static.MAX_CARD]static.TPoker
	for i := 0; i < len(myCard); i++ {
		tempPoker[i].Set(myCard[i])
	}

	var group static.TPokerGroup
	//首先找王
	for i := 14; i < 15; i++ {
		for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
			if tempPoker[j].Point == byte(i) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				slss.m_vKing = append(slss.m_vKing, group)
			}
		}
	}
	//找赖子
	for i := 0; i < len(slss.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
			if tempPoker[j].Point == slss.GetCardPoint(slss.m_allMagicPoint[i]) {
				group.Point = tempPoker[j].Point
				group.Indexes = []uint8{}
				group.Indexes = append(group.Indexes, tempPoker[j].Index)
				group.Count = 1
				group.Color = 0
				slss.m_vMagic = append(slss.m_vMagic, group)
			}
		}
	}
	//找其他的
	pokerIndex := [13]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 1, 2}
	for i := 0; i < 13; i++ {
		//赖子不重复放，只放在m_vMagic和m_vAllHandMagic中
		bIsMagic := slss.IsMagic(pokerIndex[i])
		if bIsMagic {
			continue
		}

		num := byte(0)
		indexM := [10]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
			if tempPoker[j].Point == pokerIndex[i] {
				indexM[num] = tempPoker[j].Index
				num++
				if pokerIndex[i] == 5 || pokerIndex[i] == 10 || pokerIndex[i] == 13 {
					group.Point = pokerIndex[i]
					group.Color = tempPoker[j].Color
					group.Count = 1
					group.Indexes = []uint8{}
					group.Indexes = append(group.Indexes, tempPoker[j].Index)
					slss.m_vK105 = append(slss.m_vK105, group)
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
			slss.m_vEight = append(slss.m_vEight, group)
		} else if num == 7 {
			slss.m_vSeven = append(slss.m_vSeven, group)
		} else if num == 6 {
			slss.m_vSix = append(slss.m_vSix, group)
		} else if num == 5 {
			slss.m_vFive = append(slss.m_vFive, group)
		} else if num == 4 {
			slss.m_vFour = append(slss.m_vFour, group)
		} else if num == 3 {
			slss.m_vThree = append(slss.m_vThree, group)
		} else if num == 2 {
			slss.m_vTwo = append(slss.m_vTwo, group)
		} else if num == 1 {
			slss.m_vOne = append(slss.m_vOne, group)
		}
	}
	slss.m_allHandCards, slss.m_vAllHandMagic, slss.m_vAllHandKing = slss.GetGroupTypeByPoint(myCard)
}

//计算2连对和3连对、单连、三只、两只的函数。使用StrTool请确保GetGroupType函数被调用过，这样才能保证m_allHandCards 是最新的。
func (slss *SportLogicSS510K) StrTool(outpoint byte, iLianNum int, iZhaNum int) (bool, []static.TPokerGroup) {
	var combinelist []static.TPokerGroup
	//添加代码让它适用于顺子
	if (iZhaNum == 1 && (iLianNum < int(slss.MinOnestrCount) || iLianNum > 13)) || iZhaNum < 1 || iZhaNum > 12 || iLianNum < 1 || iLianNum > 12 {
		return false, nil
	}
	iMagicNum := len(slss.m_vAllHandMagic)
	if iMagicNum < 0 || iMagicNum > 8 {
		return false, nil
	}
	byOutPoint := slss.GetCardPoint(outpoint)
	if byOutPoint != 0 && (int(byOutPoint)-iLianNum+1 < 3 || (iLianNum > 1 && byOutPoint == 1) || (iLianNum <= 1 && byOutPoint == 2)) {
		//他自己已经通天了，别人不可能比它大,连时A最大，非连时(2只或3只)2最大
		return false, nil
	}
	if len(slss.m_allHandCards) < 24 {
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
		if slss.m_allHandCards[iIndex].Count == 0 && (iIndex+iLianNum-1 < 11) {
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
			for iCardCount := 0; iCardCount < iZhaNum && iCardCount < int(slss.m_allHandCards[idx].Count); iCardCount++ {
				card_list[cardlen] = slss.m_allHandCards[idx].Indexes[iCardCount]
				cardlen++
				byTmepHaveNum++
				if iZhaNum < int(slss.m_allHandCards[idx].Count) {
					sortRight += (int(slss.m_allHandCards[idx].Count) - iZhaNum) //拆牌要加权重
				}
			}
			byPoint = slss.m_allHandCards[idx].Point //直接获取最大的byPoint
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
				fakeking[byMagic].Index = slss.m_vAllHandMagic[byMagic].Indexes[0] //替换前的值,这个要保证是唯一的
				card_list[cardlen] = slss.m_vAllHandMagic[byMagic].Indexes[0]
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
				for faker := 0; faker < static.MAX_CARD && faker < int(slss.MaxCardCount); faker++ {
					tPKstr.Fakeking[faker] = fakeking[faker]
				}
				tPKstr.Cardtype = TYPE_THREESTR
				if iZhaNum == 2 {
					tPKstr.Cardtype = TYPE_TWOSTR
				} else if iZhaNum == 1 {
					tPKstr.Cardtype = static.TYPE_ONESTR
				}
				if iLianNum == 1 {
					tPKstr.Cardtype = TYPE_THREE
					if iZhaNum == 2 {
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
func (slss *SportLogicSS510K) GetCombineOneStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(slss.m_vAllHandMagic);

	if outlen < int(slss.MinOnestrCount) || outlen > 12 {
		return
	}
	_, combinelist = slss.StrTool(outpoint, outlen, 1)
	combinelist = slss.SortBeepCardList(combinelist)
	return
}

// 得到所有的两连对
func (slss *SportLogicSS510K) GetCombineTwoStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(slss.m_vAllHandMagic);

	if outlen%2 != 0 {
		return
	}
	_, combinelist = slss.StrTool(outpoint, outlen/2, 2)
	combinelist = slss.SortBeepCardList(combinelist)
	return
}
func (slss *SportLogicSS510K) GetCombineThreeStr(outpoint byte, outlen int) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(slss.m_vAllHandMagic);
	if outlen%3 != 0 {
		return
	}
	_, combinelist = slss.StrTool(outpoint, outlen/3, 3)
	//最后可以做一次优化，让没有王的先提示，有王的后提示
	combinelist = slss.SortBeepCardList(combinelist)
	return
} //end GetCombineThreeStr

//得到所有的对子
func (slss *SportLogicSS510K) GetCombineTwo(outpoint byte) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(slss.m_vAllHandMagic);
	_, combinelist = slss.StrTool(outpoint, 1, 2)
	combinelist = slss.SortBeepCardList(combinelist)
	return
}

//得到所有的三只
func (slss *SportLogicSS510K) GetCombineThree(outpoint byte) (combinelist []static.TPokerGroup) {
	//iMagicNum := len(slss.m_vAllHandMagic);
	_, combinelist = slss.StrTool(outpoint, 1, 3)
	combinelist = slss.SortBeepCardList(combinelist)
	return
}

//得到所有大过outpoint的单牌,先直接翻译后面再优化吧
func (slss *SportLogicSS510K) GetAllOne(outpoint byte) (combinelist []static.TPokerGroup) {
	var group static.TPokerGroup
	for it := 0; it < len(slss.m_vOne); it++ {
		// 过滤掉王提示
		if slss.GetCardLevel(slss.m_vOne[it].Point) == slss.GetCardLevel(53) || slss.GetCardLevel(slss.m_vOne[it].Point) == slss.GetCardLevel(54) {
			continue
		}

		if slss.GetCardLevel(slss.m_vOne[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vOne[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vOne[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}

	bBeepSKing := false
	bBeepBKing := false
	for it := 0; it < len(slss.m_vKing); it++ {
		// 一种花色的王之提示一次就可以// 提示单张小王
		if slss.GetCardLevel(slss.m_vKing[it].Point) > slss.GetCardLevel(outpoint) && slss.GetCardLevel(slss.m_vKing[it].Point) == slss.GetCardLevel(53) && !bBeepSKing {
			group.SortRight = 100 + len(slss.m_vKing) //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vKing[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vKing[it].Indexes[0])
			combinelist = append(combinelist, group)
			bBeepSKing = true
		}
		if slss.GetCardLevel(slss.m_vKing[it].Point) > slss.GetCardLevel(outpoint) && slss.GetCardLevel(slss.m_vKing[it].Point) == slss.GetCardLevel(54) && !bBeepBKing {
			group.SortRight = 100 + len(slss.m_vKing) //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vKing[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vKing[it].Indexes[0])
			combinelist = append(combinelist, group)
			bBeepBKing = true
		}
	}

	for it := 0; it < len(slss.m_vTwo); it++ {
		// 过滤掉对王
		if slss.GetCardLevel(slss.m_vTwo[it].Point) == slss.GetCardLevel(53) || slss.GetCardLevel(slss.m_vTwo[it].Point) == slss.GetCardLevel(54) {
			continue
		}

		if slss.GetCardLevel(slss.m_vTwo[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 1 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vTwo[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vTwo[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}

	for it := 0; it < len(slss.m_vThree); it++ {
		if slss.GetCardLevel(slss.m_vThree[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 2 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vThree[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vThree[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(slss.m_vFour); it++ {
		if slss.GetCardLevel(slss.m_vFour[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 3 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vFour[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vFour[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(slss.m_vFive); it++ {
		if slss.GetCardLevel(slss.m_vFive[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 4 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vFive[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vFive[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(slss.m_vSix); it++ {
		if slss.GetCardLevel(slss.m_vSix[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 5 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vSix[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vSix[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(slss.m_vSeven); it++ {
		if slss.GetCardLevel(slss.m_vSeven[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 6 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vSeven[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vSeven[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	for it := 0; it < len(slss.m_vEight); it++ {
		if slss.GetCardLevel(slss.m_vEight[it].Point) > slss.GetCardLevel(outpoint) {
			group.SortRight = 100 + 7 //排序权重，出牌优先级，基础值100
			group.Count = 1
			group.Point = slss.m_vEight[it].Point
			group.Indexes = []byte{}
			group.Indexes = append(group.Indexes, slss.m_vEight[it].Indexes[0])
			combinelist = append(combinelist, group)
		}
	}
	return
}

//提示出牌
func (slss *SportLogicSS510K) BeepCardOut(otherout [static.MAX_CARD]byte, outtype int) (breturn bool, mybeepout []static.TPokerGroup) {
	var group static.TPokerGroup
	group.Color = 0
	len1 := slss.GetCardNum(otherout, slss.MaxCardCount)
	cardtype1 := slss.GetType(otherout, int(len1), 0, 0, 0)

	cardtype1.Cardtype = outtype //重新设置下

	outlevel := slss.GetCardLevel(cardtype1.Card)
	if outtype == TYPE_NULL || outtype == TYPE_ERROR {
		return false, nil //不需要提示
		//} else if (outtype == TYPE_BOMB_FOUR_KING){
		//	// 四个王
		//	return false,nil;
	} else if outtype == TYPE_THREESTR {
		var templist []static.TPokerGroup
		templist = slss.GetCombineThreeStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == TYPE_TWOSTR {
		var templist []static.TPokerGroup
		templist = slss.GetCombineTwoStr(cardtype1.Card, cardtype1.Len)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
		//} else if outtype == TYPE_ONESTR {
		//	var templist []public.TPokerGroup
		//	templist = slss.GetCombineOneStr(cardtype1.Card, cardtype1.Len)
		//	for it := 0; it < len(templist); it++ {
		//		mybeepout = append(mybeepout, templist[it])
		//	}
	} else if outtype == TYPE_THREE {
		var templist []static.TPokerGroup
		templist = slss.GetCombineThree(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == TYPE_TWO {
		var templist []static.TPokerGroup
		templist = slss.GetCombineTwo(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	} else if outtype == TYPE_ONE {
		var templist []static.TPokerGroup
		templist = slss.GetAllOne(cardtype1.Card)
		for it := 0; it < len(templist); it++ {
			mybeepout = append(mybeepout, templist[it])
		}
	}

	//找拱拢(炸弹)
	mybombstr := slss.GetAllXiGL()
	for i := 0; i < len(mybombstr); i++ {
		//非拱笼类型的cardtype1.count =0,
		if mybombstr[i].BombLevel > cardtype1.BombLevel || (mybombstr[i].BombLevel == cardtype1.BombLevel && slss.GetCardLevel(mybombstr[i].Point) > outlevel) {
			group.Cardtype = TYPE_BOMB_NOMORL
			group.Indexes = []byte{}
			group.Count = mybombstr[i].Count
			group.Point = mybombstr[i].Point
			for j := 0; j < len(mybombstr[i].Indexes); j++ {
				group.Indexes = append(group.Indexes, mybombstr[i].Indexes[j])
			}
			for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
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
func (slss *SportLogicSS510K) BeepFirstCardOut() (breturn bool, mybeepout []static.TPokerGroup) {
	var group static.TPokerGroup

	var templist []static.TPokerGroup

	for it := 0; it < len(slss.m_vOne); it++ { //1
		group.Indexes = []byte{}
		group.Count = 1
		group.Point = slss.m_vOne[it].Point
		for j := 0; j < len(slss.m_vOne[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vOne[it].Indexes[j])
		}
		templist = append(templist, group)
	}
	for it := 0; it < len(slss.m_vTwo); it++ { //2
		group.Indexes = []byte{}
		group.Count = 2
		group.Point = slss.m_vTwo[it].Point
		for j := 0; j < len(slss.m_vTwo[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vTwo[it].Indexes[j])
		}
		templist = append(templist, group)
	}

	//3个以内的，都用最小的提示

	size := len(templist)
	for i := 0; i < size-1; i++ {
		for j := i + 1; j < size; j++ {
			if slss.GetCardLevel(templist[i].Point) > slss.GetCardLevel(templist[j].Point) && templist[j].Point > 0 {
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

	for it := 0; it < len(slss.m_vThree); it++ { //3
		group.Indexes = []byte{}
		group.Count = 3
		group.Point = slss.m_vThree[it].Point
		for j := 0; j < len(slss.m_vThree[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vThree[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(slss.m_vFour); it++ { //4
		group.Indexes = []byte{}
		group.Count = 4
		group.Point = slss.m_vFour[it].Point
		for j := 0; j < len(slss.m_vFour[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vFour[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(slss.m_vFive); it++ { //5
		group.Indexes = []byte{}
		group.Count = 5
		group.Point = slss.m_vFive[it].Point
		for j := 0; j < len(slss.m_vFive[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vFive[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(slss.m_vSix); it++ { //6
		group.Indexes = []byte{}
		group.Count = 6
		group.Point = slss.m_vSix[it].Point
		for j := 0; j < len(slss.m_vSix[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vSix[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(slss.m_vSeven); it++ { //7
		group.Indexes = []byte{}
		group.Count = 7
		group.Point = slss.m_vSeven[it].Point
		for j := 0; j < len(slss.m_vSeven[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vSeven[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	for it := 0; it < len(slss.m_vEight); it++ { //8
		group.Indexes = []byte{}
		group.Count = 8
		group.Point = slss.m_vEight[it].Point
		for j := 0; j < len(slss.m_vEight[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vEight[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}

	for it := 0; it < len(slss.m_vKing); it++ { //king
		group.Indexes = []byte{}
		group.Count = 1
		group.Point = slss.m_vKing[it].Point
		for j := 0; j < len(slss.m_vKing[it].Indexes); j++ {
			group.Indexes = append(group.Indexes, slss.m_vKing[it].Indexes[j])
		}
		mybeepout = append(mybeepout, group)
		return true, mybeepout
	}
	return false, nil
}

//获取所有同张的炸弹，比如软4个5 ，或2个王
func (slss *SportLogicSS510K) GetAllXiGL() (mybombstr []static.TBombStr) {
	iMagicNum := len(slss.m_vAllHandMagic) //癞子的数量
	if iMagicNum > 8 || iMagicNum < 0 {
		return
	}
	if len(slss.m_allHandCards) < 24 {
		return
	}

	for i := 0; i <= iMagicNum && i < 8; i++ {
		//全由赖子组成的炸弹后面处理、赖子数最多8个
		for j := 0; j < 13; j++ {
			itemsize := len(slss.m_allHandCards[j].Indexes)
			if itemsize == 0 {
				continue
			} //数目为0不能组成喜
			var item static.TBombStr
			item.Indexes = []byte{}
			if i+itemsize >= int(slss.MinBombCount) {
				//限定为大于等于3张
				var fakeking [static.MAX_CARD]static.TFakePoker
				for k := 0; k < itemsize && k < 8; k++ {
					//同一种牌最多8个，2副牌
					item.Indexes = append(item.Indexes, slss.m_allHandCards[j].Indexes[k])
				}
				for k := 0; k < i; k++ {
					//if (i != 8)//全由赖子组成的炸弹没有特殊性，如果有特殊性可以参考肉挨肉处理
					{
						fakeking[k].Fakeindex = slss.GetCardPoint(slss.m_allHandCards[j].Indexes[0]) //替换的值
						fakeking[k].Index = slss.m_vAllHandMagic[k].Indexes[0]                       //替换前的值,这个要保证是唯一的
					}
					item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[k].Indexes[0])
				}
				item.BombLevel = (i + itemsize) * 10 //3张的炸弹nBombCount为30,4张的炸弹为40,5张炸弹为50
				item.Point = slss.m_allHandCards[j].Point
				item.MaxCount = uint8(i + itemsize)
				item.Count = uint8(i + itemsize)
				for it := 0; it < static.MAX_CARD && it < int(slss.MaxCardCount); it++ {
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
	if slss.IsMagic(logic2.CARDINDEX_BIG) {
	} else {
		mybombKingstr = slss.GetCombineKingBomb()
		for i := 0; i < len(mybombKingstr); i++ {
			mybombstr = append(mybombstr, mybombKingstr[i])
		}
	}
	mybombK105str = slss.GetAllCombineK105()
	for i := 0; i < len(mybombK105str); i++ {
		mybombstr = append(mybombstr, mybombK105str[i])
	}
	return
}

//得到用双王炸弹
func (slss *SportLogicSS510K) GetCombineKingBomb() (mybombstr []static.TBombStr) {
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
	kingNum := len(slss.m_vAllHandKing) //王的数量
	if kingNum >= 2 {
		for i := 0; i < kingNum && i < 4; i++ {
			if slss.GetCardLevel(slss.m_vAllHandKing[i].Indexes[0]) == slss.GetCardLevel(CARDINDEX_SMALL) {
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

//是否5,10,K分牌
func (slss *SportLogicSS510K) isScorePai(card byte) (bRet bool) {
	if slss.GetCardLevel(card) == slss.GetCardLevel(5) {
		return true
	} else if slss.GetCardLevel(card) == slss.GetCardLevel(10) {
		return true
	} else if slss.GetCardLevel(card) == slss.GetCardLevel(13) {
		return true
	}
	return false
}

//得到所有的k-10-5
func (slss *SportLogicSS510K) GetAllCombineK105() (mybombstr []static.TBombStr) {
	if len(slss.m_vK105) == 0 {
		return
	}

	var tempK []static.TPoker
	var temp10 []static.TPoker
	var temp5 []static.TPoker
	for it := 0; it < len(slss.m_vK105); it++ {
		var temp static.TPoker
		if slss.GetCardLevel(slss.m_vK105[it].Point) == slss.GetCardLevel(13) {
			temp.Index = slss.m_vK105[it].Indexes[0]
			temp.Color = slss.m_vK105[it].Color
			tempK = append(tempK, temp)
		} else if slss.GetCardLevel(slss.m_vK105[it].Point) == slss.GetCardLevel(10) {
			temp.Index = slss.m_vK105[it].Indexes[0]
			temp.Color = slss.m_vK105[it].Color
			temp10 = append(temp10, temp)
		} else if slss.GetCardLevel(slss.m_vK105[it].Point) == slss.GetCardLevel(5) {
			temp.Index = slss.m_vK105[it].Indexes[0]
			temp.Color = slss.m_vK105[it].Color
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
				if slss.IsMagic(logic2.CARDINDEX_BIG) { //王是赖子
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
	if len(slss.m_vAllHandMagic) == 0 {
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
			item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[0].Indexes[0])
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if slss.IsMagic(logic2.CARDINDEX_BIG) { //王是赖子
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
			item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[0].Indexes[0])
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if slss.IsMagic(logic2.CARDINDEX_BIG) { //王是赖子
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
	if len(slss.m_vAllHandMagic) >= 3 {
		//3个王相当于正黑桃510k
		var item static.TBombStr
		item.Indexes = []byte{}
		item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[0].Indexes[0])
		item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[1].Indexes[0])
		item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[2].Indexes[0])
		//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
		if slss.IsMagic(logic2.CARDINDEX_BIG) { //王是赖子
			item.BombLevel = 30 + 4 + 1 //比4张的炸弹小一点，比杂510k大(+1)
		} else {
			item.BombLevel = 50 + 4 //比5张的炸弹大一点
		}
		item.Point = 0 //这个是无效的
		item.MaxCount = 3
		item.Count = 3
		mybombstr = append(mybombstr, item)
	}
	if len(slss.m_vAllHandMagic) >= 2 {
		//2个王的510k，当于正510k
		for it510k := 0; it510k != len(slss.m_vK105); it510k++ {
			byColor = slss.m_vK105[it510k].Color
			var item static.TBombStr
			item.Indexes = []byte{}
			item.Indexes = append(item.Indexes, slss.m_vK105[it510k].Indexes[0])
			item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[0].Indexes[0])
			item.Indexes = append(item.Indexes, slss.m_vAllHandMagic[1].Indexes[0])
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			if slss.IsMagic(logic2.CARDINDEX_BIG) { //王是赖子
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

func (slss *SportLogicSS510K) GetGroupTypeByPoint(myCard [static.MAX_CARD]byte) ([]static.TPokerGroup, []static.TPokerGroup, []static.TPokerGroup) {
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
		bIsMagic := slss.IsMagic(pokerIndex[i])
		if bIsMagic {
			continue
		}
		var group static.TPokerGroup
		group.Point = pokerIndex[i]
		group.Indexes = []uint8{}
		group.Color = 0
		group.Count = 0
		for j := 0; !bIsMagic && j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
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
		for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
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
	for i := 0; i < len(slss.m_allMagicPoint); i++ {
		for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
			if tempPoker[j].Point == slss.GetCardPoint(slss.m_allMagicPoint[i]) {
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

func (slss *SportLogicSS510K) GetAllMagics() []byte {
	var magics []byte
	//双赖模式加双王双花
	if slss.m_playMode == 1 {
		magics = append(magics, logic2.CARDINDEX_SKY, logic2.CARDINDEX_SKY)
	}

	//四赖模式加双赖
	if slss.m_playMode == 2 {
		for i := 104; i < logic2.MAX_POKER_COUNTS*2; i++ {
			magics = append(magics, logic2.CARDINDEX_SKY)
		}
	}

	return magics
}

func (slss *SportLogicSS510K) CreateCards() (byte, [static.ALL_CARD]byte) {
	cbCardDataTemp := [static.ALL_CARD]byte{}
	_maxCount := byte(static.ALL_CARD)
	//初始化所有牌点数
	//经典玩法去除大小王的所有牌
	count := logic2.MAX_POKER_COUNTS - 2
	for i := 0; i < count; i++ {
		for j := 0; j < 2; j++ {
			cbCardDataTemp[i+j*count] = byte(i) + 1
		}
	}
	//双赖模式加双王双花
	if slss.m_playMode == 1 {
		cbCardDataTemp[104] = logic2.CARDINDEX_BIG
		cbCardDataTemp[105] = logic2.CARDINDEX_BIG
		cbCardDataTemp[106] = logic2.CARDINDEX_SKY
		cbCardDataTemp[107] = logic2.CARDINDEX_SKY
	}

	//四赖模式加双赖
	if slss.m_playMode == 2 {
		for i := 104; i < logic2.MAX_POKER_COUNTS*2; i++ {
			cbCardDataTemp[i] = logic2.CARDINDEX_SKY
		}
	}

	return _maxCount, cbCardDataTemp
}

//是否同点数牌>=10张
func (slss *SportLogicSS510K) CheckSamePointAndMagic(card_list [static.MAX_CARD]byte) bool {
	for i := 1; i <= 13; i++ {
		num := 0
		for j := 0; j < static.MAX_CARD && j < int(slss.MaxCardCount); j++ {
			if slss.GetCardLevel(card_list[j]) == slss.GetCardLevel(byte(i)) || slss.IsMagic(card_list[j]) {
				num++
				//if (num == 8){break;}
			}
		}
		if num >= 10 {
			return true
		}
	}
	return false
}

//获取炸弹分
func (slss *SportLogicSS510K) getBombScore(cards_power int) int {
	switch cards_power {
	case 50:
		return TYPE_BOMB5
	case 60:
		return TYPE_BOMB6
	case 70:
		if slss.m_playMode == 2 {
			return TYPE_BOMB7_SL
		} else {
			return TYPE_BOMB7_DL
		}
	case 75:
		return TYPE_SKYKINGBOMB
	case 80:
		if slss.m_playMode == 2 {
			return TYPE_BOMB8_SL
		} else {
			return TYPE_BOMB8_DL
		}
	case 90:
		if slss.m_playMode == 2 {
			return TYPE_BOMB9_SL
		} else {
			return TYPE_BOMB9_DL
		}
	case 100:
		if slss.m_playMode == 2 {
			return TYPE_BOMB10_SL
		} else {
			return TYPE_BOMB10_DL
		}
	case 110:
		return TYPE_BOMB11
	case 120:
		return TYPE_BOMB12
	default:
		return 0
	}
	return 0
}

//处理积分，四舍五入
func (slss *SportLogicSS510K) DealScore(score int) int {
	temp := score % 100
	score /= 100
	if temp >= 50 {
		return score + 1
	} else if temp <= -50 {
		return score - 1
	}
	return score
}

//牌数量
func (slss *SportLogicSS510K) GetCardNum(card_list [static.MAX_CARD]byte, cardlen byte) byte {
	var iNum byte
	for i := byte(0); i < cardlen && i < static.MAX_CARD && i < slss.MaxCardCount; i++ {
		if card_list[i] > 0 && card_list[i] <= logic2.CARDINDEX_SKY {
			iNum++
		}
	}
	return iNum
}

//混乱扑克 并发牌
func (slss *SportLogicSS510K) RandCardData(byAllCards [static.ALL_CARD]byte, tableId int, superMan uint16, bigCards []byte, superPoint byte) ([meta2.MAX_PLAYER]byte, [meta2.MAX_PLAYER][static.MAX_CARD]byte, [meta2.MAX_PLAYER]int, int) {
	cbCardDataTemp := [meta2.MAX_PLAYER][static.MAX_CARD]byte{}
	_maxCount := [meta2.MAX_PLAYER]byte{}
	_KingCount := [meta2.MAX_PLAYER]int{}
	_spade3 := -1

	_randTmp := time.Now().Unix() + int64(tableId)
	rand.Seed(_randTmp)
	rand.Seed(int64(rand.Intn(10000)*10000) + time.Now().Unix())
	//洗牌
	pokercount := int(slss.MaxCardCount * 4)
	for i := 0; i < 1000; i++ {
		rand_num := rand.Intn(1000)
		m := rand_num % (pokercount)
		rand_num_2 := rand.Intn(1000)
		n := rand_num_2 % (pokercount)
		zz := byAllCards[m]
		byAllCards[m] = byAllCards[n]
		byAllCards[n] = zz
	}

	if superMan != static.INVALID_CHAIR && len(bigCards) > 0 { // 如果要作弊则换牌
		// 得到这个superman的发牌下标
		superManCardIdx := make([]int, 0, len(bigCards))
		superManCardIdxOther := make([]int, 0)

		for i := 0; i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
			for j := 0; j < MAX_PLAYER && j < int(slss.MaxPlayerCount); j++ {
				if slss.IsValidCard(byAllCards[int(slss.MaxPlayerCount)*i+j]) {
					if uint16(j) == superMan {
						// 把superman本来的牌的索引保存起来
						if len(superManCardIdx) < len(bigCards) {
							superManCardIdx = append(superManCardIdx, int(slss.MaxPlayerCount)*i+j)
						} else {
							superManCardIdxOther = append(superManCardIdxOther, int(slss.MaxPlayerCount)*i+j)
						}
					}
				}
			}
		}

		// 开始换牌
		for bIdx, bCard := range bigCards {
			// 从头开始换
			sIdx := superManCardIdx[bIdx]
			if byAllCards[sIdx] != bCard {
				// 在牌堆找这张牌
				for i := 0; i < len(byAllCards); i++ {
					var swapped bool
					for j := 0; j < bIdx; j++ {
						if i == superManCardIdx[j] {
							swapped = true
							break
						}
					}
					if !swapped {
						if byAllCards[i] == bCard { // 找到了
							byAllCards[sIdx], byAllCards[i] = byAllCards[i], byAllCards[sIdx]
							break
						}
					}
				}
			}
		}

		for _, sIdx := range superManCardIdxOther {
			preCard := byAllCards[sIdx]
			if preCard == CARDINDEX_SKY || slss.GetCardPoint(preCard) == superPoint {
				for i := 0; i < len(byAllCards); i++ {
					var swapped bool
					for j := 0; j < len(superManCardIdx); j++ {
						if i == superManCardIdx[j] {
							swapped = true
							break
						}
					}

					if !swapped {
						for j := 0; j < len(superManCardIdxOther); j++ {
							if i == superManCardIdxOther[j] {
								swapped = true
								break
							}
						}
					}

					if !swapped {
						if byAllCards[i] != CARDINDEX_SKY && slss.GetCardPoint(byAllCards[i]) != superPoint { // 找到了
							byAllCards[sIdx], byAllCards[i] = byAllCards[i], byAllCards[sIdx]
							break
						}
					}
				}
			}
		}
	}

	//清零
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		for j := 0; j < static.MAX_CARD; j++ {
			cbCardDataTemp[i][j] = 0
		}
	}

	//发牌
	for i := 0; i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
		for j := 0; j < meta2.MAX_PLAYER && j < int(slss.MaxPlayerCount); j++ {
			if slss.IsValidCard(byAllCards[int(slss.MaxPlayerCount)*i+j]) {
				cbCardDataTemp[j][i] = byAllCards[int(slss.MaxPlayerCount)*i+j]
				_maxCount[j]++
				if cbCardDataTemp[j][i] == logic2.CARDINDEX_SMALL || cbCardDataTemp[j][i] == logic2.CARDINDEX_BIG {
					_KingCount[j]++
				}
				if _spade3 == -1 && cbCardDataTemp[j][i] == 42 {
					_spade3 = j
				}
			}
		}
	}

	return _maxCount, cbCardDataTemp, _KingCount, _spade3
}

//牌有效判断
func (slss *SportLogicSS510K) IsValidCard(cbCardData byte) bool {
	//校验
	if cbCardData == 0 || cbCardData > logic2.CARDINDEX_SKY {
		return false
	}
	return true
}

func (slss *SportLogicSS510K) Compare(typeFirst static.TCardType, typeFollow static.TCardType) bool {
	if typeFirst.Cardtype == TYPE_ERROR {
		return false
	}
	if typeFollow.Cardtype == TYPE_ERROR || typeFollow.Cardtype == TYPE_NULL {
		return false
	}

	//第一种情况，首出
	if typeFirst.Cardtype == TYPE_NULL {
		if typeFollow.Cardtype == TYPE_ONE || typeFollow.Cardtype == TYPE_TWO || typeFollow.Cardtype == TYPE_TWOSTR || typeFollow.Cardtype == TYPE_BOMB_510K || typeFollow.Cardtype == TYPE_BOMB_NOMORL {
			return true
		} else if typeFollow.Cardtype == TYPE_THREE || typeFollow.Cardtype == TYPE_THREESTR {
			//如果王是赖子：大冶王是赖子，蕲春王不是赖子，通过王是赖子来区分是否是大冶打拱，大冶有TYPE_THREE和TYPE_THREESTR
			return true
		} else {
			return false
		}
	} else { //第二种情况，跟出的人出的不是炸弹，类型必须和首出一致//其他牌的比较，非炸弹
		if (typeFollow.Cardtype == TYPE_BOMB_510K) || (typeFollow.Cardtype == TYPE_BOMB_NOMORL) {
			//非拱笼类型的BombLevel =0,
			if typeFollow.BombLevel > typeFirst.BombLevel || (typeFollow.BombLevel == typeFirst.BombLevel && slss.GetCardLevel(typeFollow.Card) > slss.GetCardLevel(typeFirst.Card)) {
				return true
			} else {
				return false
			}
		} else if typeFollow.Cardtype == typeFirst.Cardtype { //跟出的人出的不是炸弹，类型必须和首出一致
			if typeFollow.Len == typeFirst.Len && slss.GetCardLevel(typeFollow.Card) > slss.GetCardLevel(typeFirst.Card) {
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

func (slss *SportLogicSS510K) GetType(card_list [static.MAX_CARD]byte, cardlen int, outMagicNum byte, byType byte, outtype int) static.TCardType {
	var re static.TCardType
	re.Len = 0
	re.Card = 0
	re.Color = 0
	re.Cardtype = TYPE_NULL
	re.Count = 0
	re.BombLevel = 0

	card := byte(0)
	slss.SortByIndex(card_list, cardlen, true)
	re.Len = int(slss.GetCardNum(card_list, byte(cardlen)))
	if re.Len < 1 {
		return re
	}
	re.Card = card_list[re.Len-1]
	switch re.Len {
	case 0:
		re.Cardtype = TYPE_NULL
		return re
	case 1:
		re.Cardtype = TYPE_ONE
		//癞子不能单出
		if slss.IsMagic(card_list[0]) {
			re.Cardtype = TYPE_ERROR
		}
		////如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//if(slss.IsMagic(CARDINDEX_BIG) && slss.IsKing(re.Card)){
		//	re.Card = 2;//王单出算2
		//}
		return re
	case 2:
		//if !slss.IsKing(card_list[0])&&!slss.IsKing(card_list[1]){
		if slss.IsMagic(card_list[0]) && slss.IsMagic(card_list[1]) { //全是癞子
			re.Cardtype = TYPE_ERROR
		} else if slss.GetCardLevel(card_list[0]) == slss.GetCardLevel(card_list[1]) {
			re.Cardtype = TYPE_TWO
		} else if slss.IsMagic(card_list[0]) || slss.IsMagic(card_list[1]) { //有一个癞子
			re.Card = card_list[1]
			if slss.IsMagic(card_list[1]) {
				re.Card = card_list[0]
			}
			re.Cardtype = TYPE_TWO
		} else {
			re.Cardtype = TYPE_ERROR
		}
		//}else if (slss.IsKing(card_list[0]) && slss.IsKing(card_list[1])){
		////如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//if slss.IsMagic(CARDINDEX_BIG) {
		//	re.Card = 2;//大小王按2来算
		//	re.Cardtype = TYPE_TWO;
		//}else {
		//	//对王
		//	if (card_list[0] == card_list[1]) {
		//		if (card_list[0] == CARDINDEX_SMALL) {
		//			re.BombLevel = 60 + 1; //对小王，比6张的炸弹大一些
		//		} else
		//		{
		//			re.BombLevel = 60 + 1; //对大王，比6张的炸弹大一些
		//		}
		//	} else{
		//		// 大小王组成的一对
		//		re.BombLevel = 60 + 1;
		//	}
		//	re.Cardtype = TYPE_BOMB_DOUBLE_KING;
		//}
		//re.Cardtype = TYPE_TWO;
		//}else{
		//	//有一张王
		//	//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//	if slss.IsMagic(CARDINDEX_BIG) {
		//		re.Card = card_list[1];
		//		if slss.IsMagic(card_list[1]) {
		//			re.Card = card_list[0];
		//		}
		//		re.Cardtype = TYPE_TWO;
		//	}else {
		//		re.Cardtype = TYPE_ERROR;
		//	}
		//}
		return re
	case 3:
		//全是癞子不能出
		if slss.IsAllMagic(card_list, re.Len) {
			re.Cardtype = TYPE_ERROR
			return re
		}

		color := byte(0)
		is510k := false
		color, is510k = slss.Is510KBomb(card_list, 3)
		isBombNormal := false
		card, isBombNormal = slss.IsNormalBomb(card_list, cardlen)
		if is510k {
			re.Color = color
			re.Cardtype = TYPE_BOMB_510K
			//通过王是不是赖子来区分大冶和蕲春的(蕲春510k比4张炸弹大、王不是赖子；大冶510k比4张炸弹小、王是赖子)，后期还有另一种情况时建议用其他方式来区分
			//if slss.IsMagic(CARDINDEX_BIG) {//王是赖子
			re.Card = 5
			if color == 0 {
				re.BombLevel = 30 + 1 //比4张的炸弹小一点
			} else {
				//正510k不比大小
				re.BombLevel = 30 + 2 //比杂510k大一点
				//re.BombLevel = 30 + int(color) + 1;//比4张的炸弹小一点，比杂510k大(+1)
			}
			//}else {
			//	if(color == 0){
			//		re.BombLevel = 40 + 1;//比4张的炸弹大一点
			//	} else{
			//		re.BombLevel = 50 + int(color);//比5张的炸弹大一点
			//	}
			//}
			//} else if (slss.IsKing(card_list[0]) && slss.IsKing(card_list[1]) && slss.IsKing(card_list[2])){
			//	//三个王
			//	//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
			//	if slss.IsMagic(CARDINDEX_BIG) {
			//		re.Color = 4;//3个大小王按黑桃正510k来算
			//		re.BombLevel = 30 + 4+1; //比4张的炸弹小一些，比杂510k大(+1)
			//		re.Cardtype = TYPE_BOMB_510K;
			//		re.Card = 5;//3个大小王按黑桃正510k来算
			//		//re.Card = 3;//3个大小王按3来算
			//		//re.Cardtype = TYPE_THREE;
			//	}else {
			//		re.BombLevel = 60 + 4; //比6张的炸弹大一些
			//		re.Cardtype = TYPE_BOMB_DOUBLE_KING;
			//	}
			//	监利开机三个王不能出
			//	re.Cardtype = TYPE_ERROR;
		} else if isBombNormal {
			re.BombLevel = re.Len * 10
			re.Card = card
			re.Cardtype = TYPE_BOMB_NOMORL
		} else {
			//ONESTR
			//没有顺子类型，判断是否三张
			cpoint, isThree := slss.IsAllEqualExceptMagic(card_list, cardlen)
			if isThree {
				re.Cardtype = TYPE_THREE
				re.Card = cpoint
			} else {
				re.Cardtype = TYPE_ERROR
			}
			//var refakepoker []public.TFakePoker
			//iMagicNum := 0
			//iMagicNum ,refakepoker= slss.GetMagicNum(card_list,cardlen,refakepoker)
			//if(byType != 0 ) {//非0表示没有赖子模式
			//	iMagicNum= 0
			//}
			//if(iMagicNum <= 8) {//有1-4个王
			//	var ty public.FakeType
			//	ty = slss.GetTypeByMagic(card_list,cardlen,iMagicNum,outtype);
			//	re.Cardtype = ty.CardType.Cardtype
			//	re.Card = ty.CardType.Card;
			//} else {
			//	re.Cardtype = TYPE_ERROR;
			//}
		}
		return re
	default: // >= 4张
		//if (re.Len == 4) {
		//	card = byte(0)
		//	if (slss.IsKingBomb(card_list, cardlen)) {//四个王最大
		//		//如果王是赖子，大冶打拱王是赖子，蕲春王不是赖子
		//		if slss.IsMagic(logic.CARDINDEX_BIG) {
		//			re.BombLevel = 40;//4个王就是4张的炸弹
		//			re.Card = 2;//4个王就是4个2
		//			re.Cardtype = TYPE_BOMB_NOMORL;
		//		} else{
		//			//re.BombLevel = 60 + 2000;
		//			re.BombLevel = 80 + 1;//比7张的炸弹大些，比8张的炸弹小
		//			re.Cardtype = TYPE_BOMB_FOUR_KING;
		//		}
		//		return re
		//	}
		//}

		//全是癞子不能出
		if slss.IsAllMagic(card_list, re.Len) {
			re.Cardtype = TYPE_ERROR
			return re
		}

		card = byte(0)
		isBombNormal := false
		card, isBombNormal = slss.IsNormalBomb(card_list, cardlen)
		if isBombNormal {
			re.Card = card
			//if (re.Len == 8) {
			//	re.Cardtype = TYPE_BOMB_8XI;
			//} else {
			//	re.Cardtype = TYPE_BOMB_NOMORL;
			//}
			re.Cardtype = TYPE_BOMB_NOMORL
			re.BombLevel = re.Len * 10

			//若为双王加双赖
			if slss.IsKing(card) {
				re.BombLevel = 75 //大于七炸小于八炸
				//re.Cardtype = TYPE_BOMB_SKY_KING
			}
			return re
		}
		//单顺，双顺等
		var refakepoker []static.TFakePoker
		iMagicNum := 0
		iMagicNum, refakepoker = slss.GetMagicNum(card_list, cardlen, refakepoker)
		if byType != 0 { //非0表示没有赖子模式，给服务器用
			iMagicNum = 0
		}
		if iMagicNum <= 8 { //有1-4个王
			var ty static.FakeType
			ty = slss.GetTypeByMagic(card_list, cardlen, iMagicNum, outtype)
			re.Cardtype = ty.CardType.Cardtype
			re.Card = ty.CardType.Card
		} else {
			re.Cardtype = TYPE_ERROR
		}
		return re
	}
	return re
}

//判断是不是普通炸弹，如果是，返回true 对王炸、510k炸弹分开来
// 普通炸弹单纯的指的是三张以上相同的牌
func (slss *SportLogicSS510K) IsNormalBomb(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	max_bombCnt := 8
	if slss.IsMagic(logic2.CARDINDEX_SKY) {
		max_bombCnt = 12 //4个王+8张同样的牌组成的炸弹
	}
	if cardlen > max_bombCnt || cardlen < int(slss.MinBombCount) { //3到8张同样的牌组成的炸弹
		return 0, false
	}
	p, bBomb := slss.IsAllEqualExceptMagic(card_list, cardlen)
	if bBomb {
		return p, true
	} else {

	}
	return 0, false
}

//当前牌是否全是赖子
func (slss *SportLogicSS510K) IsAllMagic(card_list [static.MAX_CARD]byte, cardlen int) bool {
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
		if card_list[i] > 0 && !slss.IsMagic(card_list[i]) {
			return false
		}
	}
	return true
}

//当前牌是否除赖子外全部相同
func (slss *SportLogicSS510K) IsAllEqualExceptMagic(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	if cardlen > static.MAX_CARD || cardlen > int(slss.MaxCardCount) {
		return 0, false
	}
	k := byte(0)
	for i := cardlen - 1; i >= 0; i-- {
		if slss.IsMagic(card_list[i]) {
			continue
		}
		k = card_list[i]
		break
	}
	if k == 0 {
		return 0, false
	}
	for i := cardlen - 1; i >= 0; i-- {
		if slss.IsMagic(card_list[i]) {
			continue
		}
		if slss.GetCardLevel(k) != slss.GetCardLevel(card_list[i]) {
			return 0, false
		}
	}
	return k, true
}

//获取赖子的数量
func (slss *SportLogicSS510K) GetMagicNum(card_list [static.MAX_CARD]byte, cardlen int, fakepoker []static.TFakePoker) (int, []static.TFakePoker) {
	num := 0
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
		if card_list[i] > 0 && slss.IsMagic(card_list[i]) {
			if fakepoker != nil {
				fakepoker[num].Index = card_list[i]
			}
			num++
		}
	}
	return num, fakepoker
}

//判断是不是510k，如果是，返回花色1-4代表了方块，梅花，红桃，黑桃,0代表杂的
func (slss *SportLogicSS510K) Is510KBomb(card_list [static.MAX_CARD]byte, cardlen int) (byte, bool) {
	color := byte(0)
	if cardlen != 3 {
		return color, false
	}
	has_5 := 0
	has_10 := 0
	has_k := 0
	has_m := 0 //赖子数目
	for i := 0; i < cardlen && i < static.MAX_CARD && i < int(slss.MaxCardCount); i++ {
		if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(5) {
			has_5++
		} else if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(10) {
			has_10++
		} else if slss.GetCardLevel(card_list[i]) == slss.GetCardLevel(13) {
			has_k++
		} else if slss.IsMagic(card_list[i]) {
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

//有赖子的情况下（赖子数可以为0），得出他的牌型，这个只针对三连对或者2连对，超过4只的可以参考通山打拱的拱笼
func (slss *SportLogicSS510K) IsStrByMagic(cardlen int, iMagicNum int, iZhaNum int) static.FakeType {
	var temptype static.FakeType
	if len(slss.m_allCard) < 24 || (cardlen > iZhaNum && slss.m_allCard[12].Count != 0) || iZhaNum < 1 || iZhaNum > 12 || cardlen%iZhaNum != 0 {
		//有2，不可能组成连对
		temptype.CardType.Cardtype = TYPE_ERROR
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
		if iLianNum == 0 && slss.m_allCard[iIndex].Count == 0 {
			continue
		}
		iLianNum++
		//构造数据
		for iCardCount := 0; iCardCount < iZhaNum && byte(iCardCount) < slss.m_allCard[iIndex].Count; iCardCount++ {
			templist[iNewcardlen] = slss.m_allCard[iIndex].Indexes[iCardCount]
			iNewcardlen++
			temptype.CardType.Card = slss.m_allCard[iIndex].Point //存放最大点数，暂时赋值
		}
		if slss.m_allCard[iIndex].Count <= byte(iZhaNum) && byLeftMagicNum >= iZhaNum-int(slss.m_allCard[iIndex].Count) {
			byNeedMagicNum := iZhaNum - int(slss.m_allCard[iIndex].Count)
			byLeftMagicNum -= byNeedMagicNum
			for byMagic := 0; byMagic < byNeedMagicNum && byLeftMagicNum >= 0; byMagic++ {
				templist[iNewcardlen] = slss.m_vMagic[iUseMagicCount].Indexes[0]
				iNewcardlen++
				temptype.CardType.Card = slss.m_allCard[iIndex].Point //存放最大点数，暂时赋值
				//temptype.fakeking[iUseMagicCount].fakeindex = iIndex;//代替的值
				//temptype.fakeking[iUseMagicCount].index = fakepoker[iUseMagicCount++].index;//存放原来的值，就是王
				temptype.Fakeking[iUseMagicCount].Fakeindex = slss.m_allCard[iIndex].Point         //代替的值
				temptype.Fakeking[iUseMagicCount].Index = slss.m_vMagic[iUseMagicCount].Indexes[0] //存放原来的值，就是赖子
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
		temptype.CardType.Cardtype = TYPE_THREESTR
		if iZhaNum == 2 {
			temptype.CardType.Cardtype = TYPE_TWOSTR
			//} else if iZhaNum == 1 {
			//	temptype.CardType.Cardtype =TYPE_ONESTR
		}
		if cardlen <= 3 && static.TYPE_ONESTR != temptype.CardType.Cardtype {
			temptype.CardType.Cardtype = TYPE_THREE
			if iZhaNum == 2 {
				temptype.CardType.Cardtype = TYPE_TWO
			} else if iZhaNum == 1 {
				temptype.CardType.Cardtype = TYPE_ONE
			}
		}
	} else {
		temptype.CardType.Cardtype = TYPE_ERROR
	}
	return temptype
}

//有赖子的情况下（赖子数可以为0），得出他的牌型，这个只针对三连对或者2连对，超过4只的可以参考通山打拱的拱笼
func (slss *SportLogicSS510K) GetTypeByMagic(card_list [static.MAX_CARD]byte, cardlen int, iMagicNum int, outtype int) static.FakeType {
	var reType static.FakeType
	if iMagicNum >= 0 {
		var vTypeList []static.FakeType
		slss.m_allCard, slss.m_vMagic, slss.m_vKing = slss.GetGroupTypeByPoint(card_list)
		if cardlen >= int(slss.MinOnestrCount) { //单顺最小长度在这里判断一下，单张在gettype里面判断
			var temptypeOne static.FakeType
			temptypeOne = slss.IsStrByMagic(cardlen, iMagicNum, 1)
			if temptypeOne.CardType.Cardtype != TYPE_ERROR {
				vTypeList = append(vTypeList, temptypeOne)
			}
		}
		var temptypeTwo static.FakeType
		temptypeTwo = slss.IsStrByMagic(cardlen, iMagicNum, 2)
		if temptypeTwo.CardType.Cardtype != TYPE_ERROR {
			vTypeList = append(vTypeList, temptypeTwo)
		}
		var temptypeThree static.FakeType
		temptypeThree = slss.IsStrByMagic(cardlen, iMagicNum, 3)
		if temptypeThree.CardType.Cardtype != TYPE_ERROR {
			vTypeList = append(vTypeList, temptypeThree)
		}

		if len(vTypeList) > 0 {
			var max_re static.FakeType
			if outtype == TYPE_THREESTR || outtype == TYPE_TWOSTR || outtype == TYPE_THREE || outtype == TYPE_TWO || outtype == TYPE_ONE { //说明手先出的人已经指定了出牌类型，那么跟的就是出牌类型
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
						if slss.Compare(max_re.CardType, vTypeList[i].CardType) {
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
						if slss.Compare(max_re.CardType, vTypeList[i].CardType) {
							max_re = vTypeList[i]
						}
					}
				}
			}
			reType = max_re
		} else {
			reType.CardType.Cardtype = TYPE_ERROR
		}
	} else {
		reType.CardType.Cardtype = TYPE_ERROR
	}
	return reType
}

//出牌排序
func (slss *SportLogicSS510K) SortBeepCardList(combinelist []static.TPokerGroup) []static.TPokerGroup {
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

func (slss *SportLogicSS510K) GetStringByCard(carddata byte) string {
	strColor := ""
	strPoint := ""

	color := slss.GetCardColor(carddata)
	point := slss.GetCardPoint(carddata)

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

func (slss *SportLogicSS510K) GetWriteHandReplayRecordString(replayRecord SS510k_Replay_Record) string {
	handCardStr := ""
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d:", i)
		for j := 0; j < len(replayRecord.R_HandCards[i]); j++ {
			handCardStr += fmt.Sprintf("%s,", slss.GetStringByCard(byte(replayRecord.R_HandCards[i][j])))
		}
	}

	//写入分数
	handCardStr += "S:"
	for i := 0; i < meta2.MAX_PLAYER; i++ {
		handCardStr += fmt.Sprintf("%d,", replayRecord.R_Score[i])
	}

	return handCardStr
}

func (slss *SportLogicSS510K) GetWriteOutReplayRecordString(replayRecord SS510k_Replay_Record) string {
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
		case DG_REPLAY_OPT_HOUPAI:
			ourCardStr += fmt.Sprintf("|H%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_QIANG:
			ourCardStr += fmt.Sprintf("|Q%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_END_HOUPAI:
			ourCardStr += fmt.Sprintf("|E%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_END_GAME:
			ourCardStr += fmt.Sprintf("|G%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_DIS_GAME:
			ourCardStr += fmt.Sprintf("|J%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_OUTCARD:
			if len(record.R_Value) > 0 {
				ourCardStr += fmt.Sprintf("|C%d:", record.R_ChairId)
			} else {
				ourCardStr += fmt.Sprintf("|C%d", record.R_ChairId)
			}
			break
		case DG_REPLAY_OPT_TURN_OVER:
			if len(record.R_Opt_Ext) > 0 {
				ourCardStr += fmt.Sprintf("|T%d:", record.R_ChairId)
			} else {
				ourCardStr += fmt.Sprintf("|T%d", record.R_ChairId)
			}
			break
		case DG_REPLAY_OPT_TUOGUAN:
			ourCardStr += fmt.Sprintf("|D%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_4KINGSCORE:
			ourCardStr += fmt.Sprintf("|K%d:", record.R_ChairId)
			break
		case DG_REPLAY_OPT_RESTART:
			ourCardStr += fmt.Sprintf("|R%d:", record.R_ChairId)
			break
		default:
			break
		}

		if len(record.R_Value) > 0 {
			for i := 0; i < len(record.R_Value); i++ {
				ourCardStr += fmt.Sprintf("%s", slss.GetStringByCard(byte(record.R_Value[i])))
			}
		}

		//打出的分牌
		if len(record.R_ScoreCard) > 0 {
			fakeStr := ""
			for i := 0; i < len(record.R_ScoreCard); i++ {
				fakeStr += fmt.Sprintf("%s", slss.GetStringByCard(byte(record.R_ScoreCard[i])))
			}
			ourCardStr += fmt.Sprintf(",F%s", fakeStr)
		}

		if len(record.R_Opt_Ext) > 0 {
			for i := 0; i < len(record.R_Opt_Ext); i++ {
				if record.R_Opt_Ext[i].Ext_type == DG_EXT_HOUPAI {
					ourCardStr += fmt.Sprintf(",H%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_QIANG {
					ourCardStr += fmt.Sprintf(",Q%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_TURNSCORE {
					ourCardStr += fmt.Sprintf(",S%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_GETSCORE {
					ourCardStr += fmt.Sprintf(",G%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_MINGJI {
					ourCardStr += fmt.Sprintf(",J")
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_ENDQIANG {
					ourCardStr += fmt.Sprintf(",B%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_CARDTYPE {
					ourCardStr += fmt.Sprintf(",T%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_TUOGUAN {
					ourCardStr += fmt.Sprintf(",D%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_4KINGSCORE {
					ourCardStr += fmt.Sprintf(",K%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_RESTART {
					ourCardStr += fmt.Sprintf(",R%d", record.R_Opt_Ext[i].Ext_value)
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_CURSCORE {
					for key, value := range record.R_Opt_Ext[i].Ext_score {
						if key > 1 {
							break
						}
						ourCardStr += fmt.Sprintf(",%d:L%d", key, value)
					}
				} else if record.R_Opt_Ext[i].Ext_type == DG_EXT_TOTALSCORE {
					for key, value := range record.R_Opt_Ext[i].Ext_score {
						if key > 1 {
							break
						}
						ourCardStr += fmt.Sprintf(",%d:M%d", key, value)
					}
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
