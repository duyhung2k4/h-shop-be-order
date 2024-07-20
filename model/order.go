package model

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	ShopId            uint64  `json:"shopId"`
	ProductId         string  `json:"productId"`
	WarehouseId       uint    `json:"warehouseId"`
	TypeInWarehouseId *uint   `json:"typeInWarehouseId"`
	Amount            int     `json:"amount"`
	ProfileId         uint    `json:"profileId"`
	GroupOrderId      uint    `json:"groupOrderId"`
	Status            string  `json:"status"`
	Paid              bool    `json:"paid"`
	Total             float64 `json:"total"`

	GroupOrder GroupOrder `json:"groupOrder" gorm:"foreignKey:GroupOrderId"`
}
