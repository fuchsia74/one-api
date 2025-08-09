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
import api from '@/lib/api'

const channelSchema = z.object({
  name: z.string().min(1, 'Channel name is required'),
  type: z.number().min(1, 'Channel type is required'),
  key: z.string().min(1, 'API key is required'),
  base_url: z.string().optional(),
  other: z.string().optional(),
  models: z.array(z.string()).default([]),
  model_mapping: z.string().optional(),
  groups: z.array(z.string()).default(['default']),
  priority: z.number().default(0),
  weight: z.number().default(0),
})

type ChannelForm = z.infer<typeof channelSchema>

interface ChannelType {
  key: number
  text: string
  value: number
}

interface Model {
  id: string
  name: string
}

const CHANNEL_TYPES: ChannelType[] = [
  { key: 1, text: 'OpenAI', value: 1 },
  { key: 3, text: 'Azure OpenAI', value: 3 },
  { key: 8, text: 'Custom', value: 8 },
  { key: 11, text: 'Google PaLM', value: 11 },
  { key: 14, text: 'Anthropic Claude', value: 14 },
  { key: 15, text: 'Baidu', value: 15 },
  { key: 17, text: 'Alibaba', value: 17 },
  { key: 18, text: 'Xunfei Spark', value: 18 },
  { key: 19, text: 'Tencent Hunyuan', value: 19 },
  { key: 23, text: 'Tencent', value: 23 },
  { key: 25, text: 'Moonshot', value: 25 },
  { key: 28, text: 'Groq', value: 28 },
]

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
      groups: ['default'],
      priority: 0,
      weight: 0,
    },
  })

  const watchType = form.watch('type')

  const loadChannel = async () => {
    if (!channelId) return

    try {
      const response = await api.get(`/channel/${channelId}`)
      const { success, message, data } = response.data

      if (success && data) {
        // Convert models string to array
        if (data.models === '') {
          data.models = []
        } else {
          data.models = data.models.split(',')
        }

        // Convert groups string to array
        if (data.groups === '') {
          data.groups = ['default']
        } else {
          data.groups = data.groups.split(',')
        }

        form.reset(data)
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
      const response = await api.get('/channel/models')
      const { success, data } = response.data

      if (success && data) {
        const models = data.map((model: string) => ({
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
      const response = await api.get(`/channel/models?type=${type}`)
      const { success, data } = response.data

      if (success && data) {
        setChannelModels(data)
      }
    } catch (error) {
      console.error('Error loading channel models:', error)
    }
  }

  const loadGroups = async () => {
    try {
      const response = await api.get('/group/')
      const { success, data } = response.data

      if (success && data) {
        setGroups(data)
      }
    } catch (error) {
      console.error('Error loading groups:', error)
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

  const fillRelatedModels = () => {
    form.setValue('models', channelModels)
  }

  const fillAllModels = () => {
    form.setValue('models', allModels.map(m => m.id))
  }

  const onSubmit = async (data: ChannelForm) => {
    setIsSubmitting(true)
    try {
      let payload = { ...data }

      // Convert models array to string
      payload.models = payload.models.join(',') as any

      // Convert groups array to string
      payload.groups = payload.groups.join(',') as any

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
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <CardTitle>{isEdit ? 'Edit Channel' : 'Create Channel'}</CardTitle>
          <CardDescription>
            {isEdit ? 'Update channel configuration' : 'Create a new API channel'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Channel Name</FormLabel>
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
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Channel Type</FormLabel>
                      <Select
                        onValueChange={(value) => field.onChange(Number(value))}
                        defaultValue={String(field.value)}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select channel type" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {CHANNEL_TYPES.map((type) => (
                            <SelectItem key={type.value} value={String(type.value)}>
                              {type.text}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
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
                        placeholder="Enter API key"
                        {...field}
                      />
                    </FormControl>
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

              <div className="space-y-4">
                <Label>Supported Models</Label>
                <div className="flex gap-2 mb-3">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={fillRelatedModels}
                    disabled={channelModels.length === 0}
                  >
                    Fill Related Models ({channelModels.length})
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={fillAllModels}
                  >
                    Fill All Models ({allModels.length})
                  </Button>
                </div>
                <Input
                  placeholder="Search models..."
                  value={modelSearchTerm}
                  onChange={(e) => setModelSearchTerm(e.target.value)}
                />
                <div className="max-h-48 overflow-y-auto border rounded-md p-4 space-y-2">
                  {filteredModels.map((model) => (
                    <div key={model.id} className="flex items-center space-x-2">
                      <Checkbox
                        id={model.id}
                        checked={selectedModels.includes(model.id)}
                        onCheckedChange={() => toggleModel(model.id)}
                      />
                      <Label htmlFor={model.id} className="flex-1 cursor-pointer">
                        {model.name}
                      </Label>
                    </div>
                  ))}
                </div>
                <div className="flex flex-wrap gap-1">
                  {selectedModels.map((model) => (
                    <Badge
                      key={model}
                      variant="secondary"
                      className="cursor-pointer"
                      onClick={() => toggleModel(model)}
                    >
                      {model} ×
                    </Badge>
                  ))}
                </div>
              </div>

              <div className="space-y-4">
                <Label>Groups</Label>
                <div className="flex flex-wrap gap-2">
                  {groups.map((group) => (
                    <div key={group} className="flex items-center space-x-2">
                      <Checkbox
                        id={group}
                        checked={selectedGroups.includes(group)}
                        onCheckedChange={() => toggleGroup(group)}
                      />
                      <Label htmlFor={group} className="cursor-pointer">
                        {group}
                      </Label>
                    </div>
                  ))}
                </div>
                <div className="flex flex-wrap gap-1">
                  {selectedGroups.map((group) => (
                    <Badge
                      key={group}
                      variant="secondary"
                      className="cursor-pointer"
                      onClick={() => toggleGroup(group)}
                    >
                      {group} ×
                    </Badge>
                  ))}
                </div>
              </div>

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
                    <FormLabel>Model Mapping (JSON)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder='{"original-model": "mapped-model"}'
                        className="min-h-[100px]"
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
                    <FormLabel>Other Configuration (JSON)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Additional configuration in JSON format"
                        className="min-h-[100px]"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

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
