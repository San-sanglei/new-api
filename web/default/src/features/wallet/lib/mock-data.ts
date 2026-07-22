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
 * Mock data + types for the wallet recharge demo.
 *
 * Extends the existing wallet feature with demo-only payment methods
 * (bank transfer), a refunded status, and sample orders — all client-side
 * so the demo works without a backend.
 */

/** Demo payment methods (extends the real PAYMENT_TYPES with bank transfer) */
export const DEMO_PAYMENT_METHODS = [
  { id: 'alipay', name: '支付宝', color: '#1677FF' },
  { id: 'wxpay', name: '微信支付', color: '#07C160' },
  { id: 'bank_transfer', name: '对公转账', color: '#6B7280' },
] as const

export type DemoPaymentMethodId =
  (typeof DEMO_PAYMENT_METHODS)[number]['id']

/** Preset recharge amounts (¥) */
export const DEMO_PRESET_AMOUNTS = [10, 50, 100, 500] as const

/** Demo order status — extends real TopupStatus with 'refunded' and 'failed' */
export type DemoOrderStatus =
  | 'pending' // 待支付
  | 'success' // 支付成功
  | 'failed' // 支付失败
  | 'refunded' // 已退款

export interface DemoOrder {
  id: number
  trade_no: string
  amount: number // money paid (¥)
  quota: number // quota credited
  payment_method: DemoPaymentMethodId
  status: DemoOrderStatus
  create_time: number // unix ms
  complete_time?: number // unix ms
}

/** Bank transfer details shown for 对公转账 */
export const DEMO_BANK_INFO = {
  bank: '招商银行 · 北京分行',
  accountName: '北京量子智能科技有限公司',
  accountNumber: '6225 8801 2345 6789',
  branch: '中关村支行',
  remark: '请在转账备注中填写订单号',
}

/**
 * Build a set of sample orders with varied statuses and relative timestamps
 * so the orders table always has demo content.
 */
function buildSampleOrders(): DemoOrder[] {
  const now = Date.now()
  const minutes = (m: number) => now - m * 60 * 1000
  const hours = (h: number) => now - h * 60 * 60 * 1000
  const days = (d: number) => now - d * 24 * 60 * 60 * 1000

  return [
    {
      id: 20240601001,
      trade_no: 'ORD-20240601-001',
      amount: 100,
      quota: 50000000,
      payment_method: 'alipay',
      status: 'success',
      create_time: minutes(8),
      complete_time: minutes(7),
    },
    {
      id: 20240601002,
      trade_no: 'ORD-20240601-002',
      amount: 500,
      quota: 250000000,
      payment_method: 'wxpay',
      status: 'success',
      create_time: hours(3),
      complete_time: hours(3),
    },
    {
      id: 20240601003,
      trade_no: 'ORD-20240601-003',
      amount: 200,
      quota: 100000000,
      payment_method: 'bank_transfer',
      status: 'pending',
      create_time: hours(26),
    },
    {
      id: 20240601004,
      trade_no: 'ORD-20240601-004',
      amount: 50,
      quota: 25000000,
      payment_method: 'alipay',
      status: 'failed',
      create_time: days(2),
    },
    {
      id: 20240601005,
      trade_no: 'ORD-20240601-005',
      amount: 1000,
      quota: 500000000,
      payment_method: 'wxpay',
      status: 'success',
      create_time: days(5),
      complete_time: days(5),
    },
    {
      id: 20240601006,
      trade_no: 'ORD-20240601-006',
      amount: 200,
      quota: 100000000,
      payment_method: 'alipay',
      status: 'refunded',
      create_time: days(12),
      complete_time: days(12),
    },
    {
      id: 20240601007,
      trade_no: 'ORD-20240601-007',
      amount: 500,
      quota: 250000000,
      payment_method: 'bank_transfer',
      status: 'success',
      create_time: days(20),
      complete_time: days(19),
    },
  ]
}

/** Sample orders for the orders table demo. */
export const sampleOrders: DemoOrder[] = buildSampleOrders()

/** Status → i18n label + badge classes mapping */
export const ORDER_STATUS_META: Record<
  DemoOrderStatus,
  { label: string; badgeClass: string }
> = {
  pending: {
    label: '待支付',
    badgeClass: 'bg-warning/15 text-warning',
  },
  success: {
    label: '支付成功',
    badgeClass: 'bg-success/15 text-success',
  },
  failed: {
    label: '支付失败',
    badgeClass: 'bg-destructive/15 text-destructive',
  },
  refunded: {
    label: '已退款',
    badgeClass: 'bg-secondary text-secondary-foreground',
  },
}

/** Payment method id → display name */
export function getPaymentMethodName(
  id: DemoPaymentMethodId | string
): string {
  return (
    DEMO_PAYMENT_METHODS.find((m) => m.id === id)?.name ?? id
  )
}
