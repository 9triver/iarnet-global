"use client"

import { useState, useEffect, useMemo } from "react"
import { Sidebar } from "@/components/sidebar"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Globe, Cpu, MemoryStick, Camera, RefreshCw, Server, Activity, CheckCircle, XCircle, AlertTriangle, Plus } from "lucide-react"
import Link from "next/link"
import { formatMemory, formatNumber, formatRelativeTime } from "@/lib/utils"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { registryAPI, APIError } from "@/lib/api"
import type { DomainItem } from "@/lib/types"

// 资源域接口
interface ResourceDomain {
  id: string
  name: string
  description?: string
  nodeCount: number
  onlineNodes: number
  resourceTags: {
    cpu?: number
    gpu?: number
    memory?: number
    camera?: boolean
  }
  lastUpdated: string
}

// iarnet 节点接口
interface IarnetNode {
  id: string
  domainId: string
  name: string
  address: string
  status: "online" | "offline" | "error"
  lastSeen: string
}

export default function HomePage() {
  const [domains, setDomains] = useState<ResourceDomain[]>([])
  const [loading, setLoading] = useState(true)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [domainName, setDomainName] = useState("")
  const [domainDescription, setDomainDescription] = useState("")
  const [creating, setCreating] = useState(false)

  // Mock 数据
  const mockDomains: ResourceDomain[] = [
    {
      id: "domain-1",
      name: "生产环境域",
      description: "生产环境的算力资源域",
      nodeCount: 5,
      onlineNodes: 4,
      resourceTags: {
        cpu: 128,
        gpu: 16,
        memory: 512 * 1024 * 1024 * 1024, // 512GB
        camera: true,
      },
      lastUpdated: new Date().toISOString(),
    },
    {
      id: "domain-2",
      name: "开发环境域",
      description: "开发测试环境的算力资源域",
      nodeCount: 3,
      onlineNodes: 3,
      resourceTags: {
        cpu: 64,
        gpu: 8,
        memory: 256 * 1024 * 1024 * 1024, // 256GB
        camera: false,
      },
      lastUpdated: new Date().toISOString(),
    },
    {
      id: "domain-3",
      name: "边缘计算域",
      description: "边缘节点的算力资源域",
      nodeCount: 8,
      onlineNodes: 7,
      resourceTags: {
        cpu: 32,
        gpu: 4,
        memory: 128 * 1024 * 1024 * 1024, // 128GB
        camera: true,
      },
      lastUpdated: new Date().toISOString(),
    },
  ]

  // 将 API 返回的 DomainItem 转换为前端需要的 ResourceDomain 格式
  const convertDomainItem = (item: DomainItem): ResourceDomain => {
    return {
      id: item.id,
      name: item.name,
      description: item.description,
      nodeCount: item.node_count,
      onlineNodes: item.online_nodes,
      resourceTags: {
        cpu: item.resource_tags.cpu ? 1 : undefined,
        gpu: item.resource_tags.gpu ? 1 : undefined,
        memory: item.resource_tags.memory ? 1 : undefined,
        camera: item.resource_tags.camera || undefined,
      },
      lastUpdated: item.updated_at,
    }
  }

  // 从 API 获取域数据，并在真实数据后追加 mock 数据
  const fetchDomains = async () => {
    try {
      setLoading(true)
      
      // 调用真实 API
      let realDomains: ResourceDomain[] = []
      try {
        const response = await registryAPI.getDomains()
        realDomains = response.domains.map(convertDomainItem)
      } catch (error) {
        console.error('Failed to fetch domains from API:', error)
        // API 调用失败时，只使用 mock 数据
      }
      
      // 在真实数据后追加 mock 数据
      const allDomains = [...realDomains, ...mockDomains]
      setDomains(allDomains)
    } catch (error) {
      console.error('Failed to fetch domains:', error)
      // 出错时使用 mock 数据
      setDomains(mockDomains)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchDomains()
  }, [])

  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      fetchDomains()
    }, 5000) // 每5秒刷新一次

    return () => clearInterval(interval)
  }, [autoRefresh])

  const totalDomains = domains.length
  const totalNodes = domains.reduce((sum, d) => sum + d.nodeCount, 0)
  const totalOnlineNodes = domains.reduce((sum, d) => sum + d.onlineNodes, 0)
  
  // 计算整个集群的资源标签（汇总所有域的资源）
  const clusterResourceTags = useMemo(() => {
    const tags: {
      cpu?: boolean
      gpu?: boolean
      memory?: boolean
      camera?: boolean
    } = {}
    
    domains.forEach(domain => {
      if (domain.resourceTags.cpu !== undefined) tags.cpu = true
      if (domain.resourceTags.gpu !== undefined) tags.gpu = true
      if (domain.resourceTags.memory !== undefined) tags.memory = true
      if (domain.resourceTags.camera) tags.camera = true
    })
    
    return tags
  }, [domains])

  const getStatusBadge = (online: number, total: number) => {
    const ratio = total > 0 ? online / total : 0
    if (ratio >= 0.8) {
      return <Badge variant="default" className="bg-green-500">健康</Badge>
    } else if (ratio >= 0.5) {
      return <Badge variant="default" className="bg-yellow-500">警告</Badge>
    } else {
      return <Badge variant="destructive">异常</Badge>
    }
  }

  const handleCreateDomain = async () => {
    if (!domainName.trim()) {
      alert("请输入域名")
      return
    }

    setCreating(true)
    try {
      await registryAPI.createDomain({
        name: domainName.trim(),
        description: domainDescription.trim() || undefined,
      })

      // 创建成功，关闭对话框并刷新列表
      setIsDialogOpen(false)
      setDomainName("")
      setDomainDescription("")
      await fetchDomains()
    } catch (error) {
      console.error('Failed to create domain:', error)
      if (error instanceof APIError) {
        alert(error.message)
      } else {
        alert(error instanceof Error ? error.message : '创建域失败')
      }
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />

      <main className="flex-1 overflow-auto">
        <div className="p-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-playfair font-bold text-foreground mb-2">资源域管理</h1>
              <p className="text-muted-foreground">查看和管理各个资源域及其 iarnet 节点状态</p>
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
                onClick={fetchDomains}
                disabled={loading}
              >
                <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                刷新数据
              </Button>
              
              <Dialog 
                open={isDialogOpen} 
                onOpenChange={(open) => {
                  setIsDialogOpen(open)
                  if (!open) {
                    // 关闭时重置表单
                    setDomainName("")
                    setDomainDescription("")
                  }
                }}
              >
                <DialogTrigger asChild>
                  <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    添加域
                  </Button>
                </DialogTrigger>
                <DialogContent className="sm:max-w-[425px]">
                  <DialogHeader>
                    <DialogTitle>添加资源域</DialogTitle>
                    <DialogDescription>
                      创建一个新的资源域，用于管理一组 iarnet 节点
                    </DialogDescription>
                  </DialogHeader>
                  <div className="space-y-4 py-4">
                    <div className="space-y-2">
                      <label htmlFor="name" className="text-sm font-medium">
                        域名 <span className="text-red-500">*</span>
                      </label>
                      <Input
                        id="name"
                        placeholder="输入域名"
                        value={domainName}
                        onChange={(e) => setDomainName(e.target.value)}
                        disabled={creating}
                      />
                    </div>
                    <div className="space-y-2">
                      <label htmlFor="description" className="text-sm font-medium">
                        描述
                      </label>
                      <Textarea
                        id="description"
                        placeholder="输入域描述（可选）"
                        value={domainDescription}
                        onChange={(e) => setDomainDescription(e.target.value)}
                        disabled={creating}
                        className="min-h-[80px]"
                      />
                    </div>
                  </div>
                  <DialogFooter>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setIsDialogOpen(false)}
                      disabled={creating}
                    >
                      取消
                    </Button>
                    <Button
                      type="button"
                      onClick={handleCreateDomain}
                      disabled={creating || !domainName.trim()}
                    >
                      {creating ? (
                        <>
                          <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                          创建中...
                        </>
                      ) : (
                        "创建"
                      )}
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>

          {/* Cluster Resource Tags */}
          <Card className="mb-8">
            <CardHeader>
              <CardTitle>集群资源标签</CardTitle>
              <CardDescription>整个集群支持的所有资源类型</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-3">
                {clusterResourceTags.cpu && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <Cpu className="h-4 w-4" />
                    CPU
                  </Badge>
                )}
                {clusterResourceTags.gpu && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <Activity className="h-4 w-4" />
                    GPU
                  </Badge>
                )}
                {clusterResourceTags.memory && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <MemoryStick className="h-4 w-4" />
                    内存
                  </Badge>
                )}
                {clusterResourceTags.camera && (
                  <Badge variant="outline" className="flex items-center gap-2 px-3 py-2 text-sm">
                    <Camera className="h-4 w-4" />
                    摄像头
                  </Badge>
                )}
                {Object.keys(clusterResourceTags).length === 0 && (
                  <span className="text-sm text-muted-foreground">暂无资源标签</span>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Domains Table */}
          <Card>
            <CardHeader>
              <CardTitle>资源域列表</CardTitle>
              <CardDescription>所有已注册的资源域及其资源标签和节点状态</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>域名</TableHead>
                    <TableHead>描述</TableHead>
                    <TableHead>节点状态</TableHead>
                    <TableHead>资源标签</TableHead>
                    <TableHead>最后更新</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {loading ? (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                        加载中...
                      </TableCell>
                    </TableRow>
                  ) : domains.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                        暂无资源域
                      </TableCell>
                    </TableRow>
                  ) : (
                    domains.map((domain) => (
                      <TableRow key={domain.id}>
                        <TableCell>
                          <div className="flex items-center space-x-2">
                            <Globe className="h-4 w-4 text-primary" />
                            <div>
                              <div className="font-medium">{domain.name}</div>
                              <div className="text-xs text-muted-foreground">ID: {domain.id}</div>
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm text-muted-foreground">
                            {domain.description || "无描述"}
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center space-x-2">
                            {getStatusBadge(domain.onlineNodes, domain.nodeCount)}
                            <div className="text-sm text-muted-foreground">
                              {domain.onlineNodes} / {domain.nodeCount}
                            </div>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-2">
                            {domain.resourceTags.cpu !== undefined && (
                              <Badge variant="outline" className="flex items-center gap-1">
                                <Cpu className="h-3 w-3" />
                                CPU
                              </Badge>
                            )}
                            {domain.resourceTags.gpu !== undefined && (
                              <Badge variant="outline" className="flex items-center gap-1">
                                <Activity className="h-3 w-3" />
                                GPU
                              </Badge>
                            )}
                            {domain.resourceTags.memory !== undefined && (
                              <Badge variant="outline" className="flex items-center gap-1">
                                <MemoryStick className="h-3 w-3" />
                                内存
                              </Badge>
                            )}
                            {domain.resourceTags.camera && (
                              <Badge variant="outline" className="flex items-center gap-1">
                                <Camera className="h-3 w-3" />
                                摄像头
                              </Badge>
                            )}
                          </div>
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {formatRelativeTime(domain.lastUpdated)}
                        </TableCell>
                        <TableCell>
                          <Button variant="outline" size="sm" asChild>
                            <Link href={`/domains/${domain.id}`}>
                              查看详情
                            </Link>
                          </Button>
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

