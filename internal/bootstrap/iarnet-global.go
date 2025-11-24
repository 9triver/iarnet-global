package bootstrap

import (
	"context"
	"fmt"

	"github.com/9triver/iarnet-global/internal/config"
	"github.com/9triver/iarnet-global/internal/domain/registry"
	"github.com/9triver/iarnet-global/internal/intra/repository"
	"github.com/9triver/iarnet-global/internal/transport/http"
	"github.com/9triver/iarnet-global/internal/transport/rpc"
	"github.com/sirupsen/logrus"
)

type IarnetGlobal struct {
	// 配置
	Config *config.Config

	// 领域服务
	RegistryService registry.Service
	DomainManager   *registry.Manager
	DomainRepo      repository.DomainRepo
	// Transport 层
	HTTPServer *http.Server
	RPCManager *rpc.Manager
}

// Start 启动所有服务
func (ig *IarnetGlobal) Start(ctx context.Context) error {
	// 启动 Registry Manager（启动节点超时检测）
	if ig.DomainManager != nil {
		if err := ig.DomainManager.Start(ctx); err != nil {
			return fmt.Errorf("failed to start registry manager: %w", err)
		}
		logrus.Info("Registry manager started")
	}

	// 启动 RPC 服务器
	if ig.RPCManager != nil {
		if err := ig.RPCManager.Start(); err != nil {
			return fmt.Errorf("failed to start RPC server: %w", err)
		}
		logrus.Info("RPC server started")
	} else {
		return fmt.Errorf("rpc server is not initialized")
	}

	// 启动 HTTP 服务器
	if ig.HTTPServer != nil {
		ig.HTTPServer.Start()
		logrus.Info("HTTP server started")
	} else {
		return fmt.Errorf("http server is not initialized")
	}

	return nil
}

// Stop 停止所有服务并清理资源
func (ig *IarnetGlobal) Stop() error {
	// 停止 HTTP 服务器
	if ig.HTTPServer != nil {
		ig.HTTPServer.Stop()
		logrus.Info("HTTP server stopped")
	}

	// 停止 RPC 服务器
	if ig.RPCManager != nil {
		ig.RPCManager.Stop()
		logrus.Info("RPC server stopped")
	}

	// 停止 Registry Manager
	if ig.DomainManager != nil {
		ig.DomainManager.Stop()
		logrus.Info("Registry manager stopped")
	}

	logrus.Info("All services stopped")
	return nil
}
