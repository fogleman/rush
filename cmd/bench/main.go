package main

import (
	"fmt"
	"log"
	"time"

	. "github.com/fogleman/rush"
)

func main() {
	board, err := NewBoard([]string{
		"BCDDE.",
		"BCF.EG",
		"B.FAAG",
		"HHHI.G",
		"..JIKK",
		"LLJMM.",
		// "BB.C..",
		// ".D.CEE",
		// ".DAAFG",
		// "H.IIFG",
		// "H.JKK.",
		// "LLJ...",
	})
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(board.ReachableStates())
	start := time.Now()
	board.Unsolve()
	fmt.Println(time.Since(start))
	// var moves []Move
	// memo := NewMemo()
	// for i := 0; i < 5000000; i++ {
	// 	memo.Add(board.MemoKey(), 0)
	// 	moves = board.Moves(moves)
	// 	board.DoMove(moves[rand.Intn(len(moves))])
	// }
	// fmt.Println(memo.Size())
}
