package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NormalizedEquipment は正規化された器具名を表す
type NormalizedEquipment struct {
	Original   string `json:"original"`   // 入力された名前
	Normalized string `json:"normalized"` // 正規化された名前
}

// EquipmentNormalizer は器具名を正規化するクライアント
type EquipmentNormalizer struct {
	client *Client
}

// NewEquipmentNormalizer は新しいEquipmentNormalizerを作成
func NewEquipmentNormalizer(timeout time.Duration) *EquipmentNormalizer {
	return &EquipmentNormalizer{
		client: NewClient(timeout),
	}
}

// NormalizeEquipmentNames は器具名を正規化する
func (en *EquipmentNormalizer) NormalizeEquipmentNames(ctx context.Context, names []string) ([]NormalizedEquipment, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("正規化する名前が指定されていません")
	}

	// 入力をJSON形式に変換
	inputJSON, err := json.Marshal(names)
	if err != nil {
		return nil, fmt.Errorf("入力のJSON変換に失敗: %w", err)
	}

	prompt := fmt.Sprintf(`以下のトレーニング器具名を正規化してください。

入力: %s

出力フォーマット（JSON配列）:
[
  {
    "original": "入力された名前",
    "normalized": "正規化された名前"
  }
]

正規化のルール:
- 表記揺れを統一（例: "ダンベル", "ダンベル（10kg）" → "ダンベル"）
- 一般的な呼び名に統一（例: "バーベル", "オリンピックバー" → "バーベル"）
- 器具のカテゴリを明確に（例: "チェストプレスマシン" → "チェストプレス"）
- 重量や詳細情報は除去（例: "20kgケトルベル" → "ケトルベル"）
- 格闘技関連の器具はそのまま維持（例: "サンドバッグ", "ミット", "グラップリングダミー"）
- 自重トレーニングは "自重" として統一
- JSON形式で出力してください`, string(inputJSON))

	response, err := en.client.Execute(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("Gemini APIによる器具名正規化に失敗: %w", err)
	}

	normalizedJSON := removeCodeBlock(response.Response)

	var normalized []NormalizedEquipment
	if err := json.Unmarshal([]byte(normalizedJSON), &normalized); err != nil {
		return nil, fmt.Errorf("正規化結果のパースエラー: %w\nデータ: %s", err, normalizedJSON)
	}

	return normalized, nil
}
