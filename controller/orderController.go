package controller

import (
	"app/config"
	"app/grpc/proto"
	"app/model"
	"app/service"
	"app/utils"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type orderController struct {
	grpcWarehouseService       proto.WarehouseServiceClient
	grpcTypeInWarehouseService proto.TypeInWarehouseServiceClient
	orderService               service.OrderService
	groupOrderService          service.GroupOrderService
	jwtUtils                   utils.JwtUtils
}

type OrderController interface {
	Order(w http.ResponseWriter, r *http.Request)
}

func (c *orderController) Order(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.Split(r.Header.Get("Authorization"), " ")[1]
	mapDataRequest, errMapData := c.jwtUtils.JwtDecode(tokenString)

	if errMapData != nil {
		internalServerError(w, r, errMapData)
		return
	}

	// profileId
	if mapDataRequest["profile_id"] == nil {
		handleError(w, r, errors.New("not permission"), 401)
		return
	}
	profileId := uint(mapDataRequest["profile_id"].(float64))

	// create groupOrder
	var groupOrder GroupOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&groupOrder); err != nil {
		badRequest(w, r, err)
		return
	}
	newGroupOrder, errGroupOrderId := c.groupOrderService.CreateGroupOrder(service.PayloadCreateGroupOrder{
		Address: groupOrder.Address,
		TypePay: groupOrder.TypePay,
	})
	if errGroupOrderId != nil {
		internalServerError(w, r, errGroupOrderId)
		return
	}

	// Check count
	payloadCheckCount := []service.CheckcountPayload{}
	for _, item := range groupOrder.Orders {
		item.GroupOrderId = newGroupOrder.ID
		payloadCheckCount = append(payloadCheckCount, service.CheckcountPayload{
			ProductId:         item.ProductId,
			WarehouseId:       item.WarehouseId,
			TypeInWarehouseId: item.TypeInWarehouseId,
			GroupOrderId:      item.GroupOrderId,
			Amount:            item.Amount,
		})
	}
	errCheckCount := c.orderService.CheckCount(payloadCheckCount)
	if errCheckCount != nil {
		internalServerError(w, r, errCheckCount)
		return
	}

	orders := []service.OrderPayload{}
	for _, item := range groupOrder.Orders {
		order := service.OrderPayload{
			ProfileId:         profileId,
			ProductId:         item.ProductId,
			WarehouseId:       item.WarehouseId,
			GroupOrderId:      item.GroupOrderId,
			TypeInWarehouseId: item.TypeInWarehouseId,
		}
		orders = append(orders, order)
	}
	_, errNewOrders := c.orderService.Order(orders)
	if errNewOrders != nil {
		internalServerError(w, r, errNewOrders)
		return
	}
}

func NewOrderController() OrderController {
	return &orderController{
		grpcWarehouseService:       proto.NewWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		grpcTypeInWarehouseService: proto.NewTypeInWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		orderService:               service.NewOrderService(),
		groupOrderService:          service.NewGroupOrderService(),
		jwtUtils:                   utils.NewJwtUtils(),
	}
}

type OrderRequest struct {
	ProductId         uint  `json:"productId"`
	WarehouseId       uint  `json:"warehouseId"`
	TypeInWarehouseId *uint `json:"typeWarehouseId"`
	GroupOrderId      uint  `json:"groupOrderId"`
	Amount            uint  `json:"amount"`
}

type GroupOrderRequest struct {
	Address string         `json:"address"`
	TypePay model.TYPE_PAY `json:"typePay"`
	Orders  []OrderRequest `json:"orders"`
}

type OrderResponse struct {
}
