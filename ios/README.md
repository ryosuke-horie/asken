# ウチコミ iOS アプリ

柔術/キックボクシングなど格闘技の減量・体重コントロールを支援するiOSアプリです。

## 技術スタック

- **言語**: Swift 5.9+
- **UI**: SwiftUI
- **最小iOS**: iOS 17
- **アーキテクチャ**: MVVM + Repository
- **認証**: Keychain（JWT保存）
- **非同期処理**: async/await
- **グラフ**: Swift Charts

## ディレクトリ構成

```
ios/
├── Uchikomi/
│   ├── App/              # アプリエントリポイント
│   ├── Core/
│   │   ├── Network/      # APIClient, TokenManager
│   │   ├── Models/       # データモデル
│   │   └── Repositories/ # データアクセス層
│   ├── Features/
│   │   ├── Auth/         # ログイン画面
│   │   └── Meals/        # 食事記録画面
│   ├── Shared/
│   │   ├── Components/   # 共通UIコンポーネント
│   │   └── Extensions/   # Swift拡張
│   └── Resources/        # アセット、Info.plist
└── UchikomiTests/        # ユニットテスト
```

## セットアップ

### 1. Xcodeプロジェクト作成

1. Xcodeを起動
2. File > New > Project
3. iOS > App を選択
4. 設定:
   - Product Name: `Uchikomi`
   - Organization Identifier: `dev.exe`
   - Interface: SwiftUI
   - Language: Swift
   - ✅ Include Tests
5. `ios/` ディレクトリ内に保存

### 2. 既存ファイルの追加

作成されたプロジェクトに、このディレクトリ内のファイルを追加:
1. 不要なテンプレートファイルを削除（ContentView.swift等）
2. File > Add Files to "Uchikomi"
3. `Uchikomi/` フォルダ以下のすべてのファイルを追加

### 3. 開発用の設定

`AppEnvironment.swift` で開発環境のURLを設定:

```swift
// 実機テスト時は Mac の IP アドレスに変更
return URL(string: "http://192.168.1.100:8080")!
```

Mac の IP アドレスは `ifconfig | grep "inet "` で確認できます。

## ビルド & 実行

### シミュレータで実行

1. バックエンドを起動: `task db-up && task run`
2. Xcodeでシミュレータを選択してビルド

### 実機で実行（無料プロビジョニング）

1. iPhone を Mac に接続
2. Xcode でデバイスを選択
3. 初回は「Trust this computer」を iPhone で承認
4. Team を Personal Team に設定（Signing & Capabilities）
5. ビルド & 実行

※ 無料プロビジョニングは7日で期限切れ。週1回の再ビルドが必要。

## テストユーザー

```
メール: test@example.com
パスワード: Pass0123
```

## 主要機能

### 食事記録
- カメラ/フォトライブラリから画像選択
- Gemini APIによる自動栄養素分析
- 食事タイプ（朝食/昼食/夕食/間食）選択
- 日別の栄養素サマリー表示

## API エンドポイント

バックエンドAPI（Go）をそのまま使用:

| エンドポイント | 用途 |
|:---|:---|
| POST /api/auth/login | ログイン |
| GET /api/meals/daily | 日別食事取得 |
| POST /api/analyze | 画像分析開始 |
| GET /api/analyze/:id/status | 分析ステータス確認 |
| GET /api/analyze/:id/result | 分析結果取得 |
| POST /api/meals | 食事保存 |
