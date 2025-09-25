# Quivra Backend

Quivra は複数プレイヤーが同時に対戦できるクイズプラットフォームのバックエンド実装です。

## 技術スタック

- **言語**: Go 1.21
- **フレームワーク**: Gin
- **WebSocket**: Gorilla WebSocket
- **データベース**: MySQL 8.0
- **キャッシュ**: Go のメモリキャッシュ（sync.Map + カスタムキャッシュ）

## 機能

### Phase 1: 基本機能 ✅

- [x] WebSocket 接続管理
- [x] ルーム機能（作成・参加・状態管理）
- [x] 早押し機能（競合状態回避、タイムスタンプ管理）
- [x] HTTP API（ルーム、問題管理）

### Phase 2: ゲーム機能

- [ ] 問題管理（CRUD 操作）
- [ ] スコア管理
- [ ] ランキング表示

### Phase 3: 拡張機能

- [ ] カテゴリ管理
- [ ] 難易度設定
- [ ] 統計機能

## セットアップ

### 1. 環境変数の設定

```bash
cp env.example .env
```

`.env` ファイルを編集して、データベース設定を調整してください。

### 2. Docker Compose で起動

**開発環境（推奨）:**

```bash
docker-compose up -d
```

**本番環境:**

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 3. 手動で起動する場合

```bash
# 依存関係のインストール
go mod tidy

# アプリケーションの起動
go run main.go
```

## API エンドポイント

### HTTP API

#### ルーム関連

- `POST /api/rooms` - ルーム作成
- `GET /api/rooms/{roomId}` - ルーム情報取得
- `POST /api/rooms/join` - ルーム参加

#### 問題関連

- `POST /api/questions` - 問題作成
- `GET /api/questions` - 問題一覧取得
- `GET /api/questions/{id}` - 問題取得

### WebSocket

#### エンドポイント

- `ws://localhost:8080/ws`

#### イベント

**クライアント → サーバー**

- `join-room` - ルーム参加
- `buzz-in` - 早押しボタン
- `submit-answer` - 回答送信
- `start-game` - ゲーム開始

**サーバー → クライアント**

- `room-updated` - ルーム状態更新
- `buzz-result` - 早押し結果
- `question-result` - 回答結果

## データベース設計

### テーブル構成

1. **rooms** - ルーム情報
2. **players** - プレイヤー情報
3. **questions** - 問題情報
4. **game_sessions** - ゲームセッション情報

詳細は `database/migrations.sql` を参照してください。

## 早押し機能の実装詳細

### 競合状態の回避

早押し機能では、複数のプレイヤーが同時にボタンを押した場合の競合状態を回避するため、サーバー側で厳密な時刻管理を行います。

```go
type BuzzManager struct {
    mu          sync.RWMutex
    buzzStates  map[string]*BuzzState // roomId -> BuzzState
}
```

### タイムスタンプ管理

- サーバー側で正確な時刻を管理
- 早押しの判定はサーバー側で行う
- クライアント側の時刻は参考程度

## 開発

### プロジェクト構造

```
quivra-backend/
├── cmd/                 # アプリケーションエントリーポイント
├── config/              # 設定管理
├── database/            # データベース関連
│   ├── migrations.sql   # マイグレーション
│   └── sample_data.sql  # サンプルデータ
├── handlers/            # HTTP ハンドラー
├── models/              # データモデル
├── services/            # ビジネスロジック
├── websocket/           # WebSocket 関連
├── main.go             # メインアプリケーション
├── docker-compose.yml  # Docker Compose 設定
└── Dockerfile         # Docker 設定
```

### テスト

```bash
# 単体テスト
go test ./...

# 統合テスト
go test -tags=integration ./...
```

## デプロイメント

### 本番環境

```bash
# 本番環境での起動
docker-compose -f docker-compose.prod.yml up -d
```

- **サーバー**: EC2
- **データベース**: RDS MySQL
- **ロードバランサー**: ALB
- **SSL**: Let's Encrypt

### 監視

- **ログ**: CloudWatch Logs
- **メトリクス**: CloudWatch Metrics
- **アラート**: 接続数、レスポンス時間

## ライセンス

MIT License
