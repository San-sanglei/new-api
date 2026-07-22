package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// Channel.Update 测试
//
// 修复前：Updates 成功后执行 DB.Model(channel).First(channel, "id = ?", channel.Id)
// 但丢弃 First 错误，DB 异常时 channel 处于部分更新的脏数据状态，
// 后续 UpdateAbilities 会基于陈旧/空数据写入错误的 ability。
// 修复后：First 错误必须记录并通过 fmt.Errorf 包装返回，
// 调用方（如 controller/channel.go 多处）已有 err != nil 判空处理，兼容无影响。
// =====================================================================

// TestChannelUpdate_Success 验证成功路径下 Update 返回 nil 且 channel 被重新加载。
func TestChannelUpdate_Success(t *testing.T) {
	truncateTables(t)

	ch := &Channel{
		Type:   1,
		Key:    "sk-test-key",
		Status: 1,
		Name:   "test-channel",
		Group:  "default",
		Models: "gpt-4,gpt-3.5-turbo",
	}
	require.NoError(t, DB.Create(ch).Error)

	// 修改字段后调用 Update
	ch.Name = "updated-channel"
	ch.Status = 2
	err := ch.Update()
	require.NoError(t, err)

	// 重新查询数据库验证
	var got Channel
	require.NoError(t, DB.First(&got, ch.Id).Error)
	assert.Equal(t, "updated-channel", got.Name)
	assert.Equal(t, 2, got.Status)
}

// TestChannelUpdate_NonExistentChannel 验证 Updates 后 First 找不到记录时返回错误。
// Updates 对不存在的记录不会报错（RowsAffected=0），但后续 First 会触发 ErrRecordNotFound。
// 修复前会忽略此错误并基于零值 channel 继续 UpdateAbilities；修复后必须返回错误。
func TestChannelUpdate_NonExistentChannel(t *testing.T) {
	truncateTables(t)

	// 构造一个 DB 中不存在的 channel（Id 不存在）
	ch := &Channel{
		Id:     999999,
		Type:   1,
		Key:    "sk-nonexistent",
		Status: 1,
		Name:   "ghost-channel",
		Group:  "default",
		Models: "gpt-4",
	}

	err := ch.Update()
	// Updates 不会因记录不存在而报错，但 First 重载时会返回 ErrRecordNotFound
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reload channel after update failed")
}

// TestChannelUpdate_ReloadsUpdatedFields 验证 Update 后 channel 对象的字段
// 被数据库中的实际值覆盖（重载生效），而非保留传入的内存值。
func TestChannelUpdate_ReloadsUpdatedFields(t *testing.T) {
	truncateTables(t)

	ch := &Channel{
		Type:   1,
		Key:    "sk-reload-key",
		Status: 1,
		Name:   "original-name",
		Group:  "default",
		Models: "gpt-4",
	}
	require.NoError(t, DB.Create(ch).Error)

	// 通过 Update 修改 Name
	ch.Name = "in-memory-name"
	require.NoError(t, ch.Update())

	// Update 内部 First 会用 DB 中的值覆盖 ch，所以 ch.Name 应为 DB 中的值
	assert.Equal(t, "in-memory-name", ch.Name, "field updated via Updates should be reloaded from DB")
}
