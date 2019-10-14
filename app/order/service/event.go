package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gomodule/redigo/redis"
	"github.com/harveywangdao/ants/app/order/model"
	"github.com/harveywangdao/ants/common"
	"github.com/harveywangdao/ants/logger"
	goodspb "github.com/harveywangdao/ants/rpc/goods"
	proto "github.com/harveywangdao/ants/rpc/order"
	"github.com/harveywangdao/ants/util"
	nsq "github.com/nsqio/go-nsq"
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
					logger.Debug(channels, "subscribe success")
				case 0:
					logger.Error(channels, "subscribe fail")
					return
				}
			}
		}
	}()

	/*go func() {
		consumer, err := sarama.NewConsumer(s.Config.Kafka.Addrs, nil)
		if err != nil {
			logger.Error(err)
			return
		}

		defer func() {
			if err := consumer.Close(); err != nil {
				logger.Error(err)
			}
		}()

		partitionConsumer, err := consumer.ConsumePartition("xiaoming", 0, sarama.OffsetNewest)
		if err != nil {
			logger.Error(err)
			return
		}

		defer func() {
			if err := partitionConsumer.Close(); err != nil {
				logger.Error(err)
			}
		}()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				logger.Info("Consumed message partition", msg.Partition, "offset", msg.Offset)
				if err := s.deductStockEvent(msg.Value); err != nil {
					logger.Error(err)

				}
			}
		}
	}()*/

	go func() {
		config := sarama.NewConfig()
		config.Version = sarama.V2_2_0_0
		config.Consumer.Return.Errors = true
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

		client, err := sarama.NewClient(s.Config.Kafka.Addrs, config)
		if err != nil {
			logger.Error(err)
			return
		}
		defer func() { _ = client.Close() }()

		group, err := sarama.NewConsumerGroupFromClient("my-group-1", client)
		if err != nil {
			logger.Error(err)
			return
		}
		defer func() { _ = group.Close() }()

		go func() {
			for err := range group.Errors() {
				logger.Error("ERROR", err)
			}
		}()

		ctx := context.Background()

		for {
			topics := []string{"xiaoming"}
			handler := consumerGroupHandler{Service: s}

			logger.Debug("group.Consume start")
			err := group.Consume(ctx, topics, handler)
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Debug("group.Consume stop")
		}
	}()

	go func() {
		config := nsq.NewConfig()
		config.DefaultRequeueDelay = time.Second * 5
		config.MaxBackoffDuration = 50 * time.Millisecond
		consumer, err := nsq.NewConsumer("nsq_topic_DeductStock", "channel-1", config)
		if err != nil {
			logger.Error(err)
			return
		}
		consumer.SetLogger(nil, nsq.LogLevelError)

		h := &nsqHandler{
			Service:  s,
			Consumer: consumer,
		}
		consumer.AddHandler(h)

		err = consumer.ConnectToNSQD(s.Config.Nsq.Addrs[0])
		if err != nil {
			logger.Error(err)
			return
		}

		logger.Infof("%+v\n", consumer.Stats())

		<-consumer.StopChan

		logger.Info("nsq consumer stop")
	}()
}

type nsqHandler struct {
	Service  *Service
	Consumer *nsq.Consumer
}

func (h *nsqHandler) HandleMessage(msg *nsq.Message) error {
	logger.Info(msg.ID, msg.Timestamp, msg.Attempts, string(msg.Body))

	if err := h.Service.deductStockEvent(msg.Body); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (h *nsqHandler) LogFailedMessage(msg *nsq.Message) {
	logger.Info("Consumer Stop")
	//h.Consumer.Stop()
}

func (s *Service) produceDeductStockNsqMsg(ctx context.Context, req *proto.PayOrderRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	cfg := nsq.NewConfig()
	producer, err := nsq.NewProducer(s.Config.Nsq.Addrs[0], cfg)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer producer.Stop()
	producer.SetLogger(nil, nsq.LogLevelError)

	err = producer.Publish("nsq_topic_DeductStock", data)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

type consumerGroupHandler struct {
	Service *Service
}

func (consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	logger.Debug("group.Consume Setup")
	return nil
}

func (consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	logger.Debug("group.Consume Cleanup")
	return nil
}

func (h consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logger.Infof("Message topic:%s partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)

		if err := h.Service.deductStockEvent(msg.Value); err != nil {
			logger.Error(err)
			//sess.ResetOffset(msg.Topic, msg.Partition, msg.Offset, "")
			time.Sleep(1 * time.Second)
			return err
		}

		sess.MarkMessage(msg, "")
	}

	return nil
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

func (s *Service) produceDeductStockMsg(ctx context.Context, req *proto.PayOrderRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		logger.Error(err)
		return err
	}

	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(s.Config.Kafka.Addrs, config)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer func() {
		if err := producer.Close(); err != nil {
			logger.Error(err)
		}
	}()

	msg := &sarama.ProducerMessage{
		Topic: "xiaoming",
		Key:   sarama.StringEncoder("DeductStock"),
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("Produced message to partition", partition, "with offset", offset)

	return nil
}
