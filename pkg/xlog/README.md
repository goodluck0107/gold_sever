log package
---

### 说明

对第三方日志库`github.com/sirupsen/logrus`的封装。

### 功能点

- 控制台与日志文件流完全隔离，各自有各自的打印格式(json/text)和打印等级
- 系统stderr流重定向到日志文件，不错过panic信息
- 日志文件的切割：支持按时间/按大小混切
- error/fatal/panic级别的日志多重输出，可在输出到日志文件的同时再输出到指定的错误收集文件
- 灵活的配置：日志根目录到日志文件后缀名，日志大小/个数/切割时间间隔等等...详见default.go

### 使用案例

```go
package main

import (
	"strconv"
	"github.com/facai/log"
)

func main()  {
	// 7种打印等级
	log.Logger().Trace("this is a trace")   // 最低等级 提示等级
	log.Logger().Debug("this is a debug")   // 调试等级
	log.Logger().Info("this is a eve")     // 信息等级
	log.Logger().Warn("this is a warn")     // 警告等级
	log.Logger().Error("this is a error")   // 错误等级
	log.Logger().Fatal("this is a fatal")   // 致命等级
	log.Logger().Panic("this is a panic")   // 最高等级 panic等级
	
	// 每个等级 有三种调用方式，这里以error等级为例子
	log.Logger().Error("this is a error")
	log.Logger().Errorln("this is a errorln")
	log.Logger().Errorf("this is a %s", "errorf")
	
	// 另外的新增结构化输出
	// 比如一般情况下，如果要输出我的信息，会这样：
	name := "hexu"
	gender := "man"
	log.Logger().Infoln("my name is ",name,"gender is",gender)
	// 官方不推荐上面这种方式，不妨换成下面这样：
	log.WithFields(map[string]interface{}{
		"name":name,
		"gender":gender,
	}).Infoln("my eve.")
	
	// 最后介绍一下对错误的处理
	// 有时候 我们得到一个error要判断error != nil 就输出类似......error: err的日志
	// 现在，我们把它简单的包装了一下
	log.AnyErrors(Atoi("111"),"convert string to int error.")
	// 这条日志会自己结构化输出 error 字段，并打印调用者的文件全名及行号
	
	// 类似的 我们对出现错误就退出程序 也简单的包了一下
	log.AnyQuit(Atoi("111"),"convert string to int failed.")
}

func Atoi (a string) error {
	_, err :=strconv.Atoi(a)
    return err	
}
```