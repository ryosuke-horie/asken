'use client'

import { useState, useRef } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import type { AnalyzeMylistResponse } from '@/types/mylist'
import styles from './MylistAnalyzeInput.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

type AnalysisMethod = 'text' | 'image'

interface MylistAnalyzeInputProps {
  onAnalyze: (inputText: string) => Promise<AnalyzeMylistResponse>
  onResult: (result: AnalyzeMylistResponse, inputText: string, imagePath?: string) => void
  onCancel?: () => void
}

export default function MylistAnalyzeInput({ onAnalyze, onResult, onCancel }: MylistAnalyzeInputProps) {
  const { token } = useAuth()
  const [analysisMethod, setAnalysisMethod] = useState<AnalysisMethod>('text')
  const [inputText, setInputText] = useState('')
  const [selectedImage, setSelectedImage] = useState<File | null>(null)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const abortControllerRef = useRef<AbortController | null>(null)

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setSelectedImage(file)
    setError(null)

    const reader = new FileReader()
    reader.onload = (e) => {
      setImagePreview(e.target?.result as string)
    }
    reader.readAsDataURL(file)
  }

  const handleTextSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputText.trim()) return

    setIsLoading(true)
    setError(null)

    try {
      const result = await onAnalyze(inputText.trim())
      onResult(result, inputText.trim())
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') return
      setError(err instanceof Error ? err.message : 'AI分析に失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  const handleImageSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedImage || !token) return

    setIsLoading(true)
    setError(null)
    abortControllerRef.current = new AbortController()

    try {
      const formData = new FormData()
      formData.append('image', selectedImage)
      formData.append('meal_type', 'snack')
      formData.append('meal_date', new Date().toISOString().split('T')[0])

      const response = await fetch(`${API_BASE_URL}/api/analyze`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
        signal: abortControllerRef.current.signal,
      })

      if (!response.ok) {
        throw new Error('画像分析の開始に失敗しました')
      }

      const { id, image_path } = await response.json()

      const result = await pollForResult(id, abortControllerRef.current.signal)
      onResult(result, selectedImage.name, image_path)
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') return
      setError(err instanceof Error ? err.message : 'AI分析に失敗しました')
    } finally {
      setIsLoading(false)
      abortControllerRef.current = null
    }
  }

  const pollForResult = async (id: string, signal: AbortSignal): Promise<AnalyzeMylistResponse> => {
    const maxAttempts = 60
    const interval = 2000

    for (let i = 0; i < maxAttempts; i++) {
      if (signal.aborted) throw new DOMException('Aborted', 'AbortError')

      await new Promise((resolve) => setTimeout(resolve, interval))

      const response = await fetch(`${API_BASE_URL}/api/analyze/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
        signal,
      })

      if (!response.ok) continue

      const data = await response.json()

      if (data.status === 'completed' && data.result) {
        return {
          foods: data.result.foods,
          total_calories: data.result.total_calories,
          total_protein: data.result.total_protein,
          total_fat: data.result.total_fat,
          total_carbohydrates: data.result.total_carbohydrates,
        }
      }

      if (data.status === 'failed') {
        throw new Error(data.error_message || '分析に失敗しました')
      }
    }

    throw new Error('分析がタイムアウトしました')
  }

  const handleCancel = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort()
    }
    setIsLoading(false)
    setError(null)
    onCancel?.()
  }

  const clearImage = () => {
    setSelectedImage(null)
    setImagePreview(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  if (isLoading) {
    return (
      <div className={styles.loadingContainer}>
        <div className={styles.spinner} />
        <p className={styles.loadingText}>AIが栄養素を分析中です...</p>
        <p className={styles.loadingSubtext}>しばらくお待ちください（最大2分程度）</p>
        <button type="button" onClick={handleCancel} className={styles.cancelButton}>
          キャンセルして戻る
        </button>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.modeSelector}>
        <button
          type="button"
          onClick={() => setAnalysisMethod('text')}
          className={`${styles.modeButton} ${analysisMethod === 'text' ? styles.modeActive : ''}`}
        >
          テキストで分析
        </button>
        <button
          type="button"
          onClick={() => setAnalysisMethod('image')}
          className={`${styles.modeButton} ${analysisMethod === 'image' ? styles.modeActive : ''}`}
        >
          画像で分析
        </button>
      </div>

      {analysisMethod === 'text' ? (
        <form onSubmit={handleTextSubmit} className={styles.form}>
          <label className={styles.label}>食事メニューを入力</label>
          <textarea
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            placeholder="例: サラダチキン、ゆで卵2個、ブロッコリー100g"
            className={styles.textarea}
            rows={3}
          />
          {error && <div className={styles.error}>{error}</div>}
          <button type="submit" disabled={!inputText.trim()} className={styles.button}>
            AIで栄養素を分析
          </button>
        </form>
      ) : (
        <form onSubmit={handleImageSubmit} className={styles.form}>
          <label className={styles.label}>食事の画像をアップロード</label>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            onChange={handleImageSelect}
            className={styles.fileInput}
          />
          <button
            type="button"
            onClick={() => fileInputRef.current?.click()}
            className={styles.fileButton}
          >
            画像を選択
          </button>
          {imagePreview && (
            <div className={styles.preview}>
              <img src={imagePreview} alt="プレビュー" className={styles.previewImage} />
              <button type="button" onClick={clearImage} className={styles.clearButton}>
                画像を削除
              </button>
            </div>
          )}
          {error && <div className={styles.error}>{error}</div>}
          <button type="submit" disabled={!selectedImage} className={styles.button}>
            AIで栄養素を分析
          </button>
        </form>
      )}
    </div>
  )
}
