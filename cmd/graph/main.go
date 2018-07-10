package main

import (
	"log"
	"os"

	. "github.com/fogleman/rush"
)

func main() {
	board, err := NewBoardFromString(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	Graph(board)
}
