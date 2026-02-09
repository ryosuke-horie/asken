# ADR-001: iOSユニットテスト技術選定

## コンテキスト

ウチコミiOSアプリ（Swift/SwiftUI）のユニットテスト環境を整備する必要がある。以下の観点で技術選定を行った：

- テストフレームワークの選定（XCTest vs Swift Testing）
- Mockライブラリの選定
- スナップショットテストの導入
- テスト実行タイミングとCI/CD戦略

## 決定

### テストフレームワーク

| 分野 | 選定 |
|:---|:---|
| ユニットテスト | Swift Testing |

### ライブラリ

| 用途 | ライブラリ |
|:---|:---|
| Mock生成 | Mockolo（Uber製） |
| スナップショット | swift-snapshot-testing（Point-Free製） |

### テスト実行タイミング

| テスト種類 | ローカル | CI/PR |
|:---|:---|:---|
| ユニットテスト | 随時 | 毎回実行 |
| スナップショットテスト | 開発時 | 毎回実行 |

### テスト方針

| 項目 | 方針 |
|:---|:---|
| テストスタイル | 古典派（Classicist） |
| Mock対象 | 外部依存（API、Keychain）のみ |
| 命名規則 | 日本語「〜すべき」 |
| 構造 | Arrange-Act-Assert |
| ファイル配置 | `UchikomiTests/` 配下 |

### カバレッジ目標

| 対象 | 目標 |
|:---|:---|
| ViewModel | 80%以上 |
| Repository | 70%以上 |
| Model（ロジックあり） | 80%以上 |
| 全体 | 60%以上 |

### テストタイミング

| 場面 | アプローチ |
|:---|:---|
| ViewModel / ビジネスロジック | TDD推奨 |
| Repository / API連携 | TDD推奨 |
| バグ修正 | テスト先行必須 |
| 新規UI画面 | テスト後追いOK |

## 理由

### Swift Testing を選定した理由

- 新規プロジェクトであり、最新フレームワークを採用するメリットが大きい
- `@Test`マクロによるモダンな記法
- デフォルトで並列実行され、テスト実行が高速
- 日本語テスト名（「〜すべき」）で可読性が高い

### Mockolo を選定した理由

- Uber製で信頼性が高く、大規模プロジェクトでの実績がある
- 高速なMock生成（CLIベース）
- アクティブにメンテナンスされている（2024-2025年も更新あり）
- 生成されるコードがシンプルで、テストフレームワーク非依存

検討した他の選択肢:
- Cuckoo: コミュニティは大きいが、Swift Testing対応が不明確
- Mockable: Swift Testing明示対応だが、v0.4.1でまだ若い
- 手動Mock: ボイラープレートが多く、規模拡大時に負担

### スナップショットテストを導入する理由

- UIの意図しない変更を検知できる
- swift-snapshot-testingはSwift Testing対応済み
- CI/PRでは`record: .never`で実行し、意図しないスナップショット生成を防止

## 結果

### 現状のテスト方針（一時停止中）

macOS および Xcode のバージョンアップにより、iOS テストが頻繁に壊れる問題が発生しています。
この問題が解消するまで、iOS テストを一時的に無効化しています。

詳細は `.claude/rules/ios-testing-policy.md` を参照してください。

### 導入が必要なもの

1. Mockolo: Swift Package Managerで追加
2. swift-snapshot-testing: Swift Package Managerで追加

### 既存テストへの影響

- 既存の`UchikomiTests/`内のXCTestコードはSwift Testingと共存可能
- 段階的に移行可能

### CI/CD設定

- ユニットテスト: GitHub Actionsで毎PR実行
- スナップショットテスト: GitHub Actionsで毎PR実行（`record: .never`）

### ドキュメント

- `.claude/rules/`にiOSテストガイドラインを追加
