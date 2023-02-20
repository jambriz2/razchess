package main

import (
	"fmt"
	"os"

	"github.com/razzie/razchess/pkg/connector"
)

const (
	MaxDepth = 20
	MoveTime = 60000
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s [w|b|w+b] [session URL]\n", os.Args[0])
		os.Exit(1)
	}
	color := os.Args[1]
	sessionURL := os.Args[2]

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

	bot := NewBot(MoveTime, MaxDepth)
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
				bot.Update(update.FEN, update.PGN)
				move := bot.BestMove()
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
