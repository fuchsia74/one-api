import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { useAuthStore } from '@/lib/stores/auth'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { api } from '@/lib/api'
import { DataTable } from '@/components/ui/data-table'
import { ResponsivePageContainer } from '@/components/ui/responsive-container'
import { useResponsive } from '@/hooks/useResponsive'
import { formatNumber, cn } from '@/lib/utils'

// Quota conversion utility
const getQuotaPerUnit = () => parseFloat(localStorage.getItem('quota_per_unit') || '500000')
const getDisplayInCurrency = () => localStorage.getItem('display_in_currency') === 'true'

const renderQuota = (quota: number, precision: number = 2): string => {
  const displayInCurrency = getDisplayInCurrency()
  const quotaPerUnit = getQuotaPerUnit()

  if (displayInCurrency) {
    const amount = (quota / quotaPerUnit).toFixed(precision)
    return `$${amount}`
  }

  return formatNumber(quota)
}
import {
  ResponsiveContainer,
  LineChart,
  Line,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  BarChart,
  Bar,
  Legend,
} from 'recharts'

// Gradient definitions component for charts
const GradientDefs = () => (
  <defs>
    <linearGradient id="requestsGradient" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stopColor="#4318FF" stopOpacity={0.8} />
      <stop offset="100%" stopColor="#4318FF" stopOpacity={0.1} />
    </linearGradient>
    <linearGradient id="quotaGradient" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stopColor="#00B5D8" stopOpacity={0.8} />
      <stop offset="100%" stopColor="#00B5D8" stopOpacity={0.1} />
    </linearGradient>
    <linearGradient id="tokensGradient" x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stopColor="#FF5E7D" stopOpacity={0.8} />
      <stop offset="100%" stopColor="#FF5E7D" stopOpacity={0.1} />
    </linearGradient>
  </defs>
)

// Chart configuration
const chartConfig = {
  colors: {
    requests: '#4318FF',
    quota: '#00B5D8',
    tokens: '#FF5E7D',
  },
  gradients: {
    requests: 'url(#requestsGradient)',
    quota: 'url(#quotaGradient)',
    tokens: 'url(#tokensGradient)',
  },
  lineChart: {
    strokeWidth: 3,
    dot: false,
    activeDot: {
      r: 6,
      strokeWidth: 2,
      filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))'
    },
    grid: {
      vertical: false,
      horizontal: true,
      opacity: 0.2,
    },
  },
  barColors: [
    '#4318FF', // Deep purple
    '#00B5D8', // Cyan
    '#6C63FF', // Purple
    '#05CD99', // Green
    '#FFB547', // Orange
    '#FF5E7D', // Pink
    '#41B883', // Emerald
    '#7983FF', // Light Purple
    '#FF8F6B', // Coral
    '#49BEFF', // Sky Blue
    '#8B5CF6', // Violet
    '#F59E0B', // Amber
    '#EF4444', // Red
    '#10B981', // Emerald
    '#3B82F6', // Blue
  ],
}

export function DashboardPage() {
  const { user } = useAuthStore()
  const { isMobile } = useResponsive()
  const isAdmin = useMemo(() => (user?.role ?? 0) >= 10, [user])

  // date range defaults: last 7 days (inclusive)
  const fmt = (d: Date) => d.toISOString().slice(0, 10)
  const today = new Date()
  const last7 = new Date(today)
  last7.setDate(today.getDate() - 6)

  const [fromDate, setFromDate] = useState(fmt(last7))
  const [toDate, setToDate] = useState(fmt(today))
  const [dashUser, setDashUser] = useState<string>('all')
  const [userOptions, setUserOptions] = useState<Array<{ id: number; username: string; display_name: string }>>([])
  const [loading, setLoading] = useState(false)
  const [lastUpdated, setLastUpdated] = useState<string | null>(null)
  const [statisticsMetric, setStatisticsMetric] = useState<'tokens' | 'requests' | 'expenses'>('tokens')
  const [dateError, setDateError] = useState<string>('')

  type Row = { day: string; model_name: string; request_count: number; quota: number; prompt_tokens: number; completion_tokens: number }
  const [rows, setRows] = useState<Row[]>([])

  // Quota/Status panel state
  const [quotaStats, setQuotaStats] = useState<{ totalQuota: number; usedQuota: number; status: string }>({ totalQuota: 0, usedQuota: 0, status: '-' })

  // Fetch quota/status for selected user or all
  const fetchQuotaStats = async (userId: string) => {
    try {
      let res
      if (isAdmin) {
        // Unified API call - complete URL with /api prefix
        if (userId === 'all') {
          res = await api.get('/api/user/dashboard')
        } else {
          res = await api.get(`/api/user/dashboard?user_id=${userId}`)
        }
      } else {
        res = await api.get('/api/user/dashboard')
      }
      if (res.data?.success && res.data.data) {
        setQuotaStats({
          totalQuota: res.data.data.total_quota ?? 0,
          usedQuota: res.data.data.used_quota ?? 0,
          status: res.data.data.status ?? '-',
        })
      } else {
        setQuotaStats({ totalQuota: 0, usedQuota: 0, status: '-' })
      }
    } catch {
      setQuotaStats({ totalQuota: 0, usedQuota: 0, status: '-' })
    }
  }



  // Date validation functions
  const getMaxDate = () => {
    const today = new Date()
    return today.toISOString().split('T')[0]
  }

  const getMinDate = () => {
    if (isAdmin) {
      // Admin users can go back 1 year
      const oneYearAgo = new Date()
      oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1)
      return oneYearAgo.toISOString().split('T')[0]
    } else {
      // Regular users can only go back 7 days from today
      const sevenDaysAgo = new Date()
      sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7)
      return sevenDaysAgo.toISOString().split('T')[0]
    }
  }

  // Date validation
  const validateDateRange = (from: string, to: string): string => {
    if (!from || !to) return ''

    const fromDate = new Date(from)
    const toDate = new Date(to)
    const today = new Date()
    const minDate = new Date(getMinDate())

    if (fromDate > toDate) {
      return 'From date must be before or equal to To date'
    }

    if (toDate > today) {
      return 'To date cannot be in the future'
    }

    if (fromDate < minDate) {
      return isAdmin
        ? 'From date cannot be more than 1 year ago'
        : 'From date cannot be more than 7 days ago'
    }

    const daysDiff = Math.ceil((toDate.getTime() - fromDate.getTime()) / (1000 * 60 * 60 * 24))
    const maxDays = isAdmin ? 365 : 7

    if (daysDiff > maxDays) {
      return isAdmin
        ? 'Date range cannot exceed 1 year'
        : 'Date range cannot exceed 7 days'
    }

    return ''
  }

  const loadUsers = async () => {
    if (!isAdmin) return
    // Unified API call - complete URL with /api prefix
    const res = await api.get('/api/user/dashboard/users')
    if (res.data?.success) {
      setUserOptions(res.data.data || [])
    }
  }

  const loadStats = async () => {
    // Validate date range before making API call
    const validationError = validateDateRange(fromDate, toDate)
    if (validationError) {
      setDateError(validationError)
      return
    }

    setLoading(true)
    setDateError('')
    try {
      const params = new URLSearchParams()
      params.set('from_date', fromDate)
      params.set('to_date', toDate)
      if (isAdmin) {
        params.set('user_id', dashUser || 'all')
      }
      // Unified API call - complete URL with /api prefix
      const res = await api.get('/api/user/dashboard?' + params.toString())
      const { success, data, message } = res.data
      if (success) {
        // Handle new API response structure
        const logs = data?.logs || data || []
        setRows(
          logs.map((row: any) => ({
            day: row.Day,
            model_name: row.ModelName,
            request_count: row.RequestCount,
            quota: row.Quota,
            prompt_tokens: row.PromptTokens,
            completion_tokens: row.CompletionTokens,
          }))
        )

        // Update quota stats if available in the response
        if (data && typeof data === 'object' && 'total_quota' in data) {
          setQuotaStats({
            totalQuota: data.total_quota ?? 0,
            usedQuota: data.used_quota ?? 0,
            status: data.status ?? '-',
          })
        }

        setLastUpdated(new Date().toLocaleTimeString())
        setDateError('')
      } else {
        setDateError(message || 'Failed to fetch dashboard data')
        setRows([])
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error)
      setDateError('Failed to fetch dashboard data')
      setRows([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (isAdmin) loadUsers()
    // load initial stats and quota
    loadStats()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAdmin])

  // Fetch quota/status when dashUser changes (but not on initial load)
  useEffect(() => {
    if (isAdmin) {
      fetchQuotaStats(dashUser)
    } else {
      fetchQuotaStats('self')
    }
  }, [dashUser, isAdmin])

  // Presets
  const applyPreset = (preset: 'today' | '7d' | '30d') => {
    const today = new Date()
    const start = new Date(today)
    if (preset === 'today') start.setDate(today.getDate())
    if (preset === '7d') start.setDate(today.getDate() - 6)
    if (preset === '30d') start.setDate(today.getDate() - 29)
    setFromDate(fmt(start))
    setToDate(fmt(today))
    // defer fetch until user clicks Apply/Refresh to avoid surprise reloads
  }

  // Build time-series by day totals
  const days = useMemo(() => Array.from(new Set(rows.map(r => r.day))).sort(), [rows])
  const timeSeries = useMemo(() => {
    // Aggregate across models per day
    const quotaPerUnit = getQuotaPerUnit()
    const map: Record<string, { date: string; requests: number; quota: number; tokens: number }> = {}
    for (const r of rows) {
      if (!map[r.day]) map[r.day] = { date: r.day, requests: 0, quota: 0, tokens: 0 }
      map[r.day].requests += r.request_count || 0
      map[r.day].quota += (r.quota || 0) / quotaPerUnit // Convert to USD
      map[r.day].tokens += (r.prompt_tokens || 0) + (r.completion_tokens || 0)
    }
    return Object.values(map).sort((a, b) => a.date.localeCompare(b.date))
  }, [rows])

  // Model stacked bar per day - supports different metrics
  const uniqueModels = useMemo(() => Array.from(new Set(rows.map(r => r.model_name))), [rows])
  const stackedData = useMemo(() => {
    const quotaPerUnit = getQuotaPerUnit()
    const map: Record<string, Record<string, number>> = {}
    for (const d of days) map[d] = {}
    for (const r of rows) {
      let value: number
      switch (statisticsMetric) {
        case 'requests':
          value = r.request_count || 0
          break
        case 'expenses':
          value = (r.quota || 0) / quotaPerUnit
          break
        case 'tokens':
        default:
          value = (r.prompt_tokens || 0) + (r.completion_tokens || 0)
          break
      }
      map[r.day][r.model_name] = (map[r.day][r.model_name] || 0) + value
    }
    return days.map(d => ({ date: d, ...(map[d] || {}) }))
  }, [rows, days, statisticsMetric])

  const columns: ColumnDef<Row>[] = [
    { header: 'Day', accessorKey: 'day' },
    { header: 'Model', accessorKey: 'model_name' },
    { header: 'Requests', accessorKey: 'request_count' },
    { header: 'Quota', accessorKey: 'quota' },
    { header: 'Prompt Tokens', accessorKey: 'prompt_tokens' },
    { header: 'Completion Tokens', accessorKey: 'completion_tokens' },
  ]

  // Summaries & insights
  const dailyAgg = useMemo(() => {
    const quotaPerUnit = getQuotaPerUnit()
    const m: Record<string, { date: string; requests: number; quota: number; tokens: number }> = {}
    for (const r of rows) {
      if (!m[r.day]) m[r.day] = { date: r.day, requests: 0, quota: 0, tokens: 0 }
      m[r.day].requests += r.request_count || 0
      m[r.day].quota += (r.quota || 0) / quotaPerUnit // Convert to USD
      m[r.day].tokens += (r.prompt_tokens || 0) + (r.completion_tokens || 0)
    }
    return Object.values(m).sort((a, b) => a.date.localeCompare(b.date))
  }, [rows])

  const todayAgg = dailyAgg.length ? dailyAgg[dailyAgg.length - 1] : { requests: 0, quota: 0, tokens: 0 }
  const prevAgg = dailyAgg.length > 1 ? dailyAgg[dailyAgg.length - 2] : { requests: 0, quota: 0, tokens: 0 }
  const pct = (cur: number, prev: number) => (prev > 0 ? ((cur - prev) / prev) * 100 : 0)
  const requestTrend = pct(todayAgg.requests, prevAgg.requests)
  const quotaTrend = pct(todayAgg.quota, prevAgg.quota)
  const tokenTrend = pct(todayAgg.tokens, prevAgg.tokens)

  const byModel = useMemo(() => {
    const mm: Record<string, { model: string; requests: number; quota: number; tokens: number }> = {}
    for (const r of rows) {
      const key = r.model_name
      if (!mm[key]) mm[key] = { model: key, requests: 0, quota: 0, tokens: 0 }
      mm[key].requests += r.request_count || 0
      mm[key].quota += r.quota || 0
      mm[key].tokens += (r.prompt_tokens || 0) + (r.completion_tokens || 0)
    }
    return Object.values(mm)
  }, [rows])

  const efficiency = useMemo(() => {
    const quotaPerUnit = getQuotaPerUnit()
    return byModel
      .map(m => {
        const quotaInCurrency = m.quota / quotaPerUnit
        return {
          model: m.model,
          requests: m.requests,
          avgCost: m.requests ? quotaInCurrency / m.requests : 0,
          avgTokens: m.requests ? m.tokens / m.requests : 0,
          efficiency: quotaInCurrency > 0 ? m.tokens / quotaInCurrency : 0, // tokens per dollar
        }
      })
      .sort((a, b) => b.requests - a.requests)
  }, [byModel])

  const topModel = efficiency.length ? efficiency[0].model : ''
  const totalModels = efficiency.length

  const usagePatterns = useMemo(() => {
    const daily = dailyAgg
    let peakDay = ''
    let peakRequests = 0
    for (const d of daily) {
      if (d.requests > peakRequests) { peakRequests = d.requests; peakDay = d.date }
    }
    const avgDaily = daily.length ? Math.round(daily.reduce((s, d) => s + d.requests, 0) / daily.length) : 0
    const trend = todayAgg.requests > avgDaily ? 'Rising' : 'Declining'
    return { peakDay, avgDaily, trend }
  }, [dailyAgg, todayAgg.requests])

  // Performance metrics calculations (similar to default dashboard)
  const performanceMetrics = useMemo(() => {
    const avgTokensPerRequest = todayAgg.requests ? todayAgg.tokens / todayAgg.requests : 0
    const avgCostPerRequest = todayAgg.requests ? todayAgg.quota / todayAgg.requests : 0 // quota is already in USD

    // Simulate avgResponseTime based on token count (in real implementation, this would come from backend)
    const avgResponseTime = todayAgg.requests > 0 ?
      Math.min(2000, Math.max(200, avgTokensPerRequest * 10)) : 0

    // Simulate success rate based on cost (in real implementation, this would come from backend)
    const successRate = todayAgg.requests > 0 ?
      Math.max(85, Math.min(99.5, 100 - (avgCostPerRequest * 1000))) : 0

    // Calculate throughput based on date range
    const dateRangeLength = Math.max(1, Math.ceil((new Date(toDate).getTime() - new Date(fromDate).getTime()) / (1000 * 60 * 60 * 24)) + 1)
    const throughput = dateRangeLength > 0 ? todayAgg.requests / (dateRangeLength * 24) : 0

    return {
      avgResponseTime,
      successRate,
      throughput,
      avgTokensPerRequest,
      avgCostPerRequest
    }
  }, [todayAgg, fromDate, toDate])

  // Enhanced cost optimization recommendations
  const costOptimizationInsights = useMemo(() => {
    if (!efficiency.length) return []

    const insights: Array<{
      type: 'warning' | 'success' | 'info'
      title: string
      message: string
      icon: string
    }> = []

    // Find most expensive model
    const mostExpensive = efficiency.reduce((max, model) =>
      model.avgCost > max.avgCost ? model : max
    )

    // Find most efficient model
    const mostEfficient = efficiency.reduce((max, model) =>
      model.efficiency > max.efficiency ? model : max
    )

    // High cost model warning
    if (mostExpensive.avgCost > 0.01) {
      insights.push({
        type: 'warning',
        title: 'High Cost Model Detected',
        message: `${mostExpensive.model} has high cost per request ($${mostExpensive.avgCost.toFixed(4)}). Consider optimizing prompts or switching models.`,
        icon: '⚠️'
      })
    }

    // Most efficient model recommendation
    if (mostEfficient.efficiency > 0) {
      insights.push({
        type: 'success',
        title: 'Most Efficient Model',
        message: `${mostEfficient.model} offers the best token-to-cost ratio with ${formatNumber(mostEfficient.efficiency)} tokens per quota unit. Consider using it for similar tasks.`,
        icon: '👍'
      })
    }

    // Monthly spending projection
    const quotaPerUnit = getQuotaPerUnit()
    const avgDailyQuota = dailyAgg.length ? (dailyAgg.reduce((s, d) => s + d.quota, 0) / dailyAgg.length) : 0
    const avgDailySpending = avgDailyQuota / quotaPerUnit
    const monthlyProjection = avgDailySpending * 30

    if (monthlyProjection > 100) {
      insights.push({
        type: 'info',
        title: 'Monthly Spending Projection',
        message: `Based on recent usage, monthly spending could reach $${monthlyProjection.toFixed(2)}. Consider setting usage limits or optimizing model selection.`,
        icon: '📊'
      })
    }

    // Usage pattern insights
    if (efficiency.length > 3) {
      const topThreeUsage = efficiency.slice(0, 3).reduce((sum, m) => sum + m.requests, 0)
      const totalUsage = efficiency.reduce((sum, m) => sum + m.requests, 0)
      const concentration = (topThreeUsage / totalUsage) * 100

      if (concentration > 80) {
        insights.push({
          type: 'info',
          title: 'Model Usage Concentration',
          message: `${concentration.toFixed(0)}% of requests use only 3 models. Consider evaluating if other models might be more cost-effective for specific tasks.`,
          icon: '🎯'
        })
      }
    }

    return insights
  }, [efficiency, dailyAgg])

  if (!user) {
    return <div>Please log in to access the dashboard.</div>
  }

  // --- UI ---
  return (
    <ResponsivePageContainer
      title="Dashboard"
      description="Monitor your API usage and account statistics"
    >
      {/* Filter bar - Date Range Controls */}
      <div className="bg-white dark:bg-gray-900 rounded-lg border p-4 mb-6">
        <div className="flex flex-col lg:flex-row gap-4 items-start lg:items-end w-full">
          <div className="flex flex-col sm:flex-row gap-3 flex-1 w-full">
            <div className="flex-1 min-w-0">
              <label className="text-sm font-medium mb-2 block">From</label>
              <Input
                type="date"
                value={fromDate}
                min={getMinDate()}
                max={getMaxDate()}
                onChange={(e) => setFromDate(e.target.value)}
                className={cn("h-10", dateError ? "border-red-500" : "")}
                aria-label="From date"
              />
            </div>
            <div className="flex-1 min-w-0">
              <label className="text-sm font-medium mb-2 block">To</label>
              <Input
                type="date"
                value={toDate}
                min={getMinDate()}
                max={getMaxDate()}
                onChange={(e) => setToDate(e.target.value)}
                className={cn("h-10", dateError ? "border-red-500" : "")}
                aria-label="To date"
              />
            </div>
            {isAdmin && (
              <div className="flex-1 min-w-0">
                <label className="text-sm font-medium mb-2 block">User</label>
                <select
                  className="h-11 sm:h-10 w-full border rounded-md px-3 py-2 text-base sm:text-sm bg-background"
                  value={dashUser}
                  onChange={(e) => setDashUser(e.target.value)}
                  aria-label="Select user"
                >
                  <option value="all">All Users</option>
                  {userOptions.map(u => (
                    <option key={u.id} value={String(u.id)}>{u.display_name || u.username}</option>
                  ))}
                </select>
              </div>
            )}
          </div>

          <div className="flex flex-wrap sm:flex-nowrap gap-2 w-full sm:w-auto sm:justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={() => applyPreset('today')}
              className="h-10 flex-1 min-w-[6rem] sm:flex-none"
            >
              Today
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => applyPreset('7d')}
              className="h-10 flex-1 min-w-[6rem] sm:flex-none"
            >
              7D
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => applyPreset('30d')}
              className="h-10 flex-1 min-w-[6rem] sm:flex-none"
            >
              30D
            </Button>
            <Button
              onClick={loadStats}
              disabled={loading}
              className="h-10 flex-1 min-w-[6rem] sm:flex-none sm:px-6"
            >
              {loading ? 'Loading...' : 'Apply'}
            </Button>
          </div>
        </div>
      </div>

      {/* Error Message */}
      {dateError && (
        <div
          id="date-error"
          className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md dark:bg-red-950/20 dark:border-red-800"
          role="alert"
          aria-live="polite"
        >
          <div className="flex items-center gap-2">
            <svg className="w-4 h-4 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm font-medium text-red-800 dark:text-red-200">Error</span>
          </div>
          <p className="text-sm text-red-700 dark:text-red-300 mt-1">{dateError}</p>
        </div>
      )}

      {/* Quota/Status panels below filter bar */}
      <div className="mb-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="border-0 shadow-sm bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Quota</p>
                  <p className="text-2xl font-bold">{renderQuota(quotaStats.totalQuota)}</p>
                </div>
                <div className="w-10 h-10 bg-blue-100 dark:bg-blue-900/30 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
                  </svg>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm bg-gradient-to-br from-green-50 to-emerald-50 dark:from-green-950/20 dark:to-emerald-950/20">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Used Quota</p>
                  <p className="text-2xl font-bold">{renderQuota(quotaStats.usedQuota)}</p>
                </div>
                <div className="w-10 h-10 bg-green-100 dark:bg-green-900/30 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                  </svg>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="border-0 shadow-sm bg-gradient-to-br from-purple-50 to-violet-50 dark:from-purple-950/20 dark:to-violet-950/20">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Status</p>
                  <p className="text-2xl font-bold">{quotaStats.status}</p>
                </div>
                <div className="w-10 h-10 bg-purple-100 dark:bg-purple-900/30 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Usage Overview - Streamlined */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-xl font-semibold">Usage Overview</h2>
            <p className="text-sm text-muted-foreground">Monitor your API usage and performance</p>
          </div>
          {lastUpdated && (
            <span className="text-xs text-muted-foreground">
              Updated: {lastUpdated}
            </span>
          )}
        </div>


        {/* Key Metrics - Clean Grid */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <div className="text-sm text-muted-foreground">Requests</div>
            <div className="text-2xl font-bold mt-1">{formatNumber(todayAgg.requests)}</div>
            <div className={cn(
              "text-xs mt-2 flex items-center gap-1",
              requestTrend >= 0 ? 'text-green-600' : 'text-red-600'
            )}>
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d={requestTrend >= 0 ? "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" : "M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"} />
              </svg>
              {Math.abs(requestTrend).toFixed(1)}%
            </div>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <div className="text-sm text-muted-foreground">Quota Used</div>
            <div className="text-2xl font-bold mt-1">{renderQuota(todayAgg.quota)}</div>
            <div className={cn(
              "text-xs mt-2 flex items-center gap-1",
              quotaTrend >= 0 ? 'text-orange-600' : 'text-green-600'
            )}>
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d={quotaTrend >= 0 ? "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" : "M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"} />
              </svg>
              {Math.abs(quotaTrend).toFixed(1)}%
            </div>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <div className="text-sm text-muted-foreground">Tokens</div>
            <div className="text-2xl font-bold mt-1">{formatNumber(todayAgg.tokens)}</div>
            <div className={cn(
              "text-xs mt-2 flex items-center gap-1",
              tokenTrend >= 0 ? 'text-blue-600' : 'text-gray-600'
            )}>
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d={tokenTrend >= 0 ? "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" : "M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"} />
              </svg>
              {Math.abs(tokenTrend).toFixed(1)}%
            </div>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <div className="text-sm text-muted-foreground">Avg Cost</div>
            <div className="text-2xl font-bold mt-1">
              ${performanceMetrics.avgCostPerRequest ? performanceMetrics.avgCostPerRequest.toFixed(4) : '0.0000'}
            </div>
            <div className="text-xs text-muted-foreground mt-2">
              {Math.round(performanceMetrics.avgTokensPerRequest)} tokens/req
            </div>
          </div>
        </div>

        {/* Quick Insights */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <h3 className="font-medium mb-3">Top Model</h3>
            <div className="text-lg font-semibold text-blue-600">{topModel || 'No data'}</div>
            <div className="text-sm text-muted-foreground mt-1">{totalModels} models active</div>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <h3 className="font-medium mb-3">Performance</h3>
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Response Time</span>
                <span className="text-sm font-medium">{performanceMetrics.avgResponseTime.toFixed(0)}ms</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Success Rate</span>
                <span className="text-sm font-medium">{performanceMetrics.successRate.toFixed(1)}%</span>
              </div>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <h3 className="font-medium mb-3">Usage Trend</h3>
            <div className="text-lg font-semibold">{usagePatterns.trend}</div>
            <div className="text-sm text-muted-foreground mt-1">
              Daily avg: {formatNumber(usagePatterns.avgDaily)} requests
            </div>
          </div>
        </div>

        {/* Usage Trends - Simplified */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <h3 className="font-medium mb-4 text-blue-600">Requests</h3>
            <ResponsiveContainer width="100%" height={140}>
              <LineChart data={timeSeries}>
                <GradientDefs />
                <CartesianGrid strokeOpacity={0.1} vertical={false} />
                <XAxis dataKey="date" hide />
                <YAxis hide />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--background)',
                    border: '1px solid var(--border)',
                    borderRadius: '8px',
                    fontSize: '12px'
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="requests"
                  stroke={chartConfig.colors.requests}
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 4, fill: chartConfig.colors.requests }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <h3 className="font-medium mb-4 text-cyan-600">Quota</h3>
            <ResponsiveContainer width="100%" height={140}>
              <LineChart data={timeSeries}>
                <GradientDefs />
                <CartesianGrid strokeOpacity={0.1} vertical={false} />
                <XAxis dataKey="date" hide />
                <YAxis hide />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--background)',
                    border: '1px solid var(--border)',
                    borderRadius: '8px',
                    fontSize: '12px'
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="quota"
                  stroke={chartConfig.colors.quota}
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 4, fill: chartConfig.colors.quota }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border p-4">
            <h3 className="font-medium mb-4 text-pink-600">Tokens</h3>
            <ResponsiveContainer width="100%" height={140}>
              <LineChart data={timeSeries}>
                <GradientDefs />
                <CartesianGrid strokeOpacity={0.1} vertical={false} />
                <XAxis dataKey="date" hide />
                <YAxis hide />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--background)',
                    border: '1px solid var(--border)',
                    borderRadius: '8px',
                    fontSize: '12px'
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="tokens"
                  stroke={chartConfig.colors.tokens}
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 4, fill: chartConfig.colors.tokens }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Model Usage Statistics */}
        <div className="bg-white dark:bg-gray-900 rounded-lg border p-6 mb-8">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold">Model Usage</h3>
            <Select
              value={statisticsMetric}
              onValueChange={(value) => setStatisticsMetric(value as 'tokens' | 'requests' | 'expenses')}
            >
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="tokens">Tokens</SelectItem>
                <SelectItem value="requests">Requests</SelectItem>
                <SelectItem value="expenses">Expenses</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={stackedData}>
              <CartesianGrid strokeOpacity={0.1} vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                fontSize={12}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                width={60}
                fontSize={12}
                tickFormatter={(value) => {
                  switch (statisticsMetric) {
                    case 'requests':
                      return formatNumber(value)
                    case 'expenses':
                      return renderQuota(value, 2)
                    case 'tokens':
                    default:
                      return formatNumber(value)
                  }
                }}
              />
              <Tooltip
                content={({ active, payload, label }) => {
                  if (active && payload && payload.length) {
                    // Filter out entries with zero values and sort by value in descending order
                    const filteredAndSortedPayload = payload
                      .filter(entry => entry.value && typeof entry.value === 'number' && entry.value > 0)
                      .sort((a, b) => (b.value as number) - (a.value as number))

                    // Format value based on selected metric
                    const formatValue = (value: number) => {
                      switch (statisticsMetric) {
                        case 'requests':
                          return formatNumber(value)
                        case 'expenses':
                          return renderQuota(value, 6)
                        case 'tokens':
                        default:
                          return formatNumber(value)
                      }
                    }

                    // Determine theme (light/dark) to set an opaque background for readability
                    const isDark = typeof document !== 'undefined' && document.documentElement.classList.contains('dark')
                    const tooltipBg = isDark ? 'rgba(17,24,39,1)' : 'rgba(255,255,255,1)' // gray-900 or white
                    const tooltipText = isDark ? 'rgba(255,255,255,0.95)' : 'rgba(17,24,39,0.9)'

                    return (
                      <div style={{
                        backgroundColor: tooltipBg,
                        border: '1px solid var(--border)',
                        borderRadius: '8px',
                        padding: '12px 16px',
                        fontSize: '12px',
                        color: tooltipText,
                        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)'
                      }}>
                        <div style={{ fontWeight: '600', marginBottom: '8px', color: 'var(--foreground)' }}>
                          {label}
                        </div>
                        {filteredAndSortedPayload.map((entry, index) => (
                          <div key={index} style={{ marginBottom: '4px', display: 'flex', alignItems: 'center' }}>
                            <span style={{
                              display: 'inline-block',
                              width: '12px',
                              height: '12px',
                              backgroundColor: entry.color,
                              borderRadius: '50%',
                              marginRight: '8px'
                            }}></span>
                            <span style={{ fontWeight: '600', color: 'var(--foreground)' }}>
                              {entry.name}: {formatValue(entry.value as number)}
                            </span>
                          </div>
                        ))}
                      </div>
                    )
                  }
                  return null
                }}
              />
              <Legend />
              {uniqueModels.map((m, idx) => (
                <Bar key={m} dataKey={m} stackId="statistics" fill={barColor(idx)} radius={[2, 2, 0, 0]} />
              ))}
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Model Efficiency - Clean Table */}
        <div className="bg-white dark:bg-gray-900 rounded-lg border p-6 mb-8">
          <h3 className="text-lg font-semibold mb-6">Model Efficiency</h3>
          <div className="space-y-4">
            {efficiency.slice(0, 5).map((m, i) => (
              <div key={m.model} className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                <div className="flex items-center gap-4">
                  <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center text-sm font-semibold text-blue-600">
                    {i + 1}
                  </div>
                  <div>
                    <div className="font-medium">{m.model}</div>
                    <div className="text-sm text-muted-foreground">
                      {formatNumber(m.requests)} requests • ${m.avgCost.toFixed(4)} avg cost
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <div className="w-20 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-blue-500 to-green-500 rounded-full"
                      style={{
                        width: Math.min(100, Math.round(m.efficiency / (efficiency[0]?.efficiency || 1) * 100)) + '%'
                      }}
                    />
                  </div>
                  <span className="text-sm font-medium min-w-[3rem] text-right">
                    {m.efficiency ? m.efficiency.toFixed(0) : 0}
                  </span>
                </div>
              </div>
            ))}
            {efficiency.length === 0 && (
              <div className="py-8 text-center text-muted-foreground">
                No efficiency data available
              </div>
            )}
          </div>
        </div>

        {/* Cost Optimization Recommendations */}
        {costOptimizationInsights.length > 0 && (
          <div className="bg-white dark:bg-gray-900 rounded-lg border p-6 mb-8">
            <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
              <svg className="w-5 h-5 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
              Optimization Tips
            </h3>
            <div className="space-y-4">
              {costOptimizationInsights.map((insight, index) => (
                <div
                  key={index}
                  className={cn(
                    "flex items-start gap-3 p-4 rounded-lg border-l-4",
                    insight.type === 'warning' && "bg-orange-50 border-l-orange-500 dark:bg-orange-950/20",
                    insight.type === 'success' && "bg-green-50 border-l-green-500 dark:bg-green-950/20",
                    insight.type === 'info' && "bg-blue-50 border-l-blue-500 dark:bg-blue-950/20"
                  )}
                >
                  <div className="text-xl mt-0.5">
                    {insight.icon}
                  </div>
                  <div className="flex-1">
                    <div className="font-medium mb-1">
                      {insight.title}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {insight.message}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Data Table */}
        <div className="bg-white dark:bg-gray-900 rounded-lg border p-6">
          <h3 className="text-lg font-semibold mb-4">Detailed Data</h3>
          <DataTable
            columns={columns}
            data={rows}
            pageIndex={0}
            pageSize={rows.length || 10}
            total={rows.length}
            onPageChange={() => { }}
          />
        </div>
      </div>
    </ResponsivePageContainer>
  )
}

export function barColor(i: number) {
  return chartConfig.barColors[i % chartConfig.barColors.length]
}
