import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import DeleteHistoryButton from './DeleteHistoryButton'

const mockPush = vi.fn()
const mockRefresh = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    refresh: mockRefresh,
  }),
}))

describe('DeleteHistoryButton', () => {
  beforeEach(() => {
    mockPush.mockClear()
    mockRefresh.mockClear()
    vi.stubGlobal('confirm', vi.fn(() => true))
    vi.stubGlobal('alert', vi.fn())
    vi.stubGlobal('fetch', vi.fn())
  })

  describe('通常モード', () => {
    it('削除ボタンを表示すべき', () => {
      render(<DeleteHistoryButton historyId="test-id" />)

      expect(screen.getByRole('button', { name: '削除' })).toBeInTheDocument()
    })

    it('削除成功時に/historyへリダイレクトすべき', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true }))

      render(<DeleteHistoryButton historyId="test-id" />)
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/history')
        expect(mockRefresh).toHaveBeenCalled()
      })
    })

    it('削除をキャンセルした場合はAPIを呼ばないべき', () => {
      vi.stubGlobal('confirm', vi.fn(() => false))

      render(<DeleteHistoryButton historyId="test-id" />)
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      expect(fetch).not.toHaveBeenCalled()
    })

    it('削除失敗時にアラートを表示すべき', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: false }))

      render(<DeleteHistoryButton historyId="test-id" />)
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      await waitFor(() => {
        expect(alert).toHaveBeenCalledWith('削除に失敗しました')
      })
    })

    it('削除中はボタンが無効化されるべき', async () => {
      vi.stubGlobal('fetch', vi.fn().mockImplementation(() => new Promise(() => {})))

      render(<DeleteHistoryButton historyId="test-id" />)
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      await waitFor(() => {
        expect(screen.getByRole('button', { name: '削除中...' })).toBeDisabled()
      })
    })
  })

  describe('アイコンモード', () => {
    it('アイコンボタンを表示すべき', () => {
      render(<DeleteHistoryButton historyId="test-id" iconOnly />)

      const button = screen.getByRole('button', { name: '削除' })
      expect(button).toBeInTheDocument()
      expect(button.querySelector('svg')).toBeInTheDocument()
    })

    it('onSuccessコールバックが渡された場合はそれを呼び出すべき', async () => {
      vi.stubGlobal('fetch', vi.fn().mockResolvedValue({ ok: true }))
      const onSuccess = vi.fn()

      render(<DeleteHistoryButton historyId="test-id" iconOnly onSuccess={onSuccess} />)
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      await waitFor(() => {
        expect(onSuccess).toHaveBeenCalled()
        expect(mockPush).not.toHaveBeenCalled()
      })
    })

    it('削除中はスピナーを表示すべき', async () => {
      vi.stubGlobal('fetch', vi.fn().mockImplementation(() => new Promise(() => {})))

      render(<DeleteHistoryButton historyId="test-id" iconOnly />)
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      await waitFor(() => {
        const button = screen.getByRole('button', { name: '削除' })
        expect(button.querySelector('svg')).not.toBeInTheDocument()
      })
    })
  })

  describe('イベント伝播', () => {
    it('クリックイベントの伝播を止めるべき', () => {
      const parentClickHandler = vi.fn()
      vi.stubGlobal('confirm', vi.fn(() => false))

      render(
        <div onClick={parentClickHandler}>
          <DeleteHistoryButton historyId="test-id" />
        </div>
      )
      fireEvent.click(screen.getByRole('button', { name: '削除' }))

      expect(parentClickHandler).not.toHaveBeenCalled()
    })
  })
})
