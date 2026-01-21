import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import DateNavigation from './DateNavigation'

// useRouter mock
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

describe('DateNavigation', () => {
  beforeEach(() => {
    mockPush.mockClear()
    // 今日の日付を固定（2026-01-21 水曜日）
    jest.useFakeTimers()
    jest.setSystemTime(new Date('2026-01-21'))
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  describe('週間表示', () => {
    it('月曜日から日曜日までの7日間を表示すべき', () => {
      // 2026-01-21（水曜日）の週は 1/19(月)〜1/25(日)
      render(<DateNavigation currentDate="2026-01-21" />)

      // 19〜25の日付が表示されているか
      expect(screen.getByText('19')).toBeInTheDocument()
      expect(screen.getByText('20')).toBeInTheDocument()
      expect(screen.getByText('21')).toBeInTheDocument()
      expect(screen.getByText('22')).toBeInTheDocument()
      expect(screen.getByText('23')).toBeInTheDocument()
      expect(screen.getByText('24')).toBeInTheDocument()
      expect(screen.getByText('25')).toBeInTheDocument()
    })

    it('曜日のヘッダーを表示すべき', () => {
      render(<DateNavigation currentDate="2026-01-21" />)

      expect(screen.getByText('月')).toBeInTheDocument()
      expect(screen.getByText('火')).toBeInTheDocument()
      expect(screen.getByText('水')).toBeInTheDocument()
      expect(screen.getByText('木')).toBeInTheDocument()
      expect(screen.getByText('金')).toBeInTheDocument()
      expect(screen.getByText('土')).toBeInTheDocument()
      expect(screen.getByText('日')).toBeInTheDocument()
    })

    it('選択中の日付がハイライトされるべき', () => {
      render(<DateNavigation currentDate="2026-01-21" />)

      const selectedDayButton = screen.getByRole('button', { name: /21/ })
      expect(selectedDayButton).toHaveClass('selected')
    })

    it('今日の日付に特別なマークがあるべき', () => {
      render(<DateNavigation currentDate="2026-01-21" />)

      const todayButton = screen.getByRole('button', { name: /21/ })
      expect(todayButton).toHaveClass('today')
    })

    it('今日と選択中の日付が異なる場合、両方が区別できるべき', () => {
      render(<DateNavigation currentDate="2026-01-22" />)

      // 今日（21日）はtodayクラスのみ
      const todayButton = screen.getByRole('button', { name: /21/ })
      expect(todayButton).toHaveClass('today')
      expect(todayButton).not.toHaveClass('selected')

      // 選択中の日付（22日）はselectedクラス
      const selectedDayButton = screen.getByRole('button', { name: /22/ })
      expect(selectedDayButton).toHaveClass('selected')
      expect(selectedDayButton).not.toHaveClass('today')
    })
  })

  describe('日付クリック', () => {
    it('日付をクリックするとその日に遷移すべき', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      render(<DateNavigation currentDate="2026-01-21" />)

      const dayButton = screen.getByRole('button', { name: /19/ })
      await user.click(dayButton)

      expect(mockPush).toHaveBeenCalledWith('/?date=2026-01-19')
    })

    it('週の最後の日をクリックするとその日に遷移すべき', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      render(<DateNavigation currentDate="2026-01-21" />)

      const dayButton = screen.getByRole('button', { name: /25/ })
      await user.click(dayButton)

      expect(mockPush).toHaveBeenCalledWith('/?date=2026-01-25')
    })
  })

  describe('月をまたぐ週', () => {
    it('月をまたぐ週でも正しく表示すべき', () => {
      // 2026-01-29（木曜日）の週は 1/26(月)〜2/1(日)
      render(<DateNavigation currentDate="2026-01-29" />)

      expect(screen.getByText('26')).toBeInTheDocument()
      expect(screen.getByText('27')).toBeInTheDocument()
      expect(screen.getByText('28')).toBeInTheDocument()
      expect(screen.getByText('29')).toBeInTheDocument()
      expect(screen.getByText('30')).toBeInTheDocument()
      expect(screen.getByText('31')).toBeInTheDocument()
      expect(screen.getByText('1')).toBeInTheDocument()
    })
  })

  describe('年末年始', () => {
    it('年をまたぐ週でも正しく動作すべき', async () => {
      jest.setSystemTime(new Date('2025-12-31'))
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      // 2025-12-31（水曜日）の週は 12/29(月)〜1/4(日)
      render(<DateNavigation currentDate="2025-12-31" />)

      // 1月4日をクリック
      const dayButton = screen.getByRole('button', { name: /4日/ })
      await user.click(dayButton)

      expect(mockPush).toHaveBeenCalledWith('/?date=2026-01-04')
    })
  })
})
