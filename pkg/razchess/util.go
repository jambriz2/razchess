package razchess

import (
	"crypto/rand"
	"io"
	"strings"

	"github.com/notnil/chess"
)

const (
	charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	csLen   = byte(len(charset))
)

func GenerateID(length int) string {
	if length == 0 {
		return ""
	}
	output := make([]byte, 0, length)
	batchSize := length + length/4
	buf := make([]byte, batchSize)
	for {
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			panic(err)
		}
		for _, b := range buf {
			if b < (csLen * 4) {
				output = append(output, charset[b%csLen])
				if len(output) == length {
					return string(output)
				}
			}
		}
	}
}

func parseGame(game string) (func(*chess.Game), error) {
	switch {
	case strings.HasPrefix(game, "fen:"):
		return chess.FEN(game[4:])
	case strings.HasPrefix(game, "pgn:"):
		return chess.PGN(strings.NewReader(game[4:]))
	default:
		return chess.FEN(game)
	}
}

func gameToString(game *chess.Game, customGame bool) string {
	if customGame {
		return "fen:" + game.FEN()
	} else {
		return "pgn:" + game.String()[1:]
	}
}
