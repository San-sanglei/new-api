package common

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
)

// ---------------------------------------------------------------------------
// 应用生命周期管理
// ---------------------------------------------------------------------------
//
// 背景：原项目后台 goroutine 使用 `for { time.Sleep() }` 无限循环，无退出机制。
// 服务收到 SIGTERM/SIGINT 时所有 goroutine 无法停止，导致：
//   - 部署重启时连接泄露、数据不一致
//   - Kubernetes 滚动更新超时（默认 30s 后强杀）
//   - 优雅关闭无法实现
//
// 方案：创建根 context，所有后台任务通过 GoSafeWithContext 启动并监听 ctx.Done()。
// 服务退出时调用 LifecycleCancel()，所有任务收到取消信号后退出。

// lifecycleCtx 是应用生命周期的根 context
var lifecycleCtx context.Context

// lifecycleCancel 取消根 context，通知所有后台任务退出
var lifecycleCancel context.CancelFunc

// InitLifecycle 初始化应用生命周期 context。
// 应在 main() 早期调用，所有后台任务通过 LifecycleContext() 获取 ctx。
func InitLifecycle() {
	lifecycleCtx, lifecycleCancel = context.WithCancel(context.Background())
}

// LifecycleContext 返回应用生命周期根 context。
// 后台任务应通过 select { case <-ctx.Done(): return } 监听退出信号。
func LifecycleContext() context.Context {
	if lifecycleCtx == nil {
		// 兜底：未初始化时返回 background，避免 nil panic
		return context.Background()
	}
	return lifecycleCtx
}

// ShutdownLifecycle 取消根 context，通知所有后台任务退出。
// 由 main.go 的信号处理函数调用。
func ShutdownLifecycle() {
	if lifecycleCancel != nil {
		lifecycleCancel()
	}
}

// GoSafe 启动一个带 panic recovery 的 goroutine。
// 后台任务必须使用此函数而非直接 go func()，防止 panic 导致进程退出。
//
// 用法：
//
//	common.GoSafe(func() {
//	    for {
//	        select {
//	        case <-common.LifecycleContext().Done():
//	            return
//	        case <-time.After(60 * time.Second):
//	            doWork()
//	        }
//	    }
//	})
func GoSafe(fn func()) {
	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				SysError(fmt.Sprintf("goroutine panic recovered: %v\n%s", r, stack))
			}
		}()
		fn()
	})
}

// GoSafeWithContext 启动一个带 panic recovery 和 context 传入的 goroutine。
// 后台长循环任务应使用此函数，通过 ctx 接收退出信号。
//
// 用法：
//
//	common.GoSafeWithContext(func(ctx context.Context) {
//	    ticker := time.NewTicker(60 * time.Second)
//	    defer ticker.Stop()
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return
//	        case <-ticker.C:
//	            doWork(ctx)
//	        }
//	    }
//	})
func GoSafeWithContext(fn func(ctx context.Context)) {
	gopool.Go(func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				SysError(fmt.Sprintf("goroutine panic recovered: %v\n%s", r, stack))
			}
		}()
		fn(LifecycleContext())
	})
}

// SleepWithContext 在指定时间内阻塞，但可被 ctx 取消提前返回。
// 返回 true 表示正常超时，false 表示 ctx 被取消。
//
// 用于替代后台循环中的 time.Sleep：
//
//	for {
//	    if !common.SleepWithContext(ctx, 60*time.Second) {
//	        return // ctx 被取消
//	    }
//	    doWork()
//	}
func SleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}
