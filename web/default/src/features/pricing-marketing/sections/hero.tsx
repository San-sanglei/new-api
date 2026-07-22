import { Button } from '@/components/ui/button'
import { AnimateInView } from '@/components/animate-in-view'

export function Hero() {
  return (
    <section className='relative overflow-hidden px-4 pt-24 pb-16 sm:pt-32 sm:pb-20'>
      {/* Background gradient */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-x-0 top-0 h-[800px] opacity-30 dark:opacity-20'
        style={{
          background: [
            'radial-gradient(ellipse 70% 50% at 20% 20%, oklch(0.72 0.18 250 / 80%) 0%, transparent 70%)',
            'radial-gradient(ellipse 50% 40% at 80% 30%, oklch(0.65 0.15 200 / 60%) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 35% at 50% 80%, oklch(0.70 0.12 280 / 40%) 0%, transparent 70%)',
          ].join(', '),
          maskImage: 'linear-gradient(to bottom, black 40%, transparent 100%)',
          WebkitMaskImage: 'linear-gradient(to bottom, black 40%, transparent 100%)',
        }}
      />
      {/* Grid pattern */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 opacity-[0.03]'
        style={{
          backgroundImage: 'linear-gradient(to right, var(--border) 1px, transparent 1px), linear-gradient(to bottom, var(--border) 1px, transparent 1px)',
          backgroundSize: '60px 60px',
          maskImage: 'linear-gradient(to bottom, black 20%, transparent 90%)',
          WebkitMaskImage: 'linear-gradient(to bottom, black 20%, transparent 90%)',
        }}
      />
      <div className='relative mx-auto max-w-4xl text-center'>
        <AnimateInView animation='fade-up'>
          <div className='mb-4 inline-flex items-center gap-1.5 rounded-full border border-border/40 bg-muted/30 px-3 py-1 text-xs font-medium text-muted-foreground backdrop-blur-xs'>
            <span className='h-1.5 w-1.5 rounded-full bg-blue-500' />
            AI API 聚合平台
          </div>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={60}>
          <h1 className='text-[clamp(2.2rem,6vw,4rem)] leading-[1.1] font-bold tracking-tight'>
            灵活的 AI API 定价，<br className='hidden sm:block' />
            <span className='bg-linear-to-r from-blue-400 via-violet-400 to-purple-500 bg-clip-text text-transparent'>按需付费</span>
          </h1>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={120}>
          <p className='text-muted-foreground/80 mx-auto mt-4 max-w-2xl text-base leading-relaxed sm:text-lg'>
            从个人开发者到企业团队，找到适合你的方案
          </p>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={180}>
          <div className='mt-8 flex flex-wrap items-center justify-center gap-3'>
            <Button size='lg' render={<a href='/sign-up' />}>
              立即开始
            </Button>
            <Button variant='outline' size='lg' render={<a href='#plans' />}>
              查看方案
            </Button>
          </div>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={240}>
          <p className='text-muted-foreground/50 mt-8 text-xs'>无需信用卡 • 7 天免费试用 • 随时取消</p>
        </AnimateInView>
      </div>
    </section>
  )
}
