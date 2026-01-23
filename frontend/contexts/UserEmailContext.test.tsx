import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { UserEmailProvider, useUserEmail } from './UserEmailContext'

const TestConsumer = () => {
  const { email, setEmail, clearEmail, isLoading, storageError } = useUserEmail()
  return (
    <div>
      <span data-testid="email">{email ?? 'no-email'}</span>
      <span data-testid="loading">{isLoading ? 'loading' : 'ready'}</span>
      <span data-testid="storage-error">{storageError ? 'error' : 'ok'}</span>
      <button onClick={() => setEmail('new@example.com')}>Set Email</button>
      <button onClick={clearEmail}>Clear Email</button>
    </div>
  )
}

describe('UserEmailContext', () => {
  const localStorageMock = (() => {
    let store: Record<string, string> = {}
    return {
      getItem: vi.fn((key: string) => store[key] ?? null),
      setItem: vi.fn((key: string, value: string) => { store[key] = value }),
      removeItem: vi.fn((key: string) => { delete store[key] }),
      clear: vi.fn(() => { store = {} }),
    }
  })()

  beforeEach(() => {
    Object.defineProperty(window, 'localStorage', { value: localStorageMock })
    localStorageMock.clear()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should return null when no email in localStorage', async () => {
    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    expect(screen.getByTestId('email')).toHaveTextContent('no-email')
  })

  it('should load email from localStorage on mount', async () => {
    localStorageMock.getItem.mockReturnValue('stored@example.com')

    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    expect(screen.getByTestId('email')).toHaveTextContent('stored@example.com')
  })

  it('should persist email to localStorage when setEmail is called', async () => {
    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    fireEvent.click(screen.getByText('Set Email'))

    expect(localStorageMock.setItem).toHaveBeenCalledWith('asken_user_email', 'new@example.com')
    expect(screen.getByTestId('email')).toHaveTextContent('new@example.com')
  })

  it('should remove email from localStorage when clearEmail is called', async () => {
    localStorageMock.getItem.mockReturnValue('existing@example.com')

    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    fireEvent.click(screen.getByText('Clear Email'))

    expect(localStorageMock.removeItem).toHaveBeenCalledWith('asken_user_email')
    expect(screen.getByTestId('email')).toHaveTextContent('no-email')
  })

  it('should throw error when useUserEmail is called outside provider', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => render(<TestConsumer />)).toThrow(
      'useUserEmail must be used within a UserEmailProvider'
    )

    consoleSpy.mockRestore()
  })

  it('should handle localStorage errors gracefully on read and set storageError', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    localStorageMock.getItem.mockImplementation(() => {
      throw new Error('localStorage error')
    })

    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    expect(screen.getByTestId('email')).toHaveTextContent('no-email')
    expect(screen.getByTestId('storage-error')).toHaveTextContent('error')
    expect(consoleSpy).toHaveBeenCalledWith('Failed to access localStorage:', expect.any(Error))

    consoleSpy.mockRestore()
  })

  it('should not set storageError when localStorage read succeeds', async () => {
    localStorageMock.getItem.mockReturnValue('stored@example.com')

    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    expect(screen.getByTestId('storage-error')).toHaveTextContent('ok')
  })

  it('should not update state when localStorage write fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    // 初期状態ではメールがないことを明示
    localStorageMock.getItem.mockReturnValue(null)
    localStorageMock.setItem.mockImplementation(() => {
      throw new Error('localStorage error')
    })

    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    fireEvent.click(screen.getByText('Set Email'))

    // State should NOT be updated when localStorage fails
    expect(screen.getByTestId('email')).toHaveTextContent('no-email')
    expect(consoleSpy).toHaveBeenCalledWith('Failed to save email to localStorage:', expect.any(Error))

    consoleSpy.mockRestore()
  })

  it('should not clear state when localStorage remove fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    localStorageMock.getItem.mockReturnValue('existing@example.com')
    localStorageMock.removeItem.mockImplementation(() => {
      throw new Error('localStorage error')
    })

    render(
      <UserEmailProvider>
        <TestConsumer />
      </UserEmailProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('ready')
    })

    expect(screen.getByTestId('email')).toHaveTextContent('existing@example.com')

    fireEvent.click(screen.getByText('Clear Email'))

    // State should NOT be cleared when localStorage fails
    expect(screen.getByTestId('email')).toHaveTextContent('existing@example.com')
    expect(consoleSpy).toHaveBeenCalledWith('Failed to remove email from localStorage:', expect.any(Error))

    consoleSpy.mockRestore()
  })
})
