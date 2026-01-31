# iOSテストガイドライン

## 基本方針

| 項目 | 方針 |
|:---|:---|
| テストスタイル | 古典派（Classicist） |
| Mock対象 | 外部依存（API、Keychain）のみ |
| テストフレームワーク | Swift Testing（ユニット）、XCUITest（UI） |
| Mockライブラリ | Mockolo（Uber製） |
| スナップショット | swift-snapshot-testing |

## テスト対象と優先度

| レイヤー | 優先度 | カバレッジ目標 |
|:---|:---|:---|
| ViewModel | ✅ 高 | 80%以上 |
| Repository | ✅ 高 | 70%以上 |
| Model（ロジックあり） | ⚠️ 中 | 80%以上 |
| View（SwiftUI） | ❌ 低 | UIテスト/スナップショットでカバー |
| 全体 | - | 60%以上 |

## テストタイミング

| 場面 | アプローチ |
|:---|:---|
| ViewModel / ビジネスロジック | TDD推奨 |
| Repository / API連携 | TDD推奨 |
| バグ修正 | テスト先行必須 |
| 新規UI画面 | テスト後追いOK |
| ユーティリティ関数 | TDD推奨 |

**必須ルール:**
- PR作成前にテストが存在すること
- バグ修正は必ずテストから書く

## 命名規則

日本語「〜すべき」表現を使用（Frontendと統一）:

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
    given(mockRepo).login(email: .any, password: .any)
        .willReturn(testAuthResponse)
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
├── UchikomiTests/      # ユニットテスト
│   ├── AuthManagerTests.swift
│   └── MealsViewModelTests.swift
└── UchikomiUITests/    # UIテスト（XCUITest）
    └── LoginUITests.swift
```

## Mockの使用基準

| 対象 | Mock可否 | 理由 |
|:---|:---|:---|
| 外部API | ✅ 可 | ネットワーク依存を排除 |
| Keychain | ✅ 可 | デバイス依存を排除 |
| 現在時刻 | ✅ 可 | 再現性の確保 |
| 内部クラス | ❌ 不可 | 実装詳細への依存を避ける |
| ユーティリティ | ❌ 不可 | 実際の振る舞いを検証 |

## Mockolo使用方法

### 1. Protocolに`@mockable`を付与

```swift
/// @mockable
protocol AuthRepositoryProtocol {
    func login(email: String, password: String) async throws -> AuthResponse
}
```

### 2. Mock生成コマンド

```bash
mockolo -s Uchikomi -d UchikomiTests/Mocks/GeneratedMocks.swift
```

### 3. テストでの使用

```swift
import XCTest
@testable import Uchikomi

final class AuthManagerTests: XCTestCase {
    func testLoginSuccess() async throws {
        let mockRepo = MockAuthRepositoryProtocol()
        mockRepo.loginHandler = { email, password in
            return AuthResponse(token: "token", user: testUser)
        }

        let manager = AuthManager(repository: mockRepo)
        try await manager.login(email: "test@example.com", password: "Pass0123")

        XCTAssertTrue(manager.isAuthenticated)
    }
}
```

## UIテスト（XCUITest）

### 実行タイミング

- ローカル: PR作成前に実行（スキルで強制）
- CI: 実行しない（コスト削減）

### ベストプラクティス

```swift
import XCTest

final class LoginUITests: XCTestCase {
    let app = XCUIApplication()

    override func setUp() {
        continueAfterFailure = false
        app.launch()
    }

    func testLoginFlow() {
        // Accessibility Identifierを使用
        let emailField = app.textFields["emailTextField"]
        emailField.tap()
        emailField.typeText("test@example.com")

        let passwordField = app.secureTextFields["passwordTextField"]
        passwordField.tap()
        passwordField.typeText("Pass0123")

        app.buttons["loginButton"].tap()

        // waitForExistenceで非同期待機（sleepは禁止）
        let mealsView = app.otherElements["mealsView"]
        XCTAssertTrue(mealsView.waitForExistence(timeout: 5))
    }
}
```

## スナップショットテスト

### 実行タイミング

- ローカル: 開発時に随時
- CI: 毎回実行（`record: .never`）

### 使用方法

```swift
import SnapshotTesting
import SwiftUI
import XCTest
@testable import Uchikomi

final class NutritionCardSnapshotTests: XCTestCase {
    func testNutritionSummaryCard() {
        let view = NutritionSummaryCard(
            totalCalories: 1500,
            totalProtein: 60,
            totalFat: 50,
            totalCarbohydrates: 180
        )

        assertSnapshot(of: view, as: .image(layout: .device(config: .iPhone15)))
    }
}
```

### CI設定

```swift
// CI環境では recording: .never で実行
withSnapshotTesting(record: .never) {
    assertSnapshot(of: view, as: .image)
}
```

## 禁止事項

- 実装詳細に依存したテスト（privateメソッドの直接テスト等）
- Mockだらけで実際の振る舞いを検証していないテスト
- テスト間の依存関係（順序依存、共有状態）
- `sleep()` による固定待機（`waitForExistence` を使用）
- UIテストでのハードコードされた座標

## テストコマンド

```bash
# ユニットテスト実行
xcodebuild test -scheme Uchikomi -destination 'platform=iOS Simulator,name=iPhone 15'

# UIテスト実行
xcodebuild test -scheme UchikomiUITests -destination 'platform=iOS Simulator,name=iPhone 15'

# カバレッジ付きテスト
xcodebuild test -scheme Uchikomi -enableCodeCoverage YES -destination 'platform=iOS Simulator,name=iPhone 15'
```
