import { useState, useMemo } from 'react'
import { motion } from 'motion/react'
import { AnimateInView } from '@/components/animate-in-view'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const MODELS = [
  { id: 'gpt4o', name: 'GPT-4o', input: 2.5, output: 10 },
  { id: 'claude35', name: 'Claude 3.5 Sonnet', input: 3.0, output: 15 },
  { id: 'deepseek', name: 'DeepSeek V3', input: 0.28, output: 0.42 },
  { id: 'gemini', name: 'Gemini 1.5 Pro', input: 1.25, output: 10 },
]

export function UsageCalculator() {
  const [calls, setCalls] = useState('100000')
  const [modelId, setModelId] = useState('gpt4o')

  const estimate = useMemo(() => {
    const model = MODELS.find((m) => m.id === modelId)
    if (!model || !calls) return { monthly: 0, savings: 0 }
    const num = Number(calls)
    if (num <= 0) return { monthly: 0, savings: 0 }
    const avgTokens = 2000
    const monthly = ((num * avgTokens * (model.input + model.output)) / 1000000)
    const selfHostCost = monthly * 3.5
    const savings = selfHostCost > 0 ? Math.round((1 - monthly / selfHostCost) * 100) : 0
    return { monthly: Math.round(monthly * 100) / 100, savings }
  }, [calls, modelId])

  return (
    <section className='px-4 py-12'>
      <div className='mx-auto max-w-2xl'>
        <AnimateInView animation='fade-up'>
          <h2 className='mb-2 text-center text-2xl font-bold'>费用估算器</h2>
          <p className='text-muted-foreground/70 mb-8 text-center text-sm'>
            输入您的预期用量，快速估算每月费用
          </p>
        </AnimateInView>
        <AnimateInView animation='fade-up' delay={80}>
          <div className='rounded-xl border border-border/50 bg-card/40 backdrop-blur-sm p-6'>
            <div className='mb-5 space-y-4'>
              <div>
                <label className='mb-1.5 block text-sm font-medium'>每月预期调用次数</label>
                <Input
                  type='number'
                  value={calls}
                  onChange={(e) => setCalls(e.target.value)}
                  placeholder='例如: 100000'
                  min={0}
                />
              </div>
              <div>
                <label className='mb-1.5 block text-sm font-medium'>主要使用模型</label>
                <Select value={modelId} onValueChange={setModelId}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {MODELS.map((m) => (
                      <SelectItem key={m.id} value={m.id}>{m.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <motion.div
              key={`${calls}-${modelId}`}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className='rounded-lg border border-border/40 bg-muted/20 p-4 text-center'
            >
              <p className='text-muted-foreground text-sm'>预估费用</p>
              <p className='text-3xl font-bold'>
                ¥{estimate.monthly.toLocaleString()}
                <span className='text-muted-foreground text-base font-normal'>/月</span>
              </p>
              {estimate.savings > 0 && (
                <p className='mt-2 text-sm text-emerald-500'>
                  相比自建服务器可节省约 {estimate.savings}%
                </p>
              )}
            </motion.div>
          </div>
        </AnimateInView>
      </div>
    </section>
  )
}
