package fanlib

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	scoringlib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_scoring"
	"github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

/*说明：
七对 国番中只有7对的概念，所有这个地方要重新考虑返回值的问题
*/

const (
	_SCORING_19_ID     = 19
	_SCORING_19_NAME   = "七对"
	_SCORING_19_FANSHU = 24
)

var _SCORING_19_DISCARDID_ = []int{52}

//自己注册
func init() {
	// fmt.Println("19")
	G_ScoringManager.RegisterBaseHander(&scoring_19{
		id:           _SCORING_19_ID,
		name:         _SCORING_19_NAME,
		fanShu:       _SCORING_19_FANSHU,
		setDiscardID: _SCORING_19_DISCARDID_,
		huKind:       scoringlib2.SCORING_SPECIAL,
		humask:       scoringlib2.CANBE_ZIMO | scoringlib2.CANBE_CHIHU,
	})
}

type scoring_19 struct {
	id           int
	name         string
	fanShu       int
	setDiscardID []int
	huKind       int //
	humask       byte
}

func (this *scoring_19) GetID() int {
	return this.id
}

func (this *scoring_19) Name() string {
	return this.name
}

func (this *scoring_19) GetFanShu() int {
	return this.fanShu
}
func (this *scoring_19) GethuKind() int {
	return this.huKind
}

/*
7对，手牌区检测只是符合是不是7对
*/
func (this *scoring_19) CheckSatisfySelf(handInfoCtx *card_mgr.CheckHuCTX) bool {
	var godNum byte = 0
	if handInfoCtx.CheckGodOrg != nil {
		godNum = handInfoCtx.CheckGodOrg.GodNum
	}
	handInfoCtx.DecreaseGui()
	//result,_:=this.Check_base(handInfoCtx.CheckCardItem,godNum)
	var need byte = 0
	for i := 0; i < 34; i++ {
		if handInfoCtx.CheckCardItem[i]%2 != 0 {
			need = need + 1
		}
	}
	handInfoCtx.RecoverGui()
	if need > godNum {
		return false
	}
	return true
}

//这个用做特殊处理，如果有赖子，要根据实际情况重做手牌数据 ，这是特殊番型，在所有常规（3n+2）处理完后，再来这里
//20190212 特殊胡牌里面只有七对才会出现见字胡 鬼牌数-需要的鬼牌数 之后还有单张鬼牌，那就是见字胡
func (this *scoring_19) SpircalProcess(handInfoCtx *card_mgr.CheckHuCTX) (uint64, byte, []int) {
	// var resultMask uint16 = public.CHK_NULL
	var result uint64 = static.CHK_NULL
	//这个估计要修改了 目前4位，3豪华、2豪华、豪华、7小对 20190218 添加第5位是门清
	//卡五星有的规则7小对是必须门清的，查了下国标也是门清的，但是公司的胸弟们告诉我可以不门清。。。。
	checkmask := handInfoCtx.Mask_special
	//check7dui:=handInfoCtx.Mask_special&0xf
	//20190220 默认都是门清的
	//if handInfoCtx.GetHandCardNum() != 14&&checkmask&lib_scoring.MASK_7DUI_NOCONCEALED==0{
	handcardNm := handInfoCtx.GetBaseHandCardNum()
	if handcardNm != 13 {
		//fmt.Println(fmt.Sprintf("七对必须门清(%d)",handcardNm))
		return result, 0, nil
	}
	//-----------------------------------------
	//20190218 倒牌中的杠牌要算---倒牌中的豪华是硬性的，如果遇到那种倒牌的豪华不能拆，但是又不支持的情况呢-
	//warnCTXinfo := handInfoCtx.GetWarninfo()
	//warnHao:=0
	//if warnCTXinfo!=nil{
	//	//明杠，吃牌就不用检查了
	//	//if len(warnCTXinfo.ChipaiInfo) +len(warnCTXinfo.Triplet)>0{
	//	if len(warnCTXinfo.ChipaiInfo)>0{
	//		return result,nil
	//	}
	////如果有规则明杠和蓄杠不算的话。。。。
	//if checkmask&lib_scoring.MASK_7DUI_TRIPLET==0&&len(warnCTXinfo.Triplet) != 0{
	//	fmt.Println("7对不算明杠")
	//	return public.CHK_NULL, nil
	//}
	//	warnHao += len(warnCTXinfo.Triplet) + len(warnCTXinfo.HidTriplet)
	//	//20190219 特殊情况，warnHao最大3，但是有碰4对后，再追加蓄杠，造成4杠组合
	//	if warnHao>3{
	//		return public.CHK_NULL, nil
	//	}
	//}
	//----------20190219 这个地方有个问题，倒牌中的杠牌是按照豪华来算的，假如有规则说能拆就不用检查mask，如果不能拆，那么就要检查mask--------
	if this.CheckSatisfySelf(handInfoCtx) {
		handInfoCtx.DecreaseGui()
		groupinfo := card_mgr.ClassifyCards(handInfoCtx.CheckCardItem)
		needGui := len(groupinfo.OneCards) + len(groupinfo.ThreeCards)
		//20190219 手牌有赖子的情况才可能见字胡
		//-------如果见字胡mask存在，要检查一下----七对比较特殊，因为如果支持软豪华的情况，下面也会出现MASK_SPECIAL_MAGIC
		//至少要有god才能见字胡
		//fmt.Println(fmt.Sprintf("7对见字胡mask（%d）",handInfoCtx.FanGodHuMask^card_mgr.ORIGIN_HAND))
		if handInfoCtx.FanGodHuMask^card_mgr.ORIGIN_HAND != 0 && needGui != 0 {
			//设定FanGodHuMask的ORIGIN_TABILE位为1的时候不检查见字胡
			if this.checkGodHu(handInfoCtx) {
				//20190222 这里见字胡直接出去了，外面其实也可以做，但是这样会提高一点效率
				if handInfoCtx.FanGodHuMask != 0 && uint64(handInfoCtx.CheckCardEX.BaseInfo.Origin)&handInfoCtx.FanGodHuMask == 0 {
					//fmt.Println(fmt.Sprintf("七对见字胡不能胡(%d)mask(%d)",handInfoCtx.CheckCardEX.BaseInfo.Origin,handInfoCtx.FanGodHuMask))
					//7对见字胡是不是影响所属胡权取消，这个要根据实际情况而定，目前武汉晃晃是不能吃胡的
					handInfoCtx.ClearHuKind()
					//20190221 见字胡的位还是要置的，这样
					handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_GODHU)
					handInfoCtx.RecoverGui()
					return static.CHK_NULL, 0, nil
				}
				//设置见字胡
				handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_GODHU)
			}
		}
		baseoffset := len(groupinfo.FourCards) + len(groupinfo.ThreeCards)
		offset := 0 //手牌中最大的豪华7对数
		//手牌中赖子还原使用的可能>4 要记录豪华
		godNum := 0
		if handInfoCtx.CheckGodOrg != nil {
			godNum = int(handInfoCtx.CheckGodOrg.GodNum)
			//若有限定 20190306
			if handInfoCtx.CheckGodOrg.NeedGodNum != 0 {
				godNum = int(handInfoCtx.CheckGodOrg.NeedGodNum)
			}
		}
		//这里注意大小的问题
		listGodNum := godNum - needGui
		//20190402 这里有两种情况，如果>3 就会出现4张+散牌的情况
		haoHuaNum := listGodNum / 4
		//if haoHuaNum>0{
		//如果在硬胡的情况下，剩下2个赖子，会在balanceOffset里面配成豪华7对
		listGodNum = listGodNum % 2
		//}
		//4赖子先算豪华
		offset = baseoffset + haoHuaNum
		//手牌的应该可以直接组合
		offset = this.balanceOffset(offset, listGodNum)
		//20190219 根据mask来重新处理
		result = this.resetResult(checkmask, offset)
		//分类result的位置，软胡的位置是高4 如果手牌有单张必然是软胡
		if needGui != 0 {
			if result != static.CHK_NULL {
				handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_MAGIC)
				result = result << scoringlib2.MASK_OFFSET
			}
		} else {
			//硬胡的情况就是赖子还原或者没有赖子
			if result != static.CHK_NULL {
				handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_NOMAGIC)
			}
			if godNum != 0 {
				//有赖子就要多算一次，因为会有硬豪华，软2的情况，我们要列举出来
				offset = this.balanceOffset(baseoffset, int(godNum)-needGui)
				tempResult := this.resetResult(checkmask, offset)
				if tempResult != static.CHK_NULL {
					handInfoCtx.SetchiHuKind(scoringlib2.MASK_SPECIAL_MAGIC)
					result |= tempResult << scoringlib2.MASK_OFFSET
					//软豪华必要赖子
					needGui = godNum - needGui
				}
			}
		}
		//------------------------------------------------------
		//if needGui != 0 {
		//	//------软7---8赖子的情况下还是有可能在去掉赖子的情况，再出现可硬豪的情况，但是这个情况只可能加一豪华--
		//	handInfoCtx.SetchiHuKind(lib_scoring.MASK_SPECIAL_MAGIC)
		//	//20190124 要过滤掉其中的赖子。。。因为赖子要补，最大
		//	offset := this.balanceOffset(baseoffset+len(groupinfo.ThreeCards), int(handInfoCtx.CheckGodOrg.GodNum)-needGui)
		//	//20190219 支持warn里面的处理
		//	result |=this.resetResult_Deep(checkmask,warnHao,offset) << lib_scoring.MASK_OFFSET
		//	//result |= this.resetResult(checkmask, offset) << lib_scoring.MASK_OFFSET
		//} else {
		//	handInfoCtx.SetchiHuKind(lib_scoring.MASK_SPECIAL_NOMAGIC)
		//	//公用模块返回所有可能
		//	if handInfoCtx.CheckGodOrg.GodNum != 0 {
		//		handInfoCtx.SetchiHuKind(lib_scoring.MASK_SPECIAL_MAGIC)
		//		//这个地方要修改一下，如果godNum>4的话，在硬胡的情况下，要当豪华处理，那么
		//		haoHuaNum:=int(handInfoCtx.CheckGodOrg.GodNum)/4
		//		listGodNum:=0
		//		//4赖子先算豪华
		//		offset=baseoffset+haoHuaNum
		//		listGodNum=int(handInfoCtx.CheckGodOrg.GodNum)%4
		//		if haoHuaNum!=0 {
		//			//硬的
		//			offset = this.balanceOffset(offset, listGodNum)
		//			result |=this.resetResult_Deep(checkmask,warnHao,offset)
		//			//result |= this.resetResult(checkmask, offset)
		//		}else if int(handInfoCtx.CheckGodOrg.GodNum)/2==1{
		//			//硬的
		//			listGodNum=int(handInfoCtx.CheckGodOrg.GodNum)%2
		//			offset = this.balanceOffset(baseoffset, listGodNum)
		//			result |=this.resetResult_Deep(checkmask,warnHao,offset)
		//			//result |= this.resetResult(checkmask, offset)
		//		}
		//			//到这里最多也就是3张 但是走到这里只可能是2张
		//			listGodNum=int(handInfoCtx.CheckGodOrg.GodNum)
		//			//软的
		//			offset = this.balanceOffset(baseoffset, listGodNum)
		//		result |=this.resetResult_Deep(checkmask,warnHao,offset)<<lib_scoring.MASK_OFFSET
		//			//result |= this.resetResult(checkmask, offset)<<lib_scoring.MASK_OFFSET
		//		//
		//	}else{
		//		result |=this.resetResult_Deep(checkmask,warnHao,baseoffset)
		//		//result |= this.resetResult(checkmask, baseoffset)
		//	}
		//}
		//------------------------------------------------
		handInfoCtx.RecoverGui()
		//fmt.Println(fmt.Sprintf("七对检查result:(%b),chiHu(%b)", result<<8,handInfoCtx.ChiHuKind))
		return result<<8 | handInfoCtx.ChiHuKind, byte(needGui), this.setDiscardID
	}
	return result, 0, this.setDiscardID
}

//20190212 追加一个专门处理7对mask，拿到豪华中最大的情况
/*
考虑用这个来生成，在需求赖子完后，可能会导致以下情况：
1、一对赖子，在基础上，还有置位软豪华
2、4个赖子，在基础上，还要置 追加软2豪华（可能已经有1豪华了，那就是3豪华）
3、碰碰胡，4个赖子分摊出去，目前这个和7对是弱联系，这个要测一下
目前也考虑2种赖子牌的情况
20190218 武汉麻将里面倒牌的暗杠也算。。。
*/
func (this *scoring_19) balanceOffset(offset int, remainNum int) int {
	//这个地方没有做安全验证 如果是2的话，结果配成一个豪华了
	if remainNum < 2 {
		return offset
	}
	//到这里需要重新算 最后剩下的godNum必然是2的倍数，因为已经算过都是14张了，而最大不会超过3豪华
	//2,4,6,8 这个函数里面不做降级处理，只返回最大豪华数
	resultOffset := offset + remainNum/2
	//这个是8赖子的情况，如果除了赖子对，手牌没有豪华的情况，就成了4豪华了
	if resultOffset > 3 {
		resultOffset = 3
	}
	return resultOffset
}

//20190219
/*
//假设这样，要把倒牌的杠牌能不能拆带进去，如果不能裁，要先判断offset
//如果能拆，就全带进去，所以要多带一个参数，warn倒牌
参数1：手牌中非赖子豪华数（包含3张的）
参数2：倒牌中的豪华数（可能包含明杠） 在里面判断能不能拆
参数3：剩下的god
*/
//func (this *scoring_19) resetResult_Deep(mask_special uint16, warnHao int,offset int) uint16 {
//	//check的时候要先
//	var result uint16=static.CHK_NULL
//	fmt.Println(fmt.Sprintf("检查的mask（%b）",mask_special))
//	check7_dui:=mask_special&0xf
//	if warnHao!=0&&mask_special&lib_scoring.MASK_7DUI_WARNNOHAO ==0{
//			//默认不能拆
//			check7dui:= check7_dui >> uint(warnHao)
//			if check7dui <= 0 {
//				//不支持
//				return result
//			}
//			//还能支持的话,拿手牌中最大的豪华等级
//			result=this.resetResult_Safe(check7dui,offset,true)
//			if result==static.CHK_NULL{
//				return result
//			}
//			return result<<uint(warnHao)
//	}		//return this.resetResult(mask_special,offset+warnHao)
//		return this.resetResult_Safe(check7_dui,offset+warnHao,false)
//}
//根据mask来设置，如果规则中不支持的豪华种类，用降位
/*
考虑比较恶心的情况，直接只要是判断支持3豪华就先置全位，在与mask与
再考虑是不是保留最高位
预备废弃
*/
func (this *scoring_19) resetResult(mask_special uint64, offset int) uint64 {
	//check的时候要先
	var result uint64 = 1 << uint64(offset)
	if mask_special == 0 || result&mask_special > 0 {
		return result
	}
	temp := result
	for ; temp != 0; temp = temp >> 1 {
		if temp&mask_special != 0 {
			break
		}
	}
	return temp
}

//20190218 因为倒牌中的也要算，那么先加到数据里面，如果到最后符合了，但是又小于倒牌中豪华数，那么就失败
//这个要严格匹配 因为 check7dui是移位后的mask
func (this *scoring_19) resetResult_Safe(check7dui uint16, offset int, safe bool) uint16 {
	//check的时候要先
	var result uint16 = 1 << uint16(offset)
	if result&check7dui > 0 {
		return result
	}
	if !safe && check7dui == 0 {
		return result
	}
	temp := result
	find := false
	for ; temp != 0; temp = temp >> 1 {
		if temp&check7dui != 0 {
			find = true
			break
		}
	}
	if safe && !find {
		return static.CHK_NULL
	}
	return temp
}

//7对见字胡的判断方式
/*
如果已经确定是7对，而且有赖子的情况，才可能是见字胡
目前的想法是去掉检查的牌，再算一次，needgui和godnum的差，如果差值成单数才是见字胡
或者用检查见字胡的那个切片带进去，不管哪样，我都要用原始的那个不带判牌的序列
*/
func (this *scoring_19) checkGodHu(handInfoCtx *card_mgr.CheckHuCTX) bool {
	//最终还是直接处理手牌（最多13张）的情况
	//copy一份直接处理
	CheckCardItem := make([]byte, len(handInfoCtx.CheckCtx.CbCardIndex[:]))
	copy(CheckCardItem, handInfoCtx.CheckCtx.CbCardIndex[:])
	if handInfoCtx.CheckGodOrg == nil {
		return false
	}
	guiItem := handInfoCtx.CheckGodOrg.GodCardInfo
	var godNum byte = 0
	//因为支持双赖子，所以要先去掉赖子，同时要记录赖子牌数
	for _, v := range guiItem {
		index, _ := mahlib2.CardToIndex(v.ID)
		godNum += v.Num
		CheckCardItem[index] -= v.Num
	}
	if godNum == 0 {
		return false
	}
	//清理掉CheckCardItem的鬼牌，并计数
	var need byte = 0
	for i := 0; i < 34; i++ {
		if CheckCardItem[i]%2 != 0 {
			need += 1
		}
	}
	//到最后
	if godNum > need && (godNum-need)%2 != 0 {
		return true
	}
	return false
}
func (this *scoring_19) checkGodHu_base(cbCardIndex []byte, godNum byte) bool {
	//最终还是直接处理手牌（最多13张）的情况
	//copy一份直接处理
	//清理掉CheckCardItem的鬼牌，并计数
	var need byte = 0
	for i := 0; i < 34; i++ {
		if cbCardIndex[i]%2 != 0 {
			need += 1
		}
	}
	//到最后
	if godNum > need && (godNum-need)%2 != 0 {
		return true
	}
	return false
}
func (this *scoring_19) Check_base(cbCardIndex []byte, godNum byte) (result bool, loseGui byte) {
	var need byte = 0
	for i := 0; i < 34; i++ {
		if cbCardIndex[i]%2 != 0 {
			need = need + 1
		}
	}
	if need > godNum {
		return false, 0
	}
	return true, godNum - need
}

//普通接口
func (this *scoring_19) Check_Normal(cbCardIndex []byte, weaveItem []static.TagWeaveItem, checkCard byte, isNormalCard bool, guiCards []byte) (result uint64, needgodNum byte, err error) {
	result = static.CHK_NULL
	needgodNum = 0
	//num,err:=common.CheckHandCardsSafe(cbCardIndex)

	num, err := mahlib2.CheckHandCardsSafe_ex(cbCardIndex, checkCard)
	if err != nil {
		//目前7对只能门清
		return result, needgodNum, err
	}
	if num != 14 {
		//fmt.Println(fmt.Sprintf("七对检查手牌数不是14（%d）",num))
		return result, needgodNum, nil
	}
	for _, v := range weaveItem {
		if v.WeaveKind != 0 {
			//fmt.Println(fmt.Sprintf("七对不能有倒牌"))
			return result, needgodNum, nil
		}
	}
	if checkCard > 0x37 && checkCard != static.INVALID_BYTE {
		return result, needgodNum, errors.New(fmt.Sprintf("AnalyseHu_Normal 检查牌越界，最大值是0x37，当前值为（%x）", checkCard))
	}
	//生成，手牌数据（去赖子的），赖子计数 这里需要判断软硬胡的情况
	var checkCards []byte
	checkCards, needgodNum = card_mgr.CreateNewCheckCards(cbCardIndex, checkCard, isNormalCard, guiCards, false)
	//先判断是不是见字胡
	if this.checkGodHu_base(checkCards, needgodNum) {
		fmt.Println("见字胡")
		result |= scoringlib2.MASK_SPECIAL_GODHU
	}
	//检查普通的7对
	isok, loseGui := this.Check_base(checkCards, needgodNum)
	if !isok {
		return static.CHK_NULL, static.INVALID_BYTE, nil
	}
	//剩下的gui数可能是0也可能是原值或者是小于guinum的
	groupinfo := card_mgr.ClassifyCards(checkCards)
	needGui := len(groupinfo.OneCards) + len(groupinfo.ThreeCards)
	baseoffset := len(groupinfo.FourCards) + len(groupinfo.ThreeCards)
	//手牌中赖子还原使用的可能>4 要记录豪华
	//20190402 这里有两种情况，如果>3 就会出现4张+散牌的情况
	haoHuaNum := int(loseGui / 4)
	//if haoHuaNum>0{
	//如果在硬胡的情况下，剩下2个赖子，会在balanceOffset里面配成豪华7对
	loseGui = loseGui % 2
	//}
	//4赖子先算豪华
	offset := baseoffset + haoHuaNum
	//手牌的应该可以直接组合
	offset = this.balanceOffset(offset, int(loseGui))
	//20190219 根据mask来重新处理
	result = this.resetResult(0, offset)
	//分类result的位置，软胡的位置是高4 如果手牌有单张必然是软胡
	if needGui != 0 {
		if result != static.CHK_NULL {
			result = result << scoringlib2.MASK_OFFSET
		}
	} else {
		//硬胡的情况就是赖子还原或者没有赖子
		if needgodNum != 0 {
			//有赖子就要多算一次，因为会有硬豪华，软2的情况，我们要列举出来
			offset = this.balanceOffset(baseoffset, int(needgodNum)-needGui)
			tempResult := this.resetResult(0, offset)
			if tempResult != static.CHK_NULL {
				result |= tempResult << scoringlib2.MASK_OFFSET
				//软豪华必要赖子
				needGui = int(needgodNum) - needGui
			}
		}
	}
	//fmt.Println(fmt.Sprintf("独立七对检查result:(%b)", result<<8))
	return result << 8, byte(needGui), nil
	//----------------------------------
	//return public.CHK_NULL,public.INVALID_BYTE,nil
}
