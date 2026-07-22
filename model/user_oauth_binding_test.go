package model

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// =====================================================================
// CreateUserOAuthBindingWithTx 测试
//
// 修复前：tx.Model(...).Count(&count) 错误被丢弃，
// 即使 DB 异常也会继续走唯一性判断（count==0 视为"未占用"），
// 然后调用 tx.Create，可能写入脏数据或再次失败但调用方得到误导性错误。
// 修复后：Count 错误必须直接返回，禁止继续走唯一性判断。
// =====================================================================

// TestCreateUserOAuthBindingWithTx_Success 验证正常路径下绑定创建成功。
func TestCreateUserOAuthBindingWithTx_Success(t *testing.T) {
	truncateTables(t)

	user := &User{
		Username:    "oauth_binding_user_ok",
		Password:    "hash",
		DisplayName: "OAuth OK",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, DB.Create(user).Error)

	binding := &UserOAuthBinding{
		UserId:         user.Id,
		ProviderId:     1,
		ProviderUserId: "provider-uid-1",
	}
	err := CreateUserOAuthBindingWithTx(DB, binding)
	require.NoError(t, err)
	assert.NotZero(t, binding.Id)

	var got UserOAuthBinding
	require.NoError(t, DB.First(&got, binding.Id).Error)
	assert.Equal(t, user.Id, got.UserId)
	assert.Equal(t, "provider-uid-1", got.ProviderUserId)
}

// TestCreateUserOAuthBindingWithTx_DuplicateUniqueIndex 验证唯一占用时返回错误。
// 通过预先插入同一 provider_user_id，触发 Count > 0 分支返回 "already bound"。
func TestCreateUserOAuthBindingWithTx_DuplicateUniqueIndex(t *testing.T) {
	truncateTables(t)

	user1 := &User{
		Username: "oauth_binding_user_dup1",
		Password: "hash", DisplayName: "Dup1",
		Role: common.RoleCommonUser, Status: common.UserStatusEnabled,
		AffCode: "dup1aff",
	}
	require.NoError(t, DB.Create(user1).Error)
	binding1 := &UserOAuthBinding{UserId: user1.Id, ProviderId: 2, ProviderUserId: "dup-uid"}
	require.NoError(t, CreateUserOAuthBindingWithTx(DB, binding1))

	user2 := &User{
		Username: "oauth_binding_user_dup2",
		Password: "hash", DisplayName: "Dup2",
		Role: common.RoleCommonUser, Status: common.UserStatusEnabled,
		AffCode: "dup2aff",
	}
	require.NoError(t, DB.Create(user2).Error)

	binding2 := &UserOAuthBinding{UserId: user2.Id, ProviderId: 2, ProviderUserId: "dup-uid"}
	err := CreateUserOAuthBindingWithTx(DB, binding2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already bound")
}

// TestCreateUserOAuthBindingWithTx_CountDBError 验证 Count 查询异常时直接返回错误，
// 且不会调用 tx.Create 写入脏数据。
//
// 通过在独立的 gorm.DB 上注册 Query Callback 拦截 Count 查询（Dest 为 *int64），
// 使 Count 失败但 Create 不受影响。这样可清晰区分修复前后行为：
//   - 修复前：Count 错误被丢弃（count==0 视为未占用），函数继续走 tx.Create（在本测试中
//     会成功），返回 nil，但写入一条脏绑定记录。
//   - 修复后：Count 错误立即返回，不调用 Create，无脏数据。
//
// 使用独立的 brokenDB（而非全局 DB）注册 Callback，避免影响其他测试。
// gorm v2 的 processor 没有公开的 Callback 删除方法，因此采用独立 DB 实例。
func TestCreateUserOAuthBindingWithTx_CountDBError(t *testing.T) {
	truncateTables(t)

	user := &User{
		Username:    "oauth_binding_user_dberr",
		Password:    "hash",
		DisplayName: "DBErr",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, DB.Create(user).Error)

	// 构造独立的内存 DB，避免污染全局 DB Callback。
	brokenDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, brokenDB.AutoMigrate(&UserOAuthBinding{}))

	// 注册 Count 拦截 Callback：仅拦截 Dest 为 *int64 的查询（即 Count 调用），
	// Find/Create 等不受影响，可清晰验证修复前后行为差异。
	brokenDB.Callback().Query().Before("gorm:query").Register("fail_count_for_oauth_binding_test", func(tx *gorm.DB) {
		if _, ok := tx.Statement.Dest.(*int64); ok {
			tx.AddError(errors.New("injected count error"))
		}
	})

	binding := &UserOAuthBinding{UserId: user.Id, ProviderId: 3, ProviderUserId: "dberr-uid"}
	err = CreateUserOAuthBindingWithTx(brokenDB, binding)
	// 修复前：err 为 nil 且脏数据被写入；修复后：err 应非 nil 且无 Create 调用。
	require.Error(t, err, "CreateUserOAuthBindingWithTx should return error when Count query fails")
	assert.Contains(t, err.Error(), "injected count error")

	// 验证 Create 未被调用（Find 不受 Count 拦截影响）
	var bindings []UserOAuthBinding
	require.NoError(t, brokenDB.Where("user_id = ?", user.Id).Find(&bindings).Error)
	assert.Equal(t, 0, len(bindings), "no binding should be created when Count fails")
}

// TestCreateUserOAuthBindingWithTx_ClosedTxError 验证传入已关闭的 tx 时返回错误。
// 作为 TestCreateUserOAuthBindingWithTx_CountDBError 的补充：closed tx 上所有查询都会失败，
// 这里仅断言函数不会静默返回 nil。
func TestCreateUserOAuthBindingWithTx_ClosedTxError(t *testing.T) {
	truncateTables(t)

	user := &User{
		Username:    "oauth_binding_user_closed",
		Password:    "hash",
		DisplayName: "Closed",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
	}
	require.NoError(t, DB.Create(user).Error)

	closedDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	closedSQLDB, err := closedDB.DB()
	require.NoError(t, err)
	require.NoError(t, closedSQLDB.Close())

	binding := &UserOAuthBinding{UserId: user.Id, ProviderId: 4, ProviderUserId: "closed-uid"}
	err = CreateUserOAuthBindingWithTx(closedDB, binding)
	require.Error(t, err, "should return error when tx is closed")
}
