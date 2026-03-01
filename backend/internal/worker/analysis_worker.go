package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ryosuke-horie/uchikomi/backend/internal/repository"
	"github.com/ryosuke-horie/uchikomi/backend/internal/util"
)

// FoodService は食品分析サービスのインターフェース
type FoodService interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*repository.AnalysisResult, error)
	AnalyzeFoodText(ctx context.Context, inputText string) (*repository.AnalysisResult, error)
}

// AnalysisWorker は非同期分析ワーカー
type AnalysisWorker struct {
	foodService FoodService
	repository  repository.AnalysisRepository
	interval    time.Duration
}

// NewAnalysisWorker は新しいAnalysisWorkerを作成
func NewAnalysisWorker(
	foodService FoodService,
	repository repository.AnalysisRepository,
	interval time.Duration,
) *AnalysisWorker {
	return &AnalysisWorker{
		foodService: foodService,
		repository:  repository,
		interval:    interval,
	}
}

// Start はワーカーを起動し、context.Doneまで動作を続ける
func (w *AnalysisWorker) Start(ctx context.Context) {
	log.Printf("Analysis worker started with interval: %v", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// 起動直後に1回実行
	if err := w.processPendingRequests(ctx); err != nil {
		log.Printf("Error processing pending requests: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("Analysis worker stopped: %v", ctx.Err())
			return

		case <-ticker.C:
			if err := w.processPendingRequests(ctx); err != nil {
				log.Printf("Error processing pending requests: %v", err)
			}
		}
	}
}

// processPendingRequests はpending状態のリクエストを処理
func (w *AnalysisWorker) processPendingRequests(ctx context.Context) error {
	// pending状態のリクエストを取得（最大10件）
	requests, err := w.repository.GetPendingRequests(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to get pending requests: %w", err)
	}

	if len(requests) == 0 {
		log.Printf("No pending requests found")
		return nil
	}

	log.Printf("Found %d pending requests", len(requests))

	// 各リクエストを処理（逐次処理でシンプルに）
	for i := range requests {
		if err := w.processRequest(ctx, &requests[i]); err != nil {
			log.Printf("Error processing request %s: %v", requests[i].ID, err)
			// 1つのリクエストの失敗で全体を止めない
			continue
		}
	}

	return nil
}

// processRequest は個別のリクエストを処理
func (w *AnalysisWorker) processRequest(ctx context.Context, request *repository.AnalysisRequest) error {
	log.Printf("Processing request: %s (input_type: %s)", request.ID, request.InputType)

	// 1. ステータスをprocessingに更新
	if err := w.repository.UpdateStatus(ctx, request.ID, repository.StatusProcessing, ""); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	log.Printf("Status updated to processing for request: %s", request.ID)

	// 2. input_typeに応じて処理を分岐
	var result *repository.AnalysisResult
	var err error

	switch request.InputType {
	case repository.InputTypeImage:
		log.Printf("Analyzing image: %s", request.ImagePath)
		result, err = w.foodService.AnalyzeFoodImage(ctx, request.ImagePath)
	case repository.InputTypeText:
		log.Printf("Analyzing text: %s", util.TruncateForLog(request.InputText, 50))
		result, err = w.foodService.AnalyzeFoodText(ctx, request.InputText)
	default:
		err = fmt.Errorf("不明な入力タイプ: %s", request.InputType)
	}

	if err != nil {
		// 分析エラー時はfailedステータスに更新
		errorMessage := fmt.Sprintf("分析エラー: %v", err)
		if updateErr := w.repository.UpdateStatus(ctx, request.ID, repository.StatusFailed, errorMessage); updateErr != nil {
			log.Printf("CRITICAL: Failed to update status to failed for request %s: %v (original error: %v)", request.ID, updateErr, err)
			return fmt.Errorf("failed to update status after analysis error: %w (original: %v)", updateErr, err)
		}
		log.Printf("Analysis failed for request %s: %v", request.ID, err)
		return nil // エラーをハンドリング済みなのでnilを返す
	}

	log.Printf("Analysis completed for request: %s (total_calories: %.2f)", request.ID, result.TotalCalories)

	// 3. SaveResultで結果保存（ステータスもcompletedに更新）
	if err := w.repository.SaveResult(ctx, request.ID, result); err != nil {
		// 保存エラー時はfailedステータスに更新
		errorMessage := fmt.Sprintf("結果保存エラー: %v", err)
		if updateErr := w.repository.UpdateStatus(ctx, request.ID, repository.StatusFailed, errorMessage); updateErr != nil {
			log.Printf("CRITICAL: Failed to update status to failed for request %s: %v (original error: %v)", request.ID, updateErr, err)
			return fmt.Errorf("failed to update status after save error: %w (original: %v)", updateErr, err)
		}
		return fmt.Errorf("failed to save result: %w", err)
	}

	log.Printf("Result saved successfully for request: %s", request.ID)

	return nil
}
