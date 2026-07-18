package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/agusheryanto182/redis-playground/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func formatBytes(size int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func buildRequestMessage(ctx *fiber.Ctx) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(ctx.OriginalURL())
	result.WriteString(" - ")
	result.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
	result.WriteString(" ")
	result.WriteString(formatBytes(len(ctx.Response().Body())))

	return result.String()
}

func Logger(l logger.Interface) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		start := time.Now()

		err := ctx.Next()

		l.Info(
			"%s - duration=%s",
			buildRequestMessage(ctx),
			time.Since(start),
		)

		return err
	}
}
