'use client'

import { useState, useMemo, useCallback } from 'react'
import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import {
  ConditionRecord,
  ConditionLevel,
  FatigueLevel,
  CONDITION_LABELS,
  FATIGUE_LABELS,
  isConditionLevel,
  isFatigueLevel,
} from '@/types/condition'
import ConditionRecordForm from './ConditionRecordForm'
import styles from './ConditionSection.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface ConditionSectionProps {
  date: string
}

export default function ConditionSection({ date }: ConditionSectionProps) {
  const { token } = useAuth()
  const authFetcher = useMemo(() => createAuthFetcher(token), [token])
  const [isRecording, setIsRecording] = useState(false)

  const {
    data: record,
    error,
    isLoading,
    mutate,
  } = useSWR<ConditionRecord | null>(
    token ? `${API_BASE_URL}/api/condition-records?date=${date}` : null,
    authFetcher,
  )

  const handleRecordCondition = useCallback(
    async (condition: ConditionLevel, fatigue: FatigueLevel) => {
      if (!token) {
        throw new Error('認証が必要です。再ログインしてください')
      }

      setIsRecording(true)
      try {
        const response = await fetch(`${API_BASE_URL}/api/condition-records`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            condition,
            fatigue,
            recorded_at: date,
          }),
        })

        if (!response.ok) {
          const errorText = await response.text()
          throw new Error(errorText || '記録に失敗しました')
        }

        await mutate()
      } finally {
        setIsRecording(false)
      }
    },
    [token, date, mutate],
  )

  if (error) {
    return <div className={styles.error}>{error.message || 'データの取得に失敗しました'}</div>
  }

  if (isLoading) {
    return <div className={styles.loading}>読み込み中...</div>
  }

  return (
    <section className={styles.section}>
      <div className={styles.header}>
        <h2 className={styles.title}>体調・疲労度</h2>
      </div>

      {record && (
        <div className={styles.summary}>
          <div className={styles.item}>
            <span className={styles.label}>体調</span>
            <span className={styles.value}>
              {isConditionLevel(record.condition) ? CONDITION_LABELS[record.condition] : '不明'}
            </span>
          </div>
          <div className={styles.item}>
            <span className={styles.label}>疲労度</span>
            <span className={styles.value}>
              {isFatigueLevel(record.fatigue) ? FATIGUE_LABELS[record.fatigue] : '不明'}
            </span>
          </div>
        </div>
      )}

      <div className={styles.recordForm}>
        <ConditionRecordForm
          onSubmit={handleRecordCondition}
          isLoading={isRecording}
          initialCondition={
            record && isConditionLevel(record.condition) ? record.condition : undefined
          }
          initialFatigue={record && isFatigueLevel(record.fatigue) ? record.fatigue : undefined}
        />
      </div>
    </section>
  )
}
