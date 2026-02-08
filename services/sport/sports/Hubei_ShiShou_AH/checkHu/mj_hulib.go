package ssahCheckHu

import (
	hulib2 "github.com/open-source/game/chess.git/services/sport/backboard/lib_win"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
)

type item struct {
	eye    bool
	guiNum int
	tfn    bool
}

type Ptbl struct {
	arrayNum int
	mNums    [4]int
	match    [4]int
	m        [4][5]item
}

type mjHulib struct {
	ProbabilityItemTable Ptbl
	isTFN                bool
}

const MAX_CARD = 34

var _mjHulib_Ins__ *mjHulib = nil

func GetMjHulibInstance() *mjHulib {
	if _mjHulib_Ins__ == nil {
		_mjHulib_Ins__ = &mjHulib{}
	}
	return _mjHulib_Ins__
}

//func (this *mjHulib) tblReset() {
//	if this.checkTbl == nil {
//		syslog.Logger().Errorln("checkTbl是nil 需要load")
//		var err error
//		this.checkTbl, err = modules.GetHuTable()
//		public.HF_CheckErr(err)
//	}
//}

func (this *mjHulib) reset() {
	this.ProbabilityItemTable.arrayNum = 0
	for i := 0; i < len(this.ProbabilityItemTable.mNums); i++ {
		this.ProbabilityItemTable.mNums[i] = 0
	}
	for i := 0; i < len(this.ProbabilityItemTable.match); i++ {
		this.ProbabilityItemTable.match[i] = 0
	}
	for i := 0; i < len(this.ProbabilityItemTable.m); i++ {
		for j := 0; j < len(this.ProbabilityItemTable.m[i]); j++ {
			this.ProbabilityItemTable.m[i][j].eye = false
			this.ProbabilityItemTable.m[i][j].guiNum = 0
			this.ProbabilityItemTable.m[i][j].tfn = false
		}
	}
}

func (this *mjHulib) getHuInfo(meta *components2.PlayerMeta, cards []int, guiIndex1 int, isHaveCHH bool, agrs ...int) uint64 {
	this.reset()
	curCard := -1
	//guiIndex2 := -1
	if len(agrs) >= 1 {
		curCard = agrs[0]
	}
	//if len(agrs) == 2 {
	//	guiIndex2 = agrs[1]
	//}
	tmpCards := cards
	cardCount := 0
	for i := 0; i < len(tmpCards); i++ {
		if tmpCards[i] > 0 {
			cardCount += tmpCards[i]
		}
	}
	if cardCount%3 != 2 {
		return GameHuNull
	}
	if curCard != -1 && curCard != MAX_CARD {
		tmpCards[curCard] += 1
	}
	//两张鬼牌的索引
	guiNum := 0
	if guiIndex1 != -1 && guiIndex1 != MAX_CARD {
		guiNum += tmpCards[guiIndex1]
		tmpCards[guiIndex1] = 0
	}
	//if guiIndex2 != -1 && guiIndex2 != MAX_CARD {
	//	guiNum += tmpCards[guiIndex2]
	//	tmpCards[guiIndex2] = 0
	//}

	isHaveMagic := GameHuNull
	//if guiNum > 0 {
	//	isHaveMagic = GameHuMagic
	//}
	noPx7d := GameHuNull
	noPxJYS := GameHuNull
	noPx7d = GetmjMacthInstance().Check7D(meta, &cardData{cards: tmpCards, guiNum: guiNum}, isHaveCHH)
	noPxJYS = GetmjMacthInstance().CheckJYS(meta, &cardData{cards: tmpCards, guiNum: guiNum})
	qys := GetmjMacthInstance().CheckQYS(meta)
	if !this._split(tmpCards, guiNum) { //验证花色是否是个牌型
		return isHaveMagic | noPx7d | noPxJYS
	}
	ret := this.checkProbability(&this.ProbabilityItemTable, guiNum)
	if ret { //能胡
		huType := GetmjMacthInstance().matchCards(meta, &cardData{cards: tmpCards, guiNum: guiNum}, isHaveCHH)
		if huType != GameHuNull || qys != GameHuNull {
			return isHaveMagic | huType | qys | noPxJYS
		}
		if noPxJYS != GameHuNull {
			return isHaveMagic | noPxJYS
		}
		return isHaveMagic | GameHuPiH
	}
	return isHaveMagic | noPx7d | noPxJYS
}

func (this *mjHulib) getHuInfoLog(cards []int, guiIndex1 int, agrs ...int) bool {
	this.reset()
	curCard := -1
	if len(agrs) >= 1 {
		curCard = agrs[0]
	}
	tmpCards := cards
	cardCount := 0
	for i := 0; i < len(tmpCards); i++ {
		if tmpCards[i] > 0 {
			cardCount += tmpCards[i]
		}
	}
	if cardCount%3 != 2 {
		return false
	}
	if curCard != -1 && curCard != MAX_CARD {
		tmpCards[curCard] += 1
	}
	//两张鬼牌的索引
	guiNum := 0
	if guiIndex1 != -1 && guiIndex1 != MAX_CARD {
		guiNum += tmpCards[guiIndex1]
		tmpCards[guiIndex1] = 0
	}
	if !this._split(tmpCards, guiNum) { //验证花色是否是个牌型
		return false
	}
	ret := this.checkProbability(&this.ProbabilityItemTable, guiNum)
	return ret
}

func (this *mjHulib) checkProbability(ptbl *Ptbl, gui_num int) bool {
	// 全是鬼牌
	if ptbl.arrayNum == 0 {
		return gui_num >= 2
	}
	// 只有一种花色的牌的鬼牌
	if ptbl.arrayNum == 1 {
		return true
	}
	// 尝试组合花色，能组合则胡
	for i := 0; i < ptbl.match[0]; i++ {
		item := ptbl.m[0][i]
		eye := item.eye
		gui := gui_num - item.guiNum
		if this.checkProbabilitySub(ptbl, eye, gui, 1, ptbl.arrayNum) {
			return true
		}
	}
	return false
}

func (this *mjHulib) checkProbabilitySub(ptbl *Ptbl, eye bool, gui_num int, level int, maxLevel int) bool {
	for i := 0; i < ptbl.match[level]; i++ {
		item := ptbl.m[level][i]
		if eye && item.eye {
			continue
		}
		if gui_num < item.guiNum {
			continue
		}
		if level < maxLevel-1 {
			if this.checkProbabilitySub(ptbl, eye || item.eye, gui_num-item.guiNum, level+1, ptbl.arrayNum) {
				return true
			}
			continue
		}
		if !eye && !item.eye && item.guiNum+2 > gui_num {
			continue
		}
		return true
	}
	return false
}

/*
以下为匹配单花色是否成牌型
*/
func (this *mjHulib) _split(cards []int, gui_num int) bool {
	//Ptbl := &this.ProbabilityItemTable
	//验证四个花色   万：0-8  条：9-17  筒：18-26  字：27-33
	if !this.splitColor(cards, gui_num, 0, 0, 8, true) {
		return false
	}
	if !this.splitColor(cards, gui_num, 1, 9, 17, true) {
		return false
	}
	if !this.splitColor(cards, gui_num, 2, 18, 26, true) {
		return false
	}
	if !this.splitColor(cards, gui_num, 3, 27, 33, false) {
		return false
	}
	return true
}

func (this *mjHulib) splitColor(cards []int, gui_num int, color int, min int, max int, chi bool) bool {
	// 牌索引,癞子数量，当前花色 ，当前花色开始索引值，当前花色结束索引值，是否可以吃，结构体
	key := 0 // 表中的key
	num := 0 // 当前花色的牌 有多少张
	for i := min; i <= max; i++ {
		key = key*10 + cards[i]
		num = num + cards[i]
	}
	if num == 0 { //如果当前花色没有牌。则直接返回true 成牌型  主要做验证用 实际不一定会走进来
		return true
	}
	if !this.listProbability(color, gui_num, num, key, chi) {
		return false
	}
	return true
}

func (this *mjHulib) listProbability(color int, gui_num int, num int, key int, chi bool) bool {
	//当前花色，癞子数量，当前花色有多少个牌，验证key，是否可以吃，结构体
	ptbl := &this.ProbabilityItemTable
	cIndex := color
	anum := ptbl.arrayNum

	for i := 0; i <= gui_num; i++ {
		eye := false        //初始化 无将
		yu := (num + i) % 3 //验证当时牌数是否可以满足牌型
		if yu == 1 {        //余1 牌数不满足牌型
			continue
		} else if yu == 2 { //如果余2 则牌数量是 带将的牌数
			eye = true
		}
		var eyeStatus byte
		if eye {
			eyeStatus = 1 //status ： 0不查将 1查乱将 2查258将
		}
		//modules.huTable.CheckTBl(key,i,eye,eyeStatus,chi)
		//	this.checkTbl.CheckTBl(key,i,eye,eyeStatus,chi)
		oldCheckClass, _ := hulib2.GetOldDataTableInstance()
		if oldCheckClass.CheckTBl(key, i, eye, eyeStatus, chi) {
			ptbl.m[anum][ptbl.match[anum]].eye = eye
			ptbl.m[anum][ptbl.match[anum]].guiNum = i
			ptbl.match[anum]++
			ptbl.mNums[cIndex]++
		}
		//if common.CheckTBL(key,i,eye,eyeStatus,chi) {
		//	ptbl.m[anum][ptbl.match[anum]].eye = eye
		//	ptbl.m[anum][ptbl.match[anum]].guiNum = i
		//	ptbl.match[anum]++
		//	ptbl.mNums[cIndex]++
		//}
	}
	if ptbl.mNums[cIndex] <= 0 || ptbl.match[anum] <= 0 { //通过上面的判断 如果查表没有Key的话 必定不加,所以直接返回错误
		return false
	}
	ptbl.arrayNum++
	return true
}
