import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { api } from '@/lib/api'

const resetConfirmSchema = z.object({
  password: z.string().min(8, 'Password must be at least 8 characters'),
  password2: z.string().min(8, 'Password confirmation is required'),
}).refine((data) => data.password === data.password2, {
  message: "Passwords don't match",
  path: ["password2"],
})

type ResetConfirmForm = z.infer<typeof resetConfirmSchema>

export function PasswordResetConfirmPage() {
  const [isLoading, setIsLoading] = useState(false)
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const email = searchParams.get('email')
  const token = searchParams.get('token')

  const form = useForm<ResetConfirmForm>({
    resolver: zodResolver(resetConfirmSchema),
    defaultValues: { password: '', password2: '' },
  })

  useEffect(() => {
    if (!email || !token) {
      navigate('/login', {
        state: { message: 'Invalid or missing reset parameters' }
      })
    }
  }, [email, token, navigate])

  const onSubmit = async (data: ResetConfirmForm) => {
    if (!email || !token) return

    setIsLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      const response = await api.post('/api/user/reset', {
        email,
        token,
        password: data.password,
      })

      const { success, message } = response.data

      if (success) {
        navigate('/login', {
          state: { message: 'Password reset successful! Please login with your new password.' }
        })
      } else {
        form.setError('root', { message: message || 'Password reset failed' })
      }
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Password reset failed'
      })
    } finally {
      setIsLoading(false)
    }
  }

  if (!email || !token) {
    return null // Will redirect in useEffect
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Set New Password</CardTitle>
          <CardDescription>
            Enter your new password for {email}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>New Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Enter new password" {...field} />
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
                    <FormLabel>Confirm New Password</FormLabel>
                    <FormControl>
                      <Input type="password" placeholder="Confirm new password" {...field} />
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
                {isLoading ? 'Updating Password...' : 'Update Password'}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  )
}

export default PasswordResetConfirmPage
