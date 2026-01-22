import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import MealInputSelector from './MealInputSelector'

vi.mock('./MealTypeUpload', () => ({
  default: () => <div data-testid="meal-type-upload">MealTypeUpload</div>,
}))

vi.mock('./TextInput', () => ({
  default: () => <div data-testid="text-input">TextInput</div>,
}))

describe('MealInputSelector', () => {
  const defaultProps = {
    mealType: 'lunch' as const,
    mealDate: '2024-01-15',
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('初期表示', () => {
    it('画像タブとテキストタブが表示されるべき', () => {
      render(<MealInputSelector {...defaultProps} />)

      expect(screen.getByText('画像')).toBeInTheDocument()
      expect(screen.getByText('テキスト')).toBeInTheDocument()
    })

    it('初期状態では画像入力コンポーネントが表示されるべき', () => {
      render(<MealInputSelector {...defaultProps} />)

      expect(screen.getByTestId('meal-type-upload')).toBeInTheDocument()
      expect(screen.queryByTestId('text-input')).not.toBeInTheDocument()
    })

    it('初期状態では画像タブがアクティブになるべき', () => {
      render(<MealInputSelector {...defaultProps} />)

      const imageTab = screen.getByText('画像')
      expect(imageTab.className).toContain('active')
    })
  })

  describe('タブ切り替え', () => {
    it('テキストタブをクリックするとTextInputが表示されるべき', () => {
      render(<MealInputSelector {...defaultProps} />)

      fireEvent.click(screen.getByText('テキスト'))

      expect(screen.getByTestId('text-input')).toBeInTheDocument()
      expect(screen.queryByTestId('meal-type-upload')).not.toBeInTheDocument()
    })

    it('テキストタブをクリックするとテキストタブがアクティブになるべき', () => {
      render(<MealInputSelector {...defaultProps} />)

      fireEvent.click(screen.getByText('テキスト'))

      const textTab = screen.getByText('テキスト')
      expect(textTab.className).toContain('active')
    })

    it('画像タブに戻るとMealTypeUploadが表示されるべき', () => {
      render(<MealInputSelector {...defaultProps} />)

      fireEvent.click(screen.getByText('テキスト'))
      expect(screen.getByTestId('text-input')).toBeInTheDocument()

      fireEvent.click(screen.getByText('画像'))
      expect(screen.getByTestId('meal-type-upload')).toBeInTheDocument()
      expect(screen.queryByTestId('text-input')).not.toBeInTheDocument()
    })
  })
})
