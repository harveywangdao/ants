package server

import (
	"context"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	pb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
)

func (s *StrategyManager) StartStrategy(ctx context.Context, in *StartStrategyRequest) (*StartStrategyResponse, error) {
	// 创建策略进程
}

func (s *StrategyManager) StopStrategy(ctx context.Context, in *StopStrategyRequest) (*StopStrategyResponse, error) {
	// 找到策略进程
	// 发送命令
}

func (s *StrategyManager) StrategyExec(ctx context.Context, in *StrategyExecRequest) (*StrategyExecResponse, error) {
	// 找到策略进程
	// 发送命令
}
