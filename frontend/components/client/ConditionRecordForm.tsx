'use client'

import { useState, useEffect } from 'react'
import { ConditionLevel, FatigueLevel, CONDITION_LABELS, FATIGUE_LABELS } from '@/types/condition'
import styles from './ConditionRecordForm.module.css'

interface ConditionRecordFormProps {
  onSubmit: (condition: ConditionLevel, fatigue: FatigueLevel) => Promise<void>
  isLoading: boolean
  initialCondition?: ConditionLevel
  initialFatigue?: FatigueLevel
}

const CONDITION_LEVELS: ConditionLevel[] = [1, 2, 3]
const FATIGUE_LEVELS: FatigueLevel[] = [1, 2, 3]

export default function ConditionRecordForm({
  onSubmit,
  isLoading,
  initialCondition,
  initialFatigue,
}: ConditionRecordFormProps) {
  const [condition, setCondition] = useState<ConditionLevel | null>(initialCondition ?? null)
  const [fatigue, setFatigue] = useState<FatigueLevel | null>(initialFatigue ?? null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (initialCondition !== undefined) {
      setCondition(initialCondition)
    }
  }, [initialCondition])

  useEffect(() => {
    if (initialFatigue !== undefined) {
      setFatigue(initialFatigue)
    }
  }, [initialFatigue])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (condition === null || fatigue === null) {
      setError('体調と疲労度を選択してください')
      return
    }

    try {
      await onSubmit(condition, fatigue)
    } catch (err) {
      const message = err instanceof Error ? err.message : '記録に失敗しました'
      setError(message)
    }
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      <div className={styles.fieldGroup}>
        <label className={styles.label}>体調</label>
        <div className={styles.buttonGroup}>
          {CONDITION_LEVELS.map((level) => (
            <button
              key={level}
              type="button"
              className={`${styles.levelButton} ${condition === level ? styles.selected : ''}`}
              onClick={() => setCondition(level)}
              disabled={isLoading}
            >
              {CONDITION_LABELS[level]}
            </button>
          ))}
        </div>
      </div>

      <div className={styles.fieldGroup}>
        <label className={styles.label}>疲労度</label>
        <div className={styles.buttonGroup}>
          {FATIGUE_LEVELS.map((level) => (
            <button
              key={level}
              type="button"
              className={`${styles.levelButton} ${fatigue === level ? styles.selected : ''}`}
              onClick={() => setFatigue(level)}
              disabled={isLoading}
            >
              {FATIGUE_LABELS[level]}
            </button>
          ))}
        </div>
      </div>

      {error && <p className={styles.error}>{error}</p>}

      <button
        type="submit"
        className={styles.submitButton}
        disabled={isLoading || condition === null || fatigue === null}
      >
        {isLoading ? '記録中...' : '記録する'}
      </button>
    </form>
  )
}
