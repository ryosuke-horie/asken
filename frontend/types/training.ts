// トレーニング場所
export interface TrainingLocation {
  id: string
  user_id: string
  name: string
  sort_order: number
  created_at: string
  updated_at: string
}

// トレーニング器具
export interface TrainingEquipment {
  id: string
  location_id: string
  name: string
  original_name?: string
  sort_order: number
  created_at: string
  updated_at: string
}

// トレーニングメニュー
export interface TrainingMenu {
  id: string
  user_id?: string
  name: string
  is_default: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

// 練習記録
export interface TrainingRecord {
  id: string
  user_id: string
  location_id?: string
  location_name?: string
  recorded_at: string
  completed: boolean
  duration?: number
  intensity?: number
  satisfaction?: number
  notes?: string
  menus?: TrainingMenu[]
  created_at: string
  updated_at: string
}

// メニュー項目
export interface MenuItem {
  name: string
  duration: number
  sets: number
  reps: string
  equipment: string
  description: string
}

// 正規化された器具名
export interface NormalizedEquipment {
  original: string
  normalized: string
}

// リクエスト型
export interface CreateLocationRequest {
  name: string
}

export interface UpdateLocationRequest {
  name: string
}

export interface CreateEquipmentRequest {
  name: string
  original_name?: string
}

export interface UpdateEquipmentRequest {
  name: string
  original_name?: string
}

export interface UpsertRecordRequest {
  recorded_at: string
  location_id?: string
  completed: boolean
}

export interface CreateRecordRequest {
  recorded_at: string
  location_id?: string
  completed?: boolean
  duration?: number
  intensity?: number
  satisfaction?: number
  notes?: string
  menu_ids?: string[]
}

export interface UpdateRecordRequest {
  recorded_at?: string
  location_id?: string
  completed?: boolean
  duration?: number
  intensity?: number
  satisfaction?: number
  notes?: string
  menu_ids?: string[]
}

export interface CreateMenuRequest {
  name: string
}

export interface SuggestMenuRequest {
  equipment: string[]
  duration: number
  goals?: string[]
}

export interface NormalizeEquipmentRequest {
  names: string[]
}

// レスポンス型
export interface SuggestMenuResponse {
  menu: MenuItem[]
}

export interface NormalizeEquipmentResponse {
  normalized_names: NormalizedEquipment[]
}
