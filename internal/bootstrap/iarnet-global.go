package bootstrap

import (
	"context"
	"fmt"

	"github.com/9triver/iarnet-global/internal/config"
	"github.com/9triver/iarnet-global/internal/domain/registry"
	"github.com/9triver/iarnet-global/internal/intra/repository"
	"github.com/9triver/iarnet-global/internal/transport/http"
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
}

// Start 启动所有服务
func (ig *IarnetGlobal) Start(ctx context.Context) error {
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

	logrus.Info("All services stopped")
	return nil
}
