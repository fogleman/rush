package main

import (
	"fmt"
	// _ "net/http/pprof"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

// 1237 9.586699271s
// 1237 2.510939981s

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	best := 0
	worst := 0
	// start := time.Now()
	for i := 0; ; i++ {
		generator := rush.NewGenerator(6, 6, 16, 2, rush.Horizontal)
		board := generator.Generate()
		// start := time.Now()
		solution := board.Solve(16)
		// elapsed := time.Since(start)

		if !solution.Solvable && solution.MemoSize > worst {
			worst = solution.MemoSize
			gg.SavePNG(fmt.Sprintf("impossible-%07d-%02d.png", solution.MemoSize, solution.Depth), board.Render())
		}
		if solution.NumMoves > best {
			best = solution.NumMoves
			gg.SavePNG(fmt.Sprintf("possible-%02d.png", best), board.Render())
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
