import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { AnimateInView } from '@/components/animate-in-view'
import { FAQ_ITEMS } from '../data/faq'

export function FAQ() {
  return (
    <section className='px-4 py-12'>
      <div className='mx-auto max-w-3xl'>
        <AnimateInView animation='fade-up'>
          <h2 className='mb-2 text-center text-2xl font-bold'>常见问题</h2>
          <p className='text-muted-foreground/70 mb-8 text-center text-sm'>
            还有其他疑问？请联系我们的支持团队
          </p>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={80}>
          <Accordion type='multiple' className='space-y-2'>
            {FAQ_ITEMS.map((item, i) => (
              <AccordionItem
                key={i}
                value={`faq-${i}`}
                className='rounded-lg border border-border/40 bg-card/20 px-4'
              >
                <AccordionTrigger className='text-sm font-medium'>
                  {item.question}
                </AccordionTrigger>
                <AccordionContent className='text-muted-foreground text-sm leading-relaxed'>
                  {item.answer}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </AnimateInView>
      </div>
    </section>
  )
}
