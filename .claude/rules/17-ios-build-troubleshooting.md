# iOSビルドトラブルシューティング

## SPMキャッシュの完全クリア

SPMキャッシュ関連の問題が発生した場合、以下の手順で完全にクリアする。

### 手順

```bash
# 1. Xcodeを完全に終了
killall Xcode 2>/dev/null || true

# 2. DerivedData削除
rm -rf ~/Library/Developer/Xcode/DerivedData

# 3. SPMキャッシュ削除
rm -rf ~/Library/Caches/org.swift.swiftpm

# 4. Xcodeのパッケージキャッシュ削除
rm -rf ~/Library/Developer/Xcode/DerivedData/*/SourcePackages

# 5. プロジェクトローカルのSPMキャッシュ削除
rm -rf ios/.swiftpm
rm -rf ios/.build

# 6. Package.resolved削除（再解決を強制）
rm -f ios/Uchikomi.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved

# 7. Xcodeを起動してパッケージ解決
open ios/Uchikomi.xcodeproj
```

### Xcodeでの追加操作

Xcodeが起動したら:
1. **File > Packages > Reset Package Caches**
2. **File > Packages > Resolve Package Versions**
3. **Product > Clean Build Folder** (Shift+Cmd+K)

### Taskfileコマンド

```bash
# 完全クリア
task ios:clean-all

# 通常のクリーン（DerivedDataのみ）
task ios:clean
```

## Firebase SDK関連の問題

### 現状

Firebase SDK 11.15.0は21個の推移的依存関係を持つ:
- abseil-cpp-binary, app-check, appauth-ios, grpc-binary, nanopb など

### ベストプラクティス

| 設定 | 推奨 | 理由 |
|:---|:---|:---|
| バージョン指定 | `exactVersion` | バージョン範囲だと解決が不安定 |
| Package.resolved | gitにコミット | チーム間で同じ依存関係を使用 |
| 更新頻度 | 月1回程度 | 頻繁な更新は避ける |

### Firebase SDK更新時の注意

1. ローカルで動作確認してからコミット
2. Package.resolvedの変更も一緒にコミット
3. 問題発生時は`git checkout`で戻す

```bash
# Firebase更新で問題が出た場合
git checkout ios/Uchikomi.xcodeproj/project.xcworkspace/xcshareddata/swiftpm/Package.resolved
task ios:clean-all
```

## XcodeGen関連の問題

### プロジェクト再生成

`project.yml`を変更した場合:

```bash
cd ios
xcodegen generate
```

### 注意点

- XcodeGen実行後、SPMの参照がリセットされる場合がある
- 必要に応じて`File > Packages > Resolve Package Versions`を実行

## よくあるエラーと対処

### "Package resolution failed"

```bash
# 完全クリアして再解決
task ios:clean-all
```

### "The package manifest at '...' cannot be accessed"

ネットワーク問題の可能性:
1. VPNを切断して再試行
2. `~/.netrc`の認証情報を確認（プライベートリポジトリの場合）

### "Missing package product '...'"

XcodeGenとSPMの不整合:
```bash
cd ios
xcodegen generate
# Xcodeを再起動して Resolve Package Versions
```

### "multiple targets named '...'"

DerivedDataの破損:
```bash
rm -rf ~/Library/Developer/Xcode/DerivedData
```

## CI/CDでのiOSビルド

### 現状

- GitHub ActionsでmacOSランナーは10倍のコスト
- LinuxでiOSアプリのビルドは不可能（Xcode/iOS SDKが必要）

### 代替案

| 方法 | コスト | カバー範囲 |
|:---|:---|:---|
| Linuxでlint/format | 低 | 構文チェックのみ |
| mainマージ時のみmacOS | 中 | ビルド検証 |
| self-hosted Mac mini | 初期のみ | 完全なビルド/テスト |

### 現在の設定

- PR時: SwiftLint, SwiftFormat（Linux）
- ビルド検証: ローカルのみ
