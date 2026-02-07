# バックエンド テストコード評価レポート

日付: 2026-02-07

## 1. 概要

バックエンド（Go）の全テストコードについて、テスト設計・品質・カバレッジの観点から評価を実施した。

### 数値サマリー

| 指標 | 値 |
|:---|:---|
| プロダクションコードファイル数 | 26 |
| テストファイル数 | 20 |
| テストがないファイル数 | 7 |
| プロダクションコード行数 | 約3,840行 |
| テストコード行数 | 約4,760行 |
| テスト/プロダクション比率 | 1.24 |
| テストカバレッジ（計測可能分） | 52.3%（環境制約あり） |

環境制約: `google.golang.org/api` のダウンロード不可により、11パッケージ中8パッケージがビルドできず完全な計測は不可。計測可能だった `internal/util`（94.7%）、`pkg/gemini`（51.3%）が対象。

---

## 2. レイヤー別評価

### 2.1 Handler層（8ファイル中7ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `analyze_handler_test.go` | 607 | B- | HandleUploadImage正常系テスト欠落、JPEG生成コード重複 |
| `daily_meals_handler_test.go` | 192 | B | タイムゾーン分岐テスト欠落 |
| `health_handler_test.go` | 44 | A | 問題なし |
| `history_delete_handler_test.go` | 138 | A | 問題なし |
| `history_handler_test.go` | 425 | B- | テスト名と検証内容の矛盾、NotFound判定の不一致 |
| `image_handler_test.go` | 142 | A | パストラバーサル検証あり。模範的 |
| `skip_meal_handler_test.go` | 198 | B+ | 不正JSONデコードのテスト欠落 |
| `status_handler_test.go` | 236 | B | default分岐未テスト、GetResultエラーパス未テスト |

良い点:
- 外部依存のみをモックする古典派スタイルが徹底されている
- 全認証ハンドラーで未認証テストが存在
- `image_handler_test.go` のパストラバーサル検証が充実

Handler層の横断的問題:
- モック定義が `analyze_handler_test.go` に集約されており、このファイルを削除すると全テストが壊れる
- 一部ハンドラー（`daily_meals`, `skip_meal`, `status`, `analyze`）にHTTPメソッドチェックがない
- `history_handler.go` のNotFound判定が `strings.Contains(err.Error(), "見つかりません")` で、`history_delete_handler.go` の `errors.Is(err, repository.ErrNotFound)` と不一致

### 2.2 Service層（2ファイル中1ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `food_service_test.go` | 234 | C | Cloud Storage経由パス未テスト、detectMimeType未テスト |
| `firebase_auth_service.go` | - | テストなし | 認証サービスのテスト不在 |

問題点:
- `AnalyzeFoodImage` の `uploads/` プレフィックス判定によるCloud Storage経由パス（Download -> ClassifyFoodsFromData）が一切テストされていない
- `detectMimeTypeFromPath` の6分岐（jpg/jpeg/png/gif/webp/デフォルト）が未検証
- `firebase_auth_service.go` はFirebase依存のため直接テストが困難だが、少なくともインターフェース経由の検証を検討すべき

### 2.3 Repository層（3ファイル中2ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `analysis_repository_firestore_test.go` | 529 | B | time.Sleep x5箇所、CreateRequestFromMylist未テスト |
| `storage_repository_test.go` | 144 | F | モック自体のテストのみ、実装テストがゼロ |

`storage_repository_test.go` の致命的問題:
- `TestMockStorageRepository_Upload/GetSignedURL/Delete` は `testutil.MockStorageRepository` のモック動作を検証しているだけ
- `cloudStorageRepository` の `Upload`, `Download`, `GetSignedURL`, `Delete` の実装は一切テストされていない
- UUIDファイル名生成、io.Copyエラー、writer.Closeエラー、ErrObjectNotExist変換など全パスが未検証

`analysis_repository_firestore_test.go` の注意点:
- Firestoreエミュレータを使った統合テストとして設計されている（適切）
- 各テストでユニークuserIDを生成し、`t.Cleanup` でデータクリーンアップを実施（良い）
- `time.Sleep(100ms)` によるFirestore書き込み伝播待機が5箇所あり、CI環境で不安定化（flaky test）のリスク
- `CreateRequestFromMylist` メソッドが完全に未テスト

### 2.4 Middleware層（2ファイル中2ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `dev_auth_test.go` | 124 | A- | `os.Setenv` -> `t.Setenv` 推奨 |
| `auth_test.go` | 38 | D | Authenticate完全未テスト |

`auth_test.go` の致命的問題:
- テスト対象が `SetFirebaseUIDToContext` と `GetFirebaseUIDFromContext` の補助関数のみ
- `AuthMiddleware.Authenticate` メソッドが完全に未テスト
- Authorizationヘッダーなし / Bearer形式以外 / 無効トークン / 有効トークンの全パスが未検証
- セキュリティ上の最重要コンポーネントにテストが存在しない

### 2.5 Worker層（1ファイル中1ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `analysis_worker_test.go` | 444 | B- | default分岐、SaveResult失敗パス未テスト |

良い点:
- コールバックベースのモックで呼び出し回数/引数を検証
- Image/Text の成功/失敗パターンを網羅
- コンテキストキャンセルによる停止テストあり

問題点:
- 不明な InputType の default分岐が未テスト
- SaveResult失敗時の StatusFailed 更新パスが未検証
- `time.Sleep(250ms)` によるタイミング依存テストが存在

### 2.6 pkg/gemini パッケージ（7ファイル中5ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `http_client_test.go` | 448 | A | httptest.NewServer活用、エラーケース網羅 |
| `classifier_test.go` | 213 | B- | ClassifyFoodsFromData未テスト、API依存率高 |
| `client_test.go` | 109 | C | 重複テスト、ユニットテスト不足 |
| `nutrition_test.go` | 74 | D | ユニットテスト1件のみ |
| `text_parser_test.go` | 76 | D | ユニットテスト最低限のみ |
| `equipment_normalizer.go` | - | テストなし | 優先度: 中 |
| `training_menu.go` | - | テストなし | 優先度: 高 |

構造的問題:
- 全クライアント構造体が `*HTTPClient` を直接保持しており、インターフェースを使っていない
- そのためモックベースのユニットテストが書けず、API依存の統合テストに頼っている
- CI環境で実行可能なユニットテスト数: 24テスト関数中9件（37.5%）のみ
- 永久スキップテストが2件残存（デッドコード）
- `removeCodeBlock` テストが `client_test.go` と `http_client_test.go` で完全重複

### 2.7 Util層（1ファイル中1ファイルにテストあり）

| テストファイル | 行数 | 総合評価 | 主要な問題 |
|:---|---:|:---:|:---|
| `timezone_test.go` | 132 | A | DST境界テストなし（軽微） |

テーブル駆動テストを適切に使用し、正常系/異常系を網羅。模範的なテストファイル。

---

## 3. クリティカルな問題（対応必須）

### 3.1 `AuthMiddleware.Authenticate` が完全に未テスト

- 場所: `backend/internal/middleware/auth_test.go`
- 影響: 認証ミドルウェアの主要メソッドのテストが不在
- 原因: `FirebaseAuthService` が具象型で直接依存しており、モック注入が困難
- 対策: `FirebaseAuthService` にインターフェースを導入し、モック経由でテスト

### 3.2 `storage_repository_test.go` が実装をテストしていない

- 場所: `backend/internal/repository/storage_repository_test.go`
- 影響: `cloudStorageRepository` の全メソッド（Upload/Download/GetSignedURL/Delete）が未検証
- 原因: テストがモック自体の動作検証のみ
- 対策: Cloud Storageのインターフェースを使ったユニットテスト、またはエミュレータを使った統合テストの追加

### 3.3 NotFound判定方式の不一致

- 場所: `backend/internal/handler/history_handler.go:134,231`
- 影響: `strings.Contains(err.Error(), "見つかりません")` で判定しており脆弱
- 比較: `history_delete_handler.go` は `errors.Is(err, repository.ErrNotFound)` を使用
- 対策: `errors.Is` による判定に統一

### 3.4 テスト名と検証内容の矛盾

- 場所: `backend/internal/handler/history_handler_test.go:215`
- 影響: `TestHistoryHandler_HandleDetail_NotFound` が実際にはInternalServerErrorパスをテスト
- 原因: モックが `assert.AnError` を返すが、そのメッセージに「見つかりません」が含まれない
- 対策: テスト名の修正、または正しいNotFoundテストの追加

---

## 4. 重要な問題（対応推奨）

### 4.1 pkg/gemini のテスタビリティ

- 全クライアント構造体が `*HTTPClient` に直接依存しており、インターフェース未導入
- CI環境で実行可能なユニットテストが全体の37.5%のみ
- 対策: `HTTPClient` にインターフェースを導入し、DI可能にする

### 4.2 テストのないファイル（優先度順）

| 優先度 | ファイル | 行数 | 理由 |
|:---|:---|---:|:---|
| 高 | `training_menu.go` | 128 | バリデーション3パターン + オプショナルパラメータ構築ロジック |
| 高 | `firebase_auth_service.go` | 55 | 認証サービス |
| 中 | `equipment_normalizer.go` | 74 | 正規化ロジック |
| 中 | `analysis_models.go` | 121 | データモデル |
| 低 | `storage/client.go` | 30 | クライアント初期化 |
| 低 | `database/firestore.go` | 36 | クライアント初期化 |
| 低 | `cmd/server/main.go` | 313 | エントリーポイント |

### 4.3 テスト内の `time.Sleep` 使用

- `analysis_repository_firestore_test.go`: 5箇所
- `analysis_worker_test.go`: 1箇所
- CI環境での不安定なテスト(flaky test)の原因になり得る
- 対策: ポーリングまたはイベント駆動の待機に置き換え

### 4.4 テストコードの重複・デッドコード

- `removeCodeBlock` テストが `client_test.go` と `http_client_test.go` で重複
- `client_test.go` の2つの統合テストが実質同一内容
- 永久スキップテスト2件がデッドコードとして残存
- JPEG画像データ生成が `analyze_handler_test.go` 内で4箇所重複

### 4.5 food_service_test.go のカバレッジ不足

- `uploads/` プレフィックスによるCloud Storage経由パスが完全に未テスト
- `detectMimeTypeFromPath` の6分岐が未検証

---

## 5. 良い点

1. 古典派テストスタイルの徹底 - 外部依存のみモック、内部実装のモックは回避
2. `image_handler_test.go` のセキュリティテスト - パストラバーサル攻撃を2パターンで検証
3. `http_client_test.go` のモックサーバーパターン - httptest.NewServerを活用し、エラーケースを網羅
4. `analysis_repository_firestore_test.go` の統合テスト設計 - ユニークuserID生成とCleanupによる独立性確保
5. 全認証ハンドラーでの未認証テスト - Firebase UID欠落時の401レスポンスを全ハンドラーで検証
6. `timezone_test.go` のテーブル駆動テスト - 模範的な構造
7. testifyの使い分け - `require`(テスト中断)と`assert`(継続)の使い分けが概ね適切

---

## 6. 改善ロードマップ

### Phase 1: セキュリティ関連（最優先）

- [ ] `AuthMiddleware.Authenticate` のテスト追加（インターフェース導入含む）
- [ ] `history_handler.go` のNotFound判定を `errors.Is` に統一

### Phase 2: カバレッジ拡大

- [ ] `storage_repository_test.go` の実装テスト追加
- [ ] `training_menu.go` のテスト追加
- [ ] `food_service_test.go` のCloud Storageパス・detectMimeTypeテスト追加
- [ ] `equipment_normalizer.go` のテスト追加

### Phase 3: テスタビリティ改善

- [ ] `pkg/gemini` の `HTTPClient` にインターフェース導入
- [ ] 各Geminiクライアントのモックベースユニットテスト追加
- [ ] 統合テストとユニットテストのファイルレベル分離（ビルドタグ）

### Phase 4: テスト品質向上

- [ ] `time.Sleep` をポーリング/イベント駆動に置換
- [ ] テストコード重複の解消（removeCodeBlock、JPEG生成）
- [ ] デッドコード（永久スキップテスト）の削除
- [ ] モック定義を `mock_test.go` に分離
- [ ] HandleUploadImage正常系テスト追加
- [ ] HTTPメソッドチェックの一貫性確保

---

## 7. 総合評価

| レイヤー | 評価 | 備考 |
|:---|:---:|:---|
| Handler | B | 大部分は良好。NotFound判定とモック配置に改善余地 |
| Service | C | Cloud Storageパスと認証サービスのテスト不在 |
| Repository | C | Firestore統合テストは良好。Storageテストが致命的 |
| Middleware | C | dev_authは優秀。本番authが未テスト |
| Worker | B- | 基本パスはカバー。エラー連鎖パスに不足 |
| pkg/gemini | C+ | http_clientは優秀。他はAPI依存でCI実効性低 |
| Util | A | 模範的 |
| 全体 | C+ | 基盤はあるが、セキュリティ・テスタビリティに重要な課題あり |

テストコード自体の質は概ね良好だが、テストが存在しない重要コンポーネント（認証ミドルウェア、Storageリポジトリ実装）と、CI環境でのテスト実効性（pkg/geminiの37.5%のみ）に大きな改善余地がある。Phase 1の対応を優先的に進めることを推奨する。
