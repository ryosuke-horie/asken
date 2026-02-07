# セキュリティ診断レポート

- 対象: ウチコミ プロジェクト全体
- 診断日: 2026-02-07
- 診断者: security-reviewer エージェント (Claude Code)

---

## 総合サマリー

| 領域 | CRITICAL | HIGH | MEDIUM | LOW | INFO | リスクレベル |
|:---|:---|:---|:---|:---|:---|:---|
| バックエンド (Go) | 1 | 4 | 5 | 4 | 3 | HIGH |
| iOS (Swift) | 0 | 2 | 5 | 5 | 3 | MEDIUM |
| 設定/シークレット管理 | 0 | 2 | 4 | 4 | 3 | MEDIUM |
| CI/CD/インフラ | 1 | 5 | 6 | 4 | 8 | HIGH |
| 合計 | 2 | 13 | 20 | 17 | 17 | HIGH |

---

## 1. バックエンド (Go) セキュリティ診断

### CRITICAL (1件)

#### C-BE-1: APIエンドポイントにレート制限が存在しない

- ファイル: `backend/cmd/server/main.go` 63-131行
- 問題: 全APIエンドポイント(`/api/analyze`, `/api/upload-image`, `/api/history`, `/api/meals/daily`, `/api/meals/skip`)にレート制限が未実装。特にGemini API呼び出しを伴うエンドポイントは外部有料APIのコスト増大やサービス停止を引き起こすリスクがある
- 推奨: `golang.org/x/time/rate` でユーザーごと(Firebase UID単位)とIP単位の2段階レート制限を導入。Gemini API関連は1分あたり10リクエスト程度に制限

### HIGH (4件)

#### H-BE-1: JSONリクエストボディにサイズ制限がない

- ファイル: `backend/internal/handler/analyze_handler.go` 68行, `history_handler.go` 207行, `skip_meal_handler.go` 42行
- 問題: `json.NewDecoder(r.Body).Decode()` に `http.MaxBytesReader` による制限なし。数GBのJSONボディでメモリ枯渇DoS攻撃が可能
- 推奨: `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` (1MB) を適用

#### H-BE-2: セキュリティヘッダーが未設定

- ファイル: `backend/cmd/server/main.go` 272-292行
- 問題: `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`, `Content-Security-Policy` が未設定
- 推奨: セキュリティヘッダー設定用ミドルウェアを作成し全レスポンスに適用

#### H-BE-3: 画像配信エンドポイントに認証が不要

- ファイル: `backend/cmd/server/main.go` 67-68行, `backend/internal/handler/image_handler.go` 30-91行
- 問題: `/api/images/{filename}` はUUIDファイル名の推測困難さのみに依存(Security through Obscurity)。UUIDが漏洩すると認証なしで食事画像にアクセス可能
- 推奨: 署名付きURLの有効期限短縮、可能であれば認証追加

#### H-BE-4: Gemini APIレスポンスの読み取りにサイズ制限がない

- ファイル: `backend/pkg/gemini/http_client.go` 174行, `backend/internal/repository/storage_repository.go` 89行
- 問題: `io.ReadAll` で制限なしに読み取り。異常に大きなレスポンスでメモリ枯渇の可能性
- 推奨: `io.LimitReader` で上限設定 (Gemini: 10MB, 画像: 15MB)

### MEDIUM (5件)

#### M-BE-1: 開発用認証バイパスのモックトークンがハードコード

- ファイル: `backend/internal/middleware/dev_auth.go` 12-14行
- 問題: `DevMockToken = "dev-mock-token"` がエクスポートされた定数。本番で `APP_ENV=development` が設定された場合にバイパス可能
- 推奨: ビルドタグ (`//go:build dev`) で本番ビルドに含めない。またはローカル環境以外で動作しないようチェック追加

#### M-BE-2: meal_dateパラメータのフォーマットバリデーション不足

- ファイル: `backend/internal/handler/analyze_handler.go` 65行, `skip_meal_handler.go` 25行
- 問題: 空チェックのみでフォーマット検証なし。不正な日付がリポジトリ層まで到達し500エラーを返す
- 推奨: ハンドラー層で `time.Parse("2006-01-02", mealDate)` を実施

#### M-BE-3: ユーザー入力テキストがログに直接出力

- ファイル: `backend/internal/handler/analyze_handler.go` 103行, `backend/internal/worker/analysis_worker.go` 112行
- 問題: 食事内容テキストがサニタイズなしでログ出力。個人情報の永続化とログインジェクションリスク
- 推奨: トランケートやフィルタリング。個人情報を含むデータはログレベルを下げるか出力しない

#### M-BE-4: CORSのVaryヘッダー未設定

- ファイル: `backend/cmd/server/main.go` 274-279行
- 問題: 動的CORSオリジン設定に `Vary: Origin` がない。CDNキャッシュ汚染の可能性
- 推奨: `w.Header().Set("Vary", "Origin")` を追加

#### M-BE-5: 履歴更新エンドポイントのフードアイテム数に上限がない

- ファイル: `backend/internal/handler/history_handler.go` 205-226行
- 問題: `foods` 配列の要素数チェックなし。Firestoreドキュメントサイズ上限や計算ループに影響
- 推奨: 要素数上限 (例: 50件) を設定

### LOW (4件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| L-BE-1 | `cmd/server/main.go` 294-301行 | CORSにハードコードされたlocalhostオリジンが本番でも有効 |
| L-BE-2 | `cmd/server/main.go` 235-241行 | HTTPサーバーのRead/Write/Idleタイムアウトが150秒と長い。`ReadHeaderTimeout` 未設定 |
| L-BE-3 | `internal/handler/history_handler.go` 134行 | エラーメッセージの文字列比較で404/500を判定。`errors.Is` に統一すべき |
| L-BE-4 | `internal/handler/analyze_handler.go` 296-318行 | HEICファイルは拡張子のみで判定、マジックバイト検証なし |

### INFO (良好な設計)

- Dockerfile: distroless + nonroot + マルチステージビルド
- Gemini APIキー: 環境変数から取得、HTTPヘッダーで送信、Secret Manager管理
- 認証・認可: Firebase Auth検証が適切。Firestoreコレクション設計でデータ分離を構造的に保証

---

## 2. iOS (Swift) セキュリティ診断

### HIGH (2件)

#### H-iOS-1: APIエンドポイントのパスパラメータ未サニタイズ

- ファイル: `ios/Uchikomi/Core/Network/APIEndpoint.swift` 49, 56, 66, 74, 82行
- 問題: `id` パラメータを文字列補間でURLパスに直接埋め込み。パストラバーサル文字を含む `id` で意図しないURLへのリクエスト可能性
- 推奨: `id` に英数字・ハイフンのみ許可するバリデーション、または `addingPercentEncoding` でURLエンコード

#### H-iOS-2: サーバー応答の画像パスからURL構築時のバリデーション不足

- ファイル: `ios/Uchikomi/Features/Meals/MealInputView.swift` 250-255行
- 問題: `NSString.lastPathComponent` のみで安全性を担保。サーバー侵害時に任意URLパス構築の可能性
- 推奨: ファイル名に英数字・ハイフン・ドット・アンダースコアのみ許可するバリデーション追加

### MEDIUM (5件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| M-iOS-1 | `Core/Network/APIClient.swift` 38-42行 | SSL Pinningが未実装。MitM攻撃リスク |
| M-iOS-2 | `Core/Network/APIClient.swift` 129-133行 | DEBUGガード内だがAPIレスポンスボディ全文をログ出力。テスター配布時の情報漏洩リスク |
| M-iOS-3 | `project.yml`, `App/AppDelegate.swift` | Firebase App Checkが未導入。正規アプリ以外からのAPI呼び出し防止不可 |
| M-iOS-4 | `App/AppEnvironment.swift` 22行 | シミュレータ環境で `http://localhost:8080` への平文HTTP通信 |
| M-iOS-5 | `Core/Network/APIClient.swift` 24-26行 | テスト用 `setShared()` が `#if DEBUG` なしでプロダクションに含まれる |

### LOW (5件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| L-iOS-1 | `App/AppEnvironment.swift` 25行, `Resources/Info.plist` 37行 | GCPプロジェクト番号がソースコードにハードコード |
| L-iOS-2 | `Core/Network/APIEndpoint.swift` 23行 | `fatalError` がプロダクションコードに存在 |
| L-iOS-3 | `Core/Network/APIError.swift` 22-26行 | サーバーエラーメッセージをそのままユーザーに表示 |
| L-iOS-4 | `Features/Meals/MealInputViewModel.swift` 77-78行 | 画像アップロード前のサイズ・ピクセル数検証なし |
| L-iOS-5 | `Features/Meals/MealInputView.swift` 87-109行 | クライアント側レート制限なし |

### 良好な設計

- 認証トークン保存はFirebase SDK (Keychain) に委任
- モック認証サービスは `#if DEBUG` で適切にガード
- Apple Sign-Inのnonce生成は `SecRandomCopyBytes` + SHA256で暗号学的に安全
- 全APIエンドポイントで `requiresAuth: true` 設定
- テキスト入力は1000文字制限あり
- ATSが有効（デフォルト）

---

## 3. 設定/シークレット管理 セキュリティ診断

### HIGH (2件)

#### H-CFG-1: 開発用モックトークンのハードコード

- ファイル: `backend/internal/middleware/dev_auth.go` 13行
- 問題: (M-BE-1と同一) `dev-mock-token` がソースコードにコミット。本番環境への設定ミス時にバイパス可能
- 推奨: ビルドタグで本番ビルドに含めない

#### H-CFG-2: Cloud Storage CORSが全オリジン許可

- ファイル: `infrastructure/environments/dev/variables.tf` 61行
- 問題: `storage_cors_origins = ["*"]`。任意ドメインからバケットへのリクエスト許可
- 推奨: iOSネイティブアプリはCORS対象外のため、CORSを空リストにするか必要オリジンのみに制限

### MEDIUM (4件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| M-CFG-1 | `backend/` | `backend/.env.example` が存在しない。新規開発者が必要な環境変数を把握できない |
| M-CFG-2 | `infrastructure/environments/dev/backend.tf` | tfstateバケットのセキュリティ設定(CMEK, IAM制限, バージョニング)が未監査 |
| M-CFG-3 | `infrastructure/modules/cloud-run/variables.tf` 78行 | Cloud Run `request_timeout = 300秒` が長い |
| M-CFG-4 | `infrastructure/environments/dev/variables.tf` 33行 | Firestore削除保護が無効 (本番移行時の注意) |

### LOW (4件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| L-CFG-1 | `.mise.toml` 27-33行 | 1PasswordアイテムIDが公開 (直接的リスクは低い) |
| L-CFG-2 | `.gitignore` | `backend/.env` の明示的追加推奨 |
| L-CFG-3 | `infrastructure/environments/dev/terraform.tfvars.example` | `gemini_api_key` プレースホルダー未記載 |
| L-CFG-4 | `infrastructure/environments/dev/providers.tf` | Terraformプロバイダーのバージョン範囲が広い |

### 良好な設計

- `.gitignore`: `.env`, `.tfvars`, `sa-key.json`, `service-account*.json`, `GoogleService-Info.plist` が適切に除外
- WIF採用でGitHub Actions -> GCPのキーレス認証を実現
- Gemini APIキーはSecret Manager経由でCloud Runに注入
- git履歴にシークレットコミット痕跡なし

---

## 4. CI/CD・インフラ セキュリティ診断

### CRITICAL (1件)

#### C-CICD-1: Cloud Runサービスアカウントにランタイム+CI/CD権限が集約

- ファイル: `infrastructure/modules/cloud-run/main.tf` 46-68行
- 問題: ランタイム用SAとCI/CD用SAが同一。ランタイム侵害時にArtifact Registryへの悪意あるイメージプッシュ、Cloud Run再デプロイが可能。サプライチェーン攻撃へのエスカレーションパス
- 推奨: CI/CD専用SA (`uchikomi-api-dev-deploy-sa`) を新規作成し権限を分離

### HIGH (5件)

#### H-CICD-1: サードパーティActionsがSHAピンニングされていない

- ファイル: `.github/workflows/ci.yml` 20, 24, 43, 46, 69行, `.github/workflows/deploy.yml` 26, 30, 78, 82, 89行
- 問題: 全Actions がタグ参照。特に `deploy.yml` の `id-token: write` 権限下で侵害Actionsがトークン窃取可能
- 推奨: 全サードパーティActionsをコミットSHAにピンニング

#### H-CICD-2: WIFにブランチ制限がない

- ファイル: `infrastructure/modules/wif/main.tf` 38行
- 問題: `attribute_condition` がリポジトリのみで制限。任意ブランチからGCP認証可能。C-CICD-1と組み合わせで深刻な攻撃パス
- 推奨: `assertion.ref == 'refs/heads/main'` のブランチ制限を追加

#### H-CICD-3: IAM権限がプロジェクトレベルで付与

- ファイル: `infrastructure/modules/cloud-run/main.tf` 17-44行
- 問題: `roles/datastore.user`, `roles/storage.objectUser`, `roles/secretmanager.secretAccessor` 等がプロジェクト内の全リソースに適用
- 推奨: `google_secret_manager_secret_iam_member` 等でリソースレベルに限定

#### H-CICD-4: CIでダウンロードしたバイナリのチェックサム検証なし

- ファイル: `.github/workflows/ci.yml` 72-85行
- 問題: SwiftLint/SwiftFormatバイナリをGitHub Releasesからダウンロード後、SHA256検証なし
- 推奨: `echo "${EXPECTED_SHA256}  file.zip" | sha256sum -c -` を追加

#### H-CICD-5: deploy.ymlにトップレベルpermissions未定義

- ファイル: `.github/workflows/deploy.yml` 1-16行
- 問題: `changes` ジョブにpermissions定義なし。デフォルト権限が適用される
- 推奨: トップレベルに `permissions: contents: read` を追加

### MEDIUM (6件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| M-CICD-1 | `infrastructure/environments/dev/variables.tf` 58-62行 | Cloud Storage CORSが全オリジン許可、DELETEメソッドも許可 |
| M-CICD-2 | `infrastructure/modules/cloud-run/main.tf` 78行 | Cloud Run ingress=INGRESS_TRAFFIC_ALL。WAF層なし |
| M-CICD-3 | `infrastructure/environments/dev/variables.tf` 30-34行 | Firestore削除保護が無効 |
| M-CICD-4 | `.github/workflows/ci.yml`, `deploy.yml` 全checkout | `persist-credentials: false` 未設定 |
| M-CICD-5 | `infrastructure/environments/dev/variables.tf` 128-148行 | Terraform stateに機密データが平文格納 |
| M-CICD-6 | `infrastructure/environments/dev/variables.tf` 74-78行 | `google_oauth_client_id` に `sensitive = true` なし |

### LOW (4件)

| ID | ファイル | 問題 |
|:---|:---|:---|
| L-CICD-1 | `.github/dependabot.yml` 53-58行 | Docker ecosystemのdirectoryが `"/"` (正しくは `"/backend"`) |
| L-CICD-2 | `infrastructure/modules/cloud-run/variables.tf` 76-79行 | `request_timeout = 300秒` が長い |
| L-CICD-3 | `infrastructure/environments/dev/variables.tf` 64-68行 | Cloud Storage `force_destroy = true` |
| L-CICD-4 | `infrastructure/modules/github/main.tf` 43-46行 | GitHub Environmentに保護ルール未設定 |

### 良好な設計

- WIF (キーレス認証) の採用
- Dockerfile: distroless + nonroot + マルチステージビルド + バイナリストリッピング
- Gemini APIキーのSecret Manager管理
- Firestoreセキュリティルールのデフォルト拒否設計
- CIの最小権限パーミッション (`contents: read`, `pull-requests: read`)

---

## 対応優先度ランキング (全体)

### 最優先 (CRITICAL)

| 順位 | ID | 問題 | 影響 |
|:---|:---|:---|:---|
| 1 | C-CICD-1 | SA権限の分離 (ランタイム/CI/CD) | サプライチェーン攻撃パス |
| 2 | C-BE-1 | APIレート制限の実装 | 外部APIコスト増大、DoS |

### 早急に対応 (HIGH)

| 順位 | ID | 問題 | 推定工数 |
|:---|:---|:---|:---|
| 3 | H-CICD-2 | WIFにブランチ制限追加 | 小 |
| 4 | H-CICD-1 | Actions SHAピンニング | 小 |
| 5 | H-CICD-5 | deploy.yml permissions追加 | 小 |
| 6 | H-BE-1 | JSONボディサイズ制限 | 小 |
| 7 | H-BE-2 | セキュリティヘッダー追加 | 小 |
| 8 | H-CICD-3 | IAMリソースレベル付与 | 中 |
| 9 | H-CICD-4 | バイナリチェックサム検証 | 小 |
| 10 | H-BE-4 | io.ReadAllサイズ制限 | 小 |
| 11 | H-iOS-1 | APIパスパラメータサニタイズ | 小 |
| 12 | H-iOS-2 | 画像パスURL構築バリデーション | 小 |
| 13 | H-BE-3 | 画像配信エンドポイント認証 | 中 |
| 14 | H-CFG-2 | Cloud Storage CORS制限 | 小 |

### 計画的に対応 (MEDIUM)

| 順位 | ID | 問題 |
|:---|:---|:---|
| 15 | M-BE-1 | 開発用認証バイパスの安全化 |
| 16 | M-iOS-3 | Firebase App Check導入 |
| 17 | M-iOS-1 | SSL Pinning導入 |
| 18 | M-iOS-5 | テスト用メソッドのDEBUGガード |
| 19 | M-CICD-4 | persist-credentials: false |
| 20 | M-CICD-5 | Terraform state保護 |
| 21-35 | 残りMEDIUM/LOW | バリデーション強化、コード品質改善 |
