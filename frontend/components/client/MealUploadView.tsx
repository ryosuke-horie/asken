'use client'

import { useRouter } from 'next/navigation'
import { useMemo, useState } from 'react'
import useSWR from 'swr'
import { MealType, DailyMeals } from '@/types/nutrition'
import { createAuthFetcher } from '@/lib/fetcher'
import { useAuth } from '@/contexts/AuthContext'
import MealInputSelector from './MealInputSelector'
import DeleteHistoryButton from './DeleteHistoryButton'
import MealThumbnail from './MealThumbnail'
import styles from './MealUploadView.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface MealUploadViewProps {
  mealType: MealType
  mealDate: string
  mealLabel: string
}

const MEAL_TYPE_ICONS: Record<MealType, string> = {
  breakfast: '🌅',
  lunch: '☀️',
  dinner: '🌙',
  snack: '☕',
}

export default function MealUploadView({ mealType, mealDate, mealLabel }: MealUploadViewProps) {
  const router = useRouter()
  const { token } = useAuth()
  const authFetcher = useMemo(() => createAuthFetcher(token), [token])
  const [uploadCount, setUploadCount] = useState(0)
  const [syncError, setSyncError] = useState(false)

  // SWRで最新のデータを取得
  const { data, mutate } = useSWR<DailyMeals>(
    token ? `${API_BASE_URL}/api/meals/daily?date=${mealDate}` : null,
    authFetcher,
    {
      refreshInterval: 5000,
      onError: (err) => {
        console.error('データ同期エラー:', err)
        setSyncError(true)
      },
      onSuccess: () => {
        setSyncError(false)
      }
    }
  )

  const meals = data?.meals[mealType] || []

  const handleUploadComplete = () => {
    setUploadCount((prev) => prev + 1)
    mutate() // データを再取得
  }

  const handleDeleteSuccess = () => {
    mutate() // データを再取得
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    const month = date.getMonth() + 1
    const day = date.getDate()
    const weekday = ['日', '月', '火', '水', '木', '金', '土'][date.getDay()]
    return `${month}月${day}日（${weekday}）`
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <button onClick={() => router.push(`/?date=${mealDate}`)} className={styles.backButton}>
          ← 戻る
        </button>
        <div className={styles.dateInfo}>{formatDate(mealDate)}</div>
      </div>

      <div className={styles.titleSection}>
        <div className={styles.icon}>{MEAL_TYPE_ICONS[mealType]}</div>
        <h1 className={styles.title}>{mealLabel}</h1>
      </div>

      {syncError && (
        <div className={styles.syncError}>
          データの同期に失敗しました。ネットワーク接続を確認してください。
        </div>
      )}

      {uploadCount > 0 && (
        <div className={styles.successMessage}>
          {uploadCount}件の食事を分析しました。続けて入力できます。
        </div>
      )}

      <div className={styles.uploadSection}>
        <MealInputSelector
          mealType={mealType}
          mealDate={mealDate}
          onComplete={handleUploadComplete}
        />
      </div>

      {meals.length > 0 && (
        <div className={styles.mealsSection}>
          <h2 className={styles.sectionTitle}>登録済みの食事</h2>
          <div className={styles.mealsList}>
            {meals.map((meal) => (
              <div key={meal.id} className={styles.mealItem}>
                {meal.input_type === 'text' ? (
                  <div className={styles.textPreview}>
                    {meal.input_text}
                  </div>
                ) : (
                  <MealThumbnail
                    src={`${API_BASE_URL}/api/images/${meal.image_path?.split('/').pop()}`}
                    className={styles.thumbnail}
                  />
                )}
                <div className={styles.mealInfo}>
                  <div className={styles.calories}>
                    {Math.round(meal.total_calories)} <span className={styles.unit}>kcal</span>
                  </div>
                  <div className={styles.nutrients}>
                    <span className={styles.nutrient}>P: {meal.total_protein.toFixed(1)}g</span>
                    <span className={styles.nutrient}>F: {meal.total_fat.toFixed(1)}g</span>
                    <span className={styles.nutrient}>C: {meal.total_carbohydrates.toFixed(1)}g</span>
                  </div>
                </div>
                <div className={styles.deleteButton}>
                  <DeleteHistoryButton
                    historyId={meal.id}
                    iconOnly
                    onSuccess={handleDeleteSuccess}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
