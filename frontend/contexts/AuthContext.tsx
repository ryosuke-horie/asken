'use client'

import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react'

interface User {
  id: string
  email: string
  name?: string
}

interface AuthContextType {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>
  register: (email: string, password: string, name?: string) => Promise<{ success: boolean; error?: string }>
  logout: () => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const TOKEN_STORAGE_KEY = 'uchikomi_auth_token'
const USER_STORAGE_KEY = 'uchikomi_user'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || ''

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    try {
      const storedToken = localStorage.getItem(TOKEN_STORAGE_KEY)
      const storedUser = localStorage.getItem(USER_STORAGE_KEY)

      if (storedToken && storedUser) {
        const parsedUser = JSON.parse(storedUser)
        // 基本的なユーザーデータの検証
        if (parsedUser && typeof parsedUser.id === 'string' && typeof parsedUser.email === 'string') {
          setToken(storedToken)
          setUser(parsedUser)
        } else {
          // 不正なデータの場合はクリア
          console.error('Invalid user data in localStorage, clearing...')
          localStorage.removeItem(TOKEN_STORAGE_KEY)
          localStorage.removeItem(USER_STORAGE_KEY)
        }
      }
    } catch (error) {
      // パースエラーやその他のエラーの場合はデータをクリア
      console.error('Failed to restore auth state, clearing corrupted data:', error)
      try {
        localStorage.removeItem(TOKEN_STORAGE_KEY)
        localStorage.removeItem(USER_STORAGE_KEY)
      } catch {
        // localStorageへのアクセス自体が失敗する場合は無視
      }
    }
    setIsLoading(false)
  }, [])

  const login = useCallback(async (email: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        // ステータスコードに応じたエラーメッセージ
        if (response.status === 401) {
          return { success: false, error: errorText || 'メールアドレスまたはパスワードが正しくありません' }
        }
        if (response.status === 400) {
          return { success: false, error: errorText || '入力内容を確認してください' }
        }
        if (response.status >= 500) {
          return { success: false, error: 'サーバーエラーが発生しました。しばらく経ってからお試しください' }
        }
        return { success: false, error: errorText || 'ログインに失敗しました' }
      }

      const data = await response.json()

      localStorage.setItem(TOKEN_STORAGE_KEY, data.token)
      localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(data.user))

      setToken(data.token)
      setUser(data.user)

      return { success: true }
    } catch (error) {
      console.error('Login error:', error)
      // ネットワークエラーの詳細を判定
      if (error instanceof TypeError && error.message.includes('fetch')) {
        return { success: false, error: 'サーバーに接続できません。ネットワーク接続を確認してください' }
      }
      return { success: false, error: 'ネットワークエラーが発生しました' }
    }
  }, [])

  const register = useCallback(async (email: string, password: string, name?: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password, name }),
      })

      if (!response.ok) {
        const errorText = await response.text()
        // ステータスコードに応じたエラーメッセージ
        if (response.status === 409) {
          return { success: false, error: errorText || 'このメールアドレスは既に登録されています' }
        }
        if (response.status === 400) {
          return { success: false, error: errorText || '入力内容を確認してください' }
        }
        if (response.status >= 500) {
          return { success: false, error: 'サーバーエラーが発生しました。しばらく経ってからお試しください' }
        }
        return { success: false, error: errorText || '登録に失敗しました' }
      }

      const data = await response.json()

      localStorage.setItem(TOKEN_STORAGE_KEY, data.token)
      localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(data.user))

      setToken(data.token)
      setUser(data.user)

      return { success: true }
    } catch (error) {
      console.error('Register error:', error)
      // ネットワークエラーの詳細を判定
      if (error instanceof TypeError && error.message.includes('fetch')) {
        return { success: false, error: 'サーバーに接続できません。ネットワーク接続を確認してください' }
      }
      return { success: false, error: 'ネットワークエラーが発生しました' }
    }
  }, [])

  const logout = useCallback(() => {
    try {
      localStorage.removeItem(TOKEN_STORAGE_KEY)
      localStorage.removeItem(USER_STORAGE_KEY)
    } catch (error) {
      console.error('Failed to clear auth state:', error)
    }
    setToken(null)
    setUser(null)
  }, [])

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!token && !!user,
        isLoading,
        login,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
