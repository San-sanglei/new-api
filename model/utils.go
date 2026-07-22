package model

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	BatchUpdateTypeUserQuota = iota
	BatchUpdateTypeTokenQuota
	BatchUpdateTypeUsedQuota
	BatchUpdateTypeChannelUsedQuota
	BatchUpdateTypeRequestCount
	BatchUpdateTypeCount // if you add a new type, you need to add a new map and a new lock
)

// MaxBatchUpdateRetry 控制失败记录的最大重试次数。
// 超过此次数后，记录不再无限 requeue，改为输出 critical 告警 + 最终同步落库尝试，
// 避免 DB 长期故障时 batchUpdateStores 内存无限增长。
const MaxBatchUpdateRetry = 10

var batchUpdateStores []map[int]int
var batchUpdateLocks []sync.Mutex

// batchUpdateRetryCounters 与 batchUpdateStores 平行，记录每个 (type, id) 的连续失败次数。
// 成功后归零；超过 MaxBatchUpdateRetry 后降级处理。
var batchUpdateRetryCounters []map[int]int

// batchUpdateWG 追踪 InitBatchUpdater goroutine 的生命周期。
// P3-5 修复：shutdown 时 FlushBatchUpdate 前必须先等待 ticker goroutine 退出，
//
//	避免 ticker 正在 batchUpdate() 中段时与 FlushBatchUpdate 并发执行。
//	不持锁、不持 DB transaction 等待，仅等待 goroutine 函数返回。
var batchUpdateWG sync.WaitGroup

func init() {
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateStores = append(batchUpdateStores, make(map[int]int))
		batchUpdateLocks = append(batchUpdateLocks, sync.Mutex{})
		batchUpdateRetryCounters = append(batchUpdateRetryCounters, make(map[int]int))
	}
}

// incRetryCount 递增 (type, id) 的失败计数并返回新值。
func incRetryCount(type_, id int) int {
	batchUpdateLocks[type_].Lock()
	defer batchUpdateLocks[type_].Unlock()
	batchUpdateRetryCounters[type_][id]++
	return batchUpdateRetryCounters[type_][id]
}

// resetRetryCount 成功后归零失败计数。
func resetRetryCount(type_, id int) {
	batchUpdateLocks[type_].Lock()
	defer batchUpdateLocks[type_].Unlock()
	delete(batchUpdateRetryCounters[type_], id)
}

// getRetryCount 返回 (type, id) 的当前失败计数。
func getRetryCount(type_, id int) int {
	batchUpdateLocks[type_].Lock()
	defer batchUpdateLocks[type_].Unlock()
	return batchUpdateRetryCounters[type_][id]
}

// getMaxUserRetryCount 返回用户维度三类 (UserQuota/UsedQuota/RequestCount) 的最大失败计数。
// 因为用户维度三类 delta 通过 updateUserQuotaUsedQuotaAndRequestCountWithError 原子写入，
// 它们的失败计数应保持一致；取 max 兼容仅部分 delta 非零的场景。
func getMaxUserRetryCount(id int) int {
	maxCount := 0
	for _, t := range []int{BatchUpdateTypeUserQuota, BatchUpdateTypeUsedQuota, BatchUpdateTypeRequestCount} {
		if c := getRetryCount(t, id); c > maxCount {
			maxCount = c
		}
	}
	return maxCount
}

// incUserRetryCounters 递增用户维度三类的失败计数（原子写入失败时一起递增）。
func incUserRetryCounters(id int) {
	for _, t := range []int{BatchUpdateTypeUserQuota, BatchUpdateTypeUsedQuota, BatchUpdateTypeRequestCount} {
		incRetryCount(t, id)
	}
}

// resetUserRetryCounters 成功后归零用户维度三类的失败计数。
func resetUserRetryCounters(id int) {
	for _, t := range []int{BatchUpdateTypeUserQuota, BatchUpdateTypeUsedQuota, BatchUpdateTypeRequestCount} {
		resetRetryCount(t, id)
	}
}

// InitBatchUpdater 启动批量更新后台任务。
// P1 修复：接受 ctx，服务退出时可通过 ctx.Done() 停止循环。
// P3-5 修复：通过 batchUpdateWG 追踪 goroutine 生命周期，shutdown 时
//
//	FlushBatchUpdate 前先 Wait 等待 goroutine 退出，避免并发执行 batchUpdate()。
func InitBatchUpdater(ctx context.Context) {
	batchUpdateWG.Add(1)
	common.GoSafeWithContext(func(ctx context.Context) {
		defer batchUpdateWG.Done()
		for {
			if !common.SleepWithContext(ctx, time.Duration(common.BatchUpdateInterval)*time.Second) {
				common.SysLog("InitBatchUpdater stopping: context cancelled")
				return
			}
			batchUpdate()
		}
	})
}

// FlushBatchUpdate 在服务关闭前同步刷盘内存中的 batchUpdateStores，
// 避免 SIGTERM/容器重启导致 quota/token/channel 增量丢失。
// 必须在 CloseDB 之前调用。内部直接复用 batchUpdate() 的快照+落库+失败重试逻辑。
// P3 修复：flush 后若仍有 pending 记录（DB 失败导致 requeue），输出 critical 汇总告警，
//
//	避免进程退出后 requeue 记录静默丢失。仅告警，不清空 store、不丢弃数据。
func FlushBatchUpdate() {
	if !common.BatchUpdateEnabled {
		return
	}
	// P3-5 修复：等待 InitBatchUpdater goroutine 完全退出后再执行 flush。
	//   避免与 ticker 正在执行的 batchUpdate() 并发，造成冗余执行或锁竞争。
	//   仅等待 goroutine 函数返回，不持锁、不持 DB transaction 等待。
	//   超时保护 30s，避免 ctx 失效或 goroutine 卡死导致 shutdown 永久阻塞。
	waitDone := make(chan struct{})
	go func() {
		batchUpdateWG.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
		// goroutine 已退出，继续 flush
	case <-time.After(30 * time.Second):
		common.SysError("FlushBatchUpdate: timeout waiting for InitBatchUpdater goroutine to exit: shutdown=true flush_failed=true manual_intervention_required=true")
		// 不 return，继续尝试 flush，尽量减少数据丢失
	}

	common.SysLog("flush batch update stores before shutdown")
	batchUpdate()

	// flush 后检查残留：batchUpdate 失败的记录会 requeue 到新 store，进程退出后会丢失。
	// 遍历所有 store 统计 pending 数量和累计 delta，使用现有 batchUpdateLocks 保证线程安全。
	pendingRecords := 0
	pendingDelta := 0
	pendingSummary := make([]string, 0, BatchUpdateTypeCount)
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		for id, delta := range batchUpdateStores[i] {
			pendingRecords++
			pendingDelta += delta
			// 记录每条残留的 (type, id, delta) 供运维手动恢复
			pendingSummary = append(pendingSummary, fmt.Sprintf("type=%d id=%d delta=%d", i, id, delta))
		}
		batchUpdateLocks[i].Unlock()
	}
	if pendingRecords > 0 {
		common.SysError(fmt.Sprintf(
			"critical: FlushBatchUpdate incomplete: shutdown=true flush_failed=true pending_records=%d pending_delta=%d manual_intervention_required=true details=[%s]",
			pendingRecords, pendingDelta, strings.Join(pendingSummary, ", "),
		))
	}

	common.SysLog("flush batch update stores done")
}

func addNewRecord(type_ int, id int, value int) {
	batchUpdateLocks[type_].Lock()
	defer batchUpdateLocks[type_].Unlock()
	if _, ok := batchUpdateStores[type_][id]; !ok {
		batchUpdateStores[type_][id] = value
	} else {
		batchUpdateStores[type_][id] += value
	}
}

func batchUpdate() {
	// check if there's any data to update
	hasData := false
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		if len(batchUpdateStores[i]) > 0 {
			hasData = true
			batchUpdateLocks[i].Unlock()
			break
		}
		batchUpdateLocks[i].Unlock()
	}

	if !hasData {
		return
	}

	common.SysLog("batch update started")
	stores := make([]map[int]int, BatchUpdateTypeCount)
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		stores[i] = batchUpdateStores[i]
		batchUpdateStores[i] = make(map[int]int)
		batchUpdateLocks[i].Unlock()
	}

	// P4-5 修复：DB 写入失败的记录必须重新入队，否则额度/用量数据永久丢失。
	// P3 修复：增加 MaxBatchUpdateRetry 限制，超过后降级处理（critical 告警 + 最终落库尝试），
	//   避免 DB 长期故障时 batchUpdateStores 内存无限增长。
	//   - retryCount < MaxBatchUpdateRetry：正常 requeue
	//   - retryCount >= MaxBatchUpdateRetry：不再 requeue，输出 critical 告警，
	//     当前写入即为最终同步落库尝试；若仍失败，记录 manual_intervention_required 告警。
	type failedRecord struct {
		type_ int
		id    int
		value int
		err   error
	}
	var failedRecords []failedRecord
	var failedCount, successCount, droppedCount int

	for i, store := range stores {
		if i == BatchUpdateTypeUserQuota || i == BatchUpdateTypeUsedQuota || i == BatchUpdateTypeRequestCount {
			continue
		}
		for key, value := range store {
			switch i {
			case BatchUpdateTypeTokenQuota:
				if err := increaseTokenQuota(key, value); err != nil {
					newRetry := incRetryCount(i, key)
					if newRetry >= MaxBatchUpdateRetry {
						// 超过最大重试次数，不再 requeue，记录 critical 告警
						common.SysError(fmt.Sprintf(
							"critical: batchUpdate increaseTokenQuota permanently failed after %d retries: manual_intervention_required=true dropped_from_queue=true token_id=%d delta_quota=%d err=%v",
							newRetry, key, value, err,
						))
						resetRetryCount(i, key)
						droppedCount++
					} else {
						common.SysError(fmt.Sprintf(
							"batchUpdate: increaseTokenQuota failed: requeued=true retry_count=%d/%d token_id=%d delta_quota=%d err=%v",
							newRetry, MaxBatchUpdateRetry, key, value, err,
						))
						failedRecords = append(failedRecords, failedRecord{i, key, value, err})
						failedCount++
					}
				} else {
					resetRetryCount(i, key)
					successCount++
				}
			case BatchUpdateTypeChannelUsedQuota:
				if err := DB.Model(&Channel{}).Where("id = ?", key).
					Update("used_quota", gorm.Expr("used_quota + ?", value)).Error; err != nil {
					newRetry := incRetryCount(i, key)
					if newRetry >= MaxBatchUpdateRetry {
						common.SysError(fmt.Sprintf(
							"critical: batchUpdate updateChannelUsedQuota permanently failed after %d retries: manual_intervention_required=true dropped_from_queue=true channel_id=%d delta_quota=%d err=%v",
							newRetry, key, value, err,
						))
						resetRetryCount(i, key)
						droppedCount++
					} else {
						common.SysError(fmt.Sprintf(
							"batchUpdate: updateChannelUsedQuota failed: requeued=true retry_count=%d/%d channel_id=%d delta_quota=%d err=%v",
							newRetry, MaxBatchUpdateRetry, key, value, err,
						))
						failedRecords = append(failedRecords, failedRecord{i, key, value, err})
						failedCount++
					}
				} else {
					resetRetryCount(i, key)
					successCount++
				}
			}
		}
	}

	userQuotaStore := stores[BatchUpdateTypeUserQuota]
	usedQuotaStore := stores[BatchUpdateTypeUsedQuota]
	requestCountStore := stores[BatchUpdateTypeRequestCount]

	userIDs := make(map[int]struct{}, len(userQuotaStore)+len(usedQuotaStore)+len(requestCountStore))
	for key := range userQuotaStore {
		userIDs[key] = struct{}{}
	}
	for key := range usedQuotaStore {
		userIDs[key] = struct{}{}
	}
	for key := range requestCountStore {
		userIDs[key] = struct{}{}
	}
	for key := range userIDs {
		quota, usedQuota, requestCount := userQuotaStore[key], usedQuotaStore[key], requestCountStore[key]
		if err := updateUserQuotaUsedQuotaAndRequestCountWithError(key, quota, usedQuota, requestCount); err != nil {
			retryCount := getMaxUserRetryCount(key)
			if retryCount >= MaxBatchUpdateRetry {
				// 超过最大重试次数，不再 requeue，记录 critical 告警
				common.SysError(fmt.Sprintf(
					"critical: batchUpdate updateUserQuotaUsedQuotaAndRequestCount permanently failed after %d retries: manual_intervention_required=true dropped_from_queue=true user_id=%d quota_delta=%d used_quota_delta=%d request_count_delta=%d err=%v",
					retryCount, key, quota, usedQuota, requestCount, err,
				))
				resetUserRetryCounters(key)
				droppedCount++
			} else {
				incUserRetryCounters(key)
				common.SysError(fmt.Sprintf(
					"batchUpdate: updateUserQuotaUsedQuotaAndRequestCount failed: requeued=true retry_count=%d/%d user_id=%d quota_delta=%d used_quota_delta=%d request_count_delta=%d err=%v",
					retryCount+1, MaxBatchUpdateRetry, key, quota, usedQuota, requestCount, err,
				))
				// 用户维度三类 delta 整体重试（与原批量写入原子性一致）
				if quota != 0 {
					failedRecords = append(failedRecords, failedRecord{BatchUpdateTypeUserQuota, key, quota, err})
				}
				if usedQuota != 0 {
					failedRecords = append(failedRecords, failedRecord{BatchUpdateTypeUsedQuota, key, usedQuota, err})
				}
				if requestCount != 0 {
					failedRecords = append(failedRecords, failedRecord{BatchUpdateTypeRequestCount, key, requestCount, err})
				}
				failedCount++
			}
		} else {
			resetUserRetryCounters(key)
			successCount++
		}
	}

	// 回填失败的记录到新 store，供下一周期重试
	if len(failedRecords) > 0 {
		for _, fr := range failedRecords {
			addNewRecord(fr.type_, fr.id, fr.value)
		}
		common.SysError(fmt.Sprintf(
			"batchUpdate: incomplete: requeued_records=%d, success_records=%d, failed_records=%d, dropped_records=%d, will_retry_next_cycle=true",
			len(failedRecords), successCount, failedCount, droppedCount,
		))
	} else if droppedCount > 0 {
		common.SysError(fmt.Sprintf(
			"batchUpdate: completed with dropped records: success_records=%d, dropped_records=%d, manual_intervention_required=true",
			successCount, droppedCount,
		))
	}
	common.SysLog("batch update finished")
}

func RecordExist(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func shouldUpdateRedis(fromDB bool, err error) bool {
	return common.RedisEnabled && fromDB && err == nil
}
