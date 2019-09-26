package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/harveywangdao/ants/app/order/model"
	"github.com/harveywangdao/ants/cache/redis"
	"github.com/harveywangdao/ants/common"
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

	return &proto.GetOrderResponse{
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
	}, nil
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
后期优化
0.生成了订单但是没库存,直接撤销
1.支持实际支付

2.支付成功，扣库存失败(库存足)，抛消息队列
3.支付成功，扣库存失败(没库存)，直接撤销，退款

4.扣库存成功，修改支付状态失败，抛消息队列
*/
func (s *Service) PayOrder(ctx context.Context, req *proto.PayOrderRequest) (*proto.PayOrderResponse, error) {
	if req.OrderID == "" {
		return nil, errors.New("orderID is null")
	}

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

	// 测试，以后删除
	conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer conn.Close()

	conn.SetAdd(PayOrderPersonTime, order.BuyerID)
	conn.ZsetAdd(PayOrderPersonTimeByTimestamp, order.BuyerID, time.Now().Unix())
	conn.HyperLogLogAdd(PayOrderPersonTimeEstimate, order.BuyerID)

	/*getGoodsReq := &goodspb.GetGoodsRequest{
		GoodsID: order.GoodsID,
	}
	getGoodsResp, err := s.GoodsServiceClient.GetGoods(ctx, getGoodsReq)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if getGoodsResp.GoodsInfo.Stock < int32(order.Count) {
		return nil, errors.New("stock number is not enough")
	}*/

	// pay(order.Price)
	// 支付成功

	// 扣库存
	deductStockReq := &goodspb.DeductStockRequest{
		GoodsID: order.GoodsID,
		OrderID: req.OrderID,
		PayID:   util.GetUUID(), // 后期修改
		Number:  order.Count,
	}
	deductStockResp, err := s.GoodsServiceClient.DeductStock(ctx, deductStockReq)
	if err != nil {
		logger.Error(err)
		// 2.支付成功，扣库存失败(库存足)，抛消息队列
		// s.pushDeductStockEvent(ctx, req)
		s.publishDeductStockChannel(ctx, req)
		return nil, err
	}

	if deductStockResp.Code != 0 && deductStockResp.Code != common.ErrDeductStockRepeat {
		logger.Error("Code:", deductStockResp.Code, "CodeMsg:", deductStockResp.CodeMsg)

		if deductStockResp.Code == common.ErrStockIsNotEnough {
			// 3.支付成功，扣库存失败(没库存)，直接撤销，退款
		}

		return nil, errors.New(deductStockResp.CodeMsg)
	}

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

const (
	ActivityPrefix                = "ActivityPrefix"
	PayOrderPersonTime            = "PayOrderPersonTime"            // 历史支付人次,一个人算一次
	PayOrderPersonTimeByTimestamp = "PayOrderPersonTimeByTimestamp" // 历史支付人次,一个人算一次,按照时间戳排序
	PayOrderPersonTimeEstimate    = "PayOrderPersonTimeEstimate"    // 历史支付人次,一个人算一次,有误差,占资源小
)

func (s *Service) SetActivity(ctx context.Context, req *proto.SetActivityRequest) (*proto.SetActivityResponse, error) {
	if req.ActivityID == "" || req.ActivityName == "" || req.StartTime == "" || req.EndTime == "" {
		return nil, errors.New("param can not be null")
	}

	lock := redis.NewDistLock(s.RedisPool, "SetActivity"+req.ActivityID, s.Config.Redis.RedisLockTimeout)
	if err := lock.Lock(); err != nil {
		logger.Error(err)
		return nil, err
	}
	defer lock.Unlock()

	conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer conn.Close()

	if conn.IsKeyExist(ActivityPrefix + req.ActivityID) {
		return nil, fmt.Errorf("%s already existed", ActivityPrefix+req.ActivityID)
	}

	m := map[string]interface{}{
		"activityID":   req.ActivityID,
		"activityName": req.ActivityName,
		"startTime":    req.StartTime,
		"endTime":      req.EndTime,
	}

	if err := conn.Hmset(ActivityPrefix+req.ActivityID, m); err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.SetActivityResponse{
		CodeMsg: "add activity success",
	}, nil
}

func (s *Service) GetActivity(ctx context.Context, req *proto.GetActivityRequest) (*proto.GetActivityResponse, error) {
	if req.ActivityID == "" {
		return nil, errors.New("activityID is null")
	}

	conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer conn.Close()

	if !conn.IsKeyExist(ActivityPrefix + req.ActivityID) {
		return nil, fmt.Errorf("activity %s not existed", req.ActivityID)
	}

	activityInfo, err := conn.Hgetall(ActivityPrefix + req.ActivityID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.GetActivityResponse{
		ActivityID:   activityInfo["activityID"],
		ActivityName: activityInfo["activityName"],
		StartTime:    activityInfo["startTime"],
		EndTime:      activityInfo["endTime"],
	}, nil
}

func (s *Service) GetPayOrderPersonTime(ctx context.Context, req *proto.GetPayOrderPersonTimeRequest) (*proto.GetPayOrderPersonTimeResponse, error) {
	conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer conn.Close()

	personTime, err := conn.SetLen(PayOrderPersonTime)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	personList, err := conn.SetMembers(PayOrderPersonTime)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	personTimeByTimestamp, err := conn.ZsetLen(PayOrderPersonTimeByTimestamp)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	personListMap, err := conn.ZsetMembers(PayOrderPersonTimeByTimestamp)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	personTimeEstimate, err := conn.HyperLogLogLen(PayOrderPersonTimeEstimate)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.GetPayOrderPersonTimeResponse{
		PersonTime:            personTime,
		PersonList:            personList,
		PersonTimeByTimestamp: personTimeByTimestamp,
		PersonListMap:         personListMap,
		PersonTimeEstimate:    personTimeEstimate,
	}, nil
}
