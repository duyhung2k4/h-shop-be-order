package request

import "app/model"

type OrderRequest struct {
	ProductId         string `json:"productId"`
	WarehouseId       uint   `json:"warehouseId"`
	TypeInWarehouseId *uint  `json:"typeWarehouseId"`
	GroupOrderId      uint   `json:"groupOrderId"`
	Amount            uint   `json:"amount"`
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
