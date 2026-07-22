import { Check, X } from 'lucide-react'
import { motion } from 'motion/react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { AnimateInView } from '@/components/animate-in-view'
import { PLANS } from '../data/plans'
import type { BillingCycle } from '../types'

interface PricingCardsProps {
  billingCycle: BillingCycle
}

export function PricingCards({ billingCycle }: PricingCardsProps) {
  return (
    <section className='px-4 py-8'>
      <div className='mx-auto grid max-w-6xl gap-6 md:grid-cols-3'>
        {PLANS.map((plan, i) => (
          <AnimateInView key={plan.id} animation='fade-up' delay={i * 100}>
            <motion.div
              whileHover={{ y: -4 }}
              className={`relative flex flex-col rounded-xl border bg-card/50 backdrop-blur-sm p-6 transition-shadow ${
                plan.highlighted
                  ? 'border-blue-500/50 shadow-[0_0_30px_-8px_rgba(59,130,246,0.3)]'
                  : 'border-border/50'
              }`}
            >
              {plan.badge && (
                <Badge
                  className={`absolute -top-2.5 left-4 px-3 py-0.5 text-xs font-medium ${plan.badgeColor}`}
                  variant='outline'
                >
                  {plan.badge}
                </Badge>
              )}
              <div className='mb-4'>
                <h3 className='text-lg font-semibold'>{plan.name}</h3>
                <p className='text-muted-foreground mt-1 text-sm'>{plan.description}</p>
              </div>
              <div className='mb-2'>
                <span className='text-3xl font-bold'>{plan.price[billingCycle]}</span>
                {plan.priceSub[billingCycle] && (
                  <span className='text-muted-foreground ml-2 text-sm'>{plan.priceSub[billingCycle]}</span>
                )}
              </div>
              <Button
                variant={plan.cta.variant}
                className='mb-6 mt-2 w-full'
                render={<a href={plan.cta.href} />}
              >
                {plan.cta.text}
              </Button>
              <ul className='space-y-3 text-sm'>
                {plan.features.map((feat, fi) => (
                  <li key={fi} className='flex items-start gap-2.5'>
                    {feat.included ? (
                      <Check className='mt-0.5 h-4 w-4 shrink-0 text-emerald-500' />
                    ) : (
                      <X className='text-muted-foreground/30 mt-0.5 h-4 w-4 shrink-0' />
                    )}
                    <span className={feat.included ? 'text-foreground' : 'text-muted-foreground/40'}>
                      {feat.text}
                    </span>
                  </li>
                ))}
              </ul>
            </motion.div>
          </AnimateInView>
        ))}
      </div>
    </section>
  )
}
