package bootstrap

import (
	"context"
	"fmt"

	"github.com/9triver/iarnet-global/internal/domain/registry"
	"github.com/9triver/iarnet-global/internal/intra/repository"
	"github.com/sirupsen/logrus"
)

// bootstrapRegistry 初始化 Registry 模块
func bootstrapRegistry(ig *IarnetGlobal) error {
	// 创建 Registry Manager
	manager := registry.NewManager()
	dbConfig := ig.Config.Database
	// 初始化 Domain Repository
	var domainRepo repository.DomainRepo
	domainRepo, err := repository.NewDomainRepo(dbConfig.DomainDBPath, dbConfig.MaxOpenConns, dbConfig.MaxIdleConns, dbConfig.ConnMaxLifetimeSeconds)
	if err != nil {
		return fmt.Errorf("failed to initialize domain repository: %w", err)
	}
	// 创建 Registry Service
	service := registry.NewService(manager, domainRepo)

	// 从 repository 加载域数据到 manager
	ctx := context.Background()
	if err := service.LoadDomains(ctx); err != nil {
		return fmt.Errorf("failed to load domains from repository: %w", err)
	}

	ig.RegistryService = service
	ig.DomainManager = manager
	ig.DomainRepo = domainRepo
	logrus.Info("Registry module initialized")
	return nil
}
