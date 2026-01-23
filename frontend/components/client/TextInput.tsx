'use client'

import { useState } from 'react'
import { MealType } from '@/types/nutrition'
import { useAnalysisPolling } from '@/hooks/useAnalysisPolling'
import { useAuth } from '@/contexts/AuthContext'
import styles from './TextInput.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface TextInputProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function TextInput({ mealType, mealDate, onComplete }: TextInputProps) {
  const { token } = useAuth()
  const [inputs, setInputs] = useState<string[]>([''])

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
      setInputs([''])
      onComplete?.()
    },
  })

  function handleInputChange(index: number, value: string): void {
    setInputs(prev => {
      const newInputs = [...prev]
      newInputs[index] = value
      return newInputs
    })
  }

  function handleAddInput(): void {
    setInputs(prev => [...prev, ''])
  }

  function handleRemoveInput(index: number): void {
    if (inputs.length <= 1) {
      return
    }
    setInputs(prev => prev.filter((_, i) => i !== index))
  }

  async function handleSubmit(): Promise<void> {
    const validInputs = inputs.filter(input => input.trim() !== '')
    if (validInputs.length === 0) {
      setError('食事内容を入力してください')
      return
    }

    if (!token) {
      setError('ログインが必要です')
      return
    }

    const inputText = validInputs.join(', ')

    setIsLoading(true)
    setError(null)
    setStatusMessage('送信中...')

    try {
      const response = await fetch(`${API_BASE_URL}/api/analyze`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          input_text: inputText,
          meal_type: mealType,
          meal_date: mealDate,
        }),
      })

      if (!response.ok) {
        throw new Error(`送信に失敗しました (${response.status})`)
      }

      const data: { analysis_id: string } = await response.json()
      startAnalysis(data.analysis_id)
    } catch (err) {
      setError(err instanceof Error ? err.message : '予期しないエラー')
      setIsLoading(false)
    }
  }

  const hasValidInput = inputs.some(input => input.trim() !== '')

  return (
    <div className={styles.container}>
      <div className={styles.inputList}>
        {inputs.map((input, index) => (
          <div key={index} className={styles.inputRow}>
            <input
              type="text"
              value={input}
              onChange={(e) => handleInputChange(index, e.target.value)}
              placeholder="例: ご飯二杯、焼肉"
              disabled={isLoading}
              className={styles.textInput}
            />
            <button
              type="button"
              onClick={() => handleRemoveInput(index)}
              disabled={isLoading || inputs.length <= 1}
              className={styles.removeButton}
              aria-label="削除"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                <line x1="10" y1="11" x2="10" y2="17" />
                <line x1="14" y1="11" x2="14" y2="17" />
              </svg>
            </button>
          </div>
        ))}
      </div>

      <div className={styles.buttonGroup}>
        <button
          type="button"
          onClick={handleAddInput}
          disabled={isLoading}
          className={styles.addButton}
        >
          追加
        </button>

        <button
          type="button"
          onClick={handleSubmit}
          disabled={!hasValidInput || isLoading}
          className={styles.submitButton}
        >
          {isLoading ? '分析中...' : '解析'}
        </button>
      </div>

      {error && <div className={styles.error}>{error}</div>}
      {isLoading && <div className={styles.loading}>{statusMessage}</div>}
    </div>
  )
}
