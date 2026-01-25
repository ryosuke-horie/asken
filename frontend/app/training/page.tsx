'use client'

import { useState, useMemo, useEffect } from 'react'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { useTrainingLocations, useTrainingRecords, useTrainingMenus, useTrainingEquipment } from '@/hooks/useTraining'
import type { TrainingRecord } from '@/types/training'
import styles from './page.module.css'

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0]
}

function getMonthDates(year: number, month: number): Date[] {
  const dates: Date[] = []
  const firstDay = new Date(year, month, 1)
  const lastDay = new Date(year, month + 1, 0)

  for (let d = 1; d <= lastDay.getDate(); d++) {
    dates.push(new Date(year, month, d))
  }

  const startDayOfWeek = firstDay.getDay()
  for (let i = 0; i < startDayOfWeek; i++) {
    dates.unshift(new Date(year, month, -i))
  }

  return dates
}

interface ExerciseDetail {
  id: string
  name: string
  type: 'menu' | 'equipment'
  sets: string
  reps: string
  checked: boolean
}

interface RecordFormData {
  locationId: string
  duration: string
  exercises: ExerciseDetail[]
  intensity: number
  satisfaction: number
  notes: string
}

const initialFormData: RecordFormData = {
  locationId: '',
  duration: '',
  exercises: [],
  intensity: 0,
  satisfaction: 0,
  notes: '',
}

function formatExerciseSummary(exercises: ExerciseDetail[]): string {
  const selected = exercises.filter((e) => e.checked)
  if (selected.length === 0) return ''

  return selected
    .map((e) => {
      const parts = [e.name]
      if (e.sets && e.reps) {
        parts.push(`${e.sets}セット×${e.reps}回`)
      } else if (e.sets) {
        parts.push(`${e.sets}セット`)
      } else if (e.reps) {
        parts.push(`${e.reps}回`)
      }
      return parts.join(' ')
    })
    .join('\n')
}

function parseExerciseSummary(notes: string, exercises: ExerciseDetail[]): ExerciseDetail[] {
  if (!notes) return exercises

  const lines = notes.split('\n')
  return exercises.map((exercise) => {
    const line = lines.find((l) => l.startsWith(exercise.name))
    if (!line) return exercise

    const match = line.match(/(\d+)セット[×x](\d+)回/)
    if (match) {
      return { ...exercise, checked: true, sets: match[1], reps: match[2] }
    }

    const setsMatch = line.match(/(\d+)セット/)
    const repsMatch = line.match(/(\d+)回/)

    return {
      ...exercise,
      checked: true,
      sets: setsMatch ? setsMatch[1] : '',
      reps: repsMatch ? repsMatch[1] : '',
    }
  })
}

interface ExercisePreview {
  name: string
  sets?: string
  reps?: string
}

function parseNotesToExercises(notes: string): { exercises: ExercisePreview[]; userNotes: string } {
  if (!notes) return { exercises: [], userNotes: '' }

  const lines = notes.split('\n')
  const exercises: ExercisePreview[] = []
  const userNotesLines: string[] = []
  let isUserNotes = false

  for (const line of lines) {
    if (line.trim() === '') {
      isUserNotes = true
      continue
    }

    if (isUserNotes) {
      userNotesLines.push(line)
      continue
    }

    // エクササイズ行かどうかをチェック（名前 + オプションでセット/回数）
    const match = line.match(/^(.+?)\s*(?:(\d+)セット)?(?:[×x](\d+)回)?$/)
    if (match && (match[2] || match[3])) {
      exercises.push({
        name: match[1].trim(),
        sets: match[2],
        reps: match[3],
      })
    } else if (line.match(/セット|回/)) {
      // セットまたは回が含まれる行はエクササイズとして扱う
      const setsMatch = line.match(/(\d+)セット/)
      const repsMatch = line.match(/(\d+)回/)
      const nameMatch = line.match(/^([^\d]+)/)
      if (nameMatch) {
        exercises.push({
          name: nameMatch[1].trim(),
          sets: setsMatch?.[1],
          reps: repsMatch?.[1],
        })
      }
    } else {
      // セット/回数がない行はユーザーメモとして扱う
      userNotesLines.push(line)
    }
  }

  return { exercises, userNotes: userNotesLines.join('\n') }
}

export default function TrainingPage() {
  const today = new Date()
  const [currentMonth, setCurrentMonth] = useState({ year: today.getFullYear(), month: today.getMonth() })
  const [selectedDate, setSelectedDate] = useState<string>(formatDate(today))
  const [formData, setFormData] = useState<RecordFormData>(initialFormData)
  const [editingRecordId, setEditingRecordId] = useState<string | null>(null)
  const [isSaving, setIsSaving] = useState(false)

  const { locations } = useTrainingLocations()
  const { menus } = useTrainingMenus()
  const { equipment } = useTrainingEquipment(formData.locationId || null)

  const startDate = formatDate(new Date(currentMonth.year, currentMonth.month, 1))
  const endDate = formatDate(new Date(currentMonth.year, currentMonth.month + 1, 0))
  const { records, isLoading, error, createRecord, updateRecord, deleteRecord } = useTrainingRecords(startDate, endDate)

  // メニューと器具を exercises にマージ
  useEffect(() => {
    const menuExercises: ExerciseDetail[] = menus.map((m) => ({
      id: m.id,
      name: m.name,
      type: 'menu' as const,
      sets: '',
      reps: '',
      checked: false,
    }))

    const equipmentExercises: ExerciseDetail[] = equipment.map((e) => ({
      id: e.id,
      name: e.name,
      type: 'equipment' as const,
      sets: '',
      reps: '',
      checked: false,
    }))

    setFormData((prev) => {
      // 既存の選択状態を保持
      const existingMap = new Map(prev.exercises.map((e) => [e.id, e]))
      const newExercises = [...menuExercises, ...equipmentExercises].map((exercise) => {
        const existing = existingMap.get(exercise.id)
        if (existing) {
          return { ...exercise, sets: existing.sets, reps: existing.reps, checked: existing.checked }
        }
        return exercise
      })
      return { ...prev, exercises: newExercises }
    })
  }, [menus, equipment])

  const recordsByDate = useMemo(() => {
    const map = new Map<string, TrainingRecord[]>()
    records.forEach((r) => {
      const dateStr = r.recorded_at.split('T')[0]
      const existing = map.get(dateStr) ?? []
      map.set(dateStr, [...existing, r])
    })
    return map
  }, [records])

  const selectedDateRecords = recordsByDate.get(selectedDate) ?? []

  const monthDates = useMemo(() => {
    return getMonthDates(currentMonth.year, currentMonth.month)
  }, [currentMonth])

  const handlePrevMonth = () => {
    setCurrentMonth((prev) => {
      const newMonth = prev.month - 1
      if (newMonth < 0) {
        return { year: prev.year - 1, month: 11 }
      }
      return { ...prev, month: newMonth }
    })
  }

  const handleNextMonth = () => {
    setCurrentMonth((prev) => {
      const newMonth = prev.month + 1
      if (newMonth > 11) {
        return { year: prev.year + 1, month: 0 }
      }
      return { ...prev, month: newMonth }
    })
  }

  const handleDateSelect = (date: Date) => {
    const dateStr = formatDate(date)
    setSelectedDate(dateStr)
    resetForm()
  }

  const resetForm = () => {
    setFormData((prev) => ({
      ...initialFormData,
      exercises: prev.exercises.map((e) => ({ ...e, checked: false, sets: '', reps: '' })),
    }))
    setEditingRecordId(null)
  }

  const handleEditRecord = (record: TrainingRecord) => {
    setEditingRecordId(record.id)

    // 既存のexercisesをリセットして、記録のメニューをチェック状態にする
    const menuIds = record.menus?.map((m) => m.id) ?? []

    setFormData((prev) => {
      const updatedExercises = prev.exercises.map((e) => {
        if (menuIds.includes(e.id)) {
          return { ...e, checked: true }
        }
        return { ...e, checked: false, sets: '', reps: '' }
      })

      // notesからセット数・回数を復元
      const parsedExercises = parseExerciseSummary(record.notes ?? '', updatedExercises)

      return {
        locationId: record.location_id ?? '',
        duration: record.duration?.toString() ?? '',
        exercises: parsedExercises,
        intensity: record.intensity ?? 0,
        satisfaction: record.satisfaction ?? 0,
        notes: record.notes ?? '',
      }
    })
  }

  const handleExerciseToggle = (exerciseId: string) => {
    setFormData((prev) => ({
      ...prev,
      exercises: prev.exercises.map((e) =>
        e.id === exerciseId ? { ...e, checked: !e.checked } : e
      ),
    }))
  }

  const handleExerciseChange = (exerciseId: string, field: 'sets' | 'reps', value: string) => {
    setFormData((prev) => ({
      ...prev,
      exercises: prev.exercises.map((e) =>
        e.id === exerciseId ? { ...e, [field]: value } : e
      ),
    }))
  }

  const handleSave = async () => {
    setIsSaving(true)
    try {
      const selectedMenuIds = formData.exercises
        .filter((e) => e.checked && e.type === 'menu')
        .map((e) => e.id)

      // 選択されたエクササイズの詳細をnotesに追加
      const exerciseSummary = formatExerciseSummary(formData.exercises)
      const combinedNotes = exerciseSummary
        ? `${exerciseSummary}${formData.notes ? '\n\n' + formData.notes : ''}`
        : formData.notes

      const payload = {
        recorded_at: selectedDate,
        location_id: formData.locationId || undefined,
        completed: true,
        duration: formData.duration ? parseInt(formData.duration, 10) : undefined,
        menu_ids: selectedMenuIds.length > 0 ? selectedMenuIds : undefined,
        intensity: formData.intensity > 0 ? formData.intensity : undefined,
        satisfaction: formData.satisfaction > 0 ? formData.satisfaction : undefined,
        notes: combinedNotes || undefined,
      }

      if (editingRecordId) {
        await updateRecord(editingRecordId, payload)
      } else {
        await createRecord(payload)
      }
      resetForm()
    } catch (err) {
      alert(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setIsSaving(false)
    }
  }

  const handleDelete = async (recordId: string) => {
    if (!confirm('この記録を削除しますか？')) return

    try {
      await deleteRecord(recordId)
      if (editingRecordId === recordId) {
        resetForm()
      }
    } catch (err) {
      alert(err instanceof Error ? err.message : '削除に失敗しました')
    }
  }

  const weekDays = ['日', '月', '火', '水', '木', '金', '土']

  const menuExercises = formData.exercises.filter((e) => e.type === 'menu')
  const equipmentExercises = formData.exercises.filter((e) => e.type === 'equipment')

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <h1 className={styles.title}>トレーニング記録</h1>
          <div className={styles.headerActions}>
            <Link href="/training/locations" className={styles.settingsButton}>
              場所設定
            </Link>
            <Link href="/training/suggest" className={styles.suggestButton}>
              メニュー提案
            </Link>
          </div>
        </div>

        {error && (
          <div className={styles.error}>{error instanceof Error ? error.message : 'エラーが発生しました'}</div>
        )}

        <div className={styles.calendarContainer}>
          <div className={styles.calendarHeader}>
            <button type="button" onClick={handlePrevMonth} className={styles.navButton}>
              ←
            </button>
            <span className={styles.monthLabel}>
              {currentMonth.year}年{currentMonth.month + 1}月
            </span>
            <button type="button" onClick={handleNextMonth} className={styles.navButton}>
              →
            </button>
          </div>

          <div className={styles.weekDays}>
            {weekDays.map((day, i) => (
              <div key={day} className={`${styles.weekDay} ${i === 0 ? styles.sunday : i === 6 ? styles.saturday : ''}`}>
                {day}
              </div>
            ))}
          </div>

          {isLoading ? (
            <div className={styles.loading}>読み込み中...</div>
          ) : (
            <div className={styles.calendar}>
              {monthDates.map((date, i) => {
                const dateStr = formatDate(date)
                const isCurrentMonth = date.getMonth() === currentMonth.month
                const dateRecords = recordsByDate.get(dateStr) ?? []
                const isSelected = dateStr === selectedDate
                const isToday = dateStr === formatDate(today)
                const dayOfWeek = date.getDay()

                return (
                  <button
                    key={i}
                    type="button"
                    className={`${styles.calendarDay}
                      ${!isCurrentMonth ? styles.otherMonth : ''}
                      ${isSelected ? styles.selected : ''}
                      ${isToday ? styles.today : ''}
                      ${dayOfWeek === 0 ? styles.sunday : dayOfWeek === 6 ? styles.saturday : ''}`}
                    onClick={() => handleDateSelect(date)}
                    disabled={!isCurrentMonth}
                  >
                    <span className={styles.dayNumber}>{date.getDate()}</span>
                    {dateRecords.length > 0 && (
                      <span className={styles.recordCount}>{dateRecords.length}</span>
                    )}
                  </button>
                )
              })}
            </div>
          )}
        </div>

        <div className={styles.recordSection}>
          <h2 className={styles.sectionTitle}>{selectedDate} の記録</h2>

          {selectedDateRecords.length > 0 && (
            <div className={styles.recordList}>
              {selectedDateRecords.map((record) => {
                const { exercises: parsedExercises, userNotes } = parseNotesToExercises(record.notes ?? '')

                return (
                  <div key={record.id} className={`${styles.recordCard} ${editingRecordId === record.id ? styles.editing : ''}`}>
                    <div className={styles.recordCardHeader}>
                      <div className={styles.recordInfo}>
                        {record.location_name && <span className={styles.recordLocation}>{record.location_name}</span>}
                        {record.duration && <span className={styles.recordDuration}>{record.duration}分</span>}
                      </div>
                      <div className={styles.recordActions}>
                        <button type="button" onClick={() => handleEditRecord(record)} className={styles.editButton}>
                          編集
                        </button>
                        <button type="button" onClick={() => handleDelete(record.id)} className={styles.deleteButton}>
                          削除
                        </button>
                      </div>
                    </div>

                    {parsedExercises.length > 0 && (
                      <div className={styles.exercisePreviewList}>
                        {parsedExercises.map((ex, idx) => (
                          <div key={idx} className={styles.exercisePreviewItem}>
                            <span className={styles.exercisePreviewName}>{ex.name}</span>
                            {(ex.sets || ex.reps) && (
                              <span className={styles.exercisePreviewDetail}>
                                {ex.sets && `${ex.sets}セット`}
                                {ex.sets && ex.reps && ' × '}
                                {ex.reps && `${ex.reps}回`}
                              </span>
                            )}
                          </div>
                        ))}
                      </div>
                    )}

                    <div className={styles.recordMeta}>
                      {record.intensity && record.intensity > 0 && (
                        <span className={styles.ratingDisplay}>強度: {'★'.repeat(record.intensity)}{'☆'.repeat(5 - record.intensity)}</span>
                      )}
                      {record.satisfaction && record.satisfaction > 0 && (
                        <span className={styles.ratingDisplay}>満足度: {'★'.repeat(record.satisfaction)}{'☆'.repeat(5 - record.satisfaction)}</span>
                      )}
                    </div>

                    {userNotes && <p className={styles.recordNotes}>{userNotes}</p>}
                  </div>
                )
              })}
            </div>
          )}

          <div className={styles.recordForm}>
            <h3 className={styles.formTitle}>{editingRecordId ? '記録を編集' : '新しい記録を追加'}</h3>

            <div className={styles.formGroup}>
              <label className={styles.label}>場所</label>
              <select
                className={styles.select}
                value={formData.locationId}
                onChange={(e) => setFormData({ ...formData, locationId: e.target.value })}
              >
                <option value="">選択してください</option>
                {locations.map((loc) => (
                  <option key={loc.id} value={loc.id}>
                    {loc.name}
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>練習時間（分）</label>
              <input
                type="number"
                className={styles.input}
                value={formData.duration}
                onChange={(e) => setFormData({ ...formData, duration: e.target.value })}
                placeholder="例: 60"
                min="0"
              />
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>練習メニュー</label>
              <div className={styles.exerciseList}>
                {menuExercises.map((exercise) => (
                  <div key={exercise.id} className={`${styles.exerciseItem} ${exercise.checked ? styles.checked : ''}`}>
                    <label className={styles.exerciseCheckboxLabel}>
                      <input
                        type="checkbox"
                        checked={exercise.checked}
                        onChange={() => handleExerciseToggle(exercise.id)}
                        className={styles.exerciseCheckbox}
                      />
                      <span className={styles.exerciseName}>{exercise.name}</span>
                    </label>
                    {exercise.checked && (
                      <div className={styles.exerciseInputs}>
                        <input
                          type="number"
                          className={styles.exerciseInput}
                          value={exercise.sets}
                          onChange={(e) => handleExerciseChange(exercise.id, 'sets', e.target.value)}
                          placeholder="セット"
                          min="0"
                        />
                        <span className={styles.exerciseInputLabel}>セット</span>
                        <input
                          type="number"
                          className={styles.exerciseInput}
                          value={exercise.reps}
                          onChange={(e) => handleExerciseChange(exercise.id, 'reps', e.target.value)}
                          placeholder="回数"
                          min="0"
                        />
                        <span className={styles.exerciseInputLabel}>回</span>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>

            {formData.locationId && equipmentExercises.length > 0 && (
              <div className={styles.formGroup}>
                <label className={styles.label}>器具（{locations.find((l) => l.id === formData.locationId)?.name}）</label>
                <div className={styles.exerciseList}>
                  {equipmentExercises.map((exercise) => (
                    <div key={exercise.id} className={`${styles.exerciseItem} ${exercise.checked ? styles.checked : ''}`}>
                      <label className={styles.exerciseCheckboxLabel}>
                        <input
                          type="checkbox"
                          checked={exercise.checked}
                          onChange={() => handleExerciseToggle(exercise.id)}
                          className={styles.exerciseCheckbox}
                        />
                        <span className={styles.exerciseName}>{exercise.name}</span>
                      </label>
                      {exercise.checked && (
                        <div className={styles.exerciseInputs}>
                          <input
                            type="number"
                            className={styles.exerciseInput}
                            value={exercise.sets}
                            onChange={(e) => handleExerciseChange(exercise.id, 'sets', e.target.value)}
                            placeholder="セット"
                            min="0"
                          />
                          <span className={styles.exerciseInputLabel}>セット</span>
                          <input
                            type="number"
                            className={styles.exerciseInput}
                            value={exercise.reps}
                            onChange={(e) => handleExerciseChange(exercise.id, 'reps', e.target.value)}
                            placeholder="回数"
                            min="0"
                          />
                          <span className={styles.exerciseInputLabel}>回</span>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {formData.locationId && equipmentExercises.length === 0 && (
              <div className={styles.formGroup}>
                <label className={styles.label}>器具</label>
                <p className={styles.noEquipment}>
                  この場所には器具が登録されていません。
                  <Link href={`/training/locations/${formData.locationId}/equipment`} className={styles.addEquipmentLink}>
                    器具を追加
                  </Link>
                </p>
              </div>
            )}

            <div className={styles.formGroup}>
              <label className={styles.label}>強度</label>
              <div className={styles.ratingGroup}>
                {[1, 2, 3, 4, 5].map((value) => (
                  <button
                    key={value}
                    type="button"
                    className={`${styles.ratingButton} ${formData.intensity >= value ? styles.active : ''}`}
                    onClick={() => setFormData({ ...formData, intensity: formData.intensity === value ? 0 : value })}
                  >
                    ★
                  </button>
                ))}
              </div>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>満足度</label>
              <div className={styles.ratingGroup}>
                {[1, 2, 3, 4, 5].map((value) => (
                  <button
                    key={value}
                    type="button"
                    className={`${styles.ratingButton} ${formData.satisfaction >= value ? styles.active : ''}`}
                    onClick={() => setFormData({ ...formData, satisfaction: formData.satisfaction === value ? 0 : value })}
                  >
                    ★
                  </button>
                ))}
              </div>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>メモ</label>
              <textarea
                className={styles.textarea}
                value={formData.notes}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                placeholder="練習内容や気づきなど"
                rows={3}
              />
            </div>

            <div className={styles.formActions}>
              {editingRecordId && (
                <button type="button" onClick={resetForm} className={styles.cancelButton}>
                  キャンセル
                </button>
              )}
              <button
                type="button"
                onClick={handleSave}
                disabled={isSaving}
                className={styles.saveButton}
              >
                {isSaving ? '保存中...' : editingRecordId ? '更新' : '保存'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  )
}
