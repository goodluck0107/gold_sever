package card_mgr

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

/*
//20190305  这些结构需要重新分层了
*/

type CheckHuCTX struct {
	piZiCard byte //记录 每小局都不一样，目前还未使用
	Mask258  byte //记录 至少大局可设定
	//20190125 因为7对会有豪华、2豪华和3豪华类别，这个字段暂时就是因为这个添加的
	//20190213 修改字段分配，分上8和下8，上8的低4位保存7对的属性，下8的低4位保存可胡牌来源，因为武汉晃晃当遇到见字胡的时候会限制吃胡
	Mask_special uint64 //记录 至少大局可设定
	//20190118 设置大胡的可能性，硬胡，软胡，如果都支持的话check胡就要算两次 就一个bool就行，表示是不是要检查屁胡番
	CheckMFan bool     //记录 至少大局可设定
	CheckCtx  *BaseCTX //传进来的基本数据		//记录 每次检查的时候不一样
	//20190109 设置为结构，兼容2类赖子牌的情况
	GodInfo *GodCardOrg //记录 每小局都不一样，其中的godMask还没确定情况
	//20190213 添加见字胡mask 这个用在查胡表中使用
	GodHuMask uint64 //记录 至少大局可设定
	//20190214 查番用的
	FanGodHuMask uint64 //记录 至少大局可设定会依托GodHuMask
	//--------------以上为检查的基础数据
	ChiHuKind     uint64   //3n+2或者特殊胡都会影响
	CheckCardEX   *CardCTX // 记录 每次判断都不一定一样的
	CheckCardItem []byte   //这个是用来检查的，每个番型要重新赋值 check牌已经加进去了   	// 记录 每次判断都不一定一样的
	//20190117 追加新的属性，eye和皮子牌，目前不知道这两种牌有没有特殊处理，就不写新单元了，只是牌值设置需要用一下
	//创建检查对象的时候，会根据GodInfo来创建这个对象
	CheckGodOrg *CheckGodOrg //每次检查都会影响
	//20190117 这个reweaveInfo在手牌的组成部分可能会有N多种不同的组合，目前考虑倒牌的weave是定死的，因为操作的时候会让玩家选择
	ReweaveInfo []*static.TagWeaveItem
	//20190221 创建检查对象的时候，如果有weave结构就创建这个
	ReWarnCTXinfo *WarnCTXinfo
	//20190301 从甩字胡开始，考虑将来可能对顺子系番型的处理，需要把3n+2系中能获取的数据拿过来
	Hu_struct    *[4][2]byte //硬胡的
	GuiHu_struct *[4][2]byte //屁胡的
	//可能是多种组合的话。。。。这个是顺子系的组合，不敢估计了
	// handWeaveInfo [][]*public.TagWeaveItem
	//20190305 先丢这个地，将来都要重新整理了

	Expand_Mask *ExpandMask //目前是大局设定全面的，不排除将来每个番里面也限定，哭了
}

/*
参数说明：
1、对赖子牌的设定（别人打的赖子牌还原）
2、258mask（主要考虑将来对于：硬258将，软258将（纯赖子替换，1+赖子替换）可能会有不同的处理），其他情况无所谓
3、针对扩展可能（目前只是标志要不要判断7对（豪华、2豪华、3豪华）），添加了对见字胡的mask分两段，下8低4位用做见字胡根据实情再次扩展
4、标志是不是要检查软胡的番型 目前废弃（20190221）
*/
func NewCheckHandInfo(guimask uint, mask258 byte, mask_special uint64, checkMgFan bool) (newCheckHand *CheckHuCTX, err error) {
	newCheckHand = &CheckHuCTX{
		GodInfo:      nil,
		Mask258:      mask258,
		Mask_special: mask_special >> 8, //这个是非见字胡用的 上8位
		CheckMFan:    checkMgFan,
		GodHuMask:    mask_special & MASK_HU, //20190213 添加 查胡表用这个
		FanGodHuMask: mask_special & MASK_HU, //20190214 添加 番表7对见字胡查的是这个
		// ChiHuKind:   public.CHK_NULL,
		CheckGodOrg: nil,
	}
	if guimask != 0 {
		newCheckHand.GodInfo = NewGodCardOrg(guimask)
	}
	return
}

func (this *CheckHuCTX) ClearItemEveryCheck() {
	//20190129 发现隐藏bug
	this.ChiHuKind = static.CHK_NULL
	this.CheckGodOrg = nil
	this.ReWarnCTXinfo = nil
	this.ReweaveInfo = nil
	this.CheckCardItem = nil
	this.CheckCardEX = nil
	this.Hu_struct = nil
	this.GuiHu_struct = nil
}

//20190305 目前未用，如果皮子牌会影响到hu ，那就要用了
func (this *CheckHuCTX) SetPiZiCard(card byte) error {
	if !mahlib2.IsMahjongCard(card) {
		return errors.New(fmt.Sprintf("无效皮子牌(%x)", card))
	}
	this.piZiCard = card
	return nil
}

/*
//20190305 特殊设置 嘉鱼红中赖子杠
*/
func (this *CheckHuCTX) SetRestrictGodItem(mask *RestrictGodItem) {
	if this.Expand_Mask != nil {
		this.Expand_Mask.Restrict_GodItem = mask
	}
	this.Expand_Mask = NewExpandMask(mask)

}

//20190109  添加鬼牌数据
func (this *CheckHuCTX) SetGuiCard(guicards byte) (err error) {
	if this.GodInfo == nil {
		return errors.New("未设置赖子的mask，退出")
	}
	//这个地方，可能是清理
	err = this.GodInfo.Setguicard(guicards)
	if err != nil {
		return err
	}
	return nil
}

func (this *CheckHuCTX) GetGuiInfo() []byte {
	if this.GodInfo != nil {
		return this.GodInfo.GetGuiInfo()
	}
	return nil
}

//20190213
func (this *CheckHuCTX) SetFanGodHuMask(godHuMask uint64) {
	this.FanGodHuMask = this.GodHuMask | godHuMask
	// fmt.Println(fmt.Sprintf("**%b", this.ChiHuKind))
}

//20190109
func (this *CheckHuCTX) SetchiHuKind(chiHuKind uint64) {
	this.ChiHuKind |= chiHuKind
	// fmt.Println(fmt.Sprintf("**%b", this.ChiHuKind))
}
func (this *CheckHuCTX) ClearHuKind() {
	this.ChiHuKind &= 0
	// fmt.Println(fmt.Sprintf("**%b", this.ChiHuKind))
}
func (this *CheckHuCTX) SetSpecialMask(mask_special uint64) {
	this.Mask_special |= mask_special
	// fmt.Println(fmt.Sprintf("*----*%b", this.Mask_special))
}

//20190109 考虑到有赖子的情况，所有的检查都检查CheckCardItem
func (this *CheckHuCTX) CreatNewCheckCards() {
	this.CheckCardItem = make([]byte, len(this.CheckCtx.CbCardIndex[:]))
	copy(this.CheckCardItem, this.CheckCtx.CbCardIndex[:])

}

//20190117 判断数据进入，目前主要是给checkcard修改属性 预备初步的处理都丢这里
//调整一下
func (this *CheckHuCTX) CreateCheckObject(checkCtx *BaseCTX) error {
	if checkCtx == nil {
		return errors.New("没有手牌数据")
	}
	this.CheckCtx = checkCtx
	this.ClearItemEveryCheck()
	//20190116 现在放这里应该没问题了 ，因为初始化的时候检查牌已经进来了
	//20190220 这个设计也坑了下，太死了，如果没有Checkcard，就先空着，因为见字胡的话，会检查至少9张牌
	if checkCtx.Checkcard != nil {
		return this.SetCheckCardInfo(checkCtx.Checkcard)
	} else {
		return this.SetCheckCardInfo_ex(checkCtx.Checkcard)
	}
	return nil
}

//20190220 根据见字胡的需求，可重设检查牌信息来检查见字胡
func (this *CheckHuCTX) SetCheckCardInfo_ex(cardInfo *CardBaseInfo) error {
	//if cardInfo == nil {
	//	return errors.New("没有检查牌数据")
	//}
	var newCardEX *CardCTX = nil
	var err error
	if cardInfo != nil {
		cardClass := this.getCardClass(cardInfo.ID)
		newCardEX, err = NewCardCTX(cardInfo, cardClass)
		if err != nil {
			return err
		}
	}

	this.CheckCardEX = newCardEX
	//创建检查手牌copy
	this.CreatNewCheckCards()
	if len(this.GodInfo.guicards) != 0 {
		//这里创建	this.CheckGodOrg
		this.setCheckGodOrg(this.CheckCardItem, this.CheckCardEX)
	}
	//根据设定，修改checkcard的属性 20190117 这个目前意义不大，先不加
	//20190119 这里添加checkcode
	if this.CheckCardEX != nil {
		cardIndex, _ := mahlib2.CardToIndex(this.CheckCardEX.BaseInfo.ID)
		if cardIndex != static.INVALID_BYTE {
			this.CheckCardItem[cardIndex] += 1
		}
	}
	// var newCardEX *CardCTX = nil
	this.ReweaveInfo, _ = this.reCreateWeaveItem(this.CheckCtx.WeaveItem)
	return nil
}

//20190220 根据见字胡的需求，可重设检查牌信息来检查见字胡
func (this *CheckHuCTX) SetCheckCardInfo(cardInfo *CardBaseInfo) error {
	//if cardInfo == nil {
	//	return errors.New("没有检查牌数据")
	//}
	cardClass := this.getCardClass(cardInfo.ID)
	newCardEX, err := NewCardCTX(cardInfo, cardClass)
	if err != nil {
		return err
	}
	this.CheckCardEX = newCardEX
	//创建检查手牌copy
	this.CreatNewCheckCards()
	if len(this.GodInfo.guicards) != 0 {
		//这里创建	this.CheckGodOrg
		this.setCheckGodOrg(this.CheckCardItem, this.CheckCardEX)
	}
	//根据设定，修改checkcard的属性 20190117 这个目前意义不大，先不加
	//20190119 这里添加checkcode
	cardIndex, _ := mahlib2.CardToIndex(this.CheckCardEX.BaseInfo.ID)
	this.CheckCardItem[cardIndex] += 1
	// var newCardEX *CardCTX = nil
	this.ReweaveInfo, _ = this.reCreateWeaveItem(this.CheckCtx.WeaveItem)
	return nil
}

func (this *CheckHuCTX) DecreaseGui() {
	if this.CheckGodOrg != nil && this.CheckGodOrg.GodNum != 0 {
		//fmt.Println(fmt.Sprintf("去牌（%v）",this.CheckCardItem))
		for _, v := range this.CheckGodOrg.GodCardInfo {
			guiIndex, _ := mahlib2.CardToIndex(v.ID)
			//fmt.Println(fmt.Sprintf("赖子牌记录（%d）（%d）(%d)(%d)",guiIndex,v.ID,this.CheckCardItem[guiIndex], v.Num))
			if this.CheckCardItem[guiIndex] >= v.Num {
				this.CheckCardItem[guiIndex] -= v.Num
			} else {
				errstr := fmt.Sprintf("去赖子牌（%s)异常，手牌中数量（%x）记录数为（%d）", mahlib2.G_CardAnother[guiIndex], this.CheckCardItem[guiIndex], v.Num)
				fmt.Println(errstr)
				// panic(errstr)
			}
		}
	}
}
func (this *CheckHuCTX) RecoverGui() {
	if this.CheckGodOrg != nil && this.CheckGodOrg.GodNum != 0 {
		//fmt.Println(fmt.Sprintf("还原牌（%v）",this.CheckCardItem))
		for _, v := range this.CheckGodOrg.GodCardInfo {
			guiIndex, _ := mahlib2.CardToIndex(v.ID)
			if this.CheckCardItem[guiIndex]+v.Num > 4 {
				errstr := fmt.Sprintf("还原赖子牌（%s）异常，手牌中数量（%x）记录数为（%d）", mahlib2.G_CardAnother[guiIndex], this.CheckCardItem[guiIndex], v.Num)
				fmt.Println(errstr)
				// panic(errstr)
			}
			// fmt.Println(fmt.Sprintf("还原赖子牌（%s），还原前数量（%x）", common.G_CardAnother[guiIndex], this.CheckCardItem[guiIndex]))
			this.CheckCardItem[guiIndex] += v.Num
			// fmt.Println(fmt.Sprintf("还原赖子牌（%s），还原数量（%x）记录数为（%d）", common.G_CardAnother[guiIndex], this.CheckCardItem[guiIndex], v.Num))
		}
	}
}

//20190119 将牌设定有问题，但是当赖子可以不管这个东西
func (this *CheckHuCTX) getCardClass(checkcard byte) (cardclass int) {
	// if this.Mask258 > 0 {
	// 	cardclass |= EYECARD
	// }
	if this.piZiCard == checkcard {
		cardclass |= PICARD
		return
	}
	if this.GodInfo != nil {
		for _, v := range this.GodInfo.guicards {
			if v == checkcard {
				cardclass |= GODCARD
				return
			}
		}
	}
	return
}

// func (this *CheckHuCTX) Set258eye(eye258 bool) {
// 	this.eye258 = eye258
// }

//20190109 先移植过来，最终判断checkcard是不是能当鬼牌是要经过CardCTX的设置来判断的
//这个不能简单的处理，在清一色里面，全去掉就行，而在碰碰胡里面，还是要保留数量，用来凑刻 所以这里要拿出所有鬼牌各自的数量
/*
考虑到有这样的情况，3个赖子，其中2个做了将的话。。。那仍然不能算是清一色，所以要确定的是把所有需求的赖子扔掉后，还有没有剩下的
*/
//这个是获取初始的赖子数，判胡成功后，还要减去，用来判断是不是真正符合清一色
func (this *CheckHuCTX) setCheckGodOrg(cbCardIndex []byte, checkcard *CardCTX) {
	if this.GodInfo == nil {
		return
	}
	var sguiNum byte = 0
	for _, CardID := range this.GodInfo.guicards {
		sguiNum = 0
		v, _ := mahlib2.CardToIndex(CardID)
		if v < MAX_INDEX {
			sguiNum += cbCardIndex[v]
		}
		//只有自己起的才能算数 放放，只有自己起的赖子才算，别人打的还原
		if checkcard != nil {
			// fmt.Println(fmt.Sprintf("检查值比较ID（%d）V(%d)相等（%t）IsDraw（%t）", checkcard.ID, CardID, (checkcard.ID == CardID), checkcard.IsDraw))
			if checkcard.ExInfo.ID == CardID && this.GodInfo.CanBeGod(checkcard.BaseInfo.Origin) {
				checkcard.ExInfo.IsGod = true
				sguiNum += 1
			}
		}
		//放到god数据里面去
		if sguiNum != 0 {
			/*
				//	//20190306 嘉鱼红中，限制赖子个数胡，有个问题，目前是支持2种赖子，但是如果2个赖子都有，但是只能去一个，那就恶心了，因为有可能一个赖子会出现大胡的情况，这种就要出现两种手牌可能性
					//	//目前项目比较赶,这个修改，暂不做；目前都是1个赖子,那么就直接修改
			*/
			//----------------------------------------------
			item, err := NewGodCardInfo(CardID, sguiNum)
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				if this.CheckGodOrg == nil {
					//创建checkgodorg结构
					this.CheckGodOrg = NewCheckGodOrg()
				}
				this.CheckGodOrg.GodCardInfo = append(this.CheckGodOrg.GodCardInfo, item)
				//god总数
				// 20190306 大胡不限制赖子
				//if this.Expand_Mask!=nil&&this.Expand_Mask.Restrict_GodItem!=nil&&this.Expand_Mask.Restrict_GodItem.Restrict_NeedGodNum!=0{
				//	sguiNum=this.Expand_Mask.Restrict_GodItem.Restrict_NeedGodNum
				//}
				//this.CheckGodOrg.NeedGodNum	=sguiNum
				this.CheckGodOrg.GodNum += sguiNum
			}
		}
	}
	return
}

//20181226 重新获取倒牌信息，在判番这块，吃牌的信息从以前的左中右统一改成左吃信息
//20190117 这个只是构架的一部分，如果将来顺子真的复杂起来也要重写
func (this *CheckHuCTX) reCreateWeaveItem(weaveItem []static.TagWeaveItem) ([]*static.TagWeaveItem, error) {
	cbWeaveCount := byte(len(weaveItem))
	if cbWeaveCount == 0 {
		return nil, nil
	}
	//20190221 在这里生成
	this.ReWarnCTXinfo = this.GetWarninfo()
	// if cbWeaveCount > 4 {
	// 	return nil, errors.New(fmt.Sprintf("倒牌组合越界%d", cbWeaveCount))
	// }
	result := make([]*static.TagWeaveItem, 0)
	for index, v := range weaveItem {
		if v.WeaveKind == static.WIK_NULL {
			continue
		}
		switch v.WeaveKind {
		case static.WIK_LEFT, static.WIK_PENG, static.WIK_GANG: //左吃、碰、杠（明杠、暗杠）
			result = append(result, &weaveItem[index])
		case static.WIK_CENTER: //中吃类型
			neweave := NewTagWeaveItem(static.WIK_LEFT, v.CenterCard-1, v.PublicCard, v.ProvideUser)
			result = append(result, neweave)
		case static.WIK_RIGHT: //右吃类型
			neweave := NewTagWeaveItem(static.WIK_LEFT, v.CenterCard-2, v.PublicCard, v.ProvideUser)
			result = append(result, neweave)
		default:
			// 	fmt.Println("reCreateWeaveItem 未知类型", v.WeaveKind)
		}
	}
	return result, nil
}

/*
可以知道的东西：
统一花色
是不是都是外来牌（全求人）
吃牌的左位牌，reCreateWeave了
还有有。。。。。大于5、全中。。。。这个先放放
大于5小于5.。。
*/
//目前需要的，是不是统一花色（要是的话需要知道是什么，清一色，字一色需要）
//头疼了 先做一版 后续再修改
func (this *CheckHuCTX) GetWarninfo() *WarnCTXinfo {
	if len(this.CheckCtx.WeaveItem) == 0 {
		return nil
	}
	newinfo := &WarnCTXinfo{}
	//一次拿完所有的东西
	newinfo.is258 = 1
	haveItem := 0
	for _, v := range this.CheckCtx.WeaveItem {
		//颜色
		if v.CenterCard == 0 {
			continue
		}
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
		case static.WIK_LEFT:
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

//20190119 获取原始手牌数
func (this *CheckHuCTX) GetBaseHandCardNum() (result byte) {
	result = 0
	if this.CheckCtx != nil {
		for _, v := range this.CheckCtx.CbCardIndex {
			result += v
		}
	}
	return
}

//20190119 获取手牌数
func (this *CheckHuCTX) GetHandCardNum() (result byte) {
	numSp := this.getHandCardNumSP()
	result = 0
	for _, v := range numSp {
		result += v
	}
	return
}

//20181227 获取每个花色的牌数据 这个测试一下
func (this *CheckHuCTX) getHandCardNumSP() (result []byte) {
	var context byte = 0
	//万数据
	for i := 0; i < 9; i++ {
		context += this.CheckCardItem[i]
	}
	result = append(result, context)
	//条判断
	context = 0
	for i := 9; i < 18; i++ {
		context += this.CheckCardItem[i]
	}
	result = append(result, context)
	//筒判断
	context = 0
	for i := 18; i < 27; i++ {
		context += this.CheckCardItem[i]
	}
	result = append(result, context)
	//风牌判断
	context = 0
	for i := 27; i < 31; i++ {
		context += this.CheckCardItem[i]
	}
	result = append(result, context)
	//箭牌判断
	context = 0
	for i := 31; i < 34; i++ {
		context += this.CheckCardItem[i]
	}
	result = append(result, context)
	return result
}

//获取手牌中牌的颜色 一色牌判断 ，除了风。箭的判断
func (this *CheckHuCTX) GetHandCardColorandNum() (result byte, is258 byte, num byte) {
	checkInfo := this.getHandCardNumSP()
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
	if this.Mask258 != 0 {
		for i := 1; i <= 25; i += 3 {
			if this.CheckCardItem[i] != 0 {
				context += this.CheckCardItem[i]
			}
		}
	}
	if num-context == 0 {
		return result, 1, num
	}
	return result, 0, num
}

//给一色牌判断用 追加285（将一色判断）
func (this *CheckHuCTX) GetAllCardItembase() (result byte, is258 bool) {
	var hand258 byte = 0
	var handCardNum byte = 0
	result, hand258, handCardNum = this.GetHandCardColorandNum()
	//20190221 去掉赖子的时候 手牌就不一定是14了
	if handCardNum != MAX_COUNT && this.ReWarnCTXinfo != nil {
		result = result | this.ReWarnCTXinfo.colorInfo
		hand258 &= this.ReWarnCTXinfo.is258
		//}else{
		//	//fmt.Println("门清状态")
	}
	return result, hand258 == 1
}

// //根据两个数据来生成当前的场景mask
// func CreatCtxScene(ctxScene byte, Origin byte) byte {
// 	return 0
// }
