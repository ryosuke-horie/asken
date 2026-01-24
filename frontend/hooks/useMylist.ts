'use client'

import useSWR from 'swr'
import { useAuth } from '@/contexts/AuthContext'
import { createAuthFetcher } from '@/lib/fetcher'
import type {
  MylistItem,
  CreateMylistItemRequest,
  UpdateMylistItemRequest,
  AnalyzeMylistResponse,
} from '@/types/mylist'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

export function useMylist() {
  const { token } = useAuth()
  const fetcher = createAuthFetcher(token)

  const { data, error, isLoading, mutate } = useSWR<MylistItem[]>(
    token ? `${API_BASE_URL}/api/mylist` : null,
    fetcher,
  )

  const createItem = async (request: CreateMylistItemRequest): Promise<MylistItem> => {
    const response = await fetch(`${API_BASE_URL}/api/mylist`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'マイリストアイテムの作成に失敗しました')
    }

    const created = await response.json()
    mutate()
    return created
  }

  const updateItem = async (id: string, request: UpdateMylistItemRequest): Promise<MylistItem> => {
    const response = await fetch(`${API_BASE_URL}/api/mylist/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'マイリストアイテムの更新に失敗しました')
    }

    const updated = await response.json()
    mutate()
    return updated
  }

  const deleteItem = async (id: string): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/mylist/${id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'マイリストアイテムの削除に失敗しました')
    }

    mutate()
  }

  const reorderItems = async (itemIds: string[]): Promise<void> => {
    const response = await fetch(`${API_BASE_URL}/api/mylist/reorder`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ item_ids: itemIds }),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'マイリストの並び替えに失敗しました')
    }

    mutate()
  }

  const analyzeText = async (inputText: string): Promise<AnalyzeMylistResponse> => {
    const response = await fetch(`${API_BASE_URL}/api/mylist/analyze`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ input_text: inputText }),
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(errorText || 'AI分析に失敗しました')
    }

    return response.json()
  }

  return {
    items: data ?? [],
    isLoading,
    error,
    mutate,
    createItem,
    updateItem,
    deleteItem,
    reorderItems,
    analyzeText,
  }
}
