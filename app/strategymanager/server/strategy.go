package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	spb "github.com/harveywangdao/ants/app/strategymanager/protos/strategy"
	mgrpb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"google.golang.org/grpc"
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
	return s.config.Process.Path
}
