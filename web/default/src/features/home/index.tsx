import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Markdown } from '@/components/ui/markdown'
import { PublicLayout } from '@/components/layout'
import type { TopNavLink } from '@/components/layout'
import { useAuthStore } from '@/stores/auth-store'
import { useStatus } from '@/hooks/use-status'
import { useHomePageContent } from './hooks'
import { Hero } from './components/sections/hero'
import { SocialProof } from './components/sections/social-proof'
import { Features } from './components/sections/features'
import { HowItWorks } from './components/sections/how-it-works'
import { Stats } from './components/sections/stats'
import { CTA } from './components/sections/cta'

export function Home() {
  const { t } = useTranslation()
  const { content, isLoaded, isUrl } = useHomePageContent()
  const isAuthenticated = useAuthStore((s) => !!s.auth.user)
  const { status } = useStatus()

  const navLinks = useMemo<TopNavLink[]>(() => {
    const docsUrl =
      (status?.docs_link as string | undefined) || '/docs'
    const links: TopNavLink[] = [
      {
        title: t('产品特性'),
        href: '#features',
      },
      {
        title: t('价格'),
        href: '/pricing',
      },
      {
        title: t('文档'),
        href: docsUrl,
        external: docsUrl.startsWith('http'),
      },
    ]
    return links
  }, [status, t])

  const headerProps = {
    overrideDynamicNav: true,
    navLinks,
    showAuthButtons: true,
  }

  if (!isLoaded) {
    return (
      <PublicLayout showMainContainer={false} headerProps={headerProps}>
        <main className='flex min-h-svh items-center justify-center'>
          <div className='text-muted-foreground'>{t('Loading...')}</div>
        </main>
      </PublicLayout>
    )
  }

  if (content) {
    return (
      <PublicLayout showMainContainer={false} headerProps={headerProps}>
        <main className='flex flex-1 overflow-x-hidden pt-16'>
          {isUrl ? (
            <iframe
              src={content}
              className='h-svh w-full border-none'
              title={t('Custom Home Page')}
            />
          ) : (
            <div className='container mx-auto py-8'>
              <Markdown className='custom-home-content'>{content}</Markdown>
            </div>
          )}
        </main>
      </PublicLayout>
    )
  }

  return (
    <PublicLayout showMainContainer={false} headerProps={headerProps}>
      <main className='flex-1 overflow-x-hidden pt-16'>
        <Hero isAuthenticated={isAuthenticated} />
        <SocialProof />
        <Features />
        <HowItWorks />
        <Stats />
        <CTA isAuthenticated={isAuthenticated} />
      </main>
    </PublicLayout>
  )
}
