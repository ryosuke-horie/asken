'use client'

import { useRouter } from 'next/navigation'
import { useState, MouseEvent } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import styles from './DeleteHistoryButton.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

interface DeleteHistoryButtonProps {
  historyId: string
  iconOnly?: boolean
  onSuccess?: () => void
}

export default function DeleteHistoryButton({
  historyId,
  iconOnly = false,
  onSuccess,
}: DeleteHistoryButtonProps) {
  const router = useRouter()
  const { token } = useAuth()
  const [isDeleting, setIsDeleting] = useState(false)

  const handleDelete = async (e: MouseEvent<HTMLButtonElement>) => {
    e.preventDefault()
    e.stopPropagation()

    if (!confirm('この履歴を削除してもよろしいですか?')) {
      return
    }

    setIsDeleting(true)

    try {
      if (!token) {
        alert('認証が必要です。再度ログインしてください。')
        setIsDeleting(false)
        return
      }

      const response = await fetch(`${API_BASE_URL}/api/history/${historyId}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })

      if (!response.ok) {
        const errorText = await response.text().catch(() => '')
        const errorMessage = errorText || `削除に失敗しました (${response.status})`
        throw new Error(errorMessage)
      }

      if (onSuccess) {
        onSuccess()
      } else {
        router.push('/')
        router.refresh()
      }
    } catch (error) {
      console.error('削除エラー:', error)
      const message = error instanceof Error ? error.message : '削除に失敗しました'
      alert(message)
      setIsDeleting(false)
    }
  }

  if (iconOnly) {
    return (
      <button
        onClick={handleDelete}
        disabled={isDeleting}
        className={styles.iconButton}
        title="削除"
        aria-label="削除"
      >
        {isDeleting ? (
          <span className={styles.spinner} />
        ) : (
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="3 6 5 6 21 6" />
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
            <line x1="10" y1="11" x2="10" y2="17" />
            <line x1="14" y1="11" x2="14" y2="17" />
          </svg>
        )}
      </button>
    )
  }

  return (
    <button onClick={handleDelete} disabled={isDeleting} className={styles.button}>
      {isDeleting ? '削除中...' : '削除'}
    </button>
  )
}
