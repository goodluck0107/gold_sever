package lib_win

import (
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	cardmgr2 "github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

type HuLib struct {
	MTableMgr *TableMgr
}

func NewHU(needtable bool) (newHu *HuLib, err error) {
	newHu = &HuLib{}
	if needtable {
		newHu.MTableMgr = &TableMgr{}
		newHu.MTableMgr.Init()
		err = newHu.MTableMgr.LoadTable()
		if err != nil {
			return nil, err
		}
		err = newHu.MTableMgr.LoadFengTable()
		if err != nil {
			return nil, err
		}
	}
	return newHu, nil
}

//硬性要求 34张牌，第三个参数改为赖子牌，目前只判断一个赖子，最大值就是4张赖子
//检查普通胡的，大胡外面去
/*
参数：
1、cbCardIndex []byte 目前所有手牌的34张列表信息，不包含要判断的那张单张
2、修改为一个结构，
3、guicards []byte 鬼牌
返回：
返回的第二个参数目前当硬胡的判断
这个单元一抽为二，因为258将的判断比较恶心了
*/
//20190114 258还原胡和混将还原都是一个情况，直接抽出来
//201901116 添加赖将，如果将又是赖子的话。。
func (this *HuLib) AnalyseHuRestore(cbCardIndex []byte, checkInfo *cardmgr2.CardBaseInfo, eyeMask byte) (hu bool, hu_struct *[4][2]byte) {
	//从上层传下来的牌桌赖子，要么有要么没有
	if len(cbCardIndex) == 0 {
		return false, nil
	}
	if checkInfo != nil {
		return this.GetHuInfo_Byte(cbCardIndex, checkInfo.ID, true, static.INVALID_BYTE, static.INVALID_BYTE, eyeMask)
	} else {
		return this.GetHuInfo_Byte(cbCardIndex, static.INVALID_BYTE, true, static.INVALID_BYTE, static.INVALID_BYTE, eyeMask)
	}

}

//20190104 这个修改为赖子胡判断
func (this *HuLib) AnalyseCard(cbCardIndex []byte, checkInfo *cardmgr2.CardEx, guicards []byte, eyeMask byte) (hu bool, hu_struct *[4][2]byte) {
	var gui_1 byte = 0xff
	var gui_2 byte = 0xff
	index := len(guicards)
	switch index {
	case 1:
		gui_1 = guicards[0]
	case 2:
		gui_1 = guicards[0]
		gui_2 = guicards[1]
	}
	if gui_1 == 0xff && gui_2 == 0xff {
		if checkInfo != nil && !checkInfo.IsGod {
			return false, nil
		}
	}
	//if gui_1 == 0xff && gui_2 == 0xff && !checkInfo.IsGod {
	//	return false, nil
	//}
	//
	if checkInfo != nil {
		return this.GetHuInfo_Byte(cbCardIndex, checkInfo.ID, !checkInfo.IsGod, gui_1, gui_2, eyeMask)
	} else {
		return this.GetHuInfo_Byte(cbCardIndex, static.INVALID_BYTE, true, gui_1, gui_2, eyeMask)
	}

}

// //20190114 恶心的258
// func (this *HuLib) get258HuInfo(cards []byte, card byte, AddToHard bool, gui_num1 byte, gui_num2 byte, eyeMask byte) (hu bool, hu_struct *[4][2]byte) {
// 	//循环去将，只要有一个能胡，那就是硬258的屁胡，不然就是赖258的屁胡
// 	single258 := 0
// 	eye258 := 0
// 	for i := 1; i <= 25; i += 3 {
// 		if cards[i] == 0 {
// 			continue
// 		}
// 		//以硬胡为优先胡法 类似 {123,234,55，赖子刻}这样的牌要55为将的情况 每个将眼都要去掉2张检查一下，如果有
// 		//成对 不一定为将，如{123,234,345,456} 25都不是眼 这个时候判断都不用赖子，赖子就是将，屁胡可能，但是258这项又是硬胡
// 		if cards[i]%2 == 0 {
// 			//去掉1对检查，这个时候应该检查混将吗？
// 			cards[i] -= 2
// 			//查混将试试
// 			this.getHuInfo_Byte(cards, card, AddToHard, gui_num1, gui_num2, 1)
// 		} else {
// 			//三张
// 			single258 += 1
// 		}
// 	}
// 	return false, nil
// }

//赖子测
func (this *HuLib) GetHuInfo_Byte(cards []byte, card byte, AddToHard bool, gui_num1 byte, gui_num2 byte, eyeMask byte) (hu bool, hu_struct *[4][2]byte) {
	hu = false
	index, err := mahlib2.CardToIndex(card)
	if err != nil {
		return false, nil
	}
	var gui_num_1 byte = 0
	var gui_num_2 byte = 0

	gui1index, _ := mahlib2.CardToIndex(gui_num1)
	if gui1index < MAX_INDEX {
		gui_num_1 = cards[gui1index] % 0xff
		cards[gui1index] = 0
	}

	gui2index, _ := mahlib2.CardToIndex(gui_num2)
	if gui2index < MAX_INDEX {
		gui_num_2 = cards[gui2index] % 0xff
		cards[gui2index] = 0
	}
	if index != static.INVALID_BYTE {
		if AddToHard {
			cards[index]++
			hu, hu_struct = this.Split_Byte(cards, gui_num_1+gui_num_2, eyeMask, true)
		} else {
			hu, hu_struct = this.Split_Byte(cards, gui_num_1+gui_num_2+1, eyeMask, true)
		}
		if AddToHard {
			cards[index]--
		}
	} else {
		hu, hu_struct = this.Split_Byte(cards, gui_num_1+gui_num_2, eyeMask, true)
	}

	if gui1index < MAX_INDEX {
		cards[gui1index] = gui_num_1
	}
	if gui2index < MAX_INDEX {
		cards[gui2index] = gui_num_2
	}
	return
}

func check_Byte(gui byte, eye_num byte, gui_num byte, gui_sum byte) (bool, byte) {
	//这里应该判断 是不是255? 255代表没用到赖子
	if gui == static.INVALID_BYTE {
		return false, 0
	}
	gui_sum += gui
	if gui_sum > gui_num {
		return false, 0
	}
	if eye_num == 0 {
		return true, gui_sum
	}
	return gui_sum+(eye_num-1) <= gui_num, gui_sum
}

// 考虑可能出现缺中门的情况，所有不能用切片
func (this *HuLib) Split_Byte(cards []byte, gui_num byte, eyeMask byte, mastEye bool) (hu bool, returnstruct *[4][2]byte) {
	var eye_num byte = 0 //这个眼会叠加
	var temp_eye_num byte = 0
	var gui_sum byte = 0
	var gui byte = 0
	ret := false
	//检查万
	cards1 := cards[:9]
	gui, eye_num = this._split_Byte(cards1, gui_num, true, eye_num, eyeMask)
	ret, gui_sum = check_Byte(gui, eye_num, gui_num, gui_sum)
	if ret == false {
		return ret, nil
	}
	//创建二维数组
	tempstruct := [4][2]byte{}
	// 万需要多少个鬼，有多少个眼
	tempstruct[0] = [2]byte{gui, eye_num}
	temp_eye_num = eye_num
	//检查条
	cards2 := cards[9:18]
	gui, eye_num = this._split_Byte(cards2, gui_num-gui_sum, true, eye_num, eyeMask)
	ret, gui_sum = check_Byte(gui, eye_num, gui_num, gui_sum)
	if ret == false {
		return ret, nil
	}
	// 条需要多少个鬼，有多少个眼
	tempstruct[1] = [2]byte{gui, eye_num - temp_eye_num}
	temp_eye_num = eye_num
	//检查筒
	cards3 := cards[18:27]
	gui, eye_num = this._split_Byte(cards3, gui_num-gui_sum, true, eye_num, eyeMask)
	ret, gui_sum = check_Byte(gui, eye_num, gui_num, gui_sum)
	if ret == false {
		return ret, nil
	}
	// 筒需要多少个鬼，有多少个眼
	tempstruct[2] = [2]byte{gui, eye_num - temp_eye_num}
	temp_eye_num = eye_num
	cards4 := cards[27:34]
	//处理掉258的权位 20190109
	eyeMask &= 1
	gui, eye_num = this._split_Byte(cards4, gui_num-gui_sum, false, eye_num, eyeMask)
	ret, gui_sum = check_Byte(gui, eye_num, gui_num, gui_sum)
	if ret == false {
		return ret, nil
	}
	// 风字牌需要多少个鬼，有多少个眼
	tempstruct[3] = [2]byte{gui, eye_num - temp_eye_num}
	//记录下来 这个地方就当赖子对当将了
	if eye_num == 0 && mastEye {
		ret = (gui_sum+2 <= gui_num)
	}
	return ret, &tempstruct
}

//
// func checkPerfect(cards []int, index int) *MhjongMatrix {
// 	var marks []int
// 	var maMarks = MhjongMatrix{}
// 	// cardsList := getcardsListByRemoveTwoCards(cards)
// 	checkindex := len(cards)
// 	tempcheck := make([]int, checkindex)
// 	for i := 0; i < checkindex-2; i++ {
// 		copy(tempcheck, cards)
// 		//先去顺子再去克
// 		marks1 := removeThreeLinkCards(tempcheck[i:], index)
// 		//我去
// 		for j, _ := range marks1 {
// 			marks1[j] = marks1[j] + i + 1
// 		}
// 		marks2 := removeTheSameThreeCards(tempcheck)
// 		//这里参考的是14张牌，普通胡法是按照3n+2=14来算的，所以n最多是4 还有四暗刻，也只有4，不过手牌能达到18
// 		if len(marks1)+len(marks2) > 4 {
// 			continue
// 		}
// 		isPerfect := checkMatrixAllElemEqualZero(tempcheck)
// 		if isPerfect {
// 			marks = append(marks1, marks2...)
// 			//check一下
// 			if !checkRepByLoop(maMarks, marks) {
// 				maMarks = append(maMarks, marks)
// 			}

// 		}
// 	}
// 	//类似444这样的牌，应该还有一种记录[14,24,34],虽然两个都差不多，一个是一色4顺 211加两个赖子应该有，222这种，[15,15]
// 	return &maMarks
// }

//注意倒数第二个参数代表是不是查风字牌
//返回值修改为，第二个鬼牌个数，第三个是眼牌个数
func (this *HuLib) _split_Byte(cards []byte, gui_num byte, chi bool, eye_num byte, eyeMask byte) (byte, byte) {
	key := 0
	num := 0

	for i := 0; i < len(cards); i++ {
		key = key*10 + int(cards[i])
		num = num + int(cards[i])
	}

	if num == 0 {
		return 0, eye_num
	}
	for i := byte(0); i <= gui_num; i++ {
		yu := (num + int(i)) % 3
		if yu == 1 {
			continue
		}
		eye := (yu == 2)

		if this.MTableMgr.check(key, int(i), eye, eyeMask, chi) {
			if eye {
				eye_num++
			}
			return i, eye_num
		}
	}
	return static.INVALID_BYTE, 0
}
