package service

import (
	"context"
	"errors"

	"github.com/harveywangdao/ants/app/order/model"
	"github.com/harveywangdao/ants/cache/redis"
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

// 增加订单并不会扣库存
func (s *Service) AddOrder(ctx context.Context, req *proto.AddOrderRequest) (*proto.AddOrderResponse, error) {
	if req.BuyerID == "" || req.GoodsID == "" {
		return nil, errors.New("buyerID or goodsID is null")
	}

	if req.Count == 0 {
		return nil, errors.New("count can not be 0")
	}

	getGoodsReq := &goodspb.GetGoodsRequest{
		GoodsID: req.GoodsID,
	}
	getGoodsResp, err := s.GoodsServiceClient.GetGoods(ctx, getGoodsReq)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 这里不解决高并发下数据不一致的问题,有可能库存取出之后就被改变了
	if getGoodsResp.GoodsInfo.Stock < int32(req.Count) {
		return nil, errors.New("stock number is not enough")
	}

	orderID := util.GetUUID()

	order := &model.OrderModel{
		OrderID:   orderID,
		SellerID:  "",
		BuyerID:   req.BuyerID,
		GoodsID:   req.GoodsID,
		GoodsName: getGoodsResp.GoodsInfo.Name,
		Count:     req.Count,
		Price:     getGoodsResp.GoodsInfo.Price * float64(req.Count),
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
			Count:     order.Count,
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

/*
生成了订单但是没库存的处理
后期优化设计
1.支付
2.抛消息队列
3.修改支付状态为已支付和扣库存，两者必须是一个事务
4.修改订单和扣库存只要有失败的必须有重试机制
5.库存不足要退钱和撤销订单
*/
func (s *Service) PayOrder(ctx context.Context, req *proto.PayOrderRequest) (*proto.PayOrderResponse, error) {
	if req.OrderID == "" {
		return nil, errors.New("orderID is null")
	}

	/*conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer conn.Close()

	value := util.GetUUID()
	if err := conn.Lock("PayOrder"+req.OrderID, value, s.Config.Redis.RedisLockTimeout); err != nil {
		logger.Error(err)
		return nil, err
	}
	defer func() {
		if err := conn.Unlock("PayOrder"+req.OrderID, value); err != nil {
			logger.Error(err)
		}
	}()*/

	lock := redis.NewDistLock(s.RedisPool, "PayOrder"+req.OrderID, s.Config.Redis.RedisLockTimeout)
	if err := lock.Lock(); err != nil {
		logger.Error(err)
		return nil, err
	}
	defer lock.Unlock()

	// 查询订单
	var order model.OrderModel
	if err := s.db.Where("order_id = ?", req.OrderID).First(&order).Error; err != nil {
		logger.Error(err)
		return nil, err
	}
	if order.Status != OrderStatusUnpaid {
		return nil, errors.New("current status is not unpaid")
	}

	getGoodsReq := &goodspb.GetGoodsRequest{
		GoodsID: order.GoodsID,
	}
	getGoodsResp, err := s.GoodsServiceClient.GetGoods(ctx, getGoodsReq)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if getGoodsResp.GoodsInfo.Stock < int32(order.Count) {
		return nil, errors.New("stock number is not enough")
	}

	// pay(order.Price)
	// 支付成功

	// 扣库存
	deductStockReq := &goodspb.DeductStockRequest{
		GoodsID: order.GoodsID,
		Number:  order.Count,
	}
	deductStockResp, err := s.GoodsServiceClient.DeductStock(ctx, deductStockReq)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	_ = deductStockResp

	// 修改订单状态
	param := map[string]interface{}{
		"status": OrderStatusPaid,
		"pay":    order.Price,
	}
	if err := s.db.Model(model.OrderModel{}).Where("order_id = ?", req.OrderID).Updates(param).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.PayOrderResponse{
		CodeMsg: "pay success",
	}, nil
}
