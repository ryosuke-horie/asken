package seeder

import (
	"time"

	"github.com/ryosuke-horie/asken/backend/pkg/gemini"
)

// TestUser はテストユーザー情報を表す構造体
type TestUser struct {
	Email    string
	Password string
	Name     string
}

// DefaultTestUsers はデフォルトのテストユーザーリスト
var DefaultTestUsers = []TestUser{
	{Email: "test@example.com", Password: "Pass0123", Name: "テストユーザー"},
	{Email: "test2@example.com", Password: "Pass0123", Name: "テストユーザー2"},
	{Email: "test3@example.com", Password: "Pass0123", Name: "テストユーザー3"},
}

// MealTypes は食事タイプのリスト
var MealTypes = []string{"breakfast", "lunch", "dinner", "snack"}

// SampleNutritionData は食事タイプごとのサンプル栄養素データ
var SampleNutritionData = map[string][]gemini.NutritionInfo{
	"breakfast": {
		{Name: "トースト", EstimatedAmount: "1枚（60g）", Calories: 158, Protein: 5.6, Fat: 2.6, Carbohydrates: 28.0},
		{Name: "目玉焼き", EstimatedAmount: "1個（55g）", Calories: 91, Protein: 6.5, Fat: 6.9, Carbohydrates: 0.2},
		{Name: "コーヒー（ブラック）", EstimatedAmount: "1杯（200ml）", Calories: 8, Protein: 0.4, Fat: 0.0, Carbohydrates: 1.4},
	},
	"lunch": {
		{Name: "親子丼", EstimatedAmount: "1人前（400g）", Calories: 652, Protein: 28.5, Fat: 15.2, Carbohydrates: 92.0},
		{Name: "味噌汁", EstimatedAmount: "1杯（150ml）", Calories: 32, Protein: 2.4, Fat: 1.2, Carbohydrates: 2.8},
	},
	"dinner": {
		{Name: "焼き鮭", EstimatedAmount: "1切れ（80g）", Calories: 156, Protein: 17.8, Fat: 8.9, Carbohydrates: 0.1},
		{Name: "白ご飯", EstimatedAmount: "1膳（150g）", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
		{Name: "ほうれん草のおひたし", EstimatedAmount: "小鉢（50g）", Calories: 18, Protein: 1.5, Fat: 0.2, Carbohydrates: 2.4},
		{Name: "豆腐の味噌汁", EstimatedAmount: "1杯（150ml）", Calories: 45, Protein: 3.8, Fat: 2.1, Carbohydrates: 2.5},
	},
	"snack": {
		{Name: "おにぎり（鮭）", EstimatedAmount: "1個（110g）", Calories: 183, Protein: 5.2, Fat: 1.8, Carbohydrates: 36.5},
	},
}

// SampleTextInputs はテキスト入力のサンプル
var SampleTextInputs = []string{
	"ご飯、味噌汁、焼き魚",
	"サンドイッチとサラダ",
	"カレーライス大盛り",
	"うどんと天ぷら",
	"ラーメン",
}

// GeneratePastDates は過去N日間の日付リストを生成する
func GeneratePastDates(days int) []time.Time {
	dates := make([]time.Time, days)
	now := time.Now()

	for i := 0; i < days; i++ {
		dates[i] = now.AddDate(0, 0, -i)
	}

	return dates
}

// FormatDateForDB はDB保存用の日付フォーマットに変換する
func FormatDateForDB(t time.Time) string {
	return t.Format("2006-01-02")
}

// CalculateTotals は栄養素の合計を計算する
func CalculateTotals(foods []gemini.NutritionInfo) (calories, protein, fat, carbs float64) {
	for _, food := range foods {
		calories += food.Calories
		protein += food.Protein
		fat += food.Fat
		carbs += food.Carbohydrates
	}
	return
}
