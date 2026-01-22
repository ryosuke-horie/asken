import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import EmailSetup from './EmailSetup'

const mockSetEmail = vi.fn()
const mockClearEmail = vi.fn()

vi.mock('@/contexts/UserEmailContext', () => ({
  useUserEmail: vi.fn(() => ({
    email: null,
    setEmail: mockSetEmail,
    clearEmail: mockClearEmail,
    isLoading: false,
  })),
}))

import { useUserEmail } from '@/contexts/UserEmailContext'

describe('EmailSetup', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(useUserEmail).mockReturnValue({
      email: null,
      setEmail: mockSetEmail,
      clearEmail: mockClearEmail,
      isLoading: false,
    })
  })

  it('ローディング中は読み込み中メッセージを表示すべき', () => {
    vi.mocked(useUserEmail).mockReturnValue({
      email: null,
      setEmail: mockSetEmail,
      clearEmail: mockClearEmail,
      isLoading: true,
    })

    render(<EmailSetup />)

    expect(screen.getByText('読み込み中...')).toBeInTheDocument()
  })

  it('メールアドレスが設定済みの場合は何も表示しないべき', () => {
    vi.mocked(useUserEmail).mockReturnValue({
      email: 'existing@example.com',
      setEmail: mockSetEmail,
      clearEmail: mockClearEmail,
      isLoading: false,
    })

    const { container } = render(<EmailSetup />)

    expect(container.firstChild).toBeNull()
  })

  it('メールアドレスが未設定の場合はモーダルを表示すべき', () => {
    render(<EmailSetup />)

    expect(screen.getByText('ようこそ')).toBeInTheDocument()
    expect(screen.getByText('食事記録を保存するためにメールアドレスを入力してください')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('example@email.com')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '始める' })).toBeInTheDocument()
  })

  it('空のメールアドレスでエラーを表示すべき', () => {
    render(<EmailSetup />)

    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    expect(screen.getByText('メールアドレスを入力してください')).toBeInTheDocument()
    expect(mockSetEmail).not.toHaveBeenCalled()
  })

  it('空白のみのメールアドレスでエラーを表示すべき', () => {
    render(<EmailSetup />)

    fireEvent.change(screen.getByPlaceholderText('example@email.com'), {
      target: { value: '   ' },
    })
    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    expect(screen.getByText('メールアドレスを入力してください')).toBeInTheDocument()
    expect(mockSetEmail).not.toHaveBeenCalled()
  })

  it('不正な形式のメールアドレスでエラーを表示すべき', () => {
    render(<EmailSetup />)

    // TLDがないメールアドレス（ブラウザ検証は通過するがJS検証で失敗）
    fireEvent.change(screen.getByPlaceholderText('example@email.com'), {
      target: { value: 'test@localhost' },
    })
    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    expect(screen.getByText('有効なメールアドレスを入力してください')).toBeInTheDocument()
    expect(mockSetEmail).not.toHaveBeenCalled()
  })

  it('TLDなしのメールアドレスでエラーを表示すべき', () => {
    render(<EmailSetup />)

    // ドットがないメールアドレス
    fireEvent.change(screen.getByPlaceholderText('example@email.com'), {
      target: { value: 'test@example' },
    })
    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    expect(screen.getByText('有効なメールアドレスを入力してください')).toBeInTheDocument()
    expect(mockSetEmail).not.toHaveBeenCalled()
  })

  it('有効なメールアドレスでsetEmailを呼び出すべき', () => {
    render(<EmailSetup />)

    fireEvent.change(screen.getByPlaceholderText('example@email.com'), {
      target: { value: 'valid@example.com' },
    })
    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    expect(mockSetEmail).toHaveBeenCalledWith('valid@example.com')
  })

  it('前後の空白をトリムしてsetEmailを呼び出すべき', () => {
    render(<EmailSetup />)

    fireEvent.change(screen.getByPlaceholderText('example@email.com'), {
      target: { value: '  valid@example.com  ' },
    })
    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    expect(mockSetEmail).toHaveBeenCalledWith('valid@example.com')
  })

  it('フォーム送信後にエラーをクリアすべき', () => {
    render(<EmailSetup />)

    // 最初に空のメールアドレスでエラーを発生させる
    fireEvent.click(screen.getByRole('button', { name: '始める' }))
    expect(screen.getByText('メールアドレスを入力してください')).toBeInTheDocument()

    // 有効なメールアドレスを入力して送信
    fireEvent.change(screen.getByPlaceholderText('example@email.com'), {
      target: { value: 'valid@example.com' },
    })
    fireEvent.click(screen.getByRole('button', { name: '始める' }))

    // エラーメッセージが消えていることを確認
    expect(screen.queryByText('メールアドレスを入力してください')).not.toBeInTheDocument()
    expect(mockSetEmail).toHaveBeenCalledWith('valid@example.com')
  })
})
