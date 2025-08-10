import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTable } from '@/components/ui/data-table'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useForm } from 'react-hook-form'
import * as z from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { formatTimestamp } from '@/lib/utils'

interface RedemptionRow {
  id: number
  name: string
  key: string
  status: number
  created_time: number
  quota: number
}

const renderStatus = (status: number) => {
  const map: Record<number, { text: string; cls: string }> = {
    1: { text: 'Unused', cls: 'text-green-600' },
    2: { text: 'Disabled', cls: 'text-red-600' },
    3: { text: 'Used', cls: 'text-gray-600' },
  }
  const v = map[status] || { text: 'Unknown', cls: 'text-muted-foreground' }
  return <span className={`text-sm ${v.cls}`}>{v.text}</span>
}

export function RedemptionsPage() {
  const navigate = useNavigate()
  const [data, setData] = useState<RedemptionRow[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [sortBy, setSortBy] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [open, setOpen] = useState(false)
  const [generatedKeys, setGeneratedKeys] = useState<string[] | null>(null)

  const schema = z.object({
    name: z.string().min(1, 'Name is required').max(20, 'Max 20 chars'),
    count: z.coerce.number().int().min(1).max(100),
    quota: z.coerce.number().int().min(0),
  })
  type CreateForm = z.infer<typeof schema>
  const form = useForm<CreateForm>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', count: 1, quota: 0 },
  })

  const load = async (p = 0) => {
    setLoading(true)
    try {
      let url = `/redemption/?p=${p}`
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
      let url = `/redemption/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data } = res.data
      if (success) {
        setData(data)
        setPageIndex(0)
      }
    } finally { setLoading(false) }
  }

  const columns: ColumnDef<RedemptionRow>[] = [
    { header: 'ID', accessorKey: 'id' },
    { header: 'Name', accessorKey: 'name' },
    { header: 'Code', accessorKey: 'key' },
    { header: 'Quota', accessorKey: 'quota' },
    { header: 'Status', cell: ({ row }) => renderStatus(row.original.status) },
    { header: 'Created', cell: ({ row }) => formatTimestamp(row.original.created_time) },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/redemptions/edit/${row.original.id}`)}
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
          <Button variant="destructive" size="sm" onClick={() => manage(row.original.id, 'delete', row.index)}>Delete</Button>
        </div>
      ),
    },
  ]

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete', idx: number) => {
    let res
    if (action === 'delete') {
      res = await api.delete(`/redemption/${id}`)
    } else {
      const body: any = { id, status: action === 'enable' ? 1 : 2 }
      res = await api.put('/redemption/?status_only=true', body)
    }
    if (res.data?.success) {
      const next = [...data]
      if (action === 'delete') next.splice(idx, 1)
      else next[idx].status = action === 'enable' ? 1 : 2
      setData(next)
    }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Redemptions</CardTitle>
              <CardDescription>Manage recharge codes</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button onClick={() => navigate('/redemptions/add')}>Add Redemption</Button>
              <select className="h-9 border rounded-md px-2 text-sm" value={sortBy} onChange={(e) => { setSortBy(e.target.value); setSortOrder('desc') }}>
                <option value="">Default</option>
                <option value="id">ID</option>
                <option value="name">Name</option>
                <option value="status">Status</option>
                <option value="quota">Quota</option>
                <option value="created_time">Created Time</option>
                <option value="redeemed_time">Redeemed Time</option>
              </select>
              <Button variant="outline" size="sm" onClick={() => setSortOrder(o => o === 'asc' ? 'desc' : 'asc')}>{sortOrder.toUpperCase()}</Button>
              <Button onClick={() => load(pageIndex)} disabled={loading} variant="outline">Refresh</Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 mb-3">
            <Input placeholder="Search redemptions" value={searchKeyword} onChange={(e) => setSearchKeyword(e.target.value)} />
            <Button onClick={search} disabled={loading}>Search</Button>
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

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Generate Redemption Codes</DialogTitle>
          </DialogHeader>
          <Form {...form}>
            <form className="space-y-3" onSubmit={form.handleSubmit(async (values) => {
              const res = await api.post('/redemption/', { name: values.name, count: values.count, quota: values.quota })
              if (res.data?.success) {
                setGeneratedKeys(res.data.data || [])
                load(pageIndex)
              }
            })}>
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl><Input {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="count" render={({ field }) => (
                <FormItem>
                  <FormLabel>Count (1-100)</FormLabel>
                  <FormControl><Input type="number" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <FormField control={form.control} name="quota" render={({ field }) => (
                <FormItem>
                  <FormLabel>Quota</FormLabel>
                  <FormControl><Input type="number" {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
              <div className="pt-2 flex justify-end gap-2">
                <Button type="button" variant="outline" onClick={() => setOpen(false)}>Close</Button>
                <Button type="submit">Generate</Button>
              </div>

              {generatedKeys && (
                <div className="mt-4">
                  <div className="text-sm mb-2">Generated Codes:</div>
                  <textarea className="w-full h-32 p-2 border rounded" readOnly value={generatedKeys.join('\n')} />
                </div>
              )}
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
