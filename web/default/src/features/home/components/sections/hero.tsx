import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()

  return (
    <section className='px-6 pt-28 pb-16 lg:pt-40 lg:pb-32'>
      <div className='mx-auto flex max-w-6xl flex-col gap-12 px-6 lg:flex-row lg:items-center lg:gap-16'>
        {/* Left: 60% */}
        <div className='w-full lg:w-[60%]'>
          <p
            className='text-muted-foreground mb-4 font-semibold tracking-wide'
            style={{ fontSize: 'clamp(20px, 2.2vw, 30px)' }}
          >
            {t('源头供应商 · 满血API直连')}
          </p>
          <h1
            className='font-bold leading-tight tracking-tight'
            style={{ fontSize: 'clamp(28px, 3.5vw, 48px)' }}
          >
            {t('一站式 AI Token 聚合平台')}
          </h1>
          <p className='text-muted-foreground mt-6 max-w-xl text-base leading-relaxed lg:text-lg'>
            {t(
              '整合全球主流 AI 模型 API，统一额度管理，按需弹性计费。为开发者和企业提供最简洁的智能调用体验。'
            )}
          </p>
          <div className='mt-8 flex flex-wrap items-center gap-4'>
            {props.isAuthenticated ? (
              <Button render={<Link to='/dashboard' />} size='lg'>
                {t('Go to Dashboard')}
              </Button>
            ) : (
              <Button render={<Link to='/sign-up' />} size='lg'>
                {t('免费开始')}
              </Button>
            )}
            <Button
              variant='ghost'
              className='text-muted-foreground hover:text-foreground'
              render={<Link to='/pricing' />}
            >
              {t('查看文档')} &rarr;
            </Button>
          </div>
        </div>

        {/* Right: 40% */}
        <div className='w-full lg:w-[40%]'>
          <HeroTerminalDemo className='w-full' />
        </div>
      </div>
    </section>
  )
}
