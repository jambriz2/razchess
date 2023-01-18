package razchess

import (
	"strings"

	"github.com/notnil/chess"
	"github.com/notnil/chess/opening"
)

var book opening.Book = opening.NewBookECO()

type Move [2]string

type Update struct {
	Move          Move   `json:"move,omitempty"`
	Turn          string `json:"turn"`
	Status        string `json:"status"`
	FEN           string `json:"fen,omitempty"`
	PGN           string `json:"pgn,omitempty"`
	Opening       string `json:"opening,omitempty"`
	IsCapture     bool   `json:"isCapture"`
	IsGameOver    bool   `json:"isGameOver"`
	CheckedSquare string `json:"checkedSquare,omitempty"`
}

func newUpdate(game *chess.Game) *Update {
	u := &Update{
		Turn: game.Position().Turn().Name(),
		FEN:  game.FEN(),
		PGN:  game.String()[1:],
	}
	u.Status, u.IsGameOver = getStatus(game)
	if lastMove := getLastMove(game); lastMove != nil {
		u.Move[0] = lastMove.S1().String()
		u.Move[1] = lastMove.S2().String()
		if lastMove.HasTag(chess.Capture) {
			u.IsCapture = true
		}
		if lastMove.HasTag(chess.Check) {
			u.CheckedSquare = game.Position().Board().KingSquare(game.Position().Turn()).String()
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
