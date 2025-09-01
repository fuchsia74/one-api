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
import { api } from '@/lib/api'
import { logEditPageLayout } from '@/dev/layout-debug'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Info } from 'lucide-react'
import { useNotifications } from '@/components/ui/notifications'

// Helper function to render quota with USD conversion (USD only)
const renderQuotaWithPrompt = (quota: number): string => {
  const quotaPerUnitRaw = localStorage.getItem('quota_per_unit')
  const quotaPerUnit = parseFloat(quotaPerUnitRaw || '500000')
  const usd = Number.isFinite(quota) && quotaPerUnit > 0 ? quota / quotaPerUnit : NaN
  const usdValue = Number.isFinite(usd) ? usd.toFixed(2) : '0.00'
  console.log(`[QUOTA_DEBUG] renderQuota quota=${String(quota)} ratioRaw=${String(quotaPerUnitRaw)} ratio=${String(quotaPerUnit)} usd=${String(usd)} usdValue=${usdValue}`)
  return `$${usdValue}`
}

const tokenSchema = z.object({
  name: z.string().min(1, 'Token name is required'),
  remain_quota: z.coerce.number().min(0, 'Quota must be non-negative'),
  expired_time: z.string().optional(),
  unlimited_quota: z.boolean().default(false),
  models: z.array(z.string()).default([]),
  subnet: z.string().optional(),
})

type TokenForm = z.infer<typeof tokenSchema>

// Matches a subset of backend Token for status handling
type BackendToken = {
  id: number
  status: number
}

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
  const { notify } = useNotifications()

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
  const watchRemainQuota = form.watch('remain_quota')
  useEffect(() => {
    console.log(`[QUOTA_DEBUG] watchRemainQuota=${String(watchRemainQuota)} type=${typeof watchRemainQuota}`)
  }, [watchRemainQuota])

  const loadToken = async () => {
    if (!tokenId) return

    try {
      // Unified API call - complete URL with /api prefix
      const response = await api.get(`/api/token/${tokenId}`)
      const { success, message, data } = response.data

      if (success && data) {
        // Convert timestamp to datetime-local format
        const rawExpired = Number(data.expired_time)
        if (rawExpired && rawExpired > 0) {
          const date = new Date(rawExpired * 1000)
          data.expired_time = date.toISOString().slice(0, 16)
        } else {
          // Treat 0, -1, null, undefined as never
          data.expired_time = ''
        }

        // Convert models string to array
        const modelsRaw = data.models
        if (Array.isArray(modelsRaw)) {
          data.models = modelsRaw
        } else if (typeof modelsRaw === 'string') {
          data.models = modelsRaw === '' ? [] : modelsRaw.split(',').filter(Boolean)
        } else {
          data.models = []
        }

        // Normalize potentially nullish fields
        if (data.name == null) data.name = ''
        if (data.subnet == null) data.subnet = ''

        form.reset(data)
          // Persist original id/status for submission logic
          ; (form as any)._original = { id: data.id as number, status: data.status as number } as BackendToken
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
      // Unified API call - complete URL with /api prefix
      const response = await api.get('/api/user/available_models')
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
          notify({ type: 'error', title: 'Validation error', message: 'Invalid expiration time' })
          return
        }
        payload.expired_time = Math.ceil(time / 1000) as any
      } else {
        payload.expired_time = -1 as any
      }

      // Convert models array to string
      const modelsString = payload.models.join(',')
      payload.models = modelsString as any

      let response: any
      // Include current status and auto-adjust so Unlimited or new expiry takes effect
      const original: BackendToken | undefined = (form as any)._original
      if (original) {
        let nextStatus = original.status
        const nowSec = Math.floor(Date.now() / 1000)
        const exp = Number((payload as any).expired_time)
        const isUnlimited = !!(payload as any).unlimited_quota
        const hasQuota = Number((payload as any).remain_quota) > 0
        // Exhausted -> Enabled if unlimited or quota > 0
        if (nextStatus === 4 && (isUnlimited || hasQuota)) nextStatus = 1
        // Expired -> Enabled if never expire or a future expiry
        if (nextStatus === 3 && (exp === -1 || exp > nowSec)) nextStatus = 1
          ; (payload as any).status = nextStatus
      }
      if (isEdit && tokenId) {
        // Unified API call - complete URL with /api prefix
        response = await api.put('/api/token/', { ...payload, id: parseInt(tokenId) })
      } else {
        response = await api.post('/api/token/', payload)
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
        notify({ type: 'error', title: 'Request failed', message: message || 'Operation failed' })
      }
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Operation failed'
      })
      notify({ type: 'error', title: 'Unexpected error', message: error instanceof Error ? error.message : 'Operation failed' })
    } finally {
      setIsSubmitting(false)
    }
  }

  // RHF invalid handler
  const onInvalid = (errors: any) => {
    const firstKey = Object.keys(errors)[0]
    const firstMsg = errors[firstKey]?.message || 'Please correct the highlighted fields.'
    notify({ type: 'error', title: 'Validation error', message: String(firstMsg) })
    const el = document.querySelector(`[name="${firstKey}"]`) as HTMLElement | null
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
        ; (el as any).focus?.()
    }
  }

  // Error highlighting
  const hasError = (path: string): boolean => !!(form.formState.errors as any)?.[path]
  const errorClass = (path: string) => (hasError(path) ? 'border-destructive focus-visible:ring-destructive' : '')

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
    // Trigger layout diagnostics after render
    // eslint-disable-next-line react/jsx-no-useless-fragment
    <>
      {(() => { logEditPageLayout('EditTokenPage'); return null })()}
      <div className="container mx-auto px-4 py-8">
        <TooltipProvider>
          <Card>
            <CardHeader>
              <CardTitle>{isEdit ? 'Edit Token' : 'Create Token'}</CardTitle>
              <CardDescription>
                {isEdit ? 'Update token settings' : 'Create a new API token'}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit, onInvalid)} className="space-y-6">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-1">
                          <FormLabel>Token Name</FormLabel>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Token Name" />
                            </TooltipTrigger>
                            <TooltipContent className="max-w-xs">Human‑readable identifier for this token.</TooltipContent>
                          </Tooltip>
                        </div>
                        <FormControl>
                          <Input placeholder="Enter token name" className={errorClass('name')} {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <div className="space-y-4">
                    <div className="flex items-center gap-1">
                      <Label>Allowed Models</Label>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Allowed Models" />
                        </TooltipTrigger>
                        <TooltipContent className="max-w-xs">Restrict this token to specific models. Leave empty to allow all models available to the user/group.</TooltipContent>
                      </Tooltip>
                    </div>
                    <Input
                      placeholder="Search models..."
                      value={modelSearchTerm}
                      onChange={(e) => setModelSearchTerm(e.target.value)}
                    />
                    <div className="relative isolate max-h-48 overflow-y-auto border rounded-md p-4 space-y-2">
                      {filteredModels.map((model) => (
                        <div key={model.value} className="relative flex items-center space-x-2">
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
                        <div className="flex items-center gap-1">
                          <FormLabel>IP Restriction (Optional)</FormLabel>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: IP Restriction" />
                            </TooltipTrigger>
                            <TooltipContent className="max-w-xs">Allow requests only from these IPs or CIDR ranges (e.g., 192.168.1.0/24). Multiple entries are comma‑separated.</TooltipContent>
                          </Tooltip>
                        </div>
                        <FormControl>
                          <Input
                            placeholder="e.g., 192.168.1.0/24 or 10.0.0.1"
                            className={errorClass('subnet')}
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
                        <div className="flex items-center gap-1">
                          <FormLabel>Expiration Time</FormLabel>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Expiration Time" />
                            </TooltipTrigger>
                            <TooltipContent className="max-w-xs">Set when this token expires. Leave empty for never. Use quick buttons for common durations.</TooltipContent>
                          </Tooltip>
                        </div>
                        <FormControl>
                          <Input type="datetime-local" className={errorClass('expired_time')} {...field} />
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

                  <div className="flex flex-row items-start space-x-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        id="unlimited_quota"
                        checked={!!watchUnlimitedQuota}
                        onCheckedChange={(checked) => form.setValue('unlimited_quota', !!checked, { shouldDirty: true })}
                      />
                    </FormControl>
                    <div className="flex items-center gap-1 space-y-0">
                      <FormLabel htmlFor="unlimited_quota">Unlimited Quota</FormLabel>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Unlimited Quota" />
                        </TooltipTrigger>
                        <TooltipContent className="max-w-xs">If enabled, this token ignores remaining quota checks.</TooltipContent>
                      </Tooltip>
                    </div>
                  </div>

                  <FormField
                    control={form.control}
                    name="remain_quota"
                    render={({ field }) => {
                      const raw = field.value as any
                      const fallback = form.getValues('remain_quota') as any
                      const current = (raw ?? fallback) as any
                      const numeric = Number(current)
                      const usdLabel = Number.isFinite(numeric) && numeric >= 0 ? renderQuotaWithPrompt(numeric) : '$0.00'
                      console.log(`[QUOTA_DEBUG] remainQuota raw=${String(raw)} fallback=${String(fallback)} current=${String(current)} numeric=${String(numeric)} usd=${usdLabel} type=${typeof raw}`)
                      return (
                        <FormItem>
                          <div className="flex items-center gap-1">
                            <FormLabel>
                              Remaining Quota ({usdLabel})
                            </FormLabel>
                            <Tooltip>
                              <TooltipTrigger asChild>
                                <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Remaining Quota" />
                              </TooltipTrigger>
                              <TooltipContent className="max-w-xs">Quota is measured in tokens. USD is an estimate based on admin‑configured per‑unit pricing.</TooltipContent>
                            </Tooltip>
                          </div>
                          <FormControl>
                            <Input
                              type="number"
                              min="0"
                              disabled={watchUnlimitedQuota}
                              className={errorClass('remain_quota')}
                              {...field}
                              onChange={(e) => {
                                console.log(`[QUOTA_DEBUG] onChange remain_quota value=${String(e.target.value)}`)
                                // Pass original event to RHF (prevents libs reading value.name from breaking)
                                field.onChange(e)
                              }}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )
                    }}
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
        </TooltipProvider>
      </div>
    </>
  )
}

export default EditTokenPage
