import type { ModelPrice } from '../types'

export const MODEL_PRICES: ModelPrice[] = [
  { name: 'GPT-5', provider: 'OpenAI', inputPrice: '¥0.55', outputPrice: '¥4.40', contextWindow: '128K' },
  { name: 'GPT-5-mini', provider: 'OpenAI', inputPrice: '¥0.11', outputPrice: '¥0.88', contextWindow: '128K' },
  { name: 'o3', provider: 'OpenAI', inputPrice: '¥1.32', outputPrice: '¥5.28', contextWindow: '128K' },
  { name: 'o4-mini', provider: 'OpenAI', inputPrice: '¥0.704', outputPrice: '¥2.816', contextWindow: '128K' },
  { name: 'Claude Opus 4.5', provider: 'Anthropic', inputPrice: '¥3.30', outputPrice: '¥16.50', contextWindow: '200K' },
  { name: 'Claude Sonnet 4.5', provider: 'Anthropic', inputPrice: '¥1.98', outputPrice: '¥9.90', contextWindow: '200K' },
  { name: 'Gemini 2.5 Pro', provider: 'Google', inputPrice: '¥0.594', outputPrice: '¥4.752', contextWindow: '2M' },
  { name: 'Gemini 2.5 Flash', provider: 'Google', inputPrice: '¥0.1414', outputPrice: '¥1.178', contextWindow: '1M' },
  { name: 'DeepSeek V3.1', provider: 'DeepSeek', inputPrice: '¥0.5028', outputPrice: '¥1.5084', contextWindow: '64K' },
  { name: 'DeepSeek R1', provider: 'DeepSeek', inputPrice: '¥0.5028', outputPrice: '¥2.011', contextWindow: '64K' },
  { name: 'Qwen3-235B', provider: 'Alibaba', inputPrice: '¥0.1786', outputPrice: '¥0.7144', contextWindow: '128K' },
  { name: 'Qwen3-32B', provider: 'Alibaba', inputPrice: '¥0.1786', outputPrice: '¥0.7144', contextWindow: '128K' },
  { name: 'GLM-4.5', provider: 'Zhipu', inputPrice: '¥0.3772', outputPrice: '¥1.5088', contextWindow: '128K' },
]
