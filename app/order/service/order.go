package service

import (
	"context"
	"errors"

	"github.com/harveywangdao/ants/app/order/model"
	"github.com/harveywangdao/ants/logger"
	goodspb "github.com/harveywangdao/ants/rpc/goods"
	proto "github.com/harveywangdao/ants/rpc/order"
	"github.com/harveywangdao/ants/util"
)

const (
	OrderStatusUnpaid = 0
	OrderStatusPaid   = 1
	OrderStatusRevoke = 2
)

func (s *Service) AddOrder(ctx context.Context, req *proto.AddOrderRequest) (*proto.AddOrderResponse, error) {
	if req.BuyerID == "" || req.GoodsID == "" {
		return nil, errors.New("buyerID or goodsID is null")
	}

	getGoodsReq := &goodspb.GetGoodsRequest{
		GoodsID: req.GoodsID,
	}
	getGoodsResp, err := s.GoodsServiceClient.GetGoods(ctx, getGoodsReq)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	orderID := util.GetUUID()

	order := &model.OrderModel{
		OrderID:   orderID,
		SellerID:  "",
		BuyerID:   req.BuyerID,
		GoodsID:   req.GoodsID,
		GoodsName: getGoodsResp.GoodsInfo.Name,
		Price:     getGoodsResp.GoodsInfo.Price,
		Status:    OrderStatusUnpaid,
	}

	err = s.db.Create(order).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.AddOrderResponse{
		OrderID: orderID,
	}, nil
}

func (s *Service) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	if req.OrderID == "" {
		return nil, errors.New("orderID is null")
	}

	var order model.OrderModel
	err := s.db.Where("order_id = ?", req.OrderID).First(&order).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info(order)

	resp := &proto.GetOrderResponse{
		OrderInfo: &proto.OrderInfo{
			SellerID:  order.SellerID,
			BuyerID:   order.BuyerID,
			GoodsID:   order.GoodsID,
			GoodsName: order.GoodsName,
			Price:     order.Price,
			Pay:       order.Pay,
			Status:    uint32(order.Status),
		},
	}

	return resp, nil
}

func (s *Service) DelOrder(ctx context.Context, req *proto.DelOrderRequest) (*proto.DelOrderResponse, error) {
	if req.OrderID == "" {
		return nil, errors.New("orderID is null")
	}

	err := s.db.Where("order_id = ?", req.OrderID).Delete(model.OrderModel{}).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.DelOrderResponse{
		CodeMsg: "delete success",
	}, nil
}
