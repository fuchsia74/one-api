import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Separator } from '@/components/ui/separator'
import { api } from '@/lib/api'
import { useAuthStore } from '@/lib/stores/auth'

const topupSchema = z.object({
  redemption_code: z.string().min(1, 'Redemption code is required'),
})

type TopUpForm = z.infer<typeof topupSchema>

// Helper function to render quota with USD conversion
const renderQuotaWithPrompt = (quota: number): string => {
  const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000')
  const displayInCurrency = localStorage.getItem('display_in_currency') === 'true'

  if (displayInCurrency) {
    const usdValue = (quota / quotaPerUnit).toFixed(6)
    return `${quota.toLocaleString()} tokens ($${usdValue})`
  }
  return `${quota.toLocaleString()} tokens`
}

export function TopUpPage() {
  const { user, updateUser } = useAuthStore()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [userQuota, setUserQuota] = useState(user?.quota || 0)
  const [topUpLink, setTopUpLink] = useState('')
  const [userData, setUserData] = useState<any>(null)

  const form = useForm<TopUpForm>({
    resolver: zodResolver(topupSchema),
    defaultValues: { redemption_code: '' },
  })

  const loadUserData = async () => {
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/user/self')
      const { success, data } = res.data
      if (success) {
        setUserQuota(data.quota)
        setUserData(data)
        updateUser(data)
      }
    } catch (error) {
      console.error('Error loading user data:', error)
    }
  }

  const loadSystemStatus = () => {
    const status = localStorage.getItem('status')
    if (status) {
      try {
        const statusData = JSON.parse(status)
        if (statusData.top_up_link) {
          setTopUpLink(statusData.top_up_link)
        }
      } catch (error) {
        console.error('Error parsing system status:', error)
      }
    }
  }

  const onSubmit = async (data: TopUpForm) => {
    setIsSubmitting(true)
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.post('/api/user/topup', { key: data.redemption_code })
      const { success, message, data: responseData } = res.data

      if (success) {
        const addedQuota = responseData || 0
        setUserQuota(prev => prev + addedQuota)
        form.reset()
        form.setError('root', {
          type: 'success',
          message: `Successfully redeemed! Added ${addedQuota.toLocaleString()} tokens.`
        })
        // Reload user data to get updated quota
        loadUserData()
      } else {
        form.setError('root', { message: message || 'Redemption failed' })
      }
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Redemption failed'
      })
    } finally {
      setIsSubmitting(false)
    }
  }

  const openTopUpLink = () => {
    if (!topUpLink) {
      console.error('No top-up link configured')
      return
    }

    try {
      const url = new URL(topUpLink)
      if (userData) {
        url.searchParams.append('username', userData.username)
        url.searchParams.append('user_id', userData.id.toString())
        const uuid = (globalThis as any).crypto?.randomUUID?.() ??
          'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
            const r = (Math.random() * 16) | 0
            const v = c === 'x' ? r : (r & 0x3) | 0x8
            return v.toString(16)
          })
        url.searchParams.append('transaction_id', uuid)
      }
      window.open(url.toString(), '_blank')
    } catch (error) {
      console.error('Error opening top-up link:', error)
    }
  }

  useEffect(() => {
    loadUserData()
    loadSystemStatus()
  }, [])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-2">Top Up</h1>
          <p className="text-muted-foreground">Manage your account balance and redeem codes</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Current Balance */}
          <Card>
            <CardHeader>
              <CardTitle>Current Balance</CardTitle>
              <CardDescription>Your current quota balance</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center">
                <div className="text-3xl font-bold text-primary mb-2">
                  {renderQuotaWithPrompt(userQuota)}
                </div>
                <p className="text-sm text-muted-foreground">
                  Available quota for API usage
                </p>
                <Button variant="outline" className="mt-4" onClick={loadUserData}>
                  Refresh Balance
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* Redemption Code */}
          <Card>
            <CardHeader>
              <CardTitle>Redeem Code</CardTitle>
              <CardDescription>Enter a redemption code to add quota</CardDescription>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                  <FormField
                    control={form.control}
                    name="redemption_code"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Redemption Code</FormLabel>
                        <FormControl>
                          <Input
                            placeholder="Enter your redemption code"
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  {form.formState.errors.root && (
                    <div className={`text-sm ${form.formState.errors.root.type === 'success'
                      ? 'text-green-600'
                      : 'text-destructive'
                      }`}>
                      {form.formState.errors.root.message}
                    </div>
                  )}

                  <Button type="submit" className="w-full" disabled={isSubmitting}>
                    {isSubmitting ? 'Redeeming...' : 'Redeem Code'}
                  </Button>
                </form>
              </Form>
            </CardContent>
          </Card>
        </div>

        {/* External Top-up */}
        {topUpLink && (
          <Card>
            <CardHeader>
              <CardTitle>Online Payment</CardTitle>
              <CardDescription>
                Purchase quota through our external payment system
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center space-y-4">
                <p className="text-sm text-muted-foreground">
                  Click the button below to open our secure payment portal where you can
                  purchase additional quota for your account.
                </p>
                <Button onClick={openTopUpLink} size="lg">
                  Open Payment Portal
                </Button>
                <p className="text-xs text-muted-foreground">
                  You will be redirected to an external payment system.
                  Your account information will be automatically included.
                </p>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Usage Tips */}
        <Card>
          <CardHeader>
            <CardTitle>Tips</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm text-muted-foreground">
              <p>• Quota is consumed based on your API usage and model costs</p>
              <p>• Different models have different pricing rates</p>
              <p>• Check the Models page to see pricing for each model</p>
              <p>• Redemption codes are case-sensitive</p>
              <p>• Your balance will be automatically updated after successful redemption</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default TopUpPage
