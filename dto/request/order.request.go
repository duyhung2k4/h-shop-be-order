package request

import "app/model"

type OrderRequest struct {
	ShopId            uint64  `json:"shopId"`
	ProductId         string  `json:"productId"`
	WarehouseId       uint    `json:"warehouseId"`
	TypeInWarehouseId *uint   `json:"typeWarehouseId"`
	GroupOrderId      uint    `json:"groupOrderId"`
	Total             float64 `json:"total"`
	Amount            uint    `json:"amount"`
}

type GroupOrderRequest struct {
	Address          string         `json:"address"`
	TypePay          model.TYPE_PAY `json:"typePay"`
	OrderDescription string         `json:"orderDescription"`
	OrderType        string         `json:"orderType"`
	Orders           []OrderRequest `json:"orders"`
}

type ChangeStatusOrderRequest struct {
	OrderId string `json:"orderId"`
	Status  string `json:"status"`
}

type ChangeStatusOrderV2Request struct {
	Id                uint64 `json:"id"`
	Status            string `json:"status"`
	Paid              bool   `json:"paid"`
	Amount            int    `json:"amount"`
	WarehouseId       uint   `json:"warehouseId"`
	TypeInWarehouseId *uint  `json:"typeWarehouseId"`
}
