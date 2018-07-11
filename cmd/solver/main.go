package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

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

	gg.SavePNG(fmt.Sprintf("solver-%02d.png", 0), board.Render())
	for i, move := range solution.Moves {
		board.DoMove(move)
		gg.SavePNG(fmt.Sprintf("solver-%02d.png", i+1), board.Render())
	}
}
