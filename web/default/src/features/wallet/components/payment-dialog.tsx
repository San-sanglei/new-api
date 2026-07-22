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
import { useEffect, useState } from 'react'
import { CheckCircle2, Loader2, XCircle, Copy, Building2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  DEMO_BANK_INFO,
  getPaymentMethodName,
} from '@/features/wallet/lib/mock-data'
import type { useMockPayment } from '@/features/wallet/hooks/use-mock-payment'

interface PaymentDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  payment: ReturnType<typeof useMockPayment>
}

/**
 * Payment dialog — shows a QR code for 支付宝/微信支付, or bank transfer
 * details for 对公转账. Polls payment status and shows a success/failure
 * result screen.
 */
export function PaymentDialog({
  open,
  onOpenChange,
  payment,
}: PaymentDialogProps) {
  const { t } = useTranslation()
  const { currentOrder, status, qrPayload, confirmBankTransfer, cancelPayment } =
    payment
  const [elapsed, setElapsed] = useState(0)

  // Count up elapsed seconds while pending (for QR methods)
  useEffect(() => {
    if (!open || status !== 'pending' || !qrPayload) return
    setElapsed(0)
    const id = setInterval(() => setElapsed((e) => e + 1), 1000)
    return () => clearInterval(id)
  }, [open, status, qrPayload])

  if (!currentOrder) return null

  const isQrMethod =
    currentOrder.payment_method === 'alipay' ||
    currentOrder.payment_method === 'wxpay'
  const isBankTransfer = currentOrder.payment_method === 'bank_transfer'
  const methodName = getPaymentMethodName(currentOrder.payment_method)

  const copyBankInfo = (text: string, label: string) => {
    navigator.clipboard?.writeText(text).then(() => {
      toast.success(t('{{label}} copied', { label }))
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>
            {status === 'success'
              ? t('Payment Successful')
              : status === 'failed'
                ? t('Payment Failed')
                : t('{{method}} Payment', { method: methodName })}
          </DialogTitle>
          <DialogDescription>
            {t('Order No.')}: {currentOrder.trade_no} · ¥{currentOrder.amount.toFixed(2)}
          </DialogDescription>
        </DialogHeader>

        {/* Success state */}
        {status === 'success' ? (
          <div className='flex flex-col items-center gap-3 py-6'>
            <CheckCircle2 className='text-success size-16' />
            <p className='text-lg font-semibold'>{t('Payment Successful')}</p>
            <p className='text-muted-foreground text-sm'>
              {t('Quota credited. Thank you for your support.')}
            </p>
            <Button onClick={() => onOpenChange(false)} className='mt-2'>
              {t('Done')}
            </Button>
          </div>
        ) : status === 'failed' ? (
          <div className='flex flex-col items-center gap-3 py-6'>
            <XCircle className='text-destructive size-16' />
            <p className='text-lg font-semibold'>{t('Payment Failed')}</p>
            <p className='text-muted-foreground text-sm'>
              {t('Order cancelled. You can initiate a new payment.')}
            </p>
            <Button onClick={() => onOpenChange(false)} className='mt-2'>
              {t('Close')}
            </Button>
          </div>
        ) : isQrMethod ? (
          /* QR code payment (alipay / wxpay) */
          <div className='flex flex-col items-center gap-4 py-4'>
            <div className='bg-white rounded-lg border-2 p-4'>
              {/* Fake QR code — a grid of squares. In production this would be
                  a real QR image from the payment provider. */}
              <FakeQrCode payload={qrPayload} />
            </div>
            <div className='text-center'>
              <p className='text-sm font-medium'>
                {t('Please use')} {methodName} {t('to scan the QR code and pay')}
              </p>
              <p className='text-muted-foreground mt-1 flex items-center justify-center gap-1.5 text-xs'>
                <Loader2 className='size-3 animate-spin' />
                {t('Waiting for payment result')} · {elapsed}s
              </p>
            </div>
            <Button variant='outline' size='sm' onClick={cancelPayment}>
              {t('Cancel Payment')}
            </Button>
          </div>
        ) : isBankTransfer ? (
          /* Bank transfer details */
          <div className='space-y-4 py-2'>
            <div className='bg-muted/50 flex items-center gap-3 rounded-lg p-3'>
              <Building2 className='text-muted-foreground size-5 shrink-0' />
              <p className='text-sm'>
                {t('Please transfer to the following account, then click "I have transferred"')}
              </p>
            </div>
            <Separator />
            <div className='space-y-3'>
              {[
                { label: t('Bank'), value: DEMO_BANK_INFO.bank },
                { label: t('Account Name'), value: DEMO_BANK_INFO.accountName },
                {
                  label: t('Account Number'),
                  value: DEMO_BANK_INFO.accountNumber,
                },
                { label: t('Branch'), value: DEMO_BANK_INFO.branch },
              ].map((row) => (
                <div
                  key={row.label}
                  className='flex items-center justify-between gap-3'
                >
                  <span className='text-muted-foreground shrink-0 text-sm'>
                    {row.label}
                  </span>
                  <span className='text-right font-mono text-sm font-medium'>
                    {row.value}
                  </span>
                  <Button
                    variant='ghost'
                    size='icon'
                    className='size-7 shrink-0'
                    onClick={() => copyBankInfo(row.value, row.label)}
                  >
                    <Copy className='size-3.5' />
                  </Button>
                </div>
              ))}
              <Separator />
              <div className='bg-warning/10 text-warning rounded-md p-2 text-xs'>
                {DEMO_BANK_INFO.remark}：{currentOrder.trade_no}
              </div>
            </div>
            <div className='flex gap-2 pt-2'>
              <Button
                variant='outline'
                className='flex-1'
                onClick={cancelPayment}
              >
                {t('Cancel')}
              </Button>
              <Button className='flex-1' onClick={confirmBankTransfer}>
                {t('I have transferred')}
              </Button>
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

/**
 * Deterministic fake QR code rendered as a CSS grid. The pattern is derived
 * from the payload hash so each order has a distinct-looking code. This is
 * for demo only — not a scannable QR code.
 */
function FakeQrCode({ payload }: { payload: string | null }) {
  if (!payload) return null
  const size = 21 // standard QR module count
  // Simple hash → boolean grid
  let h = 0
  for (let i = 0; i < payload.length; i++) {
    h = (h << 5) - h + payload.charCodeAt(i)
    h |= 0
  }
  const cells: boolean[] = []
  let x = h
  for (let i = 0; i < size * size; i++) {
    x = (x * 1103515245 + 12345) & 0x7fffffff
    cells.push((x >> 16) % 2 === 0)
  }
  // Force three position-detection corners (the classic QR squares)
  const setBlock = (r: number, c: number) => {
    for (let dr = 0; dr < 7; dr++)
      for (let dc = 0; dc < 7; dc++) {
        const border = dr === 0 || dr === 6 || dc === 0 || dc === 6
        const inner = dr >= 2 && dr <= 4 && dc >= 2 && dc <= 4
        cells[(r + dr) * size + (c + dc)] = border || inner
      }
  }
  setBlock(0, 0)
  setBlock(0, size - 7)
  setBlock(size - 7, 0)

  return (
    <div
      className='grid'
      style={{
        gridTemplateColumns: `repeat(${size}, 8px)`,
        gridTemplateRows: `repeat(${size}, 8px)`,
      }}
    >
      {cells.map((on, i) => (
        <div
          key={i}
          className={on ? 'bg-black' : 'bg-white'}
          style={{ width: 8, height: 8 }}
        />
      ))}
    </div>
  )
}
