import { useEffect, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTable } from '@/components/ui/data-table'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

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
  const [data, setData] = useState<UserRow[]>([])
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
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'enable', row.index)}>Enable</Button>
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'disable', row.index)}>Disable</Button>
          <Button variant="destructive" size="sm" onClick={() => manage(row.original.id, 'delete', row.index)}>Delete</Button>
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
            <Input placeholder="Search users" value={searchKeyword} onChange={(e) => setSearchKeyword(e.target.value)} />
            <Button onClick={search} disabled={loading}>Search</Button>
          </div>
          <DataTable
            columns={columns}
            data={data}
            pageIndex={pageIndex}
            pageSize={pageSize}
            total={total}
            onPageChange={(pi) => load(pi)}
          />
        </CardContent>
      </Card>
    </div>
  )
}
