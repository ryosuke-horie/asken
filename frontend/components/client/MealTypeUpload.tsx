'use client'

import { ChangeEvent, useState, useRef } from 'react'
import { MealType } from '@/types/nutrition'
import styles from './MealTypeUpload.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

type AnalysisStatus = 'pending' | 'processing' | 'completed' | 'failed'

interface StatusResponse {
  status: AnalysisStatus
  error?: string
}

interface MealTypeUploadProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function MealTypeUpload({ mealType, mealDate, onComplete }: MealTypeUploadProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [statusMessage, setStatusMessage] = useState<string>('')
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null)

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) {
      return
    }

    // ファイルバリデーション
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

    // プレビュー生成
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
          stopPolling()
          setIsLoading(false)
          setSelectedFile(null)
          setPreviewUrl(null)
          // 親コンポーネントに通知
          onComplete?.()
          break

        case 'failed':
          setError(data.error || '分析に失敗しました')
          stopPolling()
          setIsLoading(false)
          break
      }
    } catch (err) {
      console.error('Status check error:', err)
      setError(err instanceof Error ? err.message : '予期しないエラー')
      stopPolling()
      setIsLoading(false)
    }
  }

  // ポーリング開始
  const startPolling = (id: string) => {
    stopPolling()
    checkStatus(id)
    pollingIntervalRef.current = setInterval(() => {
      checkStatus(id)
    }, 2000)
  }

  const handleUpload = async () => {
    if (!selectedFile) {
      return
    }

    setIsLoading(true)
    setError(null)
    setStatusMessage('アップロード中...')

    try {
      const formData = new FormData()
      formData.append('image', selectedFile)
      formData.append('meal_type', mealType)
      formData.append('meal_date', mealDate)

      const response = await fetch(`${API_BASE_URL}/api/analyze`, {
        method: 'POST',
        body: formData,
      })

      if (!response.ok) {
        throw new Error(`アップロードに失敗しました (${response.status})`)
      }

      const data: { analysis_id: string } = await response.json()
      startPolling(data.analysis_id)
    } catch (err) {
      setError(err instanceof Error ? err.message : '予期しないエラー')
      setIsLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.uploadArea}>
        <input
          type="file"
          accept="image/jpeg,image/png,image/heic,.heic"
          onChange={handleFileChange}
          disabled={isLoading}
          className={styles.fileInput}
          id={`file-input-${mealType}`}
        />
        <label htmlFor={`file-input-${mealType}`} className={styles.fileLabel}>
          {selectedFile ? selectedFile.name : '画像を選択'}
        </label>

        {selectedFile && previewUrl && (
          <div className={styles.preview}>
            <img src={previewUrl} alt="プレビュー" className={styles.previewImage} />
          </div>
        )}

        <button
          onClick={handleUpload}
          disabled={!selectedFile || isLoading}
          className={styles.uploadButton}
        >
          {isLoading ? '分析中...' : '📷 解析'}
        </button>

        {error && <div className={styles.error}>{error}</div>}
        {isLoading && <div className={styles.loading}>{statusMessage}</div>}
      </div>
    </div>
  )
}
