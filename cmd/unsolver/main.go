package main

import (
	"fmt"
	"log"

	"github.com/fogleman/rush"
)

var desc = []string{
	"JGCBBB",
	"JGC...",
	"....AA",
	"IIEKKH",
	"LLEF.H",
	".DDF.H",
}

func main() {
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(board)
	fmt.Println()

	unsolver := rush.NewUnsolver(board)
	unsolved := unsolver.Unsolve()
	solution := unsolved.Solve()

	fmt.Println(unsolved)
	fmt.Println()
	fmt.Println(solution)
}
