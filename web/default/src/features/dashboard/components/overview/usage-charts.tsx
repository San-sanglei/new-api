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
import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { SafeVChart } from '@/lib/safe-vchart'
import { PanelWrapper } from '../ui/panel-wrapper'
import { getUserQuotaDates } from '../../api'
import { VCHART_OPTION } from '@/lib/vchart'
import { useTheme } from '@/context/theme-provider'

const SEVEN_DAYS = 7 * 24 * 60 * 60 * 1000

function buildMockTrendData() {
  const now = Date.now()
  const models = ['GPT-4o', 'Claude 3.5', 'DeepSeek V3', 'Gemini 1.5']
  const days = Array.from({ length: 7 }, (_, i) => {
    const d = new Date(now - (6 - i) * 24 * 60 * 60 * 1000)
    return d.toISOString().slice(0, 10)
  })
  return days.map((day) =>
    models.map((m, mi) => ({
      date: day,
      model: m,
      calls: Math.floor(Math.random() * 8000 + 500 - mi * 1500 + (mi === 0 ? 2000 : 0)),
    }))
  ).flat()
}

function buildPieSpec(data: { type: string; value: number }[] | null | undefined) {
  const values = data ?? []
  return {
    type: 'pie',
    data: [{ values }],
    outerRadius: 0.7,
    innerRadius: 0.4,
    valueField: 'value',
    categoryField: 'type',
    title: { visible: false },
    label: {
      visible: true,
      formatMethod: (_v: string, datum: { value: number; type: string }) =>
        `${datum.type}: ${datum.value}%`,
    },
    legends: { visible: true, orient: 'bottom' },
  }
}

function buildAreaSpec(data: { date: string; model: string; calls: number }[] | null | undefined) {
  const values = data ?? []
  return {
    type: 'area',
    data: [{ values }],
    xField: 'date',
    yField: 'calls',
    seriesField: 'model',
    stack: false,
    smooth: true,
    line: { style: { lineWidth: 2 } },
    point: { visible: false },
    legends: { visible: true, orient: 'bottom' },
    title: { visible: false },
    axes: [
      { orient: 'left', title: { visible: false }, grid: { visible: true, style: { stroke: 'var(--border)' } } },
      { orient: 'bottom', title: { visible: false }, label: { style: { fontSize: 10 } } },
    ],
  }
}

export function UsageCharts() {
  const { t } = useTranslation()
  const { theme } = useTheme()

  const now = Date.now()

  const quotaQuery = useQuery({
    queryKey: ['dashboard', 'quota-7day'],
    queryFn: async () => {
      const res = await getUserQuotaDates({
        start_timestamp: Math.floor((now - SEVEN_DAYS) / 1000),
        end_timestamp: Math.floor(now / 1000),
      })
      return res
    },
    staleTime: 2 * 60 * 1000,
  })

  const trendData = useMemo(() => {
    if (quotaQuery.data?.success && (quotaQuery.data?.data?.length ?? 0) > 0) {
      const items = quotaQuery.data!.data!
      const grouped: Record<string, Record<string, number>> = {}
      for (const item of items) {
        if (!item.created_at) continue
        const day = new Date(item.created_at * 1000).toISOString().slice(0, 10)
        const model = item.model_name || 'unknown'
        if (!grouped[day]) grouped[day] = {}
        grouped[day][model] = (grouped[day][model] || 0) + (item.count || 0)
      }
      return Object.entries(grouped).flatMap(([date, models]) =>
        Object.entries(models).map(([model, calls]) => ({
          date,
          model,
          calls,
        }))
      )
    }
    return buildMockTrendData()
  }, [quotaQuery.data])

  const pieData = useMemo(() => {
    const grouped: Record<string, number> = {}
    for (const d of trendData) {
      grouped[d.model] = (grouped[d.model] || 0) + d.calls
    }
    const total = Object.values(grouped).reduce((a, b) => a + b, 0)
    if (total === 0) return [{ type: 'GPT-4o', value: 45 }, { type: 'Claude 3.5', value: 25 }, { type: 'DeepSeek V3', value: 18 }, { type: 'Gemini 1.5', value: 12 }]
    return Object.entries(grouped)
      .map(([type, value]) => ({ type, value: Math.round((value / total) * 100) }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 5)
  }, [trendData])

  const hasChartData = trendData.length > 0
  const loading = quotaQuery.isLoading && !hasChartData
  const empty = !loading && !hasChartData

  const areaSpec = useMemo(() => buildAreaSpec(trendData), [trendData])
  const pieSpec = useMemo(() => buildPieSpec(pieData), [pieData])

  return (
    <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
      <PanelWrapper
        title={t('API Calls Trend (7 Days)')}
        loading={loading}
        empty={empty}
        emptyMessage={t('No usage data available')}
        height='h-72'
      >
        {areaSpec && (
          <SafeVChart
            spec={areaSpec}
            option={VCHART_OPTION}
            theme={theme === 'dark' ? 'dark' : 'light'}
            style={{ height: 280, width: '100%' }}
          />
        )}
      </PanelWrapper>
      <PanelWrapper
        title={t('Model Usage Distribution')}
        loading={loading}
        empty={empty}
        emptyMessage={t('No usage data available')}
        height='h-72'
      >
        {pieSpec && (
          <SafeVChart
            spec={pieSpec}
            option={VCHART_OPTION}
            theme={theme === 'dark' ? 'dark' : 'light'}
            style={{ height: 280, width: '100%' }}
          />
        )}
      </PanelWrapper>
    </div>
  )
}
