'use client'

import { createContext, useContext, useState, useEffect, ReactNode } from 'react'

interface UserEmailContextType {
  email: string | null
  setEmail: (email: string) => boolean
  clearEmail: () => boolean
  isLoading: boolean
  storageError: boolean
}

const UserEmailContext = createContext<UserEmailContextType | undefined>(undefined)

const STORAGE_KEY = 'asken_user_email'

export function UserEmailProvider({ children }: { children: ReactNode }) {
  const [email, setEmailState] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [storageError, setStorageError] = useState(false)

  useEffect(() => {
    try {
      const storedEmail = localStorage.getItem(STORAGE_KEY)
      if (storedEmail) {
        setEmailState(storedEmail)
      }
    } catch (error) {
      console.error('Failed to access localStorage:', error)
      setStorageError(true)
    }
    setIsLoading(false)
  }, [])

  const setEmail = (newEmail: string): boolean => {
    try {
      localStorage.setItem(STORAGE_KEY, newEmail)
      setEmailState(newEmail)
      return true
    } catch (error) {
      console.error('Failed to save email to localStorage:', error)
      return false
    }
  }

  const clearEmail = (): boolean => {
    try {
      localStorage.removeItem(STORAGE_KEY)
      setEmailState(null)
      return true
    } catch (error) {
      console.error('Failed to remove email from localStorage:', error)
      return false
    }
  }

  return (
    <UserEmailContext.Provider value={{ email, setEmail, clearEmail, isLoading, storageError }}>
      {children}
    </UserEmailContext.Provider>
  )
}

export function useUserEmail() {
  const context = useContext(UserEmailContext)
  if (context === undefined) {
    throw new Error('useUserEmail must be used within a UserEmailProvider')
  }
  return context
}
