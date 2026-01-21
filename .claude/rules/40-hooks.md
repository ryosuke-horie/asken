# フックシステム

## フックの種類

| フック | タイミング | 用途 |
| :--- | :--- | :--- |
| PreToolUse | ツール実行前 | 検証、パラメータ修正 |
| PostToolUse | ツール実行後 | 自動フォーマット、チェック |
| Stop | セッション終了時 | 最終検証 |

## 現在のフック設定（.claude/hooks/hooks.json）

### PreToolUse

- tmux推奨: 長時間実行コマンド（npm install等）にtmux使用を推奨
- ドキュメントブロッカー: 不要な.md/.txtファイルの作成をブロック

### PostToolUse

- Prettier: JS/TSファイル編集後に自動フォーマット
- TypeScriptチェック: .ts/.tsxファイル編集後にtscを実行
- console.log警告: 編集ファイル内のconsole.logを警告

### Stop

- console.log監査: セッション終了前に変更ファイル内のconsole.logを最終チェック

## TodoWriteのベストプラクティス

TodoWriteツールを使用して:

- マルチステップタスクの進捗を追跡
- 指示の理解を確認
- リアルタイムでの方向修正を可能に
- 詳細な実装ステップを表示

Todoリストで明らかになる問題:

- 順序が間違っているステップ
- 欠けている項目
- 不要な項目
- 粒度の誤り
- 要件の誤解
