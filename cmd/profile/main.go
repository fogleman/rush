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

	t0 := time.Now()
	for i := 1; ; i++ {
		generator := rush.NewGenerator(6, 6, 16, 2, rush.Horizontal)
		board := generator.Generate()
		start := time.Now()
		solution := board.Solve(16)
		elapsed := time.Since(start)
		if elapsed < 100*time.Millisecond {
			continue
		}
		gps := float64(i) / time.Since(t0).Seconds()
		fmt.Printf(
			"%6d (%.1f): %8.6f, %5t, %2d, %d, %d\n",
			i, gps, elapsed.Seconds(), solution.Solvable, solution.Depth,
			solution.MemoSize, solution.MemoHits)
		if !solution.Solvable {
			gg.SavePNG(fmt.Sprintf("impossible-%d.png", int(time.Now().Unix())), board.Render())
		}
	}
}
