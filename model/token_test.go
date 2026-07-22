package model

import (
	"errors"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateUserToken_Valid 验证有效 Token（status=enabled, remain_quota=100,
// expired_time=-1）能通过验证。
func TestValidateUserToken_Valid(t *testing.T) {
	truncateTables(t)

	const tokenKey = "sk-validate-valid"
	tok := &Token{
		Id:             1,
		UserId:         1,
		Key:            tokenKey,
		Name:           "valid_token",
		Status:         common.TokenStatusEnabled,
		RemainQuota:    100,
		UnlimitedQuota: false,
		ExpiredTime:    -1, // 永不过期
	}
	require.NoError(t, DB.Create(tok).Error)

	result, err := ValidateUserToken(tokenKey)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tokenKey, result.Key)
	assert.Equal(t, common.TokenStatusEnabled, result.Status)
}

// TestValidateUserToken_Disabled 验证已禁用 Token（status=disabled）返回
// ErrTokenInvalid。
func TestValidateUserToken_Disabled(t *testing.T) {
	truncateTables(t)

	const tokenKey = "sk-validate-disabled"
	tok := &Token{
		Id:             2,
		UserId:         1,
		Key:            tokenKey,
		Name:           "disabled_token",
		Status:         common.TokenStatusDisabled,
		RemainQuota:    100,
		UnlimitedQuota: false,
		ExpiredTime:    -1,
	}
	require.NoError(t, DB.Create(tok).Error)

	result, err := ValidateUserToken(tokenKey)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenInvalid),
		"应返回 ErrTokenInvalid，实际: %v", err)
	// ValidateUserToken 对无效状态返回 token 和 error
	require.NotNil(t, result)
	assert.Equal(t, common.TokenStatusDisabled, result.Status)
}

// TestValidateUserToken_Expired 验证已过期 Token（expired_time < now）返回
// ErrTokenInvalid。
func TestValidateUserToken_Expired(t *testing.T) {
	truncateTables(t)

	const tokenKey = "sk-validate-expired"
	tok := &Token{
		Id:             3,
		UserId:         1,
		Key:            tokenKey,
		Name:           "expired_token",
		Status:         common.TokenStatusEnabled,
		RemainQuota:    100,
		UnlimitedQuota: false,
		ExpiredTime:    1, // 1970 年，已过期
	}
	require.NoError(t, DB.Create(tok).Error)

	result, err := ValidateUserToken(tokenKey)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenInvalid),
		"应返回 ErrTokenInvalid，实际: %v", err)
	require.NotNil(t, result)
	// ValidateUserToken 会将状态更新为 TokenStatusExpired
	assert.Equal(t, common.TokenStatusExpired, result.Status)
}

// TestValidateUserToken_Exhausted 验证额度耗尽 Token（remain_quota=0,
// unlimited_quota=false）返回 ErrTokenInvalid。
func TestValidateUserToken_Exhausted(t *testing.T) {
	truncateTables(t)

	const tokenKey = "sk-validate-exhausted"
	tok := &Token{
		Id:             4,
		UserId:         1,
		Key:            tokenKey,
		Name:           "exhausted_token",
		Status:         common.TokenStatusEnabled,
		RemainQuota:    0, // 额度耗尽
		UnlimitedQuota: false,
		ExpiredTime:    -1,
	}
	require.NoError(t, DB.Create(tok).Error)

	result, err := ValidateUserToken(tokenKey)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenInvalid),
		"应返回 ErrTokenInvalid，实际: %v", err)
	require.NotNil(t, result)
	// ValidateUserToken 会将状态更新为 TokenStatusExhausted
	assert.Equal(t, common.TokenStatusExhausted, result.Status)
}

// TestValidateUserToken_EmptyKey 验证空 key 返回 ErrTokenNotProvided。
func TestValidateUserToken_EmptyKey(t *testing.T) {
	result, err := ValidateUserToken("")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenNotProvided),
		"应返回 ErrTokenNotProvided，实际: %v", err)
	assert.Nil(t, result)
}
