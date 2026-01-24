import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ConditionSection from './ConditionSection'

vi.mock('@/contexts/AuthContext', () => ({
  useAuth: vi.fn(() => ({ token: 'test-token' })),
}))

vi.mock('swr', () => ({
  default: vi.fn(),
}))

import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'

const mockUseSWR = useSWR as ReturnType<typeof vi.fn>
const mockUseAuth = useAuth as ReturnType<typeof vi.fn>

describe('ConditionSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuth.mockReturnValue({ token: 'test-token' })
  })

  describe('データ表示', () => {
    it('記録がある場合に体調と疲労度を表示すべき', () => {
      mockUseSWR.mockReturnValue({
        data: {
          id: 'test-id',
          condition: 3,
          fatigue: 2,
          recorded_at: '2024-01-15',
        },
        error: null,
        isLoading: false,
        mutate: vi.fn(),
      })

      render(<ConditionSection date="2024-01-15" />)

      expect(screen.getByText('体調・疲労度')).toBeInTheDocument()
      // 「体調」ラベルが複数あることを確認（サマリーとフォームの両方）
      const conditionLabels = screen.getAllByText('体調')
      expect(conditionLabels.length).toBeGreaterThanOrEqual(2)
    })

    it('記録がない場合はフォームのみ表示すべき', () => {
      mockUseSWR.mockReturnValue({
        data: null,
        error: null,
        isLoading: false,
        mutate: vi.fn(),
      })

      render(<ConditionSection date="2024-01-15" />)

      expect(screen.getByText('体調・疲労度')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '記録する' })).toBeInTheDocument()
    })

    it('ローディング中は読み込み中と表示すべき', () => {
      mockUseSWR.mockReturnValue({
        data: undefined,
        error: null,
        isLoading: true,
        mutate: vi.fn(),
      })

      render(<ConditionSection date="2024-01-15" />)

      expect(screen.getByText('読み込み中...')).toBeInTheDocument()
    })
  })

  describe('エラーハンドリング', () => {
    it('エラー時にエラーメッセージを表示すべき', () => {
      mockUseSWR.mockReturnValue({
        data: null,
        error: { message: 'データの取得に失敗しました' },
        isLoading: false,
        mutate: vi.fn(),
      })

      render(<ConditionSection date="2024-01-15" />)

      expect(screen.getByText('データの取得に失敗しました')).toBeInTheDocument()
    })

    it('エラーメッセージがない場合はデフォルトメッセージを表示すべき', () => {
      mockUseSWR.mockReturnValue({
        data: null,
        error: {},
        isLoading: false,
        mutate: vi.fn(),
      })

      render(<ConditionSection date="2024-01-15" />)

      expect(screen.getByText('データの取得に失敗しました')).toBeInTheDocument()
    })
  })

  describe('認証', () => {
    it('トークンがない場合はAPIを呼び出さないべき', () => {
      mockUseAuth.mockReturnValue({ token: null })
      mockUseSWR.mockReturnValue({
        data: null,
        error: null,
        isLoading: false,
        mutate: vi.fn(),
      })

      render(<ConditionSection date="2024-01-15" />)

      expect(mockUseSWR).toHaveBeenCalledWith(null, expect.any(Function))
    })
  })

  describe('記録送信', () => {
    it('フォーム送信時にAPIを呼び出すべき', async () => {
      const mockMutate = vi.fn()
      mockUseSWR.mockReturnValue({
        data: null,
        error: null,
        isLoading: false,
        mutate: mockMutate,
      })

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () =>
          Promise.resolve({
            id: 'new-id',
            condition: 3,
            fatigue: 1,
            recorded_at: '2024-01-15',
          }),
      })

      render(<ConditionSection date="2024-01-15" />)

      fireEvent.click(screen.getByText('良い'))
      fireEvent.click(screen.getByText('低い'))
      fireEvent.click(screen.getByRole('button', { name: '記録する' }))

      await waitFor(() => {
        expect(global.fetch).toHaveBeenCalledWith(
          expect.stringContaining('/api/condition-records'),
          expect.objectContaining({
            method: 'POST',
            headers: expect.objectContaining({
              'Content-Type': 'application/json',
              Authorization: 'Bearer test-token',
            }),
          }),
        )
      })

      await waitFor(() => {
        expect(mockMutate).toHaveBeenCalled()
      })
    })

    it('API失敗時にエラーを表示すべき', async () => {
      mockUseSWR.mockReturnValue({
        data: null,
        error: null,
        isLoading: false,
        mutate: vi.fn(),
      })

      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        text: () => Promise.resolve('記録に失敗しました'),
      })

      render(<ConditionSection date="2024-01-15" />)

      fireEvent.click(screen.getByText('良い'))
      fireEvent.click(screen.getByText('低い'))
      fireEvent.click(screen.getByRole('button', { name: '記録する' }))

      await waitFor(() => {
        expect(screen.getByText('記録に失敗しました')).toBeInTheDocument()
      })
    })
  })
})
