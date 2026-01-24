'use client'

import { useState } from 'react'
import type { AnalyzeMylistResponse } from '@/types/mylist'
import styles from './MylistAnalyzeInput.module.css'

interface MylistAnalyzeInputProps {
  onAnalyze: (inputText: string) => Promise<AnalyzeMylistResponse>
  onResult: (result: AnalyzeMylistResponse, inputText: string) => void
}

export default function MylistAnalyzeInput({ onAnalyze, onResult }: MylistAnalyzeInputProps) {
  const [inputText, setInputText] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputText.trim()) return

    setIsLoading(true)
    setError(null)

    try {
      const result = await onAnalyze(inputText.trim())
      onResult(result, inputText.trim())
    } catch (err) {
      setError(err instanceof Error ? err.message : 'AI分析に失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <form onSubmit={handleSubmit} className={styles.form}>
        <label className={styles.label}>食事メニューを入力</label>
        <textarea
          value={inputText}
          onChange={(e) => setInputText(e.target.value)}
          placeholder="例: サラダチキン、ゆで卵2個、ブロッコリー100g"
          className={styles.textarea}
          rows={3}
          disabled={isLoading}
        />
        {error && <div className={styles.error}>{error}</div>}
        <button type="submit" disabled={!inputText.trim() || isLoading} className={styles.button}>
          {isLoading ? '分析中...' : 'AIで栄養素を分析'}
        </button>
      </form>
    </div>
  )
}
