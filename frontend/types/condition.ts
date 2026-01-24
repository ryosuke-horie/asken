export type ConditionLevel = 1 | 2 | 3
export type FatigueLevel = 1 | 2 | 3

export interface ConditionRecord {
  id: string
  user_id: string
  condition: ConditionLevel
  fatigue: FatigueLevel
  recorded_at: string
  created_at: string
}

export const CONDITION_LABELS: Record<ConditionLevel, string> = {
  1: '悪い',
  2: '普通',
  3: '良い',
}

export const FATIGUE_LABELS: Record<FatigueLevel, string> = {
  1: '低い',
  2: '普通',
  3: '高い',
}

export function isConditionLevel(value: number): value is ConditionLevel {
  return value === 1 || value === 2 || value === 3
}

export function isFatigueLevel(value: number): value is FatigueLevel {
  return value === 1 || value === 2 || value === 3
}
