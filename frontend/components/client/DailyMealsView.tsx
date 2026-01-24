'use client'

import { useMemo } from 'react'
import useSWR from 'swr'
import { DailyMeals, MealType } from '@/types/nutrition'
import { createAuthFetcher } from '@/lib/fetcher'
import { useAuth } from '@/contexts/AuthContext'
import DateNavigation from './DateNavigation'
import MealSection from './MealSection'
import DailyTotalSummary from './DailyTotalSummary'
import styles from './DailyMealsView.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface DailyMealsViewProps {
  date: string
}

export default function DailyMealsView({ date }: DailyMealsViewProps) {
  const { token } = useAuth()
  const authFetcher = useMemo(() => createAuthFetcher(token), [token])

  const { data, error, mutate } = useSWR<DailyMeals>(
    token ? `${API_BASE_URL}/api/meals/daily?date=${date}` : null,
    authFetcher,
  )

  const handleDelete = () => {
    mutate()
  }

  if (error) {
    return <div className={styles.error}>データの取得に失敗しました。再読み込みしてください。</div>
  }

  if (!data) return <div className={styles.loading}>読み込み中...</div>

  const mealTypes: MealType[] = ['breakfast', 'lunch', 'dinner', 'snack']

  return (
    <div className={styles.container}>
      <DateNavigation currentDate={date} />

      <DailyTotalSummary
        totalCalories={data.daily_total.total_calories}
        totalProtein={data.daily_total.total_protein}
        totalFat={data.daily_total.total_fat}
        totalCarbohydrates={data.daily_total.total_carbohydrates}
      />

      <div className={styles.sections}>
        {mealTypes.map((mealType) => (
          <MealSection
            key={mealType}
            mealType={mealType}
            mealDate={date}
            meals={data.meals[mealType]}
            onDelete={handleDelete}
          />
        ))}
      </div>
    </div>
  )
}
