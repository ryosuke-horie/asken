#!/bin/bash

# 2ステップアプローチの実行スクリプト
# ステップ1: 食材分類 → ステップ2: 栄養素算出

set -e  # エラーが発生したら即座に終了

echo "🚀 2ステップアプローチの実行を開始"
echo "=================================================="
echo ""

# 作業ディレクトリに移動
cd "$(dirname "$0")/.."

# ステップ1: 食材分類
echo "📸 ステップ1: 食材分類を実行中..."
echo "--------------------------------------------------"
go run scripts/step1_classify.go
echo ""

# 中間結果の確認
if [ ! -f "results/step1_classify_result.json" ]; then
    echo "❌ エラー: ステップ1の結果ファイルが見つかりません"
    exit 1
fi

echo "✅ ステップ1完了"
echo ""
echo ""

# ステップ2: 栄養素算出
echo "🥗 ステップ2: 栄養素算出を実行中..."
echo "--------------------------------------------------"
go run scripts/step2_nutrition.go
echo ""

echo "=================================================="
echo "✅ 2ステップアプローチの実行が完了しました"
echo ""
echo "📁 結果ファイル:"
echo "  - results/step1_classify_result.json  (食材分類)"
echo "  - results/step2_nutrition_result.json (栄養素算出)"
