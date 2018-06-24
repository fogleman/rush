package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	. "github.com/fogleman/rush"
)

const N = 100000

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// board, err := NewBoard([]string{
	// 	"KGGGEE",
	// 	"K.....",
	// 	"AAC..I",
	// 	"..CFFI",
	// 	"HHCJDI",
	// 	"BBBJD.",
	// })
	// if err != nil {
	// 	log.Fatal(err)
	// }

	start := time.Now()
	count := 0
	for i := 0; i < N; i++ {
		board := NewRandomBoard(6, 6, 2, 2, 10, 0)
		if board.Impossible() {
			count++
		}
	}
	elapsed := time.Since(start)
	rate := N / elapsed.Seconds()
	pct := float64(count) / N
	fmt.Println(elapsed, rate, pct)
}
