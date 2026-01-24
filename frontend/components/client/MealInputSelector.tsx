'use client'

import { useState } from 'react'
import { MealType, InputType } from '@/types/nutrition'
import MealTypeUpload from './MealTypeUpload'
import TextInput from './TextInput'
import styles from './MealInputSelector.module.css'

interface MealInputSelectorProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function MealInputSelector({
  mealType,
  mealDate,
  onComplete,
}: MealInputSelectorProps) {
  const [inputType, setInputType] = useState<InputType>('image')

  return (
    <div className={styles.container}>
      <div className={styles.tabs}>
        <button
          type="button"
          onClick={() => setInputType('image')}
          className={`${styles.tab} ${inputType === 'image' ? styles.active : ''}`}
        >
          画像
        </button>
        <button
          type="button"
          onClick={() => setInputType('text')}
          className={`${styles.tab} ${inputType === 'text' ? styles.active : ''}`}
        >
          テキスト
        </button>
      </div>

      <div className={styles.content}>
        {inputType === 'image' ? (
          <MealTypeUpload mealType={mealType} mealDate={mealDate} onComplete={onComplete} />
        ) : (
          <TextInput mealType={mealType} mealDate={mealDate} onComplete={onComplete} />
        )}
      </div>
    </div>
  )
}
