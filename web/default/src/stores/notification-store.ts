/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface NotificationState {
  // Last read Notice content signature (full trimmed message)
  lastReadNotice: string
  // Array of read announcement keys (id or content hash)
  readAnnouncementKeys: string[]
  // Timestamp of last "Close Today" action
  closedUntilDate: string | null

  // Actions
  markNoticeRead: (noticeContent: string) => void
  markAnnouncementsRead: (keys: string[]) => void
  setClosedUntilDate: (date: string | null) => void
  isAnnouncementRead: (key: string) => boolean
  isNoticeClosed: () => boolean
}

// 默认游客存储 key（未登录或无法识别用户身份时使用）
const NOTIFICATION_STORAGE_KEY_GUEST = 'notification-storage-guest'

/**
 * 根据 localStorage 中的 user 信息构造按用户隔离的 persist key。
 *
 * 设计说明：
 * - Zustand persist 的 `name` 在 store 初始化时确定，运行时不会变更。
 * - 因此本函数仅在 store 模块加载时执行一次，依赖此时 localStorage 中的 `user` 字段。
 * - 页面刷新后：auth-store 已先于此模块在加载时从 localStorage 读取 user，故此处可直接读取。
 * - 登录/退出流程：sign-out-dialog 会调用 `window.location.reload()`，整个 store 重建后读到新的 key。
 * - SSR 安全：window 不存在时降级为 guest key。
 * - 脏数据安全：JSON 解析失败或 user 无 id 时降级为 guest key。
 */
function getNotificationStorageKey(): string {
  if (typeof window === 'undefined') {
    return NOTIFICATION_STORAGE_KEY_GUEST
  }
  try {
    const userJson = window.localStorage.getItem('user')
    if (!userJson) {
      return NOTIFICATION_STORAGE_KEY_GUEST
    }
    const user = JSON.parse(userJson) as { id?: unknown }
    if (user && typeof user.id === 'number' && user.id > 0) {
      return `notification-storage-${user.id}`
    }
  } catch {
    // JSON 解析失败或 localStorage 不可用，降级为 guest key
  }
  return NOTIFICATION_STORAGE_KEY_GUEST
}

/**
 * Notification store for tracking read status of Notice and Announcements
 * Persists to localStorage to maintain state across sessions.
 *
 * 按用户隔离：每个用户使用独立的 localStorage key，避免多账号共用浏览器时
 * 已读状态污染。游客统一使用 `notification-storage-guest`。
 */
export const useNotificationStore = create<NotificationState>()(
  persist(
    (set, get) => ({
      lastReadNotice: '',
      readAnnouncementKeys: [],
      closedUntilDate: null,

      markNoticeRead: (noticeContent: string) => {
        // Persist the full trimmed content so edits beyond 100 chars register
        const normalizedContent = noticeContent.trim()
        set({ lastReadNotice: normalizedContent })
      },

      markAnnouncementsRead: (keys: string[]) => {
        set((state) => ({
          readAnnouncementKeys: [
            ...new Set([...state.readAnnouncementKeys, ...keys]),
          ],
        }))
      },

      setClosedUntilDate: (date: string | null) => {
        set({ closedUntilDate: date })
      },

      isAnnouncementRead: (key: string) => {
        return get().readAnnouncementKeys.includes(key)
      },

      isNoticeClosed: () => {
        const { closedUntilDate } = get()
        if (!closedUntilDate) return false

        const today = new Date().toDateString()
        return closedUntilDate === today
      },
    }),
    {
      name: getNotificationStorageKey(),
      partialize: (state) => ({
        lastReadNotice: state.lastReadNotice,
        readAnnouncementKeys: state.readAnnouncementKeys,
        closedUntilDate: state.closedUntilDate,
      }),
    }
  )
)
