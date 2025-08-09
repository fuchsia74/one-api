import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Checkbox } from '@/components/ui/checkbox'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import api from '@/lib/api'

const tokenSchema = z.object({
  name: z.string().min(1, 'Token name is required'),
  remain_quota: z.number().min(0, 'Quota must be non-negative'),
  expired_time: z.string().optional(),
  unlimited_quota: z.boolean().default(false),
  models: z.array(z.string()).default([]),
  subnet: z.string().optional(),
})

type TokenForm = z.infer<typeof tokenSchema>

interface Model {
  key: string
  text: string
  value: string
}

export function EditTokenPage() {
  const params = useParams()
  const tokenId = params.id
  const isEdit = tokenId !== undefined
  const navigate = useNavigate()

  const [loading, setLoading] = useState(isEdit)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [modelOptions, setModelOptions] = useState<Model[]>([])
  const [modelSearchTerm, setModelSearchTerm] = useState('')

  const form = useForm<TokenForm>({
    resolver: zodResolver(tokenSchema),
    defaultValues: {
      name: '',
      remain_quota: 500000,
      expired_time: '',
      unlimited_quota: false,
      models: [],
      subnet: '',
    },
  })

  const watchUnlimitedQuota = form.watch('unlimited_quota')

  const loadToken = async () => {
    if (!tokenId) return

    try {
      const response = await api.get(`/token/${tokenId}`)
      const { success, message, data } = response.data

      if (success && data) {
        // Convert timestamp to datetime-local format
        if (data.expired_time !== -1) {
          const date = new Date(data.expired_time * 1000)
          data.expired_time = date.toISOString().slice(0, 16)
        } else {
          data.expired_time = ''
        }

        // Convert models string to array
        if (data.models === '') {
          data.models = []
        } else {
          data.models = data.models.split(',')
        }

        form.reset(data)
      } else {
        throw new Error(message || 'Failed to load token')
      }
    } catch (error) {
      console.error('Error loading token:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadAvailableModels = async () => {
    try {
      const response = await api.get('/user/available_models')
      const { success, message, data } = response.data

      if (success && data) {
        const options = data.map((model: string) => ({
          key: model,
          text: model,
          value: model,
        }))
        setModelOptions(options)
      } else {
        throw new Error(message || 'Failed to load models')
      }
    } catch (error) {
      console.error('Error loading models:', error)
    }
  }

  useEffect(() => {
    if (isEdit) {
      loadToken()
    } else {
      setLoading(false)
    }
    loadAvailableModels()
  }, [isEdit, tokenId])

  const setExpiredTime = (months: number, days: number, hours: number, minutes: number) => {
    if (months === 0 && days === 0 && hours === 0 && minutes === 0) {
      form.setValue('expired_time', '')
      return
    }

    const now = new Date()
    const timestamp = now.getTime() +
      (months * 30 * 24 * 60 * 60 * 1000) +
      (days * 24 * 60 * 60 * 1000) +
      (hours * 60 * 60 * 1000) +
      (minutes * 60 * 1000)

    const date = new Date(timestamp)
    form.setValue('expired_time', date.toISOString().slice(0, 16))
  }

  const filteredModels = modelOptions.filter(model =>
    model.text.toLowerCase().includes(modelSearchTerm.toLowerCase())
  )

  const selectedModels = form.watch('models')

  const toggleModel = (modelValue: string) => {
    const currentModels = form.getValues('models')
    if (currentModels.includes(modelValue)) {
      form.setValue('models', currentModels.filter(m => m !== modelValue))
    } else {
      form.setValue('models', [...currentModels, modelValue])
    }
  }

  const onSubmit = async (data: TokenForm) => {
    setIsSubmitting(true)
    try {
      let payload = { ...data }

      // Convert datetime-local to timestamp
      if (payload.expired_time) {
        const time = Date.parse(payload.expired_time)
        if (isNaN(time)) {
          form.setError('expired_time', { message: 'Invalid expiration time' })
          return
        }
        payload.expired_time = Math.ceil(time / 1000) as any
      } else {
        payload.expired_time = -1 as any
      }

      // Convert models array to string
      const modelsString = payload.models.join(',')
      payload.models = modelsString as any

      let response
      if (isEdit && tokenId) {
        response = await api.put('/token/', { ...payload, id: parseInt(tokenId) })
      } else {
        response = await api.post('/token/', payload)
      }

      const { success, message } = response.data
      if (success) {
        navigate('/tokens', {
          state: {
            message: isEdit ? 'Token updated successfully' : 'Token created successfully'
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
            <span className="ml-3">Loading token...</span>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <CardTitle>{isEdit ? 'Edit Token' : 'Create Token'}</CardTitle>
          <CardDescription>
            {isEdit ? 'Update token settings' : 'Create a new API token'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Token Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter token name" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="space-y-4">
                <Label>Allowed Models</Label>
                <Input
                  placeholder="Search models..."
                  value={modelSearchTerm}
                  onChange={(e) => setModelSearchTerm(e.target.value)}
                />
                <div className="max-h-48 overflow-y-auto border rounded-md p-4 space-y-2">
                  {filteredModels.map((model) => (
                    <div key={model.value} className="flex items-center space-x-2">
                      <Checkbox
                        id={model.value}
                        checked={selectedModels.includes(model.value)}
                        onCheckedChange={() => toggleModel(model.value)}
                      />
                      <Label htmlFor={model.value} className="flex-1 cursor-pointer">
                        {model.text}
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

              <FormField
                control={form.control}
                name="subnet"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>IP Restriction (Optional)</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="e.g., 192.168.1.0/24 or 10.0.0.1"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="expired_time"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Expiration Time</FormLabel>
                    <FormControl>
                      <Input type="datetime-local" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setExpiredTime(0, 0, 0, 0)}
                >
                  Never Expire
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setExpiredTime(1, 0, 0, 0)}
                >
                  1 Month
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setExpiredTime(0, 1, 0, 0)}
                >
                  1 Day
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setExpiredTime(0, 0, 1, 0)}
                >
                  1 Hour
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setExpiredTime(0, 0, 0, 1)}
                >
                  1 Minute
                </Button>
              </div>

              <FormField
                control={form.control}
                name="unlimited_quota"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel>Unlimited Quota</FormLabel>
                    </div>
                  </FormItem>
                )}
              />

              {!watchUnlimitedQuota && (
                <FormField
                  control={form.control}
                  name="remain_quota"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Remaining Quota (tokens)</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          min="0"
                          {...field}
                          onChange={(e) => field.onChange(Number(e.target.value))}
                        />
                      </FormControl>
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
                    : (isEdit ? 'Update Token' : 'Create Token')
                  }
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate('/tokens')}
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

export default EditTokenPage
