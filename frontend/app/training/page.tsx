'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useTrainingLocations, useTrainingRecords } from '@/hooks/useTraining'
import styles from './page.module.css'

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0]
}

function getMonthDates(year: number, month: number): Date[] {
  const dates: Date[] = []
  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)

  for (let d = 1; d <= lastDay.getDate(); d++) {
    dates.push(new Date(year, month, d))
  }

  // 月初の曜日に合わせて空白を追加
  const startDayOfWeek = firstDay.getDay()
  for (let i = 0; i < startDayOfWeek; i++) {
    dates.unshift(new Date(year, month, -i))
  }

  return dates
}

export default function TrainingPage() {
  const today = new Date()
  const [currentMonth, setCurrentMonth] = useState({ year: today.getFullYear(), month: today.getMonth() })
  const [selectedDate, setSelectedDate] = useState<string>(formatDate(today))
  const [selectedLocationId, setSelectedLocationId] = useState<string | undefined>(undefined)
  const [completed, setCompleted] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  const { locations } = useTrainingLocations()

  const startDate = formatDate(new Date(currentMonth.year, currentMonth.month, 1))
  const endDate = formatDate(new Date(currentMonth.year, currentMonth.month + 1, 0))
  const { records, isLoading, error, upsertRecord } = useTrainingRecords(startDate, endDate)

  const recordsByDate = useMemo(() => {
    const map = new Map<string, (typeof records)[0]>()
    records.forEach((r) => {
      const dateStr = r.recorded_at.split('T')[0]
      map.set(dateStr, r)
    })
    return map
  }, [records])

  const monthDates = useMemo(() => {
    return getMonthDates(currentMonth.year, currentMonth.month)
  }, [currentMonth])

  const handlePrevMonth = () => {
    setCurrentMonth((prev) => {
      const newMonth = prev.month - 1
      if (newMonth < 0) {
        return { year: prev.year - 1, month: 11 }
      }
      return { ...prev, month: newMonth }
    })
  }

  const handleNextMonth = () => {
    setCurrentMonth((prev) => {
      const newMonth = prev.month + 1
      if (newMonth > 11) {
        return { year: prev.year + 1, month: 0 }
      }
      return { ...prev, month: newMonth }
    })
  }

  const handleDateSelect = (date: Date) => {
    const dateStr = formatDate(date)
    setSelectedDate(dateStr)
    const record = recordsByDate.get(dateStr)
    if (record) {
      setSelectedLocationId(record.location_id || undefined)
      setCompleted(record.completed)
    } else {
      setSelectedLocationId(undefined)
      setCompleted(false)
    }
  }

  const handleSave = async () => {
    setIsSaving(true)
    try {
      await upsertRecord({
        recorded_at: selectedDate,
        location_id: selectedLocationId,
        completed,
      })
    } catch (err) {
      alert(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setIsSaving(false)
    }
  }

  const weekDays = ['日', '月', '火', '水', '木', '金', '土']

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>トレーニング記録</h1>
          <div className={styles.headerActions}>
            <Link href="/training/locations" className={styles.settingsButton}>
              場所設定
            </Link>
            <Link href="/training/suggest" className={styles.suggestButton}>
              メニュー提案
            </Link>
          </div>
        </div>

        {error && (
          <div className={styles.error}>{error instanceof Error ? error.message : 'エラーが発生しました'}</div>
        )}

        <div className={styles.calendarContainer}>
          <div className={styles.calendarHeader}>
            <button type="button" onClick={handlePrevMonth} className={styles.navButton}>
              ←
            </button>
            <span className={styles.monthLabel}>
              {currentMonth.year}年{currentMonth.month + 1}月
            </span>
            <button type="button" onClick={handleNextMonth} className={styles.navButton}>
              →
            </button>
          </div>

          <div className={styles.weekDays}>
            {weekDays.map((day, i) => (
              <div key={day} className={`${styles.weekDay} ${i === 0 ? styles.sunday : i === 6 ? styles.saturday : ''}`}>
                {day}
              </div>
            ))}
          </div>

          {isLoading ? (
            <div className={styles.loading}>読み込み中...</div>
          ) : (
            <div className={styles.calendar}>
              {monthDates.map((date, i) => {
                const dateStr = formatDate(date)
                const isCurrentMonth = date.getMonth() === currentMonth.month
                const record = recordsByDate.get(dateStr)
                const isSelected = dateStr === selectedDate
                const isToday = dateStr === formatDate(today)
                const dayOfWeek = date.getDay()

                return (
                  <button
                    key={i}
                    type="button"
                    className={`${styles.calendarDay}
                      ${!isCurrentMonth ? styles.otherMonth : ''}
                      ${isSelected ? styles.selected : ''}
                      ${isToday ? styles.today : ''}
                      ${dayOfWeek === 0 ? styles.sunday : dayOfWeek === 6 ? styles.saturday : ''}`}
                    onClick={() => handleDateSelect(date)}
                    disabled={!isCurrentMonth}
                  >
                    <span className={styles.dayNumber}>{date.getDate()}</span>
                    {record && (
                      <span className={`${styles.recordIndicator} ${record.completed ? styles.completed : styles.notCompleted}`}>
                        {record.completed ? '○' : '×'}
                      </span>
                    )}
                  </button>
                )
              })}
            </div>
          )}
        </div>

        <div className={styles.recordForm}>
          <h2 className={styles.formTitle}>{selectedDate} の記録</h2>

          <div className={styles.formGroup}>
            <label className={styles.label}>場所</label>
            <select
              className={styles.select}
              value={selectedLocationId || ''}
              onChange={(e) => setSelectedLocationId(e.target.value || undefined)}
            >
              <option value="">選択してください</option>
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id}>
                  {loc.name}
                </option>
              ))}
            </select>
          </div>

          <div className={styles.formGroup}>
            <label className={styles.checkboxLabel}>
              <input
                type="checkbox"
                checked={completed}
                onChange={(e) => setCompleted(e.target.checked)}
                className={styles.checkbox}
              />
              練習を実施した
            </label>
          </div>

          <button
            type="button"
            onClick={handleSave}
            disabled={isSaving}
            className={styles.saveButton}
          >
            {isSaving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>
    </ProtectedRoute>
  )
}
