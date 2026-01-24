'use client'

import { useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import MylistAnalyzeInput from '@/components/client/MylistAnalyzeInput'
import MylistForm from '@/components/client/MylistForm'
import { useMylist } from '@/hooks/useMylist'
import { useAuth } from '@/contexts/AuthContext'
import type { AnalyzeMylistResponse, CreateMylistItemRequest } from '@/types/mylist'
import styles from './page.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

type EntryMode = 'select' | 'ai' | 'manual'

export default function MylistNewPage() {
  const router = useRouter()
  const { token } = useAuth()
  const { createItem, analyzeText } = useMylist()
  const [entryMode, setEntryMode] = useState<EntryMode>('select')
  const [analyzeResult, setAnalyzeResult] = useState<AnalyzeMylistResponse | null>(null)
  const [inputText, setInputText] = useState<string>('')
  const [imagePath, setImagePath] = useState<string | undefined>(undefined)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Manual mode image upload
  const [manualImagePreview, setManualImagePreview] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleAnalyzeResult = (result: AnalyzeMylistResponse, text: string, imgPath?: string) => {
    setAnalyzeResult(result)
    setInputText(text)
    setImagePath(imgPath)
    setError(null)
  }

  const handleManualImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !token) return

    // Show preview
    const reader = new FileReader()
    reader.onload = (e) => {
      setManualImagePreview(e.target?.result as string)
    }
    reader.readAsDataURL(file)

    // Upload image
    setIsUploading(true)
    setError(null)

    try {
      const formData = new FormData()
      formData.append('image', file)

      const response = await fetch(`${API_BASE_URL}/api/upload-image`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      })

      if (!response.ok) {
        throw new Error('画像のアップロードに失敗しました')
      }

      const { image_path } = await response.json()
      setImagePath(image_path)
    } catch (err) {
      setError(err instanceof Error ? err.message : '画像のアップロードに失敗しました')
      setManualImagePreview(null)
    } finally {
      setIsUploading(false)
    }
  }

  const clearManualImage = () => {
    setManualImagePreview(null)
    setImagePath(undefined)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleSubmit = async (data: CreateMylistItemRequest) => {
    setIsSubmitting(true)
    setError(null)

    try {
      const requestData = {
        ...data,
        image_path: imagePath,
      }
      await createItem(requestData)
      router.push('/mylist')
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSelectMode = (mode: 'ai' | 'manual') => {
    setEntryMode(mode)
    if (mode === 'manual') {
      setAnalyzeResult({
        foods: [],
        total_calories: 0,
        total_protein: 0,
        total_fat: 0,
        total_carbohydrates: 0,
      })
    }
  }

  const handleBack = () => {
    setEntryMode('select')
    setAnalyzeResult(null)
    setInputText('')
    setImagePath(undefined)
    setManualImagePreview(null)
    setError(null)
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <Link href="/mylist" className={styles.backLink}>
            ← マイリスト
          </Link>
          <h1 className={styles.title}>新規登録</h1>
        </div>

        {error && <div className={styles.error}>{error}</div>}

        {entryMode === 'select' && (
          <div className={styles.modeSelect}>
            <p className={styles.description}>
              登録方法を選択してください
            </p>
            <div className={styles.modeButtons}>
              <button
                type="button"
                onClick={() => handleSelectMode('ai')}
                className={styles.modeCard}
              >
                <span className={styles.modeIcon}>AI</span>
                <span className={styles.modeTitle}>AIで栄養素を分析</span>
                <span className={styles.modeDescription}>
                  テキストまたは画像からAIが自動で栄養素を計算
                </span>
              </button>
              <button
                type="button"
                onClick={() => handleSelectMode('manual')}
                className={styles.modeCard}
              >
                <span className={styles.modeIcon}>✏️</span>
                <span className={styles.modeTitle}>手動で入力</span>
                <span className={styles.modeDescription}>
                  栄養素を直接入力（画像添付も可能）
                </span>
              </button>
            </div>
          </div>
        )}

        {entryMode === 'ai' && !analyzeResult && (
          <div className={styles.analyzeSection}>
            <button type="button" onClick={handleBack} className={styles.backButton}>
              ← 戻る
            </button>
            <p className={styles.description}>
              よく食べるメニューを入力して、AIで栄養素を分析しましょう。
            </p>
            <MylistAnalyzeInput
              onAnalyze={analyzeText}
              onResult={handleAnalyzeResult}
              onCancel={handleBack}
            />
          </div>
        )}

        {(entryMode === 'ai' && analyzeResult) && (
          <div className={styles.formSection}>
            <button type="button" onClick={handleBack} className={styles.backButton}>
              ← 再分析する
            </button>
            <MylistForm
              analyzeResult={analyzeResult}
              inputText={inputText}
              onSubmit={handleSubmit}
              isSubmitting={isSubmitting}
              submitLabel="登録する"
            />
          </div>
        )}

        {entryMode === 'manual' && analyzeResult && (
          <div className={styles.formSection}>
            <button type="button" onClick={handleBack} className={styles.backButton}>
              ← 戻る
            </button>

            <div className={styles.imageUploadSection}>
              <label className={styles.imageLabel}>画像を添付（任意）</label>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleManualImageSelect}
                className={styles.fileInput}
                disabled={isUploading}
              />
              {!manualImagePreview ? (
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className={styles.fileButton}
                  disabled={isUploading}
                >
                  {isUploading ? 'アップロード中...' : '画像を選択'}
                </button>
              ) : (
                <div className={styles.imagePreview}>
                  <img src={manualImagePreview} alt="プレビュー" className={styles.previewImage} />
                  <button type="button" onClick={clearManualImage} className={styles.clearImageButton}>
                    画像を削除
                  </button>
                </div>
              )}
            </div>

            <MylistForm
              analyzeResult={analyzeResult}
              inputText={inputText}
              onSubmit={handleSubmit}
              isSubmitting={isSubmitting}
              submitLabel="登録する"
            />
          </div>
        )}
      </div>
    </ProtectedRoute>
  )
}
