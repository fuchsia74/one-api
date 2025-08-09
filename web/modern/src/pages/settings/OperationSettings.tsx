import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Checkbox } from '@/components/ui/checkbox'
import { Separator } from '@/components/ui/separator'
import api from '@/lib/api'

const operationSchema = z.object({
  QuotaForNewUser: z.number().min(0).default(0),
  QuotaForInviter: z.number().min(0).default(0),
  QuotaForInvitee: z.number().min(0).default(0),
  QuotaRemindThreshold: z.number().min(0).default(0),
  PreConsumedQuota: z.number().min(0).default(0),
  TopUpLink: z.string().default(''),
  ChatLink: z.string().default(''),
  QuotaPerUnit: z.number().min(0).default(500000),
  ChannelDisableThreshold: z.number().min(0).default(0),
  RetryTimes: z.number().min(0).default(0),
  AutomaticDisableChannelEnabled: z.boolean().default(false),
  AutomaticEnableChannelEnabled: z.boolean().default(false),
  LogConsumeEnabled: z.boolean().default(false),
  DisplayInCurrencyEnabled: z.boolean().default(false),
  DisplayTokenStatEnabled: z.boolean().default(false),
  ApproximateTokenEnabled: z.boolean().default(false),
})

type OperationForm = z.infer<typeof operationSchema>

export function OperationSettings() {
  const [loading, setLoading] = useState(true)
  const [historyTimestamp, setHistoryTimestamp] = useState('')

  const form = useForm<OperationForm>({
    resolver: zodResolver(operationSchema),
    defaultValues: {
      QuotaForNewUser: 0,
      QuotaForInviter: 0,
      QuotaForInvitee: 0,
      QuotaRemindThreshold: 0,
      PreConsumedQuota: 0,
      TopUpLink: '',
      ChatLink: '',
      QuotaPerUnit: 500000,
      ChannelDisableThreshold: 0,
      RetryTimes: 0,
      AutomaticDisableChannelEnabled: false,
      AutomaticEnableChannelEnabled: false,
      LogConsumeEnabled: false,
      DisplayInCurrencyEnabled: false,
      DisplayTokenStatEnabled: false,
      ApproximateTokenEnabled: false,
    },
  })

  const loadOptions = async () => {
    try {
      const res = await api.get('/option/')
      const { success, data } = res.data
      if (success && data) {
        const formData: any = {}
        data.forEach((item: { key: string; value: string }) => {
          const key = item.key
          if (key in form.getValues()) {
            if (key.endsWith('Enabled')) {
              formData[key] = item.value === 'true'
            } else {
              const numValue = parseFloat(item.value)
              formData[key] = isNaN(numValue) ? item.value : numValue
            }
          }
        })
        form.reset(formData)
      }
    } catch (error) {
      console.error('Error loading options:', error)
    } finally {
      setLoading(false)
    }
  }

  const updateOption = async (key: string, value: string | number | boolean) => {
    try {
      await api.put('/option/', { key, value: String(value) })
      console.log(`Updated ${key}: ${value}`)
    } catch (error) {
      console.error(`Error updating ${key}:`, error)
    }
  }

  const onSubmitGroup = async (group: 'quota' | 'general' | 'monitor') => {
    const values = form.getValues()

    switch (group) {
      case 'quota':
        await updateOption('QuotaForNewUser', values.QuotaForNewUser)
        await updateOption('QuotaForInviter', values.QuotaForInviter)
        await updateOption('QuotaForInvitee', values.QuotaForInvitee)
        await updateOption('PreConsumedQuota', values.PreConsumedQuota)
        break
      case 'general':
        await updateOption('TopUpLink', values.TopUpLink)
        await updateOption('ChatLink', values.ChatLink)
        await updateOption('QuotaPerUnit', values.QuotaPerUnit)
        await updateOption('RetryTimes', values.RetryTimes)
        break
      case 'monitor':
        await updateOption('QuotaRemindThreshold', values.QuotaRemindThreshold)
        await updateOption('ChannelDisableThreshold', values.ChannelDisableThreshold)
        break
    }
  }

  const deleteHistoryLogs = async () => {
    if (!historyTimestamp) return
    try {
      const timestamp = Date.parse(historyTimestamp) / 1000
      const res = await api.delete(`/log/?target_timestamp=${timestamp}`)
      const { success, message, data } = res.data
      if (success) {
        console.log(`Cleared ${data} logs!`)
      } else {
        console.error('Failed to clear logs:', message)
      }
    } catch (error) {
      console.error('Error clearing logs:', error)
    }
  }

  useEffect(() => {
    loadOptions()

    // Set default history timestamp to 30 days ago
    const now = new Date()
    const monthAgo = new Date(now.getTime() - 30 * 24 * 3600 * 1000)
    setHistoryTimestamp(monthAgo.toISOString().slice(0, 10))
  }, [])

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <span className="ml-3">Loading operation settings...</span>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Quota Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Quota Settings</CardTitle>
          <CardDescription>Configure user quota and invitation rewards</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="QuotaForNewUser"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Quota for New User</FormLabel>
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
                name="QuotaForInviter"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Quota for Inviter</FormLabel>
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
                name="QuotaForInvitee"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Quota for Invitee</FormLabel>
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
                name="PreConsumedQuota"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Pre-consumed Quota</FormLabel>
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
            <div className="mt-4">
              <Button onClick={() => onSubmitGroup('quota')}>Save Quota Settings</Button>
            </div>
          </Form>
        </CardContent>
      </Card>

      {/* General Settings */}
      <Card>
        <CardHeader>
          <CardTitle>General Settings</CardTitle>
          <CardDescription>Configure general operation parameters</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="TopUpLink"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Top-up Link</FormLabel>
                    <FormControl>
                      <Input placeholder="https://..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="ChatLink"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Chat Link</FormLabel>
                    <FormControl>
                      <Input placeholder="https://..." {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="QuotaPerUnit"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Quota per Unit</FormLabel>
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
                name="RetryTimes"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Retry Times</FormLabel>
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

            <Separator className="my-4" />

            <div className="space-y-4">
              <FormField
                control={form.control}
                name="LogConsumeEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center space-x-2">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) => {
                          field.onChange(checked)
                          updateOption('LogConsumeEnabled', checked)
                        }}
                      />
                    </FormControl>
                    <FormLabel>Enable Consumption Logging</FormLabel>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="DisplayInCurrencyEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center space-x-2">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) => {
                          field.onChange(checked)
                          updateOption('DisplayInCurrencyEnabled', checked)
                        }}
                      />
                    </FormControl>
                    <FormLabel>Display in Currency</FormLabel>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="DisplayTokenStatEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center space-x-2">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) => {
                          field.onChange(checked)
                          updateOption('DisplayTokenStatEnabled', checked)
                        }}
                      />
                    </FormControl>
                    <FormLabel>Display Token Statistics</FormLabel>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="ApproximateTokenEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center space-x-2">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) => {
                          field.onChange(checked)
                          updateOption('ApproximateTokenEnabled', checked)
                        }}
                      />
                    </FormControl>
                    <FormLabel>Enable Approximate Token Counting</FormLabel>
                  </FormItem>
                )}
              />
            </div>

            <div className="mt-4">
              <Button onClick={() => onSubmitGroup('general')}>Save General Settings</Button>
            </div>
          </Form>
        </CardContent>
      </Card>

      {/* Monitoring Settings */}
      <Card>
        <CardHeader>
          <CardTitle>Monitoring & Channel Settings</CardTitle>
          <CardDescription>Configure monitoring thresholds and channel management</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="QuotaRemindThreshold"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Quota Remind Threshold</FormLabel>
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
                name="ChannelDisableThreshold"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Channel Disable Threshold</FormLabel>
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

            <Separator className="my-4" />

            <div className="space-y-4">
              <FormField
                control={form.control}
                name="AutomaticDisableChannelEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center space-x-2">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) => {
                          field.onChange(checked)
                          updateOption('AutomaticDisableChannelEnabled', checked)
                        }}
                      />
                    </FormControl>
                    <FormLabel>Automatic Channel Disable</FormLabel>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="AutomaticEnableChannelEnabled"
                render={({ field }) => (
                  <FormItem className="flex items-center space-x-2">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={(checked) => {
                          field.onChange(checked)
                          updateOption('AutomaticEnableChannelEnabled', checked)
                        }}
                      />
                    </FormControl>
                    <FormLabel>Automatic Channel Enable</FormLabel>
                  </FormItem>
                )}
              />
            </div>

            <div className="mt-4">
              <Button onClick={() => onSubmitGroup('monitor')}>Save Monitoring Settings</Button>
            </div>
          </Form>
        </CardContent>
      </Card>

      {/* Log Management */}
      <Card>
        <CardHeader>
          <CardTitle>Log Management</CardTitle>
          <CardDescription>Clear historical logs to free up storage space</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-4">
            <Input
              type="date"
              value={historyTimestamp}
              onChange={(e) => setHistoryTimestamp(e.target.value)}
              className="w-auto"
            />
            <Button variant="destructive" onClick={deleteHistoryLogs}>
              Clear Logs Before This Date
            </Button>
          </div>
          <p className="text-sm text-muted-foreground mt-2">
            This will permanently delete all logs before the selected date.
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

export default OperationSettings
