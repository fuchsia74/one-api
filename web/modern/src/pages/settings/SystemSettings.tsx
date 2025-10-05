import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { api } from '@/lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Info } from 'lucide-react'
import { useNotifications } from '@/components/ui/notifications'

interface OptionRow {
  key: string
  value: string
}

interface OptionGroup {
  id: string
  title: string
  description?: string
  keys: string[]
}

const OPTION_GROUPS: OptionGroup[] = [
  {
    id: 'authentication',
    title: 'Authentication & Registration',
    description: 'Control how users sign up and sign in to your workspace.',
    keys: [
      'PasswordLoginEnabled',
      'PasswordRegisterEnabled',
      'RegisterEnabled',
      'EmailVerificationEnabled',
      'EmailDomainRestrictionEnabled',
      'EmailDomainWhitelist',
    ],
  },
  {
    id: 'oauth',
    title: 'OAuth / SSO Providers',
    description: 'Connect third-party identity providers for seamless sign-in.',
    keys: [
      'GitHubOAuthEnabled',
      'GitHubClientId',
      'GitHubClientSecret',
      'OidcEnabled',
      'OidcClientId',
      'OidcClientSecret',
      'OidcWellKnown',
      'OidcAuthorizationEndpoint',
      'OidcTokenEndpoint',
      'OidcUserinfoEndpoint',
      'LarkClientId',
      'LarkClientSecret',
      'WeChatAuthEnabled',
      'WeChatServerAddress',
      'WeChatServerToken',
      'WeChatAccountQRCodeImageURL',
    ],
  },
  {
    id: 'security',
    title: 'Anti-bot & Security',
    description: 'Configure bot protection and security checks.',
    keys: [
      'TurnstileCheckEnabled',
      'TurnstileSiteKey',
      'TurnstileSecretKey',
    ],
  },
  {
    id: 'email',
    title: 'Email (SMTP)',
    description: 'Set up outbound email delivery.',
    keys: [
      'SMTPServer',
      'SMTPPort',
      'SMTPAccount',
      'SMTPToken',
      'SMTPFrom',
    ],
  },
  {
    id: 'branding',
    title: 'Branding & Content',
    description: 'Customize the look and feel of the product experience.',
    keys: [
      'SystemName',
      'Logo',
      'Footer',
      'Notice',
      'About',
      'HomePageContent',
      'Theme',
    ],
  },
  {
    id: 'links',
    title: 'Links',
    description: 'Control external links exposed to your end users.',
    keys: [
      'TopUpLink',
      'ChatLink',
      'ServerAddress',
    ],
  },
  {
    id: 'quota',
    title: 'Quota & Billing',
    description: 'Manage quotas, billing ratios, and currency presentation.',
    keys: [
      'QuotaForNewUser',
      'QuotaForInviter',
      'QuotaForInvitee',
      'QuotaRemindThreshold',
      'PreConsumedQuota',
      'GroupRatio',
      'QuotaPerUnit',
      'DisplayInCurrencyEnabled',
      'DisplayTokenStatEnabled',
      'ApproximateTokenEnabled',
    ],
  },
  {
    id: 'channels',
    title: 'Channels & Reliability',
    description: 'Automatically react to upstream channel health and retry behavior.',
    keys: [
      'AutomaticDisableChannelEnabled',
      'AutomaticEnableChannelEnabled',
      'ChannelDisableThreshold',
      'RetryTimes',
    ],
  },
  {
    id: 'logging',
    title: 'Logging, Metrics & Integrations',
    description: 'Tune observability and downstream integrations.',
    keys: [
      'LogConsumeEnabled',
      'MessagePusherAddress',
      'MessagePusherToken',
    ],
  },
]

const SENSITIVE_OPTION_KEYS = new Set<string>([
  'SMTPToken',
  'GitHubClientSecret',
  'OidcClientSecret',
  'LarkClientSecret',
  'WeChatServerToken',
  'MessagePusherToken',
])

const OPTION_GROUP_KEY_SET = new Set(OPTION_GROUPS.flatMap((group) => group.keys))

// BOOLEAN_OPTION_KEYS must stay aligned with backend option typing in `model/option.go` and related config defaults.
// Do not rely on string suffix heuristics here—explicitly list each boolean config flag so future options remain typed correctly.
const BOOLEAN_OPTION_KEYS = new Set<string>([
  'PasswordLoginEnabled',
  'PasswordRegisterEnabled',
  'RegisterEnabled',
  'EmailVerificationEnabled',
  'EmailDomainRestrictionEnabled',
  'GitHubOAuthEnabled',
  'OidcEnabled',
  'WeChatAuthEnabled',
  'TurnstileCheckEnabled',
  'AutomaticDisableChannelEnabled',
  'AutomaticEnableChannelEnabled',
  'ApproximateTokenEnabled',
  'LogConsumeEnabled',
  'DisplayInCurrencyEnabled',
  'DisplayTokenStatEnabled',
])

const isBooleanOptionKey = (key: string) => BOOLEAN_OPTION_KEYS.has(key)

export function SystemSettings() {
  const [options, setOptions] = useState<OptionRow[]>([])
  const [loading, setLoading] = useState(false)
  const [hasLoaded, setHasLoaded] = useState(false)
  const { notify } = useNotifications()

  // Map each option key to a concise, user-friendly description for tooltips
  const descriptions = useMemo<Record<string, string>>(
    () => ({
      // Authentication & Registration
      PasswordLoginEnabled: 'Allow users to sign in with email and password. Disable to require OAuth or other providers.',
      PasswordRegisterEnabled: 'Allow new users to register with email and password.',
      RegisterEnabled: 'Master switch for user registration. Turn off to stop all new sign-ups.',
      EmailVerificationEnabled: 'Require email verification for actions like registration or password reset.',
      EmailDomainRestrictionEnabled: 'Only allow registrations from domains listed below.',
      EmailDomainWhitelist: 'Comma-separated allowed email domains (e.g., gmail.com, example.com).',

      // OAuth / SSO Providers
      GitHubOAuthEnabled: 'Enable GitHub OAuth login. Requires GitHub Client ID and Secret.',
      GitHubClientId: 'GitHub OAuth Client ID used for login.',
      GitHubClientSecret: 'GitHub OAuth Client Secret used to exchange authorization codes. Stored securely and never displayed.',
      OidcEnabled: 'Enable OpenID Connect (OIDC) login. Requires OIDC endpoints and credentials.',
      OidcClientId: 'OIDC Client ID used when initiating the OIDC login flow.',
      OidcClientSecret: 'OIDC client secret used during the token exchange. Stored securely and never displayed.',
      OidcWellKnown: 'OIDC well-known discovery URL (e.g., https://issuer/.well-known/openid-configuration).',
      OidcAuthorizationEndpoint: 'OIDC authorization endpoint URL.',
      OidcTokenEndpoint: 'OIDC token endpoint URL.',
      OidcUserinfoEndpoint: 'OIDC userinfo endpoint URL.',
      LarkClientId: 'Lark app ID for Lark OAuth login.',
      LarkClientSecret: 'Lark app secret used for completing the OAuth flow. Stored securely and never displayed.',
      WeChatAuthEnabled: 'Enable WeChat login. Requires WeChat server settings.',
      WeChatServerAddress: 'WeChat login forwarder/server base URL.',
      WeChatServerToken: 'Verification token for your WeChat server integration. Stored securely and never displayed.',
      WeChatAccountQRCodeImageURL: 'URL of the WeChat account QR code image displayed to users.',

      // Anti-bot / Security
      TurnstileCheckEnabled: 'Enable Cloudflare Turnstile on critical actions (registration/reset).',
      TurnstileSiteKey: 'Cloudflare Turnstile site key used by the frontend.',
      TurnstileSecretKey: 'Cloudflare Turnstile secret key used by the server. Keep this confidential.',

      // Email (SMTP)
      SMTPServer: 'SMTP server hostname for sending emails.',
      SMTPPort: 'SMTP server port (e.g., 587 for STARTTLS, 465 for SMTPS).',
      SMTPAccount: 'SMTP username or account email used to authenticate.',
      SMTPToken: 'SMTP password or application token used to authenticate. Stored securely and never displayed.',
      SMTPFrom: 'From address used in outgoing emails (e.g., no-reply@yourdomain).',

      // Branding & Content
      SystemName: 'System display name shown in the UI and emails.',
      Logo: 'Logo URL displayed in the header/login screens.',
      Footer: 'Footer text or HTML displayed site‑wide.',
      Notice: 'Site‑wide announcement content shown to all users.',
      About: 'About page content (markdown/HTML supported by frontend rendering).',
      HomePageContent: 'Content displayed on the home page.',
      Theme: 'UI theme (default, berry, air, modern).',

      // Links
      TopUpLink: 'External link for users to purchase or top up quota.',
      ChatLink: 'External chat/support link shown in the UI.',
      ServerAddress: 'Public base URL of this server (used in links/callbacks).',

      // Quota & Billing
      QuotaForNewUser: 'Initial quota granted to each newly registered user.',
      QuotaForInviter: 'Quota reward granted to the inviter after a successful invite.',
      QuotaForInvitee: 'Quota reward granted to the invitee upon successful registration.',
      QuotaRemindThreshold: 'When remaining quota falls below this value, users will be reminded.',
      PreConsumedQuota: 'Quota reserved at request start to avoid abuse. Unused part is returned after billing.',
      GroupRatio: 'JSON mapping of group‑specific billing ratios. Controls discounts/premiums by user group.',
      QuotaPerUnit: 'Conversion ratio for currency display. Higher value makes each $ represent more quota.',
      DisplayInCurrencyEnabled: 'Show usage and quotas as currency in the UI, based on the configured conversion.',
      DisplayTokenStatEnabled: 'Display token statistics in logs and dashboards when available.',
      ApproximateTokenEnabled: 'Use a faster approximation for token counting to improve performance (may be slightly less accurate).',

      // Channels & Reliability
      AutomaticDisableChannelEnabled: 'Automatically disable channels that show sustained failures.',
      AutomaticEnableChannelEnabled: 'Automatically re‑enable previously disabled channels when healthy.',
      ChannelDisableThreshold: 'Failure rate threshold (percentage) to auto‑disable a channel. Default 5%.',
      RetryTimes: 'Automatic retry attempts for upstream requests on transient errors.',

      // Logging / Metrics / Integrations
      LogConsumeEnabled: 'Record usage/consumption logs. Turn off to reduce storage overhead.',
      MessagePusherAddress: 'Endpoint of the alert/notification pusher service for log events.',
      MessagePusherToken: 'Authentication token for the alert/notification pusher. Stored securely and never displayed.',
    }),
    []
  )

  const load = async () => {
    setLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/option/')
      if (res.data?.success) setOptions(res.data.data || [])
    } finally {
      setLoading(false)
      setHasLoaded(true)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const save = useCallback(async (key: string, value: string) => {
    try {
      // Unified API call - complete URL with /api prefix
      await api.put('/api/option/', { key, value })
      setOptions((prev) => {
        const index = prev.findIndex((opt) => opt.key === key)
        if (index === -1) {
          return [...prev, { key, value }]
        }
        return prev.map((opt) => (opt.key === key ? { ...opt, value } : opt))
      })
      notify({ type: 'success', title: 'Setting saved', message: `${key} updated successfully.` })
    } catch (error: any) {
      console.error('Error saving option:', error)
      const errMsg = error?.response?.data?.message || error?.message || 'Unknown error'
      notify({ type: 'error', title: 'Save failed', message: String(errMsg) })
      throw error
    }
  }, [notify])

  const optionsMap = useMemo(() => {
    const map: Record<string, OptionRow> = {}
    for (const opt of options) {
      map[opt.key] = opt
    }
    return map
  }, [options])

  const uncategorizedOptions = useMemo(
    () => options.filter((opt) => !OPTION_GROUP_KEY_SET.has(opt.key)),
    [options]
  )

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <div>
          <CardTitle>System Settings</CardTitle>
          <CardDescription>
            Configure system-wide settings and options.
          </CardDescription>
        </div>
        <Button variant="outline" onClick={load} disabled={loading}>
          Refresh
        </Button>
      </CardHeader>
      <CardContent>
        {options.length > 0 ? (
          <TooltipProvider>
            <div className="space-y-10">
              {OPTION_GROUPS.map((group) => {
                const groupOptions = group.keys.map((key) => {
                  const option = optionsMap[key] ?? { key, value: '' }
                  return {
                    option,
                    isSensitive: SENSITIVE_OPTION_KEYS.has(key),
                  }
                })

                return (
                  <section key={group.id} className="space-y-4">
                    <div className="space-y-1">
                      <h3 className="text-lg font-semibold leading-6">{group.title}</h3>
                      {group.description && (
                        <p className="text-sm text-muted-foreground">{group.description}</p>
                      )}
                    </div>
                    <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                      {groupOptions.map(({ option, isSensitive }) => (
                        <OptionItem
                          key={option.key}
                          option={option}
                          description={descriptions[option.key]}
                          isSensitive={isSensitive}
                          isBoolean={isBooleanOptionKey(option.key)}
                          onSave={save}
                        />
                      ))}
                    </div>
                  </section>
                )
              })}

              {uncategorizedOptions.length > 0 && (
                <section className="space-y-4">
                  <div className="space-y-1">
                    <h3 className="text-lg font-semibold leading-6">Other Settings</h3>
                    <p className="text-sm text-muted-foreground">Configuration keys that are not yet categorized.</p>
                  </div>
                  <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                    {uncategorizedOptions.map((opt) => (
                      <OptionItem
                        key={opt.key}
                        option={opt}
                        description={descriptions[opt.key]}
                        isSensitive={SENSITIVE_OPTION_KEYS.has(opt.key)}
                        isBoolean={isBooleanOptionKey(opt.key)}
                        onSave={save}
                      />
                    ))}
                  </div>
                </section>
              )}
            </div>
          </TooltipProvider>
        ) : hasLoaded ? (
          <div className="text-center text-sm text-muted-foreground py-8">
            No options available or insufficient permissions.
          </div>
        ) : (
          <div className="text-center text-sm text-muted-foreground py-8">
            Loading options…
          </div>
        )}
      </CardContent>
    </Card>
  )
}

interface OptionItemProps {
  option: OptionRow
  description?: string
  onSave: (key: string, value: string) => Promise<void>
  isSensitive?: boolean
  isBoolean?: boolean
}

function OptionItem({ option, description, onSave, isSensitive, isBoolean }: OptionItemProps) {
  const [value, setValue] = useState(option.value)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    setValue(option.value)
  }, [option.value])

  const handleSave = useCallback(async (overrideValue?: string) => {
    const nextValue = overrideValue ?? value
    if (isSaving || nextValue === option.value) return
    setIsSaving(true)
    try {
      await onSave(option.key, nextValue)
      if (isSensitive) {
        setValue('')
      } else {
        setValue(nextValue)
      }
    } catch (_error) {
      setValue(option.value)
    } finally {
      setIsSaving(false)
    }
  }, [isSaving, isSensitive, onSave, option.key, option.value, value])

  const handleBlur = useCallback(async () => {
    if (value === option.value) return
    await handleSave()
  }, [handleSave, option.value, value])

  const handleBooleanChange = useCallback(
    (newValue: string) => {
      setValue(newValue)
      handleSave(newValue)
    },
    [handleSave]
  )

  const placeholder = isSensitive ? 'Value hidden; enter to update' : undefined

  return (
    <div className="border rounded-lg p-4 space-y-3">
      <div className="text-sm font-medium text-muted-foreground flex items-center gap-2">
        <span>{option.key}</span>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              type="button"
              className="inline-flex items-center text-muted-foreground hover:text-foreground focus:outline-none"
              aria-label={`Info about ${option.key}`}
            >
              <Info className="h-4 w-4" />
            </button>
          </TooltipTrigger>
          <TooltipContent side="top" align="start" className="max-w-[320px]">
            {description || 'No description available for this setting yet.'}
          </TooltipContent>
        </Tooltip>
      </div>
      <div className="flex flex-col gap-2 sm:flex-row">
        {isBoolean ? (
          <Select
            value={value === '' ? undefined : value}
            onValueChange={handleBooleanChange}
            disabled={isSaving}
          >
            <SelectTrigger className="flex-1" aria-label={`${option.key} value`} disabled={isSaving}>
              <SelectValue placeholder="Select value" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="true">Enabled</SelectItem>
              <SelectItem value="false">Disabled</SelectItem>
            </SelectContent>
          </Select>
        ) : (
          <Input
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onBlur={handleBlur}
            className="flex-1"
            aria-label={`${option.key} value`}
            placeholder={placeholder}
            disabled={isSaving}
          />
        )}
        <Button variant="outline" onClick={() => handleSave()} disabled={isSaving}>
          {isSaving ? 'Saving…' : 'Save'}
        </Button>
      </div>
      {isSensitive && (
        <p className="text-xs text-muted-foreground">
          Stored value is hidden. Enter a new value to overwrite the existing secret.
        </p>
      )}
    </div>
  )
}

export default SystemSettings
