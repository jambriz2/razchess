package main

import (
	"fmt"
	"math"
	"os"

	"github.com/razzie/blunder/engine"
	"github.com/razzie/razchess/pkg/connector"
)

var Search engine.Search

const (
	MaxDepth = 20
	MoveTime = 120000
)

func init() {
	engine.InitBitboards()
	engine.InitTables()
	engine.InitZobrist()
	engine.InitEvalBitboards()
	engine.InitSearchTables()

	Search.TT.Resize(engine.DefaultTTSize, engine.SearchEntrySize)
	timeLeft, increment, movesToGo, maxNodeCount := engine.InfiniteTime, engine.NoValue, int16(engine.NoValue), uint64(math.MaxUint64)
	Search.Timer.Setup(
		timeLeft,
		increment,
		MoveTime,
		movesToGo,
		MaxDepth,
		maxNodeCount,
	)
}

func getBestMove(fen string) string {
	Search.Setup(fen)
	return Search.Search().String()
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: bot [w|b] [session URL]")
		os.Exit(1)
	}
	color := os.Args[1]
	sessionURL := os.Args[2]

	if color != "w" && color != "b" {
		fmt.Println("invalid color:", color)
		os.Exit(1)
	}

	conn, err := connector.NewConnection(sessionURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	for update := range conn.C {
		if len(update.Opening) > 0 {
			fmt.Println(update.Opening, "-", update.Status)
		} else {
			fmt.Println(update.Status)
		}
		if update.IsGameOver {
			return
		} else if update.Turn == color {
			move := getBestMove(update.FEN)
			valid := conn.Move(move)
			fmt.Println("Found move:", move, ", move accepted:", valid)
		}
	}
}
