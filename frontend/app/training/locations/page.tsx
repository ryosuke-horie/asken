'use client'

import { useState } from 'react'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useTrainingLocations } from '@/hooks/useTraining'
import type { TrainingLocation } from '@/types/training'
import styles from './page.module.css'

export default function LocationsPage() {
  const { locations, isLoading, error, createLocation, updateLocation, deleteLocation } =
    useTrainingLocations()
  const [newName, setNewName] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  const handleCreate = async () => {
    if (!newName.trim()) return

    setIsCreating(true)
    try {
      await createLocation({ name: newName.trim() })
      setNewName('')
    } catch (err) {
      alert(err instanceof Error ? err.message : '作成に失敗しました')
    } finally {
      setIsCreating(false)
    }
  }

  const handleStartEdit = (location: TrainingLocation) => {
    setEditingId(location.id)
    setEditingName(location.name)
  }

  const handleCancelEdit = () => {
    setEditingId(null)
    setEditingName('')
  }

  const handleSaveEdit = async () => {
    if (!editingId || !editingName.trim()) return

    setIsSaving(true)
    try {
      await updateLocation(editingId, { name: editingName.trim() })
      setEditingId(null)
      setEditingName('')
    } catch (err) {
      alert(err instanceof Error ? err.message : '更新に失敗しました')
    } finally {
      setIsSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('この場所を削除しますか？関連する器具も削除されます。')) return

    try {
      await deleteLocation(id)
    } catch (err) {
      alert(err instanceof Error ? err.message : '削除に失敗しました')
    }
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>トレーニング場所設定</h1>
          <Link href="/training" className={styles.backButton}>
            ← 戻る
          </Link>
        </div>

        {error && (
          <div className={styles.error}>
            {error instanceof Error ? error.message : 'エラーが発生しました'}
          </div>
        )}

        <div className={styles.createForm}>
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="新しい場所の名前"
            className={styles.input}
            disabled={isCreating}
          />
          <button
            type="button"
            onClick={handleCreate}
            disabled={isCreating || !newName.trim()}
            className={styles.createButton}
          >
            {isCreating ? '作成中...' : '追加'}
          </button>
        </div>

        {isLoading ? (
          <div className={styles.loading}>読み込み中...</div>
        ) : locations.length === 0 ? (
          <div className={styles.empty}>
            <p>登録されている場所はありません</p>
          </div>
        ) : (
          <div className={styles.list}>
            {locations.map((location) => (
              <div key={location.id} className={styles.locationCard}>
                {editingId === location.id ? (
                  <div className={styles.editForm}>
                    <input
                      type="text"
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                      className={styles.editInput}
                      disabled={isSaving}
                    />
                    <div className={styles.editActions}>
                      <button
                        type="button"
                        onClick={handleSaveEdit}
                        disabled={isSaving || !editingName.trim()}
                        className={styles.saveButton}
                      >
                        {isSaving ? '保存中...' : '保存'}
                      </button>
                      <button
                        type="button"
                        onClick={handleCancelEdit}
                        disabled={isSaving}
                        className={styles.cancelButton}
                      >
                        キャンセル
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className={styles.locationInfo}>
                      <span className={styles.locationName}>{location.name}</span>
                    </div>
                    <div className={styles.locationActions}>
                      <Link
                        href={`/training/locations/${location.id}/equipment`}
                        className={styles.equipmentButton}
                      >
                        器具設定
                      </Link>
                      <button
                        type="button"
                        onClick={() => handleStartEdit(location)}
                        className={styles.editButton}
                      >
                        編集
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(location.id)}
                        className={styles.deleteButton}
                      >
                        削除
                      </button>
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </ProtectedRoute>
  )
}
