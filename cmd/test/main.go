package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

// 1237 9.586699271s
// 1237 2.510939981s

func generateAndSolve() (*rush.Board, []rush.Move, bool) {
	generator := rush.NewGenerator(6, 6, 16, 2, rush.Horizontal)
	board := generator.Generate()
	start := time.Now()
	solution := board.Solve(16)
	elapsed := time.Since(start)
	if elapsed > 50*time.Millisecond {
		fmt.Println(solution.Solvable, solution.NumMoves, elapsed)
		fmt.Println(board)
		fmt.Println()
		gg.SavePNG(fmt.Sprintf("impossible-%d.png", int(time.Now().Unix())), board.Render())
	}
	return board, solution.Moves, solution.Solvable
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	best := 0
	start := time.Now()
	for i := 0; ; i++ {
		board, moves, ok := generateAndSolve()
		if !ok {
			continue
		}
		if len(moves) > best {
			elapsed := time.Since(start)
			best = len(moves)
			fmt.Println(i, elapsed)
			fmt.Println(board)
			fmt.Println(len(moves), moves)
			fmt.Println()
			gg.SavePNG(fmt.Sprintf("possible-%d.png", int(time.Now().Unix())), board.Render())
		}
	}

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
	// desc := []string{
	// 	"J.BHHH",
	// 	"J.B.DI",
	// 	"AA.FDI",
	// 	"EEEFDI",
	// 	"...CKK",
	// 	".GGC..",
	// }
	// board, err := rush.NewBoard(desc)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(board)
	// fmt.Println()

	// // generator := rush.NewGenerator(6, 6, 16, 2, rush.Horizontal)
	// // board := generator.Generate()

	// // fmt.Println(board)
	// // fmt.Println()

	// gg.SavePNG(fmt.Sprintf("%04d.png", 0), board.Render())

	// start := time.Now()
	// moves, ok := board.Solve(16)
	// elapsed := time.Since(start)

	// if !ok {
	// 	fmt.Println("no solution")
	// 	fmt.Println(elapsed)
	// 	return
	// }

	// fmt.Println(moves)
	// sum := 0
	// for i, move := range moves {
	// 	board.DoMove(move)
	// 	fmt.Println(move)
	// 	fmt.Println(board)
	// 	fmt.Println()
	// 	sum += move.AbsSteps()
	// 	gg.SavePNG(fmt.Sprintf("%04d.png", i+1), board.Render())
	// }
	// fmt.Println(len(moves))
	// fmt.Println(sum)
	// fmt.Println(elapsed)
}
