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

