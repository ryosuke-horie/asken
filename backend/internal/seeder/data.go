package seeder

import (
	"math"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
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

// WeightSeedRecord は体重記録のシード用構造体
type WeightSeedRecord struct {
	Weight     float64
	RecordedAt string
}

// WeightSeedConfig は体重シードの設定
type WeightSeedConfig struct {
	StartWeight  float64
	TargetWeight float64
}

// DefaultWeightSeedConfig はデフォルトの体重シード設定
// 格闘技の減量を想定: 75kg → 68kg（約3ヶ月で7kg減量）
var DefaultWeightSeedConfig = WeightSeedConfig{
	StartWeight:  75.0,
	TargetWeight: 68.0,
}

// GenerateWeightRecords は指定日数分の体重記録データを生成する
// 減量トレンドを含むリアルなデータを生成
func GenerateWeightRecords(days int, config WeightSeedConfig) []WeightSeedRecord {
	if days <= 0 {
		return nil
	}

	records := make([]WeightSeedRecord, 0, days)
	now := time.Now()

	totalWeightLoss := config.StartWeight - config.TargetWeight
	dailyLoss := totalWeightLoss / float64(days)

	// 簡易的な擬似乱数生成（シード固定で再現可能）
	seed := int64(12345)
	nextRand := func() float64 {
		seed = (seed*1103515245 + 12345) & 0x7fffffff
		return float64(seed) / float64(0x7fffffff)
	}

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -(days - 1 - i))
		dayOfWeek := date.Weekday()

		// 週末は記録をスキップすることがある（約30%の確率）
		if (dayOfWeek == time.Saturday || dayOfWeek == time.Sunday) && nextRand() < 0.3 {
			continue
		}

		// ベースの体重（減量トレンド）
		baseWeight := config.StartWeight - (dailyLoss * float64(i))

		// 日々の変動を追加（±0.3kg）
		variation := (nextRand() - 0.5) * 0.6

		// 週末は少し増えやすい
		if dayOfWeek == time.Saturday || dayOfWeek == time.Sunday {
			variation += 0.2
		}

		weight := baseWeight + variation
		weight = math.Round(weight*10) / 10

		records = append(records, WeightSeedRecord{
			Weight:     weight,
			RecordedAt: FormatDateForDB(date),
		})
	}

	return records
}

// GetDefaultTargetDate は現在日から1ヶ月後の日付を目標日として返す
func GetDefaultTargetDate() string {
	return FormatDateForDB(time.Now().AddDate(0, 1, 0))
}

// MylistSeedItem はマイリストシード用の構造体
type MylistSeedItem struct {
	Name            string
	BaseAmount      string
	Unit            string
	Foods           []gemini.NutritionInfo
	SeedImageSource string // シード画像のファイル名（backend/seeds/images/内）
}

// DefaultMylistItems はデフォルトのマイリストアイテム
// 格闘技の減量に適した高タンパク・低カロリーなメニューを中心に構成
var DefaultMylistItems = []MylistSeedItem{
	{
		Name:            "鶏むね肉定食",
		BaseAmount:      "1人前",
		Unit:            "食",
		SeedImageSource: "chicken_breast_meal.jpg",
		Foods: []gemini.NutritionInfo{
			{Name: "鶏むね肉（皮なし）", EstimatedAmount: "150g", Calories: 165, Protein: 34.5, Fat: 2.3, Carbohydrates: 0.0},
			{Name: "白ご飯", EstimatedAmount: "1膳（150g）", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
			{Name: "ブロッコリー", EstimatedAmount: "50g", Calories: 17, Protein: 2.2, Fat: 0.3, Carbohydrates: 2.6},
		},
	},
	{
		Name:            "プロテインシェイク",
		BaseAmount:      "1杯",
		Unit:            "杯",
		SeedImageSource: "protein_shake.jpg",
		Foods: []gemini.NutritionInfo{
			{Name: "ホエイプロテイン", EstimatedAmount: "30g", Calories: 120, Protein: 24.0, Fat: 1.5, Carbohydrates: 3.0},
			{Name: "低脂肪牛乳", EstimatedAmount: "200ml", Calories: 92, Protein: 7.6, Fat: 2.0, Carbohydrates: 11.0},
		},
	},
	{
		Name:            "サラダチキン",
		BaseAmount:      "1パック",
		Unit:            "パック",
		SeedImageSource: "salad_chicken.jpg",
		Foods: []gemini.NutritionInfo{
			{Name: "サラダチキン", EstimatedAmount: "1パック（115g）", Calories: 127, Protein: 26.5, Fat: 1.6, Carbohydrates: 1.2},
		},
	},
	{
		Name:            "オートミール朝食",
		BaseAmount:      "1人前",
		Unit:            "食",
		SeedImageSource: "oatmeal_breakfast.jpg",
		Foods: []gemini.NutritionInfo{
			{Name: "オートミール", EstimatedAmount: "50g", Calories: 190, Protein: 6.9, Fat: 2.8, Carbohydrates: 34.5},
			{Name: "バナナ", EstimatedAmount: "1本（100g）", Calories: 86, Protein: 1.1, Fat: 0.2, Carbohydrates: 22.5},
			{Name: "はちみつ", EstimatedAmount: "10g", Calories: 30, Protein: 0.0, Fat: 0.0, Carbohydrates: 8.2},
		},
	},
	{
		Name:            "焼き鮭定食",
		BaseAmount:      "1人前",
		Unit:            "食",
		SeedImageSource: "grilled_salmon_meal.jpg",
		Foods: []gemini.NutritionInfo{
			{Name: "焼き鮭", EstimatedAmount: "1切れ（80g）", Calories: 156, Protein: 17.8, Fat: 8.9, Carbohydrates: 0.1},
			{Name: "白ご飯", EstimatedAmount: "1膳（150g）", Calories: 252, Protein: 3.8, Fat: 0.5, Carbohydrates: 55.7},
			{Name: "味噌汁", EstimatedAmount: "1杯（150ml）", Calories: 32, Protein: 2.4, Fat: 1.2, Carbohydrates: 2.8},
		},
	},
}

// CalculateMylistTotals はマイリストアイテムの栄養素合計を計算する
func CalculateMylistTotals(foods []gemini.NutritionInfo) (calories, protein, fat, carbs float64) {
	for _, food := range foods {
		calories += food.Calories
		protein += food.Protein
		fat += food.Fat
		carbs += food.Carbohydrates
	}
	return
}
