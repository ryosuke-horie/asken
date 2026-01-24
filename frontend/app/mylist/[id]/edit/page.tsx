'use client'

import { useState, useEffect } from 'react'
import { useRouter, useParams } from 'next/navigation'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import MylistForm from '@/components/client/MylistForm'
import { useMylist } from '@/hooks/useMylist'
import { useAuth } from '@/contexts/AuthContext'
import type { MylistItem, UpdateMylistItemRequest } from '@/types/mylist'
import styles from './page.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

export default function MylistEditPage() {
  const router = useRouter()
  const params = useParams()
  const id = params.id as string
  const { token } = useAuth()
  const { updateItem } = useMylist()

  const [item, setItem] = useState<MylistItem | null>(null)
  const [isLoadingItem, setIsLoadingItem] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id || !token) return

    const fetchItem = async () => {
      setIsLoadingItem(true)
      try {
        const response = await fetch(`${API_BASE_URL}/api/mylist/${id}`, {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        })

        if (!response.ok) {
          throw new Error('アイテムの取得に失敗しました')
        }

        const data = await response.json()
        setItem(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'エラーが発生しました')
      } finally {
        setIsLoadingItem(false)
      }
    }

    fetchItem()
  }, [id, token])

  const handleSubmit = async (data: UpdateMylistItemRequest) => {
    setIsSubmitting(true)
    setError(null)

    try {
      await updateItem(id, data)
      router.push('/mylist')
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <Link href="/mylist" className={styles.backLink}>
            ← マイリスト
          </Link>
          <h1 className={styles.title}>編集</h1>
        </div>

        {error && <div className={styles.error}>{error}</div>}

        {isLoadingItem ? (
          <div className={styles.loading}>読み込み中...</div>
        ) : item ? (
          <div className={styles.formSection}>
            <MylistForm
              initialData={{
                name: item.name,
                base_amount: item.base_amount,
                unit: item.unit,
                calories: item.calories,
                protein: item.protein,
                fat: item.fat,
                carbohydrates: item.carbohydrates,
                foods: item.foods,
              }}
              onSubmit={handleSubmit}
              isSubmitting={isSubmitting}
              submitLabel="更新する"
            />
          </div>
        ) : (
          <div className={styles.notFound}>アイテムが見つかりませんでした</div>
        )}
      </div>
    </ProtectedRoute>
  )
}
