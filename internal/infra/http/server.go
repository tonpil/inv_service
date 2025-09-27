package http

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	host   string
	port   int32
}

func NewServer(router *gin.Engine, host string, port int32) *Server {
	return &Server{
		router: router,
		host:   host,
		port:   port,
	}
}

func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("Starting HTTP server at %s", addr)

	if err := s.router.Run(addr); err != nil {
		log.Fatalf("failed to run http server: %v", err)
	}
}
