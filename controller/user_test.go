package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupUserTestDB 复用 setupAPITestDB，确保 User/Token/Log/TwoFA 表已迁移，
// 与 login/token 测试共享同一套 DB 初始化方式。
func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return setupAPITestDB(t)
}

// createTestUserWithRole 插入指定角色的用户，便于管理员/普通用户权限测试。
// 注意：直接 db.Create 不会触发注册流程中的 AffCode 生成，需手动设置唯一值避免 UNIQUE 约束冲突。
func createTestUserWithRole(t *testing.T, db *gorm.DB, username string, password string, status int, role int) *model.User {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Status:   status,
		Role:     role,
		Group:    "default",
		AffCode:  common.GetRandomString(8),
	}
	require.NoError(t, db.Create(user).Error)
	return user
}

// newUserAdminRouter 构建带 session 注入 + AdminAuth 的用户管理路由。
// callerRole 决定注入的角色（管理员或 root），用于权限测试。
func newUserAdminRouter(callerID int, callerUsername string, callerRole int) *gin.Engine {
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("user-test-secret"))))
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("id", callerID)
		session.Set("username", callerUsername)
		session.Set("role", callerRole)
		session.Set("status", common.UserStatusEnabled)
		session.Set("group", "default")
		c.Next()
	})

	adminGroup := r.Group("/api/user", middleware.AdminAuth())
	adminGroup.GET("/", GetAllUsers)
	adminGroup.GET("/search", SearchUsers)
	adminGroup.GET("/:id", GetUser)
	adminGroup.POST("/", CreateUser)
	adminGroup.POST("/manage", ManageUser)
	adminGroup.PUT("/", UpdateUser)
	adminGroup.DELETE("/:id", DeleteUser)
	adminGroup.DELETE("/:id/bindings/:binding_type", AdminClearUserBinding)
	adminGroup.DELETE("/:id/reset_passkey", AdminResetPasskey)
	adminGroup.DELETE("/:id/2fa", AdminDisable2FA)

	return r
}

// setupUserExtraTestDB 在 setupUserTestDB 基础上额外迁移 PasskeyCredential 和 TwoFABackupCode 表，
// 供 AdminResetPasskey / AdminDisable2FA 测试使用。
func setupUserExtraTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupUserTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.PasskeyCredential{},
		&model.TwoFABackupCode{},
	))
	return db
}

// doUserRequest 发起请求并设置 New-Api-User 头（AdminAuth 要求与 session id 一致）。
func doUserRequest(t *testing.T, r *gin.Engine, method, path string, body []byte, callerID int) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("New-Api-User", strconv.Itoa(callerID))
	r.ServeHTTP(w, req)
	return w
}

// =====================================================================
// GetAllUsers 测试
// =====================================================================

func TestGetAllUsers_Success(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	createTestUserWithRole(t, db, "user1", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	createTestUserWithRole(t, db, "user2", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/?p=1&size=10", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected success, body: %s", w.Body.String())

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data map, body: %s", w.Body.String())
	items, ok := data["items"].([]interface{})
	require.True(t, ok, "expected items array, body: %s", w.Body.String())
	assert.GreaterOrEqual(t, len(items), 3, "expected at least 3 users, body: %s", w.Body.String())
}

func TestGetAllUsers_Unauthenticated(t *testing.T) {
	setupUserTestDB(t)
	// 无 session 注入的路由，模拟未登录
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("user-test-secret"))))
	r.GET("/api/user/", middleware.AdminAuth(), GetAllUsers)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/?p=1", nil)
	r.ServeHTTP(w, req)

	// 未登录应被 AdminAuth 拦截（401 或 success=false）
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected auth failure for unauthenticated request, body: %s", w.Body.String())
}

func TestGetAllUsers_CommonUserDenied(t *testing.T) {
	db := setupUserTestDB(t)
	commonUser := createTestUserWithRole(t, db, "normal", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	// 普通用户角色注入，访问管理员接口
	r := newUserAdminRouter(commonUser.Id, commonUser.Username, common.RoleCommonUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/?p=1", nil, commonUser.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected permission denied for common user, body: %s", w.Body.String())
}

// =====================================================================
// UpdateUser 测试
// =====================================================================

func TestUpdateUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "targetuser", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	body, _ := json.Marshal(map[string]interface{}{
		"id":           target.Id,
		"username":     "targetuser-renamed",
		"display_name": "Updated Name",
		"group":        "vip",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPut, "/api/user/", body, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected update success, body: %s", w.Body.String())

	// 验证数据库已更新
	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, "Updated Name", updated.DisplayName)
	assert.Equal(t, "vip", updated.Group)
}

func TestUpdateUser_NotExists(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	body, _ := json.Marshal(map[string]interface{}{
		"id":       99999,
		"username": "ghost",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPut, "/api/user/", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
}

func TestUpdateUser_InvalidParams(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	// id=0 触发参数校验失败
	body, _ := json.Marshal(map[string]interface{}{
		"username": "no-id",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPut, "/api/user/", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for invalid params (id=0), body: %s", w.Body.String())
}

func TestUpdateUser_AdminCannotEditHigherRole(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	root := createTestUserWithRole(t, db, "root", "password123", common.UserStatusEnabled, common.RoleRootUser)

	// 管理员尝试更新 root 用户（权限不足）
	body, _ := json.Marshal(map[string]interface{}{
		"id":           root.Id,
		"username":     root.Username,
		"display_name": "hacked",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPut, "/api/user/", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected permission denied for admin editing root, body: %s", w.Body.String())
}

// =====================================================================
// DeleteUser 测试
// =====================================================================

func TestDeleteUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "todelete", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(target.Id), nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected delete success, body: %s", w.Body.String())

	// 验证已被硬删除
	var count int64
	db.Unscoped().Model(&model.User{}).Where("id = ?", target.Id).Count(&count)
	assert.Equal(t, int64(0), count, "expected user to be hard-deleted")
}

func TestDeleteUser_NotExists(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/99999", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
}

func TestDeleteUser_AdminCannotDeleteRoot(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	root := createTestUserWithRole(t, db, "root", "password123", common.UserStatusEnabled, common.RoleRootUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(root.Id), nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected permission denied for admin deleting root, body: %s", w.Body.String())
}

// =====================================================================
// ManageUser 测试
// =====================================================================

func TestManageUser_Disable(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "todisable", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	body, _ := json.Marshal(map[string]interface{}{
		"id":     target.Id,
		"action": "disable",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected disable success, body: %s", w.Body.String())

	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, common.UserStatusDisabled, updated.Status, "expected user to be disabled")
}

func TestManageUser_Enable(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "toenable", "password123", common.UserStatusDisabled, common.RoleCommonUser)

	body, _ := json.Marshal(map[string]interface{}{
		"id":     target.Id,
		"action": "enable",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected enable success, body: %s", w.Body.String())

	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, common.UserStatusEnabled, updated.Status, "expected user to be enabled")
}

func TestManageUser_AddQuota(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "quotauser", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	body, _ := json.Marshal(map[string]interface{}{
		"id":     target.Id,
		"action": "add_quota",
		"mode":   "add",
		"value":  500,
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected add_quota success, body: %s", w.Body.String())

	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, target.Quota+500, updated.Quota, "expected quota increased by 500")
}

func TestManageUser_Promote_OnlyRoot(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "promoteme", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	// 普通管理员尝试提升（应失败，仅 root 可提升）
	body, _ := json.Marshal(map[string]interface{}{
		"id":     target.Id,
		"action": "promote",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected promote failure for non-root admin, body: %s", w.Body.String())
}

func TestManageUser_Promote_ByRoot(t *testing.T) {
	db := setupUserTestDB(t)
	root := createTestUserWithRole(t, db, "root", "password123", common.UserStatusEnabled, common.RoleRootUser)
	target := createTestUserWithRole(t, db, "promoteme", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	body, _ := json.Marshal(map[string]interface{}{
		"id":     target.Id,
		"action": "promote",
	})
	r := newUserAdminRouter(root.Id, root.Username, common.RoleRootUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, root.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected promote success by root, body: %s", w.Body.String())

	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, common.RoleAdminUser, updated.Role, "expected user promoted to admin")
}

func TestManageUser_NotExists(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	// 修复后：ManageUser 使用 DB.First().Error 并在错误时返回 i18n.MsgUserNotExists，
	// 不再对不存在的 id 继续处理零值用户。此处验证返回失败且不进入具体 action。
	body, _ := json.Marshal(map[string]interface{}{
		"id":     99999,
		"action": "disable",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code, "expected HTTP 200, body: %s", w.Body.String())
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
}

func TestManageUser_InvalidAction(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "target", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	// 修复后：未知 action 进入 default 分支返回错误，不再 fall-through 到 user.Update(false)。
	// 验证返回失败且用户数据未被修改。
	body, _ := json.Marshal(map[string]interface{}{
		"id":     target.Id,
		"action": "invalid_action",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/manage", body, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code, "expected HTTP 200, body: %s", w.Body.String())
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for unknown action, body: %s", w.Body.String())

	// 验证用户状态与角色均未改变
	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, common.UserStatusEnabled, updated.Status, "expected user status unchanged for invalid action")
	assert.Equal(t, common.RoleCommonUser, updated.Role, "expected user role unchanged for invalid action")
}

// =====================================================================
// CreateUser 测试
// =====================================================================

func TestCreateUser_Success(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	body, _ := json.Marshal(map[string]interface{}{
		"username": "newuser",
		"password": "password123",
		"role":     common.RoleCommonUser,
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/", body, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected create success, body: %s", w.Body.String())

	// 数据库状态验证：用户已创建
	var created model.User
	require.NoError(t, db.Where("username = ?", "newuser").First(&created).Error)
	assert.Equal(t, common.RoleCommonUser, created.Role)
	assert.NotEmpty(t, created.AffCode, "expected AffCode auto-generated by Insert()")
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	createTestUserWithRole(t, db, "dupuser", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	body, _ := json.Marshal(map[string]interface{}{
		"username": "dupuser",
		"password": "password123",
		"role":     common.RoleCommonUser,
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for duplicate username, body: %s", w.Body.String())
}

func TestCreateUser_InvalidParams(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	// 空用户名触发参数校验失败
	body, _ := json.Marshal(map[string]interface{}{
		"username": "",
		"password": "password123",
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for empty username, body: %s", w.Body.String())
}

func TestCreateUser_CannotCreateHigherRole(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	// 管理员尝试创建另一个管理员（role >= myRole）
	body, _ := json.Marshal(map[string]interface{}{
		"username": "anotheradmin",
		"password": "password123",
		"role":     common.RoleAdminUser,
	})
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodPost, "/api/user/", body, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for creating higher/equal role, body: %s", w.Body.String())
}

func TestCreateUser_Unauthenticated(t *testing.T) {
	setupUserTestDB(t)
	// 无 session 注入的路由，模拟未登录
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("user-test-secret"))))
	r.POST("/api/user/", middleware.AdminAuth(), CreateUser)

	body, _ := json.Marshal(map[string]interface{}{
		"username": "newuser",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/", bytes.NewReader(body))
	r.ServeHTTP(w, req)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected auth failure for unauthenticated request, body: %s", w.Body.String())
}

// =====================================================================
// SearchUsers 测试
// =====================================================================

func TestSearchUsers_Success(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	createTestUserWithRole(t, db, "alice", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	createTestUserWithRole(t, db, "bob", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/search?keyword=alice", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected search success, body: %s", w.Body.String())

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data map, body: %s", w.Body.String())
	items, ok := data["items"].([]interface{})
	require.True(t, ok, "expected items array, body: %s", w.Body.String())
	assert.GreaterOrEqual(t, len(items), 1, "expected at least 1 result for 'alice', body: %s", w.Body.String())
}

func TestSearchUsers_EmptyKeyword(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	createTestUserWithRole(t, db, "alice", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	createTestUserWithRole(t, db, "bob", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	// 空关键词：LIKE "%%" 匹配所有用户
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/search?keyword=", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected search success for empty keyword, body: %s", w.Body.String())

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data map, body: %s", w.Body.String())
	items, ok := data["items"].([]interface{})
	require.True(t, ok, "expected items array, body: %s", w.Body.String())
	assert.GreaterOrEqual(t, len(items), 3, "expected at least 3 users for empty keyword, body: %s", w.Body.String())
}

func TestSearchUsers_NoResults(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/search?keyword=nonexistentuser12345", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected search success even with no results, body: %s", w.Body.String())

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data map, body: %s", w.Body.String())
	items, ok := data["items"].([]interface{})
	require.True(t, ok, "expected items array, body: %s", w.Body.String())
	assert.Equal(t, 0, len(items), "expected 0 results for non-matching keyword, body: %s", w.Body.String())
}

// =====================================================================
// GetUser 测试
// =====================================================================

func TestGetUser_Exists(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "targetuser", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/"+strconv.Itoa(target.Id), nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected get success, body: %s", w.Body.String())

	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data map, body: %s", w.Body.String())
	assert.Equal(t, float64(target.Id), data["id"], "expected correct user id, body: %s", w.Body.String())
	assert.Equal(t, "targetuser", data["username"], "expected correct username, body: %s", w.Body.String())
}

func TestGetUser_NotExists(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/99999", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
}

// =====================================================================
// AdminClearUserBinding 测试
// =====================================================================

func TestAdminClearUserBinding_Success(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "bounduser", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	// 设置 email 绑定
	require.NoError(t, db.Model(&model.User{}).Where("id = ?", target.Id).Update("email", "bound@example.com").Error)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(target.Id)+"/bindings/email", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected clear binding success, body: %s", w.Body.String())

	// 数据库状态验证：email 已被清空
	var updated model.User
	require.NoError(t, db.First(&updated, target.Id).Error)
	assert.Equal(t, "", updated.Email, "expected email binding cleared, body: %s", w.Body.String())
}

func TestAdminClearUserBinding_NotExists(t *testing.T) {
	db := setupUserTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/99999/bindings/email", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
}

// =====================================================================
// AdminResetPasskey 测试
// =====================================================================

func TestAdminResetPasskey_Success(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "passkeyuser", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	// 预置 passkey 凭证
	require.NoError(t, db.Create(&model.PasskeyCredential{
		UserID:       target.Id,
		CredentialID: "cred-id-base64-1",
		PublicKey:    "pub-key-base64-1",
	}).Error)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(target.Id)+"/reset_passkey", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected reset passkey success, body: %s", w.Body.String())

	// 数据库状态验证：passkey 已被删除
	var count int64
	db.Model(&model.PasskeyCredential{}).Where("user_id = ?", target.Id).Count(&count)
	assert.Equal(t, int64(0), count, "expected passkey credential deleted")
}

func TestAdminResetPasskey_PermissionDenied(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	root := createTestUserWithRole(t, db, "root", "password123", common.UserStatusEnabled, common.RoleRootUser)
	// 预置 root 的 passkey
	require.NoError(t, db.Create(&model.PasskeyCredential{
		UserID:       root.Id,
		CredentialID: "cred-id-base64-root",
		PublicKey:    "pub-key-base64-root",
	}).Error)

	// 管理员尝试重置 root 的 passkey（权限不足）
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(root.Id)+"/reset_passkey", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected permission denied for admin resetting root passkey, body: %s", w.Body.String())

	// 数据库状态验证：root 的 passkey 未被删除
	var count int64
	db.Model(&model.PasskeyCredential{}).Where("user_id = ?", root.Id).Count(&count)
	assert.Equal(t, int64(1), count, "expected root passkey not deleted")
}

func TestAdminResetPasskey_NotBound(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "nopasskey", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	// 不预置 passkey 凭证

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(target.Id)+"/reset_passkey", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for user without passkey, body: %s", w.Body.String())
}

// AdminResetPasskey 对不存在用户返回用户不存在错误（已修复）。
// 修复前：FillUserById 忽略 First() 错误，导致权限检查被绕过，
// 返回误导性的 "该用户尚未绑定 Passkey" 消息。
// 修复后：改用 model.GetUserById，正确返回 gorm.ErrRecordNotFound，
// 不再继续执行 Passkey 查询。
func TestAdminResetPasskey_NotExists(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/99999/reset_passkey", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	// 1. 不存在用户直接失败
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
	// 2. 不再返回 "尚未绑定 Passkey"，证明 3. 不进入 Passkey 查询流程
	msg, _ := resp["message"].(string)
	assert.NotContains(t, msg, "Passkey", "should not return misleading passkey message for non-existent user, body: %s", w.Body.String())
	assert.Contains(t, msg, "record not found", "expected gorm.ErrRecordNotFound message, body: %s", w.Body.String())
}

// =====================================================================
// AdminDisable2FA 测试
// =====================================================================

func TestAdminDisable2FA_Success(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "2fauser", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	// 预置 2FA 记录
	require.NoError(t, db.Create(&model.TwoFA{
		UserId:    target.Id,
		Secret:    "TOTPSECRET",
		IsEnabled: true,
	}).Error)

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(target.Id)+"/2fa", nil, admin.Id)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected disable 2FA success, body: %s", w.Body.String())

	// 数据库状态验证：2FA 记录已被硬删除
	var count int64
	db.Unscoped().Model(&model.TwoFA{}).Where("user_id = ?", target.Id).Count(&count)
	assert.Equal(t, int64(0), count, "expected 2FA record hard-deleted")
}

func TestAdminDisable2FA_PermissionDenied(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	root := createTestUserWithRole(t, db, "root", "password123", common.UserStatusEnabled, common.RoleRootUser)
	// 预置 root 的 2FA
	require.NoError(t, db.Create(&model.TwoFA{
		UserId:    root.Id,
		Secret:    "ROOTSECRET",
		IsEnabled: true,
	}).Error)

	// 管理员尝试禁用 root 的 2FA（权限不足）
	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(root.Id)+"/2fa", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected permission denied for admin disabling root 2FA, body: %s", w.Body.String())

	// 数据库状态验证：root 的 2FA 未被删除
	var count int64
	db.Model(&model.TwoFA{}).Where("user_id = ?", root.Id).Count(&count)
	assert.Equal(t, int64(1), count, "expected root 2FA not deleted")
}

func TestAdminDisable2FA_NotEnabled(t *testing.T) {
	db := setupUserExtraTestDB(t)
	admin := createTestUserWithRole(t, db, "admin", "password123", common.UserStatusEnabled, common.RoleAdminUser)
	target := createTestUserWithRole(t, db, "no2fa", "password123", common.UserStatusEnabled, common.RoleCommonUser)
	// 不预置 2FA 记录

	r := newUserAdminRouter(admin.Id, admin.Username, common.RoleAdminUser)
	w := doUserRequest(t, r, http.MethodDelete, "/api/user/"+strconv.Itoa(target.Id)+"/2fa", nil, admin.Id)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for user without 2FA, body: %s", w.Body.String())
}
