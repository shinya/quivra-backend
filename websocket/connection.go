package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"quivra-backend/models"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Conn     *websocket.Conn
	Send     chan []byte
	PlayerID string
	RoomID   string
}

type Hub struct {
	// 登録された接続
	connections map[*Connection]bool

	// ルーム別の接続
	rooms map[string]map[*Connection]bool

	// 接続からのメッセージを登録
	register chan *Connection

	// 接続の登録解除
	unregister chan *Connection

	// 接続にメッセージをブロードキャスト
	broadcast chan []byte

	// ルーム別ブロードキャスト
	roomBroadcast chan RoomMessage

	// 接続の保護
	mu sync.RWMutex
}

type RoomMessage struct {
	RoomID  string
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		connections:   make(map[*Connection]bool),
		rooms:         make(map[string]map[*Connection]bool),
		register:      make(chan *Connection),
		unregister:    make(chan *Connection),
		broadcast:     make(chan []byte),
		roomBroadcast: make(chan RoomMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case connection := <-h.register:
			h.mu.Lock()
			h.connections[connection] = true

			// ルームに接続を追加
			if connection.RoomID != "" {
				if h.rooms[connection.RoomID] == nil {
					h.rooms[connection.RoomID] = make(map[*Connection]bool)
				}
				h.rooms[connection.RoomID][connection] = true
			}
			h.mu.Unlock()

		case connection := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.connections[connection]; ok {
				delete(h.connections, connection)
				close(connection.Send)

				// ルームから接続を削除
				if connection.RoomID != "" {
					if room, exists := h.rooms[connection.RoomID]; exists {
						delete(room, connection)
						if len(room) == 0 {
							delete(h.rooms, connection.RoomID)
						}
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for connection := range h.connections {
				select {
				case connection.Send <- message:
				default:
					close(connection.Send)
					delete(h.connections, connection)
				}
			}
			h.mu.RUnlock()

		case roomMsg := <-h.roomBroadcast:
			h.mu.RLock()
			if room, exists := h.rooms[roomMsg.RoomID]; exists {
				for connection := range room {
					select {
					case connection.Send <- roomMsg.Message:
					default:
						close(connection.Send)
						delete(h.connections, connection)
						delete(room, connection)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToRoom(roomID string, message interface{}) {
	msg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.roomBroadcast <- RoomMessage{
		RoomID:  roomID,
		Message: msg,
	}
}

func (h *Hub) GetRoomConnections(roomID string) []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var connections []*Connection
	if room, exists := h.rooms[roomID]; exists {
		for conn := range room {
			connections = append(connections, conn)
		}
	}
	return connections
}

func (c *Connection) ReadPump(hub *Hub, wsHandler *WSHandler) {
	defer func() {
		log.Printf("WebSocket connection closed, unregistering...")
		hub.unregister <- c
		c.Conn.Close()
	}()

	log.Printf("WebSocket ReadPump started")

	for {
		var msg models.WSMessage
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		log.Printf("WebSocket message received: %+v", msg)
		// メッセージの処理
		wsHandler.HandleMessage(c, msg)
	}
}

func (c *Connection) WritePump() {
	defer c.Conn.Close()

	log.Printf("WebSocket WritePump started")

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				log.Printf("WebSocket send channel closed")
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			log.Printf("WebSocket sending message: %s", string(message))
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}
