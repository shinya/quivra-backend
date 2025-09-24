package services

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"quivra-backend/database"
	"quivra-backend/models"
)

type RoomService struct {
	db *database.DB
}

func NewRoomService(db *database.DB) *RoomService {
	return &RoomService{db: db}
}

// CreateRoom ルームを作成
func (rs *RoomService) CreateRoom(name string) (*models.Room, error) {
	roomID := generateRoomID()

	query := `INSERT INTO rooms (id, name, status) VALUES (?, ?, 'waiting')`
	_, err := rs.db.Exec(query, roomID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return &models.Room{
		ID:     roomID,
		Name:   name,
		Status: "waiting",
	}, nil
}

// GetRoom ルーム情報を取得
func (rs *RoomService) GetRoom(roomID string) (*models.Room, error) {
	var room models.Room
	query := `SELECT id, name, created_at, status FROM rooms WHERE id = ?`
	err := rs.db.QueryRow(query, roomID).Scan(&room.ID, &room.Name, &room.CreatedAt, &room.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("room not found")
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// プレイヤー情報も取得
	players, err := rs.GetRoomPlayers(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room players: %w", err)
	}
	room.Players = players

	return &room, nil
}

// GetRoomPlayers ルームのプレイヤー一覧を取得
func (rs *RoomService) GetRoomPlayers(roomID string) ([]models.Player, error) {
	query := `SELECT id, room_id, name, score, joined_at FROM players WHERE room_id = ? ORDER BY joined_at`
	rows, err := rs.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.Player
		err := rows.Scan(&player.ID, &player.RoomID, &player.Name, &player.Score, &player.JoinedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, player)
	}

	return players, nil
}

// AddPlayer プレイヤーをルームに追加
func (rs *RoomService) AddPlayer(roomID, playerName string) (*models.Player, error) {
	// ルームの存在確認
	_, err := rs.GetRoom(roomID)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}

	// プレイヤー名の重複チェック
	players, err := rs.GetRoomPlayers(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing players: %w", err)
	}

	for _, player := range players {
		if player.Name == playerName {
			return nil, fmt.Errorf("player name already exists")
		}
	}

	playerID := generatePlayerID()
	query := `INSERT INTO players (id, room_id, name, score) VALUES (?, ?, ?, 0)`
	_, err = rs.db.Exec(query, playerID, roomID, playerName)
	if err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}

	return &models.Player{
		ID:     playerID,
		RoomID: roomID,
		Name:   playerName,
		Score:  0,
	}, nil
}

// UpdateRoomStatus ルームの状態を更新
func (rs *RoomService) UpdateRoomStatus(roomID, status string) error {
	query := `UPDATE rooms SET status = ? WHERE id = ?`
	_, err := rs.db.Exec(query, status, roomID)
	if err != nil {
		return fmt.Errorf("failed to update room status: %w", err)
	}
	return nil
}

// UpdatePlayerScore プレイヤーのスコアを更新
func (rs *RoomService) UpdatePlayerScore(playerID string, score int) error {
	query := `UPDATE players SET score = ? WHERE id = ?`
	_, err := rs.db.Exec(query, score, playerID)
	if err != nil {
		return fmt.Errorf("failed to update player score: %w", err)
	}
	return nil
}

// generateRoomID ルームIDを生成（10文字の英数字）
func generateRoomID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	var result strings.Builder
	for i := 0; i < 10; i++ {
		result.WriteByte(charset[rand.Intn(len(charset))])
	}
	return result.String()
}

// generatePlayerID UUIDを生成
func generatePlayerID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
