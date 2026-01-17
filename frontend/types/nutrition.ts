export interface FoodItem {
  name: string;
  estimated_amount: string;
}

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
