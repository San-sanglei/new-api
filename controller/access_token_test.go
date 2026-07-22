package controller

import (
	"net/http"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAccessTokenSelfRouter 构建带 UserAuth 的 self 路由，用于 GenerateAccessToken 测试。
func newAccessTokenSelfRouter(callerID int, callerUsername string, callerRole int) *gin.Engine {
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("access-token-test-secret"))))
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("id", callerID)
		session.Set("username", callerUsername)
		session.Set("role", callerRole)
		session.Set("status", common.UserStatusEnabled)
		session.Set("group", "default")
		c.Next()
	})

	r.GET("/api/user/token", middleware.UserAuth(), GenerateAccessToken)
	return r
}

// TestGenerateAccessToken_Success 验证正常生成 access_token 流程。
func TestGenerateAccessToken_Success(t *testing.T) {
	db := setupUserTestDB(t)
	user := createTestUserWithRole(t, db, "tokenuser", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	r := newAccessTokenSelfRouter(user.Id, user.Username, common.RoleCommonUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/token", nil, user.Id)

	assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected success, body: %s", w.Body.String())

	data, ok := resp["data"].(string)
	require.True(t, ok, "expected data to be string, body: %s", w.Body.String())
	assert.NotEmpty(t, data, "expected non-empty access token")

	// 验证 DB 中已更新
	var updated model.User
	require.NoError(t, db.First(&updated, user.Id).Error)
	require.NotNil(t, updated.AccessToken, "expected access_token to be non-nil in DB")
	assert.Equal(t, data, *updated.AccessToken, "expected DB access_token to match response")
}

// TestGenerateAccessToken_NotExists 验证不存在用户（session id 无效）时返回失败。
func TestGenerateAccessToken_NotExists(t *testing.T) {
	db := setupUserTestDB(t)
	_ = createTestUserWithRole(t, db, "realuser", "password123", common.UserStatusEnabled, common.RoleCommonUser)

	// 用一个不存在的 id 构造 session
	r := newAccessTokenSelfRouter(99999, "ghost", common.RoleCommonUser)
	w := doUserRequest(t, r, http.MethodGet, "/api/user/token", nil, 99999)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected failure for non-existent user, body: %s", w.Body.String())
}
