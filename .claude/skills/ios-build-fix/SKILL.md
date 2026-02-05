---
name: ios-build-fix
description: iOSビルドエラーを診断し、SPMキャッシュの問題を解決。Firebase SDK関連の依存関係問題に対応。
model: sonnet
allowed-tools: Bash, Read, Glob
---

# iOS ビルド修正スキル

SPMキャッシュの問題やFirebase SDK関連のビルドエラーを診断・解決します。

## 診断手順

### 1. 症状の確認

ユーザーに以下を確認:
- エラーメッセージの内容
- 最後に成功したビルドからの変更点
- 最近Package.resolvedが変更されたか

### 2. 段階的な対処

#### レベル1: DerivedDataのみ削除

```bash
rm -rf ~/Library/Developer/Xcode/DerivedData
```

Xcodeで Product > Clean Build Folder (Shift+Cmd+K) を実行。

#### レベル2: SPMキャッシュ削除

```bash
rm -rf ~/Library/Caches/org.swift.swiftpm
rm -rf ios/.swiftpm
rm -rf ios/.build
```

Xcodeで File > Packages > Reset Package Caches を実行。

#### レベル3: Package.resolved再生成

```bash
rm -f ios/Uchikomi.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
```

Xcodeで File > Packages > Resolve Package Versions を実行。

#### レベル4: 完全クリア（最終手段）

```bash
task ios:clean-all
```

これは以下を実行:
1. Xcodeを終了
2. DerivedData削除
3. SPMキャッシュ削除
4. プロジェクトローカルキャッシュ削除
5. Package.resolved削除

### 3. Firebase SDK特有の問題

Firebase SDKは21個の推移的依存関係を持つため、問題が発生しやすい。

#### 依存関係の競合

Package.resolvedの特定パッケージのバージョンが不整合の場合:

```bash
# 既知の良いPackage.resolvedに戻す
git checkout ios/Uchikomi.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
task ios:clean-all
```

#### Firebase SDKアップデート後の問題

```bash
# project.ymlでexactVersionを確認
cat ios/project.yml | grep -A2 Firebase

# exactVersionを使用していることを確認
# from: ではなく exactVersion: を使用する
```

## Taskfileコマンド

| コマンド | 用途 |
|:---|:---|
| `task ios:clean` | DerivedDataのみ削除 |
| `task ios:clean-all` | 完全クリア（SPM含む） |
| `task ios:reset-packages` | Package.resolvedのみ削除 |

## よくあるエラーパターン

### "Package resolution failed"

→ `task ios:clean-all` を実行

### "The package manifest at '...' cannot be accessed"

→ ネットワーク問題。VPN切断して再試行。

### "Missing package product '...'"

→ XcodeGenとSPMの不整合。`cd ios && xcodegen generate` 後に Resolve Package Versions。

### "multiple targets named '...'"

→ DerivedDataの破損。`rm -rf ~/Library/Developer/Xcode/DerivedData`

## 結果

- 成功: ビルドが通ることを確認
- 失敗: 詳細なエラーログを収集し、追加調査
