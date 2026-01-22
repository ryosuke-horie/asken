'use client'

import { useState } from 'react'
import { useUserEmail } from '@/contexts/UserEmailContext'
import styles from './EmailSetup.module.css'

export default function EmailSetup() {
  const { email, setEmail, isLoading } = useUserEmail()
  const [inputEmail, setInputEmail] = useState('')
  const [error, setError] = useState('')

  if (isLoading) {
    return <div className={styles.loading}>読み込み中...</div>
  }

  if (email) {
    return null
  }

  const validateEmail = (value: string): boolean => {
    if (value.length > 255) {
      return false
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(value)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const trimmedEmail = inputEmail.trim()
    if (!trimmedEmail) {
      setError('メールアドレスを入力してください')
      return
    }

    if (!validateEmail(trimmedEmail)) {
      setError(
        trimmedEmail.length > 255
          ? 'メールアドレスは255文字以内で入力してください'
          : '有効なメールアドレスを入力してください'
      )
      return
    }

    const success = setEmail(trimmedEmail)
    if (!success) {
      setError('メールアドレスの保存に失敗しました。ブラウザの設定を確認してください。')
    }
  }

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        <h2 className={styles.title}>ようこそ</h2>
        <p className={styles.description}>
          食事記録を保存するためにメールアドレスを入力してください
        </p>
        <form onSubmit={handleSubmit} className={styles.form}>
          <input
            type="email"
            value={inputEmail}
            onChange={(e) => setInputEmail(e.target.value)}
            placeholder="example@email.com"
            className={styles.input}
            autoFocus
          />
          {error && <p className={styles.error}>{error}</p>}
          <button type="submit" className={styles.button}>
            始める
          </button>
        </form>
      </div>
    </div>
  )
}
