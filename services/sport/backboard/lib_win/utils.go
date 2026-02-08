package lib_win

import (
	"github.com/open-source/game/chess.git/pkg/static"
	mahlib2 "github.com/open-source/game/chess.git/services/sport/backboard"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	//常量定义
	MAX_WEAVE     = 4    //最大组合   （估计是杠的最大数）
	MAX_INDEX     = 34   //最大索引   （牌型索引）
	MAX_COUNT     = 14   //最大数目 （手牌最多数目）
	MAX_REPERTORY = 136  //最大库存(内地麻将个数)
	MAX_GUI_NUM   = 2    //最大鬼牌类数
	MASK_VALUE    = 0x0F //数值掩码
	GOD_CARD      = 0xFF //万能牌

	g_tablefile     = "table_%d.tbl"          //不带将的表
	g_eye_table     = "eye_table_%d.tbl"      //带混将表
	g_eye258_table  = "eye258_table_%d.tbl"   //带258将表
	g_feng_table    = "feng_table_%d.tbl"     //字牌不带将表
	gfeng_eye_table = "feng_eye_table_%d.tbl" //  字牌带将表

)

var (
	g_curpath           string
	g_tbldir            string = "tbl/"
	g_tartablefile      string
	g_tareye_table      string
	g_tareye258_table   string
	g_tarfeng_table     string
	g_tarfeng_eye_table string
)

type MhjongMatrix [][]int

//这个数据结构有点难挂钩，先放着吧
type ActRules struct {
	Chow bool `json"Chow"` //吃
	Pong bool `json"Pong"` //碰
	Kong bool `json"Kong"` //杠
	Win  bool `json"Win"`  //胡
}

//------------------------------------------
//分析子项
type TagAnalyseItem struct {
	CardEye    byte    //牌眼扑克
	WeaveKind  [4]byte //组合类型
	CenterCard [4]byte //中心扑克
}

//组合子项
type TagWeaveItem struct {
	WeaveKind   byte   `json:"weavekind"`   //组合类型 怎么凑成功的 吃碰杠
	CenterCard  byte   `json:"centercard"`  //中心扑克  重点牌，吃牌的话，就是吃的那张
	PublicCard  byte   `json:"publiccard"`  //公开标志  暗杠不公开
	ProvideUser uint16 `json:"provideuser"` //供应用户 估计是椅子号
}

//拖出来用 这个是文件操作的 将来扔出去算了
func fromSlash(path string) string {
	// Replace each '/' with '\\' if present
	var pathbuf []byte
	var lastSlash int
	for i, b := range path {
		if b == '/' {
			if pathbuf == nil {
				pathbuf = make([]byte, len(path))
			}
			copy(pathbuf[lastSlash:], path[lastSlash:i])
			pathbuf[i] = '\\'
			lastSlash = i + 1
		}
	}
	if pathbuf == nil {
		return path
	}

	copy(pathbuf[lastSlash:], path[lastSlash:])
	return string(pathbuf)
}

func init() {
	tardir := ""
	switch runtime.GOOS {
	case "windows":
		file, _ := exec.LookPath(os.Args[0])
		g_curpath, _ = filepath.Abs(filepath.Dir(file))
		g_curpath = g_curpath + string(os.PathSeparator)
		g_tbldir = fromSlash(g_tbldir)
		tardir = g_curpath + g_tbldir
		g_curpath = "./" + g_tbldir
		tardir = g_curpath
	case "linux":
		g_curpath = "./" + g_tbldir
		tardir = g_curpath
	}
	g_tartablefile = tardir + g_tablefile
	g_tareye_table = tardir + g_eye_table
	g_tareye258_table = tardir + g_eye258_table
	g_tarfeng_table = tardir + g_feng_table
	g_tarfeng_eye_table = tardir + gfeng_eye_table
}

// 是否存在n个c牌 cs是非索引序列
func Exist(c byte, codes []byte, n int) bool {
	for _, v := range codes {
		if n == 0 {
			return true
		}
		if c == v {
			n--
		}
	}
	return n == 0
}

// 对牌值从小到大排序，采用快速排序算法
func Sort(arr []byte, start, end int) {
	if start < end {
		i, j := start, end
		key := arr[(start+end)/2]
		for i <= j {
			for arr[i] < key {
				i++
			}
			for arr[j] > key {
				j--
			}
			if i <= j {
				arr[i], arr[j] = arr[j], arr[i]
				i++
				j--
			}
		}
		if start < j {
			Sort(arr, start, j)
		}
		if end > i {
			Sort(arr, i, end)
		}
	}
}

// // 验证给定的牌是否是有效的麻将牌值(内地) 136张的
// func Legal(card byte) bool {
// 	return common.IsMahjongCard(card)
// }

// //牌下标分离
// func cardsplit(card byte) (color byte, value byte) {
// 	_, color, value = common.Cardsplit(card)
// 	return
// }

// func CardToIndex(cbCardData byte) byte {
// 	result, err := common.CardToIndex(cbCardData)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return result
// }
// func CombinCard(color byte, value byte) byte {
// 	return common.CombinCard(color, value)
// }

// 去除句子(平摊方法)
func removeThreeLinkCards(mahjongMatrix []int) []int {
	marks := make([]int, 0, 4)
	for i := 0; i < len(mahjongMatrix)-2; i++ {
		if mahjongMatrix[i] > 0 && mahjongMatrix[i+1] > 0 && mahjongMatrix[i+2] > 0 {
			mahjongMatrix[i] -= 1
			mahjongMatrix[i+1] -= 1
			mahjongMatrix[i+2] -= 1
			//默认返回都是左吃，因为下标是在左
			marks = append(marks, ((i+1)*10 + 5))
			i--
		}
	}
	return marks
}

// 去除克子（平摊）
func removeTheSameThreeCards(mahjongMatrix []int) []int {
	marks := make([]int, 0, 4)
	for i := 0; i < len(mahjongMatrix); i++ {
		if mahjongMatrix[i] >= 3 {
			mahjongMatrix[i] -= 3
			marks = append(marks, ((i+1)*10 + 3))
		}
	}
	return marks
}

// 检测剩下的元素是否全为0
func checkMatrixAllElemEqualZero(mahjongMatrix []int) bool {
	for i := 0; i < len(mahjongMatrix); i++ {
		if mahjongMatrix[i] != 0 {
			return false
		}
	}
	return true
}

//检查成型 未设计完
//胡的哪张早查完了，现在直接算是什么牌型，
func checkPerfect(mahjongMatrix []int) *MhjongMatrix {
	var marks []int
	var maMarks = MhjongMatrix{}

	// mahjongMatrixList := getMahjongMatrixListByRemoveTwoCards(mahjongMatrix)
	for i := 0; i < len(mahjongMatrix); i++ {
		marks1 := removeThreeLinkCards(mahjongMatrix)
		marks2 := removeTheSameThreeCards(mahjongMatrix)
		//这里参考的是14张牌，普通胡法是按照3n+2=14来算的，所以n最多是4
		if len(marks1)+len(marks2) > 4 {
			continue
		}
		isPerfect := checkMatrixAllElemEqualZero(mahjongMatrix)
		if isPerfect {
			marks = append(marks1, marks2...)
		}
		//多种可能性的组合
		maMarks = append(maMarks, marks)
	}
	return &maMarks
}

func GetTingCheckCards(cbCardIndex []byte) (result []byte) {
	for i := byte(0); i < MAX_INDEX-2; i++ {
		if cbCardIndex[i] != 0 {
			//连续检查后面两张有没有
			if i+1 > MAX_INDEX {
				result = append(result, i)
				break
			}
			if cbCardIndex[i+1] != 0 {
				result = append(result, i+1)
			}
			if i+2 > MAX_INDEX {
				result = append(result, i)
				break
			}

		}
	}
	return nil
}

//准备独立 创建手牌新序列
func ReSetHandCards_Nomal(cbCardIndex []byte, cbCurrentCard byte, isNormalCard bool, godCards []byte, seaf bool) (checkCards []byte, handcardNum int, gui_num byte, err error) {
	if seaf {
		handcardNum, err = mahlib2.CheckHandCardsSafe_ex(cbCardIndex, cbCurrentCard)
		if err != nil {
			return nil, handcardNum, 0, err
		}
	}
	static.HF_DeepCopy(&checkCards, &cbCardIndex)
	var gui_1 byte = 0xff
	var gui_2 byte = 0xff

	index := byte(len(godCards))
	switch index {
	case 1:
		gui_1 = godCards[0]
	case 2:
		gui_1 = godCards[0]
		gui_2 = godCards[1]
	}
	//去掉赖子牌
	gui1index, _ := mahlib2.CardToIndex(gui_1)
	if gui1index < 34 {
		gui_num = checkCards[gui1index] % 0xff
		checkCards[gui1index] = 0
	}
	gui2index, _ := mahlib2.CardToIndex(gui_2)
	if gui2index < 34 {
		gui_num += checkCards[gui2index] % 0xff
		checkCards[gui2index] = 0
	}
	if cbCurrentCard != static.INVALID_BYTE {
		if isNormalCard {
			index, _ := mahlib2.CardToIndex(cbCurrentCard)
			checkCards[index]++
		} else {
			gui_num++
		}
	}
	return
}

//准备独立 追加倒牌数据
func ReSetHandwithWeave_Nomal(cbCardIndex []byte, WeaveItem []static.TagWeaveItem) (checkCards []byte) {
	//把倒牌也加进去
	static.HF_DeepCopy(&checkCards, &cbCardIndex)
	if WeaveItem == nil {
		return
	}
	for _, v := range WeaveItem {
		switch v.WeaveKind {
		case static.WIK_LEFT: //左吃
			index, _ := mahlib2.CardToIndex(v.CenterCard)
			checkCards[index] += 1
			checkCards[index+1] += 1
			checkCards[index+1] += 1
		case static.WIK_RIGHT: //右吃
			index, _ := mahlib2.CardToIndex(v.CenterCard)
			checkCards[index] += 1
			checkCards[index-1] += 1
			checkCards[index-2] += 1
		case static.WIK_CENTER: //中吃
			index, _ := mahlib2.CardToIndex(v.CenterCard)
			checkCards[index] += 1
			checkCards[index+1] += 1
			checkCards[index-1] += 1
		case static.WIK_PENG: //碰只加一次
			index, _ := mahlib2.CardToIndex(v.CenterCard)
			checkCards[index] += 1
		case static.WIK_GANG: //杠只加一次
			index, _ := mahlib2.CardToIndex(v.CenterCard)
			checkCards[index] += 1
		}
	}
	return
}
