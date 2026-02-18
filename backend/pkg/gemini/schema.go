package gemini

import (
	"fmt"
	"math"
)

// supportedUnits はiOS側のQuantityParserがパース可能な単位の一覧
// MeasurementUnit (iOS) と完全に一致させること
var supportedUnits = []string{
	"g", "ml",
	"杯", "人前", "個", "枚", "本", "切れ", "食", "皿",
	"膳", "丁", "束", "袋", "缶", "合", "玉", "粒",
	"パック", "大さじ", "小さじ",
}

// SupportedUnits はサポート単位のコピーを返す
func SupportedUnits() []string {
	result := make([]string, len(supportedUnits))
	copy(result, supportedUnits)
	return result
}

// classifierResponseItem はGemini Classifier/TextParserのレスポンス構造体
// responseSchemaで制約されたJSON出力をデシリアライズする
type classifierResponseItem struct {
	Name          string  `json:"name"`
	QuantityValue float64 `json:"quantity_value"`
	QuantityUnit  string  `json:"quantity_unit"`
}

// toFoodItem はclassifierResponseItemをFoodItemに変換する
func (r classifierResponseItem) toFoodItem() FoodItem {
	return FoodItem{
		Name:            r.Name,
		EstimatedAmount: FormatQuantity(r.QuantityValue, r.QuantityUnit),
	}
}

// nutritionResponseItem はGemini NutritionCalculatorのレスポンス構造体
// responseSchemaで制約されたJSON出力をデシリアライズする
type nutritionResponseItem struct {
	Name          string  `json:"name"`
	QuantityValue float64 `json:"quantity_value"`
	QuantityUnit  string  `json:"quantity_unit"`
	Calories      float64 `json:"calories_kcal"`
	Protein       float64 `json:"protein_g"`
	Fat           float64 `json:"fat_g"`
	Carbohydrates float64 `json:"carbohydrates_g"`
}

// toNutritionInfo はnutritionResponseItemをNutritionInfoに変換する
func (r nutritionResponseItem) toNutritionInfo() NutritionInfo {
	return NutritionInfo{
		Name:            r.Name,
		EstimatedAmount: FormatQuantity(r.QuantityValue, r.QuantityUnit),
		Calories:        r.Calories,
		Protein:         r.Protein,
		Fat:             r.Fat,
		Carbohydrates:   r.Carbohydrates,
	}
}

// FormatQuantity は数値と単位から "{値}{単位}" 形式の文字列を生成する
// 整数の場合は小数点を省略する（例: 1.0 → "1個", 1.5 → "1.5個"）
func FormatQuantity(value float64, unit string) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Sprintf("0%s", unit)
	}
	if value == float64(int64(value)) {
		return fmt.Sprintf("%d%s", int64(value), unit)
	}
	return fmt.Sprintf("%.1f%s", value, unit)
}

// FoodItemSchema はClassifier/TextParser用のresponseSchema
func FoodItemSchema() *Schema {
	return &Schema{
		Type: SchemaTypeArray,
		Items: &Schema{
			Type: SchemaTypeObject,
			Properties: map[string]*Schema{
				"name": {
					Type: SchemaTypeString,
				},
				"quantity_value": {
					Type: SchemaTypeNumber,
				},
				"quantity_unit": {
					Type: SchemaTypeString,
					Enum: SupportedUnits(),
				},
			},
			Required: []string{"name", "quantity_value", "quantity_unit"},
		},
	}
}

// NutritionInfoSchema はNutritionCalculator用のresponseSchema
func NutritionInfoSchema() *Schema {
	return &Schema{
		Type: SchemaTypeArray,
		Items: &Schema{
			Type: SchemaTypeObject,
			Properties: map[string]*Schema{
				"name": {
					Type: SchemaTypeString,
				},
				"quantity_value": {
					Type: SchemaTypeNumber,
				},
				"quantity_unit": {
					Type: SchemaTypeString,
					Enum: SupportedUnits(),
				},
				"calories_kcal": {
					Type: SchemaTypeNumber,
				},
				"protein_g": {
					Type: SchemaTypeNumber,
				},
				"fat_g": {
					Type: SchemaTypeNumber,
				},
				"carbohydrates_g": {
					Type: SchemaTypeNumber,
				},
			},
			Required: []string{"name", "quantity_value", "quantity_unit", "calories_kcal", "protein_g", "fat_g", "carbohydrates_g"},
		},
	}
}
