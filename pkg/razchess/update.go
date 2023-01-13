package razchess

import (
	"github.com/notnil/chess"
)

type Move [2]string

type Update struct {
	FEN      string `json:"fen"`
	PGN      string `json:"pgn"`
	LastMove Move   `json:"move"`
}

func newUpdate(game *chess.Game) *Update {
	u := &Update{
		FEN: game.FEN(),
		PGN: game.String()[1:],
	}
	moves := game.Moves()
	if len(moves) > 0 {
		lastMove := moves[len(moves)-1]
		u.LastMove[0] = lastMove.S1().String()
		u.LastMove[1] = lastMove.S2().String()
	}
	return u
}
