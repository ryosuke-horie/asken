'use client'

import { useMemo } from 'react'
import Link from 'next/link'
import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import type { TrainingRecord } from '@/types/training'
import styles from './TrainingSection.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface TrainingSectionProps {
  date: string
}

export default function TrainingSection({ date }: TrainingSectionProps) {
  const { token } = useAuth()
  const authFetcher = useMemo(() => createAuthFetcher(token), [token])

  const { data: records, isLoading, error, mutate } = useSWR<TrainingRecord[]>(
    token ? `${API_BASE_URL}/api/training/records?start=${date}&end=${date}` : null,
    authFetcher,
  )

  const todayRecords = records?.filter((r) => r.recorded_at.startsWith(date)) ?? []

  if (isLoading) {
    return <div className={styles.loading}>読み込み中...</div>
  }

  if (error) {
    return (
      <div className={styles.error}>
        <p>トレーニング記録の取得に失敗しました</p>
        <button onClick={() => mutate()} className={styles.retryButton}>
          再試行
        </button>
      </div>
    )
  }

  return (
    <section className={styles.section}>
      <div className={styles.header}>
        <h2 className={styles.title}>トレーニング</h2>
        <Link href="/training" className={styles.linkButton}>
          記録する
        </Link>
      </div>

      {todayRecords.length > 0 ? (
        <div className={styles.recordList}>
          {todayRecords.map((record) => (
            <div key={record.id} className={styles.recordItem}>
              <div className={styles.recordInfo}>
                {record.location_name && (
                  <span className={styles.location}>{record.location_name}</span>
                )}
                {record.duration && (
                  <span className={styles.duration}>{record.duration}分</span>
                )}
                {record.menus && record.menus.length > 0 && (
                  <span className={styles.menus}>
                    {record.menus.map((m) => m.name).join('・')}
                  </span>
                )}
              </div>
              <div className={styles.recordMeta}>
                {record.intensity && (
                  <span className={styles.rating}>強度: {'★'.repeat(record.intensity)}</span>
                )}
                {record.satisfaction && (
                  <span className={styles.rating}>満足度: {'★'.repeat(record.satisfaction)}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <p className={styles.noRecord}>本日の記録はありません</p>
      )}
    </section>
  )
}
