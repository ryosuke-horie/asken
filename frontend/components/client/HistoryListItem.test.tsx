import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import HistoryListItem from './HistoryListItem'
import { HistoryItem } from '@/types/nutrition'

const mockRefresh = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    refresh: mockRefresh,
  }),
}))

const mockHistoryItem: HistoryItem = {
  id: 'test-id-123',
  image_path: '/uploads/test-image.jpg',
  total_calories: 500.5,
  total_protein: 25.3,
  total_fat: 15.2,
  total_carbohydrates: 60.8,
  created_at: '2026-01-21T12:30:00Z',
}

describe('HistoryListItem', () => {
  beforeEach(() => {
    mockRefresh.mockClear()
  })

  describe('表示', () => {
    it('日付を日本語形式で表示すべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      expect(screen.getByText(/2026\/01\/21/)).toBeInTheDocument()
    })

    it('カロリーを整数で表示すべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      expect(screen.getByText('501 kcal')).toBeInTheDocument()
    })

    it('タンパク質を小数点1桁で表示すべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      expect(screen.getByText('P: 25.3g')).toBeInTheDocument()
    })

    it('脂質を小数点1桁で表示すべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      expect(screen.getByText('F: 15.2g')).toBeInTheDocument()
    })

    it('炭水化物を小数点1桁で表示すべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      expect(screen.getByText('C: 60.8g')).toBeInTheDocument()
    })

    it('削除ボタンを表示すべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      expect(screen.getByRole('button', { name: '削除' })).toBeInTheDocument()
    })
  })

  describe('リンク', () => {
    it('詳細ページへのリンクを含むべき', () => {
      render(<HistoryListItem item={mockHistoryItem} />)

      const link = screen.getByRole('link')
      expect(link).toHaveAttribute('href', '/history/test-id-123')
    })
  })
})
