'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import MylistAnalyzeInput from '@/components/client/MylistAnalyzeInput'
import MylistForm from '@/components/client/MylistForm'
import { useMylist } from '@/hooks/useMylist'
import type { AnalyzeMylistResponse, CreateMylistItemRequest } from '@/types/mylist'
import styles from './page.module.css'

export default function MylistNewPage() {
  const router = useRouter()
  const { createItem, analyzeText } = useMylist()
  const [analyzeResult, setAnalyzeResult] = useState<AnalyzeMylistResponse | null>(null)
  const [inputText, setInputText] = useState<string>('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleAnalyzeResult = (result: AnalyzeMylistResponse, text: string) => {
    setAnalyzeResult(result)
    setInputText(text)
    setError(null)
  }

  const handleSubmit = async (data: CreateMylistItemRequest) => {
    setIsSubmitting(true)
    setError(null)

    try {
      await createItem(data)
      router.push('/mylist')
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存に失敗しました')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <ProtectedRoute>
      <div className={styles.container}>
        <div className={styles.header}>
          <Link href="/mylist" className={styles.backLink}>
            ← マイリスト
          </Link>
          <h1 className={styles.title}>新規登録</h1>
        </div>

        {error && <div className={styles.error}>{error}</div>}

        {!analyzeResult ? (
          <div className={styles.analyzeSection}>
            <p className={styles.description}>
              よく食べるメニューを入力して、AIで栄養素を分析しましょう。
            </p>
            <MylistAnalyzeInput onAnalyze={analyzeText} onResult={handleAnalyzeResult} />
            <div className={styles.manualOption}>
              <span>または</span>
              <button
                type="button"
                onClick={() =>
                  setAnalyzeResult({
                    foods: [],
                    total_calories: 0,
                    total_protein: 0,
                    total_fat: 0,
                    total_carbohydrates: 0,
                  })
                }
                className={styles.manualButton}
              >
                手動で入力する
              </button>
            </div>
          </div>
        ) : (
          <div className={styles.formSection}>
            <button
              type="button"
              onClick={() => {
                setAnalyzeResult(null)
                setInputText('')
              }}
              className={styles.reanalyzeButton}
            >
              ← 再分析する
            </button>
            <MylistForm
              analyzeResult={analyzeResult}
              inputText={inputText}
              onSubmit={handleSubmit}
              isSubmitting={isSubmitting}
              submitLabel="登録する"
            />
          </div>
        )}
      </div>
    </ProtectedRoute>
  )
}
