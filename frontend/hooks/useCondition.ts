'use client'

import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

export interface ConditionRecord {
  id: string
  user_id: string
  condition: number
  fatigue: number
  recorded_at: string
  created_at: string
  updated_at: string
}

// 指定日の体調記録を取得するフック
export function useConditionRecord(date: string | null) {
  const { token } = useAuth()
  const fetcher = createAuthFetcher(token)

  const { data, error, isLoading, mutate } = useSWR<ConditionRecord | null>(
    token && date ? `${API_BASE_URL}/api/condition-records?date=${date}` : null,
    fetcher,
  )

  return {
    record: data ?? null,
    isLoading,
    error,
    mutate,
  }
}

// 今日の体調記録を取得するフック
export function useTodayCondition() {
  const today = new Date().toISOString().split('T')[0]
  return useConditionRecord(today)
}
