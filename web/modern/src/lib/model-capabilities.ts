// Model capability detection based on model names (simplified version of AWS adapter logic)

export interface ModelCapabilities {
  supportsTools: boolean
  supportsThinking: boolean
  supportsStop: boolean
  supportsReasoningEffort: boolean
  supportsLogprobs: boolean
  supportsTopK: boolean
  supportsTopP: boolean
  supportsFrequencyPenalty: boolean
  supportsPresencePenalty: boolean
  supportsMaxCompletionTokens: boolean
  supportsVision: boolean
}

// Check if model is AWS OpenAI OSS model
const isOpenAIOSSModel = (modelName: string): boolean => {
  const lowerName = modelName.toLowerCase()
  return lowerName.includes('gpt-oss-20b') || lowerName.includes('gpt-oss-120b')
}

// Model type detection helper
const getModelType = (modelName: string): string => {
  const lowerName = modelName.toLowerCase()

  if (lowerName.includes('claude')) return 'claude'
  if (lowerName.includes('cohere') || lowerName.includes('command')) return 'cohere'
  // Check for Vercel AI Gateway hosted models - these should be handled by vercel type
  // since they have specific capabilities but are hosted through Vercel
  if (lowerName.includes('alibaba/')) return 'vercel'
  // Check for DeepInfra hosted models - these should be handled by deepinfra type
  // since they have similar capabilities but are hosted through DeepInfra
  if (lowerName.includes('deepseek-ai/') ||
    lowerName.includes('qwen/') ||
    lowerName.includes('moonshotai/') ||
    lowerName.includes('baai/') ||
    lowerName.includes('nvidia/')) return 'deepinfra'
  // Check for Hyperbolic hosted models - these should be handled by hyperbolic type
  // since they have similar capabilities but are hosted through Hyperbolic
  if (lowerName.includes('openai/gpt-oss') ||
    lowerName.includes('qwen/qwen3-next') ||
    lowerName.includes('qwen/qwen3-235b') ||
    lowerName.includes('qwen/qwq') ||
    lowerName.includes('deepseek-ai/deepseek-r1') ||
    lowerName.includes('deepseek-ai/deepseek-v3') ||
    lowerName.includes('moonshotai/kimi-k2')) return 'hyperbolic'
  if (lowerName.includes('deepseek')) return 'deepseek'
  if (lowerName.includes('llama')) return 'llama'
  if (lowerName.includes('mistral') || lowerName.includes('mixtral')) return 'mistral'
  if (lowerName.includes('nova')) return 'nova'
  if (lowerName.includes('palmyra')) return 'writer'
  // Check for AWS OpenAI OSS models first, before general GPT check
  if (isOpenAIOSSModel(modelName)) return 'openai-oss'
  if (lowerName.includes('gpt')) return 'openai'
  if (lowerName.includes('gemini')) return 'google'

  return 'unknown'
}

// Check if Claude model supports extended thinking
const claudeSupportsThinking = (modelName: string): boolean => {
  const lowerName = modelName.toLowerCase()

  // Supported models according to the issue and documentation:
  // - Claude Opus 4.1 (claude-opus-4-1-20250805)
  // - Claude Opus 4 (claude-opus-4-20250514) 
  // - Claude Sonnet 4 (claude-sonnet-4-20250514)
  // - Claude Sonnet 3.7 (claude-3-7-sonnet-20250219)

  return (
    lowerName.includes('claude-opus-4-1-20250805') ||
    lowerName.includes('claude-opus-4-20250514') ||
    lowerName.includes('claude-sonnet-4-20250514') ||
    lowerName.includes('claude-3-7-sonnet-20250219') ||
    // More flexible patterns for model variations
    (lowerName.includes('claude') && lowerName.includes('opus') && lowerName.includes('4')) ||
    (lowerName.includes('claude') && lowerName.includes('sonnet') && lowerName.includes('4')) ||
    (lowerName.includes('claude') && lowerName.includes('sonnet') && lowerName.includes('3.7'))
  )
}

// Check if Hyperbolic model supports thinking
const hyperbolicSupportsThinking = (modelName: string): boolean => {
  const lowerName = modelName.toLowerCase()

  // Hyperbolic models that support thinking based on model names:
  // - Qwen3-Next-80B-A3B-Thinking (explicitly has "thinking" in name)
  // - DeepSeek-R1 models (reasoning models)

  return (
    lowerName.includes('thinking') ||
    lowerName.includes('deepseek-r1') ||
    lowerName.includes('qwen3-next-80b-a3b-thinking')
  )
}

// Check if DeepSeek model supports reasoning effort
const deepseekSupportsReasoningEffort = (modelName: string): boolean => {
  const lowerName = modelName.toLowerCase()

  // DeepSeek models that support reasoning_effort parameter:
  // - deepseek-v3.1: Supports reasoning effort

  return (
    lowerName.includes('deepseek-v3.1')
  )
}

// Check if model supports vision/image input
const modelSupportsVision = (modelName: string): boolean => {
  const lowerName = modelName.toLowerCase()

  // Claude models - Most Claude 3+ models support vision
  if (lowerName.includes('claude')) {
    return lowerName.includes('claude-3') ||
      lowerName.includes('claude-4') ||
      lowerName.includes('sonnet') ||
      lowerName.includes('opus') ||
      lowerName.includes('haiku')
  }

  // OpenAI GPT models with vision
  if (lowerName.includes('gpt') || lowerName.includes('chatgpt')) {
    return (lowerName.includes('gpt-4') &&
      (lowerName.includes('vision') ||
        lowerName.includes('gpt-4o') ||
        lowerName.includes('gpt-4-turbo') ||
        !lowerName.includes('gpt-4-0613') && !lowerName.includes('gpt-4-0314'))) || // Exclude older GPT-4 versions without vision
      lowerName.includes('chatgpt-4o') || // chatgpt-4o-latest and similar
      lowerName.includes('gpt-5') // gpt-5-chat-latest and similar
  }

  // Google Gemini models - Most support vision
  if (lowerName.includes('gemini')) {
    return lowerName.includes('pro') ||
      lowerName.includes('ultra') ||
      lowerName.includes('flash') ||
      lowerName.includes('exp')
  }

  // AWS Nova models - Support vision
  if (lowerName.includes('nova')) {
    return lowerName.includes('lite') ||
      lowerName.includes('pro') ||
      lowerName.includes('micro') ||
      lowerName.includes('premier')
  }

  // Llama4 vision models - New models with vision support
  if (lowerName.includes('llama4')) {
    return lowerName.includes('llama4-maverick-17b-1m') ||
      lowerName.includes('llama4-scout-17b-3.5m')
  }

  // Other vision models
  return lowerName.includes('vision') ||
    lowerName.includes('pixtral') || // Mistral vision model
    lowerName.includes('llava') ||   // LLaVA vision models
    lowerName.includes('cogvlm')     // CogVLM vision models
}

export const getModelCapabilities = (modelName: string): ModelCapabilities => {
  if (!modelName) return getDefaultCapabilities()

  const modelType = getModelType(modelName)

  switch (modelType) {
    case 'claude':
      return {
        supportsTools: true,
        supportsThinking: claudeSupportsThinking(modelName),
        supportsStop: false,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'cohere':
      return {
        supportsTools: true,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: true,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'deepseek':
      return {
        supportsTools: false,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: deepseekSupportsReasoningEffort(modelName),
        supportsLogprobs: false,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'llama':
      return {
        supportsTools: false,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: true,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'mistral':
      return {
        supportsTools: false,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: true,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'nova':
      return {
        supportsTools: false,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: false,
        supportsPresencePenalty: false,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'openai-oss':
      return {
        supportsTools: false,
        supportsThinking: false,
        supportsStop: false,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: false,
        supportsPresencePenalty: false,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'openai':
      return {
        supportsTools: true,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: true,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: true,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'deepinfra':
      return {
        supportsTools: true,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: deepseekSupportsReasoningEffort(modelName),
        supportsLogprobs: true,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: true,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'vercel':
      return {
        supportsTools: true,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: true,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: true,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'hyperbolic':
      return {
        supportsTools: true,
        supportsThinking: hyperbolicSupportsThinking(modelName),
        supportsStop: true,
        supportsReasoningEffort: deepseekSupportsReasoningEffort(modelName),
        supportsLogprobs: true,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: true,
        supportsPresencePenalty: true,
        supportsMaxCompletionTokens: true,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'google':
      return {
        supportsTools: true,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: true,
        supportsTopP: true,
        supportsFrequencyPenalty: false,
        supportsPresencePenalty: false,
        supportsMaxCompletionTokens: true,
        supportsVision: modelSupportsVision(modelName),
      }

    case 'writer':
      return {
        supportsTools: false,
        supportsThinking: false,
        supportsStop: true,
        supportsReasoningEffort: false,
        supportsLogprobs: false,
        supportsTopK: false,
        supportsTopP: true,
        supportsFrequencyPenalty: false,
        supportsPresencePenalty: false,
        supportsMaxCompletionTokens: false,
        supportsVision: modelSupportsVision(modelName),
      }

    default:
      return getDefaultCapabilities()
  }
}

const getDefaultCapabilities = (): ModelCapabilities => ({
  supportsTools: false,
  supportsThinking: false,
  supportsStop: false,
  supportsReasoningEffort: false,
  supportsLogprobs: false,
  supportsTopK: false,
  supportsTopP: true,
  supportsFrequencyPenalty: false,
  supportsPresencePenalty: false,
  supportsMaxCompletionTokens: false,
  supportsVision: false,
})
