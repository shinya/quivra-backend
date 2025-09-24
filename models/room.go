package models

import (
	"time"
)

type Room struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Status    string    `json:"status" db:"status"`
	Players   []Player  `json:"players,omitempty"`
}

type Player struct {
	ID       string    `json:"id" db:"id"`
	RoomID   string    `json:"room_id" db:"room_id"`
	Name     string    `json:"name" db:"name"`
	Score    int       `json:"score" db:"score"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

type Question struct {
	ID         int       `json:"id" db:"id"`
	Question   string    `json:"question" db:"question"`
	Answer     string    `json:"answer" db:"answer"`
	Category   string    `json:"category" db:"category"`
	Difficulty string    `json:"difficulty" db:"difficulty"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type GameSession struct {
	ID             string     `json:"id" db:"id"`
	RoomID         string     `json:"room_id" db:"room_id"`
	QuestionID     *int       `json:"question_id" db:"question_id"`
	StartedAt      time.Time  `json:"started_at" db:"started_at"`
	EndedAt        *time.Time `json:"ended_at" db:"ended_at"`
	Status         string     `json:"status" db:"status"`
	BuzzedPlayerID *string    `json:"buzzed_player_id" db:"buzzed_player_id"`
}
