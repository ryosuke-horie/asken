import { fetchDailyMeals } from '@/lib/api/daily-meals'
import DailyMealsView from '@/components/client/DailyMealsView'

export default async function HomePage({
  searchParams,
}: {
  searchParams: { date?: string }
}) {
  const today = new Date().toISOString().split('T')[0]
  const date = searchParams.date || today

  const data = await fetchDailyMeals(date)

  return <DailyMealsView initialData={data} initialDate={date} />
}
