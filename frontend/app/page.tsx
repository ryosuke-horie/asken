'use client'

import { useSearchParams } from 'next/navigation'
import DailyMealsView from '@/components/client/DailyMealsView'
import ProtectedRoute from '@/components/client/ProtectedRoute'

export default function HomePage() {
  const searchParams = useSearchParams()
  const today = new Date().toISOString().split('T')[0]
  const date = searchParams.get('date') || today

  return (
    <ProtectedRoute>
      <DailyMealsView date={date} />
    </ProtectedRoute>
  )
}
