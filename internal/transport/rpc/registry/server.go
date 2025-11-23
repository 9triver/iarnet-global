package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/9triver/iarnet-global/internal/domain/registry"
	registrypb "github.com/9triver/iarnet-global/internal/proto/registry"
	"github.com/sirupsen/logrus"
)

// Server RPC 服务器实现
type Server struct {
	registrypb.UnimplementedServiceServer
	manager *registry.Manager
}

// NewServer 创建新的 RPC 服务器
func NewServer(manager *registry.Manager) *Server {
	return &Server{
		manager: manager,
	}
}

// RegisterNode 注册节点到全局注册中心
func (s *Server) RegisterNode(ctx context.Context, req *registrypb.RegisterNodeRequest) (*registrypb.RegisterNodeResponse, error) {
	// 验证请求
	if req.DomainId == "" {
		return nil, fmt.Errorf("domain_id is required")
	}
	if req.NodeId == "" {
		return nil, fmt.Errorf("node_id is required")
	}
	if req.NodeName == "" {
		return nil, fmt.Errorf("node_name is required")
	}

	// 检查域是否存在
	domain, err := s.manager.GetDomain(registry.DomainID(req.DomainId))
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// 创建节点
	node := &registry.Node{
		ID:           registry.NodeID(req.NodeId),
		DomainID:     registry.DomainID(req.DomainId),
		Name:         req.NodeName,
		Address:      "",    // 地址在健康检查时更新
		IsHead:       false, // 默认不是 head 节点
		Status:       registry.NodeStatusOffline,
		ResourceTags: registry.NewEmptyResourceTags(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	// 添加节点到管理器
	if err := s.manager.AddNode(node); err != nil {
		return nil, fmt.Errorf("failed to add node: %w", err)
	}

	logrus.Infof("Node registered: id=%s, name=%s, domain=%s", req.NodeId, req.NodeName, req.DomainId)

	return &registrypb.RegisterNodeResponse{
		DomainName:        domain.Name,
		DomainDescription: domain.Description,
	}, nil
}

// HealthCheck 节点健康检查，定期上报节点状态和资源使用情况
func (s *Server) HealthCheck(ctx context.Context, req *registrypb.HealthCheckRequest) (*registrypb.HealthCheckResponse, error) {
	// 验证请求
	if req.NodeId == "" {
		return nil, fmt.Errorf("node_id is required")
	}
	if req.DomainId == "" {
		return nil, fmt.Errorf("domain_id is required")
	}

	nodeID := registry.NodeID(req.NodeId)
	domainID := registry.DomainID(req.DomainId)

	// 检查节点是否存在，如果不存在则自动注册
	node, err := s.manager.GetNode(nodeID)
	if err != nil {
		// 节点不存在，尝试自动注册
		_, err := s.manager.GetDomain(domainID)
		if err != nil {
			return nil, fmt.Errorf("domain not found: %w", err)
		}

		// 创建新节点
		node = &registry.Node{
			ID:           nodeID,
			DomainID:     domainID,
			Name:         req.NodeId, // 使用 node_id 作为默认名称
			Address:      req.Address,
			IsHead:       req.IsHead,
			Status:       convertProtoNodeStatus(req.Status),
			ResourceTags: convertProtoResourceTags(req.ResourceTags),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			LastSeen:     time.Now(),
		}

		if err := s.manager.AddNode(node); err != nil {
			return nil, fmt.Errorf("failed to auto-register node: %w", err)
		}

		logrus.Infof("Node auto-registered during health check: id=%s, domain=%s", req.NodeId, req.DomainId)
	} else {
		// 更新现有节点
		err := s.manager.UpdateNode(nodeID, func(n *registry.Node) {
			// 更新状态
			n.Status = convertProtoNodeStatus(req.Status)
			n.LastSeen = time.Now()
			n.UpdatedAt = time.Now()

			// 更新地址（如果提供）
			if req.Address != "" {
				n.Address = req.Address
			}

			// 更新资源标签（如果提供）
			if req.ResourceTags != nil {
				n.ResourceTags = convertProtoResourceTags(req.ResourceTags)
			}

			// 更新 head 节点状态
			if req.IsHead {
				n.IsHead = true
			}
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update node: %w", err)
		}

		// 更新节点状态（确保状态同步）
		if err := s.manager.UpdateNodeStatus(nodeID, convertProtoNodeStatus(req.Status)); err != nil {
			logrus.Warnf("Failed to update node status: %v", err)
		}
	}

	// 构建响应
	response := &registrypb.HealthCheckResponse{
		ServerTimestamp:            time.Now().UnixNano(),
		RecommendedIntervalSeconds: 30, // 建议 30 秒检查一次
		RequireReregister:          false,
		StatusCode:                 "success",
		Message:                    "Health check processed successfully",
	}

	return response, nil
}

// convertProtoNodeStatus 将 proto NodeStatus 转换为 domain NodeStatus
func convertProtoNodeStatus(status registrypb.NodeStatus) registry.NodeStatus {
	switch status {
	case registrypb.NodeStatus_NODE_STATUS_ONLINE:
		return registry.NodeStatusOnline
	case registrypb.NodeStatus_NODE_STATUS_OFFLINE:
		return registry.NodeStatusOffline
	case registrypb.NodeStatus_NODE_STATUS_ERROR:
		return registry.NodeStatusError
	default:
		return registry.NodeStatusOffline
	}
}

// convertProtoResourceTags 将 proto ResourceTags 转换为 domain ResourceTags
func convertProtoResourceTags(tags *registrypb.ResourceTags) *registry.ResourceTags {
	if tags == nil {
		return registry.NewEmptyResourceTags()
	}
	return registry.NewResourceTags(tags.Cpu, tags.Gpu, tags.Memory, tags.Camera)
}
