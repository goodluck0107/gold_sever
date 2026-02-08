package router

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/sirupsen/logrus"
)

// 注册一条路由
func NewRouter(pattern string, handler IAppHandler) {
	http.HandleFunc(pattern, httpWrapper(handler))
}

// http响应函数
type AppHandlerFunc func(req *http.Request, data interface{}) (interface{}, *xerrors.XError)

type AppWorker struct {
	DataType reflect.Type
	Handle   AppHandlerFunc
}

// http处理者协议
type IAppHandler interface {
	// 初始化
	OnInit()
	// 根据消息头得到处理函数
	AppHandler(handler string) *AppWorker
	// 判断是否加密
	EncodeInfo() (encode int, key string)
}

// http 包装器
func httpWrapper(handler IAppHandler) func(w http.ResponseWriter, req *http.Request) {
	// 非空校验
	if handler == nil {
		xlog.Logger().Panic("app handler is nil, please check at http wrapper.")
		return func(w http.ResponseWriter, req *http.Request) {
			// HTTP 510 服务器未满足响应策略
			http.Error(w, http.StatusText(http.StatusNotExtended), http.StatusNotExtended)
		}
	}
	// 初始化http响应者
	handler.OnInit()
	// 包装
	return func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				xlog.Logger().Errorln(r, string(debug.Stack()))
				// HTTP 500. 服务器异常
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		// 得到加密信息
		encode, key := handler.EncodeInfo()
		// 解析消息
		msgBase, err := httpParse(w, req, key)
		// 响应函数
		response := func(v interface{}, err error) {
			// 写入http状态码
			w.WriteHeader(http.StatusOK)
			header := ""
			if msgBase != nil {
				header = msgBase.Header
			}
			var customError *xerrors.XError
			if err != nil {
				var ok bool
				customError, ok = err.(*xerrors.XError)
				if !ok {
					customError = xerrors.NewXError(err.Error())
				}
			}
			// 正式返回
			if _, err := w.Write(_GetResponse(header, v, customError, encode, key)); err != nil {
				xlog.Logger().Errorln("http response error:", err)
			}
		}
		// 错误检查
		if err != nil {
			xlog.Logger().Errorln("http wuhan parse msg error:", err)
			response(nil, xerrors.ArgumentError)
			return
		}
		// 如果服务器开启了加密 客户端未开启，则返回参数错误。
		if encode != consts.EncodeNone && msgBase.Sign.Encode == consts.EncodeNone {
			xlog.Logger().Errorln("the wuhan turned on encryption but the client did not.")
			response(nil, xerrors.ArgumentError)
			return
		}
		// 从http处理者获取协议对应的错误函数模型
		if appHandler := handler.AppHandler(msgBase.Header); appHandler != nil {
			msgData := reflect.New(appHandler.DataType).Interface()
			err = json.Unmarshal([]byte(msgBase.Data), msgData)
			if err != nil {
				xlog.Logger().Errorln("app handler parse msg error:", err)
				response(nil, err)
				return
			}
			// 开始处理请求
			response(appHandler.Handle(req, msgData))
		} else {
			// 未受理的协议
			xlog.Logger().Errorln("未受理的协议:", msgBase.Header)
			// http 404. 资源未找到
			http.Error(w, fmt.Sprintf("%s:the requested agreement [%s] was not accepted.",
				http.StatusText(http.StatusNotFound),
				msgBase.Header,
			), http.StatusNotFound)
		}
	}
}

func httpParse(w http.ResponseWriter, req *http.Request, encodeKey string) (*static.Msg_Header, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             // 允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") // header的类型
	w.Header().Set("content-type", "application/json")             // 返回数据格式是json
	data := req.FormValue("msgdata")

	var msg static.Msg_Header
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		return nil, err
	}
	if msg.Sign.Encode == consts.EncodeAes {
		request, err := base64.URLEncoding.DecodeString(msg.Data)
		if err != nil {
			return nil, err
		}

		bytes, err := goEncrypt.AesCTR_Decrypt(request, []byte(encodeKey))
		if err != nil {
			return nil, err
		}
		msg.Data = string(bytes)
	}

	xlog.Logger().WithFields(logrus.Fields{
		"clientIp": static.HF_GetHttpIP(req),
		"head":     msg.Header,
		"data":     msg.Data,
	}).Infoln("【RECEIVED HTTP】")

	return &msg, nil
}

// 响应处理函数
// return []byte:打包好的返回数据， int: http status code.
func _GetResponse(header string, data interface{}, err *xerrors.XError, encode int, key string) []byte {
	resp := new(static.Msg_Header)
	resp.Header = header
	if err == nil {
		resp.ErrCode = xerrors.SuccessCode
	} else {
		resp.ErrCode = err.Code
		resp.ErrMsg = err.Msg
	}
	resp.Sign.Encode = encode
	resp.Sign.Time = time.Now().Unix()
	if data != nil {
		kind := reflect.TypeOf(data).Kind()
		switch kind {
		case reflect.String:
			resp.Data = data.(string)
		default:
			resp.Data = static.HF_JtoA(data)
		}
		// log
		xlog.Logger().WithFields(map[string]interface{}{
			"head":    resp.Header,
			"data":    resp.Data,
			"errCode": resp.ErrCode,
			"errMsg":  resp.ErrMsg,
		}).Info("【SEND HTTP RESP】")
		if resp.Sign.Encode == consts.EncodeAes {
			datas, err := goEncrypt.AesCTR_Encrypt([]byte(resp.Data), []byte(key))
			if err != nil {
				xlog.Logger().Errorln("encode http msg filed:", err)
			} else {
				resp.Data = base64.URLEncoding.EncodeToString(datas)
			}
			// syslog.LogField("encode_data", resp.Data).Info("【Server Encode】")
		}
	} else {
		resp.Data = ""
		xlog.WithFields(map[string]interface{}{
			"head":    resp.Header,
			"data":    resp.Data,
			"errCode": resp.ErrCode,
			"errMsg":  resp.ErrMsg,
		}).Warningln("【SEND NULL HTTP RESP】")
	}

	return static.HF_JtoB(resp)
}

// func errResp(header string, err error, encode int, key string) []byte {
// 	resp := new(public.Msg_Header)
// 	resp.Header = header
// 	if customError, ok := err.(*xerrors.XError); ok {
// 		resp.ErrCode = customError.Code
// 	} else {
// 		resp.ErrCode = xerrors.ResultErrorCode
// 	}
// 	resp.Sign.Encode = encode
// 	resp.Sign.Time = time.Now().Unix()
// 	switch resp.Sign.Encode {
// 	case constant.EncodeNone: // 不加密
// 		resp.Data = err.Error()
// 	case constant.EncodeAes: // aes + base64
// 		bytes, _ := goEncrypt.AesCTR_Encrypt([]byte(err.Error()), []byte(key))
// 		resp.Data = base64.URLEncoding.EncodeToString(bytes)
// 	}
// 	return public.HF_JtoB(resp)
// }
//
// func commonResp(header string, v interface{}, encode int, key string) []byte {
// 	resp := new(public.Msg_Header)
// 	resp.Header = header
// 	resp.ErrCode = xerrors.SuccessCode
// 	resp.Sign.Encode = encode
// 	resp.Sign.Time = time.Now().Unix()
// 	switch resp.Sign.Encode {
// 	case constant.EncodeNone: // 不加密
// 		resp.Data = public.HF_JtoA(v)
// 	case constant.EncodeAes: // aes + base64
// 		var bytes []byte
// 		if reflect.TypeOf(v).Kind() == reflect.String {
// 			bytes, _ = goEncrypt.AesCTR_Encrypt([]byte(v.(string)), []byte(key))
// 		} else {
// 			bytes, _ = goEncrypt.AesCTR_Encrypt([]byte(public.HF_JtoA(v)), []byte(key))
// 		}
// 		resp.Data = base64.URLEncoding.EncodeToString(bytes)
// 	}
// 	return public.HF_JtoB(resp)
// }
