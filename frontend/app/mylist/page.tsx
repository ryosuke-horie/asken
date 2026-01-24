'use client'

import { useState } from 'react'
import Link from 'next/link'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import { SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy } from '@dnd-kit/sortable'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import SortableMylistItemCard from '@/components/client/SortableMylistItemCard'
import { useMylist } from '@/hooks/useMylist'
import type { MylistItem } from '@/types/mylist'
import styles from './page.module.css'

export default function MylistPage() {
  const { items, isLoading, error, deleteItem, reorderItems } = useMylist()
  const [isSaving, setIsSaving] = useState(false)

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const handleEdit = (item: MylistItem) => {
    window.location.href = `/mylist/${item.id}/edit`
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteItem(id)
    } catch (err) {
      alert(err instanceof Error ? err.message : '削除に失敗しました')
    }
  }

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event

    if (!over || active.id === over.id) return

    const oldIndex = items.findIndex((item) => item.id === active.id)
    const newIndex = items.findIndex((item) => item.id === over.id)

    if (oldIndex === -1 || newIndex === -1) return

    const newItems = [...items]
    const [removed] = newItems.splice(oldIndex, 1)
    newItems.splice(newIndex, 0, removed)

    const newOrder = newItems.map((item) => item.id)

    setIsSaving(true)
    try {
      await reorderItems(newOrder)
    } catch (err) {
      alert(err instanceof Error ? err.message : '並び替えに失敗しました')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>マイリスト</h1>
          <Link href="/mylist/new" className={styles.addButton}>
            + 新規登録
          </Link>
        </div>

        {error && (
          <div className={styles.error}>{error instanceof Error ? error.message : 'エラーが発生しました'}</div>
        )}

        {isSaving && <div className={styles.saving}>保存中...</div>}

        {isLoading ? (
          <div className={styles.loading}>読み込み中...</div>
        ) : items.length === 0 ? (
          <div className={styles.empty}>
            <p>登録されているアイテムはありません</p>
            <Link href="/mylist/new" className={styles.emptyLink}>
              最初のアイテムを登録する
            </Link>
          </div>
        ) : (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={items.map((item) => item.id)} strategy={verticalListSortingStrategy}>
              <div className={styles.list}>
                {items.map((item) => (
                  <SortableMylistItemCard key={item.id} item={item} onEdit={handleEdit} onDelete={handleDelete} />
                ))}
              </div>
            </SortableContext>
          </DndContext>
        )}
      </div>
    </ProtectedRoute>
  )
}
