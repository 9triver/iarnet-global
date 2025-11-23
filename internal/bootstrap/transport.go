package bootstrap

import (
	"github.com/9triver/iarnet-global/internal/transport/http"
	"github.com/sirupsen/logrus"
)

// bootstrapTransport 初始化 Transport 层（HTTP）
func bootstrapTransport(ig *IarnetGlobal) error {
	// 创建 HTTP 服务器
	ig.HTTPServer = http.NewServer(http.Options{
		Port:            ig.Config.Transport.HTTP.Port,
		Config:          ig.Config,
		RegistryService: ig.RegistryService,
	})

	logrus.Info("Transport layer initialized")
	return nil
}
