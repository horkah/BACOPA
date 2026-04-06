package game

import "encoding/json"

type GameEngine interface {
	NewBoard() interface{}
	ValidateMove(board interface{}, player int, move int) bool
	ApplyMove(board interface{}, player int, move int) interface{}
	CheckWin(board interface{}) (bool, int)
	CheckDraw(board interface{}) bool
	GetValidMoves(board interface{}) []int
	SerializeBoard(board interface{}) json.RawMessage
	DeserializeBoard(data json.RawMessage) interface{}
}

func GetEngine(gameType string) GameEngine {
	switch gameType {
	case "tictactoe":
		return &TicTacToeEngine{}
	case "connectfour":
		return &ConnectFourEngine{}
	default:
		return nil
	}
}
