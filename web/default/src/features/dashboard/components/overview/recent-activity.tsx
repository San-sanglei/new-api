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
import { getUserLogs } from '@/features/usage-logs/api'
import { PanelWrapper } from '../ui/panel-wrapper'
import { cn } from '@/lib/utils'

interface LogEntry {
  id: number
  created_at: number
  model_name?: string
  status: boolean | number
  token_used?: number
  quota?: number
  content?: string
}

export function RecentActivity() {
  const { t } = useTranslation()

  const logsQuery = useQuery({
    queryKey: ['dashboard', 'recent-logs'],
    queryFn: async () => {
      const res = await getUserLogs({ p: 1, page_size: 10 })
      return ((res?.data?.items ?? []) as unknown) as LogEntry[]
    },
    staleTime: 30 * 1000,
  })

  const logs = logsQuery.data ?? []
  const loading = logsQuery.isLoading

  return (
    <PanelWrapper
      title={t('Recent Activity')}
      description={t('Last 10 API calls')}
      loading={loading}
      empty={!loading && logs.length === 0}
      emptyMessage={t('No recent activity')}
    >
      <div className='-mx-4 -mb-4 sm:-mx-5'>
        <table className='w-full text-sm'>
          <thead>
            <tr className='border-b border-border/40'>
              <th className='px-4 py-2.5 text-left text-xs font-medium text-muted-foreground'>{t('Time')}</th>
              <th className='px-4 py-2.5 text-left text-xs font-medium text-muted-foreground'>{t('Model')}</th>
              <th className='px-4 py-2.5 text-left text-xs font-medium text-muted-foreground'>{t('Status')}</th>
              <th className='px-4 py-2.5 text-right text-xs font-medium text-muted-foreground'>{t('Tokens')}</th>
              <th className='px-4 py-2.5 text-right text-xs font-medium text-muted-foreground'>{t('Cost')}</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log, i) => (
              <motion.tr
                key={log.id ?? i}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.03 }}
                className='border-b border-border/20 transition-colors last:border-0 hover:bg-muted/10'
              >
                <td className='px-4 py-2.5 text-xs text-muted-foreground tabular-nums'>
                  {log.created_at
                    ? new Date(log.created_at * 1000).toLocaleString('zh-CN', {
                        month: '2-digit',
                        day: '2-digit',
                        hour: '2-digit',
                        minute: '2-digit',
                      })
                    : '-'}
                </td>
                <td className='px-4 py-2.5 text-xs font-medium'>
                  {log.model_name ?? '-'}
                </td>
                <td className='px-4 py-2.5'>
                  <span
                    className={cn(
                      'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
                      log.status
                        ? 'bg-emerald-500/10 text-emerald-500'
                        : 'bg-red-500/10 text-red-500'
                    )}
                  >
                    {log.status ? t('Success') : t('Error')}
                  </span>
                </td>
                <td className='px-4 py-2.5 text-right text-xs tabular-nums text-muted-foreground'>
                  {log.token_used != null ? log.token_used.toLocaleString() : '-'}
                </td>
                <td className='px-4 py-2.5 text-right text-xs tabular-nums text-muted-foreground'>
                  {log.quota != null ? log.quota.toLocaleString() : '-'}
                </td>
              </motion.tr>
            ))}
          </tbody>
        </table>
      </div>
    </PanelWrapper>
  )
}
