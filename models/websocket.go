package models

import "time"

// WebSocket イベントの構造体
type WSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// クライアント → サーバー イベント
type JoinRoomData struct {
	RoomID     string `json:"roomId"`
	PlayerName string `json:"playerName"`
}

type BuzzInData struct {
	RoomID string `json:"roomId"`
}

type SubmitAnswerData struct {
	RoomID string `json:"roomId"`
	Answer string `json:"answer"`
}

type StartGameData struct {
	RoomID string `json:"roomId"`
}

// サーバー → クライアント イベント
type RoomUpdatedData struct {
	Players         []Player  `json:"players"`
	GameState       string    `json:"gameState"`
	CurrentQuestion *Question `json:"currentQuestion,omitempty"`
	CanBuzz         bool      `json:"canBuzz"`
}

type BuzzResultData struct {
	Success      bool    `json:"success"`
	BuzzedPlayer *Player `json:"buzzedPlayer,omitempty"`
}

type QuestionResultData struct {
	Correct       bool   `json:"correct"`
	CorrectAnswer string `json:"correctAnswer"`
	Points        int    `json:"points"`
}

// 早押し状態管理
type BuzzState struct {
	CanBuzz    bool      `json:"canBuzz"`
	BuzzedBy   string    `json:"buzzedBy"`
	BuzzedAt   time.Time `json:"buzzedAt"`
	QuestionID int       `json:"questionId"`
}
