package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gabrielmelo/tg-forward/internal/api/router"
	"github.com/gabrielmelo/tg-forward/internal/matcher"
	"github.com/gabrielmelo/tg-forward/internal/service"
)

type Server struct {
	service  *service.RulesService
	port     string
	apiToken string
	server   *http.Server
}

func NewServer(svc *service.RulesService, port, apiToken string) *Server {
	return &Server{
		service:  svc,
		port:     port,
		apiToken: apiToken,
	}
}

func (s *Server) Start() error {
	r := router.New(s.service, s.apiToken)

	addr := fmt.Sprintf(":%s", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("Starting API server on %s", addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	log.Println("Shutting down API server...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.server.Shutdown(shutdownCtx)
}

func (s *Server) GetMatcher() *matcher.Matcher {
	return s.service.GetMatcher()
}
