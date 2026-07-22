package middleware

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPStats 存储HTTP统计信息
type HTTPStats struct {
	activeConnections int64
	// P1-4: 基础监控指标（原子计数器，轻量无锁）
	totalRequests   int64
	totalErrors     int64 // status >= 400
	totalDurationNs int64 // 累计请求耗时（纳秒），用于计算平均耗时
}

var globalStats = &HTTPStats{}

// StatsMiddleware 统计中间件
func StatsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 增加活跃连接数
		atomic.AddInt64(&globalStats.activeConnections, 1)

		start := time.Now()

		// 确保在请求结束时减少连接数
		defer func() {
			atomic.AddInt64(&globalStats.activeConnections, -1)
			atomic.AddInt64(&globalStats.totalRequests, 1)
			elapsed := time.Since(start).Nanoseconds()
			atomic.AddInt64(&globalStats.totalDurationNs, elapsed)
			// 4xx/5xx 计为错误
			if c.Writer.Status() >= 400 {
				atomic.AddInt64(&globalStats.totalErrors, 1)
			}
		}()

		c.Next()
	}
}

// StatsInfo 统计信息结构
type StatsInfo struct {
	ActiveConnections int64 `json:"active_connections"`
	TotalRequests     int64 `json:"total_requests"`
	TotalErrors       int64 `json:"total_errors"`
	AvgLatencyMs      int64 `json:"avg_latency_ms"`
}

// GetStats 获取统计信息
func GetStats() StatsInfo {
	totalReq := atomic.LoadInt64(&globalStats.totalRequests)
	totalDur := atomic.LoadInt64(&globalStats.totalDurationNs)
	var avgLatencyMs int64
	if totalReq > 0 {
		avgLatencyMs = (totalDur / totalReq) / int64(time.Millisecond)
	}
	return StatsInfo{
		ActiveConnections: atomic.LoadInt64(&globalStats.activeConnections),
		TotalRequests:     totalReq,
		TotalErrors:       atomic.LoadInt64(&globalStats.totalErrors),
		AvgLatencyMs:      avgLatencyMs,
	}
}
