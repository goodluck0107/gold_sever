package lib_win

import (
	"github.com/open-source/game/chess.git/pkg/static"
)

//这个是基础判断
/*
断幺九 全幺九都是这里搞
19风都要判断，两位，19各一位，因为可能有全断幺的需求，再带上风箭
总共给4位算了
11111 从左到右 代表有字 风 非19 9 1

*/
func (this *CheckHu) CheckYaoJiu_base(handCards []byte) (result byte) {
	//判断19
	for i := 0; i < 27; i++ {
		remined := i % 9
		if handCards[i] != 0 {
			switch remined {
			case 0:
				//有1
				result |= 0x1
			case 8:
				//有9
				result |= 0x2
			default:
				//死了，非19
				result |= 0x4
			}
		}
	}
	//检查风 只要有一个就行了
	for i := 27; i < 31; i++ {
		if handCards[i] != 0 {
			result |= 0x8
			break
		}
	}
	//检查刻
	for i := 31; i < 34; i++ {
		if handCards[i] != 0 {
			result |= 16
			break
		}
	}
	return
}

//检查幺九
/*
这个里面不检查胡，胡牌在外面检查
19暂时不管赖子
*/
func (this *CheckHu) CheckYaoJiu_Normal(handCards []byte, WeaveItem []static.TagWeaveItem, cbCurrentCard byte, isNormalCard bool, godCards []byte, seaf bool) (result byte, guinum byte, err error) {
	checkCards, _, guinum, err := ReSetHandCards_Nomal(handCards, cbCurrentCard, isNormalCard, godCards, seaf)
	if err != nil {
		return 0, 0, err
	}
	//加入倒牌
	finalCards := ReSetHandwithWeave_Nomal(checkCards, WeaveItem)
	result = this.CheckYaoJiu_base(finalCards)
	return result, guinum, nil
}
