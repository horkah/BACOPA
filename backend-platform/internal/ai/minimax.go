package ai

import (
	"math"
	"math/rand"

	"github.com/horkah/bacopa/backend-platform/internal/game"
)

func GetAIMove(engine game.GameEngine, board interface{}, aiPlayer int, difficulty string) int {
	var depth int

	switch difficulty {
	case "easy":
		depth = 1 + rand.Intn(2) // 1 or 2
	case "medium":
		depth = 4
	case "hard":
		switch engine.(type) {
		case *game.TicTacToeEngine:
			depth = 6
		case *game.ConnectFourEngine:
			depth = 7
		default:
			depth = 6
		}
	default:
		depth = 4
	}

	validMoves := engine.GetValidMoves(board)
	if len(validMoves) == 0 {
		return -1
	}

	bestScore := math.MinInt32
	bestMove := validMoves[0]

	for _, move := range validMoves {
		newBoard := engine.ApplyMove(board, aiPlayer, move)
		opponent := 3 - aiPlayer
		score := minimax(engine, newBoard, depth-1, math.MinInt32, math.MaxInt32, false, aiPlayer, opponent)
		if score > bestScore {
			bestScore = score
			bestMove = move
		}
	}

	return bestMove
}

func minimax(engine game.GameEngine, board interface{}, depth int, alpha int, beta int, isMaximizing bool, aiPlayer int, opponent int) int {
	won, winner := engine.CheckWin(board)
	if won {
		if winner == aiPlayer {
			return 1000 + depth
		}
		return -1000 - depth
	}
	if engine.CheckDraw(board) {
		return 0
	}
	if depth == 0 {
		return evaluate(engine, board, aiPlayer)
	}

	validMoves := engine.GetValidMoves(board)

	if isMaximizing {
		maxScore := math.MinInt32
		for _, move := range validMoves {
			newBoard := engine.ApplyMove(board, aiPlayer, move)
			score := minimax(engine, newBoard, depth-1, alpha, beta, false, aiPlayer, opponent)
			if score > maxScore {
				maxScore = score
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
		return maxScore
	}

	minScore := math.MaxInt32
	for _, move := range validMoves {
		newBoard := engine.ApplyMove(board, opponent, move)
		score := minimax(engine, newBoard, depth-1, alpha, beta, true, aiPlayer, opponent)
		if score < minScore {
			minScore = score
		}
		if score < beta {
			beta = score
		}
		if beta <= alpha {
			break
		}
	}
	return minScore
}

func evaluate(engine game.GameEngine, board interface{}, aiPlayer int) int {
	switch engine.(type) {
	case *game.ConnectFourEngine:
		return evaluateConnectFour(board, aiPlayer)
	default:
		return 0
	}
}

func evaluateConnectFour(board interface{}, aiPlayer int) int {
	b := board.([6][7]int)
	opponent := 3 - aiPlayer
	score := 0

	// Center column preference
	for r := 0; r < 6; r++ {
		if b[r][3] == aiPlayer {
			score += 3
		} else if b[r][3] == opponent {
			score -= 3
		}
	}

	// Evaluate all windows of 4
	// Horizontal
	for r := 0; r < 6; r++ {
		for c := 0; c <= 3; c++ {
			score += evalWindow(b[r][c], b[r][c+1], b[r][c+2], b[r][c+3], aiPlayer, opponent)
		}
	}

	// Vertical
	for r := 0; r <= 2; r++ {
		for c := 0; c < 7; c++ {
			score += evalWindow(b[r][c], b[r+1][c], b[r+2][c], b[r+3][c], aiPlayer, opponent)
		}
	}

	// Diagonal down-right
	for r := 0; r <= 2; r++ {
		for c := 0; c <= 3; c++ {
			score += evalWindow(b[r][c], b[r+1][c+1], b[r+2][c+2], b[r+3][c+3], aiPlayer, opponent)
		}
	}

	// Diagonal down-left
	for r := 0; r <= 2; r++ {
		for c := 3; c < 7; c++ {
			score += evalWindow(b[r][c], b[r+1][c-1], b[r+2][c-2], b[r+3][c-3], aiPlayer, opponent)
		}
	}

	return score
}

func evalWindow(a, b, c, d, aiPlayer, opponent int) int {
	cells := [4]int{a, b, c, d}
	aiCount := 0
	oppCount := 0
	empty := 0
	for _, v := range cells {
		switch v {
		case 0:
			empty++
		default:
			if v == aiPlayer {
				aiCount++
			} else {
				oppCount++
			}
		}
	}

	if aiCount == 4 {
		return 100
	}
	if aiCount == 3 && empty == 1 {
		return 5
	}
	if aiCount == 2 && empty == 2 {
		return 2
	}
	if oppCount == 3 && empty == 1 {
		return -4
	}
	return 0
}
