package main

import (
	"fmt"
	"log"
	"math"

	"github.com/fogleman/rush"
)

func main() {
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

	// var moves []rush.Move
	// for i := 0; i < 10; i++ {
	// 	moves = board.Moves(moves)
	// 	move := moves[rand.Intn(len(moves))]
	// 	board.DoMove(move)
	// 	fmt.Println(len(moves), moves)
	// 	fmt.Println(move)
	// 	fmt.Println()
	// 	fmt.Println(board)
	// 	fmt.Println()
	// }

	moves := board.Solve(16)
	fmt.Println(moves)
	sum := 0
	for _, move := range moves {
		board.DoMove(move)
		fmt.Println(move)
		fmt.Println(board)
		fmt.Println()
		sum += int(math.Abs(float64(move.Steps)))
	}
	fmt.Println(sum)
}
