package protocolworker

import (
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

func (self *ProtocolWorker) RegisterMessage(header string, proto interface{}, handler ProtocolHandler) {
	var info ProtocolInfo
	info.ProtoType = reflect.TypeOf(proto)
	info.ProtoHandler = handler
	if self.FuncProtocol == nil {
		self.FuncProtocol = make(map[string]*ProtocolInfo)
	}
	self.FuncProtocol[header] = &info
}

func (self *ProtocolWorker) GetProtocol(protocol string) *ProtocolInfo {
	return self.FuncProtocol[protocol]
}

//// 消息解密
//func getCommonMsg(msg *public.Msg_Header) error {
//	switch msg.Sign.Encode {
//	case constant.EncodeNone: // 不加密不处理
//	case constant.EncodeAes: // aes + base64
//		data, err := base64.URLEncoding.DecodeString(msg.Data)
//		if err != nil {
//			return err
//		}
//
//		bytes, err := goEncrypt.AesCTR_Decrypt(data, []byte(wuhan.GetServer().Con.EncodePhpKey))
//		if err != nil {
//			return err
//		}
//		msg.Data = string(bytes)
//	}
//	return nil
//}
