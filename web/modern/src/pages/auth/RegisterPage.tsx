import { useEffect, useState } from 'react'
import { useNavigate, Link, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { api } from '@/lib/api'
import { buildGitHubOAuthUrl, getOAuthState } from '@/lib/oauth'

const registerSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  password2: z.string().min(8, 'Password confirmation is required'),
  email: z.string().email('Valid email is required'),
  verification_code: z.string().min(1, 'Verification code is required'),
  aff_code: z.string().optional(),
}).refine((data) => data.password === data.password2, {
  message: "Passwords don't match",
  path: ["password2"],
})

type RegisterForm = z.infer<typeof registerSchema>

export function RegisterPage() {
  const [isLoading, setIsLoading] = useState(false)
  const [isEmailSent, setIsEmailSent] = useState(false)
  const [systemStatus, setSystemStatus] = useState<any>({})
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  // Extract affiliate code from URL parameter
  const affCodeFromUrl = searchParams.get('aff') || ''

  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      username: '',
      password: '',
      password2: '',
      email: '',
      verification_code: '',
      aff_code: affCodeFromUrl,
    },
  })

  // Watch email field to enable/disable send code button
  const emailValue = form.watch('email')

  // Load cached system status (set by App on startup)
  useEffect(() => {
    const status = localStorage.getItem('status')
    if (status) {
      try { setSystemStatus(JSON.parse(status)) } catch { }
    }
  }, [])

  const onGitHubOAuth = async () => {
    const clientId = systemStatus?.github_client_id
    if (!clientId) return
    try {
      const state = await getOAuthState()
      const redirectUri = `${window.location.origin}/oauth/github`
      window.location.href = buildGitHubOAuthUrl(clientId, state, redirectUri)
    } catch (e) {
      const redirectUri = `${window.location.origin}/oauth/github`
      window.location.href = buildGitHubOAuthUrl(clientId, '', redirectUri)
    }
  }

  const sendVerificationCode = async () => {
    const email = form.getValues('email')

    // Simple email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!email || !emailRegex.test(email)) {
      form.setError('email', { message: 'Please enter a valid email address' })
      return
    }

    try {
      setIsLoading(true)
      // Unified API call - complete URL with /api prefix
      const response = await api.get(`/api/verification?email=${encodeURIComponent(email)}`)
      const { success, message } = response.data

      if (success) {
        setIsEmailSent(true)
        form.clearErrors('email')
      } else {
        form.setError('email', { message: message || 'Failed to send verification code' })
      }
    } catch (error) {
      form.setError('email', {
        message: error instanceof Error ? error.message : 'Failed to send verification code'
      })
    } finally {
      setIsLoading(false)
    }
  }

  const onSubmit = async (data: RegisterForm) => {
    setIsLoading(true)
    try {
      const payload = {
        username: data.username,
        password: data.password,
        email: data.email,
        verification_code: data.verification_code,
        ...(data.aff_code && { aff_code: data.aff_code }),
      }

      // Unified API call - complete URL with /api prefix
      const response = await api.post('/api/user/register', payload)
      const { success, message } = response.data

      if (success) {
        navigate('/login', {
          state: { message: 'Registration successful! Please login with your credentials.' }
        })
      } else {
        form.setError('root', { message: message || 'Registration failed' })
      }
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Registration failed'
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Create Account</CardTitle>
          <CardDescription>Sign up to get started</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter username" {...field} />
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
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Enter password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="password2"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Confirm Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Confirm password" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <div className="flex gap-2">
                        <Input
                          type="email"
                          placeholder="Enter email"
                          {...field}
                          className="flex-1"
                        />
                        <Button
                          type="button"
                          variant="outline"
                          onClick={sendVerificationCode}
                          disabled={isLoading || !emailValue || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(emailValue)}
                        >
                          {isLoading ? 'Sending...' : isEmailSent ? 'Sent' : 'Send Code'}
                        </Button>
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="verification_code"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Verification Code</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter verification code" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="aff_code"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Invitation Code (Optional)</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter invitation code" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {form.formState.errors.root && (
                <div className="text-sm text-destructive">
                  {form.formState.errors.root.message}
                </div>
              )}

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? 'Creating Account...' : 'Create Account'}
              </Button>

              {systemStatus?.github_oauth && (
                <div className="space-y-2">
                  <div className="relative">
                    <div className="absolute inset-0 flex items-center">
                      <span className="w-full border-t" />
                    </div>
                    <div className="relative flex justify-center text-xs">
                      <span className="bg-card px-2 text-muted-foreground">Or sign up with</span>
                    </div>
                  </div>
                  <Button type="button" variant="outline" className="w-full" onClick={onGitHubOAuth}>
                    <svg className="w-4 h-4 mr-2" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                    </svg>
                    GitHub
                  </Button>
                </div>
              )}

              <div className="text-center text-sm">
                Already have an account?{' '}
                <Link to="/login" className="text-primary hover:underline">
                  Sign in
                </Link>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  )
}

export default RegisterPage
