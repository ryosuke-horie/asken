# プラン: option.WithCredentialsFile の非推奨API対応

## Linear Issue
- Issue: EDG-610
- URL: https://linear.app/ryosuke-horie/issue/EDG-610

## 概要

`google.golang.org/api` を 0.256.0 -> 0.265.0 にアップデートすると、`option.WithCredentialsFile` が deprecated になり、staticcheck (SA1019) でエラーが発生する問題を修正する。

## 対象ファイル

| ファイル | 行 | 用途 |
|:---|:---|:---|
| `backend/pkg/database/firestore.go` | 19 | Firestoreクライアント初期化 |
| `backend/pkg/storage/client.go` | 18 | Cloud Storageクライアント初期化 |
| `backend/internal/service/firebase_auth_service.go` | 29 | Firebase Authサービス初期化 |

## 実装計画

### ステップ1: 依存関係の更新

```bash
cd backend
go get google.golang.org/api@v0.265.0
go mod tidy
```

### ステップ2: 3ファイルのコードを修正

変更パターン（共通）:

```go
// 変更前
opt := option.WithCredentialsFile(credentialsPath)

// 変更後
opt := option.WithAuthCredentialsFile(option.ServiceAccount, credentialsPath)
```

各ファイルの変更:
1. `backend/pkg/database/firestore.go:19`
2. `backend/pkg/storage/client.go:18`
3. `backend/internal/service/firebase_auth_service.go:29`

### ステップ3: 検証

```bash
task lint   # staticcheck SA1019が解消されていることを確認
task test   # 既存のテストがパスすることを確認
task build  # ビルドが成功することを確認
```

## 技術的な考慮事項

- 低リスクな変更: APIの動作は変わらない
- 既存のテストで十分: モックベースのテストが存在
- ロールバック容易: 1行の変更のみで3ファイル

## テスト計画

既存のテストで動作を確認:
- `task test` - ユニットテスト実行
- `task lint` - staticcheck (SA1019) がパスすることを確認
- `task build` - コンパイルが成功することを確認

## 検証手順

1. コード変更後、`task lint` で staticcheck エラーが解消されたことを確認
2. `task test` で既存テストがパスすることを確認
3. `task build` でビルドが成功することを確認
4. （オプション）サーバー起動して動作確認
