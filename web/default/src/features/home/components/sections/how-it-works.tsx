import { useTranslation } from 'react-i18next'

interface HowItWorksProps {
  className?: string
}

export function HowItWorks(_props: HowItWorksProps) {
  const { t } = useTranslation()

  const steps = [
    { num: '01', title: t('注册账号'), desc: t('获取统一的 API Key') },
    { num: '02', title: t('选择模型'), desc: t('浏览并选择需要的 AI 模型') },
    { num: '03', title: t('开始调用'), desc: t('通过标准接口发起请求') },
  ]

  return (
    <section className='border-border/50 border-t py-20 lg:py-28'>
      <div className='mx-auto max-w-6xl px-6'>
        <h2 className='text-center text-2xl font-bold lg:text-3xl'>
          {t('三步开始使用')}
        </h2>
        <div className='mt-14 grid grid-cols-1 gap-12 md:grid-cols-3'>
          {steps.map((step) => (
            <div key={step.num} className='text-center md:text-left'>
              <span className='text-muted-foreground/60 font-mono text-sm font-medium'>
                {step.num}
              </span>
              <h3 className='mt-3 text-xl font-semibold'>{step.title}</h3>
              <p className='text-muted-foreground mt-2 text-sm'>{step.desc}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
