package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/harveywangdao/ants/app/order/model"
	"github.com/harveywangdao/ants/common"
	"github.com/harveywangdao/ants/logger"
	goodspb "github.com/harveywangdao/ants/rpc/goods"
	proto "github.com/harveywangdao/ants/rpc/order"
	"github.com/harveywangdao/ants/util"
)

const (
	DeductStockEventList       = "DeductStockEventList"
	DeductStockEventChannel    = "DeductStockEventChannel"
	PayOrderEventTimerDuration = 2 * time.Second
)

func DeductStockEventStartListen(s *Service) {
	go func() {
		conn, err := s.RedisPool.Get()
		if err != nil {
			logger.Error(err)
			return
		}
		defer conn.Close()

		ticker := time.NewTicker(PayOrderEventTimerDuration)

		for {
			select {
			case <-ticker.C:
				data, err := conn.ListPop(DeductStockEventList)
				if err != nil {
					logger.Error(err)
					break
				}

				if data != "" {
					if err := s.deductStockEvent([]byte(data)); err != nil {
						logger.Error(err)

						if err := conn.ListPush(DeductStockEventList, data); err != nil {
							logger.Error(err)
						}
					}
				}
			}
		}
	}()

	go func() {
		conn, err := s.RedisPool.Get()
		if err != nil {
			logger.Error(err)
			return
		}
		defer conn.Close()

		var channels []string
		channels = append(channels, DeductStockEventChannel)

		psc := redis.PubSubConn{Conn: conn}
		if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
			logger.Error(err)
			return
		}
		defer psc.Unsubscribe()

		for {
			switch n := psc.Receive().(type) {
			case error:
				logger.Error(n)
				return

			case redis.Message:
				logger.Info(n.Channel, string(n.Data))
				if n.Channel == DeductStockEventChannel && len(n.Data) != 0 {
					if err := s.deductStockEvent(n.Data); err != nil {
						time.Sleep(1 * time.Second)

						conn2, err := s.RedisPool.Get()
						if err != nil {
							logger.Error(err)
							break
						}
						conn2.Publish(DeductStockEventChannel, string(n.Data))
						conn2.Close()
					}
				}

			case redis.Subscription:
				switch n.Count {
				case len(channels):
					logger.Info(channels, "subscribe success")
				case 0:
					logger.Error(channels, "subscribe fail")
					return
				}
			}
		}
	}()
}

func (s *Service) deductStockEvent(data []byte) error {
	ctx := context.Background()

	req := &proto.PayOrderRequest{}
	if err := json.Unmarshal(data, req); err != nil {
		logger.Error(err)
		return err
	}

	if req.OrderID == "" {
		return nil
	}

	// 查询订单
	var order model.OrderModel
	if err := s.db.Where("order_id = ?", req.OrderID).First(&order).Error; err != nil {
		logger.Error(err)
		return err
	}
	if order.Status != OrderStatusUnpaid {
		return nil
	}

	getGoodsReq := &goodspb.GetGoodsRequest{
		GoodsID: order.GoodsID,
	}
	getGoodsResp, err := s.GoodsServiceClient.GetGoods(ctx, getGoodsReq)
	if err != nil {
		logger.Error(err)
		return err
	}
	if getGoodsResp.GoodsInfo.Stock < int32(order.Count) {
		return nil
	}

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
		return err
	}

	if deductStockResp.Code != 0 && deductStockResp.Code != common.ErrDeductStockRepeat {
		logger.Error("Code:", deductStockResp.Code, "CodeMsg:", deductStockResp.CodeMsg)

		if deductStockResp.Code == common.ErrStockIsNotEnough {
			return nil
		}

		return errors.New(deductStockResp.CodeMsg)
	}

	// 修改订单状态
	param := map[string]interface{}{
		"status": OrderStatusPaid,
		"pay":    order.Price,
	}
	if err := s.db.Model(model.OrderModel{}).Where("order_id = ?", req.OrderID).Updates(param).Error; err != nil {
		logger.Error(err)
		// 要不要重试
		return nil
	}

	return nil
}

func (s *Service) pushDeductStockEvent(ctx context.Context, req *proto.PayOrderRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return err
	}
	defer conn.Close()

	if err := conn.ListPush(DeductStockEventList, string(data)); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (s *Service) publishDeductStockChannel(ctx context.Context, req *proto.PayOrderRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	conn, err := s.RedisPool.Get()
	if err != nil {
		logger.Error(err)
		return err
	}
	defer conn.Close()

	if err := conn.Publish(DeductStockEventChannel, string(data)); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
