package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/9triver/iarnet-global/internal/intra/repository"
	"github.com/9triver/iarnet-global/internal/util"
)

// Service 域注册服务接口
// 提供无状态的域操作服务，所有状态由 Manager 管理
type Service interface {
	// CreateDomain 创建域
	CreateDomain(ctx context.Context, name, description string) (*Domain, error)

	// GetDomain 获取域
	GetDomain(ctx context.Context, domainID DomainID) (*Domain, error)

	// GetAllDomains 获取所有域
	GetAllDomains(ctx context.Context) ([]*Domain, error)

	// UpdateDomain 更新域信息
	UpdateDomain(ctx context.Context, domainID DomainID, name, description string) error

	// DeleteDomain 删除域
	DeleteDomain(ctx context.Context, domainID DomainID) error

	// GetDomainNodes 获取域下的所有节点
	GetDomainNodes(ctx context.Context, domainID DomainID) ([]*Node, error)

	// GetDomainStats 获取域的统计信息（节点数量等）
	GetDomainStats(ctx context.Context, domainID DomainID) (*DomainStats, error)

	// LoadDomains 从 repository 加载所有域数据到 manager
	LoadDomains(ctx context.Context) error
}

// DomainStats 域统计信息
type DomainStats struct {
	TotalNodes   int // 节点总数
	OnlineNodes  int // 在线节点数
	OfflineNodes int // 离线节点数
	ErrorNodes   int // 错误节点数
}

type service struct {
	manager    *Manager
	domainRepo repository.DomainRepo
}

// NewService 创建域注册服务
func NewService(manager *Manager, domainRepo repository.DomainRepo) Service {
	return &service{
		manager:    manager,
		domainRepo: domainRepo,
	}
}

// CreateDomain 创建域
func (s *service) CreateDomain(ctx context.Context, name, description string) (*Domain, error) {
	// 验证输入
	if name == "" {
		return nil, ErrInvalidResourceTags // 暂时复用错误，后续可以定义更具体的错误
	}

	// 生成域 ID
	domainID := DomainID(util.GenIDWith("domain."))

	// 创建域
	domain := &Domain{
		ID:          domainID,
		Name:        name,
		Description: description,
		ResourceTags: &ResourceTags{
			CPU:    false,
			GPU:    false,
			Memory: false,
			Camera: false,
		},
		NodeIDs:   make([]NodeID, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.domainRepo.CreateDomain(ctx, &repository.DomainDAO{
		ID:          domain.ID,
		Name:        domain.Name,
		Description: domain.Description,
		CreatedAt:   domain.CreatedAt,
		UpdatedAt:   domain.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist domain to repository: %w", err)
	}

	// 添加到管理器
	if err := s.manager.AddDomain(domain); err != nil {
		return nil, err
	}

	return domain, nil
}

// GetDomain 获取域
func (s *service) GetDomain(ctx context.Context, domainID DomainID) (*Domain, error) {
	return s.manager.GetDomain(domainID)
}

// GetAllDomains 获取所有域
func (s *service) GetAllDomains(ctx context.Context) ([]*Domain, error) {
	return s.manager.GetAllDomains(), nil
}

// UpdateDomain 更新域信息
func (s *service) UpdateDomain(ctx context.Context, domainID DomainID, name, description string) error {
	domain, err := s.manager.GetDomain(domainID)
	if err != nil {
		return err
	}

	// 更新字段
	if name != "" {
		domain.Name = name
	}
	if description != "" {
		domain.Description = description
	}
	domain.UpdatedAt = time.Now()

	return nil
}

// DeleteDomain 删除域
func (s *service) DeleteDomain(ctx context.Context, domainID DomainID) error {
	return s.manager.RemoveDomain(domainID)
}

// GetDomainNodes 获取域下的所有节点
func (s *service) GetDomainNodes(ctx context.Context, domainID DomainID) ([]*Node, error) {
	return s.manager.GetNodesByDomain(domainID)
}

// GetDomainStats 获取域的统计信息
func (s *service) GetDomainStats(ctx context.Context, domainID DomainID) (*DomainStats, error) {
	domain, err := s.manager.GetDomain(domainID)
	if err != nil {
		return nil, err
	}

	stats := &DomainStats{
		TotalNodes: len(domain.NodeIDs),
	}

	// 统计各状态节点数量
	for _, nodeID := range domain.NodeIDs {
		status := s.manager.GetNodeStatus(nodeID)
		switch status {
		case NodeStatusOnline:
			stats.OnlineNodes++
		case NodeStatusOffline:
			stats.OfflineNodes++
		case NodeStatusError:
			stats.ErrorNodes++
		}
	}

	return stats, nil
}

// LoadDomains 从 repository 加载所有域数据到 manager
func (s *service) LoadDomains(ctx context.Context) error {
	// 从 repository 获取所有域
	domainDAOs, err := s.domainRepo.GetAllDomains(ctx)
	if err != nil {
		return fmt.Errorf("failed to load domains from repository: %w", err)
	}

	// 将 DomainDAO 转换为 Domain 并添加到 manager
	for _, dao := range domainDAOs {
		domain := &Domain{
			ID:          DomainID(dao.ID),
			Name:        dao.Name,
			Description: dao.Description,
			NodeIDs:     make([]NodeID, 0), // 节点信息需要从其他地方加载
			ResourceTags: &ResourceTags{
				CPU:    false,
				GPU:    false,
				Memory: false,
				Camera: false,
			},
			CreatedAt: dao.CreatedAt,
			UpdatedAt: dao.UpdatedAt,
		}

		// 添加到管理器（如果已存在则跳过，避免重复加载）
		if err := s.manager.AddDomain(domain); err != nil {
			// 如果域已存在，记录警告但继续处理其他域
			if err == ErrDomainAlreadyExists {
				continue
			}
			return fmt.Errorf("failed to add domain %s to manager: %w", domain.ID, err)
		}
	}

	return nil
}
