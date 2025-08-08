import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { useAuthStore } from '@/lib/stores/auth'
import api from '@/lib/api'

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
  totp_code: z.string().optional(),
})

type LoginForm = z.infer<typeof loginSchema>

export function LoginPage() {
  const [isLoading, setIsLoading] = useState(false)
  const [totpRequired, setTotpRequired] = useState(false)
  const totpRef = useRef<HTMLInputElement | null>(null)
  const navigate = useNavigate()
  const { login } = useAuthStore()

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: '', password: '', totp_code: '' },
  })

  const onSubmit = async (data: LoginForm) => {
    setIsLoading(true)
    try {
      const payload: Record<string, string> = { username: data.username, password: data.password }
      if (totpRequired && data.totp_code) payload.totp_code = data.totp_code
      const response = await api.post('/user/login', payload)
      const { success, message, data: respData } = response.data
      const m = typeof message === 'string' ? message.trim().toLowerCase() : ''
      const dataTotp = !!(respData && (respData.totp_required === true || respData.totp_required === 'true' || respData.totp_required === 1))
      const needsTotp = (!success) && (dataTotp || m === 'totp_required' || m.includes('totp'))
      if (needsTotp) {
        setTotpRequired(true)
        form.setValue('totp_code', '')
        form.setError('root', { message: 'Please enter your TOTP code' })
        return
      }
      if (success) {
        login(respData, '')
        navigate('/dashboard')
      } else {
        form.setError('root', { message: m === 'totp_required' ? 'Please enter your TOTP code' : (message || 'Login failed') })
      }
    } catch (error) {
      form.setError('root', { message: error instanceof Error ? error.message : 'Login failed' })
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => { if (totpRequired && totpRef.current) totpRef.current.focus() }, [totpRequired])

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Sign In</CardTitle>
          <CardDescription>Enter your credentials to access your account</CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField control={form.control} name="username" render={({ field }) => (
                <FormItem>
                  <FormLabel>Username</FormLabel>
                  <FormControl><Input {...field} disabled={totpRequired} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="password" render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl><Input type="password" {...field} disabled={totpRequired} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              {totpRequired && (
                <FormField control={form.control} name="totp_code" render={({ field }) => (
                  <FormItem>
                    <FormLabel>TOTP Code</FormLabel>
                    <FormControl>
                      <Input maxLength={6} placeholder="Enter 6-digit TOTP code" {...field} ref={totpRef} inputMode="numeric" pattern="[0-9]*" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )} />
              )}
              {form.formState.errors.root && (
                <div className="text-sm text-destructive">{totpRequired ? 'Please enter your TOTP code' : form.formState.errors.root.message}</div>
              )}
              <Button type="submit" className="w-full" disabled={isLoading || (totpRequired && ((form.getValues().totp_code ?? '').length !== 6))}>
                {isLoading ? 'Signing in...' : totpRequired ? 'Verify TOTP' : 'Sign In'}
              </Button>
              {totpRequired && (
                <Button type="button" variant="outline" className="w-full" onClick={() => { setTotpRequired(false); form.setValue('totp_code', ''); form.clearErrors('root') }}>
                  Back to Login
                </Button>
              )}
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  )
}

export default LoginPage
