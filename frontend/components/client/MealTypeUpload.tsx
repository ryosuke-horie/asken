'use client'

import { ChangeEvent, useState } from 'react'
import { MealType } from '@/types/nutrition'
import { useAnalysisPolling } from '@/hooks/useAnalysisPolling'
import { useAuth } from '@/contexts/AuthContext'
import styles from './MealTypeUpload.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface MealTypeUploadProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function MealTypeUpload({ mealType, mealDate, onComplete }: MealTypeUploadProps) {
  const { token } = useAuth()
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)

  const {
    isLoading,
    error,
    statusMessage,
    startAnalysis,
    setIsLoading,
    setError,
    setStatusMessage,
  } = useAnalysisPolling({
    onComplete: () => {
      setSelectedFile(null)
      setPreviewUrl(null)
      onComplete?.()
    },
  })

  function handleFileChange(e: ChangeEvent<HTMLInputElement>): void {
    const file = e.target.files?.[0]
    if (!file) {
      return
    }

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

    if (isHeic) {
      setPreviewUrl(null)
    } else {
      const reader = new FileReader()
      reader.onloadend = () => {
        setPreviewUrl(reader.result as string)
      }
      reader.onerror = () => {
        console.error('Failed to read file for preview:', reader.error)
        setPreviewUrl(null)
      }
      reader.readAsDataURL(file)
    }
  }

  async function handleUpload(): Promise<void> {
    if (!selectedFile) {
      return
    }

    if (!token) {
      setError('認証が必要です。再度ログインしてください。')
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
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      })

      if (!response.ok) {
        throw new Error(`アップロードに失敗しました (${response.status})`)
      }

      const data: { analysis_id: string } = await response.json()
      startAnalysis(data.analysis_id)
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
