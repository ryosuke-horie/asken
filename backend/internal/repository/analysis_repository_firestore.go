package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/ryosuke-horie/uchikomi/backend/internal/service"
	"github.com/ryosuke-horie/uchikomi/backend/pkg/gemini"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// firestoreAnalysisDocument はFirestoreに保存するドキュメント構造
type firestoreAnalysisDocument struct {
	ID           string         `firestore:"id"`
	Status       AnalysisStatus `firestore:"status"`
	InputType    InputType      `firestore:"inputType"`
	ImagePath    string         `firestore:"imagePath,omitempty"`
	InputText    string         `firestore:"inputText,omitempty"`
	MealType     string         `firestore:"mealType,omitempty"`
	MealDate     time.Time      `firestore:"mealDate,omitempty"`
	ErrorMessage string         `firestore:"errorMessage,omitempty"`
	CreatedAt    time.Time      `firestore:"createdAt"`
	UpdatedAt    time.Time      `firestore:"updatedAt"`
	// analysis_resultsを統合
	Result *firestoreAnalysisResult `firestore:"result,omitempty"`
}

// firestoreAnalysisResult は分析結果を保持する構造体
type firestoreAnalysisResult struct {
	Foods              []gemini.NutritionInfo `firestore:"foods"`
	TotalCalories      float64                `firestore:"totalCalories"`
	TotalProtein       float64                `firestore:"totalProtein"`
	TotalFat           float64                `firestore:"totalFat"`
	TotalCarbohydrates float64                `firestore:"totalCarbohydrates"`
}

// firestoreAnalysisRepository はFirestoreを使用したAnalysisRepositoryの実装
type firestoreAnalysisRepository struct {
	client      *firestore.Client
	storageRepo StorageRepository
}

// NewAnalysisRepositoryFirestore は新しいFirestoreベースのAnalysisRepositoryを作成します
func NewAnalysisRepositoryFirestore(client *firestore.Client, storageRepo StorageRepository) (AnalysisRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("firestore client is required")
	}
	if storageRepo == nil {
		return nil, fmt.Errorf("storage repository is required")
	}
	return &firestoreAnalysisRepository{
		client:      client,
		storageRepo: storageRepo,
	}, nil
}

// getUserAnalysisCollection はユーザーのanalysisRequestsコレクション参照を取得
func (r *firestoreAnalysisRepository) getUserAnalysisCollection(userID string) *firestore.CollectionRef {
	return r.client.Collection("users").Doc(userID).Collection("analysisRequests")
}

// CreateRequest は新しい画像分析リクエストを作成します
func (r *firestoreAnalysisRepository) CreateRequest(ctx context.Context, imagePath, mealType, mealDate string, userID *string) (uuid.UUID, error) {
	if userID == nil || *userID == "" {
		return uuid.Nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New()

	mealDateTime, err := time.Parse("2006-01-02", mealDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	// 既存のskipped記録を削除
	if err := r.deleteSkippedRecords(ctx, *userID, mealType, mealDateTime); err != nil {
		return uuid.Nil, err
	}

	doc := firestoreAnalysisDocument{
		ID:        id.String(),
		Status:    StatusPending,
		InputType: InputTypeImage,
		ImagePath: imagePath,
		MealType:  mealType,
		MealDate:  mealDateTime,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = r.getUserAnalysisCollection(*userID).Doc(id.String()).Set(ctx, doc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("分析リクエストの作成に失敗: %w", err)
	}

	return id, nil
}

// CreateRequestWithText は新しいテキスト分析リクエストを作成します
func (r *firestoreAnalysisRepository) CreateRequestWithText(ctx context.Context, inputText, mealType, mealDate string, userID *string) (uuid.UUID, error) {
	if userID == nil || *userID == "" {
		return uuid.Nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New()

	mealDateTime, err := time.Parse("2006-01-02", mealDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	// 既存のskipped記録を削除
	if err := r.deleteSkippedRecords(ctx, *userID, mealType, mealDateTime); err != nil {
		return uuid.Nil, err
	}

	doc := firestoreAnalysisDocument{
		ID:        id.String(),
		Status:    StatusPending,
		InputType: InputTypeText,
		InputText: inputText,
		MealType:  mealType,
		MealDate:  mealDateTime,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = r.getUserAnalysisCollection(*userID).Doc(id.String()).Set(ctx, doc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("分析リクエストの作成に失敗: %w", err)
	}

	return id, nil
}

// GetRequest は指定されたIDの分析リクエストを取得します（userIDでスコープ）
func (r *firestoreAnalysisRepository) GetRequest(ctx context.Context, userID string, id uuid.UUID) (*AnalysisRequest, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	// ユーザーのコレクションから直接取得
	doc, err := r.getUserAnalysisCollection(userID).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("リクエストが見つかりません: %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("リクエストの取得に失敗: %w", err)
	}

	var fsDoc firestoreAnalysisDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	return r.toAnalysisRequest(&fsDoc)
}

// UpdateStatus はリクエストのステータスを更新します
func (r *firestoreAnalysisRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status AnalysisStatus, errorMessage string) error {
	// コレクショングループクエリで対象ドキュメントを検索
	iter := r.client.CollectionGroup("analysisRequests").Where("id", "==", id.String()).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return fmt.Errorf("リクエストが見つかりません: %s", id)
	}
	if err != nil {
		return fmt.Errorf("リクエストの取得に失敗: %w", err)
	}

	updates := []firestore.Update{
		{Path: "status", Value: status},
		{Path: "updatedAt", Value: time.Now()},
	}
	if errorMessage != "" {
		updates = append(updates, firestore.Update{Path: "errorMessage", Value: errorMessage})
	}

	_, err = doc.Ref.Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("ステータスの更新に失敗: %w", err)
	}

	return nil
}

// SaveResult は分析結果を保存し、ステータスをcompletedに更新します
func (r *firestoreAnalysisRepository) SaveResult(ctx context.Context, requestID uuid.UUID, result *service.AnalysisResult) error {
	// コレクショングループクエリで対象ドキュメントを検索
	iter := r.client.CollectionGroup("analysisRequests").Where("id", "==", requestID.String()).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return fmt.Errorf("リクエストが見つかりません: %s", requestID)
	}
	if err != nil {
		return fmt.Errorf("リクエストの取得に失敗: %w", err)
	}

	fsResult := &firestoreAnalysisResult{
		Foods:              result.Foods,
		TotalCalories:      result.TotalCalories,
		TotalProtein:       result.TotalProtein,
		TotalFat:           result.TotalFat,
		TotalCarbohydrates: result.TotalCarbohydrates,
	}

	_, err = doc.Ref.Update(ctx, []firestore.Update{
		{Path: "status", Value: StatusCompleted},
		{Path: "result", Value: fsResult},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("分析結果の保存に失敗: %w", err)
	}

	return nil
}

// GetResult は指定されたリクエストIDの分析結果を取得します（userIDでスコープ）
func (r *firestoreAnalysisRepository) GetResult(ctx context.Context, userID string, requestID uuid.UUID) (*service.AnalysisResult, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	// ユーザーのコレクションから直接取得
	doc, err := r.getUserAnalysisCollection(userID).Doc(requestID.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("結果が見つかりません: %s: %w", requestID, ErrNotFound)
		}
		return nil, fmt.Errorf("結果の取得に失敗: %w", err)
	}

	var fsDoc firestoreAnalysisDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	if fsDoc.Result == nil {
		return nil, fmt.Errorf("結果が見つかりません: %s: %w", requestID, ErrNotFound)
	}

	return &service.AnalysisResult{
		Foods:              fsDoc.Result.Foods,
		TotalCalories:      fsDoc.Result.TotalCalories,
		TotalProtein:       fsDoc.Result.TotalProtein,
		TotalFat:           fsDoc.Result.TotalFat,
		TotalCarbohydrates: fsDoc.Result.TotalCarbohydrates,
	}, nil
}

// GetPendingRequests はpending状態のリクエストを取得します
func (r *firestoreAnalysisRepository) GetPendingRequests(ctx context.Context, limit int) ([]AnalysisRequest, error) {
	// コレクショングループクエリで全ユーザーからpendingを検索
	iter := r.client.CollectionGroup("analysisRequests").
		Where("status", "==", string(StatusPending)).
		OrderBy("createdAt", firestore.Asc).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	var requests []AnalysisRequest
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("pending リクエストの取得に失敗: %w", err)
		}

		var fsDoc firestoreAnalysisDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		req, err := r.toAnalysisRequest(&fsDoc)
		if err != nil {
			return nil, fmt.Errorf("AnalysisRequestへの変換に失敗: %w", err)
		}
		requests = append(requests, *req)
	}

	return requests, nil
}

// GetHistoryList は履歴一覧を取得します（userIDでスコープ、ページネーション対応）
func (r *firestoreAnalysisRepository) GetHistoryList(ctx context.Context, userID string, page, limit int) ([]HistoryItem, int, error) {
	if userID == "" {
		return nil, 0, fmt.Errorf("userIDが必要です")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 総件数を取得（ユーザーのコレクションから）
	countIter := r.getUserAnalysisCollection(userID).
		Where("status", "==", string(StatusCompleted)).
		Documents(ctx)

	total := 0
	for {
		_, err := countIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			countIter.Stop()
			return nil, 0, fmt.Errorf("総件数の取得に失敗: %w", err)
		}
		total++
	}
	countIter.Stop()

	// ページネーション用のオフセットを計算
	offset := (page - 1) * limit

	// 履歴一覧を取得（ユーザーのコレクションから）
	iter := r.getUserAnalysisCollection(userID).
		Where("status", "==", string(StatusCompleted)).
		OrderBy("createdAt", firestore.Desc).
		Offset(offset).
		Limit(limit).
		Documents(ctx)
	defer iter.Stop()

	var items []HistoryItem
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("履歴一覧の取得に失敗: %w", err)
		}

		var fsDoc firestoreAnalysisDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, 0, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		item, err := r.toHistoryItem(&fsDoc)
		if err != nil {
			return nil, 0, fmt.Errorf("HistoryItemへの変換に失敗: %w", err)
		}
		items = append(items, *item)
	}

	return items, total, nil
}

// GetHistoryDetail は履歴詳細を取得します（userIDでスコープ）
func (r *firestoreAnalysisRepository) GetHistoryDetail(ctx context.Context, userID string, id uuid.UUID) (*HistoryDetail, error) {
	if userID == "" {
		return nil, fmt.Errorf("userIDが必要です")
	}

	// ユーザーのコレクションから直接取得
	doc, err := r.getUserAnalysisCollection(userID).Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("履歴が見つかりません: %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("履歴の取得に失敗: %w", err)
	}

	var fsDoc firestoreAnalysisDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return nil, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	// completed状態のみ返す
	if fsDoc.Status != StatusCompleted {
		return nil, fmt.Errorf("履歴が見つかりません: %s: %w", id, ErrNotFound)
	}

	return r.toHistoryDetail(&fsDoc)
}

// DeleteHistory は履歴を削除します（userIDでスコープ、関連する画像も含む）
func (r *firestoreAnalysisRepository) DeleteHistory(ctx context.Context, userID string, id uuid.UUID) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}

	// ユーザーのコレクションから直接取得
	docRef := r.getUserAnalysisCollection(userID).Doc(id.String())
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("履歴が見つかりません: %s: %w", id, ErrNotFound)
		}
		return fmt.Errorf("履歴の取得に失敗: %w", err)
	}

	var fsDoc firestoreAnalysisDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	// Cloud Storageから画像を削除（画像入力の場合のみ）
	// 注: Cloud Storage削除を先に実行し、孤立したFirestoreドキュメントより
	// 孤立した画像ファイルを防ぐ（UIの整合性を優先）
	if fsDoc.InputType == InputTypeImage && fsDoc.ImagePath != "" {
		if err := r.storageRepo.Delete(ctx, fsDoc.ImagePath); err != nil {
			log.Printf("Error: Cloud Storage画像の削除に失敗: %s: %v", fsDoc.ImagePath, err)
			return fmt.Errorf("画像の削除に失敗: %w", err)
		}
	}

	// ドキュメントを削除
	_, err = docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("履歴の削除に失敗: %w", err)
	}

	return nil
}

// GetDailyMeals は指定された日付の食事データを取得します（userIDでスコープ）
func (r *firestoreAnalysisRepository) GetDailyMeals(ctx context.Context, userID string, date string) (map[string][]HistoryDetail, DailyTotal, error) {
	if userID == "" {
		return nil, DailyTotal{}, fmt.Errorf("userIDが必要です")
	}

	mealDateTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, DailyTotal{}, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	// 日付の開始と終了を計算
	startOfDay := time.Date(mealDateTime.Year(), mealDateTime.Month(), mealDateTime.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// ユーザーのコレクションから対象日の食事を取得
	iter := r.getUserAnalysisCollection(userID).
		Where("mealDate", ">=", startOfDay).
		Where("mealDate", "<", endOfDay).
		Where("status", "==", string(StatusCompleted)).
		OrderBy("mealDate", firestore.Asc).
		Documents(ctx)
	defer iter.Stop()

	meals := map[string][]HistoryDetail{
		"breakfast": {},
		"lunch":     {},
		"dinner":    {},
		"snack":     {},
	}

	var dailyTotal DailyTotal

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, DailyTotal{}, fmt.Errorf("日次食事の取得に失敗: %w", err)
		}

		var fsDoc firestoreAnalysisDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return nil, DailyTotal{}, fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		detail, err := r.toHistoryDetail(&fsDoc)
		if err != nil {
			return nil, DailyTotal{}, fmt.Errorf("HistoryDetailへの変換に失敗: %w", err)
		}

		if fsDoc.MealType != "" {
			meals[fsDoc.MealType] = append(meals[fsDoc.MealType], *detail)
		}

		// 合計を加算
		dailyTotal.TotalCalories += detail.TotalCalories
		dailyTotal.TotalProtein += detail.TotalProtein
		dailyTotal.TotalFat += detail.TotalFat
		dailyTotal.TotalCarbohydrates += detail.TotalCarbohydrates
	}

	return meals, dailyTotal, nil
}

// CreateRequestFromMylist はマイリストからの食事記録を作成します
func (r *firestoreAnalysisRepository) CreateRequestFromMylist(ctx context.Context, inputText, mealType, mealDate string, userID *string, result *service.AnalysisResult) (uuid.UUID, error) {
	if userID == nil || *userID == "" {
		return uuid.Nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New()

	mealDateTime, err := time.Parse("2006-01-02", mealDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	// 既存のskipped記録を削除
	if err := r.deleteSkippedRecords(ctx, *userID, mealType, mealDateTime); err != nil {
		return uuid.Nil, err
	}

	doc := firestoreAnalysisDocument{
		ID:        id.String(),
		Status:    StatusCompleted,
		InputType: InputTypeMylist,
		InputText: inputText,
		MealType:  mealType,
		MealDate:  mealDateTime,
		CreatedAt: now,
		UpdatedAt: now,
		Result: &firestoreAnalysisResult{
			Foods:              result.Foods,
			TotalCalories:      result.TotalCalories,
			TotalProtein:       result.TotalProtein,
			TotalFat:           result.TotalFat,
			TotalCarbohydrates: result.TotalCarbohydrates,
		},
	}

	_, err = r.getUserAnalysisCollection(*userID).Doc(id.String()).Set(ctx, doc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("マイリストからのリクエスト作成に失敗: %w", err)
	}

	return id, nil
}

// CreateSkippedMeal は「食べなかった」記録を作成します
func (r *firestoreAnalysisRepository) CreateSkippedMeal(ctx context.Context, mealType, mealDate string, userID *string) (uuid.UUID, error) {
	if userID == nil || *userID == "" {
		return uuid.Nil, fmt.Errorf("userIDが必要です")
	}

	now := time.Now()
	id := uuid.New()

	mealDateTime, err := time.Parse("2006-01-02", mealDate)
	if err != nil {
		return uuid.Nil, fmt.Errorf("日付のパースに失敗: %w", err)
	}

	// 既存記録を削除（通常記録含む）
	if err := r.deleteExistingMealRecords(ctx, *userID, mealType, mealDateTime); err != nil {
		return uuid.Nil, err
	}

	doc := firestoreAnalysisDocument{
		ID:        id.String(),
		Status:    StatusCompleted,
		InputType: InputTypeSkipped,
		MealType:  mealType,
		MealDate:  mealDateTime,
		CreatedAt: now,
		UpdatedAt: now,
		Result: &firestoreAnalysisResult{
			Foods:              []gemini.NutritionInfo{},
			TotalCalories:      0,
			TotalProtein:       0,
			TotalFat:           0,
			TotalCarbohydrates: 0,
		},
	}

	_, err = r.getUserAnalysisCollection(*userID).Doc(id.String()).Set(ctx, doc)
	if err != nil {
		return uuid.Nil, fmt.Errorf("スキップ記録の作成に失敗: %w", err)
	}

	return id, nil
}

// UpdateResult は分析結果を更新します（userIDでスコープ）
func (r *firestoreAnalysisRepository) UpdateResult(ctx context.Context, userID string, historyID uuid.UUID, foods []gemini.NutritionInfo) error {
	if userID == "" {
		return fmt.Errorf("userIDが必要です")
	}

	// ユーザーのコレクションから直接取得
	docRef := r.getUserAnalysisCollection(userID).Doc(historyID.String())
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return fmt.Errorf("履歴が見つかりません: %s: %w", historyID, ErrNotFound)
		}
		return fmt.Errorf("履歴の取得に失敗: %w", err)
	}

	var fsDoc firestoreAnalysisDocument
	if err := doc.DataTo(&fsDoc); err != nil {
		return fmt.Errorf("ドキュメントのパースに失敗: %w", err)
	}

	// completed状態のみ更新可能
	if fsDoc.Status != StatusCompleted {
		return fmt.Errorf("履歴が見つかりません: %s: %w", historyID, ErrNotFound)
	}

	// 合計値を計算
	var totalCalories, totalProtein, totalFat, totalCarbohydrates float64
	for _, food := range foods {
		totalCalories += food.Calories
		totalProtein += food.Protein
		totalFat += food.Fat
		totalCarbohydrates += food.Carbohydrates
	}

	fsResult := &firestoreAnalysisResult{
		Foods:              foods,
		TotalCalories:      totalCalories,
		TotalProtein:       totalProtein,
		TotalFat:           totalFat,
		TotalCarbohydrates: totalCarbohydrates,
	}

	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "result", Value: fsResult},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("分析結果の更新に失敗: %w", err)
	}

	return nil
}

// deleteSkippedRecords は既存のskipped記録のみを削除します
func (r *firestoreAnalysisRepository) deleteSkippedRecords(ctx context.Context, userID, mealType string, mealDate time.Time) error {
	startOfDay := time.Date(mealDate.Year(), mealDate.Month(), mealDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	iter := r.getUserAnalysisCollection(userID).
		Where("mealType", "==", mealType).
		Where("mealDate", ">=", startOfDay).
		Where("mealDate", "<", endOfDay).
		Where("inputType", "==", string(InputTypeSkipped)).
		Documents(ctx)
	defer iter.Stop()

	// BulkWriterを使用（Batchは非推奨）
	bw := r.client.BulkWriter(ctx)
	var deleteJobs []*firestore.BulkWriterJob

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("既存skipped記録の確認に失敗: %w", err)
		}
		job, err := bw.Delete(doc.Ref)
		if err != nil {
			return fmt.Errorf("skipped記録の削除ジョブ追加に失敗: %w", err)
		}
		deleteJobs = append(deleteJobs, job)
	}

	bw.Flush()
	bw.End()

	// 各削除ジョブの結果を確認
	for _, job := range deleteJobs {
		if _, err := job.Results(); err != nil {
			return fmt.Errorf("skipped記録の削除に失敗: %w", err)
		}
	}

	return nil
}

// deleteExistingMealRecords は既存の全食事記録を削除します（画像も含む）
func (r *firestoreAnalysisRepository) deleteExistingMealRecords(ctx context.Context, userID, mealType string, mealDate time.Time) error {
	startOfDay := time.Date(mealDate.Year(), mealDate.Month(), mealDate.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	iter := r.getUserAnalysisCollection(userID).
		Where("mealType", "==", mealType).
		Where("mealDate", ">=", startOfDay).
		Where("mealDate", "<", endOfDay).
		Documents(ctx)
	defer iter.Stop()

	var imagePaths []string
	var docRefs []*firestore.DocumentRef

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("既存記録の確認に失敗: %w", err)
		}

		var fsDoc firestoreAnalysisDocument
		if err := doc.DataTo(&fsDoc); err != nil {
			return fmt.Errorf("ドキュメントのパースに失敗: %w", err)
		}

		if fsDoc.InputType == InputTypeImage && fsDoc.ImagePath != "" {
			imagePaths = append(imagePaths, fsDoc.ImagePath)
		}

		docRefs = append(docRefs, doc.Ref)
	}

	// Cloud Storageから画像を先に削除
	// 注: 画像削除を先に実行し、孤立した画像ファイルを防ぐ
	for _, path := range imagePaths {
		if err := r.storageRepo.Delete(ctx, path); err != nil {
			log.Printf("Error: Cloud Storage画像の削除に失敗: %s: %v", path, err)
			return fmt.Errorf("画像の削除に失敗: %w", err)
		}
	}

	// BulkWriterでFirestoreドキュメントを削除
	bw := r.client.BulkWriter(ctx)
	var deleteJobs []*firestore.BulkWriterJob

	for _, ref := range docRefs {
		job, err := bw.Delete(ref)
		if err != nil {
			return fmt.Errorf("既存記録の削除ジョブ追加に失敗: %w", err)
		}
		deleteJobs = append(deleteJobs, job)
	}

	bw.Flush()
	bw.End()

	// 各削除ジョブの結果を確認
	for _, job := range deleteJobs {
		if _, err := job.Results(); err != nil {
			return fmt.Errorf("既存記録の削除に失敗: %w", err)
		}
	}

	return nil
}

// toAnalysisRequest はFirestoreドキュメントをAnalysisRequestに変換
func (r *firestoreAnalysisRepository) toAnalysisRequest(doc *firestoreAnalysisDocument) (*AnalysisRequest, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("不正なドキュメントID: %s: %w", doc.ID, err)
	}
	return &AnalysisRequest{
		ID:           id,
		Status:       doc.Status,
		InputType:    doc.InputType,
		ImagePath:    doc.ImagePath,
		InputText:    doc.InputText,
		ErrorMessage: doc.ErrorMessage,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}, nil
}

// toHistoryItem はFirestoreドキュメントをHistoryItemに変換
func (r *firestoreAnalysisRepository) toHistoryItem(doc *firestoreAnalysisDocument) (*HistoryItem, error) {
	id, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, fmt.Errorf("不正なドキュメントID: %s: %w", doc.ID, err)
	}
	item := &HistoryItem{
		ID:        id,
		InputType: doc.InputType,
		ImagePath: doc.ImagePath,
		InputText: doc.InputText,
		CreatedAt: doc.CreatedAt,
		MealType:  doc.MealType,
		MealDate:  doc.MealDate,
	}

	if doc.Result != nil {
		item.TotalCalories = doc.Result.TotalCalories
		item.TotalProtein = doc.Result.TotalProtein
		item.TotalFat = doc.Result.TotalFat
		item.TotalCarbohydrates = doc.Result.TotalCarbohydrates
	}

	return item, nil
}

// toHistoryDetail はFirestoreドキュメントをHistoryDetailに変換
func (r *firestoreAnalysisRepository) toHistoryDetail(doc *firestoreAnalysisDocument) (*HistoryDetail, error) {
	item, err := r.toHistoryItem(doc)
	if err != nil {
		return nil, err
	}
	detail := &HistoryDetail{
		HistoryItem: *item,
	}

	if doc.Result != nil {
		detail.Foods = doc.Result.Foods
	}

	return detail, nil
}

