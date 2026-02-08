package ssahCheckHu

import (
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
)

type mjMacth struct {
	isChi bool
}

var _mjMacth_Ins__ *mjMacth = nil

func GetmjMacthInstance() *mjMacth {
	if _mjMacth_Ins__ == nil {
		_mjMacth_Ins__ = &mjMacth{}
	}
	return _mjMacth_Ins__
}

func (this *mjMacth) CheckQYS(meta *components2.PlayerMeta) uint64 {

	arr := meta.WeaveItemArray
	n := int(meta.WeaveItemCount)
	this.isChi = false
	cpgArr := make([]int, 0)
	for i := 0; i < n; i++ {
		if !this.isChi && (arr[i].WeaveKind == static.WIK_LEFT || arr[i].WeaveKind == static.WIK_CENTER || arr[i].WeaveKind == static.WIK_RIGHT) {
			this.isChi = true
		}
		num := arr[i].CenterCard
		cpgArr = append(cpgArr, Num16to10(int(num)))
	}
	var colorArr [4]int
	for i := 0; i < len(cpgArr); i++ {
		for j := 0; j < len(colorArr); j++ {
			if (cpgArr[i]-j*10) > 0 && (cpgArr[i]-j*10) < 10 {
				if colorArr[j] < 1 {
					colorArr[j]++
				}
			}
		}
	}
	for tempIndexArr, i := meta.CardIndex, 0; i < static.MAX_INDEX; i++ {
		for j := 0; j < len(colorArr); j++ {
			cardCount := int(tempIndexArr[i])
			if cardCount > 0 {
				if cardNum := indexChangeZi(i); (cardNum-j*10) > 0 && (cardNum-j*10) < 10 {
					if colorArr[j] < 1 {
						colorArr[j]++
					}
				}
			}
		}
	}
	isHuaNum := 0
	for i := 0; i < len(colorArr); i++ {
		if colorArr[i] > 0 {
			isHuaNum++
		}
	}
	if isHuaNum < 2 {
		return GameHuQYS
	}
	return GameHuNull
}

func (this *mjMacth) matchCards(meta *components2.PlayerMeta, data *cardData, isHaveCHH bool) uint64 {
	return this.Check7D(meta, data, isHaveCHH) //| this.CheckPPH(data)
}

type cardData struct {
	cards  []int
	guiNum int
}

// 碰碰胡
func (this *mjMacth) CheckPPH(data *cardData) uint64 {
	if this.isChi { //判断有没有吃
		return GameHuNull
	}
	need := 0
	yu1 := 0
	for i := 0; i < 34; i++ {
		if data.cards[i]%3 == 1 {
			yu1 += 1
		}
		if data.cards[i]%3 != 0 {
			need += 1
		}
	}
	if yu1 > 1 {
		need += yu1 - 1
	}
	if need <= data.guiNum+1 {
		return GameHuPPH
	}
	return GameHuNull
}

// 7对
func (this *mjMacth) Check7D(meta *components2.PlayerMeta, data *cardData, isHaveCHH bool) uint64 { //isPx  是 是否成牌型 中的 牌型 拼音缩写，虽然low 但是不知道用什么更好
	if int(meta.WeaveItemCount) > 0 {
		return GameHuNull
	}
	need := 0
	temapCardCount := 0 //手牌张数做一个记录，防止上面的数据被初始化，手牌必须14张 才能胡 7 对
	for i := 0; i < 34; i++ {
		temapCardCount += data.cards[i]
		if data.cards[i]%2 != 0 {
			need += 1
		}
	}
	xlog.Logger().Debugln("手牌数量是多少", temapCardCount)
	if (temapCardCount + data.guiNum) != 14 {
		return GameHuNull
	}
	if need <= data.guiNum {
		//return GameHu7D
		isHHCount := 0 //豪华的数量
		for i := 0; i < 34; i++ {
			if data.cards[i] == 4 {
				isHHCount++
			}
		}
		if isHaveCHH {
			if isHHCount >= 3 {
				return GameHu7D3
			}
			if isHHCount == 2 {
				return GameHu7D2
			}
			if isHHCount == 1 {
				return GameHu7D1
			}
		} else {
			//20201229 苏大强 应该改这里吧
			if isHHCount >= 1 {
				return GameHu7D1
			}
		}

		if isHHCount == 0 {
			return GameHu7D
		}
	}
	return GameHuNull
}

//不成牌型单独看
func (this *mjMacth) CheckJYS(meta *components2.PlayerMeta, data *cardData) uint64 {
	cards := data.cards
	isSe := GameHuNull
	tempCount := 0
	// tfn := "1, 4, 7, 10, 13, 16, 19, 22, 25"  //	 	[1, 4, 7, 10, 13, 16, 19, 22, 25]
	tfn := map[int]struct{}{
		1:  {},
		4:  {},
		7:  {},
		10: {},
		13: {},
		16: {},
		19: {},
		22: {},
		25: {},
	}
	for i := 0; i < 34; i++ {
		if cards[i] > 0 {
			if _, ok := tfn[i]; !ok {
				tempCount++
			}
		}
	}
	for i := 0; i < len(meta.WeaveItemArray); i++ {
		if meta.WeaveItemArray[i].WeaveKind == static.WIK_LEFT ||
			meta.WeaveItemArray[i].WeaveKind == static.WIK_CENTER ||
			meta.WeaveItemArray[i].WeaveKind == static.WIK_RIGHT {
			return GameHuNull
		}
		if int(meta.WeaveItemArray[i].CenterCard) != 0 {
			num := trimCardIndex(Num16to10(int(meta.WeaveItemArray[i].CenterCard)))
			if _, ok := tfn[num]; !ok {
				tempCount++
			}
		}
	}
	if tempCount == 0 {
		isSe = GameHuJYS
	}
	return isSe
}

// 清一色
//func (this *mjMacth) CheckQYS(data *cardData) int {
//	arr := this.setColorArr()
//	mNums := GetMjHulibInstance().ProbabilityItemTable.mNums
//	hua := GetMjHulibInstance().ProbabilityItemTable.arrayNum
//	colorNum := 0
//	ci := -1
//	for i := 0; i < len(arr); i++ {
//		if arr[i] > 0 {
//			colorNum++
//			ci = i
//		}
//	}
//	if colorNum > 1 {
//		return CHK_NULL
//	}
//	if hua == 0 {
//		if colorNum == 1 {
//			return CHK_NULL
//		}
//	} else if hua == 1 {
//		if colorNum == 0 {
//			return CHK_NULL
//		}
//		if colorNum == 1 {
//			if mNums[ci] > 0 {
//				return GameHuQYS
//			}
//		}
//	}
//	return CHK_NULL
//}

type CompareFunc func(interface{}, interface{}) int

func (this *mjMacth) indexOf(a []interface{}, e interface{}, cmp CompareFunc) int {
	n := len(a)
	var i int = 0
	for ; i < n; i++ {
		if cmp(e, a[i]) == 0 {
			return i
		}
	}
	return -1
}
