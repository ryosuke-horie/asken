'use client'

import { useState } from 'react'
import { MealType, InputType } from '@/types/nutrition'
import MealTypeUpload from './MealTypeUpload'
import TextInput from './TextInput'
import MylistSelector from './MylistSelector'
import SkippedMealButton from './SkippedMealButton'
import styles from './MealInputSelector.module.css'

interface MealInputSelectorProps {
  mealType: MealType
  mealDate: string
  onComplete?: () => void
}

export default function MealInputSelector({ mealType, mealDate, onComplete }: MealInputSelectorProps) {
  const [inputType, setInputType] = useState<InputType>('mylist')

  return (
    <div className={styles.container}>
      <div className={styles.tabs}>
        <button
          type="button"
          onClick={() => setInputType('mylist')}
          className={`${styles.tab} ${inputType === 'mylist' ? styles.active : ''}`}
        >
          マイリスト
        </button>
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
        <button
          type="button"
          onClick={() => setInputType('skipped')}
          className={`${styles.tab} ${inputType === 'skipped' ? styles.active : ''}`}
        >
          食べなかった
        </button>
      </div>

      <div className={styles.content}>
        {inputType === 'mylist' ? (
          <MylistSelector mealType={mealType} mealDate={mealDate} onComplete={onComplete} />
        ) : inputType === 'image' ? (
          <MealTypeUpload mealType={mealType} mealDate={mealDate} onComplete={onComplete} />
        ) : inputType === 'text' ? (
          <TextInput mealType={mealType} mealDate={mealDate} onComplete={onComplete} />
        ) : (
          <SkippedMealButton mealType={mealType} mealDate={mealDate} onComplete={onComplete} />
        )}
      </div>
    </div>
  )
}
