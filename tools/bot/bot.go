package main

import (
	"math"

	"github.com/razzie/blunder/engine"
	"github.com/razzie/razchess/pkg/razchess"
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
	startingFEN, moves, err := razchess.ParsePGN(PGN)
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
		//bot.search.Pos.StatePly--
		if bot.search.Pos.StatePly == 99 {
			bot.search.Pos.StatePly = 0
		}
	}
	bot.moves = len(moves)
}

func (bot *Bot) BestMove() string {
	return bot.search.Search().String()
}

func (bot *Bot) Reset() {
	bot.setup = false
	bot.moves = 0
	bot.search.TT.Clear()
	bot.search.ClearHistoryTable()
	bot.search.ClearKillers()
	bot.search.ClearCounterMoves()
}
