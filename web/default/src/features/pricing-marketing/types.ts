export type BillingCycle = 'payg' | 'monthly' | 'annual'

export interface PlanFeature {
  text: string
  included: boolean
}

export interface PricingPlan {
  id: string
  name: string
  description: string
  price: Record<BillingCycle, string>
  priceSub: Record<BillingCycle, string>
  highlighted?: boolean
  badge?: string
  badgeColor?: string
  features: PlanFeature[]
  cta: { text: string; href: string; variant: 'default' | 'outline' }
}

export interface ModelPrice {
  name: string
  provider: string
  inputPrice: string
  outputPrice: string
  contextWindow: string
}

export interface FAQItem {
  question: string
  answer: string
}
