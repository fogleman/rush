package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func formatMoves(moves []rush.Move) string {
	items := make([]string, len(moves))
	for i, move := range moves {
		items[i] = move.String()
	}
	return strings.Join(items, " ")
}

func main() {
	board, err := rush.NewBoardFromString(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	solution := board.Solve()
	elapsed := time.Since(start)

	fmt.Println(solution)
	fmt.Println(elapsed)

	if !solution.Solvable {
		return
	}

	moves := board.Humanize(solution.Moves)
	if err := board.ValidateSolution(moves); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("solver:    %s (stretch %d)\n",
		formatMoves(solution.Moves), board.MoveStretch(solution.Moves))
	fmt.Printf("humanized: %s (stretch %d)\n",
		formatMoves(moves), board.MoveStretch(moves))

	gg.SavePNG(fmt.Sprintf("solver-%03d.png", 0), board.Render())
	for i, move := range moves {
		board.DoMove(move)
		gg.SavePNG(fmt.Sprintf("solver-%03d.png", i+1), board.Render())
	}
}
