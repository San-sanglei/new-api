package model

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestFillUserById_NotExists 验证 FillUserById 对不存在的用户 id 返回 gorm.ErrRecordNotFound。
// 修复前：FillUserById 丢弃 First() 错误，对不存在用户返回 nil 且不填充用户数据，
// 导致调用方权限检查被绕过。修复后：返回真实 DB 错误。
func TestFillUserById_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{Id: 99999}
	err := user.FillUserById()

	// 必须返回 error
	require.Error(t, err, "expected error for non-existent user id")
	// 必须是 gorm.ErrRecordNotFound
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"expected gorm.ErrRecordNotFound, got: %v", err)
	// 用户字段应保持零值（未被填充）
	assert.Equal(t, 0, user.Role, "expected zero Role for non-existent user")
	assert.Equal(t, "", user.Username, "expected empty Username for non-existent user")
}

// TestFillUserById_EmptyId 验证 id=0 时返回参数错误。
func TestFillUserById_EmptyId(t *testing.T) {
	truncateTables(t)

	user := &User{Id: 0}
	err := user.FillUserById()

	require.Error(t, err, "expected error for empty id")
	assert.Contains(t, err.Error(), "id 为空", "expected 'id 为空' error, got: %v", err)
}

// TestFillUserById_Success 验证存在用户能正确填充字段。
func TestFillUserById_Success(t *testing.T) {
	truncateTables(t)

	// 先创建一个用户
	created := &User{
		Username: "filltest",
		Password: "hashedpassword",
		Role:     1,
		Status:   1,
		Group:    "default",
	}
	require.NoError(t, DB.Create(created).Error)
	require.NotZero(t, created.Id)

	// 用 FillUserById 重新查询
	user := &User{Id: created.Id}
	err := user.FillUserById()

	require.NoError(t, err, "expected no error for existing user")
	assert.Equal(t, "filltest", user.Username)
	assert.Equal(t, 1, user.Role)
	assert.Equal(t, 1, user.Status)
}

// TestFillUserByGitHubId_NotExists 验证 FillUserByGitHubId 对不存在的 GitHubId 返回 gorm.ErrRecordNotFound。
// 修复前：丢弃 First() 错误，对不存在用户返回 nil 且不填充用户数据。修复后：返回真实 DB 错误。
func TestFillUserByGitHubId_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{GitHubId: "non-existent-github-id"}
	err := user.FillUserByGitHubId()

	require.Error(t, err, "expected error for non-existent GitHub id")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"expected gorm.ErrRecordNotFound, got: %v", err)
	assert.Equal(t, 0, user.Id, "expected zero Id for non-existent user")
}

// TestFillUserByDiscordId_NotExists 验证 FillUserByDiscordId 对不存在的 DiscordId 返回 gorm.ErrRecordNotFound。
func TestFillUserByDiscordId_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{DiscordId: "non-existent-discord-id"}
	err := user.FillUserByDiscordId()

	require.Error(t, err, "expected error for non-existent Discord id")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"expected gorm.ErrRecordNotFound, got: %v", err)
	assert.Equal(t, 0, user.Id, "expected zero Id for non-existent user")
}

// TestFillUserByOidcId_NotExists 验证 FillUserByOidcId 对不存在的 OidcId 返回 gorm.ErrRecordNotFound。
func TestFillUserByOidcId_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{OidcId: "non-existent-oidc-id"}
	err := user.FillUserByOidcId()

	require.Error(t, err, "expected error for non-existent Oidc id")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"expected gorm.ErrRecordNotFound, got: %v", err)
	assert.Equal(t, 0, user.Id, "expected zero Id for non-existent user")
}

// TestFillUserByWeChatId_NotExists 验证 FillUserByWeChatId 对不存在的 WeChatId 返回 gorm.ErrRecordNotFound。
func TestFillUserByWeChatId_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{WeChatId: "non-existent-wechat-id"}
	err := user.FillUserByWeChatId()

	require.Error(t, err, "expected error for non-existent WeChat id")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"expected gorm.ErrRecordNotFound, got: %v", err)
	assert.Equal(t, 0, user.Id, "expected zero Id for non-existent user")
}

// TestGetRootUser_NotExists 验证无 root 用户时 GetRootUser 返回 nil 且不 panic。
// 修复前：GetRootUser 丢弃 First() 错误，对不存在 root 用户返回 nil 指针，
// 调用方 service/user_notify.go 访问 .ToBaseUser() 时 panic。修复后：显式返回 nil。
func TestGetRootUser_NotExists(t *testing.T) {
	truncateTables(t)

	// 不创建任何 root 用户，调用应返回 nil 且不 panic
	user, err := GetRootUser()
	require.NoError(t, err, "expected no error when no root user exists")
	assert.Nil(t, user, "expected nil when no root user exists")
}

// TestGetRootUser_Success 验证存在 root 用户时 GetRootUser 返回正确用户。
func TestGetRootUser_Success(t *testing.T) {
	truncateTables(t)

	// 创建 root 用户
	created := &User{
		Username: "rootuser",
		Password: "hashedpassword",
		Role:     common.RoleRootUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "root",
	}
	require.NoError(t, DB.Create(created).Error)
	require.NotZero(t, created.Id)

	// 同时创建一个普通用户，验证不会误返回
	other := &User{
		Username: "otheruser",
		Password: "hashedpassword",
		Role:     common.RoleCommonUser,
		Status:   common.UserStatusEnabled,
		Group:    "default",
		AffCode:  "other",
	}
	require.NoError(t, DB.Create(other).Error)

	user, err := GetRootUser()
	require.NoError(t, err, "expected no error when root user exists")
	require.NotNil(t, user, "expected non-nil root user")
	assert.Equal(t, created.Id, user.Id, "expected correct root user id")
	assert.Equal(t, "rootuser", user.Username, "expected correct root username")
	assert.Equal(t, common.RoleRootUser, user.Role, "expected root role")
}

// TestFillUserByEmail_NotExists 验证 FillUserByEmail 对不存在的 email 返回 gorm.ErrRecordNotFound。
// 修复前：丢弃 First() 错误，对不存在 email 返回 nil。修复后：返回真实 DB 错误。
func TestFillUserByEmail_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{Email: "nonexistent@example.com"}
	err := user.FillUserByEmail()

	require.Error(t, err, "expected error for non-existent email")
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound),
		"expected gorm.ErrRecordNotFound, got: %v", err)
	assert.Equal(t, 0, user.Id, "expected zero Id for non-existent user")
}

// TestFillUserByEmail_Success 验证存在 email 能正确填充字段。
func TestFillUserByEmail_Success(t *testing.T) {
	truncateTables(t)

	created := &User{
		Username: "emailuser",
		Password: "hashedpassword",
		Email:    "test@example.com",
		Role:     1,
		Status:   1,
		Group:    "default",
		AffCode:  "emailtest",
	}
	require.NoError(t, DB.Create(created).Error)

	user := &User{Email: "test@example.com"}
	err := user.FillUserByEmail()

	require.NoError(t, err, "expected no error for existing email")
	assert.Equal(t, created.Id, user.Id)
	assert.Equal(t, "emailuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
}

// TestUserUpdate_NotExists 验证对不存在的 id 调用 Update 返回 error。
// 修复前：DB.First(&user, user.Id) 丢弃错误，后续 Updates 基于零值 user 执行。
// 修复后：First 失败立即返回 error，避免零值更新。
func TestUserUpdate_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{
		Id:       99999,
		Username: "ghost",
		Role:     1,
		Status:   1,
		Group:    "default",
	}
	err := user.Update(false)

	require.Error(t, err, "expected error for updating non-existent user")
}

// TestUserEdit_NotExists 验证对不存在的 id 调用 Edit 返回 error。
func TestUserEdit_NotExists(t *testing.T) {
	truncateTables(t)

	user := &User{
		Id:          99999,
		Username:    "ghost",
		DisplayName: "Ghost",
		Group:       "default",
		Remark:      "test",
	}
	err := user.Edit(false)

	require.Error(t, err, "expected error for editing non-existent user")
}
