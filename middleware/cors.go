package middleware

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}

	// 通过环境变量 CORS_ALLOWED_ORIGINS 配置允许的域名，逗号分隔
	// 示例: CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com
	allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")

	if allowedOriginsStr != "" {
		// 生产模式：白名单
		origins := strings.Split(allowedOriginsStr, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		config.AllowOrigins = origins
		config.AllowCredentials = true
	} else {
		// 开发模式默认允许 localhost
		config.AllowOrigins = []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}
		config.AllowCredentials = true
	}

	return cors.New(config)
}

func PoweredBy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-New-Api-Version", common.Version)
		c.Next()
	}
}
