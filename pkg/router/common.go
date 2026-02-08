package router

import (
	"encoding/base64"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	xerrors "github.com/open-source/game/chess.git/pkg/xerrors"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

// 消息解密
func GetCommonMsg(msg *static.Msg_Header) error {
	switch msg.Sign.Encode {
	case consts.EncodeNone: // 不加密不处理
	case consts.EncodeAes: // aes + base64
		data, err := base64.URLEncoding.DecodeString(msg.Data)
		if err != nil {
			return err
		}

		bytes, err := goEncrypt.AesCTR_Decrypt(data, []byte(httpWorker.EncodePhpKey))
		if err != nil {
			return err
		}
		msg.Data = string(bytes)
	}
	return nil
}

// 获取错误返回结果
func GetErrReturn(msgType string, err *xerrors.XError) []byte {
	resultHead := new(static.Msg_Header)
	resultHead.Header = msgType
	//resultHead.Data = err.Error()
	resultHead.ErrCode = err.Code
	resultHead.Sign.Encode = httpWorker.Encode
	resultHead.Sign.Time = time.Now().Unix()

	xlog.Logger().WithFields(logrus.Fields{
		"errhead": msgType,
		"errmsg":  err.Error(),
		"errcode": err.Code,
	}).Infoln("【SEND ERROR HTTP RESP】")

	switch resultHead.Sign.Encode {
	case consts.EncodeNone: // 不加密
		resultHead.Data = err.Error()
	case consts.EncodeAes: // aes + base64
		bytes, _ := goEncrypt.AesCTR_Encrypt([]byte(err.Error()), []byte(httpWorker.EncodePhpKey))
		resultHead.Data = base64.URLEncoding.EncodeToString(bytes)
	}

	return static.HF_JtoB(resultHead)
}

// 获取通用返回结果
func GetCommonResp(msgType string, v interface{}) *static.Msg_Header {
	resp := new(static.Msg_Header)
	resp.Header = msgType
	resp.ErrCode = xerrors.SuccessCode
	resp.Sign.Encode = httpWorker.Encode
	resp.Sign.Time = time.Now().Unix()

	xlog.Logger().WithFields(logrus.Fields{
		"head": msgType,
		"data": static.HF_JtoA(v),
	}).Infoln("【SEND HTTP RESP】")

	switch resp.Sign.Encode {
	case consts.EncodeNone: // 不加密
		resp.Data = static.HF_JtoA(v)
	case consts.EncodeAes: // aes + base64
		var bytes []byte
		if reflect.TypeOf(v).Kind() == reflect.String {
			bytes, _ = goEncrypt.AesCTR_Encrypt([]byte(v.(string)), []byte(httpWorker.EncodePhpKey))
		} else {
			bytes, _ = goEncrypt.AesCTR_Encrypt([]byte(static.HF_JtoA(v)), []byte(httpWorker.EncodePhpKey))
		}
		resp.Data = base64.URLEncoding.EncodeToString(bytes)
	}

	return resp
}
