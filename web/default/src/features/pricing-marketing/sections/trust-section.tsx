import { useEffect, useState, useRef } from 'react'
import { AnimateInView } from '@/components/animate-in-view'
import { Shield, Clock, RefreshCw } from 'lucide-react'

function useCountUp(target: number, duration: number, startOnView: boolean) {
  const [count, setCount] = useState(0)
  const ref = useRef<HTMLSpanElement>(null)
  const hasStarted = useRef(false)

  useEffect(() => {
    if (!startOnView || hasStarted.current) return
    hasStarted.current = true
    let start = 0
    const step = (timestamp: number) => {
      if (!start) start = timestamp
      const progress = Math.min((timestamp - start) / duration, 1)
      setCount(Math.floor(progress * target))
      if (progress < 1) requestAnimationFrame(step)
    }
    requestAnimationFrame(step)
  }, [startOnView, target, duration])

  return { count, ref }
}

function StatCard({ value, suffix, label }: { value: string; suffix?: string; label: string }) {
  return (
    <div className='text-center'>
      <p className='text-2xl font-bold sm:text-3xl'>
        {value}
        {suffix && <span className='text-blue-400'>{suffix}</span>}
      </p>
      <p className='text-muted-foreground mt-1 text-sm'>{label}</p>
    </div>
  )
}

export function TrustSection() {
  const [inView, setInView] = useState(false)
  const sectionRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = sectionRef.current
    if (!el) return
    const obs = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) { setInView(true); obs.disconnect() } },
      { threshold: 0.3 }
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [])

  return (
    <section ref={sectionRef} className='px-4 py-12'>
      <div className='mx-auto max-w-5xl'>
        <AnimateInView animation='fade-up'>
          <div className='mb-10 grid grid-cols-2 gap-6 md:grid-cols-4'>
            <StatCard value={inView ? '10,000+' : '0'} label='开发者信赖' />
            <StatCard value={inView ? '99.9' : '0'} suffix='%' label='服务可用性' />
            <StatCard value={inView ? '40+' : '0'} label='AI 模型支持' />
            <StatCard value={inView ? '99' : '0'} suffix='%' label='客户满意度' />
          </div>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={80}>
          <div className='mb-10 grid grid-cols-1 gap-4 sm:grid-cols-3'>
            {[
              { icon: Shield, text: '企业级安全', sub: 'SOC 2 合规，数据加密' },
              { icon: Clock, text: '99.9% 可用性', sub: '多节点高可用架构' },
              { icon: RefreshCw, text: '7 天退款', sub: '无风险试用，满意再付' },
            ].map((item) => (
              <div key={item.text} className='flex items-center gap-3 rounded-lg border border-border/40 bg-card/30 p-4 backdrop-blur-sm'>
                <div className='flex h-10 w-10 items-center justify-center rounded-lg bg-blue-500/10'>
                  <item.icon className='h-5 w-5 text-blue-400' />
                </div>
                <div>
                  <p className='text-sm font-medium'>{item.text}</p>
                  <p className='text-muted-foreground text-xs'>{item.sub}</p>
                </div>
              </div>
            ))}
          </div>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={120}>
          <p className='text-muted-foreground/50 mb-4 text-center text-xs font-medium uppercase tracking-wider'>
            合作伙伴
          </p>
          <div className='flex flex-wrap items-center justify-center gap-6 opacity-30'>
            {['OpenAI', 'Anthropic', 'Google', 'DeepSeek', 'Meta'].map((name) => (
              <div key={name} className='flex h-10 items-center rounded-md border border-border/30 bg-card/20 px-4 text-xs font-medium text-muted-foreground'>
                {name}
              </div>
            ))}
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}
