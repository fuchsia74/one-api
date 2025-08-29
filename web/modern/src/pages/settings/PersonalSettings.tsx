import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import QRCode from 'qrcode'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { useAuthStore } from '@/lib/stores/auth'
import { api } from '@/lib/api'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'

const personalSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  display_name: z.string().optional(),
  email: z.string().email('Valid email is required').optional(),
  password: z.string().optional(),
})

type PersonalForm = z.infer<typeof personalSchema>

export function PersonalSettings() {
  const { user } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [systemToken, setSystemToken] = useState('')
  const [affLink, setAffLink] = useState('')

  // TOTP related state
  const [totpEnabled, setTotpEnabled] = useState(false)
  const [showTotpSetup, setShowTotpSetup] = useState(false)
  const [totpSecret, setTotpSecret] = useState('')
  const [totpQRCode, setTotpQRCode] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [totpLoading, setTotpLoading] = useState(false)
  const [totpError, setTotpError] = useState('')
  const [setupTotpError, setSetupTotpError] = useState('')
  const [confirmTotpError, setConfirmTotpError] = useState('')
  const [disableTotpError, setDisableTotpError] = useState('')

  const form = useForm<PersonalForm>({
    resolver: zodResolver(personalSchema),
    defaultValues: {
      username: user?.username || '',
      display_name: user?.display_name || '',
      email: user?.email || '',
      password: '',
    },
  })

  // Load TOTP status when component mounts
  const loadTotpStatus = async () => {
    try {
      setTotpError('') // Clear previous error
      const res = await api.get('/api/user/totp/status')
      if (res.data.success) {
        setTotpEnabled(res.data.data.totp_enabled)
      } else {
        setTotpError(res.data.message || 'Failed to load TOTP status')
      }
    } catch (error) {
      setTotpError(error instanceof Error ? error.message : 'Failed to load TOTP status')
    }
  }

  useEffect(() => {
    loadTotpStatus()
  }, [])

  // Setup TOTP for the user
  const setupTotp = async () => {
    setTotpLoading(true)
    setSetupTotpError('') // Clear previous error
    try {
      const res = await api.get('/api/user/totp/setup')
      if (res.data.success) {
        setTotpSecret(res.data.data.secret)
        // Generate QR code from URI
        const qrCodeDataURL = await QRCode.toDataURL(res.data.data.qr_code)
        setTotpQRCode(qrCodeDataURL)
        setShowTotpSetup(true)
      } else {
        setSetupTotpError(res.data.message || 'Failed to setup TOTP')
      }
    } catch (error) {
      setSetupTotpError(error instanceof Error ? error.message : 'Failed to setup TOTP')
    }
    setTotpLoading(false)
  }

  // Confirm TOTP setup with verification code
  const confirmTotp = async () => {
    setConfirmTotpError('') // Clear previous error
    if (!totpCode) {
      setConfirmTotpError('Please enter the TOTP code')
      return
    }

    setTotpLoading(true)
    try {
      const res = await api.post('/api/user/totp/confirm', {
        totp_code: totpCode,
      })

      if (res.data.success) {
        // Success - clear error and update state
        setConfirmTotpError('')
        setTotpEnabled(true)
        setShowTotpSetup(false)
        setTotpCode('')
      } else {
        setConfirmTotpError(res.data.message || 'Failed to confirm TOTP')
      }
    } catch (error) {
      setConfirmTotpError(error instanceof Error ? error.message : 'Failed to confirm TOTP')
    }
    setTotpLoading(false)
  }

  // Disable TOTP for the user
  const disableTotp = async () => {
    setDisableTotpError('') // Clear previous error
    if (!totpCode) {
      setDisableTotpError('Please enter the TOTP code to disable')
      return
    }

    setTotpLoading(true)
    try {
      const res = await api.post('/api/user/totp/disable', {
        totp_code: totpCode,
      })

      if (res.data.success) {
        // Success - clear error and update state
        setDisableTotpError('')
        setTotpEnabled(false)
        setTotpCode('')
      } else {
        setDisableTotpError(res.data.message || 'Failed to disable TOTP')
      }
    } catch (error) {
      setDisableTotpError(error instanceof Error ? error.message : 'Failed to disable TOTP')
    }
    setTotpLoading(false)
  }

  const generateAccessToken = async () => {
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/user/token')
      const { success, message, data } = res.data
      if (success) {
        setSystemToken(data)
        setAffLink('')
        // Copy to clipboard
        await navigator.clipboard.writeText(data)
        // Show success message
      } else {
        console.error('Failed to generate token:', message)
      }
    } catch (error) {
      console.error('Error generating token:', error)
    }
  }

  const getAffLink = async () => {
    try {
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/user/aff')
      const { success, message, data } = res.data
      if (success) {
        const link = `${window.location.origin}/register?aff=${data}`
        setAffLink(link)
        setSystemToken('')
        // Copy to clipboard
        await navigator.clipboard.writeText(link)
        // Show success message
      } else {
        console.error('Failed to get aff link:', message)
      }
    } catch (error) {
      console.error('Error getting aff link:', error)
    }
  }

  const onSubmit = async (data: PersonalForm) => {
    setLoading(true)
    try {
      const payload = { ...data }
      // Don't send empty password
      if (!payload.password) {
        delete payload.password
      }

      // Unified API call - complete URL with /api prefix
      const response = await api.put('/api/user/self', payload)
      const { success, message } = response.data
      if (success) {
        // Show success message
        console.log('Profile updated successfully')
        // Update the form to clear password
        form.setValue('password', '')
      } else {
        form.setError('root', { message: message || 'Update failed' })
      }
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Update failed'
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Profile Information</CardTitle>
          <CardDescription>
            Update your personal information and account settings.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Username</FormLabel>
                      <FormControl>
                        <Input {...field} disabled />
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
                      <FormLabel>Display Name</FormLabel>
                      <FormControl>
                        <Input placeholder="Enter display name" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Email</FormLabel>
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
                      <FormLabel>New Password (leave empty to keep current)</FormLabel>
                      <FormControl>
                        <Input type="password" placeholder="Enter new password" {...field} />
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

              <Button type="submit" disabled={loading}>
                {loading ? 'Updating...' : 'Update Profile'}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Access Token & Invitation</CardTitle>
          <CardDescription>
            Generate access tokens and invitation links.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Button onClick={generateAccessToken} className="w-full">
                Generate Access Token
              </Button>
              {systemToken && (
                <div className="mt-2 p-2 bg-muted rounded text-sm font-mono break-all">
                  {systemToken}
                </div>
              )}
            </div>

            <div>
              <Button onClick={getAffLink} variant="outline" className="w-full">
                Get Invitation Link
              </Button>
              {affLink && (
                <div className="mt-2 p-2 bg-muted rounded text-sm break-all">
                  {affLink}
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Two-Factor Authentication (TOTP)</CardTitle>
          <CardDescription>
            Enhance your account security with two-factor authentication.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {totpError && (
            <div className="text-sm text-destructive font-medium mb-2">
              {totpError}
            </div>
          )}
          {totpEnabled ? (
            <Alert className="bg-green-50 dark:bg-green-950/30 border-green-200 dark:border-green-900">
              <div className="flex flex-col space-y-4">
                <div>
                  <AlertTitle className="text-green-800 dark:text-green-300">TOTP is enabled</AlertTitle>
                  <AlertDescription>
                    Your account is protected with two-factor authentication.
                  </AlertDescription>
                </div>
                <div className="flex flex-col space-y-2">
                  <Input
                    placeholder="Enter TOTP code to disable"
                    value={totpCode}
                    onChange={(e) => setTotpCode(e.target.value)}
                  />
                  {disableTotpError && (
                    <div className="text-sm text-destructive font-medium">
                      {disableTotpError}
                    </div>
                  )}
                  <Button
                    variant="destructive"
                    onClick={disableTotp}
                    disabled={totpLoading}
                    className="w-full md:w-auto"
                  >
                    {totpLoading ? 'Processing...' : 'Disable TOTP'}
                  </Button>
                </div>
              </div>
            </Alert>
          ) : (
            <Alert className="bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-900">
              <div className="flex flex-col space-y-4">
                <div>
                  <AlertTitle className="text-blue-800 dark:text-blue-300">TOTP is not enabled</AlertTitle>
                  <AlertDescription>
                    Enable two-factor authentication to secure your account.
                  </AlertDescription>
                </div>
                {setupTotpError && (
                  <div className="text-sm text-destructive font-medium">
                    {setupTotpError}
                  </div>
                )}
                <div>
                  <Button
                    variant="default"
                    onClick={setupTotp}
                    disabled={totpLoading}
                    className="w-full md:w-auto"
                  >
                    {totpLoading ? 'Processing...' : 'Enable TOTP'}
                  </Button>
                </div>
              </div>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* TOTP Setup Dialog */}
      <Dialog open={showTotpSetup} onOpenChange={(open) => !totpLoading && setShowTotpSetup(open)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Setup Two-Factor Authentication</DialogTitle>
            <DialogDescription>
              Follow these steps to secure your account with TOTP.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <Alert>
              <AlertTitle>Setup Instructions</AlertTitle>
              <AlertDescription>
                <ol className="pl-4 mt-2 space-y-1">
                  <li>Install an authenticator app (Google Authenticator, Authy, etc.)</li>
                  <li>Scan the QR code below or manually enter the secret key</li>
                  <li>Enter the 6-digit code from your authenticator app</li>
                  <li>Click "Confirm" to enable TOTP</li>
                </ol>
              </AlertDescription>
            </Alert>

            {totpQRCode && (
              <div className="flex justify-center my-4">
                <img src={totpQRCode} alt="TOTP QR Code" className="w-48 h-48" />
              </div>
            )}

            <div className="space-y-2">
              <FormLabel>Secret Key (manual entry)</FormLabel>
              <Input value={totpSecret} readOnly className="font-mono" />
            </div>

            <div className="space-y-2">
              <FormLabel>Verification Code</FormLabel>
              <Input
                placeholder="Enter 6-digit code from your authenticator app"
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                maxLength={6}
              />
              {confirmTotpError && (
                <div className="text-sm text-destructive font-medium mt-1">
                  {confirmTotpError}
                </div>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowTotpSetup(false)}
              disabled={totpLoading}
            >
              Cancel
            </Button>
            <Button
              onClick={confirmTotp}
              disabled={!totpCode || totpCode.length !== 6 || totpLoading}
            >
              {totpLoading ? 'Processing...' : 'Confirm'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default PersonalSettings
