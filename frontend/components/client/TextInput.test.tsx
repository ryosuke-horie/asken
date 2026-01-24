import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import TextInput from './TextInput'

const mockFetch = vi.fn()

vi.mock('@/contexts/AuthContext', () => ({
  useAuth: vi.fn(() => ({
    user: { id: 'user-123', email: 'test@example.com', name: 'Test User' },
    token: 'mock-jwt-token',
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
  })),
}))

describe('TextInput', () => {
  const defaultProps = {
    mealType: 'lunch' as const,
    mealDate: '2024-01-15',
  }

  beforeEach(() => {
    mockFetch.mockReset()
    global.fetch = mockFetch
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('初期表示', () => {
    it('入力フィールドが1つ表示されるべき', () => {
      render(<TextInput {...defaultProps} />)

      const inputs = screen.getAllByPlaceholderText('例: ご飯二杯、焼肉')
      expect(inputs).toHaveLength(1)
    })

    it('追加ボタンと解析ボタンが表示されるべき', () => {
      render(<TextInput {...defaultProps} />)

      expect(screen.getByText('追加')).toBeInTheDocument()
      expect(screen.getByText('解析')).toBeInTheDocument()
    })

    it('入力がない場合は解析ボタンが無効になるべき', () => {
      render(<TextInput {...defaultProps} />)

      const submitButton = screen.getByText('解析')
      expect(submitButton).toBeDisabled()
    })
  })

  describe('入力フィールドの操作', () => {
    it('テキストを入力すると解析ボタンが有効になるべき', () => {
      render(<TextInput {...defaultProps} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯' } })

      const submitButton = screen.getByText('解析')
      expect(submitButton).not.toBeDisabled()
    })

    it('追加ボタンをクリックすると入力フィールドが増えるべき', () => {
      render(<TextInput {...defaultProps} />)

      fireEvent.click(screen.getByText('追加'))

      const inputs = screen.getAllByPlaceholderText('例: ご飯二杯、焼肉')
      expect(inputs).toHaveLength(2)
    })

    it('削除ボタンをクリックすると入力フィールドが減るべき', () => {
      render(<TextInput {...defaultProps} />)

      fireEvent.click(screen.getByText('追加'))
      expect(screen.getAllByPlaceholderText('例: ご飯二杯、焼肉')).toHaveLength(2)

      const deleteButtons = screen.getAllByLabelText('削除')
      fireEvent.click(deleteButtons[0])

      expect(screen.getAllByPlaceholderText('例: ご飯二杯、焼肉')).toHaveLength(1)
    })

    it('入力フィールドが1つの場合は削除ボタンが無効になるべき', () => {
      render(<TextInput {...defaultProps} />)

      const deleteButton = screen.getByLabelText('削除')
      expect(deleteButton).toBeDisabled()
    })
  })

  describe('送信処理', () => {
    it('空の入力で送信するとエラーメッセージが表示されるべき', () => {
      render(<TextInput {...defaultProps} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: '   ' } })

      const submitButton = screen.getByText('解析')
      expect(submitButton).toBeDisabled()
    })

    it('送信成功時にAPIを呼び出すべき', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-id-123' }),
      })
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ status: 'completed' }),
      })

      const onComplete = vi.fn()
      render(<TextInput {...defaultProps} onComplete={onComplete} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯二杯' } })

      fireEvent.click(screen.getByText('解析'))

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/analyze'),
          expect.objectContaining({
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: 'Bearer mock-jwt-token',
            },
            body: JSON.stringify({
              input_text: 'ご飯二杯',
              meal_type: 'lunch',
              meal_date: '2024-01-15',
            }),
          }),
        )
      })
    })

    it('複数入力がカンマ区切りで送信されるべき', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-id-123' }),
      })

      render(<TextInput {...defaultProps} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯' } })

      fireEvent.click(screen.getByText('追加'))
      const inputs = screen.getAllByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(inputs[1], { target: { value: '味噌汁' } })

      fireEvent.click(screen.getByText('解析'))

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/analyze'),
          expect.objectContaining({
            body: JSON.stringify({
              input_text: 'ご飯, 味噌汁',
              meal_type: 'lunch',
              meal_date: '2024-01-15',
            }),
          }),
        )
      })
    })

    it('送信中は入力フィールドが無効になるべき', async () => {
      mockFetch.mockImplementation(() => new Promise(() => {}))

      render(<TextInput {...defaultProps} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯' } })

      fireEvent.click(screen.getByText('解析'))

      await waitFor(() => {
        expect(input).toBeDisabled()
      })
    })

    it('送信失敗時にエラーメッセージが表示されるべき', async () => {
      mockFetch.mockResolvedValue({
        ok: false,
        status: 500,
      })

      render(<TextInput {...defaultProps} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯' } })

      fireEvent.click(screen.getByText('解析'))

      await waitFor(() => {
        expect(screen.getByText(/送信に失敗しました/)).toBeInTheDocument()
      })
    })
  })

  describe('ステータスポーリング', () => {
    it('分析完了時にonCompleteが呼ばれるべき', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-id-123' }),
      })
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ status: 'completed' }),
      })

      const onComplete = vi.fn()
      render(<TextInput {...defaultProps} onComplete={onComplete} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯' } })

      fireEvent.click(screen.getByText('解析'))

      await waitFor(() => {
        expect(onComplete).toHaveBeenCalled()
      })
    })

    it('分析失敗時にエラーメッセージが表示されるべき', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-id-123' }),
      })
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ status: 'failed', error: 'Gemini API エラー' }),
      })

      render(<TextInput {...defaultProps} />)

      const input = screen.getByPlaceholderText('例: ご飯二杯、焼肉')
      fireEvent.change(input, { target: { value: 'ご飯' } })

      fireEvent.click(screen.getByText('解析'))

      await waitFor(() => {
        expect(screen.getByText('Gemini API エラー')).toBeInTheDocument()
      })
    })
  })
})
