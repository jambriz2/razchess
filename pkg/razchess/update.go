package razchess

import (
	"strings"

	"github.com/notnil/chess"
	"github.com/notnil/chess/opening"
)

var book opening.Book = opening.NewBookECO()

type Move [2]string

type Update struct {
	FEN           string `json:"fen,omitempty"`
	PGN           string `json:"pgn,omitempty"`
	LastMove      Move   `json:"move,omitempty"`
	Opening       string `json:"opening,omitempty"`
	Status        string `json:"status"`
	IsGameOver    bool   `json:"isGameOver"`
	Turn          string `json:"turn"`
	CheckedSquare string `json:"checkedSquare,omitempty"`
}

func newUpdate(game *chess.Game) *Update {
	u := &Update{
		FEN:  game.FEN(),
		PGN:  game.String()[1:],
		Turn: game.Position().Turn().Name(),
	}
	u.Status, u.IsGameOver = getStatus(game)
	if lastMove := getLastMove(game); lastMove != nil {
		u.LastMove[0] = lastMove.S1().String()
		u.LastMove[1] = lastMove.S2().String()
		if lastMove.HasTag(chess.Check) {
			if game.Position().Turn() == chess.White {
				u.CheckedSquare = game.Position().Board().WhiteKingSquare().String()
			} else {
				u.CheckedSquare = game.Position().Board().BlackKingSquare().String()
			}
		}
	}
	if opening := book.Find(game.Moves()); opening != nil {
		u.Opening = opening.Title()
	}
	return u
}

func getLastMove(game *chess.Game) *chess.Move {
	moves := game.Moves()
	if len(moves) > 0 {
		return moves[len(moves)-1]
	}
	return nil
}

func getStatus(game *chess.Game) (string, bool) {
	turn := game.Position().Turn().Name()
	switch game.Position().Status() {
	case chess.Checkmate:
		return "Game over: " + turn + " is in checkmate", true
	case chess.Stalemate:
		return "Game over: stalemate", true
	default:
		fallthrough
	case chess.NoMethod:
		if game.Outcome() != chess.NoOutcome && game.Method() == chess.Resignation {
			switch game.Outcome() {
			case chess.WhiteWon:
				return "Black resigned", true
			case chess.BlackWon:
				return "White resigned", true
			}
		}
		status := turn + " to move"
		if lastMove := getLastMove(game); lastMove != nil && lastMove.HasTag(chess.Check) {
			status += ", " + strings.ToLower(turn) + " is in check"
		}
		return status, false
	}
}
