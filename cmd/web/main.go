package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
)

var (
	path string
	port int
	help bool
)

func init() {
	flag.StringVar(&path, "v", "./", "web wuhan path")
	flag.IntVar(&port, "p", 8888, "listen port")
	flag.BoolVar(&help, "h", false, "see help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	p, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(fmt.Sprintf("web wuhan path = %s, listen on %d", p, port))

	http.Handle("/", http.FileServer(http.Dir(p)))
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

type Route struct {
}

func (Route) Login(arg interface{}) (int32, interface{}) {
	return 0, nil
}

func (Route) Logout(arg interface{}, a int, b string) (int32, interface{}) {
	return 0, nil
}

var m map[string]func(arg interface{}) (int32, interface{})

func R(action string, handleFn func(arg interface{}) (int32, interface{})) {
	m[action] = handleFn
}

type ErrCode uint32

type ErrMsg error

func HTTP() {
	out := Handle("Login", "123123", "wxappid", 50)

	for _, o := range out {
		switch o.(type) {
		case ErrCode:

		case ErrMsg:

		case error:

		default:

		}
	}
}

func Handle(action string, args ...interface{}) []interface{} {
	h, ok := m1[action]
	if ok {
		in := make([]reflect.Value, 0, len(args))
		for _, a := range args {
			in = append(in, reflect.ValueOf(a))
		}
		out := h.Call(in)
		res := make([]interface{}, 0, len(out))
		for _, o := range out {
			res = append(res, o.Interface())
		}
		return res
	}
	return []interface{}{-1, fmt.Errorf("not accpet")}
}

var m1 map[string]reflect.Value

func R1(router interface{}) {
	t := reflect.TypeOf(router)
	c := t.NumMethod()
	for i := 0; i < c; i++ {
		m := t.Method(i)
		m1[m.Name] = m.Func
	}
}
