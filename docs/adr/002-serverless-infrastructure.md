# ADR-002: サーバレスインフラへの移行

## ステータス

承認済み

## コンテキスト

現在のウチコミは以下の構成で運用されている：

- **ホスティング**: exe.dev（Ubuntu VM）
- **バックエンド**: Go（:8080）
- **フロントエンド**: Next.js（:3000）
- **データベース**: PostgreSQL
- **AI**: Gemini CLI（gemini-3-flash-preview）

以下の課題がある：

1. **コスト**: VMは常時稼働で課金される
2. **運用負荷**: OS/ミドルウェアの管理が必要
3. **スケーラビリティ**: 手動スケーリングが必要
4. **iOSアプリ対応**: iOSからAPIへのアクセスが主要ユースケースになる

## 決定

### インフラ構成

| コンポーネント | 現行 | 移行先 |
|:---|:---|:---|
| クライアント | Next.js + iOS | **iOSのみ** |
| バックエンドAPI | exe.dev VM (Go) | **Cloud Run (Go)** |
| データベース | PostgreSQL | **Firestore** |
| AI | Gemini CLI | **Gemini API (gemini-2.0-flash)** |
| 画像ストレージ | ローカルファイル | **Cloud Storage** |
| 認証 | 自前JWT | **Firebase Authentication** |

### アーキテクチャ図

```
┌─────────────────────────────────────────────────────────────────┐
│                      iOSアプリ (SwiftUI)                         │
│   ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐ │
│   │  認証           │  │  食事記録       │  │  体重/体調     │ │
│   │  (Firebase Auth)│  │  (カメラ/入力)  │  │  トレーニング  │ │
│   └────────┬────────┘  └────────┬────────┘  └───────┬────────┘ │
└────────────┼───────────────────┼──────────────────┼────────────┘
             │                   │                  │
             ▼                   ▼                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Cloud Run (Go API)                           │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Handler層                                                 │   │
│  │  ├── AuthHandler (Firebase Auth トークン検証)             │   │
│  │  ├── AnalyzeHandler (食事分析)                            │   │
│  │  ├── WeightHandler (体重管理)                             │   │
│  │  ├── ConditionHandler (体調記録)                          │   │
│  │  └── TrainingHandler (トレーニング)                       │   │
│  ├──────────────────────────────────────────────────────────┤   │
│  │ Service層                                                 │   │
│  │  └── GeminiService (Gemini API連携)                       │   │
│  ├──────────────────────────────────────────────────────────┤   │
│  │ Repository層 (Firestore Client)                           │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────┬───────────────────┬───────────────────┬──────────────┘
           │                   │                   │
           ▼                   ▼                   ▼
    ┌────────────┐      ┌────────────┐      ┌────────────┐
    │ Firestore  │      │ Gemini API │      │Cloud Storage│
    │ (データ)   │      │ (AI分析)   │      │ (画像)     │
    └────────────┘      └────────────┘      └────────────┘
```

### Firestoreコレクション設計

```
users/
  └── {userId}/
        ├── email: string
        ├── name: string
        ├── createdAt: timestamp
        ├── updatedAt: timestamp
        │
        ├── weightGoal/  (サブコレクション、1ドキュメント)
        │     └── goal/
        │           ├── targetWeight: number
        │           └── targetDate: timestamp
        │
        ├── weightRecords/  (サブコレクション)
        │     └── {recordId}/
        │           ├── weight: number
        │           └── recordedAt: timestamp
        │
        ├── conditionRecords/  (サブコレクション)
        │     └── {recordId}/
        │           ├── condition: number (1-3)
        │           ├── fatigue: number (1-3)
        │           └── recordedAt: timestamp
        │
        ├── trainingLocations/  (サブコレクション)
        │     └── {locationId}/
        │           ├── name: string
        │           ├── sortOrder: number
        │           │
        │           └── equipment/  (サブコレクション)
        │                 └── {equipmentId}/
        │                       ├── name: string
        │                       └── sortOrder: number
        │
        ├── trainingRecords/  (サブコレクション)
        │     └── {recordId}/
        │           ├── locationId: string
        │           ├── recordedAt: timestamp
        │           └── completed: boolean
        │
        ├── analysisRequests/  (サブコレクション)
        │     └── {requestId}/
        │           ├── status: string
        │           ├── imagePath: string
        │           ├── mealType: string
        │           ├── mealDate: timestamp
        │           ├── result: map (foods, totalCalories, etc.)
        │           └── createdAt: timestamp
        │
        └── mylist/  (サブコレクション)
              └── {itemId}/
                    ├── name: string
                    ├── foods: array
                    └── totalCalories: number
```

### 認証フロー

```
1. iOS: Firebase Auth でサインイン（Email/Password or Apple Sign-In）
2. iOS: Firebase ID Token を取得
3. iOS: Authorization: Bearer {idToken} でAPI呼び出し
4. Cloud Run: Firebase Admin SDK でトークン検証
5. Cloud Run: トークンからUID取得 → Firestoreクエリに使用
```

### Firestore Security Rules

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // ユーザーは自分のデータのみ読み書き可能
    match /users/{userId}/{document=**} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }
  }
}
```

### コスト見積もり

**想定使用量（個人利用）:**
- API呼び出し: 50リクエスト/日
- Firestore読み取り: 200回/日
- Firestore書き込み: 50回/日
- 画像ストレージ: 100MB/月
- Gemini API: 10リクエスト/日

| 項目 | 想定使用量 | 無料枠 | 超過時単価 |
|:---|:---|:---|:---|
| Cloud Run | 1,500リクエスト/月 | 200万リクエスト/月 | $0.00002400/リクエスト |
| Firestore読み取り | 6,000回/月 | 150万回/月 | $0.036/10万回 |
| Firestore書き込み | 1,500回/月 | 60万回/月 | $0.108/10万回 |
| Cloud Storage | 100MB | 5GB | $0.020/GB/月 |
| Gemini API | 300リクエスト/月 | 1分15リクエスト | 無料枠内で運用 |

**想定月額コスト: $0**（すべて無料枠内）

## 移行計画

### フェーズ1: GCPプロジェクト準備

1. GCPプロジェクト作成
2. Firebase プロジェクト連携
3. 必要なAPIの有効化（Cloud Run, Firestore, Cloud Storage, Gemini API）
4. サービスアカウント設定

### フェーズ2: バックエンド移行

1. Firestore Repositoryの実装
2. Gemini API Serviceの実装
3. Firebase Auth ミドルウェアの実装
4. Cloud Storage連携の実装
5. Cloud Runへのデプロイ

### フェーズ2.5: データ移行

1. PostgreSQL → Firestoreのマイグレーションスクリプト作成
2. 既存ユーザーデータの移行
3. データ整合性の検証
4. ロールバック手順の準備

### フェーズ3: iOSアプリ対応

1. Firebase Auth SDK導入
2. API Base URL切り替え
3. 画像アップロードをCloud Storage経由に変更

### フェーズ4: 廃止作業

1. Next.jsフロントエンドの削除
2. exe.dev VMの停止
3. PostgreSQLデータのバックアップ・廃止

### ロールバック戦略

移行中に重大な問題が発生した場合：

1. iOS APIのBase URLを旧環境（exe.dev）に切り替え
2. exe.dev VMは並行稼働期間中（1週間）は維持
3. Firestoreのデータは一時保持
4. 問題解決後に再度移行を実施

**並行稼働期間**: フェーズ3完了後、1週間は両環境を維持

## 理由

### Cloud Runを選定した理由

- Goに最適（コンテナベース、コールドスタートが速い）
- 無料枠が充実（月200万リクエスト）
- GCPサービスとの統合が容易
- スケールダウンでコスト最適化

### Firestoreを選定した理由

- サーバレスで管理不要
- 無料枠が個人利用に十分
- Firebase Authと自然に統合
- iOS SDKが充実（オフライン対応も可能）
- NoSQLだが、ウチコミのデータモデルに適合
  - ユーザーごとのデータが独立
  - 複雑なJOINが不要
  - 日付ベースのシンプルなクエリ

**Firestoreの注意点:**
- 複雑なクエリ（複数フィールドでのOR条件）に制限がある
- 複合インデックスの事前定義が必要
- PostgreSQLからのデータ移行にスクリプトが必要

### Cloud Runのコールドスタート対策

- 最小インスタンス数: 0（コスト最適化のため）
- 想定レイテンシ: 初回リクエストで数百ms〜1秒程度の遅延
- 対策: iOS側でローディングUIを適切に表示

### Gemini APIのレート制限対策

- 無料枠: 1分あたり15リクエスト
- 対策:
  - 画像分析リクエストにレート制限ミドルウェアを実装
  - 429エラー時のリトライ戦略（Exponential Backoff）
  - iOSアプリ側でのローディングUIとエラーハンドリング

### Firebase Authを選定した理由

- Apple Sign-In対応（iOSに必須）
- セキュリティルールでFirestoreと統合
- 認証実装の工数削減
- 無料枠で十分（月5万MAU）

### Next.jsを廃止する理由

- iOSアプリが主要なクライアントになる
- Web版の利用頻度が低い
- 運用コスト・複雑性の削減

## 結果

### 導入が必要なもの

**GCPサービス:**
- Cloud Run
- Firestore
- Cloud Storage
- Gemini API

**Firebaseサービス:**
- Firebase Authentication
- Firebase Admin SDK (Go)

**iOSライブラリ:**
- Firebase iOS SDK (Auth)

### 削除対象

- `frontend/` ディレクトリ全体
- exe.dev VM
- PostgreSQL関連のマイグレーション・コード

### 移行の影響

| 影響 | 対応 |
|:---|:---|
| PostgreSQL → Firestore | Repository層の書き換え |
| Gemini CLI → API | Service層の書き換え |
| 自前JWT → Firebase Auth | ミドルウェア・iOS側の変更 |
| ローカル画像 → Cloud Storage | アップロード処理の変更 |

### テスト戦略への影響

- `frontend/e2e/` のPlaywrightテストは廃止
- iOSアプリのUIテスト（XCUITest）がE2Eテストの役割を担う
- ADR-001で定義したiOSテスト戦略が主要なテスト戦略となる
- バックエンドAPIの統合テストは維持（Firestoreエミュレータ使用）

### 関連ドキュメント

- [全体アーキテクチャ](../CODEMAPS/architecture.md)（移行後に更新）
- [バックエンド構造](../CODEMAPS/backend.md)（移行後に更新）
