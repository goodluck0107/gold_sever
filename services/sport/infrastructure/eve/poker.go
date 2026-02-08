package eve

//游戏结束类型
const (
	POKER_END_LOSE        = 0 //游戏输
	POKER_END_WIN         = 1 //游戏赢
	POKER_END_NOWINNOLOSE = 2 //游戏平局
)

//统一扑克花色
const (
	CARD_C_NULL    byte = iota // value --> 0  无色
	CARD_C_DIAMOND             // value --> 1  方块
	CARD_C_CLUB                // value --> 2  梅花
	CARD_C_HEART               // value --> 3  红心
	CARD_C_SPADE               // value --> 4  黑桃
	CARD_C_TOTAL               // value --> 5
)

//统一扑克点数
const (
	CARD_P_NULL       byte = iota // value --> 0 牌背
	CARD_P_A                      // value --> 1 A
	CARD_P_2                      // value --> 2
	CARD_P_3                      // value --> 3
	CARD_P_4                      // value --> 4
	CARD_P_5                      // value --> 5
	CARD_P_6                      // value --> 6
	CARD_P_7                      // value --> 7
	CARD_P_8                      // value --> 8
	CARD_P_9                      // value --> 9
	CARD_P_10                     // value --> 10
	CARD_P_J                      // value --> 11 J
	CARD_P_Q                      // value --> 12 Q
	CARD_P_K                      // value --> 13 K
	CARD_P_SMALL_KING             // value --> 14 小王
	CARD_P_BIG_KING               // value --> 15 大王
	CARD_P_SKY_KING               // value --> 16 天牌（花牌）
	CARD_P_TOTAL                  // value --> 17
)

//统一扑克索引
const (
	CARD_INDEX_SMALL_KING = 53 // 小王牌索引
	CARD_INDEX_BIG_KING   = 54 // 大王牌索引
	CARD_INDEX_SKY_KING   = 55 // 天牌（花牌）索引
	CARD_INDEX_BACK_CARD  = 56 // 背面牌索引
	CARD_INDEX_INVALID    = 0  // 无效牌索引
)

//统一定义出牌类型
const (
	CARD_TYPE_ERROR          = -1
	CARD_TYPE_NULL           = 0
	CARD_TYPE_ONE            = 1
	CARD_TYPE_TWO            = 2
	CARD_TYPE_ONESTR         = 3
	CARD_TYPE_TWOSTR         = 4
	CARD_TYPE_THREE          = 5
	CARD_TYPE_FEIJI          = 6
	CARD_TYPE_4DAI2          = 7
	CARD_TYPE_BOMB_510K      = 30
	CARD_TYPE_BOMB_NOMORL    = 40
	CARD_TYPE_BOMB_KING_3821 = 45
	CARD_TYPE_BOMB_2KINGS    = 50
	CARD_TYPE_BOMB_3KINGS    = 60
	CARD_TYPE_BOMB_4KINGS    = 70
	CARD_TYPE_BOMB_5KINGS    = 80
	CARD_TYPE_BOMB_6KINGS    = 90
	CARD_TYPE_BOMB_SKY_KING  = 100
)

//炸弹等级
const (
	BOMB_3_LEVEL              = 30
	BOMB_4_LEVEL              = 40
	BOMB_510K_LEVEL           = 44
	BOMB_S510K_LEVEL          = 45
	BOMB_5_LEVEL              = 50
	BOMB_SMALL_KING3821_LEVEL = 51
	BOMB_KING3821_LEVEL       = 52
	BOMB_2DKING_LEVEL         = 55
	BOMB_6_LEVEL              = 60
	BOMB_2SSKING_LEVEL        = 64
	BOMB_2SBKING_LEVEL        = 66
	BOMB_7_LEVEL              = 70
	BOMB_3KING_LEVEL          = 75
	BOMB_8_LEVEL              = 80
	BOMB_4KING_LEVEL          = 85
	BOMB_9_LEVEL              = 90
	BOMB_5KING_LEVEL          = 95
	BOMB_10_LEVEL             = 100
	BOMB_6KING_LEVEL          = 105
	BOMB_11_LEVEL             = 110
	BOMB_12_LEVEL             = 120
	BOMB_13_LEVEL             = 130
	BOMB_14_LEVEL             = 140
	BOMB_15_LEVEL             = 150
	BOMB_16_LEVEL             = 160
	BOMB_17_LEVEL             = 170
	BOMB_18_LEVEL             = 180
	BOMB_2SKYKING_LEVEL       = 185
)
