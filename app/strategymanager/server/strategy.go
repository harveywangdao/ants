package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
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
		return nil, err
	}

	// 销毁进程
	sp.KillProc()

	// 从etcd删除注册信息

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

func (s *StrategyManager) delProc(uniqueId string) error {
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

type StrategyProcess struct {
	mgr *StrategyManager

	uniqueId string
	process  *os.Process

	unixFile string
	conn     *grpc.ClientConn
	Client   spb.StrategyClient

	startStrategyReq *spb.StartStrategyRequest

	once    sync.Once
	closeCh chan bool
}

func NewStrategyProcess(mgr *StrategyManager, startReq *mgrpb.StartTaskRequest) (*StrategyProcess, error) {
	unixFile := filepath.Join(os.TempDir(), util.GetUUID())
	argv := []string{unixFile}
	attr := &os.ProcAttr{
		//Env: os.Environ(),
		//Files: []*os.File{},
	}
	path := filepath.Join(mgr.getProcDir(), startReq.StrategyName)
	process, err := os.StartProcess(path, argv, attr)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	sp := &StrategyProcess{
		mgr:      mgr,
		uniqueId: fmt.Sprintf("%s-%s-%s", startReq.ApiKey, startReq.StrategyName, startReq.InstrumentId),
		process:  process,
		unixFile: unixFile,
		closeCh:  make(chan bool),
	}

	conn, err := grpc.Dial(unixFile, grpc.WithInsecure(), grpc.WithDialer(sp.unixConnect))
	if err != nil {
		logger.Error(err)
		s.process.Kill()
		s.process.Wait()
		return nil, err
	}
	sp.conn = conn
	sp.Client = spb.NewStrategyClient(sp.conn)

	go sp.task()
	go sp.waitProc()

	if err := sp.startStrategy(startReq); err != nil {
		logger.Error(err)
		s.process.Kill()
		sp.release()
		return nil, err
	}

	return sp, nil
}

func (s *StrategyProcess) task() {
	tk := time.NewTicker(time.Second * 3)
	defer tk.Stop()

	for {
		select {
		case <-s.closeCh:
			return
		case <-tk.C:
			// 向etcd注册
			// 刷新etcd
		}
	}
}

func (s *StrategyProcess) waitProc() {
	state, err := s.process.Wait()
	if err != nil {
		logger.Error(err)
	}
	logger.Infof("proc id: %s exit state: %s", state.Pid(), state.String())
	s.release()
}

func (s *StrategyProcess) startStrategy(startReq *mgrpb.StartTaskRequest) error {
	// 发送命令
	s.startStrategyReq = &spb.StartStrategyRequest{
		UserId:       startReq.UserId,
		Exchange:     startReq.Exchange,
		ApiKey:       startReq.ApiKey,
		SecretKey:    startReq.SecretKey,
		Passphrase:   startReq.Passphrase,
		StrategyName: startReq.StrategyName,
		InstrumentId: startReq.InstrumentId,
		Endpoint:     startReq.Endpoint,
		WsEndpoint:   startReq.WsEndpoint,
		Params:       startReq.Params,
	}
	_, err = sp.Client.StartStrategy(ctx, s.startStrategyReq)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (s *StrategyProcess) release() {
	s.once.Do(func() {
		close(s.closeCh)
		if err := s.conn.Close(); err != nil {
			logger.Error(err)
		}
		if err := os.RemoveAll(s.unixFile); err != nil {
			logger.Error(err)
		}
	})
}

func (s *StrategyProcess) unixConnect(addr string, t time.Duration) (net.Conn, error) {
	addr, err := net.ResolveUnixAddr("unix", s.unixFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return net.DialUnix("unix", nil, addr)
}

func (s *StrategyProcess) KillProc() {
	if err := s.process.Kill(); err != nil {
		logger.Error(err)
	}
}
