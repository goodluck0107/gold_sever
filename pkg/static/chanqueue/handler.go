package chanqueue

import (
	"fmt"
	"os"
)

type Handler interface {
	// 初始化协议处理者
	OnInit()
	// 根据协议头拿到协议处理者
	WorkerByProto(proto string) (*Worker, bool)
}

type defaultHandler struct {
}

func (d *defaultHandler) OnInit() {
	_, _ = fmt.Fprintln(os.Stderr, "implement Handler 1")
}

func (d *defaultHandler) WorkerByProto(proto string) (*Worker, bool) {
	_, _ = fmt.Fprintln(os.Stderr, "implement Handler 2")
	return nil, false
}
