package static

import (
	"regexp"
	"strconv"
)

// 校验手机号
func CheckMobile(str string) bool {
	reg := regexp.MustCompile(`^1(3|4|5|6|7|8|9)\d{9}$`)
	return reg.MatchString(str)
}

// 校验身份证
func CheckIdcard(idcard string) bool {
	if len(idcard) != 18 {
		return false
	}

	var id_card [18]byte // 'X' == byte(88)， 'X'在byte中表示为88
	var id_card_copy [17]byte

	// 将字符串，转换成[]byte,并保存到id_card[]数组当中
	for k, v := range []byte(idcard) {
		id_card[k] = byte(v)
	}

	//复制id_card[18]前17位元素到id_card_copy[]数组当中
	for j := 0; j < 17; j++ {
		id_card_copy[j] = id_card[j]
	}

	check_id := func(id [17]byte) int {
		arry := make([]int, 17)
		//强制类型转换，将[]byte转换成[]int ,变化过程
		// []byte -> byte -> string -> int
		//将通过range 将[]byte转换成单个byte,再用强制类型转换string()，将byte转换成string
		//再通过strconv.Atoi()将string 转换成int 类型
		for index, value := range id {
			arry[index], _ = strconv.Atoi(string(value))
		}

		var wi [17]int = [...]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
		var res int
		for i := 0; i < 17; i++ {
			res += arry[i] * wi[i]
		}
		return (res % 11)
	}

	byte2int := func(x byte) byte {
		if x == 88 {
			return 'X'
		}
		return (x - 48) // 'X' - 48 = 40;
	}

	verify := check_id(id_card_copy)
	last := byte2int(id_card[17])
	var temp byte
	var i int
	a18 := [11]byte{1, 0, 'X', 9, 8, 7, 6, 5, 4, 3, 2}

	for i = 0; i < 11; i++ {
		if i == verify {
			temp = a18[i]
			break
		}
	}

	if temp == last {
		return true
	} else {
		return false
	}
}
