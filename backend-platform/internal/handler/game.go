package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/horkah/bacopa/backend-platform/internal/auth"
	"github.com/horkah/bacopa/backend-platform/internal/db"
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
	if rlgbClientRef == nil {
		jsonError(w, "Game service not configured", http.StatusServiceUnavailable)
		return
	}
	games, err := rlgbClientRef.ListGames()
	if err != nil {
		log.Printf("RLGB ListGames error: %v", err)
		jsonError(w, "Game service unavailable", http.StatusServiceUnavailable)
		return
	}
	types := make([]models.GameTypeInfo, len(games))
	for i, g := range games {
		types[i] = models.GameTypeInfo{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			MinPlayers:  g.NumPlayers,
			MaxPlayers:  g.NumPlayers,
		}
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

	if req.Mode != "pvp" && req.Mode != "ai" {
		jsonError(w, "Invalid mode, must be 'pvp' or 'ai'", http.StatusBadRequest)
		return
	}

	if req.Mode == "ai" && req.AIDifficulty == "" {
		req.AIDifficulty = "medium"
	}

	if rlgbClientRef == nil {
		jsonError(w, "Game service not configured", http.StatusServiceUnavailable)
		return
	}

	// Create new game via RLGB
	newGameResp, err := rlgbClientRef.NewGame(req.GameType)
	if err != nil {
		log.Printf("RLGB NewGame error: %v", err)
		jsonError(w, "Game service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Store both state and display as a wrapper in the board column
	boardJSON, _ := json.Marshal(map[string]interface{}{
		"state":   json.RawMessage(newGameResp.State),
		"display": newGameResp.Display,
	})

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
