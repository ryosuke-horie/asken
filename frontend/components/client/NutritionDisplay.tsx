'use client'

import { AnalysisResult } from '@/types/nutrition'
import styles from './NutritionDisplay.module.css'

interface Props {
  result: AnalysisResult
}

export default function NutritionDisplay({ result }: Props) {
  return (
    <div className={styles.nutritionDisplay}>
      <h2>栄養素情報</h2>

      <table className={styles.nutritionTable}>
        <thead>
          <tr>
            <th>食材名</th>
            <th>量</th>
            <th>カロリー (kcal)</th>
            <th>タンパク質 (g)</th>
            <th>脂質 (g)</th>
            <th>炭水化物 (g)</th>
          </tr>
        </thead>
        <tbody>
          {result.foods.map((food, index) => (
            <tr key={index}>
              <td>{food.name}</td>
              <td>{food.estimated_amount}</td>
              <td>{food.calories_kcal.toFixed(0)}</td>
              <td>{food.protein_g.toFixed(1)}</td>
              <td>{food.fat_g.toFixed(1)}</td>
              <td>{food.carbohydrates_g.toFixed(1)}</td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr className={styles.totalRow}>
            <td colSpan={2}><strong>合計</strong></td>
            <td><strong>{result.total_calories.toFixed(0)}</strong></td>
            <td><strong>{result.total_protein.toFixed(1)}</strong></td>
            <td><strong>{result.total_fat.toFixed(1)}</strong></td>
            <td><strong>{result.total_carbohydrates.toFixed(1)}</strong></td>
          </tr>
        </tfoot>
      </table>
    </div>
  )
}
