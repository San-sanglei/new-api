import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { BillingCycle } from '../types'

interface BillingToggleProps {
  value: BillingCycle
  onChange: (value: BillingCycle) => void
}

const OPTIONS: { value: BillingCycle; label: string }[] = [
  { value: 'payg', label: '按量付费' },
  { value: 'monthly', label: '包月套餐' },
  { value: 'annual', label: '包年套餐（省20%）' },
]

export function BillingToggle({ value, onChange }: BillingToggleProps) {
  return (
    <section id='plans' className='px-4 pt-8 pb-4'>
      <div className='mx-auto flex max-w-4xl justify-center'>
        <Tabs value={value} onValueChange={(v) => onChange(v as BillingCycle)}>
          <TabsList>
            {OPTIONS.map((opt) => (
              <TabsTrigger key={opt.value} value={opt.value}>
                {opt.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      </div>
    </section>
  )
}
