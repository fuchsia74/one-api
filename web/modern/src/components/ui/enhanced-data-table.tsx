import * as React from 'react'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import type { ColumnDef, SortingState } from '@tanstack/react-table'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { AdvancedPagination } from '@/components/ui/advanced-pagination'
import { SearchableDropdown, type SearchOption } from '@/components/ui/searchable-dropdown'
import { Input } from '@/components/ui/input'
import { ArrowUpDown, ArrowUp, ArrowDown, Search, RotateCcw } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface EnhancedDataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  pageIndex?: number
  pageSize?: number
  total?: number
  onPageChange?: (pageIndex: number, pageSize: number) => void
  onPageSizeChange?: (pageSize: number) => void

  // Server-side sorting support
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  onSortChange?: (sortBy: string, sortOrder: 'asc' | 'desc') => void

  // Search functionality
  searchValue?: string
  searchOptions?: SearchOption[]
  searchLoading?: boolean
  onSearchChange?: (query: string) => void
  onSearchValueChange?: (value: string) => void
  onSearchSubmit?: () => void
  searchPlaceholder?: string
  allowSearchAdditions?: boolean

  // Toolbar actions
  toolbarActions?: React.ReactNode
  onRefresh?: () => void

  loading?: boolean
  className?: string
  emptyMessage?: string
}

export function EnhancedDataTable<TData, TValue>({
  columns,
  data,
  pageIndex = 0,
  pageSize = 20,
  total = 0,
  onPageChange,
  onPageSizeChange,
  sortBy = '',
  sortOrder = 'desc',
  onSortChange,
  searchValue = '',
  searchOptions = [],
  searchLoading = false,
  onSearchChange,
  onSearchValueChange,
  onSearchSubmit,
  searchPlaceholder = 'Search...',
  allowSearchAdditions = true,
  toolbarActions,
  onRefresh,
  loading = false,
  className,
  emptyMessage = 'No results found.',
}: EnhancedDataTableProps<TData, TValue>) {
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

  // Get sort icon based on current sort state
  const getSortIcon = (accessorKey: string) => {
    if (!onSortChange) return <ArrowUpDown className="ml-2 h-4 w-4 opacity-50" />

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

  const handleSearchAddition = (value: string) => {
    if (onSearchValueChange) {
      onSearchValueChange(value)
    }
  }

  return (
    <div className={cn('space-y-4', className)}>
      {/* Search and Actions Toolbar */}
      {(onSearchChange || toolbarActions || onRefresh) && (
        <div className="flex items-center justify-between gap-4">
          <div className="flex-1 flex items-center gap-2">
            {onSearchChange && (
              <>
                <div className="flex-1 max-w-md">
                  <SearchableDropdown
                    value={searchValue}
                    placeholder={searchPlaceholder}
                    searchPlaceholder={searchPlaceholder}
                    options={searchOptions}
                    onSearchChange={onSearchChange}
                    onChange={onSearchValueChange}
                    onAddItem={allowSearchAdditions ? handleSearchAddition : undefined}
                    loading={searchLoading}
                    noResultsMessage="No results found"
                    additionLabel="Search for: "
                    allowAdditions={allowSearchAdditions}
                    clearable={true}
                  />
                </div>
                {onSearchSubmit && (
                  <Button onClick={onSearchSubmit} disabled={loading} variant="outline">
                    <Search className="h-4 w-4 mr-2" />
                    Search
                  </Button>
                )}
              </>
            )}
          </div>

          <div className="flex items-center gap-2">
            {onRefresh && (
              <Button onClick={onRefresh} disabled={loading} variant="outline" size="sm">
                <RotateCcw className="h-4 w-4 mr-2" />
                Refresh
              </Button>
            )}
            {toolbarActions}
          </div>
        </div>
      )}

      {/* Data Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  Loading...
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
                  {emptyMessage}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Advanced Pagination */}
      <AdvancedPagination
        currentPage={pageIndex + 1}
        totalPages={Math.ceil(total / pageSize)}
        pageSize={pageSize}
        totalItems={total}
        onPageChange={(page) => onPageChange?.(page - 1, pageSize)}
        onPageSizeChange={(newPageSize) => {
          onPageSizeChange?.(newPageSize)
          // Reset to first page when changing page size
          onPageChange?.(0, newPageSize)
        }}
        loading={loading}
      />
    </div>
  )
}
