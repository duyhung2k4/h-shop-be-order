package controller

import (
	"app/config"
	"app/dto/request"
	"app/grpc/proto"
	"app/service"
	"app/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/render"
)

type orderController struct {
	grpcWarehouseService       proto.WarehouseServiceClient
	grpcTypeInWarehouseService proto.TypeInWarehouseServiceClient
	countPriceService          proto.CountPriceServiceClient
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
	var groupOrder request.GroupOrderRequest
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
	for i, item := range groupOrder.Orders {
		groupOrder.Orders[i].GroupOrderId = newGroupOrder.ID
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

	var errHandle error = nil
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		orders := []service.OrderPayload{}

		for _, item := range groupOrder.Orders {
			order := service.OrderPayload{
				ProfileId:         profileId,
				ProductId:         item.ProductId,
				WarehouseId:       item.WarehouseId,
				GroupOrderId:      item.GroupOrderId,
				TypeInWarehouseId: item.TypeInWarehouseId,
				Amount:            int(item.Amount),
			}
			orders = append(orders, order)
		}

		_, errNewOrders := c.orderService.Order(orders)
		if errNewOrders != nil {
			errHandle = errNewOrders
			wg.Done()
			return
		}
		wg.Done()
	}()

	go func() {
		orders := []*proto.Order{}

		for _, item := range groupOrder.Orders {
			orders = append(orders, &proto.Order{
				ProductId:   item.ProductId,
				Amount:      uint64(item.Amount),
				WarehouseId: uint64(item.WarehouseId),
			})
		}

		res, err := c.countPriceService.CountPriceOrder(context.Background(), &proto.CountPriceOrderReq{
			GroupOrderId: uint64(newGroupOrder.ID),
			Orders:       orders,
		})

		if err != nil {
			errHandle = err
			wg.Done()
			return
		}

		newGroupOrder.Total = float64(res.Price)

		wg.Done()
	}()

	wg.Wait()

	if errHandle != nil {
		internalServerError(w, r, errHandle)
		return
	}

	res := Response{
		Data:    newGroupOrder,
		Message: "OK",
		Status:  200,
		Error:   nil,
	}

	render.JSON(w, r, res)
}

func NewOrderController() OrderController {
	return &orderController{
		grpcWarehouseService:       proto.NewWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		grpcTypeInWarehouseService: proto.NewTypeInWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		countPriceService:          proto.NewCountPriceServiceClient(config.GetConnWarehouseGRPC()),
		orderService:               service.NewOrderService(),
		groupOrderService:          service.NewGroupOrderService(),
		jwtUtils:                   utils.NewJwtUtils(),
	}
}
