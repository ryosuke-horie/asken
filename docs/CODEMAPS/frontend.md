# フロントエンドアーキテクチャ

最終更新: 2026-01-30
フレームワーク: Next.js 15 (App Router)
エントリーポイント: frontend/app/layout.tsx

## ディレクトリ構造

```
frontend/
├── app/                    # App Router
│   ├── layout.tsx         # ルートレイアウト
│   ├── page.tsx           # ホーム（日次記録）
│   ├── login/             # ログイン
│   ├── register/          # 新規登録
│   ├── meals/             # 食事記録
│   │   └── [mealType]/    # 食事タイプ別
│   ├── mylist/            # マイリスト
│   │   ├── new/           # 新規作成
│   │   └── [id]/edit/     # 編集
│   ├── settings/          # 設定
│   └── training/          # トレーニング
│       ├── suggest/       # AIメニュー提案
│       └── locations/     # 場所管理
│           └── [id]/equipment/ # 器具管理
├── components/
│   ├── server/            # Server Components
│   └── client/            # Client Components
├── contexts/              # Reactコンテキスト
├── hooks/                 # カスタムフック
├── lib/                   # ユーティリティ
│   └── constants/         # 定数
├── types/                 # 型定義
└── e2e/                   # E2Eテスト
    ├── fixtures/          # テストフィクスチャ
    ├── pages/             # ページオブジェクト
    └── tests/             # テストファイル
```

## コンポーネント構成

### ルート構成 (app/)

| パス | ページ | 用途 |
|:---|:---|:---|
| / | page.tsx | ホーム（日次記録ダッシュボード） |
| /login | login/page.tsx | ログイン |
| /register | register/page.tsx | 新規登録 |
| /meals/[mealType] | meals/[mealType]/page.tsx | 食事記録（朝・昼・夕・間食） |
| /mylist | (ホームから遷移) | マイリスト一覧 |
| /mylist/new | mylist/new/page.tsx | マイリスト新規作成 |
| /mylist/[id]/edit | mylist/[id]/edit/page.tsx | マイリスト編集 |
| /settings | settings/page.tsx | 設定（目標体重等） |
| /training/suggest | training/suggest/page.tsx | AIトレーニング提案 |
| /training/locations | training/locations/page.tsx | 場所管理 |
| /training/locations/[id]/equipment | training/locations/[id]/equipment/page.tsx | 器具管理 |

### Server Components (components/server/)

| コンポーネント | 用途 |
|:---|:---|
| Pagination.tsx | ページネーション |

### Client Components (components/client/)

| コンポーネント | 用途 |
|:---|:---|
| Navigation.tsx | ナビゲーションメニュー |
| ProtectedRoute.tsx | 認証ガード |
| DateNavigation.tsx | 日付選択 |
| DailyMealsView.tsx | 日次食事表示 |
| DailyTotalSummary.tsx | 日次合計サマリー |
| MealSection.tsx | 食事セクション |
| MealThumbnail.tsx | 食事サムネイル |
| MealTypeUpload.tsx | 食事アップロード |
| MealUploadView.tsx | 食事アップロード画面 |
| MealInputSelector.tsx | 入力方法選択 |
| ImageUpload.tsx | 画像アップロード |
| TextInput.tsx | テキスト入力 |
| NutritionDisplay.tsx | 栄養素表示 |
| SkippedMealButton.tsx | 食事スキップボタン |
| WeightSection.tsx | 体重セクション |
| WeightChart.tsx | 体重グラフ |
| WeightRecordForm.tsx | 体重記録フォーム |
| WeightGoalSetting.tsx | 目標体重設定 |
| ConditionSection.tsx | 体調セクション |
| ConditionRecordForm.tsx | 体調記録フォーム |
| TrainingSection.tsx | トレーニングセクション |
| MylistForm.tsx | マイリストフォーム |
| MylistSelector.tsx | マイリスト選択 |
| MylistAnalyzeInput.tsx | マイリストAI分析入力 |
| QuantityStepper.tsx | 数量ステッパー |

## カスタムフック (hooks/)

| フック | 用途 |
|:---|:---|
| useAnalysisPolling.ts | 分析ステータスポーリング |
| useMylist.ts | マイリストデータ管理 |
| useCondition.ts | 体調データ管理 |
| useTraining.ts | トレーニングデータ管理 |
| useProfile.ts | プロフィールデータ管理 |

## コンテキスト (contexts/)

| コンテキスト | 用途 |
|:---|:---|
| AuthContext.tsx | 認証状態管理（トークン、ユーザー情報） |

## 型定義 (types/)

| ファイル | 内容 |
|:---|:---|
| nutrition.ts | 食事・栄養素関連 |
| weight.ts | 体重関連 |
| condition.ts | 体調関連 |
| training.ts | トレーニング関連 |
| mylist.ts | マイリスト関連 |
| profile.ts | プロフィール関連 |

## ユーティリティ (lib/)

| ファイル | 用途 |
|:---|:---|
| fetcher.ts | SWR用フェッチャー |
| date.ts | 日付ユーティリティ |
| storage.ts | ローカルストレージ操作 |
| constants/auth.ts | 認証関連定数 |

## データフロー

```
ユーザー操作
    ↓
Client Component
    ↓
カスタムフック (SWRでキャッシュ管理)
    ↓
fetcher (Authorization付きfetch)
    ↓
バックエンドAPI
```

## 認証フロー

```
1. AuthContext → localStorage/cookieからトークン復元
2. ProtectedRoute → 未認証時は/loginへリダイレクト
3. middleware.ts → サーバーサイドで認証チェック
4. createAuthFetcher → APIリクエストにトークン付与
```

## E2Eテスト構成 (e2e/)

### ページオブジェクト (pages/)

| ファイル | 対象ページ |
|:---|:---|
| LoginPage.ts | ログイン |
| HomePage.ts | ホーム |
| MealsPage.ts | 食事記録 |
| TrainingPage.ts | トレーニング |
| LocationsPage.ts | 場所管理 |
| EquipmentPage.ts | 器具管理 |

### テストファイル (tests/)

| ファイル | テスト内容 |
|:---|:---|
| auth.spec.ts | 認証フロー |
| meals.spec.ts | 食事記録フロー |
| training.spec.ts | トレーニングフロー |

## 依存関係図

```
app/layout.tsx
├── contexts/AuthContext.tsx
└── components/client/Navigation.tsx

app/page.tsx (ホーム)
├── components/client/DateNavigation.tsx
├── components/client/DailyMealsView.tsx
│   └── components/client/MealSection.tsx
├── components/client/WeightSection.tsx
├── components/client/ConditionSection.tsx
└── components/client/TrainingSection.tsx

app/meals/[mealType]/page.tsx
├── components/client/MealUploadView.tsx
│   ├── components/client/MealInputSelector.tsx
│   ├── components/client/ImageUpload.tsx
│   ├── components/client/TextInput.tsx
│   └── components/client/MylistSelector.tsx
└── hooks/useAnalysisPolling.ts
```

## 関連コードマップ

- [全体アーキテクチャ](./architecture.md)
- [データモデル](./data.md)
