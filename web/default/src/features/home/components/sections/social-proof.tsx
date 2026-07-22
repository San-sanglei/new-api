import { useTranslation } from 'react-i18next'

interface SocialProofProps {
  className?: string
}

export function SocialProof(_props: SocialProofProps) {
  const { t } = useTranslation()

  return (
    <section className='border-border/50 border-y'>
      <div className='mx-auto max-w-6xl px-6 py-10 text-center'>
        <p className='text-muted-foreground/60 text-sm whitespace-nowrap'>
          {t('已被 10,000+ 开发者和 500+ 企业信赖')}
        </p>
        <p className='text-muted-foreground/60 mt-4 text-sm tracking-wider whitespace-nowrap'>
          OpenAI &middot; Anthropic &middot; Google &middot; Meta &middot; Mistral
        </p>
      </div>
    </section>
  )
}
