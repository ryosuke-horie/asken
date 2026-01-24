'use client'

import { useState, FormEvent, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useAuth } from '@/contexts/AuthContext'
import styles from './page.module.css'

export default function LoginPage() {
  const router = useRouter()
  const { login, isAuthenticated, isLoading: authLoading } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (!authLoading && isAuthenticated) {
      router.push('/')
    }
  }, [isAuthenticated, authLoading, router])

  async function handleSubmit(e: FormEvent<HTMLFormElement>): Promise<void> {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    const result = await login(email, password)

    if (result.success) {
      router.push('/')
    } else {
      setError(result.error || 'ログインに失敗しました')
      setIsLoading(false)
    }
  }

  if (authLoading) {
    return <div className={styles.loading}>読み込み中...</div>
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <div className={styles.container}>
      <div className={styles.formCard}>
        <h2 className={styles.title}>ログイン</h2>

        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="email" className={styles.label}>
              メールアドレス
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={isLoading}
              className={styles.input}
              placeholder="example@email.com"
            />
          </div>

          <div className={styles.field}>
            <label htmlFor="password" className={styles.label}>
              パスワード
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={isLoading}
              className={styles.input}
              placeholder="8文字以上"
            />
          </div>

          {error && <div className={styles.error}>{error}</div>}

          <button type="submit" disabled={isLoading} className={styles.submitButton}>
            {isLoading ? 'ログイン中...' : 'ログイン'}
          </button>
        </form>

        <p className={styles.linkText}>
          アカウントをお持ちでない方は{' '}
          <Link href="/register" className={styles.link}>
            新規登録
          </Link>
        </p>
      </div>
    </div>
  )
}
