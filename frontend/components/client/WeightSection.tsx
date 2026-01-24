'use client'

import { useState, useMemo, useCallback } from 'react'
import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import { WeightRecordsResponse, WeightGoal, WeightPeriod } from '@/types/weight'
import WeightChart from './WeightChart'
import WeightRecordForm from './WeightRecordForm'
import WeightGoalSetting from './WeightGoalSetting'
import styles from './WeightSection.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

const PERIOD_LABELS: Record<WeightPeriod, string> = {
  week: '1週間',
  month: '1ヶ月',
  '3months': '3ヶ月',
}

export default function WeightSection() {
  const { token } = useAuth()
  const authFetcher = useMemo(() => createAuthFetcher(token), [token])
  const [period, setPeriod] = useState<WeightPeriod>('week')
  const [isRecording, setIsRecording] = useState(false)
  const [isUpdatingGoal, setIsUpdatingGoal] = useState(false)
  const [showGoalSetting, setShowGoalSetting] = useState(false)

  const {
    data: recordsData,
    error: recordsError,
    mutate: mutateRecords,
  } = useSWR<WeightRecordsResponse>(
    token ? `${API_BASE_URL}/api/weight-records?period=${period}` : null,
    authFetcher
  )

  const {
    data: goalData,
    error: goalError,
    mutate: mutateGoal,
  } = useSWR<WeightGoal | null>(token ? `${API_BASE_URL}/api/weight-goal` : null, authFetcher)

  const handleRecordWeight = useCallback(
    async (weight: number) => {
      if (!token) {
        throw new Error('認証が必要です。再ログインしてください')
      }

      setIsRecording(true)
      try {
        const response = await fetch(`${API_BASE_URL}/api/weight-records`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            weight,
            recorded_at: new Date().toISOString().split('T')[0],
          }),
        })

        if (!response.ok) {
          const errorText = await response.text()
          throw new Error(errorText || '記録に失敗しました')
        }

        await Promise.all([mutateRecords(), mutateGoal()])
      } finally {
        setIsRecording(false)
      }
    },
    [token, mutateRecords, mutateGoal]
  )

  const handleUpdateGoal = useCallback(
    async (targetWeight: number, targetDate: string) => {
      if (!token) {
        throw new Error('認証が必要です。再ログインしてください')
      }

      setIsUpdatingGoal(true)
      try {
        const response = await fetch(`${API_BASE_URL}/api/weight-goal`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
          },
          body: JSON.stringify({
            target_weight: targetWeight,
            target_date: targetDate,
          }),
        })

        if (!response.ok) {
          const errorText = await response.text()
          throw new Error(errorText || '目標の設定に失敗しました')
        }

        await mutateGoal()
        setShowGoalSetting(false)
      } finally {
        setIsUpdatingGoal(false)
      }
    },
    [token, mutateGoal]
  )

  if (recordsError || goalError) {
    const errorMessage = recordsError?.message || goalError?.message || 'データの取得に失敗しました'
    return <div className={styles.error}>{errorMessage}</div>
  }

  const records = recordsData?.records ?? []
  const latest = recordsData?.latest
  const stats = recordsData?.stats

  return (
    <section className={styles.section}>
      <div className={styles.header}>
        <h2 className={styles.title}>体重管理</h2>
        <button
          type="button"
          className={styles.goalButton}
          onClick={() => setShowGoalSetting(!showGoalSetting)}
        >
          {showGoalSetting ? '閉じる' : '目標設定'}
        </button>
      </div>

      {showGoalSetting && (
        <div className={styles.goalSection}>
          <WeightGoalSetting
            currentGoal={goalData ?? null}
            onSubmit={handleUpdateGoal}
            isLoading={isUpdatingGoal}
          />
        </div>
      )}

      <div className={styles.summary}>
        {latest && (
          <div className={styles.currentWeight}>
            <span className={styles.label}>現在</span>
            <span className={styles.weight}>{latest.weight} kg</span>
          </div>
        )}
        {goalData && (
          <div className={styles.goalWeight}>
            <span className={styles.label}>目標</span>
            <span className={styles.weight}>{goalData.target_weight} kg</span>
            {goalData.weight_to_lose > 0 && (
              <span className={styles.diff}>(-{goalData.weight_to_lose} kg)</span>
            )}
          </div>
        )}
        {stats && records.length > 0 && (
          <div className={styles.stats}>
            <span className={styles.label}>期間内</span>
            <span className={styles.statsValue}>
              {stats.min.toFixed(1)} - {stats.max.toFixed(1)} kg
            </span>
          </div>
        )}
      </div>

      <div className={styles.periodTabs}>
        {(Object.keys(PERIOD_LABELS) as WeightPeriod[]).map((p) => (
          <button
            key={p}
            type="button"
            className={`${styles.periodTab} ${period === p ? styles.active : ''}`}
            onClick={() => setPeriod(p)}
          >
            {PERIOD_LABELS[p]}
          </button>
        ))}
      </div>

      <WeightChart records={records} targetWeight={goalData?.target_weight} />

      <div className={styles.recordForm}>
        <WeightRecordForm onSubmit={handleRecordWeight} isLoading={isRecording} />
      </div>
    </section>
  )
}
