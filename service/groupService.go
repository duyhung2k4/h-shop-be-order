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
	CreateGroupOrder(payload PayloadCreateGroupOrder, profileId uint) (*model.GroupOrder, error)
	UpdateGroupOrder(id uint, orderId string, total float64) (*model.GroupOrder, error)
	ChangeStatusOrder(orderId string, status string) error
	GetPurchaseOrder(profileId uint) ([]*model.GroupOrder, error)
}

func (s *groupOrderService) CreateGroupOrder(payload PayloadCreateGroupOrder, profileId uint) (*model.GroupOrder, error) {
	var groupOrder = &model.GroupOrder{
		Address:   payload.Address,
		TypePay:   payload.TypePay,
		ProfileId: profileId,
	}

	if err := s.db.Model(&model.GroupOrder{}).Create(&groupOrder).Error; err != nil {
		return nil, err
	}

	return groupOrder, nil
}

func (s *groupOrderService) UpdateGroupOrder(id uint, orderId string, total float64) (*model.GroupOrder, error) {
	var groupOrder = model.GroupOrder{
		VnpTxnRef: &orderId,
		Total:     total,
		TypePay:   "online",
	}

	if err := s.db.Model(&model.GroupOrder{}).Where("id = ?", id).Updates(&groupOrder).Error; err != nil {
		return nil, err
	}

	return &groupOrder, nil
}

func (s *groupOrderService) ChangeStatusOrder(orderId string, status string) error {
	var groupOrder model.GroupOrder

	switch status {
	case "00":
		groupOrder.Paid = true
	default:
		groupOrder.Paid = false
	}

	if err := s.db.
		Model(&model.GroupOrder{}).
		Where("vnp_txn_ref = ? AND paid = ?", orderId, false).
		Updates(&groupOrder).Error; err != nil {
		return err
	}

	if err := s.db.
		Model(&model.Order{}).
		Where("group_order_id = ?", groupOrder.ID).
		Updates(&model.Order{Paid: groupOrder.Paid}).Error; err != nil {
		return err
	}

	return nil
}

func (s *groupOrderService) GetPurchaseOrder(profileId uint) ([]*model.GroupOrder, error) {
	var data []*model.GroupOrder

	if err := s.db.Debug().
		Model(&model.GroupOrder{}).
		Preload("Orders").
		Where("profile_id = ? AND paid = ?", profileId, true).
		Find(&data).Error; err != nil {
		return nil, err
	}

	return data, nil
}

func NewGroupOrderService() GroupOrderService {
	return &groupOrderService{
		db: config.GetDB(),
	}
}

type PayloadCreateGroupOrder struct {
	Address string
	TypePay model.TYPE_PAY
	OrderId string
}
