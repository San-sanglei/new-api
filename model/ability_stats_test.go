package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// GetGroupEnabledModels / GetEnabledModels / GetAllEnableAbilities 测试
//
// 修复前：Pluck/Find 错误被丢弃，DB 异常时静默返回空列表。
// 修复后：错误通过 common.SysError 记录，函数签名未变（无 error 返回值），
// 保持原有行为（返回空列表），但运维可通过日志定位 DB 异常。
// =====================================================================

// TestGetGroupEnabledModels_Success 验证成功路径下返回分组可用模型列表。
func TestGetGroupEnabledModels_Success(t *testing.T) {
	truncateTables(t)

	// 插入 ability 记录
	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "gpt-4",
		ChannelId: 1,
		Enabled:   true,
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "gpt-3.5-turbo",
		ChannelId: 1,
		Enabled:   true,
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "claude-3",
		ChannelId: 2,
		Enabled:   false, // disabled
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "vip",
		Model:     "gpt-4",
		ChannelId: 3,
		Enabled:   true,
	}).Error)

	models := GetGroupEnabledModels("default")
	// 应返回 default 分组下 enabled 的模型（去重）
	assert.Contains(t, models, "gpt-4")
	assert.Contains(t, models, "gpt-3.5-turbo")
	assert.NotContains(t, models, "claude-3", "disabled model should not be included")
}

// TestGetGroupEnabledModels_Empty 验证无数据时返回空列表。
func TestGetGroupEnabledModels_Empty(t *testing.T) {
	truncateTables(t)

	models := GetGroupEnabledModels("nonexistent")
	assert.Empty(t, models)
}

// TestGetEnabledModels_Success 验证成功路径下返回所有可用模型列表。
func TestGetEnabledModels_Success(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "gpt-4",
		ChannelId: 1,
		Enabled:   true,
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "vip",
		Model:     "claude-3",
		ChannelId: 2,
		Enabled:   true,
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "gpt-4", // duplicate model, different channel
		ChannelId: 3,
		Enabled:   true,
	}).Error)

	models := GetEnabledModels()
	// 应返回所有 enabled 的模型（去重）
	assert.Contains(t, models, "gpt-4")
	assert.Contains(t, models, "claude-3")
	// 去重后 gpt-4 只出现一次
	count := 0
	for _, m := range models {
		if m == "gpt-4" {
			count++
		}
	}
	assert.Equal(t, 1, count, "gpt-4 should appear only once after distinct")
}

// TestGetAllEnableAbilities_Success 验证成功路径下返回所有启用的 ability 记录。
func TestGetAllEnableAbilities_Success(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "gpt-4",
		ChannelId: 1,
		Enabled:   true,
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "gpt-3.5-turbo",
		ChannelId: 1,
		Enabled:   true,
	}).Error)
	require.NoError(t, DB.Create(&Ability{
		Group:     "default",
		Model:     "claude-3",
		ChannelId: 2,
		Enabled:   false, // disabled
	}).Error)

	abilities := GetAllEnableAbilities()
	assert.Len(t, abilities, 2, "should return only enabled abilities")
}

// TestGetAllEnableAbilities_Empty 验证无数据时返回空列表。
func TestGetAllEnableAbilities_Empty(t *testing.T) {
	truncateTables(t)

	abilities := GetAllEnableAbilities()
	assert.Empty(t, abilities)
}
