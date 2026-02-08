package lib_win

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
)

/*
姐妹铺
预想包含双姐妹铺
去姐妹铺的条件是2,2,2
即是顺子，又要是2个，但是赖子目前就只有4个可用
*/
type DragonObj struct {
	handCards    []byte //去掉龙后的牌数据
	lostGodCards []byte //龙牌所使用到的赖子牌数据
	lastGodNum   byte   //剩下鬼牌 这个基本上用不到
	isDragon     bool   //是不是龙
	huKind       byte
}

func (this *CheckHu) CreateDragon_base_ex(handCards []byte, godNum *byte, index byte, lostCards *[]byte) (result byte) {
	//先看牌够不够
	var lostcard byte = 0
	var needgod byte = 0
	if handCards[index]+*godNum < 1 {
		//这个地方应该销毁一些东西
		*lostCards = nil
		return 0
	}
	if handCards[index] > 0 {
		handCards[index] -= 1
		lostcard = 1
		needgod = 0
		//要记录一下
	} else {
		//不够求余，因为有为0的情况
		needgod = 1
		lostcard = 0
		handCards[index] = 0
		*godNum -= needgod
	}
	//生成记录
	*lostCards = append(*lostCards, lostcard)
	//如果是3个了，就记录一下，并且还原

	return result
}

////构建姐妹铺结果
/*
整合所有结果来创建去掉姐妹铺的牌
//所有可能去拍的总和是2，不可能超过
如果单项就已经满足2的话，另外两色就不用处理了
关键是构造结果

*/
func (this *CheckHu) CreateDragonResult(handCards []byte, objWan *DragonObj, objTiao *DragonObj, objTong *DragonObj) (result *DragonObj) {
	//这个是可能跨色的姐妹铺
	tempcards := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0}
	result = &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	result.handCards = append(result.handCards, objWan.handCards[:]...)
	result.handCards = append(result.handCards, objTiao.handCards[:]...)
	result.handCards = append(result.handCards, objTong.handCards[:]...)
	result.handCards = append(result.handCards, handCards[27:]...)
	if len(objWan.lostGodCards) == 0 {
		result.lostGodCards = append(result.lostGodCards, tempcards[:]...)
	} else {
		result.lostGodCards = append(result.lostGodCards, objWan.lostGodCards[:]...)
		result.lastGodNum = objWan.lastGodNum
		//后面没有龙也就不用加了
		return
	}
	//检查条的
	if len(objTiao.lostGodCards) == 0 {
		result.lostGodCards = append(result.lostGodCards, tempcards[:]...)
	} else {
		result.lostGodCards = append(result.lostGodCards, objTiao.lostGodCards[:]...)
		result.lastGodNum = objTiao.lastGodNum
		return
	}
	//检查筒的
	if len(objTong.lostGodCards) == 0 {
		result.handCards = append(result.handCards, handCards[18:]...)
		//后面风不会有龙
	} else {
		result.lostGodCards = append(result.lostGodCards, objTong.lostGodCards[:]...)
		result.lastGodNum = objTong.lastGodNum
	}
	return result
}

//接替 就是返回一个最多9位的切片
func (this *CheckHu) checkDragon_base(cards []byte, godnum byte, result *DragonObj) bool {
	var i byte
	find := true
	var checkcards []byte
	static.HF_DeepCopy(&checkcards, &cards)
	for i = 0; i < byte(len(checkcards)); i++ {
		//只要空一位，就不用再看了
		if checkcards[i]+godnum < 1 {
			find = false
			break
		}
		if checkcards[i] > 0 {
			checkcards[i] -= 1
			result.lostGodCards = append(result.lostGodCards, 0)
		} else {
			if godnum > 0 {
				godnum -= 1
				//代表用了一个赖子替换
				result.lostGodCards = append(result.lostGodCards, 1)
			} else {
				find = false
				break
			}
		}
	}

	result.isDragon = find
	if !result.isDragon {
		//没找到就清空
		result.lostGodCards = []byte{}
	} else {
		static.HF_DeepCopy(&result.handCards, &checkcards)
	}
	result.lastGodNum = godnum
	return find
}
func (this *CheckHu) CheckDragon_hu(checkCardObj *DragonObj, gui_num byte, eyeMask byte) byte {
	hu, _ := this.CheckHU.Split_Byte(checkCardObj.handCards, checkCardObj.lastGodNum, eyeMask, true)
	if hu {
		//不分类了，根据 好像有问题，如果是赖将，会留2个赖子，并且任意牌
		if (checkCardObj.lastGodNum == gui_num && gui_num != 0) || gui_num == 0 {
			if eyeMask > 1 {
				checkCardObj.huKind |= static.CHK_PING_HU_NOMAGIC << 2
			} else {
				checkCardObj.huKind |= static.CHK_PING_HU_NOMAGIC
			}
		} else {
			if eyeMask > 1 {
				checkCardObj.huKind |= static.CHK_PING_HU_MAGIC << 2
			} else {
				checkCardObj.huKind |= static.CHK_PING_HU_MAGIC
			}
		}
		//////独立检查一下能不能硬胡
		//checkCards := make([]byte, 0)
		//public.HF_DeepCopy(&checkCards, &checkCardObj.handCards)
		//checkCards[godindex]+=checkCardObj.lastGodNum
		//hu, _:= this.CheckHU.Split_Byte(checkCards,0, eyeMask,true)
		//if hu{
		//	if eyeMask>1{
		//		checkCardObj.huKind|=public.CHK_PING_HU_NOMAGIC<<2
		//	}else{
		//		checkCardObj.huKind|=public.CHK_PING_HU_NOMAGIC
		//	}
		//}
		//-----------------------------------------
		return checkCardObj.huKind
	}
	return static.CHK_NULL
}
func (this *CheckHu) CheckDragonResult(checkCardObj *DragonObj, gui_num byte, eyeMask byte) byte {
	hu, _ := this.CheckHU.Split_Byte(checkCardObj.handCards, checkCardObj.lastGodNum, eyeMask, true)
	if hu {
		//不分类了，根据 好像有问题，如果是赖将，会留2个赖子，并且任意牌
		if (checkCardObj.lastGodNum == gui_num && gui_num != 0) || gui_num == 0 {
			if eyeMask > 1 {
				checkCardObj.huKind |= static.CHK_PING_HU_NOMAGIC << 2
			} else {
				checkCardObj.huKind |= static.CHK_PING_HU_NOMAGIC
			}
		} else {
			if eyeMask > 1 {
				checkCardObj.huKind |= static.CHK_PING_HU_MAGIC << 2
			} else {
				checkCardObj.huKind |= static.CHK_PING_HU_MAGIC
			}
		}
		//////独立检查一下能不能硬胡
		//checkCards := make([]byte, 0)
		//public.HF_DeepCopy(&checkCards, &checkCardObj.handCards)
		//checkCards[godindex]+=checkCardObj.lastGodNum
		//hu, _:= this.CheckHU.Split_Byte(checkCards,0, eyeMask,true)
		//if hu{
		//	if eyeMask>1{
		//		checkCardObj.huKind|=public.CHK_PING_HU_NOMAGIC<<2
		//	}else{
		//		checkCardObj.huKind|=public.CHK_PING_HU_NOMAGIC
		//	}
		//}
		//-----------------------------------------
		return checkCardObj.huKind
	}
	checkCardObj.huKind = static.CHK_NULL
	return static.CHK_NULL
}

//普通接口 姐妹铺
/*
想法就是，一色牌至少要有9张以上，然后就是顺序来排下去
*/
func (this *CheckHu) CheckDragon_Normal(handCards []byte, WeaveItem []static.TagWeaveItem, cbCurrentCard byte, isNormalCard bool, godCards []byte, eyeMask byte, usegod bool, seaf bool) (result byte, err error) {

	checkGodCards := []byte{}
	if len(godCards) != 0 {
		static.HF_DeepCopy(&checkGodCards, &godCards)
	}
	if !usegod {
		checkGodCards = []byte{}
	}
	checkCards, _, gui_num, err := ReSetHandCards_Nomal(handCards, cbCurrentCard, isNormalCard, checkGodCards, seaf)
	//add
	finalCards := ReSetHandwithWeave_Nomal(checkCards, WeaveItem)
	cards_wan := &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	//cards_wan.handCards=append(cards_wan.handCards,checkCards[0:9]...)
	cards_wan.handCards = finalCards[0:9]
	cards_tiao := &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	//cards_tiao.handCards=append(cards_tiao.handCards,checkCards[9:18]...)
	cards_tiao.handCards = finalCards[9:18]
	cards_tong := &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	//cards_tong.handCards=append(cards_tong.handCards,checkCards[18:27]...)
	cards_tong.handCards = finalCards[18:27]
	//处理万，可能返回几种

	result_wan := this.checkDragon_base(cards_wan.handCards, gui_num, cards_wan)
	if !result_wan {
		//万字里面没有，处理条
		result_tiao := this.checkDragon_base(cards_tiao.handCards, gui_num, cards_tiao)
		if !result_tiao {
			//处理筒
			result_tong := this.checkDragon_base(cards_tong.handCards, gui_num, cards_tong)
			if !result_tong {
				return static.CHK_NULL, nil
			}
		}
	}
	//不检查胡
	//return public.WIK_CHI_HU,nil
	//只会有一条龙
	checkObj := this.CreateDragonResult(checkCards, cards_wan, cards_tiao, cards_tong)
	//配完一条龙后，再去掉赖子牌，然后再测一下，能不能胡
	if len(godCards) != 0 && !usegod {
		//重新去掉，因为有的赖子还原了，就在剩牌中去掉
		checkObj.handCards, _, checkObj.lastGodNum, err = ReSetHandCards_Nomal(checkObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
		if err != nil {
			fmt.Println(fmt.Sprintf("重组一条龙出错(%v)", err))
			return static.CHK_NULL, err
		}
	}
	return this.CheckDragon_hu(checkObj, gui_num, eyeMask), nil
}

func (this *CheckHu) CheckDragon(handCards []byte, WeaveItem []static.TagWeaveItem, cbCurrentCard byte, isNormalCard bool, godCards []byte, eyeMask byte, usegod bool, seaf bool) (resulthand []byte, godnum byte, result byte, err error) {

	checkGodCards := []byte{}
	if len(godCards) != 0 {
		static.HF_DeepCopy(&checkGodCards, &godCards)
	}
	if !usegod {
		checkGodCards = []byte{}
	}
	checkCards, _, gui_num, err := ReSetHandCards_Nomal(handCards, cbCurrentCard, isNormalCard, checkGodCards, seaf)
	//add
	finalCards := ReSetHandwithWeave_Nomal(checkCards, WeaveItem)
	cards_wan := &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	//cards_wan.handCards=append(cards_wan.handCards,checkCards[0:9]...)
	cards_wan.handCards = finalCards[0:9]
	cards_tiao := &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	//cards_tiao.handCards=append(cards_tiao.handCards,checkCards[9:18]...)
	cards_tiao.handCards = finalCards[9:18]
	cards_tong := &DragonObj{
		handCards:    make([]byte, 0),
		lostGodCards: make([]byte, 0),
	}
	//cards_tong.handCards=append(cards_tong.handCards,checkCards[18:27]...)
	cards_tong.handCards = finalCards[18:27]
	//处理万，可能返回几种

	result_wan := this.checkDragon_base(cards_wan.handCards, gui_num, cards_wan)
	if !result_wan {
		//万字里面没有，处理条
		result_tiao := this.checkDragon_base(cards_tiao.handCards, gui_num, cards_tiao)
		if !result_tiao {
			//处理筒
			result_tong := this.checkDragon_base(cards_tong.handCards, gui_num, cards_tong)
			if !result_tong {
				return nil, 0, static.CHK_NULL, nil
			}
		}
	}
	//不检查胡
	//return public.WIK_CHI_HU,nil
	//只会有一条龙
	resultObj := this.CreateDragonResult(checkCards, cards_wan, cards_tiao, cards_tong)
	//配完一条龙后，再去掉赖子牌，然后再测一下，能不能胡
	if len(godCards) != 0 && !usegod {
		//重新去掉，因为有的赖子还原了，就在剩牌中去掉
		resultObj.handCards, _, resultObj.lastGodNum, err = ReSetHandCards_Nomal(resultObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
		if err != nil {
			fmt.Println(fmt.Sprintf("重组一条龙出错(%v)", err))
			return nil, 0, static.CHK_NULL, err
		}
	}
	result = this.CheckDragon_hu(resultObj, gui_num, eyeMask)
	err = nil
	return resultObj.handCards, resultObj.lastGodNum, resultObj.huKind, nil
}
