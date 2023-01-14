package razchess

import (
	"github.com/notnil/chess"
	"github.com/notnil/chess/opening"
)

var book opening.Book = opening.NewBookECO()

type Move [2]string

type Update struct {
	FEN      string `json:"fen,omitempty"`
	PGN      string `json:"pgn,omitempty"`
	LastMove Move   `json:"move,omitempty"`
	Opening  string `json:"opening,omitempty"`
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
		if opening := book.Find(moves); opening != nil {
			u.Opening = opening.Title()
		}
	}
	return u
}
