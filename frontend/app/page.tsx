'use client'

import { useSearchParams } from 'next/navigation'
import DailyMealsView from '@/components/client/DailyMealsView'
import WeightSection from '@/components/client/WeightSection'
import ProtectedRoute from '@/components/client/ProtectedRoute'

export default function HomePage() {
  const searchParams = useSearchParams()
  const today = new Date().toISOString().split('T')[0]
  const date = searchParams.get('date') || today

  return (
    <ProtectedRoute>
      <WeightSection />
      <DailyMealsView date={date} />
    </ProtectedRoute>
  )
}
