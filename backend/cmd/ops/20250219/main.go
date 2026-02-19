// マイメニュー既存データへのマイクロニュートリエントバックフィルスクリプト
//
// UTK-23で追加されたマイクロニュートリエント12種について、既存のマイメニューデータには
// micronutrientsフィールドが存在しない。本スクリプトはGemini APIを使用して栄養素を
// 再推定し、micronutrientsフィールドをバックフィルする。
//
// 実行方法:
//
//	cd backend
//	GEMINI_API_KEY=xxx GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json \
//	  go run ./cmd/ops/20250219/
//
// 注意: 本スクリプトは本番Firestoreに直接書き込む。実行前にデータをバックアップすること。
// 冪等性あり: micronutrientsが既に存在するドキュメントはスキップされるため再実行可能。
// 実行時間の目安: パッチ対象メニュー1件あたり約6秒（Gemini APIクォータ待機込み）。
//
// 環境変数:
//
//	GEMINI_API_KEY                  - Gemini APIキー（必須）
//	GOOGLE_APPLICATION_CREDENTIALS  - サービスアカウントキーのパス（必須）
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/database"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

const (
	// geminiTimeout はGemini APIの1リクエストあたりのタイムアウト
	geminiTimeout = 60 * time.Second

	// rateLimitWait はGemini APIへのリクエスト間隔
	// Gemini API側のクォータ超過（429 Too Many Requests）を避けるための待機時間
	// パッチ成功後にのみ適用される（processUserMenus 参照）
	rateLimitWait = 6 * time.Second
)

// menuDoc はFirestoreのマイメニュードキュメントの読み取り用構造体
// DataTo()は未定義フィールドを無視するため、foods以外のフィールドは安全に保持される
// バックフィルに必要なfoodsフィールドのみ定義する
type menuDoc struct {
	Foods []gemini.NutritionInfo `firestore:"foods"`
}

func main() {
	ctx := context.Background()

	// Firestoreクライアント初期化（GOOGLE_APPLICATION_CREDENTIALS使用）
	fsClient, err := database.NewFirestoreClient(ctx, "")
	if err != nil {
		log.Fatalf("[FATAL] Firestore初期化失敗: %v", err)
	}
	defer func() {
		if err := fsClient.Close(); err != nil {
			log.Printf("[WARN] Firestoreクライアントのクローズに失敗: %v", err)
		}
	}()

	// NutritionCalculator初期化（GEMINI_API_KEY環境変数使用）
	calc, err := gemini.NewNutritionCalculator(geminiTimeout)
	if err != nil {
		log.Fatalf("[FATAL] NutritionCalculator初期化失敗: %v", err)
	}

	log.Printf("[INFO] マイメニューマイクロニュートリエントバックフィル開始")

	totalMenus, patchedMenus, skippedMenus, failedMenus := processAllUsers(ctx, fsClient, calc)

	log.Printf("[INFO] バックフィル完了 - 総メニュー: %d, パッチ適用: %d, スキップ: %d, 失敗: %d",
		totalMenus, patchedMenus, skippedMenus, failedMenus)

	if failedMenus > 0 {
		log.Printf("[ERROR] %d件のメニューでパッチに失敗しました。ログを確認して再実行してください。", failedMenus)
		os.Exit(1)
	}
}

// processAllUsers は全ユーザーのマイメニューを処理する
func processAllUsers(ctx context.Context, fsClient *firestore.Client, calc *gemini.NutritionCalculator) (total, patched, skipped, failed int) {
	usersIter := fsClient.Collection("users").Documents(ctx)
	defer usersIter.Stop()

	for {
		userDoc, err := usersIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// イテレータエラーは続行不可能なため即座に終了（failedをインクリメントしてos.Exit(1)を保証）
			log.Printf("[ERROR] ユーザー列挙エラーが発生しました。処理を中断します: %v", err)
			failed++
			return
		}

		userID := userDoc.Ref.ID
		log.Printf("[INFO] ユーザー処理開始: %s", userID)

		p, s, f := processUserMenus(ctx, fsClient, calc, userID)
		total += p + s + f
		patched += p
		skipped += s
		failed += f
	}
	return
}

// processUserMenus は指定ユーザーの全マイメニューを処理する
func processUserMenus(ctx context.Context, fsClient *firestore.Client, calc *gemini.NutritionCalculator, userID string) (patched, skipped, failed int) {
	menuIter := fsClient.Collection("users").Doc(userID).Collection("myMenu").Documents(ctx)
	defer menuIter.Stop()

	for {
		doc, err := menuIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// イテレータエラーは後続のNext()も同じエラーを返すため、このユーザーの処理を中断
			log.Printf("[ERROR] マイメニュー列挙エラーが発生しました。このユーザーの処理を中断します (user=%s): %v", userID, err)
			failed++
			return
		}

		menuID := doc.Ref.ID

		var menu menuDoc
		if err := doc.DataTo(&menu); err != nil {
			log.Printf("[ERROR] パースエラー (user=%s, menu=%s): %v", userID, menuID, err)
			failed++
			continue
		}

		if len(menu.Foods) == 0 {
			log.Printf("[WARN] foodsが空のメニューをスキップします。データ異常の可能性があります (user=%s, menu=%s)", userID, menuID)
			skipped++
			continue
		}

		if !needsMicronutrientPatch(menu.Foods) {
			log.Printf("[SKIP] パッチ不要 (user=%s, menu=%s)", userID, menuID)
			skipped++
			continue
		}

		if err := patchMenuMicronutrients(ctx, fsClient, calc, userID, menuID, menu.Foods); err != nil {
			log.Printf("[ERROR] パッチ失敗 (user=%s, menu=%s): %v", userID, menuID, err)
			failed++
			continue
		}

		patched++

		// レート制限のため待機
		time.Sleep(rateLimitWait)
	}
	return
}

// needsMicronutrientPatch はFoods配列内にmicronutrientsが空（nilまたは空map）の食材があるか確認する
// micronutrientsが部分的に存在する場合（一部キー欠落）はパッチ不要と判定する（設計上の意図）
func needsMicronutrientPatch(foods []gemini.NutritionInfo) bool {
	for _, food := range foods {
		if len(food.Micronutrients) == 0 {
			return true
		}
	}
	return false
}

// patchMenuMicronutrients は指定マイメニューのmicronutrientsをGemini APIで推定してFirestoreを更新する
func patchMenuMicronutrients(
	ctx context.Context,
	fsClient *firestore.Client,
	calc *gemini.NutritionCalculator,
	userID, menuID string,
	foods []gemini.NutritionInfo,
) error {
	foodItems := toFoodItems(foods)

	nutritionList, err := calc.CalculateNutrition(ctx, foodItems)
	if err != nil {
		return fmt.Errorf("Gemini API呼び出し失敗: %w", err)
	}

	if len(nutritionList) != len(foods) {
		return fmt.Errorf("Geminiレスポンス件数不一致: 期待=%d, 実際=%d", len(foods), len(nutritionList))
	}

	// Geminiレスポンスの食材順序を検証する（インデックス対応が前提）
	// 順序が異なると誤ったmicronutrientsが書き込まれるため厳密にチェックする
	for i, original := range foods {
		if nutritionList[i].Name != original.Name {
			return fmt.Errorf("Geminiレスポンスの食材名不一致 (index=%d): 期待=%q, 実際=%q",
				i, original.Name, nutritionList[i].Name)
		}
	}

	updatedFoods := applyMicronutrients(foods, nutritionList)

	docRef := fsClient.Collection("users").Doc(userID).Collection("myMenu").Doc(menuID)
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "foods", Value: updatedFoods},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("Firestore更新失敗: %w", err)
	}

	log.Printf("[OK] パッチ完了 (user=%s, menu=%s, foods=%d品)", userID, menuID, len(updatedFoods))
	return nil
}

// toFoodItems はNutritionInfo配列をFoodItem配列に変換する
func toFoodItems(foods []gemini.NutritionInfo) []gemini.FoodItem {
	items := make([]gemini.FoodItem, len(foods))
	for i, food := range foods {
		items[i] = gemini.FoodItem{
			Name:            food.Name,
			EstimatedAmount: food.EstimatedAmount,
		}
	}
	return items
}

// applyMicronutrients はGeminiの計算結果からmicronutrientsを既存foodsに適用する
// original[i]とgeminiResults[i]が同一食材に対応することを前提とする（インデックス対応）
// micronutrientsが既に存在するfoodはスキップし、不足しているfoodのみ更新する
// 既存のカロリー・タンパク質・脂質・炭水化物は保持する（Geminiのマクロ栄養素計算結果は無視）
func applyMicronutrients(original []gemini.NutritionInfo, geminiResults []gemini.NutritionInfo) []gemini.NutritionInfo {
	updated := make([]gemini.NutritionInfo, len(original))
	for i, food := range original {
		updated[i] = food
		if len(food.Micronutrients) == 0 {
			updated[i].Micronutrients = geminiResults[i].Micronutrients
		}
	}
	return updated
}
