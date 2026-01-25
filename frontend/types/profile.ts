export interface UserProfile {
  id: string
  user_id: string
  sport_type: string | null
  training_goals: string[]
  weight_class: number | null
  created_at: string
  updated_at: string
}

export interface UpdateProfileRequest {
  sport_type?: string | null
  training_goals?: string[]
  weight_class?: number | null
}

export const SPORT_TYPES = [
  '柔術',
  'キックボクシング',
  'MMA',
  'ボクシング',
  'レスリング',
  'ムエタイ',
  '空手',
  'テコンドー',
  '総合格闘技',
  'その他',
] as const

export type SportType = (typeof SPORT_TYPES)[number]

export const TRAINING_GOALS = [
  '減量',
  'スタミナ強化',
  '筋力強化',
  '技術向上',
  '柔軟性向上',
  '維持',
] as const

export type TrainingGoal = (typeof TRAINING_GOALS)[number]

export const TRAINING_GOAL_LABELS: Record<TrainingGoal, string> = {
  減量: '減量',
  スタミナ強化: 'スタミナ強化',
  筋力強化: '筋力強化',
  技術向上: '技術向上',
  柔軟性向上: '柔軟性向上',
  維持: '維持（現状維持）',
}
