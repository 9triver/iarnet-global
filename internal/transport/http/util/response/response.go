package response

import (
	"encoding/json"
	"net/http"
)

// BaseResponse 统一的API响应结构
type BaseResponse struct {
	Code    int    `json:"code"`            // HTTP状态码
	Message string `json:"message"`         // 响应消息
	Data    any    `json:"data,omitempty"`  // 响应数据
	Error   string `json:"error,omitempty"` // 错误信息
}

// Success 创建成功响应
func Success(data any) *BaseResponse {
	return &BaseResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
}

// Created 创建资源成功响应
func Created(data any) *BaseResponse {
	return &BaseResponse{
		Code:    http.StatusCreated,
		Message: "created",
		Data:    data,
	}
}

// BadRequest 创建错误请求响应
func BadRequest(error string) *BaseResponse {
	return &BaseResponse{
		Code:    http.StatusBadRequest,
		Message: "bad request",
		Error:   error,
	}
}

// InternalError 创建内部服务器错误响应
func InternalError(error string) *BaseResponse {
	return &BaseResponse{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
		Error:   error,
	}
}

// NotFound 创建资源未找到响应
func NotFound(error string) *BaseResponse {
	return &BaseResponse{
		Code:    http.StatusNotFound,
		Message: "not found",
		Error:   error,
	}
}

// WriteJSON 将响应写入HTTP响应
func (r *BaseResponse) WriteJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	return json.NewEncoder(w).Encode(r)
}
