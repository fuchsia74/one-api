import { useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Checkbox } from '@/components/ui/checkbox'
import { Separator } from '@/components/ui/separator'
import { api } from '@/lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Info } from 'lucide-react'

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

  // Descriptions for each setting used on this page
  const descriptions = useMemo<Record<string, string>>(
    () => ({
      // Quota
      QuotaForNewUser: 'Initial quota granted to each newly registered user.',
      QuotaForInviter: 'Quota reward granted to the inviter after a successful invite.',
      QuotaForInvitee: 'Quota reward granted to the invitee upon successful registration.',
      PreConsumedQuota: 'Quota reserved at request start to avoid abuse. Unused part is returned after billing.',

      // General
      TopUpLink: 'External link for users to purchase or top up quota.',
      ChatLink: 'External chat/support link shown in the UI.',
      QuotaPerUnit: 'Conversion ratio for currency display. Higher value makes each $ represent more quota.',
      RetryTimes: 'Automatic retry attempts for upstream requests on transient errors.',
      LogConsumeEnabled: 'Record usage/consumption logs. Turn off to reduce storage overhead.',
      DisplayInCurrencyEnabled: 'Show usage and quotas as currency in the UI, based on the configured conversion.',
      DisplayTokenStatEnabled: 'Display token statistics in logs and dashboards when available.',
      ApproximateTokenEnabled: 'Use a faster approximation for token counting to improve performance (may be slightly less accurate).',

      // Monitoring & Channels
      QuotaRemindThreshold: 'When remaining quota falls below this value, users will be reminded.',
      ChannelDisableThreshold: 'Failure rate threshold (percentage) to auto‑disable a channel. Default 5%.',
      AutomaticDisableChannelEnabled: 'Automatically disable channels that show sustained failures.',
      AutomaticEnableChannelEnabled: 'Automatically re‑enable previously disabled channels when healthy.',
    }),
    []
  )

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
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/option/')
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
      // Unified API call - complete URL with /api prefix
      await api.put('/api/option/', { key, value: String(value) })
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
      // Unified API call - complete URL with /api prefix
      const res = await api.delete(`/api/log/?target_timestamp=${timestamp}`)
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
    <TooltipProvider>
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
                      <FormLabel className="flex items-center gap-2">
                        Quota for New User
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Quota for New User">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.QuotaForNewUser}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Quota for Inviter
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Quota for Inviter">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.QuotaForInviter}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Quota for Invitee
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Quota for Invitee">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.QuotaForInvitee}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Pre-consumed Quota
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Pre-consumed Quota">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.PreConsumedQuota}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Top-up Link
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Top-up Link">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.TopUpLink}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Chat Link
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Chat Link">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.ChatLink}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Quota per Unit
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Quota per Unit">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.QuotaPerUnit}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Retry Times
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Retry Times">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.RetryTimes}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Enable Consumption Logging
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Enable Consumption Logging">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.LogConsumeEnabled}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Display in Currency
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Display in Currency">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.DisplayInCurrencyEnabled}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Display Token Statistics
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Display Token Statistics">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.DisplayTokenStatEnabled}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Enable Approximate Token Counting
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Approximate Token Counting">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.ApproximateTokenEnabled}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Quota Remind Threshold
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Quota Remind Threshold">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.QuotaRemindThreshold}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Channel Disable Threshold
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Channel Disable Threshold">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.ChannelDisableThreshold}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Automatic Channel Disable
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Automatic Channel Disable">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.AutomaticDisableChannelEnabled}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
                      <FormLabel className="flex items-center gap-2">
                        Automatic Channel Enable
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground" aria-label="About Automatic Channel Enable">
                              <Info className="h-4 w-4" />
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start" className="max-w-[320px]">
                            {descriptions.AutomaticEnableChannelEnabled}
                          </TooltipContent>
                        </Tooltip>
                      </FormLabel>
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
    </TooltipProvider>
  )
}

export default OperationSettings
