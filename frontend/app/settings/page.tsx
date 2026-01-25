'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useProfile } from '@/hooks/useProfile'
import { SPORT_TYPES, TRAINING_GOALS, TRAINING_GOAL_LABELS } from '@/types/profile'
import type { TrainingGoal } from '@/types/profile'
import styles from './page.module.css'

export default function SettingsPage() {
  const { profile, isLoading, error, updateProfile } = useProfile()
  const [sportType, setSportType] = useState<string>('')
  const [trainingGoals, setTrainingGoals] = useState<string[]>([])
  const [weightClass, setWeightClass] = useState<string>('')
  const [isSaving, setIsSaving] = useState(false)
  const [saveMessage, setSaveMessage] = useState<string | null>(null)

  useEffect(() => {
    if (profile) {
      setSportType(profile.sport_type || '')
      setTrainingGoals(profile.training_goals || [])
      setWeightClass(profile.weight_class ? String(profile.weight_class) : '')
    }
  }, [profile])

  const handleToggleGoal = (goal: string) => {
    setTrainingGoals((prev) =>
      prev.includes(goal) ? prev.filter((g) => g !== goal) : [...prev, goal],
    )
  }

  const handleSave = async () => {
    setIsSaving(true)
    setSaveMessage(null)

    try {
      await updateProfile({
        sport_type: sportType || null,
        training_goals: trainingGoals,
        weight_class: weightClass ? parseInt(weightClass, 10) : null,
      })
      setSaveMessage('保存しました')
      setTimeout(() => setSaveMessage(null), 3000)
    } catch (err) {
      setSaveMessage(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>設定</h1>
          <Link href="/" className={styles.backButton}>
            ← 戻る
          </Link>
        </div>

        {error && (
          <div className={styles.error}>
            {error instanceof Error ? error.message : 'エラーが発生しました'}
          </div>
        )}

        {isLoading ? (
          <div className={styles.loading}>読み込み中...</div>
        ) : (
          <div className={styles.content}>
            <section className={styles.section}>
              <h2 className={styles.sectionTitle}>プロフィール</h2>

              <div className={styles.field}>
                <label className={styles.label}>競技種別</label>
                <select
                  value={sportType}
                  onChange={(e) => setSportType(e.target.value)}
                  className={styles.select}
                  disabled={isSaving}
                >
                  <option value="">選択してください</option>
                  {SPORT_TYPES.map((type) => (
                    <option key={type} value={type}>
                      {type}
                    </option>
                  ))}
                </select>
              </div>

              <div className={styles.field}>
                <label className={styles.label}>トレーニング目標（複数選択可）</label>
                <div className={styles.checkboxGroup}>
                  {TRAINING_GOALS.map((goal) => (
                    <label key={goal} className={styles.checkboxLabel}>
                      <input
                        type="checkbox"
                        checked={trainingGoals.includes(goal)}
                        onChange={() => handleToggleGoal(goal)}
                        disabled={isSaving}
                        className={styles.checkbox}
                      />
                      <span>{TRAINING_GOAL_LABELS[goal as TrainingGoal]}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className={styles.field}>
                <label className={styles.label}>体重階級 (kg)</label>
                <input
                  type="number"
                  value={weightClass}
                  onChange={(e) => setWeightClass(e.target.value)}
                  placeholder="例: 65"
                  min="1"
                  max="200"
                  className={styles.input}
                  disabled={isSaving}
                />
              </div>

              <div className={styles.actions}>
                <button
                  type="button"
                  onClick={handleSave}
                  disabled={isSaving}
                  className={styles.saveButton}
                >
                  {isSaving ? '保存中...' : '保存'}
                </button>
                {saveMessage && (
                  <span
                    className={
                      saveMessage.includes('失敗') ? styles.errorMessage : styles.successMessage
                    }
                  >
                    {saveMessage}
                  </span>
                )}
              </div>
            </section>
          </div>
        )}
      </div>
    </ProtectedRoute>
  )
}
