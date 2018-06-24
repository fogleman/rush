package main

import (
	"fmt"
	"math/rand"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func main() {
	best := 0
	worst := 0
	generator := rush.NewDefaultGenerator()
	for i := 0; ; i++ {
		numPieces := rand.Intn(14) + 1
		numWalls := 2
		board := generator.Generate(numPieces, numWalls)
		solution := board.Solve()
		if !solution.Solvable && solution.MemoSize > worst {
			worst = solution.MemoSize
			gg.SavePNG(fmt.Sprintf("impossible-%07d-%02d.png", solution.MemoSize, solution.Depth), board.Render())
		}
		if solution.NumMoves > best {
			best = solution.NumMoves
			fmt.Println(solution)
			gg.SavePNG(fmt.Sprintf("possible-%02d.png", best), board.Render())
		}
	}
}
