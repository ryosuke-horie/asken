# =============================================================================
# マルチステージビルド: Goバックエンド用Dockerfile
# =============================================================================

# -----------------------------------------------------------------------------
# ビルドステージ
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

# ビルドに必要なパッケージをインストール
RUN apk add --no-cache ca-certificates git

WORKDIR /app

# 依存関係を先にコピー（キャッシュ効率化）
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# 静的リンクでバイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server

# -----------------------------------------------------------------------------
# 実行ステージ
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

# ビルドステージからバイナリとCA証明書をコピー（非rootユーザーに所有権を設定）
COPY --from=builder --chown=nonroot:nonroot /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chown=nonroot:nonroot /app/server /server

# 非rootユーザーで実行（distroless/nonrootはUID 65532）
USER nonroot:nonroot

# Cloud Runが使用するポート
EXPOSE 8080

# エントリポイント
ENTRYPOINT ["/server"]
