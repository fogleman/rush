package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/fogleman/rush"
)

func main() {
	// define the puzzle in ASCII
	desc := []string{
		"BBBCDE",
		"FGGCDE",
		"F.AADE",
		"HHI...",
		".JI.KK",
		".JLLMM",
	}

	// parse and create a board
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}

	// compute a solution
	solution := board.Solve()

	// print out solution information
	fmt.Printf("solvable: %t\n", solution.Solvable)
	fmt.Printf(" # moves: %d\n", solution.NumMoves)
	fmt.Printf(" # steps: %d\n", solution.NumSteps)

	// print out moves to solve puzzle
	moveStrings := make([]string, len(solution.Moves))
	for i, move := range solution.Moves {
		moveStrings[i] = move.String()
	}
	fmt.Println(strings.Join(moveStrings, ", "))

	// solvable: true
	//  # moves: 49
	//  # steps: 93
	// A-1, C+2, B+1, E+1, F-1, A-1, I-1, K-2, D+2, B+2, G+2, I-2, A+1, H+1,
	// F+4, A-1, H-1, I+2, B-2, E-1, G-3, C-1, D-2, I-1, H+4, F-1, J-1, K+2,
	// L-2, C+3, I+3, A+2, G+2, F-3, H-2, D+1, B+1, J-3, A-2, H-2, C-2, I-2,
	// K-4, C+1, I+1, M-2, D+2, E+3, A+4
}
