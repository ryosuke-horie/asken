package gemini

// MicronutrientKey はマイクロニュートリエントのキー型
type MicronutrientKey string

const (
	KeyIron       MicronutrientKey = "iron_mg"
	KeyCalcium    MicronutrientKey = "calcium_mg"
	KeyZinc       MicronutrientKey = "zinc_mg"
	KeyFiber      MicronutrientKey = "fiber_g"
	KeyVitaminA   MicronutrientKey = "vitamin_a_ug"
	KeyVitaminB1  MicronutrientKey = "vitamin_b1_mg"
	KeyVitaminB2  MicronutrientKey = "vitamin_b2_mg"
	KeyVitaminB6  MicronutrientKey = "vitamin_b6_mg"
	KeyVitaminB12 MicronutrientKey = "vitamin_b12_ug"
	KeyVitaminC   MicronutrientKey = "vitamin_c_mg"
	KeyVitaminD   MicronutrientKey = "vitamin_d_ug"
	KeyVitaminE   MicronutrientKey = "vitamin_e_mg"
)

// MicronutrientMeta はマイクロニュートリエントのメタデータ
type MicronutrientMeta struct {
	Key         MicronutrientKey
	DisplayName string
	Unit        string
	// 厚生労働省「日本人の食事摂取基準（2020年版）」に基づく成人推奨値
	DefaultTarget float64
}

// AllMicronutrients は全マイクロニュートリエントのメタデータ（表示順）
var AllMicronutrients = []MicronutrientMeta{
	{Key: KeyIron, DisplayName: "鉄分", Unit: "mg", DefaultTarget: 7.5},
	{Key: KeyCalcium, DisplayName: "カルシウム", Unit: "mg", DefaultTarget: 700},
	{Key: KeyZinc, DisplayName: "亜鉛", Unit: "mg", DefaultTarget: 10},
	{Key: KeyFiber, DisplayName: "食物繊維", Unit: "g", DefaultTarget: 21},
	{Key: KeyVitaminA, DisplayName: "ビタミンA", Unit: "μg", DefaultTarget: 800},
	{Key: KeyVitaminB1, DisplayName: "ビタミンB1", Unit: "mg", DefaultTarget: 1.3},
	{Key: KeyVitaminB2, DisplayName: "ビタミンB2", Unit: "mg", DefaultTarget: 1.5},
	{Key: KeyVitaminB6, DisplayName: "ビタミンB6", Unit: "mg", DefaultTarget: 1.3},
	{Key: KeyVitaminB12, DisplayName: "ビタミンB12", Unit: "μg", DefaultTarget: 2.4},
	{Key: KeyVitaminC, DisplayName: "ビタミンC", Unit: "mg", DefaultTarget: 100},
	{Key: KeyVitaminD, DisplayName: "ビタミンD", Unit: "μg", DefaultTarget: 8.5},
	{Key: KeyVitaminE, DisplayName: "ビタミンE", Unit: "mg", DefaultTarget: 6.5},
}

// AllMicronutrientKeys は全マイクロニュートリエントのキーを返す
func AllMicronutrientKeys() []string {
	keys := make([]string, len(AllMicronutrients))
	for i, m := range AllMicronutrients {
		keys[i] = string(m.Key)
	}
	return keys
}

// DefaultMicronutrientTargets は全マイクロニュートリエントのデフォルト目標値を返す
func DefaultMicronutrientTargets() map[string]float64 {
	targets := make(map[string]float64, len(AllMicronutrients))
	for _, m := range AllMicronutrients {
		targets[string(m.Key)] = m.DefaultTarget
	}
	return targets
}

// MergeMicronutrients は2つのmicronutrients mapを合算する
func MergeMicronutrients(dst, src map[string]float64) map[string]float64 {
	if dst == nil {
		dst = make(map[string]float64, len(src))
	}
	for k, v := range src {
		dst[k] += v
	}
	return dst
}
