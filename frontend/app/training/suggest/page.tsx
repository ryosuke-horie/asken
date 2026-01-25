'use client'

import { useState, useMemo } from 'react'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useTrainingLocations, useTrainingEquipment, useTrainingMenu } from '@/hooks/useTraining'
import type { MenuItem } from '@/types/training'
import styles from './page.module.css'

export default function SuggestMenuPage() {
  const { locations } = useTrainingLocations()
  const [selectedLocationId, setSelectedLocationId] = useState<string | null>(null)
  const { equipment } = useTrainingEquipment(selectedLocationId)
  const { suggestMenu } = useTrainingMenu()

  const [selectedEquipment, setSelectedEquipment] = useState<Set<string>>(new Set())
  const [duration, setDuration] = useState(60)
  const [goals, setGoals] = useState('')
  const [menu, setMenu] = useState<MenuItem[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleToggleEquipment = (name: string) => {
    setSelectedEquipment((prev) => {
      const next = new Set(prev)
      if (next.has(name)) {
        next.delete(name)
      } else {
        next.add(name)
      }
      return next
    })
  }

  const handleSelectAll = () => {
    setSelectedEquipment(new Set(equipment.map((e) => e.name)))
  }

  const handleClearAll = () => {
    setSelectedEquipment(new Set())
  }

  const handleSuggest = async () => {
    if (selectedEquipment.size === 0) {
      setError('器具を1つ以上選択してください')
      return
    }

    setIsLoading(true)
    setError(null)
    setMenu([])

    try {
      const goalsArray = goals
        .split(',')
        .map((g) => g.trim())
        .filter((g) => g)
      const result = await suggestMenu({
        equipment: Array.from(selectedEquipment),
        duration,
        goals: goalsArray.length > 0 ? goalsArray : undefined,
      })
      setMenu(result.menu)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'メニュー提案に失敗しました')
    } finally {
      setIsLoading(false)
    }
  }

  const totalDuration = useMemo(() => {
    return menu.reduce((sum, item) => sum + item.duration, 0)
  }, [menu])

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>メニュー提案</h1>
          <Link href="/training" className={styles.backButton}>
            ← 戻る
          </Link>
        </div>

        {error && <div className={styles.error}>{error}</div>}

        <div className={styles.form}>
          <div className={styles.formGroup}>
            <label className={styles.label}>場所を選択</label>
            <select
              className={styles.select}
              value={selectedLocationId || ''}
              onChange={(e) => {
                setSelectedLocationId(e.target.value || null)
                setSelectedEquipment(new Set())
              }}
            >
              <option value="">場所を選択してください</option>
              {locations.map((loc) => (
                <option key={loc.id} value={loc.id}>
                  {loc.name}
                </option>
              ))}
            </select>
          </div>

          {selectedLocationId && (
            <div className={styles.formGroup}>
              <div className={styles.labelRow}>
                <label className={styles.label}>使用する器具</label>
                <div className={styles.selectActions}>
                  <button type="button" onClick={handleSelectAll} className={styles.selectAllButton}>
                    全選択
                  </button>
                  <button type="button" onClick={handleClearAll} className={styles.clearAllButton}>
                    全解除
                  </button>
                </div>
              </div>
              {equipment.length === 0 ? (
                <p className={styles.noEquipment}>登録されている器具がありません</p>
              ) : (
                <div className={styles.equipmentGrid}>
                  {equipment.map((eq) => (
                    <label key={eq.id} className={styles.equipmentItem}>
                      <input
                        type="checkbox"
                        checked={selectedEquipment.has(eq.name)}
                        onChange={() => handleToggleEquipment(eq.name)}
                        className={styles.checkbox}
                      />
                      <span>{eq.name}</span>
                    </label>
                  ))}
                </div>
              )}
            </div>
          )}

          <div className={styles.formGroup}>
            <label className={styles.label}>トレーニング時間（分）</label>
            <input
              type="number"
              value={duration}
              onChange={(e) => setDuration(Number(e.target.value))}
              min={10}
              max={180}
              className={styles.input}
            />
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label}>目標・重点（任意、カンマ区切り）</label>
            <input
              type="text"
              value={goals}
              onChange={(e) => setGoals(e.target.value)}
              placeholder="例: スタミナ強化, 打撃の強化"
              className={styles.input}
            />
          </div>

          <button
            type="button"
            onClick={handleSuggest}
            disabled={isLoading || selectedEquipment.size === 0}
            className={styles.suggestButton}
          >
            {isLoading ? 'AIが考え中...' : 'メニューを提案'}
          </button>
        </div>

        {menu.length > 0 && (
          <div className={styles.menuSection}>
            <div className={styles.menuHeader}>
              <h2 className={styles.menuTitle}>提案メニュー</h2>
              <span className={styles.totalDuration}>合計: {totalDuration}分</span>
            </div>
            <div className={styles.menuList}>
              {menu.map((item, index) => (
                <div key={index} className={styles.menuItem}>
                  <div className={styles.menuItemHeader}>
                    <span className={styles.menuItemNumber}>{index + 1}</span>
                    <h3 className={styles.menuItemName}>{item.name}</h3>
                    <span className={styles.menuItemDuration}>{item.duration}分</span>
                  </div>
                  <div className={styles.menuItemDetails}>
                    <div className={styles.menuItemMeta}>
                      <span className={styles.menuItemSets}>
                        {item.sets}セット × {item.reps}
                      </span>
                      <span className={styles.menuItemEquipment}>{item.equipment}</span>
                    </div>
                    <p className={styles.menuItemDescription}>{item.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </ProtectedRoute>
  )
}
