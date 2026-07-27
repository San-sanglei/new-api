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
/**
 * Mock notification data for demonstration.
 *
 * Each mock notification reuses the existing announcement `type` field for
 * color coding (warning / error / default / success) and adds a `category`
 * field so the popover can render a distinct icon per notification kind.
 *
 * Categories map to the four requested notification types:
 * - `balance`  → 余额不足提醒 (insufficient balance)
 * - `channel`  → 渠道异常告警 (channel abnormal alert)
 * - `system`   → 系统公告 (system announcement)
 * - `feature`  → 新功能上线 (new feature release)
 */

export type NotificationCategory = 'balance' | 'channel' | 'system' | 'feature'

export interface MockNotification {
  id: number
  type: 'default' | 'warning' | 'error' | 'success'
  category: NotificationCategory
  title: string
  content: string
  extra?: string
  publishDate: string
  // Allow mock notifications to flow through the existing `Record<string, unknown>`
  // announcement pipeline without per-field casts.
  [key: string]: unknown
}

/**
 * Build mock notifications with `publishDate` relative to "now" so the
 * relative-time labels ("just now", "5 minutes ago"…) always look fresh.
 */
function buildMockNotifications(): MockNotification[] {
  const now = Date.now()
  const minutes = (m: number) => new Date(now - m * 60 * 1000).toISOString()
  const hours = (h: number) => new Date(now - h * 60 * 60 * 1000).toISOString()
  const days = (d: number) => new Date(now - d * 24 * 60 * 60 * 1000).toISOString()

  return [
    {
      id: 9001,
      type: 'warning',
      category: 'balance',
      title: '余额不足提醒',
      content:
        '您的账户余额已低于预设阈值（$2.00），为避免影响 API 调用，请及时充值。',
      extra: '当前余额：$1.25 ｜ 阈值：$2.00',
      publishDate: minutes(5),
    },
    {
      id: 9002,
      type: 'error',
      category: 'channel',
      title: '渠道异常告警',
      content:
        '渠道「Azure OpenAI (East US)」连续出现 5 次调用失败，已自动触发熔断。',
      extra: '错误率：32% ｜ 已重试 3 次 ｜ 建议检查密钥或切换渠道',
      publishDate: minutes(23),
    },
    {
      id: 9003,
      type: 'default',
      category: 'system',
      title: '系统公告',
      content:
        '平台将于本周日 02:00–04:00（北京时间）进行例行维护，期间服务可能短暂中断，请提前安排业务。',
      extra: '维护窗口：2026-06-22 02:00–04:00',
      publishDate: hours(2),
    },
    {
      id: 9004,
      type: 'success',
      category: 'feature',
      title: '新功能上线',
      content:
        'Playground 现已支持多模态对话（图像 + 文本），可在工作台直接上传图片进行调试。',
      extra: '支持模型：GPT-5、Claude Sonnet 4.5、Gemini 2.5 Pro',
      publishDate: days(1),
    },
    {
      id: 9005,
      type: 'warning',
      category: 'balance',
      title: '余额不足提醒',
      content: '令牌「sk-***prod」关联的额度即将耗尽，剩余可用额度不足 5%。',
      extra: '剩余额度：$0.48 ｜ 建议立即补充',
      publishDate: days(2),
    },
    {
      id: 9006,
      type: 'success',
      category: 'feature',
      title: '新功能上线',
      content: '渠道管理新增「自动重试 + 故障转移」策略，可在渠道详情页一键启用。',
      publishDate: days(3),
    },
  ]
}

/**
 * Stable mock notifications for the current session. Recomputed on first import
 * so the relative timestamps are fresh for each page load.
 */
export const mockNotifications: MockNotification[] = buildMockNotifications()
