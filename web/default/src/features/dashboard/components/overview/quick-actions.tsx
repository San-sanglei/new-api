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
import { useTranslation } from 'react-i18next'
import { Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Wallet, KeyRound, RadioTower } from 'lucide-react'
import { Button } from '@/components/ui/button'

const ACTIONS = [
  { title: 'Top Up', desc: 'Add credits to your account', to: '/wallet', icon: Wallet },
  { title: 'Create API Key', desc: 'Generate a new access key', to: '/keys', icon: KeyRound },
  { title: 'View Channels', desc: 'Configure provider routing', to: '/channels', icon: RadioTower },
]

export function QuickActions() {
  const { t } = useTranslation()

  return (
    <div className='grid grid-cols-1 gap-3 sm:grid-cols-2'>
      {ACTIONS.map((action, i) => (
        <motion.div
          key={action.title}
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 + i * 0.06 }}
        >
          <Button
            variant='outline'
            className='h-auto w-full justify-start rounded-xl px-4 py-3'
            render={<Link to={action.to} />}
          >
            <span className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
              <action.icon className='size-4' />
            </span>
            <span className='flex min-w-0 flex-1 flex-col gap-0.5 text-left'>
              <span className='truncate text-sm font-medium'>{t(action.title)}</span>
              <span className='text-muted-foreground line-clamp-2 text-xs leading-relaxed'>
                {t(action.desc)}
              </span>
            </span>
          </Button>
        </motion.div>
      ))}
    </div>
  )
}
