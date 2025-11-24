"use client"

import { useState, useEffect, useRef, useMemo } from "react"
import { Sidebar } from "@/components/sidebar"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { FileText, RefreshCw, Search, Download, Trash2, Filter, X } from "lucide-react"
import { formatDateTime } from "@/lib/utils"
import { AutoSizer, CellMeasurer, CellMeasurerCache, List, type ListRowProps } from "react-virtualized"
import { logsAPI, type LogEntry as APILogEntry } from "@/lib/api"

// 日志级别样式定义 - 与 iarnet 应用日志保持一致
const LOG_LEVEL_STYLES: Record<string, { badge: string; dot: string; label: string }> = {
  error: { badge: "bg-red-100 text-red-800", dot: "bg-red-500", label: "错误" },
  warn: { badge: "bg-amber-100 text-amber-800", dot: "bg-amber-500", label: "警告" },
  debug: { badge: "bg-blue-100 text-blue-800", dot: "bg-blue-500", label: "调试" },
  trace: { badge: "bg-slate-100 text-slate-800", dot: "bg-slate-400", label: "追踪" },
  info: { badge: "bg-emerald-100 text-emerald-800", dot: "bg-emerald-500", label: "信息" },
}

// 日志条目接口（前端使用）
interface LogEntry {
  id: string
  timestamp: string
  level: "debug" | "info" | "warn" | "error" | "trace" | "fatal" | "panic"
  message: string
  source?: string
  domainId?: string
  nodeId?: string
  details?: string
  caller?: LogCallerInfo
}

// 调用者信息类型
interface LogCallerInfo {
  file?: string
  line?: number
  function?: string
}

// 基础日志条目类型 - 用于 LogListViewer
type BasicLogEntry = {
  id: string
  timestamp?: string
  level?: string
  message: string
  details?: string
  caller?: LogCallerInfo
}

// LogListViewer 组件 - 与 iarnet 应用日志完全一致
const LogListViewer = ({ logs }: { logs: BasicLogEntry[] }) => {
  const cacheRef = useRef(
    new CellMeasurerCache({
      fixedWidth: true,
      defaultHeight: 72,
    })
  )

  useEffect(() => {
    cacheRef.current.clearAll()
  }, [logs])

  if (logs.length === 0) {
    return null
  }

  return (
    <AutoSizer>
      {({ height, width }: { height: number; width: number }) => (
        <List
          width={width}
          height={height}
          rowCount={logs.length}
          deferredMeasurementCache={cacheRef.current}
          rowHeight={cacheRef.current.rowHeight}
          overscanRowCount={6}
          rowRenderer={({ index, key, parent, style }: ListRowProps) => {
            const log = logs[index]
            const levelKey = (log.level || "info").toLowerCase()
            const levelStyles = LOG_LEVEL_STYLES[levelKey] || LOG_LEVEL_STYLES.info

            return (
              <CellMeasurer
                cache={cacheRef.current}
                columnIndex={0}
                key={key}
                parent={parent}
                rowIndex={index}
              >
                <div
                  style={style}
                  className="border-b border-gray-200/80 dark:border-gray-800/80 px-4 py-3 hover:bg-white dark:hover:bg-gray-900 transition-colors"
                >
                  <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                    <div className="flex items-center gap-3">
                      <span
                        className={`px-2 py-0.5 rounded-full text-[11px] font-semibold uppercase tracking-wide ${levelStyles.badge}`}
                      >
                        {(log.level ?? "INFO").toUpperCase()}
                      </span>
                      <span className="text-xs text-muted-foreground font-mono">
                        {log.timestamp ? formatDateTime(log.timestamp) : "—"}
                      </span>
                    </div>
                    <div className="text-[11px] text-muted-foreground font-mono flex items-center gap-2">
                      {log.caller && log.caller.function && (
                        <span className="hidden md:inline">
                          {log.caller.function}
                        </span>
                      )}
                    </div>
                  </div>
                  <p className="mt-2 text-sm text-gray-900 dark:text-gray-100 whitespace-pre-wrap break-words font-mono">
                    {log.message}
                  </p>
                  {log.details && (
                    <pre className="mt-2 bg-gray-100 dark:bg-gray-950 rounded-md p-2 text-xs text-gray-700 dark:text-gray-300 overflow-x-auto whitespace-pre-wrap break-words font-mono">
                      {log.details}
                    </pre>
                  )}
                </div>
              </CellMeasurer>
            )
          }}
        />
      )}
    </AutoSizer>
  )
}

export default function LogsPage() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [logFilter, setLogFilter] = useState<'all' | 'debug' | 'info' | 'warn' | 'error'>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [logLines, setLogLines] = useState(100)

  // 从后端 API 获取日志
  const fetchLogs = async () => {
    try {
      setLoading(true)
      const level = logFilter !== 'all' ? logFilter : undefined
      const response = await logsAPI.getLogs({
        start: 0,
        limit: logLines,
        level,
      })

      // 转换后端日志格式为前端格式
      const convertedLogs: LogEntry[] = response.logs.map((log: APILogEntry, index: number) => {
        // 构建 details 字符串（只包含 fields 信息，caller 单独处理）
        const detailsParts: string[] = []
        if (log.fields && Object.keys(log.fields).length > 0) {
          const fieldsStr = Object.entries(log.fields)
            .map(([key, value]) => `${key}=${JSON.stringify(value)}`)
            .join(", ")
          detailsParts.push(`Fields: ${fieldsStr}`)
        }
        const details = detailsParts.length > 0 ? detailsParts.join("\n") : undefined

        // 提取 caller 信息
        const caller = log.caller ? {
          file: log.caller.file,
          line: log.caller.line,
          function: log.caller.function,
        } : undefined

        return {
          id: `${log.timestamp}-${index}`,
          timestamp: log.timestamp,
          level: log.level.toLowerCase() as LogEntry["level"],
          message: log.message,
          details,
          caller,
        }
      })

      setLogs(convertedLogs)
    } catch (error) {
      console.error('Failed to fetch logs:', error)
      // 发生错误时保持现有日志，不清空
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchLogs()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [logFilter, logLines])

  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      fetchLogs()
    }, 2000) // 每2秒刷新一次日志

    return () => clearInterval(interval)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoRefresh, logFilter, logLines])

  // 过滤日志（后端已按级别过滤，这里只做搜索过滤）
  const filteredLogs = useMemo(() => {
    const searchTerm = searchQuery.trim().toLowerCase()
    if (!searchTerm) {
      return logs
    }

    return logs.filter((log) => {
      const content = `${log.message ?? ""} ${log.details ?? ""}`.toLowerCase()
      return content.includes(searchTerm)
    })
  }, [logs, searchQuery])

  // 转换为 BasicLogEntry 格式
  const logEntries: BasicLogEntry[] = useMemo(() => {
    return filteredLogs.map((log) => ({
      id: log.id,
      timestamp: log.timestamp,
      level: log.level,
      message: log.message,
      details: log.details,
      caller: log.caller,
    }))
  }, [filteredLogs])

  const handleClearLogs = async () => {
    if (confirm('确定要清空所有日志吗？')) {
      try {
        await logsAPI.clearLogs()
        setLogs([])
      } catch (error) {
        console.error('Failed to clear logs:', error)
        alert('清空日志失败，请稍后重试')
      }
    }
  }

  const handleExportLogs = () => {
    const logText = filteredLogs.map(log => 
      `[${formatDateTime(log.timestamp)}] [${log.level.toUpperCase()}] ${log.message}`
    ).join('\n')
    
    const blob = new Blob([logText], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `iarnet-global-logs-${new Date().toISOString()}.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />

      <main className="flex-1 overflow-auto">
        <div className="p-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-3xl font-playfair font-bold text-foreground mb-2">调度器日志</h1>
              <p className="text-muted-foreground">查看全局调度器的运行日志</p>
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
                size="sm"
                onClick={handleExportLogs}
                disabled={filteredLogs.length === 0}
              >
                <Download className="h-4 w-4 mr-2" />
                导出日志
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={handleClearLogs}
                disabled={logs.length === 0}
              >
                <Trash2 className="h-4 w-4 mr-2" />
                清空日志
              </Button>
            </div>
          </div>

          {/* Logs Card */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>调度器日志</CardTitle>
                  <CardDescription>
                    查看全局调度器的实时日志输出
                  </CardDescription>
                </div>
                <div className="flex items-center space-x-2">
                  <Select value={logLines.toString()} onValueChange={(value: string) => setLogLines(Number(value))}>
                    <SelectTrigger className="w-32">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="50">最近50条</SelectItem>
                      <SelectItem value="100">最近100条</SelectItem>
                      <SelectItem value="200">最近200条</SelectItem>
                      <SelectItem value="500">最近500条</SelectItem>
                    </SelectContent>
                  </Select>
                  <Button variant="outline" size="sm" onClick={fetchLogs} disabled={loading}>
                    <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                    刷新
                  </Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="mb-4 flex items-center space-x-2">
                <div className="flex items-center space-x-2 flex-1">
                  <Search className="h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="搜索日志内容..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="max-w-md"
                  />
                  {searchQuery && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setSearchQuery("")}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  )}
                </div>
                <Select value={logFilter} onValueChange={(value: any) => setLogFilter(value)}>
                  <SelectTrigger className="w-32">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部级别</SelectItem>
                    <SelectItem value="error">错误</SelectItem>
                    <SelectItem value="warn">警告</SelectItem>
                    <SelectItem value="info">信息</SelectItem>
                    <SelectItem value="debug">调试</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              {loading && logs.length === 0 ? (
                <div className="flex items-center justify-center h-32">
                  <RefreshCw className="h-6 w-6 animate-spin mr-2" />
                  <span>加载日志中...</span>
                </div>
              ) : filteredLogs.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-40 text-muted-foreground space-y-2 text-sm">
                  {searchQuery || logFilter !== "all" ? (
                    <Filter className="h-8 w-8 opacity-50" />
                  ) : (
                    <FileText className="h-8 w-8 opacity-50" />
                  )}
                  <span>{searchQuery || logFilter !== "all" ? "没有符合条件的日志" : "尚未获取到日志数据"}</span>
                  <span className="text-xs">
                    {searchQuery || logFilter !== "all" ? "调整筛选条件或清空搜索后重试" : "启动应用或刷新后再试"}
                  </span>
                </div>
              ) : (
                <div className="h-[500px] w-full border rounded-md bg-gray-50 dark:bg-gray-900">
                  <LogListViewer logs={logEntries} />
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </main>
    </div>
  )
}
