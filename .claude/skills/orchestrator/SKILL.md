---
name: orchestrator
description: LinearイシューURLから実装完了までの全工程を自動実行。ブランチ作成、Plan、実装、レビュー、PR作成まで一貫して行う。
---

# Linear Issue 実装オーケストレーター

LinearイシューのURLを受け取り、ブランチ作成からPR作成・レビュー依頼までを一貫して実行するスキル。

## 引数

- Linear Issue URL（必須）: 第1引数としてLinear IssueのURLを渡す
  - 例: `/orchestrator https://linear.app/ryosuke-horie/issue/EDG-687/...`

---

## Agents Team 構成

計画・実装フェーズでは、以下のチーム体制でロールプレイを行い、多角的な視点で議論・意思決定する。
各メンバーの発言はプレフィックス付きで明示する。

| ロール | 責務 |
|:---|:---|
| PdM（リーダー） | ユーザーとの窓口。要件・仕様の責任者。プラン策定の最終判断。不明点や解釈の余地がある指示はユーザーに質問する |
| Tech Lead | 技術全体の取りまとめ。アーキテクチャ判断、実装方針の決定。各エンジニアの意見を統合する |
| Backend Engineer | Go/API/Firestore領域の専門家。API設計、データモデル、パフォーマンスの観点から意見を出す |
| iOS Engineer | Swift/SwiftUI領域の専門家。UI実装、ViewModel設計、テストの観点から意見を出す |
| Security Engineer | セキュリティ観点の専門家。認証、入力検証、脆弱性リスクの観点から意見を出す |

### チーム運営ルール

1. PdMがユーザーの要件を整理し、チームに共有する
2. Tech Leadが技術的な実現方針を提示する
3. 各エンジニアが自分の専門領域から意見・懸念・提案を出す
4. 疑問や違和感、解釈の余地がある要件はPdMがユーザーに `AskUserQuestion` で質問する
   - ぶらさずに明確にすること。曖昧なまま実装に入らない
   - この質問は計画フェーズだけでなく、実装中やレビューフェーズでも行ってよい
5. チーム内の議論結果をプランファイルに反映する

### 発言フォーマット

議論の出力は以下のフォーマットで行う:

```
[PdM] 要件を整理します。今回のイシューは...
[Tech Lead] 技術的には2つのアプローチが考えられます...
[Backend] API側の懸念として、Firestoreのクエリ制約があります...
[iOS] ViewModel側では既存のパターンに合わせて...
[Security] 入力バリデーションについて確認したい点があります...
[PdM → ユーザー] 以下の点を確認させてください: ...
```

---

## ワークフロー

### Phase 1: 準備

1. Linear Issue URLからイシュー情報を取得する
   - `mcp__plugin_linear_linear__get_issue` でイシュー詳細を取得
   - イシュー番号（例: EDG-687）を抽出
2. Linear課題のステータスを「In Progress」に変更する
   - `mcp__plugin_linear_linear__update_issue` で `state: "In Progress"` に更新
   - 注意: 「Done」には変更しない（PRマージ時に自動更新される）
3. mainブランチから最新を取得し、フィーチャーブランチを作成する
   ```bash
   git checkout main
   git pull origin main
   git checkout -b edg-{番号}
   ```

### Phase 2: 計画（Plan Mode + Agents Team 議論）

1. `EnterPlanMode` を使用してPlan Modeに入る
2. PdMがイシュー要件を整理し、チームに共有する
3. チーム全体でコードベースを調査し議論する
   - Tech Lead: アーキテクチャ、既存パターンとの整合性
   - Backend: API設計、DB設計、パフォーマンス
   - iOS: UI/UX、ViewModel設計、テスト方針
   - Security: 認証・認可、入力検証、リスク評価
4. 議論中に疑問・懸念が出たら、PdMがユーザーに質問する
   - `AskUserQuestion` を使用
   - 複数の選択肢がある場合はチームの推奨案を提示する
5. プランファイルを `docs/plan/` に作成する
   - ファイル名: docs/plan/に自動生成されるファイル名のまま
   - 内容: Linear Issue情報、概要、チーム議論の要約、実装計画、技術的考慮事項、テスト計画
6. `ExitPlanMode` でユーザーの承認を待つ
7. ユーザーがプランを承認したら Phase 3 に進む

### Phase 3: 実装（Agents Team 主導）

1. Tech Leadが実装順序を決定し、タスクを各エンジニアに割り振る形で進める
2. 実装時の基本方針:
   - TDD: テストを先に書く（`tdd-guide` エージェントをプロアクティブに使用）
   - コードレビュー: コード作成後に `code-reviewer` エージェントをプロアクティブに使用
   - Go変更時: `task lint` と `task test` を実行
   - iOS変更時: `task ios:format`、`task ios:lint`、`task ios:test` を実行
   - iOS Mock再生成が必要な場合: `task ios:generate-mocks`
   - API/UI変更時: Chrome DevTools MCPで動作確認
3. 実装中に要件の不明点が出たら、PdMがユーザーに質問する
4. コミットする（Conventional Commits形式）

### Phase 4: レビューサイクル

PR Review Toolkitを使用して包括的レビューを実行し、CRITICAL/IMPORTANTイシューがなくなるまで繰り返す。

1. 並列で3つのレビューエージェントを起動する:
   - `pr-review-toolkit:code-reviewer`: コード品質、CLAUDE.md準拠
   - `pr-review-toolkit:pr-test-analyzer`: テストカバレッジ
   - `pr-review-toolkit:silent-failure-hunter`: エラーハンドリング、サイレント失敗
2. 全エージェントの結果を集約する
3. CRITICAL/IMPORTANT イシューがある場合:
   - 偽陽性でないか判断する（既知の理由がある場合はスキップ）
   - 実際の問題を修正する
   - リント/テストを再実行する
   - コミットしてプッシュする
   - レビューエージェントを再起動する（ステップ1に戻る）
4. レビュー中に要件の疑問が生じたら、PdMがユーザーに質問する
5. CRITICAL/IMPORTANT イシューがゼロになったらPhase 5に進む

### Phase 5: PR作成

1. 変更をプッシュする
   ```bash
   git push -u origin edg-{番号}
   ```
2. PRを作成する
   - タイトル: イシューのタイトルに基づく簡潔なもの（70文字以内）
   - Body: 概要 + 変更内容 + `Fixes EDG-{番号}`
   ```bash
   gh pr create --title "タイトル" --body "$(cat <<'EOF'
   ## 概要
   - 変更の説明

   ## 変更内容
   - 変更1
   - 変更2

   Fixes EDG-{番号}
   EOF
   )"
   ```
3. PR URLをユーザーに返す

### Phase 6: ユーザーレビュー依頼（difit）

1. `difit` をバックグラウンドで起動し、mainとの差分をブラウザで表示する
   ```bash
   # バックグラウンドで起動（run_in_background: true）
   # git diff をパイプで渡す形式で実行する
   git diff main...HEAD | difit
   ```
2. ユーザーにレビュー依頼を通知する
   - difitのローカルURL（通常 http://localhost:3000 ）を伝える
   - 「difitでmainとの差分を表示しました。レビューをお願いします。質問や修正要望があればお知らせください。」
3. ユーザーの応答を待機する
   - ユーザーからの質問には PdM + 該当エンジニアが回答する
   - 修正要望があれば実装し、コミット・プッシュする
   - 修正後はdifitを停止・再起動して最新の差分を反映する
   - 修正後は必要に応じて Phase 4（レビューサイクル）を再実行する
4. ユーザーが承認したら完了

---

## 重要なルール

- Phase 2（計画）完了後、必ずユーザーの承認を待つ
- Phase 4（レビュー）はCRITICAL/IMPORTANTがゼロになるまで自動で繰り返す
  - ただし、PRの変更範囲外の既存コードの問題は対象外とする
- 疑問や解釈の余地がある要件は、どのフェーズでもPdMがユーザーに質問する
  - 曖昧なまま実装に入らない
- Linearのステータスを「Done」に変更しない
- コミットメッセージはConventional Commits形式
- PRのBodyには必ず `Fixes EDG-{番号}` を含める
- Phase 6でdifitはバックグラウンドで起動し、ユーザーのレビュー完了まで待機する
- difitでレビューを依頼する前に、すべての変更をコミット・プッシュすること
  - 未コミットの変更は `git diff main...HEAD` に含まれないため、difitに反映されない
  - 修正後に再レビューを求める場合も、必ずコミットしてからdifitを再起動する
