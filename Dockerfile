# マルチステージビルド
FROM golang:1.24-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# 必要なパッケージをインストール
RUN apk add --no-cache git

# go.modとgo.sumをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 本番用イメージ
FROM alpine:latest

# セキュリティアップデートとCA証明書をインストール
RUN apk --no-cache add ca-certificates tzdata

# タイムゾーンを設定
ENV TZ=Asia/Tokyo

# 作業ディレクトリを設定
WORKDIR /root/

# ビルドしたバイナリをコピー
COPY --from=builder /app/main .

# ポートを公開
EXPOSE 8080

# ヘルスチェック
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# アプリケーションを実行
CMD ["./main"]