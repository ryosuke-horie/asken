import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import SkippedMealButton from './SkippedMealButton'
import { useAuth } from '@/contexts/AuthContext'

vi.mock('@/contexts/AuthContext', () => ({
  useAuth: vi.fn(),
}))

describe('SkippedMealButton', () => {
  const mockOnComplete = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useAuth).mockReturnValue({
      token: 'test-token',
      user: null,
      login: vi.fn(),
      logout: vi.fn(),
      isLoading: false,
    })
  })

  it('「食べませんでした」ボタンを表示すべき', () => {
    render(<SkippedMealButton mealType="lunch" mealDate="2026-01-24" />)

    expect(screen.getByText('食べませんでした')).toBeInTheDocument()
    expect(screen.getByText('🚫')).toBeInTheDocument()
    expect(
      screen.getByText('この食事を「食べなかった」として記録します'),
    ).toBeInTheDocument()
  })

  it('ボタンクリック時にAPIを呼び出すべき', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ id: 'test-id' }),
    })
    global.fetch = mockFetch

    render(
      <SkippedMealButton
        mealType="lunch"
        mealDate="2026-01-24"
        onComplete={mockOnComplete}
      />,
    )

    fireEvent.click(screen.getByText('食べませんでした'))

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith(
        '/api/meals/skip',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
            Authorization: 'Bearer test-token',
          }),
          body: JSON.stringify({
            meal_type: 'lunch',
            meal_date: '2026-01-24',
          }),
        }),
      )
    })

    await waitFor(() => {
      expect(mockOnComplete).toHaveBeenCalled()
    })
  })

  it('送信中はボタンを無効化すべき', async () => {
    let resolvePromise: (value: unknown) => void
    const pendingPromise = new Promise((resolve) => {
      resolvePromise = resolve
    })
    global.fetch = vi.fn().mockReturnValue(pendingPromise)

    render(<SkippedMealButton mealType="lunch" mealDate="2026-01-24" />)

    fireEvent.click(screen.getByText('食べませんでした'))

    await waitFor(() => {
      expect(screen.getByText('記録中...')).toBeInTheDocument()
      expect(screen.getByRole('button')).toBeDisabled()
    })

    resolvePromise!({ ok: true, json: () => Promise.resolve({ id: 'test-id' }) })
  })

  it('API失敗時にエラーメッセージを表示すべき', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      text: () => Promise.resolve('サーバーエラー'),
    })

    render(<SkippedMealButton mealType="lunch" mealDate="2026-01-24" />)

    fireEvent.click(screen.getByText('食べませんでした'))

    await waitFor(() => {
      expect(screen.getByText('サーバーエラー')).toBeInTheDocument()
    })
  })

  it('トークンがない場合はAPIを呼び出さないべき', async () => {
    vi.mocked(useAuth).mockReturnValue({
      token: null,
      user: null,
      login: vi.fn(),
      logout: vi.fn(),
      isLoading: false,
    })
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    render(<SkippedMealButton mealType="lunch" mealDate="2026-01-24" />)

    fireEvent.click(screen.getByText('食べませんでした'))

    await waitFor(() => {
      expect(mockFetch).not.toHaveBeenCalled()
    })
  })
})
