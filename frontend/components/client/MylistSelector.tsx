'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useAuth } from '@/contexts/AuthContext'
import { useMylist } from '@/hooks/useMylist'
import QuantityStepper from './QuantityStepper'
import type { MealType } from '@/types/nutrition'
import type { MylistItem } from '@/types/mylist'
import styles from './MylistSelector.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface MylistSelectorProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function MylistSelector({ mealType, mealDate, onComplete }: MylistSelectorProps) {
  const { token } = useAuth()
  const { items, isLoading, error } = useMylist()
  const [selectedItem, setSelectedItem] = useState<MylistItem | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const handleSelect = (item: MylistItem) => {
    setSelectedItem(item)
    setQuantity(1)
    setSubmitError(null)
  }

  const handleBack = () => {
    setSelectedItem(null)
    setQuantity(1)
    setSubmitError(null)
  }

  const handleSubmit = async () => {
    if (!selectedItem || !token) return

    setIsSubmitting(true)
    setSubmitError(null)

    try {
      const response = await fetch(`${API_BASE_URL}/api/meals/from-mylist`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          mylist_item_id: selectedItem.id,
          quantity,
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
      setSubmitError(err instanceof Error ? err.message : '記録に失敗しました')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isLoading) {
    return <div className={styles.loading}>読み込み中...</div>
  }

  if (error) {
    return <div className={styles.error}>{error instanceof Error ? error.message : 'エラーが発生しました'}</div>
  }

  if (items.length === 0) {
    return (
      <div className={styles.empty}>
        <p>マイリストにアイテムがありません</p>
        <Link href="/mylist/new" className={styles.emptyLink}>
          マイリストに追加する
        </Link>
      </div>
    )
  }

  if (selectedItem) {
    const scaledCalories = Math.round(selectedItem.calories * quantity)
    const scaledProtein = (selectedItem.protein * quantity).toFixed(1)
    const scaledFat = (selectedItem.fat * quantity).toFixed(1)
    const scaledCarbs = (selectedItem.carbohydrates * quantity).toFixed(1)

    return (
      <div className={styles.detail}>
        <button type="button" onClick={handleBack} className={styles.backButton}>
          ← 一覧に戻る
        </button>

        <div className={styles.selectedItem}>
          <h3 className={styles.itemName}>{selectedItem.name}</h3>
          <p className={styles.itemBase}>
            {selectedItem.base_amount} {selectedItem.unit}
          </p>
        </div>

        <div className={styles.quantitySection}>
          <label className={styles.quantityLabel}>数量</label>
          <QuantityStepper value={quantity} onChange={setQuantity} min={0.5} max={10} step={0.5} />
        </div>

        <div className={styles.nutrition}>
          <div className={styles.nutritionRow}>
            <span className={styles.nutritionLabel}>カロリー</span>
            <span className={styles.nutritionValue}>{scaledCalories} kcal</span>
          </div>
          <div className={styles.nutritionRow}>
            <span className={styles.nutritionLabel}>タンパク質</span>
            <span className={styles.nutritionValue}>{scaledProtein} g</span>
          </div>
          <div className={styles.nutritionRow}>
            <span className={styles.nutritionLabel}>脂質</span>
            <span className={styles.nutritionValue}>{scaledFat} g</span>
          </div>
          <div className={styles.nutritionRow}>
            <span className={styles.nutritionLabel}>炭水化物</span>
            <span className={styles.nutritionValue}>{scaledCarbs} g</span>
          </div>
        </div>

        {submitError && <div className={styles.error}>{submitError}</div>}

        <button type="button" onClick={handleSubmit} disabled={isSubmitting} className={styles.submitButton}>
          {isSubmitting ? '記録中...' : '記録する'}
        </button>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.list}>
        {items.map((item) => (
          <button key={item.id} type="button" onClick={() => handleSelect(item)} className={styles.itemCard}>
            <div className={styles.itemHeader}>
              <span className={styles.itemName}>{item.name}</span>
              <span className={styles.itemCalories}>{Math.round(item.calories)} kcal</span>
            </div>
            <div className={styles.itemMeta}>
              {item.base_amount} {item.unit}
            </div>
          </button>
        ))}
      </div>
    </div>
  )
}
