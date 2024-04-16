package model

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	ProductId         uint  `json:"productId"`
	WarehouseId       uint  `json:"warehouseId"`
	TypeInWarehouseId *uint `json:"typeInWarehouseId"`
	Amount            int   `json:"amount"`
	ProfileId         uint  `json:"profileId"`
	GroupOrderId      uint  `json:"groupOrderId"`

	GroupOrder GroupOrder `json:"GroupOrder" gorm:"foreignKey:GroupOrderId"`
}
