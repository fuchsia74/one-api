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
import { api } from '@/lib/api'

// Helper function to render quota with USD conversion (USD only)
const renderQuotaWithPrompt = (quota: number): string => {
  const quotaPerUnitRaw = localStorage.getItem('quota_per_unit')
  const quotaPerUnit = parseFloat(quotaPerUnitRaw || '500000')
  const usd = Number.isFinite(quota) && quotaPerUnit > 0 ? quota / quotaPerUnit : NaN
  const usdValue = Number.isFinite(usd) ? usd.toFixed(2) : '0.00'
  console.log('[QUOTA_DEBUG][Redemption] renderQuotaWithPrompt', { quota, quotaPerUnitRaw, quotaPerUnit, usd, usdValue })
  return `$${usdValue}`
}

const redemptionSchema = z.object({
  name: z.string().min(1, 'Name is required').max(20, 'Max 20 chars'),
  // Coerce numeric fields so typing works and validation runs
  quota: z.coerce.number().int().min(0, 'Quota cannot be negative'),
  count: z.coerce.number().int().min(1, 'Count must be positive').max(100, 'Count cannot exceed 100').default(1),
})

type RedemptionForm = z.infer<typeof redemptionSchema>

export function EditRedemptionPage() {
  const params = useParams()
  const redemptionId = params.id
  const isEdit = redemptionId !== undefined
  const navigate = useNavigate()

  const [loading, setLoading] = useState(isEdit)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<RedemptionForm>({
    resolver: zodResolver(redemptionSchema),
    defaultValues: {
      name: '',
  quota: 0,
  count: 1,
    },
  })

  const watchQuota = form.watch('quota')

  const loadRedemption = async () => {
    if (!redemptionId) return

    try {
      // Unified API call - complete URL with /api prefix
      const response = await api.get(`/api/redemption/${redemptionId}`)
      const { success, message, data } = response.data

      if (success && data) {
        form.reset(data)
      } else {
        throw new Error(message || 'Failed to load redemption')
      }
    } catch (error) {
      console.error('Error loading redemption:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (isEdit) {
      loadRedemption()
    } else {
      setLoading(false)
    }
  }, [isEdit, redemptionId])

  const onSubmit = async (data: RedemptionForm) => {
    setIsSubmitting(true)
    try {
      let response
      if (isEdit && redemptionId) {
        // Unified API call - complete URL with /api prefix
        response = await api.put('/api/redemption/', { ...data, id: parseInt(redemptionId) })
      } else {
        response = await api.post('/api/redemption/', data)
      }

      const { success, message } = response.data
      if (success) {
        navigate('/redemptions', {
          state: {
            message: isEdit ? 'Redemption updated successfully' : 'Redemption created successfully'
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
            <span className="ml-3">Loading redemption...</span>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <CardTitle>{isEdit ? 'Edit Redemption' : 'Create Redemption'}</CardTitle>
          <CardDescription>
            {isEdit ? 'Update redemption code settings' : 'Create a new redemption code'}
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
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter redemption name" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <FormField
                  control={form.control}
                  name="quota"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {(() => {
                          const current = (watchQuota ?? field.value ?? 0) as any
                          const numeric = Number(current)
                          const usdLabel = Number.isFinite(numeric) && numeric >= 0 ? renderQuotaWithPrompt(numeric) : '$0.00'
                          return `Quota (${usdLabel})`
                        })()}
                      </FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          min="0"
                          step="1"
                          {...field}
                          onChange={(e) => {
                            console.log('[QUOTA_DEBUG][Redemption] Input onChange', { value: e.target.value })
                            // Pass original event to RHF to keep name & target intact
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
                  name="count"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Count</FormLabel>
                      <FormControl>
                        <Input
                          type="number"
                          min="1"
                          step="1"
                          {...field}
                          onChange={(e) => {
                            // Pass original event for consistency with RHF expectations
                            field.onChange(e)
                          }}
                        />
                      </FormControl>
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
                    : (isEdit ? 'Update Redemption' : 'Create Redemption')
                  }
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate('/redemptions')}
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

export default EditRedemptionPage
