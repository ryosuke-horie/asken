'use client'

import styles from './DailyTotalSummary.module.css'

interface DailyTotalSummaryProps {
  totalCalories: number
  totalProtein: number
  totalFat: number
  totalCarbohydrates: number
}

export default function DailyTotalSummary({
  totalCalories,
  totalProtein,
  totalFat,
  totalCarbohydrates,
}: DailyTotalSummaryProps) {
  return (
    <div className={styles.container}>
      <h2 className={styles.title}>本日の合計</h2>
      <div className={styles.summary}>
        <div className={styles.mainCalories}>
          {Math.round(totalCalories)} <span className={styles.unit}>kcal</span>
        </div>
        <div className={styles.nutrients}>
          <div className={styles.nutrient}>
            <span className={styles.label}>タンパク質</span>
            <span className={styles.value}>{totalProtein.toFixed(1)}g</span>
          </div>
          <div className={styles.nutrient}>
            <span className={styles.label}>脂質</span>
            <span className={styles.value}>{totalFat.toFixed(1)}g</span>
          </div>
          <div className={styles.nutrient}>
            <span className={styles.label}>炭水化物</span>
            <span className={styles.value}>{totalCarbohydrates.toFixed(1)}g</span>
          </div>
        </div>
      </div>
    </div>
  )
}
