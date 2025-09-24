package services

import (
	"sync"
	"time"
)

type BuzzManager struct {
	mu         sync.RWMutex
	buzzStates map[string]*BuzzState // roomId -> BuzzState
}

type BuzzState struct {
	CanBuzz    bool
	BuzzedBy   string
	BuzzedAt   time.Time
	QuestionID int
}

func NewBuzzManager() *BuzzManager {
	return &BuzzManager{
		buzzStates: make(map[string]*BuzzState),
	}
}

// TryBuzz 早押しを試行する（競合状態を回避）
func (bm *BuzzManager) TryBuzz(roomId, playerId string) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	state, exists := bm.buzzStates[roomId]
	if !exists || !state.CanBuzz {
		return false
	}

	if state.BuzzedBy != "" {
		return false // 既に誰かが押している
	}

	state.BuzzedBy = playerId
	state.BuzzedAt = time.Now()
	return true
}

// SetBuzzState 早押し状態を設定
func (bm *BuzzManager) SetBuzzState(roomId string, canBuzz bool, questionId int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.buzzStates[roomId] = &BuzzState{
		CanBuzz:    canBuzz,
		BuzzedBy:   "",
		BuzzedAt:   time.Time{},
		QuestionID: questionId,
	}
}

// GetBuzzState 早押し状態を取得
func (bm *BuzzManager) GetBuzzState(roomId string) (*BuzzState, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	state, exists := bm.buzzStates[roomId]
	return state, exists
}

// ResetBuzz 早押し状態をリセット
func (bm *BuzzManager) ResetBuzz(roomId string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if state, exists := bm.buzzStates[roomId]; exists {
		state.BuzzedBy = ""
		state.BuzzedAt = time.Time{}
	}
}

// RemoveBuzzState 早押し状態を削除
func (bm *BuzzManager) RemoveBuzzState(roomId string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	delete(bm.buzzStates, roomId)
}
