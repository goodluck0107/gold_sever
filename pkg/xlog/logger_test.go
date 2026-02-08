package xlog

import (
	"testing"
)

func TestLogger(t *testing.T) {
	Logger().Error("Error")
	Logger().Warn("Warn")
	Logger().Info("Info")
	Logger().Info("Info1")

	SetFileLevel("trace")

	Logger().Trace("Trace")
	Logger().Debug("Debug")
	Logger().Info("Info2")
	Logger().Info("Info3")

	SetConsoleLevel("err")
	Logger().Trace("Trace1")
	Logger().Debug("Debug2")
	Logger().Info("Info4")
	Logger().Info("Info5")
	Logger().Error("error111")
}

type person struct {
	Name string
	Age  int
	Tel  string
	info
}

type info struct {
	nickname string
	card     int
	isBlack  bool
}

func TestLogObj(t *testing.T) {
	p := person{
		Name: "hx",
		Age:  3,
		Tel:  "10086",
	}
	p.info.card = 3000
	p.info.isBlack = true
	p.info.nickname = "hxxxxx"
	WhitAny(&p).Errorf("person")
}

// 性能测试
// 不写日志文件测试结果   100000	     11730 ns/op
// 写日志文件测试结果    20000	        72800 ns/op
func BenchmarkLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Logger().Errorln("benchmark test")
	}
}

// 性能测试
// 不写日志文件测试结果   30000	     63766 ns/op
// 写日志文件测试结果    20000	     96049 ns/op
func BenchmarkLogf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Logger().WithFields(map[string]interface{}{
			"a": "1",
			"b": "2",
			"c": "3",
			"d": "4",
			"e": "5",
			"f": "6",
			"g": "7",
			"h": "8",
			"i": "9",
			"j": "10",
		}).Infoln("benchmark test")
	}
}
