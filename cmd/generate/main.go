package main

import (
	"fmt"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func main() {
	generator := rush.NewDefaultGenerator()
	for i := 0; ; i++ {
		board := generator.Generate(1000)
		solution := board.Solve()
		fmt.Println(board)
		fmt.Println(solution)
		fmt.Println()
		gg.SavePNG(fmt.Sprintf("%02d-%d.png", solution.NumMoves, int(time.Now().Unix())), board.Render())
	}
}
