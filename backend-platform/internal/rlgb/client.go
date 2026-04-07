package rlgb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DisplayState mirrors the display object returned by RLGB.
type DisplayState struct {
	Board         interface{} `json:"board"`
	CurrentPlayer int         `json:"currentPlayer"`
	LegalActions  []int       `json:"legalActions"`
	IsTerminal    bool        `json:"isTerminal"`
	Winner        *int        `json:"winner"`
	IsDraw        bool        `json:"isDraw"`
}

// GameInfo describes a game type supported by RLGB.
type GameInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NumPlayers  int    `json:"numPlayers"`
	NumActions  int    `json:"numActions"`
}

// NewGameResponse is the response from POST /games/{type}/new.
type NewGameResponse struct {
	State   json.RawMessage `json:"state"`
	Display DisplayState    `json:"display"`
}

// MoveResponse is the response from POST /games/{type}/move.
type MoveResponse struct {
	Valid   bool            `json:"valid"`
	Error   string          `json:"error,omitempty"`
	State   json.RawMessage `json:"state"`
	Display DisplayState    `json:"display"`
}

// AIMoveResponse is the response from POST /games/{type}/ai-move.
type AIMoveResponse struct {
	Action  int             `json:"action"`
	State   json.RawMessage `json:"state"`
	Display DisplayState    `json:"display"`
}

// Client is an HTTP client for the RLGB Python service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new RLGB client. It reads the RLGB_URL environment
// variable, defaulting to http://localhost:9090.
func NewClient() *Client {
	baseURL := os.Getenv("RLGB_URL")
	if baseURL == "" {
		baseURL = "http://localhost:9090"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Health checks that the RLGB service is reachable.
func (c *Client) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("rlgb health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rlgb health check returned status %d", resp.StatusCode)
	}
	return nil
}

// ListGames returns the list of supported game types.
func (c *Client) ListGames() ([]GameInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/games")
	if err != nil {
		return nil, fmt.Errorf("rlgb list games failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rlgb list games returned status %d", resp.StatusCode)
	}
	var games []GameInfo
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		return nil, fmt.Errorf("rlgb list games decode error: %w", err)
	}
	return games, nil
}

// NewGame creates a new game of the given type.
func (c *Client) NewGame(gameType string) (*NewGameResponse, error) {
	resp, err := c.httpClient.Post(c.baseURL+"/games/"+gameType+"/new", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("rlgb new game failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rlgb new game returned status %d: %s", resp.StatusCode, body)
	}
	var result NewGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("rlgb new game decode error: %w", err)
	}
	return &result, nil
}

// MakeMove sends a move to the RLGB service.
func (c *Client) MakeMove(gameType string, state json.RawMessage, action int) (*MoveResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"state":  json.RawMessage(state),
		"action": action,
	})
	resp, err := c.httpClient.Post(c.baseURL+"/games/"+gameType+"/move", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("rlgb move failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rlgb move returned status %d: %s", resp.StatusCode, respBody)
	}
	var result MoveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("rlgb move decode error: %w", err)
	}
	return &result, nil
}

// AIMove asks the RLGB service for an AI move.
func (c *Client) AIMove(gameType string, state json.RawMessage, difficulty string) (*AIMoveResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"state":      json.RawMessage(state),
		"difficulty": difficulty,
	})
	resp, err := c.httpClient.Post(c.baseURL+"/games/"+gameType+"/ai-move", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("rlgb ai-move failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rlgb ai-move returned status %d: %s", resp.StatusCode, respBody)
	}
	var result AIMoveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("rlgb ai-move decode error: %w", err)
	}
	return &result, nil
}
