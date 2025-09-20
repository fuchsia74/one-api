import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useResponsive } from '@/hooks/useResponsive'
import { useNotifications } from '@/components/ui/notifications'
import { RefreshCw, Activity, Clock, Calendar, Zap, AlertCircle, CheckCircle, XCircle, ChevronLeft, ChevronRight } from 'lucide-react'

interface ChannelStatus {
  name: string
  status: string
  enabled: boolean
  response: {
    response_time_ms: number
    test_time: number
    created_time: number
  }
}

interface StatusResponse {
  success: boolean
  data: ChannelStatus[]
  total?: number
  message?: string
}

function StatusPageImpl() {
  const { isMobile, isTablet } = useResponsive()
  const { notify } = useNotifications()
  const [channelsData, setChannelsData] = useState<ChannelStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [refreshing, setRefreshing] = useState(false)

  // Pagination state
  const [currentPage, setCurrentPage] = useState(0)
  const [pageSize, setPageSize] = useState(6)
  const [totalCount, setTotalCount] = useState(0)
  const [totalPages, setTotalPages] = useState(0)

  const fetchStatusData = async (page: number = currentPage, size: number = pageSize) => {
    try {
      setLoading(true)
      const params = new URLSearchParams({
        p: page.toString(),
        size: size.toString()
      })
      const res = await api.get(`/api/status/channel?${params}`)
      const { success, message, data, total }: StatusResponse = res.data
      if (success) {
        setChannelsData(data || [])
        setTotalCount(total || 0)
        setTotalPages(Math.ceil((total || 0) / size))
      } else {
        notify({ message: `Failed to fetch channel status: ${message}`, type: 'error' })
      }
    } catch (error) {
      notify({ message: `Error fetching channel status: ${error}`, type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const handleRefresh = async () => {
    setRefreshing(true)
    await fetchStatusData(currentPage, pageSize)
    setRefreshing(false)
  }

  const handlePageChange = async (newPage: number) => {
    if (newPage >= 0 && newPage < totalPages) {
      setCurrentPage(newPage)
      await fetchStatusData(newPage, pageSize)
    }
  }

  const handlePreviousPage = () => {
    if (currentPage > 0) {
      handlePageChange(currentPage - 1)
    }
  }

  const handleNextPage = () => {
    if (currentPage < totalPages - 1) {
      handlePageChange(currentPage + 1)
    }
  }

  useEffect(() => {
    fetchStatusData(0, pageSize)
  }, [])

  const formatTimestamp = (timestamp: number): string => {
    if (!timestamp) return 'Never'
    const date = new Date(timestamp * 1000)
    return date.toLocaleString()
  }

  const formatResponseTime = (responseTime: number): string => {
    if (responseTime === 0) return 'N/A'
    if (responseTime < 1000) return `${responseTime}ms`
    return `${(responseTime / 1000).toFixed(2)}s`
  }

  const getStatusBadge = (status: string, enabled: boolean) => {
    if (enabled && status === 'enabled') {
      return (
        <Badge variant="default" className="bg-green-100 text-green-800 hover:bg-green-200 dark:bg-green-900 dark:text-green-200">
          <CheckCircle className="w-3 h-3 mr-1" />
          Enabled
        </Badge>
      )
    } else if (status === 'manually_disabled') {
      return (
        <Badge variant="secondary" className="bg-yellow-100 text-yellow-800 hover:bg-yellow-200 dark:bg-yellow-900 dark:text-yellow-200">
          <AlertCircle className="w-3 h-3 mr-1" />
          Manually Disabled
        </Badge>
      )
    } else if (status === 'auto_disabled') {
      return (
        <Badge variant="destructive" className="bg-red-100 text-red-800 hover:bg-red-200 dark:bg-red-900 dark:text-red-200">
          <XCircle className="w-3 h-3 mr-1" />
          Auto Disabled
        </Badge>
      )
    } else {
      return (
        <Badge variant="outline" className="bg-gray-100 text-gray-800 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-200">
          <AlertCircle className="w-3 h-3 mr-1" />
          Unknown
        </Badge>
      )
    }
  }

  const getResponseTimeBadge = (responseTime: number) => {
    if (responseTime === 0) {
      return <Badge variant="outline">N/A</Badge>
    } else if (responseTime < 1000) {
      return <Badge variant="default" className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">Fast</Badge>
    } else if (responseTime < 3000) {
      return <Badge variant="secondary" className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">Normal</Badge>
    } else {
      return <Badge variant="destructive" className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">Slow</Badge>
    }
  }

  // Filter channels based on search term
  const filteredChannels = channelsData.filter(channel => {
    if (!searchTerm) return true
    const searchLower = searchTerm.toLowerCase()
    return (
      channel.name.toLowerCase().includes(searchLower) ||
      channel.status.toLowerCase().includes(searchLower) ||
      (channel.enabled ? 'enabled' : 'disabled').includes(searchLower)
    )
  })

  const enabledChannels = filteredChannels.filter(channel => channel.enabled).length
  const disabledChannels = filteredChannels.filter(channel => !channel.enabled).length
  const displayedChannels = filteredChannels.length

  if (loading) {
    return (
      <div className="container mx-auto p-4 max-w-7xl">
        <div className="space-y-6">
          <div className="text-center space-y-2">
            <div className="animate-spin mx-auto w-8 h-8">
              <RefreshCw className="w-8 h-8" />
            </div>
            <p className="text-muted-foreground">Loading channel status...</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-4 max-w-7xl">
      <div className="space-y-6">
        {/* Header */}
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <h1 className="text-3xl font-bold">Channel Status</h1>
            <Button
              onClick={handleRefresh}
              disabled={refreshing}
              variant="outline"
              size="sm"
              className="flex items-center gap-2"
            >
              <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
              {refreshing ? 'Refreshing...' : 'Refresh'}
            </Button>
          </div>
          <p className="text-muted-foreground">
            Monitor the status and performance of all channels
          </p>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center space-x-2">
                <CheckCircle className="w-5 h-5 text-green-600" />
                <div>
                  <p className="text-2xl font-bold text-green-600">{enabledChannels}</p>
                  <p className="text-sm text-muted-foreground">Enabled</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center space-x-2">
                <XCircle className="w-5 h-5 text-red-600" />
                <div>
                  <p className="text-2xl font-bold text-red-600">{disabledChannels}</p>
                  <p className="text-sm text-muted-foreground">Disabled</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center space-x-2">
                <Activity className="w-5 h-5 text-blue-600" />
                <div>
                  <p className="text-2xl font-bold text-blue-600">{searchTerm ? displayedChannels : totalCount}</p>
                  <p className="text-sm text-muted-foreground">{searchTerm ? 'Found' : 'Total Channels'}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Search and Controls */}
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <Input
              placeholder="Search channels by name or status..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full"
            />
          </div>
          {searchTerm && (
            <Button
              variant="outline"
              onClick={() => setSearchTerm('')}
              className="whitespace-nowrap"
            >
              Clear Search
            </Button>
          )}
        </div>

        {/* Channel Status Cards */}
        <div className="space-y-4">
          {filteredChannels.length === 0 ? (
            <Card>
              <CardContent className="p-8 text-center">
                <Activity className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
                <h3 className="text-lg font-semibold mb-2">No Channels Found</h3>
                <p className="text-muted-foreground">
                  {searchTerm
                    ? `No channels match your search "${searchTerm}"`
                    : 'No channels are configured in the system'
                  }
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
              {filteredChannels.map((channel, index) => (
                <Card key={index} className="hover:shadow-md transition-shadow">
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-lg truncate">{channel.name}</CardTitle>
                      {getStatusBadge(channel.status, channel.enabled)}
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    {/* Response Time */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Clock className="w-4 h-4 text-muted-foreground" />
                        <span className="text-sm text-muted-foreground">Response Time</span>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span className="font-mono text-sm">{formatResponseTime(channel.response.response_time_ms)}</span>
                        {getResponseTimeBadge(channel.response.response_time_ms)}
                      </div>
                    </div>

                    {/* Test Time */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Zap className="w-4 h-4 text-muted-foreground" />
                        <span className="text-sm text-muted-foreground">Last Test</span>
                      </div>
                      <span className="text-sm font-mono">{formatTimestamp(channel.response.test_time)}</span>
                    </div>

                    {/* Created Time */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Calendar className="w-4 h-4 text-muted-foreground" />
                        <span className="text-sm text-muted-foreground">Created</span>
                      </div>
                      <span className="text-sm font-mono">{formatTimestamp(channel.response.created_time)}</span>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>

        {/* Pagination Controls */}
        {totalPages > 1 && !searchTerm && (
          <div className="flex items-center justify-center space-x-4 mt-6">
            <Button
              variant="outline"
              size="sm"
              onClick={handlePreviousPage}
              disabled={currentPage === 0}
              className="flex items-center gap-2"
            >
              <ChevronLeft className="w-4 h-4" />
              Previous
            </Button>

            <div className="flex items-center space-x-2">
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum;
                if (totalPages <= 5) {
                  pageNum = i;
                } else if (currentPage < 2) {
                  pageNum = i;
                } else if (currentPage >= totalPages - 2) {
                  pageNum = totalPages - 5 + i;
                } else {
                  pageNum = currentPage - 2 + i;
                }

                return (
                  <Button
                    key={pageNum}
                    variant={currentPage === pageNum ? "default" : "outline"}
                    size="sm"
                    onClick={() => handlePageChange(pageNum)}
                    className="w-10 h-10 p-0"
                  >
                    {pageNum + 1}
                  </Button>
                );
              })}
            </div>

            <Button
              variant="outline"
              size="sm"
              onClick={handleNextPage}
              disabled={currentPage >= totalPages - 1}
              className="flex items-center gap-2"
            >
              Next
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        )}

        {/* Footer Info */}
        {filteredChannels.length > 0 && (
          <div className="text-center text-sm text-muted-foreground">
            {searchTerm ? (
              `Showing ${filteredChannels.length} of ${totalCount} channels (filtered)`
            ) : (
              `Showing ${channelsData.length} of ${totalCount} channels${totalPages > 1 ? ` (Page ${currentPage + 1} of ${totalPages})` : ''}`
            )}
          </div>
        )}
      </div>
    </div>
  )
}

export function StatusPage() {
  return <StatusPageImpl />
}

export default StatusPage
