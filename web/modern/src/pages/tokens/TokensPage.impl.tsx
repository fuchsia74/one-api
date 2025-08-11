import { useEffect, useState } from 'react'
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
import { cn } from '@/lib/utils'
import { renderQuota } from '@/lib/utils'
import { Plus, Copy, Eye, EyeOff } from 'lucide-react'

interface Token {
  id: number
  name: string
  key: string
  status: number
  remain_quota: number
  unlimited_quota: boolean
  used_quota: number
  created_time: number
  accessed_time: number
  expired_time: number
  models?: string
  subnet?: string
}

// Status constants
const TOKEN_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
  EXPIRED: 3,
  EXHAUSTED: 4,
} as const

const formatQuota = (quota: number, unlimited = false) => {
  if (unlimited) return 'Unlimited'
  return renderQuota(quota)
}

const formatTimestamp = (timestamp: number) => {
  if (timestamp === -1) return 'Never'
  return new Date(timestamp * 1000).toLocaleString()
}

const getStatusBadge = (status: number) => {
  switch (status) {
    case TOKEN_STATUS.ENABLED:
      return <Badge variant="default" className="bg-green-100 text-green-800">Enabled</Badge>
    case TOKEN_STATUS.DISABLED:
      return <Badge variant="secondary" className="bg-gray-100 text-gray-800">Disabled</Badge>
    case TOKEN_STATUS.EXPIRED:
      return <Badge variant="destructive" className="bg-red-100 text-red-800">Expired</Badge>
    case TOKEN_STATUS.EXHAUSTED:
      return <Badge variant="destructive" className="bg-yellow-100 text-yellow-800">Exhausted</Badge>
    default:
      return <Badge variant="outline">Unknown</Badge>
  }
}

export function TokensPage() {
  const navigate = useNavigate()
  const { isMobile } = useResponsive()
  const [data, setData] = useState<Token[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchOptions, setSearchOptions] = useState<SearchOption[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [sortBy, setSortBy] = useState('id')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [showKeys, setShowKeys] = useState<Record<number, boolean>>({})

  const load = async (p = 0, size = pageSize) => {
    setLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      let url = `/api/token/?p=${p}&size=${size}`
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
      console.error('Failed to load tokens:', error)
      setData([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load(0, pageSize)
  }, [pageSize])

  useEffect(() => {
    if (searchKeyword.trim()) {
      performSearch()
    } else {
      load(0, pageSize)
    }
  }, [sortBy, sortOrder])

  const searchTokens = async (query: string) => {
    if (!query.trim()) {
      setSearchOptions([])
      return
    }

    setSearchLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      let url = `/api/token/search?keyword=${encodeURIComponent(query)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`

      const res = await api.get(url)
      const { success, data: responseData } = res.data

      if (success && Array.isArray(responseData)) {
        const options: SearchOption[] = responseData.map((token: Token) => ({
          key: token.id.toString(),
          value: token.name,
          text: token.name,
          content: (
            <div className="flex flex-col">
              <div className="font-medium">{token.name}</div>
              <div className="text-sm text-muted-foreground">
                ID: {token.id} • {getStatusBadge(token.status)} • Quota: {formatQuota(token.remain_quota, token.unlimited_quota)}
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
      let url = `/api/token/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`

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

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete') => {
    try {
      let res: any
      if (action === 'delete') {
        // Unified API call - complete URL with /api prefix
        res = await api.delete(`/api/token/${id}`)
      } else {
        res = await api.put('/api/token/', {
          id,
          status: action === 'enable' ? TOKEN_STATUS.ENABLED : TOKEN_STATUS.DISABLED
        })
      }

      if (res.data?.success) {
        if (searchKeyword.trim()) {
          performSearch()
        } else {
          load(pageIndex, pageSize)
        }
      }
    } catch (error) {
      console.error(`Failed to ${action} token:`, error)
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
    } catch (error) {
      console.error('Failed to copy to clipboard:', error)
    }
  }

  const toggleKeyVisibility = (tokenId: number) => {
    setShowKeys(prev => ({
      ...prev,
      [tokenId]: !prev[tokenId]
    }))
  }

  const maskKey = (key: string) => {
    if (key.length <= 8) return '***'
    return key.substring(0, 4) + '***' + key.substring(key.length - 4)
  }

  const columns: ColumnDef<Token>[] = [
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
      accessorKey: 'key',
      header: 'Key',
      cell: ({ row }) => {
        const token = row.original
        const isVisible = showKeys[token.id]
        return (
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs">
              {isVisible ? token.key : maskKey(token.key)}
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => toggleKeyVisibility(token.id)}
              className="h-6 w-6 p-0"
            >
              {isVisible ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => copyToClipboard(token.key)}
              className="h-6 w-6 p-0"
            >
              <Copy className="h-3 w-3" />
            </Button>
          </div>
        )
      },
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => getStatusBadge(row.original.status),
    },
    {
      accessorKey: 'remain_quota',
      header: 'Remaining Quota',
      cell: ({ row }) => (
        <span className="font-mono text-sm" title={`Remaining: ${formatQuota(row.original.remain_quota, row.original.unlimited_quota)}`}>
          {formatQuota(row.original.remain_quota, row.original.unlimited_quota)}
        </span>
      ),
    },
    {
      accessorKey: 'used_quota',
      header: 'Used Quota',
      cell: ({ row }) => (
        <span className="font-mono text-sm" title={`Used: ${formatQuota(row.original.used_quota)}`}>
          {formatQuota(row.original.used_quota)}
        </span>
      ),
    },
    {
      accessorKey: 'created_time',
      header: 'Created',
      cell: ({ row }) => (
        <span className="text-sm">{formatTimestamp(row.original.created_time)}</span>
      ),
    },
    {
      accessorKey: 'accessed_time',
      header: 'Last Access',
      cell: ({ row }) => (
        <span className="text-sm">{formatTimestamp(row.original.accessed_time)}</span>
      ),
    },
    {
      accessorKey: 'expired_time',
      header: 'Expires',
      cell: ({ row }) => (
        <span className="text-sm">{formatTimestamp(row.original.expired_time)}</span>
      ),
    },
    {
      header: 'Actions',
      cell: ({ row }) => {
        const token = row.original
        return (
          <div className="flex items-center gap-1 mobile-table-cell">
            <Button
              variant="outline"
              size="sm"
              onClick={() => navigate(`/tokens/edit/${token.id}`)}
              className="touch-target"
            >
              Edit
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => manage(token.id, token.status === TOKEN_STATUS.ENABLED ? 'disable' : 'enable')}
              className={cn(
                "touch-target",
                token.status === TOKEN_STATUS.ENABLED
                  ? 'text-orange-600 hover:text-orange-700'
                  : 'text-green-600 hover:text-green-700'
              )}
            >
              {token.status === TOKEN_STATUS.ENABLED ? 'Disable' : 'Enable'}
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => {
                if (confirm(`Are you sure you want to delete token "${token.name}"?`)) {
                  manage(token.id, 'delete')
                }
              }}
              className="touch-target"
            >
              Delete
            </Button>
          </div>
        )
      },
    },
  ]

  const handlePageChange = (newPageIndex: number, newPageSize: number) => {
    if (searchKeyword.trim()) {
      // For search results, we don't do server-side pagination
      setPageIndex(newPageIndex)
    } else {
      load(newPageIndex, newPageSize)
    }
  }

  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize)
    if (searchKeyword.trim()) {
      performSearch()
    } else {
      load(0, newPageSize)
    }
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

  return (
    <ResponsivePageContainer
      title="Tokens"
      description="Manage your API access tokens"
      actions={
        <Button
          onClick={() => navigate('/tokens/add')}
          className={cn(
            "gap-2",
            isMobile ? "w-full touch-target" : ""
          )}
        >
          <Plus className="h-4 w-4" />
          {isMobile ? "Add New Token" : "Add Token"}
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
            onSearchChange={searchTokens}
            onSearchValueChange={setSearchKeyword}
            onSearchSubmit={performSearch}
            searchPlaceholder="Search tokens by name..."
            allowSearchAdditions={true}
            onRefresh={refresh}
            loading={loading}
            emptyMessage="No tokens found. Create your first token to get started."
            mobileCardLayout={true}
            hideColumnsOnMobile={['created_time', 'accessed_time', 'expired_time']}
            compactMode={isMobile}
          />
        </CardContent>
      </Card>
    </ResponsivePageContainer>
  )
}

export default TokensPage
