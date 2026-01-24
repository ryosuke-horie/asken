'use client'

import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import type {
  TrainingLocation,
  TrainingEquipment,
  TrainingRecord,
  CreateLocationRequest,
  UpdateLocationRequest,
  CreateEquipmentRequest,
  UpdateEquipmentRequest,
  UpsertRecordRequest,
  SuggestMenuRequest,
  SuggestMenuResponse,
  NormalizeEquipmentResponse,
} from '@/types/training'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

// 場所管理用フック
export function useTrainingLocations() {
  const { token } = useAuth()
  const fetcher = createAuthFetcher(token)

  const { data, error, isLoading, mutate } = useSWR<TrainingLocation[]>(
    token ? `${API_BASE_URL}/api/training/locations` : null,
    fetcher,
  )

  const createLocation = async (request: CreateLocationRequest): Promise<TrainingLocation> => {
    const response = await fetch(`${API_BASE_URL}/api/training/locations`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'トレーニング場所の作成に失敗しました')
    }

    const created = await response.json()
    mutate()
    return created
  }

  const updateLocation = async (
    id: string,
    request: UpdateLocationRequest,
  ): Promise<TrainingLocation> => {
    const response = await fetch(`${API_BASE_URL}/api/training/locations/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'トレーニング場所の更新に失敗しました')
    }

    const updated = await response.json()
    mutate()
    return updated
  }

  const deleteLocation = async (id: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/training/locations/${id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'トレーニング場所の削除に失敗しました')
    }

    mutate()
  }

  return {
    locations: data ?? [],
    isLoading,
    error,
    mutate,
    createLocation,
    updateLocation,
    deleteLocation,
  }
}

// 器具管理用フック
export function useTrainingEquipment(locationId: string | null) {
  const { token } = useAuth()
  const fetcher = createAuthFetcher(token)

  const { data, error, isLoading, mutate } = useSWR<TrainingEquipment[]>(
    token && locationId ? `${API_BASE_URL}/api/training/locations/${locationId}/equipment` : null,
    fetcher,
  )

  const createEquipment = async (request: CreateEquipmentRequest): Promise<TrainingEquipment> => {
    if (!locationId) {
      throw new Error('場所が選択されていません')
    }

    const response = await fetch(
      `${API_BASE_URL}/api/training/locations/${locationId}/equipment`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(request),
      },
    )

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || '器具の作成に失敗しました')
    }

    const created = await response.json()
    mutate()
    return created
  }

  const updateEquipment = async (
    id: string,
    request: UpdateEquipmentRequest,
  ): Promise<TrainingEquipment> => {
    const response = await fetch(`${API_BASE_URL}/api/training/equipment/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || '器具の更新に失敗しました')
    }

    const updated = await response.json()
    mutate()
    return updated
  }

  const deleteEquipment = async (id: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/training/equipment/${id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || '器具の削除に失敗しました')
    }

    mutate()
  }

  return {
    equipment: data ?? [],
    isLoading,
    error,
    mutate,
    createEquipment,
    updateEquipment,
    deleteEquipment,
  }
}

// 練習記録用フック
export function useTrainingRecords(startDate?: string, endDate?: string) {
  const { token } = useAuth()
  const fetcher = createAuthFetcher(token)

  const params = new URLSearchParams()
  if (startDate) params.append('start', startDate)
  if (endDate) params.append('end', endDate)
  const queryString = params.toString()

  const { data, error, isLoading, mutate } = useSWR<TrainingRecord[]>(
    token
      ? `${API_BASE_URL}/api/training/records${queryString ? `?${queryString}` : ''}`
      : null,
    fetcher,
  )

  const upsertRecord = async (request: UpsertRecordRequest): Promise<TrainingRecord> => {
    const response = await fetch(`${API_BASE_URL}/api/training/records`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || '練習記録の作成/更新に失敗しました')
    }

    const upserted = await response.json()
    mutate()
    return upserted
  }

  return {
    records: data ?? [],
    isLoading,
    error,
    mutate,
    upsertRecord,
  }
}

// メニュー提案用フック
export function useTrainingMenu() {
  const { token } = useAuth()

  const suggestMenu = async (request: SuggestMenuRequest): Promise<SuggestMenuResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/training/suggest-menu`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'メニュー提案に失敗しました')
    }

    return response.json()
  }

  const normalizeEquipment = async (names: string[]): Promise<NormalizeEquipmentResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/training/normalize-equipment`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ names }),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || '器具名正規化に失敗しました')
    }

    return response.json()
  }

  return {
    suggestMenu,
    normalizeEquipment,
  }
}
