import { useTranslation } from 'react-i18next'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      title: t('统一接口'),
      desc: t(
        '一个 API Key 访问所有主流模型。无需分别注册和管理多个服务商账号。'
      ),
      icon: (
        <svg
          xmlns='http://www.w3.org/2000/svg'
          width='24'
          height='24'
          viewBox='0 0 24 24'
          fill='none'
          stroke='currentColor'
          strokeWidth='1.5'
          strokeLinecap='round'
          strokeLinejoin='round'
        >
          <path d='M13 2 3 14h9l-1 8 10-12h-9l1-8z' />
        </svg>
      ),
    },
    {
      title: t('弹性计费'),
      desc: t(
        '按实际调用量计费，无最低消费。用量越大单价越低，成本透明可控。'
      ),
      icon: (
        <svg
          xmlns='http://www.w3.org/2000/svg'
          width='24'
          height='24'
          viewBox='0 0 24 24'
          fill='none'
          stroke='currentColor'
          strokeWidth='1.5'
          strokeLinecap='round'
          strokeLinejoin='round'
        >
          <path d='m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z' />
          <path d='m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65' />
          <path d='m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65' />
        </svg>
      ),
    },
    {
      title: t('稳定可靠'),
      desc: t(
        '99.9% 服务可用性保障。智能路由与自动故障转移，确保调用不中断。'
      ),
      icon: (
        <svg
          xmlns='http://www.w3.org/2000/svg'
          width='24'
          height='24'
          viewBox='0 0 24 24'
          fill='none'
          stroke='currentColor'
          strokeWidth='1.5'
          strokeLinecap='round'
          strokeLinejoin='round'
        >
          <path d='M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z' />
        </svg>
      ),
    },
  ]

  return (
    <section id='features' className='py-20 lg:py-28'>
      <div className='mx-auto max-w-6xl px-6'>
        <h2 className='text-center text-2xl font-bold lg:text-3xl'>
          {t('为什么选择 Took')}
        </h2>
        <div className='mt-14 grid grid-cols-1 gap-6 md:grid-cols-3'>
          {features.map((f) => (
            <div
              key={f.title}
              className='bg-background border-border hover:border-foreground/30 rounded-xl border p-6 transition-all duration-200 hover:-translate-y-px'
            >
              <div className='mb-4'>
                {f.icon}
              </div>
              <h3 className='mb-2 text-lg font-semibold'>{f.title}</h3>
              <p className='text-muted-foreground text-sm leading-relaxed'>
                {f.desc}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
