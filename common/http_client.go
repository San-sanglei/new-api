package common

import (
	"net/http"
	"time"
)

// defaultShortRequestTransport 是用于短请求（OAuth 回调、Turnstile 校验、
// 渠道余额查询等非 relay 请求）的共享 Transport。
// P1-2 修复：统一 Transport 配置，避免各文件创建不一致的 http.Client 导致连接泄漏。
var defaultShortRequestTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     90 * time.Second,
}

// NewShortRequestClient 创建一个用于短请求的 http.Client，统一配置 Transport 和 Timeout。
// 短请求指非流式的、一次性的外部调用（如 OAuth 回调、Webhook 验签、余额查询）。
// timeout 为整体请求超时，建议 5-30 秒。
//
// 用法：
//
//	client := common.NewShortRequestClient(15 * time.Second)
//	resp, err := client.Do(req)
//	if err != nil { return err }
//	defer resp.Body.Close()
func NewShortRequestClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: defaultShortRequestTransport,
		Timeout:   timeout,
	}
}
