package scheduler

import (
	"context"

	domainscheduler "github.com/9triver/iarnet-global/internal/domain/scheduler"
	schedulerpb "github.com/9triver/iarnet-global/internal/proto/scheduler"
)

// Server 全局调度器 RPC 实现
type Server struct {
	schedulerpb.UnimplementedSchedulerServiceServer
	service domainscheduler.Service
}

// NewServer 创建调度 RPC 服务器
func NewServer(service domainscheduler.Service) *Server {
	return &Server{
		service: service,
	}
}

// DeployComponent 处理调度请求
func (s *Server) DeployComponent(ctx context.Context, req *schedulerpb.DeployComponentRequest) (*schedulerpb.DeployComponentResponse, error) {
	return s.service.DeployComponent(ctx, req)
}

// GetDeploymentStatus 暂未实现
func (s *Server) GetDeploymentStatus(ctx context.Context, req *schedulerpb.GetDeploymentStatusRequest) (*schedulerpb.GetDeploymentStatusResponse, error) {
	return &schedulerpb.GetDeploymentStatusResponse{
		Success: false,
		Error:   "not implemented",
	}, nil
}
