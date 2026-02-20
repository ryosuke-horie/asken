package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// IngredientContext はGeminiに渡す食材コンテキスト
type IngredientContext struct {
	ID         string
	Name       string
	Quantity   float64
	Unit       string
	ExpiryDate *time.Time
}

// NutritionGoalContext はGeminiに渡す栄養目標コンテキスト
type NutritionGoalContext struct {
	TargetCalories      float64
	TargetProtein       float64
	TargetFat           float64
	TargetCarbohydrates float64
	Phase               string
}

// RecentMealContext はGeminiに渡す直近食事コンテキスト
type RecentMealContext struct {
	Date          string
	MealType      string
	Name          string
	TotalCalories float64
	TotalProtein  float64
	TotalFat      float64
	TotalCarbs    float64
}

// WeightTrendContext はGeminiに渡す体重推移コンテキスト
type WeightTrendContext struct {
	Date   string
	Weight float64
}

// MenuSuggestionInput はメニュー提案生成の入力コンテキストまとめ
type MenuSuggestionInput struct {
	MealType      string
	Count         int
	Ingredients   []IngredientContext
	NutritionGoal *NutritionGoalContext
	RecentMeals   []RecentMealContext
	WeightTrend   []WeightTrendContext
}

// GeminiIngredient はGeminiレスポンスの食材
type GeminiIngredient struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

// GeminiEstimatedNutrition はGeminiレスポンスの推定栄養素
type GeminiEstimatedNutrition struct {
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Fat           float64 `json:"fat"`
	Carbohydrates float64 `json:"carbohydrates"`
}

// GeminiMenuSuggestion はGeminiレスポンスのメニュー提案
type GeminiMenuSuggestion struct {
	Title              string                   `json:"title"`
	Description        string                   `json:"description"`
	Reason             string                   `json:"reason"`
	Ingredients        []GeminiIngredient       `json:"ingredients"`
	EstimatedNutrition GeminiEstimatedNutrition `json:"estimatedNutrition"`
}

// menuSuggestionResponse はGeminiのメニュー提案レスポンス
type menuSuggestionResponse struct {
	Suggestions []GeminiMenuSuggestion `json:"suggestions"`
}

// recipeResponse はGeminiのレシピ生成レスポンス
type recipeResponse struct {
	Recipe string `json:"recipe"`
}

// MenuSuggester はメニュー提案を生成するクライアント
type MenuSuggester struct {
	httpClient GeminiHTTPClient
}

// NewMenuSuggester は新しいMenuSuggesterを作成する
// 環境変数GEMINI_API_KEYからAPIキーを読み取る
func NewMenuSuggester(timeout time.Duration) (*MenuSuggester, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	httpClient, err := NewHTTPClient(apiKey, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu suggester: %w", err)
	}
	return &MenuSuggester{httpClient: httpClient}, nil
}

// NewMenuSuggesterWithHTTPClient はHTTPClientを受け取るコンストラクタ（テスト用）
func NewMenuSuggesterWithHTTPClient(httpClient GeminiHTTPClient) *MenuSuggester {
	return &MenuSuggester{httpClient: httpClient}
}

// SuggestMenus はコンテキストに基づいてメニュー提案を生成する
func (s *MenuSuggester) SuggestMenus(ctx context.Context, input MenuSuggestionInput) ([]GeminiMenuSuggestion, error) {
	prompt := buildSuggestionPrompt(input)
	schema := buildSuggestionSchema(input.Count)

	log.Printf("MenuSuggester: メニュー提案生成を開始 (mealType: %s, count: %d)", input.MealType, input.Count)

	resp, err := s.httpClient.Execute(ctx, prompt, schema)
	if err != nil {
		log.Printf("MenuSuggester: Gemini API呼び出しエラー: %v", err)
		return nil, fmt.Errorf("Gemini APIコールエラー: %w", err)
	}

	jsonStr := removeCodeBlock(resp.Response)
	var parsed menuSuggestionResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("MenuSuggester: レスポンスのJSONパースエラー: %v\nデータ: %s", err, jsonStr)
		return nil, fmt.Errorf("メニュー提案レスポンスのパースエラー: %w", err)
	}

	log.Printf("MenuSuggester: メニュー提案生成完了 (%d件)", len(parsed.Suggestions))
	return parsed.Suggestions, nil
}

// GenerateRecipe はメニューのレシピを生成する
func (s *MenuSuggester) GenerateRecipe(ctx context.Context, title string, ingredients []GeminiIngredient) (string, error) {
	prompt := buildRecipePrompt(title, ingredients)
	schema := buildRecipeSchema()

	log.Printf("MenuSuggester: レシピ生成を開始 (title: %s)", title)

	resp, err := s.httpClient.Execute(ctx, prompt, schema)
	if err != nil {
		log.Printf("MenuSuggester: レシピ生成APIエラー: %v", err)
		return "", fmt.Errorf("レシピ生成APIコールエラー: %w", err)
	}

	jsonStr := removeCodeBlock(resp.Response)
	var parsed recipeResponse
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("MenuSuggester: レシピレスポンスのパースエラー: %v", err)
		return "", fmt.Errorf("レシピレスポンスのパースエラー: %w", err)
	}

	log.Printf("MenuSuggester: レシピ生成完了 (title: %s)", title)
	return parsed.Recipe, nil
}

// buildSuggestionPrompt はメニュー提案用プロンプトを構築する
func buildSuggestionPrompt(input MenuSuggestionInput) string {
	var sb strings.Builder

	mealTypeLabel := mealTypeToLabel(input.MealType)
	fmt.Fprintf(&sb, "アスリート向け食事管理アプリとして、%sに最適な自炊メニューを%d件提案してください。\n\n", mealTypeLabel, input.Count)

	// 食材リスト
	sb.WriteString("【利用可能な食材】\n")
	if len(input.Ingredients) == 0 {
		sb.WriteString("食材なし（一般的な食材で提案してください）\n")
	} else {
		for _, ing := range input.Ingredients {
			if ing.ExpiryDate != nil {
				fmt.Fprintf(&sb, "- %s: %.1f %s（消費期限: %s）\n", ing.Name, ing.Quantity, ing.Unit, ing.ExpiryDate.Format("2006-01-02"))
			} else {
				fmt.Fprintf(&sb, "- %s: %.1f %s\n", ing.Name, ing.Quantity, ing.Unit)
			}
		}
	}
	sb.WriteString("\n")

	// 栄養目標
	if input.NutritionGoal != nil {
		g := input.NutritionGoal
		phase := g.Phase
		if phase == "" {
			phase = "維持"
		}
		fmt.Fprintf(&sb, "【栄養目標（フェーズ: %s）】\n", phase)
		fmt.Fprintf(&sb, "- 1日のカロリー目標: %.0fkcal\n", g.TargetCalories)
		fmt.Fprintf(&sb, "- タンパク質: %.1fg / 脂質: %.1fg / 炭水化物: %.1fg\n", g.TargetProtein, g.TargetFat, g.TargetCarbohydrates)
		sb.WriteString("\n")
	}

	// 直近食事履歴
	if len(input.RecentMeals) > 0 {
		sb.WriteString("【直近7日間の食事記録】\n")
		for _, meal := range input.RecentMeals {
			fmt.Fprintf(&sb, "- %s %s: %s（%s %.0fkcal / P:%.1fg / F:%.1fg / C:%.1fg）\n",
				meal.Date, mealTypeToLabel(meal.MealType), meal.Name,
				meal.MealType, meal.TotalCalories, meal.TotalProtein, meal.TotalFat, meal.TotalCarbs)
		}
		sb.WriteString("\n")
	}

	// 体重推移
	if len(input.WeightTrend) > 0 {
		sb.WriteString("【直近30日間の体重推移】\n")
		for _, w := range input.WeightTrend {
			fmt.Fprintf(&sb, "- %s: %.1fkg\n", w.Date, w.Weight)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("提案のルール:\n")
	sb.WriteString("- 消費期限が近い食材を優先的に使用してください\n")
	sb.WriteString("- 直近の食事と重複しないメニューを提案してください\n")
	sb.WriteString("- 栄養目標の不足分を補うメニューを優先してください\n")
	sb.WriteString("- 体重が増加傾向なら低カロリーメニューを、減少傾向なら適度なカロリーのメニューを提案してください\n")
	sb.WriteString("- 各提案に「なぜこのメニューを提案したか」の理由を含めてください\n")
	sb.WriteString("- 食材の名前は提供されたリストの名前を正確に使用してください\n")
	sb.WriteString("- estimatedNutritionは1食分の推定栄養素を記載してください\n")
	fmt.Fprintf(&sb, "- mealTypeは\"%s\"に固定してください\n", input.MealType)

	return sb.String()
}

// buildRecipePrompt はレシピ生成用プロンプトを構築する
func buildRecipePrompt(title string, ingredients []GeminiIngredient) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "「%s」の調理手順を詳しく教えてください。\n\n", title)
	sb.WriteString("【使用食材】\n")
	for _, ing := range ingredients {
		fmt.Fprintf(&sb, "- %s: %.1f %s\n", ing.Name, ing.Quantity, ing.Unit)
	}
	sb.WriteString("\n")
	sb.WriteString("調理手順のルール:\n")
	sb.WriteString("- 番号付きリスト形式で記述してください\n")
	sb.WriteString("- 各手順は簡潔かつ具体的に記述してください\n")
	sb.WriteString("- 火加減や時間の目安を含めてください\n")
	sb.WriteString("- 調味料の分量も明記してください\n")
	return sb.String()
}

// buildSuggestionSchema はメニュー提案のGeminiスキーマを構築する
func buildSuggestionSchema(count int) *Schema {
	ingredientSchema := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"name":     {Type: SchemaTypeString},
			"quantity": {Type: SchemaTypeNumber},
			"unit":     {Type: SchemaTypeString},
		},
		Required: []string{"name", "quantity", "unit"},
	}

	nutritionSchema := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"calories":      {Type: SchemaTypeNumber},
			"protein":       {Type: SchemaTypeNumber},
			"fat":           {Type: SchemaTypeNumber},
			"carbohydrates": {Type: SchemaTypeNumber},
		},
		Required: []string{"calories", "protein", "fat", "carbohydrates"},
	}

	suggestionSchema := &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"title":       {Type: SchemaTypeString},
			"description": {Type: SchemaTypeString},
			"reason":      {Type: SchemaTypeString},
			"ingredients": {
				Type:  SchemaTypeArray,
				Items: ingredientSchema,
			},
			"estimatedNutrition": nutritionSchema,
		},
		Required: []string{"title", "description", "reason", "ingredients", "estimatedNutrition"},
	}

	return &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"suggestions": {
				Type:  SchemaTypeArray,
				Items: suggestionSchema,
			},
		},
		Required: []string{"suggestions"},
	}
}

// buildRecipeSchema はレシピ生成のGeminiスキーマを構築する
func buildRecipeSchema() *Schema {
	return &Schema{
		Type: SchemaTypeObject,
		Properties: map[string]*Schema{
			"recipe": {Type: SchemaTypeString},
		},
		Required: []string{"recipe"},
	}
}

// mealTypeToLabel は食事タイプを日本語ラベルに変換する
func mealTypeToLabel(mealType string) string {
	switch mealType {
	case "breakfast":
		return "朝食"
	case "lunch":
		return "昼食"
	case "dinner":
		return "夕食"
	case "snack":
		return "間食"
	default:
		return mealType
	}
}
