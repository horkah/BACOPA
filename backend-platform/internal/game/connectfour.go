package game

import "encoding/json"

type ConnectFourEngine struct{}

func (e *ConnectFourEngine) NewBoard() interface{} {
	var board [6][7]int
	return board
}

func (e *ConnectFourEngine) ValidateMove(board interface{}, player int, move int) bool {
	b := board.([6][7]int)
	if move < 0 || move > 6 {
		return false
	}
	// Top row of the column must be empty
	return b[0][move] == 0
}

func (e *ConnectFourEngine) ApplyMove(board interface{}, player int, move int) interface{} {
	b := board.([6][7]int)
	// Drop piece to lowest empty row
	for row := 5; row >= 0; row-- {
		if b[row][move] == 0 {
			b[row][move] = player
			break
		}
	}
	return b
}

func (e *ConnectFourEngine) CheckWin(board interface{}) (bool, int) {
	b := board.([6][7]int)

	// Check horizontal
	for r := 0; r < 6; r++ {
		for c := 0; c <= 3; c++ {
			if b[r][c] != 0 && b[r][c] == b[r][c+1] && b[r][c] == b[r][c+2] && b[r][c] == b[r][c+3] {
				return true, b[r][c]
			}
		}
	}

	// Check vertical
	for r := 0; r <= 2; r++ {
		for c := 0; c < 7; c++ {
			if b[r][c] != 0 && b[r][c] == b[r+1][c] && b[r][c] == b[r+2][c] && b[r][c] == b[r+3][c] {
				return true, b[r][c]
			}
		}
	}

	// Check diagonal (down-right)
	for r := 0; r <= 2; r++ {
		for c := 0; c <= 3; c++ {
			if b[r][c] != 0 && b[r][c] == b[r+1][c+1] && b[r][c] == b[r+2][c+2] && b[r][c] == b[r+3][c+3] {
				return true, b[r][c]
			}
		}
	}

	// Check diagonal (down-left)
	for r := 0; r <= 2; r++ {
		for c := 3; c < 7; c++ {
			if b[r][c] != 0 && b[r][c] == b[r+1][c-1] && b[r][c] == b[r+2][c-2] && b[r][c] == b[r+3][c-3] {
				return true, b[r][c]
			}
		}
	}

	return false, 0
}

func (e *ConnectFourEngine) CheckDraw(board interface{}) bool {
	b := board.([6][7]int)
	won, _ := e.CheckWin(board)
	if won {
		return false
	}
	// If top row is full, it's a draw
	for c := 0; c < 7; c++ {
		if b[0][c] == 0 {
			return false
		}
	}
	return true
}

func (e *ConnectFourEngine) GetValidMoves(board interface{}) []int {
	b := board.([6][7]int)
	var moves []int
	for c := 0; c < 7; c++ {
		if b[0][c] == 0 {
			moves = append(moves, c)
		}
	}
	return moves
}

func (e *ConnectFourEngine) SerializeBoard(board interface{}) json.RawMessage {
	b := board.([6][7]int)
	// Convert to [][]int for JSON
	slice := make([][]int, 6)
	for r := 0; r < 6; r++ {
		slice[r] = make([]int, 7)
		for c := 0; c < 7; c++ {
			slice[r][c] = b[r][c]
		}
	}
	data, _ := json.Marshal(slice)
	return data
}

func (e *ConnectFourEngine) DeserializeBoard(data json.RawMessage) interface{} {
	var slice [][]int
	json.Unmarshal(data, &slice)
	var board [6][7]int
	for r := 0; r < 6 && r < len(slice); r++ {
		for c := 0; c < 7 && c < len(slice[r]); c++ {
			board[r][c] = slice[r][c]
		}
	}
	return board
}
