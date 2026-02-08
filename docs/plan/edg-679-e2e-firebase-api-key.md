# プラン: E2E_FIREBASE_API_KEY シークレット設定

## Linear Issue
- Issue: EDG-679
- URL: https://linear.app/ryosuke-horie/issue/EDG-679

## Context

PR #134 で Cloud Run デプロイ後の API E2E テストを追加・マージしたが、GitHub Secrets に `E2E_FIREBASE_API_KEY` が未設定のため e2e-test ジョブが失敗している。

失敗ワークフロー: https://github.com/ryosuke-horie/utikomi/actions/runs/21791388627

CI環境での必須チェック（サイレントスキップ防止）が正しく機能し、明示的にエラー終了した。

## E2E_FIREBASE_API_KEY の用途

`backend/e2e/auth.go` で Firebase Identity Toolkit REST API を呼び出す際のクエリパラメータとして使用:

```
https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key={API_KEY}
```

カスタムトークンをIDトークンに交換し、認証が必要なAPIエンドポイント（POST /api/analyze 等）のE2Eテストに使用する。

## 対応手順

### 手順1: Firebase Web API Key を取得

1. Firebase Console (https://console.firebase.google.com/) にアクセス
2. utikomi-dev プロジェクトを選択
3. プロジェクト設定 → 全般 → 「Web API キー」をコピー

### 手順2: GitHub Secrets に登録

以下のいずれかの方法で設定:

(A) GitHub CLI:
```bash
gh secret set E2E_FIREBASE_API_KEY --repo ryosuke-horie/utikomi
# プロンプトでAPIキーを入力
```

(B) GitHub UI:
1. https://github.com/ryosuke-horie/utikomi/settings/secrets/actions にアクセス
2. 「New repository secret」をクリック
3. Name: `E2E_FIREBASE_API_KEY`、Value: 取得したAPIキーを入力
4. 「Add secret」をクリック

### 手順3: ワークフロー再実行

```bash
gh workflow run deploy.yml
```

## 検証

1. deploy.yml のワークフロー実行を監視
2. e2e-test ジョブの全ステップが成功することを確認:
   - Get Cloud Run URL: 成功
   - Run E2E tests: 成功（ヘルスチェック + 認証テスト全て PASS）
   - Report E2E results: サマリに PASS 件数が表示される
3. GitHub Step Summary でテスト結果を確認

## 注意事項

- コード変更は不要（手動設定のみ）
- Firebase Web API Key はクライアントサイドで使用されるものであり、機密性は低いが、GitHub Secrets で管理するのが適切
