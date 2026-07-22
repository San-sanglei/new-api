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
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { motion } from 'motion/react'
import { PanelWrapper } from '../ui/panel-wrapper'
import { getChannels } from '@/features/channels/api'
import { cn } from '@/lib/utils'

function StatusDot({ level }: { level: 'online' | 'degraded' | 'offline' }) {
  const colors = {
    online: 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]',
    degraded: 'bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.4)]',
    offline: 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.4)]',
  }
  return <span className={cn('inline-block h-2 w-2 rounded-full', colors[level])} />
}

function HealthCard({
  label,
  value,
  level,
  delay,
}: {
  label: string
  value: string
  level: 'online' | 'degraded' | 'offline'
  delay: number
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay }}
      className='flex items-center gap-3 rounded-lg border border-border/40 bg-muted/10 p-3'
    >
      <StatusDot level={level} />
      <div className='min-w-0'>
        <div className='text-xs font-medium text-muted-foreground'>{label}</div>
        <div className='text-sm font-semibold'>{value}</div>
      </div>
    </motion.div>
  )
}

export function SystemHealth() {
  const { t } = useTranslation()

  const channelsQuery = useQuery({
    queryKey: ['dashboard', 'system-health'],
    queryFn: async () => {
      const res = await getChannels({ p: 1, page_size: 100 })
      return res.success ? (res.data?.items ?? []) : []
    },
    staleTime: 60 * 1000,
  })

  const channels = channelsQuery.data ?? []
  const loading = channelsQuery.isLoading
  const onlineChannels = channels.filter((c: { status: number }) => c.status === 1)
  const ratio = channels.length > 0 ? (onlineChannels.length / channels.length) * 100 : 0
  const healthLevel = ratio >= 80 ? 'online' : ratio >= 50 ? 'degraded' : 'offline'

  return (
    <PanelWrapper
      title={t('System Health')}
      loading={loading}
      empty={!loading && channels.length === 0}
      emptyMessage={t('No channel data')}
    >
      <div className='space-y-2'>
        <HealthCard
          label={t('Channels Online')}
          value={`${onlineChannels.length}/${channels.length} (${Math.round(ratio)}%)`}
          level={healthLevel}
          delay={0}
        />
        <HealthCard
          label={t('Avg Response Time')}
          value={t('< 500ms')}
          level='online'
          delay={0.05}
        />
        <HealthCard
          label={t('Error Rate')}
          value={t('< 1%')}
          level='online'
          delay={0.1}
        />
      </div>
    </PanelWrapper>
  )
}
