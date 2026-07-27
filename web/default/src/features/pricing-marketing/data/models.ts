import type { ModelPrice } from '../types'

export const MODEL_PRICES: ModelPrice[] = [
  { name: 'GPT-5', provider: 'OpenAI', inputPrice: '¥5.00', outputPrice: '¥20.00', contextWindow: '128K' },
  { name: 'GPT-5-mini', provider: 'OpenAI', inputPrice: '¥0.25', outputPrice: '¥1.00', contextWindow: '128K' },
  { name: 'o3', provider: 'OpenAI', inputPrice: '¥10.00', outputPrice: '¥40.00', contextWindow: '128K' },
  { name: 'o4-mini', provider: 'OpenAI', inputPrice: '¥0.50', outputPrice: '¥2.00', contextWindow: '128K' },
  { name: 'Claude Opus 4.5', provider: 'Anthropic', inputPrice: '¥7.50', outputPrice: '¥37.50', contextWindow: '200K' },
  { name: 'Claude Sonnet 4.5', provider: 'Anthropic', inputPrice: '¥3.00', outputPrice: '¥15.00', contextWindow: '200K' },
  { name: 'Gemini 2.5 Pro', provider: 'Google', inputPrice: '¥1.25', outputPrice: '¥10.00', contextWindow: '2M' },
  { name: 'Gemini 2.5 Flash', provider: 'Google', inputPrice: '¥0.08', outputPrice: '¥0.30', contextWindow: '1M' },
  { name: 'DeepSeek V3.1', provider: 'DeepSeek', inputPrice: '¥0.28', outputPrice: '¥0.42', contextWindow: '64K' },
  { name: 'DeepSeek R1', provider: 'DeepSeek', inputPrice: '¥0.55', outputPrice: '¥2.19', contextWindow: '64K' },
  { name: 'Qwen3-235B', provider: 'Alibaba', inputPrice: '¥0.30', outputPrice: '¥0.90', contextWindow: '128K' },
  { name: 'Qwen3-32B', provider: 'Alibaba', inputPrice: '¥0.10', outputPrice: '¥0.30', contextWindow: '128K' },
  { name: 'GLM-4.5', provider: 'Zhipu', inputPrice: '¥0.50', outputPrice: '¥1.50', contextWindow: '128K' },
]
