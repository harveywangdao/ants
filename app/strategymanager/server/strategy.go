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
	pb "github.com/harveywangdao/ants/app/strategymanager/protos/strategymanager"
	"google.golang.org/grpc"
)

func (s *StrategyManager) StartStrategy(ctx context.Context, in *StartStrategyRequest) (*StartStrategyResponse, error) {
	// 创建策略进程
	// 发送命令

	// 向etcd注册
}

func (s *StrategyManager) StopStrategy(ctx context.Context, in *StopStrategyRequest) (*StopStrategyResponse, error) {
	// 找到策略进程
	// 发送命令

	// 从etcd删除注册信息
}

func (s *StrategyManager) StrategyExec(ctx context.Context, in *StrategyExecRequest) (*StrategyExecResponse, error) {
	// 找到策略进程
	// 发送命令
}

func (s *StrategyManager) createProcess(ctx context.Context, in *StartStrategyRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	procUniqueId := fmt.Sprintf("%s-%s-%s", in.ApiKey, in.StrategyName, in.InstrumentId)
	if _, ok := s.processes[procUniqueId]; ok {
		return fmt.Errorf("%s already existed", procUniqueId)
	}

	path := filepath.Join(s.config.Process.Path, in.StrategyName)
	sp, err := NewStrategyProcess(path)
	if err != nil {
		logger.Error(err)
		return err
	}
	s.processes[procUniqueId] = sp
}

type StrategyProcess struct {
	process *os.Process
	msgCh   chan []byte

	unixFile string
	conn     *grpc.ClientConn
	client   *spb.StrategyClient

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
	sp.client = client

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
