import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { EnhancedDataTable } from '@/components/ui/enhanced-data-table'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import { ResponsivePageContainer } from '@/components/ui/responsive-container'
import { useResponsive } from '@/hooks/useResponsive'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatTimestamp, cn } from '@/lib/utils'
import { Plus, TestTube, RefreshCw, Trash2, Settings, AlertCircle } from 'lucide-react'
import { useNotifications } from '@/components/ui/notifications'
import { ResponsiveActionGroup } from '@/components/ui/responsive-action-group'

interface Channel {
  id: number
  name: string
  type: number
  status: number
  response_time?: number
  created_time: number
  updated_time?: number
  priority?: number
  weight?: number
  models?: string
  group?: string
  used_quota?: number
  test_time?: number
  testing_model?: string | null
}

/**
 * Channel options defined at relay/channeltype/define.go
 */
const CHANNEL_TYPES: Record<number, { name: string; color: string }> = {
  1: { name: 'OpenAI', color: 'green' },
  50: { name: 'OpenAI Compatible', color: 'olive' },
  14: { name: 'Anthropic', color: 'black' },
  33: { name: 'AWS', color: 'orange' },
  3: { name: 'Azure', color: 'blue' },
  11: { name: 'PaLM2', color: 'orange' },
  24: { name: 'Gemini', color: 'orange' },
  51: { name: 'Gemini (OpenAI)', color: 'orange' },
  28: { name: 'Mistral AI', color: 'purple' },
  41: { name: 'Novita', color: 'purple' },
  40: { name: 'ByteDance Volcano', color: 'blue' },
  15: { name: 'Baidu Wenxin', color: 'blue' },
  47: { name: 'Baidu Wenxin V2', color: 'blue' },
  17: { name: 'Alibaba Qianwen', color: 'orange' },
  49: { name: 'Alibaba Bailian', color: 'orange' },
  18: { name: 'iFlytek Spark', color: 'blue' },
  48: { name: 'iFlytek Spark V2', color: 'blue' },
  16: { name: 'Zhipu ChatGLM', color: 'violet' },
  19: { name: '360 ZhiNao', color: 'blue' },
  25: { name: 'Moonshot AI', color: 'black' },
  23: { name: 'Tencent Hunyuan', color: 'teal' },
  26: { name: 'Baichuan', color: 'orange' },
  27: { name: 'MiniMax', color: 'red' },
  29: { name: 'Groq', color: 'orange' },
  30: { name: 'Ollama', color: 'black' },
  31: { name: '01.AI', color: 'green' },
  32: { name: 'StepFun', color: 'blue' },
  34: { name: 'Coze', color: 'blue' },
  35: { name: 'Cohere', color: 'blue' },
  36: { name: 'DeepSeek', color: 'black' },
  37: { name: 'Cloudflare', color: 'orange' },
  38: { name: 'DeepL', color: 'black' },
  39: { name: 'together.ai', color: 'blue' },
  42: { name: 'VertexAI', color: 'blue' },
  43: { name: 'Proxy', color: 'blue' },
  44: { name: 'SiliconFlow', color: 'blue' },
  45: { name: 'xAI', color: 'blue' },
  46: { name: 'Replicate', color: 'blue' },
  8: { name: 'Custom', color: 'pink' },
  22: { name: 'FastGPT', color: 'blue' },
  21: { name: 'AI Proxy KB', color: 'purple' },
  20: { name: 'OpenRouter', color: 'black' },
}

const getChannelTypeBadge = (type: number) => {
  const channelType = CHANNEL_TYPES[type] || { name: `Type ${type}`, color: 'gray' }
  return (
    <Badge variant="outline" className={`text-xs`}>
      {channelType.name}
    </Badge>
  )
}

const getStatusBadge = (status: number, priority?: number) => {
  if (status === 2) {
    return <Badge variant="destructive">Disabled</Badge>
  }
  if ((priority ?? 0) < 0) {
    return <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">Paused</Badge>
  }
  return <Badge variant="default" className="bg-green-100 text-green-800">Active</Badge>
}

const formatResponseTime = (time?: number) => {
  if (!time) return '-'
  const color = time < 1000 ? 'text-green-600' : time < 3000 ? 'text-yellow-600' : 'text-red-600'
  return <span className={cn('font-mono text-sm', color)}>{time}ms</span>
}

export function ChannelsPage() {
  const navigate = useNavigate()
  const { isMobile } = useResponsive()
  const { notify } = useNotifications()
  const [data, setData] = useState<Channel[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchOptions, setSearchOptions] = useState<SearchOption[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [sortBy, setSortBy] = useState('id')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [bulkTesting, setBulkTesting] = useState(false)
  const initializedRef = useRef(false)
  const skipFirstSortEffect = useRef(true)

  const load = async (p = 0, size = pageSize) => {
    setLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      let url = `/api/channel/?p=${p}&size=${size}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`

      const res = await api.get(url)
      const { success, data: responseData, total: responseTotal } = res.data

      if (success) {
        setData(responseData || [])
        setTotal(responseTotal || 0)
        setPageIndex(p)
        setPageSize(size)
      }
    } catch (error) {
      console.error('Failed to load channels:', error)
      setData([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  const searchChannels = async (query: string) => {
    if (!query.trim()) {
      setSearchOptions([])
      return
    }

    setSearchLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      let url = `/api/channel/search?keyword=${encodeURIComponent(query)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      url += `&size=${pageSize}`

      const res = await api.get(url)
      const { success, data: responseData } = res.data

      if (success && Array.isArray(responseData)) {
        const options: SearchOption[] = responseData.map((channel: Channel) => ({
          key: channel.id.toString(),
          value: channel.name,
          text: channel.name,
          content: (
            <div className="flex flex-col">
              <div className="font-medium">{channel.name}</div>
              <div className="text-sm text-muted-foreground flex items-center gap-2">
                ID: {channel.id} • {getChannelTypeBadge(channel.type)} • {getStatusBadge(channel.status, channel.priority)}
              </div>
            </div>
          )
        }))
        setSearchOptions(options)
      }
    } catch (error) {
      console.error('Search failed:', error)
      setSearchOptions([])
    } finally {
      setSearchLoading(false)
    }
  }

  const performSearch = async () => {
    if (!searchKeyword.trim()) {
      return load(0, pageSize)
    }

    setLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      let url = `/api/channel/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      url += `&size=${pageSize}`

      const res = await api.get(url)
      const { success, data: responseData } = res.data

      if (success) {
        setData(responseData || [])
        setPageIndex(0)
        setTotal(responseData?.length || 0)
      }
    } catch (error) {
      console.error('Search failed:', error)
    } finally {
      setLoading(false)
    }
  }

  // Load initial data
  useEffect(() => {
    load(0, pageSize)
    initializedRef.current = true
  }, [])

  // Handle sort changes (only after initialization)
  useEffect(() => {
    // Skip the very first run to avoid duplicating the initial load
    if (skipFirstSortEffect.current) {
      skipFirstSortEffect.current = false
      return
    }

    if (!initializedRef.current) return

    if (searchKeyword.trim()) {
      performSearch()
    } else {
      load(pageIndex, pageSize)
    }
  }, [sortBy, sortOrder])

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete' | 'test', index?: number) => {
    try {
      if (action === 'delete') {
        if (!confirm('Are you sure you want to delete this channel?')) return
        // Unified API call - complete URL with /api prefix
        const res = await api.delete(`/api/channel/${id}`)
        if (res.data?.success) {
          if (searchKeyword.trim()) {
            performSearch()
          } else {
            load(pageIndex, pageSize)
          }
        }
        return
      }

      if (action === 'test') {
        // Unified API call - complete URL with /api prefix
        const res = await api.get(`/api/channel/test/${id}`)
        const { success, time, message } = res.data
        if (index !== undefined) {
          const newData = [...data]
          newData[index] = { ...newData[index], response_time: time, test_time: Date.now() }
          setData(newData)
        }
        if (success) {
          notify({ type: 'success', message: 'Channel test successful.' })
        } else {
          notify({ type: 'error', title: 'Channel test failed', message: message || 'Unknown error' })
        }
        return
      }

      // Enable/disable - send status_only to avoid overwriting other fields
      const payload = { id, status: action === 'enable' ? 1 : 2 }
      const res = await api.put('/api/channel/?status_only=1', payload)
      if (res.data?.success) {
        if (searchKeyword.trim()) {
          performSearch()
        } else {
          load(pageIndex, pageSize)
        }
      }
    } catch (error) {
      console.error(`Failed to ${action} channel:`, error)
    }
  }

  const updateTestingModel = async (id: number, testingModel: string | null) => {
    try {
      const current = data.find((c) => c.id === id)
      const payload: any = { id, name: current?.name }
      // When null, let backend clear it (auto-cheapest)
      if (testingModel === null) {
        payload.testing_model = null
      } else {
        payload.testing_model = testingModel
      }
      // Unified API call - complete URL with /api prefix
      const res = await api.put('/api/channel/', payload)
      if (res.data?.success) {
        // Update local row to reflect change
        setData((prev) => prev.map((ch) => (ch.id === id ? { ...ch, testing_model: testingModel } : ch)))
        notify({ type: 'success', message: 'Testing model saved.' })
      } else {
        const msg = res.data?.message || 'Failed to save testing model'
        notify({ type: 'error', title: 'Save failed', message: msg })
      }
    } catch (error) {
      console.error('Failed to update testing model:', error)
      notify({ type: 'error', title: 'Save failed', message: 'Failed to update testing model' })
    }
  }

  const handleBulkTest = async () => {
    setBulkTesting(true)
    try {
      // Unified API call - complete URL with /api prefix
      await api.get('/api/channel/test')
      load(pageIndex, pageSize)
      notify({ type: 'info', message: 'Bulk channel test started.' })
    } catch (error) {
      console.error('Bulk test failed:', error)
      notify({ type: 'error', title: 'Bulk test failed', message: error instanceof Error ? error.message : 'Unknown error' })
    } finally {
      setBulkTesting(false)
    }
  }

  const handleDeleteDisabled = async () => {
    if (!confirm('Are you sure you want to delete all disabled channels? This action cannot be undone.')) return

    try {
      // Unified API call - complete URL with /api prefix
      await api.delete('/api/channel/disabled')
      load(pageIndex, pageSize)
      notify({ type: 'success', message: 'Disabled channels deleted.' })
    } catch (error) {
      console.error('Failed to delete disabled channels:', error)
      notify({ type: 'error', title: 'Delete failed', message: error instanceof Error ? error.message : 'Unknown error' })
    }
  }

  const columns: ColumnDef<Channel>[] = [
    {
      accessorKey: 'id',
      header: 'ID',
      cell: ({ row }) => (
        <span className="font-mono text-sm">{row.original.id}</span>
      ),
    },
    {
      accessorKey: 'name',
      header: 'Name',
      cell: ({ row }) => (
        <div className="font-medium">{row.original.name}</div>
      ),
    },
    {
      accessorKey: 'type',
      header: 'Type',
      cell: ({ row }) => getChannelTypeBadge(row.original.type),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => getStatusBadge(row.original.status, row.original.priority),
    },
    {
      accessorKey: 'group',
      header: 'Group',
      cell: ({ row }) => (
        <span className="text-sm">{row.original.group || 'default'}</span>
      ),
    },
    {
      accessorKey: 'priority',
      header: 'Priority',
      cell: ({ row }) => (
        <span className="font-mono text-sm">{row.original.priority || 0}</span>
      ),
    },
    {
      accessorKey: 'weight',
      header: 'Weight',
      cell: ({ row }) => (
        <span className="font-mono text-sm">{row.original.weight || 0}</span>
      ),
    },
    {
      accessorKey: 'response_time',
      header: 'Response',
      cell: ({ row }) => (
        <div className="text-center" title={`Response time: ${row.original.response_time ? `${row.original.response_time}ms` : 'Not tested'}${row.original.test_time ? ` (Tested: ${formatTimestamp(row.original.test_time)})` : ''}`}>
          {formatResponseTime(row.original.response_time)}
          {row.original.test_time && (
            <div className="text-xs text-muted-foreground">
              {formatTimestamp(row.original.test_time)}
            </div>
          )}
        </div>
      ),
    },
    {
      accessorKey: 'testing_model',
      header: 'Testing Model',
      cell: ({ row }) => {
        const ch = row.original
        const models = (ch.models || '')
          .split(',')
          .map((m) => m.trim())
          .filter(Boolean)
        const value = ch.testing_model ?? '' // empty => Auto (cheapest)
        return (
          <div className="w-[140px] md:w-[160px] max-w-[220px]">
            <select
              className="w-full border rounded px-2 py-1 text-sm bg-background"
              value={value}
              onChange={(e) => {
                const v = e.target.value
                updateTestingModel(ch.id, v === '' ? null : v)
              }}
            >
              <option value="">CHEAPEST</option>
              {models.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>
        )
      },
    },
    {
      accessorKey: 'created_time',
      header: 'Created',
      cell: ({ row }) => (
        <span className="text-sm">{formatTimestamp(row.original.created_time)}</span>
      ),
    },
    {
      header: 'Actions',
      cell: ({ row }) => {
        const channel = row.original
        return (
          <ResponsiveActionGroup className="sm:items-center">
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate(`/channels/edit/${channel.id}`)}
              className="gap-1"
            >
              <Settings className="h-3 w-3" />
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => manage(channel.id, channel.status === 1 ? 'disable' : 'enable')}
              className={cn(
                'gap-1',
                channel.status === 1
                  ? 'text-orange-600 hover:text-orange-700'
                  : 'text-green-600 hover:text-green-700'
              )}
            >
              {channel.status === 1 ? 'Disable' : 'Enable'}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => manage(channel.id, 'test', row.index)}
              className="gap-1"
            >
              <TestTube className="h-3 w-3" />
              Test
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => manage(channel.id, 'delete')}
              className="gap-1"
            >
              <Trash2 className="h-3 w-3" />
              Delete
            </Button>
          </ResponsiveActionGroup>
        )
      },
    },
  ]

  const handlePageChange = (newPageIndex: number, newPageSize: number) => {
    load(newPageIndex, newPageSize)
  }

  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize)
    // Don't call load here - let onPageChange handle it to avoid duplicate API calls
    setPageIndex(0)
  }

  const handleSortChange = (newSortBy: string, newSortOrder: 'asc' | 'desc') => {
    setSortBy(newSortBy)
    setSortOrder(newSortOrder)
    // Let useEffect handle the reload to avoid double requests
  }

  const refresh = () => {
    if (searchKeyword.trim()) {
      performSearch()
    } else {
      load(pageIndex, pageSize)
    }
  }

  const toolbarActions = (
    <div className={cn(
      "flex gap-2 flex-wrap max-w-full",
      isMobile ? "flex-col w-full" : "items-center"
    )}>
      <Button
        variant="outline"
        onClick={handleBulkTest}
        disabled={bulkTesting || loading}
        className={cn(
          "gap-2",
          isMobile ? "w-full touch-target" : ""
        )}
        size="sm"
      >
        {bulkTesting ? (
          <RefreshCw className="h-4 w-4 animate-spin" />
        ) : (
          <TestTube className="h-4 w-4" />
        )}
        {isMobile ? "Test All Channels" : "Test All"}
      </Button>
      <Button
        variant="destructive"
        onClick={handleDeleteDisabled}
        className={cn(
          "gap-2",
          isMobile ? "w-full touch-target" : ""
        )}
        size="sm"
      >
        <Trash2 className="h-4 w-4" />
        {isMobile ? "Delete All Disabled" : "Delete Disabled"}
      </Button>
    </div>
  )

  return (
    <ResponsivePageContainer
      title="Channels"
      description="Configure and manage API routing channels"
      actions={
        <Button
          onClick={() => navigate('/channels/add')}
          className={cn(
            "gap-2",
            isMobile ? "w-full touch-target" : ""
          )}
        >
          <Plus className="h-4 w-4" />
          {isMobile ? "Add New Channel" : "Add Channel"}
        </Button>
      }
    >
      <Card>
        <CardContent className={cn(
          isMobile ? "p-4" : "p-6"
        )}>
          <EnhancedDataTable
            columns={columns}
            data={data}
            pageIndex={pageIndex}
            pageSize={pageSize}
            total={total}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
            sortBy={sortBy}
            sortOrder={sortOrder}
            onSortChange={handleSortChange}
            searchValue={searchKeyword}
            searchOptions={searchOptions}
            searchLoading={searchLoading}
            onSearchChange={searchChannels}
            onSearchValueChange={setSearchKeyword}
            onSearchSubmit={performSearch}
            searchPlaceholder="Search channels by name, type, or group..."
            allowSearchAdditions={true}
            toolbarActions={toolbarActions}
            onRefresh={refresh}
            loading={loading}
            emptyMessage="No channels found. Create your first channel to get started."
            mobileCardLayout={true}
            hideColumnsOnMobile={['created_time', 'response_time']}
            compactMode={isMobile}
          />
        </CardContent>
      </Card>
    </ResponsivePageContainer>
  )
}
