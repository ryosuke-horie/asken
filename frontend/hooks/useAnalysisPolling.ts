'use client'

import { useState, useRef, useCallback, useEffect } from 'react'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''
const MAX_POLLING_ATTEMPTS = 60 // 2分（2秒間隔で60回）

type AnalysisStatus = 'pending' | 'processing' | 'completed' | 'failed'

interface StatusResponse {
  status: AnalysisStatus
  error?: string
}

interface UseAnalysisPollingOptions {
  onComplete?: () => void
  pollingInterval?: number
  maxPollingAttempts?: number
}

interface UseAnalysisPollingResult {
  isLoading: boolean
  error: string | null
  statusMessage: string
  startAnalysis: (analysisId: string) => void
  resetState: () => void
  setIsLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setStatusMessage: (message: string) => void
}

export function useAnalysisPolling(options: UseAnalysisPollingOptions = {}): UseAnalysisPollingResult {
  const { onComplete, pollingInterval = 2000, maxPollingAttempts = MAX_POLLING_ATTEMPTS } = options
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [statusMessage, setStatusMessage] = useState<string>('')
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null)
  const pollingCountRef = useRef(0)

  const stopPolling = useCallback(() => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current)
      pollingIntervalRef.current = null
    }
    pollingCountRef.current = 0
  }, [])

  // クリーンアップ: コンポーネントアンマウント時にポーリングを停止
  useEffect(() => {
    return () => {
      stopPolling()
    }
  }, [stopPolling])

  const checkStatus = useCallback(async (id: string) => {
    // ポーリング回数制限をチェック
    pollingCountRef.current += 1
    if (pollingCountRef.current > maxPollingAttempts) {
      setError('分析処理がタイムアウトしました。しばらく経ってから再度お試しください。')
      stopPolling()
      setIsLoading(false)
      return
    }

    try {
      const response = await fetch(`${API_BASE_URL}/api/analyze/${id}`)

      if (!response.ok) {
        throw new Error(`ステータス取得に失敗しました (${response.status})`)
      }

      const data: StatusResponse = await response.json()

      switch (data.status) {
        case 'pending':
          setStatusMessage('分析リクエストを受け付けました...')
          break

        case 'processing':
          setStatusMessage('分析処理中です...')
          break

        case 'completed':
          setStatusMessage('分析が完了しました')
          stopPolling()
          setIsLoading(false)
          onComplete?.()
          break

        case 'failed':
          setError(data.error || '分析に失敗しました')
          stopPolling()
          setIsLoading(false)
          break
      }
    } catch (err) {
      console.error('Status check error:', err)
      setError(err instanceof Error ? err.message : '予期しないエラー')
      stopPolling()
      setIsLoading(false)
    }
  }, [maxPollingAttempts, onComplete, stopPolling])

  const startAnalysis = useCallback((analysisId: string) => {
    stopPolling()
    checkStatus(analysisId)
    pollingIntervalRef.current = setInterval(() => {
      checkStatus(analysisId)
    }, pollingInterval)
  }, [checkStatus, pollingInterval, stopPolling])

  const resetState = useCallback(() => {
    setIsLoading(false)
    setError(null)
    setStatusMessage('')
    stopPolling()
  }, [stopPolling])

  return {
    isLoading,
    error,
    statusMessage,
    startAnalysis,
    resetState,
    setIsLoading,
    setError,
    setStatusMessage,
  }
}
