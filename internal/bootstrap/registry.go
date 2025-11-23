package bootstrap

import (
	"github.com/9triver/iarnet-global/internal/domain/registry"
	"github.com/sirupsen/logrus"
)

// bootstrapRegistry 初始化 Registry 模块
func bootstrapRegistry(ig *IarnetGlobal) error {
	// 创建 Registry Manager
	manager := registry.NewManager()

	// 创建 Registry Service
	service := registry.NewService(manager)
	ig.RegistryService = service

	logrus.Info("Registry module initialized")
	return nil
}
