package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	spb "github.com/harveywangdao/ants/app/strategymanager/protos/strategy"
	mgrpb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"go.etcd.io/etcd/clientv3"
)

func (s *StrategyManager) StartTask(ctx context.Context, in *mgrpb.StartTaskRequest) (*mgrpb.StartTaskResponse, error) {
	// 创建策略进程,并执行start
	if err := s.createProccesAndStartStrategy(ctx, in); err != nil {
		logger.Error(err)
		return nil, err
	}

	return &mgrpb.StartTaskResponse{}, nil
}

func (s *StrategyManager) StopTask(ctx context.Context, in *mgrpb.StopTaskRequest) (*mgrpb.StopTaskResponse, error) {
	// 找到策略进程
	uniqueId := fmt.Sprintf("%s-%s-%s", in.ApiKey, in.StrategyName, in.InstrumentId)
	sp, err := s.getProcess(uniqueId)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 发送命令
	req := &spb.StopStrategyRequest{
		UserId:       in.UserId,
		Exchange:     in.Exchange,
		ApiKey:       in.ApiKey,
		StrategyName: in.StrategyName,
		InstrumentId: in.InstrumentId,
	}
	_, err = sp.Client.StopStrategy(ctx, req)
	if err != nil {
		logger.Error(err)
		//return nil, err
	}

	// 销毁进程
	sp.Close()

	return &mgrpb.StopTaskResponse{}, nil
}

func (s *StrategyManager) TaskCommandExec(ctx context.Context, in *mgrpb.TaskCommandExecRequest) (*mgrpb.TaskCommandExecResponse, error) {
	// 找到策略进程
	uniqueId := fmt.Sprintf("%s-%s-%s", in.ApiKey, in.StrategyName, in.InstrumentId)
	sp, err := s.getProcess(uniqueId)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 发送命令
	req := &spb.StrategyExecRequest{
		UserId:       in.UserId,
		Exchange:     in.Exchange,
		ApiKey:       in.ApiKey,
		StrategyName: in.StrategyName,
		InstrumentId: in.InstrumentId,
		Params:       in.Params,
	}
	_, err = sp.Client.StrategyExec(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return &mgrpb.TaskCommandExecResponse{}, nil
}

func (s *StrategyManager) getProcess(uniqueId string) (*StrategyProcess, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sp, ok := s.processes[uniqueId]
	if !ok {
		return nil, fmt.Errorf("%s not existed", uniqueId)
	}
	return sp, nil
}

func (s *StrategyManager) delProcess(uniqueId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.processes, uniqueId)
	return nil
}

func (s *StrategyManager) createProccesAndStartStrategy(ctx context.Context, in *mgrpb.StartTaskRequest) error {
	uniqueId := fmt.Sprintf("%s-%s-%s", in.ApiKey, in.StrategyName, in.InstrumentId)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.processes[uniqueId]; ok {
		return fmt.Errorf("%s already existed", uniqueId)
	}

	sp, err := NewStrategyProcess(s, in)
	if err != nil {
		logger.Error(err)
		return err
	}
	s.processes[uniqueId] = sp
	return nil
}

func (s *StrategyManager) getProcDir() string {
	return s.Config.Process.Path
}

func (s *StrategyManager) registerTask(port int) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: s.Config.Etcd.Endpoints,
	})
	if err != nil {
		logger.Fatal(err)
		return
	}
	tk := time.NewTicker(time.Second * 3)

	var leaseid clientv3.LeaseID
	defer func() {
		tk.Stop()
		// 删除etcd注册信息
		s.unregister(cli, leaseid)
		cli.Close()
	}()

	for {
		select {
		case <-s.closeCh:
			return
		case <-tk.C:
			leaseid, err = s.register(cli, leaseid, port)
			if err != nil {
				logger.Error(err)
			}
		}
	}
}

type RegisterInfo struct {
	Uptime    int64 `json:"uptime"`
	Available bool  `json:"available"`
}

func (s *StrategyManager) register(cli *clientv3.Client, leaseid clientv3.LeaseID, port int) (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if leaseid != 0 {
		timeToLiveResp, err := cli.TimeToLive(ctx, leaseid)
		if err != nil {
			logger.Error(err)
			return leaseid, err
		}
		if timeToLiveResp.TTL == -1 {
			leaseid = 0
		}
	}

	if leaseid == 0 {
		resp, err := cli.Grant(ctx, 5)
		if err != nil {
			logger.Error(err)
			return leaseid, err
		}
		leaseid = resp.ID

		info := &RegisterInfo{
			Uptime:    time.Now().Unix(),
			Available: true,
		}
		data, err := json.Marshal(info)
		if err != nil {
			logger.Error(err)
			return leaseid, err
		}

		localIp := s.GetLocalIp()
		if localIp == "" {
			err = fmt.Errorf("can not get local ip")
			logger.Fatal(err)
			return leaseid, err
		}

		// /service/strategymanager/10.22.33.55:32154 --> {"uptime":111111111, "available":true}
		key := fmt.Sprintf("/service/strategymanager/%s:%d", localIp, port)
		_, err = cli.Put(ctx, key, string(data), clientv3.WithLease(leaseid))
		if err != nil {
			logger.Error(err)
			return leaseid, err
		}
	} else {
		resp, err := cli.KeepAliveOnce(ctx, leaseid)
		if err != nil {
			logger.Error(err)
			return leaseid, err
		}
		logger.Info(resp)
	}

	return leaseid, nil
}

func (s *StrategyManager) unregister(cli *clientv3.Client, leaseid clientv3.LeaseID) error {
	if leaseid == 0 {
		return nil
	}
	_, err := cli.Revoke(context.TODO(), leaseid)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (s *StrategyManager) GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error(err)
		return ""
	}

	ip := ""
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				logger.Info("ip:", ip)
			}
		}
	}

	return ip
}
