package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) setupRoutes() {
	router := gin.New()

	router.Use(globalServerMiddleware()...)

	router.POST("/enroll", s.handleCreateEnrollments())

	s.server.Handler = router
}
