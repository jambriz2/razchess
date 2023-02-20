package main

import (
	"math"

	"github.com/razzie/blunder/engine"
)

func init() {
	engine.InitBitboards()
	engine.InitTables()
	engine.InitZobrist()
	engine.InitEvalBitboards()
	engine.InitSearchTables()
}

type Bot struct {
	search engine.Search
	setup  bool
	moves  int
}

func NewBot(moveTime int64, maxDepth uint8) *Bot {
	bot := &Bot{}
	bot.search.TT.Resize(engine.DefaultTTSize, engine.SearchEntrySize)
	timeLeft, increment, movesToGo, maxNodeCount := engine.InfiniteTime, engine.NoValue, int16(engine.NoValue), uint64(math.MaxUint64)
	bot.search.Timer.Setup(
		timeLeft,
		increment,
		moveTime,
		movesToGo,
		maxDepth,
		maxNodeCount,
	)
	return bot
}

func (bot *Bot) Update(FEN, PGN string) {
	startingFEN, moves, err := parsePGN(PGN)
	if err != nil {
		panic(err)
	}

	if !bot.setup {
		bot.search.Setup(startingFEN)
		for _, move := range moves {
			bot.search.Pos.DoMove(moveFromCoord(&bot.search.Pos, move))
			bot.search.AddHistory(bot.search.Pos.Hash)
			bot.search.Pos.StatePly--
		}
		bot.moves = len(moves)
		bot.setup = true
	}

	newMoves := moves[bot.moves:]
	for _, move := range newMoves {
		bot.search.Pos.DoMove(moveFromCoord(&bot.search.Pos, move))
		bot.search.AddHistory(bot.search.Pos.Hash)
		bot.search.Pos.StatePly--
	}
	bot.moves = len(moves)
}

func (bot *Bot) BestMove() string {
	return bot.search.Search().String()
}
