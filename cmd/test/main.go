package main

import (
	"fmt"
	"math/rand"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

// 1237 9.586699271s
// 1237 2.510939981s

func main() {
	best := 0
	worst := 0
	generator := rush.NewDefaultGenerator()
	for i := 0; ; i++ {
		n := rand.Intn(14) + 1
		board := generator.Generate(n)
		solution := board.Solve()
		if !solution.Solvable && solution.MemoSize > worst {
			worst = solution.MemoSize
			gg.SavePNG(fmt.Sprintf("impossible-%07d-%02d.png", solution.MemoSize, solution.Depth), board.Render())
		}
		if solution.NumMoves > best {
			best = solution.NumMoves
			gg.SavePNG(fmt.Sprintf("possible-%02d.png", best), board.Render())
		}
	}
}
