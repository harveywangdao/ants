package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/harveywangdao/ants/app/scheduler/util"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	spb "github.com/harveywangdao/ants/app/strategymanager/protos/strategy"
	mgrpb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"google.golang.org/grpc"
)

func (s *StrategyManager) StartTask(ctx context.Context, in *mgrpb.StartTaskRequest) (*mgrpb.StartTaskResponse, error) {
	// 创建策略进程
	sp, err := s.createProcess(ctx, in)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 发送命令
	req := &spb.StartStrategyRequest{
		UserId:       in.UserId,
		Exchange:     in.Exchange,
		ApiKey:       in.ApiKey,
		SecretKey:    in.SecretKey,
		Passphrase:   in.Passphrase,
		StrategyName: in.StrategyName,
		InstrumentId: in.InstrumentId,
		Endpoint:     in.Endpoint,
		WsEndpoint:   in.WsEndpoint,
		Params:       in.Params,
	}
	_, err = sp.Client.StartStrategy(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 向etcd注册
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
		SecretKey:    in.SecretKey,
		Passphrase:   in.Passphrase,
		StrategyName: in.StrategyName,
		InstrumentId: in.InstrumentId,
		Endpoint:     in.Endpoint,
		WsEndpoint:   in.WsEndpoint,
		Params:       in.Params,
	}
	_, err = sp.Client.StopStrategy(ctx, req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// 销毁进程
	s.destroyProcess(uniqueId)

	// 从etcd删除注册信息
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
	s.mu.Lock()
	defer s.mu.Unlock()
	sp, ok := s.processes[uniqueId]
	if !ok {
		return nil, fmt.Errorf("%s not existed", uniqueId)
	}
	return sp, nil
}

func (s *StrategyManager) destroyProcess(uniqueId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sp, ok := s.processes[uniqueId]
	if !ok {
		return fmt.Errorf("%s not existed", uniqueId)
	}
	sp.Close()

	delete(s.processes, uniqueId)
	return nil
}

func (s *StrategyManager) createProcess(ctx context.Context, in *mgrpb.StartTaskRequest) (*StrategyProcess, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	uniqueId := fmt.Sprintf("%s-%s-%s", in.ApiKey, in.StrategyName, in.InstrumentId)
	if _, ok := s.processes[uniqueId]; ok {
		return nil, fmt.Errorf("%s already existed", uniqueId)
	}

	path := filepath.Join(s.config.Process.Path, in.StrategyName)
	sp, err := NewStrategyProcess(path)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	s.processes[uniqueId] = sp
	return sp, nil
}

type StrategyProcess struct {
	process *os.Process
	msgCh   chan []byte

	unixFile string
	conn     *grpc.ClientConn
	Client   *spb.StrategyClient

	closedCh chan bool
	once     sync.Once
}

func NewStrategyProcess(path string) (*StrategyProcess, error) {
	unixFile := filepath.Join(os.TempDir(), util.GetUUID())
	argv := []string{unixFile}
	attr := &os.ProcAttr{
		//Env: os.Environ(),
		//Files: []*os.File{},
	}
	process, err := os.StartProcess(path, argv, attr)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	sp := &StrategyProcess{
		process:  process,
		unixFile: unixFile,
		msgCh:    make(chan []byte),
		closedCh: make(chan bool),
	}

	conn, err := grpc.Dial(unixFile, grpc.WithInsecure(), grpc.WithDialer(sp.unixConnect))
	if err != nil {
		logger.Error(err)
		s.process.Kill()
		return nil, err
	}
	client := spb.NewStrategyClient(conn)
	sp.conn = conn
	sp.Client = client

	go sp.processTask()

	return sp, nil
}

func (s *StrategyProcess) unixConnect(addr string, t time.Duration) (net.Conn, error) {
	addr, err := net.ResolveUnixAddr("unix", s.unixFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return net.DialUnix("unix", nil, addr)
}

func (s *StrategyProcess) processTask() {
	for {
		select {
		case <-s.closedCh:
			s.process.Wait()
			os.RemoveAll(s.unixFile)
			return
		}
	}
}

func (s *StrategyProcess) Close() {
	s.once.Do(func() {
		close(s.closedCh)
		s.conn.Close()
		s.process.Kill()
	})
}
