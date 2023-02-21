package razchess

import (
	crand "crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"strings"
	"time"

	"github.com/notnil/chess"
	"gopkg.in/freeeve/pgn.v1"
)

const (
	StartingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	charset     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	csLen       = byte(len(charset))
)

func GenerateID(length int) string {
	if length == 0 {
		return ""
	}
	output := make([]byte, 0, length)
	batchSize := length + length/4
	buf := make([]byte, batchSize)
	for {
		if _, err := io.ReadFull(crand.Reader, buf); err != nil {
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

func GenerateFischerRandomFEN() string {
	var setup setup
	rnd := mrand.New(mrand.NewSource(time.Now().Unix()))
	// king
	king := rnd.Intn(6) + 1 // between ranks B and G to have room for rooks
	setup[king] = 'k'
	// rooks
	leftRook := rnd.Intn(king)
	rightRook := rnd.Intn(7-king) + king + 1
	setup[leftRook] = 'r'
	setup[rightRook] = 'r'
	// bishops
	emptySquares := setup.emptySquares()
	firstBishop := emptySquares[rnd.Intn(len(emptySquares))]
	setup[firstBishop] = 'b'
	emptySquares = setup.emptySquaresFilter(func(idx int) bool { return idx%2 != firstBishop%2 })
	secondBishop := emptySquares[rnd.Intn(len(emptySquares))]
	setup[secondBishop] = 'b'
	// knights
	emptySquares = setup.emptySquares()
	firstKnight := emptySquares[rnd.Intn(len(emptySquares))]
	setup[firstKnight] = 'n'
	emptySquares = setup.emptySquares()
	secondKnight := emptySquares[rnd.Intn(len(emptySquares))]
	setup[secondKnight] = 'n'
	// queen
	emptySquares = setup.emptySquares()
	queen := emptySquares[0]
	setup[queen] = 'q'
	return setup.String()
}

type setup [8]rune

func (s *setup) emptySquares() (sqs []int) {
	for i, sq := range s {
		if sq == 0 {
			sqs = append(sqs, i)
		}
	}
	return
}

func (s *setup) emptySquaresFilter(cond func(idx int) bool) (sqs []int) {
	for i, sq := range s {
		if sq == 0 && cond(i) {
			sqs = append(sqs, i)
		}
	}
	return
}

func (s setup) String() string {
	var rank string
	emptyCount := 0
	for _, sq := range s {
		if sq == 0 {
			emptyCount++
		} else {
			if emptyCount > 0 {
				rank += fmt.Sprint(emptyCount)
				emptyCount = 0
			}
			rank += string(sq)
		}
	}
	return fmt.Sprintf("%s/pppppppp/8/8/8/8/PPPPPPPP/%s w KQkq - 0 1", strings.ToLower(rank), strings.ToUpper(rank))
}

func getFENTagPairsOpt(fen string) func(*chess.Game) {
	return chess.TagPairs([]*chess.TagPair{
		{
			Key:   "SetUp",
			Value: "1",
		},
		{
			Key:   "FEN",
			Value: fen,
		},
	})
}

func parseGame(game string) ([]func(*chess.Game), error) {
	switch {
	case len(game) == 0:
		return nil, nil

	case strings.HasPrefix(game, "pgn:"):
		opt, err := chess.PGN(strings.NewReader(game[4:]))
		if err != nil {
			return nil, err
		}
		return []func(*chess.Game){opt}, nil

	case strings.HasPrefix(game, "fen:"):
		fen := game[4:]
		opt, err := chess.FEN(fen)
		if err != nil {
			return nil, err
		}
		return []func(*chess.Game){opt, getFENTagPairsOpt(fen)}, nil

	default:
		opt, err := chess.FEN(game)
		if err != nil {
			return nil, err
		}
		return []func(*chess.Game){opt, getFENTagPairsOpt(game)}, nil
	}
}

func gameToString(game *chess.Game) string {
	if len(game.Moves()) > 0 {
		return "pgn:" + strings.TrimSpace(game.String())
	}
	return "fen:" + game.FEN()
}

func ParsePGN(PGN string) (startingFEN string, moves []string, err error) {
	startingFEN = StartingFEN
	ps := pgn.NewPGNScanner(strings.NewReader(PGN))
	if ps.Next() {
		var game *pgn.Game
		game, err = ps.Scan()
		if err != nil {
			return
		}
		if fen, ok := game.Tags["FEN"]; ok {
			startingFEN = fen
		}
		moves = make([]string, len(game.Moves))
		for i, move := range game.Moves {
			moves[i] = move.String()
		}
	}
	return
}
