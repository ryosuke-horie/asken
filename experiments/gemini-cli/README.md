# Gemini CLI 画像分析実験

このディレクトリには、Gemini CLIを使った画像分析の検証実験が含まれています。

## 検証の目的

askenプロジェクトで、Gemini CLI（Gemini 3 API）を使って食事画像から食材を判定し、カロリーや栄養素を推定する機能の実現可能性を検証します。

## ディレクトリ構成

```
experiments/gemini-cli/
├── README.md            # このファイル
├── REPORT.md            # 詳細な検証レポート
├── images/              # テスト用食事画像
│   ├── IMG_0360.HEIC
│   ├── IMG_0369.HEIC
│   └── IMG_0374.JPG
├── results/             # 実行結果（JSON）
│   ├── test_jpg_result.json
│   ├── test_heic_result.json
│   └── golang_result.json
└── scripts/             # Golangサンプルコード
    ├── main.go          # メインプログラム
    └── go.mod           # Go モジュール定義
```

## クイックスタート

### 前提条件

- Gemini CLIがインストール済み（`gemini --version`で確認）
- Go 1.23以上

### 1. コマンドラインから直接実行

```bash
# JPG画像の分析
gemini "この画像に写っている食材や料理を特定し、それぞれのカロリーと主要な栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。 @./images/IMG_0374.JPG" -o json

# HEIC画像の分析
gemini "この画像に写っている食材や料理を特定し、それぞれのカロリーと主要な栄養素（タンパク質、脂質、炭水化物）を推定してJSON形式で出力してください。 @./images/IMG_0360.HEIC" -o json
```

### 2. Golangサンプルの実行

```bash
# このディレクトリから実行
go run scripts/main.go
```

**出力例**:
```
🔍 Gemini CLIで画像を分析中...
画像: images/IMG_0374.JPG
タイムアウト: 1m0s

✅ 分析完了
セッションID: 042c91f9-5c87-4278-a981-238e1bb9cd55

📊 レスポンス:
[食材リストのJSON]

💾 結果を保存しました: results/golang_result.json
```

## 検証結果サマリー

### ✅ 成功したこと

- 非対話モードでの画像分析
- JPG/HEIC両形式のサポート
- 複数食材の同時認識（9種類を正確に識別）
- JSON形式での構造化出力
- Golangからのシェルコマンド実行
- タイムアウト・エラーハンドリング

### ⚠️ 制約事項

- **実行時間**: 約40秒（実用上は問題なし）
- **ワークスペース制限**: 実行ディレクトリからアクセス可能な範囲に制限
- **標準出力の汚染**: `"Loaded cached credentials."`などのメッセージが含まれる

## 詳細レポート

詳細な検証結果、制約事項、実装方法、今後の課題については、`REPORT.md`を参照してください。

## 次のステップ

1. PostgreSQLの食品マスタデータベース設計
2. Golangバックエンドでのエンドポイント実装
3. Next.jsフロントエンドとの連携
4. 実際の食事画像での精度検証

## 関連リンク

- [Linear課題 EDG-315](https://linear.app/edge-work/issue/EDG-315/asken：geminicliをaiエージェント用途活用できるか検証)
- [Gemini CLI GitHub](https://github.com/google-gemini/gemini-cli)
- [Gemini API Documentation](https://ai.google.dev/gemini-api/docs)

---

**検証日**: 2026-01-17
**担当者**: r-horie
