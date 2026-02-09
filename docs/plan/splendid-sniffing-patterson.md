# プラン: PR #145 2回目レビュー指摘事項の修正

## Linear Issue
- Issue: EDG-608
- PR: #145 (feat: 体重記録機能を追加)

## 概要

2回目のPRレビュー（5エージェント並列実行）で検出された指摘事項を修正する。
Critical 2件、Important 8件、Suggestion 8件のうち全件を対応する。

## Phase 1: Backend修正

### 1-1. Firestoreインデックス追加 (C1)

`firestore.indexes.json` にweightRecordsサブコレクション用のインデックスを追加。
単一フィールドの範囲クエリは自動インデックスで動作するが、明示的に定義する。

```json
{
  "collectionGroup": "weightRecords",
  "queryScope": "COLLECTION",
  "fields": [
    { "fieldPath": "recordedAt", "order": "ASCENDING" }
  ]
}
```

### 1-2. GoalレスポンスにroundToOneDecimalForJSON適用 (I4)

`weight_goal_handler.go`: HandleGet(56-58行)とHandleSet(102-105行)の2箇所で
`goal.TargetWeightKg` を `roundToOneDecimalForJSON(goal.TargetWeightKg)` に変更。

`roundToOneDecimalForJSON` は `weight_record_handler.go` にあるため、同パッケージ内で参照可能。

### 1-3. コンストラクタにnilガード追加 (I7)

`weight_record_handler.go` の `NewWeightRecordHandler`:
```go
func NewWeightRecordHandler(repo repository.WeightRecordRepository, goalRepo repository.WeightGoalRepository) *WeightRecordHandler {
    if repo == nil || goalRepo == nil {
        panic("weight record handler: repositories must not be nil")
    }
    return &WeightRecordHandler{repository: repo, goalRepository: goalRepo}
}
```

`weight_goal_handler.go` の `NewWeightGoalHandler`:
```go
func NewWeightGoalHandler(repository repository.WeightGoalRepository) *WeightGoalHandler {
    if repository == nil {
        panic("weight goal handler: repository must not be nil")
    }
    return &WeightGoalHandler{repository: repository}
}
```

### 1-4. リポジトリ層にバリデーション追加 (I8)

`weight_repository_firestore.go` の `CreateRecord`, `UpdateRecord`, `SetGoal` で体重範囲チェック。
バリデーション関数をrepositoryパッケージに定義（handlerからの重複を避けるため定数も共有）。

`weight_models.go` に定数とバリデーション関数を追加:
```go
const (
    MinWeightKg = 20.0
    MaxWeightKg = 300.0
)

func ValidateWeightKg(weightKg float64) error {
    if math.IsNaN(weightKg) || math.IsInf(weightKg, 0) {
        return fmt.Errorf("weight_kgに無効な値が指定されています")
    }
    if weightKg < MinWeightKg || weightKg > MaxWeightKg {
        return fmt.Errorf("weight_kgは%.1f〜%.1fの範囲で指定してください", MinWeightKg, MaxWeightKg)
    }
    return nil
}
```

`weight_record_handler.go` の `validateWeightKg` は `repository.ValidateWeightKg` を呼び出す形に変更。
リポジトリの `CreateRecord`, `UpdateRecord`, `SetGoal` の先頭に `ValidateWeightKg` 呼び出しを追加。

### 1-5. HandleListのGoal取得失敗を非致命的に変更 (I2)

`weight_record_handler.go` HandleList (119-125行): Goal取得失敗時に500を返す代わりに、
ログを出力してnilのまま続行する:

```go
goal, err := h.goalRepository.GetGoal(r.Context(), userID)
if err != nil {
    log.Printf("Warning: failed to get weight goal, continuing without it: userID=%s, error=%v", userID, err)
    // goal remains nil, response.Goal will be null
}
```

### 1-6. extractRecordIDの改善 (S1)

`weight_record_handler.go`: `extractRecordID` を `(string, error)` 返却に変更し、
「ID未指定」と「UUID不正」を区別:

```go
func extractRecordID(path string) (string, error) {
    parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
    if len(parts) < 4 || parts[len(parts)-1] == "" {
        return "", fmt.Errorf("記録IDが指定されていません")
    }
    id := parts[len(parts)-1]
    if _, err := uuid.Parse(id); err != nil {
        return "", fmt.Errorf("記録IDの形式が不正です")
    }
    return id, nil
}
```

HandleGet, HandleUpdate, HandleDelete の呼び出し箇所を更新。

### 1-7. JSONデコード失敗時のログ追加 (S2)

`weight_record_handler.go` HandleCreate, HandleUpdate、`weight_goal_handler.go` HandleSet の3箇所:

```go
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    log.Printf("Error decoding request: userID=%s, error=%v", userID, err)
    http.Error(w, "リクエストのパースに失敗しました", http.StatusBadRequest)
    return
}
```

### 1-8. NewWeightRepositoriesコメント改善 (S7)

`weight_repository_firestore.go` 37行:
```go
// NewWeightRepositories は新しいFirestoreベースのリポジトリを作成します
// 返される WeightRecordRepository と WeightGoalRepository は同一インスタンスです
```

---

## Phase 2: Backendテスト追加 (S3, S4, S5)

### 2-1. HandleListの無効timezone/無効日付テスト (S3)

`weight_record_handler_test.go` に追加:
- 無効タイムゾーン (`tz=Invalid/Zone`) → 400
- 無効from日付 (`from=not-a-date`) → 400
- 無効to日付 (`to=invalid`) → 400

### 2-2. 不正JSONボディテスト (S4)

Create, Update, SetGoalの3エンドポイントに不正JSONボディテストを追加:
- `{invalid json}` → 400

### 2-3. validateWeightKg境界値テスト (S5)

既存のバリデーションテストテーブルに境界値を追加:
- 19.9 → reject
- 20.0 → accept
- 300.0 → accept
- 300.1 → reject

---

## Phase 3: iOS修正

### 3-1. loadChartRecords()のエラー伝播 (C2)

`WeightViewModel.swift` 89-94行: 空配列返却をthrowに変更:

```swift
private func loadChartRecords() async throws -> [WeightRecord] {
    let to = Date()
    guard let from = Calendar.current.date(byAdding: .day, value: -selectedPeriod.days, to: to) else {
        logger.error("チャート期間の日付計算に失敗: days=\(self.selectedPeriod.days)")
        throw NSError(domain: "WeightViewModel", code: 0,
                      userInfo: [NSLocalizedDescriptionKey: "チャート期間の計算に失敗しました"])
    }
    let response = try await repository.getRecords(from: from, to: to)
    return response.records
}
```

### 3-2. WeightGoalSheetにLogger追加 (I1)

`WeightGoalSheet.swift`:
- `import os` 追加
- `private let logger = Logger(...)` をファイルトップに追加
- `save()` の両catchブロックに `logger.error(...)` 追加

### 3-3. lastWeightパラメータ削除 (I3)

`WeightInputView.swift` 14行: `lastWeight _: Double? = nil` パラメータを削除。
`WeightView.swift` 90行: `WeightInputView(lastWeight: viewModel.latestWeight)` から `lastWeight:` 引数を削除。

### 3-4. todayRecords重複APIリクエスト最適化 (I5)

`WeightViewModel.swift`: `loadTodayRecords()` を削除し、`loadData()` でchartRecordsから今日分をフィルタリング。

ISO8601日付パース用のstaticフォーマッタを`WeightViewModel`に追加:
```swift
private static let iso8601Formatter: ISO8601DateFormatter = {
    let f = ISO8601DateFormatter()
    f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
    return f
}()

private static let iso8601FallbackFormatter: ISO8601DateFormatter = {
    let f = ISO8601DateFormatter()
    f.formatOptions = [.withInternetDateTime]
    return f
}()
```

`loadData()` を変更:
```swift
func loadData() async {
    isLoading = true
    errorMessage = nil

    do {
        async let chartData = loadChartRecords()
        async let goalData = repository.getGoal()

        let (chart, fetchedGoal) = try await (chartData, goalData)
        chartRecords = chart
        goal = fetchedGoal
        todayRecords = filterTodayRecords(from: chart)
    } catch ...
}
```

`filterTodayRecords` メソッドを追加:
```swift
private func filterTodayRecords(from records: [WeightRecord]) -> [WeightRecord] {
    let calendar = Calendar.current
    let today = Date()
    return records.filter { record in
        guard let date = Self.parseISO8601(record.recordedAt) else { return false }
        return calendar.isDate(date, inSameDayAs: today)
    }
}

private static func parseISO8601(_ string: String) -> Date? {
    iso8601Formatter.date(from: string) ?? iso8601FallbackFormatter.date(from: string)
}
```

### 3-5. DateFormatter最適化 (I6)

`WeightRecordRow.swift` と `WeightChartView.swift` でcomputed property内のISO8601DateFormatter生成をstatic定義に変更。

`WeightViewModel`に追加したstatic parseISO8601を再利用するか、各Viewでstatic定義するか:
→ 各Viewが独立してstatic定義する（Viewの自己完結性を保つ）

`WeightRecordRow.swift`:
```swift
private static let iso8601Formatter: ISO8601DateFormatter = { ... }()
private static let iso8601FallbackFormatter: ISO8601DateFormatter = { ... }()
private static let timeFormatter: DateFormatter = {
    let f = DateFormatter()
    f.dateFormat = "HH:mm"
    f.timeZone = TimeZone.current
    return f
}()
```

`WeightChartView.swift` も同様にstatic定義。

### 3-6. incrementWeight上限境界テスト (S6)

`WeightInputViewModelTests.swift`:
```swift
@Test
@MainActor
func incrementWeightで300kgを超えないべき() {
    let mockRepo = WeightRepositoryProtocolMock()
    let viewModel = WeightInputViewModel(repository: mockRepo)
    viewModel.weightText = "300.0"
    viewModel.incrementWeight()
    #expect(viewModel.weightText == "300.0")
}
```

---

## 変更ファイル一覧

| ファイル | Phase | 対応Issue |
|:---|:---|:---|
| `firestore.indexes.json` | 1 | C1 |
| `backend/internal/repository/weight_models.go` | 1 | I8 (定数+バリデーション関数) |
| `backend/internal/repository/weight_repository_firestore.go` | 1 | I8, S7 |
| `backend/internal/handler/weight_record_handler.go` | 1 | I2, I7, I8, S1, S2 |
| `backend/internal/handler/weight_goal_handler.go` | 1 | I4, I7, S2 |
| `backend/internal/handler/weight_record_handler_test.go` | 2 | S3, S4, S5 |
| `backend/internal/handler/weight_goal_handler_test.go` | 2 | S4 |
| `ios/Uchikomi/Features/Weight/WeightViewModel.swift` | 3 | C2, I5 |
| `ios/Uchikomi/Features/Weight/WeightInputView.swift` | 3 | I3 |
| `ios/Uchikomi/Features/Weight/WeightView.swift` | 3 | I3 |
| `ios/Uchikomi/Features/Weight/Views/WeightGoalSheet.swift` | 3 | I1 |
| `ios/Uchikomi/Features/Weight/Views/WeightRecordRow.swift` | 3 | I6 |
| `ios/Uchikomi/Features/Weight/Views/WeightChartView.swift` | 3 | I6 |
| `ios/UchikomiTests/Features/Weight/WeightInputViewModelTests.swift` | 3 | S6 |
| `ios/UchikomiTests/Features/Weight/WeightViewModelTests.swift` | 3 | I5テスト更新 |

## テスト計画

1. `task test` でGoバックエンドテスト全パス確認
2. `task lint` でGoリントクリア確認
3. `task ios:format` でSwiftFormat適用
4. Chrome DevTools MCPでAPI動作確認
