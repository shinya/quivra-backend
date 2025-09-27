# Quivra Backend

Quivra は複数プレイヤーが同時に対戦できるクイズプラットフォームのバックエンド実装です。リアルタイム WebSocket 通信と回答キュー管理システムを備えた、本格的なクイズゲームプラットフォームです。

## 🚀 技術スタック

- **言語**: Go 1.21
- **フレームワーク**: Gin
- **WebSocket**: Gorilla WebSocket
- **データベース**: MySQL 8.0 (UTF-8MB4 対応)
- **キャッシュ**: Go のメモリキャッシュ（sync.Map + カスタムキャッシュ）
- **コンテナ**: Docker & Docker Compose

## ✨ 主要機能

### 🏠 ルーム管理

- **公開/非公開ルーム**: ルーム作成時に公開設定を選択可能
- **ルーム管理者機能**: 作成者が自動的に管理者権限を取得
- **公開ルーム一覧**: 非公開ルームを除いた公開ルームのみ表示
- **ルーム参加**: ルーム ID 指定による参加（公開・非公開問わず）

### 🎮 ゲーム機能

- **回答キューシステム**: 早押し順序を厳密に管理
- **管理者による判定**: 正解・不正解のジャッジ機能
- **ポイントシステム**: 正解時の自動ポイント付与
- **ランキング表示**: リアルタイムスコア管理

### 🔐 権限管理

- **管理者権限**: 回答判定、キューリセット、ゲーム終了
- **参加者権限**: 早押しボタン、回答送信
- **権限チェック**: 全操作で適切な権限確認

## 🛠 セットアップ

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

## 📡 API エンドポイント

### HTTP API

#### ルーム関連

| メソッド | エンドポイント                | 説明                 | リクエストボディ                                                      |
| -------- | ----------------------------- | -------------------- | --------------------------------------------------------------------- |
| `POST`   | `/api/rooms`                  | ルーム作成           | `{"name": "ルーム名", "is_public": true, "creator_name": "作成者名"}` |
| `GET`    | `/api/rooms`                  | 公開ルーム一覧取得   | -                                                                     |
| `GET`    | `/api/rooms/{roomId}`         | ルーム情報取得       | -                                                                     |
| `GET`    | `/api/rooms/{roomId}/ranking` | ルームランキング取得 | -                                                                     |
| `POST`   | `/api/rooms/join`             | ルーム参加           | `{"roomId": "ルームID", "playerName": "プレイヤー名"}`                |

#### 問題関連

| メソッド | エンドポイント        | 説明         | リクエストボディ                                                                                       |
| -------- | --------------------- | ------------ | ------------------------------------------------------------------------------------------------------ |
| `POST`   | `/api/questions`      | 問題作成     | `{"question": "問題文", "answer": "答え", "category": "カテゴリ", "difficulty": "easy\|medium\|hard"}` |
| `GET`    | `/api/questions`      | 問題一覧取得 | -                                                                                                      |
| `GET`    | `/api/questions/{id}` | 問題取得     | -                                                                                                      |

### WebSocket イベント

#### エンドポイント

```
ws://localhost:8080/ws
```

#### クライアント → サーバー

| イベント        | 説明                         | データ                                                                |
| --------------- | ---------------------------- | --------------------------------------------------------------------- |
| `join-room`     | ルーム参加                   | `{"roomId": "ルームID", "playerName": "プレイヤー名"}`                |
| `buzz-in`       | 早押しボタン                 | `{"roomId": "ルームID"}`                                              |
| `submit-answer` | 回答送信                     | `{"roomId": "ルームID", "answer": "回答"}`                            |
| `start-game`    | ゲーム開始                   | `{"roomId": "ルームID"}`                                              |
| `judge-answer`  | 回答判定（管理者のみ）       | `{"roomId": "ルームID", "playerId": "プレイヤーID", "correct": true}` |
| `reset-queue`   | キューリセット（管理者のみ） | `{"roomId": "ルームID"}`                                              |
| `end-game`      | ゲーム終了（管理者のみ）     | `{"roomId": "ルームID"}`                                              |

#### サーバー → クライアント

| イベント        | 説明               | データ                                                                           |
| --------------- | ------------------ | -------------------------------------------------------------------------------- |
| `room-updated`  | ルーム状態更新     | `{"players": [...], "gameState": "waiting\|playing\|finished", "canBuzz": true}` |
| `queue-updated` | 回答キュー更新     | `{"queue": [{"player_id": "ID", "name": "名前", "buzzed_at": "時刻"}]}`          |
| `judge-result`  | 判定結果           | `{"correct": true, "player_id": "プレイヤーID"}`                                 |
| `queue-reset`   | キューリセット完了 | `{"message": "Queue has been reset"}`                                            |
| `game-ended`    | ゲーム終了         | `{"ranking": [{"player_id": "ID", "name": "名前", "score": 100, "rank": 1}]}`    |
| `success`       | 成功メッセージ     | `{"message": "メッセージ", "data": {...}}`                                       |
| `error`         | エラーメッセージ   | `{"message": "エラーメッセージ"}`                                                |

## 🗄 データベース設計

### テーブル構成

#### 1. **rooms** - ルーム情報

```sql
CREATE TABLE rooms (
    id VARCHAR(10) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('waiting', 'playing', 'finished') DEFAULT 'waiting',
    is_public BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(36) NOT NULL
);
```

#### 2. **players** - プレイヤー情報

```sql
CREATE TABLE players (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(10) NOT NULL,
    name VARCHAR(50) NOT NULL,
    score INT DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_admin BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);
```

#### 3. **questions** - 問題情報

```sql
CREATE TABLE questions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    question TEXT NOT NULL,
    answer VARCHAR(255) NOT NULL,
    category VARCHAR(50) DEFAULT 'general',
    difficulty ENUM('easy', 'medium', 'hard') DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

#### 4. **game_sessions** - ゲームセッション情報

```sql
CREATE TABLE game_sessions (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(10) NOT NULL,
    question_id INT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,
    status ENUM('waiting', 'question', 'buzzed', 'answered', 'finished') DEFAULT 'waiting',
    buzzed_player_id VARCHAR(36) NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE SET NULL,
    FOREIGN KEY (buzzed_player_id) REFERENCES players(id) ON DELETE SET NULL
);
```

#### 5. **buzz_queue** - 回答キュー管理

```sql
CREATE TABLE buzz_queue (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(10) NOT NULL,
    player_id VARCHAR(36) NOT NULL,
    buzzed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);
```

## 🔧 技術実装詳細

### 回答キューシステム

早押し機能では、複数のプレイヤーが同時にボタンを押した場合の競合状態を回避するため、サーバー側で厳密な時刻管理を行います。

```go
type BuzzQueueService struct {
    db *database.DB
}

// プレイヤーを回答キューに追加
func (bqs *BuzzQueueService) AddToQueue(roomID, playerID string) error {
    // 既にキューにいるかチェック
    exists, err := bqs.IsPlayerInQueue(roomID, playerID)
    if err != nil || exists {
        return fmt.Errorf("player already in queue")
    }

    // キューに追加
    queueID := generateQueueID()
    query := `INSERT INTO buzz_queue (id, room_id, player_id, buzzed_at, is_active) VALUES (?, ?, ?, ?, TRUE)`
    _, err = bqs.db.Exec(query, queueID, roomID, playerID, time.Now())
    return err
}
```

### 管理者権限チェック

```go
// 管理者権限チェック
func (rs *RoomService) IsPlayerAdmin(roomID, playerID string) (bool, error) {
    query := `SELECT is_admin FROM players WHERE room_id = ? AND id = ?`
    var isAdmin bool
    err := rs.db.QueryRow(query, roomID, playerID).Scan(&isAdmin)
    if err != nil {
        return false, fmt.Errorf("player not found")
    }
    return isAdmin, nil
}
```

### ランキングシステム

```go
// ルームのランキングを取得
func (rs *RoomService) GetRoomRanking(roomID string) ([]models.RoomRanking, error) {
    query := `SELECT id, name, score FROM players WHERE room_id = ? ORDER BY score DESC, joined_at ASC`
    rows, err := rs.db.Query(query, roomID)
    // ... ランキング処理
}
```

## 🏗 プロジェクト構造

```
quivra-backend/
├── cmd/                    # アプリケーションエントリーポイント
├── config/                 # 設定管理
├── database/               # データベース関連
│   ├── migrations.sql      # マイグレーション
│   ├── sample_data.sql    # サンプルデータ
│   └── database.go        # データベース接続
├── handlers/               # HTTP ハンドラー
│   ├── room_handler.go    # ルーム関連API
│   └── question_handler.go # 問題関連API
├── models/                 # データモデル
│   ├── room.go            # ルーム・プレイヤーモデル
│   └── websocket.go       # WebSocketメッセージモデル
├── services/              # ビジネスロジック
│   ├── room_service.go    # ルーム管理
│   ├── question_service.go # 問題管理
│   ├── game_service.go   # ゲーム管理
│   ├── buzz_manager.go   # 早押し管理
│   └── buzz_queue_service.go # 回答キュー管理
├── websocket/             # WebSocket 関連
│   ├── connection.go      # 接続管理
│   ├── handler.go         # イベントハンドラー
│   └── hub.go            # ハブ管理
├── main.go               # メインアプリケーション
├── docker-compose.yml    # 開発環境Docker設定
├── docker-compose.prod.yml # 本番環境Docker設定
└── Dockerfile           # Docker 設定
```

## 🧪 テスト

### 単体テスト

```bash
# 全テスト実行
go test ./...

# 特定パッケージのテスト
go test ./services/...

# カバレッジ付きテスト
go test -cover ./...
```

### 統合テスト

```bash
# 統合テスト実行
go test -tags=integration ./...
```

### WebSocket テスト

```bash
# テスト用HTMLファイルを使用
open frontend_test.html
```

## 🚀 デプロイメント

### 本番環境

```bash
# 本番環境での起動
docker-compose -f docker-compose.prod.yml up -d
```

### 推奨インフラ構成

- **サーバー**: EC2 (t3.medium 以上)
- **データベース**: RDS MySQL 8.0
- **ロードバランサー**: ALB
- **SSL**: Let's Encrypt
- **監視**: CloudWatch Logs & Metrics

### 環境変数

| 変数名        | 説明                   | デフォルト値 |
| ------------- | ---------------------- | ------------ |
| `DB_HOST`     | データベースホスト     | `localhost`  |
| `DB_PORT`     | データベースポート     | `3306`       |
| `DB_USER`     | データベースユーザー   | `quivra`     |
| `DB_PASSWORD` | データベースパスワード | `password`   |
| `DB_NAME`     | データベース名         | `quivra`     |
| `PORT`        | アプリケーションポート | `8080`       |

## 📊 監視・ログ

### ログレベル

- **INFO**: 一般的な操作ログ
- **WARN**: 警告レベルの問題
- **ERROR**: エラーレベルの問題

### 監視項目

- **接続数**: アクティブな WebSocket 接続数
- **レスポンス時間**: API 応答時間
- **エラー率**: エラー発生率
- **データベース接続**: DB 接続プール状態

## 🔒 セキュリティ

### CORS 設定

```go
// CORS設定
router.Use(func(c *gin.Context) {
    c.Header("Access-Control-Allow-Origin", "*")
    c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
    // ...
})
```

### データベースセキュリティ

- **UTF-8MB4**: 完全な Unicode サポート
- **タイムゾーン**: 日本時間（+09:00）設定
- **SQL インジェクション対策**: プリペアドステートメント使用

## 📝 ライセンス

MIT License

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 📞 サポート

技術的な質問や問題がある場合は、GitHub の Issues ページでお知らせください。

---

**Quivra Backend** - リアルタイムクイズプラットフォームのバックエンド実装
