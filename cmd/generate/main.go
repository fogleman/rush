package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	generator := rush.NewDefaultGenerator()
	for i := 0; ; i++ {
		board := generator.Generate(10000)
		solution := board.Solve()
		fmt.Println(board)
		fmt.Println(solution)
		fmt.Println()
		gg.SavePNG(fmt.Sprintf("%02d-%d.png", solution.NumMoves, int(time.Now().Unix())), board.Render())
	}
}
