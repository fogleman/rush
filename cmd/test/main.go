package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func generateAndSolve() (*rush.Board, []rush.Move, bool) {
	generator := rush.NewGenerator(6, 6, 16, 2, rush.Horizontal)
	board := generator.Generate()
	// start := time.Now()
	moves, ok := board.Solve(16)
	// elapsed := time.Since(start)
	return board, moves, ok
}

func main() {
	// best := 0
	// for {
	// 	board, moves, ok := generateAndSolve()
	// 	if !ok {
	// 		continue
	// 	}
	// 	if len(moves) > best {
	// 		best = len(moves)
	// 		fmt.Println(board)
	// 		fmt.Println(len(moves), moves)
	// 		fmt.Println()
	// 	}
	// }

	// desc := []string{
	// 	"......",
	// 	"....DD",
	// 	"..AAB.",
	// 	"..E.B.",
	// 	"..ECCC",
	// 	"......",
	// }
	// desc := []string{
	// 	"BBBCDE",
	// 	"FGGCDE",
	// 	"F.AADE",
	// 	"HHI...",
	// 	".JI.KK",
	// 	".JLLMM",
	// }
	desc := []string{
		"J.BHHH",
		"J.B.DI",
		"AA.FDI",
		"EEEFDI",
		"...CKK",
		".GGC..",
	}
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(board)
	fmt.Println()

	// generator := rush.NewGenerator(6, 6, 16, 2, rush.Horizontal)
	// board := generator.Generate()

	// fmt.Println(board)
	// fmt.Println()

	gg.SavePNG(fmt.Sprintf("%04d.png", 0), board.Render())

	start := time.Now()
	moves, ok := board.Solve(16)
	elapsed := time.Since(start)

	if !ok {
		fmt.Println("no solution")
		fmt.Println(elapsed)
		return
	}

	fmt.Println(moves)
	sum := 0
	for i, move := range moves {
		board.DoMove(move)
		fmt.Println(move)
		fmt.Println(board)
		fmt.Println()
		sum += move.AbsSteps()
		gg.SavePNG(fmt.Sprintf("%04d.png", i+1), board.Render())
	}
	fmt.Println(len(moves))
	fmt.Println(sum)
	fmt.Println(elapsed)
}
