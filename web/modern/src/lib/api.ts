import axios from 'axios'

// Unified API client - callers must provide complete URLs including /api prefix
// This eliminates ambiguity and ensures consistency across all API calls
export const api = axios.create({
  timeout: 10000,
})

// Always send cookies for session-based auth
api.defaults.withCredentials = true

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // For session-based authentication, we rely on cookies (withCredentials: true)
    // Only add Authorization header for specific API endpoints that require token auth
    // Most dashboard/web endpoints use session-based auth via cookies
    const token = localStorage.getItem('token')
    if (token && config.url?.startsWith('/v1/')) {
      // Only add token for API endpoints that require token authentication
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Helper function to handle authentication failures
const handleAuthFailure = () => {
  // Clear auth data
  localStorage.removeItem('token')
  localStorage.removeItem('user')

  // Get current path for redirect_to parameter
  const currentPath = window.location.pathname + window.location.search
  const redirectTo = encodeURIComponent(currentPath)

  // Redirect to login with redirect_to parameter
  window.location.href = `/login?redirect_to=${redirectTo}`
}

// Response interceptor
api.interceptors.response.use(
  (response) => {
    // TEMPORARY: Handle legacy 200 OK with success: false for auth errors
    // TODO: Remove this once all backend endpoints return proper HTTP status codes
    if (response.data && response.data.success === false) {
      const message = response.data.message || ''
      const isAuthError = message.includes('access token is invalid') ||
                         message.includes('not logged in') ||
                         message.includes('No permission to perform this operation')

      if (isAuthError) {
        handleAuthFailure()
        return response // Return to prevent further processing
      }
    }
    return response
  },
  (error) => {
    // Handle proper HTTP status codes for authentication/authorization failures
    if (error.response?.status === 401 || error.response?.status === 403) {
      handleAuthFailure()
    }
    return Promise.reject(error)
  }
)

export default api
