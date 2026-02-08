package static

import (
	"crypto/rand"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	stdRand "math/rand"
	"slices"
)

var nickNames = []string{
	"666",
	"嘟嘟",
	"灸一灸",
	"蔷薇涂料",
	"心如止水",
	"2233",
	"肥肥不肥",
	"旧约@@",
	"趣兜兜",
	"徐行发",
	"blessing",
	"钢铁弯弯",
	"老古董",
	"趣绘",
	"许师傅",
	"i do",
	"孤舟",
	"冷夜咆哮",
	"顺丰如意",
	"叶子",
	"ku印",
	"好运来",
	"燎原",
	"同城快修",
	"一会见",
	"OIP-C",
	"花开-花落",
	"柳姐",
	"王者归来",
	"一毛不拔",
	"today",
	"辉煌",
	"梅梅",
	"怡美",
	"莫入",
	"TT",
	"汇泉",
	"美惠子",
	"望城送果",
	"幽然",
	"whot",
	"健齿康",
	"梦天",
	"西耐开关",
	"云游四方",
	"草木",
	"禁锢",
	"梦魇",
	"下一站",
	"再来一场",
	"对酒当歌",
}

func GenIPv4() string {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err) // 仅示例，实际应处理错误
	}
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

// 获取随机昵称
func GenGuestRandomNickname() string {
	return `S745` + HF_GetRandomNumberString(5)
}

// 获取苹果随机昵称
func GenAppleRandomNickname() string {
	return `S745` + HF_GetRandomNumberString(5)
}

var allGenders = []int{
	consts.SexUnknown,
	consts.SexMale,
	consts.SexFemale,
}

func GenGender() int {
	return allGenders[stdRand.Intn(len(allGenders))]
}

var RobotIdSuffix = []int64{
	2, 3, 5,
}

func GenRobotId() int64 {
	prefix := stdRand.Int63n(900000) + 100000
	return prefix*10 + RobotIdSuffix[stdRand.Intn(len(RobotIdSuffix))]
}

func GenRobotUrl() string {
	return fmt.Sprintf("head%d.jpg", stdRand.Intn(4)+1)
}

var usedUrlAndHead = make(map[int64][]int)

func GenRobotNameUrl(hid int64) (string, string) {
	avaIdx := make([]int, 0, len(nickNames))
	for idx := range nickNames {
		if !slices.Contains(usedUrlAndHead[hid], idx) {
			avaIdx = append(avaIdx, idx)
		}
	}
	if len(avaIdx) == 0 {
		//name := GenGuestRandomNickname()
		//url := GenRobotUrl()
		return "", ""
	}
	idx := avaIdx[stdRand.Intn(len(avaIdx))]
	url := fmt.Sprintf("head%d.jpg", idx+1)
	name := nickNames[idx]
	usedUrlAndHead[hid] = append(usedUrlAndHead[hid], idx)
	return name, url
}

func RecycleRobotNameUrl(hid int64, name string) {
	if idx := slices.Index(nickNames, name); idx >= 0 {
		usedUrlAndHead[hid] = slices.DeleteFunc(usedUrlAndHead[hid], func(i int) bool {
			return i == idx
		})
	}
}
