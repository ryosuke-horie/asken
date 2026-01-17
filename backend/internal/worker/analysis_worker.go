package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ryosuke-horie/asken/backend/internal/repository"
	"github.com/ryosuke-horie/asken/backend/internal/service"
)

// FoodService は食品分析サービスのインターフェース
type FoodService interface {
	AnalyzeFoodImage(ctx context.Context, imagePath string) (*service.AnalysisResult, error)
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
	for _, request := range requests {
		if err := w.processRequest(ctx, request); err != nil {
			log.Printf("Error processing request %s: %v", request.ID, err)
			// 1つのリクエストの失敗で全体を止めない
			continue
		}
	}

	return nil
}

// processRequest は個別のリクエストを処理
func (w *AnalysisWorker) processRequest(ctx context.Context, request repository.AnalysisRequest) error {
	log.Printf("Processing request: %s (image: %s)", request.ID, request.ImagePath)

	// 1. ステータスをprocessingに更新
	if err := w.repository.UpdateStatus(ctx, request.ID, repository.StatusProcessing, ""); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	log.Printf("Status updated to processing for request: %s", request.ID)

	// 2. FoodService.AnalyzeFoodImage()を呼び出し
	result, err := w.foodService.AnalyzeFoodImage(ctx, request.ImagePath)
	if err != nil {
		// 分析エラー時はfailedステータスに更新
		errorMessage := fmt.Sprintf("分析エラー: %v", err)
		if updateErr := w.repository.UpdateStatus(ctx, request.ID, repository.StatusFailed, errorMessage); updateErr != nil {
			log.Printf("Failed to update status to failed: %v", updateErr)
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
			log.Printf("Failed to update status to failed: %v", updateErr)
		}
		return fmt.Errorf("failed to save result: %w", err)
	}

	log.Printf("Result saved successfully for request: %s", request.ID)

	// 4. 画像ファイル削除（成功時のみ）
	if err := os.Remove(request.ImagePath); err != nil {
		log.Printf("Warning: Failed to remove image file %s: %v", request.ImagePath, err)
		// ファイル削除失敗は致命的ではないため、続行
	} else {
		log.Printf("Image file removed: %s", request.ImagePath)
	}

	return nil
}
