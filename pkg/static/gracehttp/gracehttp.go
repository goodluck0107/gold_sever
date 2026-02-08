// +build !windows

package gracehttp

import (
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net/http"
	"time"
)

const (
	DEFAULT_READ_TIMEOUT  = 60 * time.Second
	DEFAULT_WRITE_TIMEOUT = DEFAULT_READ_TIMEOUT
)

// common adapt to the framework
type HttpServr struct {
	GraceHttp       *Server
	RpcCloseChan    chan byte
	SessionWaitChan chan byte
}

// call do
type call struct {
	wait time.Duration
	call chan byte
}

// refer http.ListenAndServe
func ListenAndServe(addr string, handler http.Handler) error {
	return NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT).ListenAndServe()
}

// refer http.ListenAndServeTLS
func ListenAndServeTLS(addr string, certFile string, keyFile string, handler http.Handler) error {
	return NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT).ListenAndServeTLS(certFile, keyFile)
}

// Get the wuhan pointer
func GetHttpServr(addr string, handler http.Handler) *HttpServr {
	return &HttpServr{
		GraceHttp:       NewServer(addr, handler, DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT),
		RpcCloseChan:    make(chan byte, 0),
		SessionWaitChan: make(chan byte, 0),
	}
}

// [seesion is nil]
func (this *HttpServr) Done() {
	this.SessionWaitChan <- 0
}

// [reboot] args:how long to wait all session close
// 向外暴露&实现重启&唯一接口
func (this *HttpServr) ReBoot(waittime time.Duration) bool {
	if this.GraceHttp == nil {
		xlog.Logger().Warningln("the wuhan's gracehttp is nil pointer: windows can't reboot wuhan...")
		return false
	}
	//！DO REBOOT
	xlog.Logger().Warningln(">#REBOOT NOW TIME::#", time.Now().Format(static.TIMEFORMAT))
	//！尽可能等待所有的websocket断开
	waitchan := make(chan byte, 0)
	go this.do(call{waittime, waitchan})
	xlog.Logger().Warningln("#wait for all seesion close....")
	<-waitchan
	xlog.Logger().Warningln("#all seesion closed. then close RPC...")
	//！关掉rpc
	xlog.Logger().Warningln("#send RpcCloseChan addr:", &this.RpcCloseChan)
	this.RpcCloseChan <- 0
	xlog.Logger().Warningln("#RPC closed. then reboot wuhan")
	//！重启服务器
	this.GraceHttp.ReStartServe()
	xlog.Logger().Warningln(">#REBOOT FINISHED TIME::#", time.Now().Format(static.TIMEFORMAT))
	return true
}

// [do wait]
func (this *HttpServr) do(c call) {
	defer func() {
		c.call <- 0
	}()

	timeout := time.After(c.wait)

	for {
		select {
		case <-timeout:
			xlog.Logger().Warningln("wait for webscoket all close TIME OUT....")
			return
		case <-this.SessionWaitChan:
			xlog.Logger().Warningln("one person offline")
			return
		}
	}
	return
}
