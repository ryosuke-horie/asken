export interface WeightRecord {
  id: string
  user_id: string
  weight: number
  recorded_at: string
  created_at: string
}

export interface WeightStats {
  min: number
  max: number
  average: number
}

export interface WeightRecordsResponse {
  records: WeightRecord[]
  latest: WeightRecord | null
  stats: WeightStats
}

export interface WeightGoal {
  target_weight: number
  target_date: string
  days_remaining: number
  weight_to_lose: number
}

export type WeightPeriod = 'week' | 'month' | '3months'
