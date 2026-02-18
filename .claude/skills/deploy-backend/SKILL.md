---
name: deploy-backend
description: バックエンドを開発環境（Cloud Run）にデプロイする。プリフライトチェック、Firestoreインデックス適用、デプロイ実行、E2Eテストを一連のフローで実施。
---

# バックエンドデプロイスキル

ローカルから開発環境（Cloud Run）へバックエンドをデプロイするためのスキル。
各フェーズでユーザーに確認を取りながら安全にデプロイを進める。

## 前提条件

- gcloud CLI がインストール済みで認証済みであること
- Docker が起動していること
- firebase CLI がインストール済みであること（Firestoreインデックス適用時）
- 環境変数が `.mise.toml` で設定済みであること

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

### 1.3 リモートとの同期確認

```bash
git fetch origin main
git log HEAD..origin/main --oneline
```

ローカルがリモートより遅れている場合は、`git pull` を提案する。

### 1.4 gcloud認証の確認

```bash
gcloud auth list --filter=status:ACTIVE --format='value(account)' | head -n 1
```

アクティブなアカウントがない場合は中断し、`gcloud auth login` を案内する。

### 1.5 Docker起動確認

```bash
docker info > /dev/null 2>&1
```

Docker が起動していない場合は中断する。

### 1.6 テスト・リントの実行

```bash
task lint
task test
```

lint またはテストが失敗した場合はデプロイを中断する。

すべてのチェックが通過したら、結果をまとめてユーザーに報告し、AskUserQuestion で続行確認を取る。

---

## Phase 2: 変更差分の確認

ユーザーがデプロイ内容を把握できるよう、変更差分を提示する。

### 2.1 前回デプロイからの変更

Cloud Run に現在デプロイされているイメージのタグ（git SHA）を取得し、そこからの差分を表示する。

```bash
# 現在デプロイ済みのイメージタグを取得
gcloud run services describe "${CLOUD_RUN_SERVICE_NAME}" \
  --region "${GCP_REGION}" \
  --project "${GCP_PROJECT_ID}" \
  --format 'value(spec.template.spec.containers[0].image)'
```

イメージタグから git SHA を抽出し、変更ログを表示:

```bash
git log <deployed-sha>..HEAD --oneline
```

取得できない場合は直近の10コミットを表示する。

### 2.2 Firestoreインデックスの差分

```bash
git diff <deployed-sha>..HEAD -- firestore.indexes.json
```

差分がある場合は、追加・削除されるインデックスを一覧表示する。

### 2.3 Firestoreルールの差分

```bash
git diff <deployed-sha>..HEAD -- firestore.rules
```

差分がある場合は、変更内容を表示する。

### 2.4 バックエンドコードの変更サマリ

```bash
git diff <deployed-sha>..HEAD --stat -- backend/
```

変更差分をまとめてユーザーに報告する。

---

## Phase 3: Firestoreインデックス・ルール適用

Phase 2 でインデックスまたはルールに差分があった場合のみ実行する。
差分がない場合はこのフェーズをスキップする。

### 3.1 インデックスの適用

差分がある場合、AskUserQuestion でユーザーに確認を取ってから実行:

```bash
firebase deploy --only firestore:indexes --project "${GCP_PROJECT_ID}"
```

注意: インデックスの作成には時間がかかる場合がある。Firebase Console でステータスを確認できることをユーザーに伝える。

### 3.2 ルールの適用

差分がある場合、AskUserQuestion でユーザーに確認を取ってから実行:

```bash
firebase deploy --only firestore:rules --project "${GCP_PROJECT_ID}"
```

### 3.3 データ整合性の確認

スキーマ変更（フィールド追加・変更・削除）を含むコード変更がある場合:

- 既存データとの後方互換性をコードレベルで確認する
- 新しいフィールドにデフォルト値やゼロ値ハンドリングがあるか確認する
- 問題があればユーザーに報告し、デプロイ続行の判断を委ねる

---

## Phase 4: バックエンドデプロイ

### 4.1 デプロイ実行

AskUserQuestion で最終確認を取ってから実行:

```bash
task deploy:dev
```

このコマンドは `tools/deploy/deploy-dev.sh` を実行し、以下を行う:
1. Docker イメージをビルド
2. Artifact Registry にプッシュ
3. Cloud Run にデプロイ

### 4.2 デプロイ結果の確認

デプロイ完了後、サービスURLを取得して表示:

```bash
gcloud run services describe "${CLOUD_RUN_SERVICE_NAME}" \
  --region "${GCP_REGION}" \
  --project "${GCP_PROJECT_ID}" \
  --format 'value(status.url)'
```

---

## Phase 5: デプロイ後検証

### 5.1 ヘルスチェック

```bash
curl -s -o /dev/null -w "%{http_code}" "${SERVICE_URL}/api/health"
```

200 以外の場合はユーザーに報告する。

### 5.2 E2Eテスト

AskUserQuestion で E2E テストを実行するか確認を取る:

```bash
task e2e:dev
```

注意: Gemini API のレート制限により、テスト実行には数分かかる場合がある。

### 5.3 結果報告

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
