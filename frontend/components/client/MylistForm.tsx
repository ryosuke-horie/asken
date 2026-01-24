'use client'

import { useState, useEffect } from 'react'
import type { NutritionInfo } from '@/types/nutrition'
import type {
  CreateMylistItemRequest,
  UpdateMylistItemRequest,
  AnalyzeMylistResponse,
} from '@/types/mylist'
import styles from './MylistForm.module.css'

interface MylistFormProps {
  initialData?: {
    name: string
    base_amount: string
    unit: string
    calories: number
    protein: number
    fat: number
    carbohydrates: number
    foods: NutritionInfo[]
  }
  analyzeResult?: AnalyzeMylistResponse | null
  inputText?: string
  onSubmit: (data: CreateMylistItemRequest | UpdateMylistItemRequest) => Promise<void>
  isSubmitting: boolean
  submitLabel: string
}

export default function MylistForm({
  initialData,
  analyzeResult,
  inputText,
  onSubmit,
  isSubmitting,
  submitLabel,
}: MylistFormProps) {
  const [name, setName] = useState(initialData?.name || '')
  const [baseAmount, setBaseAmount] = useState(initialData?.base_amount || '1')
  const [unit, setUnit] = useState(initialData?.unit || '人前')
  const [calories, setCalories] = useState(initialData?.calories || 0)
  const [protein, setProtein] = useState(initialData?.protein || 0)
  const [fat, setFat] = useState(initialData?.fat || 0)
  const [carbohydrates, setCarbohydrates] = useState(initialData?.carbohydrates || 0)
  const [foods, setFoods] = useState<NutritionInfo[]>(initialData?.foods || [])

  useEffect(() => {
    if (analyzeResult) {
      setCalories(analyzeResult.total_calories)
      setProtein(analyzeResult.total_protein)
      setFat(analyzeResult.total_fat)
      setCarbohydrates(analyzeResult.total_carbohydrates)
      setFoods(analyzeResult.foods)
      if (inputText && !name) {
        setName(inputText.slice(0, 100))
      }
    }
  }, [analyzeResult, inputText, name])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    await onSubmit({
      name,
      base_amount: baseAmount,
      unit,
      calories,
      protein,
      fat,
      carbohydrates,
      foods,
    })
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.field}>
        <label className={styles.label}>メニュー名 *</label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="例: いつものランチセット"
          className={styles.input}
          maxLength={100}
          required
        />
      </div>

      <div className={styles.row}>
        <div className={styles.field}>
          <label className={styles.label}>基準量 *</label>
          <input
            type="text"
            value={baseAmount}
            onChange={(e) => setBaseAmount(e.target.value)}
            placeholder="1"
            className={styles.input}
            required
          />
        </div>
        <div className={styles.field}>
          <label className={styles.label}>単位 *</label>
          <select
            value={unit}
            onChange={(e) => setUnit(e.target.value)}
            className={styles.select}
            required
          >
            <option value="人前">人前</option>
            <option value="個">個</option>
            <option value="g">g</option>
            <option value="ml">ml</option>
            <option value="パック">パック</option>
            <option value="枚">枚</option>
            <option value="本">本</option>
          </select>
        </div>
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionTitle}>栄養素</h3>
        <div className={styles.nutritionGrid}>
          <div className={styles.nutritionItem}>
            <label className={styles.nutritionLabel}>カロリー</label>
            <div className={styles.nutritionInput}>
              <input
                type="number"
                value={calories}
                onChange={(e) => setCalories(Number(e.target.value))}
                step="0.1"
                min="0"
                className={styles.input}
              />
              <span className={styles.nutritionUnit}>kcal</span>
            </div>
          </div>
          <div className={styles.nutritionItem}>
            <label className={styles.nutritionLabel}>タンパク質</label>
            <div className={styles.nutritionInput}>
              <input
                type="number"
                value={protein}
                onChange={(e) => setProtein(Number(e.target.value))}
                step="0.1"
                min="0"
                className={styles.input}
              />
              <span className={styles.nutritionUnit}>g</span>
            </div>
          </div>
          <div className={styles.nutritionItem}>
            <label className={styles.nutritionLabel}>脂質</label>
            <div className={styles.nutritionInput}>
              <input
                type="number"
                value={fat}
                onChange={(e) => setFat(Number(e.target.value))}
                step="0.1"
                min="0"
                className={styles.input}
              />
              <span className={styles.nutritionUnit}>g</span>
            </div>
          </div>
          <div className={styles.nutritionItem}>
            <label className={styles.nutritionLabel}>炭水化物</label>
            <div className={styles.nutritionInput}>
              <input
                type="number"
                value={carbohydrates}
                onChange={(e) => setCarbohydrates(Number(e.target.value))}
                step="0.1"
                min="0"
                className={styles.input}
              />
              <span className={styles.nutritionUnit}>g</span>
            </div>
          </div>
        </div>
      </div>

      {foods.length > 0 && (
        <div className={styles.section}>
          <h3 className={styles.sectionTitle}>含まれる食品</h3>
          <div className={styles.foodsList}>
            {foods.map((food, index) => (
              <div key={index} className={styles.foodItem}>
                <span className={styles.foodName}>{food.name}</span>
                <span className={styles.foodAmount}>{food.estimated_amount}</span>
                <span className={styles.foodCalories}>{Math.round(food.calories_kcal)} kcal</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <button
        type="submit"
        disabled={isSubmitting || !name.trim() || !baseAmount.trim() || !unit.trim()}
        className={styles.submitButton}
      >
        {isSubmitting ? '保存中...' : submitLabel}
      </button>
    </form>
  )
}
