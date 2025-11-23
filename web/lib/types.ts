// 域管理相关类型定义
export interface CreateDomainRequest {
  name: string        // 域名称（必填）
  description?: string // 域描述（可选）
}

export interface CreateDomainResponse {
  id: string          // 域 ID
  name: string        // 域名称
  description: string // 域描述
  created_at: string  // 创建时间
}

// API 返回的资源标签类型
export interface ResourceTagsResponse {
  cpu: boolean
  gpu: boolean
  memory: boolean
  camera: boolean
}

// API 返回的域列表项类型
export interface DomainItem {
  id: string
  name: string
  description: string
  node_count: number
  online_nodes: number
  resource_tags: ResourceTagsResponse
  created_at: string
  updated_at: string
}

// API 返回的域列表响应类型
export interface GetDomainsResponse {
  domains: DomainItem[]
  total: number
}

