import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ConditionRecordForm from './ConditionRecordForm'

describe('ConditionRecordForm', () => {
  const mockOnSubmit = vi.fn()

  beforeEach(() => {
    mockOnSubmit.mockClear()
  })

  it('体調と疲労度のボタンを表示すべき', () => {
    render(<ConditionRecordForm onSubmit={mockOnSubmit} isLoading={false} />)

    expect(screen.getByText('体調')).toBeInTheDocument()
    expect(screen.getByText('疲労度')).toBeInTheDocument()
    expect(screen.getByText('悪い')).toBeInTheDocument()
    expect(screen.getAllByText('普通')).toHaveLength(2)
    expect(screen.getByText('良い')).toBeInTheDocument()
    expect(screen.getByText('低い')).toBeInTheDocument()
    expect(screen.getByText('高い')).toBeInTheDocument()
  })

  it('体調と疲労度を選択して記録できるべき', async () => {
    mockOnSubmit.mockResolvedValue(undefined)

    render(<ConditionRecordForm onSubmit={mockOnSubmit} isLoading={false} />)

    fireEvent.click(screen.getByText('良い'))
    fireEvent.click(screen.getByText('低い'))
    fireEvent.click(screen.getByRole('button', { name: '記録する' }))

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(3, 1)
    })
  })

  it('選択せずに送信するとエラーを表示すべき', async () => {
    render(<ConditionRecordForm onSubmit={mockOnSubmit} isLoading={false} />)

    const submitButton = screen.getByRole('button', { name: '記録する' })
    expect(submitButton).toBeDisabled()
  })

  it('ローディング中はボタンを無効化すべき', () => {
    render(<ConditionRecordForm onSubmit={mockOnSubmit} isLoading={true} />)

    expect(screen.getByText('記録中...')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '記録中...' })).toBeDisabled()
  })

  it('初期値を設定できるべき', async () => {
    mockOnSubmit.mockResolvedValue(undefined)

    render(
      <ConditionRecordForm
        onSubmit={mockOnSubmit}
        isLoading={false}
        initialCondition={3}
        initialFatigue={3}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: '記録する' }))

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(3, 3)
    })
  })

  it('送信エラー時にエラーメッセージを表示すべき', async () => {
    mockOnSubmit.mockRejectedValue(new Error('記録に失敗しました'))

    render(<ConditionRecordForm onSubmit={mockOnSubmit} isLoading={false} />)

    fireEvent.click(screen.getByText('良い'))
    fireEvent.click(screen.getByText('低い'))
    fireEvent.click(screen.getByRole('button', { name: '記録する' }))

    await waitFor(() => {
      expect(screen.getByText('記録に失敗しました')).toBeInTheDocument()
    })
  })
})
