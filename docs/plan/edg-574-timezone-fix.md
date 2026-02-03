# プラン: 日付のタイムゾーンがUTC固定でローカル日付とズレる問題を修正

## Linear Issue
- Issue: EDG-574
- URL: https://linear.app/ryosuke-horie/issue/EDG-574

## 概要

現在、mealDateを`time.Parse`でUTC扱いしており、GetDailyMealsもUTC日次で検索しているため、ユーザーのローカル日付とズレる問題を修正する。

## 問題の詳細

| 処理 | iOS側 | バックエンド側 |
|:---|:---|:---|
| 日付送信 | デバイスのローカルタイムゾーン | - |
| 日付パース | - | UTC扱い |
| 日付範囲検索 | - | UTC 00:00〜24:00 |

例: JSTユーザーが「2026-02-03」の食事を取得しようとすると：
- iOSは`2026-02-03`を送信
- バックエンドは`2026-02-03 00:00:00 UTC ~ 2026-02-04 00:00:00 UTC`で検索
- これはJSTでは`2026-02-03 09:00:00 ~ 2026-02-04 09:00:00`の範囲
- ユーザーが期待する「2026-02-03」の朝食（例: 8:00 JST）が含まれない

## 実装方針

IANAタイムゾーン名（例: `Asia/Tokyo`）をクライアントから送信し、バックエンドでそのタイムゾーンを考慮して日付範囲を計算する。

```
GET /api/meals/daily?date=2026-02-03&tz=Asia/Tokyo

バックエンド処理:
  2026-02-03 00:00:00 JST = 2026-02-02 15:00:00 UTC (start)
  2026-02-04 00:00:00 JST = 2026-02-03 15:00:00 UTC (end)
```

## 実装計画

### Phase 1: バックエンド - タイムゾーンユーティリティ作成

新規ファイル: `backend/internal/util/timezone.go`

- `GetDayRangeInTimezone(dateStr, tz string) (start, end time.Time, err error)`
  - 指定タイムゾーンでの日付範囲をUTCで返す

テスト: `backend/internal/util/timezone_test.go`

### Phase 2: バックエンド - ハンドラー修正

変更ファイル: `backend/internal/handler/daily_meals_handler.go`

- `tz`クエリパラメータを受け取る
- `tz`未指定時はUTCとして処理（後方互換性）

変更ファイル: `backend/internal/handler/analyze_handler.go`

- `tz`フォームフィールドを受け取る

### Phase 3: バックエンド - リポジトリ修正

変更ファイル: `backend/internal/repository/analysis_repository_firestore.go`

- `GetDailyMeals`にタイムゾーンパラメータを追加
- 日付範囲計算をタイムゾーン対応に変更

### Phase 4: iOS - タイムゾーン送信

変更ファイル: `ios/Uchikomi/Core/Repositories/MealRepository.swift`

- `TimeZone.current.identifier`を取得
- APIリクエストに`tz`パラメータを追加

変更ファイル: `ios/Uchikomi/Core/Network/APIEndpoint.swift`

- `dailyMeals`エンドポイントに`tz`パラメータを追加

### Phase 5: テスト更新

- バックエンド: ハンドラーテスト、リポジトリテストを更新
- iOS: モック再生成（Mockolo）

## 修正対象ファイル

| ファイル | 変更内容 |
|:---|:---|
| `backend/internal/util/timezone.go` | 新規作成 |
| `backend/internal/util/timezone_test.go` | 新規作成 |
| `backend/internal/handler/daily_meals_handler.go` | tzパラメータ受け取り |
| `backend/internal/handler/analyze_handler.go` | tzパラメータ受け取り |
| `backend/internal/repository/analysis_repository_firestore.go` | タイムゾーン対応 |
| `ios/Uchikomi/Core/Repositories/MealRepository.swift` | tz送信 |
| `ios/Uchikomi/Core/Network/APIEndpoint.swift` | エンドポイント修正 |

## 後方互換性

- `tz`パラメータ未指定時はUTCとして処理
- 既存のFirestoreデータはUTCで保存されており、そのまま利用可能

## 検証方法

1. バックエンドテスト実行: `task test`
2. iOSテスト実行: `task ios:test`
3. Chrome DevTools MCPで動作確認:
   - 開発サーバー起動
   - iOSシミュレータまたはブラウザから日次食事取得APIを呼び出し
   - 正しい日付範囲で食事が取得されることを確認
