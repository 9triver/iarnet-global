package logs

import (
	"net/http"
	"strconv"

	"github.com/9triver/iarnet-global/internal/transport/http/util/response"
	"github.com/9triver/iarnet-global/internal/util"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RegisterRoutes 注册日志相关的 HTTP 路由
func RegisterRoutes(router *mux.Router) {
	api := NewAPI()
	router.HandleFunc("/logs", api.handleGetLogs).Methods("GET")
	router.HandleFunc("/logs/clear", api.handleClearLogs).Methods("POST")
}

type API struct {
	logHook *util.MemoryLogHook
}

func NewAPI() *API {
	return &API{
		logHook: util.GlobalLogHook,
	}
}

// handleGetLogs 获取日志
// 查询参数:
//   - start: 起始索引（默认 0）
//   - limit: 返回的最大数量（默认 100，最大 1000）
//   - level: 过滤的日志级别（可选：trace, debug, info, warn, error, fatal, panic）
func (api *API) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	if api.logHook == nil {
		logrus.Error("Log hook is not initialized")
		response.InternalError("log hook is not initialized").WriteJSON(w)
		return
	}

	// 解析查询参数
	start := 0
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if parsed, err := strconv.Atoi(startStr); err == nil && parsed >= 0 {
			start = parsed
		}
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			// 限制最大返回数量
			if limit > 1000 {
				limit = 1000
			}
		}
	}

	level := r.URL.Query().Get("level")

	// 获取日志
	logs := api.logHook.GetLogs(start, limit, level)
	total := api.logHook.GetTotalCount()

	resp := GetLogsResponse{
		Logs:  logs,
		Total: total,
		Start: start,
		Limit: limit,
	}

	response.Success(resp).WriteJSON(w)
}

// handleClearLogs 清空所有日志
func (api *API) handleClearLogs(w http.ResponseWriter, r *http.Request) {
	if api.logHook == nil {
		logrus.Error("Log hook is not initialized")
		response.InternalError("log hook is not initialized").WriteJSON(w)
		return
	}

	api.logHook.Clear()
	response.Success(nil).WriteJSON(w)
}

// GetLogsResponse 获取日志响应
type GetLogsResponse struct {
	Logs  []util.LogEntry `json:"logs"`
	Total int             `json:"total"`
	Start int             `json:"start"`
	Limit int             `json:"limit"`
}

