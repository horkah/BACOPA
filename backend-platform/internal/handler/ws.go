package handler

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/horkah/bacopa/backend-platform/internal/ai"
	"github.com/horkah/bacopa/backend-platform/internal/auth"
	"github.com/horkah/bacopa/backend-platform/internal/db"
	"github.com/horkah/bacopa/backend-platform/internal/game"
	"github.com/horkah/bacopa/backend-platform/internal/models"
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
	mu       sync.RWMutex
	GameID   string
	GameType string
	Mode     string
	Engine   game.GameEngine
	Board    interface{}
	Status   string
	Current  int // 1 or 2
	Winner   int
	LastMove interface{}
	Players  [2]*PlayerConn // index 0 = player1, index 1 = player2
	AIDiff   string
}

var (
	rooms   = make(map[string]*GameRoom)
	roomsMu sync.RWMutex
)

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

	engine := game.GetEngine(gs.GameType)
	if engine == nil {
		return nil, err
	}

	board := engine.DeserializeBoard(gs.Board)

	room := &GameRoom{
		GameID:   gameID,
		GameType: gs.GameType,
		Mode:     gs.Mode,
		Engine:   engine,
		Board:    board,
		Status:   gs.Status,
		Current:  gs.CurrentPlayer,
		Winner:   gs.Winner,
		LastMove: nil,
		AIDiff:   gs.AIDifficulty,
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

	serialized := room.Engine.SerializeBoard(room.Board)
	var boardData interface{}
	json.Unmarshal(serialized, &boardData)

	return models.GameStateData{
		Board:         boardData,
		CurrentPlayer: room.Current,
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

	if room.Current != pc.PlayerNumber {
		room.mu.Unlock()
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "It is not your turn"},
		})
		return
	}

	if !room.Engine.ValidateMove(room.Board, pc.PlayerNumber, moveData.Position) {
		room.mu.Unlock()
		pc.SendJSON(models.WSResponse{
			Type: "error",
			Data: models.ErrorData{Message: "Invalid move"},
		})
		return
	}

	// Apply move
	room.Board = room.Engine.ApplyMove(room.Board, pc.PlayerNumber, moveData.Position)
	room.LastMove = moveData.Position

	// Check win/draw
	gameOver := false
	won, winner := room.Engine.CheckWin(room.Board)
	if won {
		room.Status = "won"
		room.Winner = winner
		gameOver = true
	} else if room.Engine.CheckDraw(room.Board) {
		room.Status = "draw"
		room.Winner = 0
		gameOver = true
	} else {
		room.Current = 3 - room.Current
	}

	// Save to DB
	persistGame(room)
	room.mu.Unlock()

	broadcastGameState(room)

	if gameOver {
		handleGameOver(room)
		return
	}

	// If AI mode and it's AI's turn
	room.mu.RLock()
	isAI := room.Mode == "ai" && room.Current == 2 && room.Status == "playing"
	room.mu.RUnlock()

	if isAI {
		handleAIMove(room)
	}
}

func handleAIMove(room *GameRoom) {
	room.mu.Lock()

	aiMove := ai.GetAIMove(room.Engine, room.Board, 2, room.AIDiff)
	if aiMove < 0 {
		room.mu.Unlock()
		return
	}

	room.Board = room.Engine.ApplyMove(room.Board, 2, aiMove)
	room.LastMove = aiMove

	gameOver := false
	won, winner := room.Engine.CheckWin(room.Board)
	if won {
		room.Status = "won"
		room.Winner = winner
		gameOver = true
	} else if room.Engine.CheckDraw(room.Board) {
		room.Status = "draw"
		room.Winner = 0
		gameOver = true
	} else {
		room.Current = 1
	}

	persistGame(room)
	room.mu.Unlock()

	broadcastGameState(room)

	if gameOver {
		handleGameOver(room)
	}
}

func persistGame(room *GameRoom) {
	boardJSON := room.Engine.SerializeBoard(room.Board)
	gs := &models.GameSession{
		ID:            room.GameID,
		Status:        room.Status,
		Board:         boardJSON,
		CurrentPlayer: room.Current,
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
