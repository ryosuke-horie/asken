export type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack'

export interface NutritionInfo {
  name: string;
  estimated_amount: string;
  calories_kcal: number;
  protein_g: number;
  fat_g: number;
  carbohydrates_g: number;
}

export interface AnalysisResult {
  foods: NutritionInfo[];
  total_calories: number;
  total_protein: number;
  total_fat: number;
  total_carbohydrates: number;
}

export interface HistoryItem {
  id: string;
  image_path: string;
  created_at: string;
  meal_type?: MealType;
  meal_date?: string;
  total_calories: number;
  total_protein: number;
  total_fat: number;
  total_carbohydrates: number;
}

export interface HistoryDetail extends HistoryItem {
  foods: NutritionInfo[];
}

export interface HistoryListResponse {
  items: HistoryItem[];
  total: number;
  page: number;
  limit: number;
}

export interface DailyMeals {
  date: string;
  meals: {
    breakfast: HistoryDetail[];
    lunch: HistoryDetail[];
    dinner: HistoryDetail[];
    snack: HistoryDetail[];
  };
  daily_total: {
    total_calories: number;
    total_protein: number;
    total_fat: number;
    total_carbohydrates: number;
  };
}
