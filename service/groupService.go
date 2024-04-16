package service

import (
	"app/config"
	"app/model"

	"gorm.io/gorm"
)

type groupOrderService struct {
	db *gorm.DB
}

type GroupOrderService interface {
	CreateGroupOrder(payload PayloadCreateGroupOrder) (*model.GroupOrder, error)
}

func (s *groupOrderService) CreateGroupOrder(payload PayloadCreateGroupOrder) (*model.GroupOrder, error) {
	var groupOrder = &model.GroupOrder{
		Address: payload.Address,
		TypePay: payload.TypePay,
	}

	if err := s.db.Model(&model.GroupOrder{}).Create(&groupOrder).Error; err != nil {
		return nil, err
	}

	return groupOrder, nil
}

func NewGroupOrderService() GroupOrderService {
	return &groupOrderService{
		db: config.GetDB(),
	}
}

type PayloadCreateGroupOrder struct {
	Address string
	TypePay model.TYPE_PAY
}
