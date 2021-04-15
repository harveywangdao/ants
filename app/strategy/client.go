package main

import (
	"context"
	//"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	spb "github.com/harveywangdao/ants/app/strategymanager/protos/strategy"
	"google.golang.org/grpc"
)

var (
	unixFile = filepath.Join(os.TempDir(), "12345678990")
)

func init() {
	logger.SetHandlers(logger.Console)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)
}

func unixConnect(addr string, t time.Duration) (net.Conn, error) {
	raddr, err := net.ResolveUnixAddr("unix", unixFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return net.DialUnix("unix", nil, raddr)
}

func do1() {
	argv := []string{"./bottle", unixFile}
	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	process, err := os.StartProcess("./bottle", argv, attr)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info("pid:", process.Pid)

	conn, err := grpc.Dial(unixFile, grpc.WithInsecure(), grpc.WithDialer(unixConnect))
	if err != nil {
		logger.Error(err)
		return
	}
	client := spb.NewStrategyClient(conn)

	for {
		time.Sleep(time.Second / 10)
		req := &spb.StartStrategyRequest{}
		_, err = client.StartStrategy(context.Background(), req)
		if err != nil {
			logger.Error(err)
		}
		logger.Info(conn.GetState())
	}
}

func do2() {
	// create unix socket listen
	if err := os.RemoveAll(unixFile); err != nil {
		logger.Fatal(err)
		return
	}

	addr, err := net.ResolveUnixAddr("unix", unixFile)
	if err != nil {
		logger.Fatal(err)
		return
	}
	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		logger.Fatal(err)
		return
	}
	lisFile, err := lis.File()
	if err != nil {
		logger.Fatal(err)
		return
	}

	// create process
	argv := []string{"./bottle", unixFile}
	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr, lisFile},
	}
	process, err := os.StartProcess("./bottle", argv, attr)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("pid:", process.Pid)

	conn, err := grpc.Dial(unixFile, grpc.WithInsecure(), grpc.WithDialer(unixConnect))
	if err != nil {
		logger.Error(err)
		return
	}
	client := spb.NewStrategyClient(conn)

	logger.Info("lisFile.Fd() =", lisFile.Fd())
	lisFile.Close()
	//lis.Close()

	for {
		time.Sleep(time.Second / 10)

		req := &spb.StartStrategyRequest{}
		_, err = client.StartStrategy(context.Background(), req)
		if err != nil {
			logger.Error(err)
		}

		logger.Info(conn.GetState())
	}
}

func do3() {
	r, w, err := os.Pipe()
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("w.Fd() =", w.Fd())
	logger.Info("r.Fd() =", r.Fd())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("sleep 2s...")
		time.Sleep(time.Second * 2)
		logger.Info("write...")
		w.WriteString("abcd")
	}()

	b := make([]byte, 32)
	logger.Info("reading...")
	n, err := r.Read(b)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(b[:n]))

	r.Close()
	w.Close()

	wg.Wait()
}

func do4() {
	r, w, err := os.Pipe()
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("w.Fd() =", w.Fd())
	logger.Info("r.Fd() =", r.Fd())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("sleep 2s...")
		time.Sleep(time.Second * 2)
		logger.Info("write...")
		r.WriteString("abcd")
	}()

	logger.Info("sleep 4s...")
	time.Sleep(time.Second * 4)

	b := make([]byte, 32)
	logger.Info("reading...")
	n, err := w.Read(b)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(b[:n]))

	wg.Wait()
}

func do5() {
	f, err := os.Create("./abc.txt")
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("f.Fd() =", f.Fd())
	f.WriteString("abc")

	fd1, err := syscall.Dup(int(f.Fd()))
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("fd1 =", fd1)

	f1 := os.NewFile(uintptr(fd1), "name1")
	f.WriteString("def")
	f1.WriteString("123")

	fd2 := 123
	err = syscall.Dup2(int(f.Fd()), fd2)
	if err != nil {
		logger.Error(err)
		return
	}
	f2 := os.NewFile(uintptr(fd2), "name2")
	f.WriteString("ghi")
	f1.WriteString("456")
	f2.WriteString("ABC")

	f.Close()
	f1.Close()
	f2.Close()
}

func do6() {
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		logger.Error(err)
		return
	}

	f1 := os.NewFile(uintptr(fds[0]), "")
	c1, err := net.FileConn(f1)
	if err != nil {
		logger.Error(err)
		return
	}

	f2 := os.NewFile(uintptr(fds[1]), "")
	c2, err := net.FileConn(f2)
	if err != nil {
		logger.Error(err)
		return
	}

	f1.Close()
	f2.Close()

	c1.Write([]byte("abc"))
	b := make([]byte, 10)
	n, err := c2.Read(b)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(b[:n]))

	c2.Write([]byte("123"))
	b = make([]byte, 10)
	n, err = c1.Read(b)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(b[:n]))

	c1.Close()
	c2.Close()
}

func main() {
	do1()
}
