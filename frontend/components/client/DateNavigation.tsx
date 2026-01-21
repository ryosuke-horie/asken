'use client'

import { useRouter } from 'next/navigation'
import styles from './DateNavigation.module.css'

interface DateNavigationProps {
  currentDate: string // YYYY-MM-DD
}

function getWeekDates(dateStr: string): Date[] {
  const date = new Date(dateStr)
  const dayOfWeek = date.getDay()
  // 日曜日(0)の場合は前週の月曜日、それ以外は当週の月曜日を計算
  const mondayOffset = dayOfWeek === 0 ? -6 : 1 - dayOfWeek

  const monday = new Date(date)
  monday.setDate(date.getDate() + mondayOffset)

  const weekDates: Date[] = []
  for (let i = 0; i < 7; i++) {
    const d = new Date(monday)
    d.setDate(monday.getDate() + i)
    weekDates.push(d)
  }
  return weekDates
}

function formatDateToYYYYMMDD(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function isSameDate(date1: Date, date2: Date): boolean {
  return (
    date1.getFullYear() === date2.getFullYear() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getDate() === date2.getDate()
  )
}

const WEEKDAY_LABELS = ['月', '火', '水', '木', '金', '土', '日']

export default function DateNavigation({ currentDate }: DateNavigationProps) {
  const router = useRouter()
  const weekDates = getWeekDates(currentDate)
  const today = new Date()
  const selectedDate = new Date(currentDate)

  const handleDateClick = (date: Date) => {
    router.push(`/?date=${formatDateToYYYYMMDD(date)}`)
  }

  return (
    <div className={styles.container}>
      <div className={styles.weekdayRow}>
        {WEEKDAY_LABELS.map((label) => (
          <span key={label} className={styles.weekdayLabel}>
            {label}
          </span>
        ))}
      </div>
      <div className={styles.datesRow}>
        {weekDates.map((date) => {
          const isToday = isSameDate(date, today)
          const isSelected = isSameDate(date, selectedDate)
          const classNames = [
            styles.dayButton,
            isToday && styles.today,
            isSelected && styles.selected,
          ]
            .filter(Boolean)
            .join(' ')

          return (
            <button
              key={date.toISOString()}
              className={classNames}
              onClick={() => handleDateClick(date)}
              aria-label={`${date.getDate()}日`}
            >
              {date.getDate()}
            </button>
          )
        })}
      </div>
    </div>
  )
}
