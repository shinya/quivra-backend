package handlers

import (
	"net/http"

	"quivra-backend/services"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomService *services.RoomService
}

func NewRoomHandler(roomService *services.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// CreateRoom ルーム作成
func (rh *RoomHandler) CreateRoom(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := rh.roomService.CreateRoom(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roomId":  room.ID,
		"message": "ルームが作成されました",
	})
}

// GetRoom ルーム情報取得
func (rh *RoomHandler) GetRoom(c *gin.Context) {
	roomID := c.Param("roomId")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId is required"})
		return
	}

	room, err := rh.roomService.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, room)
}

// JoinRoom ルーム参加（WebSocket経由で実装予定）
func (rh *RoomHandler) JoinRoom(c *gin.Context) {
	var req struct {
		RoomID     string `json:"roomId" binding:"required"`
		PlayerName string `json:"playerName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player, err := rh.roomService.AddPlayer(req.RoomID, req.PlayerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"playerId": player.ID,
		"message":  "ルームに参加しました",
	})
}
