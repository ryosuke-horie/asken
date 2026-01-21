'use client'

import { ChangeEvent, useEffect, useState, useRef } from 'react'
import { AnalysisResult } from '@/types/nutrition'
import NutritionDisplay from './NutritionDisplay'
import { saveAnalysisId, getAnalysisId, clearAnalysisId } from '@/lib/storage'
import styles from './ImageUpload.module.css'

type AnalysisStatus = 'pending' | 'processing' | 'completed' | 'failed'

interface StatusResponse {
  status: AnalysisStatus
  message?: string
  result?: AnalysisResult
  error?: string
}

const API_BASE_URL = ''

export default function ImageUpload() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<AnalysisResult | null>(null)
  const [analysisId, setAnalysisId] = useState<string | null>(null)
  const [statusMessage, setStatusMessage] = useState<string>('')
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null)

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) {
      return
    }

    // ファイルバリデーション（JPEG, PNG, HEIC、最大10MB）
    const validTypes = ['image/jpeg', 'image/png', 'image/heic']
    const isHeic = file.name.toLowerCase().endsWith('.heic')

    if (!validTypes.includes(file.type) && !isHeic) {
      setError('JPEG, PNG, HEIC形式の画像のみアップロードできます')
      return
    }

    if (file.size > 10 * 1024 * 1024) {
      setError('ファイルサイズは10MB以下にしてください')
      return
    }

    setSelectedFile(file)
    setError(null)
    setResult(null)

    // プレビュー生成（HEICはブラウザでプレビューできないため、ファイル名のみ表示）
    if (isHeic) {
      setPreviewUrl(null)
    } else {
      const reader = new FileReader()
      reader.onloadend = () => {
        setPreviewUrl(reader.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  // ポーリング停止
  const stopPolling = () => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current)
      pollingIntervalRef.current = null
    }
  }

  // ステータスチェック
  const checkStatus = async (id: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/analyze/${id}`)

      if (!response.ok) {
        throw new Error(`ステータス取得に失敗しました (${response.status})`)
      }

      const data: StatusResponse = await response.json()

      switch (data.status) {
        case 'pending':
          setStatusMessage('分析リクエストを受け付けました...')
          break

        case 'processing':
          setStatusMessage('分析処理中です...')
          break

        case 'completed':
          setStatusMessage('分析が完了しました')
          setResult(data.result || null)
          setIsLoading(false)
          stopPolling()
          clearAnalysisId()
          setAnalysisId(null)
          break

        case 'failed':
          setError(data.error || '分析に失敗しました')
          setIsLoading(false)
          stopPolling()
          clearAnalysisId()
          setAnalysisId(null)
          break
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '予期しないエラーが発生しました')
      setIsLoading(false)
      stopPolling()
      clearAnalysisId()
      setAnalysisId(null)
    }
  }

  // ポーリング開始
  const startPolling = (id: string) => {
    // 既存のポーリングを停止
    stopPolling()

    // 即座に1回チェック
    checkStatus(id)

    // 2秒間隔でポーリング
    pollingIntervalRef.current = setInterval(() => {
      checkStatus(id)
    }, 2000)
  }

  // コンポーネントマウント時にlocalStorageから復旧
  useEffect(() => {
    const savedAnalysisId = getAnalysisId()
    if (savedAnalysisId) {
      setAnalysisId(savedAnalysisId)
      setIsLoading(true)
      setStatusMessage('分析を再開しています...')
      startPolling(savedAnalysisId)
    }

    // クリーンアップ
    return () => {
      stopPolling()
    }
  }, [])

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('画像ファイルを選択してください')
      return
    }

    setIsLoading(true)
    setError(null)
    setResult(null)
    setStatusMessage('アップロード中...')

    try {
      // FormData作成
      const formData = new FormData()
      formData.append('image', selectedFile)

      // POST /api/analyze（202 Accepted を期待）
      const response = await fetch(`${API_BASE_URL}/api/analyze`, {
        method: 'POST',
        body: formData,
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`アップロードに失敗しました (${response.status}): ${errorText}`)
      }

      // analysis_idを取得
      const data: { analysis_id: string } = await response.json()
      const newAnalysisId = data.analysis_id

      // localStorageに保存
      saveAnalysisId(newAnalysisId)
      setAnalysisId(newAnalysisId)

      // ポーリング開始
      startPolling(newAnalysisId)
    } catch (err) {
      if (err instanceof Error) {
        setError(`エラー: ${err.message}`)
      } else {
        setError('予期しないエラーが発生しました')
      }
      setIsLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.uploadSection}>
        <h2>画像をアップロード</h2>

        <input
          type="file"
          accept="image/jpeg,image/png,image/heic,.heic"
          onChange={handleFileChange}
          disabled={isLoading}
        />

        {selectedFile && (
          <div className={styles.preview}>
            {previewUrl ? (
              <img src={previewUrl} alt="プレビュー" />
            ) : (
              <p>選択したファイル: {selectedFile.name} ({(selectedFile.size / 1024 / 1024).toFixed(2)} MB)</p>
            )}
          </div>
        )}

        <button
          onClick={handleUpload}
          disabled={!selectedFile || isLoading}
          className={styles.uploadButton}
        >
          {isLoading ? '分析中...' : 'アップロードして分析'}
        </button>

        {error && (
          <div className={styles.errorMessage}>
            {error}
          </div>
        )}

        {isLoading && (
          <div className={styles.loadingMessage}>
            {statusMessage || '画像を分析しています。2分ほどお待ちください...'}
          </div>
        )}
      </div>

      {result && <NutritionDisplay result={result} />}
    </div>
  )
}
