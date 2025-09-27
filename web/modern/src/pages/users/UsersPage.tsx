import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { EnhancedDataTable } from '@/components/ui/enhanced-data-table'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import { ResponsivePageContainer } from '@/components/ui/responsive-container'
import { useResponsive } from '@/hooks/useResponsive'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { useForm } from 'react-hook-form'
import * as z from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { renderQuota, cn } from '@/lib/utils'
import { ResponsiveActionGroup } from '@/components/ui/responsive-action-group'

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
  const { isMobile } = useResponsive()
  const [data, setData] = useState<UserRow[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [total, setTotal] = useState(0)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [searchOptions, setSearchOptions] = useState<SearchOption[]>([])
  const [searchLoading, setSearchLoading] = useState(false)
  const [sortBy, setSortBy] = useState('')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [openCreate, setOpenCreate] = useState(false)
  const [openTopup, setOpenTopup] = useState<{ open: boolean, userId?: number, username?: string }>({ open: false })

  const load = async (p = 0, size = pageSize) => {
    setLoading(true)
    try {
      // Unified API call - complete URL with /api prefix
      let url = `/api/user/?p=${p}&size=${size}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      const res = await api.get(url)
      const { success, data, total } = res.data
      if (success) {
        setData(data)
        setTotal(total || data.length)
        setPageIndex(p)
        setPageSize(size)
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
      // Unified API call - complete URL with /api prefix
      const res = await api.get(`/api/user/search?keyword=${encodeURIComponent(query)}`)
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
    if (searchKeyword.trim()) {
      search()
    } else {
      load(0, pageSize)
    }
  }, [sortBy, sortOrder])

  const search = async () => {
    setLoading(true)
    try {
      if (!searchKeyword.trim()) return load(0, pageSize)
      // Unified API call - complete URL with /api prefix
      let url = `/api/user/search?keyword=${encodeURIComponent(searchKeyword)}`
      if (sortBy) url += `&sort=${sortBy}&order=${sortOrder}`
      url += `&size=${pageSize}`
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
    {
      header: 'Used Quota',
      accessorKey: 'used_quota',
      cell: ({ row }) => (
        <span className="font-mono text-sm" title={`Used: ${renderQuota(row.original.used_quota || 0)}`}>
          {row.original.used_quota ? renderQuota(row.original.used_quota) : renderQuota(0)}
        </span>
      )
    },
    {
      header: 'Remaining Quota',
      accessorKey: 'quota',
      cell: ({ row }) => (
        <span className="font-mono text-sm" title={`Remaining: ${renderQuota(row.original.quota)}`}>
          {row.original.quota === -1 ? (
            <span className="text-green-600 font-semibold">Unlimited</span>
          ) : (
            renderQuota(row.original.quota)
          )}
        </span>
      )
    },
    {
      header: 'Actions',
      cell: ({ row }) => (
        <ResponsiveActionGroup justify="start">
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
          <Button variant="outline" size="sm" onClick={() => setOpenTopup({ open: true, userId: row.original.id, username: row.original.username })}>Top Up</Button>
        </ResponsiveActionGroup>
      ),
    },
  ]

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete', idx: number) => {
    let res: any
    if (action === 'delete') {
      // Unified API call - complete URL with /api prefix
      res = await api.delete(`/api/user/${id}`)
    } else {
      const body: any = { id, status: action === 'enable' ? 1 : 2 }
      res = await api.put('/api/user/?status_only=true', body)
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

  const toolbarActions = (
    <div className={cn(
      "flex gap-2",
      isMobile ? "flex-col w-full" : "items-center"
    )}>
      <Button
        onClick={() => navigate('/users/add')}
        className={cn(
          isMobile ? "w-full touch-target" : ""
        )}
      >
        Add User
      </Button>
      <select
        className={cn(
          "h-11 sm:h-9 border rounded-md px-3 py-2 text-base sm:text-sm",
          isMobile ? "w-full" : ""
        )}
        value={sortBy}
        onChange={(e) => { setSortBy(e.target.value); setSortOrder('desc') }}
      >
        <option value="">Default</option>
        <option value="quota">Remaining Quota</option>
        <option value="used_quota">Used Quota</option>
        <option value="username">Username</option>
        <option value="id">ID</option>
        <option value="created_time">Created Time</option>
      </select>
      <Button
        variant="outline"
        size="sm"
        onClick={() => setSortOrder(o => o === 'asc' ? 'desc' : 'asc')}
        className={cn(isMobile ? "w-full touch-target" : "")}
      >
        {sortOrder.toUpperCase()}
      </Button>
    </div>
  )

  // Handlers for page change and page size change
  const handlePageChange = (newPageIndex: number, newPageSize: number) => {
    load(newPageIndex, newPageSize)
  }

  const handlePageSizeChange = (newPageSize: number) => {
    setPageSize(newPageSize)
    setPageIndex(0)
    // Don't call load here - let onPageChange handle it to avoid duplicate API calls
  }

  return (
    <ResponsivePageContainer
      title="Users"
      description="Manage users"
      actions={toolbarActions}
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
            onSortChange={(newSortBy, newSortOrder) => {
              setSortBy(newSortBy)
              setSortOrder(newSortOrder)
              // Let useEffect handle the reload to avoid double requests
            }}
            searchValue={searchKeyword}
            searchOptions={searchOptions}
            searchLoading={searchLoading}
            onSearchChange={searchUsers}
            onSearchValueChange={setSearchKeyword}
            onSearchSubmit={search}
            searchPlaceholder="Search users by username..."
            allowSearchAdditions={true}
            onRefresh={() => load(pageIndex, pageSize)}
            loading={loading}
            emptyMessage="No users found. Add your first user to get started."
            mobileCardLayout={true}
            hideColumnsOnMobile={['created_time', 'accessed_time']}
            compactMode={isMobile}
          />
        </CardContent>
      </Card>

      {/* Create User Dialog */}
      <CreateUserDialog open={openCreate} onOpenChange={setOpenCreate} onCreated={() => load(pageIndex, pageSize)} />
      {/* Top Up Dialog */}
      <TopUpDialog open={openTopup.open} onOpenChange={(v) => setOpenTopup({ open: v })} userId={openTopup.userId} username={openTopup.username} onDone={() => load(pageIndex, pageSize)} />
    </ResponsivePageContainer>
  )
}

// Create User Dialog
function CreateUserDialog({ open, onOpenChange, onCreated }: { open: boolean, onOpenChange: (v: boolean) => void, onCreated: () => void }) {
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
            // Unified API call - complete URL with /api prefix
            const res = await api.post('/api/user/', { username: values.username, password: values.password, display_name: values.display_name || values.username })
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
function TopUpDialog({ open, onOpenChange, userId, username, onDone }: { open: boolean, onOpenChange: (v: boolean) => void, userId?: number, username?: string, onDone: () => void }) {
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
            // Unified API call - complete URL with /api prefix
            const res = await api.post('/api/topup', { user_id: userId, quota: values.quota, remark: values.remark })
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
