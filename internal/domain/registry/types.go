package registry

import "time"

// DomainID 域的唯一标识符
type DomainID = string

// NodeID iarnet 节点的唯一标识符
type NodeID = string

// ResourceTags 资源标签，描述域或节点支持的计算资源类型
type ResourceTags struct {
	CPU    bool `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	GPU    bool `json:"gpu,omitempty" yaml:"gpu,omitempty"`
	Memory bool `json:"memory,omitempty" yaml:"memory,omitempty"`
	Camera bool `json:"camera,omitempty" yaml:"camera,omitempty"`
}

func NewEmptyResourceTags() *ResourceTags {
	return NewResourceTags(false, false, false, false)
}

func NewResourceTags(cpu, gpu, memory, camera bool) *ResourceTags {
	return &ResourceTags{
		CPU:    cpu,
		GPU:    gpu,
		Memory: memory,
		Camera: camera,
	}
}

// HasResource 检查是否支持指定的资源类型
func (rt *ResourceTags) HasResource(resourceType string) bool {
	switch resourceType {
	case "cpu":
		return rt.CPU
	case "gpu":
		return rt.GPU
	case "memory":
		return rt.Memory
	case "camera":
		return rt.Camera
	default:
		return false
	}
}

// NodeStatus 节点状态
type NodeStatus string

const (
	// NodeStatusOnline 节点在线
	NodeStatusOnline NodeStatus = "online"
	// NodeStatusOffline 节点离线
	NodeStatusOffline NodeStatus = "offline"
	// NodeStatusError 节点错误
	NodeStatusError NodeStatus = "error"
)

// Node iarnet 节点信息
type Node struct {
	// ID 节点唯一标识符
	ID NodeID `json:"id" yaml:"id"`
	// DomainID 所属域的 ID
	DomainID DomainID `json:"domain_id" yaml:"domain_id"`
	// Name 节点名称
	Name string `json:"name" yaml:"name"`
	// Address 节点地址，格式：host:port，例如 "192.168.1.100:50051"
	Address string `json:"address" yaml:"address"`
	// IsHead 是否为 head 节点（全局调度器跨域调度的入口）
	IsHead bool `json:"is_head" yaml:"is_head"`
	// Status 节点状态
	Status NodeStatus `json:"status" yaml:"status"`
	// ResourceTags 节点支持的资源标签
	ResourceTags *ResourceTags `json:"resource_tags,omitempty" yaml:"resource_tags,omitempty"`
	// LastSeen 最后活跃时间
	LastSeen time.Time `json:"last_seen" yaml:"last_seen"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

// Domain 资源域信息
type Domain struct {
	// ID 域的唯一标识符
	ID DomainID `json:"id" yaml:"id"`
	// Name 域名称
	Name string `json:"name" yaml:"name"`
	// Description 域描述（可选）
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// ResourceTags 域的资源标签（汇总所有节点的资源标签）
	ResourceTags *ResourceTags `json:"resource_tags,omitempty" yaml:"resource_tags,omitempty"`
	// HeadNodeID head 节点的 ID（全局调度器跨域调度的入口）
	HeadNodeID *NodeID `json:"head_node_id,omitempty" yaml:"head_node_id,omitempty"`
	// NodeIDs 域下所有节点的 ID 列表
	NodeIDs []NodeID `json:"node_ids" yaml:"node_ids"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

// GetOnlineNodeCount 获取在线节点数量（需要从节点管理器获取）
// 这个方法需要外部传入节点状态信息，因为 Domain 本身不存储节点详情
func (d *Domain) GetOnlineNodeCount(getNodeStatus func(NodeID) NodeStatus) int {
	count := 0
	for _, nodeID := range d.NodeIDs {
		if getNodeStatus(nodeID) == NodeStatusOnline {
			count++
		}
	}
	return count
}

// GetTotalNodeCount 获取节点总数
func (d *Domain) GetTotalNodeCount() int {
	return len(d.NodeIDs)
}

// // UpdateResourceTags 更新域的资源标签（汇总所有节点的资源标签）
// func (d *Domain) UpdateResourceTags(getNodeResourceTags func(NodeID) *ResourceTags) {
// 	// 汇总所有节点的资源标签
// 	aggregatedTags := &ResourceTags{}

// 	for _, nodeID := range d.NodeIDs {
// 		nodeTags := getNodeResourceTags(nodeID)
// 		if nodeTags == nil {
// 			continue
// 		}

// 		// 汇总 CPU
// 		if nodeTags.CPU != nil {
// 			if aggregatedTags.CPU == nil {
// 				aggregatedTags.CPU = new(int64)
// 			}
// 			*aggregatedTags.CPU += *nodeTags.CPU
// 		}

// 		// 汇总 GPU
// 		if nodeTags.GPU != nil {
// 			if aggregatedTags.GPU == nil {
// 				aggregatedTags.GPU = new(int64)
// 			}
// 			*aggregatedTags.GPU += *nodeTags.GPU
// 		}

// 		// 汇总 Memory（取最大值，因为内存是容量概念）
// 		if nodeTags.Memory != nil {
// 			if aggregatedTags.Memory == nil {
// 				aggregatedTags.Memory = new(int64)
// 			}
// 			if *nodeTags.Memory > *aggregatedTags.Memory {
// 				*aggregatedTags.Memory = *nodeTags.Memory
// 			}
// 		}

// 		// 汇总 Camera（任意节点支持即支持）
// 		if nodeTags.Camera != nil && *nodeTags.Camera {
// 			if aggregatedTags.Camera == nil {
// 				aggregatedTags.Camera = new(bool)
// 			}
// 			*aggregatedTags.Camera = true
// 		}
// 	}

// 	d.ResourceTags = aggregatedTags
// }

// AddNode 添加节点到域
func (d *Domain) AddNode(nodeID NodeID) {
	// 检查节点是否已存在
	for _, id := range d.NodeIDs {
		if id == nodeID {
			return
		}
	}
	d.NodeIDs = append(d.NodeIDs, nodeID)
}

// RemoveNode 从域中移除节点
func (d *Domain) RemoveNode(nodeID NodeID) {
	for i, id := range d.NodeIDs {
		if id == nodeID {
			d.NodeIDs = append(d.NodeIDs[:i], d.NodeIDs[i+1:]...)
			// 如果移除的是 head 节点，清空 HeadNodeID
			if d.HeadNodeID != nil && *d.HeadNodeID == nodeID {
				d.HeadNodeID = nil
			}
			return
		}
	}
}

// SetHeadNode 设置 head 节点
func (d *Domain) SetHeadNode(nodeID NodeID) error {
	// 验证节点是否属于该域
	found := false
	for _, id := range d.NodeIDs {
		if id == nodeID {
			found = true
			break
		}
	}
	if !found {
		return ErrNodeNotInDomain
	}

	d.HeadNodeID = &nodeID
	return nil
}

// GetHeadNode 获取 head 节点 ID
func (d *Domain) GetHeadNode() *NodeID {
	return d.HeadNodeID
}
