package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fogleman/rush"
)

func main() {
	// desc := []string{
	// 	"......",
	// 	"....DD",
	// 	"..AAB.",
	// 	"..E.B.",
	// 	"..ECCC",
	// 	"......",
	// }
	desc := []string{
		"BBBCDE",
		"FGGCDE",
		"F.AADE",
		"HHI...",
		".JI.KK",
		".JLLMM",
	}
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(board)
	fmt.Println()

	// gg.SavePNG(fmt.Sprintf("%04d.png", 0), board.Render())

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
	for _, move := range moves {
		board.DoMove(move)
		fmt.Println(move)
		fmt.Println(board)
		fmt.Println()
		sum += move.AbsSteps()
		// gg.SavePNG(fmt.Sprintf("%04d.png", i+1), board.Render())
	}
	fmt.Println(len(moves))
	fmt.Println(sum)
	fmt.Println(elapsed)
}
