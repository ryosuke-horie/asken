'use client'

import { ChangeEvent, useState } from 'react'
import { AnalysisResult } from '@/types/nutrition'
import NutritionDisplay from './NutritionDisplay'
import styles from './ImageUpload.module.css'

export default function ImageUpload() {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<AnalysisResult | null>(null)

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

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('画像ファイルを選択してください')
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      // FormData作成
      const formData = new FormData()
      formData.append('image', selectedFile)

      // POST /api/analyze（タイムアウト180秒）
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 180000)

      const response = await fetch('http://localhost:8080/api/analyze', {
        method: 'POST',
        body: formData,
        signal: controller.signal,
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        const errorText = await response.text()
        console.error('Server error:', response.status, errorText)
        throw new Error(`分析に失敗しました (${response.status}): ${errorText}`)
      }

      const data: AnalysisResult = await response.json()
      setResult(data)
    } catch (err) {
      console.error('Upload error:', err)
      if (err instanceof Error) {
        if (err.name === 'AbortError') {
          setError('タイムアウト: 分析に3分以上かかっています。もう一度お試しください。')
        } else {
          setError(`エラー: ${err.message}`)
        }
      } else {
        setError('予期しないエラーが発生しました')
      }
    } finally {
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
            画像を分析しています。2分ほどお待ちください...
          </div>
        )}
      </div>

      {result && <NutritionDisplay result={result} />}
    </div>
  )
}
