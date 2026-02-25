package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
)

// firestoreMyMenuDocument はFirestoreに保存するマイメニュードキュメント構造
type firestoreMyMenuDocument struct {
	ID                  string                 `firestore:"id"`
	Name                string                 `firestore:"name"`
	Foods               []gemini.NutritionInfo `firestore:"foods"`
	TotalCalories       float64                `firestore:"totalCalories"`
	TotalProtein        float64                `firestore:"totalProtein"`
	TotalFat            float64                `firestore:"totalFat"`
	TotalCarbohydrates  float64                `firestore:"totalCarbohydrates"`
	TotalMicronutrients map[string]float64     `firestore:"totalMicronutrients,omitempty"`
	CreatedAt           time.Time              `firestore:"createdAt"`
	UpdatedAt           time.Time              `firestore:"updatedAt"`
}

// firestoreMyMenuRepository はFirestoreを使用したMyMenuRepositoryの実装
type firestoreMyMenuRepository struct {
	client *firestore.Client
}

// NewMyMenuRepository は新しいFirestoreベースのマイメニューリポジトリを作成します
func NewMyMenuRepository(client *firestore.Client) (MyMenuRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	return &firestoreMyMenuRepository{client: client}, nil
}

// getMyMenuCollection はユーザーのmyMenuコレクション参照を取得
func (r *firestoreMyMenuRepository) getMyMenuCollection(userID string) *firestore.CollectionRef {
	return r.client.Collection("users").Doc(userID).Collection("myMenu")
}

// Create は新しいマイメニューを作成します
func (r *firestoreMyMenuRepository) Create(ctx context.Context, userID string, name string, foods []gemini.NutritionInfo) (*MyMenuItem, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	if err := validateMyMenuName(name); err != nil {
		return nil, err
	}

	if err := validateFoods(foods); err != nil {
		return nil, err
	}

	now := time.Now()
	id := uuid.New().String()

	totalCalories, totalProtein, totalFat, totalCarbs, totalMicro := calculateTotals(foods)

	doc := firestoreMyMenuDocument{
		ID:                  id,
		Name:                name,
		Foods:               foods,
		TotalCalories:       totalCalories,
		TotalProtein:        totalProtein,
		TotalFat:            totalFat,
		TotalCarbohydrates:  totalCarbs,
		TotalMicronutrients: totalMicro,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	_, err := r.getMyMenuCollection(userID).Doc(id).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("マイメニューの作成に失敗: %w", err)
	}

	return r.toMyMenuItem(&doc), nil
}

// List はユーザーのマイメニュー一覧を取得します
func (r *firestoreMyMenuRepository) List(ctx context.Context, userID string) ([]MyMenuItem, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	iter := r.getMyMenuCollection(userID).
		OrderBy("updatedAt", firestore.Desc).
		Documents(ctx)
	defer iter.Stop()

	var items []MyMenuItem
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var fsDoc firestoreMyMenuDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		items = append(items, *r.toMyMenuItem(&fsDoc))
	}

	return items, nil
}

// Get は指定されたIDのマイメニューを取得します
func (r *firestoreMyMenuRepository) Get(ctx context.Context, userID string, menuID string) (*MyMenuItem, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	if menuID == "" {
		return nil, fmt.Errorf("menuIDが必要です")
	}

	doc, err := r.getMyMenuCollection(userID).Doc(menuID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("マイメニューが見つかりません: %s: %w", menuID, ErrNotFound)
		}
		return nil, fmt.Errorf("マイメニューの取得に失敗: %w", err)
	}

	var fsDoc firestoreMyMenuDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	return r.toMyMenuItem(&fsDoc), nil
}

// Update はマイメニューを更新します
func (r *firestoreMyMenuRepository) Update(ctx context.Context, userID string, menuID string, name string, foods []gemini.NutritionInfo) (*MyMenuItem, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	if menuID == "" {
		return nil, fmt.Errorf("menuIDが必要です")
	}

	if err := validateMyMenuName(name); err != nil {
		return nil, err
	}

	if err := validateFoods(foods); err != nil {
		return nil, err
	}

	docRef := r.getMyMenuCollection(userID).Doc(menuID)

	// ドキュメントの存在確認
	existingDoc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("マイメニューが見つかりません: %s: %w", menuID, ErrNotFound)
		}
		return nil, fmt.Errorf("マイメニューの取得に失敗: %w", err)
	}

	var fsDoc firestoreMyMenuDocument
	if err := existingDoc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	now := time.Now()
	totalCalories, totalProtein, totalFat, totalCarbs, totalMicro := calculateTotals(foods)

	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "name", Value: name},
		{Path: "foods", Value: foods},
		{Path: "totalCalories", Value: totalCalories},
		{Path: "totalProtein", Value: totalProtein},
		{Path: "totalFat", Value: totalFat},
		{Path: "totalCarbohydrates", Value: totalCarbs},
		{Path: "totalMicronutrients", Value: totalMicro},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		return nil, fmt.Errorf("マイメニューの更新に失敗: %w", err)
	}

	fsDoc.Name = name
	fsDoc.Foods = foods
	fsDoc.TotalCalories = totalCalories
	fsDoc.TotalProtein = totalProtein
	fsDoc.TotalFat = totalFat
	fsDoc.TotalCarbohydrates = totalCarbs
	fsDoc.TotalMicronutrients = totalMicro
	fsDoc.UpdatedAt = now

	return r.toMyMenuItem(&fsDoc), nil
}

// Delete はマイメニューを削除します
func (r *firestoreMyMenuRepository) Delete(ctx context.Context, userID string, menuID string) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}

	if menuID == "" {
		return fmt.Errorf("menuIDが必要です")
	}

	docRef := r.getMyMenuCollection(userID).Doc(menuID)

	// 存在確認
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("マイメニューが見つかりません: %s: %w", menuID, ErrNotFound)
		}
		return fmt.Errorf("マイメニューの取得に失敗: %w", err)
	}

	_, err = docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("マイメニューの削除に失敗: %w", err)
	}

	return nil
}

// toMyMenuItem はFirestoreドキュメントをMyMenuItemに変換
func (r *firestoreMyMenuRepository) toMyMenuItem(doc *firestoreMyMenuDocument) *MyMenuItem {
	return &MyMenuItem{
		ID:                  doc.ID,
		Name:                doc.Name,
		Foods:               doc.Foods,
		TotalCalories:       doc.TotalCalories,
		TotalProtein:        doc.TotalProtein,
		TotalFat:            doc.TotalFat,
		TotalCarbohydrates:  doc.TotalCarbohydrates,
		TotalMicronutrients: doc.TotalMicronutrients,
		CreatedAt:           doc.CreatedAt,
		UpdatedAt:           doc.UpdatedAt,
	}
}

// validateMyMenuName はマイメニュー名をバリデーションします
func validateMyMenuName(name string) error {
	if name == "" {
		return fmt.Errorf("メニュー名は必須です")
	}
	if len(name) > 50 {
		return fmt.Errorf("メニュー名は50文字以内である必要があります")
	}
	return nil
}

// validateFoods は食品リストをバリデーションします
func validateFoods(foods []gemini.NutritionInfo) error {
	if len(foods) == 0 {
		return fmt.Errorf("少なくとも1つの食品が必要です")
	}
	if len(foods) > 100 {
		return fmt.Errorf("食品は100個以内である必要があります")
	}
	return nil
}

// calculateTotals は食品リストから総栄養素を計算します
func calculateTotals(foods []gemini.NutritionInfo) (calories, protein, fat, carbs float64, micro map[string]float64) {
	micro = make(map[string]float64)
	for _, food := range foods {
		calories += food.Calories
		protein += food.Protein
		fat += food.Fat
		carbs += food.Carbohydrates
		micro = gemini.MergeMicronutrients(micro, food.Micronutrients)
	}
	return
}
