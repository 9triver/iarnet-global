package rpc

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/9triver/iarnet-global/internal/domain/registry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	registrypb "github.com/9triver/iarnet-global/internal/proto/registry"
	registryrpc "github.com/9triver/iarnet-global/internal/transport/rpc/registry"
)

// server RPC 服务器
type server struct {
	Server   *grpc.Server
	Listener net.Listener
}

func (rs *server) GracefulStop() {
	if rs == nil {
		return
	}
	if rs.Server != nil {
		rs.Server.GracefulStop()
	}
	if rs.Listener != nil {
		_ = rs.Listener.Close()
	}
}

func (rs *server) Stop() {
	if rs == nil {
		return
	}
	if rs.Server != nil {
		rs.Server.Stop()
	}
	if rs.Listener != nil {
		_ = rs.Listener.Close()
	}
}

// Options RPC 服务器选项
type Options struct {
	RegistryAddr       string
	RegistryService    *registry.Manager
	RegistryServerOpts []grpc.ServerOption
}

// Manager 管理 RPC 服务器的生命周期
type Manager struct {
	Registry  *server
	Options   Options
	startOnce sync.Once
	stopOnce  sync.Once
}

// NewManager 创建新的 RPC 服务器管理器
func NewManager(opts Options) *Manager {
	return &Manager{
		Registry:  nil,
		Options:   opts,
		startOnce: sync.Once{},
		stopOnce:  sync.Once{},
	}
}

// Start 启动 RPC 服务器
func (m *Manager) Start() error {
	if m.Options.RegistryAddr == "" {
		return errors.New("registry listen address is required")
	}
	if m.Options.RegistryService == nil {
		return errors.New("registry service is required")
	}

	m.startOnce.Do(func() {
		// 配置 Registry 服务器选项
		registryOpts := append([]grpc.ServerOption{}, m.Options.RegistryServerOpts...)
		registryOpts = append(registryOpts, grpc.MaxRecvMsgSize(512*1024*1024))

		// 启动 Registry 服务器
		registry, err := startServer(m.Options.RegistryAddr, registryOpts, func(s *grpc.Server) {
			registrypb.RegisterServiceServer(s, registryrpc.NewServer(m.Options.RegistryService))
		})
		if err != nil {
			logrus.WithError(err).Error("failed to start registry server")
		} else {
			logrus.Infof("Registry RPC server listening on %s", m.Options.RegistryAddr)
			m.Registry = registry
		}
	})

	return nil
}

// Stop 停止所有 RPC 服务器
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		shutdownWithTimeout(m.Registry, 30*time.Second)
	})
}

func shutdownWithTimeout(s *server, timeout time.Duration) {
	if s == nil {
		return
	}

	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(timeout):
		logrus.Warn("grpc server graceful stop timed out, forcing stop")
		s.Stop()
	}
}

func startServer(addr string, opts []grpc.ServerOption, register func(*grpc.Server)) (*server, error) {
	lis, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(opts...)
	register(s)

	go func() {
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logrus.WithError(err).Error("grpc server stopped unexpectedly")
		}
	}()

	return &server{Server: s, Listener: lis}, nil
}
