package gemini

import (
	"fmt"
	"math"
	"strings"
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

// SupportedUnitsCSV はサポート単位をカンマ区切り文字列で返す（プロンプト埋め込み用）
func SupportedUnitsCSV() string {
	return strings.Join(supportedUnits, ", ")
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
// micronutrients.go の AllMicronutrients と同期させること（フィールド追加時は toNutritionInfo() も更新が必要）
type nutritionResponseItem struct {
	Name          string  `json:"name"`
	QuantityValue float64 `json:"quantity_value"`
	QuantityUnit  string  `json:"quantity_unit"`
	Calories      float64 `json:"calories_kcal"`
	Protein       float64 `json:"protein_g"`
	Fat           float64 `json:"fat_g"`
	Carbohydrates float64 `json:"carbohydrates_g"`
	// マイクロニュートリエント
	Iron       float64 `json:"iron_mg"`
	Calcium    float64 `json:"calcium_mg"`
	Zinc       float64 `json:"zinc_mg"`
	Fiber      float64 `json:"fiber_g"`
	VitaminA   float64 `json:"vitamin_a_ug"`
	VitaminB1  float64 `json:"vitamin_b1_mg"`
	VitaminB2  float64 `json:"vitamin_b2_mg"`
	VitaminB6  float64 `json:"vitamin_b6_mg"`
	VitaminB12 float64 `json:"vitamin_b12_ug"`
	VitaminC   float64 `json:"vitamin_c_mg"`
	VitaminD   float64 `json:"vitamin_d_ug"`
	VitaminE   float64 `json:"vitamin_e_mg"`
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
		Micronutrients: map[string]float64{
			string(KeyIron):       r.Iron,
			string(KeyCalcium):    r.Calcium,
			string(KeyZinc):       r.Zinc,
			string(KeyFiber):      r.Fiber,
			string(KeyVitaminA):   r.VitaminA,
			string(KeyVitaminB1):  r.VitaminB1,
			string(KeyVitaminB2):  r.VitaminB2,
			string(KeyVitaminB6):  r.VitaminB6,
			string(KeyVitaminB12): r.VitaminB12,
			string(KeyVitaminC):   r.VitaminC,
			string(KeyVitaminD):   r.VitaminD,
			string(KeyVitaminE):   r.VitaminE,
		},
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
// マイクロニュートリエントのプロパティはAllMicronutrientsレジストリから自動生成される
func NutritionInfoSchema() *Schema {
	properties := map[string]*Schema{
		"name":            {Type: SchemaTypeString},
		"quantity_value":  {Type: SchemaTypeNumber},
		"quantity_unit":   {Type: SchemaTypeString, Enum: SupportedUnits()},
		"calories_kcal":   {Type: SchemaTypeNumber},
		"protein_g":       {Type: SchemaTypeNumber},
		"fat_g":           {Type: SchemaTypeNumber},
		"carbohydrates_g": {Type: SchemaTypeNumber},
	}

	required := []string{
		"name", "quantity_value", "quantity_unit",
		"calories_kcal", "protein_g", "fat_g", "carbohydrates_g",
	}

	for _, m := range AllMicronutrients {
		key := string(m.Key)
		properties[key] = &Schema{Type: SchemaTypeNumber}
		required = append(required, key)
	}

	return &Schema{
		Type: SchemaTypeArray,
		Items: &Schema{
			Type:       SchemaTypeObject,
			Properties: properties,
			Required:   required,
		},
	}
}
