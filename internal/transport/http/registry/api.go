package registry

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/9triver/iarnet-global/internal/domain/registry"
	"github.com/9triver/iarnet-global/internal/transport/http/util/response"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RegisterRoutes 注册域相关的 HTTP 路由
func RegisterRoutes(router *mux.Router, service registry.Service) {
	api := NewAPI(service)
	router.HandleFunc("/registry/domains", api.handleGetDomains).Methods("GET")
	router.HandleFunc("/registry/domains", api.handleCreateDomain).Methods("POST")
	router.HandleFunc("/registry/domains/{id}", api.handleGetDomain).Methods("GET")
	router.HandleFunc("/registry/domains/{id}", api.handleUpdateDomain).Methods("PUT")
	router.HandleFunc("/registry/domains/{id}", api.handleDeleteDomain).Methods("DELETE")
	router.HandleFunc("/registry/domains/{id}/nodes", api.handleGetDomainNodes).Methods("GET")
}

type API struct {
	service registry.Service
}

func NewAPI(service registry.Service) *API {
	return &API{
		service: service,
	}
}

// handleGetDomains 获取所有域
func (api *API) handleGetDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := api.service.GetAllDomains(r.Context())
	if err != nil {
		logrus.Errorf("Failed to get domains: %v", err)
		response.InternalError("failed to get domains: " + err.Error()).WriteJSON(w)
		return
	}

	resp := GetDomainsResponse{
		Domains: make([]DomainItem, 0, len(domains)),
		Total:   len(domains),
	}

	for _, domain := range domains {
		// 获取域统计信息
		stats, err := api.service.GetDomainStats(r.Context(), domain.ID)
		if err != nil {
			logrus.Warnf("Failed to get domain stats for %s: %v", domain.ID, err)
			stats = &registry.DomainStats{
				TotalNodes:  len(domain.NodeIDs),
				OnlineNodes: 0,
			}
		}

		item := DomainItem{
			ID:          domain.ID,
			Name:        domain.Name,
			Description: domain.Description,
			NodeCount:   stats.TotalNodes,
			OnlineNodes: stats.OnlineNodes,
			ResourceTags: ResourceTagsResponse{
				CPU:    domain.ResourceTags.CPU,
				GPU:    domain.ResourceTags.GPU,
				Memory: domain.ResourceTags.Memory,
				Camera: domain.ResourceTags.Camera,
			},
			CreatedAt: domain.CreatedAt.Format(time.RFC3339),
			UpdatedAt: domain.UpdatedAt.Format(time.RFC3339),
		}

		resp.Domains = append(resp.Domains, item)
	}

	response.Success(resp).WriteJSON(w)
}

// handleCreateDomain 创建域
func (api *API) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	req := CreateDomainRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("Failed to decode create domain request: %v", err)
		response.BadRequest("invalid request body: " + err.Error()).WriteJSON(w)
		return
	}

	// 验证必填字段
	if req.Name == "" {
		response.BadRequest("domain name is required").WriteJSON(w)
		return
	}

	logrus.Infof("Creating domain: name=%s, description=%s", req.Name, req.Description)

	// 调用 service 创建域
	domain, err := api.service.CreateDomain(r.Context(), req.Name, req.Description)
	if err != nil {
		logrus.Errorf("Failed to create domain: %v", err)
		response.InternalError("failed to create domain: " + err.Error()).WriteJSON(w)
		return
	}

	logrus.Infof("Domain created successfully: id=%s, name=%s", domain.ID, domain.Name)

	resp := CreateDomainResponse{
		ID:          string(domain.ID),
		Name:        domain.Name,
		Description: domain.Description,
		CreatedAt:   domain.CreatedAt.Format(time.RFC3339),
	}

	response.Created(resp).WriteJSON(w)
}

// handleGetDomain 获取单个域
func (api *API) handleGetDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID := registry.DomainID(vars["id"])
	if domainID == "" {
		response.BadRequest("domain id is required").WriteJSON(w)
		return
	}

	domain, err := api.service.GetDomain(r.Context(), domainID)
	if err != nil {
		if err == registry.ErrDomainNotFound {
			response.NotFound("domain not found").WriteJSON(w)
			return
		}
		logrus.Errorf("Failed to get domain: %v", err)
		response.InternalError("failed to get domain: " + err.Error()).WriteJSON(w)
		return
	}

	// 获取域下的节点
	nodes, err := api.service.GetDomainNodes(r.Context(), domainID)
	if err != nil {
		logrus.Warnf("Failed to get domain nodes: %v", err)
		nodes = []*registry.Node{} // 使用空列表
	}

	resp := GetDomainResponse{
		ID:          domain.ID,
		Name:        domain.Name,
		Description: domain.Description,
		ResourceTags: ResourceTagsResponse{
			CPU:    domain.ResourceTags != nil && domain.ResourceTags.CPU,
			GPU:    domain.ResourceTags != nil && domain.ResourceTags.GPU,
			Memory: domain.ResourceTags != nil && domain.ResourceTags.Memory,
			Camera: domain.ResourceTags != nil && domain.ResourceTags.Camera,
		},
		Nodes:     convertNodes(nodes),
		CreatedAt: domain.CreatedAt.Format(time.RFC3339),
		UpdatedAt: domain.UpdatedAt.Format(time.RFC3339),
	}

	response.Success(resp).WriteJSON(w)
}

// handleUpdateDomain 更新域
func (api *API) handleUpdateDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID := registry.DomainID(vars["id"])
	if domainID == "" {
		response.BadRequest("domain id is required").WriteJSON(w)
		return
	}

	req := UpdateDomainRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("Failed to decode update domain request: %v", err)
		response.BadRequest("invalid request body: " + err.Error()).WriteJSON(w)
		return
	}

	err := api.service.UpdateDomain(r.Context(), domainID, req.Name, req.Description)
	if err != nil {
		if err == registry.ErrDomainNotFound {
			response.NotFound("domain not found").WriteJSON(w)
			return
		}
		logrus.Errorf("Failed to update domain: %v", err)
		response.InternalError("failed to update domain: " + err.Error()).WriteJSON(w)
		return
	}

	response.Success(nil).WriteJSON(w)
}

// handleDeleteDomain 删除域
func (api *API) handleDeleteDomain(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID := registry.DomainID(vars["id"])
	if domainID == "" {
		response.BadRequest("domain id is required").WriteJSON(w)
		return
	}

	err := api.service.DeleteDomain(r.Context(), domainID)
	if err != nil {
		if err == registry.ErrDomainNotFound {
			response.NotFound("domain not found").WriteJSON(w)
			return
		}
		logrus.Errorf("Failed to delete domain: %v", err)
		response.InternalError("failed to delete domain: " + err.Error()).WriteJSON(w)
		return
	}

	response.Success(nil).WriteJSON(w)
}

// handleGetDomainNodes 获取域下的所有节点
func (api *API) handleGetDomainNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	domainID := registry.DomainID(vars["id"])
	if domainID == "" {
		response.BadRequest("domain id is required").WriteJSON(w)
		return
	}

	nodes, err := api.service.GetDomainNodes(r.Context(), domainID)
	if err != nil {
		if err == registry.ErrDomainNotFound {
			response.NotFound("domain not found").WriteJSON(w)
			return
		}
		logrus.Errorf("Failed to get domain nodes: %v", err)
		response.InternalError("failed to get domain nodes: " + err.Error()).WriteJSON(w)
		return
	}

	resp := GetDomainNodesResponse{
		Nodes: convertNodes(nodes),
		Total: len(nodes),
	}

	response.Success(resp).WriteJSON(w)
}

// convertNodes 转换节点列表
func convertNodes(nodes []*registry.Node) []NodeItem {
	items := make([]NodeItem, 0, len(nodes))
	for _, node := range nodes {
		item := NodeItem{
			ID:       node.ID,
			Name:     node.Name,
			Address:  node.Address,
			Status:   string(node.Status),
			IsHead:   node.IsHead,
			LastSeen: node.LastSeen.Format(time.RFC3339),
		}

		// 转换资源标签和资源容量
		// 优先使用 ResourceCapacity.Total 中的数值，如果没有则使用 ResourceTags 的 bool 值
		resourceTags := &NodeResourceTagsResponse{}
		hasResourceTags := false

		// 从 ResourceCapacity 获取资源容量数值
		if node.ResourceCapacity != nil && node.ResourceCapacity.Total != nil {
			total := node.ResourceCapacity.Total
			if total.CPU > 0 {
				// CPU 从 millicores 转换为 cores（除以 1000）
				cpuCores := total.CPU / 1000
				resourceTags.CPU = &cpuCores
				hasResourceTags = true
			}
			if total.GPU > 0 {
				resourceTags.GPU = &total.GPU
				hasResourceTags = true
			}
			if total.Memory > 0 {
				resourceTags.Memory = &total.Memory
				hasResourceTags = true
			}
		}

		// 从 ResourceTags 获取资源标签（bool 值），主要用于 Camera
		if node.ResourceTags != nil {
			if node.ResourceTags.Camera {
				camera := true
				resourceTags.Camera = &camera
				hasResourceTags = true
			}
			// 如果 ResourceCapacity 中没有数值，但 ResourceTags 中有标记，则使用标记
			if resourceTags.CPU == nil && node.ResourceTags.CPU {
				// 如果没有容量信息，但标记支持 CPU，则不设置具体数值（前端会显示标签但不显示数值）
			}
			if resourceTags.GPU == nil && node.ResourceTags.GPU {
				// 如果没有容量信息，但标记支持 GPU，则不设置具体数值
			}
			if resourceTags.Memory == nil && node.ResourceTags.Memory {
				// 如果没有容量信息，但标记支持 Memory，则不设置具体数值
			}
		}

		if hasResourceTags {
			item.ResourceTags = resourceTags
		}

		items = append(items, item)
	}
	return items
}
