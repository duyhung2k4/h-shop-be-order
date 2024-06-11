package service

import (
	"app/config"
	"app/grpc/proto"
	"app/model"
	"context"
	"sync"

	"gorm.io/gorm"
)

type orderService struct {
	db                         *gorm.DB
	grpcWarehouseService       proto.WarehouseServiceClient
	grpcTypeInWarehouseService proto.TypeInWarehouseServiceClient
}

type OrderService interface {
	Order(payload []OrderPayload) ([]model.Order, error)
	CheckCount(payload []CheckcountPayload) error
}

func (s *orderService) Order(payload []OrderPayload) ([]model.Order, error) {
	var newOrders []model.Order

	for _, item := range payload {
		order := &model.Order{
			ProfileId:         item.ProfileId,
			ProductId:         item.ProductId,
			WarehouseId:       item.WarehouseId,
			GroupOrderId:      item.GroupOrderId,
			TypeInWarehouseId: item.TypeInWarehouseId,
			Amount:            item.Amount,
		}
		newOrders = append(newOrders, *order)
	}

	if err := s.db.Model(&model.Order{}).Create(&newOrders).Error; err != nil {
		return []model.Order{}, err
	}

	return newOrders, nil
}

func (s *orderService) CheckCount(payload []CheckcountPayload) error {
	var errCheckout error = nil
	var wg sync.WaitGroup
	var checkCountNotError []CheckcountPayload
	for _, item := range payload {
		wg.Add(1)
		go func(order CheckcountPayload) {
			if order.TypeInWarehouseId == nil {
				_, err := s.grpcWarehouseService.DownCount(context.Background(), &proto.DownCountWarehouseReq{
					Id:     uint64(order.WarehouseId),
					Amount: uint64(order.Amount),
				})

				if err != nil {
					errCheckout = err
				} else {
					checkCountNotError = append(checkCountNotError, order)
				}
			} else {
				_, err := s.grpcTypeInWarehouseService.DownCount(context.Background(), &proto.DownCountTypeInWarehouseReq{
					Id:     uint64(*order.TypeInWarehouseId),
					Amount: uint64(order.Amount),
				})

				if err != nil {
					errCheckout = err
				} else {
					checkCountNotError = append(checkCountNotError, order)
				}
			}
			wg.Done()
		}(item)
	}
	wg.Wait()

	if errCheckout != nil {
		// Push mess up count not error
		return errCheckout
	}

	return nil
}

func NewOrderService() OrderService {
	return &orderService{
		db:                         config.GetDB(),
		grpcWarehouseService:       proto.NewWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		grpcTypeInWarehouseService: proto.NewTypeInWarehouseServiceClient(config.GetConnWarehouseGRPC()),
	}
}

type OrderPayload struct {
	ProfileId         uint
	ProductId         string
	WarehouseId       uint
	GroupOrderId      uint
	TypeInWarehouseId *uint
	Amount            int
}

type CheckcountPayload struct {
	ProductId         string
	WarehouseId       uint
	TypeInWarehouseId *uint
	GroupOrderId      uint
	Amount            uint
}
