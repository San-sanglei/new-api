package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to open test db: " + err.Error())
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	model.DB = db
	common.UsingSQLite = true
	common.RedisEnabled = false
	if err := db.AutoMigrate(&model.User{}); err != nil {
		panic("failed to migrate: " + err.Error())
	}
	os.Exit(m.Run())
}

// --- test helpers ---

func createTestUser(t *testing.T, username string, role int, status int, accessToken string) *model.User {
	t.Helper()
	user := &model.User{
		Username: username,
		Password: "testpassword",
		Role:     role,
		Status:   status,
		AffCode:  username + "_aff",
		Group:    "default",
	}
	if accessToken != "" {
		user.SetAccessToken(accessToken)
	}
	require.NoError(t, model.DB.Create(user).Error)
	t.Cleanup(func() {
		model.DB.Exec("DELETE FROM users")
	})
	return user
}

func sessionSetter(username string, role int, id int, status int) func(c *gin.Context) {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("username", username)
		session.Set("role", role)
		session.Set("id", id)
		session.Set("status", status)
		session.Set("group", "default")
		_ = session.Save()
		c.Next()
	}
}

func newAuthRouter(authMiddleware gin.HandlerFunc, sessionSetup func(c *gin.Context)) *gin.Engine {
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("auth-test-secret"))))
	if sessionSetup != nil {
		r.Use(sessionSetup)
	}
	r.GET("/test", authMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	return r
}

func doAuthRequest(t *testing.T, r *gin.Engine, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

func parseAuthBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), "failed to parse response body: %s", w.Body.String())
	return body
}

func assertAuthSuccess(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, http.StatusOK, w.Code, "expected 200, got %d, body: %s", w.Code, w.Body.String())
	body := parseAuthBody(t, w)
	assert.Equal(t, true, body["success"], "expected auth to succeed, body: %s", w.Body.String())
}

func assertAuthFailed(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	body := parseAuthBody(t, w)
	assert.Equal(t, false, body["success"], "expected auth to fail, body: %s", w.Body.String())
}

// --- validUserInfo tests ---

func TestValidUserInfo_EmptyUsername_ReturnsFalse(t *testing.T) {
	assert.False(t, validUserInfo("", common.RoleCommonUser))
	assert.False(t, validUserInfo("   ", common.RoleCommonUser))
	assert.False(t, validUserInfo("\t\n", common.RoleCommonUser))
}

func TestValidUserInfo_InvalidRole_ReturnsFalse(t *testing.T) {
	assert.False(t, validUserInfo("testuser", 5))
	assert.False(t, validUserInfo("testuser", -1))
	assert.False(t, validUserInfo("testuser", 50))
}

func TestValidUserInfo_ValidInput_ReturnsTrue(t *testing.T) {
	assert.True(t, validUserInfo("testuser", common.RoleGuestUser))
	assert.True(t, validUserInfo("testuser", common.RoleCommonUser))
	assert.True(t, validUserInfo("admin", common.RoleAdminUser))
	assert.True(t, validUserInfo("root", common.RoleRootUser))
}

// --- UserAuth tests ---

func TestUserAuth_NoAuthInfo_Returns401(t *testing.T) {
	r := newAuthRouter(UserAuth(), nil)
	w := doAuthRequest(t, r, nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserAuth_ValidAccessToken_Passes(t *testing.T) {
	token := common.GetUUID()
	user := createTestUser(t, "tokenuser", common.RoleCommonUser, common.UserStatusEnabled, token)
	r := newAuthRouter(UserAuth(), nil)
	w := doAuthRequest(t, r, map[string]string{
		"Authorization": user.GetAccessToken(),
		"New-Api-User":  strconv.Itoa(user.Id),
	})
	assertAuthSuccess(t, w)
}

func TestUserAuth_InvalidAccessToken_ReturnsError(t *testing.T) {
	r := newAuthRouter(UserAuth(), nil)
	w := doAuthRequest(t, r, map[string]string{
		"Authorization": "nonexistent-token-00000000000000000000",
		"New-Api-User":  "1",
	})
	assertAuthFailed(t, w)
}

func TestUserAuth_DisabledUser_ReturnsBannedError(t *testing.T) {
	token := common.GetUUID()
	user := createTestUser(t, "disableduser", common.RoleCommonUser, common.UserStatusDisabled, token)
	r := newAuthRouter(UserAuth(), nil)
	w := doAuthRequest(t, r, map[string]string{
		"Authorization": user.GetAccessToken(),
		"New-Api-User":  strconv.Itoa(user.Id),
	})
	assertAuthFailed(t, w)
}

func TestUserAuth_CommonUser_Passes(t *testing.T) {
	r := newAuthRouter(UserAuth(), sessionSetter("normaluser", common.RoleCommonUser, 1, common.UserStatusEnabled))
	w := doAuthRequest(t, r, map[string]string{
		"New-Api-User": "1",
	})
	assertAuthSuccess(t, w)
}

// --- AdminAuth tests ---

func TestAdminAuth_CommonUser_ReturnsInsufficientPrivilege(t *testing.T) {
	r := newAuthRouter(AdminAuth(), sessionSetter("commonuser", common.RoleCommonUser, 1, common.UserStatusEnabled))
	w := doAuthRequest(t, r, map[string]string{
		"New-Api-User": "1",
	})
	assertAuthFailed(t, w)
}

func TestAdminAuth_AdminUser_Passes(t *testing.T) {
	r := newAuthRouter(AdminAuth(), sessionSetter("adminuser", common.RoleAdminUser, 1, common.UserStatusEnabled))
	w := doAuthRequest(t, r, map[string]string{
		"New-Api-User": "1",
	})
	assertAuthSuccess(t, w)
}

// --- RootAuth tests ---

func TestRootAuth_RootUser_Passes(t *testing.T) {
	r := newAuthRouter(RootAuth(), sessionSetter("rootuser", common.RoleRootUser, 1, common.UserStatusEnabled))
	w := doAuthRequest(t, r, map[string]string{
		"New-Api-User": "1",
	})
	assertAuthSuccess(t, w)
}

func TestRootAuth_AdminUser_ReturnsInsufficientPrivilege(t *testing.T) {
	r := newAuthRouter(RootAuth(), sessionSetter("adminuser", common.RoleAdminUser, 1, common.UserStatusEnabled))
	w := doAuthRequest(t, r, map[string]string{
		"New-Api-User": "1",
	})
	assertAuthFailed(t, w)
}
