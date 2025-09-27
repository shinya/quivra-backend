package services

import (
	"database/sql"
	"fmt"
	"time"

	"quivra-backend/database"
	"quivra-backend/models"
)

type BuzzQueueService struct {
	db *database.DB
}

func NewBuzzQueueService(db *database.DB) *BuzzQueueService {
	return &BuzzQueueService{db: db}
}

// AddToQueue プレイヤーを回答キューに追加
func (bqs *BuzzQueueService) AddToQueue(roomID, playerID string) error {
	// 既にキューにいるかチェック
	exists, err := bqs.IsPlayerInQueue(roomID, playerID)
	if err != nil {
		return fmt.Errorf("failed to check queue status: %w", err)
	}
	if exists {
		return fmt.Errorf("player already in queue")
	}

	queueID := generateQueueID()
	query := `INSERT INTO buzz_queue (id, room_id, player_id, buzzed_at, is_active) VALUES (?, ?, ?, ?, TRUE)`
	_, err = bqs.db.Exec(query, queueID, roomID, playerID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	return nil
}

// GetQueue ルームの回答キューを取得
func (bqs *BuzzQueueService) GetQueue(roomID string) ([]models.BuzzQueue, error) {
	query := `SELECT id, room_id, player_id, buzzed_at, is_active FROM buzz_queue WHERE room_id = ? AND is_active = TRUE ORDER BY buzzed_at ASC`
	rows, err := bqs.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query queue: %w", err)
	}
	defer rows.Close()

	var queue []models.BuzzQueue
	for rows.Next() {
		var buzz models.BuzzQueue
		err := rows.Scan(&buzz.ID, &buzz.RoomID, &buzz.PlayerID, &buzz.BuzzedAt, &buzz.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan queue: %w", err)
		}
		queue = append(queue, buzz)
	}

	return queue, nil
}

// GetNextPlayer 次の回答者を取得
func (bqs *BuzzQueueService) GetNextPlayer(roomID string) (*models.BuzzQueue, error) {
	query := `SELECT id, room_id, player_id, buzzed_at, is_active FROM buzz_queue WHERE room_id = ? AND is_active = TRUE ORDER BY buzzed_at ASC LIMIT 1`
	var buzz models.BuzzQueue
	err := bqs.db.QueryRow(query, roomID).Scan(&buzz.ID, &buzz.RoomID, &buzz.PlayerID, &buzz.BuzzedAt, &buzz.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no players in queue")
		}
		return nil, fmt.Errorf("failed to get next player: %w", err)
	}
	return &buzz, nil
}

// RemoveFromQueue プレイヤーをキューから削除
func (bqs *BuzzQueueService) RemoveFromQueue(roomID, playerID string) error {
	query := `UPDATE buzz_queue SET is_active = FALSE WHERE room_id = ? AND player_id = ?`
	_, err := bqs.db.Exec(query, roomID, playerID)
	if err != nil {
		return fmt.Errorf("failed to remove from queue: %w", err)
	}
	return nil
}

// ClearQueue ルームのキューをクリア
func (bqs *BuzzQueueService) ClearQueue(roomID string) error {
	query := `UPDATE buzz_queue SET is_active = FALSE WHERE room_id = ?`
	_, err := bqs.db.Exec(query, roomID)
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}
	return nil
}

// IsPlayerInQueue プレイヤーがキューにいるかチェック
func (bqs *BuzzQueueService) IsPlayerInQueue(roomID, playerID string) (bool, error) {
	query := `SELECT COUNT(*) FROM buzz_queue WHERE room_id = ? AND player_id = ? AND is_active = TRUE`
	var count int
	err := bqs.db.QueryRow(query, roomID, playerID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check queue: %w", err)
	}
	return count > 0, nil
}

// generateQueueID キューIDを生成
func generateQueueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
