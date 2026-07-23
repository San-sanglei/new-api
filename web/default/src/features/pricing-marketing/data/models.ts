import type { ModelPrice } from '../types'

export const MODEL_PRICES: ModelPrice[] = [
  { name: 'GPT-4o', provider: 'OpenAI', inputPrice: '¥2.50', outputPrice: '¥10.00', contextWindow: '128K' },
  { name: 'GPT-4o-mini', provider: 'OpenAI', inputPrice: '¥0.15', outputPrice: '¥0.60', contextWindow: '128K' },
  { name: 'Claude 3.5 Sonnet', provider: 'Anthropic', inputPrice: '¥3.00', outputPrice: '¥15.00', contextWindow: '200K' },
  { name: 'Claude 3.5 Haiku', provider: 'Anthropic', inputPrice: '¥0.80', outputPrice: '¥4.00', contextWindow: '200K' },
  { name: 'DeepSeek V3', provider: 'DeepSeek', inputPrice: '¥0.28', outputPrice: '¥0.42', contextWindow: '64K' },
  { name: 'DeepSeek R1', provider: 'DeepSeek', inputPrice: '¥0.55', outputPrice: '¥2.19', contextWindow: '64K' },
  { name: 'Gemini 1.5 Pro', provider: 'Google', inputPrice: '¥1.25', outputPrice: '¥10.00', contextWindow: '2M' },
  { name: 'Gemini 1.5 Flash', provider: 'Google', inputPrice: '¥0.08', outputPrice: '¥0.30', contextWindow: '1M' },
]
