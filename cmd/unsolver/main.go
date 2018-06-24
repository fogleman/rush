package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

var desc = []string{
	// "JGCBBB",
	// "JGC...",
	// "....AA",
	// "IIEKKH",
	// "LLEF.H",
	// ".DDF.H",
	".BC.DD",
	".BCEEE",
	"...FAA",
	"...FGH",
	"II..GH",
	"JJ...H",
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// board, err := rush.NewBoard(desc)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	best := 0
	for {
		board := rush.NewRandomSolvedBoard(6, 6, 2, 2, 10, 0)
		unsolver := rush.NewUnsolver(board)
		unsolved := unsolver.Unsolve()
		solution := unsolved.Solve()
		n := solution.NumMoves
		if n > best {
			fmt.Println(n)
			gg.SavePNG(fmt.Sprintf("unsolver-%02d-input.png", n), board.Render())
			gg.SavePNG(fmt.Sprintf("unsolver-%02d-output.png", n), unsolved.Render())
			best = n
		}
	}
}
