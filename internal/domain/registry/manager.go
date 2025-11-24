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
	mu              sync.RWMutex
	domains         map[DomainID]*Domain
	nodes           map[NodeID]*Node
	healthCheckStop chan struct{} // 用于停止健康检查超时监控
	timeoutDuration time.Duration // 节点超时时间（默认 90 秒）
	cleanupDuration time.Duration // 节点清理时间（默认 180 秒，即超时时间的2倍）
}

// NewManager 创建新的管理器
func NewManager() *Manager {
	timeoutDuration := 30 * time.Second // 默认 30 秒超时（便于调试，生产环境建议 90 秒）
	return &Manager{
		domains:         make(map[DomainID]*Domain),
		nodes:           make(map[NodeID]*Node),
		healthCheckStop: make(chan struct{}),
		timeoutDuration: timeoutDuration,
		cleanupDuration: timeoutDuration * 2, // 清理时间 = 超时时间的2倍（节点离线后60秒才删除）
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

// GetHeadNodes 获取所有 head 节点（返回副本以避免竞态）
func (m *Manager) GetHeadNodes() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]*Node, 0)
	for _, node := range m.nodes {
		if node.IsHead {
			nodes = append(nodes, node.Clone())
		}
	}
	return nodes
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

// updateDomainResourceTags 更新域的资源标签（汇总所有节点的资源标签）
func (m *Manager) updateDomainResourceTags(domain *Domain) {
	// 汇总所有节点的资源标签
	aggregatedTags := &ResourceTags{
		CPU:    false,
		GPU:    false,
		Memory: false,
		Camera: false,
	}

	for _, nodeID := range domain.NodeIDs {
		node, ok := m.nodes[nodeID]
		if !ok {
			continue
		}

		if node.ResourceTags == nil {
			continue
		}

		// 汇总资源标签（任意节点支持即支持）
		if node.ResourceTags.CPU {
			aggregatedTags.CPU = true
		}
		if node.ResourceTags.GPU {
			aggregatedTags.GPU = true
		}
		if node.ResourceTags.Memory {
			aggregatedTags.Memory = true
		}
		if node.ResourceTags.Camera {
			aggregatedTags.Camera = true
		}
	}

	domain.ResourceTags = aggregatedTags
	domain.UpdatedAt = time.Now()
}

// Start 启动管理器（启动节点超时检测）
func (m *Manager) Start(ctx context.Context) error {
	logrus.Info("Registry manager started")

	// 启动节点超时检测 goroutine
	go m.startHealthCheckTimeoutMonitor(ctx)

	return nil
}

// startHealthCheckTimeoutMonitor 启动健康检查超时监控
// 定期检查所有节点的 LastSeen 时间，如果超过超时时间，标记为离线
func (m *Manager) startHealthCheckTimeoutMonitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 每 10 秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkNodeTimeouts()
		case <-m.healthCheckStop:
			logrus.Info("Health check timeout monitor stopped")
			return
		case <-ctx.Done():
			logrus.Info("Health check timeout monitor stopped due to context cancellation")
			return
		}
	}
}

// checkNodeTimeouts 检查所有节点的超时状态，并清理长时间离线的节点
func (m *Manager) checkNodeTimeouts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	timeoutCount := 0
	cleanupCount := 0
	nodesToRemove := make([]NodeID, 0)

	for nodeID, node := range m.nodes {
		// 检查是否应该清理（节点离线超过清理时间）
		if node.Status == NodeStatusOffline || node.Status == NodeStatusError {
			// 计算节点离线时长（从 LastSeen 开始计算）
			offlineDuration := now.Sub(node.LastSeen)
			if offlineDuration > m.cleanupDuration {
				// 标记为待删除
				nodesToRemove = append(nodesToRemove, nodeID)
				cleanupCount++
				logrus.Infof("Node %s (domain: %s) will be removed due to extended offline (offline for %v, last seen: %v)",
					nodeID, node.DomainID, offlineDuration, node.LastSeen)
				continue
			}
		}

		// 检查在线节点是否超时
		if node.Status == NodeStatusOnline {
			// 检查是否超时
			if now.Sub(node.LastSeen) > m.timeoutDuration {
				// 标记为离线
				node.Status = NodeStatusOffline
				node.UpdatedAt = now
				timeoutCount++

				logrus.Warnf("Node %s (domain: %s) marked as offline due to timeout (last seen: %v)",
					nodeID, node.DomainID, node.LastSeen)

				// 更新域的资源标签
				if domain, ok := m.domains[node.DomainID]; ok {
					m.updateDomainResourceTagsUnsafe(domain)
				}
			}
		}
	}

	// 删除超时节点
	for _, nodeID := range nodesToRemove {
		if err := m.removeNodeUnsafe(nodeID); err != nil {
			logrus.Errorf("Failed to remove timeout node %s: %v", nodeID, err)
		}
	}

	if timeoutCount > 0 {
		logrus.Debugf("Marked %d node(s) as offline due to timeout", timeoutCount)
	}
	if cleanupCount > 0 {
		logrus.Infof("Removed %d node(s) due to extended offline", cleanupCount)
	}
}

// removeNodeUnsafe 移除节点（不加锁版本，调用者需确保已持有锁）
func (m *Manager) removeNodeUnsafe(nodeID NodeID) error {
	node, ok := m.nodes[nodeID]
	if !ok {
		return ErrNodeNotFound
	}

	domain, ok := m.domains[node.DomainID]
	if ok {
		domain.RemoveNode(nodeID)
		m.updateDomainResourceTagsUnsafe(domain)
	}

	delete(m.nodes, nodeID)
	logrus.Infof("Node removed: id=%s, name=%s, domain=%s", nodeID, node.Name, node.DomainID)
	return nil
}

// updateDomainResourceTagsUnsafe 更新域的资源标签（不加锁版本，调用者需确保已持有锁）
func (m *Manager) updateDomainResourceTagsUnsafe(domain *Domain) {
	// 汇总所有节点的资源标签
	aggregatedTags := &ResourceTags{
		CPU:    false,
		GPU:    false,
		Memory: false,
		Camera: false,
	}

	for _, nodeID := range domain.NodeIDs {
		node, ok := m.nodes[nodeID]
		if !ok {
			continue
		}

		if node.ResourceTags == nil {
			continue
		}

		// 汇总资源标签（任意节点支持即支持）
		if node.ResourceTags.CPU {
			aggregatedTags.CPU = true
		}
		if node.ResourceTags.GPU {
			aggregatedTags.GPU = true
		}
		if node.ResourceTags.Memory {
			aggregatedTags.Memory = true
		}
		if node.ResourceTags.Camera {
			aggregatedTags.Camera = true
		}
	}

	domain.ResourceTags = aggregatedTags
	domain.UpdatedAt = time.Now()
}

// Stop 停止管理器
func (m *Manager) Stop() {
	close(m.healthCheckStop)
	logrus.Info("Registry manager stopped")
}
