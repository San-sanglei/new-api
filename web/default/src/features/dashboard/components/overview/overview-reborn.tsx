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
// dashboard usage overview (DeepSeek-style)
import { Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ArrowRight, Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { getCurrencyLabel, isCurrencyDisplayEnabled } from '@/lib/currency'
import { formatNumber, formatQuota } from '@/lib/format'
import { computeTimeRange } from '@/lib/time'
import { cn } from '@/lib/utils'
import { useStatus } from '@/hooks/use-status'
import { Button } from '@/components/ui/button'
import { getUserQuotaDates } from '../../api'
import type { QuotaDataItem } from '../../types'

type RangeKey = 'today' | '7d' | '30d'

interface RangeOption {
  key: RangeKey
  label: string
  days: number
}

const RANGE_OPTIONS: RangeOption[] = [
  { key: 'today', label: 'Today', days: 1 },
  { key: '7d', label: 'Last 7 days', days: 7 },
  { key: '30d', label: 'Last 30 days', days: 30 },
]

interface ModelRow {
  name: string
  calls: number
  tokens: number
  quota: number
}

function aggregateByModel(items: QuotaDataItem[]): ModelRow[] {
  const map = new Map<string, ModelRow>()
  for (const item of items) {
    const name = item.model_name || 'unknown'
    const row = map.get(name) ?? {
      name,
      calls: 0,
      tokens: 0,
      quota: 0,
    }
    row.calls += Number(item.count ?? 0)
    row.tokens += Number(item.token_used ?? 0)
    row.quota += Number(item.quota ?? 0)
    map.set(name, row)
  }
  return Array.from(map.values()).sort((a, b) => b.quota - a.quota)
}

export function OverviewReborn() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const { status } = useStatus()
  const [range, setRange] = useState<RangeKey>('7d')

  const option = RANGE_OPTIONS.find((o) => o.key === range)!
  const timeRange = useMemo(
    () => computeTimeRange(option.days),
    [option.days]
  )

  const remainQuota = Number(user?.quota ?? 0)
  const usedQuota = Number(user?.used_quota ?? 0)

  const currencyEnabledFromStore = isCurrencyDisplayEnabled()
  const statusCurrencyFlag =
    typeof status?.display_in_currency === 'boolean'
      ? Boolean(status.display_in_currency)
      : undefined
  const currencyEnabled =
    statusCurrencyFlag !== undefined
      ? statusCurrencyFlag
      : currencyEnabledFromStore
  const currencyLabel = currencyEnabled ? getCurrencyLabel() : 'Tokens'

  const usageQuery = useQuery({
    queryKey: [
      'dashboard',
      'usage',
      range,
      timeRange.start_timestamp,
      timeRange.end_timestamp,
    ],
    queryFn: async () =>
      getUserQuotaDates({
        start_timestamp: timeRange.start_timestamp,
        end_timestamp: timeRange.end_timestamp,
        default_time: option.days === 1 ? 'hour' : 'day',
      }),
    staleTime: 60 * 1000,
  })

  const rows = useMemo(
    () => aggregateByModel(usageQuery.data?.data ?? []),
    [usageQuery.data]
  )

  const totalQuota = rows.reduce((sum, r) => sum + r.quota, 0)
  const totalCalls = rows.reduce((sum, r) => sum + r.calls, 0)
  const totalTokens = rows.reduce((sum, r) => sum + r.tokens, 0)
  const loading = usageQuery.isLoading
  const empty = !loading && rows.length === 0

  return (
    <div className='flex flex-col gap-4'>
      {/* Header: title + range switcher */}
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div>
          <h1 className='text-xl font-semibold tracking-tight md:text-2xl'>
            {t('Usage')}
          </h1>
          <p className='text-muted-foreground mt-1 text-sm'>
            {t('Monitor your consumption and model usage')}
          </p>
        </div>
        <div className='bg-muted/30 inline-flex rounded-lg border p-0.5'>
          {RANGE_OPTIONS.map((opt) => (
            <button
              key={opt.key}
              type='button'
              onClick={() => setRange(opt.key)}
              className={cn(
                'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                range === opt.key
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {t(opt.label)}
            </button>
          ))}
        </div>
      </div>

      {/* Balance card */}
      <div className='bg-card overflow-hidden rounded-2xl border shadow-xs'>
        <div className='grid gap-px bg-border/40 sm:grid-cols-3'>
          <div className='bg-card p-5 sm:col-span-1'>
            <div className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Current Balance')}
            </div>
            <div className='mt-2 flex items-baseline gap-2'>
              <span className='font-mono text-3xl font-semibold tracking-tight'>
                {formatQuota(remainQuota)}
              </span>
              <span className='text-muted-foreground text-xs'>
                {currencyLabel}
              </span>
            </div>
            <Button
              className='mt-4 h-9 w-full rounded-lg'
              render={<Link to='/wallet' />}
            >
              <Plus className='size-4' />
              {t('Top Up')}
            </Button>
          </div>

          <div className='bg-card p-5'>
            <div className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Total Used')}
            </div>
            <div className='mt-2 flex items-baseline gap-2'>
              <span className='font-mono text-2xl font-semibold tracking-tight'>
                {formatQuota(usedQuota)}
              </span>
              <span className='text-muted-foreground text-xs'>
                {currencyLabel}
              </span>
            </div>
            <div className='text-muted-foreground mt-4 text-xs'>
              {t('Cumulative consumption')}
            </div>
          </div>

          <div className='bg-card p-5'>
            <div className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Selected Range')}
            </div>
            <div className='mt-2 flex items-baseline gap-2'>
              <span className='font-mono text-2xl font-semibold tracking-tight'>
                {formatQuota(totalQuota)}
              </span>
              <span className='text-muted-foreground text-xs'>
                {currencyLabel}
              </span>
            </div>
            <div className='text-muted-foreground mt-4 text-xs'>
              {formatNumber(totalCalls)} {t('calls')} ·{' '}
              {formatNumber(totalTokens)} {t('tokens')}
            </div>
          </div>
        </div>
      </div>

      {/* Model breakdown table */}
      <div className='bg-card overflow-hidden rounded-2xl border shadow-xs'>
        <div className='flex items-center justify-between border-b px-5 py-4'>
          <h2 className='text-sm font-semibold'>
            {t('Model Consumption Breakdown')}
          </h2>
          <Link
            to='/usage-logs'
            className='text-muted-foreground hover:text-foreground inline-flex items-center gap-1 text-xs transition-colors'
          >
            {t('View Logs')}
            <ArrowRight className='size-3' />
          </Link>
        </div>

        {loading ? (
          <div className='divide-border/60 divide-y'>
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className='flex items-center gap-4 px-5 py-3.5'
              >
                <div className='bg-muted h-4 w-32 animate-pulse rounded' />
                <div className='bg-muted ml-auto h-4 w-16 animate-pulse rounded' />
                <div className='bg-muted h-4 w-20 animate-pulse rounded' />
                <div className='bg-muted h-4 w-24 animate-pulse rounded' />
              </div>
            ))}
          </div>
        ) : empty ? (
          <div className='text-muted-foreground px-5 py-16 text-center text-sm'>
            {t('No usage data available for the selected period')}
          </div>
        ) : (
          <div className='overflow-x-auto'>
            <table className='w-full text-sm'>
              <thead>
                <tr className='text-muted-foreground border-b text-xs tracking-wider uppercase'>
                  <th className='px-5 py-3 text-left font-medium'>
                    {t('Model')}
                  </th>
                  <th className='px-5 py-3 text-right font-medium'>
                    {t('Calls')}
                  </th>
                  <th className='px-5 py-3 text-right font-medium'>
                    {t('Tokens')}
                  </th>
                  <th className='px-5 py-3 text-right font-medium'>
                    {t('Consumption')}
                  </th>
                  <th className='px-5 py-3 text-left font-medium'>
                    {t('Share')}
                  </th>
                </tr>
              </thead>
              <tbody className='divide-border/60 divide-y'>
                {rows.map((row) => {
                  const share =
                    totalQuota > 0 ? (row.quota / totalQuota) * 100 : 0
                  return (
                    <tr
                      key={row.name}
                      className='hover:bg-muted/30 transition-colors'
                    >
                      <td className='px-5 py-3.5'>
                        <span className='font-medium'>{row.name}</span>
                      </td>
                      <td className='text-muted-foreground px-5 py-3.5 text-right tabular-nums'>
                        {formatNumber(row.calls)}
                      </td>
                      <td className='text-muted-foreground px-5 py-3.5 text-right tabular-nums'>
                        {formatNumber(row.tokens)}
                      </td>
                      <td className='px-5 py-3.5 text-right font-medium tabular-nums'>
                        {formatQuota(row.quota)}
                      </td>
                      <td className='px-5 py-3.5'>
                        <div className='flex items-center gap-2'>
                          <div className='bg-muted h-1.5 w-full max-w-[120px] overflow-hidden rounded-full'>
                            <div
                              className='bg-foreground h-full rounded-full transition-all'
                              style={{ width: `${Math.max(share, 2)}%` }}
                            />
                          </div>
                          <span className='text-muted-foreground w-12 text-right text-xs tabular-nums'>
                            {share.toFixed(1)}%
                          </span>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
              <tfoot>
                <tr className='bg-muted/20 border-t font-medium'>
                  <td className='px-5 py-3.5'>{t('Total')}</td>
                  <td className='px-5 py-3.5 text-right tabular-nums'>
                    {formatNumber(totalCalls)}
                  </td>
                  <td className='px-5 py-3.5 text-right tabular-nums'>
                    {formatNumber(totalTokens)}
                  </td>
                  <td className='px-5 py-3.5 text-right tabular-nums'>
                    {formatQuota(totalQuota)}
                  </td>
                  <td className='px-5 py-3.5' />
                </tr>
              </tfoot>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
