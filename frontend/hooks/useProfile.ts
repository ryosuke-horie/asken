'use client'

import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import type { UserProfile, UpdateProfileRequest } from '@/types/profile'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

async function parseJsonResponse<T>(response: Response, errorMessage: string): Promise<T> {
  try {
    return await response.json()
  } catch (error) {
    const originalError = error instanceof Error ? error.message : String(error)
    throw new Error(`${errorMessage}: レスポンスの解析に失敗しました (原因: ${originalError})`)
  }
}

export function useProfile() {
  const { token } = useAuth()
  const fetcher = createAuthFetcher(token)

  const { data, error, isLoading, mutate } = useSWR<UserProfile | null>(
    token ? `${API_BASE_URL}/api/profile` : null,
    fetcher,
  )

  const updateProfile = async (request: UpdateProfileRequest): Promise<UserProfile> => {
    const response = await fetch(`${API_BASE_URL}/api/profile`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'プロフィールの更新に失敗しました')
    }

    const updated = await parseJsonResponse<UserProfile>(response, 'プロフィールの更新')
    mutate(updated)
    return updated
  }

  return {
    profile: data,
    isLoading,
    error,
    mutate,
    updateProfile,
  }
}
