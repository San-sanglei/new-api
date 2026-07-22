import { motion } from 'motion/react'
import { AnimateInView } from '@/components/animate-in-view'
import { MODEL_PRICES } from '../data/models'
import { STAGGER_VARIANTS, TABLE_ROW_VARIANTS } from '@/lib/motion'

export function ModelPriceTable() {
  return (
    <section className='px-4 py-12'>
      <div className='mx-auto max-w-4xl'>
        <AnimateInView animation='fade-up'>
          <h2 className='mb-2 text-center text-2xl font-bold'>模型定价</h2>
          <p className='text-muted-foreground/70 mb-8 text-center text-sm'>
            按量计费，用多少付多少。价格为每百万 Token 计算。
          </p>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={80}>
          <div className='overflow-hidden rounded-xl border border-border/50 bg-card/30 backdrop-blur-sm'>
            <table className='w-full text-sm'>
              <thead>
                <tr className='border-b border-border/40 bg-muted/20'>
                  <th className='px-4 py-3 text-left font-medium text-muted-foreground'>模型</th>
                  <th className='px-4 py-3 text-left font-medium text-muted-foreground'>提供商</th>
                  <th className='px-4 py-3 text-right font-medium text-muted-foreground'>输入 (¥/1M tokens)</th>
                  <th className='px-4 py-3 text-right font-medium text-muted-foreground'>输出 (¥/1M tokens)</th>
                  <th className='px-4 py-3 text-right font-medium text-muted-foreground'>上下文</th>
                </tr>
              </thead>
              <motion.tbody
                variants={STAGGER_VARIANTS}
                initial='initial'
                animate='animate'
              >
                {MODEL_PRICES.map((model) => (
                  <motion.tr
                    key={model.name}
                    variants={TABLE_ROW_VARIANTS}
                    className='border-b border-border/20 transition-colors last:border-0 hover:bg-muted/10'
                  >
                    <td className='px-4 py-3 font-medium'>{model.name}</td>
                    <td className='text-muted-foreground px-4 py-3'>{model.provider}</td>
                    <td className='px-4 py-3 text-right tabular-nums'>{model.inputPrice}</td>
                    <td className='px-4 py-3 text-right tabular-nums'>{model.outputPrice}</td>
                    <td className='text-muted-foreground px-4 py-3 text-right'>{model.contextWindow}</td>
                  </motion.tr>
                ))}
              </motion.tbody>
            </table>
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}
