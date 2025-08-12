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

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api
