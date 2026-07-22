import { useState } from 'react'
import { PublicLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { Hero } from './sections/hero'
import { BillingToggle } from './sections/billing-toggle'
import { PricingCards } from './sections/pricing-cards'
import { ModelPriceTable } from './sections/model-price-table'
import { UsageCalculator } from './sections/usage-calculator'
import { TrustSection } from './sections/trust-section'
import { FAQ } from './sections/faq'
import { FinalCTA } from './sections/final-cta'
import type { BillingCycle } from './types'

export function PricingMarketingPage() {
  const [billingCycle, setBillingCycle] = useState<BillingCycle>('monthly')

  return (
    <PublicLayout showMainContainer={false}>
      <div className='relative'>
        <PageTransition>
          <Hero />
          <BillingToggle value={billingCycle} onChange={setBillingCycle} />
          <PricingCards billingCycle={billingCycle} />
          <ModelPriceTable />
          <UsageCalculator />
          <TrustSection />
          <FAQ />
          <FinalCTA />
        </PageTransition>
      </div>
    </PublicLayout>
  )
}
