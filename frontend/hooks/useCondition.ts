'use client'

import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import type { ConditionRecord } from '@/types/condition'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

// 指定日の体調記録を取得するフック（内部使用）
function useConditionRecord(date: string | null) {
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
