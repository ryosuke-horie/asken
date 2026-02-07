# Code Reviewer エージェント メモリ

## プロジェクト構造

- バックエンド: Go (backend/)
- iOS: SwiftUI (ios/)
- Gemini API連携: backend/pkg/gemini/
- タスク管理: Taskfile

## internal パッケージのテスト品質 (2026-02-07 評価)

### 品質ランク
- util/timezone_test.go: A (テーブル駆動、正常/異常網羅)
- middleware/dev_auth_test.go: A- (認証パスを網羅)
- repository/analysis_repository_firestore_test.go: B (統合テスト、time.Sleep脆弱性あり)
- worker/analysis_worker_test.go: B- (SaveResult失敗等のエラーパス不足)
- service/food_service_test.go: C (Cloud Storageパス分岐が完全未テスト)
- middleware/auth_test.go: D (Authenticate未テスト、ヘルパーのみ)
- repository/storage_repository_test.go: F (モック自体のテスト、実装テストなし)

### 主要な問題パターン (internal)
1. storage_repository_test.go はモック動作をテストしており実装テストではない
2. AuthMiddleware.Authenticate が完全に未テスト（セキュリティリスク）
3. FoodService の Cloud Storage パス分岐が未テスト
4. Firestore統合テストで time.Sleep(100ms) による待機が5箇所（不安定）
5. Worker の複数エラーパス未テスト（SaveResult失敗、不明InputType）
6. CreateRequestFromMylist の完全な未テスト

## pkg/gemini パッケージのテスト品質 (2026-02-07 評価)

### 品質ランク
- http_client_test.go: A (モックサーバー活用、エラーケース網羅)
- classifier_test.go: B- (detectMimeType は良好、API依存テスト多い)
- client_test.go: C (重複テスト、ユニットテスト不足)
- nutrition_test.go: D (ユニットテスト1件のみ)
- text_parser_test.go: D (ユニットテスト2件のみ)
- equipment_normalizer.go: テストなし
- training_menu.go: テストなし

### 主要な問題パターン
1. 統合テストとユニットテストの混在（CI環境でスキップ多発）
2. http_client_test.go のモックサーバーパターンが他ファイルに未展開
3. 永久スキップテストがデッドコードとして残存
4. removeCodeBlock テストが2ファイルに重複
5. インターフェース未使用によるテスタビリティ低下（HTTPClient直接依存）
