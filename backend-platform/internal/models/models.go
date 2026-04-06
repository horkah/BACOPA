package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Elo          int       `json:"elo"`
	CreatedAt    time.Time `json:"createdAt"`
}

type GameSession struct {
	ID            string          `json:"gameId"`
	GameType      string          `json:"gameType"`
	Mode          string          `json:"mode"`
	Status        string          `json:"status"`
	Player1ID     int             `json:"player1Id"`
	Player2ID     int             `json:"player2Id"`
	Board         json.RawMessage `json:"board"`
	CurrentPlayer int             `json:"currentPlayer"`
	Winner        int             `json:"winner"`
	AIDifficulty  string          `json:"aiDifficulty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type ChatMessage struct {
	From      string `json:"from"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type WSResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type MoveData struct {
	Position int `json:"position"`
}

type ChatData struct {
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Elo      int    `json:"elo"`
}

type GameTypeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MinPlayers  int    `json:"minPlayers"`
	MaxPlayers  int    `json:"maxPlayers"`
}

type CreateGameRequest struct {
	GameType     string `json:"gameType"`
	Mode         string `json:"mode"`
	AIDifficulty string `json:"aiDifficulty"`
}

type CreateGameResponse struct {
	GameID   string `json:"gameId"`
	GameType string `json:"gameType"`
	Mode     string `json:"mode"`
	Status   string `json:"status"`
	Creator  string `json:"creator"`
}

type LobbyGame struct {
	GameID    string    `json:"gameId"`
	GameType  string    `json:"gameType"`
	Mode      string    `json:"mode"`
	Creator   string    `json:"creator"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type JoinGameResponse struct {
	GameID   string `json:"gameId"`
	GameType string `json:"gameType"`
	Status   string `json:"status"`
}

type GameHistoryEntry struct {
	GameID    string    `json:"gameId"`
	GameType  string    `json:"gameType"`
	Mode      string    `json:"mode"`
	Opponent  string    `json:"opponent"`
	Result    string    `json:"result"`
	EloChange int       `json:"eloChange"`
	PlayedAt  time.Time `json:"playedAt"`
}

type GameStateData struct {
	Board         interface{}            `json:"board"`
	CurrentPlayer int                    `json:"currentPlayer"`
	Status        string                 `json:"status"`
	Winner        interface{}            `json:"winner"`
	Players       map[string]*PlayerInfo `json:"players"`
	GameType      string                 `json:"gameType"`
	LastMove      interface{}            `json:"lastMove"`
}

type PlayerInfo struct {
	Username string `json:"username"`
	Elo      int    `json:"elo"`
}

type GameOverData struct {
	Winner     interface{}    `json:"winner"`
	Reason     string         `json:"reason"`
	EloChanges map[string]int `json:"eloChanges"`
}

type ErrorData struct {
	Message string `json:"message"`
}

type PlayerJoinedData struct {
	Username     string `json:"username"`
	PlayerNumber int    `json:"playerNumber"`
}
