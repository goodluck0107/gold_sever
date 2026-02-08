package object

import (
	"fmt"
	"testing"
)

type Person struct {
	Uid int64 `json:"uid"`
	Age uint8 `json:"age"`
}

func (p *Person) SayHello(op string) {
	fmt.Println(fmt.Sprintf("%s sayhello: my uid is %d, age is %d", op, p.Uid, p.Age))
}

type PersonMgr struct {
	Persons Map
}

func (pm *PersonMgr) AddPerson(p *Person) {
	pm.Persons.Add(p.Uid, p)
}

func (pm *PersonMgr) DelPerson(uid int64) {
	pm.Persons.Remove(uid)
}

func (pm *PersonMgr) GetPerson(uid int64) *Person {
	return pm.Persons.Get(uid).(*Person)
}

func (pm *PersonMgr) PersonExist(uid int64) bool {
	return pm.Persons.Contains(uid)
}

func (pm *PersonMgr) PersonCount() int {
	return pm.Persons.Len()
}

func (pm *PersonMgr) RangePerson(f func(uid int64, p *Person) bool, do func(uid int64, p *Person)) {
	pm.Persons.Range(
		func(key, value interface{}) bool {
			return f(key.(int64), value.(*Person))
		},
		func(key, value interface{}) {
			do(key.(int64), value.(*Person))
		},
	)
}

func (pm *PersonMgr) SortPerson(less func(i, j *Person) bool) (ps []*Person) {
	vs := pm.Persons.SortVal(func(i, j interface{}) bool {
		return less(i.(*Person), j.(*Person))
	})
	for _, v := range vs {
		ps = append(ps, v.(*Person))
	}
	return
}

// 写一个常用的实例测试一下哈
func TestMap(t *testing.T) {
	pm := new(PersonMgr)

	print := func(op string) {
		fmt.Println("\n------------------------------------")
		pm.RangePerson(
			func(uid int64, p *Person) bool {
				return true
			},
			func(uid int64, p *Person) {
				p.SayHello(op)
			},
		)
		fmt.Println("------------------------------------")
		fmt.Println()
	}

	// 为了方便看日志 设置一个标签，当然这里也可以不设置
	pm.Persons.SetTag("person")
	// 关掉默认的输出
	// SetLogger(nil)
	// 代替掉默认的输出
	// SetLogger(syslog.Logger())
	// 制作实例
	p0 := &Person{2, 50}
	p1 := &Person{3, 40}
	p2 := &Person{1, 45}
	p3 := &Person{0, 42}
	// add
	pm.AddPerson(p0)
	pm.AddPerson(p1)
	pm.AddPerson(p2)
	pm.AddPerson(p3)
	print("after add")

	pm.DelPerson(2)
	print("after del")

	ps := pm.SortPerson(func(i, j *Person) bool {
		return i.Uid < j.Uid
	})

	for _, p := range ps {
		p.SayHello("after sort by uid")
	}

	fmt.Println("-------------------------------------")

	ps = pm.SortPerson(func(i, j *Person) bool {
		return i.Age > j.Age
	})

	for _, p := range ps {
		p.SayHello("after sort by age")
	}
}
