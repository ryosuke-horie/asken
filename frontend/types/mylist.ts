import type { NutritionInfo } from './nutrition'

export interface MylistItem {
  id: string
  user_id: string
  name: string
  base_amount: string
  unit: string
  calories: number
  protein: number
  fat: number
  carbohydrates: number
  foods: NutritionInfo[]
  image_path?: string
  sort_order: number
  created_at: string
  updated_at: string
}

export interface CreateMylistItemRequest {
  name: string
  base_amount: string
  unit: string
  calories: number
  protein: number
  fat: number
  carbohydrates: number
  foods: NutritionInfo[]
  image_path?: string
}

export interface UpdateMylistItemRequest {
  name: string
  base_amount: string
  unit: string
  calories: number
  protein: number
  fat: number
  carbohydrates: number
  foods: NutritionInfo[]
  image_path?: string
}

export interface AnalyzeMylistRequest {
  input_text: string
}

export interface AnalyzeMylistResponse {
  foods: NutritionInfo[]
  total_calories: number
  total_protein: number
  total_fat: number
  total_carbohydrates: number
}
