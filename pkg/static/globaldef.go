package static

import "fmt"

const (
	KIND_ID_CY         = 897 //崇阳麻将
	KIND_ID_CYFF       = 896 //崇阳放放
	KIND_ID_JM         = 981 //荆门晃晃
	KIND_ID_Test       = 123 //崇阳测试专供
	KIND_ID_Test1      = 124 //崇阳放放测试专供
	KIND_ID_JMSK       = 963 //chess麻将荆门双开
	KIND_ID_HZLZG      = 984 //chess麻将红中赖子杠
	KIND_ID_JZJL       = 599 //新江陵玩法
	KIND_ID_XTMJ_LH    = 598 //仙桃麻将赖晃
	KIND_ID_XTMJ_YLDD  = 597 //仙桃麻将一赖到底
	KIND_ID_XTMJ_YLKZC = 596 //仙桃麻将一赖可捉统
	KIND_ID_JZMJ_YJLY  = 595 //荆州麻将一脚赖油
	KIND_ID_HHHJ_YH    = 594 //晃晃合集硬晃
	KIND_ID_HHHJ_LH    = 593 //晃晃合计赖晃
	KIND_ID_HHHJ_YLKZC = 592 //晃晃合集一赖可捉统
	KIND_ID_CBMJ_DD    = 591 //赤壁麻将剁刀
	KIND_ID_CBMJ_YH    = 590 //赤壁麻将硬晃
	KIND_ID_CBMJ_HZ    = 589 //赤壁麻将红中玩法
	KIND_ID_TMMJ_YLKZC = 588 //天门麻将一赖可捉统
	KIND_ID_TMMJ_YH    = 584 //天门麻将硬晃
	KIND_ID_TMMJ_YLDD  = 583 //天门麻将一赖到底
	KIND_ID_TMMJ_PLZ   = 582 //天门麻将痞癞子
	KIND_ID_QJMJ_QJHH  = 578 //潜江麻将潜江晃晃
	KIND_ID_QJMJ_JQDT  = 577 //潜江麻将经典单挑
	KIND_ID_QJMJ_QJHZ  = 576 //潜江麻将潜江红中
	KIND_ID_HM_WHMJ    = 889 //斑马汉麻武汉麻将
	KIND_ID_HM_WHHH    = 587 //斑马汉麻武汉晃晃
	KIND_ID_JY_HZLZG   = 581 //嘉鱼红中癞子杠
	KIND_ID_JY_HH      = 580 //嘉鱼晃晃
	KIND_ID_JY_YQ      = 579 //嘉鱼硬巧
	KIND_ID_TC_HZLZG   = 572 //通城红中赖子杠
	KIND_ID_TS_TSMJ    = 574 //通山麻将
	KIND_ID_EZ_510K    = 486 //鄂州510K好友4人
	//20190403 洪湖3款麻将 苏大强
	KIND_HH_YLDD = 570
	KIND_HH_LH   = 569
	KIND_HH_YH   = 568
	//-------------------
	//----金币场游戏-----
	KIND_ID_GOLD_PDK_QC3  = 387  // 跑得快3人金币
	KIND_ID_GOLD_DouDiZhu = 382  // 斗地主金币
	KIND_ID_GOLD_HFBH     = 1005 // 鹤峰百胡金币
	KIND_ID_GOLD_ESMJ     = 1009 // 恩施麻将金币
	KIND_ID_GOLD_XTMJ_LH  = 1010 // 仙桃晃晃
	KIND_ID_GOLD_TCMJ     = 1025 // 通城麻将金币
	KIND_ID_GOLD_TCGZ     = 1006 // 通城个子金币
	//-------------------
)

// ! 游戏模式
const (
	GAME_TYPE_GOLD   = 0 //金币模式
	GAME_TYPE_MATCH  = 1 //比赛模式
	GAME_TYPE_FRIEND = 2 //好友模式
)

// 贡献值系统
const (
	Contribution_Lucky_Default = iota
	Contribution_Lucky_Good
	Contribution_Lucky_Bad
)

const (
	LuckyType_GameContribution = iota //游戏贡献值
	LuckyType_NewPlayerCtrl           //新玩家强补控制
	LuckyType_DayScoreCtrl            //日净分控制
)

const (
	ID_IPHONE_OLD               = 11                    //老手机版本号
	GAME_OPERATION_TIME_60      = 60                    //操作超时时间(60秒场)
	GAME_OPERATION_TIME_30      = 30                    //操作超时时间(30秒场)
	GAME_OPERATION_TIME_15      = 15                    //操作超时时间(15秒场)
	GAME_OPERATION_TIME_12      = 12                    //操作超时时间(12秒场)
	GAME_OPERATION_TIME         = 9                     //操作超时时间
	GAME_LIANGPAI_AUTOOUT_TIME  = 1                     //亮牌后玩家自动出牌时间
	GAME_LIANGPAI_AUTOOUT_TIME3 = 3                     //亮牌后玩家自动出牌时间
	XTMJ_OPERATION_TIME         = 12                    //操作超时时间
	IDI_MAX_TIME_ID             = 30                    //极限定时器
	IDI_OFFLINE                 = (IDI_MAX_TIME_ID + 1) //断线定时器

	DISMISS_ROOM_TIMER        = 15   // 六百秒自动解散房间
	OFFLINE_ROOM_TIMER        = 30   // 离线300秒解散房间
	VITAMIN_LOW_DISMISS_TIMER = 15   // 进入防沉迷300秒解散房间
	OFFLINE_ROOM_TIMER_hOUR   = 7200 // 离线300秒解散房间
	READT_ROOM_TIMER          = 15   // 准备倒计时

	//#define MAX_CHAIR						100								//最大椅子
	MAX_CHAIR_NORMAL = 8                //最大人数
	MAX_CHAIR        = MAX_CHAIR_NORMAL //最大椅子

	MAX_ANDROID      = 256  //最大机器
	MAX_CHAT_LEN     = 128  //聊天长度
	LIMIT_CHAT_TIMES = 1200 //限时聊天
	//游戏状态
	GS_FREE    = 0   //空闲状态
	GS_PLAYING = 100 //游戏状态
	//长度宏定义
	TYPE_LEN    = 32  //种类长度
	KIND_LEN    = 32  //类型长度
	STATION_LEN = 32  //站点长度
	SERVER_LEN  = 32  //房间长度
	MODULE_LEN  = 32  //进程长度
	RULE_LEN    = 512 //规则长度

	//性别定义
	GENDER_NULL = 0 //未知性别
	GENDER_BOY  = 1 //男性性别
	GENDER_GIRL = 2 //女性性别

	//游戏类型
	GAME_GENRE_SCORE   = 0x0001 //点值类型
	GAME_GENRE_GOLD    = 0x0002 //金币类型
	GAME_GENRE_MATCH   = 0x0004 //比赛类型
	GAME_GENRE_EDUCATE = 0x0008 //训练类型

	//用户状态定义
	US_NULL    = 0x00 //没有状态
	US_FREE    = 0x01 //站立状态
	US_SIT     = 0x02 //坐下状态
	US_READY   = 0x03 //同意状态
	US_LOOKON  = 0x04 //旁观状态
	US_PLAY    = 0x05 //游戏状态
	US_OFFLINE = 0x06 //断线状态
	//20191206 苏大强 托管
	US_TRUST = 0x10 //托管状态

	//长度宏定义
	NAME_LEN        = 32 //名字长度
	PASS_LEN        = 33 //密码长度
	EMAIL_LEN       = 32 //邮箱长度
	GROUP_LEN       = 32 //社团长度
	COMPUTER_ID_LEN = 33 //机器序列
	UNDER_WRITE_LEN = 32 //个性签名
	GAME_NUM_LEN    = 32 //游戏唯一标志串长度

	//////////////////////////////////////////////////////////////////////////
	//常用常量

	//无效数值
	INVALID_BYTE  = (byte(0xFF))         //无效数值
	INVALID_WORD  = (uint16(0xFFFF))     //无效数值
	INVALID_DWORD = (uint32(0xFFFFFFFF)) //无效数值

	//无效数值
	INVALID_TABLE = INVALID_WORD //无效桌子
	INVALID_CHAIR = INVALID_WORD //无效椅子

	//结束原因
	GER_NORMAL                   = 0x00 //常规结束
	GER_DISMISS                  = 0x01 //游戏解散
	GER_USER_LEFT                = 0x02 //用户离开
	GER_GAME_ERROR               = 0x03 //程序出错
	GER_DISMISS_OVERTIME_OUTCARD = 0x04 //20210406 荆州戳虾子，为了大结算的记录

	//常量定义
	MAX_WEAVE     = 4   //最大组合
	MAX_INDEX     = 34  //最大索引
	MAX_COUNT     = 14  //最大数目
	MAX_REPERTORY = 136 //最大库存
	ALL_CARD      = 108 //总扑克数目
	ALL_CARD_P3   = 52  //一副牌，无大小王

	MAX_CARD         = 40 //（每人的最大扑克数）这个变量要按照所有玩法中最大的来算，因为要提前定义好数组（3人是40,4人时是27，这个就是最大的40）
	MAX_CARD_3P      = 40 //3人玩法时每人的最大扑克数，确定庄家后，庄家34+6张底牌
	MAX_CARD_3P_SEND = 34 //3人玩法时每人发牌时的最大扑克数，庄家的6张底牌后面发的
	MAX_CARD_4P      = 27 //4人玩法时每人的最大扑克数
	MAX_CARD_P3      = 3  //拼三玩法，发牌数
	MAX_PLAYER_4P    = 4  //4人拱游戏人数
	MAX_PLAYER_3P    = 3  //3人拱游戏人数
	MAX_PLAYER_2P    = 2  //2人拱游戏人数
	MAX_DOWNCARDNUM  = 6  //3人拱底牌数目

	MIN_ONESTR_COUNT = 3 //顺子的最小长度为3，比如567就可以组成顺子
	MIN_BOMB_COUNT   = 3 //组成炸弹时的最小张数，蕲春打拱是3，比如555就是炸弹

	//游戏状态
	GS_MJ_FREE              = GS_FREE          //空闲状态
	GS_MJ_PLAY              = (GS_PLAYING + 1) //游戏状态
	GS_MJ_HAI_DI            = (GS_PLAYING + 2) //海底状态
	GS_MJ_END               = (GS_PLAYING + 3) //结束状态
	GS_MJ_PAO               = (GS_PLAYING + 4) //下跑下漂阶段
	GS_MJ_CONTINUE          = (GS_PLAYING + 5) //金币场中如果玩家继续玩，那么就不按照规则中当庄的人是谁就是谁
	GS_MG_WAITING_RECONNECT = (GS_PLAYING + 6) //等待重连

	//逻辑掩码
	MASK_COLOR = 0xF0 //花色掩码
	MASK_VALUE = 0x0F //数值掩码

	//消息类型
	SMT_INFO       = 0x0001 //信息消息
	SMT_EJECT      = 0x0002 //弹出消息
	SMT_GLOBAL     = 0x0004 //全局消息
	SMT_CLOSE_GAME = 0x1000 //关闭游戏
)

const (
	//乱将权位
	CHR_TIAN_HU            = 0x0001   //天胡
	CHR_DI_HU              = 0x0002   //地胡
	CHR_MEN_QIAN_QING      = 0x0004   //门前清
	CHR_RECHONG            = 0x0008   //热铳状态
	CHR_LIANGBAO           = 0x0010   //20210511 苏大强 浠水晃晃 亮宝成功
	CHR_QING_YI_SE         = 0x0100   //清色权位，包括了万一色，条一色，筒一色，将一色
	CHR_QUAN_QIU_REN       = 0x0200   //全求权位
	CHR_FENG_YI_SE         = 0x0400   //风一色权位，
	CHR_JIANG_JIANG        = 0x0800   //风一色权位，
	CHR_QIANG_GANG         = 0x1000   //抢杆权位
	CHR_HAI_DI             = 0x2000   //海底权位
	CHR_GANG_SHANG_KAI_HUA = 0x4000   //杠上开花权位
	CHR_MENG_QING          = 0x8000   //门清权位
	CHR_DA_SAN_YUAN        = 0x10000  //大三元
	CHR_XIAO_SAN_YUAN      = 0x20000  //小三元
	CHR_SI_GUI_YI_AN       = 0x40000  //暗四归一
	CHR_SI_GUI_YI_MING     = 0x80000  //明四归一
	CHR_KA_5_XING          = 0x100000 //卡五星
	CHR_GANG_SHANG_PAO     = 0x200000 //杠上炮
	CHR_HAI_DI_PAO         = 0x400000 //海底炮
)

// 监利红中赖子杠使用
const (
	CHR_JIAN_ZI_HU = 0x0008 //见字胡
)

// 胡牌定义
const (
	//小胡牌型
	CHK_NULL            = 0x0000 //非胡类型
	CHK_PING_HU_NOMAGIC = 0x0001 //平胡类型,平胡无赖子
	CHK_PING_HU_MAGIC   = 0x0002 //平胡类型,平胡有赖子
	CHK_FOUR_LAIZE      = 0x0004 //四个赖子胡牌
	CHK_PENG_PENG       = 0x0008 //碰碰胡牌(崇阳碰碰胡是小胡)

	//大胡牌型
	CHK_DA_HU_NOMAGIC = 0x1000 //大胡无赖子类型
	CHK_DA_HU_MAGIC   = 0x2000 //大胡有赖子类型

	//不需将的胡牌类型
	CHK_QING_YI_SE        = 0x0200 //清色类型，包括了万一色，条一色，筒一色，
	CHK_FENG_YI_SE        = 0x0400 //风一色，乱将
	CHK_JIANG_JIANG       = 0x0800 //将将胡牌，即将一色
	CHK_SI_LAIZI_NO_HUPAI = 0x8000 //4个赖子胡牌（不构成牌型）

	CHK_7_DUI   = 0x4000  //七对
	CHK_7_DUI_1 = 0x20000 //豪华七对
	CHK_7_DUI_2 = 0x40000 //双豪华七对
	CHK_7_DUI_3 = 0x80000 //三豪华七对

	CHK_HUN_YI_SE        = 0x100000 //混一色
	CHK_YING_QUE         = 0x200000 //硬缺
	CHK_RUANG_QUAN       = 0x400000 //软缺
	CHK_THIRTEEN_ORPHANS = 0x800000 //十三幺
	//需将的胡牌类型
	CHK_QIANG_GANG         = 0x0010     //抢杠
	CHK_HAI_DI             = 0x0020     //海底胡牌
	CHK_QUAN_QIU_REN       = 0x0040     //全求人胡牌，
	CHK_GANG_SHANG_KAI_HUA = 0x0080     //杠上开花
	CHK_MEN_QIAN_QING      = 0x0100     //门前清
	CHK_XIAO_SAN_YUAN      = 0x10000    //小三元
	CHK_DA_SAN_YUAN        = 0x1000000  //大三元
	CHK_SI_GUI_YI_MING     = 0x2000000  //明四归一
	CHK_SI_GUI_YI_AN       = 0x4000000  //暗四归一
	CHK_KA_5_XING          = 0x8000000  //卡五星
	CHK_SHOU_ZHUA_YI       = 0x10000000 //手抓一
	CHK_GANG_SHANG_PAO     = 0x20000000 //杠上炮
	CHK_HAI_DI_PAO         = 0x40000000 //海底炮
	CHK_HU_BIAN            = 0x80000000 //胡边

)

// 张家口，胡牌牌型
const (
	CHK_YI_TIAO_LONG = 0x01000000 //一条龙
	CHK_SI_GUI_YI_1  = 0x10000000 //四归一x1
	CHK_SI_GUI_YI_2  = 0x20000000 //四归一x2
	CHK_SI_GUI_YI_3  = 0x40000000 //四归一x3
	CHK_SI_GUI_YI_4  = 0x80000000 //四归一x4
)

// 滁州玩法，胡牌牌型
const (
	//CHR_CZ_HARD = 0x0001 //硬缺
	//CHR_RUANG_QUAN = 0x0002 //软缺
	CHR_DUAN_19      = 0x0010  //断幺
	CHR_YI_TIAO_LONG = 0x0020  //一条龙
	CHR_JIE_MEI_PU   = 0x0040  //姊妹铺
	CHR_JIE_MEI_PU2  = 0x0080  //双姊妹铺
	CHR_SAN_KAN      = 0x0100  //三坎
	CHR_SAN_KAN4     = 0x0200  //四坎
	CHR_QUAN_19      = 0x0400  //全19
	CHR_SI_YUNZI     = 0x0800  //四云子
	CHR_KANG_JIANG   = 0x1000  //坎将
	CHR_SI_CHA       = 0x2000  //四叉
	CHR_SAN_DA_JIANG = 0x4000  //三大将
	CHR_SI_JI_FENG   = 0x8000  //四季风
	CHR_PING_HU      = 0x10000 //平胡

)

// 动作定义
const (
	WIK_NULL      = 0x00   //没有类型
	WIK_LEFT      = 0x01   //左吃类型
	WIK_CENTER    = 0x02   //中吃类型
	WIK_RIGHT     = 0x04   //右吃类型
	WIK_PENG      = 0x08   //碰牌类型
	WIK_FILL      = 0x10   //补牌类型
	WIK_GANG      = 0x20   //杠牌类型
	WIK_CHI_HU    = 0x40   //吃胡类型
	WIK_QIANG     = 0x80   //抢暗杠类型
	WIK_BAO_QING  = 0x100  //报清类型
	WIK_TING      = 0x200  //报听类型
	WIK_BUHUA     = 0x400  //补花类型
	WIK_FENG      = 0x800  //报风一色类型
	WIK_JIANG     = 0x1000 //报将一色类型
	WIK_TAIZHUANG = 0x2000 //抬庄
	WIK_LIANGPAI  = 0x4000 //亮牌
	//20191218 苏大强 禁胡
	WIK_JINHU = 0x8000  //禁胡
	WIK_JINYP = 0x10000 //禁止养痞
	WIK_K3Z   = 0x20000 //卡字胡
)

const WIK_LIANG = 0x80 //亮倒类型

// 20190917 苏大强 log中直接显示牌权
var GongInfoMap = map[uint64]string{
	WIK_NULL:     "无操作",
	WIK_LEFT:     "左吃",
	WIK_CENTER:   "中吃",
	WIK_RIGHT:    "右吃",
	WIK_PENG:     "碰牌",
	WIK_FILL:     "补牌",
	WIK_GANG:     "杠牌",
	WIK_CHI_HU:   "吃胡",
	WIK_QIANG:    "抢杠",
	WIK_BAO_QING: "报清",
	WIK_TING:     "报听",
	WIK_BUHUA:    "补花",
	WIK_FENG:     "风一色",
	WIK_JIANG:    "将一色",
}

// 所有组合牌型
const WIK_ACTION_ALL = WIK_NULL | WIK_LEFT | WIK_CENTER | WIK_RIGHT | WIK_PENG | WIK_FILL | WIK_FILL | WIK_GANG

// 报清动作定义
const (
	WIK_BAOQING_NULL     = 0x00 //报清无
	WIK_BAOQING_QING     = 0x01 //报清一色
	WIK_BAOQING_FENG     = 0x02 //报风一色
	WIK_BAOQING_JIANG    = 0x04 //报将一色
	WIK_BAOQING_LIANGBAO = 0x10 //报亮宝
	WIK_BAOQING_QI       = 0x08 //放弃报清
)

const (
	WANFA_XN = 4 //咸宁玩法
	WANFA_JY = 5 //嘉鱼玩法
	WANFA_CY = 6 //崇阳玩法
	WANFA_TS = 7 //通山玩法
	WANFA_TC = 8 //通城玩法

	TYPE_LAIZI_GANG_FACAI = 1 //发财赖子杠
	TYPE_LAIZI_GANG_PIZI  = 2 //皮子杠
	TYPE_LAIZI_GANG_LAIZI = 3 //赖子杠
)

const (
	DING_NULL        = 0  //没有顶
	DING_FENG        = 1  //封顶
	DING_JIN         = 2  //金顶
	DING_YANGGUANG   = 3  //阳光顶
	DING_SANYANG     = 4  //三阳顶
	DING_JI          = 5  //极顶
	DING_ZUANSHI     = 6  //钻石顶
	DING_YIN         = 7  //银顶
	DING_ZHUDU       = 8  //猪肚
	DING_FENGBAO     = 9  //风豹
	DING_BIAN        = 10 //边顶
	DING_JINDING     = 11 ////金鼎
	DING_CHAOJINDING = 12 //超金鼎

)

// 会员道具
const (
	PROP_DOUBLE   = iota //双倍积分卡
	PROP_FOURDOLD        //四倍积分卡
	PROP_NEGAGIVE        //负分清零
	PROP_FLEE            //清逃跑率
	PROP_BUGLE           //小喇叭
	PROP_KICK            //防踢卡
	PROP_SHIELD          //护身符
	PROP_MEMBER_1        //会员道具
	PROP_MEMBER_2        //会员道具
	PROP_MEMBER_3        //会员道具
	PROP_MEMBER_4        //会员道具
	PROP_MEMBER_5        //会员道具
	PROP_MEMBER_6        //会员道具
	PROP_MEMBER_7        //会员道具
	PROP_MEMBER_8        //会员道具
	PROP_MEMBER_MAX
)

// 成员统计排序
const (
	DAY_RECORD_TODAY     = 0
	DAY_RECORD_YESTERDAY = 1
	DAY_RECORD_3DAYS     = 2
	DAY_RECORD_7DAYS     = 3
)

// 包厢排行榜
const (
	RANK_TIME_TODAY     = 0
	RANK_TIME_YESTERDAY = -1
	RANK_TIME_CARVE     = 1
	RANK_TIME_WEEK      = 2
	RANK_TIME_MONTH     = 3
)
const (
	RANK_TYPE_ROUND  = 0
	RANK_TYPE_WINER  = 1
	RANK_TYPE_RECORD = 2
)

const (
	SORT_PLAYTIMES_DES    = 0
	SORT_PLAYTIMES_AES    = 1
	SORT_BWTIMES_DES      = 2
	SORT_BWTIMES_AES      = 3
	SORT_TOTALSCORE_DES   = 4
	SORT_TOTALSCORE_AES   = 5
	SORT_INVALIDROUND_DES = 6
	SORT_INVALIDROUND_AES = 7

	SORT_VITAMIN_DES = 6
	SORT_VITAMIN_AES = 7
	SORT_OFFTIME_DES = 8
	SORT_OFFTIME_AES = 9
)

const (
	SORT_MEMROLE_DESC         = 0
	SORT_VITAMINCOST_DESC     = 1
	SORT_VITAMINCOST_ASC      = 2
	SORT_VITAMINLEFT_DESC     = 3
	SORT_VITAMINLEFT_ASC      = 4
	SORT_VITAMINMINUS_DESC    = 5
	SORT_VITAMINMINUS_ASC     = 6
	SORT_VITAMINWINLOSEP_DESC = 7
	SORT_VITAMINWINLOSEP_ASC  = 8
	SORT_VITAMIN_DESC         = 9
	SORT_VITAMIN_ASC          = 10
	SORT_ALARMVALUE_DESC      = 11
	SORT_ALARMVALUE_ASC       = 12
	SORT_PLAYERNUM_DESC       = 13
	SORT_PLAYERNUM_ASC        = 14
)

const (
	SORT_REWARDVITAMIN_DESC = 0
	SORT_REWARDVITAMIN_ASC  = 1
	SORT_REWARD_DESC        = 2
	SORT_REWARD_ASC         = 3
)

const (
	SORT_VITAMIN_BY_ROLE        = -1
	SORT_VITAMINCUR_DESC        = 0
	SORT_VITAMINCUR_ASC         = 1
	SORT_VITAMINPRE_DESC        = 2
	SORT_VITAMINPRE_ASC         = 3
	SORT_VITAMINWINLOSE_DESC    = 4
	SORT_VITAMINWINLOSE_ASC     = 5
	SORT_VITAMINPLAYCOST_DESC   = 6
	SORT_VITAMINPLAYCOST_ASC    = 7
	SORT_VITAMINPLAYTIMES_DESC  = 8
	SORT_VITAMINPLAYTIMES_ASC   = 9
	SORT_VITAMINBWTIMES_DESC    = 10
	SORT_VITAMINBWTIMES_ASC     = 11
	SORT_VITAMINVALIDROUND_DESC = 12
	SORT_VITAMINVALIDROUND_ASC  = 13
)

const (
	SORT_ROUND_DES = 0
	SORT_ROUND_AES = 1
	SORT_PLAYS_DES = 2
	SORT_PLAYS_AES = 3
	SORT_CARD_DES  = 4
	SORT_CARD_AES  = 5
	SORT_COST_DES  = 6
	SORT_COST_AES  = 7
)

const (
	RECORD_STATUS_BWTIMES   = 0
	RECORD_STATUS_PLAYTIMES = 1
)

const EIGHT_DAY_SECONDS = 691200
const SEVEN_DAY_SECONDS = 604800
const ONE_DAY_SECONDS = 86400

const nChihuCount = 10

// 胡牌类型掩码数组
var KindMask = [nChihuCount]int{
	CHK_GANG_SHANG_KAI_HUA,
	CHK_FENG_YI_SE,
	CHK_JIANG_JIANG,
	CHK_QING_YI_SE,
	CHK_7_DUI,
	CHK_QIANG_GANG,
	CHK_PENG_PENG,
	CHK_QUAN_QIU_REN,
	CHK_HAI_DI,
}

// 开始模式
const (
	StartMode_FullReady   = iota //满人开始
	StartMode_AllReady           //所有准备
	StartMode_Symmetry           //对称开始
	StartMode_TimeControl        //时间控制
)

// 效验类型
const (
	EstimatKind_OutCard      = iota //出牌效验
	EstimatKind_GangCard            //杠牌效验
	EstimatKind_AnGangCard          //暗杠牌效验
	EstimatKind_MingGangCard        //明杠牌效验
	EstimatKind_BuCard              //补牌效验
)

// 分数类型
const (
	ScoreKind_Win     = iota //胜
	ScoreKind_Lost           //输
	ScoreKind_Draw           //和
	ScoreKind_Flee           //逃
	ScoreKind_Service        //服务
	ScoreKind_Present        //赠送
	ScoreKind_pass           //忽略
)

// 游戏结束状态
const (
	FINISH_STA_NORMAL       = iota // 正常结束
	FINISH_STA_1ST_DISMISS         // 首局解散
	FINISH_STA_HALF_DISMISS        // 中途解散
	FINISH_STA_PLAYING             // 游戏中
)

const (
	GameBigHuKindQG    = 4  //抢杠
	GameBigHuKindQG1   = 13 //抢杠,崇阳玩法，抢杠失败游戏结束，不能播放抢杠胡的动画
	GameBigHuKindFYS   = 6  //风一色
	GameBigHuKindFYS_7 = 15 //风一色的七对
	GameBigHuKindQQR   = 8  //全求人
	GameBigHuKindQYS   = 10 //清一色
	GameBigHuKindQYS_7 = 16 //清一色的七对
	GameBigHuKindJYS   = 11 //将一色
	GameBigHuKindJYS_7 = 17 //将一色
	GameBigHuKindPP    = 9  //碰碰胡
	GameBigHuKindHDL   = 5  //海底捞
	GameBigHuKindGSK   = 7  //杠上开花
	GameBigHuKind_7    = 14 //七对
	GameBigHuKindMQQ   = 12 //门前清
	GameBigHuKindZG    = 18 //扎杠
	//add by zwj for 汉川撮虾子
	GameNoMagicHu        = 19 //硬胡
	GameMagicHu          = 20 //软胡
	GameBigHuKindK5X     = 21 //卡五星
	GameBigHuKindDSY     = 22 //大三元
	GameBigHuKindXSY     = 23 //小三元
	GameBigHuKindSZ1     = 24 //手抓一
	GameBigHuKindA4G     = 25 //暗四归一
	GameBigHuKindM4G     = 26 //明四归一
	GameBigHuKind_7Dui_1 = 27 //豪华七对
	GameBigHuKind_7Dui_2 = 28 //超豪华七对
	GameBigHuKind_7Dui_3 = 29 //超超豪华七对
	GameBigHuKindHDP     = 30 //海底炮
	GameReChongHu        = 31 //热冲
	GameRgk              = 32 //软杠开
	GameYgk              = 33 //硬杠开
)

const (
	GameBigHuKindHM         = 20 // 黑摸
	GameBigHuKindRM         = 21 // 软摸
	GameBigHuKindSK         = 22 // 刷开
	GameBigHuKindRC         = 23 // 热统
	GameBigHuKindTH         = 24 // 天胡
	GameBigHuKindShuangDaHu = 25 // 双大胡
)

const (
	K5X_QuBaJiu_No = iota
	K5X_QuBaJiu_89 //去八九
	K5X_QuBaJiu_9  //去九
)

const (
	K5X_ShangLou_BaoZi       = 1
	K5X_ShangLou_HuangZhuang = 2
)

// 游戏级别结构
type TagLevelItem struct {
	lLevelScore int       //级别积分
	szLevelName [16]uint8 //级别描述
}

type GameCardConfig struct {
	//配置文件要通过tag来指定配置文件中的名称
	Able          int    `ini:"nAble"`
	User1         string `ini:"user1"`
	User2         string `ini:"user2"`
	User3         string `ini:"user3"`
	User4         string `ini:"user4"`
	RepertoryCard string `ini:"RepertoryCard"`
}

// 子游戏相关配置
type GameConfig struct {
	PlayerCount        uint16 `json:"playercount"`
	ChairCount         uint16 `json:"chaircount"`
	LookonCount        uint16 `json:"lookoncount"`
	LookonSupport      bool   `json:"lookonsupport"`
	StartMode          int    `json:"startmode"`
	StartIgnoreOffline bool   `json:"start_ignore_offline"` // 首局开始时 忽略离线玩家
	OffLineZBFP        bool   `json:"offlinezbfp"`          // 20200708 苏大强 准备玩家离线后 可发牌  目前用于京山
}

type TagCard struct {
	HDCard    byte `json:"card"` //海底扑克
	Index     int  `json:"id"`   //海底牌索引
	UserIndex int  `json:"user"` //海底牌玩家索引
}

// 类型子项
type TagKindItem struct {
	WeaveKind  byte    //组合类型
	CenterCard byte    //中心扑克
	CardIndex  [3]byte //扑克索引
}

// 组合子项
type TagWeaveItem struct {
	WeaveKind    byte   `json:"weavekind"`    //组合类型
	ProvideUser1 uint16 `json:"provideuser1"` //蓄杠，如果是蓄杠，保留碰牌的供应者
	CenterCard   byte   `json:"centercard"`   //中心扑克
	PublicCard   byte   `json:"publiccard"`   //公开标志
	ProvideUser  uint16 `json:"provideuser"`  //供应用户
	MagicWeave   byte   `json:"MagicWeave"`   //20190322 苏大强 组合中有赖子的情况  考虑可能有限制，用byte应该是够了
}

type TagCardLeftItem struct {
	MaxCount  int     `json:"maxcount"`    //最大牌数量
	CardArray []int   `json:"cardarray"`   //牌堆,0代表有牌，-1代表没牌
	OffSet    int     `json:"offset"`      //牌堆前面索引
	EndOffSet int     `json:"end_off_set"` //牌堆后面索引
	Random    [2]byte `json:"random"`      //甩子
	Seat      int     `json:"seat"`
	Kaikou    int     `json:"kaikou"`
}

//这个先放这里，滁州老三番 全球人包牌，
/*
全球人首圈，弃牌是胡牌的上下2只，就算是包牌
*/
type BaoPaiInfo struct {
	FristQuanQiuRen bool `json:"fristQuanQiuRen"` //首听状态
	BaoPaicard      byte `json:"baoPaicard"`      //包牌值
}

type FiveGangSource struct {
	AreadyFiveGang bool    `json:"areadyfivegang"` //5杠状态
	UserStatus     [4]bool `json:"baoPaicard"`     //玩家的状态（是不是已经处理过牌权）
}

// 玩家杠明细
type TagWeaveDetail struct {
	WeaveKind   byte
	ProvideUser uint16
}

func (this TagWeaveItem) String() string {
	return fmt.Sprintf("用户组合牌牌组:\n组合类型[%d],中心扑克[%d],公开标志[%d],供应用户[%d]\n", this.WeaveKind, this.CenterCard, this.PublicCard, this.ProvideUser)
}

// 胡牌结果
type TagChiHuResult struct {
	ChiHuKind  uint64       //吃胡类型
	ChiHuRight uint16       //胡牌权位
	ChiHuUser  uint16       //吃胡玩家
	ChiHuKind2 uint64       //吃胡类型2,部分地方玩法额外新加胡牌牌型(滁州)
	HuDetail   ChuZouHuType //胡牌类型（滁州）
}

// 杠牌结果
type TagGangCardResult struct {
	MagicCard byte    //赖子牌
	CardCount byte    //扑克数目
	CardData  [6]byte //扑克数据
}

// 补牌结果
type TagBuResult struct {
	CardCount byte    //扑克数目
	CardData  [6]byte //扑克数据
}

// 分析子项
type TagAnalyseItem struct {
	CardEye    byte    //牌眼扑克
	WeaveKind  [4]byte //组合类型
	CenterCard [4]byte //中心扑克
}

// 分析子项
type TagRobotAnalyseItem struct {
	CardEye    byte      //牌眼扑克
	WeaveKind  []byte    //组合类型
	CenterCard []byte    //中心扑克
	CardIndex  [][6]byte //扑克索引
}

// 玩家第三次开口对象
type ThirdOpreate struct {
	ProvideUser  byte //第三次开口的提供者
	OperateCount byte //操作次数
}

// 用户积分信息
type TagUserScore struct {
	Vitamin     float64 `json:"vitamin"`     // 用户疲劳值
	Score       int     `json:"score"`       // 用户分数
	GameGold    int     `json:"gamegold"`    // 游戏金币//
	InsureScore int     `json:"insurescore"` // 存储金币
	WinCount    int     `json:"wincount"`    // 胜利盘数
	LostCount   int     `json:"lostcount"`   // 失败盘数
	DrawCount   int     `json:"drawcount"`   // 和局盘数
	FleeCount   int     `json:"fleecount"`   // 断线数目
}

func (v1 TagUserScore) ToV2() TagUserScoreV2 {
	v2 := TagUserScoreV2{}
	v2.Vitamin = v1.Vitamin
	v2.Score = float64(v1.Score)
	v2.GameGold = v1.GameGold
	v2.InsureScore = v1.InsureScore
	v2.WinCount = v1.WinCount
	v2.LostCount = v1.LostCount
	v2.DrawCount = v1.DrawCount
	v2.FleeCount = v1.FleeCount
	return v2
}

// 用户积分信息
type TagUserScoreV2 struct {
	Vitamin     float64 `json:"vitamin"`     // 用户疲劳值
	Score       float64 `json:"score"`       // 用户分数
	GameGold    int     `json:"gamegold"`    // 游戏金币
	InsureScore int     `json:"insurescore"` // 存储金币
	WinCount    int     `json:"wincount"`    // 胜利盘数
	LostCount   int     `json:"lostcount"`   // 失败盘数
	DrawCount   int     `json:"drawcount"`   // 和局盘数
	FleeCount   int     `json:"fleecount"`   // 断线数目
}

// 扣房卡
type TagDeleteFangKa struct {
	UserID     int64  //用户ID
	GameNum    string //牌局
	RoomNum    int    //房间号
	Count      uint32 //数量
	Players    int    //数量:人数
	Rule       string //规则
	TeaHouseID string //包厢ID
}

// 可胡牌类型
type TagHuType struct {
	HAVE_SIXI_HU            bool //定义 | 四喜
	HAVE_QUE_YISE_HU        bool //定义 | 缺一色
	HAVE_BANBAN_HU          bool //定义 | 板板胡
	HAVE_LIULIU_HU          bool //定义 | 六六顺
	HAVE_QING_YISE_HU       bool //定义 | 清一色
	HAVE_FENG_YI_SE         bool //定义 | 风一色
	HAVE_HAO_HUA_DUI_HU     bool //定义 | 豪华对
	HAVE_JIANG_JIANG_HU     bool //定义 | 将将胡
	HAVE_FENG_YISE_HU       bool //定义 | 风一色胡
	HAVE_QUAN_QIU_REN       bool //定义 | 全求人胡
	HAVE_PENG_PENG_HU       bool //定义 | 碰碰胡
	HAVE_HAI_DI_HU          bool //定义 | 海底胡牌
	HAVE_QIANG_GANG_HU      bool //定义 | 抢杠
	HAVE_GANG_SHANG_KAI_HUA bool //定义 | 杠上开花
	HAVE_LAI_ZI_GANG_KAI    bool //定义 | 癞子杠开
	HAVE_MENG_QING          bool //定义 | 门清
	HAVE_DI_HU              bool //定义 | 地胡
	HAVE_TIAN_HU            bool //定义 | 天胡
	HAVE_ZIMO_JIAO_1        bool //定义 | 自摸加1
	HAVE_QIDUI_HU           bool //定义 | 七对
	HAVE_RETONG_HU          bool //定义 | 热统
	HAVE_JIANZI_HU          bool //定义 | 见字胡
	HAVE_YI_PAO_DUO_XIANG   bool //定义 | 一炮多响
}

// ////////////////////////////////////////////////////////////////////////////
// 以下为纸牌类游戏特有的
// ////////////////////////////////////////////////////////////////////////////
const (
	//纸牌出牌类型
	TYPE_ERROR            = -1 //错误的类型
	TYPE_NULL             = 0  //没有类型
	TYPE_ONE              = 1  //单张
	TYPE_TWO              = 2  //两张
	TYPE_THREE            = 3  //三张
	TYPE_ONESTR           = 4  //顺子
	TYPE_TWOSTR           = 5  //连对 445566
	TYPE_THREESTR         = 6  //三张的连对 444555
	TYPE_KING             = 7  //王炸
	TYPE_BOMB_DOUBLE_KING = 8  //对王炸
	TYPE_BOMB_FOUR_KING   = 9  //四张王
	TYPE_BOMB_510K        = 10 //510k炸弹
	TYPE_BOMB_NOMORL      = 11 //普通炸弹
	TYPE_BOMB_7XI         = 12 //7喜，7张一样的牌组成的炸弹
	TYPE_BOMB_8XI         = 13 //8喜，8张一样的牌组成的炸弹
	TYPE_BOMB_STR         = 14 //连炸，摇摆，拱笼 44445555
	TYPE_BOMB_KING        = 15 //王炸
	TYPE_BOMB_SKY_KING    = 16 //天王炸
	TYPE_BOMB_GONGLONG    = 17 //拱拢类型
	TYPE_4KING_SCORE      = 18 //4王换分牌
	TYPE_BOMB_6XI         = 19 //6喜，6张一样的牌组成的炸弹
	TYPE_OTHR             = 20 //其他
)

// 一张牌数据
type TPoker struct {
	Point uint8 // 1-15;
	Index uint8 // 1-54
	Color uint8 // 0-4
}

func (this *TPoker) Set(byCard byte) {
	if byCard > 0 && byCard < 56 {
		this.Index = byCard
	}
	if byCard <= 0 {
		this.Point = 0
		this.Color = 0
		return
	}
	if byCard <= 13 {
		this.Point = byCard
		this.Color = 1
		return
	}
	if byCard <= 26 {
		this.Point = byCard - 13
		this.Color = 2
		return
	}
	if byCard <= 39 {
		this.Point = byCard - 26
		this.Color = 3
		return
	}
	if byCard <= 52 {
		this.Point = byCard - 39
		this.Color = 4
		return
	}
	if byCard == 53 {
		this.Point = 14
		this.Color = 0
		return
	}
	if byCard == 54 {
		this.Point = 15
		this.Color = 0
		return
	}
	//花牌
	if byCard == 55 {
		this.Point = 16
		this.Color = 0
		return
	}
	if byCard > 55 {
		this.Point = 0
		this.Color = 0
		return
	}
}

// 牌类型结构
type TCardType struct {
	Cardtype  int   //牌类型
	Len       int   //牌长度
	Card      uint8 //牌值，同种类型的牌时能证明牌大小的那张牌
	Color     uint8 //花色，这个只针对510k炸弹	0表示杂 4 表示黑桃 5表示大小王组成的对王 6 表示对小王 7 对大王
	Count     int   //这个只针对连炸中，最大炸弹的张数（暂时没有使用）
	BombLevel int   //这个只针对炸弹中，炸弹的大小，0表示不是炸弹
}

// 赖子替换信息结构，比如用赖子替换成了红桃5
type TFakePoker struct {
	Index     uint8 //牌值，原来的值  赖子
	Fakeindex uint8 //牌值，替代的值   例子中的红桃5
}

// 出牌数据中的赖子替换信息结构
type FakeType struct {
	Fakeking [8]TFakePoker //赖子替换信息
	CardType TCardType     //出牌数据
}

// 出牌数据,也包含赖子替换信息结构，用于提示出牌，或自动出牌的找牌
type TPokerGroup struct {
	Point     uint8                //牌型值1-15,k105无效
	Color     uint8                //花色，这个只针对k105、对王
	Count     uint8                //张数
	Cardtype  int                  //牌类型
	SortRight int                  //排序权重，出牌优先级,值小的优先出,基础值100
	Fakeking  [MAX_CARD]TFakePoker //赖子替换信息
	Indexes   []uint8              //出牌数据
}

// 炸弹结构
type TBombStr struct {
	BombLevel int                  //炸弹大小
	MaxCount  uint8                //炸弹张数
	Count     uint8                //所有张数
	Point     uint8                //炸弹的值，同数目的炸弹时能体现炸弹大小的值
	Cardtype  int                  //牌类型，暂时没有使用
	Fakeking  [MAX_CARD]TFakePoker //赖子替换信息
	Indexes   []uint8              //出牌数据
}

// 组合子项,注意保持和字牌的 TagWeaveItem 一致
type TagWeaveItemZP struct {
	WeaveKind   int      `json:"weavekind"`   //组合类型
	Cards       [10]byte `json:"cards"`       //扑克,大冶字牌最多4个，通城个子最多5个
	PublicCard  byte     `json:"publiccard"`  //公开标志，遮罩要用到这张牌
	ProvideUser uint16   `json:"provideuser"` //供应用户
}

// 分析子项
type TagAnalyseItemZP struct {
	CardEye    byte         `json:"cardeye"`    //牌眼扑克
	WeaveCount byte         `json:"weavecount"` //赢家组合数目
	WeaveKind  [20]int      `json:"weavekind"`  //组合类型
	WeaveInfo  [20][10]byte `json:"weaveinfo"`  //扑克数据
	IsWeave    [20]int      `json:"isweave"`    //每梯的牌是否是组合区的牌
	HuXi       [20]int      `json:"huxi"`       //胡息
	PublicCard [20]byte     `json:"publiccard"` //公开标志，遮罩要用到这张牌
}

// 听牌结果
type TagTingCardResult struct {
	Seat       uint16         `json:"seat"`
	MaxCount   int            `json:"maxcount"`
	TingIndex  [MAX_CARD]bool `json:"tingindex"`  //听牌索引
	TingNumber [MAX_CARD]int  `json:"tingnumber"` //听牌数量
	TingFanShu [MAX_CARD]int  `json:"tingfanshu"` //听牌蕃数
}

// 听牌结果
type TagTingCard struct {
	TingFlag          bool              `json:"tingflag"`  //是否有听牌标签
	TingIndex         int               `json:"tingindex"` //打哪张牌可以听牌，索引
	TagTingCardResult TagTingCardResult `json:"tinginfo"`  //听哪些牌
}

// 返回发给客户端的牌权信息 新的UserAction32是byte的
func GetPaiQuanStr(action uint64) (result string) {
	result = ""
	if action^WIK_NULL == 1 {
		return ""
	}
	//从大到小
	//胡
	if action&WIK_CHI_HU != 0 {
		result += GongInfoMap[WIK_CHI_HU] + ","
	}
	//抢杠
	if action&WIK_QIANG != 0 {
		result += GongInfoMap[WIK_QIANG] + ","
	}
	//报清
	if action&WIK_BAO_QING != 0 {
		result += GongInfoMap[WIK_BAO_QING] + ","
	}
	//报听
	if action&WIK_TING != 0 {
		result += GongInfoMap[WIK_TING] + ","
	}
	//补花
	if action&WIK_BUHUA != 0 {
		result += GongInfoMap[WIK_BUHUA] + ","
	}
	//风一色
	if action&WIK_FENG != 0 {
		result += GongInfoMap[WIK_FENG] + ","
	}
	//将一色
	if action&WIK_JIANG != 0 {
		result += GongInfoMap[WIK_JIANG] + ","
	}
	//杠
	if action&WIK_GANG != 0 {
		result += GongInfoMap[WIK_GANG] + ","
	}
	//碰
	if action&WIK_PENG != 0 {
		result += GongInfoMap[WIK_PENG] + ","
	}
	//吃ublic.WIK_LEFT, public.WIK_CENTER, public.WIK_RIGHT
	if action&WIK_LEFT != 0 {
		result += GongInfoMap[WIK_LEFT] + ","
	}
	if action&WIK_CENTER != 0 {
		result += GongInfoMap[WIK_CENTER] + ","
	}
	if action&WIK_RIGHT != 0 {
		result += GongInfoMap[WIK_RIGHT] + ","
	}
	//补
	if action&WIK_FILL != 0 {
		result += GongInfoMap[WIK_FILL] + ","
	}
	return
}

const (
	MaxFloorIndex = 30
)

const (
	HuType_Zimo         = iota //自摸
	HuType_BeiZimo             //被自摸
	HuType_DianPao             //点炮
	HuType_JiePao              //接炮,小结算面板显示胡牌
	HuType_BeiQiangGang        //被抢杠
	HuType_QiangGang           //抢杠
	HuType_ChaJiao             //查叫
	HuType_HuaZhu              //花猪
	HuTypeMax
)

const (
	ChihuLeaderType_ShangJia = iota
	ChihuLeaderType_XiaJia
	ChihuLeaderType_DuiJia
	ChihuLeaderType_LiangJia
	ChihuLeaderType_SanJia
	ChihuLeaderType_Max
)

const (
	RobortIdMax = 17669999
	RobortIdMin = 17650000
)

var HuTypeString = [HuTypeMax]string{
	"自摸", "被自摸", "点炮", "胡牌", "被抢杠", "抢杠", "查叫", "花猪",
}

var ChihuLeaderTypeString = [ChihuLeaderType_Max]string{
	"上家", "下家", "对家", "两家", "三家",
}

type ChihuDetail_xlch struct {
	ChihuLeaderType int    //胡牌责任人关系(上家下家对家)
	ChihuCard       int    //胡牌
	ChihuKind       uint64 //胡牌kind
	ChihuType       int    //胡牌类型,自摸或点炮
	ChihuProvider   uint16 //胡牌供牌者
	ChihuUser       uint16 //胡牌玩家
	ChihuFan        int    //胡牌当时番(自己的番加上点炮者的番)
	ChihuScore      int    //胡牌分
}

type HuDetails_xlch struct {
	HuType   string `json:"hutype"`   //胡牌类型
	HuFan    int    `json:"hufan"`    //胡牌番数
	HuScore  int    `json:"huscore"`  //胡牌分数
	HuLeader string `json:"huleader"` //胡牌责任人
}

func (self *HuDetails_xlch) SetHuDetails(info *ChihuDetail_xlch) {
	self.HuFan = info.ChihuFan
	self.HuScore = info.ChihuScore
	self.HuType = HuTypeString[info.ChihuType]
	self.HuLeader = ChihuLeaderTypeString[info.ChihuLeaderType]
}

type ChihuCardsInfo_xlch struct {
	HuCard         int    `json:"hucard"`   //玩家胡的牌
	HuCardProvider uint16 `json:"provider"` //胡牌供牌玩家椅子号
	HuCardOwner    uint16 `json:"owner"`    //胡牌玩家椅子号
}

// 打拱规则相关属性
type FriendRuleDG_qc struct {
	Difen  int `json:"difen"`   //底分
	SerPay int `json:"revenue"` //茶水
	Fa     int `json:"fa"`      //没逃跑处罚倍数
	Jiang  int `json:"jiang"`   //逃跑奖励别人的倍数
}

// 打拱规则相关属性
type FriendRuleDG_pdk struct {
	Difen  int    `json:"difen"`   //底分
	SerPay int    `json:"revenue"` //茶水
	Fa     int    `json:"fa"`      //没逃跑处罚倍数
	Jiang  int    `json:"jiang"`   //逃跑奖励别人的倍数
	XiPai  uint32 `json:"xipai"`   //洗牌次数
}

// 斗地主相关属性
type FriendRuleDG_DDZ struct {
	Difen             int    `json:"difen"`             //底分
	Radix             int    `json:"scoreradix"`        //底分基数
	SerPay            int    `json:"revenue"`           //茶水
	FengDing          int    `json:"fengding"`          //封顶，0不封顶
	JiaoZhuang        int    `json:"jiaozhuang"`        //叫庄，0叫分，1抢地主
	LaiMode           int    `json:"laimode"`           //癞子，0无癞子，1随机癞子，2花牌癞子，3随机+花牌癞子
	NextBankerMS      int    `json:"nextbankerms"`      //次局地主，0先出完先叫，1每局随机叫
	Piao              int    `json:"piao"`              //飘分，0不飘，1对飘，2不对飘
	ThreeTake0        string `json:"threetake0"`        //三张不带牌
	ThreeTake1        string `json:"threetake1"`        //三带一
	ThreeTake2        string `json:"threetake2"`        //三带二
	FourTake2         string `json:"fourtake2"`         //四带二
	FourTakeDouble2   string `json:"fourtakedouble2"`   //四带二对
	KingBombSplit     string `json:"kingbombsplit"`     //王炸不可拆
	HasHardBomb       string `json:"hardbombbig"`       //硬炸打软炸
	HardBomb4         string `json:"hardbomb4"`         //硬炸*4
	KingBombIsHard    string `json:"kingbombishard"`    //王炸算硬炸
	DoubleKingOrFour2 string `json:"doublekingorfour2"` //双王或四个二必叫地主
	SingleKingAndTwo2 string `json:"singlekingandtwo2"` //单王加二个二必叫地主
	Overtime_trust    int    `json:"overtime_trust"`    //超时托管
	Overtime_dismiss  int    `json:"overtime_dismiss"`  //超时解散
	Dissmiss          int    `json:"dissmiss"`          //解散次数，0不限制
	FleeTime          int    `json:"fleetime"`          // 客户端传来的 游戏开始前离线踢人时间
	Fa                int    `json:"fa"`                //没逃跑处罚倍数
	Jiang             int    `json:"jiang"`             //逃跑奖励别人的倍数
	XiPai             uint32 `json:"xipai"`             //洗牌次数
	Bomb0             uint32 `json:"bomb0"`             //炸弹0概率
	Bomb1             uint32 `json:"bomb1"`             //炸弹1概率
	Bomb2             uint32 `json:"bomb2"`             //炸弹2概率
	Bomb3             uint32 `json:"bomb3"`             //炸弹3概率
	Bomb4             uint32 `json:"bomb4"`             //炸弹4概率
	Bomb5             uint32 `json:"bomb5"`             //炸弹5概率
	BombSwitch        int    `json:"bombswitch"`        //炸弹开关
}

// 发牌信息储存
// 目前只保存了当前的发牌人和杠开信息
// 需要可以加
type TagSendCardInfo struct {
	Status      bool   `json:"status"`
	Uid         int64  `json:"uid"`
	CurrentUser uint16 `json:"current_user"`
	GangFlower  bool   `json:"gang_flower"`
	BuPai       bool   `json:"bupai"` //20191230 苏大强 主动的朝天杠不补拍
}

// //////////////////////////////////////////////////////////////////////
// 在线人数
const (
	GameOnlineTableAllKeys = "GameOnlineTable_*"
	GameOnlineTableFmt     = "GameOnlineTable_%d"
)

// 在线人数
type OnlineMap struct {
	RobotCount int `json:"RobotCount"` //! 机器人人数
	RealCount  int `json:"RealCount"`  //! 实际人数
}

////////////////////////////////////////////////////////////////////////
