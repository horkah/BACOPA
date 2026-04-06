package game

import "encoding/json"

type TicTacToeEngine struct{}

func (e *TicTacToeEngine) NewBoard() interface{} {
	board := [9]int{}
	return board
}

func (e *TicTacToeEngine) ValidateMove(board interface{}, player int, move int) bool {
	b := board.([9]int)
	if move < 0 || move > 8 {
		return false
	}
	return b[move] == 0
}

func (e *TicTacToeEngine) ApplyMove(board interface{}, player int, move int) interface{} {
	b := board.([9]int)
	b[move] = player
	return b
}

func (e *TicTacToeEngine) CheckWin(board interface{}) (bool, int) {
	b := board.([9]int)
	lines := [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
		{0, 4, 8}, {2, 4, 6}, // diags
	}
	for _, line := range lines {
		if b[line[0]] != 0 && b[line[0]] == b[line[1]] && b[line[1]] == b[line[2]] {
			return true, b[line[0]]
		}
	}
	return false, 0
}

func (e *TicTacToeEngine) CheckDraw(board interface{}) bool {
	b := board.([9]int)
	won, _ := e.CheckWin(board)
	if won {
		return false
	}
	for _, cell := range b {
		if cell == 0 {
			return false
		}
	}
	return true
}

func (e *TicTacToeEngine) GetValidMoves(board interface{}) []int {
	b := board.([9]int)
	var moves []int
	for i, cell := range b {
		if cell == 0 {
			moves = append(moves, i)
		}
	}
	return moves
}

func (e *TicTacToeEngine) SerializeBoard(board interface{}) json.RawMessage {
	b := board.([9]int)
	slice := b[:]
	data, _ := json.Marshal(slice)
	return data
}

func (e *TicTacToeEngine) DeserializeBoard(data json.RawMessage) interface{} {
	var slice []int
	json.Unmarshal(data, &slice)
	var board [9]int
	for i := 0; i < 9 && i < len(slice); i++ {
		board[i] = slice[i]
	}
	return board
}
