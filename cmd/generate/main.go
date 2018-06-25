package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	generator := rush.NewDefaultGenerator()
	for i := 0; ; i++ {
		board := generator.Generate(100000)
		board.SortPieces()
		solution := board.Solve()
		gg.SavePNG(fmt.Sprintf("%02d-%d.png", solution.NumMoves, int(time.Now().Unix())), board.Render())
	}
}
