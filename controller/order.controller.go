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
	"log"
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
	ChangeStatusOrder(w http.ResponseWriter, r *http.Request)
	ChangeStatusOrderV2(w http.ResponseWriter, r *http.Request)
	GetPurchaseOrder(w http.ResponseWriter, r *http.Request)
	AdminGetOrder(w http.ResponseWriter, r *http.Request)
}

func (c *orderController) Order(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.Split(r.Header.Get("Authorization"), " ")[1]
	mapDataRequest, errMapData := c.jwtUtils.JwtDecode(tokenString)

	log.Println(tokenString)

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
	}, profileId)
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
				ShopId:            item.ShopId,
				ProductId:         item.ProductId,
				WarehouseId:       item.WarehouseId,
				GroupOrderId:      item.GroupOrderId,
				TypeInWarehouseId: item.TypeInWarehouseId,
				Amount:            int(item.Amount),
				Status:            "pending",
				Total:             item.Total,
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
			order := proto.Order{
				ProductId:   item.ProductId,
				Amount:      uint64(item.Amount),
				WarehouseId: uint64(item.WarehouseId),
			}

			if item.TypeInWarehouseId != nil {
				order.TypeInWarehouseId = uint64(*item.TypeInWarehouseId)
			}
			orders = append(orders, &order)
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

	if groupOrder.TypePay == "offline" {
		res := Response{
			Data:    newGroupOrder,
			Message: "OK",
			Status:  200,
			Error:   nil,
		}

		render.JSON(w, r, res)
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

	_, errUpdateGroupOrder := c.groupOrderService.UpdateGroupOrder(newGroupOrder.ID, billResult.UuidBill, newGroupOrder.Total)
	if errUpdateGroupOrder != nil {
		internalServerError(w, r, errUpdateGroupOrder)
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

func (c *orderController) ChangeStatusOrder(w http.ResponseWriter, r *http.Request) {
	var payload request.ChangeStatusOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, r, err)
		return
	}

	if err := c.groupOrderService.ChangeStatusOrder(payload.OrderId, payload.Status); err != nil {
		internalServerError(w, r, err)
		return
	}

	res := Response{
		Data:    nil,
		Message: "OK",
		Status:  200,
		Error:   nil,
	}

	render.JSON(w, r, res)
}

func (c *orderController) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
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

	data, err := c.groupOrderService.GetPurchaseOrder(profileId)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	res := Response{
		Data:    data,
		Message: "OK",
		Status:  200,
		Error:   nil,
	}

	render.JSON(w, r, res)
}

func (c *orderController) AdminGetOrder(w http.ResponseWriter, r *http.Request) {
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

	orders, errOrder := c.orderService.AdminGetOrder(uint64(profileId))
	if errOrder != nil {
		internalServerError(w, r, errOrder)
		return
	}

	res := Response{
		Data:    orders,
		Message: "OK",
		Status:  200,
		Error:   nil,
	}

	render.JSON(w, r, res)
}

func (c *orderController) ChangeStatusOrderV2(w http.ResponseWriter, r *http.Request) {
	var payload request.ChangeStatusOrderV2Request
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, r, err)
		return
	}

	err := c.orderService.ChangeStatusOrderV2(payload)

	if err != nil {
		internalServerError(w, r, err)
		return
	}

	res := Response{
		Data:    nil,
		Status:  200,
		Message: "OK",
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
