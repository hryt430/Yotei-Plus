# Task Management API

Go言語で構築されたタスク管理システムのREST APIです。認証、通知、タスク管理機能を提供します。

## 🚀 特徴

- **認証・認可**: JWT ベースの認証システム
- **タスク管理**: CRUD操作、フィルタリング、検索機能
- **通知システム**: アプリ内通知、LINE通知、Webhook対応
- **セキュリティ**: CORS、CSRF、レート制限対応
- **リアルタイム通信**: WebSocket対応
- **高可用性**: Redis キャッシュ、データベース接続プール

## 🏗️ アーキテクチャ

```
├── cmd/
│   └── server/          # アプリケーションエントリーポイント
├── config/              # 設定管理
├── internal/
│   ├── common/         # 共通コンポーネント
│   │   ├── events/     # イベント定義
│   │   ├── infrastructure/ # インフラストラクチャ層
│   │   └── middleware/ # 共通ミドルウェア
│   ├── modules/        # ビジネスロジックモジュール
│   │   ├── auth/       # 認証モジュール
│   │   ├── notification/ # 通知モジュール
│   │   └── task/       # タスクモジュール
│   └── server/         # サーバー設定
└── pkg/                # 共有パッケージ
    ├── logger/         # ログ機能
    ├── token/          # JWT管理
    └── utils/          # ユーティリティ
```

## 🛠️ 技術スタック

- **言語**: Go 1.21
- **フレームワーク**: Gin
- **データベース**: MySQL 8.0
- **キャッシュ**: Redis 7
- **認証**: JWT
- **ログ**: Zap
- **コンテナ**: Docker & Docker Compose

## 📋 前提条件

- Go 1.21以上
- Docker & Docker Compose
- Make (オプション)

## 🔧 セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/hryt430/task-management-api.git
cd task-management-api
```

### 2. 環境設定

```bash
# 設定ファイルのコピー
cp .env.example .env

# 必要に応じて .env ファイルを編集
vi .env
```

### 3. 依存関係のインストール

```bash
# Go modules
go mod download

# または Makeを使用
make deps
```

### 4. データベースとRedisの起動

```bash
# Docker Composeで起動
docker-compose up -d mysql redis

# 管理ツールも含めて起動
docker-compose up -d
```

### 5. アプリケーションの起動

```bash
# 開発モード（ホットリロード）
make dev

# または通常起動
make run

# またはDocker
make docker-run
```

## 🚀 使用方法

### API エンドポイント

#### 認証
- `POST /api/v1/auth/register` - ユーザー登録
- `POST /api/v1/auth/login` - ログイン
- `POST /api/v1/auth/refresh-token` - トークン更新
- `POST /api/v1/auth/logout` - ログアウト
- `GET /api/v1/auth/me` - ユーザー情報取得

#### タスク
- `GET /api/v1/tasks` - タスク一覧
- `POST /api/v1/tasks` - タスク作成
- `GET /api/v1/tasks/:id` - タスク取得
- `PUT /api/v1/tasks/:id` - タスク更新
- `DELETE /api/v1/tasks/:id` - タスク削除
- `PUT /api/v1/tasks/:id/assign` - タスク割り当て
- `PUT /api/v1/tasks/:id/status` - ステータス変更
- `GET /api/v1/tasks/search` - タスク検索
- `GET /api/v1/tasks/my` - 自分のタスク
- `GET /api/v1/tasks/overdue` - 期限切れタスク

#### 通知
- `GET /api/v1/notifications` - 通知一覧
- `POST /api/v1/notifications` - 通知作成
- `GET /api/v1/notifications/:id` - 通知取得
- `PUT /api/v1/notifications/:id/read` - 既読マーク
- `GET /api/v1/notifications/user/:user_id/unread/count` - 未読数

### 認証の使用例

```bash
# ユーザー登録
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "password123"
  }'

# ログイン
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# タスク作成（認証が必要）
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "title": "新しいタスク",
    "description": "タスクの説明",
    "priority": "HIGH"
  }'
```

## 🧪 テスト

```bash
# すべてのテストを実行
make test

# カバレッジ付きテスト
make test-coverage

# ベンチマークテスト
make benchmark
```

## 📦 ビルド・デプロイ

```bash
# 開発用ビルド
make build

# 本番用ビルド
make build-prod

# Dockerイメージビルド
make docker-build

# 本番環境でのデプロイ
make docker-prod
```

## 🔍 開発ツール

### 管理画面
- **phpMyAdmin**: http://localhost:8081 (MySQL管理)
- **Redis Commander**: http://localhost:8082 (Redis管理)

### ログとモニタリング
- アプリケーションログ: JSON形式でコンソール出力
- ヘルスチェック: `GET /health`

## 🛡️ セキュリティ

- JWT による認証・認可
- CORS 設定
- CSRF 保護（本番環境で有効）
- セキュリティヘッダー設定
- レート制限
- SQL インジェクション対策

## ⚙️ 設定

主要な環境変数：

```bash
# アプリケーション
ENVIRONMENT=development
SERVER_PORT=8080

# データベース
DB_HOST=localhost
DB_NAME=task_management
DB_USER=root
DB_PASSWORD=password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET_KEY=your-secret-key
JWT_ACCESS_TOKEN_DURATION=1h
JWT_REFRESH_TOKEN_DURATION=168h

# 外部サービス
LINE_CHANNEL_TOKEN=your-line-token
WEBHOOK_URL=https://your-webhook.com
```

## 🤝 開発に参加

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 ライセンス

このプロジェクトは MIT ライセンスのもとで公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 📞 サポート

質問や問題がある場合は、GitHub Issues を作成してください。

---

🚀 **Happy Coding!** 🚀