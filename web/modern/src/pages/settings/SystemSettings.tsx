import { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { api } from '@/lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { Info } from 'lucide-react'

interface OptionRow {
  key: string
  value: string
}

export function SystemSettings() {
  const [options, setOptions] = useState<OptionRow[]>([])
  const [loading, setLoading] = useState(false)

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
      OidcEnabled: 'Enable OpenID Connect (OIDC) login. Requires OIDC endpoints and credentials.',
      OidcClientId: 'OIDC Client ID used when initiating the OIDC login flow.',
      OidcWellKnown: 'OIDC well-known discovery URL (e.g., https://issuer/.well-known/openid-configuration).',
      OidcAuthorizationEndpoint: 'OIDC authorization endpoint URL.',
      OidcTokenEndpoint: 'OIDC token endpoint URL.',
      OidcUserinfoEndpoint: 'OIDC userinfo endpoint URL.',
      LarkClientId: 'Lark app ID for Lark OAuth login.',
      WeChatAuthEnabled: 'Enable WeChat login. Requires WeChat server settings.',
      WeChatServerAddress: 'WeChat login forwarder/server base URL.',
      WeChatAccountQRCodeImageURL: 'URL of the WeChat account QR code image displayed to users.',

      // Anti-bot / Security
      TurnstileCheckEnabled: 'Enable Cloudflare Turnstile on critical actions (registration/reset).',
      TurnstileSiteKey: 'Cloudflare Turnstile site key used by the frontend.',
      TurnstileSecretKey: 'Cloudflare Turnstile secret key used by the server. Keep this confidential.',

      // Email (SMTP)
      SMTPServer: 'SMTP server hostname for sending emails.',
      SMTPPort: 'SMTP server port (e.g., 587 for STARTTLS, 465 for SMTPS).',
      SMTPAccount: 'SMTP username or account email used to authenticate.',
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
      MessagePusherToken: 'Authentication token for the alert/notification pusher. Keep this confidential.',
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
    }
  }

  useEffect(() => {
    load()
  }, [])

  const save = async (key: string, value: string) => {
    try {
      // Unified API call - complete URL with /api prefix
      await api.put('/api/option/', { key, value })
      // Show success message
      console.log(`Saved ${key}: ${value}`)
    } catch (error) {
      console.error('Error saving option:', error)
    }
  }

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
        <TooltipProvider>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {options.map((opt, idx) => (
              <div key={idx} className="border rounded-lg p-4">
                <div className="text-sm font-medium text-muted-foreground mb-2 flex items-center gap-2">
                  <span>{opt.key}</span>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button
                        type="button"
                        className="inline-flex items-center text-muted-foreground hover:text-foreground focus:outline-none"
                        aria-label={`Info about ${opt.key}`}
                      >
                        <Info className="h-4 w-4" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent side="top" align="start" className="max-w-[320px]">
                      {descriptions[opt.key] || 'No description available for this setting yet.'}
                    </TooltipContent>
                  </Tooltip>
                </div>
                <div className="flex gap-2">
                  <Input
                    defaultValue={opt.value}
                    onBlur={(e) => save(opt.key, e.target.value)}
                    className="flex-1"
                  />
                  <Button
                    variant="outline"
                    onClick={(e) => {
                      const target = (e.currentTarget.previousSibling as HTMLInputElement)
                      save(opt.key, target.value)
                    }}
                  >
                    Save
                  </Button>
                </div>
              </div>
            ))}
            {!options.length && (
              <div className="col-span-full text-center text-sm text-muted-foreground py-8">
                No options available or insufficient permissions.
              </div>
            )}
          </div>
        </TooltipProvider>
      </CardContent>
    </Card>
  )
}

export default SystemSettings
