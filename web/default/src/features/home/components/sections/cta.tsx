import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  if (props.isAuthenticated) {
    return null
  }

  return (
    <section className='py-20 lg:py-32'>
      <div className='mx-auto max-w-6xl px-6 text-center'>
        <h2 className='text-2xl font-bold lg:text-4xl'>
          {t('简化 AI 调用')}
        </h2>
        <div className='mt-8'>
          <Button size='lg' render={<Link to='/sign-up' />}>
            {t('立即注册')}
          </Button>
        </div>
      </div>
    </section>
  )
}
