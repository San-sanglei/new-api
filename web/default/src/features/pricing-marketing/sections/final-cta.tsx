import { Button } from '@/components/ui/button'
import { AnimateInView } from '@/components/animate-in-view'

export function FinalCTA() {
  return (
    <section className='relative overflow-hidden px-4 py-16'>
      <div
        aria-hidden
        className='pointer-events-none absolute inset-x-0 top-0 h-full opacity-20'
        style={{
          background: [
            'radial-gradient(ellipse 60% 50% at 30% 50%, oklch(0.72 0.18 250 / 60%) 0%, transparent 70%)',
            'radial-gradient(ellipse 50% 40% at 70% 50%, oklch(0.65 0.15 200 / 40%) 0%, transparent 70%)',
          ].join(', '),
        }}
      />
      <div className='relative mx-auto max-w-2xl text-center'>
        <AnimateInView animation='fade-up'>
          <h2 className='text-2xl font-bold sm:text-3xl'>
            <span className='bg-linear-to-r from-blue-400 via-violet-400 to-purple-500 bg-clip-text text-transparent'>
              还在犹豫？开始免费试用
            </span>
          </h2>
          <p className='text-muted-foreground/70 mt-3 text-sm sm:text-base'>
            无需信用卡，立即开始体验。7 天免费试用，随时取消。
          </p>
          <div className='mt-8 flex flex-wrap items-center justify-center gap-3'>
            <Button size='lg' render={<a href='/sign-up' />}>
              立即开始
            </Button>
            <Button variant='outline' size='lg' render={<a href='/contact' />}>
              联系销售
            </Button>
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}
