'use client'

import Link from 'next/link'
import { HistoryDetail, MealType } from '@/types/nutrition'
import DeleteHistoryButton from './DeleteHistoryButton'
import MealThumbnail from './MealThumbnail'
import styles from './MealSection.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

interface MealSectionProps {
  mealType: MealType
  mealDate: string
  meals: HistoryDetail[]
  onDelete?: () => void
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

export default function MealSection({ mealType, mealDate, meals, onDelete }: MealSectionProps) {
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

      {meals.length > 0 && (
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
                <div className={styles.mealCalories}>
                  {Math.round(meal.total_calories)} <span className={styles.mealUnit}>kcal</span>
                </div>
                <div className={styles.nutrients}>
                  <span>P: {meal.total_protein.toFixed(1)}g</span>
                  <span>F: {meal.total_fat.toFixed(1)}g</span>
                  <span>C: {meal.total_carbohydrates.toFixed(1)}g</span>
                </div>
              </div>
              <div className={styles.deleteButton}>
                <DeleteHistoryButton
                  historyId={meal.id}
                  iconOnly
                  onSuccess={onDelete}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
