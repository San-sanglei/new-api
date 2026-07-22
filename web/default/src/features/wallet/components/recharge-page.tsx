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
import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { ArrowLeft, Wallet, Zap } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import {
  DEMO_PAYMENT_METHODS,
  DEMO_PRESET_AMOUNTS,
  type DemoPaymentMethodId,
} from '@/features/wallet/lib/mock-data'
import { useMockPayment } from '@/features/wallet/hooks/use-mock-payment'
import { PaymentDialog } from './payment-dialog'

export function RechargePage() {
  const { t } = useTranslation()
  const [amount, setAmount] = useState<number>(100)
  const [customAmount, setCustomAmount] = useState<string>('')
  const [isCustom, setIsCustom] = useState(false)
  const [method, setMethod] = useState<DemoPaymentMethodId>('alipay')
  const [dialogOpen, setDialogOpen] = useState(false)

  const payment = useMockPayment({
    onSettled: (status) => {
      if (status === 'success') {
        toast.success(t('Payment Successful'))
      } else {
        toast.error(t('Payment Failed'))
      }
    },
  })

  const effectiveAmount = isCustom
    ? Number(customAmount) || 0
    : amount

  const canPay = effectiveAmount > 0

  const handlePay = () => {
    if (!canPay) {
      toast.error(t('Please enter a valid amount'))
      return
    }
    payment.startPayment(effectiveAmount, method)
    setDialogOpen(true)
  }

  return (
    <div className='mx-auto w-full max-w-3xl space-y-6 p-4 md:p-6'>
      {/* Header */}
      <div className='flex items-center gap-3'>
        <Button
          variant='ghost'
          size='icon'
          render={<Link to='/' />}
          aria-label={t('Back')}
        >
          <ArrowLeft className='size-4' />
        </Button>
        <div>
          <h1 className='flex items-center gap-2 text-xl font-semibold'>
            <Wallet className='size-5' />
            {t('Account Recharge')}
          </h1>
          <p className='text-muted-foreground text-sm'>
            {t('Select recharge amount and payment method')}
          </p>
        </div>
      </div>

      {/* Amount selection */}
      <Card>
        <CardHeader>
          <CardTitle>{t('Recharge Amount')}</CardTitle>
          <CardDescription>{t('Choose a preset amount or enter a custom amount')}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5'>
            {DEMO_PRESET_AMOUNTS.map((preset) => (
              <button
                key={preset}
                type='button'
                onClick={() => {
                  setAmount(preset)
                  setIsCustom(false)
                }}
                className={cn(
                  'rounded-lg border-2 p-4 text-center transition-all',
                  !isCustom && amount === preset
                    ? 'border-primary bg-primary/10 text-primary font-semibold'
                    : 'border-border hover:border-primary/50 hover:bg-accent'
                )}
              >
                <div className='text-lg font-bold'>¥{preset}</div>
                <div className='text-muted-foreground mt-0.5 text-xs'>
                  {preset * 50} {t('K quota')}
                </div>
              </button>
            ))}
          </div>

          {/* Custom amount */}
          <div className='space-y-2'>
            <label className='text-sm font-medium'>
              {t('Custom Amount')}
            </label>
            <div className='flex items-center gap-2'>
              <span className='text-muted-foreground'>¥</span>
              <Input
                type='number'
                min={1}
                placeholder={t('Enter amount')}
                value={isCustom ? customAmount : ''}
                onChange={(e) => {
                  setCustomAmount(e.target.value)
                  setIsCustom(true)
                }}
                onFocus={() => setIsCustom(true)}
                className='max-w-xs'
              />
            </div>
          </div>

          {/* Summary */}
          <div className='bg-muted/50 flex items-center justify-between rounded-lg p-3'>
            <span className='text-muted-foreground text-sm'>
              {t('Amount Due')}
            </span>
            <span className='text-primary text-2xl font-bold'>
              ¥{effectiveAmount.toFixed(2)}
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Payment method */}
      <Card>
        <CardHeader>
          <CardTitle>{t('Payment Method')}</CardTitle>
          <CardDescription>{t('Choose your payment method')}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className='grid gap-3 sm:grid-cols-3'>
            {DEMO_PAYMENT_METHODS.map((m) => (
              <button
                key={m.id}
                type='button'
                onClick={() => setMethod(m.id)}
                className={cn(
                  'flex items-center gap-3 rounded-lg border-2 p-4 transition-all',
                  method === m.id
                    ? 'border-primary bg-primary/10'
                    : 'border-border hover:border-primary/50 hover:bg-accent'
                )}
              >
                <span
                  className='inline-flex size-8 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white'
                  style={{ backgroundColor: m.color }}
                >
                  {m.name.slice(0, 1)}
                </span>
                <span className='font-medium'>{m.name}</span>
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Pay button */}
      <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
        <div className='text-muted-foreground text-sm'>
          {t('By clicking pay, you agree to the')}
          <span className='text-foreground'> {t('Terms of Service')} </span>
        </div>
        <Button
          size='lg'
          onClick={handlePay}
          disabled={!canPay}
          className='sm:min-w-40'
        >
          <Zap className='mr-1.5 size-4' />
          {t('Pay Now')} ¥{effectiveAmount.toFixed(2)}
        </Button>
      </div>

      {/* Payment dialog */}
      <PaymentDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open)
          if (!open) payment.reset()
        }}
        payment={payment}
      />
    </div>
  )
}
