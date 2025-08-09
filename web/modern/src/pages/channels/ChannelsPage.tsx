import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTable } from '@/components/ui/data-table'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { formatTimestamp } from '@/lib/utils'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { useForm } from 'react-hook-form'
import * as z from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'

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
  const navigate = useNavigate()
  const [data, setData] = useState<ChannelRow[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchOptions, setSearchOptions] = useState<SearchOption[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [sortBy, setSortBy] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [openCreate, setOpenCreate] = useState(false)

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

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete' | 'test' | 'balance', idx: number) => {
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
    if (action === 'balance') {
      await api.get(`/channel/update_balance/${id}`)
      load(pageIndex)
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
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/channels/edit/${row.original.id}`)}
          >
            Edit
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => manage(row.original.id, row.original.status === 1 ? 'disable' : 'enable', row.index)}
          >
            {row.original.status === 1 ? 'Disable' : 'Enable'}
          </Button>
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
              <Button onClick={() => navigate('/channels/add')}>Add Channel</Button>
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
            <Button variant="outline" onClick={async ()=>{ await api.get('/channel/test'); load(pageIndex) }}>Test All</Button>
            <Button variant="outline" onClick={async ()=>{ await api.get('/channel/update_balance'); load(pageIndex) }}>Update All Balances</Button>
            <Button variant="destructive" onClick={async ()=>{ await api.delete('/channel/disabled'); load(pageIndex) }}>Delete Disabled</Button>
          </div>
          <DataTable
            columns={columns}
            data={data}
            pageIndex={pageIndex}
            pageSize={pageSize}
            total={total}
            onPageChange={(pi) => load(pi)}
            sortBy={sortBy}
            sortOrder={sortOrder}
            onSortChange={(newSortBy, newSortOrder) => {
              setSortBy(newSortBy)
              setSortOrder(newSortOrder)
            }}
            loading={loading}
          />
        </CardContent>
      </Card>

      <CreateChannelDialog open={openCreate} onOpenChange={setOpenCreate} onCreated={()=>load(pageIndex)} />
    </div>
  )
}

function CreateChannelDialog({ open, onOpenChange, onCreated }: { open: boolean, onOpenChange: (v:boolean)=>void, onCreated: ()=>void }) {
  const schema = z.object({
    name: z.string().min(1),
    type: z.coerce.number().int().min(0),
    key: z.string().min(1), // supports multi-line
  })
  type FormT = z.infer<typeof schema>
  const form = useForm<FormT>({ resolver: zodResolver(schema), defaultValues: { name: '', type: 0, key: '' } })
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader><DialogTitle>Create Channel</DialogTitle></DialogHeader>
        <Form {...form}>
          <form className="space-y-3" onSubmit={form.handleSubmit(async (values) => {
            const res = await api.post('/channel/', { name: values.name, type: values.type, key: values.key })
            if (res.data?.success) {
              onOpenChange(false)
              form.reset()
              onCreated()
            }
          })}>
            <FormField control={form.control} name="name" render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl><Input {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="type" render={({ field }) => (
              <FormItem>
                <FormLabel>Type</FormLabel>
                <FormControl><Input type="number" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="key" render={({ field }) => (
              <FormItem>
                <FormLabel>Key(s)</FormLabel>
                <FormControl><textarea className="w-full h-32 p-2 border rounded" placeholder="One key per line" {...field as any} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <div className="pt-2 flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
              <Button type="submit">Create</Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
