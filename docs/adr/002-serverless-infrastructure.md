# ADR-002: サーバレスインフラ設計

## コンテキスト

iOSアプリからAPIへのアクセスが主要ユースケースとなるため、サーバレス構成で低コスト運用を実現する。

## 決定

### インフラ構成

| コンポーネント | 技術 |
|:---|:---|
| クライアント | iOS (SwiftUI) |
| バックエンドAPI | Cloud Run (Go) |
| データベース | Firestore |
| AI | Gemini API (gemini-2.0-flash) |
| 画像ストレージ | Cloud Storage |
| 認証 | Firebase Authentication |

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

#### GCP各サービスの料金詳細

Cloud Run:

| 項目 | 無料枠 | 超過時単価 |
|:---|:---|:---|
| CPU | 180,000 vCPU-seconds/月 | $0.00002400/vCPU-second |
| メモリ | 360,000 GiB-seconds/月 | $0.00000250/GiB-second |
| リクエスト | 200万リクエスト/月 | $0.40/100万リクエスト |

Firestore:

| 項目 | 無料枠 | 超過時単価 |
|:---|:---|:---|
| ドキュメント読み取り | 50,000回/日 | $0.036/10万回 |
| ドキュメント書き込み | 20,000回/日 | $0.108/10万回 |
| ストレージ | 1GB | $0.108/GB/月 |

Cloud Storage:

| 項目 | 無料枠 | 超過時単価 |
|:---|:---|:---|
| ストレージ（Standard） | 5GB | $0.020/GB/月 |
| Class A操作 | 5,000回/月 | $0.05/1万回 |

Gemini API:

| モデル | 無料枠 | 有料（入力） | 有料（出力） |
|:---|:---|:---|:---|
| gemini-2.0-flash | 10 RPM, 250 req/日 | $0.10/100万トークン | $0.40/100万トークン |

Firebase Authentication:

| 項目 | 無料枠 |
|:---|:---|
| Email/Password認証 | 50,000 MAU |
| Apple Sign-In | 50,000 MAU |

#### 使用量シナリオ別見積もり

シナリオ1: 最小利用（週2-3回の記録）

| 項目 | 月間使用量 | 無料枠 | コスト |
|:---|:---|:---|:---|
| Cloud Run | 300回 | 200万回 | $0 |
| Firestore読み取り | 3,000回 | 150万回/月 | $0 |
| Firestore書き込み | 500回 | 60万回/月 | $0 |
| Cloud Storage | 10MB | 5GB | $0 |
| Gemini API | 30回 | 250回/日 | $0 |

月額: $0

シナリオ2: 通常利用（毎日3食 + 体重 + 体調）

| 項目 | 月間使用量 | 無料枠 | コスト |
|:---|:---|:---|:---|
| Cloud Run | 3,000回 | 200万回 | $0 |
| Firestore読み取り | 30,000回 | 150万回/月 | $0 |
| Firestore書き込み | 6,000回 | 60万回/月 | $0 |
| Cloud Storage | 200MB | 5GB | $0 |
| Gemini API | 90回 | 250回/日 | $0 |

月額: $0

シナリオ3: ヘビー利用（毎日フル活用 + 頻繁な閲覧）

| 項目 | 月間使用量 | 無料枠 | コスト |
|:---|:---|:---|:---|
| Cloud Run | 10,000回 | 200万回 | $0 |
| Firestore読み取り | 100,000回 | 150万回/月 | $0 |
| Firestore書き込み | 15,000回 | 60万回/月 | $0 |
| Cloud Storage | 1GB | 5GB | $0 |
| Gemini API | 200回 | 250回/日 | $0 |

月額: $0

#### 無料枠を超える可能性があるケース

Gemini APIのレート制限:
- 1分あたり10リクエストの制限
- 対策: リクエスト間に適切な間隔、429エラー時のExponential Backoff

Cloud Storageの長期蓄積:
- 1年後: 約550MB（無料枠内）
- 3年後: 約1.6GB（無料枠内）
- 10年後: 約5.5GB（超過時 $0.01/月）

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
- ウチコミのデータモデルに適合（ユーザーごとのデータが独立、複雑なJOINが不要）

注意点:
- 複雑なクエリ（複数フィールドでのOR条件）に制限がある
- 複合インデックスの事前定義が必要

### Cloud Runのコールドスタート対策

- 最小インスタンス数: 0（コスト最適化のため）
- 想定レイテンシ: 初回リクエストで数百ms〜1秒程度の遅延
- 対策: iOS側でローディングUIを適切に表示

### Gemini APIのレート制限対策

- 無料枠: 1分あたり10リクエスト、1日250リクエスト
- 対策:
  - 画像分析リクエストにレート制限ミドルウェアを実装
  - 429エラー時のリトライ戦略（Exponential Backoff）
  - iOSアプリ側でのローディングUIとエラーハンドリング

### Firebase Authを選定した理由

- Apple Sign-In対応（iOSに必須）
- セキュリティルールでFirestoreと統合
- 認証実装の工数削減
- 無料枠で十分（月5万MAU）

## 結果

### 導入が必要なもの

GCPサービス:
- Cloud Run
- Firestore
- Cloud Storage
- Gemini API

Firebaseサービス:
- Firebase Authentication
- Firebase Admin SDK (Go)

iOSライブラリ:
- Firebase iOS SDK (Auth)

### テスト戦略

- iOSアプリのUIテスト（XCUITest）がE2Eテストの役割を担う
- ADR-001で定義したiOSテスト戦略が主要なテスト戦略となる
- バックエンドAPIの統合テストは維持（Firestoreエミュレータ使用）

### 関連ドキュメント

- [全体アーキテクチャ](../CODEMAPS/architecture.md)
- [バックエンド構造](../CODEMAPS/backend.md)
