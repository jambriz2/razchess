package razchess

import (
	"github.com/notnil/chess"
)

type Move [2]string

type Update struct {
	FEN       string `json:"fen"`
	WhiteMove Move   `json:"wm"`
	BlackMove Move   `json:"bm"`
}

func newUpdate(game *chess.Game) *Update {
	u := &Update{
		FEN: game.FEN(),
	}
	moves := game.Moves()
	positions := game.Positions()
	count := len(moves)
	if count > 1 {
		u.setMove(moves[count-1], positions[count-1])
		u.setMove(moves[count-2], positions[count-2])
	} else if count > 0 {
		u.setMove(moves[count-1], positions[count-1])
	}
	return u
}

func (u *Update) setMove(move *chess.Move, pos *chess.Position) {
	if pos.Board().Piece(move.S1()).Color() == chess.White {
		u.WhiteMove[0] = move.S1().String()
		u.WhiteMove[1] = move.S2().String()
	} else {
		u.BlackMove[0] = move.S1().String()
		u.BlackMove[1] = move.S2().String()
	}
}
