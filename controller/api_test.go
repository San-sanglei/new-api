package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.RedisEnabled = false
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	os.Exit(m.Run())
}

// setupAPITestDB creates an in-memory SQLite DB with all tables needed for
// login and token API tests. It reuses openTokenControllerTestDB from
// token_test.go for DB creation and common setting initialization, then adds
// the extra table migrations (User, Log, TwoFA) that the login flow touches.
func setupAPITestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := openTokenControllerTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Log{},
		&model.TwoFA{},
	))
	return db
}

// createTestUser inserts a user with a bcrypt-hashed password and the given
// status. The returned *model.User has the auto-assigned Id populated.
func createTestUser(t *testing.T, db *gorm.DB, username string, password string, status int) *model.User {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Status:   status,
		Role:     common.RoleCommonUser,
		Group:    "default",
	}
	require.NoError(t, db.Create(user).Error)
	return user
}

// --- Login API helpers ---

// newLoginRouter builds a gin engine wired with sessions middleware and the
// Login handler at POST /api/user/login.
func newLoginRouter() *gin.Engine {
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("api-test-secret"))))
	r.POST("/api/user/login", Login)
	return r
}

func doLoginRequest(t *testing.T, r *gin.Engine, username, password string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func decodeJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp), "failed to decode response body: %s", w.Body.String())
	return resp
}

// --- Login API tests ---

func TestLogin_Success(t *testing.T) {
	db := setupAPITestDB(t)
	common.SetPasswordLoginEnabled(true)

	createTestUser(t, db, "loginuser", "password123", common.UserStatusEnabled)

	r := newLoginRouter()
	w := doLoginRequest(t, r, "loginuser", "password123")

	assert.Equal(t, http.StatusOK, w.Code)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected success login, body: %s", w.Body.String())

	// A session cookie should be set on successful login.
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies, "expected session cookie to be set")
}

func TestLogin_WrongPassword(t *testing.T) {
	db := setupAPITestDB(t)
	common.SetPasswordLoginEnabled(true)

	createTestUser(t, db, "loginuser", "password123", common.UserStatusEnabled)

	r := newLoginRouter()
	w := doLoginRequest(t, r, "loginuser", "wrongpassword")

	assert.Equal(t, http.StatusOK, w.Code)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected login failure for wrong password, body: %s", w.Body.String())
}

func TestLogin_DisabledUser(t *testing.T) {
	db := setupAPITestDB(t)
	common.SetPasswordLoginEnabled(true)

	createTestUser(t, db, "disableduser", "password123", common.UserStatusDisabled)

	r := newLoginRouter()
	w := doLoginRequest(t, r, "disableduser", "password123")

	assert.Equal(t, http.StatusOK, w.Code)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, false, resp["success"], "expected login failure for disabled user, body: %s", w.Body.String())
}

// --- Token CRUD helpers ---

// newTokenRouter builds a gin engine with sessions middleware, a session-setting
// middleware that simulates an authenticated user, the real UserAuth middleware,
// and token CRUD routes. The New-Api-User header must still be set on each
// request (matching userID) because UserAuth requires it.
func newTokenRouter(userID int, username string) *gin.Engine {
	r := gin.New()
	r.Use(sessions.Sessions("session", cookie.NewStore([]byte("api-test-secret"))))
	r.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("id", userID)
		session.Set("username", username)
		session.Set("role", common.RoleCommonUser)
		session.Set("status", common.UserStatusEnabled)
		session.Set("group", "default")
		c.Next()
	})

	tokenGroup := r.Group("/api/token", middleware.UserAuth())
	tokenGroup.GET("/", GetAllTokens)
	tokenGroup.POST("/", AddToken)
	tokenGroup.DELETE("/:id", DeleteToken)

	return r
}

func doTokenRequest(t *testing.T, r *gin.Engine, method, path string, body []byte, userID int) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("New-Api-User", strconv.Itoa(userID))
	r.ServeHTTP(w, req)
	return w
}

// seedTokenDirectly inserts a Token row directly into the DB for testing list
// and delete endpoints.
func seedTokenDirectly(t *testing.T, db *gorm.DB, userID int, name string) *model.Token {
	t.Helper()
	token := &model.Token{
		UserId:         userID,
		Name:           name,
		Key:            "seededkey-" + name,
		Status:         common.TokenStatusEnabled,
		CreatedTime:    1,
		AccessedTime:   1,
		ExpiredTime:    -1,
		RemainQuota:    100,
		UnlimitedQuota: true,
		Group:          "default",
	}
	require.NoError(t, db.Create(token).Error)
	return token
}

// --- Token CRUD tests ---

func TestAddToken_Success(t *testing.T) {
	db := setupAPITestDB(t)
	user := createTestUser(t, db, "tokenuser", "password123", common.UserStatusEnabled)

	r := newTokenRouter(user.Id, user.Username)

	body, _ := json.Marshal(map[string]interface{}{
		"name":            "my-api-token",
		"remain_quota":    100,
		"unlimited_quota": false,
		"expired_time":    -1,
		"group":           "default",
	})
	w := doTokenRequest(t, r, http.MethodPost, "/api/token/", body, user.Id)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected token creation to succeed, body: %s", w.Body.String())

	// Verify the token was persisted in the database.
	var count int64
	db.Model(&model.Token{}).Where("user_id = ?", user.Id).Count(&count)
	assert.Equal(t, int64(1), count, "expected exactly one token in DB after creation")
}

func TestGetAllTokens_Success(t *testing.T) {
	db := setupAPITestDB(t)
	user := createTestUser(t, db, "tokenuser", "password123", common.UserStatusEnabled)

	seedTokenDirectly(t, db, user.Id, "list-token-1")
	seedTokenDirectly(t, db, user.Id, "list-token-2")

	r := newTokenRouter(user.Id, user.Username)
	w := doTokenRequest(t, r, http.MethodGet, "/api/token/?p=1&size=10", nil, user.Id)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected token list to succeed, body: %s", w.Body.String())

	// Verify paginated data structure.
	data, ok := resp["data"].(map[string]interface{})
	require.True(t, ok, "expected data to be a map, body: %s", w.Body.String())
	items, ok := data["items"].([]interface{})
	require.True(t, ok, "expected items to be an array, body: %s", w.Body.String())
	assert.Len(t, items, 2, "expected two tokens in the list")
}

func TestDeleteToken_Success(t *testing.T) {
	db := setupAPITestDB(t)
	user := createTestUser(t, db, "tokenuser", "password123", common.UserStatusEnabled)

	token := seedTokenDirectly(t, db, user.Id, "delete-me")

	r := newTokenRouter(user.Id, user.Username)
	w := doTokenRequest(t, r, http.MethodDelete, "/api/token/"+strconv.Itoa(token.Id), nil, user.Id)

	assert.Equal(t, http.StatusOK, w.Code)

	resp := decodeJSONResponse(t, w)
	assert.Equal(t, true, resp["success"], "expected token deletion to succeed, body: %s", w.Body.String())

	// Verify the token is no longer retrievable (soft-deleted via gorm.DeletedAt).
	var count int64
	db.Model(&model.Token{}).Where("id = ?", token.Id).Count(&count)
	assert.Equal(t, int64(0), count, "expected token to be soft-deleted")
}
