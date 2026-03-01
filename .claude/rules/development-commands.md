# 開発環境とコマンド操作

このプロジェクトでは`Taskfile`で開発コマンドを統一している。
コマンド実行やサーバー起動を求められた場合、以下を参照すること。

## 開発環境の立ち上げ

```bash
# バックエンドサーバーを起動
task run
```

Firestoreエミュレータを使用する場合:

```bash
# 1. Firestoreエミュレータを起動
firebase emulators:start --only firestore

# 2. 別ターミナルでバックエンドを起動（エミュレータ接続）
FIRESTORE_EMULATOR_HOST=localhost:8080 task run
```

## コード変更後の確認

```bash
# リントとテストを実行
task lint
task test
```

Firestore Repositoryのテストを含める場合:

```bash
firebase emulators:start --only firestore &
FIRESTORE_EMULATOR_HOST=localhost:8080 task test
```

## コマンド一覧

### 一般

| コマンド | 説明 |
|:---|:---|
| `task --list` | 利用可能なコマンド一覧を表示 |
| `task help` | 利用可能なコマンド一覧を表示 |
| `task hooks:install` | Lefthookフックをインストール |
| `task hooks:run` | pre-commitフックを手動実行 |

### バックエンド

| コマンド | 説明 |
|:---|:---|
| `task setup` | Go依存関係をダウンロード |
| `task clean` | ビルド成果物を削除 |
| `task test` | Goテストを実行 |
| `task test:coverage` | Goカバレッジ計測付きテスト |
| `task lint` | golangci-lintを実行 |
| `task format` | Goコードを整形 |
| `task build` | Goバイナリをビルド |
| `task run` | バックエンドサーバーを起動 |

### iOS

| コマンド | 説明 |
|:---|:---|
| `task ios:generate-mocks` | Mockoloでモックを生成 |
| `task ios:test` | iOSのテストを実行 |
| `task ios:test:coverage` | iOSカバレッジ計測付きテスト |
| `task ios:lint` | SwiftLintを実行 |
| `task ios:deadcode` | Swiftデッドコード検知（Periphery）を実行 |
| `task ios:format` | SwiftFormatを実行（コード整形） |
| `task ios:format-check` | SwiftFormatチェック（ローカル検証用） |
| `task ios:clean` | DerivedDataを削除 |
| `task ios:clean-all` | SPMキャッシュを含む完全クリア |
| `task ios:reset-packages` | Package.resolvedを削除して再解決 |

### デプロイ・E2E

| コマンド | 説明 |
|:---|:---|
| `task deploy:dev` | 開発環境へデプロイ（Docker build/push + Cloud Run deploy） |
| `task e2e:dev` | 開発環境のバックエンドE2Eテストを実行 |

## iOSビルドトラブル時

SPMキャッシュやFirebase SDK関連の問題が発生した場合:

```bash
# 段階1: DerivedDataのみ削除
task ios:clean

# 段階2: SPMキャッシュ含む完全クリア
task ios:clean-all

# 段階3: 既知の良いPackage.resolvedに戻す
git checkout ios/Uchikomi.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
task ios:clean-all
```

詳細は `.claude/rules/ios-build-troubleshooting.md` を参照。

## 注意

- データベースはFirestoreを使用
- Firestore Repositoryのテストは`FIRESTORE_EMULATOR_HOST`環境変数が設定されていない場合スキップされる
