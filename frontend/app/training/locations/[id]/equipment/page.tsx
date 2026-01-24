'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useTrainingLocations, useTrainingEquipment, useTrainingMenu } from '@/hooks/useTraining'
import type { TrainingEquipment as EquipmentType } from '@/types/training'
import styles from './page.module.css'

export default function EquipmentPage() {
  const params = useParams()
  const locationId = params.id as string

  const { locations } = useTrainingLocations()
  const location = locations.find((l) => l.id === locationId)

  const { equipment, isLoading, error, createEquipment, updateEquipment, deleteEquipment } =
    useTrainingEquipment(locationId)
  const { normalizeEquipment } = useTrainingMenu()

  const [newName, setNewName] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState('')
  const [isCreating, setIsCreating] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [isNormalizing, setIsNormalizing] = useState(false)

  const handleCreate = async () => {
    if (!newName.trim()) return

    setIsCreating(true)
    try {
      await createEquipment({ name: newName.trim() })
      setNewName('')
    } catch (err) {
      alert(err instanceof Error ? err.message : '作成に失敗しました')
    } finally {
      setIsCreating(false)
    }
  }

  const handleNormalizeAndCreate = async () => {
    if (!newName.trim()) return

    setIsNormalizing(true)
    try {
      const result = await normalizeEquipment([newName.trim()])
      if (result.normalized_names.length > 0) {
        const normalized = result.normalized_names[0]
        await createEquipment({
          name: normalized.normalized,
          original_name: normalized.original !== normalized.normalized ? normalized.original : undefined,
        })
        setNewName('')
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : '正規化に失敗しました')
    } finally {
      setIsNormalizing(false)
    }
  }

  const handleStartEdit = (eq: EquipmentType) => {
    setEditingId(eq.id)
    setEditingName(eq.name)
  }

  const handleCancelEdit = () => {
    setEditingId(null)
    setEditingName('')
  }

  const handleSaveEdit = async () => {
    if (!editingId || !editingName.trim()) return

    setIsSaving(true)
    try {
      await updateEquipment(editingId, { name: editingName.trim() })
      setEditingId(null)
      setEditingName('')
    } catch (err) {
      alert(err instanceof Error ? err.message : '更新に失敗しました')
    } finally {
      setIsSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('この器具を削除しますか？')) return

    try {
      await deleteEquipment(id)
    } catch (err) {
      alert(err instanceof Error ? err.message : '削除に失敗しました')
    }
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <div className={styles.titleSection}>
            <h1 className={styles.title}>器具設定</h1>
            {location && <p className={styles.subtitle}>{location.name}</p>}
          </div>
          <Link href="/training/locations" className={styles.backButton}>
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
            placeholder="新しい器具の名前"
            className={styles.input}
            disabled={isCreating || isNormalizing}
          />
          <div className={styles.createActions}>
            <button
              type="button"
              onClick={handleCreate}
              disabled={isCreating || isNormalizing || !newName.trim()}
              className={styles.createButton}
            >
              {isCreating ? '追加中...' : '追加'}
            </button>
            <button
              type="button"
              onClick={handleNormalizeAndCreate}
              disabled={isCreating || isNormalizing || !newName.trim()}
              className={styles.normalizeButton}
            >
              {isNormalizing ? 'AI処理中...' : 'AI正規化して追加'}
            </button>
          </div>
        </div>

        {isLoading ? (
          <div className={styles.loading}>読み込み中...</div>
        ) : equipment.length === 0 ? (
          <div className={styles.empty}>
            <p>登録されている器具はありません</p>
          </div>
        ) : (
          <div className={styles.list}>
            {equipment.map((eq) => (
              <div key={eq.id} className={styles.equipmentCard}>
                {editingId === eq.id ? (
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
                    <div className={styles.equipmentInfo}>
                      <span className={styles.equipmentName}>{eq.name}</span>
                      {eq.original_name && (
                        <span className={styles.originalName}>元: {eq.original_name}</span>
                      )}
                    </div>
                    <div className={styles.equipmentActions}>
                      <button
                        type="button"
                        onClick={() => handleStartEdit(eq)}
                        className={styles.editButton}
                      >
                        編集
                      </button>
                      <button
                        type="button"
                        onClick={() => handleDelete(eq.id)}
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
