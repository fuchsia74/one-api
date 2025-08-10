import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import { Separator } from '@/components/ui/separator'
import { AlertCircle, Info } from 'lucide-react'
import { api } from '@/lib/api'

// Enhanced channel schema with comprehensive validation
const channelSchema = z.object({
  name: z.string().min(1, 'Channel name is required'),
  // Coerce because Select returns string
  type: z.coerce.number().int().min(1, 'Channel type is required'),
  // key optional on edit; we enforce presence only on create in submit handler
  key: z.string().optional(),
  base_url: z.string().optional(),
  other: z.string().optional(),
  models: z.array(z.string()).default([]),
  model_mapping: z.string().optional(),
  model_configs: z.string().optional(),
  system_prompt: z.string().optional(),
  groups: z.array(z.string()).default(['default']),
  priority: z.number().default(0),
  weight: z.number().default(0),
  ratelimit: z.number().min(0).default(0),
  // AWS and Vertex AI specific config
  config: z.object({
    region: z.string().optional(),
    ak: z.string().optional(),
    sk: z.string().optional(),
    user_id: z.string().optional(),
    vertex_ai_project_id: z.string().optional(),
    vertex_ai_adc: z.string().optional(),
    auth_type: z.string().default('personal_access_token'),
  }).default({}),
  inference_profile_arn_map: z.string().optional(),
})

type ChannelForm = z.infer<typeof channelSchema>

interface ChannelType {
  key: number
  text: string
  value: number
  color?: string
  tip?: string
  description?: string
}

interface Model {
  id: string
  name: string
}

// Comprehensive channel types with colors and descriptions
const CHANNEL_TYPES: ChannelType[] = [
  { key: 1, text: 'OpenAI', value: 1, color: 'green' },
  { key: 50, text: 'OpenAI Compatible', value: 50, color: 'olive', description: 'OpenAI compatible channel, supports custom Base URL' },
  { key: 14, text: 'Anthropic', value: 14, color: 'black' },
  { key: 33, text: 'AWS', value: 33, color: 'black' },
  { key: 3, text: 'Azure', value: 3, color: 'olive' },
  { key: 11, text: 'PaLM2', value: 11, color: 'orange' },
  { key: 24, text: 'Gemini', value: 24, color: 'orange' },
  { key: 51, text: 'Gemini (OpenAI)', value: 51, color: 'orange', description: 'Gemini OpenAI compatible format' },
  { key: 28, text: 'Mistral AI', value: 28, color: 'orange' },
  { key: 41, text: 'Novita', value: 41, color: 'purple' },
  { key: 40, text: 'ByteDance Volcano Engine', value: 40, color: 'blue', description: 'Formerly ByteDance Doubao' },
  { key: 15, text: 'Baidu Wenxin Qianfan', value: 15, color: 'blue', tip: 'Get AK (API Key) and SK (Secret Key) from Baidu console' },
  { key: 47, text: 'Baidu Wenxin Qianfan V2', value: 47, color: 'blue', tip: 'For V2 inference service, get API Key from Baidu IAM' },
  { key: 17, text: 'Alibaba Tongyi Qianwen', value: 17, color: 'orange' },
  { key: 49, text: 'Alibaba Cloud Bailian', value: 49, color: 'orange' },
  { key: 18, text: 'iFlytek Spark Cognition', value: 18, color: 'blue', tip: 'WebSocket version API' },
  { key: 48, text: 'iFlytek Spark Cognition V2', value: 48, color: 'blue', tip: 'HTTP version API' },
  { key: 16, text: 'Zhipu ChatGLM', value: 16, color: 'violet' },
  { key: 19, text: '360 ZhiNao', value: 19, color: 'blue' },
  { key: 25, text: 'Moonshot AI', value: 25, color: 'black' },
  { key: 23, text: 'Tencent Hunyuan', value: 23, color: 'teal' },
  { key: 26, text: 'Baichuan Model', value: 26, color: 'orange' },
  { key: 27, text: 'MiniMax', value: 27, color: 'red' },
  { key: 29, text: 'Groq', value: 29, color: 'orange' },
  { key: 30, text: 'Ollama', value: 30, color: 'black' },
  { key: 31, text: '01.AI', value: 31, color: 'green' },
  { key: 32, text: 'StepFun', value: 32, color: 'blue' },
  { key: 34, text: 'Coze', value: 34, color: 'blue' },
  { key: 35, text: 'Cohere', value: 35, color: 'blue' },
  { key: 36, text: 'DeepSeek', value: 36, color: 'black' },
  { key: 37, text: 'Cloudflare', value: 37, color: 'orange' },
  { key: 38, text: 'DeepL', value: 38, color: 'black' },
  { key: 39, text: 'together.ai', value: 39, color: 'blue' },
  { key: 42, text: 'VertexAI', value: 42, color: 'blue' },
  { key: 43, text: 'Proxy', value: 43, color: 'blue' },
  { key: 44, text: 'SiliconFlow', value: 44, color: 'blue' },
  { key: 45, text: 'xAI', value: 45, color: 'blue' },
  { key: 46, text: 'Replicate', value: 46, color: 'blue' },
  { key: 8, text: 'Custom Channel', value: 8, color: 'pink', description: 'Not recommended, use OpenAI Compatible instead' },
  { key: 22, text: 'Knowledge Base: FastGPT', value: 22, color: 'blue' },
  { key: 21, text: 'Knowledge Base: AI Proxy', value: 21, color: 'purple' },
  { key: 20, text: 'OpenRouter', value: 20, color: 'black' },
  { key: 2, text: 'Proxy: API2D', value: 2, color: 'blue' },
  { key: 5, text: 'Proxy: OpenAI-SB', value: 5, color: 'brown' },
  { key: 7, text: 'Proxy: OhMyGPT', value: 7, color: 'purple' },
  { key: 10, text: 'Proxy: AI Proxy', value: 10, color: 'purple' },
  { key: 4, text: 'Proxy: CloseAI', value: 4, color: 'teal' },
  { key: 6, text: 'Proxy: OpenAI Max', value: 6, color: 'violet' },
  { key: 9, text: 'Proxy: AI.LS', value: 9, color: 'yellow' },
  { key: 12, text: 'Proxy: API2GPT', value: 12, color: 'blue' },
  { key: 13, text: 'Proxy: AIGC2D', value: 13, color: 'purple' },
]

const COZE_AUTH_OPTIONS = [
  { key: 'personal_access_token', text: 'Personal Access Token', value: 'personal_access_token' },
  { key: 'oauth_jwt', text: 'OAuth JWT', value: 'oauth_jwt' },
]

const MODEL_MAPPING_EXAMPLE = {
  'gpt-3.5-turbo-0301': 'gpt-3.5-turbo',
  'gpt-4-0314': 'gpt-4',
  'gpt-4-32k-0314': 'gpt-4-32k',
}

const MODEL_CONFIGS_EXAMPLE = {
  'gpt-3.5-turbo-0301': {
    'ratio': 0.0015,
    'completion_ratio': 2.0,
    'max_tokens': 65536,
  },
  'gpt-4': {
    'ratio': 0.03,
    'completion_ratio': 2.0,
    'max_tokens': 128000,
  }
}

const OAUTH_JWT_CONFIG_EXAMPLE = {
  "client_type": "jwt",
  "client_id": "123456789",
  "coze_www_base": "https://www.coze.cn",
  "coze_api_base": "https://api.coze.cn",
  "private_key": "-----BEGIN PRIVATE KEY-----\n***\n-----END PRIVATE KEY-----",
  "public_key_id": "***********************************************************"
}

// JSON validation functions
const isValidJSON = (jsonString: string) => {
  if (!jsonString || jsonString.trim() === '') return true
  try {
    JSON.parse(jsonString)
    return true
  } catch (e) {
    return false
  }
}

const formatJSON = (jsonString: string) => {
  if (!jsonString || jsonString.trim() === '') return ''
  try {
    const parsed = JSON.parse(jsonString)
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    return jsonString
  }
}

// Enhanced model configs validation
const validateModelConfigs = (configStr: string) => {
  if (!configStr || configStr.trim() === '') {
    return { valid: true }
  }

  try {
    const configs = JSON.parse(configStr)

    if (typeof configs !== 'object' || configs === null || Array.isArray(configs)) {
      return { valid: false, error: 'Model configs must be a JSON object' }
    }

    for (const [modelName, config] of Object.entries(configs)) {
      if (!modelName || modelName.trim() === '') {
        return { valid: false, error: 'Model name cannot be empty' }
      }

      if (typeof config !== 'object' || config === null || Array.isArray(config)) {
        return { valid: false, error: `Configuration for model "${modelName}" must be an object` }
      }

      const configObj = config as any
      // Validate ratio
      if (configObj.ratio !== undefined) {
        if (typeof configObj.ratio !== 'number' || configObj.ratio < 0) {
          return { valid: false, error: `Invalid ratio for model "${modelName}": must be a non-negative number` }
        }
      }

      // Validate completion_ratio
      if (configObj.completion_ratio !== undefined) {
        if (typeof configObj.completion_ratio !== 'number' || configObj.completion_ratio < 0) {
          return { valid: false, error: `Invalid completion_ratio for model "${modelName}": must be a non-negative number` }
        }
      }

      // Validate max_tokens
      if (configObj.max_tokens !== undefined) {
        if (!Number.isInteger(configObj.max_tokens) || configObj.max_tokens < 0) {
          return { valid: false, error: `Invalid max_tokens for model "${modelName}": must be a non-negative integer` }
        }
      }

      // Check if at least one meaningful field is provided
      if (configObj.ratio === undefined && configObj.completion_ratio === undefined && configObj.max_tokens === undefined) {
        return { valid: false, error: `Model "${modelName}" must have at least one configuration field (ratio, completion_ratio, or max_tokens)` }
      }
    }

    return { valid: true }
  } catch (error) {
    return { valid: false, error: `Invalid JSON format: ${(error as Error).message}` }
  }
}

// Helper function to get key prompt based on channel type
const getKeyPrompt = (type: number) => {
  switch (type) {
    case 15:
      return 'Please enter Baidu API Key and Secret Key in format: API_KEY|SECRET_KEY'
    case 18:
      return 'Please enter iFlytek App ID, API Secret, and API Key in format: APPID|API_SECRET|API_KEY'
    case 22:
      return 'Please enter FastGPT API Key'
    case 23:
      return 'Please enter Tencent SecretId and SecretKey in format: SECRET_ID|SECRET_KEY'
    default:
      return 'Please enter your API key'
  }
}

export function EditChannelPage() {
  const params = useParams()
  const channelId = params.id
  const isEdit = channelId !== undefined
  const navigate = useNavigate()

  const [loading, setLoading] = useState(isEdit)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [allModels, setAllModels] = useState<Model[]>([])
  const [channelModels, setChannelModels] = useState<string[]>([])
  const [modelSearchTerm, setModelSearchTerm] = useState('')
  const [groups, setGroups] = useState<string[]>([])
  const [defaultPricing, setDefaultPricing] = useState<string>('')
  const [batchMode, setBatchMode] = useState(false)
  const [customModel, setCustomModel] = useState('')

  const form = useForm<ChannelForm>({
    resolver: zodResolver(channelSchema),
    defaultValues: {
      name: '',
      type: 1,
      key: '',
      base_url: '',
      other: '',
      models: [],
      model_mapping: '',
      model_configs: '',
      system_prompt: '',
      groups: ['default'],
      priority: 0,
      weight: 0,
      ratelimit: 0,
      config: {
        region: '',
        ak: '',
        sk: '',
        user_id: '',
        vertex_ai_project_id: '',
        vertex_ai_adc: '',
        auth_type: 'personal_access_token',
      },
      inference_profile_arn_map: '',
    },
  })

  const watchType = form.watch('type')
  const watchConfig = form.watch('config')
  const selectedChannelType = CHANNEL_TYPES.find(t => t.value === watchType)

  const loadChannel = async () => {
    if (!channelId) return

    try {
      const response = await api.get(`/channel/${channelId}`)
      const { success, message, data } = response.data

      if (success && data) {
        // Parse models field - convert string to array
        let models: string[] = []
        if (data.models && typeof data.models === 'string' && data.models.trim() !== '') {
          models = data.models.split(',').map((model: string) => model.trim()).filter((model: string) => model !== '')
        }

        // Parse groups field - convert string to array
        let groups: string[] = ['default']
        if (data.group && typeof data.group === 'string' && data.group.trim() !== '') {
          groups = data.group.split(',').map((group: string) => group.trim()).filter((group: string) => group !== '')
        }

        // Parse JSON configuration
        let config = {
          region: '',
          ak: '',
          sk: '',
          user_id: '',
          vertex_ai_project_id: '',
          vertex_ai_adc: '',
          auth_type: 'personal_access_token',
        }
        if (data.config && typeof data.config === 'string' && data.config.trim() !== '') {
          try {
            config = { ...config, ...JSON.parse(data.config) }
          } catch (e) {
            console.error('Failed to parse config JSON:', e)
          }
        }

        // Format JSON fields for display
        const formatJsonField = (field: string) => {
          if (field && typeof field === 'string' && field.trim() !== '') {
            try {
              return JSON.stringify(JSON.parse(field), null, 2)
            } catch (e) {
              return field
            }
          }
          return ''
        }

        const formData: ChannelForm = {
          name: data.name || '',
          type: data.type || 1,
          key: data.key || '',
          base_url: data.base_url || '',
          other: data.other || '',
          models,
          model_mapping: formatJsonField(data.model_mapping),
          model_configs: formatJsonField(data.model_configs),
          system_prompt: data.system_prompt || '',
          groups,
          priority: data.priority || 0,
          weight: data.weight || 0,
          ratelimit: data.ratelimit || 0,
          config,
          inference_profile_arn_map: formatJsonField(data.inference_profile_arn_map),
        }

        // Load channel-specific models and default pricing
        if (data.type) {
          await Promise.all([
            loadChannelModels(data.type),
            loadDefaultPricing(data.type)
          ])
        }

        console.log('Loaded channel data:', formData)
        form.reset(formData)
        // After reset, log values (no extra setValue needed)
        console.log('Form values after reset:', form.getValues())
      } else {
        throw new Error(message || 'Failed to load channel')
      }
    } catch (error) {
      console.error('Error loading channel:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadAllModels = async () => {
    try {
      const response = await api.get('/models')
      const { success, data } = response.data

      if (success && data) {
        // Extract models from channelId2Models structure
        const allModelsSet = new Set<string>()
        Object.values(data).forEach((channelModels: any) => {
          if (Array.isArray(channelModels)) {
            channelModels.forEach((model: string) => allModelsSet.add(model))
          }
        })

        const models = Array.from(allModelsSet).map((model: string) => ({
          id: model,
          name: model,
        }))
        setAllModels(models)
      }
    } catch (error) {
      console.error('Error loading models:', error)
    }
  }

  const loadChannelModels = async (type: number) => {
    try {
      const response = await api.get('/models')
      const { success, data } = response.data

      if (success && data) {
        // data is channelId2Models object, where keys are channel types
        const typeModels = data[type] || []
        setChannelModels(typeModels)
      }
    } catch (error) {
      console.error('Error loading channel models:', error)
    }
  }

  const loadDefaultPricing = async (channelType: number) => {
    try {
      const response = await api.get(`/channel/default-pricing?type=${channelType}`)
      const { success, data } = response.data
      if (success && data?.model_configs) {
        try {
          const parsed = JSON.parse(data.model_configs)
          const formatted = JSON.stringify(parsed, null, 2)
          setDefaultPricing(formatted)
        } catch (e) {
          setDefaultPricing(data.model_configs)
        }
      }
    } catch (error) {
      console.error('Error loading default pricing:', error)
    }
  }

  const formatJSON = (jsonString: string) => {
    if (!jsonString || jsonString.trim() === '') return ''
    try {
      const parsed = JSON.parse(jsonString)
      return JSON.stringify(parsed, null, 2)
    } catch (e) {
      return jsonString
    }
  }

  const loadGroups = async () => {
    try {
      const response = await api.get('/option/')
      const { success, data } = response.data

      if (success && data) {
        // Extract available groups from system options
        const groupsOption = data.find((option: any) => option.key === 'AvailableGroups')
        if (groupsOption && groupsOption.value) {
          const availableGroups = groupsOption.value.split(',').map((g: string) => g.trim()).filter((g: string) => g !== '')
          setGroups(['default', ...availableGroups])
        } else {
          setGroups(['default'])
        }
      }
    } catch (error) {
      console.error('Error loading groups:', error)
      setGroups(['default'])
    }
  }

  useEffect(() => {
    if (isEdit) {
      loadChannel()
    } else {
      setLoading(false)
    }
    loadAllModels()
    loadGroups()
  }, [isEdit, channelId])

  useEffect(() => {
    if (watchType) {
      loadChannelModels(watchType)
      loadDefaultPricing(watchType)

      // Auto-populate models if none are selected
      const currentModels = form.getValues('models')
      if (currentModels.length === 0) {
        // Load and set channel-specific models
        loadChannelModels(watchType).then(() => {
          if (channelModels.length > 0) {
            form.setValue('models', channelModels)
          }
        })
      }
    }
  }, [watchType])

  const filteredModels = allModels.filter(model =>
    model.name.toLowerCase().includes(modelSearchTerm.toLowerCase())
  )

  const selectedModels = form.watch('models')
  const selectedGroups = form.watch('groups')

  const toggleModel = (modelValue: string) => {
    const currentModels = form.getValues('models')
    if (currentModels.includes(modelValue)) {
      form.setValue('models', currentModels.filter(m => m !== modelValue))
    } else {
      form.setValue('models', [...currentModels, modelValue])
    }
  }

  const toggleGroup = (groupValue: string) => {
    const currentGroups = form.getValues('groups')
    if (currentGroups.includes(groupValue)) {
      form.setValue('groups', currentGroups.filter(g => g !== groupValue))
    } else {
      form.setValue('groups', [...currentGroups, groupValue])
    }
  }

  const onSubmit = async (data: ChannelForm) => {
    setIsSubmitting(true)
    try {
      // Require key only on create (unless special auth path builds it)
      if (!isEdit && (!data.key || data.key.trim() === '')) {
        form.setError('key', { message: 'API key is required' })
        return
      }
      // Validate JSON fields
      if (data.model_mapping && !isValidJSON(data.model_mapping)) {
        form.setError('model_mapping', { message: 'Invalid JSON format in model mapping' })
        return
      }

      if (data.model_configs) {
        const validation = validateModelConfigs(data.model_configs)
        if (!validation.valid) {
          form.setError('model_configs', { message: validation.error || 'Invalid model configs format' })
          return
        }
      }

      if (data.other && !isValidJSON(data.other)) {
        form.setError('other', { message: 'Invalid JSON format in other configuration' })
        return
      }

      if (data.inference_profile_arn_map && !isValidJSON(data.inference_profile_arn_map)) {
        form.setError('inference_profile_arn_map', { message: 'Invalid JSON format in inference profile ARN map' })
        return
      }

      // Validate Coze OAuth JWT config if needed
      if (watchType === 34 && watchConfig.auth_type === 'oauth_jwt') {
        if (!isValidJSON(data.key)) {
          form.setError('key', { message: 'Invalid JSON format for OAuth JWT configuration' })
          return
        }

        try {
          const oauthConfig = JSON.parse(data.key)
          const requiredFields = ['client_type', 'client_id', 'coze_www_base', 'coze_api_base', 'private_key', 'public_key_id']

          for (const field of requiredFields) {
            if (!oauthConfig.hasOwnProperty(field)) {
              form.setError('key', { message: `Missing required field: ${field}` })
              return
            }
          }
        } catch (error) {
          form.setError('key', { message: `OAuth config parse error: ${(error as Error).message}` })
          return
        }
      }

      // Prepare payload
      let payload: any = { ...data }

      // Handle special key construction for AWS and Vertex AI
      if (watchType === 33 && watchConfig.ak && watchConfig.sk && watchConfig.region) {
        payload.key = `${watchConfig.ak}|${watchConfig.sk}|${watchConfig.region}`
      } else if (watchType === 42 && watchConfig.region && watchConfig.vertex_ai_project_id && watchConfig.vertex_ai_adc) {
        payload.key = `${watchConfig.region}|${watchConfig.vertex_ai_project_id}|${watchConfig.vertex_ai_adc}`
      }

      // Convert arrays to comma-separated strings for backend
      payload.models = payload.models.join(',')
      payload.group = payload.groups.join(',')
      delete payload.groups

      // Convert config object to JSON string
      payload.config = JSON.stringify(data.config)

      // Handle empty key for edit operations (don't update if empty)
      if (isEdit && (!payload.key || payload.key.trim() === '')) {
        delete payload.key
      }

      // Handle base_url - remove trailing slash
      if (payload.base_url && payload.base_url.endsWith('/')) {
        payload.base_url = payload.base_url.slice(0, -1)
      }

      // Handle Azure default API version
      if (watchType === 3 && !payload.other) {
        payload.other = '2024-03-01-preview'
      }

      // Convert empty strings to null for optional JSON fields
      const jsonFields = ['model_mapping', 'model_configs', 'other', 'inference_profile_arn_map', 'system_prompt']
      jsonFields.forEach(field => {
        if (payload[field] === '') {
          payload[field] = null
        }
      })

      let response
      if (isEdit && channelId) {
        response = await api.put('/channel/', { ...payload, id: parseInt(channelId) })
      } else {
        response = await api.post('/channel/', payload)
      }

      const { success, message } = response.data
      if (success) {
        navigate('/channels', {
          state: {
            message: isEdit ? 'Channel updated successfully' : 'Channel created successfully'
          }
        })
      } else {
        form.setError('root', { message: message || 'Operation failed' })
      }
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Operation failed'
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const testChannel = async () => {
    if (!channelId) return

    try {
      setIsSubmitting(true)
      const response = await api.get(`/channel/test/${channelId}`)
      const { success, message } = response.data

      if (success) {
        // Show success message
        alert('Channel test successful!')
      } else {
        // Show error message
        alert(`Channel test failed: ${message || 'Unknown error'}`)
      }
    } catch (error) {
      alert(`Channel test failed: ${error instanceof Error ? error.message : 'Network error'}`)
    } finally {
      setIsSubmitting(false)
    }
  }

  // Helper functions for model management
  const addCustomModel = () => {
    if (!customModel.trim()) return
    const currentModels = form.getValues('models')
    if (currentModels.includes(customModel)) return

    form.setValue('models', [...currentModels, customModel])
    setCustomModel('')
  }
  const removeModel = (modelToRemove: string) => {
    const currentModels = form.getValues('models')
    form.setValue('models', currentModels.filter(m => m !== modelToRemove))
  }

  const fillRelatedModels = () => {
    const currentModels = form.getValues('models')
    const uniqueModels = [...new Set([...currentModels, ...channelModels])]
    form.setValue('models', uniqueModels)
  }

  const fillAllModels = () => {
    const currentModels = form.getValues('models')
    const allModelIds = allModels.map(m => m.id)
    const uniqueModels = [...new Set([...currentModels, ...allModelIds])]
    form.setValue('models', uniqueModels)
  }

  const clearModels = () => {
    form.setValue('models', [])
  }

  // Helper functions for group management
  const addGroup = (groupName: string) => {
    const currentGroups = form.getValues('groups')
    if (!currentGroups.includes(groupName)) {
      form.setValue('groups', [...currentGroups, groupName])
    }
  }

  const removeGroup = (groupToRemove: string) => {
    const currentGroups = form.getValues('groups')
    const newGroups = currentGroups.filter(g => g !== groupToRemove)
    // Ensure at least 'default' group remains
    if (newGroups.length === 0) {
      newGroups.push('default')
    }
    form.setValue('groups', newGroups)
  }

  // JSON formatting helpers
  const formatModelMapping = () => {
    const current = form.getValues('model_mapping')
    const formatted = formatJSON(current)
    form.setValue('model_mapping', formatted)
  }

  const formatModelConfigs = () => {
    const current = form.getValues('model_configs')
    const formatted = formatJSON(current)
    form.setValue('model_configs', formatted)
  }

  const formatOtherConfig = () => {
    const current = form.getValues('other')
    const formatted = formatJSON(current)
    form.setValue('other', formatted)
  }

  const loadDefaultModelConfigs = () => {
    if (defaultPricing) {
      form.setValue('model_configs', defaultPricing)
    }
  }

  // Render channel-specific configuration fields
  const renderChannelSpecificFields = () => {
    const channelType = watchType

    switch (channelType) {
      case 3: // Azure OpenAI
        return (
          <div className="space-y-4 p-4 border rounded-lg bg-blue-50/50">
            <h4 className="font-medium text-blue-900">Azure OpenAI Configuration</h4>
            <FormField
              control={form.control}
              name="base_url"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Azure OpenAI Endpoint *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="https://your-resource.openai.azure.com"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="other"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>API Version</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="2024-03-01-preview"
                      {...field}
                    />
                  </FormControl>
                  <span className="text-xs text-muted-foreground">
                    Default: 2024-03-01-preview. This can be overridden by request query parameters.
                  </span>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
              <div className="flex items-center gap-2">
                <AlertCircle className="h-4 w-4 text-yellow-600" />
                <span className="text-sm text-yellow-800">
                  <strong>Important:</strong> The model name should be your deployment name, not the original model name.
                </span>
              </div>
            </div>
          </div>
        )

      case 33: // AWS Bedrock
        return (
          <div className="space-y-4 p-4 border rounded-lg bg-orange-50/50">
            <h4 className="font-medium text-orange-900">AWS Bedrock Configuration</h4>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <FormField
                control={form.control}
                name="config.region"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Region *</FormLabel>
                    <FormControl>
                      <Input placeholder="us-east-1" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="config.ak"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Access Key *</FormLabel>
                    <FormControl>
                      <Input placeholder="AKIA..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="config.sk"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Secret Key *</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Secret Key" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <div className="text-xs text-muted-foreground">
              The final API key will be constructed as: AK|SK|Region
            </div>
          </div>
        )

      case 34: // Coze
        return (
          <div className="space-y-4 p-4 border rounded-lg bg-blue-50/50">
            <h4 className="font-medium text-blue-900">Coze Configuration</h4>
            <FormField
              control={form.control}
              name="config.auth_type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Authentication Type</FormLabel>
                  <Select onValueChange={field.onChange} defaultValue={field.value}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Select authentication type" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {COZE_AUTH_OPTIONS.map(option => (
                        <SelectItem key={option.value} value={option.value}>
                          {option.text}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            {watchConfig.auth_type === 'personal_access_token' ? (
              <FormField
                control={form.control}
                name="key"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Personal Access Token *</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="pat_..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            ) : (
              <FormField
                control={form.control}
                name="key"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>OAuth JWT Configuration *</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={`OAuth JWT configuration in JSON format:\n${JSON.stringify(OAUTH_JWT_CONFIG_EXAMPLE, null, 2)}`}
                        className="font-mono text-sm min-h-[120px]"
                        {...field}
                      />
                    </FormControl>
                    <div className="text-xs text-muted-foreground">
                      Required fields: client_type, client_id, coze_www_base, coze_api_base, private_key, public_key_id
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <FormField
              control={form.control}
              name="config.user_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>User ID</FormLabel>
                  <FormControl>
                    <Input placeholder="User ID for bot operations" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        )

      case 42: // Vertex AI
        return (
          <div className="space-y-4 p-4 border rounded-lg bg-green-50/50">
            <h4 className="font-medium text-green-900">Vertex AI Configuration</h4>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <FormField
                control={form.control}
                name="config.region"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Region *</FormLabel>
                    <FormControl>
                      <Input placeholder="us-central1" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="config.vertex_ai_project_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Project ID *</FormLabel>
                    <FormControl>
                      <Input placeholder="my-project-id" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="config.vertex_ai_adc"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Service Account Credentials *</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Google service account JSON credentials"
                        className="font-mono text-xs"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>
        )

      case 18: // iFlytek Spark
        return (
          <FormField
            control={form.control}
            name="other"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Spark Version</FormLabel>
                <Select onValueChange={field.onChange} defaultValue={field.value}>
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select Spark version" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="v1.1">v1.1</SelectItem>
                    <SelectItem value="v2.1">v2.1</SelectItem>
                    <SelectItem value="v3.1">v3.1</SelectItem>
                    <SelectItem value="v3.5">v3.5</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
        )

      case 21: // Knowledge Base: AI Proxy
        return (
          <FormField
            control={form.control}
            name="other"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Knowledge ID</FormLabel>
                <FormControl>
                  <Input placeholder="Knowledge base ID" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )

      case 17: // Plugin
        return (
          <FormField
            control={form.control}
            name="other"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Plugin Parameters</FormLabel>
                <FormControl>
                  <Input placeholder="Plugin-specific parameters" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )

      case 37: // Cloudflare
        return (
          <FormField
            control={form.control}
            name="config.user_id"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Account ID</FormLabel>
                <FormControl>
                  <Input placeholder="d8d7c61dbc334c32d3ced580e4bf42b4" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        )

      case 8: // Custom Channel (Deprecated)
      case 50: // OpenAI Compatible
        return (
          <div className="space-y-4 p-4 border rounded-lg bg-purple-50/50">
            <h4 className="font-medium text-purple-900">
              {channelType === 8 ? 'Custom Channel Configuration' : 'OpenAI Compatible Configuration'}
            </h4>
            {channelType === 8 && (
              <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
                <div className="flex items-center gap-2">
                  <AlertCircle className="h-4 w-4 text-yellow-600" />
                  <span className="text-sm text-yellow-800">
                    <strong>Deprecated:</strong> Use OpenAI Compatible channel instead.
                  </span>
                </div>
              </div>
            )}
            <FormField
              control={form.control}
              name="base_url"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Base URL *</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="https://api.your-provider.com/v1"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        )

      default:
        return null
    }
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            <span className="ml-3">Loading channel...</span>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-6">
      <Card>
        <CardHeader>
          <CardTitle>{isEdit ? 'Edit Channel' : 'Create Channel'}</CardTitle>
          <CardDescription>
            {isEdit ? 'Update channel configuration' : 'Create a new API channel'}
          </CardDescription>
          {selectedChannelType?.description && (
            <div className="flex items-center gap-2 p-3 bg-blue-50 border border-blue-200 rounded-lg">
              <Info className="h-4 w-4 text-blue-600" />
              <span className="text-sm text-blue-800">{selectedChannelType.description}</span>
            </div>
          )}
          {selectedChannelType?.tip && (
            <div className="flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
              <AlertCircle className="h-4 w-4 text-yellow-600" />
              <span className="text-sm text-yellow-800" dangerouslySetInnerHTML={{ __html: selectedChannelType.tip }} />
            </div>
          )}
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              {/* Basic Configuration */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Channel Name *</FormLabel>
                      <FormControl>
                        <Input placeholder="Enter channel name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="type"
                  render={({ field }) => {
                    const currentValue = field.value
                    return (
                      <FormItem>
                        <FormLabel>Channel Type *</FormLabel>
                        <Select
                          key={`channel-type-${currentValue ?? 'unset'}`}
                          onValueChange={field.onChange}
                          value={currentValue !== undefined && currentValue !== null ? String(currentValue) : undefined}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select channel type" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent className="max-h-96 overflow-y-auto">
                            {CHANNEL_TYPES.map((t) => (
                              <SelectItem key={t.value} value={String(t.value)}>
                                {t.text}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )
                  }}
                />
              </div>

              <FormField
                control={form.control}
                name="key"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>API Key</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        placeholder={isEdit ? "Leave empty to keep existing key" : "Enter API key"}
                        {...field}
                      />
                    </FormControl>
                    {isEdit && (
                      <div className="text-xs text-muted-foreground">
                        Current API key is hidden for security. Enter a new key only if you want to update it.
                      </div>
                    )}
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="base_url"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Base URL (Optional)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g., https://api.openai.com"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="models"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Supported Models *</FormLabel>
                    <div className="flex gap-2 mb-3">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={fillRelatedModels}
                        disabled={channelModels.length === 0}
                        size="sm"
                      >
                        Fill Related Models ({channelModels.length})
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={fillAllModels}
                        size="sm"
                      >
                        Fill All Models ({allModels.length})
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={clearModels}
                        size="sm"
                      >
                        Clear All
                      </Button>
                    </div>
                    <div className="mb-2">
                      <Input
                        placeholder="Search models..."
                        value={modelSearchTerm}
                        onChange={(e) => setModelSearchTerm(e.target.value)}
                      />
                    </div>
                    <div className="max-h-48 overflow-y-auto border rounded-md p-4 space-y-2">
                      {filteredModels.map((model) => (
                        <div key={model.id} className="flex items-center space-x-2">
                          <Checkbox
                            id={model.id}
                            checked={selectedModels.includes(model.id)}
                            onCheckedChange={() => toggleModel(model.id)}
                          />
                          <Label
                            htmlFor={model.id}
                            className="flex-1 cursor-pointer text-sm"
                            onClick={() => navigator.clipboard.writeText(model.id)}
                            title="Click to copy model name"
                          >
                            {model.name}
                          </Label>
                        </div>
                      ))}
                    </div>
                    <div className="mt-2">
                      <div className="flex gap-2 mb-2">
                        <Input
                          placeholder="Add custom model..."
                          value={customModel}
                          onChange={(e) => setCustomModel(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                              e.preventDefault()
                              addCustomModel()
                            }
                          }}
                        />
                        <Button
                          type="button"
                          onClick={addCustomModel}
                          disabled={!customModel.trim()}
                          size="sm"
                        >
                          Add
                        </Button>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {selectedModels.map((model) => (
                        <Badge
                          key={model}
                          variant="secondary"
                          className="cursor-pointer"
                          onClick={() => removeModel(model)}
                        >
                          {model} ×
                        </Badge>
                      ))}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="groups"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Groups *</FormLabel>
                    <div className="space-y-2">
                      <div className="flex flex-wrap gap-2">
                        {groups.map((group) => (
                          <div key={group} className="flex items-center space-x-2">
                            <Checkbox
                              id={`group-${group}`}
                              checked={form.watch('groups').includes(group)}
                              onCheckedChange={(checked) => {
                                const currentGroups = form.getValues('groups')
                                if (checked) {
                                  if (!currentGroups.includes(group)) {
                                    form.setValue('groups', [...currentGroups, group])
                                  }
                                } else {
                                  const newGroups = currentGroups.filter(g => g !== group)
                                  if (newGroups.length === 0) {
                                    newGroups.push('default')
                                  }
                                  form.setValue('groups', newGroups)
                                }
                              }}
                            />
                            <Label htmlFor={`group-${group}`} className="cursor-pointer text-sm">
                              {group}
                            </Label>
                          </div>
                        ))}
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {form.watch('groups').map((group) => (
                          <Badge
                            key={group}
                            variant="secondary"
                            className="cursor-pointer"
                            onClick={() => removeGroup(group)}
                          >
                            {group} ×
                          </Badge>
                        ))}
                      </div>
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="priority"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Priority</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          {...field}
                          onChange={(e) => field.onChange(Number(e.target.value))}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="weight"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Weight</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          {...field}
                          onChange={(e) => field.onChange(Number(e.target.value))}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name="model_mapping"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex items-center gap-2">
                      <FormLabel>Model Mapping (JSON)</FormLabel>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={formatModelMapping}
                        disabled={!field.value || field.value.trim() === ''}
                      >
                        Format JSON
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          const example = JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2)
                          form.setValue('model_mapping', example)
                        }}
                      >
                        Fill Template
                      </Button>
                    </div>
                    <FormControl>
                      <Textarea
                        placeholder={`Model name mapping in JSON format:\n${JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2)}`}
                        className="font-mono text-sm min-h-[100px]"
                        {...field}
                      />
                    </FormControl>
                    <div className="flex justify-between items-center text-sm">
                      <span className="text-muted-foreground">
                        Map model names for this channel (optional)
                      </span>
                      {field.value && field.value.trim() !== '' && (
                        <span className={`font-bold text-xs ${isValidJSON(field.value) ? 'text-green-600' : 'text-red-600'
                          }`}>
                          {isValidJSON(field.value) ? '✓ Valid JSON' : '✗ Invalid JSON'}
                        </span>
                      )}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="model_configs"
                render={({ field }) => (
                  <FormItem>
                    <div className="flex items-center gap-2">
                      <FormLabel>Model Configs (JSON)</FormLabel>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={formatModelConfigs}
                        disabled={!field.value || field.value.trim() === ''}
                      >
                        Format JSON
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={loadDefaultModelConfigs}
                        disabled={!defaultPricing}
                      >
                        Load Defaults
                      </Button>
                    </div>
                    {defaultPricing && (
                      <div className="bg-muted/50 p-4 rounded-lg mb-2">
                        <h4 className="text-sm font-medium mb-2">
                          Default Pricing for {selectedChannelType?.text}
                        </h4>
                        <pre className="text-xs bg-background p-2 rounded border overflow-auto max-h-40">
                          {defaultPricing}
                        </pre>
                      </div>
                    )}
                    <FormControl>
                      <Textarea
                        placeholder={`Model configurations in JSON format:\n${JSON.stringify(MODEL_CONFIGS_EXAMPLE, null, 2)}`}
                        className="font-mono text-sm min-h-[120px]"
                        {...field}
                      />
                    </FormControl>
                    <div className="flex justify-between items-center text-sm">
                      <span className="text-muted-foreground">
                        Configure pricing and limits per model (optional)
                      </span>
                      {field.value && field.value.trim() !== '' && (
                        <span className={`font-bold text-xs ${isValidJSON(field.value) && validateModelConfigs(field.value).valid
                            ? 'text-green-600' : 'text-red-600'
                          }`}>
                          {isValidJSON(field.value) && validateModelConfigs(field.value).valid
                            ? '✓ Valid Config' : '✗ Invalid Config'}
                        </span>
                      )}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* Channel-specific configuration sections */}
              {renderChannelSpecificFields()}

              <FormField
                control={form.control}
                name="system_prompt"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>System Prompt</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Optional system prompt to prepend to all requests"
                        className="min-h-[100px]"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* AWS Bedrock specific field */}
              {watchType === 33 && (
                <FormField
                  control={form.control}
                  name="inference_profile_arn_map"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-2">
                        <FormLabel>Inference Profile ARN Map (AWS Bedrock)</FormLabel>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={formatOtherConfig}
                          disabled={!field.value || field.value.trim() === ''}
                        >
                          Format JSON
                        </Button>
                      </div>
                      <FormControl>
                        <Textarea
                          placeholder={`AWS Bedrock inference profile ARN mapping:\n${JSON.stringify({
                            "claude-3-5-sonnet-20241022": "arn:aws:bedrock:us-east-1:123456789012:inference-profile/us.anthropic.claude-3-5-sonnet-20241022-v2:0",
                            "claude-3-haiku-20240307": "arn:aws:bedrock:us-east-1:123456789012:inference-profile/us.anthropic.claude-3-haiku-20240307-v1:0"
                          }, null, 2)}`}
                          className="font-mono text-sm min-h-[100px]"
                          {...field}
                        />
                      </FormControl>
                      <div className="flex justify-between items-center text-sm">
                        <span className="text-muted-foreground">
                          Map model names to AWS Bedrock inference profile ARNs (optional)
                        </span>
                        {field.value && field.value.trim() !== '' && (
                          <span className={`font-bold text-xs ${isValidJSON(field.value) ? 'text-green-600' : 'text-red-600'
                            }`}>
                            {isValidJSON(field.value) ? '✓ Valid JSON' : '✗ Invalid JSON'}
                          </span>
                        )}
                      </div>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}

              {form.formState.errors.root && (
                <div className="text-sm text-destructive">
                  {form.formState.errors.root.message}
                </div>
              )}

              <div className="flex gap-2">
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting
                    ? (isEdit ? 'Updating...' : 'Creating...')
                    : (isEdit ? 'Update Channel' : 'Create Channel')
                  }
                </Button>
                {isEdit && (
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={testChannel}
                    disabled={isSubmitting}
                  >
                    Test Channel
                  </Button>
                )}
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate('/channels')}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  )
}

export default EditChannelPage
