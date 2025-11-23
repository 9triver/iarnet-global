package bootstrap

import (
	"fmt"

	"github.com/9triver/iarnet-global/internal/config"
	"github.com/sirupsen/logrus"
)

// Initialize 初始化所有模块
// 按照依赖顺序初始化：Registry -> Transport
func Initialize(cfg *config.Config) (*IarnetGlobal, error) {
	ig := &IarnetGlobal{
		Config:          cfg,
		RegistryService: nil,
		HTTPServer:      nil,
	}

	// 1. 初始化 Registry 模块
	if err := bootstrapRegistry(ig); err != nil {
		return nil, fmt.Errorf("failed to initialize registry module: %w", err)
	}

	// 2. 初始化 Transport 层
	if err := bootstrapTransport(ig); err != nil {
		return nil, fmt.Errorf("failed to initialize transport layer: %w", err)
	}

	logrus.Info("All modules initialized successfully")
	return ig, nil
}
