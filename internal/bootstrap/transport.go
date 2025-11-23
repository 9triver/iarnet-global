package bootstrap

import (
	"fmt"

	"github.com/9triver/iarnet-global/internal/transport/http"
	"github.com/9triver/iarnet-global/internal/transport/rpc"
	"github.com/sirupsen/logrus"
)

// bootstrapTransport 初始化 Transport 层（HTTP、RPC）
func bootstrapTransport(ig *IarnetGlobal) error {
	// 创建 HTTP 服务器
	ig.HTTPServer = http.NewServer(http.Options{
		Port:            ig.Config.Transport.HTTP.Port,
		Config:          ig.Config,
		RegistryService: ig.RegistryService,
	})

	// 构建 RPC 服务器地址
	registryAddr := fmt.Sprintf("0.0.0.0:%d", ig.Config.Transport.RPC.Registry.Port)

	// 创建 RPC 服务器管理器
	ig.RPCManager = rpc.NewManager(rpc.Options{
		RegistryAddr:    registryAddr,
		RegistryService: ig.DomainManager,
	})

	logrus.Info("Transport layer initialized")
	return nil
}
