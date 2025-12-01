package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/pkg/logger"
)

// Logger is a middleware that logs HTTP requests
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		fields := map[string]interface{}{
			"status":  statusCode,
			"method":  method,
			"path":    path,
			"query":   query,
			"ip":      clientIP,
			"latency": latency.String(),
		}

		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
			log.WithFields(fields).Error("Request completed with errors")
		} else if statusCode >= 500 {
			log.WithFields(fields).Error("Internal server error")
		} else if statusCode >= 400 {
			log.WithFields(fields).Warn("Client error")
		} else {
			log.WithFields(fields).Info("Request completed")
		}
	}
}
