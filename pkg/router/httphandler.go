package router

import (
	"encoding/json"
	"fmt"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/sirupsen/logrus"
	"net/http"
	"reflect"
	"runtime/debug"
)

type HttpHandler func(string, interface{}) (interface{}, *xerrors.XError)

type HttpInfo struct {
	DataType reflect.Type
	Handle   HttpHandler
}

type HttpWorker struct {
	FuncProtocol map[string]*HttpInfo
	Prefix       string
	Encode       int
	EncodePhpKey string
}

func (work *HttpWorker) SetEncode(encode int) *HttpWorker {
	work.Encode = encode
	return work
}

func (work *HttpWorker) SetPrefix(prefix string) *HttpWorker {
	work.Prefix = prefix
	return work
}
func (work *HttpWorker) SetEncodePhpKey(key string) *HttpWorker {
	work.EncodePhpKey = key
	return work
}

func (self *HttpWorker) Router(header string, proto interface{}, handler HttpHandler) {
	var info HttpInfo
	info.DataType = reflect.TypeOf(proto)
	info.Handle = handler
	if self.FuncProtocol == nil {
		self.FuncProtocol = make(map[string]*HttpInfo)
	}
	self.FuncProtocol[header] = &info
}

func (self *HttpWorker) GetProtocol(protocol string) *HttpInfo {
	return self.FuncProtocol[protocol]
}

var httpWorker *HttpWorker

func HttpCommon(w http.ResponseWriter, req *http.Request) (*static.Msg_Header, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json
	defer func() {
		x := recover()
		if x != nil {
			xlog.Logger().Errorln(x, string(debug.Stack()))
		}
	}()
	data := req.FormValue("msgdata")
	//head := req.FormValue("msghead")
	xlog.Logger().WithFields(logrus.Fields{
		"clientIp": static.HF_GetHttpIP(req),
		"data":     data,
	}).Infoln("【RECEIVED HTTP】")

	var msg static.Msg_Header
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		return nil, err
	}
	// 消息解密处理
	err = getCommonMsg(&msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

//CreateHTTP 创建
func CreateHTTP(pattern string, work *HttpWorker) {
	httpWorker = work
	httpWorker.SetPrefix(pattern)
	http.HandleFunc(pattern, HTTPServer)
	return

}

//HTTPServer http服务
func HTTPServer(w http.ResponseWriter, req *http.Request) {
	msg, err := HttpCommon(w, req)
	if err != nil {
		xlog.Logger().Errorln("admin http wuhan error:", err) //TODO 日志功能
		w.Write(GetErrReturn("", xerrors.ArgumentError))
		return
	}

	if msg == nil {
		w.Write(GetErrReturn("", xerrors.ArgumentError))
		return
	}

	if httpWorker.Encode != consts.EncodeNone && msg.Sign.Encode == consts.EncodeNone {
		w.Write(GetErrReturn("", xerrors.ArgumentError))
		return
	}
	xlog.Logger().Infoln(fmt.Sprintf("%s-msgdata:%+v", httpWorker.Prefix, msg))
	handle := httpWorker.GetProtocol(msg.Header)
	if handle != nil {
		data := reflect.New(handle.DataType).Interface()
		err := json.Unmarshal([]byte(msg.Data), data)
		if err != nil {
			xlog.Logger().Errorln(err)
			w.Write(GetErrReturn(msg.Header, xerrors.NewXError(err.Error())))
			return
		}
		res, errResp := handle.Handle(msg.Header, data)
		if errResp != xerrors.RespOk {
			w.Write(GetErrReturn(msg.Header, errResp))
			return
		}
		if resByte, ok := res.([]byte); ok {
			w.Write(static.HF_JtoB(GetCommonResp(msg.Header, string(resByte))))
			return
		}
		w.Write(static.HF_JtoB(GetCommonResp(msg.Header, res)))
		return
	}
	w.Write([]byte("method not found"))
}
