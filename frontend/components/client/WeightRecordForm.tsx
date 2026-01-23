'use client'

import { useState } from 'react'
import styles from './WeightRecordForm.module.css'

interface WeightRecordFormProps {
  onSubmit: (weight: number) => Promise<void>
  isLoading: boolean
}

export default function WeightRecordForm({ onSubmit, isLoading }: WeightRecordFormProps) {
  const [weight, setWeight] = useState('')
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const weightNum = parseFloat(weight)
    if (isNaN(weightNum) || weightNum < 0.1 || weightNum > 999.9) {
      setError('体重は0.1〜999.9kgの範囲で入力してください')
      return
    }

    try {
      await onSubmit(weightNum)
      setWeight('')
    } catch {
      setError('記録に失敗しました')
    }
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.inputGroup}>
        <input
          type="number"
          step="0.1"
          min="0.1"
          max="999.9"
          value={weight}
          onChange={(e) => setWeight(e.target.value)}
          placeholder="体重を入力"
          className={styles.input}
          disabled={isLoading}
        />
        <span className={styles.unit}>kg</span>
      </div>
      {error && <p className={styles.error}>{error}</p>}
      <button type="submit" className={styles.button} disabled={isLoading || !weight}>
        {isLoading ? '記録中...' : '記録する'}
      </button>
    </form>
  )
}
