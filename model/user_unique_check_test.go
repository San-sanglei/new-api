package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// Is*AlreadyTaken 系列测试
//
// 修复前：所有函数用 Find().RowsAffected == 1 判断，DB 错误时返回 false（"未占用"），
// 可能导致重复邮箱/第三方 ID 绕过唯一性检查。
// 修复后：返回 (bool, error)，DB 错误时返回 error，调用方可拒绝操作。
// =====================================================================

// --- IsEmailAlreadyTaken ---

func TestIsEmailAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsEmailAlreadyTaken("nobody@example.com")
	require.NoError(t, err)
	assert.False(t, taken, "expected email to not be taken")
}

func TestIsEmailAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username: "emailuser",
		Password: "hashedpassword",
		Email:    "taken@example.com",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "emailtest",
	}).Error)

	taken, err := IsEmailAlreadyTaken("taken@example.com")
	require.NoError(t, err)
	assert.True(t, taken, "expected email to be taken")
}

// TestIsEmailAlreadyTaken_SoftDeleted 验证 Unscoped 查询能检测到软删除用户的邮箱。
func TestIsEmailAlreadyTaken_SoftDeleted(t *testing.T) {
	truncateTables(t)

	user := &User{
		Username: "deleteduser",
		Password: "hashedpassword",
		Email:    "deleted@example.com",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "deleted",
	}
	require.NoError(t, DB.Create(user).Error)
	require.NoError(t, DB.Delete(user).Error) // 软删除

	taken, err := IsEmailAlreadyTaken("deleted@example.com")
	require.NoError(t, err)
	assert.True(t, taken, "expected soft-deleted user's email to be taken (unscoped)")
}

// --- IsWeChatIdAlreadyTaken ---

func TestIsWeChatIdAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsWeChatIdAlreadyTaken("nonexistent-wechat-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsWeChatIdAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username: "wechatuser",
		Password: "hashedpassword",
		WeChatId: "wechat-123",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "wechattest",
	}).Error)

	taken, err := IsWeChatIdAlreadyTaken("wechat-123")
	require.NoError(t, err)
	assert.True(t, taken)
}

// --- IsGitHubIdAlreadyTaken ---

func TestIsGitHubIdAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsGitHubIdAlreadyTaken("nonexistent-github-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsGitHubIdAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username: "githubuser",
		Password: "hashedpassword",
		GitHubId: "github-456",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "githubtest",
	}).Error)

	taken, err := IsGitHubIdAlreadyTaken("github-456")
	require.NoError(t, err)
	assert.True(t, taken)
}

// --- IsDiscordIdAlreadyTaken ---

func TestIsDiscordIdAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsDiscordIdAlreadyTaken("nonexistent-discord-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsDiscordIdAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username:  "discorduser",
		Password:  "hashedpassword",
		DiscordId: "discord-789",
		Role:      common.RoleCommonUser,
		Status:    common.UserStatusEnabled,
		Group:     "default",
		AffCode:   "discordtest",
	}).Error)

	taken, err := IsDiscordIdAlreadyTaken("discord-789")
	require.NoError(t, err)
	assert.True(t, taken)
}

// --- IsOidcIdAlreadyTaken ---

func TestIsOidcIdAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsOidcIdAlreadyTaken("nonexistent-oidc-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsOidcIdAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username: "oidcuser",
		Password: "hashedpassword",
		OidcId:   "oidc-abc",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "oidctest",
	}).Error)

	taken, err := IsOidcIdAlreadyTaken("oidc-abc")
	require.NoError(t, err)
	assert.True(t, taken)
}

// --- IsTelegramIdAlreadyTaken ---

func TestIsTelegramIdAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsTelegramIdAlreadyTaken("nonexistent-telegram-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsTelegramIdAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username:   "telegramuser",
		Password:   "hashedpassword",
		TelegramId: "telegram-xyz",
		Role:       common.RoleCommonUser,
		Status:     common.UserStatusEnabled,
		Group:      "default",
		AffCode:    "telegramtest",
	}).Error)

	taken, err := IsTelegramIdAlreadyTaken("telegram-xyz")
	require.NoError(t, err)
	assert.True(t, taken)
}

// --- IsLinuxDOIdAlreadyTaken ---

func TestIsLinuxDOIdAlreadyTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsLinuxDOIdAlreadyTaken("nonexistent-linuxdo-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsLinuxDOIdAlreadyTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Username:  "linuxdouser",
		Password:  "hashedpassword",
		LinuxDOId: "linuxdo-001",
		Role:      common.RoleCommonUser,
		Status:    common.UserStatusEnabled,
		Group:     "default",
		AffCode:   "linuxdotest",
	}).Error)

	taken, err := IsLinuxDOIdAlreadyTaken("linuxdo-001")
	require.NoError(t, err)
	assert.True(t, taken)
}

// --- IsProviderUserIdTaken ---

func TestIsProviderUserIdTaken_NotTaken(t *testing.T) {
	truncateTables(t)

	taken, err := IsProviderUserIdTaken(1, "nonexistent-provider-user-id")
	require.NoError(t, err)
	assert.False(t, taken)
}

func TestIsProviderUserIdTaken_Taken(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&UserOAuthBinding{
		UserId:         100,
		ProviderId:     1,
		ProviderUserId: "provider-user-123",
	}).Error)

	taken, err := IsProviderUserIdTaken(1, "provider-user-123")
	require.NoError(t, err)
	assert.True(t, taken)
}

// TestIsProviderUserIdTaken_DifferentProvider 验证不同 provider 的相同 userId 不算占用。
func TestIsProviderUserIdTaken_DifferentProvider(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&UserOAuthBinding{
		UserId:         100,
		ProviderId:     1,
		ProviderUserId: "shared-user-id",
	}).Error)

	// provider=2 查询相同 providerUserId 应返回 false
	taken, err := IsProviderUserIdTaken(2, "shared-user-id")
	require.NoError(t, err)
	assert.False(t, taken, "expected false for different provider")
}
