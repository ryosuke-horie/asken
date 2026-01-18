import { notFound } from 'next/navigation'
import MealUploadView from '@/components/client/MealUploadView'
import { MealType } from '@/types/nutrition'
import { fetchDailyMeals } from '@/lib/api/daily-meals'

const VALID_MEAL_TYPES: MealType[] = ['breakfast', 'lunch', 'dinner', 'snack']

const MEAL_TYPE_LABELS: Record<MealType, string> = {
  breakfast: '朝食',
  lunch: '昼食',
  dinner: '夕食',
  snack: '間食',
}

interface MealUploadPageProps {
  params: {
    mealType: string
  }
  searchParams: {
    date?: string
  }
}

export default async function MealUploadPage({ params, searchParams }: MealUploadPageProps) {
  const mealType = params.mealType as MealType

  // 無効な食事タイプの場合は404
  if (!VALID_MEAL_TYPES.includes(mealType)) {
    notFound()
  }

  const today = new Date().toISOString().split('T')[0]
  const date = searchParams.date || today

  // 日次食事データを取得
  const dailyMeals = await fetchDailyMeals(date)
  const existingMeals = dailyMeals.meals[mealType]

  return (
    <MealUploadView
      mealType={mealType}
      mealDate={date}
      mealLabel={MEAL_TYPE_LABELS[mealType]}
      initialMeals={existingMeals}
    />
  )
}
