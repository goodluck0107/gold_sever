package rule

import "github.com/open-source/game/chess.git/pkg/static"

//积分信息
type TagScoreInfo struct {
	Score     int     //游戏积分
	Score64   float64 //游戏积分
	Revenue   int     //游戏积分
	Award     int     //游戏奖励
	ScoreKind int     //分数类型
}

//开房数据
type Tag_OtherFriendCreate struct {
	UserID       int64    //用户ID
	UserPlayerID [4]int64 //用户ID1
	WinScore     [4]int
	UserName     [4]string //用户名1
	GameNum      string    //牌局
	RoomNum      int       //房间号
	KaPrice      int
	KindID       int
	Type         int // 未开始:0  开始:1  解散退卡:3  结束:2  解散退卡:4  异常解散 5
	Count        uint32
	Players      int    //人数
	Rule         string //规则
	UserName1    string //用户名
	ContextID    int
	CreateType   byte
	CurGameCount int //当前局数
}

//承包玩家
type TagContractor struct {
	ChengBaoID   byte
	ChengBaoKind int
}

//好友房相关属性
type St_FriendRule struct {
	FangZhuID            int64 //房主ID
	BankerType           int   //首局决定庄家的类型，0房主坐庄 1随机坐庄
	JuShu                int   //局数
	TypeLaizi            int   //癞子类型 = 1 //发财赖子杠  =2 //皮子杠 =3 //赖子杠
	WanFa                int   //玩法：WANFA_XN、WANFA_JY
	FengDing             int   //封顶番数
	FentDingStep1        int   //封顶的基础分倍数
	FentDingStep2        int   //封顶的基础分倍数
	FentDingStep3        int   //封顶的基础分倍数
	FentDingStep4        int   //封顶的基础分倍数
	MinFan               int   //最小番数
	MaxFan               int   //最大番数
	Bird                 int   //抓鸟个数
	DiFen                int   //底分
	TrusteeCostSharing   bool  //托管的人平摊房费
	BaseScore            int   //底注
	HasHaidi             bool  //是否有海底捞
	KouKou               bool  //是否口口番
	Has7Dui              bool  //是否七对
	NeedKaiKou           bool  //是否需要开口
	Hongzhong            bool  //是否为通用红中麻将
	NoWan                bool  //无万字牌，三人玩法时，可以勾选是否去万
	RandDiscard          bool  //是否随机去牌
	HasFaBai             bool  //咸宁麻将有：红中、中发白可选项，true表示中发白，false表示红中
	QuanQiuRenBao        bool  //是否全求人包牌
	NineSecondRoom       bool  //是否是九秒房
	RandomOutCard_9sRoom bool  //9秒场随机出牌
	HasPao               bool  //是否有跑
	Always1Round         bool  // 只有一局
	CreateType           byte  //创房类型
	BirdFan              int   //抓鸟结算番数
	CanChiPai            bool  //是否可以吃牌
	SuanFen              int   //算分类型(1整数算分, 2传统算分)
	Cardsclass           int   //牌型
	YiLaiDaoDi           bool  //一赖到底
	LaiZiKeChi           bool  //赖子可吃
	LaiZiKeChu           bool  //赖子可出
	BeiShu               int   //20191129 苏大强 小数点需求
	WithoutFaCaiBaiBan   int   //1 带中发白 2不带中发白 3不带发白 20200227 by zwj
	Havegoldtop          bool  //是否有金顶
	//!红中赖子杠规则
	GangType       int  //1发财杠,2痞子杠,3赖子杠,4红中杠
	KouKouFan      bool //口口番
	NoKouKeHu      bool //不开口可胡
	QiDuiKeHu      bool //七对可胡
	IsXiaYu        bool //下雨
	QiHuFan        int  //起胡番数
	FengDingBeiShu int  //封顶倍数,60,200,600

	//!武汉麻将
	XiaYuType    int  //1下小雨,2下大雨
	LianJin      bool //是否连金
	FanJin       bool //是否反金
	JianFengYuan bool //是否见风原
	Peibao       bool //是否包赔
	Gkbjzm       bool //杠开不计自摸

	//!嘉鱼红中赖子杠
	FengDingType  int  //1 封金顶, 2封极顶, 3封单边极顶
	GangKaiNoZiMo bool //杠开不计自摸
	HaiDiNoZiMo   bool //海底不计自摸

	//!嘉鱼晃晃
	DahuJiaBei bool //大胡加倍
	//新乡麻将 苏大强 自摸加倍
	ZmJiaBei    bool //自摸加倍
	ReChongType int  //热冲
	DingPiao    int  //定漂

	//新乡麻将
	HuMask byte //胡牌类别选项

	//!焦作推倒胡硬报到底
	OnlyZimo       bool //自摸胡
	QiDuiJiaBei    bool //七对加倍
	GangKaiJiaBei  bool //杠开加倍
	QingYiSeJiaBei bool //清一色加倍
	ZhuangJiaBei   bool //庄家加倍
	WanNengPai     bool //万能牌
	LianSixNine    bool //连6连9
	//20190507 新乡获嘉
	Mask19     bool //19句
	MaskFeng   bool //风箭刻
	MaskQueMen bool //缺门
	MingTing   bool //明听

	//!荆门钟祥推磨
	JinDingBeiShu int  //金顶倍数
	DaoChe        bool //是否可以倒车
	ShuaiPai      bool //是否可以甩牌
	NoZFB         bool //是否带中发

	//! 滁州推倒胡
	PengFengJiaFen  bool //是否碰风加分
	AnGang          bool //是否抢暗杠
	PeiFeng         bool //是否固定赔率
	IsGangKaiJiaFan bool
	Bquzheng        bool
	FengBao         bool
	//滁州老三番 自摸，捉铳一人付 三家付
	Hupattern int
	//大冶开口番
	Has_15_SecondRoom bool //是否十五秒场
	BankJiaFan        bool //庄家加番
	FengType          bool // 2人玩专属选项 true 按照金顶算 false 按照封顶算
	//20190903 苏大强 新乡黄庄不黄杠
	Hzbhg bool
	//20190917 苏大强 武汉麻将 抢蓄杠（碰后立即杠为蓄杠，碰后不立即杠的为回头杠）
	Canqiangxugang bool
	//20191018 苏大强 颍州麻将 死杠还是活杠
	ShakyGang bool
	//20191018 苏大强 颍州麻将 扎杠胡
	ZhaGang bool
	//20191109 苏大强 颍州麻将 可抢杠开
	CanQiangGK bool
	DingType   int
	Gengpi     bool
	//20191119 苏大强 超时托管和超时解散
	Overtime_trust     int //超时托管
	Overtime_dismiss   int
	Overtime_offdiss   int //离线解散时间
	Overtime_applydiss int //申请解散时间
	//20191127 苏大强 开4口加番 黄石晃晃
	K4KouJiaFan bool
	//汉川赖晃 是否可以接炮胡 by zwj
	CanPaoHu bool

	Endready bool //小结算是否自动开始下一局

	//应城晃晃
	GenHuang         bool //跟晃
	IsHaveMagic      bool //是否有癞子   ps：翻癞子玩法
	IsYingHuDouble   bool //硬胡是否翻倍*2
	IsHaveMagicziMo  bool //有赖仅自摸
	IsPiHuDPOneMagic bool //接炮屁胡一赖
	//20200418 苏大强 恩施 小血禁胡
	XiaoXueJH bool //小血禁胡
	//20200421 苏大强 恩施 每局换座
	Sjhz bool //每局换座（除庄家）

	//黄州晃晃 抢杠包胡
	QiangGangBaoHu bool

	st_FriendRule_hhs
	st_FriendRule_p3
	//!嘉鱼硬巧
	st_FriendRule_jy_yq
	//! 通城红中赖子杠
	st_FriendRule_tc_hzlzg
	//滁州
	st_FriendRule_cz
	//咸宁晃晃
	st_FriendRule_xn_xnhh
	//监利麻将
	st_FriendRule_jl_jlmj
	//蕲春打拱
	st_FriendRule_qc_qcdg
	//监利开机
	ShareScoreType int //贡献分类型
	//阳新
	st_FriendRulue_yx
	//京山麻将
	st_FriendRule_js_jsmj
	//汉川搓下虾子
	st_FriendRule_hccxz_mj
	//荆州搓下虾子
	st_FriendRule_jzcxz_mj
	//钟祥推磨

	BaoQing     bool //是否勾选报清玩法
	JianZiPaoHu bool //见字胡是否可接铳
	//应城卡五星
	st_FriendRule_K5X
	//黄石晃晃
	st_FriendRule_Mj_Hshh
	//黄冈红中
	st_FriendRule_hg_hzmj
	//浠水麻将
	st_FriendRule_xsmj

	DissmissCount int //解散次数限制,每个玩家当前大局可申请解散的次数,0无限制
	st_FriendRule_lt_hzlzg
	DeductFD_Settle int
	DissMissTeaTag  int  //包厢是否可以申请解散,0 :可以。1：不能
	DissMissMask    byte //20200806 苏大强 嘉鱼需求就算是不能解散，如果选择了托管立即解散，小解散的时候托管玩家还是能发起解散
	Radix           int  //低分基数默认1
	st_FriendRule_zjk_fhe
	st_FriendRule_xn_hzlzg
	ReadyTimeTag int //准备超时，0：代表无这个功能，other:超时时间

	SecondRoom  int  //多少秒场,15或者30
	YiSeBaoHu   bool //清一色 将一色喂第二口的包胡
	ExchangeNum int  //换牌数量，换三张3换四张4
	FanShu      int  //番数
	QueYiMen    bool //缺一门

	//黄梅麻将 add by zwj
	IsBao     int //0无宝    1宝可打     2宝不可打
	Qianggang int //0可抢杠       1不可抢
	Gskp      int // 0过手可碰        1过手不可碰
	Taizhuang int //0可抬庄        1不可台庄
	Piaofen   int // 飘分表示
	//20200326 苏大强 荆州麻将4红中可胡（首圈 ，其他圈）
	SiHZkehu int // 四红中可胡
	//武汉晃晃 add by zwj
	KyiPaoDuoXiang bool //一炮多响
	//广东推倒胡
	st_FriendRule_gd_tdh
	//转转麻将
	st_FriendRule_mj_zz
	//鹤峰麻将焖癞子
	st_FriendRule_es_hfmj
	XJSZDZB bool //小结算自动准备

	st_FriendRulue_cr //铳儿麻将
	//20200512 苏大强 武穴510k 反牌
	AntiBrand           bool
	TimesRank           int  //倍数按选项来
	Baojing             bool //要不要报警
	st_FriendRulue_ssah      //石首捱晃

	RandCardNum int //混牌次数
	//20200528 苏大强 荆门晃晃，
	SZHKZC bool //甩字胡可捉铳
	JZHKZM bool //见字胡可自摸
	//20200604 苏大强 试用 离线踢人
	OfflineKick st_OfflineKick

	st_FriendRule_tc_mj

	//麻城麻将
	st_FriendRule_mc_mj

	//乱将还是258做将 true258做将 false乱将
	JiangType bool

	ShaoSanZhang bool //少三张（发牌的时候不发13张牌，发10张牌 少三张） 潜江晃晃 潜江红中 潜江单挑

	LiangDaoType     int  //亮倒方式, 0全亮1部分亮
	ShangLouType     bool //上楼类型, 1豹子上楼2荒庄上楼
	ChaJiao          bool //false不查叫true查叫
	QuFeng           bool
	QuBaJiu          int
	MaimaClass       int
	MagiceContractor bool //20200928 苏大强 溪水麻将 全球人 癞子包胡
	Shaoyu12         bool //少于12张是否可以明牌（新孝感卡五星）
	AutoNext         bool
	KeChi            bool  //是否可吃牌
	HaiDiDouble      bool  //20201110 苏大强 海底加倍
	TrustPunish      bool  //20201111 苏大强 托管罚分
	TotalTimer       int64 //20210106 苏大强 累计时间

	MagicCardGHBP      bool //20220122 zwj 通山晃晃赖子杠后是否补牌
	Gmgh               bool //各摸各胡 zwj 荆州 揪马玩法 一脚赖油 荆州红中杠
	OfflineDismissTime int  //游戏开始后离线解散的时间，传秒数，如果不解散那就把这个时间传大一点
	ApplyDismissTime   int  //申请解散后的解散时间
	TrusteeCost        bool //托管玩家承担所有输分

}

//离线踢人
/*
20200604 苏大强 从以往经验，踢人有几个阶段
oninit 开桌阶段  面板有可能设置这个阶段踢人时间
setPao 下跑或者下漂阶段 基本上这个地方要么延续上面，要么就和游戏阶段用一个
onstartgame 游戏阶段阶段
*/
type st_OfflineKick struct {
	OfflineKick_init  int //开座阶段离线踢人时间
	OfflineKick_Piao  int //下跑下漂或者扑克那个中间选择阶段
	OfflineKick_start int //发牌后的时间

}

//石首捱晃
type st_FriendRulue_ssah struct {
	DiMa      int  //底码
	MaType    int  //配码方式
	IsAutoQGH int  //自动抢杠胡
	IsCHH     bool //是否超豪华
	IsChiHu   bool //是否吃胡
	IsDHJM    int  //大胡点炮是否奖码
}

//铳儿麻将
type st_FriendRulue_cr struct {
	QysDouble bool //清一色翻倍
	PphDouble bool //碰碰胡翻倍
	GkDouble  bool //杠开翻倍
	D7Double  bool //七对翻倍
	IsD7      bool //可胡七对
}

//阳新
type st_FriendRulue_yx struct {
	Need258   bool //: "true"二五八将，"false"乱将
	Ypkdh     bool //： 仰牌可大胡（"true"表示勾选）
	Rphznzm   bool //：软平胡只能自摸（"true"表示勾选）
	Hashhqd   bool //：豪华七对（"true"表示勾选）
	Hasgskh   bool //：杠上开花（"true"表示勾选）
	Hasqqr    bool //：全求人（"true"表示勾选）
	Haspph    bool //：碰碰胡（"true"表示勾选）
	Dpbbh     bool //点炮不包胡
	LimitTime int  //限制场时间
	SLBKK     bool //双赖不开口
	Qghfsjb   bool //抢杠胡分数减倍
}

//蕲春打拱
type st_FriendRule_qc_qcdg struct {
	ShowCardNum   bool `json:"showCardNum"`   //是否显示手牌数
	CardTypeScore bool `json:"cardtypescore"` //是否特殊牌型加分
	BombTypeSocre int  `json:"bombtypescore"` //是否炸弹有喜
}

type st_FriendRule_jl_jlmj struct {
	//！监利麻将
	HuType     int  //胡牌类型(1.自摸胡(默认) 2.点炮胡)
	BuHuaType  int  //补花类型(1.红中(默认) 2.红发白)
	WeiZeng    int  //围增(1分、2分、3分，默认1分)
	KengZhuang bool //是否坑庄
	ChuZeng    bool //是否出增(买炮或者下跑类似)
}

type st_FriendRule_hhs struct {
	//！晃晃类新增
	LockCard       bool   // 锁牌开关
	LockCard4      bool   // 4赖锁牌开关
	OnMagicAward   bool   // 飘癞子有奖开关
	OnlyBySelf     bool   // 是否只能自摸 仙桃麻将&剁刀可选项 一脚赖油只能自摸
	GkScoreType    int    // 0杠开不加分  1杠开加分  2杠开翻倍 其他 自定义
	DealerXian     bool   // 是否庄闲：庄家输赢+1分 剁刀可选项
	GangScoreAway  bool   // 杠分是否立即结算
	HardHuDoble    bool   // 硬胡是否x2 也就是是后有软硬胡之分 有得话硬胡分数x2
	FourMagicDoble bool   // 是否曾有四个癞子x2
	NoHuAfterMagic bool   // 是否场上有玩家打出癞子后 不能再捉统 晃晃合集玩家打出癞子后还可以接炮
	QGHScoreType   int    // 0抢杠胡算普通的放统 1固定算分 2杠分+胡牌分
	ShowGangScore  int    // 明杠/点杠扣多少分 （仙桃1分赤壁2分红中玩法3分）
	HuangHuangType int    // 0为硬晃 1为仙桃晃晃 2为晃晃合计 3为一赖可捉统 4为天门晃晃 5为其他
	TwoBigHu       bool   // 是否勾选双大胡
	GangDoubie     bool   // 杠的倍数是否翻倍
	XiXiangfeng    bool   // 是否勾选喜相逢
	XxfNum         int    // 喜相逢 倍数
	IWindCount     int    // 选择的风牌数量
	GameGongMask   byte   // 支持杠的类型
	DebugCardsName string // 做牌文件名字
	HuScore        int    // 指定放铳分
	SecondRoom     int    // -1代表无限制，9代表9秒场，15代表15秒场
	Trusttype      int    //20200813 苏大强 拖管类型：0不托管，1小局托管，2大局托管（仙桃添加（大局托管区别，就是小结算的时候不申请解散了））
}

type st_FriendRule_jy_yq struct {
	//！嘉鱼硬巧
	NoWind         bool //无风牌，可以勾选是否去风
	NoGangScore    bool //是否不计杠分
	PengPeng       bool //是否带碰碰胡
	IsPeiChong     bool //是否陪铳(点炮单付或点炮三人付)
	GenHuangzhuang bool //是否跟庄黄庄
	//20200205 苏大强 豪华7对
	HhQiDui bool
}

type st_FriendRule_tc_hzlzg struct {
	//! 通城红中赖子杠
	QiHuFan       int  //起胡番数
	QuFaCaiBaiBan bool //是否去发财白板
	ShuangDaHuFan int  //双大胡番数
	QiangShunGang bool //是否能抢点杠
	YingQGScore   int  //硬抢杠分数
	JianZiHuZiMo  bool //见字胡是否需要自摸
}

type st_FriendRule_tc_mj struct {
	//! 通城麻将
	Magic4Y7D bool `json:"silaiyingqidui"` //四癞子硬7对
}

//麻城麻将
type st_FriendRule_mc_mj struct {
	FengFan  bool //是否风番
	JiangFan bool //是否将番
	ManPao   bool //是否满跑
	Bi       bool //是否闭
	Cun      bool //是否存
}

type st_FriendRule_cz struct {
	//! 滁州
	BackCard bool //是否背牌
	NoFeng   bool //是否显示风牌
	Pei      bool //白赖子
	JiaDi    int  //头偏
}

type st_FriendRule_xn_xnhh struct {
	//！咸宁晃晃
	BaoPei    bool //是否包赔
	AllowGang bool //是否可杠
	MaiMa     bool //是否买马
}

type st_FriendRule_js_jsmj struct {
	//!京山麻将
	BGDY          bool //半干瞪眼
	PHDL          bool //平胡多癞
	QGDY          bool //全干瞪眼
	JoinAutoReady bool //入桌自动准备
}

type st_FriendRule_p3 struct {
	//！拼三
	Is235Enable bool // 是否有235特殊牌型
}

type st_FriendRule_hg_hzmj struct {
	//黄冈红中麻将
	FDiFen         float32 `difen"` //底分
	FengDingFanShu int     //封顶番数
}

type st_FriendRule_xsmj struct {
	//浠水麻将
	FBaseScore     float32 //胡型基础分
	MenQDL         bool    //门清多癞
	GangHBBP       bool    //杠后不补牌
	FengDingxx     int     //封顶选项
	Xs_YinDingMax  int     //银顶封顶点数
	Xs_JinDingMax  int     //金顶封顶点数
	Xs_FengDingMax int     //封顶点数
}

//黄石晃晃
type st_FriendRule_Mj_Hshh struct {
	Hshh_QiHuPoint      int  //起胡点数 bool
	Hshh_YinDingMax     int  //银顶封顶点数
	Hshh_JinDingMax     int  //金顶封顶点数
	HShh_KaiKou4_JIAFAN bool //开四口加番
	Hshh_Smart_TuoGuan  bool //智能托管
	Hshh_dlphkjp        bool //多癞屁胡可接炮
	Hshh_jzhphkjp       bool //见字胡屁胡可接炮

}

//汉川搓虾子
type st_FriendRule_hccxz_mj struct {
	Hc_12_SecondRoom bool // 是否12秒场
	HcWanfa          int  // 0一脚赖油 1半赖 2无赖到底
	Diangang         int  // 2点杠两倍 3点杠三倍
	FengDingFen      int  // 0不封顶 32分封顶 64分封顶
	Noyy             bool // 禁止语音
	No_hdbq          bool // 禁止互动表情
	No_qph           bool // 禁止俏皮话
	Fewerstart       bool // 可少人开局
	HuanSanZhang     bool // 是否换三张
	JianZiHu         bool // 是否可见字胡
}

//荆州搓虾子
type st_FriendRule_jzcxz_mj struct {
	Jz_SecondRoom      int  // 0无限制 12表示12秒场 15表示15秒场
	Jz_Wanfa           int  // 0一脚赖油 1半赖 2无赖到底
	Jz_Noyy            bool // 禁止语音
	Jz_No_hdbq         bool // 禁止互动表情
	Jz_No_qph          bool // 禁止俏皮话
	Jz_Fewerstart      bool // 可少人开局
	Jz_HuanSanZhang    bool // 是否换三张
	Jz_JianZiHu        bool // 是否可见字胡
	Jz_Gpsm            bool // 是否杠牌顺摸
	Jz_YouZhongYou     bool // 是否油中油
	Jz_Beishu          int  // 放杠倍数
	Jz_Gmgd            bool //各摸各的
	Overtime_offdiss   int  //离线解散时间
	Overtime_applydiss int  //申请解散时间
}

type st_FriendRule_K5X struct {
	MaiMaType int  //	买马类型
	QiHuX2    bool // 	2番起胡
	GangPaox4 bool //	杠上炮/花 x4
	ShuKan    bool //	数坎
	K5x4      bool //	卡五星 x4
	PPx4      bool //	碰碰胡 x4
	DLFF      bool //	对亮翻番
	GHFF      bool //	过胡翻番
	Fleetime  int  //	客户端传来的 游戏开始前离线踢人时间
	GangHuax4 bool //	杠上开花x4
	QuanPD    bool //   全频道
	HaiDix2   bool //	海底捞/炮 x2
}

// 罗田麻将
type st_FriendRule_lt_hzlzg struct {
	Lt_NoJinDing         bool // 无金顶
	Lt_NoFeng            bool // 去风
	Lt_QiangGang         bool // 抢杠胡
	Lt_FifteenSecondRoom bool // 十五秒场
	Lt_Kai2kouDpContract bool // 开两口点炮承包
	Lt_JinDingFengBao    bool // 金顶风豹
}

// 张家口番混儿
type st_FriendRule_zjk_fhe struct {
	Zjk_BaseScore int  // 胡牌底分
	Zjk_Advance   bool // 高倍玩法
}

// 咸宁红中杠
type st_FriendRule_xn_hzlzg struct {
	Xn_TimeOutPunish bool // 超时罚分
}

//广东推倒胡
type st_FriendRule_gd_tdh struct {
	Gdtdh_yongpai     int  //用牌
	Gdtdh_GangBaoQB   bool //杠爆全包
	Gdtdh_GenZhuang   bool //跟庄
	Gdtdh_JieJieGao   bool //节节高
	Gdtdh_HaiDiDouble bool //海底捞加倍
	Gdtdh_MaGenDiFen  bool //马跟底分（买马情况下有效）
	Gdtdh_MaGenGang   bool //马跟杠（买马情况下有效）
	Gdtdh_GuiPai      bool //鬼牌
	Gdtdh_NoGuiDouble bool //无鬼加倍(鬼牌情况下有效)
	Gdtdh_DianPaoKeHu bool //点炮可胡
	Gdtdh_ChiAble     bool //可吃牌

	Gdtdh_GuiPaiType  int     //鬼牌类型 0未勾选 1红中鬼 2白板鬼 3翻鬼 4翻双鬼 5 (4)花鬼 6 (8)花鬼
	Gdtdh_MaiMaNum    int     //买马类型 0未勾选 1买2马 2买4马 3买6马 4买8马 5一马全中
	Gdtdh_MaiMaDir    int     //买马方位 0未勾选 1胡牌方位 2定马159
	Gdtdh_ScoreDouble float32 //积分翻倍
}

//转转麻将
type st_FriendRule_mj_zz struct {
	Mjzz_GuoShouPeng bool //过手碰
	Mjzz_ForceHu     bool //强制胡牌
	Mjzz_XianZhuang  bool //闲庄玩法
	Mjzz_BuTongPao   bool //不通炮
	Mjzz_NoPeng      bool //不可碰
	//Gdtdh_MaiMaNum    int //买马数量 0 2 4 6
	//Gdtdh_MaiMaDir    int //买马方位 0胡牌方位中鸟 1 159鸟 2 一码全中
}

//鹤峰麻将焖癞子
type st_FriendRule_es_hfmj struct {
	FanLaiType  int  //翻癞子玩法
	LaiZiType   int  //癞子玩法
	FengDingPao bool //封顶可接炮
	TaiZhuang   bool //抬庄
	GangSH      bool //杠上花
	GangSP      bool //杠上炮
	QiangGH     bool //抢杠胡
	QuanQRBP    bool //全求人包赔
	DaHuJiePao  bool //大胡可接炮
	AnGangLai16 bool //暗杠癞子16倍
}

/*
20200604 苏大强 我认为离线踢人就3个阶段
以游戏阶段的时间为依据，如果其他的时间没设置，以游戏时间为准
*/
func (self *st_OfflineKick) SetOfflineKickTime(OfflineKick_start int, OfflineKick_Piao int, OfflineKick_init int) {
	//先判断游戏中解散时间有没有
	self.OfflineKick_start = OfflineKick_start
	if self.OfflineKick_start == 0 {
		self.OfflineKick_start = static.OFFLINE_ROOM_TIMER
	}
	//目前下跑阶段算游戏阶段，那么如果没有就按照游戏的时间来
	self.OfflineKick_Piao = OfflineKick_Piao
	if self.OfflineKick_Piao == 0 {
		self.OfflineKick_Piao = self.OfflineKick_start
	}
	//开桌的时间现在逐步都改为面板设置了，如果没有也同上
	self.OfflineKick_init = OfflineKick_init
	if self.OfflineKick_init == 0 {
		self.OfflineKick_init = self.OfflineKick_start
	}
}

//20200806 苏大强
/*
嘉鱼需求在小结算的时候，就算是选择了不能解散，如果选择了托管立即解散，还是能解散的，这个有点坑
我考虑生成一个mask，根据 现在调整一下
	GS_MJ_FREE   = iota          //空闲状态(局数够了，强退，解散，)，开局未下跑（漂）的情况。。。。这个时候好像可以离桌的，暂时保留吧
	GS_MJ_PAO     //下跑下漂阶段
	GS_MJ_PLAY    //游戏状态
	GS_MJ_HAI_DI  //海底状态
	GS_MJ_END     //结束状态(小结算)
移位判断mask里面是0（能解散）是1（不能解散）
未想到别的好方法，这个方法有个问题就是，需要根据GameStatus来判断，游戏中必须严格根据场景设置状态
参数就是一个切片吧，如果是nil，那么全状态都能申请解散，切片里面有哪个，哪个状态就不能申请解散
重点：一定要设置好GameEndStatus的状态，请使用这个func的时候，检查相关游戏的设置
*/
func (self *St_FriendRule) SetDissMissMask(setitem []uint) {
	//如果是空的，就不设置了
	if len(setitem) == 0 {
		return
	}
	self.DissMissMask = 0
	for _, v := range setitem {
		self.DissMissMask |= 1 << v
	}
}

func (self *St_FriendRule) CheckCanDissMiss(status byte) bool {
	if self.DissMissMask&(1<<uint(status-static.GS_PLAYING)) == 0 {
		return true
	}
	return false
}
