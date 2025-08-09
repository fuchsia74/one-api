import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Date/time utility functions
export function formatTimestamp(timestamp: number): string {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString()
}

export function toDateTimeLocal(timestamp: number | undefined): string {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toISOString().slice(0, 16)
}

export function fromDateTimeLocal(dateTimeLocal: string): number {
  if (!dateTimeLocal) return 0
  return Math.floor(new Date(dateTimeLocal).getTime() / 1000)
}

// Number formatting
export function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

// Quota formatting
export function formatQuota(quota: number): string {
  return `$${(quota / 500000).toFixed(4)}`
}

// Render quota with proper formatting
export function renderQuota(quota: number): string {
  return formatQuota(quota)
}
