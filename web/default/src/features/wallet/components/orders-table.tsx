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
import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { ArrowLeft, Receipt, RefreshCw, Eye } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import {
  sampleOrders,
  ORDER_STATUS_META,
  getPaymentMethodName,
  type DemoOrderStatus,
} from '@/features/wallet/lib/mock-data'

/** Format a unix-ms timestamp as a localized date-time string */
function formatTime(ms: number): string {
  const d = new Date(ms)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function OrdersTable() {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<DemoOrderStatus | 'all'>(
    'all'
  )
  // Bump to force re-render after new orders are added by the recharge flow
  const [refreshKey, setRefreshKey] = useState(0)

  const filtered = useMemo(() => {
    return sampleOrders.filter((o) => {
      if (statusFilter !== 'all' && o.status !== statusFilter) return false
      if (keyword) {
        const kw = keyword.toLowerCase()
        return (
          o.trade_no.toLowerCase().includes(kw) ||
          String(o.id).includes(kw) ||
          getPaymentMethodName(o.payment_method).includes(keyword)
        )
      }
      return true
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [keyword, statusFilter, refreshKey])

  const statusOptions: { value: DemoOrderStatus | 'all'; label: string }[] = [
    { value: 'all', label: t('All') },
    { value: 'pending', label: t('Pending') },
    { value: 'success', label: t('Success') },
    { value: 'failed', label: t('Failed') },
    { value: 'refunded', label: t('Refunded') },
  ]

  return (
    <div className='mx-auto w-full max-w-5xl space-y-6 p-4 md:p-6'>
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
            <Receipt className='size-5' />
            {t('Recharge History')}
          </h1>
          <p className='text-muted-foreground text-sm'>
            {t('View all recharge orders')}
          </p>
        </div>
        <Button
          variant='outline'
          size='sm'
          className='ml-auto'
          onClick={() => setRefreshKey((k) => k + 1)}
        >
          <RefreshCw className='mr-1.5 size-3.5' />
          {t('Refresh')}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t('Order List')}</CardTitle>
          <CardDescription>
            {t('{{count}} records in total', { count: filtered.length })}
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          {/* Filters */}
          <div className='flex flex-col gap-3 sm:flex-row sm:items-center'>
            <Input
              placeholder={t('Search by order no. / payment method...')}
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className='sm:max-w-xs'
            />
            <div className='flex flex-wrap gap-1.5'>
              {statusOptions.map((opt) => (
                <button
                  key={opt.value}
                  type='button'
                  onClick={() => setStatusFilter(opt.value)}
                  className={cn(
                    'rounded-md border px-3 py-1 text-xs font-medium transition-colors',
                    statusFilter === opt.value
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'border-border text-muted-foreground hover:bg-accent'
                  )}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>

          {/* Table — desktop */}
          <div className='hidden md:block'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('Time')}</TableHead>
                  <TableHead>{t('Order No.')}</TableHead>
                  <TableHead className='text-right'>{t('Amount')}</TableHead>
                  <TableHead>{t('Payment Method')}</TableHead>
                  <TableHead>{t('Status')}</TableHead>
                  <TableHead className='text-right'>{t('Actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={6}
                      className='text-muted-foreground py-8 text-center'
                    >
                      {t('No records')}
                    </TableCell>
                  </TableRow>
                ) : (
                  filtered.map((order) => {
                    const meta = ORDER_STATUS_META[order.status]
                    return (
                      <TableRow key={order.id}>
                        <TableCell className='font-mono text-xs'>
                          {formatTime(order.create_time)}
                        </TableCell>
                        <TableCell className='font-mono text-xs'>
                          {order.trade_no}
                        </TableCell>
                        <TableCell className='text-right font-semibold'>
                          ¥{order.amount.toFixed(2)}
                        </TableCell>
                        <TableCell>
                          {getPaymentMethodName(order.payment_method)}
                        </TableCell>
                        <TableCell>
                          <Badge
                            variant='outline'
                            className={cn('border-0', meta.badgeClass)}
                          >
                            {meta.label}
                          </Badge>
                        </TableCell>
                        <TableCell className='text-right'>
                          <Button
                            variant='ghost'
                            size='sm'
                            className='h-7'
                          >
                            <Eye className='mr-1 size-3' />
                            {t('Details')}
                          </Button>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          </div>

          {/* Card list — mobile */}
          <div className='space-y-3 md:hidden'>
            {filtered.length === 0 ? (
              <div className='text-muted-foreground py-8 text-center text-sm'>
                {t('No records')}
              </div>
            ) : (
              filtered.map((order) => {
                const meta = ORDER_STATUS_META[order.status]
                return (
                  <div
                    key={order.id}
                    className='rounded-lg border p-3'
                  >
                    <div className='flex items-center justify-between'>
                      <span className='font-mono text-xs text-muted-foreground'>
                        {order.trade_no}
                      </span>
                      <Badge
                        variant='outline'
                        className={cn('border-0', meta.badgeClass)}
                      >
                        {meta.label}
                      </Badge>
                    </div>
                    <div className='mt-2 flex items-center justify-between'>
                      <span className='text-sm'>
                        {getPaymentMethodName(order.payment_method)}
                      </span>
                      <span className='font-semibold'>¥{order.amount.toFixed(2)}</span>
                    </div>
                    <div className='text-muted-foreground mt-1 text-xs'>
                      {formatTime(order.create_time)}
                    </div>
                  </div>
                )
              })
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
