# プラン: 食事記録の手入力機能

## Linear Issue
- Issue: EDG-592
- URL: https://linear.app/ryosuke-horie/issue/EDG-592

## 概要

画像認識以外の入力方法を追加し、最初から食材を手入力できるUIを実装する。
保存後にバックエンドで栄養素を自動計算する。

## 設計方針

MealInputViewを改修して「手入力モード」と「画像入力モード」を両方サポートする。

理由:
- 既存コンポーネント（NutritionEditorViewへの遷移、ポーリング）を再利用可能
- コード重複を避けられる
- 同じ画面で入力方法を選択できる

## UI設計

```
MealInputView（改修後）
├── ExistingMealsSection（既存記録）
├── "新しく記録する" セクション
│   ├── [テキストで入力] ボタン
│   └── [画像で入力] ボタン
├── テキスト入力セクション（選択時のみ表示）
│   ├── TextEditor
│   └── [分析] ボタン
├── 画像選択セクション（選択時のみ表示、既存）
├── 分析結果 → NutritionEditorView
└── エラー表示
```

## 実装計画

### Step 1: MealRepositoryProtocol にメソッド定義追加

ファイル: `ios/Uchikomi/Core/Repositories/MealRepository.swift`

```swift
// Protocol に追加（6-14行目付近）
protocol MealRepositoryProtocol {
    // 既存メソッド...
    func analyzeText(inputText: String, mealType: MealType, mealDate: Date) async throws -> String  // NEW
}
```

### Step 2: TextAnalyzeRequest モデル定義

ファイル: `ios/Uchikomi/Core/Repositories/MealRepository.swift`（または別ファイル）

```swift
struct TextAnalyzeRequest: Encodable {
    let inputText: String
    let mealType: String
    let mealDate: String
    let tz: String  // タイムゾーン（画像アップロードと同じ形式）

    enum CodingKeys: String, CodingKey {
        case inputText = "input_text"
        case mealType = "meal_type"
        case mealDate = "meal_date"
        case tz
    }
}
```

### Step 3: MealRepository にメソッド実装

```swift
func analyzeText(inputText: String, mealType: MealType, mealDate: Date) async throws -> String {
    let request = TextAnalyzeRequest(
        inputText: inputText,
        mealType: mealType.rawValue,
        mealDate: dateFormatter.string(from: mealDate),
        tz: TimeZone.current.identifier
    )
    let response: AnalyzeResponse = try await apiClient.request(endpoint: .analyze, body: request)
    return response.id
}
```

### Step 4: MealInputViewModel にテキスト分析機能追加

ファイル: `ios/Uchikomi/Features/Meals/MealInputViewModel.swift`

```swift
@Published var inputText: String = ""

func analyzeText() async {
    // 1. バリデーション
    guard !inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
        errorMessage = "食事内容を入力してください"
        return
    }
    guard inputText.count <= 1000 else {
        errorMessage = "入力は1000文字以内にしてください"
        return
    }

    isAnalyzing = true
    errorMessage = nil

    do {
        // 2. API呼び出し
        let id = try await repository.analyzeText(
            inputText: inputText,
            mealType: selectedMealType,
            mealDate: mealDate
        )
        analysisId = id

        // 3. ポーリング（既存ロジック再利用）
        try await pollForCompletion(id: id)

        // 4. 結果取得 → エディタ表示
        analysisResult = try await repository.getAnalysisResult(id: id)
        showEditor = true
    } catch {
        errorMessage = "テキスト分析に失敗しました: \(error.localizedDescription)"
    }

    isAnalyzing = false
}
```

注意: 既存の `reset()` メソッドは画像関連の状態のみリセット。`inputText` は保持される。

### Step 5: MealInputView のUI改修

ファイル: `ios/Uchikomi/Features/Meals/MealInputView.swift`

```swift
private enum InputMode {
    case selection  // 初期状態（どちらかを選ぶ）
    case text       // テキスト入力モード
    case image      // 画像入力モード
}

@State private var inputMode: InputMode = .selection
```

変更内容:
1. InputMode enum を追加
2. 入力モード選択UI（2つのボタン）を追加
3. TextInputSection を新規作成（TextEditor + 分析ボタン）
4. inputMode に応じて表示を切り替え
5. 「キャンセル」ボタンで inputMode = .selection に戻る（入力内容は保持）

### Step 6: MockMealRepositoryProtocol の更新

ファイル: `ios/UchikomiTests/Mocks/MockMealRepositoryProtocol.swift`（Mockolo生成）

```swift
// mockoloで自動生成されるが、handlerを追加
var analyzeTextHandler: ((String, MealType, Date) async throws -> String)?

func analyzeText(inputText: String, mealType: MealType, mealDate: Date) async throws -> String {
    guard let handler = analyzeTextHandler else {
        fatalError("analyzeTextHandler not set")
    }
    return try await handler(inputText, mealType, mealDate)
}
```

## 変更対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Core/Repositories/MealRepository.swift` | Protocol + analyzeText() + TextAnalyzeRequest |
| `ios/Uchikomi/Features/Meals/MealInputViewModel.swift` | inputText, analyzeText() 追加 |
| `ios/Uchikomi/Features/Meals/MealInputView.swift` | InputMode enum, TextInputSection 追加 |
| `ios/UchikomiTests/Features/Meals/MealInputViewModelTests.swift` | テスト追加（Swift Testing） |
| `ios/UchikomiTests/Mocks/` | Mockolo再生成（`task ios:generate-mocks`）|

## テスト計画

### ユニットテスト（TDD、Swift Testing フレームワーク）

ファイル: `ios/UchikomiTests/Features/Meals/MealInputViewModelTests.swift`

```swift
import Testing
@testable import Uchikomi

@Suite struct MealInputViewModelTests {
    @Test func 空文字でテキスト分析するとエラーになるべき() async {
        let viewModel = MealInputViewModel()
        viewModel.inputText = "   "

        await viewModel.analyzeText()

        #expect(viewModel.errorMessage == "食事内容を入力してください")
        #expect(viewModel.isAnalyzing == false)
    }

    @Test func 1000文字超過でエラーになるべき() async {
        let viewModel = MealInputViewModel()
        viewModel.inputText = String(repeating: "あ", count: 1001)

        await viewModel.analyzeText()

        #expect(viewModel.errorMessage == "入力は1000文字以内にしてください")
    }

    @Test func テキスト分析成功時にエディタを表示すべき() async throws {
        let mockRepo = MockMealRepositoryProtocol()
        mockRepo.analyzeTextHandler = { _, _, _ in "test-id" }
        mockRepo.checkAnalysisStatusHandler = { _ in
            AnalysisStatusResponse(status: "completed", error: nil, message: nil)
        }
        mockRepo.getAnalysisResultHandler = { _ in
            AnalysisResultResponse(status: "completed", result: AnalysisResult(...))
        }

        let viewModel = MealInputViewModel(repository: mockRepo)
        viewModel.inputText = "鶏むね肉100g"

        await viewModel.analyzeText()

        #expect(viewModel.showEditor == true)
        #expect(viewModel.analysisId == "test-id")
    }
}
```

### 手動検証（Chrome DevTools MCP）

1. MealsView → 食事タイプをタップ → MealInputView を開く
2. 「テキストで入力」ボタンをタップ
3. TextEditor に食事内容を入力（例: 鶏むね肉100g、ご飯1杯）
4. 「分析」ボタンをタップ
5. ローディング表示 → ポーリング完了
6. NutritionEditorView が開く
7. 栄養素が計算されていることを確認
8. 「保存」→ 一覧に反映されることを確認
9. コンソールエラーがないことを確認

## エッジケース

- 空白のみの入力 → エラーメッセージ表示
- 1000文字超過 → バックエンドでエラー（iOS側でも事前チェック推奨）
- ネットワークエラー → 既存のエラーハンドリング
- 分析タイムアウト → 既存のポーリングタイムアウト（120秒）
- 入力モード切り替え → 入力内容は保持（リセットしない）

## 技術的な考慮事項

- バックエンド変更不要（POST /api/analyze でテキスト入力対応済み）
- 既存の pollForCompletion() ロジックを完全再利用
- NutritionEditorView は空の foods で初期化可能（既存仕様）
