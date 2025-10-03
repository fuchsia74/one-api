import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import api from "./api"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Date/time utility functions
export function formatTimestamp(timestamp: number): string {
  if (timestamp === undefined || timestamp === null) return '-'
  if (timestamp <= 0) return '-'
  const date = new Date(timestamp * 1000) // backend stores UTC seconds
  const pad = (n: number) => n.toString().padStart(2, '0')
  const yyyy = date.getFullYear()
  const mm = pad(date.getMonth() + 1)
  const dd = pad(date.getDate())
  const HH = pad(date.getHours())
  const MM = pad(date.getMinutes())
  const SS = pad(date.getSeconds())
  return `${yyyy}-${mm}-${dd} ${HH}:${MM}:${SS}` // local browser timezone
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

// Quota formatting with USD conversion support
export function formatQuota(quota: number): string {
  const displayInCurrency = localStorage.getItem('display_in_currency') === 'true'
  const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000')

  if (displayInCurrency) {
    const amount = (quota / quotaPerUnit).toFixed(4)
    return `$${amount}`
  }

  // Return formatted tokens
  return formatNumber(quota)
}

// Render quota with proper formatting
export function renderQuota(quota: number): string {
  return formatQuota(quota)
}

// Render quota with prompting information
export function renderQuotaWithPrompt(quota: number): string {
  const displayInCurrency = localStorage.getItem('display_in_currency') === 'true'
  const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000')

  if (displayInCurrency) {
    const amount = (quota / quotaPerUnit).toFixed(4)
    return `$${amount}`
  }

  return `${formatNumber(quota)} tokens`
}

// System status utility function
export interface SystemStatus {
  system_name?: string
  logo?: string
  footer_html?: string
  quota_per_unit?: string
  turnstile_check?: boolean
  turnstile_site_key?: string
  github_oauth?: boolean
  github_client_id?: string
  [key: string]: any
}

export const loadSystemStatus = async (): Promise<SystemStatus | null> => {
  // First try to get from localStorage
  const status = localStorage.getItem('status')
  if (status) {
    try {
      const parsedStatus = JSON.parse(status)
      return parsedStatus
    } catch (error) {
      console.error('Error parsing system status:', error)
    }
  }

  // If not in localStorage, fetch from server
  try {
    const response = await api.get('/api/status')
    const { success, data } = response.data

    if (success && data) {
      // Store in localStorage for future use (excluding usd_to_idr for real-time API usage)
      localStorage.setItem('status', JSON.stringify(data))
      localStorage.setItem('system_name', data.system_name || 'One API')
      localStorage.setItem('logo', data.logo || '')
      localStorage.setItem('footer_html', data.footer_html || '')
      localStorage.setItem('quota_per_unit', data.quota_per_unit || '500000')

      return data
    }
  } catch (error) {
    console.error('Error fetching system status:', error)
  }

  return null
}
