package registry

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager 域和节点管理器
// 负责管理域和节点的状态，提供线程安全的操作
type Manager struct {
	mu      sync.RWMutex
	domains map[DomainID]*Domain
	nodes   map[NodeID]*Node
}

// NewManager 创建新的管理器
func NewManager() *Manager {
	return &Manager{
		domains: make(map[DomainID]*Domain),
		nodes:   make(map[NodeID]*Node),
	}
}

// GetDomain 获取域
func (m *Manager) GetDomain(domainID DomainID) (*Domain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domain, ok := m.domains[domainID]
	if !ok {
		return nil, ErrDomainNotFound
	}
	return domain, nil
}

// GetAllDomains 获取所有域
func (m *Manager) GetAllDomains() []*Domain {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]*Domain, 0, len(m.domains))
	for _, domain := range m.domains {
		domains = append(domains, domain)
	}
	return domains
}

// AddDomain 添加域
func (m *Manager) AddDomain(domain *Domain) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.domains[domain.ID]; exists {
		return ErrDomainAlreadyExists
	}

	m.domains[domain.ID] = domain
	logrus.Infof("Domain added: id=%s, name=%s", domain.ID, domain.Name)
	return nil
}

// RemoveDomain 移除域
func (m *Manager) RemoveDomain(domainID DomainID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	domain, ok := m.domains[domainID]
	if !ok {
		return ErrDomainNotFound
	}

	// 移除域下的所有节点
	for _, nodeID := range domain.NodeIDs {
		delete(m.nodes, nodeID)
	}

	delete(m.domains, domainID)
	logrus.Infof("Domain removed: id=%s, name=%s", domainID, domain.Name)
	return nil
}

// GetNode 获取节点
func (m *Manager) GetNode(nodeID NodeID) (*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return nil, ErrNodeNotFound
	}
	return node, nil
}

// GetNodesByDomain 获取域下的所有节点
func (m *Manager) GetNodesByDomain(domainID DomainID) ([]*Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domain, ok := m.domains[domainID]
	if !ok {
		return nil, ErrDomainNotFound
	}

	nodes := make([]*Node, 0, len(domain.NodeIDs))
	for _, nodeID := range domain.NodeIDs {
		if node, ok := m.nodes[nodeID]; ok {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

// AddNode 添加节点
func (m *Manager) AddNode(node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[node.ID]; exists {
		return ErrNodeAlreadyExists
	}

	// 验证域是否存在
	domain, ok := m.domains[node.DomainID]
	if !ok {
		return ErrDomainNotFound
	}

	// 添加节点到管理器
	m.nodes[node.ID] = node

	// 添加节点到域
	domain.AddNode(node.ID)

	// 如果是 head 节点，设置域的 head 节点
	if node.IsHead {
		if err := domain.SetHeadNode(node.ID); err != nil {
			// 回滚：移除节点
			delete(m.nodes, node.ID)
			domain.RemoveNode(node.ID)
			return err
		}
	}

	// 更新域的资源标签
	m.updateDomainResourceTags(domain)

	logrus.Infof("Node added: id=%s, name=%s, domain=%s, isHead=%v", node.ID, node.Name, node.DomainID, node.IsHead)
	return nil
}

// UpdateNode 更新节点
func (m *Manager) UpdateNode(nodeID NodeID, updateFn func(*Node)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return ErrNodeNotFound
	}

	updateFn(node)
	node.UpdatedAt = time.Now()

	// 更新域的资源标签
	domain, ok := m.domains[node.DomainID]
	if ok {
		m.updateDomainResourceTags(domain)
	}

	logrus.Debugf("Node updated: id=%s", nodeID)
	return nil
}

// RemoveNode 移除节点
func (m *Manager) RemoveNode(nodeID NodeID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return ErrNodeNotFound
	}

	domain, ok := m.domains[node.DomainID]
	if ok {
		domain.RemoveNode(nodeID)
		m.updateDomainResourceTags(domain)
	}

	delete(m.nodes, nodeID)
	logrus.Infof("Node removed: id=%s, name=%s", nodeID, node.Name)
	return nil
}

// UpdateNodeStatus 更新节点状态
func (m *Manager) UpdateNodeStatus(nodeID NodeID, status NodeStatus) error {
	return m.UpdateNode(nodeID, func(node *Node) {
		node.Status = status
		node.LastSeen = time.Now()
	})
}

// GetNodeStatus 获取节点状态（用于 Domain.GetOnlineNodeCount）
func (m *Manager) GetNodeStatus(nodeID NodeID) NodeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return NodeStatusOffline
	}
	return node.Status
}

// GetNodeResourceTags 获取节点资源标签（用于 Domain.UpdateResourceTags）
func (m *Manager) GetNodeResourceTags(nodeID NodeID) *ResourceTags {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return nil
	}
	return node.ResourceTags
}

// updateDomainResourceTags 更新域的资源标签
func (m *Manager) updateDomainResourceTags(domain *Domain) {
	// domain.UpdateResourceTags(m.GetNodeResourceTags)
	// domain.UpdatedAt = time.Now()
}

// Start 启动管理器（预留接口，用于后续扩展）
func (m *Manager) Start(ctx context.Context) error {
	logrus.Info("Registry manager started")
	return nil
}
