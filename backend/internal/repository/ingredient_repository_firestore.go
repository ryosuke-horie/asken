package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// firestoreIngredientDocument はFirestoreに保存する食材ドキュメント構造
type firestoreIngredientDocument struct {
	ID           string     `firestore:"id"`
	Name         string     `firestore:"name"`
	Category     string     `firestore:"category"`
	Quantity     float64    `firestore:"quantity"`
	Unit         string     `firestore:"unit"`
	PurchaseDate *time.Time `firestore:"purchaseDate"`
	ExpiryDate   *time.Time `firestore:"expiryDate"`
	Source       string     `firestore:"source"`
	CreatedAt    time.Time  `firestore:"createdAt"`
	UpdatedAt    time.Time  `firestore:"updatedAt"`
}

// firestoreIngredientRepository はFirestoreを使用したIngredientRepositoryの実装
type firestoreIngredientRepository struct {
	client *firestore.Client
}

// NewIngredientRepository は新しいFirestoreベースの食材リポジトリを作成します
func NewIngredientRepository(client *firestore.Client) (IngredientRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	return &firestoreIngredientRepository{client: client}, nil
}

// getIngredientCollection はユーザーのingredientsコレクション参照を取得
func (r *firestoreIngredientRepository) getIngredientCollection(userID string) *firestore.CollectionRef {
	return r.client.Collection("users").Doc(userID).Collection("ingredients")
}

// Create は新しい食材を作成します
func (r *firestoreIngredientRepository) Create(ctx context.Context, userID string, input CreateIngredientInput) (*Ingredient, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New().String()

	doc := firestoreIngredientDocument{
		ID:           id,
		Name:         input.Name,
		Category:     input.Category,
		Quantity:     input.Quantity,
		Unit:         input.Unit,
		PurchaseDate: input.PurchaseDate,
		ExpiryDate:   input.ExpiryDate,
		Source:       input.Source,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := r.getIngredientCollection(userID).Doc(id).Set(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("食材の作成に失敗: %w", err)
	}

	return r.toIngredient(&doc), nil
}

// List はユーザーの食材一覧を取得します
func (r *firestoreIngredientRepository) List(ctx context.Context, userID string, category string) ([]Ingredient, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	collection := r.getIngredientCollection(userID)
	var iter *firestore.DocumentIterator

	if category != "" {
		iter = collection.Where("category", "==", category).OrderBy("name", firestore.Asc).Documents(ctx)
	} else {
		iter = collection.OrderBy("updatedAt", firestore.Desc).Documents(ctx)
	}
	defer iter.Stop()

	var items []Ingredient
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("食材の取得に失敗: %w", err)
		}

		var fsDoc firestoreIngredientDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		items = append(items, *r.toIngredient(&fsDoc))
	}

	if items == nil {
		items = []Ingredient{}
	}

	return items, nil
}

// GetByID は指定されたIDの食材を取得します
func (r *firestoreIngredientRepository) GetByID(ctx context.Context, userID string, ingredientID string) (*Ingredient, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}
	if ingredientID == "" {
		return nil, fmt.Errorf("ingredientIDが必要です")
	}

	doc, err := r.getIngredientCollection(userID).Doc(ingredientID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("食材が見つかりません: %s: %w", ingredientID, ErrNotFound)
		}
		return nil, fmt.Errorf("食材の取得に失敗: %w", err)
	}

	var fsDoc firestoreIngredientDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	return r.toIngredient(&fsDoc), nil
}

// Update は食材を更新します
func (r *firestoreIngredientRepository) Update(ctx context.Context, userID string, ingredientID string, input UpdateIngredientInput) (*Ingredient, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}
	if ingredientID == "" {
		return nil, fmt.Errorf("ingredientIDが必要です")
	}

	docRef := r.getIngredientCollection(userID).Doc(ingredientID)

	existingDoc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("食材が見つかりません: %s: %w", ingredientID, ErrNotFound)
		}
		return nil, fmt.Errorf("食材の取得に失敗: %w", err)
	}

	var fsDoc firestoreIngredientDocument
	if err := existingDoc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	now := time.Now()
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "name", Value: input.Name},
		{Path: "category", Value: input.Category},
		{Path: "quantity", Value: input.Quantity},
		{Path: "unit", Value: input.Unit},
		{Path: "purchaseDate", Value: input.PurchaseDate},
		{Path: "expiryDate", Value: input.ExpiryDate},
		{Path: "updatedAt", Value: now},
	})
	if err != nil {
		return nil, fmt.Errorf("食材の更新に失敗: %w", err)
	}

	fsDoc.Name = input.Name
	fsDoc.Category = input.Category
	fsDoc.Quantity = input.Quantity
	fsDoc.Unit = input.Unit
	fsDoc.PurchaseDate = input.PurchaseDate
	fsDoc.ExpiryDate = input.ExpiryDate
	fsDoc.UpdatedAt = now

	return r.toIngredient(&fsDoc), nil
}

// Delete は食材を削除します
func (r *firestoreIngredientRepository) Delete(ctx context.Context, userID string, ingredientID string) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}
	if ingredientID == "" {
		return fmt.Errorf("ingredientIDが必要です")
	}

	docRef := r.getIngredientCollection(userID).Doc(ingredientID)

	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("食材が見つかりません: %s: %w", ingredientID, ErrNotFound)
		}
		return fmt.Errorf("食材の取得に失敗: %w", err)
	}

	_, err = docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("食材の削除に失敗: %w", err)
	}

	return nil
}

// toIngredient はFirestoreドキュメントをIngredientに変換
func (r *firestoreIngredientRepository) toIngredient(doc *firestoreIngredientDocument) *Ingredient {
	return &Ingredient{
		ID:           doc.ID,
		Name:         doc.Name,
		Category:     doc.Category,
		Quantity:     doc.Quantity,
		Unit:         doc.Unit,
		PurchaseDate: doc.PurchaseDate,
		ExpiryDate:   doc.ExpiryDate,
		Source:       doc.Source,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}
