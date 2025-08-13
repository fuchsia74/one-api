import { useState, useEffect } from 'react'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

import {
  Clock,
  Globe,
  Activity,
  Zap,
  CheckCircle,
  ArrowRight,
  Play,
  Send,
  Reply,
  Flag,
  User,
  Hash,
  FileText,
  AlertCircle
} from 'lucide-react'
import { api } from '@/lib/api'
import { formatTimestamp, cn } from '@/lib/utils'

interface TraceTimestamps {
  request_received?: number
  request_forwarded?: number
  first_upstream_response?: number
  first_client_response?: number
  upstream_completed?: number
  request_completed?: number
}

interface TraceDurations {
  processing_time?: number
  upstream_response_time?: number
  response_processing_time?: number
  streaming_time?: number
  total_time?: number
}

interface TraceData {
  id: number
  trace_id: string
  url: string
  method: string
  body_size: number
  status: number
  created_at: number
  updated_at: number
  timestamps: TraceTimestamps
  durations?: TraceDurations
  log?: {
    id: number
    user_id: number
    username: string
    content: string
    type: number
  }
}

interface TracingModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  logId: number | null
  traceId?: string | null
}

const formatDuration = (milliseconds?: number): string => {
  if (!milliseconds) return 'N/A'
  if (milliseconds < 1000) {
    return `${milliseconds}ms`
  }
  return `${(milliseconds / 1000).toFixed(2)}s`
}

const getStatusColor = (status: number): string => {
  if (status >= 200 && status < 300) return 'bg-green-500'
  if (status >= 300 && status < 400) return 'bg-yellow-500'
  if (status >= 400 && status < 500) return 'bg-orange-500'
  if (status >= 500) return 'bg-red-500'
  return 'bg-gray-500'
}

const getMethodColor = (method: string): string => {
  switch (method.toUpperCase()) {
    case 'GET': return 'bg-blue-500'
    case 'POST': return 'bg-green-500'
    case 'PUT': return 'bg-yellow-500'
    case 'DELETE': return 'bg-red-500'
    case 'PATCH': return 'bg-purple-500'
    default: return 'bg-gray-500'
  }
}

const timelineEvents = [
  {
    key: 'request_received' as keyof TraceTimestamps,
    title: 'Request Received',
    icon: Play,
    color: 'text-blue-500',
    description: 'Initial request received by the gateway'
  },
  {
    key: 'request_forwarded' as keyof TraceTimestamps,
    title: 'Forwarded to Upstream',
    icon: ArrowRight,
    color: 'text-teal-500',
    description: 'Request forwarded to upstream service'
  },
  {
    key: 'first_upstream_response' as keyof TraceTimestamps,
    title: 'First Upstream Response',
    icon: Reply,
    color: 'text-purple-500',
    description: 'First response received from upstream'
  },
  {
    key: 'first_client_response' as keyof TraceTimestamps,
    title: 'First Client Response',
    icon: Send,
    color: 'text-orange-500',
    description: 'First response sent to client'
  },
  {
    key: 'upstream_completed' as keyof TraceTimestamps,
    title: 'Upstream Completed',
    icon: CheckCircle,
    color: 'text-green-500',
    description: 'Upstream response completed (streaming)'
  },
  {
    key: 'request_completed' as keyof TraceTimestamps,
    title: 'Request Completed',
    icon: Flag,
    color: 'text-green-600',
    description: 'Request fully completed'
  }
]

export function TracingModal({ open, onOpenChange, logId, traceId }: TracingModalProps) {
  const [loading, setLoading] = useState(false)
  const [traceData, setTraceData] = useState<TraceData | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)


  useEffect(() => {
    if (open && logId && traceId && traceId.trim() !== '') {
      fetchTraceData()
    } else if (open && (!traceId || traceId.trim() === '')) {
      setTraceData(null)
      setError('No trace information available for this log entry.')
    }
  }, [open, logId, traceId])

  const fetchTraceData = async () => {
    if (!logId || !traceId || traceId.trim() === '') return

    setLoading(true)
    setError(null)

    try {
      const response = await api.get(`/api/trace/log/${logId}`)
      if (response.data.success) {
        setTraceData(response.data.data)
      } else {
        setError(response.data.message || 'Failed to fetch trace data')
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to fetch trace data')
    } finally {
      setLoading(false)
    }
  }

  const copyTraceId = async () => {
    if (!traceData?.trace_id) return

    try {
      await navigator.clipboard.writeText(traceData.trace_id)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy trace ID:', err)
    }
  }

  const renderRequestInfo = () => {
    if (!traceData) return null

    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-5 w-5" />
            Request Information
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Globe className="h-4 w-4" />
                URL
              </div>
              <div className="font-mono text-sm bg-muted p-2 rounded break-all">
                {traceData.url}
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Activity className="h-4 w-4" />
                Method & Status
              </div>
              <div className="flex gap-2">
                <Badge className={cn('text-white', getMethodColor(traceData.method))}>
                  {traceData.method}
                </Badge>
                <Badge className={cn('text-white', getStatusColor(traceData.status))}>
                  {traceData.status || 'N/A'}
                </Badge>
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <FileText className="h-4 w-4" />
                Request Size
              </div>
              <div className="text-sm">
                {traceData.body_size ? `${traceData.body_size} bytes` : 'N/A'}
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <User className="h-4 w-4" />
                User
              </div>
              <div className="text-sm">
                {traceData.log?.username || 'N/A'}
              </div>
            </div>

            <div className="col-span-1 md:col-span-2 space-y-2">
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Hash className="h-4 w-4" />
                Trace ID
              </div>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <div
                      className="font-mono text-xs bg-muted p-2 rounded break-all cursor-pointer hover:bg-muted/80 transition-colors"
                      onClick={copyTraceId}
                    >
                      {traceData.trace_id}
                    </div>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{copied ? 'Copied!' : 'Click to copy trace ID'}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  const renderTimeline = () => {
    if (!traceData?.timestamps) return null

    const { timestamps, durations } = traceData
    const activeEvents = timelineEvents.filter(event => timestamps[event.key])

    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5" />
            Request Timeline
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {activeEvents.map((event, index) => {
              const timestamp = timestamps[event.key]
              const Icon = event.icon
              const isLast = index === activeEvents.length - 1

              // Calculate duration for this step
              let duration: number | undefined
              if (event.key === 'request_forwarded') duration = durations?.processing_time
              else if (event.key === 'first_upstream_response') duration = durations?.upstream_response_time
              else if (event.key === 'first_client_response') duration = durations?.response_processing_time
              else if (event.key === 'upstream_completed') duration = durations?.streaming_time

              return (
                <div key={event.key} className="relative">
                  <div className="flex items-start gap-4">
                    <div className={cn(
                      'flex items-center justify-center w-8 h-8 rounded-full border-2 bg-background',
                      event.color.replace('text-', 'border-')
                    )}>
                      <Icon className={cn('h-4 w-4', event.color)} />
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <h4 className="font-medium">{event.title}</h4>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          {duration && (
                            <Badge variant="outline" className="text-xs">
                              +{formatDuration(duration)}
                            </Badge>
                          )}
                          <span className="font-mono">
                            {formatTimestamp(Math.floor(timestamp! / 1000))}
                          </span>
                        </div>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        {event.description}
                      </p>
                    </div>
                  </div>

                  {!isLast && (
                    <div className="absolute left-4 top-8 w-px h-6 bg-border" />
                  )}
                </div>
              )
            })}
          </div>

          {durations?.total_time && (
            <div className="mt-6 p-4 bg-primary/5 rounded-lg border border-primary/20">
              <div className="flex items-center gap-2">
                <Zap className="h-5 w-5 text-primary" />
                <span className="font-semibold">Total Request Time:</span>
                <Badge variant="default" className="ml-auto">
                  {formatDuration(durations.total_time)}
                </Badge>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Request Tracing Details
          </DialogTitle>
        </DialogHeader>

        <ScrollArea className="max-h-[calc(90vh-8rem)]">
          <div className="space-y-6 pr-4">
            {loading && (
              <div className="space-y-4">
                <Skeleton className="h-32 w-full" />
                <Skeleton className="h-48 w-full" />
              </div>
            )}

            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {traceData && !loading && (
              <>
                {renderRequestInfo()}
                <Separator />
                {renderTimeline()}
              </>
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
