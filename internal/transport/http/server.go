package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/9triver/iarnet-global/internal/config"
	"github.com/9triver/iarnet-global/internal/domain/registry"
	logsAPI "github.com/9triver/iarnet-global/internal/transport/http/logs"
	registryAPI "github.com/9triver/iarnet-global/internal/transport/http/registry"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Options struct {
	Port            int
	Config          *config.Config
	RegistryService registry.Service
}

type Server struct {
	Server *http.Server
	Router *mux.Router
}

func NewServer(opts Options) *Server {
	router := mux.NewRouter()
	registryAPI.RegisterRoutes(router, opts.RegistryService)
	logsAPI.RegisterRoutes(router)

	return &Server{
		Server: &http.Server{
			Addr:    fmt.Sprintf("0.0.0.0:%d", opts.Port),
			Handler: router,
		},
		Router: router,
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	logrus.Infof("HTTP server started on %s", s.Server.Addr)
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Failed to stop HTTP server")
	}
	logrus.Info("HTTP server stopped")
}
