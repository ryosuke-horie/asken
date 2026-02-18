---
name: deploy-backend
description: バックエンドを開発環境（Cloud Run）にデプロイする。プリフライトチェック、Firestoreインデックス適用、デプロイ実行、E2Eテストを一連のフローで実施。
user_invoked_only: true
---

# バックエンドデプロイスキル

ローカルから開発環境（Cloud Run）へバックエンドをデプロイするためのスキル。
各フェーズでユーザーに確認を取りながら安全にデプロイを進める。

> このスキルはユーザーが `/deploy-backend` で明示的に呼び出した場合のみ実行すること。
> エージェントが自律的にこのスキルを参照・実行してはならない。
> デプロイは本番環境に影響を与える操作であり、必ずユーザーの意図に基づいて実行する。

## Phase 1: プリフライトチェック

デプロイ前に以下をすべて確認し、問題があれば即座にユーザーに報告して中断する。

### 1.1 mainブランチの確認

```bash
git branch --show-current
```

main ブランチ以外にいる場合は中断し、ユーザーに確認を取る。

### 1.2 ワーキングツリーの確認

```bash
git status --porcelain
```

未コミットの変更がある場合は中断し、ユーザーに確認を取る。

### 1.3 gcloud認証の確認

```bash
gcloud auth list --filter=status:ACTIVE --format='value(account)' | head -n 1
```

アクティブなアカウントがない場合は中断し、`gcloud auth login` を案内する。

### 1.4 Docker起動確認

```bash
docker info > /dev/null 2>&1
```

Docker が起動していない場合は中断する。

すべてのチェックが通過したら、結果をまとめてユーザーに報告し、AskUserQuestion で続行確認を取る。

---

## Phase 2: Firestoreインデックス・ルール適用

`firestore.indexes.json` や `firestore.rules` に変更がある場合のみ実行する。
変更がない場合はこのフェーズをスキップする。

### 2.1 インデックスの差分確認と適用

直近の変更で `firestore.indexes.json` に差分があるか確認する:

```bash
git diff HEAD~1 -- firestore.indexes.json
```

差分がある場合、AskUserQuestion でユーザーに確認を取ってから実行:

```bash
firebase deploy --only firestore:indexes --project "${GCP_PROJECT_ID}"
```

注意: インデックスの作成には時間がかかる場合がある。Firebase Console でステータスを確認できることをユーザーに伝える。

### 2.2 ルールの差分確認と適用

直近の変更で `firestore.rules` に差分があるか確認する:

```bash
git diff HEAD~1 -- firestore.rules
```

差分がある場合、AskUserQuestion でユーザーに確認を取ってから実行:

```bash
firebase deploy --only firestore:rules --project "${GCP_PROJECT_ID}"
```

### 2.3 データ整合性の確認

スキーマ変更（フィールド追加・変更・削除）を含むコード変更がある場合:

- 既存データとの後方互換性をコードレベルで確認する
- 新しいフィールドにデフォルト値やゼロ値ハンドリングがあるか確認する
- 問題があればユーザーに報告し、デプロイ続行の判断を委ねる

---

## Phase 3: バックエンドデプロイ

### 3.1 デプロイ実行

```bash
task deploy:dev
```

このコマンドは `tools/deploy/deploy-dev.sh` を実行し、以下を行う:
1. Docker イメージをビルド
2. Artifact Registry にプッシュ
3. Cloud Run にデプロイ

### 3.2 デプロイ結果の確認

デプロイ完了後、サービスURLを取得して表示:

```bash
gcloud run services describe "${CLOUD_RUN_SERVICE_NAME}" \
  --region "${GCP_REGION}" \
  --project "${GCP_PROJECT_ID}" \
  --format 'value(status.url)'
```

---

## Phase 4: デプロイ後検証

### 4.1 ヘルスチェック

```bash
curl -s -o /dev/null -w "%{http_code}" "${SERVICE_URL}/api/health"
```

200 以外の場合はユーザーに報告する。

### 4.2 E2Eテスト

AskUserQuestion で E2E テストを実行するか確認を取る:

```bash
task e2e:dev
```

注意: Gemini API のレート制限により、テスト実行には数分かかる場合がある。

### 4.3 結果報告

デプロイ結果をまとめてユーザーに報告する:

- デプロイ先URL
- ヘルスチェック結果
- E2Eテスト結果（実行した場合）
- Firestoreインデックス・ルールの適用状況

---

## 環境変数

`.mise.toml` で以下が設定されている前提:

| 変数 | 説明 |
|:---|:---|
| `GCP_PROJECT_ID` | GCPプロジェクトID |
| `GCP_REGION` | GCPリージョン |
| `CLOUD_RUN_SERVICE_NAME` | Cloud Runサービス名 |
| `ARTIFACT_REGISTRY_URL` | Artifact Registry URL |

E2Eテスト用:

| 変数 | 説明 |
|:---|:---|
| `E2E_FIREBASE_API_KEY` | Firebase APIキー |
| `SERVICE_ACCOUNT_EMAIL` | サービスアカウントメール |
| `E2E_TEST_UID` | テスト用UID |

---

## トラブルシューティング

| 問題 | 対応 |
|:---|:---|
| gcloud認証エラー | `gcloud auth login` を実行 |
| Docker認証エラー | `gcloud auth configure-docker asia-northeast1-docker.pkg.dev` |
| インデックス作成遅延 | Firebase Console でステータス確認。数分待機 |
| E2Eテスト失敗（レート制限） | Gemini APIのレート制限（5秒/リクエスト）。再実行で解決する場合がある |
| Cloud Runデプロイ失敗 | デプロイSAの権限を確認。Terraform再実行 |

---

## ロールバック手順

デプロイ後に問題が発見された場合、以下の手順でロールバックする。

### Cloud Run のロールバック

Cloud Run のリビジョン一覧を確認し、前のリビジョンにトラフィックを切り替える:

```bash
# リビジョン一覧を確認
gcloud run revisions list \
  --service "${CLOUD_RUN_SERVICE_NAME}" \
  --region "${GCP_REGION}" \
  --project "${GCP_PROJECT_ID}"

# 前のリビジョンにトラフィックを100%切り替え
gcloud run services update-traffic "${CLOUD_RUN_SERVICE_NAME}" \
  --to-revisions <previous-revision>=100 \
  --region "${GCP_REGION}" \
  --project "${GCP_PROJECT_ID}"
```

### Firestoreインデックス・ルールのロールバック

インデックスやルールをロールバックする必要がある場合:

```bash
# 前のバージョンのファイルを復元
git checkout <previous-sha> -- firestore.indexes.json firestore.rules

# 復元したファイルでデプロイ
firebase deploy --only firestore:indexes,firestore:rules --project "${GCP_PROJECT_ID}"

# ファイルを元に戻す
git checkout HEAD -- firestore.indexes.json firestore.rules
```

### 部分的な失敗への対応

| 状況 | 対応 |
|:---|:---|
| Firestore適用成功 + コードデプロイ失敗 | コードの問題を修正して再デプロイ。インデックスは後方互換のため問題なし |
| コードデプロイ成功 + E2Eテスト失敗 | テスト失敗の原因を調査。コードの問題ならロールバック |
| 全体が成功 + 本番で問題発見 | Cloud Run リビジョンロールバック + 必要に応じてFirestoreロールバック |
