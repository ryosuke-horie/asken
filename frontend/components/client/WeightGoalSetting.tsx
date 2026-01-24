'use client'

import { useState } from 'react'
import { WeightGoal } from '@/types/weight'
import { formatDateFull } from '@/lib/date'
import styles from './WeightGoalSetting.module.css'

interface WeightGoalSettingProps {
  currentGoal: WeightGoal | null
  onSubmit: (targetWeight: number, targetDate: string) => Promise<void>
  isLoading: boolean
}

export default function WeightGoalSetting({
  currentGoal,
  onSubmit,
  isLoading,
}: WeightGoalSettingProps) {
  const [isEditing, setIsEditing] = useState(!currentGoal)
  const [targetWeight, setTargetWeight] = useState(currentGoal?.target_weight?.toString() ?? '')
  const [targetDate, setTargetDate] = useState(currentGoal?.target_date ?? '')
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const weightNum = parseFloat(targetWeight)
    if (isNaN(weightNum) || weightNum < 0.1 || weightNum > 999.9) {
      setError('目標体重は0.1〜999.9kgの範囲で入力してください')
      return
    }

    if (!targetDate) {
      setError('目標日を選択してください')
      return
    }

    try {
      await onSubmit(weightNum, targetDate)
      setIsEditing(false)
    } catch (err) {
      const message = err instanceof Error ? err.message : '目標の設定に失敗しました'
      setError(message)
    }
  }

  if (!isEditing && currentGoal) {
    return (
      <div className={styles.display}>
        <div className={styles.goalInfo}>
          <div className={styles.goalItem}>
            <span className={styles.label}>目標体重</span>
            <span className={styles.value}>{currentGoal.target_weight} kg</span>
          </div>
          <div className={styles.goalItem}>
            <span className={styles.label}>目標日</span>
            <span className={styles.value}>{formatDateFull(currentGoal.target_date)}</span>
          </div>
          <div className={styles.goalItem}>
            <span className={styles.label}>残り</span>
            <span className={styles.value}>{currentGoal.days_remaining} 日</span>
          </div>
          {currentGoal.weight_to_lose > 0 && (
            <div className={styles.goalItem}>
              <span className={styles.label}>減量幅</span>
              <span className={styles.weightToLose}>-{currentGoal.weight_to_lose} kg</span>
            </div>
          )}
        </div>
        <button type="button" className={styles.editButton} onClick={() => setIsEditing(true)}>
          目標を変更
        </button>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.formGroup}>
        <label className={styles.label}>目標体重</label>
        <div className={styles.inputGroup}>
          <input
            type="number"
            step="0.1"
            min="0.1"
            max="999.9"
            value={targetWeight}
            onChange={(e) => setTargetWeight(e.target.value)}
            placeholder="66.0"
            className={styles.input}
            disabled={isLoading}
          />
          <span className={styles.unit}>kg</span>
        </div>
      </div>
      <div className={styles.formGroup}>
        <label className={styles.label}>目標日</label>
        <input
          type="date"
          value={targetDate}
          onChange={(e) => setTargetDate(e.target.value)}
          className={styles.dateInput}
          disabled={isLoading}
          min={new Date().toISOString().split('T')[0]}
        />
      </div>
      {error && <p className={styles.error}>{error}</p>}
      <div className={styles.buttons}>
        <button type="submit" className={styles.submitButton} disabled={isLoading}>
          {isLoading ? '保存中...' : '目標を設定'}
        </button>
        {currentGoal && (
          <button
            type="button"
            className={styles.cancelButton}
            onClick={() => setIsEditing(false)}
            disabled={isLoading}
          >
            キャンセル
          </button>
        )}
      </div>
    </form>
  )
}
