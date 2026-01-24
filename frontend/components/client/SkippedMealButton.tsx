'use client'

import { useState } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import type { MealType } from '@/types/nutrition'
import styles from './SkippedMealButton.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface SkippedMealButtonProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function SkippedMealButton({ mealType, mealDate, onComplete }: SkippedMealButtonProps) {
  const { token } = useAuth()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async () => {
    if (!token) return

    setIsSubmitting(true)
    setError(null)

    try {
      const response = await fetch(`${API_BASE_URL}/api/meals/skip`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          meal_type: mealType,
          meal_date: mealDate,
        }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        throw new Error(errorText || '記録に失敗しました')
      }

      onComplete?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : '記録に失敗しました')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.icon}>🚫</div>
      <p className={styles.description}>この食事を「食べなかった」として記録します</p>

      {error && <div className={styles.error}>{error}</div>}

      <button
        type="button"
        onClick={handleSubmit}
        disabled={isSubmitting}
        className={styles.submitButton}
      >
        {isSubmitting ? '記録中...' : '食べませんでした'}
      </button>
    </div>
  )
}
