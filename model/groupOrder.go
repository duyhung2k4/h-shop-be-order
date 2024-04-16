package model

import "gorm.io/gorm"

type GroupOrder struct {
	gorm.Model
	Address string   `json:"address"`
	TypePay TYPE_PAY `json:"typePay"`
	Paid    bool     `json:"paid"`
	Total   float64  `json:"total"`

	Orders []Order `json:"orders" gorm:"foreignKey:GroupOrderId"`
}

type TYPE_PAY string

var (
	CASH           TYPE_PAY = "cash"
	ONLINE_PAYMENT TYPE_PAY = "online_payment"
)
