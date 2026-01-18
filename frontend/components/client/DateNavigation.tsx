'use client'

import { useRouter } from 'next/navigation'
import styles from './DateNavigation.module.css'

interface DateNavigationProps {
  currentDate: string // YYYY-MM-DD
}

export default function DateNavigation({ currentDate }: DateNavigationProps) {
  const router = useRouter()

  const handlePreviousDay = () => {
    const date = new Date(currentDate)
    date.setDate(date.getDate() - 1)
    router.push(`/?date=${date.toISOString().split('T')[0]}`)
  }

  const handleNextDay = () => {
    const date = new Date(currentDate)
    date.setDate(date.getDate() + 1)
    router.push(`/?date=${date.toISOString().split('T')[0]}`)
  }

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    const month = date.getMonth() + 1
    const day = date.getDate()
    const weekday = ['日', '月', '火', '水', '木', '金', '土'][date.getDay()]
    return `${month}月${day}日（${weekday}）`
  }

  return (
    <div className={styles.container}>
      <button onClick={handlePreviousDay} className={styles.navButton}>
        ← 前日
      </button>
      <div className={styles.currentDate}>
        {formatDate(currentDate)}
      </div>
      <button onClick={handleNextDay} className={styles.navButton}>
        翌日 →
      </button>
    </div>
  )
}
