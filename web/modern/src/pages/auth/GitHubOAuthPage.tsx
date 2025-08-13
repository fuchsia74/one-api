import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuthStore } from '@/lib/stores/auth'
import { api } from '@/lib/api'

export function GitHubOAuthPage() {
  const [searchParams] = useSearchParams()
  const [prompt, setPrompt] = useState('Processing GitHub authentication...')
  const navigate = useNavigate()
  const { login } = useAuthStore()

  const sendCode = async (code: string, state: string, retryCount = 0): Promise<void> => {
    try {
      // Unified API call - complete URL with /api prefix
      const response = await api.get(`/api/oauth/github?code=${code}&state=${state}`)
      const { success, message, data } = response.data

      if (success) {
        if (message === 'bind') {
          // Show success toast
          navigate('/settings', {
            state: { message: 'GitHub account bound successfully!' }
          })
        } else {
          login(data, '')

          // Check for redirect_to parameter in the state
          const redirectTo = state && state.includes('redirect_to=')
            ? state.split('redirect_to=')[1]
            : null;

          if (redirectTo) {
            try {
              const decodedPath = decodeURIComponent(redirectTo);
              if (decodedPath.startsWith("/")) {
                navigate(decodedPath, {
                  state: { message: 'GitHub login successful!' }
                });
                return;
              }
            } catch (error) {
              console.error("Invalid redirect_to parameter:", error);
            }
          }

          navigate('/', {
            state: { message: 'GitHub login successful!' }
          })
        }
      } else {
        throw new Error(message || 'GitHub authentication failed')
      }
    } catch (error) {
      if (retryCount >= 3) {
        setPrompt('Authentication failed, redirecting...')
        setTimeout(() => {
          navigate('/login', {
            state: { message: 'GitHub authentication failed. Please try again.' }
          })
        }, 2000)
        return
      }

      const nextRetry = retryCount + 1
      setPrompt(`Authentication error, retrying ${nextRetry}/3...`)

      // Exponential backoff
      const delay = nextRetry * 2000
      setTimeout(() => {
        sendCode(code, state, nextRetry)
      }, delay)
    }
  }

  useEffect(() => {
    const code = searchParams.get('code')
    const state = searchParams.get('state')

    if (!code || !state) {
      navigate('/login', {
        state: { message: 'Invalid GitHub authentication parameters' }
      })
      return
    }

    sendCode(code, state)
  }, [searchParams, navigate])

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">GitHub Authentication</CardTitle>
          <CardDescription>Processing your GitHub login...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            <span className="ml-3 text-sm text-muted-foreground">{prompt}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default GitHubOAuthPage
