package lib_win

import (
	"errors"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
)

/*
姐妹铺
预想包含双姐妹铺
去姐妹铺的条件是2,2,2
即是顺子，又要是2个，但是赖子目前就只有4个可用
*/
type JieMeiPuObj struct {
	handCards  []byte //去掉姐妹铺的数据
	index      byte   //姐妹铺的起始下标
	lostCards  []byte //姐妹铺去牌数据 如果还原的话，用这个参考加回去
	lastGodNum byte   //剩下鬼牌
}

/*
最多2组
*/
type JieMeiPuResult struct {
	handCards  []byte    //去掉姐妹铺的数据
	index      [2]byte   //姐妹铺的起始下标
	lostCards  [2][]byte //姐妹铺去牌数据 如果还原的话，用这个参考加回去
	lastGodNum byte      //剩下鬼牌
	isDouble   bool      //是不是双姐妹铺
	huKind     byte      //硬胡 软胡 258标志
}

func (this *CheckHu) CreateJieMeiPu_base_ex(handCards []byte, godNum *byte, index byte, lostCards *[]byte) (result byte) {
	//先看牌够不够
	var lostcard byte = 0
	if handCards[index]+*godNum < 2 {
		//这个地方应该销毁一些东西
		*lostCards = nil
		return 0
	}
	if handCards[index]/2 > 0 {
		handCards[index] -= 2
		lostcard = 2
		//要记录一下
	} else {
		//不够求余，因为有为0的情况
		needgod := 2 - handCards[index]%2
		lostcard = handCards[index]
		handCards[index] = 0
		*godNum -= needgod
	}
	//生成记录
	*lostCards = append(*lostCards, lostcard)
	//如果是3个了，就记录一下，并且还原
	record := len(*lostCards)
	if record == 3 {
		//handCards[index]+=lostcard
		return 1
	} else {
		//继续递归
		result = this.CreateJieMeiPu_base_ex(handCards, godNum, index+1, lostCards)
		//还原
		//handCards[index]+=lostcard
	}
	return result
}

////构建姐妹铺结果
/*
整合所有结果来创建去掉姐妹铺的牌
//所有可能去拍的总和是2，不可能超过
如果单项就已经满足2的话，另外两色就不用处理了
关键是构造结果

*/
func (this *CheckHu) CreateJieMeiPuResult(handCards []byte, objWan []JieMeiPuObj, objTiao []JieMeiPuObj, objTong []JieMeiPuObj) *JieMeiPuResult {
	//这个是可能跨色的姐妹铺
	currentNum := 0    //用这个来记录当前已经有几个铺了
	var index byte = 0 //万条筒的下标不一样
	nomalResult := &JieMeiPuResult{}
	//var newHandCards []byte
	if objWan == nil {
		//没有的话就直接用原牌
		nomalResult.handCards = append(nomalResult.handCards, handCards[:9]...)
	} else {
		switch len(objWan) {
		case 1:
			//可能还有补缺的
			nomalResult.handCards = append(nomalResult.handCards, objWan[0].handCards[:]...)
			nomalResult.lastGodNum = objWan[0].lastGodNum
			nomalResult.index[0] = objWan[0].index + index
			nomalResult.lostCards[0] = objWan[0].lostCards
			nomalResult.isDouble = false
			currentNum = 1
		case 2:
			////这个就可以直接出去了
			//result := &JieMeiPuResult{
			//	lastGodNum: objWan[1].lastGodNum, //剩下鬼牌
			//}
			//只拿最后的手牌，那个是全去掉的
			nomalResult.handCards = append(nomalResult.handCards, objWan[1].handCards[:]...)
			nomalResult.handCards = append(nomalResult.handCards, handCards[9:]...)
			nomalResult.index[0] = objWan[0].index + index
			nomalResult.index[1] = objWan[1].index + index
			nomalResult.lostCards[0] = objWan[0].lostCards
			nomalResult.lostCards[1] = objWan[1].lostCards
			nomalResult.isDouble = true
			return nomalResult
		}
	}
	//检查条的
	if objTiao == nil {
		nomalResult.handCards = append(nomalResult.handCards, handCards[9:18]...)
	} else {
		//调整下标
		index = 9
		switch len(objTiao) {
		case 1:
			//这里要检查是不是存在万
			nomalResult.handCards = append(nomalResult.handCards, objTiao[0].handCards[:]...)
			nomalResult.lastGodNum = objTiao[0].lastGodNum
			if currentNum == 0 {
				nomalResult.index[0] = objTiao[0].index + index
				nomalResult.lostCards[0] = objTiao[0].lostCards
				nomalResult.isDouble = false
			} else {
				nomalResult.index[1] = objTiao[0].index + index
				nomalResult.lostCards[1] = objTiao[0].lostCards
				nomalResult.isDouble = true
			}
			currentNum++
		case 2:
			//这个就可以直接出去了
			//result:=&JieMeiPuResult{
			//	lastGodNum: objTiao[1].lastGodNum,//剩下鬼牌
			//}
			//只拿最后的手牌，那个是全去掉的
			//nomalResult.handCards=append(nomalResult.handCards,newHandCards[:]...)
			nomalResult.handCards = append(nomalResult.handCards, objTiao[1].handCards[:]...)
			nomalResult.handCards = append(nomalResult.handCards, handCards[18:]...)
			nomalResult.index[0] = objTiao[0].index + index
			nomalResult.index[1] = objTiao[1].index + index
			nomalResult.lostCards[0] = objTiao[0].lostCards
			nomalResult.lostCards[1] = objTiao[1].lostCards
			nomalResult.isDouble = true
			return nomalResult
		}
	}
	//检查筒的
	if objTong == nil {
		nomalResult.handCards = append(nomalResult.handCards, handCards[18:]...)
	} else {
		index = 18
		switch len(objTong) {
		case 1:
			nomalResult.handCards = append(nomalResult.handCards, objTong[0].handCards[:]...)
			nomalResult.lastGodNum = objTong[0].lastGodNum
			if currentNum == 1 {
				nomalResult.index[1] = objTong[0].index + index
				nomalResult.lostCards[1] = objTong[0].lostCards
				nomalResult.isDouble = true
			} else {
				nomalResult.index[0] = objTong[0].index + index
				nomalResult.lostCards[0] = objTong[0].lostCards
				nomalResult.isDouble = false
			}
			currentNum++
			//注意这里 要加上，因为不会在判断风字
			nomalResult.handCards = append(nomalResult.handCards, handCards[27:]...)
		case 2:
			////这个就可以直接出去了
			//result:=&JieMeiPuResult{
			//	lastGodNum: objTong[1].lastGodNum,//剩下鬼牌
			//}
			//只拿最后的手牌，那个是全去掉的
			//result.handCards=append(result.handCards,newHandCards[:]...)
			nomalResult.handCards = append(nomalResult.handCards, objTong[1].handCards[:]...)
			nomalResult.handCards = append(nomalResult.handCards, handCards[27:]...)
			nomalResult.index[0] = objTong[0].index + index
			nomalResult.index[1] = objTong[1].index + index
			nomalResult.lostCards[0] = objTong[0].lostCards
			nomalResult.lostCards[1] = objTong[1].lostCards
			nomalResult.isDouble = true
			return nomalResult
		}
	}
	return nomalResult
}

//func (this *CheckHu) CreateJieMeiPu(handCards []byte,index int,godNum byte) ([]byte, byte) {
//	var lostcard byte=0
//	var lostCards []byte
//	//this.CreateJieMeiPu_base(handCards,index,godNum,)
//	if handCards[index]/2>0{
//		handCards[index]-=2
//		lostcard=2
//		//要记录一下
//	}else{
//		//不够求余，因为有为0的情况
//		needgod:=2- handCards[index]%2
//		//第一位
//		lostcard=handCards[index]
//		handCards[index]=0
//		godNum-=needgod
//	}
//	//生成记录
//	lostCards=append(lostCards,lostcard)
//	return nil,0
//}
//一种花色里面可能有多个姐妹铺的牌型 单色 改成2个的情况试试
func (this *CheckHu) CheckJieMeiPu_base_ex(handCards []byte, index byte, godNum *byte, recordObj *JieMeiPuObj) (resultObj []JieMeiPuObj) {
	for i := index; i < byte(len(handCards)-2); i++ {
		checkgodNum := *godNum
		checkhardcards := make([]byte, 0)
		static.HF_DeepCopy(&checkhardcards, &handCards)
		//进构造单元
		var lostcards []byte
		if this.CreateJieMeiPu_base_ex(checkhardcards, &checkgodNum, i, &lostcards) == 1 {
			//再次进入？
			//fmt.Println(fmt.Sprintf("手牌（%v）", checkhardcards))
			//fmt.Println(fmt.Sprintf("index(%d),去牌结构（%v）godNum(%d)", i, lostcards, checkgodNum))
			//新建记录体
			if recordObj == nil {
				//如果是空就新建
				recordObj = &JieMeiPuObj{
					handCards:  checkhardcards,
					index:      i,
					lostCards:  lostcards,
					lastGodNum: checkgodNum,
				}
			} else {
				//如果有上层，那么就创建成切片
				newrecord := &JieMeiPuObj{
					handCards:  checkhardcards,
					index:      i,
					lostCards:  lostcards,
					lastGodNum: checkgodNum,
				}
				//生成切片
				resultObj = append(resultObj, *recordObj)
				resultObj = append(resultObj, *newrecord)
				return
			}
			resultObj = this.CheckJieMeiPu_base_ex(checkhardcards, i, &checkgodNum, recordObj)
			if resultObj == nil {
				resultObj = append(resultObj, *recordObj)
				return
			} else {
				return
			}
		}
	}
	return
}

//接替
func (this *CheckHu) check_base(cards []byte, godnum byte) (result [][]JieMeiPuObj) {
	//var recordObj []JieMeiPuObj
	var i byte
	for i = 0; i < byte(len(cards)-2); i++ {
		//首位不够就下一位
		if cards[i]+godnum < 2 {
			continue
		}
		recordObj := this.CheckJieMeiPu_base_ex(cards[:], i, &godnum, nil)
		if len(recordObj) != 0 {
			i = recordObj[0].index
			for _, v := range recordObj {
				//fmt.Println(fmt.Sprintf("去牌信息（%v）",v))
				_ = v
			}
			result = append(result, recordObj)
		}
	}
	return
}
func (this *CheckHu) CheckJieMeiPu_hu(checkCardObj *JieMeiPuResult, eyeMask byte) byte {
	if checkCardObj.handCards == nil {
		return static.CHK_NULL
	}
	hu, _ := this.CheckHU.Split_Byte(checkCardObj.handCards, checkCardObj.lastGodNum, eyeMask, true)
	if hu {
		//不分类了，根据
		if checkCardObj.lastGodNum != 0 {
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
		return static.WIK_CHI_HU
	}
	return static.CHK_NULL
}

//普通接口 姐妹铺
/*
想法就是，一色牌至少要有9张以上，然后就是顺序来排下去
*/
func (this *CheckHu) CheckJieMeiPu(handCards []byte, cbCurrentCard byte, isNormalCard bool, godCards []byte, eyeMask byte, usegod bool, seaf bool) (result int, err error) {
	//var resultObj []JieMeiPuResult
	resultObj, err := this.CheckJieMeiPu_Normal(handCards, cbCurrentCard, isNormalCard, godCards, eyeMask, usegod, seaf)
	if err != nil {
		return -1, err
	}
	if len(resultObj) == 0 {
		return 0, nil
	}
	for _, v := range resultObj {
		if v.isDouble {
			return 2, nil
		}
	}
	return 1, nil
}

func (this *CheckHu) CheckJieMeiPu_Normal(handCards []byte, cbCurrentCard byte, isNormalCard bool, godCards []byte, eyeMask byte, usegod bool, seaf bool) (checkCardObj []JieMeiPuResult, err error) {
	checkGodCards := []byte{}
	checkCardObj = make([]JieMeiPuResult, 0)
	if len(godCards) != 0 {
		static.HF_DeepCopy(&checkGodCards, &godCards)
	}
	if !usegod {
		checkGodCards = []byte{}
	}
	checkCards, handNum, gui_num, err := ReSetHandCards_Nomal(handCards, cbCurrentCard, isNormalCard, checkGodCards, seaf)
	if err != nil {
		return nil, err
	}
	if handNum < 7 {
		return nil, errors.New("手牌数量不足以生成姐妹铺")
	}
	//检查花色牌
	cards_wan := checkCards[:9]
	//检查条
	cards2_tiao := checkCards[9:18]
	//检查筒
	cards3_tong := checkCards[18:27]
	//处理万，可能返回几种
	var result_wan [][]JieMeiPuObj
	var result_tiao [][]JieMeiPuObj
	var result_tong [][]JieMeiPuObj
	var v []JieMeiPuObj
	var v_tiao []JieMeiPuObj
	var v_tong []JieMeiPuObj
	var newObj *JieMeiPuResult
	result_wan = this.check_base(cards_wan, gui_num)
RETWAN:
	if len(result_wan) == 0 {
		//万字里面没有，处理条
		result_tiao = this.check_base(cards2_tiao, gui_num)
	RETTIAO:
		if len(result_tiao) == 0 {
			//处理筒
			result_tong = this.check_base(cards3_tong, gui_num)
			if len(result_tong) == 0 {
				return
			} else {
				//只有筒有，那么就
				for _, v = range result_tong {
					newObj = this.CreateJieMeiPuResult(checkCards, nil, nil, v)
					//重组一下。因为这个时候可能还要用到赖子
					if len(godCards) != 0 && !usegod {
						newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
					}
					//checkCardObj=append(checkCardObj,*newObj)
					if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
						fmt.Println(fmt.Sprintf("只有筒（%v）", newObj))
						checkCardObj = append(checkCardObj, *newObj)
					}
				}
			}
		} else {
			//万是空的
			for _, v = range result_tiao {
				//如果直接就2个，那也就不用走下面的了
				switch len(v) {
				case 1:
					//newObj=this.CreateJieMeiPuResult(checkCards,nil,v,nil)
					//if len(godCards)!=0&&!usegod{
					//  newObj.handCards,_,newObj.lastGodNum,err=ReSetHandCards_Nomal(newObj.handCards,public.INVALID_BYTE,isNormalCard ,godCards)
					//}
					//if this.CheckJieMeiPu_hu(newObj,eyeMask)==public.WIK_CHI_HU{
					//  checkCardObj=append(checkCardObj,*newObj)
					//}
					//表明条子里面只有一个，再看筒里面有没有
					result_tong = this.check_base(cards3_tong, v[0].lastGodNum)
					//这个就不用再判断了
				RETTONG:
					if len(result_tong) == 0 {
						newObj = this.CreateJieMeiPuResult(checkCards, nil, v, nil)

						if len(godCards) != 0 && !usegod {
							newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
						}
						//checkCardObj=append(checkCardObj,*newObj)
						if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
							fmt.Println(fmt.Sprintf("只有条（%v）", newObj))
							checkCardObj = append(checkCardObj, *newObj)
						}
					} else {
						//这里筒最多也就一个
						for _, v_tong = range result_tong {
							newObj = this.CreateJieMeiPuResult(checkCards, nil, v, v_tong)

							if len(godCards) != 0 && !usegod {
								newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
							}
							//checkCardObj=append(checkCardObj,*newObj)
							if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
								fmt.Println(fmt.Sprintf("条-筒（%v）", newObj))
								checkCardObj = append(checkCardObj, *newObj)
							}
						}
						//这里也要追加，如果只有条的情况
						result_tong = nil
						goto RETTONG
					}
				case 2:
					//这里只有条
					newObj = this.CreateJieMeiPuResult(checkCards, nil, v, nil)
					if len(godCards) != 0 && !usegod {
						newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
					}
					//checkCardObj=append(checkCardObj,*newObj)
					if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
						fmt.Println(fmt.Sprintf("只有条（%v）", newObj))
						checkCardObj = append(checkCardObj, *newObj)
					}
				}
			}
			result_tiao = nil
			goto RETTIAO
		}
	} else {
		//万里面有东西
		for _, v = range result_wan {
			//如果直接就2个，那也就不用走下面的了
			switch len(v) {
			case 1:
				//下去之前先把自己加进去
				newObj = this.CreateJieMeiPuResult(checkCards, v, nil, nil)
				if len(godCards) != 0 && !usegod {
					newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
				}
				if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
					fmt.Println(fmt.Sprintf("万字添加初始（%v）", newObj))
					checkCardObj = append(checkCardObj, *newObj)
				}
				//这个还要走下面的
				result_tiao = this.check_base(cards2_tiao, v[0].lastGodNum)
				if len(result_tiao) == 0 {
					//表明条子里面没有，检查筒
					result_tong = this.check_base(cards3_tong, v[0].lastGodNum)
					//这个就不用再判断了
				RETTONG1:
					if len(result_tong) == 0 {
						//newObj=this.CreateJieMeiPuResult(checkCards,v,nil,nil)
						//fmt.Println(fmt.Sprintf("只有万（%v）",newObj))
						//if len(godCards)!=0&&!usegod{
						//	newObj.handCards,_,newObj.lastGodNum,err=ReSetHandCards_Nomal(newObj.handCards,public.INVALID_BYTE,isNormalCard ,godCards)
						//}
						////checkCardObj=append(checkCardObj,*newObj)
						//if this.CheckJieMeiPu_hu(newObj,eyeMask)==public.WIK_CHI_HU{
						//	checkCardObj=append(checkCardObj,*newObj)
						//}
					} else {
						//这里筒最多也就一个
						for _, v_tong = range result_tong {
							newObj = this.CreateJieMeiPuResult(checkCards, v, nil, v_tong)

							if len(godCards) != 0 && !usegod {
								newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
							}
							//checkCardObj=append(checkCardObj,*newObj)
							if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
								fmt.Println(fmt.Sprintf("万-筒（%v）", newObj))
								checkCardObj = append(checkCardObj, *newObj)
							}
						}
						//追加
						result_tong = nil
						goto RETTONG1
					}
				} else {
					//这里只能有一个条
					for _, v_tiao = range result_tiao {
						newObj = this.CreateJieMeiPuResult(checkCards, v, v_tiao, nil)
						if len(godCards) != 0 && !usegod {
							newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
						}
						//checkCardObj=append(checkCardObj,*newObj)
						if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
							fmt.Println(fmt.Sprintf("万-条（%v）", newObj))
							checkCardObj = append(checkCardObj, *newObj)
						}
					}
				}
			case 2:
				//这个就完了
				newObj = this.CreateJieMeiPuResult(checkCards, v, nil, nil)
				if len(godCards) != 0 && !usegod {
					newObj.handCards, _, newObj.lastGodNum, err = ReSetHandCards_Nomal(newObj.handCards, static.INVALID_BYTE, isNormalCard, godCards, seaf)
				}
				//checkCardObj=append(checkCardObj,*newObj)
				if this.CheckJieMeiPu_hu(newObj, eyeMask) == static.WIK_CHI_HU {
					fmt.Println(fmt.Sprintf("万字已经双了（%v）", newObj))
					checkCardObj = append(checkCardObj, *newObj)
				}
			}
			//还是要算一种情况，就是不给万配的情况，因为可能给万配了后。。。会导致不胡。直接goto吧
		}
		result_wan = nil
		goto RETWAN
	}
	return
}
