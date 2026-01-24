.PHONY: help setup clean test lint build run db-up db-down db-seed db-clean deploy

# デフォルトターゲット
.DEFAULT_GOAL := help

# 変数
BACKEND_DIR := backend
DOCKER_COMPOSE := docker-compose

##@ 基本

help: ## ターゲット一覧を表示
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

setup: ## 初期セットアップ（依存関係のダウンロード）
	@echo "=== Backend依存関係をダウンロード ==="
	cd $(BACKEND_DIR) && go mod download
	@echo "セットアップ完了"

clean: ## ビルド成果物を削除
	@echo "=== クリーンアップ ==="
	rm -f $(BACKEND_DIR)/bin/*
	@echo "クリーンアップ完了"

##@ Backend

test: ## Goテストを実行
	@echo "=== Goテスト実行 ==="
	cd $(BACKEND_DIR) && go test ./...

lint: ## Goリント（golangci-lint）を実行
	@echo "=== Goリント実行 ==="
	cd $(BACKEND_DIR) && golangci-lint run ./...

build: ## Goバイナリをビルド
	@echo "=== Goビルド ==="
	cd $(BACKEND_DIR) && mkdir -p bin && go build -o bin/server ./cmd/server

run: ## バックエンドサーバーを起動
	@echo "=== サーバー起動 ==="
	cd $(BACKEND_DIR) && go run ./cmd/server

##@ Docker/DB

db-up: ## PostgreSQLを起動
	@echo "=== PostgreSQL起動 ==="
	$(DOCKER_COMPOSE) up -d

db-down: ## PostgreSQLを停止
	@echo "=== PostgreSQL停止 ==="
	$(DOCKER_COMPOSE) down

db-seed: ## テストデータを投入
	@echo "=== シードデータ投入 ==="
	@cd $(BACKEND_DIR) && if [ -f .env ]; then export $$(cat .env | grep -v '^#' | xargs) && go run ./cmd/seed -verbose; else go run ./cmd/seed -verbose; fi

db-clean: ## DBをリセット（ボリューム削除+再起動）
	@echo "=== DBリセット ==="
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d
	@echo "DBの起動を待機中..."
	@sleep 3
	@echo "DBリセット完了"

##@ デプロイ

deploy: ## 本番環境にデプロイ（deploy.sh実行）
	@echo "=== デプロイ実行 ==="
	./deploy.sh
