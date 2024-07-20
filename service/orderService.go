package service

import (
	"app/config"
	"app/dto/request"
	"app/grpc/proto"
	"app/model"
	"app/utils"
	"context"
	"sync"

	"gorm.io/gorm"
)

type orderService struct {
	db                         *gorm.DB
	grpcWarehouseService       proto.WarehouseServiceClient
	grpcTypeInWarehouseService proto.TypeInWarehouseServiceClient
	queueProductUtils          utils.QueueProductUtils
}

type OrderService interface {
	AdminGetOrder(profileId uint64) ([]*model.Order, error)
	Order(payload []OrderPayload) ([]model.Order, error)
	CheckCount(payload []CheckcountPayload) error
	ChangeStatusOrderV2(payload request.ChangeStatusOrderV2Request) error
}

func (s *orderService) AdminGetOrder(profileId uint64) ([]*model.Order, error) {
	var orders []*model.Order

	if err := s.db.
		Model(&model.Order{}).
		Preload("GroupOrder").
		Where("shop_id = ?", profileId).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *orderService) Order(payload []OrderPayload) ([]model.Order, error) {
	var newOrders []model.Order

	for _, item := range payload {
		order := &model.Order{
			ProfileId:         item.ProfileId,
			ShopId:            item.ShopId,
			ProductId:         item.ProductId,
			WarehouseId:       item.WarehouseId,
			GroupOrderId:      item.GroupOrderId,
			TypeInWarehouseId: item.TypeInWarehouseId,
			Amount:            item.Amount,
			Status:            item.Status,
			Total:             item.Total,
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
		s.queueProductUtils.PushMessInQueue(checkCountNotError, string(model.UP_COUNT_WAREHOUSE))
		return errCheckout
	}

	return nil
}

func (s *orderService) ChangeStatusOrderV2(payload request.ChangeStatusOrderV2Request) error {
	var newOrder *model.Order = &model.Order{
		Status: payload.Status,
		Paid:   payload.Paid,
	}

	if err := s.db.Model(&model.Order{}).Where("id = ?", payload.Id).Updates(&newOrder).Error; err != nil {
		return err
	}

	if payload.Status == "accept" {
		return nil
	}

	var errRollback error
	if payload.TypeInWarehouseId != nil {
		_, err := s.grpcTypeInWarehouseService.UpCount(context.Background(), &proto.UpCountTypeInWarehouseReq{
			Id:     uint64(*payload.TypeInWarehouseId),
			Amount: uint64(payload.Amount),
		})
		errRollback = err
	} else {
		_, err := s.grpcWarehouseService.UpCount(context.Background(), &proto.UpCountWarehouseReq{
			Id:     uint64(payload.WarehouseId),
			Amount: uint64(payload.Amount),
		})
		errRollback = err
	}

	return errRollback
}

func NewOrderService() OrderService {
	return &orderService{
		db:                         config.GetDB(),
		grpcWarehouseService:       proto.NewWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		grpcTypeInWarehouseService: proto.NewTypeInWarehouseServiceClient(config.GetConnWarehouseGRPC()),
		queueProductUtils:          utils.NewQueueProductService(),
	}
}

type OrderPayload struct {
	ShopId            uint64
	ProfileId         uint
	ProductId         string
	WarehouseId       uint
	GroupOrderId      uint
	TypeInWarehouseId *uint
	Amount            int
	Status            string
	Total             float64
}

type CheckcountPayload struct {
	ProductId         string
	WarehouseId       uint
	TypeInWarehouseId *uint
	GroupOrderId      uint
	Amount            uint
}
