package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"

	"github.com/gin-gonic/gin"
)

const (
	T0RateLimitMark        = "T0RL"
	T0DailyLimit           = 10
	T0RateLimitDurationSec = 24 * 60 * 60 // 24 hours
)

// isT0Model 判断是否为 T0（高成本）模型
func isT0Model(modelName string) bool {
	name := strings.ToLower(modelName)

	// T1 排除列表（性价比模型，不限制）
	t1Excludes := []string{
		"mini", "nano", "flash", "haiku",
	}

	// GPT-4o 系列（不含 mini）
	if strings.HasPrefix(name, "gpt-4o") {
		return !containsAny(name, t1Excludes)
	}
	// GPT-4.1 系列（不含 mini/nano）
	if strings.HasPrefix(name, "gpt-4.1") {
		return !containsAny(name, t1Excludes)
	}
	// GPT-4.5
	if strings.HasPrefix(name, "gpt-4.5") {
		return true
	}
	// GPT-5 系列（不含 mini/nano）
	if strings.HasPrefix(name, "gpt-5") && !strings.HasPrefix(name, "gpt-5.") {
		return !containsAny(name, t1Excludes)
	}
	// GPT-4（不含 turbo，turbo 算 T1）
	if strings.HasPrefix(name, "gpt-4-") || name == "gpt-4" {
		return !strings.Contains(name, "turbo")
	}
	// Claude Sonnet/Opus 系列
	if strings.Contains(name, "claude-3-5-sonnet") ||
		strings.Contains(name, "claude-3-7-sonnet") ||
		strings.Contains(name, "claude-sonnet-4") ||
		strings.Contains(name, "claude-3-opus") ||
		strings.Contains(name, "claude-opus-4") {
		return true
	}
	// Gemini Pro 系列
	if strings.Contains(name, "gemini-1.5-pro") ||
		strings.Contains(name, "gemini-2.5-pro") ||
		strings.Contains(name, "gemini-3-pro") {
		return true
	}
	// o1/o3 系列（不含 mini）
	if strings.HasPrefix(name, "o1-") || name == "o1" {
		return !strings.Contains(name, "mini")
	}
	if strings.HasPrefix(name, "o3-") || name == "o3" {
		return !strings.Contains(name, "mini")
	}
	// Grok-3（不含 mini）
	if strings.HasPrefix(name, "grok-3") {
		return !strings.Contains(name, "mini")
	}

	return false
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// T0ModelRateLimit T0 模型每日限额中间件
// VIP 用户每天最多使用 10 次 T0 模型，用完后提示使用 T1 模型
func T0ModelRateLimit() func(c *gin.Context) {
	// 初始化内存限流器，24小时过期
	inMemoryRateLimiter.Init(time.Duration(T0RateLimitDurationSec) * time.Second)

	return func(c *gin.Context) {
		// 获取模型名称（由 Distribute 中间件设置）
		modelName := c.GetString("original_model")
		if modelName == "" {
			c.Next()
			return
		}

		// 不是 T0 模型，直接放行
		if !isT0Model(modelName) {
			c.Next()
			return
		}

		userId := c.GetInt("id")
		if userId == 0 {
			c.Next()
			return
		}

		// 获取用户分组
		group := common.GetContextKeyString(c, constant.ContextKeyUsingGroup)
		if group == "" {
			group = common.GetContextKeyString(c, constant.ContextKeyUserGroup)
		}

		// 免费用户不应该能访问 T0 模型（渠道分组已限制），但作为兜底
		if group != "vip" {
			c.Next()
			return
		}

		// 检查 T0 每日限额
		key := fmt.Sprintf("%s:%d", T0RateLimitMark, userId)
		// 先检查是否已达上限（使用临时 key）
		checkKey := key + "_check"
		if !inMemoryRateLimiter.Request(checkKey, T0DailyLimit, int64(T0RateLimitDurationSec)) {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests,
				fmt.Sprintf("今日 T0 高级模型（GPT-4o/Claude/Gemini Pro 等）使用次数已达上限（%d次/天），"+
					"请明天再试，或使用 GPT-4o-mini、Gemini Flash、DeepSeek 等性价比模型继续对话", T0DailyLimit))
			return
		}

		// 处理请求
		c.Next()

		// 如果请求成功，记录到实际的 T0 计数中
		if c.Writer.Status() < 400 {
			inMemoryRateLimiter.Request(key, T0DailyLimit, int64(T0RateLimitDurationSec))
		}
	}
}
