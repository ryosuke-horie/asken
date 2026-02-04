# プラン: PRレビュー指摘事項の修正

## Linear Issue
- Issue: EDG-592
- URL: https://linear.app/ryosuke-horie/issue/EDG-592

## 概要

PRレビューで指摘された以下の問題を全て修正する:
- 1件のImportant Issue（HIGH）
- 3件のMedium Issue
- 5件のテストカバレッジギャップ

---

## 修正項目一覧

### 1. [HIGH] deleteHistory()でエラー時もonSaved()が呼ばれる

**ファイル**: `MealInputViewModel.swift`, `MealInputView.swift`

**現状**:
```swift
// ViewModel
func deleteHistory(id: String) async {
    do {
        try await repository.deleteHistory(historyId: id)
    } catch {
        errorMessage = "削除に失敗しました"
    }
}

// View
Button("削除", role: .destructive) {
    Task {
        await viewModel.deleteHistory(id: meal.id)
        onSaved()  // エラーの有無に関わらず呼ばれる
    }
}
```

**修正**:
- `deleteHistory()`の戻り値を`Bool`に変更
- View側でエラー時は`onSaved()`を呼ばない
- エラーメッセージに詳細を追加

---

### 2. [MEDIUM] analyzeText()にタスクキャンセレーション処理がない

**ファイル**: `MealInputViewModel.swift`

**修正**:
- `Task.isCancelled`チェックを追加
- `CancellationError`を明示的にキャッチ

---

### 3. [MEDIUM] 画像選択時のtry?でエラーがサイレントに無視される

**ファイル**: `MealInputView.swift`

**現状**:
```swift
if let data = try? await newValue?.loadTransferable(type: Data.self),
   let image = UIImage(data: data) {
    viewModel.selectedImage = image
}
```

**修正**:
- `do-catch`に変更
- エラー時に`viewModel.errorMessage`を設定

---

### 4. [MEDIUM] テストカバレッジギャップの修正

**ファイル**: `MealInputViewModelTests.swift`

追加するテストケース:

| # | テストケース | Criticality |
|:--|:------------|:------------|
| 1 | 分析ステータスがfailedの場合エラーメッセージを表示すべき | 9/10 |
| 2 | 分析ステータスが不明な場合エラーメッセージを表示すべき | 8/10 |
| 3 | 1000文字ちょうどで分析が成功すべき | 7/10 |
| 4 | 分析結果取得失敗時にエラーメッセージを表示すべき | 7/10 |
| 5 | 非APIErrorでもエラーメッセージを表示すべき | 6/10 |
| 6 | deleteHistory成功時にtrueを返すべき | NEW |
| 7 | deleteHistory失敗時にfalseを返すべき | NEW |

---

## 実装計画

### Step 1: MealInputViewModel.swift の修正

1. `deleteHistory()`を`async -> Bool`に変更
2. `analyzeText()`にキャンセレーション処理を追加
3. `analyzeImage()`にも同様のキャンセレーション処理を追加（一貫性のため）

```swift
// deleteHistory
func deleteHistory(id: String) async -> Bool {
    do {
        try await repository.deleteHistory(historyId: id)
        return true
    } catch let error as APIError {
        errorMessage = "削除に失敗しました: \(error.localizedDescription)"
        return false
    } catch {
        errorMessage = "削除に失敗しました: \(error.localizedDescription)"
        return false
    }
}

// analyzeText - キャンセレーション対応
func analyzeText() async {
    // ... 既存のバリデーション ...

    do {
        let id = try await repository.analyzeText(...)
        guard !Task.isCancelled else { return }

        analysisId = id
        try await pollForCompletion(id: id)
        guard !Task.isCancelled else { return }

        analysisResult = try await repository.getAnalysisResult(id: id)
        showEditor = true
    } catch is CancellationError {
        return  // キャンセルは正常終了
    } catch let error as APIError {
        errorMessage = error.localizedDescription
    } catch {
        errorMessage = "テキスト分析に失敗しました: \(error.localizedDescription)"
    }

    isAnalyzing = false
}
```

### Step 2: MealInputView.swift の修正

1. 削除ボタンの処理を修正
2. 画像選択のエラーハンドリングを追加

```swift
// 削除ボタン
Button("削除", role: .destructive) {
    if let meal = deletingMeal {
        Task {
            let success = await viewModel.deleteHistory(id: meal.id)
            if success {
                onSaved()
            }
        }
    }
    deletingMeal = nil
}

// 画像選択
.onChange(of: selectedItem) { _, newValue in
    Task {
        guard let newValue else { return }
        do {
            guard let data = try await newValue.loadTransferable(type: Data.self) else {
                viewModel.errorMessage = "画像を読み込めませんでした"
                return
            }
            guard let image = UIImage(data: data) else {
                viewModel.errorMessage = "サポートされていない画像形式です"
                return
            }
            viewModel.selectedImage = image
        } catch {
            viewModel.errorMessage = "画像の読み込みに失敗しました"
        }
    }
}
```

### Step 3: MealInputViewModelTests.swift にテスト追加

```swift
// 1. 分析ステータスがfailedの場合
@Test func 分析ステータスがfailedの場合エラーメッセージを表示すべき() async { ... }

// 2. 分析ステータスが不明な場合
@Test func 分析ステータスが不明な場合エラーメッセージを表示すべき() async { ... }

// 3. 1000文字ちょうど
@Test func 1000文字ちょうどで分析が成功すべき() async { ... }

// 4. 分析結果取得失敗
@Test func 分析結果取得失敗時にエラーメッセージを表示すべき() async { ... }

// 5. 非APIError
@Test func 非APIErrorでもエラーメッセージを表示すべき() async { ... }

// 6. deleteHistory成功
@Test func deleteHistory成功時にtrueを返すべき() async { ... }

// 7. deleteHistory失敗
@Test func deleteHistory失敗時にfalseを返すべき() async { ... }
```

---

## 変更対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Features/Meals/MealInputViewModel.swift` | deleteHistory戻り値変更、キャンセレーション処理追加 |
| `ios/Uchikomi/Features/Meals/MealInputView.swift` | 削除処理修正、画像選択エラーハンドリング追加 |
| `ios/UchikomiTests/Features/Meals/MealInputViewModelTests.swift` | 7件のテスト追加 |

---

## 検証方法

1. `task ios:lint` でリントエラーがないことを確認
2. `xcodebuild build` でビルドが成功することを確認
3. ユーザーによる手動検証:
   - 削除失敗時にエラーメッセージが表示され、一覧が更新されないことを確認
   - 画像選択失敗時にエラーメッセージが表示されることを確認
   - 分析中に画面を閉じてもエラーが発生しないことを確認
