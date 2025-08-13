import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { ChannelsPage } from '../ChannelsPage'
import { api } from '@/lib/api'

// Mock the API
vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn(),
    delete: vi.fn(),
    put: vi.fn(),
  },
}))

// Mock the responsive hook
vi.mock('@/hooks/useResponsive', () => ({
  useResponsive: () => ({ isMobile: false, isTablet: false }),
}))

// Mock react-router-dom
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

const mockApiGet = vi.mocked(api.get)

const mockChannelsData = {
  success: true,
  data: Array.from({ length: 25 }, (_, i) => ({
    id: i + 1,
    name: `Channel ${i + 1}`,
    type: 1,
    status: 1,
    created_time: Date.now(),
    priority: 0,
    weight: 0,
    models: 'gpt-3.5-turbo',
    group: 'default',
    balance: 100,
    used_quota: 0,
  })),
  total: 25,
}

describe('ChannelsPage Pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockApiGet.mockResolvedValue({ data: mockChannelsData })
  })

  const renderChannelsPage = () => {
    return render(
      <BrowserRouter>
        <ChannelsPage />
      </BrowserRouter>
    )
  }

  it('should load initial data with default page size', async () => {
    renderChannelsPage()

    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledWith('/api/channel/?p=0&size=20&sort=id&order=desc')
    })

    // Allow for the initialization fix - may still have 2 calls during transition
    expect(mockApiGet).toHaveBeenCalledTimes(1)
  })

  it('should not make duplicate API calls when changing page size', async () => {
    renderChannelsPage()

    // Wait for initial load
    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledTimes(1)
    })

    // Clear the mock to track new calls
    mockApiGet.mockClear()

    // Find and click the page size selector
    const pageSizeSelect = screen.getByRole('combobox', { name: /rows per page/i })
    fireEvent.click(pageSizeSelect)

    // Select 10 rows per page
    const option10 = screen.getByRole('option', { name: '10' })
    fireEvent.click(option10)

    // Wait for the API call
    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledWith('/api/channel/?p=0&size=10&sort=id&order=desc')
    })

    // Should only make ONE API call, not multiple
    expect(mockApiGet).toHaveBeenCalledTimes(1)
  })

  it('should handle page navigation correctly', async () => {
    renderChannelsPage()

    // Wait for initial load
    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledTimes(1)
    })

    // Clear the mock to track new calls
    mockApiGet.mockClear()

    // Find and click page 2
    const page2Button = screen.getByRole('button', { name: '2' })
    fireEvent.click(page2Button)

    // Wait for the API call
    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledWith('/api/channel/?p=1&size=20&sort=id&order=desc')
    })

    expect(mockApiGet).toHaveBeenCalledTimes(1)
  })

  it('should handle sorting without duplicate calls', async () => {
    renderChannelsPage()

    // Wait for initial load
    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledTimes(1)
    })

    // Clear the mock to track new calls
    mockApiGet.mockClear()

    // Find and click a sortable column header
    const nameHeader = screen.getByRole('button', { name: /name/i })
    fireEvent.click(nameHeader)

    // Wait for the API call
    await waitFor(() => {
      expect(mockApiGet).toHaveBeenCalledWith('/api/channel/?p=0&size=20&sort=name&order=asc')
    })

    expect(mockApiGet).toHaveBeenCalledTimes(1)
  })
})
