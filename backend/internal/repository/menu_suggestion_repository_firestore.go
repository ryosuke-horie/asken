package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// firestoreMenuSuggestionIngredient はFirestoreに保存する食材要素
type firestoreMenuSuggestionIngredient struct {
	IngredientID string  `firestore:"ingredientId"`
	Name         string  `firestore:"name"`
	Quantity     float64 `firestore:"quantity"`
	Unit         string  `firestore:"unit"`
}

// firestoreEstimatedNutrition はFirestoreに保存する推定栄養素
type firestoreEstimatedNutrition struct {
	Calories      float64 `firestore:"calories"`
	Protein       float64 `firestore:"protein"`
	Fat           float64 `firestore:"fat"`
	Carbohydrates float64 `firestore:"carbohydrates"`
}

// firestoreMenuSuggestionDocument はFirestoreに保存するメニューサジェストドキュメント
type firestoreMenuSuggestionDocument struct {
	ID                 string                              `firestore:"id"`
	Title              string                              `firestore:"title"`
	Description        string                              `firestore:"description"`
	Reason             string                              `firestore:"reason"`
	IngredientsUsed    []firestoreMenuSuggestionIngredient `firestore:"ingredientsUsed"`
	Recipe             string                              `firestore:"recipe,omitempty"`
	EstimatedNutrition firestoreEstimatedNutrition         `firestore:"estimatedNutrition"`
	MealType           string                              `firestore:"mealType"`
	Status             string                              `firestore:"status"`
	CreatedAt          time.Time                           `firestore:"createdAt"`
	UpdatedAt          time.Time                           `firestore:"updatedAt"`
}

// acceptIngredientSnap はトランザクション内で読み取った食材スナップショット
type acceptIngredientSnap struct {
	ref  *firestore.DocumentRef
	snap *firestore.DocumentSnapshot
	ing  MenuSuggestionIngredient
}

// firestoreMenuSuggestionRepository はFirestoreを使用したMenuSuggestionRepositoryの実装
type firestoreMenuSuggestionRepository struct {
	client *firestore.Client
}

// NewMenuSuggestionRepository は新しいFirestoreベースのMenuSuggestionRepositoryを作成する
func NewMenuSuggestionRepository(client *firestore.Client) (MenuSuggestionRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	return &firestoreMenuSuggestionRepository{client: client}, nil
}

// getUserMenuSuggestionCollection はユーザーのmenuSuggestionsコレクション参照を取得
func (r *firestoreMenuSuggestionRepository) getUserMenuSuggestionCollection(userID string) *firestore.CollectionRef {
	return r.client.Collection("users").Doc(userID).Collection("menuSuggestions")
}

// Create は新しいメニューサジェストを作成する
func (r *firestoreMenuSuggestionRepository) Create(ctx context.Context, userID string, input CreateMenuSuggestionInput) (*MenuSuggestion, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New().String()

	fsIngs := make([]firestoreMenuSuggestionIngredient, len(input.IngredientsUsed))
	for i, ing := range input.IngredientsUsed {
		fsIngs[i] = firestoreMenuSuggestionIngredient(ing)
	}

	doc := firestoreMenuSuggestionDocument{
		ID:              id,
		Title:           input.Title,
		Description:     input.Description,
		Reason:          input.Reason,
		IngredientsUsed: fsIngs,
		EstimatedNutrition: firestoreEstimatedNutrition{
			Calories:      input.EstimatedNutrition.Calories,
			Protein:       input.EstimatedNutrition.Protein,
			Fat:           input.EstimatedNutrition.Fat,
			Carbohydrates: input.EstimatedNutrition.Carbohydrates,
		},
		MealType:  input.MealType,
		Status:    string(MenuStatusSuggested),
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := r.getUserMenuSuggestionCollection(userID).Doc(id).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("メニューサジェストの作成に失敗: %w", err)
	}

	return r.toMenuSuggestion(&doc), nil
}

// List はユーザーのメニューサジェスト一覧を取得する
func (r *firestoreMenuSuggestionRepository) List(ctx context.Context, userID string, status string, limit int) ([]MenuSuggestion, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}
	if limit <= 0 {
		limit = 10
	}

	collection := r.getUserMenuSuggestionCollection(userID)
	var iter *firestore.DocumentIterator

	if status != "" {
		iter = collection.Where("status", "==", status).OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx)
	} else {
		iter = collection.OrderBy("createdAt", firestore.Desc).Limit(limit).Documents(ctx)
	}
	defer iter.Stop()

	var items []MenuSuggestion
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("メニューサジェストの取得に失敗: %w", err)
		}

		var fsDoc firestoreMenuSuggestionDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}
		items = append(items, *r.toMenuSuggestion(&fsDoc))
	}

	if items == nil {
		items = []MenuSuggestion{}
	}
	return items, nil
}

// GetByID は指定されたIDのメニューサジェストを取得する
func (r *firestoreMenuSuggestionRepository) GetByID(ctx context.Context, userID string, id string) (*MenuSuggestion, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}
	if id == "" {
		return nil, fmt.Errorf("IDが必要です")
	}

	doc, err := r.getUserMenuSuggestionCollection(userID).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("メニューサジェストが見つかりません: %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("メニューサジェストの取得に失敗: %w", err)
	}

	var fsDoc firestoreMenuSuggestionDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}
	return r.toMenuSuggestion(&fsDoc), nil
}

// UpdateRecipe はレシピを更新する
func (r *firestoreMenuSuggestionRepository) UpdateRecipe(ctx context.Context, userID string, id string, recipe string) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}
	if id == "" {
		return fmt.Errorf("IDが必要です")
	}

	_, err := r.getUserMenuSuggestionCollection(userID).Doc(id).Update(ctx, []firestore.Update{
		{Path: "recipe", Value: recipe},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("メニューサジェストが見つかりません: %s: %w", id, ErrNotFound)
		}
		return fmt.Errorf("レシピの更新に失敗: %w", err)
	}
	return nil
}

// Accept はサジェストを採用し、食事記録と食材控除をトランザクションで実行する
func (r *firestoreMenuSuggestionRepository) Accept(ctx context.Context, userID string, id string) (*AcceptMenuSuggestionResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}
	if id == "" {
		return nil, fmt.Errorf("IDが必要です")
	}

	suggestion, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if suggestion.Status != string(MenuStatusSuggested) {
		return nil, fmt.Errorf("%w: status=%s", ErrAlreadyProcessed, suggestion.Status)
	}

	var result AcceptMenuSuggestionResult
	err = r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		result = AcceptMenuSuggestionResult{DeductedIngredients: make([]DeductedIngredient, 0)}
		now := time.Now()

		suggestionRef := r.getUserMenuSuggestionCollection(userID).Doc(id)
		fsDoc, err := r.txValidateSuggestion(tx, suggestionRef)
		if err != nil {
			return err
		}

		ingSnaps, err := r.txCollectIngredientSnaps(tx, userID, suggestion.IngredientsUsed)
		if err != nil {
			return err
		}

		if err := tx.Update(suggestionRef, []firestore.Update{
			{Path: "status", Value: string(MenuStatusAccepted)},
			{Path: "updatedAt", Value: now},
		}); err != nil {
			return fmt.Errorf("サジェストのステータス更新に失敗: %w", err)
		}

		analysisID := uuid.New()
		mealDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		analysisRef := r.client.Collection("users").Doc(userID).Collection("analysisRequests").Doc(analysisID.String())
		analysisDoc := r.buildAnalysisDoc(analysisID.String(), fsDoc, mealDate, now)
		if err := tx.Set(analysisRef, analysisDoc); err != nil {
			return fmt.Errorf("食事記録の作成に失敗: %w", err)
		}
		result.AnalysisRequestID = analysisID.String()

		deducted, err := r.txDeductIngredients(tx, ingSnaps, now)
		if err != nil {
			return err
		}
		result.DeductedIngredients = deducted

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("採用トランザクションに失敗: %w", err)
	}

	return &result, nil
}

// txValidateSuggestion はトランザクション内でサジェストを読み取り、ステータスを検証する
func (r *firestoreMenuSuggestionRepository) txValidateSuggestion(tx *firestore.Transaction, ref *firestore.DocumentRef) (*firestoreMenuSuggestionDocument, error) {
	snap, err := tx.Get(ref)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("メニューサジェストが見つかりません: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("サジェストの取得に失敗: %w", err)
	}
	var fsDoc firestoreMenuSuggestionDocument
	if err := snap.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("サジェストのパースに失敗: %w", err)
	}
	if fsDoc.Status != string(MenuStatusSuggested) {
		return nil, fmt.Errorf("%w: status=%s", ErrAlreadyProcessed, fsDoc.Status)
	}
	return &fsDoc, nil
}

// txCollectIngredientSnaps はトランザクション内で食材スナップショットを収集する
func (r *firestoreMenuSuggestionRepository) txCollectIngredientSnaps(tx *firestore.Transaction, userID string, ings []MenuSuggestionIngredient) ([]acceptIngredientSnap, error) {
	snaps := make([]acceptIngredientSnap, 0, len(ings))
	for _, ing := range ings {
		if ing.IngredientID == "" {
			continue
		}
		ingRef := r.client.Collection("users").Doc(userID).Collection("ingredients").Doc(ing.IngredientID)
		ingSnap, err := tx.Get(ingRef)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, fmt.Errorf("食材の取得に失敗 (%s): %w", ing.IngredientID, err)
			}
			ingSnap = nil
		}
		snaps = append(snaps, acceptIngredientSnap{ref: ingRef, snap: ingSnap, ing: ing})
	}
	return snaps, nil
}

// buildAnalysisDoc はaccept時に作成するanalysisRequestsドキュメントを構築する
func (r *firestoreMenuSuggestionRepository) buildAnalysisDoc(analysisID string, fsDoc *firestoreMenuSuggestionDocument, mealDate, now time.Time) firestoreAnalysisDocument {
	return firestoreAnalysisDocument{
		ID:        analysisID,
		Status:    StatusCompleted,
		InputType: InputTypeSuggestion,
		InputText: fsDoc.Title,
		MealType:  fsDoc.MealType,
		MealDate:  mealDate,
		Confirmed: true,
		CreatedAt: now,
		UpdatedAt: now,
		Result: &firestoreAnalysisResult{
			Foods: []gemini.NutritionInfo{
				{
					Name:            fsDoc.Title,
					EstimatedAmount: "1食分",
					Calories:        fsDoc.EstimatedNutrition.Calories,
					Protein:         fsDoc.EstimatedNutrition.Protein,
					Fat:             fsDoc.EstimatedNutrition.Fat,
					Carbohydrates:   fsDoc.EstimatedNutrition.Carbohydrates,
				},
			},
			TotalCalories:      fsDoc.EstimatedNutrition.Calories,
			TotalProtein:       fsDoc.EstimatedNutrition.Protein,
			TotalFat:           fsDoc.EstimatedNutrition.Fat,
			TotalCarbohydrates: fsDoc.EstimatedNutrition.Carbohydrates,
		},
	}
}

// txDeductIngredients はトランザクション内で食材の在庫を控除する
func (r *firestoreMenuSuggestionRepository) txDeductIngredients(tx *firestore.Transaction, snaps []acceptIngredientSnap, now time.Time) ([]DeductedIngredient, error) {
	results := make([]DeductedIngredient, 0, len(snaps))
	for _, is := range snaps {
		if is.snap == nil || !is.snap.Exists() {
			results = append(results, DeductedIngredient{
				IngredientID: is.ing.IngredientID,
				Name:         is.ing.Name,
				Deducted:     0,
				Remaining:    0,
			})
			continue
		}

		var fsIng firestoreIngredientDocument
		if err := is.snap.DataTo(&fsIng); err != nil {
			return nil, fmt.Errorf("食材ドキュメントのパースに失敗 (%s): %w", is.ing.IngredientID, err)
		}

		newQty := fsIng.Quantity - is.ing.Quantity
		deducted, err := r.txApplyIngredientDeduction(tx, is, fsIng, newQty, now)
		if err != nil {
			return nil, err
		}
		results = append(results, deducted)
	}
	return results, nil
}

// txApplyIngredientDeduction は1件の食材に対して控除を適用し結果を返す
func (r *firestoreMenuSuggestionRepository) txApplyIngredientDeduction(tx *firestore.Transaction, is acceptIngredientSnap, fsIng firestoreIngredientDocument, newQty float64, now time.Time) (DeductedIngredient, error) {
	if newQty <= 0 {
		if err := tx.Delete(is.ref); err != nil {
			return DeductedIngredient{}, fmt.Errorf("食材の削除に失敗 (%s): %w", is.ing.IngredientID, err)
		}
		return DeductedIngredient{
			IngredientID: is.ing.IngredientID,
			Name:         is.ing.Name,
			Deducted:     fsIng.Quantity,
			Remaining:    0,
		}, nil
	}

	if err := tx.Update(is.ref, []firestore.Update{
		{Path: "quantity", Value: newQty},
		{Path: "updatedAt", Value: now},
	}); err != nil {
		return DeductedIngredient{}, fmt.Errorf("食材の更新に失敗 (%s): %w", is.ing.IngredientID, err)
	}
	return DeductedIngredient{
		IngredientID: is.ing.IngredientID,
		Name:         is.ing.Name,
		Deducted:     is.ing.Quantity,
		Remaining:    newQty,
	}, nil
}

// Dismiss はサジェストを却下する
func (r *firestoreMenuSuggestionRepository) Dismiss(ctx context.Context, userID string, id string) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}
	if id == "" {
		return fmt.Errorf("IDが必要です")
	}

	docRef := r.getUserMenuSuggestionCollection(userID).Doc(id)
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if _, err := r.txValidateSuggestion(tx, docRef); err != nil {
			return err
		}
		return tx.Update(docRef, []firestore.Update{
			{Path: "status", Value: string(MenuStatusDismissed)},
			{Path: "updatedAt", Value: time.Now()},
		})
	})
	if err != nil {
		return fmt.Errorf("却下トランザクションに失敗: %w", err)
	}
	return nil
}

// toMenuSuggestion はFirestoreドキュメントをMenuSuggestionに変換する
func (r *firestoreMenuSuggestionRepository) toMenuSuggestion(doc *firestoreMenuSuggestionDocument) *MenuSuggestion {
	ings := make([]MenuSuggestionIngredient, len(doc.IngredientsUsed))
	for i, ing := range doc.IngredientsUsed {
		ings[i] = MenuSuggestionIngredient(ing)
	}
	return &MenuSuggestion{
		ID:              doc.ID,
		Title:           doc.Title,
		Description:     doc.Description,
		Reason:          doc.Reason,
		IngredientsUsed: ings,
		Recipe:          doc.Recipe,
		EstimatedNutrition: EstimatedNutrition{
			Calories:      doc.EstimatedNutrition.Calories,
			Protein:       doc.EstimatedNutrition.Protein,
			Fat:           doc.EstimatedNutrition.Fat,
			Carbohydrates: doc.EstimatedNutrition.Carbohydrates,
		},
		MealType:  doc.MealType,
		Status:    doc.Status,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}
