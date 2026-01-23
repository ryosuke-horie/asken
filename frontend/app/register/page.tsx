'use client'

import { useState, FormEvent, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useAuth } from '@/contexts/AuthContext'
import styles from './page.module.css'

export default function RegisterPage() {
  const router = useRouter()
  const { register, isAuthenticated, isLoading: authLoading } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [name, setName] = useState('')
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

    if (password.length < 8) {
      setError('パスワードは8文字以上で入力してください')
      return
    }

    if (password !== confirmPassword) {
      setError('パスワードが一致しません')
      return
    }

    setIsLoading(true)

    const result = await register(email, password, name || undefined)

    if (result.success) {
      router.push('/')
    } else {
      setError(result.error || '登録に失敗しました')
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
        <h2 className={styles.title}>新規登録</h2>

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
            <label htmlFor="name" className={styles.label}>
              名前（任意）
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={isLoading}
              className={styles.input}
              placeholder="お名前"
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

          <div className={styles.field}>
            <label htmlFor="confirmPassword" className={styles.label}>
              パスワード（確認）
            </label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              disabled={isLoading}
              className={styles.input}
              placeholder="パスワードを再入力"
            />
          </div>

          {error && <div className={styles.error}>{error}</div>}

          <button
            type="submit"
            disabled={isLoading}
            className={styles.submitButton}
          >
            {isLoading ? '登録中...' : '登録'}
          </button>
        </form>

        <p className={styles.linkText}>
          既にアカウントをお持ちの方は{' '}
          <Link href="/login" className={styles.link}>
            ログイン
          </Link>
        </p>
      </div>
    </div>
  )
}
