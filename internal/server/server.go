package server

import (
	"github.com/gin-gonic/gin"

	"seasons/backend/gen/dbstore"
	"seasons/backend/internal/server/handlers"
	"seasons/backend/internal/server/middleware"
)

type Dependencies struct {
	Queries dbstore.Querier
	Health  *handlers.HealthHandler
	Auth    *handlers.AuthHandler
}

type Server struct {
	router *gin.Engine
	deps   Dependencies
}

func New(deps Dependencies) *Server {
	s := &Server{
		router: gin.Default(),
		deps:   deps,
	}
	s.registerRoutes()
	return s
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) registerRoutes() {
	authMW := middleware.RequireAuth(s.deps.Queries)

	s.router.GET("/health", s.deps.Health.Handle)

	api := s.router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.POST("/login", s.deps.Auth.Login)
		auth.POST("/logout", authMW, s.deps.Auth.Logout)
		auth.GET("/me", authMW, s.deps.Auth.Me)
	}
}
