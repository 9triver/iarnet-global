// API 客户端工具函数
import type { CreateDomainRequest, CreateDomainResponse, GetDomainsResponse, GetDomainResponse } from "./types"

const API_BASE = "/api"

export class APIError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = "APIError"
  }
}

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`

  const response = await fetch(url, {
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
    ...options,
  })

  // 检查响应是否有内容
  const contentType = response.headers.get("content-type")
  const hasJsonContent = contentType && contentType.includes("application/json")
  
  let data: any = {}
  
  // 只有在响应有内容且是 JSON 格式时才解析
  if (hasJsonContent) {
    try {
      const text = await response.text()
      if (text.trim()) {
        data = JSON.parse(text)
      }
    } catch (error) {
      // JSON 解析失败，尝试从状态码获取错误信息
      if (!response.ok) {
        throw new APIError(response.status, `请求失败: ${response.statusText}`)
      }
      throw new APIError(response.status, "响应格式错误")
    }
  } else if (!response.ok) {
    // 非 JSON 响应且状态码不是成功，直接抛出错误
    throw new APIError(response.status, response.statusText || "请求失败")
  }

  if (!response.ok) {
    throw new APIError(response.status, data.message || data.error || "API request failed")
  }

  // 处理后端标准响应格式 {code, message, data}
  if (data.code !== undefined) {
    if (data.code < 200 || data.code >= 300) {
      throw new APIError(data.code, data.message || data.error || "API request failed")
    }
    return data.data
  }

  // 兼容其他响应格式
  return data.data || data
}

export const registryAPI = {
  // 获取所有域
  getDomains: () =>
    apiRequest<GetDomainsResponse>("/registry/domains", {
      method: "GET",
    }),

  // 获取单个域详情
  getDomain: (id: string) =>
    apiRequest<GetDomainResponse>(`/registry/domains/${id}`, {
      method: "GET",
    }),

  // 创建域
  createDomain: (request: CreateDomainRequest) =>
    apiRequest<CreateDomainResponse>("/registry/domains", {
      method: "POST",
      body: JSON.stringify(request),
    }),
}

// 日志相关类型
export interface LogEntry {
  timestamp: string
  level: string
  message: string
  fields?: Record<string, any>
  caller?: {
    file: string
    line: number
    function: string
  }
}

export interface GetLogsResponse {
  logs: LogEntry[]
  total: number
  start: number
  limit: number
}

export const logsAPI = {
  // 获取日志
  getLogs: (params?: { start?: number; limit?: number; level?: string }) => {
    const queryParams = new URLSearchParams()
    if (params?.start !== undefined) {
      queryParams.append("start", params.start.toString())
    }
    if (params?.limit !== undefined) {
      queryParams.append("limit", params.limit.toString())
    }
    if (params?.level) {
      queryParams.append("level", params.level)
    }
    const query = queryParams.toString()
    return apiRequest<GetLogsResponse>(`/logs${query ? `?${query}` : ""}`, {
      method: "GET",
    })
  },

  // 清空日志
  clearLogs: () =>
    apiRequest<void>("/logs/clear", {
      method: "POST",
    }),
}
