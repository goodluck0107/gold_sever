package main

import (
	"fmt"
	"unsafe"
)

func main() {
	//call(func() {
	//	panic(nil)
	//})

	fmt.Println("Nil 占用的字节数是：" + fmt.Sprint(unsafe.Alignof(Nil)))
	fmt.Println("Nil 占用的字节数是：" + fmt.Sprint(unsafe.Sizeof(Nil)))
	fmt.Println("S1.Nil 占用的字节数是：" + fmt.Sprint(unsafe.Alignof(S1{}.Nil)))
	fmt.Println("S1 占用的字节数是：" + fmt.Sprint(unsafe.Sizeof(S1{})))
}

func call(fn func()) {
	fn()
}

var Nil struct{}

type S1 struct {
	A  int32
	Nil  struct{}
}
