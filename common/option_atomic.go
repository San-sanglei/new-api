package common

import (
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// atomic 配置访问层
// ---------------------------------------------------------------------------
//
// 背景：OptionMap 本身由 OptionMapRWMutex 保护，但 updateOptionMap 在持锁期间
// 还会写入大量全局变量（common.RegisterEnabled 等）。这些全局变量在 HTTP 请求
// 路径（如 controller/github.go）中无锁读取，与 SyncOptions 后台 goroutine 的
// 周期性写入形成数据竞争。
//
// Go 中 bool/string/int 的并发读写不是原子操作，可能导致：
//   - bool 读到半写入值（虽然实践中 x86 上 bool 读写通常是原子的，但 Go 规范不保证）
//   - string 读到不一致的指针+长度（致命，可导致崩溃或读到垃圾数据）
//   - race detector 报错
//
// 解决方案：为高频并发读取的关键配置变量提供 atomic 包装。
// 写入方（updateOptionMap）调用 SetXxxAtomic，读取方调用 GetXxxAtomic。
// atomic.Value 的 Load/Store 是原子且无锁的，性能优于 RWMutex。
//
// 注意：为保持向后兼容，原始全局变量仍然保留，atomic 函数与之同步。
// 新代码应优先使用 atomic 函数。

// configAtomic 存储 bool 类型配置的原子值
type boolConfig atomic.Value

func (b *boolConfig) load() bool {
	v := (*atomic.Value)(b).Load()
	if v == nil {
		return false
	}
	return v.(bool)
}

func (b *boolConfig) store(v bool) {
	(*atomic.Value)(b).Store(v)
}

// 关键并发配置的原子存储
var (
	registerEnabledAtomic       boolConfig
	passwordLoginEnabledAtomic  boolConfig
	passwordRegisterAtomic      boolConfig
	emailVerificationAtomic     boolConfig
	githubOAuthEnabledAtomic    boolConfig
	linuxdoOAuthEnabledAtomic   boolConfig
	wechatAuthEnabledAtomic     boolConfig
	telegramOAuthEnabledAtomic  boolConfig
	turnstileCheckEnabledAtomic boolConfig
)

// InitAtomicConfig 初始化原子配置，在 InitOptionMap 之前调用
func InitAtomicConfig() {
	registerEnabledAtomic.store(RegisterEnabled)
	passwordLoginEnabledAtomic.store(PasswordLoginEnabled)
	passwordRegisterAtomic.store(PasswordRegisterEnabled)
	emailVerificationAtomic.store(EmailVerificationEnabled)
	githubOAuthEnabledAtomic.store(GitHubOAuthEnabled)
	linuxdoOAuthEnabledAtomic.store(LinuxDOOAuthEnabled)
	wechatAuthEnabledAtomic.store(WeChatAuthEnabled)
	telegramOAuthEnabledAtomic.store(TelegramOAuthEnabled)
	turnstileCheckEnabledAtomic.store(TurnstileCheckEnabled)
}

// --- RegisterEnabled ---
func GetRegisterEnabled() bool  { return registerEnabledAtomic.load() }
func SetRegisterEnabled(v bool) { RegisterEnabled = v; registerEnabledAtomic.store(v) }

// --- PasswordLoginEnabled ---
func GetPasswordLoginEnabled() bool  { return passwordLoginEnabledAtomic.load() }
func SetPasswordLoginEnabled(v bool) { PasswordLoginEnabled = v; passwordLoginEnabledAtomic.store(v) }

// --- PasswordRegisterEnabled ---
func GetPasswordRegisterEnabled() bool  { return passwordRegisterAtomic.load() }
func SetPasswordRegisterEnabled(v bool) { PasswordRegisterEnabled = v; passwordRegisterAtomic.store(v) }

// --- EmailVerificationEnabled ---
func GetEmailVerificationEnabled() bool { return emailVerificationAtomic.load() }
func SetEmailVerificationEnabled(v bool) {
	EmailVerificationEnabled = v
	emailVerificationAtomic.store(v)
}

// --- GitHubOAuthEnabled ---
func GetGitHubOAuthEnabled() bool  { return githubOAuthEnabledAtomic.load() }
func SetGitHubOAuthEnabled(v bool) { GitHubOAuthEnabled = v; githubOAuthEnabledAtomic.store(v) }

// --- LinuxDOOAuthEnabled ---
func GetLinuxDOOAuthEnabled() bool  { return linuxdoOAuthEnabledAtomic.load() }
func SetLinuxDOOAuthEnabled(v bool) { LinuxDOOAuthEnabled = v; linuxdoOAuthEnabledAtomic.store(v) }

// --- WeChatAuthEnabled ---
func GetWeChatAuthEnabled() bool  { return wechatAuthEnabledAtomic.load() }
func SetWeChatAuthEnabled(v bool) { WeChatAuthEnabled = v; wechatAuthEnabledAtomic.store(v) }

// --- TelegramOAuthEnabled ---
func GetTelegramOAuthEnabled() bool  { return telegramOAuthEnabledAtomic.load() }
func SetTelegramOAuthEnabled(v bool) { TelegramOAuthEnabled = v; telegramOAuthEnabledAtomic.store(v) }

// --- TurnstileCheckEnabled ---
func GetTurnstileCheckEnabled() bool { return turnstileCheckEnabledAtomic.load() }
func SetTurnstileCheckEnabled(v bool) {
	TurnstileCheckEnabled = v
	turnstileCheckEnabledAtomic.store(v)
}
