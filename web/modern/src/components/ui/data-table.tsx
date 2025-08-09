import * as React from 'react'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import type { ColumnDef, SortingState } from '@tanstack/react-table'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'

export interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  pageIndex?: number
  pageSize?: number
  total?: number
  onPageChange?: (pageIndex: number, pageSize: number) => void
  // Server-side sorting support
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  onSortChange?: (sortBy: string, sortOrder: 'asc' | 'desc') => void
  loading?: boolean
}

export function DataTable<TData, TValue>({
  columns,
  data,
  pageIndex = 0,
  pageSize = 20,
  total = 0,
  onPageChange,
  sortBy = '',
  sortOrder = 'desc',
  onSortChange,
  loading = false,
}: DataTableProps<TData, TValue>) {
  // Client-side sorting state (for display only when no server-side sorting)
  const [sorting, setSorting] = React.useState<SortingState>([])

  // Handle column header click for server-side sorting
  const handleSort = (accessorKey: string) => {
    if (!onSortChange) return

    // If clicking the same column, toggle order
    if (sortBy === accessorKey) {
      const newOrder = sortOrder === 'desc' ? 'asc' : 'desc'
      onSortChange(accessorKey, newOrder)
    } else {
      // New column, default to desc
      onSortChange(accessorKey, 'desc')
    }
  }

  const getSortIcon = (accessorKey: string) => {
    if (!onSortChange) return null

    if (sortBy === accessorKey) {
      return sortOrder === 'asc' ? (
        <ArrowUp className="ml-2 h-4 w-4" />
      ) : (
        <ArrowDown className="ml-2 h-4 w-4" />
      )
    }
    return <ArrowUpDown className="ml-2 h-4 w-4 opacity-50" />
  }

  // Enhanced columns with sorting support
  const enhancedColumns = columns.map((column) => {
    // Check if column has accessorKey for sorting
    const hasAccessorKey = 'accessorKey' in column && typeof column.accessorKey === 'string'
    const accessorKey = hasAccessorKey ? column.accessorKey as string : ''

    if (!accessorKey || !onSortChange) return column

    return {
      ...column,
      header: () => {
        const headerContent = typeof column.header === 'string' ? column.header : accessorKey

        return (
          <Button
            variant="ghost"
            onClick={() => handleSort(accessorKey)}
            className="h-auto p-0 font-semibold hover:bg-transparent"
          >
            <span>{headerContent}</span>
            {getSortIcon(accessorKey)}
          </Button>
        )
      },
    } as ColumnDef<TData, TValue>
  })

  const table = useReactTable({
    data,
    columns: enhancedColumns,
    state: {
      sorting,
    },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    manualSorting: !!onSortChange, // Use manual sorting if server-side sorting is available
    manualPagination: true,
    pageCount: Math.ceil(total / pageSize),
  })

  const canPrev = pageIndex > 0
  const canNext = (pageIndex + 1) * pageSize < total

  return (
    <div className="space-y-2">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id} className="text-left">
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  <div className="flex items-center justify-center space-x-2">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
                    <span>Loading...</span>
                  </div>
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => {
                    const headerDef = cell.column.columnDef.header
                    const label = typeof headerDef === 'string' ? headerDef : (cell.column.id || '')
                    return (
                      <TableCell key={cell.id} data-label={label}>
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </TableCell>
                    )
                  })}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-end space-x-2 py-2">
        <div className="text-sm text-muted-foreground mr-auto">
          {total ? `${pageIndex * pageSize + 1}-${Math.min((pageIndex + 1) * pageSize, total)} of ${total}` : '0 of 0'}
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange?.(pageIndex - 1, pageSize)}
          disabled={!canPrev || loading}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange?.(pageIndex + 1, pageSize)}
          disabled={!canNext || loading}
        >
          Next
        </Button>
      </div>
    </div>
  )
}
