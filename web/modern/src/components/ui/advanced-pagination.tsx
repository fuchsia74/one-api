import * as React from 'react'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, MoreHorizontal } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useResponsive } from '@/hooks/useResponsive'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface AdvancedPaginationProps {
  currentPage: number
  totalPages: number
  pageSize: number
  totalItems: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
  showPageSizeSelector?: boolean
  pageSizeOptions?: number[]
  className?: string
  loading?: boolean
}

export function AdvancedPagination({
  currentPage,
  totalPages,
  pageSize,
  totalItems,
  onPageChange,
  onPageSizeChange,
  showPageSizeSelector = true,
  pageSizeOptions = [10, 20, 50, 100],
  className,
  loading = false,
}: AdvancedPaginationProps) {
  const { isMobile, isTablet } = useResponsive()
  // Calculate page range to show - responsive version
  const getPageNumbers = () => {
    const pages: (number | 'ellipsis')[] = []
    const siblingCount = isMobile ? 0 : isTablet ? 1 : 1 // Fewer pages on mobile
    const maxPages = isMobile ? 3 : isTablet ? 5 : 7 // Responsive max pages

    if (totalPages <= maxPages) {
      // If total pages is small, show all pages
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i)
      }
    } else {
      // Always show first page
      pages.push(1)

      const leftBoundary = Math.max(2, currentPage - siblingCount)
      const rightBoundary = Math.min(totalPages - 1, currentPage + siblingCount)

      // Add ellipsis after first page if needed
      if (leftBoundary > 2) {
        pages.push('ellipsis')
      }

      // Add pages around current page
      for (let i = leftBoundary; i <= rightBoundary; i++) {
        pages.push(i)
      }

      // Add ellipsis before last page if needed
      if (rightBoundary < totalPages - 1) {
        pages.push('ellipsis')
      }

      // Always show last page (if more than 1 page)
      if (totalPages > 1) {
        pages.push(totalPages)
      }
    }

    return pages
  }

  const pageNumbers = getPageNumbers()
  const startItem = Math.min((currentPage - 1) * pageSize + 1, totalItems)
  const endItem = Math.min(currentPage * pageSize, totalItems)

  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages && page !== currentPage && !loading) {
      onPageChange(page)
    }
  }

  const handlePageSizeChange = (newPageSize: string) => {
    const size = parseInt(newPageSize)
    if (onPageSizeChange && size !== pageSize) {
      onPageSizeChange(size)
    }
  }

  if (totalPages <= 1 && !showPageSizeSelector) {
    return null
  }

  return (
    <div className={cn(
      "flex px-2 py-4",
      isMobile ? "flex-col gap-4" : "items-center justify-between",
      className
    )}>
      {/* Page info and size selector */}
      <div className={cn(
        "flex gap-4",
        isMobile ? "flex-col gap-2 items-center text-center" : "items-center"
      )}>
        <div className={cn(
          "text-muted-foreground",
          isMobile ? "text-xs order-1" : "text-sm"
        )}>
          Showing {startItem}-{endItem} of {totalItems} items
        </div>

        {showPageSizeSelector && onPageSizeChange && (
          <div className={cn(
            "flex items-center gap-2",
            isMobile ? "order-3 justify-center" : ""
          )}>
            <span
              aria-hidden="true"
              className={cn(
                "text-muted-foreground whitespace-nowrap",
                isMobile ? "text-xs" : "text-sm"
              )}
            >
              {isMobile ? "Per page:" : "Rows per page:"}
            </span>
            <Select
              value={pageSize.toString()}
              onValueChange={handlePageSizeChange}
              disabled={loading}
              aria-label="Rows per page"
            >
              <SelectTrigger
                aria-label="Rows per page"
                className={cn(
                isMobile ? "h-8 w-14" : "h-8 w-16"
                )}
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {pageSizeOptions.map((size) => (
                  <SelectItem key={size} value={size.toString()} aria-label={`${size}`}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}
      </div>

      {/* Pagination controls */}
      {totalPages > 1 && (
        <div className={cn(
          "flex items-center gap-1",
          isMobile ? "order-2 justify-center" : ""
        )}>
          {/* First page - Hide on mobile if too many controls */}
          {!isMobile && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => handlePageChange(1)}
              disabled={currentPage === 1 || loading}
              className="h-8 w-8 p-0 touch-target"
            >
              <ChevronsLeft className="h-4 w-4" />
              <span className="sr-only">First page</span>
            </Button>
          )}

          {/* Previous page */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(currentPage - 1)}
            disabled={currentPage === 1 || loading}
            className={cn(
              "h-8 p-0 touch-target",
              isMobile ? "w-10" : "w-8"
            )}
          >
            <ChevronLeft className="h-4 w-4" />
            <span className="sr-only">Previous page</span>
          </Button>

          {/* Page numbers */}
          {pageNumbers.map((page, index) => (
            <React.Fragment key={index}>
              {page === 'ellipsis' ? (
                <span className={cn(
                  "flex items-center justify-center",
                  isMobile ? "h-8 w-6" : "h-8 w-8"
                )}>
                  <MoreHorizontal className="h-4 w-4" />
                </span>
              ) : (
                <Button
                  variant={page === currentPage ? "default" : "outline"}
                  size="sm"
                  onClick={() => handlePageChange(page)}
                  disabled={loading}
                  aria-label={`Page ${page}`}
                  className={cn(
                    "h-8 p-0 touch-target",
                    isMobile ? "w-10 text-sm" : "w-8"
                  )}
                >
                  {page}
                </Button>
              )}
            </React.Fragment>
          ))}

          {/* Next page */}
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(currentPage + 1)}
            disabled={currentPage === totalPages || loading}
            className={cn(
              "h-8 p-0 touch-target",
              isMobile ? "w-10" : "w-8"
            )}
          >
            <ChevronRight className="h-4 w-4" />
            <span className="sr-only">Next page</span>
          </Button>

          {/* Last page - Hide on mobile if too many controls */}
          {!isMobile && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => handlePageChange(totalPages)}
              disabled={currentPage === totalPages || loading}
              className="h-8 w-8 p-0 touch-target"
            >
              <ChevronsRight className="h-4 w-4" />
              <span className="sr-only">Last page</span>
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
