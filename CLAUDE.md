# ウチコミ プロジェクトガイドライン

このドキュメントは、Claude Codeがこのプロジェクトで作業する際のガイドラインを定義します。

## 共通ルール

- 不明点がある場合はAskUserQuestionツールを使用すること
- リリース品質を前提とした実装を行うこと（品質やテストの妥協はしない）
- エージェントメモリ（MEMORY.md）を含む全ての出力は日本語で記述すること

## 詳細規約

詳細なコーディング規約、ベストプラクティス、セキュリティガイドライン等は `.claude/rules/` ディレクトリを参照してください。

### 基本ルール

| ファイル | 内容 |
|:---|:---|
| `testing-details.md` | テスト詳細（古典派スタイル、モック基準） |
| `git-workflow-details.md` | Git/GitHubワークフロー詳細（Linear連携、PRレビュー、マージ戦略） |
| `coding-style.md` | コーディングスタイル（KISS, DRY, YAGNI） |
| `security-details.md` | セキュリティ詳細（コミット前チェック） |
| `performance.md` | パフォーマンス最適化（モデル選択、コンテキスト管理） |
| `patterns.md` | 共通パターン（APIレスポンス形式、Repositoryパターン） |
| `dependencies.md` | 依存関係管理（dependabot設定） |
| `development-commands.md` | 開発コマンド（Taskfile） |
| `documentation.md` | ドキュメント更新ルール |

### 技術別ルール

| ファイル | 内容 |
|:---|:---|
| `ios-testing.md` | iOSテストガイドライン（Swift Testing, Mockolo） |
| `backend-golang.md` | バックエンド開発規約（Go） |
| `gemini-api.md` | Gemini API連携のベストプラクティス |
| `firestore.md` | Firestore規約（インデックス管理） |
| `testing-tdd.md` | テスト駆動開発（TDD） |
| `security-practices.md` | セキュリティガイドライン |
| `ios-build-troubleshooting.md` | iOSビルドトラブル対応（SPMキャッシュ、Firebase SDK） |

### Claude Code拡張

| ファイル | 内容 |
|:---|:---|
| `hooks.md` | フックシステム（PreToolUse, PostToolUse, Stop） |
| `plan-mode.md` | Plan Modeルール（プラン保存、Linear連携） |

### 即座に使用すべきエージェント

ユーザーからの明示的な指示なしでプロアクティブに使用すること:

| トリガー | エージェント |
|:---|:---|
| 複雑な機能リクエスト | planner |
| コード作成/修正後 | code-reviewer |
| バグ修正または新機能 | tdd-guide |

利用可能なエージェント一覧は `.claude/agents/` を参照。

---

## iOS実機ビルドに関する注意事項

- 現在Apple Developer Program（有料）に未登録のため、Sign In with Apple の entitlement があると実機ビルドできない
- 実機ビルド時は以下の2ファイルからSign In with Appleを一時的に無効化する（コミットしないこと）:
  - `ios/Uchikomi/Uchikomi.entitlements`: `com.apple.developer.applesignin` エントリを削除
  - `ios/project.yml`: `com.apple.developer.applesignin` プロパティを削除（`properties: {}` に変更）
- 確認後は `git checkout ios/Uchikomi/Uchikomi.entitlements ios/project.yml` で元に戻す
- Apple Developer Program登録後（本リリース時）にこのメモを削除すること

---

## iOSビルド/テストに関する注意事項

- `task ios:test` で Xcode ビルドシステムがクラッシュする場合がある
- これはmacOSとXcodeのバージョンの問題であり、アプリのコードに問題があるわけではない
- クラッシュが発生した場合はiOSテストをスキップし、ユーザーに対応を委ねること
- Goバックエンドのテスト（`task test`）とリント（`task lint`）は通常通り実行すること

### iOSテストについて

- iOSテストは作成しないこと
- macOSおよびXcodeのバージョンアップにより、iOSテストが頻繁に壊れる問題が発生している
- この問題が解消するまで、iOSのテストコードは作成せず、手動テストで代用する
- 詳細は `.claude/rules/ios-testing-policy.md` を参照

---

## Terraform認証

### ローカル開発でのTerraform実行

ローカルでTerraformを実行する場合、Google Application Default Credentials (ADC) が必要です:

```bash
# Google Cloud SDKで認証（ADCを設定）
gcloud auth application-default login

# Terraformの初期化と実行
cd infrastructure/environments/dev
terraform init
terraform plan
terraform apply
```

### Backend設定

- TerraformのstateはGCS (Google Cloud Storage) に保存
- バケット: `utikomi-dev-tfstate`
- パス: `terraform/state/dev`
- 認証: Google Application Default Credentialsを使用

### Workload Identity Federation (WIF)

GitHub ActionsからGoogle Cloudへキーレス認証で接続する仕組み:

| 設定項目 | 値 |
|:---|:---|
| OIDC Issuer | `https://token.actions.githubusercontent.com` |
| プール | GitHub Actions用 |
| プロバイダー | GitHub |
| 制限 | 特定リポジトリのみ (`attribute_condition` で制限) |

### GitHub Actionsでの認証フロー

1. GitHub ActionsがOIDCトークンを取得 (`id-token: write` パーミッションが必要)
2. `google-github-actions/auth` アクションでWIFプロバイダーとサービスアカウントを指定
3. Google Cloud一時クレデンシャルを取得してgcloud/Docker操作を実行

```yaml
permissions:
  id-token: write  # WIFで必須

- name: Authenticate to Google Cloud
  uses: google-github-actions/auth@v3
  with:
    workload_identity_provider: ${{ vars.WORKLOAD_IDENTITY_PROVIDER }}
    service_account: ${{ vars.DEPLOY_SERVICE_ACCOUNT_EMAIL }}
```

### 環境変数（GitHub Actions Variables）

以下の変数をGitHub Repository Settingsに設定済み:

- `GCP_REGION`
- `ARTIFACT_REGISTRY_URL`
- `CLOUD_RUN_SERVICE_NAME`
- `WORKLOAD_IDENTITY_PROVIDER`
- `DEPLOY_SERVICE_ACCOUNT_EMAIL`
- `SERVICE_ACCOUNT_EMAIL` (ランタイム用)

---

## 関連ドキュメント

| ドキュメント | 内容 |
|:---|:---|
| [docs/CONTRIB.md](./docs/CONTRIB.md) | 開発者ガイド（環境セットアップ、コマンド一覧） |
| [docs/RUNBOOK.md](./docs/RUNBOOK.md) | 運用手順書 |
| [docs/CODEMAPS/](./docs/CODEMAPS/) | コードマップ |
| [docs/adr/](./docs/adr/) | アーキテクチャ決定記録 |

