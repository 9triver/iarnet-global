package util

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LogEntry 表示一条日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    *CallerInfo            `json:"caller,omitempty"`
}

// CallerInfo 表示调用者信息
type CallerInfo struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

// MemoryLogHook 是一个内存日志收集器
type MemoryLogHook struct {
	mu      sync.RWMutex
	logs    []LogEntry
	maxSize int
}

// NewMemoryLogHook 创建一个新的内存日志收集器
func NewMemoryLogHook(maxSize int) *MemoryLogHook {
	if maxSize <= 0 {
		maxSize = 1000 // 默认保存 1000 条日志
	}
	return &MemoryLogHook{
		logs:    make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Levels 返回 hook 要处理的日志级别
func (h *MemoryLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 处理日志条目
func (h *MemoryLogHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	logEntry := LogEntry{
		Timestamp: entry.Time,
		Level:     entry.Level.String(),
		Message:   entry.Message,
	}

	// 复制字段
	if len(entry.Data) > 0 {
		logEntry.Fields = make(map[string]interface{}, len(entry.Data))
		for k, v := range entry.Data {
			logEntry.Fields[k] = v
		}
	}

	// 复制调用者信息
	if entry.HasCaller() && entry.Caller != nil {
		logEntry.Caller = &CallerInfo{
			File:     entry.Caller.File,
			Line:     entry.Caller.Line,
			Function: entry.Caller.Function,
		}
	}

	// 添加到日志列表
	h.logs = append(h.logs, logEntry)

	// 如果超过最大大小，删除最旧的日志
	if len(h.logs) > h.maxSize {
		h.logs = h.logs[1:]
	}

	return nil
}

// GetLogs 获取日志条目
// start: 起始索引（从 0 开始，0 表示最新的日志）
// limit: 返回的最大数量
// level: 过滤的日志级别（空字符串表示不过滤）
// 返回的日志按时间倒序排列（最新的在前）
func (h *MemoryLogHook) GetLogs(start, limit int, level string) []LogEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if limit <= 0 {
		limit = 100
	}

	total := len(h.logs)
	if total == 0 {
		return []LogEntry{}
	}

	// 日志按时间顺序存储：logs[0] 是最旧的，logs[total-1] 是最新的
	// 用户请求 start=0, limit=10 时，应该返回最新的 10 条
	// 即从 logs[total-1] 往前取 limit 条

	// 计算从末尾开始的索引
	fromEnd := start + limit
	if fromEnd > total {
		fromEnd = total
	}

	// 计算实际索引范围（从后往前）
	endIdx := total - start
	startIdx := total - fromEnd
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > total {
		endIdx = total
	}

	// 提取日志（从旧到新）
	logs := make([]LogEntry, endIdx-startIdx)
	copy(logs, h.logs[startIdx:endIdx])

	// 反转顺序，使最新的在前
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	// 如果指定了级别过滤
	if level != "" {
		filtered := make([]LogEntry, 0, len(logs))
		for _, log := range logs {
			if log.Level == level {
				filtered = append(filtered, log)
			}
		}
		return filtered
	}

	return logs
}

// GetTotalCount 获取日志总数
func (h *MemoryLogHook) GetTotalCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.logs)
}

// Clear 清空所有日志
func (h *MemoryLogHook) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logs = h.logs[:0]
}

