# EDG-700: iOSのE2Eテストを除却する

## Workflow Status
- [x] Phase 1: 準備
- [x] Phase 2: 計画
- [ ] Phase 3: 実装
- [ ] Phase 4: レビューサイクル
- [ ] Phase 5: PR 作成
- [ ] Phase 6: ユーザー確認

**Current Phase**: Phase 3
**Branch**: edg-700
**Issue**: EDG-700

## Post-Implementation Instructions

### Phase 4: レビューサイクル
Task ツールで以下3エージェントを起動し、CRITICAL/IMPORTANT=0 まで繰り返す:
- pr-review-toolkit:code-reviewer
- pr-review-toolkit:pr-test-analyzer
- pr-review-toolkit:silent-failure-hunter

### Phase 5: PR 作成
gh pr create、Body に Fixes EDG-700 を含める

### Phase 6: ユーザー確認
git diff main...HEAD | difit → ユーザーにレビュー依頼
完了後、このプランファイルを削除する

---

## Linear Issue
- Issue: EDG-700
- URL: https://linear.app/ryosuke-horie/issue/EDG-700/iosのe2eテストを除却する

## 概要
OSバージョンの問題で XCUITest（UIテスト）がエラーが発生しやすく、修正困難なため、E2Eテスト関連のドキュメント記述を全て削除する。
また、macOS バージョン問題が解消するまで、既存の iOS テストコード（ユニットテスト、スナップショットテストも含めて）を全て `Disabled/` へ移動する。

## 現状確認

### 実装状況
- UITests ターゲット: 存在しない
  - `ios/project.yml` には UchikomiTests（ユニットテスト）のみ
  - `ios/UchikomiUITests/` ディレクトリは存在しない
- CI/CD: iOS テストジョブは存在しない
  - `.github/workflows/ci.yml` には ios-lint ジョブのみ

### 既存テストファイル
- `ios/UchikomiTests/` 配下に複数のテストファイルが存在
- 一部は既に `Disabled/` へ移動済み

## 実装計画

### ステップ1: 全ての iOS テストファイルを `Disabled/` へ移動

```bash
# 既存のテストファイルを Disabled/ へ移動
cd ios/UchikomiTests
mkdir -p Disabled
mv *.swift Disabled/ 2>/dev/null || true
# Disabled/ 内のファイルは既に移動済みなので除外
mv Features/* Disabled/ 2>/dev/null || true
```

### ステップ2: `.claude/rules/ios-testing.md` の更新

1. XCUITest 関連記述を削除:
   - 14行目: テストフレームワーク表の「XCUITest（UI）」
   - 25行目: View のカバレッジ優先度
   - 95-97行目: ファイル配置図の `UchikomiUITests/`
   - 109-131行目: 「UIテスト（XCUITest）」セクション全体
   - 132行目: 禁止事項の「UIテストでのハードコードされた座標」

2. macOS バージョン問題に関する記述を追加:
   - ファイルの先頭に「現状のテスト方針」セクションを追加
   - macOS バージョン問題により、テストを一時的に無効化している経緯を記載

### ステップ3: `docs/adr/001-ios-unit-testing.md` の更新

1. XCUITest 関連記述を削除:
   - 9行目: コンテキストの「UIテスト（E2E）の導入」
   - 20行目: テストフレームワーク表の「UIテスト XCUITest」
   - 34行目: テスト実行タイミング表の「UIテスト（XCUITest）」行
   - 37行目: 「UIテストはClaude Codeスキルを使用して...」
   - 88-92行目: 「UIテストをCIで実行しない理由」セクション全体
   - 106行目: 導入が必要なものの「Claude Codeスキル: PR作成前UIテスト実行用」
   - 117行目: CI/CD設定の「UIテスト: CI実行なし（ローカルのみ）」
   - 122行目: ドキュメントの「.claude/skills/にPR作成前UIテスト実行スキルを追加」

2. macOS バージョン問題に関する追記を追加:
   - 結果セクションに「現状のテスト方針（一時停止）」を追加
   - macOS バージョン問題によりテストを無効化している経緯を記載

### ステップ4: `docs/adr/002-serverless-infrastructure.md` の更新

修正箇所:
- 287行目: 「iOSアプリのUIテスト（XCUITest）がE2Eテストの役割を担う」
  - 代替: 「バックエンドAPIのE2Eテスト（backend/e2e/）が統合テストを担う」

### ステップ5: `.claude/rules/testing-tdd.md` の更新

削除する記述:
- 68行目: 「XCUITest: UIテスト」

### ステップ6: `.claude/skills/tdd-workflow/SKILL.md` の更新

削除/修正する記述:
- 3行目: description の「ユニット、統合、E2Eテストを含む」→「ユニット、統合テストを含む」
- 25行目: カバレッジ要件の「（ユニット + 統合 + E2E）」→「（ユニット + 統合）」
- 42-44行目: E2Eテストセクション全体

## 検証方法

```bash
# 1. XCUITest関連の文字列が残っていないことを確認
grep -r "XCUITest\|UITest\|UitTests" --include="*.md" . --exclude-dir=node_modules

# 2. 全てのテストファイルが Disabled/ にあることを確認
ls ios/UchikomiTests/Disabled/

# 3. テスト実行（0 テスト成功になるはず）
task ios:test
```

## 新規ドキュメント: iOS テスト方針（一時停止）

`.claude/rules/ios-testing-policy.md` を新規作成:

```markdown
# iOS テスト方針（一時停止）

## 背景

macOS および Xcode のバージョンアップにより、iOS テストが頻繁に壊れる問題が発生している。
この問題が解消するまで、iOS テストを一時的に無効化する。

## 現状の方針

1. 全ての iOS テストコード（ユニットテスト、スナップショットテスト）を `Disabled/` ディレクトリに移動
2. CI では iOS テストを実行しない（現状維持）
3. 開発中は手動テストを中心に行う

## 再開の条件

以下のいずれかの条件が満たされた場合、テストの再開を検討する:

- macOS/Xcode のバージョン問題が安定し、テストが安定して実行できるようになった
- CI で macOS ランナーを利用可能になった（コスト面での制約が解消された）
- 代替の安定したテスト方法が見つかった

## 関連 Issue

- EDG-700: iOSのE2Eテストを除却する
```

## 変更しないもの

- `.github/workflows/ci.yml` - 現状維持（ios-lint ジョブのみ）
- `ios/project.yml` - UchikomiTests ターゲットは維持（将来的な再開のため）
