package rule

//打拱规则相关属性
type FriendRuleDG_qc struct {
	Difen            int    `json:"difen"`            //底分
	Radix            int    `json:"scoreradix"`       //底分基数
	Basescore        int    `json:"basescore"`        //变底编号（武穴510k）
	SerPay           int    `json:"revenue"`          //茶水
	Fa               int    `json:"fa"`               //没逃跑处罚倍数
	Jiang            int    `json:"jiang"`            //逃跑奖励别人的倍数
	Qiang            string `json:"qiang"`            //i是否可以抢庄，false不能抢庄，true可以抢庄
	FapaiMode        int    `json:"fapaims"`          //发牌模式，0表示起牌拱，1表示发牌拱，2表示疯狂拱，3表示变态拱
	HasPiao          string `json:"haspiao"`          //是否选漂
	HasBombstr       string `json:"hasbombstr"`       //是否有连炸
	HasBomberr       string `json:"hasbomberr"`       //是否炸错罚分
	NineSecondRoom   string `json:"sc9"`              //九秒场
	SkyCnt           int    `json:"skycnt"`           //花牌数目 ,0无花牌，
	ZhuangBuJie      string `json:"zhuangbujie"`      //庄家是否不接风，false表示接风，true表示不接风；未明鸡的对局，有一方跑了1游，如果接风为庄家的话。则由庄家下家接风
	FristOut         int    `json:"fristout"`         //首出类型 ,0黑桃3先出，1庄家首出
	BiYa             int    `json:"biya"`             //有大比压 ,0有大必压，1可以不压，
	TuoGuan          int    `json:"tuoguan"`          //托管 ,0不托管，大于0托管，
	ZhaNiao          string `json:"zhaniao"`          //是否扎鸟，红桃10是鸟
	FourTake3        string `json:"fourtake3"`        //是否可以4带3
	BombSplit        string `json:"bombsplit"`        //炸弹是否可拆
	QuickPass        string `json:"quickpass"`        //是否快速过牌
	SplitCards       string `json:"splitcards"`       //是否有切牌
	Bomb3Ace         string `json:"bomb3ace"`         //3个A是否是炸弹
	LessTake         string `json:"lesstake"`         //最后一手是否可以少带
	Jiao2King        string `json:"jiao2king"`        //双王必须叫牌
	TeamCard         string `json:"teamcard"`         //先跑可看队友牌
	Overtime_trust   int    `json:"overtime_trust"`   //托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //解散
	KeFan            string `json:"kefan"`            //是否可以反春
	Dissmiss         int    `json:"dissmiss"`         //解散次数，0不限制
	TrusteeCost      string `json:"trusteecost"`      //托管承担所有输分
	Hard510KOf4IsXi  string `json:"hard510kof4isxi"`  //4硬510K是否算喜
	CardNum          int    `json:"cardnum"`          //牌数
	FourTake2        string `json:"fourtake2"`        //是否可以4带2
	FourTake1        string `json:"fourtake1"`        //4带1是否算炸弹
	AddXiScore       string `json:"addxiscore"`       //是否带喜钱
	FleeTime         int    `json:"fleetime"`         //客户端传来的 游戏开始前离线踢人时间
	KingLai          int    `json:"kinglai"`          //王是否可做癞子，0无癞子，1有癞不分硬炸，2有癞分硬炸
	Big510k          string `json:"big510k"`          //6炸7炸不可打510k
	BombMode         string `json:"bombmode"`         //炸弹被压无分
	ExtAdd           string `json:"extadd"`           //额外加分，比如打出3王1花得1倍底分，跟打出8喜7喜的加分不一样
	Restart          string `json:"restart"`          //重新发牌
	Endready         string `json:"endready"`         //小结算是否自动准备
	Piao             int    `json:"piao"`             //选飘
	PiaoCount        int    `json:"piaocount"`        //0每局飘一次，1首局飘一次
	NotDismiss       string `json:"tuoguannotdiss"`   //托管不解散
	BombRestart      string `json:"bombrestart"`      //炸弹重新发牌
	FristOutMode     int    `json:"fristoutms"`       //首出出牌类型 ,0必带黑三或最小牌，1任意出牌
	NoBomb           string `json:"nobomb"`           //纯净玩法，勾选时不发炸弹
	//20200514 苏大强 武穴510k需要的
	AntiBrand          string `json:"antiBrand"`          //反牌
	Baojing            string `json:"baojing"`            //报警
	BombRealTime       string `json:"bombNofanbei"`       //炸弹实时计分
	OutCardDismissTime int    `json:"outcarddismisstime"` //出牌时间 超时房间强制解散 -1不限制

	LessTakeFirst string `json:"lesstakefrist"` //最后一手是否可以少带出完
	LessTakeNext  string `json:"lesstakenext"`  //最后一手是否可以少带接完

	//仙桃跑得快
	QiangChun string `json:"qiangchun"` //抢春

	//阳新二人拱
	FakeKing int `json:"fakeking"` //王单出算几

	//江汉跑得快
	Overtime_offdiss   int `json:"overtime_offdiss"`   //离线解散时间
	Overtime_applydiss int `json:"overtime_applydiss"` //申请解散时间

	//决战跑得快
	CardNum15   string `json:"cardnum15"`    //是否15张
	IsRed3First string `json:"red3firstout"` //是否红桃3首出
	ZhaNiaoFen  int    `json:"zhaniaofen"`   //2 2分   5 5分   10 10分   20 翻倍
	FengDing    int    `json:"fengding"`     //封顶，0不封顶

	LookonSupport string `json:"LookonSupport"` //本局游戏是否支持旁观

	TrustJuShu int `json:"trustjushu"` // 托管局数 不限制0
	TrustLimit int `json:"trustlimit"` // 托管限制 1 暂停 2 解散

	CardRecord string `json:"cardrecord"` //是否有记牌器，勾选为true表示可以买，可以用

	CalScoreMode int   `json:"calscoremode"` //计分模式，0 累计计分，1 不累计计分
	XiScoreMode  int   `json:"xiscoremode"`  //喜分模式
	TotalTimer   int64 `json:"TotalTimer"`   //累计时间，单位：分钟

	Card15 string `json:"card15"` // 是否为15张跑的快

	Must3 string `json:"must3"` // 3人玩法是否必须先出黑桃3
}

//通城打拱规则相关属性
type FriendRuleDG_tc struct {
	Difen            int    `json:"difen"`            //底分
	Radix            int    `json:"scoreradix"`       //底分基数
	Basescore        int    `json:"basescore"`        //变底编号（武穴510k）
	SerPay           int    `json:"revenue"`          //茶水
	Fa               int    `json:"fa"`               //没逃跑处罚倍数
	Jiang            int    `json:"jiang"`            //逃跑奖励别人的倍数
	Qiang            string `json:"qiang"`            //i是否可以抢庄，false不能抢庄，true可以抢庄
	FapaiMode        int    `json:"fapaims"`          //发牌模式，0表示起牌拱，1表示发牌拱，2表示疯狂拱，3表示变态拱
	HasPiao          int    `json:"haspiao"`          //是否选漂 0不漂 1漂1  2漂2
	HasBombstr       string `json:"hasbombstr"`       //是否有连炸
	HasBomberr       string `json:"hasbomberr"`       //是否炸错罚分
	NineSecondRoom   string `json:"sc9"`              //九秒场
	SkyCnt           int    `json:"skycnt"`           //花牌数目 ,0无花牌，
	ZhuangBuJie      string `json:"zhuangbujie"`      //庄家是否不接风，false表示接风，true表示不接风；未明鸡的对局，有一方跑了1游，如果接风为庄家的话。则由庄家下家接风
	FristOut         int    `json:"fristout"`         //首出类型 ,0黑桃3先出，1庄家首出
	BiYa             int    `json:"biya"`             //有大比压 ,0有大必压，1可以不压，
	TuoGuan          int    `json:"tuoguan"`          //托管 ,0不托管，大于0托管，
	ZhaNiao          string `json:"zhaniao"`          //是否扎鸟，红桃10是鸟
	FourTake3        string `json:"fourtake3"`        //是否可以4带3
	BombSplit        string `json:"bombsplit"`        //炸弹是否可拆
	QuickPass        string `json:"quickpass"`        //是否快速过牌
	SplitCards       string `json:"splitcards"`       //是否有切牌
	Bomb3Ace         string `json:"bomb3ace"`         //3个A是否是炸弹
	LessTake         string `json:"lesstake"`         //最后一手是否可以少带
	Jiao2King        string `json:"jiao2king"`        //双王必须叫牌
	TeamCard         string `json:"teamcard"`         //先跑可看队友牌
	Overtime_trust   int    `json:"overtime_trust"`   //托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //解散
	KeFan            string `json:"kefan"`            //是否可以反春
	Dissmiss         int    `json:"dissmiss"`         //解散次数，0不限制
	TrusteeCost      string `json:"trusteecost"`      //托管承担所有输分
	Hard510KOf4IsXi  string `json:"hard510kof4isxi"`  //4硬510K是否算喜
	CardNum          int    `json:"cardnum"`          //牌数
	FourTake2        string `json:"fourtake2"`        //是否可以4带2
	FourTake1        string `json:"fourtake1"`        //4带1是否算炸弹
	AddXiScore       string `json:"addxiscore"`       //是否带喜钱
	FleeTime         int    `json:"fleetime"`         // 客户端传来的 游戏开始前离线踢人时间
	KingLai          int    `json:"kinglai"`          //王是否可做癞子，0无癞子，1有癞不分硬炸，2有癞分硬炸
	Big510k          string `json:"big510k"`          //6炸7炸不可打510k
	BombMode         string `json:"bombmode"`         //炸弹被压无分
	ExtAdd           string `json:"extadd"`           //额外加分，比如打出3王1花得1倍底分，跟打出8喜7喜的加分不一样
	Restart          string `json:"restart"`          //重新发牌
	Endready         string `json:"endready"`         //小结算是否自动准备
	Piao             int    `json:"piao"`             //选飘
	PiaoCount        int    `json:"piaocount"`        //0每局飘一次，1首局飘一次
	NotDismiss       string `json:"tuoguannotdiss"`   //托管不解散
	BombRestart      string `json:"bombrestart"`      //炸弹重新发牌
	FristOutMode     int    `json:"fristoutms"`       //首出出牌类型 ,0必带黑三或最小牌，1任意出牌
	NoBomb           string `json:"nobomb"`           //纯净玩法，勾选时不发炸弹
	XiFenHaveOrOut   int    `json:"xifenhaveorout"`   //喜分打出算还是起到就算 0打出算  1起到算
	CardRecord       string `json:"cardrecord"`       //是否有记牌器，勾选为true表示可以买，可以用
	No6Xi            string `json:"no6xi"`            //6喜是否算喜分，勾选为true表示无6喜、不算喜分
	LookonSupport    string `json:"LookonSupport"`    //本局游戏是否支持旁观
}

//斗地主相关属性
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
	AntiSpring        string `json:"antispring"`        // 反春天
	FirstBanker       int    `json:"firstbanker"`       // 首局地主 0 随机叫  1 黑三先叫
	LookonSupport     string `json:"LookonSupport"`     //本局游戏是否支持旁观
	Endready          string `json:"endready"`          //小结算是否自动准备
	TrustJuShu        int    `json:"trustjushu"`        // 托管局数 不限制0
	TrustLimit        int    `json:"trustlimit"`        // 托管限制 1 暂停 2 解散
}

//打拱规则相关属性
type FriendRuleDG_ts_HY struct {
	Difen            int    `json:"difen"`            //底分
	KingNum          byte   `json:"kingnum"`          //最多几个王
	FakeKing         byte   `json:"fakeking"`         //王单出算几
	FapaiMode        int    `json:"fapaims"`          //发牌模式，0表示起牌拱，1表示发牌拱，2表示疯狂拱，3表示变态拱
	LongzhaDing      int    `json:"longzhading"`      //笼炸封顶
	NineSecondRoom   string `json:"sc9"`              //九秒场
	IpLimit          string `json:"ip"`               //ipgps限制
	Overtime_trust   int    `json:"overtime_trust"`   //超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //超时解散
	Radix            int    `json:"scoreradix"`       //底分基数
	TimeOutPunish    string `json:"overtime_score"`   // 超时罚分
	AddXiScore       string `json:"addxiscore"`       //是否带喜钱
	FleeTime         int    `json:"fleetime"`         // 客户端传来的 游戏开始前离线踢人时间
	HasBombstr       string `json:"hasbombstr"`       //是否有连炸
	CardRecord       string `json:"cardrecord"`       //是否有记牌器，勾选为true表示可以买，可以用
	LookonSupport    string `json:"LookonSupport"`    //本局游戏是否支持旁观
}

//打拱规则相关属性
type FriendRuleDG_xtqf_HY struct {
	Difen          int    `json:"difen"`          //底分
	MagicNum       byte   `json:"kingnum"`        //最多几个王
	FakeKing       byte   `json:"fakeking"`       //王单出算几
	GXScore        string `json:"gxscore"`        //贡献分 30-60
	AddSpecail     int    `json:"addspecial"`     //是否特殊牌型加分,0无喜钱，10表示1个喜钱10分，以此类推
	KanJie         int    `json:"kanjie"`         //坎阶梯：5分/坎，10分/坎，20分/坎
	KanScore       int    `json:"kanscore"`       //坎分：1-20，1，2，3，4....20
	BaoDi          int    `json:"baodi"`          //保底：5-100，5，10，15....100,默认5  ,
	Has510k        int    `json:"has510k"`        //510选项：510k是炸弹和510k不是炸弹，单选项，默认选中510k是炸弹，选择510k不是炸弹，癞子用法下面是禁用不能选中
	Magic510k      int    `json:"magic510k"`      //癞子用法：癞子可组510k和癞子不可组510k，单选项，默认选中癞子可组510k
	IpLimit        string `json:"ip"`             //ipgps限制
	Fleetime       int    `json:"fleetime"`       //客户端传来的 游戏开始前离线踢人时间
	Dissmiss       int    `json:"dissmiss"`       //解散次数,0不限制,12345对应限制次数
	Overtime_trust int    `json:"overtime_trust"` //托管
	LookonSupport  string `json:"LookonSupport"`  //本局游戏是否支持旁观
}

//仙桃千分必打规则相关属性
type FriendRuleDG_xtqfbd_HY struct {
	Difen            int    `json:"difen"`            //底分
	Radix            int    `json:"scoreradix"`       //底分基数
	IpLimit          string `json:"ip"`               //ipgps限制
	IsTwoKingBomb    int    `json:"isTwoKingBomb"`    //双王是否可炸 1可炸 0不可炸
	IsPressZha       int    `json:"isPressZha"`       //同花510K是否可以压4炸 1可以 0不可压
	Xifen            string `json:"xifen"`            //是否有喜分，true有 false没有
	Zongzha          string `json:"zongzha"`          //是否有总炸，true有 false没有
	Overtime_trust   int    `json:"overtime_trust"`   //托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //解散
	Endready         string `json:"endready"`         //小结算自动准备
	Sjhz             string `json:"randomseat"`       //每局换座
	Ismustmax        string `json:"ismustmax"`        //下家报双必须出最大
	Dissmiss         int    `json:"dissmiss"`         //解散次数，0不限制
	LookonSupport    string `json:"LookonSupport"`    //本局游戏是否支持旁观
}

//字牌规则相关属性
type FriendRuleZP_dyzp struct {
	Difen              int    `json:"difen"`              //底分
	Radix              int    `json:"scoreradix"`         //底分基数
	SerPay             int    `json:"revenue"`            //茶水
	DianPaoPei         string `json:"dianpaopei"`         //点炮包赔
	DianPaoBiHu        string `json:"dianpaobihu"`        //点炮必胡
	PeiChong           int    `json:"peichong"`           //数值型的陪铳（点炮包赔）
	HeiKa              string `json:"heika"`              //黑卡6番
	HunJiang           string `json:"hunjiang"`           //是否有混江
	JiaDiFen           int    `json:"jiadifen"`           //结算加的底分
	FengDing           int    `json:"fengding"`           //封顶 0不封顶
	Overtime_trust     int    `json:"overtime_trust"`     //超时托管
	Overtime_dismiss   int    `json:"overtime_dismiss"`   //超时解散
	TrustJuShu         int    `json:"trustjushu"`         //托管局数 不限制0
	TrustLimit         int    `json:"trustlimit"`         //托管限制 1 暂停 2 解散
	Overtime_offdiss   int    `json:"overtime_offdiss"`   //离线解散时间
	Overtime_applydiss int    `json:"overtime_applydiss"` //申请解散时间

	GeziShu        int    `json:"gezishu"`  //个子数
	HuaShu         int    `json:"huashu"`   //花数 ，10表示10个花，1表示溜花
	Piao           int    `json:"piao"`     //选漂:0不漂，100带漂，10定漂
	DingPiao       int    `json:"dingpiao"` //1-3定漂
	DuoHu          string `json:"duohu"`    //true 一炮多响
	BeiShu         int    `json:"beishu"`   //倍数
	QuanHei        int    `json:"quanhei"`  //全黑倍数
	NineSecondRoom string `json:"sc9"`      //九秒场

	HunJiangS          int    `json:"hunjiangs"`          //是否有混江,数字型,1有
	NoOut              string `json:"noout"`              //true不出牌（捏牌）
	KeChong            string `json:"kechong"`            //true放铳玩家出分，另外一个人不出；false放铳和另一个人都出自己的分
	FenType            int    `json:"fentype"`            //算分类型,数字型,0算胡数，1算坡数，2登庄
	Fleetime           int    `json:"fleetime"`           //客户端传来的 游戏开始前离线踢人时间
	Dissmiss           int    `json:"dissmiss"`           //解散次数,0不限制,12345对应限制次数，-1不可解散
	LessOnePlayerCard  string `json:"lessoneplayercard"`  //2人模式少一人牌
	Qiang              int    `json:"qiang"`              //枪牌，0上七八可，1夹夹枪，2见红枪
	SitAutoReady       string `json:"sitautoready"`       //入桌自动准备
	RoundOverAutoReady string `json:"roundoverautoready"` //小局结束自动准备

	MinHuxi            int    `json:"minhuxi"`            //起胡限制，30，50
	Peng5              string `json:"peng5"`              //5把不碰
	OneKan             string `json:"onekan"`             //独坎不拆
	NoShengHua         string `json:"noshenghua"`         //不可落地生花
	LastPai            string `json:"lastpai"`            //末张不胡
	ManChi             string `json:"manchi"`             //满吃满妥
	AddWuJiang         string `json:"addwujiang"`         //无将翻番，翻将玩法时如果胡牌玩家无将需要翻番
	KeZhaoHu           string `json:"kezhaohu"`           //打拱可开招胡
	NoJiang            string `json:"nojiang"`            //不翻将，为true表示没有将牌
	AddPao             int    `json:"addpao"`             //放跑加分,0放炮不加分，1放炮加1分，类推
	Sjhz               string `json:"sjhz"`               //每局换座
	HouShao            string `json:"houshao"`            //后绍必报
	HasHaidi           string `json:"haidi"`              //海底捞
	KeHuBuZhui         string `json:"kehubuzhui"`         //可胡不追，勾选了，就是 打出的牌不能再吃，不能再抵口，可以再胡，没有勾选就也不能胡
	ShengDui           string `json:"shengdui"`           //落地生花可生对，勾选了表示可以生对
	NoHuanGuan         string `json:"nohuanguan"`         //不可换观胡，勾选了表示不能换观胡
	DengZhuang         string `json:"dengzhuang"`         //登庄，false：胡牌玩家的下家当庄，true：胡牌玩家当庄。第一局随机，流局连庄
	LookonSupport      string `json:"LookonSupport"`      //本局游戏是否支持旁观
	ShowTingTag        string `json:"showtingtag"`        //是否显示听牌角标，勾选了表示显示
	Left10             string `json:"left10"`             //剩余10张流局
	LimitPeng          string `json:"limitpeng"`          //限对，限制只能碰2次
	TotalTimer         int64  `json:"TotalTimer"`         //累计时间，单位：分钟
	OutCardDismissTime int    `json:"outcarddismisstime"` // 出牌时间 超时房间强制解散 -1不限制
}

//阳新字牌规则相关属性
type FriendRuleZP_yxzp struct {
	Difen          int    `json:"difen"`         //底分
	DianPaoPei     string `json:"dianpaopei"`    //点炮包赔
	DianPaoBiHu    string `json:"dianpaobihu"`   //点炮必胡
	NineSecondRoom string `json:"sc9"`           //15秒场
	LookonSupport  string `json:"LookonSupport"` //本局游戏是否支持旁观
}

//洪湖510K规则相关属性
type FriendRule510k_ts_HongHu struct {
	Difen            int    `json:"difen"`            //底分
	Laizi            byte   `json:"laizi"`            //王单出算几 0单出3 1单出2
	NineSecondRoom   string `json:"sc9"`              //九秒场
	IpLimit          string `json:"ip"`               //ipgps限制
	Fleetime         int    `json:"fleetime"`         //客户端传来的 游戏开始前离线踢人时间
	Overtime_trust   int    `json:"overtime_trust"`   //超时托管
	Overtime_dismiss int    `json:"overtime_dismiss"` //超时解散
	Radix            int    `json:"scoreradix"`       //底分基数
	Endready         string `json:"endautoready"`     //小结算是否自动准备
	DanShu           string `json:"kechudanshun"`     //单顺是否可选
	WanFa            int    `json:"wanfa"`            //玩法 0 165小光（默认） 1 155小光
}
