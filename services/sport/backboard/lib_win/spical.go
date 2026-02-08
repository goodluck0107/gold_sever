package lib_win

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	"github.com/open-source/game/chess.git/services/sport/backboard/singleton_card"
)

//这个是基础判断，所有的牌都符合成对而已
func (this *CheckHu) CheckGodHu(checkCtx *card_mgr.CheckHuCTX) bool {
	if (checkCtx.CheckGodOrg != nil && checkCtx.CheckGodOrg.GodNum == 0) || (checkCtx.CheckGodOrg == nil) {
		return false
	}
	refList := []byte{0x13, 0x16, 0x19, 0x21, 0x24, 0x27, 0x01, 0x04, 0x07}

	var mask258 byte = 1
	if checkCtx.Mask258 != 0 {
		mask258 = 2
	}
	checkcardBaseInfo := &card_mgr.CardEx{
		IsGod: false,
	}
	hardcard := make([]byte, 0)
	static.HF_DeepCopy(&hardcard, &checkCtx.CheckCtx.CbCardIndex)
	result := true
	for i, l := 0, len(refList); i < l; i++ {
		checkcardBaseInfo.ID = refList[i]
		result, _ := this.CheckHU.AnalyseCard(hardcard, checkcardBaseInfo, checkCtx.GodInfo.GetGuiInfo(), mask258)
		if !result {
			return result
		}
	}
	//----------------------------
	return result
}

//普通接口 见字胡 必然有赖子的，没赖子的就不用检查了
//因为兼容2个赖子的情况，所以最好是传个切片
/*
eyeMask：
258将必然是大于2
乱将是1
*/
func (this *CheckHu) CheckGodHu_Normal(handCards []byte, godCards []byte, eyeMask byte) (result bool, err error) {
	_, err = mahlib2.CheckHandCardsSafe(handCards)
	if err != nil {
		return false, err
	}
	refList := []byte{0x13, 0x16, 0x19, 0x22, 0x25, 0x28, 0x01, 0x04, 0x07}
	checkCards := make([]byte, 0)
	static.HF_DeepCopy(&checkCards, &handCards)
	var gui_1 byte = 0xff
	var gui_2 byte = 0xff
	index := len(godCards)
	switch index {
	case 1:
		gui_1 = godCards[0]
	case 2:
		gui_1 = godCards[0]
		gui_2 = godCards[1]
	}
	result = true
	for i, l := 0, len(refList); i < l; i++ {
		checkcard := refList[i]
		//20200813  苏大强 遇到个bt的情况，手上已经有4张一样的牌了，checkcard也是这一张，14张手牌出现4张的有3种可能，16张是4中，check9张还是够的
		if index, err := mahlib2.CardToIndex(checkcard); err == nil {
			if checkCards[index] == 4 {
				continue
			}
			result, _ = this.CheckHU.GetHuInfo_Byte(checkCards, checkcard, true, gui_1, gui_2, eyeMask)
			//result, _ := this.CheckHU.AnalyseCard(hardcard, checkcardBaseInfo, checkCtx.GodInfo.GetGuiInfo(), eyeMask)
			if !result {
				return result, nil
			}
		}

	}
	//----------------------------
	return result, nil
}

/*
//20190227 苏大强 初版甩字胡构想
20190301 暂时放在这里，因为甩字胡，不需要番型检查，而且必须是3n+2型的
最后问了下，可以这样搞，已判断牌的为中心牌（非赖子），生成一个切片，组成顺子需要哪几张
这里有个坑的地方，因为别人打赖子是还原的，那么就太方便了，如果打的赖子是gui，那就坑的太大了
所以这个有局限，必须切记，必须是别人打的牌（赖子要还原的）
废弃，这个检查方式不对
20190402 我感觉甩字胡还有问题，因为赖子还原的情况未判断
*/
func (this *CheckHu) CheckShuaizhiHu(checkCtx *card_mgr.CheckHuCTX) bool {
	//必须只有一个赖子，手牌不低于4章或者是判牌结构不低于5章
	handNum := checkCtx.GetHandCardNum()
	//20190301 注意进来判断，必须是吃胡状态，自摸状态不进来，那么以目前的情况，没问题；如果别人打的赖子也当god，这个地方就完蛋了
	//目前这个地方可能有个问题，god可能是发到手上的，目前对甩字胡的判断，就是吃胡，这个就过去了
	//这个godnum的问题，改为godnum==0
	if (checkCtx.CheckGodOrg != nil && checkCtx.CheckGodOrg.GodNum != 1) || (checkCtx.CheckGodOrg == nil) || handNum < 5 || checkCtx.GuiHu_struct == nil {
		return false
	}
	//check 硬胡，如果手牌上有切片中的牌，那么减一，去赖子，然后检查硬胡
	hardcard := make([]byte, 0)
	static.HF_DeepCopy(&hardcard, &checkCtx.CheckCtx.CbCardIndex)
	//去赖子,根据上面的判断只会有一张呢
	var guiIndex byte = static.INVALID_BYTE
	for _, v := range checkCtx.CheckGodOrg.GodCardInfo {
		if v.Num == 1 {
			guiIndex, _ = mahlib2.CardToIndex(v.ID)
			//去赖子 因为只有一张
			hardcard[guiIndex] = 0
			break
		}
	}
	var mask258 byte = 1
	if checkCtx.Mask258 != 0 {
		mask258 = 2
	}
	return this.isShuaizhiHu(hardcard, mask258)
	//如果不是最多14章（3*3后就4章判断），然后应该就是11（3*2）、8 （3）这里就麻烦了，判断的对象是checkCtx.CheckCtx.CbCardIndex，这个时候hu_struct就只能做借鉴了
}

//20190403 甩字胡单独，不用检查来源牌，只是手牌，去1张赖子，1对，1单张，13-4=9 ，检查3n的情况即可
//目前只能有一个赖子，切记
func (this *CheckHu) CheckShuaizhiHu_Normal(handCards []byte, godCards []byte, eyeMask byte) (result bool, err error) {
	handNum := 0
	handNum, err = mahlib2.CheckHandCardsSafe(handCards)
	if err != nil {
		return false, err
	}
	//本身就要去掉一张赖子，一张普通，一对，4章；胡牌手牌，1、4、7、10、13
	if handNum < 4 {
		//不用检查，不符合甩字胡check
		return false, nil
	}
	if len(godCards) == 0 {
		return false, errors.New(fmt.Sprintf("CheckShuaizhiHu_Normal 没有赖子牌的切片"))
	}
	//check 硬胡，如果手牌上有切片中的牌，那么减一，去赖子，然后检查硬胡
	checkCards := make([]byte, 0)
	static.HF_DeepCopy(&checkCards, &handCards)
	//去赖子,根据上面的判断只会有一张呢
	for _, v := range godCards {
		guiIndex, _ := mahlib2.CardToIndex(v)
		if checkCards[guiIndex] == 0 {
			continue
		}
		if checkCards[guiIndex] > 1 {
			//fmt.Println(fmt.Sprintf("目前规则中，甩字胡牌型里面只能有一个赖子（不管有没有还原）"))
			return false, nil
		}
		checkCards[guiIndex] = 0
		break

	}
	//
	return this.isShuaizhiHu(checkCards, eyeMask), nil
	//如果不是最多14章（3*3后就4章判断），然后应该就是11（3*2）、8 （3）这里就麻烦了，判断的对象是checkCtx.CheckCtx.CbCardIndex，这个时候hu_struct就只能做借鉴了
}

/*
CbCardIndex 是未加判断牌的手牌
checkCards 是以判断判为中心凑顺子，需要的牌
20190302
这个方法不行，还是考虑去牌的方式，去掉判断的牌，手上最多13张，去掉赖子，12章，先去一对，然后再去非这张牌的一章，如果能3n，那么就是甩字胡，遍历所有可能性，
去双最多4次，最少一次
去单最少4次，最多10
只要有一次成功，那么就是甩字胡
*/
func (this *CheckHu) isShuaizhiHu(CbCardIndex []byte, mask258 byte) bool {
	var eyemask byte = 1
	isShuaiZhi := false
	checknum := 0
	for index, v := range CbCardIndex {
		if v > 1 {
			skipIndex := index
			CbCardIndex[index] -= 2
			for index1, v1 := range CbCardIndex {
				if index1 == skipIndex {
					continue
				}
				if v1 != 0 {
					CbCardIndex[index1] -= 1
					//到这里检查的符合3n
					isShuaiZhi, _ = this.CheckHU.Split_Byte(CbCardIndex, 0, eyemask, false)
					checknum++
					CbCardIndex[index1] += 1
					if isShuaiZhi {
						break
					}
				}
			}
			CbCardIndex[index] += 2
			if isShuaiZhi {
				break
			}
		}
	}
	fmt.Println(fmt.Sprintf("检查次数（%d）", checknum))
	_ = checknum
	return isShuaiZhi
}

//-------------------4对胡-------------硬胡-----------
//条件判断
//这个写死
func canFourDui(cbCardIndex []byte) bool {
	//成对数大于4即可
	fourDui := 0
	var cardNum byte = 0
	for i := 0; i < static.MAX_INDEX; i++ {
		cardNum += cbCardIndex[i]
		switch {
		case cbCardIndex[i]/4 == 1:
			fourDui += 2
		case cbCardIndex[i]/2 != 0:
			fourDui += 1
		}
	}
	//只能4对
	if fourDui > 3 && cardNum > 10 {
		return true
	}
	return false
}

//指定去掉几对
func (this *CheckHu) recursionDui(handCards []byte, eyeMask byte, num *int, index int, loseItem *[]int, loseDuiNum int) bool {
	//canhu:=false
	for i := index; i < len(handCards); i++ {
		if handCards[i] > 1 {
			handCards[i] -= 2
			*num += 1
			*loseItem = append(*loseItem, i)
			if *num == loseDuiNum {
				isFourDui, _ := this.CheckHU.Split_Byte(handCards, 0, eyeMask, true)
				//打印完了还原，i走下一
				if isFourDui {
					mahlib2.Print_cards(handCards[:])
					return isFourDui
				}
				*loseItem = (*loseItem)[:len(*loseItem)-1]
				handCards[i] += 2
				*num -= 1
			} else {
				if !this.recursionDui(handCards, eyeMask, num, i+1, loseItem, loseDuiNum) {
					//这一条打印完了，我们继续打印下一个
					loseItemNum := len(*loseItem)
					if loseItemNum != 0 {
						index := (*loseItem)[loseItemNum-1]
						*loseItem = (*loseItem)[:len(*loseItem)-1]
						handCards[index] += 2
						*num -= 1
						i = index
					}
				} else {
					return true
				}

			}
		}
	}
	return false
	//检查是不是可以胡
}

//4对，入口去掉一个，判胡留一个，实际上是去两个对，这样少递归
func (this *CheckHu) CheckFourDuiHu(handCards []byte, eyeMask byte) (result bool, err error) {
	if canFourDui(handCards) {
		//return this.getAllFourDuiItem(handCards,eyeMask)
		num := 0
		//记录下标
		lostItem := []int{}
		for i := 0; i < len(handCards); i++ {
			if handCards[i] > 1 {
				handCards[i] -= 2
				if this.recursionDui(handCards, eyeMask, &num, 0, &lostItem, 2) {
					fmt.Println("找到了")
					lostItem = append(lostItem, i)
					mahlib2.Print_cards(handCards[:])
					return true, nil
				}
				handCards[i] += 2
			}
		}
	}
	return false, nil
}

//-------------------------------------------------------------------------------
////能不能19对
//要返回一些东西
func canOneNineWeave(handCards []byte) (mask byte, num byte) {
	oneNum := handCards[0] + handCards[9] + handCards[18]
	nineNum := handCards[8] + handCards[17] + handCards[26]
	//第一步 必须大于0
	mask = 0
	num = 0
	//至少两个和要大于3
	if (oneNum == 0 || nineNum == 0) || (oneNum+nineNum)/3 == 0 {
		fmt.Println(fmt.Sprintf("没发现可以19对"))
		return
	}
	//先排等的情况 2 3 4 5 6 最多是6个1和6个9
	//这个比较坑啊，因为两个都可以作为最小的
	if oneNum == nineNum {
		//如果两个个数相等，那么必然是1,2这两种情况了
		return 3, oneNum / 2
	}
	//以下是以小为准
	if oneNum > nineNum {
		if oneNum >= nineNum*2 {
			return 1, nineNum
		} else {
			return 1, oneNum / 2
		}
	} else {
		if nineNum >= oneNum*2 {
			return 2, oneNum
		} else {
			return 2, nineNum / 2
		}
	}
	return
}

////去19对
func (this *CheckHu) CheckOneNineDuiHu(handCards []byte, eyeMask byte) (result bool, err error) {
	//初始判断，不检查eye
	mask, num := canOneNineWeave(handCards)
	_ = num
	//func recursionOneNineWeave_view1(checkCards []byte,eyeMask byte,num *int,baseindex int,createItem *[]int,loseItem *[][]int,loseDuiNum int)(bool){
	//num为可创建的最大数
	//这里要生成所有的才行，
	switch mask {
	case 3:
		//19的个数相等，或者1或者9
	case 2:
		//以1为基础
		//createNum:=0
		//		lostItem:= []int{}
		//recursionOneNineWeave_view1(handCards,1,&createNum,8,&lostItem,createNum)

	case 1:
		//以9为基础
	default:
		return false, nil
	}
	return false, nil
}
func Check13YAO(handCards []byte, godCards []byte, checkcard byte, ischihu bool) (result byte, err error) {
	//先检查下硬胡
	if result, err = Check13YAO_Normal(handCards, nil, checkcard, true); result == 0 {
		return
	}
	isNormal := true
	if mahlib2.Findcard(godCards, checkcard) && !ischihu {
		isNormal = false
	}
	return Check13YAO_Normal(handCards, godCards, checkcard, isNormal)
}

//13幺
func Check13YAO_Normal(handCards []byte, godCards []byte, checkcard byte, isNormal bool) (result byte, err error) {
	_, err = mahlib2.CheckHandCardsSafe_ex(handCards, checkcard)
	if err != nil {
		return static.INVALID_BYTE, err
	}
	checkCards := make([]byte, 0)
	static.HF_DeepCopy(&checkCards, &handCards)
	var gui_1 byte = 0xff
	var gui_2 byte = 0xff
	godNum := len(godCards)
	switch godNum {
	case 1:
		gui_1 = godCards[0]
	case 2:
		gui_1 = godCards[0]
		gui_2 = godCards[1]
	}
	var gui_num byte = 0
	gui1index, _ := mahlib2.CardToIndex(gui_1)
	if gui1index < MAX_INDEX {
		gui_num += checkCards[gui1index] % 0xff
		checkCards[gui1index] = 0
	}
	gui2index, _ := mahlib2.CardToIndex(gui_2)
	if gui2index < MAX_INDEX {
		gui_num += checkCards[gui2index] % 0xff
		checkCards[gui2index] = 0
	}
	index, err := mahlib2.CardToIndex(checkcard)
	if index != static.INVALID_BYTE {
		if isNormal {
			checkCards[index]++
		} else {
			gui_num += 1
		}
	}
	//----------------------------
	return Check13YAO_base(checkCards, gui_num), err
}

//最后就像7对一样
func Check13YAO_base(handCards []byte, godNum byte) (result byte) {
	//检查的位置是死的
	checkIndex := []byte{0, 8, 9, 17, 18, 26, 27, 28, 29, 30, 31, 32, 33}
	var needgod byte = 0
	eyeNum := 0
	cardnum := 0
	for _, v := range checkIndex {
		switch handCards[v] {
		case 0:
			needgod++
			if needgod > godNum {
				return static.INVALID_BYTE
			}
		case 2:
			eyeNum++
			if eyeNum > 1 {
				return static.INVALID_BYTE
			}
		case 3, 4:
			return static.INVALID_BYTE
		default:
			cardnum += 1
		}
	}
	//几种情况
	if cardnum == 12 && eyeNum == 1 {
		return 0
	}
	if eyeNum == 0 {
		needgod++
	}
	if needgod <= godNum {
		return needgod
	}
	return static.INVALID_BYTE
}

////4对，入口去掉一个，判胡留一个，实际上是去两个对，这样少递归
//func (this *CheckHu) CreateOneNineWeave_all(handCards []byte,baseindex int,eyeMask byte,loseItemNum int) (result bool,err error) {
//		//记录下标
//		lostItem:= []int{}
//		num:=0
//	for i:=baseindex;i<27;i+=8 {
//		num=0
//		lostItem:= []int{}
//		if handCards[i] > 0 {
//			handCards[i]-=1
//			if handCards[i]>0{
//				//func recursionOneNineWeave_view(checkCards []byte,eyeMask byte,num *int,baseindex int,loseItem *[]int,loseDuiNum int)
//				recursionOneNineWeave_view(handCards,eyeMask,&num,i,&lostItem,loseItemNum)
//			}else{
//
//			}
//
//
//		}
//	}
//		for i:=0;i<len(handCards);i++{
//			if handCards[i] > 1 {
//				handCards[i]-=2
//				if this.recursionDui(handCards,eyeMask,&num,0,&lostItem,2){
//					fmt.Println("找到了")
//					lostItem=append(lostItem,i)
//					backboard.Print_cards(handCards[:])
//					return true,nil
//				}
//				handCards[i]+=2
//			}
//		}
//
//	return false,nil
//}
////拿出所有需要成对数据的切片  显示用
//func creatOneNineWeave_view(handCards []byte,mask byte,num byte) {
//	//生成检查切片
//	switch mask {
//	case 3:
//		//19的个数相等，或者1或者9
//		//oneNum:=handCards[0]+handCards[9]+handCards[17]
//		//nineNum:=handCards[8]+handCards[17]+handCards[26]
//	case 2:
//		//以1为基础，生成9能成对数据的切片
//
//		//nineNum:=handCards[8]+handCards[17]+handCards[26]
//		createNum:=0
//		lostItem:= []int{}
//		recursionOneNineWeave_view(handCards,1,&createNum,8,&lostItem,createNum)
//	case 1:
//		//以9为基础
//		//oneNum:=handCards[0]+handCards[9]+handCards[17]
//		createNum:=0
//		lostItem:= []int{}
//		recursionOneNineWeave_view(handCards,1,&createNum,0,&lostItem,createNum)
//	}
//}
//func creatOneNineWeave_view1(handCards []byte,mask byte,num byte) {
////生成检查切片
//switch mask {
//case 3:
////19的个数相等，或者1或者9
////oneNum:=handCards[0]+handCards[9]+handCards[17]
////nineNum:=handCards[8]+handCards[17]+handCards[26]
//case 2:
////以1为基础，生成9能成对数据的切片
////nineNum:=handCards[8]+handCards[17]+handCards[26]
//createNum:=0
////创建map吧 这里坑的是如果有多个，最多4个，创建4个数组？map？
//
//lostItem:= []int{}
//recursionOneNineWeave_view1(handCards,1,&createNum,8,&lostItem,createNum)
//case 1:
////以9为基础
////oneNum:=handCards[0]+handCards[9]+handCards[17]
//createNum:=0
//lostItem:= []int{}
//recursionOneNineWeave_view1(handCards,1,&createNum,0,&lostItem,createNum)
//}
//}
//func removeDuiCards(mahjongMatrix []int,maxNum int) []int {
//	marks := make([]int, 0, maxNum)
//	for i := 0; i < len(mahjongMatrix); i++ {
//		if mahjongMatrix[i] >= 1 {
//			mahjongMatrix[i] -= 1
//			//创建东西
//			marks = append(marks, ((i+1)*10 + 3))
//		}
//	}
//	return marks
//}
func recursionOneNineWeave_view1(checkCards []byte, eyeMask byte, num *int, baseindex int, createItem *[]int, loseItem *[][]int, loseDuiNum int) bool {
	if createItem == nil {
		//createItem=new([3]int)
		createItem = &[]int{}
	}
	for i := baseindex; i < 27; i += 8 {
		if checkCards[i] > 0 {
			checkCards[i] -= 1
			if checkCards[i] == 0 {
				baseindex += 8
			}
			*createItem = append(*createItem, i)
			findItem := false
			switch len(*createItem) {
			case 1:
				//新建的，还不够进去在搞
				findItem = recursionOneNineWeave_view1(checkCards, eyeMask, num, baseindex, createItem, loseItem, loseDuiNum)
				_ = findItem
			case 2:
				//加到map里面去
				*loseItem = append(*loseItem, *createItem)
				*num += 1
				if *num == loseDuiNum {
					fmt.Println(fmt.Sprintf("凑够了"))
					mahlib2.Print_cards(checkCards[:])
					*loseItem = (*loseItem)[:len(*loseItem)-1]
					checkCards[i] += 1
					*num -= 1
				} else {
					//追加
					if recursionOneNineWeave_view1(checkCards, eyeMask, num, baseindex, nil, loseItem, loseDuiNum) {
						loseItemNum := len(*loseItem)
						if loseItemNum != 0 {
							item := (*loseItem)[loseItemNum-1]
							*loseItem = (*loseItem)[:len(*loseItem)-1]
							checkCards[item[len(item)-1]] += 1
							*num -= 1
							i = item[len(item)-1]
						}
					} else {
						return true
					}

				}
			default:
				panic("出错了")
			}
		}
	}
	return false
}

////这个特殊一点，因为
///*
//如果以1为基准 起始下标是8，终止是26 3次
//以9的话，那么就是0.18
//补偿8
// */
// //func (this *CheckHu)recursionDui(handCards []byte, eyeMask byte,num *int,index int,loseItem *[]int,loseDuiNum int)(bool)  {
//	func recursionOneNineWeave_view(checkCards []byte,eyeMask byte,num *int,baseindex int,loseItem *[]int,loseDuiNum int)(bool){
//		var newItem [3]int
//		for i:=baseindex;i<27;i+=8 {
//			if checkCards[i] > 0 {
//				checkCards[i] -= 1
//				*num += 1
//				newItem[0]=i
//				*loseItem=append(*loseItem,i)
//				if *num==loseDuiNum{
//					//isFourDui, _:= this.CheckHU.Split_Byte(handCards, 0, eyeMask, true)
//					//打印完了还原，i走下一
//					//if isFourDui {
//					//	common.Print_cards(handCards[:])
//					//	return isFourDui
//					//}
//					common.Print_cards(checkCards[:])
//					*loseItem=(*loseItem)[:len(*loseItem)-1]
//					checkCards[i]+= 1
//					*num-=1
//				}else{
//					if recursionOneNineWeave_view(checkCards,eyeMask,num,i+8,loseItem,loseDuiNum){
//					//if !this.recursionDui(checkCards,eyeMask,num,i+1,loseItem,loseDuiNum){
//						//这一条打印完了，我们继续打印下一个
//						loseItemNum:=len(*loseItem)
//						if loseItemNum!=0 {
//							index := (*loseItem)[loseItemNum-1]
//							*loseItem=(*loseItem)[:len(*loseItem)-1]
//							checkCards[index]+=1
//							*num -= 1
//							i=index
//						}
//					}else{
//						return true
//					}
//
//				}
//			}
//		}
//		return false
//}

//-------------------------------------------------------------------------

func PrintSpecial(waveType int, cbCardIndex1 []byte, cbCardIndex2 []byte) {
	logStr1 := ""
	logStr2 := ""
	if waveType == 1 {
		logStr1 = "四对牌型:"
		logStr2 = "四对剩余牌型:"
	} else if waveType == 2 {
		logStr1 = "19牌型:"
		logStr2 = "19剩余牌型:"
	} else if waveType == 3 {
		logStr1 = "特殊风牌型:"
		logStr2 = "特殊风剩余牌型:"
	} else if waveType == 4 {
		logStr1 = "特殊箭牌型:"
		logStr2 = "特殊箭剩余牌型:"
	}
	for i := 0; i < len(cbCardIndex1); i++ {
		//result:= common.IndexToCard(cbCardIndex1[i])
		logStr1 += fmt.Sprintf("%s", mahlib2.G_CardAnother[cbCardIndex1[i]])
	}
	for i := 0; i < len(cbCardIndex2); i++ {
		for j := 0; j < int(cbCardIndex2[i]); j++ {
			logStr2 += fmt.Sprintf("%s", mahlib2.G_CardAnother[byte(i)])
		}
	}
	fmt.Println(fmt.Sprintf("%v", logStr1))
	fmt.Println(fmt.Sprintf("%v", logStr2))
}

func PrintSpecial2(waveType int, cbCardIndex1 []byte) {
	logStr1 := "剩余牌"
	if waveType == 1 {
		logStr1 = "四对牌型:"
	} else if waveType == 2 {
		logStr1 = "19牌型:"
	} else if waveType == 3 {
		logStr1 = "特殊风牌型:"
	} else if waveType == 4 {
		logStr1 = "特殊箭牌型:"
	}
	if waveType != 0 {
		for i := 0; i < len(cbCardIndex1); i++ {
			logStr1 += fmt.Sprintf("%s", mahlib2.G_CardAnother[cbCardIndex1[i]])
		}
	} else {
		for i := 0; i < len(cbCardIndex1); i++ {
			for j := 0; j < int(cbCardIndex1[i]); j++ {
				logStr1 += fmt.Sprintf("%s", mahlib2.G_CardAnother[byte(i)])
			}
		}
	}
	fmt.Println(fmt.Sprintf("%v", logStr1))
}

func CheckHandCardEnough(cbCardIndex *[]byte, waveItem []byte) bool {
	for i := 0; i < len(waveItem); i++ {
		if (*cbCardIndex)[waveItem[i]] > 0 {
			(*cbCardIndex)[waveItem[i]]--
		} else {
			return false
		}
	}
	return true
}

func AppEndToCheckCards(cardIndex *map[string][]byte, addCardIndex []byte, cardIndex_left *map[string][]byte, addCardIndex_left []byte, key string) {
	(*cardIndex)[key] = addCardIndex
	(*cardIndex_left)[key] = addCardIndex_left
}

//找到所有的19组合
func FindAllSpecial19WaveItems(cbCardIndex []byte) (map[string][]byte, int) {
	//找到所有的1,9
	var oneCount int = 0
	var nineCount int = 0
	var allonenine []byte
	for i := 0; i < int(cbCardIndex[0]); i++ {
		allonenine = append(allonenine, 0)
		oneCount++
	}
	for i := 0; i < int(cbCardIndex[9]); i++ {
		allonenine = append(allonenine, 9)
		oneCount++
	}
	for i := 0; i < int(cbCardIndex[18]); i++ {
		allonenine = append(allonenine, 18)
		oneCount++
	}
	for i := 0; i < int(cbCardIndex[8]); i++ {
		allonenine = append(allonenine, 8)
		nineCount++
	}
	for i := 0; i < int(cbCardIndex[17]); i++ {
		allonenine = append(allonenine, 17)
		nineCount++
	}
	for i := 0; i < int(cbCardIndex[26]); i++ {
		allonenine = append(allonenine, 26)
		nineCount++
	}

	waveItems := map[string][]byte{}

	if len(allonenine) < 3 {
		return waveItems, 0
	}

	for i := 0; i < len(allonenine); i++ {
		for j := i + 1; j < len(allonenine); j++ {
			for k := j + 1; k < len(allonenine); k++ {
				cardValuei := mahlib2.IndexToCard(allonenine[i]) & static.MASK_VALUE
				cardValuej := mahlib2.IndexToCard(allonenine[j]) & static.MASK_VALUE
				cardValuek := mahlib2.IndexToCard(allonenine[k]) & static.MASK_VALUE

				if cardValuei != cardValuej || cardValuei != cardValuek || cardValuej != cardValuek {
					tmpWaveItem := []byte{allonenine[i], allonenine[j], allonenine[k]}
					waveItems[static.HF_Bytestoa(tmpWaveItem)] = tmpWaveItem
				}
			}
		}
	}

	minOneNineCount := oneCount
	if nineCount < minOneNineCount {
		minOneNineCount = nineCount
	}
	maxCount := len(allonenine) / 3
	if minOneNineCount < maxCount {
		maxCount = minOneNineCount
	}

	return waveItems, maxCount
}

//找到完全组合的19
func IsAllUsedSpecial19WaveItems(cbCardIndex []byte, useCount int) (bool, []byte) {
	//找到所有的1,9
	var oneCount int = 0
	var nineCount int = 0
	var allonenine []byte
	for i := 0; i < int(cbCardIndex[0]); i++ {
		allonenine = append(allonenine, 0)
		oneCount++
	}
	for i := 0; i < int(cbCardIndex[9]); i++ {
		allonenine = append(allonenine, 9)
		oneCount++
	}
	for i := 0; i < int(cbCardIndex[18]); i++ {
		allonenine = append(allonenine, 18)
		oneCount++
	}
	for i := 0; i < int(cbCardIndex[8]); i++ {
		allonenine = append(allonenine, 8)
		nineCount++
	}
	for i := 0; i < int(cbCardIndex[17]); i++ {
		allonenine = append(allonenine, 17)
		nineCount++
	}
	for i := 0; i < int(cbCardIndex[26]); i++ {
		allonenine = append(allonenine, 26)
		nineCount++
	}

	if len(allonenine) != useCount*3 {
		return false, []byte{}
	}

	if oneCount < useCount {
		return false, []byte{}
	}

	if nineCount < useCount {
		return false, []byte{}
	}

	return true, allonenine
}

//找到所有19组合以外的牌
func FindHandCardRemove19WaveItems(cbCardIndex []byte, mask19 bool) (map[string][]byte, map[string][]byte) {
	cardIndex_19 := map[string][]byte{}
	cardIndex_19_left := map[string][]byte{}
	if mask19 {
		special19WaveItems, maxCount := FindAllSpecial19WaveItems(cbCardIndex)

		for _, v1 := range special19WaveItems {
			tmp19wave := []byte{}
			tmp19wave = append(tmp19wave, v1...)

			cbCardIndexTemp := make([]byte, static.MAX_INDEX)
			copy(cbCardIndexTemp, cbCardIndex[:])

			if !CheckHandCardEnough(&cbCardIndexTemp, tmp19wave) {
				continue
			}

			key := static.HF_Bytestoa(cbCardIndexTemp)
			if cardIndex_19[key] != nil {
				continue
			}

			AppEndToCheckCards(&cardIndex_19, tmp19wave, &cardIndex_19_left, cbCardIndexTemp, key)
		}

		if maxCount >= 2 {
			for _, v1 := range special19WaveItems {
				for _, v2 := range special19WaveItems {
					cbCardIndexTemp := make([]byte, static.MAX_INDEX)
					copy(cbCardIndexTemp, cbCardIndex[:])
					tmp19wave := []byte{}
					tmp19wave = append(tmp19wave, v1...)
					tmp19wave = append(tmp19wave, v2...)

					if !CheckHandCardEnough(&cbCardIndexTemp, tmp19wave) {
						continue
					}

					key := static.HF_Bytestoa(cbCardIndexTemp)
					if cardIndex_19[key] != nil {
						continue
					}
					AppEndToCheckCards(&cardIndex_19, tmp19wave, &cardIndex_19_left, cbCardIndexTemp, key)
				}
			}
		}

		if maxCount >= 3 {
			for _, v1 := range special19WaveItems {
				for _, v2 := range special19WaveItems {
					for _, v3 := range special19WaveItems {
						cbCardIndexTemp := make([]byte, static.MAX_INDEX)
						copy(cbCardIndexTemp, cbCardIndex[:])
						tmp19wave := []byte{}
						tmp19wave = append(tmp19wave, v1...)
						tmp19wave = append(tmp19wave, v2...)
						tmp19wave = append(tmp19wave, v3...)

						if !CheckHandCardEnough(&cbCardIndexTemp, tmp19wave) {
							continue
						}

						key := static.HF_Bytestoa(cbCardIndexTemp)
						if cardIndex_19[key] != nil {
							continue
						}
						AppEndToCheckCards(&cardIndex_19, tmp19wave, &cardIndex_19_left, cbCardIndexTemp, key)
					}
				}
			}
		}

		if maxCount >= 4 {
			for i := 0; i < static.MAX_INDEX; i++ {
				cbCardIndexTemp := make([]byte, static.MAX_INDEX)
				copy(cbCardIndexTemp, cbCardIndex[:])

				if cbCardIndexTemp[i] >= 2 {
					cbCardIndexTemp[i] -= 2

					allUsed, tmp19wave := IsAllUsedSpecial19WaveItems(cbCardIndexTemp, 4)
					if allUsed {
						if !CheckHandCardEnough(&cbCardIndexTemp, tmp19wave) {
							continue
						}
						key := static.HF_Bytestoa(cbCardIndexTemp)
						if cardIndex_19[key] != nil {
							continue
						}
						AppEndToCheckCards(&cardIndex_19, tmp19wave, &cardIndex_19_left, cbCardIndexTemp, key)
					}
				}
			}
		}
	}

	//包含没有19牌方便计算
	cbCardIndexTemp := make([]byte, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex[:])

	cardIndex_19[static.HF_Bytestoa(cbCardIndexTemp)] = []byte{}
	cardIndex_19_left[static.HF_Bytestoa(cbCardIndexTemp)] = cbCardIndexTemp

	return cardIndex_19_left, cardIndex_19
}

//找到所有的四对组合
func FindAllSpecial4DuiWaveItems(cbCardIndex []byte) map[string][]byte {
	alldui := []byte{}
	for i := 0; i < len(cbCardIndex); i++ {
		if cbCardIndex[i] == 4 {
			alldui = append(alldui, byte(i))
			alldui = append(alldui, byte(i))
		} else if cbCardIndex[i] >= 2 {
			alldui = append(alldui, byte(i))
		}
	}
	waveItems := map[string][]byte{}
	if len(alldui) < 4 {
		return waveItems
	}
	for i := 0; i < len(alldui); i++ {
		for j := i + 1; j < len(alldui); j++ {
			for k := j + 1; k < len(alldui); k++ {
				for m := k + 1; m < len(alldui); m++ {
					tmpWaveItem := []byte{alldui[i], alldui[j], alldui[k], alldui[m]}
					waveItems[static.HF_Bytestoa(tmpWaveItem)] = tmpWaveItem
				}
			}
		}
	}
	return waveItems
}

//找到所有四对组合以外的牌
func FindHandCardRemove4DuiWaveItems(cbCardIndex []byte) (map[string][]byte, map[string][]byte) {
	special4duiWaveItems := FindAllSpecial4DuiWaveItems(cbCardIndex)
	cardIndex_4dui := map[string][]byte{}
	cardIndex_4dui_left := map[string][]byte{}

	for _, v := range special4duiWaveItems {
		cbCardIndexTemp := make([]byte, static.MAX_INDEX)
		copy(cbCardIndexTemp, cbCardIndex[:])

		tmpwave := []byte{}
		tmpwave = append(tmpwave, v...)
		tmpwave = append(tmpwave, v...)

		if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
			continue
		}

		key := static.HF_Bytestoa(cbCardIndexTemp)
		if cardIndex_4dui[key] != nil {
			continue
		}

		AppEndToCheckCards(&cardIndex_4dui, v, &cardIndex_4dui_left, cbCardIndexTemp, key)
	}

	//包含没有四对方便计算
	cbCardIndexTemp := make([]byte, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex[:])
	cardIndex_4dui[static.HF_Bytestoa(cbCardIndexTemp)] = []byte{}
	cardIndex_4dui_left[static.HF_Bytestoa(cbCardIndexTemp)] = cbCardIndexTemp

	return cardIndex_4dui_left, cardIndex_4dui
}

//找到所有风组合以外的牌
func FindAllSpecialFengWaveItems(cbCardIndex []byte) (map[string][]byte, int) {
	//找到所有的1,9
	var allfeng []byte
	for i := 0; i < int(cbCardIndex[27]); i++ {
		allfeng = append(allfeng, 27)
	}
	for i := 0; i < int(cbCardIndex[28]); i++ {
		allfeng = append(allfeng, 28)
	}
	for i := 0; i < int(cbCardIndex[29]); i++ {
		allfeng = append(allfeng, 29)
	}
	for i := 0; i < int(cbCardIndex[30]); i++ {
		allfeng = append(allfeng, 30)
	}
	waveItems := map[string][]byte{}

	if len(allfeng) < 3 {
		return waveItems, 0
	}

	for i := 0; i < len(allfeng); i++ {
		for j := i + 1; j < len(allfeng); j++ {
			for k := j + 1; k < len(allfeng); k++ {
				if i == j || j == k || i == k {
					continue
				}
				if allfeng[i] != allfeng[j] && allfeng[j] != allfeng[k] && allfeng[k] != allfeng[i] {
					tmpWaveItem := []byte{allfeng[i], allfeng[j], allfeng[k]}
					waveItems[static.HF_Bytestoa(tmpWaveItem)] = tmpWaveItem
				}
			}
		}
	}
	return waveItems, len(allfeng) / 3
}

//找到完全组合的风
func IsAllUsedSpecialFengWaveItems(cbCardIndex []byte, useCount int) (bool, []byte) {
	//找到所有的风
	var allfeng []byte
	var fengCount = [4]int{0, 0, 0, 0}
	for i := 0; i < int(cbCardIndex[27]); i++ {
		allfeng = append(allfeng, 27)
		fengCount[0]++
	}
	for i := 0; i < int(cbCardIndex[28]); i++ {
		allfeng = append(allfeng, 28)
		fengCount[1]++
	}
	for i := 0; i < int(cbCardIndex[29]); i++ {
		allfeng = append(allfeng, 29)
		fengCount[2]++
	}
	for i := 0; i < int(cbCardIndex[30]); i++ {
		allfeng = append(allfeng, 30)
		fengCount[3]++
	}

	if len(allfeng) != useCount*3 {
		return false, []byte{}
	}

	for i := 0; i < 4; i++ {
		if fengCount[i] > useCount {
			return false, []byte{}
		}
	}
	return true, allfeng
}

//找到所有风组合以外的牌
func FindHandCardRemoveFengWaveItems(cbCardIndex []byte, maskfeng bool) (map[string][]byte, map[string][]byte) {
	cardIndex_feng_left := map[string][]byte{}
	cardIndex_feng := map[string][]byte{}

	if maskfeng {
		specialfengWaveItems, maxCount := FindAllSpecialFengWaveItems(cbCardIndex)
		for _, v1 := range specialfengWaveItems {
			cbCardIndexTemp := make([]byte, static.MAX_INDEX)
			copy(cbCardIndexTemp, cbCardIndex[:])

			tmpwave := []byte{}
			tmpwave = append(tmpwave, v1...)

			if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
				continue
			}

			key := static.HF_Bytestoa(cbCardIndexTemp)
			if cardIndex_feng[key] != nil {
				continue
			}

			AppEndToCheckCards(&cardIndex_feng, tmpwave, &cardIndex_feng_left, cbCardIndexTemp, key)
		}

		if maxCount >= 2 {
			for _, v1 := range specialfengWaveItems {
				for _, v2 := range specialfengWaveItems {
					cbCardIndexTemp := make([]byte, static.MAX_INDEX)
					copy(cbCardIndexTemp, cbCardIndex[:])

					tmpwave := []byte{}
					tmpwave = append(tmpwave, v1...)
					tmpwave = append(tmpwave, v2...)

					if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
						continue
					}

					key := static.HF_Bytestoa(cbCardIndexTemp)
					if cardIndex_feng[key] != nil {
						continue
					}

					AppEndToCheckCards(&cardIndex_feng, tmpwave, &cardIndex_feng_left, cbCardIndexTemp, key)
				}
			}
		}

		if maxCount >= 3 {
			for _, v1 := range specialfengWaveItems {
				for _, v2 := range specialfengWaveItems {
					for _, v3 := range specialfengWaveItems {
						cbCardIndexTemp := make([]byte, static.MAX_INDEX)
						copy(cbCardIndexTemp, cbCardIndex[:])

						tmpwave := []byte{}
						tmpwave = append(tmpwave, v1...)
						tmpwave = append(tmpwave, v2...)
						tmpwave = append(tmpwave, v3...)

						if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
							continue
						}

						key := static.HF_Bytestoa(cbCardIndexTemp)
						if cardIndex_feng[key] != nil {
							continue
						}

						AppEndToCheckCards(&cardIndex_feng, tmpwave, &cardIndex_feng_left, cbCardIndexTemp, key)
					}
				}
			}
		}

		if maxCount >= 4 {
			for i := 0; i < static.MAX_INDEX; i++ {
				cbCardIndexTemp := make([]byte, static.MAX_INDEX)
				copy(cbCardIndexTemp, cbCardIndex[:])

				if cbCardIndexTemp[i] >= 2 {
					cbCardIndexTemp[i] -= 2

					allUsed, tmp19wave := IsAllUsedSpecialFengWaveItems(cbCardIndexTemp, 4)
					if allUsed {
						if !CheckHandCardEnough(&cbCardIndexTemp, tmp19wave) {
							continue
						}
						key := static.HF_Bytestoa(cbCardIndexTemp)
						if cardIndex_feng[key] != nil {
							continue
						}
						AppEndToCheckCards(&cardIndex_feng, tmp19wave, &cardIndex_feng_left, cbCardIndexTemp, key)
					}
				}
			}
		}
	}
	//包含没有风牌方便计算
	cbCardIndexTemp := make([]byte, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex[:])
	cardIndex_feng[static.HF_Bytestoa(cbCardIndexTemp)] = []byte{}
	cardIndex_feng_left[static.HF_Bytestoa(cbCardIndexTemp)] = cbCardIndexTemp
	return cardIndex_feng_left, cardIndex_feng
}

//找到所有红中白板发财
func FindAllSpecialDRAGONWaveItems(cbCardIndex []byte) (map[string][]byte, int) {
	//找到所有的1,9
	var allfeng []byte
	for i := 0; i < int(cbCardIndex[31]); i++ {
		allfeng = append(allfeng, 31)
	}
	for i := 0; i < int(cbCardIndex[32]); i++ {
		allfeng = append(allfeng, 32)
	}
	for i := 0; i < int(cbCardIndex[33]); i++ {
		allfeng = append(allfeng, 33)
	}
	waveItems := map[string][]byte{}

	if len(allfeng) < 3 {
		return waveItems, 0
	}
	for i := 0; i < len(allfeng); i++ {
		for j := i + 1; j < len(allfeng); j++ {
			for k := j + 1; k < len(allfeng); k++ {
				if i == j || j == k || i == k {
					continue
				}
				if allfeng[i] != allfeng[j] && allfeng[j] != allfeng[k] && allfeng[k] != allfeng[i] {
					tmpWaveItem := []byte{allfeng[i], allfeng[j], allfeng[k]}
					waveItems[static.HF_Bytestoa(tmpWaveItem)] = tmpWaveItem
				}
			}
		}
	}
	return waveItems, len(allfeng) / 3
}

//找到完全组合的风
func IsAllUsedSpecialDRAGONWaveItems(cbCardIndex []byte, useCount int) (bool, []byte) {
	//找到所有的风
	var allfeng []byte
	var fengCount = [3]int{0, 0, 0}
	for i := 0; i < int(cbCardIndex[31]); i++ {
		allfeng = append(allfeng, 31)
		fengCount[0]++
	}
	for i := 0; i < int(cbCardIndex[32]); i++ {
		allfeng = append(allfeng, 32)
		fengCount[1]++
	}
	for i := 0; i < int(cbCardIndex[33]); i++ {
		allfeng = append(allfeng, 33)
		fengCount[2]++
	}

	if len(allfeng) != useCount*3 {
		return false, []byte{}
	}

	for i := 0; i < 4; i++ {
		if fengCount[i] != useCount {
			return false, []byte{}
		}
	}
	return true, allfeng
}

//红中白板发财
func FindHandCardRemoveDRAGONWaveItems(cbCardIndex []byte, maskfeng bool) (map[string][]byte, map[string][]byte) {
	cardIndex_feng_left := map[string][]byte{}
	cardIndex_feng := map[string][]byte{}
	if maskfeng {
		specialfengWaveItems, maxCount := FindAllSpecialDRAGONWaveItems(cbCardIndex)

		for _, v1 := range specialfengWaveItems {
			cbCardIndexTemp := make([]byte, static.MAX_INDEX)
			copy(cbCardIndexTemp, cbCardIndex[:])

			tmpwave := []byte{}
			tmpwave = append(tmpwave, v1...)

			if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
				continue
			}

			key := static.HF_Bytestoa(cbCardIndexTemp)
			if cardIndex_feng[key] != nil {
				continue
			}

			AppEndToCheckCards(&cardIndex_feng, tmpwave, &cardIndex_feng_left, cbCardIndexTemp, key)
		}

		if maxCount >= 2 {
			for _, v1 := range specialfengWaveItems {
				for _, v2 := range specialfengWaveItems {
					cbCardIndexTemp := make([]byte, static.MAX_INDEX)
					copy(cbCardIndexTemp, cbCardIndex[:])

					tmpwave := []byte{}
					tmpwave = append(tmpwave, v1...)
					tmpwave = append(tmpwave, v2...)

					if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
						continue
					}

					key := static.HF_Bytestoa(cbCardIndexTemp)
					if cardIndex_feng[key] != nil {
						continue
					}

					AppEndToCheckCards(&cardIndex_feng, tmpwave, &cardIndex_feng_left, cbCardIndexTemp, key)
				}
			}
		}

		if maxCount >= 3 {
			for _, v1 := range specialfengWaveItems {
				for _, v2 := range specialfengWaveItems {
					for _, v3 := range specialfengWaveItems {
						cbCardIndexTemp := make([]byte, static.MAX_INDEX)
						copy(cbCardIndexTemp, cbCardIndex[:])

						tmpwave := []byte{}
						tmpwave = append(tmpwave, v1...)
						tmpwave = append(tmpwave, v2...)
						tmpwave = append(tmpwave, v3...)

						if !CheckHandCardEnough(&cbCardIndexTemp, tmpwave) {
							continue
						}

						key := static.HF_Bytestoa(cbCardIndexTemp)
						if cardIndex_feng[key] != nil {
							continue
						}

						AppEndToCheckCards(&cardIndex_feng, tmpwave, &cardIndex_feng_left, cbCardIndexTemp, key)
					}
				}
			}
		}

		if maxCount >= 4 {
			for i := 0; i < static.MAX_INDEX; i++ {
				cbCardIndexTemp := make([]byte, static.MAX_INDEX)
				copy(cbCardIndexTemp, cbCardIndex[:])

				if cbCardIndexTemp[i] >= 2 {
					cbCardIndexTemp[i] -= 2

					allUsed, tmp19wave := IsAllUsedSpecialDRAGONWaveItems(cbCardIndexTemp, 4)
					if allUsed {
						if !CheckHandCardEnough(&cbCardIndexTemp, tmp19wave) {
							continue
						}
						key := static.HF_Bytestoa(cbCardIndexTemp)
						if cardIndex_feng[key] != nil {
							continue
						}
						AppEndToCheckCards(&cardIndex_feng, tmp19wave, &cardIndex_feng_left, cbCardIndexTemp, key)
					}
				}
			}
		}
	}
	//包含没有风牌方便计算
	cbCardIndexTemp := make([]byte, static.MAX_INDEX)
	copy(cbCardIndexTemp, cbCardIndex[:])
	cardIndex_feng[static.HF_Bytestoa(cbCardIndexTemp)] = []byte{}
	cardIndex_feng_left[static.HF_Bytestoa(cbCardIndexTemp)] = cbCardIndexTemp
	return cardIndex_feng_left, cardIndex_feng
}

//19 风刻 箭刻 4对
//record为 4对，19 ，风箭句数量
func (this *CheckHu) AnalyseChiHuCard_Special(cbCardIndex []byte, cbCurrentCard byte, Origin byte, mask19 bool, maskfeng bool, record *[]int) (byte, int) {
	wChiHuKind := byte(static.CHK_NULL)
	//把牌加进去判断
	hardcard := make([]byte, 0)
	static.HF_DeepCopy(&hardcard, &cbCardIndex)
	index, _ := mahlib2.CardToIndex(cbCurrentCard)
	hardcard[index]++
	card_remove4dui_left, card_remove4dui := FindHandCardRemove4DuiWaveItems(hardcard)
	checkInfo := &card_mgr.CardBaseInfo{
		cbCurrentCard, //牌值
		Origin,        //隶属
	}
	checkcardIndex, _ := mahlib2.CardToIndex(cbCurrentCard)
	checknum := 0
	hu := false
	maxfen := 0
	for left4duicardKey, left4duicard := range card_remove4dui_left {
		//打印找到的所有四对组合以及剩余牌
		//PrintSpecial(1, card_remove4dui[left4duicardKey], card_remove4dui_left[left4duicardKey])
		card_remove19_left, card_remove19 := FindHandCardRemove19WaveItems(left4duicard, mask19)
		for left19cardKey, left19card := range card_remove19_left {
			//打印找到的所有19组合以及剩余牌
			//PrintSpecial(2, card_remove19[left19cardKey], card_remove19_left[left19cardKey])
			card_removefengleft, card_removefeng := FindHandCardRemoveFengWaveItems(left19card, maskfeng)
			for leftfengcard_key, leftfengcard := range card_removefengleft {
				//打印找到的所有风组合以及剩余牌
				//PrintSpecial(3, card_removefeng[leftfengcard_key], card_removefengleft[leftfengcard_key])
				card_removeDRAGONleft, card_removeDRAGON := FindHandCardRemoveDRAGONWaveItems(leftfengcard, maskfeng)
				for leftDRAGONcard_key, leftDRAGONcard := range card_removeDRAGONleft {
					//PrintSpecial(4, card_removeDRAGON[leftDRAGONcard_key], card_removeDRAGONleft[leftDRAGONcard_key])
					if len(card_remove4dui[left4duicardKey]) != 0 {
						//fmt.Println("不用带将判胡}}}}}")
						hu, _ = this.CheckHU.Split_Byte(leftDRAGONcard, 0, 0, false)
					} else {
						//fmt.Println("((((((带将判胡")
						hu, _ = this.CheckHU.Split_Byte(leftDRAGONcard, 0, 1, false)
					}
					checknum++
					if hu {
						//写结果，有多少个
						//PrintSpecial2(1, card_remove4dui[left4duicardIndex])
						//PrintSpecial2(2, card_remove19[left19cardIndex])
						//PrintSpecial2(3, card_removefeng[leftfengcardIndex])
						//PrintSpecial2(4, card_removeDRAGON[leftDRAGONcardIndex])
						//PrintSpecial2(0, card_removeDRAGONleft[leftDRAGONcardIndex])
						//fmt.Println(fmt.Sprintf("去四对（%d），去19（%d），去风刻（%d），去箭克（%d）",len(card_remove4dui[left4duicardIndex]),len(card_remove19[left19cardIndex]),len(card_removefeng[leftfengcardIndex]),len(card_removeDRAGON[leftDRAGONcardIndex])))
						fourDuiNum := len(card_remove4dui[left4duicardKey]) / 4
						oneNineNum := len(card_remove19[left19cardKey]) / 3
						fengNum := len(card_removefeng[leftfengcard_key]) / 3
						jianNum := len(card_removeDRAGON[leftDRAGONcard_key]) / 3
						//4对胡，如果胡的牌是4对中的，就不能捉铳
						if Origin != card_mgr.ORIGIN_NOM && fourDuiNum != 0 {
							//检查的牌存在4对中
							findcard := false
							for _, v := range card_remove4dui[left4duicardKey] {
								//fmt.Println(fmt.Sprintf("%v", v))
								if v^checkcardIndex == 0 {
									findcard = true
									break
								}
							}
							//最后剩牌
							if len(leftDRAGONcard) != 0 {
								if leftDRAGONcard[checkcardIndex] != 0 {
									findcard = false
								}
							}
							//如果4对中有判断的牌
							if findcard {
								//检查是不是别的牌型里面包含 获嘉里面如果有的话，那是可以捉铳的
								err, color, value := mahlib2.Cardsplit(cbCurrentCard)
								if err != nil {
									fmt.Println(err)
									break
								}
								if color < 3 {
									if (value == 1 || value == 9) && oneNineNum != 0 {
										//判断19句
										for _, v := range card_remove19[left19cardKey] {
											if v^checkcardIndex == 0 {
												findcard = false
												break
											}
										}
									}
								} else {
									//风箭牌
									if value < 5 {
										if fengNum != 0 {
											for _, v := range card_removefeng[leftfengcard_key] {
												if v^checkcardIndex == 0 {
													findcard = false
													break
												}
											}
										}
									} else if jianNum != 0 { //箭牌里面有没有
										for _, v := range card_removeDRAGON[leftDRAGONcard_key] {
											if v^checkcardIndex == 0 {
												findcard = false
												break
											}
										}
									}
								}
							}
							hu = !findcard
						}
						if hu {
							wChiHuKind |= static.WIK_CHI_HU
							//fmt.Println("--------------------------一轮结束, 胡-----------------------------------")
							tempfen := fourDuiNum + oneNineNum + fengNum + jianNum
							if maxfen < tempfen {
								maxfen = tempfen
								if record != nil {
									(*record)[0] = fourDuiNum
									(*record)[1] = oneNineNum
									(*record)[2] = fengNum + jianNum
								} else {
									//查听直接出去了
									//fmt.Println(fmt.Sprintf("听牌 特殊胡检查次数（%d）次",checknum))
									return wChiHuKind, 0
								}
							}
							if maxfen == 4 {
								//fmt.Println(fmt.Sprintf("特殊胡检查（%d）次,发现最大番数（%d），返回！！！！",checknum,maxfen))
								return wChiHuKind, maxfen
							}
						}
						//}else{
						//	fmt.Println("一轮结束, 没胡, 开始新的一轮")
					}
					//fmt.Println("***************************************************")
				}
			}
		}
	}
	//fmt.Println(fmt.Sprintf("特殊胡检查（%d）次,查胡成功（%t）",checknum,wChiHuKind&public.WIK_CHI_HU!=0))
	if wChiHuKind&static.WIK_CHI_HU != 0 {
		//fmt.Println(fmt.Sprintf("特殊查胡成功！！！！！！！！！！！！！！！！！！！！！！！！！！！！"))
		return wChiHuKind, maxfen
	}
	//func (this *HuLib) AnalyseHuRestore(cbCardIndex []byte, checkInfo *card_mgr.CardBaseInfo, eyeMask byte)
	//普通胡
	hu, _ = this.CheckHU.AnalyseHuRestore(cbCardIndex, checkInfo, 1)
	if hu {
		//fmt.Println("！！！！！！！！！！！！！！！！！！！！！！！！普通3n+2可胡")
		wChiHuKind |= static.WIK_CHI_HU
	} else {
		//fmt.Println("不能普通胡")
	}
	return wChiHuKind, 0
}
