package ssahCheckHu

func changeIndexArr(arr ...int) (ret []int) {
	ret = make([]int, MAX_CARD)
	for i := 0; i < len(ret); i++ {
		ret[i] = 0
	}
	for i := 0; i < len(arr); i++ {
		if arr[i] >= 1 && arr[i] <= 9 {
			ret[arr[i]-1] += 1
		} else if arr[i] >= 11 && arr[i] <= 19 {
			ret[arr[i]-2] += 1
		} else if arr[i] >= 21 && arr[i] <= 29 {
			ret[arr[i]-3] += 1
		} else if arr[i] >= 31 && arr[i] <= 37 {
			ret[arr[i]-4] += 1
		}
	}
	return
}

//将正常手牌牌值转成 0-33 索引
func trimCardIndex(num int) (n int) {
	if num <= 9 && num > 0 {
		n = num - 1
	} else if num <= 19 && num >= 11 {
		n = num - 2
	} else if num <= 29 && num >= 21 {
		n = num - 3
	} else if num <= 37 && num >= 31 {
		n = num - 4
	}
	return
}

//将 0-33 索引 转成  正常手牌牌值
func indexChangeZi(num int) (n int) {
	if num <= 8 && num >= 0 {
		n = num + 1
	} else if num <= 17 && num >= 9 {
		n = num + 2
	} else if num <= 26 && num >= 18 {
		n = num + 3
	} else if num <= 33 && num >= 27 {
		n = num + 4
	}
	return
}

//将 0-33 索引数组 转成  正常手牌牌值数组
func indexChangeZiArr(cardArr ...int) (resArr []int) {
	for i := 0; i < len(cardArr); i++ {
		if cardArr[i] <= 8 && cardArr[i] > 0 {
			resArr = append(resArr, cardArr[i]+1)
		} else if cardArr[i] <= 17 && cardArr[i] >= 9 {
			resArr = append(resArr, cardArr[i]+2)
		} else if cardArr[i] <= 26 && cardArr[i] >= 18 {
			resArr = append(resArr, cardArr[i]+3)
		} else if cardArr[i] <= 33 && cardArr[i] >= 27 {
			resArr = append(resArr, cardArr[i]+4)
		}
	}
	return
}

//将 正常手牌牌值 转成 同花色 索引区间  （清一色报清用）
func cardChangeSection(num int) (arr []int) {
	if num >= 1 && num <= 9 {
		arr = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	} else if num >= 11 && num <= 19 {
		arr = []int{9, 10, 11, 12, 13, 14, 15, 16, 17}
	} else if num >= 21 && num <= 29 {
		arr = []int{18, 19, 20, 21, 22, 23, 24, 25, 26}
	}
	return
}
