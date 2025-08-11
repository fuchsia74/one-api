import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { vi } from 'vitest'
import { LoginPage } from '../LoginPage.impl'
import { useAuthStore } from '@/lib/stores/auth'
import { api } from '@/lib/api'

// Mock the auth store
vi.mock('@/lib/stores/auth')
vi.mock('@/lib/api')

const mockLogin = vi.fn()
const mockUseAuthStore = useAuthStore as any

// Mock localStorage
const mockLocalStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
}
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

// Mock window.history
Object.defineProperty(window, 'history', {
  value: { replaceState: vi.fn() },
})

const renderLoginPage = () => {
  return render(
    <BrowserRouter>
      <LoginPage />
    </BrowserRouter>
  )
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuthStore.mockReturnValue({
      login: mockLogin,
    })
    mockLocalStorage.getItem.mockReturnValue(JSON.stringify({
      system_name: 'Test API',
      github_oauth: false,
    }))
  })

  it('renders login form correctly', () => {
    renderLoginPage()

    expect(screen.getByDisplayValue('')).toBeInTheDocument() // username input
    expect(screen.getAllByDisplayValue('')).toHaveLength(2) // username and password inputs
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('shows TOTP input when TOTP is required', async () => {
    const mockApiPost = vi.mocked(api.post)
    mockApiPost.mockResolvedValueOnce({
      data: {
        success: false,
        message: 'totp_required',
        data: { totp_required: true }
      }
    })

    renderLoginPage()

    const usernameInput = screen.getByRole('textbox')
    const passwordInput = screen.getByDisplayValue('')

    // Fill in username and password
    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    // Wait for TOTP input to appear
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/6-digit totp code/i)).toBeInTheDocument()
    })

    // Check that username and password fields are disabled
    expect(usernameInput).toBeDisabled()
    expect(passwordInput).toBeDisabled()

    // Check that the button text changed
    expect(screen.getByRole('button', { name: /verify totp/i })).toBeInTheDocument()
  })

  it('disables TOTP verify button when code is incomplete', async () => {
    const mockApiPost = vi.mocked(api.post)
    mockApiPost.mockResolvedValueOnce({
      data: {
        success: false,
        message: 'totp_required',
        data: { totp_required: true }
      }
    })

    renderLoginPage()

    const usernameInput = screen.getByRole('textbox')
    const passwordInput = screen.getByDisplayValue('')

    // Fill in username and password and trigger TOTP
    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByPlaceholderText(/6-digit totp code/i)).toBeInTheDocument()
    })

    const totpInput = screen.getByPlaceholderText(/6-digit totp code/i)
    const verifyButton = screen.getByRole('button', { name: /verify totp/i })

    // Button should be disabled initially
    expect(verifyButton).toBeDisabled()

    // Enter incomplete TOTP code
    fireEvent.change(totpInput, { target: { value: '12345' } })
    expect(verifyButton).toBeDisabled()

    // Enter complete TOTP code
    fireEvent.change(totpInput, { target: { value: '123456' } })
    expect(verifyButton).not.toBeDisabled()
  })

  it('successfully logs in with valid TOTP code', async () => {
    const mockApiPost = vi.mocked(api.post)

    // First call - TOTP required
    mockApiPost.mockResolvedValueOnce({
      data: {
        success: false,
        message: 'totp_required',
        data: { totp_required: true }
      }
    })

    // Second call - successful login
    mockApiPost.mockResolvedValueOnce({
      data: {
        success: true,
        data: { id: 1, username: 'testuser', role: 1 }
      }
    })

    renderLoginPage()

    const usernameInput = screen.getByRole('textbox')
    const passwordInput = screen.getByDisplayValue('')

    // Initial login attempt
    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    // Wait for TOTP input
    await waitFor(() => {
      expect(screen.getByPlaceholderText(/6-digit totp code/i)).toBeInTheDocument()
    })

    // Enter TOTP code and submit
    fireEvent.change(screen.getByPlaceholderText(/6-digit totp code/i), { target: { value: '123456' } })
    fireEvent.click(screen.getByRole('button', { name: /verify totp/i }))

    // Verify login was called
    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({ id: 1, username: 'testuser', role: 1 }, '')
    })
  })

  it('shows back to login button in TOTP mode', async () => {
    const mockApiPost = vi.mocked(api.post)
    mockApiPost.mockResolvedValueOnce({
      data: {
        success: false,
        message: 'totp_required',
        data: { totp_required: true }
      }
    })

    renderLoginPage()

    const usernameInput = screen.getByRole('textbox')
    const passwordInput = screen.getByDisplayValue('')

    // Trigger TOTP mode
    fireEvent.change(usernameInput, { target: { value: 'testuser' } })
    fireEvent.change(passwordInput, { target: { value: 'password123' } })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /back to login/i })).toBeInTheDocument()
    })

    // Click back to login
    fireEvent.click(screen.getByRole('button', { name: /back to login/i }))

    // Should return to normal login mode
    expect(screen.queryByPlaceholderText(/6-digit totp code/i)).not.toBeInTheDocument()
    expect(usernameInput).not.toBeDisabled()
    expect(passwordInput).not.toBeDisabled()
  })
})
