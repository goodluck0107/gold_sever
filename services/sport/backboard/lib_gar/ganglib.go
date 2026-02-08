package lib_gar

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	"github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

// 20181129 苏大强 设计想法
/*
手牌生成预处理（杠碰）


//检查是不是可以杠和碰 基础检查的
/*
//对外函数，必要检查所有参数
暗杠：判断的时候手牌有4张（原本就有或者3+1） 手牌数至少4张
	回头杠：已经倒了3张，又起了一张  至少1张
	蓄杠：已经倒了3张，手上还有一张  至少1张
	明杠：手上3张，别人打了一张（必须操作权最大，别人胡牌的情况不可能）至少4张

参数说明 ：
1、玩家的手牌，序列化
2、玩家倒下的牌
3、检查的牌
4、碰杠结果结构

根据操作的可能性：
	起一张（包括杠起）：
	 暗杠、回头杠、蓄杠
	 别人打的章：
	  明杠、碰
	  吃、碰后：
	  暗杠、蓄杠

20181203 第一版 返回值，碰杠权，错误
*/
func GetGangPengResult(cbCardIndex []byte, WeaveItem []static.TagWeaveItem, checkCard *card_mgr.CardBaseInfo) (readyforOpe *GongPengResult, err error) {
	if checkCard != nil && !mahlib2.IsMahjongCard(checkCard.ID) {
		return nil, errors.New(fmt.Sprintf("检查牌（%x）越界", checkCard.ID))
	}

	if readyforOpe == nil {
		readyforOpe = &GongPengResult{
			Chigongcards:  []byte{},
			Xugongcards:   []byte{},
			BackgongCards: []byte{},
			HidgongCards:  []byte{},
			PengCards:     []byte{},
		}
	}
	//先处理别人打出来的牌，这个时候不用去获取weave的数据,只有明杠和碰的可能，只可能一次
	if checkCard != nil && checkCard.Origin == card_mgr.ORIGIN_TABILE {
		index, err := mahlib2.CardToIndex(checkCard.ID)
		if err != nil {

			return nil, err
		}
		//直接检查手牌中这张牌有没有3张
		if cbCardIndex[index] == 3 {
			readyforOpe.Chigongcards = append(readyforOpe.Chigongcards, checkCard.ID)
			//取巧一下，因为别人打的牌，能明杠就能碰，所以只要在做明杠数据的时候再加个碰就行了
			// readyforOpe.PengCards = append(readyforOpe.PengCards, checkCard.ID)
			return readyforOpe, nil
		}
		if cbCardIndex[index] == 2 {
			//但是碰牌不能再明杠，除非朝天杠
			readyforOpe.PengCards = append(readyforOpe.PengCards, checkCard.ID)
			return readyforOpe, nil
		}
		return readyforOpe, nil
	}
	//	拿出手上4的数据 这个是暗杠
	checkHandArr := ClassifyGPCards(cbCardIndex)
	//获取weave的数据 蓄杠，回头杠
	checkWeaveArr := GetWeavePengCards(WeaveItem)
	//不加牌判断的暗杠，下面都有可能触发暗杠
	if len(checkHandArr.fourcards) != 0 {
		readyforOpe.HidgongCards = append(readyforOpe.HidgongCards, checkHandArr.fourcards[:]...)

	}
	//可能有多个蓄杠 ，吃碰，发牌都可能，所有先做了
	if len(checkWeaveArr) != 0 {
		for _, v := range checkWeaveArr {
			cardindex, err := mahlib2.CardToIndex(v)
			if err != nil {
				fmt.Println(err)
				continue
			}
			//实际上就一章
			if cbCardIndex[cardindex] == 1 {
				//如果手牌有的话，那么可以蓄杠
				readyforOpe.Xugongcards = append(readyforOpe.Xugongcards, v)
			}
		}
		//去掉已经判断的蓄杠牌 不需要多次一举，因为最后的回头杠也就只有一个了
		// for _, v := range readyforOpe {
		// 	if v.OpeChiPengGang&XUGONG > 0 {
		// 		checkWeaveArr = common.Removecard(checkWeaveArr, v.Cardbaseinfo.ID)
		// 	}
		// }
	}
	//没有判断的牌，就是吃碰后的操作，只能有暗杠和蓄杠
	if checkCard == nil {
		return
	}
	//到这里就只有发牌到手上的情况了，在上面的基础上追加可能的暗杠和回头杠就行了
	if checkCard.Origin == card_mgr.ORIGIN_NOM {
		index, err := mahlib2.CardToIndex(checkCard.ID)
		if err != nil {
			fmt.Println(err)
			//这个地方要注意一下，外面的判断是两个，先判断返回权，在看err
			return readyforOpe, err
		}
		//暗杠在上面的基础上添加，手牌3张再加起牌的情况
		if cbCardIndex[index] == 3 {
			readyforOpe.HidgongCards = append(readyforOpe.HidgongCards, checkCard.ID)
			return readyforOpe, nil
		}
		//回头杠，检查倒牌
		if len(checkWeaveArr) != 0 {
			for _, v := range checkWeaveArr {
				//实际上就一章可以回头
				if v == checkCard.ID {
					readyforOpe.BackgongCards = append(readyforOpe.BackgongCards, v)
					break
				}
			}
		}
	}
	return
}

//生成可发给客户端的数据
func AnalyseGangPeng(readyforOpe *GongPengResult) (result byte, ope []mahlib2.CardOpeInfo) {
	if readyforOpe == nil {
		return NOOPERATOR, nil
	}
	//碰和明杠只能有一个，所以有了就可以退出了
	if len(readyforOpe.Chigongcards) != 0 {
		result |= CHIGONG | PENG
		//碰牌只有一个，不需要这么麻烦
		newOpe, err := mahlib2.NewCardOpe(readyforOpe.Chigongcards[0], result)
		if err != nil {
			fmt.Println(err)
			return NOOPERATOR, nil
		}
		ope = append(ope, *newOpe)
		return
	}
	if len(readyforOpe.PengCards) != 0 {
		result |= PENG
		//碰牌只有一个，不需要这么麻烦
		newOpe, err := mahlib2.NewCardOpe(readyforOpe.PengCards[0], result)
		if err != nil {
			fmt.Println(err)
			return NOOPERATOR, nil
		}
		ope = append(ope, *newOpe)
		return
	}
	//上面的不能和下面的共存
	//暗杠、蓄杠、回头杠可以在一起，所以必须都判断完
	//蓄杠可能有多个
	if len(readyforOpe.Xugongcards) != 0 {
		result |= XUGONG
		for _, v := range readyforOpe.Xugongcards {
			newOpe, err := mahlib2.NewCardOpe(v, XUGONG)
			if err != nil {
				fmt.Println(err)
				continue
			}
			ope = append(ope, *newOpe)
		}
	}
	//回头杠就一个
	if len(readyforOpe.BackgongCards) != 0 {
		result |= BACKGONG
		newOpe, err := mahlib2.NewCardOpe(readyforOpe.BackgongCards[0], BACKGONG)
		if err != nil {
			fmt.Println(err)
			return NOOPERATOR, nil
		}
		ope = append(ope, *newOpe)
	}
	//暗杠可能有多个
	if len(readyforOpe.HidgongCards) != 0 {
		result |= HIDGONG
		for _, v := range readyforOpe.HidgongCards {
			newOpe, err := mahlib2.NewCardOpe(v, HIDGONG)
			if err != nil {
				fmt.Println(err)
				continue
			}
			ope = append(ope, *newOpe)
		}
	}
	return
}
