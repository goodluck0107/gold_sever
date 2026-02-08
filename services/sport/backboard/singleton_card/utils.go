package card_mgr

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

const (
	//常量定义
	MAX_WEAVE     = 4    //最大组合   （估计是杠的最大数）
	MAX_INDEX     = 34   //最大索引   （牌型索引）
	MAX_COUNT     = 14   //最大数目 （手牌最多数目）
	MAX_REPERTORY = 136  //最大库存(内地麻将个数)
	MAX_GUI_NUM   = 2    //最大鬼牌类数
	MASK_VALUE    = 0x0F //数值掩码
	GOD_CARD      = 0xFF //万能牌
	MASK_GODNUM   = 8    //目前赖子支持到8个

	PICARD  = 1      //皮牌
	GODCARD = 1 << 2 //赖子（万能）
	EYECARD = 1 << 3 //将牌

	//因为没看到go有word型，所有分两段 取消“自己”，底八位就为自己，高8位为他人，特殊的（上家为最高为），记录为uint16型
	/*
		设定判断牌的来源
	*/
	ORIGIN_NOMOR  = 0xff
	ORIGIN_NOM    = 1      //发牌区   这个就是自摸的状态  自己的
	ORIGIN_HAND   = 1 << 3 //手牌区    这个主要是蓄杠的状态，胡牌里不用这个  自己的
	ORIGIN_WARN   = 1 << 2 //倒牌区   这个主要是用于抢杠的状态 别人的
	ORIGIN_TABILE = 1 << 1 //桌面区    吃胡状态 别人的
	// ORIGIN_OWN    = 1 << 5 //自己
	ORIGIN_RIGHT = 1 << 8 //上家
	MASK_HU      = 0xf

	CTX_INITIAL    = iota //天听状态（天胡和地胡）
	CTX_NOM               //普通状态
	CTX_HD                //海底捞
	CTX_GANG              //杠   抢杠胡用
	CTX_GANG_AFTER        // 杠后     热铳用

	/*
			抢杠 不知道赖子能不能杠，但是可能出现在杠位上
		杠后 （别人杠后 打的热铳，和自己杠起的杠上花）
		别人的
		自己的
		设定万用，必然为1
		目前崇阳放放 11101
	*/
	GUI_OWN   = 1
	GUI_OTHER = 1 << 2

	HARD_258 = 1 // 必须是258的对将
	//下面这两种必然是有赖子的，258也可能是赖子的情况要考虑
	FART_258 = 1 << 1 // 1张258+赖子
	NO_258   = 1 << 2 // 纯赖子对替换258

)

var (
	CardClassMap = map[int]string{
		EYECARD: "将牌",
		PICARD:  "皮牌",
		GODCARD: "赖子",
	}
	CardOriginMap = map[int]string{
		ORIGIN_NOM:    "自摸",
		ORIGIN_HAND:   "手牌区",
		ORIGIN_WARN:   "抢杠",
		ORIGIN_TABILE: "吃胡",
		// ORIGIN_OWN:    "自己",
		ORIGIN_RIGHT: "上家",
	}
	CardCTXMap = map[int]string{
		CTX_INITIAL:    "天听",
		CTX_NOM:        "普通状态",
		CTX_HD:         "海底捞",
		CTX_GANG:       "杠",
		CTX_GANG_AFTER: "杠后",
	}
)

/*
20181129 根据杠牌的需求，回头杠和蓄杠要区别，这个单元修改，将IsDraw改为byte类型，名称改为来源
20190108 苏大强 考虑新的方案
*/
//--------------这段数据将来放到handInfo里面 目前碰杠单元还没归类，先不动
type GroupInfo struct {
	FourCards  []byte
	ThreeCards []byte
	TwoCards   []byte
	OneCards   []byte
}

//构想中，不完善
//20190219 重新设计
type WarnCTXinfo struct {
	colorInfo  byte //最多5位 0万、1条、2筒、3风、4箭
	is258      byte
	ChipaiInfo []byte //最多4个，有的话放入最左值，吃牌都是人家的 ，吃牌可能会算什么三步高啊
	PengInfo   []byte //20190219添加，因为7对的判断
	Triplet    []byte //明杠
	HidTriplet []byte //暗杠
}

//以自己为视角，只能得到 自己（起牌区）别人（抢杠，桌面区）
type CardBaseInfo struct {
	ID     byte //牌值
	Origin byte //隶属
}
type CardEx struct {
	ID        byte
	CardClass int //牌类型：普通、皮子、赖子、将 这个是根据mask设定的
	// IsDraw     bool //丢牌区（来源自己、来源别人）保留
	IsGod bool //20190117 要么是还原牌，要么是鬼牌
	// CheckWight byte //检查权重 目前是用来检查是不是当鬼牌 只用到最后两位   遗留属性，新改版后废弃 保留，判胡用的，实际可用废弃
}

//20190108 根据想发，这个地方就是牌的可执行的操作，根据游戏规则设定，必须属性是牌的class
//20190109 目前只是与预想的，添加一个god牌设定，天门放放里面，如果丢赖子，赖子要还原，自己起的就是god 追加一个god设置和Origin一起来判断
// type CardCTX struct {
// 	CardClass  int  //牌类型：普通、皮子、赖子、将
// 	operator   byte //可支持操作吃、碰、明杠、暗杠、续杠、回头杠、吃胡、自摸 刚好8个
// 	OriginMask byte //和cardbaseInfo里面的Origin一起设定god牌的适用范围，目前判胡就需要判断在发牌区和丢牌区
// }

//20190116
type BaseCTX struct {
	CbCardIndex []byte                //用户总共有多少张手牌 34张格局 吃碰杠的牌不算
	WeaveItem   []static.TagWeaveItem //倒牌信息
	Checkcard   *CardBaseInfo         //要检查的牌(来源：自己摸的、桌面上的)
}
type GodCardOrg struct {
	checkMask uint
	//20190116 目前判断的牌，会有几种情况，
	guicards []byte //赖子，最多应该只有2，这里还是切片吧
}

//20190305 越来越多的开关，导致越来越多的动态限制，已经不是简单的处理了
/*
嘉鱼红中赖子杠规则
设定，当选择一赖到底的时候
RestrictGodItem{
Restrict_GodNum :1     //限制只能有一张鬼牌，不管怎么用
Restrict_NeedGodNum: 0   //为0的时候不限制
}
如果是不选一赖到底
RestrictGodItem{
Restrict_GodNum :0   //不限制鬼牌数量
Restrict_NeedGodNum: 1   //限制当鬼牌的数量为1 这个需要判断后处理
}
*/
type RestrictGodItem struct {
	//目前这个限定成了大胡和屁胡通管了
	Restrict_GodNum     byte //限定胡牌时赖子的个数 嘉鱼红中赖子杠规则，动态设定的，如果选择一赖到底只能有一个赖子，不管是不是还原
	Restrict_NeedGodNum byte //嘉鱼红中赖子杠规则，动态设定的，如果不选择一赖到底可多个赖子，但是只能有一张当鬼
	//先加个临时的 就是一个开关
	Restrict_Fan bool
}

//准备7对mask 和 见字胡、甩字胡都丢这
type ExpandMask struct {
	Restrict_GodItem *RestrictGodItem
}

/*
//20190301 新增3n+2返回结果  这里有个问题，即可以硬胡又可以屁胡的情况返回的问题，要么是赖将，要么是3章赖子，区别就在于剩下的赖子数
*/
//20190116 参数修改为手牌，ctx信息(未想好怎么用)，牌信息
//20190220 调整，创建查胡环境的时候，去掉检查牌的信息，目前查胡表里的见字胡是直接加牌检查的
func NewHandBaseInfo(cbCardIndex []byte, weaveItem []static.TagWeaveItem, cardInfo *CardBaseInfo, ctxScene int) (newHand *BaseCTX, err error) {
	if len(cbCardIndex) != static.MAX_INDEX || len(weaveItem) > 4 || ctxScene == 0 {
		return nil, errors.New(fmt.Sprintf("创建手牌失败，以下条件有一个不符合：手牌序列长度（%d）必须是（%d）倒牌（%d）组最多4组,场景mask（%d）要大于0", len(cbCardIndex), static.MAX_INDEX, len(weaveItem), ctxScene))
	}
	//暂时就检查下这个牌是不是可用的
	if cardInfo.ID != static.INVALID_BYTE {
		if cardInfo != nil && !mahlib2.IsMahjongCard(cardInfo.ID) {
			return nil, errors.New(fmt.Sprintf("要检查的牌不是内地麻将牌型(%x)", cardInfo.ID))
		}
	} else {
		cardInfo = nil
	}

	return &BaseCTX{
		WeaveItem:   weaveItem,
		CbCardIndex: cbCardIndex,
		Checkcard:   cardInfo,
	}, nil
}
func NewExpandMask(Restrict_GodItem *RestrictGodItem) *ExpandMask {
	return &ExpandMask{
		Restrict_GodItem: Restrict_GodItem,
	}
}

func NewRestrictGodItem(restrictGodNum byte, restrictNeedGodNum byte, restrictfan bool) *RestrictGodItem {
	if restrictGodNum > MASK_GODNUM || restrictNeedGodNum > MASK_GODNUM {
		return nil
	}
	return &RestrictGodItem{
		restrictGodNum,
		restrictNeedGodNum,
		restrictfan,
	}
}

//平铺已经成型的牌 关于杠到底要铺几张要根据实际情况来决定
//20181226这个是以前的构思，先放着，目前不用
func CreateWeaveCard(WeaveItem []static.TagWeaveItem) (weaveCard []byte, err error) {
	if len(WeaveItem) == 0 {
		return nil, nil
	}
	cbWeaveCount := byte(len(WeaveItem))
	if cbWeaveCount > 4 {
		return nil, errors.New(fmt.Sprintf("倒牌组合越界%d", cbWeaveCount))
	}
	for _, v := range WeaveItem {
		if v.WeaveKind == static.WIK_NULL {
			continue
		}
		switch v.WeaveKind {
		case static.WIK_LEFT: //左吃类型
			weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard+1)
			weaveCard = append(weaveCard, v.CenterCard+2)
		case static.WIK_CENTER: //中吃类型
			weaveCard = append(weaveCard, v.CenterCard-1)
			weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard+1)
		case static.WIK_RIGHT: //右吃类型
			weaveCard = append(weaveCard, v.CenterCard-2)
			weaveCard = append(weaveCard, v.CenterCard-1)
			weaveCard = append(weaveCard, v.CenterCard)
		case static.WIK_PENG: //碰牌类型
			weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard)
		case static.WIK_GANG: //杠牌类型 杠牌算3张，这样算牌的时候就能14了
			// weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard)
			weaveCard = append(weaveCard, v.CenterCard)
		default:
			// fmt.Println("CreateWeaveCard 未知类型", v.WeaveKind)
		}
	}
	return weaveCard, nil
}

//20181226 简单一点，不验证参数
func NewTagWeaveItem(weaveKind byte, centerCard byte, publicCard byte, provideUser uint16) *static.TagWeaveItem {
	return &static.TagWeaveItem{
		WeaveKind:   weaveKind,
		CenterCard:  centerCard,
		PublicCard:  publicCard,
		ProvideUser: provideUser,
	}
}

/*
//目前只设定自己和其他人，应该是够了，特殊的上家应该就归为其他人
//就只有4个位置（桌面、倒牌、手牌、起牌）
两个参数分别代表自己和他人
*/
func CreateGuiCheckMask(otherOrigin byte, ownOrigin byte) uint {
	return uint(otherOrigin<<8 | ownOrigin)
}

//20190109 考虑到赖子凑牌的情况。。。
func ClassifyCards(checkcard []byte) *GroupInfo {
	maxinde := len(checkcard)
	if maxinde > static.MAX_INDEX {
		return nil
	}
	result := &GroupInfo{
		FourCards:  []byte{},
		ThreeCards: []byte{},
		TwoCards:   []byte{},
		OneCards:   []byte{},
	}
	for i := 0; i < maxinde; i++ {
		switch checkcard[i] {
		case 4: //最多3组
			result.FourCards = append(result.FourCards, mahlib2.IndexToCard(byte(i)))
		case 3: //最多4组
			result.ThreeCards = append(result.ThreeCards, mahlib2.IndexToCard(byte(i)))
		case 2: //最多7组
			result.TwoCards = append(result.TwoCards, mahlib2.IndexToCard(byte(i)))
		case 1: //正常3n+2的话，最多12章
			result.OneCards = append(result.OneCards, mahlib2.IndexToCard(byte(i)))
		}
	}
	return result
}

//这里的左吃右吃表面被吃牌的位置
func GreatDrwaForShun(color byte, value byte, exposure byte) []byte {
	//只有3种
	switch exposure {
	case static.WIK_RIGHT: //右吃类型
		return []byte{mahlib2.CombinCard(color, value-2), mahlib2.CombinCard(color, value-1)}
	case static.WIK_CENTER: //中吃类型
		return []byte{mahlib2.CombinCard(color, value-1), mahlib2.CombinCard(color, value+1)}
	default: //左吃类型
		return []byte{mahlib2.CombinCard(color, value+1), mahlib2.CombinCard(color, value+2)}
	}
}

//20190301 获取成3中的需要的牌，以中心扑克为准，成顺子，最多3组合，最少1组，4张
/*
，以中心扑克为准，成顺子，最多3组合，最少1组，4张;成刻，就一个组合，一张牌
*/

func GetDrawCards(DrawItem [][]byte) (err error, result []interface{}) {
	if len(DrawItem) == 0 {
		return nil, nil
	}
	var value []byte
	for _, v := range DrawItem {
		value = append(value, v...)
	}
	return mahlib2.SliceRemoveDuplicate(value)
}
func FindDrawForShun(centerCard byte) (error, *[][]byte) {
	if centerCard == 0 || centerCard > 0x37 {
		return errors.New(fmt.Sprintf("FindDrawForShun牌编号最大值（%b），传入值（%b）", static.MAX_INDEX, centerCard)), nil
	}
	if centerCard > 0x29 {
		return nil, nil
	}
	//剩下的就是0~9的
	index := centerCard >> 4
	value := centerCard & static.MASK_VALUE
	var tempTiem [][]byte
	switch value {
	case 1:
		//只有左吃
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_RIGHT))
	case 2:
		//中吃和左吃
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_CENTER))
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_LEFT))
	case 8:
		//右吃和中吃
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_RIGHT))
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_CENTER))
	case 9:
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_RIGHT))
	default:
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_RIGHT))
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_CENTER))
		tempTiem = append(tempTiem, GreatDrwaForShun(index, value, static.WIK_LEFT))
	}
	return nil, &tempTiem
}

func GetDrawForKe(centerCard byte) (error, *[]byte) {
	if centerCard == 0 || centerCard > 0x37 {
		return errors.New(fmt.Sprintf("GetDrawForKe牌编号最大值（%b），传入值（%b）", static.MAX_INDEX, centerCard)), nil
	}
	return nil, &[]byte{centerCard}
}

//重新确定需要个鬼牌，因为判胡只是判断每行3n+2需要多少鬼牌，会出现每行都是eye的情况，这个时候需要补赖子，普通胡里面只有一对eye
//20190116 赖子将要告知
//20190124 对于孤章+赖子组成的eye 统计有误
//20190305  这里处理一下 god的数量，如果限制了
func RestoreNeedGui(hu_struct *[4][2]byte) (guieye bool, needguinum byte) {
	Needgui, eyecolor := GetGuiNumAndEyeClorer(hu_struct)
	//eyecolor也有可能没有。。。。就是纯赖子将
	guieye = false
	switch len(eyecolor) {
	case 1:
		//正常,不排除是孤章+赖子组成的eye
	case 0:
		//赖子将
		Needgui += 2
		guieye = true
	default:
		//这里面也有可能存在孤章+赖子组成的eye 或者没用1+3这种
		Needgui += byte(len(eyecolor) - 1)
	}
	//外面校验一下
	return guieye, Needgui
}

//这个就只能按照最恶心的方式拆解了，最终就是手牌
func DecreaseGui(handCards []byte, guiCards []byte, checkCards byte, isNomail bool) (checkHandCards []byte, guiCardsNum byte) {
	guiCardsNum = 0
	static.HF_DeepCopy(&checkHandCards, &handCards)
	if len(guiCards) != 0 {
		for _, v := range guiCards {
			guiIndex, _ := mahlib2.CardToIndex(v)
			guiCardsNum += checkHandCards[guiIndex]
			checkHandCards[guiIndex] = 0
		}
	}
	if isNomail {
		index, _ := mahlib2.CardToIndex(checkCards)
		checkHandCards[index]++
	} else {
		guiCardsNum++
	}
	return checkHandCards, guiCardsNum
}

//独立处理清一色 将一色 功能
func GetWarninfo(weaveItem []static.TagWeaveItem) *WarnCTXinfo {
	if len(weaveItem) == 0 {
		return nil
	}
	newinfo := &WarnCTXinfo{}
	//一次拿完所有的东西
	newinfo.is258 = 1
	haveItem := 0
	for _, v := range weaveItem {
		//颜色
		newColor, newValue := mahlib2.DifferentColor(v.CenterCard)
		if (newColor == newValue) && (newValue == 0) {
			//没有数据
			break
		}
		haveItem++
		newinfo.colorInfo |= newColor
		if newinfo.is258 == 1 {
			if newColor&(mahlib2.CARDS_WITHOUT_BAMBOO|mahlib2.CARDS_WITHOUT_WAN|mahlib2.CARDS_WITHOUT_DOT) == 0 {
				newinfo.is258 = 0
			}
		}
		//记录顺子的左牌 进这个结构的weave重新整理了
		switch v.WeaveKind {
		case static.WIK_LEFT, static.WIK_CENTER, static.WIK_RIGHT:
			newinfo.ChipaiInfo = append(newinfo.ChipaiInfo, v.CenterCard)
			newinfo.is258 = 0
		case static.WIK_PENG:
			newinfo.PengInfo = append(newinfo.PengInfo, v.CenterCard)
		case static.WIK_GANG:
			if v.PublicCard != 0 {
				newinfo.Triplet = append(newinfo.Triplet, v.CenterCard)
			} else {
				newinfo.HidTriplet = append(newinfo.HidTriplet, v.CenterCard)
			}
		}
		if newinfo.is258 == 1 {
			if newValue != 2 && newValue != 5 && newValue != 8 {
				newinfo.is258 = 0
			}
		}
	}
	if haveItem > 0 {
		return newinfo
	}
	return nil
}
func GetAllCardItembase(checkCardItem []byte, weaveItem []static.TagWeaveItem) (result byte, is258 bool) {
	var hand258 byte = 0
	var handCardNum byte = 0
	result, hand258, handCardNum = GetHandCardColorandNum(checkCardItem)
	//20190221 去掉赖子的时候 手牌就不一定是14了
	if handCardNum != MAX_COUNT && weaveItem != nil {
		//获取组合中的东西
		reWarnCTXinfo := GetWarninfo(weaveItem)
		if reWarnCTXinfo != nil {
			result = result | reWarnCTXinfo.colorInfo
			hand258 &= reWarnCTXinfo.is258
		}
	}
	return result, hand258 == 1
}

func GetHandCardColorandNum(checkCardItem []byte) (result byte, is258 byte, num byte) {
	checkInfo := GetHandCardNumSP(checkCardItem)
	if checkInfo == nil {
		return 0, 0, 0
	}
	for index, v := range checkInfo {
		if v != 0 {
			result |= 1 << uint(index)
		}
		num += v
	}
	//检查258的数据有多少
	var context byte = 0
	// common.Print_cards(this.CheckCardItem[:])
	for i := 1; i <= 25; i += 3 {
		if checkCardItem[i] != 0 {
			context += checkCardItem[i]
		}
	}
	if num-context == 0 {
		return result, 1, num
	}
	return result, 0, num
}

func GetHandCardNumSP(checkCardItem []byte) (result []byte) {
	var context byte = 0
	//万数据
	for i := 0; i < 9; i++ {
		context += checkCardItem[i]
	}
	result = append(result, context)
	//条判断
	context = 0
	for i := 9; i < 18; i++ {
		context += checkCardItem[i]
	}
	result = append(result, context)
	//筒判断
	context = 0
	for i := 18; i < 27; i++ {
		context += checkCardItem[i]
	}
	result = append(result, context)
	//风牌判断
	context = 0
	for i := 27; i < 31; i++ {
		context += checkCardItem[i]
	}
	result = append(result, context)
	//箭牌判断
	context = 0
	for i := 31; i < 34; i++ {
		context += checkCardItem[i]
	}
	result = append(result, context)
	return result
}

//只有赖子是万用的目前
//20190708  修改，出现检查牌可能是255的情况
func CreateNewCheckCards(bCardIndex []byte, checkCard byte, isNormalCard bool, guiCards []byte, checkNoMagic bool) (checkCards []byte, godNum byte) {
	checkCards = make([]byte, 0)
	godNum = 0
	cardIndex := len(bCardIndex)

	if cardIndex == 0 || cardIndex > 34 {
		return
	}
	checkCards = make([]byte, 0)
	static.HF_DeepCopy(&checkCards, &bCardIndex)
	index, _ := mahlib2.CardToIndex(checkCard)
	if index != static.INVALID_BYTE {
		if checkNoMagic {
			//硬胡检查，直接加
			checkCards[index]++
		} else {
			//if isNormalCard{
			//	checkCards[index]++
			//}
			if len(guiCards) != 0 {
				for _, v := range guiCards {
					gui1index, err := mahlib2.CardToIndex(v)
					if err != nil {
						return nil, 0
					}
					godNum += checkCards[gui1index]
					checkCards[gui1index] = 0
				}
			}
			if isNormalCard {
				checkCards[index] += 1
			} else {
				godNum += 1
			}
		}
	} else {
		//20200305 苏大强 阳新麻将有起始亮倒这个，如果起牌是中发白，sendcard就必须是255了，追加一个
		if len(guiCards) != 0 {
			for _, v := range guiCards {
				gui1index, err := mahlib2.CardToIndex(v)
				if err != nil {
					return nil, 0
				}
				godNum += checkCards[gui1index]
				checkCards[gui1index] = 0
			}
		}
	}

	return
}

func GetCardNum(card []byte) (num byte) {
	for _, v := range card {
		num += v
	}
	return num
}
