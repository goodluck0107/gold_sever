package center

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/jinzhu/gorm"
	"github.com/open-source/game/chess.git/pkg/consts"
	"github.com/open-source/game/chess.git/pkg/models"
	"github.com/open-source/game/chess.git/pkg/xlog"
	"net/http"
	"time"
)

const (
	CallbackURL = "http://8.134.80.223:9032/callback/v1/payment"
	MchId       = "1720306540"
	AppId       = "wx7cf88f32b3c3ad62"
	SerialNo    = "4C873778AA0BE3C7CF36634353B111735EC45B42"
	APIv3Key    = "ggoldzyyl250702004533070120akaZb"
	PrivateKey  = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDTB+juDUFBxW3Z
MVN/TX4NZl0gl4QeU8CBPFNVhSFMP+iExeR3q2vQS8foaPD57FWw9MC2oaoRxTJe
YtEJIWrc63kQ53gqMOMMQz/R/6FsbkEv1sy9lel0DXjy3AGvbCvyde8XUFpVUq5h
bHaCaC7UauBgJT/RcKMwWwPup9OdX6ESyH9gzWCZpmnU+iXlhgKTEpEWV0d6TMMK
7lMYq+CR+8uhRVeuSg4iQULNrDdlckHKM2zy1gqbHZIn6c1ipBNKOI+gc/NDMIgx
XN4QMwUnDHxSd7K5veH9fpdF2zRdljw5wuE4Wq05vb25olijQyz8hpU+hmrrPmfi
ksrsuAz5AgMBAAECggEAL33/ONu4zPF+mDuWm/a3oJJh8PzIWf7lOvB1nLB6Xuuf
C4pxsVDd0aHMWoyLduNuIYXjfPaDDL7BaCZ6/sALia5gr6I4E96uqkrUKRGLpZhm
iTqhFyWOxXRmvNkwc+c3OLi71xZZTgxufDRps015BIsM9fZMF3lu8Gn7R5FzeV3s
hs8cBF/pc5pHY0nWJd3DttqCsRSAG3sHe06Iz1kKsBSLzem+vGHot93835Q7J1iT
uGKUGacrNedTRpBfVslxOcx3A0NOKGS9FloWvBMv1LVH64X4R2/nbEYUIECEE094
lTMp6ckDeas+PWY1zY0RKxin7h9MOPgEi7yazUXmEQKBgQDo+B9q5Doaaq4z9oZS
wH2gwACfHwvP3SKWRorxONzfL4ofu7CYcV4bdUWFkL/ehTGmUdATeTt3a5wHvDoG
2VHpFzARaFNRv/wrNK1W0R+XaYBh/+tHYGVyHlmT4rsCimX08IrcnRUEvDNv6fmM
n2vvVI3woBTks9xqJJwbwZQ5LQKBgQDn5JUAOPQGk/4crIeNX84w5sQ6WizG0JL6
Td+jzHDOyVU5pPN0GMpTUvCVRabsYlBZW5Hx5rBGERHYN5YxXdzQLmMH724s19/T
ZUs8yBOwz2NxCr57DNUHXIYMxYv6Aev65zC1bJ4SHSk+nNAFPttyNeu2J0hvHAGm
eYoN2onqfQKBgQDkesPFP4PEeK/UgoiGDAapauSxKe+ZstTC8Pg/T3c+5A7gxGCT
gUu8Pi0qqyWhhJuG9GHPV2x82Gq0I2P9Z5EvuvAHgnuEh3c2oHkH1hzXkD663hTP
cbjMTPupUAn8meMYb/igGOaOOE1yCtQVmBxxIkn6neUfz03yQ2lex2EpGQKBgQC/
W4h9e4Ib34olnVXqmvGqtvOc94bVtY5kEVkIcQ9yBQBYJj9kQYTMh7fSZnzdui91
3bOsu+Ign7trAkvlhwBNpsm/5Zu0U5v3dTJGAREGqcz0npobLraocXiJF4dwEp/q
F1fBjtVOO1Qqv/qFKZ6rO8W8NeR3E9RkzQzYa8u9fQKBgBHNJT4agqUKgqeTQ6k5
ZuhUQG0x+IUuXMLpsfRp8QT3gervyQtRwN0wMU/HmbPWZsxqvZWARUMWJ1CPuX+z
Ayp3+K2M2J9554pLknrrlEwBvL95tdjoDF7AGFISLb+2ZnjpCAqhYzWe3KdT6rzS
jjF/9zrUbgnlp4ZV+pOlOeLp
-----END PRIVATE KEY-----`
	PublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9kURbkO/nyiaqFri02zp
+erb2SK5vehBTA8u+7vFWX6eHwftrwb27Kb7jP/N6iHpqZSxHCgMt+wpm5G23bHh
mD9xe+3cn1Lcx8675lRQ93k3DtYxwmeQScXAiIHbqegkav1T4pO1Cn0nz0IXlwJ7
IEluAHy2WlabHUgwHbxsQRDci2TVV74djbAGf85gkcu4R5UozH7jNlaDfgWxlACX
FjbElWRZB3LvH/7HpPUtgUiMcIjfHW0/hbOYjPK/+AFmOAILfQloatz06F+/2/ai
3byBX+FMnGwixxcBzfoT1+o4P6EZhJUUbDpXq9D6MJ4cpZRgaz5wU/wtb7qhDSZs
hQIDAQAB
-----END PUBLIC KEY-----`
	PublicKeyID = "PUB_KEY_ID_0117203065402025070200181644000260"
)

var (
	payClient *wechat.ClientV3
)

func init() {
	var err error
	payClient, err = GetWechatClient()
	if err != nil {
		xlog.Logger().Error(err)
		panic(err)
	}
	http.HandleFunc("/callback/v1/payment", Callback)
}

func GetWechatClient() (*wechat.ClientV3, error) {
	wxCli, err := wechat.NewClientV3(MchId, SerialNo, APIv3Key, PrivateKey)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}
	err = wxCli.AutoVerifySignByPublicKey([]byte(PublicKey), PublicKeyID)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}

	//// 自定义配置http请求接收返回结果body大小，默认 10MB
	//client.SetBodySize() // 没有特殊需求，可忽略此配置
	//
	//// 设置自定义RequestId生成方法，非必须
	//client.SetRequestIdFunc()

	//client.V3EncryptText()
	//
	//client.V3DecryptText()

	// 打开Debug开关，输出日志，默认是关闭的
	wxCli.DebugSwitch = gopay.DebugOn

	return wxCli, nil
}

func CreateOrder(ctx context.Context, tradeNo string) (*wechat.AppPayParams, error) {
	var paymentOrder models.PaymentOrder
	err := GetDBMgr().GetDBmControl().First(&paymentOrder, "trade_no = ?", tradeNo).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = fmt.Errorf("trade_no: %s not found", tradeNo)
			xlog.Logger().Error(err)
			return nil, err
		}
		xlog.Logger().Error(err)
		return nil, err
	}

	expire := time.Now().Add(30 * time.Minute).Format(time.RFC3339)
	// 初始化 BodyMap
	bm := make(gopay.BodyMap)
	bm.Set("appid", AppId).
		Set("mchid", MchId).
		Set("description", paymentOrder.Remark).
		Set("out_trade_no", tradeNo).
		Set("time_expire", expire).
		Set("notify_url", CallbackURL).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", paymentOrder.Price).
				Set("currency", "CNY")
		})
	//.
	//SetBodyMap("payer", func(bm gopay.BodyMap) {
	//	bm.Set("sp_openid", "asdas")
	//})

	wxRsp, err := payClient.V3TransactionApp(ctx, bm)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}
	if wxRsp.Code != wechat.Success {
		err = fmt.Errorf("code:%d, msg:%s, resp:%+v", wxRsp.Code, wxRsp.Error, wxRsp.Response)
		xlog.Logger().Error(err)
		return nil, err
	}

	xlog.Logger().Infof("order: %s => wechat prepay_id: %s", tradeNo, wxRsp.Response.PrepayId)
	app, err := payClient.PaySignOfApp(AppId, wxRsp.Response.PrepayId)
	if err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}

	appStr, _ := json.Marshal(app)

	paymentOrder.PrepayId = wxRsp.Response.PrepayId
	paymentOrder.Status = models.OrderStatusPending
	paymentOrder.Params = string(appStr)
	err = GetDBMgr().GetDBmControl().Save(&paymentOrder).Error
	if err != nil {
		xlog.Logger().Error(err)
		return nil, err
	}
	return app, nil
}

type Notify struct {
	Appid         string `json:"appid"`
	Mchid         string `json:"mchid"`
	OutTradeNo    string `json:"out_trade_no"`
	TransactionId string `json:"transaction_id"`
	//JSAPI：公众号支付、小程序支付
	//NATIVE：Native支付
	//APP：APP支付
	//MICROPAY：付款码支付
	//MWEB：H5支付
	//FACEPAY：刷脸支付
	TradeType string `json:"trade_type"`
	//SUCCESS：支付成功
	//REFUND：转入退款
	//NOTPAY：未支付
	//CLOSED：已关闭
	//REVOKED：已撤销（仅付款码支付会返回）
	//USERPAYING：用户支付中（仅付款码支付会返回）
	//PAYERROR：支付失败（仅付款码支付会返回）
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	BankType       string `json:"bank_type"`
	Attach         string `json:"attach"`
	SuccessTime    string `json:"success_time"`
	Payer          struct {
		Openid    string `json:"openid"`
		SubOpenid string `json:"sub_openid"`
	} `json:"payer"`
	Amount struct {
		Total       int    `json:"total"`
		Currency    string `json:"currency"`
		PayerTotal  int    `json:"payer_total"`
		DealerTotal int    `json:"dealer_total"`
		CashTotal   int    `json:"cash_total"`
		Discount    int    `json:"discount"`
	} `json:"amount"`
}

func Callback(w http.ResponseWriter, r *http.Request) {
	var resp wechat.V3NotifyRsp
	err := handleNotify(r)
	if err != nil {
		xlog.Logger().Error(err)
		resp.Code = gopay.FAIL
		resp.Message = err.Error()
	} else {
		resp.Code = gopay.SUCCESS
		resp.Message = "成功"
	}
	// 设置响应头为 JSON 类型
	w.Header().Set("Content-Type", "application/json")
	// 设置状态码为 200
	w.WriteHeader(http.StatusOK)
	// 编码为 JSON 并写入响应
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		xlog.Logger().Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	xlog.Logger().Infof("callback resp: %+v", resp)
}

func handleNotify(r *http.Request) error {
	notifyReq, err := wechat.V3ParseNotify(r)
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}
	xlog.Logger().Infof("notifyReq: %+v", notifyReq)
	// 获取微信平台证书
	certMap := payClient.WxPublicKeyMap()
	// 验证异步通知的签名
	err = notifyReq.VerifySignByPKMap(certMap)
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}
	ntf, suc, err := ParseNotify(notifyReq)
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}
	var paymentOrder models.PaymentOrder
	err = GetDBMgr().GetDBmControl().First(&paymentOrder, "trade_no = ?", ntf.OutTradeNo).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = fmt.Errorf("trade_no: %s not found", ntf.OutTradeNo)
			xlog.Logger().Error(err)
			return err
		}
		xlog.Logger().Error(err)
		return err
	}
	if suc {
		paymentOrder.Status = models.OrderStatusPayCompleted
	} else {
		paymentOrder.Status = models.OrderStatusPayFailed
	}
	paymentOrder.Openid = ntf.Payer.Openid
	paymentOrder.SuccessTime = ntf.SuccessTime
	paymentOrder.BankType = ntf.BankType
	paymentOrder.TransactionId = ntf.TransactionId
	paymentOrder.TradeState = ntf.TradeState
	paymentOrder.TradeStateDesc = ntf.TradeStateDesc
	paymentOrder.TradeType = ntf.TradeType
	err = GetDBMgr().GetDBmControl().Save(&paymentOrder).Error
	if err != nil {
		xlog.Logger().Error(err)
		return err
	}
	xlog.Logger().Infof("order: %s => wechat notify success", ntf.OutTradeNo)
	if person := GetPlayerMgr().GetPlayer(paymentOrder.UserId); person != nil {
		xlog.Logger().Warnf("玩家 %d 充值成功， 在线状态， 发送消息给玩家:%v", paymentOrder.UserId, paymentOrder)
		person.SendMsg(consts.MsgTypePaymentResultNtf, paymentOrder)
	} else {
		xlog.Logger().Warnf("玩家 %d 充值成功， 不在线， 不发送消息给玩家:%v", paymentOrder.UserId, paymentOrder)
		GetDBMgr().GetDBrControl().RedisV2.Set(fmt.Sprintf("offline:recharge_suc:%d", paymentOrder.UserId),
			paymentOrder, time.Hour*48)
	}
	if suc {
		_, err := updateWealth(paymentOrder.UserId, paymentOrder.WealthType, paymentOrder.Num+paymentOrder.Gift, models.CostTypeBuy)
		if err != nil {
			xlog.Logger().Error(err)
			paymentOrder.Status = models.OrderStatusAmountFailed
			err = GetDBMgr().GetDBmControl().Model(&models.PaymentOrder{}).Where("id=?", paymentOrder.ID).UpdateColumn("status", paymentOrder.Status).Error
			if err != nil {
				xlog.Logger().Error(err)
			}
			return err
		} else {
			paymentOrder.Status = models.OrderStatusAmountArrived
			err = GetDBMgr().GetDBmControl().Model(&models.PaymentOrder{}).Where("id=?", paymentOrder.ID).UpdateColumn("status", paymentOrder.Status).Error
			if err != nil {
				xlog.Logger().Error(err)
			}
		}
	}
	return nil
}

func ParseNotify(notifyReq *wechat.V3NotifyReq) (*Notify, bool, error) {
	// 解密回调数据
	var notify Notify
	err := notifyReq.DecryptCipherTextToStruct(APIv3Key, &notify)
	if err != nil {
		xlog.Logger().Error(err)
		return &notify, false, err
	}
	suc := notifyReq.EventType == "TRANSACTION.SUCCESS" && notify.TradeState == "SUCCESS"
	return &notify, suc, nil
}
