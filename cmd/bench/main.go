package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fogleman/rush"
)

func solve(desc []string) {
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	moves, ok := board.Solve(16)
	elapsed := time.Since(start)

	fmt.Println(ok, len(moves), elapsed)
}

func main() {
	desc := []string{
		"BBBCDE",
		"FGGCDE",
		"F.AADE",
		"HHI...",
		".JI.KK",
		".JLLMM",
	}
	for i := 0; i < 10; i++ {
		solve(desc)
	}
}
