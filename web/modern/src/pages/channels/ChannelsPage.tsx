import { useEffect, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTable } from '@/components/ui/data-table'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { formatTimestamp } from '@/lib/utils'

interface ChannelRow {
  id: number
  name: string
  type: number
  status: number
  response_time?: number
  created_time: number
  priority?: number
}

const renderStatus = (status: number, priority?: number) => {
  let color = 'text-green-600'
  let text = 'Active'
  if (status === 2) { color = 'text-red-600'; text = 'Disabled' }
  else if ((priority ?? 0) < 0) { color = 'text-orange-600'; text = 'Paused' }
  return <span className={`text-sm ${color}`}>{text}</span>
}

export function ChannelsPage() {
  const [data, setData] = useState<ChannelRow[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [sortBy, setSortBy] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')

  const load = async (p = 0) => {
    setLoading(true)
    try {
      let url = `/channel/?p=${p}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data, total } = res.data
      if (success) {
        setData(data)
        setTotal(total)
        setPageIndex(p)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load(0) }, [])
  useEffect(() => { load(0) }, [sortBy, sortOrder])

  const search = async () => {
    if (!searchKeyword.trim()) return load(0)
    setLoading(true)
    try {
      let url = `/channel/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data } = res.data
      if (success) {
        setData(data)
        setPageIndex(0)
      }
    } finally { setLoading(false) }
  }

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete' | 'test', idx: number) => {
    if (action === 'delete') {
  const res = await api.delete(`/channel/${id}`)
      if (res.data.success) load(pageIndex)
      return
    }
    if (action === 'test') {
  const res = await api.get(`/channel/test/${id}`)
      const { success, time } = res.data
      if (success) {
        const next = [...data]
        next[idx].response_time = time
        setData(next)
      }
      return
    }
    const body: any = { id, status: action === 'enable' ? 1 : 2 }
  const res = await api.put('/channel/', body)
    if (res.data.success) load(pageIndex)
  }

  const columns: ColumnDef<ChannelRow>[] = [
    { header: 'ID', accessorKey: 'id' },
    { header: 'Name', accessorKey: 'name' },
    { header: 'Type', accessorKey: 'type' },
    { header: 'Status', cell: ({ row }) => renderStatus(row.original.status, row.original.priority) },
    { header: 'Resp (ms)', accessorKey: 'response_time' },
    { header: 'Created', cell: ({ row }) => formatTimestamp(row.original.created_time) },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <div className="space-x-2">
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'enable', row.index)}>Enable</Button>
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'disable', row.index)}>Disable</Button>
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'test', row.index)}>Test</Button>
          <Button variant="destructive" size="sm" onClick={() => manage(row.original.id, 'delete', row.index)}>Delete</Button>
        </div>
      ),
    },
  ]

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Channels</CardTitle>
              <CardDescription>Configure and manage routing channels</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <select className="h-9 border rounded-md px-2 text-sm" value={sortBy} onChange={(e) => { setSortBy(e.target.value); setSortOrder('desc') }}>
                <option value="">Default</option>
                <option value="id">ID</option>
                <option value="name">Name</option>
                <option value="type">Type</option>
                <option value="status">Status</option>
                <option value="response_time">Response Time</option>
                <option value="created_time">Created Time</option>
              </select>
              <Button variant="outline" size="sm" onClick={() => setSortOrder(o => o === 'asc' ? 'desc' : 'asc')}>{sortOrder.toUpperCase()}</Button>
              <Button onClick={() => load(pageIndex)} disabled={loading} variant="outline">Refresh</Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 mb-3">
            <input className="h-9 border rounded-md px-2 text-sm w-full" placeholder="Search channels" value={searchKeyword} onChange={(e) => setSearchKeyword(e.target.value)} />
            <Button onClick={search} disabled={loading}>Search</Button>
          </div>
          <DataTable columns={columns} data={data} pageIndex={pageIndex} pageSize={pageSize} total={total} onPageChange={(pi) => load(pi)} />
        </CardContent>
      </Card>
    </div>
  )
}
