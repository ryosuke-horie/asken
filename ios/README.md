# ウチコミ iOS アプリ

柔術/キックボクシングなど格闘技の減量・体重コントロールを支援するiOSアプリです。

## 技術スタック

- 言語: Swift 5.9+
- UI: SwiftUI
- 最小iOS: iOS 17
- アーキテクチャ: MVVM + Repository
- 認証: Firebase Authentication（Google Sign-In / Apple Sign-In）
- 非同期処理: async/await
- グラフ: Swift Charts

## ディレクトリ構成

```
ios/
├── Uchikomi/
│   ├── App/              # アプリエントリポイント
│   ├── Core/
│   │   ├── Auth/         # 認証マネージャー
│   │   ├── Extensions/   # Swift拡張
│   │   ├── Models/       # データモデル
│   │   ├── Network/      # APIClient, APIEndpoint, APIError
│   │   ├── Notification/ # 通知マネージャー
│   │   ├── Repositories/ # データアクセス層
│   │   └── Views/        # 共通ビュー（カメラ等）
│   ├── Features/
│   │   ├── Auth/              # ログイン画面
│   │   ├── CookingSuggestion/ # メニューサジェスト画面
│   │   ├── Meals/             # 食事記録画面
│   │   ├── MyMenu/            # マイメニュー画面
│   │   ├── Pantry/            # 食材管理画面
│   │   ├── Settings/          # 設定画面
│   │   └── Weight/            # 体重記録画面
│   ├── Shared/
│   │   ├── Components/   # 共通UIコンポーネント
│   │   ├── Models/       # 共有モデル（通知設定等）
│   │   └── Theme.swift   # テーマ定義
│   └── Resources/        # アセット、Info.plist
├── UchikomiCore/          # コアフレームワーク
│   ├── Auth/             # 認証関連
│   └── Models/           # 共有モデル
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
   - Include Tests にチェック
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

1. バックエンドを起動: `task run`
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

### 体重記録
- 体重の入力・記録
- 体重推移の表示
- 目標体重の設定

### マイメニュー
- よく食べるメニューの登録・管理
- マイメニューからの食事記録

### 栄養目標
- PFC（タンパク質・脂質・炭水化物）目標設定
- 微量栄養素の表示

### 食材管理（パントリー）
- 食材の登録・管理（カテゴリ・賞味期限対応）
- レシート読取による食材一括登録（Gemini API）

### メニューサジェスト
- 手持ちの食材と栄養目標に基づくメニュー提案（Gemini API）
- サジェストの採用・却下

### 設定
- 通知設定

## API エンドポイント

バックエンドAPI（Go）をそのまま使用:

| エンドポイント | 用途 |
|:---|:---|
| POST /api/analyze | 画像分析開始 |
| GET /api/analyze/{id} | 分析ステータス確認 |
| POST /api/upload-image | 画像アップロード |
| GET /api/images/{uuid} | 画像取得 |
| GET /api/history | 分析履歴一覧 |
| GET /api/history/{id} | 分析履歴詳細 |
| PUT /api/history/{id} | 分析履歴更新 |
| DELETE /api/history/{id} | 分析履歴削除 |
| GET /api/meals/daily | 日別食事取得 |
| POST /api/meals/skip | 食事スキップ |
| GET /api/weight/records | 体重記録一覧 |
| POST /api/weight/records | 体重記録作成 |
| GET /api/weight/records/{id} | 体重記録詳細 |
| PUT /api/weight/records/{id} | 体重記録更新 |
| DELETE /api/weight/records/{id} | 体重記録削除 |
| GET /api/weight/goal | 体重目標取得 |
| PUT /api/weight/goal | 体重目標更新 |
| GET /api/nutrition/goal | 栄養目標取得 |
| PUT /api/nutrition/goal | 栄養目標設定 |
| GET /api/my-menu | マイメニュー一覧 |
| POST /api/my-menu | マイメニュー作成 |
| GET /api/my-menu/{id} | マイメニュー詳細 |
| PUT /api/my-menu/{id} | マイメニュー更新 |
| DELETE /api/my-menu/{id} | マイメニュー削除 |
| POST /api/my-menu/{id}/record | マイメニューから食事記録 |
| GET /api/ingredients | 食材一覧 |
| POST /api/ingredients | 食材作成 |
| PUT /api/ingredients/{id} | 食材更新 |
| DELETE /api/ingredients/{id} | 食材削除 |
| POST /api/ingredients/scan-receipt | レシート読取 |
| POST /api/menu/suggest | メニューサジェストリクエスト |
| GET /api/menu/suggestions | サジェスト一覧 |
| GET /api/menu/suggestions/{id} | サジェスト詳細 |
| POST /api/menu/suggestions/{id}/accept | サジェスト採用 |
| POST /api/menu/suggestions/{id}/dismiss | サジェスト却下 |
