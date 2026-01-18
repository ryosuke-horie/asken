'use client'

import Link from 'next/link'
import { HistoryDetail, MealType } from '@/types/nutrition'
import styles from './MealSection.module.css'

interface MealSectionProps {
  mealType: MealType
  mealDate: string
  meals: HistoryDetail[]
}

const mealTypeLabels: Record<MealType, string> = {
  breakfast: '朝食',
  lunch: '昼食',
  dinner: '夕食',
  snack: '間食',
}

const mealTypeIcons: Record<MealType, string> = {
  breakfast: '🌅',
  lunch: '☀️',
  dinner: '🌙',
  snack: '☕',
}

export default function MealSection({ mealType, mealDate, meals }: MealSectionProps) {
  const totalCalories = meals.reduce((sum, meal) => sum + meal.total_calories, 0)

  return (
    <div className={styles.section}>
      <Link href={`/meals/${mealType}?date=${mealDate}`} className={styles.link}>
        <div className={styles.header}>
          <div className={styles.title}>
            <span className={styles.icon}>{mealTypeIcons[mealType]}</span>
            <span className={styles.label}>{mealTypeLabels[mealType]}</span>
          </div>
          <div className={styles.arrow}>→</div>
        </div>

        <div className={styles.content}>
          {meals.length > 0 ? (
            <>
              <div className={styles.total}>
                {Math.round(totalCalories)} <span className={styles.unit}>kcal</span>
              </div>
              <div className={styles.mealsCount}>
                {meals.length}件の記録
              </div>
            </>
          ) : (
            <div className={styles.empty}>
              タップして記録する
            </div>
          )}
        </div>
      </Link>
    </div>
  )
}
