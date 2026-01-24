'use client'

import { useState, useEffect, useRef } from 'react'
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
  const [isUploading, setIsUploading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [imagePath, setImagePath] = useState<string | undefined>(undefined)
  const [imagePreview, setImagePreview] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

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
        if (data.image_path) {
          setImagePath(data.image_path)
          const filename = data.image_path.split('/').pop()
          setImagePreview(`${API_BASE_URL}/api/images/${filename}`)
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'エラーが発生しました')
      } finally {
        setIsLoadingItem(false)
      }
    }

    fetchItem()
  }, [id, token])

  const handleImageSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file || !token) return

    const reader = new FileReader()
    reader.onload = (e) => {
      setImagePreview(e.target?.result as string)
    }
    reader.readAsDataURL(file)

    setIsUploading(true)
    setError(null)

    try {
      const formData = new FormData()
      formData.append('image', file)

      const response = await fetch(`${API_BASE_URL}/api/upload-image`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      })

      if (!response.ok) {
        throw new Error('画像のアップロードに失敗しました')
      }

      const { image_path } = await response.json()
      setImagePath(image_path)
    } catch (err) {
      setError(err instanceof Error ? err.message : '画像のアップロードに失敗しました')
      if (item?.image_path) {
        const filename = item.image_path.split('/').pop()
        setImagePreview(`${API_BASE_URL}/api/images/${filename}`)
        setImagePath(item.image_path)
      } else {
        setImagePreview(null)
        setImagePath(undefined)
      }
    } finally {
      setIsUploading(false)
    }
  }

  const clearImage = () => {
    setImagePreview(null)
    setImagePath(undefined)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleSubmit = async (data: UpdateMylistItemRequest) => {
    setIsSubmitting(true)
    setError(null)

    try {
      const requestData = {
        ...data,
        image_path: imagePath,
      }
      await updateItem(id, requestData)
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
            <div className={styles.imageUploadSection}>
              <label className={styles.imageLabel}>画像</label>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                onChange={handleImageSelect}
                className={styles.fileInput}
                disabled={isUploading}
              />
              {!imagePreview ? (
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className={styles.fileButton}
                  disabled={isUploading}
                >
                  {isUploading ? 'アップロード中...' : '画像を選択'}
                </button>
              ) : (
                <div className={styles.imagePreview}>
                  <img src={imagePreview} alt="プレビュー" className={styles.previewImage} />
                  <button type="button" onClick={clearImage} className={styles.clearImageButton}>
                    画像を削除
                  </button>
                </div>
              )}
            </div>

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
