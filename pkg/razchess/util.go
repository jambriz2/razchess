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

func parseGame(game string) (opts []func(*chess.Game), isCustom bool, err error) {
	var opt func(*chess.Game)
	switch {
	case len(game) == 0:
		// do nothing
	case strings.HasPrefix(game, "fen:"):
		opt, err = chess.FEN(game[4:])
		isCustom = true
	case strings.HasPrefix(game, "pgn:"):
		opt, err = chess.PGN(strings.NewReader(game[4:]))
	default:
		opt, err = chess.FEN(game)
		isCustom = true
	}
	if opt != nil {
		opts = []func(*chess.Game){opt}
	}
	return
}

func gameToString(game *chess.Game, customGame bool) string {
	if customGame {
		return "fen:" + game.FEN()
	} else {
		return "pgn:" + game.String()[1:]
	}
}
