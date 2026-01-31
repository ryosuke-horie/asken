# 依存関係管理

## dependabot設定

dependabotは`.github/dependabot.yml`で設定されており、週次（月曜）で依存関係の更新PRを作成する。

### グループ化設定

ライブラリは論理的なグループにまとめてPRを作成する設定になっている。

**npm (frontend)**
| グループ | パッケージ |
|:---|:---|
| react-ecosystem | react, react-dom, next, swr, eslint-config-next |
| testing | vitest, @testing-library/*, @playwright/test, jsdom |
| linting | eslint, prettier, knip, depcheck |
| types | typescript, @types/* |

**gomod (backend)**
| グループ | パッケージ |
|:---|:---|
| testing | testify, sqlmock |
| security | jwt, crypto |
| database | pq |

**swift (iOS)**
| グループ | パッケージ |
|:---|:---|
| testing | swift-snapshot-testing, swift-custom-dump, xctest-dynamic-overlay |
| swift-ecosystem | swift-syntax |

### ライブラリ追加時の対応

新しいライブラリを追加した場合、以下を確認すること：

1. **既存グループに該当するか確認**
   - 該当するグループがあれば、そのグループのパターンに追加
   - 例: 新しいテストユーティリティ → `testing`グループに追加

2. **新規グループが必要か検討**
   - 既存グループに該当しない場合は新規グループの作成を検討
   - グループ名は用途が明確にわかる名前にする

3. **`.github/dependabot.yml`を更新**
   - `groups`セクションの該当グループにパターンを追加
   - このドキュメントの表も同時に更新する
