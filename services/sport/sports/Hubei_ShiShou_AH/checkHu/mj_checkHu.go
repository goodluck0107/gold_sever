package ssahCheckHu

import (
	"fmt"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	components2 "github.com/open-source/game/chess.git/services/sport/components"
	"math"
)

const (
	GameHuNull uint64 = 0x0000 //不能胡
	//GameHuMagic        = 0x0001 //有癞子
	GameHuPiH = 0x0002 //屁胡
	GameHuPPH = 0x0004 //碰碰胡
	GameHuQYS = 0x0008 //清一色
	GameHu7D  = 0x0010 //7对
	GameHu7D1 = 0x0020 //豪华7对
	GameHu7D2 = 0x0040 //双豪华7对
	GameHu7D3 = 0x0080 //三豪华7对
	GameHuJYS = 0x0100 //将一色
	GameHuGK  = 0x1000 //杠开
	GameHuQGH = 0x2000 //抢杠胡
	GameHuQH  = 0x4000 //抢杠胡

)

func CheckHu(playerInfo components2.PlayerMeta, magic int, curCardIndex int, isZiMo bool, isHaveCHH bool, output func(uint16, string)) uint64 {
	cards := static.HF_BytesToInts(playerInfo.CardIndex[:])
	tmpCards := make([]int, len(cards))
	copy(tmpCards, cards)
	if !isZiMo {
		tmpCards[curCardIndex] += 1
	}
	output(static.INVALID_CHAIR,
		fmt.Sprintf("check hu Cards: %v, weatCount=%d", tmpCards, playerInfo.WeaveItemCount))
	magicIndex := 0
	if magic == 0 {
		magicIndex = -1
	} else {
		tempMagic := Num16to10(magic)
		magicIndex = trimCardIndex(tempMagic) //  赖子牌值  转  索引数组
	}
	//syslog.Logger().Print("癞子值：", tempMagic)
	//syslog.Logger().Print("癞子索引：", magicIndex)
	//syslog.Logger().Print("炮牌索引：", paoCardIndex)
	lib := mjHulib{}
	a := lib.getHuInfo(&playerInfo, tmpCards, magicIndex, isHaveCHH)
	output(static.INVALID_CHAIR, fmt.Sprintf("get hu info res=%v", a))
	if a != GameHuNull {
		qys := GetmjMacthInstance().CheckQYS(&playerInfo)
		if qys != GameHuNull {
			output(static.INVALID_CHAIR, fmt.Sprintf("qing yi se = %v", qys))
		}
		a |= qys
	}
	xlog.Logger().Print("check hu return :", a)
	return a
}

func CheckTing(playerInfo components2.PlayerMeta, magic int, curCardIndex int, isZiMo bool) []int {
	cards := static.HF_BytesToInts(playerInfo.CardIndex[:])
	tmpCards := make([]int, len(cards))
	copy(tmpCards, cards)
	if isZiMo {
		tmpCards[curCardIndex] -= 1
	}
	magicIndex := 0
	if magic == 0 {
		magicIndex = -1
	} else {
		tempMagic := Num16to10(magic)
		magicIndex = trimCardIndex(tempMagic) //  赖子牌值  转  索引数组
	}
	//syslog.Logger().Print("癞子值：", tempMagic)
	//syslog.Logger().Print("癞子索引：", magicIndex)
	//syslog.Logger().Print("炮牌索引：", paoCardIndex)
	lib := mjHulib{}
	tingNumArr := make([]int, 0)
	for j := 0; j < 34; j++ {
		tmpCards[j] += 1
		res := lib.getHuInfoLog(tmpCards, magicIndex)
		if res {
			tingNumArr = append(tingNumArr, j)
		}
		tmpCards[j] -= 1
	}

	return tingNumArr
}

//测试胡
func CheckHuTest(tmpCards []int, magic int) bool {
	magicIndex := 0
	if magic == 0 {
		magicIndex = -1
	} else {
		tempMagic := Num16to10(magic)
		magicIndex = trimCardIndex(tempMagic) //  赖子牌值  转  索引数组
	}
	lib := mjHulib{}
	a := lib.getHuInfoLog(tmpCards, magicIndex)
	return a
}

func CheckHuLog(playerInfo components2.PlayerMeta, magic int, curCardIndex int, isZiMo bool) bool {
	cards := static.HF_BytesToInts(playerInfo.CardIndex[:])
	tmpCards := make([]int, len(cards))
	copy(tmpCards, cards)
	if !isZiMo {
		tmpCards[curCardIndex] += 1
	}
	magicIndex := 0
	if magic == 0 {
		magicIndex = -1
	} else {
		tempMagic := Num16to10(magic)
		magicIndex = trimCardIndex(tempMagic) //  赖子牌值  转  索引数组
	}
	lib := mjHulib{}
	a := lib.getHuInfoLog(tmpCards, magicIndex)
	return a
}

func MaxIntObj(obj map[int]int) int {
	if obj[0] > obj[1] {
		return 0
	}
	return 1
}

func Num16to10(val int) int {
	return int(math.Floor(float64(val/16)))*10 + val%16
}
