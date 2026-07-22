package controller

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// HealthCheck 提供轻量级的健康检查端点，用于 Docker/K8s 探活。
//
// 检查项：
//   - 数据库连接（通过 model.PingDB）
//   - Redis 连接（如果启用）
//
// 返回：
//   - HTTP 200: 所有依赖正常
//   - HTTP 503: 任一依赖不可用
//
// 该端点无需鉴权，不经过限流，专供负载均衡器/容器探针使用。
// 为避免探针拖垮数据库，DB ping 内部有 10 秒缓存（见 model.PingDB）。
func HealthCheck(c *gin.Context) {
	status := gin.H{
		"status": "ok",
		"checks": gin.H{},
	}
	httpStatus := http.StatusOK

	// 检查数据库
	if err := model.PingDB(); err != nil {
		status["status"] = "degraded"
		status["checks"] = gin.H{"database": "fail"}
		httpStatus = http.StatusServiceUnavailable
	} else {
		status["checks"] = gin.H{"database": "ok"}
	}

	// 检查 Redis（如果启用）
	if common.RedisEnabled {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if common.RDB == nil {
			status["status"] = "degraded"
			status["checks"] = gin.H{"database": "ok", "redis": "fail"}
			httpStatus = http.StatusServiceUnavailable
		} else if err := common.RDB.Ping(ctx).Err(); err != nil {
			status["status"] = "degraded"
			status["checks"] = gin.H{"database": "ok", "redis": "fail"}
			httpStatus = http.StatusServiceUnavailable
		} else {
			status["checks"] = gin.H{"database": "ok", "redis": "ok"}
		}
	}

	c.JSON(httpStatus, status)
}

// GetMetrics 返回基础监控指标，供运维仪表盘或外部采集器使用。
//
// 指标包含：
//   - HTTP 层：总请求数、错误数（>=400）、平均延迟、活跃连接数
//   - 运行时：Goroutine 数、GC 次数、堆内存使用
//
// 该端点公开（无鉴权），但不包含敏感业务数据。
func GetMetrics(c *gin.Context) {
	stats := middleware.GetStats()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"http": gin.H{
			"total_requests":     stats.TotalRequests,
			"total_errors":       stats.TotalErrors,
			"avg_latency_ms":     stats.AvgLatencyMs,
			"active_connections": stats.ActiveConnections,
		},
		"runtime": gin.H{
			"goroutines":    runtime.NumGoroutine(),
			"gc_count":      memStats.NumGC,
			"heap_alloc_mb": memStats.HeapAlloc / 1024 / 1024,
			"sys_mem_mb":    memStats.Sys / 1024 / 1024,
		},
		"redis_enabled": common.RedisEnabled,
	})
}
