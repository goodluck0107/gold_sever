package router

import (
	"encoding/base64"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/static"
	"github.com/open-source/game/chess.git/pkg/static/goEncrypt"
	"reflect"
)

type ProtocolHandler func(...interface{}) (int16, interface{})

type ProtocolInfo struct {
	ProtoType    reflect.Type
	ProtoHandler ProtocolHandler
}

type ProtocolWorker struct {
	FuncProtocol map[string]*ProtocolInfo
}

func (pm *ProtocolWorker) RegisterMessage(header string, proto interface{}, handler ProtocolHandler) {
	var info ProtocolInfo
	info.ProtoType = reflect.TypeOf(proto)
	info.ProtoHandler = handler
	if pm.FuncProtocol == nil {
		pm.FuncProtocol = make(map[string]*ProtocolInfo)
	}
	pm.FuncProtocol[header] = &info
}

func (pm *ProtocolWorker) GetProtocol(protocol string) *ProtocolInfo {
	return pm.FuncProtocol[protocol]
}

// 消息解密
func getCommonMsg(msg *static.Msg_Header) error {
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
