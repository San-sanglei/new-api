import { useTranslation } from 'react-i18next'

interface StatsProps {
  className?: string
}

export function Stats(_props: StatsProps) {
  const { t } = useTranslation()

  const stats = [
    { value: '10,000+', label: t('开发者') },
    { value: '500+', label: t('企业客户') },
    { value: '50+', label: t('AI 模型') },
    { value: '99.9%', label: t('可用性') },
  ]

  return (
    <section className='bg-muted/30 py-20 lg:py-28'>
      <div className='mx-auto max-w-6xl px-6'>
        <div className='grid grid-cols-2 gap-8 text-center lg:grid-cols-4 lg:gap-12'>
          {stats.map((s) => (
            <div key={s.label}>
              <div className='text-3xl font-bold whitespace-nowrap lg:text-4xl'>
                {s.value}
              </div>
              <p className='text-muted-foreground mt-2 text-sm whitespace-nowrap'>
                {s.label}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
