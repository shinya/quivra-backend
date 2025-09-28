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
func (rs *RoomService) CreateRoom(name string, isPublic bool, creatorName string) (*models.Room, error) {
	roomID := generateRoomID()
	creatorID := generatePlayerID()

	// ルーム作成
	query := `INSERT INTO rooms (id, name, status, is_public, created_by) VALUES (?, ?, 'waiting', ?, ?)`
	_, err := rs.db.Exec(query, roomID, name, isPublic, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// 作成者を管理者として追加
	playerQuery := `INSERT INTO players (id, room_id, name, score, is_admin) VALUES (?, ?, ?, 0, TRUE)`
	_, err = rs.db.Exec(playerQuery, creatorID, roomID, creatorName)
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as admin: %w", err)
	}

	return &models.Room{
		ID:        roomID,
		Name:      name,
		Status:    "waiting",
		IsPublic:  isPublic,
		CreatedBy: creatorID,
	}, nil
}

// GetRoom ルーム情報を取得
func (rs *RoomService) GetRoom(roomID string) (*models.Room, error) {
	var room models.Room
	query := `SELECT id, name, created_at, status, is_public, created_by FROM rooms WHERE id = ?`
	err := rs.db.QueryRow(query, roomID).Scan(&room.ID, &room.Name, &room.CreatedAt, &room.Status, &room.IsPublic, &room.CreatedBy)
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
	query := `SELECT id, room_id, name, score, joined_at, is_admin FROM players WHERE room_id = ? ORDER BY joined_at`
	rows, err := rs.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.Player
		err := rows.Scan(&player.ID, &player.RoomID, &player.Name, &player.Score, &player.JoinedAt, &player.IsAdmin)
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

	// 既存のプレイヤーがいる場合はそのプレイヤー情報を返す
	for _, player := range players {
		if player.Name == playerName {
			return &player, nil
		}
	}

	playerID := generatePlayerID()
	query := `INSERT INTO players (id, room_id, name, score, is_admin) VALUES (?, ?, ?, 0, FALSE)`
	_, err = rs.db.Exec(query, playerID, roomID, playerName)
	if err != nil {
		return nil, fmt.Errorf("failed to add player: %w", err)
	}

	return &models.Player{
		ID:      playerID,
		RoomID:  roomID,
		Name:    playerName,
		Score:   0,
		IsAdmin: false,
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

// GetPublicRooms 公開ルーム一覧を取得
func (rs *RoomService) GetPublicRooms() ([]models.Room, error) {
	query := `SELECT id, name, created_at, status, is_public, created_by FROM rooms WHERE is_public = TRUE AND status = 'waiting' ORDER BY created_at DESC`
	rows, err := rs.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query public rooms: %w", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.Name, &room.CreatedAt, &room.Status, &room.IsPublic, &room.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}

		// プレイヤー数も取得
		players, err := rs.GetRoomPlayers(room.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get room players: %w", err)
		}
		room.Players = players
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// IsPlayerAdmin プレイヤーが管理者かチェック
func (rs *RoomService) IsPlayerAdmin(roomID, playerID string) (bool, error) {
	query := `SELECT is_admin FROM players WHERE room_id = ? AND id = ?`
	var isAdmin bool
	err := rs.db.QueryRow(query, roomID, playerID).Scan(&isAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("player not found")
		}
		return false, fmt.Errorf("failed to check admin status: %w", err)
	}
	return isAdmin, nil
}

// GetRoomRanking ルームのランキングを取得
func (rs *RoomService) GetRoomRanking(roomID string) ([]models.RoomRanking, error) {
	query := `SELECT id, name, score FROM players WHERE room_id = ? ORDER BY score DESC, joined_at ASC`
	rows, err := rs.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ranking: %w", err)
	}
	defer rows.Close()

	var rankings []models.RoomRanking
	rank := 1
	for rows.Next() {
		var ranking models.RoomRanking
		err := rows.Scan(&ranking.PlayerID, &ranking.Name, &ranking.Score)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ranking: %w", err)
		}
		ranking.Rank = rank
		rankings = append(rankings, ranking)
		rank++
	}

	return rankings, nil
}

// ResetAllData 全データをリセット（管理者向け）
func (rs *RoomService) ResetAllData() error {
	// 外部キー制約のため、削除順序に注意
	// 1. buzz_queue テーブルをクリア
	_, err := rs.db.Exec("DELETE FROM buzz_queue")
	if err != nil {
		return fmt.Errorf("failed to clear buzz_queue: %w", err)
	}

	// 2. game_sessions テーブルをクリア
	_, err = rs.db.Exec("DELETE FROM game_sessions")
	if err != nil {
		return fmt.Errorf("failed to clear game_sessions: %w", err)
	}

	// 3. players テーブルをクリア
	_, err = rs.db.Exec("DELETE FROM players")
	if err != nil {
		return fmt.Errorf("failed to clear players: %w", err)
	}

	// 4. rooms テーブルをクリア
	_, err = rs.db.Exec("DELETE FROM rooms")
	if err != nil {
		return fmt.Errorf("failed to clear rooms: %w", err)
	}

	// 5. questions テーブルをクリア（オプション）
	_, err = rs.db.Exec("DELETE FROM questions")
	if err != nil {
		return fmt.Errorf("failed to clear questions: %w", err)
	}

	// 6. サンプルデータを再挿入
	err = rs.insertSampleData()
	if err != nil {
		return fmt.Errorf("failed to insert sample data: %w", err)
	}

	return nil
}

// insertSampleData サンプルデータを挿入
func (rs *RoomService) insertSampleData() error {
	// サンプル問題を挿入
	sampleQuestions := []struct {
		question   string
		answer     string
		category   string
		difficulty string
	}{
		{"日本の首都は？", "東京", "地理", "easy"},
		{"1+1は？", "2", "数学", "easy"},
		{"Go言語の作者は？", "ロブ・パイク", "プログラミング", "medium"},
		{"世界で最も高い山は？", "エベレスト", "地理", "easy"},
		{"2の3乗は？", "8", "数学", "medium"},
		{"HTTPのデフォルトポートは？", "80", "プログラミング", "medium"},
		{"光の速度は？", "約30万km/s", "科学", "hard"},
		{"日本の国花は？", "桜", "文化", "easy"},
		{"Pythonの作者は？", "グイド・ヴァン・ロッサム", "プログラミング", "medium"},
		{"地球の衛星は？", "月", "科学", "easy"},
	}

	for _, q := range sampleQuestions {
		_, err := rs.db.Exec(
			"INSERT INTO questions (question, answer, category, difficulty) VALUES (?, ?, ?, ?)",
			q.question, q.answer, q.category, q.difficulty,
		)
		if err != nil {
			return fmt.Errorf("failed to insert sample question: %w", err)
		}
	}

	return nil
}

// generatePlayerID UUIDを生成
func generatePlayerID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
