# プラン: 画像パスURL構築時のバリデーション追加

## Linear Issue
- Issue: EDG-620
- URL: https://linear.app/ryosuke-horie/issue/EDG-620

## 概要

MealInputView.swiftの`ExistingMealCard.imageURL`で、サーバーから返された`imagePath`からURLを構築する際に`NSString.lastPathComponent`のみで安全性を担保している。Defense in Depth（多層防御）の観点から、ファイル名のホワイトリストバリデーションを追加する。

## 現状分析

### 現在のコード（MealInputView.swift 319-325行）

```swift
private var imageURL: URL? {
    guard let imagePath = meal.imagePath, !imagePath.isEmpty else { return nil }
    let filename = (imagePath as NSString).lastPathComponent
    let baseURL = AppEnvironment.current.baseURL
    return baseURL.appendingPathComponent("api/images/\(filename)")
}
```

### リスク評価

- サーバー側は`uploads/{UUID}.{ext}`形式で生成しており、通常運用では安全
- サーバー側（image_handler.go）でもパストラバーサル検証あり
- ただし、サーバー侵害時にFirestoreの`imagePath`が改ざんされた場合、不正なURLが構築される可能性がある
- `lastPathComponent`は基本的なパス抽出を行うが、特殊文字を含むファイル名は防げない
- 実質的なリスクは低いが、多層防御として対応する価値がある

## 実装計画

### 1. `ImageFilenameValidator`を作成

`QuantityParser`と同じパターン（enumの静的メソッド）で実装する。

- ファイル: `ios/Uchikomi/Features/Meals/Models/ImageFilenameValidator.swift`
- バリデーションルール:
  - 空文字でないこと
  - `..`を含まないこと（パストラバーサル防止）
  - 許可文字のみ: `[a-zA-Z0-9._-]`（英数字、ドット、ハイフン、アンダースコア）
  - 許可拡張子のみ: jpg, jpeg, png, heic（大文字小文字不問）

### 2. テストを作成

- ファイル: `ios/UchikomiTests/Features/Meals/ImageFilenameValidatorTests.swift`
- `QuantityParserTests`と同じスタイル（Swift Testing、日本語テスト名、`#expect`）
- テストケース:
  - 正常系: UUID形式ファイル名、各拡張子、大文字拡張子、アンダースコア含むファイル名
  - 異常系: 空文字、パストラバーサル（`../../etc/passwd`）、スラッシュ含む、スペース含む、非許可拡張子（gif, svg）、拡張子なし、特殊文字（`<script>`）、`..`のみ

### 3. MealInputViewを修正

- `ExistingMealCard.imageURL`にバリデーションguardを1行追加
- バリデーション失敗時は`nil`を返す（画像非表示 - 既存UIで対応済み）

## 変更対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `ios/Uchikomi/Features/Meals/Models/ImageFilenameValidator.swift` | 新規作成 |
| `ios/UchikomiTests/Features/Meals/ImageFilenameValidatorTests.swift` | 新規作成 |
| `ios/Uchikomi/Features/Meals/MealInputView.swift` | 1行追加（guard文） |

## 技術的な考慮事項

- `QuantityParser`と同じenumパターンを採用（テスト可能性の確保）
- Swift Regex（`#/.../#`）を使用
- `ExistingMealCard`はprivate structのため、バリデーションロジックを外部に抽出しないとテスト不可
- XcodeGenの`sources`設定によりディレクトリ配下のファイルは自動検出される

## テスト計画

- `ImageFilenameValidatorTests`で正常系・異常系を網羅的にテスト
- `task ios:test` でテスト実行
- `task ios:lint` でリント確認
- `task ios:format` でフォーマット確認

## 検証方法

1. テストがすべてPassすること
2. リントエラーがないこと
3. 既存の画像表示が正常に動作すること（UUID形式のファイル名はバリデーションを通過する）
