package rest

import "github.com/gin-gonic/gin"

// serverMiddleware represents a chain of middleware in the order in which
// they'll be applied to a *gin.Engine. I.e. the first middleware in the stack
// represents the outermost layer in the HTTP call chain.
type serverMiddleware []gin.HandlerFunc

// globalDevelopmentMiddleware returns the middleware stack used for all routes
// when running in development.
func globalServerMiddleware() serverMiddleware {
	return serverMiddleware{gin.Logger(), gin.Recovery()}
}
