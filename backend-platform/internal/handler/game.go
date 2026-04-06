package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/horkah/bacopa/backend-platform/internal/auth"
	"github.com/horkah/bacopa/backend-platform/internal/db"
	"github.com/horkah/bacopa/backend-platform/internal/game"
	"github.com/horkah/bacopa/backend-platform/internal/models"
)

type contextKey string

const contextKeyUserID contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			jsonError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			jsonError(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateToken(parts[1])
		if err != nil {
			jsonError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateGameID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetGameTypes(w http.ResponseWriter, r *http.Request) {
	types := []models.GameTypeInfo{
		{
			ID:          "tictactoe",
			Name:        "Tic Tac Toe",
			Description: "Classic 3x3 grid game. Get three in a row to win!",
			MinPlayers:  2,
			MaxPlayers:  2,
		},
		{
			ID:          "connectfour",
			Name:        "Connect Four",
			Description: "Drop discs into a 7-column grid. Connect four in a row to win!",
			MinPlayers:  2,
			MaxPlayers:  2,
		},
	}
	jsonResponse(w, http.StatusOK, types)
}

func CreateGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(int)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.GameType != "tictactoe" && req.GameType != "connectfour" {
		jsonError(w, "Invalid game type", http.StatusBadRequest)
		return
	}

	if req.Mode != "pvp" && req.Mode != "ai" {
		jsonError(w, "Invalid mode, must be 'pvp' or 'ai'", http.StatusBadRequest)
		return
	}

	if req.Mode == "ai" && req.AIDifficulty == "" {
		req.AIDifficulty = "medium"
	}

	engine := game.GetEngine(req.GameType)
	board := engine.NewBoard()
	boardJSON := engine.SerializeBoard(board)

	user, err := db.GetUserByID(userID)
	if err != nil {
		jsonError(w, "User not found", http.StatusNotFound)
		return
	}

	now := timeNow()
	gs := &models.GameSession{
		ID:            generateGameID(),
		GameType:      req.GameType,
		Mode:          req.Mode,
		Status:        "waiting",
		Player1ID:     userID,
		Board:         boardJSON,
		CurrentPlayer: 1,
		AIDifficulty:  req.AIDifficulty,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if req.Mode == "ai" {
		gs.Status = "playing"
		gs.Player2ID = 0 // AI has no real user ID
	}

	if err := db.CreateGame(gs); err != nil {
		jsonError(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusCreated, models.CreateGameResponse{
		GameID:   gs.ID,
		GameType: gs.GameType,
		Mode:     gs.Mode,
		Status:   gs.Status,
		Creator:  user.Username,
	})
}

func GetLobby(w http.ResponseWriter, r *http.Request) {
	games, err := db.GetLobbyGames()
	if err != nil {
		jsonError(w, "Failed to fetch lobby", http.StatusInternalServerError)
		return
	}
	if games == nil {
		games = []models.LobbyGame{}
	}
	jsonResponse(w, http.StatusOK, games)
}

func JoinGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(int)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	gameID := vars["id"]

	gs, err := db.GetGame(gameID)
	if err != nil {
		jsonError(w, "Game not found", http.StatusNotFound)
		return
	}

	if gs.Status != "waiting" {
		jsonError(w, "Game is not available to join", http.StatusBadRequest)
		return
	}

	if gs.Player1ID == userID {
		jsonError(w, "Cannot join your own game", http.StatusBadRequest)
		return
	}

	gs.Player2ID = userID
	gs.Status = "playing"

	if err := db.UpdateGame(gs); err != nil {
		jsonError(w, "Failed to join game", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, models.JoinGameResponse{
		GameID:   gs.ID,
		GameType: gs.GameType,
		Status:   gs.Status,
	})
}

func GetGameHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextKeyUserID).(int)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	history, err := db.GetGameHistory(userID)
	if err != nil {
		jsonError(w, "Failed to fetch game history", http.StatusInternalServerError)
		return
	}
	if history == nil {
		history = []models.GameHistoryEntry{}
	}
	jsonResponse(w, http.StatusOK, history)
}
