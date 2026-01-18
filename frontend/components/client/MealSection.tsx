'use client'

import { HistoryDetail, MealType } from '@/types/nutrition'
import MealTypeUpload from './MealTypeUpload'
import styles from './MealSection.module.css'

interface MealSectionProps {
  mealType: MealType
  mealDate: string
  meals: HistoryDetail[]
  onRefresh: () => void
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

export default function MealSection({ mealType, mealDate, meals, onRefresh }: MealSectionProps) {
  const totalCalories = meals.reduce((sum, meal) => sum + meal.total_calories, 0)

  return (
    <div className={styles.section}>
      <div className={styles.header}>
        <div className={styles.title}>
          <span className={styles.icon}>{mealTypeIcons[mealType]}</span>
          <span className={styles.label}>{mealTypeLabels[mealType]}</span>
        </div>
        {meals.length > 0 && (
          <div className={styles.total}>
            {Math.round(totalCalories)} kcal
          </div>
        )}
      </div>

      <div className={styles.content}>
        <MealTypeUpload
          mealType={mealType}
          mealDate={mealDate}
          onComplete={onRefresh}
        />

        {meals.length > 0 && (
          <div className={styles.mealsList}>
            {meals.map((meal) => (
              <div key={meal.id} className={styles.mealItem}>
                <img
                  src={`${process.env.NEXT_PUBLIC_API_URL}/api/images/${meal.image_path.split('/').pop()}`}
                  alt="食事"
                  className={styles.thumbnail}
                />
                <div className={styles.nutrients}>
                  <span className={styles.calories}>{Math.round(meal.total_calories)} kcal</span>
                  <span className={styles.nutrient}>P: {meal.total_protein.toFixed(1)}g</span>
                  <span className={styles.nutrient}>F: {meal.total_fat.toFixed(1)}g</span>
                  <span className={styles.nutrient}>C: {meal.total_carbohydrates.toFixed(1)}g</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
