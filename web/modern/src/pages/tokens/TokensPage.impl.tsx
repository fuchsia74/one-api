import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTable } from '@/components/ui/data-table'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { useForm } from 'react-hook-form'
import * as z from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { fromDateTimeLocal, toDateTimeLocal } from '@/lib/utils'

interface Token {
  id: number
  name: string
  key: string
  remain_quota: number
  status: number
  created_time: number
}

export function TokensPage() {
  const navigate = useNavigate()
  const [data, setData] = useState<Token[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchOptions, setSearchOptions] = useState<SearchOption[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [sortBy, setSortBy] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [open, setOpen] = useState(false)

  const load = async (p = 0) => {
    setLoading(true)
    try {
      let url = `/token/?p=${p}`
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

  const searchTokens = async (query: string) => {
    if (!query.trim()) {
      setSearchOptions([])
      return
    }

    setSearchLoading(true)
    try {
      const res = await api.get(`/token/search?keyword=${encodeURIComponent(query)}`)
      const { success, data } = res.data
      if (success && Array.isArray(data)) {
        const options: SearchOption[] = data.map((token: Token) => ({
          key: token.id.toString(),
          value: token.name,
          text: token.name,
          content: (
            <div className="flex flex-col">
              <div className="font-medium">{token.name}</div>
              <div className="text-sm text-muted-foreground">
                ID: {token.id} • Status: {token.status === 1 ? 'Enabled' : 'Disabled'}
              </div>
            </div>
          )
        }))
        setSearchOptions(options)
      }
    } catch (error) {
      console.error('Search failed:', error)
    } finally {
      setSearchLoading(false)
    }
  }

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete') => {
    let res
    if (action === 'delete') res = await api.delete(`/token/${id}`)
    else res = await api.put('/token/?status_only=true', { id, status: action === 'enable' ? 1 : 2 })
    if (res.data?.success) load(pageIndex)
  }

  const search = async () => {
    if (!searchKeyword.trim()) return load(0)
    setLoading(true)
    try {
      let url = `/token/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data } = res.data
      if (success) {
        setData(data)
        setPageIndex(0)
        setTotal(data.length)
      }
    } finally { setLoading(false) }
  }

  // Create token dialog
  const schema = z.object({
    name: z.string().min(1, 'Name is required').max(30),
    unlimited_quota: z.boolean().default(false),
    remain_quota: z.coerce.number().int().min(0).default(0),
    expired_time: z.string().optional(), // datetime-local, empty => -1
    subnet: z.string().optional(),
  })
  type CreateForm = z.infer<typeof schema>
  const form = useForm<CreateForm>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', unlimited_quota: false, remain_quota: 0, expired_time: '', subnet: '' },
  })

  const columns: ColumnDef<Token>[] = [
    { header: 'ID', accessorKey: 'id' },
    { header: 'Name', accessorKey: 'name' },
    { header: 'Key', accessorKey: 'key' },
    { header: 'Remain', accessorKey: 'remain_quota' },
    { header: 'Created', accessorKey: 'created_time' },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/tokens/edit/${row.original.id}`)}
          >
            Edit
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => manage(row.original.id, row.original.status === 1 ? 'disable' : 'enable')}
          >
            {row.original.status === 1 ? 'Disable' : 'Enable'}
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={() => manage(row.original.id, 'delete')}
          >
            Delete
          </Button>
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
              <CardTitle>Tokens</CardTitle>
              <CardDescription>Manage your access tokens</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button onClick={() => navigate('/tokens/add')}>
                Add Token
              </Button>
              <select className="h-9 border rounded-md px-2 text-sm" value={sortBy} onChange={(e) => { setSortBy(e.target.value); setSortOrder('desc') }}>
                <option value="">Default</option>
                <option value="id">ID</option>
                <option value="name">Name</option>
                <option value="remain_quota">Remain</option>
                <option value="created_time">Created Time</option>
              </select>
              <Button variant="outline" size="sm" onClick={() => setSortOrder(o => o === 'asc' ? 'desc' : 'asc')}>{sortOrder.toUpperCase()}</Button>
              <Button onClick={() => load(pageIndex)} disabled={loading} variant="outline">Refresh</Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 mb-3">
            <div className="flex-1">
              <SearchableDropdown
                value={searchKeyword}
                placeholder="Search tokens..."
                searchPlaceholder="Search by token name..."
                options={searchOptions}
                onSearchChange={searchTokens}
                onChange={(value) => setSearchKeyword(value)}
                onAddItem={(value) => {
                  const newOption: SearchOption = {
                    key: value,
                    value: value,
                    text: value
                  }
                  setSearchOptions([...searchOptions, newOption])
                }}
                loading={searchLoading}
                noResultsMessage="No tokens found"
                additionLabel="Use token name: "
                allowAdditions={true}
                clearable={true}
              />
            </div>
            <Button onClick={search} disabled={loading}>Search</Button>
          </div>
          <DataTable
            columns={columns}
            data={data}
            pageIndex={pageIndex}
            pageSize={pageSize}
            total={total}
            onPageChange={(pi) => load(pi)}
            onPageSizeChange={(newPageSize) => {
              setPageSize(newPageSize)
              setPageIndex(0) // Reset to first page
            }}
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
          <DialogHeader><DialogTitle>Create Token</DialogTitle></DialogHeader>
          <Form {...form}>
            <form className="space-y-3" onSubmit={form.handleSubmit(async (values) => {
              const expired = values.expired_time ? fromDateTimeLocal(values.expired_time) : -1
              const payload: any = {
                name: values.name,
                unlimited_quota: values.unlimited_quota,
                remain_quota: values.unlimited_quota ? 0 : values.remain_quota,
                expired_time: expired,
              }
              if (values.subnet) payload.subnet = values.subnet
              const res = await api.post('/token/', payload)
              if (res.data?.success) {
                setOpen(false)
                form.reset()
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
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <FormField control={form.control} name="remain_quota" render={({ field }) => (
                  <FormItem>
                    <FormLabel>Remain Quota</FormLabel>
                    <FormControl><Input type="number" {...field} disabled={form.watch('unlimited_quota')} /></FormControl>
                    <FormMessage />
                  </FormItem>
                )} />
                <FormField control={form.control} name="expired_time" render={({ field }) => (
                  <FormItem>
                    <FormLabel>Expired Time</FormLabel>
                    <FormControl><Input type="datetime-local" placeholder="Never" {...field} /></FormControl>
                    <div className="text-xs text-muted-foreground">Leave empty for never expire</div>
                    <FormMessage />
                  </FormItem>
                )} />
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <FormField control={form.control} name="unlimited_quota" render={({ field }) => (
                  <FormItem>
                    <FormLabel>Unlimited Quota</FormLabel>
                    <FormControl>
                      <input type="checkbox" className="h-4 w-4" checked={field.value} onChange={(e)=>field.onChange(e.target.checked)} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )} />
                <FormField control={form.control} name="subnet" render={({ field }) => (
                  <FormItem>
                    <FormLabel>Subnet (optional)</FormLabel>
                    <FormControl><Input placeholder="192.168.0.0/24,10.0.0.0/8" {...field} /></FormControl>
                    <FormMessage />
                  </FormItem>
                )} />
              </div>
              <div className="pt-2 flex justify-end gap-2">
                <Button type="button" variant="outline" onClick={() => setOpen(false)}>Close</Button>
                <Button type="submit">Create</Button>
              </div>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default TokensPage
