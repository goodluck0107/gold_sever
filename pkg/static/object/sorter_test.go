package object

import (
	"fmt"
	"testing"
)

func TestSort(t *testing.T) {
	p1 := &Person{100001, 20}
	p2 := &Person{100002, 19}
	p3 := &Person{100003, 18}
	p4 := &Person{100004, 17}
	p5 := &Person{100005, 16}

	ps := make([]interface{}, 0)
	print := func() {
		for _, p := range ps {
			fmt.Println(fmt.Sprintf("%+v", p))
		}
		fmt.Println()
	}

	// 乱序追加person
	ps = append(ps,
		p1,
		p3,
		p4,
		p5,
		p2,
	)
	// 打印
	print()

	// 按uid从大到小排序
	SortAny(ps, func(i, j interface{}) bool {
		return i.(*Person).Uid > j.(*Person).Uid
	})
	print()

	// 按年龄从大到小排序
	SortAny(ps, func(i, j interface{}) bool {
		return i.(*Person).Age > j.(*Person).Age
	})
	print()

	slice := []int{
		2, 3, 1, 9, 6, 5, 4, 2, 1, 6, 10,
	}
	SortAny(slice, func(i, j interface{}) bool {
		return i.(int) < j.(int)
	})

	fmt.Println("slice after sorted:", slice)
}
