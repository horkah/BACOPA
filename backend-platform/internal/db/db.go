package db

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/horkah/bacopa/backend-platform/internal/models"
)

var DB *sql.DB

func Initialize(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		elo INTEGER DEFAULT 1000,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	gamesTable := `
	CREATE TABLE IF NOT EXISTS games (
		id TEXT PRIMARY KEY,
		game_type TEXT NOT NULL,
		mode TEXT NOT NULL,
		status TEXT NOT NULL,
		player1_id INTEGER,
		player2_id INTEGER,
		board TEXT,
		current_player INTEGER DEFAULT 1,
		winner INTEGER DEFAULT 0,
		ai_difficulty TEXT DEFAULT '',
		created_at DATETIME,
		updated_at DATETIME,
		elo_change_p1 INTEGER DEFAULT 0,
		elo_change_p2 INTEGER DEFAULT 0,
		FOREIGN KEY (player1_id) REFERENCES users(id),
		FOREIGN KEY (player2_id) REFERENCES users(id)
	);`

	if _, err := DB.Exec(usersTable); err != nil {
		return err
	}
	if _, err := DB.Exec(gamesTable); err != nil {
		return err
	}
	return nil
}

func CreateUser(username, email, passwordHash string) (*models.User, error) {
	result, err := DB.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		username, email, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:       int(id),
		Username: username,
		Email:    email,
		Elo:      1000,
	}, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(
		"SELECT id, username, email, password_hash, elo, created_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Elo, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(
		"SELECT id, username, email, password_hash, elo, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Elo, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := DB.QueryRow(
		"SELECT id, username, email, password_hash, elo, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Elo, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CreateGame(g *models.GameSession) error {
	_, err := DB.Exec(
		`INSERT INTO games (id, game_type, mode, status, player1_id, player2_id, board, current_player, winner, ai_difficulty, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.ID, g.GameType, g.Mode, g.Status, g.Player1ID, g.Player2ID,
		string(g.Board), g.CurrentPlayer, g.Winner, g.AIDifficulty, g.CreatedAt, g.UpdatedAt,
	)
	return err
}

func GetGame(id string) (*models.GameSession, error) {
	g := &models.GameSession{}
	var boardStr string
	err := DB.QueryRow(
		`SELECT id, game_type, mode, status, player1_id, player2_id, board, current_player, winner, ai_difficulty, created_at, updated_at
		 FROM games WHERE id = ?`, id,
	).Scan(&g.ID, &g.GameType, &g.Mode, &g.Status, &g.Player1ID, &g.Player2ID,
		&boardStr, &g.CurrentPlayer, &g.Winner, &g.AIDifficulty, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	g.Board = json.RawMessage(boardStr)
	return g, nil
}

func UpdateGame(g *models.GameSession) error {
	g.UpdatedAt = time.Now()
	_, err := DB.Exec(
		`UPDATE games SET status=?, player2_id=?, board=?, current_player=?, winner=?, updated_at=?, elo_change_p1=?, elo_change_p2=?
		 WHERE id=?`,
		g.Status, g.Player2ID, string(g.Board), g.CurrentPlayer, g.Winner, g.UpdatedAt, 0, 0, g.ID,
	)
	return err
}

func UpdateGameWithElo(g *models.GameSession, eloP1, eloP2 int) error {
	g.UpdatedAt = time.Now()
	_, err := DB.Exec(
		`UPDATE games SET status=?, player2_id=?, board=?, current_player=?, winner=?, updated_at=?, elo_change_p1=?, elo_change_p2=?
		 WHERE id=?`,
		g.Status, g.Player2ID, string(g.Board), g.CurrentPlayer, g.Winner, g.UpdatedAt, eloP1, eloP2, g.ID,
	)
	return err
}

func GetLobbyGames() ([]models.LobbyGame, error) {
	rows, err := DB.Query(
		`SELECT g.id, g.game_type, g.mode, u.username, g.status, g.created_at
		 FROM games g JOIN users u ON g.player1_id = u.id
		 WHERE g.status = 'waiting' AND g.mode = 'pvp'
		 ORDER BY g.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []models.LobbyGame
	for rows.Next() {
		var g models.LobbyGame
		if err := rows.Scan(&g.GameID, &g.GameType, &g.Mode, &g.Creator, &g.Status, &g.CreatedAt); err != nil {
			return nil, err
		}
		games = append(games, g)
	}
	return games, nil
}

func GetGameHistory(userID int) ([]models.GameHistoryEntry, error) {
	rows, err := DB.Query(
		`SELECT g.id, g.game_type, g.mode, g.player1_id, g.player2_id, g.winner, g.updated_at, g.elo_change_p1, g.elo_change_p2,
		        COALESCE(u1.username, ''), COALESCE(u2.username, '')
		 FROM games g
		 LEFT JOIN users u1 ON g.player1_id = u1.id
		 LEFT JOIN users u2 ON g.player2_id = u2.id
		 WHERE (g.player1_id = ? OR g.player2_id = ?) AND g.status IN ('won', 'draw')
		 ORDER BY g.updated_at DESC`,
		userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.GameHistoryEntry
	for rows.Next() {
		var (
			gameID, gameType, mode     string
			p1ID, p2ID, winner         int
			updatedAt                  time.Time
			eloP1, eloP2              int
			username1, username2       string
		)
		if err := rows.Scan(&gameID, &gameType, &mode, &p1ID, &p2ID, &winner, &updatedAt, &eloP1, &eloP2, &username1, &username2); err != nil {
			return nil, err
		}

		entry := models.GameHistoryEntry{
			GameID:   gameID,
			GameType: gameType,
			Mode:     mode,
			PlayedAt: updatedAt,
		}

		if userID == p1ID {
			if mode == "ai" {
				entry.Opponent = "AI"
			} else {
				entry.Opponent = username2
			}
			entry.EloChange = eloP1
			if winner == 0 {
				entry.Result = "draw"
			} else if winner == 1 {
				entry.Result = "win"
			} else {
				entry.Result = "loss"
			}
		} else {
			entry.Opponent = username1
			entry.EloChange = eloP2
			if winner == 0 {
				entry.Result = "draw"
			} else if winner == 2 {
				entry.Result = "win"
			} else {
				entry.Result = "loss"
			}
		}

		history = append(history, entry)
	}
	return history, nil
}

func UpdateUserElo(userID int, newElo int) error {
	_, err := DB.Exec("UPDATE users SET elo = ? WHERE id = ?", newElo, userID)
	return err
}
