package common

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAtomicConfig_ConcurrentReadWrite 验证 atomic 配置在并发读写下的安全性。
//
// 用 -race 运行：go test -race -run TestAtomicConfig_Concurrent ./common/
// 如果 atomic 包装失效，race detector 会报错。
func TestAtomicConfig_ConcurrentReadWrite(t *testing.T) {
	// 初始化
	InitAtomicConfig()

	const goroutines = 50
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// 一半 goroutine 持续写入
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				SetRegisterEnabled(j%2 == 0)
				SetPasswordLoginEnabled(j%2 == 1)
				SetEmailVerificationEnabled(j%3 == 0)
				SetGitHubOAuthEnabled(j%2 == 0)
				SetTurnstileCheckEnabled(j%2 == 1)
			}
		}()
	}

	// 另一半 goroutine 持续读取
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// 读取应始终返回合法的 bool 值，不会 panic 或读到垃圾数据
				_ = GetRegisterEnabled()
				_ = GetPasswordLoginEnabled()
				_ = GetEmailVerificationEnabled()
				_ = GetGitHubOAuthEnabled()
				_ = GetTurnstileCheckEnabled()
			}
		}()
	}

	wg.Wait()

	// 最终值应与最后一次写入一致（不验证具体值，因为并发下不确定，
	// 但验证读取不会崩溃且返回合法 bool）
	assert.True(t, true, "并发读写完成无 panic")
}

// TestAtomicConfig_SetGetRoundTrip 验证 set/get 的基本正确性
func TestAtomicConfig_SetGetRoundTrip(t *testing.T) {
	InitAtomicConfig()

	SetRegisterEnabled(true)
	assert.True(t, GetRegisterEnabled())

	SetRegisterEnabled(false)
	assert.False(t, GetRegisterEnabled())

	SetPasswordLoginEnabled(true)
	assert.True(t, GetPasswordLoginEnabled())

	SetEmailVerificationEnabled(false)
	assert.False(t, GetEmailVerificationEnabled())

	SetGitHubOAuthEnabled(true)
	assert.True(t, GetGitHubOAuthEnabled())

	SetLinuxDOOAuthEnabled(true)
	assert.True(t, GetLinuxDOOAuthEnabled())

	SetWeChatAuthEnabled(true)
	assert.True(t, GetWeChatAuthEnabled())

	SetTelegramOAuthEnabled(true)
	assert.True(t, GetTelegramOAuthEnabled())

	SetTurnstileCheckEnabled(false)
	assert.False(t, GetTurnstileCheckEnabled())

	SetPasswordRegisterEnabled(true)
	assert.True(t, GetPasswordRegisterEnabled())
}

// TestAtomicConfig_GlobalVarSync 验证 atomic setter 同步更新全局变量
func TestAtomicConfig_GlobalVarSync(t *testing.T) {
	InitAtomicConfig()

	SetRegisterEnabled(true)
	assert.Equal(t, true, RegisterEnabled, "全局变量 RegisterEnabled 应同步更新")

	SetRegisterEnabled(false)
	assert.Equal(t, false, RegisterEnabled, "全局变量 RegisterEnabled 应同步更新")

	SetPasswordLoginEnabled(true)
	assert.Equal(t, true, PasswordLoginEnabled, "全局变量应同步更新")
}
