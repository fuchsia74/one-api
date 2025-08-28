import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api } from '@/lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Info } from 'lucide-react'

// Helper function to render quota with USD conversion (USD only)
const renderQuotaWithPrompt = (quota: number): string => {
  const quotaPerUnitRaw = localStorage.getItem('quota_per_unit')
  const quotaPerUnit = parseFloat(quotaPerUnitRaw || '500000')
  const usd = Number.isFinite(quota) && quotaPerUnit > 0 ? quota / quotaPerUnit : NaN
  const usdValue = Number.isFinite(usd) ? usd.toFixed(2) : '0.00'
  console.log('[QUOTA_DEBUG][User] renderQuotaWithPrompt', { quota, quotaPerUnitRaw, quotaPerUnit, usd, usdValue })
  return `$${usdValue}`
}

const userSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  display_name: z.string().optional(),
  password: z.string().optional(),
  email: z.string().email('Valid email is required').optional(),
  // Coerce to number since HTML inputs provide string values
  quota: z.coerce.number().min(0, 'Quota must be non-negative'),
  group: z.string().min(1, 'Group is required'),
})

type UserForm = z.infer<typeof userSchema>

interface Group {
  key: string
  text: string
  value: string
}

export function EditUserPage() {
  const params = useParams()
  const userId = params.id
  const isEdit = userId !== undefined
  const navigate = useNavigate()

  const [loading, setLoading] = useState(isEdit)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [groupOptions, setGroupOptions] = useState<Group[]>([])

  const form = useForm<UserForm>({
    resolver: zodResolver(userSchema),
    defaultValues: {
      username: '',
      display_name: '',
      password: '',
      email: '',
      quota: 0,
      group: 'default',
    },
  })

  const watchQuota = useWatch({ control: form.control, name: 'quota' })
  useEffect(() => {
    console.log('[QUOTA_DEBUG][User] watchQuota changed:', watchQuota, typeof watchQuota)
  }, [watchQuota])

  const loadUser = async () => {
    if (!userId) return

    try {
      // Unified API call - complete URL with /api prefix
      const response = await api.get(`/api/user/${userId}`)
      const { success, message, data } = response.data

      if (success && data) {
        form.reset({
          ...data,
          password: '', // Don't pre-fill password for security
        })
      } else {
        throw new Error(message || 'Failed to load user')
      }
    } catch (error) {
      console.error('Error loading user:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadGroups = async () => {
    try {
      // Unified API call - complete URL with /api prefix
      const response = await api.get('/api/group/')
      const { success, data } = response.data

      if (success && data) {
        const options = data.map((group: string) => ({
          key: group,
          text: group,
          value: group,
        }))
        setGroupOptions(options)
      }
    } catch (error) {
      console.error('Error loading groups:', error)
    }
  }

  useEffect(() => {
    if (isEdit) {
      loadUser()
    } else {
      setLoading(false)
    }
    loadGroups()
  }, [isEdit, userId])

  const onSubmit = async (data: UserForm) => {
    setIsSubmitting(true)
    try {
      let payload = { ...data }

      // Don't send empty password
      if (!payload.password) {
        delete payload.password
      }

      let response: any
      if (isEdit && userId) {
        // Unified API call - complete URL with /api prefix
        response = await api.put('/api/user/', { ...payload, id: parseInt(userId) })
      } else {
        response = await api.post('/api/user/', payload)
      }

      const { success, message } = response.data
      if (success) {
        navigate('/users', {
          state: {
            message: isEdit ? 'User updated successfully' : 'User created successfully'
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
            <span className="ml-3">Loading user...</span>
          </CardContent>
        </Card>
  </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <TooltipProvider>
      <Card>
        <CardHeader>
          <CardTitle>{isEdit ? 'Edit User' : 'Create User'}</CardTitle>
          <CardDescription>
            {isEdit ? 'Update user information' : 'Create a new user account'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              {/* helper for label + tooltip */}
              {(() => {
                // small inline helper component
                const LabelWithHelp = ({ label, help }: { label: string; help: string }) => (
                  <div className="flex items-center gap-1">
                    <span className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                      {label}
                    </span>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label={`Help: ${label}`} />
                      </TooltipTrigger>
                      <TooltipContent className="max-w-xs whitespace-pre-line">{help}</TooltipContent>
                    </Tooltip>
                  </div>
                )
                return null
              })()}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-1">
                        <FormLabel>Username *</FormLabel>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Username" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-xs">Unique login name. Min 3 characters.</TooltipContent>
                        </Tooltip>
                      </div>
                      <FormControl>
                        <Input placeholder="Enter username" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="display_name"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-1">
                        <FormLabel>Display Name</FormLabel>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Display Name" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-xs">Optional human‑readable name shown in the UI.</TooltipContent>
                        </Tooltip>
                      </div>
                      <FormControl>
                        <Input placeholder="Enter display name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-1">
                        <FormLabel>Email</FormLabel>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Email" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-xs">Optional contact address for password reset and notifications.</TooltipContent>
                        </Tooltip>
                      </div>
                      <FormControl>
                        <Input type="email" placeholder="Enter email" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-1">
                        <FormLabel>{isEdit ? 'New Password (leave empty to keep current)' : 'Password *'}</FormLabel>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Password" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-xs">Minimum length depends on policy. Leave empty when editing to keep unchanged.</TooltipContent>
                        </Tooltip>
                      </div>
                      <FormControl>
                        <Input type="password" placeholder="Enter password" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="quota"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-1">
                        <FormLabel>
                          {(() => {
                            const current = watchQuota ?? field.value ?? 0
                            const numeric = Number(current)
                            const usdLabel = Number.isFinite(numeric) && numeric >= 0 ? renderQuotaWithPrompt(numeric) : '$0.00'
                            return `Quota (${usdLabel})`
                          })()}
                        </FormLabel>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Quota" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-xs">Quota units are tokens. USD estimate uses the per‑unit ratio configured by admin.</TooltipContent>
                        </Tooltip>
                      </div>
                      <FormControl>
                        <Input
                          type="number"
                          min="0"
                          {...field}
                          onChange={(e) => {
                            console.log('[QUOTA_DEBUG][User] Input onChange', { value: e.target.value })
                            field.onChange(e)
                          }}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="group"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-1">
                        <FormLabel>Group *</FormLabel>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <Info className="h-4 w-4 text-muted-foreground cursor-help" aria-label="Help: Group" />
                          </TooltipTrigger>
                          <TooltipContent className="max-w-xs">User group controls access and model/channel visibility.</TooltipContent>
                        </Tooltip>
                      </div>
                      <Select onValueChange={field.onChange} value={field.value}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="Select a group" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {groupOptions.map((group) => (
                            <SelectItem key={group.value} value={group.value}>
                              {group.text}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {form.formState.errors.root && (
                <div className="text-sm text-destructive">
                  {form.formState.errors.root.message}
                </div>
              )}

              <div className="flex gap-2">
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting
                    ? (isEdit ? 'Updating...' : 'Creating...')
                    : (isEdit ? 'Update User' : 'Create User')
                  }
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate('/users')}
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
  )
}

export default EditUserPage
