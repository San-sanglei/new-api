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
import { motion } from 'motion/react'
import {
  Wallet,
  Zap,
  Activity,
  KeyRound,
  TrendingUp,
  TrendingDown,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { formatQuota } from '@/lib/format'
import { StatCard } from '../ui/stat-card'
import { getApiKeys } from '@/features/keys/api'
import { getUserLogStats } from '@/features/usage-logs/api'

function TrendBadge({ value }: { value: number }) {
  const isUp = value >= 0
  const Icon = isUp ? TrendingUp : TrendingDown
  return (
    <span className='inline-flex items-center gap-0.5 text-xs font-medium text-emerald-500'>
      <Icon className='size-3' />
      {isUp ? '+' : ''}{value}%
    </span>
  )
}

export function TopStatsCards() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)

  const remainQuota = Number(user?.quota ?? 0)
  const usedQuota = Number(user?.used_quota ?? 0)
  const requestCount = Number(user?.request_count ?? 0)

  const apiKeysQuery = useQuery({
    queryKey: ['dashboard', 'api-keys'],
    queryFn: async () => {
      const res = await getApiKeys({ p: 1, size: 100 })
      return res.success ? (res.data?.items ?? []) : []
    },
    staleTime: 60 * 1000,
  })

  const todayStatsQuery = useQuery({
    queryKey: ['dashboard', 'today-stats'],
    queryFn: async () => {
      const now = Date.now()
      const startOfDay = new Date()
      startOfDay.setHours(0, 0, 0, 0)
      const res = await getUserLogStats({
        start_timestamp: Math.floor(startOfDay.getTime() / 1000),
        end_timestamp: Math.floor(now / 1000),
      })
      return res
    },
    staleTime: 30 * 1000,
  })

  const activeKeyCount =
    apiKeysQuery.data?.filter((k: { status: number }) => k.status === 1)
      .length ?? 0

  const cards = [
    {
      title: t('Current Balance'),
      value: formatQuota(remainQuota),
      description: `${t('Used')}: ${formatQuota(usedQuota)}`,
      icon: Wallet,
      tone: 'teal' as const,
      loading: false,
      error: false,
      children: <TrendBadge value={12} />,
    },
    {
      title: t('Today Usage'),
      value: formatQuota(todayStatsQuery.data?.data?.quota ?? 0),
      description: `${t('Requests')}: ${todayStatsQuery.data?.data?.rpm ?? 0}`,
      icon: Zap,
      tone: 'rose' as const,
      loading: todayStatsQuery.isLoading,
      error: todayStatsQuery.isError,
    },
    {
      title: t('Monthly Calls'),
      value: requestCount.toLocaleString(),
      description: t('Total requests via API'),
      icon: Activity,
      tone: 'gray' as const,
      loading: false,
      error: false,
    },
    {
      title: t('Active API Keys'),
      value: String(activeKeyCount),
      description: t('Keys enabled and operational'),
      icon: KeyRound,
      tone: 'gray' as const,
      loading: apiKeysQuery.isLoading,
      error: apiKeysQuery.isError,
    },
  ]

  return (
    <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4'>
      {cards.map((card, i) => (
        <motion.div
          key={card.title}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: i * 0.08 }}
          className='bg-card rounded-2xl border p-4 shadow-xs'
        >
          <StatCard
            title={card.title}
            value={card.value}
            description={card.description}
            icon={card.icon}
            tone={card.tone}
            loading={card.loading}
            error={card.error}
          />
        </motion.div>
      ))}
    </div>
  )
}
