package main

import (
	"github.com/razzie/blunder/engine"
)

func coordinateToPos(coordinate string) uint8 {
	file := coordinate[0] - 'a'
	rank := int(coordinate[1]-'0') - 1
	return uint8(rank*8 + int(file))
}

func moveFromCoord(pos *engine.Position, move string) engine.Move {
	from := coordinateToPos(move[0:2])
	to := coordinateToPos(move[2:4])
	moved := pos.Squares[from].Type

	var moveType uint8
	flag := engine.NoFlag

	moveLen := len(move)
	if moveLen == 5 {
		moveType = engine.Promotion
		if move[moveLen-1] == 'n' {
			flag = engine.KnightPromotion
		} else if move[moveLen-1] == 'b' {
			flag = engine.BishopPromotion
		} else if move[moveLen-1] == 'r' {
			flag = engine.RookPromotion
		} else if move[moveLen-1] == 'q' {
			flag = engine.QueenPromotion
		}
	} else if move == "e1g1" && moved == engine.King {
		moveType = engine.Castle
	} else if move == "e1c1" && moved == engine.King {
		moveType = engine.Castle
	} else if move == "e8g8" && moved == engine.King {
		moveType = engine.Castle
	} else if move == "e8c8" && moved == engine.King {
		moveType = engine.Castle
	} else if to == pos.EPSq && moved == engine.Pawn {
		moveType = engine.Attack
		flag = engine.AttackEP
	} else {
		captured := pos.Squares[to]
		if captured.Type == engine.NoType {
			moveType = engine.Quiet
		} else {
			moveType = engine.Attack
		}
	}
	return engine.NewMove(from, to, moveType, flag)
}
