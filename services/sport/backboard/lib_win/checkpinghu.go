package lib_win

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
)

/*
平胡检查，顺着来，太麻烦
反着来
*/

//这个结构先不用
type PingHuObj struct {
	handCards    []byte //去掉龙后的牌数据
	lostGodCards []byte //去掉龙牌所使用到的赖子牌数据
	lastGodNum   byte   //剩下鬼牌 这个基本上用不到
	isDragon     bool   //是不是龙
	huKind       byte
}

//由以前的检牌改成去章，我觉得差不多，先看看 这个不够了
func (this *CheckHu) GetLoseCardIndex(checkcard byte) (result []byte) {
	//只有万条筒，不可能有别的
	index, err := mahlib2.CardToIndex(checkcard)
	if err != nil {
		fmt.Println(fmt.Sprintf("检查平胡获取编号错误（%v）", err))
		return nil
	}
	//如果风牌，就不生成了，因为这种情况要么是将，要么是刻，如果是将，那就是单吊，如果是刻不符合目前平胡的要求
	if checkcard >= 0x31 {
		return nil
	}
	if index != static.INVALID_BYTE {
		remined := index % 9
		if remined < 3 {
			result = append(result, index+1)
			return
		}
		if remined > 5 {
			result = append(result, index-2)
			return
		}
		result = append(result, index+1)
		result = append(result, index-2)
	}
	return
}

//自摸癞子的情况未做处理
//这个特殊了，只做白皮当癞子的处理，其他癞子情况要针对修改
func (this *CheckHu) GetLoseCardIndexEX(checkcard byte, guiCards []byte, isNomal bool) (result []byte) {
	//只有万条筒，不可能有别的
	index, err := mahlib2.CardToIndex(checkcard)
	if err != nil {
		fmt.Println(fmt.Sprintf("检查平胡获取编号错误（%v）", err))
		return nil
	}
	//如果风牌，就不生成了，因为这种情况要么是将，要么是刻，如果是将，那就是单吊，如果是刻不符合目前平胡的要求
	if len(guiCards) == 0 {
		//这个是无配字
		if checkcard >= 0x31 {
			return nil
		}

	} else {
		//这个时候就是白皮配字
		if checkcard >= 0x31 && checkcard <= 0x36 {
			return nil
		}
		if checkcard == 0x37 {
			if isNomal {
				return nil
			} else {
				//独立处理这个癞子
			}

		}
	}
	if index != static.INVALID_BYTE {
		remined := index % 9
		if remined < 3 {
			result = append(result, index+1)
			return
		}
		if remined > 5 {
			result = append(result, index-2)
			return
		}
		result = append(result, index+1)
		result = append(result, index-2)
	}
	return
}

//这里面有风箭牌的可能
func (this *CheckHu) GetCheckCards(checkcard byte) (result []byte) {
	//只有万条筒，不可能有别的
	index, err := mahlib2.CardToIndex(checkcard)
	if err != nil {
		fmt.Println(fmt.Sprintf("检查平胡获取编号错误（%v）", err))
		return nil
	}
	//如果风牌，就不生成了，因为这种情况要么是将，要么是刻，如果是将，那就是单吊，如果是刻不符合目前平胡的要求
	if checkcard >= 0x31 {
		return nil
	}
	if index != static.INVALID_BYTE {
		remined := index % 9
		if remined < 3 {
			result = append(result, checkcard+3)
			return
		}
		if remined > 5 {
			result = append(result, checkcard-3)
			return
		}
		result = append(result, checkcard+3)
		result = append(result, checkcard-3)
	}
	return
}

//
//func (this *CheckHu) CheckPingHu_hu(checkCardObj *PingHuObj,gui_num byte, eyeMask byte,godindex byte) (byte) {
//
//	hu, _:= this.CheckHU.Split_Byte(checkCardObj.handCards,checkCardObj.lastGodNum, eyeMask,true)
//	if hu{
//		//不分类了，根据 好像有问题，如果是赖将，会留2个赖子，并且任意牌
//		if (checkCardObj.lastGodNum==gui_num&&gui_num!=0)||gui_num==0{
//			if eyeMask>1{
//				checkCardObj.huKind|=static.CHK_PING_HU_NOMAGIC<<2
//			}else{
//				checkCardObj.huKind|=static.CHK_PING_HU_NOMAGIC
//			}
//		}else{
//			if eyeMask>1{
//				checkCardObj.huKind|= static.CHK_PING_HU_MAGIC<<2
//			}else{
//				checkCardObj.huKind|= static.CHK_PING_HU_MAGIC
//			}
//		}
//		////独立检查一下能不能硬胡
//		checkCards := make([]byte, 0)
//		static.HF_DeepCopy(&checkCards, &checkCardObj.handCards)
//		checkCards[godindex]+=checkCardObj.lastGodNum
//		hu, _:= this.CheckHU.Split_Byte(checkCards,0, eyeMask,true)
//		if hu{
//			if eyeMask>1{
//				checkCardObj.huKind|=static.CHK_PING_HU_NOMAGIC<<2
//			}else{
//				checkCardObj.huKind|=static.CHK_PING_HU_NOMAGIC
//			}
//		}
//		//-----------------------------------------
//		return checkCardObj.huKind
//	}
//	return static.CHK_NULL
//}
//普通接口 姐妹铺
/*
func  (self *GameLogic_cz_L3f) CheckPingHu(cbCardIndex []byte,WeaveItem []static.TagWeaveItem,cbCurrentCard byte,isNomail bool,guiCards []byte,ChiHuResult *static.TagChiHuResult,quanfeng byte, mengfeng byte)(result bool,err error){

平胡 在确定胡的前提下，再检查,仅限与白板这种赖子情况
*/
func (this *CheckHu) CheckPingHu_Normal(handCards []byte, WeaveItem []static.TagWeaveItem, cbCurrentCard byte) (result bool, err error) {
	//首先检查的牌不放进去
	result = false
	_, err = mahlib2.CheckHandCardsSafe_ex(handCards, cbCurrentCard)
	if err != nil {
		fmt.Println(fmt.Sprintf("平胡检查出问题（%v）", err))
		return
	}
	//倒牌里有碰杠，不算
	checkmask := true
	for _, v := range WeaveItem {
		if v.WeaveKind == static.CHK_NULL {
			break
		}
		if v.WeaveKind == static.WIK_PENG || v.WeaveKind == static.WIK_GANG && checkmask == true {
			checkmask = false
			break
		}
	}
	return checkmask, nil
}

//我觉得简单一点，门风将一对，去掉，检查能不能胡3n+2能胡，就是碰碰胡 可以胡，如果不能胡，那就不是平胡
//剩下的情况，只要去掉check牌，还能胡，那就是平胡
//func (this *CheckHu) CheckPingHu_Normal_sq(handCards []byte,WeaveItem []public.TagWeaveItem,cbCurrentCard byte,isNormalCard bool,godCards[]byte,eyeMask byte,quanfeng byte, mengfeng byte) (result bool,err error) {
//	//第一步，因为有碰碰胡的可能，就不能单用上面的例子，改为
//	//首先检查的牌不放进去
//	result=false
//	_,err=common.CheckHandCardsSafe_ex(handCards,cbCurrentCard)
//	if err!=nil{
//		fmt.Println(fmt.Sprintf("平胡检查出问题（%v）",err))
//		return
//	}
//	//倒牌里有碰杠，不算
//	checkmask:=true
//	for _,v:=range WeaveItem{
//		if v.WeaveKind==public.CHK_NULL{
//			break
//		}
//		if v.WeaveKind ==public.WIK_PENG||v.WeaveKind == public.WIK_GANG&&checkmask==true{
//			checkmask=false
//			break
//		}
//	}
//	if !checkmask{
//		return false,nil
//	}
//
//	//简单一点，只要取和胡牌相隔2位的牌，如果能胡，应该说明不是卡章
//	//首先，如果牌是风箭，并且手牌里只有一张，那么这个就一定是将牌，条件修改
//	if cbCurrentCard!=public.INVALID_BYTE{
//		//牌没加进去
//		if cbCurrentCard>0x31{
//			if isNormalCard{
//				checkIndex,_:=common.CardToIndex(cbCurrentCard)
//				if handCards[checkIndex]==1{
//					//这个就是单吊
//					return false,nil
//				}
//			}else{
//				//鬼牌 风牌全检查是不是==1
//				var i byte
//				for i=0x27;i<0x34;i++{
//					if handCards[i]==1{
//						//凑将
//						return false,nil
//					}
//				}
//			}
//		}
//	}
//	//牌已经加进去了 就比较恶心了 只判断是不是值钱风将
//	quanfengIndex,_:=common.CardToIndex(quanfeng)
//	menfengIndex,_:=common.CardToIndex(mengfeng)
//	//风牌要么是将要么是刻，如果手牌中的这两种牌==2必然是将
//	if handCards[quanfengIndex]==2||handCards[menfengIndex]==2{
//		return false,nil
//	}
//	//圈风和门风不能做将的
//	result=true
//	//有可能是碰碰胡，如果是碰碰胡的话，就是2张，我认为反正剩下的只要不是单张，那就是平胡
//	var checkcards []byte
//	if cbCurrentCard>0x31{
//		checkcards=[]byte{}
//	}else{
//		checkcards=this.GetCheckCards(cbCurrentCard)
//	}
//
//	var gui_1 byte = 0xff
//	var gui_2 byte = 0xff
//	index:= byte(len(godCards))
//	switch index {
//	case 1:
//		gui_1 = godCards[0]
//	case 2:
//		gui_1 = godCards[0]
//		gui_2 = godCards[1]
//	}
//	if len(checkcards)!=0{
//		for _,v:=range checkcards{
//			checkcardNomal:=true
//			if v==gui_1||v==gui_2{
//				checkcardNomal=false
//			}
//			hu,_:=this.CheckHU.GetHuInfo_Byte(handCards, v,checkcardNomal, gui_1, gui_2, eyeMask)
//			if !hu{
//				result=false
//				break
//			}
//		}
//	}else{
//		result=false
//	}
//	return
//}
