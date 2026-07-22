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
import { Link, useRouterState } from '@tanstack/react-router'
import {
  BookOpen,
  Home,
  LayoutDashboard,
  LogIn,
  Package,
  Plus,
  Receipt,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { useNotifications } from '@/hooks/use-notifications'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { LanguageSwitcher } from '@/components/language-switcher'
import { NotificationPopover } from '@/components/notification-popover'
import { ThemeSwitch } from '@/components/theme-switch'
import { HeaderLogo } from '@/components/layout/components/header-logo'
import { useTheme } from '@/context/theme-provider'

interface NavItem {
  label: string
  href: string
  icon: LucideIcon
  external?: boolean
}

interface PublicSidebarProps {
  className?: string
}

export function PublicSidebar(props: PublicSidebarProps) {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const isAuthenticated = !!auth.user
  const {
    systemName,
    logo: systemLogo,
    loading,
    logoLoaded,
  } = useSystemConfig()
  const { resolvedTheme } = useTheme()
  const notifications = useNotifications()
  const routerState = useRouterState()
  const pathname = routerState.location.pathname

  const navItems: NavItem[] = [
    { label: t('Home'), href: '/', icon: Home },
    { label: t('Models'), href: '/pricing', icon: Package },
    { label: t('Pricing'), href: '/pricing', icon: LayoutDashboard },
    { label: t('Recharge'), href: '/wallet/recharge', icon: Plus },
    { label: t('Orders'), href: '/wallet/orders', icon: Receipt },
    {
      label: t('Docs'),
      href: 'https://docs.newapi.pro',
      icon: BookOpen,
      external: true,
    },
  ]

  return (
    <aside
      className={cn(
        'bg-sidebar text-sidebar-foreground flex h-svh w-64 shrink-0 flex-col border-r border-sidebar-border',
        props.className
      )}
    >
      {/* Logo / Brand */}
      <div className='flex h-16 shrink-0 items-center gap-2.5 px-5'>
        <div className='flex size-7 shrink-0 items-center justify-center'>
          {loading ? (
            <Skeleton className='size-full rounded-lg' />
          ) : (
            <HeaderLogo
              src={resolvedTheme === 'dark' ? systemLogo.replace(/\.png$/, '-white.png') : systemLogo}
              loading={loading}
              logoLoaded={logoLoaded}
              className='size-full rounded-lg object-contain'
            />
          )}
        </div>
        <span className='text-sm font-semibold tracking-tight'>
          {loading ? (
            <Skeleton className='h-4 w-16' />
          ) : (
            systemName || 'Took'
          )}
        </span>
      </div>

      {/* Navigation */}
      <nav className='flex-1 space-y-1 px-3 py-6'>
        {navItems.map((item) => {
          const isActive =
            !item.external &&
            (item.href === '/'
              ? pathname === '/'
              : pathname.startsWith(item.href))

          const linkContent = (
            <span
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors duration-150',
                isActive
                  ? 'bg-blue-500/10 text-blue-400 font-semibold'
                  : 'text-muted-foreground/80 hover:text-foreground hover:bg-accent/50'
              )}
            >
              <item.icon className='size-4 shrink-0' aria-hidden='true' />
              <span>{item.label}</span>
            </span>
          )

          if (item.external) {
            return (
              <a
                key={item.label}
                href={item.href}
                target='_blank'
                rel='noopener noreferrer'
                className='block'
              >
                {linkContent}
              </a>
            )
          }

          return (
            <Link key={item.label} to={item.href} className='block'>
              {linkContent}
            </Link>
          )
        })}
      </nav>

      {/* Footer: quick controls + sign-in button pinned at bottom */}
      <div className='border-t border-sidebar-border px-3 py-4'>
        <div className='mb-3 flex items-center justify-center gap-1'>
          <ThemeSwitch />
          <NotificationPopover
            open={notifications.popoverOpen}
            onOpenChange={notifications.setPopoverOpen}
            unreadCount={notifications.unreadCount}
            activeTab={notifications.activeTab}
            onTabChange={notifications.setActiveTab}
            notice={notifications.notice}
            announcements={notifications.announcements}
            loading={notifications.loading}
            onMarkAllRead={notifications.markAllRead}
          />
          <LanguageSwitcher />
        </div>
        {isAuthenticated ? (
          <Button
            className='w-full rounded-lg text-sm font-medium'
            size='sm'
            render={<Link to='/dashboard' />}
          >
            {t('Dashboard')}
          </Button>
        ) : (
          <Button
            variant='outline'
            className='border-sidebar-border text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground w-full rounded-lg text-sm font-medium'
            size='sm'
            render={<Link to='/sign-in' />}
          >
            <LogIn className='mr-1.5 size-4' aria-hidden='true' />
            {t('Sign in')}
          </Button>
        )}
      </div>
    </aside>
  )
}
