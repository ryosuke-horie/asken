# Gemini CLI 画像分析検証レポート

**検証日**: 2026-01-17
**担当者**: r-horie
**関連Linear課題**: [EDG-315](https://linear.app/edge-work/issue/EDG-315/uchikomi：geminicliをaiエージェント用途活用できるか検証)

---

## 1. 検証の目的

uchikomiプロジェクトにおいて、Gemini CLI（Gemini 3 API）をAIエージェント用途で活用できるか検証する。特に、**非対話的なモード**で画像から食材を判定し、カロリーや栄養素を推定する機能の実現可能性を確認する。

---

## 2. 検証内容

### 2.1 検証項目

- ✅ Gemini CLIの非対話モードでの実行
- ✅ 画像ファイル（JPG/HEIC）の入力
- ✅ JSON形式での出力制御
- ✅ 複数食材の同時認識
- ✅ カロリー・栄養素の推定精度
- ✅ Golangからのシェルコマンド実行
- ✅ エラーハンドリングとタイムアウト処理

### 2.2 テスト環境

- **OS**: macOS (Darwin 24.5.0)
- **Gemini CLI**: インストール済み（Homebrew経由）
- **Go**: 1.23
- **テスト画像**: 実際の食事画像3枚（JPG形式1枚、HEIC形式2枚）

---

## 3. 検証結果

### 3.1 成功したこと

#### ✅ 非対話モードでの画像分析

```bash
gemini "この画像に写っている食材や料理を特定し、それぞれのカロリーと主要な栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。 @./images/IMG_0374.JPG" -o json
```

- **結果**: 正常に動作
- **認識精度**: 高い（9種類の食材を正確に認識）
- **出力形式**: JSON形式で構造化されたデータ

#### ✅ JPG/HEIC両対応

- **JPG**: 問題なく処理
- **HEIC**: Appleの画像フォーマットも正常に処理

#### ✅ 複数食材の同時認識

テスト画像（居酒屋の前菜盛り合わせ）から以下を認識:

1. ブリの刺身
2. 白身魚のカルパッチョ
3. だし巻き卵
4. ハムと野菜のサラダ
5. ホタルイカ
6. もやしのナムル
7. 枝豆
8. ガリ（生姜甘酢漬け）
9. ネギの和え物

各食材について以下の情報を推定:

- 食材名
- 推定量
- カロリー（kcal）
- タンパク質（g）
- 脂質（g）
- 炭水化物（g）

#### ✅ Golangからの実行

`exec.CommandContext`を使用して、Golangから正常に実行できることを確認。

**サンプルコード**: `experiments/gemini-cli/scripts/main.go`

**主要機能**:
- タイムアウト処理（context.WithTimeout）
- エラーハンドリング
- JSON出力のパース
- レスポンス内のJSONコードブロック抽出

### 3.2 発見した制約事項

#### ⚠️ ワークスペース制限

Gemini CLIは、実行ディレクトリから相対的にアクセス可能な範囲にセキュリティ制限があります。

**問題**:
```
Error executing tool read_file: File path must be within one of the workspace directories
```

**解決策**:
- 画像ファイルを実行ディレクトリ配下に配置
- 絶対パスを使用

#### ⚠️ 標準出力の汚染

Gemini CLIは、JSON出力の前に `"Loaded cached credentials."` などのメッセージを出力します。

**解決策**:
```go
// JSON部分を抽出
jsonStart := bytes.IndexByte(output, '{')
jsonData := output[jsonStart:]
```

#### ⚠️ レスポンスのフォーマット

Gemini CLIは、レスポンス内でJSONを `\`\`\`json ... \`\`\`` で囲むことがあります。

**解決策**:
```go
// コードブロックを除去
if strings.Contains(response.Response, "```json") {
    // 抽出処理
}
```

### 3.3 パフォーマンス

**測定結果**（test_jpg_result.jsonより）:

- **gemini-2.5-flash-lite**: 5,085ms（約5秒）
- **gemini-3-pro-preview**: 34,957ms（約35秒）
- **合計実行時間**: 約40秒

**トークン使用量**:
- Input: 4,080 tokens (lite) + 6,358 tokens (pro)
- Output: 114 tokens (lite) + 923 tokens (pro)

**考察**:
- 実用的な速度ではあるが、リアルタイム処理には向かない
- 複数モデルが呼び出されているため、タイムアウトは60秒以上推奨

---

## 4. 実装方法

### 4.1 基本的なコマンド実行

```bash
gemini "プロンプト @画像パス" -o json
```

### 4.2 Golangからの実行

```go
package main

import (
    "context"
    "os/exec"
    "time"
)

func analyzeImage(imagePath string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    prompt := fmt.Sprintf("この画像に写っている食材や料理を特定し、それぞれのカロリーと主要な栄養素を推定してJSON形式で出力してください。 @%s", imagePath)

    cmd := exec.CommandContext(ctx, "gemini", "-o", "json", prompt)
    output, err := cmd.CombinedOutput()

    // JSONパース処理
    // ...

    return nil
}
```

**詳細**: `experiments/gemini-cli/scripts/main.go` を参照

---

## 5. 今後の課題

### 5.1 データベース連携

- PostgreSQLの食品マスタデータとGeminiの推定値を組み合わせる
- Geminiで食材名を特定 → DBから正確なカロリー情報を取得

### 5.2 プロンプトの最適化

現在のプロンプト:
```
この画像に写っている食材や料理を特定し、それぞれのカロリーと主要な栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。
```

**改善点**:
- JSON Schema を明示的に指定（フォーマットの統一）
- 量の推定精度を上げる指示
- 不確実性の表現（信頼度スコア）

### 5.3 エラーハンドリングの改善

- ネットワークエラーのリトライ処理
- APIレート制限の対応
- 画像フォーマットのバリデーション

### 5.4 パフォーマンス最適化

- キャッシング機構の導入
- 並列処理（複数画像の同時分析）
- 軽量モデルの選択（`gemini-2.5-flash-lite`のみ使用）

---

## 6. 結論

### ✅ 実現可能

Gemini CLIを使った**非対話的な画像分析**は十分に実現可能です。

**推奨される使用方法**:

1. **画像からの食材認識**: Gemini CLI
2. **正確なカロリー計算**: PostgreSQL（食品マスタ）
3. **バックエンド実装**: Golang（シェルコマンド実行）

**メリット**:

- ✅ Gemini 3 APIの高精度な画像認識
- ✅ JSON出力による構造化データ
- ✅ シンプルなシェルコマンド実行
- ✅ JPG/HEIC両対応

**デメリット**:

- ⚠️ 実行時間が長い（約40秒）
- ⚠️ ワークスペース制限がある
- ⚠️ 標準出力の後処理が必要

### 次のステップ

1. **プロトタイプの拡張**: Next.jsフロントエンドとの連携
2. **データベース設計**: 食品マスタテーブルの作成
3. **API設計**: Golangバックエンドでのエンドポイント実装
4. **ユーザーテスト**: 実際の食事画像での精度検証

---

## 7. 推奨: 2ステップアプローチ

### 7.1 概要

Gemini CLIを使う際は、**1回の実行につき1つのタスク**に集中させることで、精度と保守性が向上します。

**分離アプローチの利点**:
- ✅ 各ステップの責任が明確
- ✅ エラーハンドリングが簡単
- ✅ デバッグしやすい
- ✅ テストしやすい
- ✅ 各ステップの結果をキャッシュ可能
- ✅ PostgreSQLとの連携がスムーズ

### 7.2 処理フロー

```
┌─────────┐
│  画像   │
└────┬────┘
     │
     ▼
┌─────────────────────────────────┐
│ ステップ1: 食材分類            │
│ (Gemini CLI)                    │
│ - 食材名の特定                  │
│ - 量の推定                       │
└────┬────────────────────────────┘
     │ JSON: [{name, amount}]
     ▼
┌─────────────────────────────────┐
│ ステップ2: 栄養素算出          │
│ (PostgreSQL → Gemini CLI)       │
│ 1. DBで食材を検索               │
│ 2. ヒット → DB値を使用          │
│ 3. ミス → Geminiで推定          │
└────┬────────────────────────────┘
     │ JSON: [{name, amount, calories, nutrients}]
     ▼
┌─────────────────────────────────┐
│  最終結果                       │
└─────────────────────────────────┘
```

### 7.3 ステップ1: 食材分類

**目的**: 画像から食材名と推定量のみを抽出

**プロンプト**:
```
この画像に写っている食材や料理を特定し、各食材の名前と推定量（グラム数または個数）を
JSON形式のリストで出力してください。

出力フォーマット:
[
  {
    "name": "食材名",
    "estimated_amount": "推定量（例: 100g, 3切れ, 1杯）"
  }
]

カロリーや栄養素の情報は不要です。食材の特定と量の推定のみを行ってください。
```

**実行例**:
```bash
cd experiments/gemini-cli
go run scripts/step1_classify.go
```

**出力例**:
```json
[
  {
    "name": "刺身盛り合わせ（ブリまたはハマチ）",
    "estimated_amount": "8切れ"
  },
  {
    "name": "白身魚のカルパッチョ（玉ねぎ添え）",
    "estimated_amount": "1皿 (魚約10切れ、玉ねぎ約1/4個)"
  }
]
```

### 7.4 ステップ2: 栄養素算出

**目的**: 食材リストから栄養素とカロリーを算出

**処理の優先順位**:
1. **PostgreSQL検索** (最優先)
   - 食品マスタDBから正確なデータを取得
   - `Source: "database"`
2. **Gemini推定** (フォールバック)
   - DBにない食材のみGeminiで推定
   - `Source: "gemini"`

**プロンプト**:
```
以下の食材リストについて、それぞれのカロリーと栄養素（タンパク質、脂質、炭水化物）を
推定してJSON形式で出力してください。

食材リスト:
[...]

出力フォーマット:
[
  {
    "name": "食材名",
    "estimated_amount": "推定量",
    "calories_kcal": カロリー数値,
    "protein_g": タンパク質グラム数,
    "fat_g": 脂質グラム数,
    "carbohydrates_g": 炭水化物グラム数
  }
]

一般的な食品成分表に基づいて、妥当な値を推定してください。
```

**実行例**:
```bash
cd experiments/gemini-cli
go run scripts/step2_nutrition.go
```

**出力例**:
```json
[
  {
    "name": "刺身盛り合わせ（ブリまたはハマチ）",
    "estimated_amount": "8切れ",
    "calories_kcal": 360,
    "protein_g": 30.0,
    "fat_g": 24.6,
    "carbohydrates_g": 0.4,
    "source": "gemini"
  }
]
```

### 7.5 2ステップを連続実行

```bash
cd experiments/gemini-cli
./scripts/run_two_steps.sh
```

**実行時間**:
- ステップ1: 約30-40秒
- ステップ2: 約30-40秒
- 合計: 約60-80秒

### 7.6 実装サンプル

**ファイル構成**:
```
experiments/gemini-cli/scripts/
├── step1_classify.go      # 食材分類
├── step2_nutrition.go     # 栄養素算出
├── run_two_steps.sh       # 連続実行スクリプト
└── main.go                # 統合版（参考用）
```

**サンプルコード**: `experiments/gemini-cli/scripts/` を参照

### 7.7 PostgreSQL連携の実装イメージ

```go
// ステップ2の実装イメージ（PostgreSQL連携）
func CalculateNutrition(foods []FoodItem) ([]NutritionInfo, error) {
    var results []NutritionInfo

    for _, food := range foods {
        // まずDBで検索
        nutrition, err := SearchNutritionFromDatabase(food.Name, food.EstimatedAmount)
        if err == nil && nutrition != nil {
            nutrition.Source = "database"
            results = append(results, *nutrition)
            continue
        }

        // DBにない場合のみGeminiで推定
        nutrition, err = EstimateWithGemini(food)
        if err != nil {
            return nil, err
        }
        nutrition.Source = "gemini"
        results = append(results, *nutrition)
    }

    return results, nil
}
```

---

## 8. 付録

### 8.1 ディレクトリ構成

```
experiments/gemini-cli/
├── images/              # テスト画像
│   ├── IMG_0360.HEIC
│   ├── IMG_0369.HEIC
│   └── IMG_0374.JPG
├── results/             # 実行結果
│   ├── test_jpg_result.json
│   ├── test_heic_result.json
│   ├── golang_result.json
│   ├── step1_classify_result.json
│   └── step2_nutrition_result.json
├── scripts/             # サンプルコード
│   ├── main.go               # 統合版（参考用）
│   ├── step1_classify.go     # ステップ1: 食材分類
│   ├── step2_nutrition.go    # ステップ2: 栄養素算出
│   ├── run_two_steps.sh      # 2ステップ連続実行
│   └── go.mod
├── README.md            # クイックスタートガイド
└── REPORT.md            # このレポート
```

### 8.2 参考資料

- [Gemini CLI GitHub Repository](https://github.com/google-gemini/gemini-cli)
- [Gemini API Documentation](https://ai.google.dev/gemini-api/docs)
- [Google Gemini CLI Tutorial](https://dev.to/auden/google-gemini-cli-tutorial-how-to-install-and-use-it-with-images-4phb)

---

**検証完了**: 2026-01-17
