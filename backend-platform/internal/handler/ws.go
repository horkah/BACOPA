package handler

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/horkah/bacopa/backend-platform/internal/auth"
	"github.com/horkah/bacopa/backend-platform/internal/db"
	"github.com/horkah/bacopa/backend-platform/internal/models"
	"github.com/horkah/bacopa/backend-platform/internal/rlgb"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PlayerConn struct {
	Conn         *websocket.Conn
	UserID       int
	Username     string
	Elo          int
	PlayerNumber int
	mu           sync.Mutex
}

func (p *PlayerConn) SendJSON(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Conn.WriteJSON(v)
}

type GameRoom struct {
	mu         sync.RWMutex
	GameID     string
	GameType   string
	Mode       string
	State      json.RawMessage   // Opaque RLGB state
	Display    rlgb.DisplayState // Last display state from RLGB
	Status     string
	Winner     int
	LastMove   interface{}
	Players    [2]*PlayerConn
	AIDiff     string
	RLGBClient *rlgb.Client
}

var (
	rooms   = make(map[string]*GameRoom)
	roomsMu sync.RWMutex
)

// rlgbClient is the shared RLGB client, set by SetRLGBClient from main.
var rlgbClientRef *rlgb.Client

// SetRLGBClient stores the shared RLGB client for room creation.
func SetRLGBClient(c *rlgb.Client) {
	rlgbClientRef = c
}

func getOrCreateRoom(gameID string) (*GameRoom, error) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if room, ok := rooms[gameID]; ok {
		return room, nil
	}

	gs, err := db.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	// The DB board column stores the opaque RLGB state JSON.
	// We also need a display state; for a room loaded from DB we need to
	// reconstruct it. We call RLGB move with no action to get display,
	// but since RLGB doesn't have such an endpoint, we store both state
	// and display in the DB board column as a wrapper object.
	var stored struct {
		State   json.RawMessage   `json:"state"`
		Display rlgb.DisplayState `json:"display"`
	}
	if err := json.Unmarshal(gs.Board, &stored); err != nil {
		return nil, err
	}

	room := &GameRoom{
		GameID:     gameID,
		GameType:   gs.GameType,
		Mode:       gs.Mode,
		State:      stored.State,
		Display:    stored.Display,
		Status:     gs.Status,
		Winner:     gs.Winner,
		LastMove:   nil,
		AIDiff:     gs.AIDifficulty,
		RLGBClient: rlgbClientRef,
	}

	rooms[gameID] = room
	return room, nil
}

func removeRoom(gameID string) {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	delete(rooms, gameID)
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	gameID := r.URL.Query().Get("gameId")

	if tokenStr == "" || gameID == "" {
		http.Error(w, "Missing token or gameId", http.StatusBadRequest)
		return
	}

	userID, err := auth.ValidateToken(tokenStr)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	gs, err := db.GetGame(gameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Determine player number
	var playerNumber int
	if gs.Player1ID == userID {
		playerNumber = 1
	} else if gs.Player2ID == userID {
		playerNumber = 2
	} else {
		http.Error(w, "You are not a player in this game", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	room, err := getOrCreateRoom(gameID)
	if err != nil {
		conn.WriteJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "Failed to load game"},
		})
		conn.Close()
		return
	}

	pc := &PlayerConn{
		Conn:         conn,
		UserID:       userID,
		Username:     user.Username,
		Elo:          user.Elo,
		PlayerNumber: playerNumber,
	}

	room.mu.Lock()
	room.Players[playerNumber-1] = pc

	// If player 2 just joined a pvp game that was waiting, notify player 1
	if playerNumber == 2 && room.Status == "waiting" {
		room.Status = "playing"
	}
	room.mu.Unlock()

	// Notify other player that someone joined
	if playerNumber == 2 {
		room.mu.RLock()
		otherPlayer := room.Players[0]
		room.mu.RUnlock()
		if otherPlayer != nil {
			otherPlayer.SendJSON(models.WSResponse{
				Type: "player_joined",
				Data: models.PlayerJoinedData{
					Username:     user.Username,
					PlayerNumber: 2,
				},
			})
		}
	}

	// Send initial game state
	sendGameState(room, pc)

	// Read loop
	go handleMessages(room, pc)
}

func buildGameState(room *GameRoom) models.GameStateData {
	room.mu.RLock()
	defer room.mu.RUnlock()

	players := make(map[string]*models.PlayerInfo)
	if room.Players[0] != nil {
		players["1"] = &models.PlayerInfo{
			Username: room.Players[0].Username,
			Elo:      room.Players[0].Elo,
		}
	}
	if room.Mode == "ai" {
		players["2"] = &models.PlayerInfo{
			Username: "AI",
			Elo:      0,
		}
	} else if room.Players[1] != nil {
		players["2"] = &models.PlayerInfo{
			Username: room.Players[1].Username,
			Elo:      room.Players[1].Elo,
		}
	}

	var winner interface{}
	if room.Winner == 0 {
		winner = nil
	} else {
		winner = room.Winner
	}

	return models.GameStateData{
		Board:         room.Display.Board,
		CurrentPlayer: room.Display.CurrentPlayer,
		Status:        room.Status,
		Winner:        winner,
		Players:       players,
		GameType:      room.GameType,
		LastMove:      room.LastMove,
	}
}

func sendGameState(room *GameRoom, pc *PlayerConn) {
	state := buildGameState(room)
	pc.SendJSON(models.WSResponse{
		Type: "game_state",
		Data: state,
	})
}

func broadcastGameState(room *GameRoom) {
	state := buildGameState(room)
	msg := models.WSResponse{
		Type: "game_state",
		Data: state,
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	for _, p := range room.Players {
		if p != nil {
			p.SendJSON(msg)
		}
	}
}

func broadcastToRoom(room *GameRoom, msg models.WSResponse) {
	room.mu.RLock()
	defer room.mu.RUnlock()
	for _, p := range room.Players {
		if p != nil {
			p.SendJSON(msg)
		}
	}
}

func handleMessages(room *GameRoom, pc *PlayerConn) {
	defer func() {
		pc.Conn.Close()
		room.mu.Lock()
		if room.Players[pc.PlayerNumber-1] == pc {
			room.Players[pc.PlayerNumber-1] = nil
		}
		// Check if room is empty
		empty := room.Players[0] == nil && room.Players[1] == nil
		room.mu.Unlock()

		if empty {
			removeRoom(room.GameID)
		}
	}()

	for {
		var msg models.WSMessage
		err := pc.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}

		switch msg.Type {
		case "move":
			handleMove(room, pc, msg.Data)
		case "chat":
			handleChat(room, pc, msg.Data)
		}
	}
}

func handleMove(room *GameRoom, pc *PlayerConn, data json.RawMessage) {
	var moveData models.MoveData
	if err := json.Unmarshal(data, &moveData); err != nil {
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "Invalid move data"},
		})
		return
	}

	room.mu.Lock()

	if room.Status != "playing" {
		room.mu.Unlock()
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "Game is not in progress"},
		})
		return
	}

	if room.Display.CurrentPlayer != pc.PlayerNumber {
		room.mu.Unlock()
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "It is not your turn"},
		})
		return
	}

	// Delegate move to RLGB
	moveResp, err := room.RLGBClient.MakeMove(room.GameType, room.State, moveData.Position)
	if err != nil {
		room.mu.Unlock()
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "Game service unavailable"},
		})
		return
	}

	if !moveResp.Valid {
		room.mu.Unlock()
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "Invalid move"},
		})
		return
	}

	// Update room state from RLGB response
	room.State = moveResp.State
	room.Display = moveResp.Display
	room.LastMove = moveData.Position

	// Check terminal conditions
	gameOver := false
	if moveResp.Display.IsTerminal {
		if moveResp.Display.IsDraw {
			room.Status = "draw"
			room.Winner = 0
		} else if moveResp.Display.Winner != nil {
			room.Status = "won"
			room.Winner = *moveResp.Display.Winner
		}
		gameOver = true
	}

	persistGame(room)
	room.mu.Unlock()

	broadcastGameState(room)

	if gameOver {
		handleGameOver(room)
		return
	}

	// If AI mode and it's AI's turn
	room.mu.RLock()
	isAI := room.Mode == "ai" && room.Display.CurrentPlayer == 2 && room.Status == "playing"
	room.mu.RUnlock()

	if isAI {
		handleAIMove(room)
	}
}

func handleAIMove(room *GameRoom) {
	room.mu.Lock()

	aiResp, err := room.RLGBClient.AIMove(room.GameType, room.State, room.AIDiff)
	if err != nil {
		room.mu.Unlock()
		log.Printf("RLGB AI move error: %v", err)
		return
	}

	room.State = aiResp.State
	room.Display = aiResp.Display
	room.LastMove = aiResp.Action

	gameOver := false
	if aiResp.Display.IsTerminal {
		if aiResp.Display.IsDraw {
			room.Status = "draw"
			room.Winner = 0
		} else if aiResp.Display.Winner != nil {
			room.Status = "won"
			room.Winner = *aiResp.Display.Winner
		}
		gameOver = true
	}

	persistGame(room)
	room.mu.Unlock()

	broadcastGameState(room)

	if gameOver {
		handleGameOver(room)
	}
}

func persistGame(room *GameRoom) {
	// Store both state and display as a wrapper in the board column
	wrapper, _ := json.Marshal(map[string]interface{}{
		"state":   json.RawMessage(room.State),
		"display": room.Display,
	})

	currentPlayer := room.Display.CurrentPlayer

	gs := &models.GameSession{
		ID:            room.GameID,
		Status:        room.Status,
		Board:         json.RawMessage(wrapper),
		CurrentPlayer: currentPlayer,
		Winner:        room.Winner,
	}

	// Load existing game to preserve fields
	existing, err := db.GetGame(room.GameID)
	if err == nil {
		gs.GameType = existing.GameType
		gs.Mode = existing.Mode
		gs.Player1ID = existing.Player1ID
		gs.Player2ID = existing.Player2ID
		gs.AIDifficulty = existing.AIDifficulty
		gs.CreatedAt = existing.CreatedAt
	}

	db.UpdateGame(gs)
}

func handleGameOver(room *GameRoom) {
	room.mu.RLock()
	winner := room.Winner
	status := room.Status
	mode := room.Mode
	gameID := room.GameID

	var p1 *PlayerConn
	var p2 *PlayerConn
	p1 = room.Players[0]
	p2 = room.Players[1]
	room.mu.RUnlock()

	eloChanges := make(map[string]int)

	gs, err := db.GetGame(gameID)
	if err != nil {
		return
	}

	if mode == "pvp" && p1 != nil {
		// Get current Elo from DB
		user1, err1 := db.GetUserByID(gs.Player1ID)
		user2, err2 := db.GetUserByID(gs.Player2ID)
		if err1 != nil || err2 != nil {
			return
		}

		elo1 := float64(user1.Elo)
		elo2 := float64(user2.Elo)

		var s1, s2 float64
		if status == "draw" {
			s1 = 0.5
			s2 = 0.5
		} else if winner == 1 {
			s1 = 1.0
			s2 = 0.0
		} else {
			s1 = 0.0
			s2 = 1.0
		}

		k := 32.0
		e1 := 1.0 / (1.0 + math.Pow(10, (elo2-elo1)/400.0))
		e2 := 1.0 / (1.0 + math.Pow(10, (elo1-elo2)/400.0))

		delta1 := int(math.Round(k * (s1 - e1)))
		delta2 := int(math.Round(k * (s2 - e2)))

		newElo1 := user1.Elo + delta1
		newElo2 := user2.Elo + delta2
		if newElo1 < 0 {
			newElo1 = 0
		}
		if newElo2 < 0 {
			newElo2 = 0
		}

		db.UpdateUserElo(user1.ID, newElo1)
		db.UpdateUserElo(user2.ID, newElo2)
		db.UpdateGameWithElo(gs, delta1, delta2)

		eloChanges[user1.Username] = delta1
		eloChanges[user2.Username] = delta2

		// Update cached Elo in player conns
		if p1 != nil {
			p1.Elo = newElo1
		}
		if p2 != nil {
			p2.Elo = newElo2
		}
	} else if mode == "ai" && p1 != nil {
		user1, err1 := db.GetUserByID(gs.Player1ID)
		if err1 != nil {
			return
		}

		// For AI games, use a virtual AI Elo based on difficulty
		aiElo := 1000.0
		switch gs.AIDifficulty {
		case "easy":
			aiElo = 800
		case "medium":
			aiElo = 1200
		case "hard":
			aiElo = 1600
		}

		elo1 := float64(user1.Elo)
		var s1 float64
		if status == "draw" {
			s1 = 0.5
		} else if winner == 1 {
			s1 = 1.0
		} else {
			s1 = 0.0
		}

		k := 32.0
		e1 := 1.0 / (1.0 + math.Pow(10, (aiElo-elo1)/400.0))
		delta1 := int(math.Round(k * (s1 - e1)))

		newElo1 := user1.Elo + delta1
		if newElo1 < 0 {
			newElo1 = 0
		}

		db.UpdateUserElo(user1.ID, newElo1)
		db.UpdateGameWithElo(gs, delta1, 0)

		eloChanges[user1.Username] = delta1
		if p1 != nil {
			p1.Elo = newElo1
		}
	}

	var winnerName interface{}
	if status == "draw" {
		winnerName = nil
	} else {
		room.mu.RLock()
		if winner == 1 && room.Players[0] != nil {
			winnerName = room.Players[0].Username
		} else if winner == 2 {
			if mode == "ai" {
				winnerName = "AI"
			} else if room.Players[1] != nil {
				winnerName = room.Players[1].Username
			}
		}
		room.mu.RUnlock()
	}

	reason := "win"
	if status == "draw" {
		reason = "draw"
	}

	broadcastToRoom(room, models.WSResponse{
		Type: "game_over",
		Data: models.GameOverData{
			Winner:     winnerName,
			Reason:     reason,
			EloChanges: eloChanges,
		},
	})
}

func handleChat(room *GameRoom, pc *PlayerConn, data json.RawMessage) {
	var chatData models.ChatData
	if err := json.Unmarshal(data, &chatData); err != nil {
		return
	}

	if chatData.Message == "" {
		return
	}

	broadcastToRoom(room, models.WSResponse{
		Type: "chat",
		Data: models.ChatMessage{
			From:      pc.Username,
			Message:   chatData.Message,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}
