package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

var Board = []string{
	".HHHII",
	".xDx.F",
	"AAD..F",
	"CKKE.G",
	"CJJE.G",
	"..BBBx",
}

func main() {
	board, err := rush.NewBoard(Board)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	solution := board.Solve()
	elapsed := time.Since(start)

	fmt.Println(solution)
	fmt.Println(elapsed)

	gg.SavePNG(fmt.Sprintf("%02d.png", 0), board.Render())
	for i, move := range solution.Moves {
		board.DoMove(move)
		gg.SavePNG(fmt.Sprintf("%02d.png", i+1), board.Render())
	}
}
