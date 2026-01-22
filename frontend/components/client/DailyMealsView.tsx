'use client'

import useSWR from 'swr'
import { DailyMeals, MealType } from '@/types/nutrition'
import { fetcher } from '@/lib/fetcher'
import DateNavigation from './DateNavigation'
import MealSection from './MealSection'
import DailyTotalSummary from './DailyTotalSummary'
import styles from './DailyMealsView.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface DailyMealsViewProps {
  initialData: DailyMeals
  initialDate: string
}

export default function DailyMealsView({ initialData, initialDate }: DailyMealsViewProps) {
  const { data, error, mutate } = useSWR<DailyMeals>(
    `${API_BASE_URL}/api/meals/daily?date=${initialDate}`,
    fetcher,
    { fallbackData: initialData }
  )

  const handleDelete = () => {
    mutate()
  }

  if (error) {
    return (
      <div className={styles.error}>
        データの取得に失敗しました。再読み込みしてください。
      </div>
    )
  }

  if (!data) return <div className={styles.loading}>読み込み中...</div>

  const mealTypes: MealType[] = ['breakfast', 'lunch', 'dinner', 'snack']

  return (
    <div className={styles.container}>
      <DateNavigation currentDate={initialDate} />

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
            mealDate={initialDate}
            meals={data.meals[mealType]}
            onDelete={handleDelete}
          />
        ))}
      </div>
    </div>
  )
}
