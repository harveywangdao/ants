package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/harveywangdao/ants/logger"
	goodspb "github.com/harveywangdao/ants/rpc/goods"
	"google.golang.org/grpc"
)

func (s *Service) getServiceClientConn(svcName string) (*grpc.ClientConn, error) {
	addrs, err := s.discovery.QueryServiceIpPort(svcName)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if len(addrs) == 0 {
		logger.Error("can not find service", svcName)
		return nil, errors.New("can not find service " + svcName)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(addrs))

	logger.Info("client connect:", addrs[i])

	return grpc.Dial(addrs[i], grpc.WithInsecure())
}

func (s *Service) initGoodsServiceClient() error {
	conn, err := s.getServiceClientConn(s.Config.Client.GoodsServiceName)
	if err != nil {
		logger.Error(err)
		return err
	}
	//defer conn.Close()

	s.GoodsServiceClient = goodspb.NewGoodsServiceClient(conn)

	return nil
}
