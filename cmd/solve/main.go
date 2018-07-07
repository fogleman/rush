package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fogleman/rush"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("solve DESC")
		return
	}

	board, err := rush.NewBoardFromString(args[0])
	if err != nil {
		log.Fatal(err)
	}

	solution := board.Solve()
	fmt.Println(solution)
}
