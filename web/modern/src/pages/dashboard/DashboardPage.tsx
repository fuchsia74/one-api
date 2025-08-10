import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { useAuthStore } from '@/lib/stores/auth'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { api } from '@/lib/api'
import { DataTable } from '@/components/ui/data-table'
import { formatNumber } from '@/lib/utils'
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

export function DashboardPage() {
  const { user } = useAuthStore()
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

  type Row = { day: string; model_name: string; request_count: number; quota: number; prompt_tokens: number; completion_tokens: number }
  const [rows, setRows] = useState<Row[]>([])

  const totals = useMemo(() => {
    return rows.reduce(
      (acc, r) => {
        acc.requests += r.request_count || 0
        acc.quota += r.quota || 0
        acc.tokens += (r.prompt_tokens || 0) + (r.completion_tokens || 0)
        return acc
      },
      { requests: 0, quota: 0, tokens: 0 }
    )
  }, [rows])

  const loadUsers = async () => {
    if (!isAdmin) return
    const res = await api.get('/user/dashboard/users')
    if (res.data?.success) {
      setUserOptions(res.data.data || [])
    }
  }

  const loadStats = async () => {
    setLoading(true)
    try {
      const params = new URLSearchParams()
      params.set('from_date', fromDate)
      params.set('to_date', toDate)
      if (isAdmin) {
        params.set('user_id', dashUser || 'all')
      }
      const res = await api.get('/user/dashboard?' + params.toString())
      const { success, data, message } = res.data
      if (success) {
        setRows(data || [])
        setLastUpdated(new Date().toLocaleTimeString())
      } else {
        console.warn('dashboard fetch failed', message)
        setRows([])
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (isAdmin) loadUsers()
    // load initial stats
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAdmin])
  useEffect(() => {
    loadStats()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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
    const map: Record<string, { date: string; requests: number; quota: number; tokens: number }> = {}
    for (const r of rows) {
      if (!map[r.day]) map[r.day] = { date: r.day, requests: 0, quota: 0, tokens: 0 }
      map[r.day].requests += r.request_count || 0
      map[r.day].quota += r.quota || 0
      map[r.day].tokens += (r.prompt_tokens || 0) + (r.completion_tokens || 0)
    }
    return Object.values(map).sort((a, b) => a.date.localeCompare(b.date))
  }, [rows])

  // Model stacked bar per day
  const uniqueModels = useMemo(() => Array.from(new Set(rows.map(r => r.model_name))), [rows])
  const stackedByTokens = useMemo(() => {
    const map: Record<string, Record<string, number>> = {}
    for (const d of days) map[d] = {}
    for (const r of rows) {
      const totalTokens = (r.prompt_tokens || 0) + (r.completion_tokens || 0)
      map[r.day][r.model_name] = (map[r.day][r.model_name] || 0) + totalTokens
    }
    return days.map(d => ({ date: d, ...(map[d] || {}) }))
  }, [rows, days])

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
    const m: Record<string, { date: string; requests: number; quota: number; tokens: number }> = {}
    for (const r of rows) {
      if (!m[r.day]) m[r.day] = { date: r.day, requests: 0, quota: 0, tokens: 0 }
      m[r.day].requests += r.request_count || 0
      m[r.day].quota += r.quota || 0
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
  const avgCostPerRequest = todayAgg.requests ? todayAgg.quota / todayAgg.requests : 0
  const avgTokensPerRequest = todayAgg.requests ? todayAgg.tokens / todayAgg.requests : 0

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
    return byModel
      .map(m => ({
        model: m.model,
        requests: m.requests,
        avgCost: m.requests ? m.quota / m.requests : 0,
        avgTokens: m.requests ? m.tokens / m.requests : 0,
        efficiency: m.quota > 0 ? m.tokens / m.quota : 0, // tokens per quota unit
      }))
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

  const recommendations = useMemo(() => {
    if (!efficiency.length) return [] as string[]
    const sortedByCost = [...efficiency].sort((a, b) => b.avgCost - a.avgCost)
    const highCost = sortedByCost[0]
    const mostEff = [...efficiency].sort((a, b) => b.efficiency - a.efficiency)[0]
    const dailyCost = todayAgg.quota // using current day as a simple proxy
    const monthlyProjection = Math.round((dailyAgg.length ? (dailyAgg.reduce((s, d) => s + d.quota, 0) / dailyAgg.length) : 0) * 30)
    return [
      highCost ? `High cost model detected: ${highCost.model} has high cost/request (${formatNumber(highCost.avgCost)}). Consider optimizing prompts or switching models.` : '',
      mostEff ? `Most efficient model: ${mostEff.model} with ${formatNumber(mostEff.efficiency)} tokens per quota.` : '',
      `Monthly projection based on recent daily average: ${formatNumber(monthlyProjection)} quota.`,
    ].filter(Boolean)
  }, [efficiency, todayAgg.quota, dailyAgg])

  if (!user) {
    return <div>Please log in to access the dashboard.</div>
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-6">Dashboard</h1>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Account Info</CardTitle>
              <CardDescription>Your account details</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <p><strong>Username:</strong> {user.username}</p>
                <p><strong>Display Name:</strong> {user.display_name || 'Not set'}</p>
                <p><strong>Role:</strong> {isAdmin ? 'Admin' : 'User'}</p>
                <p><strong>Group:</strong> {user.group}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Quota Usage</CardTitle>
              <CardDescription>Your API usage statistics</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <p><strong>Total Quota:</strong> {user.quota?.toLocaleString() || 'N/A'}</p>
                <p><strong>Used:</strong> {user.used_quota?.toLocaleString() || 'N/A'}</p>
                <p><strong>Remaining:</strong> {user.quota && user.used_quota ? (user.quota - user.used_quota).toLocaleString() : 'N/A'}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Status</CardTitle>
              <CardDescription>Account status</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <p><strong>Status:</strong> {user.status === 1 ? 'Active' : 'Inactive'}</p>
                <p><strong>Email:</strong> {user.email || 'Not set'}</p>
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="mt-8">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Usage Overview</CardTitle>
                  <CardDescription>Daily requests and quota by model</CardDescription>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  {lastUpdated && <span className="px-2 py-1 rounded-full bg-accent/40">Updated: {lastUpdated}</span>}
                  <Button variant="outline" onClick={loadStats} disabled={loading}>Refresh</Button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-12 gap-3 mb-4">
                <div className="md:col-span-3">
                  <label className="text-xs block mb-1">From</label>
                  <Input type="date" value={fromDate} onChange={(e)=>setFromDate(e.target.value)} />
                </div>
                <div className="md:col-span-3">
                  <label className="text-xs block mb-1">To</label>
                  <Input type="date" value={toDate} onChange={(e)=>setToDate(e.target.value)} />
                </div>
                {isAdmin && (
                  <div className="md:col-span-3">
                    <label className="text-xs block mb-1">User</label>
                    <select className="h-9 w-full border rounded-md px-2 text-sm" value={dashUser} onChange={(e)=>setDashUser(e.target.value)}>
                      <option value="all">All Users (Site-wide)</option>
                      {userOptions.map(u => (
                        <option key={u.id} value={String(u.id)}>{u.display_name || u.username}</option>
                      ))}
                    </select>
                    <div className="text-[11px] text-muted-foreground mt-1">As root, you can select up to 1 year of data.</div>
                  </div>
                )}
                <div className="md:col-span-3 flex items-end">
                  <Button size="sm" variant="outline" onClick={() => applyPreset('today')}>Today</Button>
                  <Button size="sm" variant="outline" onClick={() => applyPreset('7d')}>Last 7 Days</Button>
                  <Button size="sm" variant="outline" onClick={() => applyPreset('30d')}>Last 30 Days</Button>
                  <Button onClick={loadStats} disabled={loading}>Apply</Button>
                </div>
              </div>

              {/* Summary cards */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 mb-4">
                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground mb-1">Total Requests (Today)</div>
                  <div className="text-2xl font-semibold">{formatNumber(todayAgg.requests)}</div>
                  <div className={`text-xs mt-1 ${requestTrend>=0?'text-green-600':'text-red-600'}`}>{requestTrend>=0?'+':''}{requestTrend.toFixed(1)}%</div>
                </div>
                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground mb-1">Total Quota (Today)</div>
                  <div className="text-2xl font-semibold">{formatNumber(todayAgg.quota)}</div>
                  <div className={`text-xs mt-1 ${quotaTrend>=0?'text-green-600':'text-red-600'}`}>{quotaTrend>=0?'+':''}{quotaTrend.toFixed(1)}%</div>
                </div>
                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground mb-1">Total Tokens (Today)</div>
                  <div className="text-2xl font-semibold">{formatNumber(todayAgg.tokens)}</div>
                  <div className={`text-xs mt-1 ${tokenTrend>=0?'text-green-600':'text-red-600'}`}>{tokenTrend>=0?'+':''}{tokenTrend.toFixed(1)}%</div>
                </div>
                <div className="rounded-lg border p-3">
                  <div className="text-xs text-muted-foreground mb-1">Avg Cost / Request</div>
                  <div className="text-2xl font-semibold">{avgCostPerRequest ? avgCostPerRequest.toFixed(4) : '0'}</div>
                  <div className="text-xs mt-1">Avg Tokens: {avgTokensPerRequest ? Math.round(avgTokensPerRequest) : 0}</div>
                </div>
              </div>

              {/* Insights */}
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
                <div className="border rounded-lg p-3">
                  <div className="text-sm mb-2">Model Usage Insights</div>
                  <div className="text-xs text-muted-foreground">Most Used Model</div>
                  <div className="text-base font-medium">{topModel || '-'}</div>
                  <div className="text-xs text-muted-foreground mt-2">Active Models</div>
                  <div className="text-base font-medium">{totalModels}</div>
                </div>
                <div className="border rounded-lg p-3">
                  <div className="text-sm mb-2">Performance Metrics</div>
                  <div className="text-xs text-muted-foreground">Avg Tokens/Req</div>
                  <div className="text-base font-medium">{avgTokensPerRequest ? Math.round(avgTokensPerRequest) : 0}</div>
                  <div className="text-xs text-muted-foreground mt-2">Throughput (req/day)</div>
                  <div className="text-base font-medium">{todayAgg.requests}</div>
                </div>
                <div className="border rounded-lg p-3">
                  <div className="text-sm mb-2">Usage Patterns</div>
                  <div className="text-xs text-muted-foreground">Peak Day</div>
                  <div className="text-base font-medium">{usagePatterns.peakDay || '-'}</div>
                  <div className="text-xs text-muted-foreground mt-2">Daily Average</div>
                  <div className="text-base font-medium">{formatNumber(usagePatterns.avgDaily)}</div>
                  <div className="text-xs mt-2">Trend: <span className={usagePatterns.trend==='Rising'?'text-green-600':'text-red-600'}>{usagePatterns.trend}</span></div>
                </div>
              </div>

              {/* Trends */}
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
                <div className="border rounded-lg p-3">
                  <div className="text-sm mb-2">Model Request Trend</div>
                  <ResponsiveContainer width="100%" height={160}>
                    <LineChart data={timeSeries}>
                      <CartesianGrid strokeOpacity={0.2} vertical={false} />
                      <XAxis dataKey="date" tickLine={false} axisLine={false} />
                      <YAxis tickLine={false} axisLine={false} width={40} />
                      <Tooltip />
                      <Line type="monotone" dataKey="requests" stroke="#4318FF" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
                <div className="border rounded-lg p-3">
                  <div className="text-sm mb-2">Quota Usage Trend</div>
                  <ResponsiveContainer width="100%" height={160}>
                    <LineChart data={timeSeries}>
                      <CartesianGrid strokeOpacity={0.2} vertical={false} />
                      <XAxis dataKey="date" tickLine={false} axisLine={false} />
                      <YAxis tickLine={false} axisLine={false} width={40} />
                      <Tooltip />
                      <Line type="monotone" dataKey="quota" stroke="#00B5D8" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
                <div className="border rounded-lg p-3">
                  <div className="text-sm mb-2">Token Usage Trend</div>
                  <ResponsiveContainer width="100%" height={160}>
                    <LineChart data={timeSeries}>
                      <CartesianGrid strokeOpacity={0.2} vertical={false} />
                      <XAxis dataKey="date" tickLine={false} axisLine={false} />
                      <YAxis tickLine={false} axisLine={false} width={40} />
                      <Tooltip />
                      <Line type="monotone" dataKey="tokens" stroke="#FF5E7D" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </div>

              {/* Stacked by model (tokens) */}
              <div className="border rounded-lg p-3 mb-6">
                <div className="flex items-center justify-between mb-2">
                  <div className="text-sm">Statistics - Tokens</div>
                </div>
                <ResponsiveContainer width="100%" height={260}>
                  <BarChart data={stackedByTokens}>
                    <CartesianGrid strokeOpacity={0.2} vertical={false} />
                    <XAxis dataKey="date" tickLine={false} axisLine={false} />
                    <YAxis tickLine={false} axisLine={false} width={40} />
                    <Tooltip />
                    <Legend wrapperStyle={{ fontSize: 12 }} height={24} />
                    {uniqueModels.map((m, idx) => (
                      <Bar key={m} dataKey={m} stackId="tokens" fill={barColor(idx)} />
                    ))}
                  </BarChart>
                </ResponsiveContainer>
              </div>

              {/* Model Efficiency Analysis */}
              <div className="border rounded-lg p-3 mb-6">
                <div className="text-sm mb-3">Model Efficiency Analysis</div>
                <div className="overflow-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="text-left text-muted-foreground">
                        <th className="py-2 pr-2">#</th>
                        <th className="py-2 pr-2">Model</th>
                        <th className="py-2 pr-2">Requests</th>
                        <th className="py-2 pr-2">Avg Cost</th>
                        <th className="py-2 pr-2">Avg Tokens</th>
                        <th className="py-2 pr-2">Efficiency (tokens/quota)</th>
                      </tr>
                    </thead>
                    <tbody>
                      {efficiency.slice(0, 10).map((m, i) => (
                        <tr key={m.model} className="border-t">
                          <td className="py-2 pr-2">{i+1}</td>
                          <td className="py-2 pr-2 font-medium">{m.model}</td>
                          <td className="py-2 pr-2">{formatNumber(m.requests)}</td>
                          <td className="py-2 pr-2">{m.avgCost.toFixed(4)}</td>
                          <td className="py-2 pr-2">{Math.round(m.avgTokens)}</td>
                          <td className="py-2 pr-2">
                            <div className="flex items-center gap-2">
                              <div className="h-2 bg-accent rounded" style={{ width: Math.min(100, Math.round(m.efficiency / (efficiency[0]?.efficiency || 1) * 100)) + '%' }} />
                              <span>{m.efficiency ? m.efficiency.toFixed(0) : 0}</span>
                            </div>
                          </td>
                        </tr>
                      ))}
                      {efficiency.length === 0 && (
                        <tr><td className="py-3 text-muted-foreground" colSpan={6}>No data</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Cost Optimization Recommendations */}
              <div className="border rounded-lg p-3">
                <div className="text-sm mb-2">Cost Optimization Recommendations</div>
                <ul className="list-disc pl-5 text-sm space-y-1">
                  {recommendations.map((r, idx) => (<li key={idx}>{r}</li>))}
                </ul>
              </div>

              <DataTable columns={columns} data={rows} pageIndex={0} pageSize={rows.length || 10} total={rows.length} onPageChange={() => {}} />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

export function barColor(i: number) {
  const palette = [
    '#4318FF', '#00B5D8', '#6C63FF', '#05CD99', '#FFB547', '#FF5E7D', '#41B883', '#7983FF', '#FF8F6B', '#49BEFF',
    '#8B5CF6', '#F59E0B', '#EF4444', '#10B981', '#3B82F6',
  ]
  return palette[i % palette.length]
}
