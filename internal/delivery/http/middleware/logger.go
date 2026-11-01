package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLogger logs incoming HTTP requests with response time and status code
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		log.Printf(
			"[%s] %s | %d | %v | %s",
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			duration,
			c.IP(),
		)

		return err
	}
}
