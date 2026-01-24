export type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack'

export type InputType = 'image' | 'text' | 'mylist' | 'skipped'

export interface NutritionInfo {
  name: string
  estimated_amount: string
  calories_kcal: number
  protein_g: number
  fat_g: number
  carbohydrates_g: number
}

export interface AnalysisResult {
  foods: NutritionInfo[]
  total_calories: number
  total_protein: number
  total_fat: number
  total_carbohydrates: number
}

interface HistoryItem {
  id: string
  input_type: InputType
  image_path: string
  input_text: string
  created_at: string
  meal_type?: MealType
  meal_date?: string
  total_calories: number
  total_protein: number
  total_fat: number
  total_carbohydrates: number
}

export interface HistoryDetail extends HistoryItem {
  foods: NutritionInfo[]
}

export interface DailyMeals {
  date: string
  meals: {
    breakfast: HistoryDetail[]
    lunch: HistoryDetail[]
    dinner: HistoryDetail[]
    snack: HistoryDetail[]
  }
  daily_total: {
    total_calories: number
    total_protein: number
    total_fat: number
    total_carbohydrates: number
  }
}
