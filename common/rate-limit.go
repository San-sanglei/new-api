package common

import (
	"context"
	"sync"
	"time"
)

type InMemoryRateLimiter struct {
	store              map[string]*[]int64
	mutex              sync.Mutex
	expirationDuration time.Duration
}

func (l *InMemoryRateLimiter) Init(expirationDuration time.Duration) {
	if l.store == nil {
		l.mutex.Lock()
		if l.store == nil {
			l.store = make(map[string]*[]int64)
			l.expirationDuration = expirationDuration
			if expirationDuration > 0 {
				// P1-1 修复：使用 GoSafeWithContext 启动，支持 ctx 优雅退出 + panic recovery；
				// 使用 ticker + select 替代 time.Sleep 无限循环。
				GoSafeWithContext(l.clearExpiredItems)
			}
		}
		l.mutex.Unlock()
	}
}

// clearExpiredItems 定期清理过期的限流记录。
// P1-1 修复：接收 ctx 支持优雅退出，ticker 在退出时 Stop。
// panic recovery 由调用方的 GoSafeWithContext 提供。
func (l *InMemoryRateLimiter) clearExpiredItems(ctx context.Context) {
	ticker := time.NewTicker(l.expirationDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.mutex.Lock()
			now := time.Now().Unix()
			for key := range l.store {
				queue := l.store[key]
				size := len(*queue)
				if size == 0 || now-(*queue)[size-1] > int64(l.expirationDuration.Seconds()) {
					delete(l.store, key)
				}
			}
			l.mutex.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// Request parameter duration's unit is seconds
func (l *InMemoryRateLimiter) Request(key string, maxRequestNum int, duration int64) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	// [old <-- new]
	queue, ok := l.store[key]
	now := time.Now().Unix()
	if ok {
		if len(*queue) < maxRequestNum {
			*queue = append(*queue, now)
			return true
		} else {
			if now-(*queue)[0] >= duration {
				*queue = (*queue)[1:]
				*queue = append(*queue, now)
				return true
			} else {
				return false
			}
		}
	} else {
		s := make([]int64, 0, maxRequestNum)
		l.store[key] = &s
		*(l.store[key]) = append(*(l.store[key]), now)
	}
	return true
}
