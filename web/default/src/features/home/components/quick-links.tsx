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
import { Link } from '@tanstack/react-router'
import {
  BookOpen,
  KeyRound,
  RadioTower,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'

interface QuickLinkItem {
  title: string
  description: string
  href: string
  icon: LucideIcon
  external?: boolean
}

const quickLinks: QuickLinkItem[] = [
  {
    title: 'API Keys',
    description: 'Create and manage access tokens',
    href: '/keys',
    icon: KeyRound,
  },
  {
    title: 'Channels',
    description: 'Configure upstream providers and routing',
    href: '/channels',
    icon: RadioTower,
  },
  {
    title: 'Documentation',
    description: 'Guides, API reference, and examples',
    href: 'https://docs.newapi.pro',
    icon: BookOpen,
    external: true,
  },
]

interface QuickLinksProps {
  className?: string
}

export function QuickLinks(props: QuickLinksProps) {
  const { t } = useTranslation()

  return (
    <div className={cn('w-full max-w-2xl', props.className)}>
      <div className='text-muted-foreground/40 mb-4 text-center text-xs font-medium tracking-wider uppercase'>
        {t('Quick Links')}
      </div>
      <div className='grid grid-cols-2 gap-4'>
        {quickLinks.map((link) => {
          const card = (
            <div className='bg-card hover:bg-muted/50 group cursor-pointer rounded-xl border border-border/60 p-5 transition-all duration-150 hover:scale-[1.02] hover:shadow-lg hover:shadow-black/20'>
              <div className='bg-muted/70 group-hover:bg-blue-500/10 mb-3 inline-flex size-9 items-center justify-center rounded-lg transition-colors'>
                <link.icon className='text-blue-400 group-hover:text-blue-300 size-4' />
              </div>
              <h4 className='text-sm font-medium'>{link.title}</h4>
              <p className='text-muted-foreground/70 mt-1 text-xs leading-relaxed'>
                {link.description}
              </p>
            </div>
          )

          if (link.external) {
            return (
              <a
                key={link.title}
                href={link.href}
                target='_blank'
                rel='noopener noreferrer'
              >
                {card}
              </a>
            )
          }

          return (
            <Link key={link.title} to={link.href}>
              {card}
            </Link>
          )
        })}
      </div>
    </div>
  )
}
