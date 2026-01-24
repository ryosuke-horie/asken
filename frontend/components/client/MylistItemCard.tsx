'use client'

import type { MylistItem } from '@/types/mylist'
import styles from './MylistItemCard.module.css'

interface MylistItemCardProps {
  item: MylistItem
  onEdit: (item: MylistItem) => void
  onDelete: (id: string) => void
}

export default function MylistItemCard({ item, onEdit, onDelete }: MylistItemCardProps) {
  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation()
    onEdit(item)
  }

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (window.confirm(`「${item.name}」を削除しますか？`)) {
      onDelete(item.id)
    }
  }

  return (
    <div className={styles.card}>
      <div className={styles.content}>
        <div className={styles.header}>
          <h3 className={styles.name}>{item.name}</h3>
          <div className={styles.actions}>
            <button onClick={handleEdit} className={styles.editButton} aria-label="編集">
              ✏️
            </button>
            <button onClick={handleDelete} className={styles.deleteButton} aria-label="削除">
              🗑️
            </button>
          </div>
        </div>

        <div className={styles.amount}>
          {item.base_amount} {item.unit}
        </div>

        <div className={styles.nutrition}>
          <div className={styles.calories}>
            {Math.round(item.calories)} <span className={styles.unit}>kcal</span>
          </div>
          <div className={styles.macros}>
            <span className={styles.macro}>P: {item.protein.toFixed(1)}g</span>
            <span className={styles.macro}>F: {item.fat.toFixed(1)}g</span>
            <span className={styles.macro}>C: {item.carbohydrates.toFixed(1)}g</span>
          </div>
        </div>

        {item.foods.length > 0 && (
          <div className={styles.foods}>
            {item.foods.map((food, index) => (
              <span key={index} className={styles.foodTag}>
                {food.name}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
