package registry

// CreateDomainRequest 创建域请求
type CreateDomainRequest struct {
	Name        string `json:"name" binding:"required"`        // 域名称（必填）
	Description string `json:"description,omitempty"`          // 域描述（可选）
}

// CreateDomainResponse 创建域响应
type CreateDomainResponse struct {
	ID          string `json:"id"`          // 域 ID
	Name        string `json:"name"`        // 域名称
	Description string `json:"description"` // 域描述
	CreatedAt   string `json:"created_at"` // 创建时间
}

// UpdateDomainRequest 更新域请求
type UpdateDomainRequest struct {
	Name        string `json:"name,omitempty"`        // 域名称（可选）
	Description string `json:"description,omitempty"` // 域描述（可选）
}

// GetDomainsResponse 获取域列表响应
type GetDomainsResponse struct {
	Domains []DomainItem `json:"domains"` // 域列表
	Total   int          `json:"total"`   // 总数
}

// DomainItem 域列表项
type DomainItem struct {
	ID           string              `json:"id"`            // 域 ID
	Name         string              `json:"name"`          // 域名称
	Description  string              `json:"description"`  // 域描述
	NodeCount    int                 `json:"node_count"`   // 节点总数
	OnlineNodes  int                 `json:"online_nodes"` // 在线节点数
	ResourceTags ResourceTagsResponse `json:"resource_tags"` // 资源标签
	LastUpdated  string              `json:"last_updated"`  // 最后更新时间
}

// ResourceTagsResponse 资源标签响应（只显示是否支持，不显示具体数值）
type ResourceTagsResponse struct {
	CPU    bool `json:"cpu"`    // 是否支持 CPU
	GPU    bool `json:"gpu"`    // 是否支持 GPU
	Memory bool `json:"memory"` // 是否支持内存
	Camera bool `json:"camera"` // 是否支持摄像头
}

// GetDomainResponse 获取单个域响应
type GetDomainResponse struct {
	ID           string                `json:"id"`            // 域 ID
	Name         string                `json:"name"`          // 域名称
	Description  string                `json:"description"`  // 域描述
	ResourceTags ResourceTagsResponse  `json:"resource_tags"` // 资源标签
	Nodes        []NodeItem            `json:"nodes"`        // 节点列表
	LastUpdated  string                `json:"last_updated"`  // 最后更新时间
}

// GetDomainNodesResponse 获取域节点列表响应
type GetDomainNodesResponse struct {
	Nodes []NodeItem `json:"nodes"` // 节点列表
	Total int        `json:"total"` // 总数
}

// NodeItem 节点列表项
type NodeItem struct {
	ID           string                    `json:"id"`            // 节点 ID
	Name         string                    `json:"name"`         // 节点名称
	Address      string                    `json:"address"`      // 节点地址
	Status       string                    `json:"status"`       // 节点状态（online/offline/error）
	IsHead       bool                      `json:"is_head"`      // 是否为 head 节点
	ResourceTags *NodeResourceTagsResponse `json:"resource_tags,omitempty"` // 资源标签（显示具体数值）
	LastSeen     string                    `json:"last_seen"`    // 最后活跃时间
}

// NodeResourceTagsResponse 节点资源标签响应（显示具体数值）
type NodeResourceTagsResponse struct {
	CPU    *int64 `json:"cpu,omitempty"`    // CPU 核心数
	GPU    *int64 `json:"gpu,omitempty"`    // GPU 数量
	Memory *int64 `json:"memory,omitempty"` // 内存容量（字节）
	Camera *bool  `json:"camera,omitempty"` // 是否支持摄像头
}

