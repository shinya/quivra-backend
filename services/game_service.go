package services

import (
	"database/sql"
	"fmt"
	"time"

	"quivra-backend/database"
	"quivra-backend/models"
)

type GameService struct {
	db *database.DB
}

func NewGameService(db *database.DB) *GameService {
	return &GameService{db: db}
}

// CreateGameSession ゲームセッションを作成
func (gs *GameService) CreateGameSession(roomID string) (*models.GameSession, error) {
	sessionID := generateSessionID()

	query := `INSERT INTO game_sessions (id, room_id, status) VALUES (?, ?, 'waiting')`
	_, err := gs.db.Exec(query, sessionID, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	return &models.GameSession{
		ID:     sessionID,
		RoomID: roomID,
		Status: "waiting",
	}, nil
}

// StartQuestion 問題を開始
func (gs *GameService) StartQuestion(sessionID string, questionID int) error {
	query := `UPDATE game_sessions SET question_id = ?, status = 'question', started_at = NOW() WHERE id = ?`
	_, err := gs.db.Exec(query, questionID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to start question: %w", err)
	}
	return nil
}

// SetBuzzedPlayer 早押しプレイヤーを設定
func (gs *GameService) SetBuzzedPlayer(sessionID, playerID string) error {
	query := `UPDATE game_sessions SET buzzed_player_id = ?, status = 'buzzed' WHERE id = ?`
	_, err := gs.db.Exec(query, playerID, sessionID)
	if err != nil {
		return fmt.Errorf("failed to set buzzed player: %w", err)
	}
	return nil
}

// EndQuestion 問題を終了
func (gs *GameService) EndQuestion(sessionID string, correct bool) error {
	status := "finished"
	if correct {
		status = "answered"
	}

	query := `UPDATE game_sessions SET status = ?, ended_at = NOW() WHERE id = ?`
	_, err := gs.db.Exec(query, status, sessionID)
	if err != nil {
		return fmt.Errorf("failed to end question: %w", err)
	}
	return nil
}

// GetGameSession ゲームセッションを取得
func (gs *GameService) GetGameSession(sessionID string) (*models.GameSession, error) {
	var session models.GameSession
	query := `SELECT id, room_id, question_id, started_at, ended_at, status, buzzed_player_id FROM game_sessions WHERE id = ?`

	var questionID sql.NullInt64
	var endedAt sql.NullTime
	var buzzedPlayerID sql.NullString

	err := gs.db.QueryRow(query, sessionID).Scan(
		&session.ID, &session.RoomID, &questionID, &session.StartedAt,
		&endedAt, &session.Status, &buzzedPlayerID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("game session not found")
		}
		return nil, fmt.Errorf("failed to get game session: %w", err)
	}

	if questionID.Valid {
		qid := int(questionID.Int64)
		session.QuestionID = &qid
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	if buzzedPlayerID.Valid {
		pid := buzzedPlayerID.String
		session.BuzzedPlayerID = &pid
	}

	return &session, nil
}

// GetActiveGameSession アクティブなゲームセッションを取得
func (gs *GameService) GetActiveGameSession(roomID string) (*models.GameSession, error) {
	var session models.GameSession
	query := `SELECT id, room_id, question_id, started_at, ended_at, status, buzzed_player_id
			  FROM game_sessions
			  WHERE room_id = ? AND status IN ('waiting', 'question', 'buzzed')
			  ORDER BY started_at DESC LIMIT 1`

	var questionID sql.NullInt64
	var endedAt sql.NullTime
	var buzzedPlayerID sql.NullString

	err := gs.db.QueryRow(query, roomID).Scan(
		&session.ID, &session.RoomID, &questionID, &session.StartedAt,
		&endedAt, &session.Status, &buzzedPlayerID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active game session found")
		}
		return nil, fmt.Errorf("failed to get active game session: %w", err)
	}

	if questionID.Valid {
		qid := int(questionID.Int64)
		session.QuestionID = &qid
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	if buzzedPlayerID.Valid {
		pid := buzzedPlayerID.String
		session.BuzzedPlayerID = &pid
	}

	return &session, nil
}

// CalculateScore スコアを計算
func (gs *GameService) CalculateScore(difficulty string, timeToAnswer time.Duration) int {
	baseScore := 100

	// 難易度による基本スコア調整
	switch difficulty {
	case "easy":
		baseScore = 50
	case "medium":
		baseScore = 100
	case "hard":
		baseScore = 200
	}

	// 回答時間によるボーナス（5秒以内で最大ボーナス）
	timeBonus := 0
	if timeToAnswer <= 5*time.Second {
		timeBonus = 50
	} else if timeToAnswer <= 10*time.Second {
		timeBonus = 25
	}

	return baseScore + timeBonus
}

// generateSessionID セッションIDを生成
func generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
