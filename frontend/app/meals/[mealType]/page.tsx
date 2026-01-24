'use client'

import { useParams, useSearchParams, notFound } from 'next/navigation'
import MealUploadView from '@/components/client/MealUploadView'
import ProtectedRoute from '@/components/client/ProtectedRoute'
import { MealType } from '@/types/nutrition'

const VALID_MEAL_TYPES: MealType[] = ['breakfast', 'lunch', 'dinner', 'snack']

const MEAL_TYPE_LABELS: Record<MealType, string> = {
  breakfast: '朝食',
  lunch: '昼食',
  dinner: '夕食',
  snack: '間食',
}

export default function MealUploadPage() {
  const params = useParams()
  const searchParams = useSearchParams()
  const mealType = params.mealType as MealType

  // 無効な食事タイプの場合は404
  if (!VALID_MEAL_TYPES.includes(mealType)) {
    notFound()
  }

  const today = new Date().toISOString().split('T')[0]
  const date = searchParams.get('date') || today

  return (
    <ProtectedRoute>
      <MealUploadView mealType={mealType} mealDate={date} mealLabel={MEAL_TYPE_LABELS[mealType]} />
    </ProtectedRoute>
  )
}
