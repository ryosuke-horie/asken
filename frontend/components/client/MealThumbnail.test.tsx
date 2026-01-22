import { describe, it, expect } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import MealThumbnail from './MealThumbnail'

describe('MealThumbnail', () => {
  it('画像を表示すべき', () => {
    render(<MealThumbnail src="/test-image.jpg" />)

    const img = screen.getByRole('img', { name: '食事' })
    expect(img).toBeInTheDocument()
    expect(img).toHaveAttribute('src', '/test-image.jpg')
  })

  it('カスタムalt属性を設定できるべき', () => {
    render(<MealThumbnail src="/test-image.jpg" alt="カスタム説明" />)

    expect(screen.getByRole('img', { name: 'カスタム説明' })).toBeInTheDocument()
  })

  it('classNameを適用すべき', () => {
    render(<MealThumbnail src="/test-image.jpg" className="custom-class" />)

    const img = screen.getByRole('img', { name: '食事' })
    expect(img).toHaveClass('custom-class')
  })

  it('画像読み込みエラー時にプレースホルダーを表示すべき', () => {
    render(<MealThumbnail src="/invalid-image.jpg" />)

    const img = screen.getByRole('img', { name: '食事' })
    fireEvent.error(img)

    expect(screen.queryByRole('img')).not.toBeInTheDocument()
    expect(screen.getByText('🍽️')).toBeInTheDocument()
  })

  it('エラー時にプレースホルダーにclassNameを適用すべき', () => {
    const { container } = render(
      <MealThumbnail src="/invalid-image.jpg" className="custom-class" />
    )

    const img = screen.getByRole('img', { name: '食事' })
    fireEvent.error(img)

    const placeholder = container.querySelector('.custom-class')
    expect(placeholder).toBeInTheDocument()
  })
})
