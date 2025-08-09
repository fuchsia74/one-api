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

interface UserRow {
  id: number
  username: string
  display_name?: string
  role: number
  status: number
  email?: string
  quota: number
  used_quota: number
  group: string
}

export function UsersPage() {
  const navigate = useNavigate()
  const [data, setData] = useState<UserRow[]>([])
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
  const [openTopup, setOpenTopup] = useState<{open: boolean, userId?: number, username?: string}>({open: false})

  const load = async (p = 0) => {
    setLoading(true)
    try {
      let url = `/user/?p=${p}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data, total } = res.data
      if (success) {
        setData(data)
        setTotal(total || data.length)
        setPageIndex(p)
      }
    } finally {
      setLoading(false)
    }
  }

  const searchUsers = async (query: string) => {
    if (!query.trim()) {
      setSearchOptions([])
      return
    }

    setSearchLoading(true)
    try {
      const res = await api.get(`/user/search?keyword=${encodeURIComponent(query)}`)
      const { success, data } = res.data
      if (success && Array.isArray(data)) {
        const options: SearchOption[] = data.map((user: UserRow) => ({
          key: user.id.toString(),
          value: user.username,
          text: user.username,
          content: (
            <div className="flex flex-col">
              <div className="font-medium">{user.username}</div>
              <div className="text-sm text-muted-foreground">
                ID: {user.id} • Role: {user.role === 1 ? 'Admin' : user.role === 10 ? 'Super Admin' : 'User'} • Status: {user.status === 1 ? 'Enabled' : 'Disabled'}
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

  useEffect(() => {
    load(0)
  }, [sortBy, sortOrder])

  const search = async () => {
    setLoading(true)
    try {
      if (!searchKeyword.trim()) return load(0)
      let url = `/user/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data } = res.data
      if (success) {
        setData(data)
        setPageIndex(0)
      }
    } finally {
      setLoading(false)
    }
  }

  const columns: ColumnDef<UserRow>[] = [
    { header: 'ID', accessorKey: 'id' },
    { header: 'Username', accessorKey: 'username' },
    { header: 'Display Name', accessorKey: 'display_name' },
    { header: 'Role', cell: ({ row }) => (row.original.role >= 10 ? 'Admin' : 'Normal') },
    { header: 'Status', cell: ({ row }) => (row.original.status === 1 ? 'Enabled' : 'Disabled') },
    { header: 'Group', accessorKey: 'group' },
    { header: 'Quota', accessorKey: 'quota' },
    { header: 'Used', accessorKey: 'used_quota' },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => navigate(`/users/edit/${row.original.id}`)}
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
          <Button variant="outline" size="sm" onClick={() => setOpenTopup({open:true, userId: row.original.id, username: row.original.username})}>Top Up</Button>
        </div>
      ),
    },
  ]

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete', idx: number) => {
    let res
    if (action === 'delete') {
      res = await api.delete(`/user/${id}`)
    } else {
      const body: any = { id, status: action === 'enable' ? 1 : 2 }
      res = await api.put('/user/?status_only=true', body)
    }
    const { success } = res.data
    if (success) {
      // Optimistic update like legacy
      const next = [...data]
      if (action === 'delete') {
        next.splice(idx, 1)
      } else {
        next[idx].status = action === 'enable' ? 1 : 2
      }
      setData(next)
    }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Users</CardTitle>
              <CardDescription>Manage users</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button onClick={() => navigate('/users/add')}>Add User</Button>
              <select className="h-9 border rounded-md px-2 text-sm" value={sortBy} onChange={(e) => { setSortBy(e.target.value); setSortOrder('desc') }}>
                <option value="">Default</option>
                <option value="quota">Remaining Quota</option>
                <option value="used_quota">Used Quota</option>
                <option value="username">Username</option>
                <option value="id">ID</option>
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
                placeholder="Search users..."
                searchPlaceholder="Search by username..."
                options={searchOptions}
                onSearchChange={searchUsers}
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
                noResultsMessage="No users found"
                additionLabel="Use username: "
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
              setPageIndex(0)
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

      {/* Create User Dialog */}
      <CreateUserDialog open={openCreate} onOpenChange={setOpenCreate} onCreated={() => load(pageIndex)} />
      {/* Top Up Dialog */}
      <TopUpDialog open={openTopup.open} onOpenChange={(v)=>setOpenTopup({open:v})} userId={openTopup.userId} username={openTopup.username} onDone={()=>load(pageIndex)} />
    </div>
  )
}

// Create User Dialog
function CreateUserDialog({ open, onOpenChange, onCreated }: { open: boolean, onOpenChange: (v:boolean)=>void, onCreated: ()=>void }) {
  const schema = z.object({
    username: z.string().min(1),
    password: z.string().min(6),
    display_name: z.string().optional(),
  })
  type FormT = z.infer<typeof schema>
  const form = useForm<FormT>({ resolver: zodResolver(schema), defaultValues: { username: '', password: '', display_name: '' } })
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader><DialogTitle>Create User</DialogTitle></DialogHeader>
        <Form {...form}>
          <form className="space-y-3" onSubmit={form.handleSubmit(async (values) => {
            const res = await api.post('/user/', { username: values.username, password: values.password, display_name: values.display_name || values.username })
            if (res.data?.success) {
              onOpenChange(false)
              form.reset()
              onCreated()
            }
          })}>
            <FormField control={form.control} name="username" render={({ field }) => (
              <FormItem>
                <FormLabel>Username</FormLabel>
                <FormControl><Input {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="password" render={({ field }) => (
              <FormItem>
                <FormLabel>Password</FormLabel>
                <FormControl><Input type="password" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="display_name" render={({ field }) => (
              <FormItem>
                <FormLabel>Display Name</FormLabel>
                <FormControl><Input {...field} /></FormControl>
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

// Top Up Dialog
function TopUpDialog({ open, onOpenChange, userId, username, onDone }: { open: boolean, onOpenChange: (v:boolean)=>void, userId?: number, username?: string, onDone: ()=>void }) {
  const schema = z.object({ quota: z.coerce.number().int(), remark: z.string().optional() })
  type FormT = z.infer<typeof schema>
  const form = useForm<FormT>({ resolver: zodResolver(schema), defaultValues: { quota: 0, remark: '' } })
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader><DialogTitle>Top Up {username ? `@${username}` : ''}</DialogTitle></DialogHeader>
        <Form {...form}>
          <form className="space-y-3" onSubmit={form.handleSubmit(async (values) => {
            if (!userId) return
            const res = await api.post('/topup', { user_id: userId, quota: values.quota, remark: values.remark })
            if (res.data?.success) {
              onOpenChange(false)
              form.reset()
              onDone()
            }
          })}>
            <FormField control={form.control} name="quota" render={({ field }) => (
              <FormItem>
                <FormLabel>Quota</FormLabel>
                <FormControl><Input type="number" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name="remark" render={({ field }) => (
              <FormItem>
                <FormLabel>Remark</FormLabel>
                <FormControl><Input placeholder="Optional" {...field} /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <div className="pt-2 flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
              <Button type="submit">Submit</Button>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

async function onManageRole(username: string, action: 'promote'|'demote') {
  await api.post('/user/manage', { username, action })
}

async function onDisableTotp(id: number) {
  await api.post(`/user/totp/disable/${id}`)
}
