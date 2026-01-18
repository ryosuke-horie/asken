import { DailyMeals } from '@/types/nutrition'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export async function fetchDailyMeals(date: string): Promise<DailyMeals> {
  const response = await fetch(`${API_BASE_URL}/api/meals/daily?date=${date}`, {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch daily meals')
  }

  return response.json()
}
