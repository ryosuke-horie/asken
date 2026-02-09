---
description: iOSアプリのテストガイドライン。Swift Testing、Mockolo、swift-snapshot-testingを使用。
globs: ios/**/*.swift
---

# iOSテストガイドライン

> **現状のテスト方針（一時停止中）**
>
> macOS および Xcode のバージョンアップにより、iOS テストが頻繁に壊れる問題が発生しています。
> この問題が解消するまで、iOS テストを一時的に無効化しています。
> 全てのテストコードは `ios/UchikomiTests/Disabled/` ディレクトリに移動されています。
>
> 詳細は `.claude/rules/ios-testing-policy.md` を参照してください。

---

## 基本方針

| 項目 | 方針 |
|:---|:---|
| テストスタイル | 古典派（Classicist） |
| Mock対象 | 外部依存（API、Keychain）のみ |
| テストフレームワーク | Swift Testing |
| Mockライブラリ | Mockolo（Uber製） |
| スナップショット | swift-snapshot-testing |

## テスト対象と優先度

| レイヤー | 優先度 | カバレッジ目標 |
|:---|:---|:---|
| ViewModel | 高 | 80%以上 |
| Repository | 高 | 70%以上 |
| Model（ロジックあり） | 中 | 80%以上 |
| 全体 | - | 60%以上 |

## テストタイミング

| 場面 | アプローチ |
|:---|:---|
| ViewModel / ビジネスロジック | TDD推奨 |
| Repository / API連携 | TDD推奨 |
| バグ修正 | テスト先行必須 |
| 新規UI画面 | テスト後追いOK |
| ユーティリティ関数 | TDD推奨 |

必須ルール:
- PR作成前にテストが存在すること
- バグ修正は必ずテストから書く

## 命名規則

日本語「〜すべき」表現を使用:

```swift
@Suite struct AuthManagerTests {
    @Test func ログイン成功時に認証状態がtrueになるべき() async throws {
        // テスト実装
    }

    @Test func 無効なパスワードでログイン失敗すべき() async throws {
        // テスト実装
    }
}
```

## テスト構造

Arrange-Act-Assert パターン:

```swift
@Test func ログイン成功時に認証状態がtrueになるべき() async throws {
    // Arrange（準備）
    let mockRepo = MockAuthRepositoryProtocol()
    mockRepo.loginHandler = { email, password in
        return AuthResponse(token: "token", user: testUser)
    }
    let manager = AuthManager(repository: mockRepo)

    // Act（実行）
    try await manager.login(email: "test@example.com", password: "Pass0123")

    // Assert（検証）
    #expect(manager.isAuthenticated == true)
    #expect(manager.currentUser?.email == "test@example.com")
}
```

## ファイル配置

```
ios/
├── Uchikomi/           # プロダクションコード
│   ├── Features/
│   │   └── Auth/
│   │       ├── AuthManager.swift
│   │       └── LoginViewModel.swift
│   └── Core/
│       └── Repositories/
│           └── AuthRepository.swift
└── UchikomiTests/      # テスト（現在は無効化）
    └── Disabled/       # テストコード一時移動先
        ├── AuthManagerTests.swift
        └── MealsViewModelTests.swift
```

## Mockの使用基準

| 対象 | Mock可否 | 理由 |
|:---|:---|:---|
| 外部API | 可 | ネットワーク依存を排除 |
| Keychain | 可 | デバイス依存を排除 |
| 現在時刻 | 可 | 再現性の確保 |
| 内部クラス | 不可 | 実装詳細への依存を避ける |
| ユーティリティ | 不可 | 実際の振る舞いを検証 |

## スナップショットテスト

実行タイミング:
- ローカル: 開発時に随時
- CI: 毎回実行（`record: .never`）

## 禁止事項

- 実装詳細に依存したテスト（privateメソッドの直接テスト等）
- Mockだらけで実際の振る舞いを検証していないテスト
- テスト間の依存関係（順序依存、共有状態）
- `sleep()` による固定待機（`waitForExistence` を使用）
