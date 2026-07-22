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
import { useCallback, useEffect, useRef, useState } from 'react'
import {
  sampleOrders,
  type DemoOrder,
  type DemoOrderStatus,
  type DemoPaymentMethodId,
} from '@/features/wallet/lib/mock-data'

interface UseMockPaymentOptions {
  /** Called when payment reaches a terminal status */
  onSettled?: (status: DemoOrderStatus, order: DemoOrder) => void
}

interface MockPaymentState {
  /** Whether a payment is in progress (dialog open) */
  paying: boolean
  /** Current order being paid (if any) */
  currentOrder: DemoOrder | null
  /** Polled status of the current order */
  status: DemoOrderStatus | null
  /** Fake QR code payload for alipay/wxpay */
  qrPayload: string | null
}

/**
 * Mock payment flow for the recharge demo.
 *
 * - `startPayment(amount, method)` creates a pending order and (for
 *   alipay/wxpay) a fake QR payload, then polls status.
 * - Polling auto-succeeds after ~6s for alipay/wxpay to simulate a scan.
 * - Bank transfer stays `pending` until manually confirmed.
 * - `confirmBankTransfer()` marks a pending bank-transfer order as success.
 * - `cancelPayment()` discards the current order (marks failed).
 *
 * Orders are appended to the shared `sampleOrders` array so the orders
 * table reflects new recharges immediately.
 */
export function useMockPayment(options: UseMockPaymentOptions = {}) {
  const { onSettled } = options
  const [state, setState] = useState<MockPaymentState>({
    paying: false,
    currentOrder: null,
    status: null,
    qrPayload: null,
  })
  const pollTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const settledRef = useRef(false)

  const clearPoll = () => {
    if (pollTimer.current) {
      clearTimeout(pollTimer.current)
      pollTimer.current = null
    }
  }

  useEffect(() => clearPoll, [])

  const startPayment = useCallback(
    (amount: number, method: DemoPaymentMethodId) => {
      clearPoll()
      settledRef.current = false

      const order: DemoOrder = {
        id: Date.now(),
        trade_no: `ORD-${new Date().toISOString().slice(0, 10)}-${Math.floor(
          Math.random() * 9000 + 1000
        )}`,
        amount,
        quota: amount * 500000,
        payment_method: method,
        status: 'pending',
        create_time: Date.now(),
      }

      // Push into the shared sample orders list (newest first)
      sampleOrders.unshift(order)

      const qrPayload =
        method === 'alipay' || method === 'wxpay'
          ? `https://demo.pay.example/${method}?order=${order.trade_no}&amount=${amount}`
          : null

      setState({
        paying: true,
        currentOrder: order,
        status: 'pending',
        qrPayload,
      })

      // Simulate scan + pay success for QR methods after a short delay
      if (method === 'alipay' || method === 'wxpay') {
        pollTimer.current = setTimeout(() => {
          order.status = 'success'
          order.complete_time = Date.now()
          setState((s) => ({ ...s, status: 'success' }))
          if (!settledRef.current) {
            settledRef.current = true
            onSettled?.('success', order)
          }
        }, 6000)
      }
      // bank_transfer stays pending — user must confirm manually
    },
    [onSettled]
  )

  const confirmBankTransfer = useCallback(() => {
    setState((s) => {
      if (!s.currentOrder) return s
      s.currentOrder.status = 'success'
      s.currentOrder.complete_time = Date.now()
      if (!settledRef.current) {
        settledRef.current = true
        onSettled?.('success', s.currentOrder)
      }
      return { ...s, status: 'success' }
    })
  }, [onSettled])

  const cancelPayment = useCallback(() => {
    clearPoll()
    setState((s) => {
      if (s.currentOrder && s.status === 'pending') {
        s.currentOrder.status = 'failed'
      }
      if (!settledRef.current && s.currentOrder) {
        settledRef.current = true
        onSettled?.('failed', s.currentOrder)
      }
      return { paying: false, currentOrder: null, status: null, qrPayload: null }
    })
  }, [onSettled])

  const reset = useCallback(() => {
    clearPoll()
    settledRef.current = false
    setState({ paying: false, currentOrder: null, status: null, qrPayload: null })
  }, [])

  return {
    ...state,
    startPayment,
    confirmBankTransfer,
    cancelPayment,
    reset,
  }
}
