package scheduler

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/9triver/iarnet-global/internal/domain/registry"
	resourcepb "github.com/9triver/iarnet-global/internal/proto/resource"
	schedulerpb "github.com/9triver/iarnet-global/internal/proto/scheduler"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Service 定义全局调度能力
type Service interface {
	DeployComponent(ctx context.Context, req *schedulerpb.DeployComponentRequest) (*schedulerpb.DeployComponentResponse, error)
}

type service struct {
	manager     *registry.Manager
	dialTimeout time.Duration
	rand        *rand.Rand
}

// NewService 创建调度服务
func NewService(manager *registry.Manager) Service {
	return &service{
		manager:     manager,
		dialTimeout: 10 * time.Second,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// DeployComponent 处理调度请求
func (s *service) DeployComponent(ctx context.Context, req *schedulerpb.DeployComponentRequest) (*schedulerpb.DeployComponentResponse, error) {
	if req == nil {
		return failureResponse("request is required"), nil
	}
	if req.ResourceRequest == nil {
		return failureResponse("resource_request is required"), nil
	}

	targetNode, err := s.selectRandomNode(req.ResourceRequest)
	if err != nil {
		logrus.Warnf("Failed to select node for scheduling: %v", err)
		return failureResponse(err.Error()), nil
	}

	resp, err := s.forwardToNode(ctx, targetNode, req)
	if err != nil {
		logrus.Errorf("Failed to forward scheduling request to node %s (%s, domain=%s): %v",
			targetNode.Name, targetNode.Address, targetNode.DomainID, err)
		return failureResponse(fmt.Sprintf("node dispatch failed: %v", err)), nil
	}

	logrus.Infof("Delegated scheduling request to node %s (%s, domain=%s)", targetNode.Name, targetNode.Address, targetNode.DomainID)
	return resp, nil
}

func (s *service) selectRandomNode(resourceReq *resourcepb.Info) (*registry.Node, error) {
	type domainNodes struct {
		domainID registry.DomainID
		nodes    []*registry.Node
	}

	domains := s.manager.GetAllDomains()
	candidates := make([]domainNodes, 0, len(domains))

	for _, domain := range domains {
		nodes, err := s.manager.GetNodesByDomain(domain.ID)
		if err != nil {
			continue
		}

		eligible := make([]*registry.Node, 0, len(nodes))
		for _, node := range nodes {
			if node.Status != registry.NodeStatusOnline {
				continue
			}
			if node.Address == "" {
				continue
			}
			if len(resourceReq.Tags) > 0 && !nodeHasRequiredTags(node.ResourceTags, resourceReq.Tags) {
				continue
			}
			if !hasSufficientResources(node.ResourceCapacity, resourceReq) {
				continue
			}
			eligible = append(eligible, node.Clone())
		}

		if len(eligible) > 0 {
			candidates = append(candidates, domainNodes{
				domainID: domain.ID,
				nodes:    eligible,
			})
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no domain has nodes with sufficient capacity")
	}

	selectedDomain := candidates[s.rand.Intn(len(candidates))]
	selectedNode := selectedDomain.nodes[s.rand.Intn(len(selectedDomain.nodes))]
	return selectedNode, nil
}

func hasSufficientResources(capacity *registry.ResourceCapacity, req *resourcepb.Info) bool {
	if capacity == nil || capacity.Available == nil || req == nil {
		return false
	}

	available := capacity.Available
	if available.CPU < req.Cpu {
		return false
	}
	if available.Memory < req.Memory {
		return false
	}
	if available.GPU < req.Gpu {
		return false
	}
	return true
}

func nodeHasRequiredTags(nodeTags *registry.ResourceTags, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if nodeTags == nil {
		return false
	}

	for _, tag := range required {
		switch strings.ToLower(tag) {
		case "cpu":
			if !nodeTags.CPU {
				return false
			}
		case "gpu":
			if !nodeTags.GPU {
				return false
			}
		case "memory":
			if !nodeTags.Memory {
				return false
			}
		case "camera":
			if !nodeTags.Camera {
				return false
			}
		default:
			// 未知标签暂视为不满足
			return false
		}
	}
	return true
}

func (s *service) forwardToNode(ctx context.Context, node *registry.Node, req *schedulerpb.DeployComponentRequest) (*schedulerpb.DeployComponentResponse, error) {
	dialCtx, cancel := context.WithTimeout(ctx, s.dialTimeout)
	defer cancel()

	conn, err := grpc.NewClient(node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial head node %s: %w", node.Address, err)
	}
	defer conn.Close()

	client := schedulerpb.NewSchedulerServiceClient(conn)
	return client.DeployComponent(dialCtx, req)
}

func failureResponse(msg string) *schedulerpb.DeployComponentResponse {
	return &schedulerpb.DeployComponentResponse{
		Success: false,
		Error:   msg,
	}
}
