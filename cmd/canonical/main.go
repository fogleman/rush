package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fogleman/rush"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	board := rush.NewRandomBoard(6, 6, 2, 2, 8, 0)

	fmt.Println(board)
	fmt.Println()

	canonical := board.Canonicalize()

	fmt.Println(canonical)
}
