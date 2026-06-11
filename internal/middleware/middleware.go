package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-ID"

// RequestID is a middleware that injects a unique request ID into every request context
// and response header. If the incoming request already contains an X-Request-ID header,
// that value is reused; otherwise a new UUID is generated.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := c.Get(requestIDHeader)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		// Store in locals for use by downstream handlers and other middleware
		c.Locals("requestID", reqID)
		// Expose in response header
		c.Set(requestIDHeader, reqID)
		return c.Next()
	}
}

// RequestLogger is a middleware that logs each incoming request along with its
// duration and response status using Uber Zap structured logging.
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Retrieve the request ID set by the RequestID middleware
		reqID, _ := c.Locals("requestID").(string)

		duration := time.Since(start)

		log.Info("Request handled",
			zap.String("request_id", reqID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
		)

		return err
	}
}
