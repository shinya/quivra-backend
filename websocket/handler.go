package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"quivra-backend/models"
	"quivra-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		log.Printf("WebSocket origin check: %s", r.Header.Get("Origin"))
		return true // 本番環境では適切なCORS設定が必要
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSHandler struct {
	hub             *Hub
	roomService     *services.RoomService
	questionService *services.QuestionService
	gameService     *services.GameService
	buzzManager     *services.BuzzManager
}

func NewWSHandler(hub *Hub, roomService *services.RoomService, questionService *services.QuestionService, gameService *services.GameService, buzzManager *services.BuzzManager) *WSHandler {
	return &WSHandler{
		hub:             hub,
		roomService:     roomService,
		questionService: questionService,
		gameService:     gameService,
		buzzManager:     buzzManager,
	}
}

func (wsh *WSHandler) HandleWebSocket(c *gin.Context) {
	log.Printf("WebSocket connection attempt from %s", c.ClientIP())
	log.Printf("Request headers: %v", c.Request.Header)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		c.JSON(500, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	log.Printf("WebSocket connection established successfully")

	connection := &Connection{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		PlayerID: "",
		RoomID:   "",
	}

	wsh.hub.register <- connection

	go connection.WritePump()
	go connection.ReadPump(wsh.hub, wsh)
}

func (wsh *WSHandler) HandleMessage(conn *Connection, msg models.WSMessage) {
	switch msg.Event {
	case "join-room":
		wsh.handleJoinRoom(conn, msg.Data)
	case "buzz-in":
		wsh.handleBuzzIn(conn, msg.Data)
	case "submit-answer":
		wsh.handleSubmitAnswer(conn, msg.Data)
	case "start-game":
		wsh.handleStartGame(conn, msg.Data)
	default:
		log.Printf("Unknown event: %s", msg.Event)
	}
}

func (wsh *WSHandler) handleJoinRoom(conn *Connection, data interface{}) {
	// データを適切な構造体に変換
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling join room data: %v", err)
		wsh.sendError(conn, "Invalid message format")
		return
	}

	var joinData models.JoinRoomData
	if err := json.Unmarshal(jsonData, &joinData); err != nil {
		log.Printf("Error unmarshaling join room data: %v", err)
		wsh.sendError(conn, "Invalid message format")
		return
	}

	// プレイヤーをルームに追加
	player, err := wsh.roomService.AddPlayer(joinData.RoomID, joinData.PlayerName)
	if err != nil {
		log.Printf("Error adding player to room: %v", err)
		wsh.sendError(conn, "Room not found or player name already exists")
		return
	}

	// 接続情報を更新
	conn.PlayerID = player.ID
	conn.RoomID = joinData.RoomID

	// 成功メッセージを送信
	wsh.sendSuccess(conn, "Successfully joined room", map[string]interface{}{
		"playerId": player.ID,
		"roomId":   joinData.RoomID,
	})

	// ルーム状態を更新して全プレイヤーに送信
	wsh.broadcastRoomUpdate(joinData.RoomID)
}

func (wsh *WSHandler) handleBuzzIn(conn *Connection, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling buzz in data: %v", err)
		return
	}

	var buzzData models.BuzzInData
	if err := json.Unmarshal(jsonData, &buzzData); err != nil {
		log.Printf("Error unmarshaling buzz in data: %v", err)
		return
	}

	// 早押しを試行
	success := wsh.buzzManager.TryBuzz(buzzData.RoomID, conn.PlayerID)

	// 結果を送信
	result := models.BuzzResultData{
		Success: success,
	}

	if success {
		// ゲームセッションで早押しプレイヤーを設定
		session, err := wsh.gameService.GetActiveGameSession(buzzData.RoomID)
		if err == nil {
			wsh.gameService.SetBuzzedPlayer(session.ID, conn.PlayerID)
		}

		// プレイヤー情報を取得
		players, err := wsh.roomService.GetRoomPlayers(buzzData.RoomID)
		if err == nil {
			for _, player := range players {
				if player.ID == conn.PlayerID {
					result.BuzzedPlayer = &player
					break
				}
			}
		}
	}

	wsh.hub.SendToRoom(buzzData.RoomID, models.WSMessage{
		Event: "buzz-result",
		Data:  result,
	})
}

func (wsh *WSHandler) handleSubmitAnswer(conn *Connection, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling submit answer data: %v", err)
		return
	}

	var answerData models.SubmitAnswerData
	if err := json.Unmarshal(jsonData, &answerData); err != nil {
		log.Printf("Error unmarshaling submit answer data: %v", err)
		return
	}

	// アクティブなゲームセッションを取得
	session, err := wsh.gameService.GetActiveGameSession(answerData.RoomID)
	if err != nil {
		log.Printf("Error getting active game session: %v", err)
		return
	}

	// 問題を取得
	var question *models.Question
	if session.QuestionID != nil {
		question, err = wsh.questionService.GetQuestion(*session.QuestionID)
		if err != nil {
			log.Printf("Error getting question: %v", err)
			return
		}
	}

	// 回答の正誤判定
	correct := false
	correctAnswer := ""
	if question != nil {
		correct = answerData.Answer == question.Answer
		correctAnswer = question.Answer
	}

	points := 0
	if correct {
		// スコア計算（回答時間を考慮）
		timeToAnswer := time.Since(session.StartedAt)
		points = wsh.gameService.CalculateScore(question.Difficulty, timeToAnswer)

		// スコアを更新
		players, err := wsh.roomService.GetRoomPlayers(answerData.RoomID)
		if err == nil {
			for _, player := range players {
				if player.ID == conn.PlayerID {
					newScore := player.Score + points
					wsh.roomService.UpdatePlayerScore(player.ID, newScore)
					break
				}
			}
		}
	}

	// ゲームセッションを終了
	wsh.gameService.EndQuestion(session.ID, correct)

	// 早押し状態をリセット
	wsh.buzzManager.ResetBuzz(answerData.RoomID)

	// 結果を送信
	result := models.QuestionResultData{
		Correct:       correct,
		CorrectAnswer: correctAnswer,
		Points:        points,
	}

	wsh.hub.SendToRoom(answerData.RoomID, models.WSMessage{
		Event: "question-result",
		Data:  result,
	})

	// ルーム状態を更新
	wsh.broadcastRoomUpdate(answerData.RoomID)
}

func (wsh *WSHandler) handleStartGame(conn *Connection, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling start game data: %v", err)
		return
	}

	var startData models.StartGameData
	if err := json.Unmarshal(jsonData, &startData); err != nil {
		log.Printf("Error unmarshaling start game data: %v", err)
		return
	}

	// ルームの状態を更新
	err = wsh.roomService.UpdateRoomStatus(startData.RoomID, "playing")
	if err != nil {
		log.Printf("Error updating room status: %v", err)
		return
	}

	// ゲームセッションを作成
	session, err := wsh.gameService.CreateGameSession(startData.RoomID)
	if err != nil {
		log.Printf("Error creating game session: %v", err)
		return
	}

	// ランダムな問題を取得
	question, err := wsh.questionService.GetRandomQuestion("", "")
	if err != nil {
		log.Printf("Error getting random question: %v", err)
		return
	}

	// 問題を開始
	err = wsh.gameService.StartQuestion(session.ID, question.ID)
	if err != nil {
		log.Printf("Error starting question: %v", err)
		return
	}

	// 早押し状態を設定
	wsh.buzzManager.SetBuzzState(startData.RoomID, true, question.ID)

	// ルーム状態を更新
	wsh.broadcastRoomUpdate(startData.RoomID)
}

func (wsh *WSHandler) broadcastRoomUpdate(roomID string) {
	// ルーム情報を取得
	room, err := wsh.roomService.GetRoom(roomID)
	if err != nil {
		log.Printf("Error getting room: %v", err)
		return
	}

	// 早押し状態を取得
	buzzState, exists := wsh.buzzManager.GetBuzzState(roomID)
	canBuzz := exists && buzzState.CanBuzz && buzzState.BuzzedBy == ""

	// ルーム更新イベントを送信
	updateData := models.RoomUpdatedData{
		Players:   room.Players,
		GameState: room.Status,
		CanBuzz:   canBuzz,
	}

	// 現在の問題がある場合は追加
	if buzzState != nil && buzzState.QuestionID > 0 {
		question, err := wsh.questionService.GetQuestion(buzzState.QuestionID)
		if err == nil {
			updateData.CurrentQuestion = question
		}
	}

	wsh.hub.SendToRoom(roomID, models.WSMessage{
		Event: "room-updated",
		Data:  updateData,
	})
}

// sendError エラーメッセージを送信
func (wsh *WSHandler) sendError(conn *Connection, message string) {
	errorMsg := models.WSMessage{
		Event: "error",
		Data: map[string]interface{}{
			"message": message,
		},
	}

	msgBytes, err := json.Marshal(errorMsg)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		return
	}

	conn.Send <- msgBytes
}

// sendSuccess 成功メッセージを送信
func (wsh *WSHandler) sendSuccess(conn *Connection, message string, data map[string]interface{}) {
	successMsg := models.WSMessage{
		Event: "success",
		Data: map[string]interface{}{
			"message": message,
			"data":    data,
		},
	}

	msgBytes, err := json.Marshal(successMsg)
	if err != nil {
		log.Printf("Error marshaling success message: %v", err)
		return
	}

	conn.Send <- msgBytes
}
