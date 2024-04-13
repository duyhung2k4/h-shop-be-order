package model

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	ProductId         uint    `json:"productId"`
	WarehouseId       uint    `json:"warehouseId"`
	TypeInWarehouseId uint    `json:"typeInWarehouseId"`
	ProfileId         uint    `json:"profileId"`
	Address           string  `json:"address"`
	Amount            int     `json:"amount"`
	TypePay           string  `json:"typePay"`
	Paid              bool    `json:"paid"`
	Total             float64 `json:"total"`
}
