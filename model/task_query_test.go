package model

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// GetTimedOutUnfinishedTasks / GetAllUnFinishSyncTasks — P4-2 DB Error 修复测试
//
// 覆盖目标：
//   - 查询成功（含空结果）：返回 (tasks, nil)
//   - 查询成功（含匹配结果）：返回正确的任务集合，业务语义（超时判断/状态过滤）不变
//   - DB 查询错误：返回 (nil, error)，调用方可据此区分"无任务"与"DB 故障"
// ---------------------------------------------------------------------------

// withBrokenDB 临时将全局 DB 替换为一个底层连接已关闭的 gorm.DB 实例，
// 用于触发 Find 查询的 DB Error。函数返回前恢复原 DB，避免污染其他测试。
// 使用独立的 sqlite 内存连接并在 swap 前关闭它，确保不影响主测试 DB。
func withBrokenDB(t *testing.T, fn func()) {
	t.Helper()
	origDB := DB
	brokenDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	brokenSqlDB, err := brokenDB.DB()
	require.NoError(t, err)
	// 关闭 broken 实例的底层连接，使后续 Find 查询返回错误。
	require.NoError(t, brokenSqlDB.Close())
	DB = brokenDB
	defer func() {
		DB = origDB
	}()
	fn()
}

// TestGetTimedOutUnfinishedTasks_EmptyResult verifies that an empty result set
// returns a non-nil empty slice and nil error (not (nil, nil) which is ambiguous).
func TestGetTimedOutUnfinishedTasks_EmptyResult(t *testing.T) {
	truncateTables(t)

	tasks, err := GetTimedOutUnfinishedTasks(0, 100)
	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Equal(t, 0, len(tasks))
}

// TestGetTimedOutUnfinishedTasks_FiltersByTimeoutAndStatus verifies that:
//   - tasks with progress=100% are excluded
//   - tasks with status=FAILURE or SUCCESS are excluded
//   - tasks with submit_time >= cutoff are excluded
//
// Business semantics (timeout judgment) preserved by P4-2 fix.
func TestGetTimedOutUnfinishedTasks_FiltersByTimeoutAndStatus(t *testing.T) {
	truncateTables(t)

	cutoff := int64(1000)

	cases := []struct {
		name     string
		task     *Task
		expected bool // expected to be returned by query
	}{
		{
			name:     "not_started_before_cutoff",
			task:     &Task{TaskID: "t1", Status: TaskStatusNotStart, Progress: "0%", SubmitTime: 500},
			expected: true,
		},
		{
			name:     "in_progress_before_cutoff",
			task:     &Task{TaskID: "t2", Status: TaskStatusInProgress, Progress: "50%", SubmitTime: 800},
			expected: true,
		},
		{
			name:     "submitted_after_cutoff_excluded",
			task:     &Task{TaskID: "t3", Status: TaskStatusInProgress, Progress: "50%", SubmitTime: 1500},
			expected: false,
		},
		{
			name:     "progress_100_excluded",
			task:     &Task{TaskID: "t4", Status: TaskStatusInProgress, Progress: "100%", SubmitTime: 500},
			expected: false,
		},
		{
			name:     "status_failure_excluded",
			task:     &Task{TaskID: "t5", Status: TaskStatusFailure, Progress: "50%", SubmitTime: 500},
			expected: false,
		},
		{
			name:     "status_success_excluded",
			task:     &Task{TaskID: "t6", Status: TaskStatusSuccess, Progress: "50%", SubmitTime: 500},
			expected: false,
		},
	}

	for _, c := range cases {
		insertTask(t, c.task)
	}

	tasks, err := GetTimedOutUnfinishedTasks(cutoff, 100)
	require.NoError(t, err)
	assert.Len(t, tasks, 2, "only t1 and t2 should match")

	gotIDs := map[string]bool{}
	for _, tk := range tasks {
		gotIDs[tk.TaskID] = true
	}
	assert.True(t, gotIDs["t1"])
	assert.True(t, gotIDs["t2"])
}

// TestGetTimedOutUnfinishedTasks_LimitEnforced verifies the limit parameter is honored.
func TestGetTimedOutUnfinishedTasks_LimitEnforced(t *testing.T) {
	truncateTables(t)

	for i := 0; i < 5; i++ {
		insertTask(t, &Task{
			TaskID:     "limit_t" + string(rune('a'+i)),
			Status:     TaskStatusInProgress,
			Progress:   "10%",
			SubmitTime: 100,
		})
	}

	tasks, err := GetTimedOutUnfinishedTasks(1000, 2)
	require.NoError(t, err)
	assert.Len(t, tasks, 2, "limit=2 should cap results")
}

// TestGetTimedOutUnfinishedTasks_DBError verifies the P4-2 fix: when the underlying
// DB connection is closed, Find must surface the error instead of silently returning nil.
// This is the key regression guard — without the fix the caller would treat nil as
// "no timed-out tasks" and skip the refund flow.
func TestGetTimedOutUnfinishedTasks_DBError(t *testing.T) {
	truncateTables(t)
	// Insert one row to ensure result set is non-empty; the error must still propagate.
	insertTask(t, &Task{
		TaskID:     "dberr_t1",
		Status:     TaskStatusInProgress,
		Progress:   "10%",
		SubmitTime: 100,
	})

	withBrokenDB(t, func() {
		tasks, err := GetTimedOutUnfinishedTasks(1000, 100)
		require.Error(t, err, "broken DB connection must surface as error, not nil slice")
		assert.Nil(t, tasks, "on DB error tasks must be nil to prevent caller treating len()==0 as normal")
	})
}

// TestGetAllUnFinishSyncTasks_EmptyResult verifies empty result returns non-nil slice and nil error.
func TestGetAllUnFinishSyncTasks_EmptyResult(t *testing.T) {
	truncateTables(t)

	tasks, err := GetAllUnFinishSyncTasks(100)
	require.NoError(t, err)
	assert.NotNil(t, tasks)
	assert.Equal(t, 0, len(tasks))
}

// TestGetAllUnFinishSyncTasks_FiltersByProgressAndStatus verifies that:
//   - progress=100% excluded
//   - status=FAILURE excluded
//   - status=SUCCESS excluded
//
// Business semantics preserved by P4-2 fix.
func TestGetAllUnFinishSyncTasks_FiltersByProgressAndStatus(t *testing.T) {
	truncateTables(t)

	cases := []struct {
		name     string
		task     *Task
		expected bool
	}{
		{"not_start", &Task{TaskID: "s1", Status: TaskStatusNotStart, Progress: "0%"}, true},
		{"in_progress", &Task{TaskID: "s2", Status: TaskStatusInProgress, Progress: "50%"}, true},
		{"progress_100_excluded", &Task{TaskID: "s3", Status: TaskStatusInProgress, Progress: "100%"}, false},
		{"status_failure_excluded", &Task{TaskID: "s4", Status: TaskStatusFailure, Progress: "50%"}, false},
		{"status_success_excluded", &Task{TaskID: "s5", Status: TaskStatusSuccess, Progress: "50%"}, false},
	}

	for _, c := range cases {
		insertTask(t, c.task)
	}

	tasks, err := GetAllUnFinishSyncTasks(100)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	gotIDs := map[string]bool{}
	for _, tk := range tasks {
		gotIDs[tk.TaskID] = true
	}
	assert.True(t, gotIDs["s1"])
	assert.True(t, gotIDs["s2"])
}

// TestGetAllUnFinishSyncTasks_LimitEnforced verifies the limit parameter is honored.
func TestGetAllUnFinishSyncTasks_LimitEnforced(t *testing.T) {
	truncateTables(t)

	for i := 0; i < 5; i++ {
		insertTask(t, &Task{
			TaskID:   "sync_limit_" + string(rune('a'+i)),
			Status:   TaskStatusInProgress,
			Progress: "10%",
		})
	}

	tasks, err := GetAllUnFinishSyncTasks(2)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

// TestGetAllUnFinishSyncTasks_DBError verifies the P4-2 fix: DB error must surface as error,
// not be swallowed into a nil slice that callers would misread as "no tasks to sync".
func TestGetAllUnFinishSyncTasks_DBError(t *testing.T) {
	truncateTables(t)
	insertTask(t, &Task{
		TaskID:   "sync_dberr_t1",
		Status:   TaskStatusInProgress,
		Progress: "10%",
	})

	withBrokenDB(t, func() {
		tasks, err := GetAllUnFinishSyncTasks(100)
		require.Error(t, err, "broken DB connection must surface as error")
		assert.Nil(t, tasks)
	})
}

// Compile-time guard: ensure functions return error type (signature drift detector).
var _ func(int64, int) ([]*Task, error) = GetTimedOutUnfinishedTasks
var _ func(int) ([]*Task, error) = GetAllUnFinishSyncTasks
