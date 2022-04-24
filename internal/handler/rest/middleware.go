package rest

import (
	"net/http"

	"github.com/angusgmorrison/hexagonal/pkg/slice"
	"github.com/gin-gonic/gin"
)

// serverMiddleware represents a chain of middleware in the order in which
// they'll be applied to a *gin.Engine. I.e. the first middleware in the stack
// represents the outermost layer in the HTTP call chain.
type serverMiddleware []gin.HandlerFunc

// globalDevelopmentMiddleware returns the middleware stack used for all routes
// when running in development.
func globalServerMiddleware() serverMiddleware {
	return serverMiddleware{
		gin.Logger(),
		gin.Recovery(),
		contentTypes(applicationJSON),
	}
}

type contentType string

const (
	applicationJSON contentType = "application/json"
)

func contentTypes(contentTypes ...contentType) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slice.Includes(contentTypes, contentType(c.ContentType())) {
			c.AbortWithStatus(http.StatusUnsupportedMediaType)

			return
		}

		c.Next()
	}
}
