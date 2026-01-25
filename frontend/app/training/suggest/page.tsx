'use client'

import { useState, useMemo, useEffect, useCallback } from 'react'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useTrainingLocations, useTrainingMenu } from '@/hooks/useTraining'
import { useTodayCondition } from '@/hooks/useCondition'
import { useProfile } from '@/hooks/useProfile'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import { TRAINING_GOAL_LABELS } from '@/types/profile'
import type { MenuItem, TrainingEquipment } from '@/types/training'
import type { TrainingGoal } from '@/types/profile'
import styles from './page.module.css'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

const FATIGUE_LABELS: Record<number, string> = {
  1: '低い',
  2: '普通',
  3: '高い',
}

const CONDITION_LABELS: Record<number, string> = {
  1: '悪い',
  2: '普通',
  3: '良い',
}

export default function SuggestMenuPage() {
  const { token } = useAuth()
  const { locations } = useTrainingLocations()
  const { suggestMenu } = useTrainingMenu()
  const { record: conditionRecord, isLoading: conditionLoading } = useTodayCondition()
  const { profile, isLoading: profileLoading } = useProfile()

  const [selectedLocationIds, setSelectedLocationIds] = useState<Set<string>>(new Set())
  const [equipmentByLocation, setEquipmentByLocation] = useState<
    Record<string, TrainingEquipment[]>
  >({})
  const [selectedEquipment, setSelectedEquipment] = useState<Set<string>>(new Set())
  const [duration, setDuration] = useState(60)
  const [goals, setGoals] = useState('')
  const [menu, setMenu] = useState<MenuItem[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [equipmentFetchError, setEquipmentFetchError] = useState<string | null>(null)

  // プロフィールから目標を初期設定
  useEffect(() => {
    if (profile?.training_goals && profile.training_goals.length > 0 && goals === '') {
      const goalLabels = profile.training_goals
        .map((g) => TRAINING_GOAL_LABELS[g as TrainingGoal])
        .filter(Boolean)
      setGoals(goalLabels.join(', '))
    }
  }, [profile, goals])

  // 選択された場所の器具を取得
  const fetchEquipment = useCallback(
    async (locationId: string): Promise<TrainingEquipment[]> => {
      if (!token) return []
      try {
        setEquipmentFetchError(null)
        const authFetcher = createAuthFetcher(token)
        const data = await authFetcher(
          `${API_BASE_URL}/api/training/locations/${locationId}/equipment`,
        )
        return data as TrainingEquipment[]
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : '器具の取得に失敗しました'
        setEquipmentFetchError(errorMessage)
        return []
      }
    },
    [token],
  )

  // 場所選択時に器具を取得
  const handleToggleLocation = async (locationId: string) => {
    const newSelectedIds = new Set(selectedLocationIds)

    if (newSelectedIds.has(locationId)) {
      newSelectedIds.delete(locationId)
      // 選択解除した場所の器具を除外
      const newEquipmentByLocation = { ...equipmentByLocation }
      delete newEquipmentByLocation[locationId]
      setEquipmentByLocation(newEquipmentByLocation)

      // その場所の器具の選択も解除
      const removedEquipmentNames = (equipmentByLocation[locationId] || []).map((e) => e.name)
      setSelectedEquipment((prev) => {
        const next = new Set(prev)
        removedEquipmentNames.forEach((name) => next.delete(name))
        return next
      })
    } else {
      newSelectedIds.add(locationId)
      // 新しい場所の器具を取得
      const equipment = await fetchEquipment(locationId)
      setEquipmentByLocation((prev) => ({
        ...prev,
        [locationId]: equipment,
      }))
    }

    setSelectedLocationIds(newSelectedIds)
  }

  // 全器具を集約
  const allEquipment = useMemo(() => {
    const equipmentMap = new Map<string, TrainingEquipment>()
    Object.values(equipmentByLocation).forEach((eqList) => {
      eqList.forEach((eq) => {
        if (!equipmentMap.has(eq.name)) {
          equipmentMap.set(eq.name, eq)
        }
      })
    })
    return Array.from(equipmentMap.values())
  }, [equipmentByLocation])

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
    setSelectedEquipment(new Set(allEquipment.map((e) => e.name)))
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

        {/* 本日の体調表示 */}
        {!conditionLoading && conditionRecord && (
          <div className={styles.conditionInfo}>
            <div className={styles.conditionHeader}>本日の体調（自動適用）</div>
            <div className={styles.conditionItems}>
              <div className={styles.conditionItem}>
                <span className={styles.conditionLabel}>体調</span>
                <span className={styles.conditionValue}>
                  {CONDITION_LABELS[conditionRecord.condition] || '-'}
                </span>
              </div>
              <div className={styles.conditionItem}>
                <span className={styles.conditionLabel}>疲労度</span>
                <span className={styles.conditionValue}>
                  {FATIGUE_LABELS[conditionRecord.fatigue] || '-'}
                </span>
              </div>
            </div>
            <p className={styles.conditionNote}>この情報に基づいてメニューが調整されます</p>
          </div>
        )}

        {!conditionLoading && !conditionRecord && (
          <div className={styles.conditionInfoEmpty}>
            <p>本日の体調記録がありません。体調を記録するとより適切なメニューが提案されます。</p>
          </div>
        )}

        <div className={styles.form}>
          <div className={styles.formGroup}>
            <label className={styles.label}>場所を選択（複数可）</label>
            {locations.length === 0 ? (
              <p className={styles.noLocations}>
                登録されている場所がありません。
                <Link href="/training/locations" className={styles.link}>
                  場所を追加
                </Link>
                してください。
              </p>
            ) : (
              <div className={styles.locationGrid}>
                {locations.map((loc) => (
                  <label key={loc.id} className={styles.locationItem}>
                    <input
                      type="checkbox"
                      checked={selectedLocationIds.has(loc.id)}
                      onChange={() => handleToggleLocation(loc.id)}
                      className={styles.checkbox}
                    />
                    <span>{loc.name}</span>
                  </label>
                ))}
              </div>
            )}
          </div>

          {selectedLocationIds.size > 0 && (
            <div className={styles.formGroup}>
              <div className={styles.labelRow}>
                <label className={styles.label}>使用する器具</label>
                <div className={styles.selectActions}>
                  <button
                    type="button"
                    onClick={handleSelectAll}
                    className={styles.selectAllButton}
                  >
                    全選択
                  </button>
                  <button type="button" onClick={handleClearAll} className={styles.clearAllButton}>
                    全解除
                  </button>
                </div>
              </div>
              {equipmentFetchError && <div className={styles.warning}>{equipmentFetchError}</div>}
              {allEquipment.length === 0 && !equipmentFetchError && (
                <p className={styles.noEquipment}>選択した場所に登録されている器具がありません</p>
              )}
              {allEquipment.length > 0 && (
                <div className={styles.equipmentGrid}>
                  {allEquipment.map((eq) => (
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
            <label className={styles.label}>
              目標・重点（任意、カンマ区切り）
              {!profileLoading && profile?.training_goals && profile.training_goals.length > 0 && (
                <span className={styles.labelHint}>※プロフィールから取得</span>
              )}
            </label>
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
