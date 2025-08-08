import { useEffect, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { DataTable } from '@/components/ui/data-table'
import api from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

interface Token {
  id: number
  name: string
  key: string
  remain_quota: number
  status: number
  created_time: number
}

export function TokensPage() {
  const [data, setData] = useState<Token[]>([])
  const [loading, setLoading] = useState(false)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageSize] = useState(20)
  const [total, setTotal] = useState(0)

  const load = async (p = 0) => {
    setLoading(true)
    try {
      const res = await api.get(`/token/?p=${p}`)
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

  const manage = async (id: number, action: 'enable' | 'disable' | 'delete') => {
    let res
    if (action === 'delete') res = await api.delete(`/token/${id}`)
    else res = await api.put('/token/?status_only=true', { id, status: action === 'enable' ? 1 : 2 })
    if (res.data?.success) load(pageIndex)
  }

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
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'enable')}>Enable</Button>
          <Button variant="outline" size="sm" onClick={() => manage(row.original.id, 'disable')}>Disable</Button>
          <Button variant="destructive" size="sm" onClick={() => manage(row.original.id, 'delete')}>Delete</Button>
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
            <Button onClick={() => load(pageIndex)} disabled={loading} variant="outline">Refresh</Button>
          </div>
        </CardHeader>
        <CardContent>
          <DataTable columns={columns} data={data} pageIndex={pageIndex} pageSize={pageSize} total={total} onPageChange={(pi) => load(pi)} />
        </CardContent>
      </Card>
    </div>
  )
}

export default TokensPage
