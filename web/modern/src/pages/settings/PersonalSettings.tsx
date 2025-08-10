import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { useAuthStore } from '@/lib/stores/auth'
import { api } from '@/lib/api'

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

  const form = useForm<PersonalForm>({
    resolver: zodResolver(personalSchema),
    defaultValues: {
      username: user?.username || '',
      display_name: user?.display_name || '',
      email: user?.email || '',
      password: '',
    },
  })

  const generateAccessToken = async () => {
    try {
      const res = await api.get('/user/token')
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
      const res = await api.get('/user/aff')
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

      const response = await api.put('/user/self', payload)
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
    </div>
  )
}

export default PersonalSettings
