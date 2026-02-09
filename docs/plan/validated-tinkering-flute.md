# プラン: 体重記録機能の実装

## Context

ウチコミは格闘技向けの体重管理アプリ。現在は食事記録機能のみ実装済み。
格闘技選手にとって減量管理は重要であり、体重の記録・グラフ化・目標管理機能を追加する。
iOSアプリにタブナビゲーションを導入し、食事記録と体重記録を切り替えられるようにする。

## 要件サマリー

| 項目 | 仕様 |
|:---|:---|
| 単位 | kg のみ（小数点1桁） |
| 記録回数 | 1日に複数回可能（朝・練習後・夜など） |
| グラフ期間 | 1週間 / 1ヶ月 / 3ヶ月 |
| 目標体重 | あり（グラフに目標ラインを表示） |
| 更新 | 1日の中で特定レコードを編集可能 |

---

## Phase 1: バックエンド（体重記録 CRUD + 目標体重）

### 1-1. Firestore データモデル

コレクション構造:
```
users/{userID}/weightRecords/{recordID}    # 体重記録
users/{userID}/weightGoal/current          # 目標体重（固定ドキュメントID）
```

体重記録ドキュメント:
```go
type firestoreWeightRecordDocument struct {
    ID         string    `firestore:"id"`          // UUID
    WeightKg   float64   `firestore:"weightKg"`    // 小数点1桁 (例: 65.3)
    RecordedAt time.Time `firestore:"recordedAt"`   // 記録日時 (UTC保存)
    Note       string    `firestore:"note,omitempty"`
    CreatedAt  time.Time `firestore:"createdAt"`
    UpdatedAt  time.Time `firestore:"updatedAt"`
}
```

目標体重ドキュメント:
```go
type firestoreWeightGoalDocument struct {
    TargetWeightKg float64   `firestore:"targetWeightKg"`
    UpdatedAt      time.Time `firestore:"updatedAt"`
}
```

### 1-2. Repository インターフェース

新規ファイル: `backend/internal/repository/weight_models.go`

```go
type WeightRecordRepository interface {
    CreateRecord(ctx context.Context, userID string, weightKg float64, recordedAt time.Time, note string) (*WeightRecord, error)
    GetRecord(ctx context.Context, userID string, recordID string) (*WeightRecord, error)
    UpdateRecord(ctx context.Context, userID string, recordID string, weightKg float64, note string) (*WeightRecord, error)
    DeleteRecord(ctx context.Context, userID string, recordID string) error
    ListRecordsByDateRange(ctx context.Context, userID string, from time.Time, to time.Time) ([]WeightRecord, error)
    GetGoal(ctx context.Context, userID string) (*WeightGoal, error)
    SetGoal(ctx context.Context, userID string, targetWeightKg float64) (*WeightGoal, error)
}
```

Firestore実装: `backend/internal/repository/weight_repository_firestore.go`

### 1-3. API エンドポイント

| メソッド | パス | 説明 |
|:---|:---|:---|
| POST | `/api/weight/records` | 体重記録の作成 |
| GET | `/api/weight/records?from=&to=&tz=` | 期間指定の体重記録取得（グラフ+一覧） |
| PUT | `/api/weight/records/{id}` | 体重記録の更新 |
| DELETE | `/api/weight/records/{id}` | 体重記録の削除 |
| GET | `/api/weight/goal` | 目標体重の取得 |
| PUT | `/api/weight/goal` | 目標体重の設定・更新 |

POST リクエスト例:
```json
{
    "weight_kg": 65.3,
    "recorded_at": "2026-02-08T07:30:00+09:00",
    "note": "朝食前"
}
```

GET `/api/weight/records` レスポンス例:
```json
{
    "records": [
        {
            "id": "uuid",
            "weight_kg": 65.3,
            "recorded_at": "2026-02-08T07:30:00+09:00",
            "note": "朝食前",
            "created_at": "2026-02-08T07:35:00Z",
            "updated_at": "2026-02-08T07:35:00Z"
        }
    ],
    "daily_summary": {
        "2026-02-08": { "latest_weight": 65.3, "count": 2 },
        "2026-02-07": { "latest_weight": 65.5, "count": 1 }
    },
    "goal": {
        "target_weight_kg": 63.0,
        "updated_at": "2026-01-15T10:00:00Z"
    }
}
```

バリデーション:
- `weight_kg`: 20.0 ~ 300.0、小数点1桁に丸め
- `recorded_at`: RFC3339形式、未来日時不可（+5分まで許容）
- `note`: 最大200文字
- `from`/`to`: `2006-01-02` 形式、最大365日間

### 1-4. Handler 設計

新規ファイル:
- `backend/internal/handler/weight_record_handler.go` - CRUD
- `backend/internal/handler/weight_goal_handler.go` - 目標体重

ルーティング追加（`cmd/server/main.go`）:
```go
func setupWeightRoutes(mux *http.ServeMux, h handlers, authMiddleware middleware.Authenticator) {
    mux.Handle("/api/weight/records", authMiddleware.Authenticate(...))   // GET(list), POST
    mux.Handle("/api/weight/records/", authMiddleware.Authenticate(...))  // GET(detail), PUT, DELETE
    mux.Handle("/api/weight/goal", authMiddleware.Authenticate(...))      // GET, PUT
}
```

### 1-5. バックエンド ファイル一覧

| ファイル | 内容 |
|:---|:---|
| `backend/internal/repository/weight_models.go` | インターフェース + モデル定義 |
| `backend/internal/repository/weight_repository_firestore.go` | Firestore実装 |
| `backend/internal/repository/weight_repository_firestore_test.go` | エミュレータテスト |
| `backend/internal/handler/weight_record_handler.go` | CRUD ハンドラー |
| `backend/internal/handler/weight_record_handler_test.go` | ハンドラーテスト |
| `backend/internal/handler/weight_goal_handler.go` | 目標体重ハンドラー |
| `backend/internal/handler/weight_goal_handler_test.go` | ハンドラーテスト |
| `backend/cmd/server/main.go` | ルーティング + DI 追加 |

---

## Phase 2: iOS（タブ導入 + 体重記録画面）

### 2-1. MainTabView のタブ化

変更ファイル: `ios/Uchikomi/App/UchikomiApp.swift` 内の `MainTabView`

```swift
struct MainTabView: View {
    var body: some View {
        TabView {
            MealsView()
                .tabItem { Label("食事", systemImage: "fork.knife") }
            WeightView()
                .tabItem { Label("体重", systemImage: "scalemass") }
        }
        .tint(Theme.primary)
    }
}
```

### 2-2. 画面構成

```
WeightView (メイン画面)
  |-- 目標カード（現在体重・目標・差分）
  |-- グラフ（Swift Charts + 期間セグメント）
  |-- 今日の記録一覧
  |
  +-- [+] → WeightInputView (シート: 入力)
  +-- 記録タップ → WeightInputView (シート: 編集)
  +-- 目標アイコン → WeightGoalSheet (シート: 目標設定)
```

### 2-3. iOS ファイル構成

```
ios/Uchikomi/Features/Weight/
  WeightView.swift              # メイン画面
  WeightViewModel.swift         # メインVM
  WeightInputView.swift         # 入力/編集シート
  WeightInputViewModel.swift    # 入力VM
  Views/
    WeightGoalCard.swift        # 目標カード
    WeightChartView.swift       # Swift Charts グラフ
    WeightRecordRow.swift       # 記録行
    WeightGoalSheet.swift       # 目標設定シート
  Models/
    WeightRecord.swift          # 体重記録モデル
    WeightGoal.swift            # 目標モデル
    ChartPeriod.swift           # 期間enum

ios/Uchikomi/Core/Repositories/
  WeightRepository.swift        # Repository

ios/Uchikomi/Core/Network/
  APIEndpoint.swift             # エンドポイント追加（既存ファイル）

ios/UchikomiTests/Features/Weight/
  WeightViewModelTests.swift
  WeightInputViewModelTests.swift
```

### 2-4. Model 定義

```swift
struct WeightRecord: Codable, Identifiable {
    let id: String
    let weightKg: Double
    let recordedAt: String
    let note: String?
    let createdAt: String
    let updatedAt: String
}

struct WeightGoal: Codable {
    let targetWeightKg: Double
    let updatedAt: String
}

enum ChartPeriod: String, CaseIterable, Identifiable {
    case week, month, threeMonths
    var id: String { rawValue }
    var displayName: String { ... }  // "1週間", "1ヶ月", "3ヶ月"
    var days: Int { ... }            // 7, 30, 90
}
```

### 2-5. ViewModel 設計

WeightViewModel (メイン):
- `todayRecords: [WeightRecord]` - 今日の記録
- `chartRecords: [WeightRecord]` - グラフ用データ
- `goal: WeightGoal?` - 目標体重
- `selectedPeriod: ChartPeriod` - グラフ期間
- `latestWeight: Double?` - 最新体重（computed）
- `weightDifferenceFromGoal: Double?` - 目標との差分（computed）
- `loadData() async` / `loadChartData() async` / `deleteRecord() async`

WeightInputViewModel (入力/編集):
- `weightText: String` - 入力値
- `memo: String` - メモ
- `incrementWeight()` / `decrementWeight()` - +/-0.1 微調整
- `save() async` / `delete() async`

### 2-6. Swift Charts グラフ

- `LineMark` + `PointMark` で体重推移を描画
- `RuleMark` で目標体重の赤い破線を表示
- 期間に応じた X軸ラベル（日/週/月）
- Y軸は体重データ + 目標体重の範囲を自動計算

### 2-7. 体重入力UI の工夫

- Decimal Pad で素早い数字入力
- +0.1 / -0.1 の微調整ボタン
- クイック選択チップ: 「起床時」「練習前」「練習後」「就寝前」
- 前回値をプレースホルダーに表示
- `@FocusState` で起動時にキーボード自動表示

---

## 実装順序

| 順序 | 内容 | 依存 |
|:---|:---|:---|
| 1 | Linear Issue 作成 | - |
| 2 | バックエンド: Repository インターフェース + モデル定義 | - |
| 3 | バックエンド: Firestore 実装 + テスト（TDD） | 2 |
| 4 | バックエンド: Handler 実装 + テスト（TDD） | 3 |
| 5 | バックエンド: main.go ルーティング + 動作確認 | 4 |
| 6 | iOS: Model 定義 + Repository + APIEndpoint | 5 |
| 7 | iOS: WeightInputViewModel + テスト（TDD） | 6 |
| 8 | iOS: WeightViewModel + テスト（TDD） | 6 |
| 9 | iOS: View 実装（入力、グラフ、目標、メイン画面） | 7, 8 |
| 10 | iOS: MainTabView タブ化 | 9 |
| 11 | 結合テスト + ブラウザ動作確認 | 10 |

---

## 参照すべき既存ファイル（パターン踏襲元）

| 用途 | ファイル |
|:---|:---|
| Firestore Repository パターン | `backend/internal/repository/analysis_repository_firestore.go` |
| Handler パターン | `backend/internal/handler/daily_meals_handler.go` |
| ルーティング設定 | `backend/cmd/server/main.go` |
| タイムゾーン処理 | `backend/internal/util/timezone.go` |
| iOS ViewModel パターン | `ios/Uchikomi/Features/Meals/MealsViewModel.swift` |
| iOS View パターン | `ios/Uchikomi/Features/Meals/MealsView.swift` |
| iOS Repository パターン | `ios/Uchikomi/Core/Repositories/MealRepository.swift` |
| iOS APIEndpoint パターン | `ios/Uchikomi/Core/Network/APIEndpoint.swift` |

## 検証方法

1. バックエンド: `task test` で全テスト合格
2. バックエンド: `task run` で起動 → Chrome DevTools で API 動作確認
3. iOS: `task ios:test` で全テスト合格
4. iOS: シミュレーターで体重入力 → グラフ表示 → 編集 → 削除の一連フロー確認
5. 結合: iOS → バックエンド API → Firestore の E2E 動作確認
