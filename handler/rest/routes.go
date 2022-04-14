package rest

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) setupRoutes() {
	router := gin.New()

	router.Use(globalServerMiddleware()...)

	router.POST("/bulk_transfer", s.handleCreateBulkTransfer())

	s.server.Handler = router
}
