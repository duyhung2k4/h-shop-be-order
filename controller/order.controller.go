package controller

import (
	"app/config"
	"app/dto/request"
	"app/dto/response"
	"app/grpc/proto"
	"app/service"
	"app/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/render"
)

type orderController struct {
	grpcWarehouseService       proto.WarehouseServiceClient
	grpcTypeInWarehouseService proto.TypeInWarehouseServiceClient
	countPriceService          proto.CountPriceServiceClient
	grpcBillService            proto.BillServiceClient
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

	if len(groupOrder.Orders) == 0 {
		badRequest(w, r, errors.New("orders empty"))
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
			GroupOrderId:      newGroupOrder.ID,
			Amount:            item.Amount,
		})
	}
	errCheckCount := c.orderService.CheckCount(payloadCheckCount)
	if errCheckCount != nil {
		// push mess delete group order
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
				ProductId:         item.ProductId,
				Amount:            uint64(item.Amount),
				WarehouseId:       uint64(item.WarehouseId),
				TypeInWarehouseId: uint64(*item.TypeInWarehouseId),
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

	// push mess update total group order
	ipAddr := strings.Join([]string{
		r.Header.Get("x-forwarded-for"),
		r.RemoteAddr,
	}, ",")
	expireDate := time.Now().Add(5 * time.Minute)
	billResult, errCreateBill := c.grpcBillService.CreateBill(context.Background(), &proto.CreateBillReq{
		Amount:           float32(newGroupOrder.Total),
		ExpireDate:       expireDate.Unix(),
		OrderDescription: groupOrder.OrderDescription,
		OrderType:        groupOrder.OrderType,
		IpAddr:           ipAddr,
	})

	if errCreateBill != nil {
		internalServerError(w, r, errCreateBill)
		return
	}

	res := Response{
		Data: response.OrderResponse{
			GroupOrder: *newGroupOrder,
			VnpHref:    billResult.HrefVnp,
		},
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
		grpcBillService:            proto.NewBillServiceClient(config.GetConnPaymentGRPC()),
		orderService:               service.NewOrderService(),
		groupOrderService:          service.NewGroupOrderService(),
		jwtUtils:                   utils.NewJwtUtils(),
	}
}
