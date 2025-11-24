"use client"

import { useState, useEffect } from "react"
import { useParams, useRouter } from "next/navigation"
import { Sidebar } from "@/components/sidebar"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { ArrowLeft, Cpu, MemoryStick, Camera, RefreshCw, Server, Activity, CheckCircle, XCircle, AlertTriangle, Crown } from "lucide-react"
import { formatMemory, formatNumber, formatRelativeTime, formatDateTime } from "@/lib/utils"
import { registryAPI, APIError } from "@/lib/api"
import type { GetDomainResponse, NodeItem } from "@/lib/types"

// iarnet 节点接口（前端组件使用的格式）
interface IarnetNode {
  id: string
  name: string
  address: string
  status: "online" | "offline" | "error"
  lastSeen: string
  isHead?: boolean // 是否为 head 节点（全局调度器跨域调度的入口）
  resourceTags?: {
    cpu?: number
    gpu?: number
    memory?: number
    camera?: boolean
  }
}

// 资源域详情接口（前端组件使用的格式）
interface DomainDetail {
  id: string
  name: string
  description?: string
  resourceTags: {
    cpu?: boolean
    gpu?: boolean
    memory?: boolean
    camera?: boolean
  }
  nodes: IarnetNode[]
  lastUpdated: string
}

export default function DomainDetailPage() {
  const params = useParams()
  const router = useRouter()
  const domainId = params.id as string

  const [domain, setDomain] = useState<DomainDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [autoRefresh, setAutoRefresh] = useState(false)

  const fetchDomainDetail = async () => {
    try {
      setLoading(true)
      const data = await registryAPI.getDomain(domainId)
      
      // 转换 API 响应数据为前端组件期望的格式
      const domainDetail: DomainDetail = {
        id: data.id,
        name: data.name,
        description: data.description || "",
        resourceTags: {
          cpu: data.resource_tags.cpu,
          gpu: data.resource_tags.gpu,
          memory: data.resource_tags.memory,
          camera: data.resource_tags.camera,
        },
        nodes: data.nodes.map((node: NodeItem) => ({
          id: node.id,
          name: node.name,
          address: node.address,
          status: node.status as "online" | "offline" | "error",
          lastSeen: node.last_seen,
          isHead: node.is_head,
          resourceTags: node.resource_tags ? {
            cpu: node.resource_tags.cpu,
            gpu: node.resource_tags.gpu,
            memory: node.resource_tags.memory,
            camera: node.resource_tags.camera,
          } : undefined,
        })),
        lastUpdated: data.updated_at,
      }
      
      setDomain(domainDetail)
    } catch (error) {
      console.error('Failed to fetch domain detail:', error)
      if (error instanceof APIError && error.status === 404) {
        // 域不存在，设置为 null 以显示错误页面
        setDomain(null)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (domainId) {
      fetchDomainDetail()
    }
  }, [domainId])

  useEffect(() => {
    if (!autoRefresh || !domainId) return

    const interval = setInterval(() => {
      fetchDomainDetail()
    }, 5000) // 每5秒刷新一次

    return () => clearInterval(interval)
  }, [autoRefresh, domainId])

  const getStatusIcon = (status: IarnetNode["status"]) => {
    switch (status) {
      case "online":
        return <CheckCircle className="h-4 w-4 text-green-500" />
      case "offline":
        return <XCircle className="h-4 w-4 text-gray-500" />
      case "error":
        return <AlertTriangle className="h-4 w-4 text-red-500" />
    }
  }

  const getStatusBadge = (status: IarnetNode["status"]) => {
    switch (status) {
      case "online":
        return <Badge variant="default" className="bg-green-500">在线</Badge>
      case "offline":
        return <Badge variant="secondary">离线</Badge>
      case "error":
        return <Badge variant="destructive">错误</Badge>
    }
  }

  if (loading) {
    return (
      <div className="flex h-screen bg-background">
        <Sidebar />
        <main className="flex-1 overflow-auto flex items-center justify-center">
          <div className="text-muted-foreground">加载中...</div>
        </main>
      </div>
    )
  }

  if (!domain) {
    return (
      <div className="flex h-screen bg-background">
        <Sidebar />
        <main className="flex-1 overflow-auto flex items-center justify-center">
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>资源域不存在</CardTitle>
              <CardDescription>未找到指定的资源域</CardDescription>
            </CardHeader>
            <CardContent>
              <Button onClick={() => router.push("/")}>返回首页</Button>
            </CardContent>
          </Card>
        </main>
      </div>
    )
  }

  const onlineNodes = domain.nodes.filter(n => n.status === "online").length
  const offlineNodes = domain.nodes.filter(n => n.status === "offline").length
  const errorNodes = domain.nodes.filter(n => n.status === "error").length

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />

      <main className="flex-1 overflow-auto">
        <div className="p-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center space-x-4">
              <Button variant="ghost" size="sm" onClick={() => router.push("/")}>
                <ArrowLeft className="h-4 w-4 mr-2" />
                返回
              </Button>
              <div>
                <h1 className="text-3xl font-playfair font-bold text-foreground mb-2">{domain.name}</h1>
                <p className="text-muted-foreground">{domain.description || "无描述"}</p>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setAutoRefresh(!autoRefresh)}
                className={autoRefresh ? "bg-green-50 border-green-200" : ""}
              >
                <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? "animate-spin" : ""}`} />
                {autoRefresh ? "自动刷新" : "手动刷新"}
              </Button>
              <Button
                variant="outline"
                onClick={fetchDomainDetail}
                disabled={loading}
              >
                <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                刷新数据
              </Button>
            </div>
          </div>

          {/* Domain Resource Tags */}
          <Card className="mb-8">
            <CardHeader>
              <CardTitle>资源标签</CardTitle>
              <CardDescription>该资源域支持的计算资源类型</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-3">
                {domain.resourceTags.cpu && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <Cpu className="h-4 w-4" />
                    CPU
                  </Badge>
                )}
                {domain.resourceTags.gpu && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <Activity className="h-4 w-4" />
                    GPU
                  </Badge>
                )}
                {domain.resourceTags.memory && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <MemoryStick className="h-4 w-4" />
                    内存
                  </Badge>
                )}
                {domain.resourceTags.camera && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <Camera className="h-4 w-4" />
                    摄像头
                  </Badge>
                )}
                {!domain.resourceTags.cpu && !domain.resourceTags.gpu && !domain.resourceTags.memory && !domain.resourceTags.camera && (
                  <span className="text-sm text-muted-foreground">暂无资源标签</span>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">节点总数</CardTitle>
                <Server className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{domain.nodes.length}</div>
                <p className="text-xs text-muted-foreground">该域下的所有节点</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">在线节点</CardTitle>
                <CheckCircle className="h-4 w-4 text-green-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">{onlineNodes}</div>
                <p className="text-xs text-muted-foreground">正常运行中</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">离线节点</CardTitle>
                <XCircle className="h-4 w-4 text-gray-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-gray-600">{offlineNodes}</div>
                <p className="text-xs text-muted-foreground">已断开连接</p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">错误节点</CardTitle>
                <AlertTriangle className="h-4 w-4 text-red-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">{errorNodes}</div>
                <p className="text-xs text-muted-foreground">需要处理</p>
              </CardContent>
            </Card>
          </div>

          {/* Nodes Table */}
          <Card>
            <CardHeader>
              <CardTitle>iarnet 节点列表</CardTitle>
              <CardDescription>该资源域下的所有 iarnet 节点及其状态</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>节点名称</TableHead>
                    <TableHead>地址</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>资源标签</TableHead>
                    <TableHead>最后活跃</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {domain.nodes.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                        该域下暂无节点
                      </TableCell>
                    </TableRow>
                  ) : (
                    domain.nodes.map((node) => (
                      <TableRow key={node.id}>
                        <TableCell>
                          <div className="flex items-center space-x-2">
                            {getStatusIcon(node.status)}
                            <div className="flex flex-col">
                              <div className="font-medium">{node.name}</div>
                              <div className="text-xs text-muted-foreground font-mono">{node.id}</div>
                            </div>
                            {node.isHead && (
                              <Badge variant="default" className="bg-blue-500 flex items-center gap-1">
                                <Crown className="h-3 w-3" />
                                Head
                              </Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm font-mono">{node.address}</div>
                        </TableCell>
                        <TableCell>
                          {getStatusBadge(node.status)}
                        </TableCell>
                        <TableCell>
                          {node.resourceTags && (
                            <div className="flex flex-wrap gap-2">
                              {node.resourceTags.cpu !== undefined && (
                                <Badge variant="outline" className="flex items-center gap-1">
                                  <Cpu className="h-3 w-3" />
                                  {formatNumber(node.resourceTags.cpu)}核
                                </Badge>
                              )}
                              {node.resourceTags.gpu !== undefined && (
                                <Badge variant="outline" className="flex items-center gap-1">
                                  <Activity className="h-3 w-3" />
                                  {node.resourceTags.gpu}个
                                </Badge>
                              )}
                              {node.resourceTags.memory !== undefined && (
                                <Badge variant="outline" className="flex items-center gap-1">
                                  <MemoryStick className="h-3 w-3" />
                                  {formatMemory(node.resourceTags.memory)}
                                </Badge>
                              )}
                              {node.resourceTags.camera && (
                                <Badge variant="outline" className="flex items-center gap-1">
                                  <Camera className="h-3 w-3" />
                                  摄像头
                                </Badge>
                              )}
                            </div>
                          )}
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {formatRelativeTime(node.lastSeen)}
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  )
}

