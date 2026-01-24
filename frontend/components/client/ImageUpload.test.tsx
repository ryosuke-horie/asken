import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ImageUpload from './ImageUpload'
import * as storage from '@/lib/storage'

// storageモジュールをモック
vi.mock('@/lib/storage', () => ({
  saveAnalysisId: vi.fn(),
  getAnalysisId: vi.fn(),
  clearAnalysisId: vi.fn(),
}))

// NutritionDisplayコンポーネントをモック
vi.mock('./NutritionDisplay', () => ({
  default: ({ result }: { result: unknown }) => (
    <div data-testid="nutrition-display">{JSON.stringify(result)}</div>
  ),
}))

describe('ImageUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(storage.getAnalysisId).mockReturnValue(null)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('ファイル選択', () => {
    it('初期状態でアップロードボタンが無効であるべき', () => {
      render(<ImageUpload />)

      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      expect(uploadButton).toBeDisabled()
    })

    it('有効なJPEGファイルを選択できるべき', () => {
      render(<ImageUpload />)

      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      expect(fileInput).toBeTruthy()

      Object.defineProperty(fileInput, 'files', {
        value: [file],
      })
      fireEvent.change(fileInput)

      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      expect(uploadButton).not.toBeDisabled()
    })

    it('有効なPNGファイルを選択できるべき', () => {
      render(<ImageUpload />)

      const file = new File(['test'], 'test.png', { type: 'image/png' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement

      Object.defineProperty(fileInput, 'files', {
        value: [file],
      })
      fireEvent.change(fileInput)

      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      expect(uploadButton).not.toBeDisabled()
    })

    it('HEICファイルを選択できるべき', () => {
      render(<ImageUpload />)

      const file = new File(['test'], 'test.heic', { type: '' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement

      Object.defineProperty(fileInput, 'files', {
        value: [file],
      })
      fireEvent.change(fileInput)

      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      expect(uploadButton).not.toBeDisabled()
    })

    it('無効なファイル形式を拒否すべき', () => {
      render(<ImageUpload />)

      const file = new File(['test'], 'test.txt', { type: 'text/plain' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement

      Object.defineProperty(fileInput, 'files', {
        value: [file],
      })
      fireEvent.change(fileInput)

      expect(
        screen.getByText('JPEG, PNG, HEIC形式の画像のみアップロードできます'),
      ).toBeInTheDocument()
    })

    it('10MBを超えるファイルを拒否すべき', () => {
      render(<ImageUpload />)

      // 11MBのファイルを作成
      const largeContent = new Array(11 * 1024 * 1024).fill('a').join('')
      const file = new File([largeContent], 'large.jpg', { type: 'image/jpeg' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement

      Object.defineProperty(fileInput, 'files', {
        value: [file],
      })
      fireEvent.change(fileInput)

      expect(screen.getByText('ファイルサイズは10MB以下にしてください')).toBeInTheDocument()
    })
  })

  describe('アップロード処理', () => {
    it('アップロード成功時にポーリングを開始すべき', async () => {
      const mockFetch = vi.fn()
      global.fetch = mockFetch

      // アップロード成功
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-123' }),
      })

      // ステータスチェック - completed
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            status: 'completed',
            result: {
              foods: [{ name: 'テスト食品', quantity: '1個' }],
              total_nutrition: {
                calories: 100,
                protein: 10,
                fat: 5,
                carbohydrates: 15,
              },
            },
          }),
      })

      render(<ImageUpload />)

      // ファイル選択
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      Object.defineProperty(fileInput, 'files', { value: [file] })
      fireEvent.change(fileInput)

      // アップロードボタンをクリック
      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      fireEvent.click(uploadButton)

      await waitFor(() => {
        expect(storage.saveAnalysisId).toHaveBeenCalledWith('test-123')
      })

      await waitFor(() => {
        expect(screen.getByTestId('nutrition-display')).toBeInTheDocument()
      })
    })

    it('アップロード失敗時にエラーを表示すべき', async () => {
      const mockFetch = vi.fn()
      global.fetch = mockFetch

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        text: () => Promise.resolve('Internal Server Error'),
      })

      render(<ImageUpload />)

      // ファイル選択
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      Object.defineProperty(fileInput, 'files', { value: [file] })
      fireEvent.change(fileInput)

      // アップロードボタンをクリック
      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      fireEvent.click(uploadButton)

      await waitFor(() => {
        expect(screen.getByText(/アップロードに失敗しました/)).toBeInTheDocument()
      })
    })
  })

  describe('ステータスポーリング', () => {
    it('処理中ステータスでメッセージを表示すべき', async () => {
      const mockFetch = vi.fn()
      global.fetch = mockFetch

      // アップロード成功
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-123' }),
      })

      // ステータスチェック - processing
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ status: 'processing' }),
      })

      render(<ImageUpload />)

      // ファイル選択
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      Object.defineProperty(fileInput, 'files', { value: [file] })
      fireEvent.change(fileInput)

      // アップロードボタンをクリック
      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      fireEvent.click(uploadButton)

      await waitFor(() => {
        expect(screen.getByText('分析処理中です...')).toBeInTheDocument()
      })
    })

    it('失敗ステータスでエラーを表示すべき', async () => {
      const mockFetch = vi.fn()
      global.fetch = mockFetch

      // アップロード成功
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ analysis_id: 'test-123' }),
      })

      // ステータスチェック - failed
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            status: 'failed',
            error: '画像の分析に失敗しました',
          }),
      })

      render(<ImageUpload />)

      // ファイル選択
      const file = new File(['test'], 'test.jpg', { type: 'image/jpeg' })
      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      Object.defineProperty(fileInput, 'files', { value: [file] })
      fireEvent.change(fileInput)

      // アップロードボタンをクリック
      const uploadButton = screen.getByRole('button', {
        name: 'アップロードして分析',
      })
      fireEvent.click(uploadButton)

      await waitFor(() => {
        expect(screen.getByText('画像の分析に失敗しました')).toBeInTheDocument()
      })
    })
  })

  describe('セッション復旧', () => {
    it('保存されたanalysisIdがある場合、ポーリングを再開すべき', async () => {
      vi.mocked(storage.getAnalysisId).mockReturnValue('saved-123')

      const mockFetch = vi.fn()
      global.fetch = mockFetch

      // ステータスチェック - completed
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () =>
          Promise.resolve({
            status: 'completed',
            result: {
              foods: [{ name: 'テスト食品', quantity: '1個' }],
              total_nutrition: {
                calories: 100,
                protein: 10,
                fat: 5,
                carbohydrates: 15,
              },
            },
          }),
      })

      render(<ImageUpload />)

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith('/api/analyze/saved-123')
      })

      await waitFor(() => {
        expect(screen.getByTestId('nutrition-display')).toBeInTheDocument()
      })
    })
  })
})
