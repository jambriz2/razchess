package main

import (
	"fmt"
	"os"
	"time"

	"github.com/notnil/chess"
	"github.com/notnil/chess/uci"
	"github.com/razzie/razchess/pkg/connector"
)

const (
	MaxDepth = 20
	MoveTime = 30000
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: %s [w|b|w+b] [session URL] [UCI app path]\n", os.Args[0])
		os.Exit(1)
	}
	color := os.Args[1]
	sessionURL := os.Args[2]
	uciApp := os.Args[3]

	if color != "w" && color != "b" && color != "w+b" {
		fmt.Println("invalid color:", color)
		os.Exit(1)
	}

	conn, err := connector.NewConnection(sessionURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	eng, err := uci.New(uciApp)
	if err != nil {
		fmt.Println("failed to start UCI app:", err)
		os.Exit(1)
	}
	defer eng.Close()

	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}

	for update := range conn.C {
		if len(update.Opening) > 0 {
			fmt.Println(update.FEN, "-", update.Opening, "-", update.Status)
		} else {
			fmt.Println(update.FEN, "-", update.Status)
		}
		if update.IsGameOver {
			return
		} else if update.Turn == color || color == "w+b" {
			for {
				move := getBestMove(eng, update.PGN)
				fmt.Println("Best move:", move)
				if conn.Move(move) {
					break
				} else {
					fmt.Println("Server rejected the move (maybe someone else moved a piece?)")
					newUpdate := conn.State.Load()
					if newUpdate == update {
						fmt.Println("Bot error")
						os.Exit(1)
					} else {
						break
					}
				}
			}
		}
	}
}

func getBestMove(eng *uci.Engine, pgn string) string {
	game := chess.NewGame()
	if err := game.UnmarshalText([]byte(pgn)); err != nil {
		panic(err)
	}
	cmdPos := uci.CmdPosition{Position: game.Position()}
	cmdGo := uci.CmdGo{
		MoveTime: time.Millisecond * MoveTime,
		Depth:    MaxDepth,
	}
	if err := eng.Run(cmdPos, cmdGo); err != nil {
		panic(err)
	}
	move := eng.SearchResults().BestMove
	if err := game.Move(move); err != nil {
		panic(err)
	}
	return chess.UCINotation{}.Encode(game.Position(), move)
}
