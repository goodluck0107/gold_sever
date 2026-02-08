package models

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
)

type OrderStatus string

const (
	OrderStatusInit          OrderStatus = "init"
	OrderStatusPending       OrderStatus = "pending"
	OrderStatusPayFailed     OrderStatus = "pay_failed"
	OrderStatusPayCompleted  OrderStatus = "pay_completed"
	OrderStatusAmountFailed  OrderStatus = "amount_failed"
	OrderStatusAmountArrived OrderStatus = "amount_arrived"
)

type PaymentOrder struct {
	gorm.Model
	TradeNo    string `json:"tradeNo" gorm:"unique_index"`
	UserId     int64  `json:"userId" gorm:"index"`
	WealthType int8   `json:"wealthType"`
	ShopId     int    `json:"shopId"`
	Price      int    `json:"price"`
	Num        int    `json:"num"`
	Gift       int    `json:"gift"`
	Currency   string `json:"currency"`
	Remark     string `json:"remark"`

	Status OrderStatus `json:"status"`

	PrepayId string `json:"prepay_id"`
	Params   string `json:"params" gorm:"-"`

	Openid         string `json:"openid"`
	SuccessTime    string `json:"success_time"`
	BankType       string `json:"bank_type"`
	TransactionId  string `json:"transaction_id"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	TradeType      string `json:"trade_type"`
}

func (po *PaymentOrder) MarshalBinary() (data []byte, err error) {
	return json.Marshal(po)
}

func (po *PaymentOrder) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, po)
}

func (po *PaymentOrder) TableName() string {
	return "payment_order"
}

func initPaymentOrder(db *gorm.DB) error {
	var err error
	if db.HasTable(&PaymentOrder{}) {
		err = db.AutoMigrate(&PaymentOrder{}).Error
	} else {
		err = db.CreateTable(&PaymentOrder{}).Error
	}
	return err
}
