package ratio_setting

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/types"
)

// from songquanpeng/one-api
const (
	USD2RMB = 7.3 // 暂定 1 USD = 7.3 RMB
	USD     = 500 // $0.002 = 1 -> $1 = 500
	RMB     = USD / USD2RMB
)

// modelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// TODO: when a new api is enabled, check the pricing here
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens

var defaultModelRatio = map[string]float64{
	//"midjourney":                50,
	"gpt-4-gizmo-*":  15,
	"gpt-4o-gizmo-*": 2.5,
	"gpt-4-all":      15,
	"gpt-4o-all":     15,
	"gpt-4":          3.19, // RunAPI: 2.9 × 1.10
	//"gpt-4-0314":                   15, //deprecated
	"gpt-4-0613": 15,
	"gpt-4-32k":  30,
	//"gpt-4-32k-0314":               30, //deprecated
	"gpt-4-32k-0613":                          30,
	"gpt-4-1106-preview":                      5,    // $10 / 1M tokens
	"gpt-4-0125-preview":                      5,    // $10 / 1M tokens
	"gpt-4-turbo-preview":                     5,    // $10 / 1M tokens
	"gpt-4-vision-preview":                    5,    // $10 / 1M tokens
	"gpt-4-1106-vision-preview":               5,    // $10 / 1M tokens
	"chatgpt-4o-latest":                       2.5,  // $5 / 1M tokens
	"gpt-4o":                                  0.55, // RunAPI: 0.5 × 1.10
	"gpt-4o-audio-preview":                    0.55, // RunAPI: 0.5 × 1.10
	"gpt-4o-audio-preview-2024-10-01":         1.25, // $2.5 / 1M tokens
	"gpt-4o-2024-05-13":                       2.5,  // $5 / 1M tokens
	"gpt-4o-2024-08-06":                       1.25, // $2.5 / 1M tokens
	"gpt-4o-2024-11-20":                       1.25, // $2.5 / 1M tokens
	"gpt-4o-realtime-preview":                 2.5,
	"gpt-4o-realtime-preview-2024-10-01":      2.5,
	"gpt-4o-realtime-preview-2024-12-17":      2.5,
	"gpt-4o-mini-realtime-preview":            0.3,
	"gpt-4o-mini-realtime-preview-2024-12-17": 0.3,
	"gpt-4.1":                          0.44,   // RunAPI: 0.4 × 1.10
	"gpt-4.1-2025-04-14":               1.0,    // $2 / 1M tokens
	"gpt-4.1-mini":                     0.088,  // RunAPI: 0.08 × 1.10
	"gpt-4.1-mini-2025-04-14":          0.2,    // $0.4 / 1M tokens
	"gpt-4.1-nano":                     0.022,  // RunAPI: 0.02 × 1.10
	"gpt-4.1-nano-2025-04-14":          0.05,   // $0.1 / 1M tokens
	"o1":                               5.28,   // RunAPI: 4.8 × 1.10
	"o1-2024-12-17":                    7.5,    // $15 / 1M tokens
	"o1-preview":                       7.5,    // $15 / 1M tokens
	"o1-preview-2024-09-12":            7.5,    // $15 / 1M tokens
	"o1-mini":                          0.3509, // RunAPI: 0.319 × 1.10
	"o1-mini-2024-09-12":               0.55,   // $1.1 / 1M tokens
	"o1-pro":                           82.5,   // RunAPI: 75 × 1.10
	"o1-pro-2025-03-19":                75.0,   // $150 / 1M tokens
	"o3-mini":                          0.352,  // RunAPI: 0.32 × 1.10
	"o3-mini-2025-01-31":               0.55,
	"o3-mini-high":                     0.55,
	"o3-mini-2025-01-31-high":          0.55,
	"o3-mini-low":                      0.55,
	"o3-mini-2025-01-31-low":           0.55,
	"o3-mini-medium":                   0.55,
	"o3-mini-2025-01-31-medium":        0.55,
	"o3":                               0.66,  // RunAPI: 0.6 × 1.10
	"o3-2025-04-16":                    1.0,   // $2 / 1M tokens
	"o3-pro":                           6.6,   // RunAPI: 6 × 1.10
	"o3-pro-2025-06-10":                10.0,  // $20 / 1M tokens
	"o3-deep-research":                 5.5,   // RunAPI: 5 × 1.10
	"o3-deep-research-2025-06-26":      5.0,   // $10 / 1M tokens
	"o4-mini":                          0.352, // RunAPI: 0.32 × 1.10
	"o4-mini-2025-04-16":               0.55,  // $1.1 / 1M tokens
	"o4-mini-deep-research":            1.1,   // RunAPI: 1 × 1.10
	"o4-mini-deep-research-2025-06-26": 1.0,   // $2 / 1M tokens
	"gpt-4o-mini":                      0.033, // RunAPI: 0.03 × 1.10
	"gpt-4o-mini-2024-07-18":           0.075,
	"gpt-4-turbo":                      5, // $0.01 / 1K tokens
	"gpt-4-turbo-2024-04-09":           5, // $0.01 / 1K tokens
	"gpt-4.5-preview":                  37.5,
	"gpt-4.5-preview-2025-02-27":       37.5,
	"gpt-5":                            0.275, // RunAPI: 0.25 × 1.10
	"gpt-5-2025-08-07":                 0.275, // RunAPI: 0.25 × 1.10
	"gpt-5-chat-latest":                0.275, // RunAPI: 0.25 × 1.10
	"gpt-5-mini":                       0.055, // RunAPI: 0.05 × 1.10
	"gpt-5-mini-2025-08-07":            0.055, // RunAPI: 0.05 × 1.10
	"gpt-5-nano":                       0.011, // RunAPI: 0.01 × 1.10
	"gpt-5-nano-2025-08-07":            0.011, // RunAPI: 0.01 × 1.10
	//"gpt-3.5-turbo-0301":           0.75, //deprecated
	"gpt-3.5-turbo":          0.25,
	"gpt-3.5-turbo-0613":     0.75,
	"gpt-3.5-turbo-16k":      1.5, // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613": 1.5,
	"gpt-3.5-turbo-instruct": 0.75, // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":     0.5,  // $0.001 / 1K tokens
	"gpt-3.5-turbo-0125":     0.25,
	"babbage-002":            0.2, // $0.0004 / 1K tokens
	"davinci-002":            1,   // $0.002 / 1K tokens
	"text-ada-001":           0.2,
	"text-babbage-001":       0.25,
	"text-curie-001":         1,
	//"text-davinci-002":               10,
	//"text-davinci-003":               10,
	"text-davinci-edit-001":                     10,
	"code-davinci-edit-001":                     10,
	"whisper-1":                                 15,    // $0.006 / minute -> $0.006 / 150 words -> $0.006 / 200 tokens -> $0.03 / 1k tokens
	"tts-1":                                     0.594, // RunAPI: 0.54 × 1.10
	"tts-1-1106":                                7.5,   // 1k characters -> $0.015
	"tts-1-hd":                                  1.188, // RunAPI: 1.08 × 1.10
	"tts-1-hd-1106":                             15,    // 1k characters -> $0.03
	"davinci":                                   10,
	"curie":                                     10,
	"babbage":                                   10,
	"ada":                                       10,
	"text-embedding-3-small":                    0.0044, // RunAPI: 0.004 × 1.10
	"text-embedding-3-large":                    0.0286, // RunAPI: 0.026 × 1.10
	"text-embedding-ada-002":                    0.022,  // RunAPI: 0.02 × 1.10
	"text-search-ada-doc-001":                   10,
	"text-moderation-stable":                    0.1,
	"text-moderation-latest":                    0.1,
	"claude-3-haiku-20240307":                   0.125, // $0.25 / 1M tokens
	"claude-3-5-haiku-20241022":                 0.5,   // $1 / 1M tokens
	"claude-haiku-4-5-20251001":                 0.33,  // RunAPI: 0.3 × 1.10
	"claude-3-sonnet-20240229":                  1.5,   // $3 / 1M tokens
	"claude-3-5-sonnet-20240620":                1.5,
	"claude-3-5-sonnet-20241022":                1.5,
	"claude-3-7-sonnet-20250219":                1.5,
	"claude-3-7-sonnet-20250219-thinking":       1.5,
	"claude-sonnet-4-20250514":                  0.99, // RunAPI: 0.9 × 1.10
	"claude-sonnet-4-5-20250929":                0.99, // RunAPI: 0.9 × 1.10
	"claude-opus-4-5-20251101":                  1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-6":                           1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-6-max":                       2.5,
	"claude-opus-4-6-high":                      2.5,
	"claude-opus-4-6-medium":                    2.5,
	"claude-opus-4-6-low":                       2.5,
	"claude-opus-4-7":                           1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-7-max":                       2.5,
	"claude-opus-4-7-xhigh":                     2.5,
	"claude-opus-4-7-high":                      2.5,
	"claude-opus-4-7-medium":                    2.5,
	"claude-opus-4-7-low":                       2.5,
	"claude-opus-4-8":                           1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-8-max":                       2.5,
	"claude-opus-4-8-xhigh":                     2.5,
	"claude-opus-4-8-high":                      2.5,
	"claude-opus-4-8-medium":                    2.5,
	"claude-opus-4-8-low":                       2.5,
	"claude-3-opus-20240229":                    7.5,  // $15 / 1M tokens
	"claude-opus-4-20250514":                    4.95, // RunAPI: 4.5 × 1.10
	"claude-opus-4-1-20250805":                  4.95, // RunAPI: 4.5 × 1.10
	"ERNIE-4.0-8K":                              0.120 * RMB,
	"ERNIE-3.5-8K":                              0.012 * RMB,
	"ERNIE-3.5-8K-0205":                         0.024 * RMB,
	"ERNIE-3.5-8K-1222":                         0.012 * RMB,
	"ERNIE-Bot-8K":                              0.024 * RMB,
	"ERNIE-3.5-4K-0205":                         0.012 * RMB,
	"ERNIE-Speed-8K":                            0.004 * RMB,
	"ERNIE-Speed-128K":                          0.004 * RMB,
	"ERNIE-Lite-8K-0922":                        0.008 * RMB,
	"ERNIE-Lite-8K-0308":                        0.003 * RMB,
	"ERNIE-Tiny-8K":                             0.001 * RMB,
	"BLOOMZ-7B":                                 0.004 * RMB,
	"Embedding-V1":                              0.002 * RMB,
	"bge-large-zh":                              0.002 * RMB,
	"bge-large-en":                              0.002 * RMB,
	"tao-8k":                                    0.002 * RMB,
	"PaLM-2":                                    1,
	"gemini-1.5-pro-latest":                     1.25, // $3.5 / 1M tokens
	"gemini-1.5-flash-latest":                   0.075,
	"gemini-2.0-flash":                          0.05,
	"gemini-2.5-pro-exp-03-25":                  0.297, // RunAPI: 0.27 × 1.10
	"gemini-2.5-pro-preview-03-25":              0.297, // RunAPI: 0.27 × 1.10
	"gemini-2.5-pro":                            0.297, // RunAPI: 0.27 × 1.10
	"gemini-2.5-flash-preview-04-17":            0.075,
	"gemini-2.5-flash-preview-04-17-thinking":   0.075,
	"gemini-2.5-flash-preview-04-17-nothinking": 0.075,
	"gemini-2.5-flash-preview-05-20":            0.075,
	"gemini-2.5-flash-preview-05-20-thinking":   0.075,
	"gemini-2.5-flash-preview-05-20-nothinking": 0.075,
	"gemini-2.5-flash-thinking-*":               0.075, // 用于为后续所有2.5 flash thinking budget 模型设置默认倍率
	"gemini-2.5-pro-thinking-*":                 0.297, // RunAPI: 0.27 × 1.10 用于为后续所有2.5 pro thinking budget 模型设置默认倍率
	"gemini-2.5-flash-lite-preview-thinking-*":  0.05,
	"gemini-2.5-flash-lite-preview-06-17":       0.05,
	"gemini-2.5-flash":                          0.0707, // RunAPI: 0.0643 × 1.10
	"gemini-robotics-er-1.5-preview":            0.15,
	"gemini-embedding-001":                      0.075,
	"text-embedding-004":                        0.001,
	"chatglm_turbo":                             0.3572,     // ￥0.005 / 1k tokens
	"chatglm_pro":                               0.7143,     // ￥0.01 / 1k tokens
	"chatglm_std":                               0.3572,     // ￥0.005 / 1k tokens
	"chatglm_lite":                              0.1429,     // ￥0.002 / 1k tokens
	"glm-4":                                     7.143,      // ￥0.1 / 1k tokens
	"glm-4v":                                    0.05 * RMB, // ￥0.05 / 1k tokens
	"glm-4-alltools":                            0.1 * RMB,  // ￥0.1 / 1k tokens
	"glm-3-turbo":                               0.3572,
	"glm-4-plus":                                0.05 * RMB,
	"glm-4-0520":                                0.1 * RMB,
	"glm-4-air":                                 0.001 * RMB,
	"glm-4-airx":                                0.01 * RMB,
	"glm-4-long":                                0.001 * RMB,
	"glm-4-flash":                               0,
	"glm-4v-plus":                               0.01 * RMB,
	"qwen-turbo":                                0.077,  // RunAPI: 0.07 × 1.10
	"qwen-plus":                                 0.088,  // RunAPI: 0.08 × 1.10
	"text-embedding-v1":                         0.05,   // ￥0.0007 / 1k tokens
	"SparkDesk-v1.1":                            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v2.1":                            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.1":                            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v3.5":                            1.2858, // ￥0.018 / 1k tokens
	"SparkDesk-v4.0":                            1.2858,
	"360GPT_S2_V9":                              0.8572, // ¥0.012 / 1k tokens
	"360gpt-turbo":                              0.0858, // ¥0.0012 / 1k tokens
	"360gpt-turbo-responsibility-8k":            0.8572, // ¥0.012 / 1k tokens
	"360gpt-pro":                                0.8572, // ¥0.012 / 1k tokens
	"360gpt2-pro":                               0.8572, // ¥0.012 / 1k tokens
	"embedding-bert-512-v1":                     0.0715, // ¥0.001 / 1k tokens
	"embedding_s1_v1":                           0.0715, // ¥0.001 / 1k tokens
	"semantic_similarity_s1_v1":                 0.0715, // ¥0.001 / 1k tokens
	"hunyuan":                                   7.143,  // ¥0.1 / 1k tokens  // https://cloud.tencent.com/document/product/1729/97731#e0e6be58-60c8-469f-bdeb-6c264ce3b4d0
	// https://platform.lingyiwanwu.com/docs#-计费单元
	// 已经按照 7.2 来换算美元价格
	"yi-34b-chat-0205":       0.18,
	"yi-34b-chat-200k":       0.864,
	"yi-vl-plus":             0.432,
	"yi-large":               20.0 / 1000 * RMB,
	"yi-medium":              2.5 / 1000 * RMB,
	"yi-vision":              6.0 / 1000 * RMB,
	"yi-medium-200k":         12.0 / 1000 * RMB,
	"yi-spark":               1.0 / 1000 * RMB,
	"yi-large-rag":           25.0 / 1000 * RMB,
	"yi-large-turbo":         12.0 / 1000 * RMB,
	"yi-large-preview":       20.0 / 1000 * RMB,
	"yi-large-rag-preview":   25.0 / 1000 * RMB,
	"command":                0.5,
	"command-nightly":        0.5,
	"command-light":          0.5,
	"command-light-nightly":  0.5,
	"command-r":              0.25,
	"command-r-plus":         1.5,
	"command-r-08-2024":      0.075,
	"command-r-plus-08-2024": 1.25,
	"deepseek-chat":          0.099, // RunAPI: 0.09 × 1.10
	"deepseek-coder":         0.27 / 2,
	"deepseek-reasoner":      0.2013, // RunAPI: 0.183 × 1.10
	// Perplexity online 模型对搜索额外收费，有需要应自行调整，此处不计入搜索费用
	"llama-3-sonar-small-32k-chat":   0.11, // RunAPI: 0.1 × 1.10
	"llama-3-sonar-small-32k-online": 0.2 / 1000 * USD,
	"llama-3-sonar-large-32k-chat":   0.11, // RunAPI: 0.1 × 1.10
	"llama-3-sonar-large-32k-online": 1 / 1000 * USD,
	// grok
	"grok-3-beta":           1.5,
	"grok-3-mini-beta":      0.15,
	"grok-2":                1,
	"grok-2-vision":         1,
	"grok-beta":             2.5,
	"grok-vision-beta":      2.5,
	"grok-3-fast-beta":      2.5,
	"grok-3-mini-fast-beta": 0.3,
	// submodel
	"NousResearch/Hermes-4-405B-FP8":          0.8,
	"Qwen/Qwen3-235B-A22B-Thinking-2507":      0.6,
	"Qwen/Qwen3-Coder-480B-A35B-Instruct-FP8": 0.8,
	"Qwen/Qwen3-235B-A22B-Instruct-2507":      0.3,
	"zai-org/GLM-4.5-FP8":                     0.8,
	"openai/gpt-oss-120b":                     0.5,
	"deepseek-ai/DeepSeek-R1-0528":            0.8,
	"deepseek-ai/DeepSeek-R1":                 0.8,
	"deepseek-ai/DeepSeek-V3-0324":            0.8,
	"deepseek-ai/DeepSeek-V3.1":               0.8,
	// ===== RunAPI 新增模型（价格 = RunAPI × 1.10）=====
	// DeepSeek 系列
	"deepseek-v3":            0.1257, // RunAPI: 0.114285 × 1.10
	"deepseek-r1":            0.2514, // RunAPI: 0.228571 × 1.10
	"deepseek-v3-1":          0.2514, // RunAPI: 0.228571 × 1.10
	"deepseek-v3.1-thinking": 0.2013, // RunAPI: 0.183 × 1.10
	"deepseek-v3.2":          0.1257, // RunAPI: 0.114285 × 1.10
	"deepseek-v3.2-thinking": 0.099,  // RunAPI: 0.09 × 1.10
	"deepseek-v4-flash":      0.0629, // RunAPI: 0.057143 × 1.10
	"deepseek-v4-pro":        0.1941, // RunAPI: 0.17647 × 1.10
	// OpenAI 新增
	"gpt-5.1":                   0.275, // RunAPI: 0.25 × 1.10
	"gpt-5.2":                   0.385, // RunAPI: 0.35 × 1.10
	"gpt-5.2-pro":               4.62,  // RunAPI: 4.2 × 1.10
	"gpt-5.4":                   0.55,  // RunAPI: 0.5 × 1.10
	"gpt-5.4-mini":              0.165, // RunAPI: 0.15 × 1.10
	"gpt-5.4-nano":              0.044, // RunAPI: 0.04 × 1.10
	"gpt-5.4-pro":               6.6,   // RunAPI: 6 × 1.10
	"gpt-5.5":                   1.1,   // RunAPI: 1 × 1.10
	"gpt-5.6-luna":              0.044, // RunAPI: 0.04 × 1.10
	"gpt-5.6-sol":               1.1,   // RunAPI: 1 × 1.10
	"gpt-5.6-terra":             0.44,  // RunAPI: 0.4 × 1.10
	"gpt-5-pro":                 3.3,   // RunAPI: 3 × 1.10
	"gpt-oss-20b":               0.044, // RunAPI: 0.04 × 1.10
	"gpt-oss-120b":              0.242, // RunAPI: 0.22 × 1.10
	"gpt-4o-mini-audio-preview": 0.033, // RunAPI: 0.03 × 1.10
	// Claude 新增
	"claude-sonnet-5":                     0.66, // RunAPI: 0.6 × 1.10
	"claude-sonnet-4-6":                   0.99, // RunAPI: 0.9 × 1.10
	"claude-sonnet-4-6-thinking":          0.99, // RunAPI: 0.9 × 1.10
	"claude-sonnet-4-20250514-thinking":   0.99, // RunAPI: 0.9 × 1.10
	"claude-sonnet-4-5-20250929-thinking": 0.99, // RunAPI: 0.9 × 1.10
	"claude-opus-5":                       1.65, // RunAPI: 1.5 × 1.10
	"claude-fable-5":                      3.3,  // RunAPI: 3 × 1.10
	"claude-haiku-4-5-20251001-thinking":  0.33, // RunAPI: 0.3 × 1.10
	"claude-opus-4-20250514-thinking":     4.95, // RunAPI: 4.5 × 1.10
	"claude-opus-4-1-20250805-thinking":   4.95, // RunAPI: 4.5 × 1.10
	"claude-opus-4-5-20251101-thinking":   1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-6-thinking":            1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-7-thinking":            1.65, // RunAPI: 1.5 × 1.10
	"claude-opus-4-8-thinking":            1.65, // RunAPI: 1.5 × 1.10
	// Gemini 新增
	"gemini-3-flash-preview":        0.121,  // RunAPI: 0.11 × 1.10
	"gemini-3-pro-preview":          0.473,  // RunAPI: 0.43 × 1.10
	"gemini-3.1-pro-preview":        0.473,  // RunAPI: 0.43 × 1.10
	"gemini-3.1-flash-lite":         0.0594, // RunAPI: 0.054 × 1.10
	"gemini-3.1-flash-lite-preview": 0.0594, // RunAPI: 0.054 × 1.10
	"gemini-3.5-flash":              0.3465, // RunAPI: 0.315 × 1.10
	"gemini-3.5-flash-lite":         0.0713, // RunAPI: 0.0648 × 1.10
	"gemini-3.6-flash":              0.3465, // RunAPI: 0.315 × 1.10
	"gemini-2.5-flash-lite":         0.0237, // RunAPI: 0.0215 × 1.10
	// Grok 新增
	"grok-3":        0.495,  // RunAPI: 0.45 × 1.10
	"grok-4":        0.495,  // RunAPI: 0.45 × 1.10
	"grok-4.3":      0.2063, // RunAPI: 0.1875 × 1.10
	"grok-4.5":      0.33,   // RunAPI: 0.3 × 1.10
	"grok-4-fast":   0.033,  // RunAPI: 0.03 × 1.10
	"grok-4.1-fast": 0.033,  // RunAPI: 0.03 × 1.10
	// GLM 新增
	"glm-4.5":          0.1886, // RunAPI: 0.171428 × 1.10
	"glm-4.6":          0.1886, // RunAPI: 0.171428 × 1.10
	"glm-4.6-thinking": 0.2354, // RunAPI: 0.214 × 1.10
	"glm-4.7":          0.1886, // RunAPI: 0.171428 × 1.10
	"glm-5":            0.2514, // RunAPI: 0.228571 × 1.10
	"glm-5.1":          0.3771, // RunAPI: 0.342857 × 1.10
	"glm-5.2":          0.5029, // RunAPI: 0.457142 × 1.10
	// Kimi 新增
	"kimi-k2":          0.242,  // RunAPI: 0.22 × 1.10
	"kimi-k2-thinking": 0.2514, // RunAPI: 0.22857 × 1.10
	"kimi-k2.5":        0.2514, // RunAPI: 0.22857 × 1.10
	"kimi-k2.6":        0.4086, // RunAPI: 0.371428 × 1.10
	"kimi-k2.7-code":   0.4086, // RunAPI: 0.371428 × 1.10
	"kimi-k3":          1.2571, // RunAPI: 1.142857 × 1.10
	// Qwen 新增
	"qwen-max":                      22,     // RunAPI: 20 × 1.10
	"qwen3-max":                     0.1571, // RunAPI: 0.142857 × 1.10
	"qwen3.6-max":                   0.5657, // RunAPI: 0.514285 × 1.10
	"qwen3.6-plus":                  0.1257, // RunAPI: 0.114285 × 1.10
	"qwen3.6-flash":                 0.0754, // RunAPI: 0.068571 × 1.10
	"qwen3.5-flash":                 0.0126, // RunAPI: 0.011428 × 1.10
	"qwen3.5-plus":                  0.0503, // RunAPI: 0.045714 × 1.10
	"qwen3.5-27b":                   0.11,   // RunAPI: 0.1 × 1.10
	"qwen3.5-35b-a3b":               0.099,  // RunAPI: 0.09 × 1.10
	"qwen3.5-397b-a17b":             0.22,   // RunAPI: 0.2 × 1.10
	"qwen3.5-122b-a10b":             0.22,   // RunAPI: 0.2 × 1.10
	"qwen3.7-max":                   0.7543, // RunAPI: 0.685714 × 1.10
	"qwen3-8b":                      0.11,   // RunAPI: 0.1 × 1.10
	"qwen3-14b":                     0.0495, // RunAPI: 0.045 × 1.10
	"qwen3-30b-a3b":                 0.11,   // RunAPI: 0.1 × 1.10
	"qwen3-32b":                     0.0893, // RunAPI: 0.0812 × 1.10
	"qwen3-235b-a22b":               0.0893, // RunAPI: 0.0812 × 1.10
	"qwen3-235b-a22b-thinking-2507": 0.099,  // RunAPI: 0.09 × 1.10
	"qwen3-coder-plus":              0.1786, // RunAPI: 0.1624 × 1.10
	"qwen3-vl-plus":                 0.0629, // RunAPI: 0.057142 × 1.10
	"qwen3-vl-flash":                0.0094, // RunAPI: 0.008571 × 1.10
	"qwen3-vl-235b-a22b-instruct":   0.0893, // RunAPI: 0.0812 × 1.10
	"qwen3-vl-235b-a22b-thinking":   0.0893, // RunAPI: 0.0812 × 1.10
	// Doubao 新增
	"doubao-seed-1-6-251015":              0.055,  // RunAPI: 0.05 × 1.10
	"doubao-seed-1-6-flash-250828":        0.018,  // RunAPI: 0.0164 × 1.10
	"doubao-seed-1-6-thinking-250715":     0.055,  // RunAPI: 0.05 × 1.10
	"doubao-seed-1-6-vision-250815":       0.0963, // RunAPI: 0.0875 × 1.10
	"doubao-seed-1-8-251228":              0.0943, // RunAPI: 0.0857 × 1.10
	"doubao-seed-1-8-251228-thinking":     0.0943, // RunAPI: 0.0857 × 1.10
	"doubao-seed-2-0-lite-260215":         0.0707, // RunAPI: 0.0643 × 1.10
	"doubao-seed-2-0-mini-260215":         0.0235, // RunAPI: 0.0214 × 1.10
	"doubao-seed-2-0-pro-260215":          0.3772, // RunAPI: 0.3429 × 1.10
	"doubao-seed-2-0-code-preview-260215": 0.3772, // RunAPI: 0.3429 × 1.10
	// MiniMax 新增
	"MiniMax-M2.5": 0.132, // RunAPI: 0.12 × 1.10
	// 其他新增
	"codex-auto-review":                        1.1,    // RunAPI: 1 × 1.10
	"meta-llama/llama-3.2-3b-instruct":         0.0165, // RunAPI: 0.015 × 1.10
	"meta-llama/llama-3.1-405b":                2.75,   // RunAPI: 2.5 × 1.10
	"meta-llama/llama-3.1-405b-instruct":       3.3,    // RunAPI: 3 × 1.10
	"meta-llama/llama-3.1-70b-instruct":        0.275,  // RunAPI: 0.25 × 1.10
	"meta-llama/llama-3.1-8b-instruct":         0.0165, // RunAPI: 0.015 × 1.10
	"meta-llama/llama-3.2-1b-instruct":         0.0165, // RunAPI: 0.015 × 1.10
	"meta-llama/llama-3.2-11b-vision-instruct": 0.033,  // RunAPI: 0.03 × 1.10
	"meta-llama/llama-3.2-90b-vision-instruct": 0.22,   // RunAPI: 0.2 × 1.10
	"meta-llama/llama-3.3-70b-instruct":        0.66,   // RunAPI: 0.6 × 1.10
	"meta-llama/llama-3-70b-instruct":          0.22,   // RunAPI: 0.2 × 1.10
	"meta-llama/llama-3-8b-instruct":           0.022,  // RunAPI: 0.02 × 1.10
	"meta-llama/llama-4-scout":                 0.066,  // RunAPI: 0.06 × 1.10
	"meta-llama/llama-4-maverick":              0.11,   // RunAPI: 0.1 × 1.10
	"meta-llama/llama-guard-4-12b":             0.11,   // RunAPI: 0.1 × 1.10
	"qwen2.5-7b-instruct":                      0.022,  // RunAPI: 0.02 × 1.10
	// Embedding 新增
	"jina-embeddings-v4":            0.033, // RunAPI: 0.03 × 1.10
	"jina-embeddings-v5-text-nano":  0.033, // RunAPI: 0.03 × 1.10
	"jina-embeddings-v5-text-small": 0.033, // RunAPI: 0.03 × 1.10
	"jina-code-embeddings-1.5b":     0.033, // RunAPI: 0.03 × 1.10
	"jina-code-embeddings-0.5b":     0.033, // RunAPI: 0.03 × 1.10
	// Rerank 新增
	"rerank-english-v3.0":      0.011, // RunAPI: 0.01 × 1.10
	"rerank-multilingual-v3.0": 0.011, // RunAPI: 0.01 × 1.10
	"rerank-v4.0-pro":          0.011, // RunAPI: 0.01 × 1.10
	// 图片生成（按量计费型）
	"gemini-3.1-flash-image-preview": 0.0811, // RunAPI: 0.073746 × 1.10
	"gemini-3.1-flash-lite-image":    0.0404, // RunAPI: 0.036764 × 1.10
	"gemini-3-pro-image-preview":     0.3245, // RunAPI: 0.294985 × 1.10
}

var defaultModelPrice = map[string]float64{
	"suno_music":                     0.088,  // RunAPI: 0.08 × 1.10
	"suno_lyrics":                    0.0033, // RunAPI: 0.003 × 1.10
	"dall-e-3":                       0.022,  // RunAPI: 0.02 × 1.10
	"imagen-3.0-generate-002":        0.03,
	"black-forest-labs/flux-1.1-pro": 0.04,
	"gpt-4-gizmo-*":                  0.1,
	"mj_video":                       0.8,
	"mj_imagine":                     0.1,
	"mj_edits":                       0.1,
	"mj_variation":                   0.1,
	"mj_reroll":                      0.1,
	"mj_blend":                       0.1,
	"mj_modal":                       0.1,
	"mj_zoom":                        0.1,
	"mj_shorten":                     0.1,
	"mj_high_variation":              0.1,
	"mj_low_variation":               0.1,
	"mj_pan":                         0.1,
	"mj_inpaint":                     0,
	"mj_custom_zoom":                 0,
	"mj_describe":                    0.05,
	"mj_upscale":                     0.05,
	"swap_face":                      0.05,
	"mj_upload":                      0.05,
	"sora-2":                         0.3,
	"sora-2-pro":                     0.5,
	"gpt-4o-mini-tts":                0.473, // RunAPI: 0.43 × 1.10
	"veo-3.0-generate-001":           0.4,
	"veo-3.0-fast-generate-001":      0.15,
	"veo-3.1-generate-preview":       0.4,
	"veo-3.1-fast-generate-preview":  0.15,
	// ===== RunAPI 新增按次计费模型（价格 = RunAPI × 1.10）=====
	"gpt-image-1":                    0.0157,  // RunAPI: 0.0143 × 1.10
	"gpt-image-2":                    0.0308,  // RunAPI: 0.028 × 1.10
	"doubao-seedream-4-5-251128":     0.0308,  // RunAPI: 0.028 × 1.10
	"doubao-seedream-4-0-250828":     0.0275,  // RunAPI: 0.025 × 1.10
	"doubao-seedream-5-0-260128":     0.033,   // RunAPI: 0.03 × 1.10
	"doubao-seedream-3-0-t2i-250415": 0.0231,  // RunAPI: 0.021 × 1.10
	"doubao-seededit-3-0-i2i-250628": 0.0231,  // RunAPI: 0.021 × 1.10
	"qwen-image-edit-2509":           0.022,   // RunAPI: 0.02 × 1.10
	"flux.1.1-pro":                   0.0264,  // RunAPI: 0.024 × 1.10
	"flux-pro":                       0.0308,  // RunAPI: 0.028 × 1.10
	"flux-pro-max":                   0.0528,  // RunAPI: 0.048 × 1.10
	"flux-pro-1.1-ultra":             0.0418,  // RunAPI: 0.038 × 1.10
	"flux-kontext-max":               0.0528,  // RunAPI: 0.048 × 1.10
	"flux-kontext-pro":               0.0308,  // RunAPI: 0.028 × 1.10
	"gemini-2.5-flash-image":         0.022,   // RunAPI: 0.02 × 1.10
	"grok-imagine-image":             0.033,   // RunAPI: 0.03 × 1.10
	"grok-imagine-image-pro":         0.0968,  // RunAPI: 0.088 × 1.10
	"grok-imagine-video-1.5-fast":    0.0825,  // RunAPI: 0.075 × 1.10
	"grok-imagine-1.0-video":         0.0825,  // RunAPI: 0.075 × 1.10
	"veo3.1":                         0.495,   // RunAPI: 0.45 × 1.10
	"veo3.1-i2v":                     0.495,   // RunAPI: 0.45 × 1.10
	"veo3.1-fast":                    0.33,    // RunAPI: 0.3 × 1.10
	"kling-v1":                       1.32,    // RunAPI: 1.2 × 1.10
	"kling-v1-5":                     1.32,    // RunAPI: 1.2 × 1.10
	"kling-v1-6":                     1.32,    // RunAPI: 1.2 × 1.10
	"kling-v2-5-turbo":               0.99,    // RunAPI: 0.9 × 1.10
	"MiniMax-Hailuo-2.3":             0.517,   // RunAPI: 0.47 × 1.10
	"MiniMax-Hailuo-2.3-Fast":        0.286,   // RunAPI: 0.26 × 1.10
	"sora-2-characters":              0.00165, // RunAPI: 0.0015 × 1.10
}

var defaultAudioRatio = map[string]float64{
	"gpt-4o-audio-preview":         16,
	"gpt-4o-mini-audio-preview":    66.67,
	"gpt-4o-realtime-preview":      8,
	"gpt-4o-mini-realtime-preview": 16.67,
	"gpt-4o-mini-tts":              25,
}

var defaultAudioCompletionRatio = map[string]float64{
	"gpt-4o-realtime":      2,
	"gpt-4o-mini-realtime": 2,
	"gpt-4o-mini-tts":      1,
	"tts-1":                0,
	"tts-1-hd":             0,
	"tts-1-1106":           0,
	"tts-1-hd-1106":        0,
}

var modelPriceMap = types.NewRWMap[string, float64]()
var modelRatioMap = types.NewRWMap[string, float64]()
var completionRatioMap = types.NewRWMap[string, float64]()

var defaultCompletionRatio = map[string]float64{
	"gpt-4-gizmo-*":  2,
	"gpt-4o-gizmo-*": 3,
	"gpt-4-all":      2,
	// RunAPI 新增模型补全倍率
	"deepseek-v3":                         4,      // RunAPI cr=4
	"deepseek-r1":                         4,      // RunAPI cr=4
	"deepseek-v3-1":                       3,      // RunAPI cr=3
	"deepseek-v3.1-thinking":              3,      // RunAPI cr=3
	"deepseek-v3.2":                       1.5,    // RunAPI cr=1.5
	"deepseek-v3.2-thinking":              1.5,    // RunAPI cr=1.5
	"deepseek-v4-flash":                   2,      // RunAPI cr=2
	"deepseek-v4-pro":                     2,      // RunAPI cr=2
	"kimi-k2":                             4,      // RunAPI cr=4
	"kimi-k2-thinking":                    4,      // RunAPI cr=4
	"kimi-k2.5":                           5.25,   // RunAPI cr=5.25
	"kimi-k2.6":                           4.1538, // RunAPI cr=4.153846
	"kimi-k2.7-code":                      5,      // RunAPI cr=5
	"kimi-k3":                             5,      // RunAPI cr=5
	"glm-4.5":                             4,      // RunAPI cr=4
	"glm-4.6":                             4,      // RunAPI cr=4
	"glm-4.6-thinking":                    4,      // RunAPI cr=4
	"glm-4.7":                             4,      // RunAPI cr=4
	"glm-5":                               4,      // RunAPI cr=4
	"glm-5.1":                             4,      // RunAPI cr=4
	"glm-5.2":                             3.5,    // RunAPI cr=3.5
	"grok-3":                              5,      // RunAPI cr=5
	"grok-4":                              5,      // RunAPI cr=5
	"grok-4.3":                            2,      // RunAPI cr=2
	"grok-4.5":                            3,      // RunAPI cr=3
	"grok-4-fast":                         2.5,    // RunAPI cr=2.5
	"grok-4.1-fast":                       7.5,    // RunAPI cr=7.5
	"codex-auto-review":                   6,      // RunAPI cr=6
	"qwen-max":                            1,      // RunAPI cr=1
	"qwen3-max":                           4,      // RunAPI cr=4
	"qwen3.6-max":                         6,      // RunAPI cr=6
	"qwen3.6-plus":                        6,      // RunAPI cr=6
	"qwen3.6-flash":                       6,      // RunAPI cr=6
	"qwen3.5-flash":                       10,     // RunAPI cr=10
	"qwen3.5-plus":                        6,      // RunAPI cr=6
	"qwen3.5-27b":                         8,      // RunAPI cr=8
	"qwen3.5-35b-a3b":                     8,      // RunAPI cr=8
	"qwen3.5-397b-a17b":                   6,      // RunAPI cr=6
	"qwen3.5-122b-a10b":                   8,      // RunAPI cr=8
	"qwen3.7-max":                         3,      // RunAPI cr=3
	"qwen3-8b":                            4,      // RunAPI cr=4
	"qwen3-14b":                           4,      // RunAPI cr=4
	"qwen3-30b-a3b":                       4,      // RunAPI cr=4
	"qwen3-32b":                           4,      // RunAPI cr=4
	"qwen3-235b-a22b":                     4,      // RunAPI cr=4
	"qwen3-235b-a22b-thinking-2507":       4,      // RunAPI cr=4
	"qwen3-coder-plus":                    4,      // RunAPI cr=4
	"qwen3-vl-plus":                       10,     // RunAPI cr=10
	"qwen3-vl-flash":                      10,     // RunAPI cr=10
	"qwen3-vl-235b-a22b-instruct":         4,      // RunAPI cr=4
	"qwen3-vl-235b-a22b-thinking":         10,     // RunAPI cr=10
	"doubao-seed-1-6-251015":              10,     // RunAPI cr=10
	"doubao-seed-1-6-flash-250828":        1,      // RunAPI cr=1
	"doubao-seed-1-6-thinking-250715":     10,     // RunAPI cr=10
	"doubao-seed-1-6-vision-250815":       8,      // RunAPI cr=8
	"doubao-seed-1-8-251228":              10,     // RunAPI cr=10
	"doubao-seed-1-8-251228-thinking":     10,     // RunAPI cr=10
	"doubao-seed-2-0-lite-260215":         6,      // RunAPI cr=6
	"doubao-seed-2-0-mini-260215":         10,     // RunAPI cr=10
	"doubao-seed-2-0-pro-260215":          5,      // RunAPI cr=5
	"doubao-seed-2-0-code-preview-260215": 5,      // RunAPI cr=5
	"MiniMax-M2.5":                        4,      // RunAPI cr=4
	"claude-fable-5":                      5,      // RunAPI cr=5
	"claude-sonnet-5":                     5,      // RunAPI cr=5
	"claude-opus-5":                       5,      // RunAPI cr=5
	"gemini-3.1-flash-image-preview":      120,    // RunAPI cr=120
	"gemini-3.1-flash-lite-image":         120,    // RunAPI cr=120
	"gemini-3-pro-image-preview":          60,     // RunAPI cr=60
}

// InitRatioSettings initializes all model related settings maps
func InitRatioSettings() {
	modelPriceMap.AddAll(defaultModelPrice)
	modelRatioMap.AddAll(defaultModelRatio)
	completionRatioMap.AddAll(defaultCompletionRatio)
	cacheRatioMap.AddAll(defaultCacheRatio)
	createCacheRatioMap.AddAll(defaultCreateCacheRatio)
	imageRatioMap.AddAll(defaultImageRatio)
	audioRatioMap.AddAll(defaultAudioRatio)
	audioCompletionRatioMap.AddAll(defaultAudioCompletionRatio)
}

func GetModelPriceMap() map[string]float64 {
	return modelPriceMap.ReadAll()
}

func ModelPrice2JSONString() string {
	return modelPriceMap.MarshalJSONString()
}

func UpdateModelPriceByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelPriceMap, jsonStr, InvalidateExposedDataCache)
}

// GetModelPrice 返回模型的价格，如果模型不存在则返回-1，false
func GetModelPrice(name string, printErr bool) (float64, bool) {
	name = FormatMatchingModelName(name)

	if price, ok := modelPriceMap.Get(name); ok {
		return price, true
	}

	if strings.HasSuffix(name, CompactModelSuffix) {
		price, ok := modelPriceMap.Get(CompactWildcardModelKey)
		if !ok {
			if printErr {
				common.SysError("model price not found: " + name)
			}
			return -1, false
		}
		return price, true
	}

	if printErr {
		common.SysError("model price not found: " + name)
	}
	return -1, false
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(modelRatioMap, jsonStr, InvalidateExposedDataCache)
}

// 处理带有思考预算的模型名称，方便统一定价
func handleThinkingBudgetModel(name, prefix, wildcard string) string {
	if strings.HasPrefix(name, prefix) && strings.Contains(name, "-thinking-") {
		return wildcard
	}
	return name
}

func GetModelRatio(name string) (float64, bool, string) {
	name = FormatMatchingModelName(name)

	ratio, ok := modelRatioMap.Get(name)
	if !ok {
		if strings.HasSuffix(name, CompactModelSuffix) {
			if wildcardRatio, ok := modelRatioMap.Get(CompactWildcardModelKey); ok {
				return wildcardRatio, true, name
			}
			//return 0, true, name
		}
		return 37.5, operation_setting.SelfUseModeEnabled, name
	}
	return ratio, true, name
}

func DefaultModelRatio2JSONString() string {
	jsonBytes, err := common.Marshal(defaultModelRatio)
	if err != nil {
		common.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func GetDefaultModelRatioMap() map[string]float64 {
	return defaultModelRatio
}

func GetDefaultModelPriceMap() map[string]float64 {
	return defaultModelPrice
}

func CompletionRatio2JSONString() string {
	return completionRatioMap.MarshalJSONString()
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(completionRatioMap, jsonStr, InvalidateExposedDataCache)
}

func GetCompletionRatio(name string) float64 {
	name = FormatMatchingModelName(name)

	if strings.Contains(name, "/") {
		if ratio, ok := completionRatioMap.Get(name); ok {
			return ratio
		}
	}
	hardCodedRatio, contain := getHardcodedCompletionModelRatio(name)
	if contain {
		return hardCodedRatio
	}
	if ratio, ok := completionRatioMap.Get(name); ok {
		return ratio
	}
	return hardCodedRatio
}

type CompletionRatioInfo struct {
	Ratio  float64 `json:"ratio"`
	Locked bool    `json:"locked"`
}

func GetCompletionRatioInfo(name string) CompletionRatioInfo {
	name = FormatMatchingModelName(name)

	if strings.Contains(name, "/") {
		if ratio, ok := completionRatioMap.Get(name); ok {
			return CompletionRatioInfo{
				Ratio:  ratio,
				Locked: false,
			}
		}
	}

	hardCodedRatio, locked := getHardcodedCompletionModelRatio(name)
	if locked {
		return CompletionRatioInfo{
			Ratio:  hardCodedRatio,
			Locked: true,
		}
	}

	if ratio, ok := completionRatioMap.Get(name); ok {
		return CompletionRatioInfo{
			Ratio:  ratio,
			Locked: false,
		}
	}

	return CompletionRatioInfo{
		Ratio:  hardCodedRatio,
		Locked: false,
	}
}

func getHardcodedCompletionModelRatio(name string) (float64, bool) {

	isReservedModel := strings.HasSuffix(name, "-all") || strings.HasSuffix(name, "-gizmo-*")
	if isReservedModel {
		return 2, false
	}

	if strings.HasPrefix(name, "gpt-") {
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 3, true
			}
			if strings.HasPrefix(name, "gpt-4o-mini-tts") {
				return 20, false
			}
			return 4, false
		}
		// gpt-5 匹配
		if strings.HasPrefix(name, "gpt-5") {
			if strings.HasPrefix(name, "gpt-5.5") {
				return 6, true
			}
			if strings.HasPrefix(name, "gpt-5.4") {
				if strings.HasPrefix(name, "gpt-5.4-nano") {
					return 6, true
				}
				return 6, true
			}
			if strings.HasPrefix(name, "gpt-5.6") {
				return 6, true
			}
			return 8, true
		}
		// gpt-4.5-preview匹配
		if strings.HasPrefix(name, "gpt-4.5-preview") {
			return 2, true
		}
		if strings.HasPrefix(name, "gpt-4-turbo") || strings.HasSuffix(name, "gpt-4-1106") || strings.HasSuffix(name, "gpt-4-1105") {
			return 3, true
		}
		// 没有特殊标记的 gpt-4 模型默认倍率为 2
		return 2, false
	}
	if strings.HasPrefix(name, "o1") || strings.HasPrefix(name, "o3") || strings.HasPrefix(name, "o4") {
		return 4, true
	}
	if name == "chatgpt-4o-latest" {
		return 3, true
	}

	if strings.Contains(name, "claude-3") {
		return 5, true
	} else if strings.Contains(name, "claude-sonnet-4") || strings.Contains(name, "claude-opus-4") || strings.Contains(name, "claude-haiku-4") || strings.Contains(name, "claude-sonnet-5") || strings.Contains(name, "claude-opus-5") || strings.Contains(name, "claude-fable-5") {
		return 5, true
	}

	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3, true
		}
		if strings.HasSuffix(name, "1106") {
			return 2, true
		}
		return 4.0 / 3.0, true
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3, true
	}
	if strings.HasPrefix(name, "gemini-") {
		if strings.HasPrefix(name, "gemini-1.5") {
			return 4, true
		} else if strings.HasPrefix(name, "gemini-2.0") {
			return 4, true
		} else if strings.HasPrefix(name, "gemini-2.5-pro") { // 移除preview来增加兼容性，这里假设正式版的倍率和preview一致
			return 8, false
		} else if strings.HasPrefix(name, "gemini-2.5-flash") { // 处理不同的flash模型倍率
			if strings.HasPrefix(name, "gemini-2.5-flash-preview") {
				if strings.HasSuffix(name, "-nothinking") {
					return 4, false
				}
				return 3.5 / 0.15, false
			}
			if strings.HasPrefix(name, "gemini-2.5-flash-lite") {
				return 4, false
			}
			return 2.5 / 0.3, false
		} else if strings.HasPrefix(name, "gemini-robotics-er-1.5") {
			return 2.5 / 0.3, false
		} else if strings.HasPrefix(name, "gemini-3-pro") {
			if strings.HasPrefix(name, "gemini-3-pro-image") {
				return 60, false
			}
			return 6, false
		}
		return 4, false
	}
	if strings.HasPrefix(name, "command") {
		switch name {
		case "command-r":
			return 3, true
		case "command-r-plus":
			return 5, true
		case "command-r-08-2024":
			return 4, true
		case "command-r-plus-08-2024":
			return 4, true
		default:
			return 4, false
		}
	}
	// hint 只给官方上4倍率，由于开源模型供应商自行定价，不对其进行补全倍率进行强制对齐
	if strings.HasPrefix(name, "ERNIE-Speed-") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Lite-") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Character") {
		return 2, true
	} else if strings.HasPrefix(name, "ERNIE-Functions") {
		return 2, true
	}
	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64, true
	case "llama3-8b-8192":
		return 2, true
	case "llama3-70b-8192":
		return 0.79 / 0.59, true
	}
	return 1, false
}

func GetAudioRatio(name string) float64 {
	name = FormatMatchingModelName(name)
	if ratio, ok := audioRatioMap.Get(name); ok {
		return ratio
	}
	return 1
}

func GetAudioCompletionRatio(name string) float64 {
	name = FormatMatchingModelName(name)
	if ratio, ok := audioCompletionRatioMap.Get(name); ok {
		return ratio
	}
	return 1
}

func ContainsAudioRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := audioRatioMap.Get(name)
	return ok
}

func ContainsAudioCompletionRatio(name string) bool {
	name = FormatMatchingModelName(name)
	_, ok := audioCompletionRatioMap.Get(name)
	return ok
}

func ModelRatio2JSONString() string {
	return modelRatioMap.MarshalJSONString()
}

var defaultImageRatio = map[string]float64{}
var imageRatioMap = types.NewRWMap[string, float64]()
var audioRatioMap = types.NewRWMap[string, float64]()
var audioCompletionRatioMap = types.NewRWMap[string, float64]()

func ImageRatio2JSONString() string {
	return imageRatioMap.MarshalJSONString()
}

func UpdateImageRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonString(imageRatioMap, jsonStr)
}

func GetImageRatio(name string) (float64, bool) {
	ratio, ok := imageRatioMap.Get(name)
	if !ok {
		return 1, false // Default to 1 if not found
	}
	return ratio, true
}

func AudioRatio2JSONString() string {
	return audioRatioMap.MarshalJSONString()
}

func UpdateAudioRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(audioRatioMap, jsonStr, InvalidateExposedDataCache)
}

func AudioCompletionRatio2JSONString() string {
	return audioCompletionRatioMap.MarshalJSONString()
}

func UpdateAudioCompletionRatioByJSONString(jsonStr string) error {
	return types.LoadFromJsonStringWithCallback(audioCompletionRatioMap, jsonStr, InvalidateExposedDataCache)
}

func GetModelRatioCopy() map[string]float64 {
	return modelRatioMap.ReadAll()
}

func GetModelPriceCopy() map[string]float64 {
	return modelPriceMap.ReadAll()
}

func GetCompletionRatioCopy() map[string]float64 {
	return completionRatioMap.ReadAll()
}

func GetImageRatioCopy() map[string]float64 {
	return imageRatioMap.ReadAll()
}

func GetAudioRatioCopy() map[string]float64 {
	return audioRatioMap.ReadAll()
}

func GetAudioCompletionRatioCopy() map[string]float64 {
	return audioCompletionRatioMap.ReadAll()
}

// 转换模型名，减少渠道必须配置各种带参数模型
func FormatMatchingModelName(name string) string {

	if strings.HasPrefix(name, "gemini-2.5-flash-lite") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-flash-lite", "gemini-2.5-flash-lite-thinking-*")
	} else if strings.HasPrefix(name, "gemini-2.5-flash") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-flash", "gemini-2.5-flash-thinking-*")
	} else if strings.HasPrefix(name, "gemini-2.5-pro") {
		name = handleThinkingBudgetModel(name, "gemini-2.5-pro", "gemini-2.5-pro-thinking-*")
	}

	if strings.HasPrefix(name, "gpt-4-gizmo") {
		name = "gpt-4-gizmo-*"
	}
	if strings.HasPrefix(name, "gpt-4o-gizmo") {
		name = "gpt-4o-gizmo-*"
	}
	return name
}

// result: 倍率or价格， usePrice， exist
func GetModelRatioOrPrice(model string) (float64, bool, bool) { // price or ratio
	price, usePrice := GetModelPrice(model, false)
	if usePrice {
		return price, true, true
	}
	modelRatio, success, _ := GetModelRatio(model)
	if success {
		return modelRatio, false, true
	}
	return 37.5, false, false
}
